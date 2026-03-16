# Architecture

Primkit is a monorepo containing four primitives (**taskprim**, **stateprim**, **knowledgeprim**, and **queueprim**) and a shared infrastructure library (**primkit**). All primitives follow identical layered architecture — the only differences are the domain model, store operations, and (for knowledgeprim) the embedding layer.

## Repository Structure

```
primkit/
├── go.work                  # Go workspace (5 modules)
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
├── stateprim/               # state persistence primitive (module: github.com/propifly/primkit/stateprim)
│   ├── cmd/stateprim/       # main.go entry point
│   └── internal/
│       ├── model/           # Record, QueryFilter
│       ├── store/           # Store interface + SQLite implementation
│       ├── cli/             # Cobra commands (set, get, append, query, ...)
│       ├── api/             # HTTP API handler
│       └── mcpserver/       # MCP tool registration
├── knowledgeprim/           # knowledge graph primitive (module: github.com/propifly/primkit/knowledgeprim)
│   ├── cmd/knowledgeprim/   # main.go entry point
│   └── internal/
│       ├── model/           # Entity, Edge, SearchFilter, TraversalOpts, DiscoverOpts
│       ├── store/           # Store interface + SQLite implementation (FTS5, vectors)
│       ├── embed/           # Embedding provider abstraction (Gemini, OpenAI, custom)
│       ├── cli/             # Cobra commands (capture, search, connect, discover, ...)
│       ├── api/             # HTTP API handler
│       └── mcpserver/       # MCP tool registration
└── queueprim/               # work queue primitive (module: github.com/propifly/primkit/queueprim)
    ├── cmd/queueprim/       # main.go entry point
    └── internal/
        ├── model/           # Job, Filter, Priority, Status, QueueInfo, Stats
        ├── store/           # Store interface + SQLite implementation
        ├── cli/             # Cobra commands (enqueue, dequeue, complete, fail, ...)
        ├── api/             # HTTP API handler
        └── mcpserver/       # MCP tool registration
```

## Layered Design

Dependencies flow strictly downward. No lateral dependencies between sibling layers.

```
┌──────────────────────────────────────────────────────────┐
│                    Access Interfaces                      │
│  ┌───────┐    ┌──────────┐    ┌───────────────┐          │
│  │  CLI  │    │ HTTP API │    │  MCP Server   │          │
│  │(cobra)│    │ (net/http)│   │  (mcp-go)     │          │
│  └───┬───┘    └────┬─────┘    └──────┬────────┘          │
│      │             │                 │                    │
│      └─────────────┼─────────────────┘                    │
│                    │                                      │
│    ┌───────────────┼───────────────────┐                  │
│    │               │                   │                  │
│    │         ┌─────▼─────┐     ┌───────▼────────┐        │
│    │         │   Store    │     │    Embedder    │        │
│    │         │ (interface)│     │  (interface)   │        │
│    │         └─────┬─────┘     │ knowledgeprim  │        │
│    │               │           │     only       │        │
│    │         ┌─────▼─────┐     └────────────────┘        │
│    │         │   Model   │ ◄── structs,                  │
│    │         │           │     validation,               │
│    │         └─────┬─────┘     state machine             │
│    │               │                                      │
│    └───────┬───────┼────────────────┐                     │
│            │       │                │                     │
│  ┌─────────▼┐ ┌────▼──────┐  ┌─────▼──────┐             │
│  │  config  │ │     db     │  │  replicate │             │
│  │  (YAML)  │ │  (SQLite)  │  │(Litestream)│             │
│  └──────────┘ └────────────┘  └────────────┘             │
│                                                          │
│                 primkit (shared library)                  │
└──────────────────────────────────────────────────────────┘
```

> **Note:** The Embedder interface is unique to knowledgeprim. The background sweeper goroutine (for expired claim release) is unique to queueprim. taskprim, stateprim, and queueprim do not have an embedding layer.

### Key Constraint

**CLI, API, and MCP are sibling consumers** of the Store interface. They never depend on each other. This means:

- You can use CLI without the HTTP server
- You can use the API without MCP
- Any new access interface just imports the Store

## Store Interface

The Store is the central abstraction. Each primitive defines its own interface in `internal/store/store.go`.

### taskprim Store (21 operations)

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
| `AddDep` | Add a dependency edge (with cycle detection via recursive CTE) |
| `RemoveDep` | Remove a dependency edge |
| `Deps` | List tasks that a given task depends on |
| `Dependents` | List tasks that depend on a given task (reverse lookup) |
| `Frontier` | Open tasks with all dependencies resolved or no dependencies |
| `DepEdges` | Raw dependency edges, optionally filtered by list |
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

### knowledgeprim Store (23 operations)

