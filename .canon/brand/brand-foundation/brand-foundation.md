---
hatchable:
  domain: brand-foundation
  root: canon
  lane: brand
  artifact: brand-foundation
  role: primary
  status: draft
---

# Primkit Brand Foundation

## Source Boundary

This foundation uses the supplied `brand-family-contract.md` and `brand-id-contract.md` requirements because local contract files were not available in this workspace. It preserves upstream classification labels exactly and does not upgrade `De facto`, `Missing`, `Rejected`, `Proposed`, `Inferred`, or `Conflicting` material into `Confirmed` authority.

Controlling inputs:

| Source | Role in this foundation |
| --- | --- |
| `.canon/brand/brand-source-map/downstream-brief.md` | Sets the downstream evidence boundary, source limitations, confirmed decisions, de facto decisions, rejected constraints, and missing decisions. |
| `.canon/brand/brand-source-map/decision-ledger.md` | Records material decisions and classification labels that this foundation preserves. |
| `.canon/brand/brand-foundation/positioning.md` | Defines category, positioning, competitive exclusions, differentiators, evidence, and classification handling. |
| `.canon/brand/brand-foundation/audience.md` | Defines primary audience, secondary audience, anti-audience, audience needs, confidence, and missing audience decisions. |
| `.canon/brand/brand-foundation/promise-and-principles.md` | Defines the brand promise, values, strategy principles, behavioral implications, and unresolved foundation decisions. |
| `.canon/brand/brand-foundation/claim-boundaries.md` | Defines approved claims, proof-required claims, unsupported claims, rejected claims, and escalation rules. |
| `.canon/brand/brand-foundation/decisions-needed.md` and `.canon/brand/brand-foundation/decisions-needed.json` | Emit open foundation-owned decisions for `brand-decisions`. |

No consolidated `.canon/brand/brand-decisions/ledger.json` exists in the inspected source set, so this artifact relies on source-map and foundation package evidence. Repository implementation details, framework choices, route paths, provisional visual assets, and unaudited inspiration references are not treated as brand authority unless the source-map package already recorded them as valid evidence.

Source handling:

| Classification | Foundation handling |
| --- | --- |
| `Confirmed` | Used as fixed strategy: category, architecture, audience, promise, no-server-stack claim, operating model, precise vocabulary, and trust/relief/confidence intent. |
| `De facto` | Used only as observed operating evidence: current site phrase set, four-tool loop framing, terminal proof mode, and trust-strip claims recorded in the source map. |
| `Rejected` | Treated as hard constraints against rejected categories, inflated claim styles, mascot-led product framing, generic developer SaaS framing, and abstract productivity promises. |
| `Missing` | Routed to `brand-decisions` or downstream owning domains; not resolved by this artifact. This includes final canonical handling of the source-listed signature phrases and the four-question problem framing preserved in Claim Boundaries. |
| `Inferred`, `Proposed`, `Conflicting` | No material rows were recorded in the supporting foundation package. |

## Positioning Summary

`Confirmed`: Primkit is local infrastructure primitives for AI agents and the humans working with them.

In buyer terms, Primkit is the local persistence layer underneath agent workflows. It gives developers and teams four standalone Go CLIs backed by embedded SQLite:

| Primitive | `Confirmed` role |
| --- | --- |
| `taskprim` | Task lifecycle and dependency frontiers. |
| `stateprim` | Durable key-value state and append logs. |
| `knowledgeprim` | Searchable knowledge graphs. |
| `queueprim` | Persistent work queues. |

`Confirmed`: Primkit's operating model is explicit local file and command use that can be inspected, scripted, and replayed instead of hosted memory or an always-on server.

`Rejected`: Primkit must not be positioned as a hosted AI memory platform, enterprise agent orchestration platform, general agent framework, abstract AI productivity product, generic developer SaaS product, autonomous memory product, AI workspace, platform, operating system, orchestration layer, or agent brain.

`De facto`: Current site framing presents Primkit through a four-tool agent loop, concrete terminal proof, and SQLite-file proof. This is operating evidence, not final application-rule approval.

## Audience Summary

`Confirmed`: Primkit serves developers and teams already using AI coding agents who need durable local state across sessions.

This audience is technically capable. They already understand agent workflows and need task, state, knowledge, and queue context to survive resets, shell/agent loops, and multi-session work. They need less re-explanation, more inspectability, and a local state model that agents and humans can both reuse.

`Confirmed`: Secondary audiences are agentic coding environment builders and multi-agent workflow operators. They are adjacent because they assemble, standardize, or operate workflows that need the same persistent local primitives.

`Rejected`: Primkit is not for buyers seeking hosted AI memory, broad enterprise orchestration, a general agent framework, autonomous memory, a mascot-led developer toy, abstract AI productivity magic, generic SaaS claims, or a full platform that hides state.

Open audience decisions remain `Missing`:

| Decision | Summary |
| --- | --- |
| `DECISION:audience-first-round-emphasis` | Whether first-round brand directions should lead with individual developers using agents day to day or teams standardizing agent workflows. |
| `DECISION:audience-buyer-user-split` | Whether any separate procurement, executive buyer, or departmental segment exists beyond developers and teams. |
| `DECISION:audience-secondary-ranking` | Whether agentic coding environment builders or multi-agent workflow operators should lead as the secondary audience, or remain co-secondary. |

