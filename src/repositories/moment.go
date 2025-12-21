package repositories

import (
	"blog_api/src/model"

	"gorm.io/gorm"
)

// QueryMoments retrieves moments based on pagination and returns the list and total count.
func QueryMoments(db *gorm.DB, page, pageSize int, status string) ([]model.Moment, int64, error) {
	var moments []model.Moment
	var total int64

	query := db.Model(&model.Moment{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	if err := query.Order("created_at desc").Find(&moments).Error; err != nil {
		return nil, 0, err
	}

	return moments, total, nil
}

// GetMediaForMoments retrieves media files for a list of moment IDs.
func GetMediaForMoments(db *gorm.DB, momentIDs []int) ([]model.MomentMedia, error) {
	var media []model.MomentMedia
	if len(momentIDs) == 0 {
		return media, nil
	}
	if err := db.Where("moment_id IN ? AND is_deleted = 0", momentIDs).Find(&media).Error; err != nil {
		return nil, err
	}

	return media, nil
}

// CreateMoment creates a new moment and its associated media in a transaction.
func CreateMoment(db *gorm.DB, moment *model.Moment, media []model.MomentMedia) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(moment).Error; err != nil {
			return err
		}

		if len(media) > 0 {
			for i := range media {
				media[i].MomentID = moment.ID
			}
			if err := tx.Create(&media).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// DeleteMoment updates the status of a moment to 'deleted'.
func DeleteMoment(db *gorm.DB, id int) error {
	return db.Model(&model.Moment{}).Where("id = ?", id).Update("status", "deleted").Error
}

// MomentExistsByChannelMessage checks if a moment already exists for a channel/message pair.
func MomentExistsByChannelMessage(db *gorm.DB, channelID, messageID int64) (bool, error) {
	var count int64
	if err := db.Model(&model.Moment{}).Where("channel_id = ? AND message_id = ?", channelID, messageID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
