// Command seed populates the database with a roadmap-ordered Go curriculum in
// four languages (ru/en/uz/ja): videos, articles, quizzes, problems, projects
// and a learning track per language.
//
// It is idempotent: every row it creates is tagged "seed" (tracks are marked in
// their description), and a re-run deletes the previous seed content first. Real
// user content (without the seed tag) is left untouched.
//
//	go run ./cmd/seed            # needs DATABASE_URL
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// langs is the set of languages every topic is seeded in.
var langs = []string{"ru", "en", "uz", "ja"}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal: "+err.Error())
		os.Exit(1)
	}
}

func run() error {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	s := &seeder{pool: pool, ctx: ctx}
	if err := s.clean(); err != nil {
		return fmt.Errorf("clean: %w", err)
	}
	if err := s.seed(); err != nil {
		return fmt.Errorf("seed: %w", err)
	}
	fmt.Printf("seeded: %d videos, %d articles, %d quizzes, %d problems, %d projects, %d tracks\n",
		s.videos, s.articles, s.quizzes, s.problems, s.projects, s.tracks)
	return nil
}

type seeder struct {
	pool *pgxpool.Pool
	ctx  context.Context

	videos, articles, quizzes, problems, projects, tracks int
}

func (s *seeder) exec(sql string, args ...any) error {
	_, err := s.pool.Exec(s.ctx, sql, args...)
	return err
}

