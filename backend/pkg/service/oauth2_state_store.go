package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type OAuthStateEntry struct {
	Provider      string
	UserID        *uint
	PKCEChallenge string
	ExpiresAt     time.Time
}

type OAuthStateStore struct {
	mu    sync.RWMutex
	store map[string]*OAuthStateEntry
}

func NewOAuthStateStore() *OAuthStateStore {
	s := &OAuthStateStore{
		store: make(map[string]*OAuthStateEntry),
	}
	go s.cleanupLoop()
	return s
}

func (s *OAuthStateStore) cleanupLoop() {
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

func (s *OAuthStateStore) Generate(provider string, userID *uint, codeVerifier string) string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	state := hex.EncodeToString(b)
	s.mu.Lock()
	s.store[state] = &OAuthStateEntry{
		Provider:      provider,
		UserID:        userID,
		PKCEChallenge: codeVerifier,
		ExpiresAt:     time.Now().Add(10 * time.Minute),
	}
	s.mu.Unlock()
	return state
}

func (s *OAuthStateStore) Consume(state string) *OAuthStateEntry {
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
