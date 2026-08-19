# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **PRODUCT CONTRACT REV001 + T1→T5 OPERATOR-RATIFIED; T6 CORE COHERENT; PRE-RATIFICATION GCR FOUND BOUNDED T6 CORRECTIONS; SUMMARY RATIFICATION HELD; T7 NOT OPEN**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — architecture/design only**

## Fresh-session route

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/launch-v1-product-contract.md` — REV001
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. `wiki/architecture/r10-technical-architecture.md`
14. `docs/superpowers/analysis/2026-08-18-r10-t6-operator-material-adjudication.md`
15. `docs/superpowers/analysis/2026-08-18-r10-t6-pre-ratification-global-coherence-review.md` — **ACTIVE CORRECTION GATE**
16. T6 summary/candidate/evidence staging only for exact correction provenance
17. current implementation only as claim-specific evidence

## Current checkpoint

```text
Product Contract                         = REV001 / OPERATOR-APPROVED
Whole-Product GCR A1–A10                 = CLOSED / OPERATOR-APPROVED
Launch ownership topology                = CLOSED / OPERATOR-APPROVED / 4+1
T1 Semantic State & Invariants           = CLOSED / OPERATOR-RATIFIED
T2 Governance/Effectivity/Tx             = CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit                 = CLOSED / OPERATOR-RATIFIED
T4 Exact Content/Storage/Restore         = CLOSED / OPERATOR-RATIFIED
T5 Durable Async/Search/Effects          = CLOSED / OPERATOR-RATIFIED
Decision Registry                        = CURRENT / RECONCILED / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint      = CLOSED / OPERATOR-APPROVED
T6 material core                         = OPERATOR-APPROVED / PRESERVED
T6 pre-ratification GCR                  = COMPLETE
Formal T1→T5 reopen                      = NONE
T6 corrections C1→C8                    = OPERATOR ADJUDICATION NEXT
T6 platform-facing summary               = RATIFICATION HELD
T6 durable authority                     = NOT YET
T7 Historical Migration / Cutover        = NOT OPEN
implementation                           = BLOCKED
```

## Correction set

```text
C1 lens-scoped derived status filtering; never Document.currentStatus
C2 atomic semantic-commit + Idempotency-Key replay result; remove baseline public IN_PROGRESS state
C3 /api/v1 application contract != /auth browser integration != operations surface
C4 complete integrated lifecycle journeys in the T6 platform summary
C5 next-Revision external source copy + commit-time current-EFFECTIVE revalidation
C6 strong If-Match for provider-binding / responsible-owner / template-role singleton replacements
C7 bounded template admin metadata under template_use.manage; no implicit content/history read
C8 Company-unique normalized DocumentType.code + Area.code
```

Low refinements and future-seam pass are recorded in the review. Everything else remains frozen.

## Exact next step

Operator adjudicates C1→C8 from:

`docs/superpowers/analysis/2026-08-18-r10-t6-pre-ratification-global-coherence-review.md`

If accepted:

```text
correct T6 summary/material record
→ bounded coherence delta
→ explicit operator platform-summary ratification
→ durable T6 promotion + Decision Registry reconciliation + staging cleanup
→ only then T7
```

Do **not** ratify/promote the current summary, open T7, or write implementation plan/code before this correction gate closes.