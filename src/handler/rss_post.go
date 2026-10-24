package handler

import (
	"blog_api/src/middleware"
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RssPostHandler handles RSS post related requests
type RssPostHandler struct {
	DB *gorm.DB
}

// DeleteRssPost handles deletion of one persisted RSS article.
func (h *RssPostHandler) DeleteRssPost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: http.StatusBadRequest, Message: "无效的 RSS 文章 ID"})
		return
	}

	rowsAffected, err := friendsRepositories.DeleteRssPostByID(h.DB, id)
	if err != nil {
		log.Printf("[rss_post] failed to delete post %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: http.StatusInternalServerError, Message: "删除 RSS 文章失败"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, model.ApiResponse{Code: http.StatusNotFound, Message: "未找到指定 RSS 文章"})
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: gin.H{"rows_affected": rowsAffected}})
}

// GetRssPosts handles GET /api/rss request
// Query parameters:
//   - rss_id: filter by rss_id (optional)
//   - friend_link_id: filter by friend_link_id (optional)
//   - page: for pagination (optional, default: 1)
//   - page_size: for pagination (optional, default: 10)
func (h *RssPostHandler) GetRssPosts(c *gin.Context) {
	var query model.PostQuery

	// Bind query parameters
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid query parameters"})
		return
	}

	// Set default pagination values
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	posts, total, err := friendsRepositories.GetPosts(h.DB.WithContext(c.Request.Context()), &query)
	if err != nil {
		if middleware.AbortIfCanceled(c, err) {
			return
		}
		log.Printf("[rss_post] failed to retrieve posts: %+v", err)
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to retrieve posts"})
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: &model.PaginatedResponse{
		Items:    posts,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}})
}
