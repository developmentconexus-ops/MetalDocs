# T11 — B04 Document Work R1 — Method v2.2 locked

> **Status:** LOCKED / OPERATOR-RATIFIED.  
> **Block:** B04 — Document Work / Authoring.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Predecessors:** B01 / B01N / B02 / B03 LOCKED.  
> **Implementation:** BLOCKED.

## 1. Lock basis

The operator explicitly approved, in sequence:

```text
P7 content-first macro-layout
Eigenpal DOCX editable boundary
PDF / SUBMITTED read-only boundary
right operational rail
B04-F1 hybrid persistence UX
functional low-fi P8 R1 after browser operation/review
```

No material B04 interaction/layout finding remains open at the operator gate.

## 2. Locked Product/architecture boundary

Stable route:

```text
/documents/:document_id/work
```

Route resolution law:

```text
getDocument
→ disclosed open_revision
→ DRAFT      = editable/read-only-by-format Work lens
→ SUBMITTED  = read-only submitted Work lens
→ absent     = no current work available; no History fallback
```

Document Work owns no official truth, History truth, Governance Case truth or frontend Authorization authority.

Current operation family consumed by B04:

```text
getDocument
getRevision
getRevisionDraft
updateRevisionDraft
startRevisionDraftUpload
completeRevisionDraftUpload
getRevisionDraftSource
createSubmission
getSubmission
getSubmissionSource
withdrawSubmission
cancelRevision
```

No new application operation was created by B04 planning.

## 3. Locked composition

```text
MetalDocs minimal Work header
  code + title
  Work Revision / state
  current official Revision reference
  visible persistence state
↓
CONTENT-FIRST WORKSPACE
  main canvas
    DOCX DRAFT     → Eigenpal toolbar/chrome + editable document canvas
    PDF DRAFT      → read-only exact source viewer
    SUBMITTED      → read-only exact submitted-content viewer

  right operational rail
    1. Trabalho atual
    2. Fonte
    3. Ações
    4. Contexto do documento — collapsed by default
```

The full B03 ficha, B06 governance timeline/actions and B07 revision history remain outside B04.

## 4. Eigenpal / viewer boundary

```text
DOCX DRAFT
  Eigenpal owns internal document-editing ergonomics
  MetalDocs owns Work placement, persistence orchestration and Product-state presentation

PDF DRAFT
  read-only inspect
  source replacement is the content-change path

SUBMITTED
  read-only immutable Submission content
  no editable DRAFT controls
```

P8 toolbar controls are structural stand-ins only; vendor button order/style is not locked Product authority.

## 5. B04-F1 — hybrid persistence — CLOSED / OPERATOR-RATIFIED

```text
local Eigenpal edit
→ FORM DRAFT becomes DIRTY immediately
→ coalesced background persistence after an implementation-appropriate quiet period
→ at most one save pipeline in flight
→ edits arriving during a save remain DIRTY for the next coalesced save

Salvar agora / Ctrl+S
→ force the same persistence pipeline

DOCX body persistence
→ serialize intended bytes
→ allocate/upload/complete admission as needed
→ PATCH exact DRAFT using current If-Match
→ accept authoritative DocumentWorkView + new strong ETag
→ mark clean only for the exact local generation accepted
```

No fixed debounce duration is Product authority.

Visible persistence intents:

```text
saved
dirty / local changes
saving
save failed — local buffer preserved
conflict — reconciliation required
```

### Failure/concurrency law

```text
412 precondition.draft_changed
→ stop stale autosave progression
→ preserve local human input
→ refetch authoritative DRAFT
→ explicit human reconciliation
→ no automatic merge / no silent LWW

save/dependency failure
→ preserve local buffer
→ explicit retry opportunity
→ never show false saved state
```

### Submit law

```text
Enviar para análise
→ wait for any in-flight save
→ force-flush remaining local changes
→ require no unresolved save error/conflict/upload attachment
→ createSubmission only against the exact accepted DRAFT ETag
```

Browser local state remains transient FORM DRAFT only; no IndexedDB/localStorage/offline durable DRAFT baseline is introduced.

## 6. Exact-content replacement law represented in the lock

```text
local file chosen
→ startRevisionDraftUpload
→ provider PUT
→ completeRevisionDraftUpload establishes READY only after server admission
→ updateRevisionDraft attaches source_upload_id under current DRAFT ETag
→ authoritative WorkingContent + new ETag
```

Truth ladder:

```text
provider PUT success != READY
READY != WorkingContent
WorkingContent != Submission
```

Expired upload:

```text
preserve intended local bytes
→ allocate a fresh upload
→ repeat upload/admission
→ never revive the expired upload_id
```

## 7. Locked P8 R1 evidence

Canonical functional evidence:

```text
docs/work/current/t11-b04-document-work-functional-wireframe.html
```

The operator operated/reviewed the browser-functional prototype and explicitly approved it.

P8 locally exercises:

```text
DOCX edit → DIRTY → autosave/save-now → new ETag
coalesced edits while one save is in flight
save failure with local-buffer preservation
412 conflict + explicit reconciliation / no auto-merge
PDF DRAFT read-only viewer
source replacement allocation → PUT → READY → attach
expired upload → fresh allocation with same intended local bytes
submit force-flush → SUBMITTED read-only
withdraw Submission → same Revision DRAFT
cancel Revision reason → no-current-work state
return to B03; no History fallback
collapsed Document context
B01N global chrome reuse
responsive rail stacking
```

Review-only failure/conflict/expiry controls are Evidence, not Product UI.

## 8. State authority

```text
SERVER STATE
  DocumentOfficialView routing context
  RevisionView / DocumentWorkView + DRAFT ETag
  SubmissionView

NAVIGATION / URL
  document_id + stable Work route

FORM DRAFT
  title + Eigenpal local buffer + selected replacement file before acceptance

EPHEMERAL UI
  context disclosure, upload/save progress presentation,
  conflict/reconciliation dialog, cancellation dialog
```

No fifth durable/global frontend state class is created.

## 9. Post-LOCK progression

```text
B04 LOCK                              COMPLETE / OPERATOR-RATIFIED
P9 exact Screen Contract              NEXT
P10 bounded pattern consolidation     after P9
B05                                   NOT OPEN until P9/P10 close
```

Implementation remains blocked by the repository-wide T11/T12/Whole-R10 gate.
