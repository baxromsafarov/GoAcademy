package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goacademy/backend/internal/admin"
	"github.com/goacademy/backend/internal/httpapi/respond"
)

// adminHandler serves admin-only content CRUD (mounted behind RequireRole admin).
type adminHandler struct {
	svc    *admin.Service
	logger *slog.Logger
}

func newAdminHandler(svc *admin.Service, logger *slog.Logger) *adminHandler {
	return &adminHandler{svc: svc, logger: logger}
}

// videos ----------------------------------------------------------------------

type adminVideoRequest struct {
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	YoutubeID       string   `json:"youtube_id"`
	DurationSeconds int      `json:"duration_seconds"`
	Difficulty      string   `json:"difficulty"`
	Language        string   `json:"language"`
	Tags            []string `json:"tags"`
}

func (r adminVideoRequest) toInput() admin.VideoInput {
	return admin.VideoInput{
		Title: r.Title, Description: r.Description, YoutubeID: r.YoutubeID,
		DurationSeconds: r.DurationSeconds, Difficulty: r.Difficulty, Language: r.Language, Tags: r.Tags,
	}
}

func (h *adminHandler) createVideo(w http.ResponseWriter, r *http.Request) {
	var req adminVideoRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	v, err := h.svc.CreateVideo(r.Context(), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toVideoResponse(v))
}

func (h *adminHandler) updateVideo(w http.ResponseWriter, r *http.Request) {
	var req adminVideoRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	v, err := h.svc.UpdateVideo(r.Context(), chi.URLParam(r, "id"), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toVideoResponse(v))
}

func (h *adminHandler) deleteVideo(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteVideo(r.Context(), chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// articles --------------------------------------------------------------------

type adminArticleRequest struct {
	Title        string   `json:"title"`
	Slug         string   `json:"slug"`
	BodyMarkdown string   `json:"body_markdown"`
	Difficulty   string   `json:"difficulty"`
	Language     string   `json:"language"`
	Tags         []string `json:"tags"`
}

func (r adminArticleRequest) toInput() admin.ArticleInput {
	return admin.ArticleInput{
		Title: r.Title, Slug: r.Slug, BodyMarkdown: r.BodyMarkdown,
		Difficulty: r.Difficulty, Language: r.Language, Tags: r.Tags,
	}
}

func (h *adminHandler) createArticle(w http.ResponseWriter, r *http.Request) {
	var req adminArticleRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	a, err := h.svc.CreateArticle(r.Context(), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toArticleResponse(a))
}

func (h *adminHandler) updateArticle(w http.ResponseWriter, r *http.Request) {
	var req adminArticleRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	a, err := h.svc.UpdateArticle(r.Context(), chi.URLParam(r, "id"), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toArticleResponse(a))
}

func (h *adminHandler) deleteArticle(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteArticle(r.Context(), chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// quizzes ---------------------------------------------------------------------

type adminQuizOptionRequest struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
}

type adminQuizQuestionRequest struct {
	Prompt  string                   `json:"prompt"`
	Type    string                   `json:"type"`
	Options []adminQuizOptionRequest `json:"options"`
}

type adminQuizRequest struct {
	Title         string                     `json:"title"`
	Description   string                     `json:"description"`
	PassThreshold int                        `json:"pass_threshold"`
	Difficulty    string                     `json:"difficulty"`
	Language      string                     `json:"language"`
	Tags          []string                   `json:"tags"`
	Questions     []adminQuizQuestionRequest `json:"questions"`
}

func (r adminQuizRequest) toInput() admin.QuizInput {
	questions := make([]admin.QuizQuestionInput, 0, len(r.Questions))
	for _, q := range r.Questions {
		options := make([]admin.QuizOptionInput, 0, len(q.Options))
		for _, o := range q.Options {
			options = append(options, admin.QuizOptionInput{Text: o.Text, IsCorrect: o.IsCorrect})
		}
		questions = append(questions, admin.QuizQuestionInput{Prompt: q.Prompt, Type: q.Type, Options: options})
	}
	return admin.QuizInput{
		Title: r.Title, Description: r.Description, PassThreshold: r.PassThreshold,
		Difficulty: r.Difficulty, Language: r.Language, Tags: r.Tags, Questions: questions,
	}
}

func (h *adminHandler) createQuiz(w http.ResponseWriter, r *http.Request) {
	var req adminQuizRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	q, err := h.svc.CreateQuiz(r.Context(), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toQuizMeta(q))
}

func (h *adminHandler) updateQuiz(w http.ResponseWriter, r *http.Request) {
	var req adminQuizRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	q, err := h.svc.UpdateQuiz(r.Context(), chi.URLParam(r, "id"), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toQuizMeta(q))
}

func (h *adminHandler) deleteQuiz(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteQuiz(r.Context(), chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// problems --------------------------------------------------------------------

type adminTestCaseRequest struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	IsSample       bool   `json:"is_sample"`
}

type adminProblemRequest struct {
	Title                     string                 `json:"title"`
	Slug                      string                 `json:"slug"`
	StatementMarkdown         string                 `json:"statement_markdown"`
	ReferenceSolutionMarkdown string                 `json:"reference_solution_markdown"`
	Difficulty                string                 `json:"difficulty"`
	Language                  string                 `json:"language"`
	Tags                      []string               `json:"tags"`
	SampleIO                  json.RawMessage        `json:"sample_io"`
	TestCases                 []adminTestCaseRequest `json:"test_cases"`
}

func (r adminProblemRequest) toInput() admin.ProblemInput {
	cases := make([]admin.TestCaseInput, 0, len(r.TestCases))
	for _, tc := range r.TestCases {
		cases = append(cases, admin.TestCaseInput{Input: tc.Input, ExpectedOutput: tc.ExpectedOutput, IsSample: tc.IsSample})
	}
	return admin.ProblemInput{
		Title: r.Title, Slug: r.Slug, StatementMarkdown: r.StatementMarkdown,
		ReferenceSolutionMarkdown: r.ReferenceSolutionMarkdown, Difficulty: r.Difficulty,
		Language: r.Language, Tags: r.Tags, SampleIO: r.SampleIO, TestCases: cases,
	}
}

func (h *adminHandler) createProblem(w http.ResponseWriter, r *http.Request) {
	var req adminProblemRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	p, err := h.svc.CreateProblem(r.Context(), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toAdminProblemResponse(p))
}

func (h *adminHandler) updateProblem(w http.ResponseWriter, r *http.Request) {
	var req adminProblemRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	p, err := h.svc.UpdateProblem(r.Context(), chi.URLParam(r, "id"), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toAdminProblemResponse(p))
}

func (h *adminHandler) deleteProblem(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteProblem(r.Context(), chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
