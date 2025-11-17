package cmd

import (
	"blog_api/src/handler"
	handlerAction "blog_api/src/handler/action"
	"blog_api/src/middleware"
	"blog_api/src/model"
	"database/sql"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the Gin router
func SetupRouter(db *sql.DB, cfg *model.Config) *gin.Engine {
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

	registerRoutes(router, db)

	// Serve static files and handle SPA
	router.NoRoute(staticFileHandler(cfg))

	return router
}

func registerRoutes(router *gin.Engine, db *sql.DB) {
	// Initialize handlers
	friendLinkHandler := handler.NewFriendLinkHandler(db)
	rssPostHandler := handler.NewRssPostHandler(db)
	updataHandler := handlerAction.NewUpdataHandler(db)
	authHandler := handler.NewAuthHandler()

	// Serve SPA for /panel
	panelHandler := func(c *gin.Context) {
		c.File(filepath.Join("data", "panel", "index.html"))
	}
	router.GET("/panel", panelHandler)
	router.GET("/panel/*path", panelHandler)

	// API routes
	apiGroup := router.Group("/api")
	{
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
			rssGroup.GET("/", rssPostHandler.GetAllPostsByFriendLinkID)
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
		}
	}
}
