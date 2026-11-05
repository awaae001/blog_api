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
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ApiResponse{
				Code:    http.StatusUnauthorized,
				Message: "未提供认证token",
			})
			return
		}

		// 解析 Bearer token
		scheme, token, ok := strings.Cut(authHeader, " ")
		if !ok || scheme != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ApiResponse{
				Code:    http.StatusUnauthorized,
				Message: "token格式错误",
			})
			return
		}

		// 验证 token
		claims, err := service.ValidateJWT(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ApiResponse{
				Code:    http.StatusUnauthorized,
				Message: "无效的token: " + err.Error(),
			})
			return
		}

		// 将用户信息存入上下文
		c.Set("username", claims.Username)
		c.Next()
	}
}
