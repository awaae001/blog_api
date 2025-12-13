package repositories

import (
	"blog_api/src/model"
	"fmt"
	"log"

	"gorm.io/gorm"
)

// QueryFriendRss provides a unified interface for querying friend RSS feeds.
func QueryFriendRss(db *gorm.DB, opts model.FriendRssQueryOptions) (model.QueryFriendRssResponse, error) {
	var resp model.QueryFriendRssResponse
	query := db.Model(&model.FriendRss{})

	if opts.FriendLinkID > 0 {
		query = query.Where("friend_link_id = ?", opts.FriendLinkID)
	}
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}

	// Special case for getting all valid RSS feeds
	if opts.Status == "valid" {
		query = db.Table("friend_rss").
			Joins("JOIN friend_link ON friend_link.id = friend_rss.friend_link_id").
			Where("friend_link.status NOT IN ?", []string{"ignored", "died"})
	}

	if opts.Count {
		if err := query.Count(&resp.Count).Error; err != nil {
			return resp, fmt.Errorf("could not count friend rss feeds: %w", err)
		}
		return resp, nil
	}

	if err := query.Find(&resp.Feeds).Error; err != nil {
		return resp, fmt.Errorf("could not query friend rss feeds: %w", err)
	}

	return resp, nil
}

// CreateFriendRssFeeds creates new friend_rss records from a slice of URLs, avoiding duplicates.
func CreateFriendRssFeeds(db *gorm.DB, friendLinkID int, rssURLs []string) ([]model.FriendRss, error) {
	if len(rssURLs) == 0 {
		return nil, nil
	}

	var createdFeeds []model.FriendRss
	for _, rssURL := range rssURLs {
		var existing model.FriendRss
		err := db.Where("friend_link_id = ? AND rss_url = ?", friendLinkID, rssURL).First(&existing).Error
		if err == nil {
			log.Printf("RSS feed '%s' already exists for friend link ID %d, skipping.", rssURL, friendLinkID)
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check for existing RSS feed: %w", err)
		}

		newRSS := model.FriendRss{
			FriendLinkID: friendLinkID,
			RssURL:       rssURL,
			Status:       "survival",
		}
		if err := db.Create(&newRSS).Error; err != nil {
			log.Printf("Failed to insert RSS feed '%s' for friend link ID %d: %v", rssURL, friendLinkID, err)
			continue
		}
		createdFeeds = append(createdFeeds, newRSS)
		log.Printf("Successfully inserted RSS feed '%s' for friend link ID %d.", rssURL, friendLinkID)
	}

	return createdFeeds, nil
}

// DeleteFriendRssByURL deletes a friend_rss entry and all associated posts by its URL.
func DeleteFriendRssByURL(db *gorm.DB, rssURL string) (int64, error) {
	var deletedID int64
	err := db.Transaction(func(tx *gorm.DB) error {
		var rss model.FriendRss
		if err := tx.Where("rss_url = ?", rssURL).First(&rss).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("未找到 RSS URL: %s", rssURL)
			}
			return fmt.Errorf("查询 RSS ID 失败: %w", err)
		}

		if err := DeleteRssPostsByRssIDWithTx(tx, int64(rss.ID)); err != nil {
			return fmt.Errorf("删除关联的 RSS 文章失败: %w", err)
		}

		if err := tx.Delete(&rss).Error; err != nil {
			return fmt.Errorf("删除 RSS 源失败: %w", err)
		}

		deletedID = int64(rss.ID)
		return nil
	})
	if err != nil {
		return 0, err
	}

	log.Printf("成功删除 RSS 源 %s 及其所有文章，ID 为 %d", rssURL, deletedID)
	return deletedID, nil
}

// DeleteRssDataByFriendLinkID deletes all RSS feeds and their posts for a given friend_link_id within a transaction.
func DeleteRssDataByFriendLinkID(tx *gorm.DB, friendLinkID int) error {
	// Find all RSS feeds associated with the friend link
	var rssFeeds []model.FriendRss
	if err := tx.Where("friend_link_id = ?", friendLinkID).Find(&rssFeeds).Error; err != nil {
		return fmt.Errorf("could not query rss feeds for friend_link_id %d: %w", friendLinkID, err)
	}

	if len(rssFeeds) == 0 {
		log.Printf("No RSS feeds to delete for friend_link_id %d", friendLinkID)
		return nil // Nothing to do
	}

	// Collect all RSS feed IDs
	rssIDs := make([]int, len(rssFeeds))
	for i, feed := range rssFeeds {
		rssIDs[i] = feed.ID
	}

	// Delete all posts associated with these RSS feeds
	if err := tx.Where("friend_rss_id IN ?", rssIDs).Delete(&model.RssPost{}).Error; err != nil {
		return fmt.Errorf("could not delete rss posts for friend_link_id %d: %w", friendLinkID, err)
	}

	// Delete the RSS feeds themselves
	if err := tx.Where("friend_link_id = ?", friendLinkID).Delete(&model.FriendRss{}).Error; err != nil {
		return fmt.Errorf("could not delete rss feeds for friend_link_id %d: %w", friendLinkID, err)
	}

	log.Printf("Successfully deleted %d RSS feeds and their posts for friend_link_id %d", len(rssFeeds), friendLinkID)
	return nil
}
