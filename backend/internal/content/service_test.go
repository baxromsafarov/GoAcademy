package content

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goacademy/backend/internal/platform/apierr"
)

func strptr(s string) *string { return &s }

func TestListFilter_RejectsInvalidEnums(t *testing.T) {
	f := ListFilter{Difficulty: strptr("super-hard"), Language: strptr("fr")}
	err := f.normalizeAndValidate()

	var apiErr *apierr.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 400 {
		t.Fatalf("expected 400, got %v", err)
	}
	details, _ := apiErr.Details.(map[string]string)
	for _, k := range []string{"difficulty", "language"} {
		if _, ok := details[k]; !ok {
			t.Errorf("missing validation error for %q", k)
		}
	}
}

func TestListFilter_ClampsPagination(t *testing.T) {
	f := ListFilter{Limit: 0, Offset: -5}
	if err := f.normalizeAndValidate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Limit != DefaultPageSize {
		t.Errorf("limit = %d, want default %d", f.Limit, DefaultPageSize)
	}
	if f.Offset != 0 {
		t.Errorf("offset = %d, want 0", f.Offset)
	}

	big := ListFilter{Limit: 10_000}
	_ = big.normalizeAndValidate()
	if big.Limit != MaxPageSize {
		t.Errorf("limit = %d, want clamped to %d", big.Limit, MaxPageSize)
	}
}

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run content integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func seedVideo(t *testing.T, pool *pgxpool.Pool, marker, difficulty, language string, extraTags ...string) string {
	t.Helper()
	tags := append([]string{marker}, extraTags...)
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO videos (title, youtube_id, difficulty, language, tags)
		 VALUES ($1, $2, $3::difficulty, $4::locale, $5) RETURNING id::text`,
		"Test "+difficulty, "yt-"+marker, difficulty, language, tags).Scan(&id)
	if err != nil {
		t.Fatalf("seed video: %v", err)
	}
	return id
}

func TestService_ListAndGetVideos_Integration(t *testing.T) {
	pool := openPool(t)
	svc := NewService(pool)
	ctx := context.Background()
	marker := fmt.Sprintf("itest-%d", time.Now().UnixNano())
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM videos WHERE $1 = ANY(tags)", marker) })

	v1 := seedVideo(t, pool, marker, "beginner", "en", "go")
	seedVideo(t, pool, marker, "intermediate", "en")
	seedVideo(t, pool, marker, "advanced", "ru", "concurrency")

	// All three for this marker.
	all, err := svc.ListVideos(ctx, ListFilter{Tag: &marker})
	if err != nil {
		t.Fatalf("ListVideos: %v", err)
	}
	if all.Total != 3 || len(all.Items) != 3 {
		t.Fatalf("expected 3 videos, got total=%d items=%d", all.Total, len(all.Items))
	}

	// Filter by difficulty.
	beg, _ := svc.ListVideos(ctx, ListFilter{Tag: &marker, Difficulty: strptr("beginner")})
	if beg.Total != 1 {
		t.Errorf("difficulty=beginner total = %d, want 1", beg.Total)
	}
	// Filter by language.
	ru, _ := svc.ListVideos(ctx, ListFilter{Tag: &marker, Language: strptr("ru")})
	if ru.Total != 1 {
		t.Errorf("language=ru total = %d, want 1", ru.Total)
	}

	// Pagination: total stays 3, page size honoured.
	page1, _ := svc.ListVideos(ctx, ListFilter{Tag: &marker, Limit: 2})
	if page1.Total != 3 || len(page1.Items) != 2 {
		t.Errorf("page1: total=%d items=%d, want 3/2", page1.Total, len(page1.Items))
	}
	page2, _ := svc.ListVideos(ctx, ListFilter{Tag: &marker, Limit: 2, Offset: 2})
	if page2.Total != 3 || len(page2.Items) != 1 {
		t.Errorf("page2: total=%d items=%d, want 3/1", page2.Total, len(page2.Items))
	}

	// Get by id.
	got, err := svc.GetVideoByID(ctx, v1)
	if err != nil {
		t.Fatalf("GetVideoByID: %v", err)
	}
	if got.Difficulty != "beginner" {
		t.Errorf("difficulty = %q, want beginner", got.Difficulty)
	}

	// Unknown id → not found.
	if _, err := svc.GetVideoByID(ctx, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Error("unknown id should be not found")
	}
}

func seedArticle(t *testing.T, pool *pgxpool.Pool, marker, slug, difficulty, language string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO articles (title, slug, body_markdown, difficulty, language, tags)
		 VALUES ($1, $2, $3, $4::difficulty, $5::locale, $6)`,
		"Article "+slug, slug, "# "+slug, difficulty, language, []string{marker})
	if err != nil {
		t.Fatalf("seed article: %v", err)
	}
}

