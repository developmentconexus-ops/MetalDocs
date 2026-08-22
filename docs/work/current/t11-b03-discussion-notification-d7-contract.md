# T11 — B03 Discussion / Notification D7 Contract

> **Status:** OPERATOR-RATIFIED CANDIDATE / GCR-CORRECTED / PENDING INDEPENDENT CHALLENGE + UPSTREAM CONSOLIDATION.  
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
generic EventBus                    absent
external broker                     absent
Redis baseline                      absent
```

The old `78 / operation79-absent` census remains current upstream authority until the entire Product/T1→T9/T11 reopen is promoted. D7 authorizes only the exact candidate delta; it does not silently mutate already-ratified SSOTs.

## 2. Cross-owner atomicity and Authorization ownership

For one accepted Discussion message containing one or more valid explicit Mentions:

```text
accepted explicit Mention
<=>
required DOCUMENT_MENTION Notification exists
```

The smallest enforcement is one caller-owned local PostgreSQL transaction coordinated by application choreography.

Authority partition remains exact:

```text
Organization
  authors current User existence / Company / ENABLED / Group facts

Controlled Documents
  authors Document relationship/state/disclosure predicate facts
  owns intrinsic message/reply/content/Mention facts after admission

Authorization
  alone computes final ALLOW / default DENY

Notifications
  owns Notification identity / recipient / engagement state

application
  gathers/maps facts, owns transaction choreography, owns no semantic rule
```

Corrected create-message flow:

```text
BEGIN Scope
→ normalize unique actor + Mention-target User ids
→ acquire protected Organization subjects in deterministic user_id ASC order
→ gather exact Controlled Documents access/predicate facts
→ Authorization.DecideIn / DecideManyIn computes:
     author may document.discuss for this Document
     every Mention target may currently receive/read this exact Discussion context
→ only after every required decision = ALLOW:
     Controlled Documents validates intrinsic reply/content/Mention structure
     Controlled Documents inserts DiscussionMessage + Mention facts
     Notifications inserts one DOCUMENT_MENTION Notification per unique target/message
→ required same-Scope evidence only where separately ratified
COMMIT
```

No `document.read_discussion` Permission is introduced. D1 remains binding: reading Discussion follows the exact current ability to receive the Document Official / Discussion lens. The frontend and Controlled Documents never reconstruct the permission matrix.

### Offboarding serialization

Create DiscussionMessage joins the current T3 family whose correctness depends on ENABLED User truth.

```text
author
+ every unique Mention target
```

use the existing protected Organization eligibility mechanism inside the same Scope. Multiple Users are resolved after uniqueness in deterministic `user_id ASC` order to avoid an avoidable multi-row deadlock cycle.

If offboarding linearizes first, the operation observes DISABLED and fails closed. If the message transaction linearizes first, it may commit under valid current truth and offboarding then prevents future actions. Later access drift never rewrites the accepted Mention; D5 controls current Notification presentability.

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

Discussion list semantics remain seek/cursor based. A bounded `anchor_message_id` navigation mode may retrieve a page containing the exact target message for Notification deep-linking; this is navigation over the same resource family, not a separate message-detail operation.

Mention candidate search is purpose-built and Document-scoped. Organization candidate search is not authority by itself: application composes candidate User facts with exact Controlled Documents predicate facts and current Authorization decisions before any candidate is returned. It never reuses an administrative User list as disclosure authority.

## 4. Message wire and replay law

Message content is a minimal closed semantic sequence, not HTML/ProseMirror/Lexical persistence authority:

```text
MessageContentSegment =
  Text { kind:text, text }
  | Mention { kind:mention, user_id }
