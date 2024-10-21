package controllers

import (
	"context"
	"log"
	"net/http"
	"restaurant-management-system/database"
	"restaurant-management-system/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OrderItem Struct: This struct is designed to represent a single order item, including its attributes like quantity, unit price, food ID, and timestamps. It is focused on the details of one item.

// OrderItemPack Struct: This struct is a higher-level container that represents a collection of order items and additional context, such as the Table_id. Its designed to group multiple OrderItem instances together for a specific order.

// Example :  If a customer orders a burger, the OrderItem struct might look like this:
// {
// 	ID: "123",
// 	Quantity: "1",
// 	Unit_Price: 5.99,
// 	Food_id: "burger123",
// 	CreatedAt: "2023-08-01T10:00:00Z",
// 	UpdatedAt: "2023-08-01T10:00:00Z"
// }

// Example Usage: If a customer orders a burger and a fries, the OrderItemPack struct might look like this:
// Example Usage: If a customer orders a burger and a fries, the OrderItemPack struct might look like this:
// orderItemPack := OrderItemPack{
// 	Table_id: stringPtr("table5"),
// 	Order_items: []models.OrderItem{
// 			burger, // First item: burger
// 			{
// 					ID:          primitive.NewObjectID(),
// 					Quantity:    stringPtr("2"),
// 					Order_ID:    "order123",
// 					Order_Item_Id: primitive.NewObjectID().Hex(),
// 					Unit_Price:  float64Ptr(2.99),
// 					Food_id:     stringPtr("food_fries"),
// 					CreatedAt:   time.Now(),
// 					UpdatedAt:   time.Now(),
// 			},
// 	},
// }

// another example
// orderItemPack := OrderItemPack{
// 	Table_id: stringPtr("5"),
// 	Order_items: []models.OrderItem{
// 			burger,
// 			fries1,
// 			fries2,
// 	},
// }

type OrderItemPack struct {
	Table_id    *string // Reference to the table for the order
	Order_items []models.OrderItem
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		result, err := orderItemCollection.Find(context.TODO(), bson.M{})
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the order items"})
		}

		var allOrderItems []bson.M
		if err = result.All(ctx, &allOrderItems); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allOrderItems)
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {

		orderId := c.Param("order_id")

		allOrderItems, err := ItemsByOrder(orderId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items by order"})
		}
		c.JSON(http.StatusOK, allOrderItems)

	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		orderItemId := c.Param("order_item_id")
		var orderItem models.OrderItem

		err := orderItemCollection.FindOne(ctx, bson.M{"orderItem_id": orderItemId}).Decode(&orderItem)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the order item"})
		}
		c.JSON(http.StatusOK, orderItem)
	}
}

// this will hold the list of order items as documents retrieved from MongoDB.

