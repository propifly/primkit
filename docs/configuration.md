# Configuration Reference

All three primitives (**taskprim**, **stateprim**, and **knowledgeprim**) read the same YAML configuration format. Copy `config.example.yaml` to `config.yaml` and edit to suit your environment. knowledgeprim adds additional fields for embedding and auto-connect.

## Resolution Order

Configuration values are resolved in three layers (highest precedence wins):

1. **Defaults** — hardcoded sensible values (e.g., port 8090)
2. **YAML file** — values from your `config.yaml`, with `${ENV_VAR}` interpolation
3. **Environment overrides** — prefix-based env vars (`TASKPRIM_*`, `STATEPRIM_*`, or `KNOWLEDGEPRIM_*`)

## Full YAML Spec

```yaml
storage:
  db: ~/.taskprim/default.db
  replicate:
    enabled: false
    provider: r2
    bucket: ${R2_BUCKET}
    path: taskprim.db
    endpoint: ${R2_ENDPOINT}
    access_key_id: ${R2_ACCESS_KEY_ID}
    secret_access_key: ${R2_SECRET_ACCESS_KEY}

auth:
  keys:
    - key: "tp_sk_your_api_key_here"
      name: "johanna"

server:
  port: 8090

taskprim:
  default_list: ""

# knowledgeprim-only settings (ignored by taskprim and stateprim)
embedding:
  provider: ""              # gemini | openai | custom (empty = disabled)
  model: ""                 # e.g., text-embedding-004 (Gemini), text-embedding-3-small (OpenAI)
  dimensions: 768           # expected output dimensions
  api_key: ${EMBEDDING_API_KEY}
  endpoint: ""              # custom endpoint for local models or proxies

auto_connect:
  enabled: true             # auto-link new entities to similar ones on capture
  threshold: 0.35           # cosine distance threshold (lower = more similar)
  max_connections: 10       # max auto-connections per capture
```

## Fields

### `storage`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `db` | string | `~/.taskprim/default.db` | Path to the SQLite database file. Created automatically if it doesn't exist. Parent directories are created as needed. |

### `storage.replicate`

Litestream replication configuration. When enabled, WAL frames are continuously streamed to object storage for durability. Replication runs for all commands (CLI, serve, MCP).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable WAL replication to object storage. |
| `provider` | string | — | Object storage provider: `r2`, `s3`, `b2`, or `gcs`. |
| `bucket` | string | — | Bucket name in the object storage provider. |
| `path` | string | — | Object path within the bucket (e.g., `taskprim.db`). |
| `endpoint` | string | — | Custom S3-compatible endpoint. Required for R2 and B2. Not needed for AWS S3. |
| `access_key_id` | string | — | Access key for the object storage provider. Use `${ENV_VAR}` interpolation. |
| `secret_access_key` | string | — | Secret key for the object storage provider. Use `${ENV_VAR}` interpolation. |

### `auth`

Authentication is only used in **serve** and **MCP SSE** modes. CLI mode relies on filesystem permissions.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `keys` | list | `[]` | List of API keys and their human-readable names. When empty, authentication is disabled (open mode). |

Each key entry:

| Field | Type | Description |
|-------|------|-------------|
| `key` | string | The API key value. Passed as `Authorization: Bearer <key>`. |
| `name` | string | Human-readable name mapped to this key. Used as the `source` field when creating tasks or records via the API, so you can tell who created what. |

### `server`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | int | `8090` | HTTP server port for `serve` and MCP SSE transport. |

### `taskprim`

taskprim-specific settings (ignored by stateprim).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `default_list` | string | `""` | Default list name for new tasks when `--list` is omitted. |

### `embedding` (knowledgeprim only)

Vector embedding configuration. When a provider is configured, knowledgeprim generates embeddings on `capture` and enables vector and hybrid search modes.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `provider` | string | `""` | Embedding provider: `gemini`, `openai`, or `custom`. Empty string disables embedding. |
| `model` | string | `""` | Provider-specific model name (e.g., `text-embedding-004` for Gemini, `text-embedding-3-small` for OpenAI). |
| `dimensions` | int | `768` | Expected output vector dimensions. Must match the model's output. |
| `api_key` | string | — | API key for the embedding provider. Use `${ENV_VAR}` interpolation. |
| `endpoint` | string | `""` | Custom endpoint URL. Required for `custom` provider. Optional for `openai` (overrides default). Unused for `gemini`. |

