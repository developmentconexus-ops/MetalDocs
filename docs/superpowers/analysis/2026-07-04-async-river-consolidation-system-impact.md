# System-impact analysis — Async consolidation onto River (M5)

**Date:** 2026-07-04
**Intent (one line):** Consolidate MetalDocs' three parallel async job infrastructures onto River — migrate the 4 lease-scheduler janitors and the StagingOutboxWorker poller onto River (periodic jobs + transactional enqueue + native retention), retire the custom Postgres-lease scheduler + `job_leases`, bound outbox growth, and guarantee (or prove commutative) lifecycle fanout ordering.
**Work type:** feature (architecture/infrastructure consolidation — no new bounded-context module, no new capability)
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(see §10)*

> Mission `global-maximum-remediation` §7 M5; D7 requires this gate Green/Yellow **and** an ADR before
> the milestone plan is authored. Runtime truth verified in code 2026-07-04 (agent map, exact anchors
> below); River native-capability premise re-proven against River v0.37.1 docs (Context7) — the mission
> flagged this as the one claim M5's own gate must re-prove.

---

## 0. Runtime-truth basis (verified 2026-07-04 — code + River docs)

**Three parallel job infrastructures confirmed in code:**

| Infra | Binary | Primitive | Anchor |
|---|---|---|---|
| **River v0.37.1** — scheduled-publish + notifications fanout | `metaldocs-jobs` | River client, `InsertTx` on business tx, `AddWorker` | `apps/jobs/cmd/metaldocs-jobs/main.go:37-54,43-44`; client `internal/platform/jobs/river/client.go:23-46`; schema migrate `internal/platform/bootstrap/jobs.go:69-80` |
| **Custom Postgres-lease ticker scheduler** — 4 janitors | `metaldocs-api` | `metaldocs.acquire_lease`/`heartbeat_lease`/`release_lease` + `job_leases` table; per-job ticker; pressure-probe skip | scheduler `internal/modules/jobs/scheduler/scheduler.go:117-273,413-445`; lease fn `db/baseline/0001_current_schema.sql:59-99`; table `:1354-1361`; registration `apps/api/cmd/metaldocs-api/main.go:599-617,1005-1035` |
| **StagingOutboxWorker poller** — pdf/materialize dispatch | `metaldocs-api` | poll loop + `ClaimPending` (`FOR UPDATE SKIP LOCKED`) + duplicated exp-backoff | worker `internal/modules/render/fanout/staging_outbox_worker.go:63-119`; repo `internal/modules/render/fanout/staging_outbox.go:49-131`; tables `db/baseline/0001_current_schema.sql:1368-1382,1484-1498` |

**The 4 janitors (lease scheduler):**

| Janitor | Interval | Behavior | Idempotency / failure | Anchor |
|---|---|---|---|---|
| stuck-instance-watchdog | 5 min | auto-cancel/emit event for `in_progress` approval instances `submitted_at < now()-7d`, batch 50 | **`pg_try_advisory_lock` (H-PRE-1) + lease**; error logged, loop continues | `internal/modules/jobs/stuck_instance_watchdog/job.go:42-105,114,151`; reg `main.go:1005-1011` |
| idempotency-janitor | 15 min | pass1 `DELETE idempotency_keys WHERE expires_at<now()` batch 5000×10; pass2 count orphaned in_flight | delete is idempotent; error logged+returned | `internal/modules/jobs/idempotency_janitor/job.go:29-86`; reg `main.go:1013-1019` |
| audit-integrity-validator | 1 h | `ValidateIntegrity` hash-chain recompute (last-10k window) | read-only, no side effects | `internal/modules/jobs/audit_integrity_validator/job.go:17-41`; reg `main.go:1021-1027` |
| lease-reaper | 10 min | reclaim expired `job_leases` via CTE `FOR UPDATE SKIP LOCKED` + DELETE | delete idempotent | `internal/modules/jobs/scheduler/lease_reaper.go:21-80`; reg `main.go:1029-1035` |

Note: **lease-reaper only exists to garbage-collect the custom lease table** — it becomes vacuous once
the lease scheduler is retired (River owns its own leader/lease lifecycle in `river_leader`).

