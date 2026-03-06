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


| Flag       | Type   | Default                    | Description                                                    |
| ---------- | ------ | -------------------------- | -------------------------------------------------------------- |
| `--db`     | string | `~/<primitive>/default.db` | Path to SQLite database                                        |
| `--config` | string | `config.yaml`              | Path to config file                                            |
| `--format` | string | `text`                     | Output format: `text`, `json` (taskprim also supports `quiet`) |


Always use `--format json` for programmatic consumption.

---

## taskprim

Task management with lifecycle tracking. Tasks follow: `open` -> `done` | `killed`.

### Commands


<!-- docgen:start:taskprim:commands -->
| Command | Synopsis | Flags |
|---------|----------|-------|
| `add` | `add <what>` | `--context` — additional context or notes; `--label` (default: `[]`) — labels (repeatable or comma-separated); `--list` (default: `default`) — list to add the task to; `--parent` — parent task ID for subtasks; `--source` (default: `cli`) — who created this task; `--waiting-on` — what this task is blocked on |
| `done` | `done <id> [id...]` | — |
| `edit` | `edit <id>` | `--add-label` (default: `[]`) — add labels (repeatable); `--context` — update context notes; `--del-label` (default: `[]`) — remove labels (repeatable); `--list` — move to a different list; `--parent` — set or clear parent task ID; `--waiting-on` — set or clear (empty string) waiting_on; `--what` — update the task description |
| `export` | `export` | `--list` — export only tasks from this list; `--state` — export only tasks in this state |
| `get` | `get <id>` | — |
| `import` | `import` | `--file` — path to JSON file (default: stdin) |
| `kill` | `kill <id>` | `--reason` — why this task is being dropped (required) |
| `labels` | `labels` | `--list` — only labels from tasks in this list |
| `labels clear` | `labels clear <label>` | `--list` — only clear from tasks in this list |
| `list` | `list` | `--label` (default: `[]`) — filter by label (repeatable, AND logic); `--list` — filter by list; `--seen-by` — tasks seen by this agent (use with --since); `--since` — time window for --seen-by (e.g., 24h, 7d); `--source` — filter by source; `--stale` — tasks not updated within duration (e.g., 7d); `--state` — filter by state: open, done, killed; `--unseen-by` — tasks not seen by this agent; `--waiting` — only tasks with waiting_on set |
| `lists` | `lists` | — |
| `restore` | `restore` | — |
| `seen` | `seen <agent> [task_ids...]` | `--list` — mark all open tasks in this list as seen |
| `stats` | `stats` | — |
<!-- docgen:end:taskprim:commands -->


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

`**list --format json**` returns `[Task, ...]`. `**stats --format json**` returns:

```json
{
  "total_open": 12,
  "total_done": 45,
  "total_killed": 3
}
```

### Idempotency


| Command                                   | Idempotent | Notes                                                                             |
| ----------------------------------------- | ---------- | --------------------------------------------------------------------------------- |
| `add`                                     | No         | Creates a new task every call. Deduplicate by checking `list --format json` first |
| `list`, `get`, `labels`, `lists`, `stats` | Yes        | Read-only                                                                         |
| `done`, `kill`                            | Yes        | Calling on already-resolved task is a no-op                                       |
| `edit`                                    | Yes        | Same edit applied twice produces same result                                      |
| `seen`                                    | Yes        | Re-marking as seen updates the timestamp                                          |
| `export`                                  | Yes        | Read-only                                                                         |


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


<!-- docgen:start:stateprim:commands -->
| Command | Synopsis | Flags |
|---------|----------|-------|
| `append` | `append <namespace> <json-value>` | — |
| `delete` | `delete <namespace> <key>` | — |
| `export` | `export` | `--namespace` — export only this namespace |
| `get` | `get <namespace> <key>` | — |
| `has` | `has <namespace> <key>` | — |
| `import` | `import` | `--file` — path to JSON file (default: stdin) |
| `namespaces` | `namespaces` | — |
| `purge` | `purge <namespace> <duration>` | — |
| `query` | `query <namespace>` | `--count` — return count only; `--prefix` — filter by key prefix; `--since` — only records updated within duration (e.g., 24h, 7d) |
| `restore` | `restore` | — |
| `set` | `set <namespace> <key> <json-value>` | `--immutable` — mark the record as immutable |
| `set-if-new` | `set-if-new <namespace> <key> <json-value>` | — |
| `stats` | `stats` | — |
<!-- docgen:end:stateprim:commands -->


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

