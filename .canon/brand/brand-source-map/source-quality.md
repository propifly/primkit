---
hatchable:
  domain: brand-source-map
  root: canon
  lane: brand
  artifact: source-quality
  role: primary
  status: draft
---

# Primkit Source Quality

## Quality Vocabulary

This artifact uses only the source quality values allowed by the brand-source-map contract:

- `Authoritative`: supplied owner or brief material that directly governs product facts, source-map handling, constraints, or rejected directions.
- `Strong evidence`: current operating material or current-site summaries with clear provenance and specific brand relevance. Strong evidence can describe de facto usage, but it is not automatically owner-approved final identity.
- `Weak evidence`: readable but limited evidence with low specificity, unclear currentness, thin provenance, or limited reliability. No source in the current inventory required this value.
- `Reference only`: named context, inspiration, calibration material, upstream asset references, or demo material that may inform downstream review but should not become direct brand authority.
- `Rejected`: owner-provided anti-references, forbidden labels, or directions downstream work must avoid.
- `Unknown`: evidence whose provenance or reliability cannot be established well enough to classify. No source in the current inventory required this value; limits are recorded below instead.

Source quality is separate from decision classification. A source can be `Strong evidence` while a downstream decision remains provisional or de facto. A source can be `Authoritative` about a constraint without making every visual or verbal detail final.

## Quality Decisions

