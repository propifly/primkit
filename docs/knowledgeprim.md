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
| `instruction` | Human directive, preference, or rule the agent should follow |
| `feedback` | Human correction or validation of agent behavior |
| `goal` | Human objective or desired outcome |

These are suggestions, not rules. Use them as a starting point and let your vocabulary evolve.

### Human vs Agent Knowledge

The `--source` flag tracks who captured an entity. This matters because human knowledge carries authority that agent knowledge doesn't. A human instruction like "never use float for money" is a rule. An agent observation like "this function returns a float" is data.

Use `--source` consistently to distinguish them:

```bash
# Human captures a rule
knowledgeprim capture --type instruction --title "Never use float for money" \
  --body "Always use decimal or integer cents. Float arithmetic causes rounding errors in financial calculations." \
  --source andres

# Agent captures an observation
knowledgeprim capture --type observation --title "PaymentService uses float64 for amounts" \
  --body "Found in payment/service.go line 42. Returns float64 for transaction amounts." \
  --source coding-agent
```

When an agent searches before acting, human-sourced entities provide authoritative guidance. Agent-sourced entities provide context. Both are valuable, but they serve different roles in the graph.

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

A coding agent searches knowledge before writing code, captures what it learns, and strengthens connections it relies on.

**Before writing code — consult the graph:**

```bash
# Agent is about to implement error handling for an HTTP client.
# Step 1: Search for existing patterns and decisions
knowledgeprim search "error handling HTTP" --type pattern --format json
knowledgeprim search "error handling HTTP" --type decision --format json
knowledgeprim search "error handling HTTP" --type instruction --format json

# Step 2: Found a pattern (e_ret789). Traverse to see what it connects to.
knowledgeprim related e_ret789 --depth 2

# Step 3: Found a human instruction (e_ins012) via traversal:
#   "Always use circuit breaker for external APIs" (--source andres, weight 4.0)
# Agent uses this to inform its implementation.
```

**After writing code — capture what was learned:**

```bash
# Step 4: Agent discovered a new pattern during implementation
knowledgeprim capture --type pattern \
  --title "Retry with exponential backoff and jitter" \
  --body "Retry failed HTTP calls with 2^n * (1 + rand) backoff, capped at 30s. Jitter prevents thundering herd." \
  --source coding-agent

# Step 5: Connect to the decision it extends, with context
knowledgeprim connect e_new456 e_dec789 --relationship extends \
  --context "Circuit breaker wraps the retry pattern — retries handle transient errors, circuit breaker handles persistent outages"

# Step 6: Strengthen the edge to the instruction it followed
knowledgeprim strengthen e_new456 e_ins012 supports
```

**Periodic maintenance:**

```bash
# Find orphaned patterns that aren't connected to anything
knowledgeprim discover --orphans
```

The key loop: **search -> read -> act -> capture -> connect -> strengthen**. Every session, the agent starts with more knowledge than the last.

### Research Agent

A research agent builds evidence maps and uses them to synthesize answers.

**Capturing new evidence:**

```bash
# 1. Capture an article
knowledgeprim capture --type article \
  --title "On-Device LLMs: Performance Benchmarks 2025" \
  --url "https://example.com/on-device-llms" \
  --body "Benchmarks show 40% accuracy loss vs cloud inference on complex reasoning tasks." \
  --source research-agent

# 2. Search for related evidence to connect it
knowledgeprim search "on-device inference accuracy" --type observation

# 3. Connect with detailed context explaining the contradiction
knowledgeprim connect e_article1 e_obs123 --relationship contradicts \
  --context "The 2025 benchmarks show 40% accuracy loss on complex reasoning, directly contradicting the 2024 observation claiming near-parity for most tasks"
```

**Using the graph to answer a research question:**

```bash
# Human asks: "What's the current state of on-device LLM viability?"

# 1. Search for all evidence
knowledgeprim search "on-device LLM" --format json

# 2. From the top result, traverse to find supporting and contradicting evidence
knowledgeprim related e_article1 --depth 2 --relationship contradicts
knowledgeprim related e_article1 --depth 2 --relationship supports

# 3. Check for clusters — are there separate evidence groups?
knowledgeprim discover --clusters

# 4. Find bridge concepts connecting different research areas
knowledgeprim discover --bridges

# Agent now has: articles, observations, contradictions, supporting evidence,
# and the structure of how they relate. It synthesizes from this, not from
# a single search result.
```

### Support Agent

A support agent builds a solution graph and searches it before responding to tickets.

**When a new ticket arrives — search before responding:**

```bash
# 1. Ticket: "Getting 504 errors intermittently"
# Search for similar issues and known solutions
knowledgeprim search "504 timeout" --format json
knowledgeprim search "504 timeout" --type pattern --format json

# 2. Found a matching bug entity (e_bug1). Traverse to find solutions.
knowledgeprim related e_bug1 --relationship applies_to --direction incoming

# 3. Found "Dynamic connection pool sizing" (e_pat1, weight 3.0).
# Agent responds with this solution, citing the pattern.
# Strengthen the edge — this solution proved useful again.
knowledgeprim strengthen e_pat1 e_bug1 applies_to
```

**When the issue is new — capture it:**

```bash
# 4. No matching solution found. Capture the new issue.
knowledgeprim capture --type bug \
  --title "Memory leak in WebSocket handler after 10k connections" \
  --body "RSS grows linearly. Goroutine count stable. Suspect buffer pool not releasing." \
  --source support-agent

# 5. After resolution, capture the solution and connect it
knowledgeprim capture --type pattern \
  --title "Bounded buffer pool with explicit release" \
  --body "Use sync.Pool with a max size. Call Release() in defer, not just on success path." \
  --source support-agent

knowledgeprim connect e_pat2 e_bug2 --relationship applies_to \
  --context "Bounded pool with explicit release fixes the RSS growth by ensuring buffers return to the pool on all exit paths"

# 6. Spot trends over time
knowledgeprim discover --temporal
```

