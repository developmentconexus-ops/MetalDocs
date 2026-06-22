# Feature F3.2 — Evidence

> **Milestone:** 3 — Notifications (full-stack)  ·  **Feature:** `f3.2-notifications-backend`  ·  **Closed:** 2026-06-22
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).

## What was implemented

- **Migration `db/migrations/0247_notifications_table.sql`** — `metaldocs.notifications` table (12 cols, `status` CHECK, list keyset index, partial unique index for F3.3 idempotency, RLS 0237 pattern). Forward-only, idempotent. Commit `13a95e11`.
- **`internal/modules/notifications/domain/types.go`** — `NotificationRow` + `NotificationsPage` DTOs (`notificationsdomain` pkg). Commit `dd92af7b`.
- **`tests/integration/testdb/factory.go`** — `Notification` struct + `NewNotification` builder (no projector; seeds rows directly; auto-wires tenant/user parents). `WithRecipient`/`WithEventType`/`WithResourceType`/`WithResourceID`/`WithStatus` opts added. Commit `8923da94`.
- **`internal/modules/notifications/infrastructure/notifications_repository.go`** — `NotificationsRepository` with `List` (keyset, two-branch first/cursor, `NULL`-safe status filter, `limit+1` has_more), `UnreadCount` (PENDING+SENT self-scoped), `MarkRead` (self-scoped idempotent UPDATE, 0-rows = silent no-op). `MarkRead` tripwire-allowlisted. Commit `ab29c7eb`.
- **`tools/cilint/internal/analyzers/hgcrossmodule.go`** — `"notifications": "notifications"` ownership entry. Commit `0c4f3377`.
- **`internal/modules/notifications/delivery/http/handler.go` + `routes.go`** — `notificationshttp` pkg; `Repository` interface consumer-defined; `Handler` satisfies `notificationsapi.StrictServerInterface`; `ListNotifications` / `GetNotificationsUnreadCount` / `MarkNotificationRead`. `toProblem` + `toAPINotification` + `extractTenantAndUser` helpers. Commit `54b4998b`.
- **`apps/api/cmd/metaldocs-api/main.go`** — wires `notificationsinfra.NewNotificationsRepository(deps.SQLDB)` → `notificationshttp.NewHandler` → `RegisterRoutes`. Commit `05b935eb`.
- **`apps/api/cmd/metaldocs-api/permissions.go`** — three tier-1 rules: `GET /api/v1/notifications/unread-count` (exact), `POST /api/v1/notifications/{id}/read` (prefix+suffix), `GET /api/v1/notifications` (exact), all `CapNotificationRead` + `VisibilityPermissionGuarded`. Commit `4566867f`.

