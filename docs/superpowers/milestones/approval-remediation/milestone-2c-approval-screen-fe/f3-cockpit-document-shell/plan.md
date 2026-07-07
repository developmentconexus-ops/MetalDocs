# F3 — Plan

Seeded from master plan §F3 + the F3 investigator map (runtime truth). TDD via a fresh implementer
subagent + independent reviewer subagent (milestone Phase 3 step 6). **The implementer uses its own
Read/Edit/Bash tools directly — it does NOT spawn sub-agents.**

## Tasks

1. **[TDD] Failing tests first**
   - `features/documents/components/DocumentShell.test.tsx` (new): mount with a mocked
     `@metaldocs/editor-ui`; assert buffer loading → ready (editor mounts with `documentBuffer`) →
     error branch (`role="alert"`); `chrome` present renders header, absent renders bare editor;
     `onAutoSave` omitted ⇒ editor mounts, no throw.
   - `features/approval/pages/ApprovalCockpitPage.test.tsx` (new, replaces
     `SignoffDetailPage.test.tsx`): mock `@metaldocs/editor-ui`; mock
     `../../documents/hooks/editor/useDocumentSession` + `useDocumentAutosave` with `vi.fn()` and
     assert **call-count 0**; spy `getDocument`/`getInstance`/`fetchActiveDocumentInstance`/
     `listComments`; cases per spec Validation Gate (approval→readonly, review+eligible→review,
     non-eligible→readonly, visibility 404→inline alert no toast). Run RED before impl.
2. **Extract `DocumentShell`** (`features/documents/components/DocumentShell.tsx`): lift the
   buffer-fetch effect (`DocumentEditorPage.tsx:109-154`) + the `MetalDocsEditor` mount
   (`:519-536`) + `editorLoadError` render. `EditorChrome` wrap only when `chrome` supplied.
   Props exactly per spec §1. No `useDocumentSession`/`useDocumentAutosave` import.
3. **Refit `DocumentEditorPage`** to consume the shell: replace the `<main class=canvas>` inner
   region with `<DocumentShell … />`, passing `editorMode`, `chrome={{center,right}}`,
   `onAutoSave={handleSave}`, comment handlers, `onChange`, `editorRef`, `notice={commentsHook
   loadError banner}`. Keep rail + `ArtifactMetaSidebar` + submit dialog in the page. Buffer state
   now lives in the shell — remove the now-dead buffer state/effect from the page (the page reads
   `doc.current_revision_id` and passes it down). **`DocumentEditorPage.test.tsx` must stay green**
   — adjust the test's editor mock wiring only if the mock’s prop surface moved, never to hide a
   regression.
4. **Rename `SignoffDetailPage.tsx` → `ApprovalCockpitPage.tsx`**; add mode resolution:
   - Read the active stage from the adapter's `instance` (`stages.find(s => s.status==='active')`).
   - `resolveEditorMode(activeStage, currentUserId)` → `'review'` iff `stage_kind==='review'` AND
     current user id ∈ `activeStage.actors` with actor `status` ∈ {active,waiting}; `'readonly'`
     otherwise (incl. approval stage, non-eligible, oversee, no active stage). Fail-safe default
     `'readonly'`.
   - Replace the `tab==='documento'` A4 box + `<ReviewDocumentCanvas>` with `<DocumentShell
     documentId currentRevisionId={doc.current_revision_id} editorMode={mode} author={currentUser
     displayName} onTrackedChangesChange={setTrackedChanges} … />` (no `chrome`, no `onAutoSave`).
     Comment handlers from `useDocumentComments(documentId, approverDisplay)` if review mode needs
     comment authoring; otherwise pass read comments only.
   - Delete `canvasRef`, `flushSave`, and the `flushSave` call in `decisionSubmit` (no writable
     canvas to flush). `decisionSubmit` = `signOff(...) → refetchInstanceRef.current()`.
   - Hold `const [trackedChanges, setTrackedChanges] = useState<TrackedChange[]>([])` for F4
     (unused render in F3 — wired, not shown).
   - Keep `ArtifactApprovalScreen` + `decisionExtras` + `dialogs` + Comentários tab untouched.
   - Visibility 404: ensure the `not_found.instance_not_visible` / document `NOT_FOUND` path lands
     on the existing inline `role="alert"` not-found branch (no toast).
5. **Delete** `features/approval/components/ReviewDocumentCanvas.tsx` +
   `ReviewDocumentCanvas.test.tsx`. Delete the old `SignoffDetailPage.test.tsx` (replaced by
   `ApprovalCockpitPage.test.tsx`). Rename the CSS module import if the page module name changes
   (keep `SignoffDetailPage.module.css` or rename to `ApprovalCockpitPage.module.css` — pick one,
   update the import).
6. **Route:** `features/approval/routes.tsx` — update the `approvals/:documentId` lazy import from
   `./pages/SignoffDetailPage` → `./pages/ApprovalCockpitPage` (export name follows).
7. **Verify:** `Grep 'useDocumentSession|useDocumentAutosave' frontend/apps/web/src/features/approval`
   → zero. `vitest run` targeted (DocumentShell + ApprovalCockpitPage + DocumentEditorPage) GREEN.
   `tsc --noEmit -p tsconfig.build.json` clean. Grep the approval feature for `ReviewDocumentCanvas`
   → zero.
8. **Review pass** — independent reviewer subagent (sonnet): spec compliance (shell contract, mode
   table, W2 grep-zero, 404-not-toast), zero-visual-regression for the author page, generated-DTO
   discipline, test non-tautology (mode assertions are real, session-not-called is asserted). Apply
   accepted findings.

## Files touched

- `frontend/apps/web/src/features/documents/components/DocumentShell.tsx` (new)
- `frontend/apps/web/src/features/documents/components/DocumentShell.test.tsx` (new)
- `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx` (refit to shell)
- `frontend/apps/web/src/features/approval/pages/ApprovalCockpitPage.tsx` (renamed from SignoffDetailPage)
- `frontend/apps/web/src/features/approval/pages/ApprovalCockpitPage.test.tsx` (new, replaces SignoffDetailPage.test.tsx)
- `frontend/apps/web/src/features/approval/pages/SignoffDetailPage.module.css` (kept or renamed)
- `frontend/apps/web/src/features/approval/routes.tsx` (import swap)
- **Deleted:** `ReviewDocumentCanvas.tsx`, `ReviewDocumentCanvas.test.tsx`, `SignoffDetailPage.tsx`, `SignoffDetailPage.test.tsx`

## Risks

- **Author visual regression** — the shell extraction must preserve `EditorChrome` + layout exactly.
  Guardrail: `DocumentEditorPage.test.tsx` green + F7/F8 preview eyeball. If the shell's `chrome`
  slot can't reproduce the exact center/right layout, keep the chrome shape identical (two slots).
- **Buffer ownership move** — moving buffer state page→shell can drop the `skipInitialEditorChangeRef`
  dirty-guard or the `artifactMetadata` wiring the author page needs (`onArtifactMetadata` from
  autosave). Keep those page-owned; the shell owns only the buffer fetch + editor mount. Re-verify
  autosave still advances metadata in the author path.
- **junction drift** (memory `fe-node-modules-junction-drift`) — if vitest won't run, the real fix is
  a full `pnpm install` in `frontend/apps/web`; do NOT hack around it. Report the exact error.
- **`review` mode with no autosave** — confirm `MetalDocsEditor` does not require `onAutoSave` in
  `review` mode (ReviewDocumentCanvas passed it, but as a prop; absence should be inert). Verify.
