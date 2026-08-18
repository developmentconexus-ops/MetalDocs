# R10-T2 — Governance, Effectivity & Lifecycle Transactions

> **Status:** ACTIVE / OPERATOR-RATIFIED TECHNICAL AUTHORITY  
> **Ratified:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`  
> **Technical routing:** `wiki/architecture/r10-technical-architecture.md`  
> **Implementation:** BLOCKED

This page records the operator-ratified T2 architecture. T2 defines how the semantic facts accepted in T1 change together under concurrency and lifecycle transitions. It does not define final SQL/table/index syntax, package layout, concrete permission names, storage provider topology, worker topology, API routes, frontend screens or migration execution.

The operator accepted T2-A→T2-N and ratified the platform-facing T2 summary on 2026-08-18, with one bounded product correction: **`REV000` is the initial issuance and `REV001` is the first subsequent revision.**

---

## 1. Revision convention

```text
REV000 = initial issuance
REV001 = first revision after initial issuance
REV002 = second revision
...
```

Creation establishes `REV000 DRAFT`. First Release establishes `REV000 EFFECTIVE`. The first later business change cycle creates `REV001`.

Revision ordinals increment monotonically from zero and never reuse.

---

## 2. Core transaction law

> **One native MetalDocs business transition commits in one local ACID product-state transaction. External/provider calls never participate in that atomic commit.**

Therefore Keycloak, object storage, renderer, notification delivery and repository/provider calls cannot be required to succeed inside a local lifecycle transaction.

Rollback exposes no successful partial Submission, governance decision, Release, cancellation or obsolescence transition.

---

## 3. Lifecycle serialization and DRAFT concurrency

### Document serialization root

The stable `Document` is the lifecycle serialization root for operations capable of changing one Document's lifecycle/effectivity truth:

```text
create later Revision
SUBMIT
RETURN_FOR_CHANGES
WITHDRAW
CANCEL Revision
Step decision that advances/terminates governance
Release
obsolescence initiation/completion
```

This is a semantic serialization law, not final SQL lock syntax.

### WorkingContent OCC

DRAFT authoring uses one monotonic generation/CAS token:

```text
expected generation == current
→ accept mutation
→ increment generation once

