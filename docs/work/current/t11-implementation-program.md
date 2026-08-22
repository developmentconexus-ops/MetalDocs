# T11 — Implementation Program & Execution Graph

> **TEMPORARY T11 CANDIDATE / BRANCH-ONLY WORK.** This file is not durable authority and must be absorbed or removed before merge. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose and boundary

T11 derives the smallest bounded implementation work graph and proof obligations from accepted T1→T10 authority so that later Product implementation can proceed without architectural improvisation, duplicate authority, late integration or proof-by-ceremony.

T11 does **not** implement Product code, create schema/OpenAPI/frontend/runtime/deploy implementation, begin T12, choose a new Product capability, add an application operation, or reopen accepted authority by preference.

Fixed opening state:

```text
opening main                          cae6ba48df5d611959c0390e0f2b9b8194d62a9d
T1 → T10                               CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                                   OPEN / ACTIVE candidate
T12                                   NOT OPEN
Product implementation                BLOCKED
legacy implementation in live tree    ABSENT
application operations                78
operation 79                          ABSENT
```

The future execution graph defined here is inert while the roadmap implementation gate is BLOCKED. It becomes executable only after every final implementation-gate condition in `docs/roadmap.md` is satisfied, including T12 closure, Whole-R10 coherence, fresh independent challenge and explicit operator implementation authorization.

## 2. Decision frame

### Evidence

Accepted authority already fixes the important semantic and realization decisions:

```text
Product / ownership / lifecycle / authorization / audit / content / async semantics
backend + interface + persistence topology
executable application wire
frontend realization
runtime / process / deployment realization
Whole-T8 global coherence
T9 Golden Flows + evidence classes + cross-cutting falsifiers
T10 one-way cutover barriers B0 → B4
```

The live tree contains no Product implementation to preserve. Therefore implementation planning must optimize for the accepted target, not for migration around sunk-cost code.

### Root cause to prevent

The implementation risk is no longer missing architecture. It is **execution-order ambiguity**: layer-first work can defer real integration until late; flow-only work can duplicate foundations or leave census coverage implicit; large cross-cutting PRs can make authority and proof ownership unclear.

### Target invariant

```text
every future implementation increment
→ has one bounded graph position
→ consumes only accepted authority
→ has explicit prerequisites
→ has a falsifiable exit proof on the real protected subject
→ preserves all previously closed invariants
→ never creates a second semantic authority
```

For application operations specifically:

```text
78 accepted operations
→ each assigned to exactly one semantic implementation tranche
→ all implemented through the canonical wire SSOT
→ zero unassigned
→ zero multiply-owned
→ zero invented
→ operation 79 absent
```

### Intentionally implementation-local, not T11 authority

T11 does not freeze exact future file splits, dependency patch versions, secret values, environment identifiers or mechanical commit counts. Those are selected at execution time inside already-accepted T8 boundaries and current repository rules. Freezing them now would manufacture stale or duplicate authority rather than remove semantic ambiguity.

## 3. Credible execution shapes

```text
A  technical-layer waterfall
   contract → database → backend → frontend → runtime → tests

B  Golden-Flow-only vertical slices
   GF1 → GF2 → ... → GF6

C  minimal shared spines + semantic vertical tranches + global proof closure
```

**C is selected.**

A is rejected because it can make each layer locally green while composed-system failure remains undiscovered until the end.

B is rejected because the six Golden Flows are a validation composition basis, not the complete 78-operation implementation census; using them as the only work decomposition would either duplicate shared correctness machinery or leave non-representative operations implicit.

C is the smallest sustainable shape: establish only shared mechanisms with multiple concrete consumers, then implement semantic slices end-to-end, then close the full wire/runtime/recovery contract globally.

## 4. Global execution law

After the future implementation admission gate opens:

```text
P0  authority/admission pin
 ↓
P1  structural + executable-contract spine
 ├──────────────┐
 ↓              ↓
P2 persistence  P3 runtime/dependency shell
 └──────┬───────┘
        ↓
S1 Organization — 26 ops
 ↓
S2 Authentication + Authorization — 7 ops
 ↓
S3 Document Governance configuration — 10 ops
 ↓
S4 Document core + Work — 12 ops
 ↓
S5 Revision + content + Submission — 11 ops
 ↓
S6 Governance + Release + rendition — 8 ops
 ↓
S7 Obsolescence + Audit read — 4 ops
        ↓
P4 runtime / recovery closure
        ↓
P5 whole implementation proof closure
        ↓
T10 B2 clean seal
        ↓
T10 B3 first authoritative Product mutation
        ↓
T10 B4 recovery point + serving fence + canonical activation
```

