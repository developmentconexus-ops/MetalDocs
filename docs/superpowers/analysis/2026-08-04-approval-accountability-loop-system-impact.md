# System-impact analysis — Approval accountability loop

**Date:** 2026-08-04
**Intent (one line):** Close the accountability loop on approval — notify the eligible actors when a subject is submitted, put a per-stage deadline on it with a reminder when it lapses, and show the submitter who is holding it, by name, since when, and until when.
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(see §10)*
**Amended:** 2026-08-04, after a code-level sweep of every impacted module. See §2a — the amendment
materially narrows the work and moves the global-maximum question. Read §2a before §3.

> Same ten sections for module and feature work. Module-only rows are marked **N/A** with a one-line
> reason rather than deleted — the record should show the question was asked.

---

## 1. Classify & own

- **Work type:** feature. It adds no bounded context; every piece lands in a module that already
  exists and already owns the concept.
- **Owning module(s):**
  - `approval` — owns the routing decision (who is eligible), the stage clock (when it started, when
    it is due), and therefore owns *emitting* the fact that a subject is now waiting on named people.
    It already snapshots `eligible_actor_ids` at submit time, so it is the only module that can
    answer "who" without re-deriving the route.
  - `notifications` — owns delivery. It already consumes lifecycle events off the outbox and fans
    them out; this feature gives it two more event kinds, not a new mechanism.
- **Explicitly NOT owning:**
  - `documents` / `controlleddocuments` — they own the subject, not the routing. Putting "who must
    approve" in a document read model would fork the eligibility rule into a second place, which is
    exactly how the current route/seed mismatch went unnoticed.
  - `jobs` — hosts the scheduled tick, but must not own the SLA rule. The existing
    `approval_sla_surfacer` job already delegates the decision to
    `MarkStagesOverdueSurfaced` in `approval`; the reminder follows the same shape.
  - `iam` — supplies display names only. It does not learn anything about approval.
- **Cross-module edges (with direction):** `A → B` = A depends on B.
  - `approval → notifications` — via the published `LifecycleEventEnqueuer` port that `approval`
    already holds (`decision_service.go` uses it today). No new coupling; a new event kind on an
    existing port.
  - `approval → iam` — display-name resolution for the eligible actors. Must go through an `iam`
    application service / published interface. `approval` must never read `iam_users` directly.
  - `jobs → approval` — the periodic tick calls an `approval` application method. Already the
    established direction (ADR 0067 dual-define; the job is defined in `api` and executed in `jobs`).
  - No new edge points *into* `approval` from a module that did not already depend on it.
- **Ambiguity?** None. **AS-3 not raised.**

## 2. Foundation verdict

- **Base you'd build on:** three pieces that already exist and are sound.
  1. `approval` snapshots the resolved eligible actors per stage at submit time (all four selector
     kinds — `named_user`, `role_in_fixed_area`, `role_in_document_area`, `submit_choice` — are
     implemented in the DB CHECK, in the OpenAPI contract, and in dedicated error codes).
  2. `LifecycleEventEnqueuer.EnqueueLifecycleEventTx` enqueues a `notification_fanout` job inside the
     business transaction — the outbox discipline is already correct at the three decision call sites.
  3. `approval_sla_surfacer` already walks tenants under `WithBackgroundBypass`, seeds per-tenant, and
     marks `sla_surfaced_at`.
- **Sound, or legacy/patch/workaround?** **Sound.** This is not a rebuild — it is three wires that were
  never connected on a correctly-built frame. The gap is precisely bounded and verified:
  - Submit emits a `GovernanceEvent` (audit) and **only** that
    (`template_submit_service.go:377`, `submit_service.go:473`). The notification path is a different
    call — `EnqueueLifecycleEventTx` — invoked only from `decision_service.go:748`,
    `document_terminal_approval.go:151`, `obsolete_service.go:167`. So the system notifies on
    *decision* and never on *submit*, for documents and templates alike.
  - Confirmed at runtime, not inferred: `metaldocs.notifications` had 0 rows after two real
    submissions, and River had zero `notification_fanout` jobs, while the SLA surfacer had completed
    18 ticks finding nothing (`due_in_days` is NULL everywhere, so no stage is ever overdue).
- **Would this optimize inside a patch?** No. The one thing that *is* a patch here is the fix bundled
  into this feature: `loadInstanceAreaCode` (`read_service.go:986-999`) INNER JOINs `documents`, so a
  template instance (NULL `document_id`) yields no row and the endpoint 404s. The repository one layer
  below it was already corrected to a LEFT JOIN during the M3 kernel extraction and carries a comment
  saying so; the pre-check was missed. Fixing the pre-check the same way is *completing* the M3
  migration, not entrenching it. **AS-2 not raised.**
