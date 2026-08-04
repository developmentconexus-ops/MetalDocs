# Approval accountability loop — design

**Date:** 2026-08-04
**Status:** approved by the operator, section by section, 2026-08-04
**Gate:** [`docs/superpowers/analysis/2026-08-04-approval-accountability-loop-system-impact.md`](../analysis/2026-08-04-approval-accountability-loop-system-impact.md) — verdict 🟡 Yellow, amended §2a
**Owning modules:** `approval` (emits) · `notifications` (delivers) · `iam` (names, already wired)

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
`release_coordinator.go:631,647`). So the system notifies on *decision* and never on *submit*.
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

## 2. The three gaps

**G-A — submit emits no notification, and the current envelope cannot express one.**
`LifecycleEventArgs` (`documents/domain/notification_events.go:14`) is document-shaped:
`ControlledDocumentID`, `SubmittedBy`, five `document.*` event types. Its consumer computes
recipients **inside itself**, from either `metaldocs.v_cd_obligated_readers` or the single author
(`fanout_worker.go:93-121`). There is no "notify this explicit set of user ids" path. The worker's
`switch` also ends in `default: return nil` (`fanout_worker.go:56`) — an unrecognised event type is
dropped silently, with no error and no dead-letter.

**G-B — the SLA engine is complete but unreachable through the contract.**
`approval_route_stages.due_in_days` exists in the database with a `CHECK (due_in_days > 0)`
(`db/baseline/0001_current_schema.sql:2112`), the snapshot column exists with the same check
(`:2197`), `due_at` is computed at stage activation from the snapshot, and it is exposed on the stage
view, on the inbox, and as a `due_before` filter. But `internal/modules/approval/http` contains
**zero** references to `due_in_days` — the route-configuration write path never accepted it, and the
field appears in `api/openapi/v1/openapi.yaml` only inside description prose (lines 6691, 7033). It
can only be set by raw SQL. That is why it is NULL everywhere, and why `approval_sla_surfacer` has
completed 18 clean ticks finding nothing.

**G-C — the template 404.** `loadInstanceAreaCode` (`read_service.go:986-999`) INNER JOINs
`documents`. A template instance has a NULL `document_id`, so the pre-check returns no row, yields
`ErrNoActiveInstance`, and the endpoint 404s — hiding a roster that already works. The repository one
layer below it was already corrected to a LEFT JOIN during the M3 kernel extraction and says so in
its comment; the pre-check was missed.

## 3. The decision that shapes everything else

G-A had a cheap path and a correct one, and they differ by about half a day.

**The cheap path** adds `approval.pending` to the existing five event types, leaves
`ControlledDocumentID` empty, and teaches the fanout worker to go fetch the active stage's
`eligible_actor_ids` itself. That makes `notifications` read `approval`'s tables, which violates
non-negotiable invariant 6 (cross-module access only through a published interface) and would
therefore require an ADR recording a MUST-deviation. It also ratifies a document-shaped envelope for
an event that has no document, six months after ADR 0082 moved `approval` out of `documents`
precisely to stop this. Every future `approval` notification inherits the debt.

**The chosen path** puts the resolved recipients in the envelope. The module that resolved them is
the module that knows them: `approval` already snapshots `eligible_actor_ids` per stage.
`notifications` goes back to being delivery only — it receives a list and inserts one row per
recipient. No ADR is needed, because this is the architecture the system already declared.

The existing five `document.*` events and their worker are **not touched**. This is addition, not
replacement: no contract migration, no versioned consumer.

## 4. Architecture

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

Two event types, and no more:
- `approval.pending` — submitted; it is with you.
- `approval.overdue` — the deadline passed; it is still with you.

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

Titles and messages come from the worker's own pt-BR catalogue, the same way the five existing types
get theirs. An unrecognised event type **returns an error** — the inverse of the current
`default: return nil`. River retries and then dead-letters, so there is a trace. Silence was the
defect.

### 4.4 Data flow — submit

Inside the transaction that already exists: write the instance and its stages (`eligible_actor_ids`
is snapshotted there) → emit the audit `GovernanceEvent` (**unchanged**) → **new:** enqueue
`approval.pending` with recipients = the active stage's `eligible_actor_ids` → commit. The worker
runs after commit and inserts N rows.

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

