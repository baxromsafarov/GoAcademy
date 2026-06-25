package httpapi

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/runner"
)

// sandboxHandler runs untrusted Go in the code-runner and returns its output.
type sandboxHandler struct {
	runner *runner.Runner
	logger *slog.Logger
}

func newSandboxHandler(r *runner.Runner, logger *slog.Logger) *sandboxHandler {
	return &sandboxHandler{runner: r, logger: logger}
}

type sandboxRunRequest struct {
	Source string `json:"source"`
	Stdin  string `json:"stdin"`
}

type sandboxRunResponse struct {
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	ExitCode        int    `json:"exit_code"`
	CompileError    bool   `json:"compile_error"`
	TimedOut        bool   `json:"timed_out"`
	OOMKilled       bool   `json:"oom_killed"`
	StdoutTruncated bool   `json:"stdout_truncated"`
	StderrTruncated bool   `json:"stderr_truncated"`
	DurationMs      int64  `json:"duration_ms"`
}

// run handles POST /api/v1/sandbox/run (authenticated, rate-limited).
func (h *sandboxHandler) run(w http.ResponseWriter, r *http.Request) {
	var req sandboxRunRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	if strings.TrimSpace(req.Source) == "" {
		respond.Error(w, r, h.logger, apierr.Validation("source is required"))
		return
	}

	res, err := h.runner.Run(r.Context(), runner.Request{Source: req.Source, Stdin: req.Stdin})
	if err != nil {
		// An infrastructure failure (docker/host) — log the cause, hide it from
		// the client behind a generic 500.
		respond.Error(w, r, h.logger, apierr.Internal().WithCause(err))
		return
	}

	respond.JSON(w, http.StatusOK, sandboxRunResponse{
		Stdout:          res.Stdout,
		Stderr:          res.Stderr,
		ExitCode:        res.ExitCode,
		CompileError:    res.CompileError,
		TimedOut:        res.TimedOut,
		OOMKilled:       res.OOMKilled,
		StdoutTruncated: res.StdoutTruncated,
		StderrTruncated: res.StderrTruncated,
		DurationMs:      res.Duration.Milliseconds(),
	})
}
