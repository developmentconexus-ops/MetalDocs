# Feature F0.2 — Spec

> **Milestone:** 0 — Truth reset & structural cleanup  ·  **Folder:** `f0.2-index-route-fix`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca — *contract explicit in the consumer (router); no open decision.*

> This is the feature's **contract**, written and approved **before any code**. The milestone-validator
> judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | none needed — why | The consumer contract is fully explicit in the existing code: `dashboard/routes.tsx:5` and `operations/routes.tsx:5` both declare `index: true` at the root protected level (verified router read). The intended home is Dashboard (mission D7: Operations is dead, to be deleted). No ambiguity to interview. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** the app router (`src/app/AppRouter.tsx`, which spreads `dashboardRoutes` +
  `operationsRoutes` into the protected `AppShell` children) and any user navigating to `/`.
- **Contract:** at the **root protected layout level** there must be **exactly one** `index: true`
  route, and it must resolve to the Dashboard page (`dashboard/pages/DashboardPage` → its lazy
  `Component`). React-Router treats two sibling index routes as a conflict; today both Dashboard and
  Operations declare one, so the home is ambiguous / Operations shadows.
- **Source of truth for the contract:** the router files read this session; mission D5/D7 (Dashboard is
  the home, Operations is dead and slated for deletion in F0.3).

## What this feature implements

Remove the `index: true` route entry from `operations/routes.tsx` so the **only** root-level index
route is Dashboard's, making `/` resolve unambiguously to `DashboardPage`. Add a durable unit test
asserting the invariant (regression guard so a future change can't reintroduce a second root index).

> Scope boundary vs F0.3: F0.2 resolves the **index ambiguity** only. Deleting the `OperationsPage`
> file, the residual `path: "operations"` route, `AuditPage`, and `OperationsCenter` is **F0.3**.

## Non-goals (mandatory)

- **Not** deleting `OperationsPage` / `AuditPage` / `OperationsCenter` or the `path:"operations"`
  route — that is F0.3.
- **Not** changing the Dashboard page itself (its mock data is M1).
- **Not** restructuring the router or any other route array.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Exactly one root-level `index:true` route, and it is Dashboard | new test `AppRouter index route` in `src/app/AppRouter.test.tsx` (or `dashboard/routes.test.tsx`) — asserts root-level index count === 1 and the index route's lazy resolves to `DashboardPage`'s `Component` | real |
| TDD red→green | the test fails before the edit (operations index present) and passes after | real |
| Build clean | `npm run build` (or `tsc`/vite build) exits 0 | real |
| FE suite green | `make test` (vitest) — no regression | real |
| `operations/routes.tsx` no longer declares `index:true` | `grep -n "index: true" features/operations/routes.tsx` → 0 | real |

> TDD: failing test first (asserts single root index), then remove operations' index entry to green.

## ADR needed?

- [x] No durable decision — skip. (Router bug fix; the design decision "Dashboard is home, Operations dies" is mission D7.)
