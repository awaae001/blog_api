package handlerAction

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MediaHandler handles media related actions
type MediaHandler struct {
	DB *gorm.DB
}

// NewMediaHandler creates a new media action handler
func NewMediaHandler(db *gorm.DB) *MediaHandler {
	return &MediaHandler{DB: db}
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
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid query parameters"))
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	media, total, err := repositories.QueryMedia(h.DB, req.Page, req.PageSize, req.MomentID, req.MediaType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to get media"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(&model.QueryMediaResponse{
		Media: media,
		Total: total,
	}))
}
