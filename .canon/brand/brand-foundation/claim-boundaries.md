---
hatchable:
  domain: brand-foundation
  root: canon
  lane: brand
  artifact: claim-boundaries
  role: primary
  status: draft
---

# Primkit Claim Boundaries

## Embedded Evidence Context

This downstream slice does not carry separate `brand-decisions` or source-map artifacts, so the decision IDs used below are embedded evidence labels, not unresolved external dependencies. The source-map basis and controlling context are copied into this artifact so a claim can be treated as approved or rejected only when its row below gives a verdict, provenance, and the reason for that verdict.

| Evidence label | Verdict | Source-map basis | Controlling context |
| --- | --- | --- | --- |
| `DECISION:positioning-local-infrastructure-primitives` | `Confirmed` | `SOURCE:brief-product-summary`; `SOURCE:owner-answers-first-pass` | Primary positioning is local infrastructure primitives for AI agents and the humans working with them. |
| `DECISION:audience-ai-agent-developers` | `Confirmed` | `SOURCE:owner-answers-first-pass`; `SOURCE:brief-product-summary` | Primary audience is developers and teams already using AI coding agents who need durable local state across sessions. |
| `DECISION:architecture-four-go-sqlite-primitives` | `Confirmed` | `SOURCE:brief-product-summary` | Product architecture is four small Go CLIs backed by embedded SQLite: `taskprim` for task lifecycle and dependency frontiers, `stateprim` for durable key-value state and append logs, `knowledgeprim` for searchable knowledge graphs, and `queueprim` for persistent work queues. |
| `DECISION:promise-state-survives-session` | `Confirmed` | `SOURCE:owner-answers-first-pass`; `SOURCE:brief-observed-site-language` | Core promise is that agent state survives beyond one session instead of disappearing when the chat, agent run, or shell loop resets. |
| `DECISION:positioning-explicit-local-operating-model` | `Confirmed` | `SOURCE:brief-product-summary`; `SOURCE:owner-answers-first-pass` | Operating model is explicit local file and command use that can be inspected, scripted, and replayed. |
| `DECISION:claims-no-server-stack` | `Confirmed` | `SOURCE:owner-answers-first-pass`; `SOURCE:brief-product-summary` | The supported promise is delivered through four standalone CLIs and four SQLite files; do not imply a required server, SDK, API key, Postgres, Redis, or daemon. |
| `DECISION:vocabulary-allowed-precise-terms` | `Confirmed` | `SOURCE:owner-answers-first-pass` | Precise terms are allowed when literal and accurate: primitives, infrastructure, persistent state, local state, durable state, task frontier, knowledge graph, and queue. |
| `DECISION:positioning-reject-hosted-memory-platforms` | `Rejected` | `SOURCE:rejected-hosted-ai-memory-platforms`; source-map rejected-directions context | Hosted AI memory platform framing conflicts with the local CLI and SQLite operating model. |
| `DECISION:positioning-reject-agent-orchestration-platform` | `Rejected` | `SOURCE:rejected-agent-orchestration-platform`; source-map rejected-directions context | Enterprise agent orchestration platform framing overstates Primkit's role as local primitives underneath agent workflows. |
| `DECISION:positioning-reject-general-agent-framework` | `Rejected` | `SOURCE:rejected-general-agent-framework`; `SOURCE:owner-answers-first-pass` | General agent framework framing implies a full agent runtime or control layer; Primkit supplies lower-level local primitives. |
| `DECISION:vocabulary-reject-primary-category-labels` | `Rejected` | `SOURCE:rejected-primary-category-labels`; `SOURCE:owner-answers-first-pass` | Primary category labels such as platform, operating system, orchestration, framework, hosted memory, autonomous memory, AI workspace, and agent brain overstate the category. |
| `DECISION:claims-reject-abstract-ai-productivity-magic` | `Rejected` | `SOURCE:rejected-abstract-ai-productivity`; source-map rejected-directions context | Broad productivity or magic-intelligence claims are rejected unless they are replaced by concrete command, state, and workflow evidence. |

## Observed Evidence Context

These labels explain why a claim is listed as proof-required. They are not approval evidence and cannot promote a claim to `Confirmed`.

