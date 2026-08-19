# R10-T6 — Pre-Ratification Global Coherence Review

> **Status:** ACTIVE STAGING / ADVERSARIAL PRE-RATIFICATION REVIEW — **SUMMARY RATIFICATION HELD**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Reviewed target:** `docs/superpowers/analysis/2026-08-18-r10-t6-platform-facing-summary.md`  
> **Implementation:** BLOCKED  
> **T7:** NOT OPEN

This review was requested before T6 summary ratification to prove that the T6 platform model is coherent with the complete accepted architecture, not merely internally plausible.

Authority order used:

```text
AGENTS.md
→ DevelopmentConexus Engineering Method v1.0.0
→ Product Contract REV001
→ Whole-Product GCR A1–A10
→ Launch ownership topology 4+1
→ T1
→ T2
→ T3
→ T4
→ T5
→ Decision Registry
→ operator-approved T6 material staging
→ platform-facing summary under review
```

Current implementation was used only as prior evidence where relevant. It received no compatibility entitlement.

---

# 1. Verdict

```text
CORE T1→T5 / 4+1 COHERENCE     = PASS
T6 GLOBAL DIRECTION            = PASS
FORMAL T1→T5 REOPEN            = NONE
SUMMARY READY FOR RATIFICATION = NO — BOUNDED T6 CORRECTIONS REQUIRED
T7                              = BLOCKED
IMPLEMENTATION                  = BLOCKED
```

Finding count:

```text
BLOCKER = 0
MAJOR   = 8
LOW     = 5
```

This is **not** a redesign. The accepted T6 architecture survives. The findings close contract/concurrency/authority ambiguities at the T6 boundary before they become implementation forks.

Everything not named in this review remains frozen.

---

# 2. Cross-authority result

## Product Contract invariants

T6 correctly preserves:

```text
stable Document identity
Revision != autosave
REV000 initial / monotonic no-reuse
ordinary reader = current effective truth
immutable exact Submission
Governance binds exact Submission
return/withdraw preserve prior Submission
cancellation does not disturb older EFFECTIVE
Release sole normal effectivity authority
atomic replacement Release semantics
explicit governed obsolescence
lifecycle != physical delete
Template = ordinary governed Document / independent derived Document
Search never grants access
Audit != current state
migration truth remains T7
storage identity != semantic identity
no governed physical disposition Launch
future capabilities attach to stable references
Revision-owned title
```

## Whole-Product GCR A1–A10

T6 correctly avoids resurrecting:

```text
Artifact owner
separate Approval product/owner
quorum/reassign/fresh-auth workflow platform
Distribution / Periodic Review Launch-Core state
Dossier / Evidence / Records Launch state
Interchange/generic repository/export owner
old 5×43 AuthZ catalog
AuditChainHead global lock
Dictionary/Taxonomy/TemplateSpec/DRAFT-comments/scheduled-release adjuncts
```

## 4+1 ownership

No sixth business semantic owner is introduced. T6 surfaces remain lenses/mechanisms over:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit
```

Search, storage/admission, rendering/viewing/editing, idempotency and pagination remain mechanism/transport concerns.

---

# 3. M1 — Product Contract `status` discovery requirement is not closed

## Severity: MAJOR

### Evidence conflict

Product Contract §5-K requires core discovery filters including:

```text
code/title
Document Type
Area
responsible owner
status
```

The current T6 summary fixes Library to current-effective discovery and names Type/Area/owner filters, but drops `status` entirely.

A naive correction such as persisting `Document.current_status` would violate T1/Registry `DOC-09`, because current truth must remain derived from Revision/Release/Obsolescence rather than a second Document scalar.

### Required correction

Define **lens-scoped derived status filtering**, never a second persisted Document status:

```text
Library / GET /documents
  default status = effective
  history-authorized explicit status may select:
    effective
    obsolete
    cancelled   # only where no older EFFECTIVE remains

  DRAFT/SUBMITTED never become ordinary Library official results.

Authoring work
  filters current open Revision/work state in /work/authoring.

Governance work
  filters current attempt/Step state in /work/governance.

Document history
  filters exact Revision lifecycle states where needed.
