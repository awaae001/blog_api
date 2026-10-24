package friendsRepositories

import (
	"blog_api/src/model"
	"fmt"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InsertFriendLinks inserts friend links from the configuration if they don't already exist.
func InsertFriendLinks(db *gorm.DB, friendLinks []model.FriendWebsite) error {
	var count int64
	if err := db.Model(&model.FriendWebsite{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		log.Println("[db][friend][init]检测到已有友链，跳过初始化")
		return nil
	}

	if len(friendLinks) == 0 {
		log.Println("No friend links to insert.")
		return nil
	}

	log.Println("[db][friend][init]Start inserting friend links...")

	for _, link := range friendLinks {
		newLink := model.FriendWebsite{
			Name:      link.Name,
			Link:      link.Link,
			Avatar:    link.Avatar,
			Info:      link.Info,
			Status:    "survival",
			EnableRss: true,
		}

		result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&newLink)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			log.Printf("[db][friend][init]友链 %s 已存在，跳过", link.Name)
			continue
		}

		log.Printf("[db][friend][init]已插入友链: %s", link.Name)
	}

	log.Println("[db][friend][init]Friend links insertion process completed.")
	return nil
}

// QueryFriendLinks provides a unified interface for querying friend links.
func QueryFriendLinks(db *gorm.DB, opts model.FriendLinkQueryOptions) (model.QueryFriendLinksResponse, error) {
	var resp model.QueryFriendLinksResponse
	baseQuery := db.Model(&model.FriendWebsite{})

	// Apply status filters
	if opts.Status != "" {
		baseQuery = baseQuery.Where("status = ?", opts.Status)
	}
	if len(opts.Statuses) > 0 {
		if opts.NotIn {
			baseQuery = baseQuery.Where("status NOT IN ?", opts.Statuses)
		} else {
			baseQuery = baseQuery.Where("status IN ?", opts.Statuses)
		}
	}

	if opts.IsDied != nil {
		baseQuery = baseQuery.Where("is_died = ?", *opts.IsDied)
	}
	if opts.SkipHealthCheck != nil {
		baseQuery = baseQuery.Where("skip_health_check = ?", *opts.SkipHealthCheck)
	}
	if opts.Email != "" {
		baseQuery = baseQuery.Where("email = ?", opts.Email)
	}

	// Apply search filter
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		baseQuery = baseQuery.Where("website_name LIKE ? OR website_url LIKE ? OR description LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := baseQuery.Count(&resp.Count).Error; err != nil {
		return resp, err
	}

	// The primary-key tie-breaker keeps offset pagination stable when a trigger
	// gives several rows the same second-resolution timestamp.
	query := baseQuery.Order("updated_at DESC").Order("id DESC")

	// Apply pagination and ordering
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}
	selectFields := "id, website_name, website_url, website_icon_url, description, email, times, status, is_died, enable_rss, skip_health_check, updated_at"
	if err := query.Select(selectFields).Find(&resp.Links).Error; err != nil {
		return resp, err
	}

	return resp, nil
}

// GetFriendLinkByID fetches a single friend link by ID.
func GetFriendLinkByID(db *gorm.DB, id int) (model.FriendWebsite, error) {
	var link model.FriendWebsite
	selectFields := "id, website_name, website_url, website_icon_url, description, email, times, status, is_died, enable_rss, skip_health_check, updated_at"
	err := db.Model(&model.FriendWebsite{}).Select(selectFields).Where("id = ?", id).First(&link).Error
	return link, err
}

// UpdateFriendLink updates the details of a friend link after crawling.
func UpdateFriendLink(db *gorm.DB, link model.FriendWebsite, result model.CrawlResult) error {
	success := result.Status == "survival"
	var reachedThreshold bool
	link.Times, link.Status, reachedThreshold = model.ComputeFailureState(
		link.Times,
		success,
		4,
		"survival",
		result.Status,
		"",
	)
	if success {
		link.IsDied = false
	} else if reachedThreshold {
		link.IsDied = true
	}

	if result.RedirectURL != "" {
		link.Link = result.RedirectURL
	}

	updates := map[string]interface{}{
		"website_url": link.Link,
		"description": gorm.Expr("CASE WHEN description = '' THEN ? ELSE description END", result.Description),
		"status":      link.Status,
		"times":       link.Times,
		"is_died":     link.IsDied,
	}

	// 仅当现有 icon 为空时才覆盖，避免已有 icon 被新结果替换
	if link.Avatar == "" && result.IconURL != "" {
		updates["website_icon_url"] = resolveAvatarURL(result.IconURL, link.Link)
	}

	if err := db.Model(&model.FriendWebsite{}).Where("id = ?", link.ID).Updates(updates).Error; err != nil {
		return err
	}

	log.Printf("为 ID  %d 更新友链. 状态: %s, 时间: %d, is_died: %t", link.ID, link.Status, link.Times, link.IsDied)
	return nil
}

