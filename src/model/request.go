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
