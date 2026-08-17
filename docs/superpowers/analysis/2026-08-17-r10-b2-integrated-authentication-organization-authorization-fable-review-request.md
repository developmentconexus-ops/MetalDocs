# MetalDocs R10-B2 — Integrated Authentication / Organization / Authorization — Full Independent Review Request

> **Status:** INTEGRATED CANDIDATE / FULL REVIEW REQUEST — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Current branch evidence HEAD before this packet:** `42c61644102d249fe95892c530c94a709b5c4c31`
> **Authority baseline:** `71791dfecd4cd185684373ffcdccbf256138b741` — R10-B2-1 promotion
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Authority note:** this packet is a batch-level candidate for independent falsification. It does not amend R10 authority, handoff, program authority, ledger, code, schema, OpenAPI, frontend or Keycloak configuration.

---

# 0. Why this review is intentionally integrated

The repository has reached a stronger architecture baseline. Reviewing every field or local refinement as a separate promotion gate now risks optimizing certainty locally while hiding cross-family defects and creating ceremony without proportional decision value.

This packet therefore batches the remaining R10-B2 work and asks one independent reviewer to attack the **whole Authentication + Organization + Authorization system** as an integrated state/transaction/security model.

The reviewer MUST still be rigorous. Batch mode means **larger reasoning unit**, not weaker proof.

R10-B2-1 is already promoted authority and must be consumed rather than casually reopened. The B2-2 candidate and its independent review are evidence only. This integrated packet absorbs the single material B2-2 correction (`Area.disabled_at`) and proposes B2-3 + B2-4 together so their semantics can be judged end-to-end.

The review question is:

> **Does this integrated B2 candidate form the smallest sustainable Authentication / Organization / Authorization architecture that preserves frozen product semantics, single-company deployment law, fail-closed access, privacy separation, transaction correctness and future-safe seams without rebuilding a generic IAM platform?**

---

# 1. Mandatory authority baseline

Read `AGENTS.md` and follow its complete authority chain before reviewing this packet.

Promoted authority already fixes:

```text
R9.5 = FROZEN / GCR-REFINED / SINGLE-COMPANY-REFINED
R10-A = CLOSED / APPROVED
R10-B1 = CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED
R10-B2-1 = CLOSED / APPROVED
implementation = BLOCKED
```

The following are binding and this candidate cannot silently redefine them:

```text
one company per deployment V1
one singleton Tenant semantic root per product DB
Tenant.id immutable
expected_tenant_id ↔ DB Tenant.id fail-closed startup/readiness handshake
no universal tenant_id/company_id/deployment_id partition column
no Tenant/Area/role/Permission RLS as canonical Authorization
no customer selector/switching/routing
Authentication != Organization != Authorization
Keycloak = V1 Authentication provider
provider roles/groups/orgs/permissions/arbitrary claims never canonical AuthZ
no cross-provider DB atomicity / XA / 2PC

Organization V1:
  Tenant
  Area
  User
  Group
  GroupMembership

Authorization V1:
  Permission
  Role
  RoleAssignment
  subject = User | Group
  scope = TenantScope | AreaScope
  additive/default-deny
  grant/revocation evidence

five roles exactly:
  tenant_owner
  area_manager
  author
  approver
  viewer

permission catalog exactly:
  27 R9 base + 16 R9.5 = 43

no tenant_owner bypass
no deny engine
no generic ACL/ReBAC graph
no nested groups
no provider permission engine
```

Authorization equation remains:

```text
Permission
+ required case/resource relationship
+ Domain Governance constraints
= ALLOW
```

R10-B1 database law remains:

```text
ordinary durable entity identity = UUID PK
ordinary typed FKs
cross-owner FK actions = RESTRICT | NO ACTION only
TEXT + CHECK for frozen vocabulary by default
default isolation = READ COMMITTED
one local MetalDocs PostgreSQL transaction for required atomic cross-owner changes
required Audit + required durable intent share the authoritative local commit
external/provider effects happen after commit and reconcile idempotently
```

---

# 2. Evidence chain already available

## B2-1 — promoted

Candidate:
`docs/superpowers/analysis/2026-08-17-r10-b2-1-authentication-binding-session-assurance-fable-review-request.md` @ `9cba3acd`

Independent review:
`...-independent-fable-review.md` @ `361f6c8b`

Corrected target:
`...-adjudicated-corrected-target.md` @ `ee0a0ce0`

Bounded delta review:
`...-corrected-target-fable-delta-review.md` @ `6593c471`

Promotion:
`71791dfecd4cd185684373ffcdccbf256138b741`

B2-1 promoted shape:

```text
ProviderSubjectBinding
  id UUID PK
  user_id UUID FK → Organization.User
  issuer
  subject
  created_at
  disabled_at?

  UNIQUE(issuer, subject)
  UNIQUE(user_id) WHERE disabled_at IS NULL

ApplicationSession
  id UUID PK
  subject_binding_id UUID FK
  credential_digest
  created_at
  expires_at
  revoked_at?
  latest_reauthenticated_at?
  latest_provider_auth_time?
  latest_acr?
  latest_amr?
```

Important promoted B2-1 laws:

```text
one currently accepted ProviderSubjectBinding per User V1
issuer+subject stable identity
binding human attributes never identity authority
local opaque ApplicationSession
raw bearer never stored
finite Session expiry
no AuthZ snapshot in Session
no provider-token authority in Session
fresh-auth non-NULL state never automatically satisfies reauth
provider-only disable has bounded local-session staleness
offboarding locally revokes Sessions without provider dependency
binding disable/re-enable/replacement races with issuance must serialize
```

