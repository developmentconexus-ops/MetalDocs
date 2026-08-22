# T11 — Implementation Program & Execution Graph

> **TEMPORARY T11 CANDIDATE / BRANCH-ONLY WORK.** This is the consolidated Lead candidate after frontend-readiness derivation. It is not durable authority and must be promoted/absorbed before T11 integration. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose and boundary

T11 converts accepted T1→T10 authority into the smallest bounded implementation execution graph whose nodes have exact observable exit states and falsifiable proof obligations.

The target is not merely an ordered backlog. A future implementation node closes only when:

```text
required production behavior exists
+ accepted persistence/wire/dependency boundaries exist where owned
+ its user-facing surfaces are connected to the real backend path
+ positive proof succeeds
+ causal negative/fault proof demonstrates protected controls fire
+ previously closed invariants remain true
```

T11 does **not** implement Product code, begin T12, add Product capability, create a semantic owner, add an application operation, choose a second frontend/backend authority or reopen accepted authority by preference.

Fixed candidate envelope:

```text
opening main                          cae6ba48df5d611959c0390e0f2b9b8194d62a9d
T1 → T10                              CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                                  OPEN / ACTIVE candidate
T12                                  NOT OPEN
Product implementation                BLOCKED
legacy implementation                 ABSENT
application operations                78
operation 79                          ABSENT
Idempotency-Key creations             exact 10
ETag read / mutation domains          13 / 13
exact-byte resources                  exact 4
stable SPA Product routes             exact accepted 10
```

The future execution graph is inert until the final roadmap implementation gate opens after T11/T12/Whole-R10/fresh challenge/operator authorization.

## 2. Method outcome

Three implementation shapes were tested:

```text
A technical-layer waterfall
  contract → DB → backend → frontend → runtime → end-to-end

B Golden-Flow-only slices
  GF1 → GF2 → ... → GF6

C shared correctness spines
  + complete semantic vertical tranches
  + whole-system proof closure
```

**C is selected.**

A is rejected because locally green layers can hide integration failures until late.

B is rejected because the six Golden Flows are a validation composition basis, not the complete 78-operation implementation census.

The first T11 candidate split Document core/create from Document Work. F1/F4/F7 frontend-readiness analysis falsified that split: accepted `createDocument` terminates in Document Work, so a separate boundary would either expose a dead target or allow a node to call itself complete without a complete human journey.

The corrected Global Maximum merges Document core/create and authoring into one vertical tranche. Node size is controlled through reviewable PR slicing **inside** the node rather than by severing an accepted user flow.

## 3. Global execution graph

After future implementation admission:

```text
P0  exact authority / implementation-admission pin
 ↓
P1  structural + executable-contract spine
 ├────────────────────┐
 ↓                    ↓
P2 persistence        P3 runtime/dependency/non-serving-bootstrap shell
 └──────────┬─────────┘
            ↓
S1 Identity + Organization + Access                         33 ops
 ↓
S2 Document Governance configuration                       10 ops
 ↓
S3 Document core + creation + authoring + Submission       22 ops
   + Library + My Work authoring + History
 ↓
S4 Governance work + Governance Case + Release/rendition    9 ops
 ↓
S5 Obsolescence + Audit                                     4 ops
 ↓
P4 runtime / durable-work / recovery closure
 ↓
T10 B1 exact private target
 ↓
P5 whole implementation proof closure on that exact target
 ↓
T10 B2 real proof + verified clean seal
 ↓
T10 B3 first authoritative Product mutation / point of no return
 ↓
T10 B4 authoritative recovery point + serving fence + canonical activation
```

Operation closure:

```text
S1  33
S2  10
S3  22
S4   9
S5   4
------
    78
```

T10 B0 remains a prerequisite before target preparation is treated as cutover and is revalidated before B1.

## 4. Completion-contract law

Every node below owns these dimensions when applicable:

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

No owned claim may close with:

```text
TODO
frontend later
integration later
tests later
mock-only success
placeholder route/control
dormant future capability
```

Private file decomposition remains implementation-local where accepted T8 authority deliberately leaves it private. T11 freezes observable/architectural outcome, not arbitrary file count.

## 5. Closed-world dependency graph inherited by every node

P1 makes T8-B mechanically executable; every later node preserves it.

Representative allowed class directions remain:

```text
runtime-entrypoint → composition
transport/http     → wire-contract
transport          → application
application        → semantic-owner-public
application        → platform/txscope
application        → platform/idempotency only for admitted idempotent operations
owner              → same-owner subtree
owner              → platform/txscope only when transaction participation requires it
platform           → platform technical DAG only
composition        → owner-public / application / transport / platform
wire-contract      → no Product-runtime first-party package
```

Default-denied classes include:

```text
owner → another owner
owner → application/composition
transport → owner
transport → semantic SQL
application → application
application → owner-private
application → platform/postgres
platform → owner
foreign/cross-owner SQL
second Authorization evaluator
semantic common/shared dumping-ground package
```

Any live first-party package without a class, any forbidden edge or a negative fixture that stops firing blocks the node.

## 6. P0 — Authority / implementation-admission pin

### ENTRY

All final roadmap implementation prerequisites actually hold, including T11/T12 closure, Whole-R10 coherence, fresh independent challenge and explicit operator implementation authorization.

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
T9 validation baseline current
T10 B0 source-truth classification current
```

### EXPLICIT ABSENCE

No Product implementation begins inside P0.

### EXIT

A future implementer can prove the exact target authority and cannot start against stale or unauthorized assumptions.

## 7. P1 — Structural + executable-contract spine

### ENTRY

P0 closed.

### MUST BE IMPLEMENTED AT EXIT

```text
one Go module for backend Go code
accepted owner-first modular-monolith target classes live
one importable public surface per semantic owner
stateless application leaf is the only semantic inbound door
generated wire boundary sits between transport/http and application
composition root performs wiring only
canonical api/openapi/v1/openapi.yaml is executable SSOT
Go wire/server projection generation repeatable
TypeScript projection generation repeatable
thin browser transport consumes generated shapes
React SPA shell boots against generated client boundary
stable 10-route Product path tree exists
architecture verifier classifies live first-party tree bidirectionally
closed-world/default-deny import enforcement fires on negative fixtures
repository codegen/verification entrypoints repeatable
```

All 78 accepted operations exist in canonical OpenAPI; semantic behavior lands only in the assigned S tranche.

### PROOF

T9 E1 + V1 structural lane + V2. Negative fixtures include transport bypass, application→application, application→owner-private, owner→owner, platform→owner and foreign/cross-owner SQL.

### EXPLICIT ABSENCES

```text
handwritten parallel DTO/application contract
parallel Problem/route registry
frontend Authorization engine
parallel global server-truth store
BFF
semantic common/shared dumping ground
operation 79
```

### EXIT

A semantic tranche can land without inventing build, codegen, package, transport or frontend-client architecture.

## 8. P2 — PostgreSQL / transaction correctness spine

### ENTRY

P1 closed.

### MUST BE IMPLEMENTED AT EXIT

```text
one PostgreSQL Product-state database bootstrap/migration mechanism
provider-neutral transaction Scope with application-owned lifecycle
owner persistence participates in caller-provided local transaction
READ COMMITTED + accepted narrow serialization/CAS primitives executable
real PostgreSQL integration-test harness
idempotency key/replay persistence realizes atomic completion model
same-local-commit Audit participation realizable without owner→Audit import
replay erasure/privacy posture structurally supportable
schema compatibility check can fail serving closed
```

Owner-specific tables/queries/constraints land with their semantic tranche unless T8-D requires an earlier shared primitive.

### PROOF

Real E2 causal tests prove rollback, transaction participation, uniqueness/serialization primitives and replay/Audit atomicity mechanics.

### EXPLICIT ABSENCES

```text
platform/postgres semantic SQL
cross-owner repository
ambient hidden transaction ownership
generic workflow/event persistence
semantic tables with no current consumer
```

### EXIT

Semantic owners can persist their own truth while participating in one accepted local correctness model.

## 9. P3 — Runtime / dependency / non-serving bootstrap shell

### ENTRY

P1 closed; may progress in parallel with P2 where the graph permits.

### MUST BE IMPLEMENTED AT EXIT

```text
one modular-monolith application runtime shell
composition-root wiring executable
typed/validated fail-closed configuration
required-secret startup/readiness semantics
/livez and /readyz distinction executable
bounded startup/readiness/shutdown/drain mechanics
observability plumbing without Product authority
PostgreSQL runtime dependency boundary
accepted external mechanism boundaries attach without owner leakage
River remains in-process mechanism when named consumers activate
resource/dependency failure envelopes enforceable
```

P3 also realizes the accepted T3/T10 **non-serving bootstrap/recovery concern** that breaks the first-user/session paradox and later supports B3.

Bootstrap MUST:

```text
be explicitly operator-controlled and non-serving
be unreachable as ordinary /api/v1 Product operation
not become normal-serving RBAC bypass
not create operation 79
use accepted owners/transaction/Audit meaning rather than direct SQL as semantic authority
support disposable DEV/test/proof identity/access baseline pre-authority
support later authoritative Company/User/ProviderSubjectBinding/access baseline at B3
create no synthetic Product activation marker
remain fenced from ordinary serving after admitted use
```

T8-D provisioning/DDL trust is not semantic Product bootstrap authority.

### PROOF

Real config/dependency negatives plus proof bootstrap is inaccessible through ordinary serving and creates admitted semantic truth through accepted boundaries. E5/E6 only for actual external claims.

### EXPLICIT ABSENCES

```text
ordinary bootstrap Product endpoint
bootstrap Permission/superuser Role
second application runtime
Redis / BFF / realtime / generic event bus / external Search
provider mechanism as semantic authority
```

### EXIT

S1 can execute against a real disposable proof baseline without test-only auth bypass, and the same concern can later realize B3 under T10 gates.

## 10. S1 — Identity + Organization + Access — exactly 33 operations

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
Authorization singular ALLOW/default-DENY authority real
Company settings real
User identity/Profile/ProviderSubjectBinding/eligibility separate and real
Area identity/lifecycle real
Group identity/membership real
fixed Role + RoleAssignment semantics real
OIDC login/callback verifies protocol and resolves current binding + ENABLED User
ApplicationSession issue/read/end + session-bound CSRF real
current access revocation affects next protected request
required security/Organization Audit commits atomically
Organization/access OCC/uniqueness/offboarding invariants real in PostgreSQL
33/33 operations execute generated wire → transport → application → owners
provider-subject search serves User/binding flows without provider claims becoming Product truth
```

