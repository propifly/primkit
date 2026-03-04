# knowledgeprim Guide

knowledgeprim is a knowledge graph primitive. It stores typed entities, weighted contextualized edges, and optional vector embeddings in a single SQLite file. Agents and humans use it to capture knowledge, connect ideas, and discover patterns that emerge from usage.

This guide teaches effective usage patterns: what to capture, how to connect, when to use each search mode, and how discovery improves graph quality over time. For installation and quick start, see the [README](../README.md). For internals, see [architecture](architecture.md). For config fields, see [configuration](configuration.md).

## Entity Types

Entities are the nodes of your knowledge graph. Every entity has a **type** — a freeform string that categorizes what kind of thing it is. Types are not enforced by a schema; you define your own vocabulary.

### Starter Vocabulary

| Type | When to use |
|------|-------------|
| `article` | External content: blog posts, papers, docs, web pages |
| `thought` | Agent-generated insight, observation, or hypothesis |
| `concept` | Abstract idea or term worth defining |
| `pattern` | Recurring technique, approach, or anti-pattern |
| `observation` | Empirical data point, measurement, or finding |
| `decision` | Choice made with reasoning (architectural, product, process) |
| `bug` | Defect with reproduction steps or workaround |

These are suggestions, not rules. Use them as a starting point and let your vocabulary evolve.

### Defining Your Own Types

Use `knowledgeprim types` to see your current vocabulary and counts. A good type vocabulary has a few broad types with many entities each — not many narrow types with a few entities each.

Guidelines:

- Keep types broad enough to have 5+ entities each. If a type has fewer, merge it into a broader one.
- Types answer "what kind of thing is this?" — not "where did it come from." Use `--source` for provenance and `--properties` for metadata.
- Let types emerge from usage. Start with the starter vocabulary and add new types only when existing ones genuinely don't fit.

```bash
# Capture with a standard type
knowledgeprim capture --type pattern --title "Retry with exponential backoff" \
  --body "Retry failed HTTP calls with 2^n backoff capped at 30s." --source agent

# Capture with a custom type
knowledgeprim capture --type runbook --title "Redis failover procedure" \
  --body "1. Verify sentinel status..." --source ops-agent

# Audit your type vocabulary
knowledgeprim types
```

### Anti-Pattern: Over-Typing

If you have 20 types with 2 entities each, you have a taxonomy problem, not a knowledge graph. Merge types. A graph with 5 types and 200 entities is far more useful than one with 40 types and 200 entities.

## Relationships

Edges connect entities. Every edge has a **relationship** type — a freeform string describing how two entities relate. Like entity types, relationships are not enforced; you define your own.

### Starter Relationships

| Relationship | Reads as | When to use |
|---|---|---|
| `relates_to` | A relates to B | General connection, no specific semantic implied |
| `extends` | A extends B | A builds on, elaborates, or deepens B |
| `contradicts` | A contradicts B | A presents opposing evidence or conclusions to B |
| `supports` | A supports B | A provides evidence or validation for B |
| `inspired_by` | A was inspired by B | A was created because of B |
| `applies_to` | A applies to B | A (pattern, solution) is applicable to B (problem, bug) |
| `similar_to` | A is similar to B | Semantic similarity (auto-connect creates these) |
| `questions` | A questions B | A raises doubts or challenges an assumption in B |

### Directionality

Edges are directed: source to target. The relationship reads "source [relationship] target." For `contradicts`, the source is the entity doing the contradicting. For `extends`, the source is the entity that extends the target.

Some relationships are effectively symmetric in meaning (`relates_to`, `similar_to`), but edges are still stored as directed. Use `--direction both` in traversal to follow edges in both directions.

```bash
# Directional: this article extends that concept
knowledgeprim connect e_a1b2c3 e_d4e5f6 --relationship extends \
  --context "Adds benchmarking data to the theoretical framework"

# Check what relationships exist in your graph
knowledgeprim relationships
```

## Edge Context

Edge context is the most important concept in knowledgeprim. It's a prose explanation of **why** a connection exists — not just that it exists.

### Why Context Matters

Context transforms your graph from a structural skeleton into a queryable knowledge structure. Edge context is indexed by FTS5 and embedded by vectors, which means:

