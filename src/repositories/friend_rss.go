package repositories

import (
	"blog_api/src/model"
	"database/sql"
	"fmt"
	"log"
)

// InsertFriendRss inserts new RSS feeds for a friend link, avoiding duplicates.
func InsertFriendRss(db *sql.DB, friendLinkID int, rssURLs []string) error {
	if len(rssURLs) == 0 {
		return nil
	}

	log.Printf("开始为友链 ID: %d 插入 RSS feeds", friendLinkID)

	stmt, err := db.Prepare("INSERT INTO friend_rss (friend_link_id, rss_url) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("could not prepare insert statement for friend_rss: %w", err)
	}
	defer stmt.Close()

	for _, rssURL := range rssURLs {
		var exists bool
		// Check if the RSS feed already exists for this friend link
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM friend_rss WHERE friend_link_id = ? AND rss_url = ?)", friendLinkID, rssURL).Scan(&exists)
		if err != nil {
			log.Printf("无法为友链 %d 检查已存在的 RSS feed %s: %v", friendLinkID, rssURL, err)
			continue
		}

		if !exists {
			if _, err := stmt.Exec(friendLinkID, rssURL); err != nil {
				log.Printf("无法为友链 %d 插入 RSS feed %s: %v", friendLinkID, rssURL, err)
			} else {
				log.Printf("已为友链 ID: %d 插入 RSS feed: %s", friendLinkID, rssURL)
			}
		} else {
			log.Printf("友链 %d 的 RSS feed %s 已存在，跳过。", friendLinkID, rssURL)
		}
	}

	log.Printf("友链 ID: %d 的 RSS feed 插入流程完成", friendLinkID)
	return nil
}

// GetAllFriendRss retrieves all RSS feeds from the database.
func GetAllFriendRss(db *sql.DB) ([]model.FriendRss, error) {
	rows, err := db.Query("SELECT id, friend_link_id, rss_url, status, updated_at FROM friend_rss")
	if err != nil {
		return nil, fmt.Errorf("could not query friend_rss: %w", err)
	}
	defer rows.Close()

	var rssFeeds []model.FriendRss
	for rows.Next() {
		var rss model.FriendRss
		if err := rows.Scan(&rss.ID, &rss.FriendLinkID, &rss.RssURL, &rss.Status, &rss.UpdatedAt); err != nil {
			log.Printf("无法扫描 friend_rss 行: %v", err)
			continue
		}
		rssFeeds = append(rssFeeds, rss)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over friend_rss rows: %w", err)
	}

	return rssFeeds, nil
}
