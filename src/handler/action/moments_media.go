package handlerAction

import (
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

	if req.MomentID <= 0 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid moment_id"})
		return
	}
	if req.MediaURL == "" {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "media_url is required"})
		return
	}

	if req.MediaType == "" {
		req.MediaType = "image"
	}
	if req.MediaType != "image" && req.MediaType != "video" {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid media_type"})
		return
	}
	if req.IsLocal != 0 && req.IsLocal != 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid is_local"})
		return
	}

	media := model.MomentMedia{
		MomentID:  req.MomentID,
		Name:      req.Name,
		MediaURL:  req.MediaURL,
		MediaType: req.MediaType,
		IsLocal:   req.IsLocal,
		IsDeleted: 0,
	}

	if err := momentRepositories.CreateMomentMedia(h.DB, &media); err != nil {
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to create media"})
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

	if err := momentRepositories.DeleteMomentMedia(h.DB.WithContext(c.Request.Context()), id); err != nil {
		writeRecordError(c, err, "media not found", "failed to delete media")
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

	if req.MomentID != nil && *req.MomentID <= 0 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid moment_id"})
		return
	}
	if req.MediaType != nil && *req.MediaType != "image" && *req.MediaType != "video" {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid media_type"})
		return
	}
	if req.IsLocal != nil && *req.IsLocal != 0 && *req.IsLocal != 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid is_local"})
		return
	}

	media := model.MomentMedia{ID: id}
	if err := momentRepositories.UpdateMomentMedia(h.DB.WithContext(c.Request.Context()), &media, req); err != nil {
		writeRecordError(c, err, "media not found", "failed to update media")
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
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to get media"})
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: &model.QueryMediaResponse{
		Media: media,
		Total: total,
	}})
}
