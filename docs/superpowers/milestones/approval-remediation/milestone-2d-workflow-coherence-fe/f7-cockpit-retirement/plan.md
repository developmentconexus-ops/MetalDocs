# F2d.7 Cockpit Retirement — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement
> this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Physically delete the retired approval cockpit and the dead editor-route page, and retarget the
surviving `/edit` path constructors to the canonical `/documents/:id` workspace, completing the
single-artifact-destination migration (ADR 0080).

**Architecture:** FE-only, deletion-dominant. The `/approvals/:documentId` and `/documents/:id/edit`
redirects ALREADY exist (F2d.5 S3) and are KEPT — they preserve bookmarks/deep-links. This plan removes
the orphaned `ApprovalCockpitPage`, the wholly-cockpit-only `useDocumentApprovalArtifact` hook, and the
unmounted `DocumentEditorRoutePage`, then points live navigation directly at `/documents/:id` so it stops
bouncing through the `/edit` redirect. Zero backend diff.

**Tech Stack:** React 18, react-router-dom, TanStack Query, vitest + Testing Library.

**Locked constraints (from `docs/superpowers/analysis/2026-07-09-f7-cockpit-retirement-system-impact.md` §10):**
1. Zero backend diff — every change under `frontend/apps/web/src`.
2. KEEP both redirect routes; the `/approvals` redirect MUST keep forwarding `location.search` (`?decision=`).
3. Delete the WHOLE `useDocumentApprovalArtifact` hook (wholly cockpit-only) — not a branch.
4. Do NOT touch `ArtifactApprovalScreen`, `TemplateApprovalRoute`, or `getActiveSiblingCtaLabel` /
   `getActiveSiblingDestination` the *functions* (retarget the returned literal only). Do NOT touch
   `DashboardPage` (bare `/approvals` inbox links are correct).
5. Grep deletion gate is binding — no surviving `ApprovalCockpitPage` / `useDocumentApprovalArtifact`
   reference, no `/documents/${…}/edit` constructor, before close.

**Working dir for all commands:** `frontend/apps/web`.

---

### Task 1: Retarget the live `/edit` path constructors → `/documents/:id`

**Files:**
- Modify: `src/features/documents/lib/documentWorkflow.ts:28-33`
- Modify: `src/features/documents/adapters/useDocumentArtifact.ts:207`
- Modify: `src/features/documents/pages/DocumentDetailRoute.tsx:101,145`
- Modify: `src/features/documents/pages/NewDocumentWizardPage.tsx:179`
- Test: `src/features/documents/pages/NewDocumentWizardPage.test.tsx:293`

- [ ] **Step 1: Update the assertion test to the new target (RED)**

In `NewDocumentWizardPage.test.tsx` change the line at :293:

```ts
// before
expect(mockNavigate).toHaveBeenCalledWith('/documents/doc-xyz/edit');
// after
expect(mockNavigate).toHaveBeenCalledWith('/documents/doc-xyz');
```

- [ ] **Step 2: Run it to verify it fails**

Run: `npx vitest run src/features/documents/pages/NewDocumentWizardPage.test.tsx`
Expected: FAIL — the source still navigates to `/documents/doc-xyz/edit`.

- [ ] **Step 3: Retarget the four live constructors**

`documentWorkflow.ts` — collapse `getActiveSiblingDestination` (the `/documents/:id` workspace is
mode-adaptive: it renders the editor for `draft`/`rejected` by document status, so the destination no
longer branches). Drop the now-unused `state` param:

```ts
// before (lines 28-33)
export function getActiveSiblingDestination(documentId: string, state: ActiveSiblingState): string {
  if (state === 'draft' || state === 'rejected') {
    return `/documents/${documentId}/edit`;
  }
  return `/documents/${documentId}`;
}
// after
export function getActiveSiblingDestination(documentId: string): string {
  return `/documents/${documentId}`;
}
```

`useDocumentArtifact.ts:207` — update the single caller (drop the second arg; `activeSiblingState` stays
used by `getActiveSiblingCtaLabel`):

```ts
// before
? getActiveSiblingDestination(activeSiblingDocumentId, activeSiblingState)
// after
? getActiveSiblingDestination(activeSiblingDocumentId)
```

`DocumentDetailRoute.tsx:101`:

```ts
// before
const handleView = () => navigate(`/documents/${documentId}/edit`);
// after
const handleView = () => navigate(`/documents/${documentId}`);
```

`DocumentDetailRoute.tsx:145`:

```ts
// before
navigate(`/documents/${response.document.id}/edit`);
// after
navigate(`/documents/${response.document.id}`);
```

`NewDocumentWizardPage.tsx:179`:

```ts
// before
navigate(`/documents/${result.document.id}/edit`);
// after
navigate(`/documents/${result.document.id}`);
```

- [ ] **Step 4: Run the affected suites to verify GREEN**

