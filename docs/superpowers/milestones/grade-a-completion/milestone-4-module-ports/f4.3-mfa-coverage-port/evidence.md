# Feature F4.3 — Evidence

> **Milestone:** 4 — Module boundaries / systemic ports  ·  **Feature:** `f4.3-mfa-coverage-port`  ·  **Closed:** 2026-06-16
> **Contract:** `spec.md`.
> **Retires:** M4 accepted MfaCoverage defer from prior wave (final `iam_users` reach in `security`).

## Changes

### New: `internal/modules/iam/domain/mfa_user_reader_port.go`
`MfaUserReader` interface + `RoleMfaCounts` struct + `NoopMfaUserReader`. Off-tx, H-PRE-1.

### New: `internal/modules/iam/infrastructure/postgres/mfa_user_repository.go`
`MfaUserRepository` — two methods:
- `TenantMfaCounts`: `SELECT COUNT(*) FILTER(...) FROM iam_users WHERE tenant_id=$1`
- `TenantMfaCountsByRole`: `SELECT role_code, COUNT(*) ... FROM iam_user_roles JOIN iam_users ... GROUP BY role_code`
Both pool-read (off-tx, H-PRE-1). Compile-time guard.

### Modified: `security/infrastructure/postgres/repository.go`
- `Repository.mfaUsers iamdomain.MfaUserReader` field added.
- `NewRepository` 5th param; nil → `NoopMfaUserReader{}`.
- `MfaCoverage` rewritten: calls `r.mfaUsers.TenantMfaCounts` + `TenantMfaCountsByRole`; no
  direct SQL against `iam_users` or `iam_user_roles`; percentage computation unchanged.

### Modified: `apps/api/cmd/metaldocs-api/main.go`
`iampg.NewMfaUserRepository(sqlDB)` as 5th arg.

### Modified: integration test (drive-by)
`iampg.NewMfaUserRepository(db)` as 5th arg.

## Grep evidence: `iam_users` / `iam_user_roles` in `security/`

All remaining matches are **comments** or **integration test fixture setup** (seeding `iam_users`
rows to test port-level membership behavior) — zero SQL reads from `security` queries.

Confirmed zero SQL reads by grep:
```
grep -RIn 'FROM metaldocs.iam_users\|FROM metaldocs.iam_user_roles' internal/modules/security/
```
→ 0 matches.

## H-PRE-1 confirmation

`MfaCoverage` is called from `security.Service.MfaCoverage` (application service, plain read path)
— never inside a lock-holding tx. Port calls use pool connection (`r.db`). Off-tx confirmed.

**Prior accepted defer retired:** `MfaCoverage` was the last direct `iam_users` read in `security/`;
this feature closes it.

## Verification

| Gate | Command | Result | Real vs fixture |
|------|---------|--------|-----------------|
| G1: no SQL iam_users/iam_user_roles reads in security queries | `grep -RIn 'FROM metaldocs.iam_users\|FROM metaldocs.iam_user_roles' internal/modules/security/ --include='*.go'` | 0 matches | — |
| G2: interface+struct in iamdomain, impl in iam/postgres, Noop in iamdomain | file presence | confirmed | — |
| G3: H-PRE-1 | caller trace (see above) | off-tx confirmed | — |
| G4: build clean | `go build ./...` | no output (clean) | — |
| G5: tests green | `go test -count=1 ./internal/modules/security/... ./internal/modules/iam/...` | all `ok` | fixture |

## Bounded defers

None. C3 finding fully closed; prior wave defer retired.
