# Architecture

> **Last verified:** 2026-08-18
> **Scope:** Durable system architecture truth and active target-design routing.

## Active target design — read first

- **[launch-v1-product-contract.md](launch-v1-product-contract.md)** — accepted Launch V1 product authority; `REV000 = initial issuance`.
- **[whole-product-alignment-review.md](whole-product-alignment-review.md)** — operator-adjudicated Whole-Product GCR A1–A10.
- **[launch-v1-ownership-topology.md](launch-v1-ownership-topology.md)** — operator-approved 4+1 semantic ownership + future-evolution law.
- **[r10-t1-semantic-state-invariants.md](r10-t1-semantic-state-invariants.md)** — operator-ratified T1 semantic-state authority.
- **[r10-t2-governance-effectivity-transactions.md](r10-t2-governance-effectivity-transactions.md)** — operator-ratified T2 transaction/governance/effectivity authority.
- **[rebaseline-decision-registry.md](rebaseline-decision-registry.md)** — **operator-ratified cross-stage disposition registry for prior decisions.**
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — active T1→T7 router; T3 is active.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — exact fresh-session recovery point.
- [../standards/root-cause-global-maximum-method.md](../standards/root-cause-global-maximum-method.md) — binding DevelopmentConexus Engineering Method v1.0.0 mirror.

## Current gate

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
Decision Registry                                      CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                  ACTIVE
T4→T7                                                  NOT OPEN
implementation                                         BLOCKED
```

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

The Decision Registry classifies prior decisions as:

```text
CURRENT
PRESERVE
REFINED
REOPEN
DEFERRED
SUPERSEDED
```

Every T3–T7 design consumes `CURRENT/PRESERVE/REFINED`, designs only its `REOPEN` set, preserves `DEFERRED` seams and rejects `SUPERSEDED` inheritance.

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

## Prior redesign / evidence

- [cohesive-platform-redesign.md](cohesive-platform-redesign.md) — prior-design compatibility/evidence entrypoint; former 8+3 topology is superseded.
- [launch-v1-scope-rebaseline.md](launch-v1-scope-rebaseline.md) — narrow Records-Governance defer overlay, subordinate to current authorities.
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — historical consolidated decision inventory only.
- Former B1–B6/R10-C files are evidence whose current disposition is controlled by [rebaseline-decision-registry.md](rebaseline-decision-registry.md).

## Stable cross-cutting references

These remain usable only where they do not conflict with the Product Contract, GCR, 4+1 topology, ratified T-stage authorities or Decision Registry:

- [backend-api-structure.md](backend-api-structure.md)
- [api-contract.md](api-contract.md)
- [api-design-system.md](api-design-system.md)
- [frontend-structure.md](frontend-structure.md)
- [tenant-context.md](tenant-context.md) — current/runtime evidence; pooled-tenant target mechanics are deferred.
- [trusted-proxy.md](trusted-proxy.md)
- [rate-limiting.md](rate-limiting.md)
- [tech-stack.md](tech-stack.md)
- [deployment.md](deployment.md)

## Legacy/current-state references

These describe prior/current implementation and are evidence only, not Launch product/domain authority:

- [backend-target-architecture.md](backend-target-architecture.md)
- [backend-blueprint.md](backend-blueprint.md)
- [system-overview.md](system-overview.md)
- [data-model.md](data-model.md)
- [../backend/index.md](../backend/index.md)

When pages disagree, the current authority chain and handoff control.