Run: `npx vitest run src/features/documents/pages/NewDocumentWizardPage.test.tsx src/features/documents/pages/DocumentDetailRoute.test.tsx src/features/documents/adapters/useDocumentArtifact.test.tsx`
Expected: PASS. Then `npx tsc --noEmit -p .` → 0 errors (confirms the dropped `state` param has no other caller).

- [ ] **Step 5: Commit**

```bash
git add src/features/documents/lib/documentWorkflow.ts src/features/documents/adapters/useDocumentArtifact.ts src/features/documents/pages/DocumentDetailRoute.tsx src/features/documents/pages/NewDocumentWizardPage.tsx src/features/documents/pages/NewDocumentWizardPage.test.tsx
git commit -m "refactor(approval-fe): F2d.7 retarget live /edit constructors to /documents/:id"
```

---

### Task 2: Delete the dead `DocumentEditorRoutePage`

**Files:**
- Delete: `src/features/documents/pages/DocumentEditorRoutePage.tsx`
- Delete: `src/features/documents/pages/DocumentEditorRoutePage.test.tsx`

Context: `DocumentEditorRoutePage` exports a lazy-route `Component`, but it is mounted in NO router
(`documents/routes.tsx` uses `DocumentWorkspacePage` for `/documents/:id` and `RedirectEditToWorkspace`
for `/edit`). It is referenced ONLY by its own test (verified: `grep -rn DocumentEditorRoutePage src`).
Its `onDone` navigates to the now-dead `/edit` literal.

- [ ] **Step 1: Prove it is orphaned**

Run: `grep -rn "DocumentEditorRoutePage" src`
Expected: only `DocumentEditorRoutePage.tsx` (self) and `DocumentEditorRoutePage.test.tsx` (its test). No
router import.

- [ ] **Step 2: Delete both files**

```bash
git rm src/features/documents/pages/DocumentEditorRoutePage.tsx src/features/documents/pages/DocumentEditorRoutePage.test.tsx
```

- [ ] **Step 3: Verify build + suite clean**

Run: `npx tsc --noEmit -p .` → 0 errors. `npx vitest run src/features/documents` → all green (the deleted
`DocumentEditorRoutePage.test.tsx:38` `/edit` assertion is gone with the page).

- [ ] **Step 4: Commit**

```bash
git commit -m "refactor(approval-fe): F2d.7 delete dead DocumentEditorRoutePage (unmounted in router)"
```

---

### Task 3: Delete the orphaned `ApprovalCockpitPage`

**Files:**
- Delete: `src/features/approval/pages/ApprovalCockpitPage.tsx`
- Delete: `src/features/approval/pages/ApprovalCockpitPage.test.tsx`
- Modify (stale comment only): `src/features/approval/components/sidebar/ApprovalSidebar.tsx:39`

Context: `ApprovalCockpitPage` is orphaned — the only importer is its own test (verified by grep). The
`/approvals/:documentId` route is already `RedirectApprovalToWorkspace` (KEEP it). The long-documented
pre-existing failing test `ApprovalCockpitPage.test.tsx:342` (`?decision=reject`) is RESOLVED by this
deletion.

- [ ] **Step 1: Prove it is orphaned (non-test)**

Run: `grep -rn "ApprovalCockpitPage" src`
Expected: only `ApprovalCockpitPage.tsx` (self), `ApprovalCockpitPage.test.tsx` (its test), and comment
references at `ApprovalSidebar.tsx:39` + `DocumentWorkspacePage.tsx:180`. No live import, no route mount.

- [ ] **Step 2: Delete the page + its test**

```bash
git rm src/features/approval/pages/ApprovalCockpitPage.tsx src/features/approval/pages/ApprovalCockpitPage.test.tsx
```

- [ ] **Step 3: Remove the stale comment reference in ApprovalSidebar.tsx:39**

Open `ApprovalSidebar.tsx:39`; if the comment names `ApprovalCockpitPage` as a live collaborator, update
it to reference the workspace (`DocumentWorkspacePage`) or delete the now-inaccurate line. Keep the change
comment-only — do not alter sidebar logic.

- [ ] **Step 4: Verify build + suites clean**

Run: `npx tsc --noEmit -p .` → 0 errors. `npx vitest run src/features/approval` → GREEN (the
`?decision=reject` failure is gone). `npx vitest run src/features/documents` → GREEN.

- [ ] **Step 5: Commit**

```bash
git add -A src/features/approval
git commit -m "refactor(approval-fe): F2d.7 delete orphaned ApprovalCockpitPage (single-screen live, ADR 0080)"
```

---

### Task 4: Delete the wholly-cockpit-only `useDocumentApprovalArtifact` hook

**Files:**
- Delete: `src/features/documents/adapters/useDocumentApprovalArtifact.ts`
- Delete: `src/features/documents/adapters/useDocumentApprovalArtifact.test.tsx`
- Modify (stale comment only): `src/features/documents/pages/DocumentWorkspacePage.tsx:156,180`

Context: after Task 3, `useDocumentApprovalArtifact`'s only non-test consumer (`ApprovalCockpitPage`) is
gone. It is now fully dead. The LIVE detail adapter `useDocumentArtifact` imports only
`getActiveSiblingDestination` from `documentWorkflow` — NOT this hook (verified). The hook's internal
`href="/approvals"` breadcrumb (line ~211) dies with it.

