# T11 — B08 Notifications Full Inbox R1 — Method v2.2 locked

> **Status:** LOCKED / OPERATOR-RATIFIED / P9-P10 COMPLETE.  
> **Block:** B08 — Notifications Full Inbox.  
> **Method:** Frontend Product Experience Planning Method v2.2 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01 / B01N / B02 / B03 / B04 / B05 / B06 / B07 LOCKED.  
> **Bounded authorities:** `../../decisions/discussion-notifications-launch.md` + `../../decisions/notification-inbox-recognition-read.md`.  
> **Canonical P8 R1:** `t11-b08-notifications-full-inbox-functional-wireframe.html`.  
> **Canonical HTML blob:** `bb130535721b2381524763a4885ade5199a15596`.  
> **P9:** `t11-b08-screen-contract.md`.  
> **P10:** `t11-b08-pattern-consolidation.md`.  
> **Implementation:** BLOCKED.

## 1. Operator LOCK basis

The operator approved the B08 design, ratified the written candidate, operated P8 R1 and then explicitly approved the operated result.

Locked sequence:

```text
B08 entry / continue FP1                     OPERATOR-AUTHORIZED
B08-F1 + H1 in-chat design                    OPERATOR-APPROVED
written B08 candidate                         OPERATOR-RATIFIED
P8 R1 functional HTML                         OPERATED / OPERATOR-APPROVED
B08 final LOCK                                OPERATOR-RATIFIED
P9 Screen Contract                            COMPLETE
P10 Pattern Consolidation                     COMPLETE
```

No operator-requested structural revision was required after P8 R1.

## 2. Product boundary locked

Stable route and operation homes remain:

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

B08 is not a second Document Official, Discussion, viewer, task queue, Audit stream or Authorization surface.

Current census remains:

```text
86 operations / 11 routes / 16 PermissionCode values
```

## 3. B08-F1 — CLOSED / OPERATOR-RATIFIED

Each presentable `DOCUMENT_MENTION` op82 row is independently human-recognizable after current server-side disclosure:

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

Notification persistence remains source identities + engagement only. The following are not copied into Notifications persistence merely for rendering:

```text
Document title/code snapshots
message text/preview
author display name
Revision title
source ACL/presentability truth
```

Fixed list meanings remain:

```text
active    = presentable + non-archived
unread    = presentable + non-archived + unread
archived  = presentable + archived
```

`unread` is a subset of active. There is no unseen list view.

## 4. Locked experience — Focused Triage Inbox

```text
Notificações
  unseen + unread summary
  Marcar todas como lidas

[ Caixa de entrada ] [ Não lidas ] [ Arquivadas ]

one canonical recency-ordered list
  human-recognizable source
  Nova distinct from Não lida
  Abrir conversa
  Marcar lida / não lida
  Arquivar / desarquivar

cursor continuation
```

Canonical order stays server-owned:

```text
created_at DESC,
notification_id DESC
```

No browser grouping or re-sort by source, author, read state or Document becomes collection authority.

## 5. Engagement law locked

```text
READ => SEEN
mark unread clears read_at only
mark unread never restores unseen/new
archive/unarchive preserves read + seen
archive != delete
read != Document read / acknowledgement
```

Seen is presentation-driven:

```text
fetch/cache != seen
materially presented unseen row -> candidate for bounded op84 batch
server intersects current recipient + current presentability
```

B08 exposes no manual `mark seen` / `mark unseen` Product control.

## 6. Source handoff locked

Source activation uses exact server-returned identities:

```text
op82 source.document.document_id
+ op82 source.message_id
```

Sequence:

```text
user activates source
-> op83 read=true
-> admitted success
-> B03 Document Official route
-> B03 op79 anchor_message_id = exact source.message_id
-> current B03 disclosure rechecked
```

If op83 returns non-disclosing 404 because current presentability drifted:

```text
refetch Inbox
stale row may disappear
no source navigation
no reason/source metadata leak
```

