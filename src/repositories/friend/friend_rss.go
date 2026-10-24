package friendsRepositories

import (
	"blog_api/src/model"

	"gorm.io/gorm"
)

// QueryRssFeeds returns one page of feeds for the admin list.
func QueryRssFeeds(db *gorm.DB, opts model.RssFeedQueryOptions) ([]model.RssFeed, int64, error) {
	query := db.Model(&model.RssFeed{})
	if opts.FriendLinkID > 0 {
		query = query.Where("friend_link_id = ?", opts.FriendLinkID)
	}
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if opts.Page > 0 && opts.PageSize > 0 {
		query = query.Offset((opts.Page - 1) * opts.PageSize).Limit(opts.PageSize)
	}

	var feeds []model.RssFeed
	err := query.Order("updated_at DESC").Order("id DESC").Find(&feeds).Error
	return feeds, total, err
}

// DueRssFeeds returns feeds whose next fetch time has passed. Paused feeds and
// feeds owned by a died friend link are skipped.
func DueRssFeeds(db *gorm.DB, now int64, limit int) ([]model.RssFeed, error) {
	var feeds []model.RssFeed
	err := db.Table("friend_rss AS f").
		Select("f.*").
		Joins("LEFT JOIN friend_link l ON l.id = f.friend_link_id").
		Where("f.status <> ?", model.FeedStatusPaused).
		Where("f.next_fetch_at <= ?", now).
		Where("l.id IS NULL OR l.is_died = ?", false).
		Order("f.next_fetch_at ASC").
		Limit(limit).
		Find(&feeds).Error
	return feeds, err
}

// GetRssFeedByID returns one feed by its primary key.
func GetRssFeedByID(db *gorm.DB, id int) (model.RssFeed, error) {
	var feed model.RssFeed
	err := db.First(&feed, id).Error
	return feed, err
}

// RssFeedURLTaken reports whether another feed already uses feedURL.
func RssFeedURLTaken(db *gorm.DB, feedURL string, exceptID int) (bool, error) {
	var count int64
	err := db.Model(&model.RssFeed{}).Where("feed_url = ? AND id <> ?", feedURL, exceptID).Count(&count).Error
	return count > 0, err
}

// FriendLinkHasRssFeeds reports whether a friend link already owns a feed.
func FriendLinkHasRssFeeds(db *gorm.DB, friendLinkID int) (bool, error) {
	var count int64
	err := db.Model(&model.RssFeed{}).Where("friend_link_id = ?", friendLinkID).Count(&count).Error
	return count > 0, err
}

// CreateRssFeed inserts a new feed. feed_url is unique; callers check
// RssFeedURLTaken first for a friendly error.
func CreateRssFeed(db *gorm.DB, feed *model.RssFeed) error {
	return db.Create(feed).Error
}

// UpdateRssFeed applies column updates to one feed and reports whether it exists.
func UpdateRssFeed(db *gorm.DB, id int, updates map[string]any) (bool, error) {
	result := db.Model(&model.RssFeed{ID: id}).Updates(updates)
	return result.RowsAffected > 0, result.Error
}

// DeleteRssFeedByID deletes one feed; its posts go with it via ON DELETE CASCADE.
func DeleteRssFeedByID(db *gorm.DB, id int) (int64, error) {
	result := db.Delete(&model.RssFeed{}, id)
	return result.RowsAffected, result.Error
}

// DeleteRssFeedsByFriendLinkID drops every feed owned by a friend link.
func DeleteRssFeedsByFriendLinkID(tx *gorm.DB, friendLinkID int) error {
	return tx.Where("friend_link_id = ?", friendLinkID).Delete(&model.RssFeed{}).Error
}
