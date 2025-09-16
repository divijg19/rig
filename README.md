# rig

Project manager and task runner for Go — orchestrates the official toolchain with a single declarative manifest.

**rig = cargo’s clarity + bun’s DX + Go’s no-nonsense tooling**

[![build status](https://img.shields.io/github/actions/workflow/status/divijg19/rig/build.yml?branch=main)](https://github.com/divijg19/rig/actions)
[![latest release](https://img.shields.io/github/v/release/divijg19/rig)](https://github.com/divijg19/rig/releases)
[![license](https://img.shields.io/github/license/divijg19/rig)](./LICENSE)

`rig` is an opinionated, all-in-one project manager for Go. It enhances the native toolchain with a single, declarative manifest (`rig.toml`), solving common pain points like script cross-compatibility, reproducible tooling, and task discovery — **without replacing `go build`, `go test`, or `go mod`.**

## The Philosophy

`rig` is built on a few core beliefs:

*   **Embrace the Go Toolchain:** `rig` is a thin orchestration layer, not a replacement. It calls the `go` commands you already know and love, adding structure and reproducibility on top.
*   **Declarative is Better:** Your project's entire workflow—its tools, scripts, and build configurations—should live in a single, human-readable manifest. No more deciphering cryptic `Makefile`s or shell scripts.
*   **Reproducibility is Not Optional:** A project that works "on my machine" is a broken project. `rig` guarantees that every developer and CI machine runs the exact same setup by locking Go and tool versions.
*   **Excellent DX is a Feature:** A tool should be fast, intuitive, and a joy to use. `rig` provides clear output, helpful error messages, and powerful shortcuts to make your development cycle smoother.

> **Why not just use `make`?**
> While powerful, `make` is not cross-platform (a major friction point on Windows), has an arcane syntax, and lacks native awareness of the Go ecosystem. `rig` provides a structured, TOML-based solution that is OS-agnostic and designed specifically for the needs of a Go developer.

## Key Features

*   **Declarative Project Manifest:** A single `rig.toml` defines project metadata, tasks, development tools, and build profiles. It's the single source of truth for your project's workflow.
*   **Effortless Project Initialization:** `rig init` interactively scaffolds your project, detecting the module name, author from git, and the Go version you're using. Use `--yes` for non-interactive CI environments.
*   **Structured, Cross-Platform Tasks:** Define tasks as simple command strings or as structured objects with descriptions, environment variables, and sequential dependencies.
*   **Reproducible Tool Management:** Pin exact versions of your dev tools (linters, mock generators, etc.). `rig tools sync` installs them into a project-local `.rig/bin` directory, guaranteeing consistency for the whole team.
*   **Environment-Aware Build Profiles:** Define different build flags for different environments (e.g., `rig build --profile release` to create a stripped, optimized binary).
*   **Ephemeral Runner (`rig x`):** Run any tool from the Go ecosystem on-the-fly without permanently installing it, similar to `npx` or `bunx`.
*   **Monorepo Support:** Use `include` directives in your `rig.toml` to split your configuration into smaller, manageable files.

## Install

Requires a recent stable Go toolchain.

```bash
go install github.com/divijg19/rig@latest
```
*On Windows, the binary is placed under `%GOPATH%\bin`. On Linux/macOS, it's `$GOPATH/bin`. Ensure this directory is in your system's `PATH`.*

## Quick Start

1.  **Initialize your project:** `rig` will create a `rig.toml` file with smart defaults.
    ```bash
    cd my-go-project
    rig init

    # Interactive prompts will guide you:
    # ? project name: (my-go-project)
    # ? version: (0.1.0)
    # ? author: (Your Name <your@email.com>)
    # ? Go version detected: 1.22.1 (add to tools?) [Y/n]
    ```

2.  **Define your tasks and tools** in the newly created `rig.toml`:
    ```toml
    [project]
    name = "my-go-project"
    version = "0.1.0"
    authors = ["Your Name <your@email.com>"]

    [tools]
    go = "1.22.1"
    golangci-lint = "1.59.1"

    [tasks]
    test = "go test -v -race ./..."
    lint = "golangci-lint run"
    ```

3.  **Sync your tools:** This installs the pinned tools into `.rig/bin` and creates a `.rig/manifest.lock` file. **Commit both `rig.toml` and the lock file to version control.**
    ```bash
    rig sync        # Shortcut for `rig tools sync`

    # 🔧 Syncing tools from rig.toml...
    # ✅ go 1.22.1 installed
    # ✅ golangci-lint v1.59.1 installed
    # 🔒 Tools synced and locked in .rig/manifest.lock
    ```

4.  **Run your tasks:**
    ```bash
    # List all available tasks with their descriptions
    rig ls          # Shortcut for `rig run --list`

    # Run your defined tasks
    rig run lint
    rig run test
    ```

## The `rig.toml` Manifest Explained

The manifest is the heart of `rig`. It's designed to be comprehensive yet easy to read.

```toml
# rig.toml

# (Required) Basic information about your project.
[project]
name = "my-api"
version = "1.0.0"
authors = ["Your Name <your@email.com>"]
license = "MIT"

# (Optional) Merge other .toml files for better organization, especially in monorepos.
# Paths are relative to the root rig.toml.
# include = [".rig/tools.toml", ".rig/tasks.toml"]

# (Optional) Pin exact versions of Go and dev tools for reproducible builds.
# Tools are installed into a project-local .rig/bin directory.
[tools]
go = "1.22.1"                      # Pin the Go toolchain itself.
golangci-lint = "1.59.1"           # Use short names for common tools.
mockery = "v2.43.2"
gofumpt = "latest"                 # "latest" is also a valid version specifier.
"mvdan.cc/sh/v3/cmd/shfmt" = "v3.8.0" # Use full module paths for any other tool.

# (Required) Define your project's scripts and commands.
[tasks]
# Simple tasks are just command strings.
test = "go test -v -race ./..."
generate = "go generate ./..."

# Structured tasks offer more control for complex scenarios.
[tasks.dev]
command = "air" # Assumes 'air' is defined in [tools]
description = "Run the dev server with live reload and debug mode"
env = { GIN_MODE = "debug", PORT = "8080" }
# depends_on runs tasks sequentially before this one. If any dependency fails, this task will not run.
depends_on = ["generate"]

[tasks.ci-check]
description = "Run all verification steps for CI"
depends_on = ["test", "lint"]

# (Optional) Define build profiles for different environments.
# These profiles are activated with `rig build --profile <name>`.
[profile.release]
description = "Optimized build for production."
# Flags are passed directly to the `go build` command.
flags = ['-ldflags="-s -w"', '-trimpath']

[profile.instrumented]
description = "Build with race detector and other debug flags."
flags = ['-race', '-tags="integ_tests"']
```

## Core Concepts

### 1. Task Running (`rig run`)

Tasks are the scripts that automate your workflow, from testing and linting to deployment.

```bash
# Run a simple task
rig run lint

# Run a task with dependencies. `generate` will run before `dev`.
rig run dev

# Preview the commands rig would execute without running them
rig run dev --dry-run

# Pass extra arguments to the underlying command after a `--`
rig run test -- -short -count=1
```

### 2. Building with Profiles (`rig build`)

`rig build` is an intelligent wrapper around `go build` that uses your defined profiles.

```bash
# Standard development build (no profile)
rig build ./cmd/server

# Build using the 'release' profile from rig.toml
rig build --profile release ./cmd/server
# This would execute: go build -ldflags="-s -w" -trimpath ./cmd/server

# Build with another profile
rig build --profile instrumented ./cmd/server
# This would execute: go build -race -tags="integ_tests" ./cmd/server
```

### 3. Reproducible Tooling (`rig tools`)

`rig` ensures your entire team uses the exact same version-locked tools. The workflow is explicit, reproducible, and sandboxed to your project.

*   **How it Works:** When you run `rig tools sync`, `rig` downloads the specified tools, installs them to `.rig/bin`, and records the exact resolved versions in `.rig/manifest.lock`. The `.rig/bin` directory is then automatically added to the `PATH` for any task you run, ensuring your scripts always execute the correct, project-local tool.

*   **The Workflow:**
    1.  **Define:** Add tools to the `[tools]` section of `rig.toml`.
    2.  **Sync:** Run `rig sync`.
    3.  **Commit:** Commit `rig.toml` and `.rig/manifest.lock`.
    4.  **Verify in CI:** Run `rig check` in your CI pipeline to guarantee the environment is in sync. It will exit with a non-zero status if there's a mismatch.

### 4. Ephemeral Runner (`rig x`)

Run a tool from the Go ecosystem on-the-fly without permanently installing it. `rig` resolves, downloads, caches, and runs it in one step. This is perfect for one-off commands or trying out new tools.

```bash
# Run a specific version of a linter on your project
rig x golangci-lint@1.59.1 run

# Get help for a specific tool
rig x mockery@latest -- --help
```

### 5. Monorepo Workflows

For large projects, you can split your `rig.toml` into multiple files using the `include` directive. This keeps your configuration clean and organized.

**Example Directory Structure:**
```
.
├── .rig/
│   ├── tasks.toml
│   └── tools.toml
├── cmd/
│   └── server/
│       └── main.go
├── go.mod
└── rig.toml
```

**Root `rig.toml`:**
```toml
[project]
name = "my-monorepo"
version = "0.1.0"

# Include shared configurations
include = [".rig/tasks.toml", ".rig/tools.toml"]
```

## Command Reference

#### Shortcuts
-   `rig r` → `rig run`
-   `rig b` → `rig build`
-   `rig ls` → `rig run --list`
-   `rig sync` → `rig tools sync`
-   `rig check` → `rig tools sync --check`
-   `rig outdated` → `rig tools outdated`

#### Global Flags & UX Guarantees
-   `-C, --dir <path>`: Set the working directory for a command.
-   `-n, --dry-run`: Print the command(s) without executing.
-   `-E, --env KEY=VALUE`: Add or override environment variables for a task. Can be used multiple times.
-   `-j, --json`: Output structured JSON where supported (e.g., `rig ls -j`, `rig outdated -j`).
-   `--check`: Verify state without making changes. Exits non-zero on mismatch.
-   **Environment Variable Precedence:** Command-line `-E` flags > `[tasks.name.env]` in `rig.toml` > Shell environment.
-   **PATH Precedence:** The project-local `.rig/bin` is always prepended to the `PATH` for any executed task, ensuring your pinned tools are always used.

---

## Learn More

*   **[Philosophy](./PHILOSOPHY.md):** Understand the "why" behind `rig`.
*   **[Roadmap](./ROADMAP.md):** See where the project is headed.
*   **[Contributing](./CONTRIBUTING.md):** Learn how you can help build `rig`.

---

Made with ❤️ for the Go community, and dedicated to Tarushi, this project's origin.
