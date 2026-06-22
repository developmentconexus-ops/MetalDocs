# Feature F3.2 — Spec (notifications-backend)

> **Milestone:** 3 — Notifications (full-stack; surface + document-lifecycle emitters)  ·  **Folder:** `f3.2-notifications-backend`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-22 / leandrotca — *implementation may begin.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (`plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Consumer-contract discovery was driven by (a) the F3.1 locked contract (`api.gen.go` `Notification`
shape + the three endpoints) and ADR-0043, (b) a full read-only conformance recon of the established
Grade-A backend patterns (distribution module as the newest read-surface precedent; the 6 CI guards;
the tripwire-pairing rule; the hgcrossmodule ownership manifest; the permissions route-table tests),
and (c) operator decisions below. Q&A persisted.

| # | Question | Answer |
|---|----------|--------|
| 1 | Full table shape now (with the F3.3 projector idempotency column) or minimal-now/alter-later? | **Full table in F3.2** (operator, 2026-06-22). `source_event_id uuid` + a partial unique index `(recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL` ship in the initial migration so F3.3 needs no `ALTER`. Forward-only migrations are cheaper right-once than altered later; F3.2 is the architectural owner of the full notifications table. |
| 2 | Module layout + package naming? | Mirror **distribution** (M2, newest read-surface precedent): `api/` (F3.1, untouched) + `domain/` (`notificationsdomain`) + `delivery/http/` (`notificationshttp`) + `infrastructure/` (`notificationsinfra`). Prefixed package names per the distribution bar (search/audit use older unprefixed style — not the current standard). |
| 3 | Where does the `Repository` interface live? | Consumer-defined in the delivery handler file (`delivery/http/handler.go`), exactly as distribution/audit/search. The infra `*NotificationsRepository` satisfies it structurally. |
| 4 | How is self-scope enforced? | **Two layers.** Tier-1 route guard requires `CapNotificationRead` (`permissions.go`, mirrors the `CapAuditRead`/`CapDistributionRead` precedent). **And** every query carries `recipient_user_id = $caller` in the SQL predicate — the cap alone never returns another user's rows (R3 mitigation). The caller id comes from `authn.UserIDFromContext`; tenant from `tenant.FromContext`. |
| 5 | MarkRead behavior on wrong-owner / already-read / missing id? | **Silent idempotent 204.** `UPDATE ... WHERE id=$1 AND recipient_user_id=$2 AND tenant_id=$3 AND status != 'READ'`; 0 rows affected → still 204. No existence leak (same 204 for "yours-already-read", "not yours", "doesn't exist"). Extract-failure (no auth context) → 404 (the generated MarkRead has a 404 type; defensive, mirrors distribution's no-tenant→404). |
| 6 | Does the mutating `MarkRead` trip the `tripwire-pairing` CI guard? | **Yes — recon-decisive.** The guard (`scripts/api-lint/code_rules.go:146`) flags any `*repository.go` function with INSERT/UPDATE/DELETE and no `authz.Require` in the same body. `MarkRead`'s authz is enforced one layer up (tier-1 cap) + the self-scope predicate — the repo method is intentionally dumb, exactly like the 13 allowlisted documents/approval/CD writes. **Resolution:** add a rationale-commented entry to `scripts/api-lint/tripwire-allowlist.txt`. This is the established, reviewed path — not a new pattern. |
| 7 | Does the new table need registration in the hgcrossmodule ownership manifest? | **Yes, for completeness.** Own-table reads never flag (`hgcrossmodule.go:191` skips `owner == reader`), so F3.2 passes regardless — but Grade-A registration adds `"notifications": "notifications"` to `hgOwnerByTable` so a *future* foreign module reading the notifications base table raw is caught by the guard. |
| 8 | Display-name (iam) port dependency, like distribution? | **No.** Notification rows carry `title`/`message` pre-rendered at projection time (F3.3). The read surface returns stored columns verbatim — no `UserDisplayNameReader`, no cross-module port, no N+1. The repo constructor takes only `*sql.DB`. |
| 9 | Cursor pagination shape? | Canonical `internal/platform/pagination` keyset. `ORDER BY created_at DESC, id DESC`; cursor = `EncodeCursor(created_at.UTC().Format(RFC3339Nano), id)`; keyset filter `(created_at, id) < (cursor_ts, cursor_id)`; `limit+1` to derive `has_more`; `ClampLimit`. Matches the audit/distribution keyset dialect. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):**
  - **F3.1 generated contract** (`internal/modules/notifications/api/api.gen.go` `StrictServerInterface`) —
    F3.2 implements this interface. This is the binding shape.
  - **FE notifications surface (F3.4)** consumes the generated FE types downstream; F3.2 must keep the
    runtime response structurally equal to the generated `Notification` schema.
  - **F3.3 projector** writes rows into the table F3.2 owns (per-recipient, `status` read-state,
    `source_event_id` idempotency key).
- **Contract (the three operations from `StrictServerInterface`, F3.1):**
  - `ListNotifications(ctx, ListNotificationsRequestObject) → 200 NotificationsListResponse | 400 | 401 | 403 | 500`
    — `{ items: []Notification, page: CursorPage }`, newest-first, self-scoped, optional `status` filter, keyset cursor.
  - `GetNotificationsUnreadCount(ctx, …) → 200 UnreadCountResponse | 401 | 403 | 500` — `{ count }` of the caller's `PENDING`+`SENT` rows.
  - `MarkNotificationRead(ctx, {Id}) → 204 | 401 | 403 | 404 | 500` — idempotent, self-scoped.
  - `Notification` shape (generated, `api.gen.go:88`): `id, recipient_user_id, event_type, resource_type, resource_id, title, message, status: PENDING|SENT|READ, created_at, read_at?`.
- **Source of truth for the contract:** the F3.1 generated `notificationsapi.StrictServerInterface` +
  `Notification`/`NotificationsListResponse`/`UnreadCountResponse`/`CursorPage` types (already on disk,
  `internal/modules/notifications/api/api.gen.go`). F3.2 implements; it does not redefine the shape.

## What this feature implements

1. **Migration `db/migrations/0247_notifications_table.sql`** (forward-only, idempotent):
   `metaldocs.notifications` — `id uuid PK`, `tenant_id uuid NOT NULL`, `recipient_user_id text NOT NULL`,
   `event_type text NOT NULL`, `resource_type text NOT NULL`, `resource_id text NOT NULL`,
   `title text NOT NULL`, `message text NOT NULL`,
   `status text NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','SENT','READ'))`,
   `created_at timestamptz NOT NULL DEFAULT now()`, `read_at timestamptz`, `source_event_id uuid`.
   Index `(tenant_id, recipient_user_id, created_at DESC, id DESC)` for the keyset list.
   Partial unique index `(recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL`
   (F3.3 idempotency; NULL rows — e.g. test fixtures — unconstrained). RLS enabled + forced with the
   verbatim 0237 NULL-permissive `tenant_isolation` policy on `metaldocs.tenant_id` GUC.
2. **`domain/types.go`** (`notificationsdomain`): `NotificationRow` + `NotificationsPage` DTOs (breaks the
   infra→delivery import cycle, per the distribution `domain/types.go` precedent).
3. **`infrastructure/notifications_repository.go`** (`notificationsinfra`, ctor takes `*sql.DB`):
   - `List(ctx, tenantID, recipientUserID, statusFilter, cursor string, limit int) (NotificationsPage, error)` —
     two-branch keyset (first page / cursor), optional status via `($N::text IS NULL OR status = $N)`,
     explicit `tenant_id` + `recipient_user_id` predicate, `limit+1` for `has_more`, `ErrInvalidCursor`
     on malformed cursor.
   - `UnreadCount(ctx, tenantID, recipientUserID string) (int, error)` — `COUNT(*)` of `status IN ('PENDING','SENT')`.
   - `MarkRead(ctx, tenantID, notificationID, recipientUserID string) error` — self-scoped idempotent UPDATE.
4. **`delivery/http/handler.go`** (`notificationshttp`): implements `notificationsapi.StrictServerInterface`;
   `Repository` interface declared here (consumer-defined); `extractTenantAndUser` helper; `toProblem`
   mapper (distribution clone). Error mapping per the F3.1 generated response types (see Validation Gate).
5. **`delivery/http/routes.go`** (`notificationshttp`): `RegisterRoutes(h, mux)` — distribution clone
   (`NewStrictHandler` + `HandlerWithOptions{BaseURL:"/api/v1"}`).
6. **`apps/api/cmd/metaldocs-api/main.go`**: construct `notificationsinfra.NewNotificationsRepository(deps.SQLDB)` →
   `notificationshttp.NewHandler(repo)` → `RegisterRoutes(handler, mux)`, after the distribution block.
7. **`apps/api/cmd/metaldocs-api/permissions.go`**: three tier-1 route rules guarded by
   `CapNotificationRead`, most-specific-first, before any `/api/v1/notifications` catch-all:
   `GET /api/v1/notifications/unread-count` (exact), `POST /api/v1/notifications/{id}/read` (prefix+suffix `/read`),
   `GET /api/v1/notifications` (exact).
8. **`scripts/api-lint/tripwire-allowlist.txt`**: add
   `internal/modules/notifications/infrastructure/notifications_repository.go|MarkRead` with a rationale
   comment (tier-2 authz one layer up: tier-1 `CapNotificationRead` + self-scope SQL predicate).
9. **`tools/cilint/internal/analyzers/hgcrossmodule.go`**: add `"notifications": "notifications"` to
   `hgOwnerByTable`.
10. **`tests/integration/testdb/factory.go`**: new `NewNotification(t, db, opts...)` builder + `Notification`
    fixture struct (mints UUID, auto-seeds tenant/user FK parents, schema-qualified insert, per the
    `NewControlledDoc` shape — no projector; seeds a row directly).

## Non-goals (mandatory)

Anything here appearing in the diff is scope drift (validator C6).

- **No emitter / projector** — F3.3 owns production of rows from `governance_events`. No
  `governance_events` read, no `v_cd_obligated_readers` read, no `v_approval_instance_submitter` /
  `v_document_cd_mapping` views (those are F3.3 surfaces per ADR-0043 §6).
- **No edit to `publish_service.go` or any existing module's emit code** (`git diff` empty — milestone gate).
- **No FE wire** (F3.4) — `notifications.ts` / `NotificationsPage` stay untouched.
- **No new capability** — `CapNotificationRead` already exists (F3.1). No cap seeding (operator grants separately).
- **No display-name / iam port** — rows are self-contained.
- **No SSE / stream / channel / preference** — parked emitter mission.
- **No change to `v_cd_grantee` / `v_cd_obligated_readers`** or any search/approval/CD/iam base table.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Migration applies forward-only; table + indexes + RLS present | `.\scripts\check-system-runnable.ps1` (migrates clean); `\d metaldocs.notifications` shows the CHECK, both indexes, RLS enabled+forced | real |
| Migration sequence gapless | `migration-gapless` guard green (0247 follows 0246, no gap) | real |
| `StrictServerInterface` implemented; compiles | `go build ./...` exit 0 (handler satisfies the F3.1 interface) | real |
| **Self-scope isolation** — user A never sees B's rows | `go test -tags integration ./internal/modules/notifications/... -run TestNotifications/self_scope_isolation` | real (live PG) |
| **Unread-count accuracy** — counts PENDING+SENT only | `…-run TestNotifications/unread_count_accuracy` | real |
| **Mark-read flips + idempotent** — status→READ, read_at set, re-run no-op | `…-run TestNotifications/mark_read_flips_and_idempotent` | real |
| **Mark-read wrong-owner no-op** — A marking B's row changes nothing | `…-run TestNotifications/mark_read_wrong_owner_noop` | real |
| **Cursor stability** — 25 rows, limit 10, 3 stable DESC pages, no dup/skip | `…-run TestNotifications/cursor_stability` | real |
| **Status filter** — `status=READ` returns only READ | `…-run TestNotifications/status_filter` | real |
| `api-lint -strict` = 0 (incl. tripwire-pairing: MarkRead allowlisted) | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` → `0 violation(s)` | real |
| All 6 CI guards green (cilint/hgcrossmodule, module-boundaries, test-discipline, migration-gapless, api-design-system, spec-base-path) | `go vet ./...`; `go test ./tools/cilint/...`; `go build ./...` | real |
| Route table validated — 3 new routes resolve to `CapNotificationRead`, no fallthrough | `go test ./apps/api/cmd/metaldocs-api/ -run 'TestPermissionResolver|TestRouteCoverage|TestEveryRouteCapInRegistry'` (new cases added) | real |
| Publish path untouched | `git diff --quiet -- internal/modules/documents/approval/application/publish_service.go` | real |
| `go test ./...` green | `go test ./...` | real |

> TDD: the integration tests (self-scope, unread-count, mark-read idempotency, cursor) are written
> failing-first against the live-PG repo before the repo methods exist, then implemented to green.
> No fixture-only substitution for the self-scope or idempotency proofs — they run against real Postgres.

## ADR needed?

- [x] **No new durable decision — covered by ADR-0043.** ADR-0043 (F3.1) already records the notifications
  module, the per-recipient read-state table shape, `CapNotificationRead` self-scope, and the lifecycle-bundle
  decision. F3.2 *implements* that recorded decision; it makes no new architectural choice. The two
  conformance touch-points it adds (tripwire allow-list entry for the self-scoped `MarkRead`; ownership-manifest
  registration of the `notifications` table) are mechanical applications of existing rules, not new decisions —
  recorded in `evidence.md`, no ADR.
