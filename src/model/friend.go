package model

// FriendWebsite 单个友链站点
type FriendWebsite struct {
	ID              int    `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Name            string `json:"name" gorm:"column:website_name"`
	Link            string `json:"link" gorm:"column:website_url"`
	Avatar          string `json:"avatar" gorm:"column:website_icon_url"`
	Info            string `json:"description" gorm:"column:description"`
	Email           string `json:"email,omitempty" gorm:"column:email"`
	Times           int    `json:"times,omitempty" gorm:"column:times"`
	Status          string `json:"status,omitempty" gorm:"column:status"`
	IsDied          bool   `json:"is_died,omitempty" gorm:"column:is_died"`
	EnableRss       bool   `json:"enable_rss,omitempty" gorm:"column:enable_rss"`
	SkipHealthCheck bool   `json:"skip_health_check,omitempty" gorm:"column:skip_health_check"`
	UpdatedAt       int64  `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

// TableName sets the insert table name for this struct type.
func (FriendWebsite) TableName() string {
	return "friend_link"
}

// IsFriendOwnerField defines the fields an email owner may update.
func IsFriendOwnerField(field string) bool {
	switch field {
	case "website_name", "website_url", "website_icon_url", "description", "enable_rss":
		return true
	}
	return false
}

// FriendLinkQueryOptions defines the options for querying friend links.
type FriendLinkQueryOptions struct {
	Status          string
	Statuses        []string
	Email           string
	IsDied          *bool
	SkipHealthCheck *bool
	NotIn           bool
	Search          string
	Offset          int
	Limit           int
}

// QueryFriendLinksResponse defines the response for the unified friend link query.
type QueryFriendLinksResponse struct {
	Links []FriendWebsite
	Count int64
}

// RSS feed status values persisted in friend_rss.status.
const (
	// FeedStatusActive is fetched on its own schedule.
	FeedStatusActive = "active"
	// FeedStatusPaused is never fetched by the scheduler.
	FeedStatusPaused = "paused"
	// FeedStatusDead failed too many times; it is still probed with a long
	// interval and revives on the first successful fetch.
	FeedStatusDead = "dead"
)

// RssFeed maps to the friend_rss table: one subscription and its fetch schedule.
type RssFeed struct {
	ID            int    `json:"id" gorm:"column:id;primaryKey"`
	FriendLinkID  *int   `json:"friend_link_id" gorm:"column:friend_link_id"`
	Name          string `json:"name" gorm:"column:name"`
	FeedURL       string `json:"feed_url" gorm:"column:feed_url"`
	SiteURL       string `json:"site_url" gorm:"column:site_url"`
	FeedType      string `json:"feed_type" gorm:"column:feed_type"`
	Status        string `json:"status" gorm:"column:status"`
	ErrorCount    int    `json:"error_count" gorm:"column:error_count"`
	LastError     string `json:"last_error" gorm:"column:last_error"`
	ETag          string `json:"-" gorm:"column:etag"`
	LastModified  string `json:"-" gorm:"column:last_modified"`
	FetchInterval int64  `json:"fetch_interval" gorm:"column:fetch_interval"`
	NextFetchAt   int64  `json:"next_fetch_at" gorm:"column:next_fetch_at"`
	LastFetchedAt int64  `json:"last_fetched_at" gorm:"column:last_fetched_at"`
	LastSuccessAt int64  `json:"last_success_at" gorm:"column:last_success_at"`
	CreatedAt     int64  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     int64  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName sets the table name for RssFeed.
func (RssFeed) TableName() string {
	return "friend_rss"
}

// RssPost is one article of a feed. The JSON keys rss_id/time are kept for the
// public API consumers.
type RssPost struct {
	ID          int    `json:"id" gorm:"column:id;primaryKey"`
	FeedID      int    `json:"rss_id" gorm:"column:feed_id"`
	GUID        string `json:"-" gorm:"column:guid"`
	Title       string `json:"title" gorm:"column:title"`
	Link        string `json:"link" gorm:"column:link"`
	Description string `json:"description" gorm:"column:description"`
	Author      string `json:"author" gorm:"column:author"`
	PublishedAt int64  `json:"time" gorm:"column:published_at"`
	FetchedAt   int64  `json:"-" gorm:"column:fetched_at"`
}

// TableName sets the table name for RssPost.
func (RssPost) TableName() string {
	return "friend_rss_post"
}

// RssFeedQueryOptions filters the admin feed list.
type RssFeedQueryOptions struct {
	FriendLinkID int    // > 0 filters by owner
	Status       string // active / paused / dead
	Page         int
	PageSize     int
}

// RssFetchResult summarizes one completed RSS fetch.
type RssFetchResult struct {
	NotModified   bool `json:"not_modified"`
	CheckedItems  int  `json:"checked_items"`
	InsertedItems int  `json:"inserted_items"`
}
