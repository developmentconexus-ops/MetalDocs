# Planned Endpoints — spec-ready sketches

> **Last verified:** 2026-06-07 (captured during API Contract Hardening Phase C)
> **Status:** NOT served. NOT in the live OpenAPI contract.
> **Origin:** These four operations previously sat in `api/openapi/v1/openapi.yaml` as *published-but-unserved* paths (no runtime handler). Per [`api-contract-hardening.md`](api-contract-hardening.md) **OD-2** (remove-all unserved + backlog-the-planned), an OpenAPI spec must describe only what the API *serves*. They were removed from the live contract in **Phase C** (an unserved path is an OWASP API9 liability and a lie to every SDK/consumer), but — unlike the dead legacy-taxonomy bloc — these map to **real planned features**, so their intended shape is preserved here.

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

### Notifications schemas (as captured)

```yaml
NotificationItem:
  type: object
  required: [id, recipientUserId, eventType, resourceType, resourceId, title, message, status, createdAt]
  properties:
    id: { type: string }
    recipientUserId: { type: string }
    eventType: { type: string }
    resourceType: { type: string }
    resourceId: { type: string }
    title: { type: string }
    message: { type: string }
    status: { type: string, enum: [PENDING, SENT, READ] }
    createdAt: { type: string, format: date-time }
    readAt: { type: string, format: date-time }

ListNotificationsResponse:
  type: object
  required: [items]
  properties:
    items:
      type: array
      items: { $ref: '#/components/schemas/NotificationItem' }

MarkNotificationReadResponse:
  type: object
  required: [id, status, readAt]
  properties:
    id: { type: string }
    status: { type: string, enum: [READ] }
    readAt: { type: string, format: date-time }
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

### Workflow schemas (as captured)

```yaml
WorkflowTransitionRequest:
  type: object
  required: [toStatus]
  properties:
    toStatus: { type: string, enum: [DRAFT, IN_REVIEW, APPROVED, PUBLISHED, ARCHIVED] }
    reason: { type: string }
    assignedReviewer: { type: string }

WorkflowTransitionResponse:
  type: object
  required: [documentId, fromStatus, toStatus]
  properties:
    documentId: { type: string }
    fromStatus: { type: string }
    toStatus: { type: string }
    approvalId: { type: string }
    approvalStatus: { type: string, enum: [PENDING, APPROVED, REJECTED] }
    assignedReviewer: { type: string }

WorkflowApprovalItem:
  type: object
  required: [approvalId, documentId, requestedBy, assignedReviewer, status, requestedAt]
  properties:
    approvalId: { type: string }
    documentId: { type: string }
    requestedBy: { type: string }
    assignedReviewer: { type: string }
    decisionBy: { type: string }
    status: { type: string, enum: [PENDING, APPROVED, REJECTED] }
    requestReason: { type: string }
    decisionReason: { type: string }
    requestedAt: { type: string, format: date-time }
    decidedAt: { type: string, format: date-time }

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
