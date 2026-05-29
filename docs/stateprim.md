# stateprim Guide

stateprim is a persistent state primitive: namespaced, JSON-valued records in a single SQLite file. One simple data model — a `(namespace, key) → value` record — serves three distinct access patterns: mutable key-value state, dedup/idempotency checks, and append-only logs. Choosing the right pattern for each piece of state is what this guide is about.

This guide covers the namespace model, the three access patterns and when to reach for each, how to query records, and housekeeping. For installation and quick start, see the [README](../README.md). For every command and flag, see the [agent reference](agent-reference.md).

---

## The Namespace + Key Model

Every record lives at a `(namespace, key)` address and holds a JSON value:

```bash
stateprim set config app.theme '"dark"'
```

- **Namespace** is a partition — group related state under one name (`config`, `sent-emails`, `audit`). `stateprim namespaces` lists them.
- **Key** identifies a record within its namespace. For key-value and dedup patterns you choose the key; for append records stateprim generates one.
- **Value** is any JSON. Note the quoting above: `'"dark"'` is the JSON string `"dark"`. A bare `dark` is not valid JSON — wrap strings in quotes, and remember the shell needs its own quoting around the JSON.

Records carry `created_at` / `updated_at` timestamps and an `immutable` flag, both managed by stateprim.

---

## Pattern 1: Key-Value State

The default pattern: store a current value, read it back, overwrite it when it changes.

```bash
stateprim set    config app.theme '"dark"'
stateprim get    config app.theme --format json
stateprim delete config app.theme
```

`set` is an **upsert** — writing an existing key overwrites its value and bumps `updated_at`. This makes `set` idempotent: applying the same write twice leaves the same result. Use this pattern for configuration, cursors, last-run timestamps, feature flags — any value representing "the current state of X".

If a value should never change once written, mark it immutable:

```bash
stateprim set config schema.version '3' --immutable
```

A later `set` to an immutable key is refused, protecting write-once facts from accidental overwrite.

---

## Pattern 2: Dedup / Idempotency

"Have I already done this?" comes up constantly in agent work — did I send this email, process this webhook, handle this message? stateprim answers it with two commands.

`has` is a pure existence check, returning `yes`/`no` (or `{"exists": true}` in JSON):

```bash
if [ "$(stateprim has sent-emails msg:abc123)" = "no" ]; then
  send_email
  stateprim set sent-emails msg:abc123 '{"sent_at":"2026-03-04T10:00:00Z"}'
fi
```

That check-then-set has a race if two workers run it at once. `set-if-new` collapses it into one atomic operation: it writes only if the key is absent, and is a no-op (returning the existing record) if the key already exists:

```bash
# Atomic claim — only one worker wins
stateprim set-if-new jobs job:789 '{"claimed_by":"worker-01"}'
```

Reach for `set-if-new` whenever the existence check and the write must not be interleaved by another process — it is the safe idempotency primitive.

---

## Pattern 3: Append-Only Logs

When you need a *history* rather than a current value, use `append`. Each call writes a new immutable, timestamped record under an auto-generated key:

```bash
stateprim append audit '{"action":"deploy","version":"1.2.3"}'
stateprim append audit '{"action":"rollback","version":"1.2.2"}'
```

Append records are never overwritten — they accumulate. This is the right pattern for audit trails, event streams, and any "what happened, in order" data. Because the key is generated, you don't address append records individually; you read them back with `query`.

---

## Choosing a Pattern

| You want to… | Pattern | Commands |
|---|---|---|
| Track the current value of something | Key-value | `set` / `get` / `delete` |
| Avoid doing the same thing twice | Dedup | `has` / `set-if-new` |
| Record an ordered history of events | Append log | `append` / `query` |

The patterns share one model, so a single namespace *could* mix them — but don't. Keep a namespace to one pattern; it keeps queries and mental models clean.

---

## Querying

`query` reads multiple records from a namespace and is where append logs and bulk key-value reads pay off:

```bash
# Everything in a namespace
stateprim query audit --format json

# Only recently-updated records
stateprim query audit --since 24h --format json

# Only keys under a prefix
stateprim query config --prefix app. --format json

# Just the count
stateprim query sent-emails --count
```

`--since` takes a duration (`24h`, `7d`) and filters by recency. `--prefix` filters by key prefix — which is why hierarchical keys like `app.theme`, `app.locale` are worth adopting. `--count` returns the number of matching records instead of the records themselves, useful for cheap size checks.

---

## Housekeeping

Append logs and dedup namespaces grow without bound unless you trim them. `purge` permanently removes records in a namespace older than a duration:

```bash
# Drop audit entries older than 30 days
stateprim purge audit 30d
```

`purge` is destructive and not idempotent — it deletes whatever currently matches. Reserve it for data with a genuine retention window (logs, expired dedup markers), not for current-state key-value records you still depend on.

For a quick overview:

```bash
stateprim namespaces           # list all namespaces
stateprim stats --format json  # { "total_records": 156, "total_namespaces": 8 }
```
