# queueprim Guide

queueprim is a persistent work queue primitive. Producers enqueue jobs with a JSON payload; workers dequeue, process, and complete or fail them. Jobs survive process restarts, support priority ordering, retry automatically on failure, and route unrecoverable work to a dead-letter state.

This guide covers the concepts you need to use queueprim correctly: the visibility timeout (the central mechanism), the worker loop, how to design priorities and queues, and how to handle retries and dead-letter jobs. For installation and quick start, see the [README](../README.md). For all commands and flags, see the [agent reference](agent-reference.md).

---

## The Visibility Timeout

This is the one concept that makes everything else make sense.

When a worker calls `dequeue`, two things happen atomically: the job's status moves from `pending` to `claimed`, and a `visible_after` timestamp is set in the future (default: 30 minutes). The job is now invisible to other workers — they cannot dequeue it.

**The clock is running.** If the worker completes or fails the job before `visible_after`, all is well. If it doesn't — because the process crashed, the worker hung, or the job took longer than expected — a background sweeper goroutine detects that `visible_after` has passed and releases the job back to `pending`. No human intervention required. The job will be picked up again on the next `dequeue`.

This means:
- Workers can crash without losing work
- No explicit lock release is needed on crash — the timeout handles it
- A job being `claimed` does not mean it is being actively processed right now

The failure mode to avoid: a long-running job whose worker goes silent. The job will be re-queued at `visible_after` and claimed by another worker — or the same worker — while the original work is still in progress. The result is duplicate execution.

**The fix: extend before the timeout expires.**

```bash
# Worker is processing a long job (q_abc123, claimed with default 30m timeout)
# At the 20-minute mark, push the deadline another 30 minutes
queueprim extend q_abc123 --by 30m
```

Call `extend` periodically for any job that could outlast the claim window. Think of it as a heartbeat: the worker signals "I'm still alive and working."

---

## The Worker Loop

A worker is any process that repeatedly dequeues and processes jobs. The canonical loop:

```bash
#!/usr/bin/env bash
QUEUE="builds/ci"
WORKER="builder-01"

while true; do
  # Claim the next job. Exit 0 with no output if queue is empty.
  JOB=$(queueprim dequeue --queue "$QUEUE" --worker "$WORKER" --format json)

  if [ -z "$JOB" ]; then
    sleep 5
    continue
  fi

  ID=$(echo "$JOB" | jq -r '.id')
  PAYLOAD=$(echo "$JOB" | jq -c '.payload')

  # Process the job. For long work, heartbeat every 20 minutes.
  # (run your actual work here)
  do_work "$PAYLOAD"
  STATUS=$?

  if [ $STATUS -eq 0 ]; then
    queueprim complete "$ID" --output '{"status":"ok"}'
  else
    queueprim fail "$ID" --reason "do_work exited $STATUS"
  fi
done
```

Key points:
- `dequeue` exits 0 with no output when the queue is empty — check for empty output, not exit codes
- Always complete or fail every job you dequeue. Uncompleted jobs sit `claimed` until their timeout expires, then re-queue
- Pass `--worker` with a stable, meaningful name — it's stored on the job and visible in `list` output, which matters for debugging
- The `--output` JSON on `complete` is stored on the job and queryable later — use it to record results

### Type-Filtered Workers

If a queue contains heterogeneous job types, a worker can specialize:

```bash
# Only claim embedding jobs from the pipeline queue
queueprim dequeue --queue pipeline --worker embedder-01 --type embed_document
```

Other job types remain pending and are claimed by workers without a `--type` filter, or by workers filtering for different types.

---

## Priority and Queue Design

### Priority levels

Every job is `high`, `normal` (default), or `low`. Within a queue, workers always dequeue `high` before `normal`, `normal` before `low`, and FIFO within each level.

```bash
# Urgent: re-index a document immediately
queueprim enqueue --queue search/index --payload '{"doc":"d_xyz"}' --priority high

# Background: rebuild full corpus
queueprim enqueue --queue search/index --payload '{"corpus":"all"}' --priority low
```

