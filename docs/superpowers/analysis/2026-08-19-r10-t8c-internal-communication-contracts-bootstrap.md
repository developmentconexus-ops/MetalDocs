# R10-T8C — Internal Communication Contracts — Bootstrap

```text
ACTIVE STAGING ROUTER
NON-AUTHORITATIVE
GLOBAL MAXIMUM CANDIDATE MATERIALIZED
INDEPENDENT REVIEW NEXT
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Stage:** T8-C ACTIVE  
> **Upstream authority:** `wiki/architecture/r10-t8b-backend-module-package-topology.md`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This file is a non-authoritative T8-C staging router. The current candidate is:

`docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-global-maximum-candidate.md`

The candidate is evidence/input for independent review, not target authority.

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

For external library/framework/SDK/API/cloud behavior that is load-bearing, follow `AGENTS.md`: use current primary/official documentation and Context7 where required. External practice is falsification/reference evidence and never overrides MetalDocs semantic authority.

---

## 3. Frozen T8-B constraints

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

T8-C may propose a T8-B reopen only if an exact required contract proves a concrete contradiction.

---

## 4. Required T8-C contract families

```text
owner public capabilities and queries
provider-neutral transaction participation
same-transaction owner evidence → Audit handoff
Authorization decision + owner predicate facts
transaction-coupled durable intent
managed-content / malware / IdP / rendition mechanism ports
idempotency replay mechanism + application replay result
read-projection composition
inside-vs-outside transaction classification
fail-closed/error semantics
```

No Writer may later need to invent a material internal contract that T8-C should have frozen.

---

## 5. Candidate status

The current candidate has materially completed at staging level:

```text
interaction census
contract ownership/direction
exact txscope candidate
Audit handoff candidate
Authorization/domain-predicate candidate
GROUP mid-transition resolver candidate
Authentication / Organization / Authorization / ControlledDocs capability census
managed-content + malware ports
OfficialRendition renderer + named durable-intent port
idempotency BeginIn/CompleteIn + operation-local ReplaySnapshot
read-projection composition
inside/outside transaction law
failure/fail-closed law
T8-A selective-reuse disposition
primary/current reference pass
credible alternatives
Structural Inversion + subtractive pass
T8C-D01→D25 candidate decision set
```

Selected candidate class:

```text
AUTHORITY-ALIGNED HYBRID CONTRACT MODEL
```

It deliberately rejects:

```text
producer interfaces everywhere
consumer interfaces for every owner call
shared/common contract package
raw owner-domain types crossing owners
generic EventBus/outbox
HTTP response contract inside application
generic UnitOfWork/service locator/policy language
```

---

## 6. Review-sensitive candidate seams

Independent review must attack, not assume, at least:

```text
txscope database/sql executor shape vs a more opaque transaction token
application prohibition on Scope SQL invocation
owner-local Audit evidence → application mapping → Audit append drop/mutation risk
Authorization Check/DomainPredicate vocabulary and sole-decision authority
GROUP resolver + truthful empty snapshot/recovery
Authentication ProviderClient raw/protocol shape
Group deletion / owner eligibility / RoleAssignment target fact routing
ManagedContent/Malware exact-content boundary
OfficialRendition renderer + River-backed intent seam
idempotency BeginIn/CompleteIn + ReplaySnapshot across retries/deploy evolution
read projections / batch decisions / N+1 avoidance without foreign SQL
complete legal path for every T6 operation family
T8-D persistence and T8-E wire trespass
any removable abstraction
any materially superior contract-placement model
```

Reviewer findings are evidence only. The Lead adjudicates; the operator ratifies.

---

## 7. Stage boundaries

T8-C does **not** own:

```text
schema/table/index/constraint design                 T8-D
isolation/lock/SQL realization                       T8-D
exact OpenAPI paths/schemas/headers/status mapping   T8-E
frontend package/state/query topology                T8-F
runtime process/deployment count                     T8-G
current→target moves/deletions                       T10
implementation task decomposition                    T11
```

T8-C may name persistence needs only to explain a contract requirement; it must not design persistence by stealth.

---

## 8. Exact next action

```text
independent Fable review of the current T8-C Global Maximum candidate
→ reconstruct repository authority independently
→ apply DevelopmentConexus Engineering Method v1.0.0
→ adversarially search for a better Global Maximum
→ disposition T8C-D01→D25
→ identify BLOCKER / MAJOR / LOW findings
→ identify any T8-B reopen or T8-D/T8-E trespass
→ Lead technical adjudication
→ bounded correction/review only if a material delta survives
→ explicit operator ratification before durable promotion
```

T8-D remains **NOT OPEN**. Implementation remains **BLOCKED**.