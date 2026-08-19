# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **PRODUCT CONTRACT REV001 + T1→T5 OPERATOR-RATIFIED; POST-T5 FABLE CHECKPOINT CLOSED; T6 FINAL GLOBAL-MAXIMUM ADJUDICATION READY / OPERATOR MATERIAL ADJUDICATION NEXT; T7 NOT OPEN**  
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
14. `docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-bootstrap.md`
15. `docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-candidate.md`
16. `docs/superpowers/analysis/2026-08-18-r10-t6-external-evidence-docket.md`
17. `docs/superpowers/analysis/2026-08-18-r10-t6-global-maximum-adjudication-packet.md`
18. `docs/superpowers/analysis/2026-08-18-r10-t6-final-adjudication-refinements.md` — **FINAL OPERATOR DECISION PRECEDENCE FOR FR-1..FR-4**
19. current implementation only when claim-specific evidence is needed

Completed Fable staging is removed; Git history is the archive.

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
T6 evidence/inversion pass               = COMPLETE ENOUGH FOR ADJUDICATION
T6 base candidate                        = STAGED / NON-AUTHORITATIVE
T6 corrected adjudication packet         = STAGED / NON-AUTHORITATIVE
T6 final refinements FR-1..FR-4          = STAGED / NON-AUTHORITATIVE
operator material adjudication           = NEXT
T7 Historical Migration / Cutover        = NOT OPEN
implementation                           = BLOCKED
```

## Structural Inversion result

Current API/frontend is current-state evidence only and carries major superseded concepts. T6 has no obligation to retain routes/modules/DTOs/screens by migration cost or sunk cost.

Target direction:

```text
semantic public API instead of legacy module API
semantic-lens frontend instead of old navigation ontology
exact immutable Submission review instead of reviewer WorkingContent mutation
current-effective Library instead of polymorphic document screen
canonical Search instead of mandatory Search infrastructure
provider-neutral content/editor mechanisms behind T4/OCC
```

## Operator adjudication precedence

```text
T6 base candidate
→ corrected Global-Maximum adjudication packet
→ final refinements FR-1..FR-4
```

FR-1..FR-4 final candidate deltas:

```text
FR-1 User eligibility = singleton GET/PUT current resource; DISABLED transition executes T3 offboarding; ENABLED never restores grants.
FR-2 Governance Step Decision = singleton immutable GET/PUT resource; no Idempotency-Key replay row.
FR-3 DRAFT mutation = PATCH + strong If-Match over one generation for title + source; stale = 412.
FR-4 Idempotency-Key only for truly non-idempotent semantic POSTs; replay retention bounded; 24h only first implementation-default candidate, not architecture invariant.
```

Everything else follows the corrected packet and remains non-authoritative until operator ratification.

## Exact next step

```text
operator adjudicates final T6 material slate
→ revise only rejected/refined decisions
→ platform-facing T6 summary
→ explicit operator summary ratification
→ promote T6 durable authority
→ reconcile Decision Registry
→ remove staging
→ only then T7
```

No final SQL/index/package/process topology, Historical Migration execution plan, implementation plan or product code is authorized.
