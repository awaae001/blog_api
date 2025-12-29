package handler

import (
	"blog_api/src/config"
	"blog_api/src/model"
	"blog_api/src/repositories"
	"blog_api/src/service"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FingerprintHandler handles fingerprint related requests.
type FingerprintHandler struct {
	DB *gorm.DB
}

// NewFingerprintHandler creates a new fingerprint handler.
func NewFingerprintHandler(db *gorm.DB) *FingerprintHandler {
	return &FingerprintHandler{DB: db}
}

// CreateFingerprint handles POST /api/verify/fingerprint request.
func (h *FingerprintHandler) CreateFingerprint(c *gin.Context) {
	cfg := config.GetConfig()
	secret := cfg.Verify.Fingerprint.Secret
	if secret == "" {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "fingerprint secret is not configured"))
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	fingerprintValue := hashFingerprint(ip, userAgent, secret)

	record, err := repositories.GetFingerprintByValue(h.DB, fingerprintValue)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to query fingerprint"))
			return
		}

		record = &model.Fingerprint{
			Fingerprint:      fingerprintValue,
			UserAgent:        userAgent,
			IP:               ip,
			PermissionsLevel: "normal",
			CreatedAt:        time.Now().Unix(),
		}
		if err := repositories.CreateFingerprint(h.DB, record); err != nil {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "failed to create fingerprint"))
			return
		}
	}

	tokenService := service.NewFingerprintTokenService(secret)
	token := tokenService.Sign(record.ID)

	c.JSON(http.StatusOK, model.NewSuccessResponse(map[string]string{
		"fingerprint_token": token,
	}))
}

func hashFingerprint(ip, userAgent, secret string) string {
	sum := sha256.Sum256([]byte(ip + "|" + userAgent + "|" + secret))
	return hex.EncodeToString(sum[:])
}
