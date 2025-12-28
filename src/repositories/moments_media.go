package repositories

import (
	"blog_api/src/model"

	"gorm.io/gorm"
)

// CreateMomentMedia creates a new media record for a moment.
func CreateMomentMedia(db *gorm.DB, media *model.MomentMedia) error {
	return db.Create(media).Error
}

// DeleteMomentMedia deletes a media record by its ID.
// When hard is false, it performs a soft delete by setting is_deleted = 1.
func DeleteMomentMedia(db *gorm.DB, id int, hard bool) error {
	var result *gorm.DB
	if hard {
		result = db.Where("id = ?", id).Delete(&model.MomentMedia{})
	} else {
		result = db.Model(&model.MomentMedia{}).Where("id = ?", id).Update("is_deleted", 1)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// UpdateMomentMedia updates fields for a media record.
func UpdateMomentMedia(db *gorm.DB, id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	result := db.Model(&model.MomentMedia{}).Where("id = ?", id).Updates(updates)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// QueryMedia retrieves media based on pagination and filters, returning the list and total count.
func QueryMedia(db *gorm.DB, page, pageSize, momentID int, mediaType string) ([]model.MomentMedia, int64, error) {
	var media []model.MomentMedia
	var total int64

	query := db.Model(&model.MomentMedia{}).Where("is_deleted = 0")

	if momentID > 0 {
		query = query.Where("moment_id = ?", momentID)
	}
	if mediaType != "" {
		query = query.Where("media_type = ?", mediaType)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}
	if err := query.Order("id desc").Find(&media).Error; err != nil {
		return nil, 0, err
	}

	return media, total, nil
}
