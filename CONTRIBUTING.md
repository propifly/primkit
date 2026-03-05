# Contributing to Primkit

Thank you for considering a contribution. This guide explains how to set up the project, run tests, and submit changes.

## Development Setup

**Requirements:** Go 1.22+

```bash
git clone git@github.com:propifly/primkit.git
cd primkit
make build   # builds bin/taskprim, bin/stateprim, bin/knowledgeprim, bin/queueprim
make test    # runs all tests with race detection
```

The project uses a Go workspace (`go.work`) with five modules:
- `primkit/` — shared library (config, auth, db, server, replicate)
- `taskprim/` — task management primitive
- `stateprim/` — state persistence primitive
- `knowledgeprim/` — knowledge graph primitive
- `queueprim/` — work queue primitive

## Running Tests

```bash
make test           # all modules, verbose, race detector
make build          # compile both binaries
```

Individual modules:
```bash
cd taskprim && go test ./...
cd stateprim && go test ./...
cd knowledgeprim && go test ./...
cd queueprim && go test ./...
cd primkit && go test ./...
```

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Each CLI command lives in its own file (e.g., `add.go`, `list.go`)
- Pattern: parse flags -> call store -> format output
- Tests use `testify/assert` and `testify/require`
- In-memory SQLite for all tests (`db.OpenInMemory()`)

## Commit Messages

Use clear, imperative commit messages:
- `add restore command for point-in-time recovery`
- `fix duplicate JSON encoding in API test`
- `wire litestream replication into serve lifecycle`

One logical change per commit. Keep commits focused.

## Pull Requests

1. Fork the repo and create a branch from `main`
2. Write tests for new functionality
3. Ensure `make test` passes with no failures
4. Ensure `make build` compiles cleanly
5. Open a PR with a clear description of what changed and why

## Architecture Notes

Before making significant changes, read `docs/architecture.md` to understand the layered design. The key constraint: store interfaces are the boundary — CLI, API, and MCP are sibling consumers that never depend on each other.

## Reporting Issues

Use GitHub Issues. For bugs, include:
- What you expected to happen
- What actually happened
- Steps to reproduce
- Go version and OS
