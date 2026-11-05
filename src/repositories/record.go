package repositories

import "gorm.io/gorm"

// RecordError preserves database errors and reports mutations that matched no record.
func RecordError(result *gorm.DB) error {
	if result.Error != nil || result.RowsAffected > 0 {
		return result.Error
	}
	return gorm.ErrRecordNotFound
}
