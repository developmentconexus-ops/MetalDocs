# Milestone 3 — Notifications (full-stack; surface + document-lifecycle emitters)

> **Program:** frontend-screen-completion  ·  **Governing spec:** `../mission.md`
> **Status:** In progress — F3.1 + F3.2 **closed** (contract + read surface). **F3.3 RE-SCOPED
> 2026-06-22** per **ADR-0044**: the producer is no longer a projector-over-audit-log with compensating
> views — it is a **domain-event pattern** (owning modules emit typed domain events carrying outcome +
> recipient IDs via River **in-tx**; the notifications module subscribes and fans out idempotently).
> **HS-2 ("publish path sacred") explicitly lifted by the operator** for this redesign — the five emit
> sites gain an additive in-tx River enqueue. ADR-0043 §6 (the two compensating views) is **superseded**.
> See `f3.3-lifecycle-emitter/research-and-design.md` (evidence) + ADR-0044 (decision). The bundle table
> and F3.3 row below are updated to the ADR-0044 domain events.
> *Subagent model: Sonnet 4.6 (operator directive 2026-06-21, carried from M2).*
> **Authored:** 2026-06-22 — *before any feature in this milestone began.* Scope fork (producer
> question) resolved by operator **2026-06-22: option B — surface + real emitters**, then **widened at
> the HS-1 start gate (2026-06-22) from one emitter to the document-lifecycle bundle** (5 emitters whose
> recipient sets already resolve — *the start-gate "outbox-projector" mechanism was subsequently
> re-scoped to ADR-0044 typed domain events; see the status banner above*; approver-routing + templates +
> channels + prefs remain parked).

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates the
> milestone against *this* document.

## Objective

After M3, the Notifications center (`/notifications`, `NotificationsPage`) shows **real notifications
for the authenticated user** from a real, Grade-A backend — the stub is gone end-to-end. The operator
opening the screen sees a live inbox: each notification has a real title/message/event-type/timestamp
and a per-user read-state, **mark-as-read** persists, and the unread-count badge reflects truth. The
noop `notifications.ts` (returns `{ items: [] }`) and the empty `NotificationsPage` (renders `[]`) are
replaced by real TanStack Query hooks against new endpoints.

There are **five real producers** — the **document-lifecycle bundle**, each implemented (per
[ADR-0044](../../../../../wiki/decisions/0044-domain-event-pattern-and-river-dispatch.md)) as a **typed
River-job domain event** the owning module enqueues **in the state-change tx**, consumed by **one
idempotent notifications fan-out worker** that resolves an already-resolvable recipient set from the
event payload + the published view (additive in-tx enqueue at each emit site; publish/approval
*semantics* unchanged):

**Domain events (ADR-0044 — emitted by the owning module in the state-change tx; payload carries
outcome + recipient IDs; superseding the audit-event projection below):**

| Domain event (NEW; owner emit site) | Recipient set | Resolved by the fan-out worker via | Inbox message (pt-BR) |
|--------------------------------|---------------|---------------------------|------------------------|
| `document.published` (`publish_service.go`) | obligated readers | published `metaldocs.v_cd_obligated_readers` (event carries `controlled_document_id`) | "Novo documento controlado para leitura" |
| `document.superseded` (`supersede_service.go`) | obligated readers | `v_cd_obligated_readers` | "Documento substituído por nova revisão" |
| `document.obsoleted` (`obsolete_service.go`) | obligated readers | `v_cd_obligated_readers` | "Documento que você lê foi obsoletado" |
| `document.approved` (`decision_service.go` @ `InstanceApproved`) | submitter | `submitted_by` carried in the event | "Seu documento foi aprovado" |
| `document.rejected` (`decision_service.go` @ `InstanceRejected`) | submitter | `submitted_by` carried in the event | "Documento rejeitado — ajustes solicitados" |

