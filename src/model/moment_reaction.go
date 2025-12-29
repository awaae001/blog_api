package model

// MomentReaction represents a reaction for a moment.
type MomentReaction struct {
	ID            int    `json:"id" gorm:"column:id;primaryKey"`
	MomentID      int    `json:"moment_id" gorm:"column:moment_id"`
	FingerprintID int    `json:"fingerprint_id" gorm:"column:fingerprint_id"`
	Reaction      string `json:"reaction" gorm:"column:reaction"`
	CreatedAt     int64  `json:"created_at" gorm:"column:created_at"`
}

// TableName sets the table name for MomentReaction.
func (MomentReaction) TableName() string {
	return "moment_reactions"
}
