# Approval accountability loop — design

**Date:** 2026-08-04
**Status:** approved by the operator, section by section, 2026-08-04
**Gate:** [`docs/superpowers/analysis/2026-08-04-approval-accountability-loop-system-impact.md`](../analysis/2026-08-04-approval-accountability-loop-system-impact.md) — verdict 🟡 Yellow, amended §2a
**Owning modules:** `approval` (emits, extends) · `notifications` (delivers) · `iam` (names, new capability) · `jobs` (tick)

---

## 1. The problem, stated as the operator found it

Submitting a template for approval tells nobody. The submitter cannot see who it went to, so there is
no one to ask and no one to hold accountable. The operator's words: if this were a real company, how
would I chase anyone?

The investigation found the complaint understated the problem in one direction and overstated it in
another.

**Understated:** nothing is notified on submit — not for templates, not for documents. `approval`
emits a `GovernanceEvent` at submit (`template_submit_service.go:377`, `submit_service.go:473`),
which is an **audit** record. The notification path is a different call,
`LifecycleEventEnqueuer.EnqueueLifecycleEventTx`, invoked only from decision sites
(`decision_service.go:748`, `document_terminal_approval.go:151`, `obsolete_service.go:167`,
`release_coordinator.go:631,647`). The system notifies on *decision* and never on *submit*.
Confirmed at runtime, not inferred: zero rows in `metaldocs.notifications` and zero
`notification_fanout` jobs in River after two real submissions.

**Overstated:** the "show me who holds it, by name" half is already built and working — for
documents. `GET /approval/instances/{id}` already returns, per stage, `actors[]` with `user_id`,
`display_name`, and a status of `active` / `waiting` / `approved` / `rejected`, plus `due_at` and
`submitted_at` (`get_instance_handler.go:56-166`). Names already resolve through the iam-owned
`UserDisplayNameReader` port with a userID fallback, already off-tx per H-PRE-1
(`get_instance_handler.go:169-210`). The web client already renders that roster
(`frontend/apps/web/src/features/documents/lib/approvalWorkflow.ts:117`).

It is invisible for templates because of a JOIN bug, and the deadline beside it is always empty
because the field that sets it was never exposed on the contract.

## 2. The gaps

**G-A — submit emits no notification, and the current envelope cannot express one.**
`LifecycleEventArgs` (`documents/domain/notification_events.go:14`) is document-shaped:
`ControlledDocumentID`, `SubmittedBy`, five `document.*` event types. Its consumer computes
recipients **inside itself**, from either `metaldocs.v_cd_obligated_readers` or the single author
(`fanout_worker.go:93-121`). There is no "notify this explicit set of user ids" path. The worker's
`switch` also ends in `default: return nil` (`fanout_worker.go:56`) — an unrecognised event type is
dropped silently, with no error and no dead-letter.

**G-B — the SLA engine is complete but unreachable through the contract.**
`approval_route_stages.due_in_days` exists with `CHECK (due_in_days > 0)`
(`db/baseline/0001_current_schema.sql:2112`); the snapshot column exists with the same check
(`:2197`); `due_at` is computed at stage activation from the snapshot; it is exposed on the stage
view, on the inbox, and as a `due_before` filter. But `internal/modules/approval/http` contains
**zero** references to `due_in_days` — the route-configuration write path never accepted it, and it
appears in `api/openapi/v1/openapi.yaml` only inside description prose (lines 6691, 7033). It can
only be set by raw SQL. That is why it is NULL everywhere, and why `approval_sla_surfacer` has
completed 18 clean ticks finding nothing.

**G-C — the template 404.** `loadInstanceAreaCode` (`read_service.go:986-1001`) INNER JOINs
`documents`. A template instance has a NULL `document_id`, so the pre-check returns no row, yields
`ErrNoActiveInstance`, and the endpoint 404s — hiding a roster that already works. The repository one
layer below was already corrected to a LEFT JOIN during the M3 kernel extraction and says so in its
comment; the pre-check was missed.

