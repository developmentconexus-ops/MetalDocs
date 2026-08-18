# R10-T2 — Governance, Effectivity & Lifecycle Transactions — Integrated Candidate

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE CANDIDATE — OPERATOR ADJUDICATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Technical authority:** `wiki/architecture/r10-technical-architecture.md`  
> **T1 authority:** promoted T1 section in `wiki/architecture/r10-technical-architecture.md`  
> **Implementation:** BLOCKED

T2 derives the smallest transaction/concurrency/governance behavior needed to make the operator-ratified T1 semantic facts behave correctly in Launch V1.

It does **not** define final SQL/table/index syntax, Go package layout, concrete Authorization permission names/bundles, storage locator design, worker topology, API routes, frontend UX, or migration execution.

---

## 1. T2 decision question

> **How must MetalDocs change the accepted T1 facts together so create/edit/submit/govern/return/withdraw/cancel/release/obsolete remain exact and atomic under concurrency without turning document governance into a generic workflow platform?**

T2 succeeds when every named Launch transition has one explicit eligibility rule, one atomic business outcome, one serialization/OCC strategy and no ambiguous intermediate authority.

---

## 2. Authority / evidence boundary

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

## 3. Known / Reopened / Deferred

### Known

T2 must preserve:

```text
Document != Revision != WorkingContent != Submission
one open Revision at most per Document
one EFFECTIVE Revision at most per Document
prior EFFECTIVE may coexist with newer DRAFT/SUBMITTED Revision
Revision ordinal never reuses
WorkingContent mutable only for DRAFT
Submission immutable exact attempt
same-Revision resubmit creates new Submission
GovernanceAttempt closed to SUBMISSION | OBSOLESCENCE
one sequential Step semantic
normal human outcomes ACCEPT | RETURN_FOR_CHANGES
withdraw Submission attempt != cancel Revision
Release is system-owned effectivity authority
replacement Release atomically supersedes predecessor
OfficialRendition is a Release gate only when required
NoHumanApproval creates no fake approver/attempt
obsolescence without successor is explicit and governed
NoHumanApproval obsolescence may have zero human Step
Audit is not current state
storage/provider identity is not semantic identity
```

### Reopened from old R10

The following are not automatically retained:

```text
ANY|ALL configurable quorum
RoleInArea actor selector
strict submitter/creator SoD as broad policy engine
cross-Step same-user SoD
fresh-auth
reassign/overseer machinery
due dates/SLA/escalation
scheduled ReleasePlan.not_before
Distribution obligations in Release transaction
Approval as separate owner
nested commits/cross-owner repository imports
global SERIALIZABLE
global AuditChainHead terminal lock
```

### Deferred

```text
T3 → exact permission catalog, check sites, Audit census/facts
T4 → byte locator/staging/integrity/malware/restore
T5 → renderer worker/outbox/retry/Search/notifications
T6 → public API/frontend/viewer journeys/idempotency UX
T7 → historical migration/cutover transaction shapes
```

---

## 4. Root cause / target invariant

Failure class: **split lifecycle atomicity**.

A correct semantic model can still fail if:

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

Target invariant:

> **Every lifecycle command revalidates canonical state under one local transaction/serialization boundary, mutates only the facts it owns, freezes immutable evidence before exposing success, and never depends on an external provider call to make the local business transition atomic.**

---

# 5. Transaction and serialization posture

## 5.1 One local ACID product transaction for one business transition

Native lifecycle transitions commit in one local MetalDocs product-state transaction. No Keycloak call, object-store call, renderer call, notification delivery or repository/provider call joins that transaction.

Provider/external effects that must happen later become T4/T5 mechanisms.

## 5.2 Document is the lifecycle serialization root

All operations that can change the lifecycle/effectivity truth of one existing Document serialize on that Document before changing subordinate lifecycle state:

```text
new Revision
SUBMIT / RETURN / WITHDRAW / CANCEL
Step decision when it can advance/terminate the attempt
Release
obsolescence initiation/completion
```

This is a semantic lock-order law, not final SQL syntax.

For DRAFT content mutation, WorkingContent uses OCC as the primary race arbiter; a Document lifecycle serialization boundary is acquired when a mutation also crosses a lifecycle boundary such as SUBMIT.

## 5.3 Default isolation

Candidate posture: keep ordinary PostgreSQL `READ COMMITTED` plus explicit row-level serialization/CAS on the relevant canonical rows. Do not introduce global `SERIALIZABLE` when narrow constraints/locks/OCC can prove the invariants.

Exact SQL lock clauses and index/constraint implementation are implementation-spec work after R10 ratification, but T2 must define the admitted lock classes/order sufficiently for later proof.

