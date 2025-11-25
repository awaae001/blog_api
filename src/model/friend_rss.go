package model

import "time"

// FriendRss maps to the friend_rss table.
type FriendRss struct {
	ID           int       `json:"id" gorm:"column:id;primaryKey"`
	FriendLinkID int       `json:"friend_link_id" gorm:"column:friend_link_id"`
	RssURL       string    `json:"rss_url" gorm:"column:rss_url"`
	Status       string    `json:"status" gorm:"column:status"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// RssPost represents an article from an RSS feed.
type RssPost struct {
	ID          int       `json:"id" gorm:"column:id;primaryKey"`
	FriendRssID int       `json:"friend_rss_id" gorm:"column:friend_rss_id"`
	Title       string    `json:"title" gorm:"column:title"`
	Link        string    `json:"link" gorm:"column:link"`
	Description string    `json:"description" gorm:"column:description"`
	Time        time.Time `json:"time" gorm:"column:time"`
}

// TableName sets the table name for FriendRss.
func (FriendRss) TableName() string {
	return "friend_rss"
}

// TableName sets the table name for RssPost.
func (RssPost) TableName() string {
	return "friend_rss_post"
}
