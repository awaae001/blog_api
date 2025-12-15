package handler

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StatusHandler handles system status requests.
type StatusHandler struct {
	DB        *gorm.DB
	StartTime time.Time
}

// NewStatusHandler creates a new status handler.
func NewStatusHandler(db *gorm.DB, startTime time.Time) *StatusHandler {
	return &StatusHandler{DB: db, StartTime: startTime}
}

// GetSystemStatus handles the GET /api/status request.
func (h *StatusHandler) GetSystemStatus(c *gin.Context) {
	// Get all stats from repositories, including chart data
	stats, err := repositories.GetSystemStats(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to retrieve system stats"))
		return
	}

	// Calculate uptime
	uptime := time.Since(h.StartTime)

	// Build response
	systemStatus := model.SystemStatus{
		Uptime:     fmt.Sprintf("%v", uptime.Round(time.Second)),
		StatusData: stats,
		Time:       time.Now().Unix(),
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(systemStatus))
}
