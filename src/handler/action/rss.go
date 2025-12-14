package handlerAction

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"blog_api/src/service"
	"fmt"
	"net/http"
	"strconv"

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

	name, err := service.GetRssTitle(req.RssURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "无法获取 RSS 标题: "+err.Error()))
		return
	}

	createdFeed, err := repositories.CreateFriendRssFeeds(h.DB, req.FriendLinkID, req.RssURL, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "无法创建 friend_rss 记录: "+err.Error()))
		return
	}
	if createdFeed == nil {
		c.JSON(http.StatusConflict, model.NewErrorResponse(409, "RSS feed 已存在"))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(gin.H{"id": createdFeed.ID}))
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

	name, err := service.GetRssTitle(req.RssURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "无法获取 RSS 标题: "+err.Error()))
		return
	}

	createdFeed, err := repositories.CreateFriendRssFeeds(h.DB, friendLinkID, req.RssURL, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "创建 RSS 失败: "+err.Error()))
		return
	}
	if createdFeed == nil {
		c.JSON(http.StatusConflict, model.NewErrorResponse(http.StatusConflict, "RSS feed 已存在"))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(gin.H{"id": createdFeed.ID}))
}

// GetRss handles GET /api/action/rss
func (h *FriendRssHandler) GetRss(c *gin.Context) {
	// Parse query parameters
	status := c.Query("status")

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid page parameter"))
		return
	}

	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid page_size parameter"))
		return
	}

	// Limit maximum page size
	if pageSize > 100 {
		pageSize = 100
	}

	// Validate status parameter if provided
	if status != "" {
		validStatuses := map[string]bool{
			"survival": true,
			"timeout":  true,
			"error":    true,
			"pause":    true,
			"valid":    true,
		}
		if !validStatuses[status] {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid status parameter"))
			return
		}
	}

	// Query friend links and total count
	opts := model.FriendRssQueryOptions{
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	}
	resp, err := repositories.QueryFriendRss(h.DB, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to retrieve friend links"))
		return
	}

	// Build paginated response
	paginatedData := model.PaginatedResponse{
		Items:    resp.Feeds,
		Total:    int(resp.Total),
		Page:     page,
		PageSize: pageSize,
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(paginatedData))
}