```

Accepted Mention authority is stable `user_id`. Response presentation may resolve current bounded `UserReference` enrichment without changing persisted Mention identity.

### Completed idempotency replay

The existing global replay precedence remains binding.

For a recognized completed `createDocumentDiscussionMessage` replay:

```text
current caller session + CSRF
→ current caller document.discuss + source disclosure recheck
→ completed key/fingerprint recognition
→ return stored message_id
→ zero new DiscussionMessage
→ zero new Mention
→ zero new Notification
```

Do **not** re-run the historical Mention targets' current eligibility on completed replay. Later target access/offboarding is governed by D5 presentability; replay is not a second Mention command.

The semantic fingerprint includes the exact Document id, optional reply target and ordered normalized Text/Mention content. ReplaySnapshot remains free of message/free text.

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

`listNotifications` serves Quick Inbox and full `/notifications` Inbox from one Notifications owner read authority plus server-side current-disclosure composition.

Candidate view filter:

```text
active | unread | archived
```

First-page summary may expose the current user's presentable non-archived `unseen_count` and `unread_count`. These are derived engagement facts needed by the UI, not generic list totals or durable counters.

### Individual engagement

Conceptual PATCH:

```text
read?: boolean
archived?: boolean
```

At least one field is required. Deliberate engagement implies `seen_at` if previously absent. `seen=false` is not a command.

No ETag domain is added. Recipient engagement is personal reversible state; same-property concurrent writes linearize with last accepted action winning, while omitted properties remain unchanged.

A direct single-Notification engagement operation on an item that is absent, foreign-recipient or no longer presentable returns the ordinary non-disclosing not-found behavior; it never becomes a source-existence oracle.

### Batch seen

`markNotificationsSeen` accepts a bounded unique list of Notification ids and monotonically sets `seen_at` only on the intersection that is, at operation time:

```text
owned by current recipient
+ currently presentable
```

Absent, foreign-recipient or now-non-presentable ids produce zero mutation and no per-id status/cardinality detail. The response never reveals which requested ids existed.

### Mark all read

`markAllNotificationsRead` applies only to the current user's Notifications that are, at the operation point, presentable + non-archived + unread. It sets `seen_at` if absent and `read_at` to the accepted action time. New or currently non-presentable items are not swept into the operation.

No Launch operations exist for mark-all-unread, mark-unseen, archive-all, snooze or preferences.

## 6. Notification presentability before pagination and counts

Notifications persists only its own authority:

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

Current Launch source:

```text
DOCUMENT_MENTION
  document_id
  message_id
```

Notifications does **not** persist copied source ACL/presentability truth, Document title/code, message text or mutable User profile data merely to simplify reads.

`application/notifications` owns the cross-owner read choreography. Public pagination and counts must be over **presentable output**, never a Notification page that is post-filtered by React.

Conceptual read law:

```text
Notifications candidate scan
  current recipient + requested engagement filter
  canonical Notification ordering
→ batch source identities
→ Controlled Documents source/access predicate facts
→ Organization current subject facts
→ Authorization exact Decide/DecideMany
→ retain only currently presentable candidates
→ continue bounded candidate scan until:
     requested presentable page + lookahead is satisfied
     OR candidates are exhausted
→ compose source presentation only for retained items
```

Public cursor state is opaque and tied to the canonical candidate seek position/filter semantics. Hidden candidate identities are never returned. A page may require multiple bounded internal scans, but cannot expose sparse pages merely because inaccessible candidates were filtered after paging.

The same current-disclosure composition owns `unseen_count` and `unread_count`. Launch may evaluate counts in bounded internal chunks; no durable copied counter/current-access projection becomes authority. A measured scale failure is a reopen trigger for a rebuildable optimization that proves equivalence.

Read-time presentation may compose currently disclosable source enrichment such as actor `UserReference`, `DocumentReference` and bounded message preview. These are returned only after current disclosure succeeds.

Notifications has no new Product Permission. Ordinary Inbox access is structurally recipient-self-only under the authenticated ENABLED User; source presentation additionally rechecks current source disclosure.

## 7. Realtime signaling — +1

```text
86 GET /api/v1/notifications/events
   streamNotificationEvents
   Content-Type: text/event-stream
```

The SSE stream is ephemeral invalidation/wake-up mechanism only. Payload carries no source Document/message truth and is equivalent to a bounded `notifications.changed` signal.

Accepted call graph:

```text
transport/http
→ application/notifications stream choreography
→ narrow application-owned/consumer-owned subscription port
→ platform realtime mechanism
```

`transport -> platform` remains forbidden. The platform mechanism owns connection/subscription mechanics only; it owns no Notification truth or access rule.

Frontend reaction:

```text
SSE signal
→ invalidate/refetch Notifications query
→ canonical GET /notifications restores truth
```

Lost/disconnected SSE delivery never loses a Notification. Reconnect, focus/refetch and ordinary list reads reconcile canonical persistent state.

### Post-commit wake-up law

Every successful Notification state change that can affect another open tab emits best-effort recipient invalidation **after commit**:

```text
DOCUMENT_MENTION Notification creation
mark seen
mark read
mark unread
archive
unarchive
mark all read
```

Semantic owners never invoke realtime directly. The relevant application leaf calls the narrow wake-up mechanism only after semantic commit succeeds. Wake-up failure never rolls back committed Product state and is not retried through River merely for UI freshness.

The D8 Launch mechanism remains an in-process coalescing hub behind the narrow port. WebSocket, Redis and a broker are not implied.

## 8. Audit / History disposition

Discussion and Notifications do not create duplicate evidence streams merely because they exist.

```text
DiscussionMessage
  immutable owner record with trusted author/time/content
  → no duplicate semantic AuditEvent solely to copy it
  → not injected into Document lifecycle History timeline

