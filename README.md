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

**Official installer (recommended)**
```bash
curl -fsSL https://rig.sh/install | sh
```

On macOS/Linux, the installer installs `rig` and also creates these optional symlink entrypoints next to it:

- `rir` → `rig run`
- `ric` → `rig check`
- `rid` → `rig dev`
- `ris` → `rig start`

On Windows, invoke `rig run`, `rig check`, `rig dev`, etc directly until a PowerShell installer exists.

**Go install (single binary only)**
```bash
# Install the single main binary
go install github.com/divijg19/rig/cmd/rig@v0.3
```
Ensure `$GOPATH/bin` is in your system's `PATH`.

`go install` does not create aliases. If you want `rir`/`ric`/`rid`/`ris`, create symlinks manually:
```bash
ln -sf "$(command -v rig)" "$HOME/.local/bin/rir"
ln -sf "$(command -v rig)" "$HOME/.local/bin/ric"
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

## Upgrade notes (v0.2 → v0.3)

- `rig dev` is now a first-class command (alias: `rid`). Configure it via `[tasks.dev]` with `command` + `watch`.
- Alias model is finalized: one binary with invocation-name dispatch. See `rig alias`.

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
| **`rig sync`** | Shortcut for `rig tools sync`. |
| **`rig check`** | Verify `rig.lock` and `.rig/bin` tool parity (alias: `ric`). |
| **`rig outdated`** | Shortcut for `rig tools outdated`. |
| **`rig x`** | Run a tool ephemerally. |
| **`rig doctor`** | Verify local environment and toolchain sanity. |
| **`rig setup`** | Convenience installer for `[tools]` (similar to sync). |

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
