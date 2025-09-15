## Tools

You can pin tools in `[tools]` in `rig.toml` using either full Go module paths or supported short names:

| Short Name  | Module Path                                       |
|-------------|---------------------------------------------------|
| golangci-lint | github.com/golangci/golangci-lint/cmd/golangci-lint |
| mockery     | github.com/vektra/mockery/v2                      |
| staticcheck | honnef.co/go/tools/cmd/staticcheck                |
| revive      | github.com/mgechev/revive                         |
| air         | github.com/cosmtrek/air                           |
| reflex      | github.com/cespare/reflex                         |
| dlv         | github.com/go-delve/delve/cmd/dlv                 |
| gotestsum   | gotest.tools/gotestsum                            |
| gci         | github.com/daixiang0/gci                          |
| gofumpt     | mvdan.cc/gofumpt                                  |

These short-name mappings are centrally maintained in `internal/rig/tooling.go`. If you’d like another common tool added, open an issue or PR to extend `ToolShortNameMap`.

You can also install tools from a file (like pip requirements.txt):

```bash
rig tools sync tools.txt
```
Where `tools.txt` contains lines like:
```
# name = version or module@version
golangci-lint = 1.62.0
mockery = v2.46.0
staticcheck = latest
github.com/vektra/mockery/v2@v2.46.0
```
Blank lines and # comments are ignored. If only a name is given, "latest" is used.

**Note**: The `rig setup` command is now legacy. Use `rig tools sync` for the modern, explicit workflow.
# rig

Project manager and task runner for Go — orchestrates the official toolchain with a single declarative manifest.

**rig = cargo’s clarity + bun’s DX + Go’s no-nonsense tooling**

