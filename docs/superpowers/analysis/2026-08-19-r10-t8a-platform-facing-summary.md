# R10-T8A — Platform-Facing Summary

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **OPERATOR RATIFICATION TARGET**  
> **Date:** 2026-08-19  
> **Stage:** T8-A Technical Authority & Legacy Census  
> **Implementation:** BLOCKED

## Proposed T8-A decision

> **Derive the R10 physical architecture cleanly from ratified product/semantic truth. Give no survival entitlement to legacy technical shapes. Reuse existing mechanisms only when their value survives Structural Inversion and is independently proven for a current R10 consumer.**

Preferred strategy:

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

This is a design-freedom decision, not a package/schema/runtime design.

## Why

Current implementation materially encodes an older product:

```text
~15 legacy semantic modules
pooled tenant/RLS substrate
parallel Documents / ControlledDocuments / Templates concepts
legacy Approval delegation/SLA/review machinery
local-password authentication
legacy module-tag API surface
legacy domain-feature frontend topology
provider/storage keys in semantic rows
non-Launch jobs/capabilities
```

Current evidence also proves cross-module persistent ownership leakage and stale technical documentation that still calls parts of the old topology `canonical`.

T7 removed the compatibility justification:

```text
current business data/history = DEV / TEST / THROWAWAY
historical-data compatibility consumer = NONE
```

Therefore incremental preservation would optimize migration convenience rather than the accepted product.

## High-confidence dispositions

### Rewrite / rehome

```text
legacy semantic module/package topology
current cross-owner communication and foreign SQL boundaries
current schemas/tables and parallel document/template/approval data model
current tenant/GUC/RLS mesh as a target mechanism
current DB capability-assertion vocabulary/mechanism
current OpenAPI routes/schemas/tag-per-legacy-module shape
local-password AuthN capability and current AuthN delivery shape
current frontend feature/route topology
current storage/objectstore public contract where provider keys are semantic references
current jobs registry and non-Launch capability wiring
legacy architecture-specific verification policies as target policies
```

### Delete / defer unless a later named Launch consumer proves otherwise

```text
local credential/password capability
dormant Distribution/Periodic Review/approval-SLA/tenant-lifecycle implementation in Launch
legacy delegation/quorum/fast-forward/reassign-like machinery outside ratified baseline
DEV/test historical compatibility infrastructure
Redis/multi-replica substrate if T8-G cannot prove a current Launch need
legacy provider-key semantic fields/contracts
```

### Preserve as already-ratified or independently proven properties/mechanisms

```text
PostgreSQL product-state substrate
River durable-job mechanism for named T5 jobs
contract-first OpenAPI + generated Go/TypeScript boundaries + drift/conformance proof
tools/verify registry / local-CI verification SSOT model
runtime DB identity separated from schema/DDL ownership
reproducible deterministic DB bootstrap/proof property
exact-content SHA/size/fail-closed proof principles from T4
```

`PRESERVE` here does **not** preserve current package/file/process/table arrangement unless a later T8 subgate independently chooses it.

## Selective reuse gate

A current implementation unit may be reused only if all five are true:

```text
1. a named current R10 consumer exists;
2. its public contract contains no legacy semantic authority;
3. its dependency direction fits the accepted target;
4. its tests/proof assert the target property rather than the legacy shape;
5. reuse is still simpler/smaller than rewrite after transition cost is considered.
```

Fail one → no PRESERVE entitlement.

## Explicitly not decided yet

T8-A does not decide:

```text
exact target Go packages or module count
exact dependency graph / owner interfaces
exact target tables/constraints/RLS posture
exact OpenAPI operations/schemas
React Router / TanStack Query / Zustand survival
frontend directory/query/cache realization
interactive DOCX provider
renderer/converter provider
number of processes/containers
Redis survival
exact deployment topology
current→target transition sequence
```

These remain T8-B→T8-G and T10 decisions.

## Evidence law going forward

- Current implementation = evidence, never target authority.
- Exact stale counts are remeasured only when load-bearing to a material decision.
- Historical ADR `Accepted` status does not imply R10 inheritance.
- Current technical docs that describe the old topology remain archaeology only and must be rerouted/marked before T8-A closure so Fresh Actors cannot mistake them for target authority.

## Gate after ratification

```text
operator ratifies this summary
→ promote T8-A durable authority to wiki/
→ reconcile Decision Registry
→ repair stale technical-document / ADR routing labels
→ remove completed T8-A staging
→ mark T8-A CLOSED
→ only then open T8-B Backend Module & Package Topology
```

T8-B→T12 and product implementation remain blocked until their own gates close.
