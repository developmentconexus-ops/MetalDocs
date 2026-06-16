# Feature F4.2 — Plan

> **Spec:** `spec.md` (approved 2026-06-16).

## Tasks

### T1 — New IAM port interface + Noop
**File:** `internal/modules/iam/domain/admin_role_member_reader_port.go` (new)

`AdminRoleMemberReader` interface with `AdminRoleMembers(ctx, tenantID, roleCodes) (map[string]string, error)`.
`NoopAdminRoleMemberReader` returning empty map.

### T2 — Postgres implementation
**File:** `internal/modules/iam/infrastructure/postgres/admin_role_member_repository.go` (new)

`AdminRoleMemberRepository` queries `iam_user_roles` once:
```sql
SELECT user_id, MIN(role_code) FROM iam_user_roles WHERE tenant_id=$1 AND role_code=ANY($2) GROUP BY user_id
```
Compile-time guard: `var _ iamdomain.AdminRoleMemberReader = (*AdminRoleMemberRepository)(nil)`.

### T3 — Security repository: add field + update NewRepository
**File:** `internal/modules/security/infrastructure/postgres/repository.go`

- Add `adminRoles iamdomain.AdminRoleMemberReader` to `Repository` struct.
- Add 4th arg `adminRoles iamdomain.AdminRoleMemberReader` to `NewRepository`; nil → Noop.
- Rewrite `ListOffHoursAdminActions`: call `r.adminRoles.AdminRoleMembers` to get userRoles map,
  query `audit_events` with `actor_id = ANY($userIDs)` (no JOIN), attach role from map.

### T4 — Wire in main.go
**File:** `apps/api/cmd/metaldocs-api/main.go`

Pass `iampg.NewAdminRoleMemberRepository(sqlDB)` as 4th arg to `securitypg.NewRepository`.

### T5 — Update integration test (drive-by)
**File:** `internal/modules/security/infrastructure/postgres/repository_displayname_integration_test.go`

Pass `iampg.NewAdminRoleMemberRepository(db)` as 4th arg to `securitypg.NewRepository`.

## Files touched

| File | Change |
|------|--------|
| `internal/modules/iam/domain/admin_role_member_reader_port.go` | T1: new interface + Noop |
| `internal/modules/iam/infrastructure/postgres/admin_role_member_repository.go` | T2: new impl |
| `internal/modules/security/infrastructure/postgres/repository.go` | T3: struct + NewRepository + ListOffHoursAdminActions rewrite |
| `apps/api/cmd/metaldocs-api/main.go` | T4: wiring |
| `internal/modules/security/infrastructure/postgres/repository_displayname_integration_test.go` | T5: drive-by update |