### FRONTEND STATE AT EXIT

The exact reviewed frontend blueprint must be realized for:

```text
session gate + authenticated shell
Company settings
Users + atomic User creation
Profile / Provider Binding / Eligibility
Areas + Area lifecycle
Groups
Group memberships
Role catalog + RoleAssignments
```

No navigation/control may rely on a frontend permission matrix.

### PROOF

Close GF1 and V3 with E2/E3/E4 plus E5 for selected OIDC. Causal negatives include forged/replayed/tampered callback, unbound subject, provider-claim authorization attempt, wrong CSRF, expired/ended session, disabled/revoked access, binding-replacement revocation, stale OCC, offboarding teardown and required Audit failure rollback.

### EXPLICIT ABSENCES

```text
MetalDocs password login
provider Role/Group Product authorization
frontend Authorization evaluator
flattened User write model
cross-owner SQL
custom Role/Permission editor
operation outside 1–33
```

### EXIT

A real browser can authenticate and administer Organization/access through current server authority; every later protected tranche uses real auth rather than a bypass.

## 11. S2 — Document Governance configuration — exactly 10 operations

### ENTRY

S1 closed.

### APPLICATION ASSIGNMENT

```text
operations 34–43 exactly
```

### MUST BE IMPLEMENTED AT EXIT

```text
DocumentType identity/configuration real under Controlled Documents
base/governance/representation configuration real
eligible-Template configuration real without Template semantic owner
numbering preview non-reserving only
Template configuration projection real/disclosure-safe
10/10 operations execute real HTTP/application/owner/persistence
```

### FRONTEND STATE AT EXIT

Reviewed blueprint realized for:

```text
DocumentType list/create/base configuration
Governance + representation route editor
eligible Template set
numbering preview
Template configuration projection
```

Concrete per-Document Template-role mutation remains part of S3 because ordinary Documents are created there.

### PROOF

E2/E3 + claim-relevant E4. Negatives include stale ETags, invalid/retired references, unauthorized content inference, numbering preview as reservation and required Audit rollback.

### EXPLICIT ABSENCES

```text
Template peer lifecycle/API
generic workflow builder
number preview as number authority
implicit governance/representation default
operation outside 34–43
```

### EXIT

All governance configuration required for the complete Document/authoring tranche can be managed through the real Product.

## 12. S3 — Document core + creation + authoring + Submission + History — exactly 22 operations

### ENTRY

S2 closed and P3 content mechanisms needed by this flow are real enough for their E5 claims.

