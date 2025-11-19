package handler

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StatusHandler handles system status requests.
type StatusHandler struct {
	DB        *sql.DB
	StartTime time.Time
}

// NewStatusHandler creates a new status handler.
func NewStatusHandler(db *sql.DB, startTime time.Time) *StatusHandler {
	return &StatusHandler{DB: db, StartTime: startTime}
}

// GetSystemStatus handles the GET /api/status request.
func (h *StatusHandler) GetSystemStatus(c *gin.Context) {
	// Get counts from repositories
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
		Time:       time.Now(),
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(systemStatus))
}
