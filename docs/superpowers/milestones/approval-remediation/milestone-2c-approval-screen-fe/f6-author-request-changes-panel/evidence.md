# F6 — Evidence

## Commands + real output

- `npx vitest run src/features/documents src/features/approval` → **Test Files 51 passed (51) ·
  Tests 351 passed (351)** (implementer + reviewer). After the review-driven 404 test was added:
  `DocumentEditorPage.test.tsx` **30 passed (30)** (was 29 + 1 new graceful-404 case).
- `npx tsc --noEmit -p tsconfig.build.json` → **clean, zero output** (implementer + reviewer).
- `grep -n "onTrackedChangesChange" src/features/documents/pages/DocumentEditorPage.tsx` → present
  (`:488`, wired to `setTrackedChanges`).
- `git diff --stat -- src/features/documents/components/DocumentShell.tsx` → **EMPTY** — shared shell
  untouched (main-session + reviewer verified). Panel is page-owned.
- Task-8 grep: `ErrFreezeBlockedByUnresolvedComments` has **no entry** in
  `internal/modules/documents/approval/http/errors.go`'s `errors.Is` mapping (only
  `freeze.effective_date_missing` is freeze-related). No distinguishable `err.code` reaches the FE
  submit catch → generic toast retained (client gate is primary). No backend change.
- No new `interface/type .*Response` in the diff; no backend/Go files touched; no
  `approvalApi.ts`/`approvalTypes.ts`/`api-types` changes (reviewer confirmed full-diff scope).

## TDD proof

- Implementer confirmed RED-first: `RequestedChangesPanel.test.tsx` failed on missing import
  (`Failed to resolve import "../RequestedChangesPanel"`); the 5 extended DocumentEditorPage cases
  failed `5 failed | 24 passed (29)` before the wiring existed. GREEN after implementing. Tests first.

## Runtime proof (observable change) + fixture-vs-real

- **Detection gate bug fixed:** `approvalInstanceQuery.enabled` broadened to `docStatus ===
  'under_review' || docStatus === 'draft'` (`:390`) so the author page can fetch the instance after
  the backend reverts the doc to `draft` on `request_changes`. `changesRequested =
  approvalInstanceQuery.data?.status === 'changes_requested'` (`:393`). Panel gated at `:585`.
- **Graceful no-instance (the key correctness risk):** a never-submitted draft has no instance →
  `getApprovalInstance` 404s. `retry:false` + no `.isError`/`.error` render path → fails closed:
  no panel, no error surface, editor stays editable. **Now runtime-proven** by the review-driven test
  "a plain draft whose instance lookup 404s shows no panel and no error surface" (rejects the query,
  asserts editor `data-mode==='document-edit'`, no "Mudanças solicitadas", no permission/error alert).
- **Panel gating:** renders ONLY for instance `changes_requested` — asserted across draft-in_progress
  (no panel), under_review-in_progress (no panel), draft-changes_requested (panel). Labeled
  **fixture/mock** (vitest).
- **Per-change accept/reject:** Aceitar/Rejeitar → `editorRef.current?.acceptChange(revisionId)` /
  `rejectChange(revisionId)` (real ref names, not `…ById`). Per-comment Resolver →
  `commentsHook.resolve(c)`.
- **Re-submit gate:** `requestedChangesBlocksSubmit = changesRequested && (hasUnresolvedTrackedChanges
  || hasUnresolvedComments)` folded into the submit button `disabled`. Test proves disable→enable
  transition off the live `onTrackedChangesChange`. **Regression-safe:** `changesRequested` is false
  for a normal draft, so a first submit is unaffected (verified — pre-existing autosave-wiring submit
  tests stay green).
- **Clean-buffer re-submit ORDER (real correctness point):** for each RESOLVED comment
  `removeCommentMark(id)` (`:252`) → `saveNow` (`:258`) → `flush` (`:263`) → `submitForReviewRequest`
  (`:275`). Marks stripped BEFORE flush so the persisted buffer is clean; flush BEFORE submit. Asserted
  via a shared `callOrder` recorder (`removeIdx < flushIdx < submitIdx`).
- **Buffer-dirty subtlety:** `removeCommentMark` mutates the buffer silently (no onChange). The impl
  forces the flush path via a `resolvedAnyMark` flag (`if (editorDirty || resolvedAnyMark)`) so the
  stripped marks persist even when the author made no manual edit. Reviewer confirmed the flush + submit
  fire for a changes-requested re-submit with only resolved comments and no `onChange`.

## Key design decisions (verified against runtime truth)

- **No backend tracked-changes markup gate (D1).** The master plan §F6 assumed F0 wired
  `ScanForUnresolvedMarkup`; it does NOT exist — F0 scoped it out as HS-2 path A (F0 evidence.md:21,
  47-49). Tracked-changes cleanliness before re-submit is enforced **client-side** (disabled submit +
  frozen-content-hash chain); the backend enforces **comments-only** at freeze
  (`ErrFreezeBlockedByUnresolvedComments`). Specced to runtime truth, not the stale plan.
- **Panel page-owned, shared shell untouched.** `RequestedChangesPanel` mounts in the author frame
  beside `ArtifactMetaSidebar`; `DocumentShell` (shared with the cockpit) is byte-identical.
- **removeCommentMark scoped to resolved comments** — the `if (comment.resolved)` guard. At submit
  time the re-submit gate guarantees zero unresolved comments, so a mixed set never reaches the loop;
  the guard is defensive. (See "not-a-real-path" note below.)

## Review / QA disposition

- Independent reviewer subagent (separate from implementer, own tools, no edits): **APPROVE**,
  **0 Critical, 0 Major, 2 Minor**. All 12 adversarial checks PASS. Reviewer re-ran vitest (351/351)
  + tsc (clean) from clean state. Reviewer specifically confirmed the normal-first-submit is not
  regressed by the gate.
- **Minor findings + disposition:**
  1. **Missing runtime test for the 404 no-instance draft path** (implementation correct by inspection,
     but fixture-only). **CLOSED** — added the graceful-404 test (main-session, `DocumentEditorPage.test.tsx`
     30/30 green). This was the higher-value gap and a real reachable path.
  2. **Missing negative test for `removeCommentMark` on a mixed resolved/unresolved set.** **NOT
     ADDED — deliberately.** The path is **unreachable**: the re-submit gate blocks submit while ANY
     comment is unresolved, so `removeCommentMark`'s loop can never see an unresolved comment at submit
     time. Writing that test would require bypassing the production gate to exercise an impossible
     state — fabricated coverage. The `if (comment.resolved)` guard remains as defense-in-depth.

## Deviations to surface at HS-1

- **D1 — no backend tracked-changes freeze gate** (F0 HS-2 path A defer). Client-authoritative
  enforcement; comments-only backend gate. Operator decides at HS-1 whether a server-authoritative
  tracked-changes gate is worth a post-M2c backend feature.
- **D2 — naming.** Real ref API is `acceptChange`/`rejectChange` (design doc said `…ById`).

## Bounded defers

- None new. The server-authoritative tracked-changes freeze gate remains the program-level HS-2
  bounded defer (D1); the client disabled-submit gate + frozen-content-hash chain + comment-freeze
  gate cover today.
