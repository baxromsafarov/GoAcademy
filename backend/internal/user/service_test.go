package user

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

func TestUpdateInput_Validate_CollectsErrors(t *testing.T) {
	tooLong := strings.Repeat("x", 101)
	badLocale := "xx"
	in := UpdateInput{DisplayName: &tooLong, Locale: &badLocale}

	err := in.validate()
	var apiErr *apierr.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 400 {
		t.Fatalf("expected 400 validation error, got %v", err)
	}
	details, _ := apiErr.Details.(map[string]string)
	for _, f := range []string{"display_name", "locale"} {
		if _, ok := details[f]; !ok {
			t.Errorf("expected a validation error for %q; got %v", f, details)
		}
	}
}

func TestUpdateInput_Validate_TrimsDisplayName(t *testing.T) {
	name := "  Alice  "
	in := UpdateInput{DisplayName: &name}
	if err := in.validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Alice" {
		t.Errorf("display name not trimmed in place: %q", name)
	}
}

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run user service integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func seedUser(t *testing.T, pool *pgxpool.Pool) (id, email string) {
	t.Helper()
	email = fmt.Sprintf("profile-%d@example.com", time.Now().UnixNano())
	u, err := store.New(pool).CreateUser(context.Background(), store.CreateUserParams{
		Email:        email,
		PasswordHash: "placeholder",
		DisplayName:  "Initial",
		Locale:       store.LocaleEn,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
	return pgxutil.UUIDString(u.ID), email
}

func TestService_GetAndPartialUpdate_Integration(t *testing.T) {
	pool := openPool(t)
	svc := NewService(pool, nil) // avatar storage not exercised here
	ctx := context.Background()
	id, email := seedUser(t, pool)

	got, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Email != email || got.DisplayName != "Initial" {
		t.Fatalf("unexpected user: email=%q name=%q", got.Email, got.DisplayName)
	}

	// Partial update: only bio + is_public; other fields must stay unchanged.
	bio, pub := "I love Go", false
	updated, err := svc.Update(ctx, id, UpdateInput{Bio: &bio, IsPublic: &pub})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Bio != "I love Go" || updated.IsPublic {
		t.Errorf("bio/is_public not applied: bio=%q is_public=%v", updated.Bio, updated.IsPublic)
	}
	if updated.DisplayName != "Initial" || updated.Locale != store.LocaleEn {
		t.Error("untouched fields should remain unchanged on partial update")
	}

	// Second update: display name + locale; bio must persist.
	name, loc := "Renamed", string(store.LocaleRu)
	updated2, err := svc.Update(ctx, id, UpdateInput{DisplayName: &name, Locale: &loc})
	if err != nil {
		t.Fatalf("Update 2: %v", err)
	}
	if updated2.DisplayName != "Renamed" || updated2.Locale != store.LocaleRu {
		t.Errorf("name/locale not applied: name=%q locale=%q", updated2.DisplayName, updated2.Locale)
	}
	if updated2.Bio != "I love Go" {
		t.Error("bio from the previous update should persist")
	}

	// Invalid locale is rejected.
	bad := "xx"
	if _, err := svc.Update(ctx, id, UpdateInput{Locale: &bad}); err == nil {
		t.Error("invalid locale should be rejected")
	}

	// Unknown / malformed ids.
	if _, err := svc.GetByID(ctx, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Error("unknown id should return not found")
	}
	if _, err := svc.GetByID(ctx, "not-a-uuid"); err == nil {
		t.Error("malformed id should return an error")
	}
}
