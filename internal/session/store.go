package session

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// Store manages session tokens with automatic expiration
type Store struct {
	mu       sync.RWMutex
	sessions map[string]time.Time // token -> expiration time
	ttl      time.Duration
	done     chan struct{}
	wg       sync.WaitGroup
}

// NewStore creates a new session store with the given TTL
func NewStore(ttl time.Duration) *Store {
	s := &Store{
		sessions: make(map[string]time.Time),
		ttl:      ttl,
		done:     make(chan struct{}),
	}
	// Start background cleanup goroutine
	s.wg.Add(1)
	go s.cleanup()
	return s
}

// Create generates a new session token and stores it
func (s *Store) Create() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)

	s.mu.Lock()
	s.sessions[token] = time.Now().Add(s.ttl)
	s.mu.Unlock()

	return token, nil
}

// Valid checks if a session token is valid (exists and not expired)
func (s *Store) Valid(token string) bool {
	s.mu.RLock()
	expiry, exists := s.sessions[token]
	s.mu.RUnlock()

	if !exists {
		return false
	}
	return time.Now().Before(expiry)
}

// Delete removes a session token
func (s *Store) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

// Refresh extends the expiration of a valid token
func (s *Store) Refresh(token string) {
	s.mu.Lock()
	if expiry, exists := s.sessions[token]; exists && time.Now().Before(expiry) {
		s.sessions[token] = time.Now().Add(s.ttl)
	}
	s.mu.Unlock()
}

// cleanup periodically removes expired sessions
func (s *Store) cleanup() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			now := time.Now()
			s.mu.Lock()
			for token, expiry := range s.sessions {
				if now.After(expiry) {
					delete(s.sessions, token)
				}
			}
			s.mu.Unlock()
		}
	}
}

// Close signals the cleanup goroutine to stop and waits for it to finish
func (s *Store) Close() {
	close(s.done)
	s.wg.Wait()
}
