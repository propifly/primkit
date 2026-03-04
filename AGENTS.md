# AGENTS.md

Instructions for AI coding agents working on the primkit codebase, and pointers for agents using primkit as a tool.

## If you are using primkit (not developing it)

See the [Agent Reference](docs/agent-reference.md) for structured command tables, JSON output schemas, decision trees, and error handling patterns for all three primitives (taskprim, stateprim, knowledgeprim).

For knowledgeprim usage strategy (entity types, relationships, search modes, discovery workflows), see the [knowledgeprim Guide](docs/knowledgeprim.md).

## Build and test

```bash
make build          # Build bin/taskprim, bin/stateprim, bin/knowledgeprim
make test           # Run all tests (348 tests, 4 modules, race detector)
make lint           # go vet on all modules
make fmt            # gofmt -s -w on all modules
make tidy           # go mod tidy on all modules
```

Single module:

```bash
cd taskprim && go test -v -race -count=1 ./...
cd stateprim && go test -v -race -count=1 ./...
cd knowledgeprim && go test -v -race -count=1 ./...
cd primkit && go test -v -race -count=1 ./...
```

## Tech stack

- Go 1.22+
- Pure Go SQLite (`modernc.org/sqlite`) — no CGo
- CLI: `github.com/spf13/cobra`
- Config: `gopkg.in/yaml.v3` + env var overrides
- IDs: `github.com/matoous/go-nanoid/v2`
- Tests: `github.com/stretchr/testify`
- MCP: `github.com/mark3labs/mcp-go`

## Project structure

```
primkit/
├── primkit/           # Shared library (config, db, auth, server, mcp, replicate)
├── taskprim/          # Task management primitive
├── stateprim/         # State persistence primitive
├── knowledgeprim/     # Knowledge graph primitive
├── go.work            # Go workspace (4 modules)
└── Makefile           # Build targets
```

Each primitive follows identical internal layout:

```
<primitive>/
├── cmd/<primitive>/   # main.go entrypoint
└── internal/
    ├── model/         # Domain structs, validation, state machine
    ├── store/         # Store interface + SQLite implementation
    ├── cli/           # Cobra commands (one file per command)
    ├── api/           # HTTP API handler
    └── mcpserver/     # MCP tool registrations
```

knowledgeprim additionally has `internal/embed/` for the embedding provider abstraction.

## Code style

- Standard Go conventions: `gofmt`, `go vet`
- Each CLI command in its own file (`add.go`, `search.go`, `capture.go`)
- Command pattern: parse flags -> call store -> format output
- Tests use `testify/assert` and `testify/require`
- All tests use in-memory SQLite (`db.OpenInMemory()`) — no disk I/O, no cleanup
- No global state — store is injected via `cobra.Command` context

## Git workflow

- Branch from `main`
- Clear, imperative commit messages: `add restore command`, `fix duplicate encoding`
- One logical change per commit
- `make test` must pass before committing
- `make build` must compile cleanly

## Boundaries

### Always do

- Run `make test` after any code change
- Add tests for new functionality (use in-memory SQLite)
- Follow the existing command pattern (parse flags -> store -> format)
- Keep store interfaces as the boundary — CLI, API, MCP never depend on each other

### Ask first

- Adding new dependencies to go.mod
- Changing the store interface (affects CLI, API, and MCP consumers)
- Modifying database migrations (affects existing user data)
- Changing CLI flag names or defaults (breaks existing scripts)

### Never do

- Modify files outside the primitive you're working on without explicit instruction
- Skip tests or use `--short` flag
- Add CGo dependencies (breaks cross-compilation)
- Commit database files, config.yaml, or .env files
- Force push to main
