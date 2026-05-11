# Module: EditorChrome (Shared Editor Primitive)

> **Last verified:** 2026-05-10
> **Scope:** The `EditorChrome` shared component — purpose, slot API, sub-parts, eigenpal style overrides, design-token coverage, and how consuming pages use it.
> **Out of scope:** Eigenpal internals (see `modules/editor-ui-eigenpal.md`), template authoring business logic (see `modules/templates-v2.md`), document editing business logic (see `modules/documents.md`).
> **Key files:**
> - `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.tsx:31` — `EditorChrome` component + `EditorChromeProps` type + `editorChromeStyles` re-export
> - `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.module.css:1` — wrapper, overlay positioning, button primitives, text helpers, all eigenpal overrides
> - `frontend/apps/web/src/features/shared/components/editor-chrome/parts/VersionBadge.tsx:13` — `VersionBadge` component (monospace chip)
> - `frontend/apps/web/src/features/shared/components/editor-chrome/parts/AutosaveStatus.tsx:28` — `AutosaveStatus` component (pulsing dot / check / error)
> - `frontend/apps/web/src/features/shared/components/editor-chrome/index.ts:1` — public barrel (`EditorChrome`, `editorChromeStyles`, `VersionBadge`, `AutosaveStatus`, types)
> - `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx:1` — consumer 1 (template authoring)
> - `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:1` — consumer 2 (document editor)

---

## Why this exists

`TemplateEditorPage` and `DocumentEditorPage` both mount eigenpal (`DocxEditor` / `MetalDocsEditor`) and overlay a custom toolbar atop eigenpal's native 40px title bar. Before this primitive, the two pages were evolving independently: templates had a richer toolbar (status pill, version badge, wine formatting bar, gradient scrollbar) while the doc editor was catching up. Per the `metaldocs-frontend` skill rule — _"used by 2+ features → `features/shared/`"_ — the overlay pattern was extracted into `EditorChrome`.

Benefits:
- Eigenpal CSS overrides (compact title bar, wine tint, gradient scrollbar) live in one place — no per-page duplication or drift.
- Button primitives (`.iconBtn`, `.ghostBtn`, `.primaryBtn`) and text helpers (`.docTitle`, `.docMeta`, `.docSep`) are shared via `editorChromeStyles`, so both pages use identical visual treatment.
- Both pages pass only the buttons they need via slot props; the chrome itself is unaware of domain logic.

---

## Slot API

```ts
// EditorChrome.tsx:4
export type EditorChromeProps = {
  left?:     ReactNode;  // top-left overlay: back button, etc.
  center?:   ReactNode;  // top-center overlay: title, badges, status pill
  right?:    ReactNode;  // top-right overlay: autosave, actions
  alert?:    ReactNode;  // banner below the 40px title bar
  children:  ReactNode;  // the eigenpal editor instance
  className?: string;    // page-specific wrapper tweak
};
```

All slots are optional except `children`. The wrapper renders `position: relative`; each overlay is `position: absolute` within it. The center slot sets `pointer-events: none` so it never steals clicks intended for eigenpal.

Usage pattern:

```tsx
<EditorChrome
  left={<button className={editorChromeStyles.iconBtn}>...</button>}
  center={
    <>
      <span className={editorChromeStyles.docTitle}>{doc.name}</span>
      <StatusPill status={doc.status as DocumentStatus} />
    </>
  }
  right={<AutosaveStatus status={autosaveState} />}
>
  <MetalDocsEditor ... />
</EditorChrome>
```

---

## Sub-parts

### `VersionBadge`

Source: `parts/VersionBadge.tsx:13`. Brand-colored monospace chip for revision/version labels (e.g. `REV05`, `v5`). Accepts `children` + optional `className`. Style from `VersionBadge.module.css` — uses `--font-mono` and `--brand-*` tokens.

### `AutosaveStatus`

Source: `parts/AutosaveStatus.tsx:28`. Visual autosave indicator. Accepts:

| Prop | Type | Notes |
|------|------|-------|
| `status` | `'idle' \| 'saving' \| 'saved' \| 'error'` | Required |
| `labels` | `{ idle?, saving?, saved?, error? }` | Optional override; defaults are pt-BR |
| `className` | `string` | Optional extra class |

State rendering:
- `saving` — pulsing dot + "Salvando…"
- `saved` — green check SVG + "Salvo"
- `error` — red dot + "Erro ao salvar"
- `idle` — neutral dot + "Salvo"

