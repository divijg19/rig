# rig

Project manager and task runner for Go — orchestrates the official toolchain with a single declarative manifest.

**rig = cargo’s clarity + bun’s DX + Go’s no-nonsense tooling**

[![build status](https://img.shields.io/github/actions/workflow/status/divijg19/rig/build.yml?branch=main)](https://github.com/divijg19/rig/actions)
[![latest release](https://img.shields.io/github/v/release/divijg19/rig)](https://github.com/divijg19/rig/releases)
[![license](https://img.shields.io/github/license/divijg19/rig)](./LICENSE)

`rig` is an opinionated, all-in-one project manager and task runner for Go. It enhances the native Go toolchain with a single, declarative manifest (`rig.toml`), solving common pain points like script cross-compatibility and task discovery — without replacing `go build`, `go test`, or `go mod`.

### Key Features

*   **Declarative Tasks:** Replace your `Makefile` with a clean, cross-platform `rig.toml` file.
*   **Reproducible Toolchains:** Version-lock your linters and code generators, ensuring your entire team and CI use the exact same tools.
*   **Enhanced DX:** A beautiful and fast test runner, built-in file watching, and powerful project scaffolding.
*   **Zero-Configuration Start:** Works out of the box with sensible defaults, but is fully configurable when you need it.

### Install

- Requires: a recent stable Go toolchain

```bash
go install github.com/divijg19/rig@latest
```

On Windows, the binary is placed under `%GOPATH%\bin` (ensure it’s on your PATH).

### Quick Start

1.  **Initialize your project:**
    ```bash
    cd my-go-project
    rig init
    ```
    This will create a `rig.toml` file.

2.  **Run tasks:**
    ```bash
    # list tasks from rig.toml
    rig run --list

    # run a specific task
    rig run test

    # preview without executing
    rig run test --dry-run

    # run from a different directory
    rig run build -C ./cmd/rig
    ```

### Learn More

*   **[Philosophy](./PHILOSOPHY.md):** Understand the "why" behind `rig`.
*   **[Roadmap](./ROADMAP.md):** See where the project is headed.
*   **[Contributing](./CONTRIBUTING.md):** Learn how you can help build `rig`.

---

Made with ❤️ for the Go community, and dedicated to Tarushi, this project's origin.
