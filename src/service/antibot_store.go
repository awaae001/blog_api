package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const defaultAntiBotTTLSeconds = 300

// AntiBotStore keeps short-lived anti-bot tokens in memory.
type AntiBotStore struct {
	mu     sync.Mutex
	tokens map[string]int64
}

var antiBotStore = &AntiBotStore{
	tokens: make(map[string]int64),
}

// IssueAntiBotToken creates a new anti-bot token.
func IssueAntiBotToken() (string, int64, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", 0, err
	}
	token := hex.EncodeToString(bytes)
	expiresAt := time.Now().Add(defaultAntiBotTTLSeconds * time.Second).Unix()

	antiBotStore.mu.Lock()
	antiBotStore.tokens[token] = expiresAt
	antiBotStore.cleanupLocked(time.Now().Unix())
	antiBotStore.mu.Unlock()

	return token, expiresAt, nil
}

// AntiBotTTLSeconds returns the default TTL for anti-bot tokens.
func AntiBotTTLSeconds() int {
	return defaultAntiBotTTLSeconds
}

// ValidateAntiBotToken validates a token and checks expiration.
func ValidateAntiBotToken(token string) bool {
	now := time.Now().Unix()
	antiBotStore.mu.Lock()
	defer antiBotStore.mu.Unlock()

	expiresAt, ok := antiBotStore.tokens[token]
	if !ok || expiresAt <= now {
		if ok {
			delete(antiBotStore.tokens, token)
		}
		return false
	}

	return true
}

func (s *AntiBotStore) cleanupLocked(now int64) {
	for token, expiresAt := range s.tokens {
		if expiresAt <= now {
			delete(s.tokens, token)
		}
	}
}
