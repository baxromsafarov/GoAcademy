package httpapi

import (
	"encoding/json"
	"time"

	"github.com/goacademy/backend/internal/admin"
	"github.com/goacademy/backend/internal/content"
	"github.com/goacademy/backend/internal/gamification"
	"github.com/goacademy/backend/internal/judge"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/progress"
	"github.com/goacademy/backend/internal/quiz"
	"github.com/goacademy/backend/internal/social"
	"github.com/goacademy/backend/internal/store"
)

// userResponse is the public projection of a user. It deliberately omits
// password_hash and other sensitive/internal fields.
type userResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	Role          string    `json:"role"`
	Locale        string    `json:"locale"`
	Bio           string    `json:"bio"`
	Location      string    `json:"location"`
	AvatarURL     string    `json:"avatar_url"`
	EmailVerified bool      `json:"email_verified"`
	IsPublic      bool      `json:"is_public"`
	CreatedAt     time.Time `json:"created_at"`
}

func toUserResponse(u store.User) userResponse {
	return userResponse{
		ID:            pgxutil.UUIDString(u.ID),
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		Role:          string(u.Role),
		Locale:        string(u.Locale),
		Bio:           u.Bio,
		Location:      u.Location,
		AvatarURL:     u.AvatarUrl.String, // "" when null
		EmailVerified: u.EmailVerified,
		IsPublic:      u.IsPublic,
		CreatedAt:     u.CreatedAt.Time,
	}
}

// videoResponse is the public projection of a video.
type videoResponse struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	YoutubeID       string    `json:"youtube_id"`
	DurationSeconds int32     `json:"duration_seconds"`
	Difficulty      string    `json:"difficulty"`
	Tags            []string  `json:"tags"`
	Language        string    `json:"language"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toVideoResponse(v store.Video) videoResponse {
	return videoResponse{
		ID:              pgxutil.UUIDString(v.ID),
		Title:           v.Title,
		Description:     v.Description,
		YoutubeID:       v.YoutubeID,
		DurationSeconds: v.DurationSeconds,
		Difficulty:      string(v.Difficulty),
		Tags:            v.Tags,
		Language:        string(v.Language),
		CreatedAt:       v.CreatedAt.Time,
		UpdatedAt:       v.UpdatedAt.Time,
	}
}

// videoListResponse is a paginated list of videos.
type videoListResponse struct {
	Items  []videoResponse `json:"items"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

func toVideoListResponse(l content.VideoList) videoListResponse {
	items := make([]videoResponse, 0, len(l.Items))
	for _, v := range l.Items {
		items = append(items, toVideoResponse(v))
	}
	return videoListResponse{Items: items, Total: l.Total, Limit: l.Limit, Offset: l.Offset}
}

// articleResponse is the public projection of an article (full body included).
type articleResponse struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	BodyMarkdown string    `json:"body_markdown"`
	Difficulty   string    `json:"difficulty"`
	Tags         []string  `json:"tags"`
	Language     string    `json:"language"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func toArticleResponse(a store.Article) articleResponse {
	return articleResponse{
		ID:           pgxutil.UUIDString(a.ID),
		Title:        a.Title,
		Slug:         a.Slug,
		BodyMarkdown: a.BodyMarkdown,
		Difficulty:   string(a.Difficulty),
		Tags:         a.Tags,
		Language:     string(a.Language),
		CreatedAt:    a.CreatedAt.Time,
		UpdatedAt:    a.UpdatedAt.Time,
	}
}

// articleListResponse is a paginated list of articles. The list omits the full
// body to keep payloads small; the detail endpoint returns it.
type articleListItem struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug"`
	Difficulty string    `json:"difficulty"`
	Tags       []string  `json:"tags"`
	Language   string    `json:"language"`
	CreatedAt  time.Time `json:"created_at"`
}

type articleListResponse struct {
	Items  []articleListItem `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

func toArticleListResponse(l content.ArticleList) articleListResponse {
	items := make([]articleListItem, 0, len(l.Items))
	for _, a := range l.Items {
		items = append(items, articleListItem{
			ID:         pgxutil.UUIDString(a.ID),
			Title:      a.Title,
			Slug:       a.Slug,
			Difficulty: string(a.Difficulty),
			Tags:       a.Tags,
			Language:   string(a.Language),
			CreatedAt:  a.CreatedAt.Time,
		})
	}
	return articleListResponse{Items: items, Total: l.Total, Limit: l.Limit, Offset: l.Offset}
}

// trackListItem is a track's metadata for list responses.
type trackListItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Level       string    `json:"level"`
	Language    string    `json:"language"`
	Position    int32     `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
}

