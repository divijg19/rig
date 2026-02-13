**Configuration Reference — rig.toml**

This document is a concise reference for the `rig.toml` manifest. It describes the top-level sections, supported fields, and how `rig` loads and composes configuration files in monorepo setups.

See also: **[CLI reference](./CLI.md)** for how configuration values are used by commands such as `rig build`, `rig sync` and `rig run`.

**Table of contents**
- Top-level sections
- Tasks schema
- Tools schema
- Build profiles
- Includes / Monorepos
- Examples

---

## Top-level sections

A `rig.toml` manifest supports the following top-level sections (most common):

- `[project]` — metadata about the project.
- `[tasks]` — named commands and structured tasks used by `rig run`.
- `[tools]` — pinned developer tools installed into `.rig/bin` via `rig sync`/`rig setup`.
- `[profile.<name>]` — build-time profiles used by `rig build --profile <name>`.
- `include` — optional list of additional TOML files to include (see "Includes / Monorepos").

### `[project]`
Fields:
- `name` (string): project name.
- `version` (string): semantic version string (conventional default `0.1.0`).
- `authors` (array[string]): list of author strings.
- `license` (string): SPDX or free-form license identifier.

Example:

```toml
[project]
name = "my-service"
version = "0.1.0"
authors = ["You <you@example.com>"]
license = "MIT"
```

---

## `[tasks]` — task schema

Tasks are the primary developer-facing entrypoints.

`rig` supports two task styles:

1. Simple string (common):
  - `test = "go test ./..."`
2. Structured table (strict schema):

Supported fields for a structured task table:
- `command` (string, required): command string to execute.
- `description` (string, optional): human description shown by `rig run --list`.
- `env` (table[string], optional): map of KEY=VALUE environment variables.
- `cwd` (string, optional): working directory, resolved relative to the `rig.toml` directory.
- `depends_on` (array[string], optional): tasks to run before this task.

v0.3 adds one special-case field:
- `[tasks.dev].watch` (array[string], required for `rig dev`): file watch globs used by the watcher tool.

Notes:
- `depends_on` values are validated and resolved in deterministic topological order; cycles error.
- `rig run` and `rig dev` require `rig.lock` and will fail fast if it is missing.

Examples:

```toml
[tasks.build]
command = "go build -o bin/server ./cmd/server"

[tasks.dev]
command = "go run ."
watch = ["**/*.go"]

[tasks.release]
command = "./scripts/release.sh"
depends_on = ["build", "test"]
```

Use `rig run <task>` to execute tasks.

---

## `[tools]` — pin developer tools

The `[tools]` section allows pinning tool versions used for development and CI. `rig` installs these into `.rig/bin` using `go install`.

Key points:
- Keys may be short names (mapped by `internal/rig/tooling.go`) or full Go module paths.
- Values are versions; `latest` is supported.

Examples:

```toml
[tools]
golangci-lint = "1.62.0"
github.com/vektra/mockery/v2 = "v2.46.0"
```

Tool resolution rules:
- `rig` maps short names (e.g. `golangci-lint`) to canonical module paths for `go install`.
- When you run `rig sync`, `rig` resolves tools deterministically and writes `rig.lock` (schema=0) next to `rig.toml`.
- `rig sync` installs the resolved `module@version` pins into `.rig/bin` and also writes `.rig/manifest.lock` (a hash cache) for quick drift detection.
- For CI, use `rig sync --check --json` or `rig sync --check` to verify `rig.lock` and installed tools.
- For hermetic/offline environments, use `rig sync --offline` (fails if required modules are not already in the module cache).

---

## Build profiles (`[profile.<name>]`)

Define reusable build configuration blocks applied by `rig build`.

Supported fields for a `BuildProfile`:
- `ldflags` (string): passed to `go build -ldflags`.
- `gcflags` (string): passed to `go build -gcflags`.
- `tags` (array[string]): build tags for `go build -tags`.
- `flags` (array[string]): general extra flags.
- `env` (table[string]): environment variables to apply during build (e.g., `GOCACHE` overrides).
- `output` (string): default output path for binary.

Example:

```toml
[profile.release]
ldflags = "-s -w"
gcflags = ""
tags = []
output = "bin/myapp"
```

`rig build --profile release` will merge CLI overrides with profile values.

---

## Includes and Monorepos

`rig` supports splitting configuration across files via the `include` key (array of relative paths). Example:

```toml
include = ["rig.tasks.toml", "rig.tools.toml"]
```

Loader behavior (from `internal/config/loader.go`):
- Paths are resolved relative to the base `rig.toml` directory.
- If an include path is not present next to `rig.toml`, `rig` will attempt to find it under `.rig/<include>` (useful for monorepos where shared pieces are placed in `.rig/`).
- Included files are merged in the following way:
  - `tasks` entries are merged into the root `tasks` map (new keys override earlier ones)
  - `tools` entries are merged into the root `tools` map
  - `profile` entries are merged into `Profiles`
- Use includes when you have many projects sharing tasks/tools (monorepo), or when you want to separate auto-generated or machine-managed fragments (`.rig/`) from hand-edited top-level config.

Recommended layout for monorepos:

- Root `rig.toml` contains `[project]` and profile definitions and an `include` listing files under `.rig/`.
- Put shared or generated tasks and tools in `.rig/rig.tasks.toml` and `.rig/rig.tools.toml`.

Example monorepo structure:

```
my-monorepo/
  rig.toml           # includes .rig/rig.tasks.toml and rig.tools.toml
  packages/serviceA/
  packages/serviceB/
  .rig/rig.tasks.toml
  .rig/rig.tools.toml
```

---

## Examples

- Minimal single-module manifest: `examples/basic/rig.toml`
- Monorepo example: `examples/monorepo/rig.toml`

See `docs/CLI.md` for how the CLI uses these configuration sections.
