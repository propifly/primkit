# Architecture

Primkit is a monorepo containing two primitives (**taskprim** and **stateprim**) and a shared infrastructure library (**primkit**). Both primitives follow identical layered architecture — the only differences are the domain model and store operations.

## Repository Structure

```
primkit/
├── go.work                  # Go workspace (3 modules)
├── Makefile                 # build, test, lint, fmt, build-pi
├── config.example.yaml      # shared config format
├── primkit/                 # shared library (module: github.com/propifly/primkit/primkit)
│   ├── config/              # YAML config loader + env var interpolation
│   ├── db/                  # SQLite open + migration runner
│   ├── server/              # HTTP server, middleware, JSON helpers
│   ├── auth/                # Bearer token authentication
│   ├── mcp/                 # MCP server scaffold
│   └── replicate/           # Litestream WAL replication wrapper
├── taskprim/                # task management primitive (module: github.com/propifly/primkit/taskprim)
│   ├── cmd/taskprim/        # main.go entry point
│   └── internal/
│       ├── model/           # Task, Filter, state machine
│       ├── store/           # Store interface + SQLite implementation
│       ├── cli/             # Cobra commands (add, list, done, kill, ...)
│       ├── api/             # HTTP API handler
│       └── mcpserver/       # MCP tool registration
└── stateprim/               # state persistence primitive (module: github.com/propifly/primkit/stateprim)
    ├── cmd/stateprim/       # main.go entry point
    └── internal/
        ├── model/           # Record, QueryFilter
        ├── store/           # Store interface + SQLite implementation
        ├── cli/             # Cobra commands (set, get, append, query, ...)
        ├── api/             # HTTP API handler
        └── mcpserver/       # MCP tool registration
```

## Layered Design

Dependencies flow strictly downward. No lateral dependencies between sibling layers.

```
┌─────────────────────────────────────────────────┐
│                Access Interfaces                 │
│  ┌───────┐    ┌──────────┐    ┌───────────────┐ │
│  │  CLI  │    │ HTTP API │    │  MCP Server   │ │
│  │(cobra)│    │ (net/http)│   │  (mcp-go)     │ │
│  └───┬───┘    └────┬─────┘    └──────┬────────┘ │
│      │             │                 │           │
│      └─────────────┼─────────────────┘           │
│                    │                             │
│              ┌─────▼─────┐                       │
│              │   Store    │ ◄── interface         │
│              │ (interface)│                       │
│              └─────┬─────┘                       │
│                    │                             │
│              ┌─────▼─────┐                       │
│              │   Model   │ ◄── structs,          │
│              │           │     validation,       │
│              └─────┬─────┘     state machine     │
│                    │                             │
│      ┌─────────────┼───────────────┐             │
│      │             │               │             │
│  ┌───▼────┐  ┌─────▼─────┐  ┌─────▼─────┐      │
│  │ config │  │     db     │  │ replicate │      │
│  │ (YAML) │  │  (SQLite)  │  │(Litestream)│     │
│  └────────┘  └───────────┘  └───────────┘      │
│                                                 │
│              primkit (shared library)            │
└─────────────────────────────────────────────────┘
```

### Key Constraint

**CLI, API, and MCP are sibling consumers** of the Store interface. They never depend on each other. This means:

- You can use CLI without the HTTP server
- You can use the API without MCP
- Any new access interface just imports the Store

## Store Interface

The Store is the central abstraction. Each primitive defines its own interface in `internal/store/store.go`.

### taskprim Store (15 operations)

| Operation | Description |
|-----------|-------------|
| `CreateTask` | Persist a new task (store assigns ID, state, timestamps) |
| `GetTask` | Retrieve a single task by ID |
| `ListTasks` | Filter and list tasks (by list, state, labels, source, etc.) |
| `UpdateTask` | Partial update to mutable fields |
| `DoneTask` | Mark task as done (sets `resolved_at`) |
| `KillTask` | Mark task as killed with reason |
| `MarkSeen` | Record that an agent has seen a task |
| `MarkAllSeen` | Mark all open tasks in a list as seen by an agent |
| `ListLabels` | All labels with count of open tasks per label |
| `ClearLabel` | Remove a label from all tasks |
| `ListLists` | All lists with task counts by state |
| `Stats` | Aggregate counts (open, done, killed) |
| `ExportTasks` | Full export for data portability |
| `ImportTasks` | Bulk import preserving IDs |
| `Close` | Release database connection |

### stateprim Store (13 operations)

| Operation | Description |
|-----------|-------------|
| `Set` | Create or update a record (upsert) |
| `Get` | Retrieve by namespace + key |
| `Has` | Check existence |
| `SetIfNew` | Create only if key doesn't exist |
| `Append` | Create immutable record with auto-generated key |
| `Delete` | Remove by namespace + key |
| `Query` | Records matching filter (namespace, key prefix, time window) |
| `Purge` | Delete records older than a duration |
| `ListNamespaces` | All namespaces with record counts |
| `Stats` | Aggregate counts |
| `ExportRecords` | Full export, optionally filtered |
| `ImportRecords` | Bulk import preserving keys |
| `Close` | Release database connection |

