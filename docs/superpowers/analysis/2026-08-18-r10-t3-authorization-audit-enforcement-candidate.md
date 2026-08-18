# R10-T3 — Authorization & Audit Enforcement — Reconciled Candidate

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — OPERATOR ADJUDICATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **T1 authority:** `wiki/architecture/r10-t1-semantic-state-invariants.md`  
> **T2 authority:** `wiki/architecture/r10-t2-governance-effectivity-transactions.md`  
> **Implementation:** BLOCKED

T3 is rebuilt from the operator-ratified Decision Registry. It does **not** rediscover already-good primitives and does **not** repair the old 5×43 model.

## 1. T3 decision question

> **Given the preserved User/Group + Company/Area RoleAssignment model and the ratified T1/T2 lifecycle, what exact Launch roles, permissions, scope compatibility, administration rules, check sites and same-local-commit Audit evidence are the smallest sustainable enforcement model?**

T3 intentionally does not define SQL/index syntax, package layout, API routes, frontend screens, storage, worker topology or migration execution.

## 2. Registry baseline — NOT open for aesthetic redesign

T3 MUST consume:

```text
Group + GroupMembership
RoleAssignment subject = User | Group
RoleAssignment scope = Company | Area
static product-owned Role vocabulary
static product-owned Permission vocabulary
additive grants + default deny
live direct User grants + live Group-mediated grants
provider roles/groups/claims never canonical AuthZ
Session/JWT is not durable Role/Permission authority
Controlled Documents owns relationship/lifecycle/governance predicates
no role, including admin, bypasses domain governance
Area Manager remains a preserved operational role concept
access.manage protecting membership/grants is strong prior candidate law
offboarding preserves historical User identity
re-enable never silently restores grants/memberships/sessions
AuditEvent = evidence, not current state
critical governed/security mutations require same-local-commit Audit in principle
Audit append-only + PII-minimized
global AuditChainHead/hash-chain serialization is not Launch baseline
```

T3 MUST NOT revive:

```text
old exact 5×43 permission matrix
provider-role bridge
custom-role/custom-permission platform
generic ACL/ReBAC graph
Tenant universal partition/RLS policy engine
ROLE_IN_AREA routing
approval.oversee/reassign/cancel platform
Distribution/Periodic Review/Dossier/Evidence/Records permissions
AuditChainHead/global hash-chain
```

## 3. Credible Authorization alternatives

### A — repair old 5×43 by subtraction

**Reject — Local Maximum.** The old bundles were shaped by Distribution, Periodic Review, Dossier, Evidence, Records, Interchange and richer Approval semantics that are no longer Launch authority.

### B — generic ACL/ReBAC/policy engine

**Reject — accidental generality.** Launch already has stable organizational groups, scoped RoleAssignments and domain relationship predicates. A graph/policy platform adds another authority without a consumer.

### C — preserve the simple RBAC kernel; rederive only role/bundle surface