- Searching can surface *connections*, not just entities
- An agent traversing the graph can read context to understand why two ideas relate
- Discovery can find edges with missing context so you can fix them

Without context, an edge just says "A relates to B." With context, it says "A and B both address rate limiting but reach opposite conclusions — A advocates aggressive limiting while B suggests adaptive backoff."

### Writing Good Context

| Bad context | Good context |
|---|---|
| `"related"` | `"Both discuss rate limiting but from opposite angles — this article advocates aggressive limiting while the pattern suggests adaptive backoff"` |
| *(empty)* | `"This observation about memory leaks in long-running Go processes directly contradicts the claim that garbage collection handles everything"` |
| `"same topic"` | `"Applies the retry-with-backoff pattern to the specific case of Redis connection failures during failover"` |

Good context answers the question: **"If someone asks me why these two things are connected, what would I say?"**

### Improving Context Over Time

Use `edge-edit` to update context on existing edges:

```bash
# Add context to an edge that was created without it
knowledgeprim edge-edit e_a1b2c3 e_d4e5f6 contradicts \
  --context "The 2024 paper shows 40% accuracy loss that directly refutes the 2023 claim of parity"
```

### Discovering Weak Edges

Weak edges are edges with no context. They represent connections that exist but haven't been explained. Find them with discovery:

```bash
knowledgeprim discover --weak-edges
```

Fixing weak edges is one of the highest-leverage quality improvements you can make to a knowledge graph.

## Search

knowledgeprim supports three search modes. All return ranked results with scores.

### Search Modes

| Mode | Flag | Requires Embedding | Best for |
|------|------|-------------------|----------|
| Hybrid | `--mode hybrid` (default) | No | Most queries — combines keyword precision with semantic recall |
| FTS | `--mode fts` | No | Known terms, exact phrases, technical identifiers |
| Vector | `--mode vector` | Yes | Conceptual similarity when vocabulary differs |

### Hybrid Fallback

When no embedding provider is configured, `--mode hybrid` still works — it runs FTS-only. When embedding IS configured, it combines FTS + vector results via Reciprocal Rank Fusion (k=60). This means agents can always default to `--mode hybrid` without checking whether embedding is configured.

### Search Tips

- Filter by type: `--type pattern` narrows results to pattern entities only
- Control result count: `--limit 5` (default is 20)
- FTS supports SQLite FTS5 query syntax: `AND`, `OR`, `NOT`, phrase matching with `"quotes"`
- Vector search finds conceptually similar content even with different words ("graceful degradation" matches content about "fallback strategies")
- Use `--format json` when piping results to another tool or agent

```bash
# Default hybrid search
knowledgeprim search "distributed consensus"

# FTS for exact technical terms
knowledgeprim search "Litestream WAL" --mode fts

# Vector search for conceptual similarity
knowledgeprim search "making systems resilient to failure" --mode vector

# Filtered by type
knowledgeprim search "retry" --type pattern --limit 5

# JSON output for programmatic use
knowledgeprim search "edge computing" --format json
```

## Traversal

Graph traversal walks from a starting entity through connected edges. Use `related` to explore what an entity connects to.

### Basic Traversal

```bash
# Show everything connected to an entity (1 hop)
knowledgeprim related e_a1b2c3
```

Output includes depth, direction, relationship, weight, ID, and title for each connected entity.

### Traversal Options

| Flag | Default | Description |
|------|---------|-------------|
| `--depth` | `1` | How many hops to traverse |
| `--direction` | `both` | `outgoing`, `incoming`, or `both` |
| `--relationship` | *(all)* | Filter to a specific relationship type |
| `--min-weight` | `0` | Only follow edges at or above this weight |

### Weight-Filtered Traversal

Combining `--min-weight` with strengthened edges surfaces high-confidence paths. If you've been strengthening edges as you reference them, weight becomes a signal for which connections matter most.

```bash
# Follow only well-evidenced connections (weight >= 3.0)
knowledgeprim related e_a1b2c3 --depth 3 --min-weight 3.0

# Trace contradictions from a specific entity
knowledgeprim related e_a1b2c3 --relationship contradicts --direction outgoing

# Deep traversal through all connected knowledge
knowledgeprim related e_a1b2c3 --depth 5
```

