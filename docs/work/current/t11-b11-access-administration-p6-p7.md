# T11 — B11 Access Administration — P6 Reference Study + P7 Candidate

> **Status:** CANDIDATE / OPERATOR-APPROVED IN CHAT / WRITTEN SPEC AWAITING OPERATOR REVIEW.  
> **Block:** B11 — Access Administration.  
> **Route:** `/admin/access`.  
> **Method:** pinned DevelopmentConexus `METHOD.md + FRONTEND-METHOD.md`.  
> **Implementation:** BLOCKED.  
> **Artifact class:** temporary frontend-planning Evidence under `docs/work/**`; not Product/architecture authority and must be absent from the eventual merge candidate/main.

## 1. Purpose and authority boundary

B11 makes current accepted access configuration human-operable without creating a parallel Authorization engine or a screen-shaped backend.

Accepted human job:

```text
Administer access
→ manage GroupMembership
→ grant/revoke fixed RoleAssignment
```

Accepted permission ownership:

```text
organization.manage
  Company / User / Area / Group identity

access.manage
  GroupMembership add/remove
  RoleAssignment grant/revoke
```

B11 does not own User/Area/Group identity, role vocabulary, permission vocabulary, effective Authorization calculation, provider identity, document governance, Audit, or content access.

Primary B11 operations remain exactly:

```text
27 listGroupMembers
28 addGroupMember
29 removeGroupMember
30 listRoles
31 listRoleAssignments
32 createRoleAssignment
33 deleteRoleAssignment
```

Organization reads from B10 may be supporting reads when current server disclosure permits them; their use does not transfer ownership to B11.

## 2. P6 question

Reference study was triggered because B11 is security-sensitive and more than one materially plausible human organization exists. The study asks:

```text
How do mature access-administration products help an administrator understand:
- who receives authority;
- what authority is granted;
- where it applies;
- how group membership changes inherited authority;
- what a grant/revoke actually changes;
- how to avoid overstating effective access?
```

References are Evidence only. Visual style, feature breadth and provider-specific concepts are not authority for MetalDocs.

## 3. P6 official sources

### Microsoft Azure RBAC

Sources:

- https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments
- https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments-portal
- https://learn.microsoft.com/en-us/azure/role-based-access-control/check-access
- https://learn.microsoft.com/en-us/azure/role-based-access-control/scope-overview

**SOURCE OBSERVATION**

Azure describes a role assignment through three core dimensions: principal, role and scope. Scope constrains where the grant applies. The portal includes an explicit review step before assignment. Azure also separates grant configuration from a `Check access` experience that can show assignments at the current scope and inherited to it.

**INFERENCE**

A security-sensitive grant should make `who × what × where` perceptually distinct, and the review state should repeat those dimensions before mutation. Effective-access explanation is a separate user job from creating/revoking one assignment.

**METALDOCS DECISION PRESSURE**

Adopt the clarity of Subject × Role × Scope and a final review summary. Do not import Azure resource hierarchy, PIM, conditions, custom roles, time-bound grants or an effective-access explorer without independent MetalDocs need and authority.

### Google Cloud IAM

Sources:

- https://cloud.google.com/iam/docs/granting-changing-revoking-access
- https://docs.cloud.google.com/iam/docs/troubleshoot-policies

**SOURCE OBSERVATION**

Google Cloud grants a role to a principal at a selected resource and separately provides Policy Troubleshooter to answer whether/why a principal can use a permission on a resource.

**INFERENCE**

Grant management and access explanation are related but semantically distinct products. A configuration screen should not pretend that the removal of one grant proves removal of all effective access.

**METALDOCS DECISION PRESSURE**

Keep B11 focused on canonical membership/grant mutations. If P8 proves that administrators materially require explanation of effective access, route an upstream finding rather than computing it in the browser.

### GitHub Organizations / Teams

Sources:

- https://docs.github.com/en/organizations/organizing-members-into-teams/adding-organization-members-to-a-team
- https://docs.github.com/en/organizations/managing-user-access-to-your-organizations-repositories/managing-repository-roles

