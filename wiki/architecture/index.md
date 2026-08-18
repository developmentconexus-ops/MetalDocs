# Architecture

> **Last verified:** 2026-08-18  
> **Scope:** Durable system architecture truth and active target-design routing.

## Active target design — read first

- **[launch-v1-product-contract.md](launch-v1-product-contract.md)** — accepted Launch V1 product authority; `REV000 = initial issuance`, `REV001 = first revision`.
- **[whole-product-alignment-review.md](whole-product-alignment-review.md)** — operator-adjudicated Whole-Product GCR A1–A10.
- **[launch-v1-ownership-topology.md](launch-v1-ownership-topology.md)** — operator-approved 4+1 semantic ownership + future-evolution law.
- **[r10-t1-semantic-state-invariants.md](r10-t1-semantic-state-invariants.md)** — operator-ratified T1 authority.
- **[r10-t2-governance-effectivity-transactions.md](r10-t2-governance-effectivity-transactions.md)** — operator-ratified T2 authority.
- **[r10-t3-authorization-audit-enforcement.md](r10-t3-authorization-audit-enforcement.md)** — operator-ratified T3 authority.
- **[r10-t4-exact-content-storage-integrity-restore.md](r10-t4-exact-content-storage-integrity-restore.md)** — operator-ratified T4 exact-content/storage/restore authority.
- **[rebaseline-decision-registry.md](rebaseline-decision-registry.md)** — current operator-ratified disposition baseline for prior decisions.
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — active T1→T7 stage router; **T5 active**.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — exact fresh-session recovery point/current gate.
- [../standards/root-cause-global-maximum-method.md](../standards/root-cause-global-maximum-method.md) — binding DevelopmentConexus Engineering Method v1.0.0 mirror.

## Current gate

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                  CLOSED / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore         CLOSED / OPERATOR-RATIFIED
Decision Registry                                      CURRENT / OPERATOR-RATIFIED
T5 Durable Async, Search & External Effects           ACTIVE / DESIGN
T6→T7                                                  NOT OPEN
implementation                                         BLOCKED
```

Every T-stage requires material-decision adjudication and operator approval of a platform-facing summary before the next stage opens. Every remaining stage begins from and updates the Decision Registry.

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

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

Managed content/storage, rendering/viewers, Search, async execution, Historical Migration tooling and backup/restore are mechanisms/projections/cutover/operations, not Launch semantic owners.

## Active staging

`docs/superpowers/analysis/2026-08-18-r10-t5-durable-async-search-external-effects-candidate.md` is the active non-authoritative T5 candidate.

## Prior redesign / technical evidence

- [cohesive-platform-redesign.md](cohesive-platform-redesign.md) — historical compatibility/evidence entrypoint; old 8+3 topology/stage routing superseded.
- [launch-v1-scope-rebaseline.md](launch-v1-scope-rebaseline.md) — narrow Records-Governance defer overlay.
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — historical decision inventory only.
- Former B1–B6/R10-C and current implementation are usable only as evidence classified by the Decision Registry.

## Stable cross-cutting references

These remain usable only where they do not conflict with current authority:

- [backend-api-structure.md](backend-api-structure.md)
- [api-contract.md](api-contract.md)
- [api-design-system.md](api-design-system.md)
- [frontend-structure.md](frontend-structure.md)
- [trusted-proxy.md](trusted-proxy.md)
- [rate-limiting.md](rate-limiting.md)
- [tech-stack.md](tech-stack.md)
- [deployment.md](deployment.md)

When pages disagree, Product Contract + GCR + ownership topology + ratified T-stage authorities + Decision Registry + current handoff control the target.