## B2-2 — independent evidence, not authority

Candidate:
`docs/superpowers/analysis/2026-08-17-r10-b2-2-organization-singleton-people-groups-fable-review-request.md` @ `8dfffb65`

Independent review:
`docs/superpowers/analysis/2026-08-17-r10-b2-2-organization-singleton-people-groups-independent-fable-review.md` @ `42c61644`

Verdict:

```text
APPROVE ... WITH MATERIAL FIXES
BLOCKER = 0
MAJOR   = 1
LOW     = 5
```

The only MAJOR was missing Area retirement. The review approved:

```text
Tenant constant-expression singleton enforcement
User + UserProfile privacy split
User disabled_at lifecycle
no User home_area
flat Group
GroupMembership current-only pair key
no surrogate GroupMembership UUID
```

This integrated candidate absorbs the Area correction and the relevant successor notes; there is intentionally no separate B2-2 corrected-target/delta cycle before this full batch review.

---

# 3. Integrated target invariants

## 3.1 Authority separation

```text
Keycloak/provider
  owns credentials/provider authentication mechanics

Authentication
  owns provider-subject correlation + local Session + assurance

Organization
  owns company / Area / User / Group identity and current relationships

Authorization
  owns fixed product Role/Permission semantics + current grants + evaluation
```

A User is not a credential, provider account, role bundle or Session.

A Group is not an Authorization scope.

An Area is organizational/business identity reused by domains; Area does not own permission meaning.

Permission/Role product vocabulary is not customer-configurable IAM metadata.

## 3.2 Current-state vs history

Current mutable access truth lives in Organization/Authorization rows.

Mutation history/evidence lives in Audit.

The target does not retain current-state rows solely to simulate an authorization event store unless a real historical consumer requires it.

## 3.3 Default deny

No assignment / membership / provider binding / Session / permission path may create access by inference from human attributes or stale snapshots.

---

# 4. Organization candidate

## 4.1 Tenant

```text
Tenant
  id           UUID PRIMARY KEY
  display_name TEXT NOT NULL
```

Laws:

```text
Tenant.id immutable
Tenant.display_name mutable company identity/settings fact
no slug/status/customer lifecycle/generic settings JSON V1
```

Structural singleton:

```text
CREATE UNIQUE INDEX ... ON tenant ((true))
```

This proves **at most one** row.

The already-promoted startup/readiness handshake proves **at least one** for every serving deployment.

Combined serving invariant:

```text
exactly one Tenant root
```

No fake singleton business column and no trigger unless evidence defeats the constant-expression index.

## 4.2 Area

Corrected target after the B2-2 independent review:

```text
Area
  id          UUID PRIMARY KEY
  code        TEXT NOT NULL
  name        TEXT NOT NULL
  disabled_at TIMESTAMPTZ NULL

UNIQUE(code)
```

Laws:

```text
id immutable
code immutable V1
name mutable

disabled_at IS NULL     → Area accepts new references/assignments
disabled_at IS NOT NULL → Area retired; historical references remain valid
```

Candidate treats `disabled_at` as reversible for the same Area identity. Re-enable never rewrites historical references; it only permits future assignments again. Audit owns disable/re-enable transition history.

No hierarchy, description, owner_user_id, default approver role or generic Area metadata in Organization.

Area code canonicalization/format is implementation-spec detail fixed at creation; downstream numbering consumes the stored code verbatim.

Consumer law for retired Area:

```text
existing Documents / historical Approval / existing grants remain valid
new Document Area assignment fails closed at Controlled Information boundary
new AreaScope RoleAssignment fails closed at Authorization boundary
new future Approval policy reference to retired Area fails closed at Approval boundary
```

The full reviewer MUST attack whether reversible retirement is justified or whether the current archive evidence only justifies terminal retirement.

## 4.3 User

```text
User
  id          UUID PRIMARY KEY
  disabled_at TIMESTAMPTZ NULL
```

Laws:

```text
User.id immutable stable MetalDocs organizational participant identity
no username/email/provider subject/credential/roles/capabilities/tenant_id/home_area/employee key

disabled_at IS NULL     → organizationally eligible
disabled_at IS NOT NULL → organizationally ineligible
```

Disable/re-enable preserves User identity. Audit owns transition history.

No terminal user state is introduced. Lawful privacy cleanup is separate from eligibility lifecycle.

## 4.4 UserProfile

```text
UserProfile
  user_id      UUID PRIMARY KEY REFERENCES User(id)
  display_name TEXT NOT NULL
  email        TEXT NULL
```

No separate profile UUID.

Meaning:

```text
User        = stable governed organizational identity
UserProfile = current human-readable/contact enrichment
```

Profile row is erasable.

Normally eligible Users are profile-complete. Profile absence means either lawful enrichment erasure or a bounded provisioning transition; consumers must render a neutral/opaque fallback instead of fabricating human data.

Email/display_name remain attributes, never technical identity or provider-binding authority. `UNIQUE(email)` is not a target identity invariant.

## 4.5 Group

```text
Group
  id   UUID PRIMARY KEY
  name TEXT NOT NULL UNIQUE
```

Flat and company-wide V1.

No code, area_id, scope, provider group identity, hierarchy, dynamic rule or retirement lifecycle without new evidence.

Group UUID remains stable identity under rename.

## 4.6 GroupMembership

