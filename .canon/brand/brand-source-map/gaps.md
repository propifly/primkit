---
hatchable:
  domain: brand-source-map
  root: canon
  lane: brand
  artifact: gaps
  role: primary
  status: draft
---

# Primkit Brand Gaps

Gaps are missing decisions, missing source-grade materials, or missing proof needed before production brand guidance can be finalized. They are not conflicts unless two supplied sources disagree, and no such conflict was identified.

Gaps are grouped by owning domain so each downstream domain can see its handoff directly. Severity remains explicit per row: Blocking means the gap blocks final production brand identity, source-package release, or application rules. It does not block completion of this source-map artifact.

## brand-foundation

Owns foundational positioning, audience, promise, and voice decisions before direction generation.

| Gap ID | Severity | Gap | Evidence | Needed Decision |
| --- | --- | --- | --- | --- |
| `GAP:final-voice-system` | Blocking | Final voice system is missing. Current site language and owner answers establish strong constraints, including source-listed signature phrases, but not a complete production voice system. | `SOURCE:brief-observed-site-language`, `SOURCE:owner-answers-first-pass`, `SOURCE:site-landing-page`, `SOURCE:site-hero-terminal`, `SOURCE:site-navigation-hero-copy` | Approve the Primkit voice system: audience priority, tone range, proof-to-promise balance, source-listed signature phrases including "Composable like any Unix tool." and "Persistent state for AI agents.", allowed precise terms, forbidden claims, and examples for landing page, docs, README, CLI help, and social contexts. |
| `GAP:application-rules` | Blocking | Application rules are missing for how the brand should be used across website, docs, README, CLI help, demos, social previews, and future product surfaces. | `SOURCE:owner-answers-first-pass`, `SOURCE:site-landing-page`, `SOURCE:site-hero-terminal`, `SOURCE:site-navigation-hero-copy`, `SOURCE:site-design-note`, `SOURCE:product-demo-tape` | Define where and how Primkit's voice, proof language, mark, palette, typography, terminal demos, diagrams, CTAs, docs cards, and rejected labels apply across each production surface. |
| `GAP:de-facto-copy-examples` | Non-blocking | Current de facto copy examples, owner-listed signature phrases, and the four-question problem framing are spread across site files, summaries, and owner answers, not organized as approved voice examples. | `SOURCE:brief-observed-site-language`, `SOURCE:owner-answers-first-pass`, `SOURCE:site-landing-page`, `SOURCE:site-hero-terminal`, `SOURCE:site-navigation-hero-copy` | Select which current phrases, source-listed signature phrases, and four-question problem framing should become canonical examples, which should remain site-specific, and which should be revised after final voice decisions. |

## brand-visual-identity

Owns production visual identity, asset packaging, typography, color, accessibility, media, and token rules.

