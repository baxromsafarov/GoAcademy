package social

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

const maxNoteBody = 10000

// contentTypes are the content kinds a note (or, later, a bookmark) may attach to.
var contentTypes = map[string]bool{
	"video": true, "article": true, "quiz": true, "problem": true,
	"project": true, "track": true, "cheatsheet": true, "glossary": true,
}

func validContentType(t string) bool { return contentTypes[t] }

// Note is a user's private annotation on a piece of content.
type Note struct {
	ID          string
	ContentType string
	ContentID   string
	Body        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NotesService is the CRUD service for personal notes.
type NotesService struct {
	queries *store.Queries
}

// NewNotesService wires the notes service to the database.
func NewNotesService(pool *pgxpool.Pool) *NotesService {
	return &NotesService{queries: store.New(pool)}
}

// CreateNoteInput is a request to create a note.
type CreateNoteInput struct {
	ContentType string
	ContentID   string
	Body        string
}

func (in CreateNoteInput) validate() (pgtype.UUID, error) {
	details := map[string]string{}
	if !validContentType(in.ContentType) {
		details["content_type"] = "must be one of: video, article, quiz, problem, project, track, cheatsheet, glossary"
	}
	cid, err := pgxutil.ParseUUID(in.ContentID)
	if err != nil {
		details["content_id"] = "must be a valid uuid"
	}
	if msg := validateBody(in.Body); msg != "" {
		details["body"] = msg
	}
	if len(details) > 0 {
		return pgtype.UUID{}, apierr.Validation("invalid note").WithDetails(details)
	}
	return cid, nil
}

// validateBody returns an error message, or "" if the body is acceptable.
func validateBody(body string) string {
	switch {
	case strings.TrimSpace(body) == "":
		return "must not be empty"
	case len(body) > maxNoteBody:
		return "must be at most 10000 characters"
	default:
		return ""
	}
}

// Create stores a new note owned by the user.
func (s *NotesService) Create(ctx context.Context, userID string, in CreateNoteInput) (Note, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return Note{}, apierr.Unauthorized("invalid user")
	}
	cid, err := in.validate()
	if err != nil {
		return Note{}, err
	}
	row, err := s.queries.CreateNote(ctx, store.CreateNoteParams{
		UserID: uid, ContentType: in.ContentType, ContentID: cid, Body: in.Body,
	})
	if err != nil {
		return Note{}, err
	}
	return toNote(row), nil
}

// Update changes a note's body. Only the owner can update it; anyone else (or a
// missing note) gets not-found.
func (s *NotesService) Update(ctx context.Context, userID, noteID, body string) (Note, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return Note{}, apierr.Unauthorized("invalid user")
	}
	nid, err := pgxutil.ParseUUID(noteID)
	if err != nil {
		return Note{}, apierr.NotFound("note not found")
	}
	if msg := validateBody(body); msg != "" {
		return Note{}, apierr.Validation("invalid note").WithDetails(map[string]string{"body": msg})
	}
	row, err := s.queries.UpdateNote(ctx, store.UpdateNoteParams{Body: body, ID: nid, UserID: uid})
	if errors.Is(err, pgx.ErrNoRows) {
		return Note{}, apierr.NotFound("note not found")
	}
	if err != nil {
		return Note{}, err
	}
	return toNote(row), nil
}

// Delete removes a note. Only the owner can delete it.
func (s *NotesService) Delete(ctx context.Context, userID, noteID string) error {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return apierr.Unauthorized("invalid user")
	}
	nid, err := pgxutil.ParseUUID(noteID)
	if err != nil {
		return apierr.NotFound("note not found")
	}
	n, err := s.queries.DeleteNote(ctx, store.DeleteNoteParams{ID: nid, UserID: uid})
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("note not found")
	}
	return nil
}

// List returns the user's notes, newest first.
func (s *NotesService) List(ctx context.Context, userID string) ([]Note, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return nil, apierr.Unauthorized("invalid user")
	}
	rows, err := s.queries.ListUserNotes(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]Note, len(rows))
	for i, r := range rows {
		out[i] = toNote(r)
	}
	return out, nil
}

func toNote(n store.Note) Note {
	return Note{
		ID:          pgxutil.UUIDString(n.ID),
		ContentType: n.ContentType,
		ContentID:   pgxutil.UUIDString(n.ContentID),
		Body:        n.Body,
		CreatedAt:   n.CreatedAt.Time,
		UpdatedAt:   n.UpdatedAt.Time,
	}
}
