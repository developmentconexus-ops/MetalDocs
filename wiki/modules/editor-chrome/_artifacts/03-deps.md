# editor-chrome — Cross-Dependency Map (Phase 3)

Module path: `frontend/apps/web/src/features/shared/components/editor-chrome/`
Files: `EditorChrome.tsx`, `parts/VersionBadge.tsx`, `parts/AutosaveStatus.tsx`, `index.ts`
Last verified: 2026-05-10

---

## 1. Imports OUT

Internal MetalDocs packages this module imports. Third-party (`react`) excluded.

| Imported package | First seen in (file:line) | Symbols used | Purpose |
|---|---|---|---|
| `./EditorChrome.module.css` | `EditorChrome.tsx:2` | `styles` (wrapper, overlayLeft, overlayCenter, overlayRight, overlayAlert, docTitle, docSep, docMeta, iconBtn, ghostBtn, primaryBtn) | Layout, overlay positioning, button primitives, title text helpers |
| `./VersionBadge.module.css` | `parts/VersionBadge.tsx:2` | `styles` (badge) | Monospace chip styling |
| `./AutosaveStatus.module.css` | `parts/AutosaveStatus.tsx:1` | `styles` (status, statusError, dot, dotIdle, dotSaving, dotError, check) | Autosave indicator states + pulse animation |
| `styles/tokens.css` (CSS custom properties — logical OUT-edge) | `EditorChrome.module.css:17,26,29,46,55,58,69,72,80,87,92,108,111,131,143,148,152,166,229–242` | `--surface-2`, `--surface`, `--border`, `--border-strong`, `--text`, `--text-faint`, `--text-muted`, `--text-soft`, `--brand`, `--brand-soft`, `--accent`, `--sp-1`–`--sp-5`, `--r-1`, `--r-2`, `--font-sans` | Design tokens consumed via `var(--...)` from global token sheet |
| `styles/tokens.css` (CSS custom properties — logical OUT-edge) | `parts/AutosaveStatus.module.css:5,14,25,29,34,38` | `--sp-1`, `--text-muted`, `--danger`, `--success`, `--info`, `--font-sans` | Autosave indicator tokens |
| `styles/tokens.css` (CSS custom properties — logical OUT-edge) | `parts/VersionBadge.module.css:4,10,11` | `--font-mono`, `--r-1`, `--brand` | Badge tokens |
| `@eigenpal/docx-js-editor` (CSS-level OUT-edge only) | `EditorChrome.module.css:160,166–168,174,181,186,190,193,200,201,216,217,224,229` | `:global(.ep-root ...)` selectors targeting `[data-testid="title-bar"]`, `[data-testid="formatting-bar"]`, `[data-testid="font-size-display"]`, `[data-testid="font-size-input"]`, `[role="combobox"]`, `select`, `[role="separator"]`, `::-webkit-scrollbar*` | Scoped eigenpal override styles; no JS import, no JS coupling to eigenpal in this module |

---

## 2. Imports IN

Other internal files that import this module or its named exports.

| Importer | File:line of import | Symbols used | Notes |
|---|---|---|---|
| `features/templates/pages/TemplateEditorPage.tsx` | `:17–21` (from `'../../shared/components/editor-chrome'`) | `EditorChrome`, `editorChromeStyles`, `VersionBadge`, `AutosaveStatus`, `type AutosaveState` | All 5 public exports consumed |
| `features/documents/pages/DocumentEditorPage.tsx` | `:15–19` (from `'../../shared/components/editor-chrome'`) | `EditorChrome`, `editorChromeStyles`, `VersionBadge`, `AutosaveStatus`, `type AutosaveState` | All 5 public exports consumed |

### Name collision: `AutosaveStatus` / `AutosaveState`

`features/documents/hooks/editor/useDocumentAutosave.ts:5` declares a **local** exported type:
```ts
export type AutosaveStatus = 'idle' | 'dirty' | 'saving' | 'saved' | 'stale' | 'session_lost' | 'error';
```
This is a 7-value union. `editor-chrome`'s `AutosaveState` (`parts/AutosaveStatus.tsx:3`) is a 4-value union: `'idle' | 'saving' | 'saved' | 'error'`.

- The hook's type is named `AutosaveStatus`; the component's prop type is named `AutosaveState` — different names, no TS collision at import sites.
- `DocumentEditorPage` maps `autosave.status` (hook's `AutosaveStatus`) to `autosaveState: AutosaveState` (component's 4-value type) at `DocumentEditorPage.tsx:184`. The mapping collapses `'dirty'`, `'stale'`, `'session_lost'` — (unclear: no explicit coercion is visible at line 184 from the data collected; mapping logic not read in full).

---

## 3. Mount sites

Every JSX `<EditorChrome>` occurrence.

| File | Line | Notes |
|---|---|---|
| `features/templates/pages/TemplateEditorPage.tsx` | `274` (open), `341` (close) | Slots: center (title + VersionBadge), right (AutosaveStatus + action buttons); no alert, no left |
| `features/documents/pages/DocumentEditorPage.tsx` | `217` (open), `257` (close) | Slots: center (title + VersionBadge), right (AutosaveStatus + action button); no alert, no left |

---

## 4. Configuration surface

No `import.meta.env`, `process.env`, or `useFlag` calls found in any of the 4 module files. Module reads no runtime configuration.

---

## 5. Test surface

No test coverage found for editor-chrome primitives.

Grep over `**/*.test.{ts,tsx}` and `**/*.spec.{ts,tsx}` under `frontend/` for `EditorChrome|VersionBadge|AutosaveStatus` returned zero matches.

---

## 6. Routing / route registration

n/a — primitive, not a route owner.
