# R10-T3 — Authorization & Audit Enforcement — Integrated Candidate

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE CANDIDATE — OPERATOR ADJUDICATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Technical authority:** `wiki/architecture/r10-technical-architecture.md`  
> **T2 authority:** `wiki/architecture/r10-t2-governance-effectivity-transactions.md`  
> **Implementation:** BLOCKED

T3 derives Launch V1 Authorization and Audit enforcement from the operator-ratified Product Contract, 4+1 ownership topology, T1 semantic state and T2 lifecycle transactions.

This candidate incorporates the operator correction made during T3 discovery:

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent primitive unless current authority or a concrete failure mode disproves it; rederive only the composite bundle/decision whose justification changed; defer only the capability that actually left Launch.**

T3 therefore does **not** rebuild Authorization by aesthetic reset and does **not** preserve the old exact `5×43` matrix by subtraction.

It intentionally does not define final SQL/index syntax, Go package layout, API routes, frontend UX, storage, worker topology or migration execution.

---

# 1. T3 decision question

> **Which Launch actors may execute each accepted operation, at which Company/Area scope and under which owning-domain predicates, and which critical operations must atomically append bounded Audit evidence?**

T3 succeeds when:

```text
provider identity != User identity != Authorization grant
RoleAssignment remains one current grant authority
Groups remain first-class organizational/access inputs
permissions derive from current Launch operations only
no role bypasses Controlled Documents lifecycle/governance predicates
least-privilege Governance Viewer / Auditor exists
security-relevant grant/membership changes cannot escape Audit
critical governed lifecycle mutations cannot succeed without same-local-commit Audit
Audit remains evidence, never current state
```

---

# 2. Authority / evidence boundary

Authority:

1. `AGENTS.md`;
2. DevelopmentConexus Engineering Method v1.0.0;
3. current handoff;
4. Launch V1 Product Contract;
5. Whole-Product GCR A1–A10;
6. operator-approved 4+1 ownership topology;
7. operator-ratified T1;
8. operator-ratified T2.

Prior R9/R9.5 and R10-B2 are evidence only.

The old B2 remains particularly useful for simple primitives that current authority still independently proves, but its exact role bundles and permission catalog are superseded where they encoded Distribution, Periodic Review, Dossier, Evidence, Records, Interchange or old Approval-owner richness.

---

# 3. Prior-decision survivorship classification

## 3.1 KEEP — simple primitives still independently proven

```text
flat Group
GroupMembership = current User↔Group truth
RoleAssignment subject = User | Group
RoleAssignment scope = Company | Area
static product-owned Role vocabulary
static product-owned Permission vocabulary
static Role→Permission bundles
additive grants + default deny
live direct User grants + live Group-mediated grants
provider roles/groups/claims never canonical AuthZ
Session never carries durable Role/Permission authority
scope match is necessary but not sufficient
domain relationship/lifecycle predicates remain with Controlled Documents
no generic ACL/ReBAC graph
no custom role/custom permission platform in Launch
RoleAssignment is current truth; revoke removes current grant
GroupMembership is current truth; removal removes current membership
```

These were not invalidated by the product rebaseline and remain the smallest sustainable model.

## 3.2 RESTRUCTURE — prior composite decisions whose bundle changed

```text
old tenant_owner bundle   → rederive as governance_admin for single-company Launch
old area_manager bundle   → preserve role concept, rederive Launch-only operations
old author bundle         → preserve concept, remove old/future capabilities
old approver bundle       → preserve concept, fold old Approval owner into governance.act
old viewer bundle         → preserve concept, current effective read only
old 5-role set            → add least-privilege governance_viewer required by GCR
old approval.* catalog    → replace with bounded governance.act where Launch proves it
old admin config catalog  → remove dictionary/old policy machinery not in current semantic model
```

## 3.3 DEFER / REMOVE FROM LAUNCH CATALOG

No Launch permission is created for:

```text
Distribution / acknowledgement
Periodic Review
Dossier
Evidence
Retention / Legal Hold / Disposition
Governed Export
External Repository publish/import platform
Training/LMS
Generic Change Control
Audit export
Approval reassign/oversee/cancel platform
fresh-auth/eSignature
quorum/SLA/escalation
custom roles/permissions
```

