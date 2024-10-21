package routes

import (
	controller "restaurant-management-system/controllers"

	"github.com/gin-gonic/gin"
)

func MenuRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/menus", controller.GetMenus())              // Retrieve all menus
	incomingRoutes.GET("/menus/:menu_id", controller.GetMenu())      // Retrieve a specific menu by ID
	incomingRoutes.POST("/menus", controller.CreateMenu())           // Create a new menu
	incomingRoutes.PATCH("/menus/:menu_id", controller.UpdateMenu()) // Update a specific menu by ID
}
