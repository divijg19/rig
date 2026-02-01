# CLI entrypoint

This folder contains the only `package main` file (`main.go`).

Responsibility:
- Call `cli.Execute()` from `internal/cli`.
- No business logic or flags here; all wiring and commands live in `internal/cli`.
