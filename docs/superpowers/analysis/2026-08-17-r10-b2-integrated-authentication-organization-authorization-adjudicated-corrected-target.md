# MetalDocs R10-B2 — Integrated Authentication / Organization / Authorization — Adjudicated Corrected Target

> **Status:** ADJUDICATED CORRECTED CANDIDATE — **PENDING ONE BOUNDED DELTA REVIEW — NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Promoted architecture baseline:** `71791dfecd4cd185684373ffcdccbf256138b741` — R10-B2-1 promotion
> **Integrated candidate:** `docs/superpowers/analysis/2026-08-17-r10-b2-integrated-authentication-organization-authorization-fable-review-request.md` @ `b814f67284badd00182ff3c0abb77a66b448d7c9`
> **Independent full review:** `docs/superpowers/analysis/2026-08-17-r10-b2-integrated-authentication-organization-authorization-independent-fable-review.md` @ `34a567fda37751c24bea878c27295964ed4f9757`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Authority note:** this file records the operator-approved adjudicated integrated B2 target for bounded delta review. It does not amend R10 authority, handoff, program authority, frozen ledger, code, schema, OpenAPI, frontend or Keycloak configuration.

---

# 1. Review result and operator adjudication

The independent batch-level review returned:

```text
APPROVE R10-B2 INTEGRATED AUTHENTICATION / ORGANIZATION / AUTHORIZATION TARGET
WITH MATERIAL FIXES

BLOCKER = 0
MAJOR   = 3
LOW     = 5

IB2 decisions:
  ACCEPT          = 35
  ACCEPT WITH FIX = 5
  REJECT          = 0
  DEFER           = 0

B2-1 reopen                    = NONE
broad review required          = NO
bounded delta after correction = SUFFICIENT
promotable after adjudication  = YES
```

The review found no topology/ownership failure. The integrated structure remains the Global Maximum; corrections are bounded enforcement/coherence laws.

Operator adjudication:

| Finding | Decision | Corrected target |
|---|---|---|
| M1 — role↔scope matrix not structurally enforced | **ACCEPT** | Keep the reviewed role↔scope matrix and make invalid pairs unrepresentable with a DB CHECK plus write-path validation. |
| M2 — deterministic B2 lock order absent | **ACCEPT** | Codify one canonical lock-class order, child-row ordering and PostgreSQL lock-mode law under B1 `READ COMMITTED`. |
| M3 — Group hard-delete ignores live cross-owner references | **ACCEPT / ROUTING REFINED** | Group deletion fails closed against live typed references. Known V1 consumers ApprovalPolicy `Group` actor rule and Distribution group audience are B4 obligations. B5 inherits only if it later introduces a real Group reference. |
| L1 — offboarding says revoke/delete Session | **ACCEPT** | Offboarding terminally revokes Sessions; physical Session erasure is a separate lifecycle/privacy operation. |
| L2 — first-admin bootstrap / last-admin lockout | **ACCEPT** | Initial `tenant_owner` seed and lockout recovery use the non-serving maintenance trust surface; never a request-path bypass. R10-E may add a UX last-admin guard. |
| L3 — static catalog ↔ DB CHECK parity | **ACCEPT** | Implementation proof must fail on role-vocabulary / role-scope CHECK drift. |
| L4 — full 5×43 bundle matrix missing from durable target | **ACCEPT** | This corrected target pins the exact current bundle matrix derived from the locked R9 base bundles + locked R9.5 additions + single-company removal of `tenant.export`/`tenant.deletion.request`. |
| L5 — display-name normalization | **ACCEPT / IMPLEMENTATION-SPEC** | Blank/trim/casing constraints are implementation details; no accidental case-insensitive identity semantics. |
| D15 note — RoleAssignment UUID rationale | **ACCEPT** | UUID is structurally required because the XOR subject/scope union has no single NULL-free composite PK; partial unique indexes cannot be the table PK. |

## 1.1 Bundle-consistency refinement discovered during L4 recovery

Recovering the exact frozen role bundles removes one hypothetical capability from the original integrated candidate:

```text
access.manage exists in the V1 bundle of tenant_owner only.
tenant_owner is TenantScope-only.
area_manager is explicitly not an RBAC administrator.
```

Therefore V1 has **no Area-local access administrator**. The original candidate sentence allowing `AreaScope(A) access.manage` to administer AreaScope grants is removed rather than preserved as a dead/hypothetical path.

Corrected administration law:

```text
all GroupMembership administration = TenantScope access.manage
all RoleAssignment administration  = TenantScope access.manage
```

Because the exact bundle grants `access.manage` only to `tenant_owner`, ordinary V1 request-path access administration is tenant-owner-only. This is smaller and is consistent with the already-locked role semantics.

M1's DB CHECK remains mandatory: an illegal role↔scope assignment must be structurally unrepresentable even if a bug, migration or privileged maintenance path attempts one. The prior review's `tenant_owner@AreaScope` escalation example is therefore reduced from a legal-request-path escalation to a structural-invalid-state threat, but the required backstop is unchanged.

This refinement is part of the bounded delta review scope; it is not silently promoted here.

---

# 2. Integrated B2 invariant

R10-B2 fixes the minimum product-owned state and transaction law for Authentication + Organization + Authorization:

> **A MetalDocs request acts as one eligible organizational User reached through one accepted provider binding and one valid local ApplicationSession. Effective product authority is derived live from current direct/group RoleAssignments over typed Tenant/Area scopes, static product Role→Permission bundles, domain-owned relationship predicates and domain governance constraints. Identity, authentication, organization, authorization, audit evidence and provider execution each have one owner.**

The batch preserves:

```text
one company per deployment
one singleton Tenant semantic root
no universal company partition column
no Tenant/Area/role/Permission RLS as canonical AuthZ
Keycloak = V1 AuthN provider
Authentication != Organization != Authorization
no provider role/group/org/permission authority
no cross-provider atomicity
additive grants + default deny
no tenant_owner bypass
no generic ACL/ReBAC/deny engine/nested groups
implementation = BLOCKED
```

B2-1 remains promoted authority and is consumed, not rewritten. Its two Authentication semantic families remain exactly:

```text
ProviderSubjectBinding
ApplicationSession
```

---

# 3. Organization — corrected integrated target

## 3.1 Tenant

```text
Tenant
  id           UUID PRIMARY KEY
  display_name TEXT NOT NULL
```

Laws:

```text
Tenant.id = immutable deployment↔DB trust anchor
Tenant.display_name = mutable company identity/settings fact
no slug/status/customer lifecycle/generic settings JSON V1
```

Structural at-most-one enforcement:

```text
CREATE UNIQUE INDEX ... ON tenant ((true))
```

Promoted startup/readiness supplies at-least-one and checks `expected_tenant_id`:

```text
missing root        → FAIL CLOSED
multiple roots      → FAIL CLOSED
id mismatch         → FAIL CLOSED
```

Combined serving invariant = exactly one Tenant root.

## 3.2 Area

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
id/code immutable
name mutable

disabled_at IS NULL     → Area accepts new references/assignments
disabled_at IS NOT NULL → Area retired; existing references remain valid
```

Retirement is reversible for the same Area identity. Re-enable changes only future assignability; it never rewrites historical references or restores access state because retirement removed no grants.

No Area hierarchy, owner field, default approver role or generic metadata platform V1.

Retired Area law:

```text
existing Documents/history/grants remain valid
new Document Area assignment       → fail closed at Controlled Information boundary
new AreaScope RoleAssignment       → fail closed at Authorization boundary
new Approval policy Area reference → fail closed at Approval boundary
```

Area code format/canonicalization is fixed at creation by implementation spec; governed numbering consumes the stored immutable code verbatim.

## 3.3 User

```text
User
  id          UUID PRIMARY KEY
  disabled_at TIMESTAMPTZ NULL
```

Laws:

```text
id immutable stable organizational participant identity
no username/email/provider subject/credential/role/capability/tenant_id/home_area/employee key

disabled_at IS NULL     → organizationally eligible
disabled_at IS NOT NULL → organizationally ineligible
```

Disable/re-enable preserves identity. No terminal person-state platform is introduced.

## 3.4 UserProfile

```text
UserProfile
  user_id      UUID PRIMARY KEY REFERENCES User(id)
  display_name TEXT NOT NULL
  email        TEXT NULL
```

One-to-one subordinate Organization state; no second profile UUID.

```text
User        = stable governed participant identity
UserProfile = current human-readable/contact enrichment
```

The Profile row is erasable. Normally an eligible User is profile-complete; profile absence means lawful erasure or a bounded provisioning transition. Consumers render a neutral/opaque fallback instead of fabricating a name.

Email/display name are attributes, never technical identity or provider-binding authority. No `UNIQUE(email)` identity law.

## 3.5 Group

```text
Group
  id   UUID PRIMARY KEY
  name TEXT NOT NULL UNIQUE
