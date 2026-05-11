# Phase 2 — Data Flow: Plugin Registration & Mode Gating

> Operation: parent mounts `MetalDocsEditor` → wrapper picks plugin list → eigenpal `PluginHost` consumes.
> Source: `packages/editor-ui/src/MetalDocsEditor.tsx:49-62`

## Trace

1. Parent passes `mode: EditorMode` + optional `sidebarModel` + optional `externalPlugins`.
2. **Mode → eigenpal `mode` translation** (line 49):
   ```ts
   const libMode = props.mode === 'readonly' ? 'viewing' : 'editing';
   ```
   `template-draft` and `document-edit` both map to `'editing'` upstream.
3. **Plugin list build** (lines 55-59):
   ```ts
   [
     ...(props.mode === 'template-draft' ? [templatePlugin] : []),
     ...(props.sidebarModel ? [buildSidebarModelPlugin(props.sidebarModel)] : []),
     ...(props.externalPlugins ?? []),
   ]
   ```
4. `<PluginHost plugins={plugins}>` wraps `<DocxEditor ... />`. Eigenpal owns dispatch order from there.

## Gate decisions

| `mode` | `templatePlugin` | Autosave debounce | eigenpal `mode` |
|---|---|---|---|
| `template-draft` | included | yes (1500ms) | `editing` |
| `document-edit` | **skipped** | yes (1500ms) | `editing` |
| `readonly` | **skipped** | **off** (line 31 early-return) | `viewing` |

Rationale (per wiki + ADR 0008): document-edit shows fully-substituted output; sidebar `template-annotation-chip` items are meaningless → skipping `templatePlugin` collapses the sidebar and centers the canvas.

## Sidebar model plugin

When `sidebarModel` is passed (currently NO consumer in `frontend/apps/web/` passes it — verified grep), `buildSidebarModelPlugin` renders four optional `ReactSidebarItem` sections:
- `Used fields`, `Missing fields`, `Orphan tokens`, `Errors`

Each section renders `<section><h4>title</h4><ul><li>…</li></ul></section>` via `createElement`. No XSS surface — values are joined as text nodes, not `dangerouslySetInnerHTML`.

## External plugins

`TemplateEditorPage` bypasses the wrapper and constructs its own plugin list `[filterTransactionGuard()]`. If it ever migrates onto `MetalDocsEditor`, it would pass `externalPlugins={[filterTransactionGuard()]}`. Current adapter contract accepts this with no transformation.

## Stale wiring spec — `templatePlugin.wiring.test.tsx`

The test at `packages/editor-ui/test/templatePlugin.wiring.test.tsx:30` asserts `templatePlugin` is included for `<MetalDocsEditor mode="document-edit" />`. The current production code (line 56) gates it to `template-draft` only. The test mock returns `templatePlugin` as a static stub; under the current source, `mode='document-edit'` yields `data-plugins='0'` (no plugins, no sidebar model), failing `expect(...).toBe('1')`. Flagged for tech-debt register. Severity Major (false-pass risk / spec drift; the test as written would fail or has been silently neutralized).
