CLI Cheatsheet — Quick Reference
===============================

A compact page of common `rig` commands and recommended invocations for development and CI.

Local development
-----------------
- Initialize a new project interactively:

```pwsh
rig init
```

- Create a developer scaffold and install tools:

```pwsh
rig init --developer
rig sync
rig run dev
```

- List tasks (human):

```pwsh
rig ls
```

- List tasks (JSON for editors / automation):

```pwsh
rig run --list --json
```

- Run a task with extra environment variables:

```pwsh
rig run build -E FOO=bar -E BAZ=qux
```

- Dry-run to see what will execute:

```pwsh
rig build --dry-run
rig run test --dry-run
```

Ephemeral tools (npx-style)
---------------------------
Run a one-off tool without committing it to `[tools]`:

```pwsh
rig x golangci-lint@v1.62.0 run ./...
rig x mockery -- --help
```

Tools management
----------------
- Install/update pinned tools (writes `.rig/manifest.lock`):

```pwsh
rig sync    # shortcut for `rig tools sync`
```

- Verify tools without installing (good for CI):

```pwsh
rig sync --check
# Machine readable (CI):
rig sync --check --json | jq .
```

- List missing/outdated tools (human):

```pwsh
rig outdated
```

- List missing/outdated tools (JSON):

```pwsh
rig outdated --json
```

CI snippet (GitHub Actions)
---------------------------
Use this minimal step to assert that the project's pinned tools match the lockfile and fail the workflow if they don't.

```yaml
# .github/workflows/rig-check.yml (excerpt)
- name: Verify rig tools
  run: |
    rig sync --check --json > rig-tools.json
    cat rig-tools.json
  shell: bash
```

You can parse `rig-tools.json` to assert `missing == 0 && mismatched == 0` if you need structured checks.

Build and Release
-----------------
- Build with a named profile:

```pwsh
rig build --profile release
```

- Override output path:

```pwsh
rig build --profile release -o bin/myapp
```

Quick tips
----------
- Use `rig ls` as a low-friction way to discover project tasks.
- For CI, prefer the `--json` outputs from `rig sync --check` and `rig outdated` for stable, machine-parsable assertions.
- Put shared tasks/tools in `.rig/` and use `include` for clean monorepo manifests.

See `docs/CLI.md` and `docs/CONFIGURATION.md` for complete command and configuration references.
