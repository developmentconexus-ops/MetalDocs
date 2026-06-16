# F2.2 Scheduler Metrics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire the scheduler's per-job run/error/skip counters into `GET /api/v1/metrics` as a top-level `"scheduler"` key with a per-job grouped shape.

**Architecture:** A new `SchedulerMetricsProvider` interface in the `observability` platform package keeps it free of module imports (REQ-TOP-2). `*Scheduler` satisfies the interface via a new `SchedulerMetrics()` method that transforms `MetricsSnapshot` into the per-job grouped shape. `HTTPObservability` gains a `schedulerMetrics` field (set-once via `SetSchedulerMetrics` before serving) and merges it into `MetricsHandler()`'s payload as a top-level `"scheduler"` key.

**Tech Stack:** Go 1.22, `net/http`, `encoding/json`, `log/slog`. No new dependencies.

---

## File map

| File | Change |
|------|--------|
| `internal/platform/observability/http.go` | Add `SchedulerMetricsProvider` interface; `schedulerMetrics` field on `HTTPObservability`; `SetSchedulerMetrics` setter; nil-guard + `payload["scheduler"]` in `MetricsHandler` |
| `internal/modules/jobs/scheduler/scheduler.go` | Add package-level `schedulerMetricsFromSnapshot(MetricsSnapshot) map[string]any`; add `(*Scheduler).SchedulerMetrics() map[string]any` wrapper |
| `apps/api/cmd/metaldocs-api/main.go` | Add `httpObs.SetSchedulerMetrics(s)` after `registerScheduledJobs` (line 531), before `mux.Handle` (line 544) |
| `internal/platform/observability/http_scheduler_test.go` | **New** — `package observability_test`; `TestMetricsHandler_IncludesSchedulerCounters`, `TestMetricsHandler_NoSchedulerKey_WhenNotWired`, `stubSchedulerProvider` stub |
| `internal/modules/jobs/scheduler/scheduler_metrics_test.go` | **New** — `package scheduler`; `TestScheduler_SchedulerMetrics_GroupedByJob` |

---

## Task 1: Write the three failing tests (TDD red phase)

**Files:**
- Create: `internal/platform/observability/http_scheduler_test.go`
- Create: `internal/modules/jobs/scheduler/scheduler_metrics_test.go`

- [ ] **Step 1: Create `http_scheduler_test.go` with two handler tests**

```go
// internal/platform/observability/http_scheduler_test.go
package observability_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"metaldocs/internal/platform/observability"
)

// stubSchedulerProvider satisfies observability.SchedulerMetricsProvider.
type stubSchedulerProvider struct {
	data map[string]any
}

func (s *stubSchedulerProvider) SchedulerMetrics() map[string]any {
	return s.data
}

func TestMetricsHandler_IncludesSchedulerCounters(t *testing.T) {
	stub := &stubSchedulerProvider{
		data: map[string]any{
			"jobs": map[string]any{
				"probe": map[string]int64{"runs": 3, "errors": 1, "skips": 0},
			},
		},
	}
	o := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	o.SetSchedulerMetrics(stub)
	h := o.MetricsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	sched, ok := body["scheduler"]
	if !ok {
		t.Fatal("response body missing 'scheduler' key")
	}
	schedMap, ok := sched.(map[string]any)
	if !ok {
		t.Fatalf("scheduler value is not map[string]any: %T", sched)
	}
	jobs, ok := schedMap["jobs"].(map[string]any)
	if !ok {
		t.Fatalf("scheduler.jobs is not map[string]any: %T", schedMap["jobs"])
	}
	probe, ok := jobs["probe"].(map[string]any)
	if !ok {
		t.Fatalf("scheduler.jobs.probe is not map[string]any: %T", jobs["probe"])
	}
	// JSON numbers decode as float64.
	if got := probe["runs"]; got != float64(3) {
		t.Fatalf("scheduler.jobs.probe.runs = %v; want 3", got)
	}
	if got := probe["errors"]; got != float64(1) {
		t.Fatalf("scheduler.jobs.probe.errors = %v; want 1", got)
	}
}

func TestMetricsHandler_NoSchedulerKey_WhenNotWired(t *testing.T) {
	o := observability.NewHTTPObservability(func(*http.Request) string { return "" })
	// No SetSchedulerMetrics call — nil guard must suppress the key.
	h := o.MetricsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := body["scheduler"]; ok {
		t.Fatal("response body must not contain 'scheduler' key when no provider is wired")
	}
}
```

