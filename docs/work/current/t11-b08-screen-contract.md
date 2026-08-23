# T11 — B08 P9 Screen Contract / Bidirectional Trace

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B08 — Notifications Full Inbox.  
> **Depends on:** B08 operator LOCK, `notification-inbox-recognition-read.md`, `discussion-notifications-launch.md`, B01N global Notification chrome and B03 Discussion anchor contract.  
> **Implementation:** BLOCKED.

## 1. Goal

Prove that every material region/control in the locked B08 functional wireframe is realizable by current authority without inventing Notification truth, frontend disclosure filtering, a second Inbox store, a source workspace or a screen-shaped API.

## 2. Screen contract

| Surface / region | User goal | Current read truth | Material control / write | Identity source | Material failure / safe UX | Forbidden frontend authority | Status |
|---|---|---|---|---|---|---|---|
| B08-01 stable route | open the full personal attention Inbox | op82 `listNotifications` | route only | authenticated current User | initial failure -> Inbox unavailable + retry | no `Minha Caixa` fallback | READY |
| B08-02 global shell | preserve normal application orientation | B01 Global App Shell | normal global navigation | existing shell state | shell remains usable when Inbox fails | no Notification-specific primary IA | READY |
| B08-03 global bell / Quick Inbox | keep global attention entry coherent while full Inbox is open | B01N locked Notification chrome + same Notifications owner state | bell open/close only | current derived unseen count | Quick Inbox failure never replaces full Inbox truth | no second Notification store | READY |
| B08-04 Inbox heading | identify this as Notifications attention/triage | op82 current owner read family | none | stable `/notifications` route | unavailable state remains explicit | no task/work semantics | READY |
| B08-05 unseen/unread summaries | distinguish novelty from read state | first-page op82 `unseen_count` + `unread_count` after current presentability | none | server-derived summaries | mutation ambiguity keeps prior truth until refetch | no durable copied counters | READY |
| B08-06 mark all read | clear all currently presentable active unread attention | op82 summaries + current list | op85 `markAllNotificationsRead` | current recipient implicitly | failure -> do not declare local all-read; retry/refetch | no archived/non-presentable sweep | READY |
| B08-07 active lens | inspect current non-archived attention | op82 `view=active` | lens navigation/read only | server filter | fresh traversal on lens change | no client post-filter | READY |
| B08-08 unread lens | inspect active unread subset | op82 `view=unread` | lens navigation/read only | server filter | fresh traversal on lens change | unread != unseen | READY |
| B08-09 archived lens | inspect archived attention history | op82 `view=archived` | lens navigation/read only | server filter | first-page summaries still describe active population | archive != delete | READY |
| B08-10 Notification recognition row | recognize who mentioned me, where and when | op82 `NotificationInboxItem` + B08-F1 source projection | source-open + engagement actions | returned notification/document/message/author/revision-at-post identities | missing source projection is server-contract failure | no client source fan-out | READY |
| B08-11 unseen novelty marker | know an item is genuinely new | op82 `seen_at?` | none | exact engagement field | presentation remains conservative on op84 failure | no color-only novelty / no unread inference | READY |
| B08-12 unread/read treatment | know whether attention is currently unread | op82 `read_at?` | op83 `read=true/false` | returned `notification_id` | failure preserves visible authoritative state + retry/refetch | no mark-unseen action | READY |
| B08-13 archive/unarchive | remove/restore item from active Inbox without deleting it | op82 `archived_at?` | op83 `archived=true/false` | returned `notification_id` | failure preserves current lens truth | no delete or source mutation | READY |
| B08-14 exact source activation | continue to exact Mention context | op82 source `document.document_id + message_id` | op83 `read=true`, then navigate to B03 anchor boundary | exact returned source ids | op83 404/access drift -> refetch, item may disappear, no navigation | preview text never navigation authority | READY |
| B08-15 B03 Discussion destination | reveal exact source message after handoff | B03 op79 `listDocumentDiscussionMessages(anchor_message_id=...)` | destination read only by B08 | exact message id from op82 | B03 rechecks current disclosure and owns unavailable state | Notification never grants source access | READY |
| B08-16 presentation-driven seen | mark only actually presented unseen items as seen | op82 `seen_at?` | op84 `markNotificationsSeen(ids<=100)` | ids of rows materially presented | failure leaves novelty conservative; later refetch reconciles | fetch/cache != presentation | READY |
| B08-17 cursor continuation | continue a long lens in canonical order | op82 cursor page | load more | opaque cursor | current presentability/AuthZ rechecked by server | no total/page-number/global resort | READY |
| B08-18 continuation failure | recover without losing loaded attention | retained op82 pages + failed continuation | retry same traversal | current opaque cursor | loaded rows remain readable; focus/position preserved | no fabricated completed page | READY |
| B08-19 per-item engagement failure | understand an action was not accepted | op83 error + current cached row | retry/refetch | notification id | 404 -> neutral disappearance reconciliation; generic failure -> retain row | no optimistic business truth on ambiguity | READY |
| B08-20 mark-all failure | understand bulk command did not complete | op85 error + current op82 state | retry/refetch | current recipient | counts/rows not locally rewritten as success | no client sweep simulation | READY |
| B08-21 SSE invalidation / disconnect | regain freshness across tabs without making stream authoritative | op86 `notifications.changed {}` + canonical op82 refetch | reconnect/refetch only | current recipient stream | disconnect is freshness degradation only | no row payload/state from SSE | READY |
| B08-22 empty/responsive/accessibility states | keep all three lenses and recovery operable across viewport/input modes | same op82/op83–85 truths | presentation/focus only | none new | empty active/unread/archived distinguishable; focus preserved after row removal | no mobile-specific business semantics | READY |

