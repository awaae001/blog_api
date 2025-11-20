package repositories

import (
	"blog_api/src/model"
	"database/sql"
	"fmt"
	"log"
)

// InsertFriendRss 为友链插入新的 RSS 源，避免重复
func InsertFriendRss(db *sql.DB, friendLinkID int, rssURLs []string) error {
	if len(rssURLs) == 0 {
		return nil
	}

	log.Printf("开始为友链 ID: %d 插入 RSS 源", friendLinkID)

	stmt, err := db.Prepare("INSERT INTO friend_rss (friend_link_id, rss_url) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("无法准备 friend_rss 插入语句: %w", err)
	}
	defer stmt.Close()

	for _, rssURL := range rssURLs {
		var exists bool
		// 检查该友链是否已存在此 RSS 源
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM friend_rss WHERE friend_link_id = ? AND rss_url = ?)", friendLinkID, rssURL).Scan(&exists)
		if err != nil {
			log.Printf("无法检查友链 %d 的 RSS 源 %s 是否存在: %v", friendLinkID, rssURL, err)
			continue
		}

		if !exists {
			if _, err := stmt.Exec(friendLinkID, rssURL); err != nil {
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
func GetAllFriendRss(db *sql.DB) ([]model.FriendRss, error) {
	rows, err := db.Query("SELECT id, friend_link_id, rss_url, status, updated_at FROM friend_rss")
	if err != nil {
		return nil, fmt.Errorf("无法查询 friend_rss: %w", err)
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
		return nil, fmt.Errorf("遍历 friend_rss 行时出错: %w", err)
	}

	return rssFeeds, nil
}

// CreateFriendRss 创建新的 friend_rss 记录并返回其 ID
func CreateFriendRss(db *sql.DB, friendLinkID int, rssURL string) (int64, error) {
	// 首先检查该友链是否已存在此 RSS 源，避免重复
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM friend_rss WHERE friend_link_id = ? AND rss_url = ?)", friendLinkID, rssURL).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("无法检查已存在的 friend_rss: %w", err)
	}
	if exists {
		return 0, fmt.Errorf("RSS 地址 '%s' 已存在于友链 ID %d", rssURL, friendLinkID)
	}

	// 插入新记录
	stmt, err := db.Prepare("INSERT INTO friend_rss (friend_link_id, rss_url) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("无法准备 friend_rss 插入语句: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(friendLinkID, rssURL)
	if err != nil {
		return 0, fmt.Errorf("无法执行 friend_rss 插入语句: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("无法获取 friend_rss 最后插入 ID: %w", err)
	}

	log.Printf("成功为友链 ID %d 插入 RSS 源 %s，新 ID 为 %d", friendLinkID, rssURL, id)
	return id, nil
}