type trackListResponse struct {
	Items  []trackListItem `json:"items"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

func toTrackListResponse(l content.TrackList) trackListResponse {
	items := make([]trackListItem, 0, len(l.Items))
	for _, tr := range l.Items {
		items = append(items, trackListItem{
			ID:          pgxutil.UUIDString(tr.ID),
			Title:       tr.Title,
			Description: tr.Description,
			Level:       string(tr.Level),
			Language:    string(tr.Language),
			Position:    tr.Position,
			CreatedAt:   tr.CreatedAt.Time,
		})
	}
	return trackListResponse{Items: items, Total: l.Total, Limit: l.Limit, Offset: l.Offset}
}

// trackItemResponse is one entry in a track's ordered program (a typed reference
// to content).
type trackItemResponse struct {
	ContentType string `json:"content_type"`
	ContentID   string `json:"content_id"`
	Position    int32  `json:"position"`
}

type trackDetailResponse struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Level       string              `json:"level"`
	Language    string              `json:"language"`
	Position    int32               `json:"position"`
	Items       []trackItemResponse `json:"items"`
}

func toTrackDetailResponse(d content.TrackDetail) trackDetailResponse {
	items := make([]trackItemResponse, 0, len(d.Items))
	for _, it := range d.Items {
		items = append(items, trackItemResponse{
			ContentType: string(it.ContentType),
			ContentID:   pgxutil.UUIDString(it.ContentID),
			Position:    it.Position,
		})
	}
	return trackDetailResponse{
		ID:          pgxutil.UUIDString(d.Track.ID),
		Title:       d.Track.Title,
		Description: d.Track.Description,
		Level:       string(d.Track.Level),
		Language:    string(d.Track.Language),
		Position:    d.Track.Position,
		Items:       items,
	}
}

// project DTOs ------------------------------------------------------------------

type projectListItem struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Difficulty string    `json:"difficulty"`
	Tags       []string  `json:"tags"`
	Language   string    `json:"language"`
	CreatedAt  time.Time `json:"created_at"`
}

type projectListResponse struct {
	Items  []projectListItem `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

func toProjectListResponse(l content.ProjectList) projectListResponse {
	items := make([]projectListItem, 0, len(l.Items))
	for _, p := range l.Items {
		items = append(items, projectListItem{
			ID:         pgxutil.UUIDString(p.ID),
			Title:      p.Title,
			Difficulty: string(p.Difficulty),
			Tags:       p.Tags,
			Language:   string(p.Language),
			CreatedAt:  p.CreatedAt.Time,
		})
	}
	return projectListResponse{Items: items, Total: l.Total, Limit: l.Limit, Offset: l.Offset}
}

type projectStepResponse struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Position int32  `json:"position"`
}

type projectDetailResponse struct {
	ID                  string                `json:"id"`
	Title               string                `json:"title"`
	DescriptionMarkdown string                `json:"description_markdown"`
	Difficulty          string                `json:"difficulty"`
	Tags                []string              `json:"tags"`
	Language            string                `json:"language"`
	Steps               []projectStepResponse `json:"steps"`
}

func toProjectDetailResponse(d content.ProjectDetail) projectDetailResponse {
	steps := make([]projectStepResponse, 0, len(d.Steps))
	for _, st := range d.Steps {
		steps = append(steps, projectStepResponse{
			ID:       pgxutil.UUIDString(st.ID),
			Text:     st.Text,
			Position: st.Position,
		})
	}
	return projectDetailResponse{
		ID:                  pgxutil.UUIDString(d.Project.ID),
		Title:               d.Project.Title,
		DescriptionMarkdown: d.Project.DescriptionMarkdown,
		Difficulty:          string(d.Project.Difficulty),
		Tags:                d.Project.Tags,
		Language:            string(d.Project.Language),
		Steps:               steps,
	}
}

type projectProgressResponse struct {
	ProjectID        string   `json:"project_id"`
	CompletedStepIDs []string `json:"completed_step_ids"`
	Total            int      `json:"total"`
	Completed        int      `json:"completed"`
	ProjectComplete  bool     `json:"project_complete"`
}

func toProjectProgressResponse(r progress.ProjectProgressResult) projectProgressResponse {
	ids := r.CompletedStepIDs
	if ids == nil {
		ids = []string{}
	}
	return projectProgressResponse{
		ProjectID:        r.ProjectID,
		CompletedStepIDs: ids,
		Total:            r.Total,
		Completed:        r.Completed,
		ProjectComplete:  r.ProjectComplete,
	}
}

// cheatsheet DTOs ---------------------------------------------------------------

