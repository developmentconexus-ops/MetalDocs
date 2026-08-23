# T11 — Frontend Foundation R1 — 86/11 bounded rebaseline

> **Status:** FP0-R1 CLOSED CANDIDATE / CURRENT T11 PLANNING INPUT.  
> **Scope:** bounded successor map to the pre-Discussion `t11-frontend-blueprint.md`; does not reopen B01/B01N/B02 LOCKs.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Current Product/architecture authority:** `docs/decisions/discussion-notifications-launch.md`.  
> **Program status authority:** `docs/roadmap.md` only.  
> **Implementation:** BLOCKED.

## 1. Why this rebaseline exists

The earlier frontend blueprint was derived from the pre-reopen world:

```text
stable SPA routes     10
application ops       78
Notifications         deferred
Document Discussion   absent
```

The T11 bounded reopen has now promoted current authority:

```text
stable SPA routes              11
PermissionCode values          16
application operations         86
Idempotency-Key creations      11
semantic owners                4 business + 2 supporting
```

A frontend plan must not continue to derive coverage/surfaces from the superseded 78/10 snapshot. This record performs the smallest global map update and nothing more.

## 2. Preserved decisions — no reopen

The following remain exactly as already operator-LOCKED:

```text
B01  App Shell + Global IA + Home
B01N Notification bell / badge / Quick Inbox / responsive transformation
B02  Library / Discovery
```

The following current architecture decisions are consumed, not reopened:

```text
4+2 semantic owners
11 stable SPA routes
16 PermissionCode values
86 operations
11 Idempotency-Key creations
same-Scope Mention -> Notification
server-side Notification presentability
Lexical as replaceable composer mechanism
SSE invalidation + in-process wake-up
River as sole durable async baseline
no generic EventBus / broker / Redis
```

## 3. Client-state authority — unchanged

Exactly four frontend state classes remain:

```text
SERVER STATE
  current Product/server truth -> TanStack Query in production

NAVIGATION / URL
  route, admitted filters/cursors, deep-link anchor, route-local presentation state

FORM DRAFT
  unaccepted human input / editor buffer

EPHEMERAL UI
  dialog/drawer/focus/selection/disclosure state
```

P8 uses local deterministic fixtures only to simulate these classes; fixture state is never Product authority.

## 4. Stable Product route map — 11

```text
/documents
/documents/:document_id
/documents/:document_id/work
/documents/:document_id/history
/work
/work/governance/:attempt_id
/notifications
/audit
/admin/organization
/admin/access
/admin/document-governance
```

Browser AuthN remains outside the application operation census:

```text
/auth/login
/auth/callback
```

## 5. Human-goal coverage — R1

Preserved pre-reopen goals:

```text
establish/end session
discover official documents
create Document
inspect official/current Document
start/enter open Revision
author DRAFT
upload replacement source
submit/withdraw/cancel
see actor-relevant work
participate in governance
inspect Document history
initiate/manage obsolescence
administer Organization
administer access
administer document governance
inspect Audit
```

New current Launch goals added by the bounded reopen:

```text
G17 Discuss a stable Controlled Document
  When I need to collaborate about the Document itself,
  I need a persistent Discussion that survives Revision changes,
  so conversation is not confused with DRAFT/editor or governance feedback.

G18 Mention an eligible User in a Discussion message
  When a colleague needs explicit attention,
  I need @mention autocomplete and atomic acceptance,
  so the intended User receives exactly one in-app Notification if the message is accepted.

G19 Notice and triage in-app Notifications
  When I receive attention items,
  I need global novelty indication, Quick Inbox and a full Inbox,
  so I can find, read/unread and archive current presentable Notifications without conflating them with assigned work.

G20 Return from a Notification to the exact Discussion context
  When I open a DOCUMENT_MENTION Notification,
  I need to reach the current Document Discussion at the target message,
  so I can understand and respond under current disclosure.
```

Current coverage result:

```text
20 accepted human goals mapped
0 new Product capabilities invented by frontend planning
```

## 6. Material surface inventory — R1

Existing families remain. R1 adds only the surfaces required by current authority:

```text
Application shell                 APP-01..02       preserved
Library                           LIB-01..02       preserved
Document Official                 OFF-01..05       preserved
Document Discussion               DSC-01..02       +2
Notifications                     NTF-01..03       +3
History                           HIS-01           preserved
My Work                           WRK-01..02       preserved
Document Work                     DW-01..04        preserved
Governance Case                   GOV-01..03       preserved
Admin Organization                ORG-01..08       preserved
Admin Access                      ACC-01..02       preserved
Admin Document Governance         DGV-01..06       preserved
Audit                             AUD-01           preserved
```

New handles are planning identifiers only:

```text
DSC-01  stable-Document Discussion timeline + anchor navigation
DSC-02  Discussion composer + reply + @mention selection
NTF-01  locked global bell / unseen badge / Quick Inbox
NTF-02  full /notifications Inbox
NTF-03  realtime invalidation consumer (no business truth in signal)
```

