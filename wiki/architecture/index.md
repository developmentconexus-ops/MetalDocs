# Architecture

> **Last verified:** 2026-08-18  
> **Scope:** Durable architecture truth and active target-design routing.

## Durable authority

- `launch-v1-product-contract.md` — REV001.
- `whole-product-alignment-review.md` — A1–A10.
- `launch-v1-ownership-topology.md` — 4+1 ownership.
- `r10-t1-semantic-state-invariants.md` — T1.
- `r10-t2-governance-effectivity-transactions.md` — T2.
- `r10-t3-authorization-audit-enforcement.md` — T3.
- `r10-t3-d4-responsible-owner-eligibility-amendment.md` — operator-ratified bounded T3 precision for responsible-owner target eligibility.
- `r10-t4-exact-content-storage-integrity-restore.md` — T4.
- `r10-t5-durable-async-search-external-effects.md` — T5.
- `rebaseline-decision-registry.md` — current prior-decision disposition baseline.
- `rebaseline-decision-registry-d4-amendment.md` — bounded D4 Registry reconciliation.
- `r10-technical-architecture.md` — current stage router.
- `../references/current-agent-handoff.md` — fresh-session recovery point.

## Current gate

```text
Product Contract / GCR / ownership      APPROVED
T1                                      CLOSED
T2                                      CLOSED
T3                                      CLOSED + D4 bounded amendment
T4                                      CLOSED
T5                                      CLOSED
T6 material core                        OPERATOR-APPROVED / PRESERVED
C1→C8 + L1→L5                           CLOSED
D1→D4                                   CLOSED / OPERATOR-APPROVED
exact D1→D4 delta                       APPROVE / new material findings 0
T6 Platform Summary REV2                OPERATOR RATIFICATION NEXT
T6 durable authority                    NOT YET
T7                                      NOT OPEN
implementation                          BLOCKED
```

## Current T6 ratification target

`../../docs/superpowers/analysis/2026-08-18-r10-t6-platform-facing-summary-rev2.md`

Final bounded delta evidence:

`../../docs/superpowers/analysis/2026-08-18-r10-t6-d1-d4-exact-delta-review.md`

Final delta result:

```text
D1 = CLOSED
D2 = CLOSED
D3 = CLOSED
D4 = CLOSED
NEW MATERIAL FINDINGS = 0
DISAGREEMENT SET = EMPTY
DELTA VERDICT = APPROVE
```

## Next

```text
operator explicitly ratifies Platform Summary REV2
→ durable T6 promotion
→ reconcile T6 closure in Registry authority chain
→ staging cleanup
→ only then T7
```

No implementation plan or product code is authorized.
