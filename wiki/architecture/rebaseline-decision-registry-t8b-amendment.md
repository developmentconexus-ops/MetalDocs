# Rebaseline Decision Registry — T8-B Closure Amendment

> **Status:** ACTIVE / OPERATOR-RATIFIED REGISTRY RECONCILIATION  
> **Ratified:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **T8-B authority:** `wiki/architecture/r10-t8b-backend-module-package-topology.md`

This bounded amendment reconciles the Decision Registry after T8-B closure. It changes backend physical-realization topology only. Product Contract and T1→T8-A semantics remain unchanged.

Registry authority chain is now:

```text
rebaseline-decision-registry.md
→ rebaseline-decision-registry-d4-amendment.md
→ rebaseline-decision-registry-t6-amendment.md
→ rebaseline-decision-registry-post-t6-amendment.md
→ rebaseline-decision-registry-t7-amendment.md
→ rebaseline-decision-registry-t8a-amendment.md
→ rebaseline-decision-registry-t8b-amendment.md
```

## 1. T8-B stage disposition

```text
T8-B Backend Module & Package Topology = CLOSED / OPERATOR-RATIFIED / PROMOTED
```

Ratified Global Maximum:

```text
ONE GO MODULE FOR BACKEND GO CODE
+ OWNER-FIRST MODULAR MONOLITH
+ ONE IMPORTABLE PUBLIC SURFACE PER SEMANTIC OWNER
+ STATELESS APPLICATION LEAF ORCHESTRATION
+ ONE SEMANTIC INBOUND DOOR THROUGH APPLICATION
+ NON-SEMANTIC PLATFORM MECHANISMS
+ WIRING-ONLY COMPOSITION ROOT
+ CLOSED-WORLD / DEFAULT-DENY FIRST-PARTY DEPENDENCY GRAPH
```

## 2. Semantic-owner realization

Exactly these Launch semantic homes receive public backend surfaces:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit — supporting evidence authority
```

Each exposes exactly one importable public package path. Owner-private package/file decomposition is ungated implementation-local structure. A second public surface is a material architecture decision.

No peer Launch semantic package is created for Approval, Templates, Artifact, Search, Distribution, Periodic Review, Taxonomy, Notifications, Interchange, Records or generic Workflow.

## 3. Coordination and dependency law

```text
transport → application is the only semantic inbound door
application = stateless choreography, never semantic authority
application leaf → application leaf = forbidden
owner → owner = forbidden
composition = construction/wiring only
platform = mechanism only
foreign SQL / hidden shared write authority = forbidden
```

Every first-party Go package is classified. Every first-party dependency edge is default-deny and must match an explicitly allowed T8-B direction. Unknown package or unknown edge fails verification.

Application leaves include the ratified Session/AuthN path plus Library, My Work, Document Official, Document Work, Governance Case, History, Audit and Administration orchestration.

The generated Go OpenAPI boundary is a transport/wire technical class; exact content remains T8-E.

## 4. Required seam classes

T8-B names these seam classes solely to make package ownership and dependency direction decidable:

```text
provider-neutral transaction participation
same-local-transaction owner evidence → Audit coordination
owner-authored domain predicate facts → Authorization final decision
```

Binding laws:

```text
application owns transaction-scope lifecycle
owners participate in caller-provided scope
owner↔Audit direct imports forbidden
owner owns its evidence meaning
Audit append completes before application commit
Authorization remains sole final ALLOW/default-DENY authority
missing/invalid/unverifiable required domain predicate fact = DENY
application routes facts but does not evaluate Authorization semantics
```

Exact types/methods/contracts remain T8-C. Persistence/locking/schema realization remains T8-D.

## 5. Mechanism placement

```text
platform/txscope             provider-neutral transaction participation
platform/postgres            DB mechanism only; no semantic SQL authority
platform/managedcontent      content mechanism only
platform/identityprovider    IdP protocol mechanics only
Authentication              IdP anti-corruption semantics
platform/officialrendition   server-side OfficialRendition only
platform/river               named durable-job mechanism
platform/idempotency         opaque replay mechanism; erasure-safe
```

A future editor backend mechanism, if proven necessary by later product/provider selection, attaches as a platform mechanism and not as a semantic owner.

Owner-specific persistence adapters are owner-private, but private package placement is not T8-D authority. T8-D owns persistence mapping/constraints/query/transaction semantics.

## 6. Verification disposition

```text
tools/verify registry/profile SSOT property       PRESERVE
tools/cilint generic analyzer harness             PRESERVE / REFINE
legacy architecture analyzer policy               REWRITE
legacy target-specific fixtures/baselines         REWRITE
tools/codegen                                      target tooling class for generators
```

Target proof law includes closed-world bidirectional package classification, default-deny edge enforcement, zero semantic-owner SCCs, transport-bypass rejection, private-owner protection and RED-proven negative fixtures.

Temporary architecture exceptions must have reason + removal trigger and must **FAIL**, not merely alert, when stale.

## 7. Stage-boundary reconciliation

```text
T8-B = CLOSED / PROMOTED
T8-C = NEXT / Internal Communication Contracts
T8-D = persistence realization only
T8-E = executable wire contract
T8-F = frontend realization
T8-G = runtime/process/deployment
T8-H = whole-T8 coherence
T10  = current→target moves/deletions/cutover
```

T8-C consumes the T8-B topology. It may not restore direct owner imports, transport bypass, foreign SQL, mechanism-as-authority or a second Authorization evaluator by convenience.

Implementation remains **BLOCKED**.