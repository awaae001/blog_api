package middleware

import (
	"net/http"
	"strings"

	"blog_api/src/model"
	"blog_api/src/service"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	authService := service.NewAuthService()

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.ApiResponse{
				Code:    http.StatusUnauthorized,
				Message: "未提供认证token",
				Data:    nil,
			})
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, model.ApiResponse{
				Code:    http.StatusUnauthorized,
				Message: "token格式错误",
				Data:    nil,
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证 token
		claims, err := authService.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.ApiResponse{
				Code:    http.StatusUnauthorized,
				Message: "无效的token: " + err.Error(),
				Data:    nil,
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("username", claims.Username)
		c.Next()
	}
}