Use priority when jobs share a worker pool and some matter more than others. A single queue with three priority levels is sufficient for most cases.

### When to use separate queues

Use a separate queue when:
- Workers are specialized and should not touch certain job types
- You need independent backpressure visibility (e.g., `stats` per queue)
- Jobs need different retry policies by design (not possible within one queue)
- You want to `purge` a category independently

Queue names are arbitrary strings. Slashes are allowed and useful for namespacing:

```bash
# Two isolated queues for two tenants
queueprim enqueue --queue tenant/acme --payload '...'
queueprim enqueue --queue tenant/globex --payload '...'

# Two worker pools for two job categories
queueprim enqueue --queue pipeline/ingest --payload '...'
queueprim enqueue --queue pipeline/transform --payload '...'
```

`queues` lists all queues with counts:

```bash
queueprim queues
```

---

## Retries and Dead-Letter

### How retries work

Every job has `max_retries` (default 0) and `attempt_count`. When a worker calls `fail`:

- If `attempt_count < max_retries`: status returns to `pending`, `attempt_count` increments. The job will be claimed again.
- If `attempt_count >= max_retries`: status becomes `dead`. The job is no longer dequeued.

```bash
# This job will be attempted up to 4 times (1 original + 3 retries) before going dead
queueprim enqueue --queue jobs/send-email \
  --payload '{"to":"user@example.com"}' \
  --max-retries 3
```

### Choosing max-retries

| Job type | Suggested max-retries | Reasoning |
|---|---|---|
| Idempotent, transient failure expected | 3–5 | Network blips, rate limits, cold starts |
| Non-idempotent, side effects | 0 or 1 | Duplicate execution causes real harm |
| Fire-and-forget, best-effort | 0 | Failure is acceptable, don't retry |
| Critical, must eventually complete | 10+ | Budget retries liberally; monitor dead queue |

### Designing idempotent payloads

If a job may retry, it must be safe to execute multiple times. The payload should contain enough information to deduplicate or make execution idempotent:

```jsonc
// Fragile: retrying may double-charge
{"amount": 49.00, "action": "charge"}

// Safe: idempotency key scopes the operation
{"amount": 49.00, "action": "charge", "idempotency_key": "order_7821_charge_v1"}
```

Include entity IDs, version numbers, or explicit idempotency keys. Avoid payloads that describe relative operations ("increment counter by 1") — prefer absolute targets ("set counter to 42").

### Forcing dead-letter

When you know a job is unrecoverable regardless of remaining retries:

```bash
queueprim fail q_abc123 --dead --reason "payload schema invalid, no point retrying"
```

### Inspecting and clearing dead-letter jobs

```bash
# See all dead jobs
queueprim list --status dead

# Inspect a specific one
queueprim get q_abc123

# Discard dead jobs older than 7 days
queueprim purge --status dead --older-than 7d
```

---

## Delayed Jobs

Jobs can be scheduled to become visible in the future:

```bash
# Retry check in 1 hour
queueprim enqueue --queue notifications --payload '{"type":"reminder"}' --delay 1h
```

The job is created in `pending` status but will not be returned by `dequeue` until the delay has elapsed. Use this for rate-limiting retries, scheduling future work, or implementing backoff between coordinated steps.

---

## Monitoring

```bash
# Global counts across all queues
queueprim stats

# Per-queue breakdown
queueprim queues
```

Signs a queue is unhealthy:
- **Rising `pending` count with no change in `done`**: workers are down or too slow
- **Rising `claimed` count**: workers are claiming but not completing — check for crashes or timeout issues
- **Rising `dead` count**: jobs are exhausting retries — inspect payloads and failure reasons with `list --status dead`
- **`pending` count stuck with `claimed` > 0**: workers may be hung; check `visible_after` on claimed jobs — if it's in the past, the sweeper hasn't run yet or something is wrong

```bash
# View claimed jobs with timestamps to spot stalled workers
queueprim list --status claimed
```
