# Branding

The company logo of **Apotek Erza**: where it comes from, how to use it, and the brand colours. The source files live in `source/` and are the **only** source of the logo.

## Source files

| File | Format | Use |
|---|---|---|
| `source/logo.svg` | Vector, `viewBox="0 0 1254 1254"`, transparent background, 3 gradients | **Preferred** everywhere: app header, login page, favicon source, printed documents |
| `source/logo-mark.png` | Raster 1254 × 1254, RGBA, **transparent background** | Raster use on any background (dark, coloured, photos), app icons, share images |

The mark: a mortar and pestle held by two hands, three capsules above, and a green "Rx" inside the bowl. The mark has no wordmark; the name "Apotek Erza" is set as text next to it.

## Rules

1. **`source/` is the single source.** Never redraw, recolour, crop, trace, or re-export the logo. Do not create another copy as a second source of truth.
2. **Consume by copy at build time.** Application code uses copies placed in `apps/web/public/brand/` (and `deploy/` if needed) produced from `source/` by a build step. Copies are generated, not edited, and are listed in `.gitignore` if the build produces them. Decide the mechanism in the phase that builds the app shell (see `source/start-setup.md`, prompt 1.3).
3. **Prefer the SVG.** Use `logo-mark.png` only where a raster is required. Both files have a transparent background, so check legibility on dark backgrounds before using them there.
4. **Keep proportions** (square mark). Leave clear space around it of at least one capsule height. Do not stretch or add effects.
5. **Derived assets** (favicon, PWA icons, receipt logo) are derived from `source/logo.svg`. A receipt printed on a thermal printer needs a monochrome version: derive it from the SVG by a documented step and never by hand-editing.
6. The name text uses the application font, not a logo-style font. "Apotek Erza" is Indonesian user-facing text, so it is not translated.

## Brand colours (taken from `source/logo.svg`)

| Role | Start | End | Where in the mark |
|---|---|---|---|
| Blue (primary) | `#08B8F2` | `#0878F4` | Mortar, pestle, hands |
| Pink (accent) | `#FF168C` | `#E90083` | Capsules |
| Green (accent) | `#65E900` | `#19B900` | "Rx" |

Use these as design tokens in the Tailwind theme (`brand.blue`, `brand.pink`, `brand.green`), with the flat colour being the gradient start (to be confirmed with the designer).

**Accessibility note (contrast against white, computed):** `#08B8F2` 2.3:1, `#0878F4` 4.2:1, `#FF168C` 3.7:1, `#E90083` 4.4:1, `#65E900` 1.6:1, `#19B900` 2.6:1. WCAG AA for normal text needs 4.5:1, so none of the brand colours is safe for small text on white as is, and white text on `#08B8F2` or `#65E900` fails. Derive a darker blue (and a darker pink if used for text) for buttons and text when the app shell is built, check contrast, and record the chosen tokens here.

## Open items

- **Raster vs vector:** `logo-mark.png` is not pixel-identical to `logo.svg` (about 0.5% of pixels differ along edges and small details), so prefer the SVG when exactness matters or the logo is shown large.
- A monochrome version (for thermal receipts) is not provided. Derive it from the SVG when first needed.
- The SVG `<title>` reads "Caring Hands Pharmacy Rx", not "Apotek Erza". The file is never edited here; the app sets its own accessible name ("Apotek Erza") on the element that shows the logo.
- Printed documents (receipt, drug label, purchase order) show the company (pharmacy) name and APJ details from `masterdata`, plus the logo from `source/logo.svg`.
