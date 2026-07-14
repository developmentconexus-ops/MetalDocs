# ADR 0001: Adopt eigenpal as the document editor

> **Status:** Accepted (amended 2026-06-23 — v1.9.0 adoption, vendored fork retired)
> **Last verified:** 2026-06-23
> **Date:** ~2026-04 (verify from git log)
> **Scope:** Editor library choice for MetalDocs WYSIWYG.

## Context

We needed a DOCX-native WYSIWYG editor in the browser. Candidates:
- **CKEditor 5** — mature but HTML-first, requires DOCX↔HTML conversion (lossy)
- **BlockNote** — modern but block-model, mismatch with DOCX paragraph model
- **eigenpal/docx-editor-react** — native DOCX, ProseMirror under the hood, MS-Word-like UX (published as `@eigenpal/docx-editor-react`; at spike time referenced as `@eigenpal/docx-js-editor`)

## Decision

**Adopt `@eigenpal/docx-js-editor` as the MetalDocs DOCX editor.**

As of 2026-05-01, MetalDocs consumed a controlled EigenPal package artifact from `third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`. This kept the application dependency deterministic while EigenPal fixes were maintained in the fork and prepared for upstream/published-package consolidation.

> **Resolution (2026-06-14):** The tarball lived at the canonical, app-neutral, Go-safe home `third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`. All three consumers (`apps/docx-renderer`, `packages/editor-ui`, `frontend/apps/web`) referenced it via `file:` and both lockfiles were regenerated. A fresh checkout installed cleanly. HS-2 closed. See `docs/superpowers/specs/2026-06-14-eigenpal-vendor-path-design.md`.

> **Amendment (2026-06-23 — v1.9.0 adoption):** The vendored fork has been retired. MetalDocs now consumes the upstream published package `@eigenpal/docx-editor-react@1.9.0` directly from the npm registry. The tarball `third_party/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` has been deleted; `third_party/eigenpal/NOTICE` records the transition. Import layout changed:
> - Main API (`DocxEditor`, `DocxEditorRef`, `createEmptyDocument`, `EditorMode`): `from '@eigenpal/docx-editor-react'`
> - Plugin API (`PluginHost`, `templatePlugin`, `EditorPlugin`, `ReactSidebarItem`): `from '@eigenpal/docx-editor-react/plugin-api'`
> - CSS: `@eigenpal/docx-editor-react/styles.css`
> - `Comment` type: `from '@eigenpal/docx-editor-core/types/content'`
> - Headless (docx-renderer): `processTemplateDetailed` from `@eigenpal/docx-editor-core/headless`
>
> The ACL still holds: wrapper is `packages/editor-ui/`; all consumers import from `@metaldocs/editor-ui` only. Refresh procedure simplified: bump version in `package.json` and reinstall from registry — no tarball management.

## Reasoning

1. **DOCX round-trip integrity** — verified T2: load → edit → save with no loss
2. **Plugin extensibility** — built on ProseMirror, plugin API is mature
3. **Built-in features we need:** comments, track changes, outline nav, find/replace, paged rendering, table of contents
4. **Template/substitution path:** eigenpal ships `templatePlugin` + `docxtemplaterPlugin` natively (T4)
5. **Active development** — vs CKEditor where DOCX support is plugin-grade and BlockNote where it's absent

## Trade-offs accepted

- Eigenpal is newer / smaller community than CKEditor
- Some features still gaps:
  - **T1 restricted editing** — eigenpal ignores Word's `<w:permStart/End>` XML. Workaround: MetalDocs zones (custom) — abandoned 2026-04-25.
  - **T6 metadata plugin** — partial; MetalDocs uses toolbar instead.
- Token format diverges if not used as native (`{name}` vs MetalDocs `{{uuid}}` legacy) — see ADR 0003.

## Consequences

- All editor-related code consolidates in `packages/editor-ui/`
- EigenPal implementation details stay in the EigenPal fork; MetalDocs only documents the integration contract.
- Dependency refreshes must update `@eigenpal/docx-editor-react` version in `package.json` files and lockfiles together (no tarball management required as of 2026-06-23).
- CKEditor + BlockNote deps removed (purge plan: see `decisions/0002-zone-purge.md` companion notes)
- Future work: leverage native eigenpal capabilities instead of reinventing
- ProseMirror DOM access patterns documented for tests/debugging

## Verification

- Spike T1–T8 all run + reviewed (`references/eigenpal-spike.md`)
- Production usage: `TemplateEditorPage` (renamed from `TemplateAuthorPage` 2026-05-10), `DocumentEditorPage`

## Affected modules

- `TemplateEditorPage` — eigenpal in `template-draft` mode; `templatePlugin` detects `{name}` tokens. See `wiki/modules/editor-ui-eigenpal.md`.
- `DocumentEditorPage` — eigenpal in `document-edit` / `readonly` mode (no `templatePlugin`). See `wiki/modules/documents.md`.
- `templates` — backend module whose authoring constraints (placeholder syntax, editor wiring) are grounded in this decision. See [`wiki/modules/templates.md`](../modules/templates.md).

## Cross-refs

- [references/eigenpal-spike.md](../references/eigenpal-spike.md)
- [references/eigenpal-controlled-package.md](../references/eigenpal-controlled-package.md)
- [modules/editor-ui-eigenpal.md](../modules/editor-ui-eigenpal.md)
- [modules/templates.md](../modules/templates.md) — backend module that authors against eigenpal (§2 Architecture Constraints)
- [decisions/0002-zone-purge.md](0002-zone-purge.md)
- [decisions/0003-token-syntax-migration.md](0003-token-syntax-migration.md)
