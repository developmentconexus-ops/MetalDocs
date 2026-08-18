# R10-T3 — Operator Adjudication / Summary Ratification Gate

> **Status:** ACTIVE STAGING — T3 DECISIONS ADJUDICATED / PLATFORM SUMMARY RATIFICATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Candidate:** `docs/superpowers/analysis/2026-08-18-r10-t3-authorization-audit-enforcement-candidate.md`  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **Implementation:** BLOCKED

This record captures the operator adjudication of T3 after the candidate was rebuilt from the operator-ratified Decision Registry. It does **not** close T3 or open T4. T3 closes only after the operator explicitly ratifies the required platform-facing T3 summary.

## 1. Preserved baseline — not re-decided

T3 consumes the ratified registry baseline:

```text
Group + GroupMembership
RoleAssignment subject = User | Group
RoleAssignment scope = Company | Area
static product-owned Role / Permission vocabularies
additive grants + default deny
live direct User grants + live Group-mediated grants
provider roles/groups/claims never canonical Authorization
Session/JWT is not durable Role/Permission authority
Controlled Documents owns relationship/lifecycle/governance predicates
no role bypasses domain governance
Area Manager remains a preserved operational role concept
offboarding preserves historical User identity
re-enable never silently restores old grants/memberships/sessions
AuditEvent is evidence, not current state
critical governed/security mutations require same-local-commit Audit in principle
Audit is append-only + PII-minimized
no global AuditChainHead/hash-chain Launch requirement
```

## 2. T3 adjudication

The operator accepted all T3 recommendations as written:

```text
T3-A ACCEPT — preserve canonical live RBAC kernel: User|Group RoleAssignments, Company|Area scopes, static product catalogs, additive grants/default deny + domain predicates.
T3-B ACCEPT — Launch roles = governance_admin, area_manager, author, approver, viewer, governance_viewer.
T3-C ACCEPT — Launch permission catalog = 15 permissions; no old/future capability permissions.
T3-D ACCEPT — accepted role bundles: governance_admin configuration-only by default; area_manager Area-operational; author bounded authoring; approver exact-case governance; viewer effective-read; governance_viewer read-only governance/audit.
T3-E ACCEPT — every role may be assigned to User or Group; governance_admin = CompanyScope only; area_manager = AreaScope only; author/approver/viewer/governance_viewer = CompanyScope | AreaScope.
T3-F ACCEPT — organization.manage governs User/Area/Group identity lifecycle; access.manage governs GroupMembership + RoleAssignment mutations; no Area-local RBAC administrator in Launch.
T3-G ACCEPT — Group deletion fails while memberships, Group grants, current governance routes or live unactivated GROUP-step snapshots still depend on it; historical resolved/completed Steps do not keep Group alive solely for attribution.
T3-H ACCEPT — authoring relationship predicate = current responsible owner OR actor also has document.owner.manage; Area Manager can manage Area work without ordinary Author receiving blanket Area edit authority.
T3-I ACCEPT — governance action requires current governance.act + exact active-Step candidate snapshot + enabled User + T2 self-approval predicates; permission alone never grants a verdict.
T3-J ACCEPT — offboarding atomically disables User, revokes Sessions, removes memberships/direct grants, appends required Audit and optionally records durable provider-disable intent; re-enable restores no old access.
T3-K ACCEPT — governed/security-changing actions and offboarding total-order on current User eligibility; no claim of magical cancellation for already-linearized work.
T3-L ACCEPT — Audit = explicit semantic append-only evidence in the same local transaction for the bounded critical census; no CRUD-trigger inference, event sourcing or global hash-chain serialization.
T3-M ACCEPT — AuditEvent minimum = actor, trusted time, operation/resource attribution, immutable Company|Area visibility snapshot, bounded schema/facts and correlation when useful; PII/free-form domain payload stays out.
T3-N ACCEPT — audit.read may be Company- or Area-scoped; Area viewer sees only events attributed to that Area, not Company-wide access/identity administration.
T3-O ACCEPT — autosave/search/read/download/login/logout/notification/preview/deny telemetry is not mandatory semantic Audit in Launch; reopen on named compliance requirement.
T3-P ACCEPT — future capability permissions never silently broaden existing role bundles; material authority expansion requires explicit future design/migration.
```

## 3. Accepted role vocabulary

```text
governance_admin
area_manager
author
approver
viewer
governance_viewer
```

Roles remain static, product-owned, additive, non-hierarchical and assignable to User or Group subject according to the accepted scope matrix. No role is a superuser/domain-governance bypass.

## 4. Accepted permission vocabulary

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

## 5. Accepted role bundles

```text
viewer
  document.read_effective

author
  document.read_effective
  document.read_history
  document.read_working
  document.create
  document.edit
  document.submit

approver
  document.read_effective
  governance.act

area_manager
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

governance_admin
  organization.manage
  access.manage
  document_type.manage
  template_use.manage

governance_viewer
  document.read_effective
  document.read_history
  audit.read
```

## 6. Scope compatibility

```text
governance_admin   → CompanyScope only
area_manager       → AreaScope only
author             → CompanyScope | AreaScope
approver           → CompanyScope | AreaScope
viewer             → CompanyScope | AreaScope
governance_viewer  → CompanyScope | AreaScope
```

All accepted roles may be assigned to either a User or a Group. No special direct-User-only admin restriction is introduced.

## 7. Current gate

```text
T3 material decisions       = ADJUDICATED / ACCEPTED
T3 platform summary         = NEXT
T3 final closure/promotion  = PENDING SUMMARY RATIFICATION
Decision Registry update    = PENDING T3 CLOSURE
T4                          = NOT OPEN
implementation              = BLOCKED
```

Per the operator-approved stage protocol:

```text
T3 design
→ T3 adjudication ✅
→ platform-facing T3 summary NEXT
→ explicit operator summary ratification
→ promote/close T3
→ update Decision Registry
→ remove completed T3 staging
→ only then open T4
```
