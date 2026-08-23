---
id: api-operation-census
kind: authority
owner: architecture
summary: Owns the current 86-operation /api/v1 application census after the operator-ratified T11 Discussion/Notifications bounded reopen.
---

# API operation census

The current Launch `/api/v1` application census contains **86 operations**.

Current semantic authority is the combination of:

```text
docs/product/journeys.md
+ docs/decisions/discussion-notifications-launch.md
```

The bounded T8-E read-symmetry precision previously added two reads to the original 76-operation census:

```text
GET /api/v1/users/{user_id}/profile
operationId: getUserProfile

GET /api/v1/areas/{area_id}/lifecycle
operationId: getAreaLifecycle
```

The T11 Discussion / `@mention` / Notifications bounded reopen then added exactly eight application operations:

```text
79 GET   /api/v1/documents/{document_id}/discussion/messages
   operationId: listDocumentDiscussionMessages

80 POST  /api/v1/documents/{document_id}/discussion/messages
   operationId: createDocumentDiscussionMessage

81 GET   /api/v1/documents/{document_id}/discussion/mention-candidates
   operationId: searchDocumentDiscussionMentionCandidates

82 GET   /api/v1/notifications
   operationId: listNotifications

83 PATCH /api/v1/notifications/{notification_id}/engagement
   operationId: updateNotificationEngagement

84 PUT   /api/v1/notifications/seen
   operationId: markNotificationsSeen

85 PUT   /api/v1/notifications/read
   operationId: markAllNotificationsRead

86 GET   /api/v1/notifications/events
   operationId: streamNotificationEvents
   representation: text/event-stream
```

## Count proof

```text
original journeys census                    76
bounded read-symmetry precision             +2
T11 Discussion/Mention/Notifications reopen +8
                                            ---
current application census                  86
```

## Idempotency / ETag / exact-byte census

Exactly one of the eight T11 operations is a new non-idempotent semantic POST creation:

```text
createDocumentDiscussionMessage
```

Therefore current cross-contract counts are:

```text
application operations        86
Idempotency-Key creations     11
ETag read / mutation domains  13 / 13
exact-byte resources          4
```

The T11 creation reuses the current global durable Idempotency-Key law; its ReplaySnapshot stores stable `message_id` only and does not copy message/free text.

## Authority law

`docs/decisions/discussion-notifications-launch.md` is the bounded current authority that supersedes prior Product/T6/T8-E/T8-F/T8-H/T9 sentences asserting that operation 79 was absent or that the Launch application census was closed at 78.

All unchanged journey meaning and wire laws remain owned by `docs/product/journeys.md` and `docs/architecture/wire-contract.md`. The eight-operation delta and its exact Discussion/Notification semantics are owned by the bounded current authority until a future substantive rewrite absorbs them.

Any operation **87 or later** requires unchanged semantic normalization already permitted by current authority or a new explicit bounded Product/T6 reopen. Framework convenience, screen shape or implementation preference is not authority to add an operation.

T8-E executable realization must prove exactly **86** application operations and the current 11/13/4 supporting censuses.