Future capability promotion may add permissions/roles deliberately. It may not silently broaden an existing role merely because the feature is adjacent.

---

# 4. Canonical Authorization equation

Authorization remains live and compositional:

```text
enabled authenticated User
+
(
  current direct User RoleAssignments
  UNION
  current GroupMemberships → current Group RoleAssignments
)
+
static Role→Permission bundle
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
user_permissions cache
expanded group grants
Session role list
JWT permission snapshot as durable authority
provider-role mapping
materialized ACL
```

Implementation may cache safely later, but every critical authorization result must remain equivalent to canonical current truth.

---

# 5. Scope semantics

T1 admitted exactly:

```text
Company
Area
```

T3 preserves that scope vocabulary.

For an Area-targeted Document operation:

```text
matching AreaScope grant
OR CompanyScope grant
```

may satisfy the grant side of Authorization.

For Company-wide administration, a CompanyScope grant is required.

Area on the Document remains owning-domain context. Authorization does not invent a second Document-scope field.

---

# 6. Recommended Launch role vocabulary

T3 recommends six static, composable, non-customer-editable roles:

```text
governance_admin
area_manager
author
approver
viewer
governance_viewer
```

The first five preserve useful prior product concepts, with bundles rederived from current Launch. `governance_viewer` is the new least-privilege path required by the accepted GCR/Product Contract.

Roles are additive, not hierarchical. A User/Group may receive more than one role where business responsibility genuinely overlaps.

No role is an implicit superuser or domain-governance bypass.

---

# 7. Recommended Launch permission vocabulary

T3 recommends the following **15 Launch permissions**:

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

## 7.1 Why there is no `document.withdraw`

Withdraw is not an independently delegated business authority in the accepted Product Contract. It is a bounded action of the authorized author/submitter before Release.

Candidate check:

```text
document.submit permission
+ exact active Submission relationship / author eligibility
+ pre-Release T2 state predicate
→ withdraw allowed
```

If a real consumer later needs “may submit but may not withdraw,” add a separate permission then.

## 7.2 Why there is no `document.revise`

Beginning the next business Revision is part of the accepted authoring authority over an existing Document.

Candidate check:

```text
document.edit permission
+ Document scope
+ T2 no-open-revision / no-active-obsolescence predicate
→ may create next Revision
```

`document.create` remains specifically the authority to create a new stable Document + `REV000`.

## 7.3 `document_type.manage`

Covers current DocumentType configuration:

```text
identity/display/status
numbering rule
governance mode/route configuration
official representation requirement
```

It does not absorb Template-document lifecycle/content authority.

## 7.4 `template_use.manage`

Preserved because Template remains a real Launch capability.

Covers only:

```text
mark/unmark eligible ordinary Document Template role where product journey allows
manage DocumentType↔Template eligibility configuration
```

Template content itself remains governed through ordinary Document authoring/lifecycle permissions.

---

# 8. Recommended role bundles and scope compatibility

## 8.1 `viewer`

Bundle:

```text
document.read_effective
```

Allowed subject:

```text
User | Group
```

Allowed scope:

```text
Company | Area
```

Ordinary reader truth only. No DRAFT/SUBMITTED/history/Audit access.

## 8.2 `author`

Bundle:

```text
document.read_effective
document.read_history
document.read_working
document.create
document.edit
document.submit
```

Allowed subject:

```text
User | Group
```

Allowed scope:

```text
Company | Area
```

Author can create `REV000`, begin later Revision through `document.edit`, mutate DRAFT WorkingContent, submit/resubmit and withdraw when the T2 relationship/state predicates permit.

Author does **not** automatically gain cancellation, obsolescence, owner reassignment, governance decision or Audit authority.

## 8.3 `approver`

Bundle:

```text
document.read_effective
governance.act
```

Allowed subject:

```text
User | Group
```

Allowed scope:

```text
Company | Area
```

`governance.act` never opens arbitrary DRAFT/history. The active GovernanceAttempt/Step relationship grants bounded access to the exact Submission or obsolescence context necessary to act.

## 8.4 `area_manager`

