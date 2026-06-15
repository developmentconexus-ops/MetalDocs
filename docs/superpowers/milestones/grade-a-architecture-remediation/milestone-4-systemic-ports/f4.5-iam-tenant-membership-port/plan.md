# Feature F4.5 — Plan

> TDD. Producer-only feature (consumer is F4.6). Two new Go files + one ADR. No existing file changes
> except ADR registration/cross-links. No adjacent refactor (CLAUDE.md §5.3).

## Slices (ordered)

1. **RED — unit test.** `internal/modules/iam/domain/tenant_user_reader_port_test.go`:
   `TestNoopTenantUserReader` asserts `NoopTenantUserReader{}.TenantUserIDs(ctx,"t")` returns an empty
   non-nil slice + nil error. Compiles-red (no port type yet).

2. **GREEN — domain port.** `internal/modules/iam/domain/tenant_user_reader_port.go`:
   `TenantUserReader` interface (`TenantUserIDs(ctx, tenantID) ([]string, error)`) + `NoopTenantUserReader`
   null-object, doc comments (membership = an `iam_users` (user_id,tenant_id) row; all members, no
   `deactivated_at` filter; pool/off-tx, H-PRE-1).

3. **GREEN — pool-backed impl.** `internal/modules/iam/infrastructure/postgres/tenant_user_repository.go`:
   `TenantUserRepository{db *sql.DB}`, `NewTenantUserRepository(db)`,
   `var _ iamdomain.TenantUserReader = (*TenantUserRepository)(nil)`,
   `TenantUserIDs` → `SELECT user_id FROM metaldocs.iam_users WHERE tenant_id = $1::uuid` (empty slice
   when none, nil error).

4. **REAL — live-PG integration.** `…/postgres/tenant_user_repository_integration_test.go`
   (`//go:build integration`): seed tenant A with two members (one with `deactivated_at` set) + tenant
   B with one member; assert `TenantUserIDs(A)` returns both A members (proves no active-only filter),
   excludes B's member (tenant scope), and an unknown tenant returns empty. Mirror the F4.1
   `user_display_name_repository` live-test harness (`TEST_DATABASE_URL`, real seeded tenant ids,
   cleanup).

5. **ADR 0031.** `wiki/decisions/0031-tenant-user-reader-port.md`; register in
   `wiki/decisions/index.md`; supersede the deferred-membership note in ADR 0029.

6. **PROVE + CLOSE.** `go build ./...`; `go vet` (plain + integration) on iam; run unit + live tests;
   write `evidence.md`; commit.

## Files touched

- `internal/modules/iam/domain/tenant_user_reader_port.go` (new)
- `internal/modules/iam/domain/tenant_user_reader_port_test.go` (new)
- `internal/modules/iam/infrastructure/postgres/tenant_user_repository.go` (new)
- `internal/modules/iam/infrastructure/postgres/tenant_user_repository_integration_test.go` (new)
- `wiki/decisions/0031-tenant-user-reader-port.md` (new) + `wiki/decisions/index.md` (one row)
- `wiki/decisions/0029-user-display-name-reader-port.md` (supersede deferred-membership note)
- this feature's `spec.md` / `plan.md` / `evidence.md`
