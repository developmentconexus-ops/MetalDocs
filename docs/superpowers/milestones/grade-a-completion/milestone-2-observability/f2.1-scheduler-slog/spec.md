# Feature F2.1 — Spec

> **Milestone:** 2 — Composition / observability  ·  **Folder:** `f2.1-scheduler-slog`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-16 — leandrotca.work — *no implementation begins until this line is filled.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Driven inline (brainstorming engine flow; one question at a time, persisted below). Seed: M2
`milestone.md` F2.1 row.

| # | Question | Answer |
|---|----------|--------|
| 1 | Constructor signature shape — positional logger param, functional options, or Config struct? | **A — positional param.** `New(db, leaderID, logger)`. Matches existing two-arg style, one caller, minimal diff, no new abstractions (M2 quality goal #3). |
| 2 | Nil-logger behavior — fall back to `slog.Default()` or reject like empty `leaderID`? | **Reject.** Fail-loud; no silent fallback that could mask wiring bugs in a future second caller. Returns `errors.New("logger required")`. |
| 3 | Failing-test shape for TDD — pure unit with `slog.NewJSONHandler(&buf, nil)`, stdout-capture integration, or grep-gate only? | **A — pure unit.** Pass JSON handler over `&buf` through `New`, register a no-op job at 10ms, run `Start` under a 50ms context, assert `buf` contains a line decodable as JSON with `msg == "scheduler_job_completed"`. Deterministic, fast. Grep-gate added as a separate CI assertion. Runtime sample captured in `evidence.md` from a real `start-api.ps1` run. |
| 4 | Test file (`scheduler_test.go:188`) currently mutates private `s.logger` after `New` — migrate it inside this feature, or leave? | **Migrate.** Same file, drive-by per CLAUDE.md §5.3 (orphans from own change). Pass `slog.NewJSONHandler(io.Discard, nil)` through `New`; delete private-field mutation. Keeps grep-gate simple (no `--exclude` flag). |
| 5 | Enrich with `slog.With("component","scheduler")` at the call site? | **No.** Schema-adjacent; would tip into log-schema redesign (HS-2 rabbit hole listed in M2 milestone.md). Out of scope. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):**
  - **Primary — SRE / operator tailing `apps/api` stdout** (and any downstream log shipper: Loki,
    Datadog, CloudWatch Logs Insights, `jq` on the captured stream). Today every other line in the
    process is JSON; scheduler lines are not, so a `job=session_sweeper` filter against the JSON
    pipeline silently drops scheduler runs.
  - **Secondary — composition root** [`apps/api/cmd/metaldocs-api/main.go:525`](../../../../../apps/api/cmd/metaldocs-api/main.go).
    It already constructs `slog.SetDefault(slog.NewJSONHandler(os.Stdout, nil))` at line 105;
    expects to *inject* the resulting `slog.Default()` into the scheduler, not have the scheduler
    construct a competing handler.

- **Contract:**
  1. `jobscheduler.New` accepts a non-nil `*slog.Logger` and uses **exactly that logger** (no
     wrapping, no schema mutation) for every existing log call inside the scheduler package.
  2. With the API process's already-installed JSON default passed in, every scheduler log line on
     stdout is **one JSON object per line**, line-delimited, with the existing keys preserved:
     `time`, `level`, `msg` (e.g. `scheduler_job_completed`, `scheduler_job_failed`,
     `scheduler_job_skipped`, `scheduler_heartbeat_failed`, `scheduler_lease_held_by_other`,
     `scheduler_lease_stolen`, `scheduler_acquire_lease_failed`, `scheduler_release_lease_failed`,
     `scheduler_invalid_job_config`, `scheduler_loops_shutdown_timeout`,
     `scheduler_backpressure_skip_streak_max`), and the per-call attrs already passed
     (`job`, `epoch`, `duration`, `error`, `reason`, `streak`).
  3. A nil logger is rejected at construction time with `errors.New("logger required")`, mirroring
     the existing empty-`leaderID` check.

- **Source of truth for the contract:**
  - JSON-handler default → [`apps/api/cmd/metaldocs-api/main.go:105`](../../../../../apps/api/cmd/metaldocs-api/main.go)
    (`slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))`).
  - Existing log-key shape → [`internal/modules/jobs/scheduler/scheduler.go:146,169,211,213,221,225,251,254,264,290,294`](../../../../../internal/modules/jobs/scheduler/scheduler.go) (all current `s.logger.*` call sites — keys are preserved exactly).
  - Mission defect anchor → `mission.md` §5 D1.

