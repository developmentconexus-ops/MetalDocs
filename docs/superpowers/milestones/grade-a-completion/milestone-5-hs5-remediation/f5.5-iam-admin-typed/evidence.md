# Evidence — F5.5 iam-admin-typed (Major #4)

> **Status:** CLOSED 2026-06-19 · `GET /api/v1/iam/admin/overview` body materialized through the
> strict-server generated `iamapi.AdminOverviewResponse`; the `map[string]any` response literal and
> the `kpiToOverviewJSON` map helper are gone.

## Change

| File | Change |
|------|--------|
| `internal/modules/iam/delivery/http/admin_handler.go` | `handleAdminOverview` (`:262`) now writes `iamapi.AdminOverviewResponse{Kpi, Presence, RecentActivities}`. Presence built as `[]iamapi.OnlinePresenceItem` (reader branch sets `Status` via `&OnlinePresenceItemStatus(...)`; online-users fallback leaves it nil → omitempty). Events built as `[]iamapi.AuditEventItem`. `kpiToOverviewJSON` (map) **replaced** by `kpiToOverviewTyped` returning `iamapi.IamKpiSnapshot` (+ `[]iamapi.IamKpiRoleCount`). Added `iamapi` import. |
| `internal/modules/iam/delivery/http/admin_handler_test.go` | Added `TestHandleAdminOverview_DecodesIntoGeneratedContract`: strict `DisallowUnknownFields` decode into `iamapi.AdminOverviewResponse`, asserting kpi (`LockedAccounts`, `FailedLogins24h`, `RoleDistribution`), `Presence[0].Status==online`, `RecentActivities[0].Action`. Added `iamapi` import. |

**Field-type conversions** (explicit, no semantic change): `KpiSnapshot.FailedLogins24h` int64 →
generated `int`; `RoleCount.Role` (domain `Role`) → `iamapi.UserRole`; presence `Status`
(`iampresence.Status`) → `*OnlinePresenceItemStatus`.

## TDD record

**Refactor-class — structural lift, behavior parity.** The wire keys were already the generated
shape (established by the F2.2 contract-emit work), so the new strict-decode test is **green on the
pre-change wire too** — honestly labeled: the red→green discipline here is at the *structural* gate
(Validation Gate #1: the overview route's `map[string]any` materialization goes away), and the
existing three overview tests (`_DropsUsersField_ReturnsTypedShape`, `_PresenceCarriesStatus`,
`_TenantIsolation`) are the no-regression proof that the wire output is unchanged. The new
`DisallowUnknownFields` test is the durable contract lock: any future drift between the route body
and `iamapi.AdminOverviewResponse` now fails the suite.

## Validation Gate results (real output)

1. **No untyped map on the overview route** — `grep -n "map\[string\]any" admin_handler.go` →
   three lines, **none in `handleAdminOverview`/`kpiToOverviewTyped`**: `:249` (the audit `payload`
   decode target — `AuditEventItem.Payload` IS `map[string]interface{}` by contract, correct);
   `:341`/`:378` (role upsert/replace handler responses — out of M5 scope, see Defers).
2. **Typed response used** — `grep -n "iamapi.AdminOverviewResponse\|kpiToOverviewTyped"
   admin_handler.go` → `:264` writes `iamapi.AdminOverviewResponse{...}`, `:265` calls
   `kpiToOverviewTyped`.
3. **Build** — `go build ./...` → `BUILD OK`.
4. **Module suite** — `go test -count=1 ./internal/modules/iam/...` → all packages `ok`
   (application/authz/delivery-http/domain/infra-memory/infra-postgres/presence).
5. **Contract conformance** — `TestHandleAdminOverview_DecodesIntoGeneratedContract` green: body
   decodes strictly into `iamapi.AdminOverviewResponse` with no unknown fields; typed kpi + presence
   status + recent-activity round-trip asserted.

## Fixture-vs-real

Route-level `httptest` against in-test stub readers (kpi/presence/audit) + the real handler. F5.5 is
wire-shape only — no live SQL. No fixture stands in for a live path.

## Wire note

`last_seen_at` / `occurred_at` previously emitted via `.Format(time.RFC3339)` (second precision); now
emitted by `time.Time` JSON marshalling (RFC3339Nano). Both satisfy the OpenAPI `format: date-time`
declaration and parse identically in the FE; for the test's zero-nanosecond timestamps the bytes are
identical. Benign precision normalization toward the declared contract type.

## Defers

**Mention-don't-fix (bounded):** the role upsert/replace handler responses at
`admin_handler.go:341` (`POST .../roles`) and `:378` (`PUT .../roles`) still return `map[string]any`
— same class as Major #4 but **not** among M5's 7 cited sites. The code comment at `:290-294`
assigns the roles endpoints to **PR-5** (Roles & Caps matrix) restructuring. **Trigger:** retype
both to their generated response types when PR-5 restructures these endpoints. Widening F5.5 to cover
them would breach M5 surgical appetite (HS-6).