The role concept survives because it still has a concrete Launch consumer: area-scoped operational control of document change cycles without global configuration/RBAC administration.

Bundle:

```text
all `author` permissions

document.cancel_revision
document.obsolete
document.owner.manage

governance.act
```

Allowed subject:

```text
User | Group
```

Allowed scope:

```text
Area only
```

The route/Step still decides whether an area manager is an actual governance participant. `governance.act` alone cannot satisfy a Step.

`area_manager` cannot manage Organization, RoleAssignments, GroupMemberships, DocumentTypes or Audit by virtue of this role.

## 8.5 `governance_admin`

This is the current single-company name for the useful administrative concept formerly represented by `tenant_owner`; the old all-capabilities bundle is rejected.

Bundle:

```text
organization.manage
access.manage
document_type.manage
template_use.manage
document.owner.manage
```

Allowed subject:

```text
User | Group
```

Allowed scope:

```text
Company only
```

Important correction from the first T3 discovery proposal:

> A Group may hold `governance_admin @ Company`. There is no evidence justifying a direct-User-only restriction.

The role does **not** automatically grant document content access, governance decision authority or Audit read. Add `viewer` / `author` / `approver` / `governance_viewer` explicitly when the same people need those responsibilities.

No “last admin” product invariant is introduced in T3 absent a named product/recovery requirement; deployment recovery remains a separate trust surface.

## 8.6 `governance_viewer`

Bundle:

```text
document.read_effective
document.read_history
audit.read
```

Allowed subject:

```text
User | Group
```

Allowed scope:

```text
Company | Area
```

This is read-only governance reconstruction. It grants no edit/submit/govern/cancel/obsolete/config/access-management authority.

---

# 9. Group semantics preserved

Groups remain first-class in both Organization and Authorization.

```text
GroupMembership exists
→ User is current member

Group has RoleAssignment
+ User is current member
→ User receives that current group-mediated grant
```

No nested groups, dynamic rules or IdP-group mirroring are added.

## 9.1 GroupMembership administration

Organization owns GroupMembership truth, but T3 recommends mutation requires:

```text
access.manage
```

not merely `organization.manage`.

Reason: membership can directly change effective authority when the Group holds RoleAssignments.

This does not transfer semantic ownership to Authorization; it is an authorization rule for a security-sensitive Organization mutation.

## 9.2 Group deletion

A Group must not be deletable while live references exist, including:

```text
GroupMembership
Group RoleAssignment
current DocumentType governance route Step selecting that Group
```

An activated T2 Step freezes concrete candidate Users, so historical attempt truth does not require the Group to survive after live configuration no longer references it.

---

# 10. RoleAssignment current-truth law

RoleAssignment remains the single persisted Authorization grant family conceptually.

```text
assignment exists → grant currently exists
assignment absent → grant currently does not exist
```

Changing subject/role/scope is revoke + new grant, not in-place reinterpretation of one grant.

T3 does not introduce grant validity windows, scheduled grants, deny entries or historical grant rows.

Historical grant/revoke evidence belongs Audit.

---

# 11. Canonical domain check sites

Examples below are semantic checks, not API syntax.

## Read current effective

```text
enabled User
+ document.read_effective at Company/matching Area
+ exact Revision still EFFECTIVE
→ ALLOW
```

## Create Document / REV000

```text
enabled User
+ document.create at target Area
+ DocumentType/Area active eligibility
+ T2 create invariants
→ ALLOW
```

## Begin REV001+

```text
enabled User
+ document.edit at Document Area
+ no open Revision
+ no active obsolescence
+ current lifecycle eligible
→ ALLOW
```

## Edit DRAFT

```text
enabled User
+ document.edit at Document Area
+ Revision DRAFT
+ accepted author/owner relationship where product journey requires it
+ T2 WorkingContent OCC
→ ALLOW
```

## Submit / resubmit

```text
enabled User
+ document.submit at Document Area
+ Revision DRAFT
+ accepted author/owner relationship
+ exact expected WorkingContent generation
+ T2 submit invariants
→ ALLOW
```

## Withdraw

```text
enabled User
+ document.submit at Document Area
+ authorized author/submitter relationship
+ exact active pre-Release Submission
+ T2 withdrawal eligibility
→ ALLOW
```

