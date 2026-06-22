# Feature F3.1 — Spec (notifications-contract)

> **Milestone:** 3 — Notifications (full-stack; surface + document-lifecycle emitters)  ·  **Folder:** `f3.1-notifications-contract`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-22 / leandrotca (operator HS-1 start gate + lifecycle-bundle scope) — *implementation may begin.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (`plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Consumer-contract discovery was driven by (a) the frozen FE consumer type `NotificationItem`, (b) the
operator HS-1 start-gate decisions, and (c) a read-only codebase recon (outbox schema + the five emit
sites + reusable module/codegen/cap patterns). Q&A persisted below.

| # | Question | Answer |
|---|----------|--------|
| 1 | What is the canonical consumer contract for a notification row? | The frozen FE type `NotificationItem` (`frontend/apps/web/src/lib/types/index.ts:178`): `id, recipientUserId, eventType, resourceType, resourceId, title, message, status: "PENDING"\|"SENT"\|"READ", createdAt, readAt?`. The server `Notification` schema is the snake_case mirror of exactly this set (HS-3: FE will consume the generated type in F3.4). |
| 2 | Closed enum or open string for `event_type`? | **Open string.** The parked emitter mission adds event types additively; a closed enum would force a breaking regen each time. M3 documents the five emitted types but does not constrain the field. (`status` stays a closed enum — it is a fixed read-state machine.) |
| 3 | Cap + scope? | New tier-2 cap `CapNotificationRead`, **self-scope** (a user lists/marks only their own rows; enforced in the SQL predicate, not just by holding the cap). Registered + added to `deferredCaps` (operator grants to roles separately). Mirrors `CapDistributionRead` (M2) across 4 files. |
| 4 | Which producers does M3 emit? (operator HS-1 decision) | The **document-lifecycle bundle** (5 events): reader-targeted `document_published`/`document_superseded`/`document_obsoleted` → obligated readers; author-targeted `signoff_recorded`/`signoff.rejected` → submitter. Approver-routing + templates + channels + prefs parked. |
| 5 | Do the outbox events carry the recipient/author? (recon — decisive) | **No.** `governance_events` carries `actor_user_id` (the approver/actor), never the submitter; the five emit-site payloads carry only fact identifiers (`instance_id`, document ids, decision, reason). The submitter lives on `approval_instances.submitted_by`. Reader events carry `resource_id = document_id`, while `v_cd_obligated_readers` keys on the **CD id** — a document→CD mapping is also required. |
| 6 | How is the author/reader recipient resolved without breaking a boundary? | **Additive owner-published views**, never an emit-site edit and never a raw cross-module base-table read. Approval/documents publishes a minimal `metaldocs.v_*` exposing `(tenant_id, approval_instance_id, submitted_by, document_id)` for author-targeted events, and a document→CD mapping (or `v_cd_obligated_readers` is keyed/extended to accept document id) for reader-targeted. The "just add `submitted_by` to the signoff payload" shortcut is **rejected** — editing the approval emit sites trips **HS-2**. The additive view is the locked M2 F2.1a/b precedent (carved out as not-HS-2). *DDL is authored in F3.2/F3.3; F3.1 records the decision in the ADR.* |
| 7 | Endpoint shapes? | `GET /notifications?status=&cursor=&limit=` → `{items: Notification[], page: CursorPage}`; `GET /notifications/unread-count` → `{count}`; `POST /notifications/{id}/read` → `204`. Cursor pagination per the existing `CursorPage` (`{has_more, next_cursor}`) convention; keyset on `(created_at, id)`. |

## Consumer contract (FIRST — before any producer)

- **Consumers:**
  - **FE notifications surface (F3.4):** `frontend/apps/web/src/features/notifications/api/notifications.ts` (today a noop returning `{items: []}`), `NotificationsPage`, `NotificationsPanel`, and the reserved `QK.notifications.unreadCount` badge consumer. It will consume the **generated** types only (HS-3).
  - **F3.2 backend** implements the producer side of this contract; **F3.3** projects rows into it.
- **Contract (the server `Notification` schema — snake_case mirror of FE `NotificationItem`):**
  ```
  Notification:
    id: string (uuid)
    recipient_user_id: string
    event_type: string            # open string; M3 emits the 5 bundle types
    resource_type: string         # e.g. "document" | "approval_instance"
    resource_id: string
    title: string
    message: string
    status: "PENDING" | "SENT" | "READ"   # closed enum (read-state)
    created_at: string (date-time)
    read_at: string (date-time)   # nullable/optional
  ```
  - `GET /notifications?status=&cursor=&limit=` → `{ items: Notification[], page: { has_more: bool, next_cursor: string|null } }`
  - `GET /notifications/unread-count` → `{ count: integer }`
  - `POST /notifications/{id}/read` → `204 No Content` (idempotent)
  - Error envelope: the project-standard problem shape (same as distribution; `api-lint -strict` enforces).
- **Source of truth for the contract:** FE `NotificationItem` (`lib/types/index.ts:178`) → encoded as the `Notification` schema in `api/openapi/v1/openapi.yaml` → generated to Go (`oapi-codegen`) + FE (`openapi-typescript`). The generated `Notification` MUST be structurally equal to `NotificationItem`.

## What this feature implements

1. **OpenAPI** (`api/openapi/v1/openapi.yaml`): a `notifications` tag + the three paths above + the
   `Notification` schema, authored consumer-contract-first from `NotificationItem`. `api-lint -strict`
   parses them at **0** violations.
2. **`notifications` module declared:** `internal/modules/notifications/api/{cfg.yaml,gen.go,api.gen.go}`
   (package `notificationsapi`, `include-tags: [notifications]`, strict-server), generated via
   `oapi-codegen` — mirroring the distribution module's api/ shape. (Handlers/repo/migration are F3.2;
   the projector is F3.3. F3.1 stands up only the generated contract surface — it compiles standalone.)
3. **New cap `CapNotificationRead` (self-scope)** registered across the 4 canonical files:
   `iam/domain/model.go` (const + `validCapabilities`), `iam/domain/catalog.go` (pt-BR description),
   `iam/domain/capability_scope.go` (`ScopeTenant`/self-scope classification per the audit-read
   precedent), `scripts/api-lint/registry_rules.go` `deferredCaps` (+ root-cause comment; operator
   grants separately — never pre-seeded).
4. **Regen FE types:** `npm run gen:api` → `frontend/apps/web/src/lib/api-types/index.d.ts` carries the
   generated `Notification`/list/unread-count shapes.
5. **ADR** under `wiki/decisions/` recording the durable decisions (see ADR section).

## Non-goals (mandatory)

- **No backend implementation** — no notifications table migration, no repository, no handler logic, no
  route wiring into the main mux (all F3.2). F3.1 ends at the generated contract + cap + ADR.
- **No emitter / projector** (F3.3). No `governance_events` read, no recipient-resolution view DDL in
  this feature (the ADR *decides* the mechanism; F3.2/F3.3 *build* the views).
- **No FE wire** (F3.4) — the noop `notifications.ts`/`NotificationsPage` stay untouched this feature.
- **No edit to any approval/documents emit site** (`publish_service.go`, `decision_service.go`,
  `supersede_service.go`, `obsolete_service.go`) — enriching an outbox payload to carry the submitter
  is **HS-2** and is rejected; author resolution is an additive published view.
- **No approver-routing / template / channel / preference / SSE** contract — parked emitter mission.
- **No closed `event_type` enum** — open string (additive-extension commitment).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| New `notifications` paths parse clean | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` → `0 violation(s)` | real |
| Generated Go contract present | `internal/modules/notifications/api/api.gen.go` exists with `Notification`, list, unread-count, mark-read types; `go build ./...` exit 0 | real |
| Generated FE type present + structurally == `NotificationItem` | `npm run gen:api` then `grep -E "Notification" frontend/apps/web/src/lib/api-types/index.d.ts`; field-by-field equality vs `lib/types/index.ts:178` (id, recipient_user_id, event_type, resource_type, resource_id, title, message, status enum, created_at, read_at) | real |
| `event_type` is an open string (not enum); `status` is the closed 3-value enum | inspection of the generated schema | real |
| Cap registered across 4 files + deferred | `go test ./scripts/api-lint/... -run TestSeedRegistryParity` green (cap deferred, not unseeded-fail); `go build ./...` exit 0; grep each of the 4 files for `CapNotificationRead` | real |
| ADR present + ADR-0039 inventory note for the planned author/reader views | file exists under `wiki/decisions/`; `wiki/decisions/index.md` updated | real |
| Spec review | independent reviewer pass (separation of powers at milestone close; feature-level code review on the diff) | real |

> TDD note: F3.1 is contract+codegen (no runtime logic). The "failing test first" discipline applies
> at the guard level — `api-lint` fails before the paths exist / passes after; `TestSeedRegistryParity`
> would fail if the cap were unseeded-and-undeferred, passes once deferred. No fabricated runtime proof.

## ADR needed?

- [x] **Durable decision made → ADR.** Records: the new `notifications` module + per-recipient
  read-state table shape (built in F3.2) + `CapNotificationRead` self-scope rule + the
  **document-lifecycle bundle** emitter decision (5 `event_type → recipient-resolver` rows, selection
  criterion = existing outbox event ∧ existing recipient resolver) + the **author/reader recipient
  resolution via additive owner-published views** (no emit-site edit, no raw cross-module base-table
  read; submitter view + document→CD mapping) + the **additive-extension commitment** to the parked
  emitter mission (open `event_type`, additive views). ADR id assigned at authoring (next free under
  `wiki/decisions/`, expected `0043`); link filled in `plan.md`/`evidence.md`.