type cheatsheetListItem struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Category  string    `json:"category"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
}

type cheatsheetListResponse struct {
	Items  []cheatsheetListItem `json:"items"`
	Total  int64                `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
}

func toCheatsheetListResponse(l content.CheatsheetList) cheatsheetListResponse {
	items := make([]cheatsheetListItem, 0, len(l.Items))
	for _, c := range l.Items {
		items = append(items, cheatsheetListItem{
			ID:        pgxutil.UUIDString(c.ID),
			Title:     c.Title,
			Category:  c.Category,
			Language:  string(c.Language),
			CreatedAt: c.CreatedAt.Time,
		})
	}
	return cheatsheetListResponse{Items: items, Total: l.Total, Limit: l.Limit, Offset: l.Offset}
}

type cheatsheetDetailResponse struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Category     string    `json:"category"`
	BodyMarkdown string    `json:"body_markdown"`
	Language     string    `json:"language"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func toCheatsheetDetailResponse(c store.Cheatsheet) cheatsheetDetailResponse {
	return cheatsheetDetailResponse{
		ID:           pgxutil.UUIDString(c.ID),
		Title:        c.Title,
		Category:     c.Category,
		BodyMarkdown: c.BodyMarkdown,
		Language:     string(c.Language),
		CreatedAt:    c.CreatedAt.Time,
		UpdatedAt:    c.UpdatedAt.Time,
	}
}

// glossary DTOs -----------------------------------------------------------------

type glossaryItem struct {
	ID                 string `json:"id"`
	Term               string `json:"term"`
	DefinitionMarkdown string `json:"definition_markdown"`
	Language           string `json:"language"`
}

type glossaryListResponse struct {
	Items  []glossaryItem `json:"items"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func toGlossaryListResponse(l content.GlossaryList) glossaryListResponse {
	items := make([]glossaryItem, 0, len(l.Items))
	for _, g := range l.Items {
		items = append(items, glossaryItem{
			ID:                 pgxutil.UUIDString(g.ID),
			Term:               g.Term,
			DefinitionMarkdown: g.DefinitionMarkdown,
			Language:           string(g.Language),
		})
	}
	return glossaryListResponse{Items: items, Total: l.Total, Limit: l.Limit, Offset: l.Offset}
}

// trackProgressResponse is a user's aggregated progress over a track.
type trackItemProgressResponse struct {
	ContentType string `json:"content_type"`
	ContentID   string `json:"content_id"`
	Position    int32  `json:"position"`
	Completed   bool   `json:"completed"`
}

type trackProgressResponse struct {
	TrackID       string                      `json:"track_id"`
	Total         int                         `json:"total"`
	Completed     int                         `json:"completed"`
	Percent       int                         `json:"percent"`
	TrackComplete bool                        `json:"track_complete"`
	Items         []trackItemProgressResponse `json:"items"`
}

func toTrackProgressResponse(r progress.TrackProgressResult) trackProgressResponse {
	items := make([]trackItemProgressResponse, 0, len(r.Items))
	for _, it := range r.Items {
		items = append(items, trackItemProgressResponse{
			ContentType: it.ContentType,
			ContentID:   it.ContentID,
			Position:    it.Position,
			Completed:   it.Completed,
		})
	}
	return trackProgressResponse{
		TrackID:       r.TrackID,
		Total:         r.Total,
		Completed:     r.Completed,
		Percent:       r.Percent,
		TrackComplete: r.TrackComplete,
		Items:         items,
	}
}

// problemListItem is a problem's metadata for list responses (no statement/solution).
type problemListItem struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug"`
	Difficulty string    `json:"difficulty"`
	Tags       []string  `json:"tags"`
	Language   string    `json:"language"`
	CreatedAt  time.Time `json:"created_at"`
}

