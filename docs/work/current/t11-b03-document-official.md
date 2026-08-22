# T11 — B03 Document Official

> **Status:** CANDIDATE / NOT LOCKED.  
> **Block:** B03 — Document Official / Ficha do Documento.  
> **Predecessors:** B01 + B02 are LOCKED / OPERATOR-RATIFIED.  
> **Legacy evidence:** `docs/work/current/t11-legacy-frontend-evidence.md` — EVIDENCE / NOT AUTHORITY.  
> **Boundary:** stable Document official/current-management record plus deliberate official-content viewing surface; B04+ remain NOT OPEN.

## 1. Correction from operator + legacy evidence

The first B03 candidate made the official-content viewer the page itself. Direct operator feedback rejected that collapse: entering a Document record and viewing its exact content are distinct user intentions.

The pre-reset frontend independently supports that distinction:

```text
Library
→ content workspace / viewer / editor context
→ full Document detail/ficha
```

The old workspace even linked explicitly to `Ver ficha completa do documento`.

Decision for current B03:

```text
first viewer-first whole-page hypothesis  REJECTED / TOO COLLAPSED
Document Official route                   record/ficha first
Official content viewer                   distinct material surface entered deliberately
Document Work                             remains B04 / separate accepted route
Governance Case                           remains B06 / separate accepted route
Full Document History                     remains B07 / separate accepted route
```

No legacy route is restored by this correction.

## 2. User needs

B03 must answer, in order:

```text
What Document is this?
What is official right now?
Who is responsible and where does it belong?
What official Revision/content exists?
Is newer work underway without replacing official truth?
What can I safely do next?
How do I deliberately read the exact official content?
Where do I go for full history?
```

Primary jobs:

```text
When I open a result from the Library,
I need a trustworthy ficha of the controlled Document,
so I can understand identity, responsibility and official state before acting.

When I want to read the actual official content,
I need a deliberate View Document action,
so the reading surface can maximize content without pretending to be the Document record.

When newer work exists,
I need to see that it is separate from the official Revision,
so DRAFT/SUBMITTED never overwrites my understanding of current official truth.
```

## 3. Bounded current authority

Stable route:

```text
/documents/:document_id
```

Primary read:

```text
getDocument -> DocumentOfficialView
```

Current official shape supplies:

```text
stable Document id + code
Document Type reference
Area reference
responsible owner reference
DocumentOfficialStatus
official? ReleasedRevisionView
  revision identity / ordinal
  governed official title
  release_id + released_at
  exact source ContentSummary
  source_only | official_rendition representation
open_revision? disclosure-safe routing reference
active_obsolescence_request_id? disclosure-safe routing reference
```

Supporting current operations remain:

```text
getDocumentResponsibleOwner
replaceDocumentResponsibleOwner
createDocumentRevision
getDocumentHistory                 // B07 owns full composition; supporting use must not derive current state
getRelease / getReleaseSource
getOfficialRenditionContent
getObsolescenceRequest
createObsolescenceRequest
withdrawObsolescenceRequest
```

## 4. Truth hierarchy

```text
stable Document identity
→ current official truth
→ classification/responsibility metadata
→ bounded current work/obsolescence context
→ management affordances
→ historical evidence only as historical evidence
```

Laws:

```text
newer DRAFT/SUBMITTED never replaces older EFFECTIVE content in B03
open_revision is work routing, not official truth
active obsolescence does not make the target obsolete before completion
obsolete retains its last released official content
before first Release, official content may be absent
History/Audit never becomes current-state authority
viewer output never becomes semantic content authority
```

## 5. Legacy information disposition applied to B03

### KEEP / current authority already supports

```text
code
official governed title
status
current official Revision
Document Type
Area
responsible owner
released/effective date
source / official representation
format + file size where useful
explicit open-work context
```

### ADAPT, do not copy literally

```text
legacy revision lineage
  → B03 shows bounded current-revision context + link to full History;
    B07 owns the full timeline

legacy approval/sign-off chain
  → valuable accountability evidence, but current immutable governance facts live in History;
    do not resurrect Approval as an owner or put a generic workflow widget on the ficha

legacy content viewer
  → distinct B03 material viewing surface, not the ficha itself
```

### DEFER / DROP from Launch B03

