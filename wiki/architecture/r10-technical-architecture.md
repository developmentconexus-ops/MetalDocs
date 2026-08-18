# R10 Technical Architecture — Rebaselined Active Stage Authority

> **Status:** ACTIVE — **T1 + T2 CLOSED / OPERATOR-RATIFIED; T3 AUTHORIZATION + AUDIT DISCOVERY ACTIVE; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Revision convention:** **REV000 initial issuance / REV001 first revision**  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** `docs/engineering/standards/root-cause-global-maximum-method.md`  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`  
> **GCR authority:** `wiki/architecture/whole-product-alignment-review.md`  
> **Ownership authority:** `wiki/architecture/launch-v1-ownership-topology.md`  
> **T2 authority:** `wiki/architecture/r10-t2-governance-effectivity-transactions.md`  
> **Implementation gate:** **CLOSED — design/documentation only**

This page is the active technical-stage authority after the Whole-Product rebaseline. It supersedes the former R10-A/B1–B6/C–F stage order as active routing. Old R10 artifacts remain evidence only where the current Product Contract, GCR, 4+1 ownership topology and accepted T-stage conclusions preserve them.

---

## 1. Binding objective and evolution law

Derive the **smallest sustainable Launch V1 technical architecture** from:

```text
accepted Product Contract
+ operator-adjudicated Whole-Product GCR
+ operator-approved 4+1 ownership topology
+ ratified T-stage conclusions
+ known-future evolution law
```

Every material decision must pass:

```text
Launch correctness
+ essential vs accidental complexity
+ one semantic authority per fact
+ proof strategy before implementation
+ named-future compatibility
```

Binding law:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

Known future direction is a counterexample set, not permission to prebuild modules, tables, jobs, permissions or generic frameworks.

---

## 2. Active semantic ownership baseline

```text
BUSINESS
Authentication
Organization
Authorization
Controlled Documents

SUPPORTING SEMANTIC
Audit
```

Not Launch semantic owners:

```text
storage / staging / byte integrity / malware inspection
render/view/editor providers
Search
async/outbox/jobs/retry/lease/DLQ
notifications
Historical Migration execution machinery
backup/restore transport/readiness
```

Historical `Artifact`, separate `Approval`, `Distribution`, `Documentary Context`, `Records Governance` and generic `Interchange` must not be resurrected by technical convenience.

---

# 3. Rebaselined R10 decomposition

```text
T1 — Semantic State & Invariants                              CLOSED / OPERATOR-RATIFIED
T2 — Governance, Effectivity & Lifecycle Transactions        CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit Enforcement                       ACTIVE / DISCOVERY-DESIGN
T4 — Exact Content, Storage Integrity & Restore              NOT OPEN
T5 — Durable Async, Search & External Effects                NOT OPEN
T6 — Canonical API / Frontend Journeys                       NOT OPEN
T7 — Historical Migration & Cutover                          NOT OPEN

