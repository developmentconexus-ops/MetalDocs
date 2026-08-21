# R10 T8-B — Backend Module & Package Topology

> **Status:** CLOSED / OPERATOR-RATIFIED / PROMOTED  
> **Ratified:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** DevelopmentConexus Engineering Method v1.0.0

This page is the durable target authority for R10 **T8-B — Backend Module & Package Topology**.

It freezes backend repository/package topology, semantic-owner realization boundaries, public/private Go surfaces, dependency direction, composition, shared-mechanism placement and the mechanical-enforcement law. It does **not** freeze the detailed internal communication contracts owned by T8-C, persistence owned by T8-D, wire contracts owned by T8-E, frontend realization owned by T8-F, runtime/process/deployment owned by T8-G or transition sequencing owned by T10.

Round-1 and bounded Round-2 Fable reviews were independent non-authoritative evidence. Their staging artifacts are provenance in Git history, not authority.

---

## 1. Ratified Global Maximum

```text
ONE GO MODULE FOR BACKEND GO CODE
+
OWNER-FIRST MODULAR MONOLITH
+
ONE IMPORTABLE PUBLIC SURFACE PER SEMANTIC OWNER
+
STATELESS APPLICATION LEAF ORCHESTRATION
+
ONE SEMANTIC INBOUND DOOR THROUGH APPLICATION
+
NON-SEMANTIC PLATFORM MECHANISMS
+
WIRING-ONLY COMPOSITION ROOT
+
CLOSED-WORLD / DEFAULT-DENY FIRST-PARTY DEPENDENCY GRAPH
+
SELECTIVE T8-A PROOF-BACKED MECHANISM REUSE
```

Rejected alternatives remain:

```text
owner packages with direct owner→owner imports
separate Go modules per semantic owner without a real consumer/release/trust boundary
use cases hosted inconsistently inside primary owners
separate policy semantic owner for Authorization/domain-predicate composition
legacy module topology renamed or mapped one-for-one
```

No materially superior topology class survived the Method challenge.

---

## 2. Target backend repository projection

Repository roots not shown are not implicitly deleted or redesigned by T8-B.

```text
MetalDocs/
├── go.mod                              # one Go module for backend Go code
│
├── api/
│   └── openapi/
│       └── v1/
│           └── openapi.yaml            # contract SSOT; exact wire content = T8-E
│                                       # generated Go wire boundary = transport/wire class
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
│   ├── application/                    # stateless leaf orchestration
│   │   ├── session/
│   │   ├── library/
│   │   ├── mywork/
│   │   ├── documentofficial/
│   │   ├── documentwork/
│   │   ├── governancecase/
│   │   ├── history/
│   │   ├── audit/
│   │   ├── admin/
│   │   └── maintenance/                # T5-J managed-content GC choreography
│   │
│   ├── transport/
│   │   ├── http/                       # inbound HTTP/browser adapters
│   │   └── jobs/                       # durable-job inbound adapters
│   │
│   ├── platform/                       # non-semantic mechanisms only
│   │   ├── txscope/                    # provider-neutral transaction participation
│   │   ├── postgres/                   # pool/driver/bootstrap + tx mechanics
│   │   ├── managedcontent/             # content + malware-inspection mechanisms
│   │   ├── identityprovider/           # IdP protocol client only
│   │   ├── officialrendition/          # server-side OfficialRendition mechanism only
│   │   ├── river/                      # durable-job mechanism
│   │   ├── idempotency/                # opaque replay mechanism
│   │   ├── observability/
│   │   └── config/
│   │
│   └── composition/                    # construction/wiring only
│
├── tests/                              # dedicated cross-boundary test roots
│
└── tools/
    ├── verify/                         # verification registry / profile SSOT
    ├── cilint/                         # architecture/static-analysis host
    └── codegen/                        # contract/code-generation tooling
```

`cmd/*` means runtime entrypoints only. Code generators are tooling, not runtime shells.

`application/maintenance` is a non-semantic application leaf whose current consumer is the ratified T5-J managed-content GC choreography. It remains inside the existing `application` class and adds no semantic owner, public owner surface, dependency class or Product operation.