**Outbox growth:** `MarkDispatched` only flips `status='dispatched', dispatched_at=now()`
(`staging_outbox.go:112-131`); dead-letter sets `dead_lettered_at` (`:143`). **No purge anywhere** —
both `materialize_dispatch_outbox` and `pdf_dispatch_outbox` grow unbounded.

**Fanout ordering:** lifecycle events enqueued via `RiverLifecycleEventEnqueuer.EnqueueLifecycleEventTx`
(`internal/modules/documents/approval/jobs/lifecycle_event_enqueuer.go:28-38`, `InsertTx`, no
`ScheduledAt`); consumed by `NotificationsFanoutWorker`
(`internal/modules/notifications/infrastructure/fanout_worker.go:47-138`) with **no ordering guarantee**
between jobs. Idempotency is per-recipient-per-event via partial unique index `ON CONFLICT
(recipient_user_id, source_event_id) DO NOTHING` (`:123-136`), **not temporal** — published/superseded
for one document can fan out inverted.

**River native-capability premise (re-proven, River v0.37.1, Context7):**
- **Periodic jobs** — `Config.PeriodicJobs`, `river.NewPeriodicJob(river.PeriodicInterval(d), argsFn, &river.PeriodicJobOpts{ID, RunOnStart})`; `client.PeriodicJobs().Add(...)` for dynamic.
- **Leader election** — `river_leader` table (migration line 7); client `ID` "used for leader election";
  periodic-job insertion is a leader-only maintenance service ⇒ singleton across replicas.
- **Uniqueness belt** — `UniqueSkippedAsDuplicate` / unique opts prevent duplicate periodic insert.
- **Native retention** — `CompletedJobRetentionPeriod` / `Cancelled` / `Discarded` job cleaner (River's
  own job rows; `-1` = keep forever).
- **Transactional enqueue** — `InsertTx(ctx, tx, args, nil)` already in use.

⇒ River **can** subsume the lease scheduler (periodic + elector) and the staging poller (transactional
enqueue + worker + retry). Premise **CONFIRMED**.

---

## 1. Classify & own
*(CLAUDE.md Orientation rule)*

- **Work type:** feature (infrastructure consolidation). No new bounded-context module; no new capability.
- **Owning module(s):**
  - `internal/modules/jobs` (scheduler + the 4 janitors) — **owns** the janitor migration + lease-scheduler retirement.
  - `internal/platform/jobs/river` — **owns** the River client bundle extended with periodic-job + retention config.
  - `internal/modules/render/fanout` — **owns** the staging-outbox → River migration + retention.
  - `internal/modules/notifications` + `internal/modules/documents/approval/jobs` — **own** the fanout-ordering guarantee (emit + consume sides).
- **Explicitly NOT owning:**
  - `iam`/`authz` — not touched except *reusing* `SeedTxTenant`/`SeedTxIdentity` per message (no PDP change).
  - `audit` — the validator janitor's `ValidateIntegrity` body is unchanged; only its *scheduling* moves.
  - `docx-renderer` (Node) — no async-infra change; still a downstream materialize target.
- **Cross-module edges (with direction):**
  - `jobs → River (platform)` : janitors become River periodic jobs — via `internal/platform/jobs/river` client, published surface.
  - `render/fanout → River (platform)` : staging dispatch enqueues via River `InsertTx` (already the pattern for scheduled-publish).
  - `jobs/audit-validator → audit` : keeps its existing published `ValidateIntegrity` port (no reach into audit internals).
  - `notifications → documents` : fanout consumes `LifecycleEventArgs` (already a published contract).
  - All edges already go through published Go interfaces / platform clients — **no new cross-module repo/SQL reach introduced.**
- **Ambiguity?** None → no AS-3. The three owners are the same three that own the code today; consolidation moves *scheduling*, not *domain* ownership.

## 2. Foundation verdict
*(Global-Maximum rule)*

- **Base you'd build on:** a hand-rolled Postgres-lease scheduler (`acquire_lease`/`heartbeat`/`release` +
  `job_leases` + `lease_reaper` GC + pressure-probe ticker) **and** a hand-rolled staging poller (own
  `ClaimPending`/`ResetStaleClaims`/exp-backoff duplicated in two branches) — running **beside** a River
  deployment that already ships leader election, periodic jobs, retention, and transactional enqueue.
