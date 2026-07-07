# F6 — Plan

Seeded from master plan §F6 + design §8 C5 + F6 investigator map (agentId a2ca75b8356b29829),
**corrected against runtime truth** (no backend markup gate; detection gate bug). TDD via fresh
implementer subagent (sonnet) + independent reviewer subagent. **Implementer uses its OWN tools; does
NOT spawn sub-agents.**

## Ground truth (from investigator — do not re-derive)
- **No backend tracked-changes 409 gate exists** (F0 HS-2 path A defer). Only comment-freeze:
  `ErrFreezeBlockedByUnresolvedComments` (`freeze.go:19,50-55`), fires at freeze. Client gate is primary.
- `changes_requested` is **instance** status only (openapi:6446). Backend reverts doc→`draft` on
  `request_changes` (`review_verdict_service.go:340-347`).
- `approvalInstanceQuery` (`DocumentEditorPage.tsx:363-367`) is `enabled: docStatus==='under_review'`
  — dead for the `changes_requested`-then-draft case. Must broaden.
- F1 ref (`packages/editor-ui/src/types.ts:70-91`): `getTrackedChanges()`, `acceptChange(revisionId)`,
  `rejectChange(revisionId)`, `acceptAllChanges()`, `rejectAllChanges()`, `removeCommentMark(libraryCommentId)`.
  `onTrackedChangesChange` prop on props + threaded by `DocumentShell.tsx:32,134` but **DocumentEditorPage
  never passes it** — unwired.
- Comments: `useDocumentComments(documentID, author)` → `resolve(c)=patchComment(documentID,c.id,{done:true})`;
  comment id = `library_comment_id` = `EditorComment.id`.
- Flush sequence (proven in `submitForReview` `:239-250`): `editorRef.current?.saveNow()` →
  `autosave.queue(buf,…)` → `await autosave.flush()` (false on 409/410/422).
- Submit: `submit(documentId, body, opts)` → `POST /documents/{id}/submit` (`approvalApi.ts:122-131`),
  imported as `submitForReviewRequest` (`DocumentEditorPage.tsx:17`). Submit button disabled at `:473`
  `!canEditContent || isSubmitting`; `canEditContent = phase==='writer' && docStatus==='draft'`.
- `editorRef: useRef<MetalDocsEditorRef>(null)` (`:72`). Sidebar `ArtifactMetaSidebar` (`:498-551`).
- Greenfield: no `RequestedChangesPanel`, no `changes_requested` handling, no existing tests to extend
  beyond the `under_review` ones.

## Ordered tasks

1. **[TDD] Failing tests first** (author all before impl; run RED):
   - `features/documents/components/__tests__/RequestedChangesPanel.test.tsx` (new)
   - Extend `features/documents/pages/DocumentEditorPage.test.tsx` with the 5 gate cases.
   Cases exactly per spec Validation Gate.
2. **Broaden detection** — `DocumentEditorPage`: change `approvalInstanceQuery` `enabled` to
   `documentID.length>0 && (docStatus==='under_review' || docStatus==='draft')`. Handle no-instance
   (404/empty) gracefully → `instance` null, no panel. Derive
   `const changesRequested = approvalInstance?.status === 'changes_requested'`.
3. **Wire tracked-changes state** — add `const [trackedChanges, setTrackedChanges] =
   useState<TrackedChange[]>([])`; pass `onTrackedChangesChange={setTrackedChanges}` to the
   `DocumentShell` mount (`:442-494`). (DocumentShell already threads it — no shell edit.)
