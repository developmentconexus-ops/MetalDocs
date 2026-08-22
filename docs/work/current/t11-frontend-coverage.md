# T11 — MetalDocs Frontend Coverage Matrix

> **TEMPORARY T11 CANDIDATE WORK / BRANCH-ONLY.** Derived from accepted T6/T8-E/T8-F authority under the T11 frontend-readiness method. It is implementation-planning Evidence, not Product/API/frontend authority. Material gaps route to the smallest owning authority; they are never repaired here by inventing semantics.

## 1. Goal

Prove that the future frontend represents the already-accepted MetalDocs and that the T11 implementation graph can actually deliver each browser surface against its real backend prerequisites.

This matrix is derived **before wireframes**.

Legend:

```text
COVERED       accepted authority supplies a coherent frontend home/contract
WIREFRAME     authority is coherent but material state/action must still be drawn/reviewed
FINDING       T11 graph/readiness coherence issue requires adjudication
```

## 2. Fixed frontend envelope

```text
application operations                78
operation 79                          absent
stable Product routes                 exact T6/T8-F set
frontend semantic owner               none
frontend Authorization engine         absent
parallel DTO/API authority             absent
parallel global server-truth store     absent
server state                          TanStack Query
navigation/filter context             router / URL
unaccepted form input                 local feature/form state
ephemeral presentation state          local React state
generated TypeScript shapes           wire authority
```

All 78 application operations require `MetalDocsSession`. Unsafe mutations additionally obey exact T8-E CSRF/conditional/idempotency profiles. OIDC `/auth/login` and `/auth/callback` remain browser integration routes outside the 78-operation census.

## 3. Accepted human-goal coverage — 16/16

| # | Accepted human goal | Primary frontend home | Accepted backend/wire input | Material interaction obligation | Status |
|---:|---|---|---|---|---|
| 1 | Establish/end session | application shell + `/auth/login`/`/auth/callback` | `getSession`, `endSession`; OIDC browser routes; `SessionView` | no password input; HttpOnly session; CSRF material; real 401/end-session behavior; server is current access authority | WIREFRAME |
| 2 | Discover official documents | `/documents` Library | `listDocuments` → `DocumentPage` | official/effective discovery; admitted q/filters/cursor; no DRAFT/SUBMITTED as official; navigation by returned `document_id` | WIREFRAME |
| 3 | Create Document | `/documents` creation interaction | `getDocumentCreationOptions`, supporting `getDocumentTypeNumberingPreview`, `createDocument` | options guidance only; preview never reserves; idempotent logical create; result routes to returned Document identity | WIREFRAME |
| 4 | Inspect official/current Document truth | `/documents/:document_id` | `getDocument`; supporting owner/Release/rendition/obsolescence reads as they become reachable | official truth primary; DRAFT never rendered as official; progressive accepted lens enrichment is explicit | WIREFRAME |
| 5 | Start/enter open Revision | Document Official → Work | `getDocument` disclosure-safe `open_revision`; `createDocumentRevision` | server supplies open Revision identity; create-next and Work target become available in same implementation tranche | WIREFRAME |
| 6 | Author DRAFT | `/documents/:document_id/work` | `getRevision`, `getRevisionDraft`, `getRevisionDraftSource`, `updateRevisionDraft` | exact DRAFT ETag; explicit 412 reconciliation; no silent merge/LWW | WIREFRAME |
| 7 | Upload/replace DRAFT source | Document Work | `startRevisionDraftUpload`, provider PUT, `completeRevisionDraftUpload`, `updateRevisionDraft` | PUT != READY != WorkingContent; expired capability replaced, never revived; admission stays server authority | WIREFRAME |
| 8 | Submit / withdraw / cancel | Document Work | `createSubmission`, `getSubmission`, `getSubmissionSource`, `withdrawSubmission`, `cancelRevision` | Submission idempotency + DRAFT condition; immutable subject; exact withdrawal/cancel outcomes | WIREFRAME |
| 9 | See actor-relevant work | `/work` | `listAuthoringWork`, `listGovernanceWork` | projection only; each lane is implemented only when its owner target surface exists | WIREFRAME |
| 10 | Participate in governance | `/work/governance/:attempt_id` | governance attempt/feedback/decision operations + admitted exact source reads | immutable governed subject; current AuthZ; `allowed_actions` hints only; decision != publish | WIREFRAME |
| 11 | Inspect Document history | `/documents/:document_id/history` | `getDocumentHistory` + admitted supporting facts/content | semantic history only; not current-resource resolver; surface is regression-enriched as later fact classes become reachable | WIREFRAME |
| 12 | Initiate/manage obsolescence | Document Official | create/get/withdraw obsolescence operations | active request identity from current server truth; synchronous NoHumanApproval handled without fake Step | WIREFRAME |
| 13 | Administer Organization | `/admin/organization` | provider-subject search + Company/User/Profile/Binding/Eligibility/Area/Group operations | independent ETag domains; opaque provider ref; atomic offboarding; no provider claim Product truth | WIREFRAME |
| 14 | Administer access | `/admin/access` | GroupMembership + Role operations, supported by admitted User/Group/Area reads | fixed roles/Permissions; no custom editor; security-bearing membership and RoleAssignment remain server authority | WIREFRAME |
| 15 | Administer document governance | `/admin/document-governance` | DocumentType/governance/eligibility/numbering/config + concrete Document template-role reads/writes when Documents exist | config ETags; Template not peer owner; route is explicitly enriched after Document creation exists | WIREFRAME |
| 16 | Inspect Audit | `/audit` | `listAuditEvents` → `AuditEventPage` | paging only; evidence not current truth; no inferred filters | WIREFRAME |

