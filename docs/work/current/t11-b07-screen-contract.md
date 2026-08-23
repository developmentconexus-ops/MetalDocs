# T11 — B07 P9 Screen Contract / Bidirectional Trace

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B07 — Document History.  
> **Depends on:** B07 operator LOCK, `document-history-recognition-read.md`, current T6/T8-E/T8-F authority.  
> **Implementation:** BLOCKED.

## 1. Goal

Prove that every material region/control in the locked B07 functional wireframe is realizable by current authority without inventing frontend historical truth, a screen-shaped API, a History mutation surface or an Audit dependency.

## 2. Screen contract

| Surface / region | User goal | Current read truth | Material control / write | Identity source | Material failure / safe UX | Forbidden frontend authority | Status |
|---|---|---|---|---|---|---|---|
| B07-01 History route | open one Document's controlled history | op53 `getDocumentHistory` | route only | `document_id` | 404/non-disclosable -> neutral unavailable state | no Audit fallback | READY |
| B07-02 orientation header | know which Document/history is open | op47 `getDocument` for current orientation only | return to B03 | returned `DocumentReference` / official context | orientation may refresh independently | op47 never becomes History event authority | READY |
| B07-03 Revision marker | recognize exact business Revision for each chronological segment | op53 `DocumentHistoryItem.revision` | none | returned `RevisionIdentity` | missing identity is contract failure, not client inference | no UUID->ordinal graph reconstruction | READY |
| B07-04 revision-created event | know when a business Revision began | op53 `revision_created` | none | returned `revision` + `occurred_at` | render without fabricated title snapshot | no title-at-creation invention | READY |
| B07-05 Submission event | recognize immutable submitted attempt | op53 `submission_created` | open exact historical Submission | returned `submission_id`, `revision`, title/submitter | exact-content failure preserves event truth | no current DRAFT substitution | READY |
| B07-06 governance feedback | understand immutable review context | op53 `feedback_added` | none | returned attempt, subject kind, revision, actor | pagination/content around it may fail without erasing event | no Document Discussion substitution | READY |
| B07-07 governance Decision | understand who decided, outcome and human Step | op53 `governance_decision` | none | returned `step_label`, `subject_kind`, revision, actor | RETURN reason follows union presence law | no current route lookup for Step label | READY |
| B07-08 return reason | understand why content returned | same Decision event | none | returned `reason` iff `RETURN_FOR_CHANGES` | absent on ACCEPT remains truthful | no feedback-as-reason inference | READY |
| B07-09 Submission withdrawal | understand attempt termination without fake reject | op53 `submission_withdrawn` | none | returned submission/revision/actor | event remains historical fact | no current lifecycle mutation | READY |
| B07-10 Revision cancellation | understand stopped business change cycle | op53 `revision_cancelled` | none | returned revision/actor/reason | prior EFFECTIVE truth is not reconstructed by browser | no cancel/reopen control | READY |
| B07-11 Release event | understand when exact Revision became official | op53 `release_completed` | open exact released content | returned `release_id`, revision, optional predecessor Revision | content failure never substitutes current/newer Release | no compare/restore action | READY |
| B07-12 official-rendition event | understand derived official representation completion | op53 `official_rendition_completed` | display only in current P8 | returned rendition/submission/revision ids | event survives exact viewer failure elsewhere | no representation-as-lifecycle authority | READY |
| B07-13 obsolescence events | understand request/withdrawal/completion against exact target Revision | op53 obsolescence variants | none | returned request id + exact revision | later event may target older Revision; chronology stays server-owned | no event reordering into prior pages | READY |
| B07-14 later older-Revision marker | preserve chronology when an older Revision receives later events | same ordered op53 items | presentation-only repeated Revision heading | each event's returned `revision` | no event is moved backward in time | no client global regroup/re-sort authority | READY |
| B07-15 cursor continuation | continue long History in canonical order | op53 cursor page | load next page | opaque cursor | current disclosure/AuthZ rechecked every page | no total count/frozen snapshot | READY |
| B07-16 continuation failure | recover without losing already loaded history | op53 continuation failure + retained pages | retry same continuation intent | current opaque cursor | preserve loaded events/read position; retry | no synthetic completed page | READY |
| B07-17 exact Submission viewer | inspect immutable submitted bytes | op63 `getSubmission` + op64 `getSubmissionSource` | open/close read-only viewer | exact `submission_id` | 403/404/integrity failure -> no substitute bytes | no edit/submit/governance controls | READY |
| B07-18 exact Release viewer | inspect exact historical official bytes | op72 `getRelease` + op73/74 as representation requires | open/close read-only viewer | exact `release_id` and returned representation identity | integrity/dependency failure -> no current-version fallback | no restore/delete/compare | READY |
| B07-19 global notification chrome | preserve app-wide attention behavior | B01N Notifications authority | Quick Inbox open/close only | Notification source identities | B07 does not interpret Notification state | no B07 Notification store | READY |
| B07-20 responsive/accessibility | preserve chronological meaning on narrow/accessibility paths | same server truths | presentation/focus only | none new | headings/event text/viewer close remain operable | no mobile-specific business semantics | READY |

