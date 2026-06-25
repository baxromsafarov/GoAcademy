// Package respond renders HTTP responses in the GoAcademy unified JSON format.
// Success payloads are written with JSON; errors with Error, which guarantees the
// shape {"error":{"code","message","details"}} and never leaks internal causes.
package respond

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/goacademy/backend/internal/platform/apierr"
)

type envelope struct {
	Error payload `json:"error"`
}

type payload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ErrorEnvelope builds (but does not write) the unified error body. Useful where a
// ResponseWriter is available but the *http.Request is not (e.g. middleware).
func ErrorEnvelope(code, message string, details any) any {
	return envelope{Error: payload{Code: code, Message: message, Details: details}}
}

// JSON writes v as a JSON response with the given status code. A nil v writes only
// the status (no body).
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// Error renders err in the unified error format. An *apierr.APIError is rendered
// with its status, code, message and details; any other error becomes a 500 whose
// cause is logged (when logger is non-nil) and never exposed to the client.
func Error(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	var apiErr *apierr.APIError
	if !errors.As(err, &apiErr) {
		apiErr = apierr.Internal().WithCause(err)
	}

	if apiErr.Status >= http.StatusInternalServerError && logger != nil {
		logger.LogAttrs(r.Context(), slog.LevelError, "request error",
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("code", apiErr.Code),
			slog.String("error", err.Error()),
		)
	}

	JSON(w, apiErr.Status, ErrorEnvelope(apiErr.Code, apiErr.Message, apiErr.Details))
}