**G-D — a deadline set on the route cannot be varied for one case.** Added at the operator's
request. Without per-instance extension, the only way to give one document longer is to create
another route — multiplying route rows to express an exception. The route stays the standard; the
extension is the recorded exception.

## 3. The decision that shapes everything else

G-A had a cheap path and a correct one, and they differ by about half a day.

**The cheap path** adds `approval.pending` to the existing five event types, leaves
`ControlledDocumentID` empty, and teaches the fanout worker to fetch the active stage's
`eligible_actor_ids` itself. That makes `notifications` read `approval`'s tables, violating
non-negotiable invariant 6 (cross-module access only through a published interface), and would
require an ADR recording a MUST-deviation. It also ratifies a document-shaped envelope for an event
that has no document, six months after ADR 0082 moved `approval` out of `documents` precisely to stop
this. Every future `approval` notification inherits the debt.

**The chosen path** puts the resolved recipients in the envelope. The module that resolved them is
the module that knows them: `approval` already snapshots `eligible_actor_ids` per stage.
`notifications` goes back to being delivery only. No ADR is needed, because this is the architecture
the system already declared.

The existing five `document.*` events and their worker are **not touched**. Addition, not
replacement: no contract migration, no versioned consumer, and no way for this feature to break
document-published notification, because it does not touch that code.

## 4. Architecture — the notification loop

### 4.1 The envelope — `approval/domain`

`ApprovalNotificationArgs`, with its own `Kind()` (a distinct River job kind from
`notification_fanout`):

| Field | Purpose |
|---|---|
| `EventID` | uuid minted at emit — the idempotency key |
| `TenantID` | single-tenant by construction, as the existing envelope is |
| `EventType` | `approval.pending` or `approval.overdue` |
| `SubjectKind` | `document` or `template` — the M3 kernel's subject discrimination |
| `SubjectID` | the subject key |
| `RecipientUserIDs` | **the resolved list.** The field that makes the boundary hold |
| `ActorUserID` | who submitted |
| `OccurredAt` | emit time |

Two event types, and no more: `approval.pending` (submitted; it is with you) and `approval.overdue`
(the deadline passed; it is still with you).

### 4.2 The port — `approval/domain`

`ApprovalNotificationEnqueuer`, with its River adapter in `approval/jobs` beside the existing
`RiverLifecycleEventEnqueuer`. Same shape, sibling file. It takes `db.Tx` so the domain stays
infra-free, exactly as the existing port does.

### 4.3 The consumer — `notifications/infrastructure`

A second `river.Worker`, registered next to the current one at
`apps/jobs/cmd/metaldocs-jobs/main.go:127`. It is **delivery only**: begin tx, `SeedTxTenant`,
iterate `RecipientUserIDs`, insert with the same
`ON CONFLICT (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL DO NOTHING`
that already guarantees idempotency. **No `approval` SQL.** That is what keeps invariant 6 standing.

Titles and messages come from the worker's own pt-BR catalogue, as the five existing types do. An
unrecognised event type **returns an error** — the inverse of the current `default: return nil`.
River retries and then dead-letters, so there is a trace. Silence was the defect.

### 4.4 Data flow — submit

Inside the transaction that already exists: write the instance and its stages (`eligible_actor_ids`
is snapshotted there) → emit the audit `GovernanceEvent` (**unchanged**) → **new:** enqueue
`approval.pending` with recipients = the active stage's `eligible_actor_ids` → commit. The worker
runs after commit and inserts one row per recipient.

The `livre` route (zero stages, auto-approve, ADR 0087) resolves itself: no active stage means an
empty recipient list means nothing is emitted. The degenerate case needs no special branch.

### 4.5 Data flow — reminder

`approval_sla_surfacer` already walks tenant by tenant under `WithBackgroundBypass`, seeds each
tenant, and marks `sla_surfaced_at` on overdue stages. It gains one effect: enqueue
`approval.overdue` for each stage it marked, recipients = that stage's `eligible_actor_ids`.

