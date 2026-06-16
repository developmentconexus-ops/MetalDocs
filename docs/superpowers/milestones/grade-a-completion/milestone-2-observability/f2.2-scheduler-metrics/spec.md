# Feature F2.2 — Spec

> **Milestone:** 2 — Composition / observability  ·  **Folder:** `f2.2-scheduler-metrics`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-16 — leandrotca.work — *no implementation begins until this line is filled.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Driven inline (brainstorming engine flow; one question at a time, persisted below). Seed: M2
`milestone.md` F2.2 row + industry SaaS research on metrics payload structure.

| # | Question | Answer |
|---|----------|--------|
| 1 | Payload placement — top-level `"scheduler"` key, `"runtime.scheduler"`, or `"runtime.worker.scheduler"`? | **Top-level `"scheduler"` key** (sibling to `"items"` and `"runtime"`). Industry precedent (Prometheus naming hierarchy, GitHub internal health endpoint, Sidekiq, Kubernetes scheduler vs. kubelet): scheduler is a distinct system domain — not a sub-layer of runtime state or worker infrastructure. Worker = queue depth (infrastructure); scheduler = run/error/skip counters (execution orchestration). Different concerns, different top-level keys. Research confirms Option A. |
| 2 | Per-job payload shape — flat `{"runsTotal": {"job-a": 4}, "errorsTotal": {}}` (mirrors Go struct) or per-job grouped `{"jobs": {"job-a": {"runs": 4, "errors": 0, "skips": 0}}}`? | **Per-job grouped** (`"jobs"` map keyed by job name, value has `runs`/`errors`/`skips`). SRE looks up one job and gets all its stats in a single subtree. Matches Sidekiq and Kubernetes per-job metric objects. Requires a small transform from `MetricsSnapshot` but produces a significantly more consumer-friendly shape. |
| 3 | Injection mechanism — new constructor variadic, functional option, or setter method on `HTTPObservability`? | **Setter method** `SetSchedulerMetrics(s SchedulerMetricsProvider)`. `httpObs` is constructed at `main.go:277` and scheduler `s` at `main.go:524` — 250 lines apart. A setter called between lines 531 and 544 (after `registerScheduledJobs`, before `mux.Handle`) requires zero code reorder. Matches the `runtimeProvider` set-once-at-startup pattern; safe because it is called before `ListenAndServe`. |
| 4 | Interface placement — define `SchedulerMetricsProvider` in `observability` package or a shared platform package? | **In `observability` package**, alongside `RuntimeStatusProvider`. Keeps platform package self-contained; the `*Scheduler` in the jobs package satisfies it without the `observability` package importing the jobs package (REQ-TOP-2: platform packages stay free of module imports). |
| 5 | Jobs not yet in `RunsTotal` but present in `ErrorsTotal` or `SkipsTotal` — include them in the grouped output? | **Yes.** The `SchedulerMetrics()` transform walks all three maps and merges by job name so no job with any counter is silently omitted. A job that errored on its first attempt (before a completed run) would be invisible in a `RunsTotal`-only walk. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):**
  - **Primary — SRE / operator scraping `GET /api/v1/metrics`** (curl, monitoring agent, Datadog,
    Grafana JSON datasource). They need to answer: "Did job X run? Did it error? Was it skipped?"
    with a single JSON lookup `scheduler.jobs["job-name"]` without cross-referencing three
    separate flat maps.
  - **Secondary — handler integration test** asserting the `"scheduler"` key and its `"jobs"`
    sub-map exist and have the correct shape.

- **Contract:**
  1. `GET /api/v1/metrics` response body includes a top-level `"scheduler"` key alongside
     `"items"` and `"runtime"`. This key is **absent** when no `SchedulerMetricsProvider` is
     wired (nil-guard — existing payload shape is never broken for deployments that do not wire
     the scheduler).
  2. `"scheduler"` value is `{ "jobs": { "<job-name>": { "runs": <int64>, "errors": <int64>,
     "skips": <int64> } } }`. Each registered job that has any non-zero counter appears as a
     key. Jobs with all-zero counters may be absent (empty map is also valid; no guarantee of
     presence until the first tick).
  3. `"runs"` = total completed invocations regardless of error; `"errors"` = invocations that
     returned non-nil error; `"skips"` = invocations skipped due to backpressure.
  4. The existing `"items"` and `"runtime"` keys are **unmodified** — no key renamed, no key
     removed, no value type changed.
  5. `SchedulerMetricsProvider` is a new interface in `internal/platform/observability/` with a
     single method `SchedulerMetrics() map[string]any`. `*jobscheduler.Scheduler` satisfies it
     by implementing `SchedulerMetrics()` which transforms `MetricsSnapshot` into the
     per-job grouped shape.

- **Source of truth for the contract:**
  - Existing payload handler → [`internal/platform/observability/http.go:147`](../../../../../internal/platform/observability/http.go)
    (`MetricsHandler`).
  - Existing `MetricsSnapshot` struct → [`internal/modules/jobs/scheduler/scheduler.go:75`](../../../../../internal/modules/jobs/scheduler/scheduler.go).
  - Mission defect anchor → `mission.md` §5 D2.

## What this feature implements

