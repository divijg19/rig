Basic example
=============

Location: `examples/basic/rig.toml`

This minimal example demonstrates a single-module Go project with sensible defaults:

- `[project]` metadata
- `[tasks]`: `build`, `test`, `run`, `fmt`, `vet`
- `[profile.release]` for reproducible release builds
- `[tools]` left empty (add pins for reproducible dev/CI)

Quick start

```pwsh
# From the example directory
cd examples/basic

# Copy into a real project or use as a template
cp rig.toml /path/to/your/project/rig.toml
cd /path/to/your/project

# Optional: install pinned tools if you add any
rig sync

# Inspect available tasks
rig run --list

# Run verification
rig run test