```text
enabled User
+
(current direct RoleAssignments
 UNION current GroupMembership → Group RoleAssignments)
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

**Recommended Global Maximum.**

## 4. Recommended Launch role vocabulary

Preserve the useful prior role concepts and refine only what changed:

```text
governance_admin   // refinement of old tenant_owner for single-company Launch
area_manager       // preserved operational Area manager
author             // preserved
author? no — one role only: author
approver           // preserved
viewer             // preserved
governance_viewer  // new least-privilege Auditor/Governance Viewer required by GCR
```

Canonical set:

```text
governance_admin
area_manager
author
approver
viewer
governance_viewer
```

Roles are static, additive and non-hierarchical. A User/Group may hold multiple roles when responsibilities overlap. `governance_admin` is not a superuser/bypass.

## 5. Recommended Launch permission vocabulary

Preserve current-use names where still semantically correct and remove permissions tied only to deferred/superseded capability.

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

Total: **15 Launch permissions**.

### Why no separate `document.withdraw`

Withdraw is a bounded continuation of submitter/author authority on the exact pre-Release attempt:

```text
document.submit
+ responsible-author / owner-management relationship
+ exact active Submission
+ pre-Release T2 predicate
→ withdraw
```

Add a separate permission only if a real consumer later needs “may submit but may not withdraw”.

### Why no separate `document.revise`

Starting the next business Revision is an authoring operation over an existing effective Document:

```text
document.edit
+ relationship predicate
+ T2 new-Revision eligibility
→ create next Revision
```

Add a separate permission only if later business policy needs different delegation.

### Why no `session.manage`

Offboarding/revocation consequences are part of the User lifecycle operation. Launch has no named standalone “session administrator” journey.

### Why no separate governance-config permission

Current route + representation policy are DocumentType configuration. `document_type.manage` owns current DocumentType configuration; no separate Approval owner/policy platform exists.

## 6. Recommended Role → Permission bundles

### `viewer`

```text
document.read_effective
```

### `author`

```text
document.read_effective
document.read_history
document.read_working
document.create
document.edit
document.submit
```

This preserves the old author concept while removing deferred comment/periodic-review/Evidence/Dossier permissions.

### `approver`

```text
document.read_effective
governance.act
```

`governance.act` does not grant blanket draft/history access. Exact active-Step participation opens the exact Submission/governance context required for the decision.

### `area_manager`

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

This preserves the prior operational manager meaning while deleting old Approval-overseer/reassign, Distribution, Evidence and Dossier capabilities.

### `governance_admin`

```text
organization.manage
access.manage
document_type.manage
template_use.manage
```

This role configures organization/access/document-governance settings but does **not** automatically gain content read/edit/governance/Audit rights. Add another role when the same person needs them.

### `governance_viewer`

```text
document.read_effective
document.read_history
audit.read
```

Read-only governance/auditor path. No content mutation, governance decision, configuration or access administration.

## 7. RoleAssignment subject/scope matrix

All roles remain assignable to either subject kind:

```text
User
Group
```

No special “admin must be direct User” restriction is introduced without evidence.

Scope compatibility:

| Role | CompanyScope | AreaScope |
|---|---:|---:|
| governance_admin | yes | no |
| area_manager | no | yes |
| author | yes | yes |
| approver | yes | yes |
| viewer | yes | yes |
| governance_viewer | yes | yes |

Rationale:

- `governance_admin` manages whole-company Organization/AuthZ/DocumentType configuration;
- `area_manager` is deliberately an Area-operational role, not company RBAC administrator;
- author/approver/viewer/governance_viewer may be delegated company-wide or Area-locally.

No “last direct admin” rule is introduced. If a real lockout failure proves a product-level invariant is needed, reopen explicitly; bootstrap/maintenance recovery is not an ordinary RBAC bypass.

## 8. Access administration law

### `organization.manage`

Company-scoped current organization administration:

```text
Company display/current organizational settings where exposed
User / UserProfile lifecycle
Area lifecycle
Group identity/lifecycle
```

### `access.manage`

Company-scoped current access administration:

```text
GroupMembership add/remove
RoleAssignment grant/revoke
```

Reason GroupMembership belongs behind `access.manage` even though Organization owns its semantics:

```text
Group has RoleAssignment
+ add User to Group
= effective authority changes
```

Someone with only `organization.manage` must not be able to grant authority indirectly by changing security-bearing membership.

Launch has no Area-local RBAC administrator. `area_manager` is operational, not access-administrative.

## 9. Group lifecycle/deletion law

Group identity remains Organization-owned.

Group deletion requires `organization.manage` and is fail-closed while **any live dependency** remains:

```text
current GroupMembership
current Group RoleAssignment
current GovernanceRoute GROUP selector
live GovernanceAttempt route snapshot containing an unactivated GROUP Step that still needs resolution
```

Once an active Group Step has resolved to its concrete User candidate snapshot, that resolved Step no longer needs Group identity for historical attribution. Completed historical attempts therefore do not keep a Group alive solely for history.

No nested/dynamic/provider-mirrored Group model is introduced.

## 10. Document relationship predicates / canonical check sites

A Permission grant is never sufficient by itself.

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
→ authorized immutable revision/submission/governance history
```

History permission does not grant current DRAFT mutation.

### Working read/edit/submit/withdraw/revise

Authoring relationship:

```text
actor is current Document responsible owner
OR actor also has document.owner.manage in scope
```

Then:

```text
document.read_working → inspect current open work
document.edit         → mutate DRAFT under T2 OCC; also start next Revision when eligible
document.submit       → submit exact generation; withdraw own/managed pre-Release attempt
```