```

Flat, company-wide V1. No code, area scope, provider-group mirror, nested group, dynamic rule or retirement lifecycle.

### Group hard-deletion law — M3 closure

Hard deletion remains the smallest lifecycle, but it fails closed while **any live reference** exists.

Within B2, memberships and RoleAssignments must first be absent. Across owners, every persisted live Group reference must be an ordinary typed FK to `Group(id)` with `RESTRICT` / `NO ACTION`.

Known V1 live-reference consumers:

```text
B4 ApprovalPolicy Step actor_rule = Group
B4 Distribution audience configuration targeting Group before release-time snapshot
```

Historical participant/audience snapshots resolve concrete Users and therefore do not require Group to survive after the live configuration no longer references it.

B5 receives no speculative Group requirement; it must add the same reference law only if its design proves a real Group reference.

## 3.6 GroupMembership

```text
GroupMembership
  user_id  UUID NOT NULL REFERENCES User(id)
  group_id UUID NOT NULL REFERENCES Group(id)

PRIMARY KEY (user_id, group_id)
```

Current truth only:

```text
row exists → current member
row absent → not current member
```

No surrogate UUID, interval/tombstone or membership-history family. Audit owns add/remove transition evidence. The pair is two internal UUIDs identifying a pure relationship fact and is a legitimate PK under B1.

---

# 4. Authorization — static product catalogs

## 4.1 Permission and Role are product authority, not deployment data

V1 persists no editable:

```text
permissions table
roles table
role_permissions table
custom-role bundle state
```

Authorization owns versioned-with-product static catalogs:

```text
Permission vocabulary
Role vocabulary
Role → exact Permission bundle
```

Database rows contain only current RoleAssignments using the frozen role vocabulary.

This avoids a second authority and avoids manufacturing unsupported custom-role/custom-permission capability.

## 4.2 Exact current 5×43 Role→Permission matrix — L4 closure

Provenance:

- locked R9 base catalog/bundles originally had 29 permissions;
- single-company refinement removed only `tenant.export` and `tenant.deletion.request`, yielding the current 27 base permissions;
- locked R9.5 adds 16 permissions with explicit per-role additions;
- everything else in the role bundles is unchanged.

The resulting current V1 bundles are pinned below.

### viewer — 3 permissions

```text
document.read_effective
evidence.read
dossier.read
```

### author — 15 permissions

```text
document.read_effective
document.read_history
document.read_working
document.create
document.edit
document.comment
document.submit
document.review_periodic

evidence.read
evidence.create
evidence.edit
evidence.capture

dossier.read
dossier.create
dossier.manage
```

### approver — 4 permissions

```text
document.read_effective
approval.act
evidence.read
dossier.read
```

Approver has no blanket working/history access; exact Approval participation opens the case-specific Submission/evidence required by frozen relationship rules.

### area_manager — 25 permissions

```text
document.read_effective
document.read_history
document.read_working
document.create
document.edit
document.comment
document.submit
document.review_periodic
document.cancel_revision
document.obsolete
document.owner.manage

approval.act
approval.oversee
approval.reassign
approval.cancel

distribution.manage
distribution.oversee

evidence.read
evidence.create
evidence.edit
evidence.capture
evidence.void

dossier.read
dossier.create
dossier.manage
```

Area manager remains an operational manager, not an RBAC/configuration administrator. It has no `access.manage`, `organization.manage`, tenant/config, audit/session or whole-company lifecycle administration.

### tenant_owner — all 43 permissions

```text
tenant.settings.manage
organization.manage
access.manage
document_type.manage
approval_policy.manage
template_use.manage
dictionary.manage

document.read_effective
document.read_history
document.read_working
document.create
document.edit
document.comment
document.submit
document.cancel_revision
document.obsolete
document.review_periodic
document.owner.manage

approval.act
approval.oversee
approval.reassign
approval.cancel

distribution.manage
distribution.oversee

audit.read
audit.export
session.manage

evidence_type.manage
evidence.read
evidence.create
evidence.edit
evidence.capture
evidence.void

dossier_type.manage
dossier.read
dossier.create
dossier.manage

