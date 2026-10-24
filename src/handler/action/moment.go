package handlerAction

import (
	"blog_api/src/model"
	momentRepositories "blog_api/src/repositories/moment"
	"blog_api/src/service"
	botService "blog_api/src/service/bot"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MomentHandler handles moment related actions
type MomentHandler struct {
	DB *gorm.DB
}

// CreateMoment handles POST /api/action/moments request
func (h *MomentHandler) CreateMoment(c *gin.Context) {
	var req model.CreateMomentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid request body"})
		return
	}
	if req.GuildID != nil && *req.GuildID < 0 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid guild_id"})
		return
	}
	if req.ChannelID != nil && *req.ChannelID < 0 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid channel_id"})
		return
	}
	if req.MessageID != nil && *req.MessageID < 0 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid message_id"})
		return
	}

	if err := service.CreateMoment(h.DB, req); err != nil {
		log.Printf("[moments] create moment failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to create moment"})
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: nil})
}

// GetMoments handles GET /api/action/moments request
func (h *MomentHandler) GetMoments(c *gin.Context) {
	var req struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Status   string `form:"status"`
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

	resp, err := service.GetMomentsWithMedia(h.DB, req.Page, req.PageSize, req.Status, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to get moments"})
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: resp})
}

// DeleteMoment handles DELETE /api/action/moments/:id request
func (h *MomentHandler) DeleteMoment(c *gin.Context) {
	id, ok := parseRecordID(c, "invalid moment id")
	if !ok {
		return
	}

	if err := botService.DeleteMomentWithSync(h.DB.WithContext(c.Request.Context()), id); err != nil {
		writeRecordError(c, err, "moment not found", "failed to delete moment")
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: nil})
}

// UpdateMoment handles PUT /api/action/moments/:id request
func (h *MomentHandler) UpdateMoment(c *gin.Context) {
	id, ok := parseRecordID(c, "invalid moment id")
	if !ok {
		return
	}

	var req model.UpdateMomentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid request body"})
		return
	}

	if req.Status != nil {
		switch *req.Status {
		case "visible", "hidden", "deleted":
		default:
			c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid status"})
			return
		}
	}
	if req.GuildID != nil && *req.GuildID < 0 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid guild_id"})
		return
	}
	if req.ChannelID != nil && *req.ChannelID < 0 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid channel_id"})
		return
	}
	if req.MessageID != nil && *req.MessageID < 0 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid message_id"})
		return
	}

	if req == (model.UpdateMomentRequest{}) {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "no fields to update"})
		return
	}

	moment := model.Moment{ID: id}
	if err := momentRepositories.UpdateMoment(h.DB.WithContext(c.Request.Context()), &moment, req); err != nil {
		writeRecordError(c, err, "moment not found", "failed to update moment")
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: moment})
}

// DeleteMomentReaction handles DELETE /api/action/moments/:id/reactions request
func (h *MomentHandler) DeleteMomentReaction(c *gin.Context) {
	id, ok := parseRecordID(c, "invalid moment id")
	if !ok {
		return
	}

	reaction := c.Query("reaction")
	if reaction == "" {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "reaction query parameter is required"})
		return
	}

	if err := momentRepositories.ClearReactionsByType(h.DB, id, reaction); err != nil {
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to delete reactions"})
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: nil})
}
