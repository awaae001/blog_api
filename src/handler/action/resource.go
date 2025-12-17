package handlerAction

import (
	"blog_api/src/model"
	"blog_api/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResourceHandler 封装了处理资源相关请求的逻辑。
type ResourceHandler struct {
	resourceService *service.ResourceService
}

// NewResourceHandler 创建一个新的 ResourceHandler 实例。
func NewResourceHandler(cfg *model.Config) *ResourceHandler {
	return &ResourceHandler{
		resourceService: service.NewResourceService(cfg),
	}
}

// UploadResource 处理文件上传请求。
func (h *ResourceHandler) UploadResource(c *gin.Context) {
	// 从表单中获取文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "获取文件失败: "+err.Error()))
		return
	}

	// 绑定表单字段
	var req model.UploadResourceReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "无效的表单参数: "+err.Error()))
		return
	}

	// 调用服务层保存文件
	_, urlPath, err := h.resourceService.SaveFile(file, req.Path, req.Overwrite)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{
		"url": urlPath,
	}))
}

// DeleteResource 处理文件删除请求。
func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	// 从 URL 通配符参数中获取文件路径
	filePath := c.Param("file_path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件路径不能为空"})
		return
	}

	// Gin 的通配符参数会包含一个前导斜杠，需要去掉
	filePath = filePath[1:]
	if err := h.resourceService.DeleteFile(filePath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件删除成功"})
}