`26 + 7 + 10 + 12 + 11 + 8 + 4 = 78`.

Frontend is deliberately **not** a late standalone semantic phase. P1 establishes the accepted SPA/generated-TypeScript shell; each S tranche adds only the T8-F lens/query/editor/viewer behavior whose real backend contract is available in that tranche. This preserves one semantic owner and prevents a second frontend authorization/state model.

## 5. Program nodes

| Node | Deliverable boundary | Depends on | App ops | Minimum closure basis |
|---|---|---:|---:|---|
| P0 | exact implementation-admission and authority snapshot | final roadmap gate | 0 | roadmap/T11/T12/Whole-R10/operator authorization all current; census still 78 |
| P1 | build/package spine, canonical OpenAPI realization, generated Go/TS projections, structural architecture verifier, SPA shell | P0 | 0 new semantics | T9 E1; V1 + V2 structural lane; generated projections compile; operation 79 absent |
| P2 | PostgreSQL persistence, migration/transaction correctness primitives, required Audit/idempotency/OCC mechanisms used by later semantic owners | P1 | 0 | T9 E2; causal DB/transaction negatives for V4/V5/V6 primitives |
| P3 | one runtime shell plus accepted dependency boundaries/config/readiness/shutdown/observability and only actually-consumed external mechanisms | P1 | 0 | E1/E5/E6 as claim-relevant; fail-closed config/dependency probes; no extra runtime component |
| S1 | Organization owner and Admin organization lenses | P2 + P3 | 26 | real E2/E3; E4 where browser behavior is claimed; current-state/OCC/atomicity negatives |
| S2 | browser OIDC/session trust path plus Authorization owner/access lenses | S1 | 7 | GF1; E2+E3+E4 and E5 for selected OIDC profile; V3 |
| S3 | Document Governance configuration | S1 + S2 | 10 | E2+E3; E4 where claimed; stale-ETag/current-eligibility negatives; GF2 prerequisite |
| S4 | Document creation/core, responsibility/template role, Work and History composition | S3 | 12 | closes GF2 with prior nodes; E2+E3; V4/V5/V6 representative causal negatives |
| S5 | Revision DRAFT, upload/admission, exact source and Submission paths | S4 + P3 content mechanisms | 11 | GF3; E2+E3+E4+E5 as claim-relevant; V5/V6/V7 |
| S6 | GovernanceAttempt/Step/feedback/decision, Release and OfficialRendition | S5 + P3 River/renderer/scanner mechanisms | 8 | GF4; E2+E3+E5+E6; V4/V5/V6/V7/V8; browser lens when claimed |
| S7 | obsolescence commands/reads plus Audit event read | S6 | 4 | GF5 with discovery/history from prior slices; E2+E3+E4 as claim-relevant; disclosure/history/Audit falsifiers |
| P4 | final runtime failure isolation, durable-work, shutdown and recovery behavior | S1→S7 + P3 | 0 | GF6; E2+E5+E6; V8/V9/V10 |
| P5 | exact whole-implementation candidate and proof dossier | all prior nodes | 78 closed | 6/6 GF; 10/10 V; T8-E runtime conformance classes; 78/78 census; real-dependency proof where claimed |

A node may use more evidence classes than its minimum when the protected claim requires them. The table is not permission to replace a stronger accepted proof with a weaker one.

## 6. Exact 78-operation tranche assignment

This section owns **implementation assignment only**. Endpoint meaning, schemas and wire behavior remain owned by T6/T8-E and the canonical OpenAPI. Repeating a method/path here never creates a second contract authority.

### S1 — Organization — exactly 26

```text
GET    /api/v1/company
PUT    /api/v1/company
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/{user_id}
GET    /api/v1/users/{user_id}/profile
PUT    /api/v1/users/{user_id}/profile
DELETE /api/v1/users/{user_id}/profile
GET    /api/v1/users/{user_id}/provider-binding
PUT    /api/v1/users/{user_id}/provider-binding
GET    /api/v1/users/{user_id}/eligibility
PUT    /api/v1/users/{user_id}/eligibility
GET    /api/v1/areas
POST   /api/v1/areas
GET    /api/v1/areas/{area_id}
PUT    /api/v1/areas/{area_id}
GET    /api/v1/areas/{area_id}/lifecycle
PUT    /api/v1/areas/{area_id}/lifecycle
GET    /api/v1/groups
POST   /api/v1/groups
GET    /api/v1/groups/{group_id}
PUT    /api/v1/groups/{group_id}
DELETE /api/v1/groups/{group_id}
GET    /api/v1/groups/{group_id}/members
PUT    /api/v1/groups/{group_id}/members/{user_id}
DELETE /api/v1/groups/{group_id}/members/{user_id}
```

