// Package social serves cross-user features: the leaderboard (and later notes,
// bookmarks and certificates).
package social

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

const (
	defaultLeaderboardLimit = 20
	maxLeaderboardLimit     = 100
)

// LeaderboardEntry is one ranked user.
type LeaderboardEntry struct {
	Rank        int
	UserID      string
	DisplayName string
	AvatarURL   string
	XP          int64
}

// Service serves social features.
type Service struct {
	queries *store.Queries
}

// NewService wires the social service to the database.
func NewService(pool *pgxpool.Pool) *Service { return &Service{queries: store.New(pool)} }

// Leaderboard ranks public, non-blocked users for the given period
// ("all" | "week" | "month"). now is injected so period windows are testable.
func (s *Service) Leaderboard(ctx context.Context, period string, now time.Time, limit, offset int) ([]LeaderboardEntry, error) {
	limit, offset = clampPage(limit, offset)
	from, to, allTime, err := periodWindow(period, now)
	if err != nil {
		return nil, err
	}

	if allTime {
		rows, err := s.queries.LeaderboardAllTime(ctx, store.LeaderboardAllTimeParams{
			Lim: int32(limit), Off: int32(offset),
		})
		if err != nil {
			return nil, err
		}
		out := make([]LeaderboardEntry, len(rows))
		for i, r := range rows {
			out[i] = LeaderboardEntry{
				Rank: offset + i + 1, UserID: pgxutil.UUIDString(r.ID),
				DisplayName: r.DisplayName, AvatarURL: r.AvatarUrl.String, XP: r.Xp,
			}
		}
		return out, nil
	}

	rows, err := s.queries.LeaderboardPeriod(ctx, store.LeaderboardPeriodParams{
		FromTs: pgxutil.Timestamptz(from), ToTs: pgxutil.Timestamptz(to),
		Lim: int32(limit), Off: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]LeaderboardEntry, len(rows))
	for i, r := range rows {
		out[i] = LeaderboardEntry{
			Rank: offset + i + 1, UserID: pgxutil.UUIDString(r.ID),
			DisplayName: r.DisplayName, AvatarURL: r.AvatarUrl.String, XP: r.Xp,
		}
	}
	return out, nil
}

// periodWindow resolves a period name into a UTC window. "all"/"" => all-time
// (allTime=true, dates unused); "week"/"month" => the last 7/30 days ending today.
func periodWindow(period string, now time.Time) (from, to time.Time, allTime bool, err error) {
	t := now.UTC()
	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	toExclusive := today.AddDate(0, 0, 1)
	switch period {
	case "", "all":
		return time.Time{}, time.Time{}, true, nil
	case "week":
		return today.AddDate(0, 0, -6), toExclusive, false, nil
	case "month":
		return today.AddDate(0, 0, -29), toExclusive, false, nil
	default:
		return time.Time{}, time.Time{}, false, apierr.Validation("invalid period: use all, week or month")
	}
}

func clampPage(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultLeaderboardLimit
	}
	if limit > maxLeaderboardLimit {
		limit = maxLeaderboardLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