type problemListResponse struct {
	Items  []problemListItem `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

func toProblemListResponse(l content.ProblemList) problemListResponse {
	items := make([]problemListItem, 0, len(l.Items))
	for _, p := range l.Items {
		items = append(items, problemListItem{
			ID:         pgxutil.UUIDString(p.ID),
			Title:      p.Title,
			Slug:       p.Slug,
			Difficulty: string(p.Difficulty),
			Tags:       p.Tags,
			Language:   string(p.Language),
			CreatedAt:  p.CreatedAt.Time,
		})
	}
	return problemListResponse{Items: items, Total: l.Total, Limit: l.Limit, Offset: l.Offset}
}

// problemDetailResponse deliberately omits reference_solution_markdown (hidden
// until the user solves the problem) but exposes the sample I/O examples.
type problemDetailResponse struct {
	ID                string          `json:"id"`
	Title             string          `json:"title"`
	Slug              string          `json:"slug"`
	StatementMarkdown string          `json:"statement_markdown"`
	Difficulty        string          `json:"difficulty"`
	Tags              []string        `json:"tags"`
	Language          string          `json:"language"`
	SampleIO          json.RawMessage `json:"sample_io"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func toProblemDetailResponse(p store.Problem) problemDetailResponse {
	sampleIO := json.RawMessage(p.SampleIo)
	if len(sampleIO) == 0 {
		sampleIO = json.RawMessage("[]")
	}
	return problemDetailResponse{
		ID:                pgxutil.UUIDString(p.ID),
		Title:             p.Title,
		Slug:              p.Slug,
		StatementMarkdown: p.StatementMarkdown,
		Difficulty:        string(p.Difficulty),
		Tags:              p.Tags,
		Language:          string(p.Language),
		SampleIO:          sampleIO,
		CreatedAt:         p.CreatedAt.Time,
		UpdatedAt:         p.UpdatedAt.Time,
	}
}

// problemSubmissionResponse is the saved submission. The reference solution is
// included only once the problem is solved (omitted otherwise); verdict is set
// only when the online judge graded the submission.
type problemSubmissionResponse struct {
	ID                        string        `json:"id"`
	Status                    string        `json:"status"`
	Language                  string        `json:"language"`
	CreatedAt                 time.Time     `json:"created_at"`
	Verdict                   *judge.Result `json:"verdict,omitempty"`
	ReferenceSolutionMarkdown string        `json:"reference_solution_markdown,omitempty"`
}

func toProblemSubmissionResponse(r progress.ProblemSubmissionResult) problemSubmissionResponse {
	return problemSubmissionResponse{
		ID:                        pgxutil.UUIDString(r.Submission.ID),
		Status:                    string(r.Submission.Status),
		Language:                  r.Submission.Language,
		CreatedAt:                 r.Submission.CreatedAt.Time,
		ReferenceSolutionMarkdown: r.ReferenceSolution,
	}
}

func toJudgedSubmissionResponse(sub store.ProblemSubmission, verdict judge.Result) problemSubmissionResponse {
	v := verdict
	return problemSubmissionResponse{
		ID:        pgxutil.UUIDString(sub.ID),
		Status:    string(sub.Status),
		Language:  sub.Language,
		CreatedAt: sub.CreatedAt.Time,
		Verdict:   &v,
	}
}

// quizListItem is a quiz's metadata (no questions) for list responses.
type quizListItem struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	PassThreshold int32     `json:"pass_threshold"`
	Difficulty    string    `json:"difficulty"`
	Tags          []string  `json:"tags"`
	Language      string    `json:"language"`
	CreatedAt     time.Time `json:"created_at"`
}

