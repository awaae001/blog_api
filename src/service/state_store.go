package service

import (
	"blog_api/src/model"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"
)

var (
	ErrInvalidStatePayload     = errors.New("state payload must be a JSON object")
	ErrStatePayloadTooLarge    = errors.New("state payload is too large")
	ErrStateCapacityReached    = errors.New("state capacity reached")
	ErrStateMemoryLimitReached = errors.New("state memory limit reached")
)

type stateEntry struct {
	payload   json.RawMessage
	expiresAt time.Time
}

// StoredState is an immutable snapshot returned by StateStore.
type StoredState struct {
	Payload   json.RawMessage
	ExpiresAt time.Time
}

type StateStore struct {
	mu              sync.Mutex
	entries         map[string]stateEntry
	maxEntries      int
	maxPayloadBytes int64
	maxTotalBytes   int64
	totalBytes      int64
	now             func() time.Time
}

func NewStateStore(maxEntries int, maxPayloadBytes, maxTotalBytes int64) *StateStore {
	return &StateStore{
		entries:         make(map[string]stateEntry),
		maxEntries:      maxEntries,
		maxPayloadBytes: maxPayloadBytes,
		maxTotalBytes:   maxTotalBytes,
		now:             time.Now,
	}
}

func (s *StateStore) RunCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.mu.Lock()
			s.cleanupLocked(now)
			s.mu.Unlock()
		}
	}
}

func (s *StateStore) Put(key string, payload []byte, expiresAt time.Time) (state StoredState, created bool, err error) {
	if int64(len(payload)) > s.maxPayloadBytes {
		return StoredState{}, false, ErrStatePayloadTooLarge
	}
	if !isJSONObject(payload) {
		return StoredState{}, false, ErrInvalidStatePayload
	}

	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(now)

	existing, exists := s.entries[key]
	if !exists && len(s.entries) >= s.maxEntries {
		return StoredState{}, false, ErrStateCapacityReached
	}
	existingBytes := int64(len(existing.payload))
	newTotalBytes := s.totalBytes - existingBytes + int64(len(payload))
	if newTotalBytes > s.maxTotalBytes {
		return StoredState{}, false, ErrStateMemoryLimitReached
	}

	ownedPayload := bytes.Clone(payload)
	s.entries[key] = stateEntry{payload: ownedPayload, expiresAt: expiresAt}
	s.totalBytes = newTotalBytes
	return snapshot(s.entries[key]), !exists, nil
}

// Get returns a copied snapshot for key and lazily removes expired states.
func (s *StateStore) Get(key string) (StoredState, bool) {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(now)

	entry, ok := s.entries[key]
	if !ok {
		return StoredState{}, false
	}
	return snapshot(entry), true
}

// Delete removes key and reports whether it existed and had not expired.
func (s *StateStore) Delete(key string) bool {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(now)

	entry, ok := s.entries[key]
	if !ok {
		return false
	}
	delete(s.entries, key)
	s.totalBytes -= int64(len(entry.payload))
	return true
}

// Stats returns a consistent usage snapshot after removing expired entries.
func (s *StateStore) Stats() model.KVCacheStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(s.now())

	return model.KVCacheStatus{
		EntryCount:   len(s.entries),
		PayloadBytes: s.totalBytes,
	}
}

func (s *StateStore) cleanupLocked(now time.Time) {
	for key, entry := range s.entries {
		if !entry.expiresAt.After(now) {
			delete(s.entries, key)
			s.totalBytes -= int64(len(entry.payload))
		}
	}
}

func snapshot(entry stateEntry) StoredState {
	return StoredState{
		Payload:   bytes.Clone(entry.payload),
		ExpiresAt: entry.expiresAt,
	}
}

func isJSONObject(payload []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	var object map[string]json.RawMessage
	if err := decoder.Decode(&object); err != nil || object == nil {
		return false
	}
	var trailing interface{}
	return errors.Is(decoder.Decode(&trailing), io.EOF)
}
