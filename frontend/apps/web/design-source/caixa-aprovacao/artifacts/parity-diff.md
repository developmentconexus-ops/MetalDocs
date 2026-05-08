# Phase 3b — Parity Diff
> Screen: Caixa de Aprovação (`caixa-aprovacao`)
> Date: 2026-05-08
> Method: `preview_inspect` + `preview_eval` computed-style queries against running app at `http://localhost:4174/approvals`

---

## Region 1 — InboxToolbar

| Property | Spec | Computed | Match |
|---|---|---|---|
| `display` | `flex` | `flex` | ✓ |
| `align-items` | `center` | `center` | ✓ |
| `padding-left` | `24px` (var(--sp-6)) | `24px` | ✓ |
| `padding-right` | `24px` (var(--sp-6)) | `24px` | ✓ |
| `height` | `48px` | `48px` | ✓ |
| `background-color` | `rgb(255,255,255)` (--surface) | `rgb(255, 255, 255)` | ✓ |
| `flex-shrink` | `0` | `0` | ✓ |

**View switcher — active button (`Foco`)**

| Property | Spec | Computed | Match |
|---|---|---|---|
| `background-color` | `var(--surface)` = white | `rgb(255, 255, 255)` | ✓ |
| `color` | `var(--text)` = `#1a0e0e` | `rgb(26, 14, 14)` | ✓ |
| `box-shadow` | `var(--shadow-1)` | `rgba(74, 33, 33, 0.06) 0px 1px 2px 0px` | ✓ |

**View switcher — inactive button (`Linha do tempo`)**

| Property | Spec | Computed | Match |
|---|---|---|---|
| `background-color` | `transparent` | `rgba(0, 0, 0, 0)` | ✓ |
| `color` | `var(--text-muted)` = `#8a7575` | `rgb(138, 117, 117)` | ✓ |
| `box-shadow` | `none` | `none` | ✓ |

**View switcher container**

| Property | Spec | Computed | Match |
|---|---|---|---|
| `background-color` | `var(--surface-2)` | `rgb(250, 246, 246)` | ✓ |
| `border-radius` | `var(--r-2)` = `6px` | `6px` | ✓ |
| `padding` | `2px` | `2px` | ✓ |

---

## Region 2 — Queue Rail (active item)

| Property | Spec | Computed | Match |
|---|---|---|---|
| `display` | `flex` | `flex` | ✓ |
| `gap` | `var(--sp-3)` = `12px` | `12px` | ✓ |
| `padding` | `14px 24px` | `14px 24px` | ✓ |
| `border-left` | `3px solid var(--brand)` | `2.4px solid rgb(107, 31, 42)` | ~✓ (subpixel rendering) |
| `background-color` | `var(--surface-2)` = `#faf6f6` | `rgb(250, 246, 246)` | ✓ |

Note: `border-left` computes as `2.4px` due to subpixel scaling at the test DPR — spec value is `3px` which rounds to `2.4px` at 0.8× device scale. This is a rendering artefact, not a CSS error.

---

## Region 3 — InboxApprovalCard

| Property | Spec | Computed | Match |
|---|---|---|---|
| `border-radius` | `16px` (var(--r-4)) | `16px` | ✓ |
| `box-shadow` | `0 24px 60px rgba(0,0,0,0.12), 0 4px 12px rgba(0,0,0,0.04)` | matches | ✓ |
| `border` | `1px solid var(--border)` | `0.8px solid rgb(230, 220, 220)` | ~✓ (subpixel) |
| `width` | `min(640px, 90%)` | `576px` (90% of 640px viewport) | ✓ |

**Card header**

| Property | Spec | Computed | Match |
|---|---|---|---|
| `background-color` | `var(--text)` = `#1a0e0e` | `rgb(26, 14, 14)` | ✓ |
| `color` | `var(--text-on-brand)` = `#ffffff` | `rgb(255, 255, 255)` | ✓ |
| `padding` | `var(--sp-5) 28px` = `20px 28px` | `20px 28px` | ✓ |

---

## Region 4 — Stats Grid

| Property | Spec | Computed | Match |
|---|---|---|---|
| `display` | `grid` | `grid` | ✓ |
| `grid-template-columns` | `1fr 1fr 1fr` | `162.125px 162.137px 162.137px` (equal thirds) | ✓ |
| `gap` | `var(--sp-4)` = `16px` | `16px` | ✓ |
| `border-top` | `1px solid var(--border)` | `0.8px solid rgb(230, 220, 220)` | ~✓ (subpixel) |

---

## Region 5 — InboxTimeline (verified via compiled CSS sheet, not live DOM — timeline view has static `view='stack'` in Phase 3a placeholder)

All timeline CSS rules verified via `document.styleSheets` enumeration. Token references confirmed preserved as `var(--token)` in compiled output.

| Class | Key rules | Token usage |
|---|---|---|
| `.timelineContainer` | `flex: 1; overflow: auto; background: var(--bg)` | ✓ |
| `.timelineTitle` | `font-size: 1.75rem; font-weight: 600` | ✓ |
| `.timelineRail` | `background: linear-gradient(180deg, var(--accent)…var(--text-faint))` | ✓ |
| `.bucketDotUrgent` | `background: var(--accent); border: 4px solid var(--bg)` | ✓ |
| `.bucketCount` | `background: var(--text); border-radius: var(--r-5)` | ✓ |
| `.bucketCountUrgent` | `background: var(--accent)` | ✓ |
| `.itemRow` | `display: grid; grid-template-columns: 80px 1fr auto 200px auto` | ✓ |
| `.stageBarFilled` | `background: var(--brand)` | ✓ |
| `.stageBarUrgent` | `background: var(--accent)` | ✓ |

Visual screenshot of timeline view taken and confirmed via HMR temp-patch (`view='timeline'`): layout, heatmap, bucket sections, stage bars, and "Revisar" buttons all render correctly.

---

## Summary

| Region | Status | Notes |
|---|---|---|
| InboxToolbar | PASS | All values match spec exactly |
| Queue rail (active item) | PASS | 3px border renders as 2.4px at 0.8× DPR — not a CSS error |
| InboxApprovalCard | PASS | 16px radius, correct shadow, correct header dark bg |
| Stats grid | PASS | Equal thirds, 16px gap, correct borders |
| InboxTimeline CSS | PASS | All token references verified from compiled stylesheet |

**Overall: PASS — no token violations detected, no raw hex, no spacing regressions.**
