package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"blog_api/src/model"

	"github.com/gin-gonic/gin"
)

// StateMasterAuth authenticates server-to-server state API requests.
//
// Requests must carry the configured password as an Authorization Bearer token.
func StateMasterAuth(password string) gin.HandlerFunc {
	expected := sha256.Sum256([]byte(password))
	return func(c *gin.Context) {
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		actual := sha256.Sum256([]byte(token))
		if !ok || token == "" || subtle.ConstantTimeCompare(actual[:], expected[:]) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ApiResponse{Code: http.StatusUnauthorized, Message: "unauthorized"})
			return
		}
		c.Next()
	}
}
