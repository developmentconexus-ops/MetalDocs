# T11 — Implementation Program & Execution Graph

> **TEMPORARY T11 CANDIDATE / BRANCH-ONLY WORK.** This file is not durable authority and must be absorbed or removed before merge. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose and boundary

T11 derives the smallest bounded implementation work graph and proof obligations from accepted T1→T10 authority so that later Product implementation can proceed without architectural improvisation, duplicate authority, late integration, frontend/backend drift or proof-by-ceremony.

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
Idempotency-Key creations             exact 10
ETag read / mutation domains          13 / 13
exact-byte resources                  exact 4
```

The future execution graph defined here is inert while the roadmap implementation gate is BLOCKED. It becomes executable only after every final implementation-gate condition in `docs/roadmap.md` is satisfied, including T12 closure, Whole-R10 coherence, fresh independent challenge and explicit operator implementation authorization.

## 2. Candidate composition

The current T11 candidate is deliberately split into three **temporary review artifacts**, not three durable authorities:

```text
docs/work/current/t11-implementation-program.md
  → owns execution DAG, operation assignment, proof/cutover integration and T11 closure

docs/work/current/t11-node-completion-contracts.md
  → defines the observable MUST-BE-IMPLEMENTED exit state for every P/S node

docs/work/current/t11-frontend-readiness.md
  → defines Coverage → Screen Contract → Navigation/Data Graph → Wireframe → Interaction Ledger → bidirectional trace closure
```

The companions refine this Lead and are mandatory parts of the same T11 candidate. They may not be ignored to satisfy the shorter summary tables in this file.

At durable promotion, their binding result is consolidated into the minimum T11 authority/implementation-readiness artifact set and these temporary work files are removed so no parallel authority remains.

## 3. Decision frame

### Evidence

Accepted authority already fixes the important semantic and realization decisions:

```text
Product / ownership / lifecycle / authorization / audit / content / async semantics
backend + internal-interface + persistence topology
closed-world/default-deny first-party import graph
executable application wire
frontend route/lens/state/transport realization
runtime / process / deployment realization
Whole-T8 global coherence
T9 Golden Flows + evidence classes + cross-cutting falsifiers
T10 one-way cutover barriers B0 → B4
```

The live tree contains no Product implementation to preserve. Therefore implementation planning must optimize for the accepted target, not for migration around sunk-cost code.

### Root cause to prevent

The remaining risk is no longer only missing architecture. It is **implementation ambiguity across boundaries**:

```text
execution-order ambiguity
+ node-exit ambiguity
+ late integration
+ frontend designed after backend without bidirectional trace
+ screen-shaped API invention
+ proof deferred until the system is too large to localize defects
```

Layer-first work can leave a locally-green backend that has never been consumed coherently by the browser. Flow-only work can duplicate foundations or omit non-representative operations. A graph that says what to start but not what MUST exist at exit still forces the implementer to design while coding.

### Target invariant

```text
every future implementation increment
→ has one bounded graph position
→ consumes only accepted authority
→ has explicit prerequisites
→ has an exact observable completion contract
→ implements its backend + persistence + wire + frontend obligations as one coherent tranche where applicable
→ satisfies the accepted closed-world dependency graph
→ has a falsifiable exit proof on the real protected subject
→ preserves all previously closed invariants
→ never creates a second semantic authority
```

For application operations specifically:

```text
78 accepted operations
→ each assigned to exactly one semantic implementation tranche
→ all implemented through the canonical wire SSOT
→ all accepted frontend consumers traced before implementation
→ zero unassigned
→ zero multiply-owned
→ zero invented
→ operation 79 absent
```

### Intentionally implementation-local, not T11 authority

T11 does not freeze exact future private file splits, dependency patch versions, secret values, environment identifiers, mechanical commit counts or ornamental UI pixels. Those are selected at execution time inside already-accepted T8 boundaries and current repository rules.

What T11 **does** freeze before implementation is every material outcome needed to prevent semantic guessing: graph position, node exit state, dependency class/edge law, accepted operation ownership, frontend screen/action/data trace, material interaction behavior and proof obligation.

## 4. Credible execution shapes

```text
A  technical-layer waterfall
   contract → database → backend → frontend → runtime → tests

B  Golden-Flow-only vertical slices
   GF1 → GF2 → ... → GF6