## 5.4 No nested semantic commits

A composed lifecycle use case owns the transaction boundary. Lower semantic seams must be transactionally composable and must not commit independently or import another owner's repository merely to obtain atomicity.

T3 later adds required same-commit Audit without changing this business transaction ownership.

---

# 6. Create / numbering / template-based creation

## 6.1 Blank creation

One successful create command atomically establishes:

```text
allocated stable Document code
Document identity
REV001 DRAFT
initial WorkingContent
```

A partial shell is never successful.

Committed Document codes never rebind to another Document. T2 does not promise gap-free human numbering; uniqueness and no reuse matter, not cosmetic contiguity.

Read-only number preview reserves nothing.

## 6.2 Template-based creation

Create-from-template additionally proves at commit that:

```text
source Document still has Template role/eligibility
selected source Revision is still that Template's current EFFECTIVE Revision
exact source-content provenance is pinned
new Document receives an independent WorkingContent copy/seed
```

Later Template changes never rebind the derived Document.

The source Template lifecycle is revalidated/serialized narrowly for this operation so a concurrent Template Release cannot silently switch the source between selection and commit.

---

# 7. DRAFT WorkingContent / autosave concurrency

WorkingContent keeps one monotonically increasing `working_generation`/equivalent OCC token.

Every accepted governed DRAFT mutation carries the caller-observed expected generation:

```text
expected generation == current
→ accept mutation
→ increment generation once

expected generation != current
→ reject stale mutation / conflict
```

Laws:

- no last-write-wins silent overwrite for governed DRAFT state;
- autosave never allocates a business Revision;
- autosave/checkpoint recovery may use technical mechanism state later, but never becomes official history;
- no EditorSession/checkout is required for correctness;
- future CRDT may replace the DRAFT concurrency mechanism without changing Revision/Submission semantics.

---

# 8. SUBMIT transaction

SUBMIT serializes the Document/lifecycle boundary and requires:

```text
Revision is current open DRAFT
caller supplies expected WorkingContent generation
expected generation still equals canonical WorkingContent generation
all Launch submit requirements satisfied
current DocumentType governance + representation configuration can be read as one coherent committed configuration
```

One successful SUBMIT atomically:

1. freezes immutable Submission from exactly that WorkingContent generation;
2. freezes the exact governance mode/route snapshot for that attempt;
3. freezes official-representation requirement;
4. changes Revision from DRAFT → SUBMITTED;
5. creates a GovernanceAttempt only when mode = `UseGovernanceRoute`.

NoHumanApproval creates no fake GovernanceAttempt/System decision.

Later DocumentType/route changes affect future attempts only.

---

# 9. Governance route — smallest Launch participant semantics

## 9.1 Actor selector

Candidate Launch selector vocabulary:

```text
NAMED_USER
GROUP
```

`ROLE_IN_AREA` is rejected for Launch unless a real journey proves it. Authorization roles are access bundles and should not automatically become business-routing identity.

## 9.2 One active Step at a time

The route is sequential. Exactly one Step is active at a time for a live human GovernanceAttempt.

At activation:

- `NAMED_USER` resolves the configured User;
- `GROUP` resolves a concrete snapshot of current enabled Group members;
- later Group membership drift does not rewrite that active Step's candidate set;
- current Authorization is still rechecked when someone actually acts; snapshot membership is not a permission grant.

## 9.3 Group completion

Launch Group Step completion is **ANY-one**:

> one currently authorized User from the activation snapshot may make the Step decision.

No configurable ALL/quorum engine is introduced.

This keeps Groups useful for resilient routing while avoiding dormant voting/quorum machinery.

## 9.4 Human independence — bounded SoD

Recommended minimum Launch rule:

```text
Submission submitter cannot satisfy a human Step on that Submission attempt.
Obsolescence initiator cannot satisfy a human Step on that obsolescence attempt.
```

This gives `UseGovernanceRoute` actual independent-human meaning without restoring the old broad strict-SoD engine.

No baseline rule forbids the same non-initiating User from satisfying two different Steps if the configured NamedUser/Group selectors legitimately include that User. A future regulatory/customer requirement may add cross-Step SoD deliberately.

This bounded rule is a material T2 operator decision; see T2-K.

## 9.5 No reassign/overseer baseline

Launch has no general reassign/overseer operation.

If a route snapshot becomes impossible (for example an unavailable NamedUser), the safe bounded escape is:

```text
withdraw current Submission attempt
→ same Revision DRAFT
→ administrator fixes current route configuration
→ resubmit creates a new immutable Submission/attempt snapshot
```

