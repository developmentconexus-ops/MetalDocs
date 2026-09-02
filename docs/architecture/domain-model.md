# R10-T1 — Semantic State & Invariants

> **Status:** ACTIVE / OPERATOR-RATIFIED TECHNICAL AUTHORITY  
> **Ratified baseline:** 2026-08-18  
> **T11 bounded reopen:** 2026-08-22 — stable-Document Discussion / `@mention` / Notifications; Lead GCR + Fable CONVERGED  
> **Revision convention:** `REV000` initial issuance / `REV001` first revision  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Implementation:** BLOCKED

This page records the current semantic state families. The T11 bounded reopen adds only the proven Discussion/Mention and Notifications families; all unrelated T1 semantics remain unchanged.

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
ConfidentialityClass vocabulary
ConfidentialityGrant
```

Role/Permission semantics are product-owned, not customer-defined platform data. RoleAssignment is current grant truth over `User | Group` and `Company | Area` scopes.

`ConfidentialityClass` is a Company-configured vocabulary whose authorization semantics are
product-owned, exactly like Role. Classes are additive and non-hierarchical. Exactly one
distinguished default class per Company means "unrestricted"; every Document carries exactly
one class at all times, so absence never has to be distinguished from "unknown".
`ConfidentialityGrant` is current clearance truth over the same `User | Group` and
`Company | Area` vocabularies. It is a second independent axis, never a Role and never a
Permission. Exact current roles, permissions, bundles and check sites belong T3. No role is a domain-governance bypass.

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
DocumentDiscussionMessage
Mention
```

The stable Document identity is not silently retitled in place. Human-readable title belongs to the Revision being governed. Therefore a newer DRAFT/SUBMITTED Revision may carry a new title while ordinary readers continue seeing the title of the current EFFECTIVE Revision. Historical Revisions preserve their own governed titles.

`DocumentDiscussionMessage` belongs to the stable Document across Revisions. It is not WorkingContent, DRAFT EditorialComment, SubmissionFeedback or governance feedback. Accepted Launch messages are immutable; a message may optionally reference one earlier message in the same Document Discussion. `Mention` binds a stable Organization `User` identity and never grants access.

A DiscussionMessage may snapshot the exact official Revision identity present at acceptance when one exists. That reference is contextual provenance only; Discussion ownership remains at the stable Document.

### Audit

```text
AuditEvent
```

Audit is append-only supporting semantic evidence and never current domain state. Deployment-wide `AuditChainHead` / global hash-chain serialization is not a Launch requirement absent a concrete assurance trigger.

### Notifications

```text
Notification
```

Notification is persistent recipient attention state about an already-valid source fact.

Current Launch source union:

```text
DOCUMENT_MENTION
  document_id
  message_id
```

Semantic minimum:

```text
notification_id
recipient_user_id
kind
closed source identity
created_at
seen_at?
read_at?
archived_at?
```

Binding laws:

```text
seen_at is monotonic
read_at absent/present = unread/read; read may return to unread
archived_at absent/present = active/archived; archive is reversible
READ => SEEN
archive/unarchive preserves read/seen
Notification read != Document read/acknowledgement/governance evidence
Notification never grants/preserves source access
source presentation is resolved under current disclosure, not copied as Notification authority
```

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

Document Discussion belongs to stable Document, never DRAFT/Submission/Governance authority
DiscussionMessage accepted state is immutable in Launch
Mention identity = stable User id, never reparsed display text
reply reference cannot cross Document Discussion
explicit accepted Mention requires one persistent Notification per unique target/message in the same local transaction
Notification persistence != current presentability
Notification presentability always rechecks current source disclosure
Notification engagement never becomes Read & Acknowledge
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
per-document ACL / access control entry
external / guest / public-link access subject
standalone Artifact semantic owner
DocumentTypeCategory taxonomy
generic Dictionary/System Value platform
TemplateSpec platform
DRAFT EditorialComment platform
message editing/version/tombstone platform
Notification email/push/preferences/digest state
Periodic Review state
Distribution / acknowledgement state
Dossier / Evidence / Records Governance
generic Interchange / governed export / repository receipt state
global AuditChainHead/hash chain
business WorkingSnapshot history
EditorSession as business authority
generic EventBus/EventStore state
```

Stable-Document Discussion and persistent in-app Mention Notifications are current Launch; their richer adjacent platforms remain absent.

## 6. Future-evolution anchors

```text
Notification channels → persistent Notification intent + named T5 durable delivery effect; channel never owns Notification truth
Distribution          → Release + effective Revision + User/Group
Periodic Review       → Document + current EFFECTIVE Revision
Dossier               → stable Document identity
Evidence              → Organization/AuthZ + future shared exact-content mechanism
Records               → stable governed identities + immutable lifecycle history
Governed Export       → stable semantic relationships + exact-content facts
Repository connector  → target-owner seams + exact-content snapshots
Training/LMS          → released/effective document + future Distribution
Change Control        → stable Document/Revision lifecycle seams
pooled tenancy        → stable Company identity + reopenable substrate
CRDT                  → replaceable WorkingContent concurrency mechanism
```

Binding law:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

T1 may reopen only on a material counterexample that invalidates a specific accepted invariant. Current bounded cross-layer details are routed by `docs/decisions/discussion-notifications-launch.md`.