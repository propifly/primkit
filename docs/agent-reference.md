# Agent Reference

Structured reference for AI agents using primkit primitives. All commands, flags, JSON output schemas, decision trees, and error patterns.

For human-readable guides, see the [README](../README.md) and [knowledgeprim Guide](knowledgeprim.md). For configuration, see [configuration](configuration.md).

## Install

Install only the primitives you need. Each is a standalone binary.

**From GitHub releases (pre-built):**

```bash
# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m); [[ "$ARCH" == "x86_64" ]] && ARCH="amd64"

# Install all four (or pick the ones you need)
for bin in taskprim stateprim knowledgeprim queueprim; do
  curl -sL "https://github.com/propifly/primkit/releases/latest/download/${bin}_0.1.0_${OS}_${ARCH}.tar.gz" | tar xz
done
sudo mv taskprim stateprim knowledgeprim queueprim /usr/local/bin/
```

**From source (requires Go 1.22+):**

```bash
git clone https://github.com/propifly/primkit.git && cd primkit && make build
# Binaries: bin/taskprim, bin/stateprim, bin/knowledgeprim, bin/queueprim
```

**Verify:**

```bash
taskprim --help
stateprim --help
knowledgeprim --help
queueprim --help
```

No configuration required. Databases auto-create on first use at `~/<primitive>/default.db`.

---

## Global Flags (all primitives)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--db` | string | `~/<primitive>/default.db` | Path to SQLite database |
| `--config` | string | `config.yaml` | Path to config file |
| `--format` | string | `text` | Output format: `text`, `json` (taskprim also supports `quiet`) |

Always use `--format json` for programmatic consumption.

---

## taskprim

Task management with lifecycle tracking. Tasks follow: `open` -> `done` | `killed`.

### Commands

| Command | Args | Required Flags | Optional Flags |
|---------|------|---------------|----------------|
| `add <what>` | task description | `--list` | `--source`, `--label` (repeatable), `--waiting-on`, `--context`, `--parent` |
| `list` | — | — | `--list`, `--state` (open/done/killed), `--label`, `--source`, `--unseen-by`, `--seen-by`, `--since`, `--stale`, `--parent`, `--waiting`, `--format` |
| `get <id>` | task ID | — | — |
| `done <id>` | task ID | — | `--reason` |
| `kill <id>` | task ID | — | `--reason` |
| `edit <id>` | task ID | — | `--what`, `--list`, `--waiting-on`, `--context`, `--parent`, `--add-label`, `--del-label` |
| `seen <agent>` | agent name | — | `--ids` (repeatable) |
| `labels` | — | — | — |
| `label clear <id> <label>` | task ID, label | — | — |
| `lists` | — | — | — |
| `stats` | — | — | — |
| `export` | — | — | `--list`, `--state` |
| `import` | — | — | — |
| `restore` | — | — | `--timestamp`, `--source` |

### JSON Schemas

**Task object:**

```json
{
  "id": "t_x7k2m9p4n1",
  "list": "ops",
  "what": "Deploy v2 to staging",
  "source": "johanna",
  "state": "open",
  "waiting_on": "CI pipeline",
  "parent_id": "t_abc123",
  "context": "Blocked on flaky test fix",
  "labels": ["deploy", "urgent"],
  "created": "2026-03-04T10:00:00Z",
  "updated": "2026-03-04T10:00:00Z",
  "resolved_at": null,
  "resolved_reason": null
}
```

**`list --format json`** returns `[Task, ...]`. **`stats --format json`** returns:

```json
{
  "total_open": 12,
  "total_done": 45,
  "total_killed": 3
}
```

### Idempotency

| Command | Idempotent | Notes |
|---------|-----------|-------|
| `add` | No | Creates a new task every call. Deduplicate by checking `list --format json` first |
| `list`, `get`, `labels`, `lists`, `stats` | Yes | Read-only |
| `done`, `kill` | Yes | Calling on already-resolved task is a no-op |
| `edit` | Yes | Same edit applied twice produces same result |
| `seen` | Yes | Re-marking as seen updates the timestamp |
| `export` | Yes | Read-only |

### Decision Tree

- **Create a task**: `add "<what>" --list <list> --source <agent-name>`
- **Check for existing work**: `list --list <list> --state open --format json`
- **Track what you've seen**: `seen <agent-name>` then later `list --unseen-by <agent-name>`
- **Mark complete**: `done <id>` (successful) or `kill <id>` (abandoned)
- **Subtasks**: `add "<what>" --list <list> --parent <parent-id>`

---

## stateprim

Namespaced key-value state with three access patterns: key-value, dedup, and append.

### Commands

