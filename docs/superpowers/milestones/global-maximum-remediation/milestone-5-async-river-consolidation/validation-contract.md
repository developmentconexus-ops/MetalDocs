# M5 validation contract (D4 — binding, authored before implementation)

> **Program:** global-maximum-remediation · **Milestone:** M5 (async consolidation onto River)
> **Authored:** 2026-07-04, **before any implementation** (mission D4). Committed before the first code
> change.
> **Binding:** the `milestone-validator` checks the implementation against this document **section by
> section**. Any divergence between shipped code and this contract is **HS-7**: fix the code to the
> contract, or re-open this contract **with operator approval** — never silently edit the contract to
> match the code (mission §9 HS-7). The load-bearing clauses are the **§2 per-job migration table**
> (every janitor: schedule + idempotency key + failure behavior + post-migration equivalence), the
> **§2.5 lease-scheduler retirement census**, the **§2.6 watchdog single-runner-lock removal + River
> singleton proof** (see the §2.6 HS-7 erratum — does NOT retire H-PRE-1), the **§3 staging
> dispatch idempotency + backoff-deletion**, the **§4 retention policy + bounded-purge proof**, and the
> **§5 fanout commutativity proof**.
>
> **Design rails locked before this contract (D7):** gate verdict 🟡 Yellow
> (`../../../analysis/2026-07-04-async-river-consolidation-system-impact.md`, cd2bceb3) + **ADR 0067**
> Accepted (`wiki/decisions/0067-async-job-infrastructure-consolidated-onto-river.md`, 5eb270c3). This
> contract is the concrete enumeration ADR 0067's decision implies.

---

## 0. Runtime-truth basis (the facts this contract is built on)

All claims traced to source at authoring time (2026-07-04; agent code-map + River v0.37.1 docs). Runtime
truth beats docs (CLAUDE.md).

### 0.1 The three infrastructures today

| Infra | Binary | Primitive | Anchor |
|---|---|---|---|
| River v0.37.1 (scheduled-publish + notifications fanout) | `metaldocs-jobs` | `InsertTx` on business tx; `AddWorker` | `apps/jobs/cmd/metaldocs-jobs/main.go:37-54,43-44`; `internal/platform/jobs/river/client.go:23-46` |
| Custom Postgres-lease ticker scheduler (4 janitors) | `metaldocs-api` | `metaldocs.acquire_lease/heartbeat_lease/release_lease` + `job_leases`; per-job ticker; pressure-probe | `internal/modules/jobs/scheduler/scheduler.go:117-273,413-445`; lease fn `db/baseline/0001_current_schema.sql:59-99`; table `:1354-1361`; reg `apps/api/cmd/metaldocs-api/main.go:599-617,1005-1035` |
| StagingOutboxWorker poller (pdf + materialize dispatch) | `metaldocs-api` | poll loop + `ClaimPending` (`FOR UPDATE SKIP LOCKED`) + duplicated backoff | `internal/modules/render/fanout/staging_outbox_worker.go:63-119`; repo `staging_outbox.go:49-131`; tables `db/baseline/0001_current_schema.sql:1368-1382,1484-1498` |

### 0.2 River native capabilities (re-proven, River v0.37.1)

- **Periodic jobs** — `Config.PeriodicJobs` / `river.NewPeriodicJob(river.PeriodicInterval(d), argsFn, &river.PeriodicJobOpts{ID, RunOnStart})`.
- **Leader election** — `river_leader` table (migration line 7); client `ID` "used for leader election"; periodic-job insertion is leader-only ⇒ **singleton cluster-wide**.
- **Uniqueness belt** — unique opts / `UniqueSkippedAsDuplicate` prevent duplicate concurrent insert.
- **Native retention** — `CompletedJobRetentionPeriod` / `CancelledJobRetentionPeriod` / `DiscardedJobRetentionPeriod` cleaner for River's own rows (`-1` = keep forever).
- **Transactional enqueue** — `InsertTx(ctx, tx, args, nil)`.

