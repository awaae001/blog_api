package handler

import (
	"blog_api/src/model"
	imageRepositories "blog_api/src/repositories/image"
	"errors"
	"log"
	"net/http"
	"strconv"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// ImagePublicHandler 处理与图片相关的公开API请求
type ImagePublicHandler struct {
	db *gorm.DB
}

// NewImagePublicHandler 创建一个新的 ImagePublicHandler 实例
func NewImagePublicHandler(db *gorm.DB) *ImagePublicHandler {
	return &ImagePublicHandler{db: db}
}

// GetImage 根据ID获取图片信息或随机获取图片
func (h *ImagePublicHandler) GetImage(c *gin.Context) {
	idStr := c.Param("id")
	var image *model.Image
	var err error

	if idStr == "" || idStr == "/" {
		// 如果没有提供ID，则随机获取一张图片
		image, err = imageRepositories.GetRandomImage(h.db)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "没有可用的图片"})
			} else {
				log.Printf("[handler][image_public][ERR] 获取随机图片失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
			}
			return
		}
	} else {
		// 如果提供了ID，则解析并获取指定的图片
		id, errConv := strconv.Atoi(idStr)
		if errConv != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的图片ID"})
			return
		}

		image, err = imageRepositories.GetImageByID(h.db, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "图片未找到"})
			} else {
				log.Printf("[handler][image_public][ERR] 查询图片失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
			}
			return
		}
	}

	// 检查查询参数 `type`
	queryType := c.DefaultQuery("type", "image")

	if queryType == "metadata" {
		// 如果 type=metadata，返回图片的元数据
		c.JSON(http.StatusOK, image)
	} else {
		// 否则，执行302重定向到图片URL
		c.Redirect(http.StatusFound, image.URL)
	}
}
