# Module: Editor UI (Eigenpal Integration)

> _Changelog: 2026-04-26 — added note that `applyVariables` is NOT used in writer mode (ADR 0008). 2026-05-04 — DocumentEditorPage consumer updated: isEditable gate, PDF polling wired. 2026-05-06 — eigenpal CSS overrides moved to `EditorChrome.module.css` (shared primitive); consumer paths updated (no longer under `v2/` sub-folder). 2026-05-06 — `templatePlugin` gated to `template-draft` mode only; `EditorMode` type expanded to three values; document editor layout switched to left rail._
>
> **Last verified:** 2026-05-06
> **Scope:** How MetalDocs wraps `@eigenpal/docx-js-editor`, what plugins are registered, autosave wiring, ProseMirror access patterns.
> **Out of scope:** EigenPal fork internals (see `vendor/eigenpal/README.md` and the fork docs), placeholder semantics (see `concepts/placeholders.md`), template authoring page UX (see `modules/templates-v2.md`), toolbar overlay + eigenpal CSS overrides (see `modules/editor-chrome.md`), deferred editor backlog items (see `backlog/editor.md`).
> **Key files:**
> - `packages/editor-ui/src/MetalDocsEditor.tsx:49-59` — plugin list build; `templatePlugin` gate on `mode === 'template-draft'`
> - `packages/editor-ui/src/types.ts:5` — `EditorMode` type: `'template-draft' | 'document-edit' | 'readonly'`
> - `packages/editor-ui/src/index.ts` — package public API
> - `packages/editor-ui/src/plugins/OutlinePlugin.tsx` — heading nav (custom MetalDocs plugin)
> - `packages/editor-ui/src/plugins/sidebarModelBridge.ts` — sidebar item bridge for placeholders/etc
> - `packages/editor-ui/src/plugins/mergefieldPlugin.ts` — (legacy? verify)
> - `vendor/eigenpal/README.md` - controlled EigenPal package artifact and refresh command
> - `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` - package artifact consumed by MetalDocs

---

## Stack

- **Eigenpal:** `@eigenpal/docx-js-editor` — DOCX WYSIWYG editor, ProseMirror under the hood.
- **Current package source:** controlled fork artifact vendored at `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`.
- **MetalDocsEditor:** thin React wrapper at `packages/editor-ui/src/MetalDocsEditor.tsx`. Adds:
  - Debounced autosave (1500ms)
  - Plugin registration order (see `templatePlugin` gating below)
  - Imperative `ref` exposing `getDocumentBuffer()` for parent to grab DOCX bytes
- **Consumers:**
  - `frontend/apps/web/src/features/templates/TemplateAuthorPage.tsx` (template authoring, mode=`template-draft`)
  - `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx` (document fill-in/view, mode=`document-edit` when `isEditable`, otherwise `readonly`; non-draft docs also show `PDFCell` via `useDocumentPdfStatus`)

  Both consumers wrap `MetalDocsEditor` inside `EditorChrome` (see `modules/editor-chrome.md`), which owns the toolbar overlay and all eigenpal CSS overrides.

## Package contract

MetalDocs intentionally treats EigenPal as a package dependency, not as application code.

- MetalDocs consumes `@eigenpal/docx-js-editor` from `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`.
- The source for that artifact is the controlled fork documented in `vendor/eigenpal/README.md`.
- Deep implementation details for header/footer tables, body pagination, template overlays, table selection, and DOCX round-trip behavior live in the EigenPal fork docs and local lab dossier.
- The MetalDocs Wiki should only document how the editor is consumed, where it is wired, how to refresh the artifact, and how to validate the integration.
- Do not patch `node_modules`, reintroduce frontend-only `pnpm patch` files, or duplicate EigenPal fork internals here.

The practical rule: if the question is "how does MetalDocs use the editor?", document it here. If the question is "how does EigenPal render or serialize DOCX internals?", document it in the fork.

## Plugin registration

`MetalDocsEditor.tsx:55-59`:
```ts
const plugins: EditorPlugin[] = [
  ...(props.mode === 'template-draft' ? [templatePlugin] : []),    // eigenpal native — only in template authoring
  ...(props.sidebarModel ? [buildSidebarModelPlugin(props.sidebarModel)] : []),  // sidebar bridge
  ...(props.externalPlugins ?? []),                                 // page-specific extras (e.g., filterTransactionGuard)
];
```

Order matters: plugins later in the array can react to earlier plugins' state.

### `templatePlugin` mode gating

`templatePlugin` is now included **only when `mode === 'template-draft'`**. It is skipped for `document-edit` and `readonly` modes.

**Rationale:** `templatePlugin` injects `template-annotation-chip` items into eigenpal's `docx-unified-sidebar`. In template authoring this is the desired behaviour — authors see the token list alongside the canvas. In document editing, documents contain fully-substituted output (no live `{tokens}`), so the sidebar chips are meaningless. With no chips and no comments open, eigenpal collapses the sidebar, centering the canvas — the correct visual outcome for the document editor.

