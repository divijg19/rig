# Cassia

## Rig - all-in-one runtime, server and package manager for the Go lang &amp; community

[rig logo]

**rig = cargo’s clarity + bun’s DX + Go’s no-nonsense tooling**

[![build status](https://img.shields.io/github/actions/workflow/status/your-org/rig/build.yml?branch=main)](https://github.com/your-org/rig/actions)
[![latest release](https://img.shields.io/github/v/release/your-org/rig)](https://github.com/your-org/rig/releases)
[![license](https://img.shields.io/github/license/your-org/rig)](./LICENSE)

`rig` is an opinionated, all-in-one project management tool and task runner for Go. It enhances the natural Go toolchain with a single, declarative manifest, solving common pain points like toolchain management and script cross-compatibility without replacing the tools you already love.

### Key Features

*   **Declarative Tasks:** Replace your `Makefile` with a clean, cross-platform `rig.toml` file.
*   **Reproducible Toolchains:** Version-lock your linters and code generators, ensuring your entire team and CI use the exact same tools.
*   **Enhanced DX:** A beautiful and fast test runner, built-in file watching, and powerful project scaffolding.
*   **Zero-Configuration Start:** Works out of the box with sensible defaults, but is fully configurable when you need it.

### Quick Start

1.  **Install `rig`:**
    ```bash
    go install github.com/your-org/rig@latest
    ```

2.  **Initialize your project:**
    ```bash
    cd my-go-project
    rig init
    ```
    This will create a `rig.toml` file.

3.  **Run a task:**
    ```bash
    rig run test
    ```

### Learn More

*   **[Philosophy](./PHILOSOPHY.md):** Understand the "why" behind `rig`.
*   **[Roadmap](./ROADMAP.md):** See where the project is headed.
*   **[Contributing](./CONTRIBUTING.md):** Learn how you can help build `rig`.

---

Made with ❤️ for the Go community, and dedicated to Tarushi, this project's origin.
