# MetalDocs — Post-T5 Fable Review — Author Adjudication Round 1

> **Status:** ACTIVE STAGING / AUTHOR ADJUDICATION — OPERATOR RATIFICATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Fable evidence:** `docs/superpowers/analysis/2026-08-18-t1-t5-integrated-independent-fable-review.md` @ `bdef5fc3c4004aa3ab4deefc9e8373dd3efcf856`  
> **Implementation:** BLOCKED  
> **T6:** NOT OPEN

This file is the architecture-author response to the independent Fable review. It is staging/evidence only. It does **not** amend T1→T5 authority by existing. Proposed bounded amendments require operator ratification before durable authorities or the Decision Registry are changed.

The review is accepted as a successful cold checkpoint: it found no systemic contradiction and no decision requiring a formal T-stage reopen. The material issues are corrections/completions at existing seams.

---

# 1. Executive adjudication

```text
Fable verdict                         = APPROVE T1→T5 WITH MATERIAL FIXES
Author disposition                    = ACCEPT WITH REFINEMENTS
Formal T-stage reopen                 = NONE
T6                                    = BLOCKED pending operator ratification + bounded amendments
Implementation                         = BLOCKED
```

Disposition summary:

```text
M1  ACCEPT — projection concurrency counterexample is valid.
M2  ACCEPT ROOT CAUSE / REFINE FIX — restore security non-resurrection is required;
     do not prematurely freeze a generic security-teardown journal.
M3  ACCEPT — choose option (b): materialized Search projection + search_refresh are conditional
     on T6 proving at least one derived/expensive searchable fact that canonical query/view cannot
     sustainably satisfy. Current Product Contract search facts do not prove materialization.

L1  ACCEPT / REFINE — title is governed Revision metadata, not mutable Document-current metadata.
L2  ACCEPT — late rendition result is a no-op for a terminated/ineligible Submission.
L3  ACCEPT — live admission claim/binding protects READY content from GC until bounded release/expiry.
L4  ACCEPT — bounded initiator withdrawal of active human-governed obsolescence request.
L5  ACCEPT — align T3 provider-disable wording with T5-L.

N1  ACCEPT as architecture note/reopen guard.
N2  ACCEPT cosmetic registry wording cleanup.
N3  ACCEPT — explicitly add source upload/admission UX to T6 REOPEN set.
```

Everything not named here remains frozen.

---

# 2. M1 — Search concurrent refresh overlap

## Disposition: ACCEPT

Fable's counterexample is valid. `reload latest state` solves out-of-order **start/order**, but not two overlapping executions where the older execution reads first and writes last.

The missing invariant is:

```text
IF a materialized Search projection exists:
  for one Document,
  projection-write serialization is acquired BEFORE canonical state is read
  and held through the projection rewrite/removal.

  full/rebuild writes obey the same per-Document serialization law.
```

Consequences:

```text
job FIFO remains unnecessary
broker/queue ordering remains unnecessary
duplicate delivery remains acceptable
concurrent overlap cannot leave an older canonical observation as final projection state
```

This is a narrow correctness law, not a new scheduler/lock platform. Exact SQL/lock primitive remains implementation design.

Because M3 below makes materialization conditional, this law is conditional too: if Launch Search is satisfied by a canonical query/view, there is no async projection write to serialize.

**Proposed durable amendment:** T5-G + conditional projection proof obligations; registry Search wording follows.

**Formal reopen:** NONE.

---

# 3. M2 — Restore security non-resurrection

## Disposition: ACCEPT ROOT CAUSE / REFINE FIX

Fable is correct that T4 currently proves exact-content restore and post-snapshot `UserProfile` erasure non-resurrection, but does not prevent a pre-offboarding restore point from reviving application sessions/access state.

### Accepted minimum security law

A restored deployment may not enter ordinary serving mode while historical authentication/authorization state can silently resurrect access that was revoked after the recovery point.

Two layers are required:

```text
A. ApplicationSession — structural fail-safe
   ALL ApplicationSessions restored from the recovery point are invalidated before serving.
   Users must establish fresh post-restore sessions.

B. User/access teardown — reconciliation gate
   before ordinary serving, post-snapshot security teardown known to the recovery process
   (at minimum User offboarding/disable and security-bearing access revocations that must survive
   the chosen recovery objective) is reconciled from independently retained recovery/control evidence.

   if completeness required by the recovery contract cannot be proven:
     ordinary authenticated serving remains BLOCKED / fail closed.
```

### Refinement versus Fable proposal

Do **not** freeze a generic journal of every RoleAssignment/GroupMembership mutation in T4 now. That would choose a recovery mechanism before T7/operations derives the smallest proof source.

T4 should own only the readiness invariant:

```text
restored sessions invalidated
+ required post-snapshot security teardown reconciled/proven
→ security restore gate may pass
```