C  minimal shared spines + semantic vertical tranches + global proof closure
```

**C is selected.**

A is rejected because it can make each layer locally green while composed-system and browser/backend failure remain undiscovered until the end.

B is rejected because the six Golden Flows are a validation composition basis, not the complete 78-operation implementation census; using them as the only work decomposition would either duplicate shared correctness machinery or leave non-representative operations implicit.

C is the smallest sustainable shape: establish only shared mechanisms with multiple concrete consumers, then implement semantic slices end-to-end, then close the full wire/runtime/recovery contract globally.

## 5. Global execution law

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
T10 B1 private target
        ↓
P5 whole implementation proof closure on the exact private candidate
        ↓
T10 B2 real proof + clean seal
        ↓
T10 B3 first authoritative Product mutation / point of no return
        ↓
T10 B4 recovery point + serving fence + canonical activation
```

T10 B0 remains a prerequisite to treating any target preparation as cutover and must be revalidated before B1. It is not converted into a Product implementation node.

`26 + 7 + 10 + 12 + 11 + 8 + 4 = 78`.

Frontend is deliberately **not** a late standalone semantic phase. P1 establishes the accepted SPA/generated-TypeScript shell; each S tranche closes its assigned backend/persistence/wire **and** the reviewed T11 frontend Screen Contracts/wireframes whose real backend contract becomes available in that tranche.

A semantic S node cannot close with `backend done / frontend later`.

## 6. Program nodes

| Node | Work boundary | Depends on | App ops | Required exit basis |
|---|---|---:|---:|---|
| P0 | exact implementation-admission and authority snapshot | final roadmap gate | 0 | current authority/census pinned; no Product code; B0 classification current |
| P1 | structural/package + executable-contract + SPA transport spine | P0 | 0 semantic implementations | T9 E1/V1/V2 structural; generated Go/TS compile; closed-world import verifier fires |
| P2 | PostgreSQL/transaction/idempotency correctness spine | P1 | 0 | real E2 shared persistence mechanics; no database-first semantic prebuild |
| P3 | runtime/composition/config/dependency shell | P1 | 0 | accepted one-runtime shell + fail-closed technical mechanisms; E5/E6 only when claimed |
| S1 | Organization owner + Admin Organization frontend | P2 + P3 | 26 | all S1 completion-contract dimensions + real E2/E3/E4 as claim-relevant |
| S2 | browser AuthN/session + Authorization/access frontend | S1 | 7 | GF1 + V3 + exact S2 browser/server completion contract |
| S3 | Document Governance configuration + frontend | S1 + S2 | 10 | accepted config/OCC/disclosure behavior + frontend contract |
| S4 | Document creation/core + Library/Official/My Work/History | S3 | 12 | GF2 closure + numbering/idempotency/OCC + reviewed frontend surfaces |
| S5 | Revision DRAFT/content/Submission + Document Work | S4 + P3 content mechanisms | 11 | GF3 + exact-content/OCC/idempotency + editor/upload/recovery UI contract |
| S6 | Governance/Release/OfficialRendition + browser presentation | S5 + P3 River/renderer/scanner | 8 | GF4 + durable-effect/rendition/fidelity integrity + browser contract |
| S7 | obsolescence + Audit read + frontend | S6 | 4 | GF5 + disclosure/history/Audit separation + reviewed frontend surfaces |
| P4 | runtime failure isolation/shutdown/recovery | S1→S7 + P3 | 0 | GF6 + V8/V9/V10 real failure/recovery evidence |
| P5 | exact private whole-implementation proof candidate | all prior nodes + B1 | 78 closed | 6/6 GF; 10/10 V; wire runtime closure; all cross-cutting censuses; frontend drift check |

The mandatory exact exit meaning for every row is `t11-node-completion-contracts.md`. A row in this table is only a summary and cannot weaken its companion completion contract.

## 7. Exact 78-operation tranche assignment

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

## 8. Node completion contract — mandatory, not advisory

`t11-node-completion-contracts.md` defines for every P/S node:

```text
ENTRY
PRODUCTION STATE
PERSISTENCE STATE
WIRE STATE
DEPENDENCY GRAPH STATE
FRONTEND STATE
PROOF STATE
EXPLICIT ABSENCES
EXIT
```

