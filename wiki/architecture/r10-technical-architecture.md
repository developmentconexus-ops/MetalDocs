# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **PRODUCT CONTRACT REV001 + T1→T5 OPERATOR-RATIFIED; T6 MATERIAL CORE PRESERVED; PRE-RATIFICATION GCR FOUND BOUNDED T6 CORRECTIONS; SUMMARY RATIFICATION HELD; T7 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Revision convention:** `REV000` initial issuance / `REV001` first revision  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation gate:** **CLOSED — architecture/design only**

This file owns current technical-stage status and exact next action.

## 1. Binding authority chain

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md` — REV001
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. this router
14. `docs/superpowers/analysis/2026-08-18-r10-t6-operator-material-adjudication.md`
15. `docs/superpowers/analysis/2026-08-18-r10-t6-pre-ratification-global-coherence-review.md` — **ACTIVE REVIEW / CORRECTION GATE**
16. T6 platform summary/candidate/evidence staging only for the bounded correction set
17. current implementation only as claim-specific evidence

Current implementation has **no compatibility entitlement**. Structural Inversion controls T6.

## 2. Binding Method laws

```text
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
Structural Inversion
```

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed.**

## 3. Technical descent

```text
Product Contract                                        REV001 / OPERATOR-APPROVED
T1 — Semantic State & Invariants                       CLOSED / OPERATOR-RATIFIED
T2 — Governance, Effectivity & Lifecycle Transactions CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit Enforcement                CLOSED / OPERATOR-RATIFIED
T4 — Exact Content, Storage Integrity & Restore       CLOSED / OPERATOR-RATIFIED
T5 — Durable Async, Search & External Effects         CLOSED / OPERATOR-RATIFIED
Decision Registry                                      CURRENT / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint                   CLOSED / OPERATOR-APPROVED
T6 material core                                      OPERATOR-APPROVED / PRESERVED
T6 pre-ratification GCR                               COMPLETE / BOUNDED CORRECTIONS FOUND
T6 correction set C1→C8                               OPERATOR ADJUDICATION NEXT
T6 platform-facing summary                            RATIFICATION HELD
T6 durable authority                                  NOT YET
T7 — Historical Migration & Cutover                   NOT OPEN
implementation                                         BLOCKED
```

## 4. Pre-ratification Global Coherence Review

Authority/review:

`docs/superpowers/analysis/2026-08-18-r10-t6-pre-ratification-global-coherence-review.md`

Verdict:

```text
core T1→T5 / 4+1 coherence     PASS
T6 Global-Maximum direction    PASS
formal T1→T5 reopen            NONE
summary ready for ratification NO
```

Required bounded T6 corrections:

```text
C1 status discovery is lens-scoped/derived; never persisted Document.currentStatus
C2 Idempotency-Key replay result commits atomically with semantic transition; no baseline public IN_PROGRESS state
C3 distinguish /api/v1 application contract from /auth integration and operations surfaces
C4 restore complete Launch lifecycle journeys to platform-facing summary
C5 next-Revision source copy revalidates current EFFECTIVE source after external copy and before commit
C6 If-Match on provider-binding / responsible-owner / template-role current singleton resources
C7 template admin uses bounded template_use.manage metadata surface without implicit document content/history access
C8 normalized DocumentType.code + Area.code are Company-unique numbering inputs
```

Low refinements L1→L5 remain in the review and require no T1→T5 reopen.

Everything else in the operator-approved T6 material core remains frozen.

## 5. Current gate

```text
operator adjudicates C1→C8
→ incorporate accepted corrections into T6 summary/material record
→ bounded coherence delta against Product Contract + T1→T5
→ explicit operator platform-summary ratification
→ promote durable T6 authority to wiki/
→ reconcile Decision Registry
→ remove completed T6 staging
→ only then open T7
```

A previous material-decision approval does not override this later evidence-driven correction gate.

## 6. Final gate after T7

```text
Integrated Whole-R10 Global Coherence Review
→ cold independent final review
→ operator final ratification
→ implementation spec/plan
→ code
```

**Implementation remains BLOCKED.**