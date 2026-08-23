# T11 — B08 Notifications Full Inbox R1 — Method v2.2 candidate

> **Status:** OPEN / ACTIVE / ENTRY RECOVERY COMPLETE / P6 COMPLETE / B08-F1 + P7 H1 DESIGN APPROVED IN CHAT / WRITTEN OPERATOR RATIFICATION PENDING.  
> **Block:** B08 — Notifications Full Inbox.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 / B06 / B07 LOCKED.  
> **Current bounded authority:** `../../decisions/discussion-notifications-launch.md`.  
> **Implementation:** BLOCKED.  
> **P8:** NOT OPEN until written design ratification.

## 1. Entry recovery

B08 is the full persistent Notification Inbox already admitted by the current Launch Notification authority. It does not create a new Product space, semantic owner or task system.

```text
stable route
  /notifications

primary reads
  82 listNotifications
  86 streamNotificationEvents

writes
  83 updateNotificationEngagement
  84 markNotificationsSeen
  85 markAllNotificationsRead

semantic owner
  Notifications

source owner
  Controlled Documents for current DOCUMENT_MENTION source context

sidebar
  no permanent Notifications item

global entry
  B01N Notification bell + Quick Inbox
```

Mental-model boundary remains:

```text
Notifications = personal attention / triage
Minha Caixa   = assigned authoring/governance work
```

Quick Inbox and B08 are two presentations of the same Notifications owner/read family. They are never separate stores or engagement authorities.

## 2. Current Notification law preserved

Persistent engagement state:

```text
created_at
seen_at?
read_at?
archived_at?
```

Meaning:

```text
unseen -> seen_at absent
seen   -> seen_at present
unread -> read_at absent
read   -> read_at present
archive -> archived_at present
```

Binding laws:

```text
READ => SEEN
mark unread clears read_at only; it never makes an item unseen
archive/unarchive preserves read/seen
any deliberate per-item engagement implies seen if absent
archive != delete
read != source acknowledgement
read != Document read
read != Read & Acknowledge
Notification != access grant
Notification != assigned work
```

Presentability remains server-side and current:

```text
recipient == current User
+ current User ENABLED
+ current Document Discussion source disclosable
```

A non-presentable Notification is omitted before public paging/counts/source composition. React never receives prohibited source metadata and never post-filters Notification disclosure.

## 3. Human jobs

### J1 — see what is new versus merely unread

```text
When I open Notifications,
I need to distinguish genuinely new/unseen attention from older items I intentionally marked unread,
so that novelty does not become another name for unread.
```

### J2 — triage attention without losing history

```text
When an item no longer needs active attention,
I need to archive it without deleting it or rewriting its read/seen state,
and later recover it from Archived when needed.
```

### J3 — return to the exact source context

```text
When a Notification came from an accepted @Mention,
I need to recognize who mentioned me and in which controlled Document,
then enter the exact source Discussion message rather than a generic Document home.
```

### J4 — handle long-lived Inbox truth

```text
When the Inbox spans many items or changes in another tab,
I need cursor continuation and realtime refresh without client-side reordering, copied counters or SSE payload authority.
```

### J5 — remain safe under access drift

```text
When source disclosure disappears between list rendering and engagement/navigation,
I need the Inbox to reconcile neutrally without leaking why the source vanished.
```

## 4. P6 bounded reference study — COMPLETE

### GitHub Notifications Inbox

Useful evidence:

```text
read/unread is distinct from removing an item from the active Inbox
Done-style triage preserves a separate historical surface
single-item and bounded bulk triage are recognizable Inbox mechanics
```

Accepted lesson:

> Notification engagement and Inbox membership are separate concepts. MetalDocs `read_at` and `archived_at` should stay separate.

Rejected breadth:

```text
rich reason/repository filters
saved/custom filters
arbitrary bulk selection platform
subscription/watch administration
```

### Linear Inbox

Useful evidence:

```text
Inbox is an attention center, not the user's assigned-work database
read/unread and removal/triage are separate interactions
source object remains the place where business work occurs
```

Rejected breadth:

```text
snooze
reminders
notification preferences
search/filter platform
```

### Slack Activity

Useful evidence:

```text
attention surface leads back to source context
compact scanning is more useful than rebuilding the source workspace inside the Inbox
```

