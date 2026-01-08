package handlerAction

import (
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UpdataHandler handles updata related requests
type UpdataHandler struct {
	DB *gorm.DB
}

// NewUpdataHandler creates a new updata handler
func NewUpdataHandler(db *gorm.DB) *UpdataHandler {
	return &UpdataHandler{DB: db}
}

// CreateFriendLink handles POST /api/updata/friend request
func (h *UpdataHandler) CreateFriendLink(c *gin.Context) {
	log.Println("[handler][updata] Received friend link creation request")
	var req model.FriendWebsite
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[handler][updata][ERR] JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid request body"))
		return
	}

	log.Printf("[handler][updata] Received friend link data: %+v", req)

	// Insert into database
	id, err := friendsRepositories.CreateFriendLink(h.DB, req)
	if err != nil {
		log.Printf("[handler][updata][ERR] 创建友情链接失败: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to create friend link"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"id": id}))
}

// DeleteFriendLink handles DELETE /api/action/friend/:id request
func (h *UpdataHandler) DeleteFriendLink(c *gin.Context) {
	log.Println("[handler][updata] Received friend link deletion request")
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid friend link ID"))
		return
	}

	log.Printf("[handler][updata] Received friend link deletion request for ID: %d", id)

	deletedLink, err := friendsRepositories.DeleteFriendLinkByID(h.DB, uint(id))
	if err != nil {
		log.Printf("[handler][updata][ERR] 删除友情链接失败: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to delete friend link"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"deleted_link": deletedLink}))
}

// EditFriendLink handles PUT /api/action/friend/:id request
func (h *UpdataHandler) EditFriendLink(c *gin.Context) {
	log.Println("[handler][updata] Received friend link edit request")
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid friend link ID"))
		return
	}

	var req model.EditFriendLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[handler][updata][ERR] JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "invalid request body"))
		return
	}

	log.Printf("[handler][updata] Received friend link edit data for ID %d: %+v", id, req)

	rowsAffected, err := friendsRepositories.UpdateFriendLinkByID(h.DB, uint(id), req)
	if err != nil {
		log.Printf("[handler][updata][ERR] 更新友情链接失败: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to update friend link"))
		return
	}

	if rowsAffected == 0 {
		log.Printf("[handler][updata] No friend link found with ID %d or no fields to update", id)
		c.JSON(http.StatusNotFound, model.NewErrorResponse(404, "no friend link found with the given ID or no fields needed update"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"rows_affected": rowsAffected}))
}
