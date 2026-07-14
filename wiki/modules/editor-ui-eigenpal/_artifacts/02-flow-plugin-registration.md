# Phase 2 — Data Flow: Plugin Registration & Mode Gating

> Operation: parent mounts `MetalDocsEditor` → wrapper picks plugin list → eigenpal `PluginHost` consumes.
> Source: `packages/editor-ui/src/MetalDocsEditor.tsx:50-75`

## Trace

1. Parent passes `mode: EditorMode` + optional `documentBuffer` + optional `sidebarModel` + optional `externalPlugins`.
2. **Mode → eigenpal `mode` translation** (line 49):
   ```ts
   const libMode = props.mode === 'readonly' ? 'viewing' : 'editing';
   ```
   `template-draft` and `document-edit` both map to `'editing'` upstream.
3. **Blank editable document seed** (lines 53-56):
   ```ts
   const blankDocument = !props.documentBuffer && props.mode !== 'readonly'
     ? createEmptyDocument()
     : undefined;
   ```
   Persisted DOCX buffers take precedence. Readonly empty mounts are not seeded.
4. **Plugin list build** (lines 55-59):
   ```ts
   [
     ...(props.mode === 'template-draft' ? [templatePlugin] : []),
     ...(props.sidebarModel ? [buildSidebarModelPlugin(props.sidebarModel)] : []),
     ...(props.externalPlugins ?? []),
   ]
   ```
5. `<PluginHost plugins={plugins}>` wraps `<DocxEditor documentBuffer={...} document={blankDocument} ... />`. Eigenpal owns dispatch order from there.

## Gate decisions

| `mode` | `templatePlugin` | Autosave debounce | eigenpal `mode` |
|---|---|---|---|
| `template-draft` | included | yes (1500ms) | `editing` |
| `document-edit` | **skipped** | yes (1500ms) | `editing` |
| `readonly` | **skipped** | **off** (line 31 early-return) | `viewing` |

Editable blank modes also pass `document={createEmptyDocument()}` to avoid Eigenpal's no-document placeholder.

Rationale (per wiki + ADR 0008): document-edit shows fully-substituted output; sidebar `template-annotation-chip` items are meaningless → skipping `templatePlugin` collapses the sidebar and centers the canvas.

## Sidebar model plugin

When `sidebarModel` is passed (currently NO consumer in `frontend/apps/web/` passes it — verified grep), `buildSidebarModelPlugin` renders four optional `ReactSidebarItem` sections:
- `Used fields`, `Missing fields`, `Orphan tokens`, `Errors`

Each section renders `<section><h4>title</h4><ul><li>…</li></ul></section>` via `createElement`. No XSS surface — values are joined as text nodes, not `dangerouslySetInnerHTML`.

## External plugins

`TemplateEditorPage` bypasses the wrapper and constructs its own plugin list `[filterTransactionGuard()]`. If it ever migrates onto `MetalDocsEditor`, it would pass `externalPlugins={[filterTransactionGuard()]}`. Current adapter contract accepts this with no transformation.

## Wiring tests

`packages/editor-ui/test/templatePlugin.wiring.test.tsx` covers plugin mode-gating. `packages/editor-ui/test/MetalDocsEditor.mount.test.tsx` covers the blank editable mount contract, including the invariant that existing `documentBuffer` wins and readonly empty mounts remain unseeded.
