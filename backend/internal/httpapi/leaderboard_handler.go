package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/social"
)

// leaderboardHandler serves the public leaderboard (/api/v1/leaderboard).
type leaderboardHandler struct {
	svc    *social.Service
	logger *slog.Logger
	now    func() time.Time // injectable for tests
}

func newLeaderboardHandler(svc *social.Service, logger *slog.Logger) *leaderboardHandler {
	return &leaderboardHandler{svc: svc, logger: logger, now: time.Now}
}

// get handles GET /api/v1/leaderboard?period=&limit=&offset=.
func (h *leaderboardHandler) get(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	period := q.Get("period")
	entries, err := h.svc.Leaderboard(r.Context(), period, h.now(), queryInt(q, "limit", 0), queryInt(q, "offset", 0))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toLeaderboardResponse(period, entries))
}
