package model

// FriendTag belongs exclusively to friend links.
type FriendTag struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	CreatedAt int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt int64  `json:"updated_at" gorm:"autoUpdateTime:false"`
}

func (FriendTag) TableName() string { return "friend_tag" }

// FriendLinkTag's composite key prevents duplicate assignments.
type FriendLinkTag struct {
	FriendLinkID int `gorm:"primaryKey;autoIncrement:false"`
	TagID        int `gorm:"primaryKey;autoIncrement:false"`
}

func (FriendLinkTag) TableName() string { return "friend_link_tag" }
