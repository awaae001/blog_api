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

	log.Printf("Start inserting RSS feeds for friend link ID: %d", friendLinkID)

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
			log.Printf("Could not check for existing RSS feed %s for friend link %d: %v", rssURL, friendLinkID, err)
			continue
		}

		if !exists {
			if _, err := stmt.Exec(friendLinkID, rssURL); err != nil {
				log.Printf("Could not insert RSS feed %s for friend link %d: %v", rssURL, friendLinkID, err)
			} else {
				log.Printf("Inserted RSS feed: %s for friend link ID: %d", rssURL, friendLinkID)
			}
		} else {
			log.Printf("RSS feed %s already exists for friend link %d, skipping.", rssURL, friendLinkID)
		}
	}

	log.Printf("RSS feed insertion process completed for friend link ID: %d", friendLinkID)
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
			log.Printf("could not scan friend_rss row: %v", err)
			continue
		}
		rssFeeds = append(rssFeeds, rss)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over friend_rss rows: %w", err)
	}

	return rssFeeds, nil
}