- **Sound, or legacy/patch/workaround?** **Local maximum.** Per the frameworks-catalog rule ("a
  hand-rolled equivalent of any platform framework is a defect — it bypasses the invariant the framework
  exists to enforce"), the lease scheduler and the staging backoff are hand-rolled equivalents of
  primitives River already runs in this very system. Three retry/idempotency/election code paths are
  maintained where one trusted primitive exists. The stuck-instance-watchdog's belt-and-suspenders
  (advisory lock **+** lease) and the H-PRE-1 deadlock constraint are *symptoms* of not having one
  trusted scheduling primitive — exactly the 2026-07-03 review's dimension-5 finding.
- **Global-maximum structure (named):** **River as the single job infrastructure.** Janitors → River
  periodic jobs (leader-elected singleton); staging dispatch → River transactional job; retention → River
  native cleaner (River rows) + a River periodic purge job (staging-outbox business rows); lease scheduler
  + `job_leases` + `lease_reaper` deleted.
- **AS-2?** **No.** The work moves *to* the global maximum; it does not optimize *inside* the patch. This
  is the mission's own thesis (§2 goal 3, review P3-7). No hard-stop. The **trade-off** (carried into the
  ADR, D7): migration cost + re-verifying/retiring H-PRE-1 under River semantics + making River a
  dependency of the **api** binary's maintenance work (today the lease scheduler runs in api; the janitors
  must keep running when the jobs binary is where River workers live — a deployment-topology decision the
  ADR must settle: *which binary hosts the janitor periodic jobs*).

## 3. Invariant alignment
*(the 6 non-negotiables — `references/invariant-checklist.md`)*

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | No (indirect) | Janitors are system-path jobs (no PDP); no capability added/changed. Watchdog keeps authz-recording reads **off-tx** (H-PRE-1). | — (no `authz.Require` inside job tx) |
| Contract-first (OpenAPI + oapi-codegen) | **No** | Pure infrastructure — no HTTP route added/changed. No `openapi.yaml` edit. (If any ops/inspection endpoint is later wanted, that would be contract-first — out of M5 scope.) | — |
| Multi-tenant pooled (`tenant_id` / tx-local GUC / 404) | **Yes** | Staging dispatch + fanout must keep per-message `SeedTxTenant` so FORCE-RLS backstops async (M3 F3.2). Janitors that touch tenant rows seed identity per run. River job args carry `tenant_id` where the work is tenant-scoped. | `authz.SeedTxTenant` (`staging_outbox.go:185`), `authz.SeedTxIdentity` |
| Async = transactional outbox | **Yes (central)** | River `InsertTx` on the business tx **preserves** the outbox guarantee (state-write + enqueue same tx; network call in the idempotent consumer). Staging migration keeps `Enqueue` in the caller tx with `ON CONFLICT (tenant_id, revision_id) DO NOTHING`. | Outbox repo pattern (`staging_outbox.go:49-67`), `river.InsertTx` |
| DB enforces invariants (triggers/constraints) | **Yes** | River schema migrated via `rivermigrate` (`river_job/leader/queue/notification`). Staging-outbox status CHECK + FORCE-RLS stay. Retention purge is a bounded DELETE, not a new trigger. | `rivermigrate` (`bootstrap/jobs.go:69-80`) |
| Cross-module via published interface only | **Yes** | Audit-validator keeps its `ValidateIntegrity` port; fanout keeps `LifecycleEventArgs`; janitors depend on platform River client, not other modules' tables. | published ports (existing) |

**No violation → no AS-1.** The outbox invariant is *strengthened* (one enqueue primitive, retention added), not weakened.

## 4. Capability wiring
**N/A** — no IAM capability is added or changed. Janitors are system-path maintenance jobs with no
route→capability gate and no in-tx `authz.Require`. (Registry size `TestCapabilityRegistrySize` untouched.)

## 5. Module wiring
**N/A** — no new bounded-context module is born. All owners (`jobs`, `render/fanout`, `notifications`,
`platform/jobs/river`) exist. `internal/modules/jobs/scheduler/` is **deleted**, not created; no
`module.go`/`port.go`/OpenAPI-tag birth steps apply.