| Command | Args | Required Flags | Optional Flags |
|---------|------|---------------|----------------|
| `set <ns> <key> <json-value>` | namespace, key, JSON | — | — |
| `get <ns> <key>` | namespace, key | — | — |
| `has <ns> <key>` | namespace, key | — | — |
| `set-if-new <ns> <key> <json-value>` | namespace, key, JSON | — | — |
| `delete <ns> <key>` | namespace, key | — | — |
| `append <ns> <json-value>` | namespace, JSON | — | — |
| `query <ns>` | namespace | — | `--since`, `--prefix`, `--count` |
| `purge <ns>` | namespace | — | `--before` |
| `namespaces` | — | — | — |
| `stats` | — | — | — |
| `export` | — | — | — |
| `import` | — | — | — |
| `restore` | — | — | `--timestamp`, `--source` |

### JSON Schemas

**Record object:**

```json
{
  "namespace": "config",
  "key": "app.theme",
  "value": "dark",
  "immutable": false,
  "created_at": "2026-03-04T10:00:00Z",
  "updated_at": "2026-03-04T10:00:00Z"
}
```

**`has` output (text):** `yes` or `no`. **`has --format json`:**

```json
{
  "exists": true
}
```

**`stats --format json`:**

```json
{
  "total_records": 156,
  "total_namespaces": 8
}
```

### Idempotency

| Command | Idempotent | Notes |
|---------|-----------|-------|
| `set` | Yes | Upsert — same key+namespace overwrites |
| `get`, `has`, `query`, `namespaces`, `stats` | Yes | Read-only |
| `set-if-new` | Yes | No-op if key exists (returns existing record) |
| `append` | No | Creates a new immutable record every call. Key is auto-generated |
| `delete` | Yes | Deleting non-existent key is a no-op |
| `purge` | No | Permanently removes records |

### Decision Tree

- **Store state**: `set <ns> <key> '<json>'`
- **Read state**: `get <ns> <key> --format json`
- **Check before acting (dedup)**: `has <ns> <key>` — returns `yes`/`no`
- **Create only if missing**: `set-if-new <ns> <key> '<json>'`
- **Log an event**: `append <ns> '<json>'` (immutable, timestamped)
- **Read recent events**: `query <ns> --since 24h --format json`

---

## knowledgeprim

Knowledge graph with typed entities, weighted edges, and hybrid search.

### Commands

| Command | Args | Required Flags | Optional Flags |
|---------|------|---------------|----------------|
| `capture` | — | `--type`, `--title` | `--body`, `--url`, `--source` (default: hostname), `--properties` (JSON), `--no-auto-connect`, `--threshold` (0.35), `--force` |
| `search <query>` | search text | — | `--type`, `--limit` (20), `--mode` (hybrid/fts/vector), `--force` |
| `get <id>` | entity ID | — | — |
| `edit <id>` | entity ID | — | `--title`, `--body`, `--properties` |
| `delete <id>` | entity ID | — | — |
| `connect <src-id> <tgt-id>` | source and target IDs | `--relationship` | `--context`, `--weight` (1.0) |
| `strengthen <src-id> <tgt-id> <rel>` | source ID, target ID, relationship | — | — |
| `edge-edit <src-id> <tgt-id> <rel>` | source ID, target ID, relationship | — | `--context`, `--weight` |
| `disconnect <src-id> <tgt-id> <rel>` | source ID, target ID, relationship | — | — |
| `related <id>` | entity ID | — | `--depth` (1), `--direction` (both/outgoing/incoming), `--relationship`, `--min-weight` (0) |
| `discover` | — | — | `--orphans`, `--clusters`, `--bridges`, `--temporal`, `--weak-edges` (no flags = all) |
| `types` | — | — | — |
| `relationships` | — | — | — |
| `stats` | — | — | — |
| `export` | — | — | `--type` |
| `import` | — | — | — |
| `re-embed` | — | — | — |
| `strip-vectors` | — | — | `--confirm` |
| `restore` | — | — | `--timestamp`, `--source` |

### JSON Schemas

**Entity object:**

```json
{
  "id": "e_x7k2m9p4n1",
  "type": "pattern",
  "title": "Retry with exponential backoff",
  "body": "Retry failed HTTP calls with 2^n backoff capped at 30s.",
  "url": "https://example.com/patterns/retry",
  "source": "coding-agent",
  "properties": {"language": "go", "category": "resilience"},
  "created_at": "2026-03-04T10:00:00Z",
  "updated_at": "2026-03-04T10:00:00Z",
  "edges": []
}
```

**Edge object (inside entity or discover results):**

```json
{
  "source_id": "e_x7k2m9p4n1",
  "target_id": "e_a3b4c5d6e7",
  "relationship": "extends",
  "weight": 2.0,
  "context": "Adds jitter to the basic exponential backoff pattern",
  "created_at": "2026-03-04T10:00:00Z",
  "updated_at": "2026-03-04T10:00:00Z"
}
```

**`search --format json`** returns:

```json
[
  {
    "entity": { "id": "e_...", "type": "...", "title": "...", "..." : "..." },
    "score": 0.8542
  }
]
```

**`related --format json`** returns:

```json
[
  {
    "entity": { "id": "e_...", "type": "...", "title": "...", "..." : "..." },
    "relationship": "extends",
    "direction": "outgoing",
    "depth": 1,
    "weight": 2.0
  }
]
```

**`discover --format json`** returns:

```json
{
  "orphans": [{"id": "e_...", "type": "...", "title": "..."}],
  "clusters": [{"entities": [...], "size": 7}],
  "bridges": [{"entity": {...}, "edge_count": 12, "cluster_ids": [0, 2]}],
  "temporal": [{"period": "2026-W09", "type": "pattern", "count": 5}],
  "weak_edges": [{"source_id": "e_...", "target_id": "e_...", "relationship": "...", "weight": 1.0}]
}
```

**`stats --format json`:**

```json
{
  "entity_count": 234,
  "edge_count": 567,
  "vector_count": 234,
  "orphan_count": 12,
  "type_count": 6,
  "db_size_bytes": 1048576,
  "db_path": "/home/user/.knowledgeprim/default.db"
}
```

### Idempotency

| Command | Idempotent | Notes |
|---------|-----------|-------|
| `capture` | No | Creates a new entity every call. Search first to avoid duplicates |
| `search`, `get`, `related`, `discover`, `types`, `relationships`, `stats` | Yes | Read-only |
| `connect` | No | Fails if edge already exists (same source, target, relationship) |
| `strengthen` | No | Additive — increments weight by 1.0 each call |
| `edge-edit` | Yes | Same edit applied twice produces same result |
| `disconnect` | Yes | Deleting non-existent edge is a no-op |
| `edit` | Yes | Same edit applied twice produces same result |
| `delete` | Yes | Deleting non-existent entity is a no-op |
| `re-embed` | No | Re-generates all vectors with current provider. Expensive (API calls per entity) |
| `strip-vectors` | Yes | Deleting already-empty vectors is a no-op |

### Search Mode Decision Tree

```
Need to find something?
├── Know the exact term or identifier?
│   └── Use --mode fts
├── Looking for conceptually similar content?
│   ├── Embedding configured?
│   │   └── Use --mode vector
│   └── No embedding?
│       └── Use --mode fts (try synonyms in query)
└── General search (most cases)?
    └── Use --mode hybrid (default — gracefully degrades to FTS-only without embedding)
```

### Common Workflows

**Capture and connect:**

```bash
knowledgeprim capture --type pattern --title "..." --body "..." --source agent
# Parse the ID from the JSON output
knowledgeprim connect <new-id> <existing-id> --relationship extends \
  --context "Why these connect"
```

**Search before capture (avoid duplicates):**

```bash
RESULTS=$(knowledgeprim search "retry backoff" --type pattern --format json)
# If empty array, safe to capture. If not, consider strengthening existing edges instead.
```

**Periodic maintenance:**

```bash
knowledgeprim discover --orphans --weak-edges --format json
# Process orphans: connect or delete
# Process weak edges: add context via edge-edit
```

---

## queueprim

Persistent work queues with priority, retries, and atomic dequeue. Jobs follow: `pending` → `claimed` → `done` / `failed` / `dead`.

### Commands

| Command | Args | Required Flags | Optional Flags |
|---------|------|---------------|----------------|
| `enqueue` | — | `--queue`, `--payload` (JSON) | `--type`, `--priority` (high/normal/low), `--max-retries`, `--delay` (e.g. `5m`) |
| `dequeue` | — | `--queue` | `--worker`, `--timeout` (default 30m), `--type` |
| `peek` | — | `--queue` | — |
| `complete <id>` | job ID | — | `--output` (JSON) |
| `fail <id>` | job ID | — | `--reason`, `--dead` (force dead-letter) |
| `release <id>` | job ID | — | — |
| `extend <id>` | job ID | — | `--by` (duration, default 30m) |
| `get <id>` | job ID | — | — |
| `list` | — | — | `--queue`, `--status`, `--type`, `--older-than` |
| `queues` | — | — | — |
| `stats` | — | — | — |
| `purge` | — | `--queue`, `--status` | `--older-than` |
| `export` | — | — | `--queue` |
| `import` | — | — | — |
| `restore` | — | — | `--timestamp`, `--source` |

### JSON Schemas

**Job object:**

```json
{
  "id": "q_x7k2m9p4n1",
  "queue": "infra/fixes",
  "type": "ssh_auth_fail",
  "priority": "normal",
  "payload": {"host": "web-01", "issue": "disk_full"},
  "status": "claimed",
  "claimed_by": "johanna",
  "claimed_at": "2026-03-05T10:00:00Z",
  "visible_after": "2026-03-05T10:30:00Z",
  "completed_at": null,
  "output": null,
  "failure_reason": null,
  "attempt_count": 1,
  "max_retries": 3,
  "created_at": "2026-03-05T09:58:00Z",
  "updated_at": "2026-03-05T10:00:00Z"
}
```

