# Feature F3.4 — Plan (notification menu: bell → popover → spotlight)

> Implements [`spec.md`](spec.md) (Approved 2026-06-22). Order: **backend `read-all` (contract-first, TDD)
> → FE api/queries/lib → FE components (row → bell → spotlight) → AppToolbar wire → remove page/route →
> retire legacy → reviewers → evidence.** TDD throughout (failing test first).

## Files touched

**Backend (additive `POST /notifications/read-all`):**
- `api/openapi/v1/openapi.yaml` — add the `read-all` path (mirror `/notifications/{id}/read`) + a tiny
  `MarkAllReadResponse {updated:int}` schema.
- `internal/modules/notifications/api/api.gen.go` — regen (`go generate`).
- `internal/modules/notifications/infrastructure/notifications_repository.go` — add `MarkAllRead(ctx, tenantID, recipientUserID) (int, error)`.
- `internal/modules/notifications/delivery/http/handler.go` — add `MarkAllNotificationsRead` handler.
- `apps/api/cmd/metaldocs-api/permissions.go` — add the tier-1 `CapNotificationRead` guard for `POST /api/v1/notifications/read-all` (exact path, **before** the `/read` suffix + bare-collection rules so it matches first).
- `internal/modules/notifications/infrastructure/notifications_repository_integration_test.go` — add `TestMarkAllRead` subtests.

**Frontend:**
- `frontend/apps/web/src/lib/api-types/index.d.ts` — regen (`npm run gen:api`).
- `frontend/apps/web/src/lib/queryKeys.ts` — add `QK.notifications.list(params)`.
- `frontend/apps/web/src/features/notifications/api/notifications.ts` — rewrite (typed `api.GET`/`api.POST`; delete noops).
- `frontend/apps/web/src/features/notifications/queries/` — `useNotificationsQuery.ts`, `useUnreadCountQuery.ts`, `useMarkNotificationReadMutation.ts`, `useMarkAllNotificationsReadMutation.ts`.
- `frontend/apps/web/src/features/notifications/lib/` — `formatNotificationTimestamp.ts`, `eventTypeLabel.ts` + `__tests__/`.
- `frontend/apps/web/src/features/notifications/components/` — `NotificationRow.tsx`+`.module.css`+test, `NotificationBell.tsx`+`.module.css`+test, `NotificationsSpotlight.tsx`+`.module.css`+test.
- `frontend/apps/web/src/features/shell/components/AppToolbar.tsx` — replace placeholder bell with `<NotificationBell/>`.
- **Remove:** `frontend/apps/web/src/features/notifications/pages/NotificationsPage.tsx`, `frontend/apps/web/src/features/notifications/routes.tsx` (+ the spread/import in the app router), `frontend/apps/web/src/components/NotificationsPanel.tsx`. Remove `NotificationItem` from `lib/types/index.ts` iff unused (check `AppShellHeader.tsx`).
- **Tracker:** `wiki/implementation/screen-redesign-tracker.md` (or M0 evidence) — record `/notifications` route removal.

## Plan (task list — each starts with a failing test)

### Phase A — Backend `read-all` (Grade-A, contract-first)
1. **OpenAPI:** add `/notifications/read-all` POST (operationId `markAllNotificationsRead`, tags `[notifications]`, 200 `MarkAllReadResponse {updated:int}` + standard error refs). `go generate ./internal/modules/notifications/api` → `api.gen.go` gains the op.
2. **Failing integration test** (`TestMarkAllRead`): seed user A with 2 PENDING + 1 READ, user B with 1 PENDING; assert `MarkAllRead(A)` returns `updated==2`, A's rows all READ with `read_at` set, B's row untouched (self-scope), re-run returns `updated==0` (idempotent). Red (no repo method).
3. **Repo `MarkAllRead`:** `UPDATE metaldocs.notifications SET status='READ', read_at=now() WHERE tenant_id=$1 AND recipient_user_id=$2 AND status IN ('PENDING','SENT')` → `RowsAffected`. Green.
4. **Handler `MarkAllNotificationsRead`:** `extractTenantAndUser` → `repo.MarkAllRead` → `200 {updated}`. Wire into the strict server (generated interface now requires it).
5. **Permissions:** add the exact-path guard before the suffix rule. `go build ./...` + `go vet` + cilint guards + `api-lint -strict`=0 + `go test ./...`.

