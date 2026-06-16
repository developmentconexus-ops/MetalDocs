# Feature F4.3 — Plan

> **Spec:** `spec.md` (approved 2026-06-16).

## Tasks

### T1 — New IAM port interface + Noop
**File:** `internal/modules/iam/domain/mfa_user_reader_port.go` (new)

`RoleMfaCounts` struct. `MfaUserReader` interface with `TenantMfaCounts` and `TenantMfaCountsByRole`.
`NoopMfaUserReader` returning zeros / empty slice.

### T2 — Postgres implementation
**File:** `internal/modules/iam/infrastructure/postgres/mfa_user_repository.go` (new)

`MfaUserRepository`:
- `TenantMfaCounts` → single `QueryRowContext` against `iam_users`.
- `TenantMfaCountsByRole` → `QueryContext` against `iam_user_roles JOIN iam_users`.
Compile-time guard: `var _ iamdomain.MfaUserReader = (*MfaUserRepository)(nil)`.

### T3 — Security repository: add field + update NewRepository + rewrite MfaCoverage
**File:** `internal/modules/security/infrastructure/postgres/repository.go`

- Add `mfaUsers iamdomain.MfaUserReader` to `Repository` struct.
- Add 5th arg to `NewRepository`; nil → `NoopMfaUserReader{}`.
- Rewrite `MfaCoverage`: call `r.mfaUsers.TenantMfaCounts` + `TenantMfaCountsByRole`, compute
  percentages, assemble `securitydomain.MfaCoverage` (logic unchanged; SQL reads removed).

### T4 — Wire in main.go
**File:** `apps/api/cmd/metaldocs-api/main.go`

Pass `iampg.NewMfaUserRepository(sqlDB)` as 5th arg to `securitypg.NewRepository`.

### T5 — Update integration test (drive-by)
**File:** `internal/modules/security/infrastructure/postgres/repository_displayname_integration_test.go`

Pass `iampg.NewMfaUserRepository(db)` as 5th arg.

## Files touched

| File | Change |
|------|--------|
| `internal/modules/iam/domain/mfa_user_reader_port.go` | T1: new interface + Noop |
| `internal/modules/iam/infrastructure/postgres/mfa_user_repository.go` | T2: new impl |
| `internal/modules/security/infrastructure/postgres/repository.go` | T3: struct + NewRepository + MfaCoverage rewrite |
| `apps/api/cmd/metaldocs-api/main.go` | T4: wiring |
| `internal/modules/security/infrastructure/postgres/repository_displayname_integration_test.go` | T5: drive-by update |
