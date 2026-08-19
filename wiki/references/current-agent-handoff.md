# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **T1→T5 OPERATOR-RATIFIED; C1→C8 + D1→D4 CLOSED; T6 PLATFORM SUMMARY REV2 RATIFICATION NEXT; T7 NOT OPEN**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED — architecture/design only

## Fresh-session route

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/launch-v1-product-contract.md`
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t3-d4-responsible-owner-eligibility-amendment.md`
11. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
12. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
13. `wiki/architecture/rebaseline-decision-registry.md`
14. `wiki/architecture/rebaseline-decision-registry-d4-amendment.md`
15. `wiki/architecture/r10-technical-architecture.md`
16. `docs/superpowers/analysis/2026-08-18-r10-t6-platform-facing-summary-rev2.md` — **CURRENT OPERATOR RATIFICATION TARGET**
17. `docs/superpowers/analysis/2026-08-18-r10-t6-d1-d4-exact-delta-review.md` — final delta evidence
18. older T6 candidate/review files only for provenance when needed

## Current checkpoint

```text
Product Contract                         REV001 / OPERATOR-APPROVED
Whole-Product GCR / ownership            CLOSED / APPROVED
T1                                       CLOSED / OPERATOR-RATIFIED
T2                                       CLOSED / OPERATOR-RATIFIED
T3                                       CLOSED / OPERATOR-RATIFIED + D4 amendment
T4                                       CLOSED / OPERATOR-RATIFIED
T5                                       CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + D4 amendment
T6 material core                         OPERATOR-APPROVED / PRESERVED
C1→C8 + L1→L5                           CLOSED
D1→D4                                   CLOSED / OPERATOR-APPROVED
exact D1→D4 delta                        APPROVE / new material findings 0
Platform Summary REV2                    RATIFICATION NEXT
T6 durable authority                     NOT YET
T7                                       NOT OPEN
implementation                           BLOCKED
```

## Final T6 corrections now incorporated

```text
C1 derived lens-scoped status; no Document.currentStatus
C2 semantic fact + idempotency replay proof commit atomically
C3 /api/v1 != /auth integration != operations surface
C4 complete Launch lifecycle journeys in implementation-facing summary
C5 external source copy + commit-time current-EFFECTIVE revalidation
C6 If-Match on authority-bearing singleton current resources
C7 bounded template-admin metadata without content/history leakage
C8 Company-unique normalized DocumentType.code + Area.code
D1 current AuthZ before replay disclosure; no historical command re-execution
D2 GroupMembership current list under access.manage
D3 purpose-built least-privilege document-creation/options projection
D4 eligible responsible target = existing ENABLED User in same Company; relation grants no permission
```

Everything else remains frozen.

## Exact next step

Operator reviews and explicitly ratifies:

`docs/superpowers/analysis/2026-08-18-r10-t6-platform-facing-summary-rev2.md`

If ratified:

```text
promote T6 durable authority to wiki/
→ reconcile T6 closure in Registry authority chain
→ update router/handoff/index/PR
→ remove completed T6 staging
→ only then open T7
```

Do not open T7 or write implementation plan/code before T6 durable promotion.