[![build status](https://img.shields.io/github/actions/workflow/status/divijg19/rig/build.yml?branch=main)](https://github.com/divijg19/rig/actions)
[![latest release](https://img.shields.io/github/v/release/divijg19/rig)](https://github.com/divijg19/rig/releases)
[![license](https://img.shields.io/github/license/divijg19/rig)](./LICENSE)

`rig` is an opinionated, all-in-one project manager and task runner for Go. It enhances the native Go toolchain with a single, declarative manifest (`rig.toml`), solving common pain points like script cross-compatibility, reproducible tooling, and task discovery — without replacing `go build`, `go test`, or `go mod`.

### Key Features

• **Interactive Setup**: `rig init` with smart defaults
    - Detects project name from directory
    - Pulls author info from git config
    - Auto-detects Go version for tooling
    - `--yes` flag for non-interactive CI environments

• **Unified Manifest**: one `rig.toml` to rule them all
    - [project] metadata: name, version, authors, license
    - [tasks] supporting both simple strings and structured configurations
    - [tools] for version-locked dev tools with explicit sync
    - [profile.*] for build flags per environment (e.g., release)
    - include = ["..."] to split configs (e.g., monorepos via .rig/)

• **Structured Tasks**: beyond simple command strings
    ```toml
    [tasks]
    test = "go test ./..."  # Simple string
    
    [tasks.dev]            # Structured task
    command = "go run ./cmd/server"
    description = "Runs the development server"
    env = { GIN_MODE = "debug" }
    depends_on = ["generate-mocks"]
    ```

• **Explicit Tool Management**: no-magic approach
    - `rig tools sync` installs tools and creates manifest.lock
    - `rig tools sync --check` verifies state (now with `--json` summary output for CI)
    - `rig tools outdated` reports missing/mismatched (supports `--json`)
    - `rig run` performs lightning-fast lock verification
    - Project-local installs in `.rig/bin` with PATH precedence (deduplicated PATH)
    - CI and team members guaranteed same tool versions

• **Friendly DX**: emoji output, clear errors, and cross-platform behavior

### Install

- Requires: a recent stable Go toolchain

```bash
go install github.com/divijg19/rig@latest
```

On Windows, the binary is placed under `%GOPATH%\bin` (ensure it’s on your PATH).

### Quick Start

1.  **Initialize your project with smart defaults:**
    ```bash
    cd my-go-project
    rig init
    # Interactive prompts with smart defaults:
    # ? project name: (my-go-project)
    # ? version: (0.1.0) 
    # ? author: (Your Name <your@email.com>)
    # ? Go version detected: 1.21.5 (add to tools?) [Y/n]
    
    # For CI/automation:
    rig init --yes
    ```

2.  **Sync your tools once:**
    ```bash
    rig sync        # shortcut for `rig tools sync`
    # 🔧 Syncing tools from rig.toml
    # ✅ golangci-lint v1.62.0 installed
    # ✅ mockery v2.46.0 installed  
    # 🔒 Tools synced and locked in .rig/manifest.lock
    ```

3.  **Run tasks with enhanced features:**
    ```bash
    # list tasks with descriptions
    rig ls          # shortcut for `rig run --list`

    # run simple tasks
    rig run test

    # structured tasks with dependencies and env vars work automatically
    rig run dev

    # preview without executing
    rig run test --dry-run

    # ephemeral runs like bunx/uvx
    rig x golangci-lint@v1.62.0 run
    rig x mockery -- --help
    ```

### Advanced Task Configuration

You can define tasks as simple strings or structured configurations:

```toml
[tasks]
# Simple tasks (most common)
test = "go test -v -race ./..."
build = "go build ./..."

# Structured tasks for complex scenarios
[tasks.dev]
command = "go run ./cmd/server"
description = "Start development server with live reload"
env = { GIN_MODE = "debug", PORT = "8080" }
depends_on = ["generate-mocks", "build-assets"]

[tasks.deploy]
command = "docker build -t myapp . && docker push myapp:latest"
description = "Build and deploy to production"
env = { DOCKER_BUILDKIT = "1" }
depends_on = ["test", "lint"]
```

### Tool Management Workflow

```toml
# rig.toml
[tools]
go = "1.21.5"                    # Pin Go version
golangci-lint = "1.62.0"         # Use short names when available
github.com/vektra/mockery/v2 = "v2.46.0"  # Or full module paths
```

```bash
# One-time setup per project
rig sync

# Fast verification on every run (sub-millisecond)
rig run test
# ✅ Tools in sync
# 🚀 Running task "test"...

# Check status without changing state (human & JSON)
rig outdated
rig outdated --json | jq .

# Verify manifest lock without changing state
rig check
rig sync --check
# Machine-readable summary for CI
rig sync --check --json | jq .
```

### Learn More

*   **[Philosophy](./PHILOSOPHY.md):** Understand the "why" behind `rig`.
*   **[Roadmap](./ROADMAP.md):** See where the project is headed.
*   **[Contributing](./CONTRIBUTING.md):** Learn how you can help build `rig`.

---

Made with ❤️ for the Go community, and dedicated to Tarushi, this project's origin.

## Shared flags & UX

Keep flags minimal and consistent across commands:

- -C, --dir: set the working directory (run, build).
- -n, --dry-run: print the command(s) without executing (run, build).
- -E, --env KEY=VALUE: add/override environment variables for a run task; can be repeated. Merges with task env. `.rig/bin` is always prepended to PATH with deduplication.
- -l, --list: for `rig run`, lists tasks. Combine with `-j/--json` to get a machine-readable list including envKeys and shell when present.
- -j, --json: structured output where meaningful and deterministic:
    - `rig run --list --json` or `rig ls -j`
    - `rig outdated --json`
    - `rig sync --check --json` (only valid with `--check`)
- --check: verify state without making changes (e.g., `rig sync --check`). Exits non‑zero on mismatch; pair with `--json` in CI.

Other UX guarantees:

- PATH precedence: `.rig/bin` is prepended exactly once and deduplicated for clean environments.
- Deterministic outputs: names are sorted for stable CI logs and tests.
- Ephemeral runner: `rig x <tool[@version]> -- [args...]` resolves short names and pins versions when provided.

Shortcuts summary:

- `rig r` = `rig run`
- `rig b` = `rig build`
- `rig ls` = `rig run --list`
- `rig sync` = `rig tools sync`
- `rig check` = `rig tools check` = `rig tools sync --check`
- `rig outdated` = `rig tools outdated`
