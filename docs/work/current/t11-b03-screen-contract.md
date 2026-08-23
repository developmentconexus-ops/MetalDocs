# T11 — B03 P9 Screen Contract / Bidirectional Trace

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B03 — Document Official.  
> **Depends on:** B03 LOCK, Discussion/Notifications authority, responsible-owner precision, Document Official action-hint precision.  
> **Implementation:** BLOCKED.

## 1. Goal

Prove that every material region/control in the locked B03 wireframe is realizable by current authority without inventing frontend business truth or a screen-shaped API.

## 2. Screen contract

| Surface / region | User goal | Current read truth | Material control / write | Identity source | Material failure / safe UX | Forbidden frontend authority | Status |
|---|---|---|---|---|---|---|---|
| B03-01 Document hero | know exact Document + official state | `getDocument -> DocumentOfficialView` | View / Download / History navigation | returned `document_id`, official Revision/release refs | 404 non-disclosable; official absent means no fake viewer | no status/lifecycle derivation | READY |
| B03-02 current-work context | know newer work exists without replacing official truth | `DocumentOfficialView.open_revision?` | `Abrir trabalho` navigation only | returned `document_id` + open Revision routing ref | absent means no current work disclosed | no History fallback/current resolver | READY |
| B03-03 ficha | understand Type/Area/owner/release metadata | `getDocument` | none by virtue of display | returned references | stale display reconciles through refetch | no normalized global Document entity | READY |
| B03-04 official preview | recognize current official content | official `ReleasedRevisionView` + representation/source summary | open same official viewer | returned release/rendition/source refs | integrity failure -> no partial preview-as-truth | preview never semantic/exact-content authority | READY |
| B03-05 official viewer | deliberately read exact official bytes | `getRelease`, `getReleaseSource`, `getOfficialRenditionContent` as applicable | Download exact resource | exact returned release/rendition ids | 404/403 non-disclosing; integrity failure -> no partial success | no edit/submit/governance controls | READY |
| B03-06 responsible owner | inspect/change responsibility | `getDocument`, `responsible_owner_candidates?`, `getDocumentResponsibleOwner + ETag` | `replaceDocumentResponsibleOwner(target user_id, If-Match)` | selected returned UserReference + exact owner ETag | 412 -> refetch owner while preserving selection intent; target disabled -> fail closed | no target-eligibility evaluator | READY |
| B03-07 management actions | see currently meaningful Document commands | `DocumentOfficialView.allowed_actions` | create Revision / create or withdraw obsolescence | Document id + current request ref where disclosed | hint may stale; command denial/conflict remains authoritative | no permission/status matrix | READY |
| B03-08 create Revision | enter next authoring work when currently admissible | `allowed_actions` + current Document | `createDocumentRevision` | current `document_id` | command rechecks current eligibility; conflict stays on ficha with refreshed truth | no client lifecycle transition | READY |
| B03-09 obsolescence | start/manage obsolescence | `allowed_actions` + `active_obsolescence_request_id?` + request read | `createObsolescenceRequest`, `withdrawObsolescenceRequest` | `document_id` / returned request id | current command errors preserve official truth; no fake obsolete state | no client obsolescence state machine | READY |
| B03-10 revisions context | distinguish official vs current open work | `official?` + `open_revision?` | View official / Work / History navigation | returned identities | absent open work stays absent; no reconstructed lineage | no second History authority | READY |
| B03-11 Discussion list | inspect stable-Document conversation | op79 `listDocumentDiscussionMessages` | page older messages | `document_id`, opaque cursor | 404/non-disclosable anchor follows normal law | no chat/thread aggregate invented | READY |
| B03-12 Discussion reply/send | post stable-Document message | current Discussion disclosure + `document.discuss` enforced server-side | op80 `createDocumentDiscussionMessage` | `document_id`, optional returned prior `message_id`, stable Mention user ids | invalid Mention rejects whole message; ambiguous retry uses same Idempotency-Key | no parsed `@Name` identity authority | READY |
| B03-13 Mention autocomplete | select eligible Mention target | op81 `searchDocumentDiscussionMentionCandidates` | select UserReference into composer | stable returned `user_id` | candidate drift revalidated at op80 commit | no local eligibility directory | READY |
| B03-14 Notification deep-link | return to exact mention context | B01N/Notifications source identity + op79 anchor semantics | Notification engagement handled by Notifications; B03 reveals anchor | `document_id + message_id` from Notification source | access drift -> Notification omitted/non-presentable; anchor non-disclosable -> normal not-found | Notification never access grant | READY |
| B03-15 responsive reflow | preserve same meaning on narrow screens | same server truths | presentation-only reorder/stack | none new | no hidden-only material action | no mobile-specific business semantics | READY |