func TestService_ListAndGetArticles_Integration(t *testing.T) {
	pool := openPool(t)
	svc := NewService(pool)
	ctx := context.Background()
	now := time.Now().UnixNano()
	marker := fmt.Sprintf("aitest-%d", now)
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM articles WHERE $1 = ANY(tags)", marker) })

	s1 := fmt.Sprintf("intro-go-%d", now)
	s2 := fmt.Sprintf("go-concurrency-%d", now)
	seedArticle(t, pool, marker, s1, "beginner", "en")
	seedArticle(t, pool, marker, s2, "advanced", "ru")

	all, err := svc.ListArticles(ctx, ListFilter{Tag: &marker})
	if err != nil {
		t.Fatalf("ListArticles: %v", err)
	}
	if all.Total != 2 || len(all.Items) != 2 {
		t.Fatalf("expected 2 articles, got total=%d items=%d", all.Total, len(all.Items))
	}

	beg, _ := svc.ListArticles(ctx, ListFilter{Tag: &marker, Difficulty: strptr("beginner")})
	if beg.Total != 1 {
		t.Errorf("difficulty=beginner total = %d, want 1", beg.Total)
	}

	got, err := svc.GetArticleBySlug(ctx, s1)
	if err != nil {
		t.Fatalf("GetArticleBySlug: %v", err)
	}
	if got.Slug != s1 {
		t.Errorf("slug = %q, want %q", got.Slug, s1)
	}

	// citext slug: an uppercase request must match.
	if _, err := svc.GetArticleBySlug(ctx, strings.ToUpper(s1)); err != nil {
		t.Errorf("case-insensitive slug lookup should match: %v", err)
	}

	// Unknown slug → not found.
	if _, err := svc.GetArticleBySlug(ctx, "no-such-slug-here"); err == nil {
		t.Error("unknown slug should be not found")
	}

	// Slug uniqueness: a duplicate insert must fail.
	if _, err := pool.Exec(ctx,
		"INSERT INTO articles (title, slug, tags) VALUES ('dup', $1, ARRAY[$2])", s1, marker,
	); err == nil {
		t.Error("duplicate slug should violate the unique constraint")
	}
}

func TestService_Quizzes_Integration(t *testing.T) {
	pool := openPool(t)
	svc := NewService(pool)
	ctx := context.Background()
	marker := fmt.Sprintf("qztest-%d", time.Now().UnixNano())
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM quizzes WHERE $1 = ANY(tags)", marker) })

	var quizID string
	if err := pool.QueryRow(ctx,
		"INSERT INTO quizzes (title, pass_threshold, tags) VALUES ('Quiz', 60, ARRAY[$1]) RETURNING id::text", marker,
	).Scan(&quizID); err != nil {
		t.Fatalf("seed quiz: %v", err)
	}

	var q1, q2 string
	if err := pool.QueryRow(ctx,
		"INSERT INTO quiz_questions (quiz_id, prompt, type, position) VALUES ($1,'Q1','single',1) RETURNING id::text", quizID,
	).Scan(&q1); err != nil {
		t.Fatalf("seed q1: %v", err)
	}
	if err := pool.QueryRow(ctx,
		"INSERT INTO quiz_questions (quiz_id, prompt, type, position) VALUES ($1,'Q2','multiple',2) RETURNING id::text", quizID,
	).Scan(&q2); err != nil {
		t.Fatalf("seed q2: %v", err)
	}
	if _, err := pool.Exec(ctx,
		"INSERT INTO quiz_options (question_id, text, is_correct, position) VALUES ($1,'a',true,1),($1,'b',false,2)", q1,
	); err != nil {
		t.Fatalf("seed q1 options: %v", err)
	}
	if _, err := pool.Exec(ctx,
		"INSERT INTO quiz_options (question_id, text, is_correct, position) VALUES ($1,'x',true,1),($1,'y',true,2),($1,'z',false,3)", q2,
	); err != nil {
		t.Fatalf("seed q2 options: %v", err)
	}

	detail, err := svc.GetQuizDetail(ctx, quizID)
	if err != nil {
		t.Fatalf("GetQuizDetail: %v", err)
	}
	if detail.Quiz.PassThreshold != 60 {
		t.Errorf("pass_threshold = %d, want 60", detail.Quiz.PassThreshold)
	}
	if len(detail.Questions) != 2 {
		t.Fatalf("questions = %d, want 2", len(detail.Questions))
	}
	if detail.Questions[0].Question.Prompt != "Q1" || detail.Questions[1].Question.Prompt != "Q2" {
		t.Error("questions are not ordered by position")
	}
	if len(detail.Questions[0].Options) != 2 || len(detail.Questions[1].Options) != 3 {
		t.Errorf("option counts = %d/%d, want 2/3", len(detail.Questions[0].Options), len(detail.Questions[1].Options))
	}
	if detail.Questions[0].Options[0].Text != "a" || detail.Questions[0].Options[1].Text != "b" {
		t.Error("options are not ordered by position")
	}

	list, err := svc.ListQuizzes(ctx, ListFilter{Tag: &marker})
	if err != nil {
		t.Fatalf("ListQuizzes: %v", err)
	}
	if list.Total != 1 {
		t.Errorf("list total = %d, want 1", list.Total)
	}

	if _, err := svc.GetQuizDetail(ctx, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Error("unknown quiz should be not found")
	}
}