4. **`components/RequestedChangesPanel.tsx`** (new) — props `{ trackedChanges, comments,
   onAcceptChange, onRejectChange, onResolveComment }`. Renders: header "Mudanças solicitadas" +
   counts; tracked-change cards (author, type, excerpt) with Aceitar/Rejeitar; unresolved-comment
   list (`comments.filter(c=>!c.resolved)`) with Resolver. Empty/all-clean state ("Nenhuma marcação
   pendente. Você pode reenviar."). PT-BR sentence case, wine tokens, visible focus, a11y labels.
   Pure presentational — actions via callbacks; does NOT own editorRef/mutations.
5. **Mount panel** page-owned in `DocumentEditorPage` (only when `changesRequested`) —
   `onAcceptChange={(id)=>editorRef.current?.acceptChange(id)}`,
   `onRejectChange={(id)=>editorRef.current?.rejectChange(id)}`,
   `onResolveComment={(c)=>commentsHook.resolve(c)}`. Place beside `ArtifactMetaSidebar` / in the
   author frame (NOT a DocumentShell slot). Do not touch the shared shell.
6. **Re-submit gating** — compute `hasUnresolvedTrackedChanges = trackedChanges.length>0` and
   `hasUnresolvedComments = comments.some(c=>!c.resolved)`. Extend submit button `disabled`
   (`:473`) with `|| (changesRequested && (hasUnresolvedTrackedChanges || hasUnresolvedComments))`.
   Inline note/tooltip: "Resolva todas as marcações e comentários antes de reenviar."
7. **Clean-buffer re-submit sequence** — in `submitForReview`, when `changesRequested`, BEFORE the
   existing flush: for each resolved comment call `editorRef.current?.removeCommentMark(c.id)`; then
   the existing `saveNow → queue → flush` (already there) persists the cleaned buffer; then
   `submitForReviewRequest`. Keep the flush-false abort. Assert the call order in test.
8. **Backend comment-freeze 409 (belt-and-suspenders)** — grep the approval HTTP error-code table
   (`internal/modules/documents/approval/http`) for the code mapped from
   `ErrFreezeBlockedByUnresolvedComments`. If a distinguishable `err.code` exists, add an inline
   branch in the submit catch (`:307-320`) mapping it to a problem detail; else leave the generic
   toast (client gate is primary) and note it. No backend change.
9. **Verify:** `grep -n "onTrackedChangesChange" src/features/documents/pages/DocumentEditorPage.tsx`
   → present; panel renders only for `changes_requested`. Targeted `vitest run src/features/documents
   src/features/approval` GREEN; `tsc --noEmit -p tsconfig.build.json` clean.
10. **Review pass** — independent reviewer subagent (sonnet): C5 (panel gated on instance
    `changes_requested`, per-change accept/reject via real ref names, per-comment resolve, re-submit
    gated on clean buffer, removeCommentMark→flush→submit order), detection-gate-bug fixed, shared
    shell untouched, no backend/type changes, no fabricated markup-gate 409, D1/D2 flagged, test
    non-tautology, wine tokens. Apply accepted findings.

## Files touched
- `features/documents/pages/DocumentEditorPage.tsx` (detection broaden, wire onTrackedChangesChange,
  mount panel, gate submit, clean-buffer sequence)
- `features/documents/pages/DocumentEditorPage.test.tsx` (extend)
- `features/documents/components/RequestedChangesPanel.tsx` (new) + `.module.css`
- `features/documents/components/__tests__/RequestedChangesPanel.test.tsx` (new)
- possibly a small selector/util for unresolved-count (inline is fine)

## Risks
- **Detection over-fire** — enabling the instance query for ALL drafts fires `getApprovalInstance`
  on never-submitted drafts. Must handle no-instance (404 visibility) as null, no error, no panel.
  Guardrail: test asserts a plain draft (no instance) shows no panel and no error.
- **Stale-plan markup gate** — do NOT implement or assume a backend tracked-changes 409. Client
  disabled-submit is the gate (D1). Reviewer checks none was fabricated.
- **Shared shell** — panel must be page-owned; `git diff DocumentShell.tsx` MUST be empty.
- **Call-order correctness** — `removeCommentMark` must run BEFORE flush (so the flushed buffer is
  clean), flush BEFORE submit. Assert order in test with a mock call-sequence.
- **removeCommentMark scope** — only for RESOLVED comments (don't strip marks of still-open comments).
- **junction drift** — vitest broken → full `pnpm install` in `frontend/apps/web`; no config hack.
- **onTrackedChangesChange timing** — the editor fires it on mount/edit/accept/reject; ensure the
  panel + gate react to the live set, not a stale snapshot. Accepting/rejecting a change should
  shrink the list and (when empty + comments clean) enable re-submit without a manual refresh.
