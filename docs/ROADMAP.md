# Project `rig` Roadmap

This document outlines the planned development for `rig`, reflecting our ambition to deliver a best-in-class developer environment for Go. It is a living document, and priorities may shift based on community feedback.

Our development is planned in phases, each with a clear theme and a goal to deliver significant, tangible value.

---

### ✅ **Phase 1: The Declarative Foundation (v0.1.0-alpha)**

**Goal:** Establish `rig` as a superior, cross-platform replacement for `make` and shell scripts by providing a single, declarative manifest.

- [x] **Interactive `rig init`:** Intelligent, interactive setup wizard.
- [x] **Core Task Runner (`rig run`):** Robust executor for simple and structured tasks.
- [x] **Build Profiles (`rig build`):** `go build` orchestration with flags from `[profile]` tables.

---

### ✅ **Phase 2: The Reproducible Workspace (v0.1 - v0.4)**

**Goal:** Solve the "it works on my machine" problem, making `rig` a critical tool for team collaboration and reliable CI/CD.

- [x] **Tooling Engine (`rig tools sync`):** Sandboxed, version-locked tool installation.
- [x] **Lockfile Generation (`rig.lock`):** Deterministic tool resolution and installation.
- [x] **PATH Injection:** Automatic use of project-local tools.
- [x] **CI Verification (`rig check` / `rig sync --check`):** Verify environment sync against `rig.lock`.
- [x] **Tool Maintenance (`rig outdated`):** Check tools that are missing or mismatched.

---

### 🏃 **Phase 3: The Virtual Runtime Experience (v0.5 - v0.8)**

**Goal:** Deliver the "Bun DX" by transforming `rig` from a task runner into a true development environment with a lightning-fast feedback loop. This is the current focus.

- [ ] **Native File Watching & Hot-Reloading (`rig dev`):** Implement a native, high-performance file watcher to drive `rig dev` and `rig test --watch`. This is the flagship feature.
- [ ] **Polished Test Runner (`rig test`):** A dedicated, rich TUI for running and visualizing test results, inspired by `bun test`.
- [ ] **Ephemeral Runner (`rig x`):** Implement the `npx`-style on-the-fly tool runner.
- [ ] **Monorepo Support (`include` directive):** Enhance the configuration loader to support splitting `rig.toml` across multiple files.
- [ ] **JSON Output & Argument Passthrough:** Implement scripting-friendly output and flexible argument handling (`--`).

---

### 🚀 **Phase 4: The Production & Ecosystem Platform (Before v1.0)**

**Goal:** Bridge the gap from development to deployment, solidifying `rig` as an end-to-end platform for the entire Go application lifecycle.

- [ ] **Production Supervisor (`rig start`):** A lightweight, production-ready process manager that handles graceful shutdowns, signal trapping, and secrets injection.
- [ ] **Project Scaffolding (`rig new`):** A powerful scaffolding engine that can generate new projects from Git templates (e.g., `rig new my-api --template=github:user/clean-arch-starter`).
- [ ] **Streamlined Publishing (`rig publish`):** A single command to automate the release workflow: tagging, cross-compiling, and creating a GitHub release with binaries.
- [ ] **Advanced Build Caching:** The "Turborepo" feature. For large monorepos, cache task outputs to avoid re-executing work on unchanged code.
- [ ] **Plugin System & IDE Integration:** The long-term vision. Allow the community to extend `rig` and build first-class editor integrations.

---

### How to Contribute

Your feedback is crucial! If you have ideas for features, want to see a specific item prioritized, or are ready to contribute, please [open an issue](https://github.com/divijg19/rig/issues) and let's build the future of Go tooling together.
