**CLI Reference (v0.4)**

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
- `ril` → `rig tools ls`
- `rip` → `rig tools path`
- `riw` → `rig tools why`
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

### `rig tools ls` (entrypoint alias: `ril`)

Lists tools from `rig.lock` in deterministic name order.

### `rig tools path <name>` (entrypoint alias: `rip`)

Prints the absolute path for a locked tool binary in `.rig/bin`.

Validation:
- tool exists in `rig.lock`
- binary exists in `.rig/bin`
- file checksum matches lock SHA256

### `rig tools why <name>` (entrypoint alias: `riw`)

Shows lock-backed provenance for one tool:
- requested
- resolved module@version
- sha256
- resolved binary path

### `rig tools doctor [name]`

Diagnoses tool health for all tools or one tool:
- present / missing
- executable bit
- sha256 parity

Deterministic output ordering is preserved.

### `rig status`

Read-only overview of current state:
- config path
- lock presence and parity
- tool counts (missing/mismatched/extras)
- Go toolchain status (if applicable)

### `rig doctor [name]`

- Without args: runs environment + toolchain doctor checks.
- With `<name>`: delegates to `rig tools doctor <name>`.

### `rig upgrade`

Self-updates the `rig` binary from GitHub Releases.

Behavior:
- Compares current build version to latest `tag_name`; if equal, prints up-to-date and exits.
- Selects asset by OS/arch:
  - Unix: `rig_<os>_<arch>.tar.gz`
  - Windows: `rig_windows_<arch>.zip`
- Requires a matching `<asset>.sha256` and verifies SHA256 before extraction.
- Requires archive contract: exactly one binary entry (`rig` or `rig.exe`).
- Replaces the current executable only; does not mutate `rig.toml`, `rig.lock`, PATH, aliases, or project config.
- Exits non-zero on any failure (network, checksum mismatch, unsupported platform, permission denied, extraction/replace errors).

Windows note:
- If replacement fails due to a running/locked executable, close active `rig` processes and retry.

### `rig start` (alias: `ris`)

Stubbed for future releases. Currently returns “not implemented”.