No accepted human goal lacks a semantic frontend home. No current row requires a Product/T8-F semantic reopen.

## 4. Exact 78-operation frontend reconciliation from T8-F

T8-F's accepted frontend consumer partition remains unchanged:

```text
Shell/session                         operations 1–2                  = 2
Admin / Organization                 operations 3–26                 = 24
Admin / Access                       operations 27–33                = 7
Admin / Document Governance          operations 34–43, 50–51         = 12
Library / creation                   operations 44–46                = 3
Document Official / management       operations 47–49,52,72–73,75,77 = 8
My Work / History supporting reads   operations 53–56,76             = 5
Document Work                        operations 57–66                = 10
Governance Case                      operations 67–71                = 5
Official rendition                   operation 74                    = 1
Audit                                operation 78                    = 1

TOTAL                                                              = 78
```

Primary frontend home does not transfer semantic ownership. Supporting reads may cross lenses only under accepted disclosure and never create duplicate client truth.

## 5. Cross-cutting architecture → UX obligations

| Accepted invariant | Required frontend behavior | Status |
|---|---|---|
| All 78 operations require current session | real AuthN/session + canonical AuthZ must precede protected semantic browser proof | COVERED by corrected T11 graph |
| Server is Authorization authority | visibility may improve UX but hidden UI is never correctness | WIREFRAME |
| `SessionView` has no effective Permission snapshot | no client role/permission navigation matrix | WIREFRAME |
| One semantic owner per meaning | composed screen may read several owners; each write goes to exactly one accepted operation | WIREFRAME |
| TanStack Query owns server state | no durable normalized Product entity/global store | COVERED |
| Forms bind to exact ETag | stale whole-replacement/DRAFT states need explicit safe reconciliation | WIREFRAME |
| 10 idempotent POST creations | same logical retry keeps key; changed/new command gets new key | WIREFRAME |
| Branch on `Problem.code` | materially different safe-next-action failures cannot collapse into generic toast | WIREFRAME |
| Exact bytes have semantic URLs | provider location never Product identity; integrity failure precedes success presentation | WIREFRAME |
| Provider PUT != admitted content | upload UI cannot claim WorkingContent before completion/admission + DRAFT attachment | WIREFRAME |
| `allowed_actions` are hints | command always rechecks canonical truth | COVERED |
| Cursor paging | no offset/page-number/total-count contract | WIREFRAME |
| Disclosure-safe current references | work/obsolescence navigation uses returned current references, not History/Audit inference | WIREFRAME |
| Release system-owned | no Publish button/public Release mutation | COVERED |
| Audit != History != current truth | cross-links may exist only through admitted facts; no merged lifecycle authority | WIREFRAME |
| T8-B default-deny import graph | screen organization never justifies new backend dependency edge | COVERED |
| Hard no screen-shaped API | visual convenience cannot authorize operation 79/new read | COVERED |

## 6. Coverage findings and adjudication

