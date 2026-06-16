# Feature F4.5 — Live parity proof (HS-4 fix for F4.1/F4.2/F4.3)

> **Milestone:** 4 — Module boundaries / systemic ports  ·  **Feature:** `f4.5-live-parity-proof`
> **Status:** Approved 2026-06-16 — HS-4 fix feature; code change may begin.
> **Opened by:** HS-4 (validator FAIL — F4.1/F4.2/F4.3 acceptance gates were spec-stated tests that did not exist).

## What this feature adds

Three tests that were required by F4.1/F4.2/F4.3 spec gates but never written:

1. **F4.1 wire-value invariant** (unit test, no DB): `string(templatesdomain.VersionStatusPublished) == "published"`.
   Proves the constant chosen in the fix resolves to the correct wire value.

2. **F4.2 live parity** (`//go:build integration`): seeds `iam_user_roles` + `audit_events`; calls
   `repo.ListOffHoursAdminActions` with the port-backed repository; asserts the correct
   `OffHoursAction` set (right actor + right role) and that non-admin actors and wrong-tenant
   actors are excluded.

3. **F4.3 live parity** (`//go:build integration`): seeds `iam_users` + `iam_user_roles`; calls
   `repo.MfaCoverage` with the port-backed repository; asserts total, mfaEnabled, and per-role
   slice values match the known seeded state.

## Non-goals

- No logic change in production code.
- No migration, no new endpoint, no FE change.
- F4.4 is already closed (no-code decision); not touched.

## Validation Gate

1. `go test -count=1 ./internal/modules/documents/application/...` includes and passes
   `TestVersionStatusPublished_WireValue`.
2. `go test -count=1 -tags integration -run TestSecurityRepository_PortParity_Live ./internal/modules/security/infrastructure/postgres/` skips when `DATABASE_URL`/`METALDOCS_DATABASE_URL` not set, but compiles and passes when a real DB is available.
3. `go build ./...` clean.
4. `go test -count=1 ./...` (non-integration) green — no new test failures.
