# Feature F5.7 — Evidence — close proof gaps

> **Milestone:** 5 · **Feature:** `f5.7-close-proof-gaps` · **Closed:** 2026-07-04
> **Origin:** milestone-validator FAIL verdict (`../qa/milestone-qa.md`, 2026-07-04) — HS-4. The
> validator had live DB access and executed the F5.2/F5.3 integration proofs that their evidence had
> recorded as "authored; run deferred to M5 close". 6 of 12 deferred assertions were red on first real
> run. This feature makes those named acceptance gates actually pass.
> **Nature:** proof-layer + test-harness + fixture + doc-truth fixes. **No production consolidation
> code changed** — the validator confirmed the migration itself sound at root (C3/C5).
> **Contract:** `spec.md`. Validation Gate proved below.
> Engine: `superpowers:subagent-driven-development` — sonnet subagents for the harness/fixture work;
> main session reviewed diffs + committed + drove the census-truth gates. **T2 was REMOVED** (the
> watchdog "fixture" was a genuine production defect — routed to F5.8/ADR 0068 under the HS-2/HS-6
> clause, not folded in).

## What was implemented

- **T1 `be66e96e` — River-provision the testdb template (harness, global-max).**
  `tests/integration/testdb/db.go` — `ApplyCuratedBootstrap` now calls
  `bootstrap.MigrateRiverSchema(ctx, db, "")` **after** the curated bundles, so every per-test clone
  has `river_job`/`river_leader`/`river_queue`/etc. Fixes `relation "river_job" does not exist
  (42P01)`. Additive — no baseline/migration/reference bundle altered. This is the correct global-max
  fix (unblocks the entire River-backed integration-test class permanently), per the Global-Maximum
  rule and the validator's C5 recommendation, not a per-test patch.