retention.extend
legal_hold.manage
disposition.manage
historical_migration.manage
governed_subject.export
external_repository.publish
```

`tenant_owner` is an ordinary Role bundle, never a bypass. Domain relationship/state/SoD/fresh-auth invariants remain binding.

## 4.3 Catalog / enforcement parity — L3 closure

Static catalogs own meaning. DB CHECKs only enforce representable vocabulary.

Implementation must carry a mechanical parity proof so that:

```text
static Role codes
== RoleAssignment role_code CHECK vocabulary
== role↔scope CHECK vocabulary
```

Drift must fail CI/migration verification. The CHECK is enforcement, never a second semantic catalog.

---

# 5. RoleAssignment — the single persisted Authorization family

```text
RoleAssignment
  id UUID PRIMARY KEY

  user_id  UUID NULL REFERENCES User(id)
  group_id UUID NULL REFERENCES Group(id)

  role_code TEXT NOT NULL

  tenant_scope_id UUID NULL REFERENCES Tenant(id)
  area_scope_id   UUID NULL REFERENCES Area(id)
```

Cross-owner FK actions = `RESTRICT` / `NO ACTION`.

## 5.1 Typed closed unions

Structural subject XOR:

```text
exactly one of user_id / group_id is non-NULL
```

Structural scope XOR:

```text
exactly one of tenant_scope_id / area_scope_id is non-NULL
```

No generic polymorphic `subject_type/id` or `scope_type/id` registry.

## 5.2 Role vocabulary CHECK

```text
role_code IN (
  'tenant_owner',
  'area_manager',
  'author',
  'approver',
  'viewer'
)
```

## 5.3 Role↔scope structural CHECK — M1 closure

Accepted compatibility matrix:

```text
tenant_owner → TenantScope only
area_manager → AreaScope only
author       → TenantScope | AreaScope
approver     → TenantScope | AreaScope
viewer       → TenantScope | AreaScope
```

Invalid pair must be unrepresentable:

```text
CHECK (
     (role_code = 'tenant_owner' AND tenant_scope_id IS NOT NULL)
  OR (role_code = 'area_manager' AND area_scope_id IS NOT NULL)
  OR (role_code IN ('author','approver','viewer'))
)
```

The scope XOR makes the first two branches exclusive to their allowed scope kind. Application write paths validate the same invariant for friendly failure; DB constraint is the authority-independent backstop.

Successor check-site law: every B3–B5/R10-E authorization use declares whether its check target is Tenant-wide or Area-targeted. Permissions frozen as tenant-owner-only whole-company administration/governance remain Tenant-wide checks; an Area relationship cannot silently downgrade a whole-company permission into an Area permission.

## 5.4 Duplicate current-grant backstops

Four partial unique indexes prevent duplicate semantic grants:

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

## 5.5 Current-truth mutation law

```text
INSERT → grant currently exists
DELETE → grant revoked
```

Grant shape is immutable while row exists. Changing subject/role/scope is revoke + new grant. No retained `revoked_at`, effective interval or temporal-grant scheduler V1.

Required grant/revocation evidence is appended to Audit in the same local transaction. Re-grant creates a new RoleAssignment UUID and new Audit evidence.

## 5.6 Why RoleAssignment has UUID while GroupMembership does not — D15 closure

The XOR subject/scope shape has nullable columns, so there is no single NULL-free natural column set that can serve as a PostgreSQL primary key. The four partial unique indexes prove semantic duplicate rejection but cannot together be one table PK.

Therefore:

```text
RoleAssignment.id UUID PK = structurally necessary technical identity
GroupMembership(user_id,group_id) PK = sufficient relationship identity
```

This is not an arbitrary surrogate-key convention.

## 5.7 TenantScope is semantic, not partitioning

`tenant_scope_id → Tenant(id)` means “this grant applies to the whole company represented by the singleton Tenant root”. It is the real semantic relation and is allowed by B1.

It is not a universal partition column and appears only on RoleAssignment where the relation itself is meaningful.

---

# 6. Canonical Authorization evaluation and administration

## 6.1 Live evaluation

No persisted semantic:

```text
user_permissions
effective_permissions
cached group-expanded grants
Session roles/permissions
materialized ACL
provider-role mapping
```

Canonical evaluation:

```text
current direct User RoleAssignments
UNION
current GroupMemberships → current Group RoleAssignments

→ static Role → Permission bundle
→ scope match
→ owner/domain relationship predicate when required
→ owner/domain governance constraints
→ ALLOW or default DENY
```

Role/grant/membership changes therefore take effect on the next canonical check without Session regeneration.

## 6.2 Scope application

```text
Tenant-wide check
  → qualifying TenantScope assignment required

