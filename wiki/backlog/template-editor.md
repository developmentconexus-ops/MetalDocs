# Backlog: Template editor screen (`/templates/:templateId/versions/:n`)

> Last updated: 2026-05-10 (Phase 5 docs after rebuild)

Tracks deferred work from the 2026-05-10 template-editor rebuild. The screen now mirrors `DocumentEditorPage` layout (inner rail 48px + variables/outline panels + EditorChrome + eigenpal canvas). Items below were intentionally cut or deferred during the rebuild — none are bugs.

---

## version-history

**Cut from rebuild — backend gap.** Right-rail "Versões" panel (a la `EditorMetaSidebar` Revisões tab) was not wired because there is no `GET /api/v2/templates/:id/versions` list endpoint today. Only the latest version is returned by `GET /api/v2/templates/:id`.

Backend work:
- Add `GET /api/v2/templates/:id/versions` returning `[{ id, version_number, status, author_id, created_at, published_at }]`.
- Frontend can then add a right sidebar (or modal) listing versions with status pills + jump-to action.

Frontend impact: new component `TemplateVersionTimeline.tsx` (mirror `DocumentVersionTimeline`).

---

## comments

**Cut from rebuild — model gap.** Templates have no comment thread today (documents do via `GET /api/v2/documents/:id/comments`). Reviewer feedback during `in_review` flows through `VersionActionPanel` reason field only.

Decide before implementing:
1. Reuse the document comments table with a `template_version_id` column.
2. Separate `template_review_comments` table.
3. Inline-only annotations via eigenpal native comments (different UX).

---

## outline-future-enhancements

**Phase 3c shipped a read-only outline panel.** Headings derived from `agent.getAgentContext().outline` (eigenpal-filtered `ParagraphOutline[]` with `isHeading` + `headingLevel`). See `lib/readHeadings.ts`.

Future iterations:
- Click a heading to scroll/select that paragraph in the editor (needs `agent.scrollToParagraph(index)` or equivalent surface — verify in eigenpal types).
- Drag-to-reorder sections (large lift; coupling to eigenpal mutation API).
- Persist `leftActive` panel choice across reloads via `localStorage` (lazy initializer).

---

## design-toolbar-parity (Decision A — won't fix)

The Phase 0 design HTML drew a custom React toolbar (undo/redo/B/I/U/lists/etc.) sitting above the canvas. We **kept eigenpal's built-in toolbar** instead — same call as `DocumentEditorPage`. Rationale:

- Eigenpal owns selection/inspector internals; reimplementing the toolbar would require reverse-engineering its command surface and re-validating against every eigenpal upgrade.
- Two stacked toolbars = bad UX.
- Live document editor ships eigenpal's toolbar already; consistency wins.

If we ever swap eigenpal out, revisit. Until then: do not flag absence of the design toolbar as a defect.

---

## placeholder-catalog-panel-restyle

`PlaceholderCatalogPanel.tsx` still uses inline styles with raw hex/px (predates the tokens-only rule). Functional, but drifts from `wiki/architecture/frontend-structure.md`. Port to a CSS Module + tokens to match `TemplateOutlinePanel.module.css`. No data-shape change, low risk. Defer until next templates iteration.

---

## convergence-test-rewrite

`features/templates/__tests__/template-author-page-convergence.test.tsx` is currently `describe.skip` — pre-existing red on `main` (4/5 fails before rename) because mock paths still pointed at the dissolved `v2/` dir (commit b1e7ae00) and the `useFakeTimers` flow predates the new outline + ApiError effects.

Fix: rewrite as a TanStack Query integration test using `QueryClientProvider` + real `fetchPlaceholderCatalog` mocked at the network layer (`vi.spyOn(global, 'fetch')`), drop fake timers, assert on visible DOM via `findByTestId`. Keep the assertions about catalog rendering, detected-token marking, and computed-placeholder save.

---

## submitForReview server-error UX

`submitForReview` now throws `ApiError`; the page surfaces the message via `resolveErrorMessage`. If we ever expand the error catalogue (e.g. `template.review.no_pending_role`, `template.review.docx_missing`), add the codes to `lib/api/errors.ts` so users see pt-BR copy instead of the raw backend message.
