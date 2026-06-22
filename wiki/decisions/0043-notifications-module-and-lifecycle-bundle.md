# ADR 0043 — new `notifications` module + `CapNotificationRead` + document-lifecycle bundle contract

> **Status:** Accepted 2026-06-22
> **Last verified:** 2026-06-22
> **Deciders:** leandrotca.work (operator), MetalDocs backend
> **Context window:** Mission `frontend-screen-completion` · Milestone M3 (Notifications) · Feature F3.1.
> **Supersedes:** none.
> **Related ADRs:**
> - [0039 — Cross-module read boundary](./0039-cross-module-base-table-read-boundary.md) — notifications reads only `metaldocs.v_*` published views (additive views defined here); no raw cross-module base-table read.
> - [0042 — `distribution` module + `CapDistributionRead`](./0042-distribution-module-and-cap.md) — structural precedent for a new module with a deferred cap, contract-first with oapi-codegen.
> - [0040 — `v_cd_obligated_readers`](./0040-cd-obligated-readers-view.md) — reader-targeted recipient source (for `document_published`/`document_superseded`/`document_obsoleted`).
> - [0022 — Authz tiers](./0022-authz-capability-coherence.md) — cap registry + scope; self-scope SQL predicate.
> - [0012 — Contract-first API](./0012-contract-first-api.md); [0024 — Single base path](./0024-openapi-single-base-path.md).
>
> **Related code (Last verified 2026-06-22):**
> - `api/openapi/v1/openapi.yaml` — tag `notifications`; ops `listNotifications`, `getNotificationsUnreadCount`, `markNotificationRead`; schemas `Notification`, `NotificationsListResponse`, `UnreadCountResponse`.
> - `internal/modules/notifications/api/` — `cfg.yaml`, `gen.go`, `api.gen.go` (generated; `package notificationsapi`).
> - `internal/modules/iam/domain/model.go` — `CapNotificationRead = "notification.read"`.
> - `internal/modules/iam/domain/catalog.go` — pt-BR description.
> - `internal/modules/iam/domain/capability_scope.go` — `ScopeTenant` (self-scope enforced in SQL, not this enum).
> - `scripts/api-lint/registry_rules.go` + `apps/api/cmd/metaldocs-api/permissions_test.go` — `deferredCaps`.

## Context

Mission `frontend-screen-completion` M3 builds a real notification inbox end-to-end. The operator
selected the **document-lifecycle bundle** (5 events) as the emitter scope at the HS-1 start gate
(2026-06-22). The existing outbox (`governance_events`, migration 0125) was identified as the source
of record; however, recon revealed that the five emit-sites do not carry the recipient — requiring
additive published views rather than emit-site enrichment.

## Decision

### 1 — New `notifications` module

A new `internal/modules/notifications/` module owns the per-recipient notification inbox. Structure
mirrors the `distribution` module (F3.1 = contract surface only; F3.2 = table + repo + handler + route
wiring; F3.3 = projector; F3.4 = FE wire).

### 2 — `Notification` schema: open `event_type`, closed `status`

The `Notification` schema has `event_type` as an **open string** — the parked emitter mission adds
types additively without breaking generated clients. `status` is a closed 3-value enum (`PENDING |
SENT | READ`) because it is a fixed read-state machine.

The schema is the snake_case mirror of the frozen FE consumer type `NotificationItem`
(`frontend/apps/web/src/lib/types/index.ts:178`) — structural equality is a hard gate in F3.1.

### 3 — Three endpoints, cursor pagination, self-scope

- `GET /notifications` — list (status filter, cursor pagination keyset on `(created_at, id)`)
- `GET /notifications/unread-count` — badge count
- `POST /notifications/{id}/read` — idempotent mark-read (204)

Self-scope: a caller can only see and act on their own rows. Enforced by a SQL predicate
`recipient_user_id = <caller_id>` in every query — not by the ScopeArea/ScopeTenant classification
(which has no "self" concept). `CapNotificationRead` is classified `ScopeTenant` in
`capability_scope.go` solely because the area-vs-tenant grade does not apply; the real guard is the
SQL predicate (F3.2).