### APPLICATION ASSIGNMENT

```text
operations 44–54 except 55
+ operations 56–66
= 22 operations
```

Exact set:

```text
44 getDocumentCreationOptions
45 listDocuments
46 createDocument
47 getDocument
48 getDocumentResponsibleOwner
49 replaceDocumentResponsibleOwner
50 getDocumentTemplateRole
51 replaceDocumentTemplateRole
52 createDocumentRevision
53 getDocumentHistory
54 listAuthoringWork
56 getRevision
57 getRevisionDraft
58 updateRevisionDraft
59 startRevisionDraftUpload
60 completeRevisionDraftUpload
61 getRevisionDraftSource
62 createSubmission
63 getSubmission
64 getSubmissionSource
65 withdrawSubmission
66 cancelRevision
```

### MUST BE IMPLEMENTED AT EXIT

```text
creation options expose only actor-usable accepted references
Library official discovery real/default effective
Document create allocates final code + REV000/current truth atomically
createDocument idempotency + required Audit atomicity real
create success resolves directly into real Document Work — no dead target
Document Official core read real; DRAFT never presented as official
operator-approved T8-E-RO responsible-owner candidate projection real/disclosure-safe
responsible-owner relation + OCC + D4 target eligibility real
Document Template-role read/write real
create-next-Revision semantics real
listAuthoringWork projection real and targets live Work
Revision/open semantics real
DRAFT + exact strong ETag real
DRAFT title/source mutation under If-Match real
bounded upload allocation/provider PUT/completion/admission/attachment distinction real
ExactContentDescriptor remains content authority
exact DRAFT/submitted source reads real
Submission creation snapshots exact governed subject
withdrawal/cancellation semantics real
History operation real for all facts reachable through this tranche
22/22 operations execute end-to-end
```

### FRONTEND STATE AT EXIT

Reviewed blueprint realized end-to-end for:

```text
Library + filters
Create Document → Document Work
Document Official core
Responsible Owner management
Create/Open Revision
History
My Work authoring lane
Document Work DRAFT DOCX/PDF
upload/admission/source replacement
DRAFT OCC reconciliation
submit / submitted state / withdrawal / cancellation
concrete Document Template-role management
```

### PROOF

Close GF2 Document-create composition and GF3 with E2/E3/E4/E5 as claim-relevant; V4/V5/V6/V7 relevant portions. Required causal negatives include distinct concurrent code allocation, same-key changed create/submit, required Audit failure, stale owner/template/DRAFT ETags, target offboarding race, DRAFT-as-official leakage, expired upload, wrong size/hash/format, provider PUT without semantic attachment, malicious content and unauthorized exact-source disclosure.

### EXPLICIT ABSENCES

```text
Document.currentStatus authority
screen-shaped read API
History/Audit current truth
provider identity as Product content identity
EditorSession/lease correctness
silent stale merge/LWW
client content descriptor authority
Governance Work row before S4 target exists
Release/OfficialRendition claimed before S4
operation outside assigned 22
```

### EXIT

A real browser completes the entire accepted path:

```text
Library
→ create/open Document
→ live Document Work
→ DRAFT edit/upload
→ submit
→ submitted-state handling
```

with no deferred backend/frontend seam.

## 13. S4 — Governance work + Governance Case + Release/rendition — exactly 9 operations

### ENTRY

S3 closed and P3 River/renderer/MalwareInspector mechanisms required by accepted representation paths are real.

### APPLICATION ASSIGNMENT

```text
55 listGovernanceWork
67 getGovernanceAttempt
68 listGovernanceFeedback
69 createGovernanceFeedback
70 getGovernanceStepDecision
71 recordGovernanceStepDecision
72 getRelease
73 getReleaseSource
74 getOfficialRenditionContent
= 9 operations
```

### MUST BE IMPLEMENTED AT EXIT

```text
Governance Work projection real and targets live Case
GovernanceAttempt + immutable Step/candidate snapshot real
feedback + Step Decision under current Authorization
required owner mutation + Audit + durable intent commit coherently
River named work redelivery-safe when activated
submitted PDF + required PDF uses exact admitted bytes with no renderer transform
submitted DOCX + required PDF uses private renderer + re-admission + current eligibility revalidation
Release/EFFECTIVE system-owned canonical truth real
Release source + OfficialRendition exact-byte reads real
9/9 operations execute end-to-end
History enrichment for governance/release facts real
```

