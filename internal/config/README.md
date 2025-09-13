# Config: loading rig.toml

This package contains:
- Data models for the `rig.toml` manifest (see `config.go`).
- A Viper-based loader that searches upward from CWD and unmarshals into `Config` (see `loader.go`).

Contract:
- Look for `rig.toml` in the current directory or any parent.
- On success, return `(*Config, path, nil)`; otherwise return a helpful error.
- Do not own business logic (just loading/validation). Consumers live in `internal/cli` or future `internal/rig`.