```

For a Document with an older EFFECTIVE Revision plus a newer CANCELLED Revision, derived catalog status remains `effective` because the older EFFECTIVE remains current reader truth.

No persisted `Document.current_status` is introduced.

### Proof

```text
status filter satisfies Product Contract
AND
no duplicate Document current-state authority exists
AND
DRAFT/SUBMITTED never appear as ordinary official Library truth
```

---

# 4. M2 — Idempotency exact replay lacks crash-consistent atomicity

## Severity: MAJOR

### Counterexample

Current T6 promises:

```text
semantic POST commits
→ exact status/body stored for replay
```

but does not require the durable replay result to commit atomically with the semantic fact.

Failure:

```text
business fact COMMIT
→ process crashes before idempotency result COMMIT
→ client retries same key
→ endpoint cannot provide promised exact replay
```

The existing `409 conflict.idempotency_in_progress` requirement also implies a separate visible reservation state that is not required by any product invariant.

### Global-Maximum correction

Because semantic state and replay storage are both local PostgreSQL mechanisms, use the smallest atomic law:

```text
BEGIN local business transaction
→ insert/lock scoped Idempotency-Key
→ compare semantic fingerprint if key already exists
→ execute semantic transition
→ store durable replay result sufficient to reconstruct response
→ required Audit / durable intents as already mandated
→ COMMIT
```

Concurrent exact same-key requests may serialize/wait on the key. After the winner commits, the loser returns the stored replay.

Baseline therefore does **not** require a durable public `IN_PROGRESS` idempotency state or `409 conflict.idempotency_in_progress` behavior.

```text
same key + same fingerprint after commit → exact replay
same key + different fingerprint          → explicit 422 key-reuse validation error
rollback                                  → no committed semantic fact and no completed replay record
```

External preflight work may duplicate transiently before the local transaction; duplicate mechanism output remains reclaimable and never creates duplicate semantic facts.

### Everything frozen

Targeted `Idempotency-Key` operation set, bounded retention and natural HTTP idempotency decisions remain accepted.

---

# 5. M3 — `/api/v1` contract SSOT is conflated with browser/operations HTTP surfaces

## Severity: MAJOR

The T6 summary says:

```text
api/openapi/v1/openapi.yaml = public HTTP contract SSOT
```

while separately defining:

```text
GET /auth/login
GET /auth/callback
```

outside the generated JSON API.

Also, T5 requires operational visibility, and a production deployment still needs non-product liveness/readiness/metrics surfaces. A “closed public operation census” must not accidentally delete operational endpoints.

### Required correction

Use three explicit surface classes:

```text
A. /api/v1 application API
   → OpenAPI contract SSOT
   → closed T6 product/application operation census

B. browser AuthN integration
   → /auth/login
   → /auth/callback
   → fixed T6 web-integration surface, not application JSON API

C. operations surface
   → liveness/readiness/metrics/diagnostics as later deployment/observability design requires
   → non-product semantics
   → cannot become business authority or a bypass to product access
```

T4 restore non-serving readiness must eventually be reflected truthfully by readiness behavior; exact ops-route layout remains implementation/deployment design.

---

# 6. M4 — Platform-facing summary omits required integrated lifecycle journeys

## Severity: MAJOR

The detailed candidate contains the lifecycle decisions, but the platform-facing summary is intended to be the implementation-facing comprehension artifact and currently under-specifies several Launch-Core flows.

Before ratification it must explicitly restate the following integrated outcomes:

```text
Create blank/template → REV000 DRAFT

NoHumanApproval + SourceOnly SUBMIT
→ Submission + Release may commit in same local transaction
→ REV000/next Revision becomes EFFECTIVE immediately

UseGovernanceRoute SUBMIT
→ immutable Submission + GovernanceAttempt
→ one active sequential Step

RETURN_FOR_CHANGES
→ immutable old Submission/Decision/Feedback preserved
→ same Revision DRAFT

Submission withdrawal
→ same Revision DRAFT
→ no fake RETURN

Revision cancellation
→ terminal CANCELLED
→ older EFFECTIVE remains

required OfficialRendition pending
→ Revision remains SUBMITTED
→ no fake RENDERING/FAILED lifecycle state
→ mechanism processing/failure may be shown separately from lifecycle truth

replacement Release
→ predecessor SUPERSEDED + successor EFFECTIVE atomically

human-governed obsolescence active
→ target remains EFFECTIVE
→ final ACCEPT revalidates and makes OBSOLETE
→ RETURN/withdraw leaves EFFECTIVE

