# Phase 0 — Context Load

**Date:** 2026-05-10
**Module:** editor-chrome
**Scope decision:** **Frontend-only**. No backend companion. Module is the React shell that overlays a custom toolbar around the eigenpal editor canvas.

## Inputs read

- `wiki/README.md` — index. editor-chrome listed under Modules; stub entry exists.
- `wiki/modules/editor-chrome.md` — existing stub (175 lines). Already has slot API, parts, eigenpal overrides, consumers. Living doc upgrade target.
- `wiki/modules/editor-ui-eigenpal.md` — sibling module: `MetalDocsEditor` wrapper around `@eigenpal/docx-js-editor`. editor-chrome wraps its output.
- `wiki/concepts/placeholders.md` — fixed 7-token catalog; tokens stay literal in editor; substitution at freeze. editor-chrome surfaces tokens via the eigenpal canvas, not its own UI.
- `wiki/decisions/0001-eigenpal-adoption.md` — ADR for adopting eigenpal. editor-chrome is downstream consequence.

## Module shape (preview from glob)

```
frontend/apps/web/src/features/shared/components/editor-chrome/
  EditorChrome.tsx           # main component + slot API + re-export
  EditorChrome.module.css    # wrapper + overlays + button/text primitives + eigenpal overrides
  index.ts                   # barrel
  parts/
    VersionBadge.tsx
    VersionBadge.module.css
    AutosaveStatus.tsx
    AutosaveStatus.module.css
```

7 files total. No backend, no migrations, no SQL.

## Phase adjustments (FE-only module)

- **Phase 4 (Persistence map):** n/a. Will record "n/a — frontend module, no persistence" in `04-persistence.md`. No SQL, no migrations, no tripwire, no GUC.
- **Phase 2 (Data-flow):** ops are pure-FE interactions, not HTTP → handler → DB. Selected ops:
  1. **Mount + slot composition** — how `EditorChrome` is rendered by a consumer page (TemplateEditorPage path).
  2. **Autosave status rendering** — `AutosaveStatus` state driven by parent's autosave lifecycle (state prop → visual).
  3. **Eigenpal CSS override scope** — how `.wrapper :global(.ep-root)` overrides reach eigenpal DOM at runtime.
- **Phase 1 (Surface):** public exports from `index.ts` barrel; no HTTP routes; no migrations.

## Cross-deps to capture (Phase 3)

IN-edges expected:
- `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx`
- `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`

OUT-edges expected:
- `styles/tokens.css` (design tokens)
- eigenpal DOM via global selectors (`:global(.ep-root)` — not a JS import, a CSS coupling)
- `components/ui/StatusPill` (consumed BY pages, not by EditorChrome itself — verify)

## ADR / concept anchors

- ADR 0001 (eigenpal-adoption) — chrome exists because eigenpal's native title bar is insufficient for MetalDocs branding/UX
- `concepts/placeholders.md` — chrome does not render placeholders; it wraps the eigenpal canvas where placeholders live
- `wiki/architecture/frontend-structure.md` — `features/shared/` rule (2+ consumers ⇒ extract)

## Open questions to resolve in later phases

1. Does any test file (`*.test.tsx`) cover the chrome primitive? → check in Phase 1.
2. `mergefieldPlugin` mention in editor-ui-eigenpal — unrelated to chrome, skip.
3. Is `editorChromeStyles` re-export consumed by anything beyond the two pages? → check in Phase 3.
4. Does `AutosaveStatus` have a11y semantics (aria-live)? → audit in Phase 6.75 self-review.

## Out of scope (explicit)

- Eigenpal internals — owned by `modules/editor-ui-eigenpal.md` and the EigenPal fork
- Template authoring business logic — `modules/templates_v2.md` (backend) + `modules/templates-v2.md` (frontend, retired pending R-100)
- Document editor business logic — `modules/documents.md`
- StatusPill component — owned by `frontend-primitives.md` story (verify which module)

## No skipped gates

Phase 4 is recorded as `n/a` with rationale, not silently skipped.
