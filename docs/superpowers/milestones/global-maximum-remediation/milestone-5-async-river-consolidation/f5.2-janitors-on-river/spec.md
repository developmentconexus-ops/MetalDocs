# F5.2 — janitors on River (spec)

> **Milestone:** M5 · **Status:** in progress
> **Binding parent:** `../validation-contract.md` §2 (the per-janitor migration table). This spec is a
> distillation; on any conflict the **contract wins** and the divergence is HS-7.
> **Rails:** ADR 0067 (River single primitive; janitors in `metaldocs-jobs`). **HS-7 correction:** M5 removes
> the watchdog's *unrelated* single-runner advisory lock; **H-PRE-1 stays LIVE** (governs authz-reads in
> lock-holding txs — audit writer, not the watchdog). See ADR 0067 §H-PRE-1 erratum.

## Consumer contract (who consumes the output, and the shape required)

- **Consumer 1 — the `metaldocs-jobs` runtime.** Requires each of the 3 surviving janitors registered as
  a **River periodic job** on a dedicated **`maintenance`** queue, inserted leader-only (cluster-wide
  singleton via the River elector), body byte-behavior-identical to today. Shape: a River `Worker[Args]`
  + a `river.NewPeriodicJob(PeriodicInterval(d), argsFn, &PeriodicJobOpts{ID, RunOnStart:false})` entry
  in `Config.PeriodicJobs`, mirroring the existing scheduled-publish job wiring.
- **Consumer 2 — `metaldocs-api` runtime.** Requires the custom lease scheduler **gone**: api starts with
  **no scheduler goroutine**, hosts no janitor. Stateless sync + authz.
- **Consumer 3 — the milestone-validator.** Requires: the §2.5 retirement census returns 0 (scheduler
  pkg deleted, no `job_leases` / `acquire_lease`/`heartbeat_lease`/`release_lease` / `pg_try_advisory_lock`
  references), the §2.6 H-PRE-1 singleton integration proof green, and a per-janitor equivalence
  integration test (testdb) for each of the 3 survivors.

## What to implement (per contract §2)

| Job | Interval | River `ID` | Body | Idempotency |
|---|---|---|---|---|
| stuck-instance-watchdog | 5 min | `stuck-instance-watchdog` | **unchanged** — cancel/emit `in_progress` where `submitted_at < now()-7d`, batch 50. **Advisory lock removed** (§2.6). | River elector (single insert) + queue (single dequeue); body idempotent |
| idempotency-janitor | 15 min | `idempotency-janitor` | **unchanged** — pass1 DELETE `expires_at<now()` batch 5000×≤10; pass2 count orphaned in_flight past 5-min grace | River elector + unique kind; DELETE idempotent |
| audit-integrity-validator | 1 h | `audit-integrity-validator` | **unchanged** — `validator.ValidateIntegrity(ctx)` via the published `audit` port | read-only, naturally idempotent |
| lease-reaper | — | — | **DELETED** (§2.4) — `job_leases` dropped, River owns `river_leader` lifecycle | n/a |

`RunOnStart:false` for all three. Each preserves its `SeedTxIdentity`/`SeedTxTenant` tenant seed inside
the job tx (M3 RLS backstop). Failure behavior: River `MaxAttempts` retry + backoff → `discarded` DLQ on
exhaustion (REQ-ASYNC-3) — net-equivalent to "retried next tick", now with an inspectable DLQ.

### Migration ordering (contract §1, expand/contract — do NOT reorder)
1. Add the 3 River periodic jobs + `maintenance` queue in `metaldocs-jobs`; verify **both** binaries build
   and start.
2. Delete `internal/modules/jobs/scheduler/` (scheduler + lease_reaper), remove the api registration, remove
   the watchdog advisory lock.
3. Forward-only migration dropping `metaldocs.job_leases` + `acquire_lease`/`heartbeat_lease`/`release_lease`
   — ordered **after** step 2 (no writer left).

## Non-goals

- **No body logic change.** Not "improve" any janitor — same query, batch, drift policy, seed. This is a
  host/primitive migration, not a behavior change. Changing a janitor's semantics is out of scope (HS-6).
- **No staging / retention / fanout work** — those are F5.3/F5.4/F5.5.
- **No new capability, no openapi edit, no capability-registry-size change.**
- **No metrics/Prometheus** (M8 defer, contract §8).
- The `maintenance` queue's `MaxWorkers` is a throttle (platform substitute for the retired pressure-probe),
  **not** a re-implementation of the pressure-probe — do not port the probe.

## Validation Gate (acceptance — all must hold)

1. **3 equivalence integration tests (testdb, targeted `-run`)** — one per surviving janitor, each proving
   post-migration body identical to pre-migration (contract §2.1/§2.2/§2.3 "equivalence proof" rows):
   - watchdog: seed a 7-day-stuck `in_progress` instance → run job once → cancelled / governance event emitted.
   - idempotency: seed expired + in_flight keys → run → expired deleted, orphan count reported.
   - audit: seed a tampered chain row → run → integrity failure flagged/returned.
2. **§2.6 H-PRE-1 singleton proof (testdb, real Postgres):** two River clients on one DB → the watchdog
   periodic job executes **exactly once per tick** (elector singleton), advisory lock removed. Green ⇒
   watchdog single-runner lock removal is safe; red ⇒ advisory lock stays, HS-2 surfaced (no silent contract
   satisfaction). NOT an H-PRE-1 retirement proof (HS-7 — H-PRE-1 stays LIVE).
3. **§2.5 retirement census = 0** (grep, recorded in evidence): `internal/modules/jobs/scheduler/` deleted;
   no `acquire_lease`/`heartbeat_lease`/`release_lease`/`job_leases`/`pg_try_advisory_lock` refs;
   `registerScheduledJobs` gone from `apps/api/cmd/metaldocs-api/main.go`.
4. **Both binaries build + start:** `go build ./...` green; `metaldocs-api` starts with no scheduler
   goroutine; `metaldocs-jobs` starts with the 3 periodic jobs on `maintenance`.
5. **Doc updates:** `developing-new-work/references/invariant-checklist.md` H-PRE-1 line + memory
   `advisory-lock-deadlock-constraint` record M5 removed the watchdog's *unrelated* single-runner lock;
   **H-PRE-1 stays LIVE** (HS-7 correction).
6. Section-by-section match to contract §2. No divergence, or divergence surfaced as HS-7.

## Interview record

No operator interview — the contract §2 (operator-committed at D4) IS the locked spec; F5.2 is its
mechanical execution. The "questions" were settled upstream in ADR 0067 (topology, H-PRE-1 disposition)
and the contract (per-job intervals, batch sizes, equivalence-proof definitions). The only open question
F5.2 execution resolves empirically is the §2.6 singleton proof — pre-committed as an acceptance gate, not
an interview.

## ADR

No new ADR — F5.2 executes ADR 0067's already-accepted decisions. The **watchdog advisory-lock removal**
it enacts is ADR 0067 §H-PRE-1 (which — per its HS-7 erratum — keeps **H-PRE-1 LIVE**; the lock removed
is the watchdog's *unrelated* `pg_try_advisory_lock`, not H-PRE-1's audit-writer lock); its doc/memory
updates land in this feature.
