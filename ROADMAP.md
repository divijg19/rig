# Project `rig` Roadmap

This document outlines the planned development for `rig`, reflecting our ambition to deliver a best-in-class developer experience for Go. It is a living document, and priorities may shift based on community feedback.

Our development is planned in phases, each with a clear theme and a goal to deliver significant, tangible value.

---

### ✅ **Phase 1: The Core Utility (v0.1 - v0.3)**

**Goal:** To establish `rig` as a superior, cross-platform replacement for `make` and shell scripts. This phase focused on the core workflow of defining and running tasks.

- [x] **Interactive `rig init`:** Create an intelligent, interactive setup wizard with a `--yes` flag for automation.
- [x] **Core Task Runner (`rig run`):** Implement a robust executor for simple and structured tasks.
- [x] **Structured Tasks:** Support for task tables with `description`, `env`, and `depends_on`.
- [x] **Build Profiles (`rig build`):** Implement `go build` orchestration with flags from `[profile]` tables.

---

### ✅ **Phase 2: The Reproducible Workspace (v0.4 - v0.6)**

**Goal:** To solve the "it works on my machine" problem once and for all, making `rig` a critical tool for team collaboration and reliable CI/CD.

- [x] **Tooling Engine (`rig tools sync`):** Implement the core logic for the `[tools]` manifest, installing version-locked tools into a project-local `.rig/bin`.
- [x] **Lockfile Generation:** Create the `.rig/manifest.lock` file to guarantee deterministic tool installation.
- [x] **PATH Injection:** Ensure the `.rig/bin` directory is always prepended to the `PATH` for any executed task.
- [x] **Verification (`rig check`):** Implement the CI-focused command to verify that the environment is in sync with the lockfile.
- [x] **Tool Maintenance (`rig tools outdated`):** Add a command to check defined tools against their latest available versions.
- [x] **Shortcuts & Usability:** Implement all documented shortcuts (`sync`, `check`, `outdated`, `ls`) and global flags (`--dry-run`, `-C, --dir`).

---

### 🏃 **Phase 3: The Modern DX Powerhouse (v0.7 - v0.9)**

**Goal:** To deliver the "wow" factor features that make `rig` feel like a best-in-class tool from the broader software ecosystem, focusing on speed and convenience.

- [ ] **Ephemeral Runner (`rig x`):** Implement the `npx`-style on-the-fly tool runner. This is the flagship feature of this phase and a massive DX win.
- [ ] **Monorepo Support (`include` directive):** Enhance the configuration loader to support splitting `rig.toml` across multiple files, enabling clean monorepo workflows.
- [ ] **JSON Output (`--json` flag):** Implement structured JSON output for commands like `ls` and `outdated` to make `rig` more scriptable.
- [ ] **Argument Passthrough:** Implement the ability to pass extra arguments to a task using `--`.

---

### 🚀 **Phase 4: The Ecosystem & The Future (v1.0 and Beyond)**

**Goal:** To build upon the stable foundation with features that create an unparalleled development feedback loop and solidify `rig` as a true platform.

- [ ] **The Polished Test Runner (`rig test`):** A dedicated, rich TUI for running and visualizing test results, inspired by `bun test` and Jest. This will include performance insights and rich diffs.
- [ ] **Native File Watching (`--watch` mode):** The ultimate fast feedback loop. Implement a native, configuration-driven file watcher for `rig run dev` and `rig test`.
- [ ] **Project Scaffolding (`rig new`):** A powerful scaffolding engine that can generate new projects from Git templates.
- [ ] **Streamlined Publishing (`rig publish`):** A single command to automate the entire release workflow: tagging, cross-compiling, creating a GitHub release, and uploading binaries.
- [ ] **Advanced Build Caching:** The "Turborepo" feature. For large monorepos, cache task outputs to avoid re-executing work on unchanged code, dramatically speeding up CI builds.
- [ ] **Plugin System & IDE Integration:** The long-term vision. Allow the community to extend `rig` with custom commands and build first-class editor integrations.

---

### How to Contribute

Your feedback is crucial! If you have ideas for features, want to see a specific item prioritized, or are ready to contribute, please [open an issue](https://github.com/divijg19/rig/issues) and let's build the future of Go tooling together.