expected generation != current
→ reject stale mutation/conflict
```

There is no silent last-write-wins for governed DRAFT content. Autosave never allocates a business Revision. Future CRDT/realtime collaboration may replace the DRAFT concurrency mechanism without changing Revision or Submission meaning.

### Isolation posture

Accepted design posture:

```text
PostgreSQL READ COMMITTED
+ narrow explicit lifecycle serialization
+ OCC/CAS
+ later structural constraints where required
```

Global `SERIALIZABLE` is not a Launch default. Reopen only if implementation evidence proves the accepted invariants cannot be enforced sustainably with the narrower posture.

---

## 4. Create

A successful blank create command atomically establishes:

```text
stable Document code
Document identity
REV000 DRAFT
initial WorkingContent
```

A successful operation never leaves an accepted partial Document shell.

Committed Document codes never rebind to another Document. Human numbering need not be gap-free; uniqueness and no reuse are the invariant.

### Template-based creation

Create-from-template additionally revalidates at commit that:

```text
source Document still has Template role/eligibility
selected source Revision is still the current EFFECTIVE Revision
exact source-content provenance is pinned
```

The new Document receives an independent WorkingContent seed. Later Template changes never rebind the derived Document.

---

## 5. SUBMIT

SUBMIT crosses the lifecycle boundary and therefore revalidates under Document serialization:

```text
Revision is current open DRAFT
caller expected WorkingContent generation is still current
all Launch submit requirements are satisfied
DocumentType governance + representation configuration is read as one coherent committed configuration
```

One successful SUBMIT atomically:

1. freezes an immutable Submission from exactly the accepted WorkingContent generation;
2. freezes the governance mode/route snapshot for that attempt;
3. freezes the official-representation requirement;
4. changes the Revision from `DRAFT` to `SUBMITTED`;
5. creates a `GovernanceAttempt` only for `UseGovernanceRoute`.

`NoHumanApproval` creates no fake GovernanceAttempt or System approver.

Later configuration edits affect later attempts only.

---

## 6. Governance route

### Launch actor selectors

The Launch route admits only:

```text
NAMED_USER
GROUP
```

`ROLE_IN_AREA` is not a baseline routing selector. Authorization roles are access bundles and do not automatically become business-routing identity.

### Sequential execution

Exactly one Step is active at a time for a live human GovernanceAttempt.

At activation:

- `NAMED_USER` resolves the configured User;
- `GROUP` resolves a concrete snapshot of the current enabled Group members;
- later Group membership drift does not rewrite the active Step candidate set;
- current Authorization is still rechecked when a User acts, because snapshot membership is not an access grant.

### Group Step completion

A Group Step is `ANY-one` in Launch:

> one currently authorized User from the activation snapshot may make the Step decision.

There is no baseline `ALL`, N-of-M or generic quorum engine.

### Bounded separation of duties

The minimum Launch independence rule is:

```text
Submission submitter cannot satisfy a human Step on that Submission attempt.
Obsolescence initiator cannot satisfy a human Step on that obsolescence attempt.
```

Launch does not prohibit the same non-initiating User from satisfying two different Steps when the configured selectors legitimately include that User. Cross-Step SoD remains a reopen trigger for a real regulatory/customer requirement.

### No baseline reassign/overseer engine

Launch has no generic live reassignment, delegation or overseer mechanism.

If a frozen route becomes impossible, the bounded recovery is:

```text
withdraw the active Submission attempt
→ same Revision returns DRAFT
→ fix current route configuration
→ resubmit
→ new immutable Submission/attempt snapshot
```

Reassignment may be introduced later if real operations prove this insufficient.

---

## 7. Governance outcomes

### ACCEPT

An ACCEPT decision binds the exact active GovernanceAttempt + Step + governed subject and records the actual User and trusted decision time as immutable domain evidence.

If more Steps remain, the next Step activates. Final ACCEPT satisfies the human governance gate.

### RETURN_FOR_CHANGES — Submission

One transaction:

```text
record immutable RETURN decision/reason
terminate current GovernanceAttempt as returned
Revision SUBMITTED → DRAFT
preserve old Submission + prior decisions/feedback
```

A later SUBMIT creates a new Submission and, where required, a new GovernanceAttempt.

### RETURN — Obsolescence

RETURN of an obsolescence request does not put the Document into DRAFT. It terminates that obsolescence attempt while the target Revision remains EFFECTIVE. A later retry is a new request/attempt.

---

## 8. Withdraw versus cancellation

### Withdraw

Before Release, an authorized author may withdraw an active Submission attempt to continue editing the same Revision:

```text
active attempt → terminated/WITHDRAWN
Revision SUBMITTED → DRAFT
old Submission + decisions/feedback remain immutable
```

Withdraw fabricates no ACCEPT/RETURN verdict.

A governance-satisfied Submission that is only waiting for a required OfficialRendition is still pre-Release and may be withdrawn because Release is the effectivity boundary.

### Cancel Revision

An authorized actor may cancel the current open Revision while it is DRAFT or pre-Release SUBMITTED:

```text
terminate any live governance attempt without fake participant verdict
write immutable RevisionCancellation reason/actor/time
Revision → CANCELLED
```

An older EFFECTIVE Revision remains EFFECTIVE. A cancelled Revision never reopens and its ordinal never reuses.

---

## 9. Release / effectivity

Release is system-owned and binds the exact winning Submission.

Release gates are orthogonal:

```text
human gate:
  NoHumanApproval     → satisfied by absence
  UseGovernanceRoute  → satisfied after final ACCEPT

representation gate:
  SourceOnly                 → satisfied by absence
  RequireOfficialRendition    → satisfied by exact successful OfficialRendition for the Submission
