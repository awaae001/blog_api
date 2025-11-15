package model

import "time"

// FriendRss maps to the friend_rss table.
type FriendRss struct {
	ID           int       `json:"id"`
	FriendLinkID int       `json:"friend_link_id"`
	RssURL       string    `json:"rss_url"`
	Status       string    `json:"status"`
	UpdatedAt    time.Time `json:"updated_at"`
}
