package admin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/goacademy/backend/internal/platform/pgxutil"
	"github.com/goacademy/backend/internal/store"
)

func randUUID(t *testing.T) string {
	t.Helper()
	u, err := pgxutil.NewUUID()
	if err != nil {
		t.Fatalf("uuid: %v", err)
	}
	return pgxutil.UUIDString(u)
}

func TestAdminTrackCRUD_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)
	q := store.New(pool)

	tr, err := svc.CreateTrack(ctx, TrackInput{
		Title: "Track", Description: "d", Level: "beginner", Position: 1, Language: "en",
		Items: []TrackItemInput{
			{ContentType: "video", ContentID: randUUID(t)},
			{ContentType: "article", ContentID: randUUID(t)},
		},
	})
	if err != nil {
		t.Fatalf("CreateTrack: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM tracks WHERE id = $1", tr.ID) })
	if items, _ := q.ListTrackItems(ctx, tr.ID); len(items) != 2 {
		t.Fatalf("want 2 track items, got %d", len(items))
	}

	// Update replaces items with one.
	_, err = svc.UpdateTrack(ctx, pgxutil.UUIDString(tr.ID), TrackInput{
		Title: "Track v2", Level: "advanced", Position: 2, Language: "ru",
		Items: []TrackItemInput{{ContentType: "quiz", ContentID: randUUID(t)}},
	})
	if err != nil {
		t.Fatalf("UpdateTrack: %v", err)
	}
	if items, _ := q.ListTrackItems(ctx, tr.ID); len(items) != 1 {
		t.Errorf("update should replace items, got %d", len(items))
	}

	// Validation: bad content type.
	if _, err := svc.CreateTrack(ctx, TrackInput{Title: "x", Level: "beginner", Language: "en",
		Items: []TrackItemInput{{ContentType: "bogus", ContentID: randUUID(t)}}}); err == nil {
		t.Error("bad item content_type should fail")
	}
	if err := svc.DeleteTrack(ctx, pgxutil.UUIDString(tr.ID)); err != nil {
		t.Fatalf("DeleteTrack: %v", err)
	}
}

func TestAdminProjectCRUD_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)
	q := store.New(pool)

	p, err := svc.CreateProject(ctx, ProjectInput{
		Title: "Proj", DescriptionMarkdown: "d", Difficulty: "beginner", Language: "en",
		Steps: []ProjectStepInput{{Text: "s1"}, {Text: "s2"}, {Text: "s3"}},
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM mini_projects WHERE id = $1", p.ID) })
	if steps, _ := q.ListProjectSteps(ctx, p.ID); len(steps) != 3 {
		t.Fatalf("want 3 steps, got %d", len(steps))
	}

	if _, err := svc.UpdateProject(ctx, pgxutil.UUIDString(p.ID), ProjectInput{
		Title: "Proj v2", Difficulty: "advanced", Language: "ru", Steps: []ProjectStepInput{{Text: "only"}},
	}); err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if steps, _ := q.ListProjectSteps(ctx, p.ID); len(steps) != 1 {
		t.Errorf("update should replace steps, got %d", len(steps))
	}
	// Empty step text fails.
	if _, err := svc.CreateProject(ctx, ProjectInput{Title: "x", Difficulty: "beginner", Language: "en", Steps: []ProjectStepInput{{Text: " "}}}); err == nil {
		t.Error("blank step should fail")
	}
	if err := svc.DeleteProject(ctx, pgxutil.UUIDString(p.ID)); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
}

func TestAdminCheatsheetCRUD_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)

	c, err := svc.CreateCheatsheet(ctx, CheatsheetInput{Title: "Slices", Category: "stdlib", BodyMarkdown: "...", Language: "en"})
	if err != nil {
		t.Fatalf("CreateCheatsheet: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM cheatsheets WHERE id = $1", c.ID) })
	if _, err := svc.UpdateCheatsheet(ctx, pgxutil.UUIDString(c.ID), CheatsheetInput{Title: "Slices v2", Category: "core", Language: "ru"}); err != nil {
		t.Fatalf("UpdateCheatsheet: %v", err)
	}
	if _, err := svc.CreateCheatsheet(ctx, CheatsheetInput{Title: "", Language: "en"}); err == nil {
		t.Error("empty title should fail")
	}
	if err := svc.DeleteCheatsheet(ctx, pgxutil.UUIDString(c.ID)); err != nil {
		t.Fatalf("DeleteCheatsheet: %v", err)
	}
}

