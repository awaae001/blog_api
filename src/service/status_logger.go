package service

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gorm.io/gorm"
)

// LogSystemStatus fetches system status and writes it to a log file.
func LogSystemStatus(db *gorm.DB) {

	logDir := "data/log"
	logFile := filepath.Join(logDir, "system_status.log")

	log.Println("[StatusLogger] Running system status logger...")

	// Get runtime memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	goroutineCount := runtime.NumGoroutine()
	dbStats, _ := repositories.GetSystemStats(db)
	statusLog := model.SystemStatusLog{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		GoroutineCount: goroutineCount,
		MemStats:       memStats,
		DbStats:        dbStats,
	}

	logData, _ := json.Marshal(statusLog)

	// Ensure the log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("[StatusLogger] Failed to create log directory '%s': %v", logDir, err)
		return
	}

	// Open the log file for appending
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[StatusLogger] Failed to open log file '%s': %v", logFile, err)
		return
	}
	defer file.Close()

	// Write the log entry
	if _, err := fmt.Fprintln(file, string(logData)); err != nil {
		log.Printf("[StatusLogger] Failed to write to log file '%s': %v", logFile, err)
	}
}
