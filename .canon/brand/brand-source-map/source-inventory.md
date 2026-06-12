---
hatchable:
  domain: brand-source-map
  root: canon
  lane: brand
  artifact: source-inventory
  role: primary
  status: draft
---

# Primkit Brand Source Map

## Source Boundary

Detected mode: archaeology mode. The supplied material includes real operating material: the live `primkit-site` repository references, current landing-page copy, current public assets, interim token files, implemented CSS tokens, a site design note, product demo assets, owner answers, inspiration references, anti-references, and rejected directions.

This artifact inventories evidence only. It does not create final brand strategy, final voice, final visual identity, application rules, production asset guidance, or implementation tasks.

Valid brand evidence for this pass is limited to supplied brand material and explicit `.canon/brand` upstream artifacts. No prior `.canon/brand` artifacts were present in this workspace. Repository implementation files are excluded unless the brief or supplied site design note explicitly named them as current brand material.

Absolute filesystem paths supplied by the brief are evidence references for read-only inspection. They should not become downstream emission evidence pointers. Downstream emissions should use stable `SOURCE:` or `ASSET:` ids, `input:<relative-path>` pointers, or project-relative `.canon/brand/` paths.

## Source Inventory

| ID | Supplied source reference | Captured material | Source role | Source quality | Readability |
| --- | --- | --- | --- | --- | --- |
| `SOURCE:brief-product-summary` | `brand-input/brief.md` / Product Summary | Primkit is a family of local infrastructure primitives for AI agents and the humans working with them. It provides four small Go CLIs backed by embedded SQLite: `taskprim` for task lifecycle and dependency frontiers, `stateprim` for durable key-value state and append logs, `knowledgeprim` for searchable knowledge graphs, and `queueprim` for persistent work queues; public promise centers on session-surviving state, four CLIs, four SQLite files, and no server. | Strategy source | Authoritative | Direct text, readable. |
| `SOURCE:brief-current-evidence-map` | `brand-input/brief.md` / Current Brand Evidence | Declares the second source-map pass should treat the live `primkit-site` repository, site copy, terminal demos, robot assets, social preview, interim tokens, implemented CSS, and design note as current brand evidence. | Strategy source | Authoritative | Direct text, readable. |
| `SOURCE:brief-observed-site-language` | `brand-input/brief.md` / Observed current site language | Groups the supplied public-site language: "Give your agents state that survives the session.", "Install once, call from any shell or agent loop.", "The agent resets. The work shouldn't.", "Four tools, one agent loop.", "Explicit, not automatic. A file, not a fleet.", real terminal proof mode, trust strip, and "Stop re-explaining your project every session." | Verbal source | Strong evidence | Direct text, readable. |
| `SOURCE:brief-observed-visual-language` | `brand-input/brief.md` / Observed current visual language | Groups the supplied visual summary: blocky navy robot mark, cyan glow accents, dark navy social preview, white `primkit` wordmark, cyan tagline, robot-derived palette including `#181848`, `#14152b`, `#00d8f0`, `#0b7285`, `#fbfcfe`, `#f3f6fb`, and `#e9edf6`, and the `#0b7285` role as AA-safe accent text on light surfaces. | Visual source | Strong evidence | Direct text, readable. |
| `SOURCE:owner-answers-first-pass` | `brand-input/brief.md` / Owner Answers From First Source-Map Pass | Owner decisions and constraints across primary/secondary audiences, lead promise, durable/explicit personality, agent-native restraint, abstract/platform-y avoidance, admired references, anti-references, serious local infrastructure stance, balanced promise-plus-proof language, trust/relief/confidence feelings, website/docs priority, CLI-help priority, the four-question problem framing, signature phrases including "Composable like any Unix tool." and "Persistent state for AI agents.", rejected category labels, allowed precise terms, robot restraint, persistence-layer interpretation, and navy/cyan preservation. | Strategy source | Authoritative | Direct text, readable. |
| `SOURCE:source-map-instructions-this-pass` | `brand-input/brief.md` / Source-Map Instructions For This Pass | Run instructions: treat as archaeology mode; classify product facts and owner decisions as confirmed; current site behavior as de facto; production-ready identity gaps as missing; do not invent new visual directions. | Constraint source | Authoritative | Direct text, readable. |
| `SOURCE:primkit-site-repo-root` | `/Users/avergara/development/propifly/primkit-site` | Supplied absolute root for the sibling live site repository. Use only named site files below as brand evidence, not the whole implementation tree. | Application source | Reference only | Directory reference, readable as location only. |
| `SOURCE:site-landing-page` | `/Users/avergara/development/propifly/primkit-site/src/app/(home)/page.tsx` | Current landing-page presentation and copy: nav, hero, trust strip, problem framing, four-question block, composability section, Why-Primkit principles, docs cards, closing CTA, footer, and repeated mark usage. | Application source | Strong evidence | Direct text/code, readable. |
| `SOURCE:site-hero-terminal` | `/Users/avergara/development/propifly/primkit-site/src/app/(home)/hero-terminal.tsx` | Current tabbed terminal proof demos for all four primitives, including captured CLI transcripts, proof behavior, and accessibility/readability notes. | Application source | Strong evidence | Direct text/code, readable. |
| `SOURCE:site-navigation-hero-copy` | `/Users/avergara/development/propifly/primkit-site/docs-navigation.json` | Current docs navigation, landing hero title, tagline, CTA, and home-card titles/descriptions. | Verbal source | Strong evidence | Direct JSON, readable. |
| `SOURCE:site-design-note` | `/Users/avergara/development/propifly/primkit-site/.design/design.md` | Current site design source of truth: public docs renderer posture, content boundary, landing/docs relationship, brand-token need, logo/OG usage, asset boundary, and open brand-token question. | Constraint source | Strong evidence | Direct markdown, readable. |
| `SOURCE:site-design-tokens` | `/Users/avergara/development/propifly/primkit-site/.design/design-tokens.tokens.json` | Interim hand-authored DTCG-style token file: robot-derived palette, token values, cyan usage rules, light-mode status, provisional status, shadows, and font token slots. | Visual source | Strong evidence | Direct JSON, readable. |
| `SOURCE:site-token-css` | `/Users/avergara/development/propifly/primkit-site/src/app/globals.css` | Implemented CSS token mapping and Fumadocs theme aliases using `pk` variables; repeats provisional status and accessibility constraint for cyan usage. | Application source | Strong evidence | Direct CSS, readable. |
| `ASSET:site-robot-mark` | `/Users/avergara/development/propifly/primkit-site/public/logo.png` | Current public site robot mark used by the landing page header/footer. File metadata: PNG, 256 x 256, RGBA. Visual interpretation should rely on supplied brief/design text, not unsupported pixel inference. | Asset source | Strong evidence | Binary image; metadata readable, visual extraction partial. |
| `ASSET:site-social-preview` | `/Users/avergara/development/propifly/primkit-site/public/social-preview.png` | Current public site social/OG preview. File metadata: PNG, 1280 x 640, indexed color. Visual interpretation should rely on supplied brief/design text, not unsupported pixel inference. | Asset source | Strong evidence | Binary image; metadata readable, visual extraction partial. |
| `ASSET:product-logo` | `docs/assets/logo.png` | Product-repo logo asset named by the site design note as the source vendored into the site `public/` directory. File metadata: PNG, 256 x 256, RGBA. | Asset source | Reference only | Binary image; metadata readable, visual extraction partial. |
| `ASSET:product-social-preview` | `docs/assets/social-preview.png` | Product-repo social preview asset named by the site design note as the source vendored into the site `public/` directory. File metadata: PNG, 1280 x 640, indexed color. | Asset source | Reference only | Binary image; metadata readable, visual extraction partial. |
| `ASSET:product-demo-gif` | `docs/assets/demo.gif` | Product demo media named by the site design note as a README/product demo asset and optional landing-page reuse candidate. File metadata: GIF, 1200 x 600. | Asset source | Reference only | Binary image animation; metadata readable, content extraction partial. |
| `ASSET:product-demo-mp4` | `docs/assets/demo.mp4` | Product demo media named by the site design note as a README/product demo asset and optional landing-page reuse candidate. | Asset source | Reference only | Binary video; file type readable, content extraction partial. |
| `SOURCE:product-demo-tape` | `docs/assets/demo.tape` | VHS script for the product demo, including session-resilient state narrative, commands, outputs, dimensions, theme, and generation targets. | Verbal source | Reference only | Direct text, readable. |
| `SOURCE:product-demo-setup-script` | `docs/assets/demo-setup.sh` | Helper script for the VHS demo that stores structured migration context with `stateprim`. | Application source | Reference only | Direct shell text, readable. |
| `SOURCE:inspiration-sqlite` | Owner admired reference: SQLite | Durability, trust, and boring-in-the-best-way infrastructure. | Inspiration source | Reference only | Named reference only; external material not fetched. |
| `SOURCE:inspiration-git` | Owner admired reference: Git | Local-first inspectability and durable project memory. | Inspiration source | Reference only | Named reference only; external material not fetched. |
| `SOURCE:inspiration-ripgrep` | Owner admired reference: ripgrep | Sharp CLI utility and plain developer credibility. | Inspiration source | Reference only | Named reference only; external material not fetched. |
| `SOURCE:inspiration-tailscale-docs` | Owner admired reference: Tailscale docs | Practical technical clarity. | Inspiration source | Reference only | Named reference only; external material not fetched. |
| `SOURCE:inspiration-fly-io` | Owner admired reference: Fly.io | Confident infrastructure copy without enterprise weight. | Inspiration source | Reference only | Named reference only; external material not fetched. |
| `SOURCE:inspiration-unix-tool-philosophy` | Owner admired reference: Unix tool philosophy | Small tools that compose. | Inspiration source | Reference only | Named reference only; external material not fetched. |
| `SOURCE:rejected-hosted-ai-memory-platforms` | Owner anti-reference: hosted AI memory platforms | Rejected direction for positioning and category feel. | Rejected source | Rejected | Direct text, readable. |
| `SOURCE:rejected-agent-orchestration-platform` | Owner anti-reference: enterprise agent orchestration platform branding | Rejected direction for language and product category posture. | Rejected source | Rejected | Direct text, readable. |
| `SOURCE:rejected-cute-mascot-tools` | Owner anti-reference: cute mascot-first developer tools | Rejected direction for mark usage and voice. | Rejected source | Rejected | Direct text, readable. |
| `SOURCE:rejected-crypto-web3-aesthetics` | Owner anti-reference: crypto/web3 infrastructure aesthetics | Rejected direction for visual language and credibility cues. | Rejected source | Rejected | Direct text, readable. |
| `SOURCE:rejected-abstract-ai-productivity` | Owner anti-reference: abstract AI productivity tools promising magic instead of commands | Rejected direction for claims, abstraction level, and proof posture. | Rejected source | Rejected | Direct text, readable. |
| `SOURCE:rejected-generic-developer-saas` | Owner anti-reference: generic developer SaaS pages with inflated claims and vague diagrams | Rejected direction for page structure, copy density, and evidence style. | Rejected source | Rejected | Direct text, readable. |
| `SOURCE:rejected-primary-category-labels` | Owner rejected primary category labels | Rejected primary labels: platform, operating system, orchestration, framework, hosted memory, autonomous memory, AI workspace, and agent brain. | Rejected source | Rejected | Direct text, readable. |
| `SOURCE:rejected-mascot-led-voice` | Owner rejected direction: robot as mascot-led voice or cute character system | Rejected direction for turning the robot mark into a playful mascot identity. | Rejected source | Rejected | Direct text, readable. |
| `SOURCE:rejected-general-agent-framework` | Owner rejected direction: general agent framework | Rejected interpretation; Primkit should remain the persistence layer underneath agents. | Rejected source | Rejected | Direct text, readable. |
| `SOURCE:rejected-casual-visual-departure` | Owner rejected direction: casually discarding the navy/cyan system | Rejected direction unless downstream work gives a clear reason for intentional evolution. | Rejected source | Rejected | Direct text, readable. |

