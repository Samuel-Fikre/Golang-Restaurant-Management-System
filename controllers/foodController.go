package controllers

import (
	"context"
	"log"
	"math"
	"net/http"
	"restaurant-management-system/database"
	"restaurant-management-system/models"

	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10" // Add this import
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Ensure cancel is called before returning

		// c.Query("recordPerPage") extracts the value of the recordPerPage query parameter from the URL (e.g., in http://example.com?page=2&recordPerPage=10, the value of recordPerPage would be "10").

		// strconv is part of Go's standard library. It contains functions for string conversions, including converting strings to integers.

		// strconv.Atoi(...) attempts to convert the string returned by c.Query("recordPerPage") into an integer.

		// The default value is chosen to ensure the application has a reasonable, valid value to work with for pagination purposes. In this case, 10 means that the system will show 10 records per page if the user input is invalid or missing.

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		// For page 1, startIndex is 0 (starting from the first record).
		//For page 2, startIndex is 10 (starting from the 11th record).
		//For page 3, startIndex is 20 (starting from the 21st record).

		startIndex := (page - 1) * recordPerPage // Remove the conflicting reassign

		// "$match": This is an aggregation operator used in MongoDB to filter documents.
		// It selects documents that match the specified condition.
		// Value: bson.D{{}} - This indicates that no specific matching criteria are set, allowing all documents to be included.
		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		// "$group": This is an aggregation operator that groups documents together based on a specified key
		groupStage := bson.D{
			{Key: "$group", Value: bson.D{
				// Setting _id to nil means that theres no specific grouping criterion. Instead, it groups all documents into a single group.
				// total_count:-This is the name of the new field in the output document that will store the count of the documents
				//For each document processed, MongoDB sees 1 and adds it to a total count.
				//If there are three documents in the group, total_count will be 3.
				// The $push operator adds items to an array. In this case, it pushes the entire document being processed into the array.
				// So when you push $$ROOT, you are pushing the complete document, not just specific fields.

				//{
				//"_id": null,
				//"total_count": 3,
				//"data": [
				//{ "name": "Pizza", "type": "fast food" },
				//{ "name": "Sushi", "type": "Japanese" },
				//{ "name": "Burger", "type": "fast food" }
				//]
				//}

				{Key: "_id", Value: nil},
				{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
			}},
		}

		// Index is page in this case
		//Suppose $data contains the following array of food items:
		//[
		//{"name": "Pizza"},
		//{"name": "Burger"},
		//{"name": "Sushi"},
		//{"name": "Pasta"},
		//{"name": "Salad"},
		//{"name": "Tacos"},
		//{"name": "Ice Cream"},
		//{"name": "Cake"}
		//]

		// If startIndex is 4 and recordPerPage is 3, the $slice operation would yield:
		//[
		//{"name": "Salad"},
		//{"name": "Tacos"},
		//{"name": "Ice Cream"}
		//]

		// This stage is crucial for returning a clean and relevant set of results for the client, ensuring they get only what they need to display on their UI.
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "food_items", Value: bson.D{{Key: "$slice", Value: bson.A{"$data", startIndex, recordPerPage}}}},
			}},
		}

		// example
		//Filter: Retrieve only the items in the "Fruits" category.
		//Group: Calculate the total quantity of fruits.
		//Project: Return the total quantity and the names of the fruit

		result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while fetching the food items"})
			return
		}
		// bson.M is used to hold the data returned from MongoDB queries in Go.in here it is stored that in allFoods

		// allFoods = []bson.M{
		//{"_id": nil, "total_count": 5, "data": [{"name": "Apple", "category": "Fruits", "price": 0.5}, ...]},
		//}

		var allFoods []bson.M

		// This function call retrieves all the documents from the MongoDB cursor (result) and decodes them into the allFoods slice.
		if err = result.All(ctx, &allFoods); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allFoods[0])
	}
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		foodId := c.Param("food_id")
		var food models.Food

		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the food item"})
		}
		c.JSON(http.StatusOK, food)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var menu models.Menu
		var food models.Food
		// The validator package allows you to define rules for each field in your structs. For example, you can specify that a certain field must be a valid email address, must not be empty, must be a specific length, etc. This is done using struct tags.
		// Check your model files
		var validate = validator.New()

		//  A client application (like a web app) sends a POST request to create a new food item:

		// Receives the request.
		//Unmarshals the JSON body into the food struct.
		//Validates the data.
		//Creates a new food item in the database.

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate.Struct(food)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		// This line of code is attempting to find a menu item in the MongoDB collection based on the menu_id associated with the food item being created. If a matching menu document is found, it will be decoded into the menu struct

		// The data is temporarily stored in the menu struct for the duration of the CreateFood function execution. It allows you to use the information from the database (like validating that the menu exists) without having to keep the data in MongoDB at that moment.

		// Therefore, there will be no storage of the menu data in the menuCollection during the execution of this function.
		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "menu was not found"})
			return
		}
		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.ID = primitive.NewObjectID()
		// this line converts the newly generated ObjectID (which is used as the primary key for the food item in MongoDB) into a hexadecimal string representation.
		food.Food_id = food.ID.Hex()
		var num = toFixed(*food.Price, 2) // Dereference food.Price
		food.Price = &num                 // Assign the result back as a pointer

		//  In Go, when you use the MongoDB driver to insert a document into a collection, you don't need to manually marshal (serialize) your struct into a BSON format before insertion. The MongoDB Go driver handles this for you automatically

		result, insertErr := foodCollection.InsertOne(ctx, food)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": insertErr.Error()})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)

	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))

}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output

}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var menu models.Menu
		var food models.Food

		foodId := c.Param("food_id")
		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Think of primitive.D like a list where each item in the list is a pair of a key and a value. This is similar to how a dictionary or a map works in other programming languages

		// Each item in updateObj will represent a field in the food item that you want to change. For example, if you want to change the food's name, you'd add a key-value pair where the key is "name" and the value is the new name.
		var updateObj primitive.D

		// By checking if food.Name is not nil, you ensure that you only update the name field if a new value has actually been provided.

		// This line is crucial for ensuring that you only attempt to update fields in the database if new values are provided in the request.

		//It helps keep the update operation efficient and avoids overwriting existing data with nil values

		// bson.E represents a single entry in a BSON document (which MongoDB uses to store data). It consists of a key (in this case, "name") and a value (the new name that was provided in the request, food.Name).

		// This work for all below codes not only name

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
		}

		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key: "price", Value: food.Price})
		}

		if food.Food_image != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.Food_image})
		}

		if food.Menu_id != nil {
			err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)

			defer cancel()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "menu was not found"})
				return
			}

			updateObj = append(updateObj, bson.E{Key: "menu", Value: food.Menu_id})
		}
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.Updated_at})
		// In the context of MongoDB operations, "upsert" is a combination of "update" and "insert"

		// When upsert is set to true, it means that if the specified document to be updated does not exist, a new document should be created with the specified values.

		// For example, if you attempt to update a food item with food_id "12345" and that item does not exist, a new document will be created with the food_id set to "12345" and the fields specified in updateObj

		upsert := true
		// Suppose you want to update a food item with a specific food_id (for example, "12345").

		//If foodId holds the value "12345",  such a document exists, the updates specified in updateObj will be applied to it
		filter := bson.M{"food_id": foodId}

		// options.UpdateOptions?

		//This is a struct provided by the MongoDB Go driver that contains options you can set when you want to update a document in the database.
		// Think of it as a way to customize how your update operation will behave.
		// The Upsert field in UpdateOptions is a boolean (true or false) that tells MongoDB whether to create a new document if it doesnt find one to update.
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		//  bson.D is commonly used for creating documents to be inserted or updated in the database
		result, err := foodCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Food item update failed"})
		}

		c.JSON(http.StatusOK, result)
	}
}