## Domain Models

### taskprim: Task

```
Task {
    ID             string      // t_<nanoid>, assigned by store
    List           string      // required: which list
    What           string      // required: task description
    Source         string      // required: who created it
    State          State       // open → done | killed
    WaitingOn      *string     // optional: blocking dependency
    ParentID       *string     // optional: subtask relationship
    Context        *string     // optional: background info
    Labels         []string    // freeform tags
    Created        time.Time   // assigned by store
    Updated        time.Time   // assigned by store
    ResolvedAt     *time.Time  // set on done/kill
    ResolvedReason *string     // why it was killed
}
```

**State machine:**

```
          done()
  open ──────────► done
    │
    │  kill(reason)
    └──────────► killed
```

Tasks start as `open`. Transitions to `done` or `killed` are one-way. There is no restore/reopen.

### stateprim: Record

```
Record {
    Namespace  string           // required: scope
    Key        string           // required: identifier
    Value      json.RawMessage  // required: JSON payload
    Immutable  bool             // true for append records
    CreatedAt  time.Time        // assigned by store
    UpdatedAt  time.Time        // assigned by store
}
```

**Three access patterns share the same model:**

1. **Key-value state** (`set`/`get`): current state, updatable. `Immutable=false`.
2. **Dedup lookups** (`has`/`set-if-new`): existence checks, create-once semantics.
3. **Append log** (`append`): immutable, timestamped entries. `Immutable=true`, auto-generated key.

## Data Flow

### CLI Command

```
User → cobra command → parse flags → store.Operation() → format output → stdout
```

### HTTP API Request

```
Client → HTTP request
  → RequestID middleware (assigns/propagates X-Request-ID)
  → Logging middleware (logs method, path, status, duration)
  → Recovery middleware (catches panics → 500)
  → Auth middleware (validates Bearer token → 401 if invalid)
  → API handler → store.Operation()
  → JSON response
```

### MCP Tool Call

```
Agent → MCP protocol (stdio or SSE)
  → mcp-go framework → tool handler → store.Operation()
  → MCP response
```

## Replication

Litestream is embedded as a Go library (not a sidecar process). WAL frames are continuously streamed to S3-compatible object storage (R2, S3, B2, GCS).

### Lifecycle

Replication runs for **every command** — CLI, serve, and MCP alike. This is managed in `root.go` via Cobra's `PersistentPreRunE` and `PersistentPostRunE`:

```
PersistentPreRunE:
  1. Resolve database path (flag → env var → home default)
  2. Load config (YAML + env overrides)
  3. RestoreIfNeeded: if replication enabled and local DB missing,
     download from replica before opening
  4. Open SQLite database
  5. Start Litestream replication (continuous WAL streaming)

Command executes (add, list, serve, mcp, etc.)

PersistentPostRunE:
  6. Stop replication with final sync
```

For short-lived CLI commands, the final sync ensures the last WAL changes reach the replica. For long-running serve/MCP commands, replication streams continuously.

### Restore

Two restore paths:

- **Auto-restore** (`RestoreIfNeeded`): On startup, if the local DB file doesn't exist but replication is configured, the DB is automatically downloaded from the replica. This enables stateless deployments.
- **Manual restore** (`taskprim restore` / `stateprim restore`): Point-in-time recovery. Overwrites the local database with the latest replica.

## Authentication

Authentication is only active in **serve** and **MCP SSE** modes. CLI mode uses filesystem permissions.

- API keys are configured in `config.yaml` and mapped to human-readable names
- Keys are validated using **constant-time comparison** (prevents timing attacks)
- When no keys are configured, the server runs in **open mode** (all requests allowed)
- The authenticated key's `name` is injected into the request context and used as the `source` field for created tasks/records

## HTTP Server

The HTTP server wraps `net/http` with:

- **Graceful shutdown**: Listens for SIGINT/SIGTERM, gives in-flight requests 10 seconds to complete
- **Timeouts**: 30s read, 30s write, 60s idle
- **Middleware chain**: RequestID → Logging → Recovery → Auth → Handler

## SQLite

Pure Go SQLite via `modernc.org/sqlite` (no CGo). This simplifies cross-compilation, especially for ARM64 (Raspberry Pi).

- **WAL mode**: concurrent readers during writes, required for Litestream and serve mode
- **Foreign keys**: enforced for referential integrity
- **Busy timeout**: 5 seconds to prevent lock contention errors
- **Embedded migrations**: SQL files are embedded via `embed.FS` for single-binary deployment
- **In-memory mode**: available for tests (`db.OpenInMemory()`)

## Build

The monorepo uses a Go workspace (`go.work`) with three modules. The Makefile provides:

| Target | Description |
|--------|-------------|
| `make build` | Compile `bin/taskprim` and `bin/stateprim` |
| `make build-pi` | Cross-compile for ARM64 Linux |
| `make test` | Run all tests with race detector |
| `make lint` | Run `go vet` across all modules |
| `make fmt` | Format all code with `gofmt` |
| `make tidy` | Run `go mod tidy` for all modules |
| `make all` | tidy → fmt → lint → test → build |