Rejected breadth:

```text
DM/thread/activity multiplexing
channel-specific filtering
reminder/task semantics
```

P6 saturated because additional products stopped changing the bounded MetalDocs decision space.

## 5. B08-F1 — human-recognizable Notification read projection — WRITTEN CANDIDATE

### 5.1 Root cause

Current Notification authority closes persistence, presentability, paging, counts and engagement but deliberately leaves source presentation broad:

```text
read-time presentation may compose currently disclosable source enrichment
such as actor UserReference, DocumentReference and bounded message preview
```

That is sufficient architecture posture but not sufficiently closed for a full Inbox Screen Contract. B08 must not invent actor names, Document titles or message previews from client-side fan-out or copied Notification persistence.

### 5.2 Target invariant

> Every presentable DOCUMENT_MENTION Inbox item is independently human-recognizable from one server-authored op82 row after current source disclosure. The row carries enough bounded source context to identify the mention and navigate to the exact Discussion message, while Notifications persistence continues to own only Notification/source identities and engagement state.

### 5.3 Bounded op82 projection refinement

No new operation is created. Operation 82 remains:

```text
GET /api/v1/notifications
listNotifications
```

Conceptual current item shape becomes:

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

This is a read projection, not a new semantic owner or persistence shape.

### 5.4 Source law

```text
document
  = current-disclosable stable DocumentReference for source document_id

author
  = immutable DiscussionMessage.author_user_id
    enriched to currently admissible bounded UserReference
    display_name may be absent after lawful profile erasure

message_id
  = exact immutable DiscussionMessage source identity

official_revision_at_post
  = present only when the immutable DiscussionMessage recorded an official Revision identity at post time
  = server resolves that exact Revision to RevisionReference
  = absent before first Release
  = never replaced by whichever Revision is official now

message_preview
  = server-composed bounded plain-text preview derived from the immutable MessageContentSegment sequence
  = generated only after current disclosure succeeds
  = may use current admissible Mention display enrichment
  = bounded by ShortText (<=256)
  = presentation only; never persisted Notification authority
  = never used for navigation, equality, deduplication or Authorization
```

If a current display name is unavailable, presentation must degrade neutrally while stable source/User identities remain truthful. The preview may truncate with an ellipsis; exact typography/truncation mechanism is presentation detail so long as it does not fabricate semantic content.

### 5.5 Persistence boundary unchanged

Notifications persistence remains identity + engagement only:

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

Do not copy into Notifications persistence merely for rendering:

```text
Document title/code snapshots
message text/preview
author display name
Revision title
source ACL/presentability truth
```

### 5.6 View semantics closed

First-page filter remains exactly:

```text
view = active | unread | archived
```

Meaning after current presentability filtering:

```text
active
  archived_at absent

unread
  archived_at absent
  AND read_at absent

archived
  archived_at present
```

`unread` is a subset of active. There is no `unseen` list view.

Every first page continues to expose derived current presentable non-archived:

```text
unseen_count
unread_count
```

These summaries describe the active Inbox population even when the current displayed lens is `archived`; they are not total-count pagination metadata.

### 5.7 Census impact

```text
new application operation       0
new stable SPA route            0
new Permission                  0
new semantic owner              0
new lifecycle state             0
new ETag domain                 0
new Idempotency-Key creation    0
new exact-byte resource         0
new async worker                0
```

Current accepted census remains 86 operations / 11 routes / 16 PermissionCode values.

## 6. P7 approaches

### H1 — Focused Triage Inbox — SELECTED / DESIGN APPROVED IN CHAT

```text
minimal Inbox header
  Notificações
  unseen/unread summary
  Mark all read

three fixed lenses
  Caixa de entrada = active
  Não lidas        = unread
  Arquivadas       = archived

single dominant list
  Notification recognition
  novelty/read state
  bounded source preview
  exact created time
  source-open affordance
  per-item engagement actions

continuation
  Load more / cursor failure recovery
```

Why selected:

```text
matches the owner wire directly
keeps attention distinct from assigned work
keeps source business work in B03
preserves one global chronological Notification order
supports reversible read/archive engagement without a filter platform
scales to long Inbox via current cursor law
```

### H2 — two-column master/detail Inbox — REJECTED AS LEADING

```text
notification list | source preview/detail pane
```

