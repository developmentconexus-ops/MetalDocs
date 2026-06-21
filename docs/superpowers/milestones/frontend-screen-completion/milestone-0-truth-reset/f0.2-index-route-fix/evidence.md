# Feature F0.2 — Evidence

> **Milestone:** 0  ·  **Feature:** `f0.2-index-route-fix`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (consumer = the app router; one root-level `index:true` → Dashboard).

## What was implemented

- Removed the `{ index: true, … }` entry from `frontend/apps/web/src/features/operations/routes.tsx`.
  The file now declares only `{ path: "operations", … }`, so the **sole** root-level `index:true`
  route is Dashboard's (`dashboard/routes.tsx:5`). `/` resolves unambiguously to `DashboardPage`.
- Added durable regression guard `frontend/apps/web/src/app/AppRouter.test.tsx` (2 tests): asserts
  exactly one root-level index route across the index-carrying arrays, and that the one index belongs
  to `dashboardRoutes`, not `operationsRoutes`.
  - Asserted on the **source route-config arrays**, not the instantiated `router` — instantiating
    `createBrowserRouter` starts a real navigation (and crashes under node v26's undici `AbortSignal`),
    a side effect a unit test must not trigger. The arrays are the only root-index contributors.
- **Producer matches consumer contract:** the router (`AppRouter.tsx`, which spreads
  `dashboardRoutes` + `operationsRoutes` into the protected `AppShell` children) now receives a single
  root index. Verified at runtime: authenticated `/` renders the Dashboard, not the Operations shell.
- Scope held to F0.2: the `path:"operations"` route, `OperationsPage`, `AuditPage`, and
  `OperationsCenter` are **left in place** — their deletion is F0.3 (spec non-goal).
- Not yet committed (M0 commits at milestone close / operator discretion).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first | wrote `AppRouter.test.tsx` while `operations/routes.tsx` still had `index:true` → `toHaveLength(1)` failed (2 root indexes) | RED confirmed before edit | real |
| TDD — green after edit | removed operations' `index:true`, re-ran targeted test | `npx vitest run src/app/AppRouter.test.tsx` → **2 passed** (3 ms) | real |
| `operations/routes.tsx` drops `index:true` | `grep -n "index: true" src/features/operations/routes.tsx` | **exit 1, 0 matches** | real |
| Build/typecheck clean | `npx tsc --noEmit -p tsconfig.json` | **TSC_EXIT:0** | real |
| FE suite — no new failures | `npx vitest run` | **36 failed / 405 passed / 5 skipped** — identical failure count to baseline-36; the 2 new tests pass; 0 new failures | real |
| Runtime: `/` → Dashboard not Operations | preview (web 4173 + API 8081), authenticated, navigate `/` | `preview_snapshot` shows Dashboard workspace (Início hero "Bom dia, Administrator", AGUARDANDO VOCÊ stats, AGUARDANDO SUA ASSINATURA inbox, SEU PULSO) — `OperationsPage` shell absent; `location.pathname === "/"` | real |
| Runtime console clean | `preview_console_logs level=error` | **No console logs** | real |

> The 36 suite failures are the pre-existing, baseline-accepted set (InboxPage / DocumentEditorPage /
> templates.create — masked before M0 because the suite couldn't collect; see
> [[fe-node-modules-junction-drift]] and the mission baseline-accept). Gate is **no-new-failures**,
> which holds: count unchanged at 36, new test file green.
>
> Note: `preview_screenshot` itself times out (renderer/canvas hang under node v26) — `preview_snapshot`
> is the workflow-preferred text proof for content/structure and is sufficient here. The login *form*
> submit did not wire the mutation in-preview; login was driven by a direct `POST /api/v1/auth/login`
> (200, `system_admin`) to set the session cookie, then `/` was loaded. The form-submit gap is a
> Login-screen concern, out of F0.2 scope — recorded as a bounded defer.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Exactly one root-level `index:true`, and it is Dashboard | yes | `AppRouter.test.tsx` both assertions green; runtime `/`→Dashboard |
| TDD red→green | yes | RED (2 indexes) before edit → GREEN (2 passed) after |
| Build clean | yes | `tsc --noEmit` exit 0 |
| FE suite green (no regression) | yes | 36 failed unchanged vs baseline; new tests pass; 0 new failures |
| `operations/routes.tsx` no longer declares `index:true` | yes | `grep` exit 1 / 0 matches |

All 5 criteria **met**.

## Review disposition

- Spec-compliance review: self-review against `spec.md` — PASS. Index ambiguity resolved; non-goals
  respected (no deletion of OperationsPage/AuditPage/OperationsCenter, no router restructure, Dashboard
  page untouched).
- Code-quality review: the test asserts on route-config (no side-effect router build); rationale
  documented in the test header. Independent judgement deferred to the M0 `milestone-validator`
  (separation of powers) — it re-runs the gate from clean state.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `path:"operations"` route + `OperationsPage`/`AuditPage`/`OperationsCenter` still mounted | Explicit F0.2 non-goal; deletion is F0.3 (`dead-stub-disposition`) | Trigger: F0.3 executes the delete + grep `OperationsCenter`=0. Owner: M0 |
| Login form submit does not fire the login mutation in-preview | Login-screen behavior, not the index-route contract; did not block F0.2 runtime proof (cookie set via direct API call) | Trigger: verify/repair when the Login screen is taken up (tracker row). Owner: operator / later milestone |
| `preview_screenshot` hangs under node v26 | Tooling/runtime env issue, not app code; snapshot is the preferred text proof | Trigger: revisit after the durable `pnpm install` env fix ([[fe-node-modules-junction-drift]]). Owner: operator |
