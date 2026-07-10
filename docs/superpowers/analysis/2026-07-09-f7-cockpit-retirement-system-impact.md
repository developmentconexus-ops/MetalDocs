# System-impact analysis — F2d.7 cockpit retirement

**Date:** 2026-07-09
**Intent (one line):** Retire the approval cockpit now that the single mode-adaptive working screen
(`/documents/:id`) is live (F2d.5 S3, ADR 0080) — delete the orphaned `ApprovalCockpitPage` + its
wholly-cockpit-only adapter, and retarget the surviving `/edit` and `/approvals` path constructors so
worklist navigation stops bouncing through redirects.
**Work type:** feature (FE-only; deletion + reference-retargeting)
**Author:** developing-new-work skill
**Verdict:** 🟢 Green *(see §10)*

> FE-only cleanup completing an already-ratified migration. Same ten sections; module-birth rows
> marked **N/A** with reason. **Targeted-verified the full deletion/reference surface before this run**
> (cavecrew-investigator + two router reads) — findings below correct three drifts from the milestone row.

---

## 1. Classify & own

- **Work type:** feature (FE-only, deletion-dominant).
- **Owning module(s):** frontend `features/approval` (the retired `ApprovalCockpitPage` + `routes.tsx`
  redirect) and frontend `features/documents` (the cockpit-only adapter `useDocumentApprovalArtifact`,
  the `/edit` path constructors, `routes.tsx`). Also `features/dashboard` (worklist deep-links).
