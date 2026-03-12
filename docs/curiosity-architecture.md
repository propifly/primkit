# Building Curious Agents with PrimKit: A Knowledge Graph Approach to Self-Improving AI Agents

*Based on building Lukas, a strategic advisor agent that investigates, learns, and builds knowledge autonomously across sessions.*

---

## TL;DR

Stateless agents start every session from zero. PrimKit's knowledgeprim (knowledge graph) and stateprim (key-value state) give agents persistent memory that survives across sessions — but persistence alone doesn't produce curiosity. Curiosity requires three coordinated layers: **identity** (the agent believes investigation is who it is), **workflow** (dedicated time for exploration, not just reactive processing), and **state** (themes that track open investigations across sessions). The result is a flywheel: exploration deepens themes, observations accumulate into patterns, patterns generate recommendations, and new findings spawn new themes. Human involvement is needed for seeding initial knowledge, validating recommendations before they become actions, and pruning themes that drift from the agent's domain — but not for the investigation itself.

---

## The Problem: Agents That Only React

Most agent systems are reactive. A cron fires, the agent scans for changes, processes what it finds, and reports. Next session: same thing, starting from zero context.

This works for operational tasks (scan email, triage tasks). It fails for strategic work — the kind where value comes from noticing patterns across multiple observations over time, connecting a finding from Tuesday to something that happened last week, or following a thread through several sessions until it resolves into something worth saying.

The gap isn't intelligence. The model is capable of deep analysis within a single session. The gap is **continuity** — there's no mechanism for carrying an investigation forward, and no incentive to start one.

## What PrimKit Provides

PrimKit is a suite of SQLite-backed primitives for agent state. Two are relevant here:

**knowledgeprim** — A knowledge graph with typed entities, weighted edges, and hybrid search (FTS5 + vector embeddings). Entities have types (observation, pattern, article, instruction, recommendation), edges have relationships (supports, contradicts, extends, applies_to) with prose context explaining *why* the connection exists. Auto-connect creates `similar_to` edges when new entities are semantically close to existing ones.

**stateprim** — A namespaced key-value store with event logging. Keys are mutable, namespaces provide isolation. No embeddings — it's for operational state, not knowledge.

Together: knowledgeprim is what the agent knows, stateprim is what the agent is doing about it.

### Why a Graph, Not a Document Store

A document store (or flat memory files) gives you retrieval: "find things related to X." A graph gives you traversal: "what connects to this observation, and what connects to *those* things?" Traversal is what produces non-obvious connections — the kind where an observation about client revenue leakage links through a pattern about tribal knowledge to a recommendation about discovery phase methodology. The agent didn't search for that chain. It followed edges.

Auto-connect (cosine distance on capture) seeds the graph with structural connections the agent didn't explicitly create. In the Lukas deployment, we used threshold 0.28 (cosine distance — lower means more similar) with max 5 connections (primkit defaults are 0.35 / 10); each new entity arrives pre-linked to its most relevant neighbors. The agent's job is to upgrade these mechanical `similar_to` edges into meaningful relationships with context — or to discover that the similarity was superficial and the real connection is elsewhere.

## The Three Layers of Curiosity

Persistence is necessary but not sufficient. An agent with a knowledge graph but no curiosity is just a better filing cabinet. Three layers, coordinated, produce the behavior:

### Layer 1: Identity — Curiosity as Character Trait

Per the principle that personality shapes behavior more reliably than instructions (agents mirror the register and posture of their identity files), curiosity must be part of who the agent is, not a task it's told to perform.

