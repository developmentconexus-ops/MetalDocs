---
id: notification-inbox-recognition-read
kind: authority
owner: architecture
summary: Operator-ratified B08 precision making each presentable Notification Inbox item independently human-recognizable from one server-authored op82 row without copied source authority or browser fan-out.
---

# Notification Inbox human-recognition read precision

> **Status:** OPERATOR-RATIFIED / CURRENT BOUNDED AUTHORITY  
> **Ratified:** 2026-08-23  
> **Block:** B08 — Notifications Full Inbox  
> **Method:** DevelopmentConexus Engineering Method + Frontend Product Experience Planning Method v2.2  
> **Impacts:** T6 Notification Inbox meaning, T8-E `listNotifications` read shape, T8-F B01N/B08 consumption.  
> **Implementation:** BLOCKED.

## 1. Decision outcome

DevelopmentConexus Decision Core outcome:

```text
CURRENT NOTIFICATION OWNER/LIFECYCLE STRUCTURE CONFIRMED
+ BOUNDED HUMAN-RECOGNITION READ PROJECTION PRECISION
```

The current route, owner, engagement lifecycle and operations remain sufficient:

```text
GET   /api/v1/notifications
      operation 82 listNotifications

PATCH /api/v1/notifications/{notification_id}/engagement
      operation 83 updateNotificationEngagement

PUT   /api/v1/notifications/seen
      operation 84 markNotificationsSeen

PUT   /api/v1/notifications/read
      operation 85 markAllNotificationsRead

GET   /api/v1/notifications/events
      operation 86 streamNotificationEvents

stable SPA route
      /notifications
```

No operation 87+, route, Permission, semantic owner, lifecycle state, ETag domain or async worker is introduced.

## 2. Root cause

`discussion-notifications-launch.md` already closes Notification persistence, current presentability, paging, counts, engagement and realtime behavior, but deliberately leaves source presentation broad:

```text
read-time presentation may compose currently disclosable source enrichment
such as actor UserReference, DocumentReference and bounded message preview
```

That is sufficient owner architecture, but a full Inbox cannot be implementation-ready while the browser is left to decide which source facts are mandatory or to fan out to Controlled Documents merely to render one row.

Rejected baseline:

```text
Notification row contains only source ids
-> browser fetches Document/message/author context per row
-> browser joins source facts
-> UI becomes a parallel source-recognition graph
```

> A presentable Inbox row must arrive independently human-recognizable from the canonical Notification list read after current source disclosure has already succeeded.

## 3. Target invariant

> Every presentable `DOCUMENT_MENTION` Inbox item is independently recognizable from one server-authored `listNotifications` row and carries the exact stable source identities required to enter the source Discussion context. Human presentation enrichment remains read-time projection only; Notifications persistence continues to own only Notification/source identity plus engagement state.

## 4. Bounded op82 projection

Operation 82 remains:

```text
GET /api/v1/notifications
listNotifications
```

The current item projection is:

```text
NotificationInboxItem {
  notification_id: Uuid
  kind: DOCUMENT_MENTION
  created_at: UtcInstant
  seen_at?: UtcInstant
  read_at?: UtcInstant
  archived_at?: UtcInstant

  source: {
    document: DocumentReference
    message_id: Uuid
    author: UserReference
    official_revision_at_post?: RevisionReference
    message_preview: ShortText
  }
}
```

The exact wire encoding remains owned by the executable wire authority. This decision closes semantic presence/source meaning only.

## 5. Source laws

### 5.1 Document

```text
source.document
  = current-disclosable stable DocumentReference for the persisted source document_id
```

The reference is composed only after current `DocumentDiscussionDisclosure` succeeds for the recipient.

No current source disclosure means no public Inbox row and no Document code/title leak.

### 5.2 Message identity

```text
source.message_id
  = exact immutable DiscussionMessage source identity persisted by Notification
```

This is the navigation/anchor identity. The browser never derives identity from preview text or visual position.

### 5.3 Author

```text
source.author
  stable identity source
    = immutable DiscussionMessage.author_user_id

  human enrichment
    = currently admissible bounded UserReference
```

`display_name` may be absent after lawful profile erasure. Stable `user_id` remains truthful. The Notification row must degrade neutrally rather than fabricate a historical profile snapshot.

### 5.4 Official Revision at post

```text
source.official_revision_at_post
  present iff the immutable DiscussionMessage recorded official_revision_at_post
  resolved from that exact Revision identity
  never replaced by whichever Revision is official now
```

Before first Release it is absent.

This field is contextual provenance only. It does not make Notification a Revision-owned record and is not required for source navigation.

### 5.5 Message preview

```text
source.message_preview
  = bounded server-composed plain-text preview
    derived from immutable MessageContentSegment[]
    after current source disclosure succeeds
```

Rules:

```text
bounded by ShortText (<=256)
may use currently admissible Mention display enrichment
may truncate with an ellipsis
presentation only
not persisted Notification authority
not used for navigation
not used for equality/deduplication
not used for Authorization/presentability
```

Exact typography and truncation mechanics remain presentation detail; the preview may not fabricate semantic content absent from the accepted DiscussionMessage.

## 6. Persistence boundary unchanged

Notifications persistence remains:

```text
notification_id
recipient_user_id
kind
source document_id
source message_id
created_at
seen_at?
read_at?
archived_at?
```

