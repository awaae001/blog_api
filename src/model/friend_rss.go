package model

// FriendRss maps to the friend_rss table.
type FriendRss struct {
	ID           int    `json:"id" gorm:"column:id;primaryKey"`
	FriendLinkID int    `json:"friend_link_id" gorm:"column:friend_link_id"`
	Name         string `json:"name" gorm:"column:name"`
	RssURL       string `json:"rss_url" gorm:"column:rss_url"`
	Status       string `json:"status" gorm:"column:status"`
	UpdatedAt    int64  `json:"updated_at" gorm:"column:updated_at"`
}

// RssPost represents an article from an RSS feed.
type RssPost struct {
	ID          int    `json:"id" gorm:"column:id;primaryKey"`
	RssID       int    `json:"rss_id" gorm:"column:rss_id"`
	Title       string `json:"title" gorm:"column:title"`
	Link        string `json:"link" gorm:"column:link"`
	Description string `json:"description" gorm:"column:description"`
	Time        int64  `json:"time" gorm:"column:time"`
}

// TableName sets the table name for FriendRss.
func (FriendRss) TableName() string {
	return "friend_rss"
}

// TableName sets the table name for RssPost.
func (RssPost) TableName() string {
	return "friend_rss_post"
}

// FriendRssQueryOptions defines the options for querying friend RSS feeds.
type FriendRssQueryOptions struct {
	FriendLinkID int    // Filter by friend link ID
	Status       string // Filter by status
	Page         int    // Page number for pagination
	PageSize     int    // Number of items per page
}

// QueryFriendRssResponse defines the response for the unified friend RSS query.
type QueryFriendRssResponse struct {
	Feeds []FriendRss `json:"feeds"`
	Total int64       `json:"total"`
}
