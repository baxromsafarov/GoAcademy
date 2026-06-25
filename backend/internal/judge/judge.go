// Package judge grades a code submission against a problem's test cases using the
// sandbox runner, producing a verdict (OK/WA/TLE/RE/CE).
package judge

import (
	"context"
	"strings"

	"github.com/goacademy/backend/internal/runner"
)

// Verdict is the outcome of judging a submission (or a single case).
type Verdict string

const (
	OK  Verdict = "OK"  // accepted — output matched on every case
	WA  Verdict = "WA"  // wrong answer — output mismatch
	TLE Verdict = "TLE" // time limit exceeded
	RE  Verdict = "RE"  // runtime error — non-zero exit or out-of-memory
	CE  Verdict = "CE"  // compile error
)

// TestCase is one input/expected-output pair to judge against.
type TestCase struct {
	Input          string
	ExpectedOutput string
	IsSample       bool
}

// CaseResult is the per-case outcome. It deliberately omits the actual program
// output and the expected output so hidden test cases are never leaked.
type CaseResult struct {
	Index      int     `json:"index"`
	IsSample   bool    `json:"is_sample"`
	Verdict    Verdict `json:"verdict"`
	DurationMs int64   `json:"duration_ms"`
}

// Result is the overall judging outcome.
type Result struct {
	Verdict      Verdict      `json:"verdict"`
	Passed       int          `json:"passed"`
	Total        int          `json:"total"`
	CompileError string       `json:"compile_error,omitempty"`
	Cases        []CaseResult `json:"cases"`
}

// Judge grades submissions with a sandbox runner.
type Judge struct {
	runner *runner.Runner
	limits runner.Limits
}

// New builds a Judge using the given runner and per-case limits (zero-value
// limits fall back to the runner defaults).
func New(r *runner.Runner, limits runner.Limits) *Judge {
	return &Judge{runner: r, limits: limits}
}

// Run grades code against cases (which must be non-empty). The source is compiled
// once and run against every case input; the overall verdict is the first failing
// case's verdict, or OK when all pass. A returned error is an infrastructure
// failure, not a verdict.
func (j *Judge) Run(ctx context.Context, code string, cases []TestCase) (Result, error) {
	stdins := make([]string, len(cases))
	for i, c := range cases {
		stdins[i] = c.Input
	}

	results, err := j.runner.RunBatch(ctx, code, stdins, j.limits)
	if err != nil {
		return Result{}, err
	}
	if len(results) > 0 && results[0].CompileError {
		return Result{Verdict: CE, Total: len(cases), CompileError: results[0].Stderr, Cases: []CaseResult{}}, nil
	}

	out := Result{Verdict: OK, Total: len(cases), Cases: make([]CaseResult, 0, len(cases))}
	for i, res := range results {
		v := classify(res, cases[i].ExpectedOutput)
		if v == OK {
			out.Passed++
		} else if out.Verdict == OK {
			out.Verdict = v // first failing case decides the overall verdict
		}
		out.Cases = append(out.Cases, CaseResult{
			Index:      i,
			IsSample:   cases[i].IsSample,
			Verdict:    v,
			DurationMs: res.Duration.Milliseconds(),
		})
	}
	return out, nil
}

// classify turns a single run result + expected output into a verdict.
func classify(res runner.Result, expected string) Verdict {
	switch {
	case res.TimedOut:
		return TLE
	case res.OOMKilled:
		return RE
	case res.ExitCode != 0:
		return RE
	case normalize(res.Stdout) == normalize(expected):
		return OK
	default:
		return WA
	}
}

// normalize makes output comparison forgiving of trailing whitespace and line-end
// differences: CRLF/CR → LF, trailing spaces/tabs per line removed, and trailing
// blank lines dropped. Internal structure is preserved.
func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}