```text
GroupMembership
  user_id  UUID NOT NULL REFERENCES User(id)
  group_id UUID NOT NULL REFERENCES Group(id)

PRIMARY KEY (user_id, group_id)
```

Meaning:

```text
row exists → current membership
row absent → not currently a member
```

No surrogate UUID, joined_at/left_at interval, tombstone or nested membership.

Audit owns add/remove transition evidence. Approval/Distribution snapshot concrete Users when their own semantics require historical denominator/participants, so they never depend on historical GroupMembership rows.

The pair consists of two internal technical UUIDs and identifies a pure relationship fact; candidate treats this as compatible with B1's UUID identity law rather than an external/business composite key.

---

# 5. Authorization candidate

## 5.1 Role and Permission are static product catalogs, not configurable tables

Exactly five roles and 43 permissions are already frozen product semantics.

Candidate therefore does **not** persist:

```text
permissions table
roles table
role_permissions table
editable/custom role bundle state
```

Instead Authorization owns a versioned-with-product static catalog:

```text
Role code → frozen Permission bundle
```

The exact bundles are those already accepted in R9/R9.5; this batch does not redesign them.

Database persistence contains only current RoleAssignments using frozen `role_code` vocabulary.

Rationale:

- persisted editable catalog would create a second authority against frozen product semantics;
- it would create unsupported custom-role/custom-permission capability;
- current legacy 8-role/capability code is evidence of the old model, not target entitlement;
- Admin UI can read the product catalog without DB configurability.

The reviewer MUST attack whether any deployment-specific consumer requires Role/Permission rows despite fixed semantics.

## 5.2 RoleAssignment — the only Authorization semantic persistent family

Candidate:

```text
RoleAssignment
  id UUID PRIMARY KEY

  user_id  UUID NULL REFERENCES User(id)
  group_id UUID NULL REFERENCES Group(id)

  role_code TEXT NOT NULL

  tenant_scope_id UUID NULL REFERENCES Tenant(id)
  area_scope_id   UUID NULL REFERENCES Area(id)
```

Structural checks:

```text
exactly one subject:
  user_id XOR group_id

exactly one scope:
  tenant_scope_id XOR area_scope_id

role_code CHECK in:
  tenant_owner
  area_manager
  author
  approver
  viewer
```

No generic `subject_type/subject_id` or `scope_type/scope_id` polymorphic registry.

The closed typed union preserves real FKs to every possible subject/scope.

### Duplicate-grant backstops

Candidate uses four partial uniqueness constraints/indexes:

```text
UNIQUE(user_id,  role_code, tenant_scope_id)
  WHERE user_id IS NOT NULL AND tenant_scope_id IS NOT NULL

UNIQUE(user_id,  role_code, area_scope_id)
  WHERE user_id IS NOT NULL AND area_scope_id IS NOT NULL

UNIQUE(group_id, role_code, tenant_scope_id)
  WHERE group_id IS NOT NULL AND tenant_scope_id IS NOT NULL

UNIQUE(group_id, role_code, area_scope_id)
  WHERE group_id IS NOT NULL AND area_scope_id IS NOT NULL
```

So two semantically identical current grants cannot coexist.

### Grant mutation law

RoleAssignment represents **current grant truth**.

```text
INSERT row → grant exists
DELETE row → grant revoked
```

The grant shape is immutable while the row exists; changing subject/role/scope is revoke + new grant.

No `revoked_at`, `revoked_by`, effective interval or retained historical RoleAssignment row by default.

Required grant/revocation evidence is written to Audit in the same local commit.

Re-grant after revocation produces a new RoleAssignment id and new Audit evidence.

The reviewer MUST attack whether RoleAssignment actually needs its surrogate UUID if current truth is otherwise fully identified by subject+role+scope; unlike GroupMembership the candidate currently keeps UUID because each grant is administered/revoked as an independent Authorization fact and may be attributed in Audit.

## 5.3 TenantScope is a real semantic Tenant reference

`tenant_scope_id` is not universal tenant partitioning.

It represents the real product meaning:

```text
RoleAssignment applies to the whole company represented by this singleton Tenant root
```

This is an allowed semantic Tenant FK under B1.

`area_scope_id` represents a real AreaScope.

No magic UUID/sentinel means TenantScope.

## 5.4 Proposed role↔scope compatibility

Candidate proposes:

```text
tenant_owner → TenantScope only
area_manager → AreaScope only

author       → TenantScope | AreaScope
approver     → TenantScope | AreaScope
viewer       → TenantScope | AreaScope
```

This is NOT assumed frozen by the packet; it is a material B2-3 proposal that the full reviewer must verify against the accepted role meanings/bundles and end-to-end consumers.

Reject or correct it if frozen authority requires a different compatibility matrix.

## 5.5 No persisted effective permissions

Forbidden as semantic authority:

```text
user_permissions
effective_permissions
cached group-expanded grants
Session roles/permissions
materialized ACLs
provider-role mappings
```

Canonical evaluation derives current authority from:

```text
User direct RoleAssignments
UNION
current GroupMemberships → Group RoleAssignments

→ static Role → Permission bundle
→ scope match
→ owner/domain relationship predicate
→ owner/domain governance constraint
→ ALLOW or default DENY
```

A cache/projection may later exist only as rebuildable mechanism if evidence proves it necessary; it cannot become a second authority.

## 5.6 Scope application

Candidate rule:

```text
Tenant-wide target/check
  requires a qualifying TenantScope assignment

Area-targeted target/check
  may be satisfied by:
    qualifying TenantScope assignment
    OR matching AreaScope assignment
```