**SOURCE OBSERVATION**

When adding an organization member to a team, GitHub asks the administrator to review repositories that the new member will be able to access before confirmation. GitHub also distinguishes individual and team repository access.

**INFERENCE**

Group membership is security-bearing, not merely organization metadata. Consequence communication before membership mutation materially reduces surprise. Direct vs group-derived access should not be conflated.

**METALDOCS DECISION PRESSURE**

Preserve B10 Group identity vs B11 membership separation. Test whether consequence copy is sufficient without a full effective-access projection; do not invent one preemptively.

### Okta

Sources:

- https://help.okta.com/en-us/content/topics/security/administrators-group-membership-admin.htm
- https://help.okta.com/en-us/content/topics/security/administrators-admin-comparison.htm

**SOURCE OBSERVATION**

Okta has a distinct Group Membership Administrator role with the ability to add/remove members from managed groups while not owning group creation/deletion. Fixed administrator roles also make role meaning inspectable without requiring a custom role editor for every administrative task.

**INFERENCE**

Separating group identity from group membership is a legitimate least-privilege boundary. Fixed roles can be explained rather than edited.

**METALDOCS DECISION PRESSURE**

Retain `organization.manage` for Group identity and `access.manage` for GroupMembership. Explain the six fixed MetalDocs roles from server-returned `RoleView`; never create a Role/Permission builder.

## 4. P6 convergence

The sources converge on four useful principles:

```text
1. make Subject × Role × Scope explicit;
2. review the intended grant before mutation;
3. treat membership as authority-changing and communicate consequence;
4. separate configuring grants from explaining complete effective access.
```

They do **not** justify importing:

```text
custom Role builder
custom Permission editor
ABAC conditions
PIM / eligible / time-bound grants
nested or dynamic Groups
IdP Group mirroring
resource-set platform
bulk grants
access certification
permission simulator
generic IAM framework
```

Reference saturation was reached when additional mature products repeated the same principal/role/scope and inherited-access distinction without changing the MetalDocs decision space.

## 5. P7 credible alternatives

### A — Scope-first IAM

```text
Access
→ Company or Area
→ assignments at that scope
→ grant/revoke
```

**Strength:** directly answers “who has access here?” and resembles mature resource-scoped IAM consoles.

**Blocking mismatch:** current `listRoleAssignments` has no accepted scope filter. Filtering only one loaded page would fabricate completeness; crawling every page to manufacture a scope index would turn browser traversal into a hidden query engine.

**Disposition:** REJECTED under current authority. Reopen only if P8 proves scope-first administration is a material human need and current read authority is insufficient.

### B — Subject-first / Effective Access

```text
Access
→ User or Group
→ direct grants
→ group-derived grants
→ effective access explanation
```

**Strength:** excellent for “what can this person do?” and “why do they have access?”

**Blocking mismatch:** current wire has no effective-access or access-explanation read. Browser composition would be incomplete and could become a false Authorization authority.

**Disposition:** REJECTED as the current P7 structure. Preserve as a falsification hypothesis through B11-A3.

### C — Task-oriented Access Workspace

```text
/admin/access

Access
├── Memberships
└── Role grants
```

**Strength:** maps directly to the two accepted authority-changing jobs while keeping their semantics distinct. Supporting Organization reads supply selectable identity without creating a new directory owner. Subject × Role × Scope remains explicit inside the grant flow.

**Disposition:** LEADING CANDIDATE / OPERATOR-APPROVED IN CHAT.

## 6. Candidate structure

### 6.1 Access route

Stable route remains exactly:

```text
/admin/access
```

The accepted B01/B01N shell is inherited. B11 creates no new top-level Product space or stable route family.

Local B11 navigation:

```text
Access
├── Memberships
└── Role grants
```

The local labels are UX organization only; semantic ownership remains Organization for GroupMembership state and Authorization for RoleAssignment state, with both mutations guarded by `access.manage`.