### `editorChromeStyles` (re-exported CSS module)

`EditorChrome.tsx:47` re-exports the CSS module as `editorChromeStyles`. Consumers import it alongside the component to use shared button and text class names without redefining them:

```ts
import { EditorChrome, editorChromeStyles } from '../../shared/components/editor-chrome';
// then: <button className={editorChromeStyles.iconBtn}>
```

Available class names:

| Class | Purpose |
|-------|---------|
| `.iconBtn` | 32px square icon-only button |
| `.ghostBtn` | text + optional icon, transparent bg |
| `.primaryBtn` | filled brand-color action button |
| `.docTitle` | truncated title text in center overlay |
| `.docSep` | `/` separator between title segments |
| `.docMeta` | secondary metadata label (muted) |

---

## Eigenpal overrides

All eigenpal overrides are scoped to `.wrapper :global(.ep-root)` in `EditorChrome.module.css`. Moving them here prevents per-page duplication. Key overrides:

- **Compact title bar:** reduces eigenpal's default title bar padding so the 40px overlay sits flush.
- **Wine formatting bar:** tints the eigenpal formatting toolbar with `var(--brand-soft)` background and `var(--brand)` accent — the "wine pill" effect.
- **Doc icon recolor:** SVG doc icon colored to `var(--brand-pale)`.
- **Font-size input fixed width:** pins the font-size dropdown to a stable width to avoid layout shift on value change.
- **Gradient scrollbar:** custom thin scrollbar with `var(--brand-pale)` thumb inside the editor canvas.

All values use design tokens from `styles/tokens.css` — no hardcoded hex colors.

---

## Design-token coverage

`EditorChrome.module.css` is fully token-driven. Token namespaces used:

- `--brand`, `--brand-deep`, `--brand-soft`, `--brand-pale` — wine palette overrides
- `--surface-2` — wrapper background
- `--sp-*` — spacing (overlay gaps, padding)
- `--r-*` — border radii on button primitives
- `--text`, `--text-muted` — text helper colors
- `--shadow-*` — button focus rings

No hardcoded colors, no magic pixel values outside the `--sp-*` / `--r-*` system.

---

## Consumers

### `TemplateEditorPage`

Source: `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx`. Uses `<EditorChrome left={...} center={...} right={...}>` with:
- `left`: back button + sidebar toggle icons (`.iconBtn`)
- `center`: template title + `VersionBadge` + `StatusPill`
- `right`: `AutosaveStatus` + submit action (`.primaryBtn`)

CSS module shrank from ~495 → ~155 lines (only rails, side panel, and canvas layout remain — eigenpal overrides moved to `EditorChrome.module.css`).

### `DocumentEditorPage`

Source: `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`. Slot assignment (2026-05-06):
- `left`: **unused** — back button lives in a dedicated `<aside class={styles.rail}>` element outside `EditorChrome`, to avoid collision with eigenpal's File icon overlay.
- `center`: `CodeChip` (doc code) + document name + `VersionBadge` (revision) + `StatusPill`
- `right`: `AutosaveStatus` + "Submeter para revisão" button only

Removed from right slot: `CheckpointsDialog` mount, `checkpointsOpen` state, `handleRestored`, `ExportMenuButton`, and the "Revisões" button. These components still exist as standalone files but are not mounted from this page.

The intermediate `EditorDocBar.tsx` + `EditorDocBar.module.css` were deleted when this page was migrated to `EditorChrome`. No shims remain.

---

## Reuses

- `StatusPill` from `components/ui/` — status badges for both pages (SSOT for `DocumentStatus` type). See `modules/documents.md` for the 8-state catalog.
- Design tokens from `styles/tokens.css:2` — no cross-primitive hardcoding.

---

## Cross-refs

- [modules/editor-ui-eigenpal.md](editor-ui-eigenpal.md) — eigenpal wrapper (`MetalDocsEditor`); this primitive wraps the output of that layer
- [modules/templates-v2.md](templates-v2.md) — `TemplateEditorPage` consumer (frontend doc)
- [modules/templates_v2.md](templates_v2.md) — backend module the editor chrome surfaces authoring UX for (Arc42 doc)
- [modules/documents.md](documents.md) — `DocumentEditorPage` consumer; `StatusPill` 8-state catalog
- [architecture/frontend-structure.md](../architecture/frontend-structure.md) — `features/shared/` placement rule
- [styles/tokens.css](../../frontend/apps/web/src/styles/tokens.css) — design token definitions
