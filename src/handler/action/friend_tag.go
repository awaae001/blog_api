package handlerAction

import (
	"blog_api/src/handler/response"
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

type FriendTagHandler struct{ DB *gorm.DB }

func friendTagError(c *gin.Context, err error, missing, failure string) bool {
	if errors.Is(err, friendsRepositories.ErrInvalidTagIDs) {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: err.Error()})
		return true
	}
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
		c.JSON(http.StatusConflict, model.ApiResponse{Code: 409, Message: "tag name already exists"})
		return true
	}
	return response.RecordError(c, err, missing, failure)
}

func friendTagID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid ID"})
		return 0, false
	}
	return id, true
}

func (h *FriendTagHandler) CreateTag(c *gin.Context) {
	var req model.CreateFriendTagReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "tag name is required"})
		return
	}
	tag := model.FriendTag{Name: strings.TrimSpace(req.Name), Color: req.Color, SortOrder: req.SortOrder}
	if friendTagError(c, friendsRepositories.CreateFriendTag(h.DB.WithContext(c.Request.Context()), &tag), "tag not found", "failed to create tag") {
		return
	}
	c.JSON(http.StatusOK, model.ApiResponse{Code: 200, Message: "success", Data: tag.ToDTO(true)})
}

func (h *FriendTagHandler) UpdateTag(c *gin.Context) {
	id, ok := friendTagID(c)
	if !ok {
		return
	}
	var req model.UpdateFriendTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid request body"})
		return
	}
	updates := make(map[string]any)
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "tag name is required"})
			return
		}
		updates["name"] = name
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "no fields provided for update"})
		return
	}
	tag, err := friendsRepositories.UpdateFriendTag(h.DB.WithContext(c.Request.Context()), id, updates)
	if friendTagError(c, err, "tag not found", "failed to update tag") {
		return
	}
	c.JSON(http.StatusOK, model.ApiResponse{Code: 200, Message: "success", Data: tag.ToDTO(true)})
}

func (h *FriendTagHandler) DeleteTag(c *gin.Context) {
	id, ok := friendTagID(c)
	if !ok {
		return
	}
	if friendTagError(c, friendsRepositories.DeleteFriendTag(h.DB.WithContext(c.Request.Context()), id), "tag not found", "failed to delete tag") {
		return
	}
	c.JSON(http.StatusOK, model.ApiResponse{Code: 200, Message: "success"})
}

func (h *FriendTagHandler) ReplaceTags(c *gin.Context) {
	id, ok := friendTagID(c)
	if !ok {
		return
	}
	var req model.ReplaceFriendTagsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "tag_ids must be an array of positive IDs; use [] to clear"})
		return
	}
	tags, err := friendsRepositories.ReplaceFriendTags(h.DB.WithContext(c.Request.Context()), id, req.TagIDs)
	if friendTagError(c, err, "friend link not found", "failed to replace friend tags") {
		return
	}
	dtos := make([]model.FriendTagDTO, 0, len(tags))
	for _, tag := range tags {
		dtos = append(dtos, tag.ToDTO(true))
	}
	c.JSON(http.StatusOK, model.ApiResponse{Code: 200, Message: "success", Data: dtos})
}