No new "already reminded" guard is invented. `sla_surfaced_at` **is** that guard — the job only marks
stages not yet marked, so the reminder fires once. Reuse, not reinvention.

### 4.6 What `notifications` knows about `approval` when this is done

Nothing. It receives a tenant, a recipient list, and an event type.

## 5. G-B — making the deadline configurable on the route

One field, three places.

1. **Contract.** `due_in_days` (`integer`, nullable, `minimum: 1`) on the route-configuration stage
   schema in `api/openapi/v1/openapi.yaml`, readable and writable. Full spec regeneration — a partial
   regen is forbidden drift, since any spec edit churns every module's embedded `swaggerSpec`.
2. **Handler.** Only the HTTP boundary is missing, and the gap is narrower than it looks:
   `domain.Stage.DueInDays` already exists (`approval/domain/route.go:62`) and the service already
   persists and re-reads it (`route_admin_service.go:992,1106`). What drops the value is the wire
   mapping — `mapStageRequests`, `mapStagesToResponse` and `mapListRoute`
   (`approval/http/route_admin_handler.go`) simply never carry the field. Boundary validation is the
   friendly first line; the DB `CHECK` is the authority (invariant 5, already satisfied).
3. **Semantics.** Nullable stays nullable and means **no deadline** — never "zero days", never a
   hidden default. That is the no-fallback principle, and the engine already behaves that way:
   `due_at` stays NULL and the surfacer skips the stage.

**Route rows are per subject kind, and that is deliberate.** `approval_routes` carries `subject_kind`
constrained to `'document'` or `'template'` (`db/baseline/0001_current_schema.sql:2136-2141`). The
same table, same stage configuration and same `due_in_days` field serve both — but a document profile
and a template profile are separate route rows. This is the M3 subject-generic kernel, and it is
correct: approving *the template of a Procedure* and approving *a filled-in Procedure* legitimately
warrant different stages, approvers and deadlines. Collapsing them into one row would lose capability.

Consequence to surface in the UI, not to hide: configuring 30 days on the document route does not
configure it on the template route. If that proves awkward, the fix belongs in the screen (show both,
or offer "copy configuration") — never in the data model.

Frontend: the route-admin `StageCard` gains the field, fitted into the existing card rather than a
redesign.

## 6. G-D — per-instance extension

The route is the standard. The extension is the recorded exception. Its reason for existing is
precisely to avoid multiplying route rows to express a one-off.

**Shape.** `POST /api/v1/approval/instances/{instance_id}/extend-sla`, body carrying the new due date
and a **mandatory reason**. It applies to the instance's **active** stage.

**Rules, each with its reason:**

- **Forward only.** The new date must be strictly after the current `due_at`. Shortening a deadline
  is a different act with different governance consequences and is not in scope (YAGNI — the operator
  asked to postpone).
- **Reason mandatory.** Extending a deadline in a regulated system is a governance event. Reuses the
  existing reason-for-change facility from GMR M6; no parallel mechanism.
- **Audited.** Who extended, from when to when, and why — through the existing `GovernanceEvent`
  emitter. **No new history table:** the audit event *is* the history, and a table duplicating it
  would be exactly the redundancy §9 forbids.
- **`sla_surfaced_at` is left alone — deliberately.** The obvious worry is that a stage which already
  lapsed carries a "reminded" mark, so extending it would silently disarm the reminder. It does not,
  because the marker is **per cycle, not boolean**: the surfacer's eligibility predicate is
  `sla_surfaced_at IS NULL OR sla_surfaced_at < due_at` (`approval/domain/sla_port.go:52`). Moving
  `due_at` forward makes that predicate true again, so the reminder re-arms by construction. Clearing
  the marker would be strictly worse — it erases the evidence that the first reminder was sent. A
  test pins this, because the behaviour is load-bearing and non-obvious.
