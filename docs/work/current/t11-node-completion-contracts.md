# T11 — Node Completion Contracts

> **TEMPORARY T11 CANDIDATE COMPANION / BRANCH-ONLY WORK.** This file refines the T11 Lead work graph; it is not durable authority and must be absorbed or removed before T11 integration. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose

The T11 execution graph is not complete if a node says only what work to attempt. Each node must define the **observable implementation state that MUST exist when that node closes**.

This companion therefore converts every P/S node into an exit contract. It does not choose new Product semantics, APIs, package topology, persistence meaning, frontend authority, runtime topology or cutover behavior. Those meanings remain owned by accepted T1→T10 authority.

A future implementation node closes only when its entire completion contract is true on the real implementation subject. File existence, scaffolding, TODOs, mocks-only success or an unintegrated layer are not completion.

## 2. Completion-contract law

Every implementation node has these mandatory dimensions:

```text
ENTRY
  exact predecessor/gate state required before work starts

PRODUCTION STATE
  concrete behavior/surface that exists and is usable at node exit

PERSISTENCE STATE
  schema/constraints/transactions/durable mechanism state required by that behavior

WIRE STATE
  accepted operations/schemas/headers/problems/exact-byte behavior realized at node exit

DEPENDENCY GRAPH STATE
  live packages/components and all relevant allowed/forbidden first-party edges satisfy T8-B

FRONTEND STATE
  accepted T8-F screen/lens behavior implemented for the node, consuming real generated wire

PROOF STATE
  positive + causal negative/fault evidence required by T9 on the real protected subject

EXPLICIT ABSENCES
  tempting but unaccepted mechanisms/authority that MUST still be absent

EXIT
  one falsifiable statement that determines whether the node may close
```

No field may be replaced by `TODO`, “follow-up”, “frontend later”, “integration later”, “tests later” or equivalent deferred ambiguity when the node already owns the affected claim.

Mechanical file/private-package decomposition may remain implementation-local only where accepted T8 authority deliberately leaves it private. A node completion contract freezes **observable and architectural outcome**, not arbitrary file layout.

## 3. Global dependency-graph condition

Every node inherits the accepted T8-B closed-world dependency law. The implementation verifier must classify every live first-party Go package bidirectionally and reject any edge not affirmatively allowed.

Normative class-level edges remain T8-B, including:

```text
runtime-entrypoint → composition
transport/http     → wire-contract
transport          → application
transport          → platform/observability only as technical instrumentation
application        → semantic-owner-public
application        → platform/txscope
application        → platform/idempotency only for admitted idempotent use cases
application        → platform/observability only as technical instrumentation
owner              → same-owner subtree
owner              → platform/txscope only when transaction participation requires it
platform           → platform technical DAG only
composition        → owner-public / application / transport / platform
wire-contract      → no Product-runtime first-party package
```

Among the default-denied edges that must remain impossible:

```text
owner → another owner
owner → application/composition
transport → owner
transport → SQL/persistence
application → application
application → owner-private
application → platform/postgres
platform → owner
foreign SQL as owner communication
second Authorization evaluator
cross-owner semantic dumping-ground package
```

A node cannot close while a newly introduced package is unclassified, an allowed edge is merely documented but unenforced, or a forbidden-edge negative fixture fails to make the verifier fail.

## 4. P0 — Authority / implementation-admission pin

### ENTRY

All roadmap implementation-gate prerequisites are actually satisfied: T11/T12 closed and operator-ratified as required, Whole-R10 coherence PASS, fresh independent challenge converged and explicit operator implementation authorization granted.

### MUST BE TRUE AT EXIT

```text
current main/authority snapshot is pinned
current implementation branch/base is pinned
78-operation canonical census is pinned
operation 79 remains absent
Idempotency-Key creations = exact 10
ETag read/mutation domains = 13/13
exact-byte resources = exact 4
accepted T8 topology/wire/frontend/runtime authorities are current
T9 proof baseline is current
T10 B0 source-truth classification is current before cutover preparation is treated as such
```

No Product implementation is introduced by P0 itself.

### PROOF

Mechanically/readably reconcile the pinned implementation input against the then-current roadmap and durable authorities. Any changed material authority after the pin invalidates P0 until reconciled.

### EXIT

A future implementer can identify the exact accepted target and prove that no Product code was started against stale or unauthorized authority.

## 5. P1 — Structural + executable-contract spine

### ENTRY

P0 closed.

### MUST BE IMPLEMENTED AT EXIT

