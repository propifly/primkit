---
hatchable:
  domain: brand-source-map
  root: canon
  lane: brand
  artifact: asset-inventory
  role: primary
  status: draft
---

# Primkit Asset Inventory

## Asset Boundary

This inventory follows the evidence boundary in `.canon/brand/brand-source-map/source-inventory.md`. It catalogs supplied asset evidence only; it does not approve final visual identity, logo usage, production packaging, or application rules.

Stable asset ids come from the source inventory:

| Asset ID | Supplied evidence | Classification | Provenance status |
| --- | --- | --- | --- |
| `ASSET:site-robot-mark` | Current site `public/logo.png`; PNG, 256 x 256, RGBA. | Export-only | Supplied current-site evidence. It is a presentation-owned site asset and matches the product logo export by checksum, but no editable master, ownership record, or license record was supplied. |
| `ASSET:site-social-preview` | Current site `public/social-preview.png`; PNG, 1280 x 640, indexed color. | Export-only | Supplied current-site evidence. It is a presentation-owned site asset and matches the product social-preview export by checksum, but no editable master, ownership record, or license record was supplied. |
| `ASSET:product-logo` | `docs/assets/logo.png`; PNG, 256 x 256, RGBA. | Export-only | Product-repo raster export named by `SOURCE:site-design-note` as the source vendored into the site. It is upstream of the site copy, but still not source-grade artwork. |
| `ASSET:product-social-preview` | `docs/assets/social-preview.png`; PNG, 1280 x 640, indexed color. | Export-only | Product-repo raster export named by `SOURCE:site-design-note` as the source vendored into the site. It is upstream of the site copy, but still not source-grade artwork. |
| `ASSET:product-demo-gif` | `docs/assets/demo.gif`; GIF, 1200 x 600. | Export-only | Product demo media and proof material. It can support current product presentation, but it is not approved brand identity source material. |
| `ASSET:product-demo-mp4` | `docs/assets/demo.mp4`; MP4 media file. | Export-only | Product demo media and proof material. It can support current product presentation, but it is not approved brand identity source material. |

The visual descriptions in `SOURCE:brief-observed-visual-language`, `SOURCE:owner-answers-first-pass`, and `SOURCE:site-design-note` are valid evidence for what the current assets are meant to represent. Pixel-level logo geometry, color correctness, accessibility, and font licensing are not inferred from the image files.

Screenshots, PNGs, GIFs, MP4s, and inspiration material are not treated as approved source assets. They are evidence or exports unless a source-grade file, ownership record, and owner approval are supplied.

## Source-Grade Assets

No supplied asset qualifies as source-grade for production visual identity work.

| Needed source-grade asset | Classification | Evidence | Why it is not source-grade yet |
| --- | --- | --- | --- |
| Editable robot mark master | Missing | `ASSET:site-robot-mark`, `ASSET:product-logo`, `SOURCE:brief-observed-visual-language` | Only raster PNG exports are supplied. No SVG, vector master, layered file, design-file source, construction rules, clearspace, minimum-size rule, or approval record is present. |
| Editable social-preview master | Missing | `ASSET:site-social-preview`, `ASSET:product-social-preview`, `SOURCE:brief-observed-visual-language` | Only raster PNG exports are supplied. No editable layout source, crop system, template, or production export recipe is present. |
| Wordmark or logo lockup source | Missing | `ASSET:site-social-preview`, `SOURCE:site-landing-page`, `SOURCE:site-design-note` | The word `primkit` appears in current presentation contexts, but no wordmark source, lockup file, or typography approval is supplied. |
| Production color-token package | Missing | `SOURCE:site-design-tokens`, `SOURCE:site-token-css`, `SOURCE:owner-answers-first-pass` | Current navy/cyan values are strong de facto evidence, but the token source is explicitly provisional and downstream work must decide whether to preserve or deliberately evolve it. |
| Typography source and licensing | Missing | `SOURCE:site-design-tokens`, `SOURCE:site-token-css` | Font token slots and implemented CSS are evidence of current presentation, but no approved typeface decision, license record, or fallback policy is supplied. |
| Icon, illustration, and diagram source system | Missing | `SOURCE:site-landing-page`, `SOURCE:owner-answers-first-pass` | Current site icon usage and the robot mark are evidence, but no approved iconography, illustration, screenshot, or diagram style source package is supplied. |

## Export-Only Assets

The supplied asset files are export-only evidence. They can be used to understand the current public presentation, but they should not be treated as editable or fully approved production identity sources.

