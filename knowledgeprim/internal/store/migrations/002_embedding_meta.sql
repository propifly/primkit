-- Embedding metadata: tracks which embedding provider/model produced the
-- vectors in this database. Single-row table enforced by CHECK constraint.
-- Prevents silent degradation when switching embedding providers.

CREATE TABLE IF NOT EXISTS embedding_meta (
    id         INTEGER PRIMARY KEY CHECK (id = 1),
    provider   TEXT NOT NULL,
    model      TEXT NOT NULL,
    dimensions INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
