# Feature F3.4 — Spec (notification menu: bell → popover → spotlight)

> **Milestone:** 3 — Notifications (full-stack; surface + document-lifecycle emitters)  ·  **Folder:** `f3.4-notifications-wire`
> **Status:** **Approved (pre-code) — 2026-06-22 / leandrotca.** Operator re-shaped page→menu and
> directed mark-all-read be **designed properly** (additive `POST /notifications/read-all`, no workaround).
> `plan.md` + TDD execute this session.
> **Evolution:** v1 (rewire legacy panel) → rejected (anti-pattern). v2 (grade-A full *page*) → **re-shaped by
> the operator (2026-06-22)** from a page to a **top-bar notification menu**: a bell + unread badge → click
> opens a popover of recent notifications → "Mostrar Todas" opens a **spotlight** (Dialog overlay) with the
> full list + mark-all-read. The standalone `/notifications` page+route is **removed**. One small **additive
> backend endpoint** (`POST /notifications/read-all`) backs grade-A mark-all-read.
> **Governing decisions:** F3.1 contract (generated types) + `wiki/architecture/frontend-structure.md`
> (query/API + §3 "consume generated types, never hand-roll") + `wiki/quality/screen-definition-of-done.md` (D2)
> + the M0 route-tracker (route removal recorded there).

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (`plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Grade-A pattern (binding standard, from the shipped screens)

- **CSS-Module components + `tokens.css` vars only.** No legacy `NotificationsPanel`/`WorkspaceViewFrame`/
  `WorkspaceDataState`/`catalog-*` (pre-redesign, de-coupled in refactor `8f4eed41`).
- **Consume generated `components['schemas']['X']` directly** (Dashboard `audit.ts` precedent); snake_case
  end-to-end; **no snake→camel mapper**, no hand-rolled domain wrapper (`frontend-structure §3 rule 4`).
- **Decomposition:** `api/` + `queries/` + `lib/` (tested pure helpers) + `components/` (sub-components) + `__tests__/`.
- **Typed `api.GET`/`api.POST`** (openapi-fetch) with `asApiError`; explicit loading/error/empty guards; honest states.
- **Existing primitives reused, not redesigned (HS-2):** `components/ui/Dialog` for the spotlight, `components/ui/Icon` for the bell.

## Interview record (fail-closed gate)

| # | Question | Answer (operator-decided / recon-grounded) |
|---|----------|--------------------------|
| 1 | Surface shape? | **Top-bar menu, not a page (operator 2026-06-22).** Bell + unread badge in `AppToolbar` → popover (recent ~8) → "Mostrar Todas" → spotlight (Dialog) full list + mark-all-read. |
| 2 | `/notifications` route/page? | **Removed (operator).** Delete `features/notifications/pages/NotificationsPage.tsx` + its route registration; update the M0 route tracker so it is not a dead stub. The bell menu is the only surface. |
| 3 | Spotlight contents? | **Full scrollable list (reusing `NotificationRow`) + "Marcar todas como lidas" (operator).** No search/filter this feature. |
| 4 | Endpoints + generated types exist? | **List/unread/mark-read: yes (F3.1).** `Notification`, `NotificationsListResponse {items,page}`, `UnreadCountResponse {count}`; ops `listNotifications` GET `/notifications?status&cursor&limit`, `getNotificationsUnreadCount` GET `/notifications/unread-count`, `markNotificationRead` POST `/notifications/{id}/read`→204. **Mark-all-read: NO endpoint exists** → F3.4 adds one (Q5). |
| 5 | **Mark-all-read backend? (operator-confirmed approach in Approval)** | Grade-A = a small **additive `POST /notifications/read-all`** on the notifications module (self-scope `CapNotificationRead`; UPDATE the caller's `PENDING/SENT` rows → `READ`, set `read_at`; idempotent; returns `{updated: number}` or `204`). Contract-first: OpenAPI → `oapi-codegen` → `npm run gen:api`; `api-lint -strict`=0; integration-tested; all 6 CI guards. *(Alternative, if operator downgrades: client-side fan-out of `markNotificationRead` over the unread set — N requests, not grade-A. Default = the bulk endpoint.)* |
| 6 | Consume generated or map? | **Consume generated directly** (Q grade-A). The legacy camelCase `NotificationItem` + `NotificationsPanel` are retired. |
| 7 | API client + hooks? | typed `api.GET`/`api.POST` + `asApiError`. `queries/`: `useNotificationsQuery(params)` (`QK.notifications.list`, `staleTime 30_000`), `useUnreadCountQuery()` (`QK.notifications.unreadCount`, `staleTime 30_000`, polled via staleTime/refetch — no SSE), `useMarkNotificationReadMutation()`, `useMarkAllNotificationsReadMutation()` — both invalidate `QK.notifications.list` + `QK.notifications.unreadCount`; toast via `resolveErrorMessage`. |
| 8 | Bell anchor? | `AppToolbar.tsx:43` already has a placeholder bell button → replace with the `NotificationBell` component. |
| 9 | SSE / audit stream? | **Out of scope — no stream rebuilt.** **Deviation (recorded at code time):** the `subscribeOperationsStream`/`listAuditEvents` SSE noops did not live in a shared module — they were members of the dead `features/notifications/useNotifications.ts` hook, which had **zero consumers** (its only callers were the legacy page/panel removed by this feature). Leaving them would have stranded `items: []`/`never[]` noop bodies that the acceptance grep (line ~117) forbids. Resolution: deleted `useNotifications.ts` wholesale as dead code (not an SSE rebuild, no behavior removed from any live path). No live `subscribeOperationsStream` consumer existed to preserve. |
| 10 | Design-source for visual review? | **N/A — none exists; operator specified the interaction directly.** `frontend-screen-reviewer` judges vs `tokens.css` + the reused primitives + the operator-described bell/popover/spotlight UX; design-source N/A recorded in evidence. |
| 11 | Loading/error/empty? | Badge: hidden while loading/error/0 (honest). Popover/spotlight: loading "Carregando notificações…"; error `role="alert"` + retry; empty "Sua caixa está limpa." No fabricated rows. |
| 12 | Popover open/close + a11y? | Controlled open state in `NotificationBell`; close on outside-click + Esc; `aria-haspopup`/`aria-expanded` on the bell; badge has an accessible label (e.g. "3 não lidas"). |

## Consumer contract (FIRST — components consume the generated types)

- **Consumers:** `NotificationBell` (badge + popover) and `NotificationsSpotlight` (Dialog) in `AppToolbar`.
- **Binding shapes (generated, F3.1 + the new F3.4 op — never hand-rolled, HS-3):**
  - `api.GET('/notifications', {params:{query:{limit,…}}})` → `NotificationsListResponse`.
  - `api.GET('/notifications/unread-count')` → `UnreadCountResponse` (`{count}`).
  - `api.POST('/notifications/{id}/read', {params:{path:{id}}})` → `204`.
  - `api.POST('/notifications/read-all')` → `{updated:number}`/`204` (NEW, F3.4 backend leg).
- **Row rendering:** generated `Notification` (snake_case) — `title`, `message`, `event_type`, `created_at`,
  `status` (drives read/unread), `read_at` when set.
- **Query keys:** `QK.notifications.list(params)` (new) + `QK.notifications.unreadCount()` (reserved).

## What this feature implements

### Backend leg (additive — the only backend touch)
1. **`POST /notifications/read-all`** in `internal/modules/notifications/`: OpenAPI path + `oapi-codegen`
   regen + handler + repository method (self-scoped `UPDATE … SET status='READ', read_at=now()
   WHERE recipient_user_id = :caller AND status IN ('PENDING','SENT')`), gated by `CapNotificationRead`
   self-scope (predicate-enforced, like F3.2). Returns `{updated:int}` (or 204). Idempotent. Route guarded
   tier-1 in `permissions.go` per the F3.1 precedent. `npm run gen:api` for the FE type.
   *(No new cap, no new table, no migration — reuses F3.2's table + cap.)*

### Frontend leg
2. **`features/notifications/api/notifications.ts` (rewrite):** re-export generated types
   (`Notification`, `NotificationsListResponse`, `UnreadCountResponse`, list query params from
   `operations['listNotifications']`); `fetchNotifications(params)`, `fetchUnreadCount()`,
   `markNotificationRead(id)`, `markAllNotificationsRead()` — all typed `api.GET`/`api.POST` + `asApiError`.
   **Delete** the `{ items: [] }`/`never[]` noops (list + mark-read). The `subscribeOperationsStream`/`listAuditEvents`
   SSE noops lived in the dead `useNotifications.ts` hook (zero consumers) → deleted wholesale as dead code (see Q9 deviation).
3. **`features/notifications/queries/` (new):** `useNotificationsQuery`, `useUnreadCountQuery`,
   `useMarkNotificationReadMutation`, `useMarkAllNotificationsReadMutation` (Q7).
4. **`features/notifications/lib/` (new, pure + unit-tested):** `formatNotificationTimestamp(created_at)`
   (relative/pt-BR), `eventTypeLabel(event_type)` (chip label). `lib/__tests__/`.
5. **`features/notifications/components/` (new, CSS-Module + token vars):**
   - `NotificationRow.tsx` (+`.module.css`, +test) — one row: title, message, event-type chip, timestamp,
     read/unread state, "Marcar como lida" when unread (fires mark-read mutation). Stateless.
   - `NotificationBell.tsx` (+`.module.css`, +test) — bell button + unread badge (`useUnreadCountQuery`) +
     controlled popover (latest ~8 via `useNotificationsQuery({limit:8})`, loading/error/empty guards,
     `NotificationRow`s, footer "Mostrar Todas" opening the spotlight). Outside-click/Esc close; a11y per Q12.
   - `NotificationsSpotlight.tsx` (+`.module.css`, +test) — `components/ui/Dialog` overlay: full scrollable
     list (`useNotificationsQuery`), "Marcar todas como lidas" (`useMarkAllNotificationsReadMutation`),
     loading/error/empty guards.
6. **`AppToolbar` wire:** replace the placeholder bell button (`AppToolbar.tsx:43`) with `<NotificationBell/>`.
7. **`QK.notifications.list(params)`** added to `lib/queryKeys.ts`.
8. **Remove the page+route:** delete `features/notifications/pages/NotificationsPage.tsx` + its route
   registration; **update the M0 route tracker** (`wiki/implementation/screen-redesign-tracker.md` or the
   M0 evidence) so `/notifications` removal is recorded, not a dead stub.
9. **Retire legacy:** delete `components/NotificationsPanel.tsx` (unused after step 8); remove
   `NotificationItem` from `lib/types/index.ts` **iff** fully unused (if `AppShellHeader.tsx` is itself
   dead/legacy, note it; do **not** refactor a live unrelated component).

## Non-goals (mandatory — in the diff = scope drift, validator C6)

- **No SSE / operations-stream rebuild**; **no real-time push** — unread refreshes via query staleTime/refetch only.
- **No search/filter, no status-filter UI, no pagination UI** beyond the popover's `limit` + spotlight scroll; no preferences/channels (parked mission).
- **No hand-rolled types / no snake→camel mapper** (HS-3).
- **No new/modified shared token or shared primitive** (HS-2) — reuse `Dialog`/`Icon`/`tokens.css`; feature-local `.module.css` only.
- **Backend touch is bounded to the one additive `read-all` endpoint** — no new cap, no migration, no change to F3.1/F3.2/F3.3 behavior, no publish/approval touch.
- **No refactor of unrelated live components**; legacy deletion bounded to genuinely-dead notifications scaffolding.

## Validation Gate (concrete — to be approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| **`read-all` endpoint** — marks only the caller's `PENDING/SENT` rows READ; idempotent; self-scope (user A can't mark B's) | `go test -tags integration ./internal/modules/notifications/... -run TestMarkAllRead` (live PG) | real |
| **`api-lint -strict` = 0** (new path parsed) + **all 6 CI guards green** | `go run ./scripts/api-lint -strict …` → 0; `go vet`/`go build`/cilint | real |
| **`go build`/`go test ./...` green** | `go test ./...` | real |
| **Noop deleted at root** — list + mark-read stubs gone | `grep -nE "items: \[\]\|never\[\]\|Stubbed pending" frontend/apps/web/src/features/notifications/api/notifications.ts` → **0** (list/mark-read) | real (static) |
| **Legacy retired** — no import of `NotificationsPanel`/`WorkspaceViewFrame`/`catalog-`/`NotificationsPage` remains | `grep -rn "NotificationsPanel\|WorkspaceViewFrame\|catalog-\|NotificationsPage" frontend/apps/web/src` → **0** | real (static) |
| **Route removed** — `/notifications` no longer registered; tracker updated | `grep -rn "/notifications" frontend/apps/web/src/**/routes*.tsx` → **0**; tracker diff present | real (static) |
| **Generated-type consumption** — api types from `components['schemas']`/`operations`; no mapper | `npm run typecheck` green | real (type) |
| **List hook → rows** | vitest `useNotificationsQuery.test.ts` (renderHook + QueryClient + `vi.mock` api) | fixture |
| **Unread-count hook → badge** | vitest `useUnreadCountQuery.test.ts` | fixture |
| **Mark-read mutation** — POST `{id}/read` + invalidate list+unread | vitest `useMarkNotificationReadMutation.test.ts` | fixture |
| **Mark-all mutation** — POST `read-all` + invalidate list+unread | vitest `useMarkAllNotificationsReadMutation.test.ts` | fixture |
| **NotificationRow** — renders fields/state; unread → mark-read click fires mutation | vitest `NotificationRow.test.tsx` | fixture |
| **NotificationBell** — badge shows count; click opens popover with rows; "Mostrar Todas" opens spotlight; Esc/outside-click closes | vitest `NotificationBell.test.tsx` | fixture |
| **Spotlight** — full list renders; "Marcar todas como lidas" fires mark-all | vitest `NotificationsSpotlight.test.tsx` | fixture |
| **lib helpers pure + tested** | vitest `lib/__tests__/*.test.ts` | fixture |
| **`frontend-screen-reviewer` APPROVE on record (D2)** — vs tokens/primitives + the operator UX (design-source N/A) | verdict in `evidence.md` | real (review) |
| **`frontend-code-reviewer` APPROVE on record (D2)** | verdict in `evidence.md` | real (review) |
| **FE suite green** | `make test` | real |
| **Build clean** | `npm run build` | real |
| **Runtime functional pass** — bell badge = real unread; popover lists real rows; per-row mark-read flips to READ + badge decrements; spotlight lists all; mark-all flips every unread + badge → 0 (per `screen-qa-checklist.md`) | preview workflow (start → snapshot → click bell → snapshot → mark-read → mark-all → snapshot) | real (runtime) |

> TDD: backend `read-all` written failing-first against live PG; FE hooks/components written failing-first
> against `vi.mock`-ed api responses shaped as generated types, then implemented to green. Runtime live-data
> proof uses F3.3's producers (or one seeded real row).

## ADR needed?

- [x] **No new ADR.** Consumes the F3.1 contract + existing FE architecture rules; the `read-all` endpoint
  is an additive application of the F3.1/F3.2 notifications-module pattern (same cap, same table, same
  self-scope rule). The page→menu re-shape is a UX decision recorded here + in `milestone.md` (operator
  redirect 2026-06-22), not a durable architecture decision. Recorded in `evidence.md`.

## Approval

- [x] **Approved (pre-code) — 2026-06-22 / leandrotca.** Bulk `POST /notifications/read-all` endpoint
  confirmed (designed properly per Grade-A rules; no client-side fan-out workaround).
