# ADR 0001: Adopt eigenpal as the document editor

> **Last verified:** 2026-05-11
> **Status:** Accepted
> **Date:** ~2026-04 (verify from git log)
> **Scope:** Editor library choice for MetalDocs WYSIWYG.

## Context

We needed a DOCX-native WYSIWYG editor in the browser. Candidates:
- **CKEditor 5** — mature but HTML-first, requires DOCX↔HTML conversion (lossy)
- **BlockNote** — modern but block-model, mismatch with DOCX paragraph model
- **eigenpal/docx-js-editor** — native DOCX, ProseMirror under the hood, MS-Word-like UX

## Decision

**Adopt `@eigenpal/docx-js-editor` as the MetalDocs DOCX editor.**

As of 2026-05-01, MetalDocs consumes a controlled EigenPal package artifact from `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`. This keeps the application dependency deterministic while EigenPal fixes are maintained in the fork and prepared for upstream/published-package consolidation.

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
- Dependency refreshes must update `vendor/eigenpal/`, package manifests, and lockfiles together.
- CKEditor + BlockNote deps removed (purge plan: see `decisions/0002-zone-purge.md` companion notes)
- Future work: leverage native eigenpal capabilities instead of reinventing
- ProseMirror DOM access patterns documented for tests/debugging

## Verification

- Spike T1–T8 all run + reviewed (`references/eigenpal-spike.md`)
- Production usage: `TemplateEditorPage` (renamed from `TemplateAuthorPage` 2026-05-10), `DocumentEditorPage`

## Affected modules

- `TemplateEditorPage` — eigenpal in `template-draft` mode; `templatePlugin` detects `{name}` tokens. See `wiki/modules/editor-ui-eigenpal.md`.
- `DocumentEditorPage` — eigenpal in `document-edit` / `readonly` mode (no `templatePlugin`). See `wiki/modules/documents.md`.
- `templates_v2` — backend module whose authoring constraints (placeholder syntax, editor wiring) are grounded in this decision. See [`wiki/modules/templates_v2.md`](../modules/templates_v2.md).

## Cross-refs

- [references/eigenpal-spike.md](../references/eigenpal-spike.md)
- [references/eigenpal-controlled-package.md](../references/eigenpal-controlled-package.md)
- [modules/editor-ui-eigenpal.md](../modules/editor-ui-eigenpal.md)
- [modules/templates_v2.md](../modules/templates_v2.md) — backend module that authors against eigenpal (§2 Architecture Constraints)
- [decisions/0002-zone-purge.md](0002-zone-purge.md)
- [decisions/0003-token-syntax-migration.md](0003-token-syntax-migration.md)
