# F2.1 Scheduler Slog Injection — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the hardcoded `slog.NewTextHandler` in the scheduler with composition-root injection of `*slog.Logger`, so every scheduler log line on the API process's stdout is JSON-parseable by the shared log shipper.

**Architecture:** Add a required `logger *slog.Logger` positional parameter to `jobscheduler.New`. Nil → reject with `errors.New("logger required")` (mirrors the existing empty-`leaderID` check). The single caller [`apps/api/cmd/metaldocs-api/main.go:525`](../../../../../apps/api/cmd/metaldocs-api/main.go) passes `slog.Default()`, which `main.go:105` already sets to `slog.NewJSONHandler(os.Stdout, nil)`. No new abstractions; the other hardcoded scheduler knobs (`heartbeatEvery`, `drainWait`, `forceWait`, `maxSkipStreak`) stay untouched.

**Tech Stack:** Go 1.22+, `log/slog`, existing `*testing.T` table tests, `bytes.Buffer` + `encoding/json` for fixture proof.

**Source spec:** [`spec.md`](spec.md) (approved 2026-06-16). Non-goals there are binding.

---

## File structure

| File | Role | Action |
|------|------|--------|
| `internal/modules/jobs/scheduler/scheduler.go` | Scheduler type + `New` constructor (the producer) | Modify: signature + nil-check + drop hardcoded handler + drop `os` import if unused after edit |
| `internal/modules/jobs/scheduler/scheduler_test.go` | Scheduler unit tests | Modify: `newTestScheduler` helper signature update, drop private-field override at line 188; existing `TestNew_LeaderIDRequired` updated to new 3-arg shape; add `TestScheduler_New_RejectsNilLogger`, `TestScheduler_LoggerEmitsJSON` |
| `apps/api/cmd/metaldocs-api/main.go` | API composition root | Modify: single call site at line 525 — append `slog.Default()` arg |

No file split; no new package.

---

### Task 1: Failing test — `TestScheduler_New_RejectsNilLogger`

**Files:**
- Modify: `internal/modules/jobs/scheduler/scheduler_test.go` (append at end of file)

- [ ] **Step 1: Add the failing test**

Append to `internal/modules/jobs/scheduler/scheduler_test.go`:

```go
func TestScheduler_New_RejectsNilLogger(t *testing.T) {
	db := openSchedulerTestDB(t)
	_, err := New(db, "test-leader", nil)
	if err == nil || err.Error() != "logger required" {
		t.Fatalf("New(db, \"test-leader\", nil) error = %v, want \"logger required\"", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails (compile error expected)**

Run:
```
go test ./internal/modules/jobs/scheduler/ -run TestScheduler_New_RejectsNilLogger -count=1
```
Expected: build failure — `too many arguments in call to New` (current signature is `New(db, leaderID)`).

- [ ] **Step 3: Do NOT implement yet — proceed to Task 2 to add the second failing test before any constructor change.**

---

### Task 2: Failing test — `TestScheduler_LoggerEmitsJSON`

**Files:**
- Modify: `internal/modules/jobs/scheduler/scheduler_test.go` (append; add `encoding/json`, `bytes` imports if absent)

- [ ] **Step 1: Add the failing test**

Append to `internal/modules/jobs/scheduler/scheduler_test.go`:

```go
func TestScheduler_LoggerEmitsJSON(t *testing.T) {
	db := openSchedulerTestDB(t)
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	s, err := New(db, "test-leader", logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	s.heartbeatEvery = 5 * time.Millisecond
	s.drainWait = 50 * time.Millisecond
	s.forceWait = 50 * time.Millisecond

	ran := make(chan struct{}, 1)
	s.Register(JobConfig{
		Name:     "json_probe",
		Interval: 10 * time.Millisecond,
		Fn: func(ctx context.Context, epoch int64) error {
			select {
			case ran <- struct{}{}:
			default:
			}
			return nil
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.Start(ctx)
	}()

	select {
	case <-ran:
	case <-ctx.Done():
		t.Fatalf("job did not run before context deadline")
	}

	cancel()
	<-done

	var foundCompleted bool
	for _, line := range bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal(line, &rec); err != nil {
			t.Fatalf("scheduler log line is not JSON: %q (err: %v)", string(line), err)
		}
		if rec["msg"] == "scheduler_job_completed" && rec["job"] == "json_probe" {
			foundCompleted = true
		}
	}
	if !foundCompleted {
		t.Fatalf("no JSON line with msg=scheduler_job_completed job=json_probe in %q", buf.String())
	}
}
```

- [ ] **Step 2: Add required imports if missing**

Ensure the `import (...)` block in `scheduler_test.go` contains `"bytes"`, `"context"`, `"encoding/json"`. Add any that are missing. Leave others untouched.

- [ ] **Step 3: Run both new tests to verify they fail**

Run:
```
go test ./internal/modules/jobs/scheduler/ -run "TestScheduler_New_RejectsNilLogger|TestScheduler_LoggerEmitsJSON" -count=1
```
Expected: build failure — `too many arguments in call to New`.

---

### Task 3: Update `New` signature + drop hardcoded handler

**Files:**
- Modify: `internal/modules/jobs/scheduler/scheduler.go:118-133` and the import block at lines 3–11

- [ ] **Step 1: Replace the constructor**

Replace lines 118–133 of `internal/modules/jobs/scheduler/scheduler.go`:

```go
func New(db *sql.DB, leaderID string, logger *slog.Logger) (*Scheduler, error) {
	if leaderID == "" {
		return nil, errors.New("leaderID required")
	}
	if logger == nil {
		return nil, errors.New("logger required")
	}
	return &Scheduler{
		db:             db,
		leaderID:       leaderID,
		metrics:        newMetrics(),
		inFlight:       map[*inFlightJob]struct{}{},
		heartbeatEvery: time.Minute,
		drainWait:      30 * time.Second,
		forceWait:      5 * time.Second,
		maxSkipStreak:  10,
		logger:         logger,
	}, nil
}
```

- [ ] **Step 2: Remove the now-unused `os` import**

`os` was used only by the deleted `slog.NewTextHandler(os.Stdout, nil)` literal. Grep to confirm no other use, then delete the `"os"` line from the import block at the top of `scheduler.go`.

Run:
```
grep -n "\bos\." internal/modules/jobs/scheduler/scheduler.go
```
Expected: no output. Then edit the import block to drop `"os"`.

- [ ] **Step 3: Verify the package still builds (tests will still fail — call sites not yet migrated)**

Run:
```
go build ./internal/modules/jobs/scheduler/
```
Expected: success.

- [ ] **Step 4: Do NOT commit yet — `go test` of this package will still fail (the existing `TestNew_LeaderIDRequired` and `newTestScheduler` helper call `New` with the old 2-arg shape). Migrate them in Task 4.**

---

### Task 4: Migrate existing test call sites + delete private-field override

**Files:**
- Modify: `internal/modules/jobs/scheduler/scheduler_test.go:180-189` (helper) and `:204-208` (`TestNew_LeaderIDRequired`)

- [ ] **Step 1: Replace `newTestScheduler`**

Replace lines 180–190 of `internal/modules/jobs/scheduler/scheduler_test.go`:

```go
func newTestScheduler(db *sql.DB) *Scheduler {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	s, err := New(db, "test-leader", logger)
	if err != nil {
		panic(err)
	}
	s.heartbeatEvery = 5 * time.Millisecond
	s.drainWait = 200 * time.Millisecond
	s.forceWait = 100 * time.Millisecond
	return s
}
```

(The private-field override `s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))` is gone — the constructor already received the discard logger. `NewTextHandler` no longer appears anywhere in `internal/modules/jobs/`.)

- [ ] **Step 2: Update `TestNew_LeaderIDRequired` to the new 3-arg shape**

Replace lines 204–208 of `internal/modules/jobs/scheduler/scheduler_test.go`:

```go
func TestNew_LeaderIDRequired(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	if _, err := New(nil, "", logger); err == nil || err.Error() != "leaderID required" {
		t.Fatalf("New(nil, \"\", logger) error = %v, want leaderID required", err)
	}
}
```

- [ ] **Step 3: Run the full scheduler test suite — all four target tests must pass green**

Run:
```
go test ./internal/modules/jobs/scheduler/ -count=1 -v
```
Expected: PASS, including `TestNew_LeaderIDRequired`, `TestScheduler_New_RejectsNilLogger`, `TestScheduler_LoggerEmitsJSON`, and every pre-existing test in the package.

- [ ] **Step 4: Run the grep gate (Validation Gate row 3)**

Run:
```
grep -RIn 'NewTextHandler' internal/modules/jobs/
```
Expected: no output, exit code 1 (no matches). If any line prints, find the leak and remove it before continuing.

- [ ] **Step 5: Commit scheduler-package change**

```
git add internal/modules/jobs/scheduler/scheduler.go internal/modules/jobs/scheduler/scheduler_test.go
git commit -m "feat(scheduler): inject logger via New; require JSON handler from composition root (F2.1)"
```

---

### Task 5: Wire the composition-root call site

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/main.go:525`