## 5. G-B — making the deadline configurable

One field, three places.

1. **Contract.** `due_in_days` (`integer`, nullable, `minimum: 1`) on the route-configuration stage
   schema in `api/openapi/v1/openapi.yaml`, readable and writable. Full spec regeneration — a partial
   regen is forbidden drift, since any spec edit churns every module's embedded `swaggerSpec`.
2. **Handler.** The route-configuration write path reads and persists it. The boundary validation is
   the friendly first line; the DB `CHECK` is the authority (invariant 5, already satisfied).
3. **Semantics.** Nullable stays nullable and means **no deadline** — never "zero days", never a
   hidden default. That is the no-fallback principle, and the engine already behaves that way:
   `due_at` stays NULL and the surfacer skips the stage.

Frontend: the route-admin `StageCard` gains the field, fitted into the existing card rather than a
redesign.

## 6. G-C — the 404

`loadInstanceAreaCode` becomes a LEFT JOIN on `documents`, matching the repository below it, with a
comment citing the M3 kernel extraction so the next reader sees it as the same correction rather than
an exception.

The non-obvious care: under a LEFT JOIN, `d.process_area_code_snapshot` is NULL for a template.
Downstream code must treat "no document area" as a legitimate case and **must not** fall back to a
default area — a substitute value in an authorization path is the worst class of fallback in this
system. Templates use `asi.area_code_snapshot`, which the query already selects.

## 7. Migrations

**None.** The SLA columns already exist, the event is a River job with no schema, and
`metaldocs.notifications` already exists. `db/migrations/` is untouched, which also means the Class 22
lesson from 0317 has no branch to get wrong here.

## 8. Error handling

- An enqueue failure inside the submit transaction **aborts the submit**. It is not swallowed. If the
  notification cannot be promised, the submission is not confirmed — that is the outbox discipline; a
  half-transaction is worse than an error.
- An unknown event type in the worker **returns an error**, so River retries and then dead-letters.
- A recipient with no `display_name` falls back to their `user_id`, exactly as the existing roster
  does. No invented "Unknown user".
- An empty recipient list emits nothing. Not an error — that is the `livre` route.

## 9. Test plan

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

**QA gates that apply:** contract, authz, multi-tenant isolation, async/idempotency, docs.
**DB-invariant gate: N/A** — no migration.

**Compilation trap to respect:** untagged `go test` does **not** compile `//go:build integration`
files. This feature changes seam signatures (a new port, a new worker), so `go vet -tags integration`
runs before any commit.

## 10. Evidence for closure

Compiling does not count.

Commands: `go build ./...`, `go test ./...`, `go test -tags=integration` (touched packages plus guard
suites, per the selective ladder), `go vet -tags integration`,
`.\scripts\check-system-runnable.ps1`.

Plus **runtime proof** on the container stack, because this feature is runtime-observable: perform a
real submit and show a row in `metaldocs.notifications` and a job in River. That exact check is what
proved the current behaviour — 0 rows, 0 jobs. The same check must return N rows.

## 11. Docs

- `wiki/modules/approval.md` — submit now emits two events, not one; the SLA becomes configurable.
- `wiki/modules/notifications.md` — a second worker; delivery by explicit recipient list.
- `Last verified` refreshed on both.

**ADR: none.** The chosen path is the architecture ADR 0082 already declared, so there is no
MUST-deviation to record. (The cheap path would have required one.)

## 12. Explicitly out of scope

Named so the operator can pull them in deliberately, not so they can be quietly forgotten.

- **BPMN claim step.** Converting pool eligibility into a single named holder changes standing
  approval policy and earns its own ADR. This design does not foreclose it: the roster already
  distinguishes `active` from `waiting` per actor, so "the N in the pool" and "the one who took it"
  can coexist later without a contract break.
- **External channels (email).** Notification stays in-app, in `metaldocs.notifications`, which is
  what exists today. Email brings bounces, opt-out and retry — a different problem and a different
  spec.
- **Escalation after a deadline lapses** (notify a manager, reassign). Requires an org hierarchy
  `iam` does not model.
