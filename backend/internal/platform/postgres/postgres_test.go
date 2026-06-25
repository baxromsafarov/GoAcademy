package postgres

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestConnect_BadDSN ensures a malformed DSN fails fast at parse time (no retry
// loop, so the test stays quick).
func TestConnect_BadDSN(t *testing.T) {
	_, err := Connect(context.Background(), "host=localhost port=notanumber user=x")
	if err == nil {
		t.Fatal("expected error for malformed DSN, got nil")
	}
}

// TestConnect_Integration verifies a real connection and ping. It is skipped
// unless TEST_DATABASE_URL is set, keeping the default `go test ./...` hermetic.
func TestConnect_Integration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run the postgres integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}
