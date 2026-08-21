# R10-T8A — Technical Authority & Legacy Disposition

> **Status:** CLOSED / OPERATOR-RATIFIED / PROMOTED  
> **Ratified:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Current program stage / implementation permission / next action:** `../roadmap.md`

This document is the durable T8-A authority for how R10 treats the existing MetalDocs technical realization while deriving the target physical architecture. Its closure/ratification facts are immutable; mutable program state is owned exclusively by `../roadmap.md`.

It does **not** define the final backend package topology, internal owner contracts, relational schema, exact wire contract, frontend realization, process/deployment topology, or transition plan. Those remain owned by T8-B→T8-G and T10 respectively.

## 1. Ratified Global Maximum

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

Binding interpretation:

> Derive the R10 physical architecture from the ratified product and semantic truth. Existing implementation is evidence, not target authority. A legacy shape receives no survival entitlement from existence, sunk cost, current test coverage, migration convenience, or prior ADR acceptance. Reuse is allowed only when the reusable mechanism independently remains the smallest sustainable solution for a named current R10 consumer.

This posture applies to code, packages, modules, tables, schemas, triggers, API routes, frontend features, runtime processes, deployment components, tests, guards, and technical documentation.

## 2. T7 compatibility consequence

T7 ratified:

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Therefore:

```text
current MetalDocs DB/content/history = DEV / TEST / THROWAWAY
current MetalDocs business history   = NONE
historical-data compatibility consumer = NONE
```

No current DEV/test data shape, API behavior, storage key, schema layout, or internal implementation receives preservation rights for historical-business compatibility.

T10 still owns technical current→target transition, reset/cutover/rollback, and legacy deletion.

## 3. Structural Inversion result

If the current implementation had been organized differently, the following already-ratified target properties would still be required:

```text
PostgreSQL-backed product-state transactions
single-company product semantics
Authentication / Organization / Authorization / Controlled Documents / Audit ownership topology
Document != Revision != WorkingContent != Submission
system-owned Release/effectivity
exact-content identity independent of storage-provider identity
one sequential governance route baseline
same-local-commit Audit for critical business changes
River-backed named durable jobs
OpenAPI 3.0.3 contract-first generated boundaries
purpose-built frontend user journeys/lenses
verification before implementation
```

Consequently, current physical shapes are not architecture authority merely because they embody some of those properties today.

## 4. High-confidence REWRITE / REHOME dispositions

The following current shapes are not R10 target authority and must be rederived by their owning T8 subgate:

```text
legacy semantic module/package topology
cross-owner communication and foreign-SQL boundaries
current persistent schemas/table families
current tenant_id + GUC + RLS mesh as target mechanism
current DB capability-assertion vocabulary/mechanism
current OpenAPI path/schema/tag-per-legacy-module surface
local-password authentication capability and delivery shape
current frontend feature/route topology
provider/storage-key semantic storage contract
current jobs registry and non-Launch capability wiring
legacy architecture-specific verification policies
```

A `REWRITE` disposition in T8-A means the existing shape is not inherited. It does not preselect the replacement shape.

## 5. DELETE / DEFER dispositions

Absent a later named current Launch consumer, the following implementation has no Launch survival entitlement:

```text
local credential/password capability
Distribution implementation
Periodic Review implementation
approval-SLA machinery
tenant lifecycle / tenant export / tenant erasure product machinery
legacy approval delegation/quorum/fast-forward/reassign-like machinery outside the ratified baseline
DEV/test historical compatibility infrastructure
legacy provider-key semantic fields/contracts
```

Redis, multi-replica-specific substrate, process count, container count, and current provider topology are not preserved by T8-A. T8-G must independently prove any current Launch need before they survive.

The rule is: **defer the capability, not dormant implementation**.

## 6. PRESERVE properties / mechanisms

The following survive T8-A because they are already ratified upstream or have independently demonstrated a current R10-relevant property:

```text
PostgreSQL product-state substrate
River durable-job mechanism for named T5 jobs
contract-first OpenAPI + generated Go/TypeScript boundaries
verification registry / local-CI SSOT model
runtime database identity separated from schema/DDL ownership
reproducible deterministic database bootstrap/proof property
exact-content SHA-256 / size / fail-closed proof principles
```

