# Contributing to `rig`

First off, thank you for considering contributing! `rig` is a community-driven project, and every contribution, from a bug report to a new feature, is valuable.

### How Can I Contribute?

#### Reporting Bugs

If you find a bug, please create a [bug report issue](https://github.com/divijg19/rig/issues/new?assignees=&labels=bug&template=bug_report.md). Be sure to include:
*   Your `rig` version (`rig --version`).
*   Your Go version (`go version`).
*   Your operating system.
*   A clear description of the bug and steps to reproduce it.

#### Suggesting Enhancements

If you have an idea for a new feature or an improvement, please open a [feature request issue](https://github.com/divijg19/rig/issues/new?assignees=&labels=enhancement&template=feature_request.md). This lets us discuss the idea before any code is written.

#### Your First Code Contribution

Ready to write some code? Here’s how to get started:

1.  **Fork the repository** on GitHub.
2.  **Clone your fork** to your local machine: `git clone https://github.com/<your-username>/rig.git`
3.  **Create a new branch** for your changes: `git checkout -b feature/my-awesome-feature`
4.  **Make your changes.** Ensure you adhere to the code style.
5.  **Add tests** for your changes. We take testing seriously.
6.  **Ensure code builds and vets cleanly:** `rig run build` then `rig run vet`
7.  **Commit your changes** with a clear and descriptive commit message.
8.  **Push your branch** to your fork: `git push origin feature/my-awesome-feature`
9.  **Open a Pull Request** against the `main` branch of `divijg19/rig`.

### Style Guide

*   All Go code must be formatted with `gofmt`.
*   Linting: we aim to add `golangci-lint` in a future phase. For now, ensure `go vet` is clean and code is formatted.

### Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](./CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.