- [ ] **Step 1: Prove it is fully dead (post-Task-3)**

Run: `grep -rn "useDocumentApprovalArtifact" src`
Expected: only `useDocumentApprovalArtifact.ts` (self), `useDocumentApprovalArtifact.test.tsx` (its test),
and comment references in `DocumentWorkspacePage.tsx`. NO live import.

- [ ] **Step 2: Delete the hook + its test**

```bash
git rm src/features/documents/adapters/useDocumentApprovalArtifact.ts src/features/documents/adapters/useDocumentApprovalArtifact.test.tsx
```

- [ ] **Step 3: Clean the stale comment references in DocumentWorkspacePage.tsx:156,180**

Open `DocumentWorkspacePage.tsx`; the comments at ~:156 and ~:180 reference the retired
`useDocumentApprovalArtifact` / `ApprovalCockpitPage` as if live. Update them to describe the current
`DocumentWorkspacePage` derivation, or delete the now-inaccurate lines. Comment-only — no logic change.

- [ ] **Step 4: Verify build + full documents suite clean**

Run: `npx tsc --noEmit -p .` → 0 errors. `npx vitest run src/features/documents` → GREEN.

- [ ] **Step 5: Commit**

```bash
git add -A src/features/documents
git commit -m "refactor(approval-fe): F2d.7 delete dead useDocumentApprovalArtifact cockpit adapter"
```

---

### Task 5: ADR 0080 closure note + full feature gates + independent review + evidence.md

**Files:**
- Modify: `wiki/decisions/0080-single-artifact-destination.md` (append F2d.7 closure note)
- Create: `docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/f7-cockpit-retirement/evidence.md`
- Modify: `docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/milestone.md` (F2d.7 DONE marker)

- [ ] **Step 1: Amend ADR 0080 with the physical-retirement closure**

Append a dated note: the cockpit page + `useDocumentApprovalArtifact` adapter + dead
`DocumentEditorRoutePage` were physically deleted in F2d.7; the `/approvals/:documentId` and
`/documents/:id/edit` redirects remain as the bookmark-preservation surface; all live navigation now
targets `/documents/:id` directly. No new ADR (0080 governs).

- [ ] **Step 2: Run the full feature gates from clean state**

```bash
# grep deletion gates — all must return NOTHING
grep -rn "ApprovalCockpitPage" src | grep -v "^Binary"          # expect: empty
grep -rn "useDocumentApprovalArtifact" src                        # expect: empty
grep -rEn "/documents/\$\{[^}]+\}/edit" src                       # expect: empty
# build + suites
npx tsc --noEmit -p .                                             # expect: 0 errors
npx vitest run src/features/documents                             # expect: all green
npx vitest run src/features/approval                              # expect: all green (cockpit fail gone)
# zero-backend gate (BASE = this feature's base SHA)
git diff --name-only <BASE_SHA> HEAD | grep -v "^frontend/apps/web/src" | grep -v "^docs/" | grep -v "^wiki/decisions/0080"   # expect: empty
```

- [ ] **Step 3: Independent whole-feature review**

Dispatch an independent `cavecrew-reviewer` on `git diff <BASE_SHA> HEAD -- frontend/apps/web/src`.
Fix any 🔴/🟡 with a follow-up implementer + re-review before closing. 🔵 nits at reviewer discretion.

- [ ] **Step 4: Write evidence.md**

Mirror the F2d.5b/F2d.6 evidence shape: per-task commits, the grep deletion-gate outcomes, tsc + both
suites green (note the `?decision=reject` pre-existing fail is RESOLVED by the cockpit deletion, not
carried), zero-backend proof, review disposition, ADR 0080 amendment, NOT pushed.

- [ ] **Step 5: Mark the F2d.7 milestone row DONE + commit**

```bash
git add wiki/decisions/0080-single-artifact-destination.md docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/
git commit -m "docs(approval-fe): F2d.7 close — ADR 0080 closure note + evidence.md + milestone DONE marker"
```

---

## Self-review

- **Spec coverage:** milestone F2d.7 items (a) redirect — already done, KEPT (constraint 2); (b) delete
  cockpit + reduced model — Tasks 3+4 (whole hook, not a branch — global-max); (c) worklist deep-links —
  already satisfied (no `/approvals/${id}` constructor survives), verified; (d) `/edit` constructors —
  Task 1 (4 live) + Task 2 (1 dead page deleted). ADR — Task 5 (amend 0080, no new ADR). All covered.
- **Placeholder scan:** none — every step has exact paths, before/after code, and expected output.
- **Type consistency:** `getActiveSiblingDestination` signature change (drop `state`) has exactly one
  caller (`useDocumentArtifact.ts:207`), updated in the same task; `activeSiblingState` remains defined
  for `getActiveSiblingCtaLabel`. tsc gate in Step 4 confirms no orphaned reference.