`PRESERVE` does **not** freeze their current package, file, table, process, or deployment arrangement unless a later T8 subgate independently chooses it.

Examples:

- preserving PostgreSQL does not preserve current tables, RLS policies, triggers, or schemas;
- preserving River does not preserve the current jobs registry or process topology;
- preserving contract-first does not preserve the current OpenAPI document or tag/module arrangement;
- preserving the verification control plane does not preserve guards that encode superseded architecture;
- preserving exact-content proof principles does not preserve storage provider keys as semantic references.

## 7. Selective reuse gate

A current implementation unit may be reused only when **all five** are proven:

```text
1. a named current R10 consumer exists;
2. the unit's public contract contains no legacy semantic authority;
3. its dependency direction fits the accepted target;
4. its tests/proof assert the target property rather than the legacy shape;
5. reuse remains simpler/smaller than rewrite after transition cost is considered.
```

Fail any one condition → the unit has no `PRESERVE` entitlement.

High local code quality alone is insufficient.

## 8. Current-state evidence conclusions

T8-A established enough current evidence to classify the estate without turning archaeology into a preservation exercise:

- API, worker, and jobs composition roots directly wire a broad set of legacy modules/capabilities;
- current durable jobs include capabilities outside Launch;
- current PostgreSQL baseline spreads pooled-tenancy/RLS/GUC logic across many table families;
- current OpenAPI still exposes legacy/local-auth/tenant/delegation/etc. surfaces inconsistent with T6;
- current frontend routes/features are organized around legacy domains rather than the T6 user lenses;
- current storage code contains useful hashing/fail-closed behavior but leaks provider/storage keys into semantic/persistent contracts;
- current architecture lint baseline directly proves cross-owner raw-SQL reads/writes;
- several technical wiki/ADR routing documents still described the superseded topology as canonical/target.

The old Aug-09 `55 foreign reads + 12 foreign writes` metric remains `LAST-REPRODUCED`. Exact recount was not load-bearing to the T8-A Global Maximum because current leakage itself is directly proven. Remeasure when a later transition/proof decision actually depends on exact magnitude.

## 9. Explicitly undecided

T8-A intentionally does not decide:

```text
exact target Go packages or module count
one Go module vs another source layout
exact owner interfaces / dependency graph
exact target tables / constraints / RLS posture
exact OpenAPI operations / schemas
React Router / TanStack Query / Zustand survival
frontend folder/query/cache realization
interactive DOCX provider
renderer/converter provider
number of processes or containers
Redis survival
exact deployment topology
current→target transition sequence
```

These belong to T8-B→T8-G and T10.

## 10. Technical-document authority law

Historical/current-state technical documentation remains useful evidence, but words such as `canonical`, `target`, `MUST`, or `industry-grade` in documents describing the old topology do not override R10.

For target architecture the authority chain is:

```text
Product Contract REV001
→ Whole-Product GCR + 4+1 ownership topology
→ T1→T8-A durable authorities
→ current routed T8 authorities
→ current evidence only for concrete implementation facts
```

Current routing is resolved through `../index.md` and mutable stage/implementation/next-action state through `../roadmap.md`. Imported pre-reset `wiki/...` strings are provenance only under documentation governance.

Historical ADR `Accepted` status does not imply R10 inheritance unless a current R10 authority explicitly preserves the property.

## 11. Reopen triggers

Reopen T8-A only if material evidence proves one of the following:

```text
an upstream ratified R10 decision materially changes
an existing physical shape is discovered to be an externally binding compatibility contract
repository/toolchain constraints prove clean target realization impossible or materially inferior
one of the preserved mechanisms loses its named R10 consumer or proof basis
a supposedly deleted/deferred capability gains a concrete Launch consumer
```

Preference, sunk cost, implementation convenience, or hypothetical future extensibility are not reopen triggers.

## 12. Ratification closure and handoff

```text
T8-A Technical Authority & Legacy Census = CLOSED / OPERATOR-RATIFIED / PROMOTED
```

At T8-A ratification, the immediate downstream consumer was T8-B — Backend Module & Package Topology. That handoff is immutable historical progression evidence, not current program status.

T8-B must derive backend/package boundaries from ratified ownership and semantics without inheriting the legacy module map.

Current stage, integration status, implementation permission and exact next action are owned exclusively by `../roadmap.md`.
