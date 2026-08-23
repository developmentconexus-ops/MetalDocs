# T11 — B04 P9 Screen Contract / Bidirectional Trace

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B04 — Document Work / Authoring.  
> **Depends on:** B04 LOCK, current T2 DRAFT OCC, T4 exact-content admission, T6/T8-E Work operations, T8-F frontend realization.  
> **Implementation:** BLOCKED.

## 1. Goal

Prove that every material region/control in the locked B04 functional wireframe is realizable by current authority without inventing frontend business truth, a second DRAFT authority or a screen-shaped API.

## 2. Screen contract

| Surface / region | User goal | Current read truth | Material control / write | Identity source | Material failure / safe UX | Forbidden frontend authority | Status |
|---|---|---|---|---|---|---|---|
| B04-01 Work route resolver | enter exact current open work | `getDocument -> open_revision?` | route only | returned `document_id + revision_id + open state` | absent/non-disclosable -> no current work; no History fallback | no History/current-work reconstruction | READY |
| B04-02 Work header | know Document, Work Revision/state and official-vs-work separation | `getDocument` + `getRevision` | back to B03 only | returned Document/Revision identities | stale context reconciles by refetch | no official/current lifecycle derivation | READY |
| B04-03 DOCX DRAFT editor | edit exact current WorkingContent | `getRevisionDraft + ETag` + `getRevisionDraftSource` | local title/Eigenpal buffer edits | exact returned `revision_id` + current DRAFT ETag | source load/integrity failure -> no editable false-success | local buffer is FORM DRAFT only | READY |
| B04-04 hybrid save status | understand dirty/saving/saved/failure truth | local buffer state + authoritative accepted DRAFT ETag | same persistence pipeline | local generation + current ETag | failure preserves buffer; never false `saved` | save badge is not lifecycle truth | READY |
| B04-05 Save now / Ctrl+S | force persistence deliberately | same DRAFT read truth | `startRevisionDraftUpload` + provider PUT + `completeRevisionDraftUpload` as needed, then `updateRevisionDraft` | current revision + fresh upload id + current ETag | upload/admission/PATCH failure preserves local intent | no second save endpoint / bypass | READY |
| B04-06 coalesced background save | avoid data loss without save storms | same as B04-05 | same pipeline, at most one in flight | exact local generation captured for the save | edits arriving during save remain dirty; next save follows | debounce timing is not Product authority | READY |
| B04-07 DRAFT OCC reconciliation | recover from concurrent edit safely | `412 precondition.draft_changed` + refetched `getRevisionDraft` | explicit discard/rebase-for-manual-copy choice only | stale ETag vs newly returned ETag | preserve local input; stop stale autosave; no auto-merge | no LWW / CRDT / client conflict resolver authority | READY |
| B04-08 PDF DRAFT viewer | inspect non-editable working PDF | `getRevisionDraft` + `getRevisionDraftSource` | replace source only | revision id + exact working source | integrity/dependency failure -> no partial success | no fake PDF editing | READY |
| B04-09 source replacement | replace working source deliberately | current DRAFT + ETag | op59 allocation -> provider PUT -> op60 completion -> op58 attach | returned upload id + revision id + current ETag | provider success != READY; attach may fail/stale | client filename/type/hash never exact-content authority | READY |
| B04-10 expired-upload recovery | recover without losing selected file | `410 state.upload_expired` | start a new allocation and repeat | same intended local bytes + **new** upload id | old claim never revived; local file preserved | no upload-id resurrection | READY |
| B04-11 submit force-flush | submit exactly what the author sees | current accepted DRAFT + ETag | op62 `createSubmission` after all saves/attach settle | revision id + exact DRAFT If-Match + logical Idempotency-Key | save/conflict/upload pending blocks submit; submit errors preserve Work truth | no optimistic `SUBMITTED` before accepted result | READY |
| B04-12 SUBMITTED read-only view | understand current immutable Submission/gates | `getRevision.current_submission_id` + `getSubmission` + `getSubmissionSource` | none by virtue of viewing | returned submission/revision identities | inaccessible/terminated context refetches route truth | no DRAFT editor while submitted | READY |
| B04-13 withdraw Submission | resume editing same Revision before Release | current Submission truth | op65 `withdrawSubmission` | returned current submission id | conflict/denial keeps current server state; then refetch | no fabricated DRAFT transition | READY |
| B04-14 cancel Revision | deliberately end current open Revision | current Revision state | op66 `cancelRevision(reason)` | exact revision id | missing reason blocked locally; command state conflict remains authoritative | no client cancellation state machine | READY |
| B04-15 no-current-work terminal | understand that Work ended | route resolver after withdrawal/cancel/release drift | navigate B03 | document id | no History fallback; no stale editor | no resurrection of closed Revision | READY |
| B04-16 operational rail + context | stay oriented without duplicating B03/B06/B07 | Work reads + bounded `getDocument` context | source/actions/context disclosure only | returned Type/Area/owner/official refs | stale display reconciles through normal refetch | no second ficha/history/governance authority | READY |
| B04-17 B01N global chrome reuse | preserve application-wide attention affordance | locked Notifications chrome | source navigation remains owned by B01N/B08/B03 destinations | Notification source identity | B04 does not interpret source business truth | no B04 Notification store | READY |
| B04-18 responsive reflow | preserve same Work semantics on narrow screens | same server truths | presentation-only stack/reorder | none new | no material control hidden without accessible alternative | no mobile-specific business semantics | READY |

