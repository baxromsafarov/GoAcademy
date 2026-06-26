// Package content implements reading of learning content (videos, and later
// articles, quizzes, problems, ...).
package content

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// Pagination bounds applied to list endpoints.
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Service reads content from the database.
type Service struct {
	queries *store.Queries
}

// NewService wires the content service to the database.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{queries: store.New(pool)}
}

// ListFilter holds the optional list filters and pagination window shared by all
// content list endpoints (videos, articles, ...).
type ListFilter struct {
	Difficulty   *string
	Tag          *string
	Language     *string
	Q            *string // case-insensitive title search (videos/articles)
	IncludeHidden bool   // admin-only: also return items tagged "hidden"
	Limit        int
	Offset       int
}

// VideoList is a paginated list result.
type VideoList struct {
	Items  []store.Video
	Total  int64
	Limit  int
	Offset int
}

var validDifficulties = map[string]struct{}{
	string(store.DifficultyBeginner):     {},
	string(store.DifficultyIntermediate): {},
	string(store.DifficultyAdvanced):     {},
}

var validContentLocales = map[string]struct{}{
	string(store.LocaleRu): {},
	string(store.LocaleEn): {},
	string(store.LocaleUz): {},
	string(store.LocaleJa): {},
}

// normalizeAndValidate validates filter values and clamps pagination.
func (f *ListFilter) normalizeAndValidate() error {
	details := map[string]string{}
	if f.Difficulty != nil {
		if _, ok := validDifficulties[*f.Difficulty]; !ok {
			details["difficulty"] = "must be one of beginner, intermediate, advanced"
		}
	}
	if f.Language != nil {
		if _, ok := validContentLocales[*f.Language]; !ok {
			details["language"] = "must be one of ru, en, uz, ja"
		}
	}

	switch {
	case f.Limit <= 0:
		f.Limit = DefaultPageSize
	case f.Limit > MaxPageSize:
		f.Limit = MaxPageSize
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	if len(details) > 0 {
		return apierr.Validation("invalid query parameters").WithDetails(details)
	}
	return nil
}

// ListVideos returns a filtered, paginated list of videos plus the total count.
func (s *Service) ListVideos(ctx context.Context, f ListFilter) (VideoList, error) {
	if err := f.normalizeAndValidate(); err != nil {
		return VideoList{}, err
	}

	listParams := store.ListVideosParams{Lim: int32(f.Limit), Off: int32(f.Offset)}
	countParams := store.CountVideosParams{}
	if f.Difficulty != nil {
		d := store.NullDifficulty{Difficulty: store.Difficulty(*f.Difficulty), Valid: true}
		listParams.Difficulty, countParams.Difficulty = d, d
	}
	if f.Language != nil {
		l := store.NullLocale{Locale: store.Locale(*f.Language), Valid: true}
		listParams.Language, countParams.Language = l, l
	}
	if f.Tag != nil {
		t := pgtype.Text{String: *f.Tag, Valid: true}
		listParams.Tag, countParams.Tag = t, t
	}
	if f.Q != nil {
		q := pgtype.Text{String: *f.Q, Valid: true}
		listParams.Q, countParams.Q = q, q
	}
	listParams.ShowHidden, countParams.ShowHidden = f.IncludeHidden, f.IncludeHidden

	items, err := s.queries.ListVideos(ctx, listParams)
	if err != nil {
		return VideoList{}, err
	}
	total, err := s.queries.CountVideos(ctx, countParams)
	if err != nil {
		return VideoList{}, err
	}
	return VideoList{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset}, nil
}

// GetVideoByID returns a single video, or a 404 if it does not exist.
func (s *Service) GetVideoByID(ctx context.Context, id string) (store.Video, error) {
	uid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.Video{}, apierr.NotFound("video not found")
	}
	v, err := s.queries.GetVideoByID(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.Video{}, apierr.NotFound("video not found")
	}
	if err != nil {
		return store.Video{}, err
	}
	return v, nil
}

// ArticleList is a paginated list of articles.
type ArticleList struct {
	Items  []store.Article
	Total  int64
	Limit  int
	Offset int
}

