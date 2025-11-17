package handler

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UpdataHandler handles updata related requests
type UpdataHandler struct {
	DB *sql.DB
}

// NewUpdataHandler creates a new updata handler
func NewUpdataHandler(db *sql.DB) *UpdataHandler {
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

	// Set default avatar if not provided
	if req.Avatar == "" {
		req.Avatar = "/Rss.webp"
	}

	// Insert into database
	id, err := repositories.CreateFriendLink(h.DB, req)
	if err != nil {
		log.Printf("[handler][updata][ERR] Failed to create friend link: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to create friend link"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"id": id}))
}
