package gamification

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/store"
)

// TestCriterionMet is hermetic: the xp/streak/unknown branches never touch the
// database, so a nil Queries is safe.
func TestCriterionMet(t *testing.T) {
	mk := func(typ, params string) store.Badge {
		return store.Badge{CriteriaType: typ, CriteriaParams: []byte(params)}
	}
	var uid pgtype.UUID
	ctx := context.Background()

	cases := []struct {
		name    string
		badge   store.Badge
		xp      int
		streak  int
		wantMet bool
	}{
		{"xp meets threshold", mk("xp_at_least", `{"xp":100}`), 100, 0, true},
		{"xp below threshold", mk("xp_at_least", `{"xp":100}`), 99, 0, false},
		{"streak meets threshold", mk("streak_at_least", `{"days":7}`), 0, 7, true},
		{"streak below threshold", mk("streak_at_least", `{"days":7}`), 0, 6, false},
		{"unknown criterion never met", mk("future_thing", `{}`), 99999, 99999, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := criterionMet(ctx, nil, uid, tc.badge, tc.xp, tc.streak)
			if err != nil {
				t.Fatalf("criterionMet: %v", err)
			}
			if got != tc.wantMet {
				t.Errorf("criterionMet = %v, want %v", got, tc.wantMet)
			}
		})
	}
}

func TestRecorder_AwardsBadges_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	rec := NewRecorder(pool)
	svc := NewService(pool)
	uid := seedUser(t, pool)

	codes := func() map[string]int {
		bs, err := svc.GetBadges(ctx, uid)
		if err != nil {
			t.Fatalf("GetBadges: %v", err)
		}
		m := map[string]int{}
		for _, b := range bs {
			m[b.Code]++
		}
		return m
	}

	if len(codes()) != 0 {
		t.Fatal("fresh user should have no badges")
	}

	// Solve one problem (30 XP): earns first_problem, not yet xp_100.
	if err := rec.Record(ctx, activity.Event{UserID: uid, Type: "problem_solved", RefType: "problem", RefID: newUUIDString(t)}); err != nil {
		t.Fatalf("record: %v", err)
	}
	got := codes()
	if got["first_problem"] != 1 {
		t.Errorf("first_problem should be awarded once, got %d", got["first_problem"])
	}
	if got["xp_100"] != 0 {
		t.Error("xp_100 must not be awarded at 30 XP")
	}

	// Solve three more distinct problems -> 120 XP total, crossing the 100 XP badge.
	for i := 0; i < 3; i++ {
		if err := rec.Record(ctx, activity.Event{UserID: uid, Type: "problem_solved", RefType: "problem", RefID: newUUIDString(t)}); err != nil {
			t.Fatalf("record solve: %v", err)
		}
	}
	got = codes()
	if got["first_problem"] != 1 {
		t.Errorf("first_problem must not be duplicated, got %d", got["first_problem"])
	}
	if got["xp_100"] != 1 {
		t.Errorf("xp_100 should be awarded after 120 XP, got %d", got["xp_100"])
	}
}
