# T11 — MetalDocs Frontend Coverage Matrix

> **TEMPORARY T11 CANDIDATE WORK / BRANCH-ONLY.** Derived from accepted T6/T8-E/T8-F authority under the T11 frontend-readiness method. It is implementation-planning Evidence, not Product/API/frontend authority. Material gaps route to the smallest owning authority; they are never repaired here by inventing semantics.

## 1. Goal

Prove that the future frontend represents the already-accepted MetalDocs and that the T11 implementation graph can actually deliver each browser surface against its real backend prerequisites.

This matrix is derived **before wireframes**.

Legend:

```text
COVERED       accepted authority supplies a coherent frontend home/contract
WIREFRAME     authority is coherent but material state/action must still be drawn/reviewed
FINDING       T11 graph/readiness coherence issue must be corrected before wireframing
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

All 78 application operations require `MetalDocsSession`. Unsafe mutations additionally obey the exact T8-E CSRF/conditional/idempotency profile. OIDC `/auth/login` and `/auth/callback` remain browser integration routes outside the 78-operation census.

## 3. Accepted human-goal coverage

| # | Accepted human goal | Primary frontend home | Accepted backend/wire input | Material interaction obligations | Status |
|---:|---|---|---|---|---|
| 1 | Establish/end session | application shell + external `/auth/login`/`/auth/callback` integration | `getSession`, `endSession`; OIDC browser routes; `SessionView` | no password input; HttpOnly session cookie; in-memory/server-state CSRF token; 401/session-end behavior; server remains current access authority | WIREFRAME |
| 2 | Discover official documents | `/documents` Library | `listDocuments` → `DocumentPage` | official/effective discovery only by default; admitted q/filters/cursor; no DRAFT/SUBMITTED as official; selection routes by returned `document_id` | WIREFRAME |
| 3 | Create Document | `/documents` creation interaction | `getDocumentCreationOptions`, supporting `getDocumentTypeNumberingPreview`, `createDocument` | options are guidance only; numbering preview never reserves; one Idempotency-Key per logical create; accepted result routes to returned Document identity | WIREFRAME |
| 4 | Inspect official/current Document truth | `/documents/:document_id` Document Official | `getDocument`; supporting `getDocumentResponsibleOwner`, `getRelease`, `getReleaseSource`, `getOfficialRenditionContent`, `getObsolescenceRequest` when referenced | official truth primary; DRAFT never silently rendered; release/source/rendition distinction; disclosure-safe current references; management only through accepted operations | WIREFRAME |
| 5 | Start/enter open Revision | Document Official → `/documents/:document_id/work` | `getDocument` disclosure-safe `open_revision`; `createDocumentRevision` when absent/eligible | existing open Revision identity comes from server truth; create-next uses idempotent command; History is not current-resource resolver | WIREFRAME |
| 6 | Author DRAFT | `/documents/:document_id/work` | `getRevision`, `getRevisionDraft`, `getRevisionDraftSource`, `updateRevisionDraft` | exact DRAFT ETag binds local editor/form; stale DRAFT always explicit 412 reconciliation; no silent merge/LWW | WIREFRAME |
| 7 | Upload/replace DRAFT source | Document Work | `startRevisionDraftUpload`, provider capability PUT, `completeRevisionDraftUpload`, `updateRevisionDraft` | provider PUT != READY != WorkingContent; expired allocation gets a new allocation; intended local bytes preserved where accepted; exact admission remains server authority | WIREFRAME |
| 8 | Submit / withdraw / cancel | Document Work | `createSubmission`, `getSubmission`, `getSubmissionSource`, `withdrawSubmission`, `cancelRevision` | Submission uses same logical key + semantic DRAFT If-Match on retry; immutable submitted subject; withdrawal/cancel exact state; failures separated from transport ambiguity | WIREFRAME |
| 9 | See actor-relevant work | `/work` My Work | `listAuthoringWork`, `listGovernanceWork` | projection only; authoring rows route to owner work lens; governance rows route to exact case; no lifecycle authority in My Work | FINDING COV-04 |
| 10 | Participate in governance | `/work/governance/:attempt_id` | `getGovernanceAttempt`, `listGovernanceFeedback`, `createGovernanceFeedback`, `getGovernanceStepDecision`, `recordGovernanceStepDecision`; exact governed source through admitted reads | immutable governed subject; feedback/decision under current server AuthZ; `allowed_actions` hints only; decision != publish | WIREFRAME |
| 11 | Inspect Document history | `/documents/:document_id/history` | `getDocumentHistory` + explicitly referenced supporting facts/content | Controlled Documents history only; not current-resource resolver; Audit separate; historical referenced content only when currently authorized | WIREFRAME |
| 12 | Initiate/manage obsolescence | Document Official | `createObsolescenceRequest`, `getObsolescenceRequest`, `withdrawObsolescenceRequest` | active request identity comes from disclosure-safe current truth; NoHumanApproval can complete synchronously; no fake human Step; target remains effective until success when governance is pending | WIREFRAME |
| 13 | Administer Organization | `/admin/organization` | `searchProviderSubjects`; Company/User/Profile/ProviderBinding/Eligibility/Area/Group operations | independent ETag domains remain independent; provider search ref is opaque; offboarding/security effects remain atomic; provider claims never Product truth | FINDING COV-01/COV-02 |
| 14 | Administer access | `/admin/access` | GroupMembership ops + `listRoles`, RoleAssignment ops; supporting User/Group/Area reads | exact static role vocabulary; exact Permission bundles; membership and RoleAssignment writes remain access-sensitive; no custom Role/Permission editor | FINDING COV-01/COV-02 |
| 15 | Administer document governance | `/admin/document-governance` | DocumentType/governance/eligible-template/numbering/config operations + `getDocumentTemplateRole`/`replaceDocumentTemplateRole` | config ETags; Template is not peer semantic owner; template administration does not grant content/history access; preview non-reserving | FINDING COV-02 |
| 16 | Inspect Audit | `/audit` | `listAuditEvents` → `AuditEventPage` | paging only; action evidence not current truth; no inferred filter; current historical-visibility AuthZ before paging | WIREFRAME |

No accepted human goal is currently missing a semantic frontend home. The findings are **T11 execution/readiness graph findings**, not evidence that T8-F Product semantics are wrong.

## 4. Exact 78-operation frontend reconciliation from T8-F

T8-F already provides the authoritative frontend consumer partition. T11 preserves it as frontend coverage truth:

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

Primary frontend home does not transfer semantic ownership. A supporting read may be consumed by another lens only under accepted disclosure and must not create duplicate client truth.

## 5. Cross-cutting architecture → UX obligations

| Accepted invariant | Required frontend behavior | Status |
|---|---|---|
| All 78 operations require current session | every protected screen assumes real session bootstrap; implementation graph cannot prove a protected tranche before AuthN/session exists | FINDING COV-01 |
| Server is Authorization authority | navigation/action hiding may improve UX only when based on actual returned truth; hidden UI is never correctness | WIREFRAME |
| `SessionView` has no effective Permission snapshot | stable Product navigation cannot be made correct by a client role/permission matrix | WIREFRAME |
| One semantic owner per meaning | composed screen may read several owner views; each write control routes to exactly one accepted owner operation | WIREFRAME |
| TanStack Query owns server state | no durable normalized client Product entity/global store becomes truth | COVERED |
| URL/router owns navigation context | admitted filters/cursors/route ids stay navigation context, not Product state | WIREFRAME |
| Forms bind to exact loaded ETag | stale whole-replacement and DRAFT-specific preconditions have explicit reconciliation UX; no silent overwrite | WIREFRAME |
| 10 accepted idempotent POST creations | same logical retry keeps same key; changed/new semantic command gets new key; key never shown as business identity | WIREFRAME |
| Problems are machine-branched by `Problem.code` | material 401/403/409/410/412/422/503 states cannot be collapsed into one generic toast when safe next action differs | WIREFRAME |
| Exact bytes have semantic URLs | provider bucket/key/version never displayed/parsed as Product identity; corruption fails before success presentation | WIREFRAME |
| Provider PUT != admitted content | upload progress/success cannot claim WorkingContent before completion/admission + DRAFT attachment | WIREFRAME |
| `allowed_actions` are hints | every command still rechecks canonical server truth | COVERED |
| Cursor paging is seek/integrity protected | no page-number/offset/total-count UX contract is invented | WIREFRAME |
| DocumentOfficialView current references are disclosure-safe | work/obsolescence navigation must use returned current references rather than History/Audit inference | WIREFRAME |
| Release is system-owned | no Publish button/public Release mutation is invented | COVERED |
| Audit != History != current truth | UI may cross-link admitted facts but never merges them into lifecycle authority | WIREFRAME |
| T8-B default-deny import graph | React feature layout never justifies new Go owner/application/package edges | COVERED |
| Hard no screen-shaped API | visual convenience cannot authorize operation 79 or a new read model | COVERED |

## 6. Coverage findings

### COV-01 — Original T11 S1→S2 order cannot close real HTTP/browser proof

Current candidate before this matrix says:

```text
S1 Organization
→ S2 Authentication + Authorization
```

But T8-E requires `MetalDocsSession` on **all 78 application operations**. S1's own completion contract claims real E3/E4 Organization administration. That is impossible before Authentication/session and canonical Authorization are executable.

Accepted T3/T10 already provide the escape from a bootstrap paradox:

```text
bootstrap/recovery = explicit non-serving operations concern
not an ordinary RBAC bypass
not a Product endpoint
not operation 79
```

T10 further states that exact runtime realization of semantic bootstrap belongs to T11 implementation planning.

**Disposition:** MATERIAL T11 graph defect. Do not weaken S1 proof. Correct the graph and make non-serving bootstrap realization a P3 prerequisite to the first protected semantic tranche.

### COV-02 — Original T11 tranche boundaries do not align with accepted frontend dependency homes

Examples:

```text
searchProviderSubjects
  semantic Authentication read
  primary frontend consumer = Admin / Organization provider-binding workflow

