Basic example
==============

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
rig sync    # (optional) install tools if you added any
rig run --list
rig run test
```

When to use this example

- Small services and libraries
- When you prefer a single-file manifest and minimal tooling

Customization

- Add pins to `[tools]` (see `docs/CONFIGURATION.md`) and run `rig sync`.
- Add structured tasks using `argv`, `command`, and `depends_on` for complex flows.