```text
next periodic review       DEFER Launch+
distribution/coverage      DEFER Launch+
comments                   DROP current Launch
related artifacts          DROP/DEFER
classification/tags        DROP current Launch
public/share link          NOT ADMITTED
user Publish command       DROP; Release is system-owned
```

## 6. Candidate B03 record hierarchy

Leading design question is now record composition, not viewer composition.

### A — Sectioned ficha + explicit full viewer — LEADING

```text
breadcrumb / back to Biblioteca
↓
Document hero
  code
  official title or truthful no-official-title state
  status
  official Revision
  primary: Visualizar documento when official content exists
  secondary: Baixar / Histórico as supported
↓
current-context banner
  open Revision when disclosed
  active obsolescence when disclosed
↓
Ficha / Sobre
  Tipo
  Área
  Responsável
  Revision oficial
  Liberado em
  source/official-representation label
  format / size where useful
↓
Revisões
  current official Revision
  open Revision if disclosed
  explicit Full History entry
  no invented all-revisions snapshot
↓
Management
  actions from server-derived hints only
```

`Visualizar documento` enters a full B03-owned read-only content surface and provides an explicit Back to ficha control.

### B — Tabbed ficha

```text
Visão geral | Conteúdo | Revisões
```

All tabs remain presentation state on the accepted Document Official route. Benefit: compact organization. Risk: important official/work separation and responsibility facts can become hidden behind tabs; “Revisões” may drift into B07 ownership.

### C — Two-column dossier

```text
left: identity / current Revision / responsibility / work context
right: official-content summary card + actions
lower: revisions / management
```

Benefit: dense professional scan. Risk: looks dashboard-like and can over-prioritize metadata cards over the Document record narrative.

## 7. Distinct official-content viewer surface

This is not B04.

```text
entry = explicit Visualizar documento from ficha
mode = read-only
```

Current presentation laws:

```text
SourceOnly PDF
  → exact Release source in PDF viewer

SourceOnly DOCX
  → accepted read-only DOCX adapter on exact Release source

RequireOfficialRendition(PDF)
  → exact OfficialRendition PDF is primary official view
  → exact Release source remains separately available/labeled

no official Release
  → no fake viewer / no draft substitution
```

Candidate viewer shell:

```text
Back to ficha
Document code + official Revision
exact-content label
page/zoom mechanics where technically supported
Download exact official/source resource where admitted
large content canvas
```

No edit/submit/governance controls appear in this viewer.

## 8. B03-F1 — command-affordance source

Frontend must not rebuild T3 Authorization. Candidate smallest precision remains:

```text
DocumentOfficialAction =
  create_revision
  replace_responsible_owner
  create_obsolescence_request
  withdraw_obsolescence_request

DocumentOfficialView.allowed_actions: unique DocumentOfficialAction[]
```

Law:

```text
server-derived UX hints only
same canonical authorization/domain predicates or provably shared equivalent
every command rechecks current truth
no operation / owner / Permission / route added
```

This follows the already accepted `GovernanceCaseView.allowed_actions` pattern.

B03 cannot LOCK until the precision is reconciled into the effective current read authority or otherwise proven unnecessary.

## 9. History boundary

The legacy ficha showed a full lineage and sign-off chain. Current architecture deliberately owns full semantic history at:

```text
/documents/:document_id/history
getDocumentHistory -> DocumentHistoryPage
```

B03 may show only information that does not create a second history authority:

```text
current official Revision from DocumentOfficialView
current disclosed open Revision from open_revision
release timestamp / representation from official
entry/link to full History
```

Any richer inline historical summary must trace to `getDocumentHistory`, remain historical-only, tolerate 403/non-disclosure, and be proven not to duplicate B07 composition before it becomes part of the locked B03 baseline.

## 10. Responsive/accessibility structure

```text
desktop ficha
  sectioned readable record with clear primary action

narrow ficha
  single-column semantic order
  no loss of official/work distinction

viewer
  content-first, focus-managed entry/exit
  exact title/revision also exposed as semantic text outside rendered document bytes
```

No critical context depends on color, hover or the document renderer itself.

## 11. P8 next disposition

Render B03 record-first A/B/C plus the distinct read-only viewer transition.

Do not design:

```text
Document Work editor/content mutation
Governance Case decision UI
full History timeline
Audit
Admin
```

The previous viewer-first B03 artifact remains rejected evidence and is not a baseline.
