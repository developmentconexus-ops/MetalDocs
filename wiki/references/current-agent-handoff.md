# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **T1→T5 + DECISION REGISTRY OPERATOR-RATIFIED; POST-T5 FABLE REVIEW RECEIVED; AUTHOR ROUND-1 ADJUDICATION PENDING OPERATOR RATIFICATION; T6 NOT OPEN**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/launch-v1-product-contract.md`
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. `wiki/architecture/r10-technical-architecture.md`
14. `docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-review-request.md` — review request / evidence
15. `docs/superpowers/analysis/2026-08-18-t1-t5-integrated-independent-fable-review.md` — **FABLE REVIEW RECEIVED**
16. `docs/superpowers/analysis/2026-08-18-t1-t5-fable-author-adjudication-round1.md` — **AUTHOR RESPONSE / OPERATOR RATIFICATION NEXT**
17. `wiki/architecture/launch-v1-scope-rebaseline.md`
18. old R3–R9.5 / old R10/current implementation only as evidence allowed by current authority/registry

## Current checkpoint

```text
Product Contract                         = ACTIVE / OPERATOR-APPROVED
Whole-Product GCR A1–A10                 = CLOSED / OPERATOR-APPROVED
Launch ownership topology                = CLOSED / OPERATOR-APPROVED / 4+1
T1 Semantic State & Invariants           = CLOSED / OPERATOR-RATIFIED
T2 Governance/Effectivity/Tx             = CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit                 = CLOSED / OPERATOR-RATIFIED
T4 Exact Content/Storage/Restore         = CLOSED / OPERATOR-RATIFIED
T5 Durable Async/Search/Effects          = CLOSED / OPERATOR-RATIFIED
Decision Registry                        = CURRENT / OPERATOR-RATIFIED
Fable independent review                 = RECEIVED / EVIDENCE ONLY
Fable verdict                            = APPROVE T1→T5 WITH MATERIAL FIXES
Author Round-1 adjudication              = WRITTEN / OPERATOR RATIFICATION NEXT
Durable amendments                       = NOT YET APPLIED
Post-T5 checkpoint                       = OPEN
T6 Canonical API / Frontend Journeys     = NOT OPEN
T7 Historical Migration / Cutover        = NOT OPEN
implementation                           = BLOCKED
```

## Fable verdict

Commit:

`bdef5fc3c4004aa3ab4deefc9e8373dd3efcf856`

```text
BLOCKER = 0
MAJOR   = 3
LOW     = 5
NOTE    = 3
formal T-stage reopen = NONE
```

Major findings:

```text
M1 Search projection concurrent overlap can end stale despite latest-state reads.
M2 restore can resurrect revoked sessions/access teardown.
M3 materialized Search/search_refresh lack a named current consumer proving materialization.
```

## Author Round-1 recommendation — NON-AUTHORITATIVE until operator ratification

Detailed file:

`docs/superpowers/analysis/2026-08-18-t1-t5-fable-author-adjudication-round1.md`

Headline:

```text
M1 ACCEPT
  conditional materialized projection must serialize per-Document projection write before canonical read through write; no FIFO.

M2 ACCEPT ROOT CAUSE / REFINE FIX
  invalidate all restored ApplicationSessions before serving;
  fail closed until required post-snapshot security teardown is reconciled/proven;
  T7 chooses smallest recovery proof mechanism — no generic security journal frozen now.

M3 ACCEPT — OPTION (b)
  Search journey remains required;
  canonical PostgreSQL query/view is baseline for current canonical search facts;
  materialized projection + search_refresh + rebuild become conditional on T6 proving a real derived/expensive consumer or measured need.
```

LOW dispositions proposed:

```text
L1 title = Revision-governed metadata, preserving DRAFT/EFFECTIVE separation
L2 late renderer result = semantic no-op if Submission/Revision no longer eligible
L3 live admission claim/binding prevents GC until bounded release/expiry
L4 bounded initiator/manager withdraw of active human-governed obsolescence request
L5 T3 provider-disable wording aligned to T5-L
```

Notes proposed:

```text
same-DB job recovery coherence recorded as reopen guard
registry SUPERSEDED wording tightened cosmetically
T6 explicitly names source upload/admission UX
```

Everything else remains frozen.

## Exact next step

Operator ratification of the Round-1 adjudication.

If ratified:

```text
apply only the ratified bounded amendments to durable T1→T5 authorities + Decision Registry
→ update router/handoff/PR
→ Fable reads the adjudication/delta from GitHub and challenges only remaining material disagreement if operator dispatches it
→ explicitly close post-T5 checkpoint when disagreement set is empty/adjudicated
→ only then open T6
```

Do **not** open T6 before this checkpoint closes.

No final SQL/index/package/process topology, public API/frontend contract, Historical Migration execution plan, implementation plan or product code is authorized.
