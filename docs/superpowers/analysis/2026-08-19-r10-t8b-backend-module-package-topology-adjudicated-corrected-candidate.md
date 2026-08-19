# R10-T8B — Backend Module & Package Topology — Adjudicated Corrected Candidate

```text
ADJUDICATED CORRECTED CANDIDATE
NON-AUTHORITATIVE STAGING
ROUND-2 FABLE INPUT
NOT TARGET AUTHORITY
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Baseline HEAD before this staging artifact:** `254d98a9c1481f1b2f1a7573b820be90ef6146d8`  
> **Round-1 Fable review:** `docs/superpowers/analysis/2026-08-19-r10-t8b-backend-module-package-topology-independent-fable-review.md`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Stage:** T8-B ACTIVE  
> **Implementation:** BLOCKED

This artifact materializes the Lead's adjudicated correction of the T8-B candidate after the independent Fable Round-1 review. It is **review input only**. It does not close T8-B, amend durable authority, update the Decision Registry, update the router, open T8-C or authorize implementation.

The operator approved the Lead adjudication before this materialization. Durable promotion still requires the bounded Round-2 review to converge, final Lead adjudication and the repository's normal operator-ratification/promotion flow.

---

## 0. Authority and review posture

Repository authority remains local. The governing order is the one routed by `AGENTS.md` and `wiki/architecture/r10-technical-architecture.md`.

Binding stage law:

```text
T8-B may freeze:
  target repository/package layout
  semantic-owner realization boundaries
  layering within owners
  public/internal Go package surfaces
  allowed dependency graph
  forbidden dependency graph
  composition root / dependency injection
  location/law of shared mechanisms

T8-B may name a required seam class to justify dependency direction.
T8-B must NOT invent the detailed T8-C contract by stealth.
```

Binding T8-A realization law remains:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

Current implementation is evidence only.

---

## 1. Lead adjudication verdict

Round-1 Fable verdict:

```text
APPROVE R10 T8-B BACKEND MODULE & PACKAGE TOPOLOGY
WITH MATERIAL FIXES

