# Phase 1 — Surface Scan

**Module:** editor-chrome (FE-only React primitive)
**Path:** `frontend/apps/web/src/features/shared/components/editor-chrome/`
**Method note:** Codex subagent blocked by policy; main agent performed manual surface scan via Read/Grep on the 7-file module. No AST tooling used.

## 1. File tree

```
features/shared/components/editor-chrome/
├── EditorChrome.tsx              # main wrapper component + slot API + styles re-export
├── EditorChrome.module.css       # wrapper, overlays, button/text primitives, eigenpal :global overrides
├── index.ts                      # barrel: re-exports EditorChrome, editorChromeStyles, parts, types
└── parts/
    ├── VersionBadge.tsx          # monospace brand chip for version/revision labels
    ├── VersionBadge.module.css   # single .badge class
    ├── AutosaveStatus.tsx        # autosave indicator (idle / saving / saved / error states)
    └── AutosaveStatus.module.css # .status / .dot / .check + keyframes pulse + prefers-reduced-motion
```

No test files (`*.test.tsx` / `*.spec.tsx`) co-located with the module — `(undocumented test coverage)`.

## 2. Public surface

| File:line | Kind | Name | Signature / type | Doc comment first line |
|---|---|---|---|---|
| `EditorChrome.tsx:4` | type (exported) | `EditorChromeProps` | `{ left?: ReactNode; center?: ReactNode; right?: ReactNode; alert?: ReactNode; children: ReactNode; className?: string }` | (per-field JSDoc; type itself undocumented) |
| `EditorChrome.tsx:31` | component (function) | `EditorChrome` | `(props: EditorChromeProps) => JSX.Element` | `EditorChrome — shared wrapper for eigenpal-based document editors.` |
| `EditorChrome.tsx:47` | const (re-export) | `editorChromeStyles` | `typeof import('./EditorChrome.module.css').default` | `Re-export the module CSS classes so consumers can use the button primitives ...` |
| `parts/VersionBadge.tsx:13` | component (function) | `VersionBadge` | `(props: { children: ReactNode; className?: string }) => JSX.Element` | `Brand-colored monospace chip for revision/version labels.` |
| `parts/AutosaveStatus.tsx:3` | type (exported) | `AutosaveState` | `'idle' \| 'saving' \| 'saved' \| 'error'` | (undocumented) |
| `parts/AutosaveStatus.tsx:28` | component (function) | `AutosaveStatus` | `(props: { status: AutosaveState; labels?: Partial<Record<AutosaveState,string>>; className?: string }) => JSX.Element` | `Editor autosave indicator. Shows pulsing dot while saving, green check when saved, red label on error, neutral idle.` |
| `index.ts:1–5` | barrel | re-exports `EditorChrome`, `editorChromeStyles`, `EditorChromeProps`, `VersionBadge`, `AutosaveStatus`, `AutosaveState` | | barrel — no doc |

Notes:
- `VersionBadgeProps` and `AutosaveStatusProps` are NOT exported (local types). Consumers reach component props inline.
- `CheckIcon` in `AutosaveStatus.tsx:65` is module-local (unexported helper).
- `DEFAULT_LABELS` in `AutosaveStatus.tsx:17` is module-local (unexported const). Labels are pt-BR.

## 3. HTTP operations

**n/a — frontend primitive, no HTTP routes.**

## 4. Migration list

**n/a — frontend module, no SQL migrations.**

## 5. CSS surface

### 5.1 `EditorChrome.module.css`

Local class selectors (consumed via CSS Modules; available through `styles.*` and re-exported `editorChromeStyles.*`):

