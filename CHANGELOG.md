# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

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
