# T11 — Node Completion Contracts

> **TEMPORARY T11 CANDIDATE COMPANION / BRANCH-ONLY WORK.** This file refines the T11 Lead work graph; it is not durable authority and must be absorbed or removed before T11 integration. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose

The T11 graph is incomplete if a node says only what work to attempt. Each node must define the **observable implementation state that MUST exist when that node closes**.

This companion converts every P/S node into an exit contract. It does not choose new Product semantics, APIs, package topology, persistence meaning, frontend authority, runtime topology or cutover behavior. Those meanings remain owned by accepted T1→T10 authority.

A future node closes only when its entire completion contract is true on the real implementation subject. File existence, scaffolding, TODOs, mocks-only success or an unintegrated layer are not completion.

## 2. Completion-contract law

Every node has these mandatory dimensions when applicable:

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

No field may be replaced by `TODO`, “follow-up”, “frontend later”, “integration later”, “tests later” or equivalent deferred ambiguity when the node owns the affected claim.

Mechanical private file/package decomposition may remain implementation-local only where accepted T8 authority deliberately leaves it private. A completion contract freezes **observable and architectural outcome**, not arbitrary file layout.

## 3. Global dependency-graph condition

Every node inherits T8-B's closed-world/default-deny first-party dependency law. P1 must make that law executable; every later node must preserve it.

Normative class-level directions remain T8-B, including:

```text
runtime-entrypoint → composition
transport/http     → wire-contract
transport          → application
application        → semantic-owner-public
application        → platform/txscope
application        → platform/idempotency only for admitted idempotent use cases
owner              → same-owner subtree
owner              → platform/txscope only when transaction participation requires it
platform           → platform technical DAG only
composition        → owner-public / application / transport / platform
wire-contract      → no Product-runtime first-party package
```

Default-denied examples remain:

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

A node cannot close while a newly introduced package is unclassified or a forbidden-edge negative fixture fails to make the same CI verifier fail.

## 4. P0 — Authority / implementation-admission pin

### ENTRY

All roadmap implementation-gate prerequisites are actually satisfied: T11/T12 closed and operator-ratified as required, Whole-R10 coherence PASS, fresh independent challenge converged and explicit operator implementation authorization granted.

### MUST BE TRUE AT EXIT

```text
current main/authority snapshot pinned
implementation branch/base pinned
application operations = 78
operation 79 = absent
Idempotency-Key creations = exact 10
ETag read/mutation domains = 13/13
exact-byte resources = exact 4
accepted T8 topology/wire/frontend/runtime authority current
T9 proof baseline current
T10 B0 source-truth classification current before cutover preparation is treated as such
```

P0 introduces no Product implementation.

### EXIT

A future implementer can identify the exact accepted target and prove no Product work started against stale/unauthorized authority.

## 5. P1 — Structural + executable-contract spine

### ENTRY

P0 closed.

### MUST BE IMPLEMENTED AT EXIT

```text
one Go module for backend Go code
accepted owner-first modular-monolith target classes available to the live target
one public importable surface per semantic owner
stateless application-leaf class is the only semantic inbound door
generated wire boundary sits between transport/http and application
composition root performs wiring only
canonical api/openapi/v1/openapi.yaml is executable SSOT
Go wire/server projection generation is repeatable
TypeScript projection generation is repeatable
thin browser transport consumes generated shapes
React SPA/router shell boots against the generated application client boundary
architecture verifier classifies the live first-party tree bidirectionally
closed-world/default-deny import enforcement fires on negative fixtures
repository codegen/verification entrypoints are repeatable
```

All 78 accepted operations exist in canonical OpenAPI with zero invented operation, but semantic behavior is implemented only by its assigned S tranche.

### PROOF

T9 E1 + V1 structural lane + V2. Negative fixtures demonstrate the verifier firing for transport bypass, application→application, application→owner-private, owner→owner, platform→owner and foreign/cross-owner SQL.

### EXPLICIT ABSENCES

```text
handwritten parallel DTO/application contract
parallel route/Problem registry
frontend Authorization engine
BFF
semantic common/shared dumping-ground
operation 79
```

### EXIT

A vertical Product slice can land without inventing build, contract, transport, package-boundary or frontend-client architecture.

## 6. P2 — PostgreSQL / transaction correctness spine

### ENTRY

P1 closed.

### MUST BE IMPLEMENTED AT EXIT

