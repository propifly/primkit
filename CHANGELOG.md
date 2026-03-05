# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- **queueprim**: Work queue primitive — CLI, HTTP API, MCP server
  - Persistent SQLite-backed job queue with priority, retries, and dead-letter support
  - Atomic dequeue with visibility timeout (`claimed` status prevents double-processing)
  - Job lifecycle: `pending` → `claimed` → `done` / `failed` / `dead`
  - Delayed jobs via `visible_after` timestamp; background sweeper releases expired claims
  - Slash-containing queue names supported (e.g., `infra/prod`)
  - MCP tools: `enqueue_job`, `dequeue_job`, `complete_job`, `fail_job`, `release_job`,
    `extend_job`, `list_jobs`, `get_job`, `list_queues`, `get_stats`, `purge_queue`
  - CLI commands mirroring all API/MCP operations

### Fixed

- **all prims**: Global env vars (`TASKPRIM_DB`, `STATEPRIM_DB`, `KNOWLEDGEPRIM_DB`,
  `*_SERVER_PORT`, `*_REPLICATE_*`) no longer override per-agent config files in
  multi-agent deployments. Previously, `LoadWithEnvOverrides` always applied env var
  overrides on top of the YAML file, so a single global env var would silently stomp
  the `storage.db` (or any other field) from every agent's `--config` file. The fix
  makes env var overrides conditional on the config file being absent: when `--config`
  is provided the file is authoritative; when it is absent env vars serve as the
  primary configuration mechanism (unchanged for container / CI deployments). Use
  `${VAR}` interpolation inside the YAML to inject secrets while keeping the file
  authoritative. The same guard is applied to the direct `os.Getenv` call in each
  prim's `PersistentPreRunE`.

  **Before (broken):**
  ```
  Effective precedence (always): --db flag → *PRIM_DB env var → storage.db config → default
  ```
  **After (fixed):**
  ```
  With --config:    --db flag → storage.db from config file → default
  Without --config: --db flag → *PRIM_DB env var → default
  ```

## [v0.3.0] - 2026-03-05

### Fixed
- **stateprim / taskprim / knowledgeprim**: `storage.db` config key is now
  honored as the database path. Previously, the DB path fallback chain ran
  before the config file was loaded, so `cfg.Storage.DB` was never consulted
  and the hardcoded default (`~/.{prim}/default.db`) always won. Precedence
  is now: `--db` flag → `*PRIM_DB` env var → `storage.db` config → default.
- **knowledgeprim**: `${...}` environment variable references in
  `embedding.api_key` (and all other `embedding.*` fields) are now expanded
  before YAML parsing. Previously `loadEmbedConfig` passed raw bytes directly
  to `yaml.Unmarshal` without calling `InterpolateEnvVars`, so a value like
  `api_key: ${OPENAI_API_KEY}` was forwarded literally to the provider,
  causing authentication failures at runtime.
- **all prims**: `storage.replicate.bucket` and `storage.replicate.endpoint`
  now have `${R2_BUCKET}` and `${R2_ENDPOINT}` references in the example
  configs. Previously these fields were left as empty strings with no `${...}`
  placeholder, making them impossible to set via env var interpolation even
  though the credential fields showed the pattern. All four R2 fields now use
  consistent `R2_*` naming (`R2_BUCKET`, `R2_ENDPOINT`, `R2_ACCESS_KEY_ID`,
  `R2_SECRET_ACCESS_KEY`). The `docs/configuration.md` "Full YAML Spec"
  example had the same stale `REPLICATE_ACCESS_KEY_ID` naming and is fixed.

## [v0.2.0] - 2026-03-04

### Added
- **knowledgeprim**: Knowledge graph primitive — CLI, HTTP API, MCP server
  - Typed entities with freeform properties and source tracking
  - Weighted, contextualized edges (store *why* things connect)
  - Hybrid search: FTS5 (BM25) + vector (cosine) + Reciprocal Rank Fusion
  - Graph traversal with depth, direction, and weight filters
  - Discovery operations: orphans, clusters, bridges, temporal patterns, weak edges
  - Optional vector embedding via Gemini, OpenAI, or any OpenAI-compatible endpoint
  - Auto-connect: new entities automatically linked to semantically similar ones
  - Export/import for data portability
  - **Embedding safety**: prevent silent degradation when switching embedding providers
    - `embedding_meta` table tracks which provider/model produced stored vectors
    - Mismatch detection on capture and search with clear error messages
    - `--force` flag to bypass mismatch check when needed
    - `re-embed` command to migrate all vectors to a new embedding provider
    - `strip-vectors` command to remove all embeddings and revert to FTS5-only
    - `Provider()` and `Model()` methods on Embedder interface for identity tracking

## [v0.1.0] - 2026-03-04

### Added
- **taskprim**: Full task management primitive — CLI, HTTP API, MCP server
  - Task lifecycle: open -> done | killed
  - Lists, labels, per-agent seen-tracking
  - Export/import for data portability
- **stateprim**: State persistence primitive — CLI, HTTP API, MCP server
  - Three access patterns: key-value state, dedup lookups, append log
  - Namespace-scoped records with JSON values
  - Query filtering by namespace, key prefix, time window
- **primkit**: Shared infrastructure library
  - SQLite with WAL mode, embedded migrations
  - YAML config with `${ENV_VAR}` interpolation
  - Bearer token authentication
  - HTTP server with middleware chain (logging, recovery, request ID)
  - MCP helpers for tool registration
- **Litestream replication**: Embedded WAL streaming to S3/R2/B2
  - Runs for all commands (CLI, serve, MCP)
  - Auto-restore on startup when local DB is missing
  - `restore` command for point-in-time recovery
  - Configurable via YAML or environment variables
