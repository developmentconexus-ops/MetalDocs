# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs rebaselined R10 technical design.  
> **Current gate:** **T6 PRE-RATIFICATION GCR FOUND BOUNDED CORRECTIONS; SUMMARY RATIFICATION HELD; T7 NOT OPEN; IMPLEMENTATION BLOCKED.**

Durable accepted truth belongs in `wiki/`. Active not-yet-promoted design/review evidence belongs here. Completed staging is removed and Git history is the archive.

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
- `analysis/2026-08-18-r10-t6-operator-material-adjudication.md`
- `analysis/2026-08-18-r10-t6-platform-facing-summary.md` — summary under review, **not ratifiable yet**.
- `analysis/2026-08-18-r10-t6-pre-ratification-global-coherence-review.md` — **ACTIVE / C1→C8 OPERATOR ADJUDICATION NEXT.**

## Review outcome

```text
T1→T5 / 4+1 core coherence      PASS
T6 Global-Maximum direction     PASS
formal T1→T5 reopen             NONE
T6 platform summary             HELD for bounded corrections
```

Required C1→C8 and low refinements are owned by the review artifact. Everything not named there remains frozen.

## Current path

```text
T1→T5                       CLOSED / OPERATOR-RATIFIED
Post-T5 Fable               CLOSED / OPERATOR-APPROVED
Decision Registry           CURRENT / RECONCILED
T6 material core            OPERATOR-APPROVED / PRESERVED
T6 pre-ratification GCR     COMPLETE
T6 corrections C1→C8       OPERATOR ADJUDICATION NEXT
T6 summary ratification     HELD
T7                          NOT OPEN
implementation              BLOCKED

→ operator adjudicates C1→C8
→ corrected T6 summary + bounded coherence delta
→ explicit operator summary ratification
→ durable T6 promotion + Registry update + staging cleanup
→ only then T7
→ Whole-R10 GCR
→ final cold review
→ final operator ratification
→ implementation spec/plan
→ code
```

No implementation plan/product code is authorized while these gates remain open.