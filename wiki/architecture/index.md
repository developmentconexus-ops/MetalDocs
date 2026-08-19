# Architecture

> **Last verified:** 2026-08-18  
> **Scope:** Durable architecture truth and active target-design routing.

## Durable authority — read first

- `launch-v1-product-contract.md` — Product Contract REV001.
- `whole-product-alignment-review.md` — Whole-Product GCR A1–A10.
- `launch-v1-ownership-topology.md` — operator-approved 4+1 ownership.
- `r10-t1-semantic-state-invariants.md` — T1 operator-ratified.
- `r10-t2-governance-effectivity-transactions.md` — T2 operator-ratified.
- `r10-t3-authorization-audit-enforcement.md` — T3 operator-ratified.
- `r10-t3-d4-responsible-owner-eligibility-amendment.md` — bounded T3 D4 authority.
- `r10-t4-exact-content-storage-integrity-restore.md` — T4 operator-ratified.
- `r10-t5-durable-async-search-external-effects.md` — T5 operator-ratified.
- `r10-t6-canonical-api-frontend-journeys.md` — **T6 operator-ratified durable authority.**
- `rebaseline-decision-registry.md` — prior-decision disposition baseline.
- `rebaseline-decision-registry-d4-amendment.md` — bounded D4 Registry reconciliation.
- `rebaseline-decision-registry-t6-amendment.md` — **T6 closure Registry reconciliation.**
- `r10-technical-architecture.md` — current stage router.
- `../references/current-agent-handoff.md` — fresh-session recovery point.

## Current gate

```text
Product Contract / GCR / ownership      APPROVED
T1                                      CLOSED / OPERATOR-RATIFIED
T2                                      CLOSED / OPERATOR-RATIFIED
T3                                      CLOSED / OPERATOR-RATIFIED + D4 amendment
T4                                      CLOSED / OPERATOR-RATIFIED
T5                                      CLOSED / OPERATOR-RATIFIED
T6                                      CLOSED / OPERATOR-RATIFIED / PROMOTED
Decision Registry                       CURRENT + D4 + T6 amendments
T7                                      ACTIVE / EVIDENCE CENSUS NEXT
implementation                          BLOCKED
```

## Active T7 staging

`../../docs/superpowers/analysis/2026-08-18-r10-t7-historical-migration-cutover-bootstrap.md`

T7 owns only:

```text
actual source evidence census
CURRENT_STATE / FULL_HISTORY or smaller real migration-mode set
imported target-owned fact shapes
ordinal/content/governance provenance
plan / dry-run / idempotency / reconciliation
semantic-unit atomicity
cutover / readiness / rollback / deletion map
restore/erasure/post-snapshot security-teardown reconciliation choreography when required
```

T7 is not a generic Interchange platform and may not weaken the target architecture for migration convenience.

## Final R10 gate after T7

```text
Integrated Whole-R10 Global Coherence Review
→ cold independent final review
→ operator final ratification
→ implementation spec/plan
→ code
```

No implementation plan or product code is authorized.