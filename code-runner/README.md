# code-runner

Sandbox for executing **untrusted Go** safely (D-027). The implementation lives
in the backend as the package [`backend/internal/runner`](../backend/internal/runner)
and is orchestrated in-process — the security boundary is the container, not a
service boundary, so a separate microservice would add operational complexity
without isolation benefit. This directory holds the design notes; a standalone
extraction (its own binary/HTTP service) can be lifted out later without changing
the model.

## Isolation model

1. **Compile on the host**, not in the sandbox. The untrusted source is
   cross-compiled with `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOPROXY=off`
   into a static binary. The Go compiler does not execute the code it compiles,
   so building untrusted source is safe; `GOPROXY=off` makes the sandbox
   **stdlib-only** (no module fetching). Host compilation keeps runs fast
   (~sub-second vs ~10s for a cold in-container build) and avoids the *compiler*
   being OOM-killed by the container's memory limit.
2. **Run the static binary in a one-shot container** with:
   - `--network=none` — no network at all
   - `--cap-drop=ALL --security-opt=no-new-privileges --user=65534:65534` — no
     capabilities, no privilege escalation, runs as `nobody`
   - `--memory` / `--memory-swap` (equal → swap disabled) — RAM cap; the
     `OOMKilled` flag is read back via `docker inspect`
   - `--cpus`, `--pids-limit` — CPU and fork limits
   - `--tmpfs=/tmp` — small writable scratch; the rest of the fs is ephemeral
     and discarded with the container
   - a **wall-clock timeout** enforced by the runner (the container is
     `docker kill`ed if it overruns)
   - **output capping** — stdout/stderr are read through a capped writer that
     never blocks the program
3. **Binary delivery**: `docker create -i` → `docker cp prog → /prog` →
   `docker start -a -i`. The rootfs is *not* `--read-only` because `docker cp`
   cannot write into a read-only rootfs; the remaining controls plus the
   throwaway container make this an acceptable trade.

## Run image

A minimal linux image able to exec a static binary — `busybox` by default
(`runner.New("busybox", workRoot)`). No Go toolchain is needed inside the
sandbox.

## Deployment note (CH18)

The host running the backend needs the **Go toolchain** (for cross-compilation)
and **Docker CLI + daemon access**. In a containerized deploy this means either
mounting the Docker socket (Docker-out-of-Docker) or running the runner on a host
with Docker, and bundling Go in that image. Hardening (seccomp profile, a
dedicated low-trust runner host, image pinning) is tracked for the security pass.
