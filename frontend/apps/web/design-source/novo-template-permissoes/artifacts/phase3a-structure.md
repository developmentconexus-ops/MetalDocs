# Phase 3a — Structure mirror · novo-template-permissoes

**Executor:** Main agent inline (Heavy tier)

## DOM diff vs design reference

| Level | Design | Impl | Match? |
|---|---|---|---|
| Root | `.card` div | `<div className="card">` | ✓ |
| L1 | `.kicker` div | `<div className="kicker">` | ✓ |
| L1 | `h2.h2` | `<h2 className="h2">` | ✓ |
| L1 | `p.caption` | `<p className="caption {intro}">` | ✓ |
| L1 | `div[role=tablist]` (mode segmented) | `<div role="radiogroup" className={modeSegmented}>` | ✓ (role fixed to radiogroup — mutex, not tablist) |
| L2 | `button[role=tab] × 3` | `<button role="radio" aria-checked × 3>` | ✓ (radio semantics for mutex) |
| L3 | span label + span tiny desc | `<span className={modeTabLabel}> + <span className="tiny {modeTabDesc}>` | ✓ |
| L1 (all) | `div` banner | `<div className={allBanner}>` | ✓ |
| L2 | `span` icon circle | `<span className={allIcon}> <Icon name="home"/>` | ✓ |
| L2 | `div` body | `<div className={allBody}>` | ✓ |
| L3 | div title + p.caption | `<div className={allTitle}> + <p className="caption">` | ✓ |
| L1 (areas) | `div` 3-col grid | `<div role="group" className={areaGrid}>` | ✓ |
| L2 | `button × 6` | `<button role="checkbox" aria-checked × 6>` | ✓ |
| L3 | `span` icon | `<span className={cardIcon}>` | ✓ |
| L3 | `div` body | `<div className={cardBody}>` | ✓ |
| L4 | mono code + tiny count | `<div className="mono {cardCode}"> + <div className="tiny {cardCount}>` | ✓ |
| L3 | `span` check icon (conditional) | `<span className={cardCheck}> <Icon name="check"/>` | ✓ |
| L1 (roles) | `div` 2-col grid | `<div role="group" className={roleGrid}>` | ✓ |
| L2 | `button × 6` | `<button role="checkbox" aria-checked × 6>` | ✓ |
| L3 | `span` icon | `<span className={cardIcon}>` | ✓ |
| L3 | `div` body | `<div className={cardBody}>` | ✓ |
| L4 | `div.roleIdRow { mono id · tiny area }` | `<div className={roleIdRow}> <span mono> <span tiny>` | ✓ |
| L4 | `div` name + `div.tiny` count | `<div className={cardName}> + <div className="tiny {cardCount}>` | ✓ |
| L3 | `span` check icon (conditional) | `<span className={cardCheck}>` | ✓ |
| L1 | coverage summary (mode !== 'all') | `<div className={coverageSummary}>` | ✓ |
| L2 | `.kicker` + coverage row | `<div className="kicker {coverageKicker}"> + <div className={coverageRow}>` | ✓ |
| L3 | `span.mono` count + `span.caption` text | `<span className="mono {coverageCount}"> + <span className="caption">` | ✓ |
| L1 | `WizardFooter` | `<WizardFooter .../>` | ✓ |

## Notes

- `role="tablist"` → `role="radiogroup"` on mode control (mutex selection → radio semantics, matches Step 3 pattern).
- `role="group"` + `role="checkbox"` + `aria-checked` on card grids (multi-select, not mutex).
- DOM order preserved exactly per design.
- No semantic restructuring.

## Status

Structure mirrors design. Phase 3b (style) dispatched as subagent.
