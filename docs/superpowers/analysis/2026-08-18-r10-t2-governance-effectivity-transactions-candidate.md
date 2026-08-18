# R10-T2 — Governance, Effectivity & Lifecycle Transactions — Corrected Integrated Candidate

> **Status:** ACTIVE STAGING — MATERIAL DECISIONS ADJUDICATED / PLATFORM SUMMARY RATIFICATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Technical authority:** `wiki/architecture/r10-technical-architecture.md`  
> **T1 authority:** promoted T1 section in `wiki/architecture/r10-technical-architecture.md`  
> **Adjudication:** `docs/superpowers/analysis/2026-08-18-r10-t2-operator-adjudication.md`  
> **Implementation:** BLOCKED

T2 derives the smallest transaction/concurrency/governance behavior needed to make the operator-ratified T1 semantic facts behave correctly in Launch V1.

The operator accepted T2-A→T2-N on 2026-08-18 with one bounded correction to the candidate: **initial issuance is `REV000`; `REV001` is the first subsequent revision.** This file is corrected to that authority. T2 is not closed until the operator ratifies the required platform-facing summary.

It does **not** define final SQL/table/index syntax, Go package layout, concrete Authorization permission names/bundles, storage locator design, worker topology, API routes, frontend UX, or migration execution.

---

## 1. T2 decision question

> **How must MetalDocs change the accepted T1 facts together so create/edit/submit/govern/return/withdraw/cancel/release/obsolete remain exact and atomic under concurrency without turning document governance into a generic workflow platform?**

T2 succeeds when every named Launch transition has one explicit eligibility rule, one atomic business outcome, one serialization/OCC strategy and no ambiguous intermediate authority.

---

## 2. Accepted revision convention

```text
REV000 = initial issuance
REV001 = first revision after initial issuance
REV002 = second revision
...
```

Therefore:

```text
create Document → REV000 DRAFT
first Release   → REV000 EFFECTIVE
first later change cycle → REV001 DRAFT
first replacement Release → REV000 SUPERSEDED + REV001 EFFECTIVE
```

Revision ordinals never reuse.

---

## 3. Authority / evidence boundary

Authority, in order:

1. `AGENTS.md`;
2. DevelopmentConexus Engineering Method v1.0.0;
3. current handoff;
4. accepted Product Contract;
5. operator-adjudicated Whole-Product GCR;
6. operator-approved 4+1 ownership topology + future-evolution law;
7. operator-ratified T1 semantic state/invariants promoted into `wiki/architecture/r10-technical-architecture.md`.

Prior B1/B3/B4/B6 and paused R10-C are evidence only. Their old context ownership, Artifact model, Distribution coupling, strict workflow richness and global Audit lock law are not entitlement.

---

## 4. Target invariant

> **Every lifecycle command revalidates canonical state under one local transaction/serialization boundary, mutates only the facts it owns, freezes immutable evidence before exposing success, and never depends on an external provider call to make the local business transition atomic.**

Failure classes T2 prevents include:

```text
Submission freezes one generation while Revision moves on another
approval accepts content A while Release publishes content B
concurrent releases create two EFFECTIVE revisions
withdraw/cancel destroys prior Submission history
group membership drift rewrites who could act on an active Step
route edits reinterpret in-flight governance
obsolescence races a replacement Revision
external rendition/provider calls become part of business atomicity
```

---

# 5. Accepted transaction / concurrency posture

## 5.1 One local ACID transaction per native business transition

Native lifecycle transitions commit in one local MetalDocs product-state transaction. No Keycloak call, object-store call, renderer call, notification delivery or repository/provider call joins that transaction.

## 5.2 Document as lifecycle serialization root

Operations that can change lifecycle/effectivity truth of one existing Document serialize on that Document before changing subordinate lifecycle state:

```text
new Revision
SUBMIT / RETURN / WITHDRAW / CANCEL
Step decision when it can advance/terminate the attempt
Release
obsolescence initiation/completion
```

DRAFT WorkingContent mutation uses OCC as the primary race arbiter.

## 5.3 WorkingContent OCC

Every governed DRAFT mutation carries the caller-observed generation:

```text
expected == current → accept + increment once
expected != current → stale/conflict; no silent overwrite
```

Autosave never allocates a business Revision. Future CRDT may replace the DRAFT concurrency mechanism without changing Revision/Submission meaning.

## 5.4 Isolation posture

Accepted design posture:

```text
PostgreSQL READ COMMITTED
+ explicit narrow serialization/CAS/constraints
```

No global `SERIALIZABLE` requirement is introduced. Exact SQL enforcement waits for implementation design after final R10 ratification.

---

# 6. Create / numbering / template-based creation

A successful blank creation atomically establishes:

```text
allocated stable Document code
Document identity
REV000 DRAFT
initial WorkingContent
```

No successful partial shell exists. Committed codes never rebind/reuse. Gap-free numbering is not required; uniqueness and no reuse are.

Create-from-template additionally revalidates at commit that:

```text
source Document still has Template role/eligibility
selected source Revision is still the Template's current EFFECTIVE Revision
exact source-content provenance is pinned
new Document receives an independent REV000 WorkingContent seed
```

Later Template changes never rebind the derived Document.

---

# 7. SUBMIT

SUBMIT serializes the lifecycle boundary and requires:

```text
Revision is current open DRAFT
caller supplies expected WorkingContent generation
expected generation is still canonical
all Launch submit requirements are satisfied
DocumentType governance + representation config is read coherently
```

One successful SUBMIT atomically:

1. freezes immutable Submission from exactly that WorkingContent generation;
2. freezes governance mode/route snapshot;
3. freezes official-representation requirement;
4. changes Revision `DRAFT → SUBMITTED`;
5. creates GovernanceAttempt only for `UseGovernanceRoute`.

`NoHumanApproval` creates no fake GovernanceAttempt/System approver.

---

# 8. Governance route — smallest Launch semantics

## 8.1 Actor selector

Accepted Launch selector vocabulary:

```text
NAMED_USER
GROUP
```

`ROLE_IN_AREA` is absent from the Launch baseline.

## 8.2 Sequential one-active-Step rule

Exactly one human Step is active at a time.

At activation:

- `NAMED_USER` resolves the configured User;
- `GROUP` resolves a concrete snapshot of current enabled Group members;
- later Group membership drift does not rewrite that active candidate set;
- current Authorization is rechecked when a User actually acts.

## 8.3 Group completion

Group Step completion is **ANY-one**: one currently authorized User from the activation snapshot may make the Step decision.

No `ALL`, N-of-M or generic quorum engine is part of Launch.

## 8.4 Bounded SoD

Accepted minimum independence rule:

```text
Submission submitter cannot satisfy a human Step on that same Submission attempt.
Obsolescence initiator cannot satisfy a human Step on that same obsolescence attempt.
```

No baseline cross-Step same-user prohibition exists. Stronger SoD is a reopen trigger on concrete regulation/customer requirement.

## 8.5 No reassign/overseer baseline

If a frozen route becomes impossible, the safe Launch escape is:

```text
withdraw current Submission attempt
→ same Revision DRAFT
→ admin fixes current route config
→ resubmit creates new immutable Submission/attempt snapshot
```

No invisible mutation of an active historical route.

---

# 9. ACCEPT / RETURN / resubmit

ACCEPT binds exact attempt + Step + subject, records actual User/trusted time and is immutable evidence. If more Steps remain, the next activates; final ACCEPT satisfies the human gate.

For Submission governance, RETURN atomically:

```text
record immutable RETURN decision/reason
terminate GovernanceAttempt as RETURNED
Revision SUBMITTED → DRAFT
preserve old Submission + decisions + feedback
```

A later SUBMIT creates a new Submission and, if human governance applies, a new GovernanceAttempt.

For obsolescence governance, RETURN terminates only that obsolescence attempt; the target Revision remains EFFECTIVE.

---

# 10. Withdraw and cancel

## Withdraw

Before Release, authorized withdrawal:

```text
terminate active Submission governance execution as WITHDRAWN
Revision SUBMITTED → DRAFT
preserve old Submission + decisions/feedback
```

No participant verdict is fabricated.