| Evidence label | Status | Captured context |
| --- | --- | --- |
| `OBSERVED:current-trust-strip` | `De facto` | Current trust-strip wording contains license, packaging, named tool compatibility, and interface claims: MIT licensed, pure Go single binaries, Claude Code, Cursor, Codex, OpenClaw, Hermes, CLI, MCP, and HTTP. Each of those claims still needs legal, build, compatibility, or interface proof before approval. |
| `OBSERVED:current-site-wording` | `De facto` | Current site wording implies install-once usage, shell or agent-loop use, work preserved across agent resets, explicit files instead of automatic fleet-like infrastructure, and reduced project re-explanation across sessions. These are observed copy inputs, not final approved claims. |
| `OBSERVED:current-operating-framing` | `De facto` | Current operating framing combines four tools in one agent loop and uses terminal proof. This supports proof-required workflow language but does not approve it as canonical positioning. |
| `OBSERVED:source-listed-signature-phrases` | `Missing` | Source-listed phrases include composable Unix-tool language and persistent-state-for-agents language, but owner confirmation is missing for using them as canonical claims. |

## Approved Claims

These claims are approved for brand-foundation use because the embedded evidence context above classifies them as `Confirmed`.

| Approved claim | Classification | Evidence |
| --- | --- | --- |
| Primkit is local infrastructure primitives for AI agents and humans working with them. | `Confirmed` | `DECISION:positioning-local-infrastructure-primitives` |
| Primkit serves developers and teams already using AI coding agents who need durable local state across sessions. | `Confirmed` | `DECISION:audience-ai-agent-developers` |
| Primkit provides four small Go CLIs backed by embedded SQLite. | `Confirmed` | `DECISION:architecture-four-go-sqlite-primitives` |
| The four primitives are `taskprim`, `stateprim`, `knowledgeprim`, and `queueprim`. | `Confirmed` | `DECISION:architecture-four-go-sqlite-primitives` |
| `taskprim` supports task lifecycle and dependency frontiers. | `Confirmed` | `DECISION:architecture-four-go-sqlite-primitives` |
| `stateprim` supports durable key-value state and append logs. | `Confirmed` | `DECISION:architecture-four-go-sqlite-primitives` |
| `knowledgeprim` supports searchable knowledge graphs. | `Confirmed` | `DECISION:architecture-four-go-sqlite-primitives` |
| `queueprim` supports persistent work queues. | `Confirmed` | `DECISION:architecture-four-go-sqlite-primitives` |
| Primkit gives agents state that survives the session. | `Confirmed` | `DECISION:promise-state-survives-session` |
| Primkit's operating model is explicit local file and command use that can be inspected, scripted, and replayed. | `Confirmed` | `DECISION:positioning-explicit-local-operating-model` |
| The core promise is supported by four standalone CLIs, four SQLite files, and no server, SDK, API key, Postgres, Redis, or daemon. | `Confirmed` | `DECISION:claims-no-server-stack` |
| Precise terms such as primitives, infrastructure, persistent state, local state, durable state, task frontier, knowledge graph, and queue are allowed when accurate. | `Confirmed` | `DECISION:vocabulary-allowed-precise-terms` |

Approved here means source support is sufficient for the claim itself. It does not add separate rules beyond claim support.

## Proof-Required Claims

These claims are observed, current, or plausible, but they are not approved as final brand claims in the supplied source context. They need authoritative proof capture, owner confirmation, or a later `brand-decisions` verdict before being promoted.

