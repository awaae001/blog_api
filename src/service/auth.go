package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"time"

	"blog_api/src/model"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte
var expectedUsername string
var expectedPassword string

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

// InitAuth initializes the shared signing key and admin credentials at startup.
func InitAuth(cfg *model.Config) {
	InitJWTSecret()
	expectedUsername = cfg.WebPanelUser
	expectedPassword = cfg.WebPanelPwd
	if expectedPassword == "" {
		expectedUsername = "admin_" + randomHex(3)
		expectedPassword = randomHex(12)
		log.Printf("[auth] 未检测到 WEB_PANEL_PWD，已为本次启动生成临时后台账号: username=%s password=%s", expectedUsername, expectedPassword)
		return
	}
	if expectedUsername == "" {
		expectedUsername = "admin"
	}
	log.Printf("[auth] 已读取到后台凭证（仅显示前3位）: username=%s*** password=%s***", prefix3(expectedUsername), prefix3(expectedPassword))
}

// ValidateCredentials 验证用户名和密码
func ValidateCredentials(username, password string) bool {
	return username == expectedUsername && password == expectedPassword
}

func randomHex(byteLen int) string {
	b := make([]byte, byteLen)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func prefix3(s string) string {
	runes := []rune(s)
	if len(runes) <= 3 {
		return s
	}
	return string(runes[:3])
}

// GenerateJWT 生成 JWT token
func GenerateJWT(username string) (string, time.Time, error) {
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
func ValidateJWT(tokenString string) (*model.JWTClaims, error) {
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
