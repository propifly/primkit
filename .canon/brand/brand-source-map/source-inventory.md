---
hatchable:
  domain: brand-source-map
  root: canon
  lane: brand
  artifact: source-inventory
  role: primary
  status: draft
---

# Source Inventory

## Source Boundary

Detected mode: intake. The supplied material is a prose brand brief about what Primkit is, who it serves, its practical infrastructure-focused public promise, and the desired brand feel.

This inventory treats the enriched Primkit Brand Brief as the complete brand source set for this run. No operating material, brand guidelines, website copy set, product screenshots, decks, proposals, stakeholder notes, rejected directions, or asset files accompany the prose brief.

Repository implementation files, package manifests, route structure, local documentation, and existing repo assets are outside this source boundary and are not treated as brand evidence.

## Source Inventory

| Source ID | Source | Pointer | Type | Inventory Status |
| --- | --- | --- | --- | --- |
| `SOURCE:primkit-brand-brief` | Primkit Brand Brief | `input:brand-input/brief.md` | Prose intake brief | Included as the sole source |

`SOURCE:primkit-brand-brief` states that Primkit is a family of local infrastructure primitives for AI agents and humans working with them. It describes four Go CLIs backed by embedded SQLite: taskprim for task lifecycle and dependency frontiers, stateprim for durable key-value state and append logs, knowledgeprim for searchable knowledge graphs, and queueprim for persistent work queues. It also states the target users and workflows, the local-first operating model, the practical and infrastructure-focused current public promise, and directional brand feel.

## Source Roles

| Source ID | Allowed Role | What It Supports |
| --- | --- | --- |
| `SOURCE:primkit-brand-brief` | Strategy source | Product category, primitive roles, audience, operating model, practical infrastructure-focused public promise, and high-level positioning. |
| `SOURCE:primkit-brand-brief` | Verbal source | Core words and phrases such as local infrastructure primitives, AI agents, state that survives the session, and no server to run. |
| `SOURCE:primkit-brand-brief` | Constraint source | Directional constraints that the brand should feel local-first, sharp, durable, and agent-native without becoming abstract, cute, or overbuilt. |

No `Visual source`, `Asset source`, `Application source`, `Inspiration source`, or `Rejected source` material is identified in the supplied source set.

## Source Quality

| Source ID | Allowed Quality | Rationale |
| --- | --- | --- |
| `SOURCE:primkit-brand-brief` | Authoritative | It is the supplied controlling intake brief and the declared complete source set for this source-inventory task. |

The source quality vocabulary used here is limited to the allowed values: `Authoritative`, `Strong evidence`, `Weak evidence`, `Reference only`, `Rejected`, and `Unknown`.

## Readability

The brief is fully readable as plain markdown prose. It identifies the company, product family, primitive roles, intended users, operating model, practical infrastructure-focused current public promise, and broad brand-feel constraints.

No source text is unreadable, partially extracted, binary, image-only, PDF-only, or blocked.

## Limitations

This is intake-mode evidence, not operating archaeology. There is no observed real-world brand usage to classify as de facto, and no asset file to inspect for logo geometry, color correctness, typography, layout, or production suitability.

No operating material or asset files accompany the prose brief. Downstream domains should treat this as a clean intake foundation, not as proof of existing market-facing identity.