### Phase B — FE data layer
6. **`npm run gen:api`** → FE types gain `markAllNotificationsRead` + `MarkAllReadResponse`. `QK.notifications.list` added.
7. **`api/notifications.ts` rewrite:** `fetchNotifications(params)`, `fetchUnreadCount()`, `markNotificationRead(id)`, `markAllNotificationsRead()` — typed `api.GET`/`api.POST`, `asApiError`. Delete `{items:[]}`/`never[]` noops. `npm run typecheck`.
8. **`lib/` helpers + failing unit tests** → implement → green.
9. **`queries/` hooks + failing hook tests** (`renderHook` + `QueryClient` + `vi.mock('../api/notifications')`): each hook test red → implement → green. Mutations assert POST called + `invalidateQueries` on list + unread-count; toast via `resolveErrorMessage` (mirror `useRouteAdminMutations`).

### Phase C — FE components (TDD: test → component)
10. **`NotificationRow`** — test renders title/message/chip/timestamp + state; unread → "Marcar como lida" click fires `onMarkRead`. Implement (CSS-Module + tokens). Green.
11. **`NotificationsSpotlight`** — test: full list renders; "Marcar todas como lidas" fires mark-all; loading/error/empty guards. Implement with `components/ui/Dialog` (`open`/`title`/`onClose`/`footer`). Green.
12. **`NotificationBell`** — test: badge shows unread count (hidden at 0/loading/error); click opens popover w/ rows (limit 8); "Mostrar Todas" opens spotlight; Esc/outside-click closes; a11y (`aria-haspopup`/`aria-expanded`, badge label). Implement (CSS-Module popover + `Icon name="bell"`). Green.

### Phase D — Wire + cleanup
13. **`AppToolbar`:** replace placeholder bell button with `<NotificationBell/>`. (AppShell test mocks AppToolbar — unaffected.)
14. **Remove page+route:** delete `NotificationsPage.tsx` + `routes.tsx`; remove `notificationsRoutes` import/spread from the app router. `grep` `/notifications` route refs = 0.
15. **Retire legacy:** delete `components/NotificationsPanel.tsx`; remove `NotificationItem` iff unused (inspect `AppShellHeader.tsx` — if dead, note; else leave + record).
16. **Tracker update:** record `/notifications` route removal in the redesign tracker / M0 evidence.

### Phase E — Gates
17. `make test` green; `npm run build` clean; `npm run typecheck` clean.
18. Backend: `go build ./...`, `go test ./...`, `api-lint -strict`=0, 6 CI guards green, `.\scripts\check-system-runnable.ps1`.
19. Runtime: preview workflow — bell badge = real unread; popover lists; per-row mark-read flips + badge decrements; spotlight full list; mark-all → badge 0.
20. **Reviewers:** dispatch `frontend-screen-reviewer` + `frontend-code-reviewer`; both APPROVE on record (design-source N/A — vs tokens/primitives + operator UX). Fix-by-family on findings, re-review.
21. **`evidence.md`** — commands + real output, TDD proof, reviewer verdicts, runtime proof, bounded defers.

## Test strategy
- **Backend:** integration (live PG via `testdb.Open` + `testdb.New{Tenant,User,Notification}` factories) — the `read-all` self-scope + idempotency proof is real, not fixture.
- **FE hooks:** vitest `renderHook` + `QueryClient(retry:false)` + `vi.mock` the api module (fixture responses shaped as generated types).
- **FE components:** vitest + `@testing-library/react`, factory data, DOM + `fireEvent` interaction assertions (mirror `ApprovalTimelinePanel.test.tsx`).
- **Runtime:** preview workflow for the live-data functional pass.

## Ordering rationale
Backend first (FE `read-all` call needs the generated type). Data layer before components (components consume hooks). Row before bell/spotlight (both compose it). Wire + removals last (after the menu works) so the app never has a broken intermediate notifications surface. Reviewers + evidence final.

## Risks / hard-stops
- **HS-2:** `read-all` must be **additive** (new path/handler/repo method) — no change to F3.1/F3.2/F3.3 behavior, no migration, no new cap, no publish/approval touch. Reuse table + `CapNotificationRead`.
- **HS-3:** contract-first regen order (OpenAPI → oapi-codegen → gen:api); FE consumes generated types only.
- **HS-6:** no search/filter/preferences/channels/SSE — bell+popover+spotlight + read-all only.
- Permission-rule ordering: the exact `/read-all` rule must precede the `/read` suffix + bare-collection rules.
- Route removal must also drop the `notificationsRoutes` spread or the build breaks (dangling import).
