package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goacademy/backend/internal/content"
	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/judge"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/progress"
)

// problemHandler serves problem read endpoints (solution hidden) plus the
// authenticated submission and solution-reveal endpoints.
type problemHandler struct {
	svc      *content.Service
	progress *progress.Service
	judge    *judge.Service // nil when the online judge is disabled
	logger   *slog.Logger
}

func newProblemHandler(svc *content.Service, prog *progress.Service, jdg *judge.Service, logger *slog.Logger) *problemHandler {
	return &problemHandler{svc: svc, progress: prog, judge: jdg, logger: logger}
}

// list handles GET /api/v1/problems?difficulty=&tag=&language=&limit=&offset=.
func (h *problemHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := content.ListFilter{
		Difficulty: optionalQuery(q, "difficulty"),
		Tag:        optionalQuery(q, "tag"),
		Language:      optionalQuery(q, "language"),
		Q:             optionalQuery(q, "q"),
		IncludeHidden: adminWantsHidden(r),
		Limit:         queryInt(q, "limit", 0),
		Offset:        queryInt(q, "offset", 0),
	}

	list, err := h.svc.ListProblems(r.Context(), filter)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toProblemListResponse(list))
}

// get handles GET /api/v1/problems/{slug} (statement + sample I/O, no solution).
func (h *problemHandler) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetProblemBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toProblemDetailResponse(p))
}

type submitProblemRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Solved   bool   `json:"solved"`
}

// submit handles POST /api/v1/problems/{slug}/submissions (authenticated). When
// the online judge is enabled and the problem has test cases, the submission is
// auto-graded (verdict + auto-solved); otherwise it falls back to manual marking.
func (h *problemHandler) submit(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}

	var req submitProblemRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	slug := chi.URLParam(r, "slug")

	if h.judge != nil {
		// Resolve the problem (also yields its reference solution to reveal on OK).
		problem, err := h.svc.GetProblemBySlug(r.Context(), slug)
		if err != nil {
			respond.Error(w, r, h.logger, err)
			return
		}
		sub, verdict, jerr := h.judge.Submit(r.Context(), uid, pgxutil.UUIDString(problem.ID), req.Code, req.Language)
		switch {
		case jerr == nil:
			resp := toJudgedSubmissionResponse(sub, verdict)
			if sub.Status == "solved" {
				resp.ReferenceSolutionMarkdown = problem.ReferenceSolutionMarkdown
			}
			respond.JSON(w, http.StatusCreated, resp)
			return
		case errors.Is(jerr, judge.ErrNoTestCases):
			// No test set — fall through to the manual flow below.
		default:
			respond.Error(w, r, h.logger, jerr)
			return
		}
	}

	res, err := h.progress.SubmitProblem(r.Context(), uid, slug, progress.SubmitProblemInput{
		Code:     req.Code,
		Language: req.Language,
		Solved:   req.Solved,
	})
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toProblemSubmissionResponse(res))
}

// getSolution handles GET /api/v1/problems/{slug}/solution (authenticated): it
// returns the reference solution only if the user has solved the problem.
func (h *problemHandler) getSolution(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}

	sol, err := h.progress.GetProblemSolution(r.Context(), uid, chi.URLParam(r, "slug"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"reference_solution_markdown": sol})
}
