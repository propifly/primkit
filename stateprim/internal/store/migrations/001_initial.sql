-- Initial stateprim schema.
--
-- Single table design: all three access patterns (key-value, dedup, append)
-- share the same records table. The immutable flag distinguishes append
-- records from regular key-value entries.

CREATE TABLE IF NOT EXISTS records (
    namespace   TEXT NOT NULL,
    key         TEXT NOT NULL,
    value       TEXT NOT NULL,  -- JSON payload
    immutable   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  TIMESTAMP NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    PRIMARY KEY (namespace, key)
);

-- Index for namespace-scoped queries (the most common access pattern).
CREATE INDEX IF NOT EXISTS idx_records_namespace
    ON records(namespace);

-- Index for time-range queries within a namespace (since, purge).
CREATE INDEX IF NOT EXISTS idx_records_created
    ON records(namespace, created_at);

CREATE INDEX IF NOT EXISTS idx_records_updated
    ON records(namespace, updated_at);
