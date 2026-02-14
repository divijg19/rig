# Changelog

All notable changes to this project are documented in this file.

The format is based on Keep a Changelog, and this project aims to follow Semantic Versioning.

## Documentation
- Expanded command documentation for v0.4 command surface (`rig tools ls/path/why/doctor`, `rig doctor [name]`, `rig upgrade`).
- Aligned installation and alias docs with invocation-name entrypoints (`ril`, `rip`, `riw`) and current installer behavior.
- Added explicit self-upgrade safety notes: binary-only scope, up-to-date short-circuit, checksum verification, and Windows retry guidance.

## v0.4

### Added
- Tool observability commands:
	- `rig tools ls`
	- `rig tools path <name>`
	- `rig tools why <name>`
	- `rig tools doctor [name]`
- Top-level shortcuts for observability:
	- `rig ls` (delegates to `rig tools ls`)
	- `rig path <name>` (delegates to `rig tools path`)
	- `rig why <name>` (delegates to `rig tools why`)
- Invocation-name aliases:
	- `ril` → `rig tools ls`
	- `rip` → `rig tools path`
	- `riw` → `rig tools why`
- `rig upgrade` command for self-updating the current `rig` executable.
- New core diagnostics/reporting:
	- `internal/rig/tools_inspect.go`
	- `internal/rig/doctor.go`
	- `internal/rig/upgrade.go`
- New observability + upgrade test coverage in `internal/rig/observability_upgrade_test.go`.

### Changed
- `rig doctor` now accepts optional `name` and dispatches to per-tool diagnostics when provided.
- Root CLI help text now documents observability and self-upgrade features.
- Self-upgrade behavior now enforces:
	- latest release tag lookup via GitHub Releases API,
	- OS/arch asset selection,
	- SHA256 verification via `<asset>.sha256`,
	- strict archive contract (exactly one binary entry),
	- replace-write preflight checks,
	- up-to-date short-circuit when `current == latest`.
- Windows replacement path hardened to handle rename/lock behavior with actionable retry messaging.

### Security
- Self-upgrade now fails hard on checksum mismatch, missing checksum asset, invalid archive structure, or unsupported platform.

### Notes
- Upgrade scope is intentionally narrow: it replaces only the `rig` binary and does not mutate `rig.toml`, `rig.lock`, PATH, aliases, or project configuration.

## v0.3

### Added
- `rig dev` (alias: `rid`): tool-backed dev loop driven by `[tasks.dev]` and `rig.lock`.
- Finalized alias model via invocation-name dispatch (`rig`/`rir`/`ric`/`rid`/`ris`) and informational `rig alias`.
- `docs/INSTALLATION.md` and installer script intended for `curl -fsSL https://rig.sh/install | sh`.
- Optional `description` field for tasks (metadata-only), shown by `rig run --list`.

### Installer
- On macOS/Linux, `install.sh` creates optional alias symlinks (`rir`/`ric`/`rid`/`ris`) pointing to the installed `rig` binary.

### Changed
- Enforced single-binary distribution: only `cmd/rig` exists; aliases are invocation-name only.
- Task configuration is strict for determinism: tasks are either strings or tables with `{command, description, env, cwd, depends_on}`.
- `[tasks.dev]` additionally supports `watch` for `rig dev`.

### Notes
- `rig start` remains stubbed for future releases.

## v0.2

### Added
- Strict lock-backed execution for `rig run` and `rig check`.
- `rig check` CI-oriented validation for lock presence and tooling parity.
- `rig status` project health reporting around lock and tool parity.
- Offline tooling mode for sync/check workflows (`--offline`), using module-download suppression when requested.

### Changed
- `rig.lock` became the authoritative source of resolved tool state.
- `.rig/manifest.lock` kept as a deterministic hash cache for fast drift/parity checks.
- Tool sync/check workflows were tightened around deterministic lock semantics.

## v0.1

### Added
- Interactive `rig init` scaffolding.
- Core task runner (`rig run`) for string and table task forms.
- Build profile orchestration (`rig build`) over `go build`.
- Reproducible tooling workflow foundation (`rig tools sync`, lockfile generation).