## Source Roles

Allowed role vocabulary used in this inventory:

- `Strategy source`: product facts, owner stance, mode instructions, and current evidence boundary.
- `Verbal source`: current copy, phrase inventory, demo language, tagline, navigation labels, and terminal proof text.
- `Visual source`: visual-system descriptions, token values, palette constraints, and visual summaries.
- `Asset source`: image, animation, video, and source-grade media files.
- `Application source`: current public site implementation files explicitly named by the brief or design note as brand evidence.
- `Constraint source`: rules, limits, provisional status, source-map instructions, and site design boundaries.
- `Inspiration source`: owner-named references to evaluate for direction generation without copying or canonizing them.
- `Rejected source`: owner-named anti-references, rejected labels, and directions downstream work should avoid.

Grouped rows are intentional where one supplied section contains many related references, such as owner answers, observed current site language, observed current visual language, rejected category labels, and the site evidence map. The grouped rows preserve the individual evidence items in the Captured material column.

## Source Quality

Allowed quality vocabulary used in this inventory:

- `Authoritative`: supplied owner or brief material that can directly govern source-map handling, product facts, or brand constraints.
- `Strong evidence`: current operating material from the live site or explicit current-site summaries; useful as de facto evidence, but not automatically final identity.
- `Weak evidence`: allowed value; not assigned in this inventory because every included source is either stronger current/owner evidence, reference-only material, rejected material, or binary material with limited extraction.
- `Reference only`: named context, inspiration, upstream asset references, or demo materials that should inform review without becoming direct brand authority.
- `Rejected`: owner-provided anti-references, forbidden primary labels, and directions not to follow.
- `Unknown`: allowed value; not assigned in this inventory because provenance limits are recorded in limitations without treating the source itself as unknown.