| Gap ID | Severity | Gap | Evidence | Needed Decision |
| --- | --- | --- | --- | --- |
| `GAP:final-visual-identity` | Blocking | Final visual identity is missing. The navy/cyan robot system is current operating evidence, but the supplied sources do not approve it as the finished identity. | `SOURCE:brief-observed-visual-language`, `SOURCE:owner-answers-first-pass`, `SOURCE:site-design-note`, `SOURCE:site-design-tokens`, `SOURCE:site-token-css`, `ASSET:site-robot-mark`, `ASSET:site-social-preview` | Decide whether the current navy/cyan robot-led visual system is final, intentionally evolved, or replaced; define the approved palette, mark role, composition rules, imagery posture, and production identity boundaries. |
| `GAP:source-grade-asset-package` | Blocking | Source-grade asset package is missing. All supplied image and media files are export-only or reference-only evidence. | `ASSET:site-robot-mark`, `ASSET:site-social-preview`, `ASSET:product-logo`, `ASSET:product-social-preview`, `SOURCE:site-design-note`, `.canon/brand/brand-source-map/asset-inventory.md#source-grade-assets` | Provide or approve editable source files, logo lockups, mark variants, social-preview templates, source/export naming, versioning, checksums, clearspace, minimum-size rules, and export recipes. |
| `GAP:typography-system` | Blocking | Final typography system is missing. Token slots and CSS usage exist, but no approved type choices, license records, or usage hierarchy were supplied. | `SOURCE:site-design-tokens`, `SOURCE:site-token-css`, `SOURCE:site-design-note` | Approve typefaces, font sources, licenses, fallback stacks, display/UI/mono roles, sizing rhythm, code/terminal text treatment, and loading policy. |
| `GAP:accessibility-rules` | Blocking | Accessibility rules are incomplete. Sources note cyan usage constraints and provisional tokens, but do not provide final contrast, state, or alternative-text rules. | `SOURCE:site-design-tokens`, `SOURCE:site-token-css`, `SOURCE:site-design-note`, `ASSET:site-social-preview`, `ASSET:product-demo-gif`, `ASSET:product-demo-mp4` | Approve accessibility rules for color contrast, cyan-on-light usage, focus states, non-color cues, reduced-motion behavior, media captions or alternatives, social-preview readability, and dark/light surface pairings. |
| `GAP:demo-media-production-use` | Non-blocking | Demo GIF/MP4 production use is not settled. The media supports product proof, but reuse rights, quality, captions, poster frames, and update workflow are not approved. | `ASSET:product-demo-gif`, `ASSET:product-demo-mp4`, `SOURCE:product-demo-tape`, `SOURCE:product-demo-setup-script` | Decide whether demo media is part of the production brand asset package and, if yes, approve source scripts, replay process, captions, poster frames, compression targets, and update ownership. |
| `GAP:icon-diagram-screenshot-system` | Non-blocking | Iconography, diagrams, screenshots, and product-proof illustration rules are not defined. | `SOURCE:site-landing-page`, `SOURCE:site-hero-terminal`, `SOURCE:owner-answers-first-pass`, `.canon/brand/brand-source-map/asset-inventory.md#missing-obvious-variants` | Define whether Primkit uses icons, diagrams, terminal captures, screenshots, or other proof visuals; specify style, density, labels, accessibility, and when each visual form is appropriate. |
| `GAP:token-lineage-and-export-rules` | Non-blocking | Token lineage and export governance are incomplete. Current values are known, but the sources do not settle final token ownership, release format, or change policy. | `SOURCE:site-design-tokens`, `SOURCE:site-token-css`, `SOURCE:site-design-note` | Decide the canonical token source, final token names, color role mapping, dark/light behavior, export format, versioning, and rules for future token changes. |

## brand-decisions

Owns owner verdict consolidation for decisions that require explicit approval, rights confirmation, or rejection.

| Gap ID | Severity | Gap | Evidence | Needed Decision |
| --- | --- | --- | --- | --- |
| `GAP:licensing-and-ownership` | Blocking | Licensing and ownership are unresolved for the mark, social preview, typography, and demo media. | `ASSET:site-robot-mark`, `ASSET:site-social-preview`, `ASSET:product-logo`, `ASSET:product-social-preview`, `ASSET:product-demo-gif`, `ASSET:product-demo-mp4`, `SOURCE:site-design-tokens`, `SOURCE:site-token-css` | Confirm authorship, ownership, license status, permitted uses, redistribution rights, attribution needs, and approval authority for visual assets, fonts, and demo media. |

## brand-directions

Owns use of source-map constraints and reference qualities when developing brand direction options.

| Gap ID | Severity | Gap | Evidence | Needed Decision |
| --- | --- | --- | --- | --- |
| `GAP:inspiration-audit-boundary` | Non-blocking | Inspiration references are owner-named but not externally audited. | `SOURCE:inspiration-sqlite`, `SOURCE:inspiration-git`, `SOURCE:inspiration-ripgrep`, `SOURCE:inspiration-tailscale-docs`, `SOURCE:inspiration-fly-io`, `SOURCE:inspiration-unix-tool-philosophy` | Decide whether downstream direction work should audit these references directly or use only the owner-stated qualities: durable, local-first, sharp CLI, practical docs, confident infrastructure, and composable tools. |

## Owning Domain

Owning domains identify where the decision should be resolved or prepared downstream:

- `brand-foundation`: owns foundational positioning, audience, promise, and voice decisions before direction generation.
- `brand-directions`: owns use of source-map constraints and reference qualities when developing brand direction options.
- `brand-decisions`: owns owner verdict consolidation for decisions that require explicit approval, rights confirmation, or rejection.
- `brand-visual-identity`: owns production visual identity, asset packaging, typography, color, accessibility, media, and token rules.

## Severity

Severity values used here:

- Blocking: must be resolved before final production identity, production asset packaging, or application rules are treated as complete.
- Non-blocking: does not prevent `brand-foundation` or `brand-directions` from proceeding, but must be resolved before the affected production surface or asset type is finalized.

## Needed Decision

Each needed decision is written as reviewable decision text for the owning domain. Downstream work should preserve the distinction between missing decisions and rejected directions: rejected sources constrain what not to do, while the gaps above define what still needs a positive owner or domain decision.
