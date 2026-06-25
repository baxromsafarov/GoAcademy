// Package admin provides admin-only CRUD over content. All of it is mounted
// behind RequireRole("admin").
package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// Service is the admin content-management service.
type Service struct {
	pool    *pgxpool.Pool
	queries *store.Queries
}

// NewService wires the admin service to the database.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, queries: store.New(pool)}
}

var difficulties = map[string]bool{"beginner": true, "intermediate": true, "advanced": true}
var locales = map[string]bool{"ru": true, "en": true, "uz": true, "ja": true}

// trackContentTypes are the values of the track_content_type enum (used by track
// items and daily challenges).
var trackContentTypes = map[string]bool{
	"video": true, "article": true, "quiz": true, "problem": true, "project": true,
}

// validateMeta records errors for the content fields common to all types.
func validateMeta(details map[string]string, title, difficulty, language string) {
	if strings.TrimSpace(title) == "" {
		details["title"] = "must not be empty"
	}
	if !difficulties[difficulty] {
		details["difficulty"] = "must be beginner, intermediate or advanced"
	}
	if !locales[language] {
		details["language"] = "must be ru, en, uz or ja"
	}
}

func normalizeTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// VideoInput is the create/update payload for a video.
type VideoInput struct {
	Title           string
	Description     string
	YoutubeID       string
	DurationSeconds int
	Difficulty      string
	Language        string
	Tags            []string
}

func (in VideoInput) validate() error {
	details := map[string]string{}
	validateMeta(details, in.Title, in.Difficulty, in.Language)
	if strings.TrimSpace(in.YoutubeID) == "" {
		details["youtube_id"] = "must not be empty"
	}
	if in.DurationSeconds < 0 {
		details["duration_seconds"] = "must be >= 0"
	}
	if len(details) > 0 {
		return apierr.Validation("invalid video").WithDetails(details)
	}
	return nil
}

// CreateVideo inserts a new video.
func (s *Service) CreateVideo(ctx context.Context, in VideoInput) (store.Video, error) {
	if err := in.validate(); err != nil {
		return store.Video{}, err
	}
	return s.queries.CreateVideo(ctx, store.CreateVideoParams{
		Title: in.Title, Description: in.Description, YoutubeID: in.YoutubeID,
		DurationSeconds: int32(in.DurationSeconds), Difficulty: store.Difficulty(in.Difficulty),
		Tags: normalizeTags(in.Tags), Language: store.Locale(in.Language),
	})
}

// UpdateVideo replaces an existing video's fields.
func (s *Service) UpdateVideo(ctx context.Context, id string, in VideoInput) (store.Video, error) {
	vid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.Video{}, apierr.NotFound("video not found")
	}
	if err := in.validate(); err != nil {
		return store.Video{}, err
	}
	v, err := s.queries.UpdateVideo(ctx, store.UpdateVideoParams{
		ID: vid, Title: in.Title, Description: in.Description, YoutubeID: in.YoutubeID,
		DurationSeconds: int32(in.DurationSeconds), Difficulty: store.Difficulty(in.Difficulty),
		Tags: normalizeTags(in.Tags), Language: store.Locale(in.Language),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return store.Video{}, apierr.NotFound("video not found")
	}
	return v, err
}

// DeleteVideo removes a video.
func (s *Service) DeleteVideo(ctx context.Context, id string) error {
	vid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return apierr.NotFound("video not found")
	}
	n, err := s.queries.DeleteVideo(ctx, vid)
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("video not found")
	}
	return nil
}

// ArticleInput is the create/update payload for an article.
type ArticleInput struct {
	Title        string
	Slug         string
	BodyMarkdown string
	Difficulty   string
	Language     string
	Tags         []string
}

func (in ArticleInput) validate() error {
	details := map[string]string{}
	validateMeta(details, in.Title, in.Difficulty, in.Language)
	if strings.TrimSpace(in.Slug) == "" {
		details["slug"] = "must not be empty"
	}
	if len(details) > 0 {
		return apierr.Validation("invalid article").WithDetails(details)
	}
	return nil
}

// CreateArticle inserts a new article (slug must be unique).
func (s *Service) CreateArticle(ctx context.Context, in ArticleInput) (store.Article, error) {
	if err := in.validate(); err != nil {
		return store.Article{}, err
	}
	a, err := s.queries.CreateArticle(ctx, store.CreateArticleParams{
		Title: in.Title, Slug: in.Slug, BodyMarkdown: in.BodyMarkdown,
		Difficulty: store.Difficulty(in.Difficulty), Tags: normalizeTags(in.Tags), Language: store.Locale(in.Language),
	})
	if isUniqueViolation(err) {
		return store.Article{}, apierr.Conflict("an article with this slug already exists")
	}
	return a, err
}

// UpdateArticle replaces an existing article's fields.
func (s *Service) UpdateArticle(ctx context.Context, id string, in ArticleInput) (store.Article, error) {
	aid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.Article{}, apierr.NotFound("article not found")
	}
	if err := in.validate(); err != nil {
		return store.Article{}, err
	}
	a, err := s.queries.UpdateArticle(ctx, store.UpdateArticleParams{
		ID: aid, Title: in.Title, Slug: in.Slug, BodyMarkdown: in.BodyMarkdown,
		Difficulty: store.Difficulty(in.Difficulty), Tags: normalizeTags(in.Tags), Language: store.Locale(in.Language),
	})
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return store.Article{}, apierr.NotFound("article not found")
	case isUniqueViolation(err):
		return store.Article{}, apierr.Conflict("an article with this slug already exists")
	}
	return a, err
}

// DeleteArticle removes an article.
func (s *Service) DeleteArticle(ctx context.Context, id string) error {
	aid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return apierr.NotFound("article not found")
	}
	n, err := s.queries.DeleteArticle(ctx, aid)
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("article not found")
	}
	return nil
}
