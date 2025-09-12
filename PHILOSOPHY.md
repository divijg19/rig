# The `rig` Philosophy

`rig` is an opinionated tool. It is built on a set of core beliefs about what makes a development experience productive and joyful. This document outlines those beliefs.

Our philosophy is captured in a single mantra:

> **rig = cargo’s clarity + bun’s DX + Go’s no-nonsense tooling**

This is not just a slogan; it is the design principle against which every feature and decision is measured.

### The Three Pillars

#### 1. Cargo’s Clarity: A Single Source of Truth

We believe that a project's configuration should be declarative, unambiguous, and centralized. The Go ecosystem is powerful but fragmented, often requiring a `go.mod` for dependencies, a `Makefile` for tasks, a `.golangci.yml` for linting, and various shell scripts for orchestration.

`rig` brings clarity by introducing a single `rig.toml` manifest. Inspired by Rust's `Cargo.toml`, this file is the single source of truth for project metadata, tasks, and toolchain definitions. When you want to know how a project works, you look in one place.

#### 2. Bun’s DX: Development Should Be Fast and Fun

We believe that developer tools should be fast, intuitive, and a joy to use. A great Developer Experience (DX) is not a luxury; it is a core feature that enables higher productivity and better software.

`rig` is obsessed with DX. This means:
*   **Speed:** Commands should be instant. File watchers should react immediately.
*   **Polish:** Interfaces should be clean and informative. Our test runner doesn't just run tests; it presents the results in a beautiful, actionable way.
*   **Thoughtfulness:** Features like the native file watcher and scaffolding engine are designed to anticipate developer needs and eliminate common frustrations.

#### 3. Go’s No-Nonsense Tooling: Orchestrate, Don't Replace

We believe in the power and simplicity of the official Go toolchain. The Go compiler is fast, `go mod` is effective, and `go test` is robust. `rig` is not here to replace them.

`rig` is an orchestration layer. It enhances the existing tools by filling in the gaps.
*   It uses `go install` to manage your toolchain.
*   It calls `go test` to run your tests.
*   It invokes `go build` to compile your code.

`rig` respects the Go ecosystem. It is a no-nonsense tool designed to manage the "meta-work" so you can focus on writing Go.

### What `rig` is Not

*   **It is not a package manager.** `go mod` is the package manager.
*   **It is not a compiler or bundler.** The `go` command handles that.
*   **It is not a framework.** It is a tool for managing your project, regardless of the libraries or frameworks you use.
