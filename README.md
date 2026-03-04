# primkit

Agent-native operational primitives. Single Go binaries with embedded SQLite that give AI agents structured, queryable infrastructure for the two things every production agent needs: **task management** and **operational state**.

---

## Why primkit?

Production AI agents need to persist state, track tasks, and coordinate across sessions. Most teams duct-tape this together with markdown files, JSON blobs, or external databases. primkit provides purpose-built primitives that are:

- **Single-binary** — no external dependencies, no Docker, no database server
- **Three interfaces** — CLI, HTTP API, and MCP (Model Context Protocol) out of the box
- **SQLite-native** — embedded WAL-mode database with optional cloud replication
- **Agent-first** — designed for the access patterns agents actually use

## Primitives

### taskprim

Task management for agents and the humans they work with. Tasks have an explicit lifecycle (`open` → `done` | `killed`), belong to lists, carry freeform labels, and support per-agent seen-tracking.

```bash
# Create a task
taskprim add "Deploy v2 to staging" --list ops --label deploy --source johanna

# List open tasks
taskprim list --list ops --state open

# Mark as done
taskprim done t_abc123

# What hasn't agent "johanna" seen yet?
taskprim list --unseen-by johanna
```

### stateprim

Operational state persistence. Three access patterns unified under a single namespaced key-value model:

- **Key-value state** — `set` / `get` / `update` for current state
- **Dedup lookups** — `has` / `set-if-new` for existence checks
- **Append records** — immutable, timestamped log entries

```bash
# Store configuration state
stateprim set config theme '"dark"'

# Check if an email was already sent (dedup)
stateprim has sent-emails msg:abc123

# Append an audit log entry (immutable)
stateprim append audit '{"action":"deploy","version":"1.2.3"}'

# Query recent events
stateprim query audit --since 24h
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   primkit (shared)                   │
│  config · db · auth · server · mcp scaffold          │
└──────────────┬──────────────────────┬────────────────┘
               │                      │
       ┌───────┴───────┐      ┌───────┴───────┐
       │   taskprim    │      │   stateprim   │
       │               │      │               │
       │  CLI · API    │      │  CLI · API    │
       │  MCP · Store  │      │  MCP · Store  │
       └───────────────┘      └───────────────┘
```

Each primitive is a **single Go binary** with three access modes:

| Mode | Use case | Auth |
|------|----------|------|
| **CLI** (default) | Local shell, scripts, cron | Filesystem permissions |
| **serve** | Remote clients, web UIs | Bearer token |
| **mcp** | AI agent integration (Claude, Cursor) | stdio: none / SSE: Bearer token |

## Installation

### From source