```text
one Go module for backend Go code
accepted owner-first modular-monolith root classes exist as needed by the live target
one public importable surface per semantic owner
stateless application-leaf class available as the only semantic inbound route
transport/http wired only through generated wire boundary + application
composition root performs wiring only
canonical api/openapi/v1/openapi.yaml is executable SSOT
canonical Go server/wire projection generation works
canonical TypeScript projection generation works
one thin browser transport boundary consumes generated shapes
React SPA shell can boot against the real generated application client boundary
architecture verifier classifies the live first-party tree bidirectionally
closed-world/default-deny import enforcement fires on forbidden-edge fixtures
codegen/verification commands are repeatable from repository tooling
```

P1 does **not** need Product lifecycle implementation to prove structural contract/codegen behavior. Structural package roots that have no current P1 consumer must not be populated with speculative business implementation.

### WIRE

All 78 accepted operations exist in the canonical OpenAPI contract with zero invented operation and generated projections compile from that contract. P1 proves contract/projection closure; semantic behavior is implemented only by its assigned S tranche.

### FRONTEND

The SPA shell, router foundation, generated TypeScript consumption boundary and technical request mechanics can be exercised without creating a second DTO/route/Problem/Authorization authority.

### PROOF

T9 E1 + V1 structural lane + V2. Negative fixtures must demonstrate at least transport bypass, application→application, application→owner-private, owner→owner, platform→owner and foreign/cross-owner SQL rejection through the same verifier used in CI.

### EXPLICIT ABSENCES

```text
handwritten parallel application DTO contract
parallel route/Problem registry
frontend Authorization engine
BFF
common/shared semantic dumping-ground
operation 79
```

### EXIT

A semantic slice can be added vertically without inventing build, contract, transport, package-boundary or frontend-client architecture.

## 6. P2 — PostgreSQL / transaction correctness spine

### ENTRY

P1 closed.

### MUST BE IMPLEMENTED AT EXIT

```text
one PostgreSQL Product-state database bootstrap/migration mechanism
provider-neutral transaction scope whose lifecycle is application-owned
owner persistence can participate in caller-provided local transaction scope
READ COMMITTED baseline + accepted narrow serialization/CAS primitives are executable
constraint/transaction integration-test harness runs against real PostgreSQL
idempotency replay persistence mechanism supports the accepted atomic-completion model
same-local-commit Audit participation can be realized without owner→Audit imports
erasure/privacy obligations of replay storage are structurally supportable
migration/schema compatibility check can fail serving against incompatible schema
```

P2 establishes cross-cutting persistence machinery. Owner-specific tables/queries/constraints land with the semantic S tranche that owns and proves them unless T8-D explicitly requires an earlier shared primitive. P2 must not become a database-first implementation of all future semantic slices.

### PROOF

Real E2 causal tests prove transaction rollback, uniqueness/serialization primitives and idempotency/Audit atomic-participation mechanics on PostgreSQL rather than repository fakes.

### EXPLICIT ABSENCES

```text
platform/postgres owning semantic SQL
cross-owner repositories
ambient hidden transaction ownership
generic event store
generic workflow persistence
semantic tables with no current tranche consumer
```

### EXIT

Every later semantic owner can persist through its own authority while participating in one accepted local transaction and real PostgreSQL correctness model.

## 7. P3 — Runtime / dependency shell

### ENTRY

P1 closed. It may progress in parallel with P2 where the DAG permits.

### MUST BE IMPLEMENTED AT EXIT

```text
one accepted modular-monolith application runtime shell
composition-root wiring is executable
configuration is typed/validated/fail-closed
required secrets have explicit startup/readiness semantics
/livez and /readyz distinction is executable
bounded startup/readiness/shutdown/drain mechanics exist
observability plumbing exists without Product authority
one PostgreSQL pool/runtime dependency boundary exists
accepted external mechanism boundaries are attachable without owner leakage
River runs only in the accepted in-process topology when its named consumer becomes active
ManagedContentStore / identity-provider / renderer / MalwareInspector boundaries preserve T8 ownership/trust law
resource envelopes and dependency-specific failure isolation are enforceable
```

Mechanism adapters may be implemented when they have an accepted current S/P consumer and their selected profile is already authorized. P3 must not prebuild dormant providers/workers/frameworks merely because later capability is imaginable.

### PROOF

Real fail-closed config and dependency probes as claim-relevant; E5/E6 only where an actual external mechanism claim is made. The process remains live while readiness truthfully fails for a required unavailable dependency where T8-G requires that distinction.

