package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// ProjectStepInput is one checklist step of a mini-project.
type ProjectStepInput struct {
	Text string
}

// ProjectInput is the create/update payload for a mini-project and its steps.
type ProjectInput struct {
	Title               string
	DescriptionMarkdown string
	Difficulty          string
	Language            string
	Tags                []string
	Steps               []ProjectStepInput
}

func (in ProjectInput) validate() error {
	details := map[string]string{}
	validateMeta(details, in.Title, in.Difficulty, in.Language)
	for i, st := range in.Steps {
		if strings.TrimSpace(st.Text) == "" {
			details[fmt.Sprintf("steps[%d].text", i)] = "must not be empty"
		}
	}
	if len(details) > 0 {
		return apierr.Validation("invalid project").WithDetails(details)
	}
	return nil
}

// CreateProject inserts a mini-project and its steps atomically.
func (s *Service) CreateProject(ctx context.Context, in ProjectInput) (store.MiniProject, error) {
	if err := in.validate(); err != nil {
		return store.MiniProject{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return store.MiniProject{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)

	project, err := q.CreateProject(ctx, store.CreateProjectParams{
		Title: in.Title, DescriptionMarkdown: in.DescriptionMarkdown, Difficulty: store.Difficulty(in.Difficulty),
		Tags: normalizeTags(in.Tags), Language: store.Locale(in.Language),
	})
	if err != nil {
		return store.MiniProject{}, err
	}
	if err := insertProjectSteps(ctx, q, project.ID, in.Steps); err != nil {
		return store.MiniProject{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.MiniProject{}, err
	}
	return project, nil
}

// UpdateProject replaces a mini-project's metadata and its steps atomically.
func (s *Service) UpdateProject(ctx context.Context, id string, in ProjectInput) (store.MiniProject, error) {
	pid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return store.MiniProject{}, apierr.NotFound("project not found")
	}
	if err := in.validate(); err != nil {
		return store.MiniProject{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return store.MiniProject{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)

	project, err := q.UpdateProject(ctx, store.UpdateProjectParams{
		ID: pid, Title: in.Title, DescriptionMarkdown: in.DescriptionMarkdown, Difficulty: store.Difficulty(in.Difficulty),
		Tags: normalizeTags(in.Tags), Language: store.Locale(in.Language),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return store.MiniProject{}, apierr.NotFound("project not found")
	}
	if err != nil {
		return store.MiniProject{}, err
	}
	if err := q.DeleteProjectSteps(ctx, pid); err != nil {
		return store.MiniProject{}, err
	}
	if err := insertProjectSteps(ctx, q, pid, in.Steps); err != nil {
		return store.MiniProject{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return store.MiniProject{}, err
	}
	return project, nil
}

func insertProjectSteps(ctx context.Context, q *store.Queries, projectID pgtype.UUID, steps []ProjectStepInput) error {
	for i, st := range steps {
		if _, err := q.CreateProjectStep(ctx, store.CreateProjectStepParams{
			ProjectID: projectID, Text: st.Text, Position: int32(i + 1),
		}); err != nil {
			return err
		}
	}
	return nil
}

// DeleteProject removes a mini-project (cascading its steps and progress).
func (s *Service) DeleteProject(ctx context.Context, id string) error {
	pid, err := pgxutil.ParseUUID(id)
	if err != nil {
		return apierr.NotFound("project not found")
	}
	n, err := s.queries.DeleteProject(ctx, pid)
	if err != nil {
		return err
	}
	if n == 0 {
		return apierr.NotFound("project not found")
	}
	return nil
}