- **Global-maximum note carried into design:** the industry answer to "a pool holds it, so nobody
  holds it" is BPMN 2.0's **claim** step — `candidateGroups` (which MetalDocs has, as area+role
  selectors) plus an explicit claim that converts group authority into named accountability. The
  regulatory frame agrees from both ends: ISO 13485 §4.2.4 / ISO 9001 §7.5.2 define approval authority
  by *function*, while 21 CFR Part 11 §11.50 requires the signer's *name* — "route by function,
  account by name." Whether to add a claim step is a design question for brainstorming; this analysis
  only records that visibility-of-the-pool is the prerequisite for it either way, and must not be
  designed in a way that forecloses it.

## 2a. Amendment — what the module-by-module sweep found

§2 was written from the system map and was directionally right but materially wrong about *what is
missing*. A code-level read of `approval`, `notifications`, `documents/domain`, `iam` and the
frontend changes the shape of the work. Runtime truth beats the map, so the amendment governs.

**Already built — do not design it again.**
`GET /approval/instances/{id}` already returns, per stage, `actors[]` carrying `user_id`,
`display_name` and a status of `active` / `waiting` / `approved` / `rejected`, alongside `due_at` and
`submitted_at` (`get_instance_handler.go:56-166`). Display names already resolve through the
iam-owned `UserDisplayNameReader` port with a userID fallback and already do so off-tx
(`get_instance_handler.go:169-210`, comment cites M4/F4.1 and H-PRE-1). The web client already
renders that roster (`frontend/apps/web/src/features/documents/lib/approvalWorkflow.ts:117`). So the
"expose who holds it, by name" half of the intent exists and works — for documents.

**The three real gaps.**

- **G-A — submit emits no notification, and the envelope cannot express one.** Confirmed at runtime
  (0 rows in `metaldocs.notifications`, 0 `notification_fanout` jobs after two real submits). The
  deeper finding is structural: `LifecycleEventArgs` (`documents/domain/notification_events.go:14`)
  is *document-shaped* — `ControlledDocumentID`, `SubmittedBy`, five `document.*` event types — and
  the consumer computes recipients **inside itself**, from either
  `metaldocs.v_cd_obligated_readers` or the single author (`fanout_worker.go:93-121`). **There is no
  "notify this explicit set of user ids" path.** The worker's `switch` also ends in
  `default: return nil` (`fanout_worker.go:56`), so an unrecognised event type is dropped silently,
  with no error and no dead-letter.
- **G-B — the SLA engine is complete but unreachable through the contract.** `due_at` is computed at
  stage activation from `due_in_days_snapshot`, stays NULL when no SLA is configured (no-fallback,
  correct), and is exposed on the stage view, on the inbox, and as a `due_before` filter. But
  `due_in_days` appears in `api/openapi/v1/openapi.yaml` **only inside description prose** (lines
  6691, 7033) — it is not a settable field on any route-configuration schema. It can only be set by
  raw SQL. That is why it is NULL everywhere and why `approval_sla_surfacer` has completed 18 clean
  ticks finding nothing. This gap is contract/configuration work, not engine work.
- **G-C — the template 404** (`read_service.go:986`) hides a roster that already works.

**Where the global maximum actually sits — corrected.**
§2 placed the global-maximum question on the BPMN claim step. That question is real but secondary.
The load-bearing one is G-A: ADR 0082 promoted `approval` to a top-level module, but the lifecycle
event envelope stayed behind in `documents/domain`, still shaped around a controlled document. A
template approval has no controlled document at all. Adding an `approval.pending` event type to that
envelope — with `ControlledDocumentID` empty and the recipient list smuggled in somehow — is
precisely "optimizing inside a patch": it would ratify a boundary that ADR 0082 already moved, and it
would put approval-specific recipient logic inside a worker that today only knows documents.

The global-maximum structure is instead: **an explicit-recipient lifecycle event owned by the module
that knows the recipients.** Concretely — the envelope carries the resolved recipient list, the
emitter (`approval`) supplies it from the `eligible_actor_ids` snapshot it already holds, and
`notifications` stops being the place where "who gets this" is decided for events it does not own.
The trade-off is one migration of the event contract plus a versioned consumer, against permanently
forking recipient resolution across two modules. The operator decides; this analysis records that
taking the cheap path here is a known local maximum, not an oversight.

