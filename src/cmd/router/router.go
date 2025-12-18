package cmd

import (
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
	router.NoRoute(staticFileHandler(cfg))
	return router
}
