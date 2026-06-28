# Backlog: Template editor screen (`/templates/:templateId/versions/:n`)

> Last updated: 2026-06-21 (verify-and-archive sweep; see _cleanup-2026-06-21.md)

Tracks deferred work from the 2026-05-10 template-editor rebuild. The screen now mirrors `DocumentEditorPage` layout (inner rail 48px + variables/outline panels + EditorChrome + eigenpal canvas). Items below were intentionally cut or deferred during the rebuild — none are bugs.

---

## version-history

**Cut from rebuild — backend gap.** Right-rail "Versões" panel (a la `EditorMetaSidebar` Revisões tab) was not wired because there is no `GET /api/v1/templates/:id/versions` list endpoint today. Only the latest version is returned by `GET /api/v1/templates/:id`.

Backend work:
- Add `GET /api/v1/templates/:id/versions` returning `[{ id, version_number, status, author_id, created_at, published_at }]`.
- Frontend can then add a right sidebar (or modal) listing versions with status pills + jump-to action.

Frontend impact: new component `TemplateVersionTimeline.tsx` (mirror `DocumentVersionTimeline`).

---

## comments

**Cut from rebuild — model gap.** Templates have no comment thread today (documents do via `GET /api/v1/documents/:id/comments`). Reviewer feedback during `in_review` flows through `VersionActionPanel` reason field only.

Decide before implementing:
1. Reuse the document comments table with a `template_version_id` column.
2. Separate `template_review_comments` table.
3. Inline-only annotations via eigenpal native comments (different UX).

---

## outline-future-enhancements

**Superseded (SP-0, 2026-06-27):** the custom read-only outline panel (`lib/readHeadings.ts`, `TemplateOutlinePanel.tsx`) was deleted in favor of the eigenpal native `showOutlineButton`. Any future outline enhancement should extend the vendor control, not reintroduce the `getAgent`-based heading reader.

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

## ~~placeholder-catalog-panel-restyle~~ (resolved, SP-0 2026-06-27)

`PlaceholderCatalogPanel.tsx` was deleted and replaced by `AvailableTokensPanel.tsx`, which already uses a CSS Module + design tokens. No longer applicable.

---

## ~~convergence-test-rewrite~~ (resolved, SP-0 follow-up 2026-06-28)

`features/templates/__tests__/template-author-page-convergence.test.tsx` was rewritten as the SP-0 schema-write regression guard (Major-3 in the Grade-A review). The old `describe.skip` test mocked the dissolved `getAgent().getVariables()` interface and asserted the now-removed schema-corruption write path; it is gone.

The replacement is a single negative-assertion test under `QueryClientProvider`: it mocks the editor + `useTemplateSchemas`, fires a body `onChange`, lets the 400ms classify debounce settle, then asserts (a) detected tokens flow into the read-only panel (`data-used="true"`) and (b) `schemaState.save` is **never** called. If a future refactor reattaches a schema write to the editor change handler, the test goes red. Runs green on default heap (~0.5s).

---

## submitForReview server-error UX

`submitForReview` now throws `ApiError`; the page surfaces the message via `resolveErrorMessage`. If we ever expand the error catalogue (e.g. `template.review.no_pending_role`, `template.review.docx_missing`), add the codes to `lib/api/errors.ts` so users see pt-BR copy instead of the raw backend message.

---

## VersionActionPanel inline styles

`VersionActionPanel.tsx` uses inline styles for the review/approve/publish footer panel. Restyle to CSS Module + tokens to match the rest of the screen. Low risk — no data-shape change.

---