// ListArticles returns a filtered, paginated list of articles plus the total count.
func (s *Service) ListArticles(ctx context.Context, f ListFilter) (ArticleList, error) {
	if err := f.normalizeAndValidate(); err != nil {
		return ArticleList{}, err
	}

	listParams := store.ListArticlesParams{Lim: int32(f.Limit), Off: int32(f.Offset)}
	countParams := store.CountArticlesParams{}
	if f.Difficulty != nil {
		d := store.NullDifficulty{Difficulty: store.Difficulty(*f.Difficulty), Valid: true}
		listParams.Difficulty, countParams.Difficulty = d, d
	}
	if f.Language != nil {
		l := store.NullLocale{Locale: store.Locale(*f.Language), Valid: true}
		listParams.Language, countParams.Language = l, l
	}
	if f.Tag != nil {
		t := pgtype.Text{String: *f.Tag, Valid: true}
		listParams.Tag, countParams.Tag = t, t
	}
	if f.Q != nil {
		q := pgtype.Text{String: *f.Q, Valid: true}
		listParams.Q, countParams.Q = q, q
	}
	listParams.ShowHidden, countParams.ShowHidden = f.IncludeHidden, f.IncludeHidden

	items, err := s.queries.ListArticles(ctx, listParams)
	if err != nil {
		return ArticleList{}, err
	}
	total, err := s.queries.CountArticles(ctx, countParams)
	if err != nil {
		return ArticleList{}, err
	}
	return ArticleList{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset}, nil
}

// GetArticleBySlug returns a single article by id or slug, or a 404 if it does
// not exist. A ref that parses as a UUID is looked up by id (so track-program
// items, which reference content by id, resolve); anything else is a slug.
func (s *Service) GetArticleBySlug(ctx context.Context, slug string) (store.Article, error) {
	var (
		a   store.Article
		err error
	)
	if id, perr := pgxutil.ParseUUID(slug); perr == nil {
		a, err = s.queries.GetArticleByID(ctx, id)
	} else {
		a, err = s.queries.GetArticleBySlug(ctx, slug)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return store.Article{}, apierr.NotFound("article not found")
	}
	if err != nil {
		return store.Article{}, err
	}
	return a, nil
}

// QuizList is a paginated list of quizzes (metadata only, no questions).
type QuizList struct {
	Items  []store.Quiz
	Total  int64
	Limit  int
	Offset int
}

// ListQuizzes returns a filtered, paginated list of quizzes plus the total count.
func (s *Service) ListQuizzes(ctx context.Context, f ListFilter) (QuizList, error) {
	if err := f.normalizeAndValidate(); err != nil {
		return QuizList{}, err
	}

	listParams := store.ListQuizzesParams{Lim: int32(f.Limit), Off: int32(f.Offset)}
	countParams := store.CountQuizzesParams{}
	if f.Difficulty != nil {
		d := store.NullDifficulty{Difficulty: store.Difficulty(*f.Difficulty), Valid: true}
		listParams.Difficulty, countParams.Difficulty = d, d
	}
	if f.Language != nil {
		l := store.NullLocale{Locale: store.Locale(*f.Language), Valid: true}
		listParams.Language, countParams.Language = l, l
	}
	if f.Tag != nil {
		t := pgtype.Text{String: *f.Tag, Valid: true}
		listParams.Tag, countParams.Tag = t, t
	}
	if f.Q != nil {
		q := pgtype.Text{String: *f.Q, Valid: true}
		listParams.Q, countParams.Q = q, q
	}
	listParams.ShowHidden, countParams.ShowHidden = f.IncludeHidden, f.IncludeHidden

	items, err := s.queries.ListQuizzes(ctx, listParams)
	if err != nil {
		return QuizList{}, err
	}
	total, err := s.queries.CountQuizzes(ctx, countParams)
	if err != nil {
		return QuizList{}, err
	}
	return QuizList{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset}, nil
}

// QuizQuestionDetail is a question with its options.
type QuizQuestionDetail struct {
	Question store.QuizQuestion
	Options  []store.QuizOption // includes is_correct; the HTTP layer strips it for students
}

// QuizDetail is a quiz with its ordered questions and options.
type QuizDetail struct {
	Quiz      store.Quiz
	Questions []QuizQuestionDetail
}

