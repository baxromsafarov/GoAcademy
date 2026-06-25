package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goacademy/backend/internal/content"
	"github.com/goacademy/backend/internal/httpapi/respond"
)

// referenceHandler serves the public reference material: cheatsheets and glossary.
type referenceHandler struct {
	svc    *content.Service
	logger *slog.Logger
}

func newReferenceHandler(svc *content.Service, logger *slog.Logger) *referenceHandler {
	return &referenceHandler{svc: svc, logger: logger}
}

// listCheatsheets handles GET /api/v1/cheatsheets?category=&q=&language=&limit=&offset=.
func (h *referenceHandler) listCheatsheets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.svc.ListCheatsheets(r.Context(), content.CheatsheetFilter{
		Category: optionalQuery(q, "category"),
		Query:    optionalQuery(q, "q"),
		Language: optionalQuery(q, "language"),
		Limit:    queryInt(q, "limit", 0),
		Offset:   queryInt(q, "offset", 0),
	})
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toCheatsheetListResponse(list))
}

// getCheatsheet handles GET /api/v1/cheatsheets/{id}.
func (h *referenceHandler) getCheatsheet(w http.ResponseWriter, r *http.Request) {
	c, err := h.svc.GetCheatsheetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toCheatsheetDetailResponse(c))
}

// listGlossary handles GET /api/v1/glossary?q=&language=&limit=&offset=.
func (h *referenceHandler) listGlossary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.svc.ListGlossary(r.Context(), content.GlossaryFilter{
		Query:    optionalQuery(q, "q"),
		Language: optionalQuery(q, "language"),
		Limit:    queryInt(q, "limit", 0),
		Offset:   queryInt(q, "offset", 0),
	})
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toGlossaryListResponse(list))
}
