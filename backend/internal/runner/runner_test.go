package runner

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// dockerReady reports whether a usable Docker daemon and the sandbox image are
// present. Runner tests are skipped otherwise so the suite stays green on hosts
// without Docker.
func dockerReady(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping docker-backed runner test in -short mode")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not on PATH")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker daemon not available")
	}
	if err := exec.Command("docker", "image", "inspect", "busybox").Run(); err != nil {
		t.Skip("busybox image not present (docker pull busybox)")
	}
}

func newTestRunner() *Runner { return New("busybox", "") }

func run(t *testing.T, req Request) Result {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	res, err := newTestRunner().Run(ctx, req)
	if err != nil {
		t.Fatalf("Run infra error: %v", err)
	}
	return res
}

func TestRunner_HelloAndExit(t *testing.T) {
	dockerReady(t)
	res := run(t, Request{Source: `package main
import "fmt"
func main(){ fmt.Print("hello") }`})
	if res.CompileError {
		t.Fatalf("unexpected compile error: %s", res.Stderr)
	}
	if res.Stdout != "hello" {
		t.Errorf("stdout = %q, want %q", res.Stdout, "hello")
	}
	if res.ExitCode != 0 {
		t.Errorf("exit = %d, want 0", res.ExitCode)
	}
	if res.TimedOut || res.OOMKilled {
		t.Errorf("unexpected sandbox flags: %+v", res)
	}
}

func TestRunner_Stdin(t *testing.T) {
	dockerReady(t)
	res := run(t, Request{
		Source: `package main
import ("bufio";"fmt";"os")
func main(){ s:=bufio.NewScanner(os.Stdin); s.Scan(); fmt.Print("got:"+s.Text()) }`,
		Stdin: "from-stdin\n",
	})
	if res.Stdout != "got:from-stdin" {
		t.Errorf("stdout = %q, want %q", res.Stdout, "got:from-stdin")
	}
}

func TestRunner_NonZeroExit(t *testing.T) {
	dockerReady(t)
	res := run(t, Request{Source: `package main
import "os"
func main(){ os.Exit(3) }`})
	if res.ExitCode != 3 {
		t.Errorf("exit = %d, want 3", res.ExitCode)
	}
}

func TestRunner_CompileError(t *testing.T) {
	dockerReady(t)
	res := run(t, Request{Source: `package main
func main(){ thisIsNotDefined() }`})
	if !res.CompileError {
		t.Fatalf("expected CompileError, got %+v", res)
	}
	if !strings.Contains(res.Stderr, "undefined") {
		t.Errorf("compile stderr = %q, want it to mention 'undefined'", res.Stderr)
	}
}

func TestRunner_NonStdlibImportFails(t *testing.T) {
	dockerReady(t)
	// GOPROXY=off means anything outside the stdlib cannot be fetched: the
	// sandbox is stdlib-only by design.
	res := run(t, Request{Source: `package main
import _ "github.com/some/package"
func main(){}`})
	if !res.CompileError {
		t.Errorf("expected non-stdlib import to fail to build, got %+v", res)
	}
}

func TestRunner_Timeout(t *testing.T) {
	dockerReady(t)
	res := run(t, Request{
		Source: `package main
func main(){ for {} }`,
		Limits: Limits{WallTime: 2 * time.Second},
	})
	if !res.TimedOut {
		t.Fatalf("expected TimedOut, got %+v", res)
	}
	if res.Duration > 6*time.Second {
		t.Errorf("timeout took too long: %s", res.Duration)
	}
}

func TestRunner_OOM(t *testing.T) {
	dockerReady(t)
	res := run(t, Request{
		Source: `package main
func main(){
	var chunks [][]byte
	for i := 0; i < 4096; i++ { b := make([]byte, 1<<20); for j := range b { b[j] = byte(j) }; chunks = append(chunks, b) }
	_ = chunks
}`,
		Limits: Limits{Memory: "64m", WallTime: 10 * time.Second},
	})
	if !res.OOMKilled {
		t.Fatalf("expected OOMKilled, got %+v (exit=%d)", res, res.ExitCode)
	}
}

func TestRunner_LargeOutputTruncated(t *testing.T) {
	dockerReady(t)
	res := run(t, Request{
		Source: `package main
import "os"
func main(){ b := make([]byte, 1024); for i := range b { b[i]='x' }; for i:=0;i<200;i++ { os.Stdout.Write(b) } }`,
		Limits: Limits{OutputBytes: 4096, WallTime: 10 * time.Second},
	})
	if !res.StdoutTruncated {
		t.Fatalf("expected StdoutTruncated, got len=%d", len(res.Stdout))
	}
	if len(res.Stdout) > 4096 {
		t.Errorf("stdout len = %d, want <= 4096", len(res.Stdout))
	}
}

func TestRunner_NoNetwork(t *testing.T) {
	dockerReady(t)
	// The program runs; the dial must fail because the container has no network.
	res := run(t, Request{
		Source: `package main
import ("fmt";"net";"time")
func main(){ _, err := net.DialTimeout("tcp","1.1.1.1:80",3*time.Second); fmt.Println("dialErrIsNil:", err==nil) }`,
		Limits: Limits{WallTime: 8 * time.Second},
	})
	if res.TimedOut {
		t.Fatalf("network test timed out: %+v", res)
	}
	if !strings.Contains(res.Stdout, "dialErrIsNil: false") {
		t.Errorf("expected dial to fail (no network); stdout = %q", res.Stdout)
	}
}
