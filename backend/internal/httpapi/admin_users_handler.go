package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goacademy/backend/internal/admin"
	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
)

// listUsers handles GET /api/v1/admin/users?q=&limit=&offset=.
func (h *adminHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.svc.ListUsers(r.Context(), q.Get("q"), queryInt(q, "limit", 0), queryInt(q, "offset", 0))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toAdminUserListResponse(list))
}

type adminUpdateUserRequest struct {
	Role      *string `json:"role"`
	IsBlocked *bool   `json:"is_blocked"`
}

// updateUser handles PATCH /api/v1/admin/users/{id} (change role / block status).
func (h *adminHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	acting, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	var req adminUpdateUserRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	u, err := h.svc.UpdateUser(r.Context(), acting, chi.URLParam(r, "id"), admin.UpdateUserInput{
		Role: req.Role, IsBlocked: req.IsBlocked,
	})
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toAdminUserResponse(u))
}
