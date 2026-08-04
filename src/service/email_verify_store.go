package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

const (
	defaultEmailCodeTTLSeconds  = 600
	defaultEmailTokenTTLSeconds = 86400
	maxEmailCodes               = 10000
	maxEmailTokens              = 10000
	// maxEmailCodeAttempts bounds brute-force tries against one code before it is invalidated.
	maxEmailCodeAttempts = 5
	// minEmailCodeResendIntervalSeconds throttles code re-issue per email so an
	// attacker cannot trade re-sends for fresh brute-force budgets.
	minEmailCodeResendIntervalSeconds = 60
)

// ErrEmailCodeResendTooSoon is returned when a new code is requested before the resend cooldown elapsed.
var ErrEmailCodeResendTooSoon = errors.New("email verification code requested too frequently")

type emailCodeEntry struct {
	code      string
	expiresAt int64
	issuedAt  int64
	attempts  int
}

type emailTokenEntry struct {
	email     string
	expiresAt int64
}

// EmailVerifyStore keeps short-lived email verification codes and tokens in memory.
type EmailVerifyStore struct {
	mu     sync.Mutex
	codes  map[string]emailCodeEntry
	tokens map[string]emailTokenEntry
}

var emailVerifyStore = &EmailVerifyStore{
	codes:  make(map[string]emailCodeEntry),
	tokens: make(map[string]emailTokenEntry),
}

// EmailCodeTTLSeconds returns the default TTL for email verification codes.
func EmailCodeTTLSeconds() int {
	return defaultEmailCodeTTLSeconds
}

// EmailTokenTTLSeconds returns the default TTL for email auth tokens.
func EmailTokenTTLSeconds() int {
	return defaultEmailTokenTTLSeconds
}

// IssueEmailVerifyCode creates a new verification code for the given email.
func IssueEmailVerifyCode(email string) (string, int64, error) {
	code, err := generateEmailCode()
	if err != nil {
		return "", 0, err
	}
	expiresAt := time.Now().Add(defaultEmailCodeTTLSeconds * time.Second).Unix()

	now := time.Now().Unix()
	emailVerifyStore.mu.Lock()
	defer emailVerifyStore.mu.Unlock()
	emailVerifyStore.cleanupLocked(now)
	if existing, exists := emailVerifyStore.codes[email]; exists {
		if now-existing.issuedAt < minEmailCodeResendIntervalSeconds {
			return "", 0, ErrEmailCodeResendTooSoon
		}
	} else if len(emailVerifyStore.codes) >= maxEmailCodes {
		return "", 0, fmt.Errorf("email verification code capacity reached")
	}
	emailVerifyStore.codes[email] = emailCodeEntry{
		code:      code,
		expiresAt: expiresAt,
		issuedAt:  now,
	}

	return code, expiresAt, nil
}

// ValidateEmailVerifyCode verifies and consumes a verification code for the email.
func ValidateEmailVerifyCode(email, code string) bool {
	now := time.Now().Unix()
	emailVerifyStore.mu.Lock()
	defer emailVerifyStore.mu.Unlock()

	entry, ok := emailVerifyStore.codes[email]
	if !ok || entry.expiresAt <= now {
		if ok {
			delete(emailVerifyStore.codes, email)
		}
		return false
	}
	if entry.code != code {
		entry.attempts++
		if entry.attempts >= maxEmailCodeAttempts {
			// Invalidate the code after too many wrong guesses; the owner must request a new one.
			delete(emailVerifyStore.codes, email)
			return false
		}
		emailVerifyStore.codes[email] = entry
		return false
	}
	delete(emailVerifyStore.codes, email)
	return true
}

// IssueEmailToken creates a new short-lived token bound to the email.
func IssueEmailToken(email string) (string, int64, error) {
	token, err := generateEmailToken()
	if err != nil {
		return "", 0, err
	}
	expiresAt := time.Now().Add(defaultEmailTokenTTLSeconds * time.Second).Unix()

	emailVerifyStore.mu.Lock()
	defer emailVerifyStore.mu.Unlock()
	emailVerifyStore.cleanupLocked(time.Now().Unix())
	if len(emailVerifyStore.tokens) >= maxEmailTokens {
		return "", 0, fmt.Errorf("email token capacity reached")
	}
	emailVerifyStore.tokens[token] = emailTokenEntry{
		email:     email,
		expiresAt: expiresAt,
	}

	return token, expiresAt, nil
}

// ValidateEmailToken validates a token without consuming it.
func ValidateEmailToken(token string) (string, bool) {
	now := time.Now().Unix()
	emailVerifyStore.mu.Lock()
	defer emailVerifyStore.mu.Unlock()

	entry, ok := emailVerifyStore.tokens[token]
	if !ok || entry.expiresAt <= now {
		if ok {
			delete(emailVerifyStore.tokens, token)
		}
		return "", false
	}
	return entry.email, true
}

func (s *EmailVerifyStore) cleanupLocked(now int64) {
	for email, entry := range s.codes {
		if entry.expiresAt <= now {
			delete(s.codes, email)
		}
	}
	for token, entry := range s.tokens {
		if entry.expiresAt <= now {
			delete(s.tokens, token)
		}
	}
}

func generateEmailCode() (string, error) {
	max := big.NewInt(1000000)
	num, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("generate email code: %w", err)
	}
	return fmt.Sprintf("%06d", num.Int64()), nil
}

func generateEmailToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate email token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