## Promise

`Confirmed`: Primkit gives AI agents local durable state that survives the session.

Approved short-form promise:

> Give your agents state that survives the session.

The promise is concrete, not aspirational. It is supported by four standalone CLIs, four SQLite-backed primitives, local files and commands, and no required server, SDK, API key, Postgres, Redis, or daemon.

Promise boundaries:

| Boundary | Classification | Implication |
| --- | --- | --- |
| The promise centers on durable local state. | `Confirmed` | Keep the promise tied to surviving task, state, knowledge, and queue context. |
| Proof should follow the promise quickly. | `Confirmed` | Use concrete commands, SQLite files, and terminal examples. |
| The promise does not claim broad productivity magic. | `Rejected` | Do not imply general productivity, intelligence, autonomy, or self-management without concrete proof and decision approval. |
| Current site phrases are preserved for traceability. | `De facto` | Do not treat them as canonical voice, final copy, or direction requirements; route adoption through `brand-decisions`. |

## Personality Constraints

This foundation records strategy-level behavior and decision boundaries only. It does not approve a verbal personality, visual identity, palette, mascot, mark treatment, reference set, or surface system.

Evidence-supported constraints:

| Constraint | Classification | Use |
| --- | --- | --- |
| Primkit's confirmed behavior is local, inspectable, scriptable, and replayable. | `Confirmed` | Downstream work may use this as product strategy, not as an approved tone, mood, or personality system. |
| Primkit should support trust that the product is simple, inspectable, and real, with relief and confidence as secondary outcomes. | `Confirmed` | This is an audience outcome rooted in product behavior; verbal and visual expression remain downstream decisions. |
| Precise terms such as primitives, infrastructure, persistent state, local state, durable state, task frontier, knowledge graph, and queue are supported when accurate. | `Confirmed` | This preserves product vocabulary without approving final copy style. |
| Current public phrases and terminal proof patterns are observed operating evidence. | `De facto` | They remain traceability evidence only; final voice, copy, and application rules are unresolved. |
| Hosted memory, enterprise orchestration, general framework, crypto/web3, generic SaaS, abstract productivity, and mascot-led product framing are rejected. | `Rejected` | These are category and claim boundaries, not a complete personality or visual system. |

Not defined here: downstream verbal systems, visual systems, palette approval, color preservation rules, logo, mascot or robot handling, typography, imagery, layout, templates, product UI, per-surface copy, inspiration references, admired reference qualities, or production application rules. This package only records evidence-backed strategy constraints and routes unresolved expression decisions to `brand-directions`, downstream verbal/visual identity, or `brand-decisions`.

## Principles

| Principle | Classification | Practical implication |
| --- | --- | --- |
| Lead with surviving state. | `Confirmed` | Start from the durable local state promise when explaining the brand. |
| Prove the promise. | `Confirmed` | Follow claims with concrete commands, SQLite files, and terminal examples. |
| Name the primitives when architecture matters. | `Confirmed` | Use `taskprim`, `stateprim`, `knowledgeprim`, and `queueprim` when the claim depends on product structure. |
| Keep the category narrow. | `Confirmed` | Treat Primkit as local infrastructure primitives, not a platform, framework, workspace, or hosted memory layer. |
| Keep state inspectable. | `Confirmed` | Favor claims that show where state lives and how humans or agents can inspect it. |
| Keep work scriptable. | `Confirmed` | Preserve the CLI, shell, and explicit-command operating model. |
| Keep claims replayable. | `Confirmed` | Avoid claims that cannot be checked through concrete commands, SQLite files, or terminal examples. |
| Reject magic language. | `Rejected` | Do not replace concrete proof with vague claims about productivity, intelligence, autonomy, or automatic memory. |

These principles are strategy constraints. They are not downstream verbal systems, visual systems, surface templates, or product UI directions.

## Differentiators

| Differentiator | Classification | Why it matters |
| --- | --- | --- |
| Four small Go CLIs: `taskprim`, `stateprim`, `knowledgeprim`, and `queueprim`. | `Confirmed` | Gives the product a concrete architecture instead of a vague agent-memory claim. |
| Embedded SQLite backing for task, state, knowledge, and queue data. | `Confirmed` | Makes the durable local state promise inspectable and practical. |
| Local durable state that survives agent sessions. | `Confirmed` | Directly addresses the audience's reset, re-explanation, and continuity problem. |
| Explicit local file and command operating model. | `Confirmed` | Makes Primkit scriptable, replayable, and usable by both humans and agents. |
| No required server, SDK, API key, Postgres, Redis, or daemon for the core promise. | `Confirmed` | Separates Primkit from hosted memory, orchestration platforms, and heavier infrastructure layers. |
| Terminal-proof posture. | `De facto` | Current site proof supports credibility, but final proof-mode application remains unresolved. |
| Four tools composing into one agent loop. | `De facto` | Current public framing is useful operating evidence, but not yet a final canonical messaging system. |