For domain resources, the semantic owner provides the relevant relationship/scope context and governance predicates. Authorization does not embed Document/Approval/Evidence lifecycle meaning.

This preserves the frozen equation:

```text
Permission + relationship + Domain Governance = ALLOW
```

## 5.7 Authorization administration boundaries

Candidate maps existing frozen admin permissions to state owners:

```text
tenant.settings.manage
  → Tenant editable identity/settings

organization.manage
  → Area/User/UserProfile/Group identity & lifecycle operations

access.manage
  → GroupMembership and RoleAssignment access configuration

session.manage
  → explicit administrative Session management
```

### Target-scope grant administration

Creating/revoking a RoleAssignment requires `access.manage` effective at the **target scope**:

```text
TenantScope access.manage
  → may manage TenantScope grants

AreaScope(A) access.manage
  → may manage AreaScope(A) grants
```

No Area A authority grants Area B or TenantScope access.

### GroupMembership administration

Because Groups are company-wide and membership can activate every RoleAssignment attached to the Group, candidate requires **TenantScope `access.manage`** to add/remove GroupMembership.

This deliberately prevents an area-scoped administrator from adding themselves or another User to a globally privileged Group.

The reviewer MUST attack whether this is too restrictive and whether a safe area-local membership model can exist without making Group scope/history complex.

Group identity create/rename/delete remains `organization.manage`; membership is `access.manage` because membership changes effective access composition.

## 5.8 Disabled targets

New direct RoleAssignment to a disabled User fails closed.

New GroupMembership for a disabled User fails closed.

New AreaScope RoleAssignment to a retired Area fails closed.

Existing AreaScope grants remain valid when Area is retired; retirement stops **new assignment/reference**, not historical/current access to resources already associated with that Area.

---

# 6. Integrated offboarding / re-enable target

## 6.1 Why offboarding is decided here, not left as a microdecision

When Organization, Authentication and Authorization are viewed together, retaining dormant memberships/grants across User disable creates a predictable re-enable privilege-restoration hazard.

The integrated candidate therefore closes the semantic policy now.

## 6.2 Offboarding = destructive removal of effective access configuration

Candidate local transaction:

```text
BEGIN

lock User eligibility row

User.disabled_at = now

revoke/delete all ApplicationSessions for User

delete all GroupMembership rows for User

delete all direct User RoleAssignments

append required Audit evidence

insert durable provider-disable intent IF provider-side disable is required/possible

COMMIT
```

ProviderSubjectBinding is **not** automatically disabled; the issuer+subject→User correlation remains truthful after employee offboarding, consistent with promoted B2-1.

Keycloak/provider disable runs after commit through R10-D durable execution/reconciliation.

Group RoleAssignments remain because they belong to the Group, not the departing User; removing GroupMembership removes the User's indirect access.

## 6.3 Re-enable never silently restores prior access

Candidate:

```text
BEGIN
lock User
clear User.disabled_at
append required Audit
insert durable provider-enable intent when applicable
COMMIT
```

After re-enable:

```text
no prior GroupMembership is restored automatically
no prior direct RoleAssignment is restored automatically
```

Default deny remains until fresh explicit memberships/grants are created.

Rationale:

```text
short leave convenience
<
fail-secure return after long absence / role change / rehire
```

Identity is reversible; authorization configuration is not implicitly resurrected.

The reviewer MUST attack whether destructive offboarding loses any configuration/history that current Audit cannot safely preserve, and whether there is an accepted product need for temporary suspension distinct from offboarding.

## 6.4 Privacy cleanup remains separate

Offboarding does not automatically erase UserProfile or provider correlation.

Lawful privacy cleanup may later, according to accepted privacy policy:

```text
erase ApplicationSession rows
erase ProviderSubjectBinding where lawful
erase UserProfile
retain User.id skeleton where governed references require it
```

No generic PrivacyCase/workflow is introduced.

---

# 7. B2-4 transaction / concurrency target

Default isolation remains B1 `READ COMMITTED`.

Use narrow row locks + DB constraints, not global SERIALIZABLE/advisory-lock platforms.

## 7.1 C1 — Session issuance vs User offboarding

Promoted invariant retained:

```text
Either issuance commits before offboarding and offboarding revokes it,
or offboarding commits first and issuance sees User disabled and creates nothing.

Forbidden:
offboarding success + concurrently surviving newly issued valid Session.
```

Candidate realization uses a shared/compatible lock on the User eligibility row during issuance and exclusive lock during offboarding.

## 7.2 C2 — Binding acceptance vs Session issuance

Promoted B2-1 law retained:

Session issuance validates currently accepted ProviderSubjectBinding under the binding row serialization boundary.

Binding disable/re-enable/replacement uses the same lock discipline; disable/replacement revokes affected Sessions in the same local transaction; re-enable never revives revoked Sessions.

## 7.3 C3 — GroupMembership add vs User offboarding

Membership add:

```text
lock/check User eligible
insert (user_id, group_id)
```

Offboarding:

```text
exclusive User lock
disable User
delete all User memberships
```

Valid total orders:

```text
membership first → offboarding removes it
offboarding first → add sees disabled User and fails
```

## 7.4 C4 — direct User RoleAssignment grant vs offboarding

Direct User grant:

```text
lock/check User eligible
validate target scope
insert RoleAssignment
```

Offboarding under exclusive User lock deletes all direct User RoleAssignments.

Same total-order invariant as C3.

