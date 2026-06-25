package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/goacademy/backend/internal/auth"
	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/store"
)

const (
	refreshCookieName = "refresh_token"
	refreshCookiePath = "/api/v1/auth"
)

// CookieConfig controls how the refresh cookie is written.
type CookieConfig struct {
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

// authHandler serves the /api/v1/auth endpoints.
type authHandler struct {
	svc    *auth.Service
	cookie CookieConfig
	logger *slog.Logger
}

func newAuthHandler(svc *auth.Service, cookie CookieConfig, logger *slog.Logger) *authHandler {
	return &authHandler{svc: svc, cookie: cookie, logger: logger}
}

// --- registration / verification ---

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Locale      string `json:"locale"`
}

func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}

	user, err := h.svc.Register(r.Context(), auth.RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		Locale:      store.Locale(req.Locale),
	})
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}

	respond.JSON(w, http.StatusCreated, toUserResponse(user))
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

func (h *authHandler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	if err := h.svc.VerifyEmail(r.Context(), req.Token); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "verified"})
}

// --- password reset ---

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (h *authHandler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	if err := h.svc.RequestPasswordReset(r.Context(), req.Email); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	// Always 200, regardless of whether the email exists (no enumeration).
	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (h *authHandler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	if err := h.svc.ResetPassword(r.Context(), req.Token, req.Password); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "password_reset"})
}

// --- login / refresh / logout ---

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// tokenResponse returns the access token in the body; the refresh token is sent
// only as an httpOnly cookie (never in the body).
type tokenResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	User        userResponse `json:"user"`
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}

	pair, err := h.svc.Login(r.Context(), req.Email, req.Password, r.UserAgent())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}

	h.setRefreshCookie(w, pair.RefreshToken, pair.RefreshExpiresAt)
	respond.JSON(w, http.StatusOK, tokenResponse{
		AccessToken: pair.AccessToken,
		TokenType:   "Bearer",
		User:        toUserResponse(pair.User),
	})
}

func (h *authHandler) refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		respond.Error(w, r, h.logger, apierr.Unauthorized("missing refresh token"))
		return
	}

	pair, err := h.svc.Refresh(r.Context(), cookie.Value, r.UserAgent())
	if err != nil {
		h.clearRefreshCookie(w) // drop the now-invalid cookie
		respond.Error(w, r, h.logger, err)
		return
	}

	h.setRefreshCookie(w, pair.RefreshToken, pair.RefreshExpiresAt)
	respond.JSON(w, http.StatusOK, tokenResponse{
		AccessToken: pair.AccessToken,
		TokenType:   "Bearer",
		User:        toUserResponse(pair.User),
	})
}

func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(refreshCookieName); err == nil {
		if err := h.svc.Logout(r.Context(), cookie.Value); err != nil {
			respond.Error(w, r, h.logger, err)
			return
		}
	}
	h.clearRefreshCookie(w)
	respond.JSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h *authHandler) setRefreshCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     refreshCookiePath,
		Domain:   h.cookie.Domain,
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: h.cookie.SameSite,
	})
}

func (h *authHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     refreshCookiePath,
		Domain:   h.cookie.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: h.cookie.SameSite,
	})
}