### 6.2 Memberships

```text
Memberships

Groups
  paginated browse
  → select Group

Selected Group
  current members
  Add member
  Remove member
```

Primary operations:

```text
listGroupMembers
addGroupMember
removeGroupMember
```

Supporting reads may use accepted Group/User Organization views. No B11-specific User/Group directory API is invented.

#### Add member

The flow selects one existing User. `UserPage` exposes eligibility; a disabled User may be shown as unavailable in the selector, but the server remains mutation authority and revalidates security-sensitive current truth.

Confirmation must communicate bounded consequence without claiming complete effective access:

```text
Add <User> to <Group>?

Joining this group may give the user access through current or future
RoleAssignments granted to the group.

Cancel | Add to group
```

It must not claim the exact final permission set unless a later accepted read makes that claim truthful.

#### Remove member

Confirmation must communicate the narrow effect:

```text
Remove <User> from <Group>?

Access derived from this group will no longer apply after removal.
Other direct grants or grants through other groups may still apply.

Cancel | Remove from group
```

It must not claim “all access will be removed.”

### 6.3 Role grants

```text
Role grants

Current assignments
  paginated ledger

Grant access
  Subject
  Role
  Scope
  Review

Revoke one assignment
```

Primary operations:

```text
listRoles
listRoleAssignments
createRoleAssignment
deleteRoleAssignment
```

#### Grant composer

The composer exposes exactly three semantic dimensions:

```text
1. Subject
   USER | GROUP
   exact selected identity

2. Role
   one fixed RoleView
   show server-returned permissions as explanatory meaning

3. Scope
   COMPANY | AREA(area_id)
   only scope kinds admitted by the selected RoleView
```

Review repeats the exact intended command before mutation:

```text
Subject   <User or Group>
Role      <fixed Role>
Scope     <Company or Area>

Role allows:
<server-returned RoleView.permissions>

This creates an additive RoleAssignment.
Existing assignments are unchanged.

Cancel | Grant access
```

The permissions list explains the selected fixed Role. It is not a client-maintained permission matrix and never substitutes for server authorization.

`governance_admin` remains Company-only; `area_manager` remains Area-only; the other accepted roles admit Company and Area according to `RoleView.allowed_scope_kinds`.

#### Revoke

Revoke targets one exact assignment identity:

```text
Revoke grant?

Subject  <identity>
Role     <role>
Scope    <scope>

This revokes this RoleAssignment only.
The subject may still have access through other direct assignments
or Group memberships.

Cancel | Revoke
```

There is no “edit assignment” abstraction. Changing Subject/Role/Scope means an explicit revoke decision and a separate new grant decision. B11 does not pretend delete+create is one atomic edit.

## 7. State and authority classes

B11 uses only accepted frontend state classes:

```text
SERVER STATE
  Group pages, User/Area supporting pages, GroupMemberPage,
  RoleListView, RoleAssignmentPage

NAVIGATION / URL
  /admin/access + bounded local presentation state when useful

FORM DRAFT
  unaccepted Subject/Role/Scope grant composition

EPHEMERAL UI
  selected Group/assignment, open confirmation, focus, review scenario controls
```

No normalized global User/Group/Role/Permission entity store becomes authority.

## 8. Failure and retry intent for P8

P8 must make these classes operable where relevant:

```text
403 permission.denied
  route/action denied; never render successful empty access data

404 notfound.resource
  selected Group/assignment became absent or non-disclosable

409 state.conflict
  createRoleAssignment referential/current-state conflict; preserve deliberate grant draft for correction/review

422 validation.failed / validation.idempotency_key_reused
  keep grant draft; bind safe validation information without inventing rejected-value truth

ambiguous transport outcome for createRoleAssignment
  retry the same logical command with the same Idempotency-Key
  do not generate a second grant command silently

CSRF recovery
  re-bootstrap session/CSRF and retry only the same logical unsafe command when safe

membership mutation vs offboarding
  server serializes current eligibility; frontend never predicts the winner
```