> **Why these names changed (recon truth, 2026-06-22):** the old audit events don't mean what the
> first draft assumed — `signoff.rejected` is an *eligibility-failure* (not document rejection), and
> `signoff_recorded` fires per-stage with no terminal marker. The new `document.approved`/
> `document.rejected` domain events fire **only at the instance terminal transition** — the clean
> final-outcome signal. Reader events carry `controlled_document_id` (resolved via the approval
> module's existing `CDFieldReader` port), so **`v_document_cd_mapping` is unnecessary**; author events
> carry `submitted_by` (already in `instance.SubmittedBy`), so **`v_approval_instance_submitter` is
> unnecessary**. Both ADR-0043 §6 views are eliminated.

So the screen renders genuinely-live rows across the real document lifecycle the first time any of these
events fires after M3, not a fabricated or seeded demo. The five share one mechanism — a typed domain
event + one fan-out worker that switches on `event_type` to a recipient query; the marginal cost of each
additional lifecycle event after the first is a typed event + an emit-site enqueue + a worker case + an
integration assertion.

**Out of scope for this milestone (parked, not faked):** the **approver-routing** trigger ("a document
or template needs *your* approval" → next approver) is **deferred-with-trigger** to the parked
**notification-emitter mission** because **no eligible-approver resolver exists in the codebase** (grep
empty) — it requires net-new approval-routing logic, unlike the five bundle events whose recipients
already resolve. Also parked: template-approval lifecycle events, `route.config.*` → admins, reminder/SLA
(needs the parked read-tracking numerator), all out-of-app delivery (email/push), per-channel
preferences, digests, and any real-time stream rebuild. The screen does not invent notifications it has
no producer for; absence of those event types is correct, not a gap.

**Quality bar moved:** the program's "real API data, no empty no-API shell, no `MOCK_`/stub" bar
(`wiki/quality/screen-definition-of-done.md` D2 + mission §8) goes from **violated** on Notifications
(the page renders a hard-coded `[]` behind a noop API) to **met**: a real query hook, real rows from a
real producer, working mark-read, redesign tokens, both reviewer agents APPROVE on record. Re-measured
by: `grep -nE "items: \[\]|never\[\]|return \{ items: \[\] \}" notifications/api/notifications.ts` = 0
(the noop is deleted), plus runtime proof a published-document notification appears in the recipient's
inbox and mark-read flips its state.

### Operator-locked scope decision (2026-06-22) — the producer fork

Runtime truth at authoring time (recon this session, HEAD on `main`): the FE consumer contract
(`NotificationItem` in `frontend/apps/web/src/lib/types/index.ts:178` — `id, recipientUserId,
eventType, resourceType, resourceId, title, message, status PENDING|SENT|READ, createdAt, readAt?`)
exists and a `notifications` unread-count query key is reserved, but **no backend notification table,
module, endpoint, or producer exists**, and **audit/governance events are actor-centric**
(`Event.ActorID`, `GovernanceEvent.ActorUserID`) — they record *who acted*, never *who should be
notified*. A list/mark-read surface built over a per-recipient table therefore has a real schema and,
without a producer, **nothing real to list** (the M2 numerator trap restated).

The operator was shown the fork and **chose option B — surface + real emitter(s)**, then at the HS-1
start gate (2026-06-22) **widened the emitter scope to the document-lifecycle bundle**:
- M3 builds the read surface (table + list/unread-count/mark-read + wire) **and** wires the **five**
  document-lifecycle producers above (five typed domain events + one notifications fan-out worker, per
  [ADR-0044](../../../../../wiki/decisions/0044-domain-event-pattern-and-river-dispatch.md)) so the
  screen demonstrably renders live data across the real lifecycle.
- The **approver-routing** trigger and all other triggers are **parked** to a designed
  `notification-emitter` mission, rendered as their honest absence (no event of that type yet), never
  fabricated.

**Selection criterion (re-scoped 2026-06-22 per [ADR-0044](../../../../../wiki/decisions/0044-domain-event-pattern-and-river-dispatch.md)):**
an event is in the M3 bundle iff **(a)** it corresponds to a **canonical business outcome** at an
existing module's transition point (publish / supersede / obsolete / final approval / final rejection),
and **(b)** its recipient set resolves from data the owning module **already has in scope at emit time**
— `controlled_document_id` (via the approval `CDFieldReader` port) → published
`metaldocs.v_cd_obligated_readers` for reader events; `instance.SubmittedBy` for author events. The
five bundle events meet both. **Approver-requested fails (a)+(b)** — no eligible-approver resolver
exists (grep empty) — so it is parked, not built. **Note the F3.3-recon correction:** the original
criterion assumed `governance_events` already carried the bundle outcomes, but it is an **actor-centric
audit log** (records *who acted*, not the *outcome* or *recipient*); `signoff.rejected` is an
eligibility-failure (not document rejection) and `signoff_recorded` has no terminal marker. So the two
prior author events are replaced by the clean terminal-transition events `document.approved` /
`document.rejected`. See ADR-0044 §3.

### Boundary decision — emitter is a typed domain event, not a projector over the audit log

The original M3 plan made the emitter a **notifications-owned projector over `governance_events`**.
F3.3 pre-spec recon (2026-06-22) found that to be a **workaround over a category error** (audit log ≠
domain-event substrate), which the operator rejected, directing a root-cause fix researched against
industry standards (ADR-0044). The emitter is now: at each of the five canonical business actions, the
**owning module enqueues a typed River-job domain event in the same `*sql.Tx`** as the state change
(River's same-tx enqueue = the transactional-outbox guarantee), carrying the outcome + recipient IDs;
a new **notifications fan-out River worker** consumes each job **after commit**, resolves recipients
(readers → published `v_cd_obligated_readers`; author → the event's `submitted_by`), and inserts
idempotent per-recipient rows (`source_event_id` = domain-event id; F3.2 unique index dedupes). This
respects ADR-0039 (the event **is** the cross-module contract; the worker reads only the published view,
no base-table reads) and H-PRE-1 (worker is off-tx). **HS-2 (publish path sacred) is explicitly lifted
by the operator (2026-06-22) for the additive in-tx enqueue only** — publish/approval *semantics* stay
unchanged (additive enqueue, covered by existing service tests + new emit assertions). Adding a parked
event later is additive (new typed job + worker subscription), never a rewrite.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F3.1 | `f3.1-notifications-contract` | ADR + OpenAPI for the notifications read surface, **consumer-contract-first** from `NotificationItem`: `GET /notifications?status=&cursor=&limit=` → `{items: Notification[], page: CursorPage}` (`Notification` shape == the FE `NotificationItem`: `id, recipient_user_id, event_type, resource_type, resource_id, title, message, status: "PENDING"\|"SENT"\|"READ", created_at, read_at?`); `GET /notifications/unread-count` → `{count}`; `POST /notifications/:id/read` → `204`/`{...}`. The `event_type` is an **open string** in the contract (not a closed enum) so the parked mission adds types additively without a breaking regen; the **five bundle types** (`document_published`, `document_superseded`, `document_obsoleted`, `signoff_recorded`, `signoff.rejected`) are documented as the M3-emitted set. New `notifications` module declared. New tier-2 cap `CapNotificationRead` (**self-scope** — a user reads/marks only their own notifications) registered + added to `deferredCaps`. ADR records: the new module + table shape (per-recipient, read-state) + cap + self-scope rule + the **document-lifecycle bundle decision** (five `event_type → recipient-resolver` rows, all via outbox projection; the selection criterion: existing outbox event ∧ existing recipient resolver) + the additive-extension commitment to the parked emitter mission (approver-routing etc.). Regen Go server types (`oapi-codegen`) + FE types (`npm run gen:api`). | `api-lint -strict` parses the new `notifications` paths = **0** violations; generated Go + FE types present; the generated `Notification` shape is structurally equal to FE `NotificationItem`; ADR present under `wiki/decisions/`; ADR-0039 inventory updated if a view is added; cap registered in `internal/modules/iam/domain/model.go` + scoped in `capability_scope.go` + entry in `scripts/api-lint/registry_rules.go` `deferredCaps`; `go build ./...` green; spec review. |
| F3.2 | `f3.2-notifications-backend` | Implement the **read surface** to the Grade-A bar: forward-only migration creating the notifications table owned by the `notifications` module (per-recipient rows, `status` read-state, `read_at`, tenant-scoped, RLS consistent with the 0237 pattern); repository + handlers in `internal/modules/notifications/` for list (cursor pagination per the existing `CursorPage` convention, keyset on `(created_at, id)`), unread-count, and mark-read. Reads/writes gated by `CapNotificationRead` **self-scoped** (a user can only list/mark their own rows — enforced in the query predicate, not just the cap). **No emitter in this feature** (F3.3 owns production). | Integration test against live PG: seeded notification rows for user A are listed for A and **not** for user B (self-scope holds); unread-count matches; mark-read flips `status`→`READ` + sets `read_at` and is idempotent; cursor pagination returns stable keyset pages. `api-lint -strict` = **0**; all **6 CI guards green** (notably `hgcrossmodule` — notifications owns its table; any cross-module read is a published view/port only); `go build`/`go vet`/`go test ./...` green; `git diff …/approval/application/publish_service.go` = empty. |
| F3.3 | `f3.3-lifecycle-emitter` | **(RE-SCOPED 2026-06-22 per [ADR-0044](../../../../../wiki/decisions/0044-domain-event-pattern-and-river-dispatch.md) — domain-event pattern, not a projector. HS-2 lifted by operator for the additive enqueue only.)** At each of the five canonical business actions, the **owning module emits a typed domain event** (a River job) **in the same `*sql.Tx`** as the state change, carrying outcome + recipient IDs: reader events (`document.published`/`.superseded`/`.obsoleted`) carry `controlled_document_id` (via the approval `CDFieldReader` port, `s.cdRead`); author events (`document.approved` @ the `InstanceApproved` transition, `document.rejected` @ `InstanceRejected`) carry `submitted_by` (`instance.SubmittedBy`). A new **notifications fan-out River worker** consumes each job, resolves recipients (readers → published `v_cd_obligated_readers`; author → the event's `submitted_by`), inserts one idempotent row per recipient (`source_event_id` = domain-event id; F3.2 unique index dedupes), with the real pt-BR title/message per the bundle table. Worker runs **after commit** (off-tx; H-PRE-1). Domain-event type defs live in the **owning module**; notifications **subscribes**. The five emit sites gain an **additive** in-tx enqueue — publish/approval **semantics unchanged**. | Integration test per event type: the triggering action yields exactly one notification per correct recipient (reader set == `v_cd_obligated_readers`; author == `submitted_by`) and **zero** for non-recipients; redelivering the job is a no-op (idempotent, no duplicate per `(recipient_user_id, source_event_id)`); `document.approved`/`.rejected` fire **only** at the terminal instance transition (not per-stage); each emit site asserts exactly one domain-event job enqueued in the state-change tx; `git diff` on the five services shows **additive enqueue only** — publish/approval *semantics* untouched (existing service-test assertions unchanged); `hgcrossmodule` green (worker reads only the published view; emit sites use existing in-module ports); `go build`/`vet`/`test ./...` green; all 6 CI guards green. |
| F3.4 | `f3.4-notifications-wire` | Replace the noop `notifications.ts` (`listNotifications`/`markNotificationRead`) + the empty `NotificationsPage` with real TanStack Query hooks against the F3.1 endpoints; wire `unread-count` to the reserved `QK.notifications.unreadCount` consumer (badge); restyle the page/panel to redesign tokens. Delete the `{ items: [] }`/`never[]` stub. The SSE operations-stream rebuild is **out of scope** (rabbit hole) — leave the existing `subscribeOperationsStream` noop untouched or behind its current behavior; do not fabricate a stream. | `grep -nE "items: \[\]\|never\[\]\|Stubbed pending" notifications/api/notifications.ts` = **0** for the list/mark-read functions (noop deleted); the page renders live notifications + unread-count from real queries; mark-read calls the real endpoint and the row flips to READ; consumer shape == generated types; a query-hook test passes against fixtured responses; **`frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE on record** (D2); FE tests green; `npm run build` clean. |

Order: **F3.1 → F3.2 → F3.3 → F3.4** — contract first (consumer-contract-first regen), then the
read-surface producer-agnostic backend, then the five real emitters (five typed domain events + one
fan-out worker, isolated so its existing-code-touch risk is gated alone), then the wire. F3.4 can only prove *live* data once F3.3 produces rows; until
then F3.2's read surface is exercised by integration fixtures. Each "what to validate" is objectively
checkable (a grep = 0, an `api-lint` count, a contracted shape, a passing integration/hook test, a
reviewer APPROVE-on-record), never "works".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What that gate
enforces for this milestone:

1. **Per-feature acceptance** — F3.1/F3.2/F3.3/F3.4 each meet their declared "what to validate", and
   each feature's **consumer contract** (`spec.md`) was honored: the generated `Notification` shape
   matches FE `NotificationItem`; notifications owns its table and reads cross-module data only via the
   published `v_cd_obligated_readers` view (no base-table reads); self-scope is enforced in the query,
   not just the cap; the emitter is five typed River-job domain events (one per canonical action) + one
   notifications fan-out worker per [ADR-0044](../../../../../wiki/decisions/0044-domain-event-pattern-and-river-dispatch.md);
   the five emit sites gain only an **additive in-tx enqueue** (publish/approval *semantics* unchanged);
   author events resolve the recipient from `instance.SubmittedBy` carried in the event, reader events
   from `controlled_document_id` → the published `v_cd_obligated_readers` view; the worker reads only
   the published view (no base-table reads) and runs off-tx.
2. **Workflow-class QA** — **FE:** `wiki/quality/screen-definition-of-done.md` (D2: both reviewer
   APPROVEs on record) + the runtime functional pass by reference to `wiki/quality/screen-qa-checklist.md`.
   **BE:** `wiki/quality/backend-api-qa-checklist.md` + the 6 CI guards + `api-lint -strict` = 0 + the
   F3.2 and F3.3 integration tests.
3. **Regression** — **M0 + M1 + M2** still pass their gates (single index route; 0 dead-stub routes;
   M1 dashboard greps hold; M2 Distribuição greps + the obligated-readers view intact); the publish
   path is edited by **additive enqueue only** — publish/approval **semantics unchanged** (existing
   `publish_service.go`/`decision_service.go` service-test assertions still pass; the only diff is the
   in-tx domain-event enqueue, per the ADR-0044 / HS-2 lift); `v_cd_grantee` and
   `v_cd_obligated_readers` untouched (`git diff` on their migrations = empty); the FE suite holds at
   the operator-accepted baseline; `go build ./...` / `go test ./...` green.
4. **Quality-bar re-measure (root cause, not symptom)** —
   `grep -nE "items: \[\]|never\[\]|Stubbed pending" notifications/api/notifications.ts` = **0** (noop
   deleted at root, not flag-hidden); the inbox renders **live** rows at runtime from real producers
   (proven by the F3.3 per-event integration tests: publish/supersede/obsolete → obligated-reader
   notification, final-signoff/reject → submitter notification); mark-read flips state and persists;
   the **parked** triggers (approver-routing, templates, channels, prefs) render as honest absence,
   **not** fabricated rows.
5. **No unplanned scope** — only F3.1 + F3.2 + F3.3 + F3.4 are implemented; the emitter is **exactly
   the five document-lifecycle domain events** (no approver-routing emitter, no template-lifecycle
   emitter), **no** email/push channel, **no** notification preferences, **no** SSE-stream rebuild,
   **no** publish-path *semantic* change (additive enqueue only, per the ADR-0044 / HS-2 lift), **no**
   modification of `v_cd_grantee`/`v_cd_obligated_readers` or any search/approval-owned business logic.
   Any pull toward those is routed to the parked emitter mission.

## Dependencies & constraints

- **Depends on:** M0 + M1 + M2 passed; the M2-published `metaldocs.v_cd_obligated_readers` view (the
  reader-targeted recipient set); River (Postgres-backed queue, already in production for scheduled
  publish) as the domain-event dispatch mechanism; the five emit sites that will gain the additive
  in-tx enqueue (`approval/application/publish_service.go`, `supersede_service.go`,
  `obsolete_service.go`, `decision_service.go` at the `InstanceApproved`/`InstanceRejected`
  transitions); the approval `CDFieldReader` port (`s.cdRead`) for `controlled_document_id` and
  `instance.SubmittedBy` for the author recipient (both already in scope); the FE `NotificationItem`
  contract + reserved `QK.notifications.unreadCount` query key; the existing `NotificationsPanel`
  component as the wire target.
- **Appetite:** medium-plus — one full-stack slice: a **greenfield `notifications` module** (table +
  read endpoints + cap), **five typed domain events** (one per canonical document-lifecycle action) +
  **one** notifications fan-out River worker (per
  [ADR-0044](../../../../../wiki/decisions/0044-domain-event-pattern-and-river-dispatch.md)), and the
  screen wire. Bounded well below the parked emitter mission (lifecycle bundle only — no
  approver-routing, no channels, no preferences, no stream rebuild). One new mutable table
  (notifications rows); no new table for the events (River jobs are the outbox); the five emit sites
  gain only an additive in-tx enqueue (publish/approval *semantics* unchanged).
- **Quality goals (ranked):** 1) **truthfulness** (render only notifications a real producer emits;
  parked triggers are honestly absent, never faked) > 2) **module-boundary cleanliness** (notifications
  owns its table; cross-module data only via the published `v_cd_obligated_readers` view + the
  domain-event payload; the event **is** the cross-module contract; ADR-0039 honored) > 3)
  **contract-correctness** (generated `Notification` == FE `NotificationItem`; the contract is
  **forward-compatible** — the parked mission adds event types additively, never breaks it) > 4)
  **simplicity** (smallest backend that serves real data: one table, three read endpoints, five typed
  events + one idempotent fan-out worker, no new event table).
- **Architectural constraints (validator can fail on these):**
  - **Self-scope authz:** `CapNotificationRead` is **self-scope** — list/unread/mark-read return and
    mutate **only the caller's own** rows; enforced in the SQL predicate (`recipient_user_id = caller`),
    not merely by holding the cap. Cap registered in `iam/domain/model.go` + scoped in
    `capability_scope.go` + added to `scripts/api-lint/registry_rules.go` `deferredCaps` (operator
    grants to roles separately — never pre-granted by the agent). Route guarded tier-1 in
    `apps/api/cmd/metaldocs-api/permissions.go` per the tenant-grade-GET precedent (`CapAuditRead`).
  - **Publish/approval semantics are sacred — enqueue is additive (HS-2 lifted for the enqueue only,
    2026-06-22, per [ADR-0044](../../../../../wiki/decisions/0044-domain-event-pattern-and-river-dispatch.md)):**
    the five emit sites (`publish_service.go`, `supersede_service.go`, `obsolete_service.go`,
    `decision_service.go` ×2) gain **only** an in-tx typed-domain-event River enqueue next to the
    existing audit emit. **Zero** change to existing publish/approval *behavior* — existing service-test
    assertions must still pass; the only diff is the additive enqueue. A change that alters publish/approval
    *semantics* (not just adds an enqueue) trips **HS-2**.
  - **No raw cross-module base-table reads:** the notifications fan-out worker reads cross-module data
    only via `metaldocs.v_cd_obligated_readers` (reader events) + the recipient IDs carried in the
    domain-event payload (author events) + any iam display-name port already in use (`hgcrossmodule` =
    0). It does **not** read CD/taxonomy/iam/approval base tables raw. The domain event **is** the
    cross-module contract (ADR-0044 §5).
  - **Author-targeted recipient resolution (ADR-0044):** author events (`document.approved` @
    `InstanceApproved`, `document.rejected` @ `InstanceRejected`) carry `submitted_by` directly in the
    event payload — the approval module has `instance.SubmittedBy` in scope at emit time, so the worker
    never reads the approval instance base table. Reader events carry `controlled_document_id` (resolved
    at emit via the approval `CDFieldReader` port) → the worker joins the published
    `v_cd_obligated_readers`. **This eliminates both ADR-0043 §6 compensating views.**
  - **`v_cd_grantee` / `v_cd_obligated_readers` are sacred:** untouched (their migrations diff empty).
  - **Idempotent consumer:** the fan-out worker inserts at most one row per `(recipient_user_id,
    source_event_id=domain-event id)`; at-least-once redelivery is a no-op (F3.2 partial unique index).
    No duplicate inbox spam.
  - **H-PRE-1:** no authz-recording read inside a lock-holding atomic tx; the fan-out worker runs
    **after commit**, off any publish-path lock.
  - **Contract-first regen order:** spec → OpenAPI → `oapi-codegen` → FE types; the FE consumes the
    **generated** types only, never hand-rolled shapes (HS-3).
  - **Grade-A backend bar:** new endpoints pass `api-lint -strict` = 0, all 6 CI guards, and
    integration tests against live PG; `go build`/`vet`/`test` green.
  - **Design system is consumed, not redesigned:** use `tokens.css` + existing primitives /
    `NotificationsPanel`; changing a shared primitive trips HS-2.
  - **Migration policy:** forward-only (`wiki/database/migration-policy.md`); RLS consistent with the
    0237 tenant-table pattern for the new table.
  - Reads stay live (no caching workaround); **no merge / no push by the agent** (commits allowed after
    verified work).
- **Risks:**
  - **R1 — emit-site edit (additive enqueue vs semantic change).** The five emit sites are now edited
    (HS-2 lifted), but only to add an in-tx enqueue. *Mitigation:* the enqueue is additive next to the
    existing audit emit; existing service-test assertions must still pass; any change to publish/approval
    *semantics* trips HS-2 (stop + replan). Each emit site asserts exactly one domain-event job enqueued.
  - **R2 — idempotency / duplicate inbox spam.** At-least-once River redelivery could double-insert.
    *Mitigation:* F3.2 partial unique `(recipient_user_id, source_event_id=domain-event id)`; F3.3
    integration test asserts redelivery = no-op.
  - **R3 — self-scope leak.** A cap-only check (no row predicate) would let any holder read others'
    inboxes. *Mitigation:* F3.2 integration test asserts user A never sees user B's rows; predicate
    enforced in SQL.
  - **R4 — company-scope publish fan-out volume.** Publishing a company-scope CD notifies all active
    tenant users (the M2 company-scope leg). *Mitigation:* the fan-out worker runs after commit (off-tx)
    and batches inserts; if a real measurement shows unacceptable volume/latency, bound it — but only if
    proven (no premature optimization).
  - **R5 — "honest absence" misread as broken.** With the lifecycle bundle, the inbox is empty until a
    lifecycle event (publish/supersede/obsolete/approve/reject) happens. *Mitigation:* the F3.4
    empty-state copy already reads as a truthful empty inbox ("Quando houver eventos operacionais…");
    the F3.3 per-event integration tests are the live-data proof.
  - **R6 — author-recipient cross-module read (eliminated by ADR-0044).** Author events carry
    `submitted_by` directly in the event payload (the approval module has `instance.SubmittedBy` in
    scope at emit time), so the fan-out worker never reads the approval instance base table —
    `hgcrossmodule` clean. Both ADR-0043 §6 compensating views are eliminated; no submitter-resolution
    risk remains.
- **Rabbit holes (do NOT chase):**
  - *A second-class emitter beyond the lifecycle bundle (approver-routing/next-approver, template
    lifecycle, route.config, reminder/SLA)* — parked emitter mission. (The five lifecycle events —
    published/superseded/obsoleted/signoff-recorded/signoff-rejected — ARE in M3 scope.)
  - *Email / push / any out-of-app delivery channel* — parked mission.
  - *Per-user notification preferences / mute / digest* — parked mission.
  - *Rebuilding the SSE operations stream* (`subscribeOperationsStream`) — out of scope; leave as-is,
    do not fabricate a stream.
  - *Changing publish/approval **semantics** at an emit site* — HS-2. (Adding an in-tx domain-event
    enqueue next to the existing audit emit is explicitly allowed per ADR-0044 / the HS-2 lift; altering
    existing behavior is not.)
  - *Inventing notification rows with no producer (demo seed in prod path)* — fabrication; violates
    quality goal 1.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | Milestone close: operator review gate before M4 and before any merge — mandatory. **Start gate CLEARED 2026-06-22:** operator approved the start and **widened** the emitter scope to the document-lifecycle bundle; **re-scoped 2026-06-22 (HS-2/HS-6, F3.3 recon)** from outbox-projector to the ADR-0044 typed-domain-event pattern (5 domain events + 1 fan-out worker); approver-routing + templates + channels + prefs parked. |
| HS-2 | **Re-scoped 2026-06-22 (ADR-0044): partially LIFTED for the additive enqueue.** The emit sites (`publish_service.go`, `supersede_service.go`, `obsolete_service.go`, `decision_service.go`) **may** gain an **additive in-tx domain-event River enqueue** — explicitly authorized by the operator. HS-2 still trips on: changing publish/approval **semantics** (behavior diff, not just an added enqueue), modifying `v_cd_grantee`/`v_cd_obligated_readers`, changing the *behavior* of other search-/approval-owned logic/existing views/tables, or changing a shared primitive/token. **Stop** and report on any of those. |
| HS-3 | A prerequisite fails at runtime: app won't start, no auth session, the notifications route is broken, or contract↔generated types drift. Repair the prerequisite (contract-first regen order), rerun the checkpoint, then resume the feature. |
| HS-4 | `milestone-validator` returns FAIL → open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Scope drift: an emitter **beyond the five-event lifecycle bundle** (approver-routing, template lifecycle, route.config, reminder/SLA), a delivery channel, preferences, a stream rebuild, or any non-notification concern appears in M3 → stop and replan; route to the parked emitter mission. |
