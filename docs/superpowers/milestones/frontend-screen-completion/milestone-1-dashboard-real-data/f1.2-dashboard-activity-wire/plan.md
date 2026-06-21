# Feature F1.2 — Plan

> Input: `spec.md` (approved 2026-06-21). The "how".

## Files touched

| File | Change |
|------|--------|
| `features/dashboard/api/audit.ts` | **new** — local `AuditEventItem` / `ListAuditEventsResponse` types (mirror backend `auditapi`), `fetchRecentAuditEvents(limit=8)` via `apiFetch`. |
| `features/dashboard/lib/deriveActivity.ts` | **new** — pure `deriveActivityItems(events): ActivityItem[]`; action→PT humanizer (known map + dotted fallback); exports `ActivityItem`. |
| `features/dashboard/lib/__tests__/deriveActivity.test.ts` | **new** — TDD (failing first): known action, unknown action fallback, who/code/occurredAt passthrough, empty input. |
| `features/dashboard/queries/useDashboardActivityQuery.ts` | **new** — `useQuery({ queryKey: QK.audit.recent(8), queryFn: () => fetchRecentAuditEvents(8), staleTime: 30_000 })`. |
| `features/dashboard/pages/DashboardPage.tsx` | **edit** — delete `MOCK_ACTIVITY`; wire feed; loading / empty / error (403-graceful) states; reuse `formatRelative`. |

## Test strategy

- TDD unit: `deriveActivityItems` (the branching logic). Failing first.
- Build/type: `npm run build` (tsc).
- Runtime: preview `/` — feed renders live events + non-crashing empty/error — captured in `evidence.md`.

## Ordering

1. Failing `deriveActivity.test.ts`.
2. Implement `deriveActivity.ts` → green.
3. `api/audit.ts` + `useDashboardActivityQuery.ts`.
4. Edit `DashboardPage.tsx` (remove `MOCK_ACTIVITY`, wire live + states).
5. `npm run build` + targeted `make test`; runtime preview check.
6. Reviewer agents (D2); evidence.