NoHumanApproval obsolescence
→ synchronous governed completion after checks/reason
→ zero fake human Step
```

This is summary completeness, not a T1/T2/T5 reopen.

---

# 7. M5 — Create-next-Revision source copy needs an explicit external-work race law

## Severity: MAJOR

T6 correctly says a new Revision seeds from the exact current EFFECTIVE released source, and T2 forbids provider calls inside the local semantic transaction.

Counterexample:

```text
actor reads EFFECTIVE R1
→ copies R1 source outside transaction
→ concurrent NoHumanApproval obsolescence makes R1 OBSOLETE
→ actor commits new Revision from stale pre-obsolescence assumption
```

### Required correction

Use the same preflight/revalidation pattern already accepted for template creation:

```text
read candidate current EFFECTIVE Revision + exact source
→ create independent managed copy outside semantic tx
→ BEGIN Document-serialized local tx
→ revalidate source Revision is still current EFFECTIVE and next-Revision creation is eligible
→ revalidate copied source descriptor/provenance matches selected source
→ create new Revision + WorkingContent
→ required Audit
→ COMMIT
```

If revalidation fails, no Revision is created; copied mechanism bytes are reclaimable.

Template-based creation continues to use T2's existing current-EFFECTIVE/eligibility revalidation law.

---

# 8. M6 — Security/authority-sensitive current-resource PUTs need lost-update preconditions

## Severity: MAJOR

T6 already requires `If-Match` for mutable admin configuration and User eligibility, but omits three current-resource mutations where silent last-write-wins can change authority or governed future behavior:

```text
User ProviderSubjectBinding replacement
Document responsible owner
Document Template role
```

### Required correction

These singleton current resources return strong ETags and require `If-Match` on replacement PUT:

```text
PUT /users/{user_id}/provider-binding
PUT /documents/{document_id}/responsible-owner
PUT /documents/{document_id}/template-role
```

Stale precondition:

```text
412 precondition.resource_changed
zero mutation
```

Exact same desired current value may be an idempotent no-op and must not fabricate duplicate semantic Audit transitions.

This preserves T3 current-authorization semantics and prevents two administrators from silently overwriting security/governance-bearing current truth.

---

# 9. M7 — Template administration must preserve governance_admin least privilege

## Severity: MAJOR

T3 deliberately says:

```text
governance_admin
→ organization/access/document-governance configuration
→ NO automatic document content read/history/governance/Audit authority
```

T6 Admin Center requires template role/eligibility management. A naive template picker that calls ordinary effective-document reads would silently grant content visibility to governance_admin.

### Required correction

`template_use.manage` authorizes a **bounded template-configuration read model** sufficient to administer role/eligibility without granting general document content/history access.

It may expose only the minimum governed identity metadata needed for selection, for example:

```text
Document id
stable code
current Template-role flag
current eligible/effective identity label needed for selection
```

Exact source/download/history remains protected by `document.read_effective` / `document.read_history` or exact case authorization.

UI placement in `Administration / Document Governance` does not change permission ownership:

```text
DocumentType route/representation config → document_type.manage
Template role/eligibility                → template_use.manage
Group identity                           → organization.manage
Group membership                         → access.manage
```

No new permission is introduced.

---

# 10. M8 — Numbering code uniqueness is missing

## Severity: MAJOR

T6 introduces `DocumentType.code` and `Area.code` as formatted numbering inputs but currently states syntax/immutability without explicit uniqueness.

Without uniqueness:

```text
Type A code = PO
Type B code = PO
→ both can generate PO-001
```

or:

```text
Area A code = RH
Area B code = RH
→ TYPE_AREA prefixes collide
```

The final Document-code uniqueness constraint would detect the collision too late and turn normal allocation into avoidable failure.

### Required correction

Within one Company:

```text
normalized DocumentType.code = unique
normalized Area.code         = unique
```

Normalization is the already-proposed trim + uppercase ASCII-alphanumeric rule.

Committed Document code remains globally unique within Company and never reuses.

The exact `1..16` code-length ceiling has no current Product Contract evidence. Before durable promotion either:

```text
(a) justify 16 from a current/source code census,
OR
(b) retain only "bounded length" in architecture and freeze the exact API maximum during implementation-contract design.
```

This LOW sub-point does not block the uniqueness correction.

---

# 11. L1 — One interactive DOCX provider != one renderer

## Severity: LOW

Clarify summary wording:

```text
exactly one interactive DOCX editor/viewer adapter in Launch
```

This does **not** collapse T5's separate server-side OfficialRendition renderer mechanism. Renderer/converter selection remains independently governed by T5 and may use a different product.

No dual interactive editor runtime remains the T6 rule.

---

# 12. L2 — Range reads were promoted beyond T4 evidence

## Severity: LOW

T4's managed-content contract requires exact full read and marks bounded/Range read optional.

T6 currently says the semantic byte gateway baseline supports full **and Range** reads. No product invariant currently requires Range.

Correct to:

```text
exact full semantic read = required
Range support = optional mechanism activated when the selected PDF/viewer/file-size evidence proves benefit
semantic URL/authorization does not change when Range is added
```

This avoids turning a performance optimization into a product invariant.

---

# 13. L3 — `allowed_actions` must not become a second policy implementation

## Severity: LOW

Retain:

```text
allowed_actions = UX hint only
command rechecks canonical truth
```

Add proof law:

```text
allowed_actions must be derived through the same canonical T3 permission/scope and Controlled-Documents predicate components (or a provably shared equivalent), never through a parallel frontend/backend role matrix.
```

A stale hint can cause a harmless UX mismatch; it can never authorize a command.

---

# 14. L4 — User creation should explicitly compose AuthN/Organization truth atomically

## Severity: LOW

Provider-directory lookup is external preflight. After a provider subject is resolved, successful `POST /users` should establish the intended local product facts as one local business transition:

```text
User
+ required UserProfile fields
+ ProviderSubjectBinding
+ required Audit evidence
```

No successful create should expose an accidental half-created local User/binding combination.

This follows T2's one-local-transition law and does not merge Authentication/Organization ownership.

---

# 15. L5 — Future-evolution seam review

## Severity: LOW / PASS

T6 was checked against every named future horizon and no structural reopen is required:

```text
Distribution / Read & Acknowledge
  → Release + User/Group remain stable seams

