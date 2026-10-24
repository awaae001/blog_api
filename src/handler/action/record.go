package handlerAction

import (
	"blog_api/src/middleware"
	"blog_api/src/model"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func parseRecordID(c *gin.Context, message string) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, model.ApiResponse{Code: http.StatusBadRequest, Message: message})
		return 0, false
	}
	return id, true
}

func writeRecordError(c *gin.Context, err error, notFound, failure string) {
	if middleware.AbortIfCanceled(c, err) {
		return
	}
	status, message := http.StatusInternalServerError, failure
	if errors.Is(err, gorm.ErrRecordNotFound) {
		status, message = http.StatusNotFound, notFound
	}
	c.JSON(status, model.ApiResponse{Code: status, Message: message})
}
