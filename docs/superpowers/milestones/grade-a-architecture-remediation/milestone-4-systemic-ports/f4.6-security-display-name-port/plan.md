# Feature F4.6 — Plan

> The "how" for [`spec.md`](spec.md). Consumer contract approved pre-code.

## Steps

1. **RED** — add live-PG integration test
   `internal/modules/security/infrastructure/postgres/repository_displayname_integration_test.go`
   (`//go:build integration`, package `postgres_test`): self-seed two tenants + members (one
   deactivated, one other-tenant), `auth_identities` lockout/failed-login rows, `auth_sessions`
   new-device rows. Construct `NewRepository(db, displayNameRepo, tenantUserRepo)` (real ports).
   Assert per method: tenant isolation, membership filter, `missing→user_id` fallback, anti-join.
   Confirm RED (compile fail: `NewRepository` arity) before GREEN.
2. **GREEN** — edit `repository.go`:
   - struct + `NewRepository(db, displayNames, members)`; import `iamdomain`.
   - private helper `resolveNames(ctx, tenantID, ids []string) (map[string]string, error)` →
     `DisplayNames` + `id→id` fallback for missing/empty.
   - `ListLockouts`: SQL drops JOIN + display_name col; `WHERE i.user_id = ANY($1)` (member ids);
     post-scan enrich via helper.
   - `CountRecentFailedLoginsByUser`: same JOIN→`ANY` swap; enrich.
   - `CountRecentLockouts`: drop JOIN; `WHERE i.user_id = ANY($1)`.
   - `ListNewDeviceLogins`: drop JOIN; keep `s.tenant_id`; enrich.
   - member-id fetch: call `members.TenantUserIDs(ctx, tenantID)` once per coupled method; empty set →
     short-circuit (`ANY('{}')` is safe but skip the round-trip when no members).
3. **Wire** — `apps/api/cmd/metaldocs-api/main.go`: pass the two pool-backed iam repos into
   `securitypg.NewRepository`.
4. **PROVE + CLOSE** — `go build ./...`; `go vet` (plain + integration) on security + api; run security
   unit tests (interface unchanged → green) + the new live integration tests (verify RUN, not skipped);
   grep proofs (0 `iam_users.display_name`, only `MfaCoverage` JOIN remains); add security consumer to
   ADR 0029 Key files; write `evidence.md`; commit.

## Risk / rollback
- Behavior-identical by construction (same rows, same names, same order/limit). Risk = an off-by-one in
  the member-id filter vs JOIN; mitigated by the live test asserting exact membership incl. deactivated
  + other-tenant exclusion.
- Rollback = revert the single repository.go + main.go diff; ports are additive.