**AS-2 is raised but resolved-by-choice, not unresolved:** the work *would* optimize inside a patch
if it took the cheap path, so the design must choose the envelope question explicitly rather than
default into it. Recording it here is what keeps the verdict at Yellow rather than Red.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | Yes | The new read (who is eligible / since when / due when) is a field set on the existing approval-instance read, which is already capability-gated. `CapApprovalOversee` (`model.go:99`, `ScopeTenant`) covers tenant-wide oversight; the submitter's own-instance view rides the existing gate on that route. **No new capability** ⇒ no `TestCapabilityRegistrySize` bump. Reason in capabilities, never "the author can see it". | `authz.Require` in-tx |
| Contract-first (OpenAPI + oapi-codegen) | Yes | Response fields for the eligible-actor list, `waiting_since`, `due_at` are added to `api/openapi/v1/openapi.yaml` first, then regenerated. Note the standing rule: any spec edit churns every module's embedded `swaggerSpec` — **full regen only**, partial regen is forbidden drift. | `oapi-codegen`, generated DTOs |
| Multi-tenant pooled | Yes | Every new read and every reminder query is `tenant_id`-predicated and runs under the tx-local GUC. The background reminder is the dangerous case: it must per-tenant seed (`BypassSystem` + `SeedTxTenant`) and carry an **explicit `tenant_id` predicate** — the M6 surfacer bug (§F6.4) was exactly this, and `metaldocs_app` is superuser+BYPASSRLS in dev so RLS will not catch a miss. | `tenant.FromContext`, `SeedTxTenant` |
| Async = transactional outbox | Yes | The submit-time fanout enqueues **inside the business transaction** via the existing `EnqueueLifecycleEventTx`. No network call in the tx; the fanout consumer stays idempotent. Inlining any send would be AS-1. | `LifecycleEventEnqueuer`, outbox repo |
| DB enforces invariants | Already satisfied — **see §2a** | No new constraint is needed. `approval_route_stages.due_in_days` and `approval_stage_instances.due_in_days_snapshot` already exist with `CHECK (… > 0)` (`db/baseline/0001_current_schema.sql:2112`, `:2197`), and `due_at` is already derived at stage activation. The gap is contractual, not schematic. | the existing CHECK constraints |
| Cross-module via published interface only | Yes | `approval → notifications` on the existing port; `approval → iam` via an `iam` application service for names. No repo/SQL/domain reach-through in either direction. | module `ports.go` |

**AS-1 not raised.**

## 4. Capability wiring

**N/A** — the work adds no capability. It widens the payload of reads that are already gated, and the
eligible-actor visibility is bounded by the same gate that already governs seeing the instance at all.
Touchpoint 10 (H-PRE-1) is *not* N/A and is carried in §10.

## 5. Module wiring

**N/A** — no module is born. All four modules involved (`approval`, `notifications`, `iam`, `jobs`)
exist and are wired at the composition root.

## 6. Frameworks to reuse, not reinvent

- `TxRunner` — the submit path is already a writable tx; the fanout enqueue joins it. Do not open a
  second tx.
- `authz.Require` — in-tx tier-2 check on the widened read. Note the standing constraint: it needs a
  **writable** tx (G1, `817abd59`).
- `authz.SeedTxIdentity` / `SeedTxTenant` / `BypassSystem` — the reminder job's per-tenant seeding.
- `LifecycleEventEnqueuer.EnqueueLifecycleEventTx` — the notification path. Do **not** hand-roll a
  second enqueuer for submit.
- `audit.RecordTx` / the existing `GovernanceEvent` emitter — stays as-is. The audit event is correct
  and is not replaced; the fanout is *added alongside* it. Two paths, two purposes.
- `problem.New` / `problem.Write` — any new error uses a registered `problem.Code` from the closed
  vocabulary (ADR 0089). A bare string literal will not compile.
- `testdb.Open` + factory builders — the integration suite.
- The `iam` published name-lookup service — not a hand-rolled `SELECT` against `iam_users`.

## 7. Contract & data

- **OpenAPI-first:** widen the approval-instance read schema (eligible actors with display names,
  `waiting_since`, `due_at`, overdue flag). Full spec regen; no partial. Four stale pre-0089 code
  names remain in openapi.yaml *prose* (lines 1273, 1289, 1290, 5710) — out of scope here, noted so
  the next spec edit does not treat them as truth.
- **Migration: none — superseded by §2a.** This row originally planned migration 0318 for the stage
  deadline. The sweep found the columns already exist (`approval_route_stages.due_in_days` and
  `approval_stage_instances.due_in_days_snapshot`, both with `CHECK (… > 0)`), the event is a River
  job with no schema, and `metaldocs.notifications` already exists. `db/migrations/` stays untouched,
  which also means the Class 22 lesson from 0317 has no branch here to get wrong.
- **Destructive change?** No. Purely additive fields; the 404 fix is a JOIN correction with no schema
  change. The one behavioural break is intentional and desirable: submits start producing
  notifications where they previously produced silence.

## 8. Test & QA plan

- **Canonical framework:** `tests/integration/testdb/` — `Open(t)`, factory builders, `SeedWithCaps`,
  `Qualified`, tagged `//go:build integration`. R1–R4 per `check-test-discipline.sh`. No bespoke
  harness.
