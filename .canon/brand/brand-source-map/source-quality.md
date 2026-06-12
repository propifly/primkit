---
hatchable:
  domain: brand-source-map
  root: canon
  lane: brand
  artifact: source-quality
  role: primary
  status: draft
---

# Source Quality

## Quality Vocabulary

Allowed source quality values are exactly:

| Quality Value | Meaning |
| --- | --- |
| `Authoritative` | Supplied or owner-declared material that controls the source boundary for this run. |
| `Strong evidence` | Specific, current, and reliable material that supports a source-backed conclusion but is not itself the controlling authority. |
| `Weak evidence` | Limited, ambiguous, incomplete, or low-specificity material that can inform questions but should not settle direction by itself. |
| `Reference only` | Material useful for context, inspiration, comparison, or orientation, but not valid authority for brand decisions. |
| `Rejected` | Material explicitly marked as not to be used, or useful only as evidence of what to avoid. |
| `Unknown` | Material whose provenance, authority, or relevance cannot be established from the supplied source set. |

No numeric scores, extra authority labels, approval-state labels, or decision classifications are used as source quality values.

Source quality is separate from decision classification. A source can be `Authoritative` while specific downstream decisions remain `Missing`, `Inferred`, or otherwise unresolved if the source does not settle them.

## Quality Decisions

| Source ID | Quality Value | Rationale |
| --- | --- | --- |
| `SOURCE:primkit-brand-brief` | `Authoritative` | The brief is the sole supplied source, is identified by the upstream inventory as the complete brand source set for this run, and directly states Primkit's product category, primitive roles, audience, operating model, practical infrastructure-focused public promise, and directional brand constraints. |

This quality assignment is conservative in scope. `SOURCE:primkit-brand-brief` is authoritative for this intake run because it is the declared controlling source, not because it proves existing public usage, production asset correctness, customer reception, or owner approval of every possible downstream brand decision.

## Source Limitations

The source set contains readable prose only. The supplied brief is useful for strategy, verbal direction, audience context, product framing, and broad constraints.

The source set does not include operating archaeology: no brand guidelines, website copy set, product screenshots, decks, proposals, posts, stakeholder notes, rejected directions, or observed market-facing usage were supplied.

The source set does not include inspectable visual or production assets. It cannot verify logo geometry, color correctness, typography, layout systems, file formats, licensing, export quality, accessibility, or application behavior.

Repository implementation files, package manifests, route structure, local documentation, and existing repo assets are not brand evidence for this run.

## Blocked Evidence

No unreadable binary, PDF, image, or asset source was supplied.

No source was blocked by extraction failure, unreadable formatting, missing file access, or partial binary parsing.

Blocked evidence is therefore empty for this artifact. The limitations above are intake-mode limits, not blocked-source failures.
