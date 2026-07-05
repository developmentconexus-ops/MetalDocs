# Feature F5.8 — Evidence — watchdog alert-only collapse (ADR 0068)

> **Milestone:** 5 · **Feature:** `f5.8-watchdog-alert-only` · **Closed:** 2026-07-04
> **Origin:** F5.7 proof run surfaced a failing watchdog equivalence test. Three-investigator
> read-only audit (2026-07-04) proved it was **not** a fixture bug but a genuine, pre-existing
> production defect — an orphaned dead-code branch from a concept (`auto_cancel` stuck-timeout action)
> abandoned before it ever reached schema. Per F5.7's HS-2/HS-6 clause the defect was surfaced, not
> folded in; the operator selected the global-maximum fix (AskUserQuestion 2026-07-04: **"Collapse to
> alert-only, remove orphans"**). **Rails:** ADR 0068.
> **Contract:** `spec.md`. Validation Gate proved below.
> Engine: `superpowers:subagent-driven-development` — sequential sonnet subagent per task (shared git
> index; T1 changes signatures T2 depends on); main session reviewed diffs, committed, and drove the
> census/bypass-surface gates itself.

## Root cause (audited, not assumed)

`on_eligibility_drift[_snapshot]` fused two orthogonal concepts under one name. Concept **A**
(eligibility-drift quorum policy `{reduce_quorum, fail_stage, keep_snapshot}`) won — it is the only
CHECK-allowed value set and the only thing signoff evaluates. Concept **B** (a stuck-timeout action
`{auto_cancel, alert_only, none}`) was abandoned before schema. The frontend was corrected to A
(`14e48071`); the Go watchdog was never synced, so it read `on_eligibility_drift_snapshot` and compared
it to `"auto_cancel"` — a value the column can never hold (schema CHECK rejects it). The `auto_cancel`
branch, `SystemCancelInstance`, and the `system` authz-bypass path it invoked were all unreachable.
**Not introduced by M5** (origin phase-8 `84b0507f`); M5 F5.2 migrated the read source but preserved the
orphan. Full detail + evidence anchors in ADR 0068.

## What was implemented

Three commits (+ one census-cleanup follow-up):

- **T1 `64be3145` — production removal.**
  - `internal/modules/jobs/stuck_instance_watchdog/job.go`: deleted the dead
    `if inst.DriftPolicy == "auto_cancel" { … SystemCancelInstance … }` branch, the `autoCancelled`
    counter/log field, the `cancelSvcInterface` type, and the `cancelSvc`+`runner` struct fields.
    `NewWorker(database *sql.DB, emitter governanceEmitter)` and `run(ctx, database, emitter)` lost the
    cancel-service params. Kept `authz.WithBackgroundBypass` (comment updated to alert-only rationale —
    still needed for `listStuckInstances`/`emitStuckAlert`'s own `BypassSystem` reads), kept
    `StuckInstance.DriftPolicy` (read + emitted as informational context in the alert payload) and the
    `on_eligibility_drift_snapshot` SELECT.
  - `internal/modules/documents/approval/application/cancel_service.go`: deleted `SystemCancelInstance`;
    folded `cancelInstance(…, system bool)` into `CancelInstance` (dropped the `system` param, deleted
    the `if system { authz.BypassSystem + authz.SeedTxIdentity }` block, made
    `authz.Require(ctx, tx, CapDocumentEdit, areaCode.String)` unconditional). Shared body (GUC set,
    instance/stage cancel, doc→draft OCC, governance event) byte-for-byte unchanged — the live path
    always passed `system=false`, so behavior is identical.
  - `apps/jobs/cmd/metaldocs-jobs/main.go`: composition root updated to
    `stuck_instance_watchdog.NewWorker(db, approvalEmitter)` (dropped the `services.Cancel` arg;
    `services.Cancel` still wired for the HTTP cancel handler).
  - `cancel_service_test.go`: removed only `TestSystemCancelInstance_BypassesUserCapability`.

- **T2 `d46d607f` — honest alert-only tests.**
  - `job_test.go`: removed `mockCancelService`, updated `run(...)` call sites to the 2-arg signature,
    deleted `TestWatchdog_AutoCancel` + `TestWatchdog_CancelError`; kept `TestWatchdog_NoStuck`,
    `TestWatchdog_AlertOnly`, and the read-source invariant `TestListStuckInstances_UsesStageSnapshotDriftPolicy`;
    added `TestWatchdog_AlertOnly_MultipleInstances`.
  - `job_integration_test.go`: removed the `..._AutoCancelEquivalence` fiction (it seeded the
    schema-impossible `"auto_cancel"`); renamed the honest test to `TestIntegration_Watchdog_P1_AlertOnly`
    (extended to prove a fresh <7d instance is untouched and zero alerts fire for it); added
    `TestIntegration_Watchdog_P1_AlertOnly_AnyDriftPolicy` (valid `keep_snapshot`).

- **T3 `3942b4d1` — wiki alert-only.** `wiki/modules/approval.md`: C4 System/Rel nodes (78, 87),
  the stuck-instance failure-mode row (535), the `CancelInstance` chokepoint desc + anchor refresh to
  `cancel_service.go:46`, and a new **Last verified** stamp — all corrected to alert-only, citing ADR 0068.

- **`0080d195` — grep-census cleanup.** Reworded 4 test comments that still spelled the defunct
  `auto_cancel` / `SystemCancelInstance` tokens so the census gate (item 3) is literally zero; meaning
  preserved via ADR 0068 references.

## Verification

| Gate (spec.md Validation Gate) | Command / action | Result |
|---|---|---|
| 1. Build (all 4 binaries, new `NewWorker` sig) | `go build ./...` | **clean** |
| 2. Vet | `go vet ./...` + `go vet -tags=integration ./…/stuck_instance_watchdog/… ./…/approval/application/…` | **clean** |
| 3. Grep census = 0 (prod **and** test) | `grep -rn "auto_cancel\|SystemCancelInstance\|stuck_watchdog_auto_cancel" internal/ apps/ --include=*.go` | **CENSUS=0 CLEAN** (post `0080d195`) |
| 4. Honest alert-only integration proof (real Postgres, targeted `-run`) | `go test -tags=integration -run TestIntegration_Watchdog ./internal/modules/jobs/stuck_instance_watchdog/... -v` | **PASS** — T2 subagent run on real Postgres (`.env` DB), ~132s; stuck instance stays `in_progress` + doc `under_review` + exactly one `approval.instance.stuck_alert`; fresh <7d instance untouched. Independently re-run by milestone-validator. |
| 5. User-cancel path regression | `go test -tags=integration -run Cancel ./internal/modules/documents/approval/...` | live path unchanged (T1 fold preserved the shared body; `CancelInstance` still `document.edit`-gated). Re-run by validator. |
| 6. Bypass-surface check | `grep -rn "BypassSystem" internal/modules/documents/approval/ --include=*.go \| grep -v _test.go` | **2 callers remain** (`scheduler_service.go:65`, `scheduled_publish_job.go:56`); the removed system-cancel path dropped the approval module's `BypassSystem` surface by one. Watchdog still roots in `WithBackgroundBypass` for its list/alert reads. |
| 7. ADR 0068 Accepted + indexed; wiki alert-only | `wiki/decisions/0068-*.md` + `wiki/decisions/index.md` row; `wiki/modules/approval.md` | **done** (T3 `3942b4d1`) |

### Live drive (M5-close, real running system — the F5.8 binary in `metaldocs-jobs`)

`.\scripts\start-api.ps1 -Build` rebuilt all 3 Go hosts (api/worker/**jobs**) from current source
(census-0, so the jobs binary is the F5.8 build) and launched them; `.\scripts\check-system-runnable.ps1`
**PASS** (login/session/auth-me/target-route all 200). River DB ground truth (`public.river_job` /
`river_leader`):

- Leader elected `MN-NTB-LEANDROTH_2` @ `2026-07-05 00:22:45Z` — advisory-lock-free election (ADR 0067), jobs host connected.
- **`stuck_instance_watchdog` fresh tick `id=30` — `state=completed, attempt=1, finalized 2026-07-05 00:27:50Z`** — i.e. produced *after* this rebuild's leader election, by the F5.8 alert-only binary. It **completed on the first attempt with no error and no cancel path** — the live proof that the alert-only watchdog runs clean in the running system. (Prior ticks id≤29 stop at 22:34Z, from the pre-rebuild binary.)
- Other consolidated River jobs completing on their correct queues: `idempotency_janitor` (maintenance), `materialize_dispatch` (temporal), `notification_fanout` (temporal, F5.6). The lone `notification_fanout|default|available` row is **id=6** (created 21:18Z, `attempt=0`) — the pre-F5.6 orphan already recorded as a bounded defer in F5.6 evidence, not a regression.

### Bypass-surface before/after (spec gate 6)

- **Before:** approval-module `BypassSystem` callers = 3 (`cancel_service.go` system path + `scheduler_service.go` + `scheduled_publish_job.go`).
- **After:** 2 (`scheduler_service.go:65`, `scheduled_publish_job.go:56`). Net **−1** authz-bypass surface, as intended by ADR 0068 §Decision.

## Acceptance vs spec Validation Gate

| Criterion (spec.md) | Met? | Evidence |
|---|---|---|
| 1. Build green, jobs comp-root compiles new sig | **yes** | `go build ./...` clean; `metaldocs-jobs/main.go` updated (T1) |
| 2. Vet clean, no unused field/param/import | **yes** | `go vet` clean (plain + `-tags=integration`) |
| 3. Grep census = 0 (prod + test) | **yes** | CENSUS=0 CLEAN after `0080d195` |
| 4. Honest alert-only integration proof passes on real Postgres | **yes** | T2 run, real `.env` DB, ~132s PASS; no schema-bypassing mock |
| 5. User-cancel path unchanged | **yes** | shared body byte-for-byte; `system` was always false for `CancelInstance`; regression `-run Cancel` |
| 6. `BypassSystem` approval caller count drops by one | **yes** | 3 → 2, before/after above |
| 7. ADR 0068 Accepted + indexed; wiki alert-only | **yes** | T3 `3942b4d1`, ADR indexed |

## Review disposition

- T1 production diff read in full by the main session before accepting (`git show 64be3145`): confirmed
  the live cancel path is behaviorally identical (the removed `system` branch was never reached by
  `CancelInstance`), no shared-body change, `authz` import still used (`Require`, `WithCapCache`).
- T2 tests reviewed: no test seeds `"auto_cancel"` (a CHECK violation); the alert-only proof asserts
  real behavior on real Postgres, not a mock.
- **HS-2/HS-6 discipline honored end to end:** the defect was surfaced to the operator (not fixture-swapped
  to green), the operator chose the global-maximum path, and F5.7 T2 was formally superseded rather than
  silently repurposed.

## Bounded defers

| Defer | Why bounded | Trigger / owner |
|---|---|---|
| A real auto-cancel (`on_timeout_action`) feature | Explicitly out of scope (ADR 0068 §Future) — needs a successor ADR + new column + snapshot + OpenAPI + frontend. Alert-only is the safe default for a regulated doc system (human-in-the-loop). | New operator-approved feature if/when auto-cancel is genuinely wanted. |
| `-race` on the watchdog integration suite | Env-capped (`CGO_ENABLED=0` on this box) — same constraint as F5.5/F5.6 | Re-run under a cgo-capable CI; not blocking. |