A governance-satisfied Submission waiting only for required Rendition is still pre-Release and may be withdrawn.

## Cancel Revision

Cancellation of a DRAFT or eligible pre-Release SUBMITTED Revision atomically:

```text
terminate any live GovernanceAttempt without fake verdict
record immutable cancellation reason/actor/time
Revision → CANCELLED
```

Older EFFECTIVE Revision remains EFFECTIVE. Cancelled ordinal never reuses.

---

# 11. Release gates and effectivity

Release gates are orthogonal:

```text
human gate:
  NoHumanApproval     → satisfied by absence
  UseGovernanceRoute  → final required ACCEPT

representation gate:
  SourceOnly                 → satisfied by absence
  RequireOfficialRendition   → exact successful OfficialRendition for winning Submission
```

## 11.1 Immediate first Release

For example:

```text
REV000 DRAFT
+ NoHumanApproval
+ SourceOnly
→ freeze Submission
→ system Release in same local transaction
→ REV000 EFFECTIVE
```

Submission and Release remain distinct semantic facts even when committed together.

## 11.2 Final ACCEPT / Rendition waiting

If final ACCEPT satisfies the last missing gate and required Rendition already exists, system Release may occur in that same transaction.

If Rendition is still missing:

```text
GovernanceAttempt satisfied
Revision remains SUBMITTED
Release waits truthfully
```

Later exact Rendition completion may trigger Release after canonical revalidation. Renderer/provider execution stays outside the human decision transaction.

## 11.3 Replacement Release

Serialized on the stable Document and atomic:

```text
prior EFFECTIVE → SUPERSEDED
winning successor SUBMITTED → EFFECTIVE
Release records predecessor + exact winning Submission
```

First replacement example:

```text
REV000 EFFECTIVE
REV001 SUBMITTED + all gates satisfied
→ REV000 SUPERSEDED
→ REV001 EFFECTIVE
```

No successful observable state contains two EFFECTIVE revisions. Distribution/Acknowledgement is absent from Launch-Core Release atomicity.

---

# 12. Governed obsolescence without replacement

Obsolescence may start only when:

```text
target Revision is current EFFECTIVE
mandatory reason is nonblank
no open replacement Revision exists
no other active obsolescence attempt exists
```

While obsolescence is active, creation of a new Revision is blocked. This prevents simultaneous contradictory intents: “replace it” and “remove it without successor.”

Launch reuses the same current DocumentType governance mode/route for obsolescence; no separate obsolescence workflow family exists.

## NoHumanApproval

```text
authorized initiation
+ mandatory reason
+ eligibility/conflict checks
+ NoHumanApproval
→ zero human Step / no fake approver
→ system completes in same transaction
→ current EFFECTIVE → OBSOLETE
```

## Human-governed

The target remains EFFECTIVE while route governance is active. Final ACCEPT rechecks target/current state/no replacement/active attempt under Document serialization, then atomically changes the current EFFECTIVE Revision to OBSOLETE. RETURN ends the request and leaves it EFFECTIVE.

No OBSOLETE reactivation exists in Launch.

---

# 13. Route/configuration concurrency

Current DocumentType governance/representation configuration is mutable current truth, but every Submission/obsolescence attempt freezes one coherent snapshot.

Accepted law:

```text
configuration edit = atomic whole-config mutation
attempt creation    = atomic coherent snapshot
```

Concurrent admin edit vs attempt creation resolves to complete old or complete new config, never a mixed route. A first-class browsable `PolicyVersion` object remains unnecessary without another consumer.

---

# 14. Failure / restart laws

- rollback leaves no successful partial Submission/decision/Release/obsolescence transition;
- external provider failure cannot retroactively invalidate committed semantic history;
- governance-satisfied Submission waiting for Rendition is truthful durable state and resumes after restart;
- repeated Release trigger is idempotent against canonical Release eligibility/fact identity;
- retries cannot fabricate duplicate semantic decisions;
- stale commands fail/reload rather than overwrite newer lifecycle truth.

API/transport idempotency contract is T6; T2 defines semantic uniqueness.

---

# 15. Named-future compatibility