## 3. Exact operation homes used by B07

```text
47  getDocument
53  getDocumentHistory
63  getSubmission
64  getSubmissionSource
72  getRelease
73  getReleaseSource
74  getOfficialRenditionContent
```

Notifications operations remain owned by B01N/B08; B07 only inherits the already-locked global chrome.

No operation 87+ is needed.

## 4. Bidirectional trace

### Product/backend -> frontend

```text
Document Official current read
-> B07 orientation only

DocumentHistoryPage + B07-F1 projection
-> Revision markers + chronological lifecycle/event spine

submission_created.submission_id
-> exact historical Submission viewer

release_completed.release_id
-> exact historical Release/source/rendition viewer
```

### Frontend -> Product/backend

```text
open History
-> op53

return to Document
-> B03 route; B03 rereads its own truth

load more
-> op53 cursor continuation

Ver conteúdo submetido
-> op63 -> op64

Ver versão oficial
-> op72 -> op73 or op74 according to returned representation
```

Unbound material controls: **0**.  
Invented application operations: **0**.  
Screen-shaped APIs: **0**.

## 5. Client state classes

```text
SERVER STATE
  op47 orientation
  ordered op53 History pages
  exact Submission / Release descriptors and bytes

NAVIGATION / URL
  document_id

EPHEMERAL UI
  loaded-page presentation
  selected historical event/viewer
  viewer page/focus state
  continuation retry state
  Quick Inbox visibility inherited from B01N
```

B07 has no FORM DRAFT and no durable client business state.

The frontend may visually segment consecutive ordered events by returned `RevisionIdentity`, including repeating a Revision marker later. It may not build a cross-page entity graph, globally reorder history or reconstruct missing semantic identities.

## 6. Material failure intent

```text
404 / non-disclosable History
  -> neutral unavailable state; safe return to B03; no disclosure reason leak

continuation failure
  -> preserve already loaded truthful events and reading position; retry continuation

current op47 orientation drift
  -> refresh orientation without rewriting accepted historical events

Submission exact-content failure
  -> event remains visible; no other/current Submission is rendered

Release exact-content failure
  -> event remains visible; no newer/current Release is rendered
```

Final copywriting remains outside P9; these are semantic message intents only.

## 7. Access / disclosure proof

```text
History presence != Audit access
History event identity != current lifecycle authority
op47 current official context != historical event source
historical content identity != permission grant
viewer availability != restore/edit authority
loaded prior cursor page != authority for interpreting a later event
```

Every read performs current server disclosure/AuthZ according to its own operation.

## 8. Backend sufficiency

B07-F1 closed the only material read-symmetry gap discovered before P8:

```text
every event has exact RevisionIdentity
Decision has frozen step_label + subject_kind
feedback has subject_kind
Release predecessor has RevisionIdentity
```

Therefore the locked B07 requires:

```text
new operations        0
new routes            0
new Permissions       0
new lifecycle state   0
new History mutation  0
Audit joins            0
```

Current census remains **86 operations / 11 routes / 16 PermissionCode values**.

## 9. P9 closure

```text
material B07 regions/controls traced        20 / 20
unbound material controls                    0
invented operations                          0
operation 87+                                0
screen-shaped APIs                           0
frontend historical graph authority          0
frontend Authorization evaluator             0
History mutations                            0
Audit reconstruction dependencies            0
material B07 Screen Contract findings        0
```

P9 is complete for the locked B07 scope.