**Supported providers:**

| Provider | Model | Dimensions | Notes |
|----------|-------|------------|-------|
| `gemini` | `text-embedding-004` | 768 | Google Gemini |
| `openai` | `text-embedding-3-small` | 1536 | OpenAI |
| `openai` | `text-embedding-3-large` | 3072 | OpenAI (highest quality) |
| `custom` | Any | Configurable | Any OpenAI-compatible endpoint (local models, proxies) |

### `auto_connect` (knowledgeprim only)

Auto-connect configuration. When embedding is enabled, new entities are automatically linked to semantically similar existing entities on capture.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Enable auto-connect when capturing with embeddings. |
| `threshold` | float | `0.35` | Cosine distance threshold. Lower = more similar. Entities below this threshold get automatic `similar_to` edges. |
| `max_connections` | int | `10` | Maximum number of auto-connections created per capture. |

## Environment Variable Interpolation

Any value in the YAML file can reference environment variables using `${VAR_NAME}` syntax. Variables are interpolated at startup. Missing variables resolve to an empty string.

```yaml
storage:
  replicate:
    access_key_id: ${R2_ACCESS_KEY_ID}
    secret_access_key: ${R2_SECRET_ACCESS_KEY}
```

This lets you keep secrets out of the config file while still having a single file for all other settings.

## Environment Variable Overrides

Environment variables with the appropriate prefix override any value from the YAML file. This is useful for container deployments or CI where you don't want to mount a config file.

### taskprim prefix: `TASKPRIM_`

| Env Var | Overrides |
|---------|-----------|
| `TASKPRIM_DB` | `storage.db` |
| `TASKPRIM_SERVER_PORT` | `server.port` |
| `TASKPRIM_REPLICATE_ENABLED` | `storage.replicate.enabled` |
| `TASKPRIM_REPLICATE_PROVIDER` | `storage.replicate.provider` |
| `TASKPRIM_REPLICATE_BUCKET` | `storage.replicate.bucket` |
| `TASKPRIM_REPLICATE_PATH` | `storage.replicate.path` |
| `TASKPRIM_REPLICATE_ENDPOINT` | `storage.replicate.endpoint` |
| `TASKPRIM_REPLICATE_ACCESS_KEY_ID` | `storage.replicate.access_key_id` |
| `TASKPRIM_REPLICATE_SECRET_ACCESS_KEY` | `storage.replicate.secret_access_key` |

### stateprim prefix: `STATEPRIM_`

Same mapping as taskprim, with `STATEPRIM_` prefix:

| Env Var | Overrides |
|---------|-----------|
| `STATEPRIM_DB` | `storage.db` |
| `STATEPRIM_SERVER_PORT` | `server.port` |
| `STATEPRIM_REPLICATE_ENABLED` | `storage.replicate.enabled` |
| `STATEPRIM_REPLICATE_PROVIDER` | `storage.replicate.provider` |
| `STATEPRIM_REPLICATE_BUCKET` | `storage.replicate.bucket` |
| `STATEPRIM_REPLICATE_PATH` | `storage.replicate.path` |
| `STATEPRIM_REPLICATE_ENDPOINT` | `storage.replicate.endpoint` |
| `STATEPRIM_REPLICATE_ACCESS_KEY_ID` | `storage.replicate.access_key_id` |
| `STATEPRIM_REPLICATE_SECRET_ACCESS_KEY` | `storage.replicate.secret_access_key` |

### knowledgeprim prefix: `KNOWLEDGEPRIM_`

Same storage/server mapping as above, plus embedding-specific overrides:

