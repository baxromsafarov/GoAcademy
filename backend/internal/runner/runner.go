// Package runner executes untrusted Go source in a locked-down, network-less,
// resource-limited Docker container and returns its captured output.
//
// Isolation model (D-027): the untrusted source is cross-compiled on the host
// with CGO disabled and the module proxy off (the Go compiler does not execute
// the code it compiles, so building untrusted source is safe; only stdlib
// programs build because GOPROXY=off blocks fetching anything else). The
// resulting static linux binary is copied into a one-shot container that has no
// network, all capabilities dropped, no-new-privileges, a non-root user, and
// memory/CPU/PID limits, then run under a wall-clock timeout with its output
// capped. Compiling on the host (rather than inside the sandbox) keeps runs fast
// (~sub-second vs ~10s cold) and avoids the compiler itself being OOM-killed by
// the container memory limit.
package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// maxSourceBytes caps accepted source size.
	maxSourceBytes = 64 * 1024
	// buildTimeout bounds the host cross-compile step.
	buildTimeout = 20 * time.Second
	// goModule is the throwaway module written next to the source.
	goModule = "module sandbox\n\ngo 1.25\n"
)

// Limits constrains a single run.
type Limits struct {
	WallTime    time.Duration // total execution budget for the program
	Memory      string        // docker --memory value, e.g. "128m"
	CPUs        string        // docker --cpus value, e.g. "0.5"
	Pids        int           // docker --pids-limit value
	OutputBytes int           // cap per stream (stdout, stderr)
}

// DefaultLimits returns conservative limits suitable for a learning sandbox.
func DefaultLimits() Limits {
	return Limits{
		WallTime:    5 * time.Second,
		Memory:      "128m",
		CPUs:        "0.5",
		Pids:        64,
		OutputBytes: 64 * 1024,
	}
}

func (l Limits) withDefaults() Limits {
	d := DefaultLimits()
	if l.WallTime <= 0 {
		l.WallTime = d.WallTime
	}
	if l.Memory == "" {
		l.Memory = d.Memory
	}
	if l.CPUs == "" {
		l.CPUs = d.CPUs
	}
	if l.Pids <= 0 {
		l.Pids = d.Pids
	}
	if l.OutputBytes <= 0 {
		l.OutputBytes = d.OutputBytes
	}
	return l
}

// Request is one execution request.
type Request struct {
	Source string
	Stdin  string
	Limits Limits
}

// Result is the outcome of a run. A non-zero ExitCode is a normal result (the
// program chose to exit non-zero); CompileError/TimedOut/OOMKilled describe
// sandbox-level outcomes instead.
type Result struct {
	Stdout          string
	Stderr          string
	ExitCode        int
	CompileError    bool
	TimedOut        bool
	OOMKilled       bool
	StdoutTruncated bool
	StderrTruncated bool
	Duration        time.Duration
}

// Runner executes requests against a Docker daemon.
type Runner struct {
	docker   string // docker binary name/path
	image    string // minimal image to run the static binary in
	workRoot string // host directory for transient build dirs ("" = os.TempDir)
}

// New builds a Runner. image must be a minimal linux image able to exec a static
// binary (e.g. "busybox"); workRoot is where transient build dirs are created.
func New(image, workRoot string) *Runner {
	if image == "" {
		image = "busybox"
	}
	return &Runner{docker: "docker", image: image, workRoot: workRoot}
}

// Run compiles and executes req, returning the captured result. A returned error
// indicates an infrastructure failure (docker/host problem), not a problem with
// the user's program — those are reported in Result.
func (r *Runner) Run(ctx context.Context, req Request) (Result, error) {
	if len(req.Source) > maxSourceBytes {
		return Result{}, fmt.Errorf("source exceeds %d bytes", maxSourceBytes)
	}
	lim := req.Limits.withDefaults()

	dir, bin, compileErr, err := r.compile(ctx, req.Source)
	if dir != "" {
		defer os.RemoveAll(dir)
	}
	if err != nil {
		return Result{}, err
	}
	if compileErr != nil {
		return *compileErr, nil
	}
	return r.execOnce(ctx, bin, req.Stdin, lim)
}

// RunBatch compiles the source once and runs the resulting binary against each
// stdin in its own fresh sandbox container. It is used by the judge to run a
// submission across many test cases without recompiling per case. On a compile
// error every returned result carries CompileError; the results align 1:1 with
// stdins.
func (r *Runner) RunBatch(ctx context.Context, source string, stdins []string, lim Limits) ([]Result, error) {
	if len(source) > maxSourceBytes {
		return nil, fmt.Errorf("source exceeds %d bytes", maxSourceBytes)
	}
	lim = lim.withDefaults()

	dir, bin, compileErr, err := r.compile(ctx, source)
	if dir != "" {
		defer os.RemoveAll(dir)
	}
	if err != nil {
		return nil, err
	}

	out := make([]Result, len(stdins))
	if compileErr != nil {
		for i := range out {
			out[i] = *compileErr
		}
		return out, nil
	}
	for i, in := range stdins {
		res, err := r.execOnce(ctx, bin, in, lim)
		if err != nil {
			return nil, err
		}
		out[i] = res
	}
	return out, nil
}

