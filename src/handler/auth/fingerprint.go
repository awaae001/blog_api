package authHandler

import (
	"blog_api/src/config"
	"blog_api/src/middleware"
	"blog_api/src/model"
	"blog_api/src/repositories"
	"blog_api/src/service"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FingerprintHandler handles fingerprint related requests.
type FingerprintHandler struct {
	DB *gorm.DB
}

// CreateFingerprint handles POST /api/verify/fingerprint request.
func (h *FingerprintHandler) CreateFingerprint(c *gin.Context) {
	cfg := config.GetConfig()
	secret := cfg.Verify.Fingerprint.Secret
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	tokenService := service.NewFingerprintTokenService(secret)

	// 请求级 context：客户端断开后查询会被取消
	db := h.DB.WithContext(c.Request.Context())

	if token, ok, err := h.reuseFingerprintToken(c, db, tokenService, ip, userAgent, secret); err != nil {
		if !middleware.AbortIfCanceled(c, err) {
			c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to update fingerprint"})
		}
		return
	} else if ok {
		c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: map[string]string{
			"fingerprint_token": token,
		}})
		return
	}

	fingerprintValue := hashFingerprint(ip, userAgent, secret)

	record, err := repositories.GetFingerprintByValue(db, fingerprintValue)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		if !middleware.AbortIfCanceled(c, err) {
			c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to query fingerprint"})
		}
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		record = &model.Fingerprint{
			Fingerprint:      fingerprintValue,
			UserAgent:        userAgent,
			IP:               ip,
			PermissionsLevel: "normal",
			CreatedAt:        time.Now().Unix(),
		}
		if err := db.Create(record).Error; err != nil {
			if !middleware.AbortIfCanceled(c, err) {
				c.JSON(http.StatusInternalServerError, model.ApiResponse{Code: 500, Message: "failed to create fingerprint"})
			}
			return
		}
	}

	token := tokenService.Sign(record.ID)

	c.JSON(http.StatusOK, model.ApiResponse{Code: http.StatusOK, Message: "success", Data: map[string]string{
		"fingerprint_token": token,
	}})
}

func (h *FingerprintHandler) reuseFingerprintToken(
	c *gin.Context,
	db *gorm.DB,
	tokenService *service.FingerprintTokenService,
	ip, userAgent, secret string,
) (string, bool, error) {
	token := middleware.ExtractFingerprintToken(c)
	if token == "" {
		return "", false, nil
	}

	id, ok := tokenService.Verify(token)
	if !ok {
		return "", false, nil
	}

	record, err := repositories.GetFingerprintByID(db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", false, nil
		}
		return "", false, err
	}

	fingerprintValue := hashFingerprint(ip, userAgent, secret)
	if record.Fingerprint == fingerprintValue && record.UserAgent == userAgent && record.IP == ip {
		return token, true, nil
	}

	if err := repositories.UpdateFingerprintIdentity(db, record.ID, fingerprintValue, userAgent, ip); err != nil {
		return "", false, err
	}

	return token, true, nil
}

func hashFingerprint(ip, userAgent, secret string) string {
	sum := sha256.Sum256([]byte(ip + "|" + userAgent + "|" + secret))
	return hex.EncodeToString(sum[:])
}