```text
one PostgreSQL Product-state database bootstrap/migration mechanism
provider-neutral transaction scope with application-owned lifecycle
owner persistence can participate in caller-provided local transaction
READ COMMITTED baseline + accepted narrow serialization/CAS primitives executable
real PostgreSQL integration-test harness
idempotency replay persistence supports accepted atomic-completion model
same-local-commit Audit participation is realizable without owner→Audit import
erasure/privacy posture of replay storage is structurally supportable
schema compatibility check can fail serving on incompatible schema
```

P2 establishes shared persistence mechanics only. Owner-specific tables/queries/constraints land with the semantic S tranche that owns/proves them unless T8-D explicitly requires an earlier shared primitive.

### PROOF

Real E2 causal tests prove rollback, uniqueness/serialization primitives and idempotency/Audit transaction-participation mechanics.

### EXPLICIT ABSENCES

```text
platform/postgres owning semantic SQL
cross-owner repository
ambient hidden transaction ownership
generic event/workflow persistence
semantic tables with no current tranche consumer
```

### EXIT

Later owners can persist their own truth while participating in one accepted local transaction and real PostgreSQL correctness model.

## 7. P3 — Runtime / dependency / non-serving bootstrap shell

### ENTRY

P1 closed. It may progress in parallel with P2 where the DAG permits.

### MUST BE IMPLEMENTED AT EXIT

```text
one accepted modular-monolith application runtime shell
composition-root wiring executable
configuration typed/validated/fail-closed
required secrets have explicit startup/readiness semantics
/livez and /readyz distinction executable
bounded startup/readiness/shutdown/drain mechanics executable
observability plumbing exists without Product authority
one PostgreSQL runtime dependency boundary exists
accepted external mechanism boundaries attach without owner leakage
River remains accepted in-process mechanism when named consumers become active
resource/dependency failure envelopes are enforceable
```

P3 also owns the implementation realization of the already-accepted T3/T10 **non-serving bootstrap/recovery concern** needed to break the first-user/session paradox and later cross B3 truthfully.

The bootstrap realization MUST:

```text
be non-serving / explicitly operator-controlled
be unreachable as an ordinary /api/v1 Product operation
not become an RBAC bypass available to normal serving
not create operation 79
use accepted semantic owners/transaction/Audit meaning rather than direct SQL as business authority
support disposable DEV/test/proof identity/access baseline before authority begins
support the later explicit B3 authoritative Company/User/ProviderSubjectBinding/access baseline without a synthetic activation marker
remain fenced from ordinary serving after its admitted use
```

T8-D `bootstrap/provisioner` remains DDL/provisioning trust only and is not semantic Product-bootstrap authority.

### PROOF

Real fail-closed config/dependency probes; prove bootstrap is inaccessible through ordinary serving and that it creates its admitted semantic truth through accepted owner/transaction boundaries. E5/E6 only where an actual external mechanism claim is made.

### EXPLICIT ABSENCES

```text
ordinary bootstrap Product endpoint
bootstrap permission/superuser role
second application runtime
Redis / BFF / realtime / generic event bus / external Search
provider mechanism as semantic authority
```

### EXIT

The first protected semantic tranche can authenticate/authorize against a disposable non-serving-created proof baseline without inventing a public bootstrap API, and the same concern can later realize T10 B3 under its separate launch gate.

## 8. S1 — Identity + Organization + Access — exactly 33 application operations plus two browser AuthN routes

### ENTRY

P2 + P3 closed sufficiently to provide real PostgreSQL/runtime/bootstrap subjects.

### APPLICATION ASSIGNMENT

```text
operations 1–33 exactly
+ GET /auth/login and GET /auth/callback outside application census
```

### MUST BE IMPLEMENTED AT EXIT

```text
Authentication semantic owner/session boundary real
Organization semantic owner real
Authorization semantic owner real and canonical ALLOW/default-DENY authority singular
Company current settings real
User identity/profile/provider-binding/eligibility real and separate concerns
Area identity/lifecycle real
Group identity/membership real
fixed Role + RoleAssignment semantics real
OIDC login/callback verifies protocol and resolves current binding + ENABLED User
ApplicationSession issue/read/end + session-bound CSRF real
current access revocation affects next protected request
all required security/Organization Audit meaning commits atomically
Organization/access OCC/uniqueness/offboarding invariants real in PostgreSQL
33/33 assigned application operations execute through generated wire → transport → application → owners
Admin / Organization and Admin / Access frontend surfaces consume real generated reads/mutations
provider-subject search supplies the provider-binding workflow without provider claims becoming Product truth
Admin / Access can use admitted User/Group/Area reference truth without a parallel directory model
```

