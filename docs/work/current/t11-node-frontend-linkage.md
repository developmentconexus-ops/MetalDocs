# T11 — Node ↔ Frontend Readiness Linkage

> **TEMPORARY T11 CANDIDATE WORK / BRANCH-ONLY.** This closes the handoff between the Node Completion Contracts and the completed F1→F9 frontend-readiness pack. It adds no new implementation node or Product meaning.

## 1. Law

A semantic node cannot close as "backend complete; frontend later".

For every S node:

```text
assigned application operations implemented on real generated HTTP path
+
owned semantic/persistence invariants implemented
+
its named frontend wireframes realized against those real operations
+
its material F6 controls proven
+
no actionable navigation target points to a downstream surface that does not yet exist
=
node eligible to close
```

The F5 wireframe IDs are planning handles only; final React component/file boundaries remain implementation-local inside T8-F topology.

## 2. P1 — structural/executable contract spine

Frontend completion contribution at P1 exit:

```text
React SPA shell boots
stable 10-route Product path tree exists
browser integration routes remain outside Product tree
OpenAPI-generated TypeScript projection is repeatable and consumed through one thin transport
the four accepted client state classes have concrete homes
no handwritten DTO/Problem/route contract authority
```

P1 does **not** claim Product route contents are semantically complete.

## 3. S1 — Identity + Organization + Access — 33 operations

Must realize at exit:

```text
WF-00  unauthenticated/session gate
WF-01  authenticated shell/logout
WF-16  Company settings
WF-17  Users + atomic User creation/provider preflight
WF-18  User Profile / Provider Binding / Eligibility
WF-19  Areas / Area lifecycle / Groups
WF-20  Group memberships
WF-21  Role catalog + RoleAssignments
```

Cross-cutting overlays required here:

```text
WF-X1 permission denied
WF-X2 not found/non-disclosable
WF-X3 CSRF recovery
WF-X4 OCC conflict for S1 ETag domains
WF-X5 ambiguous retry for S1 idempotent creates
WF-X6 dependency unavailable for OIDC/provider reads
```

Exit means a real browser can authenticate, navigate the shell and administer Organization/access end-to-end with current server authority.

## 4. S2 — Document Governance configuration — 10 operations

Must realize at exit:

```text
WF-22 DocumentType list/create/base + numbering preview
WF-23 governance/representation configuration
WF-24 eligible Template set
WF-25 Template configuration list/projection base
```

Concrete per-Document Template-role mutation in WF-25 remains explicitly incomplete until S3 because ordinary Documents do not exist before S3.

Exit means later Document creation can consume real configured DocumentType/governance truth; there is no generic workflow/Template product.

## 5. S3 — Library + Document core + Template-role + History — 9 operations

Must realize at exit:

```text
WF-02 Library official discovery
WF-03 Create Document
WF-04 Document Official core (without S5 release viewer enrichment)
WF-05 Responsible-owner management using approved T8-E-RO candidate projection
WF-14 History for facts reachable through S3
WF-25 concrete Document Template-role management enrichment
```

Navigation law:

```text
createDocument success
→ /documents/:document_id/work is the accepted final route target
BUT S3 cannot expose that post-create navigation as complete until S4 Work exists.
```

Therefore the S3 implementation increment may complete backend Document creation and its returned IDs, but the **user-visible create flow cannot be declared node-complete until the S4 target is present**. To preserve node independence without a dead action, implementation slicing should land the final LIB-02 submit/navigation control in the first S4 PR or land S3+the minimum S4 route target in one reviewable vertical increment. The architectural node dependency remains S3→S4; no placeholder route is allowed.

This is a deliberate completion-boundary precision, not permission to merge a user-facing dead link.

## 6. S4 — Revision authoring + My Work authoring + content + Submission — 13 operations

Must realize at exit:

```text
WF-06 Work route current resolver
WF-07 DRAFT DOCX authoring/save/submit
WF-08 DRAFT PDF + upload/admission/source replacement
WF-09 DRAFT OCC explicit reconciliation
WF-10 Submitted Revision/Submission state + termination
WF-11 Authoring lane
WF-04 Create/Open Revision action now live end-to-end
WF-03 Create Document success target now live end-to-end
```

