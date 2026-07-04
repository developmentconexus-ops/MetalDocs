# F5.7 — close proof gaps (spec)

> **Milestone:** M5 · **Status:** in progress
> **Origin:** milestone-validator FAIL verdict (`../qa/milestone-qa.md`, 2026-07-04) — HS-4. The
> validator had live DB access and **executed** the F5.2/F5.3 integration proofs that F5.2/F5.3
> evidence had recorded as "authored; run deferred to M5 close". 6 of 12 deferred assertions are red
> on first real run. Each affected feature's evidence states verbatim "a failure here is an HS-4
> (validator FAIL)". This feature makes those named acceptance gates actually pass.
> **Nature:** proof-layer + test-harness + doc-truth fixes. **No production consolidation code changes**
> — the validator confirmed the migration itself is sound at root (C3/C5). We are closing the gap
> between "authored proof" and "green proof", not re-doing M5.

## Consumer contract

The **consumers** are the named acceptance proofs of F5.2 and F5.3, plus the doc-truth invariant. Each
must move from red/lagging to green/correct:

1. **F5.3 §3.3 — the 4 `dispatchjobs` integration proofs** (`internal/modules/render/fanout/dispatchjobs/dispatch_integration_test.go`): `TestPDFDispatchWorker_Integration_...`, `TestMaterializeDispatchWorker_Integration_...`, `TestEnqueuer_Integration_EnqueuePDFTx_InsertsOutboxRowAndRiverJob`, `TestEnqueuer_Integration_DedupSkip_NoSecondRiverInsert`. They call `testdb.Open(t)` and assert on `river_job` rows, but the testdb template never provisions the River schema → `relation "river_job" does not exist (42P01)`. **Contract:** these run **green** against a testdb whose template includes the River schema.
2. **F5.2 §2.1 — watchdog `TestIntegration_Watchdog_P1_AutoCancelEquivalence`** (`internal/modules/jobs/stuck_instance_watchdog/job_integration_test.go`): its `seedActiveStageSnapshot` writes `on_eligibility_drift_snapshot = 'auto_cancel'`, which is not in the column's CHECK set `{reduce_quorum, fail_stage, keep_snapshot}` → `23514`. **Contract:** the fixture seeds a **constraint-valid** value while still exercising the auto-cancel watchdog path the test asserts.
3. **F5.2 §2.3 — audit `TestIntegration_AuditValidator_P3_DetectsTamperedChain`** (`internal/modules/jobs/audit_integrity_validator/job_integration_test.go`): its tamper `UPDATE ... SET row_hash = 'tampered0000…'` writes a 52-char non-hex value, violating `audit_events_row_hash_shape` (`row_hash = '' OR row_hash ~ '^[0-9a-f]{64}$'`, `NOT VALID` so writes are still checked) → `23514`. **Contract:** the tamper writes a **well-formed-but-wrong** 64-lowercase-hex hash — still a genuine tamper (mismatches the recomputed chain hash → detected), without violating the shape CHECK.
4. **F5.1 doc-truth** (`../f5.1-gate-and-adr/spec.md`, `evidence.md`): residual "H-PRE-1 retired/RETIRED" wording contradicts the authoritative sources — ADR 0067's HS-7 erratum ("**H-PRE-1 stays LIVE**, retirement withdrawn"), contract §2.6, `invariant-checklist.md:58`, F5.2 docs. **Contract:** F5.1's wording aligns to **H-PRE-1 LIVE**, citing the ADR 0067 §H-PRE-1 erratum.

## What to implement

