package gamification

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func seedDaily(t *testing.T, pool *pgxpool.Pool, date, contentType, contentID string, bonus int) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO daily_challenges (challenge_date, content_type, content_id, bonus_xp) VALUES ($1, $2::track_content_type, $3::uuid, $4)",
		date, contentType, contentID, bonus)
	if err != nil {
		t.Fatalf("seed daily challenge: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM daily_challenges WHERE challenge_date = $1", date)
	})
}

func TestDailyChallenge_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	rec := NewRecorder(pool)
	daily := NewDailyService(pool, rec)
	stats := NewService(pool)
	uid := seedUser(t, pool)

	today := time.Now().UTC()
	todayStr := today.Format("2006-01-02")

	// No challenge configured for today -> 404 on both endpoints.
	if _, err := daily.Today(ctx, uid, today); err == nil {
		t.Error("Today should 404 when no challenge exists")
	}
	if _, err := daily.Complete(ctx, uid, today); err == nil {
		t.Error("Complete should 404 when no challenge exists")
	}

	// Configure today's challenge: a quiz worth 25 bonus XP.
	contentID := newUUIDString(t)
	seedDaily(t, pool, todayStr, "quiz", contentID, 25)

	// Today returns it, not yet completed.
	ch, err := daily.Today(ctx, uid, today)
	if err != nil {
		t.Fatalf("Today: %v", err)
	}
	if ch.Completed || ch.BonusXP != 25 || ch.ContentType != "quiz" || ch.ContentID != contentID || ch.Date != todayStr {
		t.Fatalf("unexpected challenge: %+v", ch)
	}

	// Complete it: newly completed, bonus XP awarded and counted toward the streak.
	res, err := daily.Complete(ctx, uid, today)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if !res.NewlyCompleted {
		t.Error("first completion should be newly completed")
	}
	st, _ := stats.GetStats(ctx, uid)
	if st.TotalXP != 25 {
		t.Errorf("bonus XP not awarded: total=%d, want 25", st.TotalXP)
	}
	if st.CurrentStreak != 1 {
		t.Errorf("daily completion should count toward streak, got %d", st.CurrentStreak)
	}

	// Now reported as completed.
	if ch2, _ := daily.Today(ctx, uid, today); !ch2.Completed {
		t.Error("challenge should show completed after Complete")
	}

	// Repeat completion: idempotent, no extra reward.
	res2, err := daily.Complete(ctx, uid, today)
	if err != nil {
		t.Fatalf("Complete repeat: %v", err)
	}
	if res2.NewlyCompleted {
		t.Error("repeat completion must not be newly completed")
	}
	st2, _ := stats.GetStats(ctx, uid)
	if st2.TotalXP != 25 {
		t.Errorf("repeat must not re-award XP, got %d", st2.TotalXP)
	}
}
