package repositories

import (
	"blog_api/src/model"

	"gorm.io/gorm"
)

// GetFingerprintByValue retrieves a fingerprint by its hash value.
func GetFingerprintByValue(db *gorm.DB, fingerprint string) (*model.Fingerprint, error) {
	var record model.Fingerprint
	if err := db.Where("fingerprint = ?", fingerprint).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// CreateFingerprint inserts a new fingerprint record.
func CreateFingerprint(db *gorm.DB, fingerprint *model.Fingerprint) error {
	return db.Create(fingerprint).Error
}