### EXPLICIT ABSENCES

```text
second application runtime
external job platform
Redis
BFF
realtime bus
generic event bus
external Search
provider mechanism as semantic authority
```

### EXIT

The first semantic tranche can run through the real process/composition/config/observability shell without inventing runtime mechanics locally.

## 8. S1 — Organization — exactly 26 application operations

### ENTRY

P2 + P3 closed sufficiently to provide real PostgreSQL and runtime subjects.

### MUST BE IMPLEMENTED AT EXIT

```text
Organization semantic owner public surface is real
Company current settings are persisted and conditionally mutable as accepted
User identity/profile/provider-binding/eligibility semantics are real and remain separate concerns
Area identity/lifecycle semantics are real
Group identity + membership semantics are real
all required Organization Audit meaning participates atomically
all Organization OCC/uniqueness/offboarding invariants are enforced in real PostgreSQL
26/26 S1 operations execute through transport → application → owner on the generated wire
Admin / Organization frontend surfaces consume only generated transport/read models
frontend edit forms bind to the exact accepted current representation/ETag where applicable
server remains Authorization/disclosure authority
```

The exact 26 methods/paths are the S1 assignment in the T11 Lead; no extra Organization application operation may appear.

### PROOF

E2 + E3 on real owner/transaction paths, E4 for material browser claims. Include causal negatives for stale OCC, normalized uniqueness where applicable, disabled/ineligible target races, required Audit failure rollback and disclosure/permission denial.

### EXPLICIT ABSENCES

```text
frontend permission matrix
User flattened into one write model
provider claim as Organization truth
cross-owner SQL
operation outside assigned 26
```

### EXIT

A real user can exercise accepted Organization administration end-to-end through the SPA/API while every current-state/atomicity/import invariant remains enforced.

## 9. S2 — Authentication + Authorization — exactly 7 application operations plus two browser AuthN routes

### ENTRY

S1 closed.

### MUST BE IMPLEMENTED AT EXIT

```text
GET /auth/login and GET /auth/callback implement the accepted browser OIDC flow outside the application census
verified issuer+subject resolves only through current ProviderSubjectBinding
existing ENABLED User requirement is enforced
ApplicationSession issue/read/end behavior is real
session-bound CSRF behavior is real
Authorization owner holds the one canonical ALLOW/default-DENY decision authority
fixed Role + RoleAssignment semantics are real
current access revocation affects the next protected request as accepted
7/7 S2 application operations run on the real composed path
SPA auth/session shell and Admin / Access surfaces consume real server truth
```

### PROOF

Close GF1 and V3 with real E2/E3/E4 and E5 for the selected OIDC profile. Causal negatives include forged/replayed/tampered callback, unbound subject, provider claim attempting to grant Product permission, wrong CSRF, expired/ended session, disabled/revoked access and ProviderSubjectBinding replacement session revocation.

### EXPLICIT ABSENCES

```text
MetalDocs password login
provider roles/groups as Product Authorization
second Authorization evaluator
client-side authorization correctness
session application operation beyond accepted census
```

### EXIT

The real browser can establish/end a session and current authorization can allow/deny Product requests without any parallel identity or authorization authority.

## 10. S3 — Document Governance configuration — exactly 10 application operations

### ENTRY

S1 + S2 closed.

### MUST BE IMPLEMENTED AT EXIT

```text
DocumentType identity/configuration is real under Controlled Documents authority
governance/representation configuration is conditionally mutable as accepted
eligible-Template configuration is real without creating a Template semantic owner
numbering preview is non-reserving guidance only
Document Governance Template configuration projection is real and disclosure-safe
10/10 assigned operations run through real HTTP/application/owner/persistence
Admin / Document Governance frontend consumes purpose-built reads and exact ETags
configuration access does not imply governed document content/history access
```

### PROOF

E2/E3 and E4 where browser behavior is claimed. Negatives cover stale ETag, retired/ineligible referenced configuration, unauthorized content inference, numbering preview reservation attempts and required Audit rollback.

### EXPLICIT ABSENCES

```text
Template peer lifecycle/API
custom Role/Permission editor
generic workflow builder
number preview as number authority
operation outside assigned 10
```

### EXIT

All accepted governance configuration needed by document creation/governance can be managed through the real product without inventing a peer semantic owner or reserving future lifecycle behavior.

## 11. S4 — Document core + Work — exactly 12 application operations

### ENTRY

S3 closed.

### MUST BE IMPLEMENTED AT EXIT

