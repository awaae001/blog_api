package handler

import (
	"blog_api/src/model"
	"blog_api/src/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MomentHandler handles moment related requests
type MomentHandler struct {
	DB *gorm.DB
}

// NewMomentHandler creates a new moment handler
func NewMomentHandler(db *gorm.DB) *MomentHandler {
	return &MomentHandler{DB: db}
}

// GetMoments handles GET /api/moments request
func (h *MomentHandler) GetMoments(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid page parameter"))
		return
	}

	pageSizeStr := c.DefaultQuery("page_size", "10")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid page_size parameter"))
		return
	}

	if pageSize > 100 {
		pageSize = 100
	}

	resp, err := service.GetMomentsWithMedia(h.DB, page, pageSize, "visible")
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to retrieve moments"))
		return
	}

	publicMoments := make([]model.PublicMomentWithMedia, len(resp.Moments))
	for i, moment := range resp.Moments {
		publicMoments[i] = model.PublicMomentWithMedia{
			ID:          moment.ID,
			Content:     moment.Content,
			Status:      moment.Status,
			MessageLink: moment.MessageLink,
			CreatedAt:   moment.CreatedAt,
			UpdatedAt:   moment.UpdatedAt,
			Media:       moment.Media,
		}
	}

	paginatedData := model.PaginatedResponse{
		Items:    publicMoments,
		Total:    int(resp.Total),
		Page:     page,
		PageSize: pageSize,
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(paginatedData))
}