- [ ] **Step 2: Create `scheduler_metrics_test.go` with transform test**

```go
// internal/modules/jobs/scheduler/scheduler_metrics_test.go
package scheduler

import "testing"

func TestScheduler_SchedulerMetrics_GroupedByJob(t *testing.T) {
	snap := MetricsSnapshot{
		RunsTotal:   map[string]int64{"job-a": 3, "job-b": 1},
		ErrorsTotal: map[string]int64{"job-b": 1},
		SkipsTotal:  map[string]int64{"job-c": 2},
	}
	got := schedulerMetricsFromSnapshot(snap)

	jobsRaw, ok := got["jobs"]
	if !ok {
		t.Fatal("result missing 'jobs' key")
	}
	jobs, ok := jobsRaw.(map[string]any)
	if !ok {
		t.Fatalf("jobs is not map[string]any: %T", jobsRaw)
	}

	cases := []struct {
		name                string
		runs, errors, skips int64
	}{
		{"job-a", 3, 0, 0}, // only in RunsTotal
		{"job-b", 1, 1, 0}, // in RunsTotal + ErrorsTotal
		{"job-c", 0, 0, 2}, // only in SkipsTotal — merge edge case
	}
	for _, tc := range cases {
		entryRaw, ok := jobs[tc.name]
		if !ok {
			t.Fatalf("jobs[%q] missing", tc.name)
		}
		entry, ok := entryRaw.(map[string]int64)
		if !ok {
			t.Fatalf("jobs[%q] is not map[string]int64: %T", tc.name, entryRaw)
		}
		if entry["runs"] != tc.runs {
			t.Errorf("jobs[%q].runs = %d; want %d", tc.name, entry["runs"], tc.runs)
		}
		if entry["errors"] != tc.errors {
			t.Errorf("jobs[%q].errors = %d; want %d", tc.name, entry["errors"], tc.errors)
		}
		if entry["skips"] != tc.skips {
			t.Errorf("jobs[%q].skips = %d; want %d", tc.name, entry["skips"], tc.skips)
		}
	}
}
```

- [ ] **Step 3: Run both test packages — verify build failures (expected: symbols not defined yet)**

```
go test ./internal/platform/observability/ -run "TestMetricsHandler_IncludesSchedulerCounters|TestMetricsHandler_NoSchedulerKey_WhenNotWired" -count=1
go test ./internal/modules/jobs/scheduler/ -run TestScheduler_SchedulerMetrics_GroupedByJob -count=1
```

Expected: compile errors — `o.SetSchedulerMetrics` undefined; `schedulerMetricsFromSnapshot` undefined. Any other failure is a problem.

---

## Task 2: Add `SchedulerMetricsProvider` interface + field + setter + MetricsHandler merge

**Files:**
- Modify: `internal/platform/observability/http.go`

- [ ] **Step 1: Add interface and field to `http.go`**

After the existing `RuntimeStatusProvider` interface (line 12 of `runtime.go` — but the interface lives in `runtime.go`, not `http.go`). Add the new interface and field directly in `http.go`, above `HTTPObservability` struct definition (currently at line 29).

Insert before `type HTTPObservability struct {`:

```go
// SchedulerMetricsProvider is satisfied by any scheduler whose per-job counters
// should appear in the /api/v1/metrics payload. Defined here so the platform
// package stays free of module imports (REQ-TOP-2); the jobs package implements
// it implicitly via duck typing.
type SchedulerMetricsProvider interface {
	SchedulerMetrics() map[string]any
}
```

Add `schedulerMetrics SchedulerMetricsProvider` field to `HTTPObservability`:

```go
type HTTPObservability struct {
	runtimeProvider  RuntimeStatusProvider
	schedulerMetrics SchedulerMetricsProvider
	userIDResolver   func(*http.Request) string
	mu               sync.RWMutex
	byKey            map[string]*routeMetrics
}
```

- [ ] **Step 2: Add `SetSchedulerMetrics` setter after `NewHTTPObservability`**

Insert after the closing `}` of `NewHTTPObservability` (currently around line 65):

```go
// SetSchedulerMetrics wires a scheduler whose counters will appear as a top-level
// "scheduler" key in MetricsHandler's response. Call once during startup, before
// ListenAndServe — same set-once pattern as runtimeProvider.
func (o *HTTPObservability) SetSchedulerMetrics(s SchedulerMetricsProvider) {
	o.schedulerMetrics = s
}
```