**What this looks like in practice (from Lukas's SOUL.md):**

> I pull threads. When something catches my attention — a gap in the methodology, a shift in the market, a pattern across clients — I don't just note it. I follow it. I check what I already know, I look at what's changed, I search for what others are doing. Not endlessly — to a conclusion. But the curiosity comes first. If I'm not actively chewing on something, I'm coasting.

The key phrase: "If I'm not actively chewing on something, I'm coasting." This frames inactivity as failure — not because a rule says "you must explore," but because the character wouldn't let a thread go unpulled.

**Why identity, not instructions:** An instruction like "explore your knowledge graph during exploration crons" produces mechanical compliance. An identity trait like "I investigate to a conclusion" produces judgment about *what* to investigate, *how deep* to go, and *when* a thread is exhausted. The instruction tells the agent what to do. The identity tells it what kind of agent it is — and the behavior follows.

### Layer 2: Workflow — Dedicated Exploration Time

Identity creates the drive. Workflow allocates time for it.

The mistake is expecting curiosity to happen inside reactive crons. If the agent's morning job is "check what changed in the vault and report," it will scan, report, and stop. There's no room for "and also, follow up on that thing from three days ago."

The fix: a dedicated exploration cron with no output obligation. The agent's only job is to investigate.

**Exploration cron design:**

1. **Pick a theme** from stateprim. Prioritize themes with the most observations (near synthesis), themes not explored recently, or themes connected to upcoming work.
2. **Recall what you know** — search knowledgeprim for the theme and related concepts. Traverse from key entities to find adjacent knowledge.
3. **Search primary sources** — read content in the agent's domain (vault files, project docs, whatever it works with).
4. **Search external sources** — web search for developments, best practices, competitor approaches, published research.
5. **Capture findings** as typed entities (observation, article, pattern). Connect them to existing knowledge with relationship context.
6. **Look for patterns** — if a theme has 3+ observations, synthesize into a pattern entity. Check for meta-patterns across themes.
7. **Spawn new threads** — exploration often opens new questions. Create new themes for sub-investigations.
8. **Update the theme** — increment observation count, set last_explored date. If the theme is exhausted, mark it resolved.

**Critical design choice: no output to the user.** The exploration cron is silent. Insights surface through existing communication crons (morning briefings, weekly reviews) when they're ready. This prevents two failure modes: (a) half-formed findings presented prematurely, and (b) the agent optimizing for message production instead of knowledge building.

### Layer 3: State — Themes as Continuity Mechanism

Without themes, every exploration session starts cold: "What should I investigate?" With themes, the agent picks up where it left off: "I have four open investigations. The competitive landscape theme hasn't been explored in three days. Let me pull that thread."

**Theme lifecycle:**

```
Create → Explore → Mature → Resolve → Spawn
```

1. **Create** — When the agent notices something worth following (a gap, a trend, a pattern forming), it creates a theme in stateprim with a slug, status, observation count, summary, and last_explored timestamp.

2. **Explore** — During exploration crons, the agent picks a theme and investigates. After exploring, it updates the theme's metadata (increment observations, set last_explored).

3. **Mature** — When a theme accumulates 3+ observations, the agent synthesizes them into a `pattern` entity in knowledgeprim. The pattern connects the observations and represents a higher-order finding.

4. **Resolve** — When a pattern leads to a recommendation the agent surfaces to the user, or when the thread runs dry, the theme is marked `resolved`. Not deleted — it's history.

5. **Spawn** — Exploration often opens new threads. A theme about "competitive landscape" might spawn "governance tooling trends" when the agent discovers that competitors are differentiating on governance specifically.

**Themes in stateprim (real example):**

```
themes  theme:walk-phase-readiness        active  observations:1
themes  theme:stage3-prompt-gap           active  observations:1
themes  theme:organic-growth-pain-points  active  observations:3
themes  theme:competitive-landscape       active  observations:0
```

The `organic-growth-pain-points` theme has 3 observations — it's ready for pattern synthesis. The `competitive-landscape` theme has 0 — it needs its first exploration pass. This metadata drives prioritization without the agent needing to remember session-to-session what it was doing.

## The Flywheel

The three layers create a self-reinforcing cycle:

```
Exploration deepens themes
    → Observations accumulate in knowledgeprim
        → 3+ observations synthesize into patterns
            → Patterns generate recommendations
                → Recommendations surface in briefings
                    → New findings spawn new themes
                        → (back to exploration)
```

Each cron type feeds the next:

- **Exploration crons** (silent) build depth in the knowledge graph
- **Morning crons** start by checking active themes, not just scanning for changes — so the agent's daily work is informed by its ongoing investigations
- **Weekly review crons** draw on accumulated knowledge to evaluate an area substantively
- **Trend scan crons** connect external findings to active themes, and spawn new themes when a finding opens a new thread

The flywheel is self-sustaining once seeded: themes create the need for exploration, exploration creates observations, observations mature into patterns, patterns drive recommendations, and new findings along the way spawn new themes.

## Seeding: Where Humans Start the Flywheel

An empty knowledge graph produces no curiosity — there's nothing to connect to, no threads to pull. The graph needs initial mass.

**What to seed (and what not to):**

Seed **distilled strategic knowledge** — the kind of understanding a human expert would carry in their head. Not raw documents (the agent can read those via search tools), but the synthesis: what the methodology actually is, how client engagements work, what patterns recur, what the competitive landscape looks like.

Don't seed raw content. If the agent has access to source documents (a vault, a codebase, a wiki), those are searchable at runtime. The knowledge graph should contain what the agent *thinks about* the content, not the content itself.

**Practical seeding approach:**

| Phase | What | Entity Types | Size |
|-------|------|-------------|------|
| 1. Methodology skeleton | How the practice/product/system works — its structure, phases, deliverables, quality gates | instruction, observation | 10-15 entities |
| 2. Domain intelligence | Client patterns, engagement learnings, cross-cutting observations | observation, pattern | 10-15 entities |
| 3. External context | Competitive intelligence, market positioning, industry frameworks | article, competitor | Agent-driven (web search) |

Phases 1 and 2 are human-guided: the human knows what matters, what the methodology actually is (vs. what the docs say), and which patterns are real vs. coincidental. Phase 3 is agent-driven: once the graph has enough internal knowledge, the agent can seek external context through web search and connect it to what it already knows.

**Seeding with caveats:** If the domain is evolving (methodology being built, product in flux), capture that as an explicit instruction entity: "The methodology is evolving — there will be discrepancies between documentation and actual deliverables. This is expected, not a quality problem." Without this, the agent will flag every discrepancy as a finding, drowning signal in noise.

**Our experience:** 50 entities seeded across 4 types (23 instructions, 17 observations, 5 articles, 5 patterns) with auto-connect at cosine distance threshold 0.28 (tighter than the default 0.35) produced 296 edges automatically. Zero orphans — every entity connected to at least one other. The graph was immediately navigable.

## When Human Involvement Is Needed

The curiosity architecture is designed to minimize human intervention in the investigation loop while keeping humans in control of what matters.

### Humans Required

**Initial knowledge seeding (Phase 1-2).** The agent can't bootstrap its own domain knowledge from scratch — it needs the human's understanding of what matters. This is a one-time investment (hours, not days) that the agent builds on indefinitely.

**Validating recommendations before they become actions.** The agent produces recommendations (`recommendation` entities in knowledgeprim, surfaced through briefing crons). These are proposals, not directives. The human decides whether to act on them, ignore them, or redirect the investigation. The agent never writes to the source material directly — findings go through the human.

**Pruning themes that drift.** Over time, theme spawning can drift from the agent's core domain. A strategic advisor investigating competitive landscape might spawn a theme about "pricing models," which spawns "subscription vs. project billing," which spawns "payment processing integrations" — and now the agent is investigating technical infrastructure, not business strategy. Periodic human review of active themes (quarterly, or when theme count exceeds ~10) keeps investigations on track.

**Correcting bad patterns.** If the agent synthesizes a pattern from observations that the human knows are wrong or misleading, the human needs to correct it — either by editing the pattern entity or by adding a contradicting instruction. Left uncorrected, bad patterns propagate: they inform recommendations, which inform future observations, which reinforce the bad pattern.

### Humans Not Required

**Day-to-day exploration.** Once themes are seeded and the flywheel is running, the agent picks themes, investigates, captures findings, and manages its own knowledge graph. This is the whole point — the agent does the tedious work of reading, searching, comparing, and connecting so the human gets synthesized insights instead of raw data.

**Theme lifecycle management.** The agent creates, explores, matures, resolves, and spawns themes autonomously. The lifecycle rules are in its workflow (AGENTS.md), not enforced by external tooling.

**Graph maintenance.** Auto-connect handles initial edge creation. The agent upgrades `similar_to` edges to meaningful relationships during exploration. `knowledgeprim discover --orphans` and `--weak-edges` are available for self-diagnosis — the agent can identify gaps in its own graph and prioritize filling them.

**External research.** Web search, source verification, competitive scanning — these are agent-driven activities that require no human input. The bar is set by the agent's identity and workflow: verified findings with sources, specific events with dates and URLs, no "consultant-flavored air."

### The Human Review Cadence

| Frequency | What | Why |
|-----------|------|-----|
| Daily | Read the agent's briefings | React to recommendations, provide feedback |
| Weekly | Skim the agent's weekly review | Validate strategic direction |
| Monthly | Review active themes list | Prune drift, suggest new directions |
| Quarterly | Review pattern entities | Correct any bad syntheses |
| As needed | Seed new knowledge after major changes | Keep the graph current after pivots, new clients, strategy shifts |

## knowledgeprim: The Specific Mechanisms

### Entity Types That Support Curiosity

| Type | Role in Curiosity Loop |
|------|----------------------|
| `instruction` | Human-provided ground truth. Anchors the graph — observations and patterns connect back to these. |
| `observation` | What the agent notices. The raw material for patterns. Each exploration session should produce 1-3 of these. |
| `pattern` | Synthesized from 3+ observations. The payoff of the curiosity loop — higher-order findings that individual observations couldn't produce. |
| `article` | External findings with sources. Connects internal knowledge to the outside world. |
| `recommendation` | What the agent proposes to the human. The output of the curiosity loop. |

### Edge Context as Knowledge

Edges are not just structural links — they're knowledge in themselves. The context field on each edge is indexed by both FTS5 and vector search. When the agent searches for "revenue leakage," it finds not just entities about revenue leakage but *connections* that mention it: "This observation about phantom intake connects to the revenue leakage pattern because unbilled services flow through the same gap."

This means the graph's knowledge is in two places: entities (nodes) and edge context (connections). An agent that only captures entities but doesn't write edge context is building a skeleton without tendons.

### Auto-Connect as Serendipity Engine

Auto-connect (cosine distance on capture) creates connections the agent didn't explicitly plan. In this deployment we used cosine distance threshold 0.28 (tighter than primkit's default 0.35 — lower means more similar, so fewer but higher-confidence connections). New entities arrive pre-linked to their most relevant neighbors. Sometimes these connections are obvious (two observations about the same client). Sometimes they're surprising (a methodology observation linked to a competitive intelligence article because they both discuss governance frameworks).

