# T11 — B08 Notifications Full Inbox R1 — Method v2.2 ratified

> **Status:** OPEN / ACTIVE / B08-F1 OPERATOR-RATIFIED / P6 COMPLETE / P7 H1 OPERATOR-RATIFIED / P8 OPEN.  
> **Block:** B08 — Notifications Full Inbox.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 / B06 / B07 LOCKED.  
> **Bounded authorities:** `../../decisions/discussion-notifications-launch.md` + `../../decisions/notification-inbox-recognition-read.md`.  
> **Implementation:** BLOCKED.  
> **P8:** OPEN / functional low-fidelity HTML is the next evidence gate.

## 1. Ratification basis

The operator approved in sequence:

```text
B08 entry direction / continue FP1
B08-F1 + P7 H1 design in chat
written B08 candidate exactly as recorded
```

The written candidate self-review had:

```text
TBD / TODO / undecided semantic field = 0
new operation = 0
new route = 0
new Permission = 0
new semantic owner = 0
```

B08-F1 is now promoted to durable authority:

```text
../../decisions/notification-inbox-recognition-read.md
```

## 2. Product/owner boundary

B08 is the full persistent Notification Inbox already admitted by Launch.

```text
stable route
  /notifications

reads
  82 listNotifications
  86 streamNotificationEvents

writes
  83 updateNotificationEngagement
  84 markNotificationsSeen
  85 markAllNotificationsRead

semantic owner
  Notifications

source owner
  Controlled Documents for DOCUMENT_MENTION context
```

B01N remains the global bell + Quick Inbox.

Mental model:

```text
Notifications = personal attention / triage
Minha Caixa   = assigned authoring/governance work
```

Quick Inbox and B08 consume one owner/read family and never become separate engagement stores or counters.

## 3. B08-F1 — CLOSED / OPERATOR-RATIFIED

Every currently presentable `DOCUMENT_MENTION` row returned by op82 is independently human-recognizable after server-side current disclosure.

Current semantic item projection:

```text
NotificationInboxItem {
  notification_id
  kind=DOCUMENT_MENTION
  created_at
  seen_at?
  read_at?
  archived_at?

  source {
    document: DocumentReference
    message_id
    author: UserReference
    official_revision_at_post?: RevisionReference
    message_preview: ShortText
  }
}
```

Source laws:

```text
document
  current-disclosable stable DocumentReference

message_id
  exact immutable DiscussionMessage identity
  navigation authority for B03 anchor

author
  immutable author_user_id + current admissible bounded UserReference enrichment
  display_name may be absent after lawful profile erasure

official_revision_at_post
  exact Revision recorded by the immutable DiscussionMessage when one existed
  never silently replaced by whichever Revision is official now

message_preview
  server-composed after disclosure from immutable MessageContentSegment[]
  <= ShortText
  presentation only
  not persistence / navigation / dedupe / AuthZ authority
```

Notifications persistence remains only:

```text
Notification identity
recipient
kind
source document_id + message_id
created_at
seen_at?
read_at?
archived_at?
```

No source message/title/profile/ACL copy is promoted.

Current census remains:

```text
86 operations / 11 routes / 16 PermissionCode values
```

## 4. Engagement law preserved

```text
READ => SEEN
mark unread clears read_at only
mark unread never makes unseen
archive/unarchive preserves read/seen
per-item deliberate engagement implies seen if absent
archive != delete
read != Document read
read != Read & Acknowledge
Notification != access grant
```

Fixed op82 views after presentability:

```text
active
  archived_at absent

unread
  archived_at absent
  AND read_at absent

archived
  archived_at present
```

`unread` is a subset of active. There is no unseen list lens.

First-page summaries remain presentable non-archived:

```text
unseen_count
unread_count
```

## 5. P6 — COMPLETE

Reference study saturated after GitHub Notifications, Linear Inbox and Slack Activity.

Accepted lessons:

```text
Inbox = attention triage, not assigned-work authority
read/unread is distinct from removing/archiving active attention
source work remains in the source owner lens
compact recognition + source entry beats rebuilding source work inside Inbox
```

Rejected breadth:

```text
search/filter platform
saved/custom views
snooze/reminders/priority
preferences/subscriptions
arbitrary bulk selection
```

## 6. P7 — H1 Focused Triage Inbox — OPERATOR-RATIFIED

Selected structure:

```text
minimal Full Inbox header
  Notificações
  unseen/unread summary
  Marcar todas como lidas

three fixed lenses
  Caixa de entrada = active
  Não lidas        = unread
  Arquivadas       = archived

single dominant list
  Notification recognition
  unseen novelty distinct from unread
  bounded source preview
  exact created time
  primary source-open affordance
  per-item read/unread
  per-item archive/unarchive

continuation
  cursor + retry preserving loaded items
```

