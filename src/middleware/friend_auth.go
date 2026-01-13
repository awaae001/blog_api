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
	authService := service.NewAuthService()

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(401, "authorization token is required"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(401, "token format is invalid"))
			c.Abort()
			return
		}

		token := parts[1]
		if claims, err := authService.ValidateJWT(token); err == nil {
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

		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(401, "invalid token"))
		c.Abort()
	}
}
