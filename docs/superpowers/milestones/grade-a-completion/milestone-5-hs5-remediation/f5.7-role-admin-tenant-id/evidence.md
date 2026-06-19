# Evidence — F5.7 role-admin `tenant_id` upsert (re-audit Major #2 — FIXED)

> **Status:** CLOSED 2026-06-19 · **Major #2 confirmed real and fixed.** The `iam_users` upsert in
> both role-admin write paths now carries the caller's `tenant_id`; new rows no longer default to the
> sentinel placeholder tenant.

## Disposition

Re-audit 2026-06-16 Major #2 was a **true defect**. `metaldocs.iam_users.tenant_id` is
`NOT NULL DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff'` (`0001_current_schema.sql:1423`); both
production upserts (`UpsertUserAndAssignRoleTx`, `ReplaceUserRolesTx`) omitted `tenant_id` from the
INSERT, so a newly-created user landed under the **sentinel tenant** instead of the caller's real one
— invisible to its tenant's `UNIQUE (tenant_id,user_id)` reads and the ADR-0031 `TenantUserReader`
membership/display-name resolution. Callers already had the real tenant
(`auth/application/service.go:213` `tenant.DevTenantID`, `:615` `fields.tenantID`); it was dropped only
at this one INSERT. The sibling `iam_user_roles` write was already correct.

## Change (behavior fix)

| File | Change |
|------|--------|
| `internal/modules/iam/infrastructure/postgres/role_admin_repository.go` | Both upserts: column list `(user_id, display_name, tenant_id, is_active, updated_at)`, `VALUES ($1,$2,$3::uuid,…)`, `tenantID` passed as 3rd arg, `ON CONFLICT … DO UPDATE SET … tenant_id = EXCLUDED.tenant_id …` (repairs sentinel rows on next upsert). Anchor comment cites Major #2 / F5.7. |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go` | 2 exact-match expectations updated to the 3-arg INSERT carrying `testTenant`. |
| `internal/modules/auth/application/service_test.go:992` | Rollback test's `iam_users` expectation updated to the 3-arg INSERT carrying `tenantID`. |

## TDD record (failing-first)

1. **Red:** updated `TestUpsertUserAndAssignRole_PassesTenantID` /
   `TestReplaceUserRoles_DeleteThenInsert_PersistsSingleRole` to expect
   `(user_id, display_name, tenant_id, is_active, updated_at)` with `WithArgs("alice","Alice",testTenant)`.
   Ran against unchanged source →
   `FAIL: could not match actual sql "INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, updated_at) …"`
   and `arguments do not match: expected 2, but got 3`. The drop was observable.
2. **Green:** added `tenant_id` to both upserts (column list + `$3::uuid` value + `EXCLUDED.tenant_id`
   on conflict). Tests pass.
3. A third sqlmock expectation (`service_test.go:992`, `TestCreateUser_RollbackWhenReplaceUserRolesFails`)
   also bound the old 2-arg INSERT and was updated to the 3-arg form — caught by the suite, then fixed.

## Validation Gate results (real output)

1. **TDD** — red→green captured above (`expected 2, but got 3 arguments` → all green).
2. **Build** — `go build ./...` → exit 0.
3. **Tests** — `go test -count=1 ./internal/modules/iam/... ./internal/modules/auth/...` → all packages
   `ok` (iam: application, authz, delivery/http, domain, infra/memory, infra/postgres, presence; auth:
   application, delivery/http, domain, infra/memory, infra/postgres).
4. **Grep gate** — `grep "iam_users (user_id, display_name, is_active, updated_at)" internal/ (non-test)`
   → `GATE CLEAN: tenant-less iam_users upsert eliminated`. (`e2e_seed.go:509` already carried
   `tenant_id`.)

## Defers

None. Historical sentinel-tenant rows (if any pre-date this fix) are repaired lazily by
`EXCLUDED.tenant_id` on the next upsert of that user; a one-shot backfill was explicitly scoped out
(spec Non-goals) and is **not** required for the fix's correctness — recorded as a note, not a defer.