Wire `MetricsSnapshot()` (already exists on `*Scheduler`) into `/api/v1/metrics` via a new
platform-level interface. Concrete changes, scoped to three files:

1. **`internal/platform/observability/http.go`**
   - Add `SchedulerMetricsProvider` interface (one method: `SchedulerMetrics() map[string]any`).
   - Add `schedulerMetrics SchedulerMetricsProvider` field to `HTTPObservability`.
   - Add `SetSchedulerMetrics(s SchedulerMetricsProvider)` setter (assigns the field; called
     once during startup, before `ListenAndServe`).
   - In `MetricsHandler()`: if `o.schedulerMetrics != nil`, add `payload["scheduler"] =
     o.schedulerMetrics.SchedulerMetrics()`.

2. **`internal/modules/jobs/scheduler/scheduler.go`**
   - Add `SchedulerMetrics() map[string]any` method on `*Scheduler`. Calls `s.MetricsSnapshot()`,
     walks all three maps (`RunsTotal`, `ErrorsTotal`, `SkipsTotal`) merged by job name, returns
     `map[string]any{"jobs": <per-job-map>}`. No change to existing methods or struct fields.

3. **`apps/api/cmd/metaldocs-api/main.go`**
   - Add `httpObs.SetSchedulerMetrics(s)` after line 531 (`registerScheduledJobs`) and before
     line 544 (`mux.Handle("/api/v1/metrics", ...)`). One line change.

## Non-goals (mandatory)

Explicitly out of scope. Anything here that later appears in the diff is scope drift (validator C6).

- **No aggregated totals** (`"total": {"runs": N}` across all jobs). Per-job is sufficient; an
  aggregate can be derived by the scraper. YAGNI.
- **No zero-counter job entries guaranteed.** Jobs not yet ticked are absent from the map.
  The contract does not require pre-populating all registered job names; the first tick creates
  the entry.
- **No change to `"items"` or `"runtime"` payload shape.** R2 (milestone.md risk) — existing
  keys are purely additive additions.
- **No Prometheus/OpenMetrics format change.** The endpoint stays JSON; no exporter swap or
  content-type negotiation. (HS-2 rabbit hole.)
- **No new endpoint.** Scheduler counters land in the existing `/api/v1/metrics` response only.
- **No touch to `StaticRuntimeStatusProvider` or `PostgresRuntimeStatusProvider`.** The scheduler
  is injected into `HTTPObservability` directly, not threaded through `RuntimeStatusProvider`.
- **No change to `MetricsSnapshot` or `Metrics` struct.** The transform in
  `SchedulerMetrics()` reads the existing snapshot as-is.
- **No other `jobs/` package changes** beyond adding one method to `scheduler.go`.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| 1. `MetricsHandler` includes `"scheduler"` top-level key with a `"jobs"` sub-map when a `SchedulerMetricsProvider` is wired. | `go test ./internal/platform/observability/ -run TestMetricsHandler_IncludesSchedulerCounters -count=1` — stub provider returning `{"jobs":{"probe":{"runs":3,"errors":1,"skips":0}}}`; assert decoded response body has `scheduler.jobs.probe.runs == 3`. | fixture |
| 2. `MetricsHandler` does NOT include `"scheduler"` key when no provider is wired (nil guard — existing payload shape preserved). | `go test ./internal/platform/observability/ -run TestMetricsHandler_NoSchedulerKey_WhenNotWired -count=1` — no `SetSchedulerMetrics` call; assert decoded body lacks `"scheduler"` key. | fixture |
| 3. `Scheduler.SchedulerMetrics()` produces per-job grouped shape including jobs present only in `ErrorsTotal` or `SkipsTotal`. | `go test ./internal/modules/jobs/scheduler/ -run TestScheduler_SchedulerMetrics_GroupedByJob -count=1` — seed known `MetricsSnapshot`; assert per-job map shape and merged coverage. | fixture |
| 4. Whole-repo regression — existing `"items"` and `"runtime"` keys unaffected. | `go test ./...` exits 0; no FAIL lines. | fixture |
| 5. Runtime proof — `curl /api/v1/metrics` response body includes `"scheduler"` key with non-empty `"jobs"` map after at least one job tick. | `.\scripts\start-api.ps1` started; after first scheduler tick, `curl http://localhost:8081/api/v1/metrics` (with auth token); response body parsed; `scheduler.jobs` map has at least one entry with `runs > 0`. Sample JSON snippet pasted verbatim into `evidence.md`, labeled **real-provider**. | **real-provider** (per mission §8 label) |

> TDD: rows 1, 2, 3 are the failing tests, written first. Row 4 is the regression guard. Row 5 is
> runtime evidence captured from a real `start-api.ps1` run after implementation.

## ADR needed?

- [x] No durable decision — skip. Top-level payload placement is a tactical wiring choice;
      the durable architecture statement ("composition-root injection, not in-package wiring") is
      already recorded in `milestone.md` §Dependencies & constraints and `mission.md` §5 D2. The
      `SchedulerMetricsProvider` interface follows the existing `RuntimeStatusProvider` pattern
      (same package, same injection seam) — no new pattern introduced.
- [ ] Durable decision made → record an ADR under `wiki/decisions/` and link it here.
