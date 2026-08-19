# Architecture

> **Last verified:** 2026-08-18  
> **Scope:** Durable architecture truth and active target-design routing.

## Durable authority

- `launch-v1-product-contract.md` — REV001.
- `whole-product-alignment-review.md` — A1–A10.
- `launch-v1-ownership-topology.md` — 4+1 ownership.
- `r10-t1-semantic-state-invariants.md` — T1.
- `r10-t2-governance-effectivity-transactions.md` — T2.
- `r10-t3-authorization-audit-enforcement.md` — T3; only the D4 phrase is currently evidence-reopened for operator adjudication.
- `r10-t4-exact-content-storage-integrity-restore.md` — T4.
- `r10-t5-durable-async-search-external-effects.md` — T5.
- `rebaseline-decision-registry.md` — current disposition baseline.
- `r10-technical-architecture.md` — current stage router.
- `../references/current-agent-handoff.md` — fresh-session recovery point.

## Current gate

```text
Product Contract / GCR / ownership      APPROVED
T1                                      CLOSED
T2                                      CLOSED
T3                                      CLOSED except D4 precision question
T4                                      CLOSED
T5                                      CLOSED
T6 material core                        OPERATOR-APPROVED / PRESERVED
C1→C8 + L1→L5                           CLOSED
T6 bounded delta D1→D4                  OPERATOR ADJUDICATION NEXT
T6 platform summary                     RATIFICATION HELD
T7                                      NOT OPEN
implementation                          BLOCKED
```

Current delta review:

`../../docs/superpowers/analysis/2026-08-18-r10-t6-bounded-coherence-delta-review.md`

New bounded findings:

```text
D1 current access authorization before Idempotency-Key replay disclosure
D2 cursor GroupMembership read surface under access.manage
D3 purpose-built least-privilege document-creation/options reference projection
D4 define responsible-owner target eligibility = current enabled Company User; no implicit access grant
```

Minimal reopen:

```text
T3 §9 phrase only
T6 D1→D3 only
```

Everything else remains frozen.

## Next

```text
operator adjudicates D1→D4
→ exact bounded corrections
→ exact D1→D4 delta review
→ operator T6 platform-summary ratification
→ durable T6 promotion + Registry update + staging cleanup
→ only then T7
```

No implementation plan or product code is authorized.