The agent's exploration job includes reviewing auto-connected edges and asking: "Is this connection meaningful? Should I upgrade it to a specific relationship with context? Or is the similarity superficial?" This review process itself generates insights — the agent discovers connections it wouldn't have searched for.

### discover as Self-Diagnosis

`knowledgeprim discover` provides structural analysis that drives exploration priorities:

- `--orphans` — Entities with no edges. These are captured knowledge that hasn't been integrated. Exploration should connect them or determine they're noise.
- `--clusters` — Densely connected groups. These are the agent's areas of deep knowledge. If a cluster is isolated from others, the agent should look for bridges.
- `--bridges` — Cross-cluster connectors. These are the most valuable entities — they connect otherwise separate knowledge domains.
- `--weak-edges` — Edges without context prose. These are connections the agent asserted but didn't explain. During exploration, the agent should fill in the "why."

An agent that periodically runs `discover` and acts on the results is maintaining its own knowledge quality — a form of metacognitive self-improvement.

## stateprim: The Operational Backbone

Stateprim doesn't do knowledge. It does logistics: which themes are active, which areas have been reviewed, which recommendations have already been surfaced.

**Three namespaces for a curious agent:**

| Namespace | Purpose | Prevents |
|-----------|---------|----------|
| `themes` | Active investigations with metadata (status, observation count, last explored) | Starting cold every session |
| `recommendations` | What's already been recommended to the human | Repeating yourself |
| `review-state` | Which areas have been deeply reviewed and when | Always reviewing the same area |

