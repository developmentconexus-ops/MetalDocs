# R10-T3 — Authorization & Audit Enforcement

> **Status:** ACTIVE / OPERATOR-RATIFIED TECHNICAL AUTHORITY  
> **Ratified:** 2026-08-18  
> **Post-T5 Fable bounded amendment:** 2026-08-18 — obsolescence-withdraw AuthZ + provider-disable alignment  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **T1 authority:** `wiki/architecture/r10-t1-semantic-state-invariants.md`  
> **T2 authority:** `wiki/architecture/r10-t2-governance-effectivity-transactions.md`  
> **Implementation:** BLOCKED

This page records the operator-ratified T3 architecture plus bounded completeness amendments ratified through the post-T5 independent-review checkpoint. T3 defines Launch Authorization and semantic Audit enforcement over the already-ratified Organization, Controlled Documents and lifecycle model. It does not define final SQL/index syntax, package layout, API routes, frontend screens, storage, worker topology or migration execution.

The operator accepted T3-A→T3-P and explicitly ratified the platform-facing T3 summary on 2026-08-18.

---

## 1. Preserved baseline

T3 does not rediscover the following decisions:

```text
Group + GroupMembership
RoleAssignment subject = User | Group
RoleAssignment scope = Company | Area
static product-owned Role vocabulary
static product-owned Permission vocabulary
additive grants + default deny
live direct User grants + live Group-mediated grants
provider roles/groups/claims never canonical Authorization
Session/JWT is not durable Role/Permission authority
Controlled Documents owns relationship/lifecycle/governance predicates
no role bypasses domain governance
offboarding preserves historical User identity
re-enable never silently restores grants/memberships/sessions
AuditEvent = evidence, not current state
critical governed/security mutations require same-local-commit Audit
Audit append-only + PII-minimized
no global AuditChainHead/hash-chain Launch requirement
```

The old exact `5×43` catalog remains superseded.

---

## 2. Canonical Authorization equation

Authorization is live and compositional:

```text
enabled authenticated User
+
(
  current direct User RoleAssignments
  UNION
  current GroupMemberships → current Group RoleAssignments
)
+
static Role → Permission bundle
+
scope match
+
Controlled Documents relationship/state/governance predicates
=
ALLOW
```

Otherwise default DENY.

No persisted semantic authority exists for:

```text
user_permissions
expanded group permission cache
Session role list
JWT permission snapshot
provider-role mapping
materialized ACL
```

Caches may exist later only when provably equivalent to current canonical truth and immediately invalidatable/reconcilable; they never become authority.

---

## 3. Launch Role vocabulary

Accepted static product roles:

```text
governance_admin
area_manager
author
approver
viewer
governance_viewer
```

Roles are additive and non-hierarchical. A User or Group may hold multiple roles where responsibilities genuinely overlap. No role is an implicit superuser or domain-governance bypass.

---

## 4. Launch Permission vocabulary

Accepted Launch permissions:

```text
organization.manage
access.manage
document_type.manage
template_use.manage

document.read_effective
document.read_history
document.read_working
document.create
document.edit
document.submit
document.cancel_revision
document.obsolete
document.owner.manage

governance.act

audit.read
```

No dormant Launch permission is created for Distribution, Periodic Review, Dossier, Evidence, Records/Hold/Disposition, Governed Export, repository connectors, Training/LMS, generic Change Control, fresh-auth/eSignature, reassign/overseer/quorum/SLA or Audit export.

### Bounded non-permission operations

No separate `document.withdraw` is required in Launch. Submission withdraw is permitted through:

```text
document.submit
+ exact active Submission relationship
+ responsible-owner/managed-document relationship
+ T2 pre-Release predicate
```

No separate `document.revise` is required in Launch. Creating the next Revision is permitted through:

```text
document.edit
+ Document relationship predicate
+ T2 next-Revision eligibility
```

No separate session-administration permission exists because Launch has no standalone session-administrator journey; required session revocation is part of the governed User offboarding operation and restore readiness may invalidate restored sessions structurally under T4.

Current governance route and representation configuration are DocumentType configuration under `document_type.manage`; no separate Approval-policy permission family exists.

---

## 5. Accepted Role bundles

### viewer

```text
document.read_effective
```

### author

```text
document.read_effective
document.read_history
document.read_working
document.create
document.edit
document.submit
```

### approver

```text
document.read_effective
governance.act
```

`governance.act` never grants blanket working/history access. Exact active-Step participation opens only the exact Submission/governance context needed for that decision.

### area_manager

