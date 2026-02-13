**Installation (v0.3)**

`rig` ships as one binary.

Aliases (`rir`/`ric`/`rid`/`ris`) are optional and work by invocation name (argv[0]).

- On macOS/Linux, the shell installer creates these aliases as symlinks next to `rig`.
- On Windows, invoke `rig run`, `rig check`, etc directly.

---

## Shell installer

Single line:
```sh
curl -fsSL https://rig.sh/install | sh
```

This installer is expected to download the `rig` binary for your platform and place it on your `PATH`.

On macOS/Linux, the installer also creates these optional symlink entrypoints:

- `rir` → `rig run`
- `ric` → `rig check`
- `rid` → `rig dev`
- `ris` → `rig start`

---

## Go install

Install the single main binary:
```sh
go install github.com/divijg19/rig/cmd/rig@v0.3
```

Local development from the repo:
```sh
go install ./cmd/rig
```

## Installer vs go install behavior

- `go install` installs only the `rig` binary.
- `install.sh` installs `rig` and creates alias symlinks (`rir`, `ric`, `rid`, `ris`) on macOS/Linux.

Windows note:

- Until a PowerShell installer exists, Windows users should invoke `rig run` / `rig check` / `rig dev` directly.

---

## Notes

- Ensure `$GOPATH/bin` (or `$(go env GOPATH)/bin`) is in your `PATH`.
- Use `rig alias` to see the reserved alias list and how invocation-name dispatch works.

If you installed via `go install` and want aliases, create symlinks manually:
```sh
ln -sf "$(command -v rig)" "$HOME/.local/bin/rir"
ln -sf "$(command -v rig)" "$HOME/.local/bin/ric"
ln -sf "$(command -v rig)" "$HOME/.local/bin/rid"
ln -sf "$(command -v rig)" "$HOME/.local/bin/ris"
```
