# Module: Editor UI (Eigenpal Integration)

> _Changelog: 2026-04-26 — added note that `applyVariables` is NOT used in writer mode (ADR 0008)._
>
> **Last verified:** 2026-05-01
> **Scope:** How MetalDocs wraps `@eigenpal/docx-js-editor`, what plugins are registered, autosave wiring, ProseMirror access patterns.
> **Out of scope:** EigenPal fork internals (see `vendor/eigenpal/README.md` and the fork docs), placeholder semantics (see `concepts/placeholders.md`), template authoring page UX (see `modules/templates-v2.md`).
> **Key files:**
> - `packages/editor-ui/src/MetalDocsEditor.tsx` — main wrapper component
> - `packages/editor-ui/src/types.ts` — props, ref interface
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
  - Plugin registration order
  - Imperative `ref` exposing `getDocumentBuffer()` for parent to grab DOCX bytes
- **Consumers:**
  - `frontend/apps/web/src/features/templates/v2/TemplateAuthorPage.tsx` (template authoring, mode=editing)
  - `frontend/apps/web/src/features/documents/v2/DocumentEditorPage.tsx` (document fill-in/view, mode=editing or readonly)

## Package contract

MetalDocs intentionally treats EigenPal as a package dependency, not as application code.

- MetalDocs consumes `@eigenpal/docx-js-editor` from `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`.
- The source for that artifact is the controlled fork documented in `vendor/eigenpal/README.md`.
- Deep implementation details for header/footer tables, body pagination, template overlays, table selection, and DOCX round-trip behavior live in the EigenPal fork docs and local lab dossier.
- The MetalDocs Wiki should only document how the editor is consumed, where it is wired, how to refresh the artifact, and how to validate the integration.
- Do not patch `node_modules`, reintroduce frontend-only `pnpm patch` files, or duplicate EigenPal fork internals here.

The practical rule: if the question is "how does MetalDocs use the editor?", document it here. If the question is "how does EigenPal render or serialize DOCX internals?", document it in the fork.

## Plugin registration

`MetalDocsEditor.tsx:53–58`:
```ts
const plugins: ReactEditorPlugin[] = [
  templatePlugin,                                                  // eigenpal native — placeholder detection
  ...(props.mode !== 'readonly' ? [outlinePlugin] : []),           // headings nav (custom MetalDocs)
  ...(props.sidebarModel ? [buildSidebarModelPlugin(props.sidebarModel)] : []),  // sidebar bridge
  ...(props.externalPlugins ?? []),                                // page-specific extras (e.g., filterTransactionGuard)
];
```

Order matters: plugins later in the array can react to earlier plugins' state.

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
mode: 'editing' | 'readonly'
```

Maps to eigenpal's `mode: 'editing' | 'viewing'`. Readonly hides the outline panel and disables autosave.

## Autosave

`MetalDocsEditor.tsx:31–48`. On every editor `onChange`:
1. Debounce 1500ms (`AUTOSAVE_DEBOUNCE_MS`)
2. Skip if previous save still in flight (`inFlightRef`)
3. Call `inner.current.save()` → returns DOCX `Uint8Array | null`
4. Pass buffer to parent via `props.onAutoSave(buf)`
5. Parent uploads to API/S3

Parent is responsible for handling failures + retry. Editor doesn't surface save state — parent does (via title bar "Saved" badge in `TemplateAuthorPage`).

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

1. **`templatePlugin` only detects `{name}` (single brace).** MetalDocs migrated to this format (2026-04-25). Legacy `{{uuid}}` templates will not get highlighting. See `concepts/placeholders.md`.
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
- [modules/templates-v2.md](templates-v2.md) — TemplateAuthorPage consumer
- [modules/documents-v2.md](documents-v2.md) — DocumentEditorPage consumer
- [references/eigenpal-spike.md](../references/eigenpal-spike.md) — T7 outline plugin origin + caveats