| Selector | File:line | Targets | Purpose |
|---|---|---|---|
| `.wrapper` | `EditorChrome.module.css:9` | root `<div>` | flex column, position relative, `--surface-2` bg, overflow hidden |
| `.overlayLeft` | `:22` | absolute slot div | top-left, 40px tall, `z-index:60`, gap `--sp-2` |
| `.overlayCenter` | `:33` | absolute slot div | centered (`left:50%; transform:translateX(-50%)`), `pointer-events:none`, `z-index:60`, white-space:nowrap |
| `.overlayRight` | `:52` | absolute slot div | top-right, 40px tall, `z-index:100` |
| `.overlayAlert` | `:63` | absolute banner | below 40px title bar, full width, `z-index:100`, `--surface` bg, bottom border |
| `.docTitle` | `:77` | center-slot title text | 15px / 600, truncate with ellipsis, max-width 320px |
| `.docSep` | `:87` | "/" separator | `--text-faint`, weight 400 |
| `.docMeta` | `:92` | secondary metadata | `--text-muted`, 12px |
| `.iconBtn` | `:100` | 26×26 icon-only button (NOT 32px as stated in stub) | transparent bg, border, `--text-soft`, `:hover` darkens |
| `.ghostBtn` | `:117` | text/icon transparent button | 26px tall, border, `--text-soft`, `:disabled` 0.55 opacity |
| `.primaryBtn` | `:137` | filled brand action | `--brand` bg, white text, `:hover` → `--brand-soft`, `:disabled` 0.55 opacity |

Eigenpal `:global` overrides (couple to eigenpal DOM, marked with `!important`):

| Selector | File:line | External target | Purpose |
|---|---|---|---|
| `.wrapper :global(.ep-root [data-testid="title-bar"])` | `:160` | eigenpal title bar | compact: 2px top/bottom padding |
| `.wrapper :global(.ep-root) svg path[fill="#cbd5e1"]` | `:166` | eigenpal doc-icon SVG path | recolor to `var(--brand)` |
| `.wrapper :global(.ep-root) svg path[fill="#94a3b8"]` | `:167` | eigenpal doc-icon SVG path | recolor to `var(--brand-soft)` |
| `.wrapper :global(.ep-root) svg[viewBox="0 0 32 40"]` | `:168` | eigenpal doc-icon SVG | resize 16×20 |
| `.wrapper :global(.ep-root [data-testid="formatting-bar"])` | `:174` | eigenpal formatting bar | wine tint bg, overflow hidden (BOTH axes — comment notes capture-phase scroll bug w/ font-size dropdown) |
| `.wrapper :global(.ep-root [data-testid="formatting-bar"] button:not(.docx-advanced-color-picker-dropdown *))` | `:181` | formatting-bar buttons (excl. color picker) | recolor + hover state |
| `.wrapper :global(.ep-root [data-testid="formatting-bar"] button[aria-pressed="true"] ...)` | `:193` | active toolbar buttons | pressed-state bg |
| `.wrapper :global(.ep-root [data-testid="font-size-display/input"])` | `:200–201` | eigenpal font-size widget | fixed 56×28 to prevent scroll-on-focus that closes dropdown |
| `.wrapper :global(.ep-root [data-testid="formatting-bar"] [role="combobox"])` | `:216` | font-family / style selectors | recolor |
| `.wrapper :global(.ep-root [data-testid="formatting-bar"] [role="separator"])` | `:224` | toolbar group dividers | opacity tweak |
| `.wrapper :global(.ep-root) ::-webkit-scrollbar(-track/-thumb)` | `:229–238` | eigenpal scroll region | wine gradient scrollbar |
| `.wrapper :global(.ep-root) { scrollbar-width; scrollbar-color }` | `:239–242` | Firefox scrollbar | thin + brand color |

### 5.2 `parts/VersionBadge.module.css`

| Selector | File:line | Purpose |
|---|---|---|
| `.badge` | `:1` | inline-flex, `--font-mono` 10.5px / 600, `--brand` bg, white text, `--r-1` corners, letter-spacing 0.06em |

### 5.3 `parts/AutosaveStatus.module.css`

| Selector | File:line | Purpose |
|---|---|---|
| `.status` | `:1` | inline-flex, `--text-muted`, 12px, `--font-sans`, gap `--sp-1`, `min-width:60px` |
| `.statusError` | `:13` | overrides color → `--danger` |
| `.dot` | `:17` | 8×8 round flex-item |
| `.dotIdle` | `:24` | `--success` bg |
| `.dotSaving` | `:28` | `--info` bg + `pulse` animation 1.2s ease-in-out infinite |
| `.dotError` | `:33` | `--danger` bg |
| `.check` | `:37` | `--success` color, 14×14 |
| `@keyframes pulse` | `:44` | 0/100% opacity 1 → 50% opacity 0.4 |
| `@media (prefers-reduced-motion: reduce)` | `:49` | disables `.dotSaving` animation |

