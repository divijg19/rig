CLI Cheatsheet — Quick Reference
===============================

A compact page of common `rig` commands and recommended invocations for development and CI.

Local development
-----------------

- Initialize a new project interactively:

```sh
rig init
```

- Create a developer scaffold and install tools:

```sh
rig init --developer
rig sync
rig run dev
```

- List tasks (human):

```sh
rig run --list
```

- List tasks (JSON for editors / automation):

```sh
rig check
```

- Run a task with extra environment variables:

```sh
rig run build -E FOO=bar -E BAZ=qux
```

- Dry-run to see what will execute:

```sh
rig build --dry-run
rig run test --dry-run
```

Ephemeral tools (npx-style)
---------------------------

Run a one-off tool without committing it to `[tools]`:

```sh
rig x golangci-lint@v1.62.0 run ./...
rig x mockery -- --help
```

Tools management
----------------

- Install/update pinned tools (writes `rig.lock` + `.rig/manifest.lock`):

```sh
rig sync    # shortcut for `rig tools sync`
```

- Verify tools without installing (good for CI):

```sh
rig sync --check

# Machine readable (CI):
rig sync --check --json | jq .

- Hermetic/offline (no downloads; requires module cache):

```sh
rig sync --offline
rig sync --check --offline --json | jq .
```

- List missing/outdated tools (human):

```sh
rig outdated
```

- List missing/outdated tools (JSON):

```sh
rig outdated --json
```

CI snippet (GitHub Actions)
----------------------------

Use this minimal step to assert that the project's pinned tools match the lockfile and fail the workflow if they don't.

```yaml
# .github/workflows/rig-check.yml (excerpt)
- name: Verify rig tools
  run: |
    rig sync --check --json > rig-tools.json
    cat rig-tools.json
  shell: bash
```

Build and Release
------------------

- Build with a named profile:

```sh
rig build --profile release
```

- Override output path:

```sh
rig build --profile release -o bin/myapp
```

Quick tips
----------

- Use `rig run --list` to discover project tasks.
- For CI, prefer the `--json` outputs from `rig sync --check` and `rig outdated` for stable, machine-parsable assertions.

See `docs/CLI.md` and `docs/CONFIGURATION.md` for complete command and configuration references.
