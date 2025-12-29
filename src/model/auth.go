package model

import "github.com/golang-jwt/jwt/v5"

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// JWTClaims JWT 载荷
type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Fingerprint represents a verified visitor identity.
type Fingerprint struct {
	ID               int    `json:"id" gorm:"column:id;primaryKey"`
	Fingerprint      string `json:"fingerprint" gorm:"column:fingerprint"`
	UserAgent        string `json:"user_agent,omitempty" gorm:"column:user_agent"`
	IP               string `json:"ip,omitempty" gorm:"column:ip"`
	PermissionsLevel string `json:"permissions_level" gorm:"column:permissions_level"`
	CreatedAt        int64  `json:"created_at" gorm:"column:created_at"`
}

// TableName sets the table name for Fingerprint.
func (Fingerprint) TableName() string {
	return "fingerprints"
}