A future node closes only when all applicable dimensions are true on the integrated real subject.

This prevents ambiguous milestones such as:

```text
backend complete / frontend later
schema exists / behavior later
endpoint exists / owner path not integrated
UI works against mocks / real API later
imports compile / forbidden architecture edge unverified
test exists / causal falsifier absent
```

The accepted T8-B closed-world/default-deny import graph is inherited by every node. P1 must make the classifier/verifier executable; every later node must keep the live tree classified and legal. Package existence never grants a new dependency edge.

## 9. Frontend implementation-readiness program — mandatory T11 work

T8-F remains closed. It already owns the semantic frontend realization contract. T11 now derives the implementation-ready **screen/action/data realization** from that authority before Product implementation begins.

The mandatory sequence is defined in `t11-frontend-readiness.md`:

```text
F0 authority recovery
→ F1 Frontend Coverage Matrix
→ F2 material interaction-surface inventory
→ F3 Screen Contracts
→ F4 Navigation/Data Graph
→ F5 functional low/mid-fidelity wireframes
→ F6 Material Interaction Ledger
→ F7 bidirectional frontend↔backend trace
→ F8 material finding classification / smallest justified reopen
→ F9 frontend readiness closure
```

No wireframe is accepted merely because the layout is sensible. A material surface must identify its Product goal, semantic owner, exact reads/writes, generated schemas, target-identity source, state class, ETag/idempotency/exact-byte mechanics, material Problems/recovery and what the frontend must not own.

Every material control must trace to one accepted owner operation. Every material data block must trace to accepted read truth. Every navigation edge must identify where the target identity comes from and which initial operation loads the target lens.

### Hard no screen-shaped API

If a wireframe needs information or an action absent from the accepted backend contract:

```text
STOP affected screen
→ prove whether the need follows from an accepted Product/human goal
→ classify the gap
→ if material, reopen only the smallest owning Product/T6/T8-E/T8-F authority
→ never add operation 79 or a convenience endpoint silently
```

Conversely, the frontend may not be forced to guess material backend truth merely to avoid a justified bounded reopen.

### Wireframe fidelity law

Before implementation, wireframes freeze interaction meaning, not ornamental pixels. They must show structure, material data, actions, navigation, state/recovery distinctions and editor/viewer behavior. Final color, micro-animation, exact spacing/shadow/radius and other non-functional styling are not architecture unless a material requirement depends on them.

## 10. Frontend realization during future implementation

The completed T11 frontend pack binds to semantic tranches rather than becoming a late frontend phase:

```text
P1
→ React SPA shell + router foundation + canonical generated TypeScript transport consumption

S1
→ reviewed Admin / Organization screen contracts/wireframes

S2
→ reviewed auth/session shell + Admin / Access screen contracts/wireframes

S3
→ reviewed Admin / Document Governance screen contracts/wireframes

S4
→ reviewed Library / creation / Document Official / My Work / History screen contracts/wireframes

S5
→ reviewed Document Work DRAFT/editor/upload/Submission screen contracts/wireframes

S6
→ reviewed Governance Case + Release/Official presentation screen contracts/wireframes

S7
→ reviewed obsolescence + Audit screen contracts/wireframes
```

Every browser mutation still relies on server authority. `allowed_actions` remain hints. No frontend Authorization engine, parallel global server store, handwritten application DTO contract or duplicate lifecycle state is permitted.

A browser proof closes only against the actual SPA and real application origin/path required by T9 E4.

## 11. Proof law per implementation node

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
Node Completion Contract clauses advanced/closed
new or modified production subjects
positive proof
causal negative/fault proof
T9 GF/V obligations advanced or closed
frontend Screen Contracts/interactions advanced or closed when applicable
application-census effect limited to already-assigned operations
cross-cutting census effect remains within accepted T8-E counts
remaining downstream prerequisites
```

No test exemption, baseline waiver, placeholder or dormant capability may be used to turn an unresolved material contradiction green.

## 12. Integration and PR slicing law

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
no semantic S tranche declared complete while its assigned reviewed frontend contract remains unimplemented
```

A downstream node may start only when its prerequisite semantics are sufficiently closed to provide the real subject it consumes. Parallelism is allowed only where the DAG permits it; P2 and P3 are the principal intentional early parallel branches.

