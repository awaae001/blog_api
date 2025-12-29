package middleware

import (
	"net/http"
	"strings"

	"blog_api/src/config"
	"blog_api/src/model"
	"blog_api/src/service"

	"github.com/gin-gonic/gin"
)

// FingerprintAuth verifies a fingerprint token and stores fingerprint_id in context.
func FingerprintAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := config.GetConfig().Verify.Fingerprint.Secret
		if secret == "" {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "fingerprint secret is not configured"))
			c.Abort()
			return
		}

		token := extractFingerprintToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(401, "fingerprint token is required"))
			c.Abort()
			return
		}

		tokenService := service.NewFingerprintTokenService(secret)
		id, ok := tokenService.Verify(token)
		if !ok || id <= 0 {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(401, "invalid fingerprint token"))
			c.Abort()
			return
		}

		c.Set("fingerprint_id", id)
		c.Next()
	}
}

func extractFingerprintToken(c *gin.Context) string {
	if token := c.GetHeader("X-Fingerprint-Token"); token != "" {
		return token
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Fingerprint") {
			return parts[1]
		}
	}

	return ""
}
