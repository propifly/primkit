---
hatchable:
  domain: brand-source-map
  root: canon
  lane: brand
  artifact: downstream-brief
  role: primary
  status: draft
---

# Downstream Brief

## Mode

Intake mode - the supplied source set is one authoritative prose brief, `SOURCE:primkit-brand-brief`, with no operating material or asset files to observe.

## Run Readiness

Ready for downstream brand work.

The brief identifies Primkit, explains its product family and primitive roles, names its audience context, states the local operating model, provides the practical infrastructure-focused current public promise, and gives broad personality constraints. The recorded gaps are non-blocking for the brand chain: they should guide downstream work and owner review, not stop the run.

Do not treat missing visual, verbal, asset, reference, or application decisions as approved direction. They are open decisions to resolve through downstream brand work and owner verdicts.

## Controlling Sources

| Source ID | Role | Quality | Downstream Use |
| --- | --- | --- | --- |
| `SOURCE:primkit-brand-brief` | `Strategy source`, `Verbal source`, `Constraint source` | `Authoritative` | Controls product category, primitive roles, audience context, operating model, practical infrastructure-focused public promise, useful phrases, and broad brand-feel constraints for this intake run. |

Supporting `.canon/brand` artifacts for downstream handoff:

- `.canon/brand/brand-source-map/source-inventory.md`
- `.canon/brand/brand-source-map/asset-inventory.md`
- `.canon/brand/brand-source-map/source-quality.md`
- `.canon/brand/brand-source-map/conflicts.md`
- `.canon/brand/brand-source-map/gaps.md`
- `.canon/brand/brand-source-map/decision-ledger.md`
- `.canon/brand/brand-source-map/owner-questions.md`
- `.canon/brand/brand-source-map/decisions-needed.md`
- `.canon/brand/brand-source-map/decisions-needed.json`

## Source Limitations

The source set is readable prose only. It does not include brand guidelines, website copy, product screenshots, decks, proposals, posts, stakeholder notes, rejected directions, inspiration sources, application examples, or observed public usage.

No visual or production asset files were supplied. There is no source-grade logo, mark, color palette, typography specification, iconography, imagery, diagram style, template, licensing record, export set, or approved application system.

Repository implementation files and possible existing repo assets are outside the brand evidence boundary for this run. Do not use them as brand authority unless a later brand-domain run explicitly brings them into scope as valid brand material.

## Classification Coverage

| Classification | Coverage |
| --- | --- |
| `Confirmed` | Represented. Seven decisions are captured directly from `SOURCE:primkit-brand-brief`. |
| `Missing` | Represented. Eight decisions remain source-silent or source-incomplete and are carried forward for downstream resolution. |
| `De facto` | Not identified in the supplied sources. Intake mode has no operating material to observe. |
| `Inferred` | Not identified in the supplied sources. No model-only conclusions are recorded as decisions. |
| `Proposed` | Not identified in the supplied sources. Source-silent needs remain `Missing`, not proposed defaults. |
| `Rejected` | Not identified in the supplied sources. The avoidance language is a confirmed constraint, not an inventoried rejected source. |
| `Conflicting` | Not identified in the supplied sources. No competing source statements were supplied. |

## Confirmed Decisions

| Decision ID | Confirmed Decision | Owning Domain |
| --- | --- | --- |
| `DECISION:positioning-local-infrastructure-primitives` | Primkit is a family of local infrastructure primitives for AI agents and the humans working with them. | `brand-foundation` |
| `DECISION:positioning-four-sqlite-clis` | Primkit provides four small Go CLIs backed by embedded SQLite: taskprim for task lifecycle and dependency frontiers, stateprim for durable key-value state and append logs, knowledgeprim for searchable knowledge graphs, and queueprim for persistent work queues. | `brand-foundation` |
| `DECISION:audience-agentic-local-state-users` | The named audience includes agentic coding environments, multi-agent workflows, and developers who need durable local state beyond one chat session or process. | `brand-foundation` |
| `DECISION:positioning-explicit-local-operating-model` | Primkit's operating model is explicit local file and command use that can be inspected, scripted, and replayed instead of hosted memory or an always-on server. | `brand-foundation` |
| `DECISION:positioning-current-public-promise` | The current public promise is practical and infrastructure-focused: state that survives the session; four CLIs, four SQLite files, no server to run. | `brand-foundation` |
| `DECISION:personality-local-sharp-durable-agent-native` | The brand should feel local-first, sharp, durable, and agent-native. | `brand-foundation` |
| `DECISION:personality-avoid-abstract-cute-overbuilt` | The brand should avoid becoming abstract, cute, or overbuilt. | `brand-foundation` |

