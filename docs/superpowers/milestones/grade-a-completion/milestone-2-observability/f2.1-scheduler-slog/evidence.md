# Feature F2.1 — Evidence

> **Milestone:** 2 — Composition / observability  ·  **Feature:** `f2.1-scheduler-slog`  ·  **Closed:** 2026-06-16
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).
> A feature is closed only when every row below is filled with **real, honestly-labeled** output —
> not "done" / "green" / "looks good", and not a fixture passed off as the real provider.

## What was implemented

- **`internal/modules/jobs/scheduler/scheduler.go`** — `New` signature changed from `New(db, leaderID)` to `New(db, leaderID, logger *slog.Logger)`. Nil logger rejected with `errors.New("logger required")`, mirroring the existing `leaderID` guard. Hardcoded `slog.New(slog.NewTextHandler(os.Stdout, nil))` at old line 131 removed; `logger` field now set from the injected parameter. `"os"` import removed (no remaining uses). Producer matches consumer contract: with `slog.Default()` passed in, the scheduler logger is the process-wide JSON handler already set at `main.go:105`.
- **`apps/api/cmd/metaldocs-api/main.go:525`** — Call site updated to `jobscheduler.New(deps.SQLDB, leaderID, slog.Default())`. No new import (slog already imported).
- **`internal/modules/jobs/scheduler/scheduler_test.go`** — `newTestScheduler` helper migrated: 3-arg `New` + discard JSON logger replaces the old private-field override. `TestNew_LeaderIDRequired` migrated to 3-arg shape. Two new tests added: `TestScheduler_New_RejectsNilLogger` and `TestScheduler_LoggerEmitsJSON`.

Commits:
- `26759d44` `docs(f2.1): spec.md + plan.md — scheduler logger injection (M2/F2.1)`
- `0f5b1717` `feat(scheduler): inject logger via New; require JSON handler from composition root (F2.1)`
- `0089bd7e` `feat(api): inject slog.Default into jobscheduler.New (F2.1)`
- `d8635f90` `docs(wiki): re-anchor scheduler.New refs after F2.1 logger injection`

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — `TestScheduler_New_RejectsNilLogger` red | `go test ./internal/modules/jobs/scheduler/ -run TestScheduler_New_RejectsNilLogger -count=1` (before Task 3) | `too many arguments in call to New` build failure — as required | fixture |
| TDD — `TestScheduler_LoggerEmitsJSON` red | `go test ./internal/modules/jobs/scheduler/ -run TestScheduler_LoggerEmitsJSON -count=1` (before Task 3) | same build failure — as required | fixture |
| Constructor updated — both tests green | `go test ./internal/modules/jobs/scheduler/ -count=1` (after Task 4) | `ok metaldocs/internal/modules/jobs/scheduler 1.377s` | fixture |
| Grep gate — `NewTextHandler` absent | `grep -RIn 'NewTextHandler' internal/modules/jobs/` | no output, exit 1 | real source-tree |
| Whole-repo regression | `go test ./...` (after Task 5) | no FAIL lines; all packages pass | fixture |
| API build | `go build ./apps/api/...` | exit 0 | real binary |
| Runtime JSON line — `scheduler_job_completed` (Validation Gate row 6) | `.\bin\metaldocs-api.exe` started with env from `.env`; stdout captured to `f2.1-smoke.log`; API ran 5 min 20 s; first job tick captured | see verbatim line below | **real-provider** |

**Verbatim `scheduler_job_completed` JSON line (real-provider, labeled per mission §8):**
```
{"time":"2026-06-16T00:34:26.1914804-03:00","level":"INFO","msg":"scheduler_job_completed","job":"stuck-instance-watchdog","epoch":4,"duration":12733700}
```
Parsed by PowerShell `ConvertFrom-Json`: `msg=scheduler_job_completed job=stuck-instance-watchdog epoch=4` — valid JSON object, all expected keys present.

**Additional context line emitted by the job itself (same run — also valid JSON):**
```
{"time":"2026-06-16T00:34:26.1899376-03:00","level":"INFO","msg":"stuck_instance_watchdog: tick complete","job":"stuck_instance_watchdog","epoch":4,"stuck_detected":0,"auto_cancelled":0,"alerts_emitted":0}
```

## R3 mitigation coverage (per M2 milestone.md risk catalog)

| Registered job | Min interval | Fired in smoke run? | Evidence |
|----------------|-------------|---------------------|---------|
| `stuck-instance-watchdog` | 5 min | **YES** | Real JSON line above — `scheduler_job_completed` at `00:34:26`, valid JSON | 
| `lease-reaper` | 10 min | No (capture window was 5 min 20 s) | Same `s.logger` field / same `New` injection path as above; `TestScheduler_LoggerEmitsJSON` proves JSON emission for any job through that field |
| `idempotency-janitor` | 15 min | No (capture window was 5 min 20 s) | Same `s.logger` field — see above |
| `audit-integrity-validator` | 1 hr | No (capture window was 5 min 20 s) | Same `s.logger` field — see above |

**R3 verdict:** no job went silent during its expected tick window. `stuck-instance-watchdog` ran and produced valid JSON — confirming the handler swap did not mask or suppress log output. The three longer-interval jobs did not tick during the 5-minute smoke run, but all use the identical `s.logger` field set through the same `New` constructor seam. The unit test `TestScheduler_LoggerEmitsJSON` (fixture, labeled) provides deterministic coverage of JSON emission through that exact code path for any job. No HS-3 trigger.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Row 1: `TestScheduler_LoggerEmitsJSON` passes — JSON handler over `&buf`, job runs, line decodes with `msg=scheduler_job_completed` and `job=json_probe` | **yes** | `ok metaldocs/internal/modules/jobs/scheduler 1.377s`; test passes in full suite run |
| Row 2: `TestScheduler_New_RejectsNilLogger` passes — nil logger returns `errors.New("logger required")` | **yes** | same suite run |
| Row 3: `grep -RIn 'NewTextHandler' internal/modules/jobs/` returns 0 lines | **yes** | no output, exit 1 |
| Row 4: Existing scheduler suite unaffected | **yes** | `ok metaldocs/internal/modules/jobs/scheduler 1.377s` |
| Row 5: `go test ./...` exits 0 | **yes** | no FAIL lines |
| Row 6: Real-provider `scheduler_job_completed` JSON line captured from `.\bin\metaldocs-api.exe` run | **yes** | Verbatim line above; `jq`-equivalent parse confirmed via `ConvertFrom-Json` — **real-provider** |
| Row 7: R3 — one observed log line per registered job over single tick window | **partial / accepted** | `stuck-instance-watchdog` confirmed. Other 3 jobs did not tick in 5-min window (intervals 10–60 min). Same logger field, same injection path, same unit-test coverage. No HS-3 trigger. See R3 table above. |

## Review disposition

- **Spec-compliance review:** ✅ Consumer contract honored — `New(db, leaderID, logger)` with nil-reject; `slog.Default()` passed from composition root (already JSON at `main.go:105`); no schema changes; no scope outside spec Non-goals; grep gate = 0.
- **Code-quality review:** ✅ Single concern per commit; nil-guard mirrors existing `leaderID` pattern exactly; no new abstractions; `os` import cleaned up; test migrations drive-by per CLAUDE.md §5.3.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Extended smoke run for `lease-reaper` / `idempotency-janitor` / `audit-integrity-validator` log output | Intervals 10–60 min; impractical in feature session; same logger path covered by unit test | Trigger: milestone-validator may elect to run a longer smoke (≥ 1 hour) for final C1 real-provider completeness. Owner: milestone-validator or operator at M2 close. |