### FRONTEND STATE AT EXIT

Reviewed blueprint realized for:

```text
My Work governance lane
Submission Governance Case
Obsolescence Governance Case presentation base
feedback + Step Decision controls
read-only exact governed Submission content
Document Official Release/source/OfficialRendition viewer enrichment
History governance/release/rendition enrichment
```

### PROOF

Close GF4 with E2/E3/E5/E6 and E4 where browser claimed; V4/V6/V7/V8. Negatives include frozen-candidate drift, current AuthZ denial, Audit/intent failure, duplicate River delivery, renderer failure, fidelity failure, provider corruption, malware-gate failure and scan-evidence mismatch.

### EXPLICIT ABSENCES

```text
user publish/Release mutation
queue state as Product truth
reviewer mutation of governed Submission bytes
renderer output as authority before semantic admission/finalization
second governance/workflow owner
operation outside assigned 9
```

### EXIT

A real actor can move from Governance Work to an exact Case and canonical Release/Official presentation without any client or worker becoming semantic authority.

## 14. S5 — Obsolescence + Audit — exactly 4 operations

### ENTRY

S4 closed.

### APPLICATION ASSIGNMENT

```text
75 createObsolescenceRequest
76 getObsolescenceRequest
77 withdrawObsolescenceRequest
78 listAuditEvents
```

### MUST BE IMPLEMENTED AT EXIT

```text
obsolescence create/read/withdraw real
NoHumanApproval synchronous completion has no fake human Step
human-governed request routes through existing S4 Governance Case
current target remains EFFECTIVE until accepted completion
Official/Library/History truth reflects result
Audit evidence list real/disclosure-safe
Audit remains evidence only
4/4 operations execute end-to-end
```

### FRONTEND STATE AT EXIT

Reviewed blueprint realized for:

```text
Document Official obsolescence management
active / returned / withdrawn / completed request states
obsolescence Governance Case subject interaction
History obsolescence enrichment
Audit evidence list
```

### PROOF

Close GF5 with E2/E3/E4 and relevant V3/V4/V6. Negatives include unauthorized absence inference, stale active-request routing, DRAFT-as-official leakage, History/Audit disclosure bypass and any operation79 attempt.

### EXPLICIT ABSENCES

```text
fake System approver
fake human Step under NoHumanApproval
Audit current lifecycle authority
History current resolver
operation 79
operation outside 75–78
```

### EXIT

Official discovery, obsolescence, history and Audit compose coherently without a second current-truth path.

## 15. P4 — Runtime / durable-work / recovery closure

### ENTRY

S1→S5 closed plus P3 runtime mechanisms real.

### MUST BE IMPLEMENTED AT EXIT

```text
all activated named River work classes prove duplicate/redelivery safety
terminal required work visible/inspectable/redrivable
renderer/scanner/OIDC/PostgreSQL blast radii match accepted runtime law
/livez and /readyz behavior proven
resource ceilings/time/memory/scratch/concurrency proven
SIGTERM unready-first bounded drain proven
backup capture + exact-content manifest/pin/GC exclusion proven
restore invalidates sessions and completes security/privacy reconciliation before readiness
observability material failures attributable without secret/PII leakage
```

Frontend contribution is only truthful/sanitized dependency/integrity states; there is no Product restore/jobs control surface.

### PROOF

Close GF6 and V8/V9/V10 with real E2/E5/E6 subjects.

### EXPLICIT ABSENCES

```text
generic retry-jobs Product UI
queue state as business truth
restore Product operation
silent degraded serving after readiness failure
```

### EXIT

The full implementation can survive the accepted runtime/recovery falsifiers and is eligible to be instantiated as private B1 target.

## 16. B1 + P5 — exact private target / whole implementation proof closure

T10 B1 privately provisions **the exact implementation candidate** using only accepted components.

P5 runs against that exact target/profile.

### P5 MUST PROVE

```text
6/6 Golden Flows positive + causal negative lanes
10/10 T9 cross-cutting properties
T8-E runtime wire-conformance fixture classes on real composed HTTP path
78/78 application operations in canonical wire census
0 orphaned / 0 invented operations
operation 79 absent
10/10 Idempotency-Key creation semantics
13/13 ETag read/mutation domains
4/4 exact-byte resources
closed-world dependency graph on live tree
all required same-local-commit Audit classes represented
all claim-relevant external mechanisms have real proof lanes
frontend matches reviewed T11 blueprint
```

