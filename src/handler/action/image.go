package handlerAction

import (
	"blog_api/src/model"
	imageRepositories "blog_api/src/repositories/image"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ImageHandler handles image related requests
type ImageHandler struct {
	DB *gorm.DB
}

func NewImageHandler(db *gorm.DB) *ImageHandler {
	return &ImageHandler{DB: db}
}

// GetImages handles GET /api/action/image request
func (h *ImageHandler) GetImages(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "无效的页面参数"))
		return
	}

	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "无效的页面大小参数"))
		return
	}

	if pageSize > 100 {
		pageSize = 100
	}

	opts := model.ImageQueryOptions{
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := imageRepositories.QueryImages(h.DB, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "获取图片列表失败"))
		return
	}

	paginatedData := model.PaginatedResponse{
		Items:    resp.Images,
		Total:    int(resp.Total),
		Page:     page,
		PageSize: pageSize,
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(paginatedData))
}
