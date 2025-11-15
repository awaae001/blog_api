package repositories

import (
	"blog_api/src/model"
	"database/sql"
	"fmt"
	"log"
)

// InsertRssPost inserts a new post into the database, avoiding duplicates.
func InsertRssPost(db *sql.DB, post *model.RssPost) error {
	// Check if the post already exists based on the link
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM friend_rss_post WHERE link = ?)", post.Link).Scan(&exists)
	if err != nil {
		return fmt.Errorf("could not check for existing post: %w", err)
	}

	if exists {
		log.Printf("Post with link %s already exists, skipping.", post.Link)
		return nil
	}

	stmt, err := db.Prepare("INSERT INTO friend_rss_post (friend_rss_id, title, link, description, time) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("could not prepare insert statement for friend_rss_post: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(post.FriendRssID, post.Title, post.Link, post.Description, post.Time)
	if err != nil {
		return fmt.Errorf("could not insert post: %w", err)
	}

	log.Printf("Inserted new post: %s", post.Title)
	return nil
}
