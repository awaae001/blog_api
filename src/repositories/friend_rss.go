package repositories

import (
	"blog_api/src/model"
	"fmt"
	"log"

	"gorm.io/gorm"
)

// InsertFriendRss 为友链插入新的 RSS 源，避免重复
func InsertFriendRss(db *gorm.DB, friendLinkID int, rssURLs []string) error {
	if len(rssURLs) == 0 {
		return nil
	}

	log.Printf("开始为友链 ID: %d 插入 RSS 源", friendLinkID)

	for _, rssURL := range rssURLs {
		var exists bool
		// 检查该友链是否已存在此 RSS 源
		err := db.Model(&model.FriendRss{}).
			Select("count(*) > 0").
			Where("friend_link_id = ? AND rss_url = ?", friendLinkID, rssURL).
			Find(&exists).Error
		if err != nil {
			log.Printf("无法检查友链 %d 的 RSS 源 %s 是否存在: %v", friendLinkID, rssURL, err)
			continue
		}

		if !exists {
			newRSS := model.FriendRss{
				FriendLinkID: friendLinkID,
				RssURL:       rssURL,
				Status:       "survival", // 避免空字符串触发 CHECK 约束
			}
			if err := db.Create(&newRSS).Error; err != nil {
				log.Printf("无法为友链 %d 插入 RSS 源 %s: %v", friendLinkID, rssURL, err)
			} else {
				log.Printf("已为友链 ID: %d 插入 RSS 源: %s", friendLinkID, rssURL)
			}
		} else {
			log.Printf("友链 %d 的 RSS 源 %s 已存在，跳过", friendLinkID, rssURL)
		}
	}

	log.Printf("友链 ID: %d 的 RSS 源插入流程完成", friendLinkID)
	return nil
}

// GetAllFriendRss 从数据库获取所有 RSS 源
func GetAllFriendRss(db *gorm.DB) ([]model.FriendRss, error) {
	var rssFeeds []model.FriendRss
	if err := db.Find(&rssFeeds).Error; err != nil {
		return nil, fmt.Errorf("无法查询 friend_rss: %w", err)
	}

	return rssFeeds, nil
}

// CreateFriendRss 创建新的 friend_rss 记录并返回其 ID
func CreateFriendRss(db *gorm.DB, friendLinkID int, rssURL string) (int64, error) {
	// 首先检查该友链是否已存在此 RSS 源，避免重复
	var count int64
	if err := db.Model(&model.FriendRss{}).Where("friend_link_id = ? AND rss_url = ?", friendLinkID, rssURL).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("无法检查已存在的 friend_rss: %w", err)
	}
	if count > 0 {
		return 0, fmt.Errorf("RSS 地址 '%s' 已存在于友链 ID %d", rssURL, friendLinkID)
	}

	// 插入新记录
	newRSS := model.FriendRss{FriendLinkID: friendLinkID, RssURL: rssURL}
	if err := db.Create(&newRSS).Error; err != nil {
		return 0, fmt.Errorf("无法执行 friend_rss 插入语句: %w", err)
	}

	log.Printf("成功为友链 ID %d 插入 RSS 源 %s，新 ID 为 %d", friendLinkID, rssURL, newRSS.ID)
	return int64(newRSS.ID), nil
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
