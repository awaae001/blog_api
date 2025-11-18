package model

import "time"

// StatusData holds the statistical data for the system.
type StatusData struct {
	FriendLinkCount int `json:"friend_link_count"`
	RssCount        int `json:"rss_count"`
	RssPostCount    int `json:"rss_post_count"`
}

// SystemStatus represents the overall system status response.
type SystemStatus struct {
	Uptime string     `json:"uptime"`
	Data   StatusData `json:"data"`
	Time   time.Time  `json:"time"`
}
