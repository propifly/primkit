# primkit

[![CI](https://github.com/propifly/primkit/actions/workflows/ci.yml/badge.svg)](https://github.com/propifly/primkit/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/propifly/primkit)](https://github.com/propifly/primkit/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/propifly/primkit/primkit)](https://goreportcard.com/report/github.com/propifly/primkit/primkit)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)

**Task management and state persistence for AI agents.** Single Go binaries with embedded SQLite — no dependencies, no Docker, no database server.

- **CLI-first** — agents with shell access (Claude Code, Codex, Cursor) use the CLI directly. No MCP required.
- **Three interfaces** — CLI, HTTP API, and MCP (Model Context Protocol) from the same binary
- **Zero config** — database auto-creates on first use. Install the binary, start using it.
- **SQLite-native** — embedded WAL-mode database with optional cloud replication (S3, R2, B2, GCS)

---

## Table of Contents

- [Primitives](#primitives) — taskprim + stateprim
- [Installation](#installation) — pre-built binaries or from source
- [Quick Start](#quick-start) — get running in 30 seconds
- [Agent Quick Start](#agent-quick-start) — verify the install programmatically
- [HTTP API](#http-api) — REST endpoints
- [MCP](#mcp-model-context-protocol) — IDE integration for Claude Desktop, Cursor, etc.
- [Configuration](#configuration) — YAML, env vars, global flags
- [Development](#development) — build, test, lint
- [Design Decisions](#design-decisions) — why we built it this way
- [Documentation](#documentation) — full reference docs

---

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

### Pre-built binaries (recommended)

Download the latest release for your platform:

```bash
# macOS (Apple Silicon)
curl -sL https://github.com/propifly/primkit/releases/latest/download/taskprim_0.1.0_darwin_arm64.tar.gz | tar xz
curl -sL https://github.com/propifly/primkit/releases/latest/download/stateprim_0.1.0_darwin_arm64.tar.gz | tar xz

# macOS (Intel)
curl -sL https://github.com/propifly/primkit/releases/latest/download/taskprim_0.1.0_darwin_amd64.tar.gz | tar xz
curl -sL https://github.com/propifly/primkit/releases/latest/download/stateprim_0.1.0_darwin_amd64.tar.gz | tar xz

# Linux (x86_64)
curl -sL https://github.com/propifly/primkit/releases/latest/download/taskprim_0.1.0_linux_amd64.tar.gz | tar xz
curl -sL https://github.com/propifly/primkit/releases/latest/download/stateprim_0.1.0_linux_amd64.tar.gz | tar xz

# Linux (ARM64 / Raspberry Pi)
curl -sL https://github.com/propifly/primkit/releases/latest/download/taskprim_0.1.0_linux_arm64.tar.gz | tar xz
curl -sL https://github.com/propifly/primkit/releases/latest/download/stateprim_0.1.0_linux_arm64.tar.gz | tar xz
```

Move to your PATH:

```bash
sudo mv taskprim stateprim /usr/local/bin/
```

Or use `gh`:

```bash
gh release download v0.1.0 --repo propifly/primkit --pattern '*darwin_arm64*'
```

### From source

Requires [Go 1.22+](https://go.dev/dl/):

```bash
git clone https://github.com/propifly/primkit.git
cd primkit
make build
# Binaries: bin/taskprim, bin/stateprim
```

## Quick Start

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

## Agent Quick Start

Three commands to verify the install:

```bash
# 1. Create a task (auto-creates ~/.taskprim/default.db)
taskprim add "test task" --list default --source agent

# 2. List it back
taskprim list --format json

# 3. Mark it done (use the ID from step 2)
taskprim done t_<id>
```

For stateprim:

```bash
stateprim set test hello '"world"'
stateprim get test hello
```

No config file needed. The database is created automatically on first use.

## HTTP API

Both primitives can run as HTTP servers:

```bash
# taskprim on port 8090
taskprim serve --port 8090

# stateprim on port 8091
stateprim serve --port 8091
```

<details>
<summary><strong>taskprim endpoints</strong></summary>

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

</details>

<details>
<summary><strong>stateprim endpoints</strong></summary>

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

</details>

### Authentication

When auth keys are configured, all API requests require a Bearer token:

```bash
curl -H "Authorization: Bearer tp_sk_your_key_here" \
  http://localhost:8090/v1/tasks
```

## MCP (Model Context Protocol)

> **When to use MCP vs CLI:** If your agent has shell access (Claude Code, Codex, terminal-based agents), the CLI is the simplest and most reliable option — agents are trained on CLI tools and can pipe, chain, and compose them naturally. MCP is ideal for IDE integrations (Claude Desktop, Cursor, VS Code) where there's no terminal access.

Both primitives can run as MCP servers for direct AI agent integration:

```bash
# stdio transport (local agent on same machine)
taskprim mcp --transport stdio
stateprim mcp --transport stdio

# SSE transport (remote agent over HTTP)
taskprim mcp --transport sse --port 8091
stateprim mcp --transport sse --port 8092
```

### Claude Code configuration

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "taskprim": {
      "command": "taskprim",
      "args": ["mcp", "--transport", "stdio"]
    },
    "stateprim": {
      "command": "stateprim",
      "args": ["mcp", "--transport", "stdio"]
    }
  }
}
```

If the binaries aren't on your PATH, use the full path (e.g., `/usr/local/bin/taskprim`).

### Claude Desktop configuration

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "taskprim": {
      "command": "/usr/local/bin/taskprim",
      "args": ["mcp", "--transport", "stdio"]
    },
    "stateprim": {
      "command": "/usr/local/bin/stateprim",
      "args": ["mcp", "--transport", "stdio"]
    }
  }
}
```

<details>
<summary><strong>Available MCP tools</strong></summary>

**taskprim** (11 tools): `taskprim_add`, `taskprim_list`, `taskprim_get`, `taskprim_done`, `taskprim_kill`, `taskprim_edit`, `taskprim_seen`, `taskprim_label_clear`, `taskprim_labels`, `taskprim_lists`, `taskprim_stats`

**stateprim** (10 tools): `stateprim_set`, `stateprim_get`, `stateprim_has`, `stateprim_set_if_new`, `stateprim_append`, `stateprim_delete`, `stateprim_query`, `stateprim_purge`, `stateprim_namespaces`, `stateprim_stats`

</details>

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

<details>
<summary><strong>Project structure</strong></summary>

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

</details>

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

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Pure Go SQLite** (`modernc.org/sqlite`) | No CGo = simpler cross-compilation for ARM (Raspberry Pi) |
| **Embedded migrations** (`embed.FS`) | Single binary, no external SQL files to ship |
| **Go workspace** (`go.work`) | Three modules share code without publishing packages |
| **Interface-based store** | CLI, API, and MCP are sibling consumers of the same contract |
| **In-memory SQLite for tests** | Catches real SQL bugs that mocks would miss |
| **Cobra for CLI** | De facto Go CLI standard, good completions and help |
| **CLI-first design** | Agents with shell access prefer CLI over MCP — trained on man pages, composable with pipes |

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

---

If primkit is useful to you, consider giving it a ⭐ — it helps others discover it.
