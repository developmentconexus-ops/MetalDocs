# Feature F1.2 — Evidence

> **Milestone:** 1 — Dashboard real data  ·  **Feature:** `f1.2-dashboard-activity-wire`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (live `/audit/events` → `§ MURMÚRIOS` feed; generated types).
> Every row below is real, honestly-labeled output.

## What was implemented

- Deleted `MOCK_ACTIVITY` from `DashboardPage.tsx`; the `§ MURMÚRIOS` feed now renders live audit
  events from `GET /audit/events?limit=8`.
- New audit client (`api/audit.ts`) — **generated contract types**:
  `AuditEventItem = components['schemas']['AuditEventItem']`,
  `ListAuditEventsResponse = components['schemas']['ListAuditEventsResponse']`; `fetchRecentAuditEvents`
  uses the typed `api.GET('/audit/events', { params: { query: { limit } } })` with the `library.ts`
  `asApiError` pattern. **No hand-rolled shape** (the earlier hand-typed version, written on a
  false-negative-grep premise, was reworked — see spec.md Q1 correction).
- New pure `deriveActivityItems(events): ActivityItem[]` (`lib/deriveActivity.ts`): `who = actor_id`,
  `code = resource_id`, `occurredAt = occurred_at`, `what = humanizeAction(action)` (document-domain
  `ACTION_LABELS` map + de-dotted fallback so the line is **never blank**). `ActivityItem.id = e.id`
  threads the event id as the **stable React key**.
- New `useDashboardActivityQuery` hook — `useQuery` keyed `QK.audit.recent(8)`, `staleTime: 30_000`.
- Feed has loading / empty / error states; the error node carries `role="alert"`; muted text uses the
  new `.activityMuted` CSS-module class (no inline theming added by M1).
- Producer matches the consumer contract: consumer reads only `.items` (`AuditEventItem[]`) from the
  generated `ListAuditEventsResponse`; producer is the live `/audit/events` endpoint. No backend change.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | `npx vitest run src/features/dashboard/lib/__tests__/deriveActivity.test.ts` | derive fn absent → red; implemented → **4 passed** | real (pure-logic) |
| Static (types) | `npx tsc --noEmit -p tsconfig.build.json` | **exit 0** (generated audit types resolve) | — |
| Targeted test | `npx vitest run src/features/dashboard` | **7 passed (2 files)** | real (pure-logic) |
| Mock removed | `grep -rE "MOCK_" frontend/apps/web/src/features/dashboard` | **0 matches** (exit 1) | real |
| Generated types, not hand-rolled | code-reviewer worksheet + `audit.ts:8-9` import from `lib/api-types` | PASS — no hand-rolled audit interface in M1 files | real |
| Runtime proof — live feed | Dashboard `/` on live stack, `preview_snapshot` | `§ MURMÚRIOS` renders **8 live audit events** (`admin` · `authz bypass system admin` · `route.manage` · `há 6 dias`; `… auth session revoked …`; `FAM-WAVEF família deactivated`; real trace/resource ids), `/audit/events` 200 | **real (live API)** |
| Runtime proof — unknown-action fallback | same snapshot | open-set actions (`authz.bypass.system_admin`, `iam.user.role.upserted`, `auth.session.revoked`) render de-dotted/readable, never blank | **real (live API)** |
| Runtime proof — empty/error states | code path + state branches (loading/empty/error each rendered conditionally; error `role="alert"`) | non-crashing honest states; seeded data exercised the populated path at runtime | real (populated path live); empty/error branches code-verified |

> The live feed is real-provider proof. The empty and error branches were not force-triggered against
> the live API (data was present); they are code-verified conditional branches matching the
> `useDashboardInboxQuery` precedent — labeled honestly, not asserted as exercised.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| `MOCK_ACTIVITY` / all `MOCK_` gone from dashboard | yes | Mock-removed row (0 matches) |
| `deriveActivityItems` maps event→{id,who,what,code,occurredAt}; humanizes known + unknown | yes | 4 unit tests + runtime fallback row |
| Feed renders live events at runtime; states non-crashing | yes | Runtime live-feed row + state-branch verification |
| Audit client uses generated types + typed `api.GET` | yes | Generated-types row; spec.md Q1 correction |
| Type-safe + suite green | yes | tsc exit 0; 7 passed |
| Both reviewer agents APPROVE (D2) | yes | Review disposition below |

## Review disposition

- **Visual (`frontend-screen-reviewer`): APPROVE.** 0 Critical, 0 Major. Confirmed the humanizer
  fallback de-dots open-set actions with no blank lines, `role="alert"` present, generated types used.
- **Code-quality (`frontend-code-reviewer`): APPROVE.** 0 Critical, 0 Major in M1 scope. Worksheet
  PASS: generated types (`audit.ts:8-9`), `api.GET('/audit/events', …)`, `QK.audit.recent`,
  `staleTime 30_000`, stable `key={a.id}`, loading/empty/error trifecta. Minor §1 (legacy
  `lib/types` `AuditEventItem`) and three pre-existing `PendingRow`/`dashboard.ts` nits deferred.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `lib/types/index.ts:191` legacy hand-rolled `AuditEventItem` (0 importers) | IAM-domain dead code outside M1 boundary; nothing imports it | Trigger: IAM milestone / lib/types cleanup PR — delete. |
| `features/dashboard/api/dashboard.ts` `fetchDashboardInbox` uses raw `apiFetch` | Pre-existing; routes through shared error layer; type-hygiene only | Trigger: convert to `api.GET('/approval/inbox', …)` in a dashboard-client hardening pass. |
| `PendingRow` nested `<button>` in `role="button"` div + missing Space key | Pre-existing a11y; not touched by mock→live wiring | Trigger: dashboard a11y pass — restructure to `<Link>`/single interactive element, add Space. |
| `PendingRow` inline theming styles | Pre-existing `§3 rule 8` violation; not introduced by M1 | Trigger: same a11y/CSS pass — move to module classes. |
| `activityMuted` class name doubles as layout wrapper | Cosmetic naming; no visual defect | Trigger: rename to `.activityBody` in a polish pass. |
| Empty/error feed branches not force-triggered against live API | Live seed had data; branches are code-verified, matching inbox precedent | Trigger: add a forced-error/empty harness when E2E fixtures land (M-later). |