BLOCKER  4
MAJOR    8
LOW      6
```

Lead adjudication:

```text
GLOBAL MAXIMUM CLASS                  CONFIRMED
ORIGINAL T8-B CANDIDATE               BOUNDED CORRECTION REQUIRED
MATERIALLY BETTER TOPOLOGY CLASS      NOT FOUND
UPSTREAM REOPEN                       NONE
T8-C                                  NOT OPEN
IMPLEMENTATION                        BLOCKED
```

The corrected candidate preserves the selected architecture class:

```text
ONE GO MODULE FOR BACKEND GO CODE
+ OWNER-FIRST MODULAR MONOLITH
+ NON-SEMANTIC APPLICATION ORCHESTRATION
+ SINGLE INBOUND ADAPTATION PATH
+ NON-SEMANTIC TECHNICAL MECHANISMS
+ WIRING-ONLY COMPOSITION ROOT
+ COMPLETE MECHANICAL DEPENDENCY ENFORCEMENT
```

The Round-1 findings exposed missing realization laws and ambiguous wording. They did **not** justify replacing the topology class.

---

## 2. Alternatives and Global Maximum

### Alternative A — direct owner-to-owner imports

Rejected.

T3 makes Controlled Documents commands authorization-composed while Authorization decisions consume owner-owned resource relationship/state predicates. Direct reciprocal imports therefore create immediate cycle pressure between the two largest semantic owners and then expand through Organization/Audit.

Local benefit:

```text
fewer orchestration hops
```

Global cost:

```text
reciprocal semantic coupling
owner SCC pressure
authority leakage
foreign persistence/read shortcuts
difficult mechanical enforcement
```

### Alternative B — one-module owner-first modular monolith + orchestration + adapters

**Selected Global Maximum class.** Corrected below.

### Alternative C — one Go module per semantic owner

Rejected on current evidence.

No independent external Go consumer, release lifecycle, repository/trust boundary or deployment boundary exists. Go `internal` plus complete import-graph enforcement provides owner-private isolation without module-versioning and multi-module coordination cost.

Reopen only on concrete evidence such as:

```text
independent external Go consumer
independent release train
repository/trust boundary
required independent build provenance
deployment boundary that materially benefits from module separation
```

### Alternative D — use cases hosted inside primary owners, application only for multi-owner writes

Rejected.

It creates two inbound rules:

```text
sometimes transport/application → owner
sometimes transport/application → cross-owner orchestrator
```

That distinction becomes semantic knowledge distributed across callers and weakens the single composition point for authorization/read-model behavior.

### Alternative E — separate `policy` semantic package for Authorization + domain predicate composition

Rejected.

T3 already owns final ALLOW/default-DENY semantics in Authorization. A separate policy semantic home would duplicate authority.

### Global Maximum

```text
Alternative B, corrected by this artifact
```

No materially superior fourth topology class survived the Method challenge.

---

## 3. Target repository topology

This is the **backend target projection**. Repository roots not shown are not implicitly deleted or redesigned by T8-B. Frontend and non-Go runtime realization remain owned by later stages where routed.

```text
MetalDocs/
├── go.mod                              # one Go module for backend Go code
│
├── api/
│   └── openapi/
│       └── v1/
│           └── openapi.yaml            # contract SSOT; exact wire content = T8-E
│
├── cmd/
│   └── <runtime-shells>/               # names/count/process topology = T8-G
│
├── db/                                 # DB/migration/bootstrap asset home
│                                        # exact schema/content = T8-D
│
├── internal/
│   ├── authentication/                 # semantic owner — one public surface
│   ├── organization/                   # semantic owner — one public surface
│   ├── authorization/                  # semantic owner — one public surface
│   ├── controlleddocs/                 # semantic owner — one public surface
│   ├── audit/                          # supporting semantic owner — one public surface
│   │
│   ├── application/                    # stateless leaf use-case/lens orchestration
│   │   ├── library/
│   │   ├── mywork/
│   │   ├── documentofficial/
│   │   ├── documentwork/
│   │   ├── governancecase/
│   │   ├── history/
│   │   ├── audit/
│   │   └── admin/
│   │
│   ├── transport/
│   │   ├── http/                       # inbound HTTP/browser adapters
│   │   └── jobs/                       # durable-job inbound adapters
│   │
│   ├── platform/                       # non-semantic mechanisms only
│   │   ├── txscope/                    # provider-neutral transaction-scope abstraction
│   │   ├── postgres/                   # pool/driver/bootstrap + txscope realization
│   │   ├── managedcontent/             # managed content + malware inspection mechanisms
│   │   ├── identityprovider/           # IdP protocol client only
│   │   ├── officialrendition/          # server-side OfficialRendition mechanism only
│   │   ├── river/                      # durable-job mechanism
│   │   ├── idempotency/                # opaque replay mechanism
│   │   ├── observability/
│   │   └── config/
│   │
│   └── composition/                    # construction/wiring only
│
├── tests/                              # test roots; exact T9 validation content later
│
└── tools/
    ├── verify/                         # verification registry / profile SSOT
    ├── cilint/                         # architecture/static-analysis analyzer host
    └── codegen/                        # contract/code-generation tooling
