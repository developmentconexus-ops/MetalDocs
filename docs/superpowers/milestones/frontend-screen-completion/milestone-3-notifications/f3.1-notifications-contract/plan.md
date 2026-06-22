# Feature F3.1 — Plan (notifications-contract)

> Input: `spec.md` (approved 2026-06-22). Engine: writing-plans (inline). This is the **how**.
> ADR: **0043** (next free; `wiki/decisions/0042` is the last). Link recorded in `evidence.md`.

## Files touched

| # | File | Change |
|---|------|--------|
| 1 | `api/openapi/v1/openapi.yaml` | + `notifications` tag (after `distribution` tag, line ~34); + 3 paths before `components:` (line ~3779): `/notifications`, `/notifications/unread-count`, `/notifications/{id}/read`; + schemas `Notification`, `NotificationsListResponse`, `UnreadCountResponse` (after the Distribution* schemas, ~line 4119). |
| 2 | `internal/modules/notifications/api/cfg.yaml` | NEW — `package: notificationsapi`, `include-tags: [notifications]`, strict-server (mirror distribution `cfg.yaml`). |
| 3 | `internal/modules/notifications/api/gen.go` | NEW — `package notificationsapi` + `//go:generate` oapi-codegen line (mirror distribution `gen.go`). |
| 4 | `internal/modules/notifications/api/api.gen.go` | GENERATED via `go generate ./internal/modules/notifications/api/...`. |
| 5 | `internal/modules/iam/domain/model.go` | + `CapNotificationRead = "notification.read"` const; add to `validCapabilities`. |
| 6 | `internal/modules/iam/domain/catalog.go` | + pt-BR description row for `CapNotificationRead`. |
| 7 | `internal/modules/iam/domain/capability_scope.go` | + self-scope classification for `CapNotificationRead` (mirror the read-cap precedent). |
| 8 | `scripts/api-lint/registry_rules.go` | + `CapNotificationRead` to `deferredCaps` (+ root-cause comment: operator-granted, not seeded). |
| 9 | `frontend/apps/web/src/lib/api-types/index.d.ts` | GENERATED via `npm run gen:api`. |
| 10 | `wiki/decisions/0043-notifications-module-and-lifecycle-bundle.md` | NEW ADR. |
| 11 | `wiki/decisions/index.md` | + ADR-0043 row. |

## OpenAPI schema (exact)

```yaml
# tag (after distribution tag)
  - name: notifications
    description: Per-recipient notification inbox — read surface (list / unread-count / mark-read). Projected from governance_events for the document-lifecycle bundle (ADR-0043). Self-scope.

# paths (before components:)
  /notifications:
    get:
      summary: List the caller's notifications (self-scope), newest first
      tags: [notifications]
      operationId: listNotifications
      parameters:
        - { name: status, in: query, schema: { type: string, enum: [PENDING, SENT, READ] } }
        - name: cursor
          in: query
          description: Opaque forward keyset cursor from a prior page's page.next_cursor.
          schema: { type: string }
        - name: limit
          in: query
          schema: { type: integer, minimum: 1, maximum: 100 }
      responses:
        '200': { description: ok, content: { application/json: { schema: { $ref: '#/components/schemas/NotificationsListResponse' } } } }
        '400': { $ref: '#/components/responses/BadRequest' }
        '401': { $ref: '#/components/responses/Unauthorized' }
        '403': { $ref: '#/components/responses/Forbidden' }
        '500': { $ref: '#/components/responses/InternalServerError' }
  /notifications/unread-count:
    get:
      summary: Count the caller's unread (PENDING/SENT) notifications
      tags: [notifications]
      operationId: getNotificationsUnreadCount
      responses:
        '200': { description: ok, content: { application/json: { schema: { $ref: '#/components/schemas/UnreadCountResponse' } } } }
        '401': { $ref: '#/components/responses/Unauthorized' }
        '403': { $ref: '#/components/responses/Forbidden' }
        '500': { $ref: '#/components/responses/InternalServerError' }
  /notifications/{id}/read:
    post:
      summary: Mark one of the caller's notifications read (idempotent)
      tags: [notifications]
      operationId: markNotificationRead
      parameters:
        - { name: id, in: path, required: true, schema: { type: string, format: uuid } }
      responses:
        '204': { description: marked read }
        '401': { $ref: '#/components/responses/Unauthorized' }
        '403': { $ref: '#/components/responses/Forbidden' }
        '404': { $ref: '#/components/responses/NotFound' }
        '500': { $ref: '#/components/responses/InternalServerError' }

# schemas (after Distribution* schemas)
    Notification:
      type: object
      required: [id, recipient_user_id, event_type, resource_type, resource_id, title, message, status, created_at]
      properties:
        id: { type: string, format: uuid }
        recipient_user_id: { type: string }
        event_type:
          type: string
          description: Open string. M3 emits document_published / document_superseded / document_obsoleted / signoff_recorded / signoff.rejected; the parked emitter mission adds more additively (ADR-0043). Not a closed enum.
        resource_type: { type: string }
        resource_id: { type: string }
        title: { type: string }
        message: { type: string }
        status: { type: string, enum: [PENDING, SENT, READ] }
        created_at: { type: string, format: date-time }
        read_at: { type: string, format: date-time, nullable: true }
    NotificationsListResponse:
      type: object
      required: [items, page]
      properties:
        items: { type: array, items: { $ref: '#/components/schemas/Notification' } }
        page: { $ref: '#/components/schemas/CursorPage' }
    UnreadCountResponse:
      type: object
      required: [count]
      properties:
        count: { type: integer, minimum: 0 }
```

## Test strategy (guard-level TDD — no runtime logic in F3.1)

1. **api-lint red→green:** run `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` *before* adding `deferredCaps` entry → expect the new cap to surface as unseeded/undocumented (red); add registration → `0 violation(s)` (green).
2. **Registry parity:** `go test ./scripts/api-lint/... -run TestSeedRegistryParity` green after cap deferred.
3. **Build:** `go build ./...` exit 0 (generated `api.gen.go` compiles standalone; no handler implements the interface yet — that is F3.2, and an unused generated interface compiles).
4. **Generated-type equality:** `npm run gen:api`; diff generated `Notification` field set vs FE `NotificationItem` (`lib/types/index.ts:178`) — same 10 fields, `status` 3-enum, `event_type` open string.
5. **Cap presence:** grep `CapNotificationRead` in all 4 registration files.

## Ordering

1. OpenAPI tag + schemas + paths.
2. `cfg.yaml` + `gen.go`; `go generate` → `api.gen.go`.
3. Cap registration across 4 files.
4. `go build ./...`; `api-lint -strict`; `TestSeedRegistryParity`.
5. `npm run gen:api`; verify FE type equality.
6. ADR-0043 + index row.
7. `evidence.md`.