A Notification being read never proves the source Document/message was viewed.

## 7. Realtime locked

Operation 86 remains invalidation only:

```text
event: notifications.changed
data: {}
```

Frontend behavior:

```text
signal
-> invalidate/refetch canonical op82 state
```

SSE contains no Notification/Document/message/User business payload. Disconnect is a freshness degradation only; reconnect/focus/refetch restores canonical truth.

## 8. P8 R1 evidence

Canonical file:

```text
docs/work/current/t11-b08-notifications-full-inbox-functional-wireframe.html
```

R1 is behaviorally truthful but technically disposable:

```text
HTML + CSS + vanilla JavaScript
local deterministic fixtures
no React
no backend/API call
no OpenAPI client
no real SSE transport
no frontend Authorization evaluator
```

Structural verification captured before repository write:

```text
HTML parse                          PASS
static ids                          33
duplicate static ids                0
inline JavaScript node --check      PASS
Product forbidden-control hits      0
local Git blob                      bb130535721b2381524763a4885ade5199a15596
repository blob                     bb130535721b2381524763a4885ade5199a15596
```

Material behavior exercised:

```text
active / unread / archived lenses
unseen + unread item
seen + intentionally unread item
read item
read / unread
archive / unarchive preserving state
archived unread item
mark-all-read affecting active unread only
presentation-driven seen batching
loaded-but-off-screen rows staying unseen until presentation
exact B03 source boundary
access-drift 404 disappearance
cursor continuation + failure/retry
per-item + mark-all failure
initial unavailable state
three empty states
SSE invalidation/refetch + disconnect/reconnect
Quick Inbox / Full Inbox shared engagement/count state
responsive/accessibility recovery structure
```

## 9. P9 Screen Contract closure

Canonical P9:

```text
docs/work/current/t11-b08-screen-contract.md
```

Closure proof:

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

Operation homes remain 82–86; B03 op79 is only the destination owner read after explicit source handoff.

## 10. P10 Pattern Consolidation closure

Canonical P10:

```text
docs/work/current/t11-b08-pattern-consolidation.md
```

Existing shared patterns reused:

```text
Global App Shell
Notification Quick Inbox
```

No new shared Product/component pattern graduated.

B08-local patterns retained:

```text
Focused Triage Inbox
fixed active/unread/archived lens set
human-recognizable Notification row
novelty separate from read state
presentation-driven seen batching
Notification engagement reconciliation
source engagement + owner-lens handoff
SSE invalidation-only reconciliation
```

P10 explicitly rejects a generic Inbox, Activity/Event feed, generic filter engine, realtime entity sync, deep-link resolver or generic seen/read state machine.

Closure:

```text
existing locked shared patterns reused           2
new shared semantic patterns graduated            0
B08-local semantic/composition patterns retained  8
false abstractions introduced                     0
Notifications/Minha Caixa semantic merges         0
source-workspace duplications                      0
```

## 11. Explicit non-goals preserved

B08 Launch does not add:

```text
search
free-form/custom/saved filters
filter by author/Document
bulk row selection/archive
mark-all-unread / mark-unseen
snooze / priority / reminders
Notification preferences
email/push settings
Notification deletion
Notification-kind selector
source reply/editor/viewer inside Inbox
```

## 12. Final block closure

```text
entry recovery                         COMPLETE
P6                                     COMPLETE
B08-F1                                 CLOSED / OPERATOR-RATIFIED
P7 H1                                  OPERATOR-RATIFIED
P8 R1                                  LOCKED / OPERATOR-RATIFIED
P9                                     COMPLETE
P10                                    COMPLETE
B08                                    LOCKED / OPERATOR-RATIFIED
B09                                    NOT OPEN / NEXT ELIGIBLE
B10-B12                                NOT OPEN
```

Implementation remains blocked. B08 LOCK does not authorize Product code/schema/OpenAPI/runtime/deploy work, T12, merge, B09 opening or T11 closeout.
