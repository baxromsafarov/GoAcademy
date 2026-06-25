package httpapi

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/goacademy/backend/internal/store"
)

// TestProblemDetailResponse_HidesReferenceSolution is the leak guard: the problem
// detail response must never serialize the reference solution, but must keep the
// statement and sample I/O.
func TestProblemDetailResponse_HidesReferenceSolution(t *testing.T) {
	p := store.Problem{
		Title:                     "Two Sum",
		Slug:                      "two-sum",
		StatementMarkdown:         "Find two numbers that add up to the target.",
		ReferenceSolutionMarkdown: "SECRET-SOLUTION-CODE",
		SampleIo:                  []byte(`[{"input":"1 2","output":"3"}]`),
		Difficulty:                store.DifficultyBeginner,
		Language:                  store.LocaleEn,
	}

	raw, err := json.Marshal(toProblemDetailResponse(p))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(raw)

	for _, leak := range []string{"reference_solution", "ReferenceSolution", "SECRET-SOLUTION-CODE"} {
		if strings.Contains(body, leak) {
			t.Errorf("response leaks the reference solution via %q: %s", leak, body)
		}
	}
	if !strings.Contains(body, "Find two numbers") {
		t.Errorf("statement missing: %s", body)
	}
	if !strings.Contains(body, `"sample_io"`) || !strings.Contains(body, `"3"`) {
		t.Errorf("sample_io missing or not passed through: %s", body)
	}
}
