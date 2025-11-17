package cmd

import (
	"blog_api/src/model"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func staticFileHandler(cfg *model.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
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
				return
			}

			c.String(http.StatusNotFound, "Not Found")
			return
		}

		// Serve the file
		c.File(fsPath)
	}
}
