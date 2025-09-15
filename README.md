## Tools

You can pin tools in `[tools]` in `rig.toml` using either full Go module paths or supported short names:

| Short Name      | Module Path                                 |
|-----------------|---------------------------------------------|
| golangci-lint   | github.com/golangci/golangci-lint/cmd/golangci-lint |
| mockery         | github.com/vektra/mockery/v2                |
| staticcheck     | honnef.co/go/tools/cmd/staticcheck          |
| revive          | github.com/mgechev/revive                   |

These short-name mappings are centrally maintained in `internal/rig/tooling.go`. If you’d like another common tool added, open an issue or PR to extend `ToolShortNameMap`.

You can also install tools from a file (like pip requirements.txt):

```
rig setup tools.txt
```
Where `tools.txt` contains lines like:
```
# name = version or module@version
golangci-lint = 1.62.0
mockery = v2.46.0
staticcheck = latest
github.com/vektra/mockery/v2@v2.46.0
```
Blank lines and # comments are ignored. If only a name is given, "latest" is used.
# rig

Project manager and task runner for Go — orchestrates the official toolchain with a single declarative manifest.

**rig = cargo’s clarity + bun’s DX + Go’s no-nonsense tooling**

[![build status](https://img.shields.io/github/actions/workflow/status/divijg19/rig/build.yml?branch=main)](https://github.com/divijg19/rig/actions)
[![latest release](https://img.shields.io/github/v/release/divijg19/rig)](https://github.com/divijg19/rig/releases)
[![license](https://img.shields.io/github/license/divijg19/rig)](./LICENSE)

`rig` is an opinionated, all-in-one project manager and task runner for Go. It enhances the native Go toolchain with a single, declarative manifest (`rig.toml`), solving common pain points like script cross-compatibility, reproducible tooling, and task discovery — without replacing `go build`, `go test`, or `go mod`.

### Key Features

• Unified Manifest: one `rig.toml` to rule them all
    - [project] metadata: name, version, authors, license
    - [tasks] to replace Makefile scripts
    - [tools] for version-locked dev tools
    - [profile.*] for build flags per environment (e.g., release)
    - include = ["..."] to split configs (e.g., monorepos via .rig/)

• Reproducible Tooling: project-local installs in `.rig/bin`
    - `rig setup` reads `[tools]` and installs with `GOBIN=.rig/bin`
    - `rig run` and `rig build` prepend `.rig/bin` to PATH
    - CI and team members are guaranteed to use the same tool versions

• Friendly DX: emoji output, clear errors, and cross-platform behavior
• Zero-config start: sensible defaults; customize as you go

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

3.  **Pin and use tools reproducibly:**
    ```toml
    # rig.toml
    [tools]
    golangci-lint = "1.62.0"
    github.com/vektra/mockery/v2 = "v2.46.0"
    ```
    ```bash
    rig setup        # installs into ./.rig/bin
    rig run lint     # now uses the pinned golangci-lint
    ```

    Prefer short names where supported:
    ```toml
    [tools]
    golangci-lint = "1.62.0"  # short name
    mockery = "v2.46.0"       # short name mapped to github.com/vektra/mockery/v2
    ```
    Verify without installing:
    ```bash
    rig setup --check
    # output shows ✅ when versions match or ❌ for mismatches/not-found
    ```
    ```

### Learn More

*   **[Philosophy](./PHILOSOPHY.md):** Understand the "why" behind `rig`.
*   **[Roadmap](./ROADMAP.md):** See where the project is headed.
*   **[Contributing](./CONTRIBUTING.md):** Learn how you can help build `rig`.

---

Made with ❤️ for the Go community, and dedicated to Tarushi, this project's origin.