`**has` output (text):** `yes` or `no`. `**has --format json`:**

```json
{
  "exists": true
}
```

`**stats --format json`:**

```json
{
  "total_records": 156,
  "total_namespaces": 8
}
```

### Idempotency


| Command                                      | Idempotent | Notes                                                            |
| -------------------------------------------- | ---------- | ---------------------------------------------------------------- |
| `set`                                        | Yes        | Upsert — same key+namespace overwrites                           |
| `get`, `has`, `query`, `namespaces`, `stats` | Yes        | Read-only                                                        |
| `set-if-new`                                 | Yes        | No-op if key exists (returns existing record)                    |
| `append`                                     | No         | Creates a new immutable record every call. Key is auto-generated |
| `delete`                                     | Yes        | Deleting non-existent key is a no-op                             |
| `purge`                                      | No         | Permanently removes records                                      |


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


<!-- docgen:start:knowledgeprim:commands -->
| Command | Synopsis | Flags |
|---------|----------|-------|
| `capture` | `capture` | `--body` — entity body text; `--force` — bypass embedding model mismatch check; `--no-auto-connect` — skip auto-connect; `--properties` — JSON properties; `--source` (default: `cli`) — who captured this; `--threshold` (default: `0.35`) — auto-connect cosine distance threshold; `--title` *(required)* — entity title; `--type` *(required)* — entity type (article, thought, concept, pattern, etc.); `--url` — source URL |
| `connect` | `connect <source-id> <target-id>` | `--context` — edge context (why this connection exists); `--relationship` *(required)* — relationship type; `--weight` (default: `1`) — edge weight |
| `delete` | `delete <id>` | — |
| `disconnect` | `disconnect <source-id> <target-id> <relationship>` | — |
| `discover` | `discover` | `--bridges` — find cross-cluster connectors; `--clusters` — find densely connected groups; `--orphans` — find entities with no edges; `--temporal` — show type distribution over time; `--weak-edges` — find edges with no context |
| `edge-edit` | `edge-edit <source-id> <target-id> <relationship>` | `--context` — edge context; `--weight` — edge weight |
| `edit` | `edit <id>` | `--body` — new body; `--properties` — JSON properties; `--title` — new title |
| `export` | `export` | `--type` — export only entities of this type |
| `get` | `get <id>` | — |
| `import` | `import` | — |
| `re-embed` | `re-embed` | — |
| `related` | `related <id>` | `--depth` (default: `1`) — traversal depth (hops); `--direction` (default: `both`) — edge direction: outgoing, incoming, both; `--min-weight` — minimum edge weight; `--relationship` — filter by relationship type |
| `relationships` | `relationships` | — |
| `restore` | `restore` | — |
| `search` | `search <query>` | `--force` — bypass embedding model mismatch check; `--limit` (default: `20`) — max results; `--mode` (default: `hybrid`) — search mode: hybrid, fts, vector; `--type` — filter by entity type |
| `stats` | `stats` | — |
| `strengthen` | `strengthen <source-id> <target-id> <relationship>` | — |
| `strip-vectors` | `strip-vectors` | `--confirm` — confirm destructive operation |
| `types` | `types` | — |
<!-- docgen:end:knowledgeprim:commands -->


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

`**search --format json**` returns:

```json
[
  {
    "entity": { "id": "e_...", "type": "...", "title": "...", "..." : "..." },
    "score": 0.8542
  }
]
```

`**related --format json**` returns:

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

`**discover --format json**` returns:

