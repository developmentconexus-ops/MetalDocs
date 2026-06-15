# Feature F0.1 — authz `effective_from`

> **Milestone:** 0 — Auth / authz / session correctness  ·  **Folder:** `f0.1-authz-effective-from`
> **Status:** Planning

## Source

- Milestone spec row: *implement* — `authz.Require` honors `effective_from <= now()` (matching
  `ResolveEligibleActors`), at the shared authz layer. *validate* — future-dated membership denied;
  current granted; existing authz tests green.
- Governing-spec reference: mission §5 B1, §7 F0.1.

## Plan

**Files touched**
1. `internal/modules/iam/authz/authz.go` — add one clause to the capability-grant query (lines 115-126).
2. `internal/modules/iam/authz/authz_effective_from_integration_test.go` — **new**, `//go:build integration`,
   `package authz_test` (external) — real-Postgres proof.

**Test strategy (TDD — failing test first)**

The existing `authz_test.go` uses a fake driver returning canned `granted` regardless of SQL, so it
cannot exercise the predicate. The proof must be real-DB. New integration file, mirroring the
`testdb.Open` factory pattern (`tests/integration/testdb`) and the GUC-seeding pattern from
`internal/modules/iam/integration_test.go` (`TestProbeD`):

Shared helper `seedMembership(t, db, effFrom time.Time) (tenantID, userID, area, cap string)`:
- `testdb.NewTenant` → tenantID; `testdb.NewUser(WithTenant)` → userID.
- `testdb.NewTaxonomy(WithTenant)` → gives a real `document_process_areas` code (FK target for `area_code`).
- Insert `public.user_process_areas (user_id, tenant_id, area_code, role, effective_from)` with
  `role='qms_admin'`, `area_code=tax.ProcessAreaCode`, `effective_from=effFrom`, wrapped in
  `testdb.SeedWithCaps(..., '[{"cap":"membership.manage"}]', ...)` (table carries the membership tripwire).
- Capability proven: curated `role_capabilities` already maps `qms_admin → document.submit`
  (`db/reference-data/0001_product_reference_data.sql:52`).

`TestRequire_FutureDatedMembershipDenied` (the RED test):
- `effFrom = now()+1h`. Begin tx; `authz.SeedTxIdentity(ctx, tx, tenantID, userID)`; call
  `authz.Require(authz.WithCapCache(ctx), tx, "document.submit", area)`.
- Assert `errors.As(err, &authz.ErrCapDenied{})`. **Against current code this FAILS** (current `Require`
  ignores `effective_from`, so it grants → returns nil).

`TestRequire_CurrentMembershipGranted` (guards against over-fix):
- `effFrom = now()-1h`. Same call. Assert `err == nil` (grant).

**Implementation (make it green)**

Add `AND upa.effective_from <= now()` to the JOIN-side predicate in the `Require` query:

```
  JOIN metaldocs.user_process_areas upa
    ON upa.role = rc.role
   AND upa.tenant_id = $4::uuid
   AND upa.user_id   = $3
   AND upa.effective_from <= now()      -- F0.1: deny future-dated memberships (matches ResolveEligibleActors)
   AND upa.effective_to IS NULL
```

`effective_to IS NULL` left untouched (Non-goal; see spec interview Q2).

**Ordering**
1. Write both integration tests; run `go test -tags integration ./internal/modules/iam/authz/...` →
   confirm `FutureDated` **fails**, `Current` passes (proves the test is real and the bug exists).
2. Add the predicate clause.
3. Re-run integration → both green.
4. `go test ./internal/modules/iam/authz/...` (unit, no tag) → green (no regression).
5. `go test ./...` → green.

## Execution notes

- Implemented directly (single-line predicate change + one test file) under review discipline — too
  small to warrant subagent-driven-development fan-out. Code-review pass before evidence close.
- Integration tests require a live Postgres (`DATABASE_URL`/`METALDOCS_DATABASE_URL` or the testdb
  harness). If the harness is unavailable in this environment, that is recorded honestly in
  `evidence.md` (the test exists and is correct; real-provider run is the gate).