A generic toast may not erase a failure that changes the administrator's safe next action.

## 9. Falsification targets

### B11-A1 — Findability

```text
Paginated browse is operationally sufficient to locate User / Group / Area /
RoleAssignment for Launch access administration without global search/filter APIs.
```

P8 must include deterministic fixture scale large enough to force real pagination. Forbidden proof shortcuts:

```text
page-local search presented as global search
client crawl of every page presented as a complete query
invented q/filter operation
```

Failure disposition: UPSTREAM FINDING to the smallest read-contract owner if operator Evidence proves the job materially impractical.

### B11-A2 — Membership consequence

```text
Bounded truthful consequence language is sufficient for safe membership add/remove
without a full effective-access projection.
```

P8 must force an administrator to decide membership changes on Groups that have plausible security-bearing assignments in the fixture narrative, while refusing to fabricate the exact derived permission result.

Failure disposition: identify the smallest missing semantic read needed to make the decision safe; do not simulate a complete answer in frontend state.

### B11-A3 — Access explanation

```text
Launch administrators can safely manage canonical memberships and grants without a
separate “why does this subject have access?” / effective-access explanation surface.
```

P8 must make direct grants and membership-mediated possibilities understandable while remaining honest about what it cannot prove. If the operator cannot confidently administer access without explanation, route an upstream finding rather than adding an inferred permission matrix.

## 10. Accessibility and responsive structure

P8 must prove at least:

```text
semantic page/local navigation landmarks
keyboard-operable Group/assignment selection
labeled Subject/Role/Scope controls
visible focus
confirmation dialogs with deterministic focus entry/return
status/error announcements
no security meaning conveyed by color alone
```

Narrow viewport:

```text
global shell follows accepted responsive behavior
local Memberships / Role grants navigation reflows/collapses
list/detail stacks rather than forcing horizontal compression
grant composer becomes a vertical sequence
review summary remains before the mutation action
```

No material security action may depend on hover.

## 11. YAGNI / explicit exclusions

B11 P7 excludes:

```text
custom Role or Permission editor
scope hierarchy browser
conditions / ABAC
PIM / time-bound / eligible grants
nested/dynamic/provider-mirrored Groups
bulk membership/grant mutation
access certification/review campaigns
permission simulator
client-side effective-access engine
generic IAM framework
generic reference-data service
new B11 search/filter/effective-access API without proven need
B12 Document Governance concerns
```

## 12. Global Maximum decision

Current Evidence supports:

```text
CURRENT STRUCTURE CONFIRMED
→ Task-oriented Access Workspace
→ Memberships + Role grants
→ Subject × Role × Scope review for grants
→ bounded consequence communication
→ no effective-access fabrication
```

Why this is the current Global Maximum:

- it preserves the essential security complexity already present in MetalDocs;
- it removes accidental complexity from enterprise IAM products that MetalDocs does not need;
- it uses current accepted owners and operations without screen-shaped backend additions;
- it keeps the three strongest unresolved UX risks falsifiable rather than suppressing them;
- every rejected richer capability can be added later through the smallest owning reopen if real Evidence proves it necessary.

Reopen triggers are material operator/P8 Evidence that A1, A2 or A3 fails; changed Launch scale; a new named security/compliance consumer; or changed Product authority. Preference for a richer IAM console is not a reopen trigger.

## 13. P7 exit gate

This written specification is ready for operator review, not P8 execution.

```text
P6 reference study             COMPLETE
credible alternatives          3 compared
leading candidate              Task-oriented Access Workspace
operator in-chat disposition   APPROVED
written-spec disposition       AWAITING OPERATOR REVIEW
proven upstream Findings       0
open falsifiable assumptions   B11-A1 / B11-A2 / B11-A3
P8                              NOT STARTED
LOCK                            NONE
```

Only after explicit operator approval of this written specification may P8 be materialized under `docs/work/current/*.html`.
