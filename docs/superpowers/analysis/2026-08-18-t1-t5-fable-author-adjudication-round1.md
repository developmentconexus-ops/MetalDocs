# MetalDocs — Post-T5 Fable Review — Author Adjudication Round 1

> **Status:** OPERATOR-RATIFIED ROUND-1 ADJUDICATION / DURABLE AMENDMENTS APPLIED — FABLE DELTA REVIEW NEXT  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Fable evidence:** `docs/superpowers/analysis/2026-08-18-t1-t5-integrated-independent-fable-review.md` @ `bdef5fc3c4004aa3ab4deefc9e8373dd3efcf856`  
> **Delta review request:** `docs/superpowers/analysis/2026-08-18-t1-t5-fable-delta-review-request.md`  
> **Implementation:** BLOCKED  
> **T6:** NOT OPEN

This file records the operator-ratified disposition of the independent post-T5 Fable review. The detailed pre-ratification reasoning remains in Git history. The ratified bounded amendments have been promoted into the Product Contract/T1→T5 authorities and Decision Registry; no formal T-stage reopen occurred.

## 1. Ratified disposition

```text
Fable verdict                = APPROVE T1→T5 WITH MATERIAL FIXES
Operator disposition         = ACCEPT AUTHOR ROUND-1 ADJUDICATION
Formal T-stage reopen        = NONE
Durable bounded amendments   = APPLIED
Fable delta review           = NEXT
T6                           = NOT OPEN
implementation               = BLOCKED
```

Everything not named below remains frozen.

## 2. Material findings

```text
M1 ACCEPT
  Materialized Search is conditional.
  If activated, per-Document projection-write serialization is acquired before canonical read
  and held through rewrite/removal; rebuild obeys the same law; FIFO remains unnecessary.

M2 ACCEPT ROOT CAUSE / REFINED FIX
  All restored ApplicationSessions are invalidated before ordinary serving.
  Required known post-snapshot offboarding/access teardown/security revocations must be
  reconciled/proven before ordinary authenticated serving.
  T7 chooses the smallest recovery proof/choreography; no generic per-grant journal is frozen now.

M3 ACCEPT — OPTION (b)
  Search journey remains Launch-required.
  Baseline = canonical PostgreSQL query/view over current canonical search facts.
  Materialized projection + search_refresh + rebuild activate only when T6/current evidence proves
  a real derived/expensive searchable fact or measured need canonical query/view cannot sustain.
```

## 3. LOW findings

```text
L1 ACCEPT
  title = Revision-governed metadata; ordinary current reader/search title follows EFFECTIVE Revision.

L2 ACCEPT
  late renderer result for returned/withdrawn/cancelled dead candidate = semantic no-op;
  output becomes reclaimable mechanism content after claim release/expiry.

L3 ACCEPT
  live bounded admission claim/binding protects READY content from GC until consumed/released/expired.

L4 ACCEPT
  authorized initiator/manager may withdraw active human-governed obsolescence request;
  target remains EFFECTIVE; no fake participant verdict; later retry = new request.

L5 ACCEPT
  T3 explicitly follows T5-L: local offboarding is access-correct without provider-disable durable job.
```

## 4. Notes / guardrails

```text
N1 ACCEPT
  same-PostgreSQL durable intents restore coherently with semantic facts; a future separate job
  recovery domain must re-prove this property.

N2 ACCEPT
  registry disposition wording tightened where old SUPERSEDED labels could mislead cold readers.

N3 ACCEPT
  T6 REOPEN set explicitly includes source upload/T4 admission UX.
```

## 5. Durable promotion map

```text
Product Contract REV001
  title-as-Revision product meaning
  bounded active human-governed obsolescence withdrawal

T1
  Revision-governed title

T2
  obsolescence withdrawal lifecycle
  late-rendition eligibility consequence

T3
  obsolescence-withdraw authorization/Audit
  provider-disable wording aligned to T5-L

T4
  restore security non-resurrection
  admission-claim GC liveness
  same-DB durable-intent recovery coherence

T5
  canonical Search query/view baseline
  conditional materialized projection/search_refresh/rebuild
  conditional projection-write serialization
  late-rendition no-op

Decision Registry
  reconciled to all above
  T6 Search materialization proof + upload/admission UX routed explicitly
  T7 post-snapshot security-teardown recovery choreography routed explicitly
```

## 6. Round-2 / delta contract

Fable must read the ratified authorities directly from GitHub and review only the promoted delta using:

`docs/superpowers/analysis/2026-08-18-t1-t5-fable-delta-review-request.md`

Expected output path:

`docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-delta-review.md`

The desired close condition is:

```text
DISAGREEMENT SET = EMPTY
T6 READINESS = MAY OPEN
```

A disagreement must provide a concrete counterexample and the smallest exact correction; it does not reopen the whole architecture review.

## 7. Current gate

```text
T1→T5 + bounded amendments    OPERATOR-RATIFIED
Decision Registry             RECONCILED
Fable delta review request    STAGED
Fable delta verdict           PENDING
Post-T5 checkpoint            OPEN
T6                            NOT OPEN
implementation                BLOCKED
```
