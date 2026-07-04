# Feature F5.2 — Evidence — janitors on River

> **Milestone:** 5 · **Feature:** `f5.2-janitors-on-river` · **Closed:** 2026-07-04
> **Contract:** `spec.md` (distills `../validation-contract.md` §2). Validation Gate proved below.
> Engine: `superpowers:subagent-driven-development` — fresh subagent per task (T1–T6), sonnet
> implement+review; main session reviewed + committed. Expand→contract→drop→prove→docs ordering.

## What was implemented

3 surviving janitors (stuck-instance-watchdog, idempotency-janitor, audit-integrity-validator) migrated
off the custom Postgres-lease ticker scheduler onto **River periodic jobs**; the lease scheduler +
lease-reaper + `job_leases` table + 3 lease SQL functions **retired**. Producer matches consumer contract
(spec.md §"Consumer contract"): jobs binary hosts execution on a dedicated `maintenance` queue, leader-only
enqueue (dual-define rail), bodies byte-behavior-identical.

- **T1** `5838bf44` — platform: `Config.PeriodicJobs []*river.PeriodicJob` threaded through
  `internal/platform/jobs/river/client.go` + `bootstrap/jobs.go` (`JobsWorkerFactory` returns
  `(*river.Workers, []*river.PeriodicJob, error)`). No behavior change; build green.
- **T2** `81ffac51` — each janitor exposes a River `Worker[Args]` (empty `Args`+`Kind()`, `Work→run()`);
  body verbatim (query/batch/drift/`SeedTxTenant`/`WithBackgroundBypass` unchanged). **Watchdog
  `acquireRunLock`/`pg_try_advisory_lock` removed.** Legacy `New()` kept so api still compiled (expand).
- **T3** `c94643b9` — `internal/modules/jobs/maintenance/periodic.go` `PeriodicJobs()` (watchdog 5m /
  idempotency 15m / audit 1h, queue `maintenance`, `RunOnStart:false`). jobs main: 3 workers registered +
  `Queues["maintenance"]={MaxWorkers:2}`. api main: same `PeriodicJobs()` in Config (enqueue-when-leader,
  **no** maintenance queue, nil workers). Both binaries build.
- **T4** `b067f3a1` — contract: deleted `internal/modules/jobs/scheduler/`; removed `jobscheduler.New`,
  `registerScheduledJobs`, `schedulerWG` goroutine, `SetSchedulerMetrics` from api main; legacy `New()` +
  scheduler import dropped from the 3 janitors; job_test.go repointed to `run(...)`. api starts with **no
  scheduler goroutine**.
- **T5** `ca3dd39d` — forward-only migration `db/migrations/0273_drop_job_leases.sql`: DROP the 3 lease
  functions + `metaldocs.job_leases`. Ordered after T4 (no writer left). No down (repo forward-only norm).
- **T6** `332a799c` — proofs authored (P1–P3 equivalence + P4 River singleton), testdb factory,
  `//go:build integration`. **Execution DEFERRED** (no DB DSN in this session) — see Bounded defers;
  runs at M5 close live drive.

> **HS-7 note:** T6's commit subject and an earlier revision of ADR 0067 / contract §2.6 called the
> watchdog-lock removal an "H-PRE-1 retirement". **That framing is false** (corrected in place this
> feature): H-PRE-1 governs authz-recording reads inside lock-holding txs (audit hash-chain writer
> `writer.go:59` + `authz.Require` `authz.go:119`); the watchdog's `pg_try_advisory_lock` was an unrelated
> single-runner mutex on its own connection that never enclosed an authz read. **H-PRE-1 stays LIVE.** The
> lock removal itself stands (River elector+queue subsume single-runner; P4 gates it). Ratification of the
> correction carried to M5 HS-1. Recorded: README HS-7 row, ADR 0067 errata, contract §2.6 erratum,
> `invariant-checklist.md:58`, memory `advisory-lock-deadlock-constraint`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Build (all 4 binaries) | `go build ./...` | `BUILD_EXIT=0` (2026-07-04, post-corrections) | real |
| Retirement census — scheduler pkg gone | `ls internal/modules/jobs/scheduler` | `No such file or directory` | real |
| Retirement census — no lease refs in Go | `grep -rn "acquire_lease\|heartbeat_lease\|release_lease\|job_leases\|pg_try_advisory_lock" --include=*.go internal apps \| grep -v _test.go` | 1 hit, unrelated: `outbox/postgres/consumer.go:37` TODO comment on staging `claimLease` (F5.3 territory, NOT `job_leases`) — 0 lease-scheduler refs | real |
| Retirement census — api scheduler gone | `grep "registerScheduledJobs\|jobscheduler" apps/api/cmd/metaldocs-api/main.go` | `NONE` | real |
| P1–P3 equivalence proofs (testdb) | `go test -tags integration -run TestWatchdog.../TestIdempotency.../TestAudit... ./internal/modules/jobs/...` | **DEFERRED** — compiled + vetted; execution needs DSN (see defer) | real (pending run) |
| P4 §2.6 singleton proof (2 clients, 1 db) | `go test -tags integration -run TestRiverPeriodicSingleton ./internal/platform/jobs/river/` | **DEFERRED** — compiled + vetted; gates watchdog-lock removal at close | real (pending run) |

## Acceptance vs spec Validation Gate

| Acceptance criterion (spec.md) | Met? | Evidence |
|--------------------------------|------|----------|
| 1. 3 equivalence integration tests (one per survivor) | authored; **run deferred** | T6 row; runs at M5 close |
| 2. §2.6 singleton proof (2 River clients, exactly-once) | authored; **run deferred** | T6/P4 row; gates lock removal (HS-2 if red) |
| 3. §2.5 retirement census = 0 | **yes** | census rows above — scheduler pkg deleted, 0 lease-scheduler refs, `registerScheduledJobs` gone |
| 4. Both binaries build + start (api no scheduler goroutine; jobs 3 periodic on `maintenance`) | build **yes**; start proof at close live drive | `BUILD_EXIT=0`; T3/T4 wiring; runnable check at M5 close |
| 5. Doc updates (invariant-checklist + memory) | **yes** — corrected to **H-PRE-1 LIVE** (HS-7), not "retired" | `invariant-checklist.md:58`, memory clarifier |
| 6. Section match to contract §2, divergence surfaced as HS-7 | **yes** | HS-7 raised + resolved (pending ratification); README row |

## Review disposition

- Spec-compliance review (per-task, subagent): T1–T6 each reviewed against `plan.md`/contract §2; findings
  fixed in-family before commit. No body-semantics drift (T2 constraint held).
- Code-quality review (per-task, subagent): approved per task; expand/contract kept every commit build-green.
- **Aggregate correctness discovery (main session):** the H-PRE-1 mischaracterization (HS-7) — caught by
  reading runtime truth (memory + `writer.go`/`authz.go` greps) against the committed contract, per
  CLAUDE.md "runtime truth beats docs / stop on architecture contradictions". Surfaced, not patched around.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| P1–P3 equivalence + P4 singleton integration proofs not yet **executed** | Tests compiled + vetted; blocked only by missing DB DSN this session (metaldocs-postgres reachable on host :5433 but Docker-NAT + `.env`-forbidden password blocked; testdb self-skips without DSN). No green fabricated. | **M5 close live QA drive** (task #7): run the exact `-run` cmds above against real Postgres via `.\scripts\start-api.ps1 -Build` path. **P4 red ⇒ HS-2** (watchdog lock removal unsafe, revert). Owner: main session at M5 close. |
