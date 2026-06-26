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

// trackHandler serves learning-track read endpoints and the authenticated
// per-user track progress.
type trackHandler struct {
	svc      *content.Service
	progress *progress.Service
	logger   *slog.Logger
}

func newTrackHandler(svc *content.Service, prog *progress.Service, logger *slog.Logger) *trackHandler {
	return &trackHandler{svc: svc, progress: prog, logger: logger}
}

// list handles GET /api/v1/tracks?difficulty=&language=&limit=&offset=
// (difficulty filters the track level).
func (h *trackHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := content.ListFilter{
		Difficulty: optionalQuery(q, "difficulty"),
		Language:   optionalQuery(q, "language"),
		Q:          optionalQuery(q, "q"),
		Limit:      queryInt(q, "limit", 0),
		Offset:     queryInt(q, "offset", 0),
	}

	list, err := h.svc.ListTracks(r.Context(), filter)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toTrackListResponse(list))
}

// get handles GET /api/v1/tracks/{id} (ordered program of content references).
func (h *trackHandler) get(w http.ResponseWriter, r *http.Request) {
	detail, err := h.svc.GetTrackDetail(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toTrackDetailResponse(detail))
}

// progress handles GET /api/v1/tracks/{id}/progress (authenticated).
func (h *trackHandler) getProgress(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	res, err := h.progress.TrackProgress(r.Context(), uid, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toTrackProgressResponse(res))
}

// enroll handles POST /api/v1/tracks/{id}/enroll.
func (h *trackHandler) enroll(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	if err := h.svc.EnrollTrack(r.Context(), uid, chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// unenroll handles DELETE /api/v1/tracks/{id}/enroll.
func (h *trackHandler) unenroll(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	if err := h.svc.UnenrollTrack(r.Context(), uid, chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// myTracks handles GET /api/v1/me/tracks — the tracks the user enrolled in.
func (h *trackHandler) myTracks(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	tracks, err := h.svc.ListEnrolledTracks(r.Context(), uid)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toTrackListResponse(content.TrackList{
		Items: tracks, Total: int64(len(tracks)), Limit: len(tracks),
	}))
}

// recentCompletions handles GET /api/v1/me/recent — the user's most recently
// finished content (videos, articles, quizzes, problems).
func (h *trackHandler) recentCompletions(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	items, err := h.svc.ListRecentCompletions(r.Context(), uid, queryInt(r.URL.Query(), "limit", 8))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toRecentCompletionsResponse(items))
}
