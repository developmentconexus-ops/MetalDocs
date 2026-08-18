# R10 Technical Architecture — Rebaselined Active Stage Authority

> **Status:** ACTIVE — **T1 CLOSED / OPERATOR-RATIFIED; T2 DECISIONS ADJUDICATED / PLATFORM-SUMMARY RATIFICATION NEXT; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Revision-numbering amendment:** 2026-08-18 — **REV000 initial issuance / REV001 first revision**  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** `docs/engineering/standards/root-cause-global-maximum-method.md`  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`  
> **GCR authority:** `wiki/architecture/whole-product-alignment-review.md`  
> **Ownership authority:** `wiki/architecture/launch-v1-ownership-topology.md`  
> **Implementation gate:** **CLOSED — design/documentation only**

This page is the active technical-stage authority after the Whole-Product rebaseline. It supersedes the former R10-A/B1–B6/C–F stage order as active routing. Old R10 artifacts remain evidence only where the current Product Contract, GCR, 4+1 ownership topology and accepted T-stage conclusions preserve them.

---

## 1. Binding objective and evolution law

Derive the **smallest sustainable Launch V1 technical architecture** from:

```text
accepted Product Contract
+ operator-adjudicated Whole-Product GCR
+ operator-approved 4+1 ownership topology
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
T2 — Governance, Effectivity & Lifecycle Transactions        DECISIONS ACCEPTED / SUMMARY RATIFICATION PENDING
T3 — Authorization & Audit Enforcement                       NOT OPEN
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

Authentication provider subject identity, organizational User identity and product Authorization remain distinct. Fresh-auth/e-signature state is absent until a named T2/T3 consumer proves it.

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

## 4.3 T1-J — NoHumanApproval obsolescence

Accepted:

> If the Document Type is configured `NoHumanApproval`, governed obsolescence may complete with zero human Step after authorized initiation, mandatory reason and all eligibility/invariant checks.

This is **not** a raw status toggle and creates no fake System approver. Domain evidence remains mandatory; T3 later establishes required Audit.

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

## 4.5 T1 future-evolution proof

Stable attachment anchors are intentionally preserved:

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

T1 is durable architecture authority. Its completed staging packets are removed from the live tree; Git history is the archive.

---

# 5. T2 — Governance, Effectivity & Lifecycle Transactions — DECISIONS ACCEPTED / SUMMARY PENDING

T2 asks:

> **How do the accepted T1 facts change together under concurrency so every Launch journey is atomic, exact and unambiguous without creating generic workflow machinery?**

On 2026-08-18 the operator accepted T2-A→T2-N as recommended, with one bounded correction: initial document creation establishes `REV000 DRAFT`, not `REV001 DRAFT`. `REV001` is the first subsequent business revision. T3 remains closed until the required platform-facing T2 summary is explicitly ratified.

T2 covers:

```text
create / code allocation / REV000 DRAFT
blank and template-based creation
DRAFT WorkingContent mutation/autosave concurrency
SUBMIT exact generation
NoHumanApproval release behavior
sequential human governance
ACCEPT / RETURN_FOR_CHANGES
resubmit
withdraw Submission attempt
cancel open Revision
first Release of REV000
replacement Release / supersession (REV000 → REV001 for first revision)
required-Rendition release gate
governed obsolescence without replacement
```

Accepted T2 direction:

```text
one local ACID transaction per native business transition
Document = lifecycle serialization root
WorkingContent OCC for DRAFT races
create = code + Document + REV000 DRAFT + WorkingContent atomically
SUBMIT freezes exact expected generation + coherent governance/representation snapshots
route selector = NAMED_USER | GROUP
Group Step = ANY-one from activation membership snapshot
one active sequential Step
bounded initiator self-approval prohibition only
RETURN / withdraw / cancel preserve immutable Submission history
Release gates = human gate + optional OfficialRendition gate
system Release may occur in same tx when all gates are already satisfied
replacement Release = predecessor SUPERSEDED + successor EFFECTIVE atomically
Distribution remains outside Launch-Core Release atomicity
obsolescence requires current EFFECTIVE + reason + no open replacement + no competing obsolescence
same DocumentType route reused for obsolescence
NoHumanApproval obsolescence = zero human Step
route edits never reinterpret an in-flight attempt
READ COMMITTED + narrow explicit serialization/CAS posture
```

T2 owns:

```text
local transaction boundaries
serialization roots
OCC/concurrency law
state-transition eligibility
smallest participant-selection semantics
route snapshot timing
attempt activation/termination semantics
one-EFFECTIVE atomicity
obsolescence mutual-exclusion behavior
Release gate composition
```

T2 does not decide the exact permission catalog/Audit census (T3), storage locator/integrity implementation (T4), async worker topology (T5), API/frontend routes (T6), or historical migration execution (T7).

Active staging packets:

- `docs/superpowers/analysis/2026-08-18-r10-t2-governance-effectivity-transactions-candidate.md`
- `docs/superpowers/analysis/2026-08-18-r10-t2-operator-adjudication.md`

Current gate:

```text
T2 decisions ACCEPTED
→ platform-facing T2 summary
→ explicit operator summary ratification
→ T2 promotion/closure
→ T3
```

---

# 6. T3–T7 routing

## T3 — Authorization & Audit Enforcement — NOT OPEN

```text
personas → operations → resources → scopes → domain predicates
→ permissions → role bundles → check sites
```

Regenerates the Launch AuthZ catalog from zero and establishes the same-local-commit Audit census/minimum facts. Includes least-privilege Auditor/Governance Viewer. No role bypasses domain governance.

## T4 — Exact Content, Storage Integrity & Restore — NOT OPEN

Proves exact-content storage/integrity/restore without restoring Artifact semantic ownership: immutable governed bytes, mutable DRAFT recovery, hash/size/format validation, safe untrusted-content admission, no overwrite, provider outage truth, temporary cleanup, backup/restore and fail-closed restore.

## T5 — Durable Async, Search & External Effects — NOT OPEN

Designs only genuinely required durable intents/workers/retries, renderer execution, Search projection/rebuild/freshness, notifications and external effects. Worker/Search/provider state never becomes domain truth.

## T6 — Canonical API / Frontend Journeys — NOT OPEN

Proves admin/create/edit/submit/govern/return/withdraw/cancel/release/revise/obsolete/search/read/download/history/audit journeys. Search hit always re-resolves canonical state + Authorization + exact intended content before serving.

## T7 — Historical Migration & Cutover — NOT OPEN

Proves truthful native/imported handling, unknown preservation, ordinal/content provenance, idempotency/reconciliation, atomic semantic import units, cutover/readiness/rollback and legacy retirement. It may not recreate generic Interchange product scope.

---

# 7. Old R10 evidence classification

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

# 8. Proof and review discipline

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