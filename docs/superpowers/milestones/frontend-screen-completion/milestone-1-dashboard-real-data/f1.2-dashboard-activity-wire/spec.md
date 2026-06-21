# Feature F1.2 — Spec

> **Milestone:** 1 — Dashboard real data  ·  **Folder:** `f1.2-dashboard-activity-wire`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca — *mission F1.2 contract; producer confirmed at runtime.*

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | What is the real producer + shape for the "§ MURMÚRIOS" feed? | `GET /api/v1/audit/events?limit=N` → `ListAuditEventsResponse { items: AuditEventItem[], page }`. `AuditEventItem = { id, occurred_at, actor_id, action, resource_type, resource_id, payload, trace_id }`. **Correction (post-review):** these ARE in the web app's generated `lib/api-types/index.d.ts` (`components['schemas']['AuditEventItem']` line 2439, `['ListAuditEventsResponse']` line 2452, path `/audit/events` line 524, `operations['listAuditEvents']` line 3983 with `query.limit?: number`). So the client uses the **generated types + typed `api.GET`** (the `library.ts` pattern) — NOT a hand-typed shape. The earlier "not in index.d.ts" claim came from a false-negative grep and is wrong. |
| 2 | `/audit/events` is `CapAuditRead`-guarded — a non-privileged user gets 403/empty. How should the feed behave? | Render a truthful empty state on empty, and a graceful non-crashing state on error/403 — never a mock, never a thrown render. |
| 3 | How to map an event to who/what/code? | who = `actor_id`; code = `resource_id`; what = humanized `action` (known-action map + dotted-action fallback); time = relative from `occurred_at` (reuse existing `formatRelative`). |

## Consumer contract (FIRST — before any producer)

- **Consumer:** `DashboardPage.tsx` "§ MURMÚRIOS" `activityList` — currently maps `MOCK_ACTIVITY`
  (`{ who, what, code, time }`).
- **Contract:** an `ActivityItem[]` = `{ id: string; who: string; what: string; code: string; occurredAt: string }`
  (`id` = event id, used as the stable React key), newest-first, derived from the live audit events;
  the component formats `occurredAt` via the existing `formatRelative`. Empty list → existing empty
  styling; query error → honest non-crash state.
- **Producer (must match):** `GET /api/v1/audit/events?limit=8` → `ListAuditEventsResponse`.
- **Source of truth:** `audit/delivery/http/handler.go` (`ListAuditEventsResponse`, `AuditEventItem`).

## What this feature implements

Delete `MOCK_ACTIVITY` from `DashboardPage.tsx`. Add an audit client
(`features/dashboard/api/audit.ts`: `AuditEventItem`/`ListAuditEventsResponse` re-exported from the
generated `components['schemas']` + `fetchRecentAuditEvents` via typed `api.GET('/audit/events',
{ params: { query: { limit } } })`, `library.ts` error pattern), a pure
`deriveActivityItems(events): ActivityItem[]` (who/what/code map), a `useDashboardActivityQuery` hook
(`QK.audit.recent(8)`), and wire the feed with loading / empty / error states. No backend change.

## Non-goals (mandatory)

- No backend/Go change; no new audit endpoint or field.
- No audit export, filtering UI, pagination, or "view all" — just the recent feed.
- No change to the stats cards (F1.1) or any other dashboard surface.
- No reformatting of `formatRelative` — reuse as-is.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `MOCK_ACTIVITY` gone from the dashboard feature | `grep -rnE "MOCK_ACTIVITY" frontend/apps/web/src/features/dashboard` → exit 1 / 0 | real |
| `MOCK_` fully gone from dashboard (milestone bar) | `grep -rnE "MOCK_" frontend/apps/web/src/features/dashboard` → exit 1 / 0 | real |
| `deriveActivityItems` maps event→{who,what,code,occurredAt}; humanizes known + unknown actions | new unit test passes against fixtured `AuditEventItem[]` | fixture (labeled) |
| Feed renders live events at runtime; empty + error states render without crashing | preview `/` with seeded/real audit data + forced-error case | real |
| Type-safe + suite green | `npm run build` (tsc) exit 0; `make test` dashboard scope green | real |
| Both reviewer agents APPROVE (D2) | `frontend-screen-reviewer` + `frontend-code-reviewer` verdicts on record in `evidence.md` | real |

> TDD: `deriveActivityItems` test failing first → implement to green.

## ADR needed?

- [x] No durable architectural decision — skip. The audit client uses the generated contract types +
  typed `api.GET` (the canonical `library.ts` pattern, frontend-structure §3 rule 4: consumers use
  generated types, never hand-rolled shapes). No new rule, no deviation.