## What this feature implements

The scheduler stops constructing its own logger and instead receives one through its constructor.
Concrete change, scoped to four lines of source + two tests:

1. **`internal/modules/jobs/scheduler/scheduler.go`** — `New` signature becomes
   `New(db *sql.DB, leaderID string, logger *slog.Logger) (*Scheduler, error)`. Nil logger →
   `errors.New("logger required")`. Hardcoded `slog.New(slog.NewTextHandler(os.Stdout, nil))`
   literal at line 131 deleted; the injected `logger` is assigned to the `logger` struct field.
   Unused `os` import removed if no other reference remains.
2. **`apps/api/cmd/metaldocs-api/main.go:525`** — call site becomes
   `jobscheduler.New(deps.SQLDB, leaderID, slog.Default())`.
3. **`internal/modules/jobs/scheduler/scheduler_test.go`** — every `jobscheduler.New(...)` call
   site updated to pass `slog.New(slog.NewJSONHandler(io.Discard, nil))`. The private-field
   override at line 188 (`s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))`) is deleted —
   the constructor already takes the discard logger.
4. **New failing test** `TestScheduler_LoggerEmitsJSON` added to `scheduler_test.go` per the
   Validation Gate row 1 below.

## Non-goals (mandatory)

Explicitly out of scope. Anything here that later appears in the diff is scope drift (validator C6).

- **No functional-options or Config-struct refactor of `New`.** Other hardcoded knobs
  (`heartbeatEvery`, `drainWait`, `forceWait`, `maxSkipStreak`) stay hardcoded this feature.
- **No log-schema redesign.** No new keys, no field renames, no level changes, no
  `slog.With("component", "scheduler")` enrichment at the call site (HS-2 rabbit hole).
- **No worker / jobs binary changes.** `apps/worker` and `apps/jobs` already JSON-set their
  default; scheduler runs in `apps/api`.
- **No touch outside `internal/modules/jobs/scheduler/` + the single call site** (composition
  root). No adjacent refactor of other `jobs/` packages (CLAUDE.md §5.3 surgical rule).
- **No `slog.SetDefault` change.** It is already correct at `main.go:105`.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| 1. With a JSON handler passed through `New`, a scheduled job run emits exactly one line that decodes as a JSON object with `msg == "scheduler_job_completed"` and `job == <registered name>`. | `go test ./internal/modules/jobs/scheduler/ -run TestScheduler_LoggerEmitsJSON -count=1` | fixture (JSON-handler over `&bytes.Buffer`) |
| 2. Nil logger is rejected at construction. | `go test ./internal/modules/jobs/scheduler/ -run TestScheduler_New_RejectsNilLogger -count=1` (asserts `errors.New("logger required")`) | fixture |
| 3. No hardcoded text-handler construction remains in `internal/modules/jobs/`. | `grep -RIn 'NewTextHandler' internal/modules/jobs/` returns 0 lines (exit 1) | real source-tree gate |
| 4. Existing scheduler test suite unaffected by the signature change. | `go test ./internal/modules/jobs/scheduler/...` exits 0 | fixture |
| 5. Whole-repo regression. | `go test ./...` exits 0 | fixture |
| 6. Real-provider runtime proof. | `.\scripts\start-api.ps1` started; stdout captured; one line containing `"msg":"scheduler_job_completed"` and `"job":"<name>"` decodes as valid JSON (`jq .` over the line exits 0). Sample line pasted verbatim into `evidence.md` with the job name and timestamp. | **real-provider** (per mission §8 label) |
| 7. R3 mitigation (M2 milestone.md) — no scheduled job goes silent under the new logger. | `evidence.md` records one observed log line per registered scheduled job over a single tick window of the same `start-api.ps1` run; any silent job triggers HS-3 before close. | real-provider |

> TDD: rows 1 and 2 are the failing tests, written first. Row 3 (grep gate) is asserted before
> the constructor edit. Rows 4–5 are the regression guard. Rows 6–7 are runtime evidence.

## ADR needed?

- [x] No durable decision — skip. Mechanical wiring change; the durable architecture statement
      ("composition-root injection, not in-package wiring") is already recorded in
      `milestone.md` §Dependencies & constraints and in `mission.md` §5 D1.
- [ ] Durable decision made → record an ADR under `wiki/decisions/` and link it here.