// clean removes previously-seeded content (tagged "seed"; tracks marked in
// description). Cascades handle questions/options/test-cases/steps/items.
func (s *seeder) clean() error {
	stmts := []string{
		`DELETE FROM videos      WHERE 'seed' = ANY(tags)`,
		`DELETE FROM articles    WHERE 'seed' = ANY(tags)`,
		`DELETE FROM quizzes     WHERE 'seed' = ANY(tags)`,
		`DELETE FROM problems    WHERE 'seed' = ANY(tags)`,
		`DELETE FROM mini_projects WHERE 'seed' = ANY(tags)`,
		`DELETE FROM tracks      WHERE description LIKE '%[seed]%'`,
	}
	for _, st := range stmts {
		if err := s.exec(st); err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) seed() error {
	for li, lang := range langs {
		// Videos come from the language's real Go course playlist, not the
		// curriculum topics, so each language shows ~17–44 genuine videos.
		videoIDs, err := s.seedVideos(lang)
		if err != nil {
			return fmt.Errorf("videos %s: %w", lang, err)
		}
		var articleIDs, quizIDs, problemIDs []string
		for _, t := range curriculum {
			// Article
			var aid string
			if err := s.pool.QueryRow(s.ctx,
				`INSERT INTO articles (title, slug, body_markdown, difficulty, tags, language)
				 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id::text`,
				t.Title.get(lang), fmt.Sprintf("%s-%s", t.Tag, lang), t.Article.get(lang),
				t.Difficulty, []string{"seed", "go", t.Tag}, lang,
			).Scan(&aid); err != nil {
				return fmt.Errorf("article %s/%s: %w", t.Tag, lang, err)
			}
			articleIDs = append(articleIDs, aid)
			s.articles++

			// Quiz (+ questions + options)
			qid, err := s.insertQuiz(t, lang)
			if err != nil {
				return fmt.Errorf("quiz %s/%s: %w", t.Tag, lang, err)
			}
			quizIDs = append(quizIDs, qid)
			s.quizzes++

			// Problem (+ test cases)
			pid, err := s.insertProblem(t, lang)
			if err != nil {
				return fmt.Errorf("problem %s/%s: %w", t.Tag, lang, err)
			}
			problemIDs = append(problemIDs, pid)
			s.problems++
		}

		// Roadmap track linking this language's content in order.
		if err := s.insertTrack(lang, li, videoIDs, articleIDs, quizIDs, problemIDs); err != nil {
			return fmt.Errorf("track %s: %w", lang, err)
		}

		// Projects for this language.
		for _, p := range projects {
			if err := s.insertProject(p, lang); err != nil {
				return fmt.Errorf("project %s/%s: %w", p.Tag, lang, err)
			}
		}
	}
	return nil
}

// seedVideos inserts every video from a language's playlist (falling back to
// the English one) and returns their ids in playlist order. Difficulty is
// bucketed by position; duration is unknown from a playlist page, so it is left
// at 0 (the UI simply omits the duration badge).
func (s *seeder) seedVideos(lang string) ([]string, error) {
	vids := videoPlaylists[lang]
	if len(vids) == 0 {
		vids = videoPlaylists["en"]
	}
	ids := make([]string, 0, len(vids))
	for i, v := range vids {
		diff := "beginner"
		switch {
		case i >= len(vids)*7/10:
			diff = "advanced"
		case i >= len(vids)*4/10:
			diff = "intermediate"
		}
		var id string
		if err := s.pool.QueryRow(s.ctx,
			`INSERT INTO videos (title, description, youtube_id, duration_seconds, difficulty, tags, language)
			 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id::text`,
			v.Title, videoDesc(lang, i+1), v.ID, 0, diff, []string{"seed", "go"}, lang,
		).Scan(&id); err != nil {
			return nil, fmt.Errorf("video %d: %w", i+1, err)
		}
		ids = append(ids, id)
		s.videos++
	}
	return ids, nil
}

func videoDesc(lang string, n int) string {
	switch lang {
	case "ru":
		return fmt.Sprintf("Видео-урок №%d курса Go.", n)
	case "uz":
		return fmt.Sprintf("Go kursining %d-video darsi.", n)
	case "ja":
		return fmt.Sprintf("Go コースのビデオレッスン %d。", n)
	default:
		return fmt.Sprintf("Go course — video lesson %d.", n)
	}
}

func (s *seeder) insertQuiz(t topic, lang string) (string, error) {
	var qid string
	if err := s.pool.QueryRow(s.ctx,
		`INSERT INTO quizzes (title, description, pass_threshold, difficulty, tags, language)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id::text`,
		t.Title.get(lang), t.Blurb.get(lang), 60, t.Difficulty, []string{"seed", "go", t.Tag}, lang,
	).Scan(&qid); err != nil {
		return "", err
	}
	for qi, q := range t.Quiz {
		var questionID string
		if err := s.pool.QueryRow(s.ctx,
			`INSERT INTO quiz_questions (quiz_id, prompt, type, position) VALUES ($1,$2,$3,$4) RETURNING id::text`,
			qid, q.Prompt.get(lang), q.Type, qi+1,
		).Scan(&questionID); err != nil {
			return "", err
		}
		for oi, o := range q.Options {
			if err := s.exec(
				`INSERT INTO quiz_options (question_id, text, is_correct, position) VALUES ($1,$2,$3,$4)`,
				questionID, o.Text.get(lang), o.Correct, oi+1,
			); err != nil {
				return "", err
			}
		}
	}
	return qid, nil
}

func (s *seeder) insertProblem(t topic, lang string) (string, error) {
	var pid string
	if err := s.pool.QueryRow(s.ctx,
		`INSERT INTO problems (title, slug, statement_markdown, reference_solution_markdown, sample_io, difficulty, tags, language)
		 VALUES ($1,$2,$3,$4,$5::jsonb,$6,$7,$8) RETURNING id::text`,
		t.Prob.Title.get(lang), fmt.Sprintf("%s-%s", t.Prob.Slug, lang),
		t.Prob.Statement.get(lang), t.Prob.Solution, t.Prob.sampleJSON(),
		t.Difficulty, []string{"seed", "go", t.Tag}, lang,
	).Scan(&pid); err != nil {
		return "", err
	}
	for ci, c := range t.Prob.Cases {
		if err := s.exec(
			`INSERT INTO problem_test_cases (problem_id, input, expected_output, is_sample, position) VALUES ($1,$2,$3,$4,$5)`,
			pid, c.In, c.Out, c.Sample, ci+1,
		); err != nil {
			return "", err
		}
	}
	return pid, nil
}

func (s *seeder) insertProject(p project, lang string) error {
	var pid string
	if err := s.pool.QueryRow(s.ctx,
		`INSERT INTO mini_projects (title, description_markdown, difficulty, tags, language)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id::text`,
		p.Title.get(lang), p.Desc.get(lang), p.Difficulty, []string{"seed", "go", p.Tag}, lang,
	).Scan(&pid); err != nil {
		return err
	}
	for si, step := range p.Steps {
		if err := s.exec(
			`INSERT INTO mini_project_steps (project_id, text, position) VALUES ($1,$2,$3)`,
			pid, step.get(lang), si+1,
		); err != nil {
			return err
		}
	}
	s.projects++
	return nil
}

func (s *seeder) insertTrack(lang string, position int, videoIDs, articleIDs, quizIDs, problemIDs []string) error {
	var tid string
	title := map[string]string{
		"ru": "Go: путь обучения", "en": "Go learning roadmap",
		"uz": "Go oʻrganish yoʻli", "ja": "Go 学習ロードマップ",
	}[lang]
	if err := s.pool.QueryRow(s.ctx,
		`INSERT INTO tracks (title, description, level, language, position)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id::text`,
		title, "[seed] roadmap", "beginner", lang, position+1,
	).Scan(&tid); err != nil {
		return err
	}
	pos := 1
	add := func(ct, id string) error {
		err := s.exec(
			`INSERT INTO track_items (track_id, content_type, content_id, position) VALUES ($1,$2,$3::uuid,$4)`,
			tid, ct, id, pos)
		pos++
		return err
	}
	// Interleave per topic: article -> video -> quiz -> problem. There may be
	// fewer playlist videos than topics, so only link a video when one exists.
	for i := range articleIDs {
		if err := add("article", articleIDs[i]); err != nil {
			return err
		}
		if i < len(videoIDs) {
			if err := add("video", videoIDs[i]); err != nil {
				return err
			}
		}
		if err := add("quiz", quizIDs[i]); err != nil {
			return err
		}
		if err := add("problem", problemIDs[i]); err != nil {
			return err
		}
	}
	s.tracks++
	return nil
}
