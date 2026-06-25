package gamification

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// Stats is a user's gamification summary.
type Stats struct {
	TotalXP        int
	Level          int
	CurrentStreak  int
	LongestStreak  int
	LastActiveDate string // "YYYY-MM-DD", or "" if never active
}

// Service reads gamification stats.
type Service struct {
	queries *store.Queries
}

// NewService wires the gamification read service to the database.
func NewService(pool *pgxpool.Pool) *Service { return &Service{queries: store.New(pool)} }

// GetStats returns the user's XP/level/streak summary. A user with no recorded
// activity has no user_stats row yet and reports zero XP at level 1.
func (s *Service) GetStats(ctx context.Context, userID string) (Stats, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return Stats{}, apierr.Unauthorized("invalid user")
	}
	st, err := s.queries.GetUserStats(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return Stats{Level: 1}, nil
	}
	if err != nil {
		return Stats{}, err
	}
	lastActive := ""
	if st.LastActiveDate.Valid {
		lastActive = st.LastActiveDate.Time.Format("2006-01-02")
	}
	return Stats{
		TotalXP:        int(st.TotalXp),
		Level:          int(st.Level),
		CurrentStreak:  int(st.CurrentStreak),
		LongestStreak:  int(st.LongestStreak),
		LastActiveDate: lastActive,
	}, nil
}

// Badge is an earned achievement.
type Badge struct {
	Code        string
	Title       string
	Description string
	Icon        string
	AwardedAt   time.Time
}

// GetBadges returns the user's earned badges, oldest award first.
func (s *Service) GetBadges(ctx context.Context, userID string) ([]Badge, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return nil, apierr.Unauthorized("invalid user")
	}
	rows, err := s.queries.ListUserBadges(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]Badge, 0, len(rows))
	for _, r := range rows {
		out = append(out, Badge{
			Code:        r.Code,
			Title:       r.Title,
			Description: r.Description,
			Icon:        r.Icon,
			AwardedAt:   r.AwardedAt.Time,
		})
	}
	return out, nil
}