func ItemsByOrder(id string) (orderItems []primitive.M, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	// $match: The $match is like a filter in a search. In this case, were saying, "Hey MongoDB, find documents (which are like rows in SQL databases) where the order_id equals the id that we passed into the function
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "order_id", Value: id}}}}
	//Its purpose is to join two collections in MongoDB, much like how a SQL JOIN works.
	// In this line, you're trying to get more details about the food associated with each item in the order. The order items are stored in one collection, and the food details are stored in another collection. This stage connects the two.
	// from: "food" This tells MongoDB that the additional information you need is in the food collection
	// localField: "food_id" This is the field in the current collection (order items) that will be used to match the documents. Here, its food_id, which links each order item to a specific food item.
	// foreignField: "food_id"  This is the field in the other collection (food) that MongoDB will use to find matching documents. The food_id field in the food collection needs to match the food_id in the order item
	// as: "food"  This is the name of the new field that will hold the matched data from the food collection. After the lookup, each order item will have a new field called food, which contains details about the food item (like its name, price, etc.).
	// Before lookupStage:

	// You just have:

	// [
	//   { "order_id": "123", "food_id": "001", "quantity": 2 },
	// 	  { "order_id": "123", "food_id": "002", "quantity": 1 }
	// ]

	// This only tells you the food_id and quantity, but not much else.

	// After lookupStage:

	// [
	// {
	// "order_id": "123",
	// "food_id": "001",
	// "quantity": 2,
	// "food": {
	// 	"name": "Pizza",
	// 	"price": 10.0,
	// 	"food_image": "pizza.jpg"
	// }
	// },
	// {
	// "order_id": "123",
	// "food_id": "002",
	// "quantity": 1,
	// "food": {
	// "name": "Burger",
	// "price": 5.0,
	// "food_image": "burger.jpg"
	// }
	// }
	// ]

	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "food"}, {Key: "localField", Value: "food_id"}, {Key: "foreignField", Value: "food_id"}, {Key: "as", Value: "food"}}}}
	// $unwind takes an array from a document and splits it into separate documents for each item in that array.
	// It helps make it easier to work with data that has arrays.
	//  path

	// What It Means: The path specifies which field you want to unwind. In this case, its "$food".
	// How It Works: By saying path: "$food", you are telling MongoDB that you want to take the food array from each order document and create separate documents for each food item.

	// Before unwindStage:

	// {
	// "order_id": 1,
	// "customer_name": "Alice",
	// "food_ids": ["f1", "f2"],
	// "food_items": [
	// { "food_id": "f1", "name": "Pizza" },
	// { "food_id": "f2", "name": "Burger" }
	// ]
	// }

	// After unwindStage:

	// {
	// "order_id": 1,
	// "customer_name": "Alice",
	// "food_ids": ["f1", "f2"],
	// "food_items": { "food_id": "f1", "name": "Pizza" }
	// }

	// {
	// "order_id": 1,
	// "customer_name": "Alice",
	// "food_ids": ["f1", "f2"],
	// "food_items": { "food_id": "f2", "name": "Burger" }
	// }

	// The primary purpose of the $unwind stage in MongoDB is to deconstruct an array field from the input documents. This means that if you have a document containing an array, $unwind will create a new document for each element in that array, effectively "flattening" it

	// Simply ANTI-ARRAY :- Benefit: This allows for easier manipulation and querying of individual array elements in MongoDB.

	//  preserveNullAndEmptyArrays: true ensures that documents with null values or empty arrays for the field being unwound are not excluded from the results. Instead, they will appear in the output, but the unwound field will have a null value.

	unwindStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$food"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	// Seek Mock Data Example if confused
	// Purpose of the as Key:

	// The as key in the $lookup stage specifies the name of the new field that will be created in the resulting documents. In your case, this is the "order" field.
	// The documents from the order collection that match the join condition will be included in this new field as an array.

	// Why the Food Name Is Not Included:

	// The $lookup only brings in data from the order collection based on the matching order_id.
	// It does not include fields from the food collection in the result of the lookup because that part is handled separately in your pipeline.

	// We are concerned with key as it's value
	lookupOrderStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "order"}, {Key: "localField", Value: "order_id"}, {Key: "foreignField", Value: "order_id"}, {Key: "as", Value: "order"}}}}
	unwindOrderStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$order"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	lookupTableStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "table"}, {Key: "localField", Value: "order.table_id"}, {Key: "foreignField", Value: "table_id"}, {Key: "as", Value: "table"}}}}
	unwindTableStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$table"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	// The $project stage is crucial for shaping the final output of your aggregation pipeline. It allows you to control which fields are included, excluded, or renamed in the output documents, helping you create a cleaner and more relevant data structure for further processing or displaying in your application.

	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			// This line excludes the id field from the output documents. Setting a field to 0 in the $project stage means it won't appear in the resulting documents.

			// FIX THE STATIC VALUES THEY ARE NOT SAFE
			{Key: "id", Value: 0},
			{Key: "amount", Value: "$food_price"},
			{Key: "total_count", Value: 1},
			{Key: "food_name", Value: "$food.name"},
			{Key: "food_image", Value: "$food.food_image"},
			{Key: "table_number", Value: "$table.table_number"},
			{Key: "table_id", Value: "$table.table_id"},
			{Key: "order_id", Value: "$order.order_id"},
			{Key: "price", Value: "$food.price"},
			{Key: "quantity", Value: 1},
		}}}

	// The aggregation groups the documents based on both the order_id and table_number. This means that if multiple items belong to the same order (i.e., they have the same order_id) and were placed at the same table (i.e., they have the same table_number), they will be grouped together in a single result.

	groupStage := bson.D{{Key: "$group", Value: bson.D{
		{Key: "_id", Value: bson.D{
			{Key: "order_id", Value: "$order_id"},
			{Key: "table_number", Value: "$table_number"},
		}},
		{Key: "table_id", Value: bson.D{{Key: "$first", Value: "$table_id"}}},
		{Key: "order_id", Value: bson.D{{Key: "$first", Value: "$order_id"}}},
		{Key: "table_number", Value: bson.D{{Key: "$first", Value: "$table_number"}}},
		{Key: "total_count", Value: bson.D{{Key: "$sum", Value: "$quantity"}}},
		{Key: "order_items", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		{Key: "payment", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
		{Key: "total_amount", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
	}}}

	//  In this code, the projectStage2 is applying a projection on the results that were produced by the previous groupStage.

	// The earlier stage typically comes early in the pipeline, before the data is grouped. It helps to reduce the amount of data that is processed by focusing only on the fields you actually need.

	// This second projection stage is used to restructure the grouped results and prepare the final output

	projectStage2 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "id", Value: 0},
			{Key: "payment_due", Value: 1},
			{Key: "table_number", Value: "$_id.table_number"},
			{Key: "order_items", Value: 1},
			{Key: "total_count", Value: 1},
		}}}

	//  the MongoDB aggregation pipeline is executed in sequence, with each stage processing the output of the previous one. This means the order of the stages is very important because each stage depends on the results from the stages before it.
	cursor, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupStage,
		unwindStage,
		lookupOrderStage,
		unwindOrderStage,
		lookupTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		projectStage2,
	})
	defer cancel()
	if err != nil {
		panic(err)
	}

	if err = cursor.All(ctx, &orderItems); err != nil {
		panic(err)
	}

	return orderItems, nil

}