```json
{
  "orphans": [{"id": "e_...", "type": "...", "title": "..."}],
  "clusters": [{"entities": [...], "size": 7}],
  "bridges": [{"entity": {...}, "edge_count": 12, "cluster_ids": [0, 2]}],
  "temporal": [{"period": "2026-W09", "type": "pattern", "count": 5}],
  "weak_edges": [{"source_id": "e_...", "target_id": "e_...", "relationship": "...", "weight": 1.0}]
}
```

`**stats --format json`:**

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


| Command                                                                   | Idempotent | Notes                                                                            |
| ------------------------------------------------------------------------- | ---------- | -------------------------------------------------------------------------------- |
| `capture`                                                                 | No         | Creates a new entity every call. Search first to avoid duplicates                |
| `search`, `get`, `related`, `discover`, `types`, `relationships`, `stats` | Yes        | Read-only                                                                        |
| `connect`                                                                 | No         | Fails if edge already exists (same source, target, relationship)                 |
| `strengthen`                                                              | No         | Additive — increments weight by 1.0 each call                                    |
| `edge-edit`                                                               | Yes        | Same edit applied twice produces same result                                     |
| `disconnect`                                                              | Yes        | Deleting non-existent edge is a no-op                                            |
| `edit`                                                                    | Yes        | Same edit applied twice produces same result                                     |
| `delete`                                                                  | Yes        | Deleting non-existent entity is a no-op                                          |
| `re-embed`                                                                | No         | Re-generates all vectors with current provider. Expensive (API calls per entity) |
| `strip-vectors`                                                           | Yes        | Deleting already-empty vectors is a no-op                                        |


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

Queue names and payloads are **positional arguments**, not flags. Commands that act on a specific job take the job ID as a positional argument.


<!-- docgen:start:queueprim:commands -->
| Command | Synopsis | Flags |
|---------|----------|-------|
| `complete` | `complete <id>` | `--output` — JSON output payload from the worker |
| `dequeue` | `dequeue <queue>` | `--timeout` (default: `30m`) — visibility timeout for claimed job; `--timeout-wait` — max time to wait (e.g. 5m); 0 = wait forever; `--type` — only claim jobs of this type; `--wait` — block until a job appears; `--worker` — worker name for claimed_by tracking (default: hostname) |
| `enqueue` | `enqueue <queue> <json_payload>` | `--delay` — delay before job is visible, e.g. 5m, 1h; `--max-retries` — max retries before dead-letter (default 0 = one-shot); `--priority` (default: `normal`) — priority: high, normal, or low; `--type` — job type category for workers (e.g. ssh_auth_fail) |
| `export` | `export` | `--queue` — export only jobs in this queue (default: all) |
| `extend` | `extend <id>` | `--by` (default: `30m`) — extension duration, e.g. 30m, 1h |
| `fail` | `fail <id>` | `--dead` — force to dead-letter regardless of retry count; `--reason` — human-readable failure reason |
| `get` | `get <id>` | — |
| `import` | `import` | — |
| `list` | `list` | `--older-than` — only jobs created before now-duration, e.g. 1h; `--queue` — filter to this queue; `--status` — filter by status: pending, claimed, done, failed, dead; `--type` — filter by job type |
| `peek` | `peek <queue>` | — |
| `purge` | `purge <queue>` | `--older-than` — only purge jobs older than this duration, e.g. 7d, 24h; `--status` *(required)* — status to purge: done, dead, failed (required) |
| `queues` | `queues` | — |
| `release` | `release <id>` | — |
| `restore` | `restore` | — |
| `stats` | `stats` | — |
<!-- docgen:end:queueprim:commands -->


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

`**queues` output:**

```json
[
  {"queue": "infra/fixes", "pending": 3, "claimed": 1, "done": 12, "failed": 0, "dead": 1}
]
```

`**stats` output:**

```json
{
  "total_pending": 5,
  "total_claimed": 2,
  "total_done": 148,
  "total_failed": 3,
  "total_dead": 1
}
```

`**dequeue` exits 1 with `"queue is empty"` on stderr when the queue is empty (without `--wait`). Use `--wait` to block and poll every 2s until a job appears; combine with `--timeout-wait` to cap the wait duration.**