## Governance decision

```text
enabled User
+ governance.act at Document Area
+ exact active GovernanceAttempt
+ exact active Step
+ User is in frozen candidate set
+ current Authorization still valid
+ User != Submission submitter / obsolescence initiator
→ ALLOW
```

For Group-routed Steps, being added to the Group after activation does not add the User to the frozen candidate set. Being removed from the Group may remove the group-mediated `governance.act`; another current qualifying grant can still satisfy the permission side if the User remains in the frozen candidate set.

## Cancel Revision

```text
enabled User
+ document.cancel_revision at Document Area
+ T2 DRAFT/pre-Release cancellation eligibility
→ ALLOW
```

## Initiate obsolescence

```text
enabled User
+ document.obsolete at Document Area
+ current EFFECTIVE target
+ mandatory reason
+ no open replacement Revision
+ no active competing obsolescence
→ ALLOW initiation
```

Human completion still requires governance route participant predicates. `NoHumanApproval` follows ratified T1/T2 zero-human-Step behavior.

## Change responsible owner

```text
enabled User
+ document.owner.manage at Document Area
+ target owner relationship eligible
→ ALLOW
```

---

# 12. Offboarding / re-enable enforcement

Prior B2 offboarding semantics survive because T1 independently requires future access termination without rewriting history.

One local product-state transaction conceptually:

```text
disable User
revoke active ApplicationSessions
remove current GroupMemberships
remove direct User RoleAssignments
append required Audit evidence
commit
```

Group RoleAssignments remain because they belong to the Group; removed membership means the offboarded User no longer inherits them.

Provider-side disable/reconciliation, if required, occurs outside the local transaction through later mechanism design.

Re-enable:

```text
re-enable same stable User identity
```

does **not** restore old Sessions, GroupMemberships or RoleAssignments. Fresh access must be explicitly granted.

Historical Submission/Decision/Release/Audit actor references remain unchanged.

---

# 13. Audit approach — alternatives

## Alternative A — generic DB CRUD triggers

Reject.

Database row mutation cannot reliably express business meaning such as:

```text
RETURN vs withdraw
first Release vs replacement Release
role revoke vs data cleanup
obsolescence initiation vs completion
```

It also encourages request-body/row copying and duplicated domain authority.

## Alternative B — event sourcing / deployment-wide cryptographic chain

Reject for Launch.

Would reintroduce a second state model and/or global serialization without a named assurance requirement, contrary to accepted GCR A9/T1.

## Alternative C — explicit semantic AuditEvent in the same local transaction for a bounded critical census

**Recommended.**

```text
BEGIN
  authoritative business/security mutation
  owning-domain immutable evidence where applicable
  required AuditEvent
COMMIT
```

If a required AuditEvent cannot be committed, that critical operation does not report success.

---

# 14. Audit minimum enforcement model

AuditEvent remains append-only supporting evidence.

Minimum conceptual facts:

```text
event identity
trusted server occurrence time
actor kind = USER | SYSTEM
stable actor User id or bounded System actor code
operation code
resource kind + stable resource id
visibility scope = Company | Area(area_id)
bounded operation-specific facts
correlation id when materially useful
```

`visibility scope` is historical evidence for least-privilege Audit read. It is not current resource authority.

## 14.1 Scope classification

Examples:

```text
Document lifecycle event
→ Document Area

Area-scoped RoleAssignment grant/revoke
→ that Area

Company-scoped RoleAssignment
→ Company

GroupMembership add/remove
→ Company
  // membership may affect grants across more than one Area

User offboard/re-enable
→ Company

DocumentType/governance-route config
→ Company
```

Area-scoped `governance_viewer` therefore receives only Audit events whose immutable visibility scope matches that Area, plus domain history it is separately authorized to read.

---

# 15. Required same-local-commit Audit census

T3 recommends mandatory Audit for the following accepted Launch mutations.

## 15.1 Authentication / Organization / Authorization security mutations

```text
provider-subject binding accept/disable/replace when product-owned operation exists
User create
offboard User
re-enable User
Area create/rename/retire/re-enable
Group create/rename/delete
GroupMembership add/remove
RoleAssignment grant/revoke
```