Rejected because the detail pane quickly reconstructs Discussion/Document Official semantics inside Notifications and creates pressure for reply/viewer/source business actions in the wrong owner lens.

### H3 — grouped-by-Document Inbox — REJECTED AS LEADING

Rejected because it replaces the canonical Notification order:

```text
created_at DESC,
notification_id DESC
```

with a browser-authored aggregation axis, making recency harder to scan and encouraging a second source relationship model in the client.

## 7. Selected H1 region model

```text
R1  global App Shell inherited from B01
R2  Notification bell / Quick Inbox inherited from B01N
R3  Full Inbox heading
R4  unseen + unread summary
R5  Mark all read
R6  active / unread / archived fixed lenses
R7  Notification row human recognition
R8  unseen novelty marker distinct from unread treatment
R9  read / unread action
R10 archive / unarchive action
R11 exact source open to B03 Discussion anchor
R12 cursor continuation
R13 continuation failure preserving loaded items
R14 per-item engagement failure/reconciliation
R15 mark-all-read failure/reconciliation
R16 access-drift item disappearance after neutral 404/refetch
R17 empty active state
R18 empty unread state
R19 empty archived state
R20 SSE invalidation/refetch state
R21 responsive narrow/mobile list/action treatment
R22 accessibility/focus/live reconciliation
```

No B08 Document viewer, Discussion composer or Governance action zone exists.

## 8. Interaction law

### 8.1 Open source

Deliberate source activation targets the exact stable Document + immutable message identity returned by op82:

```text
NotificationInboxItem.source.document.document_id
+ source.message_id
```

Sequence:

```text
user activates source
-> op83 set read=true (therefore seen as required by owner law)
-> on admitted success, navigate to B03 Document Official
-> B03 uses the exact message_id as Discussion anchor intent
-> op79 listDocumentDiscussionMessages(anchor_message_id=...)
-> current B03 disclosure rechecked
```

Exact router/query-state encoding of the anchor is not a new stable Product route and remains implementation choice. The identity reaching op79 must be the exact server-returned `message_id`, never parsed preview text.

If op83 returns non-disclosing 404 because presentability drifted:

```text
refetch current Inbox
item may disappear
no source navigation from stale leaked metadata
no explanation of whether source was deleted/foreign/inaccessible
```

If B03 source disclosure changes after successful engagement but before destination read, B03 owns its normal unavailable/non-disclosing state. A Notification being read does not prove the source was viewed.

### 8.2 Read / unread

```text
unread -> op83 read=true
read   -> op83 read=false
```

Mark unread preserves seen. No action may set `seen=false`.

### 8.3 Archive / unarchive

```text
active item   -> op83 archived=true
archived item -> op83 archived=false
```

Archive/unarchive preserves read and seen. Archive is not delete.

### 8.4 Mark all read

Header command maps only to op85:

```text
presentable + non-archived + unread at operation point
-> read + seen
```

It does not include archived or currently non-presentable Notifications and does not create a mark-all-unread opposite command.

### 8.5 Seen presentation

B08 exposes no manual "mark seen" action.

Only Notifications that are actually presented to the user may become seen candidates. Fetching/caching a cursor page does not by itself prove presentation.

Conceptual behavior:

```text
loaded row becomes materially presented in the Inbox viewport
-> collect its notification_id when still unseen
-> bounded batch op84
-> server intersects with current recipient + current presentability
-> 204 reveals no per-id cardinality
```

Exact viewport threshold, debounce and batching implementation are not Product authority. P8 must prove at least that hidden/off-screen loaded rows are not blindly marked seen merely because a page was fetched.

## 9. Ordering / pagination law

Operation 82 keeps server order:

```text
created_at DESC,
notification_id DESC
```

B08 never re-sorts loaded cursor pages by Document, author, seen/read or archive state.

First page:

```text
view + optional limit
```

Continuation:

```text
cursor + optional limit only
```

Changing a lens starts a fresh first-page traversal. Already loaded items remain readable if continuation fails; retry continues the same lens traversal.

No total count, page number, generic sort, search or saved filter is introduced.

## 10. Realtime law

Operation 86 remains invalidation only:

```text
event: notifications.changed
data: {}
```

B08 behavior:

```text
signal
-> invalidate/refetch current Notification summaries/lens
-> canonical op82 restores truth
```

