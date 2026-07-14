package oauth2

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type StateEntry struct {
	Provider      string
	UserID        *uint
	PKCEChallenge string
	ExpiresAt     time.Time
}

type StateStore struct {
	mu    sync.RWMutex
	store map[string]*StateEntry
}

func NewStateStore() *StateStore {
	s := &StateStore{
		store: make(map[string]*StateEntry),
	}
	go s.cleanupLoop()
	return s
}

func (s *StateStore) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.store {
			if now.After(entry.ExpiresAt) {
				delete(s.store, key)
			}
		}
		s.mu.Unlock()
	}
}

func (s *StateStore) Generate(provider string, userID *uint, codeVerifier string) string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	state := hex.EncodeToString(b)
	s.mu.Lock()
	s.store[state] = &StateEntry{
		Provider:      provider,
		UserID:        userID,
		PKCEChallenge: codeVerifier,
		ExpiresAt:     time.Now().Add(10 * time.Minute),
	}
	s.mu.Unlock()
	return state
}

func (s *StateStore) Consume(state string) *StateEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.store[state]
	if !ok {
		return nil
	}
	delete(s.store, state)
	if time.Now().After(entry.ExpiresAt) {
		return nil
	}
	return entry
}
