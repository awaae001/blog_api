package handler

import (
	"blog_api/src/config"
	"blog_api/src/middleware"
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

// GetMoments handles GET /api/moments request
func (h *MomentHandler) GetMoments(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid page parameter"})
		return
	}

	pageSizeStr := c.DefaultQuery("page_size", "10")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid page_size parameter"})
		return
	}

	if pageSize > 100 {
		pageSize = 100
	}

	tokenService := service.NewFingerprintTokenService(config.GetConfig().Verify.Fingerprint.Secret)
	var fingerprintID *int
	if id, ok := tokenService.Verify(middleware.ExtractFingerprintToken(c)); ok {
		fingerprintID = &id
	}

	resp, err := service.GetMomentsWithMedia(h.DB.WithContext(c.Request.Context()), page, pageSize, "visible", fingerprintID)
	if err != nil {
		if middleware.AbortIfCanceled(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to retrieve moments"})
		return
	}

	publicMoments := make([]model.PublicMomentWithMedia, len(resp.Moments))
	for i, moment := range resp.Moments {
		publicMoments[i] = model.PublicMomentWithMedia{
			ID:               moment.ID,
			Content:          moment.Content,
			Status:           moment.Status,
			MessageLink:      moment.MessageLink,
			CreatedAt:        moment.CreatedAt,
			UpdatedAt:        moment.UpdatedAt,
			Media:            moment.Media,
			Reactions:        moment.Reactions,
			SelectedReaction: moment.SelectedReaction,
		}
	}

	paginatedData := model.PaginatedResponse{
		Items:    publicMoments,
		Total:    int(resp.Total),
		Page:     page,
		PageSize: pageSize,
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: paginatedData})
}