getDocumentTemplateRole / replaceDocumentTemplateRole
  paths under Document
  accepted T8-F primary frontend home = Admin / Document Governance

Admin / Access
  GroupMembership + RoleAssignment behavior
  requires admitted User/Group/Area reference reads for usable selection/context
```

The prior 26/7/10/... partition is mathematically correct but creates artificial cross-tranche UI incompleteness.

**Disposition:** MATERIAL T11 decomposition defect, not Product/T8-F defect. Prefer fewer vertical tranches aligned to complete user-facing capability clusters.

### COV-03 — Stable route != one-and-done implementation tranche

`/documents/:document_id` is one stable Document Official lens, but accepted behavior is progressively supplied by different capability slices:

```text
core current official truth / owner management
→ later Release/source/OfficialRendition presentation
→ later active obsolescence create/read/withdraw
```

Likewise `/work` contains authoring and governance projections whose target owner surfaces close at different times.

**Disposition:** completion contracts must state the exact **route capability state at each tranche exit**. A route may be implemented progressively without creating a second frontend phase; later slices enrich the same accepted lens.

### COV-04 — My Work governance row must not become a dead navigation path

`listGovernanceWork` supplies governance projection rows whose target is `/work/governance/:attempt_id` loaded by `getGovernanceAttempt`.

Implementing the projection materially before the target case surface would either expose a dead action or require an implementation exception.

**Disposition:** assign `listGovernanceWork` to the Governance/Release tranche that also implements the target Governance Case, while `listAuthoringWork` stays with the Document core/authoring tranche. This changes T11 implementation assignment only; operation meaning remains T8-E/T8-F.

## 7. Candidate corrected implementation tranche partition

The smallest coherent vertical partition after COV-01→04 is:

```text
S1  Identity + Organization + Access
    operations 1–33                                      = 33
    + /auth/login + /auth/callback outside census

