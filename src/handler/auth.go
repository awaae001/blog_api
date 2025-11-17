package handler

import (
	"net/http"

	"blog_api/src/model"
	"blog_api/src/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

// Login 处理登录请求
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// 验证用户名和密码
	if !h.authService.ValidateCredentials(req.Username, req.Password) {
		c.JSON(http.StatusUnauthorized, model.ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "用户名或密码错误",
			Data:    nil,
		})
		return
	}

	// 生成 JWT token
	token, expiresAt, err := h.authService.GenerateJWT(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "生成token失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{
		Code:    http.StatusOK,
		Message: "登录成功",
		Data: model.LoginResponse{
			Token:     token,
			ExpiresAt: expiresAt.Format("2006-01-02 15:04:05"),
		},
	})
}
