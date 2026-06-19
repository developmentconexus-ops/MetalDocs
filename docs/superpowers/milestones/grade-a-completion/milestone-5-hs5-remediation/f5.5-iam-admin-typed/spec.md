# Feature F5.5 — IAM admin overview typed response (Major #4)

> **Milestone:** 5 — HS-5 remediation  ·  **Feature:** `f5.5-iam-admin-typed`
> **Status:** Approved 2026-06-19 — code change may begin.

## Consumer contract

**Consumer:** the Admin Center Overview tab (FE), via `GET /api/v1/iam/admin/overview`, served by
`(*AdminHandler).handleAdminOverview` in `internal/modules/iam/delivery/http/admin_handler.go`.

The route currently materializes its body through a nested `map[string]any` literal at
`admin_handler.go:262` (the top-level `{kpi, presence, recent_activities}` map) plus the
`kpiToOverviewJSON` map helper (`:272`). Re-audit **Major #4** = a public IAM route returning an
untyped map.

**What the consumer needs:** the body materialized through the **strict-server generated** type
already declared for this exact operation — `iamapi.AdminOverviewResponse` (`internal/modules/iam/api/api.gen.go:266`):

```go
type AdminOverviewResponse struct {
	Kpi              IamKpiSnapshot       `json:"kpi"`
	Presence         []OnlinePresenceItem `json:"presence"`
	RecentActivities []AuditEventItem     `json:"recent_activities"`
}
```

with the nested generated types `IamKpiSnapshot` (+ `IamKpiRoleCount`), `OnlinePresenceItem`
(+ `*OnlinePresenceItemStatus`), `AuditEventItem`. The `iamapi` alias is already imported by sibling
handlers in this package (`people_handler.go:27`, `routes_roles_caps.go:7`).

**Wire output is unchanged.** The current map literal already emits exactly the generated shape
(the F2.2 contract-emit work established it). F5.5 is a **structural retype** — kill the
`map[string]any`, keep byte-identical JSON. This is the same mechanical pattern as F5.3/F5.4.

**Required shape after this feature:**

1. `handleAdminOverview` builds and writes `iamapi.AdminOverviewResponse` (not a map literal).
2. `kpiToOverviewJSON` is **replaced** by a typed builder returning `iamapi.IamKpiSnapshot`
   (it has exactly one caller — `handleAdminOverview` — confirmed by GitNexus).
3. Field-type conversions handled explicitly:
   - `KpiSnapshot.FailedLogins24h` is `int64` → generated `IamKpiSnapshot.FailedLogins24h` is `int`:
     `int(k.FailedLogins24h)`.
   - `RoleCount.Role` (domain `Role`) → `IamKpiRoleCount.Role` (`iamapi.UserRole`):
     `iamapi.UserRole(string(rc.Role))`.
   - presence `Status` (`iampresence.Status`) → `*OnlinePresenceItemStatus`: set the pointer in the
     presence-reader branch; leave it `nil` (omitempty) in the legacy online-users fallback branch
     (matches today's output, which omits `status` there).

## Interview record (B1.5)

| Question | Resolution | Source |
|----------|-----------|--------|
| Generated type for this op exists? | Yes — `AdminOverviewResponse` + nested types, mapped 1:1 to the current map keys. | `api.gen.go:266` |
| Is wire output changing? | No. Current map already emits the generated shape (F2.2). Retype is byte-identical. | `admin_handler.go:262`, `api.gen.go:266` |
| `kpiToOverviewJSON` shared with `/iam/kpi`? | No — exactly one caller (`handleAdminOverview`). Safe to replace, not duplicate. | GitNexus call-graph |
| Field-type mismatches? | `FailedLogins24h` int64→int; `Role`→`iamapi.UserRole`; presence `Status`→`*OnlinePresenceItemStatus`. All explicit conversions, no semantic change. | domain vs `api.gen.go` |
| `writeJSON` accepts a struct? | Yes — `var writeJSON = httpresponse.WriteJSON` (`middleware.go:17`), payload is `any`. | grep |
| Presence else-branch status? | Today's fallback (online-users) omits `status`; typed nil-pointer omitempty preserves that. | `admin_handler.go:236-243` |

## Non-goals

- No OpenAPI/spec change.
- No change to the three composition reads (KPI / presence / audit) or the errgroup concurrency.
- No frontend change (wire keys + values unchanged).
- Do not touch the `/iam/kpi` standalone endpoint or `ObservabilityHandler` (out of scope; Major #4
  names only the overview route).

## Validation Gate

1. **No untyped map on the overview route:** the `:262` response literal, the `kpiToOverviewJSON`
   helper, and the presence/event intermediate `map[string]any` builders are all gone — the overview
   body is materialized solely through `iamapi.AdminOverviewResponse`. Residual `map[string]any` in
   the file is limited to: (a) the audit `payload` decode target (`AuditEventItem.Payload` IS a
   `map[string]interface{}` by contract — correct, not a finding); (b) the role upsert/replace
   handler responses at `:341`/`:378` — **out of M5 scope** (not among the 7 cited sites; the code
   comment at `:290-294` assigns the roles endpoints to PR-5 restructuring). Recorded as a bounded
   defer with trigger = PR-5. Verify: `grep -n "map\[string\]any" admin_handler.go` shows only those
   three lines, none in `handleAdminOverview`/`kpiToOverviewTyped`.
2. **Typed response used:** `handleAdminOverview` constructs `iamapi.AdminOverviewResponse`.
3. **Build:** `go build ./...` clean.
4. **Tests:** `go test -count=1 ./internal/modules/iam/...` all green — including the existing
   `admin_handler_test.go` overview tests (`TestHandleAdminOverview_DropsUsersField_ReturnsTypedShape`,
   `_PresenceCarriesStatus`, `_TenantIsolation`), which decode the body and assert keys/values that
   are unchanged.
5. **Contract conformance (new test):** a test decodes the overview body strictly
   (`json.Decoder` + `DisallowUnknownFields`) into `iamapi.AdminOverviewResponse` and asserts a
   representative typed field round-trips (kpi + presence-status + recent-activity), locking the
   route to the generated contract so future drift fails.
