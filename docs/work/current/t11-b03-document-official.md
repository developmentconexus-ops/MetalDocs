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
What are people discussing about this stable Document?
Was I explicitly mentioned in that discussion?
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

When colleagues need to discuss the Document itself,
I need a persistent Document-level discussion that survives Revision changes,
so conversation is not confused with DRAFT editorial comments or immutable governance feedback.

When a colleague explicitly mentions me,
I need an in-application notification that routes me back to the exact Document discussion context,
subject to current disclosure and Authorization,
so mentions create useful attention without becoming an access bypass.
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

**Current accepted Product/T3/T6/T8-E authority does not yet admit Document-level discussion, @mentions or in-application Notifications.** Existing forward authority had Notifications deferred in the absence of a named consumer. The operator requirement recorded in §9 now provides that named Launch consumer and therefore requires a bounded upstream reopen before B03 can LOCK.

## 4. Truth hierarchy

```text
stable Document identity
→ current official truth
→ classification/responsibility metadata
→ bounded current work/obsolescence context
→ management affordances
→ Document-level human discussion when the bounded reopen is ratified
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
Document Discussion never becomes WorkingContent, SubmissionFeedback or governance-decision authority
Notification never grants or preserves access to a Document after current disclosure/Authorization denies it
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

legacy published-page CommentsCard / discussion idea
  → strong UX/product evidence for stable-Document discussion;
    do not reuse editor comments or governance feedback as the semantic model

legacy Notifications feature
  → evidence that the interaction was useful, not authority for current routes/DTOs;
    current Launch design must be re-derived from the mention consumer
```

### DEFER / DROP from Launch B03

```text
next periodic review       DEFER Launch+
distribution/coverage      DEFER Launch+
DRAFT/editor comments      remain deferred; not the new Document Discussion
related artifacts          DROP/DEFER
classification/tags        DROP current Launch
public/share link          NOT ADMITTED
user Publish command       DROP; Release is system-owned
email/push/realtime notif.  UNDECIDED; not implied by required in-app notifications
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
↓
Discussão do Documento
  persistent conversation on the stable Document
  @mention interaction
  exact placement/density pending §9 mini-design
```

`Visualizar documento` enters a full B03-owned read-only content surface and provides an explicit Back to ficha control.

### B — Tabbed ficha

```text
Visão geral | Conteúdo | Revisões | Discussão
```

All tabs remain presentation state on the accepted Document Official route. Benefit: compact organization. Risk: important official/work separation and responsibility facts can become hidden behind tabs; “Revisões” may drift into B07 ownership; a Discussion tab can also hide active conversation too aggressively if mentions/unread state matters.

### C — Two-column dossier

```text
left: identity / current Revision / responsibility / work context
right: official-content summary card + actions
lower: revisions / management / discussion
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

## 9. B03-F2 — operator-required Document Discussion + @Mention + in-app Notification

**Operator decision:** Document-level human discussion and in-application notification for explicit `@mention` are required before Launch V1. This is not a speculative future feature.

The current authority gap is real:

```text
T1 explicitly excludes DRAFT EditorialComment from Launch
T3 has no document-discussion Permission
T5/forward obligations deferred Notifications absent a concrete consumer
T6/T8-E have no Document Discussion / Notification operations or schemas
T8-F has no Notification material surface or shell state
```

The operator-provided consumer changes the premise:

```text
stable Document discussion
+ explicit @mention of a current MetalDocs User
→ in-application Notification
→ click returns to exact Document discussion context
→ destination rechecks current disclosure/Authorization
```

Semantic separation required before any implementation design:

```text
Document Discussion != DRAFT/editor comments
Document Discussion != SubmissionFeedback
Document Discussion != GovernanceCase feedback
Notification != access grant
Notification != lifecycle/history authority
```

Current candidate direction, **not yet ratified as upstream contract**:

```text
Discussion belongs to stable Document identity and survives Revision changes
message carries stable author/time/content identity
reply/reference behavior may be admitted in a bounded form
@mention binds stable User identity, not display-text parsing as authority
mention candidate discovery is server-derived and disclosure-safe
mention validation is repeated when the message is accepted
in-app notification is required
unread/read presentation and exact notification lifecycle remain to be designed
B01 may require the smallest evidence-backed reopen for a notification indicator/surface
```

Explicitly NOT frozen yet:

```text
exact operation count / operation numbers
exact Idempotency-Key census delta
new Permission name/bundle mapping
persistence schema
notification semantic owner vs projection ownership boundary
email
push
WebSocket / SSE / polling
realtime presence / typing
framework / library / external service
attachments / reactions
message edit/delete semantics
```

The existing 78-operation / 10-idempotent-create census remains the **current accepted authority until this bounded reopen is designed and ratified**. T11 must not silently invent operation 79+ merely to render the wireframe; after ratification, the census will change only by the exact approved delta.

### Mini-design gate before B03 continues

Pause B03 structural adjudication long enough to close the user-visible semantics that can change B01/B03/backend contracts:

```text
1. read/write eligibility for Discussion
2. message/reply semantics
3. @mention candidate + validation semantics
4. in-app Notification create/read/unread/navigation semantics
5. information-disclosure and offboarding behavior
6. smallest B01 shell impact
7. smallest Product/T3/T5/T6/T8-E/T8-F reopen set
```

Technology selection remains later evidence work after these requirements are frozen. Framework/repository/library choice must not define Product semantics.

## 10. History boundary

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

## 11. Responsive/accessibility structure

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

## 12. P8 next disposition

B03 is temporarily paused at the bounded `B03-F2` mini-design gate because required Discussion/@Mention/Notification semantics can alter B01/B03 structure and current Product/API authority.

After the mini-design is operator-ratified:

```text
consolidate smallest approved upstream reopen
→ update exact operation/Permission/async/frontend authority as required
→ apply smallest evidence-backed B01 reopen if notification chrome is required
→ render revised B03 record-first A/B/C with real Discussion semantics
→ adjudicate B03 visually
→ resolve B03-F1
→ only operator may LOCK B03
```

Do not design B04+ while this material dependency is open.

The previous viewer-first B03 artifact remains rejected evidence and is not a baseline.