`document.owner.manage` therefore lets `area_manager` manage work across its Area while ordinary `author` remains bounded to Documents for which the actor is responsible.

### Create

```text
document.create in Company or Area
+ active DocumentType / Area eligibility
→ create Document + REV000
```

Ordinary author creation makes the actor responsible owner by default unless an actor with `document.owner.manage` deliberately chooses another eligible responsible User.

### Cancel Revision

```text
document.cancel_revision
+ matching scope
+ T2 cancellation eligibility
→ cancel open Revision
```

### Obsolescence

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

### Change responsible owner

```text
document.owner.manage
+ matching scope
+ target User eligible
→ change current responsibility
```

### Governance decision / feedback context

```text
enabled User
+ governance.act in Company or matching Area
+ exact live GovernanceAttempt
+ exact active Step
+ User is in activated candidate snapshot
+ User is not Submission submitter / obsolescence initiator
→ allow governance action
```

Case participation permits inspection of the exact governed Submission/attempt context required to decide. It does not create general `document.read_working`/history authority.

## 11. Canonical grant evaluation

No persisted semantic authority exists for:

```text
user_permissions
expanded group permission cache
Session role list
JWT permission snapshot
provider-role mapping
materialized ACL
```

Canonical authorization uses current truth:

```text
current User direct RoleAssignments
UNION
current GroupMemberships → current Group RoleAssignments
→ static role bundle
→ scope match
→ domain relationship/state/governance predicate
```

Caches may exist later only if they are provably equivalent/revocable and never become source authority.

## 12. Offboarding exact transaction law

An authorized `organization.manage` offboarding operation is one local semantic transaction:

```text
BEGIN
serialize User eligibility against concurrent security-sensitive mutations
mark User disabled
revoke all live ApplicationSessions
remove all current GroupMemberships for User
remove all direct User RoleAssignments
append required Audit evidence for access teardown + offboarding
insert durable provider-disable intent only if provider-side effect is required
COMMIT
```

`ProviderSubjectBinding` remains because provider-subject→User historical correlation remains truthful.

Group RoleAssignments remain because they belong to Group; removing membership removes the User's inherited access.

Re-enable:

```text
mark User enabled
append required Audit
```

and never resurrects:

```text
old Sessions
old direct RoleAssignments
old GroupMemberships
```

Fresh grants/memberships are explicit new access decisions.

## 13. Authorization-sensitive concurrency with offboarding

For governed/security-changing mutations whose correctness depends on an enabled User, action transaction and offboarding must serialize on current User eligibility.

Required outcome:

```text
action serializes first
→ action may commit under the pre-offboarding valid state
→ offboarding waits, then commits and stops future actions

offboarding serializes first
→ later action observes disabled User and fails closed
```

This applies at least to:

```text
new Session issuance
GroupMembership add
new direct User RoleAssignment
governance ACCEPT / RETURN
Submission/withdraw/cancel/obsolescence user mutations
```

T3 does not claim magical cancellation of already committed work or ordinary reads that completed authorization before offboarding linearized. Stronger global request cancellation requires a concrete future invariant.

## 14. Audit alternatives

### A — infer Audit from DB CRUD triggers

**Reject.** Row mutation does not know business operation meaning and duplicates domain interpretation in DB triggers.

### B — event sourcing / global hash-chain serialization

**Reject for Launch.** Audit would become shadow history authority and the global chain reintroduces contention explicitly removed by GCR A9.

### C — explicit semantic Audit appended in same local transaction for a bounded census

**Recommended.**

```text
BEGIN
business/security semantic mutation
required domain evidence
required AuditEvent(s)
COMMIT
```

If required Audit append fails, the governed/security mutation fails with the transaction.

## 15. AuditEvent semantic minimum

```text
AuditEvent
  stable id
  trusted server/transaction time

  actor:
    USER → stable User id
    SYSTEM → product-owned system actor code

  operation code
  resource kind + stable resource id

  visibility attribution:
    COMPANY
    OR AREA + area id snapshot

  bounded facts schema + PII-minimized facts
  correlation id when useful
```

`resource_kind/resource_id` may be generic because Audit explicitly does not own resource lifecycle.

Audit does **not** copy by convenience:

```text
name/email/username
password/token/JWT/provider claims
request body/headers
free-form Submission feedback
free-form return/obsolescence/cancellation reason
full governed content
IP/user-agent by default
```

