# Phase 3b — Style Port: distribuicao

**Date:** 2026-05-08
**Status:** Complete

## Tokens added

Added to `frontend/apps/web/src/styles/tokens.css` under a new `/* Font sizes */` section:

| Token | Value | Usage |
|-------|-------|-------|
| `--font-size-2xs` | `9.5px` | Kicker text, breadcrumb, DocRefCard header, KPI kicker, col headers |
| `--font-size-xs` | `11px` | Secondary labels, subtitles, pagination, deadline badge |
| `--font-size-sm` | `12.5px` | Fact values, legend rows |
| `--font-size-md` | `14px` | CoverageByArea pct values, hero subtitle approximation |
| `--font-size-lg` | `22px` | KPI cell values |
| `--font-size-xl` | `32px` | DonutCard center percentage |
| `--font-size-2xl` | `36px` | Hero H1 title |

No color, spacing, radius, or shadow tokens were missing — all those used `var(--token)` references directly from the existing token set.

## Leakage fixes applied

1. **Page height collapse** — `.page` changed from `height: 100%` to `min-height: 100%`. The shell's `<main>` is `overflow: auto` with flex layout; `height: 100%` caused `.hero` to be crushed to ~72px by the `flex: 1` `.main`. `min-height: 100%` allows the page to grow beyond the viewport while still filling it.

2. **Hero flex-shrink** — Added `flex-shrink: 0` to `.hero` to prevent it from being squeezed by the flex container even if the page total is constrained.

3. **RecipientsCard search input** — Phase 3a uses the `SearchBar` primitive (not a raw `<input>`), so no explicit reset was needed in `RecipientsCard.module.css`. A pre-existing leakage in `SearchBar.module.css` (global rule wins by specificity) is documented in `leakage-probe.md` but not fixed here — it is a cross-cutting concern for the SearchBar component, not specific to distribuicao.

## Notes per region

### Hero
- DocRefCard is rendered centered inside a `div` wrapper (Phase 3a structure); the `heroGrid` column of 210px accommodates the card + centering.
- The heroOverlay uses CSS `background-image` with two `linear-gradient` calls to create the grid dot pattern, matching the design's `backgroundImage` approach.
- Hero title uses `text-wrap: balance` for multi-line balance (matches design).

### KPIStrip
- The `.kicker` class inside KPIStrip.module.css overrides only `font-size`; the global `.kicker` class (from base.css) handles the rest (color, font-family, letter-spacing, text-transform).
- Progress bar shows fixed 62.9% (mock data in Phase 3a TSX).

### DonutCard
- SVG is static (no animation in Phase 3a skeleton) — this is expected. The ring circles use inline `var(--surface-3)`, `var(--brand)`, `var(--success)` directly on SVG attributes, which is correct.
- The `.donutSvg` has `transform: rotate(-90deg)` to start arcs from the top.

### CoverageByArea
- Bar colors are statically `var(--success)` for `.barAck` in CSS. The design dynamically colors based on pct threshold (success/brand/warning/danger). Phase 3a uses static colors — dynamic coloring is a Phase 4 behavior concern, not a style port issue.
- Goal marker at `left: 92%` matches the design's 92% target line.

### TimelineCard
- SVG chart is fully rendered with static data (Phase 3a). No animation needed for static skeleton.
- The `.legendLine` (solid) and `.legendDash` (dashed) chips correctly represent the two series.

### RecipientsCard
- Uses `TabBar` and `SearchBar` primitives for filter controls (Phase 3a decision). These primitives have their own module CSS — no style conflicts detected.
- Status pills (`statusAck`, `statusRead`, `statusPending`, `statusOverdue`) use semantic tokens matching `STATUS_TONE` from the design.

## Screenshots
Saved at `artifacts/screenshots/`:
- `1440-impl.png` — 1440×900 viewport
- `1024-impl.png` — 1024×768 viewport
- `375-impl.png` — 375×812 viewport

## User approval
<!-- Do NOT mark "User approved screenshot diff" — leave that for the user -->
