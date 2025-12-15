package cmd

import (
	"blog_api/src/handler"
	handlerAction "blog_api/src/handler/action"
	"blog_api/src/middleware"
	"blog_api/src/model"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter initializes and configures the Gin router
func SetupRouter(db *gorm.DB, cfg *model.Config, startTime time.Time) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Safe.CorsAllowHostlist,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	registerRoutes(router, db, startTime)
	pprof.Register(router)

	// Serve static files and handle SPA
	router.NoRoute(staticFileHandler(cfg))
	return router
}

func registerRoutes(router *gin.Engine, db *gorm.DB, startTime time.Time) {
	// Initialize handlers
	friendLinkHandler := handler.NewFriendLinkHandler(db)
	rssPostHandler := handler.NewRssPostHandler(db)
	updataHandler := handlerAction.NewUpdataHandler(db)
	RssHandler := handlerAction.NewRssHandler(db)
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
				friendActionGroup.GET("/", friendLinkHandler.GetFullFriendLinks)
				friendActionGroup.POST("/", updataHandler.CreateFriendLink)
				friendActionGroup.PUT("/", updataHandler.EditFriendLink)
				friendActionGroup.DELETE("/", updataHandler.DeleteFriendLink)
			}
			rssActionGroup := actionGroup.Group("/rss")
			{
				rssActionGroup.GET("/", RssHandler.GetRss)
				rssActionGroup.POST("/", RssHandler.CreateRss)
				rssActionGroup.PUT("/", RssHandler.EditRss)
				rssActionGroup.DELETE("/", RssHandler.DeleteFriendRss)
			}
		}
	}
}