- **No `If-Match`.** The forward-only check reads and writes the same row inside one transaction, so
  concurrent extensions are already correct. An OCC precondition would be contract surface solving a
  problem the transaction already solves.
- **`Idempotency-Key` required**, failing closed before the service is invoked — the module's
  standing mutation contract (`submit_handler.go:18`).

**Authorization — a new capability, `approval.sla_extend`.**

This is the one number that moved. `approval.oversee` cannot be reused: its own declaration says
*read-only oversight* (`internal/modules/iam/domain/model.go:95-99`). Authorizing a write with a
capability declared read-only is precisely the silent erosion §9 forbids.

All ten capability touchpoints apply:

| # | Touchpoint | This feature |
|---|---|---|
| 1 | const + `validCapabilities` | `CapApprovalSLAExtend = "approval.sla_extend"` (`iam/domain/model.go`) |
| 2 | scope classify | `ScopeArea` — extension is judged in the subject's area, like `approval.review` |
| 3 | tier-1 route→cap | **mandatory.** The generic `/api/v1/approval/` fallback was deleted (`permissions.go:266-271`), so an unmapped POST falls through to `VisibilitySessionRequired` — reachable by any authenticated session. This is a real privilege escalation, not a theoretical one |
| 4 | tier-2 in-tx | `authz.Require(ctx, tx, CapApprovalSLAExtend, areaCode)` after `SeedTxIdentity` |
| 5 | seed grants | `db/reference-data/0001_product_reference_data.sql` |
| 6 | DB tripwire | **N/A, deliberately.** `approval_stage_instances` is not a gated table (`internal/platform/tripwire/arms.go` arms only `approval_instances` and `approval_signoffs`). Arming it would break the background jobs that write it under bypass. Considered and rejected, not overlooked — so no tripwire re-render and no forward migration |
| 7 | guard tests | `TestEveryCapabilityClassified`, `TestAreaGradeCapabilitySet` stay green |
| 8 | registry size | `TestCapabilityRegistrySize`: **38 → 39** (`iam/domain/model_test.go:96`), appending the reason to the running ledger comment in its established style |
| 9 | CI capability-coherence | the five surfaces must agree (REQ-AUTHZ-5) |
| 10 | H-PRE-1 | no authz-recording read inside a lock-holding tx |

**Error codes**, registered in `approval/http/errors.go` in the existing
`problem.Register("approval", "<family>.<name>", status)` form — the vocabulary is closed (ADR 0089),
so a bare string does not compile:
- `validation.sla_extension_not_forward` (422)
- `state.sla_extension_no_active_stage` (409)

## 7. G-C — the 404

`loadInstanceAreaCode` becomes a LEFT JOIN on `documents`, matching the repository below it, with a
comment citing the M3 kernel extraction so the next reader sees the same correction rather than an
exception.

The function's existing precedence chain already does the right thing for templates: it prefers
`asi.area_code_snapshot` (the active stage), then the document snapshot, then the controlled-document
area, and bakes in **no** empty-area default — callers make the "area filter off" decision explicitly
at the call site. Under a LEFT JOIN a template simply resolves on the first branch.

The care that must not be lost: `d.process_area_code_snapshot` is NULL for a template, and downstream
code must treat "no document area" as legitimate and **never** substitute a default area. A
substitute value in an authorization path is the worst class of fallback in this system. A test
covers the completed-template case (no active stage), where the chain yields `""` and the caller's
explicit `"" → "tenant"` widening must be proven to *tighten* rather than loosen access.

## 8. Migrations

**None.**

- The SLA columns already exist, with their CHECK constraints.
- The events are River jobs — no schema.
- `metaldocs.notifications` already exists, with its idempotency index.
- The new capability needs no tripwire arm (§6, touchpoint 6), so no golden re-render.

`db/migrations/` is untouched. This also means the Class 22 lesson from 0317 — the guarded branch no
environment executes — has no branch here to get wrong.