## Training an Agent

knowledgeprim is persistent memory that survives across sessions. Without it, an agent starts fresh every time. With it, the agent starts every session with accumulated knowledge. "Training" an agent means deliberately building up a graph that changes the agent's behavior — because the agent searches before acting and finds your guidance instead of guessing.

### Seed Phase: Human Knowledge First

The most valuable knowledge in the graph comes from humans. An agent can observe and accumulate, but humans provide the authority: rules, preferences, domain concepts, and strategic decisions.

Start by seeding the graph with what you know:

```bash
# Domain rules — things the agent must always follow
knowledgeprim capture --type instruction \
  --title "Never use float for money" \
  --body "Always use decimal types or integer cents. Float arithmetic causes rounding errors in financial calculations. This applies to all payment, billing, and accounting code." \
  --source andres

# Domain concepts — define terms in YOUR context
knowledgeprim capture --type concept \
  --title "Primitive (primkit)" \
  --body "A self-contained infrastructure component shipped as a single binary. Each primitive owns one domain (tasks, state, knowledge) and exposes CLI, HTTP, and MCP interfaces from the same binary." \
  --source andres

# Strategic decisions — capture the reasoning, not just the choice
knowledgeprim capture --type decision \
  --title "Pure Go SQLite over CGo" \
  --body "Use modernc.org/sqlite instead of mattn/go-sqlite3. Trades ~10% performance for zero CGo dependency, which simplifies ARM cross-compilation for Raspberry Pi deployments." \
  --source andres

# Connect related decisions with context
knowledgeprim connect e_dec1 e_dec2 --relationship supports \
  --context "The pure Go decision enables the single-binary goal — no shared libraries to ship"
```

### Correction Loops

When an agent does something wrong, don't just fix it — capture the correction. The next time the agent encounters a similar situation, it finds the correction instead of repeating the mistake.

```bash
# Agent used sync.Mutex where sync.RWMutex was better.
# Capture the correction.
knowledgeprim capture --type feedback \
  --title "Use RWMutex for read-heavy shared state" \
  --body "Agent used sync.Mutex in the cache layer. This serializes all access including reads. Use sync.RWMutex when reads vastly outnumber writes — allows concurrent reads." \
  --source andres

# Connect it to the original decision if it exists
knowledgeprim connect e_feedback1 e_dec_cache --relationship contradicts \
  --context "The Mutex choice was wrong for this use case — read-heavy access patterns need RWMutex for acceptable performance"
```

Over time, the graph accumulates corrections. An agent that searches before writing concurrency code finds both the patterns AND the corrections — avoiding past mistakes.

### Building Domain Expertise

To train an agent on a specific domain, build up interconnected knowledge:

```bash
# 1. Capture foundational concepts
knowledgeprim capture --type concept --title "WAL mode" \
  --body "Write-Ahead Logging. SQLite writes changes to a separate WAL file before the main database. Enables concurrent readers during writes." \
  --source andres

knowledgeprim capture --type concept --title "Busy timeout" \
  --body "SQLite waits this many milliseconds for a lock before returning SQLITE_BUSY. Set to 5000ms minimum for concurrent access patterns." \
  --source andres

# 2. Connect them with rich context
knowledgeprim connect e_wal e_busy --relationship relates_to \
  --context "WAL mode reduces lock contention but doesn't eliminate it — busy timeout is still needed for write-write conflicts"

# 3. Add practical patterns
knowledgeprim capture --type pattern --title "SQLite connection setup" \
  --body "Open with WAL mode, foreign keys, busy timeout 5s. Single writer, multiple readers. Use connection pooling with max 1 writer connection." \
  --source andres

# 4. Connect pattern to concepts
knowledgeprim connect e_pattern1 e_wal --relationship applies_to \
  --context "The setup pattern implements WAL mode as part of the standard connection initialization"

# 5. Add edge cases and gotchas
knowledgeprim capture --type observation --title "WAL checkpoint stalls under heavy write load" \
  --body "Observed 200ms stalls during WAL checkpoint when write rate exceeds 1000 tx/sec. Mitigation: manual checkpointing during idle periods." \
  --source andres
```

As the domain graph grows denser, the agent's ability to navigate it improves. Searching "SQLite setup" returns the pattern. Traversing from the pattern reveals WAL mode, busy timeout, and the checkpoint gotcha. The agent writes better code because it has contextual domain knowledge, not just a pattern in isolation.

### The Source Authority Pattern

Use `--source` consistently. When the agent searches and finds multiple results, source helps it prioritize:

- `--source andres` (human) — authoritative rules, corrections, strategic decisions
- `--source coding-agent` — observations, accumulated patterns, auto-discovered connections
- `--source auto` — auto-connect edges (lowest authority, useful for discovery)

The agent doesn't need special logic for this. Human-sourced entities naturally accumulate more strengthened edges (because humans create targeted, high-value connections), which means `--min-weight` filtering naturally surfaces human knowledge first.

### What Good Training Looks Like

A well-trained agent's graph has:

- **Human instructions** at the center, heavily connected to patterns and decisions
- **Agent observations** at the edges, supporting or questioning the human knowledge
- **Strong weight** on connections the agent has relied on repeatedly
- **Corrections** connected to the mistakes they fix, so the agent finds both together
- **Domain concepts** forming a backbone that agent-discovered patterns attach to

The graph is not static. Every session, the agent adds observations, strengthens useful connections, and occasionally surfaces contradictions for the human to resolve. Training is ongoing, not one-time.

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
