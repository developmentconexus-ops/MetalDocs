# T1→T5 Post-Fable Ratified Delta — Independent Fable Review Request

> **Status:** ACTIVE STAGING / DELTA REVIEW REQUEST — NOT TARGET AUTHORITY  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Original Fable review:** `docs/superpowers/analysis/2026-08-18-t1-t5-integrated-independent-fable-review.md` @ `bdef5fc3c4004aa3ab4deefc9e8373dd3efcf856`  
> **Author adjudication:** `docs/superpowers/analysis/2026-08-18-t1-t5-fable-author-adjudication-round1.md`  
> **Operator disposition:** Round-1 adjudication APPROVED; bounded amendments promoted  
> **T6:** NOT OPEN  
> **Implementation:** BLOCKED

This is a **delta review**, not a second full architecture review. Reconstruct authority from the repository and verify only whether the operator-ratified corrections close the original findings without introducing a new material contradiction.

## 1. Read order

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md` — now Product Contract `REV001`
5. `wiki/architecture/r10-t1-semantic-state-invariants.md`
6. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
7. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
8. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
9. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
10. `wiki/architecture/rebaseline-decision-registry.md`
11. `wiki/architecture/r10-technical-architecture.md`
12. original independent review
13. Round-1 author adjudication
14. this request

Review evidence is not authority. Do not alter durable authorities directly.

## 2. Ratified delta to verify

### M1 — optional materialized Search concurrency

Materialized Search is no longer baseline. **If** T6 activates it, one Document's projection-write serialization is acquired **before canonical state read** and held through rewrite/removal. Rebuild obeys the same law. FIFO/broker ordering remains rejected.

Verify that overlapping older/newer workers cannot leave an older observation as final projection state.

### M2 — restore security non-resurrection

T4 now requires:

```text
all restored ApplicationSessions invalidated before ordinary serving
+
required known post-snapshot offboarding/access teardown/security revocations reconciled/proven
+
fail-closed ordinary authenticated serving until selected recovery proof is complete
```

T4 intentionally does **not** freeze a generic per-grant security journal. T7 owns the concrete recovery evidence/choreography.

Verify that this closes the access-resurrection defect while preserving the "prepare the seam, not the entire platform" law.

### M3 — Search mechanism before consumer

Search journey remains Launch-required, but baseline is now:

```text
canonical PostgreSQL query/view over current canonical searchable facts
```

Only if T6 names a real derived/expensive searchable fact or measured scale/ranking/language need may it activate:

```text
materialized PostgreSQL Search projection
+ search_refresh(document_id)
+ rebuild/reconciliation
+ M1 serialization law
```

Verify that no dormant materialized Search machinery remains mandatory and that T6 has the correct proof/activation seam.

### L1 — title

Product Contract `REV001` + T1 now bind human-readable title to Revision. Ordinary readers/search use the current EFFECTIVE Revision's title; a newer DRAFT/SUBMITTED retitle cannot rewrite current reader truth.

### L2 — late rendition

A renderer result for a returned/withdrawn/cancelled no-longer-eligible Submission is a semantic no-op; no OfficialRendition/Release is created and physical output becomes reclaimable after T4 admission-claim release/expiry.

### L3 — GC liveness

A live bounded admission claim/binding prevents a READY handle from becoming GC-eligible until consumed/released/expired; GC rechecks claims immediately before delete.

### L4 — active obsolescence withdrawal

Product Contract `REV001`, T2 and T3 now allow bounded withdrawal of an active **human-governed** obsolescence request by authorized initiator/manager using existing `document.obsolete` semantics. Target remains EFFECTIVE; no fake RETURN/ACCEPT; Audit is required; later retry is a new request.

### L5 — provider disable

T3 now explicitly follows T5-L: local offboarding is access-correct without a provider-disable durable job. Provider convergence is future assurance scope only when explicitly promoted.

### Notes / registry

- same-DB durable-intent restore coherence is now recorded as a positive recovery property/reopen guard;
- ambiguous `SUPERSEDED` registry wording was tightened;
- T6 REOPEN set explicitly includes source upload/T4 admission UX and the Search materialization proof question;
- T7 REOPEN set explicitly includes post-snapshot security-teardown recovery choreography.

## 3. Required output

Write the delta verdict to:

`docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-delta-review.md`

Commit/push to the same branch/PR.

Return exactly:

```text
DELTA VERDICT = APPROVE | MATERIAL DISAGREEMENT

ORIGINAL FINDINGS:
M1 = CLOSED | OPEN
M2 = CLOSED | OPEN
M3 = CLOSED | OPEN
L1 = CLOSED | OPEN
L2 = CLOSED | OPEN
L3 = CLOSED | OPEN
L4 = CLOSED | OPEN
L5 = CLOSED | OPEN

NEW MATERIAL FINDINGS = <count>
DISAGREEMENT SET = EMPTY | <exact minimal set>
T6 READINESS = MAY OPEN | BLOCKED
```

If `DISAGREEMENT SET = EMPTY`, no long prose is needed. Briefly state why the delta closes the original defect classes.

If there is disagreement, provide only a concrete counterexample, exact authority affected, smallest correction and everything that remains frozen. Do not restart a general architecture critique.

## 4. Gate

```text
T1→T5 authorities = operator-ratified with bounded post-Fable amendments
Decision Registry = reconciled
Fable delta verdict = PENDING
T6 = NOT OPEN
implementation = BLOCKED
```

The checkpoint closes only after the delta verdict is read/adjudicated through GitHub.