### 4 — `CapNotificationRead` (self-scope cap, deferred)

New tier-2 cap `notification.read`. Deliberately **not seeded** to any tenant role; the operator
grants it to roles separately (same deferral pattern as `CapDistributionRead`). Registered in:
`iam/domain/model.go`, `catalog.go`, `capability_scope.go`, `scripts/api-lint/registry_rules.go`
(`deferredCaps`), and `apps/api/cmd/metaldocs-api/permissions_test.go` (`TestEveryCapSeededOrDeferred`
deferred allow-list).

### 5 — Document-lifecycle bundle emitter scope (5 events)

| event_type | Recipient resolver | Notes |
|---|---|---|
| `document_published` | `v_cd_obligated_readers` (reader-targeted) | resource_id = document_id; needs doc→CD mapping (see §6) |
| `document_superseded` | `v_cd_obligated_readers` | same |
| `document_obsoleted` | `v_cd_obligated_readers` | same |
| `signoff_recorded` | submitter via approval instance | author-targeted; needs additive submitter view (see §6) |
| `signoff.rejected` | submitter via approval instance | author-targeted; same |

Selection criterion: existing outbox event (`governance_events.event_type` already emitted) AND
existing recipient resolver. Approver-routing, templates, channels, preferences, reminders/SLA, and
SSE rebuild are **parked** to a separate emitter mission.

### 6 — Recipient resolution: additive owner-published views, NOT emit-site enrichment

Recon finding (decisive): `governance_events` carries `actor_user_id` (the approver), never the
submitter. The five emit-sites carry only fact identifiers; `approval_instances.submitted_by` holds
the submitter. `v_cd_obligated_readers` keys on CD id; governance events carry `resource_id =
document_id` → a document→CD mapping is needed.

**Decision: additive `metaldocs.v_*` published views** — the same mechanism as ADR-0040/0041 (M2
F2.1a/b precedent). Two views are needed:

1. **Submitter view** (approval/approval_instances module publishes): exposes
   `(tenant_id, approval_instance_id, submitted_by, document_id)` at minimum. Named
   `v_approval_instance_submitter` (exact name confirmed in F3.3 spec).
2. **Document→CD mapping** (documents/controlled-documents module publishes): exposes
   `(tenant_id, document_id, controlled_document_id)` to let the projector join
   `v_cd_obligated_readers` on CD id. Named `v_document_cd_mapping` (exact name confirmed in F3.3
   spec).

**Rejected alternative:** enriching the signoff emit-sites (`decision_service.go`) to carry
`submitted_by` in the payload JSON. This edits approval-owned emit-site code — **HS-2**
(approval/workflow boundary). The additive-view path is structurally identical to the ADR-0039 carve-out
used in M2 and is not an HS-2 trigger.

**ADR-0039 inventory note (additive views to be created in F3.2/F3.3):**
- `metaldocs.v_approval_instance_submitter` — authored in F3.2 (approval module publishes; notifications
  projector reads via `hgcrossmodule` view protocol).
- `metaldocs.v_document_cd_mapping` — authored in F3.2 (documents module publishes; notifications
  projector reads via `hgcrossmodule` view protocol).

### 7 — Additive-extension commitment

The `open event_type` + additive-view design commits to extending the notification surface without
breaking existing generated clients: new event types require no schema change, and new recipient
resolvers require new views (not emit-site edits). This is the load-bearing contract for the parked
emitter mission.

## Consequences

- **Positive:** self-scope inbox owned entirely by the new module; no cross-module write path; full
  contract-first codegen flow (oapi-codegen + openapi-typescript); cap deferred = operator controls
  rollout.
- **Negative / to watch:** the projector (F3.3) requires two new published views from other modules;
  those are tracked as HS-2 carve-outs (additive-only, no base-table edits, no emit-site change).
- **F3.2 gate (HS-2 carve-out):** the two views above MUST be authored as pure `CREATE VIEW … AS
  SELECT` statements against existing stable columns, published to `metaldocs.*` schema, with no
  modifications to base tables, triggers, or approval/document service code.
