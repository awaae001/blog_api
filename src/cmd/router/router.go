package cmd

import (
	"blog_api/src/handler"
	"blog_api/src/model"
	"database/sql"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the Gin router
func SetupRouter(db *sql.DB, cfg *model.Config) *gin.Engine {
	// Set Gin mode (release mode in production)
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Configure CORS middleware
	corsConfig := cors.Config{
		AllowOrigins:     cfg.Safe.CorsAllowHostlist,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// Initialize handlers
	friendLinkHandler := handler.NewFriendLinkHandler(db)

	// API routes
	api := router.Group("/api")
	{
		// Friend link routes
		friend := api.Group("/friend")
		{
			friend.GET("/", friendLinkHandler.GetAllFriendLinks)
		}
	}

	return router
}
