# Phase 3 — Cross-deps

> **Last verified:** 2026-06-23
> Scope: imports IN to `packages/editor-ui/`, imports OUT, runtime callers.

## OUT-edges (runtime)

| Source | Target | Purpose |
|---|---|---|
| `MetalDocsEditor.tsx:2` | `@eigenpal/docx-editor-react` — `DocxEditor`, `DocxEditorRef`, `createEmptyDocument`, `EditorMode` | Editor surface + blank-document seed |
| `MetalDocsEditor.tsx:3` | `@eigenpal/docx-editor-react/plugin-api` — `PluginHost`, `templatePlugin`, `EditorPlugin`, `ReactSidebarItem` | Plugin host + template plugin |
| `MetalDocsEditor.tsx:4` | `@eigenpal/docx-editor-react/styles.css` | Eigenpal stylesheet (consumed once at adapter level) |
| `index.ts:3` | `@eigenpal/docx-editor-core/types/content` — `Comment` | Re-exported Comment type for documents module consumers |
| `sidebarModelBridge.ts:1` | `@eigenpal/docx-editor-react/plugin-api` — `EditorPlugin`, `ReactSidebarItem` | Plugin type |
| `OutlinePlugin.tsx:2` | `@eigenpal/docx-editor-react/plugin-api` — `EditorPlugin`, `PluginPanelProps` | Plugin type |
| `mergefieldPlugin.ts:1` | `@metaldocs/shared-tokens` — `diffTokensVsSchema`, `classifyBlacklist`, `ParseError`, `Token` | Token diff math |

Total external runtime deps: 3 (`@eigenpal/docx-editor-react`, `@eigenpal/docx-editor-core`, `@metaldocs/shared-tokens`).

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
| ~~`frontend/apps/web/src/editor-adapters/eigenpal-template-mode.ts:1`~~ | ~~`BlockContent`, `Paragraph`, `Table` from `@eigenpal/docx-editor-core/types/document`~~ | **Fixed 2026-06-23** — now imports via `@metaldocs/editor-ui` (re-exported from `packages/editor-ui/src/types.ts`) |
| `frontend/apps/web/src/editor-adapters/__spike__/eigenpal-placeholder-spike.test.ts:7` | `createEmptyDocument`, `DocumentAgent`, `parseDocx`, `serializeDocx` from `@eigenpal/docx-editor-react` or `@eigenpal/docx-editor-core/headless` | Test/spike only |
| `frontend/apps/web/src/features/templates/__tests__/template-author-page-convergence.test.tsx` | eigenpal mocks | Test-only |

Template runtime pages consume `@metaldocs/editor-ui`; no production page mounts eigenpal directly.

## docx-renderer OUT-edge

`apps/docx-renderer/package.json` declares `@eigenpal/docx-editor-react` (server-side headless substitution via `processTemplateDetailed` from `@eigenpal/docx-editor-core/headless`). `@eigenpal/docx-editor-core` is pulled only **transitively** through `@eigenpal/docx-editor-react` — it is a phantom dep tracked by ADR 0046; no direct declaration exists. That is OUT of the editor-ui package scope but shares the same npm-registry dependency — no longer coupled via the vendored tarball (tarball retired 2026-06-23).

## Build wiring

- `packages/editor-ui/package.json` — `@eigenpal/docx-editor-react: 1.9.0` (installed from npm registry; vendored tarball retired 2026-06-23)
- `packages/editor-ui/package.json:5` — `main: ./src/index.ts` (no compile step shipped; consumers compile source via path alias)
- TS path alias + Vite alias resolve `@metaldocs/editor-ui` to source directly.

## 2026-06-23 sync note

Eigenpal migration: `@eigenpal/docx-js-editor@0.2.0` (vendored tarball) retired. All OUT-edges updated to new package names: main API from `@eigenpal/docx-editor-react`, plugin API from `@eigenpal/docx-editor-react/plugin-api`, `Comment` type from `@eigenpal/docx-editor-core/types/content`. `processTemplateDetailed` (docx-renderer) from `@eigenpal/docx-editor-core/headless`.

## 2026-05-17 sync note

Blank editable mount behavior added adapter OUT-edge for `createEmptyDocument`. Now re-exported from `@eigenpal/docx-editor-react` directly (no longer a separate `/core` subpath for this symbol). Affected consumers are both runtime `MetalDocsEditor` mounts; no API, DB, or backend dependency changed.