```

Important non-decisions:

```text
runtime shell count/names                      T8-G
exact DB layout/schema/constraints             T8-D
exact OpenAPI operations/schemas               T8-E
frontend source topology                        T8-F
Node/TS renderer/runtime deployment topology    T8-G
current→target moves/deletions                   T10
```

`cmd/*` means runtime entrypoints only. Code generators are tooling and do not inherit the runtime dependency law.

---

## 4. Semantic owner realization law

Exactly five semantic homes exist for Launch:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit — supporting evidence authority
```

### 4.1 One public surface per owner

The corrected unit is **public surface**, not package count.

```text
Each semantic owner exposes exactly ONE importable public package path.
```

A semantic owner may freely use files or owner-private nested packages:

```text
internal/<owner>/internal/<responsibility>
```

No architecture approval/evidence gate is required for purely owner-private decomposition because it cannot create another cross-owner semantic authority or public dependency path.

Example:

```text
internal/controlleddocs/                         public owner surface
internal/controlleddocs/internal/revision/       private
internal/controlleddocs/internal/governance/     private
internal/controlleddocs/internal/content/        private
internal/controlleddocs/internal/release/        private
```

The exact private decomposition is implementation-local and may evolve without reopening T8-B as long as the public surface/dependency law remains true.

### 4.2 Second public surface rule

A **second public package path** for one semantic owner is material architecture and requires a named consumer plus proof that the single public surface cannot sustainably serve it.

Entity/family names never justify a second public surface by themselves.

### 4.3 No false peer owners

No Launch peer semantic package/owner is created for:

```text
Approval
Templates
Artifact
Search
Distribution
Periodic Review
Taxonomy
Notifications
Interchange
Records
generic Workflow
```

Reasons remain those already ratified by Product Contract/GCR/T1→T6:

```text
approval/governance/release     inside Controlled Documents authority
Template                        ordinary governed Document role
Artifact                        removed as semantic owner
Search                          discovery/read mechanism, not authority
Distribution/Periodic Review    Launch+
Records/etc.                    Future absent Launch consumer
```

Future law:

```text
prepare the seam, not dormant implementation
```

---

## 5. Application orchestration law

`internal/application/*` is **not a semantic owner**.

It is a set of stateless leaf orchestration packages aligned to ratified T6 use-case/lens vocabulary.

### 5.1 Allowed responsibilities

An application leaf may:

```text
coordinate one application/business operation
open/commit/rollback the shared local transaction scope
invoke semantic-owner public surfaces
gather owner-authored predicate facts/results for Authorization
coordinate required same-transaction Audit append
coordinate justified idempotency/durable intent
return purpose-built use-case/read-model results
```

### 5.2 Forbidden responsibilities

An application leaf must not:

```text
own User/Role/Permission/Document lifecycle state
reimplement semantic owner rules
compute the canonical Authorization ALLOW/default-DENY equation itself
invent owner-intrinsic Audit meaning
become current-state authority
contain owner semantic SQL
become a generic workflow/event-bus/domain platform
hold persistent or process-global state
import another application leaf package
```

Dependency law:

```text
application/<a> → application/<b>    FORBIDDEN
owner           → application         FORBIDDEN
```

This prevents a hidden `application/core` God orchestrator while preserving cross-owner choreography.

Package leaf names may be refined only within already-ratified lens/use-case meaning; frontend feature realization remains T8-F.

---

## 6. Transport law — one inbound door

Inbound adapters are technical translation only.

```text
transport/http  → application
transport/jobs  → application
```

Forbidden:

```text
transport → semantic owner public surface
transport → semantic owner private implementation
transport → SQL/persistence
transport → transaction-scope creation
transport → business rules
```

There is one inbound composition door for commands and queries:

```text
transport → application
```

This keeps Authorization/domain-predicate composition and allowed-actions/read-model behavior on one provable path.

Durable job handlers are inbound adapters. They reload canonical state and do not become semantic/job-state authorities.

---

## 7. Composition law

`internal/composition` owns construction only.

Allowed:

```text
construct semantic owners
construct application leaves
construct transport adapters
construct platform mechanisms/adapters
bind interfaces/seams to implementations
assemble runtime dependency graph for cmd shells/tests
```

Forbidden:

```text
business conditional logic
request-scoped behavior
semantic decisions
current-state authority
SQL/business queries
use-case orchestration
```

Runtime process count remains T8-G.

---

## 8. Platform law — mechanism never authority

`internal/platform/*` contains replaceable non-semantic technical mechanisms.

The corrected law is over **imports/types/authority**, not only SQL.

### 8.1 General platform law

Forbidden:

```text
platform package → any semantic owner package import
platform signature/state → semantic owner type as platform-owned meaning
platform → semantic SQL authority
platform → cross-owner business query
platform → shared semantic write authority
platform → product workflow/state-machine ownership
```

Allowed:

```text
primitive/mechanism values
provider SDK/protocol types kept inside the relevant mechanism boundary
consumer-owned technical port types
provider-neutral mechanism abstractions
```

Semantic identifiers may cross a mechanism seam only as opaque values required by the consumer-owned contract; the mechanism does not acquire their business meaning.

### 8.2 `platform/postgres`

Owns only technical PostgreSQL mechanism concerns such as:

```text
pool/driver lifecycle
connection/bootstrap primitives
transaction implementation mechanics
```

Does not own:

```text
Organization SQL
Authorization SQL
Controlled Documents SQL
Audit SQL
foreign/cross-owner SQL
```

Positive corollary:

```text
owner-specific persistence adapters are owner-private
```

Their exact locations/schema/query design remain T8-D.

### 8.3 `platform/txscope`

See §9. It is provider-neutral transaction participation infrastructure, not a semantic owner.

### 8.4 `platform/managedcontent`

Consumes T4 law:

```text
mechanism does not parse/own Document/Revision/Submission semantics
no owner_type/owner_id generic semantic registry
exact-content descriptors remain owned by semantic records
malware inspection remains a technical admission mechanism
```

### 8.5 `platform/identityprovider`

Owns protocol mechanics only, for example:

```text
OIDC discovery
authorization/token exchange
JWKS/protocol verification
raw provider protocol response handling
```

Authentication owns the IdP anti-corruption boundary:

```text
provider-subject meaning
provider claim → MetalDocs identity interpretation
ProviderSubjectBinding semantics
ApplicationSession issuance/revocation
authentication assurance/fresh-auth semantics
```

Provider roles/groups/organizations/permissions never become canonical MetalDocs Authorization.

Raw provider claims/protocol types must not leak from the protocol + Authentication boundary into Organization, Authorization, Controlled Documents, Audit or business/read-model surfaces.

### 8.6 `platform/officialrendition`

Represents the server-side OfficialRendition rendering mechanism only.

Interactive DOCX editor/viewer realization is a different concern and is not collapsed into this backend mechanism. Its exact realization remains for the later frontend/runtime stages.

### 8.7 `platform/river`

River remains the selected durable-job mechanism only for named T5 consumers. It does not own semantic truth, generic event-bus authority or product workflow state.

### 8.8 `platform/idempotency`

Idempotency is an opaque replay mechanism.

```text
platform/idempotency never authorizes replay disclosure
application performs live Authorization/resource-visibility recheck before disclosure
replay persistence must be erasure-safe and must not become an unintended PII retention root
```

Exact contract, record schema, retention and whether erasure-safety is achieved by PII-free representation vs purge/redaction remain T8-C/T8-D.

### 8.9 Observability/config

Remain technical mechanisms and may not become business authority or semantic registries.

---

## 9. Required transaction-scope seam class

Round-1 B-02 identified a real T8-B omission: the package/dependency graph is incomplete unless one shared local transaction has a named technical carrier class.

### 9.1 Frozen T8-B law

```text
transaction-scope abstraction home   = internal/platform/txscope
semantic status                      = mechanism only
scope lifecycle owner                = application
semantic owner participation         = explicit in owner write surfaces
provider identity leakage            = forbidden
ambient/hidden transaction ownership = forbidden
```

Application:

```text
opens shared local scope
coordinates all participants
commits or rolls back
```

Semantic owners:

```text
participate in the caller-provided scope
must not open/commit/rollback an independent scope for one composed transition
must not expose/require *sql.Tx, *pgx.Tx, pool/driver concrete types on public surfaces
```

Transport must not open transaction scope.

### 9.2 Deferred deliberately

T8-B does not freeze:

```text
exact Go interface/type/method names          T8-C
isolation level realization                   T8-D
lock modes/order                              T8-D
Document serialization primitive realization T8-D
PostgreSQL concrete transaction mapping       T8-D
```

The seam exists to make dependency direction and atomic composition decidable; its detailed contract belongs to T8-C/T8-D.

---

## 10. Required same-transaction Audit evidence seam class

Round-1 B-01 correctly identified that T3 same-local-commit Audit must be structurally realizable. The review's exact `AuditAppender` prescription is **not** adopted as T8-B contract authority.

### 10.1 Frozen T8-B law

```text
owner → audit direct import            FORBIDDEN
audit → other owner direct import      FORBIDDEN
owner-intrinsic evidence meaning       produced/owned by the mutating semantic owner
cross-owner routing/composition        application
Audit append                           must complete inside the same local transaction
application commit                     may occur only after required Audit append succeeds
Audit                                  remains evidence, never lifecycle/current-state authority
```

A compliant realization therefore has a bounded evidence handoff seam:

```text
owner mutation
→ owner-authored required evidence facts/result
→ application coordinates Audit append in the same tx scope
→ commit
```

Application transports/composes the evidence but must not invent owner-intrinsic event meaning.

### 10.2 Deferred deliberately

T8-C owns:

```text
exact evidence handoff interface
exact returned/callback/port pattern
exact fact/result types
```

T8-D owns persistence realization of the same-transaction guarantee.

T9 must prove a required governed/security mutation cannot commit without its required Audit evidence.

---

## 11. Required Authorization/domain-predicate decision seam class

Round-1 B-04 identified a real contradiction in a zero-owner-import graph unless the substitute composition route is named.

Binding authority partition remains:

```text
Authorization       owns grants, scope composition and final ALLOW/default-DENY decision
business owner      owns resource relationship/state/governance predicate meaning
application         owns choreography only
```

### 11.1 Frozen T8-B direction

```text
application obtains owner-authored predicate facts/results
application supplies required facts/context to Authorization
Authorization performs the canonical access decision
Authorization returns ALLOW / DENY
```

Forbidden:

```text
Authorization → Controlled Documents import
Controlled Documents → Authorization import
application computing the canonical ALLOW/DENY equation itself
second authorization evaluator elsewhere in the tree
```

The same decision authority is the basis for T6 `allowed_actions`; no hand-maintained second ruleset is permitted.

For commands that require current canonical truth inside the transaction, the decision path must be able to participate in the same transaction scope from §9.

### 11.2 Deferred deliberately

T8-C owns:

```text
exact Authorization decision interface
exact owner predicate fact vocabulary
exact actor/scope/context/result shapes
exact transaction participation signature
```

T8-B freezes only authority and dependency direction.

---

## 12. Allowed dependency graph

```text
cmd/<runtime-shell> → composition

transport/http      → application leaves
transport/jobs      → application leaves

application leaf    → semantic owner public surfaces
application leaf    → Authorization public decision surface
application leaf    → Audit public surface
application leaf    → platform/txscope
application leaf    → platform/idempotency where the use case requires it

semantic owner      → its own private implementation
semantic owner      → its own consumer-declared technical seams
semantic owner      → platform/txscope abstraction where transaction participation requires it

platform adapter    → external SDK/protocol/mechanism dependencies

composition         → public constructors/surfaces and adapters it wires

tools/*             → tooling dependencies only; no product runtime authority
```

The exact consumer-declared interfaces are T8-C where they are inter-owner/mechanism contracts; T8-B freezes their direction/class only.

---

## 13. Forbidden dependency graph

```text
semantic owner      → another semantic owner, public or private
semantic owner      → application
semantic owner      → composition
semantic owner      → concrete DB/provider/SDK adapter by convenience
semantic owner      → connection pool / driver concrete type

transport           → any semantic owner package
transport           → persistence/SQL
transport           → transaction-scope creation

application/<a>     → application/<b>
application         → owner-private implementation
application         → semantic SQL
application         → duplicate semantic state/rules

platform            → semantic owner package
platform            → semantic owner current-state authority
platform            → cross-owner semantic SQL/query
platform            → shared semantic write authority

composition         → business/request logic

foreign SQL as owner communication
common/shared/utils/models/repositories/services/domain as cross-owner semantic dumping grounds
second evaluation of canonical Authorization semantics
provider roles/groups/claims as MetalDocs Organization/Authorization truth
```

Direct owner-to-owner imports are **forbidden**, not merely discouraged. A future exception requires a named contradiction showing the zero-direct-import model is materially worse; convenience is not a reopen trigger.

---

## 14. Mechanical enforcement and proof

Documentation-only dependency laws are insufficient. The target requires a closed-world, falsifiable architecture classifier.

### 14.1 Tooling disposition

T8-A five-part gate applied to current verification machinery:

```text
tools/verify registry/profile SSOT property       PRESERVE
tools/cilint generic analyzer harness             PRESERVE / REFINE
legacy target-specific analyzer policy            REWRITE
legacy policy fixtures/baselines                  REWRITE for target semantics
```

Target homes:

```text
tools/verify   verification registry / profile SSOT
tools/cilint   architecture/static-analysis analyzer host
tools/codegen  contract/code-generation tooling
```

Tooling lives outside `internal/` and must not become product authority.

### 14.2 Closed-world package classification

Every Go package in the target backend universe must map to exactly one architecture class:

```text
semantic-owner-public
semantic-owner-private
application
transport
platform
composition
runtime-entrypoint
tooling
test
```

Proof is bidirectional:

```text
tree → architecture catalog
architecture catalog → tree
```

An unmapped package fails verification. Silence is not a classification.

### 14.3 Import/dependency proofs

Target analyzers must reject at minimum:

```text
forbidden import edges
semantic-owner SCCs
foreign owner-private reachability
transport bypass of application
owner dependency on application/composition/concrete adapters
prohibited peer semantic roots
platform importing/naming semantic owner packages as authority
application leaf → application leaf
foreign/cross-owner SQL ownership
unmapped target packages
```

Same-transaction Audit and Authorization single-source requirements require proof at the appropriate later stage once their exact contracts exist; T8-B records them as mandatory proof obligations, not fake current analyzers.

### 14.4 Negative fixtures

Every mechanical rejection class must have a fixture proven to fire against the forbidden shape.

A verifier that merely stays green on the accepted tree is not sufficient proof.

### 14.5 Exception decay

Any temporary architecture exception must:

```text
be explicit
have a named reason
have a removal trigger
fail/alert when the underlying violation no longer exists
```

Target state should prefer no allowlist where a clean graph makes exceptions unnecessary.

---

## 15. Legacy/current disposition

```text
legacy `internal/modules/*` semantic topology    REWRITE / REHOME
legacy peer Approval/Templates/etc. modules     DELETE / REHOME by Launch ownership
legacy composition placement                     concept survives; contents not inherited
current platform packages                        selective reuse only through T8-A five-part gate
current runtime `apps/*/cmd` shape               current evidence; target process topology T8-G
current generators under root `cmd/`             REHOME to tooling class; exact move T10
current `db/` assets                              target root survives as class; exact content T8-D/T10
current architecture analyzer harness             selective harness reuse
legacy analyzer policy/allowlists                 REWRITE / RETIRE
OpenAPI contract-first property                   PRESERVE; exact contract T8-E
one Go module                                      independently reselected
```

