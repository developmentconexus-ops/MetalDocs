# F5.3 — Signoff Reconcile · Consumer Contract Spec

> Feature: `f5.3-signoff-reconcile` · Milestone 5 · added 2026-07-10 as HS-4 remediation of the
> milestone-validator FAIL (F5.1's deliverable was retired by ADR 0080 and never reconciled).
> **Type:** verify-only reconciliation — **no rebuild, no behavior change.**

## Why this feature exists

F5.1 ("Detalhe Signoff") closed 2026-06-23 against a **standalone approval cockpit**
(`SignoffDetailPage.tsx` mounting `ControlledDocumentDetailPanel`, reusing `ApprovalTimelinePanel`
+ `SignoffDialog`, at route `/approvals/:documentId`). **ADR 0080** ("single artifact destination",
commit `0c96dfb2`, 2026-07-07) — from the parallel `approval-remediation` M2d program — retired the
cockpit pattern and **relocated** the sign-off decision surface into the mode-adaptive document
workspace. F2d.7 then deleted the cockpit files. F5.1's `evidence.md` still proves against those
deleted files, so the milestone-validator FAILed (C2 can't re-run deleted tests; C4/C5 APPROVE is
for a deleted screen; C6 docs⊥code split-brain).

The **function is live** (signoff moved, not removed) and **0 backend changed** — so the remedy is a
verify-only reconcile, not a rebuild (validator recommendation; orchestrator authorized 2026-07-10).

## The current surface this feature contracts against (runtime truth)

| Concern | Current owner (file:line) |
|---|---|
| Route `/approvals/:documentId` | `features/approval/routes.tsx:7-22` → `RedirectApprovalToWorkspace` → `/documents/:id` (preserves `?decision=`) |
| Decision screen | `features/documents/pages/DocumentWorkspacePage.tsx` — mode-adaptive; *approving* mode via `deriveWorkspaceMode` (`:68`) |
| Decision model | `buildDocumentSignoffDecision` (`DocumentWorkspacePage.tsx:206-213`); submit closure `:193-200` |
| Decision mount | `WorkspaceSidebar` → `DecisionFooter` → `ApprovalModeFooter` → shared `ArtifactDecisionPanel` (password re-auth + legal-effect) |
| Timeline | `features/approval/components/sidebar/ApprovalTimeline.tsx` (fork of the deleted `ApprovalTimelinePanel`), mounted `WorkspaceSidebar.tsx:120-125` |
| Sign-off network call | `features/approval/hooks/useSignoffMutation.ts` → `approvalApi.signoff` → `POST /documents/{id}/signoff`, body `{decision,reason,password,content_hash}`, If-Match `"v{rev}"`, idempotency `resourceId=documentId`; invalidates approval/documents/controlledDocuments |
| Error classification | `features/approval/lib/signoffErrors.ts` (`mapSignoffError`; 412→`stale`) — lifted verbatim from the retired `SignoffDialog` |
| Inbox navigation | `InboxPage.tsx:97-114` → `/documents/{id}?decision=` (direct; redirect only catches legacy links) |
| PDF/A4 official view | `PdfCanvas` for `approved\|scheduled\|published`; in-approval read stays on the source `DocumentShell` canvas (ADR 0080 amendment F2d.5b) |

## Acceptance (objectively checkable)

1. **No stale pointers:** F5.1 `evidence.md` carries a superseded-by-ADR-0080 banner pointing here
   (history preserved, not deleted); `milestone.md` objective + status reflect the workspace surface.
2. **Current surface re-proven by runnable tests** (re-runnable from clean state):
   `DocumentWorkspacePage.test.tsx` (approving-mode panel renders / SoD-gates / `?decision=` preselect),
   `useSignoffMutation.test.tsx` (**new** — asserts `POST /signoff` body `content_hash` + If-Match
   `"v{rev}"` + 412→stale, re-establishing the guard the deleted `SignoffDialog.test` held),
   `DecisionFooter.test.tsx`, `WorkspaceSidebar.test.tsx`, `signoffErrors.test.ts`, `InboxPage.test.tsx`.
3. **No dead trace:** orphan `SignoffDetailPage.module.css` removed; no remaining import/ref to the
   deleted cockpit files.
4. `tsc --noEmit` clean; **zero behavior change** (only a test + docs added, one orphan CSS removed —
   no product-source `.tsx` behavior touched); 0 backend changed.
5. Both `frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE the **current** surface + this
   reconciliation.

**Out of scope (→ chip, not inline):** any real defect in the current workspace signoff surface
(coverage gaps beyond the added test, `ArtifactDecisionPanel` a11y, etc.), and any behavior change.
