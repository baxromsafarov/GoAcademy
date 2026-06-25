package progress

import "testing"

func TestSummarize(t *testing.T) {
	cases := []struct {
		name        string
		flags       []bool
		wantTotal   int
		wantDone    int
		wantPercent int
		wantDone100 bool
	}{
		{"empty track", nil, 0, 0, 0, false},
		{"none done", []bool{false, false, false}, 3, 0, 0, false},
		{"half done", []bool{true, false, true, false}, 4, 2, 50, false},
		{"two thirds", []bool{true, true, false}, 3, 2, 67, false}, // 66.67 -> 67
		{"all done", []bool{true, true}, 2, 2, 100, true},
		{"single done", []bool{true}, 1, 1, 100, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			total, done, percent, complete := summarize(tc.flags)
			if total != tc.wantTotal || done != tc.wantDone || percent != tc.wantPercent || complete != tc.wantDone100 {
				t.Errorf("summarize(%v) = (total=%d done=%d pct=%d complete=%v), want (%d %d %d %v)",
					tc.flags, total, done, percent, complete, tc.wantTotal, tc.wantDone, tc.wantPercent, tc.wantDone100)
			}
		})
	}
}
