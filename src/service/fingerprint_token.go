package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// FingerprintTokenService signs and verifies fingerprint tokens.
type FingerprintTokenService struct {
	secret []byte
}

// NewFingerprintTokenService creates a new token service.
func NewFingerprintTokenService(secret string) *FingerprintTokenService {
	return &FingerprintTokenService{secret: []byte(secret)}
}

// Sign generates a signed token for a fingerprint id.
func (s *FingerprintTokenService) Sign(id int) string {
	payload := strconv.Itoa(id)
	return payload + "." + s.sign(payload)
}

// Verify validates a signed token and returns the fingerprint id.
func (s *FingerprintTokenService) Verify(token string) (int, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, false
	}
	if !hmac.Equal([]byte(parts[1]), []byte(s.sign(parts[0]))) {
		return 0, false
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false
	}
	return id, true
}

func (s *FingerprintTokenService) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
