package handlerAction

import (
	"blog_api/src/model"
	imageRepositories "blog_api/src/repositories/image"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ImageHandler handles image related requests
type ImageHandler struct {
	DB *gorm.DB
}

// GetImages handles GET /api/action/image request
func (h *ImageHandler) GetImages(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "无效的页面参数"})
		return
	}

	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "无效的页面大小参数"})
		return
	}

	if pageSize > 100 {
		pageSize = 100
	}

	opts := model.ImageQueryOptions{
		Page:     page,
		PageSize: pageSize,
		Status:   c.Query("status"),
		Name:     c.Query("search"),
	}

	resp, err := imageRepositories.QueryImages(h.DB, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "获取图片列表失败"})
		return
	}

	paginatedData := model.PaginatedResponse{
		Items:    resp.Images,
		Total:    int(resp.Total),
		Page:     page,
		PageSize: pageSize,
	}
	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: paginatedData})
}

// CreateImage handles POST /api/action/image request
func (h *ImageHandler) CreateImage(c *gin.Context) {
	var image model.Image
	if err := c.ShouldBindJSON(&image); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "无效的请求体: " + err.Error()})
		return
	}

	if image.URL == "" {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "图片 URL 不能为空"})
		return
	}
	// 忽略前端参数，默认状态为正常
	image.Status = "normal"

	if err := imageRepositories.CreateImage(h.DB, &image); err != nil {
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "创建图片配置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model.ApiResponse{Code: http.StatusCreated, Message: "success", Data: image})
}

// UpdateImage handles PUT /api/action/image/:id request
func (h *ImageHandler) UpdateImage(c *gin.Context) {
	id, ok := parseRecordID(c, "无效的图片 ID")
	if !ok {
		return
	}

	var req model.UpdateImageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "无效的请求体: " + err.Error()})
		return
	}
	if req.Status != nil {
		switch *req.Status {
		case "normal", "pause", "broken", "pending":
		default:
			c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "无效的状态参数"})
			return
		}
	}

	image := model.Image{ID: id}
	if err := imageRepositories.UpdateImage(h.DB.WithContext(c.Request.Context()), &image, req); err != nil {
		writeRecordError(c, err, "图片配置不存在", "更新图片配置失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: image})
}

// DeleteImage handles DELETE /api/action/image/:id request
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	id, ok := parseRecordID(c, "无效的图片 ID")
	if !ok {
		return
	}

	if err := imageRepositories.DeleteImage(h.DB.WithContext(c.Request.Context()), id); err != nil {
		writeRecordError(c, err, "图片配置不存在", "删除图片配置失败: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: nil})
}