## 6. Design-token usage

| Token | File:line | Used for |
|---|---|---|
| `--surface` | `EditorChrome.module.css:71` | alert banner bg |
| `--surface-2` | `:16, 115, 134, 234` | wrapper bg, button hover bg, scrollbar thumb border |
| `--border` | `:108, 124, 72` | button borders, alert bottom border |
| `--border-strong` | `:115, 134` | button hover border |
| `--brand` | `:143, 166, 232, 233, 241` (also `VersionBadge.module.css:10`) | primary button bg, doc-icon fill, scrollbar gradient, scrollbar-color, badge bg |
| `--brand-soft` | `:154, 167, 232, 237` | primary hover bg, doc-icon secondary fill, scrollbar gradient |
| `--accent` | `:237` | scrollbar hover gradient stop |
| `--text` | `:80, 188, 195` | title color, formatting-bar button hover/pressed |
| `--text-soft` | `:110, 126, 184, 191, 204, 218` | icon button color, ghost button color, formatting-bar buttons |
| `--text-muted` | `:94` (also `AutosaveStatus.module.css:5`) | secondary meta label, autosave label |
| `--text-faint` | `:88` | "/" separator |
| `--success` | `AutosaveStatus.module.css:25, 38` | idle dot, check icon |
| `--danger` | `AutosaveStatus.module.css:14, 34` | error text, error dot |
| `--info` | `AutosaveStatus.module.css:29` | saving dot |
| `--font-sans` | `EditorChrome.module.css:46, 131, 151` (also `AutosaveStatus.module.css:7`) | overlay text font, button font |
| `--font-mono` | `VersionBadge.module.css:4` | badge font |
| `--sp-1` | `:120, 140` (also `AutosaveStatus.module.css:4`) | button gap, status gap |
| `--sp-2` | `:30, 41, 59` | overlay gap, button gap |
| `--sp-3` | `:25, 121, 141` | overlay left padding, ghost/primary horizontal padding |
| `--sp-5` | `:69` | alert banner horizontal padding |
| `--r-1` | `:109, 205` (also `VersionBadge.module.css:9`) | icon button radius, font-size widget radius, badge radius |
| `--r-2` | `:125, 145` | ghost button radius, primary button radius |

**Hardcoded values present (not token-driven):**

| Value | File:line | Context |
|---|---|---|
| `40px` (overlay top heights) | `:26, 36, 56, 65` | overlay row height — mirrors eigenpal's title-bar height; no `--editor-titlebar-h` token exists |
| `13px` font-size | `:45` | overlay center font-size |
| `15px` / `12px` / `12.5px` font-size | `:78, 95, 128, 147, 70` | typography hardcoded — token catalog has only `--font-sans/--font-mono`, no `--fs-*` |
| `600`, `500`, `400` font-weight | `:79, 128, 148` | hardcoded weights |
| `26px` button height/width | `:104–105, 121, 141` | mirrors button-primitive convention; no `--btn-h-sm` token |
| `320px` max-width | `:84` | title truncation cap |
| `rgba(107, 31, 42, 0.18)` / `0.2` | `:175, 176` | formatting-bar wine tint — duplicates `--brand` rgb (107,31,42) but inlined for alpha |
| `rgba(0, 0, 0, 0.05–0.15)` | `:187, 194, 202, 203, 220, 225` | neutral hover/border tints — not token-backed |
| `10px` / `2px` scrollbar dims | `:229, 234` | scrollbar dimensions |
| `10.5px` font-size, `0.06em` letter-spacing | `VersionBadge.module.css:5, 8` | badge typography hardcoded |
| `2px 6px` badge padding | `VersionBadge.module.css:8` | hardcoded |
| `8px` dot dimensions | `AutosaveStatus.module.css:18–19` | hardcoded |
| `14px` SVG dims | `AutosaveStatus.module.css:39–40` (also TSX:69–70) | hardcoded |
| `1.2s` animation duration | `AutosaveStatus.module.css:30` | hardcoded |
| `#fff` (white) | `:146` (also `VersionBadge.module.css:11`) | primary button + badge text — no `--text-on-brand` token |

