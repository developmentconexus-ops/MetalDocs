# Feature F0.3 — Spec

> **Milestone:** 0 — Truth reset & structural cleanup  ·  **Folder:** `f0.3-dead-stub-disposition`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca — *deletion set + scope locked by mission D7 and milestone.md F0.3; blast-radius verified this session; no open decision.*

> This is the feature's **contract**, written and approved **before any code**. The milestone-validator
> judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | none needed — why | The deletion set is fixed by mission D7 (IAM Admin Center already owns metrics/audit/sessions; Operations + Audit are redundant dead shells) and enumerated in `milestone.md` F0.3. Blast radius was verified by grep this session (refs confined to the two feature folders + `OperationsCenter`; helpers are shared and stay). No ambiguity to interview. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** the app router (`src/app/AppRouter.tsx`, which spreads `operationsRoutes` +
  `auditRoutes` into the protected `AppShell` children); the F0.2 regression guard
  (`src/app/AppRouter.test.tsx`, which imports `operationsRoutes`); future maintainers reading
  `src/` for live code.
- **Contract:** after F0.3, the router mounts **no** empty no-API shell. The `operations` and `audit`
  routes, the `OperationsPage`/`AuditPage` page modules, and the orphaned `OperationsCenter` component
  are **gone from source**. The router still compiles and every surviving route renders. The F0.2
  invariant (`/` → exactly one root index = Dashboard) still holds and is re-expressed without the
  deleted `operationsRoutes` import.
- **Source of truth for the contract:** mission D7 + `milestone.md` F0.3 acceptance cell; the
  verified-this-session grep that `OperationsCenter`/`OperationsPage`/`AuditPage` have **no other**
  consumers, and that `OperationsCenter`'s helper imports (`WorkspaceDataState`, `TimelineRail`,
  `WorkspaceHeroHeader`, `metalNobreProcessAreaHint`) are **shared primitives with other live
  consumers** (TemplatesListPage, ControlledDocumentsExplorer, NotificationsPanel, DocumentsHubHeader)
  — so they are **not** deleted.

## What this feature implements

Delete the dead Operations/Audit surface at the root (not a redirect, not a flag — D7 / validation §4
root-cause rule):

1. Delete files:
   - `src/features/operations/pages/OperationsPage.tsx`
   - `src/features/operations/routes.tsx`
   - `src/features/audit/pages/AuditPage.tsx`
   - `src/features/audit/routes.tsx`
   - `src/components/OperationsCenter.tsx`
   - `src/components/OperationsCenter.module.css`
   - the now-empty `src/features/operations/` and `src/features/audit/` directories.
2. Edit `src/app/AppRouter.tsx`: remove the `operationsRoutes` + `auditRoutes` imports and their two
   spread entries from the protected children.
3. Edit `src/app/AppRouter.test.tsx` (the F0.2 guard): drop the `operationsRoutes` import and re-express
   the single-root-index invariant against the surviving carrier (`dashboardRoutes` declares exactly
   one root `index:true`). The F0.2 contract is preserved, not weakened.

## Non-goals (mandatory)

- **Not** deleting the shared helpers `WorkspaceDataState` / `TimelineRail` / `WorkspaceHeroHeader` /
  `metalNobreProcessAreaHint` — they have other live consumers (deleting one would trip HS-2).
- **Not** rebuilding Operations/Audit function anywhere — IAM Admin Center already owns it (D7).
- **Not** touching any other route array, the Dashboard page, or any token/primitive.
- **Not** repurposing the `operations`/`audit` route paths.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Page/component source files gone | `ls` of the 6 deleted paths → all "No such file" | real |
| Zero `OperationsCenter` references remain | `grep -rn "OperationsCenter" frontend/apps/web/src` → **0** | real |
| Zero `OperationsPage`/`AuditPage` references remain | `grep -rEn "OperationsPage|AuditPage" frontend/apps/web/src` → **0** | real |
| Route registrations removed | `grep -nE "operationsRoutes|auditRoutes" src/app/AppRouter.tsx` → **0** | real |
| F0.2 invariant preserved | `AppRouter.test.tsx` still asserts a single root index = Dashboard, no `operationsRoutes` import; targeted vitest run green | real |
| Build/typecheck clean | `npx tsc --noEmit -p tsconfig.json` exits 0 (no orphan import) | real |
| FE suite — no new failures | `npx vitest run` — failure count ≤ baseline-36 (no regression from deletion) | real |

> TDD note: this is a deletion feature. The "test" is the compiler + the F0.2 guard — removing a mounted
> module that another module imports fails `tsc`/vitest until every reference is cleaned. The red→green
> is: delete files → `tsc` red (dangling imports in AppRouter/AppRouter.test) → clean the references →
> `tsc` + vitest green. The F0.2 guard is the durable regression assertion.

## ADR needed?

- [x] No durable decision — skip. The decision "Operations/Audit are dead, IAM owns their function,
  delete them" is mission **D7** (already operator-locked). F0.3 executes it; it records no new decision.
