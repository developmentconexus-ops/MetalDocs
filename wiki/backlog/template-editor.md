# Backlog: Template editor screen (`/templates/:templateId/versions/:n`)

> Last updated: 2026-05-17 (v1 route memory sync)

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

---

## Integration Audit (2026-05-17)

Plan 12.5 audit run after the required gates:

- `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/templates` -> PASS
- `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module templates` -> PASS

Classification rule used here: runtime truth + OpenAPI/generated types + current frontend wrappers/hooks + module wiki truth. No mock-only surface is treated as implementable.

| Item | Source | Runtime/API reality | Frontend reality | Classification | Action |
|---|---|---|---|---|---|
| Back button inner rail | current screen + design | No backend dependency; route ownership for `/templates/:templateId/versions/:versionNum` is real | Implemented in `TemplateEditorPage`; local callback/nav only | implemented and aligned | Keep in screen slice |
| Variables rail toggle | current screen + design | No backend dependency | Implemented with local `leftActive` state | implemented and aligned | Keep in screen slice |
| Outline rail toggle | current screen + design + backlog | No backend endpoint needed; eigenpal adapter exposes enough local document context for read-only headings | Implemented via `readHeadings` + `TemplateOutlinePanel` | implemented and aligned | Keep; leave advanced behavior deferred |
| Placeholder catalog panel | design + backlog + module wiki | `GET /api/v1/templates/placeholder-catalog` exists in runtime, OpenAPI, generated backend, and generated frontend types | Visible and functional, but fetched twice (`TemplateEditorPage` + `PlaceholderCatalogPanel`), and panel styling is still inline/raw | implemented but legacy-wired | Normalize fetch ownership and restyle in screen PR |
| Template name / version / status chrome | design + current screen | `GET /api/v1/templates/{id}` and `GET /api/v1/templates/{id}/versions/{n}` exist in runtime/spec/codegen | Rendered, but `useTemplateDraft` loads via direct `fetch` wrappers instead of canonical client/query flow | implemented but legacy-wired | Move detail loading to canonical wrapper/query slice |
| Autosave status indicator | design + current screen + editor-chrome wiki | Autosave presign/commit endpoints exist and are mounted/generated | Visible and working, but template editor only maps `idle/saving/saved/error`; no richer cache/session state and hook is custom direct-fetch logic | implemented but legacy-wired | Keep behavior; normalize hook/wrapper path in API/query slice |
| Import `.docx` action | current screen + module wiki | Autosave presign/commit runtime is real; Plan 12.4 proved imported DOCX handoff | Implemented with direct `fetch` wrapper path and local refetch | implemented but legacy-wired | Keep; fold into wrapper/query normalization slice |
| Submit for review CTA | design + current screen + backlog | `POST /api/v1/templates/{id}/versions/{n}/submit` exists in runtime/spec/generated backend/frontend types | Implemented and user-visible; uses `apiFetch`, but page still owns bespoke local state/error banner logic | implemented but legacy-wired | Keep; fold error mapping cleanup into screen/API slice |
| Success/error alert banner | current screen | No backend dependency beyond mutation outcomes | Implemented in `EditorChrome.alert`; copy is pt-BR and honest | implemented and aligned | Keep in screen slice |
| Eigenpal canvas + built-in toolbar | design + design notes + editor-ui docs | Real editor capability exists; module wiki explicitly treats custom toolbar parity as a won't-fix | Implemented through `MetalDocsEditor` adapter in `template-draft` or `readonly` mode | implemented and aligned | Preserve as-is; do not reintroduce design toolbar |
| Draft vs readonly editor state | current screen + module wiki | Template version status is real; non-draft versions are not editable in current product flow | Implemented by `isDraft ? 'template-draft' : 'readonly'` | implemented and aligned | Keep in screen slice |
| Review / approve / published footer panel | current screen + backlog + module wiki | Review/approve routes exist in runtime/spec/generated backend types | Visible for non-draft states, but `VersionActionPanel` uses inline styles and direct `fetch` wrappers | implemented but legacy-wired | Restyle + canonical wrapper cleanup in screen/API slice |
| `version-history` backlog item | backlog + templates wiki | No `GET /api/v1/templates/{id}/versions` list endpoint exists today; only latest template and specific version fetches are available | No timeline/modal/sidebar is wired | missing backend capability | Keep deferred; requires backend/API prerequisite before UI |
| `comments` backlog item | backlog + templates wiki | No template comments model or endpoint exists; document comments are separate | No comments UI is wired | missing backend capability | Keep deferred; requires backend/API prerequisite before UI |
| `outline-click-to-scroll` | backlog `outline-future-enhancements` | No confirmed public editor adapter method for scroll/select targeting is exposed in current `MetalDocsEditorRef` | Current outline panel is read-only | defer | Defer until adapter exposes a stable navigation surface |
| `outline-drag-to-reorder` | backlog `outline-future-enhancements` | No supported backend/editor mutation contract exists for section reorder | No UI present | defer | Defer; would need a separate editor capability decision |
| `outline-panel-persistence` | backlog `outline-future-enhancements` | No backend dependency | Not implemented; local `leftActive` resets on reload | screen-local integration fix | Include only if it fits after core cleanup; otherwise keep deferred |
| `design-toolbar-parity` | backlog + design notes | Product/runtime truth is eigenpal toolbar; module docs already record Decision A won't-fix | No custom React toolbar is rendered | defer | Keep documented as intentional non-goal |
| `placeholder-catalog-panel-restyle` | backlog + current screen | Backend capability exists and is stable | Panel still uses inline styles and duplicate local fetch | screen-local integration fix | Include in screen PR |
| `convergence-test-rewrite` | backlog + skipped test | No backend blocker; existing screen behaviors are real | `template-author-page-convergence.test.tsx` remains `describe.skip` and predates current wiring | screen-local integration fix | Rewrite as current integration test in verification slice |
| `submitForReview-error-codes` | backlog | Current generic error path works; no expanded template-specific code catalogue is documented yet | `resolveErrorMessage` fallback works, but code-specific pt-BR mapping is not complete | defer | Leave deferred unless backend exposes stable new codes during implementation |

