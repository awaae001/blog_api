package cmd

import (
	"blog_api/src/handler"
	handlerAction "blog_api/src/handler/action"
	authHandler "blog_api/src/handler/auth"
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

	friendLinkHandler := &handler.FriendLinkHandler{DB: db}
	rssPostHandler := &handler.RssPostHandler{DB: db}
	updataHandler := &handlerAction.UpdataHandler{DB: db}
	RssHandler := &handlerAction.FriendRssHandler{DB: db}
	authHandlerInstance := authHandler.NewAuthHandler()
	verifyHandler := &authHandler.VerifyHandler{DB: db}
	statusHandler := &handler.StatusHandler{
		DB:           db,
		StartTime:    startTime,
		DatabasePath: cfg.Data.Database.Path,
	}
	imageHandler := &handlerAction.ImageHandler{DB: db}
	resourceHandler := handlerAction.NewResourceHandler(ossService)
	imagePublicHandler := handler.NewImagePublicHandler(db)
	verifyPublicHandler := &authHandler.VerifyPublicHandler{}
	momentHandler := &handler.MomentHandler{DB: db}
	momentReactionHandler := &handler.MomentReactionHandler{DB: db}
	momentActionHandler := &handlerAction.MomentHandler{DB: db}
	mediaHandler := &handlerAction.MediaHandler{DB: db}
	configHandler := &handlerAction.ConfigHandler{}
	fingerprintHandler := authHandler.NewFingerprintHandler(db)
	systemHandler := &handlerAction.SystemHandler{}

	// API routes
	apiGroup := router.Group("/api")
	{
		if cfg.StateAPIMasterPassword != "" {
			registerStateRoutes(apiGroup, cfg.StateAPIMasterPassword)
		} else {
			log.Print("[state] STATE_API_MASTER_PASSWORD is empty; in-memory state API is disabled")
		}
		verifyGroup := apiGroup.Group("/verify")
		{
			verifyGroup.POST("/passwd", middleware.TurnstileVerify(), authHandlerInstance.Login)
			verifyGroup.POST("/email", middleware.RequirePublicAPI(model.PublicAPIEmail), middleware.AntiBotAuth(), verifyHandler.SendEmailCode)
			verifyGroup.POST("/turnstile", middleware.TurnstileVerify(), verifyHandler.IssueVerifyToken)
			verifyGroup.POST("/fingerprint", middleware.AntiBotAuth(), fingerprintHandler.CreateFingerprint)
		}
		publicGroup := apiGroup.Group("/public")
		{
			publicGroup.GET("/verify_conf", verifyPublicHandler.GetVerifyConfig)
			friendPublicGroup := publicGroup.Group("/friend")
			friendPublicGroup.Use(middleware.RequirePublicAPI(model.PublicAPIFriend))
			{
				friendPublicGroup.GET("/", friendLinkHandler.GetAllFriendLinks)
				friendPublicGroup.GET("/self", middleware.FriendLinkAuth(), friendLinkHandler.GetFriendLinkByEmailToken)
				friendPublicGroup.GET("/:id", friendLinkHandler.GetFriendLinkByID)
				friendPublicGroup.POST("", middleware.FriendLinkAuth(), updataHandler.CreateFriendLink)
				friendPublicGroup.PUT("/:id", middleware.FriendLinkAuth(), updataHandler.EditFriendLink)
				friendPublicGroup.DELETE("/:id", middleware.FriendLinkAuth(), updataHandler.DeleteOwnedFriendLink)
			}
			publicGroup.GET("/rss/", middleware.RequirePublicAPI(model.PublicAPIRSS), rssPostHandler.GetRssPosts)
			publicGroup.GET("/image/*id", middleware.RequirePublicAPI(model.PublicAPIImage), imagePublicHandler.GetImage)
			momentsPublicGroup := publicGroup.Group("/moments")
			momentsPublicGroup.Use(middleware.RequirePublicAPI(model.PublicAPIMoments))
			{
				momentsPublicGroup.GET("/", momentHandler.GetMoments)
				momentsPublicGroup.POST("/:id/reactions", middleware.AntiBotAuth(), middleware.FingerprintAuth(), momentReactionHandler.AddReaction)
				momentsPublicGroup.DELETE("/:id/reactions", middleware.AntiBotAuth(), middleware.FingerprintAuth(), momentReactionHandler.DeleteReaction)
			}
		}
		apiGroup.GET("/status", middleware.JWTAuth(), statusHandler.GetSystemStatus)

		actionGroup := apiGroup.Group("/action")
		actionGroup.Use(middleware.JWTAuth())
		{
			friendActionGroup := actionGroup.Group("/friend")
			{
				friendActionGroup.GET("", friendLinkHandler.GetFullFriendLinks)
				friendActionGroup.GET("/:id", friendLinkHandler.GetFullFriendLinkByID)
				friendActionGroup.POST("", updataHandler.CreateFriendLink)
				friendActionGroup.PUT("/:id", updataHandler.EditFriendLink)
				friendActionGroup.DELETE("/:id", updataHandler.DeleteFriendLink)
				friendActionGroup.POST("/recheck", friendLinkHandler.RecheckAllFriendLinks)
				friendActionGroup.GET("/recheck", friendLinkHandler.GetRecheckProgress)
				friendActionGroup.POST("/recheck/:id", friendLinkHandler.RecheckFriendLink)
				friendActionGroup.GET("/recheck/:id", friendLinkHandler.GetFriendLinkRecheckProgress)
			}
			rssActionGroup := actionGroup.Group("/rss")
			{
				rssActionGroup.GET("", RssHandler.GetRss)
				rssActionGroup.POST("", RssHandler.CreateRss)
				rssActionGroup.DELETE("/posts/:id", rssPostHandler.DeleteRssPost)
				rssActionGroup.PUT("/:id", RssHandler.EditRss)
				rssActionGroup.DELETE("/:id", RssHandler.DeleteFriendRss)
				rssActionGroup.POST("/:id/fetch", RssHandler.FetchRss)
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
				resourceActionGroup.DELETE("/local/*file_path", resourceHandler.DeleteResourceLocal)
				resourceActionGroup.DELETE("/oss/*file_path", resourceHandler.DeleteResourceOSS)
			}
			actionGroup.PUT("/config", configHandler.UpdateConfig)
			systemActionGroup := actionGroup.Group("/system")
			{
				systemActionGroup.POST("/restart", systemHandler.Restart)
			}
			momentsActionGroup := actionGroup.Group("/moments")
			{
				momentsActionGroup.GET("", momentActionHandler.GetMoments)
				momentsActionGroup.POST("", momentActionHandler.CreateMoment)
				momentsActionGroup.PUT("/:id", momentActionHandler.UpdateMoment)
				momentsActionGroup.DELETE("/:id", momentActionHandler.DeleteMoment)
				momentsActionGroup.DELETE("/:id/reactions", momentActionHandler.DeleteMomentReaction)
			}
			mediaActionGroup := actionGroup.Group("/moments/media")
			{
				mediaActionGroup.GET("", mediaHandler.GetMedia)
				mediaActionGroup.POST("", mediaHandler.CreateMedia)
				mediaActionGroup.PUT("/:id", mediaHandler.UpdateMedia)
				mediaActionGroup.DELETE("/:id", mediaHandler.DeleteMedia)
			}
		}
	}
}
