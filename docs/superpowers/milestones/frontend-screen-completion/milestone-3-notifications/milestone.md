# Milestone 3 — Notifications (full-stack; surface + document-lifecycle emitters)

> **Program:** frontend-screen-completion  ·  **Governing spec:** `../mission.md`
> **Status:** In progress — operator HS-1 start-gate **approved 2026-06-22**; emitter scope **widened**
> at the gate to the **document-lifecycle bundle** (see below). Executing F3.1.
> *Subagent model: Sonnet 4.6 (operator directive 2026-06-21, carried from M2).*
> **Authored:** 2026-06-22 — *before any feature in this milestone began.* Scope fork (producer
> question) resolved by operator **2026-06-22: option B — surface + real emitters**, then **widened at
> the HS-1 start gate (2026-06-22) from one emitter to the document-lifecycle bundle** (5 cheap
> outbox-projector emitters whose recipient sets already resolve; approver-routing + templates +
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

There are **five real producers** — the **document-lifecycle bundle**, all implemented as one
idempotent **notifications-owned projector** over the existing `governance_events` outbox, each event
type mapped to an already-resolvable recipient set (no new recipient-resolution logic, no publish-path
edit):

| Outbox event (already emitted) | Recipient set | Resolver (already exists) | Inbox message (pt-BR) |
|--------------------------------|---------------|---------------------------|------------------------|
| `document_published` | obligated readers | `metaldocs.v_cd_obligated_readers` (M2) | "Novo documento controlado para leitura" |
| `document_superseded` | obligated readers | `metaldocs.v_cd_obligated_readers` | "Documento substituído por nova revisão" |
| `document_obsoleted` | obligated readers | `metaldocs.v_cd_obligated_readers` | "Documento que você lê foi obsoletado" |
| `signoff_recorded` (final approval) | submitter | approval instance `submitted_by` | "Seu documento foi aprovado" |
| `signoff.rejected` | submitter | approval instance `submitted_by` | "Documento rejeitado — ajustes solicitados" |

So the screen renders genuinely-live rows across the real document lifecycle the first time any of these
events fires after M3, not a fabricated or seeded demo. The five share one mechanism — the projector is
parameterized by an `event_type → recipient-query` map; the marginal cost of each additional lifecycle
event after the first is a map row + a recipient query + an integration assertion.

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
  document-lifecycle producers above (one projector, five event→recipient map rows) so the screen
  demonstrably renders live data across the real lifecycle.
- The **approver-routing** trigger and all other triggers are **parked** to a designed
  `notification-emitter` mission, rendered as their honest absence (no event of that type yet), never
  fabricated.

**Selection criterion (grounded in runtime truth — operator-ratified at the HS-1 start gate):** an
event is in the M3 bundle iff **(a)** it is already persisted to the `governance_events` outbox by an
existing module (so M3 can **project** without touching any emit site), **and (b)** its recipient set
**already resolves** from an existing view/column — `metaldocs.v_cd_obligated_readers` (M2) for
reader-targeted events, the approval instance `submitted_by` column for author-targeted events. The
five bundle events meet both; **approver-requested fails (b)** — no eligible-approver resolver exists
(grep empty) — so it is parked, not built. Every projector reads the outbox `+` a recipient resolver
and writes only notifications-owned rows: **no edit to `publish_service.go` or any existing emit
code**, blast radius held to one new module.

### Boundary decision — emitter is a projector, not a publish-path edit

To respect ADR-0039 (no cross-module raw base-table writes) **and** the program's standing rule that
the publish path is sacred (M2 guarded `publish_service.go` = empty diff), the emitter is a single
**notifications-owned projector over the already-emitted `governance_events` outbox**: for each of the
five bundle event types, the notifications module reads `governance_events` ⋈ the event's recipient
resolver (`metaldocs.v_cd_obligated_readers` for reader-targeted events; the approval instance
`submitted_by` for author-targeted events) and inserts per-recipient notification rows it owns. The
event→recipient mapping is a single table in the projector — adding a parked event later is additive
(a map row), never a rewrite. **No existing module's code is edited to emit.** The exact projection
trigger (a `jobs` scheduler tick consuming the outbox vs. an idempotent catch-up projection on read)
is an **F3.3 plan decision**, constrained here to: idempotent (no duplicate rows per
`(recipient, source_event)`), publish-path untouched, and respecting H-PRE-1 (no authz-recording read
inside a lock-holding atomic tx). If F3.3 discovers the only correct mechanism requires editing the
publish transaction, that trips **HS-2** — stop and replan.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F3.1 | `f3.1-notifications-contract` | ADR + OpenAPI for the notifications read surface, **consumer-contract-first** from `NotificationItem`: `GET /notifications?status=&cursor=&limit=` → `{items: Notification[], page: CursorPage}` (`Notification` shape == the FE `NotificationItem`: `id, recipient_user_id, event_type, resource_type, resource_id, title, message, status: "PENDING"\|"SENT"\|"READ", created_at, read_at?`); `GET /notifications/unread-count` → `{count}`; `POST /notifications/:id/read` → `204`/`{...}`. The `event_type` is an **open string** in the contract (not a closed enum) so the parked mission adds types additively without a breaking regen; the **five bundle types** (`document_published`, `document_superseded`, `document_obsoleted`, `signoff_recorded`, `signoff.rejected`) are documented as the M3-emitted set. New `notifications` module declared. New tier-2 cap `CapNotificationRead` (**self-scope** — a user reads/marks only their own notifications) registered + added to `deferredCaps`. ADR records: the new module + table shape (per-recipient, read-state) + cap + self-scope rule + the **document-lifecycle bundle decision** (five `event_type → recipient-resolver` rows, all via outbox projection; the selection criterion: existing outbox event ∧ existing recipient resolver) + the additive-extension commitment to the parked emitter mission (approver-routing etc.). Regen Go server types (`oapi-codegen`) + FE types (`npm run gen:api`). | `api-lint -strict` parses the new `notifications` paths = **0** violations; generated Go + FE types present; the generated `Notification` shape is structurally equal to FE `NotificationItem`; ADR present under `wiki/decisions/`; ADR-0039 inventory updated if a view is added; cap registered in `internal/modules/iam/domain/model.go` + scoped in `capability_scope.go` + entry in `scripts/api-lint/registry_rules.go` `deferredCaps`; `go build ./...` green; spec review. |
| F3.2 | `f3.2-notifications-backend` | Implement the **read surface** to the Grade-A bar: forward-only migration creating the notifications table owned by the `notifications` module (per-recipient rows, `status` read-state, `read_at`, tenant-scoped, RLS consistent with the 0237 pattern); repository + handlers in `internal/modules/notifications/` for list (cursor pagination per the existing `CursorPage` convention, keyset on `(created_at, id)`), unread-count, and mark-read. Reads/writes gated by `CapNotificationRead` **self-scoped** (a user can only list/mark their own rows — enforced in the query predicate, not just the cap). **No emitter in this feature** (F3.3 owns production). | Integration test against live PG: seeded notification rows for user A are listed for A and **not** for user B (self-scope holds); unread-count matches; mark-read flips `status`→`READ` + sets `read_at` and is idempotent; cursor pagination returns stable keyset pages. `api-lint -strict` = **0**; all **6 CI guards green** (notably `hgcrossmodule` — notifications owns its table; any cross-module read is a published view/port only); `go build`/`go vet`/`go test ./...` green; `git diff …/approval/application/publish_service.go` = empty. |
| F3.3 | `f3.3-lifecycle-emitter` | Wire the **five** real producers as a **single notifications-owned projector over the `governance_events` outbox**, parameterized by an `event_type → recipient-resolver` map: **reader-targeted** (`document_published`, `document_superseded`, `document_obsoleted`) → one row per obligated reader from `metaldocs.v_cd_obligated_readers`; **author-targeted** (`signoff_recorded`, `signoff.rejected`) → one row for the approval instance `submitted_by`. Each inserted idempotently (no duplicate per `(recipient_user_id, source_event_id)`), with the event's `resource_type`/`resource_id` and a real pt-BR title/message per the bundle table. **No edit to `publish_service.go` or any existing module's emit code.** Exact trigger (scheduler tick vs. catch-up projection) per this feature's `plan.md`, within the milestone constraints. | Integration test per event type: the triggering action (publish / supersede / obsolete / final-signoff / signoff-reject) yields exactly one notification per correct recipient (reader set == `v_cd_obligated_readers` for the CD; author == `submitted_by`) and **zero** for non-recipients; re-running the projection is a no-op (idempotent, no duplicates); the `event_type → recipient` map has a test row per bundle event; `git diff internal/modules/documents/approval/application/publish_service.go` = **empty**; `hgcrossmodule` green (projector reads only the published view + the outbox table per its declared ownership); `go build`/`vet`/`test ./...` green; all 6 CI guards green. |
| F3.4 | `f3.4-notifications-wire` | Replace the noop `notifications.ts` (`listNotifications`/`markNotificationRead`) + the empty `NotificationsPage` with real TanStack Query hooks against the F3.1 endpoints; wire `unread-count` to the reserved `QK.notifications.unreadCount` consumer (badge); restyle the page/panel to redesign tokens. Delete the `{ items: [] }`/`never[]` stub. The SSE operations-stream rebuild is **out of scope** (rabbit hole) — leave the existing `subscribeOperationsStream` noop untouched or behind its current behavior; do not fabricate a stream. | `grep -nE "items: \[\]\|never\[\]\|Stubbed pending" notifications/api/notifications.ts` = **0** for the list/mark-read functions (noop deleted); the page renders live notifications + unread-count from real queries; mark-read calls the real endpoint and the row flips to READ; consumer shape == generated types; a query-hook test passes against fixtured responses; **`frontend-screen-reviewer` + `frontend-code-reviewer` APPROVE on record** (D2); FE tests green; `npm run build` clean. |

Order: **F3.1 → F3.2 → F3.3 → F3.4** — contract first (consumer-contract-first regen), then the
read-surface producer-agnostic backend, then the five real emitters (one projector, isolated so its
existing-code-touch risk is gated alone), then the wire. F3.4 can only prove *live* data once F3.3 produces rows; until
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
   not just the cap; the emitter is a single projector covering all **five** bundle event types
   (`event_type → recipient` map) that does **not** edit the publish path; author-targeted events
   resolve the recipient from the approval instance `submitted_by`, reader-targeted from the published
   view.
2. **Workflow-class QA** — **FE:** `wiki/quality/screen-definition-of-done.md` (D2: both reviewer
   APPROVEs on record) + the runtime functional pass by reference to `wiki/quality/screen-qa-checklist.md`.
   **BE:** `wiki/quality/backend-api-qa-checklist.md` + the 6 CI guards + `api-lint -strict` = 0 + the
   F3.2 and F3.3 integration tests.
3. **Regression** — **M0 + M1 + M2** still pass their gates (single index route; 0 dead-stub routes;
   M1 dashboard greps hold; M2 Distribuição greps + the obligated-readers view intact); the publish
   path is untouched (`git diff …/approval/application/publish_service.go` = empty); `v_cd_grantee`
   and `v_cd_obligated_readers` untouched (`git diff` on their migrations = empty); the FE suite holds
   at the operator-accepted baseline; `go build ./...` / `go test ./...` green.
4. **Quality-bar re-measure (root cause, not symptom)** —
   `grep -nE "items: \[\]|never\[\]|Stubbed pending" notifications/api/notifications.ts` = **0** (noop
   deleted at root, not flag-hidden); the inbox renders **live** rows at runtime from real producers
   (proven by the F3.3 per-event integration tests: publish/supersede/obsolete → obligated-reader
   notification, final-signoff/reject → submitter notification); mark-read flips state and persists;
   the **parked** triggers (approver-routing, templates, channels, prefs) render as honest absence,
   **not** fabricated rows.
5. **No unplanned scope** — only F3.1 + F3.2 + F3.3 + F3.4 are implemented; the emitter is **exactly
   the five document-lifecycle bundle events** (no approver-routing emitter, no template-lifecycle
   emitter), **no** email/push channel, **no** notification preferences, **no** SSE-stream rebuild,
   **no** publish-path edit, **no** modification of `v_cd_grantee`/`v_cd_obligated_readers` or any
   search/approval-owned code. Any pull toward those is routed to the parked emitter mission.

## Dependencies & constraints

- **Depends on:** M0 + M1 + M2 passed; the M2-published `metaldocs.v_cd_obligated_readers` view (the
  reader-targeted recipient set); the existing `governance_events` outbox + the five bundle events
  already emitted by existing modules (`document_published`, `publish_scheduled`-adjacent
  `document_superseded`/`document_obsoleted`, `signoff_recorded`, `signoff.rejected` —
  `approval/application/events.go`, `supersede_service.go`, `obsolete_service.go`,
  `decision_service.go`); the approval instance `submitted_by` column (author-targeted recipient); the
  FE `NotificationItem` contract + reserved `QK.notifications.unreadCount` query key; the existing
  `NotificationsPanel` component as the wire target.
- **Appetite:** medium-plus — one full-stack slice: a **greenfield `notifications` module** (table +
  read endpoints + cap), **one** outbox projector covering the **five document-lifecycle events** (an
  `event_type → recipient-resolver` map, not five separate projectors), and the screen wire. Bounded
  well below the parked emitter mission (lifecycle bundle only — no approver-routing, no channels, no
  preferences, no stream rebuild). One new mutable table (notifications rows); no change to any existing
  table or to the publish path.
- **Quality goals (ranked):** 1) **truthfulness** (render only notifications a real producer emits;
  parked triggers are honestly absent, never faked) > 2) **module-boundary cleanliness** (notifications
  owns its table; cross-module data only via the published `v_cd_obligated_readers` view + the outbox;
  the emitter is a projector, never a publish-path edit; ADR-0039 honored) > 3) **contract-correctness**
  (generated `Notification` == FE `NotificationItem`; the contract is **forward-compatible** — the
  parked mission adds event types additively, never breaks it) > 4) **simplicity** (smallest backend
  that serves real data: one table, three read endpoints, one idempotent projector).
