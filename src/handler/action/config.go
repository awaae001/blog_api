package handlerAction

import (
	"blog_api/src/config"
	"blog_api/src/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 处理与配置相关的请求
type ConfigHandler struct{}

// UpdateConfig 处理 PATCH /api/action/config，请求体只需包含变化的字段。
func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	var req model.UpdateConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: http.StatusBadRequest, Message: "无效的配置补丁: " + err.Error()})
		return
	}
	result, err := config.UpdateAndSaveConfig(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: http.StatusInternalServerError, Message: "更新配置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: result})
}
