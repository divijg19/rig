# rig

**The all-in-one modern toolchain, task runner, and developer environment for Go.**

> **`rig` = Cargo’s clarity and reliability + Bun’s DX + uv's hygiene + Go’s simplicity and no-nonsense ideology**

[![build status](https://img.shields.io/github/actions/workflow/status/divijg19/rig/build.yml?branch=main)](https://github.com/divijg19/rig/actions)
[![latest release](https://img.shields.io/github/v/release/divijg19/rig)](https://github.com/divijg19/rig/releases)
[![license](https://img.shields.io/github/license/divijg19/rig)](./LICENSE)

`rig` is an opinionated project orchestrator: it helps you define tasks, pin dev tools, and compose build profiles via a single `rig.toml`. It complements the Go toolchain; it does not replace `go build`, `go test`, or `go mod`.

## Why Rig?

- **One manifest:** `rig.toml` is the source of truth for project tasks, tools, and build intent.
- **Project-local tooling:** installs pinned tools into `.rig/bin` to avoid global conflicts.
- **Reproducible tooling:** `rig sync` writes `rig.lock` (deterministic, schema=0) and uses it to install and verify tools.
- **Fast parity checks:** also writes `.rig/manifest.lock` (a hash cache) so commands can quickly detect drift.
- **Ergonomic DX:** dry-runs, task listing, JSON output (where supported), and sensible `init` templates.

---

## Install

**Current installer**

```bash
curl -fsSL https://raw.githubusercontent.com/divijg19/rig/main/install.sh | sh
```

**Eventual official installer**
```bash
curl -fsSL https://rig.sh/install | sh
```

On macOS/Linux, the installer installs `rig` and also creates these optional symlink entrypoints next to it:

- `rir` → `rig run`
- `ric` → `rig check`
- `rid` → `rig dev`
- `ris` → `rig start`

Additional entrypoints are supported by invocation name and can be created manually:

- `ril` → `rig tools ls`
- `rip` → `rig tools path`
- `riw` → `rig tools why`

On Windows, invoke `rig run`, `rig check`, `rig dev`, etc directly until a PowerShell installer exists.

**Go install (single binary only)**
```bash
# Install the single main binary
go install github.com/divijg19/rig/cmd/rig@latest
```
Ensure `$GOPATH/bin` is in your system's `PATH`.

`go install` does not create aliases. If you want aliases, create symlinks manually:
```bash
ln -sf "$(command -v rig)" "$HOME/.local/bin/rir"
ln -sf "$(command -v rig)" "$HOME/.local/bin/ric"
ln -sf "$(command -v rig)" "$HOME/.local/bin/ril"
ln -sf "$(command -v rig)" "$HOME/.local/bin/rip"
ln -sf "$(command -v rig)" "$HOME/.local/bin/riw"
ln -sf "$(command -v rig)" "$HOME/.local/bin/rid"
ln -sf "$(command -v rig)" "$HOME/.local/bin/ris"
```

---

## Quick Start

```bash
cd my-go-project

# scaffold a rig.toml
rig init

# install tools declared in [tools] (writes rig.lock + .rig/manifest.lock)
rig sync

# start the dev loop (requires rig.lock)
rig dev

# discover tasks
rig run --list

# run a task
rig run test
```

Example tooling pins:

```toml
[tools]
golangci-lint = "1.62.0"
github.com/vektra/mockery/v2 = "v2.46.0"
```

---

## Quick upgrade notes 
### (v0.2) 

- `rig run` (alias: rir) now requires `rig.lock` for deterministic tool validation and execution. List tasks set in rig.toml quickly with `rig run --list` or `rir --list` if you use aliases. Run `rig sync` to generate the lock file and install tools.
- `rig check` (alias: `ric`) verifies `rig.lock` presence and tool parity. Use it in CI to ensure a clean state before running tasks.
- `rig status` checks for `rig.lock` and `.rig/bin` parity and reports the current state of the project environment.
- `rig sync` now writes `rig.lock` (schema=0) and `.rig/manifest.lock` (a hash cache) for deterministic tool resolution and fast parity checks.

### (v0.3)

- `rig dev` is now a first-class command (alias: `rid`). Configure it via `[tasks.dev]` with `command` + `watch`.
- Alias model is finalized: one binary with invocation-name dispatch. See `rig alias`.

### (v0.4)

- Tool observability commands are available: `rig tools ls`, `rig tools path <name>`, `rig tools why <name>`, `rig tools doctor [name]`.
- New invocation-name aliases are supported: `ril`, `rip`, and `riw`.
- `rig doctor` now accepts an optional tool name and dispatches to tool diagnostics.
- `rig upgrade` performs a checksum-verified binary self-update with an up-to-date short-circuit.

---

## Command Reference

| Command | Description |
| :--- | :--- |
| **`rig init`** | Generate a `rig.toml` (interactive or flags). |
| **`rig run <task>`** | Run a task from `[tasks]`. |
| **`rig dev`** | Run the watcher-backed dev loop (alias: `rid`). |
| **`rig status`** | Show current state (read-only). |
| **`rig build`** | Compose and run `go build` using optional profiles. |
| **`rig tools`** | Manage tools declared in `[tools]` (sync/check/outdated). |
| **`rig tools ls`** | List locked tools in deterministic order (alias entrypoint: `ril`). |
| **`rig tools path <name>`** | Print absolute path in `.rig/bin` with lock+checksum validation (alias entrypoint: `rip`). |
| **`rig tools why <name>`** | Show requested/resolved/sha/path provenance for a tool (alias entrypoint: `riw`). |
| **`rig tools doctor [name]`** | Diagnose tool presence, executable bit, and checksum parity. |
| **`rig sync`** | Shortcut for `rig tools sync`. |
| **`rig check`** | Verify `rig.lock` and `.rig/bin` tool parity (alias: `ric`). |
| **`rig outdated`** | Shortcut for `rig tools outdated`. |
| **`rig x`** | Run a tool ephemerally. |
| **`rig doctor`** | Verify local environment and toolchain sanity. |
| **`rig upgrade`** | Upgrade the `rig` binary with asset+SHA verification and safe replacement semantics. |
| **`rig setup`** | Convenience installer for `[tools]` (similar to sync). |

---

## Self-upgrade behavior

- Scope: `rig upgrade` only replaces the current `rig` executable. It does not modify `rig.toml`, `rig.lock`, PATH, aliases, or project config.
- Version gate: if latest release tag exactly matches current version, it prints up-to-date and does not replace the binary.
- Integrity: downloads both release asset and `.sha256`, validates filename and SHA256 before extraction.
- Archive contract: requires exactly one binary entry (`rig` on Unix, `rig.exe` on Windows).
- Replacement: uses temp-file replacement; on Windows, if the running binary is locked, it returns an actionable message to close running `rig` processes and retry.

---

## Documentation

- [docs/CONFIGURATION.md](docs/CONFIGURATION.md): `rig.toml` schema and behavior.
- [docs/CLI.md](docs/CLI.md): CLI commands, flags, and workflows.
- [docs/CHEATSHEET.md](docs/CHEATSHEET.md): quick reference.
- [examples/README.md](examples/README.md): copy-pasteable manifests.

#### Bonus End-goals
- Zig: Introduce `zig cc` as a linker or plausible build tool for all cgo and c, c++ code managed through `rig`.
- Glyph: Integrate `glyph *` commands directly into rig.
---

Made with ❤️ for the Go community, and dedicated to Tarushi, this project's origin.
