package gamification

import (
	"context"
	"testing"
	"time"

	"github.com/goacademy/backend/internal/activity"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestComputeStreak(t *testing.T) {
	base := day(2026, 6, 20)

	cases := []struct {
		name         string
		prevActive   time.Time
		hasPrev      bool
		prevCurrent  int
		prevLongest  int
		newDate      time.Time
		wantCurrent  int
		wantLongest  int
		wantLastDate time.Time
	}{
		{"first ever activity", time.Time{}, false, 0, 0, base, 1, 1, base},
		{"first ever, longest preserved", time.Time{}, false, 0, 5, base, 1, 5, base},
		{"same day, no change", base, true, 3, 5, base, 3, 5, base},
		{"next day increments", base, true, 3, 5, base.AddDate(0, 0, 1), 4, 5, base.AddDate(0, 0, 1)},
		{"next day sets new longest", base, true, 5, 5, base.AddDate(0, 0, 1), 6, 6, base.AddDate(0, 0, 1)},
		{"two-day gap resets to 1", base, true, 4, 7, base.AddDate(0, 0, 2), 1, 7, base.AddDate(0, 0, 2)},
		{"backfilled older day ignored", base, true, 4, 7, base.AddDate(0, 0, -1), 4, 7, base},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeStreak(tc.prevActive, tc.hasPrev, tc.prevCurrent, tc.prevLongest, tc.newDate)
			if got.Current != tc.wantCurrent || got.Longest != tc.wantLongest || !got.LastActive.Equal(tc.wantLastDate) {
				t.Errorf("computeStreak = {current:%d longest:%d last:%s}, want {current:%d longest:%d last:%s}",
					got.Current, got.Longest, got.LastActive.Format("2006-01-02"),
					tc.wantCurrent, tc.wantLongest, tc.wantLastDate.Format("2006-01-02"))
			}
		})
	}
}

func TestRecorder_StreakIncrement_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	rec := NewRecorder(pool)
	svc := NewService(pool)
	uid := seedUser(t, pool)

	// First activity today: streak starts at 1.
	if err := rec.Record(ctx, activity.Event{UserID: uid, Type: "video_completed", RefType: "video", RefID: newUUIDString(t)}); err != nil {
		t.Fatalf("record day 1: %v", err)
	}
	st, _ := svc.GetStats(ctx, uid)
	if st.CurrentStreak != 1 || st.LongestStreak != 1 {
		t.Fatalf("after first activity: current=%d longest=%d, want 1/1", st.CurrentStreak, st.LongestStreak)
	}

	// Simulate that the recorded activity happened yesterday, then record again
	// today: a consecutive day must advance the streak to 2.
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	if _, err := pool.Exec(ctx, "UPDATE user_stats SET last_active_date = $2 WHERE user_id = $1::uuid", uid, yesterday); err != nil {
		t.Fatalf("backdate: %v", err)
	}
	if err := rec.Record(ctx, activity.Event{UserID: uid, Type: "article_read", RefType: "article", RefID: newUUIDString(t)}); err != nil {
		t.Fatalf("record day 2: %v", err)
	}
	st2, _ := svc.GetStats(ctx, uid)
	if st2.CurrentStreak != 2 || st2.LongestStreak != 2 {
		t.Errorf("after consecutive day: current=%d longest=%d, want 2/2", st2.CurrentStreak, st2.LongestStreak)
	}

	// A same-day repeat does not change the streak.
	if err := rec.Record(ctx, activity.Event{UserID: uid, Type: "problem_solved", RefType: "problem", RefID: newUUIDString(t)}); err != nil {
		t.Fatalf("record same day: %v", err)
	}
	st3, _ := svc.GetStats(ctx, uid)
	if st3.CurrentStreak != 2 {
		t.Errorf("same-day activity must not change streak, got %d", st3.CurrentStreak)
	}
}
