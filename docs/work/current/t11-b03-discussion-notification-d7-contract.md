# T11 — B03 Discussion / Notification D7 Contract

> **Status:** OPERATOR-RATIFIED CANDIDATE / PENDING UPSTREAM CONSOLIDATION.  
> **Parent:** `t11-b03-discussion-notification-mini-design.md`.  
> **Reasoning authority:** `developmentconexus-ops/conexus-methodology/METHOD.md`.  
> **Implementation:** BLOCKED.  
> **Current upstream 78-operation authority remains effective until the full bounded reopen is coherently consolidated.**

## 1. Decision

D7 closes the candidate technical contract for the operator-ratified Launch V1 Document Discussion + `@mention` + Notifications capability.

Candidate post-reopen invariants:

```text
semantic owners                     4 business + 2 supporting
stable SPA routes                   11
PermissionCode values               16
application operations              86
Idempotency-Key creations           11
ETag read / mutation domains        13 / 13
exact-byte resources                4
generic EventBus                     absent
external broker                      absent
Redis baseline                       absent
```

The old `78 / operation79-absent` census remains historical/current upstream authority until the entire Product/T1→T9/T11 reopen is promoted. D7 authorizes the exact candidate delta; it does not silently mutate already-ratified SSOTs.

## 2. Cross-owner atomicity

For one accepted Discussion message containing one or more valid explicit Mentions:

```text
accepted explicit Mention
<=>
required DOCUMENT_MENTION Notification exists
```

The smallest enforcement is one caller-owned local PostgreSQL transaction coordinated by application choreography:

```text
BEGIN Scope
→ current author AuthZ/disclosure recheck
→ Controlled Documents validates reply + Mentions and creates DiscussionMessage/Mention facts
→ Notifications creates one DOCUMENT_MENTION Notification per unique target/message
→ required same-scope evidence only where separately ratified
COMMIT
```

No semantic owner imports another owner. Notification persistence failure aborts the message transaction. No River job, generic outbox, EventBus or external broker mediates this local invariant.

## 3. Discussion operations — +3

```text
79 GET  /api/v1/documents/{document_id}/discussion/messages
   listDocumentDiscussionMessages

80 POST /api/v1/documents/{document_id}/discussion/messages
   createDocumentDiscussionMessage

81 GET  /api/v1/documents/{document_id}/discussion/mention-candidates?query=...
   searchDocumentDiscussionMentionCandidates
```

`createDocumentDiscussionMessage` is the 11th `Idempotency-Key` creation. Its replay snapshot contains only stable result identity such as `message_id`; message/free text is excluded from durable replay data.

Discussion list semantics remain seek/cursor based. A bounded `anchor_message_id` navigation mode may retrieve a page that contains the exact target message for Notification deep-linking; this is continuation/navigation of the same resource family, not a separate message-detail operation.

Mention candidate search is purpose-built and Document-scoped. It never reuses an administrative User list and always revalidates exact eligibility when the message is accepted.

## 4. Message wire law

Message content is a minimal closed semantic sequence, not HTML/ProseMirror authority:

```text
MessageContentSegment =
  Text { kind:text, text }
  | Mention { kind:mention, user_id }
```

Accepted Mention authority is stable `user_id`. Response presentation may resolve current bounded `UserReference` enrichment without changing persisted Mention identity.

## 5. Notification operations — +4 state operations

```text
82 GET   /api/v1/notifications
   listNotifications

83 PATCH /api/v1/notifications/{notification_id}/engagement
   updateNotificationEngagement

84 PUT   /api/v1/notifications/seen
   markNotificationsSeen

85 PUT   /api/v1/notifications/read
   markAllNotificationsRead
```

`listNotifications` serves Quick Inbox and full `/notifications` Inbox from one Notifications owner read authority. Cursor/seek laws remain consistent with the existing wire contract.

Candidate view filter:

```text
active | unread | archived
```

First-page summary may expose the current user's presentable non-archived `unseen_count` and `unread_count`. These are derived engagement facts required by the UI, not generic list total counts or separate durable counters.

### Individual engagement

Conceptual PATCH:

```text
read?: boolean
archived?: boolean
```

At least one field is required. Deliberate engagement implies `seen_at` if previously absent. `seen=false` is not a command.

No ETag domain is added. Recipient engagement is personal reversible state; same-property concurrent writes linearize with last accepted action winning, while omitted properties remain unchanged.

### Batch seen

