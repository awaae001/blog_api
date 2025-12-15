package model

import "runtime"

// StatusData holds the statistical data for the system.
type StatusData struct {
	FriendLinkCount int `json:"friend_link_count"`
	RssCount        int `json:"rss_count"`
	RssPostCount    int `json:"rss_post_count"`
}

// SystemStatus represents the overall system status response.
type SystemStatus struct {
	Uptime     string     `json:"uptime"`
	StatusData StatusData `json:"status_data"`
	Time       int64      `json:"time"`
}

// SystemStatusLog represents the data to be logged.
type SystemStatusLog struct {
	Timestamp      string           `json:"timestamp"`
	GoroutineCount int              `json:"goroutine_count"`
	MemStats       runtime.MemStats `json:"mem_stats"`
	DbStats        StatusData       `json:"db_stats"`
}
