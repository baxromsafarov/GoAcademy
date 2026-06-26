package httpapi

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/goacademy/backend/internal/content"
	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/progress"
)

// videoHandler serves the video read endpoints and the authenticated progress endpoint.
type videoHandler struct {
	svc      *content.Service
	progress *progress.Service
	logger   *slog.Logger
}

func newVideoHandler(svc *content.Service, prog *progress.Service, logger *slog.Logger) *videoHandler {
	return &videoHandler{svc: svc, progress: prog, logger: logger}
}

// list handles GET /api/v1/videos?difficulty=&tag=&language=&q=&limit=&offset=.
func (h *videoHandler) list(w http.ResponseWriter, r *http.Request) {
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

	list, err := h.svc.ListVideos(r.Context(), filter)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toVideoListResponse(list))
}

// get handles GET /api/v1/videos/{id}.
func (h *videoHandler) get(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.GetVideoByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toVideoResponse(v))
}

type videoProgressRequest struct {
	Percent   int   `json:"percent"`
	Position  int   `json:"position"`
	Completed *bool `json:"completed"`
}

// postProgress handles POST /api/v1/videos/{id}/progress (authenticated).
func (h *videoHandler) postProgress(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}

	var req videoProgressRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}

	p, err := h.progress.RecordVideoProgress(r.Context(), uid, chi.URLParam(r, "id"), progress.VideoProgressInput{
		Percent:   req.Percent,
		Position:  req.Position,
		Completed: req.Completed,
	})
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toVideoProgressResponse(p))
}

// getProgress handles GET /api/v1/videos/{id}/progress (authenticated).
func (h *videoHandler) getProgress(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	p, err := h.progress.GetVideoProgress(r.Context(), uid, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toVideoProgressResponse(p))
}

// optionalQuery returns a pointer to the trimmed query value, or nil if absent/empty.
func optionalQuery(q url.Values, key string) *string {
	if v := q.Get(key); v != "" {
		return &v
	}
	return nil
}

// queryInt parses an int query value, returning fallback when absent or invalid.
func queryInt(q url.Values, key string, fallback int) int {
	if v := q.Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// adminWantsHidden reports whether the caller is an authenticated admin asking
// to include hidden content (?show_hidden=true). Content lists use it so admins
// can manage hidden items while students never see them.
func adminWantsHidden(r *http.Request) bool {
	role, ok := RoleFromContext(r.Context())
	return ok && role == "admin" && r.URL.Query().Get("show_hidden") == "true"
}
