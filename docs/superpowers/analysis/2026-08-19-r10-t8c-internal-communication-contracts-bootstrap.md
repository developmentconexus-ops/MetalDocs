# R10-T8C — Internal Communication Contracts — Bootstrap

```text
ACTIVE STAGING
NON-AUTHORITATIVE
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Stage:** T8-C ACTIVE  
> **Upstream authority:** `wiki/architecture/r10-t8b-backend-module-package-topology.md`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This is the active non-authoritative bootstrap for R10 **T8-C — Internal Communication Contracts**. It routes the next architecture work; it is not target authority.

---

## 1. Exact T8-C question

> **What is the smallest complete set of internal contracts that lets the ratified owners and non-semantic application/mechanism layers realize T1→T8-B semantics without direct owner imports, foreign SQL, duplicate authority, hidden write ownership or unnecessary interface ceremony?**

T8-C consumes T8-B. It does not reopen package topology by preference.

---

## 2. Binding inputs

Read current authority in repository order, including:

```text
Product Contract REV001
Whole-Product GCR + 4+1 ownership
T1 semantic state/invariants
T2 governance/effectivity/transaction laws
T3 Authorization/Audit + D4
T4 exact content/storage integrity
T5 durable async/search/external effects
T6 canonical API/frontend journeys + amendment
T7 migration truth
T8-A technical authority/legacy disposition
T8-B backend topology
Decision Registry + amendments through T8-B
post-T6 implementation-readiness program
```

Current code/interfaces are evidence only.

---

## 3. Frozen T8-B constraints T8-C must preserve

```text
one Go module for backend Go code
one public surface per semantic owner
owner-private decomposition stays private/ungated
transport → application is the only semantic inbound door
application leaves are stateless choreography
application leaf → application leaf forbidden
owner → owner forbidden
platform = mechanism only
composition = wiring only
first-party dependency graph = closed-world/default-deny
foreign SQL and hidden shared write authority forbidden
```

Semantic homes remain:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit
```

T8-C may propose a T8-B reopen only if an exact required contract proves a concrete contradiction with this topology.

---

## 4. Contract families T8-C must freeze

### C1 — Owner public capabilities and queries

For each ratified application/use-case family, identify the smallest owner-facing capability/query set needed by application choreography.

Do not expose internal aggregates/repositories merely because they exist.

### C2 — Transaction participation contract

T8-B requires:

```text
home class                    platform/txscope
application                   opens + commits/rolls back
owner writes                  participate explicitly
provider-specific tx types    forbidden on owner public surfaces
```

T8-C must freeze the exact provider-neutral contract/type/method ownership needed for participation without deciding T8-D isolation/locking/PostgreSQL mapping.

### C3 — Same-transaction Audit evidence handoff

T8-B requires:

```text
owner owns intrinsic evidence meaning
application coordinates Audit append
append succeeds before commit
owner↔Audit direct imports forbidden
```

T8-C must freeze the smallest handoff contract and its producer/consumer ownership. It must support one or multiple required semantic events in one transaction without turning Audit into lifecycle authority or application into event-meaning authority.

### C4 — Authorization/domain-predicate decision contract

T8-B requires:

```text
business owner authors relationship/state/governance predicate facts
application gathers/routes facts
Authorization alone performs final ALLOW/default-DENY
missing/invalid/unverifiable required fact = DENY
no second evaluator
```

T8-C must freeze:

```text
exact decision capability
actor/scope/context inputs
owner predicate fact vocabulary/ownership
transaction participation where current truth must be checked in-scope
result/decision shape
allowed_actions consumption of the same decision authority
```

Do not move domain predicate meaning into Authorization and do not move ALLOW/DENY composition into application.

### C5 — Transaction-coupled durable intent

Where T5/T6 already require a future external/durable effect to share the authoritative local commit, freeze the smallest enqueue/intent contract and ownership.

River remains mechanism; jobs/events are not semantic authority.

### C6 — Managed-content / provider / rendition mechanism ports

Freeze only the consumer/producer contracts that current Launch semantics actually require:

```text
ManagedContentStore / admission-related technical seams
malware inspection
IdP protocol client consumption by Authentication anti-corruption boundary
OfficialRendition mechanism where required
idempotency replay mechanism
observability/config seams where materially contract-bearing
```

Do not create a generic infrastructure service locator or a universal mechanism interface framework.

### C7 — Read projections

T6 purpose-built read models remain lenses, not owners. Freeze how application obtains bounded owner facts and composes read results without foreign SQL or persistent duplicate truth.

Library/Search remains application read orchestration over canonical owner truth, not a Search owner.

---

## 5. Interaction census T8-C must walk

