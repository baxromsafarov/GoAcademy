package judge

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/goacademy/backend/internal/runner"
)

// --- pure unit tests (no Docker) ---

func TestNormalize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"5\n", "5"},
		{"5", "5"},
		{"5   \n", "5"},
		{"a\r\nb\r\n", "a\nb"},
		{"x\n\n\n", "x"},
		{"line1 \nline2\t\n", "line1\nline2"},
		{"", ""},
	}
	for _, c := range cases {
		if got := normalize(c.in); got != c.want {
			t.Errorf("normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name     string
		res      runner.Result
		expected string
		want     Verdict
	}{
		{"match", runner.Result{Stdout: "5\n"}, "5", OK},
		{"match-trailing-ws", runner.Result{Stdout: "5  "}, "5\n", OK},
		{"mismatch", runner.Result{Stdout: "6"}, "5", WA},
		{"timeout", runner.Result{TimedOut: true}, "5", TLE},
		{"oom", runner.Result{OOMKilled: true}, "5", RE},
		{"nonzero-exit", runner.Result{Stdout: "5", ExitCode: 1}, "5", RE},
	}
	for _, c := range cases {
		if got := classify(c.res, c.expected); got != c.want {
			t.Errorf("%s: classify = %q, want %q", c.name, got, c.want)
		}
	}
}

// --- integration tests (real sandbox runner) ---

func dockerReady(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping docker-backed judge test in -short mode")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not on PATH")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker daemon not available")
	}
	if err := exec.Command("docker", "image", "inspect", "busybox").Run(); err != nil {
		t.Skip("busybox image not present")
	}
}

func newTestJudge() *Judge {
	return New(runner.New("busybox", ""), runner.Limits{WallTime: 3 * time.Second})
}

const sumProgram = `package main
import ("bufio";"fmt";"os")
func main(){ r := bufio.NewReader(os.Stdin); var a, b int; fmt.Fscan(r, &a, &b); fmt.Println(a+b) }`

func sumCases() []TestCase {
	return []TestCase{
		{Input: "2 3\n", ExpectedOutput: "5", IsSample: true},
		{Input: "10 20\n", ExpectedOutput: "30"},
		{Input: "-1 1\n", ExpectedOutput: "0"},
	}
}

func judge(t *testing.T, code string, cases []TestCase) Result {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	res, err := newTestJudge().Run(ctx, code, cases)
	if err != nil {
		t.Fatalf("judge infra error: %v", err)
	}
	return res
}

func TestJudge_Accepted(t *testing.T) {
	dockerReady(t)
	res := judge(t, sumProgram, sumCases())
	if res.Verdict != OK {
		t.Fatalf("verdict = %s, want OK (%+v)", res.Verdict, res)
	}
	if res.Passed != 3 || res.Total != 3 {
		t.Errorf("passed/total = %d/%d, want 3/3", res.Passed, res.Total)
	}
}

func TestJudge_WrongAnswer(t *testing.T) {
	dockerReady(t)
	// Prints a*b instead of a+b: case "2 3" → 6 ≠ 5.
	wrong := `package main
import ("bufio";"fmt";"os")
func main(){ r := bufio.NewReader(os.Stdin); var a, b int; fmt.Fscan(r, &a, &b); fmt.Println(a*b) }`
	res := judge(t, wrong, sumCases())
	if res.Verdict != WA {
		t.Fatalf("verdict = %s, want WA (%+v)", res.Verdict, res)
	}
}

func TestJudge_RuntimeError(t *testing.T) {
	dockerReady(t)
	re := `package main
import "os"
func main(){ os.Exit(1) }`
	res := judge(t, re, sumCases())
	if res.Verdict != RE {
		t.Fatalf("verdict = %s, want RE (%+v)", res.Verdict, res)
	}
}

func TestJudge_TimeLimit(t *testing.T) {
	dockerReady(t)
	tle := `package main
func main(){ for {} }`
	res := judge(t, tle, sumCases()[:1]) // one case is enough
	if res.Verdict != TLE {
		t.Fatalf("verdict = %s, want TLE (%+v)", res.Verdict, res)
	}
}

func TestJudge_CompileError(t *testing.T) {
	dockerReady(t)
	ce := `package main
func main(){ notDefined() }`
	res := judge(t, ce, sumCases())
	if res.Verdict != CE {
		t.Fatalf("verdict = %s, want CE (%+v)", res.Verdict, res)
	}
	if res.CompileError == "" {
		t.Error("expected non-empty CompileError message")
	}
}
