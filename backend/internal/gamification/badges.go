package gamification

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/store"
)

// awardBadges grants every not-yet-earned badge whose criterion the user now
// meets. It runs inside the recorder's transaction so awards are consistent with
// the stats that earned them, and is idempotent: ListUnearnedBadges skips earned
// badges and AwardBadge is ON CONFLICT DO NOTHING.
func awardBadges(ctx context.Context, q *store.Queries, uid pgtype.UUID, totalXP, currentStreak int) error {
	badges, err := q.ListUnearnedBadges(ctx, uid)
	if err != nil {
		return err
	}
	for _, b := range badges {
		met, err := criterionMet(ctx, q, uid, b, totalXP, currentStreak)
		if err != nil {
			return err
		}
		if met {
			if err := q.AwardBadge(ctx, store.AwardBadgeParams{UserID: uid, BadgeID: b.ID}); err != nil {
				return err
			}
		}
	}
	return nil
}

// criterionMet reports whether the user satisfies a badge's criterion. An unknown
// criteria_type is never met, so introducing a new badge type cannot break the
// existing ones — code support can be added later without a data migration.
func criterionMet(ctx context.Context, q *store.Queries, uid pgtype.UUID, b store.Badge, totalXP, currentStreak int) (bool, error) {
	switch b.CriteriaType {
	case "xp_at_least":
		var p struct {
			XP int `json:"xp"`
		}
		if err := json.Unmarshal(b.CriteriaParams, &p); err != nil {
			return false, err
		}
		return totalXP >= p.XP, nil

	case "streak_at_least":
		var p struct {
			Days int `json:"days"`
		}
		if err := json.Unmarshal(b.CriteriaParams, &p); err != nil {
			return false, err
		}
		return currentStreak >= p.Days, nil

	case "activity_count_at_least":
		var p struct {
			ActivityType string `json:"activity_type"`
			Count        int    `json:"count"`
		}
		if err := json.Unmarshal(b.CriteriaParams, &p); err != nil {
			return false, err
		}
		n, err := q.ActivityRefCount(ctx, store.ActivityRefCountParams{UserID: uid, ActivityType: p.ActivityType})
		if err != nil {
			return false, err
		}
		return int(n) >= p.Count, nil

	default:
		return false, nil // unknown criterion: awarded only once code supports it
	}
}
