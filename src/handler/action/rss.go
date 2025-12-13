package handlerAction

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FriendRssHandler 处理与 friend_rss 相关的请求
type FriendRssHandler struct {
	DB *gorm.DB
}

// NewFriendRssHandler 创建一个新的 FriendRssHandler
func NewRssHandler(db *gorm.DB) *FriendRssHandler {
	return &FriendRssHandler{DB: db}
}

// CreateFriendRss 处理 POST /api/action/rss 请求
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

	createdFeeds, err := repositories.CreateFriendRssFeeds(h.DB, req.FriendLinkID, []string{req.RssURL})
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "无法创建 friend_rss 记录"))
		return
	}
	if len(createdFeeds) == 0 {
		c.JSON(http.StatusConflict, model.NewErrorResponse(409, "RSS feed 已存在"))
		return
	}
	id := createdFeeds[0].ID

	c.JSON(http.StatusCreated, model.NewSuccessResponse(gin.H{"id": id}))
}

// DeleteFriendRss handles DELETE /api/action/rss
func (h *FriendRssHandler) DeleteFriendRss(c *gin.Context) {
	var req model.DeleteFriendRssReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "无效的请求体: "+err.Error()))
		return
	}

	id, err := repositories.DeleteFriendRssByURL(h.DB, req.RssURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "删除 RSS 失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"id": id}))
}

// CreateRss handles PUT /api/action/rss
func (h *FriendRssHandler) CreateRss(c *gin.Context) {
	var req model.CreateRssReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "无效的请求体: "+err.Error()))
		return
	}

	friendLinkID := req.FriendLinkID
	if friendLinkID == 0 {
		friendLinkID = -1
	}

	if friendLinkID != -1 {
		// 检查 friend_link_id 是否真实存在
		exists, err := repositories.FriendLinkExists(h.DB, friendLinkID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "检查友链是否存在时出错: "+err.Error()))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(http.StatusNotFound, fmt.Sprintf("ID 为 %d 的友链不存在", friendLinkID)))
			return
		}
	}

	createdFeeds, err := repositories.CreateFriendRssFeeds(h.DB, friendLinkID, []string{req.RssURL})
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "创建 RSS 失败: "+err.Error()))
		return
	}
	if len(createdFeeds) == 0 {
		c.JSON(http.StatusConflict, model.NewErrorResponse(http.StatusConflict, "RSS feed 已存在"))
		return
	}
	id := createdFeeds[0].ID

	c.JSON(http.StatusCreated, model.NewSuccessResponse(gin.H{"id": id}))
}