The owning domain record remains the authority for reason/comment/content. Audit references IDs/outcomes and other bounded facts needed for action reconstruction.

## 16. Audit visibility / `audit.read`

Audit events snapshot visibility attribution at action time.

```text
audit.read @ Company
→ may read Company + all Area-attributed Audit events

audit.read @ Area X
→ may read only Area-X attributed Audit events
→ may not read Company-wide User/access/config administration events merely because a User later belongs to Area X
```

Current resource relocation/renaming never rewrites historical Audit visibility attribution.

This enables least-privilege Area Governance Viewer/Auditor without making Audit query current resource state for authorization.

## 17. Required same-local-commit Audit census

### Authentication / Organization / Access

```text
ProviderSubjectBinding accepted / disabled / replaced
User created
offboarded
re-enabled
UserProfile erased when lawful-erasure action itself requires evidence
Area created / renamed / retired / re-enabled
Group created / renamed / deleted
GroupMembership added / removed
RoleAssignment granted / revoked
```

Offboarding-generated membership/grant/session teardown must remain reconstructible. It may append multiple semantic Audit events in the same transaction plus the final User offboarding event.

### Controlled Document configuration

```text
DocumentType created / materially reconfigured / activated / inactivated
GovernanceRoute/representation configuration changed as part of DocumentType config
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
OfficialRendition semantic completion when required record is established
Release completed
obsolescence requested
obsolescence completed
```

`SubmissionFeedback` already carries immutable actor/time/content authority and does not require a duplicate semantic Audit event merely to copy its text.

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

These may produce security/operational telemetry, but not immutable semantic Audit merely because they occur.

A regulation/customer requirement for audited reads/downloads is an explicit T3 reopen trigger.

## 18. Minimum bounded Audit facts by operation family

### RoleAssignment grant/revoke

```text
assignment id
subject kind + User/Group id
role code
scope kind + Company/Area id
```

Allows historical access reconstruction after the current grant row is gone.

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

Reason/feedback text remains domain evidence.

### Release

```text
Document id
Revision id
winning Submission id
predecessor Revision id when replacement
```

### Cancellation/obsolescence

```text
Document id
Revision id
cancellation/obsolescence evidence id
```

Reason text remains domain evidence.

### DocumentType/config changes

Facts contain bounded changed configuration identifiers/codes/version-independent operation facts, not arbitrary before/after JSON dumps.

## 19. Audit retention boundary

T3 does **not** claim a statutory or indefinite retention period.

Launch has no Audit-disposition capability, so required Audit is not pruned by ordinary product paths. A future accepted Records/compliance requirement must define retention/pruning/checkpoint semantics deliberately.

No global hash chain is introduced to justify retention.

## 20. Future capability attack

| Future capability | T3 seam preserved |
|---|---|
| Distribution | add explicit permissions/roles when promoted; User/Group + Company/Area grant model survives; no silent role broadening |
| Periodic Review | add review permission/role deliberately; current role bundles do not gain it automatically |
| Dossier | future links never grant access; target resources still use canonical AuthZ |
| Evidence | may introduce its own permissions/lifecycle while reusing Organization/AuthZ primitives |
| Records/Hold/Disposition | adds records permissions without turning current Document status/storage into policy authority |
| Governed Export | add explicit export permission only when capability promoted |
| Repository connector | provider identity never becomes AuthZ identity |
| Training/LMS | training permissions remain separate from document read/effectivity |
| Change Control | orchestration cannot bypass each Document's lifecycle/AuthZ predicates |
| pooled tenancy | substrate may reopen around Company identity without invalidating User/Group/Area semantic grant model |
| CRDT | DRAFT collaboration changes do not alter authorization vocabulary or immutable Submission governance |

Binding rule:

> **Adding a material future permission to an existing Role bundle is itself an authority change and requires explicit design/migration; future features never silently broaden current roles.**

## 21. Proof obligations before implementation

Later implementation design/tests must falsifiably prove at least:

```text
provider roles/groups cannot create product authority
current Group membership changes affect inherited authority on next canonical check
removed RoleAssignment no longer grants authority on next canonical check
AreaScope never grants another Area
CompanyScope applies where explicitly allowed
area_manager cannot administer RBAC merely through operational role
governance_admin does not automatically gain content/Audit access
ordinary author cannot edit/submit a Document outside responsible-owner predicate
area_manager can manage Area documents through document.owner.manage predicate
governance.act cannot act outside active Step snapshot
submitter/initiator self-approval remains impossible regardless of role
Group deletion fails while any live access/governance dependency remains
offboarding and security-sensitive mutations total-order on User eligibility
re-enable does not resurrect old access
required audited mutation cannot commit without AuditEvent
Area-scoped audit.read cannot expose Company-wide security/admin events
UserProfile erasure does not destroy stable historical actor attribution
Audit facts contain no forbidden PII/secrets/free-form domain payload
```

## 22. Reopen triggers

Reopen the implicated T3 decision only on material evidence that Launch needs:

```text
customer-editable custom roles/permissions
explicit deny rules / ACL / ReBAC graph
Area-local RBAC administration
separate submit vs withdraw authority
separate edit vs create-next-Revision authority
role-specific relationship semantics that current permission+predicate composition cannot express cleanly
strict cross-Step SoD / fresh-auth/eSignature
read/download/search Audit as compliance evidence
Audit export
cryptographic tamper-evidence/global anchoring
finite Audit retention/pruning
standalone session administration
mandatory last-admin product invariant beyond bootstrap/ops recovery
```

## 23. Explicit non-decisions

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

## 24. Operator adjudication packet

Recommended dispositions:

```text
T3-A ACCEPT — preserve canonical live RBAC kernel: User|Group RoleAssignments, Company|Area scopes, static product catalogs, additive grants/default deny + domain predicates.
T3-B ACCEPT — Launch roles = governance_admin, area_manager, author, approver, viewer, governance_viewer; first five preserve/refine prior concepts and governance_viewer adds required least-privilege audit path.
T3-C ACCEPT — Launch permission catalog = the 15 permissions in §5; no old/future capability permissions.
T3-D ACCEPT — Role bundles in §6; admin is configuration-only by default, area_manager remains Area-operational, approver gets exact-case governance via governance.act, governance_viewer is read-only.
T3-E ACCEPT — all roles may be assigned to User or Group; scope matrix in §7: governance_admin Company-only, area_manager Area-only, the other four Company|Area.
T3-F ACCEPT — `organization.manage` owns User/Area/Group identity lifecycle; `access.manage` owns GroupMembership + RoleAssignment mutations; no Area-local RBAC admin Launch.
T3-G ACCEPT — Group deletion fails while memberships, Group grants, current governance routes or live unactivated GROUP-step snapshots still depend on it; historical resolved/completed Steps do not keep Group alive solely for attribution.
T3-H ACCEPT — authoring relationship predicate = current responsible owner OR actor also has document.owner.manage; area_manager can therefore manage Area work without giving ordinary authors blanket Area edit authority.
T3-I ACCEPT — governance action requires current governance.act + exact active-Step candidate snapshot + enabled User + T2 self-approval predicates; permission alone never grants a verdict.
T3-J ACCEPT — offboarding atomically disables User, revokes Sessions, removes memberships/direct grants, appends Audit and optionally records durable provider-disable intent; re-enable restores no old access.
T3-K ACCEPT — governed/security-changing actions and offboarding total-order on current User eligibility; no claim of magical cancellation for already-linearized work.
T3-L ACCEPT — Audit = explicit semantic append-only evidence in the same local transaction for the bounded census in §17; no CRUD-trigger inference, event sourcing or global hash-chain serialization.
T3-M ACCEPT — AuditEvent minimum = actor, trusted time, operation/resource attribution, immutable Company|Area visibility snapshot, bounded schema/facts and correlation when useful; PII/free-form domain payload stays out.
T3-N ACCEPT — audit.read may be Company or Area scoped; Area viewer sees only events attributed to that Area, not Company-wide access/identity administration.
T3-O ACCEPT — autosave/search/read/download/login/logout/notification/preview/deny telemetry is not mandatory semantic Audit in Launch; reopen on named compliance requirement.
T3-P ACCEPT — future capability permissions never silently broaden existing roles; adding material authority requires explicit future design/migration.
```

T3 remains non-authoritative until operator adjudication. After adjudication, **T4 still must not open**: the mandatory platform-facing T3 summary must be presented and explicitly ratified first.