T8-B does not authorize deleting or moving any current code. T10 owns the transition/deletion sequence.

---

## 16. Decision disposition — adjudicated T8-B set

```text
T8B-D01  ACCEPT
  one Go module for backend Go code

T8B-D02  ACCEPT
  semantic owner public roots = authentication / organization / authorization /
  controlleddocs / audit

T8B-D03  ACCEPT WITH CORRECTION
  exactly one importable PUBLIC surface per semantic owner

T8B-D04  ACCEPT WITH CORRECTION
  owner-private nested packages/files are freely decomposable;
  architecture gate applies to a SECOND public surface, not private package count

T8B-D05  ACCEPT WITH CORRECTION
  application = stateless leaf orchestration aligned to ratified use-case/lens boundaries;
  no application→application imports

T8B-D06  ACCEPT WITH CORRECTION
  direct semantic-owner→semantic-owner imports are forbidden;
  required substitute seam classes are named by T8-B, exact contracts by T8-C

T8B-D07  ACCEPT WITH CORRECTION
  transport depends on application ONLY; owner public surfaces are not a transport bypass

T8B-D08  ACCEPT
  composition = wiring/construction only; no business/request behavior

T8B-D09  ACCEPT WITH CORRECTION
  platform = technical mechanisms only; law covers imports/types/authority, not only SQL

T8B-D10  ACCEPT
  platform/postgres owns no semantic owner SQL;
  owner-specific persistence adapters are owner-private

T8B-D11  ACCEPT
  no generic common/shared semantic dumping ground

T8B-D12  ACCEPT
  no Launch peer semantic package for Approval

T8B-D13  ACCEPT
  no Launch peer semantic package for Templates

T8B-D14  ACCEPT
  no Launch peer semantic package for Artifact

T8B-D15  ACCEPT
  no Launch semantic owner for Search; Library read orchestration is application-level

T8B-D16  ACCEPT
  no dormant Launch+/Future packages

T8B-D17  ACCEPT
  legacy module topology has no inheritance entitlement

T8B-D18  ACCEPT
  selective reuse remains subject to all five T8-A gates

T8B-D19  ACCEPT
  Go nested `internal` visibility participates in owner-private enforcement

T8B-D20  ACCEPT WITH CORRECTION
  target architecture enforcement = tools/cilint analyzer host + tools/verify registry,
  complete/bidirectional classification + negative fixtures + exception decay

T8B-D21  ACCEPT
  exact inter-owner/mechanism contracts remain T8-C;
  T8-B may name seam classes/directions only

T8B-D22  ACCEPT
  exact persistence remains T8-D

T8B-D23  ACCEPT
  runtime/process/deployment topology remains T8-G

T8B-D24  ACCEPT
  transition/deletion sequencing remains T10

T8B-D25  ADD — ACCEPT
  one provider-neutral transaction-scope mechanism class is required;
  home = platform/txscope; application owns scope lifecycle;
  owner writes participate explicitly; exact contract T8-C/T8-D

T8B-D26  ADD — ACCEPT
  Authorization remains sole final ALLOW/default-DENY authority;
  application gathers owner-authored predicate facts/results and routes them;
  exact decision contract/fact vocabulary T8-C

T8B-D27  ADD — ACCEPT WITH LEAD CORRECTION
  same-transaction Audit evidence seam is required;
  no direct owner↔Audit imports;
  owner owns evidence meaning, application coordinates Audit append before commit;
  exact handoff contract T8-C

T8B-D28  ADD — ACCEPT
  repository target projection explicitly homes DB assets, tests and tooling;
  generators are tooling, not runtime cmd shells

T8B-D29  ADD — ACCEPT
  IdP protocol mechanism lives in platform/identityprovider;
  provider anti-corruption semantics live in Authentication

T8B-D30  ADD — ACCEPT
  backend OfficialRendition mechanism is distinct from interactive editor/viewer realization

T8B-D31  ADD — ACCEPT
  idempotency is opaque/non-authoritative; replay disclosure is live-authorized by application;
  storage must be erasure-safe; exact realization T8-C/T8-D
```

