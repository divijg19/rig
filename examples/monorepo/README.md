# Monorepo example
=================

Location: `examples/monorepo/rig.toml`

This example demonstrates a realistic monorepo layout where shared tasks and tools would commonly live under `.rig/` and be referenced via `include` from the root `rig.toml`.

Key ideas demonstrated:

- Separate `[profile.*]` blocks for `dev` and `release`.
- Shared tasks and tools can be placed in `.rig/rig.tasks.toml` and `.rig/rig.tools.toml`.
- Developer commands like `dev` use a file-watching tool (e.g. `reflex`) to re-run on change.

Quick start

```pwsh
# Copy the example into a repo root
cp examples/monorepo/rig.toml /path/to/monorepo/rig.toml

# Optionally create .rig/ and split tasks/tools into the include files
mkdir -p .rig

# Then run:
rig sync
rig run dev
```

CI and reproducibility

- `rig sync` writes `rig.lock` next to `rig.toml` and installs tools into `.rig/bin`.
- Use `rig sync --check --json` in CI to validate `rig.lock` matches `rig.toml` and that the pinned tools are installed.
- In CI you can fail the job when the JSON payload indicates mismatches.

Notes

- The example inlines `.rig/` fragments for readability — in real monorepos we prefer the include pattern to keep the root manifest concise.