Exact private package/file decomposition inside a semantic owner is **not architecture authority** and is not owned by T8-D. T8-D owns persistence mapping and correctness realization, not owner-private folder structure.

---

## 3. Semantic-owner realization law

Exactly five Launch semantic homes exist:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit — supporting evidence authority
```

### 3.1 One public surface per owner

Each semantic owner exposes exactly **one importable public package path**.

An owner may freely decompose its private implementation using files or nested packages such as:

```text
internal/<owner>/internal/<responsibility>
```

No architecture approval/evidence gate applies to purely owner-private decomposition. A **second public package path** for one semantic owner is material architecture and requires a named consumer plus proof that the single public surface cannot sustainably serve it.

Entity/family names do not justify public package boundaries.

### 3.2 No false peer semantic packages

No Launch peer semantic owner/package is created for:

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

Governance/release/template semantics remain under Controlled Documents where already ratified. Search is discovery/read behavior, not authority. Launch+/Future capabilities remain deferred until Product authority creates a real consumer.

```text
prepare the seam, not dormant implementation
```

---

## 4. Application orchestration law

`internal/application/*` is non-semantic choreography.

Human/API application leaves are derived from ratified T6 use-case/lens meaning and must provide a legal application route for every ratified inbound use-case family. Accepted non-semantic maintenance choreography may occupy the same `application` class when a named upstream consumer requires it; the current instance is T5-J managed-content GC. An omitted inbound leaf never creates a `transport → owner` exception.

Application leaves are stateless and may:

```text
coordinate one application/business operation
open/commit/rollback the shared local transaction scope
invoke semantic-owner public surfaces
gather owner-authored predicate facts/results for Authorization
coordinate required same-transaction Audit append
coordinate justified idempotency/durable intent
return purpose-built use-case/read-model results
```

They must not:

```text
own semantic lifecycle/current state
reimplement semantic owner rules
compute the canonical Authorization ALLOW/default-DENY equation
invent owner-intrinsic Audit meaning
contain semantic SQL
hold persistent/process-global state
become a generic workflow/event-bus/domain platform
import another application leaf
```

```text
application/<a> → application/<b>    FORBIDDEN
owner           → application         FORBIDDEN
```

Frontend folder topology is not inferred one-for-one from these Go leaves.

---

## 5. Transport law — one semantic inbound door

Inbound adapters translate protocols and invocation only.

```text
transport/http → generated wire contract boundary
transport/http → application
transport/jobs → application
```

The generated Go OpenAPI boundary is a **transport/wire technical class**, never a semantic owner. Exact generated content remains T8-E.

Forbidden:

```text
transport → semantic owner public/private implementation
transport → persistence/SQL
transport → transaction-scope creation
transport → business rules
```

There is exactly one **semantic** inbound composition door:

```text
transport → application
```

Durable job handlers are inbound adapters; they reload canonical state and do not become semantic/job-state authorities.

---

## 6. Composition law

`internal/composition` owns construction and wiring only.

Allowed:

```text
construct semantic owners
construct application leaves
construct transport adapters
construct platform mechanisms/adapters
bind technical seams/interfaces to implementations
assemble runtime dependency graphs for cmd shells/tests
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

## 7. Platform law — mechanism never authority

`internal/platform/*` contains replaceable non-semantic mechanisms.

Forbidden:

```text
platform → semantic-owner package imports
platform state/signatures acquiring semantic-owner meaning
platform → semantic SQL authority
platform → cross-owner business query
platform → shared semantic write authority
platform → product workflow/state-machine authority
```

Semantic identifiers may cross a mechanism seam only as opaque values required by a consumer-owned technical contract; the mechanism does not acquire their business meaning.

### 7.1 PostgreSQL

`platform/postgres` owns pool/driver/bootstrap and transaction mechanics only. It does not own Organization, Authorization, Controlled Documents or Audit SQL.

Owner-specific persistence adapters are owner-private. Their **package placement is ungated private implementation**. T8-D owns their schema/query/constraint/transaction mapping and persistence semantics, not private folder placement.

### 7.2 Managed content

Managed content remains mechanism only:

```text
no Document/Revision/Submission semantic ownership
no generic owner_type/owner_id semantic registry
exact-content descriptors remain semantic-record facts
malware inspection remains technical admission
```

### 7.3 Identity provider

`platform/identityprovider` owns protocol mechanics only: discovery, token exchange, JWKS/protocol verification and raw provider protocol handling.

Authentication owns provider anti-corruption meaning, ProviderSubjectBinding, ApplicationSession and assurance semantics.

Provider claims/roles/groups/organizations/permissions do not escape the protocol + Authentication boundary as MetalDocs Organization/Authorization truth.

### 7.4 Official rendition

`platform/officialrendition` is the server-side OfficialRendition mechanism only. It is distinct from interactive editor/viewer realization.

If a later selected editor proves a backend technical mechanism is required, that mechanism attaches additively under the existing `platform` mechanism class. No editor mechanism becomes a semantic owner merely because a provider needs it.

### 7.5 River

River remains a durable-job mechanism only for named T5 consumers. It does not create generic event-bus/workflow authority.

### 7.6 Idempotency

Idempotency is opaque replay mechanism state.

```text
platform/idempotency never authorizes replay disclosure
application performs live Authorization/resource-visibility recheck
replay persistence must be erasure-safe and not become a PII retention root
```

Exact replay contract/schema/retention and PII-free-vs-purge realization remain T8-C/T8-D.

---

## 8. Required seam classes — exact contracts deferred

T8-B names three seam classes because the dependency graph is otherwise incomplete. Their method/type signatures remain T8-C.

### 8.1 Transaction participation

```text
home                         = internal/platform/txscope
semantic status              = mechanism only
scope lifecycle owner        = application
semantic owner participation = explicit
provider identity leakage    = forbidden
ambient/hidden tx ownership  = forbidden
```

Application opens and commits/rolls back the shared local scope. Owner writes participate in the caller-provided scope and do not create an independent scope for one composed transition.

Exact Go type/method names are T8-C; isolation/locks/serialization/PostgreSQL mapping are T8-D.

### 8.2 Same-transaction Audit evidence

```text
owner → Audit direct import            FORBIDDEN
Audit → other owner direct import      FORBIDDEN
owner-intrinsic evidence meaning       produced/owned by mutating owner
cross-owner routing/composition        application
required Audit append                  inside same local transaction
application commit                     only after required append succeeds
Audit                                  evidence, never lifecycle authority
```

Exact evidence-handoff interface/fact/result types remain T8-C. Persistence realization remains T8-D.

### 8.3 Authorization/domain-predicate decision

```text
Authorization  owns grants/scope + final ALLOW/default-DENY
business owner owns relationship/state/governance predicate meaning
application    gathers/routes facts only
```

Direction:

```text
application obtains owner-authored predicate facts/results
→ supplies required context to Authorization
→ Authorization performs canonical access decision
→ ALLOW / DENY
```

If a required predicate fact is absent, invalid or unverifiable, Authorization **DENIES**.

No second evaluator of canonical Authorization semantics is allowed. `allowed_actions` uses the same decision authority.

Exact decision interface/fact vocabulary/actor-scope shapes remain T8-C.

---

## 9. Closed-world dependency law

Architecture dependency enforcement is **default deny**, not a finite denylist.

Every first-party Go package maps to exactly one target class. The target classes include:

```text
semantic-owner-public
semantic-owner-private
application
transport
wire-contract
platform
composition
runtime-entrypoint
tooling
dedicated-test
```

In-package `_test.go` files inherit the production package's architecture class and edge law. Dedicated cross-boundary tests under `tests/` use the dedicated-test class. Test-only imports never create production dependency entitlement.

### 9.1 Normative allowed first-party edges

The following list is exhaustive at the class/direction level. Any first-party edge not affirmatively allowed is forbidden.

```text
runtime-entrypoint → composition

transport/http     → wire-contract
transport          → application
transport          → platform/observability            # technical instrumentation only

application        → semantic-owner-public
application        → platform/txscope
application        → platform/idempotency               # only for idempotent use cases
application        → platform/observability             # technical instrumentation only

semantic-owner-public/private
                   → same-owner subtree
semantic-owner-public/private
                   → platform/txscope                    # only when tx participation requires it

wire-contract      → no product-runtime first-party package

platform           → platform                           # technical DAG only; no semantic authority

composition        → semantic-owner-public
composition        → application
composition        → transport
composition        → platform

tooling            → tooling
codegen tooling    → wire-contract                      # generation/validation only

dedicated-test     → public/technical target surfaces needed by the test
```

All other technical mechanism access by owners is through consumer-owned technical seams whose exact contracts are T8-C; this prevents direct mechanism coupling from becoming the accidental default.

First-party import cycles remain forbidden by Go and semantic-owner SCCs must be zero.

### 9.2 Important forbidden classes

The default-deny rule subsumes, among others:

```text
owner → another owner
owner → application/composition
transport → owner
application → application
application → owner-private
application → platform/postgres
platform → owner
foreign SQL as owner communication
common/shared/utils/models/repositories/services/domain as cross-owner semantic dumping grounds
provider claims as Organization/AuthZ truth
second Authorization evaluator
```

---

## 10. Mechanical enforcement

Target architecture laws must be falsifiable.

```text
tools/verify  = verification registry / profile SSOT
tools/cilint  = architecture/static-analysis host
tools/codegen = contract/code-generation tooling
```

T8-A selective-reuse result:

```text
tools/verify registry/profile property       PRESERVE
tools/cilint generic analyzer harness         PRESERVE / REFINE
legacy target-specific analyzer policy        REWRITE
legacy target-specific fixtures/baselines     REWRITE
```

Required target proofs:

```text
closed-world package classification
class catalog ↔ live tree in BOTH directions
default-deny allowed-edge enforcement
semantic-owner SCC = zero
foreign owner-private reachability rejected
transport bypass rejected
prohibited peer semantic roots rejected
platform→owner imports rejected
foreign/cross-owner SQL ownership rejected
negative fixture for each mechanical rejection class
```

Any temporary architecture exception must be explicit, have a reason and removal trigger, and **FAIL** verification when the underlying violation no longer exists. An alert is not sufficient.

Same-transaction Audit and single-source Authorization proofs become mechanically exact once T8-C/T8-D provide their contracts/realization; T9 must prove their composed behavior.

---

## 11. Legacy/current disposition

```text
legacy internal/modules/* semantic topology     REWRITE / REHOME
legacy peer Approval/Templates/etc. modules     DELETE / REHOME by Launch ownership
legacy composition placement                     concept survives; contents not inherited
current platform packages                        selective reuse only through T8-A gate
current apps/*/cmd runtime shape                 current evidence; target topology T8-G
current root cmd generators                      REHOME to tools/codegen in T10
current db assets                                 root class survives; exact content T8-D/T10
current analyzer harness                         selective harness reuse
legacy analyzer policy/allowlists                REWRITE / RETIRE
OpenAPI contract-first property                  PRESERVE; exact contract T8-E
one Go module                                     independently reselected
```

T8-B authorizes no code moves/deletions. T10 owns current→target transition sequencing.

---

## 12. Reopen triggers

Reopen only on material evidence such as:

```text
independent external Go consumer / release train / trust boundary
single public owner surface cannot sustainably serve a real named consumer
a ratified T6 use case cannot pass through an application leaf without material harm
closed-world dependency matrix requires unstable exception proliferation
T8-C proves an accepted seam cannot express required semantics without owner leakage
T8-D proves provider-neutral transaction participation cannot realize accepted atomicity
future Product authority introduces a new semantic owner/capability
```

Preference, legacy familiarity, sunk cost, migration convenience and hypothetical future capability are not reopen triggers.

---

## 13. T8-B closure and next stage

```text
T8-B Backend Module & Package Topology = CLOSED / OPERATOR-RATIFIED / PROMOTED
```

The next open stage is:

```text
T8-C — Internal Communication Contracts
```

T8-C must freeze exact internal owner/application/mechanism contracts while preserving this topology. It must not reopen direct owner imports, mechanism-as-authority, foreign SQL or hidden shared write authority by convenience.

Implementation remains **BLOCKED**.