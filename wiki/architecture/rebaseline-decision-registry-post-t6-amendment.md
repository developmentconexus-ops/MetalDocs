# Rebaseline Decision Registry — Post-T6 Program Amendment

> **Status:** ACTIVE / OPERATOR-RATIFIED REGISTRY RECONCILIATION  
> **Ratified:** 2026-08-19  
> **Parent registry:** `wiki/architecture/rebaseline-decision-registry.md`  
> **Prior amendments:** D4 + T6 closure  
> **Program authority:** `wiki/architecture/r10-post-t6-implementation-readiness-program.md`  
> **Implementation:** BLOCKED

This amendment changes only the **stage ownership of unresolved post-T6 design questions** after the operator-ratified Stage-Decomposition Global Coherence Review.

It does not change the accepted meaning of Product Contract REV001 or T1→T6.

Everything in the parent registry and prior amendments not named here remains unchanged.

---

## 1. Prior T7 allocation is superseded

The former T7 REOPEN set combined two materially different layers:

```text
A. truthful historical/source semantic mapping
B. physical migration/cutover/recovery realization
```

That combined stage is superseded because physical realization depends on a target technical architecture that was not yet frozen.

Method outcome:

```text
RESTRUCTURE NOW — stage decomposition only
```

---

## 2. New T7 ownership — Historical Migration Truth & Semantic Mapping

The following unresolved meanings remain in T7:

```text
actual source evidence census
PROVEN / INFERABLE WITH EXPLICIT RULE / UNKNOWN classification
CURRENT_STATE / FULL_HISTORY or a smaller real migration-mode set
imported target-owned fact shapes vs provenance-only evidence
source revision/ordinal mapping
exact-content provenance quality
source actor/governance provenance quality
semantic migration unit definition
truthful representation of partial/unknown historical evidence
```

T7 must not fabricate native history and may not change target semantics for migration convenience.

---

## 3. T8 ownership — Technical Realization Architecture

The Stage-Decomposition GCR exposed material target-realization questions not previously closed by the registry:

```text
backend module/package topology
allowed/forbidden dependency graph
internal owner communication realization
physical persistence/table/constraint ownership
exact executable OpenAPI wire contract
frontend route/feature/query/cache realization
runtime/process/jobs/deploy/trust/observability realization
```

These are now explicitly owned by T8 and its A→H subgates.

They are not implementation-task freedom.

---

## 4. T9 ownership — Golden Flows & Validation Baseline

T9 owns the composed proof architecture for the ratified system:

```text
cross-layer Golden Flows
negative/failure/race/restart/security cases
proof-method allocation
implementation Validation Baseline
```

Existing safety controls remain live until replaced by accepted equal-or-stronger proof/enforcement.

---

## 5. T10 ownership — Transition / Refactor / Migration / Cutover

The following items are removed from T7 and reassigned to T10 after T8 physical realization is frozen:

```text
concrete migration plan/runtime tooling
concrete dry-run execution mechanism
concrete migration idempotency/restart mechanism
concrete reconciliation mechanism
physical semantic-unit transaction/atomicity realization
current→target schema/data transition
current→target code/API/frontend/runtime transition
cutover readiness barrier
rollback choreography/windows
legacy freeze/deletion map
concrete restore/erasure/post-snapshot security-teardown reconciliation choreography
production cutover execution design
```

T10 remains constrained by T1→T9 and cannot weaken them for easier migration.

---

## 6. T11/T12 ownership

```text
T11 = bounded implementation Execution Graph; architecture decisions prohibited inside Writer tasks
T12 = independent/adversarial implementation-readiness challenge
```

Implementation remains blocked through both stages and final integrated review.

---

## 7. Current registry-routing state

```text
T1→T6  CLOSED / OPERATOR-RATIFIED
T7     NOT OPEN
T8     NOT OPEN
T9     NOT OPEN
T10    NOT OPEN
T11    NOT OPEN
T12    NOT OPEN
```

Prerequisite before T7:

`docs/superpowers/analysis/2026-08-19-r10-technical-realization-reconciliation-baseline.md`

The operator reviews that census for coverage/evidence classification before the redefined T7 may open.
