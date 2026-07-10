# Feature F2d.7 — Evidence (approval cockpit retirement)

**Intent:** Physically retire the approval cockpit now that the single mode-adaptive working screen
(`/documents/:id`) is live (F2d.5 S3, ADR 0080) — delete the orphaned page + its dead adapter + the dead
editor-route page, cascade-clean the cockpit-only sidebar stack, and retarget the surviving `/edit` path
constructors so navigation lands directly instead of bouncing through the redirect.

**System-impact gate:** 🟢 Green — `docs/superpowers/analysis/2026-07-09-f7-cockpit-retirement-system-impact.md`
(+ a same-day correction commit: Dashboard's bare `/approvals` is the inbox LIST, kept; milestone item
(c) already satisfied). FE-only; no invariant / capability / backend / contract / migration change.

**Base SHA:** `a22d2289` (plan). **Gate correction of the milestone row:** items (a) redirects were
**already built** in F2d.5 S3 — F2d.7 KEEPS them and deletes the orphaned page they replaced.

---

## Commits (base `a22d2289` → HEAD `645d7199`)
1. `8852c122` — **Task 1** retarget the 4 live `/edit` constructors → `/documents/:id`
   (`getActiveSiblingDestination` collapsed + `state` param dropped + sole caller `useDocumentArtifact`
   updated; `DocumentDetailRoute` handleView + create-revision; `NewDocumentWizardPage` onSuccess; test
   assertion updated). RED→GREEN shown; 31/31 affected tests; tsc 0.
2. `a90275e7` — **Task 2** delete dead `DocumentEditorRoutePage` (+test) — unmounted in the router,
   referenced only by its own test (grep-proven).
3. `4c3fdb37` — **Task 3** delete orphaned `ApprovalCockpitPage` (+test); ApprovalSidebar comment
   updated. The long-documented pre-existing `?decision=reject` cockpit test failure is **resolved by
   this deletion** (the file is gone). approval suite 150/150.
4. `cc4f873a` — **Task 4** delete wholly-cockpit-only `useDocumentApprovalArtifact` (+test); stale
   cockpit comments in `DocumentWorkspacePage` cleaned. 632 deletions.
5. `e0b3caaf` — **Task 4b** sweep 7 stale comment references to the deleted symbols across 6 files so
   grep for the retired symbols is clean (satisfies the binding deletion gate — comments are references).
6. `61485000` — **review correction** delete dead `ApprovalSidebar` (+test) — the cockpit's sidebar,
   superseded by the live `WorkspaceSidebar`; plus fix two comments the sweep pointed at the wrong
   symbols (`mapApprovalChain` callers → `useDocumentArtifact`+`WorkspaceSidebar`; sign-off model →
   `DocumentWorkspacePage`).
7. `645d7199` — **cascade cleanup** delete the sidebar helpers orphaned by ApprovalSidebar's removal
   (`StageContextHeader`, `SuggestionList`, `ApprovalSidebar.module.css`); reword the one stale
   `dates.ts` comment (kept `formatDueRelative` — live in the inbox surfaces).

---

## Scope reconciliation vs the milestone row (targeted-verified)
- **(a) redirects** — already built in F2d.5 S3 (`RedirectApprovalToWorkspace` forwards `location.search`;
  `RedirectEditToWorkspace`). **KEPT** (bookmark/deep-link preservation); F2d.7 never rebuilds or deletes
  them (they are absent from this feature's diff, confirmed by the whole-feature reviewer).
- **(b) delete cockpit + reduced model** — `useDocumentApprovalArtifact` was **wholly** cockpit-only, so
  the **entire hook** was deleted (global-max: not a half-hook branch).
- **(c) worklist deep-links → /documents/:id** — **already satisfied**; no `/approvals/${id}` per-document
  constructor survives (the only one, `InboxPage.openDecisionFlow`, was retargeted in F2d.5). Dashboard's
  bare `/approvals` (inbox list) is correct and untouched.
- **(d) `/edit` constructors** — 4 live constructors retargeted (Task 1) + a 5th dead one deleted with its
  page (`DocumentEditorRoutePage`, Task 2).
- **Cascade beyond the row:** deleting the cockpit orphaned `ApprovalSidebar` (superseded by
  `WorkspaceSidebar`) and its helpers `StageContextHeader` / `SuggestionList` — all deleted so the
  retirement leaves **zero dead cockpit code** (surfaced by the whole-feature review, commits 6–7).

## Bounded defer — `CancelInstanceDialog` (NOT deleted)
The whole-feature review flagged `CancelInstanceDialog` as also-orphaned (its sole caller was the deleted
cockpit). It is **deliberately kept**: `WorkspaceSidebar.tsx:40-43` documents that the single-screen
"Cancelar instância" lifecycle action is **deferred to S2b** — the dialog is *not-yet-wired planned work*,
not superseded dead code (unlike `ApprovalSidebar`, which had a live replacement). Deleting it would
destroy planned work. Flagged for the deferred cancel-affordance wiring; the component + its test remain.

## Feature gates (Task 5, self-verified from clean state)
- **Grep deletion gates — all EMPTY ✓:** `ApprovalCockpitPage`, `useDocumentApprovalArtifact`,
  `DocumentEditorRoutePage`, `ApprovalSidebar`, `StageContextHeader`, `SuggestionList`, and
  `/documents/${…}/edit` — zero surviving references (code AND comments).
- **tsc:** `npx tsc --noEmit -p .` → **0 errors**.
- **documents + approval suites:** `vitest run src/features/documents src/features/approval` →
  **399/399 PASS** (58 files). The cockpit's `?decision=reject` failure is gone (page deleted), so the
  long-standing pre-existing fail is **resolved, not carried**.
- **Zero-backend gate:** `git diff --name-only a22d2289 HEAD` = all under `frontend/apps/web/src`
  (+ `docs/` + `wiki/decisions/0080`). No backend / OpenAPI / contract / migration / non-FE file.
- **Redirects preserved:** `documents/routes.test.tsx` redirect tests (`/documents/:id/edit` →
  `/documents/:id`; `/approvals/:id?decision=approve` → `/documents/:id?decision=approve`) remain GREEN.

## Independent review (cavecrew-reviewer, whole-feature `a22d2289..HEAD`)
- Task 1 diff review: **0 findings** (clean).
- Whole-feature review: **0🔴 5🟡**. Disposition:
  - dead `ApprovalSidebar` + `CancelInstanceDialog` → ApprovalSidebar **deleted** (superseded);
    CancelInstanceDialog **kept as bounded defer** (planned S2b wiring) — see above.
  - 3 comment-accuracy nits (`approvalWorkflow.ts` mapApprovalChain callers; `templateApprovalDecision.ts`
    sign-off model location; `ApprovalSidebar` comment) → **fixed** (commits 6–7; the ApprovalSidebar one
    resolved by deleting the file). Real callers re-verified by grep before rewording.
- Post-correction gates re-run green (399/399, tsc 0, all deletion greps empty).

---
## F2d.7 — FEATURE COMPLETE
The approval cockpit is physically retired: page, adapter, dead editor-route page, and the cockpit-only
sidebar stack are deleted with **zero dead code left**; the `/approvals/:id` + `/documents/:id/edit`
redirects are kept for bookmarks (search-forwarding intact); all live navigation targets `/documents/:id`
directly. ADR 0080 amended with a closure note (no new ADR). `CancelInstanceDialog` kept as a bounded
defer for the deferred single-screen cancel action. Zero backend. **NOT pushed.**

Next: F2d.8 (UI-driven live QA close — the milestone's last feature).
