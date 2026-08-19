# Rebaseline Decision Registry — T8-A Closure Amendment

> **Status:** ACTIVE / OPERATOR-RATIFIED REGISTRY RECONCILIATION  
> **Ratified:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **T8-A authority:** `wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md`

This bounded amendment reconciled the Decision Registry after T8-A closure. It changes technical-realization inheritance/disposition only; it does not rewrite unrelated Product Contract or T1→T7 semantics.

The registry chain has since been extended by the operator-ratified T8-B amendment:

```text
rebaseline-decision-registry.md
→ rebaseline-decision-registry-d4-amendment.md
→ rebaseline-decision-registry-t6-amendment.md
→ rebaseline-decision-registry-post-t6-amendment.md
→ rebaseline-decision-registry-t7-amendment.md
→ rebaseline-decision-registry-t8a-amendment.md
→ rebaseline-decision-registry-t8b-amendment.md
```

For current stage/status use the R10 router, not this historical amendment's former "next stage" section.

## 1. T8-A stage disposition

```text
T8-A Technical Authority & Legacy Census = CLOSED / OPERATOR-RATIFIED / PROMOTED
```

Ratified Global Maximum:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

## 2. Registry inheritance law

For later realization stages:

```text
existing implementation = evidence only
historical ADR acceptance = evidence only unless current R10 explicitly retains the property
sunk cost / test count / migration ease = no survival entitlement
PRESERVE must be independently justified
REWRITE / REHOME / DELETE are valid first-class outcomes
```

No later stage may use current package/table/API/frontend/runtime existence as a substitute for deriving the target from ratified semantics.

## 3. Preserved properties/mechanisms

The following are retained into later realization stages:

```text
PostgreSQL product-state substrate
River durable-job mechanism for named T5 jobs
contract-first OpenAPI + generated Go/TypeScript boundaries
verification registry / local-CI SSOT model
runtime DB identity separated from schema/DDL ownership
reproducible deterministic DB bootstrap/proof property
exact-content SHA-256 / size / fail-closed proof principles
```

Retention is property-level unless a later subgate independently selects the current physical implementation.

## 4. Reopened / rewrite-owned technical surfaces

The following are explicitly not inherited as target shapes:

```text
legacy semantic module/package topology
current cross-owner communication / foreign SQL
current schemas/table families
current tenant/GUC/RLS mesh
current DB capability-assertion mechanism/vocabulary
current OpenAPI route/schema/tag topology
local-password AuthN realization
current frontend feature/route topology
provider-key semantic storage contracts
current jobs registry/non-Launch wiring
legacy target-specific architecture guards
```

Replacement ownership remains:

```text
T8-B = backend module/package topology — NOW CLOSED / PROMOTED
T8-C = internal communication contracts — CURRENT ACTIVE STAGE
T8-D = persistence realization
T8-E = executable wire contract
T8-F = frontend realization
T8-G = runtime/process/deployment realization
T8-H = integrated T8 coherence
T10  = current→target transition/deletion/cutover
```

## 5. Delete/defer law

Non-Launch implementation has no preservation right absent a named current Launch consumer.

This includes legacy implementation for:

```text
local passwords
Distribution
Periodic Review
approval SLA machinery
tenant lifecycle/export/erasure
legacy approval delegation/quorum/fast-forward/reassign-like behavior outside the ratified baseline
DEV/test historical compatibility
```

Do not preserve dormant implementation merely to keep a future seam.

## 6. Selective reuse gate

A current implementation unit may survive only if all are proven:

```text
named current R10 consumer
+ public contract free of legacy semantic authority
+ dependency direction fits target
+ proof asserts target property rather than legacy shape
+ reuse remains smaller than rewrite after transition cost
```

Failure of any condition removes `PRESERVE` entitlement.

## 7. Measurement law

Old exact mechanical metrics remain `LAST-REPRODUCED` until remeasured.

Remeasure only when the metric is load-bearing to a material later decision. T8-A does not require ritual reproduction of old counts when direct current evidence already proves the property needed for disposition.

## 8. Successor routing

T8-A no longer owns next-stage routing. Current routing is:

```text
wiki/architecture/r10-technical-architecture.md
```

T8-B closure authority:

```text
wiki/architecture/r10-t8b-backend-module-package-topology.md
wiki/architecture/rebaseline-decision-registry-t8b-amendment.md
```

Current active stage is T8-C. Implementation remains **BLOCKED**.