T7/operations must choose the concrete recovery source/choreography together with the already-open erasure reconciliation design. Credible mechanisms may include a bounded independently retained security delta/barrier or a stricter fail-closed recovery procedure; T7 must prove completeness for the selected recovery objective.

This is intentionally distinct from normal point-in-time business-data recovery: security revocation/offboarding must not silently become usable authority merely because the database was rewound.

**Proposed durable amendments:** T4-M/T4-N or adjacent restore-readiness section; T7 REOPEN wording; registry STO/SEC recovery wording.

**Formal reopen:** NONE.

---

# 4. M3 — Search machinery before a named materialization consumer

## Disposition: ACCEPT — SELECT OPTION (b), CONDITIONALIZE

This is the strongest Method finding in the review.

Current Product Contract Search requires discovery/filtering by:

```text
code/title
Document Type
Area
responsible owner
status/current-effective truth
```

Those facts are canonical relational facts in the same PostgreSQL product database. They prove a Search **journey**, but do not by themselves prove a materialized projection or durable `search_refresh` job.

Therefore the corrected architecture should be:

```text
Search capability/journey = REQUIRED
Search authority boundary = CURRENT and unchanged

Launch baseline implementation:
  canonical PostgreSQL query/view over current semantic state

ONLY IF T6 names a derived/expensive searchable fact or measured query requirement that canonical
query/view cannot sustainably satisfy:
  activate rebuildable PostgreSQL materialized Search projection
  + transaction-coupled search_refresh(document_id)
  + conditional M1 per-Document projection-write serialization
  + full rebuild/reconciliation path
```

Examples of evidence that could activate materialization later:

```text
full-text search over extracted DOCX/PDF body text
expensive derived searchable facts that cannot be sustainably computed on read
measured scale/ranking/latency requirement requiring precomputation
```

Do **not** invent full-text content search now merely to justify the mechanism.

External Search engine remains rejected as Launch baseline absent measured need.

### Consequences for T5

Replace:

```text
always-required durable job = search_refresh
Search = mandatory materialized rebuildable projection
```

with:

```text
conditional durable job = search_refresh
  only when T6/current consumer proves materialized projection is required

Search baseline = canonical PostgreSQL query/view
materialized projection/rebuild = conditional optimization/derived projection, never authority
```

This removes M1's complexity entirely in the baseline case while preserving the exact seam if a real consumer appears.

**Proposed durable amendments:** T5-B/T5-C examples/T5-F→I/proof obligations; registry ASY-01/ASY-09 equivalent wording; router T6 Search work.

**Formal reopen:** NONE.

**Operator-visible choice:** YES — this adjudication recommends option (b). Operator ratification is required before changing durable T5 authority.

---

# 5. L1 — Document title ownership

## Disposition: ACCEPT / REFINE

The product contract requires `code/title` discovery but T1 never names title ownership. The smallest coherent law is **not** mutable Document-current title, because that would allow a title edit for a new DRAFT to leak ahead of Release while readers still consume the prior EFFECTIVE Revision.

Proposed law:

```text
stable Document identity:
  stable technical identity + stable controlled code

Document title:
  governed Revision metadata
  initialized on REV000 DRAFT
  editable only as part of the open DRAFT Revision
  frozen by Submission with the rest of decision-relevant governed state
  ordinary current-reader/search title comes from the current EFFECTIVE Revision
```

Therefore changing the official title of an EFFECTIVE Document is a business revision, not an out-of-band Document metadata edit.

Author/governance workspaces may show the authorized open-DRAFT/submitted title without changing ordinary reader truth.

No new semantic owner is created; Controlled Documents already owns Revision.

**Formal reopen:** NONE; T1 completeness amendment.

---

# 6. L2 — Late OfficialRendition result

## Disposition: ACCEPT

`eligible` must be explicit.

A renderer output may become semantic `OfficialRendition` only if, in the final serialized local transaction:

```text
Revision is still SUBMITTED / pre-Release
this exact Submission is still the live release candidate
its frozen policy still requires this rendition format
the attempt has not been RETURNED or WITHDRAWN
Revision has not been CANCELLED
no winning OfficialRendition already exists
T4 READY/exact-content/admission proofs still pass
```

If any condition fails:

```text
semantic finalization = NO-OP
rendered output remains non-governed mechanism content and may later be reclaimed
```

Do not permanently freeze a dead-attempt rendition solely because rendering finished late.

**Formal reopen:** NONE; T5-D/E precision amendment.

---

# 7. L3 — GC liveness for not-yet-referenced READY handles

## Disposition: ACCEPT

Absence of a semantic reference is not sufficient while a handle still has a live legal path to first attachment.

Proposed T4/T5 law:

```text
GC eligibility requires:
  not current WorkingContent
  no immutable governed reference
  no backup exclusion
  AND no live admission claim/binding/lease that may still legally attach the handle
```

