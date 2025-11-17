package cmd

import (
	"blog_api/src/handler"
	handlerAction "blog_api/src/handler/action"
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
	rssPostHandler := handler.NewRssPostHandler(db)
	updataHandler := handlerAction.NewUpdataHandler(db)

	// API routes
	api := router.Group("/api")
	{
		// Friend link routes
		friend := api.Group("/friend")
		{
			friend.GET("/", friendLinkHandler.GetAllFriendLinks)
		}
		// RSS post routes
		rss := api.Group("/rss")
		{
			rss.GET("/", rssPostHandler.GetAllPostsByFriendLinkID)
		}
		// Update routes
		// Action routes for friend links
		action := api.Group("/action")
		{
			friendAction := action.Group("/friend")
			{
				friendAction.POST("/", updataHandler.CreateFriendLink)
				friendAction.PUT("/", updataHandler.EditFriendLink)
				friendAction.DELETE("/", updataHandler.DeleteFriendLink)
			}
		}
	}

	return router
}
