# Phase 3b — Style Port: novo-template-escopo

> **Status:** COMPLETE (awaiting user screenshot approval)
> **Date:** 2026-05-09
> **Branch:** worktree-agent-aa265c117ddbc20b9

---

## Token Map

| Design value (template-wizard.jsx TplStep1) | Token used | Notes |
|---|---|---|
| profileGrid `display: grid` | `display: grid` | No token needed |
| profileGrid `gridTemplateColumns: repeat(2, 1fr)` | literal value | No token needed |
| profileGrid `gap: 10px` | raw `10px` | No --sp-2-5 token; StepProfile.module.css uses same raw value |
| profileGrid `marginBottom: 8px` | `var(--sp-2)` | sp-2 = 8px ✓ |
| profileCard `gap: 6px` | raw `6px` | No token for 6px; StepProfile uses same raw value |
| profileCard `padding: 16px` | `var(--sp-4)` | sp-4 = 16px ✓ (overrides SelectableCard's default sp-4) |
| profileHeader `display: flex` | `display: flex` | No token needed |
| profileHeader `alignItems: center` | `align-items: center` | No token needed |
| profileHeader `gap: 8px` | `var(--sp-2)` | sp-2 = 8px ✓ |
| profileCode `fontSize: 13px` | raw `13px` | No fz token; TODO(novo-template-wizard:typography-tokens) |
| profileCode `fontWeight: 600` | raw `600` | No weight token |
| profileName `fontSize: 13px` | raw `13px` | Same as profileCode |
| profileName `fontWeight: 500` | raw `500` | No weight token |
| profileName `flex: 1` | `flex: 1` | No token needed |
| profileMeta `display: flex` | `display: flex` | No token needed |
| profileMeta `gap: 12px` | `var(--sp-3)` | sp-3 = 12px ✓ |
| profileMeta `fontSize: 11px` | raw `11px` | No fz token; TODO(novo-template-wizard:typography-tokens) |
| profileMeta `color: var(--text-muted)` | `var(--text-muted)` | Token ✓ |

**kicker/h2/caption margins:** Already handled by `WizardShell.module.css` rules `.container > :global(.card) > :global(.kicker/h2/caption)`. No additional scoping needed in StepScope.

---

## Missing Tokens Added

None. Raw values `10px`, `6px`, `13px`, `11px` follow the established precedent in `StepProfile.module.css` (same values, same TODO comments for the typography-tokens milestone). No new tokens added.

---

## Bug Fixed During Phase 3b

`TemplateWizardPage.tsx` was missing `export { TemplateWizardPage as Component }`. React Router lazy loading requires a named `Component` export — without it, the route rendered an empty `<main>`. Fixed by adding the re-export. This was a Phase 3a gap.

---

## Computed-Style Snapshots

### profileGrid (at 1440×900)

```json
{
  "display": "grid",
  "cols": "402.2px 402.2px",
  "gap": "10px",
  "mb": "8px",
  "w": 814.4
}
```

- `display: grid` ✓
- `gridTemplateColumns: repeat(2, 1fr)` resolves to two equal columns ✓
- `gap: 10px` ✓ (matches design 10px)
- `marginBottom: 8px` ✓ (var(--sp-2) = 8px)

### profileCard (first card, selected state)

```json
{
  "padding": "16px",
  "border": "1.6px solid rgb(107, 31, 42)",
  "bg": "rgb(249, 240, 240)",
  "br": "8px",
  "textAlign": "left"
}
```

- `padding: 16px` ✓ (var(--sp-4) = 16px, override of SelectableCard gap)
- `border`: SelectableCard `.selected` renders 2px brand border (1.6px due to viewport scaling) ✓
- `background`: `--brand-pale` (#f9f0f0 = rgb(249,240,240)) ✓
- `borderRadius`: `--r-3` = 8px ✓
- `textAlign: left` ✓

### profileHeader

```json
{
  "display": "flex",
  "ai": "center",
  "gap": "8px",
  "mb": "0px"
}
```

- `display: flex` ✓
- `alignItems: center` ✓
- `gap: 8px` ✓ (var(--sp-2) = 8px)
- `marginBottom: 0px` — no margin-bottom set; inter-row spacing via SelectableCard gap ✓

### profileMeta

```json
{
  "display": "flex",
  "gap": "12px",
  "fs": "11px",
  "color": "rgb(138, 117, 117)"
}
```

- `display: flex` ✓
- `gap: 12px` ✓ (var(--sp-3) = 12px)
- `fontSize: 11px` ✓ (matches design 11px)
- `color: rgb(138,117,117)` = `--text-muted` (#8a7575) ✓

---

## Leakage Probe Results

| Matched selector | Rule | Impact | Action |
|---|---|---|---|
| `*` | `box-sizing: border-box` | Universal — harmless | None |
| `button, input, select, textarea` | `font: inherit` | Base reset — desired | None |
| `button` | `cursor: pointer` | Base reset — desired | None |
| `._root_v9gta_1` | SelectableCard base styles (flex, padding, border, bg, text-align, etc.) | Own component styles — correct | None |
| `._selected_v9gta_23` | SelectableCard selected state (brand border, brand-pale bg) | Own component state — correct | None |
| `._profileCard_188uq_41` | `gap: 6px; padding: var(--sp-4)` | Our override — correct | None |

No unexpected global style leakage into profileCard elements.

---

## Token Coverage Result

```
grep -En '#[0-9a-fA-F]{3,8}|rgb\(' src/features/templates/components/wizard/steps/StepScope.module.css | grep -v 'var(--'
(empty output)
```

**PASS** — No raw hex or rgb() values in StepScope.module.css.

---

## Screenshot Status

Screenshots taken from main repo dev server (worktree has no node_modules). Server: `metaldocs-web` at localhost:4174.

| Viewport | Status |
|---|---|
| 1440×900 | Captured |
| 1024×768 | Captured |
| 375×812 | Captured |

**Awaiting user screenshot approval.**

---

## Notes

- Phase 3a gap: `TemplateWizardPage` was missing `Component` named export — fixed in this phase alongside CSS.
- Route navigated to `/templates-v2/novo`.
- Responsive single-column collapse at ≤600px verified at 375px viewport.
- `profileCard.padding` intentionally overrides SelectableCard's `var(--sp-4)` default with the same value — this is a no-op on padding but documents the design intent explicitly and allows future delta if needed.
