**CLI Reference (v0.3)**

This document describes the stable, daily workflow commands:

`init → sync → dev → check`.

See also:
- [docs/CONFIGURATION.md](./CONFIGURATION.md) for `rig.toml` schema
- [docs/INSTALLATION.md](./INSTALLATION.md) for install options

---

## Alias model (final)

`rig` ships as one binary. Behavior is selected by invocation name (argv[0]).

Reserved entrypoints:
- `rig` → main CLI
- `rir` → `rig run`
- `ric` → `rig check`
- `rid` → `rig dev`
- `ris` → `rig start` (stub / future)

Use `rig alias` for the canonical explanation.

---

## Commands

### `rig run <task>` (alias: `rir`)

Runs a named task from `[tasks]`.

- Requires `rig.lock`.
- Validates tools in `.rig/bin` against `rig.lock` before executing.
- Supports `depends_on` with deterministic ordering and cycle detection.
- Arguments after `--` are passed only to the root task.

Examples:
```
rig run --list
rig run test
rig run test -- -count=1
```

### `rig dev` (alias: `rid`)

Runs the long-lived dev loop: execute `[tasks.dev].command` and restart on changes.

Requirements:
- `rig.lock` must exist (run `rig sync` first).
- A watcher tool must be pinned in `[tools]` and installed into `.rig/bin`.
  - v0.3 watcher: `reflex`

Signals:
- `SIGINT` (Ctrl+C) triggers a restart.
- `SIGTERM` exits.

Example config:
```toml
[tasks.dev]
command = "go run ."
watch = ["**/*.go"]

[tools]
github.com/cespare/reflex = "latest"
```

### `rig check` (alias: `ric`)

Verifies that:
- `rig.lock` exists
- tools in `.rig/bin` match the lock
- Go toolchain requirements (if pinned) match the lock

Output:
- Always prints stable JSON to stdout.
- Exits non-zero if the check fails.

### `rig status`

Read-only overview of current state:
- config path
- lock presence and parity
- tool counts (missing/mismatched/extras)
- Go toolchain status (if applicable)

### `rig start` (alias: `ris`)

Stubbed for future releases. Currently returns “not implemented”.
