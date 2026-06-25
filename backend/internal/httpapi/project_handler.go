package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goacademy/backend/internal/content"
	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/progress"
)

// projectHandler serves mini-project read endpoints and the authenticated
// checklist progress (read + step toggle).
type projectHandler struct {
	svc      *content.Service
	progress *progress.Service
	logger   *slog.Logger
}

func newProjectHandler(svc *content.Service, prog *progress.Service, logger *slog.Logger) *projectHandler {
	return &projectHandler{svc: svc, progress: prog, logger: logger}
}

// list handles GET /api/v1/projects?difficulty=&tag=&language=&limit=&offset=.
func (h *projectHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.svc.ListProjects(r.Context(), content.ListFilter{
		Difficulty: optionalQuery(q, "difficulty"),
		Tag:        optionalQuery(q, "tag"),
		Language:   optionalQuery(q, "language"),
		Limit:      queryInt(q, "limit", 0),
		Offset:     queryInt(q, "offset", 0),
	})
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toProjectListResponse(list))
}

// get handles GET /api/v1/projects/{id} (project + ordered checklist).
func (h *projectHandler) get(w http.ResponseWriter, r *http.Request) {
	detail, err := h.svc.GetProjectDetail(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toProjectDetailResponse(detail))
}

// getProgress handles GET /api/v1/projects/{id}/progress (authenticated).
func (h *projectHandler) getProgress(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	res, err := h.progress.ProjectProgress(r.Context(), uid, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toProjectProgressResponse(res))
}

// toggleStep handles POST /api/v1/projects/{id}/steps/{stepId}/toggle (authenticated).
func (h *projectHandler) toggleStep(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	res, err := h.progress.ToggleProjectStep(r.Context(), uid, chi.URLParam(r, "id"), chi.URLParam(r, "stepId"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toProjectProgressResponse(res))
}
