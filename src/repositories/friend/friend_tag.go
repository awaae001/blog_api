package friendsRepositories

import (
	"blog_api/src/model"
	"blog_api/src/repositories"
	"errors"

	"gorm.io/gorm"
)

var ErrInvalidTagIDs = errors.New("tag_ids must contain existing positive tag IDs")

func QueryFriendTags(db *gorm.DB, includeEmpty bool) ([]model.FriendTagWithCount, error) {
	tags := make([]model.FriendTagWithCount, 0)
	query := db.Table("friend_tag AS t").
		Select("t.*, COUNT(lt.friend_link_id) AS link_count").
		Joins("LEFT JOIN friend_link_tag AS lt ON lt.tag_id = t.id").
		Group("t.id").Order("t.sort_order, t.id")
	if !includeEmpty {
		query = query.Having("COUNT(lt.friend_link_id) > 0")
	}
	err := query.Scan(&tags).Error
	return tags, err
}

func CreateFriendTag(db *gorm.DB, tag *model.FriendTag) error {
	return db.Create(tag).Error
}

func UpdateFriendTag(db *gorm.DB, id int, updates map[string]any) (model.FriendTag, error) {
	var tag model.FriendTag
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := repositories.RecordError(tx.Model(&model.FriendTag{}).Where("id = ?", id).Updates(updates)); err != nil {
			return err
		}
		return tx.First(&tag, id).Error
	})
	return tag, err
}

func DeleteFriendTag(db *gorm.DB, id int) error {
	// Foreign keys cascade assignment deletion, never friend link deletion.
	return repositories.RecordError(db.Delete(&model.FriendTag{}, id))
}

// LoadFriendTags fetches assignments for the entire page in one query.
func LoadFriendTags(db *gorm.DB, links []model.FriendWebsite) error {
	if len(links) == 0 {
		return nil
	}
	ids := make([]int, 0, len(links))
	positions := make(map[int]int, len(links))
	for i := range links {
		ids = append(ids, links[i].ID)
		positions[links[i].ID] = i
		links[i].Tags = make([]model.FriendTag, 0)
	}
	var rows []struct {
		model.FriendTag
		FriendLinkID int
	}
	if err := db.Table("friend_tag AS t").Select("t.*, lt.friend_link_id").
		Joins("JOIN friend_link_tag AS lt ON lt.tag_id = t.id").
		Where("lt.friend_link_id IN ?", ids).Order("t.sort_order, t.id").Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		i := positions[row.FriendLinkID]
		links[i].Tags = append(links[i].Tags, row.FriendTag)
	}
	return nil
}

func validateFriendTags(db *gorm.DB, ids []int) ([]int, error) {
	unique := make([]int, 0, len(ids))
	seen := make(map[int]bool, len(ids))
	for _, id := range ids {
		if id < 1 {
			return nil, ErrInvalidTagIDs
		}
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	if len(unique) > 0 {
		var count int64
		if err := db.Model(&model.FriendTag{}).Where("id IN ?", unique).Count(&count).Error; err != nil {
			return nil, err
		}
		if count != int64(len(unique)) {
			return nil, ErrInvalidTagIDs
		}
	}
	return unique, nil
}

func insertFriendTags(db *gorm.DB, friendID int, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	rows := make([]model.FriendLinkTag, 0, len(ids))
	for _, id := range ids {
		rows = append(rows, model.FriendLinkTag{FriendLinkID: friendID, TagID: id})
	}
	return db.Create(&rows).Error
}

// ReplaceFriendTags validates before deleting, so invalid input preserves assignments.
func ReplaceFriendTags(db *gorm.DB, friendID int, ids []int) ([]model.FriendTag, error) {
	links := []model.FriendWebsite{{ID: friendID}}
	err := db.Transaction(func(tx *gorm.DB) error {
		var link model.FriendWebsite
		if err := tx.Select("id").First(&link, friendID).Error; err != nil {
			return err
		}
		unique, err := validateFriendTags(tx, ids)
		if err != nil {
			return err
		}
		if err := tx.Where("friend_link_id = ?", friendID).Delete(&model.FriendLinkTag{}).Error; err != nil {
			return err
		}
		if err := insertFriendTags(tx, friendID, unique); err != nil {
			return err
		}
		return LoadFriendTags(tx, links)
	})
	return links[0].Tags, err
}

// CreateFriendLinkWithTags commits the profile and assignments together.
func CreateFriendLinkWithTags(db *gorm.DB, link model.FriendWebsite, ids []int) (int64, error) {
	var id int64
	err := db.Transaction(func(tx *gorm.DB) error {
		unique, err := validateFriendTags(tx, ids)
		if err != nil {
			return err
		}
		id, err = CreateFriendLink(tx, link)
		if err != nil {
			return err
		}
		return insertFriendTags(tx, int(id), unique)
	})
	return id, err
}
