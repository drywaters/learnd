package session

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		pool.Close()
		t.Fatalf("failed to ensure sessions table: %v", err)
	}

	if _, err := pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)`); err != nil {
		pool.Close()
		t.Fatalf("failed to ensure sessions index: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM sessions`)
		pool.Close()
	})

	return pool
}

func TestStoreLifecycle(t *testing.T) {
	pool := setupTestPool(t)
	store := NewStore(pool, 200*time.Millisecond)
	t.Cleanup(store.Close)

	token, err := store.Create()
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if !store.Valid(token) {
		t.Fatalf("expected session to be valid after create")
	}

	time.Sleep(100 * time.Millisecond)
	store.Refresh(token)

	time.Sleep(120 * time.Millisecond)
	if !store.Valid(token) {
		t.Fatalf("expected session to be valid after refresh")
	}

	store.Delete(token)
	if store.Valid(token) {
		t.Fatalf("expected session to be invalid after delete")
	}
}

func TestStoreExpires(t *testing.T) {
	pool := setupTestPool(t)
	store := NewStore(pool, 150*time.Millisecond)
	t.Cleanup(store.Close)

	token, err := store.Create()
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	if store.Valid(token) {
		t.Fatalf("expected session to expire")
	}
}
