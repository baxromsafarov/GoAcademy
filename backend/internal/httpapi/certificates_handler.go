package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/social"
)

// certificatesHandler serves certificate issuance, listing and public verification.
type certificatesHandler struct {
	svc    *social.CertificatesService
	logger *slog.Logger
}

func newCertificatesHandler(svc *social.CertificatesService, logger *slog.Logger) *certificatesHandler {
	return &certificatesHandler{svc: svc, logger: logger}
}

// issue handles POST /api/v1/tracks/{id}/certificate (issue if the track is done).
func (h *certificatesHandler) issue(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	cert, err := h.svc.IssueForTrack(r.Context(), uid, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toCertificateResponse(cert))
}

// list handles GET /api/v1/me/certificates.
func (h *certificatesHandler) list(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	certs, err := h.svc.ListForUser(r.Context(), uid)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toCertificatesListResponse(certs))
}

// verify handles GET /api/v1/certificates/{code} (public).
func (h *certificatesHandler) verify(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Verify(r.Context(), chi.URLParam(r, "code"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toCertificateVerificationResponse(v))
}
