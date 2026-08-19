# R10-T7 — Historical Migration Truth & Semantic Mapping — Stage Bootstrap

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **T7 OPEN / SOURCE EVIDENCE CENSUS NEXT**  
> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

T7 is the first stage opened under the operator-ratified post-T6 Implementation Readiness Program.

T7 owns **historical/source truth and semantic migration mapping only**. It does not own target package/database/API/frontend/runtime realization and it does not own concrete migration/cutover implementation.

---

## 1. Binding authority chain

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md`
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t3-d4-responsible-owner-eligibility-amendment.md`
11. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
12. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
13. `wiki/architecture/r10-t6-canonical-api-frontend-journeys.md`
14. `wiki/architecture/rebaseline-decision-registry.md`
15. `wiki/architecture/rebaseline-decision-registry-d4-amendment.md`
16. `wiki/architecture/rebaseline-decision-registry-t6-amendment.md`
17. `wiki/architecture/rebaseline-decision-registry-post-t6-amendment.md`
18. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
19. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
20. `wiki/architecture/r10-technical-architecture.md`
21. this bootstrap
22. actual source-system/runtime/data evidence only for a concrete T7 claim

Legacy code/schema/content is evidence, never target authority by existence.

---

## 2. T7 official scope

T7 decides only:

```text
actual legacy/source evidence census
PROVEN / INFERABLE WITH EXPLICIT RULE / UNKNOWN classification
CURRENT_STATE / FULL_HISTORY or a smaller real migration-mode set
which source facts become target-owned imported facts
which source facts remain provenance-only evidence
source document/revision identity quality
source revision/ordinal mapping
exact-content/file provenance quality
source actor/owner identity quality
source approval/governance provenance quality
semantic migration unit definition
truthful representation of partial/ambiguous/unknown historical evidence
```

T7 answers:

> **What historical/source truth can MetalDocs legitimately carry forward, and how may that truth map into the already-ratified semantic target without fabrication?**

---

## 3. Explicitly out of T7

T7 does **not** decide:

```text
backend package/module topology                    → T8-B
internal package/owner communication realization  → T8-C
target relational schema/tables/constraints        → T8-D
exact executable OpenAPI schemas/codegen           → T8-E
frontend route/feature/query/cache topology        → T8-F
runtime binaries/processes/deploy/observability    → T8-G
whole physical realization coherence               → T8-H
Golden Flow proof architecture                      → T9
migration tooling/runtime implementation            → T10
dry-run implementation                              → T10
migration idempotency/reconciliation implementation → T10
production cutover/readiness/rollback/deletion      → T10
restore/offboarding cutover choreography            → T10
implementation decomposition                        → T11
product code                                        → BLOCKED
```

If a T7 question cannot be answered without selecting one of those physical mechanisms, preserve the semantic requirement/unknown and route the realization to its owning later stage.

---

## 4. Inherited non-negotiable laws

```text
unknown source truth remains unknown
inference is labeled and governed by an explicit rule
imported history never becomes fake native MetalDocs history
migration convenience never changes Document/Revision/Submission/Release meaning
native governance evidence is never fabricated
provider/storage identity never becomes semantic identity
exact imported bytes remain governed by T4 exact-content law
current T3 security/offboarding truth remains authoritative for serving
historical side effects/jobs/notifications are never replayed as new current effects
Historical Migration is not a generic Interchange/ETL domain
prepare migration seams only where the actual migration requires them
```

---

## 5. First required work — actual source evidence census

Before selecting migration modes or mapping shapes, inspect what the actual current/legacy MetalDocs system can prove.

The census must examine, where applicable:

```text
stable document identifiers/codes
revision/version identifiers and ordinals
current/effective/obsolete/cancelled state evidence
exact source files/bytes and formats
title and metadata by historical revision
area/document-type/template relationships
responsible owner/author identities
approval/governance actors, decisions, outcomes and timestamps
release/effectivity evidence
source audit/history evidence and its trust limits
current authn/user/group/access facts relevant to identity mapping
content gaps, duplicates, orphan records and contradictions
erasure/offboarding/deletion evidence relevant to truthful migration
```

Every material source claim is classified:

```text
PROVEN
INFERABLE WITH EXPLICIT RULE
UNKNOWN / NOT RELIABLY PROVABLE
```

No default may convert `UNKNOWN` into plausible historical truth.

---

## 6. Evidence-source priority

For current source truth, prefer direct source evidence in this order as applicable:

```text
current DB schema/data constraints and actual rows
current object-store/file evidence
current application/runtime semantics where needed to interpret stored facts
current OpenAPI/code only where needed to understand source meaning
current maintained data dictionary/current-state docs as supporting evidence
historical docs/Git only to explain provenance or older shapes
```

A historical document cannot overrule current source data merely because it describes a cleaner model.

---

## 7. Core T7 proof questions

The eventual T7 candidate must answer at least:

1. Which source truth is actually required to launch after replacement of the current system?
2. What evidence exists for current state versus historical revisions?
3. Is `CURRENT_STATE` sufficient, is `FULL_HISTORY` justified, or is a smaller mixed mode the Global Maximum?
4. How are source document/revision identities and ordinals represented without falsifying target `REV000`/`REV001...` semantics?
5. Which historical metadata becomes imported target-owned state and which remains provenance-only?
6. How are source content bytes associated with the correct historical semantic unit?
7. When approval/governance actor/time/outcome is incomplete, what can be represented truthfully?
8. What is the smallest semantic migration unit whose truth must remain internally coherent?
9. Which contradictions/gaps must block migration of a unit versus remain explicit unknown provenance?
10. What source facts are deliberately not migrated because they are legacy implementation accidents or deferred capabilities?

These are semantic truth questions. Concrete tables, import workers and cutover scripts belong later.

---

## 8. Global Maximum test

Reject the local maximum:

```text
preserve legacy schema/status/workflow semantics because they are easy to copy
```

Reject the speculative maximum:

```text
build a generic ETL/interchange/history framework for hypothetical future imports
```

Target:

```text
actual source evidence
+ explicit truth classification
+ smallest required migration modes
+ target-owner semantic mapping
+ explicit provenance for uncertainty
→ truthful historical migration contract
```

---

## 9. T7 process

T7 is architectural and evidence-first:

```text
actual source evidence census
→ PROVEN / INFERABLE / UNKNOWN
→ identify only material unknowns
→ compare 2–3 truthful migration-semantic approaches
→ Global Maximum candidate
→ material adjudication
→ platform-facing T7 summary
→ explicit operator summary ratification
→ durable T7 promotion + Registry reconciliation + staging cleanup
```

Only after T7 closes may T8 open.

T8→T12 and implementation remain blocked.
