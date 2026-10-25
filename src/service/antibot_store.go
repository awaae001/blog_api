package service

import (
	"errors"
	"sync"
	"time"
)

const (
	AntiBotTTLSeconds = 300
	maxAntiBotTokens  = 10000
)

var errAntiBotCapacityReached = errors.New("anti-bot token capacity reached")

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
	token := randomHex(32)
	expiresAt := time.Now().Add(AntiBotTTLSeconds * time.Second).Unix()

	antiBotStore.mu.Lock()
	defer antiBotStore.mu.Unlock()
	antiBotStore.cleanupLocked(time.Now().Unix())
	if len(antiBotStore.tokens) >= maxAntiBotTokens {
		return "", 0, errAntiBotCapacityReached
	}
	antiBotStore.tokens[token] = expiresAt

	return token, expiresAt, nil
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
