# Feature F5.4 — Evidence — outbox retention

> **Milestone:** 5 · **Feature:** `f5.4-outbox-retention` · **Closed:** 2026-07-04
> **Contract:** `spec.md` (distills `../validation-contract.md` §4). Validation Gate proved below.
> Engine: `superpowers:subagent-driven-development` — fresh subagent per task (T1–T3), sonnet
> implement+review; main session reviewed + committed.

## What was implemented

River-native job-row retention (config only) + a new periodic purge job for the two staging
dispatch tables (`pdf_dispatch_outbox`, `materialize_dispatch_outbox`), bounded-batch, 7-day cutoff,
dead-lettered rows excluded by construction.

- **T1** `d9ff68d3` — `riverjobs.Config` gained 3 retention fields (forwarded verbatim to
  `river.Config`; zero-value = River's own documented default, no special-casing needed — confirmed
  via `go doc`); `StagingOutboxRepository.PurgeDispatched(ctx, cutoff, batchSize, maxIterations)`
  added (bounded `ctid IN (SELECT ctid ... LIMIT $2)` loop, matching the house-style bounded-batch
  idiom already used by `idempotency_janitor/job.go:56-78`); new `render/fanout/retention` package
  (`PurgeArgs`, `PurgeWorker`, binding constants `RetentionPeriod=7×24h`/`BatchSize=5000`/
  `MaxIterations=10`, `PeriodicJob()` mirroring `maintenance/periodic.go`'s shape).
- **T2** `91445c90` — `metaldocs-jobs` set the 3 retention values explicitly (24h/24h/7×24h) on its
  actual `riverjobs.Config` construction site (`internal/platform/bootstrap/jobs.go` — found by
  reading, not assumed to be in `main.go`); registered `PurgeWorker`; merged `retention.PeriodicJob()`
  into its periodic-jobs slice. `metaldocs-api` appended the same periodic-job definition to its
  enqueue-only Config for leader-election parity (mirrors the F5.2 janitor dual-define pattern
  exactly) — no `maintenance` queue subscription, no `PurgeWorker` registration, retention fields
  left at zero (harmless: same defaults apply if api's own cleaner ever runs).
- **T3** `40c9ec52` — 3 real-DB integration tests (`retention_integration_test.go`, testdb factory,
  `//go:build integration`, mirrors F5.3's `dispatch_integration_test.go` convention): purge-only-old
  equivalence (both tables, table-driven), `PurgeWorker.Work` end-to-end (both repos via the real
  River entrypoint), batch-bound proof (10 seeded rows, `batchSize=3/maxIterations=2` ⇒ exactly 6
  deleted, 4 survive — proves the cap mechanism without needing a 5000-row fixture).
  **Execution DEFERRED** (no DB DSN this session, same precedent as F5.2/F5.3 T6).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|--------------------|-----------------|
| Build (all 4 binaries) | `go build ./...` | exit 0 | real |
| Vet | `go vet ./...` | exit 0 | real |
| Unit tests (T1/T2 packages) | `go test ./internal/modules/render/fanout/... ./internal/platform/jobs/river/... ./apps/...` | all `ok` | real |
| T3 integration tests — compile/vet | `go build -tags=integration ./...`, `go vet -tags=integration ./...` | clean, no output | real |
| T3 integration tests — attempted run | `go test -tags=integration -run 'TestRetentionPurge' ./internal/modules/render/fanout/retention/... -v` | all 3 `SKIP`: `DATABASE_URL/METALDOCS_DATABASE_URL not set` — expected precedent failure mode, not a logic/compile bug | real (pending run) |

T3 exact `-run` commands for later live execution (M5 close):
```
go test -tags=integration -run TestRetentionPurge_Integration_PurgesOnlyOldDispatchedRows ./internal/modules/render/fanout/retention/... -v
go test -tags=integration -run TestRetentionPurge_Integration_PurgeWorkerRunsBothTables ./internal/modules/render/fanout/retention/... -v
go test -tags=integration -run TestRetentionPurge_Integration_BatchBounded_CapsAtBatchSizeTimesMaxIterations ./internal/modules/render/fanout/retention/... -v
```

## Acceptance vs spec Validation Gate

| Acceptance criterion (spec.md) | Met? | Evidence |
|---------------------------------|------|----------|
| 1. Retention integration proof (seed old/recent/dead-lettered → purge → correct survivors, both tables) | authored; **run deferred** | T3 row; runs at M5 close |
| 2. Batch-bounded proof (loop caps, not unbounded) | authored; **run deferred** | T3's `BatchBounded` test (10 rows, 3×2 cap ⇒ 6 deleted); mechanism-equivalent proof at smaller scale |
| 3. River native retention config set (24h/24h/7×24h on jobs binary) | **yes** | `bootstrap/jobs.go` diff (T2), `go doc` confirmation these match River's own field semantics |
| 4. `go build ./...` green, all 4 binaries | **yes** | exit 0 throughout T1–T3 |
| 5. Section-by-section match to contract §4, no divergence | **yes** | values (7d/5000/10/24h/24h/7×24h) taken verbatim from contract; no HS-7 this feature |

## Review disposition

- Spec-compliance + code-quality review (per-task, main session read every landed diff directly via
  `git show` before approval): T1 (batch-SQL house-style match, zero-value handling), T2 (correctly
  located the real Config construction site in `bootstrap/jobs.go` rather than assuming `main.go`;
  leader-election-parity pattern matched F5.2 exactly), T3 (test conventions matched F5.3's
  integration-test template) — all 3 tasks approved first pass, no re-review loops needed.
- No subagent-dispatch anomaly this feature (unlike F5.3's T3/T6) — all 3 tasks landed on the first
  dispatch with concrete, verified evidence.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|--------------------------|
| T3 retention + batch-bound integration tests not yet **executed** | Compiled + vetted; blocked only by missing DB DSN this authoring session (F5.2/F5.3 T6 precedent). No green fabricated — `SKIP` output quoted above. | **M5 close live QA drive** (task #7): run the 3 `-run` commands above against real Postgres via `.\scripts\start-api.ps1 -Build` path. A failure here is an HS-4 (validator FAIL). Owner: main session at M5 close. |
| Dead-lettered `*_dispatch_outbox` row pruning | Kept inspectable per REQ-ASYNC-3; contract §8 bounded defer, not this feature's scope | M8 ops-readiness, or when a DLQ-retention policy is set |
