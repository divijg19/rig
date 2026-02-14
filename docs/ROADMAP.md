# Project `rig` Roadmap

This document outlines the planned development for `rig`, reflecting my ambition to deliver a best-in-class developer environment for Go. It is a living document, and priorities may shift based on community feedback.

`rig` is a deterministic, lock-backed toolchain orchestrator for Go, with:
> A single declarative manifest (`rig.toml`) that defines how a Go project builds, runs, and installs tools - reproducibly.

This roadmap reflects that philosophy. It prioritizes correctness, determinism, and clarity over surface features. The development is planned in phases, each with a clear theme and a goal to deliver significant, tangible value.

---

## ✅ Phase 1 — Declarative Core (v0.1)

**Theme:** Replace shell scripts with a strict manifest.

`rig` becomes a structured task runner.

Delivered:

* [x] `rig init` (interactive + non-interactive)
* [x] `[project]`, `[tasks]`, `[profile.*]`
* [x] `rig run <task>`
* [x] `rig build`
* [x] Strict task schema enforcement
* [x] Deterministic help + version output

**Result:**
A cross-platform alternative to Makefiles and ad-hoc scripts.

---

## ✅ Phase 2 — Reproducible Toolchain (v0.2–v0.3)

**Theme:** Lock-backed, hermetic tool execution.

`rig` becomes a reproducible workspace engine.

Delivered:

* [x] `[tools]` declaration
* [x] `rig sync`
* [x] `rig.lock`
* [x] Deterministic tool resolution
* [x] `.rig/bin` sandbox
* [x] `rig check`
* [x] `rig outdated`
* [x] Strict lock enforcement (run/dev require lock)
* [x] Stable `rig dev` (regex watcher freeze)
* [x] Unified help/version surface
* [x] Hardened installers (Unix + Windows)
* [x] Cross-platform release pipeline
* [x] Checksum-verified distribution

**Result:**
Team-safe, CI-safe deterministic environments.

---

## 🟢 Phase 3 — Tool Observability & Self-Management (v0.4)

**Theme:** Make the toolchain inspectable and self-healing.

`rig` becomes introspectable.

Delivered / In Progress:

* [x] `rig tools ls`
* [x] `rig tools path`
* [x] `rig tools why`
* [x] `rig tools doctor`
* [x] Root shortcuts for tools subcommands (`rig ls`, `rig path`, etc.)
* [x] `rig doctor` (toolchain/system health), with optional tool argument for focused diagnostics
* [x] `rig upgrade` (checksum-verified atomic self-update)
* [x] Windows-safe binary replacement

**Result:**
Users can now understand and repair their toolchain.

---

## 🔜 Phase 4 — Ephemeral Execution (v0.5)

**Theme:** Expand execution surface safely.

Planned:

* [ ] `rig x` (alias: `rix`, ephemeral execution, similar to `npx` but lock-aware)
* [ ] Safe remote tool execution model
* [ ] Strict isolation (no global pollution)
* [ ] Built-in alias (`rix`) for direct invocation

Constraints:

* Must not weaken determinism.
* Must not bypass lockfile guarantees.
* Must remain single-binary.

---

## 🔭 Future (Pre-v1.0 Research)

These are research areas, not commitments.

* Monorepo graph introspection
* Structured JSON output mode
* `--dry-run` for mutating commands
* Semver-aware upgrade comparisons
* Release signing / verification hardening
* Performance improvements for large toolchains

---

## Philosophy Guardrails

Every future feature must preserve:

1. Single binary distribution
2. Lockfile authority
3. Explicit over implicit
4. No hidden PATH mutation
5. No global state pollution
6. Deterministic behavior across machines

I label these as ideal invariants for `rig`.

---

## Release Model

* Development happens on `main`.
* Releases are immutable tags (`vX.Y.Z`).
* Tags trigger cross-platform builds.
* Installers consume GitHub release artifacts.
* `rig upgrade` pulls from `releases/latest`.

No version branches are required unless maintaining legacy versions.

---


**Goal:** Bridge the gap from development to deployment, solidifying `rig` as an end-to-end platform for the entire Go application lifecycle.

- [ ] **Production Supervisor (`rig start`):** A lightweight, production-ready process manager that handles graceful shutdowns, signal trapping, and secrets injection.
- [ ] **Project Scaffolding (`rig init --template`):** A powerful scaffolding engine that can generate new projects from Git templates (e.g., `rig init my-api --template=github:user/clean-arch-starter`).
- [ ] **Streamlined Publishing (`rig publish`):** A single command to automate the release workflow: tagging, cross-compiling, and creating a GitHub release with binaries.
- [ ] **Advanced Build Caching:** The "Turborepo" feature. For large monorepos, cache task outputs to avoid re-executing work on unchanged code.
- [ ] **Zig CC Integration:** Use Zig as a universal C compiler for cross-compiling CGo dependencies in a hermetic way, improving the experience of working with CGo-heavy projects. Also explore using Zig's caching and remote execution features for CGo builds.
- [ ] **Plugin System & IDE Integration:** Moonshot. Allow the community to extend `rig` and build first-class editor integrations. Not imminent, pretty or necessary in any way unless priorities change.

---

### How to Contribute

The golden question:

> Does this improve determinism, clarity, or toolchain integrity?

Open an issue:
[https://github.com/divijg19/rig/issues](https://github.com/divijg19/rig/issues)

Your feedback is crucial! If you have ideas for features, want to see a specific item prioritized, or are ready to contribute, please [open an issue](https://github.com/divijg19/rig/issues) and enable us to build the future of Go tooling together :)