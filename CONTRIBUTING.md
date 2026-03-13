# Contributing to Primkit

Thank you for considering a contribution. This guide explains how to set up the project, run tests, and submit changes.

## Development Setup

**Requirements:** Go 1.26+, [golangci-lint](https://golangci-lint.run/) v2, [gofumpt](https://github.com/mvdan/gofumpt)

```bash
git clone git@github.com:propifly/primkit.git
cd primkit
make all   # tidy → fmt → lint → test → build → docs-check → check-registration
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
make build          # compile all four binaries
make lint           # golangci-lint across all modules
make fmt            # gofumpt across all modules
make all            # full pre-PR checklist (recommended before pushing)
```

Individual modules:
```bash
cd taskprim && go test ./...
cd stateprim && go test ./...
cd knowledgeprim && go test ./...
cd queueprim && go test ./...
cd primkit && go test ./...
```

### Golden File Tests

Each prim has golden file tests that snapshot CLI output for regression testing. To update golden files after intentional output changes:

```bash
cd taskprim && go test -run TestGolden -update ./internal/cli/
cd stateprim && go test -run TestGolden -update ./internal/cli/
cd knowledgeprim && go test -run TestGolden -update ./internal/cli/
cd queueprim && go test -run TestGolden -update ./internal/cli/
```

Review the diff in `internal/cli/testdata/*.golden` before committing — golden files should only change when output format changes intentionally.

### Fuzz Testing

The FTS5 query sanitizer in knowledgeprim has a fuzz test. Run it locally for deeper coverage:

```bash
cd knowledgeprim && go test -fuzz=FuzzSanitizeFTS5Query -fuzztime=60s ./internal/store/
```

CI runs this for 30 seconds on every push.

## Code Style

- **Formatter:** [gofumpt](https://github.com/mvdan/gofumpt) (stricter `gofmt`). Run `make fmt`.
- **Linter:** [golangci-lint](https://golangci-lint.run/) v2 with 15 linters enabled. Run `make lint`. See `.golangci.yml` for the full configuration.
- Each CLI command lives in its own file (e.g., `add.go`, `list.go`)
- Pattern: parse flags → call store → format output
- Tests use `testify/assert` and `testify/require`
- Prefer table-driven tests:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {name: "basic case", input: "hello", want: "HELLO"},
        {name: "empty", input: "", want: ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := MyFunction(tt.input)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

- In-memory SQLite for all tests (`db.OpenInMemory()`)

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `ci`, `chore`

**Scopes:** `taskprim`, `stateprim`, `knowledgeprim`, `queueprim`, `primkit`, or omit for cross-cutting changes.

Examples:
- `feat(taskprim): add restore command for point-in-time recovery`
- `fix(knowledgeprim): quote hyphenated terms in FTS5 queries`
- `ci: add govulncheck to CI pipeline`
- `docs: update CONTRIBUTING with conventional commits`

One logical change per commit. Keep commits focused.

## Pull Requests

1. Fork the repo and create a branch from `main`
2. Write tests for new functionality
3. Run `make all` — it must pass with no failures
4. Open a PR with a clear description of what changed and why

## Verifying Releases

Release binaries are signed with [cosign](https://docs.sigstore.dev/cosign/overview/) via keyless signing (Sigstore OIDC). To verify a release:

```bash
# Download the release assets (checksums + signature + certificate)
cosign verify-blob \
  --certificate checksums.txt.pem \
  --signature checksums.txt.sig \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp 'github\.com/propifly/primkit' \
  checksums.txt
```

## Architecture Notes

Before making significant changes, read `docs/architecture.md` to understand the layered design. The key constraint: store interfaces are the boundary — CLI, API, and MCP are sibling consumers that never depend on each other.

## Reporting Issues

Use GitHub Issues. For bugs, include:
- What you expected to happen
- What actually happened
- Steps to reproduce
- Go version and OS