- **QA gates that apply** (feature ⇒ subset, the rest named N/A):
  - **Contract** — applies. Spec-first + full regen + generated DTOs on the consumer side.
  - **AuthZ** — applies. Prove a non-eligible, non-oversight actor cannot see the eligible list; prove
    the tier-2 in-tx check runs.
  - **Multi-tenant isolation** — applies, and is the highest-risk gate. Cross-tenant read → 404; the
    reminder job must not leak across tenants. Test with the non-owner `metaldocs_ci` role, because
    `metaldocs_app` is BYPASSRLS in dev and will produce a false green.
  - **Async / idempotency** — applies. Two submits must not double-notify; the fanout consumer is
    idempotent; the reminder must not re-fire every tick for the same stage.
  - **DB-invariant** — applies narrowly (the deadline CHECK).
  - **Docs** — applies (§9).
- **Integration-tag compile gap:** untagged `go test` does not compile `//go:build integration`
  files. After any seam signature change, run `go vet -tags integration` before committing.
- **Evidence shape:** `go build ./...`, `go test ./...`, `go test -tags=integration ./...` (touched
  packages + guard suites per the selective ladder), `.\scripts\check-system-runnable.ps1`, plus
  **runtime proof** — this feature is runtime-observable, so closure requires driving a real submit in
  the container stack and showing rows in `metaldocs.notifications` and a `notification_fanout` job in
  River. Not "the code enqueues it."

## 9. Docs / ADR

- **Wiki:** update `wiki/modules/approval.md` (the submit path now emits two events, not one; the SLA
  clock becomes load-bearing) and `wiki/modules/notifications.md` (new event kinds). Refresh both
  `Last verified` stamps. No new module doc.
- **REQ IDs cited:** to be pinned against `wiki/architecture/backend-target-architecture.md` in the
  spec — the async/outbox, multi-tenant isolation, and authz REQ IDs are the ones this feature must
  answer to. The spec cites them explicitly; a review that cannot cite them is not a review.
- **ADR required?** **Conditionally yes — this is the Yellow.** No ADR is needed to *connect* the
  existing wires: emitting a fanout event at submit, honouring an SLA that the schema already models,
  and fixing a JOIN are all in-bounds work against MUSTs the system already accepts. But if the design
  chooses to add a **claim step** (converting pool eligibility into a single named holder), that
  changes standing approval policy — the meaning of "eligible" and the relationship between routing
  and accountability — and requires its own ADR, superseding nothing but extending ADR 0082's kernel
  ruling. Decide in brainstorming, before writing the spec, not after.

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — proceed to design. The foundation is sound and no invariant is
  violated. Yellow, not Green, for one reason: the design may reach for a claim step, and that is a
  policy change requiring an ADR. Nothing blocks design; the ADR question must be *decided* in
  brainstorming rather than discovered in implementation.
- **Open hard-stops:** none. AS-1 no, AS-2 no, AS-3 no.
- **Locked constraints handed to brainstorming** — the design must honour all of these:
  1. **Submit-time fanout enqueues inside the business transaction**, reusing
     `LifecycleEventEnqueuer.EnqueueLifecycleEventTx`. No second enqueuer, no network call in the tx.
  2. **The audit `GovernanceEvent` stays.** The fanout is added alongside it. Do not repurpose the
     audit path as a delivery path — they are two events with two consumers and two retention rules.
  3. **Recipients come from the existing `eligible_actor_ids` snapshot.** Never re-resolve the route
     at notify time — re-resolution can disagree with what was recorded, and the snapshot is what the
     audit trail asserts.
  4. **`approval → iam` for display names only, through a published interface.** No direct read of
     `iam_users`.
  5. **H-PRE-1:** display-name resolution and any authz-recording read stay **off** the lock-holding
     transaction. Never call an authz-recording read inside a lock-holding atomic tx.
  6. **The reminder job seeds per tenant and carries an explicit `tenant_id` predicate.** RLS is inert
     in dev (`metaldocs_app` is superuser + BYPASSRLS) and will not catch the omission — the M6
     surfacer bug is the precedent.
  7. **No new capability.** If the design finds it needs one, that is a signal the read was placed on
     the wrong route — revisit the placement before adding a cap.
  8. **Contract-first with a full spec regen.** Partial regen is forbidden drift.
  9. **Migration 0318**: `to_regclass` every referenced table up front, schema-qualify explicitly, and
     test the branch against a database seeded into the state the migration repairs (Class 22).
  10. **The 404 fix (`read_service.go:986`) is completing the M3 kernel extraction**, not a workaround —
      match the LEFT JOIN the repository below it already uses, and say so in the comment.
  11. **Closure requires runtime evidence**, not compile evidence: a real submit in the container
      stack producing a row in `metaldocs.notifications` and a `notification_fanout` job in River.
  12. **Do not foreclose a claim step.** Whether or not this feature adds one, the data shape must
      allow "the pool of N" and "the one who took it" to coexist later.
