// Package logging builds the application's structured slog.Logger from config.
package logging

import (
	"log/slog"
	"os"
)

// New returns a *slog.Logger writing to stdout. The level is one of
// debug|info|warn|error (unknown values fall back to info) and the format is
// text|json (anything other than "text" yields JSON).
func New(level, format string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var handler slog.Handler
	if format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
