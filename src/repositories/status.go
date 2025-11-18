package repositories

import (
	"blog_api/src/model"
	"database/sql"
	"fmt"
)

// GetSystemStats retrieves all system statistics in a single query.
func GetSystemStats(db *sql.DB) (model.StatusData, error) {
	var stats model.StatusData
	query := `
		SELECT
			(SELECT COUNT(*) FROM friend_link),
			(SELECT COUNT(*) FROM friend_rss),
			(SELECT COUNT(*) FROM friend_rss_post)
	`
	err := db.QueryRow(query).Scan(&stats.FriendLinkCount, &stats.RssCount, &stats.RssPostCount)
	if err != nil {
		return stats, fmt.Errorf("could not query system stats: %w", err)
	}
	return stats, nil
}
