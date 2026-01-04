package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbTimeout         = 5 * time.Second
	tokenSize         = 32
	maxCreateAttempts = 5
)

// Store manages session tokens with automatic expiration.
type Store struct {
	pool *pgxpool.Pool
	ttl  time.Duration
	done chan struct{}
	wg   sync.WaitGroup
}

// NewStore creates a new session store with the given TTL.
func NewStore(pool *pgxpool.Pool, ttl time.Duration) *Store {
	s := &Store{
		pool: pool,
		ttl:  ttl,
		done: make(chan struct{}),
	}
	// Start background cleanup goroutine.
	s.wg.Add(1)
	go s.cleanup()
	return s
}

// Create generates a new session token and stores it.
func (s *Store) Create() (string, error) {
	for i := 0; i < maxCreateAttempts; i++ {
		token, err := generateToken()
		if err != nil {
			return "", err
		}

		ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
		cmdTag, err := s.pool.Exec(ctx, `
			INSERT INTO sessions (token, expires_at)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, token, time.Now().Add(s.ttl))
		cancel()

		if err != nil {
			return "", fmt.Errorf("failed to create session: %w", err)
		}
		if cmdTag.RowsAffected() == 1 {
			return token, nil
		}
	}

	return "", errors.New("failed to create unique session token")
}

// Valid checks if a session token is valid (exists and not expired).
func (s *Store) Valid(token string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var expiry time.Time
	err := s.pool.QueryRow(ctx, `SELECT expires_at FROM sessions WHERE token = $1`, token).Scan(&expiry)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false
		}
		slog.Error("session lookup failed", "error", err)
		return false
	}

	if time.Now().After(expiry) {
		s.Delete(token)
		return false
	}

	return true
}

// Delete removes a session token.
func (s *Store) Delete(token string) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if _, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token); err != nil {
		slog.Error("session delete failed", "error", err)
	}
}

// Refresh extends the expiration of a valid token.
func (s *Store) Refresh(token string) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if _, err := s.pool.Exec(ctx, `
		UPDATE sessions
		SET expires_at = $2
		WHERE token = $1 AND expires_at > NOW()
	`, token, time.Now().Add(s.ttl)); err != nil {
		slog.Error("session refresh failed", "error", err)
	}
}

// cleanup periodically removes expired sessions.
func (s *Store) cleanup() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
			_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= NOW()`)
			cancel()
			if err != nil {
				slog.Error("session cleanup failed", "error", err)
			}
		}
	}
}

// Close signals the cleanup goroutine to stop and waits for it to finish.
func (s *Store) Close() {
	close(s.done)
	s.wg.Wait()
}

func generateToken() (string, error) {
	b := make([]byte, tokenSize)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
