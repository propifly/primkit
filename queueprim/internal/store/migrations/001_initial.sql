-- Initial queueprim schema.
--
-- Core table:
--   jobs — the central entity with full queue lifecycle
--
-- Design: the primitive is dumb, the agent is smart. The schema stores
-- structured state and enforces ordering; agents provide job semantics
-- via the opaque payload and type fields.

CREATE TABLE IF NOT EXISTS jobs (
    id              TEXT PRIMARY KEY,
    queue           TEXT NOT NULL,
    type            TEXT,
    priority        TEXT NOT NULL DEFAULT 'normal'
                    CHECK (priority IN ('high', 'normal', 'low')),
    priority_rank   INTEGER NOT NULL DEFAULT 1,  -- 0=high, 1=normal, 2=low; for ORDER BY
    payload         TEXT NOT NULL,               -- arbitrary JSON
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'claimed', 'done', 'failed', 'dead')),
    claimed_by      TEXT,
    claimed_at      TIMESTAMP,
    visible_after   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at    TIMESTAMP,
    output          TEXT,                        -- JSON, optional worker result
    failure_reason  TEXT,
    attempt_count   INTEGER NOT NULL DEFAULT 0,
    max_retries     INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Primary dequeue index: filters pending+visible jobs, ordered by priority then FIFO.
CREATE INDEX IF NOT EXISTS idx_jobs_dequeue
    ON jobs (queue, priority_rank, created_at)
    WHERE status = 'pending';

-- Secondary index for list/filter queries.
CREATE INDEX IF NOT EXISTS idx_jobs_queue_status
    ON jobs (queue, status);

-- Index for the timeout sweeper: finds claimed jobs whose visibility has expired.
CREATE INDEX IF NOT EXISTS idx_jobs_timeout
    ON jobs (visible_after)
    WHERE status = 'claimed';
