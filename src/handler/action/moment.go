package handlerAction

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"blog_api/src/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MomentHandler handles moment related actions
type MomentHandler struct {
	DB *gorm.DB
}

// NewMomentHandler creates a new moment action handler
func NewMomentHandler(db *gorm.DB) *MomentHandler {
	return &MomentHandler{DB: db}
}

// CreateMoment handles POST /api/action/moments request
func (h *MomentHandler) CreateMoment(c *gin.Context) {
	var req model.CreateMomentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid request body"))
		return
	}

	if err := service.CreateMoment(h.DB, req); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to create moment"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}

// GetMoments handles GET /api/action/moments request
func (h *MomentHandler) GetMoments(c *gin.Context) {
	var req struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Status   string `form:"status"`
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

	resp, err := service.GetMomentsWithMedia(h.DB, req.Page, req.PageSize, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to get moments"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(resp))
}

// DeleteMoment handles DELETE /api/action/moments/:id request
func (h *MomentHandler) DeleteMoment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid moment id"))
		return
	}

	if err := repositories.DeleteMoment(h.DB, id); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to delete moment"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}