### 0.3 Invariants that STAY the last line (do not move to the app)

- **Transactional outbox** (REQ-ASYNC-1): enqueue in the business tx via `InsertTx`; network call only in the idempotent consumer. **Preserved, not moved.**
- **Idempotent consumers** (REQ-ASYNC-2): `ON CONFLICT` dedup on staging (`(tenant_id, revision_id)`) and notifications (`(recipient_user_id, source_event_id)`). **Preserved.**
- **M3 FORCE-RLS backstop** (F3.2): per-message/run `authz.SeedTxTenant`/`SeedTxIdentity` before any tenant-scoped write. **Preserved on every migrated job.**
- **M2 write-tripwire** (`enforce_capability_asserted`) — janitor system paths do not assert caps; unchanged.
- **DB status CHECK + FORCE-RLS on the staging tables** — unchanged; retention is a bounded DELETE, no new trigger.
- **REQ-ASYNC-6** background-job authz uses the fail-closed bypass bridge, never a synthetic HTTP principal — unchanged (janitors are system paths).

---

## 1. Post-migration target architecture (binding shape)

**River is the single async job infrastructure.** After M5:

- **Host binary:** all periodic janitors + the staging-dispatch workers + the retention purge job run in
  **`metaldocs-jobs`** (already runs the River client/elector/workers). `metaldocs-api` runs **no
  scheduler goroutine** — it is stateless sync + authz.
- **Maintenance queue:** a dedicated River queue **`maintenance`** with a low `MaxWorkers` hosts the
  janitor + purge periodic jobs, so maintenance never starves business jobs. This is the platform
  substitute for the retired hand-rolled pressure-probe (ADR 0067 §1) — behavior-equivalent
  (maintenance is throttled below business traffic), not line-equivalent.
- **Retired:** `internal/modules/jobs/scheduler/` (scheduler + lease_reaper), `metaldocs.job_leases`,
  `acquire_lease`/`heartbeat_lease`/`release_lease`, the stuck-instance-watchdog advisory lock, the
  duplicated staging backoff math.
- **Migration ordering (expand/contract):** (1) add River periodic + staging + purge jobs, verify both
  binaries run; (2) delete scheduler code + advisory lock; (3) drop `job_leases` + the 3 lease functions
  in a forward-only migration ordered **after** step 2.

---

## 2. ★ F5.2 — janitors on River (the per-job migration table — binding)

Each janitor migrates as a **River periodic job** with the **same interval** and **unchanged body**.
"Post-migration equivalence" is the binding claim the per-janitor integration proof must demonstrate.

### 2.1 stuck-instance-watchdog

| Property | Current (lease scheduler) | Post-migration (River periodic job) |
|---|---|---|
| Anchor | `internal/modules/jobs/stuck_instance_watchdog/job.go:42-105`; reg `main.go:1005-1011` | River periodic job in `metaldocs-jobs`, queue `maintenance` |
| Schedule | 5 min ticker, `SkipOnPressure` | `PeriodicInterval(5*time.Minute)`, `PeriodicJobOpts{ID:"stuck-instance-watchdog", RunOnStart:false}` |
| Body | cancel/emit for `in_progress` instances `submitted_at < now()-7d`, batch 50 | **unchanged** — same query, batch, drift_policy |
| Idempotency key | lease (`acquire_lease`) + `pg_try_advisory_lock` (job.go:114) | **River elector** (single insert cluster-wide) + **single queue dequeue**; unique-job key = job kind (no overlapping run). **Advisory lock removed** (§2.6). Body is idempotent (re-run cancels nothing new). |
| Failure behavior | error logged, scheduler loop continues next tick | River retry per `MaxAttempts` + backoff; exhaustion → **discarded** (inspectable DLQ, REQ-ASYNC-3). Same net effect (retried next cycle), now with a DLQ. |
| Tenant seed | seeds identity before tenant-scoped writes | **preserved** — `SeedTxIdentity`/`SeedTxTenant` in the job tx (M3 RLS backstop) |
| **Equivalence proof (F5.2)** | — | Integration test: seed a 7-day-stuck `in_progress` instance → run the periodic job once → instance cancelled / governance event emitted, identical to pre-migration. |

