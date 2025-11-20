package handler

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RssPostHandler handles RSS post related requests
type RssPostHandler struct {
	DB *sql.DB
}

// NewRssPostHandler creates a new RSS post handler
func NewRssPostHandler(db *sql.DB) *RssPostHandler {
	return &RssPostHandler{DB: db}
}

// GetRssPosts handles GET /api/rss request
// Query parameters:
//   - friend_link_id: filter by friend_link_id (optional)
//   - page: for pagination (optional, default: 1)
//   - page_size: for pagination (optional, default: 10)
func (h *RssPostHandler) GetRssPosts(c *gin.Context) {
	friendLinkIDStr := c.Query("friend_link_id")

	if friendLinkIDStr != "" {
		// Handle request with friend_link_id
		friendLinkID, err := strconv.Atoi(friendLinkIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid friend_link_id parameter"))
			return
		}

		posts, err := repositories.GetPostsByFriendLinkID(h.DB, friendLinkID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to retrieve posts"))
			return
		}
		c.JSON(http.StatusOK, model.NewSuccessResponse(posts))
	} else {
		// Handle request without friend_link_id (with pagination)
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			page = 1
		}

		pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		if err != nil || pageSize <= 0 {
			pageSize = 10
		}

		posts, total, err := repositories.GetAllPosts(h.DB, page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to retrieve posts"))
			return
		}

		c.JSON(http.StatusOK, model.NewSuccessResponse(&model.PaginatedResponse{
			Items:    posts,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}))
	}
}
