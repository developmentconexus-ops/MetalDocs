# EigenPal Controlled Package

> **Last verified:** 2026-06-23
> **Scope:** What MetalDocs needs to know about the EigenPal package adoption.
> **Out of scope:** Internal EigenPal implementation details.
> **Key files:**
> - `third_party/eigenpal/NOTICE` - records the transition from vendored tarball to published npm package
> - `packages/editor-ui/src/MetalDocsEditor.tsx` - React wrapper used by MetalDocs
> - `apps/docx-renderer/package.json` - server-side docgen dependency
> - `packages/editor-ui/package.json` - editor-ui dependency
> - `frontend/apps/web/package.json` - web dependency

---

## Current state

MetalDocs consumes the upstream published EigenPal package:

```text
@eigenpal/docx-editor-react@1.9.0  (npm registry)
@eigenpal/docx-editor-core          (npm registry, for headless + type subpaths)
```

The vendored fork era ended 2026-06-23: `third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` was deleted; `third_party/eigenpal/NOTICE` records the transition. All `package.json` `file:` references have been replaced with npm registry version pins.

## Why no longer vendored

The upstream PR series landed and `@eigenpal/docx-editor-react@1.9.0` was published to the npm registry. MetalDocs can now consume it as a standard dependency, avoiding the tarball management overhead.

## What belongs in MetalDocs docs

- Which package versions MetalDocs consumes.
- Which MetalDocs modules import it.
- How to refresh/reinstall after a new EigenPal release.
- Which smoke checks prove the integration still works.

## What does not belong here

- Header/footer rendering internals.
- Table layout and ProseMirror table command internals.
- DOCX serialization implementation details.
- Historical debugging notes from the EigenPal lab.

## Import layout

| Import | Symbol(s) | Used in |
|---|---|---|
| `@eigenpal/docx-editor-react` | `DocxEditor`, `DocxEditorRef`, `createEmptyDocument`, `EditorMode` | `packages/editor-ui/src/MetalDocsEditor.tsx` |
| `@eigenpal/docx-editor-react/plugin-api` | `PluginHost`, `templatePlugin`, `EditorPlugin`, `ReactSidebarItem` | `packages/editor-ui/src/MetalDocsEditor.tsx`, plugin files |
| `@eigenpal/docx-editor-react/styles.css` | CSS stylesheet | `packages/editor-ui/src/MetalDocsEditor.tsx` |
| `@eigenpal/docx-editor-core/types/content` | `Comment` | `packages/editor-ui/src/index.ts` (re-export) |
| `@eigenpal/docx-editor-core/headless` | `processTemplateDetailed` | `apps/docx-renderer/src/render/fanout.ts` |

## Refresh checklist

1. Bump `@eigenpal/docx-editor-react` and `@eigenpal/docx-editor-core` versions in the relevant `package.json` files.
2. Run `pnpm install` at the MetalDocs root.
3. Commit lockfile updates together with the version bump.
4. Verify CSS overrides in `EditorChrome.module.css` still match eigenpal DOM selectors.
5. Run the editor validation checklist in `wiki/modules/editor-ui-eigenpal.md`.
