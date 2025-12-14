package model

// DeleteFriendLinkReq defines the request body for deleting friend links
type DeleteFriendLinkReq struct {
	IDs []int `json:"ids" binding:"required"`
}

// EditFriendLinkReq defines the request body for editing a friend link.
type EditFriendLinkReq struct {
	ID   int                    `json:"id" binding:"required"`
	Data map[string]interface{} `json:"data" binding:"required"`
	Opt  struct {
		OverwriteIfBlank bool `json:"overwrite_if_blank"`
	} `json:"opt"`
}

// CreateFriendRssReq defines the request body for creating a friend rss link.
type CreateFriendRssReq struct {
	FriendLinkID int    `json:"friend_link_id" binding:"required"`
	RssURL       string `json:"rss_url" binding:"required"`
}

// PostQuery defines the query parameters for fetching posts.
type PostQuery struct {
	RssID        *int `form:"rss_id"`
	FriendLinkID *int `form:"friend_link_id"`
	Page         int  `form:"page"`
	PageSize     int  `form:"page_size"`
}

// DeleteFriendRssReq defines the request body for deleting a friend rss link.
type DeleteFriendRssReq struct {
	RssURL string `json:"rss_url" binding:"required"`
}

// CreateRssReq defines the request body for creating a rss link.
type CreateRssReq struct {
	FriendLinkID int    `json:"friend_link_id"`
	RssURL       string `json:"rss_url" binding:"required"`
}
