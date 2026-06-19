# Plan — F5.2 auth-usertenant-port

## Files

**New:**
- `internal/modules/iam/domain/user_tenant_reader_port.go` — `UserTenantReader` interface +
  `NoopUserTenantReader`. Mirror `tenant_user_reader_port.go`.
- `internal/modules/iam/infrastructure/postgres/user_tenant_repository.go` — `UserTenantRepository`
  + `NewUserTenantRepository(db)`. Pool-backed, off-tx. Query identical to the replaced auth query.
- `internal/modules/iam/infrastructure/postgres/user_tenant_repository_test.go` — white-box unit
  test pinning the query string (no DB).
- `internal/modules/iam/infrastructure/postgres/user_tenant_repository_integration_test.go` —
  `//go:build integration` live parity test. Mirror `tenant_user_repository_integration_test.go`.

**Modified:**
- `internal/modules/auth/infrastructure/postgres/repository.go` — add `iamdomain` import; add
  `userTenants iamdomain.UserTenantReader` field; `NewRepository(db, userTenants)` with nil→Noop;
  `GetUserTenants` delegates to `r.userTenants.UserTenantIDs(ctx, userID)`.
- `internal/platform/bootstrap/api.go:80` — `authpg.NewRepository(db, iampg.NewUserTenantRepository(db))`.
- 6 auth-pg test callers of `NewRepository(db)` → `NewRepository(db, nil)` (Noop; none test
  postgres GetUserTenants).

## Steps (TDD)

1. **Red — unit:** write `user_tenant_repository_test.go` pinning the query; it fails to compile
   (type/constructor absent). That is the red.
2. **Green — port + impl:** add the domain port + Noop, the Postgres repo with the identical query;
   unit test compiles + passes.
3. **Live parity test:** add the integration test (build-tagged). Attempt a live run if a DB is
   reachable; else record skip.
4. **Rewire consumer:** change auth `NewRepository` signature + field + `GetUserTenants` body; wire
   bootstrap; update the 6 test callers to pass `nil`.
5. **Gate:** H-G grep → 0; `go build ./...`; `go test -count=1 ./internal/modules/auth/... ./internal/modules/iam/...`.

## Test strategy

- **Unit** (no DB, always runs): query-shape pin — guards the parity contract in CI without a DB.
- **Integration** (`-tags integration`, needs DATABASE_URL): behavioral parity against live
  `iam_user_roles`. Honestly labeled in evidence as run-or-skipped.
- The replaced query and the new query are character-identical on the SQL that matters
  (DISTINCT/WHERE/ORDER BY), so parity is by construction; the tests are the guard.
