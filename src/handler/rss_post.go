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

// GetAllPostsByFriendLinkID handles GET /api/rss request
// Query parameters:
//   - friend_link_id: filter by friend_link_id (required)
func (h *RssPostHandler) GetAllPostsByFriendLinkID(c *gin.Context) {
	friendLinkIDStr := c.Query("friend_link_id")
	if friendLinkIDStr == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "friend_link_id is required"))
		return
	}

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
}
