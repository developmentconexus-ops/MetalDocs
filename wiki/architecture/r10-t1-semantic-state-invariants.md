# R10-T1 — Semantic State & Invariants

> **Status:** ACTIVE / OPERATOR-RATIFIED TECHNICAL AUTHORITY  
> **Ratified:** 2026-08-18  
> **Post-T5 Fable bounded amendment:** 2026-08-18 — Revision-governed title metadata  
> **Revision convention:** `REV000` initial issuance / `REV001` first revision  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`  
> **Ownership authority:** `wiki/architecture/launch-v1-ownership-topology.md`  
> **Implementation:** BLOCKED

This page records the operator-ratified T1 conclusions plus bounded completeness amendments ratified through the post-T5 independent-review checkpoint. No semantic owner or lifecycle state family is reopened.

## 1. Accepted semantic owners and families

### Authentication

```text
ProviderSubjectBinding
ApplicationSession
```

Provider subject identity, MetalDocs organizational User identity and product Authorization remain distinct. Provider roles/groups/organizations/permissions never become canonical MetalDocs Authorization. Fresh-auth/e-signature state is absent until a named consumer proves it.

### Organization

```text
Company
User
UserProfile
Area
Group
GroupMembership
```

`User` is stable historical participant identity. `UserProfile` is separately erasable human-readable/contact enrichment. Area and Group remain small flat organizational concepts. No Area hierarchy, dynamic/nested groups, provider-group mirroring or generic User↔Area membership exists without a named consumer.

### Authorization

```text
product Role vocabulary
product Permission vocabulary
RoleAssignment
```

Role/Permission semantics are product-owned, not customer-defined platform data. RoleAssignment is current grant truth over `User | Group` and `Company | Area` scopes. Exact Launch roles, permissions, bundles and check sites belong T3. No role is a domain-governance bypass.

### Controlled Documents

```text
DocumentType + numbering semantics
Document + stable code/type/Area/responsibility + Template role
DocumentOrigin
Revision + governed human-readable title metadata
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

The stable Document identity is not silently retitled in place. Human-readable title belongs to the Revision being governed. Therefore a newer DRAFT/SUBMITTED Revision may carry a new title while ordinary readers continue seeing the title of the current EFFECTIVE Revision. Historical Revisions preserve their own governed titles.

### Audit

```text
AuditEvent
```

Audit is append-only supporting semantic evidence and never current domain state. Deployment-wide `AuditChainHead` / global hash-chain serialization is not a Launch requirement absent a concrete assurance trigger.

## 2. Binding lifecycle/content laws

```text
Document != Revision != WorkingContent != Submission
REV000 = initial issuance
REV001 = first revision after initial issuance
revision ordinals increment monotonically from zero and never reuse
human-readable title = Revision-governed metadata
current reader title = title of current EFFECTIVE Revision
WorkingContent = sole mutable DRAFT authority
Submission = immutable exact governed attempt
same-Revision resubmit = new Submission
Template = ordinary governed Document role
GovernanceAttempt is bounded to SUBMISSION | OBSOLESCENCE; not generic BPM
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

## 3. `NoHumanApproval` obsolescence

If a Document Type is configured `NoHumanApproval`, governed obsolescence may complete with zero human Step after authorized initiation, mandatory reason and all eligibility/invariant checks.

This is not a raw status toggle and creates no fake System approver. Immutable domain evidence remains required; T3 establishes the required Audit census.

## 4. Template origin / imported-history seam

A derived Document origin pins:

```text
source Template Document
+ exact source EFFECTIVE Revision
+ bounded exact-content provenance
```

It does not require a native source Submission because truthful imported history may contain a real effective Revision/content without a native MetalDocs Submission.

Native history must never be fabricated to normalize imported history. T7 chooses the smallest concrete imported-history persistence forms from actual source evidence.

## 5. Explicitly absent from Launch T1

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

These are deferred/future/mechanism unless a named consumer later reopens them.

## 6. Future-evolution anchors

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

Binding law:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

T1 is closed and may reopen only on a material counterexample that invalidates a specific accepted invariant.