**Comments are unaffected.** Comments are a built-in eigenpal feature wired through `DocxEditor` props (`comments`, `onCommentAdd`, etc.), not through the plugin system. Removing `templatePlugin` does not disable comments.

**Future consideration:** if a feature ever needs template-style annotations inside the document editor, do not simply re-add `templatePlugin` unconditionally. Use CSS to hide `.template-annotation-chip` items instead, to preserve the gating contract. See `backlog/editor.md` — cross-cutting notes.

## Plugins

### `templatePlugin` (eigenpal native)
Imported from `@eigenpal/docx-js-editor`. Detects docxtemplater tokens (`{name}`, `{#section}`, etc.) and:
- Adds orange decoration to canvas
- Provides sidebar chips
- Exposes `TemplateTag[]` via plugin state

**Status:** Active. MetalDocs now uses `{name}` syntax (post-migration 2026-04-25), so tokens are highlighted orange and listed in the sidebar natively. In template authoring, `TemplateAuthorPage` also reads `editorRef.current.getAgent().getVariables()` after editor changes and auto-syncs schema metadata from detected token names. See `concepts/placeholders.md`.

**`applyVariables` is NOT called in writer mode.** Tokens remain as literal `{name}` strings in the editor DOCX. Substitution occurs server-side at freeze/finalize via the fanout pipeline. Reason: eigenpal autosaves on every change — calling `applyVariables` in-editor would persist substituted values in the DOCX, destroying original tokens. A future "preview mode" (two-buffer design) would allow ephemeral browser-side substitution without affecting the autosaved edit buffer. See `decisions/0008-placeholder-fixed-catalog.md`.

### `outlinePlugin` (custom MetalDocs)
Source: `packages/editor-ui/src/plugins/OutlinePlugin.tsx`. Walks the ProseMirror doc tree, finds paragraphs with heading style (`outlineLevel` attr or `styleId` matching `Título1` / `Heading1`), surfaces them as a left panel for navigation.

**Status:** Not currently in the plugins array in `MetalDocsEditor.tsx` (removed in the 2026-05-06 plugin-registration refactor). The source file still exists. Eigenpal's own `docx-outline-nav` button still appears in the canvas via eigenpal internals — the MetalDocs outline panel on top of it is dormant.

**Spike origin:** Verified in eigenpal-spike T7. Module-level `cachedDoc` singleton bug (breaks with multiple editor instances) was fixed at port time via factory pattern + `useMemo` per instance.

Toggle: button `docx-outline-nav` injected by eigenpal at top-left of editor. Click opens/closes the panel.

### `sidebarModelBridge` (custom MetalDocs)
Source: `packages/editor-ui/src/plugins/sidebarModelBridge.ts`. Optional. When the parent passes `sidebarModel` prop, this plugin renders MetalDocs-specific sidebar items (placeholders/etc) inside eigenpal's sidebar slot.

### `mergefieldPlugin` (status: VERIFY)
Source: `packages/editor-ui/src/plugins/mergefieldPlugin.ts`. Loaded by Vite (per network log) but not in the plugins array of `MetalDocsEditor.tsx`. May be legacy or invoked elsewhere. **Action item:** confirm whether to remove or document its real entry point.

### `filterTransactionGuard` (page-specific)
`frontend/apps/web/src/editor-adapters/filter-transaction-guard.ts`. Passed as `externalPlugins` from `TemplateAuthorPage`. Filters specific transactions to prevent unwanted edits in template mode.

## Modes

```ts
// packages/editor-ui/src/types.ts:5
type EditorMode = 'template-draft' | 'document-edit' | 'readonly';
```

MetalDocs uses three modes instead of eigenpal's two (`editing` / `viewing`):

| MetalDocs mode | eigenpal `mode` | `templatePlugin` | Autosave | Consumer |
|---|---|---|---|---|
| `template-draft` | `editing` | included | yes | `TemplateAuthorPage` |
| `document-edit` | `editing` | **skipped** | yes | `DocumentEditorPage` (writer session) |
| `readonly` | `viewing` | **skipped** | no | `DocumentEditorPage` (no writer session) |

Eigenpal's `outlinePlugin` is no longer in the plugin array (removed in the same pass). The outline nav button shipped by eigenpal itself is still available inside the canvas.

## Autosave

`MetalDocsEditor.tsx:30-47`. On every editor `onChange`:
1. Debounce 1500ms (`AUTOSAVE_DEBOUNCE_MS`)
2. Skip if previous save still in flight (`inFlightRef`)
3. Call `inner.current.save()` → returns DOCX `Uint8Array | null`
4. Pass buffer to parent via `props.onAutoSave(buf)`
5. Parent uploads to API/S3

