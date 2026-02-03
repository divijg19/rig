# Examples
========

This directory contains small, copy-pasteable `rig.toml` manifests that demonstrate common project layouts and workflows.

Included examples:

- `basic/rig.toml` — minimal single-module project. Good starting point for small services and libraries.
- `monorepo/rig.toml` — merged monorepo example showing `.rig/` include-style layout inlined for clarity.

How to use an example locally

1. Copy the desired `rig.toml` into the root of a working directory (or edit it in place):

```sh
cp examples/basic/rig.toml /path/to/your/project/rig.toml
```

2. Initialize or inspect the manifest:

```sh
rig run --list
```

3. Sync tools and run tasks:

```sh
rig sync           # installs pinned tools into .rig/bin
rig run dev        # run a development task (if present)
rig build --profile release
```

Notes

- For monorepo examples, prefer placing shared fragments in `.rig/rig.tasks.toml` and `.rig/rig.tools.toml` and using `include = [".rig/rig.tasks.toml", ".rig/rig.tools.toml"]` in the root `rig.toml`.
- `rig sync` writes `rig.lock` (deterministic tool pins) next to `rig.toml`.
- Use `rig sync --check --json` in CI to verify `rig.lock` + installed tools and fail builds on mismatch.