## De Facto Decisions

There are no `De facto` decisions in intake mode.

No operating material was supplied, so this run cannot identify consistently observed real-world usage. Downstream work should not treat any name, phrase, visual choice, asset, layout, or application pattern as de facto identity unless later archaeology-mode evidence supports it.

## Blocking Gaps

No blocking gaps are identified for the brand-source-map stage or for continuing the brand chain.

Non-blocking gaps that downstream work must keep visible:

| Decision ID | Missing Decision | Owning Domain | Production Impact |
| --- | --- | --- | --- |
| `DECISION:audience-primary-priority` | Primary audience priority is not settled among agentic coding environments, multi-agent workflows, and developers working with durable local state. | `brand-foundation` | Shapes strategic emphasis but does not block the next domain. |
| `DECISION:positioning-lead-emphasis` | The lead positioning emphasis is not settled among local infrastructure primitives, session-surviving state, no-server operation, and agent-native workflow support. | `brand-foundation` | Shapes strategic emphasis but does not block the next domain. |
| `DECISION:personality-trait-balance` | The final weighting of local-first, sharp, durable, agent-native, and avoidance constraints is not settled. | `brand-foundation` | Shapes tone and direction choices but does not block the next domain. |
| `DECISION:verbal-system-final` | Final voice, tagline, message hierarchy, terminology rules, and example copy are not settled. | `brand-directions` | Blocks production copy system until resolved. |
| `DECISION:visual-identity-system` | Logo, mark, color palette, typography, iconography, imagery, diagram style, and layout direction are not settled. | `brand-directions` | Blocks production visual identity until resolved. |
| `DECISION:asset-production-sources` | Source-grade asset formats, approved variants, file ownership, licensing, and production readiness are not settled. | `brand-directions` | Blocks production asset use until resolved. |
| `DECISION:references-admired-disliked-anti-directions` | Admired references, disliked references, and explicit anti-directions are not settled. | `brand-decisions` | Owner taste should be collected before treating references as direction constraints. |
| `DECISION:application-priority-contexts` | Priority brand application contexts are not settled. | `brand-directions` | Shapes examples and checks but does not block direction generation. |

## Conflicts Not To Paper Over

No conflicts are identified in the supplied sources.

The absence of conflicts should not be used to imply that missing decisions are settled. The brief's avoidance constraint - do not become abstract, cute, or overbuilt - is confirmed source guidance, not a rejected-source inventory and not a conflict.

## Suggested Next Domains

1. `brand-foundation` is always next. It should use the confirmed positioning, primitive roles, audience, operating model, practical infrastructure-focused public promise, and personality constraints while keeping unresolved audience priority, lead positioning emphasis, and personality weighting visible.
2. `brand-directions` should run after `brand-foundation`. It should explore verbal, visual, asset, and application directions without treating missing source evidence as approval.
3. `brand-decisions` can run at any point to consolidate emissions and collect owner verdicts, especially for audience priority, positioning emphasis, references, anti-references, signature language, avoided words, and other open owner judgments.

## Blocked Downstream Domains

No downstream brand domain is blocked by this source map.

The current blockers are production-level decisions, not run blockers: final voice, final visual identity, source-grade assets, and approved application rules are not ready for production use until downstream work and owner verdicts settle them.