### 2.2 idempotency-janitor

| Property | Current | Post-migration |
|---|---|---|
| Anchor | `internal/modules/jobs/idempotency_janitor/job.go:29-86`; reg `main.go:1013-1019` | River periodic job, queue `maintenance` |
| Schedule | 15 min, `SkipOnPressure` | `PeriodicInterval(15*time.Minute)`, `ID:"idempotency-janitor"`, `RunOnStart:false` |
| Body | pass1 `DELETE idempotency_keys WHERE expires_at<now()` batch 5000×≤10; pass2 count orphaned in_flight past 5-min grace | **unchanged** |
| Idempotency key | lease | River elector + unique job kind; **DELETE is idempotent** (tombstone) |
| Failure behavior | error logged+returned | River retry → discarded on exhaustion |
| **Equivalence proof** | — | Integration test: seed expired + in_flight keys → run → expired rows deleted, orphan count reported; identical to pre-migration. |

### 2.3 audit-integrity-validator

| Property | Current | Post-migration |
|---|---|---|
| Anchor | `internal/modules/jobs/audit_integrity_validator/job.go:17-41`; reg `main.go:1021-1027` | River periodic job, queue `maintenance` |
| Schedule | 1 h, `SkipOnPressure` | `PeriodicInterval(1*time.Hour)`, `ID:"audit-integrity-validator"`, `RunOnStart:false` |
| Body | `validator.ValidateIntegrity(ctx)` (published `audit` port; last-10k window) | **unchanged** — same published-port call, no reach into audit internals |
| Idempotency key | lease | River elector + unique job kind; **read-only, no side effects** (naturally idempotent) |
| Failure behavior | error logged+returned (integrity issue surfaced) | River retry → discarded on exhaustion; the surfaced-error semantics preserved |
| **Equivalence proof** | — | Integration test: seed a tampered chain row → run → the job returns/flags the integrity failure, identical to pre-migration. |

### 2.4 lease-reaper — **DELETED (not migrated)**

| Property | Current | Post-migration |
|---|---|---|
| Anchor | `internal/modules/jobs/scheduler/lease_reaper.go:21-80`; reg `main.go:1029-1035` | **removed** |
| Purpose | GC expired rows in `metaldocs.job_leases` | **vacuous** — `job_leases` is dropped; River owns its own leader lifecycle in `river_leader` |
| **Proof** | — | `grep` census: no `lease_reaper` registration, no `job_leases` reference anywhere (§2.5). Recorded as an intentional removal in `evidence.md`. |

### 2.5 ★ Lease-scheduler retirement census (binding — the anti-shadow clause)

The whole point is **one** primitive. A second scheduler that merely stops being *called* is a **C4
FAIL**. After F5.2, a `grep` census returns **0** (or each residual explicitly allowlisted with a
reason) for:

- `internal/modules/jobs/scheduler/` (scheduler.go, lease_reaper.go) — package deleted.
- `metaldocs.acquire_lease` / `heartbeat_lease` / `release_lease` — no call sites; functions dropped by migration.
- `metaldocs.job_leases` — no reads/writes; table dropped by migration.
- `pg_try_advisory_lock` in `stuck_instance_watchdog` — removed (§2.6).
- Scheduler registration in `apps/api/cmd/metaldocs-api/main.go` (`registerScheduledJobs`, lines ~599-617,1005-1035) — removed; **`metaldocs-api` builds and starts with no scheduler goroutine.**

Migration ordering per §1 (drop lease objects only after the writing code is gone).

### 2.6 ★ Watchdog single-runner advisory-lock removal + River singleton proof (binding)

