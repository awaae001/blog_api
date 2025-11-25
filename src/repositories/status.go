package repositories

import (
	"blog_api/src/model"
	"fmt"

	"gorm.io/gorm"
)

// GetSystemStats retrieves all system statistics in a single query.
func GetSystemStats(db *gorm.DB) (model.StatusData, error) {
	var stats model.StatusData
	query := `
		SELECT
			(SELECT COUNT(*) FROM friend_link) AS friend_link_count,
			(SELECT COUNT(*) FROM friend_rss) AS rss_count,
			(SELECT COUNT(*) FROM friend_rss_post) AS rss_post_count
	`
	err := db.Raw(query).Scan(&stats).Error
	if err != nil {
		return stats, fmt.Errorf("could not query system stats: %w", err)
	}
	return stats, nil
}
