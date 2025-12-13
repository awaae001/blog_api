package repositories

import (
	"blog_api/src/model"
	"fmt"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InsertFriendLinks inserts friend links from the configuration if they don't already exist.
func InsertFriendLinks(db *gorm.DB, friendLinks []model.FriendWebsite) error {
	if len(friendLinks) == 0 {
		log.Println("No friend links to insert.")
		return nil
	}

	log.Println("[db][friend][init]Start inserting friend links...")

	for _, link := range friendLinks {
		var exists bool
		err := db.Model(&model.FriendWebsite{}).Select("count(*) > 0").Where("website_url = ?", link.Link).Find(&exists).Error
		if err != nil {
			log.Printf("[db][friend][ERR]无法检查已存在的链接 %s: %v", link.Link, err)
			continue // Or return error, depending on desired strictness
		}

		if !exists {
			newLink := model.FriendWebsite{
				Name:      link.Name,
				Link:      link.Link,
				Avatar:    link.Avatar,
				Info:      link.Info,
				Status:    "survival",
				EnableRss: true,
			}
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&newLink).Error; err != nil {
				log.Printf("[db][friend][ERR]无法插入友链 %s: %v", link.Name, err)
				// Decide if one failure should stop the whole process
			} else {
				log.Printf("[db][friend][init]已插入友链: %s", link.Name)
			}
		} else {
			log.Printf("[db][friend][init]友链 %s 已存在，跳过", link.Name)
		}
	}

	log.Println("[db][friend][init]Friend links insertion process completed.")
	return nil
}

// GetAllFriendLinks retrieves all friend links from the database, excluding 'died' and 'ignored' ones.
func GetAllFriendLinks(db *gorm.DB) ([]model.FriendWebsite, error) {
	var links []model.FriendWebsite
	if err := db.Where("status NOT IN ?", []string{"died", "ignored"}).
		Select("id, website_name, website_url, website_icon_url, description, times, status").
		Find(&links).Error; err != nil {
		return nil, fmt.Errorf("could not query friend links: %w", err)
	}

	return links, nil
}

// GetAllDiedFriendLinks retrieves all friend links from the database with 'died' status.
func GetAllDiedFriendLinks(db *gorm.DB) ([]model.FriendWebsite, error) {
	var links []model.FriendWebsite
	if err := db.Where("status = ?", "died").
		Select("id, website_name, website_url, website_icon_url, description, times, status").
		Find(&links).Error; err != nil {
		return nil, fmt.Errorf("could not query died friend links: %w", err)
	}

	return links, nil
}

