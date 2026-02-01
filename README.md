# rig

**The all-in-one modern toolchain, full-stack orchestrator, process & project manager and developer environment for Go.**

> **`rig` = Cargo’s clarity and reliability + Bun’s DX + uv's hygiene + Go’s simplicity and no-nonsense ideology**

[![build status](https://img.shields.io/github/actions/workflow/status/divijg19/rig/build.yml?branch=main)](https://github.com/divijg19/rig/actions)
[![latest release](https://img.shields.io/github/v/release/divijg19/rig)](https://github.com/divijg19/rig/releases)
[![license](https://img.shields.io/github/license/divijg19/rig)](./LICENSE)

`rig` is an opinionated meta-framework orchestrator, project manager and process supervisor. It replaces Makefiles, `air`, `npm run dev`, and shell scripts with a single, deterministic workflow. 

It solves script cross-compatibility, tool versioning, and hot-reloading—**without replacing `go build`, `go test`, or `go mod`.** `rig` orchestrates the official Go toolchain with a single declarative manifest, bridging the gap between a **Build Tool**, a **Process Manager**, and a **Developer Experience Platform**. 

Whether you are building a simple Go CLI or a multi-faceted **Go + Flutter + HTMX** stack, `rig` manages the storm you bring to your workspace.

## Why Rig?

*   **⚡ Virtual Runtime (`rig dev`):** Native hot-reloading, environment variable injection, and instant feedback loops.
*   **🎯 Process Multiplexing:** Concurrently run your Backend (Go), Web (Templ/Tailwind), and Mobile (Flutter) in one terminal window.
*   **🔒 Hermetic Tooling:** `rig` manages non-Go tools too; they are version-locked in `rig.lock` and sandboxed per project. It downloads and version-locks `tailwindcss`, `templ`, and `sqlc` inside the project. No global version conflicts.
*   **📦 Cargo-like Management:** A single `rig.toml` acts as the source of truth for tasks (scripts), tools, and build profiles.
*   **🌉 Automated Pipelines:** Define "glue" tasks. `rig` watches files and triggers `sqlc`, `swag`, or codegen tools before your build runs.
*   **🚀 Production Supervisor (`rig start`):** In production, `rig` acts as PID 1; a lightweight process manager for your binaries that handles graceful shutdowns, signal trapping, log formatting and secrets for your binary.

---

## Install

**Via Shell (Recommended for CI/Mac/Linux)**
```bash
curl -fsSL https://rig.sh/install | sh
```

**Via Go Install**
```bash
go install github.com/divijg19/rig@latest
```
*Ensure `$GOPATH/bin` is in your system's `PATH`.*

---

## Quick Start

### 1. Initialize the project
> **`rig init`**

`rig` scans your project and creates a `rig.toml` with smart defaults.
```bash
cd my-go-project
rig init
# Or scaffold a full stack: rig init --stack goth-flutter
```

### 2. The Dev Loop
> **`rig dev`**

Stop opening 4 terminal tabs. No need to configure `air` or write a Makefile. If `rig` detects a main file, it just works.
```bash
rig dev

# ⚡ Watching . for changes...
# 🔨 Building... (12ms)
# 🚀 Started (PID: 1234)
```

*   **What happens?** `rig` verifies tool versions, runs generators (SQL/OpenAPI), starts the Go server (hot-reload), watches Tailwind CSS, and boots the Flutter emulator—all in one stream.

### 3. Sync Tools
> **`rig sync`**

If your team needs specific linters or generators, pin them in `rig.toml` and sync.
```bash
rig sync

# ✅ golangci-lint v1.59.1 installed (sandboxed)
# 🔒 Versions locked in rig.lock
```

---

## Core Features

### 1. ⚡ The Virtual Runtime (`rig dev`) & Multiplexing (The "Vite" Replacement)
Development often requires running multiple things at once. `rig` can multiplex multiple processes (like Tailwind or Flutter) alongside your Go server, managing them as a unified stream, acting as a "Vite" solution to make Go feel like a scripting language. It watches your files, handles rebuilds incrementally, and manages child processes.


```toml
[tasks.dev]
description = "Start the Full Stack"
mode = "parallel" 

[tasks.dev.processes]
backend = { cmd = "go run main.go", watch = ["."], env = { PORT = "8080" } }
styles = { cmd = "tailwindcss -i input.css -o public/output.css --watch" }
mobile = { cmd = "flutter run", cwd = "./mobile", optional = true }
```

### 2. 🔒 Hermetic Tooling (`rig.toml` & `rig.lock`, No `node_modules`)
Stop asking your team to `go install` tools globally. `rig` installs tools into a project-local `.rig/bin`, updates your `PATH` automatically during tasks, and locks versions in `rig.lock`.

```toml
[tools]
go = "1.23.0"
templ = "v0.2.707"
golangci-lint = "1.59.1"
# Rig downloads the standalone binary. No Node.js required.
tailwindcss = { version = "v3.4", url = "..." } 
```

### 3. 🚀 Production Supervisor (`rig start`)
In production, `rig` acts as the parent process (PID 1), wrapping your binary to provide modern observability and reliability features without changing your code.

```bash
# Runs with graceful shutdown handling and JSON log formatting
rig start --bin ./my-app
```

### 4. 🪄 Ephemeral Runner (`rig x`)
Run a tool from the Go ecosystem on-the-fly without permanently installing it (inspired by `npx`/`bunx`).

```bash
rig x cobra-cli@latest init
```

## The `rig.toml` Manifest

The manifest is the heart of `rig`.

```toml
[project]
name = "payment-service"
version = "0.1.0"

# Pin exact versions for reproducible builds
[tools]
go = "1.22.1"
golangci-lint = "1.59.1"

# Define reproducible scripts
[tasks]
test = "go test -v -race ./..."
lint = "golangci-lint run"

[tasks.dev]
description = "Start dev server with hot-reload"
command = "go run cmd/api/main.go"
watch = ["cmd/", "pkg/"]
env = { APP_ENV = "dev" }

# Build profiles for different environments
[profile.release]
flags = ['-ldflags="-s -w"', '-trimpath']
```

---

## Command Reference

| Command | Description |
| :--- | :--- |
| **`rig dev`** | Start the multiplexed development environment/server with file watching and live hot reload. |
| **`rig build`** | Build the project using defined pipelines and profiles (e.g., `--profile release`). |
| **`rig test`** | Run tests (wraps `go test` with better output). |
| **`rig start`** | Run the binary in production mode (Supervisor/PID 1). |
| **`rig run <task>`** | Execute a task (script) defined in `rig.toml`. |
| **`rig sync`** | Download and lock pinned tools in generated `rig.lock`. |
| **`rig x <tool>`** | Download and execute a tool ephemerally (`rig x mockery`). |
| **`rig init`** | Scaffold a new `rig.toml` in the current directory. |

#### Global Flags
-   `-C, --dir <path>`: Set working directory.
-   `-E, --env KEY=VALUE`: Override environment variables.
-   `--json`: Output structured JSON (where supported).

---

## Documentation

For advanced usage, please refer to the documentation folder:

*   **[Configuration Reference](./docs/CONFIGURATION.md):** Full documentation of the `rig.toml` schema, workspaces, and build profiles.
*   **[CLI Reference](./docs/CLI.md):** Detailed list of all commands (e.g., `rig build`, `rig test`) and global flags.
*   **[Production Guide](./docs/PRODUCTION.md):** How to use `rig` as a process supervisor in Docker/Kubernetes. Deep dive into `rig start`, PID 1 strategies, and Docker/Kubernetes integration.
*   **[The Golden Stack](./docs/GOLDEN_STACK.md):** Guide to Go + Templ + Flutter development with `rig`.

---


Made with ❤️ for the Go community, and dedicated to Tarushi, this project's origin.