## 6. Frameworks to reuse, not reinvent
*(`references/frameworks-catalog.md`)*

- **River client** (`internal/platform/jobs/river/client.go`) — **extend** with `PeriodicJobs` + retention config; do **not** stand up a parallel scheduler. (The whole milestone is "stop reinventing River.")
- **`TxRunner` (`Do`/`DoReadOnly`)** — janitor + purge DB work goes through the tx port, nil-tx rejected.
- **`authz.SeedTxTenant`/`SeedTxIdentity`** — per-message/per-run GUC seed so RLS backstops async (M3).
- **Outbox repo pattern** (`staging_outbox.go`) — `Enqueue` stays in the business tx; only the *dispatch* mechanism moves to River.
- **`rivermigrate`** — River schema management (already used at api startup).
- **`audit` published `ValidateIntegrity` port** — reused unchanged by the validator periodic job.
- **`testdb.Open` + factory** — every janitor + staging + purge + ordering proof is a `//go:build integration` testdb test.
- **No hand-rolled equivalent** of leader election, periodic scheduling, retry backoff, or job retention is permitted — River owns each. Every deleted line of `scheduler.go`/`lease_reaper.go`/duplicated backoff is the point.

## 7. Contract & data

- **OpenAPI-first:** no route change → no `openapi.yaml` edit, no regen. (Explicitly in-bounds: this milestone touches *no wire contract*.)
- **Migration(s):**
  - River schema already migrated (`rivermigrate`); confirm the migration runs in whichever binary hosts the periodic janitors.
  - **Retirement migration:** drop `metaldocs.job_leases` + `acquire_lease`/`heartbeat_lease`/`release_lease` functions **after** the scheduler code is deleted (expand/contract: stop using → delete). Forward-only (repo convention); a row pre-check is unnecessary (lease rows are ephemeral) but the migration must be ordered *after* the code that writes them is gone.
  - **Retention:** a River periodic **purge job** DELETEs `*_dispatch_outbox` rows where `status='dispatched' AND dispatched_at < now()-<policy>` (and a dead-letter policy for `dead_lettered_at`) — bounded batch, `tenant_id`-agnostic system path or per-tenant seed. Policy value fixed in the validation contract.
- **Destructive change?** Yes — lease scheduler + `job_leases` deletion. Sequence: migrate janitors onto River (both run) → verify → delete scheduler code → drop lease objects. Never delete the lease table while a running api still writes it.

## 8. Test & QA plan
*(`references/test-qa-gates.md`)*

- **Canonical framework:** `testdb` integration factory, `//go:build integration`, targeted `-run` only (full suite is the 20-min box — bounded defer per mission §10). R1–R4 discipline.
- **Per-job integration proof (mission F5.2/F5.3 acceptance — each janitor + staging dispatch):**
  - stuck-instance-watchdog: seeds a stuck `in_progress` instance, runs the River periodic job, asserts cancel/event + **singleton** (two client instances → one execution via elector) + **H-PRE-1 retired/re-verified** (advisory lock removed or proven still off-tx).
  - idempotency-janitor: seeds expired + in_flight keys, asserts expired deleted / orphans counted.
  - audit-integrity-validator: unchanged body; asserts it still runs on schedule and flags a tampered chain.
  - lease-reaper: **deleted** (asserts the janitor no longer exists; River owns lease GC) — recorded, not migrated.
  - staging dispatch (pdf + materialize): enqueue in business tx → River job dispatches → `ON CONFLICT` idempotency preserved; duplicated backoff deleted (asserted by grep census).
- **F5.4 retention proof:** seed dispatched rows older than policy → purge job → rows gone, recent rows kept; growth bounded.
- **F5.5 ordering proof:** race a `published` then `superseded` fanout for one document; assert no lost/inverted terminal notification state — **or** an explicit idempotent-commutative argument in the ADR (per-recipient-per-event `ON CONFLICT` already makes redelivery safe; the argument must show terminal state is order-independent, or add a guard/ordering key).
- **QA gates that apply:** async/idempotency, multi-tenant isolation (RLS backstop on the migrated jobs), DB-invariant (River schema + retention), docs. Contract + authz gates → mostly N/A (no route/capability). Name each in the milestone gate.
- **Close-out (mission §7 M5):** `go build ./...`; all 4 binaries build; `.\scripts\check-system-runnable.ps1`; **live QA drive** — `.\scripts\start-api.ps1 -Build`, drive a scheduled publish + a notification fanout, capture proof (network/logs). `milestone-validator` → `qa/milestone-qa.md`.
- **Evidence shape:** commands + outcomes + review/QA disposition + bounded defers. No bare "done".

