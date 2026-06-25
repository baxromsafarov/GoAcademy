package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/goacademy/backend/internal/gamification"
	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
)

// dailyHandler serves the daily challenge (/api/v1/daily-challenge).
type dailyHandler struct {
	svc    *gamification.DailyService
	logger *slog.Logger
	now    func() time.Time // injectable for tests
}

func newDailyHandler(svc *gamification.DailyService, logger *slog.Logger) *dailyHandler {
	return &dailyHandler{svc: svc, logger: logger, now: time.Now}
}

// utcDay truncates the current time to its UTC calendar day (D-010).
func utcDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// get handles GET /api/v1/daily-challenge.
func (h *dailyHandler) get(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	ch, err := h.svc.Today(r.Context(), uid, utcDay(h.now()))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toDailyChallengeResponse(ch))
}

// complete handles POST /api/v1/daily-challenge/complete.
func (h *dailyHandler) complete(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	res, err := h.svc.Complete(r.Context(), uid, utcDay(h.now()))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toDailyCompletionResponse(res))
}
