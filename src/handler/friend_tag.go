package handler

import (
	"blog_api/src/handler/response"
	"blog_api/src/model"
	friendsRepositories "blog_api/src/repositories/friend"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FriendTagHandler struct{ DB *gorm.DB }

func (h *FriendTagHandler) GetTags(public bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		private := !public
		tags, err := friendsRepositories.QueryFriendTags(h.DB.WithContext(c.Request.Context()), private)
		if response.Error(c, err, "failed to retrieve friend tags") {
			return
		}
		dtos := make([]model.FriendTagDTO, 0, len(tags))
		for _, tag := range tags {
			dto := tag.FriendTag.ToDTO(private)
			count := tag.LinkCount
			dto.LinkCount = &count
			dtos = append(dtos, dto)
		}
		c.JSON(http.StatusOK, model.ApiResponse{Code: 200, Message: "success", Data: dtos})
	}
}