- [ ] **Step 1: Update the call site**

Open `apps/api/cmd/metaldocs-api/main.go`. Locate line 525:

```go
s, err := jobscheduler.New(deps.SQLDB, leaderID)
```

Replace it with:

```go
s, err := jobscheduler.New(deps.SQLDB, leaderID, slog.Default())
```

`slog` is already imported in this file (it's used at `main.go:105` for `SetDefault` and at `:527` for `slog.Error`). No import edit required — verify by re-reading the import block if unsure.

- [ ] **Step 2: Build the API binary**

Run:
```
go build ./apps/api/...
```
Expected: success.

- [ ] **Step 3: Whole-repo regression (Validation Gate row 5)**

Run:
```
go test ./...
```
Expected: PASS across the repo. Any failure outside the scheduler package indicates a leaked dependency on the old 2-arg `New`; locate, fix narrowly (no scope drift), then re-run.

- [ ] **Step 4: Commit the call-site wiring**

```
git add apps/api/cmd/metaldocs-api/main.go
git commit -m "feat(api): inject slog.Default into jobscheduler.New (F2.1)"
```

---

### Task 6: Real-provider runtime proof + R3 mitigation

**Files:**
- Create: `docs/superpowers/milestones/grade-a-completion/milestone-2-observability/f2.1-scheduler-slog/evidence.md` (use `templates/feature-evidence.md` as the skeleton)

> Validation Gate rows 6 and 7. These are runtime evidence steps — not code edits. They produce the artifacts the milestone-validator (C1) reads.

- [ ] **Step 1: Start the API per the script-truth policy (CLAUDE.md §1)**

Run from a PowerShell window in the repo root:
```
.\scripts\start-api.ps1 -Build
```
Tee stdout to a capture file:
```
.\scripts\start-api.ps1 -Build *>&1 | Tee-Object -FilePath .\f2.1-stdout.log
```
Let it run long enough for **every registered scheduled job** to tick at least once. Inspect `registerScheduledJobs` in `apps/api/cmd/metaldocs-api/main.go` (the call at `:531`) to enumerate the job names; the slowest registered interval is the minimum capture window.

- [ ] **Step 2: Verify at least one `scheduler_job_completed` JSON line**

Run:
```
Select-String -Path .\f2.1-stdout.log -Pattern 'scheduler_job_completed' | Select-Object -First 1
```
Pipe the matched line through `jq .` to confirm it decodes:
```
(Select-String -Path .\f2.1-stdout.log -Pattern 'scheduler_job_completed' | Select-Object -First 1 -ExpandProperty Line) | jq .
```
Expected: a single JSON object printed by `jq` with keys including `time`, `level`, `msg`, `job`, `epoch`, `duration`. Exit code 0.

- [ ] **Step 3: R3 — confirm no scheduled job is silent**

Build the registered-job list from `registerScheduledJobs` (read source). For each job name `<n>`, run:
```
Select-String -Path .\f2.1-stdout.log -Pattern '"job":"<n>"' | Select-Object -First 1
```
Every registered job must produce at least one matching line within its interval window. If any job is silent → **HS-3 trigger**: stop, repair via `runtime-contract-prereq`, re-run the failed checkpoint, then resume.

- [ ] **Step 4: Populate `evidence.md`**

Copy `.claude/skills/milestone/templates/feature-evidence.md` to `docs/superpowers/milestones/grade-a-completion/milestone-2-observability/f2.1-scheduler-slog/evidence.md` and fill:
- Validation Gate rows 1–5: paste the exact `go test` / `grep` commands run in Tasks 1–5 with their real output (PASS lines, exit codes).
- Validation Gate row 6: paste **verbatim** one JSON line from `f2.1-stdout.log` containing `"msg":"scheduler_job_completed"`. Label it `real-provider`.
- Validation Gate row 7: a checklist of registered job names with the first observed log line for each. Label it `real-provider`.
- Bounded defers: none unless something surfaced; if anything did, write the trigger and an owner.

- [ ] **Step 5: Commit evidence**

```
git add docs/superpowers/milestones/grade-a-completion/milestone-2-observability/f2.1-scheduler-slog/evidence.md
git commit -m "docs(f2.1): evidence — JSON scheduler log line + R3 per-job coverage"
```

---

### Task 7: Wiki sync + module-doc-sync dispatch

**Files:**
- Possibly modify: any wiki doc whose `Last verified:` anchor pointed at the deleted line range in `scheduler.go`.

- [ ] **Step 1: Find wiki references to the scheduler constructor**

Run:
```
grep -RIn 'scheduler.go:13[0-9]\|scheduler.go:1[12][0-9]\|jobscheduler.New' wiki/
```
For every hit: re-anchor the `file:line` to the new line numbers in `scheduler.go`, and bump `Last verified:` to today's date.

Two hits are already known from the earlier audit:
- `wiki/backend/_artifacts/stage1/async-runtime.md:186` — anchor `apps/api/main.go:523` (now `:525` after the M1-era shift, still `:525` here) and the New-arg shape.
- `wiki/backend/binaries/worker.md:127` — same anchor.

Update both to the current line and the new 3-arg signature; bump their `Last verified:` stamps.

- [ ] **Step 2: Per CLAUDE.md §2, dispatch `wiki-curator` if anything beyond the two known hits exists**

If `grep` surfaced other references → dispatch the `wiki-curator` agent on the named change ("F2.1 scheduler logger injection; New signature now `(db, leaderID, logger)`; hardcoded text handler removed").

- [ ] **Step 3: Commit wiki updates**

```
git add wiki/
git commit -m "docs(wiki): re-anchor scheduler.New refs after F2.1 logger injection"
```

---

## Self-review (run before handoff)

- **Spec coverage** — every Validation Gate row in `spec.md` is implemented:
  - Row 1 (`TestScheduler_LoggerEmitsJSON`) → Task 2 + Task 4 step 3.
  - Row 2 (`TestScheduler_New_RejectsNilLogger`) → Task 1 + Task 4 step 3.
  - Row 3 (grep-gate) → Task 4 step 4.
  - Row 4 (scheduler suite) → Task 4 step 3.
  - Row 5 (whole-repo) → Task 5 step 3.
  - Row 6 (real-provider JSON line) → Task 6 steps 1–2 + evidence row.
  - Row 7 (R3 per-job coverage) → Task 6 step 3 + evidence row.
- **Placeholder scan** — no TBD / TODO / "implement later" / "similar to Task N". Every code step shows the code; every command step shows the command + expected outcome.
- **Type consistency** — `New(db, leaderID, logger)` used identically in Tasks 1, 2, 3, 4, 5. `jobscheduler.New` is the import-qualified name in `main.go` (the package is imported as `jobscheduler`); package-local `New` everywhere else. Error string `"logger required"` matches across the test, the constructor, and the spec's Validation Gate row 2.
- **Non-goal guard** — no plan step touches `apps/worker`, `apps/jobs`, the scheduler's other hardcoded knobs, `slog.SetDefault`, or any package outside `internal/modules/jobs/scheduler/` + the single API call site + (read-only) wiki re-anchoring. ADR not produced (per spec).

---

## Execution handoff

Plan saved to [`plan.md`](plan.md) (this file). Two execution options:

1. **Subagent-Driven (recommended)** — dispatch a fresh subagent per task, review between tasks, fast iteration. Uses `superpowers:subagent-driven-development`.
2. **Inline Execution** — execute tasks in this session via `superpowers:executing-plans`, batch with checkpoints.

Which approach?
