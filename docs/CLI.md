**CLI Reference**

This reference documents the `rig` CLI — global flags, important commands, and the most common workflows.

Cross-links:
- Configuration reference: `docs/CONFIGURATION.md`
- Examples: `examples/`

---

## Global/Shared flags
Many `rig` subcommands share a small set of useful flags. Where a flag is only applicable to a subcommand it is noted in that command's section.

- `-C, --dir <path>`: set the working directory for the command (applies to `run`, `build`, `x`).
- `-n, --dry-run`: print what `rig` would execute without running it (applies to `run`, `build`, `x`).
- `-E, --env KEY=VALUE`: provide environment variables (can be repeated) merged with task env. Example: `-E FOO=bar -E BAZ=qux`.
- `-l, --list`: for `rig run` — list tasks instead of running.
- `-j, --json`: machine-readable structured output where supported (e.g., `run --list --json`, `tools outdated --json`, `tools sync --check --json`).
- `--check`: check-only (verify state without making changes), used in `tools sync --check` and `rig setup --check`.

Note: `rig` ensures `.rig/bin` is prepended once to `PATH` for commands that install or run local tools.

---

## Commands

Top-level commands implemented in `internal/cli/*.go`.

### `rig init`
Create a new `rig.toml` in the current directory with sensible defaults. Options:

- `--yes, -y`: accept defaults non-interactively.
- `--developer`: generate developer-focused template (linting, watchers, dev profile).
- `--minimal`: minimal template (release profile only).
- `--monorepo`: generate `.rig/` includes for monorepo layouts.
- `--dev-watcher {none|reflex|air}`: pick a file-watcher for developer experience.
- `--ci`: add a simple CI task.

What it writes:
- A root `rig.toml` with `[project]` and `profile` blocks.
- Optionally `.rig/rig.tasks.toml` and `.rig/rig.tools.toml` for monorepo layouts.

See `docs/CONFIGURATION.md` for the generated fields and structure.

### `rig run <task>` (aliases: `rig r`)
Run a named task from `[tasks]`.

Flags:
- `-l, --list`: list tasks (use `-j` for JSON list).
- `-j, --json`: valid only with `--list` to print deterministic JSON describing tasks.
- `-n, --dry-run`: print the resolved command(s) without executing.
- `-E, --env KEY=VALUE`: override/extend task env during run.
- `-C, --dir`: working directory for the executed command.

Behavior:
- Tasks may be simple strings or structured tables (see `docs/CONFIGURATION.md`).
- `depends_on` is supported and will be executed in deterministic topological order. Cycles are detected and reported.
- `rig run` will perform a fast `.rig` lock verification if tools are defined, unless run in list mode or with `--dry-run`.

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
- `-n, --dry-run`: show the composed go build command.

`rig build` uses `internal/rig.ComposeBuildCommand` to merge profile and CLI values.

### `rig tools` / `rig sync` / `rig check` / `rig outdated` (aliases: `rig t`, `rig sync`)
Manage declared tools in `[tools]`.

- `rig tools sync` (alias `rig sync`): install or update tools into `.rig/bin`.
  - `--check`: do not install; verify lockfile parity.
  - `--json` (only valid with `--check`): print machine-readable summary for CI.

- `rig tools check`: convenience alias for `rig tools sync --check`.

- `rig tools outdated` (alias `rig outdated`): list missing or mismatched tools.
  - `--json`: print machine-readable list. When no tools are declared the command returns `[]` in JSON mode.

Notes:
- `rig sync` writes a lockfile (e.g., `.rig/manifest.lock`) used by `rig run` to make run-time checks fast.
- For CI prefer `rig sync --check --json` to assert that tools are installed and locked.

### `rig setup`
Legacy helper that mirrors `rig tools sync` semantics but remains as a convenience command.

Flags:
- `--check`: verify installed tool versions instead of installing.

### `rig doctor`
Sanity-check local environment: verifies `go` installation and, if a `rig.toml` exists, checks pinned tools in `.rig/bin`.

Example:
```
rig doctor
```

### `rig x` (ephemeral execution)
Ephemeral runner similar to `npx`/`bunx` — install a specified tool (by short name or module) into `.rig/bin` if missing, then run it immediately.

Usage:
```
rig x <tool[@version]|module[@version]> [-- args...]
```

Flags:
- `--no-install`: fail if missing or mismatched instead of installing.
- `--dry-run`: show what would run.
- `-C, --dir`: working directory to run the tool in.
- `--env KEY=VALUE`: environment overrides.

Behavior:
- Resolves short names via internal mapping (see `internal/rig/tooling.go`).
- If version omitted `rig` will consult `[tools]` in the project, or default to `latest`.

---

## Quick workflows
- Developer first-run:
```
rig init --developer
rig sync
rig run dev
```

- CI validation (fail fast on tool mismatch):
```
rig sync --check --json | jq .
```

- Ephemeral use (run specific tool without adding to repo files):
```
rig x golangci-lint@v1.62.0 run ./...
```

---

If you'd like, I can also generate a small cheatsheet file (CLI quick reference) with common one-liners for your README or CONTRIBUTING guide.