Ordinary profile enrichment edit is not mandatory semantic Audit by default because display name/email are erasable, non-authoritative attributes. A future compliance requirement may reopen it.

## 15.2 Controlled Documents configuration / responsibility

```text
DocumentType create/update/activate/deactivate
current governance route/config change
template role / eligibility configuration change
Document responsible-owner change
```

## 15.3 Controlled Documents lifecycle/governance

```text
Document + REV000 creation
REV001+ creation
Submission creation
ACCEPT
RETURN_FOR_CHANGES
withdraw Submission attempt
Revision cancellation
OfficialRendition semantic confirmation when one is required
Release
obsolescence initiation
obsolescence completion
```

Where the owning domain already has immutable evidence (`Submission`, Decision, `RevisionCancellation`, `Release`, Obsolescence result, OfficialRendition), Audit references that evidence rather than duplicating its full payload.

---

# 16. Deliberately not mandatory semantic Audit in Launch

```text
every WorkingContent mutation/autosave
every keystroke/editor event
Search query
read current document
history read
download
preview/render request
ordinary login/logout
notification delivery
SubmissionFeedback as a second duplicate Audit event
```

`SubmissionFeedback` is already immutable domain evidence with actor/time/context. Audit does not duplicate it solely for volume.

A named regulatory/customer requirement such as “every controlled-document download must be audited” is a direct reopen trigger.

Security telemetry (IP, user-agent, failed login, rate-limit events, provider security signals) remains observability/security telemetry unless promoted by a concrete Audit requirement.

---

# 17. Audit PII minimization

Audit must not copy by convenience:

```text
passwords/tokens/session secrets
provider claims
request/response bodies
full document content
Submission bytes
free-form feedback/reason text
email/display name
IP/user-agent by default
```

Use stable semantic identifiers and bounded operation-specific codes.

Examples:

RoleAssignment revoke may preserve enough facts to reconstruct the deleted current grant:

```text
assignment id
subject kind + stable subject id
role code
scope kind + stable scope id
```

GroupMembership removal may preserve:

```text
User id
Group id
```

A governance decision AuditEvent references the immutable Decision id and outcome code; rationale/feedback remains Controlled Documents evidence.

Erasing `UserProfile` must not rewrite Audit actor identity.

---

# 18. Future-evolution compatibility

| Future capability | T3 seam |
|---|---|
| Distribution | add explicit distribution permissions/roles; current reader bundle is not silently broadened |
| Periodic Review | add review operation permission only when capability returns |
| Evidence | independent permissions/lifecycle; Group/User/Company/Area grant model reusable |
| Records/Hold/Disposition | dedicated permissions added deliberately; no current area_manager/admin bundle silently gains them |
| Governed Export | explicit export permission added with concrete completeness/security contract |
| Repository connector | explicit import/publish permissions only when promoted |
| Training | independent training roles/permissions; reader does not imply training admin |
| Change Control | orchestration permissions may be added without becoming Document lifecycle bypass |
| pooled tenancy | Company anchor remains; T3 does not embed shared-tenant policy engine |
| CRDT | no AuthZ rewrite; DRAFT WorkingContent mechanism changes only |

Binding future privilege law:

> **Adding a material permission to an existing static role bundle is itself an authority change and requires explicit product migration/adjudication; future capabilities never silently widen existing Launch roles.**

---

# 19. Proof obligations before implementation

Later implementation design/tests must falsifiably prove at least:

```text
provider role/group cannot grant MetalDocs permission
revoked direct grant disappears from next canonical check
removed GroupMembership removes group-mediated authority from next canonical check
new Group membership cannot join an already activated T2 governance Step candidate snapshot
frozen Step candidate still cannot act after losing all current governance.act grants
offboarded User cannot retain valid Session or current direct/group-mediated authority
re-enable does not silently restore previous access
Company grant applies to Area-targeted operation; unrelated Area grant does not
area_manager cannot perform Company admin operations
governance_admin cannot read governed content without an explicit content role
governance.act cannot bypass active Step / frozen candidate / T2 SoD predicates
viewer cannot read DRAFT/SUBMITTED/history
area governance_viewer cannot read Company/other-Area Audit events
critical mutation rolls back if required same-commit Audit cannot be appended
Audit deletion/update through normal serving path is impossible
Audit reason/content/PII minimization rules are enforced
```

