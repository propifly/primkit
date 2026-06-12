---
hatchable:
  domain: brand-source-map
  root: canon
  lane: brand
  artifact: decisions-needed
  role: primary
  status: draft
---

# Decisions Needed

## Emission Summary

This first-run emission mirrors every decision currently recorded in `.canon/brand/brand-source-map/decision-ledger.md` because no consolidated `.canon/brand/brand-decisions/ledger.json` is present.

Emitted decisions: 15 total; 7 `Confirmed` decisions newly captured from `SOURCE:primkit-brand-brief`; 8 `Missing` decisions remain open for downstream resolution.

## Confirmed Decisions

These decisions are recorded for downstream consolidation but do not require downstream resolution in this artifact.

| Decision ID | Area | Classification | Decision | Evidence | Owning Domain |
| --- | --- | --- | --- | --- | --- |
| `DECISION:positioning-local-infrastructure-primitives` | positioning | `Confirmed` | Primkit is a family of local infrastructure primitives for AI agents and the humans working with them. | `SOURCE:primkit-brand-brief` | `brand-foundation` |
| `DECISION:positioning-four-sqlite-clis` | positioning | `Confirmed` | Primkit provides four small Go CLIs backed by embedded SQLite: taskprim for task lifecycle and dependency frontiers, stateprim for durable key-value state and append logs, knowledgeprim for searchable knowledge graphs, and queueprim for persistent work queues. | `SOURCE:primkit-brand-brief` | `brand-foundation` |
| `DECISION:audience-agentic-local-state-users` | audience | `Confirmed` | The named audience includes agentic coding environments, multi-agent workflows, and developers who need durable local state beyond one chat session or process. | `SOURCE:primkit-brand-brief` | `brand-foundation` |
| `DECISION:positioning-explicit-local-operating-model` | positioning | `Confirmed` | Primkit's operating model is explicit local file and command use that can be inspected, scripted, and replayed instead of hosted memory or an always-on server. | `SOURCE:primkit-brand-brief` | `brand-foundation` |
| `DECISION:positioning-current-public-promise` | positioning | `Confirmed` | The current public promise is practical and infrastructure-focused: state that survives the session; four CLIs, four SQLite files, no server to run. | `SOURCE:primkit-brand-brief` | `brand-foundation` |
| `DECISION:personality-local-sharp-durable-agent-native` | personality | `Confirmed` | The brand should feel local-first, sharp, durable, and agent-native. | `SOURCE:primkit-brand-brief` | `brand-foundation` |
| `DECISION:personality-avoid-abstract-cute-overbuilt` | personality | `Confirmed` | The brand should avoid becoming abstract, cute, or overbuilt. | `SOURCE:primkit-brand-brief`; `.canon/brand/brand-source-map/conflicts.md#rejected-directions` | `brand-foundation` |

## Open Decisions

Only `Missing` decisions are listed here; these are the unresolved items requiring downstream resolution.

| Decision ID | Area | Classification | Decision | Evidence | Owning Domain |
| --- | --- | --- | --- | --- | --- |
| `DECISION:audience-primary-priority` | audience | `Missing` | Primary audience priority is not settled among agentic coding environments, multi-agent workflows, and developers working with durable local state. | `SOURCE:primkit-brand-brief`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | `brand-foundation` |
| `DECISION:positioning-lead-emphasis` | positioning | `Missing` | The lead positioning emphasis is not settled among local infrastructure primitives, session-surviving state, no-server operation, and agent-native workflow support. | `SOURCE:primkit-brand-brief`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | `brand-foundation` |
| `DECISION:personality-trait-balance` | personality | `Missing` | The final weighting of local-first, sharp, durable, agent-native, and avoidance constraints is not settled. | `SOURCE:primkit-brand-brief`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | `brand-foundation` |
| `DECISION:verbal-system-final` | verbal | `Missing` | Final voice, tagline, message hierarchy, terminology rules, and example copy are not settled. | `SOURCE:primkit-brand-brief`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | `brand-directions` |
| `DECISION:visual-identity-system` | visual | `Missing` | Logo, mark, color palette, typography, iconography, imagery, diagram style, and layout direction are not settled. | `.canon/brand/brand-source-map/asset-inventory.md#source-grade-assets`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | `brand-directions` |
| `DECISION:asset-production-sources` | asset | `Missing` | Source-grade asset formats, approved variants, file ownership, licensing, and production readiness are not settled. | `.canon/brand/brand-source-map/asset-inventory.md#missing-obvious-variants`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | `brand-directions` |
| `DECISION:references-admired-disliked-anti-directions` | references | `Missing` | Admired references, disliked references, and explicit anti-directions are not settled. | `SOURCE:primkit-brand-brief`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | `brand-decisions` |
| `DECISION:application-priority-contexts` | application | `Missing` | Priority brand application contexts are not settled. | `.canon/brand/brand-source-map/source-inventory.md#source-boundary`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | `brand-directions` |

## Emission Agreement

This file mirrors `.canon/brand/brand-source-map/decisions-needed.json` exactly across the `Confirmed Decisions` and `Open Decisions` tables for emitted decision ids, areas, classifications, statements, evidence pointers, and owning domains. `decisions-needed.json` is the typed machine emission consumed by the brand-decisions consolidator.
