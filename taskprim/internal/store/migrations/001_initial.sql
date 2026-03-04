-- Initial taskprim schema.
--
-- Core tables:
--   tasks       — the central entity with lifecycle (open → done | killed)
--   task_labels — many-to-many join table for freeform labels
--   seen        — per-agent seen-tracking for cross-session awareness
--
-- Design principle: "the primitive is dumb, the agent is smart." The schema
-- stores structured state; agents layer their own semantics via labels and
-- the context field.

CREATE TABLE IF NOT EXISTS tasks (
    id              TEXT PRIMARY KEY,
    list            TEXT NOT NULL,
    what            TEXT NOT NULL,
    source          TEXT NOT NULL,
    state           TEXT NOT NULL DEFAULT 'open' CHECK (state IN ('open', 'done', 'killed')),
    waiting_on      TEXT,
    parent_id       TEXT REFERENCES tasks(id),
    context         TEXT,
    created         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at     TIMESTAMP,
    resolved_reason TEXT
);

CREATE TABLE IF NOT EXISTS task_labels (
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    label   TEXT NOT NULL,
    PRIMARY KEY (task_id, label)
);

CREATE TABLE IF NOT EXISTS seen (
    agent   TEXT NOT NULL,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (agent, task_id)
);

-- Indexes for the most common query patterns.
-- (list, state) covers the primary listing query: "show me open tasks in this list."
CREATE INDEX IF NOT EXISTS idx_tasks_list_state ON tasks(list, state);
CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state);
CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);
CREATE INDEX IF NOT EXISTS idx_task_labels_label ON task_labels(label);
CREATE INDEX IF NOT EXISTS idx_seen_agent ON seen(agent);
