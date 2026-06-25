package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/social"
)

// bookmarksHandler serves a user's bookmarks.
type bookmarksHandler struct {
	svc    *social.BookmarksService
	logger *slog.Logger
}

func newBookmarksHandler(svc *social.BookmarksService, logger *slog.Logger) *bookmarksHandler {
	return &bookmarksHandler{svc: svc, logger: logger}
}

type createBookmarkRequest struct {
	ContentType string `json:"content_type"`
	ContentID   string `json:"content_id"`
}

// create handles POST /api/v1/bookmarks.
func (h *bookmarksHandler) create(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	var req createBookmarkRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	b, err := h.svc.Add(r.Context(), uid, req.ContentType, req.ContentID)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toBookmarkResponse(b))
}

// delete handles DELETE /api/v1/bookmarks/{id}.
func (h *bookmarksHandler) delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	if err := h.svc.Remove(r.Context(), uid, chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// list handles GET /api/v1/me/bookmarks.
func (h *bookmarksHandler) list(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	bs, err := h.svc.List(r.Context(), uid)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toBookmarksListResponse(bs))
}