## Discovery

Discovery analyzes graph structure to surface patterns that aren't obvious from individual entities or edges. It's the feedback loop that tells you where your graph needs work.

### Discovery Operations

| Operation | Flag | What it finds | Action to take |
|-----------|------|---------------|----------------|
| Orphans | `--orphans` | Entities with no edges | Connect them or question if they belong |
| Clusters | `--clusters` | Densely connected groups | Name the domain, look for cross-cluster bridges |
| Bridges | `--bridges` | High-degree entities connecting clusters | Ensure they have good titles and bodies |
| Temporal | `--temporal` | Entity type distribution over time | Spot knowledge growth trends and gaps |
| Weak Edges | `--weak-edges` | Edges with no context prose | Add context to make them searchable |

### The Quality Loop

Run this periodically to improve your graph:

1. Run `knowledgeprim discover` for a full report
2. Process orphans: connect them to related entities, or delete if no longer relevant
3. Review weak edges: add context explaining why each connection exists
4. Examine clusters: do they make sense as knowledge domains? Are there missing cross-cluster connections?
5. Check bridges: are your bridge entities well-documented with clear titles and bodies?

```bash
# Full discovery report
knowledgeprim discover

# Just orphans (quick quality check)
knowledgeprim discover --orphans

# Weak edges in JSON for programmatic processing
knowledgeprim discover --weak-edges --format json
```

## Auto-Connect and Edge Weight

### Auto-Connect

When an embedding provider is configured, `capture` automatically searches for similar existing entities and creates `similar_to` edges. This seeds your graph with initial connections.

How it works:

1. The new entity's title and body are embedded
2. Cosine distance search against all existing embeddings
3. Entities below the threshold get automatic `similar_to` edges
4. Each auto-created edge includes context like `"auto-connected: cosine distance 0.28"`

Configuration (in `config.yaml`):

```yaml
auto_connect:
  enabled: true
  threshold: 0.35    # cosine distance (lower = more similar)
  max_connections: 10
```

Per-capture overrides:

```bash
# Skip auto-connect for this capture
knowledgeprim capture --type thought --title "..." --no-auto-connect --source agent

# Use a tighter threshold (fewer, higher-quality connections)
knowledgeprim capture --type article --title "..." --threshold 0.25 --source agent
```

### Tuning Auto-Connect

- **Threshold 0.35** (default) is conservative — produces moderate connections
- **Lower (0.25)**: fewer connections, higher quality. Use when your graph is large and you want precision.
- **Higher (0.50)**: more connections, more noise. Use when your graph is small and you want seeding.
- Monitor auto-connect quality: `knowledgeprim discover --weak-edges` shows auto-connect edges (they have context starting with "auto-connected:")
- If auto-connect produces too many spurious connections, lower the threshold

### Edge Weight

Edges start at weight 1.0. The `strengthen` command increments weight by 1.0 each time it's called. Weight represents accumulated evidence that a connection matters.

Use case: every time an agent references or relies on a connection, strengthen it. Over time, heavily-used connections naturally rise to the top.

```bash
# Strengthen an edge (weight goes from 1.0 to 2.0)
knowledgeprim strengthen e_a1b2c3 e_d4e5f6 relates_to

# After multiple strengthens, filter traversal by weight
knowledgeprim related e_a1b2c3 --min-weight 3.0
```

## Agent Workflows

These examples show end-to-end workflows for different agent types using actual CLI syntax.

### Coding Agent

A coding agent encounters patterns, makes decisions, and tracks bugs during development.