- **Explicitly NOT owning / must NOT touch:** the backend `documents`/`approval` modules (zero backend
  diff); `features/shared/controlled-artifact/ArtifactApprovalScreen` (SHARED — templates still render
  it via `TemplateApprovalRoute`, so it survives; only the cockpit's usage of it goes); `features/
  templates` `TemplateApprovalRoute` `screenModel` variant (a separate parallel surface, out of scope);
  `documentWorkflow.getActiveSiblingDestination` — retarget its returned path, do NOT delete it (the
  LIVE `useDocumentArtifact` adapter imports it).
- **Cross-module edges (with direction):** none new; edges are DELETED. Removing `ApprovalCockpitPage`
  drops FE→`useDocumentApprovalArtifact`; removing that hook drops its edge to the approval instance
  adapters. No Go cross-module edge touched.
- **Ambiguity?** None. AS-3 not triggered — every owner is a FE feature folder, verified by grep.

## 2. Foundation verdict

- **Base you'd build on:** the single-screen destination is already the ratified global maximum
  (ADR 0080, `docs/superpowers/specs/…single-artifact-destination`), **already live** — F2d.5 S3
  (`a19ea6c3`) mounted `/documents/:id` as the working screen and converted BOTH `/approvals/:documentId`
  and `/documents/:id/edit` to redirects (`RedirectApprovalToWorkspace` forwards `location.search`;
  `RedirectEditToWorkspace`). The cockpit is the **retired legacy surface** left dangling.
- **Sound, or legacy/patch/workaround?** Deleting the cockpit IS the global-maximum move — it removes
  the legacy, it does not optimize inside it. Not AS-2. Verified: `ApprovalCockpitPage` is already
  **orphaned** — the only importer is its own test file; it is mounted in no route. So this feature is
  the *physical* completion of the *logical* retirement F2d.5 already performed.
- **Milestone-row drifts corrected (targeted-verified):**
  1. Milestone item (a) "`/approvals/:documentId` **becomes** a redirect" — **already done** in F2d.5
     S3 (`approval/routes.tsx:7,20`, search-forwarding confirmed). F2d.7 keeps that redirect and
     deletes the orphaned page it replaced.
  2. Milestone item (b) "delete `useDocumentApprovalArtifact`'s cockpit-only reduced **model**" — the
     hook is **wholly** cockpit-only (sole non-test consumer = `ApprovalCockpitPage`, verified by grep).
     Global-max: delete the **entire hook + its test suite**, not a branch — leaving a half-hook is dead
     code, a defect.
  3. Beyond the milestone's cited constructor list: a **5th** `/edit` constructor at
     `DocumentEditorRoutePage.tsx:16`, and that page is **referenced only by its own test** — mounted in
     no router (dead). Captured for the plan: delete the dead page + test rather than retarget a dead path.
  4. Milestone item (c) "worklist deep links target `/documents/:id`" is **already satisfied** — a grep
     for `/approvals/${…}` per-document constructors returns **none** (the only one, `InboxPage.
     openDecisionFlow`, was retargeted in F2d.5). `DashboardPage.tsx:65,66,106` navigate to **bare
     `/approvals`** — the inbox LIST route (which survives), not a per-document cockpit URL — so they are
     correct as-is and are NOT retargeted.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | No | No authz code touched; deletion of FE surfaces only. The surviving `/documents/:id` screen's gating is unchanged. | — |
| Contract-first (OpenAPI + oapi-codegen) | No | No route/DTO/spec change. Redirects are FE router entries, not API routes. | — |
| Multi-tenant pooled | No | No query/table touched. | — |
| Async = transactional outbox | No | No side effect touched. | — |
| DB enforces invariants | No | No DB touched. | — |
| Cross-module via published interface only | No | FE-only; edges are removed, none added. | — |

No violation. AS-1 not triggered.

## 4. Capability wiring

**N/A** — no capability added, changed, or removed. `TestCapabilityRegistrySize` unchanged.

## 5. Module wiring

**N/A** — no module born or retired (a FE *page* retires, not a bounded context).

## 6. Frameworks to reuse, not reinvent

Reuse the existing react-router `<Navigate replace>` redirect primitive already in place (no new redirect
mechanism); the existing `/documents/:id` workspace screen (already the live destination); the existing
`getActiveSiblingDestination` helper (retarget its literal, keep the helper). No new hook, no new route
machinery, no hand-rolled navigation. Nothing reinvented — this feature only **removes** and **retargets**.

## 7. Contract & data

- **OpenAPI-first:** no route added/changed.
- **Migration:** none.
- **Destructive change?** Yes — **code deletion** (a page + a hook + their test suites). This is safe:
  the deletion targets are verified-orphaned (cockpit) or wholly-cockpit-only (hook). Grep gates in the
  plan enforce zero surviving references before the delete lands.

## 8. Test & QA plan

- **Canonical framework:** FE = vitest + Testing Library. Backend = zero change.
- **QA gates that apply:** the FE component/route-behavior gate + **grep deletion gates** (no surviving
  `ApprovalCockpitPage` / `useDocumentApprovalArtifact` reference; no `/documents/${…}/edit` or
  `/approvals/${…}` path *constructor* outside the redirect routes themselves). Contract / multi-tenant /
  async / DB gates = N/A. Build-clean-after-deletion gate applies (tsc 0 + suites green).
- **Evidence shape:** route-redirect tests (params + `?decision=` query preserved — already covered at
  `documents/routes.test.tsx:85`, keep green); worklist-link tests target `/documents/:id`; grep gates;
  `tsc --noEmit 0`; documents + approval suites green (the cockpit's `?decision=reject` failing test dies
  WITH the page — the long-documented pre-existing fail is *resolved by deletion*, not left behind);
  zero-backend `git diff --name-only` gate; independent cavecrew-review; evidence.md.

## 9. Docs / ADR

- **Wiki:** light — refresh `wiki/architecture/frontend-structure.md` route-map stamp if it names the
  cockpit; remove any cockpit reference.
- **ADR:** **no new ADR.** ADR 0080 (single-artifact-destination) already governs this; F2d.7 is its
  physical completion. **Amend 0080** with a "cockpit physically retired in F2d.7" closure note.
- **REQ IDs cited:** design brief single-screen destination; ADR 0080.

## 10. Verdict & locked constraints

- **Verdict:** 🟢 Green — proceed to plan. FE-only deletion + retargeting completing an already-live,
  already-ratified migration; no invariant, no capability, no backend/contract/migration change.
- **Open hard-stops:** none (AS-1/AS-2/AS-3 all clear; the orphan status of the cockpit and the
  wholly-cockpit-only status of the hook were targeted-verified by grep before this verdict).
- **Locked constraints handed to design:**
  1. **Zero backend diff** — `git diff --name-only` gate; all changes under `frontend/apps/web/src`.
  2. **Keep the two redirect routes** (`approvals/:documentId`, `documents/:documentId/edit`) — they are
     the bookmark/deep-link preservation surface; F2d.7 deletes the orphaned *page*, never the redirect.
     The `?decision=` search-forwarding on the `/approvals` redirect MUST remain.
  3. **Delete the WHOLE dead hook** — `useDocumentApprovalArtifact.ts` + its test are wholly cockpit-only;
     remove entirely, do not leave a reduced branch. Prove zero surviving references (grep gate) first.
  4. **Do NOT touch shared/live survivors** — `ArtifactApprovalScreen` (templates use it),
     `TemplateApprovalRoute` `screenModel`, and `getActiveSiblingDestination` the *function* (retarget its
     returned literal only). Templates domain is out of scope.
  5. **Retarget every LIVE `/edit` path constructor** to `/documents/${id}` (the 4 live ones:
     `documentWorkflow.ts:30`, `DocumentDetailRoute.tsx:101,145`, `NewDocumentWizardPage.tsx:179`) so
     navigation lands directly, not via a redirect bounce. `getActiveSiblingDestination` collapses to
     returning `/documents/${id}` for all states (the `/documents/:id` workspace is mode-adaptive — it
     renders the editor for draft/rejected by document status, so the caller no longer branches). Resolve
     the dead `DocumentEditorRoutePage.tsx` (+ its `/edit` constructor at :16, + test) by **deletion**
     (referenced only by its own test) rather than retargeting a dead path. **Do NOT touch `DashboardPage`
     — its bare `/approvals` inbox links are correct (item c already satisfied, §2).**
  6. **Grep deletion gate is binding** — the delete does not land until no `ApprovalCockpitPage` /
     `useDocumentApprovalArtifact` reference and no stray `/edit` or `/approvals` constructor survives
     outside the redirect route files.