Do not persist copied rendering/source truth merely to satisfy the Inbox:

```text
Document title/code snapshots
DiscussionMessage text/preview
author display name
Revision title
source ACL/presentability truth
```

Controlled Documents remains source-content authority. Organization remains User/profile authority. Authorization remains final ALLOW/default-DENY authority.

## 7. Fixed Inbox views

Operation 82 retains exactly:

```text
view = active | unread | archived
```

After server-side current presentability composition:

```text
active
  archived_at absent

unread
  archived_at absent
  AND read_at absent

archived
  archived_at present
```

Consequences:

```text
unread is a subset of active
there is no unseen list view
mark-unread never makes an item unseen
archive/unarchive never changes read/seen
```

Every first page continues to expose derived current presentable non-archived:

```text
unseen_count
unread_count
```

These are active-Inbox engagement summaries even while the displayed lens is `archived`; they are not generic total-count pagination metadata.

## 8. Source-open boundary

A Notification source-open action uses only identities returned by op82:

```text
source.document.document_id
+ source.message_id
```

Accepted flow:

```text
user activates source
-> op83 read=true
-> admitted success implies seen under existing engagement law
-> navigate to B03 Document Official lens
-> B03 sends exact message_id as anchor intent to op79
-> current source disclosure is rechecked by B03/op79
```

A Notification being marked read never proves that the Document or source message was successfully viewed.

If op83 returns non-disclosing not-found because presentability drifted:

```text
refetch canonical Inbox
item may disappear
no source navigation from stale presentation metadata
no reason for disappearance is leaked
```

## 9. Seen presentation law preserved

There is no manual `seen=false` or "mark unseen" command.

```text
fetch/cache row
  != actual presentation

materially present unseen row to recipient
  -> row may become candidate for bounded op84 batch
```

Operation 84 continues to intersect submitted ids with current recipient + presentability and returns count-free `204` semantics. Exact viewport threshold/debounce/batching implementation remains frontend realization detail.

## 10. Ordering / pagination unchanged

Canonical order remains:

```text
created_at DESC,
notification_id DESC
```

B08/B01N must not globally re-sort loaded cursor pages by Document, author, seen/read or archive state.

First page:

```text
view + optional limit
```

Continuation:

```text
cursor + optional limit only
```

Changing a fixed view starts a new first-page traversal. No total count, page number, generic sort/search/filter DSL or saved view is added.

## 11. Realtime unchanged

Operation 86 remains invalidation only:

```text
event: notifications.changed
data: {}
```

Client behavior:

```text
signal
-> invalidate/refetch Notification summaries/current lens
-> op82 restores canonical truth
```

No business row is inserted/updated/deleted from the SSE payload because no Notification/source business payload is carried.

## 12. B01N / B08 ownership boundary

```text
B01N Quick Inbox
  global glance/attention entry
  bounded recent presentable rows
  active/unread quick lenses
  mark-all-read
  source open
  Ver todas -> /notifications

B08 Full Inbox
  complete persistent triage surface
  active/unread/archived
  cursor continuation
  read/unread
  archive/unarchive
  mark-all-read
  source-open recovery
```

Both consume the same Notifications owner/read family. They never maintain separate engagement stores/counters.

Notifications remains outside `Minha Caixa`; attention is not assigned work.

## 13. Explicit non-goals

This precision does not add:

```text
search
free-form/custom filters
filter by author/Document
saved views
bulk row selection
bulk archive/unarchive
mark-all-unread
mark-unseen
snooze
priority/reminders
Notification preferences
email/push settings
Notification deletion
Notification-kind selector
source reply/editor/viewer inside Inbox
```

Current Notification kind remains only `DOCUMENT_MENTION`.

## 14. Census impact

```text
new application operations       0
new stable SPA routes            0
new PermissionCode values        0
new semantic owners              0
new lifecycle states             0
new ETag domains                 0
new Idempotency-Key creations    0
new exact-byte resources         0
new async workers                0
```

Current accepted census remains:

```text
86 operations / 11 routes / 16 PermissionCode values
```

## 15. Proof obligations

Later executable-contract/frontend/implementation proof must falsify at least:

```text
presentable Inbox row requires per-row browser fan-out for Document/message/author recognition
source presentation leaks before current source disclosure succeeds
browser hides non-presentable Notification after receiving leaked source data
message_preview is persisted as Notification authority
message_preview is used as navigation identity
current official Revision silently replaces official_revision_at_post
profile erasure rewrites stable author identity
unread lens includes archived rows
archived lens changes read/seen semantics by virtue of listing
loaded but never presented rows are blindly marked seen
client globally re-sorts cursor pages
SSE payload becomes Notification/source business truth
Quick Inbox and Full Inbox diverge into separate engagement authorities
Notification source-open bypasses B03/op79 disclosure recheck
```

## 16. Reopen triggers

Reopen only the implicated decision if material evidence proves a need for:

```text
multiple Notification kinds requiring a different closed source union
source recognition fields insufficient for real Inbox triage
server-composed preview unsustainable at measured scale
Notification search/filter requirements beyond the three fixed views
bulk triage beyond current mark-all-read
snooze/priority/preferences/delivery configuration
external email/push/webhook delivery as current Product scope
Notification semantic Audit requirement
```

None is implied by B08-F1.
