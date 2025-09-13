# Installing `rig`

This document provides detailed instructions for installing `rig` on various platforms.

### Recommended Method: `go install`

The simplest and recommended way to install `rig` is using the `go install` command, which works on any platform with the Go toolchain installed.

```bash
The recommended way to install `rig` is with the Go toolchain.

## Using `go install`

Works on all platforms with Go installed:

```bash
go install github.com/divijg19/rig@latest
```

Windows PATH note: the binary is placed under `%GOPATH%\\bin` (or `%USERPROFILE%\\go\\bin`). Ensure this directory is on your PATH to run `rig` from any shell.

## Verify

```bash
rig --help
```

If the command is not found, open a new shell or update your PATH.
