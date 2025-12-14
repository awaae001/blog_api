package repositories

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

// GetPostsByFriendLinkID retrieves all posts associated with a given friend_link_id.
func GetPostsByFriendLinkID(db *gorm.DB, friendLinkID int) ([]model.RssPost, error) {
	var posts []model.RssPost
	if err := db.Table("friend_rss_post AS p").
		Select("p.id, p.rss_id, p.title, p.link, p.description, p.time").
		Joins("JOIN friend_rss r ON p.rss_id = r.id").
		Where("r.friend_link_id = ?", friendLinkID).
		Order("p.time DESC").
		Scan(&posts).Error; err != nil {
		return nil, fmt.Errorf("could not query posts by friend_link_id: %w", err)
	}

	return posts, nil
}

// GetAllPosts retrieves all posts with pagination.
func GetAllPosts(db *gorm.DB, page, pageSize int) ([]model.RssPost, int, error) {
	var total int64
	if err := db.Model(&model.RssPost{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("could not query total posts count: %w", err)
	}

	var posts []model.RssPost
	offset := (page - 1) * pageSize
	if err := db.Model(&model.RssPost{}).
		Order("time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&posts).Error; err != nil {
		return nil, 0, fmt.Errorf("could not query posts with pagination: %w", err)
	}

	return posts, int(total), nil
}
