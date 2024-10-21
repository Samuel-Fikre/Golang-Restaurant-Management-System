package routes

import (
	controller "restaurant-management-system/controllers"

	"github.com/gin-gonic/gin"
)

func OrderRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/orders", controller.GetOrders())               // Retrieve all orders
	incomingRoutes.GET("/orders/:order_id", controller.GetOrder())      // Retrieve a specific order by ID
	incomingRoutes.POST("/orders", controller.CreateOrder())            // Create a new order
	incomingRoutes.PATCH("/orders/:order_id", controller.UpdateOrder()) // Update a specific order by ID
}