| Source ID | Quality | Basis for decision |
| --- | --- | --- |
| `SOURCE:brief-product-summary` | `Authoritative` | High authority and specificity: supplied brief states what Primkit is and the core product promise. Provenance is direct; currentness is tied to the run input. |
| `SOURCE:brief-current-evidence-map` | `Authoritative` | High authority for source boundaries: the brief explicitly names which live-site and asset references count as current evidence for this pass. |
| `SOURCE:brief-observed-site-language` | `Strong evidence` | Specific and current site-language summary, but it reports operating material rather than final approved voice. Reliable for observed phrasing; not final authority. |
| `SOURCE:brief-observed-visual-language` | `Strong evidence` | Specific visual summary from supplied material. Reliable for observed palette, `#0b7285` AA-safe accent-text role, and mark posture, but not enough to approve final identity or pixel-level detail. |
| `SOURCE:owner-answers-first-pass` | `Authoritative` | Highest authority in the set for audience, promise, personality, anti-references, category labels, and owner constraints. Direct provenance from owner answers. |
| `SOURCE:source-map-instructions-this-pass` | `Authoritative` | Direct run instructions with governing force over mode, classification handling, and evidence interpretation. |
| `SOURCE:primkit-site-repo-root` | `Reference only` | Useful as a read-only location reference. Too broad and implementation-shaped to function as direct brand evidence on its own. |
| `SOURCE:site-landing-page` | `Strong evidence` | Current operating public-site material with specific copy and presentation patterns. It is reliable for observed usage, not final owner approval. |
| `SOURCE:site-hero-terminal` | `Strong evidence` | Current operating proof/demo material with concrete CLI transcripts and interaction framing. Strong for product proof language, not final voice system. |
| `SOURCE:site-navigation-hero-copy` | `Strong evidence` | Current JSON source for navigation, hero, tagline, CTA, and cards. Specific and machine-readable, but still site-operating evidence. |
| `SOURCE:site-design-note` | `Strong evidence` | Current design-source note with specific site and asset boundaries. Strong for constraints and open questions; not final brand-system approval. |
| `SOURCE:site-design-tokens` | `Strong evidence` | Specific current token values and usage rules, including the `#0b7285` AA-safe accent-text role. Provenance is clear, but the source itself marks the tokens provisional. |
| `SOURCE:site-token-css` | `Strong evidence` | Implemented current CSS tokens and theme aliases. Reliable for what the site runs today, while still downstream from provisional token decisions. |
| `ASSET:site-robot-mark` | `Strong evidence` | Current public asset used by the site. Metadata is readable; detailed visual interpretation is limited to supplied prose and any future visual tooling. |
| `ASSET:site-social-preview` | `Strong evidence` | Current public social preview asset. Metadata is readable; detailed color, composition, and content claims remain limited. |
| `ASSET:product-logo` | `Reference only` | Product-repo source asset named by the design note, but used here as upstream reference rather than current presentation-owned authority. |
| `ASSET:product-social-preview` | `Reference only` | Product-repo source asset named by the design note. Relevant for lineage, not sufficient for final visual approval. |
| `ASSET:product-demo-gif` | `Reference only` | Product demo media named by the design note. Useful for proof/demo context; binary animation content was not fully extracted. |
| `ASSET:product-demo-mp4` | `Reference only` | Product demo media named by the design note. Useful for proof/demo context; video content was not fully extracted. |
| `SOURCE:product-demo-tape` | `Reference only` | Readable VHS script with demo narrative and command behavior. It informs product proof but is not a governing brand source. |
| `SOURCE:product-demo-setup-script` | `Reference only` | Readable helper script. Reliable for demo setup behavior, but too implementation-specific for brand authority. |
| `SOURCE:inspiration-sqlite` | `Reference only` | Owner-named calibration reference for durable, boring infrastructure. External source material was not fetched or audited. |
| `SOURCE:inspiration-git` | `Reference only` | Owner-named calibration reference for local-first inspectability. External source material was not fetched or audited. |
| `SOURCE:inspiration-ripgrep` | `Reference only` | Owner-named calibration reference for sharp CLI credibility. External source material was not fetched or audited. |
| `SOURCE:inspiration-tailscale-docs` | `Reference only` | Owner-named calibration reference for practical technical clarity. External source material was not fetched or audited. |
| `SOURCE:inspiration-fly-io` | `Reference only` | Owner-named calibration reference for confident infrastructure copy. External source material was not fetched or audited. |
| `SOURCE:inspiration-unix-tool-philosophy` | `Reference only` | Owner-named calibration reference for small composable tools. External source material was not fetched or audited. |
| `SOURCE:rejected-hosted-ai-memory-platforms` | `Rejected` | Direct owner anti-reference. Reliable as a constraint against hosted memory-platform positioning. |
| `SOURCE:rejected-agent-orchestration-platform` | `Rejected` | Direct owner anti-reference. Reliable as a constraint against enterprise orchestration-platform branding. |
| `SOURCE:rejected-cute-mascot-tools` | `Rejected` | Direct owner anti-reference. Reliable as a constraint against mascot-first developer-tool identity. |
| `SOURCE:rejected-crypto-web3-aesthetics` | `Rejected` | Direct owner anti-reference. Reliable as a constraint against crypto or web3 infrastructure aesthetics. |
| `SOURCE:rejected-abstract-ai-productivity` | `Rejected` | Direct owner anti-reference. Reliable as a constraint against vague AI productivity promises. |
| `SOURCE:rejected-generic-developer-saas` | `Rejected` | Direct owner anti-reference. Reliable as a constraint against generic SaaS page structure and inflated claims. |
| `SOURCE:rejected-primary-category-labels` | `Rejected` | Direct owner rejection of primary category labels. Specific, high-authority constraint for naming and positioning. |
| `SOURCE:rejected-mascot-led-voice` | `Rejected` | Direct owner rejection of robot-as-cute-character direction. Specific constraint for voice and mark behavior. |
| `SOURCE:rejected-general-agent-framework` | `Rejected` | Direct owner rejection of general agent framework positioning. Specific constraint for category interpretation. |
| `SOURCE:rejected-casual-visual-departure` | `Rejected` | Direct owner rejection of casually discarding the navy/cyan system. Specific constraint for visual evolution. |

Quality decisions follow these rules:

- Owner answers and brief instructions are `Authoritative` because they directly govern facts, constraints, and run handling.
- Current site files and current site summaries are `Strong evidence` because they show operating usage, but the brief explicitly prevents upgrading them into final owner authority by default.
- Inspiration references are `Reference only` because they are calibration references, not sources to copy or canonize.
- Product demo assets and upstream product-repo assets are `Reference only` unless the site design note or current site usage gives them stronger presentation relevance.
- Rejected labels, anti-references, and forbidden directions are `Rejected` because they are explicit owner constraints.

## Source Limitations

- Absolute paths in the source inventory are read-only supplied evidence references. Downstream evidence pointers should use stable `SOURCE:` ids, `ASSET:` ids, `input:<relative-path>` pointers, or project-relative `.canon/brand/` paths.
- The live `primkit-site` material is current operating evidence. It should support de facto observations, but it should not be treated as confirmed owner-approved final identity.
- The design token JSON and implemented CSS tokens are useful, specific evidence, but the supplied material identifies them as provisional. They should be preserved or deliberately evolved, not silently canonized.
- Binary image, animation, and video assets were inventoried even where extraction was partial. File type and dimensions were readable where noted; detailed logo geometry, image content, animation content, color correctness, and licensing were not inferred.
- Screenshots or social-preview images are visual evidence, not automatically approved production source assets.
- External inspiration references were not fetched or audited. They should calibrate direction generation only at the level described by the owner-supplied notes.
- The source inventory does not settle final voice, final visual identity, font licensing, logo ownership, production asset packaging, or application UI rules.
- Repository implementation files are not brand evidence unless the brief or supplied site design note explicitly named them as brand material.
- Rejected sources are constraints, not gaps. They should prevent downstream drift into disallowed territory.
- `brand-family-contract.md` and `brand-id-contract.md` were not discoverable in this workspace or surrounding Hatchable state tree during upstream inspection. The supplied task/domain contract text governed this artifact.

## Blocked Evidence

| Blocked or limited evidence | Affected source IDs | What is blocked | Downstream impact |
| --- | --- | --- | --- |
| Pixel-level logo interpretation | `ASSET:site-robot-mark`, `ASSET:product-logo` | Detailed logo geometry, exact color sampling, edge quality, mark construction, and production-readiness cannot be verified from metadata alone. | Visual identity work must validate the mark with explicit visual inspection/tooling before treating details as canonical. |
| Social-preview visual details | `ASSET:site-social-preview`, `ASSET:product-social-preview` | Composition, text rendering quality, color correctness, and export suitability cannot be fully verified from metadata alone. | Social/OG guidance should not infer final layout rules from these images without visual review. |
| Demo animation and video content | `ASSET:product-demo-gif`, `ASSET:product-demo-mp4` | Frame content, timing, readability, compression quality, and narrative accuracy were not extracted. | Demo media can inform proof posture only after media review; it should not drive final brand rules from metadata alone. |
| External inspiration source specifics | `SOURCE:inspiration-sqlite`, `SOURCE:inspiration-git`, `SOURCE:inspiration-ripgrep`, `SOURCE:inspiration-tailscale-docs`, `SOURCE:inspiration-fly-io`, `SOURCE:inspiration-unix-tool-philosophy` | External pages, docs, marks, copy, and current presentation were not fetched or audited. | Downstream direction work may use owner-stated qualities, but should not borrow specific language or visual systems. |
| Contract file verification | Source-map contract context | The named `brand-family-contract.md` and `brand-id-contract.md` files were not available for direct file inspection. | This artifact follows the supplied contract text; future runs should compare against the files if they become available. |
| Asset ownership and licensing | All binary/image/media asset IDs | Copyright ownership, font licensing, export rights, and production packaging status were not established. | Production brand and asset packaging work needs separate verification before final release guidance. |

No supplied source set was all-unreadable. The run is not blocked for source-quality documentation, but the blocked evidence above should stop downstream domains from making unsupported visual, media, licensing, or final-approval claims.
