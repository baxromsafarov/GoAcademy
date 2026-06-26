package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goacademy/backend/internal/content"
	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/quiz"
)

// quizHandler serves quiz read endpoints (answers hidden) and the authenticated
// attempt submission (answers revealed in the result).
type quizHandler struct {
	svc    *content.Service
	quiz   *quiz.Service
	logger *slog.Logger
}

func newQuizHandler(svc *content.Service, quizSvc *quiz.Service, logger *slog.Logger) *quizHandler {
	return &quizHandler{svc: svc, quiz: quizSvc, logger: logger}
}

// list handles GET /api/v1/quizzes?difficulty=&tag=&language=&limit=&offset=.
func (h *quizHandler) list(w http.ResponseWriter, r *http.Request) {
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

	list, err := h.svc.ListQuizzes(r.Context(), filter)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toQuizListResponse(list))
}

// get handles GET /api/v1/quizzes/{id} (questions and options, without answers).
func (h *quizHandler) get(w http.ResponseWriter, r *http.Request) {
	detail, err := h.svc.GetQuizDetail(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toQuizDetailResponse(detail))
}

type submitAttemptRequest struct {
	Answers map[string][]string `json:"answers"`
}

// submit handles POST /api/v1/quizzes/{id}/attempts (authenticated).
func (h *quizHandler) submit(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}

	var req submitAttemptRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}

	res, err := h.quiz.Submit(r.Context(), uid, chi.URLParam(r, "id"), quiz.SubmitInput{Answers: req.Answers})
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toQuizAttemptResponse(res))
}
