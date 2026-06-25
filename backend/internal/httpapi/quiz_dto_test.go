package httpapi

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/goacademy/backend/internal/content"
	"github.com/goacademy/backend/internal/store"
)

// TestQuizDetailResponse_OmitsCorrectAnswers is the answer-leak guard: the quiz
// detail response must never serialize the is_correct flag.
func TestQuizDetailResponse_OmitsCorrectAnswers(t *testing.T) {
	detail := content.QuizDetail{
		Quiz: store.Quiz{
			Title:         "Sample",
			PassThreshold: 70,
			Difficulty:    store.DifficultyBeginner,
			Language:      store.LocaleEn,
		},
		Questions: []content.QuizQuestionDetail{
			{
				Question: store.QuizQuestion{Prompt: "2+2?", Type: store.QuizQuestionTypeSingle},
				Options: []store.QuizOption{
					{Text: "four", IsCorrect: true},
					{Text: "five", IsCorrect: false},
				},
			},
		},
	}

	raw, err := json.Marshal(toQuizDetailResponse(detail))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(raw)

	for _, leak := range []string{"is_correct", "IsCorrect", "isCorrect"} {
		if strings.Contains(body, leak) {
			t.Errorf("response leaks correctness via %q: %s", leak, body)
		}
	}
	// Sanity: option text is still present.
	if !strings.Contains(body, "four") || !strings.Contains(body, "five") {
		t.Errorf("option text missing from response: %s", body)
	}
}
