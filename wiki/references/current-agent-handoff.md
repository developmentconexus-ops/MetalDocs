# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **PRODUCT CONTRACT REV001 + T1→T5 + DECISION REGISTRY OPERATOR-RATIFIED; POST-T5 FABLE ROUND-1 AMENDMENTS PROMOTED; DELTA REVIEW PENDING; T6 NOT OPEN**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/launch-v1-product-contract.md` — **REV001**
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. `wiki/architecture/r10-technical-architecture.md`
14. `docs/superpowers/analysis/2026-08-18-t1-t5-integrated-independent-fable-review.md` — independent review evidence
15. `docs/superpowers/analysis/2026-08-18-t1-t5-fable-author-adjudication-round1.md` — **OPERATOR-RATIFIED ROUND 1**
16. `docs/superpowers/analysis/2026-08-18-t1-t5-fable-delta-review-request.md` — **ACTIVE DELTA REVIEW REQUEST**
17. `wiki/architecture/launch-v1-scope-rebaseline.md`
18. old R3–R9.5 / old R10/current implementation only as evidence allowed by current authority/registry

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
Fable original verdict                   = APPROVE T1→T5 WITH MATERIAL FIXES
Formal T-stage reopen                    = NONE
Round-1 adjudication                     = OPERATOR-RATIFIED
Durable bounded amendments               = APPLIED
Fable delta review                       = PENDING
Post-T5 checkpoint                       = OPEN
T6 Canonical API / Frontend Journeys     = NOT OPEN
T7 Historical Migration / Cutover        = NOT OPEN
implementation                           = BLOCKED
```

## Ratified post-T5 delta

```text
M1
  materialized Search is conditional;
  if activated, serialize per-Document projection write before canonical read through rewrite/removal;
  FIFO remains unnecessary.

M2
  all restored ApplicationSessions invalid before ordinary serving;
  required known post-snapshot access teardown must be reconciled/proven before ordinary authenticated serving;
  T7 chooses smallest recovery proof mechanism; no generic per-grant journal frozen.

M3
  Search journey required;
  baseline = canonical PostgreSQL query/view over current canonical facts;
  materialized projection + search_refresh + rebuild activate only on proven derived/expensive/measured need.

L1 title = Revision-governed metadata.
L2 late renderer result for dead candidate = semantic no-op/reclaimable output.
L3 live bounded admission claim protects in-flight READY content from GC.
L4 active human-governed obsolescence request has bounded initiator/manager withdrawal.
L5 T3 provider-disable wording follows T5-L.
```

Additional guardrails:

```text
same-DB durable-intent restore coherence recorded
registry ambiguous SUPERSEDED wording tightened
T6 REOPEN explicitly includes source upload/T4 admission UX and Search materialization proof
T7 REOPEN explicitly includes post-snapshot security-teardown recovery choreography
```

Everything not named remains frozen.

## Exact next step

Operator dispatches Fable to read:

`docs/superpowers/analysis/2026-08-18-t1-t5-fable-delta-review-request.md`

Fable must write:

`docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-delta-review.md`

Close condition:

```text
DELTA VERDICT = APPROVE
DISAGREEMENT SET = EMPTY
T6 READINESS = MAY OPEN
```

If Fable disagrees, adjudicate only the exact remaining delta through GitHub. Do **not** restart T1→T5 from zero.

T6 remains NOT OPEN until the post-T5 checkpoint explicitly closes. No final SQL/index/package/process topology, public API/frontend contract, Historical Migration execution plan, implementation plan or product code is authorized.