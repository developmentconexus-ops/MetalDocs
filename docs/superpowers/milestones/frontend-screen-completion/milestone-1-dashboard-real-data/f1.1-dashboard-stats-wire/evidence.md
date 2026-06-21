# Feature F1.1 — Evidence

> **Milestone:** 1 — Dashboard real data  ·  **Feature:** `f1.1-dashboard-stats-wire`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (operator-re-scoped to honest `/documents/stats` `by_status` counts).
> Every row below is real, honestly-labeled output.

## What was implemented

- Deleted `MOCK_STATS` from `DashboardPage.tsx`; the hero stat pills and the `§ SEU PULSO`
  sidebar now render live values derived from `/documents/stats`.
- New pure `deriveDashboardStats(stats): StatItem[]` (`lib/deriveDashboardStats.ts`) — the
  operator-locked re-scope map: **Aprovados** = `by_status.approved`; **Em revisão** =
  `by_status.under_review + by_status.rejected`; **Publicados** = `by_status.published`; any
  missing key counts as `0`. Labels/subs/colors are static scaffolding; only the numbers are live.
- New `useDashboardStatsQuery` hook (`queries/useDashboardStatsQuery.ts`) — `useQuery` keyed on the
  existing `QK.documents.stats()`, `queryFn: fetchLibraryStats` (the canonical typed client in
  `features/documents/api/library.ts`), `staleTime: 30_000`.
- `statValue(i)` honesty guard in `DashboardPage.tsx`: `'…'` while loading, `'—'` on error / no
  data, the live count otherwise — **never a fabricated number**.
- Producer matches the consumer contract: consumer reads `DocumentStatsResponse.by_status` (generated
  type); producer is the live `/documents/stats` endpoint already shipped. No producer was built or
  changed (operator decision: no approval-throughput endpoint).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | `npx vitest run src/features/dashboard/lib/__tests__/deriveDashboardStats.test.ts` | derive fn absent → red; implemented → **3 passed** | real (pure-logic) |
| Static (types) | `npx tsc --noEmit -p tsconfig.build.json` | **exit 0** | — |
| Targeted test | `npx vitest run src/features/dashboard` | **7 passed (2 files)** | real (pure-logic) |
| Mock removed | `grep -rE "MOCK_" frontend/apps/web/src/features/dashboard` | **0 matches** (exit 1) | real |
| Runtime proof | Dashboard `/` on live stack (web :4173 + api :8081), `preview_snapshot` | Hero pills + `§ SEU PULSO` render live `APROVADOS 0`, `EM REVISÃO 0`, `PUBLICADOS 0` from `/documents/stats` 200 — honest `0` (seed DB has none in those statuses), not `'—'` error, not fabricated | **real (live API)** |

> Runtime `0` values are the truthful live counts for the current seed DB, not an error state — the
> `'—'` error path is reserved for query failure. Fixture-labeled rows are pure-logic unit tests;
> the runtime row is the real-provider proof.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| `MOCK_STATS` gone from dashboard feature | yes | Mock-removed row (0 matches) |
| Cards map `/documents/stats` `by_status` per re-scope (approved / under_review+rejected / published; missing→0) | yes | `deriveDashboardStats` unit tests (3) + runtime row |
| Loading → `'…'`, error → `'—'`, never fabricated | yes | `statValue` guard; TDD + runtime |
| Live values render at runtime | yes | Runtime proof row (snapshot) |
| Type-safe + suite green | yes | tsc exit 0; 7 passed |
| Both reviewer agents APPROVE (D2) | yes | Review disposition below |

## Review disposition

- **Spec-compliance / visual (`frontend-screen-reviewer`): APPROVE.** 0 Critical, 0 Major. Confirmed
  visual parity preserved, `deriveDashboardStats` `?? 0` matches the contract, `statValue` never
  fabricates. Minors (both out of M1 scope, deferred): stale `"5 dias úteis"` label;
  `activityMuted` class naming.
- **Code-quality (`frontend-code-reviewer`): APPROVE.** 0 Critical, 0 Major in M1 scope. Cross-check
  worksheet PASS on every gate (QK reuse, staleTime, honest loading/error, no `MOCK_`). Minors are
  pre-existing/adjacent (deferred below).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Stale `"5 dias úteis"` sub-label on `§ SEU PULSO` | Pre-existing text; M1 re-scope made it semantically empty (no time window). Cosmetic, not data-wiring. | Trigger: next dashboard copy/polish pass — remove or replace with a window-free label. |
| `who` = raw `actor_id` (e.g. `admin`) | Honest per spec Q3 contract; readability nicety only | Trigger: when a user-display-name lookup exists, map `actor_id`→name. |
| `lib/types/index.ts:191` legacy hand-rolled `AuditEventItem` (0 importers) | IAM-domain dead code, outside M1 mock→live boundary; not imported by dashboard or IAM (both use generated `components['schemas']`) | Trigger: IAM screen-completion milestone, or a lib/types cleanup PR — delete the dead interface. |
