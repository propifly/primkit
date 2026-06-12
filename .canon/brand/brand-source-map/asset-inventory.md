---
hatchable:
  domain: brand-source-map
  root: canon
  lane: brand
  artifact: asset-inventory
  role: primary
  status: draft
---

# Asset Inventory

## Asset Boundary

This asset inventory follows the source boundary established in `.canon/brand/brand-source-map/source-inventory.md`.

The supplied source set contains one authoritative prose intake brief, `SOURCE:primkit-brand-brief`, and no operating material or visual asset files. Repository implementation files, package manifests, route structure, local documentation, and existing repo assets are outside the source boundary and are not treated as brand evidence.

No `ASSET:` ids are created in this inventory because no approved source-grade assets, export-only assets, screenshots, inspiration references, or production visual files were supplied.

## Source-Grade Assets

No source-grade assets were supplied.

| Asset Category | Inventory Status | Provenance |
| --- | --- | --- |
| Logo source files | Missing | No approved vector, editable, or master logo source file is present in the supplied source set. |
| Brand marks and lockups | Missing | No approved mark, wordmark, lockup, symbol, or responsive logo system is present in the supplied source set. |
| Color source files | Missing | No approved palette file, design token file, swatch file, or color specification is present in the supplied source set. |
| Typography source files | Missing | No approved font files, font licenses, type scale, or typography specification is present in the supplied source set. |
| Iconography source files | Missing | No approved icon set, icon source file, or icon usage rule is present in the supplied source set. |
| Illustration or photography | Missing | No approved illustration, photography, image library, or art-direction source is present in the supplied source set. |
| Design-system or template sources | Missing | No approved editable design file, template, component library, or production layout source is present in the supplied source set. |

The prose brief provides directional brand constraints, but it does not approve any logo geometry, color values, typography, iconography, imagery, or production layout assets.

## Export-Only Assets

No export-only assets were supplied.

The supplied material contains no screenshots, rendered logo exports, raster images, PDFs, decks, social graphics, website captures, product captures, exported icons, or application mockups.

Screenshots and inspiration references would be visual evidence only, not approved source assets, unless an owner explicitly marked them as production-approved source material. No such screenshot or inspiration material is present in this run.

## Missing Obvious Variants

Because no approved master asset exists in the supplied source set, the obvious production variants are absent rather than incomplete variants of a known system.

Missing or unconfirmed variants include:

- Primary logo, secondary logo, wordmark, symbol, and responsive lockups.
- Light-background, dark-background, single-color, reversed, and small-size logo variants.
- Favicon, app icon, social avatar, repository icon, and documentation icon variants.
- Approved color palette, neutral palette, semantic colors, accessibility pairings, and color tokens.
- Approved typefaces, type scale, font licensing, fallback stack, and usage rules.
- Approved icon style, illustration style, photography style, screenshot style, and diagram style.
- Approved templates for website, documentation, slide, social, and product-interface uses.

These absences are not visual direction. They are asset-boundary facts for downstream work to resolve.

## Unknown Provenance

Any visual material outside the supplied source set has unknown provenance for this run.

That includes possible existing repository images, icons, interface styling, screenshots, documentation examples, package badges, or implementation assets. They are not inventoried as brand assets because the source boundary excludes repository implementation files and existing repo assets from brand evidence.

Unknown-provenance material should not be treated as approved identity, source-grade artwork, or production-ready visual guidance without owner confirmation and downstream visual identity review.

## Downstream Visual Checks

Future production use of visuals, color, typography, logo, or iconography should route to downstream visual identity work.

Before production visual use, downstream work should verify or create:

- Approved logo source files and usage rules.
- Approved color palette, accessibility checks, and production token mapping.
- Approved typography, font licensing, fallback stack, and type scale.
- Approved iconography, illustration, photography, screenshot, and diagram direction.
- Approved export formats, file naming, versioning, and ownership/provenance records.
- Clear distinction between source-grade assets, export-only references, inspiration references, and rejected visual directions.

Until those checks are completed, production visual decisions for Primkit should be treated as missing or unknown, not source-backed.