S2  Document Governance administration
    operations 34–43 + 50–51                            = 12

S3  Library + Document core + authoring/history base
    operations 44–49 + 52–54                             = 9

S4  Document Work + Revision/content/Submission
    operations 56–66                                    = 11

S5  Governance work + Governance Case + Release/rendition
    operation 55 + operations 67–74                      = 9

S6  Obsolescence + Audit
    operations 75–78                                     = 4

TOTAL                                                    = 78
```

Count proof:

```text
33 + 12 + 9 + 11 + 9 + 4 = 78
```

Properties:

```text
zero operation reassigned semantically
zero Product/API meaning changed
zero new application operation
operation 79 absent
fewer artificial tranche seams
AuthN/session/Authorization available before every protected Product surface
Admin Organization + Access close together against usable identity/reference truth
Admin Document Governance owns its accepted T8-F primary operation set
My Work governance projection closes with its real Governance Case target
Document Official remains one route/lens progressively enriched by S3/S5/S6
```

This is a T11 implementation-graph correction only.

## 8. Coverage Matrix closure status

```text
accepted human goals mapped                 16 / 16
stable route/lens homes                     covered
T8-F operation consumer reconciliation      78 / 78
operation 79                                absent
material T11 graph findings                 4
Product/T8-F semantic reopen                not justified by current evidence
wireframes                                  not started
```

COV-01→04 must be folded into the T11 Lead + Node Completion Contracts before Screen Contract derivation proceeds. Wireframing remains blocked until that graph correction is coherent.