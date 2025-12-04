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

### The File Structure Strategy

1.  **`README.md`**: The high-level overview, installation, and value prop.
2.  **`docs/CONFIGURATION.md`**: (New) The detailed `rig.toml` reference (moved from the original README).
3.  **`docs/CLI.md`**: (New) The full command reference and global flags (moved from the original README).
4.  **`docs/PRODUCTION.md`**: (New) Deep dive into `rig start` and the Supervisor/PID 1 features (Original feature preserved here).

---

# rig

**The full-stack orchestrator, modern toolchain, and process manager for Go.**

> **rig = Cargo’s clarity + Bun’s DX + uv's hygiene + Go’s reliability.**

[![build status](https://img.shields.io/github/actions/workflow/status/divijg19/rig/build.yml?branch=main)](https://github.com/divijg19/rig/actions)
[![latest release](https://img.shields.io/github/v/release/divijg19/rig)](https://github.com/divijg19/rig/releases)
[![license](https://img.shields.io/github/license/divijg19/rig)](./LICENSE)

`rig` is an opinionated meta-framework orchestrator. It replaces Makefiles, `air`, `npm run dev`, and shell scripts with a single, deterministic workflow.

It bridges the gap between a **Build Tool**, a **Process Manager**, and a **Developer Experience Platform**. Whether you are building a simple Go CLI or a complex **Go + Flutter + HTMX** stack, Rig manages the chaos.

## Why Rig?

*   **⚡ Process Multiplexing:** Run your Backend (Go), Web (Templ/Tailwind), and Mobile (Flutter) simultanously in one terminal window.
*   **🔒 Hermetic Tooling:** Rig manages non-Go tools too. It downloads and version-locks `tailwindcss`, `templ`, and `sqlc` inside the project. No global version conflicts. No `npm install`.
*   **🌉 Automated Pipelines:** Define "glue" tasks. Rig watches files and triggers `sqlc`, `swag`, or codegen tools before your build runs.
*   **🚀 Production Supervisor:** In production, `rig start` acts as PID 1. It handles graceful shutdowns, signal trapping, and log formatting for your binary.

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

## The "Bun-Like" Experience (Quick Start)

**1. Initialize the Project**
Rig scans your project and creates a `rig.toml` with smart defaults.
```bash
rig init
# Or scaffold a full stack: rig init --stack goth-flutter
```

**2. The One Command (`rig dev`)**
Stop opening 4 terminal tabs. Rig handles the entire dev loop.
```bash
rig dev
```
*   **What happens?** Rig verifies tool versions, runs generators (SQL/OpenAPI), starts the Go server (hot-reload), watches Tailwind CSS, and boots the Flutter emulator—all in one stream.

**3. Sync your Tools**
If your team needs specific linters or generators, pin them in `rig.toml` and sync.
```bash
rig sync
# ✅ golangci-lint v1.59.1 installed (sandboxed)
# 🔒 Versions locked in rig.lock
```

---

## Core Features

### 1. Multiplexing (The "Vite" Replacement)
Development often requires running multiple things at once. Rig manages them as a unified stream.

```toml
[tasks.dev]
description = "Start the Full Stack"
mode = "parallel" 

[tasks.dev.processes]
backend = { cmd = "go run main.go", watch = ["."], env = { PORT = "8080" } }
styles = { cmd = "tailwindcss -i input.css -o public/output.css --watch" }
mobile = { cmd = "flutter run", cwd = "./mobile", optional = true }
```

### 2. Hermetic Tooling (No `node_modules`)
Stop asking your team to `go install` tools globally. Rig installs tools into a project-local `.rig/bin` and updates your `PATH` automatically.

```toml
[tools]
go = "1.23.0"
templ = "v0.2.707"
# Rig downloads the standalone binary. No Node.js required.
tailwindcss = { version = "v3.4", url = "..." } 
```

### 3. Production Supervisor (`rig start`)
In production, Rig wraps your binary to provide modern observability and reliability features without changing your code.

```bash
# Runs with graceful shutdown handling and JSON log formatting
rig start --bin ./my-app
```

### 4. Ephemeral Runner (`rig x`)
Run a tool from the Go ecosystem on-the-fly without permanently installing it (inspired by `npx`/`bunx`).

```bash
rig x cobra-cli@latest init
```

---

## Documentation

*   **[Configuration Reference](./docs/CONFIGURATION.md):** Full documentation of `rig.toml` schema, workspaces, and profiles.
*   **[CLI Reference](./docs/CLI.md):** Detailed list of all commands and global flags.
*   **[Production Guide](./docs/PRODUCTION.md):** How to use Rig as a process supervisor in Docker/Kubernetes.
*   **[The Golden Stack](./docs/GOLDEN_STACK.md):** Guide to Go + Templ + Flutter development with Rig.

---

## Command Summary

| Command | Description |
| :--- | :--- |
| **`rig dev`** | Start the multiplexed development environment with hot-reload. |
| **`rig build`** | Build the project using defined pipelines and profiles. |
| **`rig start`** | Run the binary in production mode (Supervisor/PID 1). |
| **`rig sync`** | Download and lock tool versions in `rig.lock`. |
| **`rig x`** | Run a tool ephemerally (`rig x mockery`). |

---

Made with ❤️ for the Go community, and dedicated to Tarushi, this project's origin.
