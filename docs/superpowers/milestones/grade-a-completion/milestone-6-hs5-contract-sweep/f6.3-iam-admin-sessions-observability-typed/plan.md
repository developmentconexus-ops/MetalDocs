# F6.3 — Plan

## Tasks

1. **admin_handler.go** — Replace 2 `map[string]any` writeJSON calls:
   - Line ~341: `map[string]any{user_id, role, display_name}` → `iamapi.UpsertUserRoleResponse`
   - Line ~378: `map[string]any{user_id, display_name, roles}` → `iamapi.ReplaceUserRolesResponse`
   - Note: `iamapi` import already present (line 16)
   - `DisplayName *string` in `UpsertUserRoleResponse` — use ptr helper for non-empty or nil for empty
   - Run `go build ./...`

2. **sessions_handler.go** — Replace `[]map[string]any` + outer `map[string]any`:
   - Add import `iamapi "metaldocs/internal/modules/iam/api"`
   - Replace `out []map[string]any` slice with `out []iamapi.SessionItem`
   - Per-item: map `sql.NullTime` → `time.Time` (zero time for invalid)
   - Optional fields: `IpAddress *string`, `UserAgent *string`, `DeviceLabel *string`
   - Outer writeJSON: `iamapi.ListSessionsResponse{Items: out, Page: iamapi.CursorPage{HasMore: false}}`
   - Run `go build ./...`

3. **observability_handler.go** — Retype two helper functions:
   - `usageToJSON(u iamdomain.UsageSnapshot) iamapi.UsageSnapshot`
     - `int64` → `int` casts for CountWindows and StorageUsage fields
     - `PlanTier`: `*iamapi.UsageSnapshotPlanTier` — nil when empty
   - `kpiToJSON(k iamdomain.KpiSnapshot) iamapi.IamKpiSnapshot`
     - `FailedLogins24h int64` → `int` cast
     - `RoleDistribution []iamapi.IamKpiRoleCount` with `UserRole(string(rc.Role))`
   - Add import `iamapi "metaldocs/internal/modules/iam/api"`
   - Run `go build ./...`

4. **Full test run**: `go test -count=1 ./...`

5. **H-D grep proof**: confirm 0 hits in the three files

6. **Commit**: `feat(m6/f6.3): iam admin/sessions/observability typed responses`
