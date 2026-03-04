-- knowledgeprim initial schema: entities, edges, FTS5 index, vector storage.

CREATE TABLE entities (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL,
    title       TEXT NOT NULL,
    body        TEXT,
    url         TEXT,
    source      TEXT NOT NULL,
    properties  TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  TIMESTAMP NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE edges (
    source_id    TEXT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    target_id    TEXT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    relationship TEXT NOT NULL,
    weight       REAL NOT NULL DEFAULT 1.0,
    context      TEXT,
    created_at   TIMESTAMP NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at   TIMESTAMP NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE(source_id, target_id, relationship)
);

-- FTS5 external content table. Kept in sync via triggers below.
CREATE VIRTUAL TABLE entities_fts USING fts5(
    title,
    body,
    content='entities',
    content_rowid='rowid'
);

-- Vector storage: embeddings as BLOBs with dimension tracking.
CREATE TABLE entity_vectors (
    entity_id  TEXT PRIMARY KEY REFERENCES entities(id) ON DELETE CASCADE,
    embedding  BLOB NOT NULL,
    dimensions INTEGER NOT NULL
);

-- Triggers to keep FTS5 index in sync with entities table.
CREATE TRIGGER entities_ai AFTER INSERT ON entities BEGIN
    INSERT INTO entities_fts(rowid, title, body)
    VALUES (new.rowid, new.title, new.body);
END;

CREATE TRIGGER entities_ad AFTER DELETE ON entities BEGIN
    INSERT INTO entities_fts(entities_fts, rowid, title, body)
    VALUES ('delete', old.rowid, old.title, old.body);
END;

CREATE TRIGGER entities_au AFTER UPDATE ON entities BEGIN
    INSERT INTO entities_fts(entities_fts, rowid, title, body)
    VALUES ('delete', old.rowid, old.title, old.body);
    INSERT INTO entities_fts(rowid, title, body)
    VALUES (new.rowid, new.title, new.body);
END;

-- Indexes for common query patterns.
CREATE INDEX idx_entities_type ON entities(type);
CREATE INDEX idx_entities_source ON entities(source);
CREATE INDEX idx_entities_created ON entities(created_at);
CREATE INDEX idx_edges_source ON edges(source_id);
CREATE INDEX idx_edges_target ON edges(target_id);
CREATE INDEX idx_edges_relationship ON edges(relationship);
