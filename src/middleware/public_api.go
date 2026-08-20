package middleware

import (
	"blog_api/src/config"
	"blog_api/src/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequirePublicAPI hides a public API unless its feature key is enabled.
// The configuration is read for every request so settings updates take effect
// without rebuilding the router or restarting the process.
func RequirePublicAPI(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.GetConfig().Safe.IsPublicAPIEnabled(feature) {
			c.AbortWithStatusJSON(http.StatusNotFound, model.NewErrorResponse(http.StatusNotFound, "not found"))
			return
		}
		c.Next()
	}
}
