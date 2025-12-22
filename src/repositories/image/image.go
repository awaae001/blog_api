package imageRepositories

import (
	"blog_api/src/model"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BatchInsertImages 批量插入图片信息到数据库
// 使用 OnConflict 来避免插入重复的 URL
func BatchInsertImages(db *gorm.DB, images []model.Image) error {
	if len(images) == 0 {
		log.Println("[db][image] No images to insert.")
		return nil
	}

	// 使用 OnConflict 子句，当 url 冲突时，不执行任何操作
	// 这可以防止重复插入相同的图片
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "url"}},
		DoNothing: true,
	}).Create(&images)

	if result.Error != nil {
		log.Printf("[db][image][ERR] 无法批量插入图片: %v", result.Error)
		return result.Error
	}

	log.Printf("[db][image] 成功插入 %d 条图片记录", result.RowsAffected)
	return nil
}

// FilterNonExistingImages takes a slice of images and returns only those that do not exist in the database based on URL.
func FilterNonExistingImages(db *gorm.DB, images []model.Image) ([]model.Image, error) {
	if len(images) == 0 {
		return images, nil
	}

	var urls []string
	for _, img := range images {
		urls = append(urls, img.URL)
	}

	var existingURLs []string
	// Find all URLs from the input list that already exist in the DB
	if err := db.Model(&model.Image{}).Where("url IN ?", urls).Pluck("url", &existingURLs).Error; err != nil {
		return nil, err
	}

	// Create a map for faster lookup
	existingMap := make(map[string]bool)
	for _, url := range existingURLs {
		existingMap[url] = true
	}

	var newImages []model.Image
	for _, img := range images {
		if !existingMap[img.URL] {
			newImages = append(newImages, img)
		}
	}

	return newImages, nil
}

// QueryImages 根据提供的选项查询图片，并返回分页结果和总数
func QueryImages(db *gorm.DB, opts model.ImageQueryOptions) (model.QueryImageResponse, error) {
	var resp model.QueryImageResponse
	query := db.Model(&model.Image{})

	// Apply status filter
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}

	// Apply name filter for fuzzy search
	if opts.Name != "" {
		query = query.Where("name LIKE ?", "%"+opts.Name+"%")
	}
	if err := query.Count(&resp.Total).Error; err != nil {
		return resp, err
	}
	if opts.Page > 0 && opts.PageSize > 0 {
		offset := (opts.Page - 1) * opts.PageSize
		query = query.Offset(offset).Limit(opts.PageSize)
	}

	if err := query.Order("id desc").Find(&resp.Images).Error; err != nil {
		return resp, err
	}

	return resp, nil
}

// CreateImage inserts a single image record into the database.
func CreateImage(db *gorm.DB, image *model.Image) error {
	err := db.Create(image).Error
	if err != nil {
		log.Printf("[db][image][ERR] 无法创建图片: %v", err)
		return err
	}
	log.Printf("[db][image] 成功创建图片记录，ID: %d", image.ID)
	return nil
}

// UpdateImage updates an existing image record in the database.
func UpdateImage(db *gorm.DB, image *model.Image) error {
	result := db.Model(&model.Image{}).Where("id = ?", image.ID).Updates(map[string]interface{}{
		"name":       image.Name,
		"url":        image.URL,
		"local_path": image.LocalPath,
		"is_local":   image.IsLocal,
		"is_oss":     image.IsOss,
		"status":     image.Status,
	})

	if result.RowsAffected == 0 {
		log.Printf("[db][image][WARN] 未找到要更新的图片，ID: %d", image.ID)
		return gorm.ErrRecordNotFound
	}

	log.Printf("[db][image] 成功更新图片记录，ID: %d", image.ID)
	return nil
}

// DeleteImage deletes an image record from the database by its ID.
func DeleteImage(db *gorm.DB, id int) error {
	result := db.Delete(&model.Image{}, id)
	if result.RowsAffected == 0 {
		log.Printf("[db][image][WARN] 未找到要删除的图片，ID: %d", id)
		return gorm.ErrRecordNotFound
	}

	log.Printf("[db][image] 成功删除图片记录，ID: %d", id)
	return nil
}

// GetImageByID retrieves a single image by its ID.
func GetImageByID(db *gorm.DB, id int) (*model.Image, error) {
	var image model.Image
	err := db.First(&image, id).Error
	if err != nil {
		log.Printf("[db][image][ERR] 无法通过ID %d 找到图片: %v", id, err)
		return nil, err
	}
	return &image, nil
}

// GetRandomImage retrieves a random image from the database.
func GetRandomImage(db *gorm.DB) (*model.Image, error) {
	var image model.Image
	err := db.Order("RANDOM()").First(&image).Error
	if err != nil {
		log.Printf("[db][image][ERR] 无法获取随机图片: %v", err)
		return nil, err
	}
	return &image, nil
}
