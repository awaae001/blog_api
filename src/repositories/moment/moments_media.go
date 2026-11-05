package momentRepositories

import (
	"blog_api/src/model"
	"blog_api/src/repositories"

	"gorm.io/gorm"
)

// DeleteMomentMedia deletes a media record by its ID.
func DeleteMomentMedia(db *gorm.DB, id int) error {
	return repositories.RecordError(db.Where("id = ?", id).Delete(&model.MomentMedia{}))
}

// UpdateMomentMedia updates fields for a media record.
func UpdateMomentMedia(db *gorm.DB, media *model.MomentMedia, updates model.UpdateMomentMediaReq) error {
	if updates == (model.UpdateMomentMediaReq{}) {
		return nil
	}

	return repositories.RecordError(db.Model(media).Updates(updates))
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
