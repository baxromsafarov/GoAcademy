package progress

import (
	"context"

	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

// resolveArticle fetches an article by id (when ref parses as a UUID) or by
// slug. It mirrors content.Service so the article progress endpoints accept
// either form — track-program items reference content polymorphically by id.
func (s *Service) resolveArticle(ctx context.Context, ref string) (store.Article, error) {
	if id, err := pgxutil.ParseUUID(ref); err == nil {
		return s.queries.GetArticleByID(ctx, id)
	}
	return s.queries.GetArticleBySlug(ctx, ref)
}

// resolveProblem fetches a problem by id (when ref parses as a UUID) or by slug.
func (s *Service) resolveProblem(ctx context.Context, ref string) (store.Problem, error) {
	if id, err := pgxutil.ParseUUID(ref); err == nil {
		return s.queries.GetProblemByID(ctx, id)
	}
	return s.queries.GetProblemBySlug(ctx, ref)
}
