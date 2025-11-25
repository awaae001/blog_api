package handler

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
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

// GetAllFriendLinks handles GET /api/friend/ request
// Query parameters:
//   - status: filter by status (optional)
//   - page: page number, default 1 (optional)
//   - page_size: items per page, default 20, max 100 (optional)
func (h *FriendLinkHandler) GetAllFriendLinks(c *gin.Context) {
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
			"died":     true,
			"pending":  true,
		}
		if !validStatuses[status] {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid status parameter"))
			return
		}
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get total count
	total, err := repositories.CountFriendLinks(h.DB, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to count friend links"))
		return
	}

	// Get friend links
	links, err := repositories.GetFriendLinksWithFilter(h.DB, status, offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to retrieve friend links"))
		return
	}

	// Convert to DTO (Data Transfer Object) to hide sensitive fields like 'times'
	dtoLinks := make([]model.FriendLinkDTO, 0, len(links))
	for _, link := range links {
		dtoLinks = append(dtoLinks, model.FriendLinkDTO{
			ID:             link.ID,
			WebsiteName:    link.Name,
			WebsiteURL:     link.Link,
			WebsiteIconURL: link.Avatar,
			Description:    link.Info,
			Status:         link.Status,
			UpdatedAt:      link.UpdatedAt,
		})
	}

	// Build paginated response
	paginatedData := model.PaginatedResponse{
		Items:    dtoLinks,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(paginatedData))
}
