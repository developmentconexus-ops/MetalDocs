# F5.8 — watchdog alert-only collapse (plan)

> Executes `spec.md`. Engine: `superpowers:subagent-driven-development` — fresh subagent per task,
> sonnet implement+review; main session reviews diffs + commits. Tasks are **sequential** (shared
> git index; T1 changes signatures T2/T3 depend on).

## Load-bearing facts (verified 2026-07-04 by audit + direct read)

1. **Watchdog** `internal/modules/jobs/stuck_instance_watchdog/job.go`:
   - Dead branch `job.go:100–116` (`if inst.DriftPolicy == "auto_cancel"` → `SystemCancelInstance`).
   - `cancelSvcInterface` (33–36), `cancelSvc`+`runner` fields (59, 61), `NewWorker` param (65),
     `run` params (86), `Work` calls `run(ctx, w.database, w.runner, w.cancelSvc, w.emitter)` (80).
   - `runner` used ONLY to feed `SystemCancelInstance`; `listStuckInstances`/`emitStuckAlert` use raw
     `db.BeginTx`. Both become unused after removal.
   - `authz.WithBackgroundBypass` (79) still needed (list/alert `BypassSystem` at 143, 191).
   - `DriftPolicy` still read (153) + emitted in alert payload (207) — keep.
   - Composition-root call site of `NewWorker` — find it (likely `apps/jobs/**` or
     `internal/platform/bootstrap/jobs.go`); update to drop the cancelSvc arg.
2. **CancelService** `internal/modules/documents/approval/application/cancel_service.go`:
   - `CancelInstance` (46–48) → `cancelInstance(…, false)`; `SystemCancelInstance` (50–54) →
     `cancelInstance(…, true)`; shared `cancelInstance(…, system bool)` (56–204).
   - `system`-gated: `if system { BypassSystem + SeedTxIdentity }` (66–80); `if !system { Require }`
     (114–118). Fold into one path: always `Require`, no bypass/seed.
   - Live caller: `http/cancel_handler.go:18` → `CancelInstance` (system=false). Unchanged.
   - `newCancelService` (207) unchanged.
3. **Tests:**
   - `job_test.go`: `mockCancelService` (20–44) incl. `SystemCancelInstance` (40); NewWorker
     construction at 193/235/272/308; auto-cancel cases (229–232, 269, 303); read-source invariant
     (216–221).
   - `job_integration_test.go`: `TestIntegration_Watchdog_P1_AutoCancelEquivalence` (27–71, seeds
     `"auto_cancel"`, asserts cancelled/draft); `..._AlertOnlyEquivalence` (76–114, seeds
     `reduce_quorum`, asserts in_progress/under_review + 1 stuck_alert); `seedActiveStageSnapshot`
     (134–148); `assertInstanceStatus`/`assertDocumentStatus` helpers.
   - `cancel_service_test.go`: `SystemCancelInstance` unit test at 324–326 — remove; keep the
     `CancelInstance` cases.

## Task breakdown (sequential)

### T1 (sonnet) — remove the orphaned system-cancel path (production)
- Edit `cancel_service.go`: delete `SystemCancelInstance`; fold `cancelInstance` into `CancelInstance`
  (drop `system` param, delete bypass/seed block, unconditional `Require`). Keep all shared body.
- Edit `job.go`: delete dead branch + `autoCancelled`; remove `cancelSvcInterface`, `cancelSvc` +
  `runner` fields/params through `NewWorker`/`Work`/`run`; keep `WithBackgroundBypass` (fix comment);
  keep `DriftPolicy` read + alert payload.
- Update the `NewWorker` composition-root call site.
- `cancel_service_test.go`: remove the `SystemCancelInstance` unit case only.
- `go build ./...` + `go vet ./...` clean. Grep census (`auto_cancel`, `SystemCancelInstance`,
  `stuck_watchdog_auto_cancel`) = 0 in non-test production code.
- Commit: `refactor(jobs,approval): F5.8 T1 — remove orphaned watchdog auto-cancel + SystemCancelInstance (ADR 0068)`.

### T2 (sonnet) — honest alert-only tests
- `job_test.go`: drop `mockCancelService` / auto-cancel cases; fix `NewWorker` calls to new sig;
  keep the read-source guard if still valid.
- `job_integration_test.go`: replace the AutoCancel fiction with an honest alert-only proof (or merge
  into AlertOnly): stuck instance + valid drift value → `in_progress` + `under_review` + exactly one
  `stuck_alert`; fresh instance untouched. Remove now-unused helpers/asserts only.
- `go test -tags=integration -run TestIntegration_Watchdog ./internal/modules/jobs/stuck_instance_watchdog/... -v` → PASS on real Postgres (needs T1's testdb River provisioning from F5.7 T1 if the suite touches River — verify).
- Commit: `test(jobs): F5.8 T2 — replace watchdog auto-cancel fiction with honest alert-only proof`.

### T3 (main session, doc-only) — wiki alert-only
- Edit `wiki/modules/approval.md` watchdog section → alert-only; cite ADR 0068. Refresh Last verified.
- Commit: `docs(wiki): F5.8 T3 — approval.md watchdog is alert-only (ADR 0068)`.

## Test strategy
Targeted `-run` per changed proof only (no full suite). T1 is production removal (build+vet+grep
gate); T2 is the honest integration proof; T3 is docs. The user-cancel regression
(`-run Cancel ./internal/modules/documents/approval/...`) runs after T1 to prove the live path
intact.

## Files touched (census)
MODIFY `internal/modules/documents/approval/application/cancel_service.go` (T1),
`internal/modules/jobs/stuck_instance_watchdog/job.go` (T1), the `NewWorker` composition root (T1),
`internal/modules/documents/approval/application/cancel_service_test.go` (T1),
`internal/modules/jobs/stuck_instance_watchdog/job_test.go` (T2), `..._integration_test.go` (T2),
`wiki/modules/approval.md` (T3). No migration/DDL, no OpenAPI, no capability registry.
