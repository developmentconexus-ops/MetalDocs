# Parity Diff — distribuicao Phase 3b

**Date:** 2026-05-08
**Viewport measured:** 1440×900
**Route:** `/documents/PR-EHS-014/distribution`

Computed-style observations compared to design reference (`onda1-fanout-v5.jsx`).

| Region | Field | Impl value | Note |
|--------|-------|------------|------|
| Hero | padding | `32px 56px 40px` | Matches design exactly |
| Hero | background | `linear-gradient(#f9f0f0 → #fff)` | Correct brand-pale → surface |
| Hero | heroGrid columns | `210px 1006.8px` | Matches `210px 1fr` at 1440px |
| Hero | heroGrid gap | `40px` | Matches `var(--sp-8)` |
| Hero breadcrumb | font-size | `11px` | Matches `var(--font-size-xs)` |
| Hero breadcrumb | text-transform | `uppercase` | Correct |
| Hero .codeBadge | height | `24px` | Matches design |
| Hero .deadlineBadge | background | `var(--warning-bg)` | Correct |
| Hero title | font-size | `36px` | Matches `var(--font-size-2xl)` |
| Hero CTAs | opacity | `0.5` | Correct disabled state |
| DocRefCard | width/height | `168px / 224px` | Matches design |
| DocRefCard | transform | `perspective(1200px) rotateY(-12deg) rotateX(4deg)` | Matches design |
| DocRefCard | box-shadow | `var(--shadow-doc-card)` | Correct |
| DocRefCard headerStrip | background | `linear-gradient(135deg, var(--brand) → var(--brand-deep))` | Correct |
| KPIStrip | grid | `repeat(5, 1fr)` | Correct |
| KPIStrip | cell padding | `18px 20px` | Matches design (sp-5=20px used for horizontal) |
| KPIStrip .value | font-size | `22px` | Matches `var(--font-size-lg)` |
| KPIStrip .valueDanger | color | `var(--danger)` | Correct |
| KPIStrip .progressFill | background | `var(--success)` | Correct |
| DonutCard | padding | `24px` | Matches `var(--sp-6)` |
| DonutCard .donutPct | font-size | `32px` | Matches `var(--font-size-xl)` |
| DonutCard .legendDotRead | opacity | `0.4` | Matches design `opacity: faded ? 0.4 : 1` |
| DonutCard footer | border-top | `1px solid var(--border)` | Correct |
| DistributionFacts | card bg | `var(--surface)` | Correct |
| DistributionFacts .iconBoxWarning | background | `var(--warning-bg)` | Correct |
| DistributionFacts .factValue | font-size | `12.5px` | Matches `var(--font-size-sm)` |
| CoverageByArea | areaRow grid | `180px 1fr 100px` | Matches design |
| CoverageByArea .barRead | opacity | `0.22` | Matches design |
| CoverageByArea .goalMarker | left | `92%` | Correct goal line position |
| CoverageByArea .pctValue | font-size | `14px` | Matches `var(--font-size-md)` |
| TimelineCard | padding | `20px 20px` | Matches design `20px 22px` (sp-5 = 20px, minor rounding) |
| TimelineCard .legendLine | width/height | `16px / 2px` | Correct |
| RecipientsCard .filterBar | background | `var(--surface-2)` | Correct |
| RecipientsCard .headerRow | font-size | `9.5px` | Matches `var(--font-size-2xs)` |
| RecipientsCard .headerRow | columns | `32px 1.6fr 1.1fr 1fr 130px 130px 90px` | Matches design |
| RecipientsCard .statusAck | background | `var(--success-bg)` | Correct |
| RecipientsCard .statusOverdue | background | `var(--danger-bg)` | Correct |
| RecipientsCard .paginationFooter | font-size | `11px` | Matches `var(--font-size-xs)` |
| RecipientsCard SearchBar | border leakage | `0.8px solid var(--border)` | Pre-existing SearchBar issue; see leakage-probe.md |
| Page .page | height | `min-height: 100%` | Adjusted from `height: 100%` — prevents hero flex-collapse in shell context |
| Hero overlay | grid pattern | `linear-gradient brand 1px, transparent; 32px 32px` | Correct CSS background grid |

## Layout adjustment made

The design reference uses `height: 1200` (explicit fixed height) on the outermost container. In the real shell, `.page` is nested inside a `flex: 1; overflow: auto` main. Using `height: 100%` caused the hero to flex-shrink to ~72px. Fixed by changing `.page` to `min-height: 100%` and adding `flex-shrink: 0` to `.hero`. All other dimensions match the reference.
