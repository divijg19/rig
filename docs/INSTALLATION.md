**Installation (v0.3)**

`rig` ships as one binary.

Aliases (`rir`/`ric`/`rid`/`ris`) are optional and work by invocation name (argv[0]). `rig` does not auto-create symlinks.

---

## Shell installer

Single line:
```sh
curl -fsSL https://rig.sh/install | sh
```

This installer is expected to download the `rig` binary for your platform and place it on your `PATH`.

Optional: create aliases (no auto-symlinks):
```sh
ln -sf "$(command -v rig)" "$HOME/.local/bin/rir"
ln -sf "$(command -v rig)" "$HOME/.local/bin/ric"
ln -sf "$(command -v rig)" "$HOME/.local/bin/rid"
ln -sf "$(command -v rig)" "$HOME/.local/bin/ris"
```

---

## Go install

Install the single main binary:
```sh
go install github.com/divijg19/rig/cmd/rig@v0.3.0
```

Install all reserved entrypoints intentionally:
```sh
go install github.com/divijg19/rig/cmd/...@v0.3.0
```

Local development from the repo:
```sh
go install ./cmd/...
```

---

## Notes

- Ensure `$GOPATH/bin` (or `$(go env GOPATH)/bin`) is in your `PATH`.
- Use `rig alias` to see the reserved alias list and how invocation-name dispatch works.
