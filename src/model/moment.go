package model

// Moment represents a moment entry in the database.
type Moment struct {
	ID          int    `json:"id" gorm:"column:id;primaryKey"`
	Content     string `json:"content" gorm:"column:content"`
	Status      string `json:"status" gorm:"column:status"`
	GuildID     int64  `json:"guild_id,omitempty" gorm:"column:guild_id"`
	ChannelID   int64  `json:"channel_id,omitempty" gorm:"column:channel_id"`
	MessageID   int64  `json:"message_id,omitempty" gorm:"column:message_id"`
	MessageLink string `json:"message_link,omitempty" gorm:"column:message_link"`
	CreatedAt   int64  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   int64  `json:"updated_at" gorm:"column:updated_at"`
}

// TableName sets the table name for Moment.
func (Moment) TableName() string {
	return "moments"
}

// MomentMedia represents a media file associated with a moment.
type MomentMedia struct {
	ID        int    `json:"id" gorm:"column:id;primaryKey"`
	MomentID  int    `json:"moment_id" gorm:"column:moment_id"`
	Name      string `json:"name,omitempty" gorm:"column:name"`
	MediaURL  string `json:"media_url" gorm:"column:media_url"`
	MediaType string `json:"media_type" gorm:"column:media_type"`
	IsLocal   int    `json:"is_local" gorm:"column:is_local"`
	IsDeleted int    `json:"is_deleted" gorm:"column:is_deleted"`
}

// TableName sets the table name for MomentMedia.
func (MomentMedia) TableName() string {
	return "moments_media"
}