---

# 20. Explicit non-decisions

T3 does not decide:

```text
final table/check/index syntax
exact repository/package/check middleware layout
JWT/cache implementation
API endpoint names/error envelopes
frontend role-management screens
manual session-management UI
manual provider-binding administration UI
Audit export
read/download Audit requirement absent a named consumer
Audit cryptographic hash chain
custom roles/custom permissions
nested/dynamic groups
role inheritance
explicit deny rules
scheduled/temporary grants
Area hierarchy
```

---

# 21. Reopen triggers

Reopen the implicated decision only on concrete evidence that:

- Launch needs a role beyond the six static roles to avoid materially unsafe overgranting;
- `area_manager` has no actual operational consumer or needs materially different authority;
- a customer requires custom roles/permissions;
- Area-scoped access is insufficient and a new resource scope is genuinely required;
- Group membership administration must be delegated separately from RoleAssignment administration;
- mandatory download/read auditing is contractually/regulatorily required;
- Audit tamper-evidence requires a cryptographic chain/signature assurance level;
- fresh-auth/eSignature becomes required for a named governance decision;
- future capability promotion creates a real permission/bundle migration requirement.

---

# 22. T3 operator adjudication packet

Recommended dispositions:

```text
T3-A ACCEPT — preserve prior simple AuthZ primitives: flat Groups, current GroupMembership, RoleAssignment(User|Group, Company|Area), static product catalogs, additive/default-deny live evaluation; no provider-role bridge/ACL/ReBAC/custom-role platform.

T3-B ACCEPT — six composable static roles: governance_admin, area_manager, author, approver, viewer, governance_viewer; prior role concepts survive where still useful, governance_viewer is added for least-privilege audit/history access.

T3-C ACCEPT — 15 Launch permissions in §7; no dormant Launch+/Future permissions and no old Approval-owner reassign/oversee/cancel catalog.

T3-D ACCEPT — role bundles/scope matrix in §8; all six roles may be assigned to User or Group; governance_admin is Company-only; area_manager Area-only; author/approver/viewer/governance_viewer Company|Area.

T3-E ACCEPT — organization.manage owns organizational identity/lifecycle operations; access.manage is required for both RoleAssignment and GroupMembership mutation because membership may change effective authority.

T3-F ACCEPT — canonical Authorization is live direct + group-mediated grants → static bundle → scope → Controlled Documents predicates; Session/provider/caches never become durable authority.

T3-G ACCEPT — governance.act never grants generic approval access; exact active attempt/Step, frozen candidate set, current AuthZ, enabled User and T2 initiator/submitter SoD must all pass.

T3-H ACCEPT — offboarding disables User, revokes Sessions, removes current memberships/direct grants in one local semantic transaction; re-enable preserves identity but restores no old access automatically.

T3-I ACCEPT — Group deletion is blocked by live memberships, group grants or live route/config references; historical activated Step truth relies on frozen concrete candidate Users rather than requiring Group survival.

T3-J ACCEPT — Audit uses explicit semantic append-only AuditEvent, not CRUD triggers/event sourcing/global hash-chain; bounded critical operations require same-local-commit Audit.

T3-K ACCEPT — AuditEvent carries immutable Company|Area visibility attribution sufficient for least-privilege governance_viewer access; security/global events remain Company-scoped.

T3-L ACCEPT — same-local-commit Audit census in §15; autosave/search/read/download/login/logout are not mandatory semantic Audit events in Launch absent a named requirement.

T3-M ACCEPT — Audit is PII-minimized and references owning-domain evidence rather than duplicating reasons/comments/content; deleted current grant/membership evidence retains only bounded stable IDs/codes needed for reconstruction.

T3-N ACCEPT — future capability permissions are added deliberately; adding material permission to an existing role bundle is an authority migration and may not happen silently.
```

T3 remains **non-authoritative** until the operator adjudicates these recommendations. After adjudication, the mandatory platform-facing T3 summary must be presented and explicitly ratified before T3 can close or T4 open.

Implementation remains **BLOCKED**.
