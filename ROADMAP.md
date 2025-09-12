# Project `rig` Roadmap

This document outlines the planned development for `rig`. It is a living document and priorities may shift based on community feedback.

Our development is planned in phases, with each phase delivering a significant set of valuable features.

### Phase 1: v0.1 - The Foundation (The `Makefile` Killer)
**Goal:** Provide an immediate, valuable replacement for `Makefile` with a native, cross-platform solution.
- [ ] Define the initial `rig.toml` specification (`[project]`, `[tasks]`).
- [ ] Implement the `rig run <task>` command.
- [ ] Implement the `rig init` command for bootstrapping.
- [ ] Basic support for build profiles via `[profile]`.

### Phase 2: v0.2 - The Differentiator (Solving Toolchain Hell)
**Goal:** Solve the problem of reproducible tooling to make `rig` an indispensable part of the Go workflow.
- [ ] Implement the `[tools]` section in `rig.toml`.
- [ ] Build the transparent toolchain installation and management logic.
- [ ] Create the high-level `rig check` command for project validation.

### Phase 3: v0.5 - The DX Revolution (The `bun` Experience)
**Goal:** Elevate the user experience to be on par with the best modern toolchains.
- [ ] Build `rig test`: the enhanced, interactive test runner.
- [ ] Implement `rig run --watch`: a native, configuration-driven file watcher.
- [ ] Build `rig new`: a powerful scaffolding engine using Git templates.
- [ ] Add support for task dependencies.

### Phase 4: v1.0 and Beyond - The Ecosystem
**Goal:** Solidify `rig` as the de facto project manager for the modern Go ecosystem.
- [ ] Full monorepo support with deep Go Workspace integration.
- [ ] Explore a plugin system for custom command extensions.
- [ ] Research advanced build caching inspired by Turbopack.
- [ ] Design `rig publish` for streamlined release workflows.

### How to Contribute

Your feedback is crucial! If you have ideas for features or want to see a specific item prioritized, please [open an issue](https://github.com/divijg19/rig/issues) and let's discuss it.
