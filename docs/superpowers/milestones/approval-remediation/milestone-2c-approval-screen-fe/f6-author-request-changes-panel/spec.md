# F6 — Author request-changes panel + clean-buffer re-submit (C5) — Spec

**Feature:** M2c `approval-screen-fe` / F6. Governing: design spec §8 **C5**
(`docs/superpowers/specs/2026-07-07-approval-remediation-design.md:147`) + master plan §F6, **corrected
against runtime truth** (F6 investigator agentId a2ca75b8356b29829).

> C5 (verbatim): "*autor pós-request-changes: editor opens with "mudanças solicitadas" panel —
> per-change accept/reject (`acceptChangeById`/`rejectChangeById`), per-comment resolve; all resolved
> → re-submit enabled.*"

## Interview record (fail-closed contract discovery)

| # | Question | Resolution |
|---|----------|------------|
| 1 | Does the master plan's F0 backend markup gate exist? | **NO.** `ScanForUnresolvedMarkup` / `ErrUnresolvedTrackedChanges` / `ErrUnstrippedComments` do not exist anywhere (grep zero; F0 evidence.md:21,47-49 records they were **scoped out at F0 as HS-2 path A and deferred**). The ONLY backend freeze gate is **comment-only**: `ErrFreezeBlockedByUnresolvedComments` via `repo.HasUnresolvedInstanceComments` (`freeze.go:19,50-55`), fired at **freeze time**. **Consequence:** tracked-changes cleanliness before re-submit is enforced **client-side only** (disabled submit button + hash-chain integrity), NOT by a parallel backend 409. Runtime beats docs — F6 specs to this. Flagged D1 at HS-1. |
| 2 | How does the author page detect `changes_requested`? | **It can't today — a gate bug.** `changes_requested` is an **instance**-level status (`ApprovalInstanceByDocumentResponse.status`, openapi:6446), never a document status. On a `request_changes` verdict the backend reverts `documents.status → draft` in-tx (`review_verdict_service.go:340-347`). But `approvalInstanceQuery` is `enabled: docStatus === 'under_review'` (`DocumentEditorPage.tsx:363-367`) — so once the doc is back to `draft` the instance query is OFF and the author never sees the instance. F6 must broaden `enabled` to also fire for `draft`, then render the panel only when the fetched `instance.status === 'changes_requested'` (no-instance/other-status → no panel). |
| 3 | Is the F1 accept/reject/resolve API ready? | **Yes.** `MetalDocsEditorRef` (`packages/editor-ui/src/types.ts:70-91`): `getTrackedChanges(): TrackedChange[]`, `acceptChange(revisionId)`, `rejectChange(revisionId)`, `acceptAllChanges()`, `rejectAllChanges()`, `removeCommentMark(libraryCommentId)`. `onTrackedChangesChange?: (changes) => void` on props. **Naming note:** the design doc says `acceptChangeById`/`rejectChangeById`; the REAL ACL-wall names are `acceptChange`/`rejectChange` — F6 uses the real names. |
| 4 | Is tracked-changes state wired on the author page? | **No — unwired.** `DocumentShell` threads `onTrackedChangesChange` (`DocumentShell.tsx:32,134`) but `DocumentEditorPage` never passes it (comment says "F4 consumes"). F6 wires it: `DocumentEditorPage` passes `onTrackedChangesChange={setTrackedChanges}` and holds `trackedChanges` state. |
| 5 | Comment resolve + mark-strip primitives? | `useDocumentComments(documentID, author)` → `resolve(c)` = `patchComment(documentID, c.id, {done:true})` (query.ts:100). Comment id used by `removeCommentMark` = `library_comment_id` = `EditorComment.id` (rowToEditorComment:15-24). Resolving (done:true) satisfies the backend comment-freeze gate; `removeCommentMark` strips the visual mark from the docx buffer (output hygiene, no longer gate-critical since the F0 markup gate doesn't exist). |
| 6 | Buffer-flush primitive? | Already implemented + proven in `submitForReview` (`DocumentEditorPage.tsx:239-250`): `editorRef.current?.saveNow()` → `autosave.queue(buf,…)` → `await autosave.flush()` (returns `false` on 409/410/422). F6's clean-buffer re-submit reuses this exact sequence, run AFTER accept/reject + `removeCommentMark` mutate the buffer. |
| 7 | Where does the panel mount? | Page-owned in `DocumentEditorPage` — **do NOT add a slot to the shared `DocumentShell`** (shared with the cockpit; parallels F4's don't-edit-shared discipline). None of DocumentShell's slots (`chrome.center/right`, `notice`) fit a persistent working panel. Mount `RequestedChangesPanel` in the author page's own frame (sidebar column beside `ArtifactMetaSidebar`, or a panel region in `<main>`). |
| 8 | Submit error handling today? | `submitForReview` catch (`DocumentEditorPage.tsx:307-320`) maps 3 validation codes inline, else generic `toast.error`. If the backend comment-freeze 409 surfaces with a distinguishable `err.code`, F6 maps it to an inline problem detail; if it falls through to generic toast, that's acceptable defense-in-depth (the client gate is primary). Implementer greps the approval HTTP error-code table to find the code. |

## Consumer contract (what F6 must deliver)

**Consumer = the author whose document came back `changes_requested`, reopening it in the editor.**

1. **Detection:** broaden the instance query so the author page fetches the active instance when
   `docStatus === 'under_review' || docStatus === 'draft'`; render the panel **only** when
   `instance.status === 'changes_requested'`. Graceful no-instance (404 visibility / no instance) →
   no panel, no error.
2. **`RequestedChangesPanel`** (new, page-owned): lists (a) `TrackedChange[]` cards — author, type,
   excerpt — each with **Aceitar** / **Rejeitar** → `editorRef.current.acceptChange(revisionId)` /
   `rejectChange(revisionId)`; (b) unresolved comments (`commentsHook.comments` where `!resolved`) —
   each with **Resolver** → `commentsHook.resolve(comment)`. Live-updates off `onTrackedChangesChange`
   + the comments query. Teaching header ("Mudanças solicitadas") + count.
3. **Re-submit gating:** the submit button's `disabled` gains
   `hasUnresolvedTrackedChanges || hasUnresolvedComments`; tooltip/inline note explains why
   ("Resolva todas as marcações e comentários antes de reenviar.").
4. **Clean-buffer re-submit sequence** (on re-submit, in order): for each resolved comment call
   `editorRef.current.removeCommentMark(library_comment_id)` (strip marks) → `saveNow → queue →
   flush` (persist the cleaned buffer) → `submitForReviewRequest(...)`. Abort with the existing
   flush-failure toast if flush returns false.
5. **Backend 409 (comment-freeze) belt-and-suspenders:** if the submit/freeze path returns the
   comment-freeze problem+json, surface it inline (not a generic toast) when a distinguishable code
   exists.

## Non-goals (explicit)

- **No backend/contract changes** — no new tracked-changes freeze gate (F0 deferred it, HS-2 path A;
  D1). The client disabled-submit gate + hash chain + existing comment-freeze gate are the enforcement.
- **No `DocumentShell` slot changes** — panel is page-owned (shared shell untouched).
- **No editor-engine work** — accept/reject/resolve/removeCommentMark are F1 primitives, reused.
- **No cockpit/reviewer changes** — F6 is the author side only.
- **No full a11y audit** (F7); F6 adds visible focus on the new panel controls only.

## Validation Gate

- **New:** `RequestedChangesPanel.test.tsx` — panel lists tracked changes + unresolved comments;
  Aceitar/Rejeitar call `acceptChange`/`rejectChange` with the right `revisionId`; Resolver calls
  `commentsHook.resolve`; empty (all-clean) state.
- **Extend `DocumentEditorPage.test.tsx`:** (a) panel mounts only when instance
  `status==='changes_requested'` (and NOT for a plain draft / under_review); (b) instance query
  enabled for `draft`; (c) re-submit disabled while tracked changes OR unresolved comments remain,
  enabled when clean; (d) on re-submit with resolved comments: `removeCommentMark` called per resolved
  comment, then flush, then `submitForReviewRequest` — asserted call order; (e) the existing
  `under_review` submit + readonly tests stay green.
- **Proof commands:** `npx vitest run src/features/documents src/features/approval` GREEN; `npx tsc
  --noEmit -p tsconfig.build.json` clean; `grep -rn "onTrackedChangesChange" src/features/documents/pages/DocumentEditorPage.tsx`
  → present (now wired).

## Deviations to surface at HS-1

- **D1 — no backend tracked-changes freeze gate.** Master plan §F6 assumed F0 wired
  `ScanForUnresolvedMarkup`; it does not exist (F0 HS-2 path A defer). Tracked-changes cleanliness
  before re-submit is **client-authoritative** (disabled submit + hash chain); the backend enforces
  **comments-only** at freeze (`ErrFreezeBlockedByUnresolvedComments`). Operator decides at HS-1
  whether a server-authoritative tracked-changes gate is worth a post-M2c backend feature.
- **D2 — naming.** Real ref API is `acceptChange`/`rejectChange` (design doc said `…ById`).