Area-targeted check
  → qualifying TenantScope assignment
     OR matching AreaScope assignment
```

Domain owners supply resource relationship/context; Authorization does not absorb Document/Approval/Evidence lifecycle semantics.

## 6.3 Administration permissions — corrected after exact-bundle recovery

```text
tenant.settings.manage
  → Tenant editable identity/settings

organization.manage
  → Area/User/UserProfile/Group identity & lifecycle

access.manage
  → GroupMembership + RoleAssignment access configuration

session.manage
  → explicit administrative ApplicationSession management
```

The exact frozen bundles grant `organization.manage`, `access.manage`, `tenant.settings.manage` and `session.manage` only to `tenant_owner`. Because `tenant_owner` is TenantScope-only:

```text
Organization administration = TenantScope tenant_owner V1
GroupMembership administration = TenantScope access.manage only
RoleAssignment grant/revoke administration = TenantScope access.manage only
```

There is **no Area-local RBAC administrator V1**. `area_manager` is operational, not RBAC administrative. Area-local delegation remains represented by Tenant owner creating AreaScope RoleAssignments for Users/Groups; it does not delegate access administration itself.

This removes the original candidate's unused `AreaScope access.manage` administration path.

## 6.4 Disabled/retired targets

```text
new direct RoleAssignment to disabled User → fail closed
new GroupMembership for disabled User      → fail closed
new AreaScope RoleAssignment to retired Area → fail closed
```

Existing AreaScope grants remain valid after Area retirement. Retirement freezes new structural use; it is not an implicit deny-all.

---

# 7. Integrated lifecycle operations

## 7.1 User offboarding = destructive access teardown

One local MetalDocs transaction:

```text
BEGIN

lock User
set User.disabled_at = now

revoke all ApplicationSessions for User
  // terminal revoked_at mutation; DO NOT erase Session rows here

delete all GroupMemberships for User
delete all direct User RoleAssignments

append required Audit evidence
insert durable provider-disable intent when provider-side effect is required

COMMIT
```

ProviderSubjectBinding remains because `(issuer,subject) → User` correlation remains truthful after employment/access ends.

Provider effect executes after commit via R10-D. No provider call is part of local atomicity.

Group RoleAssignments remain because they belong to the Group; deleting memberships removes the offboarded User's inherited access.

## 7.2 Re-enable never restores authority silently

```text
BEGIN
lock User
clear User.disabled_at
append required Audit
insert provider-enable durable intent when required
COMMIT
```

After commit:

```text
no prior GroupMembership restored
no prior direct RoleAssignment restored
default deny until explicit fresh grants/memberships are created
```

Identity may return; old authority does not.

No distinct temporary-suspension product state exists V1. A real requirement for leave/suspension with intentional access restoration is a reopen trigger, not a reinterpretation of offboarding.

## 7.3 Area retirement / re-enable

```text
retire Area:
  disabled_at = now
  existing references/grants remain valid
  new references/grants/policy bindings fail closed

re-enable same Area:
  disabled_at = NULL
  future references become legal again
  no access is restored because retirement removed none
```

## 7.4 Group deletion

Hard deletion requires all of:

```text
no GroupMembership rows
no Group RoleAssignments
no live cross-owner typed references
```

If any live B4 Approval/Distribution configuration references the Group, `RESTRICT` / `NO ACTION` must block deletion.

---

# 8. B2 concurrency and deterministic lock law — M2 closure

B1 default isolation remains `READ COMMITTED`. B2 uses narrow row locks + FK/UNIQUE/CHECK enforcement; no global SERIALIZABLE or advisory-lock framework.

## 8.1 Canonical lock acquisition order

Classes may be skipped, but a transaction never revisits an earlier class after acquiring a later one:

```text
1. User row

2. ProviderSubjectBinding rows of that User
   ascending id

3. Area row

4. child-row sets, each in ascending primary-key order:
   ApplicationSession → ascending id
   GroupMembership    → ascending (user_id, group_id)
   RoleAssignment     → ascending id
```

Group deletion is an isolated class:

```text
Group row FOR UPDATE
→ GroupMembership rows ascending user_id
→ Group RoleAssignments ascending id
```

A Group-deletion transaction does not subsequently acquire User / ProviderSubjectBinding / Area locks. FK enforcement supplies the required race with concurrent membership creation; the Group row lock makes deletion/mutation ordering explicit.

## 8.2 Lock modes

```text
eligibility/acceptance readers
  Session issuance
  GroupMembership add
  direct User RoleAssignment insert
  AreaScope RoleAssignment insert
