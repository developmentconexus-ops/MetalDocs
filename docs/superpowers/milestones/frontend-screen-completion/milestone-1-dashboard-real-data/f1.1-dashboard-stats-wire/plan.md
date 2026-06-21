# Feature F1.1 — Plan

> Input: `spec.md` (approved 2026-06-21). Output of `writing-plans`-equivalent. The "how".

## Approach

Frontend-only wire. Put the `by_status` → 3-card derivation in a **pure, testable function**, cover it
with a unit test (TDD), add a query hook mirroring the existing inbox hook, then swap `DashboardPage`
from `MOCK_STATS` to the live data. No backend touch.

## Files touched

| File | Change |
|------|--------|
| `features/dashboard/lib/deriveDashboardStats.ts` | **new** — pure `deriveDashboardStats(stats: DocumentStatsResponse): StatItem[]`; the status→label map + missing-key→0 logic. Exports `StatItem` type. |
| `features/dashboard/lib/__tests__/deriveDashboardStats.test.ts` | **new** — TDD unit test (written first, failing): fixtured `DocumentStatsResponse` → asserts the 3 items (approved; under_review+rejected; published) and missing-key→0. |
| `features/dashboard/queries/useDashboardStatsQuery.ts` | **new** — `useQuery({ queryKey: QK.documents.stats(), queryFn: fetchLibraryStats, staleTime: 30_000 })`. Reuses existing `fetchLibraryStats`. |
| `features/dashboard/pages/DashboardPage.tsx` | **edit** — delete `MOCK_STATS`; call `useDashboardStatsQuery`; derive items; render in hero pills + "§ SEU PULSO" rows with `…` (loading) / `—` (error/empty). |

## Test strategy

- TDD unit: `deriveDashboardStats` — the only logic with branches. Failing test first.
  - case: full `by_status` → correct 3 values + labels/subs.
  - case: missing keys → 0 (no fabricated value, no NaN).
  - case: `under_review`+`rejected` summed into "Em revisão".
- Build/type: `npm run build` (tsc) — consumer uses generated `DocumentStatsResponse`.
- Runtime: preview `/` renders numeric pills (real-provider proof) — captured in `evidence.md`.

## Ordering

1. Write failing `deriveDashboardStats.test.ts`.
2. Implement `deriveDashboardStats.ts` → green.
3. Add `useDashboardStatsQuery.ts`.
4. Edit `DashboardPage.tsx` (remove `MOCK_STATS`, wire live).
5. `npm run build` + targeted `make test`; runtime preview check.
6. Dispatch reviewer agents (D2); record evidence.