## 3. Exact operation homes used by B03

Current retained operation numbering:

```text
47  getDocument
48  getDocumentResponsibleOwner
49  replaceDocumentResponsibleOwner
52  createDocumentRevision
72  getRelease
73  getReleaseSource
74  getOfficialRenditionContent
75  createObsolescenceRequest
76  getObsolescenceRequest
77  withdrawObsolescenceRequest
79  listDocumentDiscussionMessages
80  createDocumentDiscussionMessage
81  searchDocumentDiscussionMentionCandidates
```

Notifications operations 82–86 remain primarily owned by B01N/B08. B03 consumes only the admitted source navigation contract; it does not duplicate Notification state ownership.

## 4. Bidirectional trace

### Product/backend -> frontend

```text
Document Official read
-> B03 hero/ficha/current-work/revision/preview context

exact official content resources
-> B03 viewer/download

responsible-owner read + candidate precision
-> B03 owner drawer

DocumentOfficialView.allowed_actions
-> B03 management affordances

Discussion operations 79–81
-> B03 Discussion timeline/composer/Mention selection

Notification DOCUMENT_MENTION source identity
-> B03 exact Discussion anchor destination
```

### Frontend -> Product/backend

```text
Visualizar / preview
-> exact Release/source/rendition reads

Alterar responsável
-> get owner ETag -> replace owner

Criar revisão
-> createDocumentRevision

Solicitar / retirar obsolescência
-> existing obsolescence operations

Responder / enviar Discussion
-> op80

@
-> op81

Notification source click
-> Notifications engagement owner + B03 op79 anchor read
```

Unbound material controls: **0**.  
Invented application operations: **0**.  
Screen-shaped APIs: **0**.

## 5. Client state classes

B03 uses only accepted state classes:

```text
SERVER STATE
  DocumentOfficialView, owner ETag view, Discussion pages, exact content refs

NAVIGATION / URL
  document_id, destination route, Discussion anchor intent

FORM DRAFT
  owner selection before submit, Discussion composer text/reply target

EPHEMERAL UI
  viewer open/closed, Quick Inbox visibility inherited from B01N,
  action menu, owner drawer, anchor highlight/focus
```

No additional durable/global client state class is justified.

## 6. Material failure intent

```text
404 / non-disclosable
  user understands resource/context is unavailable; no disclosure reason leaked

403 permission.denied on command
  user understands action is no longer admitted; refetch current read truth

412 responsible-owner resource_changed
  preserve intended target selection, refetch current owner/ETag, require deliberate retry

409/state conflict on Revision/obsolescence
  keep official truth intact, refetch current Document, explain that state changed

ambiguous op80 outcome
  retry same logical message with same Idempotency-Key; never create duplicate Message/Mention/Notification

invalid Mention
  message not accepted; preserve composer for correction; no partial Mention/Notification success

exact-content integrity failure
  no partial viewer success; keep ficha available
```

Final copywriting remains outside P9; these are semantic message intents only.

## 7. Access / disclosure proof

```text
button visibility != Authorization
allowed_actions != Authorization snapshot
responsible_owner_candidates != target guarantee
Mention candidate != commit guarantee
Notification presence != access grant
History/preview/viewer != current lifecycle authority
```

Every material command rechecks server truth.

## 8. P9 closure

```text
material B03 regions/controls traced        15 / 15
unbound material controls                   0
invented operations                         0
screen-shaped APIs                          0
frontend Authorization evaluator            0
navigation identities unsourced             0
material B03 Screen Contract findings       0
```

P9 is complete for the locked B03 scope.
