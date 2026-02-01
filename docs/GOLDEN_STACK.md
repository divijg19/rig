**The Golden Stack: Go + Templ + Tailwind + Flutter**

This recipe shows how to configure a `rig.toml` in a monorepo to orchestrate a modern full-stack development workflow:
- Go backend (server)
- Templ (static site/templating generation)
- Tailwind CSS (asset pipeline / watcher)
- Flutter frontend (mobile/web)

The example below is derived from the monorepo sandbox and provides a copy-pasteable manifest you can adapt to your repository layout.

---

Why this stack?
- Go: fast backend and API server.
- Templ: static site generation or codegen step used by many Go projects.
- Tailwind: modern utility-first CSS with a file-watching dev loop.
- Flutter: single codebase for web & mobile UI.

This `rig.toml` demonstrates how to run and compose these pieces using `rig` tasks and tools.

```toml
[project]
name = "golden-stack"
version = "0.1.0"
license = "MIT"

[profile.dev]
flags = ["-race"]

[profile.release]
ldflags = "-s -w"
output = "bin/golden"

[tasks]
# Basic tasks
list = "rig run --list"
help = "rig --help"

# Backend
backend.build = "go build -o ./bin/backend ./cmd/backend"
backend.run = "go run ./cmd/backend"

# Templ (assumes `templ` is a tool available via `rig x templ` or installed in tools)
templ.build = "templ generate --out ./static"

# Tailwind (assumes node/tailwind is available)
assets.watch = "npx tailwindcss -i ./assets/input.css -o ./static/css/output.css --watch"
assets.build = "npx tailwindcss -i ./assets/input.css -o ./static/css/output.css --minify"

# Flutter
flutter.run = "flutter run -d web"
flutter.build = "flutter build web --release"

# Composite: run all dev services concurrently (POSIX shell example)
# Note: On Windows, prefer using `powershell -Command "Start-Process ... -NoNewWindow -Wait"` or a cross-platform supervisor
dev-all = "sh -c \"rig run templ.build & rig run assets.watch & rig run backend.run & rig run flutter.run & wait\""

# Convenience task for CI: build everything
dist = "sh -c \"rig run templ.build && rig run assets.build && rig build --profile release && rig run flutter.build\""

[tools]
# Pin toolchain and important tools
go = "1.21.5"
# Node is not installed via go tools; assume CI installs node separately for tailwind/flutter

# Helpful dev tools
golangci-lint = "1.62.0"
# templ might be a small Go-based generator; if so, map it via short name or module
# templ = "latest"  # replace with actual module path if applicable

# Flutter and Tailwind are not Go modules; install them with system package managers or CI scripts
```

Notes & recommendations
- `dev-all` runs multiple processes in parallel via a shell pipeline which works for UNIX-like dev environments. For cross-platform local dev, prefer a small supervisor script or use `tmux`/editor-integrated tasks.
- Use `rig x <tool>` for ephemeral runs when you don't want to add tools to the repository `tools` section. Example: `rig x golangci-lint@v1.62.0 run ./...`.
- For CI, ensure Node and Flutter are available in the runner environment before calling `rig run flutter.build` or `assets.build`.

See `docs/CONFIGURATION.md` for more on task syntax, and `docs/CLI.md` for commands to drive this workflow.
