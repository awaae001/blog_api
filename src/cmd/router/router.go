package cmd

import (
	"blog_api/src/handler"
	handlerAction "blog_api/src/handler/action"
	"blog_api/src/middleware"
	"blog_api/src/model"
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	authHandler := handler.NewAuthHandler()

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes
		api.POST("/verify", authHandler.Login)

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
		// Action routes for friend links (requires JWT authentication)
		action := api.Group("/action")
		action.Use(middleware.JWTAuth())
		{
			friendAction := action.Group("/friend")
			{
				friendAction.POST("/", updataHandler.CreateFriendLink)
				friendAction.PUT("/", updataHandler.EditFriendLink)
				friendAction.DELETE("/", updataHandler.DeleteFriendLink)
			}
		}
	}
	//
	router.NoRoute(func(c *gin.Context) {
		dir := "data"
		reqPath := c.Request.URL.Path

		// 1. Prevent directory traversal
		if strings.Contains(reqPath, "..") {
			c.String(http.StatusBadRequest, "Bad Request")
			return
		}

		// 2. Check against general excluded paths from config
		for _, excludedPath := range cfg.Safe.ExcludePaths {
			if !strings.HasPrefix(excludedPath, "/") {
				excludedPath = "/" + excludedPath
			}
			if strings.HasPrefix(reqPath, excludedPath) {
				c.String(http.StatusForbidden, "Forbidden")
				return
			}
		}

		// 3. Check against database file
		if cfg.Data.Database.Path != "" {
			dbFileName := filepath.Base(cfg.Data.Database.Path)
			if reqPath == "/"+dbFileName {
				c.String(http.StatusForbidden, "Forbidden")
				return
			}
		}

		fsPath := filepath.Join(dir, reqPath)
		originalPathIsDir := false
		if info, err := os.Stat(fsPath); err == nil && info.IsDir() {
			originalPathIsDir = true
			fsPath = filepath.Join(fsPath, "index.html")
		}

		// Check if the file exists and handle errors
		if _, err := os.Stat(fsPath); os.IsNotExist(err) {
			if originalPathIsDir {
				c.String(http.StatusBadRequest, "Bad Request: index.html not found in directory")
			} else {
				c.String(http.StatusNotFound, "Not Found")
			}
			return
		}

		// Serve the file
		c.File(fsPath)
	})
	return router
}
