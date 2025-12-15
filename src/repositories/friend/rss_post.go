package friendsRepositories

import (
	"blog_api/src/model"
	"fmt"
	"log"

	"gorm.io/gorm"
)

// InsertRssPost inserts a new post into the database, avoiding duplicates.
func InsertRssPost(db *gorm.DB, post *model.RssPost) error {
	var count int64
	if err := db.Model(&model.RssPost{}).Where("link = ?", post.Link).Count(&count).Error; err != nil {
		return fmt.Errorf("could not check for existing post: %w", err)
	}

	if count > 0 {
		log.Printf("链接为 %s 的文章已存在，跳过", post.Link)
		return nil
	}

	if err := db.Create(post).Error; err != nil {
		return fmt.Errorf("could not insert post: %w", err)
	}

	log.Printf("已插入新文章: %s", post.Title)
	return nil
}

// GetPosts retrieves posts based on the provided query parameters.
func GetPosts(db *gorm.DB, query *model.PostQuery) ([]model.RssPost, int, error) {
	var posts []model.RssPost
	var total int64

	tx := db.Table("friend_rss_post AS p").Select("p.id, p.rss_id, p.title, p.link, p.description, p.time")
	if query.FriendLinkID != nil {
		tx = tx.Joins("JOIN friend_rss r ON p.rss_id = r.id").Where("r.friend_link_id = ?", *query.FriendLinkID)
	}
	if query.RssID != nil {
		tx = tx.Where("p.rss_id = ?", *query.RssID)
	}

	// Count total records
	countTx := tx
	if err := countTx.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("could not query total posts count: %w", err)
	}

	// Handle pagination
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		tx = tx.Limit(query.PageSize).Offset(offset)
	}

	if err := tx.Order("p.time DESC").Scan(&posts).Error; err != nil {
		return nil, 0, fmt.Errorf("could not query posts: %w", err)
	}

	return posts, int(total), nil
}
