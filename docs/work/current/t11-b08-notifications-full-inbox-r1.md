# T11 — B08 Notifications Full Inbox R1 — Method v2.2

> **Status:** OPEN / ACTIVE / B08-F1 OPERATOR-RATIFIED / P6 COMPLETE / P7 H1 OPERATOR-RATIFIED / P8 R1 READY FOR OPERATOR USE.  
> **Block:** B08 — Notifications Full Inbox.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 / B06 / B07 LOCKED.  
> **Bounded authorities:** `../../decisions/discussion-notifications-launch.md` + `../../decisions/notification-inbox-recognition-read.md`.  
> **Canonical P8 R1:** `t11-b08-notifications-full-inbox-functional-wireframe.html`.  
> **Canonical HTML blob:** `bb130535721b2381524763a4885ade5199a15596`.  
> **Implementation:** BLOCKED.  
> **LOCK:** NOT YET — operator must operate/iterate P8 first.

## 1. Ratified basis

The operator approved:

```text
B08 entry / continue FP1
B08-F1 + P7 H1 design in chat
written B08 candidate exactly as recorded
```

B08-F1 durable authority:

```text
../../decisions/notification-inbox-recognition-read.md
```

Current census remains:

```text
86 operations / 11 routes / 16 PermissionCode values
```

## 2. Product boundary

```text
/notifications

82 listNotifications
83 updateNotificationEngagement
84 markNotificationsSeen
85 markAllNotificationsRead
86 streamNotificationEvents
```

B01N remains the global bell + Quick Inbox.

```text
Notifications = personal attention / triage
Minha Caixa   = assigned authoring/governance work
```

B08 does not become a second Document Official, Discussion, viewer, task queue or Authorization surface.

## 3. B08-F1 — CLOSED / OPERATOR-RATIFIED

Each presentable `DOCUMENT_MENTION` op82 row is independently human-recognizable after server-side current disclosure:

```text
NotificationInboxItem
  notification identity + engagement state

  source
    current-disclosable DocumentReference
    exact immutable message_id
    author UserReference
    exact official_revision_at_post? RevisionReference
    bounded server-composed message_preview
```

Persistence remains Notification/source identities + engagement only. Source text/title/profile/ACL is not copied into Notifications authority.

Fixed views:

```text
active    = presentable + non-archived
unread    = presentable + non-archived + unread
archived  = presentable + archived
```

`unread` is a subset of active. There is no unseen list view.

## 4. P7 — H1 Focused Triage Inbox — OPERATOR-RATIFIED

```text
Notificações
  unseen + unread summary
  Marcar todas como lidas

[ Caixa de entrada ] [ Não lidas ] [ Arquivadas ]

one server-order list
  source recognition
  Nova distinct from Não lida
  Abrir conversa
  Marcar lida / não lida
  Arquivar / desarquivar

cursor continuation
```

Source-open remains a boundary to B03:

```text
op83 read=true
-> exact source.document.document_id + source.message_id
-> B03
-> op79 anchor_message_id
-> current disclosure rechecked
```

## 5. Engagement law represented

```text
READ => SEEN
mark unread preserves seen
archive/unarchive preserves read + seen
archive != delete
read != Document read / acknowledgement
```

Seen remains presentation-driven:

```text
fetch/cache != seen
materially present unseen row -> candidate for bounded op84 batch
```

P8 deliberately exposes the seen-batch ids in a review-only console so the operator can falsify whether loaded-but-off-screen rows are being marked seen prematurely.

## 6. P8 R1 — functional low-fidelity evidence

Canonical file:

```text
docs/work/current/t11-b08-notifications-full-inbox-functional-wireframe.html
```

The artifact is pure:

```text
HTML
CSS
vanilla JavaScript
deterministic local fixtures
IntersectionObserver only for P8 presentation/seen simulation
```

It contains no:

```text
React
backend/API calls
OpenAPI client
real SSE transport
frontend Authorization evaluator
Product persistence/schema implementation
```

### Structural verification before repository write

```text
HTML parse                          PASS
static ids                          33
duplicate static ids                0
inline JavaScript node --check      PASS
Product forbidden-control hits      0
local Git blob                      bb130535721b2381524763a4885ade5199a15596
repository blob                     bb130535721b2381524763a4885ade5199a15596
```

## 7. Material P8 interactions implemented

```text
active / unread / archived lenses
unseen + unread item
seen + intentionally unread item
read item
per-item read / unread
archive / unarchive preserving read state
archived unread item
mark-all-read affecting active unread only
IntersectionObserver seen queue + bounded batch simulation
loaded page-2 rows remaining unseen until materially presented
exact source-open B03 boundary with document_id + message_id
one-shot access-drift 404 / row disappearance
cursor continuation
cursor failure + retry preserving loaded rows
per-item mutation failure preserving state
mark-all-read failure preserving state
initial Inbox unavailable + retry
empty active / unread / archived fixtures
SSE `notifications.changed {}` invalidation + canonical refetch simulation
SSE disconnect/reconnect freshness behavior
Quick Inbox using the same local engagement/count state
responsive one-column treatment
Escape/focus/live-region recovery structure
```

Review-only fixture controls are explicitly outside Product UI.

## 8. Ordering / realtime represented

Canonical fixture order follows:

```text
created_at DESC,
notification_id DESC
```

No Product sort/search/custom filter exists.

Realtime fixture law:

```text
notifications.changed {}
-> invalidate
-> refetchCanonical fixture
```

The SSE simulation carries no Notification row payload.

## 9. Explicit non-goals preserved

```text
search
free-form/custom filters
filter by author/Document
saved views
bulk row selection/archive
mark-all-unread / mark-unseen
snooze / priority / reminders
Notification preferences
email/push settings
Notification deletion
Notification-kind selector
source reply/editor/viewer inside Inbox
```

## 10. Operator-use focus

The operator should especially exercise:

```text
1. Compare "Nova" vs "Não lida".
2. Mark a seen row unread and confirm it does not become Nova.
3. Archive a read row and inspect the same read state under Arquivadas.
4. Open an archived unread row, unarchive it, and inspect its engagement continuity.
5. Load page 2 and inspect the Seen batching console before scrolling lower rows into view.
6. Arm Próximo abrir -> access drift 404 and open one source.
7. Arm page/action/mark-all failures and confirm current visible truth is preserved.
8. Emit SSE changed {} and confirm the new row appears only after the simulated refetch.
9. Open the bell Quick Inbox and confirm counts/engagement match the Full Inbox.
10. Resize to narrow/mobile and inspect row actions + Quick Inbox sheet behavior.
```

## 11. Current gate

```text
entry recovery                         COMPLETE
P6                                     COMPLETE
B08-F1                                 CLOSED / OPERATOR-RATIFIED
P7 H1                                  OPERATOR-RATIFIED
P8 R1 functional HTML                  EXISTS / READY FOR OPERATOR USE
P8 material finding                    NONE RECORDED YET — OPERATOR USE PENDING
B08 LOCK                               NOT YET
P9 / P10                               NOT OPEN
B09+                                   NOT OPEN
```

Next gate:

> Operator opens/operates B08 P8 R1. If friction/finding exists, revise the same B08 HTML. Only explicit B08 LOCK opens P9 Screen Contract then P10 pattern consolidation. B09 remains closed.