Notification creation / engagement / realtime
  → no mandatory semantic AuditEvent in Launch
```

This mirrors the existing SubmissionFeedback principle. A future regulatory/customer requirement for messaging or notification audit is a bounded T3 reopen trigger.

## 9. Persistence enforcement obligations

Upstream T8-D consolidation must preserve at minimum:

```text
new notifications.* owner namespace
identity-only cross-owner referential integrity only
unique DOCUMENT_MENTION Notification per recipient + message
read_at present      -> seen_at present
archived_at present  -> seen_at present
immutable accepted DiscussionMessage / Mention through application privileges/structure
reply_to_message_id cannot cross Document Discussion
```

Exact table decomposition/constraints remain T8-D/T11 realization detail, but the protected properties are not optional.

## 10. Event/broker posture

```text
typed event semantics / event-friendly boundaries   allowed
current generic internal EventBus                    absent
current external broker                              absent
River durable future-work mechanism                 preserved
SSE realtime invalidation                            admitted
```

Reopen an internal EventStore/EventBus only when real producer-consumer pressure proves that direct application choreography causes producers to know multiple independently evolving consumers or otherwise creates a structural dead end. Reopen an external broker only on independent distributed-scale/process/trust/fan-out evidence.

## 11. Exact candidate census delta

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

## 12. Runtime / OpenAPI closure gate

Operation 86 remains part of the candidate `/api/v1` census only if the selected Go OpenAPI boundary proves server-side `text/event-stream` generation/handling without a manual parallel route or DTO registry.

Until that proof exists:

```text
SSE Product/UX mechanism candidate = accepted
manual contract escape hatch        = forbidden
```

Failure of one generator/toolchain is a bounded mechanism/tooling reopen; it does not authorize silently dropping realtime or moving the route outside contract authority.

T8-G consolidation owns heartbeat/flush/proxy-timeout/shutdown/resource-limit behavior for the long-lived response while preserving the one-application-runtime baseline.

## 13. Proof strategy

Implementation/readiness proof must be able to falsify at least:

```text
Authorization is the only final Mention-target ALLOW/DENY owner
protected author/target eligibility serializes with offboarding in deterministic id order
Notification persistence failure -> zero accepted DiscussionMessage
same Idempotency-Key replay -> same message_id + zero duplicate Notification
completed replay does not re-run historical Mention-target eligibility
same target Mention repeated -> one Notification per target/message
Mention invalid before commit -> zero Message + zero Notification
non-presentable Notification -> absent from items/counts + zero metadata leak
presentable paging has no frontend post-filter authority / sparse-page leak
batch seen cannot disclose absent/foreign/non-presentable ids
mark unread -> never clears seen
archive/unarchive -> preserves seen/read
mark-all-read -> excludes currently non-presentable Notifications
lost SSE signal -> canonical GET still recovers Notification
SSE transport -> application -> mechanism direction is preserved
engagement mutations invalidate other tabs only after commit
SSE payload -> zero source business truth
Discussion/Notification does not create duplicate Audit/History truth
persistence constraints enforce unique source + engagement implications
OpenAPI boundary realizes server-side text/event-stream without manual parallel route
contract census -> exactly 86 application operations
first-party dependency graph -> zero owner→owner imports + zero generic EventBus
```

## 14. Reopen triggers

Reopen D7 if evidence proves:

```text
same-transaction Message/Notification creation causes greater accidental cross-owner complexity than it removes;
server-side presentability composition cannot sustainably satisfy pagination/count correctness at measured scale;
a generic EventStore becomes justified by multiple real independent consumers;
an external broker becomes justified by measured/process/trust/fan-out requirements;
SSE cannot satisfy the required same-origin browser notification-update experience;
Notification engagement needs a concurrency invariant stronger than current recipient-action linearization;
wire implementation reveals a smaller operation set that preserves every ratified Product/UX invariant without screen-shaped APIs or hidden authority.
```
