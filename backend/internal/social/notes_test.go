package social

import (
	"context"
	"testing"

	"github.com/goacademy/backend/internal/platform/pgxutil"
)

func newUUID(t *testing.T) string {
	t.Helper()
	u, err := pgxutil.NewUUID()
	if err != nil {
		t.Fatalf("uuid: %v", err)
	}
	return pgxutil.UUIDString(u)
}

func TestNotes_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewNotesService(pool)

	owner := seedUserStats(t, pool, "note-owner", true, false, 0)
	other := seedUserStats(t, pool, "note-other", true, false, 0)
	contentID := newUUID(t)

	// Owner creates a note.
	note, err := svc.Create(ctx, owner, CreateNoteInput{ContentType: "article", ContentID: contentID, Body: "first"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if note.Body != "first" || note.ContentType != "article" || note.ContentID != contentID {
		t.Fatalf("unexpected note: %+v", note)
	}

	// It appears in the owner's list, and not in the other user's.
	if ns, _ := svc.List(ctx, owner); len(ns) != 1 || ns[0].ID != note.ID {
		t.Fatalf("owner list = %+v, want the one note", ns)
	}
	if ns, _ := svc.List(ctx, other); len(ns) != 0 {
		t.Errorf("other user must not see the note, got %d", len(ns))
	}

	// Owner edits it.
	upd, err := svc.Update(ctx, owner, note.ID, "edited")
	if err != nil || upd.Body != "edited" {
		t.Fatalf("owner Update = (%+v, %v), want body 'edited'", upd, err)
	}

	// A non-owner can neither update nor delete it.
	if _, err := svc.Update(ctx, other, note.ID, "hacked"); err == nil {
		t.Error("non-owner Update should fail")
	}
	if err := svc.Delete(ctx, other, note.ID); err == nil {
		t.Error("non-owner Delete should fail")
	}
	// ...and the note is untouched.
	if ns, _ := svc.List(ctx, owner); len(ns) != 1 || ns[0].Body != "edited" {
		t.Errorf("note must be unchanged after non-owner attempts, got %+v", ns)
	}

	// Owner deletes it.
	if err := svc.Delete(ctx, owner, note.ID); err != nil {
		t.Fatalf("owner Delete: %v", err)
	}
	if ns, _ := svc.List(ctx, owner); len(ns) != 0 {
		t.Errorf("note should be gone, got %d", len(ns))
	}

	// Validation and not-found paths.
	if _, err := svc.Create(ctx, owner, CreateNoteInput{ContentType: "bogus", ContentID: contentID, Body: "x"}); err == nil {
		t.Error("invalid content_type should fail validation")
	}
	if _, err := svc.Create(ctx, owner, CreateNoteInput{ContentType: "article", ContentID: "not-a-uuid", Body: "x"}); err == nil {
		t.Error("invalid content_id should fail validation")
	}
	if _, err := svc.Create(ctx, owner, CreateNoteInput{ContentType: "article", ContentID: contentID, Body: "  "}); err == nil {
		t.Error("blank body should fail validation")
	}
	if _, err := svc.Update(ctx, owner, newUUID(t), "x"); err == nil {
		t.Error("updating a missing note should be not-found")
	}
	if err := svc.Delete(ctx, owner, newUUID(t)); err == nil {
		t.Error("deleting a missing note should be not-found")
	}
}