## 3. Exact operation homes used by B04

Current retained operation numbering:

```text
47  getDocument                     route/current official orientation
56  getRevision                     exact Revision identity/state
57  getRevisionDraft                DRAFT body + strong ETag
58  updateRevisionDraft             title/source attach under If-Match
59  startRevisionDraftUpload        allocate upload capability
60  completeRevisionDraftUpload     server admission / READY
61  getRevisionDraftSource          exact working source bytes
62  createSubmission                freeze exact accepted DRAFT generation
63  getSubmission                   immutable submitted truth/gates
64  getSubmissionSource             exact immutable Submission source
65  withdrawSubmission              same Revision SUBMITTED -> DRAFT
66  cancelRevision                  open Revision -> CANCELLED
```

Provider PUT is an external upload mechanism explicitly orchestrated between ops59 and 60; it is not an application API operation and never becomes Product authority.

B04 adds no operation 87+.

## 4. Bidirectional trace

### Product/backend -> frontend

```text
DocumentOfficialView.open_revision
-> B04 route resolver

RevisionView + DocumentWorkView + ETag
-> Work header / editable or read-only DRAFT mode / persistence state

exact DRAFT source
-> Eigenpal DOCX input or PDF DRAFT viewer

upload allocation/admission + DRAFT PATCH
-> source replacement and DOCX persistence pipeline

SubmissionView + exact Submission source
-> SUBMITTED read-only mode + gate summary

withdraw/cancel commands
-> refetch current route truth / DRAFT or no-current-work state
```

### Frontend -> Product/backend

```text
local title/DOCX edit
-> transient FORM DRAFT
-> save pipeline
-> ops59/60 when bytes change
-> op58 If-Match

Salvar agora / Ctrl+S
-> same pipeline, no alternate mutation

Substituir arquivo
-> op59 -> provider PUT -> op60 -> op58

Enviar para análise
-> wait/flush save pipeline
-> op62 with exact current DRAFT If-Match + one logical Idempotency-Key

Retirar submissão
-> op65

Cancelar revisão
-> op66 with reason

Voltar para ficha / no-current-work
-> B03 stable Document route
```

Unbound material controls: **0**.  
Invented application operations: **0**.  
Screen-shaped APIs: **0**.

## 5. Client state classes

B04 uses only the accepted four classes:

```text
SERVER STATE
  DocumentOfficialView routing/orientation
  RevisionView
  DocumentWorkView + response ETag
  SubmissionView

NAVIGATION / URL
  document_id + stable /documents/:document_id/work route

FORM DRAFT
  editable title
  Eigenpal local DOCX buffer
  selected local replacement file until server acceptance
  cancellation reason before command

EPHEMERAL UI
  save/upload progress presentation
  right-rail disclosure
  conflict/reconciliation dialog
  cancellation dialog
  transient failure/retry state
```

No IndexedDB/localStorage/global entity store/offline DRAFT authority is admitted.

## 6. Material failure intent

```text
404 / non-disclosable Work
  no current work available; offer B03 return; never infer History target

403 permission.denied
  server remains authority; preserve safe local input where relevant

412 precondition.draft_changed
  stop stale autosave
  preserve local buffer
  refetch DRAFT + ETag
  explicit human reconciliation
  no automatic merge / no LWW

410 state.upload_expired
  preserve selected local bytes
  allocate a fresh upload id
  never revive expired claim

422 validation.content_invalid / content_malicious
  source not attached; local file remains user intent; expose correction path

503 dependency.* during upload/save/submit
  no false saved/submitted state; local buffer preserved

ambiguous createSubmission transport outcome
  retry same logical command with same Idempotency-Key and semantic If-Match law

exact-content integrity failure
  no partial successful viewer/editor content
```

Final copywriting remains outside P9; these are semantic message intents.

## 7. Access / authority proof

```text
editable Eigenpal presence != Authorization
save button presence != mutation admission guarantee
save status != DRAFT lifecycle authority
selected local file != READY
provider PUT success != READY
READY != WorkingContent
WorkingContent != Submission
Submission gate display != Governance authority
B04 context rail != Document Official authority
no-current-work != historical non-existence
```

Every material command rechecks current server truth.

## 8. P9 closure

```text
material B04 regions/controls traced        18 / 18
unbound material controls                   0
invented operations                         0
operation 87+                               absent
screen-shaped APIs                          0
frontend Authorization evaluator            0
second client DRAFT authority               0
navigation identities unsourced             0
material B04 Screen Contract findings       0
```

P9 is complete for the locked B04 scope.