// clampPage applies pagination bounds (default 20, max 100, offset >= 0).
func clampPage(limit, offset int) (int, int) {
	switch {
	case limit <= 0:
		limit = DefaultPageSize
	case limit > MaxPageSize:
		limit = MaxPageSize
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func languageValid(lang *string) bool {
	if lang == nil {
		return true
	}
	_, ok := validContentLocales[*lang]
	return ok
}

func invalidLanguageErr() error {
	return apierr.Validation("invalid query parameters").
		WithDetails(map[string]string{"language": "must be one of ru, en, uz, ja"})
}

// CheatsheetFilter holds the cheatsheet search/filter and pagination.
type CheatsheetFilter struct {
	Category *string
	Query    *string
	Language *string
	Limit    int
	Offset   int
}

// CheatsheetList is a paginated list of cheatsheets.
type CheatsheetList struct {
	Items  []store.Cheatsheet
	Total  int64
	Limit  int
	Offset int
}

// ListCheatsheets returns cheatsheets filtered by category/language and an
// optional free-text query over title and category.
func (s *Service) ListCheatsheets(ctx context.Context, f CheatsheetFilter) (CheatsheetList, error) {
	if !languageValid(f.Language) {
		return CheatsheetList{}, invalidLanguageErr()
	}
	limit, offset := clampPage(f.Limit, f.Offset)

	lp := store.ListCheatsheetsParams{Lim: int32(limit), Off: int32(offset)}
	cp := store.CountCheatsheetsParams{}
	if f.Category != nil {
		t := pgtype.Text{String: *f.Category, Valid: true}
		lp.Category, cp.Category = t, t
	}
	if f.Query != nil {
		t := pgtype.Text{String: *f.Query, Valid: true}
		lp.Q, cp.Q = t, t
	}
	if f.Language != nil {
		l := store.NullLocale{Locale: store.Locale(*f.Language), Valid: true}
		lp.Language, cp.Language = l, l
	}

	items, err := s.queries.ListCheatsheets(ctx, lp)
	if err != nil {
		return CheatsheetList{}, err
	}
	total, err := s.queries.CountCheatsheets(ctx, cp)
	if err != nil {
		return CheatsheetList{}, err
	}
	return CheatsheetList{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

// GetCheatsheetByID returns a single cheatsheet, or a 404.
func (s *Service) GetCheatsheetByID(ctx context.Context, id string) (store.Cheatsheet, error) {
	uid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.Cheatsheet{}, apierr.NotFound("cheatsheet not found")
	}
	c, err := s.queries.GetCheatsheetByID(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return store.Cheatsheet{}, apierr.NotFound("cheatsheet not found")
	}
	if err != nil {
		return store.Cheatsheet{}, err
	}
	return c, nil
}

// GlossaryFilter holds the glossary search and pagination.
type GlossaryFilter struct {
	Query    *string
	Language *string
	Limit    int
	Offset   int
}

// GlossaryList is a paginated list of glossary terms.
type GlossaryList struct {
	Items  []store.GlossaryTerm
	Total  int64
	Limit  int
	Offset int
}

// ListGlossary returns glossary terms filtered by language and an optional
// free-text query over term and definition.
func (s *Service) ListGlossary(ctx context.Context, f GlossaryFilter) (GlossaryList, error) {
	if !languageValid(f.Language) {
		return GlossaryList{}, invalidLanguageErr()
	}
	limit, offset := clampPage(f.Limit, f.Offset)

	lp := store.ListGlossaryTermsParams{Lim: int32(limit), Off: int32(offset)}
	cp := store.CountGlossaryTermsParams{}
	if f.Query != nil {
		t := pgtype.Text{String: *f.Query, Valid: true}
		lp.Q, cp.Q = t, t
	}
	if f.Language != nil {
		l := store.NullLocale{Locale: store.Locale(*f.Language), Valid: true}
		lp.Language, cp.Language = l, l
	}

	items, err := s.queries.ListGlossaryTerms(ctx, lp)
	if err != nil {
		return GlossaryList{}, err
	}
	total, err := s.queries.CountGlossaryTerms(ctx, cp)
	if err != nil {
		return GlossaryList{}, err
	}
	return GlossaryList{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

// ProblemList is a paginated list of problems (metadata only).
type ProblemList struct {
	Items  []store.Problem
	Total  int64
	Limit  int
	Offset int
}

// ListProblems returns a filtered, paginated list of problems plus the total count.
func (s *Service) ListProblems(ctx context.Context, f ListFilter) (ProblemList, error) {
	if err := f.normalizeAndValidate(); err != nil {
		return ProblemList{}, err
	}

	listParams := store.ListProblemsParams{Lim: int32(f.Limit), Off: int32(f.Offset)}
	countParams := store.CountProblemsParams{}
	if f.Difficulty != nil {
		d := store.NullDifficulty{Difficulty: store.Difficulty(*f.Difficulty), Valid: true}
		listParams.Difficulty, countParams.Difficulty = d, d
	}
	if f.Language != nil {
		l := store.NullLocale{Locale: store.Locale(*f.Language), Valid: true}
		listParams.Language, countParams.Language = l, l
	}
	if f.Tag != nil {
		t := pgtype.Text{String: *f.Tag, Valid: true}
		listParams.Tag, countParams.Tag = t, t
	}
	if f.Q != nil {
		q := pgtype.Text{String: *f.Q, Valid: true}
		listParams.Q, countParams.Q = q, q
	}
	listParams.ShowHidden, countParams.ShowHidden = f.IncludeHidden, f.IncludeHidden

	items, err := s.queries.ListProblems(ctx, listParams)
	if err != nil {
		return ProblemList{}, err
	}
	total, err := s.queries.CountProblems(ctx, countParams)
	if err != nil {
		return ProblemList{}, err
	}
	return ProblemList{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset}, nil
}

// ProjectList is a paginated list of mini-projects (metadata only).
type ProjectList struct {
	Items  []store.MiniProject
	Total  int64
	Limit  int
	Offset int
}

// ListProjects returns a filtered, paginated list of mini-projects.
func (s *Service) ListProjects(ctx context.Context, f ListFilter) (ProjectList, error) {
	if err := f.normalizeAndValidate(); err != nil {
		return ProjectList{}, err
	}

	listParams := store.ListProjectsParams{Lim: int32(f.Limit), Off: int32(f.Offset)}
	countParams := store.CountProjectsParams{}
	if f.Difficulty != nil {
		d := store.NullDifficulty{Difficulty: store.Difficulty(*f.Difficulty), Valid: true}
		listParams.Difficulty, countParams.Difficulty = d, d
	}
	if f.Language != nil {
		l := store.NullLocale{Locale: store.Locale(*f.Language), Valid: true}
		listParams.Language, countParams.Language = l, l
	}
	if f.Tag != nil {
		t := pgtype.Text{String: *f.Tag, Valid: true}
		listParams.Tag, countParams.Tag = t, t
	}
	if f.Q != nil {
		q := pgtype.Text{String: *f.Q, Valid: true}
		listParams.Q, countParams.Q = q, q
	}
	listParams.ShowHidden, countParams.ShowHidden = f.IncludeHidden, f.IncludeHidden

	items, err := s.queries.ListProjects(ctx, listParams)
	if err != nil {
		return ProjectList{}, err
	}
	total, err := s.queries.CountProjects(ctx, countParams)
	if err != nil {
		return ProjectList{}, err
	}
	return ProjectList{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset}, nil
}

// ProjectDetail is a mini-project with its ordered checklist steps.
type ProjectDetail struct {
	Project store.MiniProject
	Steps   []store.MiniProjectStep
}

// GetProjectDetail returns a project with its ordered steps, or a 404.
func (s *Service) GetProjectDetail(ctx context.Context, id string) (ProjectDetail, error) {
	uid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return ProjectDetail{}, apierr.NotFound("project not found")
	}
	project, err := s.queries.GetProjectByID(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProjectDetail{}, apierr.NotFound("project not found")
	}
	if err != nil {
		return ProjectDetail{}, err
	}
	steps, err := s.queries.ListProjectSteps(ctx, uid)
	if err != nil {
		return ProjectDetail{}, err
	}
	return ProjectDetail{Project: project, Steps: steps}, nil
}

// GetProblemBySlug returns a problem by id or slug (including the reference
// solution, which the HTTP layer hides until the user has solved it), or a 404.
// A ref that parses as a UUID is looked up by id (so track-program items resolve).
func (s *Service) GetProblemBySlug(ctx context.Context, slug string) (store.Problem, error) {
	var (
		p   store.Problem
		err error
	)
	if id, perr := pgxutil.ParseUUID(slug); perr == nil {
		p, err = s.queries.GetProblemByID(ctx, id)
	} else {
		p, err = s.queries.GetProblemBySlug(ctx, slug)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return store.Problem{}, apierr.NotFound("problem not found")
	}
	if err != nil {
		return store.Problem{}, err
	}
	return p, nil
}

// TrackList is a paginated list of tracks (metadata only).
type TrackList struct {
	Items  []store.Track
	Total  int64
	Limit  int
	Offset int
}

// ListTracks returns a filtered, paginated list of tracks. Difficulty maps to the
// track level; the tag filter does not apply to tracks.
func (s *Service) ListTracks(ctx context.Context, f ListFilter) (TrackList, error) {
	if err := f.normalizeAndValidate(); err != nil {
		return TrackList{}, err
	}

	listParams := store.ListTracksParams{Lim: int32(f.Limit), Off: int32(f.Offset)}
	countParams := store.CountTracksParams{}
	if f.Difficulty != nil {
		d := store.NullDifficulty{Difficulty: store.Difficulty(*f.Difficulty), Valid: true}
		listParams.Level, countParams.Level = d, d
	}
	if f.Language != nil {
		l := store.NullLocale{Locale: store.Locale(*f.Language), Valid: true}
		listParams.Language, countParams.Language = l, l
	}
	if f.Q != nil {
		q := pgtype.Text{String: *f.Q, Valid: true}
		listParams.Q, countParams.Q = q, q
	}

	items, err := s.queries.ListTracks(ctx, listParams)
	if err != nil {
		return TrackList{}, err
	}
	total, err := s.queries.CountTracks(ctx, countParams)
	if err != nil {
		return TrackList{}, err
	}
	return TrackList{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset}, nil
}

// TrackDetail is a track with its ordered program of heterogeneous content items.
type TrackDetail struct {
	Track store.Track
	Items []store.TrackItem
}

// GetTrackDetail returns a track with its items ordered by position, or a 404.
func (s *Service) GetTrackDetail(ctx context.Context, id string) (TrackDetail, error) {
	uid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return TrackDetail{}, apierr.NotFound("track not found")
	}
	track, err := s.queries.GetTrackByID(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return TrackDetail{}, apierr.NotFound("track not found")
	}
	if err != nil {
		return TrackDetail{}, err
	}
	items, err := s.queries.ListTrackItems(ctx, uid)
	if err != nil {
		return TrackDetail{}, err
	}
	return TrackDetail{Track: track, Items: items}, nil
}

// EnrollTrack opts the user into a track so it surfaces on their dashboard. It
// is idempotent — enrolling twice is a no-op.
func (s *Service) EnrollTrack(ctx context.Context, userID, trackID string) error {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return apierr.Validation("invalid user")
	}
	tid, err := pgxutil.ParseUUID(trackID)
	if err != nil {
		return apierr.NotFound("track not found")
	}
	return s.queries.EnrollTrack(ctx, store.EnrollTrackParams{UserID: uid, TrackID: tid})
}

// UnenrollTrack removes the user's enrollment in a track.
func (s *Service) UnenrollTrack(ctx context.Context, userID, trackID string) error {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return apierr.Validation("invalid user")
	}
	tid, err := pgxutil.ParseUUID(trackID)
	if err != nil {
		return apierr.NotFound("track not found")
	}
	_, err = s.queries.UnenrollTrack(ctx, store.UnenrollTrackParams{UserID: uid, TrackID: tid})
	return err
}

// ListEnrolledTracks returns the tracks the user enrolled in, most recent first.
func (s *Service) ListEnrolledTracks(ctx context.Context, userID string) ([]store.Track, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return nil, apierr.Validation("invalid user")
	}
	return s.queries.ListEnrolledTracks(ctx, uid)
}

