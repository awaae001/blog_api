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
		info, err := os.Stat(fsPath)

		// If file doesn't exist, check if it's a SPA route
		if os.IsNotExist(err) {
			// If the path is under /panel/, it's a SPA route, serve index.html
			if strings.HasPrefix(reqPath, "/panel/") {
				spaIndex := filepath.Join(dir, "panel", "index.html")
				if _, err := os.Stat(spaIndex); err == nil {
					c.File(spaIndex)
					return
				}
			}
			// Otherwise, it's a 404
			c.String(http.StatusNotFound, "Not Found")
			return
		}

		// If it's a directory, serve index.html from it
		if info.IsDir() {
			indexPath := filepath.Join(fsPath, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				c.File(indexPath)
				return
			}
			c.String(http.StatusForbidden, "Directory listing is not allowed")
			return
		}

		// Serve the file
		c.File(fsPath)
	}
}