Periodic Review
  → Document + exact current EFFECTIVE Revision remain stable

Dossier
  → stable Document reference remains available

Evidence
  → can own independent exact content through T4 without Artifact

Retention/Hold/Disposition
  → can attach to stable semantic identities/history without storage becoming lifecycle

Governed Export
  → semantic IDs + exact-content descriptors/byte routes remain usable

External Repository IMPORT/PUBLISH
  → target-owner + T4 copy/admission seam remains provider-neutral

Training/LMS
  → Release/effective truth remains attachable

multi-document Change Control
  → stable Document/Revision + explicit next-Revision transition remain orchestratable

pooled tenancy
  → stable Company identity remains; substrate/API context may reopen later without rewriting document history

CRDT/realtime collaboration
  → can replace DRAFT OCC/API mechanism without changing Revision/Submission identity
```

No dormant future module is justified.

---

# 16. Correction set / everything frozen

## Required before T6 summary ratification

```text
C1 lens-scoped status discovery semantics; no Document currentStatus
C2 idempotency replay result atomically coupled to semantic commit; remove baseline IN_PROGRESS public state
C3 distinguish /api/v1 application contract from /auth integration and operations surfaces
C4 restore integrated lifecycle journeys to platform-facing summary
C5 next-Revision source-copy preflight + commit-time current-EFFECTIVE revalidation
C6 If-Match on provider-binding / responsible-owner / template-role current resources
C7 bounded template-admin metadata read under template_use.manage; no implicit content read
C8 Company-unique normalized DocumentType.code + Area.code
```

## Low refinements

```text
L1 clarify interactive DOCX adapter != server rendition renderer
L2 Range read remains optional until evidence
L3 allowed_actions uses shared canonical policy components
L4 local User creation composes User/Profile/Binding/Audit atomically after external provider lookup
L5 record future-seam pass
```

## Frozen

Everything else in:

```text
Product Contract REV001
Whole-Product GCR A1–A10
4+1 ownership
T1→T5
Decision Registry
operator-approved T6 material decisions not named above
```

remains unchanged.

Formal T1→T5 reopen set:

```text
EMPTY
```

---

# 17. Recommended gate

```text
T6 material core                           OPERATOR-APPROVED / PRESERVED
Pre-ratification Global Coherence Review  COMPLETE
T6 correction set C1→C8                    OPERATOR ADJUDICATION NEXT
Platform-facing summary ratification       HELD
T6 durable promotion                       NOT YET
T7                                         NOT OPEN
implementation                             BLOCKED
```

After C1→C8 are adjudicated and incorporated:

```text
correct platform-facing T6 summary
→ rerun bounded coherence delta against Product Contract + T1→T5
→ explicit operator summary ratification
→ durable T6 promotion / Registry reconciliation / staging cleanup
→ only then T7
```