| Asset ID | Export role | Current evidence value | Production implication |
| --- | --- | --- | --- |
| `ASSET:site-robot-mark` | Current public site mark used as the restrained technical robot mark. | Strong evidence of current mark usage and the navy/cyan visual system. | Keep as de facto evidence for downstream visual identity. Do not ship broader production identity from this PNG alone. |
| `ASSET:site-social-preview` | Current public site social/OG preview. | Strong evidence of current social presentation: robot asset, dark navy surface, `primkit` name, and cyan tagline as described in the brief. | Use as evidence for downstream social template design. Confirm editable source, crops, text rendering, and ownership before production expansion. |
| `ASSET:product-logo` | Product-repo logo export used as the source for the current site mark copy. | Reference-grade upstream export; checksum matches `ASSET:site-robot-mark`. | Treat as the product-side copy of the same export, not as a master logo source. |
| `ASSET:product-social-preview` | Product-repo social-preview export used as the source for the current site preview copy. | Reference-grade upstream export; checksum matches `ASSET:site-social-preview`. | Treat as the product-side copy of the same export, not as a master social template. |
| `ASSET:product-demo-gif` | README/product demo animation. | Reference-grade proof media for the product's CLI workflow. | Can inform product proof presentation, but should not drive logo, palette, typography, or identity decisions. |
| `ASSET:product-demo-mp4` | README/product demo video. | Reference-grade proof media for the product's CLI workflow. | Can inform product proof presentation, but should not drive logo, palette, typography, or identity decisions. |

## Missing Obvious Variants

Missing or unconfirmed production variants:

- Editable logo master: SVG/vector, layered source, or design-file source.
- Mark-only, wordmark-only, horizontal lockup, stacked lockup, and compact lockup variants.
- Light-background, dark-background, reversed, single-color, and small-size logo variants.
- Favicon, app icon, repository avatar, social avatar, and docs-site icon variants with source provenance.
- Social-preview source template plus square, wide, cropped, and text-safe variants.
- Production color palette, semantic colors, dark-mode behavior, accessibility pairings, and final token mapping.
- Typography source decisions: display, UI, mono, fallback stack, font files or hosted-font source, and license records.
- Iconography, diagram, screenshot, illustration, and product-demo media rules.
- Export naming, file organization, versioning, checksums, and approval records for production release.

These are missing asset-package facts, not proposed visual directions.

## Unknown Provenance

The current robot mark and social preview have useful operating evidence but incomplete production provenance.

| Provenance question | Status | Evidence affected | Downstream owner |
| --- | --- | --- | --- |
| Who created or approved the robot mark? | Unknown provenance | `ASSET:site-robot-mark`, `ASSET:product-logo` | `brand-visual-identity` |
| Where is the editable source for the robot mark? | Missing | `ASSET:site-robot-mark`, `ASSET:product-logo` | `brand-visual-identity` |
| Who created or approved the social preview? | Unknown provenance | `ASSET:site-social-preview`, `ASSET:product-social-preview` | `brand-visual-identity` |
| Where is the editable source or template for the social preview? | Missing | `ASSET:site-social-preview`, `ASSET:product-social-preview` | `brand-visual-identity` |
| Are the current font choices licensed and approved? | Unknown provenance | `SOURCE:site-design-tokens`, `SOURCE:site-token-css` | `brand-visual-identity` |
| Are the current PNG colors accurate to intended brand values? | Unknown provenance | `ASSET:site-robot-mark`, `ASSET:site-social-preview`, `SOURCE:site-design-tokens` | `brand-visual-identity` |
| Are product demo GIF/MP4 assets approved for marketing reuse? | Unknown provenance | `ASSET:product-demo-gif`, `ASSET:product-demo-mp4` | `brand-visual-identity` |

Inspiration references and rejected directions are not asset provenance. They calibrate downstream direction work but do not approve copied visual material or source files.

## Downstream Visual Checks

Before production visual identity work, downstream domains should verify:

- Source-grade mark files exist, are editable, and are approved for production use.
- The robot mark remains a restrained technical mark rather than becoming mascot-led.
- The existing navy/cyan system is either preserved or deliberately evolved with a clear reason.
- Raster exports match approved source files after color-management and size checks.
- Logo variants work at favicon, nav, README, docs, social, and dark/light surface sizes.
- Social-preview templates have safe crop zones, readable text, export recipes, and current messaging.
- Color tokens pass accessibility checks, especially the `#0b7285` AA-safe accent-text role on light surfaces.
- Typography choices have clear source, license, loading, fallback, and usage rules.
- Demo media has source scripts, replay instructions, poster frames, captions or alternatives, and approval for website reuse.
- Production files distinguish source-grade assets, export-only files, reference media, screenshots, inspiration references, and rejected visual directions.

Until those checks are complete, the supplied assets should be treated as current operating evidence and export-only references, not as a complete production visual identity package.
