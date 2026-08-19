# R10-T6 — Operator Material Adjudication

> **Status:** OPERATOR-APPROVED MATERIAL ADJUDICATION / PLATFORM SUMMARY RATIFICATION NEXT  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED  
> **T7:** NOT OPEN

This file records the operator's explicit material adjudication of T6 after the greenfield Structural-Inversion pass, external/current evidence review, corrected Global-Maximum packet and second adversarial simplification pass.

It is a **stage-adjudication record**, not yet the durable T6 architecture authority. T6 promotion remains gated on explicit operator ratification of the platform-facing summary.

## 1. Adjudicated precedence

The operator approved the final T6 proposal in this precedence order:

```text
1. docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-candidate.md
2. docs/superpowers/analysis/2026-08-18-r10-t6-global-maximum-adjudication-packet.md
3. docs/superpowers/analysis/2026-08-18-r10-t6-final-adjudication-refinements.md
```

Where a later file is more specific, the later decision controls the adjudicated T6 candidate.

The evidence docket remains evidence only:

`docs/superpowers/analysis/2026-08-18-r10-t6-external-evidence-docket.md`

## 2. Operator-approved material slate

```text
T6-A  REFINED ACCEPT
  Rebuild the pre-launch /api/v1 contract from current product semantics.
  No /api/v2 compatibility layer or legacy shim.
  OpenAPI remains the contract SSOT with generated Go/TypeScript boundaries.
  OAS 3.0.3 remains the Launch feature set absent a named 3.1 consumer.

T6-B  ACCEPT
  Frontend uses stable semantic lenses: Library, My Work, exact Governance case,
  Document official/work/history, Audit, Administration.

T6-C  REFINED ACCEPT
  Public API is a closed semantic operation census, not a mirror of backend modules.
  No generic /actions endpoint and no separate Approval/Template/ControlledDocument product API.

T6-D  ACCEPT
  Server-derived allowed_actions are UX hints only; commands always re-evaluate T2/T3.

T6-E  REFINED ACCEPT
  Numbering = DOCUMENT_TYPE | DOCUMENT_TYPE_AREA, fixed '-', decimal sequence with
  minimum display width 3 and natural growth after 999. No custom formatting language.

T6-F  REFINED ACCEPT + FR-3
  Revision title + WorkingContent share one T2 DRAFT generation.
  HTTP expression = strong ETag + If-Match on PATCH /revisions/{id}/draft.
  Stale edit = 412 with zero mutation.

T6-G  REFINED ACCEPT
  Browser upload = bound upload_id OPEN → direct create-only provider upload → server READY
  verification → OCC attachment. Client never owns ExactContentDescriptor.
  Malware inspection remains governed-boundary server preflight.

T6-H  ACCEPT
  My Work aggregates actor work; governance decision always targets exact immutable Submission.
  Case participation never grants WorkingContent mutation.

T6-I  REFINED ACCEPT
  Semantic exact-byte URLs hide provider/storage identity.
  SourceOnly versus OfficialRendition presentation is explicit and truthful.

T6-J  REFINED ACCEPT
  Provider-neutral browser-buffer DOCX adapter is baseline.
  EigenPal-class is first candidate only after representative fidelity corpus proof.
  ONLYOFFICE-class is fallback on material fidelity failure.
  Exactly one Launch DOCX provider; no EditorSession correctness dependency baseline.

T6-K  ACCEPT
  Search materialization, search_refresh and external Search engine are OFF for Launch.
  Library uses canonical PostgreSQL current-effective code/title + explicit filters.

T6-L  ACCEPT
  Domain Document history and Audit remain separate read surfaces/authorities.

T6-M  REFINED ACCEPT
  Admin Center = Organization / Access / Document Governance.
  Current mutable administration uses strong ETag/If-Match where lost update matters.
  Launch binds pre-existing provider identities; no local credential/JIT-user path.

T6-N  REFINED ACCEPT
  RFC 9457 Problem Details transport.
  MetalDocs code is the single machine problem identifier; type is mechanically derived.
  Closed semantic families include dependency and ratelimit.

T6-O  REFINED ACCEPT + FR-1 + FR-2 + FR-4
  Natural HTTP idempotency first.
  User eligibility is a singleton current PUT resource.
  Governance Step Decision is a singleton immutable PUT resource.
  Durable Idempotency-Key replay remains only for genuinely non-idempotent semantic POST creation.
  Replay retention is bounded operational policy; 24h is first implementation-default candidate,
  not an architecture invariant.

T6-P  REFINED ACCEPT
  Unbounded lists use opaque cursor pagination; default 20 / maximum 100.
  No mandatory totals, offset pagination or generic filter/sort DSL.

T6-Q  REFINED ACCEPT
  Purpose-built semantic read models; no DB table/module DTO leakage.

T6-R  ACCEPT
  Preserve feature-sliced React + TanStack Query mechanism pattern while replacing legacy
  frontend feature taxonomy with current product semantics.

T6-S  ACCEPT
  Keycloak Authorization Code browser flow → MetalDocs ApplicationSession.
  No MetalDocs password input/change-password API, no ROPC/Direct Grant, no JIT User.
  Unsafe same-origin browser API requests use session-bound CSRF protection.

T6-V  ACCEPT
  Blank seed is trusted product mechanism content.
  Template-based create and create-next-Revision copy exact governed/released SOURCE,
  never OfficialRendition, and create independent WorkingContent.
```

## 3. Explicit subtraction accepted by the operator

The operator's approval includes permission to remove/rewrite current implementation surfaces that conflict with or are unnecessary under the adjudicated target, including current shapes for:

```text
Tenant/RLS product ontology
local password login/change-password
separate Approval product/API/workspace
separate Template lifecycle/API
public ControlledDocument peer object
Distribution/Notifications Launch surfaces
Taxonomy/Dictionary/Tokens platforms absent promoted requirement
legacy roles/capabilities as frontend authority
writer/editor session correctness dependency
scheduled publish
universal PDF/export machinery
materialized Search without consumer
reviewer mutation of WorkingContent during governance
legacy API compatibility shims
```

Sunk cost and migration convenience are not target requirements.

## 4. Frozen upstream authority

This adjudication does **not** reopen Product Contract REV001 or T1→T5. The following remain frozen unless material counterevidence explicitly triggers bounded reopen:

```text
Document != Revision != WorkingContent != Submission
Revision-owned title
immutable exact Submission
Release/effectivity semantics
T3 Authorization/Audit authority
T4 exact-content/admission authority
viewer/preview != OfficialRendition
Search never grants authority
canonical Search baseline / optional future materialization seam
no generic notification/event/integration platform Launch baseline
```

## 5. Gate consequence

```text
T6 material decisions             OPERATOR-APPROVED
platform-facing T6 summary        NEXT
platform-facing summary approval  REQUIRED BEFORE PROMOTION
T6 durable authority              NOT YET
Decision Registry reconciliation  NOT YET
T6 staging cleanup                NOT YET
T7                                NOT OPEN
implementation                    BLOCKED
```

A material-decision approval alone does not authorize T7 or implementation.