| Operation | Description |
|-----------|-------------|
| `CaptureEntity` | Persist a new entity with optional embedding vector |
| `GetEntity` | Retrieve a single entity by ID (includes edges) |
| `UpdateEntity` | Partial update to mutable fields |
| `DeleteEntity` | Remove entity and all connected edges |
| `CreateEdge` | Create a weighted, contextualized connection |
| `UpdateEdge` | Update edge context or weight |
| `StrengthenEdge` | Increment an edge's weight by 1.0 |
| `DeleteEdge` | Remove a connection between entities |
| `SearchFTS` | Full-text search via FTS5 (BM25 ranking) |
| `SearchVector` | Semantic search via cosine distance on embeddings |
| `SearchHybrid` | Combined FTS + vector via Reciprocal Rank Fusion (k=60) |
| `Related` | Multi-hop graph traversal with direction and weight filters |
| `Discover` | Pattern detection: orphans, clusters, bridges, temporal groups, weak edges |
| `ListTypes` | All entity types with counts |
| `ListRelationships` | All relationship types with counts |
| `Stats` | Aggregate counts (entities, edges, vectors, orphans, DB size) |
| `ExportEntities` | Full export with optional type filter |
| `ImportEntities` | Bulk import preserving IDs |
| `GetEmbeddingMeta` | Fetch the stored embedding provider/model metadata for this database |
| `SetEmbeddingMeta` | Write or overwrite the embedding metadata record |
| `StripVectors` | Delete all embedding vectors and metadata (reverts to FTS5-only) |
| `UpdateEntityVector` | Upsert a single entity's embedding vector (used by `re-embed`) |
| `Close` | Release database connection |

### queueprim Store (16 operations)

| Operation | Description |
|-----------|-------------|
| `EnqueueJob` | Persist a new job (store assigns ID, status, timestamps) |
| `DequeueJob` | Atomically claim the next available job in a queue (status=pending AND visible_after ≤ now) |
| `CompleteJob` | Mark a claimed job as done; optionally store output payload |
| `FailJob` | Mark a claimed job as failed; retries if retries remain, otherwise moves to dead |
| `ReleaseJob` | Return a claimed job to pending immediately (unclaim) |
| `ExtendJob` | Extend a claimed job's visibility timeout to prevent auto-release |
| `PeekJob` | Inspect the next available job without claiming it |
| `GetJob` | Retrieve a single job by ID |
| `ListJobs` | Filter and list jobs (by queue, status, type, age) |
| `ListQueues` | All named queues with job counts by status |
| `Stats` | Aggregate counts across all queues |
| `PurgeJobs` | Delete jobs matching queue + status + age criteria; returns count |
| `ExportJobs` | Full export of all jobs in a queue |
| `ImportJobs` | Bulk import preserving original IDs |
| `SweepExpiredClaims` | Release claimed jobs whose visibility_after has passed; called by background sweeper |
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

**Dependency graph:**

```
DepEdge {
    TaskID    string  // the task that is blocked
    DependsOn string  // the task it depends on
}
```

Stored in `task_deps` table with composite primary key `(task_id, depends_on)` and a self-reference check (`task_id != depends_on`). Cycle detection is enforced via recursive CTE on `AddDep`. `waiting_on` (freeform text for external/human blockers) and `task_deps` (structural task-to-task edges) coexist — they serve different purposes.

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

### knowledgeprim: Entity + Edge

```
Entity {
    ID             string           // e_<nanoid>, assigned by store
    Type           string           // required: entity type (article, concept, pattern, etc.)
    Title          string           // required: entity title
    Body           *string          // optional: entity body text
    URL            *string          // optional: source URL
    Source         string           // required: who captured it
    Properties     json.RawMessage  // optional: custom JSON
    CreatedAt      time.Time        // assigned by store
    UpdatedAt      time.Time        // assigned by store
    Edges          []*Edge          // populated on GetEntity
}

Edge {
    SourceID       string           // required: source entity ID
    TargetID       string           // required: target entity ID
    Relationship   string           // required: relationship type (relates_to, extends, etc.)
    Weight         float64          // starts at 1.0, grows via strengthen
    Context        *string          // optional: WHY this connection exists
    CreatedAt      time.Time        // assigned by store
    UpdatedAt      time.Time        // assigned by store
}

EmbeddingMeta {
    Provider   string     // embedding provider name (e.g., "gemini", "openai")
    Model      string     // model name (e.g., "text-embedding-004")
    Dimensions int        // vector dimensions produced by this model
    CreatedAt  time.Time  // when the metadata was first recorded
}
```

`EmbeddingMeta` is a single-row record (enforced by `CHECK (id = 1)` in SQLite) that tracks which embedding provider and model generated the vectors in this database. One row per `.db` file. Used by `CheckEmbeddingMeta` to prevent silent degradation when the configured provider changes.

**Entity types** are freeform strings — agents define their own vocabulary (e.g., `article`, `thought`, `concept`, `pattern`, `observation`, `decision`, `bug`).

**Relationship types** are also freeform (e.g., `relates_to`, `contradicts`, `extends`, `inspired_by`, `applies_to`, `similar_to`).

**Three search modes:**

1. **FTS** — keyword search via SQLite FTS5, BM25 ranking
2. **Vector** — semantic search via cosine distance on embeddings
3. **Hybrid** — combines FTS + vector results via Reciprocal Rank Fusion (k=60)

