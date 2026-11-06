package model

import "encoding/json"

// CreateFriendLinkReq accepts initial tags only on the admin API.
type CreateFriendLinkReq struct {
	FriendWebsite
	TagIDs []int `json:"tag_ids" binding:"omitempty,dive,gt=0"`
}

type CreateFriendTagReq struct {
	Name      string `json:"name" binding:"required"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
}

// Nil fields stay untouched; empty color and zero sort_order are valid updates.
type UpdateFriendTagReq struct {
	Name      *string `json:"name"`
	Color     *string `json:"color"`
	SortOrder *int    `json:"sort_order"`
}

// An explicit empty array clears tags; missing and null arrays are invalid.
type ReplaceFriendTagsReq struct {
	TagIDs []int `json:"tag_ids" binding:"required,dive,gt=0"`
}

// EditFriendLinkReq defines the request body for editing a friend link.
type EditFriendLinkReq struct {
	Data map[string]interface{} `json:"data" binding:"required"`
	Opt  struct {
		OverwriteIfBlank bool `json:"overwrite_if_blank"`
	} `json:"opt"`
}

// EditRssFeedReq updates one feed. Nil fields stay untouched; FriendLinkID 0
// clears the owner.
type EditRssFeedReq struct {
	Name         *string `json:"name"`
	FeedURL      *string `json:"feed_url"`
	FriendLinkID *int    `json:"friend_link_id"`
	Status       *string `json:"status" binding:"omitempty,oneof=active paused"`
}

// PostQuery defines the query parameters for fetching posts.
type PostQuery struct {
	RssID        *int `form:"rss_id"`
	FriendLinkID *int `form:"friend_link_id"`
	Page         int  `form:"page"`
	PageSize     int  `form:"page_size"`
}

// UploadResourceReq 定义了上传资源请求的表单字段。
type UploadResourceReq struct {
	Path      string `form:"path"`
	Overwrite bool   `form:"overwrite"`
}

// CreateRssReq subscribes to one feed URL. FriendLinkID 0 means no owner; an
// empty name falls back to the feed title.
type CreateRssReq struct {
	FeedURL      string `json:"feed_url" binding:"required"`
	FriendLinkID int    `json:"friend_link_id"`
	Name         string `json:"name"`
}

// DiscoverRssReq asks for the feeds a web page advertises.
type DiscoverRssReq struct {
	URL string `json:"url" binding:"required"`
}

// ImageQueryOptions defines the query parameters for fetching images.
type ImageQueryOptions struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Name     string `form:"name"`
}

// UpdateImageReq defines the request body for updating an image.
type UpdateImageReq struct {
	Name      *string `json:"name"`
	URL       *string `json:"url"`
	LocalPath *string `json:"local_path"`
	IsLocal   *int    `json:"is_local"`
	IsOss     *int    `json:"is_oss"`
	Status    *string `json:"status" binding:"omitempty,oneof=normal pause broken pending"`
}

// UpdateConfigReq 只携带 system_conf 的 JSON Merge Patch，不包含启动环境配置。
type UpdateConfigReq struct {
	SystemConf json.RawMessage `json:"system_conf" binding:"required"`
}

// MediaRequest represents the media data in the create moment request.
type MediaRequest struct {
	MediaURL  string `json:"media_url" binding:"required"`
	MediaType string `json:"media_type" binding:"required"`
}

// CreateMomentRequest represents the request body for creating a new moment.
type CreateMomentRequest struct {
	Content     string         `json:"content" binding:"required"`
	Media       []MediaRequest `json:"media"`
	GuildID     *int64         `json:"guild_id" binding:"omitempty,gte=0"`
	ChannelID   *int64         `json:"channel_id" binding:"omitempty,gte=0"`
	MessageID   *int64         `json:"message_id" binding:"omitempty,gte=0"`
	MessageLink *string        `json:"message_link"`
}

// CreateMomentMediaReq defines the request body for creating moment media.
type CreateMomentMediaReq struct {
	MomentID  int    `json:"moment_id" binding:"gt=0"`
	Name      string `json:"name"`
	MediaURL  string `json:"media_url" binding:"required"`
	MediaType string `json:"media_type" binding:"omitempty,oneof=image video"`
	IsLocal   int    `json:"is_local" binding:"oneof=0 1"`
}

// UpdateMomentMediaReq defines the request body for updating moment media.
type UpdateMomentMediaReq struct {
	MomentID  *int    `json:"moment_id" binding:"omitempty,gt=0"`
	Name      *string `json:"name"`
	MediaURL  *string `json:"media_url"`
	MediaType *string `json:"media_type" binding:"omitempty,oneof=image video"`
	IsLocal   *int    `json:"is_local" binding:"omitempty,oneof=0 1"`
}

// UpdateMomentRequest defines the request body for updating a moment.
type UpdateMomentRequest struct {
	Content     *string `json:"content"`
	Status      *string `json:"status" binding:"omitempty,oneof=visible hidden deleted"`
	GuildID     *int64  `json:"guild_id" binding:"omitempty,gte=0"`
	ChannelID   *int64  `json:"channel_id" binding:"omitempty,gte=0"`
	MessageID   *int64  `json:"message_id" binding:"omitempty,gte=0"`
	MessageLink *string `json:"message_link"`
}

// MomentReactionRequest defines the request body for reacting to a moment.
type MomentReactionRequest struct {
	Reaction string `json:"reaction" binding:"required,oneof=👍 👎 ❤ 👀 💩"`
}