`markNotificationsSeen` accepts a bounded unique list of Notification ids and monotonically sets `seen_at` where absent. It is naturally idempotent and exists to avoid one request per viewport item.

### Mark all read

`markAllNotificationsRead` applies only to the current user's Notifications that are, at the operation point, presentable + non-archived + unread. It sets `seen_at` if absent and `read_at` to the accepted action time. New or currently non-presentable items are not swept into the operation.

No Launch operations exist for mark-all-unread, mark-unseen, archive-all, snooze or preferences.

## 6. Notification read composition

Notifications persists only its authority:

```text
notification_id
recipient_user_id
kind
closed source identities
created_at
seen_at?
read_at?
archived_at?
```

For current Launch:

```text
DOCUMENT_MENTION
  document_id
  message_id
```

List presentation may compose currently disclosable source enrichment such as actor `UserReference`, `DocumentReference` and a bounded message preview. These values are resolved at read time and are not copied into Notifications as source-business authority.

Notifications has no new Product Permission. Ordinary Inbox access is structurally recipient-self-only under the authenticated ENABLED User. Source presentation additionally rechecks current source disclosure.

## 7. Realtime signaling — +1

```text
86 GET /api/v1/notifications/events
   streamNotificationEvents
   Content-Type: text/event-stream
```

The SSE stream is ephemeral invalidation/wake-up mechanism only. Payload contains no source Document/message truth and is equivalent to a bounded `notifications.changed` signal.

Frontend reaction:

```text
SSE signal
→ invalidate/refetch Notifications query
→ canonical GET /notifications restores truth
```

Lost/disconnected SSE delivery never loses a Notification. Reconnect, focus/refetch and ordinary list reads reconcile canonical persistent state.

A successful business transaction does not depend on successful realtime wake-up. Wake-up mechanism failure therefore does not roll back already-committed Notification truth and is not retried through River merely to preserve realtime UI freshness.

The exact runtime wake-up implementation is intentionally deferred to D8 evidence: in-process hub, PostgreSQL LISTEN/NOTIFY or another proven mechanism may satisfy this seam. WebSocket, Redis and a broker are not implied.

## 8. Event/broker posture

```text
typed event semantics / event-friendly boundaries   allowed
current generic internal EventBus                    absent
current external broker                             absent
River durable future-work mechanism                 preserved
SSE realtime invalidation                           admitted
```

Reopen an internal EventStore/EventBus only when real producer-consumer pressure proves that direct application choreography causes producers to know multiple independently evolving consumers or otherwise creates a structural dead end. Reopen an external broker only on independent distributed-scale/process/trust/fan-out evidence.

## 9. Exact candidate census delta

```text
Discussion / Mention               +3
Notification engagement state      +4
Notification SSE signaling         +1
                                  ----
                                   +8

current upstream                   78
candidate after reopen             86
```

Exactly one of the eight new operations is a non-idempotent semantic creation requiring durable `Idempotency-Key`:

```text
createDocumentDiscussionMessage
```

Therefore:

```text
Idempotency-Key creations 10 → 11
ETag domains              13/13 unchanged
exact-byte resources       4 unchanged
```

## 10. Proof strategy

Implementation/readiness proof must be able to falsify at least:

```text
Notification persistence failure -> zero accepted DiscussionMessage
same Idempotency-Key retry -> same message_id + zero duplicate Notification
same target Mention repeated -> one Notification per target/message
Mention invalid before commit -> zero Message + zero Notification
non-presentable Notification -> absent from items/counts + zero metadata leak
mark unread -> never clears seen
archive/unarchive -> preserves seen/read
mark-all-read -> excludes currently non-presentable Notifications
lost SSE signal -> canonical GET still recovers Notification
SSE payload -> zero source business truth
contract census -> exactly 86 application operations
first-party dependency graph -> zero owner→owner imports + zero generic EventBus
```

## 11. Reopen triggers

Reopen D7 if evidence proves:

```text
same-transaction Message/Notification creation causes greater accidental cross-owner complexity than it removes;
a generic EventStore becomes justified by multiple real independent consumers;
an external broker becomes justified by measured/process/trust/fan-out requirements;
SSE cannot satisfy the required same-origin browser notification-update experience;
Notification engagement needs a concurrency invariant stronger than current recipient-action linearization;
wire implementation reveals a smaller operation set that preserves every ratified Product/UX invariant without screen-shaped APIs or hidden authority.
```
