**Running Rig in Production**

This document explains recommended patterns for using `rig` in production environments (containers, process supervisors, and CI runners). It covers how `rig` executes tasks, how to integrate with Docker, and best practices for signal handling and observability.

---

## How Rig executes tasks

`rig` executes tasks using the `internal/rig` executor which provides two execution modes:

- `Execute(name, args, ExecOptions)`: run a binary directly (no shell), streaming `stdout`/`stderr` to the parent process.
- `ExecuteShell(command, ExecOptions)`: run a string command via the platform shell (`sh -c` on Unix, `cmd /c` on Windows), streaming stdio.

Important characteristics for production:
- `rig` streams child's `stdout` and `stderr` directly to its own `stdout`/`stderr` (no buffering). This keeps logs real-time and compatible with Docker / journald.
- Environment and working directory can be controlled via `ExecOptions.Dir` and `ExecOptions.Env`.

Note: `rig` does not currently implement advanced logging or structured JSON emission for every task; you can wrap task commands with your own JSON logger or use `rig x` to run pinned tools that emit structured logs.

---

## Signals and supervision (PID 1)

`rig` is designed to be a process runner for tasks, but it does not implement a full process-supervisor layer (no automatic reaping of orphaned children beyond the normal Go `exec.Cmd` behavior). Key points:

- When `rig` runs a command it uses `exec.Cmd` and connects `stdin/stdout/stderr` to the child.
- `exec.Cmd.Run()` blocks until the child exits and returns the child's exit code.
- `rig` does not currently perform explicit `signal.Notify` forwarding (SIGINT/SIGTERM) nor does it reparent children with a dedicated reaper loop in the codebase. Because of this you should be careful running `rig` directly as PID 1 inside a container.

Recommendations:
- In container environments, run `rig` under a minimal init system (recommended):
  - Use `tini` or `dumb-init` as PID 1, and make `rig` a child process. These tiny init systems forward signals correctly and reap zombies.
  - Example (Docker): `ENTRYPOINT ["/sbin/tini", "--", "rig", "start"]`.
- If you must run `rig` as PID 1, consider wrapping your task commands with a small shell script that traps signals and forwards them to the child processes.

---

## Observability and logging

- `rig` streams stdout/stderr of tasks directly — logs from tasks will appear in the container logs as-is.
- For structured logs consider running your service under a structured-logging wrapper or use tools that output JSON.
- Use `rig check` (always prints stable JSON) and `rig tools sync --check --json` / `rig tools outdated --json` to drive automation and monitoring checks from CI systems.

---

## Docker integration

A simple Dockerfile pattern using `tini` and `rig`:

```dockerfile
# Build stage
FROM golang:1.21 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/rig ./cmd/rig

# Runtime stage
FROM debian:bookworm-slim
# Install tini
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tini && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/rig /usr/local/bin/rig
WORKDIR /app
# Copy project files or mount at runtime
COPY . .

# Recommended: run under tini so signals are forwarded and zombies reaped
ENTRYPOINT ["/usr/bin/tini", "--", "rig", "run", "start"]
# Or run a configured task directly
# ENTRYPOINT ["/usr/bin/tini", "--", "rig", "run", "server"]
```

Replace `start` or `server` with the appropriate task for your project.

---

## Example: minimal production run

1. Build a release binary via CI using the release profile:

```pwsh
rig build --profile release
```

2. Create a minimal container image that only contains your compiled binary and a tiny init like `tini`.

3. Run the container with an init system to ensure signal forwarding.

---

## Caveats and future work

- If your project needs advanced supervision (auto-restart, healthchecks, concurrency limiting), run `rig` under a true process supervisor (systemd, supervisor, Kubernetes). `rig` is intended primarily as a developer- and CI-oriented task orchestrator and build tool.

- If you want `rig` to act as a full supervisor (PID 1) with proper signal forwarding and reaping behavior, we can add a small signal-forwarding loop to the project that registers `signal.Notify` and forwards signals to the active child process; I can draft this change if you'd like.