```

### Same-transaction Release

If all required gates are synchronously satisfied, Release may occur in the same local business transaction that satisfied the last gate.

Example:

```text
REV000 DRAFT
NoHumanApproval + SourceOnly
→ freeze Submission
→ system Release
→ REV000 EFFECTIVE
```

Submission and Release remain distinct semantic facts even when committed together.

Final human ACCEPT may likewise perform Release in the same transaction if the representation gate is already satisfied.

### Waiting for OfficialRendition

If human governance is complete but the required OfficialRendition is not yet available:

```text
Governance = satisfied
Revision = SUBMITTED
Release = not yet established
```

Later rendition success may trigger Release only after canonical eligibility is rechecked. Renderer/provider execution is outside the business transaction.

### First Release

```text
REV000 SUBMITTED → EFFECTIVE
Release binds the exact winning Submission
```

### Replacement Release

Serialized on the stable Document and atomic:

```text
prior current EFFECTIVE Revision → SUPERSEDED
winning successor SUBMITTED      → EFFECTIVE
Release records predecessor + winning Submission
```

For the first revision this is:

```text
REV000 → SUPERSEDED
REV001 → EFFECTIVE
```

No successful externally observable state may contain two EFFECTIVE Revisions for one Document.

Distribution/Acknowledgement is not part of Launch-Core Release atomicity.

---

## 10. Governed obsolescence without replacement

An obsolescence request may start only when:

```text
target Revision is the current EFFECTIVE Revision
mandatory reason is nonblank
no open replacement Revision exists
no other active obsolescence attempt exists for the Document
```

While an obsolescence attempt is active, creation of a new Revision is blocked. Launch does not simultaneously model “replace this document” and “remove it without replacement” intents.

The same DocumentType governance route/configuration is reused for obsolescence in Launch; there is no second obsolescence-specific workflow family.

### NoHumanApproval

```text
authorized initiation
+ reason
+ eligibility/conflict checks
+ NoHumanApproval
→ zero human Step
→ no fake approver
→ same transaction changes current EFFECTIVE Revision to OBSOLETE
```

Immutable obsolescence domain evidence remains required.

### Human-governed obsolescence

The current EFFECTIVE Revision remains EFFECTIVE while the route is active.

Final ACCEPT revalidates the canonical target/conflict conditions under Document serialization and then atomically:

```text
current EFFECTIVE Revision → OBSOLETE
no successor becomes EFFECTIVE
immutable obsolescence result/time is written
```

RETURN ends that request and leaves the Revision EFFECTIVE. There is no Launch reactivation of an OBSOLETE Document.

---

## 11. Configuration concurrency

Current DocumentType governance/representation configuration is mutable current truth. Each Submission/obsolescence attempt freezes one coherent configuration snapshot.

Required law:

```text
configuration mutation = atomic whole-configuration mutation
attempt creation        = atomic coherent snapshot
```

Concurrent admin edit and attempt creation resolves to the complete old or complete new configuration, never a mixed route.

No standalone browsable `PolicyVersion` aggregate is required by Launch.

---

## 12. Failure / restart laws

- rollback leaves no successful partial lifecycle transition;
- external provider failure cannot retroactively invalidate committed semantic history;
- a governance-satisfied Submission waiting for OfficialRendition is truthful durable state and resumes safely after restart;
- repeated system Release triggers must be idempotent against canonical eligibility/fact identity;
- transport retries may not fabricate duplicate semantic decisions;
- stale commands fail/reload instead of overwriting newer lifecycle truth.

Public request-idempotency shape belongs to T6; T2 defines the semantic uniqueness requirements.

---

## 13. Future-evolution compatibility

```text
Distribution        → reacts after Release; does not enter effectivity atomicity
Periodic Review     → later serializes against exact current EFFECTIVE Revision
Dossier             → references stable Document; owns no lifecycle transition
Evidence            → may own an independent lifecycle
Records             → attaches to immutable Revision/Submission/Release history
Governed Export     → reads stable consistent facts without owning them
Repository connector→ external provider effects remain outside lifecycle transaction
Training/LMS        → consumes released/effective truth; is not a Release gate
Change Control      → may orchestrate multiple Documents without taking each Document lifecycle authority
pooled tenancy      → may reopen substrate around stable Company identity
CRDT                → may replace DRAFT concurrency; SUBMIT still freezes one exact state
```

---

## 14. Deferred from T2

```text
T3 → exact permission catalog, role bundles, check sites, Audit census/minimum facts
T4 → content locator, staging, integrity, malware, restore
T5 → renderer/outbox/worker/retry/Search/notifications
T6 → API/frontend/idempotency contract
T7 → historical migration/cutover transaction shapes
```

Also deferred absent a named requirement:

```text
ALL/N-of-M quorum
ROLE_IN_AREA routing
cross-Step strict SoD
fresh-auth/eSignature
live reassign/overseer
SLA/escalation
separate obsolescence route
scheduled/future-dated effectivity
```

---

## 15. Reopen triggers

Reopen only the implicated decision when material evidence proves one of:

- Launch needs `ALL`/N-of-M quorum;
- a real route must select by product Role-in-Area;
- regulation/customer contract requires strict cross-Step SoD or fresh-auth/eSignature;
- live reassignment is operationally required;
- obsolescence needs a distinct route;
- a concrete Change-Control journey needs replacement and obsolescence intents to coexist;
- scheduled effectivity becomes a real requirement;
- implementation evidence shows READ COMMITTED + narrow locks/CAS cannot enforce the accepted invariants sustainably.

---

## 16. Proof obligations for later implementation

Later implementation planning must make these falsifiable:

```text
concurrent code allocation cannot create duplicate committed Document codes
stale DRAFT update cannot overwrite newer WorkingContent
SUBMIT cannot freeze a stale generation
route edit cannot create a mixed in-flight snapshot
Group membership drift cannot rewrite an activated Step candidate set
RETURN/withdraw/cancel cannot mutate prior immutable history
concurrent Releases cannot create two EFFECTIVE Revisions
replacement Release cannot expose predecessor+successor EFFECTIVE together
wrong/mismatched OfficialRendition cannot Release a different Submission
provider outage cannot create false EFFECTIVE truth
obsolescence cannot complete against a no-longer-current EFFECTIVE target
new Revision cannot race active obsolescence under the accepted mutual-exclusion law
retry/restart cannot duplicate one semantic Release/decision result
```

T2 is architecture authority only. Implementation remains **BLOCKED** until T3→T7, integrated Whole-R10 GCR, cold review and final operator ratification close.