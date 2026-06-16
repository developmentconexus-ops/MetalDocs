# Feature F4.3 — IAM MfaCoverage port

> **Milestone:** 4 — Module boundaries / systemic ports  ·  **Feature:** `f4.3-mfa-coverage-port`
> **Status:** Approved 2026-06-16 — code change may begin.

## Consumer contract

**Consumer:** `security.MfaCoverage` in
`internal/modules/security/infrastructure/postgres/repository.go`.

**What it needs:** total active user count + MFA-enabled count for a tenant, and a per-role
breakdown of the same counts — so it can assemble `securitydomain.MfaCoverage` without
reading `iam_users` or `iam_user_roles` directly. This retires the M4 accepted defer from the
prior wave.

**Required shape after this feature:**
- A new `iamdomain.MfaUserReader` interface (in `internal/modules/iam/domain/`) with:
  - `TenantMfaCounts(ctx, tenantID) (total, mfaEnabled int, err error)` — active-user total + mfa count.
  - `TenantMfaCountsByRole(ctx, tenantID) ([]RoleMfaCounts, error)` — per-role breakdown.
- A `RoleMfaCounts` struct in `iamdomain` (`Role string`, `Total int`, `MfaEnabled int`).
- A Postgres impl `iam.MfaUserRepository` (in `internal/modules/iam/infrastructure/postgres/`)
  that issues the two queries IAM already owns against `iam_users` and `iam_user_roles`.
- A `NoopMfaUserReader` returning zeros and an empty slice.
- `security.Repository` struct gains `mfaUsers iamdomain.MfaUserReader` field.
- `NewRepository` accepts it as a 5th arg (nil → Noop default).
- `MfaCoverage` calls `r.mfaUsers.TenantMfaCounts` + `TenantMfaCountsByRole`, then assembles
  `securitydomain.MfaCoverage` with computed percentages (no direct SQL against iam tables).

## Anchor (re-verified 2026-06-16)

- `internal/modules/security/infrastructure/postgres/repository.go:63` — `func (r *Repository) MfaCoverage(...)`
- `internal/modules/security/infrastructure/postgres/repository.go:67` — `FROM metaldocs.iam_users u`
- `internal/modules/security/infrastructure/postgres/repository.go:83` — `FROM metaldocs.iam_user_roles ur`

## Non-goals

- No new IAM HTTP endpoint.
- No redesign of IAM's port family.
- No change to `securitydomain.Repository` interface or `securitydomain.MfaCoverage` shape.
- No new schema/migration.
- No F4.2 (iam-role-port) re-scope — already committed.

## Validation Gate

1. `grep -RIn 'iam_users\|iam_user_roles' internal/modules/security/ --include='*.go'` →
   0 SQL references (struct comments are named in `evidence.md`); C3 finding closed.
2. `MfaUserReader` interface + `RoleMfaCounts` in `iamdomain`; impl in `iam/postgres`; Noop in `iamdomain`.
3. H-PRE-1: `MfaCoverage` is not called inside a lock-holding tx — confirmed in `evidence.md`.
4. `go build ./...` clean.
5. `go test -count=1 ./internal/modules/security/... ./internal/modules/iam/...` green.
