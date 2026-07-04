# ADR 0067 — Async job infrastructure consolidated onto River (lease scheduler + staging poller retired; H-PRE-1 retired)

> **Status:** Accepted
> **Date:** 2026-07-04
> **Program:** global-maximum-remediation · **Milestone:** M5 (async consolidation onto River), feature F5.1
> **Supersedes/retires:** the custom Postgres-lease scheduler (`metaldocs.job_leases` + `acquire_lease`/`heartbeat_lease`/`release_lease`); **H-PRE-1** advisory-lock constraint (retired, §H-PRE-1).

## Context

The 2026-07-03 architecture review (dimension 5) and M5's `developing-new-work` gate
(`docs/superpowers/analysis/2026-07-04-async-river-consolidation-system-impact.md`, verdict Yellow)
found **three parallel async job infrastructures** in one system, verified in code 2026-07-04:

1. **River v0.37.1** — `metaldocs-jobs` binary; scheduled-publish cutover + notifications fanout;
   transactional enqueue via `InsertTx` on the business tx (outbox pattern by construction).
   (`apps/jobs/cmd/metaldocs-jobs/main.go:37-54`; client `internal/platform/jobs/river/client.go:23-46`.)
2. **Custom Postgres-lease ticker scheduler** — `metaldocs-api` binary; 4 janitors
   (stuck-instance-watchdog 5 min, idempotency-janitor 15 min, audit-integrity-validator 1 h,
   lease-reaper 10 min) each acquiring `metaldocs.acquire_lease(...)` with heartbeat + pressure-probe
   skip. (`internal/modules/jobs/scheduler/scheduler.go:117-273,413-445`; lease fn
   `db/baseline/0001_current_schema.sql:59-99`; `job_leases` `:1354-1361`; reg `main.go:599-617,1005-1035`.)
3. **StagingOutboxWorker poller** — `metaldocs-api` binary; pdf + materialize dispatch;
   `ClaimPending` (`FOR UPDATE SKIP LOCKED`) + exponential backoff **duplicated in two branches**.
   (`internal/modules/render/fanout/staging_outbox_worker.go:63-119`; repo `staging_outbox.go:49-131`.)

River **already runs in this system** and natively ships every primitive the other two hand-roll
(re-proven against River v0.37.1 docs, 2026-07-04):

- **Leader election** — `river_leader` table (migration line 7); client `ID` is "used for leader
  election"; periodic-job insertion is a leader-only maintenance service ⇒ **singleton cluster-wide**.
- **Periodic jobs** — `Config.PeriodicJobs` / `river.NewPeriodicJob(river.PeriodicInterval(d), argsFn,
  &river.PeriodicJobOpts{ID, RunOnStart})`; uniqueness belt (`UniqueSkippedAsDuplicate`).
- **Native retention** — `CompletedJobRetentionPeriod` / `Cancelled` / `Discarded` job cleaner for
  River's own rows (`-1` = keep forever).
- **Transactional enqueue** — `InsertTx(ctx, tx, args, nil)`.

Per the frameworks-catalog rule ("a hand-rolled equivalent of a platform framework is a defect — it
bypasses the invariant the framework exists to enforce"), the lease scheduler and the staging backoff
are **local maxima**: three retry/idempotency/election code paths maintained where **one trusted
primitive already runs**. The stuck-instance-watchdog's belt-and-suspenders (advisory lock **plus**
lease) and the **H-PRE-1** deadlock constraint are *symptoms* of not having one trusted scheduling
primitive.

Two correctness debts ride along (review dimension 5 DEBT): the staging outbox tables are **never
purged** (`MarkDispatched` only flips status — unbounded growth: `staging_outbox.go:112-131`), and the
lifecycle fanout has **no ordering guarantee** between River jobs (`fanout_worker.go:47-138`).

## Decision

**River is MetalDocs' single async job infrastructure.** The custom lease scheduler and the staging
poller are retired; both concerns move onto River.