### Idempotency


| Command                                  | Idempotent | Notes                                           |
| ---------------------------------------- | ---------- | ----------------------------------------------- |
| `enqueue`                                | No         | Creates a new job every call                    |
| `dequeue`                                | No         | Atomically claims and removes from pending pool |
| `peek`, `get`, `list`, `queues`, `stats` | Yes        | Read-only                                       |
| `complete`                               | No         | Terminal — cannot re-complete                   |
| `fail`                                   | No         | Transitions status; retries decrement           |
| `release`                                | Yes        | Releasing an already-pending job is a no-op     |
| `extend`                                 | No         | Additive — each call pushes the timeout further |
| `purge`                                  | No         | Permanently deletes matching jobs               |


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


| Error                                            | Cause                                                                    | Recovery                                                                |
| ------------------------------------------------ | ------------------------------------------------------------------------ | ----------------------------------------------------------------------- |
| `"type is required"`                             | Missing `--type` flag on capture                                         | Add `--type <type>`                                                     |
| `"title is required"`                            | Missing `--title` flag on capture                                        | Add `--title "<title>"`                                                 |
| `"relationship is required"`                     | Missing `--relationship` on connect                                      | Add `--relationship <rel>`                                              |
| `"self-edges are not allowed"`                   | Source and target are the same entity                                    | Use different entity IDs                                                |
| `"vector search requires an embedding provider"` | Used `--mode vector` without embedding config                            | Configure embedding in config.yaml or use `--mode fts`                  |
| `"embedding model mismatch"`                     | Configured embedding provider differs from what's in the db              | Use `--mode fts`, run `re-embed`, match config to db, or pass `--force` |
| `"entity not found"`                             | Entity ID doesn't exist                                                  | Verify ID with `search` or `types`                                      |
| `"database is locked"`                           | Another process holds the SQLite lock                                    | Retry after a short delay (SQLite busy timeout handles most cases)      |
| `"list is required"`                             | taskprim `add` without `--list`                                          | Add `--list <list>`                                                     |
| `"namespace is required"`                        | stateprim command without namespace arg                                  | Provide namespace as first positional arg                               |
| `"queue is required"`                            | queueprim command without `--queue`                                      | Add `--queue <name>`                                                    |
| `"payload is required"`                          | queueprim `enqueue` without `--payload`                                  | Add `--payload '<json>'`                                                |
| `"payload must be valid JSON"`                   | queueprim payload is not valid JSON                                      | Wrap in single quotes; ensure valid JSON                                |
| `"job not found"`                                | Job ID doesn't exist                                                     | Verify ID with `list --format json`                                     |
| `"invalid status transition"`                    | Operation not valid for current job status (e.g., completing a done job) | Check job status with `get <id>` first                                  |
| `"queue is empty"` (CLI exit 0)                  | `dequeue` or `peek` on an empty queue                                    | Normal — poll again or exit worker loop                                 |


## Environment Variables


| Variable           | Primitive     | Overrides                       |
| ------------------ | ------------- | ------------------------------- |
| `TASKPRIM_DB`      | taskprim      | `storage.db` path               |
| `STATEPRIM_DB`     | stateprim     | `storage.db` path               |
| `KNOWLEDGEPRIM_DB` | knowledgeprim | `storage.db` path               |
| `QUEUEPRIM_DB`     | queueprim     | `storage.db` path               |
| `TASKPRIM_LIST`    | taskprim      | Default list name for new tasks |


## ID Formats


| Primitive              | Prefix | Example        | Generated by       |
| ---------------------- | ------ | -------------- | ------------------ |
| taskprim               | `t_`   | `t_x7k2m9p4n1` | Store on `add`     |
| knowledgeprim (entity) | `e_`   | `e_a3b4c5d6e7` | Store on `capture` |
| queueprim              | `q_`   | `q_x7k2m9p4n1` | Store on `enqueue` |


stateprim uses user-provided namespace + key pairs, not generated IDs.