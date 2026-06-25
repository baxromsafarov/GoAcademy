package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goacademy/backend/internal/httpapi/respond"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/social"
)

// notesHandler serves personal notes CRUD.
type notesHandler struct {
	svc    *social.NotesService
	logger *slog.Logger
}

func newNotesHandler(svc *social.NotesService, logger *slog.Logger) *notesHandler {
	return &notesHandler{svc: svc, logger: logger}
}

type createNoteRequest struct {
	ContentType string `json:"content_type"`
	ContentID   string `json:"content_id"`
	Body        string `json:"body"`
}

// create handles POST /api/v1/notes.
func (h *notesHandler) create(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	var req createNoteRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	note, err := h.svc.Create(r.Context(), uid, social.CreateNoteInput{
		ContentType: req.ContentType, ContentID: req.ContentID, Body: req.Body,
	})
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toNoteResponse(note))
}

type updateNoteRequest struct {
	Body string `json:"body"`
}

// update handles PATCH /api/v1/notes/{id}.
func (h *notesHandler) update(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	var req updateNoteRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	note, err := h.svc.Update(r.Context(), uid, chi.URLParam(r, "id"), req.Body)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toNoteResponse(note))
}

// delete handles DELETE /api/v1/notes/{id}.
func (h *notesHandler) delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	if err := h.svc.Delete(r.Context(), uid, chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// list handles GET /api/v1/me/notes.
func (h *notesHandler) list(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, r, h.logger, apierr.Unauthorized("authentication required"))
		return
	}
	notes, err := h.svc.List(r.Context(), uid)
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toNotesListResponse(notes))
}