### COV-01 — Protected HTTP ordering cycle — MATERIAL / CORRECTED

Initial T11 ordered `Organization → Authentication + Authorization` while every application operation requires `MetalDocsSession` and S1 claimed real E3/E4 proof.

Accepted T3/T10 already provide explicit non-serving bootstrap/recovery concern; T10 assigns its exact runtime realization to T11 planning.

**Correction:** P3 owns fenced non-serving bootstrap realization; first semantic tranche combines Identity + Organization + Access, establishing real AuthN/AuthZ before all later protected Product work.

### COV-02 — Artificial Organization/Access frontend split — MATERIAL / CORRECTED

Admin Organization uses Authentication-owned provider-subject lookup; Admin Access needs current Organization identity/reference truth. The old split created cross-tranche incomplete screens.

**Correction:** operations 1–33 close together in S1. Semantic owners remain separate internally; the tranche is vertical implementation, not merged authority.

### COV-03 — Stable route incorrectly treated as single completion moment — MATERIAL / CORRECTED

Document Official, Admin Document Governance, My Work and History acquire admitted capability at different later slices.

**Correction:** Node Completion Contracts now specify progressive route/lens enrichment explicitly and forbid claiming future regions/actions early.

### COV-04 — My Work governance could expose a dead target — MATERIAL / CORRECTED

`listGovernanceWork` targets Governance Case loaded by `getGovernanceAttempt`.

**Correction:** operation 55 moves to S5 with operations 67–74 so governance projection and real target close together.

### COV-05 — Create-next Revision / authoring work could expose dead Document Work target — MATERIAL / CORRECTED

A second dependency pass found the same issue for `createDocumentRevision` and `listAuthoringWork`: their safe user consequence is entry into the real Document Work surface.

**Correction:** operation 52 and operation 54 move to S4 with operations 56–66. S3 Document Official explicitly does not claim the create/enter-Revision affordance before S4.

## 7. Corrected T11 implementation tranche partition

The smallest coherent partition after COV-01→05 is:

```text
S1  Identity + Organization + Access
    operations 1–33                                      = 33
    + /auth/login + /auth/callback outside census

S2  Document Governance base configuration
    operations 34–43                                     = 10

S3  Library + Document core + template-role + History
    operations 44–51 + 53                                 = 9

S4  Revision authoring + My Work authoring + content + Submission
    operation 52 + operation 54 + operations 56–66       = 13

S5  Governance work + Governance Case + Release/rendition
    operation 55 + operations 67–74                       = 9

S6  Obsolescence + Audit
    operations 75–78                                      = 4

TOTAL                                                     = 78
```

Count proof:

```text
33 + 10 + 9 + 13 + 9 + 4 = 78
```

Non-contiguous assignments deliberately prevent dead user paths:

```text
50–51 concrete Document template-role
  → S3 after ordinary Document creation exists; frontend home remains Admin Document Governance

52 createDocumentRevision + 54 listAuthoringWork
  → S4 with live Document Work target

55 listGovernanceWork
  → S5 with live Governance Case target
```

No Product/API meaning changes. No operation is added or removed.

## 8. Progressive lens implementation map

```text
/admin/document-governance
  S2 base DocumentType/governance/eligible-template/numbering/config
  → S3 concrete Document template-role administration

/documents/:document_id
  S3 official/core + responsible-owner
  → S4 create/enter Revision
  → S5 Release/source/OfficialRendition presentation
  → S6 obsolescence state/actions

/work
  S4 authoring projection + live Work target
  → S5 governance projection + live Governance Case target

/documents/:document_id/history
  S3 reachable semantic history
  → S4/S5/S6 regression/enrichment as additional accepted facts become reachable
```

This is one stable T8-F route set, not duplicate routes or a late second frontend phase.

## 9. Coverage Matrix closure status

```text
accepted human goals mapped                 16 / 16
stable route/lens homes                     covered
T8-F operation consumer reconciliation      78 / 78
T11 implementation rows                     78 / 78 exactly once
operation 79                                absent
material T11 coverage findings              5 found / 5 adjudicated
unresolved MATERIAL coverage finding        0
Product/T8-F semantic reopen                not justified by current evidence
wireframes                                  not started
```

F1 is complete. The next frontend-readiness step is **F2 material interaction-surface inventory**, followed by Screen Contracts; wireframing remains intentionally later.