- **T1 (harness, global-max):** provision the River schema in the testdb template. `tests/integration/testdb/db.go`'s `ApplyCuratedBootstrap` applies `db/prerequisites` + `db/baseline` + `db/migrations` but never runs River's migrator. Add a call mirroring production's `bootstrap.MigrateRiverSchema` — i.e. `rivermigrate.New(riverdatabasesql.New(db), &rivermigrate.Config{Schema: ""})` then `Migrate(ctx, DirectionUp, nil)` — after the curated bundles, so the per-test clone has `river_job`, `river_leader`, etc. Use the same empty/default schema production uses locally (River tables live in `public`). This is the correct global-max fix (unblocks the entire River-job integration-test class permanently), not a per-test workaround. Then the 4 dispatch tests must pass.
- **T2 (fixture):** in the watchdog test, change the `AutoCancelEquivalence` seed's `on_eligibility_drift_snapshot` from `'auto_cancel'` to a valid enum value that preserves the test's intent. The watchdog's auto-cancel behavior is **not** encoded by this column (it is the watchdog policy/config under test); this column is the approval-stage eligibility-drift snapshot and just needs a valid value. Pick the valid value consistent with what the test asserts (read the test body; do not weaken any assertion). Test passes green.
- **T3 (fixture):** in the audit test, change the tamper `row_hash` literal to a valid 64-char lowercase-hex string that differs from the real computed hash (e.g. 64×`'0'`, or any `^[0-9a-f]{64}$` value guaranteed ≠ the genuine row hash). Verify the test still asserts tamper **is detected** (the validator recomputes and compares — a wrong-but-well-formed hash still mismatches). Test passes green.
- **T4 (doc):** correct the F5.1 `spec.md` + `evidence.md` H-PRE-1 wording to state H-PRE-1 remains **LIVE** (retirement withdrawn), citing ADR 0067 §H-PRE-1 erratum. Keep the historical interview line but annotate the disposition as corrected/withdrawn so the record shows the decision was revisited, matching how ADR 0067 handled its own erratum in place.

## Non-goals

- **No change to production consolidation code** (River wiring, janitors, purge, enqueuers, fanout worker). The validator already passed C3/C5 on the migration itself. If any proof fix reveals a genuine production defect (not a fixture/harness bug), that is an HS-2/HS-6 stop — surface it, do not fold it in silently.
- **No new janitor/dispatch behavior**; no change to the CHECK constraints themselves (the constraints are correct; the fixtures were wrong).
- **No re-running the full integration suite** (20+ min). Targeted `-run` per fixed test only.
- **No change to F5.4/F5.5** proofs — the validator already ran them green.

## Validation Gate (acceptance — all must hold)

1. **T1:** `go test -tags=integration -run Integration ./internal/modules/render/fanout/dispatchjobs/...` → all 4 **PASS** against a River-provisioned testdb.
2. **T2:** `go test -tags=integration -run TestIntegration_Watchdog_P1 ./internal/modules/jobs/stuck_instance_watchdog/...` → **PASS** (both variants).
3. **T3:** `go test -tags=integration -run TestIntegration_AuditValidator_P3 ./internal/modules/jobs/audit_integrity_validator/...` → **PASS** (both variants).
4. **No regression:** `go build ./...` + `go vet ./...` clean; the already-green proofs (F5.2 P4 singleton, F5.2 P2 idempotency, F5.4 ×3, F5.5 ×3) still pass; T1's harness change provisions River **additively** (no baseline/migration bundle altered, no existing integration test broken).
5. **T4:** no file under M5 (or authoritative wiki) says H-PRE-1 is retired; F5.1 wording cites ADR 0067 §H-PRE-1 (LIVE).
6. **Re-dispatch milestone-validator** → PASS (this feature exists to clear the C1/C2/C3/C4/C6 findings in `../qa/milestone-qa.md`).

## Interview record

No operator interview — the fix is fully determined by the validator's verdict + the investigator's root-cause scoping (testdb harness gap + two constraint-violating fixture literals + one lagging doc). The one design choice (fix the shared testdb factory vs patch 4 tests) is resolved to the global-max: fix the factory, per the Global-Maximum rule and the validator's own C5 recommendation.

## ADR

No new ADR — no durable architectural decision. T1 is a test-harness completeness fix consistent with the existing testdb factory design (ADR 0034 integration-test fixture framework) and with ADR 0067 (River is the async primitive); it makes the harness able to test River-backed code, which the framework always intended.
