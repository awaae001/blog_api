package handler

import (
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FriendLinkHandler handles friend link related requests
type FriendLinkHandler struct {
	DB *gorm.DB
}

// NewFriendLinkHandler creates a new friend link handler
func NewFriendLinkHandler(db *gorm.DB) *FriendLinkHandler {
	return &FriendLinkHandler{DB: db}
}

// toFriendLinkDTOs converts a slice of FriendWebsite models to a slice of FriendLinkDTOs.
// If isPrivate is true, it includes sensitive fields like Email and Times.
func toFriendLinkDTOs(links []model.FriendWebsite, isPrivate bool) []model.FriendLinkDTO {
	dtoLinks := make([]model.FriendLinkDTO, 0, len(links))
	for _, link := range links {
		dto := model.FriendLinkDTO{
			ID:          link.ID,
			Name:        link.Name,
			Link:        link.Link,
			Avatar:      link.Avatar,
			Description: link.Info,
			Status:      link.Status,
			EnableRss:   link.EnableRss,
			UpdatedAt:   link.UpdatedAt,
		}
		if isPrivate {
			dto.Email = link.Email
			dto.Times = link.Times
			dto.IsDied = link.IsDied
		}
		dtoLinks = append(dtoLinks, dto)
	}
	return dtoLinks
}

// getFriendLinks is a helper function to get friend links with common logic.
func (h *FriendLinkHandler) getFriendLinks(c *gin.Context, isPrivate bool) {
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
			"pending":  true,
		}
		if !validStatuses[status] {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid status parameter"))
			return
		}
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Query friend links and total count
	opts := model.FriendLinkQueryOptions{
		Status: status,
		Search: search,
		Offset: offset,
		Limit:  pageSize,
		IsDied: isDied,
	}
	resp, err := friendsRepositories.QueryFriendLinks(h.DB, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to retrieve friend links"))
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

	c.JSON(http.StatusOK, model.NewSuccessResponse(paginatedData))
}

// GetAllFriendLinks handles GET /api/friend/ request
func (h *FriendLinkHandler) GetAllFriendLinks(c *gin.Context) {
	h.getFriendLinks(c, false)
}

// GetFullFriendLinks handles GET /api/action/friend/ request (authenticated)
// It returns the full friend link data, including sensitive fields.
func (h *FriendLinkHandler) GetFullFriendLinks(c *gin.Context) {
	h.getFriendLinks(c, true)
}
