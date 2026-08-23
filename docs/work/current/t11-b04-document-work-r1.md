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

The P7 wireframe represents Eigenpal toolbar zones structurally; it does not freeze vendor button order, visual styling or implementation internals.

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

## 5. Current P7 interaction/state findings

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
  refetch current server DRAFT
  require explicit reconciliation
  no silent LWW / automatic merge

state.upload_expired
  preserve intended local bytes
  allocate a fresh upload
  never revive the expired upload id
```

## 6. Open material decision before P8

### B04-F1 — DOCX persistence UX

Still **OPEN**:

```text
manual save
vs
autosave
vs
hybrid save
```

The final choice must preserve:

```text
DocumentWorkView + strong DRAFT ETag = server truth
local editor buffer = FORM DRAFT only
save status is presentation state, not Product lifecycle
412 preserves local input and forces reconciliation
submit must never race ahead of unresolved local edits
```

Legacy evidence used autosave, but legacy timing/mechanics are evidence only and do not define current Product semantics.

## 7. P8 gate

```text
P7 macro-layout                         OPERATOR-APPROVED
Eigenpal DOCX / viewer format boundary  OPERATOR-APPROVED
right-rail composition                  OPERATOR-APPROVED
B04-F1 persistence UX                   OPEN
P8 functional HTML                      NOT YET GENERATED
B04 LOCK                                NO
```

Next:

```text
adjudicate B04-F1
→ render one functional P8
→ operator operates / iterates
→ operator-only LOCK
→ P9 exact Screen Contract
→ P10 bounded pattern consolidation
```

B05+ remain NOT OPEN.
