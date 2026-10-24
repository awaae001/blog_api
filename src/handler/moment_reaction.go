package handler

import (
	"blog_api/src/middleware"
	"blog_api/src/model"
	momentRepositories "blog_api/src/repositories/moment"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MomentReactionHandler handles reactions to moments.
type MomentReactionHandler struct {
	DB *gorm.DB
}

// AddReaction handles POST /api/public/moments/:id/reactions request.
func (h *MomentReactionHandler) AddReaction(c *gin.Context) {
	reaction, ok := bindReaction(c)
	if !ok {
		return
	}
	reaction.CreatedAt = time.Now().Unix()
	if err := momentRepositories.CreateMomentReaction(h.DB.WithContext(c.Request.Context()), reaction); err != nil {
		if middleware.AbortIfCanceled(c, err) {
			return
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			c.JSON(http.StatusConflict, model.ApiResponse{Code: 409, Message: "reaction already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to add reaction"})
		return
	}
	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: nil})
}

// DeleteReaction handles DELETE /api/public/moments/:id/reactions request.
func (h *MomentReactionHandler) DeleteReaction(c *gin.Context) {
	reaction, ok := bindReaction(c)
	if !ok {
		return
	}
	if err := momentRepositories.DeleteMomentReaction(h.DB.WithContext(c.Request.Context()), reaction.MomentID, reaction.FingerprintID, reaction.Reaction); err != nil {
		if middleware.AbortIfCanceled(c, err) {
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, model.ApiResponse{Code: 404, Message: "reaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to delete reaction"})
		return
	}
	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: nil})
}

func bindReaction(c *gin.Context) (*model.MomentReaction, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid moment id"})
		return nil, false
	}
	var req model.MomentReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid request body"})
		return nil, false
	}
	switch req.Reaction {
	case "👍", "👎", "❤", "👀", "💩":
	default:
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid reaction"})
		return nil, false
	}
	fingerprintID, ok := c.Get("fingerprint_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ApiResponse{Code: 401, Message: "fingerprint token is required"})
		return nil, false
	}
	return &model.MomentReaction{
		MomentID:      id,
		FingerprintID: fingerprintID.(int),
		Reaction:      req.Reaction,
	}, true
}
