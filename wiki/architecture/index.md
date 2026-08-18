# Architecture

> **Last verified:** 2026-08-18
> **Scope:** Durable system architecture truth and active target-design routing.

## Active target design — read first

- **[launch-v1-product-contract.md](launch-v1-product-contract.md)** — **ACCEPTED Launch V1 product authority.** Includes `REV000 = initial issuance`, `REV001 = first revision`.
- **[whole-product-alignment-review.md](whole-product-alignment-review.md)** — **OPERATOR-ADJUDICATED Whole-Product GCR authority.** A1–A10 accepted.
- **[launch-v1-ownership-topology.md](launch-v1-ownership-topology.md)** — **OPERATOR-APPROVED Launch semantic ownership authority.** Launch topology is 4 business owners + Audit, with the binding future-evolution law.
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — **ACTIVE REBASELINED T1→T7 TECHNICAL-STAGE AUTHORITY.** T1 and T2 are operator-ratified; Decision Reconciliation is the active gate and T3 is paused.
- **[r10-t2-governance-effectivity-transactions.md](r10-t2-governance-effectivity-transactions.md)** — **OPERATOR-RATIFIED T2 authority.** Transaction, concurrency, governance and effectivity laws.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — exact fresh-session recovery point / current gate.
- [../standards/root-cause-global-maximum-method.md](../standards/root-cause-global-maximum-method.md) — binding DevelopmentConexus Engineering Method v1.0.0 mirror.

Active staging reconciliation:

- `../../docs/superpowers/analysis/2026-08-18-rebaseline-decision-reconciliation-candidate.md` — **NON-AUTHORITATIVE candidate** reconciling prior R3–R9.5 / old R10 decisions into CURRENT / PRESERVE / REFINED / REOPEN / DEFERRED / SUPERSEDED.

Current gate:

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
Decision Reconciliation Baseline                     ACTIVE / OPERATOR REVIEW NEXT
T3 Authorization & Audit Enforcement                  PAUSED ON RECONCILIATION
T4→T7                                                  NOT OPEN
implementation                                         BLOCKED
```

Every T-stage requires material-decision adjudication **and** operator approval of a platform-facing summary before the next T-stage may open. After reconciliation promotion, every T-stage must also consume/update the durable Decision Registry.

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

The reconciliation gate exists to prevent both sunk-cost legacy inheritance and needless redesign of already-good decisions.

## Prior redesign / technical evidence

- [cohesive-platform-redesign.md](cohesive-platform-redesign.md) — prior-design evidence compatibility entrypoint. Its former 8+3 Launch topology and old stage routing are superseded.
- [launch-v1-scope-rebaseline.md](launch-v1-scope-rebaseline.md) — narrow operator-approved Records-Governance defer overlay; subordinate to Product Contract/GCR/topology.
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — historical consolidated decision inventory; not current authority.
- Former B1–B6/R10-C material remains evidence whose survivorship is being explicitly classified by the active reconciliation candidate.

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

These remain valid where they do not conflict with the accepted Product Contract, adjudicated GCR, approved Launch topology, ratified T-stage conclusions or the eventual Decision Registry:

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

When pages disagree on Launch capability, semantic ownership or current routing, the Product Contract + adjudicated GCR + Launch ownership topology + ratified T-stage authorities + current handoff control the target. Once operator-ratified, the Decision Registry additionally controls the disposition of prior decisions.