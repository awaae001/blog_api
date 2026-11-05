package middleware

import (
	"net/http"
	"strings"

	"blog_api/src/model"
	"blog_api/src/service"

	"github.com/gin-gonic/gin"
)

// FriendLinkAuth allows either admin JWTs or one-time email tokens.
func FriendLinkAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ApiResponse{Code: 401, Message: "authorization token is required"})
			return
		}

		scheme, token, ok := strings.Cut(authHeader, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ApiResponse{Code: 401, Message: "token format is invalid"})
			return
		}
		if claims, err := service.ValidateJWT(token); err == nil {
			c.Set("username", claims.Username)
			c.Set("auth_type", "jwt")
			c.Next()
			return
		}

		if email, ok := service.ValidateEmailToken(token); ok {
			c.Set("auth_email", email)
			c.Set("auth_type", "email")
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, model.ApiResponse{Code: 401, Message: "invalid token"})
	}
}