**Discovery operations** surface non-obvious patterns:
- **Orphans** — entities with no edges
- **Clusters** — densely connected entity groups
- **Bridges** — high-degree entities connecting separate clusters
- **Temporal** — entity type distribution over time periods
- **Weak edges** — edges missing context prose

### queueprim: Job

```
Job {
    ID            string           // q_<nanoid>, assigned by store
    Queue         string           // required: named queue (slashes allowed, e.g., infra/prod)
    Type          string           // optional: job type category for type-filtered dequeue
    Priority      Priority         // high | normal (default) | low
    Payload       json.RawMessage  // required: arbitrary JSON work description
    Status        Status           // pending → claimed → done | failed | dead
    ClaimedBy     *string          // set on dequeue: worker name
    ClaimedAt     *time.Time       // set on dequeue
    VisibleAfter  time.Time        // delayed jobs: not visible until this time
    CompletedAt   *time.Time       // set on complete
    Output        json.RawMessage  // optional: worker result payload
    FailureReason *string          // set on fail
    AttemptCount  int              // incremented on each dequeue
    MaxRetries    int              // 0 = one-shot; >0 = retry up to N times before dead
    CreatedAt     time.Time        // assigned by store
    UpdatedAt     time.Time        // assigned by store
}
```

**State machine:**

```
  enqueue()           dequeue()           complete()
  ─────────► pending ──────────► claimed ──────────► done
                                    │
                                    │  fail() + retries remain
                                    ├──────────────────────── → pending (re-queued)
                                    │
                                    │  fail() + retries exhausted
                                    │  fail(--dead)
                                    └──────────────────────── → dead
```

**Priority ordering:** high → normal → low. Within a priority level, ordering is FIFO.

**Visibility timeout:** Claimed jobs hold a `visible_after` lock. If a worker crashes without completing, a background sweeper goroutine releases the claim once `visible_after` passes, returning the job to `pending`.

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
- **Manual restore** (`taskprim restore` / `stateprim restore` / `knowledgeprim restore` / `queueprim restore`): Point-in-time recovery. Overwrites the local database with the latest replica.

## Embedding (knowledgeprim only)

knowledgeprim supports optional vector embeddings for semantic search. The embedding layer is a pluggable interface:

```go
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
    Dimensions() int
    Provider() string  // "gemini", "openai", or "custom"
    Model() string     // e.g., "text-embedding-004"
}
```

`Provider()` and `Model()` are used by the metadata safety layer to detect provider changes. Three implementations ship out of the box:

| Provider | Model | Dimensions |
|----------|-------|------------|
| `gemini` | `text-embedding-004` | 768 |
| `openai` | `text-embedding-3-small` | 1536 |
| `custom` | Any OpenAI-compatible endpoint | Configurable |

**Embedding is optional.** Without it, knowledgeprim still provides:
- Full-text search (FTS5/BM25)
- Manual edge creation
- Graph traversal
- Discovery operations

You only lose vector search and auto-connect.

### Embedding Metadata Safety

Each knowledgeprim database stores a single `EmbeddingMeta` row (in the `embedding_meta` table, `CHECK (id = 1)`) recording which provider and model generated the stored vectors. This prevents **silent degradation** when switching embedding providers — old 768-dimension Gemini vectors are incompatible with a new 1536-dimension OpenAI config.

**Flow on `capture` or `search --mode vector/hybrid`:**

```
CheckEmbeddingMeta(provider, model, dimensions)
  ├── No meta yet → OK (first embed will call EnsureEmbeddingMeta)
  ├── Meta matches config → OK
  └── Meta differs → ErrEmbeddingMismatch with clear message:
        "db uses gemini/text-embedding-004 (768d),
         config uses openai/text-embedding-3-small (1536d).
         Use --mode fts, run re-embed, or pass --force"
```

**Recovery options:**

| Option | When to use |
|--------|-------------|
| `knowledgeprim re-embed` | Switching to a new provider — re-generates all vectors |
| `knowledgeprim strip-vectors --confirm` | Dropping back to FTS5-only — removes all vectors and metadata |
| `--force` flag | Bypassing the check temporarily (risky — mixed-dimension vectors in one DB) |
| `--mode fts` on search | Read-only fallback that skips vector operations entirely |

### Auto-Connect

When embedding is configured, `CaptureEntity` can automatically link new entities to semantically similar existing ones:

1. Embed the new entity's title + body
2. Cosine distance search against all existing embeddings
3. Entities below the threshold get automatic `similar_to` edges
4. Configurable: threshold (default 0.35), max connections (default 10)

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

The monorepo uses a Go workspace (`go.work`) with five modules. The Makefile provides:

| Target | Description |
|--------|-------------|
| `make build` | Compile `bin/taskprim`, `bin/stateprim`, `bin/knowledgeprim`, and `bin/queueprim` |
| `make build-pi` | Cross-compile for ARM64 Linux |
| `make test` | Run all tests with race detector |
| `make lint` | Run `go vet` across all modules |
| `make fmt` | Format all code with `gofmt` |
| `make tidy` | Run `go mod tidy` for all modules |
| `make all` | tidy → fmt → lint → test → build |
