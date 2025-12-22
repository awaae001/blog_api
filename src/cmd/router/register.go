package cmd

import (
	"blog_api/src/handler"
	handlerAction "blog_api/src/handler/action"
	"blog_api/src/middleware"
	"blog_api/src/model"
	"blog_api/src/service/oss"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerRoutes(router *gin.Engine, db *gorm.DB, cfg *model.Config, startTime time.Time) {
	ossService, err := oss.NewOSSService()
	if err != nil {
		// 记录错误但不中断启动，因为 OSS 可能不是必须的
		log.Printf("Failed to initialize OSS service: %v", err)
	}

	friendLinkHandler := handler.NewFriendLinkHandler(db)
	rssPostHandler := handler.NewRssPostHandler(db)
	updataHandler := handlerAction.NewUpdataHandler(db)
	RssHandler := handlerAction.NewRssHandler(db)
	authHandler := handler.NewAuthHandler()
	statusHandler := handler.NewStatusHandler(db, startTime)
	imageHandler := handlerAction.NewImageHandler(db)
	resourceHandler := handlerAction.NewResourceHandler(cfg, ossService)
	imagePublicHandler := handler.NewImagePublicHandler(db)
	momentHandler := handler.NewMomentHandler(db)
	momentActionHandler := handlerAction.NewMomentHandler(db)
	configHandler := handlerAction.NewConfigHandler()

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
			publicGroup.GET("/friend/", friendLinkHandler.GetAllFriendLinks)
			publicGroup.GET("/rss/", rssPostHandler.GetRssPosts)
			publicGroup.GET("/image/*id", imagePublicHandler.GetImage)
			publicGroup.GET("/moments/", momentHandler.GetMoments)
		}
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
				resourceActionGroup.GET("/*file_path", resourceHandler.GetResource)
				resourceActionGroup.POST("/local", resourceHandler.UploadResourceLocal)
				resourceActionGroup.POST("/oss", resourceHandler.UploadResourceOSS)
				resourceActionGroup.DELETE("/*file_path", resourceHandler.DeleteResource)
			}
			actionGroup.PUT("/config", configHandler.UpdateConfig)
			momentsActionGroup := actionGroup.Group("/moments")
			{
				momentsActionGroup.GET("", momentActionHandler.GetMoments)
				momentsActionGroup.POST("", momentActionHandler.CreateMoment)
				momentsActionGroup.DELETE("/:id", momentActionHandler.DeleteMoment)
			}
		}
	}
}
