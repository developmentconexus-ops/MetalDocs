# Phase 3b — Style Port Summary
> Screen: Caixa de Aprovação (`caixa-aprovacao`)
> Date: 2026-05-08
> Phase: 3b — CSS Module fill (token-based style port from reference HTML)

---

## What was done

### 1. Token additions (`tokens.css`)

Two new radius tokens added to `src/styles/tokens.css`:

```css
--r-4: 16px;   /* card border-radius (was 12px — updated) */
--r-5: 12px;   /* badge/pill border-radius (was --r-4) */
```

The pre-existing `--r-4: 12px` was renamed to `--r-5: 12px`. The single existing consumer (`LibraryPage.module.css` `.tableCard`) was updated from `var(--r-4)` → `var(--r-5)` to preserve its 12px radius.

### 2. CSS Modules filled (5 files)

All five CSS Module files for the `features/approval` inbox screens were filled from empty skeletons to full token-based implementations:

| File | Classes | Key structures |
|---|---|---|
| `pages/InboxPage.module.css` | 1 | Full-height flex column, overflow hidden |
| `components/InboxToolbar.module.css` | 8 | 48px toolbar, breadcrumb, view switcher with `data-active` styling |
| `components/InboxStack.module.css` | 20 | 320px/1fr grid, queue rail, card stack with ghost cards, keyboard hints |
| `components/InboxApprovalCard.module.css` | 18 | 16px radius card, dark header, stats grid, action buttons |
| `components/InboxTimeline.module.css` | 30 | Timeline rail gradient, bucket sections, item rows (5-col grid), stage bars, heatmap |

### 3. Token rule compliance

- **Colors**: 100% `var(--token)` — 0 raw hex values
- **Spacing**: 100% `var(--sp-N)` for all spacing — `1px` borders and intentional design geometry (28px card padding, 4px stage bar height, 32px button height, 8px heatmap bar) documented as spec-permitted exceptions
- **Radii**: `var(--r-1)` through `var(--r-5)` + `var(--r-pill)` as applicable
- **Shadows**: `var(--shadow-1)` for UI shadows; literal box-shadow for card drop shadow (spec-specified)
- **Typography**: `rem` values for font sizes (no token mismatch); `var(--font-sans)` / `var(--font-mono)` for font families
- **Animations**: `urgentBlink` (queue dot), `cardIn` (card entrance), `urgentPulse` (bucket dot) — all CSS keyframe animations defined in the relevant modules

### 4. Global leakage resets

- `.summary` in `InboxApprovalCard.module.css`: `margin-top: 0` resets browser default `p` margin
- No other resets needed — no `<input>`, `<table>`, or other leakage-prone elements used

---

## Verification

### TypeScript
`pnpm tsc --noEmit -p tsconfig.build.json` — same pre-existing errors as Phase 2 (6 errors in unrelated files: auth tests, documents queries, shell Rail). Zero new errors introduced.

### Runtime rendering
App verified live at `http://localhost:4174/approvals` (1440×900):

- **Stack view**: Two-column grid (320px queue rail + 1fr card area) renders correctly. Queue items show active state (brand left border, surface-2 background). Ghost cards stack behind approval card at correct z-indices. Keyboard hints visible at bottom.
- **Timeline view**: Timeline rail gradient, urgent pulse animation on "Hoje" bucket dot, heatmap widget, 5-column item rows with stage progress bars and "Revisar" buttons all render correctly.
- **1024px**: Layout holds — queue rail and card both visible, card truncated as expected.
- **375px (mobile)**: Queue rail fills full width; card area off-screen (no mobile breakpoint in Phase 3b — Phase 3c scope).

### Computed style parity
See `parity-diff.md` for numerical comparison. All regions PASS. Only subpixel rendering differences (1px → 0.8px at 0.8× DPR) noted — not CSS errors.

### Leakage probe
See `leakage-probe.md`. Clean — 0 raw hex, 0 spacing token violations.

---

## Screenshots taken

All screenshots at `http://localhost:4174/approvals`:

- Stack view @ 1440×900 (main)
- Stack view @ 1024×768
- Stack view @ 375×812 (mobile)
- Timeline view @ 1440×900 (via HMR temp-patch `view='timeline'`)

Note: Reference design HTML (`caixa-aprovacao.html`) screenshot timed out via preview tool — likely due to CSS animations blocking headless renderer. Reference screenshots not saved; parity assessment performed directly against computed styles.

---

## Known gaps / Phase 3c scope

1. **No mobile breakpoint** — at 375px the card area is hidden (off-screen). Phase 3c should add `@media (max-width: 768px)` to `InboxStack` to stack queue + card vertically.
2. **`InboxPage` view switcher is static** — `const view = 'stack'` hardcoded. Phase 3c replaces with `useState`.
3. **`.cardHeaderUrgent`** — gradient variant defined but not wired to data. Phase 3c connects to `item.urgent` flag.
4. **`.queueItemNumberActive`** — spec uses a descendant selector `.queueItemActive .queueItemNumber` in the CSS; the TSX does not apply a separate `queueItemNumberActive` class. The CSS module uses the descendant rule pattern which works correctly.

---

## Commits

1. `feat(tokens): add --r-4, --r-5 for card/badge radii`
2. `feat(approval): Phase 3b style port for caixa-aprovacao`
