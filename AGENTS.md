# AGENTS.md

Instructions for AI coding agents working on the primkit codebase, and pointers for agents using primkit as a tool.


## Setup

One-time setup after cloning:

```bash
bash scripts/install-hooks.sh
```

This installs a pre-commit hook that runs `make docs-check` before each commit,
preventing you from committing stale documentation.

## If you are using primkit (not developing it)

See the [Agent Reference](docs/agent-reference.md) for structured command tables, JSON output schemas, decision trees, and error handling patterns for all four primitives (taskprim, stateprim, knowledgeprim, queueprim).

For knowledgeprim usage strategy (entity types, relationships, search modes, discovery workflows), see the [knowledgeprim Guide](docs/knowledgeprim.md).

## Build and test

```bash
make build          # Build bin/taskprim, bin/stateprim, bin/knowledgeprim, bin/queueprim
make test           # Run all tests across all modules, race detector
make lint           # go vet on all modules
make fmt            # gofmt -s -w on all modules
make tidy           # go mod tidy on all modules
```

Single module:

```bash
cd taskprim && go test -v -race -count=1 ./...
cd stateprim && go test -v -race -count=1 ./...
cd knowledgeprim && go test -v -race -count=1 ./...
cd queueprim && go test -v -race -count=1 ./...
cd primkit && go test -v -race -count=1 ./...
```

## Tech stack

- Go 1.26+
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
├── queueprim/         # Work queue primitive
├── go.work            # Go workspace (5 modules)
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

knowledgeprim additionally has `internal/embed/` for the embedding provider abstraction. queueprim's store runs a background sweeper goroutine (in serve/mcp modes) that releases claimed jobs whose visibility timeout has expired.

## Code style

- Standard Go conventions: `gofmt`, `go vet`
- Each CLI command in its own file (`add.go`, `search.go`, `capture.go`)
- Command pattern: parse flags -> call store -> format output
- Tests use `testify/assert` and `testify/require`
- All tests use in-memory SQLite (`db.OpenInMemory()`) — no disk I/O, no cleanup
- No global state — store is injected via `cobra.Command` context


## When adding a new prim

```bash
1. bash scripts/new-prim.sh <name>
2. Implement commands in <name>/internal/cli/
3. make all   # fails until everything is wired correctly
4. Update README.md, llms.txt, REPOS.md manually
```

The scaffold script registers the new prim in go.work, Makefile, scripts/docgen.sh,
and docs/agent-reference.md. `make check-registration` validates all registrations
are in place.

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

## Documentation maintenance

The `docs/agent-reference.md` command tables are auto-generated from each primitive's Cobra command tree. Hand-written sections (schemas, idempotency tables, decision trees) are preserved — only the content between anchor comments is replaced.

### Regenerate docs

```bash
make docs        # Regenerate docs/agent-reference.md
make docs-check  # Check that docs are up to date (used in CI)
```

### How it works

1. Each primitive has a `cmd/docgen/main.go` that walks its Cobra command tree and emits JSON metadata to stdout.
2. `primkit/cmd/docupdater/main.go` reads the JSON files and rewrites the anchored sections in `docs/agent-reference.md`.
3. `scripts/docgen.sh` orchestrates both steps.

### Anchor format

Each prim's Commands table is wrapped in HTML comment anchors:

```
<!-- docgen:start:<primname>:commands -->
| Command | Synopsis | Flags |
...
<!-- docgen:end:<primname>:commands -->
```

Do not edit content inside these anchors manually — it will be overwritten by the next `make docs` run.

### When to run make docs

Run `make docs` after any change to:
- CLI flags (adding, removing, or renaming flags)
- Adding or removing commands
- Changing command `Use` strings or `Short` descriptions
- Adding `MarkFlagRequired` to a command

CI runs `make docs-check` on every pull request and will fail if the docs are out of date.
