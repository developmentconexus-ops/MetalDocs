# Feature F3.4 — Evidence

> **Milestone:** 3 — Notifications  ·  **Feature:** `f3.4-notifications-wire`  ·  **Closed:** 2026-06-22
> **Contract:** [`spec.md`](spec.md) (consumer contract + Validation Gate this proves against).
> A feature is closed only when every row below is filled with **real, honestly-labeled** output.

## What was implemented

Outcome: the empty `/notifications` page+route is gone; notifications now live as a **top-bar bell →
unread-badge → popover → spotlight** menu, backed by the real F3.1–F3.3 producers plus one additive
bulk endpoint. Producer matches the **consumer contract** in `spec.md` (consumers `NotificationBell` +
`NotificationsSpotlight` drive the generated shapes; types were not back-fitted to the producer).

**Backend leg (additive — only backend touch):**
- `POST /notifications/read-all` — OpenAPI path + `oapi-codegen` regen (`api/api.gen.go`), strict
  handler (`delivery/http/handler.go`), repository `MarkAllRead` (self-scoped
  `UPDATE … SET status='READ', read_at=now() WHERE recipient_user_id = :caller AND status IN ('PENDING','SENT')`),
  tier-1 route guard (`cmd/metaldocs-api/permissions.go`) under existing `CapNotificationRead`. Idempotent,
  returns `{updated:int}`. No new cap, no migration — reuses F3.2's table + cap.

**Frontend leg (grade-A: CSS-Module + `tokens.css`, generated snake_case types, TanStack Query, typed openapi-fetch):**
- `features/notifications/api/notifications.ts` — re-exports generated `Notification`,
  `NotificationsListResponse`, `UnreadCountResponse`, `MarkAllReadResponse`, list params from
  `operations[...]`; `fetchNotifications`/`fetchUnreadCount`/`markNotificationRead`/`markAllNotificationsRead`
  via typed `api.GET`/`api.POST` + `asApiError`. Legacy noops deleted.
- `lib/eventTypeLabel.ts`, `lib/formatNotificationTimestamp.ts` (pure, unit-tested).
- `queries/useNotificationQueries.ts` (`useNotificationsQuery`, `useUnreadCountQuery` — `staleTime 30_000`,
  badge `refetchInterval 60_000`), `queries/useNotificationMutations.ts` (mark-read + mark-all; both
  invalidate `QK.notifications.unreadCount()` + `QK.notifications.list()` on settle; toast on error).
- `components/NotificationRow.tsx`, `NotificationBell.tsx` (badge + popover, outside-click/Esc close,
  `aria-haspopup`/`aria-expanded`/labeled badge, `:focus-visible`), `NotificationsSpotlight.tsx`
  (`components/ui/Dialog` overlay, full list + "Marcar todas como lidas") — each + `.module.css` + test.
- `features/shell/components/AppToolbar.tsx` — placeholder bell replaced with `<NotificationBell/>`;
  user-name span migrated to `.userName` CSS class; `type="button"` + `--text-on-brand` token cleanups.
- `lib/queryKeys.ts` — `QK.notifications.{unreadCount,list}`; `lib/api-types/index.d.ts` — regenerated.
- **Removed:** `pages/NotificationsPage.tsx`, `routes.tsx`, `components/NotificationsPanel.tsx`,
  `components/AppShellHeader.tsx`, `useNotifications.ts`; `AppRouter.tsx` route entry; `NotificationItem`
  from `lib/types/index.ts`. Tracker updated (`wiki/implementation/screen-redesign-tracker.md`).

Commit: see `git log` (committed this session; sha recorded at close).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — backend failing-first → green | `go test -tags integration ./internal/modules/notifications/... -run TestNotifications` | `mark_all_read_flips_only_callers_unread_and_idempotent` red (assertion mismatch on un-transitioned seed row) → fixed assertion to scope `read_at` to transitioned IDs → **PASS** | real (live PG) |
| TDD — FE failing-first → green | vitest hooks/components vs `vi.mock`-ed api shaped as generated types | red → green | fixture |
| Backend build | `go build ./...` | `BUILD_EXIT=0` | real |
| API contract lint | `go run ./scripts/api-lint -strict ./api/openapi/v1/openapi.yaml .` | **`0 violation(s)`**, `APILINT_EXIT=0` | real |
| Integration — `read-all` self-scope + idempotent | `go test -tags integration ./internal/modules/notifications/...` | `TestNotifications` **7/7 PASS** (incl. mark_all_read self-scope/idempotent), `TestNotificationsFanoutWorker` 7/7 PASS, `ok …infrastructure 149.308s` | real (live PG) |
| Go unit suite | `go test ./...` | **`GOTEST_EXIT=0`** — no `FAIL` lines | real |
| FE typecheck (generated types, no mapper) | `npx tsc --noEmit -p tsconfig.build.json` | clean (no output / exit 0) | real (type) |
| Notifications FE tests | `npx vitest run src/features/notifications` | **33 passed (7 files)** | fixture |
| FE production build | `npx vite build` | **`✓ built in 11.55s`, `VITEBUILD_EXIT=0`** (chunk-size warning pre-existing/unrelated) | real |
| Static — noop deleted | grep `items: []\|never[]\|Stubbed pending` in `api/notifications.ts` | **0** | real (static) |
| Static — legacy retired | grep `NotificationsPanel\|NotificationsPage\|useNotifications\|NotificationItem` in `src` | **0** | real (static) |
| Static — route removed | grep `/notifications` in `**/routes*.tsx` | **0** | real (static) |
| Runtime — full flow | preview: seed 2 unread + 1 read → reload → bell → popover → spotlight → mark-all | badge **"2 não lidas"**; popover 3 rows (chips PUBLICADO/APROVADO/REPROVADO, times há 7 min/2 h/3 d, titles+messages); spotlight full list + mark-all enabled; **mark-all → bell "Notificações" (0 unread), button disabled, DB `READ\|3\|3` all `read_at` stamped** | real (runtime, live PG) |

