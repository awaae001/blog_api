package handlerAction

import (
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	crawlerService "blog_api/src/service/crawler"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FriendRssHandler serves the admin RSS subscription API.
type FriendRssHandler struct {
	DB *gorm.DB
}

var feedStatuses = map[string]bool{
	model.FeedStatusActive: true,
	model.FeedStatusPaused: true,
	model.FeedStatusDead:   true,
}

// GetRss handles GET /api/action/rss?status=&friend_link_id=&page=&page_size=
func (h *FriendRssHandler) GetRss(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	friendLinkID, _ := strconv.Atoi(c.Query("friend_link_id"))
	status := c.Query("status")
	if page < 1 || pageSize < 1 || (status != "" && !feedStatuses[status]) {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "无效的查询参数"))
		return
	}
	pageSize = min(pageSize, 100)

	feeds, total, err := friendsRepositories.QueryRssFeeds(h.DB, model.RssFeedQueryOptions{
		FriendLinkID: friendLinkID,
		Status:       status,
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		log.Printf("[handler][rss] query feeds: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "获取订阅源失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(model.PaginatedResponse{
		Items:    feeds,
		Total:    int(total),
		Page:     page,
		PageSize: pageSize,
	}))
}

// DiscoverRss handles POST /api/action/rss/discover: list the feeds a page advertises.
func (h *FriendRssHandler) DiscoverRss(c *gin.Context) {
	var req model.DiscoverRssReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "无效的请求体: "+err.Error()))
		return
	}

	candidates, err := crawlerService.DiscoverFeeds(c.Request.Context(), strings.TrimSpace(req.URL))
	if err != nil {
		if c.Request.Context().Err() != nil {
			return
		}
		c.JSON(http.StatusBadGateway, model.NewErrorResponse(http.StatusBadGateway, "无法读取页面: "+err.Error()))
		return
	}
	if candidates == nil {
		candidates = []model.FeedCandidate{}
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(candidates))
}

// CreateRss handles POST /api/action/rss: validate the feed and subscribe.
func (h *FriendRssHandler) CreateRss(c *gin.Context) {
	var req model.CreateRssReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "无效的请求体: "+err.Error()))
		return
	}
	if !h.checkFriendLink(c, req.FriendLinkID) {
		return
	}

	feed, result, err := crawlerService.Subscribe(c.Request.Context(), h.DB, req.FeedURL, req.FriendLinkID, req.Name)
	if err != nil {
		h.writeFeedError(c, err)
		return
	}
	c.JSON(http.StatusCreated, model.NewSuccessResponseWithCode(http.StatusCreated, gin.H{"feed": feed, "result": result}))
}

// EditRss handles PUT /api/action/rss/:id.
func (h *FriendRssHandler) EditRss(c *gin.Context) {
	id, ok := feedID(c)
	if !ok {
		return
	}
	var req model.EditRssFeedReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "无效的请求体: "+err.Error()))
		return
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.FriendLinkID != nil {
		if !h.checkFriendLink(c, *req.FriendLinkID) {
			return
		}
		updates["friend_link_id"] = nil
		if *req.FriendLinkID > 0 {
			updates["friend_link_id"] = *req.FriendLinkID
		}
	}
	if req.FeedURL != nil {
		feedURL := strings.TrimSpace(*req.FeedURL)
		if _, err := crawlerService.ValidatePublicHTTPURL(feedURL); err != nil {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, err.Error()))
			return
		}
		taken, err := friendsRepositories.RssFeedURLTaken(h.DB, feedURL, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "检查订阅地址失败"))
			return
		}
		if taken {
			c.JSON(http.StatusConflict, model.NewErrorResponse(http.StatusConflict, "该订阅地址已存在"))
			return
		}
		updates["feed_url"] = feedURL
		resetSchedule(updates)
	}
	if req.Status != nil {
		switch *req.Status {
		case model.FeedStatusPaused:
			updates["status"] = model.FeedStatusPaused
		case model.FeedStatusActive:
			updates["status"] = model.FeedStatusActive
			resetSchedule(updates)
		default:
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "status 只能是 active 或 paused"))
			return
		}
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "没有需要更新的字段"))
		return
	}

	found, err := friendsRepositories.UpdateRssFeed(h.DB, id, updates)
	if err != nil {
		log.Printf("[handler][rss] update feed %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "更新订阅源失败"))
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(http.StatusNotFound, "未找到指定订阅源"))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}

// DeleteFriendRss handles DELETE /api/action/rss/:id; posts go with the feed.
func (h *FriendRssHandler) DeleteFriendRss(c *gin.Context) {
	id, ok := feedID(c)
	if !ok {
		return
	}
	rows, err := friendsRepositories.DeleteRssFeedByID(h.DB, id)
	if err != nil {
		log.Printf("[handler][rss] delete feed %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "删除订阅源失败"))
		return
	}
	if rows == 0 {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(http.StatusNotFound, "未找到指定订阅源"))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}

// FetchRss handles POST /api/action/rss/:id/fetch: refresh one feed now.
func (h *FriendRssHandler) FetchRss(c *gin.Context) {
	id, ok := feedID(c)
	if !ok {
		return
	}
	feed, err := friendsRepositories.GetRssFeedByID(h.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(http.StatusNotFound, "未找到指定订阅源"))
			return
		}
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "获取订阅源失败"))
		return
	}

	result, err := crawlerService.RefreshFeed(c.Request.Context(), h.DB, feed)
	if err != nil {
		h.writeFeedError(c, err)
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(result))
}

// resetSchedule makes the scheduler pick the feed up on its next tick with a
// clean slate.
func resetSchedule(updates map[string]any) {
	updates["etag"] = ""
	updates["last_modified"] = ""
	updates["error_count"] = 0
	updates["last_error"] = ""
	updates["next_fetch_at"] = 0
}

func feedID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, "无效的订阅源 ID"))
		return 0, false
	}
	return id, true
}

// checkFriendLink accepts 0 (no owner) or an existing friend link ID.
func (h *FriendRssHandler) checkFriendLink(c *gin.Context, id int) bool {
	if id == 0 {
		return true
	}
	exists, err := friendsRepositories.FriendLinkExists(h.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "检查友链失败"))
		return false
	}
	if !exists {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(http.StatusNotFound, "友链不存在"))
		return false
	}
	return true
}

func (h *FriendRssHandler) writeFeedError(c *gin.Context, err error) {
	switch {
	case c.Request.Context().Err() != nil:
	case errors.Is(err, crawlerService.ErrFeedExists):
		c.JSON(http.StatusConflict, model.NewErrorResponse(http.StatusConflict, "该订阅地址已存在"))
	case errors.Is(err, crawlerService.ErrRssSource):
		c.JSON(http.StatusBadGateway, model.NewErrorResponse(http.StatusBadGateway, err.Error()))
	default:
		log.Printf("[handler][rss] %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(http.StatusInternalServerError, "保存订阅源失败"))
	}
}