## 13. T10 barrier overlay — binding order preserved

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
  must be current before any target preparation is treated as cutover; source/business-authority classification is revalidated before B1 work

B1
  follows implementation/runtime preparation only when the exact target candidate is private and ordinary canonical serving remains fenced

P5
  runs the complete T9 proof closure against that exact B1 private production candidate

B2
  requires the exact P5 candidate to satisfy real T9 proof plus the accepted operations/provenance clean seal

B3
  occurs only after B2 and explicit launch authorization; the first authoritative R10 Product mutation is the point of no return

B4
  requires an authoritative recovery point, disposable user-serving estate fenced and canonical R10 serving activation
```

Post-B3 recovery is forward on R10 authority only. No dual Product authority, legacy read fallback, disposable DEV/test restore as Product truth or compatibility bridge is introduced.

## 14. Stop / reopen routing

Classify implementation or frontend-readiness evidence before changing architecture:

```text
implementation/proof defect
  → fix inside the owning implementation node; accepted architecture stays closed

frontend placement/presentation issue with adequate accepted backend truth
  → correct inside T11 frontend readiness; no Product/T8 reopen

accepted mechanism assumption falsified
  → STOP that path; reopen only the smallest implicated T8 technical owner

required accepted human goal cannot be safely represented from current wire/read models
  → STOP affected surface; smallest Product/T6/T8-E/T8-F owner reopens as evidence requires

missing/extra application operation or consumer
  → STOP; smallest Product/T6/T8-E reopen; operation 79 never appears silently

semantic owner/lifecycle contradiction
  → STOP; smallest accepted Product/T1→T8 semantic owner reopens

real pre-R10 authoritative business truth discovered
  → STOP cutover planning/execution; smallest T7/T10 reopen

graph dependency/proof boundary itself shown unsound
  → correct T11 before ratification, or bounded T11 reopen after ratification
```

Dependency-version churn, provider patch releases, private file splits, ornamental UI choice or mechanical refactoring do not by themselves reopen architecture when accepted semantics/boundaries remain intact.

## 15. T11 closure contract

T11 can be ratified only when all are true:

```text
implementation DAG has no unresolved dependency cycle or unowned material work
Node Completion Contract exists for every P0→P5 / S1→S7 node
all node exit contracts specify production/persistence/wire/dependency/frontend/proof outcome as applicable
78/78 application operations are assigned exactly once
operation 79 is absent
Idempotency-Key creation census remains exact 10
ETag read / mutation domains remain 13 / 13
exact-byte resources remain exact 4
6/6 Golden Flows have a future implementation/proof owner
10/10 T9 cross-cutting properties have a future implementation/proof owner
T8-E runtime wire-conformance closure is assigned to P5 on the real composed path
real external-mechanism claims retain E5 proof lanes
failure/recovery claims retain E6 proof lanes
accepted T8-B closed-world import graph has a future executable verifier/negative-proof owner
Frontend Coverage Matrix is complete
material frontend interaction-surface inventory is complete
Screen Contract exists for every material surface
Navigation/Data Graph is complete for every material transition
functional wireframes exist for all material screens and materially different safe-action states
Material Interaction Ledger is complete
backend→frontend trace is complete
frontend→backend trace is complete
78/78 frontend operation coverage is reconciled
zero unresolved MATERIAL frontend/backend coverage finding remains
frontend remains a semantic consumer, never a second semantic authority
T10 B0→B4 ordering and point-of-no-return law are preserved exactly
no T12 work has begun
no Product implementation has begun
independent bounded T11 review has converged with no unresolved MATERIAL finding
operator ratification is explicit
```

The frontend readiness pack must be complete **before** independent final T11 convergence/ratification. T12 then receives an implementation program whose screens/actions/backend relationships are already falsifiable rather than deferred implementation design.

Expected durable promotion after convergence:

```text
docs/architecture/implementation-program.md
docs/decisions/t11-ratification.md
docs/index.md route update
minimum retained implementation-readiness/wireframe pack required by T12/future implementation
```

All temporary T11 work files must be removed before integration. T11 closure still does **not** authorize Product implementation; T12, Whole-R10 coherence, fresh independent challenge and explicit operator implementation authorization remain downstream gates owned by `docs/roadmap.md`.