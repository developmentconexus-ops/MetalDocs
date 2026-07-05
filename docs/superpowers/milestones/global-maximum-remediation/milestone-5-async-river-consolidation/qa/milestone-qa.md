# Milestone 5 — Validation Verdict (C1–C7)

> **Verdict history (newest last):**
> - **Run 1 — 2026-07-04 — FAIL** (spawned fix-features F5.7 + F5.8). Recorded verbatim below.
> - **Run 2 — 2026-07-04 (re-validation) — PASS.** Re-judged after F5.7 (proof-gap closure) + F5.8
>   (watchdog alert-only, ADR 0068) cleared every Run-1 finding. See the **"Run 2"** section at the
>   bottom of this file. The prior FAIL is preserved unedited as the audit trail.

---

# Run 1 (2026-07-04) — FAIL — history, preserved unedited

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` + `../validation-contract.md` + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-04  ·  **Verdict:** see C7 — **FAIL**.
> The validator judges and writes this file only; it edited no source and flipped no status.

## Inputs loaded

All present and readable: `milestone.md`, `validation-contract.md`, program `README.md`, governing
`mission.md` (§7 M5), F5.1–F5.6 `spec.md`/`evidence.md` (F5.1 is a decision feature with no `plan.md` by
design; F5.6 an operator-approved HS-6 fix). Aggregate diff `cd2bceb3..ed0ef52c` reviewed. Live DB
(`metaldocs-postgres` :5433) reachable — so the deferred integration proofs were **executed by the
validator**, not accepted on their SKIP transcript.

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F5.1 gate+ADR | ✅ | ✅ | ✅ | Gate Yellow (cd2bceb3), ADR 0067 Accepted + indexed (5eb270c3), cited by F5.x. **Note:** F5.1 `spec.md` interview ("H-PRE-1 disposition? Retire") + `evidence.md` item 2 ("H-PRE-1 RETIRED") retain the **withdrawn** framing; authoritative sources (ADR 0067, contract §2.6, invariant-checklist:58, F5.2 spec/evidence) are all corrected to **LIVE**. Lagging feature-doc, not a live runtime split-brain — finding, not blocker. |
| F5.2 janitors | ⚠ code ✅ / proof ❌ | **❌** | ✅ | Census clean, migration applied, both binaries build. But 2 of 3 equivalence proofs **fail on first real run** (see C2). Acceptance criteria 1 (3 equivalence tests) and 2 (§2.6 singleton) were **deferred, never executed** at close; validator ran them — P4 singleton PASS, but P1 watchdog auto-cancel + P3 audit tamper FAIL. |
| F5.3 staging | ⚠ code ✅ / proof ❌ | **❌** | ✅ | Backoff/claim census = 0, poll worker deleted, build green. But **all 4** dispatch integration proofs **fail on first real run** (`river_job` absent in testdb — C2). Acceptance criterion 1 not met. |
| F5.4 retention | ✅ | ✅ | ✅ | Config set (24h/24h/7×24h), purge job wired to `maintenance`; 3/3 integration proofs **PASS for real** (C2). |
| F5.5 fanout | ✅ | ✅ | ✅ | Race test 3/3 subtests **PASS for real**, both interleavings commutative + redelivery no-op (C2). |
| F5.6 queue-fix | ✅ | ✅ | ✅ | Enqueuer→`temporal` unit-asserted; live drive materialized `document.published` for admin; job 23 `temporal/completed/attempt 1` confirmed by validator against live `river_job`. |

**C1 = FAIL** — F5.2 and F5.3 do not meet their declared "what to validate" (named integration proofs
that pass). Their evidence marked the proofs "authored; run deferred," and each states verbatim "A
failure here is an HS-4 (validator FAIL)." On execution, they fail.

## C2 — Gates re-run, isolated

Build/vet/census re-run by validator; **integration proofs executed against live Postgres** (DSN derived
from the running container's non-secret env, not from `.env`).

| Gate | Command | Real output | Pass? |
|------|---------|-------------|-------|
| Build (4 binaries) | `go build ./...` | `BUILD_EXIT=0` | ✅ |
| Vet | `go vet ./...` | `VET_EXIT=0` | ✅ |
| Scheduler pkg gone | `ls internal/modules/jobs/scheduler` | No such file | ✅ |
| Lease-ref census | grep `acquire_lease\|heartbeat_lease\|release_lease\|job_leases\|pg_try_advisory_lock` *.go | 2 hits, both **comments** (outbox consumer TODO re staging `claimLease`; singleton_test.go comment) — 0 functional | ✅ |
| api scheduler census | grep `registerScheduledJobs\|jobscheduler\|StagingOutboxWorker\|ClaimPending\|ResetStaleClaims\|startOutboxWorkers` | 0 matches | ✅ |
| Backoff census (render/fanout) | grep `1<<` | 0 matches | ✅ |
| DB lease objects dropped | `pg_proc` + `to_regclass('metaldocs.job_leases')` | functions 0 rows; table NULL — migration 0273 applied | ✅ |
| River queues live | `SELECT DISTINCT queue FROM river_job` | temporal, maintenance, default | ✅ |
| **F5.2 P4 singleton** | `-run TestSingleton_P4_HPRE1_ExactlyOnceAcrossTwoClients ./internal/platform/jobs/river/` | `--- PASS (97.02s)` | ✅ |
| **F5.2 P2 idempotency (×2)** | `-run TestIntegration_Janitor_P2 ./.../idempotency_janitor/` | both `PASS` (deleted=4; orphan survives) | ✅ |
| **F5.2 P1 watchdog auto-cancel** | `-run TestIntegration_Watchdog_P1 ./.../stuck_instance_watchdog/` | `--- FAIL` — `seedActiveStageSnapshot: violates check constraint "approval_stage_instances_on_eligibility_drift_snapshot_check" (SQLSTATE 23514)` (AlertOnly variant PASS) | **❌** |
| **F5.2 P3 audit tamper** | `-run TestIntegration_AuditValidator_P3 ./.../audit_integrity_validator/` | `--- FAIL` — `tamper row: violates check constraint "audit_events_row_hash_shape" (SQLSTATE 23514)` (CleanChain variant PASS) | **❌** |
| **F5.3 dispatch (×4)** | `-run Integration ./internal/modules/render/fanout/dispatchjobs/` | **all 4 FAIL** — `relation "river_job" does not exist (SQLSTATE 42P01)` (testdb never provisions River schema) | **❌** |
| **F5.4 retention (×3)** | `-run TestRetentionPurge_Integration ./.../retention/` | all `PASS` (purge-only-old both tables; worker both tables; batch-bounded 6/4) | ✅ |
| **F5.5 fanout race (×3)** | `-run TestNotificationsFanoutWorker_ConcurrentRaceCommutativity ./.../notifications/infrastructure/` | all `PASS` (order_A, order_B, redelivery no-op) | ✅ |
| Unit — touched pkgs | `go test ./.../notifications/infrastructure/... ./.../approval/jobs/... ./.../render/fanout/... ./.../jobs/...` | all `ok` | ✅ |
| Live fanout ground truth | `SELECT ... FROM river_job WHERE kind='notification_fanout'` | job 23 temporal/completed/1; job 12 temporal/completed/7 (self-healed); job 6 default/available (orphan) | ✅ |

**C2 = FAIL.** 6 of the 12 deferred integration assertions do not run green on first real execution:
F5.2 watchdog auto-cancel-equivalence, F5.2 audit tamper-detection, and all 4 F5.3 dispatch proofs. The
evidence transcripts recorded these as SKIP ("run deferred to M5 close") and the close live drive (F5.6)
only exercised the fanout-publish path — the F5.2 and F5.3 proofs were **never actually executed** before
this validation. Two are broken fixtures (bad seed vs a real CHECK constraint); four are a harness gap
(the testdb factory does not run `rivermigrate`, so `river_job` is absent for any InsertTx-in-tx proof).
Either way, the named acceptance gates for F5.2 §2.1/§2.3 and F5.3 §3.3 are **not** green.

## C3 — Senior review of the aggregate milestone diff

- **Code-wise the migration is sound and senior-level.** One primitive (River) genuinely replaces three:
  scheduler package deleted, lease functions + table dropped by forward-only migration 0273 (verified live),
  duplicated backoff math removed (census 0), staging poll loop removed, dual-define leader-parity pattern
  applied consistently (janitors, purge, retention). No split-brain of runtime truth; no dead scheduler left
  beside River (the anti-shadow clause §2.5 holds). F5.6's T3 set-based `INSERT...SELECT` rewrite is a genuine
  global-max improvement over the buffer-then-insert local max, and eliminates the open-cursor bug class.
- **Findings (non-blocking):** (a) F5.1 spec/evidence retain the withdrawn "H-PRE-1 retired" wording while
  every authoritative doc says LIVE — lagging feature-doc note. (b) `StagingOutboxWorkerConfig` vestigial
  fields left unread (self-disclosed, harmless).
- **Blocking gap surfaced only in aggregate:** the milestone's *proof layer* is uneven — F5.4/F5.5 proofs are
  real and green, but F5.2/F5.3 proofs were shipped unexecuted and are red on first run. A staff engineer
  would not approve closing a "behavior-preserving migration" milestone whose behavior-preservation proofs
  for two of the migrated jobs (watchdog, audit) and the entire staging-dispatch path do not pass.
- Staff-engineer bar met? **❌** (code diff yes; proof completeness no).

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Async/idempotency + tenancy QA subset | pass-with-failures | Transactional-outbox preserved (InsertTx-in-tx), `ON CONFLICT` dedup kept, `SeedTxTenant` retained on migrated jobs — but the proofs that *demonstrate* watchdog/audit equivalence and staging dispatch idempotency/tenant-seed are red (C2). |
| Regression vs M0–M4 | all still pass | `go build ./...` + `go vet ./...` clean across all 4 binaries; touched-pkg unit suites green; no route/openapi/capability-registry change; live `river_job` shows scheduled-publish/fanout still on `temporal`. No prior-milestone gate regressed. |

**C4 = FAIL** on the async proof subset (root-cause deferral, below); regression itself is clean.

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Dim-5 DEBT: "3 parallel job infras" | 3 (River + lease scheduler + staging poller) | **1** (River) at the **root** | Scheduler pkg deleted; lease fns+table dropped (live-verified); poll worker deleted; census 0. Root-caused, not shadowed. ✅ |
| Unbounded outbox growth | unbounded | bounded | F5.4 purge job + retention config; **3/3 proofs green for real**. ✅ |
| Unverified fanout ordering | unverified | proven commutative | F5.5 race test **3/3 green for real**. ✅ |
| Behavior-preservation of migrated janitors/dispatch | asserted | **NOT proven** | F5.2 watchdog auto-cancel + audit tamper proofs and all F5.3 dispatch proofs **fail on first run**. The consolidation's *structure* is at root; its *equivalence guarantee* for watchdog/audit/staging is unproven. ⚠ |

- Could it be built better? The testdb factory should provision the River schema via `rivermigrate` so
  InsertTx-in-tx proofs (F5.3, and any future River-job test) can run — right now that whole test class is
  structurally un-runnable. That is next-milestone/fix-feature input, and it directly blocks C2 here.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence
- [x] **Fixture/authored proof passed off as satisfied** — F5.2 (watchdog, audit) and F5.3 (×4) proofs were
      recorded "authored; run deferred" and treated as closing the feature, but had **never been executed**;
      on execution they are red. Authored-but-unrun ≠ green.
- [ ] Consumer contract guessed rather than read from the consumer
- [ ] Split-brain (one fact, two sources of truth) — F5.1 stale wording is a lagging doc, corrected
      everywhere authoritative; not a live runtime split-brain.
- [ ] Self-judged close / validator edited or fixed code — validator only judged + wrote this file.
- [ ] Scope drift — F5.6 is recorded with rationale (HS-6, operator-approved); no undocumented scope.
- [ ] Symptom-patch — consolidation is root-level.

**C6 = one hit** (fixture/authored proof treated as satisfied).

## C7 — Verdict

- **VERDICT: FAIL**
- **Failed checks:** C1 (F5.2/F5.3 acceptance not met), C2 (6/12 integration proofs red on isolated re-run),
  C3 (proof layer incomplete), C4 (async proof subset), C6 (authored-but-unexecuted proofs treated as green).
- **Minimum fix feature to open:** `f5.7-close-proof-gaps` — must, from a River-provisioned testdb:
  1. Fix the testdb factory (or the F5.3 tests) so `river_job` exists for InsertTx-in-tx proofs, then run
     the 4 `dispatchjobs` integration tests **green**.
  2. Fix the F5.2 watchdog `AutoCancelEquivalence` fixture seed to satisfy
     `approval_stage_instances_on_eligibility_drift_snapshot_check`, run **green**.
  3. Fix the F5.2 audit `DetectsTamperedChain` fixture seed to satisfy `audit_events_row_hash_shape`, run
     **green**.
  4. Correct the F5.1 `spec.md`/`evidence.md` residual "H-PRE-1 retired" wording to LIVE (doc-truth).
  (F5.2 P4 singleton, F5.4 ×3, F5.5 ×3 already verified green by this validator — no rework needed there.)
- Milestone stays **active**; the main session does **not** advance and does **not** flip status.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): **not reached** — validator FAIL precedes the operator gate.
> - Status flipped in `README.md`: **no** (FAIL).

---

# Run 2 (2026-07-04, re-validation) — PASS

> **Written by:** the `milestone-validator` subagent — a **fresh** dispatch, *not* the main session
> (separation of powers). This run re-judges M5 after fix-features **F5.7** (close-proof-gaps) and
> **F5.8** (watchdog alert-only, ADR 0068) closed the Run-1 findings (HS-4). It re-runs every Run-1 red
> assertion **from clean state** and re-reviews the aggregate diff since the Run-1 FAIL.
> **Validates against:** `../milestone.md` + `../validation-contract.md` + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> The validator judges and writes this file only; it edited no source, fixed no finding, and flipped no
> status.

## Inputs loaded (all present + readable)

`milestone.md` (F5.1–F5.8 feature table, validation-definition, HS list), `validation-contract.md`
(§2–§5), program `README.md` (M0–M4 passed; M5 in-progress), governing `mission.md` (§7 M5), and every
feature's `spec.md`/`plan.md`/`evidence.md` for F5.1–F5.8 (F5.1 is a decision feature — no `plan.md` by
design). The Run-1 FAIL verdict (preserved above) was loaded as the baseline. Aggregate re-validation
diff `git diff 566328ca~1..HEAD` reviewed (18 files, +634/−245). Real Postgres reachable via the repo
`.env` PG* credentials, assembled in-memory per command; **the DSN/secret was never printed, echoed,
logged, or written to any file** — only `ok`/`PASS`/`FAIL` summaries are recorded here.

## What was re-judged (the fixes since Run 1)

- **F5.7** (test/harness/doc only, **no production consolidation change**): T1 provisions the River schema
  in the testdb template via the same `bootstrap.MigrateRiverSchema` production uses (single source of
  truth) — unblocking the whole River-backed integration-test class; a folded fixture fix populates
  `river.Job.JobRow` (a field real River always sets); T3 corrects the audit tamper literal to a
  shape-valid-but-wrong hash; T4 aligns F5.1's lagging "H-PRE-1 retired" wording to LIVE. Run-1's
  watchdog "fixture" was NOT swapped — it was surfaced as a genuine production defect (HS-2/HS-6) and
  routed to F5.8.
- **F5.8** (production removal + ADR 0068): removed the dead `auto_cancel` branch, `SystemCancelInstance`,
  and the `system` authz-bypass path; watchdog is now alert-only; user-facing `CancelInstance`
  (`document.edit`-gated) shared body preserved byte-for-byte; F5.2's auto-cancel-equivalence fiction
  replaced by an honest alert-only integration proof; `wiki/modules/approval.md` corrected.

## C1 — Spec & plan conformance (per feature)

Each feature has `spec.md` (approval line + interview record), `plan.md` (execution-shaped), and
`evidence.md` (acceptance table). F5.1 decision feature: no `plan.md` by design (C1 note allows equivalent
inline output). All acceptance tables map to the milestone/contract gates. Non-goals respected; every
deviation (F5.6, F5.7 T2→F5.8) carries a written rationale.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence (Run 2) |
|---------|----------------------------|-----------------|----------------------|------------------|
| F5.1 gate+ADR | ✅ | ✅ | ✅ | Gate Yellow (cd2bceb3), ADR 0067 Accepted+indexed, cited by F5.x. **Run-1 finding CLEARED:** F5.1 `spec.md:15,54` + `evidence.md:33,40` now state **H-PRE-1 LIVE** (retirement withdrawn, ADR 0067 §H-PRE-1 erratum), interview line annotated as revisited. No live doc asserts retirement (grep confirms only corrected/annotated occurrences). |
| F5.2 janitors | ✅ | ✅ | ✅ | Census clean, migration applied, both binaries build. **Run-1 proof gaps CLEARED:** watchdog proof now honest alert-only (F5.8, C2 green); audit tamper P3 green (F5.7 T3, C2); P4 singleton + P2 idempotency still green (regression). |
| F5.3 staging | ✅ | ✅ | ✅ | Backoff/claim census 0. **Run-1 gap CLEARED:** all 4 `dispatchjobs` proofs green on the River-provisioned testdb (F5.7 T1, C2). |
| F5.4 retention | ✅ | ✅ | ✅ | 3/3 retention proofs re-run green under the new harness (C2) — additive, no regression. |
| F5.5 fanout | ✅ | ✅ | ✅ | Race commutativity proof re-run green under the new harness (C2). |
| F5.6 queue-fix | ✅ | ✅ | ✅ | `temporal` queue fix + set-based fanout; live fanout confirmed Run 1; unchanged since. |
| F5.7 proof-gaps | ✅ | ✅ | ✅ | Consumers = the named F5.2/F5.3 proofs + doc-truth; all now green/correct (C2). Non-goal "no production consolidation change" honored — diff is test/harness/doc only; the watchdog production defect was surfaced to F5.8, not folded in. |
| F5.8 watchdog | ✅ | ✅ | ✅ | Consumers: jobs runtime (new `NewWorker` sig — builds), validator honest proof (green), approval public cancel surface (regression green). Non-goals honored (no `on_timeout_action`, no `CancelInstance`/route/authz/OpenAPI/migration change). ADR 0068 Accepted+indexed; wiki alert-only. |

**C1 = PASS.** Every feature's declared "what to validate" is met and each consumer contract honored;
non-goals respected; deviations carry written rationale. The Run-1 C1 fails (F5.2/F5.3 acceptance) and the
F5.1 doc-truth finding are all cleared.

## C2 — Gates re-run, isolated (validator ran these, from clean state)

All integration proofs run against real Postgres (repo `.env` PG* creds, assembled in-memory; secret
never surfaced). Fixture-vs-real is labeled; all integration proofs below are **real Postgres**, not
mock/fixture.

| Gate | Command (validator-run) | Real output | Pass? |
|------|-------------------------|-------------|-------|
| Build (all Go pkgs) | `go build ./...` | `BUILD_EXIT=0` | ✅ |
| 3 Go hosts build | `go build -o /dev/null ./apps/{api,worker,jobs}/cmd/*` | api/worker/jobs all OK (docx-renderer is not a Go binary — no `apps/**/*.go`; `go build ./...` covers all Go) | ✅ |
| Vet | `go vet ./...` | clean (no unused field/param/import from the F5.8 removal) | ✅ |
| **Census** `auto_cancel` / `SystemCancelInstance` / `stuck_watchdog_auto_cancel` (prod+test) | `grep -rn … internal/ apps/ --include=*.go` | **0** hits | ✅ |
| Scheduler pkg gone | `ls internal/modules/jobs/scheduler` | No such file | ✅ |
| Lease primitive census | grep `acquire_lease\|heartbeat_lease\|release_lease\|job_leases\|pg_try_advisory_lock` (non-test) | 1 hit — a pre-existing `heartbeat_lease` **comment** (outbox consumer TODO), 0 functional; unchanged by M5 | ✅ |
| H-PRE-1 doc-truth (T4) | grep `H-PRE-1 .* retir/RETIRED` in M5 docs | every occurrence is the corrected "stays LIVE / retirement withdrawn" or an erratum annotation — no live retirement claim | ✅ |
| **F5.3 dispatch (×4)** — *real PG* | `go test -tags=integration -run Integration ./internal/modules/render/fanout/dispatchjobs/... -v` | **4/4 PASS** — PDF (77.55s), Materialize (4.21s), Enqueue-PDF-tx (4.79s), Dedup-skip (4.40s) | ✅ |
| **F5.7 T3 audit tamper P3 (×2)** — *real PG* | `go test -tags=integration -run TestIntegration_AuditValidator_P3 ./internal/modules/jobs/audit_integrity_validator/... -v` | **2/2 PASS** — `_DetectsTamperedChain` (120.87s, `issue_count=2 first_kind=row_hash_mismatch` — genuine detection, not masked) + `_CleanChainNoIssue` (0.91s, `issue_count=0`) | ✅ |
| **F5.8 watchdog alert-only (×2)** — *real PG* | `go test -tags=integration -run TestIntegration_Watchdog ./internal/modules/jobs/stuck_instance_watchdog/... -v` | **2/2 PASS** — `_AlertOnly` (156.03s) + `_AlertOnly_AnyDriftPolicy` (7.69s); test body seeds valid drift (`reduce_quorum`/`keep_snapshot`), asserts stuck instance stays `in_progress` + doc `under_review` + exactly one `stuck_alert`, fresh <7d instance untouched (0 alerts). No `auto_cancel` seed, no cancel assertion — honest. | ✅ |
| **F5.8 user-cancel regression** — *real PG* | `go test -tags=integration -run Cancel ./internal/modules/documents/approval/...` | all `ok` (application/domain/http/contracts) — `CancelInstance` behaviorally unchanged | ✅ |
| F5.8 unit (watchdog + cancel_service) | `go test ./…/stuck_instance_watchdog/... ./…/approval/application/...` | both `ok` | ✅ |
| **F5.4 retention (regression)** — *real PG* | `go test -tags=integration -run TestRetentionPurge_Integration ./…/render/fanout/...` | `ok …/retention 163.34s` — still green under new harness | ✅ |
| **F5.5 fanout race (regression)** — *real PG* | `go test -tags=integration -run ConcurrentRaceCommutativity ./…/notifications/...` | `ok …/notifications/infrastructure 163.65s` — still green under new harness | ✅ |
| F5.8 gate6 bypass surface | grep `authz.BypassSystem(` in approval (non-test) | actual call sites reduced to 1 in approval/application (`scheduler_service.go:65`); the removed system-cancel path dropped the surface by one — net −1 as intended | ✅ |
| ADR 0068 + wiki | `head` ADR 0068 + `wiki/decisions/index.md` + `wiki/modules/approval.md` | ADR 0068 **Accepted 2026-07-04**, indexed row present; approval.md C4 node + chokepoint + failure-mode all alert-only, cite ADR 0068 | ✅ |

**C2 = PASS.** Every Run-1 red assertion (F5.2 watchdog, F5.2 audit P3, F5.3 ×4 — 6 total) now runs
**green from clean state on real Postgres**, and the previously-green proofs (F5.4 ×N, F5.5) still pass
under the additive River-provisioned harness. No transcript trusted — each command was executed by this
validator.

## C3 — Senior review of the aggregate milestone diff (since Run-1 FAIL, `566328ca~1..HEAD`)

Reviewed the whole re-validation diff (18 files) as one unit:

- **Production removals are clean and senior-level.** `cancel_service.go` collapses
  `CancelInstance`/`SystemCancelInstance`/`cancelInstance(…, system bool)` into a single
  `CancelInstance` with unconditional `authz.Require` — the shared body (GUC set, instance/stage cancel,
  doc→draft OCC, governance event) is preserved byte-for-byte, and the live path always passed
  `system=false`, so behavior is provably identical. `job.go` deletes the dead `auto_cancel` branch,
  `cancelSvcInterface`, and the now-unused `cancelSvc`/`runner` fields; `WithBackgroundBypass` stays
  (still needed for list/alert reads) with a corrected comment; `DriftPolicy` is still read and emitted
  in the alert payload. `main.go` composition root updated to the new `NewWorker` signature; `services.Cancel`
  remains wired for the HTTP handler. **Dead code removed, not shadowed; bypass surface reduced.**
- **Harness change is the global maximum, not a per-test patch.** `testdb/db.go` reuses production's
  `bootstrap.MigrateRiverSchema(ctx, db, "")` — one source of truth, additive (after the curated bundles),
  no baseline/migration/reference bundle altered. The dispatch-test `JobRow` population matches real
  River's always-set field (fixture-class, not a production change).
- **No split-brain, no guessed contract, no feature breaking another.** F5.4/F5.5 re-run green proves the
  harness change is additive. H-PRE-1 doc-truth is now single-valued (LIVE everywhere).
- **Findings (non-blocking):** none new. The Run-1 non-blocking findings (F5.1 wording; vestigial
  `StagingOutboxWorkerConfig` fields) — the wording is now fixed (T4); the vestigial-config note was
  already self-disclosed and outside the F5.7/F5.8 scope (not touched, not a regression).
- Staff-engineer bar met? **✅** — the diff is exactly the two surfaced fixes, tightly scoped, with the
  behavior-preservation of the live cancel path and the harness additivity both proven by re-run.

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Async/idempotency + multi-tenant QA subset (`qa-operating-system.md`) | **pass** | Transactional-outbox preserved (InsertTx-in-tx), `ON CONFLICT` dedup kept, `SeedTxTenant`/`SeedTxIdentity` on migrated jobs — and now the proofs that *demonstrate* watchdog (alert-only), audit tamper-detection, and staging dispatch idempotency/tenant-seed all **run green** (C2). The Run-1 pass-with-failures is resolved. |
| Test-discipline (`test-discipline.md`, ADR 0034) | **pass** | testdb factory for every proof; targeted `-run` only (full suite not run — 20-min box, bounded defer recorded); the harness now provisions River, completing the fixture framework's intent. |
| Regression vs M0–M4 | **all still pass** | `go build ./...` + `go vet ./...` clean; touched-pkg unit + integration suites green; no `openapi.yaml`/route/capability-registry change (F5.8 removed only an internal method + bypass path); F5.4/F5.5 re-run green; lease scheduler stays retired (census). No prior-milestone gate regressed. |
| Root-cause (binds C4) | **at root** | River is the single primitive (scheduler/lease/backoff deleted, census 0); the watchdog defect was removed at source (dead branch + orphan method + bypass path), not masked; the audit/dispatch proofs were fixed by correcting the fixture/harness truth, not by weakening assertions. |

**C4 = PASS.**

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before (Run 1) | After (Run 2) | Root-cause-fixed evidence |
|-------------|----------------|---------------|---------------------------|
| Dim-5 DEBT: "3 parallel job infras" | 3 → 1 (structure at root) | **1** (River), structure + proof both at root | Census 0 (scheduler/lease/backoff); dispatch proofs now **green** proving River actually dispatches. ✅ |
| Unbounded outbox growth | bounded (proven) | bounded, **re-proven green** | F5.4 3/3 re-run green under new harness. ✅ |
| Unverified fanout ordering | proven commutative | **re-proven green** | F5.5 race re-run green. ✅ |
| Behavior-preservation of migrated janitors/dispatch | **NOT proven** (Run-1 blocker) | **PROVEN** | Watchdog alert-only proof green (honest, real PG); audit tamper P3 green (genuine detection); all 4 dispatch proofs green. The Run-1 gap is closed. ✅ |
| Watchdog auto-cancel (Run-1 "fixture" red) | fictional equivalence test on a schema-impossible seed | **dead branch removed** (ADR 0068), alert-only proven | Root-caused as an orphaned pre-schema concept (not introduced by M5); removed at source, not fixture-swapped to green. Census 0; honest proof. ✅ |

- **Root-cause vs symptom-patch:** confirmed root-cause on both surfaced defects. The watchdog was **not**
  patched by swapping the fixture to a passing value (which would have masked the dead branch) — the
  HS-2/HS-6 "surface, don't fold" clause fired, the operator chose collapse-to-alert-only, and the orphan
  was deleted. The dispatch/audit reds were harness/fixture truth-corrections (River schema provisioned;
  shape-valid wrong hash), with assertions intact (tamper still detected).
- **Could it be built better?** The Run-1 retrospective item (testdb factory should provision River) is
  now **done** (F5.7 T1) — the whole River-backed integration-test class is permanently unblocked. Future
  input only: a real `on_timeout_action` config if auto-cancel is ever genuinely wanted (recorded as a
  bounded defer in F5.8/ADR 0068 §Future). Not a defect in the current construction.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-level "all green" reported as a pass without per-feature acceptance mapped to evidence — **no**; each feature's named proof was run individually and mapped (C2).
- [ ] Fixture/mock passed off as real-provider proof — **no**; every integration proof ran on **real Postgres** (labeled), and the Run-1 fixture-as-satisfied hit is cleared (proofs now actually executed green).
- [ ] Consumer contract guessed rather than read from the consumer — **no**; F5.7/F5.8 consumers (named proofs, jobs runtime, approval public surface) read and honored.
- [ ] Split-brain (one fact, two sources of truth) — **no**; H-PRE-1 doc-truth is now single-valued LIVE everywhere (T4).
- [ ] Self-judged close / validator edited or fixed code — **no**; this validator only ran proofs, judged, and wrote this verdict file. No source/spec/doc edited; no status flipped.
- [ ] Scope drift — **no**; F5.6/F5.7/F5.8 are each recorded with rationale (HS-6/HS-4/HS-2), inside the async-consolidation cluster; no undocumented scope.
- [ ] Symptom-patch — **no**; both surfaced defects fixed at root (dead-code removed; harness/fixture truth corrected without weakening assertions).

**C6 = clean (no hit).** The single Run-1 hit (authored-but-unexecuted proof treated as green) is cleared.

## C7 — Verdict

- **VERDICT: PASS**
- **All checks pass:** C1 (per-feature conformance, all consumer contracts honored, Run-1 doc-truth
  finding cleared), C2 (all 6 Run-1 red proofs now green from clean state on real Postgres; regression
  proofs re-run green), C3 (aggregate diff is senior-level, tightly scoped, behavior-preservation +
  harness additivity both proven), C4 (async/tenancy QA + M0–M4 regression clean; root-cause), C5
  (quality bar re-measured and now fully proven, not symptom-patched), C6 (forbidden-list clean).
- Every Run-1 failing check (C1/C2/C3/C4/C6) is cleared by F5.7 (proof/harness/doc) + F5.8 (production
  dead-code removal, ADR 0068). No new finding.
- Milestone M5 is validated. Handed back to the main session to flip status and present the **HS-1
  operator gate** (which also carries the two pending ratifications recorded in the program README: the
  ADR 0067 §H-PRE-1 erratum and the ADR 0066 M4 split — operator-facing, not validator-facing).

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): **pending** — now reached (validator PASS); operator ratifies the M5 close +
>   the carried ADR 0067 §H-PRE-1 erratum.
> - Status flipped in `README.md`: **yes, on PASS** — main session action, not the validator's.
> - Never push without explicit operator approval (mission §2/§10).