→ FOR SHARE on the governing eligibility/acceptance row

lifecycle mutators
  User offboard/re-enable
  Area retire/re-enable
  Binding disable/re-enable/replace
→ FOR UPDATE
```

`FOR KEY SHARE` is not sufficient for User/Area eligibility serialization because updates to `disabled_at` are non-key updates and may take `FOR NO KEY UPDATE`; the required conflict would not be guaranteed.

## 8.3 Required race outcomes

### C1 Session issuance ↔ User offboarding

```text
issuance first  → offboarding revokes resulting Session
offboarding first → issuance sees disabled User and creates nothing
```

Forbidden: offboarding success + newly issued surviving valid Session.

### C2 Binding acceptance ↔ Session issuance

Session issuance validates currently accepted binding inside binding serialization. Disable/replacement revokes affected Sessions in same local tx; re-enable never revives revoked Sessions.

### C3 GroupMembership add ↔ offboarding

```text
membership first → offboarding deletes it
offboarding first → add sees disabled User → fail
```

### C4 direct User RoleAssignment ↔ offboarding

Same User-row total order; no new direct grant survives a completed offboarding.

### C5 AreaScope grant ↔ Area retirement

```text
grant first → it becomes an existing grant and survives retirement
retirement first → insert sees retired Area → fail
```

No stale-check grant may be born after retirement commit.

### C6 re-enable

Re-enable only changes eligibility/assignability. Deleted access rows or revoked Sessions never resurrect.

### C7 GroupMembership ↔ Group RoleAssignment

No atomic coupling is required. Effective group-mediated authority exists exactly when both current facts exist. Both mutation families are tenant-owner-administered under corrected §6.3, removing cross-tier privilege composition.

## 8.4 In-flight request posture

B2 guarantees fail-closed **future** Session resolution/Authorization after lifecycle commit. It does not claim magical cancellation of work that completed its relevant authn/authz decision before offboarding/retirement committed unless that business transaction participates in a shared lock/invariant.

No frozen requirement demands impossible global request cancellation; stronger linearization requires a concrete future invariant.

---

# 9. Persistence class × mutation law

```text
Tenant
  SEMANTIC AUTHORITY
  id immutable; display_name mutable

Area
  SEMANTIC AUTHORITY
  id/code immutable; name/disabled_at mutable

User
  SEMANTIC AUTHORITY
  id immutable; disabled_at mutable

UserProfile
  SEMANTIC AUTHORITY — subordinate human-readable enrichment
  mutable; erasable

Group
  SEMANTIC AUTHORITY
  id immutable; name mutable; hard-deletable only when unreferenced

GroupMembership
  SEMANTIC AUTHORITY — current relationship
  INSERT / DELETE

RoleAssignment
  SEMANTIC AUTHORITY — current grant
  immutable grant shape while row exists; INSERT / DELETE

ProviderSubjectBinding
  promoted Authentication semantic authority

ApplicationSession
  promoted Authentication semantic authority
```

No historical grant/membership interval family is introduced. Audit is the transition timeline, not current grant authority.

---

# 10. Audit / durable provider-intent law

Administrative mutation that changes B2 identity, eligibility, provider-binding acceptance or effective access appends required Audit evidence in the **same local transaction** as the authoritative mutation.

At minimum:

```text
Tenant display/settings mutation
Area create/rename/retire/re-enable
User create/offboard/re-enable
UserProfile create/update/erasure when identity evidence is required
Group create/rename/delete
GroupMembership add/remove
RoleAssignment grant/revoke
ProviderSubjectBinding acceptance/replacement
administrative Session revocation
offboarding
```

Audit grant/revocation events must preserve enough PII-minimized facts for forensic reconstruction after the current RoleAssignment row is gone: assignment id, subject reference, role code, scope reference/type, actor reference, operation and trusted time according to B6's final Audit field-classification law.

Ordinary login/logout traffic does not become governed semantic Audit merely because it exists; security telemetry is an R10-E/operations concern unless a later accepted requirement says otherwise.

When a B2 semantic mutation requires a provider-side effect:

```text
BEGIN
  local semantic truth
  required Audit
  required durable provider intent
COMMIT

