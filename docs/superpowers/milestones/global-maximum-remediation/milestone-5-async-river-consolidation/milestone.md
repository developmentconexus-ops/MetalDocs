# Milestone 5 — Async consolidation onto River

> **Program:** global-maximum-remediation  ·  **Governing spec:** `../mission.md` (§7 M5)
> **Status:** Spec (drafting)
> **Authored:** 2026-07-04 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates the
> milestone against *this* document. The full expected behaviors are pinned in
> `validation-contract.md` (D4), authored and committed **before** the first feature's
> implementation.
>
> **D7 pre-design gate (done before this spec):** the `developing-new-work` gate returned
> **🟡 Yellow** (`../../../analysis/2026-07-04-async-river-consolidation-system-impact.md`,
> committed cd2bceb3) and **ADR 0067** was accepted (`wiki/decisions/0067-...`, committed
> 5eb270c3) — both **before** this plan, per mission D7. HS-8 (Red ⇒ design blocked) did
> not fire.

## Objective

**After this milestone, MetalDocs runs on ONE async job infrastructure — River — and the
custom Postgres-lease scheduler no longer exists.** Concretely, observable at close:

1. The 3 surviving maintenance janitors (stuck-instance-watchdog, idempotency-janitor,
   audit-integrity-validator) run as **River periodic jobs** on the leader-elected
   `metaldocs-jobs` binary, on the same intervals, with the same behavior — proven by a
   per-janitor integration test. lease-reaper is **deleted** (it only GC'd the retired
   lease table).
2. `metaldocs.job_leases` + `acquire_lease`/`heartbeat_lease`/`release_lease` and the
   scheduler code (`internal/modules/jobs/scheduler/`) are **gone** — `grep` census = 0.
   `metaldocs-api` starts and runs with **no scheduler goroutine** (stateless sync+authz).
3. The staging outbox (pdf + materialize) **dispatches via River** transactional jobs; the
   duplicated exponential-backoff math is deleted (River owns retry).
4. The staging outbox tables **stop growing unbounded** — a River periodic purge job
   removes dispatched rows older than policy; a seeded-then-verified test proves it.
5. Lifecycle fanout (`published`/`superseded`) has a **proven** correctness story: an
   idempotent-commutative argument backed by a concurrent race test — no lost or inverted
   terminal notification state under reordering.
