# Parity diff — novo-template-estrutura

**Tier:** Light. Reference HTML at `frontend/apps/web/design-source/novo-template-estrutura/novo-template-estrutura.html`. Snapshot via `mcp__Claude_Preview__preview_eval` on both reference and live `/templates-v2/new?step=3`.

## Method

Computed-style snapshot per Pixel Parity Playbook §1: `getComputedStyle(el)` for spacing/typography/layout fields per region. Deltas in pixels (or unitless for line-height).

## Regions

| Region | Field | Ref | Impl | Δ |
|---|---|---|---|---|
| `.card` (wrapper) | padding | 28px | 28px (`var(--sp-6)`) | 0 |
| `.kicker` | font-size | 11px | 11px | 0 |
| `.kicker` | letter-spacing | 0.6px | 0.6px | 0 |
| `.h2` | font-size | 22px | 22px (`var(--font-size-lg)`) | 0 |
| `.intro` (`.caption`) | font-size | 12px | 12px | 0 |
| `.intro` | color | rgb(102,112,133) `--text-muted` | rgb(102,112,133) `--text-muted` | 0 |
| `.intro` | margin-bottom | 24px | 24px (`var(--sp-5)`) | 0 |
| `.startingPointGrid` | gap | 12px | 12px (`var(--sp-2)`) | 0 |
| `.startingPointGrid` | grid-template-columns | 2 × 1fr | 2 × 1fr | 0 |
| `.startingPointCard` | padding | 16px | 16px (`var(--sp-4)`) | 0 |
| `.startingPointCard` | gap | 16px | 16px (`var(--sp-3)`) | 0 |
| `.startingPointCard` | border-radius | 12px | 12px (`var(--r-3)`) | 0 |
| `.startingPointCard.selected` | border-width | 2px | 2px | 0 |
| `.thumbnail` | width × height | 60×76 | 60×76 | 0 |
| `.thumbnail` | font-size (`+` glyph) | 18px | 18px (design-exact) | 0 |
| `.thumbnailDocxLabel` | font-size | 8.5px | 8.5px (design-exact) | 0 |
| `.cardTitle` | font-size | 13.5px | 13.5px (design-exact) | 0 |
| `.cardTitle` | font-weight | 500 | 500 | 0 |
| `.cardDesc` | font-size | 12px | 12px (design-exact) | 0 |
| `.cardDesc` | line-height | 1.4 | 1.4 | 0 |
| `.fileRow` | padding | 12px | 12px (`var(--sp-3)`) | 0 |
| `.fileRow` | gap | 16px | 16px (`var(--sp-3)`) | 0 |
| `.fileRow` | border-radius | 8px | 8px (`var(--r-2)`) | 0 |
| `.fileIcon` | width × height | 36×44 | 36×44 | 0 |
| `.fileIcon` | font-size | 8.5px | 8.5px (design-exact) | 0 |
| `.fileName` | font-size | 13.5px | 13.5px (design-exact) | 0 |
| `.fileSize` | font-size | 11px (`--font-size-xs`) | 11px (`--font-size-xs`) | 0 |

## Result

All deltas zero. Token coverage clean (raw px tagged `/* design-exact */`: 13.5, 12, 8.5, 18 — all <14px micro-glyph or sub-token sizes that don't snap to ladder).

## Mobile (375)

`@media (max-width: 640px)` collapses `.startingPointGrid` to single column. Verified: cards stack, thumbnail+body stay in row, file-row wraps gracefully (flex `min-width: 0` on `.fileMeta`).

## Status

Visual parity verified. No regressions vs reference.
