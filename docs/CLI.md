**CLI Reference**

This reference documents the `rig` CLI — global flags, important commands, and common workflows.

Cross-links:
- Configuration reference: `docs/CONFIGURATION.md`
- Examples: `examples/`

---

## Global/Shared flags

Many `rig` subcommands share a small set of useful flags. Where a flag is only applicable to a subcommand it is noted in that command's section.

- `-C, --dir <path>`: set the working directory for the command (applies to `run`, `build`, `x`).
- `-n, --dry-run`: print what `rig` would execute without running it (applies to `run`, `build`, `x`).
- `-E, --env KEY=VALUE`: provide environment variables (can be repeated) merged with task env.
- `-l, --list`: for `rig run` — list tasks instead of running.
- `-j, --json`: machine-readable structured output where supported (e.g., `run --list --json`, `tools outdated --json`, `tools sync --check --json`).
- `--check`: check-only (verify state without making changes), used in `tools sync --check` and `rig setup --check`.

Note: `rig` ensures `.rig/bin` is prepended once to `PATH` for commands that install or run local tools.

---

## Commands

Top-level commands are implemented in `internal/cli/*.go`.

### `rig init`

Create a new `rig.toml` in the current directory with sensible defaults or interactive prompts.

Common flags:
- `--yes, -y`: accept defaults non-interactively.
- `--developer`: generate developer-focused template (linting, watchers, dev profile).
- `--minimal`: minimal template.
- `--monorepo`: generate `.rig/` includes for monorepo layouts.
- `--dev-watcher {none|reflex|air}`: pick a file-watcher task for developer experience.
- `--ci`: add a simple CI task.
- `-C, --dir <path>`: write manifest into a target directory.
- `--force`: overwrite existing `rig.toml`.

### `rig run <task>` (aliases: `rig r`, `rig ls`)

Run a named task from `[tasks]`.

Flags:
- `-l, --list`: list tasks (use `-j` for JSON list).
- `-j, --json`: valid only with `--list` to print deterministic JSON describing tasks.
- `-n, --dry-run`: print the resolved command(s) without executing.
- `-E, --env KEY=VALUE`: override/extend task env during run.
- `-C, --dir`: working directory for the executed command.

Behavior:
- Tasks may be simple strings or structured tables (see `docs/CONFIGURATION.md`).
- `depends_on` is supported and will be executed in deterministic topological order.

Examples:
```
rig run dev
rig run build --dry-run
rig run --list --json | jq .
```

### `rig build` (aliases: `rig b`)

Compose and run `go build` with optional `--profile` or CLI overrides.

Flags:
- `--profile <name>`: apply `[profile.<name>]` from `rig.toml`.
- `-o, --output <path>`: override profile output.
- `-t, --tags`: build tags.
- `--ldflags`, `--gcflags`: override profile flags.
- `-C, --dir`: working directory for the build.
- `-n, --dry-run`: show the composed `go build` command.

### `rig tools` / `rig sync` / `rig check` / `rig outdated` (aliases: `rig t`, `rig sync`)

Manage tools declared in `[tools]`.

- `rig tools sync` (alias `rig sync`): install or update tools into `.rig/bin`.
	- `--check`: do not install; verify lockfile parity.
	- `--json` (only valid with `--check`): print machine-readable summary for CI.
	- `--offline`: do not download modules (sets `GOPROXY=off`, `GOSUMDB=off`).

- `rig tools check`: convenience alias for `rig tools sync --check`.

- `rig tools outdated` (alias `rig outdated`): list missing or mismatched tools.
	- `--json`: print machine-readable list.

Notes:
- `rig sync` writes `rig.lock` (schema=0) next to `rig.toml`. This is the source of truth for resolved tool versions.
- `rig sync` also writes `.rig/manifest.lock` (a hash cache) used for fast parity checks.
- `rig sync --check` validates that `rig.lock` matches `rig.toml` and that `.rig/bin` matches `rig.lock`.
- For CI prefer `rig sync --check --json`.

### `rig setup`

Convenience installer for `[tools]` (similar semantics to `rig tools sync`).

Flags:
- `--check`: verify installed tool versions instead of installing.

### `rig doctor`

Sanity-check local environment: verifies `go` installation and, if a `rig.toml` exists, checks pinned tools in `.rig/bin`.

### `rig x` (ephemeral execution)

Ephemeral runner similar to `npx`/`bunx` — install a specified tool (by short name or module) into `.rig/bin` if missing, then run it immediately.

Usage:
```
rig x <tool[@version]|module[@version]> [-- args...]
```

---