- **Architectural constraints (validator can fail on these):**
  - **Self-scope authz:** `CapNotificationRead` is **self-scope** — list/unread/mark-read return and
    mutate **only the caller's own** rows; enforced in the SQL predicate (`recipient_user_id = caller`),
    not merely by holding the cap. Cap registered in `iam/domain/model.go` + scoped in
    `capability_scope.go` + added to `scripts/api-lint/registry_rules.go` `deferredCaps` (operator
    grants to roles separately — never pre-granted by the agent). Route guarded tier-1 in
    `apps/api/cmd/metaldocs-api/permissions.go` per the tenant-grade-GET precedent (`CapAuditRead`).
  - **Publish path is sacred:** **zero** change to `publish_service.go` or any existing module's emit
    code. The emitter is a **notifications-owned projector over `governance_events`**. A fix that
    requires editing the publish transaction trips **HS-2**.
  - **No raw cross-module base-table reads:** notifications reads cross-module data only via
    `metaldocs.v_cd_obligated_readers` + the `governance_events` outbox (per declared ownership /
    ADR-0039 disposition) + any iam display-name port already in use (`hgcrossmodule` = 0). It does
    **not** read CD/taxonomy/iam/approval base tables raw.
  - **Author-targeted recipient resolution (F3.1 contract decision):** the author-targeted events
    (`signoff_recorded`, `signoff.rejected`) need the document **submitter** as the recipient, but
    `submitted_by` lives on the approval instance (an approval-module base table) and the outbox event's
    `actor_user_id` is the *approver*, not the submitter. The projector must therefore resolve the
    submitter **without a raw approval base-table read** — F3.1 chooses one of: (a) the
    `governance_events` payload already carries `submitted_by`/document author (preferred — verify at
    F3.1 recon), or (b) the approval/documents module publishes a minimal owner-published view
    (`metaldocs.v_*`, the M2 F2.1a/b pattern — additive, not a publish-path edit). Resolving it any
    other way (raw read of the approval instance) **fails `hgcrossmodule`** and is forbidden. If only
    option (b) works and it would require a non-additive change to an existing module, that is an F3.1
    surface, not an HS-2 (publishing a new view is additive).
  - **`v_cd_grantee` / `v_cd_obligated_readers` are sacred:** untouched (their migrations diff empty).
  - **Idempotent producer:** the projector inserts at most one row per `(recipient_user_id,
    source_event_id)`; re-running it is a no-op (unique constraint or upsert). No duplicate inbox spam.
  - **H-PRE-1:** no authz-recording read inside a lock-holding atomic tx; the projector runs off any
    publish-path lock.
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
  - **R1 — emitter boundary (projector vs publish-path edit).** The clean recipient set is M2's view,
    but the *mechanism* must not touch the sacred publish path. *Mitigation:* locked as an outbox
    projector reading `governance_events`; F3.3's plan picks the trigger within constraints; a forced
    publish-path edit trips HS-2 (stop + replan).
  - **R2 — idempotency / duplicate inbox spam.** A projector re-run or an at-least-once outbox could
    double-insert. *Mitigation:* unique `(recipient_user_id, source_event_id)`; F3.3 integration test
    asserts re-run = no-op.
  - **R3 — self-scope leak.** A cap-only check (no row predicate) would let any holder read others'
    inboxes. *Mitigation:* F3.2 integration test asserts user A never sees user B's rows; predicate
    enforced in SQL.
  - **R4 — company-scope publish fan-out volume.** Publishing a company-scope CD notifies all active
    tenant users (the M2 company-scope leg). *Mitigation:* projector is off-tx and batched; if a real
    measurement shows unacceptable volume/latency, bound it — but only if proven (no premature
    optimization).
  - **R5 — "honest absence" misread as broken.** With the lifecycle bundle, the inbox is empty until a
    lifecycle event (publish/supersede/obsolete/signoff) happens. *Mitigation:* the F3.4 empty-state
    copy already reads as a truthful empty inbox ("Quando houver eventos operacionais…"); the F3.3
    per-event integration tests are the live-data proof.
  - **R6 — author-recipient cross-module read.** Author-targeted events need the approval-instance
    `submitted_by`; a raw read of the approval base table fails `hgcrossmodule`. *Mitigation:* F3.1
    resolves the submitter via the outbox payload or an additive owner-published view (see the
    author-targeted constraint above); if neither is available without a non-additive approval-module
    change, narrow M3 back to the reader-targeted three and re-park the author-targeted two (operator
    decision at that point).
