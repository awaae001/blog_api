package model

import "time"

// RssPost represents an article from an RSS feed.
type RssPost struct {
	ID          int       `json:"id"`
	FriendRssID int       `json:"friend_rss_id"`
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	Time        time.Time `json:"time"`
}
