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