### PROOF

Close GF1 and V3 with E2/E3/E4 plus E5 for the selected OIDC profile. Include stale OCC, normalized uniqueness where applicable, offboarding/access teardown atomicity, required Audit failure rollback, forged/replayed/tampered callback, unbound subject, provider-claim authorization attempt, wrong CSRF, expired/ended session, disabled/revoked access and binding-replacement session-revocation negatives.

### EXPLICIT ABSENCES

```text
MetalDocs password login
provider roles/groups as Product Authorization
frontend permission matrix
second Authorization evaluator
flattened User write model
cross-owner SQL
custom Role/Permission editor
operation outside assigned 1–33
```

### EXIT

A real browser can establish a session and fully exercise Organization/access administration through current server authority; every later protected Product slice has a real authenticated/authorized path rather than test bypass.

## 9. S2 — Document Governance configuration — exactly 10 application operations

### ENTRY

S1 closed.

### APPLICATION ASSIGNMENT

```text
operations 34–43 exactly
```

### MUST BE IMPLEMENTED AT EXIT

```text
DocumentType identity/configuration real under Controlled Documents
base/governance/representation configuration conditionally mutable as accepted
eligible-Template configuration real without Template semantic owner
numbering preview non-reserving guidance only
Template configuration projection real/disclosure-safe; empty is truthful before template Documents exist
10/10 assigned operations execute through real HTTP/application/owner/persistence
Admin / Document Governance base configuration surface consumes purpose-built reads + exact ETags
configuration access does not imply governed content/history access
```

Template-role administration for concrete Documents (`get/replaceDocumentTemplateRole`) is intentionally **not** declared complete here because no ordinary Document exists until S3. S3 owns that later enrichment of the same `/admin/document-governance` lens.

### PROOF

E2/E3 and E4 for claim-relevant browser configuration. Negatives cover stale ETag, invalid/retired configuration reference, unauthorized content inference, numbering-preview reservation and required Audit rollback.

### EXPLICIT ABSENCES

```text
Template peer lifecycle/API
generic workflow builder
number preview as number authority
Document template-role sub-surface falsely marked complete before S3
operation outside assigned 34–43
```

### EXIT

All accepted DocumentType/governance configuration required by later Document creation can be managed through the real Product; concrete Document template-role administration remains explicitly pending S3 rather than hidden.

## 10. S3 — Library + Document core + template-role + History — exactly 9 application operations

### ENTRY

S2 closed.

### APPLICATION ASSIGNMENT

```text
operations 44–51 + 53 exactly
```

### MUST BE IMPLEMENTED AT EXIT

```text
creation options expose only actor-usable accepted references
Library official discovery real and defaults to official/effective truth
Document creation allocates final code/current truth atomically
createDocument idempotency + required Audit atomicity real
Document Official core read real; DRAFT never presented as official
responsible-owner current relation + OCC/current-target eligibility real
concrete Document template-role read/write real under accepted template_use.manage authority
Document History operation real for every history fact reachable by the implementation so far
9/9 assigned operations execute end-to-end
Library + creation + Document Official core + History surfaces wired to real generated client
Admin / Document Governance is enriched with the concrete Document template-role sub-surface
Document Official release/rendition and obsolescence regions are explicitly not yet claimed; S5/S6 own those enrichments
```

Creating/entering a next Revision and My Work authoring navigation are intentionally **not** exposed as complete in S3 because their target Document Work surface is owned by S4.

### PROOF

E2/E3 + claim-relevant E4. Causal negatives cover concurrent distinct Document-code allocation, same-key changed create request, required Audit failure, stale owner/template-role ETag, offboarding race, DRAFT-as-official leakage and attempts to use History as current truth.

### EXPLICIT ABSENCES

```text
Document.currentStatus authority
screen-shaped read API
Search semantic owner
History/Audit as current truth
create-next-Revision affordance pointing to an unimplemented Work target
release/obsolescence UI claimed before S5/S6
```

### EXIT

A user can discover/create/open/manage the accepted Document core and inspect currently reachable semantic history, while every later route enrichment is named rather than disguised as completed frontend.

## 11. S4 — Revision authoring + My Work authoring + exact content + Submission — exactly 13 application operations

