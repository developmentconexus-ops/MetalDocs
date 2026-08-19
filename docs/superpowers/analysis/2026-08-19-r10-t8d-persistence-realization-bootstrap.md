# R10-T8D — Persistence Realization — Bootstrap

```text
ACTIVE STAGING ROUTER
NON-AUTHORITATIVE
GLOBAL MAXIMUM CANDIDATE MATERIALIZED
INDEPENDENT FABLE REVIEW = NEXT
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Stage:** T8-D ACTIVE  
> **Upstream authority:** `wiki/architecture/r10-t8c-internal-communication-contracts.md`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This file is the active non-authoritative router for R10 **T8-D — Persistence Realization**. It does not contain target authority.

Current candidate:

`docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md`

No T8-D decision is durable until independent challenge, Lead adjudication and explicit operator ratification promote it into `wiki/`.

---

## 1. Exact T8-D question

> **What is the smallest PostgreSQL persistence realization that makes T1→T8-C invariants and internal contracts structurally enforceable, assigns every persistent fact to its ratified semantic/mechanism owner, and maps required ACID/OCC/serialization behavior to explicit schema/constraint/query/lock rules without foreign SQL, duplicate truth, hidden shared write authority, wire leakage or speculative persistence?**

---

## 2. Binding inputs

Read in repository authority order:

```text
Product Contract REV001
Whole-Product GCR + 4+1 ownership
T1 semantic state/invariants
T2 governance/effectivity/transaction law
T3 Authorization/Audit + D4
T4 exact content/storage/restore
T5 durable async/search/external effects
T6 canonical API/frontend journeys
T7 migration truth
T8-A technical authority/legacy disposition
T8-B backend topology
T8-C internal communication contracts
Decision Registry + amendments through T8-C
post-T6 implementation-readiness program
TRRB
this bootstrap
current T8-D candidate
current schema/migrations/SQL only for concrete evidence/reuse claims
```

Current implementation is evidence only. Table names, RLS/GUC, legacy module ownership and current migrations receive no survival entitlement.

---

## 3. Frozen upstream laws

```text
semantic homes = Authentication / Organization / Authorization /
                 Controlled Documents / Audit

one local ACID product-state transaction per native business transition
PostgreSQL READ COMMITTED
+ narrow explicit serialization
+ OCC/CAS
+ structural constraints where required

database/sql transaction family
owner-private semantic SQL only
no owner→owner imports
no foreign SQL as communication
same-Scope required Audit
same-Scope required River intent
same-Scope Idempotency claim/completion
same-Scope AdmissionClaim consumption
protected eligibility serialization with offboarding
self-contained PII-free ReplaySnapshot
ManagedContent mechanism != semantic authority
Search materialization OFF
```

---

## 4. Current candidate class

The operator-approved design materialized for independent review selects:

```text
OWNER-NAMESPACED POSTGRESQL RELATIONAL CORE
+
DECLARATIVE CORRECTNESS
+
PRIVILEGE-ENFORCED IMMUTABLE HISTORY
+
READ COMMITTED NARROW SERIALIZATION
+
EXPLICIT CAS
+
IDENTITY-ONLY CROSS-OWNER REFERENTIAL INTEGRITY
+
TRANSACTIONAL KEY↔REPLAY COMPLETION
+
THIRD-PARTY RIVER SCHEMA ISOLATION
+
PROOF-BACKED SELECTIVE LEGACY PROPERTY REUSE
-
LEGACY PHYSICAL SHAPE INHERITANCE
-
GENERIC PERSISTENCE FRAMEWORKS
-
DUPLICATE CURRENT TRUTH
```

Principal candidate decisions include:

```text
one PostgreSQL database
schemas: authn / org / authz / controlled_docs / audit / platform / river
fully-qualified first-party SQL
PostgreSQL-16-compatible primitive floor
complete bidirectional DB-object ownership catalog
Role/Permission static; RoleAssignment persisted
no Launch RLS/tenant substrate
explicit BIGINT VersionToken and WorkingContent generation
Revision.state canonical lifecycle + immutable Release fact
partial uniqueness for one open and one EFFECTIVE Revision
closed relational governance model
live GROUP dependency separate from activated candidate snapshot
semantic exact descriptors + technical ManagedContent
row-existence AdmissionClaim + repeated two-phase GC proof
paired Idempotency Key + Replay with deferred completion FK
River under third-party river.*
identity/existence-only cross-owner FKs; no semantic cascades
runtime DB role separate from DDL owner
append-only DB grants for immutable history
zero semantic lifecycle triggers baseline
protected actor FOR SHARE; Document FOR UPDATE lifecycle root
explicit owner-private database/sql SQL; no generic ORM/repository framework
```

---

## 5. Stage boundaries

T8-D owns:

```text
schemas/tables/material columns and types
correctness PK/FK/unique/check/partial constraints
immutable/history relational shapes
owner-private SQL/query boundaries
OCC/VersionToken physical realization
transaction/serialization/lock mapping
managed-content/idempotency/River persistence boundaries
```

T8-D does not own:

```text
semantic/product changes                         T1→T7
package/dependency topology                     T8-B
internal contract signatures                    T8-C
exact HTTP/OpenAPI/ETag encoding                T8-E
frontend realization                            T8-F
runtime/process/deploy                          T8-G
Golden Flow matrix                              T9
current→target cutover                          T10
implementation task graph                       T11
```

T8-E remains **NOT OPEN**. Implementation remains **BLOCKED**.

---

## 6. Independent-review scope

The next reviewer must reconstruct authority independently and attack the exact candidate rather than assuming this router or chat history is correct.

Material challenge surfaces:

```text
namespace/ownership strategy
cross-owner FK boundary
static Role/Permission vs RoleAssignment persistence
ProviderSubjectBinding current/history shape
VersionToken/OCC completeness
Revision.state + Release effectivity split
open/EFFECTIVE partial uniqueness under concurrency
current_submission_id boundedness
closed governance relational completeness
GROUP deletion/snapshot law
runtime DB-grant immutability
ManagedContent/AdmissionClaim/GC race safety
malware proof persistence
paired idempotency Key↔Replay deferred-FK feasibility
READ COMMITTED same-key loser behavior
River v0.37.1 schema/InsertTx compatibility
runtime-role vs DDL-owner grant model
lock ordering/deadlock surface
zero-trigger baseline
current-schema T8-A reuse dispositions
future-capability subtraction
T8-D/T8-E/T8-G/T10 boundary discipline
```

Reviewer evidence never becomes authority by itself.

---

## 7. Exact next action

```text
independent Fable review
of:
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md

→ reviewer reports BLOCKER / MAJOR / LOW findings
→ reviewer decides Global Maximum confirmed yes/no
→ reviewer decides upstream/T8-C reopen yes/no
→ reviewer checks T8-E/T8-G/T10 trespass
→ Lead confrontation/adjudication
→ bounded correction/re-review only if material
→ explicit operator ratification before durable T8-D promotion
```

Do **not** promote T8-D, open T8-E, create migrations or implement product code from this staging router/candidate.
