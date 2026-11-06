package handlerAction

import (
	"blog_api/src/handler/response"
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FriendLinkHandler handles friend link mutations for admins and email owners.
type FriendLinkHandler struct {
	DB *gorm.DB
}

// CreateFriendLink builds a handler; public routes cannot assign tags.
func (h *FriendLinkHandler) CreateFriendLink(public bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.CreateFriendLinkReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid request body"})
			return
		}

		if public && req.TagIDs != nil {
			c.JSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "tags can only be assigned through the admin API"})
			return
		}

		if authType, ok := c.Get("auth_type"); ok && authType == "email" {
			authEmail, _ := c.Get("auth_email")
			email, _ := authEmail.(string)
			if email == "" || (req.Email != "" && req.Email != email) {
				c.JSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "email does not match the token"})
				return
			}
			req.Email = email
			req.Status = "pending"
		} else {
			req.Status = "survival"
		}

		// Insert into database
		id, err := friendsRepositories.CreateFriendLinkWithTags(h.DB.WithContext(c.Request.Context()), req.FriendWebsite, req.TagIDs)
		if friendTagError(c, err, "tag not found", "failed to create friend link") {
			return
		}

		c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: gin.H{"id": id}})
	}
}

// DeleteOwnedFriendLink handles deletion by an email-authenticated owner.
func (h *FriendLinkHandler) DeleteOwnedFriendLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid friend link ID"})
		return
	}
	authEmail, _ := c.Get("auth_email")
	email, _ := authEmail.(string)
	if email == "" {
		c.JSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "email token is invalid"})
		return
	}

	link, err := friendsRepositories.GetFriendLinkByID(h.DB, int(id))
	if response.RecordError(c, err, "friend link not found", "failed to retrieve friend link") {
		return
	}
	if link.Email != email {
		c.JSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "friend link does not belong to this email"})
		return
	}

	deleted, err := friendsRepositories.DeleteFriendLink(h.DB, uint(id), &email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "friend link ownership changed"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to delete friend link"})
		return
	}
	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: gin.H{"deleted_link": deleted.ToDTO(false)}})
}

// DeleteFriendLink handles DELETE /api/action/friend/:id request
func (h *FriendLinkHandler) DeleteFriendLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid friend link ID"})
		return
	}

	deletedLink, err := friendsRepositories.DeleteFriendLink(h.DB, uint(id), nil)
	if response.Error(c, err, "failed to delete friend link") {
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: gin.H{"deleted_link": deletedLink}})
}

// EditFriendLink handles PUT /api/action/friend/:id request
func (h *FriendLinkHandler) EditFriendLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid friend link ID"})
		return
	}

	var req model.EditFriendLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "invalid request body"})
		return
	}

	if authType, ok := c.Get("auth_type"); ok && authType == "email" {
		for field := range req.Data {
			if !model.IsFriendOwnerField(field) {
				c.JSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "field cannot be updated with this token: " + field})
				return
			}
		}
		authEmail, _ := c.Get("auth_email")
		email, _ := authEmail.(string)
		link, err := friendsRepositories.GetFriendLinkByID(h.DB, int(id))
		if response.RecordError(c, err, "friend link not found", "failed to retrieve friend link") {
			return
		}
		if link.Email == "" {
			c.JSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "email is not set for this friend link"})
			return
		}
		if email == "" || link.Email != email {
			c.JSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "email does not match the token"})
			return
		}

		rowsAffected, err := friendsRepositories.UpdateFriendLinkByOwner(h.DB, uint(id), email, req)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusForbidden, model.ApiResponse{Code: 403, Message: "friend link ownership changed"})
				return
			}
			c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: gin.H{"rows_affected": rowsAffected}})
		return
	}

	if value, exists := req.Data["skip_health_check"]; exists {
		if _, ok := value.(bool); !ok {
			c.JSON(http.StatusBadRequest, model.ApiResponse{Code: 400, Message: "skip_health_check must be a boolean"})
			return
		}
	}

	rowsAffected, err := friendsRepositories.UpdateFriendLinkByID(h.DB, uint(id), req)
	if response.Error(c, err, "failed to update friend link") {
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, model.ApiResponse{Code: 404, Message: "no friend link found with the given ID or no fields needed update"})
		return
	}

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: gin.H{"rows_affected": rowsAffected}})
}
