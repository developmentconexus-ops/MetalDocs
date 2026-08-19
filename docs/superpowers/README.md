# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs rebaselined R10 technical design.  
> **Current gate:** **T6 MATERIAL DECISIONS OPERATOR-APPROVED; PLATFORM-FACING SUMMARY RATIFICATION NEXT; T7 NOT OPEN; IMPLEMENTATION BLOCKED.**

Durable accepted truth belongs in `wiki/`. Active not-yet-promoted design/evidence belongs here. Completed staging is removed and Git history is the archive.

## Durable authority

```text
wiki/architecture/launch-v1-product-contract.md          REV001
→ wiki/architecture/whole-product-alignment-review.md
→ wiki/architecture/launch-v1-ownership-topology.md
→ wiki/architecture/r10-t1-semantic-state-invariants.md
→ wiki/architecture/r10-t2-governance-effectivity-transactions.md
→ wiki/architecture/r10-t3-authorization-audit-enforcement.md
→ wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md
→ wiki/architecture/r10-t5-durable-async-search-external-effects.md
→ wiki/architecture/rebaseline-decision-registry.md
→ wiki/architecture/r10-technical-architecture.md
```

## Active T6 staging

Decision provenance:

- `analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-bootstrap.md`
- `analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-candidate.md`
- `analysis/2026-08-18-r10-t6-external-evidence-docket.md`
- `analysis/2026-08-18-r10-t6-global-maximum-adjudication-packet.md`
- `analysis/2026-08-18-r10-t6-final-adjudication-refinements.md`

Current gate artifacts:

- `analysis/2026-08-18-r10-t6-operator-material-adjudication.md` — **MATERIAL DECISIONS OPERATOR-APPROVED.**
- `analysis/2026-08-18-r10-t6-platform-facing-summary.md` — **OPERATOR SUMMARY RATIFICATION NEXT.**

The material adjudication does not yet make T6 durable authority. Promotion waits for explicit summary ratification.

## T6 greenfield law

```text
Product Contract REV001 + T1→T5
→ Structural Inversion
→ smallest sustainable API/UX
```

Current routes/modules/screens/DTOs are evidence only. Do not preserve a legacy surface because migration is easier.

## Current path

```text
T1→T5                        CLOSED / OPERATOR-RATIFIED
Post-T5 Fable                CLOSED / OPERATOR-APPROVED
Decision Registry            CURRENT / RECONCILED
T6 material decisions        OPERATOR-APPROVED
T6 platform-facing summary   STAGED / RATIFICATION NEXT
T6 durable authority         NOT YET
T7                           NOT OPEN
implementation               BLOCKED

→ operator ratifies platform-facing T6 summary
→ durable T6 promotion + Registry update + staging cleanup
→ only then T7
→ Whole-R10 GCR
→ final cold review
→ final operator ratification
→ implementation spec/plan
→ code
```

No implementation plan/product code is authorized while these gates remain open.