→ Integrated Whole-R10 Global Coherence Review
→ cold independent review
→ operator final ratification
→ implementation spec/plan
→ code
```

The sequence is by failure class and proof dependency, not by old module/context boundaries.

## 3.0 Mandatory T-stage closure protocol — OPERATOR-APPROVED

For every `Tn`:

```text
Tn candidate/design
→ material decision adjudication
→ platform-facing summary
→ explicit operator summary ratification
→ promote durable Tn conclusions
→ remove completed staging
→ only then open Tn+1
```

The platform summary must explain what problem the stage solved, what was decided, how MetalDocs behaves because of it, what remains deferred, how the named future horizon remains attachable, and material reopen triggers. Technical A/B/C approval alone never opens the next stage.

---

# 4. T1 — Semantic State & Invariants — CLOSED / RATIFIED

The operator accepted T1-A→T1-I, accepted T1-J Option 1, then explicitly ratified the platform-facing T1 summary on 2026-08-18. The operator subsequently clarified the business revision convention: `REV000` is the initial issuance and `REV001` is the first revision after that issuance. This is a bounded semantic correction to T1 and the Product Contract, not a reopen of the T1 architecture.

## 4.1 Accepted semantic family set

### Authentication

```text
ProviderSubjectBinding
ApplicationSession
```

Authentication provider subject identity, organizational User identity and product Authorization remain distinct. Fresh-auth/e-signature state is absent until a named later consumer proves it.

### Organization

```text
Company
User
UserProfile
Area
Group
GroupMembership
```

`User` is stable historical participant identity; human-readable profile enrichment is separately erasable. Area and Group remain small flat organizational concepts. No Area hierarchy, dynamic/nested groups, provider-group mirroring or generic User↔Area membership exists without a consumer.

### Authorization

```text
product Role vocabulary
product Permission vocabulary
RoleAssignment
```

Role/Permission semantics are product-owned, not customer-defined platform data. The exact Launch catalog/bundles/check sites are T3. RoleAssignment remains current grant truth over User|Group and Company|Area scopes. No role is a domain-governance bypass.

### Controlled Documents

```text
DocumentType + numbering semantics
Document + Area/responsibility + Template role
DocumentOrigin
Revision
WorkingContent
Submission
current GovernanceRoute configuration
bounded GovernanceAttempt over SUBMISSION | OBSOLESCENCE
governance Step / Decision evidence
SubmissionFeedback
RevisionCancellation
Release
OfficialRendition only when required
Obsolescence request/result semantics
native/imported provenance seam
```

### Audit

```text
AuditEvent
```

Audit is append-only supporting semantic evidence, never current domain state. Deployment-wide `AuditChainHead`/global hash-chain serialization remains deferred absent a concrete Launch assurance requirement.

## 4.2 T1 lifecycle/content laws

```text
Document != Revision != WorkingContent != Submission
REV000 = initial issuance
REV001 = first revision after initial issuance
revision ordinals increment monotonically from zero and never reuse
WorkingContent = sole mutable DRAFT authority
Submission = immutable exact governed attempt
same-Revision resubmit = new Submission
Template = ordinary governed Document role
GovernanceAttempt is bounded to SUBMISSION|OBSOLESCENCE; not generic BPM
feedback/decisions never mutate Submission
withdraw governance attempt != cancel Revision
Release = sole normal native effectivity authority
replacement Release = predecessor SUPERSEDED + successor EFFECTIVE as one business transition
at most one EFFECTIVE Revision per Document
required OfficialRendition binds exact Submission
SourceOnly preview != semantic Rendition
obsolescence without replacement is explicit governed history
Search never establishes effectivity/access
storage/provider identity never becomes semantic content identity
native history != imported history
future contexts attach by reference rather than duplicate core authority
```

## 4.3 NoHumanApproval obsolescence

If the Document Type is configured `NoHumanApproval`, governed obsolescence may complete with zero human Step after authorized initiation, mandatory reason and all eligibility/invariant checks. This is not a raw status toggle and creates no fake System approver. Domain evidence remains mandatory; T3 establishes the required Audit census.

## 4.4 Explicitly absent from Launch T1

```text
standalone Artifact semantic owner
DocumentTypeCategory taxonomy
generic Dictionary/System Value platform
TemplateSpec platform
DRAFT EditorialComment platform
Periodic Review state
Distribution / acknowledgement state
Dossier / Evidence / Records Governance
generic Interchange / governed export / repository receipt state
global AuditChainHead/hash chain
business WorkingSnapshot history
EditorSession as business authority
```

These remain Launch+/Future or mechanism unless a later named Launch consumer proves otherwise.

## 4.5 T1 future-evolution anchors

```text
Distribution         → Release + effective Revision + User/Group
Periodic Review      → Document + current EFFECTIVE Revision
Dossier              → stable Document identity
Evidence             → Organization/AuthZ + future shared exact-content mechanism
Records              → stable governed identities + immutable lifecycle history
Governed Export      → stable semantic relationships + exact-content facts
Repository connector → target-owner seams + exact-content snapshots
Training/LMS         → released/effective document + future Distribution
Change Control       → stable Document/Revision lifecycle seams
pooled tenancy       → stable Company identity + reopenable substrate
CRDT                 → replaceable WorkingContent concurrency mechanism
```

T1 is durable architecture authority. Its completed staging is removed; Git history is the archive.

---

# 5. T2 — Governance, Effectivity & Lifecycle Transactions — CLOSED / RATIFIED

Durable detailed authority: `wiki/architecture/r10-t2-governance-effectivity-transactions.md`.

The operator accepted T2-A→T2-N with the `REV000` correction and explicitly ratified the platform-facing T2 summary on 2026-08-18.

## 5.1 Accepted transaction / concurrency laws

```text
one local ACID transaction per native business transition
no external/provider call joins local lifecycle atomicity
Document = lifecycle serialization root
WorkingContent = OCC/CAS for DRAFT races
READ COMMITTED + narrow explicit serialization/CAS posture
no silent last-write-wins governed DRAFT mutation
```

## 5.2 Create / SUBMIT

```text
create = code + Document + REV000 DRAFT + initial WorkingContent atomically
first later business revision = REV001
template-based create revalidates exact current EFFECTIVE template source at commit
SUBMIT freezes exact expected WorkingContent generation
SUBMIT freezes coherent governance + representation configuration snapshots
Revision DRAFT → SUBMITTED atomically with the immutable Submission
NoHumanApproval creates no fake GovernanceAttempt
```

## 5.3 Governance

```text
route selector vocabulary = NAMED_USER | GROUP
one active sequential Step at a time
GROUP Step = ANY-one from enabled membership snapshot captured at activation
current Authorization is rechecked when a User acts
Submission submitter cannot satisfy a human Step on that Submission attempt
obsolescence initiator cannot satisfy a human Step on that obsolescence attempt
no baseline cross-Step same-user prohibition
no baseline ALL/N-of-M quorum
no baseline ROLE_IN_AREA routing
no baseline reassign/overseer engine
```

If a frozen route becomes impossible, the Launch recovery is withdraw → fix current route → resubmit, preserving immutable history.

## 5.4 Return / withdraw / cancel

```text
RETURN preserves prior immutable Submission/decision/feedback history
Submission RETURN returns the same Revision to DRAFT
obsolescence RETURN leaves the target Revision EFFECTIVE
WITHDRAW terminates the current Submission attempt and returns the same Revision DRAFT
CANCEL terminally cancels the open Revision and never reuses its ordinal
older EFFECTIVE Revision remains EFFECTIVE after successor cancellation
```

No operation fabricates a participant verdict.

## 5.5 Release

Release gates are orthogonal:

```text
human gate:
  NoHumanApproval      → satisfied by absence
  UseGovernanceRoute   → satisfied by final ACCEPT

