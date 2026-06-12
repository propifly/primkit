---
hatchable:
  domain: brand-source-map
  root: canon
  lane: brand
  artifact: decision-ledger
  role: primary
  status: draft
---

# Decision Ledger

Primary runtime artifact: `.canon/brand/brand-source-map/decision-ledger.md`.

## Classification Coverage

This ledger represents only the classifications actually identified for this intake-mode source set:

| Classification | Coverage |
| --- | --- |
| `Confirmed` | Represented. These decisions are captured directly from `SOURCE:primkit-brand-brief`, the authoritative prose intake brief. |
| `Missing` | Represented. These decisions are source-silent or source-incomplete needs carried forward from `.canon/brand/brand-source-map/gaps.md`. |
| `De facto` | Not identified in the supplied sources. Intake mode has no operating material to observe, so no de facto decisions are recorded. |
| `Inferred` | Not identified in the supplied sources. No model-only conclusions are recorded as decisions in this seed ledger. |
| `Proposed` | Not identified in the supplied sources. Source-silent needs are recorded as `Missing`, not proposed defaults. |
| `Rejected` | Not identified in the supplied sources. The brief's avoidance language is recorded as a confirmed constraint, not as an inventoried rejected source or rejected direction. |
| `Conflicting` | Not identified in the supplied sources. `.canon/brand/brand-source-map/conflicts.md` records no conflicting decisions. |

Source role, source quality, evidence, confidence, and approval state are not collapsed into decision classification. Source role and quality remain in `.canon/brand/brand-source-map/source-inventory.md` and `.canon/brand/brand-source-map/source-quality.md`; this ledger uses the Evidence and Confidence columns only to explain why each classification was assigned.

Production Blocker is assessed independently from source-map gap blocking status. Gap evidence identifies source silence or incompleteness for this intake stage; this column answers whether the missing decision prevents production-ready output in the owning domain. `Yes` is reserved for unresolved decisions that block final copy, visual identity, or approved asset use, and those rows cite the source or asset evidence directly instead of the non-blocking gap bucket. Strategic prioritization, reference, and application-context gaps remain `No` unless they are required to approve a production deliverable.

| Decision ID | Decision | Area | Classification | Evidence | Confidence | Owning Domain | Production Blocker |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `DECISION:positioning-local-infrastructure-primitives` | Primkit is a family of local infrastructure primitives for AI agents and the humans working with them. | positioning | `Confirmed` | `SOURCE:primkit-brand-brief` | High - direct source statement. | `brand-foundation` | No |
| `DECISION:positioning-four-sqlite-clis` | Primkit provides four small Go CLIs backed by embedded SQLite: taskprim for task lifecycle and dependency frontiers, stateprim for durable key-value state and append logs, knowledgeprim for searchable knowledge graphs, and queueprim for persistent work queues. | positioning | `Confirmed` | `SOURCE:primkit-brand-brief` | High - direct source statement. | `brand-foundation` | No |
| `DECISION:audience-agentic-local-state-users` | The named audience includes agentic coding environments, multi-agent workflows, and developers who need durable local state beyond one chat session or process. | audience | `Confirmed` | `SOURCE:primkit-brand-brief` | High - direct source statement. | `brand-foundation` | No |
| `DECISION:positioning-explicit-local-operating-model` | Primkit's operating model is explicit local file and command use that can be inspected, scripted, and replayed instead of hosted memory or an always-on server. | positioning | `Confirmed` | `SOURCE:primkit-brand-brief` | High - direct source statement. | `brand-foundation` | No |
| `DECISION:positioning-current-public-promise` | The current public promise is practical and infrastructure-focused: state that survives the session; four CLIs, four SQLite files, no server to run. | positioning | `Confirmed` | `SOURCE:primkit-brand-brief` | High - direct source statement. | `brand-foundation` | No |
| `DECISION:personality-local-sharp-durable-agent-native` | The brand should feel local-first, sharp, durable, and agent-native. | personality | `Confirmed` | `SOURCE:primkit-brand-brief` | High - direct source statement. | `brand-foundation` | No |
| `DECISION:personality-avoid-abstract-cute-overbuilt` | The brand should avoid becoming abstract, cute, or overbuilt. | personality | `Confirmed` | `SOURCE:primkit-brand-brief`; `.canon/brand/brand-source-map/conflicts.md#rejected-directions` | High - direct source constraint; conflicts artifact confirms this is not a rejected-source record. | `brand-foundation` | No |
| `DECISION:audience-primary-priority` | Primary audience priority is not settled among agentic coding environments, multi-agent workflows, and developers working with durable local state. | audience | `Missing` | `SOURCE:primkit-brand-brief`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | High - the brief names audiences but does not rank them. | `brand-foundation` | No - strategic priority gap, not a production deliverable blocker. |
| `DECISION:positioning-lead-emphasis` | The lead positioning emphasis is not settled among local infrastructure primitives, session-surviving state, no-server operation, and agent-native workflow support. | positioning | `Missing` | `SOURCE:primkit-brand-brief`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | High - the gap is explicitly recorded and no hierarchy is supplied. | `brand-foundation` | No - emphasis hierarchy gap, not a production deliverable blocker. |
| `DECISION:personality-trait-balance` | The final weighting of local-first, sharp, durable, agent-native, and avoidance constraints is not settled. | personality | `Missing` | `SOURCE:primkit-brand-brief`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | High - the brief supplies traits but no priority, examples, or edge cases. | `brand-foundation` | No - trait-weighting gap, not a production deliverable blocker. |
| `DECISION:verbal-system-final` | Final voice, tagline, message hierarchy, terminology rules, and example copy are not settled. | verbal | `Missing` | `SOURCE:primkit-brand-brief` | High - the brief supplies phrases and promise language, not a complete verbal system. | `brand-directions` | Yes - blocks production copy system. |
| `DECISION:visual-identity-system` | Logo, mark, color palette, typography, iconography, imagery, diagram style, and layout direction are not settled. | visual | `Missing` | `.canon/brand/brand-source-map/asset-inventory.md#source-grade-assets` | High - no visual or asset source was supplied. | `brand-directions` | Yes - blocks production visual identity. |
| `DECISION:asset-production-sources` | Source-grade asset formats, approved variants, file ownership, licensing, and production readiness are not settled. | asset | `Missing` | `.canon/brand/brand-source-map/asset-inventory.md#missing-obvious-variants` | High - no approved master assets, variants, formats, or licensing records were supplied. | `brand-directions` | Yes - blocks production asset use. |
| `DECISION:references-admired-disliked-anti-directions` | Admired references, disliked references, and explicit anti-directions are not settled. | references | `Missing` | `SOURCE:primkit-brand-brief`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | High - the brief gives broad avoidance language but no reference or anti-reference sources. | `brand-decisions` | No - reference context gap, not a production deliverable blocker. |
| `DECISION:application-priority-contexts` | Priority brand application contexts are not settled. | application | `Missing` | `.canon/brand/brand-source-map/source-inventory.md#source-boundary`; `.canon/brand/brand-source-map/gaps.md#non-blocking-gaps` | High - no website copy set, docs style sample, product screenshot, deck, proposal, social asset, or application source was supplied. | `brand-directions` | No - application-priority gap, not a production deliverable blocker. |
