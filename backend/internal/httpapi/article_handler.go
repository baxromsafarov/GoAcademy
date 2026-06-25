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

// articleHandler serves article read endpoints and the authenticated "mark read".
type articleHandler struct {
	svc      *content.Service
	progress *progress.Service
	logger   *slog.Logger
}

func newArticleHandler(svc *content.Service, prog *progress.Service, logger *slog.Logger) *articleHandler {
	return &articleHandler{svc: svc, progress: prog, logger: logger}
}

// list handles GET /api/v1/articles?difficulty=&tag=&language=&limit=&offset=.
func (h *articleHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := content.ListFilter{
		Difficulty: optionalQuery(q, "difficulty"),
		Tag:        optionalQuery(q, "tag"),
		Language:   optionalQuery(q, "language"),
		Limit:      queryInt(q, "limit", 0),
		Offset:     queryInt(q, "offset", 0),
	}

	list, err := h.svc.ListArticles(r.Context(), filter)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toArticleListResponse(list))
}

// get handles GET /api/v1/articles/{slug}.
func (h *articleHandler) get(w http.ResponseWriter, r *http.Request) {
	a, err := h.svc.GetArticleBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toArticleResponse(a))
}

// complete handles POST /api/v1/articles/{slug}/complete (authenticated).
func (h *articleHandler) complete(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	read, err := h.progress.MarkArticleRead(r.Context(), uid, chi.URLParam(r, "slug"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toArticleReadResponse(read))
}

// readStatus handles GET /api/v1/articles/{slug}/read (authenticated).
func (h *articleHandler) readStatus(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	read, found, err := h.progress.GetArticleReadStatus(r.Context(), uid, chi.URLParam(r, "slug"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toArticleReadStatusResponse(found, read))
}
