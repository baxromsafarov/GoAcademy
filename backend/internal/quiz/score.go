// Package quiz scores quiz attempts and persists them.
package quiz

import "math"

// Question is the scoring view of a quiz question: its id and the set of correct
// option ids. (Type-specific input validation happens before scoring.)
type Question struct {
	ID         string
	CorrectIDs []string
}

// Submission maps a question id to the option ids the user selected.
type Submission map[string][]string

// QuestionResult is the per-question scoring outcome plus the correct answers
// (revealed after submission).
type QuestionResult struct {
	QuestionID string
	Correct    bool
	CorrectIDs []string
}

// ScoreResult is the overall outcome of scoring a submission.
type ScoreResult struct {
	Score   int // percent 0..100
	Results []QuestionResult
}

// Score grades a submission against the quiz's questions. A question is correct
// iff the set of selected options equals the set of correct options (this covers
// both single- and multiple-choice). The score is the percentage of correct
// questions, rounded to the nearest integer.
func Score(questions []Question, sub Submission) ScoreResult {
	results := make([]QuestionResult, 0, len(questions))
	correct := 0
	for _, q := range questions {
		ok := equalSet(sub[q.ID], q.CorrectIDs)
		if ok {
			correct++
		}
		results = append(results, QuestionResult{
			QuestionID: q.ID,
			Correct:    ok,
			CorrectIDs: q.CorrectIDs,
		})
	}

	score := 0
	if len(questions) > 0 {
		score = int(math.Round(float64(correct) * 100 / float64(len(questions))))
	}
	return ScoreResult{Score: score, Results: results}
}

// equalSet reports whether a and b contain the same distinct elements.
func equalSet(a, b []string) bool {
	sa, sb := toSet(a), toSet(b)
	if len(sa) != len(sb) {
		return false
	}
	for k := range sa {
		if _, ok := sb[k]; !ok {
			return false
		}
	}
	return true
}

func toSet(xs []string) map[string]struct{} {
	m := make(map[string]struct{}, len(xs))
	for _, x := range xs {
		m[x] = struct{}{}
	}
	return m
}
