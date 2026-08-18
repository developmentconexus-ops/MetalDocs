# T1→T5 Integrated Architecture — Independent Fable Review Request

> **Status:** ACTIVE STAGING / INDEPENDENT COLD-REVIEW REQUEST — NOT TARGET AUTHORITY  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Checkpoint:** T1→T5 operator-ratified; T6 NOT OPEN  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This packet requests an independent adversarial review of the complete ratified T1→T5 technical architecture **before T6 Canonical API / Frontend Journeys opens**.

This file is review staging only. It does not amend product or technical authority. Any finding is evidence and must be adjudicated against the Method before any durable authority changes.

---

## 0. Reviewer bootstrap — true cold start

**Reviewer role:** senior independent adversarial architecture reviewer.

Reconstruct MetalDocs exclusively from the repository. Do not use prior conversation memory, author explanations outside the repository, old runtime shape, sunk cost or this packet as requirement authority.

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
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
14. this review packet
15. historical/current implementation evidence only when needed to falsify or validate a specific material claim

Do not treat old R3–R9.5, old R10-A→C, current schema/OpenAPI/packages or removed T-stage staging as target authority.

### Required Method posture

Apply directly:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ Credible Alternatives
→ Local Maximum vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / Overengineering / Future Cost
→ Authority / Boundary
→ Enforcement
→ Proof Strategy
→ Adversarial Challenge
→ Decision
→ Reopen Triggers
```

Apply Structural Inversion aggressively:

> If the current implementation were opposite in every relevant respect, would the ratified T1→T5 conclusion still follow from product authority?

Do not reward compatibility with current code. Do not demand redesign merely because another design is aesthetically attractive.

---

## 1. Stage gate under review

The checkpoint to reconstruct and challenge is:

```text
Product Contract                                   ACTIVE / OPERATOR-APPROVED
Whole-Product GCR A1–A10                           CLOSED / OPERATOR-APPROVED
Launch ownership topology                          CLOSED / OPERATOR-APPROVED / 4+1
T1 — Semantic State & Invariants                  CLOSED / OPERATOR-RATIFIED
T2 — Governance, Effectivity & Transactions       CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit Enforcement            CLOSED / OPERATOR-RATIFIED
T4 — Exact Content, Storage Integrity & Restore   CLOSED / OPERATOR-RATIFIED
T5 — Durable Async, Search & External Effects     CLOSED / OPERATOR-RATIFIED
Decision Registry                                  CURRENT / OPERATOR-RATIFIED
T6                                                 NOT OPEN
T7                                                 NOT OPEN
implementation                                     BLOCKED
```

The purpose of this checkpoint is to find **cross-stage contradictions or missing seams before API/UX begins encoding them**.

A finding that merely prefers a different mechanism is not sufficient to reopen a ratified decision.

---

## 2. Current product/ownership anchor

Launch semantic owners are exactly:

```text
BUSINESS
Authentication
Organization
Authorization
Controlled Documents

