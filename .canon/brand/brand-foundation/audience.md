---
hatchable:
  domain: brand-foundation
  root: canon
  lane: brand
  artifact: brand-foundation
  role: primary
  status: draft
---

# Primkit Audience

## Primary Audience

`Confirmed`: Primkit serves developers and teams already using AI coding agents who need durable local state across sessions.

This audience is not beginner-generic "AI users." They are capable technical users who already understand agent workflows and need the work to survive beyond an agent session, shell loop, or coding environment reset.

Evidence:

- `DECISION:audience-ai-agent-developers`
- `SOURCE:owner-answers-first-pass`
- `SOURCE:brief-product-summary`
- `.canon/brand/brand-source-map/decision-ledger.md`

## Secondary Audience

`Confirmed`: Secondary audiences are agentic coding environment builders and multi-agent workflow operators.

These audiences are adjacent to the primary audience because they assemble, standardize, or operate agent workflows that need the same local persistence layer: tasks, state, knowledge, and queues that can be inspected and reused by agents and humans.

Evidence:

- `DECISION:audience-ai-agent-developers`
- `DECISION:architecture-four-go-sqlite-primitives`
- `SOURCE:owner-answers-first-pass`
- `SOURCE:brief-product-summary`

## Anti-Audience

`Rejected`: Primkit is not for buyers looking for a hosted AI memory platform, a generic enterprise agent orchestration platform, a general agent framework, an autonomous memory product, an AI workspace, an agent brain, or a mascot-led developer toy.

`Rejected`: Primkit should not be framed for audiences seeking abstract AI productivity magic, generic developer SaaS promise language, crypto/web3-coded infrastructure, vague diagrams in place of proof, or a full platform that hides the underlying state.

This anti-audience is intentionally narrow. The sources reject specific categories, labels, and reference directions; they do not support inventing broader excluded segments such as all non-developers, all enterprise teams, or all non-local infrastructure users.

Evidence:

- `DECISION:positioning-reject-hosted-memory-platforms`
- `DECISION:positioning-reject-agent-orchestration-platform`
- `DECISION:positioning-reject-general-agent-framework`
- `DECISION:vocabulary-reject-primary-category-labels`
- `DECISION:claims-reject-abstract-ai-productivity-magic`
- `DECISION:application-reject-generic-developer-saas`
- `DECISION:tone-reject-cute-mascot-tools`
- `DECISION:voice-reject-mascot-led-voice`
- `.canon/brand/brand-source-map/conflicts.md#rejected-directions`

## Needs

`Confirmed`: The audience needs agent state that survives the session.

`Confirmed`: The audience needs durable local primitives for the recurring parts of agent work: task lifecycle and dependency frontiers, key-value state and append logs, searchable knowledge graphs, and persistent work queues.

`Confirmed`: The audience needs inspectability. Primkit's state should remain legible as local files and SQLite-backed data instead of disappearing into a hosted memory layer or opaque platform.

`Confirmed`: The audience needs scriptability. The product should work through CLIs and explicit commands that developers, teams, shells, agent loops, MCP clients, or HTTP clients can call directly.

`Confirmed`: The audience needs replayability. The operating model should make work reconstructable through files, commands, and terminal proof rather than only through the previous agent's conversational context.

`Confirmed`: The audience needs reduced re-explanation. Primkit should help developers stop re-stating project context, task state, and workflow memory every time an agent session resets.

Evidence:

- `DECISION:promise-state-survives-session`
- `DECISION:claims-no-server-stack`
- `DECISION:positioning-explicit-local-operating-model`
- `DECISION:application-terminal-proof-mode`
- `DECISION:messaging-current-site-phrase-set`
- `SOURCE:brief-observed-site-language`
- `SOURCE:site-hero-terminal`

## Source Confidence

High confidence:

- Primary audience, secondary audience, session-surviving state need, and local infrastructure posture are `Confirmed` in the source-map decision ledger.
- Product architecture is `Confirmed`: four small Go CLIs backed by embedded SQLite.
- The local, inspectable, scriptable, replayable operating model is `Confirmed`.
- Rejected audience-adjacent categories and anti-references are explicit `Rejected` constraints.

Medium confidence:

- The exact downstream emphasis between individual developers and teams is not settled. Both are included in the confirmed primary audience, but owner question 1 asks which should lead in first brand directions.

Not used as audience authority:

- Route paths, framework choices, and unrelated implementation details.
- Export-only visual assets beyond their recorded role as current operating evidence.
- Named inspiration references beyond the owner-stated qualities already captured by the source map.

## Missing Audience Decisions

`Missing`: The first-round audience emphasis is unresolved: should brand directions speak more to individual developers using agents day to day, or to teams standardizing agent workflows?

Evidence:

- `.canon/brand/brand-source-map/owner-questions.md`
- `SOURCE:owner-answers-first-pass`

`Missing`: The sources do not define a buyer-versus-user split beyond "developers and teams." Do not invent procurement roles, executive buyers, or departmental segments without a later owner decision.

Evidence:

- `DECISION:audience-ai-agent-developers`
- `.canon/brand/brand-source-map/decision-ledger.md`

`Missing`: The sources do not rank the two secondary audiences. Treat agentic coding environment builders and multi-agent workflow operators as co-secondary until an owner verdict or downstream decision says otherwise.

Evidence:

- `DECISION:audience-ai-agent-developers`
- `.canon/brand/brand-source-map/decision-ledger.md`