## 7.5 C5 — Area retirement vs new AreaScope RoleAssignment

New AreaScope grant and Area disable/re-enable serialize on the same Area row.

Valid order:

```text
grant first → commit → Area retires; existing grant survives
OR
retirement first → new grant sees retired Area → fail closed
```

No grant may be born after Area retirement while believing the Area was active.

The same successor law must be consumed by B3/B4 when they create new Area references/policies.

## 7.6 C6 — re-enable races

User re-enable uses exclusive User lock. After re-enable, new access configuration may be granted normally, but no deleted grant/membership is resurrected by the re-enable transaction.

Area re-enable uses exclusive Area lock. New references become legal only after re-enable commits.

ProviderSubjectBinding re-enable remains under promoted B2-1 C3 discipline.

## 7.7 C7 — grant/member changes relative to one another

A Group RoleAssignment and GroupMembership need not be one transaction. Effective access becomes true when both current facts exist.

Each mutation must independently pass its administrative Authorization check and DB invariants.

The reviewer MUST attack whether any race can create privilege that neither administrator intended, especially:

```text
concurrent GroupMembership add + new Group RoleAssignment
concurrent GroupMembership remove + RoleAssignment change
group deletion + membership/grant mutations
```

## 7.8 Lock-order / deadlock challenge

Some operations may touch more than one lockable semantic row:

```text
Session issuance → User + Binding
Area-scoped direct User grant → User + Area
Offboarding → User + dependent Session/Membership/RoleAssignment rows
binding replacement → User/binding uniqueness surface + Sessions
```

This packet does not freeze exact SQL syntax, but B2-4 must leave an implementable deterministic lock-order law rather than hand-wave deadlock risk.

The reviewer MUST either:

1. confirm a simple deterministic order exists under the candidate shapes; or
2. identify the minimum corrected order/data-shape change required.

Do not solve with global SERIALIZABLE unless narrow locks/constraints genuinely fail.

## 7.9 In-flight request semantics after offboarding

The candidate guarantees fail-closed **future Session resolution / authorization decisions** after offboarding commits through revoked Sessions + disabled User eligibility.

It does **not** claim magical cancellation of a request that completed authentication/authorization before offboarding committed unless that particular business transaction shares a relevant lock.

The reviewer MUST judge whether this bounded revocation semantics is acceptable or whether frozen product/security requirements demand stronger linearization across every authenticated operation.

---

# 8. Persistence classification / mutation law

Candidate classifications:

```text
Tenant
  SEMANTIC AUTHORITY
  id immutable
  display_name mutable

Area
  SEMANTIC AUTHORITY
  id/code immutable
  name/disabled_at mutable

User
  SEMANTIC AUTHORITY
  id immutable
  disabled_at mutable

UserProfile
  SEMANTIC AUTHORITY — subordinate human-readable enrichment
  mutable
  erasable

Group
  SEMANTIC AUTHORITY
  id immutable
  name mutable

GroupMembership
  SEMANTIC AUTHORITY — current relationship fact
  INSERT / DELETE

RoleAssignment
  SEMANTIC AUTHORITY — current grant fact
  immutable grant shape while present
  INSERT / DELETE

ProviderSubjectBinding
  already promoted Authentication semantic authority

ApplicationSession
  already promoted Authentication semantic authority
```

The reviewer should correct terminology if a class is semantically wrong, but must distinguish vocabulary preference from a real ownership/correctness defect.

---

# 9. Audit / durable-intent law for B2

## 9.1 Candidate same-commit Audit rule

Administrative mutations that change organizational identity, eligibility, binding acceptance or effective access must append required Audit evidence in the same local transaction.

Includes at minimum:

```text
Tenant display identity/settings mutation
Area create/rename/disable/re-enable
User create/disable/re-enable
UserProfile create/update/erasure where governance requires identity evidence
Group create/rename/delete
GroupMembership add/remove
RoleAssignment grant/revoke
ProviderSubjectBinding create/disable/re-enable/replacement
administrative Session revocation
offboarding
```

Ordinary Session issuance/login and user-initiated logout are not automatically declared critical governed mutations merely because they exist; review whether security/audit frozen law requires additional cases.

Audit does not FK-depend on current Organization/Authentication/Authorization rows for historical validity. It records PII-minimized identifiers/facts according to B6 privacy proof.

## 9.2 Provider effects

No Keycloak/provider HTTP call occurs inside a MetalDocs DB transaction.

When a local authoritative mutation requires a future provider effect:

```text
business/semantic mutation
+ required Audit
+ durable provider-effect intent
COMMIT
```

Then R10-D executes/retries/reconciles.

Examples:

```text
User provisioning/provider create
User offboarding/provider disable
User re-enable/provider enable
binding reconciliation
```

No XA/2PC.

---

# 10. Privacy target

The integrated privacy strategy is structural, not a workflow platform.

## 10.1 Direct human-readable PII boundary

```text
UserProfile
ProviderSubjectBinding
ApplicationSession/support telemetry if later introduced
```

can be erased according to lawful lifecycle.

`User.id` remains the stable organizational skeleton used by governed history where deleting the root would break retained facts.

## 10.2 Governed historical evidence

Approval, Distribution, Controlled Information and Audit must retain only the accepted PII-minimized reference/evidence they own. They must not depend on `UserProfile` existence for historical validity.

B6 later proves exact surviving Audit fields; R10-C proves restore does not silently resurrect erased enrichment.

## 10.3 No generic privacy engine

No:

```text
PrivacyCase
UserErasureWorkflow
PII tombstone platform
per-User crypto key system
```

without a named requirement.

The reviewer MUST attack whether retaining bare `User.id` plus `disabled_at` is compatible with the already-frozen architectural privacy posture, without turning the review into jurisdiction-specific legal advice.

---

# 11. Authorization read/evaluation surface

The candidate expects Authorization to expose a canonical application/domain read surface conceptually equivalent to:

```text
ResolveCurrentGrants(user_id)
CheckPermission(user_id, permission, scope/context, relationship facts)
```

Exact API names are not frozen.

Evaluation reads:

```text
User eligibility/current principal context
current direct RoleAssignments
current GroupMemberships
current Group RoleAssignments
static role-permission catalog
scope match
consumer/domain relationship predicates
consumer/domain governance constraints
```

No separate persistent grant projection is required for correctness.

The reviewer MUST attack query complexity/performance only insofar as it can create a foreseeable structural dead end at realistic V1 scale; do not invent scale problems without evidence.

---

# 12. Candidate administrative permission matrix

This matrix is proposed and must be independently verified against frozen role bundles/permissions:

| Operation family | Required semantic permission candidate |
|---|---|
| Tenant display/settings mutation | `tenant.settings.manage` |
| Area/User/UserProfile/Group identity/lifecycle | `organization.manage` |
| GroupMembership add/remove | `access.manage` at TenantScope |
| RoleAssignment grant/revoke | `access.manage` at target assignment scope |
| explicit administrative Session revoke/manage | `session.manage` |
| User offboarding | `organization.manage`; composition may revoke access/Sessions without separately requiring actor to hold `access.manage`/`session.manage` because those are mandatory consequences of the authorized offboarding operation |
| User re-enable | `organization.manage`; no access is restored |

The reviewer MUST attack especially:

- whether offboarding under `organization.manage` may legitimately delete grants/memberships and revoke Sessions without requiring additional permissions;
- whether GroupMembership must always be TenantScope-managed;
- whether Group lifecycle is `organization.manage` vs `access.manage`;
- whether any operation enables privilege escalation through self-management or Group indirection;
- whether the fixed role bundles actually contain the necessary permissions at the proposed scopes.

Do not change the 43-permission catalog in B2. If a real missing semantic permission is discovered, classify it as a material reopen rather than silently inventing one.

---

# 13. End-to-end scenarios the reviewer MUST walk

## S1 — New employee/user provisioning

```text
create User + UserProfile
write Audit
write provider-provisioning intent
commit

provider creates/returns causally correlated subject
create ProviderSubjectBinding
Audit

explicit RoleAssignments / GroupMemberships may be configured
login creates ApplicationSession only from eligible User + accepted binding
```

No email auto-binding.

## S2 — Direct Area role grant

```text
actor has access.manage at Area A
User eligible
Area A active
insert RoleAssignment(User, role, AreaScope A)
Audit same commit
```

## S3 — Group-mediated grant

```text
Tenant access admin adds User to Group
Area access admin grants Group role at Area A
current conjunction confers access
```

Attack both operation orders.

## S4 — Area retirement

```text
Area A active
existing Documents and grants reference it
retire Area A
existing references remain valid
new Document/grant/policy references fail
```

Attack re-enable.

## S5 — User offboarding with direct + group + live Sessions

```text
User has direct grants, memberships, sessions, binding
offboard
→ disabled
→ Sessions revoked
→ memberships deleted
→ direct grants deleted
→ Audit + provider-disable intent same commit
→ binding correlation retained
```

Verify no access resurrection.

## S6 — User rehire/re-enable much later

```text
same User id
clear disabled_at
provider re-enable if needed
old memberships/grants remain absent
login possible after provider/binding state allows
Authorization default-denies until fresh grants
```

## S7 — Privacy cleanup after offboarding

```text
erase Session/Binding/Profile where lawful
retain User id skeleton + governed history
historical UI degrades to neutral actor fallback
restore cannot later resurrect erased profile
```

## S8 — Group deletion

```text
must first remove memberships and RoleAssignments
then delete Group
historical Approval/Distribution snapshots remain valid because they reference concrete Users, not live Group
```

## S9 — Provider outage

Existing Sessions continue under local rules; new login/reauth/provider effect fails/retries visibly; Authorization current-state remains local.

## S10 — Grant revoke while Session remains live

Delete RoleAssignment + Audit. Existing Session remains valid authentication but next canonical Authorization evaluation sees grant absent and denies.

## S11 — Group membership removal while Session remains live

Delete membership + Audit. Next canonical Authorization evaluation excludes Group grants; Session need not be revoked.

## S12 — Concurrent offboarding vs login/member-add/direct-grant

Prove total-order invariants under READ COMMITTED/narrow locks.

## S13 — Concurrent Area retirement vs new AreaScope grant

Prove no grant born after retirement from a stale active check.

## S14 — Retired Area existing access

Prove retirement does not accidentally become a deny-all kill switch for historical/current governed content.

## S15 — Duplicate grant attempts

Prove partial unique indexes reject semantic duplicates across all subject/scope variants.

## S16 — Actor attempts privilege escalation via Group

Area-scoped admin cannot add self to globally privileged company Group under candidate membership administration law.

## S17 — ProviderSubjectBinding disabled/re-enabled while User eligible

Consume promoted B2-1; ensure Authorization does not invent alternate identity paths.

## S18 — User disabled but stale grant rows due transaction failure attempt

Same transaction must prevent partial offboarding success.