| Env Var | Overrides |
|---------|-----------|
| `KNOWLEDGEPRIM_DB` | `storage.db` |
| `KNOWLEDGEPRIM_SERVER_PORT` | `server.port` |
| `KNOWLEDGEPRIM_REPLICATE_ENABLED` | `storage.replicate.enabled` |
| `KNOWLEDGEPRIM_REPLICATE_PROVIDER` | `storage.replicate.provider` |
| `KNOWLEDGEPRIM_REPLICATE_BUCKET` | `storage.replicate.bucket` |
| `KNOWLEDGEPRIM_REPLICATE_PATH` | `storage.replicate.path` |
| `KNOWLEDGEPRIM_REPLICATE_ENDPOINT` | `storage.replicate.endpoint` |
| `KNOWLEDGEPRIM_REPLICATE_ACCESS_KEY_ID` | `storage.replicate.access_key_id` |
| `KNOWLEDGEPRIM_REPLICATE_SECRET_ACCESS_KEY` | `storage.replicate.secret_access_key` |
| `KNOWLEDGEPRIM_EMBEDDING_PROVIDER` | `embedding.provider` |
| `KNOWLEDGEPRIM_EMBEDDING_MODEL` | `embedding.model` |
| `KNOWLEDGEPRIM_EMBEDDING_DIMENSIONS` | `embedding.dimensions` |
| `KNOWLEDGEPRIM_EMBEDDING_API_KEY` | `embedding.api_key` |
| `KNOWLEDGEPRIM_EMBEDDING_ENDPOINT` | `embedding.endpoint` |

## Examples

### Minimal (local development)

```yaml
storage:
  db: ~/.taskprim/default.db
```

Everything else uses defaults. No auth, no replication.

### With Cloudflare R2 replication

```yaml
storage:
  db: /data/taskprim.db
  replicate:
    enabled: true
    provider: r2
    bucket: primkit-backups
    path: taskprim.db
    endpoint: https://abc123.r2.cloudflarestorage.com
    access_key_id: ${R2_ACCESS_KEY_ID}
    secret_access_key: ${R2_SECRET_ACCESS_KEY}

auth:
  keys:
    - key: "tp_sk_prod_key_1"
      name: "api-server"
    - key: "tp_sk_prod_key_2"
      name: "agent-01"

server:
  port: 8090
```

### Environment-only (no config file)

```bash
export TASKPRIM_DB=/data/taskprim.db
export TASKPRIM_SERVER_PORT=9090
export TASKPRIM_REPLICATE_ENABLED=true
export TASKPRIM_REPLICATE_PROVIDER=s3
export TASKPRIM_REPLICATE_BUCKET=my-bucket
export TASKPRIM_REPLICATE_ACCESS_KEY_ID=AKIA...
export TASKPRIM_REPLICATE_SECRET_ACCESS_KEY=secret...

taskprim serve
```

When no `--config` flag is provided, defaults are used and then env overrides are applied.

### knowledgeprim with Gemini embedding

```yaml
storage:
  db: ~/.knowledgeprim/default.db

embedding:
  provider: gemini
  model: text-embedding-004
  dimensions: 768
  api_key: ${GEMINI_API_KEY}

auto_connect:
  enabled: true
  threshold: 0.35
  max_connections: 10

server:
  port: 8092
```

### knowledgeprim without embedding (FTS-only)

```yaml
storage:
  db: ~/.knowledgeprim/default.db
```

Works immediately. Full-text search, manual edges, graph traversal, and discovery all work without embedding.

## Database Path Resolution

The database path is resolved in this order:

1. `--db` flag (if provided)
2. Environment variable (`TASKPRIM_DB`, `STATEPRIM_DB`, or `KNOWLEDGEPRIM_DB`)
3. Home directory default (`~/.taskprim/default.db`, `~/.stateprim/default.db`, or `~/.knowledgeprim/default.db`)

The directory is created automatically if it doesn't exist.

## SQLite Pragmas

The following SQLite pragmas are applied automatically when the database is opened:

| Pragma | Value | Purpose |
|--------|-------|---------|
| `journal_mode` | `WAL` | Allows concurrent reads during writes. Required for Litestream replication and HTTP serve mode. |
| `foreign_keys` | `ON` | Enforces referential integrity (e.g., task labels reference tasks). |
| `busy_timeout` | `5000` | Waits up to 5 seconds for locks instead of failing immediately, preventing "database is locked" errors. |
