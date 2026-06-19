# Plan — F5.7 role-admin `tenant_id` upsert

## Files touched
- `internal/modules/iam/infrastructure/postgres/role_admin_repository.go` — two `iam_users` upserts.
- `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go` — update 2 existing
  exact-match expectations; assert 3-arg INSERT carrying `testTenant`.

## Tasks (TDD order)
1. **Red:** update `TestUpsertUserAndAssignRole_PassesTenantID` and
   `TestReplaceUserRoles_DeleteThenInsert_PersistsSingleRole` so the `iam_users` INSERT expectation
   uses `WithArgs("alice", "Alice", testTenant)`. Run → fails (HEAD passes only `"alice","Alice"`).
2. **Green:** in both `UpsertUserAndAssignRoleTx` and `ReplaceUserRolesTx`, change the upsert to:
   ```sql
   INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id, is_active, updated_at)
   VALUES ($1, $2, $3::uuid, TRUE, NOW())
   ON CONFLICT (user_id)
   DO UPDATE SET display_name = EXCLUDED.display_name,
                 tenant_id    = EXCLUDED.tenant_id,
                 is_active    = TRUE,          -- (Replace variant keeps its existing set-list, +tenant_id; no is_active)
                 updated_at   = NOW()
   ```
   and pass `tenantID` as the third exec arg.
3. **Verify:** `go build ./...`; `go test -count=1 ./internal/modules/iam/... ./internal/modules/auth/...`.
4. **Grep gate:** confirm no tenant-less `iam_users` column list remains in `internal/` (non-test).

## Test strategy
sqlmock exact-text + arg matching (same harness as existing tests). The 3rd arg `testTenant` is the
behavioral proof; the SQL column-list change is what the arg assertion enforces.