```text
document.read_effective
document.read_history
document.read_working
document.create
document.edit
document.submit
document.cancel_revision
document.obsolete
document.owner.manage
governance.act
```

`area_manager` remains an operational Area role, not RBAC administration.

### governance_admin

```text
organization.manage
access.manage
document_type.manage
template_use.manage
```

`governance_admin` configures organization/access/document-governance settings but receives no automatic content read/edit/governance/Audit authority. Additional roles must be assigned deliberately when the same person needs those capabilities.

### governance_viewer

```text
document.read_effective
document.read_history
audit.read
```

This is the least-privilege read-only governance/auditor path. It has no content mutation, governance decision, configuration or access-administration authority.

---

## 6. RoleAssignment subject/scope matrix

All accepted roles may be assigned to either:

```text
User
Group
```

No special direct-User-only administrative restriction exists.

Scope compatibility:

| Role | CompanyScope | AreaScope |
|---|---:|---:|
| `governance_admin` | yes | no |
| `area_manager` | no | yes |
| `author` | yes | yes |
| `approver` | yes | yes |
| `viewer` | yes | yes |
| `governance_viewer` | yes | yes |

`governance_admin` is Company-wide because it manages Organization/AuthZ/DocumentType configuration. `area_manager` is intentionally Area-operational. The other roles can be delegated company-wide or per Area.

No product-level last-direct-admin rule is introduced without a concrete Launch failure requirement. Bootstrap/recovery remains an explicit non-serving operations concern, never an ordinary RBAC bypass.

---

## 7. Organization administration vs access administration

### organization.manage

Controls company organization identity/lifecycle where exposed:

```text
Company display/current organization settings
User / UserProfile lifecycle
Area lifecycle
Group identity/lifecycle
```

### access.manage

Controls effective access configuration:

```text
GroupMembership add/remove
RoleAssignment grant/revoke
```

GroupMembership remains Organization-owned semantic state, but its mutation is access-sensitive because:

```text
Group has RoleAssignment
+ add User to Group
= effective authority changes
```

A caller with only `organization.manage` may not grant authority indirectly through security-bearing membership.

Launch has no Area-local RBAC administrator. `area_manager` is operational, not access-administrative.

---

## 8. Group lifecycle / deletion law

Group identity remains Organization-owned.

Group deletion requires `organization.manage` and fails closed while any live dependency remains:

```text
current GroupMembership
current Group RoleAssignment
current GovernanceRoute GROUP selector
live GovernanceAttempt route snapshot containing an unactivated GROUP Step that still needs resolution
```

Once a Group Step activates and resolves to its concrete User candidate snapshot, that resolved Step no longer needs Group identity for historical participant attribution. Completed historical attempts therefore do not keep Group alive solely because the Group once participated.

No nested/dynamic/provider-mirrored Group model exists in Launch.

---

## 9. Canonical document authorization predicates

A Permission grant is necessary but never sufficient.

### Effective read

```text
enabled User
+ document.read_effective in Company or matching Area
+ target Revision = current EFFECTIVE
→ allow
```

### Historical read

```text
enabled User
+ document.read_history in Company or matching Area
→ read authorized immutable Revision/Submission/governance history
```

### Working read/edit/submit/withdraw/revise

Authoring relationship predicate:

```text
actor is current Document responsible owner
OR
actor also has document.owner.manage in scope
```

Then:

```text
document.read_working → inspect current open work
document.edit         → mutate DRAFT under T2 OCC; create next Revision when eligible
document.submit       → submit exact generation; withdraw own/managed pre-Release attempt
```

Ordinary `author` therefore does not receive blanket Area edit authority. `area_manager` can manage Area work because its bundle contains `document.owner.manage`.

### Create

```text
document.create in Company or matching Area
+ active DocumentType / Area eligibility
→ create Document + REV000
```

Ordinary author creation makes the actor the responsible owner by default unless an actor with `document.owner.manage` deliberately selects another eligible responsible User.

### Cancel Revision

```text
document.cancel_revision
+ matching scope
+ T2 cancellation eligibility
→ cancel open Revision
```

### Obsolescence

Initiate/complete under the existing permission:

```text
document.obsolete
+ matching scope
+ exact current EFFECTIVE target
+ mandatory reason
+ no open replacement
+ no competing obsolescence
+ T2 governance mode/route
→ initiate/complete governed obsolescence
```

Withdraw an active **human-governed** obsolescence request:

```text
document.obsolete
+ matching scope
+ active pre-completion obsolescence request
+ (
    actor is the request initiator
    OR actor also has document.owner.manage in scope
  )
+ T2 withdrawal eligibility
→ terminate request as WITHDRAWN
→ current target remains EFFECTIVE
```

This creates no fake participant `RETURN_FOR_CHANGES` decision. `NoHumanApproval` obsolescence completes synchronously and therefore has no live request window to withdraw.

### Change responsible owner

```text
document.owner.manage
+ matching scope
+ eligible target User
→ change current responsibility
```

### Governance action

```text
enabled User
+ governance.act in Company or matching Area
+ exact live GovernanceAttempt
+ exact active Step
+ User is in activated candidate snapshot
+ User is not Submission submitter / obsolescence initiator
→ allow ACCEPT / RETURN_FOR_CHANGES and exact-case feedback context
```

Case participation permits inspection of the exact governed Submission/attempt context needed to decide. It does not create general working/history authority.

---

## 10. Offboarding transaction law

An authorized `organization.manage` offboarding operation is one local semantic transaction:

```text
BEGIN
serialize User eligibility against concurrent security-sensitive mutations
mark User disabled
revoke all live ApplicationSessions
remove all current GroupMemberships for User
remove all direct User RoleAssignments
append required Audit evidence for access teardown + offboarding
COMMIT
```

`ProviderSubjectBinding` remains where needed because provider-subject→User historical correlation remains truthful after employment/access ends.

Group RoleAssignments remain because they belong to the Group; removing membership removes the offboarded User's inherited access.

**T5-L is the baseline for provider-side disable:** MetalDocs access correctness does not require a durable IdP-disable job. If a future assurance requirement explicitly makes provider-state convergence mandatory, that same offboarding transaction may insert the smallest named durable provider-disable intent under T5's transaction-coupled-job rules. No generic identity-sync engine is implied.

Re-enable changes User eligibility and appends required Audit, but never resurrects:

```text
old ApplicationSessions
old direct RoleAssignments
old GroupMemberships
```

Fresh grants/memberships are explicit new access decisions.

---

## 11. Authorization-sensitive concurrency with offboarding

Governed/security-changing operations whose correctness depends on an enabled User must serialize with offboarding on current User eligibility.

Required outcome:

```text
action serializes first
→ action may commit under valid pre-offboarding state
→ offboarding then commits and stops future actions

offboarding serializes first
→ later action observes disabled User
→ fail closed
```

This applies at least to:

```text
new Session issuance
GroupMembership add
new direct User RoleAssignment
governance ACCEPT / RETURN_FOR_CHANGES
Submission / withdraw / cancel / obsolescence user mutations
```

T3 does not claim magical cancellation of already committed/linearized work or ordinary reads that completed authorization before offboarding linearized.

Historical restore is a separate readiness boundary: T4 requires restored sessions to be invalidated and post-snapshot security teardown to be reconciled before ordinary serving resumes.

---

## 12. Audit model

Audit is explicit semantic evidence, never inferred from database CRUD and never event sourcing.

Required form:

```text
BEGIN
business/security semantic mutation
required owning-domain evidence
required AuditEvent(s)
COMMIT
```

If required Audit append fails, the governed/security mutation fails with the same local transaction.

No deployment-wide AuditChainHead/global hash-chain serialization is required in Launch.

---

## 13. AuditEvent semantic minimum

Each AuditEvent contains:

```text
stable event id
trusted server/transaction time

actor:
  USER   → stable User id
  SYSTEM → product-owned system actor code

operation code
resource kind + stable resource id

visibility attribution:
  COMPANY
  OR AREA + Area id snapshot

bounded facts schema + PII-minimized facts
correlation id when useful
```

Generic `resource_kind/resource_id` is acceptable here because Audit explicitly does not own resource lifecycle.

Audit does not copy by convenience:

```text
name/email/username
password/token/JWT/provider claims
request body/headers
free-form Submission feedback
free-form return/obsolescence/cancellation reason
full governed content
IP/user-agent by default
```

The owning domain record remains authority for reasons/comments/content. Audit references IDs/outcomes and the minimum bounded facts required to reconstruct the action.

---

## 14. Audit read visibility

Audit visibility is snapshotted at action time.

```text
audit.read @ Company
→ Company events + all Area-attributed events

audit.read @ Area X
→ only events historically attributed to Area X
→ no Company-wide User/access/config administration merely because a User later belonged to Area X
```

Current resource relocation/renaming never rewrites historical Audit visibility attribution.

This permits least-privilege Area Governance Viewer/Auditor without requiring Audit to resolve current resource state to decide historical visibility.