SUPPORTING SEMANTIC
Audit
```

Mechanisms/projections are not semantic owners:

```text
managed content/storage/malware
viewer/editor/render providers
Search
async jobs/queues/retry
notifications
Historical Migration execution
backup/restore transport/readiness
```

Future law:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

Revalidation law:

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

---

## 3. Ratified T1→T5 kernel to attack

### T1 — Semantic state

At minimum:

```text
Document != Revision != WorkingContent != Submission
REV000 = initial issuance
one open Revision max per Document
one EFFECTIVE Revision max per Document
WorkingContent = sole mutable DRAFT authority
Submission = immutable exact attempt
same Revision may have multiple Submissions
GovernanceRoute current config; GovernanceAttempt frozen snapshot
NoHumanApproval creates no fake approver
Release = sole normal effectivity authority
explicit governed obsolescence
AuditEvent = action evidence, not current state
native/imported provenance remains distinguishable
```

### T2 — Transaction/lifecycle rules

At minimum:

```text
one local ACID transaction per native business transition
provider calls never join semantic transaction
Document = lifecycle serialization root
WorkingContent OCC/CAS for DRAFT
SUBMIT freezes exact generation + governed snapshot
one active sequential Step
GROUP activation freezes concrete candidate Users
current AuthZ rechecked at action
bounded self-approval prohibition
RETURN/withdraw preserve immutable Submission and return same Revision to DRAFT
cancel Revision terminal/non-reused ordinal
Release gates human + required representation
replacement Release atomically SUPERSEDES predecessor + EFFECTIVE successor
obsolescence has explicit guarded journey
READ COMMITTED + narrow row locks/CAS; no global SERIALIZABLE
```

### T3 — Authorization/Audit

At minimum:

```text
RoleAssignment subject = User | Group
scope = Company | Area
static 6-role / 15-permission vocabulary
additive grants + default deny
provider/JWT/session never durable permission authority
ordinary author working authority bounded by responsible-owner relation unless document.owner.manage
approval verdict requires governance.act + exact active-Step candidate participation + T2 predicates
offboarding atomically disables User/revokes sessions/removes memberships/direct grants
same-local-commit Audit for bounded critical census
Audit PII-minimized with frozen Company|Area visibility attribution
ordinary reads/downloads/search/autosave/login/logout are not mandatory semantic Audit
```

### T4 — Exact content/storage/restore

At minimum:

```text
ExactContentDescriptor = SHA-256 + exact size + ContentFormat
opaque managed_content_id = retrieval mechanism only
no Artifact semantic owner
one provider-neutral ManagedContentStore / one active store
OPEN→READY server-verified admission
create-once/no-overwrite
UNTRUSTED_EXTERNAL CLEAN malware proof before governed admission
WorkingContent = DRAFT recovery baseline; no WorkingSnapshot business history
SUBMIT/Rendition transaction revalidates READY exact content with zero provider/scanner calls
only unreferenced/non-governed content reclaimable
backup couples DB recovery point + all required exact content
restore remains non-serving until exact content verifies
historical restore reconciles post-snapshot lawful UserProfile erasures
```

### T5 — Async/Search/external effects

At minimum:

```text
one PostgreSQL-backed durable-job mechanism; River selected/reference mechanism
required future job inserted in same local transaction that creates requirement
provider/renderer execution outside semantic transaction
search_refresh(document_id) = always-required durable projection job
official_rendition_render = conditional only when frozen representation policy requires it
PDF source direct-view by default
DOCX + SourceOnly direct read-only viewer; no persisted PDF merely for viewing
OfficialRendition render/finalization at-least-once + idempotent/revalidating
Search = PostgreSQL rebuildable projection keyed by Document
Search worker reloads latest canonical state; duplicates/out-of-order converge safely
Search may lag by omission but never grants stale authority/effectivity
full Search rebuild mandatory; always-on crawler not baseline
GC = periodic reconciliation over GC_PENDING with immediate canonical recheck
no mandatory Launch notifications/event bus
no mandatory durable IdP-disable job
no generic ExternalEffectReceipt
bounded retry + terminal visibility + redrive + bounded-ID job payloads
minimum async operational visibility required
```

---

## 4. Mandatory cross-stage attacks

The reviewer MUST explicitly test these integrated flows rather than reviewing each T in isolation.

### A. DRAFT → SUBMIT → governance → optional Rendition → Release → Search

Challenge:

```text
WorkingContent generation
→ Submission exactness
→ governance snapshot/Step candidacy
→ optional conditional OfficialRendition
→ Release atomicity
→ Search convergence
```

Look for duplicate authority, impossible intermediate state, lost required durable work, stale approval target, renderer race, Release race or Search stale-authority leak.

### B. RETURN / withdraw / resubmit

Challenge whether immutable old Submission + same Revision DRAFT + new Submission + prior rendering/search jobs can coexist without old async work becoming semantic authority or corrupting later truth.

### C. Offboarding during governance/async work

Challenge total ordering among:

```text
User disable
session revocation
group/grant removal
active Step decision
Submission/withdraw/cancel/obsolescence
long-lived jobs that reload canonical state
```

Ensure no job payload or cached identity silently restores authority.

### D. Replacement Release and Search ordering

Attack:

```text
REV001 EFFECTIVE
REV002 Release
search refresh jobs duplicated/reordered
```

Prove the latest-state projector cannot re-expose predecessor EFFECTIVE truth after replacement.

### E. Obsolescence and Search

Attack successful obsolescence with stale Search hits. Verify stale projection can never serve obsolete content as current effective truth.

### F. Managed-content GC vs WorkingContent / Submission / Rendition / backup

Attack races between:

```text
DRAFT autosave replacement
SUBMIT freeze
rendition output
GC_PENDING
backup pin/capture
physical delete
restore
```

Look for a handle that can become required after GC eligibility, or a stale worker that can delete newly required content.

### G. Restore + offboarding/privacy

Challenge whether restored User/Profile/Session/RoleAssignment/GroupMembership and managed content can resurrect access or erased PII despite T3/T4 laws.

### H. Viewer vs OfficialRendition

Attack the split:

```text
viewer/preview mechanism
!=
OfficialRendition semantic state
```

Ensure a SourceOnly DOCX viewer cannot accidentally become release authority and an OfficialRendition policy cannot silently fall back to preview output.

### I. Audit vs async

Challenge whether business mutation + required Audit + required durable job can compose in one local transaction without Audit becoming event bus or jobs becoming evidence authority.

### J. Search/AuthZ boundary

Attack Area changes, Group membership changes, RoleAssignment removal, User offboarding and Document lifecycle changes while projection is stale. A Search row must never act as canonical authorization/effectivity.

---

## 5. Authority completeness / uniqueness audit

Perform a fact-owner sweep across T1→T5.

For every durable semantic fact, identify exactly one authority.

Explicitly challenge whether any of these are ownerless or duplicated:

```text
representation policy snapshot
active GovernanceAttempt / Step candidate snapshots
responsible Document owner
Submission exact descriptor
OfficialRendition exact descriptor
Release/effectivity
obsolescence result
Audit visibility snapshot
managed-content admission/GC mechanism state
Search projection state
job state / terminal failure state
restore erasure barrier/journal mechanism
```

Expected distinction:

```text
business semantic fact → semantic owner
mechanism/projection/operations fact → durable mechanism, never promoted to business authority
```

Flag any mechanism that has accidentally become a fifth business owner.

---

## 6. Decision Registry audit

Audit `wiki/architecture/rebaseline-decision-registry.md` for:

```text
T1→T5 decisions still incorrectly marked REOPEN
SUPERSEDED decisions leaked back into CURRENT/PRESERVE
DEFERRED capabilities creating Launch backward pressure
old Artifact/Approval/Tenant/RLS/Distribution/Records/Interchange assumptions surviving indirectly
contradictions between detailed T authorities and registry wording
```

The registry must route T6 only to its genuine REOPEN set.

---

## 7. Essential vs accidental complexity attack

For every retained mechanism, name its current Launch consumer.

Specifically challenge whether these are justified or overengineered:

```text
ApplicationSession
Group-mediated RoleAssignments
GovernanceAttempt snapshots
ExactContentDescriptor
OPEN→READY managed-content state
malware gate
admission binding
backup manifest / GC exclusion
post-snapshot erasure reconciliation
River durable jobs
Search projection
full Search rebuild
async terminal visibility
```

Also challenge the opposite failure: whether removing any of them would accept a concrete failure mode already prohibited by product authority.

Do not classify complexity as accidental merely because it is technical. Do not classify it as essential merely because it already exists.

---

## 8. Future seam attack

Test T1→T5 against known future horizons without requiring those capabilities now:

```text
Launch+ Distribution / Read & Acknowledge
Launch+ Periodic Review
Dossier
Evidence
Retention / Legal Hold / Disposition
Governed Export
External Repository IMPORT/PUBLISH
Training/LMS
multi-document Change Control
pooled multi-customer tenancy
CRDT/realtime coauthoring
```

For each future capability ask:

1. Can it attach by stable reference without duplicating current authority?
2. Does any current decision make the future capability materially harder than necessary?
3. Would fixing that later require rewriting immutable historical meaning?
4. Is the candidate accidentally implementing future state today?

The correct result is a preserved seam, not a dormant implementation.

---

## 9. T6 readiness attack

The review must decide whether T1→T5 is coherent enough that T6 can safely encode user/API journeys.

T6 should not begin if a material unresolved contradiction exists in:

```text
semantic identity/state
transaction boundaries
AuthZ check sites
Audit obligations
exact-content ownership
viewer/OfficialRendition distinction
durable-effect guarantees
Search authority boundary
```

Do **not** block T6 merely because final table names, API routes, UI screens, queue names or renderer product are intentionally not decided yet.

---

## 10. Finding format — mandatory

For each finding provide:

```text
ID / severity
claim
repository evidence anchors
authority/invariant affected
root cause
credible alternatives
Global Maximum analysis
essential vs accidental complexity
required correction
whether any ratified T-stage must reopen
smallest exact reopen set
what remains frozen
proof required to close
```

Severity:

```text
BLOCKER = current architecture cannot truthfully proceed to T6
MAJOR   = material contradiction/owner/race/future-cost defect; smallest correction required before T6
LOW     = precision/clarity/proof issue that does not invalidate current architecture
NOTE    = useful observation; no required correction
```

A request to reopen T1→T5 MUST provide:

1. material new evidence or integrated counterexample;
2. exact ratified invariant/decision invalidated;
3. why existing seam cannot solve it;
4. smallest decision IDs/stage sections to reopen;
5. everything that remains frozen.

---

## 11. Required final verdict

Return exactly one top-level verdict:

```text
APPROVE T1→T5 CHECKPOINT / T6 MAY OPEN

APPROVE T1→T5 WITH MATERIAL FIXES
  → list exact fixes
  → T6 remains blocked until adjudicated

REOPEN MINIMAL T1→T5 SET
  → list exact stage decisions/sections only
  → everything else remains frozen

BLOCK T1→T5 CHECKPOINT
  → only if architecture has a systemic contradiction that cannot be corrected by a bounded reopen
```

Also provide:

```text
finding count by severity
minimal reopen set
anti-overengineering verdict
future-seam verdict
T6 readiness verdict
```

Do not write implementation code or an implementation plan.

---

## 12. Independence law

The review itself is **evidence, not authority**.

After Fable returns:

```text
independent findings
→ operator/author adjudication against repo authority + Method
→ bounded corrections/reopen only where justified
→ optional delta review if material fixes were required
→ explicit checkpoint closure
→ only then T6 may open
```

Implementation remains **BLOCKED**.