## 9. Docs / ADR
*(`references/docs-adr-governance.md`)*

- **ADR required? YES (D7 mandates it).** **ADR 0067** — "Async job infrastructure consolidated onto
  River" — records: (a) River as the single primitive (subsumes lease scheduler + staging poller);
  (b) which binary hosts the janitor periodic jobs (deployment topology); (c) the **H-PRE-1
  re-verification/retirement** decision under River semantics (advisory lock removed because River's
  elector guarantees singleton — with evidence — or retained with rationale); (d) outbox retention policy;
  (e) fanout ordering guarantee vs idempotent-commutative proof. Supersedes/annotates nothing that names a
  MUST, but it *changes a standing infrastructure policy* ⇒ ADR is the right instrument. Accepted **before**
  the milestone plan is authored.
- **Wiki:** refresh `wiki/modules/jobs.md` (+ its tech-debt doc) for the scheduler retirement; note the
  staging-outbox change in the render/fanout doc; `Last verified` stamps refreshed on touched docs.
  wiki-curator pass after implementation.
- **REQ IDs cited:** REQ-ASYNC-* (transactional outbox / DLQ inspectability) from
  `wiki/architecture/backend-target-architecture.md` — cite the exact IDs in the ADR and milestone.
- **H-PRE-1:** the memory note + `invariant-checklist.md:58` say the constraint "holds until M5 formally
  re-verifies or retires it." ADR 0067 is where that formal disposition lands.

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — fits cleanly (owning modules clear, no invariant violated, River premise
  re-proven), proceed to design; **but** three named items are carried into the design as locked
  constraints and an ADR is mandatory (D7). Not Green only because of the ADR requirement + the H-PRE-1
  disposition + the deployment-topology decision, each a real design choice the brainstorm must resolve.
- **Open hard-stops:** AS-1 none · AS-2 none (moving to global max) · AS-3 none. No Red → HS-8 does not fire; design is **not** blocked.
- **Locked constraints handed to brainstorming / milestone:**
  1. **River is the only scheduling primitive** — janitors → River periodic jobs (leader-elected); staging → River transactional job; **no hand-rolled scheduler/backoff survives** (grep census of deleted primitives is a gate).
  2. **Transactional-outbox invariant preserved** — every enqueue stays `InsertTx` on the business tx; consumers idempotent; `ON CONFLICT` dedup kept.
  3. **RLS backstop preserved (M3 F3.2)** — per-message/per-run `SeedTxTenant`/`SeedTxIdentity` on every migrated tenant-scoped job.
  4. **H-PRE-1 formally dispositioned in ADR 0067** — retire (advisory lock removed, singleton proven by elector + integration test) or retain-with-rationale; evidence required.
  5. **Deployment topology decided in ADR 0067** — which binary hosts the janitor periodic jobs (api vs jobs), and the migration/retirement ordering (migrate → verify → delete scheduler → drop `job_leases`).
  6. **Outbox growth bounded** — River native retention (River rows) + a periodic purge job (staging business rows) with a policy value fixed in the validation contract.
  7. **Fanout ordering guaranteed or proven commutative** — race test or formal ADR argument; no lost/inverted terminal state.
  8. **lease-reaper is deleted, not migrated** — River owns its own lease lifecycle; recorded as intentional removal.
  9. **No wire-contract change** — M5 touches no `openapi.yaml`; if an inspection endpoint is ever wanted it is a separate contract-first change.
  10. **Test discipline** — testdb factory, targeted `-run`, bounded defers recorded; per-job integration proof for each janitor + staging dispatch + retention + ordering.

**Next step (D7):** author ADR 0067 (accepted) → then invoke `milestone` skill → author `milestone.md` +
`validation-contract.md` (every job enumerated: schedule, idempotency key, failure behavior, post-migration
equivalence) **before** any implementation.