func TestService_Problems_Integration(t *testing.T) {
	pool := openPool(t)
	svc := NewService(pool)
	ctx := context.Background()
	now := time.Now().UnixNano()
	marker := fmt.Sprintf("pitest-%d", now)
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM problems WHERE $1 = ANY(tags)", marker) })

	slug := fmt.Sprintf("two-sum-%d", now)
	if _, err := pool.Exec(ctx,
		`INSERT INTO problems (title, slug, statement_markdown, reference_solution_markdown, sample_io, difficulty, language, tags)
		 VALUES ('Two Sum', $1, 'Statement', 'SOLUTION', '[{"input":"a","output":"b"}]'::jsonb, 'beginner'::difficulty, 'en'::locale, ARRAY[$2])`,
		slug, marker); err != nil {
		t.Fatalf("seed problem: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO problems (title, slug, difficulty, tags) VALUES ('P2', $1, 'advanced'::difficulty, ARRAY[$2])`,
		fmt.Sprintf("p2-%d", now), marker); err != nil {
		t.Fatalf("seed problem 2: %v", err)
	}

	// The service returns the full row (reference solution included); the HTTP
	// layer is what hides it.
	got, err := svc.GetProblemBySlug(ctx, slug)
	if err != nil {
		t.Fatalf("GetProblemBySlug: %v", err)
	}
	if got.ReferenceSolutionMarkdown != "SOLUTION" {
		t.Error("service should return the reference solution")
	}
	if len(got.SampleIo) == 0 {
		t.Error("sample_io should be present")
	}

	// citext slug: uppercase request matches.
	if _, err := svc.GetProblemBySlug(ctx, strings.ToUpper(slug)); err != nil {
		t.Errorf("case-insensitive slug lookup should match: %v", err)
	}

	list, _ := svc.ListProblems(ctx, ListFilter{Tag: &marker})
	if list.Total != 2 {
		t.Errorf("list total = %d, want 2", list.Total)
	}
	beg, _ := svc.ListProblems(ctx, ListFilter{Tag: &marker, Difficulty: strptr("beginner")})
	if beg.Total != 1 {
		t.Errorf("difficulty=beginner total = %d, want 1", beg.Total)
	}

	if _, err := svc.GetProblemBySlug(ctx, "no-such-problem-slug"); err == nil {
		t.Error("unknown slug should be not found")
	}

	if _, err := pool.Exec(ctx, "INSERT INTO problems (title, slug, tags) VALUES ('dup', $1, ARRAY[$2])", slug, marker); err == nil {
		t.Error("duplicate slug should violate the unique constraint")
	}
}

func TestService_Tracks_Integration(t *testing.T) {
	pool := openPool(t)
	svc := NewService(pool)
	ctx := context.Background()
	title := fmt.Sprintf("Go Basics %d", time.Now().UnixNano())

	var trackID string
	if err := pool.QueryRow(ctx,
		"INSERT INTO tracks (title, level, language) VALUES ($1, 'beginner'::difficulty, 'en'::locale) RETURNING id::text", title,
	).Scan(&trackID); err != nil {
		t.Fatalf("seed track: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM tracks WHERE id = $1", trackID) })

	// Insert heterogeneous items with positions out of insertion order.
	for _, it := range []struct {
		ctype string
		pos   int
	}{{"video", 3}, {"article", 1}, {"quiz", 2}} {
		if _, err := pool.Exec(ctx,
			"INSERT INTO track_items (track_id, content_type, content_id, position) VALUES ($1, $2::track_content_type, gen_random_uuid(), $3)",
			trackID, it.ctype, it.pos); err != nil {
			t.Fatalf("seed item %s: %v", it.ctype, err)
		}
	}

	detail, err := svc.GetTrackDetail(ctx, trackID)
	if err != nil {
		t.Fatalf("GetTrackDetail: %v", err)
	}
	if len(detail.Items) != 3 {
		t.Fatalf("items = %d, want 3", len(detail.Items))
	}
	// Ordered by position: article(1), quiz(2), video(3).
	if detail.Items[0].ContentType != "article" || detail.Items[1].ContentType != "quiz" || detail.Items[2].ContentType != "video" {
		t.Errorf("items not ordered by position: %q, %q, %q",
			detail.Items[0].ContentType, detail.Items[1].ContentType, detail.Items[2].ContentType)
	}

	// Level filter includes our beginner track.
	list, err := svc.ListTracks(ctx, ListFilter{Difficulty: strptr("beginner")})
	if err != nil {
		t.Fatalf("ListTracks: %v", err)
	}
	if list.Total < 1 {
		t.Error("list should include at least the seeded track")
	}

	if _, err := svc.GetTrackDetail(ctx, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Error("unknown track should be not found")
	}
}

