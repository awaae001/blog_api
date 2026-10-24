package friendsRepositories

import (
	"blog_api/src/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InsertRssPosts inserts posts and silently skips those already stored, either
// under the same (feed_id, guid) or with the same link. It returns how many
// rows were actually inserted.
func InsertRssPosts(db *gorm.DB, posts []model.RssPost) (int, error) {
	inserted := 0
	for start := 0; start < len(posts); start += 100 {
		end := min(start+100, len(posts))
		result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(posts[start:end])
		if result.Error != nil {
			return inserted, result.Error
		}
		inserted += int(result.RowsAffected)
	}
	return inserted, nil
}

// DeleteRssPostByID deletes one RSS article by its primary key.
// It returns zero when the article does not exist.
func DeleteRssPostByID(db *gorm.DB, id int) (int64, error) {
	result := db.Delete(&model.RssPost{}, id)
	return result.RowsAffected, result.Error
}

// GetPosts retrieves posts based on the provided query parameters.
func GetPosts(db *gorm.DB, query *model.PostQuery) ([]model.RssPost, int, error) {
	var posts []model.RssPost
	var total int64

	baseTx := db.Table("friend_rss_post AS p")
	if query.FriendLinkID != nil {
		baseTx = baseTx.Joins("JOIN friend_rss f ON p.feed_id = f.id").Where("f.friend_link_id = ?", *query.FriendLinkID)
	}
	if query.RssID != nil {
		baseTx = baseTx.Where("p.feed_id = ?", *query.RssID)
	}

	if err := baseTx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataTx := baseTx.Select("p.id, p.feed_id, p.title, p.link, p.description, p.author, p.published_at")
	if query.Page > 0 && query.PageSize > 0 {
		dataTx = dataTx.Limit(query.PageSize).Offset((query.Page - 1) * query.PageSize)
	}

	err := dataTx.Order("p.published_at DESC").Order("p.id DESC").Scan(&posts).Error
	return posts, int(total), err
}