Required cross-cutting frontend states:

```text
upload expired → new allocation/reupload same local bytes
provider PUT success != Product attachment
DRAFT 412 → local input preserved + explicit reconcile
ambiguous createSubmission → same Idempotency-Key
submitted source immutable/read-only
```

Exit means no Library/Official/My Work action points to an unimplemented Work target.

## 7. S5 — Governance work + Governance Case + Release/rendition — 9 operations

Must realize at exit:

```text
WF-11 Governance lane
WF-12 Submission Governance Case
WF-13 Obsolescence Governance Case presentation base
WF-04 official Release/source/OfficialRendition viewer enrichment
WF-14 governance/release/rendition History enrichment
```

Required user-visible distinctions:

```text
Governance Decision != Release/publish
submitted DOCX rendition pending != business failure state
submitted PDF + required PDF = zero-transform semantic reuse
exact governed Submission source is read-only
allowed_actions are hints only
```

Exit means actor-relevant governance projection has a real target and successful governance/release behavior is visible only from canonical owner truth.

## 8. S6 — Obsolescence + Audit — 4 operations

Must realize at exit:

```text
WF-05 obsolescence create/read/withdraw/completion states
WF-13 full obsolescence Governance subject interaction as applicable
WF-14 obsolescence History enrichment
WF-15 Audit evidence list
```

Required distinctions:

```text
active human-governed request
returned / withdrawn / completed request
NoHumanApproval synchronous obsolete result with no fake Step
Audit evidence != current Product state
```

Exit means GF5 browser composition is complete with no operation 79 or current-state resolver through History/Audit.

## 9. P4 — runtime/recovery closure

Frontend contribution is deliberately small:

```text
dependency failure states reached by Product interactions remain truthful/sanitized
/livez vs /readyz is not exposed as Product business UI
serving remains unavailable when runtime readiness/recovery law says so
restore/recovery has no Product "restore" screen or operation
```

WF-X6/WF-X7 may demonstrate safe browser failure presentation, but E6 runtime/recovery evidence remains primary.

## 10. P5 — whole implementation proof closure

P5 must compare the realized SPA against the reviewed T11 frontend pack:

```text
16/16 accepted human goals still have a real home
36/36 material surfaces are implemented or intentionally composed exactly as reviewed
10/10 stable Product routes retain accepted meanings
78/78 generated application operations retain at least one admitted consumer context
10/10 Idempotency-Key creation interactions preserve logical-command semantics
13/13 ETag domains preserve explicit current-representation binding/reconciliation
4/4 exact-byte resources retain exact-content viewer/editor behavior
T8-E-RO Responsible Owner candidates are disclosure-safe, complete and non-authoritative
no frontend Authorization engine appeared
no parallel handwritten DTO/API authority appeared
no parallel global server-truth store appeared
no screen-shaped operation appeared
operation 79 remains absent
```

P5 browser/composed proof uses real SPA + real application origin where T9 E4/E3 claims require it.

## 11. Implementation slicing constraint

A node may span multiple PRs, but no merged PR may knowingly leave a material user control pointing to an impossible target.

Allowed split example:

```text
S3 backend/core read foundations
→ internal/non-user-visible closure
→ S4 first vertical increment lands Document Work target
→ LIB-02/Create and OFF-04 controls become user-visible only with the target
```

Also allowed:

```text
one bounded vertical PR spans the S3→S4 seam when that is the smallest independently coherent user increment
```

Not allowed:

```text
button/link enabled now
→ target "coming later"
```

## 12. Closure

```text
S1 frontend completion home defined  YES
S2 frontend completion home defined  YES
S3 frontend completion home defined  YES + explicit S3→S4 no-dead-link seam
S4 frontend completion home defined  YES
S5 frontend completion home defined  YES
S6 frontend completion home defined  YES
P1/P4/P5 frontend obligations         defined
unowned F1→F9 frontend work           0
```

The Node Completion Contracts and this linkage together answer what **must be implemented and proven at the end of each execution stage**. They do not authorize implementation while the roadmap gate remains BLOCKED.