- [ ] **Step 3: Merge `"scheduler"` key in `MetricsHandler`**

In `MetricsHandler` (currently at line 147), change the payload construction block. Current:

```go
items := o.snapshot()
payload := map[string]any{"items": items}
if o.runtimeProvider != nil {
    payload["runtime"] = o.runtimeProvider.RuntimeMetrics(r.Context())
}
```

Replace with:

```go
items := o.snapshot()
payload := map[string]any{"items": items}
if o.runtimeProvider != nil {
    payload["runtime"] = o.runtimeProvider.RuntimeMetrics(r.Context())
}
if o.schedulerMetrics != nil {
    payload["scheduler"] = o.schedulerMetrics.SchedulerMetrics()
}
```

- [ ] **Step 4: Run handler tests — both must pass now**

```
go test ./internal/platform/observability/ -run "TestMetricsHandler_IncludesSchedulerCounters|TestMetricsHandler_NoSchedulerKey_WhenNotWired" -count=1
```

Expected:
```
ok  	metaldocs/internal/platform/observability
```

Any FAIL at this point is a bug in the implementation — fix before continuing.

---

## Task 3: Add `schedulerMetricsFromSnapshot` + `SchedulerMetrics()` to scheduler

**Files:**
- Modify: `internal/modules/jobs/scheduler/scheduler.go`

- [ ] **Step 1: Add `schedulerMetricsFromSnapshot` package function and `SchedulerMetrics` method**

Append after `MetricsSnapshot()` (currently at line 275–277 of `scheduler.go`):

```go
// schedulerMetricsFromSnapshot transforms a MetricsSnapshot into the per-job
// grouped shape expected by the /api/v1/metrics "scheduler" payload:
//
//	{"jobs": {"job-name": {"runs": N, "errors": N, "skips": N}}}
//
// All three counter maps are merged so a job present only in ErrorsTotal or
// SkipsTotal (e.g. skipped before its first successful run) still appears.
func schedulerMetricsFromSnapshot(snap MetricsSnapshot) map[string]any {
	seen := make(map[string]struct{})
	jobs := make(map[string]any)

	for name, runs := range snap.RunsTotal {
		seen[name] = struct{}{}
		jobs[name] = map[string]int64{
			"runs":   runs,
			"errors": snap.ErrorsTotal[name],
			"skips":  snap.SkipsTotal[name],
		}
	}
	for name, errors := range snap.ErrorsTotal {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		jobs[name] = map[string]int64{
			"runs":   0,
			"errors": errors,
			"skips":  snap.SkipsTotal[name],
		}
	}
	for name, skips := range snap.SkipsTotal {
		if _, ok := seen[name]; ok {
			continue
		}
		jobs[name] = map[string]int64{
			"runs":   0,
			"errors": 0,
			"skips":  skips,
		}
	}
	return map[string]any{"jobs": jobs}
}

// SchedulerMetrics satisfies observability.SchedulerMetricsProvider. Returns the
// current per-job run/error/skip counters as a map suitable for JSON encoding in
// the /api/v1/metrics payload.
func (s *Scheduler) SchedulerMetrics() map[string]any {
	return schedulerMetricsFromSnapshot(s.MetricsSnapshot())
}
```

- [ ] **Step 2: Run scheduler transform test — must pass**

```
go test ./internal/modules/jobs/scheduler/ -run TestScheduler_SchedulerMetrics_GroupedByJob -count=1
```

Expected:
```
ok  	metaldocs/internal/modules/jobs/scheduler
```

---

## Task 4: Wire `SetSchedulerMetrics` in `main.go`

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/main.go`

- [ ] **Step 1: Add `httpObs.SetSchedulerMetrics(s)` after `registerScheduledJobs`**

Current lines 531–532 in `main.go`:
```go
registerScheduledJobs(s, deps, approvalServices.Cancel, approvalEmitter)

var schedulerWG sync.WaitGroup
```

Change to:
```go
registerScheduledJobs(s, deps, approvalServices.Cancel, approvalEmitter)
httpObs.SetSchedulerMetrics(s)

