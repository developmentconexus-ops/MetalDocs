# Feature F5.5 — Plan

## Files touched
- `internal/modules/iam/delivery/http/admin_handler.go` — retype the overview response.
- `internal/modules/iam/delivery/http/admin_handler_test.go` — add strict-contract conformance test.

## Tasks (TDD)
1. **Red:** add `TestHandleAdminOverview_DecodesIntoGeneratedContract` — wire stub kpi + presence
   readers, GET the route, decode body with `json.NewDecoder(...).DisallowUnknownFields()` into
   `iamapi.AdminOverviewResponse`, assert: kpi field (e.g. `LockedAccounts`), `Presence[0].Status`
   == `online`, `RecentActivities[0].Action` round-trip. (May pass on current wire — see note.)
2. **Green/refactor:** in `admin_handler.go`
   - add `import iamapi "metaldocs/internal/modules/iam/api"`.
   - replace the presence `map[string]any` builders (both branches) with `[]iamapi.OnlinePresenceItem`;
     reader branch sets `Status` via `s := iamapi.OnlinePresenceItemStatus(string(item.Status)); &s`,
     fallback branch leaves it nil.
   - replace the event `map[string]any` builder with `[]iamapi.AuditEventItem` (payload stays
     `map[string]interface{}`).
   - replace `kpiToOverviewJSON` with `kpiToOverviewTyped(k iamdomain.KpiSnapshot) iamapi.IamKpiSnapshot`
     building `IamKpiRoleCount` slice; convert `int(k.FailedLogins24h)` and
     `iamapi.UserRole(string(rc.Role))`.
   - replace the `:262` map literal with `iamapi.AdminOverviewResponse{Kpi:..., Presence:..., RecentActivities:...}`.
3. **Verify gate:** grep `map[string]any` in `admin_handler.go` → 0; `go build ./...`;
   `go test -count=1 ./internal/modules/iam/...`.

## Test strategy
- New strict-decode test locks the contract.
- The 3 existing overview tests are the regression proof of byte-identical wire output.

## Note on TDD red
Wire output already conforms (F2.2), so the new strict-decode test may be green before the retype.
That is honest: F5.5 is a structural quality fix, not a wire change. The red→green discipline is
satisfied at the *structural* gate (Validation Gate #1: `map[string]any` count goes 5→0); the
existing+new tests are the no-regression proof. Labeled as such in evidence.