| Proof-required claim | Current classification | Why proof is required |
| --- | --- | --- |
| Primkit is MIT licensed. | `De facto` | Listed under `OBSERVED:current-trust-strip`, but not recorded as a `Confirmed` brand claim; legal/license proof is still required. |
| Primkit ships pure Go single binaries. | `De facto` | Listed under `OBSERVED:current-trust-strip`; the confirmed architecture supports "four small Go CLIs", but the stronger packaging claim needs build or release proof. |
| Primkit works with Claude Code, Cursor, Codex, OpenClaw, and Hermes. | `De facto` | Listed under `OBSERVED:current-trust-strip`, but no confirmed compatibility matrix is supplied. |
| Primkit works through CLI, MCP, or HTTP. | `De facto` | Listed under `OBSERVED:current-trust-strip`, but not confirmed as an approved interface boundary. |
| Primkit can be installed once and called from any shell or agent loop. | `De facto` | Listed under `OBSERVED:current-site-wording`; it needs confirmation before becoming a broad supported claim. |
| Primkit preserves work across agent resets. | `De facto` | Listed under `OBSERVED:current-site-wording`; it aligns with the confirmed state-survival promise but is not separately confirmed. |
| Primkit combines four tools in one agent loop. | `De facto` | Listed under `OBSERVED:current-operating-framing`; it should not be upgraded to confirmed positioning without owner verdict. |
| Primkit uses explicit files rather than automatic fleet-like infrastructure. | `De facto` | Listed under `OBSERVED:current-site-wording`; it aligns with confirmed principles but is not approved as a separate claim. |
| Primkit reduces repeated project re-explanation across sessions. | `De facto` | Listed under `OBSERVED:current-site-wording`; it remains operating evidence until confirmed. |
| Primkit is composable like a Unix tool and provides persistent state for AI agents. | `Missing` | Listed under `OBSERVED:source-listed-signature-phrases`, but owner confirmation is missing. |

Rule: current trust-strip and source-listed claims remain proof-required unless a later source-map capture or owner verdict explicitly confirms them. In this artifact, those phrases identify observed working-copy evidence only; they do not function as approval evidence.

## Unsupported Claims

These claims are not supported enough to use. They should be routed into missing decisions or proof work, not turned into promises.

| Unsupported or unresolved claim | Boundary status | Handling |
| --- | --- | --- |
| Primkit improves developer productivity in general terms. | No valid support in supplied sources | Do not use without concrete proof and a decision; current sources reject broad, unsupported productivity positioning. |
| Primkit makes agents autonomous, smarter, self-managing, or generally more intelligent. | No valid support in supplied sources | Do not use; the confirmed promise is durable local state, not intelligence or autonomy. |
| Primkit replaces project management, knowledge management, orchestration, or agent frameworks. | No valid support in supplied sources | Do not use; the confirmed role is infrastructure underneath agent workflows. |

Unsupported claims are not softer approved claims. If a claim cannot be tied to confirmed product facts or proof-required evidence, leave it out and record the missing decision in the next allowed decision-emission pass.

## Rejected Claims

These claims and category labels are hard constraints. They should not be used as primary positioning, approved claim language, or category framing.

| Rejected claim or category | Classification | Evidence |
| --- | --- | --- |
| Primkit is a hosted AI memory platform. | `Rejected` | `DECISION:positioning-reject-hosted-memory-platforms` |
| Primkit is an enterprise agent orchestration platform. | `Rejected` | `DECISION:positioning-reject-agent-orchestration-platform` |
| Primkit is a general agent framework. | `Rejected` | `DECISION:positioning-reject-general-agent-framework` |
| Primkit's primary category is platform, operating system, orchestration, framework, hosted memory, autonomous memory, AI workspace, or agent brain. | `Rejected` | `DECISION:vocabulary-reject-primary-category-labels` |
| Primkit makes broad, unsupported productivity promises instead of showing concrete commands. | `Rejected` | `DECISION:claims-reject-abstract-ai-productivity-magic` |

Rejected claims are not proof-required claims. They are blocked unless a later owner decision explicitly reopens them.

## Escalation

Use this escalation rule for any claim not listed as approved:

1. If the claim matches a `Rejected` row, do not use it.
2. If the claim matches a `Proof-Required Claims` row, use it only as current operating evidence or with explicit proof status.
3. If the claim is unresolved, route it to the relevant missing decision instead of turning it into a promise.
4. If the claim is new and material, send it to `brand-decisions` or a later source-map capture before downstream brand work treats it as approved.
5. If a claim requires product, legal, licensing, compatibility, implementation, or production proof, do not approve it from brand-foundation alone.

Scope note: this run's declared output scope permits only `promise-and-principles.md` and `claim-boundaries.md`, so escalation is documented here without creating sibling decision-emission files.