- **Rabbit holes (do NOT chase):**
  - *A second-class emitter beyond the lifecycle bundle (approver-routing/next-approver, template
    lifecycle, route.config, reminder/SLA)* — parked emitter mission. (The five lifecycle events —
    published/superseded/obsoleted/signoff-recorded/signoff-rejected — ARE in M3 scope.)
  - *Email / push / any out-of-app delivery channel* — parked mission.
  - *Per-user notification preferences / mute / digest* — parked mission.
  - *Rebuilding the SSE operations stream* (`subscribeOperationsStream`) — out of scope; leave as-is,
    do not fabricate a stream.
  - *Editing `publish_service.go` or any existing emit site to push notifications* — HS-2; use the
    outbox projector.
  - *Inventing notification rows with no producer (demo seed in prod path)* — fabrication; violates
    quality goal 1.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | Milestone close: operator review gate before M4 and before any merge — mandatory. **Start gate CLEARED 2026-06-22:** operator approved the start and **widened** the emitter scope to the document-lifecycle bundle (5 outbox-projector emitters); approver-routing + templates + channels + prefs parked. |
| HS-2 | A "fix" turns out to require editing the publish path / `publish_service.go`, modifying `v_cd_grantee`/`v_cd_obligated_readers`, changing the *behavior* of search- or approval-owned code (logic, emit sites, existing views/tables), or changing a shared primitive/token. **Stop**; report the boundary; route to the parked mission — do not symptom-patch the publish path. **Carve-out (not HS-2):** an *additive* owner-published view from the approval/documents module to expose the submitter for author-targeted recipient resolution (the M2 F2.1a/b precedent) — additive new `metaldocs.v_*`, zero change to existing behavior — is an F3.1 surface, not a boundary breach. |
| HS-3 | A prerequisite fails at runtime: app won't start, no auth session, the notifications route is broken, or contract↔generated types drift. Repair the prerequisite (contract-first regen order), rerun the checkpoint, then resume the feature. |
| HS-4 | `milestone-validator` returns FAIL → open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | Scope drift: an emitter **beyond the five-event lifecycle bundle** (approver-routing, template lifecycle, route.config, reminder/SLA), a delivery channel, preferences, a stream rebuild, or any non-notification concern appears in M3 → stop and replan; route to the parked emitter mission. |
