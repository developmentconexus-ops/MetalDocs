# F5.7 — close proof gaps (plan)

> Executes `spec.md`. Engine: `superpowers:subagent-driven-development` — fresh subagent per task,
> sonnet implement+review; main session reviews diffs + commits + re-dispatches the validator.
> All 4 tasks touch **disjoint files** but share one git index, so they are dispatched **sequentially**
> (each commits before the next) to avoid index races — not parallel.

## Load-bearing facts (verified 2026-07-04 by scoping investigator + ADR read)

1. **testdb template builder:** `tests/integration/testdb/db.go` → `ApplyCuratedBootstrap(ctx, db)` at
   ~line 256 applies `db/prerequisites/0001_extensions.sql`, `db/baseline/0001_current_schema.sql`,
   `db/reference-data/0001_product_reference_data.sql`, then every file in `db/migrations`. **It never
   runs River's migrator.** Add River provisioning **after** the curated bundles.
2. **Production River provisioning to mirror:** `internal/platform/bootstrap/jobs.go:83` —
   `MigrateRiverSchema(ctx, db, schema)` does
   `rivermigrate.New(riverdatabasesql.New(db), &rivermigrate.Config{Schema: schema})` then
   `migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)`. Imports:
   `github.com/riverqueue/river/rivermigrate`, `github.com/riverqueue/river/riverdriver/riverdatabasesql`
   (both already in `go.mod` at v0.37.1). Local prod schema is empty/default → River tables in `public`.
   Simplest correct call in testdb: reuse `bootstrap.MigrateRiverSchema(ctx, db, "")` directly if import
   is clean, else inline the same 2 calls. Prefer reusing the bootstrap function (single source of truth).
3. **4 dispatch tests:** `internal/modules/render/fanout/dispatchjobs/dispatch_integration_test.go`,
   helper `openDispatchDB(t)` (line ~49) = `testdb.Open(t)` + `SET search_path TO metaldocs, public`.
   Tests at lines ~126/201/277/315. They need `river_job` present. Once T1 provisions River in the
   template, they should pass unchanged (do NOT modify the tests unless an assertion is genuinely wrong).
4. **Watchdog fixture:** `internal/modules/jobs/stuck_instance_watchdog/job_integration_test.go`,
   `seedActiveStageSnapshot` (~line 131) sets `on_eligibility_drift_snapshot = $2` = `'auto_cancel'` for
   the AutoCancel variant. CHECK `approval_stage_instances_on_eligibility_drift_snapshot_check` allows
   only `{reduce_quorum, fail_stage, keep_snapshot}` (`db/baseline/0001_current_schema.sql`). The watchdog
   auto-cancel path under test is driven by the watchdog's own policy, NOT this column — so this column
   just needs any valid value. Change `'auto_cancel'` → a valid enum (read the test to pick the one that
   keeps every assertion true; `keep_snapshot` or `fail_stage` most likely neutral).
5. **Audit fixture:** `internal/modules/jobs/audit_integrity_validator/job_integration_test.go`, tamper
   `UPDATE metaldocs.audit_events SET row_hash = 'tampered0000…' WHERE id=$1` (~line 57). CHECK
   `audit_events_row_hash_shape` = `row_hash = '' OR row_hash ~ '^[0-9a-f]{64}$'` (NOT VALID → writes
   still checked). Change the literal to a valid 64-hex wrong hash, e.g.
   `'0000000000000000000000000000000000000000000000000000000000000000'` (64 zeros) — well-formed, ≠ the
   real hash, so still detected as tamper. Confirm the test asserts detection (ErrIntegrityViolation).
6. **F5.1 doc-truth:** ADR 0067 line 3 erratum + §H-PRE-1 (line 122+) — "**H-PRE-1 stays LIVE
   (retirement withdrawn)**"; it governs the **audit-writer** `pg_advisory_xact_lock`, NOT the watchdog
   lock M5 removed. Correct `f5.1-gate-and-adr/evidence.md:32` ("H-PRE-1 … RETIRED") and
   `f5.1-gate-and-adr/spec.md:15,22,52` to LIVE, citing the ADR 0067 §H-PRE-1 erratum.

## Task breakdown (sequential)

### T1 (sonnet) — River-provision the testdb template + green the 4 dispatch proofs
- Edit `tests/integration/testdb/db.go`: after `ApplyCuratedBootstrap`'s curated bundles apply
  successfully, run River's up-migration on the same `*sql.DB` (reuse `bootstrap.MigrateRiverSchema(ctx,
  db, "")` if the import graph is clean — check for an import cycle testdb↔bootstrap; if a cycle exists,
  inline the two `rivermigrate` calls locally in the testdb package instead). Provision **once per
  template**, in the same place the curated bundles run, so every per-test clone inherits `river_job`.
- Run `go test -tags=integration -run Integration ./internal/modules/render/fanout/dispatchjobs/... -v`
  → all 4 PASS. If any test asserts something genuinely wrong (not schema-absence), STOP and report
  (do not weaken assertions to force green).
- `go build ./...` + `go build -tags=integration ./...` + `go vet -tags=integration ./...` clean.
- Commit: `test(harness): F5.7 T1 — provision River schema in testdb template; green dispatchjobs integration proofs`.

### T2 (sonnet) — fix watchdog AutoCancel fixture
- Edit `job_integration_test.go` `seedActiveStageSnapshot` (or the caller) so the AutoCancel variant's
  `on_eligibility_drift_snapshot` is a CHECK-valid value; keep the test's assertions intact.
- `go test -tags=integration -run TestIntegration_Watchdog_P1 ./internal/modules/jobs/stuck_instance_watchdog/... -v` → PASS.
- Commit: `test(jobs): F5.7 T2 — watchdog fixture uses valid on_eligibility_drift_snapshot enum`.

### T3 (sonnet) — fix audit tamper fixture
- Edit `job_integration_test.go` tamper literal → valid 64-hex wrong hash. Keep detection assertion.
- `go test -tags=integration -run TestIntegration_AuditValidator_P3 ./internal/modules/jobs/audit_integrity_validator/... -v` → PASS.
- Commit: `test(jobs): F5.7 T3 — audit tamper uses shape-valid wrong hash so 23514 no longer masks detection`.

### T4 (main session, doc-only) — F5.1 H-PRE-1 wording → LIVE
- Main session edits `f5.1-gate-and-adr/spec.md` + `evidence.md` per fact 6; cite ADR 0067 §H-PRE-1 erratum.
- Commit: `docs(milestone): F5.7 T4 — F5.1 H-PRE-1 wording corrected to LIVE (ADR 0067 erratum)`.

## Test strategy
Targeted `-run` per fixed proof only (no full suite). T1 is the only structural change (shared harness);
T2/T3 are one-literal fixture corrections; T4 is docs. After all four land, main session re-runs the
full set of previously-red proofs green, writes `evidence.md`, and re-dispatches the milestone-validator.

## Files touched (census)
MODIFY `tests/integration/testdb/db.go` (T1). MODIFY
`internal/modules/jobs/stuck_instance_watchdog/job_integration_test.go` (T2). MODIFY
`internal/modules/jobs/audit_integrity_validator/job_integration_test.go` (T3). MODIFY
`f5.1-gate-and-adr/spec.md` + `evidence.md` (T4). No production `.go` under `internal/modules/**`
application/domain/infra changes; no migration/DDL changes.