```text
creation options expose only actor-usable accepted references
Document creation allocates final code and current Document truth atomically
responsible-owner relation + OCC/current eligibility rules are real
template-role relation + OCC/eligibility rules are real
next-Revision creation semantics are real
Library official discovery is real and never treats DRAFT/SUBMITTED as official
Document Official read/management lens is real
My Work authoring/governance projections are real and remain projection only
Document History is real Controlled Documents semantic history and not current-resource authority
12/12 assigned operations run end-to-end
SPA Library / Document Official / My Work / History surfaces are wired to real reads/mutations
navigation identities used by these surfaces come only from accepted server representations
```

### PROOF

Complete GF2 with prior nodes. Use E2/E3 plus claim-relevant E4. Causal negatives cover concurrent distinct Document code allocation, same-key changed semantic request, required Audit append failure, stale owner/template-role ETag, offboarding race and attempts to use History/projections as current mutation authority.

### EXPLICIT ABSENCES

```text
Document.currentStatus authority
screen-shaped read API
Search semantic owner
History/Audit as current truth
frontend lifecycle authority
```

### EXIT

A user can discover/create/navigate/manage the accepted Document core through real server truth, with numbering/idempotency/OCC/atomicity already proven rather than deferred to later authoring work.

## 12. S5 — Revision + exact content + Submission — exactly 11 application operations

### ENTRY

S4 closed and P3 content mechanisms needed by the accepted flow are real enough for E5 claims.

### MUST BE IMPLEMENTED AT EXIT

```text
Revision current/open semantics are real
DRAFT representation + exact strong ETag generation is real
DRAFT title/source update under If-Match is real
bounded upload allocation is real
provider upload is admitted only after accepted completion verification
ExactContentDescriptor remains authority over provider identity
WorkingContent attachment is distinct from provider PUT/admission
exact DRAFT/source read is real
Submission creation snapshots the accepted governed subject
Submission source read is exact and immutable as accepted
withdrawal/cancellation semantics are real
11/11 assigned operations run end-to-end
Document Work frontend implements the accepted DRAFT/editor/viewer/upload/submit/withdraw/cancel interactions
DOCX/PDF modes obey the accepted editor/viewer boundary
stale DRAFT reconciliation preserves local input and never silently LWW/merges
expired upload recovery allocates a new upload and never revives the old capability
```

### PROOF

Close GF3 with E2/E3/E4/E5 as claim-relevant and V5/V6/V7. Causal negatives include stale real DRAFT ETag, expired allocation, wrong size/hash/format, provider PUT without READY/WorkingContent, same-key changed Submission request and unauthorized source disclosure.

### EXPLICIT ABSENCES

```text
provider object identity as Product identity
EditorSession/lease correctness dependency
automatic stale merge
client hash/descriptor authority
operation outside assigned 11
```

### EXIT

A real browser can author exact DRAFT content through Submission using the real storage/admission/OCC/idempotency path, and every unsafe recovery branch has a defined/proven user outcome.

## 13. S6 — Governance + Release + OfficialRendition — exactly 8 application operations

### ENTRY

S5 closed and the P3 River/renderer/MalwareInspector mechanisms required by the accepted representation path are real.

### MUST BE IMPLEMENTED AT EXIT

```text
GovernanceAttempt + immutable Step/candidate snapshot semantics are real
feedback and Step Decision execute under current Authorization
required owner mutation + Audit + durable-effect intent commit atomically where required
River named durable work is redelivery-safe when activated
submitted PDF + required PDF reuses exact admitted bytes without renderer transformation
submitted DOCX + required PDF uses private renderer, re-admission and eligibility revalidation
Release/EFFECTIVE truth is system-owned and canonical
OfficialRendition semantics/exact-byte read are real
8/8 assigned application operations execute end-to-end
Governance Case browser surface renders exact governed subject and admitted actions
Document Official presentation resolves accepted Release/source/OfficialRendition representations without exposing mechanism state as truth
```

### PROOF

Close GF4 with E2/E3/E5/E6 and E4 for browser governance/presentation. Cover frozen-candidate drift, current AuthZ denial, same-transaction Audit/intent failure, duplicate River delivery, renderer failure, fidelity failure, provider corruption, malware-gate failure and exact scan-evidence binding.

### EXPLICIT ABSENCES

```text
public Release mutation
Approval peer product
queue state as Product truth
renderer output bypassing content admission
reviewer mutation of WorkingContent
```

### EXIT

A governed submission can reach canonical Release/OfficialRendition truth through the accepted human/system boundaries, and every required asynchronous/external effect remains mechanism rather than authority.

