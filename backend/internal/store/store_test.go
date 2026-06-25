package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/store"
)

// TestNow_Integration exercises the generated sqlc Now query against a real
// database. It is skipped unless TEST_DATABASE_URL is set, keeping the default
// `go test ./...` hermetic.
func TestNow_Integration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run the store integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	q := store.New(pool)
	got, err := q.Now(ctx)
	if err != nil {
		t.Fatalf("Now: %v", err)
	}
	if !got.Valid || got.Time.IsZero() {
		t.Fatalf("Now returned invalid timestamp: %+v", got)
	}
}
