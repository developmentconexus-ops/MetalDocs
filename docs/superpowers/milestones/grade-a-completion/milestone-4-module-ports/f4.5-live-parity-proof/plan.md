# Feature F4.5 — Plan

> **Milestone:** 4 — Module boundaries / systemic ports  ·  **Feature:** `f4.5-live-parity-proof`
> **Input:** `spec.md` (approved 2026-06-16).

## Plan

1. **F4.1 wire-value unit test**
   - File: `internal/modules/documents/application/version_status_test.go` (new)
   - Test: `TestVersionStatusPublished_WireValue` — asserts `string(templatesdomain.VersionStatusPublished) == "published"`
   - No DB, no build tag. Runs in standard `go test ./...`.
   - Verify: `go test -count=1 -run TestVersionStatusPublished_WireValue ./internal/modules/documents/application/` → PASS.

2. **F4.2 + F4.3 live parity integration tests**
   - File: `internal/modules/security/infrastructure/postgres/repository_port_parity_integration_test.go` (new)
   - Build tag: `//go:build integration`
   - Reuses `openLiveSecurityDB` and `asArray` from the existing `repository_displayname_integration_test.go` (same package `postgres_test`)
   - Seeds: `tenants`, `iam_users` (with `mfa_enabled`), `iam_user_roles`, `audit_events`
   - `F4.2_ListOffHoursAdminActions_port_parity`: seeds one off-hours + one in-hours event for admin user; one off-hours for non-admin; one off-hours for wrong-tenant admin. Asserts exactly 1 result (right actor, right role, in-hours excluded, non-admin excluded, wrong-tenant excluded).
   - `F4.3_MfaCoverage_port_parity`: 2 tenantA users (1 mfa=true, 1 mfa=false). Asserts TotalUsers=2, MfaEnabled=1, MfaEnabledPct=50, per-role slices match seed.
   - Verify compile: `go vet -tags integration ./internal/modules/security/infrastructure/postgres/` → exit 0.

3. **Update F4.1/F4.2/F4.3 evidence.md**
   - F4.1 G2: replace "confirmed by read" with named test command.
   - F4.2 G5/G6: add live parity test command.
   - F4.3 G5/G6: add live parity test command.

## Files touched

| File | Action |
|------|--------|
| `internal/modules/documents/application/version_status_test.go` | new |
| `internal/modules/security/infrastructure/postgres/repository_port_parity_integration_test.go` | new |
| `docs/.../f4.1-published-constant/evidence.md` | update G2 |
| `docs/.../f4.2-iam-role-port/evidence.md` | add G6 |
| `docs/.../f4.3-mfa-coverage-port/evidence.md` | add G6 |
| `docs/.../f4.5-live-parity-proof/spec.md` | pre-existing |
| `docs/.../f4.5-live-parity-proof/plan.md` | this file |
| `docs/.../f4.5-live-parity-proof/evidence.md` | new |

## No production code changes

This feature adds tests only. All production paths (`service.go`, `repository.go`, port impls) are unchanged.