> **⚠ ERRATUM 2026-07-04 (HS-7 — false premise corrected in place, ratification carried to M5 HS-1).**
> This section as originally written claimed removing the watchdog's advisory lock **retires H-PRE-1**.
> That premise is **FALSE** and is withdrawn. Runtime-truth verification (during F5.2 T6/T7) established:
> **H-PRE-1** ("never call an authz-recording read inside a lock-holding atomic tx") is motivated by the
> **audit hash-chain writer's `pg_advisory_xact_lock`** (`internal/modules/audit/infrastructure/postgres/writer.go:59`)
> combined with `authz.Require` recording a system_admin bypass audit **in the caller's tx**
> (`internal/modules/iam/authz/authz.go:119`) — it governs authz-recording reads nested in ANY lock-holding
> tx (audit-writer, auth-repo, documents-repo, migrate). The **stuck-instance-watchdog's
> `pg_try_advisory_lock(hashtext(JobName))`** was pure **single-runner mutual exclusion** with **no**
> authz-recording read — a *different, unrelated* advisory lock. Removing it therefore **neither triggers
> nor retires H-PRE-1**. H-PRE-1 stays **LIVE**. `[[advisory-lock-deadlock-constraint]]`.

**What M5 actually does (corrected, binding):** the watchdog's redundant single-runner advisory lock is
**removed** because River's leader elector (single periodic insert) + queue dequeue (`FOR UPDATE SKIP
LOCKED`, single claim) subsume cluster-wide single-runner. The watchdog body is additionally idempotent
(a second run finds no new stuck instances), so even a hypothetical double-run is harmless.

**Required evidence (binding — proves the single-runner guarantee that justifies removing the lock):** an
integration test that starts **two** River clients against one database and asserts the watchdog job body
**executes exactly once** (elector single-insert + queue single-claim), with the advisory lock removed
(`internal/platform/jobs/river/singleton_integration_test.go`, P4). This validates River's single-runner
guarantee. It is **not** an H-PRE-1 retirement proof (H-PRE-1 is unaffected). If it fails, the watchdog lock
removal is unsafe and must be reverted (HS-2 boundary call) — recorded, never silently satisfied.

**Doc updates (in this milestone):** `invariant-checklist.md` H-PRE-1 line reaffirmed **LIVE** with a
clarifying M5 note (watchdog lock removal is unrelated); the memory `advisory-lock-deadlock-constraint` is
**kept** (not retired) with the same clarifier. No "retired" claim anywhere.

---

## 3. ★ F5.3 — staging dispatch on River (binding)

### 3.1 Dispatch migration

| Property | Current (poller) | Post-migration (River) |
|---|---|---|
| Anchor | `staging_outbox_worker.go:63-119` | River worker per dispatch kind (pdf, materialize) in `metaldocs-jobs` |
| Enqueue | `Enqueue` in caller business tx, `ON CONFLICT (tenant_id, revision_id) DO NOTHING` (`staging_outbox.go:49-67`) | **unchanged** — `Enqueue` stays in the business tx; the enqueue *also* inserts a River job via `InsertTx` on the same tx (or the existing row is claimed by a River job) — the transactional-outbox guarantee is preserved |
| Dispatch trigger | poll `ClaimPending` (`FOR UPDATE SKIP LOCKED`) every `pollEvery` | River job dequeue (no bespoke poll loop) |
| **Idempotency key** | `(tenant_id, revision_id)` unique on the outbox row | **preserved** — duplicate enqueue = one outbox row = one dispatch; River job keyed to the row so redelivery re-dispatches the same content idempotently (the downstream render consumer is already idempotent per ADR 0009/0015) |
| Retry/backoff | **duplicated** exp-backoff `1<<min(attempts,30)*30s cap 30m` in **two** branches (`worker.go:99,109`) | **DELETED** — River `MaxAttempts` + `RetryPolicy` own retry/backoff; terminal failure = River **discarded** state (inspectable DLQ, REQ-ASYNC-3) |
| Tenant seed | `MarkDispatched`/`MarkFailed` seed `SeedTxTenant` (`staging_outbox.go:185`) before write | **preserved** — RLS engages on the dispatch write (M3 F3.2) |

### 3.2 Backoff-deletion census (binding)

After F5.3, `grep` for the duplicated backoff expression (`1<<`…`30*time.Second` / `staleAfter`
hand-rolled retry) in `render/fanout` returns **0** — River owns retry. `ClaimPending`/`ResetStaleClaims`
hand-rolled claim logic is removed in favor of River dequeue (or explicitly allowlisted if a thin
claim-adapter is retained, with reason).

### 3.3 F5.3 exit criteria

pdf + materialize dispatch integration proof (testdb): enqueue-in-business-tx → River job dispatches the
render request → `ON CONFLICT` idempotency holds (duplicate enqueue = one dispatch) → tenant seed present
(RLS engages). Duplicated-backoff census = 0. `go build ./...` green. Matches this §3.

---

## 4. ★ F5.4 — outbox retention (binding policy + proof)

### 4.1 Retention policy (locked values)

| Rows | Policy | Mechanism |
|---|---|---|
| River's own job rows | `CompletedJobRetentionPeriod = 24h`; `CancelledJobRetentionPeriod = 24h`; `DiscardedJobRetentionPeriod = 7*24h` (DLQ inspection window) | River native cleaner (config only, no code) |
| `materialize_dispatch_outbox` / `pdf_dispatch_outbox` — `status='dispatched'` | delete `dispatched_at < now() - 7d` | River periodic **purge job** (`ID:"staging-outbox-purge"`, `PeriodicInterval(24*time.Hour)`, queue `maintenance`), bounded batch **5000 rows × ≤10 iterations/run** |
| `*_dispatch_outbox` — `dead_lettered_at IS NOT NULL` | **retained** (not auto-purged) | kept for ops inspection (REQ-ASYNC-3 inspectable DLQ); a separate future policy may prune — recorded as a bounded defer (§8) |

Rationale: dispatched rows have no residual value once the render succeeded (7d is a generous audit
tail); dead-lettered rows must stay inspectable. Values are **binding** — changing them is HS-7.

### 4.2 F5.4 exit criteria

Retention integration proof (testdb): seed dispatched rows with `dispatched_at` older than 7d **and**
recent dispatched rows **and** a dead-lettered row → run the purge job → **old dispatched rows gone,
recent dispatched rows kept, dead-lettered row kept**; the DELETE is batch-bounded (no unbounded single
statement). Growth demonstrably bounded. Matches §4.1.

---

## 5. ★ F5.5 — fanout ordering (binding commutativity proof)

### 5.1 The claim (ADR 0067 §4)

Lifecycle fanout is **idempotent-commutative**, so strict cross-job ordering is **not** required:

- Each lifecycle event fans out to **additive per-recipient-per-event notification rows**: `INSERT INTO
  metaldocs.notifications (..., source_event_id) ... ON CONFLICT (recipient_user_id, source_event_id)
  WHERE source_event_id IS NOT NULL DO NOTHING` (`fanout_worker.go:123-136`). `source_event_id` is minted
  at emit time (`documents/domain/notification_events.go:15`).
- There is **no shared mutable "current status" row** a later event could clobber — `published` fanout and
  `superseded` fanout write **distinct** rows keyed by their own `source_event_id`.
- Terminal state = the **set of notification rows** = union of per-event fanouts ⇒ **commutative** (order-
  independent) and **idempotent** (redelivery-safe via `ON CONFLICT`).

### 5.2 F5.5 exit criteria

Concurrent race test (testdb, real Postgres): for **one** document, race a `published` fanout and a
`superseded` fanout in **both** interleavings (deterministic barrier). Assert:

| Assertion | Expected |
|---|---|
| Terminal notification-row set | **Identical** across both interleavings (order-independent). |
| Idempotency | Redelivering either event inserts **no** duplicate row (`ON CONFLICT` holds). |
| No lost/inverted state | Every obligated recipient has the correct per-event rows; none dropped, none overwritten. |

The commutativity argument (§5.1) is recorded in `evidence.md` and cross-referenced to ADR 0067 §4. **If**
the test reveals a shared mutable projection (breaking commutativity), F5.5 adds an ordering key/guard and
re-proves — that is an HS-6 scope surface (surface it), not a silent contract edit.

---

## 6. DB-as-last-line invariants preserved (binding — cross-feature)

- **Transactional outbox stays by construction** — River `InsertTx` on the business tx; no network call
  shares the state-write tx. F5.3 keeps `Enqueue` in the caller tx.
- **Idempotent consumers** — staging `ON CONFLICT (tenant_id, revision_id)`; notifications `ON CONFLICT
  (recipient_user_id, source_event_id)` — both preserved.
- **M3 FORCE-RLS backstop** — per-message/run `SeedTxTenant`/`SeedTxIdentity` on every migrated tenant-
  scoped job; a missing seed silently absorbed by NULL-permissive RLS is a defect the per-job proof
  guards against.
- **River schema via `rivermigrate`** (`bootstrap/jobs.go:69-80`); staging status CHECK + FORCE-RLS stay;
  retention is a bounded DELETE, **no new trigger**, no new business logic in triggers.
- **No wire-contract change** — no `openapi.yaml` edit; capability registry size untouched.
- **ADR 0009/0015** (pdf outbox / freeze-materialize split) semantics unchanged — M5 moves *dispatch/
  scheduling*, not the render worker execution pipeline.

## 7. Cross-feature constraints (bind all features)

- Runtime truth beats docs; River v0.37.1 native primitives are the mechanism (no hand-rolled equivalent survives).
- **Subagent-driven implementation** (`superpowers:subagent-driven-development`; sonnet implement/review, haiku mechanical, never fable, ≤15 concurrent); main session orchestrates/reviews/commits; the `milestone-validator` judges and writes `qa/milestone-qa.md`; main session flips status only on PASS.
- **testdb factory** for every proof (real concurrent for F5.5, not sqlmock); **targeted `-run`** only — the full integration suite is NOT run locally (20+ min box); bounded defers recorded in `evidence.md` with triggers.
- **All 4 binaries build** + `.\scripts\check-system-runnable.ps1` + **live QA drive** (scheduled publish + notification fanout on the consolidated stack) at close.
- **Commit after verified work; never push; never commit `docs/release/`; plans dir gitignored (never force-add).**
- **HS-7:** implementation compared section-by-section to this contract; drift → fix code to contract or re-open contract with operator approval.
- Module boundary: janitor/staging refactors stay within their owning modules; cross-module access via published interfaces only (audit `ValidateIntegrity` port, `LifecycleEventArgs`, the River platform client).

## 8. Bounded defers (recorded, with triggers)

| Defer | Rationale | Trigger / owner |
|---|---|---|
| Dead-lettered `*_dispatch_outbox` row pruning | Kept inspectable per REQ-ASYNC-3; auto-pruning them needs an ops-retention decision beyond M5 | M8 ops-readiness, or when a DLQ-retention policy is set |
| Prometheus queue-depth / oldest-item-age metric (REQ-ASYNC-4 metric half) | Observability is M8's scope; M5 preserves the watchdog, not the metrics posture | M8 F8.3 metrics-backup |
| Platform outbox **execution** (metaldocs-worker pdf/materialize render) onto River | M5 consolidates dispatch/scheduling; render-worker execution is a separate pipeline (ADR 0009/0015) with no consumer contract asking to move it | Future async-consolidation increment, if a seam demands it |
| Full `-tags integration` run of the M5 proofs on the local box | 20-min box constraint (mission §10) | Run on CI / capable box before program close-out; targeted `-run` drives authored regardless (M1–M4 precedent) |
| Full `If-Match` OCC unification / `documents/approval` promotion / CLAUDE.md janitor-wording rewrite | Other milestones (ADR 0066 target; M1; M9 F9.4) | Their milestones |