// Order and OrderItem are two different types of documents in the database.
// Order: This document represents a complete order placed by a customer. It includes details like the order date, table ID, and a list of order items.
// OrderItem: This document represents a single item within an order. It includes details like the order ID, food ID, quantity, and unit price.

// Order
// ├── ID: order123
// ├── Order_Date: 2024-10-19
// ├── Table_ID: table5
// ├── OrderItems
// │   ├── OrderItem 1 (e.g., Burger)
// │   │   ├── Quantity: 1
// │   │   ├── Unit_Price: 9.99
// │   ├── OrderItem 2 (e.g., Fries)
// │   │   ├── Quantity: 2
// │   │   ├── Unit_Price: 2.99

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var order models.Order
		var orderItemPack OrderItemPack
		var validate = validator.New()

		if err := c.BindJSON(&orderItemPack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order.Order_Date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		// orderItemsToBeInserted: This is the name of the variable being declared. It's intended to hold a collection of order items that will later be inserted into a database (MongoDB in this case).
		orderItemsToBeInserted := []interface{}{}
		order.Table_ID = orderItemPack.Table_id
		// The line order_id := OrderItemOrderCreator(order) creates a new order and retrieves its unique ID, which is crucial for associating the order items with the correct order.
		// This step establishes the link between the Order and its OrderItems, allowing the application to maintain the integrity of data and relationships in the database
		order_id := OrderItemOrderCreator(order)

		// I take out a car from the box. I check if its okay. I put a sticker on it with my name. Then I put it in a special spot to keep it safe.

		// Why We Need the Loop
		// The loop is essential because it allows you to handle multiple order items efficiently and cleanly.
		// It saves time, reduces mistakes, and keeps the code neat and organized.
		// Rememeber this it is for multiplicity of order items.if it was single item we dont need loop.
		for _, orderItem := range orderItemPack.Order_items {
			orderItem.Order_ID = order_id

			validationErr := validate.Struct(orderItem)
			if validationErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}

			orderItem.ID = primitive.NewObjectID()
			orderItem.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Order_Item_Id = orderItem.ID.Hex()
			var num = toFixed(*orderItem.Unit_Price, 2)
			orderItem.Unit_Price = &num
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
		}

		insertedOrderItems, err := orderItemCollection.InsertMany(ctx, orderItemsToBeInserted)
		if err != nil {
			log.Fatal(err)
		}
		defer cancel()
		c.JSON(http.StatusOK, insertedOrderItems)
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var orderItem models.OrderItem
		orderItemId := c.Param("order_item_id")

		filter := bson.M{"orderItem_id": orderItemId}

		var updateObj primitive.D

		if orderItem.Unit_Price != nil {
			updateObj = append(updateObj, bson.E{Key: "unit_price", Value: orderItem.Unit_Price})
		}
		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{Key: "quantity", Value: orderItem.Quantity})
		}
		if orderItem.Food_id != nil {
			updateObj = append(updateObj, bson.E{Key: "food_id", Value: orderItem.Food_id})
		}

		orderItem.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: orderItem.UpdatedAt})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderItemCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while updating the order item"})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}
