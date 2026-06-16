# Feature F2.2 — Evidence

> **Milestone:** 2 — Composition / observability  ·  **Feature:** `f2.2-scheduler-metrics`  ·  **Closed:** 2026-06-16
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).
> A feature is closed only when every row below is filled with **real, honestly-labeled** output —
> not "done" / "green" / "looks good", and not a fixture passed off as the real provider.

## What was implemented

- **`internal/platform/observability/http.go`** — Added `SchedulerMetricsProvider` interface (single method `SchedulerMetrics() map[string]any`); `schedulerMetrics SchedulerMetricsProvider` field on `HTTPObservability`; `SetSchedulerMetrics(s SchedulerMetricsProvider)` setter; nil-guard in `MetricsHandler()` that adds `payload["scheduler"] = o.schedulerMetrics.SchedulerMetrics()` only when provider is wired.
- **`internal/modules/jobs/scheduler/scheduler.go`** — Added `schedulerMetricsFromSnapshot(snap MetricsSnapshot) map[string]any` (three-pass merge across `RunsTotal`, `ErrorsTotal`, `SkipsTotal` with `seen` set to handle jobs present in only one counter map); `SchedulerMetrics() map[string]any` method on `*Scheduler` that calls `s.MetricsSnapshot()` and delegates to the transform. No change to existing methods or struct fields.
- **`apps/api/cmd/metaldocs-api/main.go`** — Added `httpObs.SetSchedulerMetrics(s)` after `registerScheduledJobs` (line 531), before `mux.Handle("/api/v1/metrics", ...)` (line 545). One line; zero new imports; called before `ListenAndServe`.
- **New test files:**
  - `internal/platform/observability/http_scheduler_test.go` — `TestMetricsHandler_IncludesSchedulerCounters`, `TestMetricsHandler_NoSchedulerKey_WhenNotWired`
  - `internal/modules/jobs/scheduler/scheduler_metrics_test.go` — `TestScheduler_SchedulerMetrics_GroupedByJob`

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD red — 3 failing tests (before implementation) | `go test ./internal/platform/observability/ ./internal/modules/jobs/scheduler/` (before Tasks 2–3) | `undefined: SetSchedulerMetrics`, `undefined: schedulerMetricsFromSnapshot` — build failures as required | fixture |
| AC row 1 — `TestMetricsHandler_IncludesSchedulerCounters` | `go test ./internal/platform/observability/ -run TestMetricsHandler_IncludesSchedulerCounters -count=1 -v` | `--- PASS: TestMetricsHandler_IncludesSchedulerCounters (0.00s)` | fixture |
| AC row 2 — `TestMetricsHandler_NoSchedulerKey_WhenNotWired` | `go test ./internal/platform/observability/ -run TestMetricsHandler_NoSchedulerKey_WhenNotWired -count=1 -v` | `--- PASS: TestMetricsHandler_NoSchedulerKey_WhenNotWired (0.00s)` | fixture |
| AC row 3 — `TestScheduler_SchedulerMetrics_GroupedByJob` | `go test ./internal/modules/jobs/scheduler/ -run TestScheduler_SchedulerMetrics_GroupedByJob -count=1 -v` | `--- PASS: TestScheduler_SchedulerMetrics_GroupedByJob (0.00s)` | fixture |
| AC row 4 — whole-repo regression | `go test ./...` | no FAIL lines; all packages pass or `[no test files]` | fixture |
| API build | `go build ./apps/api/...` | exit 0 | real binary |
| AC row 5 — runtime proof (Validation Gate row 5) | Fresh binary started; scheduler tick observed; `curl http://localhost:8081/api/v1/metrics` with auth session; `scheduler.jobs.stuck-instance-watchdog.runs == 1` | see verbatim JSON below | **real-provider** |

**Verbatim `GET /api/v1/metrics` JSON (real-provider, labeled per mission §8):**
```json
{
  "items": [...],
  "runtime": {
    "auth": { "sessions": {"active":6,"expired":175,"revoked":1}, "users": {"active":72,"inactive":0,"locked":0,"mustChangePassword":0} },
    "authEnabled": true,
    "repositoryMode": "postgres",
    "repositoryStatus": "up",
    "storageProvider": "minio",
    "worker": { "outbox": {"claimable":0,"deadLettered":1,"pending":0} }
  },
  "scheduler": {
    "jobs": {
      "stuck-instance-watchdog": {
        "errors": 0,
        "runs": 1,
        "skips": 0
      }
    }
  }
}
```

Key observations:
- `"scheduler"` is a top-level key, sibling to `"items"` and `"runtime"` — matches spec contract §1.
- `"jobs"` sub-map keyed by job name, each value `{"runs":N,"errors":N,"skips":N}` — matches spec contract §2.
- `"items"` and `"runtime"` values unchanged — spec contract §4 satisfied.
- `stuck-instance-watchdog` fired once (`runs: 1`, `errors: 0`, `skips: 0`) after ~5-min tick — **real-provider**.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Row 1: `TestMetricsHandler_IncludesSchedulerCounters` passes — stub with `{"jobs":{"probe":{"runs":3,"errors":1,"skips":0}}}`; decoded body has `scheduler.jobs.probe.runs == float64(3)` | **yes** | `--- PASS` above |
| Row 2: `TestMetricsHandler_NoSchedulerKey_WhenNotWired` passes — no `SetSchedulerMetrics`; decoded body lacks `"scheduler"` key | **yes** | `--- PASS` above |
| Row 3: `TestScheduler_SchedulerMetrics_GroupedByJob` passes — seeded `MetricsSnapshot` with jobs across all three counter maps; per-job grouped shape verified including edge-case job-c (only in SkipsTotal) | **yes** | `--- PASS` above |
| Row 4: `go test ./...` exits 0; no FAIL lines | **yes** | no FAIL lines in full suite run |
| Row 5: Runtime proof — `curl /api/v1/metrics` body includes `"scheduler"` key with non-empty `"jobs"` map after at least one job tick; `runs > 0` | **yes** | Verbatim JSON above — `stuck-instance-watchdog.runs: 1` — **real-provider** |

## Review disposition

- **Spec-compliance review (per-task):** ✅ All three implementation tasks reviewed — interface placement in `observability` package (REQ-TOP-2 satisfied), nil-guard present, three-pass merge covers edge case, one-line main.go wiring after `registerScheduledJobs` before `mux.Handle`.
- **Code-quality review (per-task):** ✅ Setter follows existing `runtimeProvider` set-once pattern; duck-typing via interface avoids import cycle; no new abstractions beyond spec; `seen` set handles counter-map merge correctly.

## Bounded defers

None. All spec Validation Gate rows met with real-provider evidence for row 5. No partial mitigations.