// GetFriendLinksWithFilter retrieves friend links with filtering and pagination support.
func GetFriendLinksWithFilter(db *gorm.DB, status string, search string, offset int, limit int) ([]model.FriendWebsite, error) {
	query := db.Model(&model.FriendWebsite{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("website_name LIKE ? OR website_url LIKE ? OR description LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	var links []model.FriendWebsite
	if err := query.Select("id, website_name, website_url, website_icon_url, description, email, times, status, enable_rss, updated_at").Order("updated_at DESC").Offset(offset).Limit(limit).Find(&links).Error; err != nil {
		return nil, fmt.Errorf("could not query friend links: %w", err)
	}

	return links, nil
}

// CountFriendLinks counts the total number of friend links matching the filter.
func CountFriendLinks(db *gorm.DB, status string, search string) (int, error) {
	query := db.Model(&model.FriendWebsite{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("website_name LIKE ? OR website_url LIKE ? OR description LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("could not count friend links: %w", err)
	}

	return int(count), nil
}

// UpdateFriendLink updates the details of a friend link after crawling.
func UpdateFriendLink(db *gorm.DB, link model.FriendWebsite, result model.CrawlResult) error {
	if result.Status == "survival" {
		link.Times = 0 // Reset times on success
	} else {
		link.Times++
	}

	if link.Times >= 4 {
		link.Status = "died"
	} else {
		link.Status = result.Status
	}
	if result.RedirectURL != "" {
		link.Link = result.RedirectURL
	}

	updates := map[string]interface{}{
		"website_url": link.Link,
		"description": gorm.Expr("CASE WHEN description = '' THEN ? ELSE description END", result.Description),
		"status":      link.Status,
		"times":       link.Times,
		"updated_at":  gorm.Expr("CURRENT_TIMESTAMP"),
	}

	// 仅当现有 icon 为空时才覆盖，避免已有 icon 被新结果替换
	if link.Avatar == "" && result.IconURL != "" {
		updates["website_icon_url"] = result.IconURL
	}

	if err := db.Model(&model.FriendWebsite{}).Where("id = ?", link.ID).Updates(updates).Error; err != nil {
		return fmt.Errorf("could not update friend link with id %d: %w", link.ID, err)
	}

	log.Printf("为 ID  %d 更新友链. 状态: %s, 时间: %d", link.ID, link.Status, link.Times)
	return nil
}

// CreateFriendLink inserts a single new friend link into the database.
func CreateFriendLink(db *gorm.DB, link model.FriendWebsite) (int64, error) {
	newLink := model.FriendWebsite{
		Name:      link.Name,
		Link:      link.Link,
		Avatar:    link.Avatar,
		Info:      link.Info,
		Email:     link.Email,
		Status:    "pending",
		EnableRss: link.EnableRss,
	}

	if err := db.Create(&newLink).Error; err != nil {
		return 0, fmt.Errorf("could not execute insert statement for friend link: %w", err)
	}

	log.Printf("[db][friend] 已插入新友链: %s，ID 为: %d", link.Name, newLink.ID)
	return int64(newLink.ID), nil
}

// DeleteFriendLinksByID deletes friend links by their IDs and returns the deleted links.
func DeleteFriendLinksByID(db *gorm.DB, ids []int) ([]model.FriendWebsite, error) {
	if len(ids) == 0 {
		return []model.FriendWebsite{}, nil
	}

	var deletedLinks []model.FriendWebsite
	var rowsDeleted int64
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id IN ?", ids).Find(&deletedLinks).Error; err != nil {
			return fmt.Errorf("could not query friend links for deletion: %w", err)
		}

		res := tx.Where("id IN ?", ids).Delete(&model.FriendWebsite{})
		if res.Error != nil {
			return fmt.Errorf("could not delete friend links: %w", res.Error)
		}
		rowsDeleted = res.RowsAffected
		return nil
	})
	if err != nil {
		return nil, err
	}

	log.Printf("[db][friend] 已删除 %d 个友链", rowsDeleted)

	return deletedLinks, nil
}

// UpdateFriendLinkByID updates a friend link by its ID and handles cascading deletes for RSS data if necessary.
func UpdateFriendLinkByID(db *gorm.DB, req model.EditFriendLinkReq) (int64, error) {
	if len(req.Data) == 0 {
		return 0, fmt.Errorf("no data provided for update")
	}

	// Whitelist of updatable columns
	updatableColumns := map[string]bool{
		"website_name":     true,
		"website_url":      true,
		"website_icon_url": true,
		"description":      true,
		"email":            true,
		"status":           true,
		"enable_rss":       true,
	}

	updates := map[string]interface{}{}
	for col, val := range req.Data {
		if !updatableColumns[col] {
			log.Printf("[db][friend][WARN] 尝试更新不可更新的列: %s", col)
			continue
		}
		if !req.Opt.OverwriteIfBlank {
			if s, ok := val.(string); ok && s == "" {
				continue
			}
		}
		updates[col] = val
	}

	if len(updates) == 0 {
		log.Println("[db][friend] No valid fields to update after filtering.")
		return 0, nil
	}

	updates["updated_at"] = gorm.Expr("CURRENT_TIMESTAMP")

	// Check if enable_rss is being set to false
	disableRss := false
	if val, ok := updates["enable_rss"].(bool); ok && !val {
		disableRss = true
	}

	var rowsAffected int64
	err := db.Transaction(func(tx *gorm.DB) error {
		// If disabling RSS, delete related data first
		if disableRss {
			if err := DeleteRssDataByFriendLinkID(tx, req.ID); err != nil {
				return err
			}
		}

		// Perform the update
		result := tx.Model(&model.FriendWebsite{}).Where("id = ?", req.ID).Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("could not execute update for friend link id %d: %w", req.ID, result.Error)
		}
		rowsAffected = result.RowsAffected
		return nil
	})

	if err != nil {
		return 0, err
	}

	log.Printf("[db][friend] 为 ID: %d 更新友链. Rows affected: %d", req.ID, rowsAffected)
	return rowsAffected, nil
}

// FriendLinkExists checks if a friend link with the given ID exists.
func FriendLinkExists(db *gorm.DB, id int) (bool, error) {
	var count int64
	if err := db.Model(&model.FriendWebsite{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("could not check for existing friend_link: %w", err)
	}
	return count > 0, nil
}