## 9. Implementation rules — binding, not aspirational

Recorded at the operator's explicit instruction, because a promise in conversation evaporates and a
rule in a spec is enforceable at review.

1. **Nothing is born optional for compatibility.** When a consumer must change, it changes. No
   relaxing a field to optional so an old caller keeps working.
2. **No field enters without a consumer.** A field nobody reads does not ship, however convenient it
   seems while the file is open.
3. **Legacy found in a touched path is deleted, not worked around** — and the deletion is reported,
   not buried in the diff.
4. **Zero fallback on an integrity path.** No default area, no default deadline, no invented display
   name.

**Deletions this feature commits to, named now rather than promised:**

- `default: return nil` in `fanout_worker.go:56` — an unrecognised event type currently vanishes with
  no error and no dead-letter. It becomes an error.
- The four pre-ADR-0089 code names surviving in `openapi.yaml` prose (`APPROVAL_ROUTE_MISSING`,
  `ALREADY_EXISTS`; lines 1273, 1289, 1290, 5710). The real codes are `state.approval_route_missing`
  and `conflict.already_exists`. A contract that lies is worse than a contract that is silent.

The principle also shaped the design before it was demanded, in two places worth pointing at: no
"already reminded" flag was invented because `sla_surfaced_at` already is one (two flags that must
agree are a bug with a date on it), and no extension-history table is created because the audit event
already is that history.

## 10. Error handling

- An enqueue failure inside the submit transaction **aborts the submit**. Not swallowed. If the
  notification cannot be promised, the submission is not confirmed — that is the outbox discipline; a
  half-transaction is worse than an error. The operator has seen and accepted the visible cost: a
  rare "failed, try again" instead of a silent success that notifies nobody.
- An unknown event type in the worker **returns an error**, so River retries and then dead-letters.
- A recipient with no `display_name` falls back to their `user_id`, as the existing roster does. No
  invented "Unknown user".
- An empty recipient list emits nothing. Not an error — that is the `livre` route.

## 11. Test plan

Canonical framework only: `tests/integration/testdb/` — `Open(t)`, factory builders, `SeedWithCaps`,
tagged `//go:build integration`, R1–R4 discipline. No bespoke harness.

| # | Test | What it proves |
|---|---|---|
| 1 | Template submit emits, with the right recipients | the case that started this |
| 2 | Document submit emits | the working path does not regress |
| 3 | Same `EventID` twice → one row | the `ON CONFLICT` idempotency |
| 4 | Worker tenant isolation, run as **`metaldocs_ci`** | `metaldocs_app` is superuser+BYPASSRLS in dev and gives a false green — this is how the M6 surfacer bug passed |
| 5 | `livre` route emits nothing | zero stages, zero notifications |
| 6 | `due_in_days` NULL ⇒ `due_at` NULL ⇒ surfacer skips | the no-fallback principle |
| 7 | Two surfacer ticks → one reminder | `sla_surfaced_at` is the guard |
| 8 | `GET /approval/instances/{id}` for a template returns 200 with populated `actors[]` | the 404 is dead |
| 9 | Unknown event type returns an error, not `nil` | the silent-drop defect is not reproduced |
| 10 | Extension without a reason is rejected | governance, not decoration |
| 11 | Extension to an earlier date is rejected | forward-only |
| 12 | Extension after lapse re-arms the reminder on the new date, with `sla_surfaced_at` untouched | the per-cycle marker in §6 |
| 13 | A principal without `approval.sla_extend` gets 403 at both tiers | the capability is real, not decorative |
| 14 | Completed template instance resolves to `""` area and the caller's widening tightens access | the §7 care |

**QA gates that apply:** contract, authz, multi-tenant isolation, async/idempotency, docs.
**DB-invariant gate: N/A** — no migration.