This preserves history instead of mutating an active route invisibly.

---

# 10. ACCEPT / RETURN / resubmit

## 10.1 ACCEPT

An ACCEPT decision:

```text
binds exact active GovernanceAttempt + Step + governed subject
records actual User and trusted time
requires current T3 Authorization when implemented
is immutable evidence
```

If more Steps remain, the next Step activates.

If this was the final required Step, the human governance gate becomes satisfied.

## 10.2 RETURN_FOR_CHANGES — Submission governance

Any eligible actor on the active Step may return the Submission for changes.

One transaction:

```text
record immutable RETURN decision/reason
terminate GovernanceAttempt as RETURNED
Revision SUBMITTED → DRAFT
preserve old Submission + prior Step decisions/feedback
preserve/increment WorkingContent generation continuity; do not reset business history
```

A later SUBMIT creates a new Submission and, when human governance applies, a new GovernanceAttempt.

## 10.3 RETURN_FOR_CHANGES — Obsolescence governance

For an obsolescence attempt, RETURN does **not** put the Document/Revision into DRAFT.

It terminates that obsolescence request/attempt as returned/unsuccessful; the target Revision remains EFFECTIVE. A later retry creates a new obsolescence request/attempt with a new immutable reason/snapshot.

No mutable generic “obsolescence draft” workflow state is introduced.

---

# 11. Withdraw Submission attempt

Before Release, an authorized author may withdraw an active Submission attempt to continue editing the same business Revision.

One transaction:

```text
active Submission governance execution → WITHDRAWN/terminated
Revision SUBMITTED → DRAFT
old Submission + decisions/feedback remain immutable history
WorkingContent remains the same Revision's mutable DRAFT authority
```

Withdraw creates no ACCEPT/RETURN decision and is not Revision cancellation.

A governance-satisfied Submission that is only waiting on a required Rendition is still pre-Release and may be withdrawn if the accepted product journey permits authorized withdrawal before Release; T2 recommends YES because Release, not human-gate completion, is the irreversible effectivity boundary.

---

# 12. Cancel open Revision

An authorized actor may cancel the current open Revision while it is DRAFT or pre-Release SUBMITTED.

One transaction:

```text
if a GovernanceAttempt is live, terminate it without fabricating a participant verdict
write immutable RevisionCancellation reason/actor/time
Revision → CANCELLED
```

Prior EFFECTIVE Revision, if any, remains EFFECTIVE.

All prior Submissions/decisions/feedback remain history.

A cancelled Revision never reopens and its ordinal never reuses.

---

# 13. Release gate and effectivity

Release is system-owned and can occur only for the exact winning Submission.

Release gates are orthogonal:

```text
human governance gate:
  NoHumanApproval → satisfied by absence
  UseGovernanceRoute → satisfied only after final ACCEPT

representation gate:
  SourceOnly → satisfied by absence
  RequireOfficialRendition → satisfied only by exact successful OfficialRendition for that Submission
```

## 13.1 Immediate Release when all gates are already satisfied

If SUBMIT creates a Submission where all Release gates are synchronously satisfied, the system may execute Release in the **same local business transaction** rather than exposing a meaningless durable intermediate state.

Canonical example:

```text
NoHumanApproval + SourceOnly
DRAFT
→ freeze Submission
→ system Release
→ EFFECTIVE
```

Submission and Release remain distinct immutable semantic facts even if committed together.

## 13.2 Final ACCEPT and Release

If the final human ACCEPT satisfies the last missing gate and any required OfficialRendition already exists, the system may perform Release in the same transaction as that final decision.

If the required Rendition does not yet exist:

```text
GovernanceAttempt = satisfied
Revision remains SUBMITTED
Release waits
```

Later successful rendition completion may trigger system Release after rechecking canonical eligibility. Renderer/provider execution is T4/T5 mechanism, not part of the human decision transaction.

## 13.3 First Release

Atomically:

```text
winning Revision SUBMITTED → EFFECTIVE
Release fact written for exact winning Submission
```

## 13.4 Replacement Release

Serialized on the stable Document and atomically:

```text
prior current EFFECTIVE Revision → SUPERSEDED
winning successor SUBMITTED → EFFECTIVE
Release records predecessor identity + winning Submission
```

No externally observable successful state may contain two EFFECTIVE revisions for one Document.

Distribution/Acknowledgement is absent from this transaction in Launch Core.

---

# 14. Governed obsolescence without replacement

## 14.1 Initiation eligibility

T2 recommends the smallest unambiguous Launch rule:

An obsolescence request may start only when:

```text
target Revision is the Document's current EFFECTIVE Revision
mandatory reason is nonblank
no open replacement Revision exists
no other active obsolescence attempt exists for the Document
```

While an obsolescence attempt is active, creation of a new Revision is blocked. This prevents simultaneous “replace it” and “remove it without replacement” business intents.

The Product Contract only requires the replacement Revision to be resolved before completion; this stricter Launch admission rule removes an otherwise useless contradictory intermediate state. Reopen if a concrete future Change-Control journey needs concurrent intents.

## 14.2 Governance mode

Launch uses the same current DocumentType governance mode/route semantics for obsolescence as for Submission governance. T2 does not add a separate obsolescence workflow configuration family.

A future concrete requirement for a different retirement route is an additive configuration reopen, not a reason to build it now.

## 14.3 NoHumanApproval obsolescence

Operator-ratified T1-J applies:

```text
authorized initiation
+ reason
+ eligibility/conflict checks
+ NoHumanApproval
→ no GovernanceAttempt / no human Step / no fake approver
→ system completes obsolescence in the same transaction
→ EFFECTIVE → OBSOLETE
```

Immutable obsolescence domain evidence remains mandatory.

## 14.4 Human-governed obsolescence

`UseGovernanceRoute` creates an immutable obsolescence request + bounded GovernanceAttempt using the frozen route snapshot.

The target stays EFFECTIVE while governance is active.

Final ACCEPT rechecks under Document serialization that:

```text
target is still current EFFECTIVE
no open replacement Revision exists
attempt is still the active obsolescence attempt
```

Then one transaction changes:

```text
current EFFECTIVE Revision → OBSOLETE
no successor becomes EFFECTIVE
immutable obsolescence result/time is recorded
```

RETURN terminates the attempt and leaves the target EFFECTIVE.

No reactivation of OBSOLETE exists in Launch.

---

# 15. Route/configuration concurrency

Current DocumentType governance/representation configuration is mutable current truth, but each Submission/obsolescence attempt freezes one coherent configuration snapshot.

T2 requires:

```text
configuration mutation = atomic whole configuration mutation
attempt creation = atomic coherent snapshot
```

A concurrent admin edit and SUBMIT/obsolescence initiation must resolve to either the complete old configuration or the complete new configuration, never a mixed route.

This requirement does not force a first-class browsable `ApprovalPolicyVersion` object.

---

# 16. Failure / restart laws

- rollback leaves no successful partial Submission/decision/Release/obsolescence transition;
- external provider failure cannot retroactively invalidate committed semantic history;
- a governance-satisfied Submission waiting for Rendition is truthful durable state and may be resumed after restart;
- repeated system Release trigger must be idempotent against canonical Release eligibility/fact identity;
- repeated user command transport retries must not fabricate duplicate semantic decisions; request/API idempotency realization is T6, while T2 defines which semantic results are unique/append-only;
- any stale command must fail/reload rather than overwrite newer lifecycle truth.

---

# 17. Named-future compatibility attack

| Future capability | T2 seam preserved |
|---|---|
| Distribution / Read & Acknowledge | reacts after Release; no audience obligation inside Release atomicity |
| Periodic Review | serializes against exact current EFFECTIVE Revision later; does not enter Release gate |
| Dossier | references stable Document; no lifecycle transaction ownership |
| Evidence | may have its own lifecycle/transactions rather than reusing Revision/Release |
| Records/Hold/Disposition | attaches to immutable Release/Submission/Revision history; business status still never means delete |
| Governed Export | reads stable transactionally consistent semantic facts later; does not own them |
| Repository connectors | provider effects remain outside local lifecycle transaction |
| Training/LMS | consumes effective/released truth later; never becomes Release gate by accident |
| Change Control | may orchestrate multiple Documents later but must not take over each Document's lifecycle serialization authority |
| pooled tenancy | local transaction laws remain per product-state authority; substrate may reopen around Company identity |
| CRDT | replaces DRAFT mutation/concurrency mechanism; SUBMIT still freezes one exact accepted generation/state |

---

# 18. Proof strategy before implementation

T2 acceptance requires a later implementation plan to include falsifiable proofs for at least:

```text
concurrent create/code allocation cannot produce duplicate committed Document code
stale autosave/update cannot overwrite newer WorkingContent
SUBMIT cannot freeze stale WorkingContent generation
route edit cannot create mixed in-flight snapshot
Group membership drift cannot rewrite an activated Step candidate set
RETURN/withdraw/cancel never mutate prior Submission/decision history
concurrent Release attempts cannot create two EFFECTIVE Revisions
replacement Release cannot expose predecessor+successor EFFECTIVE together
required Rendition mismatch cannot Release a different Submission
renderer/provider outage cannot make false EFFECTIVE truth
obsolescence cannot complete against a no-longer-current EFFECTIVE target
new Revision cannot race an active obsolescence attempt under accepted mutual-exclusion law
retry/restart cannot duplicate one semantic Release/decision result
```

