package admin

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

func strptr(s string) *string { return &s }

func TestAdminUsers_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)
	q := store.New(pool)

	seed := func(tag, role string) (id, email string) {
		t.Helper()
		email = fmt.Sprintf("%s-%d@example.com", tag, time.Now().UnixNano())
		u, err := q.CreateUser(ctx, store.CreateUserParams{Email: email, PasswordHash: "x", DisplayName: tag, Locale: store.LocaleEn})
		if err != nil {
			t.Fatalf("seed %s: %v", tag, err)
		}
		t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email) })
		if role == "admin" {
			if _, err := pool.Exec(ctx, "UPDATE users SET role='admin' WHERE id = $1", u.ID); err != nil {
				t.Fatalf("promote: %v", err)
			}
		}
		return pgxutil.UUIDString(u.ID), email
	}

	adminID, _ := seed("uadmin", "admin")
	studentID, _ := seed("ustudent", "student")
	_, victimEmail := seed("uvictim", "student")

	// Search by exact (unique) email returns just that user.
	list, err := svc.ListUsers(ctx, victimEmail, 50, 0)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if list.Total != 1 || len(list.Items) != 1 || list.Items[0].Email != victimEmail {
		t.Fatalf("search = total %d items %d, want 1 matching %s", list.Total, len(list.Items), victimEmail)
	}

	// Promote a student to admin.
	u, err := svc.UpdateUser(ctx, adminID, studentID, UpdateUserInput{Role: strptr("admin")})
	if err != nil || u.Role != store.UserRoleAdmin {
		t.Fatalf("promote = (%s, %v), want admin", u.Role, err)
	}

	// Block a user.
	blocked := true
	u2, err := svc.UpdateUser(ctx, adminID, studentID, UpdateUserInput{IsBlocked: &blocked})
	if err != nil || !u2.IsBlocked {
		t.Fatalf("block = (%v, %v), want blocked", u2.IsBlocked, err)
	}

	// Self-protection: an admin cannot block or demote themselves.
	if _, err := svc.UpdateUser(ctx, adminID, adminID, UpdateUserInput{IsBlocked: &blocked}); err == nil {
		t.Error("blocking yourself should be forbidden")
	}
	if _, err := svc.UpdateUser(ctx, adminID, adminID, UpdateUserInput{Role: strptr("student")}); err == nil {
		t.Error("demoting yourself should be forbidden")
	}
	// A self no-op (staying admin) is allowed.
	if _, err := svc.UpdateUser(ctx, adminID, adminID, UpdateUserInput{Role: strptr("admin")}); err != nil {
		t.Errorf("self no-op should be allowed: %v", err)
	}

	// Invalid role and unknown user.
	if _, err := svc.UpdateUser(ctx, adminID, studentID, UpdateUserInput{Role: strptr("superuser")}); err == nil {
		t.Error("invalid role should fail validation")
	}
	if _, err := svc.UpdateUser(ctx, adminID, randUUID(t), UpdateUserInput{Role: strptr("admin")}); err == nil {
		t.Error("unknown user should be not-found")
	}
}
