package model

// Moment represents a moment entry in the database.
type Moment struct {
	ID        int    `json:"id" gorm:"column:id;primaryKey"`
	Content   string `json:"content" gorm:"column:content"`
	Status    string `json:"status" gorm:"column:status"`
	GuildID   int64  `json:"guild_id,omitempty" gorm:"column:guild_id"`
	ChannelID int64  `json:"channel_id,omitempty" gorm:"column:channel_id"`
	MessageID int64  `json:"message_id,omitempty" gorm:"column:message_id"`
	CreatedAt int64  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt int64  `json:"updated_at" gorm:"column:updated_at"`
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
	IsDeleted int    `json:"is_deleted" gorm:"column:is_deleted"`
}

// TableName sets the table name for MomentMedia.
func (MomentMedia) TableName() string {
	return "moments_media"
}

// MomentWithMedia represents a moment with its associated media files.
type MomentWithMedia struct {
	Moment
	Media []MomentMedia `json:"media"`
}

// QueryMomentsResponse defines the response for querying moments.
type QueryMomentsResponse struct {
	Moments []MomentWithMedia `json:"moments"`
	Total   int64             `json:"total"`
}

// MediaRequest represents the media data in the create moment request.
type MediaRequest struct {
	MediaURL  string `json:"media_url" binding:"required"`
	MediaType string `json:"media_type" binding:"required"`
}

// CreateMomentRequest represents the request body for creating a new moment.
type CreateMomentRequest struct {
	Content string         `json:"content" binding:"required"`
	Media   []MediaRequest `json:"media"`
}