type quizListResponse struct {
	Items  []quizListItem `json:"items"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func toQuizListResponse(l content.QuizList) quizListResponse {
	items := make([]quizListItem, 0, len(l.Items))
	for _, q := range l.Items {
		items = append(items, quizListItem{
			ID:            pgxutil.UUIDString(q.ID),
			Title:         q.Title,
			Description:   q.Description,
			PassThreshold: q.PassThreshold,
			Difficulty:    string(q.Difficulty),
			Tags:          q.Tags,
			Language:      string(q.Language),
			CreatedAt:     q.CreatedAt.Time,
		})
	}
	return quizListResponse{Items: items, Total: l.Total, Limit: l.Limit, Offset: l.Offset}
}

// quizOptionResponse deliberately omits is_correct so students cannot see answers.
type quizOptionResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type quizQuestionResponse struct {
	ID      string               `json:"id"`
	Prompt  string               `json:"prompt"`
	Type    string               `json:"type"`
	Options []quizOptionResponse `json:"options"`
}

type quizDetailResponse struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	PassThreshold int32                  `json:"pass_threshold"`
	Difficulty    string                 `json:"difficulty"`
	Tags          []string               `json:"tags"`
	Language      string                 `json:"language"`
	Questions     []quizQuestionResponse `json:"questions"`
}

func toQuizDetailResponse(d content.QuizDetail) quizDetailResponse {
	questions := make([]quizQuestionResponse, 0, len(d.Questions))
	for _, qd := range d.Questions {
		options := make([]quizOptionResponse, 0, len(qd.Options))
		for _, o := range qd.Options {
			options = append(options, quizOptionResponse{ID: pgxutil.UUIDString(o.ID), Text: o.Text})
		}
		questions = append(questions, quizQuestionResponse{
			ID:      pgxutil.UUIDString(qd.Question.ID),
			Prompt:  qd.Question.Prompt,
			Type:    string(qd.Question.Type),
			Options: options,
		})
	}
	return quizDetailResponse{
		ID:            pgxutil.UUIDString(d.Quiz.ID),
		Title:         d.Quiz.Title,
		Description:   d.Quiz.Description,
		PassThreshold: d.Quiz.PassThreshold,
		Difficulty:    string(d.Quiz.Difficulty),
		Tags:          d.Quiz.Tags,
		Language:      string(d.Quiz.Language),
		Questions:     questions,
	}
}

// quizAttemptResponse is the graded result of a submission. Correct answers are
// revealed here (after the attempt), unlike the quiz read endpoint.
type quizQuestionReview struct {
	QuestionID       string   `json:"question_id"`
	Correct          bool     `json:"correct"`
	CorrectOptionIDs []string `json:"correct_option_ids"`
}

type quizAttemptResponse struct {
	AttemptID string               `json:"attempt_id"`
	Score     int                  `json:"score"`
	Passed    bool                 `json:"passed"`
	CreatedAt time.Time            `json:"created_at"`
	Review    []quizQuestionReview `json:"review"`
}

func toQuizAttemptResponse(res quiz.AttemptResult) quizAttemptResponse {
	review := make([]quizQuestionReview, 0, len(res.Review))
	for _, r := range res.Review {
		review = append(review, quizQuestionReview{
			QuestionID:       r.QuestionID,
			Correct:          r.Correct,
			CorrectOptionIDs: r.CorrectIDs,
		})
	}
	return quizAttemptResponse{
		AttemptID: res.AttemptID,
		Score:     res.Score,
		Passed:    res.Passed,
		CreatedAt: res.CreatedAt,
		Review:    review,
	}
}

// articleReadResponse confirms an article was marked read.
type articleReadResponse struct {
	ArticleID   string    `json:"article_id"`
	CompletedAt time.Time `json:"completed_at"`
}

func toArticleReadResponse(r store.ArticleRead) articleReadResponse {
	return articleReadResponse{
		ArticleID:   pgxutil.UUIDString(r.ArticleID),
		CompletedAt: r.CompletedAt.Time,
	}
}

// articleReadStatusResponse reports whether the user has read an article.
type articleReadStatusResponse struct {
	Read        bool       `json:"read"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func toArticleReadStatusResponse(found bool, r store.ArticleRead) articleReadStatusResponse {
	resp := articleReadStatusResponse{Read: found}
	if found {
		ts := r.CompletedAt.Time
		resp.CompletedAt = &ts
	}
	return resp
}

// statsResponse is the authenticated user's gamification summary (XP, level, streaks).
type statsResponse struct {
	TotalXP        int    `json:"total_xp"`
	Level          int    `json:"level"`
	CurrentStreak  int    `json:"current_streak"`
	LongestStreak  int    `json:"longest_streak"`
	LastActiveDate string `json:"last_active_date"` // "YYYY-MM-DD", "" if never active
}

func toStatsResponse(s gamification.Stats) statsResponse {
	return statsResponse{
		TotalXP:        s.TotalXP,
		Level:          s.Level,
		CurrentStreak:  s.CurrentStreak,
		LongestStreak:  s.LongestStreak,
		LastActiveDate: s.LastActiveDate,
	}
}

// noteResponse is a user's note on a piece of content.
type noteResponse struct {
	ID          string    `json:"id"`
	ContentType string    `json:"content_type"`
	ContentID   string    `json:"content_id"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toNoteResponse(n social.Note) noteResponse {
	return noteResponse{
		ID:          n.ID,
		ContentType: n.ContentType,
		ContentID:   n.ContentID,
		Body:        n.Body,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
}

type notesListResponse struct {
	Notes []noteResponse `json:"notes"`
}

func toNotesListResponse(ns []social.Note) notesListResponse {
	items := make([]noteResponse, 0, len(ns))
	for _, n := range ns {
		items = append(items, toNoteResponse(n))
	}
	return notesListResponse{Notes: items}
}

// toQuizMeta projects a quiz row to its metadata response (admin create/update).
func toQuizMeta(q store.Quiz) quizListItem {
	return quizListItem{
		ID:            pgxutil.UUIDString(q.ID),
		Title:         q.Title,
		Description:   q.Description,
		PassThreshold: q.PassThreshold,
		Difficulty:    string(q.Difficulty),
		Tags:          q.Tags,
		Language:      string(q.Language),
		CreatedAt:     q.CreatedAt.Time,
	}
}

// adminUserResponse is the admin view of a user (includes role/block status, omits
// password_hash).
type adminUserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	Role          string    `json:"role"`
	IsBlocked     bool      `json:"is_blocked"`
	EmailVerified bool      `json:"email_verified"`
	Locale        string    `json:"locale"`
	CreatedAt     time.Time `json:"created_at"`
}

func toAdminUserResponse(u store.User) adminUserResponse {
	return adminUserResponse{
		ID: pgxutil.UUIDString(u.ID), Email: u.Email, DisplayName: u.DisplayName,
		Role: string(u.Role), IsBlocked: u.IsBlocked, EmailVerified: u.EmailVerified,
		Locale: string(u.Locale), CreatedAt: u.CreatedAt.Time,
	}
}

type adminUserListResponse struct {
	Items  []adminUserResponse `json:"items"`
	Total  int64               `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

func toAdminUserListResponse(l admin.UserList) adminUserListResponse {
	items := make([]adminUserResponse, 0, len(l.Items))
	for _, u := range l.Items {
		items = append(items, toAdminUserResponse(u))
	}
	return adminUserListResponse{Items: items, Total: l.Total, Limit: l.Limit, Offset: l.Offset}
}

// toTrackMeta projects a track row to its metadata response (admin create/update).
func toTrackMeta(t store.Track) trackListItem {
	return trackListItem{
		ID: pgxutil.UUIDString(t.ID), Title: t.Title, Description: t.Description,
		Level: string(t.Level), Language: string(t.Language), Position: t.Position, CreatedAt: t.CreatedAt.Time,
	}
}

// toProjectMeta projects a mini-project row to its metadata response.
func toProjectMeta(p store.MiniProject) projectListItem {
	return projectListItem{
		ID: pgxutil.UUIDString(p.ID), Title: p.Title, Difficulty: string(p.Difficulty),
		Tags: p.Tags, Language: string(p.Language), CreatedAt: p.CreatedAt.Time,
	}
}

// toGlossaryItem projects a glossary row to its response.
func toGlossaryItem(g store.GlossaryTerm) glossaryItem {
	return glossaryItem{
		ID: pgxutil.UUIDString(g.ID), Term: g.Term, DefinitionMarkdown: g.DefinitionMarkdown, Language: string(g.Language),
	}
}

// adminBadgeResponse is the admin view of a badge definition.
type adminBadgeResponse struct {
	ID             string          `json:"id"`
	Code           string          `json:"code"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	Icon           string          `json:"icon"`
	CriteriaType   string          `json:"criteria_type"`
	CriteriaParams json.RawMessage `json:"criteria_params"`
	CreatedAt      time.Time       `json:"created_at"`
}

func toAdminBadgeResponse(b store.Badge) adminBadgeResponse {
	cp := json.RawMessage(b.CriteriaParams)
	if len(cp) == 0 {
		cp = json.RawMessage("{}")
	}
	return adminBadgeResponse{
		ID: pgxutil.UUIDString(b.ID), Code: b.Code, Title: b.Title, Description: b.Description,
		Icon: b.Icon, CriteriaType: b.CriteriaType, CriteriaParams: cp, CreatedAt: b.CreatedAt.Time,
	}
}

// adminDailyChallengeResponse is the admin view of a scheduled daily challenge.
type adminDailyChallengeResponse struct {
	ID            string    `json:"id"`
	ChallengeDate string    `json:"challenge_date"`
	ContentType   string    `json:"content_type"`
	ContentID     string    `json:"content_id"`
	BonusXP       int32     `json:"bonus_xp"`
	CreatedAt     time.Time `json:"created_at"`
}

func toAdminDailyChallengeResponse(d store.DailyChallenge) adminDailyChallengeResponse {
	return adminDailyChallengeResponse{
		ID: pgxutil.UUIDString(d.ID), ChallengeDate: d.ChallengeDate.Time.Format("2006-01-02"),
		ContentType: string(d.ContentType), ContentID: pgxutil.UUIDString(d.ContentID),
		BonusXP: d.BonusXp, CreatedAt: d.CreatedAt.Time,
	}
}

// adminProblemResponse is the admin view of a problem (includes the reference
// solution, unlike the public detail response).
type adminProblemResponse struct {
	ID                        string          `json:"id"`
	Title                     string          `json:"title"`
	Slug                      string          `json:"slug"`
	StatementMarkdown         string          `json:"statement_markdown"`
	ReferenceSolutionMarkdown string          `json:"reference_solution_markdown"`
	Difficulty                string          `json:"difficulty"`
	Tags                      []string        `json:"tags"`
	Language                  string          `json:"language"`
	SampleIO                  json.RawMessage `json:"sample_io"`
	CreatedAt                 time.Time       `json:"created_at"`
	UpdatedAt                 time.Time       `json:"updated_at"`
}

func toAdminProblemResponse(p store.Problem) adminProblemResponse {
	sampleIO := json.RawMessage(p.SampleIo)
	if len(sampleIO) == 0 {
		sampleIO = json.RawMessage("[]")
	}
	return adminProblemResponse{
		ID:                        pgxutil.UUIDString(p.ID),
		Title:                     p.Title,
		Slug:                      p.Slug,
		StatementMarkdown:         p.StatementMarkdown,
		ReferenceSolutionMarkdown: p.ReferenceSolutionMarkdown,
		Difficulty:                string(p.Difficulty),
		Tags:                      p.Tags,
		Language:                  string(p.Language),
		SampleIO:                  sampleIO,
		CreatedAt:                 p.CreatedAt.Time,
		UpdatedAt:                 p.UpdatedAt.Time,
	}
}

// certificateResponse is a certificate the user has earned.
type certificateResponse struct {
	Code       string    `json:"code"`
	TrackID    string    `json:"track_id"`
	TrackTitle string    `json:"track_title"`
	IssuedAt   time.Time `json:"issued_at"`
}

func toCertificateResponse(c social.Certificate) certificateResponse {
	return certificateResponse{Code: c.Code, TrackID: c.TrackID, TrackTitle: c.TrackTitle, IssuedAt: c.IssuedAt}
}

type certificatesListResponse struct {
	Certificates []certificateResponse `json:"certificates"`
}

func toCertificatesListResponse(cs []social.Certificate) certificatesListResponse {
	items := make([]certificateResponse, 0, len(cs))
	for _, c := range cs {
		items = append(items, toCertificateResponse(c))
	}
	return certificatesListResponse{Certificates: items}
}

// certificateVerificationResponse is the public verification of a certificate.
type certificateVerificationResponse struct {
	Code        string    `json:"code"`
	DisplayName string    `json:"display_name"`
	TrackTitle  string    `json:"track_title"`
	IssuedAt    time.Time `json:"issued_at"`
}

func toCertificateVerificationResponse(v social.CertificateVerification) certificateVerificationResponse {
	return certificateVerificationResponse{
		Code: v.Code, DisplayName: v.DisplayName, TrackTitle: v.TrackTitle, IssuedAt: v.IssuedAt,
	}
}

// recentCompletionResponse is one finished item in the user's activity feed.
type recentCompletionResponse struct {
	ContentType string    `json:"content_type"`
	ContentID   string    `json:"content_id"`
	Title       string    `json:"title"`
	CompletedAt time.Time `json:"completed_at"`
}

type recentCompletionsListResponse struct {
	Items []recentCompletionResponse `json:"items"`
}

func toRecentCompletionsResponse(rs []content.RecentCompletion) recentCompletionsListResponse {
	items := make([]recentCompletionResponse, 0, len(rs))
	for _, r := range rs {
		items = append(items, recentCompletionResponse{
			ContentType: r.ContentType,
			ContentID:   r.ContentID,
			Title:       r.Title,
			CompletedAt: r.CompletedAt,
		})
	}
	return recentCompletionsListResponse{Items: items}
}

// bookmarkResponse is a user's saved reference to content.
type bookmarkResponse struct {
	ID          string    `json:"id"`
	ContentType string    `json:"content_type"`
	ContentID   string    `json:"content_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func toBookmarkResponse(b social.Bookmark) bookmarkResponse {
	return bookmarkResponse{
		ID:          b.ID,
		ContentType: b.ContentType,
		ContentID:   b.ContentID,
		CreatedAt:   b.CreatedAt,
	}
}

type bookmarksListResponse struct {
	Bookmarks []bookmarkResponse `json:"bookmarks"`
}

func toBookmarksListResponse(bs []social.Bookmark) bookmarksListResponse {
	items := make([]bookmarkResponse, 0, len(bs))
	for _, b := range bs {
		items = append(items, toBookmarkResponse(b))
	}
	return bookmarksListResponse{Bookmarks: items}
}

// leaderboardEntryResponse is one ranked user on the leaderboard.
type leaderboardEntryResponse struct {
	Rank        int    `json:"rank"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	XP          int64  `json:"xp"`
}

type leaderboardResponse struct {
	Period  string                     `json:"period"`
	Entries []leaderboardEntryResponse `json:"entries"`
}

func toLeaderboardResponse(period string, entries []social.LeaderboardEntry) leaderboardResponse {
	if period == "" {
		period = "all"
	}
	items := make([]leaderboardEntryResponse, 0, len(entries))
	for _, e := range entries {
		items = append(items, leaderboardEntryResponse{
			Rank: e.Rank, UserID: e.UserID, DisplayName: e.DisplayName, AvatarURL: e.AvatarURL, XP: e.XP,
		})
	}
	return leaderboardResponse{Period: period, Entries: items}
}

// dailyChallengeResponse is today's challenge plus the user's completion state.
type dailyChallengeResponse struct {
	Date        string `json:"date"`
	ContentType string `json:"content_type"`
	ContentID   string `json:"content_id"`
	BonusXP     int    `json:"bonus_xp"`
	Completed   bool   `json:"completed"`
}

func toDailyChallengeResponse(c gamification.DailyChallenge) dailyChallengeResponse {
	return dailyChallengeResponse{
		Date:        c.Date,
		ContentType: c.ContentType,
		ContentID:   c.ContentID,
		BonusXP:     c.BonusXP,
		Completed:   c.Completed,
	}
}

// dailyCompletionResponse confirms a daily-challenge completion. newly_completed
// is false when it was already done today (no extra reward).
type dailyCompletionResponse struct {
	Challenge      dailyChallengeResponse `json:"challenge"`
	NewlyCompleted bool                   `json:"newly_completed"`
}

func toDailyCompletionResponse(r gamification.DailyCompletion) dailyCompletionResponse {
	return dailyCompletionResponse{
		Challenge:      toDailyChallengeResponse(r.Challenge),
		NewlyCompleted: r.NewlyCompleted,
	}
}

// badgeResponse is one earned achievement.
type badgeResponse struct {
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	AwardedAt   time.Time `json:"awarded_at"`
}

type badgesResponse struct {
	Badges []badgeResponse `json:"badges"`
}

func toBadgesResponse(bs []gamification.Badge) badgesResponse {
	items := make([]badgeResponse, 0, len(bs))
	for _, b := range bs {
		items = append(items, badgeResponse{
			Code:        b.Code,
			Title:       b.Title,
			Description: b.Description,
			Icon:        b.Icon,
			AwardedAt:   b.AwardedAt,
		})
	}
	return badgesResponse{Badges: items}
}

// progressSummaryResponse is the authenticated user's per-section completion counts.
type progressSummaryResponse struct {
	VideosCompleted   int64 `json:"videos_completed"`
	ArticlesRead      int64 `json:"articles_read"`
	QuizzesPassed     int64 `json:"quizzes_passed"`
	ProblemsSolved    int64 `json:"problems_solved"`
	ProjectsCompleted int64 `json:"projects_completed"`
}

func toProgressSummaryResponse(s progress.ProgressSummaryResult) progressSummaryResponse {
	return progressSummaryResponse{
		VideosCompleted:   s.VideosCompleted,
		ArticlesRead:      s.ArticlesRead,
		QuizzesPassed:     s.QuizzesPassed,
		ProblemsSolved:    s.ProblemsSolved,
		ProjectsCompleted: s.ProjectsCompleted,
	}
}

// activityDayResponse is one heatmap bucket (a UTC day).
type activityDayResponse struct {
	Day   string `json:"day"`
	Count int64  `json:"count"`
	XP    int64  `json:"xp"`
}

// activityResponse echoes the resolved inclusive date window and the daily buckets.
type activityResponse struct {
	From string                `json:"from"`
	To   string                `json:"to"`
	Days []activityDayResponse `json:"days"`
}

func toActivityResponse(from, toExclusive time.Time, days []progress.ActivityDay) activityResponse {
	items := make([]activityDayResponse, 0, len(days))
	for _, d := range days {
		items = append(items, activityDayResponse{Day: d.Day, Count: d.Count, XP: d.XP})
	}
	return activityResponse{
		From: from.Format("2006-01-02"),
		To:   toExclusive.AddDate(0, 0, -1).Format("2006-01-02"), // back to inclusive last day
		Days: items,
	}
}

// videoProgressResponse is a user's progress for a single video.
type videoProgressResponse struct {
	VideoID             string    `json:"video_id"`
	WatchedPercent      int32     `json:"watched_percent"`
	LastPositionSeconds int32     `json:"last_position_seconds"`
	Completed           bool      `json:"completed"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func toVideoProgressResponse(p store.VideoProgress) videoProgressResponse {
	return videoProgressResponse{
		VideoID:             pgxutil.UUIDString(p.VideoID),
		WatchedPercent:      p.WatchedPercent,
		LastPositionSeconds: p.LastPositionSeconds,
		Completed:           p.Completed,
		UpdatedAt:           p.UpdatedAt.Time,
	}
}
