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

// RssPost represents an article from an RSS feed.
type RssPost struct {
	ID          int       `json:"id"`
	FriendRssID int       `json:"friend_rss_id"`
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	Time        time.Time `json:"time"`
}
