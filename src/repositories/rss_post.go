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
		log.Printf("链接为 %s 的文章已存在，跳过。", post.Link)
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

	log.Printf("已插入新文章: %s", post.Title)
	return nil
}

// GetPostsByFriendLinkID retrieves all posts associated with a given friend_link_id.
func GetPostsByFriendLinkID(db *sql.DB, friendLinkID int) ([]model.RssPost, error) {
	query := `
		SELECT
			p.id,
			p.friend_rss_id,
			p.title,
			p.link,
			p.description,
			p.time
		FROM
			friend_rss_post p
		JOIN
			friend_rss r ON p.friend_rss_id = r.id
		WHERE
			r.friend_link_id = ?
		ORDER BY
			p.time DESC
	`

	rows, err := db.Query(query, friendLinkID)
	if err != nil {
		return nil, fmt.Errorf("could not query posts by friend_link_id: %w", err)
	}
	defer rows.Close()

	var posts []model.RssPost
	for rows.Next() {
		var post model.RssPost
		if err := rows.Scan(&post.ID, &post.FriendRssID, &post.Title, &post.Link, &post.Description, &post.Time); err != nil {
			return nil, fmt.Errorf("could not scan post row: %w", err)
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return posts, nil
}

// GetAllPosts retrieves all posts with pagination.
func GetAllPosts(db *sql.DB, page, pageSize int) ([]model.RssPost, int, error) {
	// First, get the total count of posts
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM friend_rss_post").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("could not query total posts count: %w", err)
	}

	// Then, get the posts for the current page
	query := `
		SELECT
			id,
			friend_rss_id,
			title,
			link,
			description,
			time
		FROM
			friend_rss_post
		ORDER BY
			time DESC
		LIMIT ? OFFSET ?
	`

	offset := (page - 1) * pageSize
	rows, err := db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("could not query posts with pagination: %w", err)
	}
	defer rows.Close()

	var posts []model.RssPost
	for rows.Next() {
		var post model.RssPost
		if err := rows.Scan(&post.ID, &post.FriendRssID, &post.Title, &post.Link, &post.Description, &post.Time); err != nil {
			return nil, 0, fmt.Errorf("could not scan post row: %w", err)
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error during rows iteration: %w", err)
	}

	return posts, total, nil
}