```bash
# 1. Capture a pattern discovered during code review
knowledgeprim capture --type pattern \
  --title "Retry with exponential backoff and jitter" \
  --body "Retry failed HTTP calls with 2^n * (1 + rand) backoff, capped at 30s. Jitter prevents thundering herd." \
  --source coding-agent

# 2. Auto-connect finds a related decision entity (e_prev123)
# Output: "Auto-connected to 2 similar entities."

# 3. Capture an architectural decision
knowledgeprim capture --type decision \
  --title "Use circuit breaker for external API calls" \
  --body "After three consecutive failures, open circuit for 60s before retrying." \
  --source coding-agent

# 4. Manually connect the decision to the pattern with context
knowledgeprim connect e_new456 e_prev789 --relationship extends \
  --context "Circuit breaker wraps the retry pattern — retries handle transient errors, circuit breaker handles persistent outages"

# 5. When the pattern comes up again, strengthen the edge
knowledgeprim strengthen e_new456 e_prev789 extends

# 6. Search for existing patterns when writing new code
knowledgeprim search "error handling HTTP" --type pattern

# 7. Periodic quality check
knowledgeprim discover --orphans
```

### Research Agent

A research agent captures articles, tracks evidence, and discovers contradictions.

```bash
# 1. Capture an article
knowledgeprim capture --type article \
  --title "On-Device LLMs: Performance Benchmarks 2025" \
  --url "https://example.com/on-device-llms" \
  --body "Benchmarks show 40% accuracy loss vs cloud inference on complex reasoning tasks." \
  --source research-agent

# 2. Find it contradicts an earlier observation
knowledgeprim search "on-device inference accuracy" --type observation

# 3. Connect with detailed context
knowledgeprim connect e_article1 e_obs123 --relationship contradicts \
  --context "The 2025 benchmarks show 40% accuracy loss on complex reasoning, directly contradicting the 2024 observation claiming near-parity for most tasks"

# 4. Discover clusters of evidence on a topic
knowledgeprim discover --clusters

# 5. Find bridge entities connecting different research areas
knowledgeprim discover --bridges

# 6. Deep traversal from a key concept
knowledgeprim related e_concept1 --depth 3 --min-weight 2.0
```

### Support Agent

A support agent captures issues, maps solutions, and spots trends.

```bash
# 1. Capture an issue and its solution
knowledgeprim capture --type bug \
  --title "Connection timeout during peak hours" \
  --body "Users report 504 errors between 2-4pm UTC. Root cause: connection pool exhaustion." \
  --source support-agent

knowledgeprim capture --type pattern \
  --title "Dynamic connection pool sizing" \
  --body "Scale pool size based on concurrent request count with high-water mark." \
  --source support-agent

# 2. Connect solution to problem
knowledgeprim connect e_pattern1 e_bug1 --relationship applies_to \
  --context "Dynamic pool sizing directly addresses the pool exhaustion that causes the 504 errors during peak"

# 3. When a similar issue arrives, search for existing solutions
knowledgeprim search "connection pool timeout" --type pattern

# 4. Spot incident trends over time
knowledgeprim discover --temporal
```

## Growing Your Graph

### Maturity Stages

**Seeding (0-50 entities):** Capture broadly. Let auto-connect create initial links. Don't worry about type consistency yet — you're exploring what kinds of knowledge this graph will hold.

**Connecting (50-200 entities):** Start creating manual connections with rich context. Run discovery regularly to find orphans. Your type vocabulary stabilizes during this phase.

**Strengthening (200+ entities):** Weight important edges by strengthening them when referenced. Use clusters and bridges to understand your knowledge domains. Temporal analysis shows growth trends.

**Pruning (ongoing):** Delete entities that no longer serve a purpose. Fix weak edges. Merge overlapping concepts. A curated graph is more valuable than a large one.

### Anti-Patterns

| Anti-Pattern | Symptom | Fix |
|---|---|---|
| Over-typing | 20 types with 2 entities each | Merge into broader types, let them emerge from usage |
| Under-connecting | High orphan count in discovery | Connect after every capture, or enable auto-connect |
| Empty context | Many weak edges in discovery | Write why the connection exists, not just that it exists |
| Premature structure | Designing a taxonomy before you have 50 entities | Capture first, structure later |
| Ignoring weight | All edges at weight 1.0 | Strengthen edges when you reference or rely on them |

### The Core Principle

**Primitive is dumb, agent is smart.** knowledgeprim stores, indexes, and retrieves. It doesn't decide what's worth capturing, how to type entities, or what edges mean. That's the agent's job. No taxonomy is enforced. No relationship types are required. Structure emerges from usage, not from upfront design.

Start capturing. Connect as you go. Let discovery show you what your knowledge looks like.
