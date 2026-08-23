# T11 — B04 Document Work R1 — Method v2.2 candidate

> **Status:** CURRENT FP1 BLOCK / CANDIDATE / NOT LOCKED.  
> **Block:** B04 — Document Work / Authoring.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Predecessors:** B01 / B01N / B02 / B03 LOCKED.  
> **Implementation:** BLOCKED.

## 1. Current Product/architecture boundary

Stable route:

```text
/documents/:document_id/work
```

Route resolution law:

```text
getDocument
→ disclosed open_revision
→ DRAFT      = editable Work lens
→ SUBMITTED  = read-only submitted Work lens
→ absent     = no current work available; no History fallback
```

Document Work owns no official truth and no Governance Case truth.

Current operations used by the block remain the accepted Work family:

```text
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

No new operation is admitted by this planning block.

## 2. Operator-approved P7 macro-layout

The operator reviewed the final B04 low-fidelity layout after explicit legacy comparison and **APPROVED** it.

Selected composition:

```text
MetalDocs minimal Work header
  code + title
  REV / current Work state
  current official Revision reference
  visible persistence state
↓
CONTENT-FIRST WORKSPACE
  main canvas
    DOCX DRAFT     → Eigenpal editor surface
                     Eigenpal toolbar/chrome + editable document canvas
    PDF DRAFT      → read-only viewer
    SUBMITTED      → read-only exact submitted content viewer

  right operational rail
    1. Trabalho atual
    2. Fonte
    3. Ações
    4. Contexto do documento — collapsed by default
```

## 3. Eigenpal boundary

For Launch B04:

```text
DOCX DRAFT
  only currently editable content format
  Eigenpal owns internal document-editing ergonomics / toolbar / canvas behavior
  MetalDocs owns placement, Work semantics, persistence orchestration and Product state presentation

PDF DRAFT
  read-only inspect
  replace source when current command truth admits it

SUBMITTED
  read-only
  never mounts editable DRAFT controls
```

The P7/P8 wireframes represent Eigenpal toolbar zones structurally; they do not freeze vendor button order, visual styling or implementation internals.

## 4. Right-rail disposition after legacy review

Legacy Work sidebar evidence contained identity, revision lineage, approval chain/timeline, contextual approval panels, decision footer and ficha navigation.

Current disposition:

```text
KEEP / ADAPT
  current Work state
  source file/format/size where useful
  persistence/save state
  Work actions
  compact collapsed Document context
  explicit return to B03 ficha

DO NOT RESTORE
  full revision history          → B07
  approval chain/timeline        → B06
  approval decision footer       → B06
  legacy mode-adaptive Work/Approval owner
  legacy editorial-comments semantics
  duplicated full B03 ficha
```

Collapsed Document context may show only bounded orientation facts already available through the route-entry/current read composition:

```text
Document Type
Area
Responsible owner
current official Revision/status
Ver ficha completa → B03
```

It is orientation, not a second Document Official authority.

## 5. P7/P8 interaction-state laws

Must remain visible/understandable in P8:

```text
DRAFT vs SUBMITTED distinction
DOCX editable vs PDF read-only distinction
current official Revision remains separate from current Work Revision
source replacement progression:
  local file chosen
  → provider upload
  → server verification/admission
  → attach under current DRAFT ETag
  → authoritative WorkingContent updated

provider PUT success != READY
READY != WorkingContent
WorkingContent != Submission

412 precondition.draft_changed
  preserve unsaved/local human input
  stop autosave progression
  refetch current server DRAFT
  require explicit reconciliation
  no silent LWW / automatic merge

state.upload_expired
  preserve intended local bytes
  allocate a fresh upload
  never revive the expired upload id
```

## 6. B04-F1 — DOCX persistence UX — CLOSED / OPERATOR-RATIFIED

The operator approved the **hybrid persistence** model.

Contract:

```text
local Eigenpal edit
→ FORM DRAFT becomes DIRTY immediately
→ after an implementation-appropriate quiet period, coalesce background persistence
→ at most one save pipeline in flight
→ additional edits while saving remain DIRTY and schedule the next coalesced save

Salvar agora / Ctrl+S
→ force the same save pipeline immediately

save pipeline for DOCX body
→ serialize intended DOCX bytes
→ allocate/upload/complete admission as required
→ PATCH exact DRAFT under current If-Match
→ accept returned DocumentWorkView + new strong ETag
→ mark clean only for the exact local generation persisted
```

Timing/debounce duration is implementation tuning, not Product/frontend authority.

Visible presentation states:

```text
saved
local changes / dirty
saving
save failed — local buffer preserved
conflict / reconciliation required
```

Concurrency/error law:

```text
412 precondition.draft_changed
→ stop normal autosave for that stale base
→ preserve local human input
→ refetch authoritative DRAFT
→ explicit human reconciliation
→ no automatic merge / silent overwrite

save/dependency failure
→ preserve local buffer
→ expose retry
→ no false "saved" state
```

Submit law:

```text
Enviar para análise
→ if a save is in flight, wait for it to resolve
→ if local changes remain, force flush through the same save pipeline
→ require no unresolved save error/conflict/upload attachment
→ only then createSubmission against the exact current accepted DRAFT ETag
```

Browser state law:

```text
local editor buffer = transient FORM DRAFT only
no IndexedDB/localStorage/offline durable DRAFT baseline
no second client content authority
```

## 7. Canonical P8 R1

Current functional evidence:

```text
t11-b04-document-work-functional-wireframe.html
```

Medium:

```text
HTML
CSS
vanilla JavaScript
local deterministic fixtures/state simulation
```

P8 R1 exercises:

```text
DOCX Eigenpal-like toolbar + editable canvas
DIRTY → coalesced autosave → saved/new ETag
Salvar agora / Ctrl+S
one-save-in-flight behavior
save failure + explicit retry opportunity
412 conflict + local-buffer preservation + manual reconciliation choice
PDF DRAFT read-only viewer
source replacement:
  allocation → provider PUT → completion/READY → DRAFT attach
expired upload → preserve local file → new allocation / re-upload
submit forces flush before SUBMITTED
SUBMITTED read-only gates
withdraw Submission → same Revision back to DRAFT
cancel Revision with mandatory reason → no current work
no-History fallback + explicit return to B03
collapsed Document context
locked B01N notification chrome reuse
responsive rail stacking
```

Review-only fixture controls may force conflict/failure/expiry states. They are P8 Evidence controls, not Product UI.

## 8. Current gate

```text
P7 macro-layout                         OPERATOR-APPROVED
Eigenpal DOCX / viewer format boundary  OPERATOR-APPROVED
right-rail composition                  OPERATOR-APPROVED
B04-F1 hybrid persistence UX            CLOSED / OPERATOR-RATIFIED
P8 functional R1                        RENDERED / OPERATOR OPERATION+REVIEW
B04 LOCK                                NO
```

Next:

```text
operator operates P8 R1
→ iterate only material B04 layout/interaction findings
→ operator-only B04 LOCK
→ P9 exact Screen Contract
→ P10 bounded pattern consolidation
```

B05+ remain NOT OPEN.