**Compilation trap to respect:** untagged `go test` does **not** compile `//go:build integration`
files. This feature changes seam signatures (a new port, a new worker, a new capability), so
`go vet -tags integration` runs before any commit.

## 12. Evidence for closure

Compiling does not count.

Commands: `go build ./...`, `go test ./...`, `go test -tags=integration` (touched packages plus guard
suites, per the selective ladder), `go vet -tags integration`, the api-lint suite (tripwire parity,
capability coherence), `.\scripts\check-system-runnable.ps1`.

Plus **runtime proof** on the container stack, because this feature is runtime-observable: perform a
real submit and show a row in `metaldocs.notifications` and a job in River. That exact check is what
proved the current behaviour — 0 rows, 0 jobs. The same check must return N rows.

## 13. Docs

- `wiki/modules/approval.md` — submit emits two events, not one; the SLA becomes configurable;
  per-instance extension and its capability.
- `wiki/modules/notifications.md` — a second worker; delivery by explicit recipient list.
- `Last verified` refreshed on both.

**ADR: none.** The chosen envelope path is the architecture ADR 0082 already declared, and the new
capability is ordinary in-bounds wiring, not a policy change. (The cheap path would have required
one.)

## 14. Numbers

| | |
|---|---|
| Migrations | **0** |
| ADRs | **0** |
| New capabilities | **1** — `approval.sla_extend`, registry 38 → 39 |
| Existing events touched | **0** |
| Tests | **14** |
| Estimate | **4.5–5.5 days** |

## 15. Deliberately out of scope — and what comes next

Named so the operator pulls them in deliberately, never so they are quietly forgotten.

**Spec 2 — assignment and escalation.** Planned, with its shape already decided.

Assignment is the answer to "Maria has three open, João has none". It presupposes something that does
not exist yet: an *owner*. Today a stage has a pool, and "Maria is holding three" is not a sentence
the system can form. So load-balancing is not a feature beside the claim step — it is a *policy* of
the assignment the claim step introduces.

The form is settled and it is the house pattern, not an invention: **a closed set of named
strategies, selected by configuration** — exactly how route selectors already work (`named_user`,
`role_in_fixed_area`, `role_in_document_area`, `submit_choice`, with a DB CHECK). Concretely
`assignment_strategy` on the stage (`pool` / `least_loaded` / `round_robin`) and `on_overdue`
(`notify_assignees` / `notify_oversight` / `reassign`), each backed by a Go interface so a new
strategy is a small, tested, versioned change.

**Customer-authored logic is explicitly rejected.** It is the Sankhya trade-off: total power, and
every customer becomes a snowflake nobody can support or upgrade. In a regulated eQMS it is worse —
customer code on the approval path means each customer must re-validate, so you no longer ship
validated software.

Two questions spec 2 must answer before it is designed, both capable of making the system worse if
skipped: **less load of what, exactly** (open assignments? including blocked ones? across areas?) and
**what if João is on holiday** — balancing without availability routes everything to the absent
person, which is strictly worse than a pool. Delegation already exists as a concept
(`get_instance_handler.go:109`) and is where availability should attach.

One correction carried forward: escalation was earlier said to be impossible without an org chart.
That holds only for "notify Maria's manager". **"Escalate to whoever oversees" is expressible today**
via `approval.oversee`, tenant-scoped (`iam/domain/model.go:99`) — which in a quality organisation is
usually exactly who should learn that something has been stuck for thirty days.

**Spec 1 is a prerequisite of spec 2, not a neighbour:** assigning without notifying is worse than a
pool; escalating requires a deadline that today is NULL everywhere; and judging whether a
distribution is fair requires seeing the roster. Spec 1 blocks none of it — the envelope already
carries an explicit recipient list, so when assignment exists the list becomes "the assignee" without
a contract break.

**Email.** Accepted as coming, deliberately after the platform works. Notification stays in-app for
now. Email brings a sending service, bounce handling, legally-required opt-out and retry — its own
spec, not a trailing `if`.
