# T11 — B03 Document Official

> **Status:** CANDIDATE / NOT LOCKED.  
> **Block:** B03 — Document Official.  
> **Predecessors:** B01 + B02 are LOCKED / OPERATOR-RATIFIED.  
> **Boundary:** stable Document official/current-management lens only; B04+ remain NOT OPEN.

## 1. User need / mental model

B03 must answer the central MetalDocs question immediately:

```text
What is official right now?
```

Primary jobs:

```text
When I open a document from the Library,
I need the current official content to dominate the page,
so that I can read the company truth without draft noise.

When I need context,
I need to understand code, title, current official Revision, status, Type, Area,
responsible owner and effectivity/release timing,
so that I know exactly what I am reading.

When I am authorized to manage the Document,
I need clear next actions without confusing management with official content,
so that I can continue/open work, start a new Revision, manage responsibility or
initiate/manage obsolescence safely.
```

The official lens never silently becomes DRAFT based on caller identity.

## 2. Bounded accepted authority

Route:

```text
/documents/:document_id
```

Primary read:

```text
getDocument -> DocumentOfficialView
```

Current shape supplies:

```text
stable document code/id
Document Type reference
Area reference
responsible owner reference
DocumentOfficialStatus
ReleasedRevisionView official? when official truth exists
  revision identity / ordinal
  official title
  release_id / released_at
  exact source ContentSummary
  representation = source_only | official_rendition
open_revision? disclosure-safe routing reference
active_obsolescence_request_id? disclosure-safe routing reference
```

Supporting reads/actions already admitted:

```text
getDocumentResponsibleOwner
replaceDocumentResponsibleOwner
createDocumentRevision
getRelease / getReleaseSource
getOfficialRenditionContent
getObsolescenceRequest
createObsolescenceRequest
withdrawObsolescenceRequest
```

The operator-approved T8-E-RO precision additionally makes responsible-owner candidates available only when `document.owner.manage` applies to the exact Document. It adds no operation.

## 3. Truth hierarchy

B03 structure must preserve:

```text
stable Document identity
→ current official truth
→ supporting metadata
→ disclosed work/governance context
→ management actions
```

Key laws:

```text
newer DRAFT/SUBMITTED never replaces older EFFECTIVE content in this route
open_revision is a routing hint, not official truth
active obsolescence request does not make target obsolete before completion
obsolete Document retains its last released official content
before first Release, official content may be absent
viewer output never becomes semantic content authority
```

## 4. Reference study — P6

| Reference | Source observation | Candidate lesson | Mismatch / disconfirming evidence |
|---|---|---|---|
| Veeva Vault Doc Info | Document viewer is primary and sits beside a collapsible/resizable right information pane; header carries identity/state/actions. | Strong evidence for viewer-first split with metadata beside, not above, the document. | Vault exposes broad workflow, favorites, notifications, relationships, sharing and configurable lifecycle machinery outside MetalDocs Launch. |
| M-Files metadata card | Selected document/content is paired with a right metadata card containing title and object properties. | Metadata can remain visible without displacing the document itself. | M-Files metadata card also edits generic workflow/permissions; MetalDocs must not merge Admin or DRAFT authority into the official lens. |
| Qualio effective-vs-work separation | Normal document Library is for effective truth while draft/review/approval work is separately permissioned/workspace-oriented. | Official reader truth should stay clean even when work exists. | Qualio has broader QMS lifecycle/tasks and does not determine MetalDocs field hierarchy. |

Reference URLs:

```text
https://quality.veevavault.help/en/lr/9753/
https://www.userguide.m-files.com/user-guide/latest/eng/metadata_card.html
https://docs.qualio.com/en/articles/6526420-user-permissions
```

## 5. P7 — structural hypotheses

### A — Viewer-first + collapsible right info pane — LEADING

```text
Document header / identity / status
+ disclosed work/obsolescence banner when relevant
+ large official-content viewer
+ compact right pane for metadata + secondary actions
```

Strengths:

```text
reading dominates
metadata stays available
management stays spatially secondary
right pane can collapse for focused reading
works naturally for PDF and read-only DOCX viewer modes
```

### B — Metadata summary first + full-width viewer below

Strength: excellent scan of document facts before content.

