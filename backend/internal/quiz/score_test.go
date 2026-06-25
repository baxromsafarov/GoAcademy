package quiz

import "testing"

func TestScore(t *testing.T) {
	questions := []Question{
		{ID: "q1", CorrectIDs: []string{"a"}},      // single
		{ID: "q2", CorrectIDs: []string{"x", "y"}}, // multiple
		{ID: "q3", CorrectIDs: []string{"m"}},      // single
	}

	cases := []struct {
		name       string
		sub        Submission
		wantScore  int
		wantQ2True bool
	}{
		{
			name:      "all correct",
			sub:       Submission{"q1": {"a"}, "q2": {"x", "y"}, "q3": {"m"}},
			wantScore: 100, wantQ2True: true,
		},
		{
			name:      "multiple order-independent",
			sub:       Submission{"q1": {"a"}, "q2": {"y", "x"}, "q3": {"m"}},
			wantScore: 100, wantQ2True: true,
		},
		{
			name:      "multiple partial is wrong",
			sub:       Submission{"q1": {"a"}, "q2": {"x"}, "q3": {"m"}},
			wantScore: 67, wantQ2True: false, // 2/3 -> 66.67 -> 67
		},
		{
			name:      "multiple extra is wrong",
			sub:       Submission{"q1": {"a"}, "q2": {"x", "y", "z"}, "q3": {"m"}},
			wantScore: 67, wantQ2True: false,
		},
		{
			name:      "single wrong answer",
			sub:       Submission{"q1": {"b"}, "q2": {"x", "y"}, "q3": {"m"}},
			wantScore: 67, wantQ2True: true,
		},
		{
			name:      "empty submission scores zero",
			sub:       Submission{},
			wantScore: 0, wantQ2True: false,
		},
		{
			name:      "duplicates in selection do not change the set",
			sub:       Submission{"q1": {"a", "a"}, "q2": {"x", "x", "y"}, "q3": {"m"}},
			wantScore: 100, wantQ2True: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := Score(questions, tc.sub)
			if res.Score != tc.wantScore {
				t.Errorf("score = %d, want %d", res.Score, tc.wantScore)
			}
			if len(res.Results) != len(questions) {
				t.Fatalf("results = %d, want %d", len(res.Results), len(questions))
			}
			// q2 is the second result.
			if res.Results[1].Correct != tc.wantQ2True {
				t.Errorf("q2 correct = %v, want %v", res.Results[1].Correct, tc.wantQ2True)
			}
			// Correct ids are always revealed.
			if len(res.Results[1].CorrectIDs) != 2 {
				t.Errorf("q2 correct ids = %v, want 2", res.Results[1].CorrectIDs)
			}
		})
	}
}

func TestScore_NoQuestions(t *testing.T) {
	res := Score(nil, Submission{})
	if res.Score != 0 {
		t.Errorf("score = %d, want 0 for no questions", res.Score)
	}
}