Quality is not the same as final decision status. For example, current site files are `Strong evidence` for operating behavior, while their identity decisions may still be provisional or de facto rather than confirmed. Likewise, inspiration references are owner-supplied but remain `Reference only` because no external source was inspected and no borrowed identity should be inferred.

## Readability

Directly readable text sources: `brand-input/brief.md`, the named site TSX files, `docs-navigation.json`, `.design/design.md`, `.design/design-tokens.tokens.json`, `globals.css`, `docs/assets/demo.tape`, and `docs/assets/demo-setup.sh`.

Partially readable binary sources: `public/logo.png`, `public/social-preview.png`, `docs/assets/logo.png`, `docs/assets/social-preview.png`, `docs/assets/demo.gif`, and `docs/assets/demo.mp4`. File type and dimensions were readable for images and GIFs; detailed image content, logo geometry, color correctness, animation content, and licensing/ownership were not inferred from pixels.

Named inspiration references were not fetched or audited externally. They are recorded only as owner-provided references that downstream direction work may use for calibration.

The requested contract filenames `brand-family-contract.md` and `brand-id-contract.md` were not present under this workspace or the surrounding Hatchable state tree during inspection. The role vocabulary, quality vocabulary, ID style, output scope, and artifact headings were applied from the supplied task/domain contract text.

## Limitations

- Absolute paths are supplied evidence references for this archaeology pass. They should not be copied into downstream decision emissions as evidence pointers.
- The sibling `primkit-site` files are current operating evidence, not final brand authority by default.
- The interim token file and CSS are explicitly provisional. They should be preserved or deliberately evolved, not silently treated as final production identity.
- Binary and image sources are inventoried even when extraction is partial. Visual claims should rely on supplied prose or explicit visual tooling, not unsupported pixel inference.
- The current inventory does not settle final voice, final visual identity, logo source ownership, font licensing, production asset packaging, or application rules.
- Rejected directions are inventoried as owner constraints. They are not gaps.
- Unrelated repository implementation details remain outside the brand evidence boundary unless a supplied source explicitly names them as brand material.
- This task is scoped to `source-inventory.md` only; companion artifacts such as `asset-inventory.md`, `source-quality.md`, `conflicts.md`, `gaps.md`, `decision-ledger.md`, and `downstream-brief.md` are intentionally not created here.