The admission liveness mechanism must be bounded by explicit consumption/release and/or expiry so abandoned uploads do not become immortal.

Exact timeout/schema is implementation design.

**Formal reopen:** NONE.

---

# 8. L4 — Withdraw active obsolescence request

## Disposition: ACCEPT BOUNDED ESCAPE

Participant `RETURN` is not a sufficient universal escape because an in-flight frozen route can lose all usable participants while active obsolescence blocks new Revision creation.

Add the smallest symmetric escape:

```text
authorized initiator/manager may withdraw an active human-governed obsolescence request
before successful obsolescence completion

→ terminate GovernanceAttempt as WITHDRAWN
→ current target Revision remains EFFECTIVE
→ no fake ACCEPT/RETURN decision
→ immutable withdrawal actor/time/reason evidence + required Audit
→ new Revision creation becomes eligible again under normal rules
```

No generic cancellation/reassignment/workflow engine is introduced. `NoHumanApproval` obsolescence normally completes in its initiating transaction, so there is no intermediate attempt to withdraw.

Existing `document.obsolete` + scope/domain relationship should protect the action; no new permission is justified.

**Formal reopen:** NONE; bounded T1/T2/T3 journey completion.

---

# 9. L5 — T3 provider-disable wording

## Disposition: ACCEPT

T3 should explicitly defer to T5-L:

```text
Launch offboarding correctness is complete in MetalDocs after local disable/session/grant/membership teardown.
No provider-disable durable intent/job is baseline.
If a future assurance requirement mandates provider-side eventual disable, reopen T5-L and add one named effect.
```

This prevents a local reader from reintroducing `provider_subject_disable` by interpreting T3's conditional sentence as a current requirement.

**Formal reopen:** NONE.

---

# 10. Notes

## N1 — same-DB durable-work restore coherence — ACCEPT

Record as a positive design property/reopen guard:

```text
while required durable jobs live in the same PostgreSQL recovery domain as the semantic facts that
create them, a DB recovery point rewinds both together; restored pre-snapshot required jobs may replay
idempotently and post-snapshot jobs disappear with the post-snapshot facts that required them.
```

If a future job mechanism leaves the same recovery domain, restore consistency becomes an explicit reopen/proof obligation.

This does not solve M2 security non-resurrection; that is a separate readiness property.

## N2 — registry wording ambiguity — ACCEPT COSMETIC

Tighten rows whose disposition says `SUPERSEDED` while the meaning column describes the replacement current law. No decision changes.

## N3 — source upload/admission journey in T6 — ACCEPT

Explicitly add to T6 REOPEN set:

```text
source upload / managed-content admission / malware-preflight UX
```

T6 must encode the browser/client journey for T4 OPEN→READY/admission without moving exact-content authority into frontend/provider state.

---

# 11. Proposed minimal amendment set after operator ratification

```text
T1
  title = Revision-governed metadata
  bounded obsolescence-withdraw semantic evidence

T2
  bounded obsolescence initiator/manager withdrawal

T3
  authorization rule for obsolescence withdrawal using existing document.obsolete
  provider-disable wording aligned to T5-L

T4
  restore security non-resurrection readiness
  GC live-admission protection
  same-recovery-domain job coherence note

T5
  Search materialization/search_refresh conditional on proven T6/current consumer
  conditional per-Document projection-write serialization if projection exists
  late OfficialRendition finalization no-op rule

Decision Registry
  reflect bounded amendments and clarify ambiguous SUPERSEDED wording

Router / T6 REOPEN
  source upload/admission UX explicit
  Search journey first proves canonical-query vs derived-projection need
```

No other T1→T5 decision is touched.

---

# 12. Round-2 challenge contract for Fable

If Fable is asked to review this adjudication, it should read this file from GitHub and challenge only material disagreement.

In particular:

```text
M1 — is conditional per-Document projection-write serialization sufficient without FIFO?
M2 — does the refined readiness invariant close access resurrection without prematurely inventing a generic security journal?
M3 — does option (b) correctly satisfy the Product Contract while removing dormant Search machinery?
L1 — does Revision-owned title preserve draft/effective separation?
L4 — is bounded obsolescence withdrawal the smallest non-deadlocking escape?
```

If it agrees, no prose restatement is needed: return a short delta verdict and exact remaining disagreement set (ideally EMPTY).

If it disagrees, write a new GitHub evidence artifact with the smallest concrete counterexample. Do not alter durable authority directly.

---

# 13. Current gate

```text
Fable independent review                   RECEIVED
Author Round-1 adjudication                 WRITTEN / OPERATOR RATIFICATION PENDING
Durable T1→T5 amendments                    NOT YET APPLIED
Post-T5 checkpoint                          OPEN
T6                                          NOT OPEN
implementation                              BLOCKED
```