## 3. Exact operation homes used by B08

```text
82  listNotifications
83  updateNotificationEngagement
84  markNotificationsSeen
85  markAllNotificationsRead
86  streamNotificationEvents
```

B08 source continuation terminates at the already-locked B03 boundary:

```text
79  listDocumentDiscussionMessages(anchor_message_id=source.message_id)
```

Operation 79 remains owned by B03 / Controlled Documents. B08 does not call it to enrich Inbox rows.

No operation 87+ is needed.

## 4. Bidirectional trace

### Product/backend -> frontend

```text
op82 presentable Notification page + B08-F1 projection
-> summaries + fixed lenses + human-recognizable rows

seen_at / read_at / archived_at
-> independent novelty/read/archive presentation

source.document.document_id + source.message_id
-> exact B03 Discussion handoff

op86 invalidation only
-> refetch op82
```

### Frontend -> Product/backend

```text
switch lens
-> new first-page op82 with exact view

mark read/unread or archive/unarchive
-> op83

actually present unseen rows
-> bounded op84 ids

mark all read
-> op85

SSE signal / reconnect / focus
-> canonical op82 refetch

open source
-> op83 read=true
-> B03 route
-> op79 anchor read owned by B03
```

## 5. Client state classes

B08 uses only accepted frontend state classes:

```text
SERVER STATE
  Notification pages, unseen/unread summaries, engagement fields, current source projection

NAVIGATION / URL
  /notifications + fixed view intent + B03 source destination intent

FORM DRAFT
  none

EPHEMERAL UI
  Quick Inbox open/closed, row action menu, retry banners,
  observed/presented-row batching buffer, focus restoration, demo/reconciliation state
```

The presented-row buffer is transient interaction bookkeeping only. It is not a durable seen authority; op84 remains canonical mutation authority.

## 6. Wire / ordering / concurrency mechanics

```text
op82 first page
  view + optional limit

op82 continuation
  cursor + optional limit only

canonical order
  created_at DESC,
  notification_id DESC

op83
  read?:boolean and/or archived?:boolean
  no ETag domain
  server rechecks recipient + current presentability

op84
  unique ids <=100
  count-free 204
  server intersects current recipient + current presentability

op85
  current presentable + non-archived + unread set at operation point
  count-free 204

op86
  text/event-stream
  notifications.changed / {}
  invalidation only
```

No generic mutation retry or optimistic lifecycle fabrication is introduced.

## 7. Material failure intent

```text
initial list failure
  Inbox is temporarily unavailable; shell/other Product spaces remain usable; retry

continuation failure
  preserve loaded rows, lens and reading position; retry continuation

op83 non-disclosing 404
  refetch; stale row may disappear; do not explain source access state

op83 transport/generic failure
  keep current row state; action may retry/refetch

op84 failure
  keep novelty conservative until canonical refetch succeeds

op85 failure
  never locally announce all-read success; preserve current rows/counts

SSE disconnect
  no Product state lost; reconnect/focus/refetch restores freshness

B03 destination unavailable
  B03 owns neutral source-unavailable state after current disclosure recheck
```

## 8. Access / disclosure proof

```text
Notification presence              != source access grant
message_preview                    != source identity
read                               != Document read / acknowledgement
archive                            != delete
unread                             != unseen
loaded                             != seen
Quick Inbox state                  != second Notification store
SSE event                          != Notification truth
client visibility/filtering        != presentability authority
```

Public op82 paging/counts/source projection are formed only after server-side current presentability composition.

## 9. B08 negative contract

The locked B08 contains no current Product surface for:

```text
search
free-form/saved filters
filter by author/Document
bulk row selection
bulk archive/unarchive
mark all unread
mark unseen
snooze / priority / reminders
Notification preferences
email/push settings
Notification delete
Notification-kind selector
Discussion reply/composer inside Inbox
Document viewer inside Inbox
frontend Authorization/presentability matrix
```

## 10. P9 closure

```text
material B08 regions/controls traced         22 / 22
unbound material controls                    0
invented operations                          0
operation 87+                                absent
screen-shaped APIs                           0
frontend presentability/AuthZ evaluator      0
second Notification state store              0
source-workspace duplication                 0
SSE business-payload authority               0
material B08 Screen Contract findings        0
```

P9 is complete for the operator-locked B08 scope.