then R10-D executes/retries/reconciles provider effect
```

No Keycloak/provider HTTP call participates in local PostgreSQL atomicity.

---

# 11. Privacy

B2 retains the accepted separation:

```text
erasable when lawful:
  UserProfile
  ApplicationSession rows after lifecycle/evidence need ends
  ProviderSubjectBinding when lawful under B2-1

retained skeleton where governed history requires it:
  User.id
  User.disabled_at
  governed domain records referencing User UUID
  PII-minimized/non-PII Audit skeleton
```

Offboarding is not privacy erasure. Session revocation is not Session deletion. Privacy cleanup occurs separately and in RESTRICT-safe order.

Lawful erasure of ProviderSubjectBinding surrenders its DB-level no-recorrelation guarantee exactly as promoted in B2-1; any later correlation is a new trusted binding decision.

B6 must still classify surviving Audit fields; R10-C must prove restore does not silently resurrect lawfully erased PII. No generic privacy workflow/platform is introduced.

---

# 12. Bootstrap / lockout recovery — L2 closure

Default deny + no bypass means the very first access-administration grant cannot be created by an ordinary request-path actor in a fresh deployment, and an operator can otherwise lock out all tenant access administrators.

Correct boundary:

```text
initial tenant_owner RoleAssignment seeding
admin-lockout recovery
→ distinct non-serving maintenance trust surface
```

This identity is never the ordinary serving role, never request-reachable, never an implicit tenant-content principal, and never a permanent authorization bypass.

R10-F/operations must specify the exact bootstrap/cutover procedure. R10-E may add a friendly last-admin/self-offboard warning or guard, but UX is defense-in-depth, not the recovery authority.

---

# 13. Name/format normalization — L5 closure

Implementation specifications must reject unusable blank/whitespace forms for required names/codes and normalize inputs deliberately, but B2 does not define human-display casing as identity.

In particular:

```text
Group.name case-insensitive uniqueness is NOT implied
Tenant.display_name is not a routing/business key
Area.name is not Area identity; immutable Area.code is
```

Exact trim/length/character constraints belong implementation specs unless a later consumer proves product semantics.

---

# 14. Successor obligations

These are routed consequences of accepted B2, not reopenings.

## B3–B5 authorization check sites

Every canonical permission check site declares its scope target:

```text
Tenant-wide
or
Area-targeted
```

Tenant-owner-only whole-company families remain Tenant-wide even when the underlying resource has an Area relation. Domain owners provide relationship/governance predicates; Authorization does not reinterpret them.

## B4

Must consume:

```text
retired Area rejects new Approval policy Area reference
live ApprovalPolicy Group actor reference → typed FK Group(id) RESTRICT/NO ACTION
live Distribution Group audience reference → typed FK Group(id) RESTRICT/NO ACTION
historical activated participants / release audiences snapshot concrete Users
fresh-auth consumer policy remains bounded per promoted B2-1
```

## B5

No Group requirement is presumed. If a real B5 persistent family references Group, it must obey the same typed-FK RESTRICT law.

## B6

Must finalize:

```text
Audit skeleton field-by-field privacy classification
role-grant/revocation Audit event fields sufficient for forensic reconstruction
User skeleton/privacy consequence
same-commit cross-owner Audit matrix
```

## R10-C

Restore non-resurrection of lawfully erased user PII remains mandatory.

## R10-D

Provider provisioning/disable/enable/reconciliation executes durable intent with retry/idempotency/lease/DLQ and never becomes Authentication/Organization authority.

## R10-E

Must consume:

```text
provider-hosted auth journeys
Session TTL security value
per-check-site scope classification
neutral historical actor fallback when UserProfile absent
optional last-admin/self-offboard UX guard
optional bounded session/device/group display metadata only with a real consumer
```

## R10-F / operations

Must specify:

```text
initial tenant_owner seeding
admin-lockout maintenance recovery
legacy 8-role/capability cutover
legacy global/area dual grant-table removal
legacy tenant_id/RLS/context removal from accepted mapping
static-catalog ↔ DB CHECK parity proof/gate
```

---

# 15. Integrated proof obligations

A later implementation specification/test plan must prove at minimum:

1. Tenant singleton at-most-one DB constraint + at-least-one/matching readiness handshake.
2. Tenant.id immutable.
3. Area.code immutable/deployment-wide unique; retirement/re-enable semantics.
4. User eligibility is one `disabled_at` fact; no provider/AuthZ/PII identity duplication.
5. UserProfile row absence is a valid erasable enrichment state with neutral fallback.
6. Group flat/company-wide; hard delete fails on any live cross-owner reference.
7. GroupMembership pair PK prevents duplicates; no surrogate identity drift.
8. Static role/permission catalogs contain exactly the accepted 5 roles/43 permissions and exact bundles above.
9. Static catalog ↔ role CHECK ↔ role-scope CHECK parity failure is mechanically detected.
10. RoleAssignment subject XOR and scope XOR hold at DB level.
11. Illegal role↔scope pairs fail at DB level, including `tenant_owner@AreaScope` and `area_manager@TenantScope`.
12. Four partial uniqueness backstops reject duplicate current grants.
13. RoleAssignment UUID is stable technical PK; current grant shape is immutable while present.
14. Authorization is live-state/additive/default-deny with no effective-permission semantic store or Session AuthZ snapshot.
15. Ordinary V1 access administration is TenantScope tenant-owner-only because only `tenant_owner` carries `access.manage`.
16. Disabled User cannot receive new direct grant, membership or Session.
17. Retired Area cannot receive new AreaScope grant/reference while existing refs/grants remain usable.
18. Offboarding revokes all Sessions, deletes memberships/direct grants and appends Audit in one local transaction.
19. Offboarding retains binding correlation and provider work is durable-post-commit choreography.
20. Re-enable restores identity eligibility only and never old access rows/Sessions.
21. Canonical lock acquisition order and `FOR SHARE`/`FOR UPDATE` modes eliminate the reviewed deadlock/race classes under READ COMMITTED.
22. Group deletion vs concurrent membership/grant/reference creation fails closed.
23. Same-commit Audit exists for material B2 identity/eligibility/access mutations.
24. Grant/revocation Audit evidence remains forensic after current RoleAssignment deletion without becoming current AuthZ authority.
25. Provider calls never participate in local DB atomicity.
26. Privacy cleanup erases enrichment/auth state without breaking governed User references.
27. No universal tenant/company/deployment partition column re-enters through B2.
28. No provider role/group/org/claim bridge re-enters Authorization.
29. No custom-role/permission/ACL/ReBAC/deny/nested-group platform appears without reopen evidence.
30. Maintenance bootstrap/recovery is non-serving/request-unreachable and does not become a bypass.

---

# 16. Reopen triggers

B2 reopens only on material evidence such as:

```text
simultaneously active multiple MetalDocs-facing provider identities per User becoming required
provider subject handover/reuse between different Users becoming a legitimate product requirement
immediate provider-initiated Session revocation becoming mandatory
real HR/workforce Area placement independent of authorization
nested/dynamic/scoped Groups becoming a real requirement
custom Roles/Permission bundles becoming an accepted product capability
temporary/scheduled grants becoming required by a named consumer
explicit deny semantics becoming necessary
arbitrary resource sharing/relationship graph requiring ReBAC/OpenFGA-class machinery
temporary User suspension requiring intentional automatic authority restoration
Area retirement requiring different semantics than preserve-existing/block-new
a new permission that changes the frozen role bundles
```

Current implementation inconvenience or legacy table shape is never a reopen trigger.

---

# 17. Bounded delta review contract

The next review is **one bounded delta review of this integrated corrected target**, not a new broad B2 review and not B2-2/B2-3/B2-4 microreviews.

It must verify only that the adjudicated corrections close the prior findings without introducing contradiction:

```text
M1  role↔scope DB CHECK + matrix unchanged
A1  exact bundle matrix implies tenant-owner-only access administration V1
M2  canonical lock order/modes consistent with every B2 transaction/race
M3  Group live-reference RESTRICT law with B4 routing
L1  Session revoke vs erase separation
L2  maintenance bootstrap/lockout boundary
L3  catalog↔CHECK parity proof obligation
L4  exact 5×43 bundles match frozen R9/R9.5 + single-company refinement
L5  normalization remains implementation detail
D15 RoleAssignment UUID rationale
```

The delta reviewer must also mechanically verify that all prior `BLOCKER=0`, three MAJOR findings and five LOW findings are closed, and explicitly attack the new A1 administration simplification for any lost legitimate V1 delegation.

If the bounded delta returns `BLOCKER=0 / MAJOR=0 / no new material contradiction`, the next gate is operator promotion of **the entire R10-B2 batch** into R10 authority, followed by R10-B3. No product implementation is authorized by this file.