| Future capability | T2 seam preserved |
|---|---|
| Distribution / Read & Acknowledge | reacts after Release; no audience obligation inside Release atomicity |
| Periodic Review | later serializes against exact current EFFECTIVE Revision; not a Release gate |
| Dossier | references stable Document; owns no lifecycle transaction |
| Evidence | may have independent lifecycle/transactions rather than reuse Revision/Release |
| Records/Hold/Disposition | attaches to immutable Release/Submission/Revision history; lifecycle status never means delete |
| Governed Export | reads stable transactionally consistent facts; owns none |
| Repository connectors | provider effects remain outside local lifecycle transaction |
| Training/LMS | consumes effective/released truth later; does not become Release gate accidentally |
| Change Control | may orchestrate multiple Documents but cannot take each Document's lifecycle authority |
| pooled tenancy | transaction laws remain local; substrate may reopen around stable Company identity |
| CRDT | may replace DRAFT concurrency; SUBMIT still freezes exact accepted state |

---

# 16. Proof obligations before implementation

Later implementation planning must make these falsifiable, at minimum:

```text
concurrent create/code allocation cannot duplicate committed Document code
initial create always produces REV000, never REV001
stale DRAFT mutation cannot overwrite newer WorkingContent
SUBMIT cannot freeze stale generation
route edit cannot create a mixed in-flight snapshot
Group membership drift cannot rewrite activated Step candidates
RETURN/withdraw/cancel never mutate prior Submission/decision history
concurrent Release cannot create two EFFECTIVE Revisions
replacement Release cannot expose predecessor+successor EFFECTIVE together
required Rendition mismatch cannot Release a different Submission
provider outage cannot create false EFFECTIVE truth
obsolescence cannot complete against a no-longer-current EFFECTIVE target
new Revision cannot race active obsolescence under the accepted mutual-exclusion law
retry/restart cannot duplicate one semantic Release/decision result
```

These are design proof obligations, not claims that implementation/tests exist today.

---

# 17. Explicit non-decisions

T2 does not decide:

```text
final SQL table/index/trigger names
exact lock SQL / full global lock ordering
exact Role/Permission names/bundles
AuditEvent census/facts
object-store handles / staging keys / scan implementation
renderer/outbox/worker/retry topology
Search technology
API route/error envelope/idempotency-key contract
frontend workflow screens
migration import transaction shapes
cross-Step strict SoD beyond bounded initiator rule
fresh-auth/eSignature
quorum ALL/N-of-M
reassignment/overseer/SLA/escalation
separate obsolescence-specific route configuration
scheduled Release
```

---

# 18. Reopen triggers

Reopen only the implicated T2 decision if material evidence shows:

- Launch needs `ALL`/N-of-M quorum rather than ANY-one Group Step;
- a real route must select by product Role-in-Area rather than NamedUser/Group;
- regulation/customer contract requires strict cross-Step SoD, eSignature/fresh-auth or mandatory reauthentication;
- live reassignment/overseer is required and withdraw→fix→resubmit is insufficient;
- obsolescence must use a distinct route;
- Change Control needs active replacement Revision and obsolescence intent to coexist;
- scheduled/future-dated effectivity becomes a real requirement;
- implementation evidence proves `READ COMMITTED` + narrow locks/CAS cannot enforce invariants without disproportionate complexity.

---

# 19. Adjudication result / current gate

Accepted dispositions:

```text
T2-A ACCEPT
T2-B ACCEPT
T2-C ACCEPT WITH REV000 CORRECTION
T2-D ACCEPT
T2-E ACCEPT
T2-F ACCEPT
T2-G ACCEPT
T2-H ACCEPT
T2-I ACCEPT
T2-J ACCEPT
T2-K ACCEPT
T2-L ACCEPT
T2-M ACCEPT
T2-N ACCEPT
```

Current gate:

```text
T2 material decisions       = ACCEPTED
T2 platform-facing summary  = NEXT
T2 promotion/closure        = PENDING SUMMARY RATIFICATION
T3                          = NOT OPEN
implementation              = BLOCKED
```

The pre-adjudication candidate is preserved in Git history; this corrected staging file is the active T2 design packet until summary ratification and durable promotion.