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
	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "url"}},
		DoNothing: true,
	}).Create(&images).Error

	if err != nil {
		log.Printf("[db][image][ERR] 无法批量插入图片: %v", err)
		return err
	}

	log.Printf("[db][image] 成功插入 %d 条图片记录", len(images))
	return nil
}

// QueryImages 根据提供的选项查询图片，并返回分页结果和总数
func QueryImages(db *gorm.DB, opts model.ImageQueryOptions) (model.QueryImageResponse, error) {
	var resp model.QueryImageResponse
	query := db.Model(&model.Image{})

	// Get total count
	if err := query.Count(&resp.Total).Error; err != nil {
		return resp, err
	}

	// Apply pagination
	if opts.Page > 0 && opts.PageSize > 0 {
		offset := (opts.Page - 1) * opts.PageSize
		query = query.Offset(offset).Limit(opts.PageSize)
	}

	if err := query.Find(&resp.Images).Error; err != nil {
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
