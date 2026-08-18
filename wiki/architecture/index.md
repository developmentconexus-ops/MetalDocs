# Architecture

> **Last verified:** 2026-08-18
> **Scope:** Durable system architecture truth and active target-design routing.

## Active target design — read first

- **[launch-v1-product-contract.md](launch-v1-product-contract.md)** — **ACCEPTED Launch V1 product authority.** Defines required product capabilities, journeys, invariants and scope tiers.
- **[whole-product-alignment-review.md](whole-product-alignment-review.md)** — **OPERATOR-ADJUDICATED Whole-Product GCR authority.** A1–A10 are accepted.
- **[launch-v1-ownership-topology.md](launch-v1-ownership-topology.md)** — **OPERATOR-APPROVED Launch semantic ownership authority.** Launch topology is 4 business owners + Audit, with the binding future-evolution law: defer capability, preserve the evolution seam.
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — **ACTIVE REBASELINED T1→T7 TECHNICAL-STAGE AUTHORITY.** Every T-stage requires material-decision adjudication **and** operator approval of a platform-facing summary before the next T-stage may open.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — exact fresh-session recovery point / current gate.
- [../standards/root-cause-global-maximum-method.md](../standards/root-cause-global-maximum-method.md) — binding DevelopmentConexus Engineering Method v1.0.0 mirror.

Current gate:

```text
T1 technical decisions = ACCEPTED
T1 platform summary    = OPERATOR RATIFICATION NEXT
T2                     = NOT OPEN
implementation         = BLOCKED
```

## Prior redesign / technical evidence

- [cohesive-platform-redesign.md](cohesive-platform-redesign.md) — prior-design evidence compatibility entrypoint. Its former 8+3 Launch topology and old stage routing are superseded.
- [launch-v1-scope-rebaseline.md](launch-v1-scope-rebaseline.md) — narrow operator-approved Records-Governance defer overlay; subordinate to Product Contract/GCR/topology.
- Former B1–B6/R10-C staging material remains evidence only where current authorities re-prove its conclusions.

## Active Launch ownership

```text
BUSINESS
Authentication
Organization
Authorization
Controlled Documents

SUPPORTING SEMANTIC
Audit
```

Storage/integrity, rendering/viewers, Search, async execution, Historical Migration tooling and backup/restore are mechanisms/projections/cutover/operations, not Launch semantic owners.

Known Launch+/Future capabilities remain an explicit architectural horizon in `launch-v1-ownership-topology.md`. They receive stable attachment seams but no dormant implementation.

## Stable cross-cutting references

These remain valid where they do not conflict with the accepted Product Contract, adjudicated GCR, approved Launch topology or active T1→T7 technical authority:

- [backend-api-structure.md](backend-api-structure.md) — backend route/OpenAPI structure.
- [api-contract.md](api-contract.md) — contract-first policy and generated-surface rules.
- [api-design-system.md](api-design-system.md) — API behavior standards.
- [frontend-structure.md](frontend-structure.md) — frontend structure and boundary rules.
- [tenant-context.md](tenant-context.md) — current/runtime trust-boundary reference; target tenancy mechanics must survive the current single-company authority.
- [trusted-proxy.md](trusted-proxy.md) — network trust boundary.
- [rate-limiting.md](rate-limiting.md) — request throttling architecture.
- [tech-stack.md](tech-stack.md), [deployment.md](deployment.md) — supporting runtime/operations references.

## Legacy/current-state architecture references

The following describe prior target programs or the current implementation and **must not be used as current Launch product/domain authority**:

- [backend-target-architecture.md](backend-target-architecture.md) — prior normative backend target.
- [backend-blueprint.md](backend-blueprint.md) — current/prior-state composition evidence.
- [system-overview.md](system-overview.md) — current-state overview.
- [data-model.md](data-model.md) — current/legacy data-model reference; target data model is re-derived only after the new technical design settles.
- [../backend/index.md](../backend/index.md) — forensic/current-state backend atlas; evidence, not target design.

When pages disagree on Launch capability, semantic ownership or current routing, the Product Contract + adjudicated GCR + Launch ownership topology + rebaselined R10 technical authority + current handoff control the target.