### Ready for implementation

- Inner rail navigation and panel toggles
- Read-only outline panel as currently shipped
- Editor chrome title/version/status/autosave/import/submit surface
- Eigenpal canvas with native toolbar
- Review/approve footer panel
- Placeholder catalog restyle and fetch ownership cleanup
- Canonical frontend API/wrapper cleanup for editor-facing template routes
- Convergence test rewrite

### Prerequisites

- No blocking runtime prerequisite after the two required gates passed.
- No blocking shared contract prerequisite was found for Plan 12.5.
- Backend/API prerequisite for any real version-history UI.
- Backend/API prerequisite for any real template comments UI.

### Deferred

- Version history UI until a template versions list endpoint exists
- Template comments until a template comment model/endpoint exists
- Outline click-to-scroll until the editor adapter exposes a stable navigation command
- Outline drag-to-reorder
- Design-toolbar parity (intentional won't-fix)
- Submit-for-review error-code polish beyond the current generic error mapping

### Verification needed next

- `cd frontend/apps/web; pnpm.cmd tsc --noEmit -p tsconfig.build.json`
- `cd frontend/apps/web; pnpm test -- --runInBand src/features/templates`
- Manual smoke on `/templates/:templateId/versions/:versionNum` covering draft, in_review/approved, and imported-DOCX cases

### Proposed implementation slices

1. `screen-local chrome/panels`
   - Ownership: `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx`, `TemplateEditorPage.module.css`, `PlaceholderCatalogPanel.tsx`, `TemplateOutlinePanel.tsx`
   - Scope: panel polish, fetch ownership cleanup, tokenized styles, local state cleanup

2. `template editor API/query normalization`
   - Ownership: `frontend/apps/web/src/features/templates/api/templates.ts`, `api/catalog.ts`, `hooks/useTemplateDraft.ts`, `hooks/useTemplateAutosave.ts`, `hooks/useTemplateSchemas.ts`
   - Scope: move editor-facing data paths onto canonical client/error handling and reduce direct `fetch` drift

3. `workflow actions and verification`
   - Ownership: `frontend/apps/web/src/features/templates/VersionActionPanel.tsx`, `__tests__/template-author-page-convergence.test.tsx`, related template API tests
   - Scope: footer action polish, current-flow test rewrite, regression coverage for placeholder/submit states