---

# 14. Strong alternatives the reviewer MUST compare

Do not merely inspect the candidate. Compare credible alternatives.

## A — DB-configurable RBAC

```text
Role table
Permission table
RolePermission table
RoleAssignment
```

Question: does any real V1 consumer justify configurability, or is this a local maximum/generic IAM platform?

## B — Separate UserRoleAssignment / GroupRoleAssignment tables

Question: are two typed tables simpler/safer than one XOR-typed RoleAssignment?

## C — Generic subject/scope polymorphism

```text
subject_type + subject_id
scope_type + scope_id
```

Question: does it reduce complexity enough to justify losing real FKs and opening generic graph pressure?

## D — Retained revoked RoleAssignments / effective intervals

Question: does Audit + current-state row deletion lose any required historical authority?

## E — Preserve memberships/grants on offboarding

Question: is destructive access cleanup too aggressive, or is silent privilege resurrection on re-enable the stronger defect?

## F — Scoped Groups / Area-local Groups

Question: could this safely permit area-local membership administration, or does it introduce unnecessary Group scope semantics duplicating RoleAssignment?

## G — Materialized effective-permission projection

Question: is current live evaluation structurally insufficient at V1 scale, or would a projection create synchronization/second-authority complexity?

## H — Monolithic User with nullable/scrubbed PII

Re-attack despite prior review; determine whether the integrated Authorization/offboarding model changes the UserProfile conclusion.

---

# 15. DevelopmentConexus Method review contract

The independent reviewer MUST apply the Method explicitly, not perform a generic architecture opinion pass.