func TestAdminGlossaryCRUD_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)
	term := "Goroutine-" + marker()

	g, err := svc.CreateGlossaryTerm(ctx, GlossaryInput{Term: term, DefinitionMarkdown: "a lightweight thread", Language: "en"})
	if err != nil {
		t.Fatalf("CreateGlossaryTerm: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM glossary_terms WHERE id = $1", g.ID) })
	// Duplicate term -> conflict.
	if _, err := svc.CreateGlossaryTerm(ctx, GlossaryInput{Term: term, DefinitionMarkdown: "x", Language: "en"}); err == nil {
		t.Error("duplicate term should conflict")
	}
	if _, err := svc.UpdateGlossaryTerm(ctx, pgxutil.UUIDString(g.ID), GlossaryInput{Term: term + "-v2", DefinitionMarkdown: "y", Language: "ru"}); err != nil {
		t.Fatalf("UpdateGlossaryTerm: %v", err)
	}
	if err := svc.DeleteGlossaryTerm(ctx, pgxutil.UUIDString(g.ID)); err != nil {
		t.Fatalf("DeleteGlossaryTerm: %v", err)
	}
}

func TestAdminBadgeCRUD_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)
	code := "test-badge-" + marker()

	b, err := svc.CreateBadge(ctx, BadgeInput{
		Code: code, Title: "Test Badge", Description: "d", Icon: "*",
		CriteriaType: "xp_at_least", CriteriaParams: json.RawMessage(`{"xp":1000}`),
	})
	if err != nil {
		t.Fatalf("CreateBadge: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM badges WHERE id = $1", b.ID) })
	// Duplicate code -> conflict.
	if _, err := svc.CreateBadge(ctx, BadgeInput{Code: code, Title: "x", CriteriaType: "xp_at_least"}); err == nil {
		t.Error("duplicate code should conflict")
	}
	// Invalid criteria_params JSON.
	if _, err := svc.CreateBadge(ctx, BadgeInput{Code: "z-" + marker(), Title: "x", CriteriaType: "xp_at_least", CriteriaParams: json.RawMessage("{bad")}); err == nil {
		t.Error("invalid criteria_params should fail")
	}
	if _, err := svc.UpdateBadge(ctx, pgxutil.UUIDString(b.ID), BadgeInput{Code: code, Title: "Test Badge v2", CriteriaType: "streak_at_least", CriteriaParams: json.RawMessage(`{"days":30}`)}); err != nil {
		t.Fatalf("UpdateBadge: %v", err)
	}
	if err := svc.DeleteBadge(ctx, pgxutil.UUIDString(b.ID)); err != nil {
		t.Fatalf("DeleteBadge: %v", err)
	}
}

func TestAdminDailyChallengeCRUD_Integration(t *testing.T) {
	pool := openPool(t)
	ctx := context.Background()
	svc := NewService(pool)
	// A spread-out future date to avoid colliding with other rows.
	day := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(time.Now().UnixNano()%360))
	dateStr := day.Format("2006-01-02")
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM daily_challenges WHERE challenge_date = $1", dateStr)
	})

	d, err := svc.CreateDailyChallenge(ctx, DailyChallengeInput{ChallengeDate: dateStr, ContentType: "quiz", ContentID: randUUID(t), BonusXP: 25})
	if err != nil {
		t.Fatalf("CreateDailyChallenge: %v", err)
	}
	// Duplicate date -> conflict.
	if _, err := svc.CreateDailyChallenge(ctx, DailyChallengeInput{ChallengeDate: dateStr, ContentType: "problem", ContentID: randUUID(t), BonusXP: 10}); err == nil {
		t.Error("duplicate date should conflict")
	}
	// Bad date.
	if _, err := svc.CreateDailyChallenge(ctx, DailyChallengeInput{ChallengeDate: "not-a-date", ContentType: "quiz", ContentID: randUUID(t)}); err == nil {
		t.Error("bad date should fail validation")
	}
	if _, err := svc.UpdateDailyChallenge(ctx, pgxutil.UUIDString(d.ID), DailyChallengeInput{ChallengeDate: dateStr, ContentType: "problem", ContentID: randUUID(t), BonusXP: 50}); err != nil {
		t.Fatalf("UpdateDailyChallenge: %v", err)
	}
	if err := svc.DeleteDailyChallenge(ctx, pgxutil.UUIDString(d.ID)); err != nil {
		t.Fatalf("DeleteDailyChallenge: %v", err)
	}
}
