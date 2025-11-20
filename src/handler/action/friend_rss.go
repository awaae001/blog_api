package handlerAction

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// FriendRssHandler 处理与 friend_rss 相关的请求。
type FriendRssHandler struct {
	DB *sql.DB
}

// NewFriendRssHandler 创建一个新的 FriendRssHandler。
func NewFriendRssHandler(db *sql.DB) *FriendRssHandler {
	return &FriendRssHandler{DB: db}
}

// CreateFriendRss 处理 POST /api/action/rss 请求。
func (h *FriendRssHandler) CreateFriendRss(c *gin.Context) {
	var req model.CreateFriendRssReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "无效的请求体: "+err.Error()))
		return
	}

	// 检查 friend_link_id 是否真实存在
	exists, err := repositories.FriendLinkExists(h.DB, req.FriendLinkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "检查友链是否存在时出错: "+err.Error()))
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(http.StatusNotFound, fmt.Sprintf("ID 为 %d 的友链不存在", req.FriendLinkID)))
		return
	}

	id, err := repositories.CreateFriendRss(h.DB, req.FriendLinkID, req.RssURL)
	if err != nil {
		// 可以考虑检查特定错误，例如重复条目。
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "创建友链 RSS 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(gin.H{"id": id}))
}