For each material finding use proportionally:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ Credible Alternatives
→ Local Maximum vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / Future Cost
→ Authority / Boundary
→ Enforcement
→ Proof Strategy
→ Adversarial Challenge
→ Decision
→ Reopen Triggers
```

Required specific Method passes:

### Structural Inversion

Answer:

> If MetalDocs had been designed from day one with Keycloak, one company per deployment, separate Organization/AuthZ and no legacy IAM tables, what B2 state and transaction laws would still necessarily exist?

### Subtractive pass

Ask:

> What can still be deleted from the integrated B2 target without weakening a real invariant?

Attack at least:

```text
Tenant.display_name
Area.disabled_at
Area reversibility
User.disabled_at
UserProfile table/email
Group table/lifecycle omission
GroupMembership explicit row
RoleAssignment.id
RoleAssignment current-only semantics
static Role/Permission catalogs
four partial unique indexes
TenantScope explicit FK
GroupMembership Tenant-only admin rule
offboarding grant/membership deletion
provider disable intent
each Audit same-commit category
```

### Authority duplication pass

Search for any two sources of truth for:

```text
person identity
provider identity
group membership
role grant
permission bundle
scope
User eligibility
Area retirement
Session validity
historical actor display
```

### Failure-class pass

Attack at least:

```text
privilege escalation
stale grants
stale memberships
re-enable privilege resurrection
session/grant races
offboarding partial commit
Area retirement race
deadlocks
provider uncertainty
privacy erasure/restore
historical-reference breakage
unbounded IAM configurability
```

---

# 16. Decision list for reviewer disposition

Disposition every item `ACCEPT / ACCEPT WITH FIX / REJECT / DEFER`, with material fixes made explicit.

```text
IB2-D01  Tenant = id + display_name only
IB2-D02  singleton = constant-expression unique + readiness at-least-one
IB2-D03  Area = id/code/name/disabled_at; no hierarchy
IB2-D04  Area code immutable; retirement blocks new references only
IB2-D05  Area disabled_at reversible
IB2-D06  User = id + disabled_at only
IB2-D07  User/UserProfile split
IB2-D08  email/username never technical identity; no UNIQUE(email)
IB2-D09  no User home_area
IB2-D10  Group = id + unique name; flat/company-wide/no lifecycle
IB2-D11  GroupMembership pair PK/current-only/no UUID/history
IB2-D12  Role/Permission catalogs are static product authority, not DB tables
IB2-D13  one persisted Authorization family = RoleAssignment
IB2-D14  RoleAssignment typed subject XOR + typed scope XOR
IB2-D15  RoleAssignment UUID retained
IB2-D16  RoleAssignment current-only INSERT/DELETE; Audit owns revoke history
IB2-D17  four duplicate-grant uniqueness backstops
IB2-D18  explicit Tenant FK is semantic TenantScope, not tenancy partitioning
IB2-D19  proposed role↔scope compatibility matrix
IB2-D20  no persisted effective permissions
IB2-D21  TenantScope grant can satisfy matching area-level checks; AreaScope only matching Area
IB2-D22  GroupMembership mutation requires TenantScope access.manage
IB2-D23  RoleAssignment mutation requires access.manage at target scope
IB2-D24  Group identity uses organization.manage; membership uses access.manage
IB2-D25  offboarding deletes Sessions + Memberships + direct User RoleAssignments atomically
IB2-D26  offboarding retains ProviderSubjectBinding correlation
IB2-D27  re-enable never restores deleted access configuration
IB2-D28  privacy cleanup deletes enrichment/auth state but retains User skeleton
IB2-D29  User lock serializes issuance/member-add/direct-grant/offboarding
IB2-D30  Area lock serializes retirement/re-enable vs new AreaScope grants
IB2-D31  Binding lock discipline remains promoted B2-1 law
IB2-D32  no transaction couples Group grant + GroupMembership; conjunction is live truth
IB2-D33  deterministic narrow-lock ordering is sufficient under READ COMMITTED
IB2-D34  same-commit Audit for material B2 admin/identity/access mutations
IB2-D35  provider durable intent shares local commit when future provider effect is required
IB2-D36  no cross-provider transaction
IB2-D37  in-flight request semantics are bounded; no claim of universal cancellation
IB2-D38  canonical Authorization evaluates live current state + static bundles + domain predicates
IB2-D39  retired Area existing grants/content remain usable; only new refs blocked
IB2-D40  no generic IAM/privacy/RBAC platform introduced
```

---

# 17. Required proof outputs

The review must answer, with concrete reasoning:

1. Is integrated B2 at a **Global Maximum**, or did batching hide an incorrect local maximum?
2. Is the six-family Organization shape still correct after Authorization/transactions are included?
3. Does `User + UserProfile` still beat monolithic User?
4. Is reversible Area retirement correct, or should retirement be terminal?
5. Is Group lifecycle still genuinely YAGNI?
6. Is pair-keyed GroupMembership still correct under access-management/offboarding semantics?
7. Do Roles/Permissions belong as static product catalogs rather than DB tables?
8. Does RoleAssignment need a UUID?
9. Is one XOR-typed RoleAssignment table superior to separate typed tables?
10. Is current-only grant state + Audit sufficient evidence?
11. Is the role↔scope compatibility matrix correct?
12. Is explicit `tenant_scope_id` the right representation of TenantScope?
13. Can an area-scoped admin privilege-escalate through Groups under the proposed administration model?
14. Is TenantScope-only management of GroupMembership too restrictive or correctly conservative?
15. Is destructive offboarding of memberships/direct grants the right fail-secure behavior?
16. Does re-enable/default-deny behavior create operational dead ends?
17. Does any accepted temporary-suspension use case require retained access config?
18. Can all declared concurrency invariants be implemented cheaply under READ COMMITTED?
19. Is there a deadlock cycle implied by User/Binding/Area locking?
20. Is the in-flight request revocation posture honest and sufficient?
21. Does Area retirement correctly preserve existing grants/content while blocking new refs?
22. Does same-commit Audit scope include too much or too little?
23. Is provider-disable/enable durable intent correctly coupled to User lifecycle without moving provider truth into Organization?
24. Can lawful privacy cleanup preserve governed history without a generic privacy engine?
25. Does any surviving `tenant_id`-like semantic column reappear accidentally?
26. Can any table/field/index/lifecycle still be deleted?
27. Does any real current implementation consumer prove a missing target fact?
28. Does the candidate preserve all 43-permission / five-role frozen semantics without inventing a sixth role or new permission?
29. Does it preserve strict default-deny and no-bypass laws?
30. Does B2 now close cleanly enough to proceed to B3 after one adjudication/delta cycle at most?

---

# 18. Scope controls

This is a **full integrated B2 review**, not a whole-product redesign restart.

You MAY reopen B2-1 only if the integrated candidate produces a concrete contradiction with promoted B2-1 authority.

You MAY identify successor obligations for B3/B4/B6/C/D/E/F when the batch exposes them, but do not design those stages beyond the minimum needed to prove B2 coherence.

Do NOT:

```text
invent custom roles/permissions
add pooled tenancy
add RLS as policy engine
add Keycloak role/group authority
add generic privacy workflow
add generic ACL/ReBAC graph
add implementation code/schema/OpenAPI/frontend
configure Keycloak
promote authority
merge
```

Current implementation/schema/runtime are evidence only.

---

# 19. Required verdict

Return exactly one primary verdict:

```text
APPROVE R10-B2 INTEGRATED AUTHENTICATION / ORGANIZATION / AUTHORIZATION TARGET
```

or

```text
APPROVE R10-B2 INTEGRATED AUTHENTICATION / ORGANIZATION / AUTHORIZATION TARGET
WITH MATERIAL FIXES
```

or

```text
DO NOT APPROVE R10-B2 INTEGRATED AUTHENTICATION / ORGANIZATION / AUTHORIZATION TARGET
```

Required output structure:

- verdict;
- BLOCKER / MAJOR / LOW counts;
- decision disposition `IB2-D01..IB2-D40`;
- Method/Global-Maximum verdict;
- Organization verdict;
- Authorization relational-model verdict;
- Role/Permission static-catalog verdict;
- RoleAssignment UUID/current-history verdict;
- role↔scope compatibility verdict;
- GroupMembership admin/security verdict;
- offboarding/re-enable verdict;
- Area retirement/re-enable verdict;
- canonical evaluation/default-deny verdict;
- transaction/concurrency/deadlock verdict;
- Audit/durable-intent verdict;
- privacy verdict;
- subtractive/YAGNI findings;
- current-implementation evidence findings classified as `KNOWN REQUIREMENT / LEGACY MECHANISM / ACCIDENTAL COMPLEXITY / UNKNOWN / DEFERRED`;
- any material B2-1 reopen;
- any material reopen outside B2;
- exact corrected integrated target if fixes are required;
- whether another broad review is required;
- whether a bounded delta review would be sufficient after corrections;
- whether R10-B2 would be promotable after operator adjudication.

Write the independent review artifact to:

`docs/superpowers/analysis/2026-08-17-r10-b2-integrated-authentication-organization-authorization-independent-fable-review.md`

Explicit authorization is limited to:

- creating that independent review artifact;
- committing it to the current branch;
- pushing to the same branch.

Do not alter candidate/authority/handoff/program/ledger or implementation surfaces.
