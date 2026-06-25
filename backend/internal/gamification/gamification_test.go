package gamification

import (
	"context"
	"testing"

	"github.com/goacademy/backend/internal/activity"
)

func TestXPFor(t *testing.T) {
	cases := map[string]int{
		"video_completed":   10,
		"article_read":      5,
		"quiz_passed":       20,
		"quiz_attempt":      2,
		"problem_solved":    30,
		"project_completed": 50,
		"unknown_thing":     0,
		"":                  0,
	}
	for typ, want := range cases {
		if got := XPFor(typ); got != want {
			t.Errorf("XPFor(%q) = %d, want %d", typ, got, want)
		}
	}
}

func TestLevelForXP(t *testing.T) {
	cases := []struct {
		xp   int
		want int
	}{
		{-50, 1}, {0, 1}, {1, 1}, {99, 1},
		{100, 2}, {399, 2},
		{400, 3}, {899, 3},
		{900, 4},
		{10000, 11}, // sqrt(100)=10 -> level 11
	}
	for _, tc := range cases {
		if got := LevelForXP(tc.xp); got != tc.want {
			t.Errorf("LevelForXP(%d) = %d, want %d", tc.xp, got, tc.want)
		}
	}
}

// TestRecorder_InvalidUser is hermetic: an invalid user id is rejected before any
// database access, so a nil pool is never dereferenced.
func TestRecorder_InvalidUser(t *testing.T) {
	rec := NewRecorder(nil)
	if err := rec.Record(context.Background(), activity.Event{UserID: "not-a-uuid", Type: "video_completed"}); err == nil {
		t.Error("invalid user id should error")
	}
}