## 7. New Screen Contract inputs — lightweight FP0 map

Full exact per-control contracts remain P9 after each block LOCK.

```text
DSC-01
  truth       listDocumentDiscussionMessages
  identity    document_id + optional anchor_message_id
  disclosure  DocumentDiscussionDisclosure
  state       cursor/anchor in URL/navigation; rows server state
  law         Discussion is stable-Document truth, never DRAFT/governance feedback

DSC-02
  truth       current composer draft + mention candidates
  read        searchDocumentDiscussionMentionCandidates
  write       createDocumentDiscussionMessage
  identity    document_id + optional reply_to_message_id + Mention user_id
  law         author current document.discuss + disclosure; all Mentions revalidated atomically
  retry       11th Idempotency-Key creation

NTF-01
  truth       same Notifications read family as NTF-02
  action      mark seen / mark all read where admitted
  law         unseen != unread; opening bell never blindly marks unloaded items seen

NTF-02
  truth       listNotifications(view=active|unread|archived)
  writes      updateNotificationEngagement / markNotificationsSeen / markAllNotificationsRead
  law         recipient-self + current source disclosure; no client post-filter authority

NTF-03
  input       streamNotificationEvents
  effect      invalidate/refetch canonical Notification queries
  law         SSE carries zero source business truth and may be lost without semantic loss
```

## 8. Operation coverage — 86 / 86

Operations 1–78 preserve their previously assigned frontend homes.

Bounded delta:

```text
79 listDocumentDiscussionMessages
   -> B03 / DSC-01

80 createDocumentDiscussionMessage
   -> B03 / DSC-02

81 searchDocumentDiscussionMentionCandidates
   -> B03 / DSC-02

82 listNotifications
   -> B01N Quick Inbox + B08 Full Inbox

83 updateNotificationEngagement
   -> B08; bounded source-open/read interaction may be initiated from B01N

84 markNotificationsSeen
   -> B01N + B08 presentation reconciliation

85 markAllNotificationsRead
   -> B01N + B08

86 streamNotificationEvents
   -> B01N + B08 realtime invalidation mechanism
```

Reconciliation:

```text
operations mapped          86 / 86
unassigned operations       0
invented operations         0
Idempotency-Key creations  11
ETag domains               13 / 13
exact-byte resources        4
```

## 9. Frontend block program — FP1 sequence

Block IDs are planning handles; `docs/roadmap.md` alone owns live status.

```text
B01   App Shell + Global IA + Home
B01N  Notification chrome + Quick Inbox
B02   Library / Discovery
B03   Document Official / Ficha + Viewer + Discussion
B04   Document Work / Authoring
B05   My Work
B06   Governance Case
B07   Document History
B08   Notifications Full Inbox
B09   Audit
B10   Organization Administration
B11   Access Administration
B12   Document Governance Administration
```

Boundary laws:

```text
B01N owns global attention chrome, not full triage.
B03 owns stable Document Discussion because it is contextual to the exact Document.
B08 owns the full Notifications Inbox route/lenses/engagement workspace.
B04/B06/B07 remain separate accepted Product lenses and are not pre-designed by B03.
```

## 10. Block dependencies

```text
B01 -> all blocks inherit shell/IA
B01N -> B03 and B08 inherit Notification entry/deep-link semantics
B02 -> B03 Library result entry
B03 -> B04 Work transition only; does not design B04
B03 -> B07 History transition only; does not design B07
B03 Discussion -> B08 source-navigation contract
B08 -> B03 exact DOCUMENT_MENTION deep-link
```

No dependency permits an unopened downstream block to be generated as baseline.

## 11. FP0-R1 exit proof

```text
current 11-route map represented                    PASS
current operations 79–86 receive frontend homes    PASS
Discussion human goals represented                  PASS
Notification human goals represented                PASS
B01/B01N/B02 existing LOCKS preserved               PASS
B03 current scope includes Discussion               PASS
B08 full Inbox explicitly exists in roadmap         PASS
frontend Authorization authority invented           0
screen-shaped APIs invented                          0
```

## 12. Disposition of older T11 frontend artifacts

```text
t11-frontend-blueprint.md
  PRE-REOPEN SNAPSHOT / EVIDENCE
  its 78/10 counts do not override R1/current authority

t11-wireframes.md
  PRE-v2.2 / PRE-REOPEN planning evidence
  may inform later blocks but cannot serve as current P8 LOCK evidence
```

Do not rewrite old evidence merely to make historical counts look current. Current program/status comes from `docs/roadmap.md`; current bounded frontend map is this R1 record plus current Product/architecture authorities.

## 13. Next gate

```text
FP0-R1 complete
-> B03 current candidate record under Method v2.2
-> canonical functional low-fi HTML/CSS/JS P8
-> operator operates / discusses / iterates
-> B03-F1 closes before final LOCK
-> operator-only B03 LOCK
-> P9 exact Screen Contract
-> P10 bounded pattern pass
```