Producer matches consumer contract: `Handler` implements the F3.1 generated `notificationsapi.StrictServerInterface`; `toAPINotification` maps stored rows to the generated `notificationsapi.Notification` shape verbatim.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing tests first, then green | Integration tests written before `NewNotificationsRepository` existed; verified compile-fail before impl | All 6 subtests green (see integration row) | real (live PG) |
| `go build ./...` | `go build ./...` | exit 0, no output | — |
| `go vet ./...` | `go vet ./...` | exit 0, no output | — |
| `go test ./...` (unit suite) | `go test ./...` | all packages PASS | — |
| **Self-scope isolation** | `go test -tags integration ./internal/modules/notifications/... -run TestNotifications/self_scope_isolation -v` | `--- PASS: TestNotifications/self_scope_isolation (0.71s)` | real (live PG) |
| **Status filter** | `…-run TestNotifications/status_filter` | `--- PASS (0.59s)` | real |
| **Cursor stability** (25 rows, 3 pages, no dup/skip) | `…-run TestNotifications/cursor_stability` | `--- PASS (3.92s)` | real |
| **Unread count accuracy** (PENDING+SENT only, self-scoped) | `…-run TestNotifications/unread_count_accuracy` | `--- PASS (1.17s)` | real |
| **Mark-read flips + idempotent** | `…-run TestNotifications/mark_read_flips_and_idempotent` | `--- PASS (0.79s)` | real |
| **Mark-read wrong-owner no-op** | `…-run TestNotifications/mark_read_wrong_owner_noop` | `--- PASS (0.93s)` | real |
| Full integration suite | `go test -tags integration ./internal/modules/notifications/... -v` | `--- PASS: TestNotifications (83.83s)`, all 6 subtests PASS | real |
| `api-lint -strict` = 0 | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` | — |
| cilint / hgcrossmodule | `go test ./tools/cilint/...` | `ok metaldocs/tools/cilint/internal/analyzers` | — |
| Permission resolver — 3 notifications cases | `go test -count=1 ./apps/api/cmd/metaldocs-api/ -run TestPermissionResolver -v` | `PASS: notifications_list`, `notifications_unread-count`, `notifications_mark-read` | — |
| Publish path untouched | `git diff --quiet -- internal/modules/documents/approval/application/publish_service.go && echo CLEAN` | `CLEAN` | — |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Migration applies forward-only; table + indexes + RLS present | yes | `0247_notifications_table.sql` committed; `check-system-runnable.ps1` would validate schema (system runnable pre-condition for F3.3) |
| Migration sequence gapless | yes | `go test ./tools/cilint/...` PASS (migration-gapless guard) |
| `StrictServerInterface` implemented; compiles | yes | `go build ./...` exit 0 |
| Self-scope isolation — user A never sees B's rows | yes | `TestNotifications/self_scope_isolation` PASS (real PG) |
| Unread-count accuracy — counts PENDING+SENT only | yes | `TestNotifications/unread_count_accuracy` PASS (real PG) |
| Mark-read flips + idempotent — status→READ, read_at set, re-run no-op | yes | `TestNotifications/mark_read_flips_and_idempotent` PASS (real PG) |
| Mark-read wrong-owner no-op | yes | `TestNotifications/mark_read_wrong_owner_noop` PASS (real PG) |
| Cursor stability — 25 rows, 3 stable DESC pages, no dup/skip | yes | `TestNotifications/cursor_stability` PASS (real PG) |
| Status filter — `status=READ` returns only READ | yes | `TestNotifications/status_filter` PASS (real PG) |
| `api-lint -strict` = 0 (incl. tripwire-pairing: MarkRead allowlisted) | yes | `0 violation(s)` |
| All 6 CI guards green | yes | `go test ./tools/cilint/...` PASS + `go build`/`vet`/`test` exit 0 |
| Route table validated — 3 new routes resolve to `CapNotificationRead` | yes | `TestPermissionResolver/notifications_*` all PASS |
| Publish path untouched | yes | `git diff --quiet … publish_service.go` → `CLEAN` |
| `go test ./...` green | yes | all packages PASS |

## Review disposition

- **Spec-compliance:** Implementation follows the distribution module precedent exactly (four-layer layout, consumer-defined `Repository` interface, `toProblem` helper, `RegisterRoutes` via `HandlerWithOptions`). Self-scope enforced in SQL predicate AND tier-1 cap (two layers as spec). `MarkRead` tripwire-allowlisted with rationale comment. Hgcrossmodule ownership registered. No publish-path edit, no FE wire, no emitter — non-goals hold. Compliant.
- **Code-quality:** `go vet` clean; `go build` clean; `go test ./...` clean. No cross-module raw reads (`hgcrossmodule` PASS). Standard platform pagination (keyset, `ClampLimit`, `EncodeCursor`/`DecodeCursor`). `MarkRead` correctly uses `status != 'READ'` to keep `read_at` stable on re-runs. No TODOs or stubs in delivered files.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| F3.3 projector (emitter) | Intentionally not in F3.2 (non-goal); F3.3 is the next feature in the milestone plan | F3.3 execution |
| F3.4 FE wire | Intentionally not in F3.2 (non-goal) | F3.4 execution |
| `check-system-runnable.ps1` schema verification (`\d metaldocs.notifications`) | API server not started in this session (pure backend library + test pass is sufficient for F3.2); runtime boot verified at F3.3 system-runnable gate | F3.3 Task 1 system-runnable check |
