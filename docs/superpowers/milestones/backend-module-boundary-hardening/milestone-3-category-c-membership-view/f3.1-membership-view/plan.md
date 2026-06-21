# Feature F3.1 — Plan

> Input: `spec.md` (approved pre-code). Output of `superpowers:writing-plans` run inline.

## Plan

### Files touched
- **NEW** `db/migrations/0242_iam_v_active_user_areas_view.sql` — the view DDL + `schema_migrations` row.
- **NEW** `internal/modules/iam/infrastructure/postgres/active_user_areas_view_parity_integration_test.go` —
  the view-vs-base parity test (`-tags integration`).
- **EDIT** `wiki/decisions/0039-cross-module-base-table-read-boundary.md` — annotate "Related code" with the
  concrete migration path (one line; no decision change).

### Task order (TDD — RED first)
1. Write `active_user_areas_view_parity_integration_test.go`:
   - `testdb.Open(t)` (applies baseline + `db/migrations`).
   - Seed via `testdb.SeedWithCaps(…, [{"cap":"membership.manage"}], …)` direct `INSERT INTO
     public.user_process_areas`: (a) active row (role `qms_admin`, `effective_to NULL`); (b) a 2nd active row
     same user+area different role (`approver`) — multi-role; (c) a revoked row (past `effective_from`, past
     `effective_to > effective_from`, `revoked_by` set) — must be excluded; (d) a wrong-tenant active row.
   - Assert `SELECT tenant_id,user_id,area_code,role FROM metaldocs.v_active_user_areas WHERE tenant_id = $1
     ORDER BY area_code, role` equals the same projection from `public.user_process_areas WHERE effective_to
     IS NULL AND tenant_id = $1`. Also assert the revoked row's role is absent and the active roles present.
2. Run RED: `go test -tags integration ./internal/modules/iam/infrastructure/postgres/ -run ActiveUserAreasView`
   → fails (relation `metaldocs.v_active_user_areas` does not exist).
3. Add `0242_iam_v_active_user_areas_view.sql` (BEGIN/COMMIT, `CREATE VIEW … WHERE effective_to IS NULL`,
   `COMMENT ON VIEW`, `schema_migrations` insert, `ON CONFLICT DO NOTHING`). Match the existing
   `metaldocs.user_process_areas` view's security posture (no `security_invoker`).
4. Run GREEN: same test → ok. Then `go test -tags integration ./tests/integration/testdb/...` (bootstrap still
   applies cleanly with the new migration).
5. `go build ./...`; `go run ./tools/cilint ./...` (exit 0, no ledger/guard change).
6. Annotate ADR-0039 "Related code". Write `evidence.md`. Commit.

### Test strategy
- Parity = **set-equality** (ORDER BY both sides, compare row slices) — not a count. Revoked-exclusion +
  wrong-tenant-exclusion + multi-role inclusion are the authz-drift guards.
- Real PG :5434 only; HS-3 if down.

### Notes / risks
- The seeded revoked row must satisfy `revoked_by_required_when_revoked` (set `revoked_by`) and
  `effective_interval_valid` (`effective_to > effective_from`).
- `area_code` from `testdb.NewTaxonomy` is lowercase (`pa-…`); direct INSERT bypasses the grant-fn regex (as
  the `authz_effective_from` analog does) — fine, the table has no such CHECK.
