# Phase 1 — Surface Scan

> Module root: `packages/editor-ui/`
> Scope: public exports, file tree, no HTTP routes, no migrations.

## File tree (source)

```
packages/editor-ui/
├── package.json
├── tsconfig.json
├── vitest.config.ts
├── src/
│   ├── index.ts                          # public barrel
│   ├── MetalDocsEditor.tsx                # wrapper component (88 LOC)
│   ├── types.ts                           # public types
│   └── plugins/
│       ├── OutlinePlugin.tsx              # createOutlinePlugin() (dormant export)
│       ├── sidebarModelBridge.ts          # buildSidebarModelPlugin()
│       └── mergefieldPlugin.ts            # computeSidebarModel() + SidebarModel type
└── test/
    ├── MetalDocsEditor.mount.test.tsx
    ├── mergefieldPlugin.diff.test.ts
    ├── props.contract.test.tsx
    ├── smoke.test.ts
    └── templatePlugin.wiring.test.tsx     # STALE — asserts gating opposite of current code
```

## Public exports (`src/index.ts`)

| Symbol | Kind | Notes |
|---|---|---|
| `MetalDocsEditor` | React component (forwardRef) | Main wrapper around `DocxEditor` |
| `MetalDocsEditorProps` | type | Adapter contract surface |
| `MetalDocsEditorRef` | type | `{ getDocumentBuffer, focus }` |
| `EditorMode` | type | `'template-draft' \| 'document-edit' \| 'readonly'` |
| `Comment` | type re-export | from `@eigenpal/docx-js-editor` |
| `computeSidebarModel` | function | Diffs tokens vs schema → `SidebarModel` |
| `SidebarModel` | type | `{ used, missing, orphans, bannerError, errorCategories }` |
| `buildSidebarModelPlugin` | function | Builds `EditorPlugin` from `SidebarModel` |
| `createOutlinePlugin` | function | Builds outline `EditorPlugin` — **dormant** (not registered in `MetalDocsEditor`) |

## MetalDocsEditor props surface

`MetalDocsEditorProps` (types.ts:7-27):
- `mode: EditorMode` — discriminator (required)
- `documentBuffer?: ArrayBuffer` — initial DOCX
- `author?`, `documentName?`, `documentNameEditable?`, `onDocumentNameChange?`
- Comments: `comments`, `onCommentsChange`, `onCommentAdd`, `onCommentResolve`, `onCommentDelete`, `onCommentReply`
- `renderTitleBarRight?: () => ReactNode`
- `sidebarModel?: SidebarModel`
- `externalPlugins?: EditorPlugin[]`
- `onAutoSave?: (buf: ArrayBuffer) => Promise<void>`
- `onLockLost?: () => void` — **declared but not wired** in `MetalDocsEditor.tsx`
- `showRuler?: boolean`

## Imperative ref surface

`MetalDocsEditorRef` (types.ts:29-32):
- `getDocumentBuffer(): Promise<ArrayBuffer | null>`
- `focus(): void` — no-op (`MetalDocsEditor.tsx:23`)

## HTTP routes

None. FE adapter package — zero network surface.

## Migrations

None. No persistence.

## Plugin registration order (MetalDocsEditor.tsx:55-59)

```ts
[
  ...(props.mode === 'template-draft' ? [templatePlugin] : []),         // eigenpal native, gated
  ...(props.sidebarModel ? [buildSidebarModelPlugin(props.sidebarModel)] : []),
  ...(props.externalPlugins ?? []),
]
```

## Key constants

- `AUTOSAVE_DEBOUNCE_MS = 1500` (MetalDocsEditor.tsx:7)
- `SIDEBAR_PLUGIN_ID = 'metaldocs-sidebar-model'` (sidebarModelBridge.ts:5)
- `PLACEHOLDER_TAG_PREFIX = 'placeholder:'` — lives in `frontend/apps/web/src/editor-adapters/eigenpal-template-mode.ts:16` (outside this package, but part of the wider adapter surface in repo).

## Eigenpal symbols imported

From `@eigenpal/docx-js-editor`:
- `DocxEditor` (component) · `PluginHost` (component) · `templatePlugin` (instance)
- Types: `DocxEditorRef`, `EditorPlugin`, `Comment`, `ReactSidebarItem`, `PluginPanelProps`, `DocxEditorProps`

No imports from `/core` or `/react` subpaths from inside the package (TemplateEditorPage uses those directly, bypassing the wrapper).
