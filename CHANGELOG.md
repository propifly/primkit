# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

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
