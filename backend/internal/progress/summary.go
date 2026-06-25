package progress

import (
	"context"
	"fmt"
	"time"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

const (
	defaultHeatmapDays = 365 // GitHub-style one-year window when no range is given
	maxHeatmapDays     = 366 // bound on the response size (one leap year)
	dateLayout         = "2006-01-02"
)

// ProgressSummaryResult is the authenticated user's per-section completion counts.
type ProgressSummaryResult struct {
	VideosCompleted   int64
	ArticlesRead      int64
	QuizzesPassed     int64
	ProblemsSolved    int64
	ProjectsCompleted int64
}

// ActivityDay is one bucket of the activity heatmap: a UTC calendar day with the
// number of actions and XP earned that day.
type ActivityDay struct {
	Day   string // UTC calendar day, "YYYY-MM-DD"
	Count int64
	XP    int64
}

// ProgressSummary returns the user's completion counts across all content sections.
func (s *Service) ProgressSummary(ctx context.Context, userID string) (ProgressSummaryResult, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return ProgressSummaryResult{}, apierr.Unauthorized("invalid user")
	}
	row, err := s.queries.ProgressSummary(ctx, uid)
	if err != nil {
		return ProgressSummaryResult{}, err
	}
	return ProgressSummaryResult{
		VideosCompleted:   row.VideosCompleted,
		ArticlesRead:      row.ArticlesRead,
		QuizzesPassed:     row.QuizzesPassed,
		ProblemsSolved:    row.ProblemsSolved,
		ProjectsCompleted: row.ProjectsCompleted,
	}, nil
}

// ActivityHeatmap returns daily activity buckets over the half-open range
// [from, to) (UTC). Days with no activity are absent (callers fill gaps).
func (s *Service) ActivityHeatmap(ctx context.Context, userID string, from, to time.Time) ([]ActivityDay, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return nil, apierr.Unauthorized("invalid user")
	}
	rows, err := s.queries.ActivityHeatmap(ctx, store.ActivityHeatmapParams{
		UserID: uid,
		FromTs: pgxutil.Timestamptz(from),
		ToTs:   pgxutil.Timestamptz(to),
	})
	if err != nil {
		return nil, err
	}
	days := make([]ActivityDay, 0, len(rows))
	for _, r := range rows {
		days = append(days, ActivityDay{Day: r.Day, Count: r.Count, XP: r.Xp})
	}
	return days, nil
}

// ParseHeatmapRange resolves the optional from/to query dates (UTC "YYYY-MM-DD")
// into a half-open range [from, toExclusive). Empty values default to a one-year
// window ending today (now). It enforces from<=to and a maximum span.
func ParseHeatmapRange(fromStr, toStr string, now time.Time) (from, toExclusive time.Time, err error) {
	startOfDay := func(t time.Time) time.Time {
		t = t.UTC()
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}
	parseDate := func(s string) (time.Time, error) {
		t, e := time.Parse(dateLayout, s) // no zone in layout => UTC
		if e != nil {
			return time.Time{}, apierr.Validation("invalid date: expected YYYY-MM-DD")
		}
		return t, nil
	}

	toIncl := startOfDay(now)
	if toStr != "" {
		if toIncl, err = parseDate(toStr); err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	fromIncl := toIncl.AddDate(0, 0, -(defaultHeatmapDays - 1))
	if fromStr != "" {
		if fromIncl, err = parseDate(fromStr); err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if fromIncl.After(toIncl) {
		return time.Time{}, time.Time{}, apierr.Validation("'from' must be on or before 'to'")
	}
	if days := int(toIncl.Sub(fromIncl).Hours()/24) + 1; days > maxHeatmapDays {
		return time.Time{}, time.Time{}, apierr.Validation(fmt.Sprintf("range too large: at most %d days", maxHeatmapDays))
	}
	return fromIncl, toIncl.AddDate(0, 0, 1), nil
}