// RecentCompletion is one finished item in the user's activity feed.
type RecentCompletion struct {
	ContentType string
	ContentID   string
	Title       string
	CompletedAt time.Time
}

// ListRecentCompletions returns the user's most recently finished content
// (videos watched, articles read, quizzes passed, problems solved).
func (s *Service) ListRecentCompletions(ctx context.Context, userID string, limit int) ([]RecentCompletion, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return nil, apierr.Validation("invalid user")
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	rows, err := s.queries.ListRecentCompletions(ctx, store.ListRecentCompletionsParams{
		UserID: uid, Limit: int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]RecentCompletion, 0, len(rows))
	for _, r := range rows {
		out = append(out, RecentCompletion{
			ContentType: r.ContentType,
			ContentID:   pgxutil.UUIDString(r.ContentID),
			Title:       r.Title,
			CompletedAt: r.CompletedAt.Time,
		})
	}
	return out, nil
}

// GetQuizDetail returns a quiz with its questions and options (ordered).
func (s *Service) GetQuizDetail(ctx context.Context, id string) (QuizDetail, error) {
	uid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return QuizDetail{}, apierr.NotFound("quiz not found")
	}
	quiz, err := s.queries.GetQuizByID(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return QuizDetail{}, apierr.NotFound("quiz not found")
	}
	if err != nil {
		return QuizDetail{}, err
	}

	questions, err := s.queries.ListQuizQuestions(ctx, uid)
	if err != nil {
		return QuizDetail{}, err
	}
	options, err := s.queries.ListQuizOptionsByQuiz(ctx, uid)
	if err != nil {
		return QuizDetail{}, err
	}

	byQuestion := make(map[pgtype.UUID][]store.QuizOption, len(questions))
	for _, o := range options {
		byQuestion[o.QuestionID] = append(byQuestion[o.QuestionID], o)
	}

	details := make([]QuizQuestionDetail, 0, len(questions))
	for _, q := range questions {
		details = append(details, QuizQuestionDetail{Question: q, Options: byQuestion[q.ID]})
	}
	return QuizDetail{Quiz: quiz, Questions: details}, nil
}
