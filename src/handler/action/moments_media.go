package handlerAction

import (
	"blog_api/src/handler/response"
	"blog_api/src/model"
	momentRepositories "blog_api/src/repositories/moment"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MediaHandler handles media related actions
type MediaHandler struct {
	DB *gorm.DB
}

// CreateMedia handles POST /api/action/moments/media request
func (h *MediaHandler) CreateMedia(c *gin.Context) {
	var req model.CreateMomentMediaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid request body"})
		return
	}

	if req.MediaType == "" {
		req.MediaType = "image"
	}

	media := model.MomentMedia{
		MomentID:  req.MomentID,
		Name:      req.Name,
		MediaURL:  req.MediaURL,
		MediaType: req.MediaType,
		IsLocal:   req.IsLocal,
		IsDeleted: 0,
	}

	if response.Error(c, h.DB.Create(&media).Error, "failed to create media") {
		return
	}

	c.JSON(http.StatusCreated, model.ApiResponse{Code: http.StatusCreated, Message: "success", Data: media})
}

// DeleteMedia handles DELETE /api/action/moments/media/:id request
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	id, ok := parseRecordID(c, "invalid media id")
	if !ok {
		return
	}

	if response.RecordError(c, momentRepositories.DeleteMomentMedia(h.DB.WithContext(c.Request.Context()), id), "media not found", "failed to delete media") {
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: nil})
}

// UpdateMedia handles PUT /api/action/moments/media/:id request
func (h *MediaHandler) UpdateMedia(c *gin.Context) {
	id, ok := parseRecordID(c, "invalid media id")
	if !ok {
		return
	}

	var req model.UpdateMomentMediaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid request body"})
		return
	}

	media := model.MomentMedia{ID: id}
	if response.RecordError(c, momentRepositories.UpdateMomentMedia(h.DB.WithContext(c.Request.Context()), &media, req), "media not found", "failed to update media") {
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: media})
}

// GetMedia handles GET /api/action/moments/media request
func (h *MediaHandler) GetMedia(c *gin.Context) {
	var req struct {
		Page      int    `form:"page"`
		PageSize  int    `form:"page_size"`
		MomentID  int    `form:"moment_id"`
		MediaType string `form:"type"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid query parameters"})
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	media, total, err := momentRepositories.QueryMedia(h.DB, req.Page, req.PageSize, req.MomentID, req.MediaType)
	if response.Error(c, err, "failed to get media") {
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: &model.QueryMediaResponse{
		Media: media,
		Total: total,
	}})
}
