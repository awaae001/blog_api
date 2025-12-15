package model

// EditFriendLinkReq defines the request body for editing a friend link.
type EditFriendLinkReq struct {
	Data map[string]interface{} `json:"data" binding:"required"`
	Opt  struct {
		OverwriteIfBlank bool `json:"overwrite_if_blank"`
	} `json:"opt"`
}

// EditFriendRssReq defines the request body for editing a friend rss link.
type EditFriendRssReq struct {
	Data map[string]interface{} `json:"data" binding:"required"`
}

// PostQuery defines the query parameters for fetching posts.
type PostQuery struct {
	RssID        *int `form:"rss_id"`
	FriendLinkID *int `form:"friend_link_id"`
	Page         int  `form:"page"`
	PageSize     int  `form:"page_size"`
}

// CreateRssReq defines the request body for creating a rss link.
type CreateRssReq struct {
	FriendLinkID int    `json:"friend_link_id"`
	RssURL       string `json:"rss_url" binding:"required"`
	Name         string `json:"name"`
}