No decision above changes product semantics or opens a Future capability.

---

## 17. Round-1 finding adjudication summary

```text
B-01  FINDING ACCEPTED
      exact reviewer AuditAppender prescription NOT frozen;
      bounded same-tx evidence seam frozen, exact contract → T8-C

B-02  ACCEPTED — MATERIAL T8-B OMISSION
      transaction-scope class/home/direction now frozen

B-03  ACCEPTED
      transport → application ONLY now explicit

B-04  ACCEPTED — MATERIAL T8-B OMISSION
      Authorization final-decision authority + fact-routing direction now frozen;
      exact fact/interface contract → T8-C

M-01  ACCEPTED
      platform law covers semantic imports/types/authority

M-02  ACCEPTED WITH PRECISION
      protocol client mechanism vs Authentication anti-corruption meaning separated

M-03  ACCEPTED
      one public surface per owner replaces one package per owner

M-04  ACCEPTED
      application becomes stateless leaf packages; no intra-application imports

M-05  ACCEPTED
      closed-world completeness + bidirectional proof + exception decay

M-06  ACCEPTED
      tools/cilint = analyzer host; tools/verify = verification SSOT;
      harness reuse only, policy/fixtures rewritten

M-07  ACCEPTED WITH LEAD CORRECTION
      explicit target homes added;
      generators → tools/codegen, not runtime cmd

M-08  FINDING ACCEPTED
      exact purge/redact prescription NOT frozen;
      erasure-safe invariant frozen, realization → T8-D

L-01  ACCEPTED
      this staging artifact closes candidate traceability gap

L-02  ACCEPTED
      wording = one Go module for backend Go code

L-03  ACCEPTED WITH LEAD CORRECTION
      OfficialRendition backend mechanism separated from editor/viewer concern

L-04  ACCEPTED
      managed-content non-semantic law explicit

L-05  ACCEPTED
      Library/Search orchestration has application home, no Search owner

L-06  ACCEPTED
      owner persistence adapters stated affirmatively as owner-private
```

