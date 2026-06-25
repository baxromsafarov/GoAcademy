package progress

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/goacademy/backend/internal/activity"
	"github.com/goacademy/backend/internal/platform/apierr"
	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// ProjectProgressResult is a user's checklist progress for a mini-project.
type ProjectProgressResult struct {
	ProjectID        string
	CompletedStepIDs []string
	Total            int
	Completed        int
	ProjectComplete  bool
}

// ProjectProgress returns the user's current checklist progress for a project.
func (s *Service) ProjectProgress(ctx context.Context, userID, projectID string) (ProjectProgressResult, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return ProjectProgressResult{}, apierr.Unauthorized("invalid user")
	}
	pid, err := pgxutil.ParseUUID(projectID)
	if err != nil {
		return ProjectProgressResult{}, apierr.NotFound("project not found")
	}

	steps, err := s.projectSteps(ctx, pid)
	if err != nil {
		return ProjectProgressResult{}, err
	}
	completed, err := s.completedSteps(ctx, uid, pid)
	if err != nil {
		return ProjectProgressResult{}, err
	}
	return buildProjectProgress(projectID, steps, completed), nil
}

// ToggleProjectStep flips the completion of one checklist step (idempotent per
// final state) and records a "project_completed" activity the first time all
// steps are done.
func (s *Service) ToggleProjectStep(ctx context.Context, userID, projectID, stepID string) (ProjectProgressResult, error) {
	uid, err := pgxutil.ParseUUID(userID)
	if err != nil {
		return ProjectProgressResult{}, apierr.Unauthorized("invalid user")
	}
	pid, err := pgxutil.ParseUUID(projectID)
	if err != nil {
		return ProjectProgressResult{}, apierr.NotFound("project not found")
	}

	steps, err := s.projectSteps(ctx, pid)
	if err != nil {
		return ProjectProgressResult{}, err
	}
	valid := false
	for _, st := range steps {
		if pgxutil.UUIDString(st.ID) == stepID {
			valid = true
			break
		}
	}
	if !valid {
		return ProjectProgressResult{}, apierr.NotFound("step not found in this project")
	}

	completed, err := s.completedSteps(ctx, uid, pid)
	if err != nil {
		return ProjectProgressResult{}, err
	}
	wasComplete := allStepsDone(steps, completed)

	if completed[stepID] {
		delete(completed, stepID)
	} else {
		completed[stepID] = true
	}

	// Persist only valid step ids, in step order (drops any stale ids).
	ordered := make([]string, 0, len(completed))
	for _, st := range steps {
		id := pgxutil.UUIDString(st.ID)
		if completed[id] {
			ordered = append(ordered, id)
		}
	}
	raw, err := json.Marshal(ordered)
	if err != nil {
		return ProjectProgressResult{}, err
	}
	if _, err := s.queries.UpsertProjectProgress(ctx, store.UpsertProjectProgressParams{
		UserID: uid, ProjectID: pid, CompletedSteps: raw,
	}); err != nil {
		return ProjectProgressResult{}, err
	}

	result := buildProjectProgress(projectID, steps, setOf(ordered))
	if !wasComplete && result.ProjectComplete {
		_ = s.activity.Record(ctx, activity.Event{
			UserID: userID, Type: "project_completed", RefType: "project", RefID: projectID,
		})
	}
	return result, nil
}

// projectSteps loads a project's ordered steps, returning 404 if the project does
// not exist.
func (s *Service) projectSteps(ctx context.Context, pid pgtype.UUID) ([]store.MiniProjectStep, error) {
	if _, err := s.queries.GetProjectByID(ctx, pid); errors.Is(err, pgx.ErrNoRows) {
		return nil, apierr.NotFound("project not found")
	} else if err != nil {
		return nil, err
	}
	return s.queries.ListProjectSteps(ctx, pid)
}

// completedSteps returns the set of step ids the user has marked done.
func (s *Service) completedSteps(ctx context.Context, uid, pid pgtype.UUID) (map[string]bool, error) {
	prog, err := s.queries.GetProjectProgress(ctx, store.GetProjectProgressParams{UserID: uid, ProjectID: pid})
	if errors.Is(err, pgx.ErrNoRows) {
		return map[string]bool{}, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := json.Unmarshal(prog.CompletedSteps, &ids); err != nil {
		return nil, err
	}
	return setOf(ids), nil
}

func buildProjectProgress(projectID string, steps []store.MiniProjectStep, completed map[string]bool) ProjectProgressResult {
	ids := make([]string, 0, len(steps))
	for _, st := range steps {
		id := pgxutil.UUIDString(st.ID)
		if completed[id] {
			ids = append(ids, id)
		}
	}
	total := len(steps)
	return ProjectProgressResult{
		ProjectID:        projectID,
		CompletedStepIDs: ids,
		Total:            total,
		Completed:        len(ids),
		ProjectComplete:  total > 0 && len(ids) == total,
	}
}

func allStepsDone(steps []store.MiniProjectStep, completed map[string]bool) bool {
	if len(steps) == 0 {
		return false
	}
	for _, st := range steps {
		if !completed[pgxutil.UUIDString(st.ID)] {
			return false
		}
	}
	return true
}

func setOf(ids []string) map[string]bool {
	m := make(map[string]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m
}
