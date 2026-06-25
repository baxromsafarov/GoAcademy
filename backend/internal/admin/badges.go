package admin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// BadgeInput is the create/update payload for a badge.
type BadgeInput struct {
	Code           string
	Title          string
	Description    string
	Icon           string
	CriteriaType   string
	CriteriaParams json.RawMessage
}

func (in BadgeInput) validate() error {
	details := map[string]string{}
	if strings.TrimSpace(in.Code) == "" {
		details["code"] = "must not be empty"
	}
	if strings.TrimSpace(in.Title) == "" {
		details["title"] = "must not be empty"
	}
	if strings.TrimSpace(in.CriteriaType) == "" {
		details["criteria_type"] = "must not be empty"
	}
	if len(in.CriteriaParams) > 0 && !json.Valid(in.CriteriaParams) {
		details["criteria_params"] = "must be valid JSON"
	}
	if len(details) > 0 {
		return apierr.Validation("invalid badge").WithDetails(details)
	}
	return nil
}

func criteriaParamsBytes(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}

// CreateBadge inserts a new badge (code must be unique).
func (s *Service) CreateBadge(ctx context.Context, in BadgeInput) (store.Badge, error) {
	if err := in.validate(); err != nil {
		return store.Badge{}, err
	}
	b, err := s.queries.CreateBadge(ctx, store.CreateBadgeParams{
		Code: in.Code, Title: in.Title, Description: in.Description, Icon: in.Icon,
		CriteriaType: in.CriteriaType, CriteriaParams: criteriaParamsBytes(in.CriteriaParams),
	})
	if isUniqueViolation(err) {
		return store.Badge{}, apierr.Conflict("a badge with this code already exists")
	}
	return b, err
}

// UpdateBadge replaces a badge's fields.
func (s *Service) UpdateBadge(ctx context.Context, id string, in BadgeInput) (store.Badge, error) {
	bid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.Badge{}, apierr.NotFound("badge not found")
	}
	if err := in.validate(); err != nil {
		return store.Badge{}, err
	}
	b, err := s.queries.UpdateBadge(ctx, store.UpdateBadgeParams{
		ID: bid, Code: in.Code, Title: in.Title, Description: in.Description, Icon: in.Icon,
		CriteriaType: in.CriteriaType, CriteriaParams: criteriaParamsBytes(in.CriteriaParams),
	})
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return store.Badge{}, apierr.NotFound("badge not found")
	case isUniqueViolation(err):
		return store.Badge{}, apierr.Conflict("a badge with this code already exists")
	}
	return b, err
}

// DeleteBadge removes a badge.
func (s *Service) DeleteBadge(ctx context.Context, id string) error {
	bid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return apierr.NotFound("badge not found")
	}
	n, err := s.queries.DeleteBadge(ctx, bid)
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("badge not found")
	}
	return nil
}

// DailyChallengeInput is the create/update payload for a daily challenge.
type DailyChallengeInput struct {
	ChallengeDate string // "YYYY-MM-DD"
	ContentType   string
	ContentID     string
	BonusXP       int
}

func (in DailyChallengeInput) validate() (pgtype.Date, pgtype.UUID, error) {
	details := map[string]string{}
	date, dateErr := time.Parse("2006-01-02", in.ChallengeDate)
	if dateErr != nil {
		details["challenge_date"] = "must be a date in YYYY-MM-DD form"
	}
	if !trackContentTypes[in.ContentType] {
		details["content_type"] = "must be video, article, quiz, problem or project"
	}
	cid, cidErr := pgxutil.ParseUUID(in.ContentID)
	if cidErr != nil {
		details["content_id"] = "must be a valid uuid"
	}
	if in.BonusXP < 0 {
		details["bonus_xp"] = "must be >= 0"
	}
	if len(details) > 0 {
		return pgtype.Date{}, pgtype.UUID{}, apierr.Validation("invalid daily challenge").WithDetails(details)
	}
	return pgtype.Date{Time: date, Valid: true}, cid, nil
}

// CreateDailyChallenge schedules a challenge for a date (date must be unique).
func (s *Service) CreateDailyChallenge(ctx context.Context, in DailyChallengeInput) (store.DailyChallenge, error) {
	date, cid, err := in.validate()
	if err != nil {
		return store.DailyChallenge{}, err
	}
	d, err := s.queries.CreateDailyChallenge(ctx, store.CreateDailyChallengeParams{
		ChallengeDate: date, ContentType: store.TrackContentType(in.ContentType),
		ContentID: cid, BonusXp: int32(in.BonusXP),
	})
	if isUniqueViolation(err) {
		return store.DailyChallenge{}, apierr.Conflict("a challenge already exists for this date")
	}
	return d, err
}

// UpdateDailyChallenge replaces a daily challenge's fields.
func (s *Service) UpdateDailyChallenge(ctx context.Context, id string, in DailyChallengeInput) (store.DailyChallenge, error) {
	did, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.DailyChallenge{}, apierr.NotFound("daily challenge not found")
	}
	date, cid, err := in.validate()
	if err != nil {
		return store.DailyChallenge{}, err
	}
	d, err := s.queries.UpdateDailyChallenge(ctx, store.UpdateDailyChallengeParams{
		ID: did, ChallengeDate: date, ContentType: store.TrackContentType(in.ContentType),
		ContentID: cid, BonusXp: int32(in.BonusXP),
	})
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return store.DailyChallenge{}, apierr.NotFound("daily challenge not found")
	case isUniqueViolation(err):
		return store.DailyChallenge{}, apierr.Conflict("a challenge already exists for this date")
	}
	return d, err
}

// DeleteDailyChallenge removes a daily challenge.
func (s *Service) DeleteDailyChallenge(ctx context.Context, id string) error {
	did, err := pgxutil.ParseUUID(id)
	if err != nil {
		return apierr.NotFound("daily challenge not found")
	}
	n, err := s.queries.DeleteDailyChallenge(ctx, did)
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("daily challenge not found")
	}
	return nil
}