var schedulerWG sync.WaitGroup
```

No new imports needed — `s` is `*jobscheduler.Scheduler` which now satisfies `observability.SchedulerMetricsProvider` via duck typing; `httpObs` is already `*observability.HTTPObservability`.

- [ ] **Step 2: Build the API binary — verify no compile errors**

```
go build ./apps/api/...
```

Expected: exit 0, no output.

---

## Task 5: Full regression

- [ ] **Step 1: Run whole-repo test suite**

```
go test ./...
```

Expected: no `FAIL` lines. Note: DB integration tests are skipped in environments without a live DB — that is expected (look for `[no test files]` or `ok`, not `FAIL`).

---

## Task 6: Runtime proof (real-provider)

- [ ] **Step 1: Start the API**

```powershell
.\scripts\start-api.ps1 -Build
```

Wait for the line `"msg":"application_started"` (or equivalent startup log) to appear in stdout before proceeding.

- [ ] **Step 2: Get an auth token**

```powershell
$resp = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/login" `
    -Method POST `
    -ContentType "application/json" `
    -Body '{"identifier":"admin","password":"AdminMetalDocs123!"}'
$token = $resp.token
```

- [ ] **Step 3: Wait for first scheduler tick (~5 minutes for `stuck-instance-watchdog`)**

Watch stdout for a line containing `"msg":"scheduler_job_completed"`. Once one appears, proceed.

- [ ] **Step 4: Curl `/api/v1/metrics` and capture payload**

```powershell
$metrics = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/metrics" `
    -Method GET `
    -Headers @{ Authorization = "Bearer $token" }
$metrics | ConvertTo-Json -Depth 10
```

Expected: response body includes `"scheduler"` key with `"jobs"` sub-map containing at least one job entry where `runs` > 0.

Example expected snippet:
```json
{
  "scheduler": {
    "jobs": {
      "stuck-instance-watchdog": { "runs": 1, "errors": 0, "skips": 0 }
    }
  }
}
```

- [ ] **Step 5: Paste verbatim JSON snippet into `evidence.md`**

Record the raw JSON line from PowerShell output showing `scheduler.jobs` with at least one job entry. Label it **real-provider** per mission §8.

---

## Task 7: Commit

- [ ] **Step 1: Stage and commit**

```
git add internal/platform/observability/http.go \
        internal/platform/observability/http_scheduler_test.go \
        internal/modules/jobs/scheduler/scheduler.go \
        internal/modules/jobs/scheduler/scheduler_metrics_test.go \
        apps/api/cmd/metaldocs-api/main.go
git commit -m "feat(metrics): wire scheduler per-job counters into /api/v1/metrics (F2.2)"
```

---

## Self-review

**Spec coverage:**

| Spec requirement | Task |
|-----------------|------|
| `SchedulerMetricsProvider` interface in `observability` | Task 2 Step 1 |
| `schedulerMetrics` field + `SetSchedulerMetrics` setter | Task 2 Steps 1–2 |
| Top-level `"scheduler"` key in `MetricsHandler` | Task 2 Step 3 |
| Nil guard (absent when not wired) | Task 2 Step 3 (`if o.schedulerMetrics != nil`) |
| `schedulerMetricsFromSnapshot` per-job grouped transform | Task 3 Step 1 |
| Merge edge case — jobs only in ErrorsTotal/SkipsTotal | Task 3 Step 1 (three-pass loop) |
| `SchedulerMetrics()` method on `*Scheduler` | Task 3 Step 1 |
| `httpObs.SetSchedulerMetrics(s)` in `main.go` | Task 4 Step 1 |
| `TestMetricsHandler_IncludesSchedulerCounters` | Task 1 Step 1 |
| `TestMetricsHandler_NoSchedulerKey_WhenNotWired` | Task 1 Step 1 |
| `TestScheduler_SchedulerMetrics_GroupedByJob` | Task 1 Step 2 |
| Whole-repo regression (`go test ./...`) | Task 5 |
| Runtime real-provider proof | Task 6 |

No gaps found.

**Placeholder scan:** No TBD / TODO / "similar to" patterns. All code blocks are complete.

**Type consistency:**
- `SchedulerMetricsProvider.SchedulerMetrics()` returns `map[string]any` — matches field type, MetricsHandler usage, and stub return type ✅
- `schedulerMetricsFromSnapshot` takes `MetricsSnapshot` (exported struct from `scheduler.go:75`), used identically in test and implementation ✅
- Inner job entries are `map[string]int64` stored as `any` in outer `map[string]any` — test asserts `.(map[string]int64)` (pre-JSON), handler test asserts `float64` (post-JSON round-trip) ✅
- `SetSchedulerMetrics` receiver is `*HTTPObservability` — matches `o` variable type throughout ✅