### ENTRY

S3 closed and P3 content mechanisms needed by the accepted flow are real enough for their E5 claims.

### APPLICATION ASSIGNMENT

```text
operation 52
operation 54
operations 56–66
= 13 operations
```

### MUST BE IMPLEMENTED AT EXIT

```text
create-next-Revision semantics real
listAuthoringWork projection real and routes to a real Document Work target
Revision current/open semantics real
DRAFT representation + exact strong ETag generation real
DRAFT title/source update under If-Match real
bounded upload allocation real
provider upload admitted only after accepted completion verification
ExactContentDescriptor remains authority over provider identity
WorkingContent attachment distinct from provider PUT/admission
exact DRAFT/source read real
Submission creation snapshots accepted governed subject
Submission source exact/immutable as accepted
withdrawal/cancellation semantics real
13/13 assigned operations execute end-to-end
Document Official gains create/enter-Revision interaction only now, because target Work is live
My Work authoring lane is live end-to-end
Document Work implements accepted DRAFT/editor/viewer/upload/submit/withdraw/cancel interactions
stale DRAFT reconciliation preserves local input; no silent LWW/merge
expired upload recovery uses a new capability; old capability never revived
```

### PROOF

Close GF3 and the authoring portion of GF2 with E2/E3/E4/E5 as claim-relevant and V5/V6/V7. Negatives include stale real DRAFT ETag, expired allocation, wrong size/hash/format, provider PUT without READY/WorkingContent, same-key changed Submission request and unauthorized exact-source disclosure.

### EXPLICIT ABSENCES

```text
provider object identity as Product identity
EditorSession/lease correctness dependency
automatic stale merge
client hash/descriptor authority
My Work governance row before its S5 target exists
operation outside assigned set
```

### EXIT

A real browser can move from official Document → Revision → DRAFT authoring/upload → Submission and My Work authoring without any dead navigation or deferred backend/frontend integration.

## 12. S5 — Governance work + Governance Case + Release + OfficialRendition — exactly 9 application operations

### ENTRY

S4 closed and P3 River/renderer/MalwareInspector mechanisms required by accepted representation paths are real.

### APPLICATION ASSIGNMENT

```text
operation 55
operations 67–74
= 9 operations
```

### MUST BE IMPLEMENTED AT EXIT

```text
listGovernanceWork projection real and routes to real Governance Case
GovernanceAttempt + immutable Step/candidate snapshot real
feedback + Step Decision execute under current Authorization
required owner mutation + Audit + durable-effect intent commit atomically where required
River named durable work redelivery-safe when activated
submitted PDF + required PDF reuses exact admitted bytes without renderer transformation
submitted DOCX + required PDF uses private renderer + re-admission + eligibility revalidation
Release/EFFECTIVE truth system-owned/canonical
Release source + OfficialRendition exact-byte reads real
9/9 assigned operations execute end-to-end
My Work governance lane live end-to-end
Governance Case renders exact governed subject and admitted actions
Document Official is enriched with accepted Release/source/OfficialRendition presentation
```

Document History is regression-tested/enriched for governance/release facts now reachable; it remains the same S3 operation/route, not a new API.

### PROOF

Close GF4 with E2/E3/E5/E6 and E4 for browser governance/presentation. Cover frozen-candidate drift, current AuthZ denial, same-transaction Audit/intent failure, duplicate River delivery, renderer/fidelity failure, provider corruption, malware-gate failure and exact scan-evidence binding.

### EXPLICIT ABSENCES

```text
public Release mutation
Approval peer product
queue state as Product truth
renderer output bypassing content admission
reviewer mutation of WorkingContent
governance My Work row with dead target
```

### EXIT

A real user can traverse My Work governance → exact case → decision/feedback and observe resulting Release/OfficialRendition truth without asynchronous mechanism becoming Product authority.

## 13. S6 — Obsolescence + Audit — exactly 4 application operations

### ENTRY

S5 closed.

### APPLICATION ASSIGNMENT

```text
operations 75–78 exactly
```

### MUST BE IMPLEMENTED AT EXIT

```text
obsolescence create/read/withdraw real under Controlled Documents authority
active-request routing disclosure-safe
NoHumanApproval creates no fake human Step
official discovery reflects resulting official truth without DRAFT leakage
Audit event paging real as evidence inspection only
4/4 assigned operations execute through real composed paths
Document Official gains complete accepted obsolescence interaction/state
Audit browser lens consumes real generated wire
History is enriched/regression-proven for obsolescence facts
History remains semantic history; Audit remains evidence; neither current truth
```