> Runtime proof exercised the real provider end-to-end (seeded real rows, real `POST /notifications/read-all`,
> verified DB state + UI). `preview_screenshot` timed out this session → proof captured via
> `preview_eval`/`preview_snapshot` (accessibility tree + DOM text) + direct `psql` state assertions.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| `read-all` marks only caller's PENDING/SENT; idempotent; self-scope | yes | Integration row (7/7 PASS) |
| `api-lint -strict` = 0 | yes | API contract lint row |
| `go build` / `go test ./...` green | yes / yes | Backend build + Go unit suite rows |
| Noop deleted at root | yes | Static — noop row (0) |
| Legacy retired (no panel/page/hook/type refs) | yes | Static — legacy row (0) |
| Route removed; tracker updated | yes | Static — route row (0); tracker diff |
| Generated-type consumption, no mapper | yes | FE typecheck row |
| List/unread/mark-read/mark-all hooks | yes | Notifications FE tests (33) |
| NotificationRow / Bell / Spotlight behavior | yes | Notifications FE tests (33) |
| lib helpers pure + tested | yes | Notifications FE tests (eventTypeLabel 8, formatTimestamp 6) |
| `frontend-screen-reviewer` APPROVE | yes | Review disposition |
| `frontend-code-reviewer` APPROVE | yes | Review disposition |
| FE suite green | yes (notifications subset) / pre-existing unrelated failures | see Bounded defers |
| Build clean | yes | FE production build row |
| Runtime functional pass | yes | Runtime row |

## Review disposition

- **Spec-compliance / visual (`frontend-screen-reviewer`): APPROVE WITH NITS.** No Critical, no Major
  architecture finding. Token/primitive/a11y nits resolved at root by family:
  - *a11y focus rings* — added `:focus-visible` to `.bell`, `.showAll`, `.markAll`, and the retry buttons.
  - *unstyled retry button* — added `.retry` class (both popover + spotlight error states).
  - *inline style / bare `color:white`* (AppToolbar, touched-file migration) — `.userName` class + `--text-on-brand`.
  - *popover focus-trap* — see Bounded defers (non-modal popover; Esc + outside-click present).
- **Code-quality (`frontend-code-reviewer`): APPROVE WITH NITS.** No Critical. Majors resolved:
  - *M-ARCH-1 inline query key* — `['notifications','list']` → `QK.notifications.list()` (+ test assertion updated).
  - *M-A11Y-1 missing `type=`* — `type="button"` on the Novo-documento button.
  - *barrel imports* (`resolveErrorMessage` ×3) → `../../../lib/api`.
  - *M-TEST-1 E2E smoke* — see Bounded defers.
  - All re-verified: `npx vitest run src/features/notifications` 33/33, `tsc` clean.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Playwright E2E smoke for the bell (`e2e/flows/notification_bell.spec.ts`) | Behavior fully covered by 33 vitest tests + real runtime proof (seed→popover→spotlight→mark-all + DB assert). E2E adds CI regression guard only. | Add when the `e2e/` suite is next touched, or before the v1 release E2E pass. Owner: frontend-screen-completion mission. |
| Popover focus-trap (non-modal `div[role=dialog]`) | Spotlight (the modal surface) uses native `<dialog>` `showModal()` with real focus containment. The popover is a transient menu that closes on Esc **and** outside-click; WCAG 2.4.3 gap is low-severity for a non-modal popover. | Revisit if the popover gains interactive depth (filters/actions) or an a11y audit flags it. Owner: frontend-screen-completion mission. |
| Full `make test` FE suite has pre-existing failures (InboxPage, DocumentEditorPage, templates.create) | Unrelated to this feature — confirmed pre-existing via `git stash` baseline; matches known `node_modules` junction drift (entities, lru-cache). Notifications subset 33/33 green. | Resolved by completing the pnpm install (FE node_modules drift memo). Not in F3.4 scope. |
