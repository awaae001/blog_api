package cmd

import (
	"blog_api/src/handler"
	handlerAction "blog_api/src/handler/action"
	"blog_api/src/middleware"
	"blog_api/src/model"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NotImplemented is a handler for features that are not yet implemented.
func NotImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Not Implemented",
	})
}

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

	registerRoutes(router, db, cfg, startTime)
	pprof.Register(router)

	// Serve static files and handle SPA
	router.NoRoute(staticFileHandler(cfg))
	return router
}

func registerRoutes(router *gin.Engine, db *gorm.DB, cfg *model.Config, startTime time.Time) {
	// Initialize handlers
	friendLinkHandler := handler.NewFriendLinkHandler(db)
	rssPostHandler := handler.NewRssPostHandler(db)
	updataHandler := handlerAction.NewUpdataHandler(db)
	RssHandler := handlerAction.NewRssHandler(db)
	authHandler := handler.NewAuthHandler()
	statusHandler := handler.NewStatusHandler(db, startTime)
	imageHandler := handlerAction.NewImageHandler(db)
	resourceHandler := handlerAction.NewResourceHandler(cfg)
	imagePublicHandler := handler.NewImagePublicHandler(db)

	// API routes
	apiGroup := router.Group("/api")
	{

		// Authentication routes
		verifyGroup := apiGroup.Group("/verify")
		{
			verifyGroup.POST("/passwd", authHandler.Login)
			verifyGroup.POST("/email", NotImplemented)
		}
		publicGroup := apiGroup.Group("/public")
		{
			friendGroup := publicGroup.Group("/friend")
			{
				friendGroup.GET("/", friendLinkHandler.GetAllFriendLinks)
			}
			rssGroup := publicGroup.Group("/rss")
			{
				rssGroup.GET("/", rssPostHandler.GetRssPosts)
			}
			imageGroup := publicGroup.Group("/image")
			{
				imageGroup.GET("/*id", imagePublicHandler.GetImage)
			}
		}
		// Status router (protected)
		apiGroup.GET("/status", middleware.JWTAuth(), statusHandler.GetSystemStatus)

		actionGroup := apiGroup.Group("/action")
		actionGroup.Use(middleware.JWTAuth())
		{
			friendActionGroup := actionGroup.Group("/friend")
			{
				friendActionGroup.GET("", friendLinkHandler.GetFullFriendLinks)
				friendActionGroup.POST("", updataHandler.CreateFriendLink)
				friendActionGroup.PUT("/:id", updataHandler.EditFriendLink)
				friendActionGroup.DELETE("/:id", updataHandler.DeleteFriendLink)
			}
			rssActionGroup := actionGroup.Group("/rss")
			{
				rssActionGroup.GET("", RssHandler.GetRss)
				rssActionGroup.POST("", RssHandler.CreateRss)
				rssActionGroup.PUT("/:id", RssHandler.EditRss)
				rssActionGroup.DELETE("/:id", RssHandler.DeleteFriendRss)
			}
			imageActionGroup := actionGroup.Group("/image")
			{
				imageActionGroup.GET("", imageHandler.GetImages)
				imageActionGroup.POST("", imageHandler.CreateImage)
				imageActionGroup.PUT("/:id", imageHandler.UpdateImage)
				imageActionGroup.DELETE("/:id", imageHandler.DeleteImage)
			}
			resourceActionGroup := actionGroup.Group("/resource")
			{
				resourceActionGroup.POST("", resourceHandler.UploadResource)
				resourceActionGroup.DELETE("/*file_path", resourceHandler.DeleteResource)
			}
		}
	}
}
