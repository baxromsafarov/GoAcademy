package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/goacademy/backend/internal/httpapi/respond"
)

// readyCheckTimeout bounds each readiness probe so a hung dependency cannot hang
// the /readyz request itself.
const readyCheckTimeout = 2 * time.Second

// Check is a named readiness probe for a dependency (e.g. the database).
type Check struct {
	Name string
	Func func(ctx context.Context) error
}

// healthz is a liveness probe: it returns 200 as long as the process serves HTTP.
func healthz(w http.ResponseWriter, _ *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readyz runs every registered check. It returns 200 only when all pass,
// otherwise 503 with a per-check report.
func readyz(checks []Check) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := make(map[string]string, len(checks))
		ready := true
		for _, c := range checks {
			ctx, cancel := context.WithTimeout(r.Context(), readyCheckTimeout)
			err := c.Func(ctx)
			cancel()
			if err != nil {
				ready = false
				results[c.Name] = err.Error()
				continue
			}
			results[c.Name] = "ok"
		}

		status := http.StatusOK
		state := "ready"
		if !ready {
			status = http.StatusServiceUnavailable
			state = "not_ready"
		}
		respond.JSON(w, status, map[string]any{"status": state, "checks": results})
	}
}
