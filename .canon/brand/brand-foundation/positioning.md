---
hatchable:
  domain: brand-foundation
  root: canon
  lane: brand
  artifact: positioning
  role: primary
  status: draft
---

# Primkit Positioning

## Category

Primkit's category is local infrastructure primitives for AI agents and humans working with them. `Confirmed`

In buyer terms, Primkit is not a hosted memory product, agent framework, orchestration layer, or AI workspace. It is the local persistence layer underneath agent workflows: four small Go CLIs, backed by embedded SQLite, that let agents and the people operating them keep task, state, knowledge, and queue context across sessions. `Confirmed`

## Positioning Statement

Primkit gives developers and teams using AI coding agents local durable state that survives the session. It does this through four standalone Go CLIs and four SQLite-backed primitives: `taskprim` for task lifecycle and dependency frontiers, `stateprim` for durable key-value state and append logs, `knowledgeprim` for searchable knowledge graphs, and `queueprim` for persistent work queues. `Confirmed`

Primkit's operating model is explicit local file and command use: inspectable, scriptable, and replayable without requiring a server, SDK, API key, Postgres, Redis, or daemon. `Confirmed`

## Context

The audience already works with AI coding agents and needs continuity across resets, tool switches, and multi-session work. The core buyer problem is not "more agent intelligence"; it is durable local operating context that agents and humans can both inspect and reuse. `Confirmed`

Primkit should be understood as infrastructure beneath the agent workflow. It supports the promise that agents can resume work with durable local state instead of making users re-explain project context every session. `Confirmed`

Current public copy and examples already frame the product through session-surviving state, a four-tool loop, concrete shell usage, and file-backed proof. Those site examples are operating evidence only; they remain `De facto` until consolidated by owner decision. `De facto`

## Competitive Context

Primkit should be positioned against failure modes and rejected categories.

| Context | Classification | Implication |
| --- | --- | --- |
| Hosted AI memory platforms | `Rejected` | Do not frame Primkit as remote memory, cloud account memory, or a hosted continuity service. |
| Enterprise agent orchestration platforms | `Rejected` | Do not make orchestration, control-plane management, or enterprise workflow governance the primary category. |
| General agent frameworks | `Rejected` | Keep Primkit as the persistence layer underneath agents, not the framework that builds or runs agents. |
| Abstract AI productivity tools | `Rejected` | Do not claim vague productivity magic; show concrete commands, files, and primitives. |
| Generic developer SaaS | `Rejected` | Avoid broad SaaS framing, inflated claims, and vague diagrams as the source of credibility. |
| Primary category labels: platform, operating system, orchestration, framework, hosted memory, autonomous memory, AI workspace, agent brain | `Rejected` | These labels should not be used as Primkit's primary category. |

## Differentiators

| Differentiator | Classification | Evidence |
| --- | --- | --- |
| Four small Go CLIs: `taskprim`, `stateprim`, `knowledgeprim`, and `queueprim`. | `Confirmed` | `SOURCE:brief-product-summary`; `DECISION:architecture-four-go-sqlite-primitives` |
| Embedded SQLite backing for durable local task, state, knowledge, and queue data. | `Confirmed` | `SOURCE:brief-product-summary`; `DECISION:architecture-four-go-sqlite-primitives` |
| Local durable state that survives agent sessions. | `Confirmed` | `SOURCE:owner-answers-first-pass`; `SOURCE:brief-observed-site-language`; `DECISION:promise-state-survives-session` |
| Explicit local file and command operating model that is inspectable, scriptable, and replayable. | `Confirmed` | `SOURCE:brief-product-summary`; `SOURCE:owner-answers-first-pass`; `DECISION:positioning-explicit-local-operating-model` |
| No server stack: no server, SDK, API key, Postgres, Redis, or daemon required to support the core promise. | `Confirmed` | `SOURCE:owner-answers-first-pass`; `SOURCE:brief-product-summary`; `DECISION:claims-no-server-stack` |
| Terminal-proof posture: current public material proves the promise with terminal examples for all four primitives. | `De facto` | `SOURCE:brief-observed-site-language`; `SOURCE:site-hero-terminal`; `DECISION:application-terminal-proof-mode` |
| Current site framing presents the four tools as one agent loop. | `De facto` | `SOURCE:brief-observed-site-language`; `SOURCE:site-landing-page`; `SOURCE:site-navigation-hero-copy`; `DECISION:messaging-four-tool-agent-loop` |

## Evidence

| Source or decision | Classification | Positioning use |
| --- | --- | --- |
| `DECISION:positioning-local-infrastructure-primitives` | `Confirmed` | Category anchor: Primkit is local infrastructure primitives for AI agents and humans working with them. |
| `DECISION:architecture-four-go-sqlite-primitives` | `Confirmed` | Product proof: four Go CLIs, embedded SQLite, and the named primitive roles. |
| `DECISION:audience-ai-agent-developers` | `Confirmed` | Audience anchor: developers and teams already using AI coding agents; secondary audiences are environment builders and multi-agent workflow operators. |
| `DECISION:promise-state-survives-session` | `Confirmed` | Promise anchor: state survives the session. |
| `DECISION:claims-no-server-stack` | `Confirmed` | Claim boundary: four CLIs, four SQLite files, no server stack. |
| `DECISION:positioning-explicit-local-operating-model` | `Confirmed` | Operating model: explicit local files and commands, inspectable, scriptable, and replayable. |
| `DECISION:vocabulary-allowed-precise-terms` | `Confirmed` | Allowed precise terms: primitives, infrastructure, persistent state, local state, durable state, task frontier, knowledge graph, and queue. |
| `DECISION:messaging-current-site-phrase-set` | `De facto` | Current phrase evidence only; do not treat the full phrase set as approved canonical copy. |
| `DECISION:application-terminal-proof-mode` | `De facto` | Current proof mode evidence only; supports the terminal-proof posture without upgrading it to `Confirmed`. |
| `DECISION:claims-current-trust-strip` | `De facto` | Current trust-strip evidence only; MIT license, pure Go single binaries, and CLI/MCP/HTTP compatibility are observed site claims, not upgraded here. |
| `DECISION:vocabulary-reject-primary-category-labels` | `Rejected` | Category guardrail: do not use the rejected labels as Primkit's primary category. |
| `DECISION:positioning-reject-general-agent-framework` | `Rejected` | Category guardrail: do not frame Primkit as a general agent framework. |

## Classification

This artifact preserves upstream labels exactly and does not upgrade observed operating usage into final authority.

| Classification | Handling in this artifact |
| --- | --- |
| `Confirmed` | Used for category, product architecture, audience, promise, operating model, no-server claim, and allowed precise vocabulary. |
| `De facto` | Used only for current site language, current terminal proof mode, current four-tool loop framing, and current trust-strip claims. These remain observed operating evidence. |
| `Rejected` | Used for rejected categories, anti-references, and claim directions that positioning must avoid. |
| `Missing` | Acknowledged only where upstream inputs withhold final decisions; those gaps do not block this positioning statement. |
| `Inferred` | Not used; upstream source-map artifacts recorded no material `Inferred` decisions. |
| `Proposed` | Not used; upstream source-map artifacts recorded no material `Proposed` decisions. |
| `Conflicting` | Not used; upstream source-map artifacts recorded no actual `Conflicting` decisions. |

Positioning boundary: this file defines category, operating position, competitive exclusions, differentiators, evidence, and classification handling only. It does not set downstream application-system choices beyond these positioning claims.
