# R10-T7 — Historical Migration & Cutover — Stage Bootstrap

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **T7 OPEN / EVIDENCE CENSUS NEXT**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

T7 is the final technical design stage of the active R10 rebaseline. It is opened only after T6 was operator-ratified, promoted to durable `wiki/` authority, reconciled into the Decision Registry and its completed staging removed from the live tree.

T7 owns **Historical Migration & Cutover only**. It does not reopen the accepted product architecture by default and it is not a generic Interchange/integration-platform design stage.

---

## 1. Binding authority chain

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md` — REV001
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
17. `wiki/architecture/r10-technical-architecture.md`
18. this bootstrap
19. actual source-system evidence only where needed

Current implementation/source data is evidence, never target authority by existence.

---

## 2. T7 official REOPEN set

T7 deliberately designs only:

```text
actual source evidence census
CURRENT_STATE / FULL_HISTORY or a smaller real migration-mode set
imported target-owned fact shapes
ordinal/content/governance provenance
plan / dry-run / idempotency / reconciliation
semantic-unit atomicity
cutover / readiness / rollback / deletion map
concrete restore/erasure and post-snapshot security-teardown reconciliation choreography where cutover/recovery requires it
```

Everything else is frozen unless actual migration evidence proves a material contradiction.

---

## 3. T7 non-negotiable inherited laws

```text
imported history never becomes fake native MetalDocs history
unknown source truth stays unknown
migration writes through owning semantic seams
Historical Migration != generic Interchange domain
provider/storage identity never becomes semantic identity
Document/Revision/Submission/Release meanings do not change for migration convenience
native governance evidence is never fabricated
external/provider work never joins a local semantic transaction
T4 exact-content admission/integrity laws apply to imported bytes
T3 current authorization/offboarding truth remains authoritative at cutover
T4 restore readiness must not resurrect erased PII, restored sessions or known post-snapshot access teardown
T5 jobs/effects remain mechanisms and historical migration never replays old side effects as new current events
```

---

## 4. Hard boundaries

T7 does **not** decide or implement:

```text
new Launch product capabilities
new semantic owners absent a proven source/target contradiction
new generic import/export/repository platform
new workflow engine
new Search architecture
new API/frontend product journeys
implementation code
final SQL migration scripts
production cutover execution
```

A migration convenience is not sufficient reason to weaken a ratified invariant.

---

## 5. First required work — source evidence census

Before choosing migration modes or imported fact shapes, T7 must establish what the actual legacy/source system can truthfully prove.

Evidence census must identify, where present:

```text
stable source document identifiers/codes
source revision/version identifiers and ordinals
current/effective/obsolete/cancelled state evidence
exact source content/files and content format
source title/metadata by version
source authors/owners/actors and identity quality
source approvals/reviews/decisions and whether actor/time/outcome are provable
source timestamps and their trust/meaning
source template relationships/provenance
source area/type/category mappings relevant to accepted target concepts
source current access/security facts needed for cutover
source deletion/erasure/offboarding evidence relevant to restore/cutover safety
source inconsistencies/duplicates/gaps/unknowns
```

The census must distinguish:

```text
PROVEN
INFERABLE WITH EXPLICIT RULE
UNKNOWN / NOT RELIABLY PROVABLE
```

T7 may not convert UNKNOWN into invented native truth.

---

## 6. Core proof questions

The T7 candidate must answer at least:

1. What smallest migration modes are actually required by source evidence and rollout needs?
2. Which source facts become target-owned imported facts, and which remain provenance-only?
3. How are source revision ordinals mapped without falsifying the `REV000` target convention?
4. How are exact bytes admitted and bound under T4 without provider identity leakage?
5. How are source approvals/actors/timestamps represented when they are trustworthy, partial or absent?
6. What is one semantic migration unit and what must be atomic inside it?
7. How does retry avoid duplicate imported facts/content while remaining restart-safe?
8. What does dry-run prove before writes are allowed?
9. How does reconciliation prove imported target truth matches the migration plan/source evidence?
10. What is the cutover readiness barrier and when are legacy writes/read paths disabled?
11. What rollback is still possible before/after target serving begins?
12. What source data may be deleted, retained or frozen after cutover, and who owns that decision?
13. How does cutover/restore prove no known erased PII or post-snapshot offboarded/revoked access is resurrected?
14. How are historical side effects/jobs/notifications prevented from replaying as new current effects?

---

## 7. Global-Maximum test for T7

The preferred T7 design is the **smallest truthful migration/cutover mechanism that can move the actual required source truth into the ratified target without fabricating history or creating a permanent migration platform**.

Reject both extremes:

```text
Local Maximum:
  preserve legacy schema/semantics because migration is easier

Speculative Maximum:
  build a generic ETL/interchange/repository framework for hypothetical future imports
```

Target:

```text
actual source evidence
+ explicit target-owner mapping
+ bounded migration tooling
+ deterministic plan/dry-run/reconcile
+ safe cutover/readiness
→ truthful one-time/operational migration capability
```

---

## 8. Stage process

T7 is architectural. Before substantive design:

```text
source evidence census
→ clarify only material unknowns
→ propose 2–3 migration/cutover approaches
→ compare truthfulness / complexity / rollback / operational risk
→ candidate T7 design
→ material decision adjudication
→ platform-facing T7 summary
→ explicit operator summary ratification
→ durable promotion + Registry reconciliation + staging cleanup
```

After T7 closes:

```text
Integrated Whole-R10 Global Coherence Review
→ cold independent final review
→ operator final ratification
→ implementation spec/plan
→ code
```

Implementation remains **BLOCKED**.