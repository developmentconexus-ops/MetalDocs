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

## Authority / historical-snapshot law

`docs/decisions/discussion-notifications-launch.md` is the bounded current authority for operations 79–86 and supersedes prior current-tense Product/T6/T8-E/T8-F/T8-H/T9 statements asserting that operation 79 was absent or that the Launch application census was closed at 78.

This page is the **sole current numeric census authority**. Therefore any older `78`, `78/78`, `operation 79 absent`, or `no operation 79` statement found in:

```text
Product/T6 authority
T4/T8/T9/T10 architecture pages
T8-F/T8-H/T9/T10 ratification records
frontend-read/responsible-owner precision provenance
transition/content-integrity closure prose
other pre-T11 closure snapshots
```

is interpreted as one of:

```text
historical stage snapshot, truthful for that stage when ratified
OR
bounded current-tense clause superseded by this T11 census authority
```

It never overrides the current `86 / 11 / 13-13 / 4` census.

All unchanged journey and wire laws remain owned by `docs/product/journeys.md` and `docs/architecture/wire-contract.md`. The eight-operation delta and exact Discussion/Notification semantics are owned by `docs/decisions/discussion-notifications-launch.md` until a future substantive rewrite absorbs them.

Any operation **87 or later** requires unchanged semantic normalization already permitted by current authority or a new explicit bounded Product/T6 reopen. Framework convenience, screen shape or implementation preference is not authority to add an operation.

T8-E executable realization must prove exactly **86** application operations and the current 11/13/4 supporting censuses.