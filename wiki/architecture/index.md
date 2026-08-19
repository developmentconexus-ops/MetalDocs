# Architecture

> **Last verified:** 2026-08-18
> **Scope:** Durable system architecture truth and active target-design routing.

## Active target design — read first

- **[launch-v1-product-contract.md](launch-v1-product-contract.md)** — Launch V1 product authority, **REV001**.
- **[whole-product-alignment-review.md](whole-product-alignment-review.md)** — operator-adjudicated Whole-Product GCR A1–A10.
- **[launch-v1-ownership-topology.md](launch-v1-ownership-topology.md)** — operator-approved 4+1 semantic ownership.
- **[r10-t1-semantic-state-invariants.md](r10-t1-semantic-state-invariants.md)** — operator-ratified T1 authority.
- **[r10-t2-governance-effectivity-transactions.md](r10-t2-governance-effectivity-transactions.md)** — operator-ratified T2 authority.
- **[r10-t3-authorization-audit-enforcement.md](r10-t3-authorization-audit-enforcement.md)** — operator-ratified T3 authority.
- **[r10-t4-exact-content-storage-integrity-restore.md](r10-t4-exact-content-storage-integrity-restore.md)** — operator-ratified T4 authority.
- **[r10-t5-durable-async-search-external-effects.md](r10-t5-durable-async-search-external-effects.md)** — operator-ratified T5 authority.
- **[rebaseline-decision-registry.md](rebaseline-decision-registry.md)** — current operator-ratified cross-stage disposition baseline.
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — current router; **T6 PRE-RATIFICATION CORRECTION GATE ACTIVE**.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — exact fresh-session recovery point.

## Current gate

```text
Product Contract                                      REV001 / OPERATOR-APPROVED
T1 Semantic State & Invariants                       CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                CLOSED / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore       CLOSED / OPERATOR-RATIFIED
T5 Durable Async, Search & External Effects         CLOSED / OPERATOR-RATIFIED
Decision Registry                                    CURRENT / RECONCILED / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint                  CLOSED / OPERATOR-APPROVED
T6 material core                                    OPERATOR-APPROVED / PRESERVED
T6 pre-ratification GCR                              COMPLETE
T6 corrections C1→C8                                OPERATOR ADJUDICATION NEXT
T6 platform-facing summary                          RATIFICATION HELD
T6 durable authority                                NOT YET
T7 Historical Migration & Cutover                   NOT OPEN
implementation                                       BLOCKED
```

## Active T6 review gate

- `../../docs/superpowers/analysis/2026-08-18-r10-t6-pre-ratification-global-coherence-review.md` — **current operator correction target**.
- `../../docs/superpowers/analysis/2026-08-18-r10-t6-platform-facing-summary.md` — summary remains staging until corrections are adjudicated/incorporated.

Review result:

```text
core T1→T5 + 4+1 coherence = PASS
T6 direction               = PASS
formal T1→T5 reopen        = NONE
summary ratification       = HOLD
```

The review found bounded T6 contract/concurrency/authority corrections, not a need to redesign the platform core. Current runtime remains evidence only and receives no target entitlement from sunk cost.

## Exact next gate

```text
operator adjudicates C1→C8
→ corrected platform-facing T6 summary
→ bounded coherence delta
→ explicit operator summary ratification
→ durable T6 promotion + Decision Registry reconciliation + staging cleanup
→ only then T7
```

No implementation plan/product code is authorized.