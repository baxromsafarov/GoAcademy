package social

import (
	"context"
	"testing"
)

func TestBookmarks_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewBookmarksService(pool)

	owner := seedUserStats(t, pool, "bm-owner", true, false, 0)
	other := seedUserStats(t, pool, "bm-other", true, false, 0)
	cid := newUUID(t)

	// Add a bookmark; it shows up in the owner's list.
	b, err := svc.Add(ctx, owner, "video", cid)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if bs, _ := svc.List(ctx, owner); len(bs) != 1 || bs[0].ID != b.ID {
		t.Fatalf("owner list = %+v, want the one bookmark", bs)
	}

	// Adding the SAME content again is a no-op: same id, no duplicate.
	b2, err := svc.Add(ctx, owner, "video", cid)
	if err != nil {
		t.Fatalf("Add duplicate: %v", err)
	}
	if b2.ID != b.ID {
		t.Errorf("duplicate add should return the same bookmark, got %s vs %s", b2.ID, b.ID)
	}
	if bs, _ := svc.List(ctx, owner); len(bs) != 1 {
		t.Errorf("no duplicate expected, got %d", len(bs))
	}

	// Different content adds a second bookmark.
	if _, err := svc.Add(ctx, owner, "article", newUUID(t)); err != nil {
		t.Fatalf("Add second: %v", err)
	}
	if bs, _ := svc.List(ctx, owner); len(bs) != 2 {
		t.Errorf("want 2 bookmarks, got %d", len(bs))
	}

	// Another user sees none of them.
	if bs, _ := svc.List(ctx, other); len(bs) != 0 {
		t.Errorf("other user must not see bookmarks, got %d", len(bs))
	}

	// A non-owner cannot remove, and the bookmark survives.
	if err := svc.Remove(ctx, other, b.ID); err == nil {
		t.Error("non-owner Remove should fail")
	}
	if bs, _ := svc.List(ctx, owner); len(bs) != 2 {
		t.Error("bookmark must survive a non-owner remove attempt")
	}

	// Owner removes one.
	if err := svc.Remove(ctx, owner, b.ID); err != nil {
		t.Fatalf("owner Remove: %v", err)
	}
	if bs, _ := svc.List(ctx, owner); len(bs) != 1 {
		t.Errorf("want 1 bookmark after remove, got %d", len(bs))
	}

	// Validation and not-found.
	if _, err := svc.Add(ctx, owner, "bogus", cid); err == nil {
		t.Error("invalid content_type should fail validation")
	}
	if _, err := svc.Add(ctx, owner, "video", "not-a-uuid"); err == nil {
		t.Error("invalid content_id should fail validation")
	}
	if err := svc.Remove(ctx, owner, newUUID(t)); err == nil {
		t.Error("removing a missing bookmark should be not-found")
	}
}
