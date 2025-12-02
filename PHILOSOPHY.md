# The `rig` Philosophy

`rig` is an opinionated tool built on a set of core beliefs about what makes a development experience productive, reproducible, and joyful.

Our philosophy is captured in a single equation:

> **rig = Cargo’s clarity + Bun’s DX + uv's hygiene + Go’s simplicity and no-nonsense ideology**

This is not just a slogan; it is the design principle against which every feature is measured. We are building the "missing piece" of the Go ecosystem—an all-in-one developer environment that feels like a modern runtime, without sacrificing the performance of compiled Go.

### The Three Pillars

#### 1. Cargo’s Clarity & uv's Hygiene: Determinism is King
We believe that a project that works "on my machine" but fails in CI is a broken project. The current Go ecosystem often relies on global tool installations (`go install ...`) and ad-hoc Makefiles, leading to version drift and "works for me" syndrome.

`rig` solves this with **Strict Determinism**:
*   **Single Source of Truth:** Your `rig.toml` defines everything—metadata, scripts, and toolchain versions.
*   **Hermetic Tooling:** `rig` never touches your global `$PATH`. Tools (like linters and generators) are sandboxed per-project and version-locked in `rig.lock`.
*   **Structure:** Like Cargo, `rig` provides opinionated scaffolding (`rig new`) to standardize project layouts, ending the debate on where folders should go.

#### 2. Bun’s DX: The "Virtual Runtime" Experience
We believe developers shouldn't have to stitch together five different tools just to write code. While Go is a compiled language, the *development loop* should feel as instant and fluid as a scripting language.
We believe that developer tools should be fast, intuitive, and a joy to use. A great Developer Experience (DX) is not a luxury; it is a core feature that enables higher productivity and better software.

`rig` is obsessed with DX. This means:
*   **Speed:** Commands should be instant. File watchers should react immediately.
*   **Polish:** Interfaces should be clean and informative. Our test runner doesn't just run tests; it presents the results in a beautiful, actionable way.
*   **Thoughtfulness:** Features like the native file watcher and scaffolding engine are designed to anticipate developer needs and eliminate common frustrations.

`rig` provides an **All-in-One Environment**:
*   **Batteries Included:** You shouldn't need a separate binary for hot-reloading, another for testing, and another for env vars. `rig dev` handles watching, rebuilding, and restarting natively.
*   **Speed as a Feature:** `rig` is written in Go. It starts instantly, caches aggressively (inspired by `uv`), and respects your time.
*   **Unified Interface:** Whether you are testing (`rig test`), building (`rig build`), or running a one-off tool (`rig x`), the interface is consistent, pretty, and human-readable.

#### 3. Go’s Ideology: An Exoskeleton, Not a Cage
We believe in the power of the official Go toolchain. `rig` is not here to replace `go build`, `go mod`, or `go test`. It is here to orchestrate them.
`rig` respects the Go ecosystem. It is a no-nonsense tool designed to manage the "meta-work" so you can focus on writing Go.

We believe in the power and simplicity of the official Go toolchain. The Go compiler is fast, `go mod` is effective, and `go test` is robust. `rig` is not here to replace them.

`rig` is an orchestration layer. It enhances the existing tools by filling in the gaps.
*   It uses `go install` to manage your toolchain.
*   It calls `go test` to run your tests.
*   It invokes `go build` to compile your code.

`rig` operates with **Zero Lock-In**:
*   **Transparent Wrapper:** Under the hood, `rig` constructs standard `go` commands. It doesn't use a custom compiler or a proprietary build process.
*   **Ejectable:** You can delete `rig` at any time and your project is still valid Go code.
*   **Process Supervisor:** In production, `rig start` acts as a lightweight process manager (handling signals, logs, and secrets), but the binary it runs is standard Go.

### What `rig` is Not

*   **It is not a Language Runtime.** It does not interpret Go code like Node.js interprets JavaScript. It compiles Go using the standard toolchain, but manages the lifecycle so fluidly it *feels* like a runtime.
*   **It is not a Package Manager.** `go mod` is the package manager. `rig` complements it by managing *tool* dependencies and project workflows.
*   **It is not a Framework.** It doesn't care if you use Gin, Echo, or the stdlib. It manages the *environment* around your code, not the code itself.
