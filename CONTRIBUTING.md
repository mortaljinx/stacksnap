# Contributing to StackSnap

Thanks for your interest in contributing. StackSnap is a focused tool — contributions that keep it simple and self-hosted-friendly are most welcome.

---

## What's Welcome

- Bug fixes
- Improved error messages
- ARM / platform compatibility fixes
- Documentation improvements
- Docker fallback reconstruction improvements

## What to Discuss First

If you want to add a significant new feature, open an issue first so we can talk through whether it fits the scope of the project. StackSnap is intentionally small — a persistent daemon mode, a web UI, or a database backend are examples of things that probably don't belong here.

---

## Getting Started

```sh
git clone https://github.com/<YOUR_GITHUB_USERNAME>/stacksnap
cd stacksnap
go mod tidy
go build .
```

Requirements: Go 1.23+, Docker installed locally for testing the fallback path.

---

## Before Submitting a PR

- `go mod tidy` should be clean
- `go build ./...` should succeed with no errors
- Test against a real Portainer instance if touching the Portainer client
- Test the Docker fallback path if touching `internal/docker`
- Keep commits focused — one thing per PR where possible

---

## Reporting Bugs

Open a GitHub issue with:
- Your OS and architecture
- How you're running StackSnap (binary or Docker)
- Whether you're using Portainer or Docker fallback
- The exact error output
- Relevant log lines

---

## License

By contributing you agree your changes will be licensed under the MIT License.
