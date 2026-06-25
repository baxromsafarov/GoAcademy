package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/progress"
)

// meProgressHandler serves the authenticated user's progress summary and activity
// heatmap (/api/v1/me/progress, /api/v1/me/activity).
type meProgressHandler struct {
	progress *progress.Service
	logger   *slog.Logger
	now      func() time.Time // injectable for tests; defaults to time.Now
}

func newMeProgressHandler(prog *progress.Service, logger *slog.Logger) *meProgressHandler {
	return &meProgressHandler{progress: prog, logger: logger, now: time.Now}
}

// summary handles GET /api/v1/me/progress.
func (h *meProgressHandler) summary(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	s, err := h.progress.ProgressSummary(r.Context(), uid)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toProgressSummaryResponse(s))
}

// activity handles GET /api/v1/me/activity?from=&to= (heatmap, daily UTC buckets).
func (h *meProgressHandler) activity(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}

	q := r.URL.Query()
	from, toExclusive, err := progress.ParseHeatmapRange(q.Get("from"), q.Get("to"), h.now())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}

	days, err := h.progress.ActivityHeatmap(r.Context(), uid, from, toExclusive)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toActivityResponse(from, toExclusive, days))
}