### S2 — Authentication + Authorization — exactly 7

```text
GET    /api/v1/session
DELETE /api/v1/session
GET    /api/v1/authentication/provider-subjects?query=...
GET    /api/v1/roles
GET    /api/v1/role-assignments
POST   /api/v1/role-assignments
DELETE /api/v1/role-assignments/{assignment_id}
```

`GET /auth/login` and `GET /auth/callback` are implemented in S2 but remain browser AuthN integration routes outside the 78-operation application census.

### S3 — Document Governance configuration — exactly 10

```text
GET  /api/v1/document-types
POST /api/v1/document-types
GET  /api/v1/document-types/{document_type_id}
PUT  /api/v1/document-types/{document_type_id}
GET  /api/v1/document-types/{document_type_id}/governance
PUT  /api/v1/document-types/{document_type_id}/governance
GET  /api/v1/document-types/{document_type_id}/eligible-templates
PUT  /api/v1/document-types/{document_type_id}/eligible-templates
GET  /api/v1/document-types/{document_type_id}/numbering-preview?area_id=...
GET  /api/v1/document-governance/templates
```

### S4 — Document core + Work — exactly 12

```text
GET  /api/v1/document-creation/options
GET  /api/v1/documents
POST /api/v1/documents
GET  /api/v1/documents/{document_id}
GET  /api/v1/documents/{document_id}/responsible-owner
PUT  /api/v1/documents/{document_id}/responsible-owner
GET  /api/v1/documents/{document_id}/template-role
PUT  /api/v1/documents/{document_id}/template-role
POST /api/v1/documents/{document_id}/revisions
GET  /api/v1/documents/{document_id}/history
GET  /api/v1/work/authoring
GET  /api/v1/work/governance
```

### S5 — Revision + content + Submission — exactly 11

```text
GET   /api/v1/revisions/{revision_id}
GET   /api/v1/revisions/{revision_id}/draft
PATCH /api/v1/revisions/{revision_id}/draft
POST  /api/v1/revisions/{revision_id}/draft/uploads
POST  /api/v1/revisions/{revision_id}/draft/uploads/{upload_id}/complete
GET   /api/v1/revisions/{revision_id}/draft/source
POST  /api/v1/revisions/{revision_id}/submissions
GET   /api/v1/submissions/{submission_id}
GET   /api/v1/submissions/{submission_id}/source
PUT   /api/v1/submissions/{submission_id}/withdrawal
PUT   /api/v1/revisions/{revision_id}/cancellation
```

### S6 — Governance + Release + rendition — exactly 8

```text
GET  /api/v1/governance-attempts/{attempt_id}
GET  /api/v1/governance-attempts/{attempt_id}/feedback
POST /api/v1/governance-attempts/{attempt_id}/feedback
GET  /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
PUT  /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
GET  /api/v1/releases/{release_id}
GET  /api/v1/releases/{release_id}/source
GET  /api/v1/official-renditions/{rendition_id}/content
```

### S7 — Obsolescence + Audit read — exactly 4

```text
POST /api/v1/documents/{document_id}/obsolescence-requests
GET  /api/v1/obsolescence-requests/{request_id}
PUT  /api/v1/obsolescence-requests/{request_id}/withdrawal
GET  /api/v1/audit/events
```

Census closure equation:

```text
S1 26
S2  7
S3 10
S4 12
S5 11
S6  8
S7  4
-----
   78
```

Any path/method outside this assignment that is proposed as a new application operation is a STOP and requires the smallest Product/T6/T8-E reopen. It cannot be repaired locally inside T11 or implementation.

## 7. Frontend realization rule

Frontend work follows semantic-slice availability rather than forming a second implementation graph:

```text
P1
→ React SPA shell + canonical generated TypeScript projection consumption

S1/S3
→ bounded Admin Center lenses

S2
→ auth/session shell + access behavior

S4
→ Library / creation / Document Work / My Work / History lenses as their real reads become available

S5
→ DRAFT/editor/upload/Submission paths

S6
→ Governance Case + Release/Official presentation

S7
→ obsolescence completion + Audit lens
```

Every browser mutation still relies on server authority. `allowed_actions` remain hints. No frontend authorization engine, parallel global server store, handwritten application DTO contract or duplicate lifecycle state is permitted.

A browser proof closes only against the actual SPA and real application origin/path required by T9 E4.

## 8. Proof law per implementation node

A semantic/runtime node closes only when all applicable claims satisfy the T9 global validation law:

```text
real production subject executed or mechanically inspected
+
positive invariant observed
+
causal negative/fault case proves the control fires
```

Mocks/fakes may support local unit tests, but cannot close a claim about PostgreSQL semantics, composed HTTP, browser trust, an external mechanism or recovery.

Each mergeable implementation increment must identify:

```text
accepted authority consumed
node/tranche owned
new or modified production subjects
positive proof
causal negative/fault proof
T9 GF/V obligations advanced or closed
78-operation census delta = 0 unless implementing an already-assigned operation
remaining downstream prerequisites
```

No test exemption, baseline waiver, placeholder or dormant capability may be used to turn an unresolved material contradiction green.

## 9. Integration and PR slicing law

The graph defines dependency/closure boundaries, not a mandate for giant PRs.

A graph node may require multiple PRs when that reduces review blast radius, but every PR must leave a coherent, independently testable increment. Split at a material ownership/proof seam, not mechanically by technical layer or file count.

Required laws:

```text
no direct commits to main
no Product implementation before P0 admission
no speculative cross-node scaffolding with no current consumer
no dormant tables/endpoints/workers/frameworks for future capability
no hand-edited generated application contract projection
no merge that knowingly breaks a previously closed invariant
no proof deferred merely because another layer is unfinished when the current claim is testable now
```

A downstream node may start only when its prerequisite semantics are sufficiently closed to provide the real subject it consumes. Parallelism is allowed only where the DAG permits it; P2 and P3 are the principal intentional early parallel branches.

## 10. T10 barrier overlay — binding order preserved

T11 does not replace T10 with an implementation-shaped cutover. The accepted monotonic barriers remain the only authority edge:

```text
B0 source truth
→ B1 private target
→ B2 real proof + clean seal
→ B3 first authoritative Product mutation / point of no return
→ B4 recovery point + serving fence + canonical activation
```

Execution mapping:

```text
B0
  revalidate source/business-authority classification before target preparation is treated as cutover

B1
  may be reached only by a privately prepared target using accepted runtime components; ordinary canonical serving remains fenced

B2
  requires the exact P5 candidate to satisfy real T9 proof and the accepted operations/provenance clean seal

B3
  occurs only after B2 and explicit launch authorization; the first authoritative R10 Product mutation is the point of no return

B4
  requires an authoritative recovery point, disposable user-serving estate fenced and canonical R10 serving activation
```

Post-B3 recovery is forward on R10 authority only. No dual Product authority, legacy read fallback, disposable DEV/test restore as Product truth or compatibility bridge is introduced.

## 11. Stop / reopen routing

Classify implementation evidence before changing architecture:

```text
implementation/proof defect
  → fix inside the owning implementation node; accepted architecture stays closed

accepted mechanism assumption falsified
  → STOP that path; reopen only the smallest implicated T8 technical owner

missing/extra application operation or consumer
  → STOP; smallest Product/T6/T8-E reopen; operation 79 never appears silently

semantic owner/lifecycle contradiction
  → STOP; smallest accepted Product/T1→T8 semantic owner reopens

real pre-R10 authoritative business truth discovered
  → STOP cutover planning/execution; smallest T7/T10 reopen

graph dependency/proof boundary itself shown unsound
  → correct T11 before ratification, or bounded T11 reopen after ratification
```

Dependency-version churn, provider patch releases, file splits or mechanical refactoring do not by themselves reopen architecture when accepted semantics/boundaries remain intact.

## 12. T11 closure contract

T11 can be ratified only when all are true:

```text
implementation DAG has no unresolved dependency cycle or unowned material work
78/78 application operations are assigned exactly once
operation 79 is absent
6/6 Golden Flows have a future implementation/proof owner
10/10 T9 cross-cutting properties have a future implementation/proof owner
T8-E runtime wire-conformance closure is assigned to P5 on the real composed path
real external-mechanism claims retain E5 proof lanes
failure/recovery claims retain E6 proof lanes
frontend is integrated by semantic slice, not a second semantic authority
T10 B0→B4 ordering and point-of-no-return law are preserved exactly
no T12 work has begun
no Product implementation has begun
independent bounded review has converged with no unresolved MATERIAL finding
operator ratification is explicit
```

Expected durable promotion after convergence:

```text
docs/architecture/implementation-program.md
docs/decisions/t11-ratification.md
docs/index.md route update
```

This temporary Lead file must be removed before integration. T11 closure still does **not** authorize Product implementation; T12, Whole-R10 coherence, fresh independent challenge and explicit operator implementation authorization remain downstream gates owned by `docs/roadmap.md`.
