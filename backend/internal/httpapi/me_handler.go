package httpapi

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/user"
)

// maxAvatarBytes caps avatar uploads.
const maxAvatarBytes = 5 << 20 // 5 MiB

// meHandler serves the authenticated user's own profile (/api/v1/me).
type meHandler struct {
	svc    *user.Service
	logger *slog.Logger
}

func newMeHandler(svc *user.Service, logger *slog.Logger) *meHandler {
	return &meHandler{svc: svc, logger: logger}
}

// get handles GET /api/v1/me.
func (h *meHandler) get(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	u, err := h.svc.GetByID(r.Context(), uid)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toUserResponse(u))
}

type updateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	Location    *string `json:"location"`
	Locale      *string `json:"locale"`
	IsPublic    *bool   `json:"is_public"`
}

// patch handles PATCH /api/v1/me (partial update; only the fields present change).
func (h *meHandler) patch(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}

	var req updateProfileRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}

	u, err := h.svc.Update(r.Context(), uid, user.UpdateInput{
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		Location:    req.Location,
		Locale:      req.Locale,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toUserResponse(u))
}

// uploadAvatar handles POST /api/v1/me/avatar (multipart, field "avatar").
func (h *meHandler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarBytes)
	if err := r.ParseMultipartForm(maxAvatarBytes); err != nil {
		respond.Error(w, r, h.logger, apierr.Validation("upload is invalid or larger than 5 MiB").WithCause(err))
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		respond.Error(w, r, h.logger, apierr.Validation("missing 'avatar' file field").WithCause(err))
		return
	}
	defer file.Close()

	// Sniff the real content type from the first bytes, not the client's claim.
	head := make([]byte, 512)
	n, _ := io.ReadFull(file, head)
	ext, ok := avatarExt(http.DetectContentType(head[:n]))
	if !ok {
		respond.Error(w, r, h.logger, apierr.Validation("unsupported image type (allowed: jpeg, png, gif, webp)"))
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}

	u, err := h.svc.SetAvatar(r.Context(), uid, ext, file)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toUserResponse(u))
}

// avatarExt maps an allowed image content type to a file extension.
func avatarExt(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}