T2 is design only; these are proof obligations, not claims that tests currently exist.

---

# 19. Explicit non-decisions

T2 does not decide:

```text
final SQL table/index/trigger names
exact lock SQL / all global lock ordering beyond T2 lifecycle roots
exact Role/Permission names/bundles
AuditEvent census/facts
object-store handles / staging keys / scan implementation
renderer/outbox/worker/retry topology
Search technology
API route/error envelope/idempotency key contract
frontend workflow screens
migration import transaction shapes
cross-Step strict SoD beyond the bounded initiator rule proposed here
fresh-auth/eSignature
quorum ALL/N-of-M
reassignment/overseer/SLA/escalation
separate obsolescence-specific route configuration
scheduled release
```

---

# 20. Reopen triggers

Reopen the implicated T2 decision only when material evidence shows:

- Launch truly needs `ALL`/N-of-M quorum rather than ANY-one Group Step;
- a real business route must select by product Role-in-Area rather than NamedUser/Group;
- regulation/customer contract requires strict cross-Step SoD, eSignature/fresh-auth or mandatory reauthentication;
- live reassign/overseer is required and withdraw→fix route→resubmit is operationally insufficient;
- obsolescence must use a distinct governance route;
- a concrete Change-Control journey needs an active replacement Revision and obsolescence intent to coexist;
- scheduled/future-dated effectivity becomes a real requirement;
- implementation evidence proves `READ COMMITTED` + explicit locks/CAS cannot enforce the accepted invariants without disproportionate complexity.

---

# 21. T2 operator adjudication packet

Recommended dispositions:

```text
T2-A ACCEPT — one local ACID transaction per native business transition; no external/provider call joins it.
T2-B ACCEPT — stable Document is lifecycle serialization root; WorkingContent uses OCC for DRAFT races.
T2-C ACCEPT — create atomically establishes code + Document + REV001 DRAFT + initial WorkingContent; template creation revalidates exact current EFFECTIVE source at commit.
T2-D ACCEPT — SUBMIT freezes exact expected WorkingContent generation + coherent governance/representation snapshots and moves Revision to SUBMITTED; NoHumanApproval creates no GovernanceAttempt.
T2-E ACCEPT — Launch route selector = NAMED_USER | GROUP only; no ROLE_IN_AREA baseline.
T2-F ACCEPT — Group Step = ANY-one from concrete enabled membership snapshot captured at Step activation; current AuthZ still rechecked at decision.
T2-G ACCEPT — one active sequential Step; ACCEPT advances; RETURN terminates attempt and preserves immutable history; no generic quorum/reassign/overseer engine.
T2-H ACCEPT — withdraw pre-Release returns same Revision to DRAFT and terminates the attempt without fake verdict; cancellation terminally cancels the Revision and preserves older EFFECTIVE/history.
T2-I ACCEPT — Release gates = human gate + optional official-rendition gate; system may Release in the same transaction as SUBMIT/final ACCEPT when all gates are already satisfied, otherwise truthful SUBMITTED state remains until the missing gate completes.
T2-J ACCEPT — replacement Release atomically sets predecessor SUPERSEDED + successor EFFECTIVE and never includes Distribution obligations in Launch-Core atomicity.
T2-K ACCEPT RECOMMENDED — bounded SoD only: Submission submitter / obsolescence initiator cannot satisfy a human Step on that same attempt; no baseline cross-Step same-user prohibition.
T2-L ACCEPT — obsolescence initiation requires current EFFECTIVE + reason + no open replacement Revision + no active obsolescence; active obsolescence blocks new Revision; same DocumentType governance route is reused; NoHumanApproval completes with zero human Step.
T2-M ACCEPT — route/config edits and attempt snapshotting are atomic whole-config operations; in-flight attempts never reinterpret after admin edits; no mandatory standalone PolicyVersion object.
T2-N ACCEPT — ordinary posture remains READ COMMITTED + explicit narrow serialization/CAS rather than global SERIALIZABLE; exact SQL enforcement waits for implementation design.
```

T2 remains **non-authoritative** until the operator adjudicates these recommendations. After adjudication, the mandatory platform-facing T2 summary must be presented and explicitly ratified before T2 can close or T3 open.

Implementation remains **BLOCKED**.