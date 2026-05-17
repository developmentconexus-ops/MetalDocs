# Phase 3 — Cross-deps

> Scope: imports IN to `packages/editor-ui/`, imports OUT, runtime callers.

## OUT-edges (runtime)

| Source | Target | Purpose |
|---|---|---|
| `MetalDocsEditor.tsx:2` | `@eigenpal/docx-js-editor` — `DocxEditor`, `PluginHost`, `templatePlugin`, `DocxEditorRef`, `EditorPlugin` | Editor surface |
| `MetalDocsEditor.tsx:3` | `@eigenpal/docx-js-editor/core` - `createEmptyDocument` | Seeds editable no-buffer mounts with an Eigenpal blank document |
| `MetalDocsEditor.tsx:4` | `@eigenpal/docx-js-editor/styles.css` | Eigenpal stylesheet (consumed once at adapter level) |
| `sidebarModelBridge.ts:1` | `@eigenpal/docx-js-editor` — `EditorPlugin`, `ReactSidebarItem` | Plugin type |
| `OutlinePlugin.tsx:2` | `@eigenpal/docx-js-editor` — `EditorPlugin`, `PluginPanelProps` | Plugin type |
| `mergefieldPlugin.ts:1` | `@metaldocs/shared-tokens` — `diffTokensVsSchema`, `classifyBlacklist`, `ParseError`, `Token` | Token diff math |

Total external runtime deps: 2 (`@eigenpal/docx-js-editor`, `@metaldocs/shared-tokens`).

## IN-edges (consumers of `@metaldocs/editor-ui`)

| Consumer | Imports | Notes |
|---|---|---|
| `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:2` | `MetalDocsEditor`, `MetalDocsEditorRef`, `Comment` | Runtime mount |
| `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx:3` | `MetalDocsEditor`, `MetalDocsEditorRef` | Runtime mount; blank-template editor relies on adapter empty-document seed |
| `frontend/apps/web/src/features/documents/hooks/editor/useDocumentComments.ts:2` | `type { Comment }` | Type-only |
| `frontend/apps/web/src/features/documents/hooks/editor/__tests__/useDocumentComments.add.test.tsx:1` | `type { Comment }` | Type-only |
| `frontend/apps/web/src/features/documents/hooks/editor/__tests__/useDocumentComments.orphan.test.tsx:1` | `type { Comment }` | Type-only |
| `frontend/apps/web/src/features/documents/__tests__/DocumentEditorPage.test.tsx:36` | `MetalDocsEditor` (vi.mock target) | Mock for tests |
| `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.test.tsx:7` | `MetalDocsEditor` (vi.mock target) | Mock for tests |
| `frontend/apps/web/vite.config.ts:36` | alias `@metaldocs/editor-ui` → `packages/editor-ui/src/index.ts` | Build wiring |
| `frontend/apps/web/tsconfig.json:21` | TS path alias | Type-resolution wiring |

## IN-edges (eigenpal direct, bypassing the wrapper)

| Consumer | Imports | Drift severity |
|---|---|---|
| `frontend/apps/web/src/editor-adapters/eigenpal-template-mode.ts:1` | `BlockContent`, `Paragraph`, `Table` from `@eigenpal/docx-js-editor/core` | Type-only adapter spike; not a runtime editor mount |
| `frontend/apps/web/src/editor-adapters/__spike__/eigenpal-placeholder-spike.test.ts:7` | `createEmptyDocument`, `DocumentAgent`, `parseDocx`, `serializeDocx` from `@eigenpal/docx-js-editor/core` | Test/spike only |
| `frontend/apps/web/src/features/templates/__tests__/template-author-page-convergence.test.tsx` | eigenpal mocks | Test-only |

Template runtime pages consume `@metaldocs/editor-ui`; no production page mounts eigenpal directly.

## docgen-v2 OUT-edge

`apps/docgen-v2/package.json:15` declares `@eigenpal/docx-js-editor` (server-side substitution at freeze). That is OUT of the editor-ui package scope but shares the same vendored tarball — captured as a coupled-dependency risk in tech-debt T-001.

## Build wiring

- `packages/editor-ui/package.json:29` — `@eigenpal/docx-js-editor: file:../../vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` (path present after Plan 3 restoration)
- `packages/editor-ui/package.json:5` — `main: ./src/index.ts` (no compile step shipped; consumers compile source via path alias)
- TS path alias + Vite alias resolve `@metaldocs/editor-ui` to source directly.

## 2026-05-17 sync note

Blank editable mount behavior adds a new adapter OUT-edge to `@eigenpal/docx-js-editor/core` for `createEmptyDocument`. Affected consumers are both runtime `MetalDocsEditor` mounts; no API, DB, or backend dependency changed.
