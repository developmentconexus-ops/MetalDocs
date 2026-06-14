# Planned Endpoints — spec-ready sketches

> **Last verified:** 2026-06-08 (Phase E1: all schema field names normalised to snake_case; prior: 2026-06-07)
> **Status:** NOT served. NOT in the live OpenAPI contract.
> **Origin:** These four operations previously sat in `api/openapi/v1/openapi.yaml` as *published-but-unserved* paths (no runtime handler). Per [`api-contract-hardening.md`](../_archive/backlog/api-contract-hardening.md) **OD-2** (remove-all unserved + backlog-the-planned), an OpenAPI spec must describe only what the API *serves*. They were removed from the live contract in **Phase C** (an unserved path is an OWASP API9 liability and a lie to every SDK/consumer), but — unlike the dead legacy-taxonomy bloc — these map to **real planned features**, so their intended shape is preserved here.

## Re-introduction rule

When a feature below is built, re-add its path **and** its schemas to `api/openapi/v1/openapi.yaml` **in the same change that wires the handler** — never spec-ahead-of-handler again. Add a `permissions.go` route rule with the capability noted below, and a `permissions_test.go` row. Migrate `ApiErrorEnvelope` → `Problem` (RFC 9457) per AD-2 when re-adding (the sketches below still show the legacy envelope as captured; do not copy it forward).

---

## Feature A — Notifications

User-facing operational notification feed. Capability: **`CapDocumentView`** (read-side) for both ops (the mark-read is a read-side state flip, gated the same as the list per the original `permissions.go` rows).

### A.1 `GET /notifications`

List operational notifications for the authenticated user.

Query params:
- `recipientUserId` (string, optional)
- `status` (string, optional, enum `PENDING | SENT | READ`)
- `limit` (integer, optional, 1..200 — **clamp to 100** on re-add to match the design system / Phase E)

Response `200` → `ListNotificationsResponse`.

### A.2 `POST /notifications/{notificationId}/read`

Mark a notification as read. Path param `notificationId` (string, required).

Response `200` → `MarkNotificationReadResponse`. `404` if not found.

### Notifications schemas (as captured, casing normalised to snake_case per E1)

```yaml
NotificationItem:
  type: object
  required: [id, recipient_user_id, event_type, resource_type, resource_id, title, message, status, created_at]
  properties:
    id: { type: string }
    recipient_user_id: { type: string }
    event_type: { type: string }
    resource_type: { type: string }
    resource_id: { type: string }
    title: { type: string }
    message: { type: string }
    status: { type: string, enum: [PENDING, SENT, READ] }
    created_at: { type: string, format: date-time }
    read_at: { type: string, format: date-time }

ListNotificationsResponse:
  type: object
  required: [items]
  properties:
    items:
      type: array
      items: { $ref: '#/components/schemas/NotificationItem' }

MarkNotificationReadResponse:
  type: object
  required: [id, status, read_at]
  properties:
    id: { type: string }
    status: { type: string, enum: [READ] }
    read_at: { type: string, format: date-time }
```

---

## Feature B — Workflow transitions & approvals

Document workflow status machine + approval trail. Distinct from the live `approval/*` instance/sign-off surface — this is the higher-level document-status transition API the audit found speced-but-unserved.

### B.1 `POST /workflow/documents/{documentId}/transitions`

Transition a document's workflow status. Capability: **`CapDocumentSubmit`**.

Path param `documentId` (via `#/components/parameters/DocumentId`). Request body required → `WorkflowTransitionRequest`. Response `200` → `WorkflowTransitionResponse`. `409` on invalid transition for the current status.

### B.2 `GET /workflow/documents/{documentId}/approvals`

List the approval trail for a document. Capability: **`CapDocumentView`**.

Path param `documentId`. Response `200` → `ListWorkflowApprovalsResponse`.

### Workflow schemas (as captured, casing normalised to snake_case per E1)

```yaml
WorkflowTransitionRequest:
  type: object
  required: [to_status]
  properties:
    to_status: { type: string, enum: [DRAFT, IN_REVIEW, APPROVED, PUBLISHED, ARCHIVED] }
    reason: { type: string }
    assigned_reviewer: { type: string }

WorkflowTransitionResponse:
  type: object
  required: [document_id, from_status, to_status]
  properties:
    document_id: { type: string }
    from_status: { type: string }
    to_status: { type: string }
    approval_id: { type: string }
    approval_status: { type: string, enum: [PENDING, APPROVED, REJECTED] }
    assigned_reviewer: { type: string }

WorkflowApprovalItem:
  type: object
  required: [approval_id, document_id, requested_by, assigned_reviewer, status, requested_at]
  properties:
    approval_id: { type: string }
    document_id: { type: string }
    requested_by: { type: string }
    assigned_reviewer: { type: string }
    decision_by: { type: string }
    status: { type: string, enum: [PENDING, APPROVED, REJECTED] }
    request_reason: { type: string }
    decision_reason: { type: string }
    requested_at: { type: string, format: date-time }
    decided_at: { type: string, format: date-time }

ListWorkflowApprovalsResponse:
  type: object
  required: [items]
  properties:
    items:
      type: array
      items: { $ref: '#/components/schemas/WorkflowApprovalItem' }
```

---

## Not backlogged (deleted outright in Phase C)

These were also unserved but are **dead, not planned** — superseded or abandoned. They were removed without capture:

- Legacy taxonomy bloc (`/document-profiles*`, `/process-areas*`, `/document-subjects*`, `/document-types*`, `/document-families`, `/document-templates`, `/document-departments*`, `/document-areas/{code}`) — superseded by the canonical `/taxonomy/*` surface.
- `/operations/stream` (SSE Operations Center) — no handler, no consumer.
- `/attachments/{attachmentId}/content` — no handler, no consumer.
- `/telemetry/mddm-shadow-diff` — no handler, no consumer.