func TestService_CheatsheetsAndGlossary_Integration(t *testing.T) {
	pool := openPool(t)
	svc := NewService(pool)
	ctx := context.Background()
	now := time.Now().UnixNano()

	// --- cheatsheets ---
	cat := fmt.Sprintf("cat-%d", now)
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM cheatsheets WHERE category = $1", cat) })
	var csID string
	if err := pool.QueryRow(ctx,
		"INSERT INTO cheatsheets (title, category, body_markdown) VALUES ($1, $2, 'b') RETURNING id::text",
		fmt.Sprintf("Slices %d", now), cat).Scan(&csID); err != nil {
		t.Fatalf("seed cheatsheet: %v", err)
	}
	if _, err := pool.Exec(ctx,
		"INSERT INTO cheatsheets (title, category, body_markdown) VALUES ($1, $2, 'b')",
		fmt.Sprintf("Maps %d", now), cat); err != nil {
		t.Fatalf("seed cheatsheet 2: %v", err)
	}

	byCat, err := svc.ListCheatsheets(ctx, CheatsheetFilter{Category: &cat})
	if err != nil {
		t.Fatalf("ListCheatsheets: %v", err)
	}
	if byCat.Total != 2 {
		t.Errorf("category total = %d, want 2", byCat.Total)
	}

	maps := "Maps"
	search, _ := svc.ListCheatsheets(ctx, CheatsheetFilter{Category: &cat, Query: &maps})
	if search.Total != 1 {
		t.Errorf("search 'Maps' total = %d, want 1", search.Total)
	}

	if c, err := svc.GetCheatsheetByID(ctx, csID); err != nil || c.Category != cat {
		t.Errorf("GetCheatsheetByID = (%+v, %v)", c, err)
	}
	if _, err := svc.GetCheatsheetByID(ctx, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Error("unknown cheatsheet should be not found")
	}

	// --- glossary ---
	marker := fmt.Sprintf("gm%d", now)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM glossary_terms WHERE term ILIKE '%' || $1 || '%'", marker)
	})
	term1 := "Goroutine " + marker
	if _, err := pool.Exec(ctx, "INSERT INTO glossary_terms (term, definition_markdown) VALUES ($1, 'concurrent function')", term1); err != nil {
		t.Fatalf("seed term1: %v", err)
	}
	if _, err := pool.Exec(ctx, "INSERT INTO glossary_terms (term, definition_markdown) VALUES ($1, 'typed conduit')", "Channel "+marker); err != nil {
		t.Fatalf("seed term2: %v", err)
	}

	all, _ := svc.ListGlossary(ctx, GlossaryFilter{Query: &marker})
	if all.Total != 2 {
		t.Errorf("glossary search '%s' total = %d, want 2", marker, all.Total)
	}
	gq := "Goroutine " + marker
	one, _ := svc.ListGlossary(ctx, GlossaryFilter{Query: &gq})
	if one.Total != 1 {
		t.Errorf("glossary search '%s' total = %d, want 1", gq, one.Total)
	}

	// Term uniqueness is case-insensitive (citext).
	if _, err := pool.Exec(ctx, "INSERT INTO glossary_terms (term, definition_markdown) VALUES ($1, 'dup')", strings.ToUpper(term1)); err == nil {
		t.Error("duplicate term (different case) should violate the unique constraint")
	}
}
