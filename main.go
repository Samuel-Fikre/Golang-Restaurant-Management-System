package main

import (
	"os"
	"restaurant-management-system/database"
	"restaurant-management-system/middleware"
	"restaurant-management-system/routes"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// foodCollection :- This is a variable that holds a reference to the food collection in your MongoDB database. It has the type *mongo.Collection, meaning it's a pointer to a mongo.Collection object (part of the official MongoDB Go driver

// used to create a MongoDB collection
var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

// OpenCollection function to get a collection from the database
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	// Assuming you have a specific database, like "mydb"
	collection := client.Database("mydb").Collection(collectionName)
	return collection
}

func main() {
	port := os.Getenv(("PORT"))

	if port == "" {
		port = "8000"
	}
	// t creates a new instance of the Gin engine (router) without any default middleware.
	router := gin.New()
	// gin.Logger() is a built-in middleware provided by the Gin framework. It logs details about each HTTP request the router receives and the corresponding response. This helps in debugging and monitoring the behavior of your application
	// gin.Logger() logs important information like HTTP methods, paths, response status codes, client IP addresses, and request processing time.
	router.Use(gin.Logger())
	router.UserRoutes(router)
	// used to attach custom authentication middleware to your Gin router. Middleware in Gin acts like a filter that processes every request before it reaches your route handlers. This particular middleware is for authentication, ensuring that only users who are authenticated (logged in or have valid credentials) can access certain routes.
	router.Use(middleware.Authentication())

	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)

	router.Run(":" + port)
}