- **`9c95d214` — dispatch-worker `JobRow` fixture (fixture; folded into T1's charter).**
  Provisioning River exposed a *second*, independent blocker in 2 of the 4 dispatch proofs: the tests
  built `&river.Job[T]{Args: …}` with a **nil embedded `*rivertype.JobRow`**, so production's
  `job.Attempt`/`job.MaxAttempts` (fields on `rivertype.JobRow`) nil-deref'd. Production
  (`workers.go:126,150`) is correct — real River always populates `JobRow`. This is a fixture gap, in
  F5.7's fixture-fix charter (**not** a production defect → not HS-2/HS-6). Fixed by populating
  `JobRow: &rivertype.JobRow{Attempt: 1, MaxAttempts: 25}` (River's non-terminal first-work state) in
  both `dispatch_integration_test.go` constructions. The first T1 subagent correctly *stopped* rather
  than patch it blind (HS discipline); main session verified it was fixture-class and folded it in.

- **T3 `0871b138` — audit tamper fixture (fixture).**
  `internal/modules/jobs/audit_integrity_validator/job_integration_test.go` — the tamper `UPDATE …
  SET row_hash = 'tampered0000…'` wrote a 52-char non-hex literal that violated
  `audit_events_row_hash_shape` (`^[0-9a-f]{64}$`, `23514`) before the validator could even run.
  Replaced with 64 lowercase-hex zeros — still a genuine tamper (mismatches the recomputed chain hash
  → detected as `row_hash_mismatch`), without violating the shape CHECK.

- **T4 `ca9cb618` (prior commit) — F5.1 H-PRE-1 doc-truth.**
  Corrected residual "H-PRE-1 retired" wording in F5.1 `spec.md`/`evidence.md` (and across F5.2/
  milestone.md) to **H-PRE-1 LIVE**, citing the ADR 0067 §H-PRE-1 erratum. The historical interview
  line is kept but annotated as withdrawn, matching how ADR 0067 recorded its own erratum in place.

- **T2 — REMOVED (superseded by F5.8).** The scoping premise ("swap the watchdog fixture to a valid
  value") was wrong: the audit proved the watchdog IS keyed off `on_eligibility_drift_snapshot` and the
  `auto_cancel` branch is dead code. Correct resolution = the production fix in F5.8 (watchdog →
  alert-only, ADR 0068), not a fixture swap. `spec.md` items 2 + T2 + gate 2 all annotated REMOVED.

## Verification (real Postgres via `.env` DB — value never exposed)

| Gate (spec.md) | Command | Result |
|---|---|---|
| 1. T1 — 4 dispatch proofs green on River-provisioned testdb | `go test -tags=integration -run Integration ./internal/modules/render/fanout/dispatchjobs/... -v` | **all 4 PASS** — `ok metaldocs/…/dispatchjobs 97.680s`: `TestPDFDispatchWorker_Integration_PublishesAndMarksDispatched` (77.55s), `TestMaterializeDispatchWorker_Integration_PublishesAndMarksDispatched` (4.21s), `TestEnqueuer_Integration_EnqueuePDFTx_InsertsOutboxRowAndRiverJob` (4.79s), `TestEnqueuer_Integration_DedupSkip_NoSecondRiverInsert` (4.40s) |
| 2. T2 | — | **REMOVED** — honest alert-only proof is an F5.8 acceptance gate (see `../f5.8-watchdog-alert-only/evidence.md`) |
| 3. T3 — audit P3 both variants green | `go test -tags=integration -run TestIntegration_AuditValidator_P3 ./internal/modules/jobs/audit_integrity_validator/... -v` | **PASS** — `ok …/audit_integrity_validator 129.465s`: `_DetectsTamperedChain` (120.87s, `issue_count=2 first_kind=row_hash_mismatch`) + `_CleanChainNoIssue` (0.91s, `issue_count=0`) |
| 4. No regression | `go build ./...` + `go vet ./...` | **clean**; T1's harness change is additive (River provisioned after curated bundles); the enqueuer proofs (already green pre-fix) still pass |
| 5. T4 — no source says H-PRE-1 retired; F5.1 cites ADR 0067 §H-PRE-1 LIVE | `grep -rn "H-PRE-1 retired\|RETIRED" docs/superpowers/milestones/.../milestone-5-*` | corrected in `ca9cb618` |
| 6. Re-dispatch milestone-validator | milestone Phase 4 | pending — runs at M5 close, this feature exists to clear its C1–C6 findings |

### Note on the `.env` DB access
The integration proofs connect to the local Postgres via the repo `.env` credentials
(`PGHOST/PGPORT/PGDATABASE/PGUSER/PGPASSWORD/PGSSLMODE`, assembled in-memory in the PowerShell
session). The connection string was **never printed, echoed, logged, or written to a file** — only the
`ok`/`PASS` summaries are recorded here.

## Acceptance vs spec Validation Gate

| Criterion (spec.md) | Met? | Evidence |
|---|---|---|
| 1. T1 — 4 dispatch tests PASS on River testdb | **yes** | `ok …/dispatchjobs 97.680s`, 4/4 PASS |
| 2. T2 | **N/A** | REMOVED → F5.8 |
| 3. T3 — audit P3 PASS (both variants) | **yes** | `ok …/audit_integrity_validator 129.465s`, 2/2 PASS |
| 4. No regression; harness change additive | **yes** | build+vet clean; no bundle altered; enqueuer proofs still green |
| 5. T4 — H-PRE-1 LIVE wording | **yes** | `ca9cb618` |
| 6. Validator re-dispatch → PASS | **pending** | M5 close (Phase 4) |

## Review disposition

- All fixture diffs read in full before accepting (`git show be66e96e / 9c95d214 / 0871b138`): each is
  additive/fixture-only, zero production-code change — matching F5.7's stated nature.
- **HS discipline honored:** the T1 subagent stopped on the nil-`JobRow` failure rather than patch it
  blind; the main session verified it was fixture-class (production reads a field River always
  populates) and folded it into F5.7's charter — while the genuinely-production watchdog defect was
  routed out to F5.8, not folded in. Two failures, two correct dispositions.

## Bounded defers

| Defer | Why bounded | Trigger / owner |
|---|---|---|
| Full integration suite (20+ min) not run | Targeted `-run` per fixed proof only, per spec non-goal | Runs in CI / at validator's discretion |
| `-race` on these integration proofs | Env-capped (`CGO_ENABLED=0` on this box) | Re-run under cgo-capable CI |
