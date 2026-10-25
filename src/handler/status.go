package handler

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"blog_api/src/service"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StatusHandler handles system status requests.
type StatusHandler struct {
	DB           *gorm.DB
	StartTime    time.Time
	DatabasePath string
	StateStore   *service.StateStore
}

// GetSystemStatus handles the GET /api/status request.
func (h *StatusHandler) GetSystemStatus(c *gin.Context) {
	stats, err := repositories.GetSystemStats(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to retrieve system stats"})
		return
	}
	var databaseSize int64
	if info, err := os.Stat(h.DatabasePath); err == nil {
		databaseSize = info.Size()
	}

	var dataFolderSize int64
	filepath.Walk("data", func(_ string, info os.FileInfo, err error) error {
		if err == nil && info.Mode().IsRegular() {
			dataFolderSize += info.Size()
		}
		return nil
	})

	var kvCache model.KVCacheStatus
	if h.StateStore != nil {
		kvCache = h.StateStore.Stats()
	}

	now := time.Now()
	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: model.SystemStatus{
		Uptime:              now.Sub(h.StartTime).Round(time.Second).String(),
		StatusData:          stats,
		DatabaseSizeBytes:   databaseSize,
		DataFolderSizeBytes: dataFolderSize,
		Time:                now.Unix(),
		KVCache:             kvCache,
	}})
}
