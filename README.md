# rig

**The all-in-one modern toolchain, process and project management developer environment for Go.**

> **rig = Cargo’s clarity + Bun’s DX + uv's hygiene + Go’s simplicity and no-nonsense ideology**

[![build status](https://img.shields.io/github/actions/workflow/status/divijg19/rig/build.yml?branch=main)](https://github.com/divijg19/rig/actions)
[![latest release](https://img.shields.io/github/v/release/divijg19/rig)](https://github.com/divijg19/rig/releases)
[![license](https://img.shields.io/github/license/divijg19/rig)](./LICENSE)

`rig` is an opinionated project manager and process supervisor. It orchestrates the official Go toolchain with a single declarative manifest, bridging the gap between a build tool and a modern runtime experience.

It solves script cross-compatibility, tool versioning, and hot-reloading—**without replacing `go build`, `go test`, or `go mod`.**

## Why Rig?

*   **⚡ Virtual Runtime (`rig dev`):** Native hot-reloading, environment variable injection, and instant feedback loops.
*   **🔒 Strict Determinism:** Tools (linters, generators) are version-locked in `rig.lock` and sandboxed per project. No more global version conflicts.
*   **🚀 Production Supervisor (`rig start`):** A lightweight process manager for your binaries that handles graceful shutdowns, signal trapping, and secrets.
*   **📦 Cargo-like Management:** A single `rig.toml` acts as the source of truth for tasks, tools, and build profiles.

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

1.  **Initialize your project:**
    `rig` scans your project and creates a `rig.toml` with smart defaults.
    ```bash
    cd my-go-project
    rig init
    ```

2.  **Start the Dev Loop:**
    No need to configure `air` or write a Makefile. If `rig` detects a main file, it just works.
    ```bash
    rig dev
    
    # ⚡ Watching . for changes...
    # 🔨 Building... (12ms)
    # 🚀 Started (PID: 1234)
    ```

3.  **Sync your Tools:**
    If your team needs specific linters or generators, pin them in `rig.toml` and sync.
    ```bash
    rig sync
    
    # ✅ golangci-lint v1.59.1 installed (sandboxed)
    # 🔒 Versions locked in rig.lock
    ```

---

## Core Features

### 1. The Virtual Runtime (`rig dev`)
Rig makes Go feel like a scripting language during development. It watches your files, handles rebuilds incrementally, and manages the child process.

```toml
# rig.toml
[tasks.dev]
command = "go run main.go"
watch = ["."]                 # Native file watching
env = { PORT = "8080" }       # Injected env vars
ignore = ["tmp/", "vendor/"]
```

### 2. Production Supervisor (`rig start`)
In production, `rig` acts as the parent process (PID 1). It wraps your binary to provide modern observability and reliability features without changing your code.

```bash
# Runs the compiled binary with graceful shutdown handling and JSON log formatting
rig start --bin ./my-app
```

### 3. Hermetic Tooling
Stop asking your team to `go install` tools globally. Rig installs tools into a project-local `.rig/bin`, updates your `PATH` automatically during tasks, and locks versions in `rig.lock`.

```toml
[tools]
go = "1.23.0"
golangci-lint = "1.59.1"
mockery = "latest"
```

### 4. Ephemeral Runner (`rig x`)
Run a tool from the Go ecosystem on-the-fly without permanently installing it (inspired by `npx`/`bunx`).

```bash
rig x cobra-cli@latest init
```

---

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
| **`rig dev`** | Start the development server with file watching and live reload. |
| **`rig build`** | Build the project using defined profiles (e.g., `--profile release`). |
| **`rig test`** | Run tests (wraps `go test` with better output). |
| **`rig start`** | Run the binary in production mode (supervisor). |
| **`rig run <task>`** | Execute a script defined in `rig.toml`. |
| **`rig sync`** | Download pinned tools and generate `rig.lock`. |
| **`rig x <tool>`** | Download and execute a tool ephemerally (`rig x mockery`). |
| **`rig init`** | Scaffold a new `rig.toml` in the current directory. |

#### Global Flags
-   `-C, --dir <path>`: Set working directory.
-   `-E, --env KEY=VALUE`: Override environment variables.
-   `--json`: Output structured JSON (where supported).

---

## Learn More

*   **[Philosophy](./PHILOSOPHY.md):** Why we built `rig`, and why it's different from Makefiles.
*   **[Roadmap](./ROADMAP.md):** Upcoming features (Workspaces, Docker exports).
*   **[Contributing](./CONTRIBUTING.md):** Join the development.

---

Made with ❤️ for the Go community, and dedicated to Tarushi, this project's origin.
