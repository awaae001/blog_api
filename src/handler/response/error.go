package response

import (
	"blog_api/src/middleware"
	"blog_api/src/model"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Error handles a server error and reports whether the handler should return.
func Error(c *gin.Context, err error, message string) bool {
	if err == nil {
		return false
	}
	if !middleware.AbortIfCanceled(c, err) {
		log.Printf("[handler] %s %s: %s: %v", c.Request.Method, c.FullPath(), message, err)
		c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: http.StatusInternalServerError, Message: message})
	}
	return true
}

// RecordError also maps missing database records to 404.
func RecordError(c *gin.Context, err error, notFound, failure string) bool {
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Error(c, err, failure)
	}
	if !middleware.AbortIfCanceled(c, err) {
		c.JSON(http.StatusNotFound, model.ApiResponse{Code: http.StatusNotFound, Message: notFound})
	}
	return true
}
