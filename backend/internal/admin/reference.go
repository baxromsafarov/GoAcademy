package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// CheatsheetInput is the create/update payload for a cheatsheet.
type CheatsheetInput struct {
	Title        string
	Category     string
	BodyMarkdown string
	Language     string
}

func (in CheatsheetInput) validate() error {
	details := map[string]string{}
	if strings.TrimSpace(in.Title) == "" {
		details["title"] = "must not be empty"
	}
	if !locales[in.Language] {
		details["language"] = "must be ru, en, uz or ja"
	}
	if len(details) > 0 {
		return apierr.Validation("invalid cheatsheet").WithDetails(details)
	}
	return nil
}

// CreateCheatsheet inserts a new cheatsheet.
func (s *Service) CreateCheatsheet(ctx context.Context, in CheatsheetInput) (store.Cheatsheet, error) {
	if err := in.validate(); err != nil {
		return store.Cheatsheet{}, err
	}
	return s.queries.CreateCheatsheet(ctx, store.CreateCheatsheetParams{
		Title: in.Title, Category: in.Category, BodyMarkdown: in.BodyMarkdown, Language: store.Locale(in.Language),
	})
}

// UpdateCheatsheet replaces a cheatsheet's fields.
func (s *Service) UpdateCheatsheet(ctx context.Context, id string, in CheatsheetInput) (store.Cheatsheet, error) {
	cid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.Cheatsheet{}, apierr.NotFound("cheatsheet not found")
	}
	if err := in.validate(); err != nil {
		return store.Cheatsheet{}, err
	}
	c, err := s.queries.UpdateCheatsheet(ctx, store.UpdateCheatsheetParams{
		ID: cid, Title: in.Title, Category: in.Category, BodyMarkdown: in.BodyMarkdown, Language: store.Locale(in.Language),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return store.Cheatsheet{}, apierr.NotFound("cheatsheet not found")
	}
	return c, err
}

// DeleteCheatsheet removes a cheatsheet.
func (s *Service) DeleteCheatsheet(ctx context.Context, id string) error {
	cid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return apierr.NotFound("cheatsheet not found")
	}
	n, err := s.queries.DeleteCheatsheet(ctx, cid)
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("cheatsheet not found")
	}
	return nil
}

// GlossaryInput is the create/update payload for a glossary term.
type GlossaryInput struct {
	Term               string
	DefinitionMarkdown string
	Language           string
}

func (in GlossaryInput) validate() error {
	details := map[string]string{}
	if strings.TrimSpace(in.Term) == "" {
		details["term"] = "must not be empty"
	}
	if !locales[in.Language] {
		details["language"] = "must be ru, en, uz or ja"
	}
	if len(details) > 0 {
		return apierr.Validation("invalid glossary term").WithDetails(details)
	}
	return nil
}

// CreateGlossaryTerm inserts a new glossary term (term must be unique).
func (s *Service) CreateGlossaryTerm(ctx context.Context, in GlossaryInput) (store.GlossaryTerm, error) {
	if err := in.validate(); err != nil {
		return store.GlossaryTerm{}, err
	}
	g, err := s.queries.CreateGlossaryTerm(ctx, store.CreateGlossaryTermParams{
		Term: in.Term, DefinitionMarkdown: in.DefinitionMarkdown, Language: store.Locale(in.Language),
	})
	if isUniqueViolation(err) {
		return store.GlossaryTerm{}, apierr.Conflict("a term with this name already exists")
	}
	return g, err
}

// UpdateGlossaryTerm replaces a glossary term's fields.
func (s *Service) UpdateGlossaryTerm(ctx context.Context, id string, in GlossaryInput) (store.GlossaryTerm, error) {
	gid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.GlossaryTerm{}, apierr.NotFound("glossary term not found")
	}
	if err := in.validate(); err != nil {
		return store.GlossaryTerm{}, err
	}
	g, err := s.queries.UpdateGlossaryTerm(ctx, store.UpdateGlossaryTermParams{
		ID: gid, Term: in.Term, DefinitionMarkdown: in.DefinitionMarkdown, Language: store.Locale(in.Language),
	})
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return store.GlossaryTerm{}, apierr.NotFound("glossary term not found")
	case isUniqueViolation(err):
		return store.GlossaryTerm{}, apierr.Conflict("a term with this name already exists")
	}
	return g, err
}

// DeleteGlossaryTerm removes a glossary term.
func (s *Service) DeleteGlossaryTerm(ctx context.Context, id string) error {
	gid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return apierr.NotFound("glossary term not found")
	}
	n, err := s.queries.DeleteGlossaryTerm(ctx, gid)
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("glossary term not found")
	}
	return nil
}