---

## 18. Successor-stage obligations — recorded, not designed

### T8-C — Internal Communication Contracts

Must freeze exact contracts for the seam classes named here:

```text
transaction-scope type/method/signature contract
same-transaction owner-evidence → Audit handoff contract
Authorization decision contract + owner predicate fact vocabulary
consumer-owned mechanism ports where material
transaction-coupled durable-intent contract
```

T8-C must preserve:

```text
no direct owner→owner imports
application = choreography, not authority
Authorization = single final access-decision authority
owner = source of its own relationship/lifecycle predicate meaning
Audit = evidence only
```

### T8-D — Persistence

Owns:

```text
schema/table/index/constraint mapping
persistence adapter exact package/file layout inside owner-private code
transaction implementation mapping
isolation/locking/serialization mechanics
idempotency replay record + erasure realization
```

### T8-E

Exact OpenAPI and generated wire contracts. `allowed_actions` must derive from the same canonical Authorization decision semantics, not a second ruleset.

### T8-F

Frontend realization. Do not infer React feature folders one-for-one from Go application leaf package names.

### T8-G

Runtime shell count, process/deployment topology and placement of non-Go runtime services.

### T8-H

Whole-T8 coherence proof.

### T9

Must prove at minimum:

```text
forbidden import edges fire
architecture classification is complete
semantic-owner SCC = zero
transport bypass impossible
same-tx Audit omission impossible for required mutations
no second Authorization evaluator
cross-owner atomic operations share one local transaction
```

