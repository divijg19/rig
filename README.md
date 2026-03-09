# rig

**The all-in-one modern toolchain, task runner, and developer environment for Go.**

> **`rig` = Cargo’s clarity and reliability + Bun’s DX + uv's hygiene + Go’s simplicity and no-nonsense ideology**

[![build status](https://img.shields.io/github/actions/workflow/status/divijg19/rig/build.yml?branch=main)](https://github.com/divijg19/rig/actions)
[![latest release](https://img.shields.io/github/v/release/divijg19/rig)](https://github.com/divijg19/rig/releases)
[![license](https://img.shields.io/github/license/divijg19/rig)](./LICENSE)

`rig` is an opinionated project orchestrator: it helps you define tasks, pin dev tools, and compose build profiles via a single `rig.toml`. It complements the Go toolchain; it does not replace `go build`, `go test`, or `go mod`.

## Why Rig?

- *One manifest:* `rig.toml` is the source of truth for project tasks, tools, and build intent.
- *Project-local tooling:* installs pinned tools into `.rig/bin` to avoid global conflicts.
- *Reproducible tooling:* `rig sync` writes `rig.lock` (deterministic, schema=0) and uses it to install and verify tools.
- *Fast parity checks:* also writes `.rig/manifest.lock` (a hash cache) so commands can quickly detect drift.
- *Ergonomic DX:* dry-runs, task listing, JSON output (where supported), and sensible `init` templates.

## Core Features:

- **⚡ Virtual Runtime (rig dev):** Native hot-reloading, environment variable injection, and instant feedback loops.
- **🎯 Process Multiplexing:** Concurrently run your Backend (Go), Web (Templ/Tailwind), and Mobile (Flutter) in one terminal window.
- **🔒 Hermetic Tooling:** rig manages non-Go tools too; they are version-locked in rig.lock and sandboxed per project. It downloads and version-locks tailwindcss, templ, and sqlc inside the project. No global version conflicts.
- **📦 Cargo-like Management:** A single rig.toml acts as the source of truth for tasks (scripts), tools, and build profiles.
- **🌉 Automated Pipelines:** Define "glue" tasks. rig watches files and triggers sqlc, swag, or codegen tools before your build runs.
- **🚀 Production Supervisor (rig start):** In production, rig acts as PID 1; a lightweight process manager for your binaries that handles graceful shutdowns, signal trapping, log formatting and secrets for your binary.

---

## Install

**Current installer**

```bash
curl -fsSL https://raw.githubusercontent.com/divijg19/rig/main/install.sh | sh
```

For complete installation options (`go install`, alias symlinks, Windows notes), see `docs/INSTALLATION.md`.

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

## Command Reference

| Command | Description |
| :--- | :--- |
| **`rig init`** | Generate a `rig.toml` (interactive or flags). |
| **`rig run <task>`** | Run a task from `[tasks]`. (alias entrypoint: `rir`) |
| **`rig dev`** | Run the watcher-backed dev loop (alias entrypoint: `rid`). |
| **`rig status`** | Show current state (read-only). |
| **`rig build`** | Compose and run `go build` using optional profiles. |
| **`rig tools`** | Manage tools declared in `[tools]` (sync/check/outdated). |
| **`rig ls`** | List locked tools in deterministic order. Shortcut for `rig tools ls` (alias entrypoint: `ril`). |
| **`rig path <name>`** | Print absolute path in `.rig/bin` with lock+checksum validation. Shortcut for `rig tools path <name>` (alias entrypoint: `rip`). |
| **`rig why <name>`** | Show requested/resolved/sha/path provenance for a tool. Shortcut for `rig tools why <name>` (alias entrypoint: `riw`). |
| **`rig doctor`** | Verify local environment and toolchain sanity. |
| **`rig doctor [name]`** | Diagnose tool presence, executable bit, and checksum parity. |
| **`rig sync`** | Shortcut for `rig tools sync`. |
| **`rig check`** | Verify `rig.lock` and `.rig/bin` tool parity (alias entrypoint: `ric`). |
| **`rig outdated`** | Shortcut for `rig tools outdated`. |
| **`rig x`** | Run a tool ephemerally. (alias entrypoint: `rix`) |
| **`rig upgrade`** | Upgrade the `rig` binary with asset+SHA verification and safe replacement semantics. |
| **`rig setup`** | Convenience installer for `[tools]` (similar to sync). Shortcut for `rig tools setup`. |

---

## Self-upgrade behavior

- Scope: `rig upgrade` only replaces the current `rig` executable. It does not modify `rig.toml`, `rig.lock`, PATH, aliases, or project config.
- Version gate: if latest release tag exactly matches current version, it prints up-to-date and does not replace the binary.
- Integrity: downloads both release asset and `.sha256`, validates filename and SHA256 before extraction.
- Archive contract: requires exactly one binary entry (`rig` on Unix, `rig.exe` on Windows).
- Replacement: uses temp-file replacement; on Windows, if the running binary is locked, it returns an actionable message to close running `rig` processes and retry.

---

## Documentation

- [docs/INSTALLATION.md](docs/INSTALLATION.md): installation methods, aliases, and platform notes.
- [SECURITY.md](SECURITY.md): vulnerability reporting and security policy.
- [docs/CLI.md](docs/CLI.md): CLI commands, flags, and workflows.
- [docs/CONFIGURATION.md](docs/CONFIGURATION.md): `rig.toml` schema and behavior.
- [docs/CHEATSHEET.md](docs/CHEATSHEET.md): quick reference.
- [docs/PRODUCTION.md](docs/PRODUCTION.md): production-oriented guidance and operational notes.
- [docs/GOLDEN_STACK.md](docs/GOLDEN_STACK.md): reference stack and example task layout.
- [docs/PHILOSOPHY.md](docs/PHILOSOPHY.md): project principles and design direction.
- [docs/ROADMAP.md](docs/ROADMAP.md): planned features and release direction.
- [examples/README.md](examples/README.md): copy-pasteable manifests.

#### Bonus End-goals
- Zig: Introduce `zig cc` as a linker or plausible build tool for all cgo and c, c++ code managed through `rig`.
- Glyph: Integrate `glyph *` commands directly into rig.
---

Made with ❤️ for the Go community, and dedicated to Tarushi, this project's origin.