Parent is responsible for handling failures + retry. Editor doesn't surface save state — parent does (via `AutosaveStatus` in the `EditorChrome` right slot; see `modules/editor-chrome.md`).

## Imperative ref

```ts
type MetalDocsEditorRef = {
  getDocumentBuffer(): Promise<Uint8Array | null>;
  focus(): void;
}
```

Used by parent to:
- Grab DOCX bytes on demand (e.g., for download, manual save trigger)
- Focus editor programmatically (no-op currently)

## Layout

The eigenpal `DocxEditor` renders inside `PluginHost`:
```
┌─ ep-root.docx-editor ─────────────────────────────────────────┐
│ ┌─ toolbar (z-50) ─────────────────────────────────────────┐ │
│ │ File  Format  Insert  ...                                │ │
│ └──────────────────────────────────────────────────────────┘ │
│ ┌─ paged-editor ──────────────────────────────────────────┐  │
│ │ ┌─ paged-editor__hidden-pm (.ProseMirror) ────────────┐ │  │
│ │ │ [actual editable ProseMirror]                       │ │  │
│ │ └──────────────────────────────────────────────────────┘ │  │
│ │ ┌─ rendered pages ────────────────────────────────────┐ │  │
│ │ │ [paginated visual rendering]                        │ │  │
│ │ └──────────────────────────────────────────────────────┘ │  │
│ │ [image-selection-overlay] [decoration-overlay]         │  │
│ └─────────────────────────────────────────────────────────┘  │
│ [docx-outline-nav button — top-left, fixed position]        │
└─────────────────────────────────────────────────────────────┘
```

`.ProseMirror` is the actual editable element. Reach via `document.querySelector('.ProseMirror')` for tests/debugging. Has `pmViewDesc` property exposing the node hierarchy.

## ProseMirror access

The editor doesn't expose its `EditorView` directly. To do programmatic edits:
- Synthetic `KeyboardEvent` does NOT work (PM filters)
- `document.execCommand('insertText' | 'selectAll' | 'delete')` DOES work
- `ClipboardEvent('paste', { clipboardData })` DOES work for HTML paste

## Common pitfalls

1. **`templatePlugin` only detects `{name}` (single brace) and is only active in `template-draft` mode.** MetalDocs migrated to this format (2026-04-25). Legacy `{{uuid}}` templates will not get highlighting. In document editing the plugin is not loaded at all — see "Plugin registration" above. See `concepts/placeholders.md`.
2. **Outline panel won't render until `docx-outline-nav` button is clicked.** It's an eigenpal toggle, not a passive plugin display.
3. **Multiple `MetalDocsEditor` instances** — the spike's outline plugin had a module-level cache bug. Confirmed fixed in our port via factory pattern. If you ever see "second editor sees stale headings", check this regression first.
4. **Autosave race** — parent must handle 409/etag conflicts itself. The editor doesn't track server state.

## Freeze Integration

Eigenpal's headless substitution API is **not** called in the editor (writer mode). Substitution happens exclusively at freeze time, server-side, triggered by the final signoff approval:

1. `FreezeService.Freeze` resolves each catalog token via `resolvers.Registry`.
2. The `{name: value}` map is posted to docgen-v2 via `fanout.Client.Fanout`.
3. docgen-v2 calls eigenpal headless substitution on the stored template DOCX and uploads the result as `frozen.docx`.

For the full pipeline, see [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md).

## Validation checklist

Use this when refreshing the vendored EigenPal artifact or checking that MetalDocs still consumes it correctly.

1. Run `npm run typecheck -w packages/editor-ui`.
2. Run `npm run test -w packages/editor-ui -- --run`.
3. Run `npm run typecheck -w apps/docgen-v2`.
4. Build the web app from `frontend/apps/web` with `npx vite build`.
5. Browser smoke: open template authoring and a DOCX template that uses headers, tables, and placeholders; confirm the editor loads without console errors.

This checklist validates MetalDocs integration only. EigenPal rendering fidelity tests belong in the fork.

## Cross-refs

- [concepts/placeholders.md](../concepts/placeholders.md) — placeholder schema and `{name}` token format
- [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md) — approve → freeze → fanout → PDF artifact
- [modules/editor-chrome.md](editor-chrome.md) — toolbar overlay primitive + eigenpal CSS overrides (used by both consumers)
- [modules/templates-v2.md](templates-v2.md) — TemplateAuthorPage consumer (`template-draft` mode)
- [modules/documents.md](documents.md) — DocumentEditorPage consumer (`document-edit` / `readonly` modes); left-rail layout
- [backlog/editor.md](../backlog/editor.md) — deferred Metadados, Revisões, Aprovadores sidebar items; cross-cutting note on templatePlugin gating
- [references/eigenpal-spike.md](../references/eigenpal-spike.md) — T7 outline plugin origin + caveats
