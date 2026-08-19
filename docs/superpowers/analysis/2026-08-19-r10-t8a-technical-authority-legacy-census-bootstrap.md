# R10-T8A — Technical Authority & Legacy Census — Stage Bootstrap

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **T8-A OPEN / CURRENT TECHNICAL CENSUS + REMEASUREMENT NEXT**  
> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

T8-A is the first substage of T8 Technical Realization Architecture. It classifies the current technical estate before any target package/database/API/frontend/runtime design is allowed.

T8-A is **not** a target architecture stage by itself. It answers what exists, what is still technically valuable, what is legacy accident, and which structures must be preserved/refined/re-homed/rewritten/deleted or treated as current-state evidence only.

## 1. Binding authority chain

Read in order:

1. `AGENTS.md`
2. DevelopmentConexus Engineering Method mirror
3. `wiki/references/current-agent-handoff.md`
4. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
5. T1→T7 durable authorities
6. Decision Registry + D4/T6/post-T6/T7 amendments
7. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
8. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
9. `wiki/architecture/r10-technical-architecture.md`
10. this bootstrap
11. current repo/runtime/schema/API/frontend/deploy/test evidence for each concrete claim

Current implementation proves **what exists**, not what survives.

## 2. Inherited T7 boundary

T7 closed with:

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Therefore:

```text
current MetalDocs DB/content/history = DEV / TEST / THROWAWAY
historical-data compatibility consumer = NONE
```

No current package/table/route/process receives survival entitlement to preserve disposable DEV/test business data.

Technical structures may still survive if current evidence proves they are the smallest sustainable realization of an already-ratified property.

## 3. T8-A disposition vocabulary

Every material current technical structure must receive one disposition:

```text
PRESERVE
  current structure/mechanism already fits the ratified target and has no concrete reason to change

REFINE
  current structure is fundamentally correct but needs bounded correction/hardening

REHOME
  useful behavior/mechanism survives but semantic/physical ownership must move

REWRITE
  required capability survives but current realization is structurally incompatible with target authority

DELETE
  no target consumer/property justifies the structure

CURRENT-STATE ONLY
  useful evidence about what runs today, but not a target decision/structure

SUPERSEDED
  prior technical target/design authority explicitly displaced by current R10 authority
```

Existence, sunk cost, test count or migration convenience are not disposition criteria.

## 4. Required census surfaces

T8-A must inspect at least:

### Backend/repository

```text
Go module/repository topology
apps/binaries roots
internal/modules/*
internal/platform/*
package import graph
module-level dependency graph/SCCs
composition/bootstrap roots
cross-owner/private imports
shared mechanism packages
```

### Persistence

```text
current schemas/tables/views/functions/triggers
module/table ownership records
raw SQL readers/writers
foreign-table reads/writes
cross-owner transactions
current constraints/invariants worth preserving
DEV-only data/state that has no migration entitlement
```

### API / contracts

```text
api/openapi/v1/openapi.yaml
operation/tag/package organization
Go generated packages
frontend generated TypeScript boundary
runtime request validation/conformance
legacy API structures contradicted by T6
```

### Frontend

```text
route tree
feature directories
cross-feature imports
query/cache/state organization
transport/generated-client usage
legacy Approval/Templates/Taxonomy/IAM/Documents/ControlledDocuments feature shapes
editor/viewer integration boundaries
```

### Async/render/runtime/deploy

```text
API/worker/jobs/renderer process roots
River/job ownership
outbox/worker mechanisms
renderer/providers
Docker/Compose/runtime scripts
startup/readiness/shutdown
configuration/secrets/trust/network evidence
observability/backup/restore runtime evidence
```

### Verification / CI

```text
tools/verify registry/profile truth
GitHub workflow composition
architecture guards
SQL ownership guards
contract/codegen checks
integration/E2E coverage
known allowlists/baselines/bypass surfaces
```

### Technical documentation authority

```text
backend-target-architecture.md
backend-blueprint.md
backend-api-structure.md
frontend-structure.md
data-model.md
repo-topology.md
module pages
ADRs/technical standards that may still constrain current mechanisms
```

Each document is classified as active reusable technical constraint, current-state evidence, superseded target, or deletion/rewrite candidate. No stale document may silently route T8-B+.

## 5. Evidence discipline

Use TRRB evidence classes while collecting the census:

```text
CURRENT-PROVEN
LAST-REPRODUCED
STALE / SUPERSEDED
UNKNOWN / REMEASURE
```

Old Aug-09 audit metrics are `LAST-REPRODUCED` until mechanically rerun against the current PR base/current source state when load-bearing.

Do not copy old counts into T8-A as current facts.

## 6. Structural Inversion test

Reject:

```text
legacy module/package/table exists
→ therefore fit the target around it
```

Instead:

```text
ratified semantic/property requirement
→ inspect current realization
→ prove current structure fits or fails
→ disposition it
```

Examples:

```text
legacy Approval module exists
  ≠ Approval is a target owner

legacy Templates module exists
  ≠ Template requires an independent target lifecycle/module

legacy public.documents exists
  ≠ target schema must retain that row shape

legacy API/worker/jobs process split exists
  ≠ target runtime must preserve that process count
```

## 7. First required work product

Create a T8-A technical census/disposition matrix containing, for each material structure:

```text
surface / structure
current evidence source
current role/behavior
ratified target property it serves, if any
current evidence class
preliminary disposition
reason
material dependency/coupling
unknown/re-measurement required
future owning substage if target design remains open
```

The first pass should prioritize load-bearing structures and systemic coupling rather than enumerate every leaf file equally.

## 8. T8-A must not decide

T8-A does not freeze:

```text
final package/module topology                 → T8-B
final owner-to-owner contracts                → T8-C
final relational schema/constraints           → T8-D
final OpenAPI executable schemas              → T8-E
final frontend route/feature/query topology   → T8-F
final process/deployment topology             → T8-G
whole realization coherence                   → T8-H
Golden Flow proof baseline                    → T9
transition/cutover implementation             → T10
implementation task graph                     → T11
```

A disposition may say `REWRITE` without prematurely deciding the exact replacement shape.

## 9. T8-A close gate

T8-A closes only after:

```text
current technical census sufficiently complete
→ load-bearing stale metrics remeasured
→ material structures dispositioned
→ documentation authority drift reconciled
→ material disagreements adjudicated
→ operator ratifies T8-A disposition summary
```

Only then may T8-B open.

T8-B→T12 and implementation remain blocked.