Cost: every ordinary reader must pass through metadata before reaching the actual official document; poor for repeated reading.

### C — Full-width viewer + details/action drawer on demand

Strength: maximum content space and simple narrow-screen adaptation.

Cost: important identity/owner/type/area context becomes too hidden for controlled-document confidence and management.

Leading candidate is A.

## 6. P7 lightweight data/action feasibility

| Need | Required truth | Feasibility |
|---|---|---|
| Stable code | `document.code` | PRESENT-IN-AUTHORITY |
| Official title + revision | `official.title` + revision ordinal | PRESENT-IN-AUTHORITY |
| Status | `DocumentOfficialStatus` | PRESENT-IN-AUTHORITY |
| Type / Area / responsible owner | references on `DocumentOfficialView` | PRESENT-IN-AUTHORITY |
| Released/effective timing | `official.released_at` | PRESENT-IN-AUTHORITY |
| Official content mode | source/representation descriptors + referenced exact-byte reads | PRESENT-IN-AUTHORITY |
| Open work banner/routing | `open_revision?` | PRESENT-IN-AUTHORITY |
| Active obsolescence context | `active_obsolescence_request_id?` | PRESENT-IN-AUTHORITY |
| Owner-management selector | T8-E-RO responsible-owner candidates | APPROVED PRECISION / consolidation pending |
| Exact command affordances | actor-safe hints for revision/owner/obsolescence commands | **FINDING B03-F1** |

## 7. B03-F1 — command-affordance source

Frontend authority forbids a client permission matrix. Yet the official page has four material write controls whose visibility/usefulness depends on current server Authorization + relationship/lifecycle predicates:

```text
createDocumentRevision
replaceDocumentResponsibleOwner
createObsolescenceRequest
withdrawObsolescenceRequest
```

Current routing references solve context identity, but do not completely answer whether the actor may issue each command now.

Rejected repairs:

```text
reconstruct Role/Permission logic in React
infer edit authority from open_revision presence/absence
infer obsolescence authority from status alone
show every dangerous command and use 403 as the normal discovery mechanism
invent operation 79
```

Candidate smallest precision:

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
allowed_actions is UX guidance only
values derive from the same canonical T3 permission/scope + Controlled Documents predicates
used by command authorization, or a provably shared equivalent
every command rechecks current canonical truth
absence grants nothing and proves only current hint result
```

This mirrors the already-accepted `GovernanceCaseView.allowed_actions` pattern and adds no operation/owner/Permission/route.

B03 cannot LOCK while B03-F1 remains unresolved.

## 8. Official-content presentation candidates

```text
SourceOnly PDF
→ exact Release source is primary viewer content

SourceOnly DOCX
→ accepted read-only DOCX adapter on exact Release source

RequireOfficialRendition(PDF)
→ exact OfficialRendition PDF is primary viewer content
→ exact Release source separately available/labeled

no official Release yet
→ no fake preview; explicit "Ainda não existe versão oficial" structural state
```

Viewer controls may provide ordinary reading mechanics (page/zoom/download where technically supported) but may not invent lifecycle actions or alter exact-content authority.

## 9. Candidate hierarchy

Leading A candidate:

```text
breadcrumb back to Biblioteca
↓
code + official title + official status/revision
↓
context banner only when server discloses open work / active obsolescence
↓
main split
  70–75% official viewer
  25–30% collapsible information pane
↓
information pane
  Tipo
  Área
  Responsável
  Revisão oficial
  Vigente/liberado em
  representação/source semantics
  link to Histórico (route only; B07 owns structure)
↓
management actions
  only from server-returned hints
```

Dangerous obsolescence action should not visually compete with ordinary reading or the primary work continuation action.

## 10. Responsive/accessibility candidate

```text
desktop
  viewer + right pane

narrow
  viewer primary
  metadata pane becomes inline/collapsible section below header
  management actions remain keyboard reachable
```

The viewer cannot be the only source of title/status understanding for screen readers. Header/metadata remain semantic HTML. Pane collapse has explicit accessible control and focus return.

## 11. P8 disposition

Render B03 only, comparing A/B/C. Do not design Document Work content/editor, Governance Case, History timeline or Admin screens inside this artifact.

Operator visual adjudication is required before B03 LOCK.