6. **H-PRE-1 is retired** — the stuck-instance-watchdog's `pg_try_advisory_lock` is removed
   (River's elector + single-dequeue subsume it), proven by a two-client singleton test.

**Quality bar moved:** the 2026-07-03 review's dimension-5 DEBT ("three parallel job
infrastructures where River could be one"; unbounded outbox; unverified fanout ordering) is
closed. Re-measured by: `grep` census of the retired primitives = 0; all 4 binaries build +
`.\scripts\check-system-runnable.ps1` green; the live QA drive of a scheduled publish + a
notification fanout succeeds on the consolidated stack.

Coherent slice: exactly the async-consolidation cluster (findings 11, 12, 13), no more. It
follows the correctness milestone (M4) and precedes the product features that depend on
consolidated scheduling (M6 periodic-review rides River periodic jobs).

## Appetite

- **Appetite:** the 5 features below (F5.1 gate/ADR already done). No new bounded-context
  module, no new capability, no wire-contract change. Behavior-preserving migration + two
  bounded correctness fixes (retention, fanout proof).
- **Rabbit holes (do not chase):**
  - **Re-implementing the pressure-probe "skip on DB pressure" by hand** — River's native
    queue config (`MaxWorkers`, `FetchCooldown`, a dedicated maintenance queue) is the
    platform mechanism; the bespoke connection-ratio probe is retired, not ported (ADR 0067
    §1). Reason: porting it would rebuild a hand-rolled primitive River already provides.
  - **Prometheus queue-depth / oldest-item metric (REQ-ASYNC-4 metric half)** — observability
    is **M8**'s scope; M5 preserves the watchdog, not the metrics posture. Reason: wrong
    milestone.
  - **Full `If-Match` OCC unification, `documents/approval` promotion, CLAUDE.md rewrites** —
    other milestones (M4 ADR 0066, M1/M9). Reason: not async.
  - **Migrating the platform outbox consumer (`metaldocs-worker` pdf/materialize *execution*)
    onto River** — worker execution is a separate pipeline (ADR 0009/0015); M5 consolidates
    *dispatch/scheduling*, not the render worker. Reason: no consumer contract asks for it;
    would balloon scope. Recorded as a bounded defer in `validation-contract.md` if a seam
    is touched.
  - **Any push to origin.** Never, this milestone (mission §2, §10).

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F5.1 | `f5.1-gate-and-adr` | *(done before this spec — D7)* `developing-new-work` system-impact analysis + **ADR 0067** for the job-infrastructure consolidation, incl. the H-PRE-1 re-verification-under-River decision. **Consumer:** the milestone planner + the M5 validator consume the gate verdict + the accepted ADR as the design rails. | Gate artifact committed, verdict **Green/Yellow** (is Yellow); ADR 0067 **Accepted**, indexed, and cited by later F5.x commits. Both committed before any F5.2+ implementation. |
| F5.2 | `f5.2-janitors-on-river` | The 3 surviving janitors become **River periodic jobs** (leader-elected, same intervals, unchanged bodies) hosted in `metaldocs-jobs`; the custom lease scheduler (`internal/modules/jobs/scheduler/`), `job_leases` table + 3 lease SQL functions, and **lease-reaper** are deleted; the stuck-instance-watchdog `pg_try_advisory_lock` is removed (H-PRE-1 retirement). **Consumer:** the `metaldocs-jobs` River client registers each as a `NewPeriodicJob`; ops observe the same maintenance effects. | Per-janitor integration proof (testdb): each preserves its behavior (watchdog cancels a seeded 7-day-stuck instance; idempotency-janitor deletes expired keys; audit-validator flags a tampered chain). **Singleton proof:** two River clients on one DB → watchdog periodic job executes **exactly once/tick** (elector, advisory lock removed). `grep` census of `job_leases`/`acquire_lease`/`scheduler.go`/`lease_reaper`/`pg_try_advisory_lock` = **0** (or allowlisted w/ reason). `metaldocs-api` builds + starts with no scheduler. H-PRE-1 retirement recorded with the singleton evidence. All matches `validation-contract.md` §2 per-job table. |
| F5.3 | `f5.3-staging-on-river` | StagingOutboxWorker **dispatch** moves to River transactional jobs (one worker per dispatch kind: pdf, materialize); `Enqueue` stays in the business tx (`ON CONFLICT` idempotency kept); the **duplicated exponential-backoff math is deleted** (River `MaxAttempts`+`RetryPolicy` own retry; terminal failure = River discarded/DLQ). Per-message `SeedTxTenant` kept (M3 RLS backstop). **Consumer:** the freeze/materialize + pdf-dispatch call sites enqueue via the River enqueuer; render fanout dispatches through River. | pdf + materialize dispatch integration proof (testdb): enqueue-in-tx → River job dispatches the render request; idempotency preserved (duplicate enqueue = one dispatch via `ON CONFLICT (tenant_id, revision_id)`); tenant seed present (RLS engages on the write). Duplicated-backoff `grep` census = 0. Matches `validation-contract.md` §3. |
| F5.4 | `f5.4-outbox-retention` | River native retention configured for River's own rows (`CompletedJobRetentionPeriod` etc); a **River periodic purge job** DELETEs `*_dispatch_outbox` rows `status='dispatched' AND dispatched_at < now()-<retention>` (+ dead-letter policy) in bounded batches. **Consumer:** the purge periodic job; ops see bounded table growth. | Retention integration proof (testdb): seed dispatched rows older than policy + recent rows → run purge → old rows gone, recent + non-dispatched rows kept; batch-bounded (no unbounded single statement). Retention policy value = the one fixed in `validation-contract.md` §4. Growth demonstrably bounded. |
| F5.5 | `f5.5-fanout-ordering` | Prove lifecycle fanout (`published`/`superseded`) is **idempotent-commutative** (ADR 0067 §4): additive per-recipient-per-event notification rows, no shared mutable status row → order-independent terminal state; backed by a concurrent race test. If the proof fails (a shared mutable projection exists), add an ordering key/guard. **Consumer:** `NotificationsFanoutWorker`; the notification reader observes a correct terminal set regardless of job order. | Concurrent race test (testdb): both interleavings of `published`→`superseded` fanout for one document yield the **identical** terminal notification-row set; no lost/inverted state; `ON CONFLICT (recipient_user_id, source_event_id)` dedup holds under redelivery. Commutativity argument recorded in `evidence.md` + ADR 0067 §4. Matches `validation-contract.md` §5. |

For each feature, "what to validate" is objectively checkable: a named integration test that
passes, a `grep` census that returns 0, a build that starts clean, a race test with a
deterministic single-winner/identical-set assertion. No "works"/"looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it
judges and writes `qa/milestone-qa.md`; the main session flips status only on its PASS), per
the binding C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`.
For M5:

1. **Per-feature acceptance** — every feature meets its declared "what to validate", and each
   feature's consumer contract (`spec.md`) was honored. F5.2 matches `validation-contract.md`
   §2 (per-janitor schedule/idempotency/failure table + singleton + H-PRE-1 retirement); F5.3
   matches §3 (dispatch idempotency + backoff deletion); F5.4 matches §4 (retention policy +
   bounded proof); F5.5 matches §5 (commutativity + race test).
2. **Workflow-class QA checklist** — `wiki/quality/qa-operating-system.md` async/idempotency +
   multi-tenant-isolation subsets; `wiki/quality/test-discipline.md` (testdb factory for every
   proof; targeted `-run` only — the full integration suite is NOT run locally, 20-min box).
3. **Regression** — M0–M4 gates still pass; the transactional-outbox invariant, the M3 RLS
   backstop (per-message seed), and the M2 write-tripwire are not regressed by any migrated
   job; `go build ./...` green; **all 4 binaries build**; `.\scripts\check-system-runnable.ps1`
   green.
4. **Quality-bar / root-cause check** — consolidation is at the **root**: River is the single
   primitive and the hand-rolled scheduler/backoff are **deleted** (census = 0), NOT left
   beside River (a second scheduler that merely stops being called is a **FAIL**). H-PRE-1 is
   **retired with evidence** (advisory lock removed + singleton proof), not merely "still
   satisfied". Retention + fanout are **proven** (seeded tests / race test), not asserted.
5. **No unplanned scope** — anything beyond F5.1–F5.5 recorded with rationale; the rabbit-hole
   list above is the scope-drift baseline.
6. **Live QA drive (runtime-visible milestone, mission §7 M5)** — `.\scripts\start-api.ps1
   -Build`; drive a **scheduled publish** cutover and a **notification fanout**; capture proof
   (logs/network) that both run on the consolidated River stack. Evidence in the closing
   feature's `evidence.md`.

**HS-7 (mission-specific):** the implementation is compared **section-by-section** against the
committed `validation-contract.md`. Any drift = stop; fix the code to the contract, or re-open
the contract **with operator approval** — never silently edit the contract to match the code.

## Dependencies & constraints

- **Depends on:** M0–M4 (committed; M4 passed operator HS-1 2026-07-04). River v0.37.1 already
  deployed (`metaldocs-jobs`). M3 RLS backstop (per-message `SeedTxTenant`) — migrated jobs must
  preserve it. ADR 0067 (accepted) + the M5 gate (Yellow) — both done.
- **Quality goals (ranked):** **correctness-preservation > deletion-of-duplication > simplicity.**
  (1) No janitor/staging behavior regresses — the migration is behavior-preserving, proven per
  job. (2) The hand-rolled primitives are actually *deleted*, not shadowed — the whole point is
  one primitive. (3) River config stays idiomatic (native levers, no bespoke wrappers). The
  validator uses this rank on any trade-off.
- **Architectural constraints (hard rules the validator can fail on):**
  - **River is the only scheduling/retry primitive** — no hand-rolled scheduler, elector, retry
    backoff, or job-lease survives (census gate).
  - **Transactional-outbox invariant preserved** — every enqueue stays `InsertTx` on the
    business tx; consumers idempotent; `ON CONFLICT` dedup kept (REQ-ASYNC-1/2).
  - **RLS backstop preserved (M3 F3.2)** — per-message/run `SeedTxTenant`/`SeedTxIdentity` on
    every migrated tenant-scoped job (a missing seed silently absorbed by NULL-permissive RLS is
    a defect).
  - **Migration ordering (expand/contract, destructive)** — migrate jobs onto River + verify
    both run → delete scheduler code + advisory lock → **then** drop `job_leases` + lease
    functions (forward-only migration ordered after the writing code is gone). Never drop lease
    objects while a running api still writes them.
  - **No wire-contract change** — M5 edits no `openapi.yaml`; no route/capability added
    (registry size untouched).
  - **DB stays the last line** — River schema via `rivermigrate`; staging status CHECK +
    FORCE-RLS stay; retention purge is a bounded DELETE, not a new trigger.
  - **Root-cause over symptom-patch** — binds C4.
- **Risks (named, with disposition):**
  - *Topology shift* (janitors api→jobs): `metaldocs-jobs` is already a required binary
    (scheduled-publish/notifications) — no new required binary. Mitigation: ADR 0067 §5 records
    it; live drive + system-runnable check prove both binaries still work. Accepted.
  - *H-PRE-1 retirement is wrong* (River doesn't actually singleton the periodic job):
    mitigation — the two-client singleton integration test is a **gate**; retirement is not
    declared until it is green. If it fails → advisory lock stays, H-PRE-1 not retired, recorded
    (HS-2 boundary call).
  - *Behavior drift under River retry semantics* (a janitor that relied on the lease's exactly-
    every-N-minutes cadence): mitigation — per-janitor integration proof + interval parity;
    River `RunOnStart`/interval mapped explicitly in `validation-contract.md` §2.
  - *Full integration suite can't run on the box* (20-min): mitigation — targeted `-run` only;
    bounded defer recorded with a trigger (run on CI/capable box before program close), per
    M1–M4 precedent.
- **Test discipline:** testdb factory for every proof (real concurrent for F5.5, not sqlmock);
  targeted `-run` filters; full suite NOT run locally; bounded defers in `evidence.md`.
- **Model policy:** implement via subagents (`superpowers:subagent-driven-development`; sonnet
  implement/review, haiku mechanical, never fable, ≤15 concurrent); main session
  orchestrates/reviews/commits.
- **Commit after verified work; never push; never commit `docs/release/`; plans dir gitignored
  (never force-add).**

## Applicable hard-stops

| ID | What would trip it here |
|----|-------------------------|
| HS-1 | Milestone boundary — operator review gate after validator PASS; no M6, no merge/push without approval. |
| HS-2 | A fix implies redesign outside M5's boundary — e.g. the singleton proof fails and making the watchdog safe needs a cross-binary contract change beyond "register a River periodic job", or migrating staging dispatch forces the render-worker execution pipeline to move too. Stop, report the boundary + minimum prerequisite, no symptom-patch. |
| HS-3 | A prerequisite boundary fails (build / all-4-binaries / system-runnable / River schema migrate) — repair first, rerun, resume. |
| HS-4 | `milestone-validator` returns FAIL — open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Scope drift / off-plan discovery mid-milestone (e.g. a 5th janitor found, a second staging dispatch kind, a shared mutable fanout projection that breaks commutativity) — stop, surface, replan before continuing. |
| HS-7 | Implementation deviates from the committed `validation-contract.md` — fix code to contract, or re-open contract with operator approval; never silently adjust the contract. |
| HS-8 | *(already cleared)* the M5 `developing-new-work` gate returned Yellow, not Red — design was not blocked. Recorded for the audit trail. |
