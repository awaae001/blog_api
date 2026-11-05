package handlerAction

import (
	"blog_api/src/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseRecordID(c *gin.Context, message string) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: http.StatusBadRequest, Message: message})
		return 0, false
	}
	return id, true
}