---

## 15. Required same-local-commit Audit census

### Authentication / Organization / Access

```text
ProviderSubjectBinding accepted / disabled / replaced
User created / offboarded / re-enabled
UserProfile erased when lawful-erasure action itself requires evidence
Area created / renamed / retired / re-enabled
Group created / renamed / deleted
GroupMembership added / removed
RoleAssignment granted / revoked
```

Offboarding-generated membership/grant/session teardown must remain reconstructible. Multiple semantic events may be appended in the same transaction plus the final User offboarding event.

### Controlled Document configuration

```text
DocumentType created / materially reconfigured / activated / inactivated
GovernanceRoute/representation configuration changed as DocumentType configuration
Template role/eligibility changed
Document responsible owner changed
```

### Controlled Document lifecycle/governance

```text
Document + REV000 created
later Revision created
Submission created
ACCEPT
RETURN_FOR_CHANGES
Submission withdrawn
Revision cancelled
OfficialRendition semantic completion when a required record is established
Release completed
obsolescence requested
obsolescence withdrawn
obsolescence completed
```

`SubmissionFeedback` already has immutable actor/time/content authority and does not require a duplicate semantic Audit event merely to copy its text.

### Not mandatory semantic Audit in Launch

```text
every autosave/WorkingContent keystroke
Search query
ordinary effective read
ordinary download
login/logout
notification delivery
preview rendering
authorization denial
```

Those may generate operational/security telemetry, but not immutable semantic Audit solely because they occurred.

A named regulatory/customer requirement for audited reads/downloads is a reopen trigger.

---

## 16. Minimum bounded Audit facts

### RoleAssignment grant/revoke

```text
assignment id
subject kind + User/Group id
role code
scope kind + Company/Area id
```

### GroupMembership add/remove

```text
User id
Group id
```

### Governance decision

```text
GovernanceAttempt id
Step id
Decision id
subject kind = SUBMISSION | OBSOLESCENCE
subject id
outcome = ACCEPT | RETURN_FOR_CHANGES
```

Reason/feedback text remains owning-domain evidence.

### Release

```text
Document id
Revision id
winning Submission id
predecessor Revision id when replacement
```

### Cancellation / obsolescence

```text
Document id
Revision id
cancellation/obsolescence evidence id
```

Reason text remains owning-domain evidence.

### DocumentType/configuration changes

Facts contain bounded changed configuration identifiers/codes/operation facts, not arbitrary before/after JSON dumps.

---

## 17. Audit retention boundary

T3 does not claim a statutory or indefinite Audit retention period.

Launch has no Audit-disposition capability, so required Audit is not pruned by ordinary product paths. A later accepted Records/compliance requirement must define retention/pruning/checkpoint semantics deliberately.

---

## 18. Future-evolution law

Future capability permissions never silently broaden existing role bundles.

When Distribution, Periodic Review, Dossier, Evidence, Records, Export, Repository connectors, Training or Change Control are promoted, their material permissions/role changes require explicit future design and migration.

Stable seams remain:

```text
User / Group / Company / Area grant model
current RoleAssignment authority
static product Role/Permission vocabulary
Controlled Documents lifecycle predicates
exact active governance participation
Audit as evidence, not state
```

Pooled tenancy may reopen deployment/substrate mechanics around stable Company identity without changing the semantic meaning of User/Group/Area grants. CRDT may replace DRAFT collaboration mechanics without changing Authorization or immutable Submission governance.

---

## 19. Reopen triggers

Reopen only the implicated T3 decision on material evidence that Launch/future promoted scope requires:

```text
customer-editable custom roles/permissions
explicit deny rules / ACL / ReBAC graph
Area-local RBAC administration
separate submit vs withdraw authority
separate edit vs create-next-Revision authority
strict cross-Step SoD / fresh-auth/eSignature
read/download/search Audit as compliance evidence
Audit export
cryptographic tamper-evidence/global anchoring
finite Audit retention/pruning
standalone session administration
mandatory provider-side identity convergence for assurance
mandatory last-admin product invariant beyond bootstrap/ops recovery
```

---

## 20. Explicit non-decisions

T3 does not decide:

```text
final SQL constraints/indexes/lock clauses
Go package boundaries
HTTP routes/error envelopes
frontend role/admin screens
storage/digest implementation
async/outbox mechanism
Search technology
Historical Migration execution
future Launch+/Future permission catalogs
```

Those belong to later stages/implementation design.

Implementation remains **BLOCKED**.