**`queues` output:**

```json
[
  {"queue": "infra/fixes", "pending": 3, "claimed": 1, "done": 12, "failed": 0, "dead": 1}
]
```

**`stats` output:**

```json
{
  "total_pending": 5,
  "total_claimed": 2,
  "total_done": 148,
  "total_failed": 3,
  "total_dead": 1
}
```

**`dequeue` returns `204 No Content` (CLI: exit 0 with no output) when the queue is empty.**

### Idempotency

| Command | Idempotent | Notes |
|---------|-----------|-------|
| `enqueue` | No | Creates a new job every call |
| `dequeue` | No | Atomically claims and removes from pending pool |
| `peek`, `get`, `list`, `queues`, `stats` | Yes | Read-only |
| `complete` | No | Terminal — cannot re-complete |
| `fail` | No | Transitions status; retries decrement |
| `release` | Yes | Releasing an already-pending job is a no-op |
| `extend` | No | Additive — each call pushes the timeout further |
| `purge` | No | Permanently deletes matching jobs |

### Decision Tree

- **Producer enqueues work**: `enqueue --queue <q> --payload '<json>'`
- **Worker claims next job**: `dequeue --queue <q> --worker <name> --format json` → parse `id`
- **Long-running job (heartbeat)**: `extend <id> --by 30m` before timeout expires
- **Mark success**: `complete <id>` (optionally `--output '<json>'`)
- **Mark failure (retriable)**: `fail <id> --reason "..."` → retries if `max_retries > attempt_count`
- **Mark failure (permanent)**: `fail <id> --dead`
- **Inspect queue without consuming**: `peek --queue <q>`
- **Clean up old done jobs**: `purge --queue <q> --status done --older-than 7d`

---

## Error Patterns

All primitives return non-zero exit codes on error. Error messages go to stderr.

| Error | Cause | Recovery |
|-------|-------|----------|
| `"type is required"` | Missing `--type` flag on capture | Add `--type <type>` |
| `"title is required"` | Missing `--title` flag on capture | Add `--title "<title>"` |
| `"relationship is required"` | Missing `--relationship` on connect | Add `--relationship <rel>` |
| `"self-edges are not allowed"` | Source and target are the same entity | Use different entity IDs |
| `"vector search requires an embedding provider"` | Used `--mode vector` without embedding config | Configure embedding in config.yaml or use `--mode fts` |
| `"embedding model mismatch"` | Configured embedding provider differs from what's in the db | Use `--mode fts`, run `re-embed`, match config to db, or pass `--force` |
| `"entity not found"` | Entity ID doesn't exist | Verify ID with `search` or `types` |
| `"database is locked"` | Another process holds the SQLite lock | Retry after a short delay (SQLite busy timeout handles most cases) |
| `"list is required"` | taskprim `add` without `--list` | Add `--list <list>` |
| `"namespace is required"` | stateprim command without namespace arg | Provide namespace as first positional arg |
| `"queue is required"` | queueprim command without `--queue` | Add `--queue <name>` |
| `"payload is required"` | queueprim `enqueue` without `--payload` | Add `--payload '<json>'` |
| `"payload must be valid JSON"` | queueprim payload is not valid JSON | Wrap in single quotes; ensure valid JSON |
| `"job not found"` | Job ID doesn't exist | Verify ID with `list --format json` |
| `"invalid status transition"` | Operation not valid for current job status (e.g., completing a done job) | Check job status with `get <id>` first |
| `"queue is empty"` (CLI exit 0) | `dequeue` or `peek` on an empty queue | Normal — poll again or exit worker loop |

## Environment Variables

| Variable | Primitive | Overrides |
|----------|-----------|-----------|
| `TASKPRIM_DB` | taskprim | `storage.db` path |
| `STATEPRIM_DB` | stateprim | `storage.db` path |
| `KNOWLEDGEPRIM_DB` | knowledgeprim | `storage.db` path |
| `QUEUEPRIM_DB` | queueprim | `storage.db` path |
| `TASKPRIM_LIST` | taskprim | Default list name for new tasks |

## ID Formats

| Primitive | Prefix | Example | Generated by |
|-----------|--------|---------|-------------|
| taskprim | `t_` | `t_x7k2m9p4n1` | Store on `add` |
| knowledgeprim (entity) | `e_` | `e_a3b4c5d6e7` | Store on `capture` |
| queueprim | `q_` | `q_x7k2m9p4n1` | Store on `enqueue` |

stateprim uses user-provided namespace + key pairs, not generated IDs.
