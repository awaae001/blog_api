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
		token := ExtractFingerprintToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ApiResponse{Code: 401, Message: "fingerprint token is required"})
			return
		}

		tokenService := service.NewFingerprintTokenService(secret)
		id, ok := tokenService.Verify(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ApiResponse{Code: 401, Message: "invalid fingerprint token"})
			return
		}

		c.Set("fingerprint_id", id)
		c.Next()
	}
}

// ExtractFingerprintToken reads the fingerprint header or Authorization scheme.
func ExtractFingerprintToken(c *gin.Context) string {
	if token := c.GetHeader("X-Fingerprint-Token"); token != "" {
		return token
	}

	scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
	if ok && strings.EqualFold(scheme, "Fingerprint") {
		return token
	}

	return ""
}