Rejected as leading:

```text
master/detail Inbox that rebuilds B03
Document-grouped Inbox that replaces canonical recency order
```

No B08 viewer, Discussion composer or governance action zone exists.

## 7. Source-open interaction

Exact identity:

```text
source.document.document_id
+ source.message_id
```

Flow:

```text
activate source
-> op83 read=true
-> on admitted success navigate to B03 Document Official
-> B03 uses exact source.message_id as Discussion anchor intent
-> op79 anchored read
-> B03/current disclosure rechecked
```

If op83 returns non-disclosing 404 because presentability drifted:

```text
refetch Inbox
item may disappear
no navigation from stale source presentation
no disclosure reason leaked
```

If B03 becomes unavailable after engagement but before destination read, B03 owns its own neutral unavailable state. Notification read never proves source view.

## 8. Seen presentation

There is no manual mark-seen/unseen command.

```text
fetch/cache row != present row

actually present unseen row to recipient
-> candidate id for bounded op84 batch
-> server intersects with current recipient + presentability
-> 204 exposes no per-id cardinality
```

P8 must visibly prove that loaded-but-off-screen rows are not blindly marked seen merely because they were fetched.

Exact viewport threshold/debounce/batching mechanism is not Product authority.

## 9. Ordering / pagination

Server order remains:

```text
created_at DESC,
notification_id DESC
```

B08 never globally re-sorts cursor pages.

```text
first page
  view + optional limit

continuation
  cursor + optional limit only

change lens
  fresh first-page traversal
```

No total count, page number, generic sort, search or saved filter.

## 10. Realtime

Operation 86 remains invalidation only:

```text
event: notifications.changed
data: {}
```

```text
signal
-> invalidate/refetch current Notification summaries/lens
-> op82 restores canonical truth
```

SSE never carries a Notification/source business row. Disconnect is freshness degradation, not Product-state loss.

## 11. Failure / reconciliation targets

```text
initial op82 failure
  Inbox unavailable + retry; shell stays available

continuation failure
  preserve loaded rows + current lens + retry cursor

op83 404/non-presentable
  refetch; row may disappear; neutral explanation

op83 ambiguous/transport failure
  do not fabricate changed engagement state

op84 failure
  novelty stays conservative until refetch

op85 failure
  do not locally declare all read

SSE disconnect
  reconnect/focus/refetch restores truth
```

## 12. P8 region inventory

```text
R1  Global App Shell inherited from B01
R2  Notification bell / Quick Inbox inherited from B01N
R3  Full Inbox heading
R4  unseen + unread summary
R5  Mark all read
R6  active / unread / archived lenses
R7  human-recognizable Notification row
R8  unseen novelty distinct from unread treatment
R9  read / unread
R10 archive / unarchive
R11 exact source open to B03 Discussion anchor
R12 cursor continuation
R13 continuation failure/retry preserving rows
R14 per-item engagement failure/reconciliation
R15 mark-all-read failure/reconciliation
R16 access-drift disappearance after neutral 404/refetch
R17 empty active
R18 empty unread
R19 empty archived
R20 SSE invalidation/refetch
R21 narrow/mobile row/action treatment
R22 accessibility/focus/live reconciliation
```

## 13. P8 R1 proof targets

Functional HTML must exercise at least:

```text
active / unread / archived fixed lenses
unseen + unread item
seen + intentionally unread item
read item
archive preserving read state
archived unread item + unarchive
mark-all-read affects active unread only
loaded-but-not-presented rows not blindly seen
bounded seen batching evidence
exact source-open boundary to B03 anchor
source engagement 404/access-drift disappearance
cursor continuation + failure/retry
mark-all-read failure
per-item engagement failure
empty states for all three lenses
SSE invalidation/refetch
Quick Inbox coherence with Full Inbox
responsive row/actions
keyboard/focus/live-region recovery
```

## 14. Explicit non-goals

```text
search
free-form/custom filters
filter by author/Document
saved views
bulk selection/archive
mark-all-unread / mark-unseen
snooze / priority / reminders
Notification preferences
email/push settings
Notification deletion
Notification-kind selector
source reply/editor/viewer inside Inbox
```

## 15. Current gate

```text
entry recovery                         COMPLETE
P6                                     COMPLETE
B08-F1                                 CLOSED / OPERATOR-RATIFIED
P7 H1 Focused Triage Inbox             OPERATOR-RATIFIED
P8                                     OPEN / NEXT
B08 LOCK / P9 / P10                    NOT OPEN
B09+                                   NOT OPEN
```

Next action:

> Write the temporary P8 realization plan, create the functional low-fidelity B08 HTML with deterministic local fixtures, verify structural behavior, deliver the exact HTML to the operator, and remain in B08 until explicit LOCK.
