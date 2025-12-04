# Installing `rig`

This guide shows the recommended way to install `rig` using the Go toolchain.

## Recommended: `go install`

Works on all platforms with Go installed:

```pwsh
go install github.com/divijg19/rig@latest
```

### PATH notes

- Windows: the binary is placed under `%GOPATH%\bin` or `%USERPROFILE%\go\bin`. Ensure this folder is in your PATH so you can run `rig` from any shell (PowerShell, Command Prompt, etc.). You may need to restart your terminal.
- macOS and Linux: the binary is placed under `$GOPATH/bin` or `$HOME/go/bin`. Add this directory to your shell PATH (e.g., in `~/.zshrc`, `~/.bashrc`, or `~/.profile`).

Example (bash/zsh):

```bash
export PATH="$HOME/go/bin:$PATH"
```

## Verify

```pwsh
rig --help
rig ls -j
rig sync --check --json
```

If a command is not found, open a new shell or update your PATH as noted above.
