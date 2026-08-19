# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T5 OPERATOR-RATIFIED; C1→C8 CLOSED; T6 BOUNDED DELTA FOUND D1→D4; SUMMARY RATIFICATION HELD; T7 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

This file owns current technical-stage status and exact next action.

## Binding authority chain

```text
AGENTS.md
→ DevelopmentConexus Engineering Method v1.0.0
→ wiki/references/current-agent-handoff.md
→ Product Contract REV001
→ Whole-Product GCR A1–A10
→ 4+1 ownership topology
→ T1
→ T2
→ T3
→ T4
→ T5
→ Decision Registry
→ this router
→ active T6 review/adjudication staging only
```

Current implementation has no compatibility entitlement. Structural Inversion controls.

## Current descent

```text
Product Contract                         REV001 / OPERATOR-APPROVED
T1                                       CLOSED / OPERATOR-RATIFIED
T2                                       CLOSED / OPERATOR-RATIFIED
T3                                       CLOSED / OPERATOR-RATIFIED, except D4 precision question now evidence-open
T4                                       CLOSED / OPERATOR-RATIFIED
T5                                       CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT / OPERATOR-RATIFIED
Post-T5 Fable checkpoint                 CLOSED / OPERATOR-APPROVED
T6 material core                         OPERATOR-APPROVED / PRESERVED
C1→C8 + L1→L5                           CLOSED / INCORPORATED
T6 bounded coherence delta               COMPLETE / NEW D1→D4 FOUND
T6 summary                               RATIFICATION HELD
T7                                       NOT OPEN
implementation                           BLOCKED
```

## Bounded delta authority/evidence

`docs/superpowers/analysis/2026-08-18-r10-t6-bounded-coherence-delta-review.md`

Delta verdict:

```text
C1→C8 = CLOSED
L1→L5 = CLOSED
NEW MATERIAL FINDINGS = 4
DELTA VERDICT = MATERIAL PRECISION DELTA
```

D1→D4:

```text
D1  current T3 access authorization must be rechecked before an Idempotency-Key replay response is disclosed; completed mutation predicates are not re-executed
D2  Access Admin needs cursor-paginated GET GroupMembership read surface under access.manage
D3  create/owner journeys need a purpose-built least-privilege document-creation/options projection instead of reusing admin/PII reads
D4  T3 must define `eligible target User` for responsible owner; recommended = current ENABLED User in same Company, with no implicit Role/Permission grant
```

Minimal reopen set:

```text
T1 = EMPTY
T2 = EMPTY
T3 = §9 responsible-owner target eligibility phrase only
T4 = EMPTY
T5 = EMPTY
T6 = D1→D3 contract/read-surface precision
```

Everything else remains frozen.

## Current gate

```text
operator adjudicates D1→D4
→ if accepted: apply D1→D3 to T6 summary
→ apply D4 bounded T3 clarification + Registry reconciliation
→ rerun exact bounded delta only over D1→D4
→ if clean: operator platform-summary ratification
→ durable T6 promotion
→ staging cleanup
→ only then T7
```

A previous T6 material approval does not override later evidence. Implementation remains **BLOCKED**.