package cmd

import (
	"blog_api/src/handler"
	handlerAction "blog_api/src/handler/action"
	"blog_api/src/middleware"
	"blog_api/src/model"
	"database/sql"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the Gin router
func SetupRouter(db *sql.DB, cfg *model.Config, startTime time.Time) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Configure CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Safe.CorsAllowHostlist,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	registerRoutes(router, db, startTime)

	// Serve static files and handle SPA
	router.NoRoute(staticFileHandler(cfg))
	return router
}

func registerRoutes(router *gin.Engine, db *sql.DB, startTime time.Time) {
	// Initialize handlers
	friendLinkHandler := handler.NewFriendLinkHandler(db)
	rssPostHandler := handler.NewRssPostHandler(db)
	updataHandler := handlerAction.NewUpdataHandler(db)
	friendRssHandler := handlerAction.NewFriendRssHandler(db)
	authHandler := handler.NewAuthHandler()
	statusHandler := handler.NewStatusHandler(db, startTime)

	// API routes
	apiGroup := router.Group("/api")
	{
		// Status router
		apiGroup.GET("/status", statusHandler.GetSystemStatus)
		// Authentication routes
		apiGroup.POST("/verify", authHandler.Login)

		// Friend link routes
		friendGroup := apiGroup.Group("/friend")
		{
			friendGroup.GET("/", friendLinkHandler.GetAllFriendLinks)
		}

		// RSS post routes
		rssGroup := apiGroup.Group("/rss")
		{
			rssGroup.GET("/", rssPostHandler.GetRssPosts)
		}

		// Action routes (requires JWT authentication)
		actionGroup := apiGroup.Group("/action")
		actionGroup.Use(middleware.JWTAuth())
		{
			friendActionGroup := actionGroup.Group("/friend")
			{
				friendActionGroup.POST("/", updataHandler.CreateFriendLink)
				friendActionGroup.PUT("/", updataHandler.EditFriendLink)
				friendActionGroup.DELETE("/", updataHandler.DeleteFriendLink)
			}
			actionGroup.POST("/rss", friendRssHandler.CreateFriendRss)
		}
	}
}
