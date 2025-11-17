package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"blog_api/src/model"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

// InitJWTSecret 初始化 JWT 密钥
func InitJWTSecret() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// 如果环境变量未设置，生成一个随机密钥
		randomBytes := make([]byte, 32)
		rand.Read(randomBytes)
		secret = hex.EncodeToString(randomBytes)
	}
	jwtSecret = []byte(secret)
}

// AuthService 认证服务
type AuthService struct{}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	if len(jwtSecret) == 0 {
		InitJWTSecret()
	}
	return &AuthService{}
}

// ValidateCredentials 验证用户名和密码
func (s *AuthService) ValidateCredentials(username, password string) bool {
	expectedUsername := os.Getenv("WEB_PANEL_USER")
	expectedPassword := os.Getenv("WEB_PANEL_PWD")

	if expectedUsername == "" {
		expectedUsername = "admin"
	}
	if expectedPassword == "" {
		expectedPassword = "password"
	}

	return username == expectedUsername && password == expectedPassword
}

// GenerateJWT 生成 JWT token
func (s *AuthService) GenerateJWT(username string) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour)

	claims := &model.JWTClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateJWT 验证 JWT token
func (s *AuthService) ValidateJWT(tokenString string) (*model.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*model.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
