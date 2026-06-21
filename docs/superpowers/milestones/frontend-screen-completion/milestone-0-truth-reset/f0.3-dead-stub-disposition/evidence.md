# Feature F0.3 — Evidence

> **Milestone:** 0  ·  **Feature:** `f0.3-dead-stub-disposition`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (consumer = the router; delete the dead Operations/Audit surface at the root).

## What was implemented

Deleted the dead, no-API Operations/Audit surface at the root (D7 — IAM Admin Center already owns
metrics/audit/sessions). Not a redirect, not a flag — files removed and routes unregistered.

- **Deleted (6 files + 2 emptied dirs):**
  `features/operations/pages/OperationsPage.tsx`, `features/operations/routes.tsx`,
  `features/audit/pages/AuditPage.tsx`, `features/audit/routes.tsx`,
  `components/OperationsCenter.tsx`, `components/OperationsCenter.module.css`, and the now-empty
  `features/operations/` + `features/audit/` directories.
- **`src/app/AppRouter.tsx`:** removed the `operationsRoutes` + `auditRoutes` imports and their two
  spread entries from the protected `AppShell` children. All other route arrays and the
  `{ path: '*', element: <Navigate to="/" replace/> }` catch-all are untouched.
- **`src/app/AppRouter.test.tsx`:** dropped the `operationsRoutes` import (the file it pointed at is
  gone) and re-expressed the F0.2 single-root-index invariant against the sole surviving carrier
  (`dashboardRoutes`). The F0.2 contract (`/` → one root index = Dashboard) is preserved, not weakened.
- **Shared primitives kept:** `OperationsCenter`'s helper imports (`WorkspaceDataState`, `TimelineRail`,
  `WorkspaceHeroHeader`, `metalNobreProcessAreaHint`) have other live consumers (TemplatesListPage,
  ControlledDocumentsExplorer, NotificationsPanel, DocumentsHubHeader) and were **not** touched —
  deleting one would trip HS-2.
- Not yet committed (M0 commits at milestone close / operator discretion).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| 6 source files gone | `[ -e <path> ]` test of each | **6/6 gone** | real |
| Zero `OperationsCenter` refs | `grep -rn "OperationsCenter" src` | **exit 1, 0 matches** | real |
| Zero `OperationsPage`/`AuditPage` refs | `grep -rEn "OperationsPage\|AuditPage" src` | **exit 1, 0 matches** | real |
| Route registrations removed | `grep -nE "operationsRoutes\|auditRoutes" src/app/AppRouter.tsx` | **exit 1, 0 matches** | real |
| Compiler red→green (dangling imports cleaned) | `npx tsc --noEmit -p tsconfig.json` | **TSC_EXIT:0** | real |
| F0.2 guard preserved | `npx vitest run src/app/AppRouter.test.tsx` | **2 passed** (no `operationsRoutes` import) | real |
| FE suite — no new failures | `npx vitest run` | **36 failed / 405 passed / 5 skipped** — identical to baseline; deletion added **0** new failures | real |
| Runtime: dead routes gone, no shell | preview (web 4173 + API 8081), authenticated | `/operations` → redirects to `/` (Dashboard); `/audit` → redirects to `/` (Dashboard) — catch-all `Navigate to="/"`; no empty shell rendered | real |
| Runtime: surviving routes intact | navigate `/` and `/documents` after hard reload | `/` → Dashboard ("Bom dia, Administrator"); `/documents` → Library ("Documentos / Biblioteca") — router mounts cleanly post-deletion | real |

> **TDD (compiler-as-test):** files deleted first → `tsc` would red-flag every dangling reference →
> references cleaned in `AppRouter.tsx` + `AppRouter.test.tsx` → `tsc` exit 0, vitest green. The F0.2
> unit guard is the durable regression assertion and stayed green.
>
> **Console note:** `preview_console_logs` shows a frozen buffer of 12 stale HMR errors
> (`operationsRoutes is not defined` / `auditRoutes is not defined`) carrying the **transient
> edit-window timestamps** (`t=...251453`, `t=...254939`) — emitted in the moment between removing the
> imports and removing the spreads. They did not recur after the final edit or a hard reload: the app
> renders Dashboard and Library end-to-end, which is impossible if the live `AppRouter` module still
> threw. `tsc` exit 0 is the authoritative static proof; the full render is the runtime proof.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Page/component source files gone | yes | 6/6 `[ -e ]` gone |
| Zero `OperationsCenter` references | yes | grep exit 1 / 0 |
| Zero `OperationsPage`/`AuditPage` references | yes | grep exit 1 / 0 |
| Route registrations removed | yes | grep on AppRouter.tsx exit 1 / 0 |
| F0.2 invariant preserved | yes | `AppRouter.test.tsx` green, no operations import; runtime `/`→Dashboard |
| Build/typecheck clean | yes | `tsc --noEmit` exit 0 |
| FE suite — no new failures | yes | 36 failed unchanged vs baseline; 0 new |

All 7 criteria **met**.

## Review disposition

- Spec-compliance review: self-review against `spec.md` — PASS. Exact deletion set; root-cause removal
  (files gone, routes unregistered) not a redirect/flag (validation §4); shared primitives preserved
  (HS-2 respected); F0.2 contract carried forward.
- Code-quality review: deletion + 2 surgical edits, no logic added. Independent judgement deferred to
  the M0 `milestone-validator` (separation of powers) — it re-runs the grep set + build from clean state.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Frozen stale HMR error buffer in preview | Tooling artifact from the mid-edit window, not live app state; `tsc`=0 + full render disprove a real error | Trigger: clears on next dev-server restart. Owner: n/a |
| `authRoutes` / `AuthRoutePage` unmounted dead export still present | Out of F0.3 scope — D7 limits deletion to Operations/Audit | Recorded in tracker as `not-started / out-of-scope`; sweep at program close if still unused. Owner: operator |