representation gate:
  SourceOnly                → satisfied by absence
  RequireOfficialRendition  → exact successful OfficialRendition required
```

System Release may occur in the same transaction as SUBMIT/final ACCEPT when all gates are already satisfied. If a required OfficialRendition is missing, the truthful state remains SUBMITTED until it exists and Release eligibility is revalidated.

Replacement Release is atomic:

```text
prior EFFECTIVE → SUPERSEDED
successor       → EFFECTIVE
```

For the first revision:

```text
REV000 → SUPERSEDED
REV001 → EFFECTIVE
```

Distribution/Acknowledgement is not part of Launch-Core Release atomicity.

## 5.6 Obsolescence

Obsolescence initiation requires:

```text
current EFFECTIVE target
mandatory reason
no open replacement Revision
no competing active obsolescence
```

While obsolescence is active, a new Revision cannot be created. The same DocumentType governance route is reused for obsolescence in Launch.

`NoHumanApproval` obsolescence may complete with zero human Step after authorization/reason/conflict checks. Human-governed obsolescence leaves the target EFFECTIVE until final ACCEPT. Successful completion changes the exact current EFFECTIVE Revision to OBSOLETE with no successor.

## 5.7 Configuration and restart laws

```text
route/config mutation is atomic as a whole
attempt snapshot is coherent old-or-new configuration, never mixed
in-flight attempts never reinterpret after admin edits
rollback exposes no successful partial transition
provider failure cannot retroactively change committed domain truth
retries/restarts cannot fabricate duplicate Release/decision results
```

## 5.8 T2 deferred/reopen surface

Deferred absent real requirement:

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

T2 is durable architecture authority. Its completed staging is removed; Git history is the archive.

---

# 6. T3 — Authorization & Audit Enforcement — ACTIVE / DISCOVERY-DESIGN

T3 asks:

> **Which current grants and domain predicates must permit each already-ratified Launch operation, at what Company/Area scope, and which operations require same-local-commit Audit evidence?**

Required derivation order:

```text
accepted personas + T1/T2 journeys
→ named operations
→ canonical resource owner/state predicates
→ scopes
→ permissions
→ product role bundles
→ administration law
→ check sites
→ same-local-commit Audit census
→ minimum bounded Audit facts
```

T3 must regenerate Launch Authorization **from zero**. The former exact `5×43` catalog is evidence/counterexample only and cannot be preserved by subtraction.

T3 must prove at least:

```text
least-privilege Reader path
Author / Document Owner path
Reviewer / Approver path requiring both grant + active Step participation
Governance Admin path
least-privilege Auditor / Governance Viewer path
no role/domain-governance bypass
User/Group + Company/Area RoleAssignment semantics
current grant re-evaluation at action time
historical actor attribution after offboarding/profile erasure
role-assignment administration without circular privilege
same-local-commit Audit for required governed/security mutations
Audit facts remain bounded/PII-minimized and never duplicate domain reasons/comments as authority
```

T3 does not own storage integrity (T4), Search/async (T5), public API/frontend composition (T6) or migration execution (T7).

Current gate:

```text
T3 discovery/design
→ compare credible access-model alternatives
→ present T3 material design choices to operator
→ operator adjudication
→ platform-facing T3 summary
→ explicit summary ratification
→ promote/close T3
→ T4
```

No T3 catalog is accepted merely because a prior role name existed.

---

# 7. T4–T7 routing

## T4 — Exact Content, Storage Integrity & Restore — NOT OPEN

Proves exact-content storage/integrity/restore without restoring Artifact semantic ownership: immutable governed bytes, mutable DRAFT recovery, hash/size/format validation, safe untrusted-content admission, no overwrite, provider outage truth, temporary cleanup, backup/restore and fail-closed restore.

## T5 — Durable Async, Search & External Effects — NOT OPEN

Designs only genuinely required durable intents/workers/retries, renderer execution, Search projection/rebuild/freshness, notifications and external effects. Worker/Search/provider state never becomes domain truth.

## T6 — Canonical API / Frontend Journeys — NOT OPEN

Proves admin/create/edit/submit/govern/return/withdraw/cancel/release/revise/obsolete/search/read/download/history/audit journeys. Search hit always re-resolves canonical state + Authorization + exact intended content before serving.

## T7 — Historical Migration & Cutover — NOT OPEN

Proves truthful native/imported handling, unknown preservation, ordinal/content provenance, idempotency/reconciliation, atomic semantic import units, cutover/readiness/rollback and legacy retirement. It may not recreate generic Interchange product scope.

---

# 8. Old R10 evidence classification

```text
former R10-A 8+3                  → SUPERSEDED FOR LAUNCH
former R10-B1                     → evidence for surviving substrate/transaction laws
former R10-B2                     → AuthN/Org/AuthZ evidence; exact catalog reopened
former R10-B3                     → core semantic evidence; Artifact/adjunct shape rejected
former R10-B4                     → one Step + Release evidence; workflow richness/Distribution reopened
former R10-B5                     → Future Dossier/Evidence/Records evidence only
former R10-B6                     → Audit/migration evidence; Interchange/hash-chain scope reopened
former R10-C                      → paused physical-safety evidence only; Artifact-owned solution rejected
former R10-D/E/F                   → old stage labels superseded by T5/T6/T7
```

Do not repair old candidate files into the target. Extract only surviving evidence into the active T-stage.

---

# 9. Proof and review discipline

Each T-stage must include:

1. authority/evidence boundary;
2. Known / Inferred / Unknown / Deferred;
3. root cause and target invariant;
4. credible alternatives;
5. smallest sustainable decision;
6. named future-horizon attack;
7. proof strategy before implementation;
8. reopen triggers;
9. explicit non-decisions;
10. material decision adjudication gate;
11. operator-facing platform summary;
12. explicit summary ratification before next-stage opening.

After T7:

```text
Integrated Whole-R10 GCR
→ cold independent review
→ operator final ratification
→ implementation spec/plan
→ code
```

Implementation remains **BLOCKED**.