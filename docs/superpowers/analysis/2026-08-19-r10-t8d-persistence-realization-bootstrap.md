# R10-T8D — Persistence Realization — Bootstrap

```text
ACTIVE STAGING ROUTER
NON-AUTHORITATIVE
ROUND-1 INDEPENDENT REVIEW COMPLETE
LEAD ADJUDICATION COMPLETE / OPERATOR-APPROVED
ADJUDICATED CORRECTED CANDIDATE MATERIALIZED
BOUNDED FABLE ROUND 2 = NEXT
NOT IMPLEMENTATION AUTHORIZATION
```

> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Stage:** T8-D ACTIVE  
> **Upstream authority:** `wiki/architecture/r10-t8c-internal-communication-contracts.md`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This file is the active non-authoritative staging router for R10 **T8-D — Persistence Realization**. It does not contain target authority.

## Active staging chain

```text
original Global Maximum candidate
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-global-maximum-candidate.md

Round-1 independent Fable review
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-independent-fable-review.md

Lead-adjudicated corrected candidate — ACTIVE ROUND-2 INPUT
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-adjudicated-corrected-candidate.md
```

For Round 2, the effective corrected candidate is the original candidate plus the adjudicated correction overlay. Where they conflict, the corrected candidate controls staging.

No T8-D decision is durable until bounded Round 2, final Lead adjudication and explicit operator ratification promote one consolidated authority into `wiki/`.

---

## 1. Exact T8-D question

> **What is the smallest PostgreSQL persistence realization that makes T1→T8-C invariants and internal contracts structurally enforceable, assigns every persistent fact to its ratified semantic/mechanism owner, and maps required ACID/OCC/serialization behavior to explicit schema/constraint/query/lock rules without foreign SQL, duplicate truth, hidden shared write authority, wire leakage or speculative persistence?**

---

## 2. Binding authority chain

Read in repository authority order:

```text
AGENTS.md
root-cause-global-maximum-method.md
current-agent-handoff.md
r10-technical-architecture.md — sole status/next-action router
Product Contract REV001 + Whole-Product GCR + 4+1 ownership
T1→T8-C durable authorities
Decision Registry + amendments through T8-C
post-T6 implementation-readiness program
TRRB / technical-realization reconciliation baseline
this bootstrap
original candidate
Round-1 review evidence
adjudicated corrected candidate
current schema/migrations/SQL only for concrete evidence/reuse claims
```

Current implementation is evidence only. Reviewer artifacts and staging candidates are evidence/input only, never target authority.

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
same-Scope AdmissionClaim consumption where applicable
protected eligibility serialization with offboarding
self-contained PII-free ReplaySnapshot
ManagedContent mechanism != semantic authority
Search materialization OFF
```

---

## 4. Global Maximum class — Round 1 confirmed

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

Round-1 Fable verdict:

```text
APPROVE WITH MATERIAL FIXES
BLOCKER 2 / MAJOR 11 / LOW 10
Global Maximum class = CONFIRMED
T8-C reopen = NO
T8-B/T1→T7 reopen = NO
stage trespass = NO
```

Lead adjudication accepted both blockers, accepted or narrowed the material corrections, rejected M7's proposed Company/company_id subtraction, and selected exact corrected realizations. The operator explicitly approved that adjudication.

---

## 5. Corrected material delta

Round 2 must verify only the adjudicated delta unless it creates a new contradiction:

```text
River custom schema ownership/provisioning/grants on PG16
+ River self-REINDEX OFF
+ third-party catalog class

universal semantic ManagedContent attachment FOR SHARE
vs GC FOR UPDATE

governance Decision -> frozen candidate composite FK
immutable platform.malware_inspections evidence
current-only provider issuer+subject uniqueness
actor/target User one sorted lock class
complete transaction-operation census
GovernanceAttempt subject uniqueness
M7 Lead rejection: Company is current semantic scope; company_id retained
static Go <-> DDL vocabulary parity proof
idempotency retention Replay->Key
HMAC-SHA-256 semantic fingerprint + key version
provisioner/owner/runtime/CI DB trust classes
accepted LOW constraint/proof precision corrections
```

---

## 6. Stage boundaries

T8-D owns relational persistence, constraints, owner-private queries, immutable-history enforcement, OCC/VersionToken physical realization, and transaction/serialization/lock mapping.

It does **not** own:

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

## 7. Exact next action

```text
BOUNDED FABLE ROUND 2

review the exact corrected delta in:
  docs/superpowers/analysis/2026-08-19-r10-t8d-persistence-realization-adjudicated-corrected-candidate.md

using original candidate + Round-1 review only as provenance/context

→ verify every Round-1 blocker is actually closed
→ attack accepted/narrowed corrections and M7 Lead rejection
→ report new/surviving BLOCKER / MAJOR / LOW
→ report Global Maximum still confirmed yes/no
→ report T8-C/T8-B/T1→T7 reopen yes/no
→ report T8-E/F/G/T10/T11 trespass yes/no
→ state whether any further review round is materially required
→ final Lead adjudication
→ explicit operator ratification
→ only then durable T8-D promotion and T8-E opening
```

Do **not** promote T8-D, open T8-E, create target migrations or implement product code from this staging chain.