Frontend whole-candidate closure includes:

```text
16/16 accepted human goals
36/36 material surfaces
10/10 stable SPA Product route meanings
no frontend Authorization engine
no handwritten parallel DTO/API authority
no global server-truth mirror
no screen-shaped operation
T8-E-RO candidates disclosure-safe/complete/non-authoritative
```

### EXIT

The exact private target is eligible for T10 B2 proof + verified clean seal. P5 itself creates no business authority.

## 17. T10 barrier overlay

T11 never replaces accepted cutover law:

```text
B0 source truth
→ B1 private target
→ B2 exact production candidate proven + clean seal
→ B3 first authoritative Product mutation / point of no return
→ B4 authoritative recovery point + serving fence + canonical activation
```

Mapping:

```text
B0
  current before target preparation is treated as cutover

B1
  exact private realization after implementation/runtime preparation

P5
  full implementation/T9 proof against exact B1 target

B2
  exact candidate passes proof + operations/provenance clean seal

B3
  first post-seal authoritative Product commit through accepted non-serving bootstrap concern

B4
  authoritative recovery point exists + disposable serving estate fenced + canonical R10 serving activated
```

Post-B3 recovery is forward on R10 authority only.

## 18. PR / implementation slicing law

A graph node may span multiple PRs. Split only at independently coherent ownership/proof seams.

Required:

```text
no direct main commits
no Product implementation before P0
no speculative cross-node scaffolding with no current consumer
no dormant tables/endpoints/workers/frameworks
no hand-edited generated projections
no knowingly broken previously closed invariant
no user-visible control pointing to an impossible downstream target
```

S3 may be multiple PRs internally, for example persistence/owner → wire/application → frontend, **only if intermediate merges remain non-user-visible or independently coherent**. The final S3 exit is still the complete create→Work→submit vertical journey.

## 19. T11 frontend implementation blueprint

The complete F1→F9 result is consolidated in:

```text
docs/work/current/t11-frontend-blueprint.md
```

That companion owns implementation-readiness detail only:

```text
capability/human-goal coverage
36 material surfaces
screen contracts
navigation/data graph
functional wireframes
material interaction ledger
78-operation bidirectional trace
frontend-specific findings/reopen triggers
```

It creates no Product/frontend semantic authority beyond accepted T6/T8-E/T8-F and the operator-approved bounded precision:

```text
docs/decisions/responsible-owner-selection-read.md
```

## 20. Stop / reopen routing

```text
implementation/proof defect
→ fix owning node; architecture stays closed

accepted mechanism assumption falsified
→ STOP; smallest implicated T8 technical owner

missing/extra application operation or consumer
→ STOP; smallest Product/T6/T8-E owner; operation79 never silent

semantic owner/lifecycle contradiction
→ STOP; smallest Product/T1→T8 owner

frontend accepted human goal not representable by current read/write truth
→ STOP that surface; smallest justified T6/T8-E/T8-F precision

real pre-R10 authoritative business truth discovered
→ STOP cutover; smallest T7/T10 owner

graph dependency/proof boundary shown unsound
→ correct T11 before ratification / bounded T11 reopen later
```

Preference, file split, patch-version churn or implementation convenience are not reopen triggers.

## 21. T11 closure contract

T11 candidate can proceed to independent review only when:

```text
implementation DAG has no unresolved dependency cycle or dead human-flow seam
all P/S node exit contracts are exact
78/78 application operations assigned exactly once
operation 79 absent
10 Idempotency-Key creations owned/proven
13/13 ETag domains owned/proven
4 exact-byte resources owned/proven
6/6 Golden Flows have implementation/proof owner
10/10 V properties have implementation/proof owner
frontend blueprint complete and bidirectionally traced
0 unresolved MATERIAL frontend finding
T8-E-RO precision exactly specified/routed
T10 B0→B4 law preserved
no T12 work begun
no Product implementation begun
required CI green on exact candidate
explicit operator approval granted before independent Fable review
```

After bounded independent review convergence, T11 durable promotion must:

```text
promote this implementation-program outcome to durable architecture authority
promote/retain the minimum frontend blueprint needed by T12/future implementation
retain approved bounded precision provenance without duplicate contradictory authority
remove superseded temporary T11 work artifacts
record immutable T11 ratification evidence
keep Product implementation BLOCKED until T12 + remaining final roadmap gates close
```

T11 closure does **not** authorize implementation and does not open T12 automatically.
