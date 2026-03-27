<div align="center">
  <img src="docs/assets/logo.png" width="128" alt="primkit" />
  <h1>primkit</h1>
  <p><strong>Persistent state for AI agents. Four CLI tools, embedded SQLite, zero infrastructure.</strong></p>

  [![CI](https://github.com/propifly/primkit/actions/workflows/ci.yml/badge.svg)](https://github.com/propifly/primkit/actions/workflows/ci.yml)
  [![Release](https://img.shields.io/github/v/release/propifly/primkit)](https://github.com/propifly/primkit/releases/latest)
  [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
  [![Go 1.26+](https://img.shields.io/badge/go-1.26+-00ADD8.svg)](https://go.dev/dl/)
</div>

Each primitive is a standalone binary. No server, no configuration, no runtime dependencies. State survives session ends, terminal closes, and context window limits.

- **taskprim**: task lifecycle, dependencies, frontier queries
- **stateprim**: key-value state, dedup checks, append-only logs
- **knowledgeprim**: knowledge graph with hybrid search (FTS5 + vectors)
- **queueprim**: persistent work queues with priority, retries, and dead-letter

---

**Not sure if you need this?** Paste this into your agent:

> *Read https://github.com/propifly/primkit/blob/main/docs/agent-reference.md, then tell me whether primkit would help with what we've been doing, and which primitives are relevant.*

It will map primkit to your actual workflow and give you a straight answer.

---

```bash
# One command. Database auto-created on first use.
taskprim add "Deploy v2 to staging" --list ops --label deploy
taskprim list --list ops --state open
taskprim done t_abc123
```

<div align="center">

*An agent creates a task, queries open work, and marks it done, all from the shell.*

![primkit demo](docs/assets/demo.gif)

</div>


### Install

```bash
go install github.com/propifly/primkit/taskprim/cmd/taskprim@latest
go install github.com/propifly/primkit/stateprim/cmd/stateprim@latest
go install github.com/propifly/primkit/knowledgeprim/cmd/knowledgeprim@latest
go install github.com/propifly/primkit/queueprim/cmd/queueprim@latest
```

Or download pre-built binaries from [the latest release](https://github.com/propifly/primkit/releases/latest). See [Installation](#installation) for all options.

> **Setting up an agent?** Point it at the [Agent Reference](docs/agent-reference.md) for command tables, JSON schemas, and decision trees for all four primitives.

---

## Table of Contents

- [Primitives](#primitives): taskprim, stateprim, knowledgeprim, queueprim
- [Installation](#installation): pre-built binaries or from source
- [Quick Start](#quick-start): get running in 30 seconds
- [Agent Quick Start](#agent-quick-start): verify the install programmatically
- [HTTP API](#http-api): REST endpoints
- [MCP](#mcp-model-context-protocol): IDE integration for Claude Desktop, Cursor, etc.
- [Configuration](#configuration): YAML, env vars, global flags
- [Development](#development): build, test, lint
- [Design Decisions](#design-decisions): why we built it this way
- [Documentation](#documentation): full reference docs

---

## Primitives

### taskprim

Task management for agents and the humans they work with. Tasks have an explicit lifecycle (`open` → `done` | `killed`), belong to lists, carry freeform labels, and support per-agent seen-tracking. Structural task-to-task dependencies enable dependency graphs with cycle detection and frontier queries ("what can I work on next?").

```bash
# Create a task
taskprim add "Deploy v2 to staging" --list ops --label deploy --source johanna

# List open tasks
taskprim list --list ops --state open

# Mark as done
taskprim done t_abc123

# What hasn't agent "johanna" seen yet?
taskprim list --unseen-by johanna

# Dependencies: B depends on A
taskprim dep add t_taskB t_taskA

# What's ready to work on? (no unresolved dependencies)
taskprim frontier --list ops
```

### stateprim

Operational state persistence. Three access patterns unified under a single namespaced key-value model:

- **Key-value state**: `set` / `get` / `update` for current state
- **Dedup lookups**: `has` / `set-if-new` for existence checks
- **Append records**: immutable, timestamped log entries

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

### knowledgeprim

Knowledge graphs with semantic search and discovery. Typed entities, weighted edges with context, and hybrid retrieval (FTS5 + vector + RRF).

```bash
# Capture a knowledge entity (auto-creates ~/.knowledgeprim/default.db)
knowledgeprim capture --type article --title "Agents prefer CLI over MCP" \
  --body "Benchmarks show CLI tools are cheaper and agents perform better..." \
  --source thomas

# Search (hybrid: keyword + semantic)
knowledgeprim search "agent tool preferences" --mode hybrid

# Connect two entities with context
knowledgeprim connect e_abc e_def \
  --relationship extends --context "Builds on the CLI-first argument with cost data"

# Traverse the graph
knowledgeprim related e_abc --depth 2 --direction both

# Discover patterns (orphans, clusters, bridges)
knowledgeprim discover --clusters --bridges
```

**[knowledgeprim Guide](docs/knowledgeprim.md)**: entity types, relationship design, edge context patterns, search strategy, discovery workflows, and agent playbooks.

### queueprim

Persistent work queues for multi-agent pipelines. Jobs have priority, retries, and a dead-letter path. Atomic dequeue prevents double-processing across concurrent workers.

```bash
# Enqueue a job (auto-creates ~/.queueprim/default.db)
queueprim enqueue infra/fixes '{"host":"web-01","issue":"disk_full"}'

# Worker atomically claims the next job
queueprim dequeue infra/fixes --worker johanna

# Mark it done with output
queueprim complete q_abc123 --output '{"freed_gb":12}'

# Inspect without claiming
queueprim peek infra/fixes

# List all queues with job counts
queueprim queues
```

Job lifecycle: `pending` → `claimed` → `done` / `failed` / `dead`

**[queueprim Guide](docs/queueprim.md)**: visibility timeout, the worker loop, priority and queue design, retry and dead-letter strategy, monitoring.

## Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          primkit (shared)                                │
│  config · db · auth · server · mcp scaffold · replicate                  │
└──────┬───────────────────┬───────────────────┬───────────────┬───────────┘
       │                   │                   │               │
┌──────┴──────┐   ┌────────┴──────┐   ┌────────┴───────┐  ┌───┴──────────┐
│  taskprim   │   │  stateprim    │   │ knowledgeprim  │  │  queueprim   │
│             │   │               │   │                │  │              │
│ CLI · API   │   │  CLI · API    │   │  CLI · API     │  │ CLI · API    │
│ MCP · Store │   │  MCP · Store  │   │  MCP · Store   │  │ MCP · Store  │
└─────────────┘   └───────────────┘   │  Embed · Search│  │ Sweeper      │
                                      └────────────────┘  └──────────────┘
```

Each primitive is a **single Go binary** with three access modes:

| Mode | Use case | Auth |
|------|----------|------|
| **CLI** (default) | Local shell, scripts, cron | Filesystem permissions |
| **serve** | Remote clients, web UIs | Bearer token |
| **mcp** | AI agent integration (Claude, Cursor) | stdio: none / SSE: Bearer token |

## Installation

### Pre-built binaries (recommended)

Download the latest release for your platform. Install only the primitives you need:

```bash
# Set the version (check https://github.com/propifly/primkit/releases for latest)
VERSION="0.5.1"

# macOS (Apple Silicon)
for bin in taskprim stateprim knowledgeprim queueprim; do
  curl -sL "https://github.com/propifly/primkit/releases/download/v${VERSION}/${bin}_${VERSION}_darwin_arm64.tar.gz" | tar xz
done

# macOS (Intel)
for bin in taskprim stateprim knowledgeprim queueprim; do
  curl -sL "https://github.com/propifly/primkit/releases/download/v${VERSION}/${bin}_${VERSION}_darwin_amd64.tar.gz" | tar xz
done

# Linux (x86_64)
for bin in taskprim stateprim knowledgeprim queueprim; do
  curl -sL "https://github.com/propifly/primkit/releases/download/v${VERSION}/${bin}_${VERSION}_linux_amd64.tar.gz" | tar xz
done

# Linux (ARM64 / Raspberry Pi)
for bin in taskprim stateprim knowledgeprim queueprim; do
  curl -sL "https://github.com/propifly/primkit/releases/download/v${VERSION}/${bin}_${VERSION}_linux_arm64.tar.gz" | tar xz
done
```

Move to your PATH:

```bash
sudo mv taskprim stateprim knowledgeprim queueprim /usr/local/bin/
```

Or use `gh`:

```bash
gh release download --latest --repo propifly/primkit --pattern '*darwin_arm64*'
```

### go install

Requires [Go 1.26+](https://go.dev/dl/):

```bash
go install github.com/propifly/primkit/taskprim/cmd/taskprim@latest
go install github.com/propifly/primkit/stateprim/cmd/stateprim@latest
go install github.com/propifly/primkit/knowledgeprim/cmd/knowledgeprim@latest
go install github.com/propifly/primkit/queueprim/cmd/queueprim@latest
```

### From source

```bash
git clone https://github.com/propifly/primkit.git
cd primkit
make build
# Binaries: bin/taskprim, bin/stateprim, bin/knowledgeprim, bin/queueprim
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

### knowledgeprim

```bash
# Capture a knowledge entity (auto-creates ~/.knowledgeprim/default.db)
knowledgeprim capture --type concept --title "Eventual consistency" \
  --body "A consistency model where replicas converge over time" --source agent

# Search for it
knowledgeprim search "consistency models"

# Connect two entities
knowledgeprim connect e_abc e_def --relationship relates_to \
  --context "Both deal with distributed state"

# Traverse from an entity
knowledgeprim related e_abc --depth 2
```

### queueprim

```bash
# Enqueue a job (auto-creates ~/.queueprim/default.db)
queueprim enqueue tasks/default '{"action":"reindex"}'

# Worker claims and processes it
queueprim dequeue tasks/default --worker agent-01

# Complete or fail it
queueprim complete q_abc123
queueprim fail q_abc123 --reason "upstream timeout"
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

For knowledgeprim:

```bash
knowledgeprim capture --type test --title "Hello world" --source agent
knowledgeprim search "hello"
```

For queueprim:

```bash
queueprim enqueue test '{"ping":true}'
queueprim dequeue test --worker agent
```

No config file needed. The database is created automatically on first use.

## HTTP API

All primitives can run as HTTP servers:

```bash
# taskprim on port 8090
taskprim serve --port 8090

# stateprim on port 8091
stateprim serve --port 8091

# knowledgeprim on port 8092
knowledgeprim serve --port 8092

# queueprim on port 8093
queueprim serve --port 8093
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
| `POST` | `/v1/tasks/{id}/deps` | Add a dependency |
| `DELETE` | `/v1/tasks/{id}/deps/{dep-id}` | Remove a dependency |
| `GET` | `/v1/tasks/{id}/deps` | List dependencies |
| `GET` | `/v1/tasks/{id}/dependents` | List reverse dependencies |
| `GET` | `/v1/frontier` | Tasks ready for execution |
| `GET` | `/v1/dep-edges` | Raw dependency edges |
| `POST` | `/v1/labels/{name}/clear` | Remove label from all tasks |
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

<details>
<summary><strong>knowledgeprim endpoints</strong></summary>

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/entities` | Capture an entity |
| `GET` | `/v1/entities/{id}` | Get an entity with edges |
| `PATCH` | `/v1/entities/{id}` | Update entity fields |
| `DELETE` | `/v1/entities/{id}` | Delete entity and edges |
| `GET` | `/v1/search` | Search (hybrid, FTS, or vector) |
| `GET` | `/v1/entities/{id}/related` | Graph traversal |
| `POST` | `/v1/edges` | Create an edge |
| `PATCH` | `/v1/edges/{source}/{target}/{rel}` | Update edge |
| `POST` | `/v1/edges/{source}/{target}/{rel}/strengthen` | Increment weight |
| `DELETE` | `/v1/edges/{source}/{target}/{rel}` | Delete edge |
| `GET` | `/v1/discover` | Discovery (orphans, clusters, bridges) |
| `GET` | `/v1/types` | List entity types |
| `GET` | `/v1/relationships` | List relationship types |
| `GET` | `/v1/stats` | Aggregate stats |

</details>

<details>
<summary><strong>queueprim endpoints</strong></summary>

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/jobs` | Enqueue a job |
| `GET` | `/v1/jobs` | List jobs (filter by queue, status, type) |
| `GET` | `/v1/jobs/{id}` | Get a job |
| `POST` | `/v1/queues/{queue}/dequeue` | Atomically claim the next job |
| `POST` | `/v1/jobs/{id}/complete` | Mark a job done |
| `POST` | `/v1/jobs/{id}/fail` | Mark a job failed (retries or dead-letter) |
| `POST` | `/v1/jobs/{id}/release` | Return a claimed job to pending |
| `POST` | `/v1/jobs/{id}/extend` | Extend a claimed job's visibility timeout |
| `GET` | `/v1/queues` | List all queues with job counts |
| `GET` | `/v1/stats` | Aggregate stats |
| `DELETE` | `/v1/queues/{queue}` | Purge jobs by status and age |

Queue names may contain slashes (e.g., `infra/prod`). `dequeue` returns `204 No Content` when the queue is empty.

</details>

### Authentication

When auth keys are configured, all API requests require a Bearer token:

```bash
curl -H "Authorization: Bearer tp_sk_your_key_here" \
  http://localhost:8090/v1/tasks
```

## MCP (Model Context Protocol)

> **When to use MCP vs CLI:** If your agent has shell access (Claude Code, Codex, terminal-based agents), the CLI is the simplest and most reliable option. Agents are trained on CLI tools and can pipe, chain, and compose them naturally. MCP is ideal for IDE integrations (Claude Desktop, Cursor, VS Code) where there's no terminal access.

All primitives can run as MCP servers for direct AI agent integration:

```bash
# stdio transport (local agent on same machine)
taskprim mcp --transport stdio
stateprim mcp --transport stdio
knowledgeprim mcp --transport stdio
queueprim mcp --transport stdio

# SSE transport (remote agent over HTTP)
taskprim mcp --transport sse --port 8091
stateprim mcp --transport sse --port 8092
knowledgeprim mcp --transport sse --port 8093
queueprim mcp --transport sse --port 8094
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
    },
    "knowledgeprim": {
      "command": "knowledgeprim",
      "args": ["mcp", "--transport", "stdio"]
    },
    "queueprim": {
      "command": "queueprim",
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
    },
    "knowledgeprim": {
      "command": "/usr/local/bin/knowledgeprim",
      "args": ["mcp", "--transport", "stdio"]
    },
    "queueprim": {
      "command": "/usr/local/bin/queueprim",
      "args": ["mcp", "--transport", "stdio"]
    }
  }
}
```

<details>
<summary><strong>Available MCP tools</strong></summary>

**taskprim** (16 tools): `taskprim_add`, `taskprim_list`, `taskprim_get`, `taskprim_done`, `taskprim_kill`, `taskprim_edit`, `taskprim_seen`, `taskprim_dep_add`, `taskprim_dep_remove`, `taskprim_deps`, `taskprim_dependents`, `taskprim_frontier`, `taskprim_label_clear`, `taskprim_labels`, `taskprim_lists`, `taskprim_stats`

**stateprim** (10 tools): `stateprim_set`, `stateprim_get`, `stateprim_has`, `stateprim_set_if_new`, `stateprim_append`, `stateprim_delete`, `stateprim_query`, `stateprim_purge`, `stateprim_namespaces`, `stateprim_stats`

**knowledgeprim** (14 tools): `knowledgeprim_capture`, `knowledgeprim_search`, `knowledgeprim_get`, `knowledgeprim_related`, `knowledgeprim_connect`, `knowledgeprim_strengthen`, `knowledgeprim_edge_edit`, `knowledgeprim_disconnect`, `knowledgeprim_edit`, `knowledgeprim_delete`, `knowledgeprim_discover`, `knowledgeprim_types`, `knowledgeprim_relationships`, `knowledgeprim_stats`

**queueprim** (11 tools): `queueprim_enqueue`, `queueprim_dequeue`, `queueprim_complete`, `queueprim_fail`, `queueprim_release`, `queueprim_extend`, `queueprim_peek`, `queueprim_list`, `queueprim_get`, `queueprim_queues`, `queueprim_stats`

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
| `KNOWLEDGEPRIM_DB` | `storage.db` (for knowledgeprim) |
| `QUEUEPRIM_DB` | `storage.db` (for queueprim) |
| `TASKPRIM_LIST` | Default list for new tasks |

### Global flags

All binaries accept:

```
--db <path>        Path to SQLite database
--config <path>    Path to config file
--format <fmt>     Output format (knowledgeprim: text/json, others: table/json/quiet)
```

<details>
<summary><strong>Project structure</strong></summary>

```
primkit/
├── primkit/                  # Shared foundation library
│   ├── auth/                 #   API key validation (constant-time)
│   ├── cmd/docupdater/       #   Auto-generated docs updater tool
│   ├── config/               #   YAML + env var config loader
│   ├── db/                   #   SQLite (WAL mode) + migration runner
│   ├── docgen/               #   Documentation generation library
│   ├── mcp/                  #   MCP server scaffold
│   ├── replicate/            #   Litestream WAL replication wrapper
│   └── server/               #   HTTP server, middleware, JSON helpers
├── taskprim/                 # Task management primitive
│   ├── cmd/taskprim/         #   Binary entrypoint
│   └── internal/
│       ├── model/            #   Task, Filter, state machine
│       ├── store/            #   Store interface + SQLite impl
│       ├── cli/              #   Cobra commands (21 commands)
│       ├── api/              #   HTTP API handler
│       └── mcpserver/        #   MCP tool registrations
├── stateprim/                # State persistence primitive
│   ├── cmd/stateprim/        #   Binary entrypoint
│   └── internal/
│       ├── model/            #   Record, QueryFilter
│       ├── store/            #   Store interface + SQLite impl
│       ├── cli/              #   Cobra commands (16 commands)
│       ├── api/              #   HTTP API handler
│       └── mcpserver/        #   MCP tool registrations
├── knowledgeprim/            # Knowledge graph primitive
│   ├── cmd/knowledgeprim/    #   Binary entrypoint
│   └── internal/
│       ├── model/            #   Entity, Edge, SearchFilter, TraversalOpts
│       ├── store/            #   Store interface + SQLite impl (FTS5, vectors)
│       ├── embed/            #   Embedding provider (Gemini, OpenAI, custom)
│       ├── cli/              #   Cobra commands (22 commands)
│       ├── api/              #   HTTP API handler
│       └── mcpserver/        #   MCP tool registrations
├── queueprim/                # Work queue primitive
│   ├── cmd/queueprim/        #   Binary entrypoint
│   └── internal/
│       ├── model/            #   Job, Filter, Priority, Status, QueueInfo, Stats
│       ├── store/            #   Store interface + SQLite impl
│       ├── cli/              #   Cobra commands (18 commands)
│       ├── api/              #   HTTP API handler
│       └── mcpserver/        #   MCP tool registrations
├── go.work                   # Go workspace (5 modules)
├── Makefile                  # build, test, lint, fmt, tidy, build-pi
└── config.example.yaml       # Configuration template
```

</details>

## Development

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)

### Build & test

```bash
make build          # Build all four binaries
make test           # Run all tests across all modules
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
cd knowledgeprim && go test -v ./...
cd queueprim && go test -v ./...
cd primkit && go test -v ./...
```

Tests use **in-memory SQLite**: no disk I/O, no cleanup, fast and isolated.

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Pure Go SQLite** (`modernc.org/sqlite`) | No CGo = simpler cross-compilation for ARM (Raspberry Pi) |
| **Embedded migrations** (`embed.FS`) | Single binary, no external SQL files to ship |
| **Go workspace** (`go.work`) | Four modules share code without publishing packages |
| **Interface-based store** | CLI, API, and MCP are sibling consumers of the same contract |
| **In-memory SQLite for tests** | Catches real SQL bugs that mocks would miss |
| **Cobra for CLI** | De facto Go CLI standard, good completions and help |
| **CLI-first design** | Agents with shell access prefer CLI over MCP; trained on man pages, composable with pipes |
| **FTS5 + vectors in SQLite** | No external search engine needed; hybrid retrieval in a single file |
| **Optional embedding** | knowledgeprim works without vectors (FTS5 search, manual edges, discovery all work) |
| **Contextualized edges** | Edges store *why* things connect, not just *that* they connect |

### CLI-first, not MCP-first

primkit ships an MCP interface for IDE-first workflows (Cursor, Claude Desktop), but the product is CLI-native. An agent with shell access uses it like any Unix tool. No IDE required, no host process, works anywhere bash works.

### Why not files?

Files break under two conditions: two agents writing simultaneously (corruption), and structured queries (`--status=in-progress`). primkit handles both. If your agent does single-session, single-step tasks, files are fine. The transition point is the third time you paste state back into a new session by hand.

## Documentation

- [Agent Reference](docs/agent-reference.md): structured command tables, JSON schemas, decision trees, error patterns (for agents)
- [knowledgeprim Guide](docs/knowledgeprim.md): entity types, relationships, edge context, search strategy, discovery, agent workflows
- [queueprim Guide](docs/queueprim.md): visibility timeout, worker loop, priority and queue design, retry and dead-letter strategy
- [Configuration Reference](docs/configuration.md): full YAML spec, env var overrides, examples
- [Architecture](docs/architecture.md): layered design, store interfaces, data flow, replication
- [Building Curious Agents](docs/curiosity-architecture.md): using knowledgeprim's knowledge graph and stateprim to build agents that investigate and learn autonomously across sessions
- [Setup Guide](SETUP.md): R2/S3 setup, replication testing, MCP configuration
- [Contributing](CONTRIBUTING.md): dev setup, code style, PR process
- [Security Policy](SECURITY.md): vulnerability reporting
- [Changelog](CHANGELOG.md): release history

## License

MIT License. See [LICENSE](./LICENSE).

Copyright (c) 2026 Propifly, Inc.

---

If primkit is useful to you, consider giving it a ⭐. It helps others discover it.
