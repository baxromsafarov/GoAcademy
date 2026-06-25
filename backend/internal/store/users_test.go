package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/store"
)

// openTestPool returns a pool against TEST_DATABASE_URL, skipping the test when it
// is unset so the default `go test ./...` stays hermetic.
func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run store integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestUsers_CreateAndFetch_Integration(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	// Run inside a transaction we roll back, so no rows persist in the dev DB.
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback(ctx)

	q := store.New(tx)
	created, err := q.CreateUser(ctx, store.CreateUserParams{
		Email:        "Case@Example.com",
		PasswordHash: "argon2id-placeholder",
		DisplayName:  "Tester",
		Locale:       store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// Server-side defaults.
	if created.Role != store.UserRoleStudent {
		t.Errorf("role = %q, want %q", created.Role, store.UserRoleStudent)
	}
	if created.IsBlocked {
		t.Error("is_blocked default should be false")
	}
	if created.EmailVerified {
		t.Error("email_verified default should be false")
	}
	if !created.IsPublic {
		t.Error("is_public default should be true")
	}
	if created.Locale != store.LocaleEn {
		t.Errorf("locale = %q, want en", created.Locale)
	}
	if !created.ID.Valid {
		t.Error("id should be generated")
	}
	if !created.CreatedAt.Valid || !created.UpdatedAt.Valid {
		t.Error("created_at/updated_at should be set")
	}

	// citext: a lookup with different casing must return the same user.
	got, err := q.GetUserByEmail(ctx, "case@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail (different case): %v", err)
	}
	if got.ID != created.ID {
		t.Error("case-insensitive email lookup returned a different user")
	}

	byID, err := q.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if byID.Email != created.Email {
		t.Errorf("email mismatch: %q vs %q", byID.Email, created.Email)
	}
}

func TestUsers_UniqueEmailCaseInsensitive_Integration(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback(ctx)

	q := store.New(tx)
	base := store.CreateUserParams{
		Email:        "Dup@Example.com",
		PasswordHash: "x",
		DisplayName:  "A",
		Locale:       store.LocaleEn,
	}
	if _, err := q.CreateUser(ctx, base); err != nil {
		t.Fatalf("first insert: %v", err)
	}

	dup := base
	dup.Email = "dup@example.com" // same address, different case
	if _, err := q.CreateUser(ctx, dup); err == nil {
		t.Fatal("expected a unique-violation error for a case-insensitive duplicate email")
	}
}
