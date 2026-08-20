# R10-T8E — Executable Wire Contract — Bootstrap

```text
ACTIVE STAGING
NON-AUTHORITATIVE
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-20  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Stage:** T8-E ACTIVE  
> **Upstream persistence authority:** `wiki/architecture/r10-t8d-persistence-realization.md`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This is the active non-authoritative bootstrap for R10 **T8-E — Executable Wire Contract**. It routes exact `/api/v1` contract work; it is not target authority.

---

## 1. Exact T8-E question

> **What is the smallest exact OpenAPI 3.0.3 wire contract that realizes the already-ratified T6 semantic journeys and T8-D persistence/concurrency laws without inventing new state, leaking internal persistence/mechanism shape, leaving Writers to choose missing fields/enums/errors, or turning the wire into a second lifecycle/AuthZ authority?**

---

## 2. Binding inputs

Read in repository authority order, including:

```text
Product Contract REV001
Whole-Product GCR + 4+1 ownership
T1→T6 durable semantic/API journey authority
T7 migration truth
T8-A technical disposition
T8-B backend topology
T8-C internal communication contracts
T8-D persistence realization
Decision Registry + amendments through T8-D
post-T6 implementation-readiness program
TRRB
current OpenAPI/generated/runtime implementation only as concrete evidence
```

Current `/api/v1` OpenAPI, generated Go/TypeScript, handlers and frontend consumers are evidence only. Legacy fields/routes/errors receive no survival entitlement.

---

## 3. T8-E owns

Freeze exact executable application-wire authority for the T6 operation census:

```text
OpenAPI 3.0.3 paths + operationIds
request schemas
success response schemas
fields / enums / nullability / requiredness
path/query/header parameters
strong ETag representation
If-Match requirements and stale-precondition wire outcomes
Idempotency-Key presence/scope-visible behavior
cursor pagination syntax / default / max / binding semantics
RFC 9457 Problem Details envelope
closed application problem-code vocabulary
upload allocation / completion / attachment wire contract
exact-byte semantic resource responses where T6 requires them
browser session/CSRF application API representation
purpose-built read projections/lenses
ReplaySnapshot-exposed success representation constraints
bounded body/response sizes where contract feasibility requires exact values
generated Go boundary
generated TypeScript boundary
runtime request/response validation and contract-conformance obligations
```

No implementation task may invent a missing wire field, enum, problem code, success shape or concurrency header after T8-E closes.

---

## 4. T8-E does not own

```text
semantic lifecycle / owner meaning                 T1→T6
persistent tables / constraints / locks            T8-D
owner/package/internal contract topology            T8-B/T8-C
frontend feature/query/cache realization            T8-F
runtime binaries/processes/secrets/deploy topology  T8-G
Golden Flow proof architecture                      T9
current→target cutover                               T10
implementation tranche decomposition                T11
```

T8-E may expose a real contradiction requiring a bounded reopen, but may not silently change upstream meaning to simplify JSON/OpenAPI.

---

## 5. Frozen upstream wire laws already inherited

```text
application API prefix          /api/v1
OpenAPI feature set             3.0.3
JSON naming                     snake_case
technical identifiers          opaque UUIDs
trusted instants               RFC3339 UTC
browser AuthN routes            /auth/login + /auth/callback outside generated application API
OIDC only; no password/ROPC/JIT User creation
session-bound CSRF on unsafe same-origin application requests
Library defaults to current official/effective truth
Document route meaning never silently switches by caller identity
Search q = code + current EFFECTIVE title only
materialized Search OFF
numbering preview non-reserving
DRAFT title+source one generation/ETag authority
whole replacements use strong ETag/If-Match
stale precondition => zero mutation
idempotent replay rechecks current session/CSRF/AuthZ/minimum visibility
potentially unbounded lists use cursor pagination; T6 default 20 / max 100
natural idempotency before Idempotency-Key
Idempotency-Key only on already-ratified non-idempotent creation operations
```

T8-D adds no wire-specific state; its VersionToken/generation/idempotency/persistence facts must be represented without leaking table layout or lock mechanics.

---

## 6. T8-D consequences T8-E must encode exactly

At minimum:

```text
ProviderSubjectBinding replacement current VersionToken
Company/UserProfile/Area/Group/DocumentType/owner/Template-role replacement VersionTokens where T6 exposes them
DRAFT WorkingContent generation strong ETag
Submission/current_submission boundedness without exposing a latest-history pointer
Revision lifecycle states without persisted Document.current_status
immutable Submission/Decision/Release/Rendition/history projections
GovernanceCase exact frozen candidate/step truth without Group-history coupling
ManagedContent upload_id/claim abstraction without exposing provider locator/DB handle internals
OPEN upload allocation vs complete READY vs semantic attach distinction
malware failure/dependency outcomes without exposing malware-inspection table shape
OfficialRendition pending/ready semantics consistent with named River intent
Idempotency replay success shape self-contained and PII-free
HMAC fingerprint/key version remain internal; never become public API fields
River/DB trust/lock details remain internal
```

---

## 7. Required operation census

Reconstruct the complete T6 route/operation census before designing schemas. Do not derive the target from current OpenAPI.

At minimum cover lenses/families:

```text
Session
Provider subject lookup / User create / provider-binding replacement
Organization Company/Users/UserProfile/Areas/Groups/GroupMemberships
Authorization roles/read vocabulary + RoleAssignments
Document governance administration
Document creation options / numbering preview
Document create / next Revision / responsible owner / Template role
Document Work DRAFT read/update
upload allocate / complete / attach
SUBMIT / feedback / withdraw / cancel
Governance Case / ACCEPT / RETURN_FOR_CHANGES
Library / Document Official / History / My Work
Official exact content / OfficialRendition exact bytes where applicable
obsolescence initiation / governance / withdrawal
Audit list/read
```

The exact census must reconcile one-to-one with T6 semantic operations. Missing route = unresolved architecture; extra route requires a named authority consumer.

---

## 8. Problem Details / error law

T8-E must freeze one closed application error contract based on RFC 9457 Problem Details without turning transport errors into business authority.

For every operation, classify at least:

```text
authentication/session failure
authorization/default-deny
not found / minimum disclosure
validation / malformed input
stale If-Match / generation
Idempotency-Key semantic conflict
state/lifecycle conflict
route/config impossible/recovery-required conditions
exact-content/upload admission failure
malware malicious / scanner dependency unavailable
provider/renderer/dependency unavailable
internal invariant failure
```

Exact status/problem-code mapping belongs T8-E. Do not inherit current legacy problem strings by existence.

---

## 9. Contract-generation / validation law

`api/openapi/v1/openapi.yaml` remains the `/api/v1` SSOT.

T8-E must freeze:

```text
OpenAPI schema authority
required generated Go boundary
required generated TypeScript boundary
no handwritten duplicate wire DTO authority
blocking generation/parity verification
runtime conformance proving handlers do not emit undeclared success/error shapes
```

Exact generator/tool commands are implementation evidence unless a material compatibility constraint requires freezing them here.

---

## 10. Alternatives / Method challenge

For material wire choices compare the smallest credible alternatives, including where applicable:

```text
resource-shaped vs journey/lens-shaped success schemas
full object replay vs minimal operation-local ReplaySnapshot
opaque cursor encoding alternatives
ETag representation derived from VersionToken/generation without exposing internal integer semantics
problem-code vocabulary placement
exact-byte response headers/content disposition
upload completion response vs DRAFT attachment response separation
admin directory vs purpose-built least-privilege option projections
```

Choose by semantic fidelity, client determinism, security/minimum disclosure, generated-boundary clarity, proof cost and future retrofit — not current route familiarity.

---

## 11. Required T8-E outputs

A promotable candidate must contain:

```text
complete operation/path/operationId census
exact request/response component inventory
exact field/enum/nullability ledger
exact success status/body/header matrix
exact error/problem-code matrix
ETag/If-Match matrix
Idempotency-Key matrix
cursor pagination contract
session/CSRF representation
upload/admission/exact-byte contract
Governance/History/Library/MyWork lens schemas
admin/read projection schemas
generated Go/TypeScript authority rules
runtime OpenAPI conformance rules
current-wire selective-reuse/disposition ledger
Structural Inversion / YAGNI challenge
T8-F/T8-G/T10 boundary check
independent adversarial review
operator-ratifiable Global Maximum
```

---

## 12. Exact next action

```text
reconstruct complete T6 semantic operation census
→ map each operation to exact T8-D VersionToken/generation/idempotency/content consequences
→ inventory current OpenAPI/generated boundary only as evidence
→ classify each current path/schema/problem code PRESERVE/REFINE/REWRITE/DELETE
→ derive exact success schema/header/status matrix before convenience
→ derive exact RFC 9457 problem-code matrix
→ derive pagination/cursor + ETag/If-Match + Idempotency-Key matrices
→ derive upload/exact-byte wire contract
→ prove generated Go/TS/runtime conformance feasibility
→ compare credible wire alternatives
→ apply Method + Structural Inversion + subtractive pass
→ independent Fable challenge
→ explicit operator ratification
```

Do **not** open T8-F or implement product code from this bootstrap.
