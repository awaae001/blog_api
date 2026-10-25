package cmd

import (
	"blog_api/src/handler"
	"blog_api/src/middleware"
	"blog_api/src/service"

	"github.com/gin-gonic/gin"
)

func registerStateRoutes(apiGroup *gin.RouterGroup, masterPassword string) *service.StateStore {
	stateHandler := handler.NewStateHandler()
	stateGroup := apiGroup.Group("/internal/states")
	stateGroup.Use(middleware.StateMasterAuth(masterPassword))
	stateGroup.PUT("/:key", stateHandler.PutState)
	stateGroup.GET("/:key", stateHandler.GetState)
	stateGroup.DELETE("/:key", stateHandler.DeleteState)
	return stateHandler.Store
}