The `themes` namespace is the continuity mechanism. Without it, the agent has a knowledge graph (what it knows) but no research agenda (what it's investigating). Themes bridge the gap.

## What We Built (Grounded Data)

For Lukas, a strategic advisor agent for a consulting practice:

**Knowledge graph (knowledgeprim):** 50 entities, 296 auto-connected edges, 50 vectors, 868KB. Entity types: 23 instructions (methodology ground truth), 17 observations (what the agent noticed), 5 articles (external findings), 5 patterns (synthesized insights). Zero orphans.

**Operational state (stateprim):** 4 active themes seeded from existing graph knowledge. Observation counts range from 0 (unexplored) to 3 (ready for pattern synthesis).

**Exploration cron:** Runs Tuesday/Thursday at 2:37pm, silent (no delivery to user). The agent picks a theme, investigates through vault search + web search + graph traversal, captures findings, looks for patterns, and spawns new themes.

**Communication crons (modified):** Morning insight now starts from active themes, not vault changes. Weekly review draws on accumulated knowledge. Trend scan spawns themes when findings open new threads.

**Identity update:** SOUL.md includes curiosity as a core trait, with the framing that inactivity is coasting — the agent should always be investigating something.

**What we haven't proven yet:** As of this writing, the exploration cron has been deployed but hasn't fired its first run. The design is grounded in the architecture that works for Lukas's other crons and in the principles from building 5 agents over several months — but the specific flywheel (exploration → observation → pattern → recommendation → new theme) is untested at runtime. We'll know after the first week of Tuesday/Thursday exploration runs whether the playbook produces the intended behavior or needs adjustment.

## Principles

1. **Curiosity is identity, not instruction.** "Investigate to a conclusion" in SOUL.md produces better exploration than "run the exploration playbook" in a cron message.

2. **Separate exploration from reporting.** Silent exploration crons build depth. Communication crons surface insights when they're ready. Combining them optimizes for message production, not knowledge building.

3. **Themes are the continuity mechanism.** Without them, a knowledge graph is a filing cabinet. With them, it's a research agenda.

4. **3+ observations → pattern.** This threshold prevents premature synthesis (a single observation is anecdotal) while ensuring the agent doesn't accumulate observations indefinitely without drawing conclusions.

5. **Auto-connect seeds serendipity.** Conservative threshold + agent review of auto-edges produces connections the agent wouldn't have searched for.

6. **Seed the graph with what the human knows, not what the documents say.** The agent can search documents at runtime. The graph should contain synthesized understanding — the expert's mental model, not a document index.

7. **Human review cadence decreases over time.** Heavy involvement at seeding (hours), light touch at steady state (read briefings, prune themes quarterly).

8. **The exploration cron is the only new infrastructure.** Everything else — knowledge graph, state store, communication crons — existed before. Curiosity was achieved by restructuring identity, workflow, and state, not by adding new tools.

---