At minimum cover the contract implications of:

```text
browser login / callback / session resolution / logout
User creation + UserProfile + ProviderSubjectBinding + Audit
User offboarding across Organization/AuthN/AuthZ/Audit
Document + REV000 + WorkingContent creation
working-content read/write with OCC semantics
submit/freeze immutable Submission
governance ACCEPT / RETURN / withdraw / cancel
Release/effectivity and replacement
obsolescence
responsible-owner eligibility/change
Library / My Work / Document Official / Document Work / Governance Case / History
allowed_actions
exact-content admission/retrieval
OfficialRendition durable work where activated
transaction-coupled provider/durable intents
idempotent command replay disclosure
```

For every interaction classify:

```text
synchronous query
synchronous capability/mutation
same-transaction participant
durable-intent producer
durable-job invocation/readback
technical mechanism call
read projection composition
```

---

## 6. Required decision questions

T8-C must explicitly answer:

1. Which side owns each contract: consumer, producer or application boundary?
2. Which contracts are public owner surfaces vs owner-private implementation details?
3. Which data is semantic fact vs bounded DTO/fact projection?
4. Which calls must participate in one shared transaction?
5. Which calls must not occur inside the transaction because they are external/provider effects?
6. How does Audit evidence cross the boundary without direct owner coupling?
7. How are Controlled Documents predicate facts supplied without duplicating Authorization?
8. How does a contract fail closed when required current truth is absent/stale/unverifiable?
9. Which T5 durable seams are genuinely required vs generic event-bus overengineering?
10. Which current interfaces pass the T8-A five-part reuse gate, if any?
11. Can any proposed contract be deleted without weakening a ratified invariant?
12. Does any contract force a T8-B reopen? If yes, prove the contradiction rather than preferring another topology.

---

## 7. Credible alternatives to compare

For material contract families compare at least the relevant forms, for example:

```text
producer-owned interface/DTO
consumer-owned capability/port
application-owned bounded orchestration DTO
raw owner domain types crossing boundaries
shared generic common contract package
```

Do not choose by pattern familiarity. Apply Global Maximum, authority duplication, interface count, testability, future retrofit and enforcement cost.

Generic shared contracts/common models are suspect because they can become hidden authority.

---

## 8. Required T8-C outputs

A promotable T8-C candidate must include:

```text
complete internal interaction matrix
exact owner public query/capability contracts
exact transaction-participation contract
exact Audit evidence handoff contract
exact Authorization decision/domain-fact contract
exact transaction-coupled durable-intent contract
material mechanism-port contracts
contract ownership for every interface/DTO
allowed call direction for each contract
transaction-inside vs transaction-outside classification
error/fail-closed semantics where material
reuse disposition against T8-A gate
proof/enforcement strategy
2–3 materially credible alternatives for disputed contract families
subtractive/YAGNI pass
adversarial challenge
platform-facing summary
operator ratification
```

No Writer may later need to invent a material internal contract that T8-C should have frozen.

---

## 9. Stage boundaries

T8-C does **not** own:

```text
schema/table/index/constraint design                 T8-D
isolation/lock realization                           T8-D
exact OpenAPI paths/schemas/headers                  T8-E
frontend package/state/query topology                T8-F
runtime process/deployment count                     T8-G
current→target moves/deletions                       T10
implementation task decomposition                    T11
```

T8-C may name persistence needs only to explain a contract requirement; it must not design the schema by stealth.

---

## 10. Method challenge

Apply explicitly:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ Credible Alternatives
→ Local vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / Future Cost
→ Authority / Boundary
→ Enforcement
→ Proof Strategy
→ Adversarial Challenge
→ Decision
→ Reopen Triggers
```

Mandatory passes:

```text
Structural Inversion
Subtractive
Authority Duplication
Failure Class
```

Primary failure classes include:

```text
direct owner coupling through interface types
shared DTO/common-model authority
cross-owner raw persistence/SQL
transaction ownership ambiguity
Audit evidence drop channel
second Authorization evaluator
provider/external call inside semantic transaction
generic event bus / interface explosion
read projection becoming truth
mechanism contract leaking provider identity
```

---

## 11. Exact next action

```text
derive the complete interaction census from T2/T3/T4/T5/T6 + T8-B
→ classify every interaction by sync/query/mutation/tx/durable/mechanism/read-projection
→ freeze contract ownership and direction
→ resolve txscope / Audit evidence / Authorization predicate contracts first
→ derive remaining owner/mechanism contracts
→ compare credible contract-placement approaches
→ apply Method + subtractive pass
→ adversarial challenge
→ operator-ratifiable T8-C candidate
```

Implementation remains **BLOCKED**.