## 14. S7 — Obsolescence + Audit read — exactly 4 application operations

### ENTRY

S6 closed.

### MUST BE IMPLEMENTED AT EXIT

```text
obsolescence create/read/withdraw behavior is real under Controlled Documents authority
active request routing remains disclosure-safe
NoHumanApproval behavior creates no fake human Step
ordinary official discovery reflects resulting official truth without exposing DRAFT as official
Audit event paging is real evidence inspection only
4/4 assigned operations execute through real HTTP/application/owner paths
Document Official obsolescence interactions and Audit browser lens consume real generated wire
History remains semantic history; Audit remains evidence; neither becomes current business authority
```

### PROOF

Close GF5 with E2/E3/E4 as claim-relevant. Negatives include unauthorized absence inference, DRAFT-as-official leakage, stale active-obsolescence routing, History/Audit disclosure bypass and any attempted operation-79 navigation repair.

### EXPLICIT ABSENCES

```text
operation 79
generic Search endpoint
Audit-driven lifecycle reconstruction
generic approval/obsolescence action endpoint
```

### EXIT

Official discovery, obsolescence and Audit inspection are complete without introducing a second truth authority or a screen-shaped operation.

## 15. P4 — Runtime / failure isolation / recovery closure

### ENTRY

S1→S7 and P3 are closed on the integrated implementation candidate.

### MUST BE IMPLEMENTED AT EXIT

```text
bad config/missing required secret fails closed
incompatible schema prevents serving
PostgreSQL outage drives readiness false without falsifying liveness semantics
renderer outage is isolated from unrelated serving as accepted
MalwareInspector outage blocks required CLEAN admission without becoming global Product failure
OIDC outage blocks new provider-dependent login while otherwise-valid sessions retain their own validity
SIGTERM makes process unready first and performs bounded drain
River terminal required work is visible/inspectable/redrivable without becoming Product truth
resource/time/memory/scratch/concurrency envelopes are enforced
backup capture protects the required DB + exact-content set against GC race
restore invalidates sessions and completes privacy/security reconciliation before readiness
logs/metrics/traces prove material failure attribution without leaking secrets/PII
```

### PROOF

Close GF6 and V8/V9/V10 with real E2/E5/E6 subjects and causal failure/recovery tests.

### EXPLICIT ABSENCES

```text
serving while restored security/privacy state is unreconciled
dual Product authority
generic retry-jobs Product screen
unbounded resource assumptions
queue state as recovery truth
```

### EXIT

The implementation survives/rejects the accepted failure classes truthfully and has a real restorable recovery posture before it can be treated as a production candidate.

## 16. P5 — Whole implementation proof closure

### ENTRY

All prior P/S nodes are closed; T10 B1 exact private target exists; the exact candidate under proof is immutable for the proof run.

### MUST BE TRUE AT EXIT

```text
6/6 Golden Flows close on real production subjects with causal negatives
10/10 T9 cross-cutting properties close
T8-E runtime wire-conformance fixture classes close on real transport → application → owner/mechanism path
78/78 application operations are implemented and accounted exactly once
orphaned operations = 0
invented operations = 0
operation 79 = absent
Idempotency-Key creations = exact 10
ETag read/mutation domains = 13/13
exact-byte resources = exact 4
all T3 required same-local-commit Audit classes are represented
all claim-relevant external mechanisms have real E5 proof
all claim-relevant recovery/failure claims have real E6 proof
frontend coverage/wireframe/interaction contracts ratified by T11 are realized without drift
closed-world import graph and structural verifier pass on the live tree
no unresolved material proof waiver/exemption exists
```

P5 does not mutate Product authority. It proves the exact private B1 candidate that may become eligible for the T10 B2 clean seal.

### EXIT

One exact private production candidate has sufficient real evidence to proceed to B2 without any remaining implementation ambiguity disguised as a later integration task.

## 17. Cross-node regression law

Closing a downstream node never permits regression of an upstream completion contract.

Every implementation increment must state which completion-contract clauses it touches and rerun the smallest sufficient upstream regression proof. P5 reruns the whole accepted proof closure.

If a node cannot satisfy its completion contract without changing accepted Product/T1→T10 meaning, STOP and classify the evidence. Do not weaken the exit contract or add dormant machinery to make the stage appear complete.

## 18. Promotion law

This companion is temporary review structure. If T11 converges, its binding outcome is absorbed into the durable T11 implementation-program authority; this file is removed before integration so it cannot become a second permanent implementation-plan authority.