// CreateFriendLink inserts a single new friend link into the database.
func CreateFriendLink(db *gorm.DB, link model.FriendWebsite) (int64, error) {
	status := link.Status
	if status == "" {
		status = "pending"
	}

	newLink := model.FriendWebsite{
		Name:      link.Name,
		Link:      link.Link,
		Avatar:    link.Avatar,
		Info:      link.Info,
		Email:     link.Email,
		Status:    status,
		EnableRss: link.EnableRss,
	}

	if err := db.Create(&newLink).Error; err != nil {
		return 0, err
	}

	log.Printf("[db][friend] 已插入新友链: %s，ID 为: %d", link.Name, newLink.ID)
	return int64(newLink.ID), nil
}

// DeleteFriendLink deletes a link and its RSS data. A non-nil ownerEmail
// restricts deletion to that owner, including when the email is empty.
func DeleteFriendLink(db *gorm.DB, id uint, ownerEmail *string) (model.FriendWebsite, error) {
	match := map[string]any{"id": id}
	if ownerEmail != nil {
		match["email"] = *ownerEmail
	}
	var deletedLink model.FriendWebsite
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(match).First(&deletedLink).Error; err != nil {
			return err
		}
		if err := DeleteRssFeedsByFriendLinkID(tx, deletedLink.ID); err != nil {
			return err
		}
		result := tx.Where(match).Delete(&deletedLink)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		return model.FriendWebsite{}, err
	}
	return deletedLink, nil
}

// UpdateFriendLinkByID updates a friend link by its ID and handles cascading deletes for RSS data if necessary.
func UpdateFriendLinkByID(db *gorm.DB, id uint, req model.EditFriendLinkReq) (int64, error) {
	if len(req.Data) == 0 {
		return 0, fmt.Errorf("no data provided for update")
	}

	updates := make(map[string]any, len(req.Data))
	for col, val := range req.Data {
		switch col {
		case "website_name", "website_url", "website_icon_url", "description", "email",
			"times", "status", "enable_rss", "is_died", "skip_health_check":
		default:
			continue
		}
		if text, ok := val.(string); ok && text == "" && !req.Opt.OverwriteIfBlank {
			continue
		}
		updates[col] = val
	}

	if len(updates) == 0 {
		log.Println("[db][friend] No valid fields to update after filtering.")
		return 0, nil
	}

	return updateFriendLink(db, id, nil, updates)
}

// UpdateFriendLinkByOwner updates owner-editable profile fields for a friend link.
// Ownership is included in the update predicate. Disabling RSS removes related data.
func UpdateFriendLinkByOwner(db *gorm.DB, id uint, email string, req model.EditFriendLinkReq) (int64, error) {
	if len(req.Data) == 0 {
		return 0, fmt.Errorf("no data provided for update")
	}

	updates := make(map[string]any, len(req.Data))
	for column, value := range req.Data {
		if !model.IsFriendOwnerField(column) {
			return 0, fmt.Errorf("field %q cannot be updated by owner", column)
		}
		if text, ok := value.(string); ok && text == "" && !req.Opt.OverwriteIfBlank {
			continue
		}
		updates[column] = value
	}
	if len(updates) == 0 {
		return 0, fmt.Errorf("no data provided for update")
	}

	return updateFriendLink(db, id, &email, updates)
}

func updateFriendLink(db *gorm.DB, id uint, ownerEmail *string, updates map[string]any) (int64, error) {
	var rowsAffected int64
	err := db.Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&model.FriendWebsite{}).Where("id = ?", id)
		if ownerEmail != nil {
			query = query.Where("email = ?", *ownerEmail)
		}
		result := query.Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		rowsAffected = result.RowsAffected
		if rowsAffected == 0 && ownerEmail != nil {
			return gorm.ErrRecordNotFound
		}
		// Ownership is checked by the update before deleting related RSS data.
		if enabled, ok := updates["enable_rss"].(bool); ok && !enabled && rowsAffected > 0 {
			return DeleteRssFeedsByFriendLinkID(tx, int(id))
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// FriendLinkExists checks if a friend link with the given ID exists.
func FriendLinkExists(db *gorm.DB, id int) (bool, error) {
	var count int64
	if err := db.Model(&model.FriendWebsite{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