Requires [Go 1.22+](https://go.dev/dl/):

```bash
git clone https://github.com/propifly/primkit.git
cd primkit
make build
```

Binaries are built to `./bin/`:

```
bin/taskprim
bin/stateprim
```

### Cross-compile for Raspberry Pi

```bash
make build-pi
# Produces bin/taskprim-linux-arm64 and bin/stateprim-linux-arm64
```

## Quick start

### taskprim

```bash
# Add a task (auto-creates the database at ~/.taskprim/default.db)
taskprim add "Review PR #42" --list andres --label code-review

# List all open tasks
taskprim list

# Full lifecycle
taskprim add "Fix login bug" --list andres --source johanna
taskprim list --list andres
taskprim done t_abc123
taskprim list --state done
```

### stateprim

```bash
# Set a value (auto-creates the database at ~/.stateprim/default.db)
stateprim set config app.theme '"dark"'

# Retrieve it
stateprim get config app.theme

# Dedup check
stateprim set-if-new sent-emails msg:abc123 '{"to":"alice@example.com"}'
stateprim has sent-emails msg:abc123  # → yes

# Append an immutable log entry
stateprim append audit '{"action":"deploy","version":"1.2.3"}'

# Query a namespace
stateprim query audit --since 24h --format json
```

## HTTP API

Both primitives can run as HTTP servers:

```bash
# taskprim on port 8090
taskprim serve --port 8090

# stateprim on port 8091
stateprim serve --port 8091
```

### taskprim endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/tasks` | Create a task |
| `GET` | `/v1/tasks` | List/query tasks |
| `GET` | `/v1/tasks/{id}` | Get a task |
| `PATCH` | `/v1/tasks/{id}` | Edit a task |
| `POST` | `/v1/tasks/{id}/done` | Mark done |
| `POST` | `/v1/tasks/{id}/kill` | Mark killed |
| `POST` | `/v1/seen/{agent}` | Mark tasks as seen |
| `GET` | `/v1/labels` | List labels |
| `GET` | `/v1/lists` | List all lists |
| `GET` | `/v1/stats` | Aggregate stats |

### stateprim endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/records` | Set (upsert) a record |
| `GET` | `/v1/records/{ns}/{key}` | Get a record |
| `DELETE` | `/v1/records/{ns}/{key}` | Delete a record |
| `POST` | `/v1/records/{ns}/set-if-new` | Create if not exists |
| `POST` | `/v1/records/{ns}/append` | Append immutable record |
| `GET` | `/v1/records/{ns}/has/{key}` | Existence check |
| `GET` | `/v1/records/{ns}` | Query namespace |
| `POST` | `/v1/records/{ns}/purge` | Purge old records |
| `GET` | `/v1/namespaces` | List namespaces |
| `GET` | `/v1/stats` | Aggregate stats |
| `GET` | `/v1/export` | Export records |
| `POST` | `/v1/import` | Import records |

### Authentication

When auth keys are configured, all API requests require a Bearer token:

```bash
curl -H "Authorization: Bearer tp_sk_your_key_here" \
  http://localhost:8090/v1/tasks
```

## MCP (Model Context Protocol)

Both primitives can run as MCP servers for direct AI agent integration:

```bash
# stdio transport (local agent on same machine)
taskprim mcp --transport stdio
stateprim mcp --transport stdio

# SSE transport (remote agent over HTTP)
taskprim mcp --transport sse --port 8091
stateprim mcp --transport sse --port 8092
```

### Claude Desktop configuration

Add to your Claude Desktop `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "taskprim": {
      "command": "/path/to/taskprim",
      "args": ["mcp", "--transport", "stdio"]
    },
    "stateprim": {
      "command": "/path/to/stateprim",
      "args": ["mcp", "--transport", "stdio"]
    }
  }
}
```

### Available MCP tools

**taskprim** (11 tools): `taskprim_add`, `taskprim_list`, `taskprim_get`, `taskprim_done`, `taskprim_kill`, `taskprim_edit`, `taskprim_seen`, `taskprim_label_clear`, `taskprim_labels`, `taskprim_lists`, `taskprim_stats`

**stateprim** (10 tools): `stateprim_set`, `stateprim_get`, `stateprim_has`, `stateprim_set_if_new`, `stateprim_append`, `stateprim_delete`, `stateprim_query`, `stateprim_purge`, `stateprim_namespaces`, `stateprim_stats`

## Configuration

Copy `config.example.yaml` to `config.yaml` and edit:

```yaml
storage:
  db: ~/.taskprim/default.db

auth:
  keys:
    - key: "tp_sk_your_api_key_here"
      name: "johanna"

server:
  port: 8090
```

### Environment variable overrides

Every config value can be overridden with environment variables:

| Variable | Overrides |
|----------|-----------|
| `TASKPRIM_DB` | `storage.db` (for taskprim) |
| `STATEPRIM_DB` | `storage.db` (for stateprim) |
| `TASKPRIM_LIST` | Default list for new tasks |

### Global flags

Both binaries accept:

```
--db <path>        Path to SQLite database
--config <path>    Path to config file
--format <fmt>     Output format: table (default), json, quiet
```

## Project structure

```
primkit/
├── primkit/                  # Shared foundation library
│   ├── auth/                 #   API key validation (constant-time)
│   ├── config/               #   YAML + env var config loader
│   ├── db/                   #   SQLite (WAL mode) + migration runner
│   ├── mcp/                  #   MCP server scaffold
│   ├── replicate/             #   Litestream WAL replication wrapper
│   └── server/               #   HTTP server, middleware, JSON helpers
├── taskprim/                 # Task management primitive
│   ├── cmd/taskprim/         #   Binary entrypoint
│   └── internal/
│       ├── model/            #   Task, Filter, state machine
│       ├── store/            #   Store interface + SQLite impl
│       ├── cli/              #   Cobra commands (14 commands)
│       ├── api/              #   HTTP API handler
│       └── mcpserver/        #   MCP tool registrations
├── stateprim/                # State persistence primitive
│   ├── cmd/stateprim/        #   Binary entrypoint
│   └── internal/
│       ├── model/            #   Record, QueryFilter
│       ├── store/            #   Store interface + SQLite impl
│       ├── cli/              #   Cobra commands (14 commands)
│       ├── api/              #   HTTP API handler
│       └── mcpserver/        #   MCP tool registrations
├── go.work                   # Go workspace (3 modules)
├── Makefile                  # build, test, lint, fmt, tidy, build-pi
└── config.example.yaml       # Configuration template
```

## Development

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)

### Build & test

```bash
make build          # Build both binaries
make test           # Run all tests (322 tests across 3 modules)
make lint           # Run go vet on all modules
make fmt            # Format all code
make tidy           # Tidy all go.mod files
make clean          # Remove built binaries
```

### Running tests

```bash
# All modules
make test

# Single module
cd taskprim && go test -v ./...
cd stateprim && go test -v ./...
cd primkit && go test -v ./...
```

Tests use **in-memory SQLite** — no disk I/O, no cleanup, fast and isolated.

## Design decisions

| Decision | Rationale |
|----------|-----------|
| **Pure Go SQLite** (`modernc.org/sqlite`) | No CGo = simpler cross-compilation for ARM (Raspberry Pi) |
| **Embedded migrations** (`embed.FS`) | Single binary, no external SQL files to ship |
| **Go workspace** (`go.work`) | Three modules share code without publishing packages |
| **Interface-based store** | CLI, API, and MCP are sibling consumers of the same contract |
| **In-memory SQLite for tests** | Catches real SQL bugs that mocks would miss |
| **Cobra for CLI** | De facto Go CLI standard, good completions and help |

## Roadmap

- [x] Shared foundation (config, db, auth, server, mcp scaffold)
- [x] taskprim (model, store, CLI, HTTP API, MCP)
- [x] stateprim (model, store, CLI, HTTP API, MCP)
- [x] Litestream replication to object storage (S3, R2, B2, GCS)
- [x] GitHub Actions CI pipeline
- [x] Pre-built binaries (GoReleaser)

## Documentation

- [Configuration Reference](docs/configuration.md) — full YAML spec, env var overrides, examples
- [Architecture](docs/architecture.md) — layered design, store interfaces, data flow, replication
- [Setup Guide](SETUP.md) — R2/S3 setup, replication testing, MCP configuration
- [Contributing](CONTRIBUTING.md) — dev setup, code style, PR process
- [Security Policy](SECURITY.md) — vulnerability reporting
- [Changelog](CHANGELOG.md) — release history

## License

MIT License — see [LICENSE](./LICENSE).

Copyright (c) 2026 Propifly, Inc.
