package model

import "time"

// FriendWebsite 单个友链站点
type FriendWebsite struct {
	ID        int       `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Name      string    `json:"name" gorm:"column:website_name"`
	Link      string    `json:"link" gorm:"column:website_url"`
	Avatar    string    `json:"avatar" gorm:"column:website_icon_url"`
	Info      string    `json:"description" gorm:"column:description"`
	Email     string    `json:"email,omitempty" gorm:"column:email"`
	Times     int       `json:"times,omitempty" gorm:"column:times"`
	Status    string    `json:"status,omitempty" gorm:"column:status"`
	EnableRss bool      `json:"enable_rss,omitempty" gorm:"column:enable_rss"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

// TableName sets the insert table name for this struct type.
func (FriendWebsite) TableName() string {
	return "friend_link"
}

// FriendLinkQueryOptions defines the options for querying friend links.
type FriendLinkQueryOptions struct {
	Status   string   // Single status filter, e.g., "pending"
	Statuses []string // Multiple statuses for IN or NOT IN clauses, e.g., {"died", "ignored"}
	NotIn    bool     // If true, use NOT IN for Statuses
	Search   string   // Search keyword
	Offset   int      // Pagination offset
	Limit    int      // Pagination limit
	Count    bool     // If true, only return the count
}

// QueryFriendLinksResponse defines the response for the unified friend link query.
type QueryFriendLinksResponse struct {
	Links []FriendWebsite
	Count int64
}
