package social

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/progress"
	"github.com/goacademy/backend/internal/store"
)

// Certificate is a user's earned track-completion certificate.
type Certificate struct {
	Code       string
	TrackID    string
	TrackTitle string
	IssuedAt   time.Time
}

// CertificateVerification is the public view of a certificate (proof of who
// completed which track, and when).
type CertificateVerification struct {
	Code        string
	DisplayName string
	TrackTitle  string
	IssuedAt    time.Time
}

// CertificatesService issues and verifies track-completion certificates.
type CertificatesService struct {
	queries  *store.Queries
	progress *progress.Service
}

// NewCertificatesService wires the certificates service; it uses the progress
// service to confirm a track is 100% complete before issuing.
func NewCertificatesService(pool *pgxpool.Pool, prog *progress.Service) *CertificatesService {
	return &CertificatesService{queries: store.New(pool), progress: prog}
}

// IssueForTrack issues the user's certificate for a track, but only once the
// track is fully complete. It is idempotent: an already-certified track returns
// the existing certificate.
func (s *CertificatesService) IssueForTrack(ctx context.Context, userID, trackID string) (Certificate, error) {
	prog, err := s.progress.TrackProgress(ctx, userID, trackID)
	if err != nil {
		return Certificate{}, err // not-found for an unknown track, etc.
	}
	if !prog.TrackComplete {
		return Certificate{}, apierr.Conflict("track is not yet completed")
	}

	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return Certificate{}, apierr.Unauthorized("invalid user")
	}
	tid, err := pgxutil.ParseUUID(trackID)
	if err != nil {
		return Certificate{}, apierr.NotFound("track not found")
	}

	code, err := newCertificateCode()
	if err != nil {
		return Certificate{}, err
	}
	row, err := s.queries.IssueCertificate(ctx, store.IssueCertificateParams{
		UserID: uid, TrackID: tid, CertificateCode: code,
	})
	switch {
	case err == nil:
		return Certificate{Code: row.CertificateCode, TrackID: trackID, IssuedAt: row.IssuedAt.Time}, nil
	case errors.Is(err, pgx.ErrNoRows):
		existing, err := s.queries.GetUserTrackCertificate(ctx, store.GetUserTrackCertificateParams{UserID: uid, TrackID: tid})
		if err != nil {
			return Certificate{}, err
		}
		return Certificate{Code: existing.CertificateCode, TrackID: trackID, IssuedAt: existing.IssuedAt.Time}, nil
	default:
		return Certificate{}, err
	}
}

// ListForUser returns the user's certificates, newest first.
func (s *CertificatesService) ListForUser(ctx context.Context, userID string) ([]Certificate, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return nil, apierr.Unauthorized("invalid user")
	}
	rows, err := s.queries.ListUserCertificates(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]Certificate, len(rows))
	for i, r := range rows {
		out[i] = Certificate{
			Code: r.CertificateCode, TrackID: pgxutil.UUIDString(r.TrackID),
			TrackTitle: r.TrackTitle, IssuedAt: r.IssuedAt.Time,
		}
	}
	return out, nil
}

// Verify looks up a certificate by its public code (no auth required).
func (s *CertificatesService) Verify(ctx context.Context, code string) (CertificateVerification, error) {
	row, err := s.queries.GetCertificateByCode(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return CertificateVerification{}, apierr.NotFound("certificate not found")
	}
	if err != nil {
		return CertificateVerification{}, err
	}
	return CertificateVerification{
		Code: row.CertificateCode, DisplayName: row.DisplayName,
		TrackTitle: row.TrackTitle, IssuedAt: row.IssuedAt.Time,
	}, nil
}

// newCertificateCode returns a random certificate code like "GOAC-XXXXXXXXXXXXXXXX".
func newCertificateCode() (string, error) {
	var b [10]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "GOAC-" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b[:]), nil
}