### T10

Owns actual current→target moves, rewrites, deletions and retirement of legacy analyzer policies/allowlists.

---

## 19. Future-evolution check

This topology is intentionally additive for foreseeable future capabilities without implementing them now.

Expected attachment law:

```text
future semantic capability
→ receives its own owner only when Product authority introduces that semantic meaning
→ integrates through application/contract seams
→ does not require current Launch owners to be dismantled
```

Examples checked conceptually:

```text
Distribution / Read & Acknowledge
Periodic Review
Evidence / Records
Retention / Legal Hold
Governed Export
multi-document Change Control
CRDT/realtime authoring
```

None justifies a dormant Launch package today.

Pooled tenancy remains Future. The transaction-scope seam is justified entirely by current local atomic composition; no tenant semantics are prebuilt into it.

---

## 20. Reopen triggers

Reopen a T8-B decision only on material evidence such as:

```text
real external/independently released Go consumer              → reconsider D01
single owner public surface cannot sustainably serve a real consumer
                                                               → reconsider second public surface
single transport→application inbound path cannot realize a ratified T6 operation
                                                               → reconsider D07
zero direct owner imports cannot express a ratified semantic composition
                                                               → reconsider D06/D26/D27
T8-D proves provider-neutral explicit tx participation structurally impossible
                                                               → reconsider D25 shape, not atomicity requirement
T8-G proves composition package provides no independent construction value
                                                               → reconsider separate composition location
```

