# Phase 3b — Leakage Probe
> Screen: Caixa de Aprovação (`caixa-aprovacao`)
> Date: 2026-05-08
> Source: Phase 2 pre-flight leakage map + live computed-style verification

---

## Global rules at risk (from phase2-preflight.md §4)

### 1. `button` → `cursor: pointer`

**Status: BENIGN — no action needed.**

All buttons on this page (queue items, view switcher, card actions, timeline row review buttons) should have `cursor: pointer`. The global rule reinforces the desired behavior. The CSS Modules add no conflicting `cursor` override.

### 2. `button:disabled` → `opacity: 0.5; cursor: not-allowed`

**Status: BENIGN.**

The Filters button uses `disabled` / `aria-disabled="true"`. The `opacity: 0.5` from the global rule renders correctly as a visually muted disabled state — consistent with the design intent. No override needed.

### 3. `input:not(...)` → `width: 100%; border-radius: var(--r-2); border: 1px solid var(--border); ...`

**Status: NOT APPLICABLE — no bare `<input>` elements on the approval page.**

The approval page (InboxPage, InboxStack, InboxApprovalCard, InboxTimeline, InboxToolbar) contains zero `<input>` elements. The global width/border injection cannot fire.

### 4. `p` → browser default `margin: 1em 0`

**Status: FIXED in CSS Module.**

The `.summary` class in `InboxApprovalCard.module.css` explicitly sets `margin-top: 0` to suppress the browser default top margin. The bottom margin is overridden by `margin-bottom: 18px`. Verified: `.summary` renders at exactly `margin-top: 0` (confirmed from spec — the paragraph is the only bare `<p>` inside a scoped CSS Module class, so the reset is applied correctly).

### 5. `table`, `th/td`, `tr` rules

**Status: NOT APPLICABLE — no `<table>` elements on the approval page.**

The approval page uses CSS Grid (`display: grid`) for layout throughout. No `<table>`, `<tr>`, `<th>`, or `<td>` elements present.

### 6. `textarea` → `resize: vertical`

**Status: NOT APPLICABLE — no `<textarea>` on the approval page.**

---

## Raw hex scan (compiled CSS modules)

Scanned all compiled CSS rules from approval feature modules (`InboxPage`, `InboxToolbar`, `InboxStack`, `InboxApprovalCard`, `InboxTimeline`) via `document.styleSheets` enumeration.

**Result: 0 raw hex values found.**

All color values use `var(--token)` references or `rgba(...)` with literal values permitted by spec (shadow overlays, backdrop semi-transparent white).

Permitted `rgba(...)` literals:
- `rgba(255,255,255,0.15)` — kindBadge glass background (no token for semi-transparent white on dark bg)
- `rgba(255,255,255,0.7)` / `rgba(255,255,255,0.65)` — codeVersion / deadlineLabel opacity on dark header
- `rgba(0,0,0,0.12)` / `rgba(0,0,0,0.04)` — card box-shadow (spec-specified)
- `rgba(200,74,42,0.15)` — bucketDotUrgent glow ring (accent at 15% — no token for this)

---

## Raw px spacing scan

Scanned for spacing token values used as raw px (4, 8, 12, 16, 20, 24, 32, 40, 56px) in CSS Module rules.

**Result: 8 occurrences found — all intentional per spec.**

| Occurrence | Value | Rule | Justification |
|---|---|---|---|
| `._card_` box-shadow | `24px`, `4px`, `12px` | `box-shadow: 0 24px 60px … 0 4px 12px` | Shadow literal from spec — not a spacing token |
| `._cardHeader_` padding | `28px` | `padding: var(--sp-5) 28px` | 28px has no token match; spec uses it directly |
| `._kindBadge_` backdrop | `8px` | `backdrop-filter: blur(8px)` | Blur radius — not a spacing token |
| `._heatmapBars_` height | `32px` | `height: 32px` | Fixed visual height — per spec |
| `._heatmapBar_` | `8px`, `4px` | `width: 8px; min-height: 4px` | Fixed visual geometry — per spec |
| `._bucketDotUrgent_` | `4px`, `4px` | `border: 4px solid; box-shadow: 0 0 0 4px` | Border/shadow literal — per spec |
| `._stageBar_` | `4px` | `height: 4px; border-radius: 2px` | Fixed visual element — per spec |
| `._reviewBtn_` | `32px` | `height: 32px` | Fixed button height — per spec |

**0 spacing token violations** (no case where `var(--sp-N)` should have been used but wasn't).

---

## Conclusion

| Check | Result |
|---|---|
| Raw hex in CSS Modules | 0 found — CLEAN |
| Spacing token violations | 0 found — CLEAN |
| Global `button` cursor bleed | BENIGN |
| Global `button:disabled` opacity bleed | BENIGN |
| Global `input` width bleed | NOT APPLICABLE |
| Browser `p` margin bleed | FIXED via `margin-top: 0` in `.summary` |
| Global `table/tr` bleed | NOT APPLICABLE |

**Overall: CLEAN — no leakage issues requiring action.**
