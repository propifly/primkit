-- Task dependency graph: directed edges between tasks.
--
-- A row (task_id, depends_on) means "task_id depends on depends_on."
-- Self-loops are rejected by the CHECK constraint.
-- Cascading delete removes edges when either task is deleted.

CREATE TABLE IF NOT EXISTS task_deps (
    task_id    TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, depends_on),
    CHECK (task_id != depends_on)
);

CREATE INDEX IF NOT EXISTS idx_task_deps_depends_on ON task_deps(depends_on);
