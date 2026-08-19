# `docs/superpowers` — Active Design Staging Only

> **Status:** **T6 PLATFORM SUMMARY REV2 OPERATOR RATIFICATION NEXT; T7 NOT OPEN; IMPLEMENTATION BLOCKED.**

Durable accepted truth belongs in `wiki/`. Active not-yet-promoted analysis/review belongs here; completed staging is removed and Git history is the archive.

## Durable authority

```text
Product Contract REV001
→ Whole-Product GCR A1–A10
→ 4+1 ownership topology
→ T1
→ T2
→ T3
→ T3 D4 bounded amendment
→ T4
→ T5
→ Decision Registry
→ Registry D4 amendment
→ active router
```

## Active T6 staging

Decision/review provenance remains available while T6 is open. Current gate artifacts are:

- `analysis/2026-08-18-r10-t6-c1-c8-operator-adjudication.md` — C1→C8 approved.
- `analysis/2026-08-18-r10-t6-d1-d4-operator-adjudication.md` — D1→D4 approved.
- `analysis/2026-08-18-r10-t6-platform-facing-summary-rev2.md` — **CURRENT OPERATOR RATIFICATION TARGET.**
- `analysis/2026-08-18-r10-t6-d1-d4-exact-delta-review.md` — final bounded delta, `APPROVE`.

Earlier T6 candidate/review artifacts remain provenance only until T6 closes.

## Final T6 review result

```text
C1→C8 = CLOSED
L1→L5 = CLOSED
D1→D4 = CLOSED
NEW MATERIAL FINDINGS = 0
DISAGREEMENT SET = EMPTY
DELTA VERDICT = APPROVE
```

## Current path

```text
T1→T5                       CLOSED / OPERATOR-RATIFIED
T3 D4 amendment             OPERATOR-RATIFIED
Decision Registry D4        RECONCILED BY BOUNDED AMENDMENT
T6 material/corrections     OPERATOR-APPROVED
T6 Platform Summary REV2    OPERATOR RATIFICATION NEXT
T6 durable authority        NOT YET
T7                          NOT OPEN
implementation              BLOCKED

→ operator ratifies Platform Summary REV2
→ durable T6 promotion + T6 closure reconciliation
→ staging cleanup
→ only then T7
→ Whole-R10 GCR
→ final cold review
→ final operator ratification
→ implementation spec/plan
→ code
```

No implementation plan or product code is authorized while these gates remain open.