Differentiators should stay tied to confirmed product facts or clearly labeled de facto evidence. Unsupported claims should not be converted into softer differentiators.

## Claim Boundaries

Approved foundation claims are `Confirmed` and may be used when accurate:

| Approved claim | Classification |
| --- | --- |
| Primkit is local infrastructure primitives for AI agents and humans working with them. | `Confirmed` |
| Primkit serves developers and teams already using AI coding agents who need durable local state across sessions. | `Confirmed` |
| Primkit provides four small Go CLIs backed by embedded SQLite. | `Confirmed` |
| The four primitives are `taskprim`, `stateprim`, `knowledgeprim`, and `queueprim`. | `Confirmed` |
| Primkit gives agents state that survives the session. | `Confirmed` |
| Primkit's operating model is explicit local file and command use that can be inspected, scripted, and replayed. | `Confirmed` |
| The core promise does not require a server, SDK, API key, Postgres, Redis, or daemon. | `Confirmed` |
| Precise terms such as primitives, infrastructure, persistent state, local state, durable state, task frontier, knowledge graph, and queue are allowed when accurate. | `Confirmed` |

Proof-required claims remain current or plausible but not approved as final brand claims:

| Proof-required claim area | Current status |
| --- | --- |
| MIT license, pure Go single binaries, named compatibility with Claude Code, Cursor, Codex, OpenClaw, and Hermes, and CLI/MCP/HTTP interface claims. | `De facto`; needs proof or owner confirmation before becoming approved brand claims. |
| Installation phrases, reset/work contrast, four-tool loop phrasing, file-not-fleet phrasing, and stop-re-explaining phrasing. | `De facto`; may inform downstream work but is not canonical copy. |
| Terminal examples as the final proof mode. | `De facto`; final application rules are still `Missing`. |
| Source-listed phrase adoption and four-question problem-framing adoption. | Source text is preserved below for traceability. Canonical handling is `Missing`; route through `DECISION:messaging-de-facto-copy-examples`. |

Source-listed messaging inputs retained for traceability:

| Source-listed input | Foundation handling |
| --- | --- |
| "Give your agents state that survives the session." | `Confirmed` as the approved short-form promise above. |
| "State that survives the session." | Source-listed input. Canonical handling is `Missing`; do not treat as approved copy. |
| "The agent resets. The work shouldn't." | Observed current-site use is `De facto` where covered by `DECISION:messaging-current-site-phrase-set`; canonical handling is `Missing`. |
| "Four standalone CLIs. Four SQLite files. No daemon." | Underlying no-server-stack support is `Confirmed`; exact copy adoption is `Missing`. |
| "Explicit, not automatic. A file, not a fleet." | Observed current-site use is `De facto` where covered by `DECISION:messaging-current-site-phrase-set`; canonical handling is `Missing`. |
| "Four tools, one agent loop." | Observed current-site use is `De facto` where covered by `DECISION:messaging-current-site-phrase-set`; canonical handling is `Missing`. |
| "Composable like any Unix tool." | Owner-listed reference-quality input. Canonical copy handling is `Missing`; do not treat as observed current-site usage. |
| "Persistent state for AI agents." | Source-listed phrase input. Canonical handling is `Missing`; do not treat as observed current-site usage. |
| Four-question problem framing: "what was I doing"; "did we already do this"; "what did we learn"; "what still needs to run". | Owner-listed problem-framing input. Canonical use is `Missing`; do not treat as observed current-site usage. |

Unsupported or rejected claim handling:

| Claim type | Handling |
| --- | --- |
| Final voice system, production application rules, canonical current-site copy, buyer segmentation, and current-site proof approvals. | Route through `brand-decisions` using the existing `Missing` emissions. |
| Final visual identity, asset package, typography, accessibility, imagery, token governance, licensing, and ownership. | Do not resolve in brand-foundation; these belong to downstream owning domains or `brand-decisions`. |
| Abstract productivity, autonomy, intelligence, self-management, replacement of project management or agent frameworks, hosted memory, generic SaaS, and orchestration-platform claims. | Do not use; these are unsupported or `Rejected` by the source package. |

Escalation rule: if a material claim is not approved here, either label it as proof-required operating evidence, route it through the relevant `Missing` decision, or leave it out.

## Downstream Handoff

`brand-directions` is the immediate downstream consumer and owns the first coherent direction options within the fixed strategy stated here.

`Confirmed`: First-round brand direction should prioritize the website landing page and docs; CLI help and command examples are secondary.

`Missing`: Palette handling, current color equity, mascot or robot treatment, admired reference handling, and inspiration-audit boundaries remain unresolved expression decisions. This foundation does not make them confirmed requirements for downstream verbal or visual identity work.

`Missing`: `DECISION:direction-inspiration-audit-boundary` remains open. Direction work should treat named references and owner-stated reference qualities as unapproved inputs until audited or resolved through `brand-decisions`.

`brand-verbal-identity` and `brand-visual-identity` consume foundation strategy through the selected direction, not directly from this package.

Open decisions route through `brand-decisions`, where the owner records verdicts.