### PROOF

Close GF5 with E2/E3/E4 as claim-relevant. Negatives include unauthorized absence inference, DRAFT-as-official leakage, stale active-obsolescence routing, History/Audit disclosure bypass and any operation-79 navigation repair.

### EXPLICIT ABSENCES

```text
operation 79
generic Search endpoint
Audit-driven lifecycle reconstruction
generic approval/obsolescence action endpoint
```

### EXIT

Every accepted user-facing Product capability is now represented through the reviewed frontend surfaces and all 78 operations have a real consumer path.

## 14. P4 — Runtime / failure isolation / recovery closure

### ENTRY

S1→S6 and P3 closed on the integrated implementation candidate.

### MUST BE IMPLEMENTED AT EXIT

```text
bad config/missing required secret fails closed
incompatible schema prevents serving
PostgreSQL outage drives readiness false while preserving accepted liveness semantics
renderer outage isolated from unrelated serving
MalwareInspector outage blocks required CLEAN admission without global Product failure
OIDC outage blocks new provider-dependent login while valid sessions retain own validity
SIGTERM makes process unready first + bounded drain
River terminal required work visible/inspectable/redrivable without Product truth
resource/time/memory/scratch/concurrency envelopes enforced
backup capture protects required DB + exact-content set against GC race
restore invalidates sessions + privacy/security reconciliation before readiness
logs/metrics/traces attribute material failure without secret/PII leakage
```

### PROOF

Close GF6 and V8/V9/V10 with real E2/E5/E6 subjects and causal failure/recovery tests.

### EXIT

The implementation survives/rejects accepted failure classes truthfully and has a real restorable recovery posture before becoming a production candidate.

## 15. P5 — Whole implementation proof closure

### ENTRY

All prior P/S nodes closed; T10 B1 exact private target exists; the exact candidate under proof is immutable for the proof run.

### MUST BE TRUE AT EXIT

```text
6/6 Golden Flows close with causal negatives
10/10 T9 cross-cutting properties close
T8-E runtime wire-conformance classes close on real composed path
78/78 application operations implemented/accounted once
orphaned = 0
invented = 0
operation 79 = absent
Idempotency-Key creations = exact 10
ETag read/mutation domains = 13/13
exact-byte resources = exact 4
all T3 required same-local-commit Audit classes represented
claim-relevant external mechanisms have real E5 proof
claim-relevant recovery/failure claims have real E6 proof
reviewed T11 frontend Screen Contracts/wireframes realized without drift
closed-world import graph verifier passes live tree + negative fixtures
no unresolved material proof waiver/exemption
```

P5 does not mutate Product authority. It proves the exact private B1 candidate eligible for T10 B2.

### EXIT

One exact private production candidate has sufficient real evidence to proceed to B2 without implementation ambiguity disguised as later integration.

## 16. Cross-node progressive-lens law

A stable route/lens may be enriched by later slices when accepted capability prerequisites appear. This is not permission for “frontend later”; the exact progression must be named in the upstream completion contract.

Current accepted progression includes:

```text
/admin/document-governance
  S2 base DocumentType/governance/eligibility/numbering/config
  → S3 concrete Document template-role administration

/documents/:document_id
  S3 core official/current + responsible-owner
  → S4 create/enter Revision
  → S5 Release/source/OfficialRendition presentation
  → S6 obsolescence create/read/withdraw state

/work
  S4 authoring projection with live Work target
  → S5 governance projection with live Governance Case target

/documents/:document_id/history
  S3 operation/surface for reachable history
  → S4/S5/S6 regression/enrichment as new accepted semantic facts become reachable
```

No slice may expose an actionable navigation edge whose target surface is knowingly absent.

## 17. Cross-node regression law

Closing a downstream node never permits regression of an upstream completion contract. Every increment states which prior clauses it touches and reruns the smallest sufficient regression proof. P5 reruns whole closure.

If a node cannot satisfy its contract without changing accepted Product/T1→T10 meaning, STOP and classify the evidence. Do not weaken the exit contract or add dormant machinery to make the stage appear complete.

## 18. Promotion law

This companion is temporary review structure. If T11 converges, its binding outcome is absorbed into durable T11 authority; this file is removed before integration so it cannot become a second permanent implementation-plan authority.