The existing wiki stub claims "fully token-driven · no hardcoded hex colors · no magic pixel values outside the `--sp-*` / `--r-*` system." That claim is partly inaccurate — there are no hardcoded hex *colors* (the `rgba(107,31,42,...)` inlines brand RGB; `#fff` is white-on-brand), but font sizes, font weights, button heights, animation durations, and the overlay 40px height are not tokenised.

## 7. Eigenpal coupling (selector inventory)

Pixel-/attribute-level couplings to eigenpal DOM (breaks silently if eigenpal renames):

- `[data-testid="title-bar"]`
- `[data-testid="formatting-bar"]`
- `[data-testid="font-size-display"]`
- `[data-testid="font-size-input"]`
- `[role="combobox"]` inside formatting-bar
- `[role="separator"]` inside formatting-bar
- `[aria-pressed="true"]` button selector inside formatting-bar
- `svg path[fill="#cbd5e1"]` / `svg path[fill="#94a3b8"]` (HARDCODED hex match — fragile to eigenpal palette changes)
- `svg[viewBox="0 0 32 40"]` (geometry-anchored)
- `.docx-advanced-color-picker-dropdown` (class-name selector)
- `.ep-root` (root anchor, via `:global`)

If eigenpal changes any of these attribute values, the override silently no-ops. No runtime check exists.

## 8. Accessibility surface

- `AutosaveStatus`: dot `<span>` carries `aria-hidden="true"` (cosmetic), but the wrapper `<span>` has NO `role="status"` / `aria-live="polite"`. State changes (idle → saving → saved → error) are NOT announced to screen readers. The check SVG has `aria-hidden="true"` and no visible label is associated — text content provides the only accessible name.
- `VersionBadge`: pure visual chip, no a11y annotations. Treated as flow text.
- `EditorChrome`: wrapper `<div>` has no landmark role. `.overlayCenter` has `pointer-events:none` — focusable children inside the center slot would not receive click but COULD still receive focus (keyboard tab) and trigger keyboard activation; comment at `:49–50` warns that interactive children must opt back in with `pointer-events:auto` — not enforced.

## 9. Public-API observations (for §5 of doc + tech-debt feed)

- Slot API uses `ReactNode` (untyped slot contents — no constraint on what consumers pass). No discriminated union on slot identity.
- `editorChromeStyles` re-export is the SOLE mechanism by which consumers obtain the shared button/text class names. Type is `typeof styles` (inferred CSS-module record); the consuming pages can typo a class name and get `undefined` at runtime (only flagged if `noUncheckedIndexedAccess` is on — verify in Phase 6.75).
- `AutosaveState` (`'idle' | 'saving' | 'saved' | 'error'`) is a 4-state enum exported from this module. The Documents feature has a parallel local type `AutosaveStatus` in `features/documents/hooks/editor/useDocumentAutosave.ts:5` with a 7-state superset (`'idle' | 'dirty' | 'saving' | 'saved' | 'stale' | 'session_lost' | 'error'`). The page-level code maps the larger enum to this 4-state enum (see `DocumentEditorPage.tsx:184` and `TemplateEditorPage.tsx:217` — `autosaveState: AutosaveState = ...`).

## Summary

- **6 exported symbols** (2 components from main file + 2 components from parts + 1 type alias from `AutosaveState` + 1 props type `EditorChromeProps` + 1 const re-export `editorChromeStyles`; barrel re-exports them)
- **15 local CSS class selectors** (across 3 CSS modules) + **17 `:global(.ep-root ...)` overrides**
- **22 design-token references** + **~15 hardcoded magic-value sites** (px, weights, durations)
- **0 HTTP routes · 0 migrations** (n/a — FE primitive)