// compile cross-compiles source on the host to a static linux/amd64 binary. On a
// build failure it returns a non-nil *Result with CompileError set; on success it
// returns the build dir and binary path. The caller must os.RemoveAll(dir) when
// dir != "".
func (r *Runner) compile(ctx context.Context, source string) (dir, bin string, compileErr *Result, err error) {
	dir, err = os.MkdirTemp(r.workRoot, "goacademy-run-")
	if err != nil {
		return "", "", nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(source), 0o600); err != nil {
		return dir, "", nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goModule), 0o600); err != nil {
		return dir, "", nil, err
	}

	bin = filepath.Join(dir, "prog")
	buildCtx, cancelBuild := context.WithTimeout(ctx, buildTimeout)
	defer cancelBuild()
	build := exec.CommandContext(buildCtx, "go", "build", "-trimpath", "-o", bin, ".")
	build.Dir = dir
	build.Env = append(os.Environ(),
		"GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0",
		"GO111MODULE=on", "GOFLAGS=-mod=mod", "GOPROXY=off",
	)
	var buildErr bytes.Buffer
	build.Stdout = io.Discard
	build.Stderr = &buildErr
	if err := build.Run(); err != nil {
		// A build failure is a user-facing compile error, not an infra error,
		// unless the build context itself was cancelled by the caller.
		if ctx.Err() != nil {
			return dir, "", nil, ctx.Err()
		}
		return dir, "", &Result{CompileError: true, Stderr: cleanBuildError(buildErr.String(), dir)}, nil
	}
	return dir, bin, nil, nil
}

// execOnce runs the already-compiled binary once in a fresh locked-down container.
func (r *Runner) execOnce(ctx context.Context, bin, stdin string, lim Limits) (Result, error) {
	cid, err := r.create(ctx, lim)
	if err != nil {
		return Result{}, err
	}
	defer r.remove(cid)

	if out, err := exec.CommandContext(ctx, r.docker, "cp", bin, cid+":/prog").CombinedOutput(); err != nil {
		return Result{}, fmt.Errorf("docker cp: %w: %s", err, out)
	}

	runCtx, cancelRun := context.WithTimeout(ctx, lim.WallTime)
	defer cancelRun()
	stdout := &cappedWriter{limit: lim.OutputBytes}
	stderr := &cappedWriter{limit: lim.OutputBytes}
	start := time.Now()
	run := exec.CommandContext(runCtx, r.docker, "start", "--attach", "--interactive", cid)
	run.Stdin = strings.NewReader(stdin)
	run.Stdout = stdout
	run.Stderr = stderr
	runErr := run.Run()
	dur := time.Since(start)

	res := Result{
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		StdoutTruncated: stdout.truncated,
		StderrTruncated: stderr.truncated,
		Duration:        dur,
	}

	if runCtx.Err() == context.DeadlineExceeded {
		res.TimedOut = true
		r.kill(cid) // the CommandContext only killed the CLI; stop the container too
		return res, nil
	}
	_ = runErr // a non-zero program exit surfaces via ExitCode below, not as an error

	ec, oom := r.inspect(cid)
	res.ExitCode = ec
	res.OOMKilled = oom
	return res, nil
}

// create starts a stopped, locked-down container and returns its id.
func (r *Runner) create(ctx context.Context, lim Limits) (string, error) {
	args := []string{
		"create",
		"--interactive", // keep STDIN open so `docker start -i` can feed the program
		"--network=none",
		"--memory=" + lim.Memory,
		"--memory-swap=" + lim.Memory, // equal swap disables swap entirely
		"--cpus=" + lim.CPUs,
		"--pids-limit=" + strconv.Itoa(lim.Pids),
		"--cap-drop=ALL",
		"--security-opt=no-new-privileges",
		"--user=65534:65534", // nobody
		"--tmpfs=/tmp:rw,size=16m",
		r.image,
		"/prog",
	}
	out, err := exec.CommandContext(ctx, r.docker, args...).Output()
	if err != nil {
		return "", fmt.Errorf("docker create: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// inspect returns the container's exit code and OOMKilled flag.
func (r *Runner) inspect(cid string) (int, bool) {
	out, err := exec.Command(r.docker, "inspect", cid, "--format", "{{.State.ExitCode}} {{.State.OOMKilled}}").Output()
	if err != nil {
		return 0, false
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) != 2 {
		return 0, false
	}
	code, _ := strconv.Atoi(fields[0])
	return code, fields[1] == "true"
}

func (r *Runner) kill(cid string) {
	_ = exec.Command(r.docker, "kill", cid).Run()
}

func (r *Runner) remove(cid string) {
	_ = exec.Command(r.docker, "rm", "-f", cid).Run()
}

// cleanBuildError strips the host build directory path from compiler messages so
// the sandbox's temp layout is not leaked to the user.
func cleanBuildError(msg, dir string) string {
	msg = strings.ReplaceAll(msg, dir+string(filepath.Separator), "")
	msg = strings.ReplaceAll(msg, dir, "")
	return strings.TrimSpace(msg)
}

// cappedWriter buffers up to limit bytes and then silently discards the rest,
// recording that truncation happened. It never errors, so the attached pipe is
// always drained (a program that floods stdout cannot block on a full pipe).
type cappedWriter struct {
	limit     int
	buf       bytes.Buffer
	truncated bool
}

func (w *cappedWriter) Write(p []byte) (int, error) {
	if remaining := w.limit - w.buf.Len(); remaining > 0 {
		if len(p) <= remaining {
			w.buf.Write(p)
		} else {
			w.buf.Write(p[:remaining])
			w.truncated = true
		}
	} else if len(p) > 0 {
		w.truncated = true
	}
	return len(p), nil
}

func (w *cappedWriter) String() string { return w.buf.String() }