No Notification row is inserted, updated or removed from SSE payload data because there is no business payload.

Stream disconnect/failure is a freshness degradation, not Product-state failure. Reconnect/focus/refetch restores canonical truth.

## 11. Failure / reconciliation law

```text
initial op82 failure
  -> Inbox unavailable state + retry; shell remains truthful

continuation failure
  -> preserve already loaded items + current lens + retry continuation

op83 404/non-presentable
  -> refetch; item may disappear; no disclosure explanation

op83 generic/transport failure
  -> preserve visible current server state; action remains retryable

op84 failure
  -> novelty remains conservative until canonical refetch succeeds

op85 failure
  -> do not locally declare all read; preserve current rows/counts + retry/refetch

SSE disconnect
  -> no Product truth lost; reconnect/focus/refetch path
```

Frontend does not fabricate engagement truth optimistically when the mutation outcome is unknown.

## 12. Responsive / accessibility direction

Desktop:

```text
one dominant Inbox column
fixed lens controls near heading
row source/content recognition dominates; secondary engagement actions trail
Quick Inbox remains header utility and does not duplicate full-page controls while page is open
```

Narrow/mobile:

```text
heading + counts
fixed lenses remain touch-operable
one-column Notification rows
secondary actions may move to bounded row menu/sheet
source-open remains a clear primary row action
```

Accessibility obligations:

```text
unseen novelty is announced in text/semantics, not color alone
unread/read is distinguishable independently of unseen
row action labels identify the specific action
source-open target has a clear accessible name
lens selection uses real tab/navigation semantics appropriate to implementation
focus is preserved after engagement removes a row from the current lens
load-more and realtime reconciliation announce meaningful list changes without stealing focus
```

## 13. Explicit non-goals / YAGNI

B08 current Launch does not add:

```text
search
free-form filters
filter by author or Document
custom/saved views
bulk row selection
bulk archive/unarchive
mark-all-unread
mark-unseen
snooze
priority
reminders
Notification preferences
email/push settings
Notification deletion
Notification-kind selector
source reply/editor/viewer inside Inbox
```

Current Notification kind remains only `DOCUMENT_MENTION`. A richer Notification taxonomy is a separate Product reopen, not a reason to pre-build generic Inbox infrastructure.

## 14. P8 proof targets after written ratification

P8 R1 must exercise at least:

```text
active / unread / archived fixed lenses
unseen item that is also unread
seen item intentionally marked unread
read item
archive preserving read state
archived unread item + unarchive
mark-all-read affecting active unread only
loaded-but-not-presented rows not blindly marked seen
bounded seen batching fixture
exact source open to B03 anchor
source engagement 404/access-drift disappearance
cursor continuation + failure/retry
mark-all-read failure
per-item engagement failure
empty states for all three lenses
SSE invalidation/refetch
Quick Inbox state coherence with full Inbox
responsive row/action treatment
keyboard/focus/live-region recovery
```

## 15. Self-review

Placeholder scan:

```text
TBD / TODO / undecided semantic field = 0
```

Consistency check:

```text
B01N remains global entry / glance surface
B08 remains full triage surface
op82 remains sole Inbox list authority
op83–85 remain sole engagement writes
op86 remains invalidation only
B03 remains exact source Discussion owner lens
Notifications persistence remains identity + engagement only
```

Scope check:

```text
no second Notification capability introduced
no generic filtering/preferences/delivery platform
no new API operation or route
```

Ambiguity check closed by this candidate:

```text
exact human-recognition source projection
active/unread/archived meanings
source-open engagement/navigation boundary
seen presentation rule
SSE invalidation behavior
```

## 16. Current gate

```text
B08 entry recovery                         COMPLETE
P6                                         COMPLETE
B08-F1 recognition projection              DESIGN APPROVED IN CHAT / WRITTEN RATIFICATION PENDING
P7 H1 Focused Triage Inbox                  DESIGN APPROVED IN CHAT / WRITTEN RATIFICATION PENDING
P8                                         NOT OPEN
B08 LOCK / P9 / P10                        NOT OPEN
B09+                                       NOT OPEN
```

Next gate:

> Operator reviews and ratifies this written B08 design. Only after written ratification may B08-F1 be promoted to durable bounded authority and P8 functional HTML open.
