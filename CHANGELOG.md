# Changelog

All notable changes to this project are documented in this file.

The format is based on Keep a Changelog, and this project aims to follow Semantic Versioning.

## v0.3.0

### Added
- `rig dev` (alias: `rid`): tool-backed dev loop driven by `[tasks.dev]` and `rig.lock`.
- Finalized alias model via invocation-name dispatch (`rig`/`rir`/`ric`/`rid`/`ris`) and informational `rig alias`.
- `docs/INSTALLATION.md` and installer script intended for `curl -fsSL https://rig.sh/install | sh`.

### Changed
- Task configuration is strict for determinism: tasks are either strings or tables with `{command, env, cwd, depends_on}`.
- `[tasks.dev]` additionally supports `watch` for `rig dev`.

### Notes
- `rig start` remains stubbed for future releases.
