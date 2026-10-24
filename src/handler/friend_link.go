package handler

import (
	"blog_api/src/middleware"
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	crawlerService "blog_api/src/service/crawler"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FriendLinkHandler handles friend link related requests
type FriendLinkHandler struct {
	DB *gorm.DB
}

// toFriendLinkDTOs converts a slice of FriendWebsite models to a slice of FriendLinkDTOs.
// If isPrivate is true, it includes sensitive fields like Email and Times.
func toFriendLinkDTOs(links []model.FriendWebsite, isPrivate bool) []model.FriendLinkDTO {
	dtoLinks := make([]model.FriendLinkDTO, 0, len(links))
	for _, link := range links {
		dtoLinks = append(dtoLinks, link.ToDTO(isPrivate))
	}
	return dtoLinks
}

// getFriendLinks is a helper function to get friend links with common logic.
func (h *FriendLinkHandler) getFriendLinks(c *gin.Context, isPrivate bool, email string) {
	// Parse query parameters
	status := c.Query("status")
	search := c.Query("search")
	isDiedStr := c.Query("is_died")
	var isDied *bool
	if isDiedStr != "" {
		val, err := strconv.ParseBool(isDiedStr)
		if err == nil {
			isDied = &val
		}
	}

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid page parameter"})
		return
	}

	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid page_size parameter"})
		return
	}

	// Limit maximum page size
	if pageSize > 1000 {
		pageSize = 1000
	}

	// Validate status parameter if provided
	switch status {
	case "", "survival", "timeout", "error", "pending":
	default:
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid status parameter"})
		return
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Query friend links and total count
	opts := model.FriendLinkQueryOptions{
		Status: status,
		Search: search,
		Email:  email,
		Offset: offset,
		Limit:  pageSize,
		IsDied: isDied,
	}
	resp, err := friendsRepositories.QueryFriendLinks(h.DB.WithContext(c.Request.Context()), opts)
	if err != nil {
		if middleware.AbortIfCanceled(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to retrieve friend links"})
		return
	}

	// Convert to DTO based on the context (public or private)
	dtoLinks := toFriendLinkDTOs(resp.Links, isPrivate)

	// Build paginated response
	paginatedData := model.PaginatedResponse{
		Items:    dtoLinks,
		Total:    int(resp.Count),
		Page:     page,
		PageSize: pageSize,
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: paginatedData})
}

func (h *FriendLinkHandler) loadFriendLink(c *gin.Context) (model.FriendWebsite, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid friend link ID"})
		return model.FriendWebsite{}, false
	}

	link, err := friendsRepositories.GetFriendLinkByID(h.DB.WithContext(c.Request.Context()), id)
	if err == nil {
		return link, true
	}
	if !middleware.AbortIfCanceled(c, err) {
		status, message := http.StatusInternalServerError, "failed to retrieve friend link"
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status, message = http.StatusNotFound, "friend link not found"
		}
		c.JSON(status, model.ApiResponse{Code: status, Message: message})
	}
	return model.FriendWebsite{}, false
}

func (h *FriendLinkHandler) getFriendLinkByID(c *gin.Context, isPrivate bool) {
	if link, ok := h.loadFriendLink(c); ok {
		c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: link.ToDTO(isPrivate)})
	}
}

// GetAllFriendLinks handles GET /api/public/friend/ request
func (h *FriendLinkHandler) GetAllFriendLinks(c *gin.Context) {
	h.getFriendLinks(c, false, "")
}

// GetFriendLinkByID handles GET /api/public/friend/:id request
func (h *FriendLinkHandler) GetFriendLinkByID(c *gin.Context) {
	h.getFriendLinkByID(c, false)
}

// GetFriendLinkByEmailToken handles GET /api/public/friend/self request (email token).
func (h *FriendLinkHandler) GetFriendLinkByEmailToken(c *gin.Context) {
	authType, ok := c.Get("auth_type")
	if !ok || authType != "email" {
		c.JSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "email token is required"})
		return
	}

	authEmail, _ := c.Get("auth_email")
	email, _ := authEmail.(string)
	if email == "" {
		c.JSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "email token is invalid"})
		return
	}

	h.getFriendLinks(c, false, email)
}

// GetFullFriendLinks handles GET /api/action/friend/ request (authenticated)
// It returns the full friend link data, including sensitive fields.
func (h *FriendLinkHandler) GetFullFriendLinks(c *gin.Context) {
	h.getFriendLinks(c, true, "")
}

// GetFullFriendLinkByID handles GET /api/action/friend/:id request (authenticated)
func (h *FriendLinkHandler) GetFullFriendLinkByID(c *gin.Context) {
	h.getFriendLinkByID(c, true)
}

// RecheckFriendLink handles POST /api/action/friend/recheck/:id.
// It performs one inspection even when scheduled health checks are disabled.
func (h *FriendLinkHandler) RecheckFriendLink(c *gin.Context) {
	link, ok := h.loadFriendLink(c)
	if !ok {
		return
	}

	result := crawlerService.CrawlWebsite(c.Request.Context(), link.Link)
	if c.Request.Context().Err() != nil {
		return
	}
	if err := friendsRepositories.UpdateFriendLink(h.DB, link, result); err != nil {
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: http.StatusInternalServerError, Message: "failed to update friend link inspection"})
		return
	}

	crawlerService.SubscribeDiscovered(c.Request.Context(), h.DB, link, result.Feeds)

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: nil})
}

// RecheckAllFriendLinks handles POST /api/action/friend/recheck.
// The scan covers every friend link that has not opted out of health checks and
// runs in the background, because one full pass can take minutes.
func (h *FriendLinkHandler) RecheckAllFriendLinks(c *gin.Context) {
	if err := crawlerService.StartFriendLinksInspection(h.DB, crawlerService.FullInspectionScope()); err != nil {
		if errors.Is(err, crawlerService.ErrInspectionBusy) {
			c.JSON(http.StatusConflict, model.ApiResponse{Code: http.StatusConflict, Message: "a friend link inspection is already running"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: http.StatusInternalServerError, Message: "failed to start friend link inspection"})
		return
	}

	c.JSON(http.StatusAccepted, model.ApiResponse{
		Code:    http.StatusAccepted,
		Message: "inspection started",
		Data:    crawlerService.InspectionProgressSnapshot(),
	})
}

// GetRecheckProgress handles GET /api/action/friend/recheck.
// It reports the current or most recent inspection run of this process.
func (h *FriendLinkHandler) GetRecheckProgress(c *gin.Context) {
	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: crawlerService.InspectionProgressSnapshot()})
}

// GetFriendLinkRecheckProgress handles GET /api/action/friend/recheck/:id.
// It reports one link's position inside the current or most recent run together
// with the inspection state stored for that link.
func (h *FriendLinkHandler) GetFriendLinkRecheckProgress(c *gin.Context) {
	link, ok := h.loadFriendLink(c)
	if !ok {
		return
	}

	progress := crawlerService.LinkInspectionSnapshot(link.ID)
	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: gin.H{
		"id":         link.ID,
		"in_run":     progress.InRun,
		"done":       progress.Done,
		"run_status": progress.Status,
		"run":        progress.Run,
		"link":       link.ToDTO(true),
	}})
}