Not reopen triggers:

```text
preference
legacy path familiarity
sunk cost
ceremony aversion without failure evidence
hypothetical future feature
```

---

## 21. Platform-facing summary

```text
R10 T8-B — ADJUDICATED CORRECTED CANDIDATE

BACKEND GO BUILD UNIT
  one Go module

SEMANTIC PUBLIC SURFACES
  authentication
  organization
  authorization
  controlleddocs
  audit

OWNER INTERNALS
  private decomposition free inside owner boundary
  exactly one public surface per owner

APPLICATION
  stateless leaf orchestration
  no application→application imports
  no semantic authority

TRANSPORT
  HTTP/jobs inbound adapters
  transport → application ONLY

TRANSACTION
  one provider-neutral txscope mechanism class
  application owns lifecycle
  owner writes participate explicitly
  exact contract later

AUTHORIZATION
  final ALLOW/default-DENY remains Authorization-owned
  domain owners provide predicate meaning
  application only gathers/routes facts

AUDIT
  evidence remains Audit-owned
  owner owns event meaning for its mutation
  application coordinates same-tx append before commit
  exact handoff later

PLATFORM
  technical mechanisms only
  no semantic-owner imports/types/SQL authority

IDP
  protocol client in platform
  anti-corruption semantics in Authentication

PERSISTENCE
  owner-specific adapters private to owner
  shared PostgreSQL mechanism owns no business SQL

TOOLING
  tools/verify = verification SSOT
  tools/cilint = analyzer host
  tools/codegen = generator tooling

ENFORCEMENT
  Go internal visibility
  + closed-world architecture classification
  + import graph / SCC rules
  + bidirectional completeness
  + RED-proven negative fixtures
  + exception decay

NO LAUNCH PEER OWNER FOR
  Approval / Templates / Artifact / Search /
  Distribution / Periodic Review / Interchange / Records / generic Workflow

T8-C exact contracts NOT YET OPEN
T8-D+ untouched
IMPLEMENTATION BLOCKED
```

---

## 22. Round-2 Fable contract

Another broad review is **not** warranted.

A bounded delta review is sufficient and should attack only the material corrections introduced by the Lead adjudication:

```text
1. B-01 correction:
   is same-transaction Audit structurally guaranteed without freezing an AuditAppender contract now?

2. B-02:
   is platform/txscope + application-owned lifecycle the smallest correct T8-B decision,
   without stealing T8-D?

3. B-03:
   does transport → application ONLY create unnecessary pass-through complexity or remain the Global Maximum?

4. B-04 correction:
   does owner-authored predicate fact routing preserve Authorization as sole ALLOW/DENY authority
   without prematurely designing T8-C?

5. M-01/M-02:
   is platform sufficiently non-semantic, including IdP protocol vs anti-corruption separation?

6. M-03/M-04:
   does one public surface per owner + private decomposition + stateless application leaves
   avoid both semantic fragmentation and God packages?

7. M-05/M-06:
   are tools/cilint + tools/verify roles and closed-world/decay proof obligations sufficient?

8. M-07 Lead correction:
   are runtime cmd vs tools/codegen vs db/tests homes coherent without deciding T8-G/T10?

9. M-08 Lead correction:
   is `erasure-safe` the right T8-B invariant while exact purge/redact/PII-free realization remains T8-D?

10. L-03 Lead correction:
    does `officialrendition` correctly separate backend server rendering from interactive editor/viewer realization?
```

Round 2 must explicitly test for T8-C/T8-D trespass.

Required Round-2 outcome:

```text
APPROVE CORRECTED T8-B DELTA
or
APPROVE CORRECTED T8-B DELTA WITH MATERIAL FIXES
or
DO NOT APPROVE CORRECTED T8-B DELTA
```

Reviewer output remains evidence only.

---

## 23. Current gate

```text
T8-A   CLOSED / OPERATOR-RATIFIED

T8-B   ACTIVE
       Global Maximum class CONFIRMED
       Fable Round 1 COMPLETE
       Lead adjudication OPERATOR-APPROVED
       corrected candidate MATERIALIZED AS NON-AUTHORITATIVE STAGING
       bounded Fable Round 2 = NEXT

T8-C   NOT OPEN
T8-D+  NOT OPEN
implementation BLOCKED
```

Do not promote this artifact before the bounded Round-2 review converges and final operator ratification occurs.
