# primkit

Agent-native operational primitives. Single Go binaries with embedded SQLite that give AI agents structured, queryable infrastructure for the two things every production agent needs: **task management** and **operational state**.

## Primitives

### taskprim

Task management for agents and the humans they work with. Replaces Todoist-style tools with a purpose-built primitive that treats tasks as first-class objects with explicit lifecycle (`open` → `done` | `killed`), list-based partitioning, freeform labels, and per-agent seen-tracking.

### stateprim

Operational state persistence. Key-value state, dedup lookups, and append-only action logs — the three access patterns every production agent duct-tapes together with markdown files. Namespaced, queryable, and durable.

### primkit (shared)

The shared foundation: config management, SQLite + Litestream replication, HTTP server, MCP server, and authentication. Both primitives are built on the same chassis.

## Architecture

Each primitive is a **single Go binary** with three interfaces:

- **CLI mode** (default) — direct shell access, no infrastructure needed
- **serve mode** — HTTP API for remote clients
- **mcp mode** — Model Context Protocol server for agent ecosystem integration

Storage is **embedded SQLite** (WAL mode) with built-in **Litestream** replication to object storage (S3, R2, B2, GCS).

Configuration follows **12-factor**: YAML config files with environment variable overrides.

## Status

Under active development. See [EXECUTION_GUIDE.md](./EXECUTION_GUIDE.md) for the implementation roadmap.

## License

MIT License — see [LICENSE](./LICENSE).