### 1. Janitors → River periodic jobs (F5.2)
The 3 surviving janitors become **River periodic jobs** (`NewPeriodicJob` + `PeriodicInterval` + stable
`PeriodicJobOpts.ID`), leader-elected (singleton by River's elector), with the **same intervals and the
same job bodies** (behavior-preserving migration; each has an integration proof):

| Periodic job | Interval | Body (unchanged) |
|---|---|---|
| stuck-instance-watchdog | 5 min | auto-cancel/emit for `in_progress` instances `submitted_at < now()-7d` |
| idempotency-janitor | 15 min | delete expired `idempotency_keys`; count orphaned in_flight |
| audit-integrity-validator | 1 h | `ValidateIntegrity` hash-chain recompute (published `audit` port, unchanged) |

**lease-reaper is deleted, not migrated** — it existed solely to GC the custom `job_leases` table.
River owns its own leader/lease lifecycle in `river_leader`; there is nothing left to reap.

The pressure-probe "skip on DB pressure" behavior is **not** re-implemented by hand — River's own
fetch/queue configuration (`FetchCooldown`, `MaxWorkers`, queue separation) is the platform mechanism.
If a maintenance queue needs isolation from business jobs, it gets its own River queue with a low
`MaxWorkers`, rather than a bespoke connection-ratio probe. (Behavior-equivalence, not line-equivalence:
the *guarantee* — maintenance never starves business traffic — is preserved via River's native levers;
the hand-rolled probe is retired. Recorded as an intentional behavior substitution in the validation
contract.)

### 2. Staging dispatch → River transactional job (F5.3)
`Enqueue` **stays** in the caller's business tx (`ON CONFLICT (tenant_id, revision_id) DO NOTHING`
idempotency preserved). The **dispatch** moves from the hand-rolled poller to a River job (transactional
enqueue via `InsertTx`, one worker per dispatch kind). The **duplicated exponential-backoff math is
deleted** — River's `MaxAttempts` + `RetryPolicy` own retry/backoff; failed terminal jobs are River's
discarded state (the inspectable DLQ, satisfying the async DLQ requirement). Per-message `SeedTxTenant`
before every tenant-scoped write **stays** so FORCE-RLS backstops the async path (M3 F3.2).

### 3. Outbox retention (F5.4)
- **River's own rows:** configure `CompletedJobRetentionPeriod` (and Cancelled/Discarded) — bounded by
  the platform cleaner, no code.
- **Staging business rows** (`materialize_dispatch_outbox`, `pdf_dispatch_outbox`): a **River periodic
  purge job** DELETEs `status='dispatched' AND dispatched_at < now() - <retention>` in bounded batches,
  plus a dead-letter policy for `dead_lettered_at`. Retention value fixed in the M5 validation contract.
  Growth is now bounded; the SKIP-LOCKED scan cost stops rising.

### 4. Fanout ordering — proven idempotent-commutative (F5.5)
Strict cross-job ordering is **not** imposed. The fanout is **order-independent by construction**: each
lifecycle event fans out to **additive, per-recipient-per-event notification rows**
(`INSERT ... ON CONFLICT (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL DO
NOTHING`, `fanout_worker.go:123-136`). There is **no shared mutable "current status" row** a later event
could clobber — a `published` fanout and a `superseded` fanout write **distinct** rows keyed by their own
`source_event_id`. Therefore:

- **Terminal state = the set of notification rows = union of per-event fanouts**, which is commutative
  and idempotent under redelivery and reordering.
- No lost or inverted terminal state is possible from job reordering.

This is proven by a **concurrent race test** (both interleavings of `published`/`superseded` for one
document) asserting the terminal notification-row set is identical regardless of order, plus the formal
commutativity argument above. (Should a future event type introduce a shared mutable projection, ordering
must be revisited — recorded as the trigger in the validation contract.)

### 5. Deployment topology
The janitor periodic jobs and the staging dispatch workers are hosted in **`metaldocs-jobs`** — the
binary that already runs the River client, elector, and workers. `metaldocs-jobs` is **already a
required binary** (scheduled-publish + notifications run there), so this adds **no new required binary**.
Consequence: **`metaldocs-api` becomes a truly stateless sync + authz service** — it no longer hosts a
scheduler or async workers. (This corrects the CLAUDE.md "api also hosts the 4 leader-elected janitors"
wording; the correction is already scoped to M9 F9.4 doc-truth and is applied there, not silently here.)

### H-PRE-1 — **RETIRED** (formal disposition)
**H-PRE-1** ("never call an authz-recording read inside a lock-holding atomic tx; hoist it off-tx")
existed because the stuck-instance-watchdog took a **`pg_try_advisory_lock`** (`job.go:114`) to prevent
concurrent runs across replicas, creating a lock-holding window that must not enclose an authz-recording
read. Under River:

- The periodic job is **inserted once cluster-wide by the elected leader** and **dequeued once** by a
  single worker (River's queue guarantees single execution). The advisory lock's sole purpose —
  single-runner mutual exclusion — is **subsumed by River's elector + queue**.
- The `pg_try_advisory_lock` is therefore **removed**. With no application-held advisory lock in the
  janitor, **the H-PRE-1 hazard no longer exists** — the constraint is **retired**, not merely satisfied.

**Evidence required (F5.2):** an integration test that runs two River clients against one database and
asserts the watchdog periodic job **executes exactly once** per tick (singleton via elector), with the
advisory lock removed. Retirement is contingent on that proof; until it is green, the constraint is not
declared retired.

## Consequences

- **~3 duplicated code paths deleted permanently:** the lease scheduler (`scheduler.go`,
  `lease_reaper.go`), the `job_leases` table + 3 lease SQL functions, the watchdog advisory lock, and the
  duplicated staging backoff math. One primitive (River) owns election, scheduling, retry, retention, DLQ.
- **The transactional-outbox invariant is preserved and strengthened** — every enqueue stays `InsertTx`
  on the business tx; consumers stay idempotent; retention is added.
- **RLS backstop preserved** — migrated tenant-scoped jobs seed `SeedTxTenant`/`SeedTxIdentity` per
  message/run (M3 F3.2).
- **H-PRE-1 retired with evidence** — the memory note and `invariant-checklist.md:58` are updated to
  reflect retirement (M5 disposition); the advisory-lock class is gone.
- **Migration ordering (destructive, expand/contract):** (a) add River periodic jobs + staging River
  jobs + purge job, verify both run; (b) delete the lease scheduler code + watchdog advisory lock;
  (c) drop `job_leases` + the 3 lease functions in a forward-only migration ordered **after** the code
  that writes them is gone. Never drop the lease objects while a running api still writes them.
- **No wire-contract change** — M5 touches no `openapi.yaml`; no route/capability added.
- **Topology change** — deployments that ran only `metaldocs-api` for maintenance must run
  `metaldocs-jobs` (already true for scheduled-publish/notifications; documented in DEPLOY/compose).

## Alternatives considered

- **Keep three infrastructures, patch the two debts in place** (add a lease-scheduler purge job + a
  hand-rolled ordering key). Rejected: optimizes *inside* the patch (locks in the local maximum), leaves
  three election/retry code paths and the H-PRE-1 class standing. Fails the Global-Maximum rule.
- **Migrate janitors onto River but leave staging on the poller.** Rejected: still two dispatch/backoff
  implementations; the staging duplicated-backoff debt survives. Half-consolidation.
- **Host the janitors in `metaldocs-api`** (a River client with periodic jobs inside api). Rejected:
  keeps async work in the binary the target architecture wants stateless; splits River across two binaries
  for no benefit. `metaldocs-jobs` already owns River.

## References

- M5 gate: `docs/superpowers/analysis/2026-07-04-async-river-consolidation-system-impact.md` (Yellow; 10 locked constraints)
- 2026-07-03 review §5 (async): `docs/superpowers/analysis/2026-07-03-final-architecture-global-maximum-review.md`
- Mission: `docs/superpowers/milestones/global-maximum-remediation/mission.md` §7 M5, §9 (HS-8 gate cleared: Yellow)
- River v0.37.1 native capabilities (periodic jobs, `river_leader` elector, retention, `InsertTx`) — River docs, verified 2026-07-04
- Retired: lease scheduler `internal/modules/jobs/scheduler/scheduler.go`; `db/baseline/0001_current_schema.sql:59-99,1354-1361`
- Staging: `internal/modules/render/fanout/staging_outbox_worker.go`, `staging_outbox.go`
- Fanout idempotency: `internal/modules/notifications/infrastructure/fanout_worker.go:123-136`
- Async requirements: `wiki/architecture/backend-target-architecture.md:250-254` — **REQ-ASYNC-1** (transactional outbox), **REQ-ASYNC-2** (idempotent consumer), **REQ-ASYNC-3** (backoff+jitter+cap → inspectable DLQ), **REQ-ASYNC-4** (watchdog for stuck work + queue-depth/oldest-item metric), **REQ-ASYNC-5** (retry ownership declared once per pipeline), **REQ-ASYNC-6** (background jobs use the fail-closed bypass bridge, never a synthetic HTTP principal) — all preserved by this consolidation
- H-PRE-1 origin: `wiki/decisions/` advisory-lock constraint; memory `advisory-lock-deadlock-constraint`; `invariant-checklist.md:58`
