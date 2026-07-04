# F5.4 — outbox retention (plan)

> Executes `spec.md` / contract §4. Engine: `superpowers:subagent-driven-development` (fresh subagent
> per task; sonnet implement+review; main session reviews + commits).

## Load-bearing facts (verified 2026-07-04)

1. **River's own retention defaults already match contract §4.1** (`go doc github.com/riverqueue/river.Config`):
   `CompletedJobRetentionPeriod` defaults to 24h, `CancelledJobRetentionPeriod` defaults to 24h,
   `DiscardedJobRetentionPeriod` defaults to 7 days — identical to the contract's locked values. Setting
   them **explicitly** in `apps/jobs/cmd/metaldocs-jobs/main.go`'s River client config is still required
   (binding per contract, and defends against a future River default change silently drifting the
   policy) — this is config-only, no new code path.
2. **`riverjobs.Config` doesn't yet expose the 3 retention fields** (`internal/platform/jobs/river/client.go:11-16`)
   — needs 3 new fields threaded through to `river.Config` (mirrors how `PeriodicJobs`/`Queues` are
   already threaded).
3. **Periodic-job pattern to mirror:** `internal/modules/jobs/maintenance/periodic.go` — a
   `river.NewPeriodicJob(river.PeriodicInterval(d), argsFn, &river.PeriodicJobOpts{ID, RunOnStart:false})`
   per job, queue `"maintenance"`. The purge job's Args/Worker/periodic-entry should live in a **new**
   package (owning module: `render` — the tables are `render/fanout`'s) analogous to how `dispatchjobs`
   sits under `internal/modules/render/fanout/`.
4. **Only `metaldocs-jobs` needs the new periodic entry in its Config.PeriodicJobs** AND registers the
   Worker (unlike the 3 janitors, which api's Config also lists for leader-election parity — the purge
   job has no api-side enqueue-site equivalent needing that; but check: F5.2's pattern put
   `PeriodicJobs()` in BOTH api and jobs configs so whichever wins leader election enqueues it. For
   consistency and to avoid a single-point-of-enqueue gap if jobs is ever not the leader, add the purge
   job's periodic entry to **both** api's and jobs' `Config.PeriodicJobs`, mirroring the janitor pattern
   exactly — only jobs subscribes `maintenance` + registers the Worker, matching `periodic.go`'s own doc
   comment convention.
5. **Staging tables:** `pdf_dispatch_outbox` / `materialize_dispatch_outbox` — columns `status` (CHECK
   `pending|processing|dispatched|failed`), `dispatched_at`, `dead_lettered_at` (per F5.3's repo:
   `internal/modules/render/fanout/staging_outbox.go`). Purge query: `DELETE ... WHERE status='dispatched'
   AND dispatched_at < now() - interval '7 days'` — a dead-lettered row's status is `'failed'`, never
   `'dispatched'`, so it's excluded by the status filter alone; no extra guard needed. Batch via
   `... LIMIT 5000` in a loop capped at 10 iterations per table per run (so max 50,000 rows/table/run —
   still bounded, not unbounded).

## Task breakdown

### T1 (sonnet) — retention config fields + purge job package
- `internal/platform/jobs/river/client.go`: add `CompletedJobRetentionPeriod`, `CancelledJobRetentionPeriod`,
  `DiscardedJobRetentionPeriod time.Duration` to `Config`; thread into the constructed `river.Config`.
- New package `internal/modules/render/fanout/retention` (or `dispatchjobs` if the reviewer judges it
  tightly enough coupled — your call, document the choice): `PurgeArgs{}` (`Kind()` →
  `"staging-outbox-purge"`), `PurgeWorker` embedding `river.WorkerDefaults[PurgeArgs]`, holding a
  `*sql.DB` (or a narrow repo interface — prefer a repo method so the SQL lives with the other staging
  outbox SQL in `fanout` package, not duplicated). Add a `PurgeDispatched(ctx, cutoff time.Time, batchSize
  int, maxIterations int) (deleted int, err error)` method (or two, one per table, or one parameterized
  by table) to `StagingOutboxRepository` (or a sibling type) implementing the bounded-batch-loop DELETE
  described in fact 5 above, for **both** pdf and materialize tables.
- `PeriodicJobs()`-style helper (new file, e.g. `internal/modules/render/fanout/retention/periodic.go`)
  returning the one `*river.PeriodicJob` (`PeriodicInterval(24*time.Hour)`, `ID:"staging-outbox-purge"`,
  queue `"maintenance"`, `RunOnStart:false`), mirroring `maintenance/periodic.go`'s shape exactly (own
  file, no worker construction, no DB — just schedule/args wiring) so it composes the same way in both
  binaries' Config.
- `go build ./...` green.

### T2 (sonnet) — wire into both binaries
- `apps/jobs/cmd/metaldocs-jobs/main.go`: add the 3 retention fields to the `riverjobs.Config` passed to
  `NewClientBundle` (24h/24h/7×24h — explicit, not relying on defaults per fact 1); append the purge
  periodic job to `PeriodicJobs` (alongside the 3 janitor entries already there from F5.2); register the
  `PurgeWorker` via `river.AddWorker`.
- `apps/api/cmd/metaldocs-api/main.go`: append the purge periodic job to its enqueue-only Config's
  `PeriodicJobs` too (leader-election parity with the 3 janitors — mirror exactly how F5.2 T3 added the
  janitor periodic entries to api's Config without api subscribing `maintenance` or registering workers).
- `go build ./...` green; both binaries build.

### T3 (sonnet) — proofs (testdb, targeted -run)
- Retention integration test (contract §4.2), testdb factory, `//go:build integration`: seed (a) old
  dispatched rows (`dispatched_at` > 7d old), (b) recent dispatched rows, (c) a dead-lettered row, for
  BOTH pdf and materialize tables → run `PurgeWorker.Work` (or the repo purge method directly) once →
  assert (a) gone, (b)+(c) survive. A second test (or the same, extended) seeds > 5000 eligible rows and
  asserts a single `Work()` call deletes at most the bounded cap (5000×10 = 50,000 max; a smaller,
  practical seed count like 5001-6000 rows proving "not all deleted in one unbounded statement, loop
  actually iterates/caps" is sufficient — no need to literally seed 50k+ rows).
- Targeted `-run`; execution deferred to M5-close live drive if no DSN this session (F5.2/F5.3 T6
  precedent) — author + compile + vet regardless; record exact `-run` commands in evidence.

## Test strategy
TDD where it bites: T1's repo purge method gets a focused unit/integration test for the batch-cap logic;
T3 is the acceptance proof. Build-green after every task. Commit per task, never push.

## Files touched (census)
NEW `internal/modules/render/fanout/retention/*` (Args, Worker, periodic.go) or equivalent home decided
in T1; MODIFY `internal/platform/jobs/river/client.go` (3 new Config fields), `internal/modules/render/fanout/staging_outbox.go`
(purge method), `apps/jobs/cmd/metaldocs-jobs/main.go`, `apps/api/cmd/metaldocs-api/main.go`. NEW
integration test(s).
