# R10-T8C — Internal Communication Contracts — Bootstrap

```text
ACTIVE STAGING ROUTER
NON-AUTHORITATIVE
ROUND-1 REVIEW + LEAD ADJUDICATION COMPLETE
ADJUDICATED CORRECTED CANDIDATE MATERIALIZED
BOUNDED ROUND-2 FABLE NEXT
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Stage:** T8-C ACTIVE  
> **Upstream authority:** `wiki/architecture/r10-t8b-backend-module-package-topology.md`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This file is a non-authoritative T8-C staging router.

Provenance:

```text
original candidate
  docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-global-maximum-candidate.md

Round-1 independent review
  docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-independent-fable-review.md

current adjudicated corrected candidate / Round-2 input
  docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-adjudicated-corrected-candidate.md
```

None is durable target authority.

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

## 4. Round-1 disposition

Independent Round-1 verdict:

```text
APPROVE T8-C GLOBAL MAXIMUM CANDIDATE WITH MATERIAL FIXES
BLOCKER 5 / MAJOR 6 / LOW 5
GLOBAL MAXIMUM CLASS CONFIRMED
T8-B REOPEN NO
T1→T7 REOPEN NO
```

Lead adjudication:

```text
B1 ACCEPT
B2 ACCEPT
B3 ACCEPT
B4 ACCEPT
B5 REJECT AS BLOCKER

M1 ACCEPT
M2 ACCEPT
M3 ACCEPT WITH NARROWING
M4 ACCEPT
M5 ACCEPT AS REAL T8-C DECISION -> PII-FREE REPLAY BY CONSTRUCTION
M6 ACCEPT

L1-L5 ACCEPT
```

No accepted correction changes the authority-aligned hybrid model class.

---

## 5. Current corrected candidate

Current Round-2 input:

`docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-adjudicated-corrected-candidate.md`

Key corrected delta:

```text
explicit database/sql-family transaction substrate
+ sealed txscope + platform-only native SQL binding for River
+ Audit historical-visibility read contract
+ Authorization AuthorizedScopes query
+ protected User-eligibility serialization contract
+ owner VersionToken / expected-version law
+ verified primitive issuer+subject provider seam
+ ManagedContent admission-claim lifecycle
+ DeleteReclaimable + T5-J GC contract family
+ malware verdict digest correlation
+ concurrent idempotency outcome law without mandatory savepoint/retry
+ PII-free ReplaySnapshot by construction
+ OfficialRendition content-read contract
+ bounded operation-census precision fixes
```

T8-D and T8-E boundaries remain explicit in the corrected candidate.

---

## 6. Bounded Round-2 scope

Round 2 is a delta review only. It must not re-derive already-confirmed unchanged T8-C decisions without a new contradiction.

Attack only:

```text
txscope/River corrected binding + deliberate database/sql constraint
Audit read visibility
Authorization AuthorizedScopes
ManagedContent claims + GC completeness/legal path
protected eligibility serialization semantics
owner VersionToken vs T8-E ETag boundary
provider primitive identity seam
malware digest correlation
B5 disagreement using PostgreSQL primary behavior
PII-free replay decision
OfficialRendition exact-content read
operation-census delta closure
```

Another broad review is justified only if this delta changes the confirmed Global Maximum class or proves a real T8-B/upstream contradiction.

---

## 7. Stage boundaries

T8-C does **not** own:

```text
schema/table/index/constraint design                 T8-D
exact lock/SQL/upsert/savepoint realization          T8-D
exact OpenAPI paths/schemas/headers/status mapping   T8-E
frontend package/state/query topology                T8-F
runtime process/deployment count                     T8-G
current→target moves/deletions                       T10
implementation task decomposition                    T11
```

---

## 8. Exact next action

```text
bounded Fable Round 2 on the adjudicated corrected candidate
→ verify only the material Round-1 correction delta
→ attack B5 Lead rejection with PostgreSQL primary evidence
→ attack new PII-free replay decision
→ verify B1/B2/B3/B4/M1/M2/M3/M4/M6 closure
→ verify operation-census delta closure
→ report any surviving material contradiction
→ final Lead adjudication
→ explicit operator ratification before durable T8-C promotion
```

T8-D remains **NOT OPEN**. Implementation remains **BLOCKED**.