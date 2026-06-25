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

// youTubePool is a small set of real Go YouTube video IDs. NOTE: sourcing 30
// genuine language-specific tutorials per language is out of scope, so videos
// reuse this pool — replace these IDs with real per-language videos as desired.
var youTubePool = []string{
	"YS4e4q9oBaU", // freeCodeCamp — Learn Go Programming
	"un6ZyFkqFKo", // Go in 100 seconds / intro
	"446E-r0rXHI", // Go tutorial
}

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
		var videoIDs, articleIDs, quizIDs, problemIDs []string
		for ti, t := range curriculum {
			// Video
			var vid string
			if err := s.pool.QueryRow(s.ctx,
				`INSERT INTO videos (title, description, youtube_id, duration_seconds, difficulty, tags, language)
				 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id::text`,
				t.Title.get(lang), t.Blurb.get(lang), youTubePool[ti%len(youTubePool)], 600+ti*30,
				t.Difficulty, []string{"seed", "go", t.Tag}, lang,
			).Scan(&vid); err != nil {
				return fmt.Errorf("video %s/%s: %w", t.Tag, lang, err)
			}
			videoIDs = append(videoIDs, vid)
			s.videos++

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
	// Interleave per topic: article -> video -> quiz -> problem.
	for i := range articleIDs {
		if err := add("article", articleIDs[i]); err != nil {
			return err
		}
		if err := add("video", videoIDs[i]); err != nil {
			return err
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
