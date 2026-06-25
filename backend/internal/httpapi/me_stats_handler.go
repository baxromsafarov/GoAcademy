package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/goacademy/backend/internal/gamification"
	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
)

// meStatsHandler serves the authenticated user's gamification stats (/api/v1/me/stats).
type meStatsHandler struct {
	svc    *gamification.Service
	logger *slog.Logger
}

func newMeStatsHandler(svc *gamification.Service, logger *slog.Logger) *meStatsHandler {
	return &meStatsHandler{svc: svc, logger: logger}
}

// get handles GET /api/v1/me/stats.
func (h *meStatsHandler) get(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	st, err := h.svc.GetStats(r.Context(), uid)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toStatsResponse(st))
}

// badges handles GET /api/v1/me/badges.
func (h *meStatsHandler) badges(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	bs, err := h.svc.GetBadges(r.Context(), uid)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toBadgesResponse(bs))
}
