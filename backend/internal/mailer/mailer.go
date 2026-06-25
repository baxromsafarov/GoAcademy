// Package mailer abstracts outbound email. On startup a logging stub is used
// (decision D-007): it records what would be sent instead of contacting an SMTP
// server, so the rest of the code is written against the interface and a real
// transport can be added later without changing callers.
package mailer

import (
	"context"
	"log/slog"
)

// Mailer sends transactional emails.
type Mailer interface {
	SendEmailVerification(ctx context.Context, to, token string) error
	SendPasswordReset(ctx context.Context, to, token string) error
}

// LogMailer is the development stub. It logs the recipient and token via slog
// instead of sending mail.
//
// WARNING: it logs the verification token in cleartext. This is intentional for
// local development only — never enable LogMailer in production (use real SMTP).
type LogMailer struct {
	logger *slog.Logger
}

// NewLogMailer returns a LogMailer writing through logger.
func NewLogMailer(logger *slog.Logger) *LogMailer {
	return &LogMailer{logger: logger}
}

func (m *LogMailer) SendEmailVerification(ctx context.Context, to, token string) error {
	m.logger.InfoContext(ctx, "stub mailer: email verification (not actually sent)",
		"to", to,
		"token", token,
	)
	return nil
}

func (m *LogMailer) SendPasswordReset(ctx context.Context, to, token string) error {
	m.logger.InfoContext(ctx, "stub mailer: password reset (not actually sent)",
		"to", to,
		"token", token,
	)
	return nil
}
