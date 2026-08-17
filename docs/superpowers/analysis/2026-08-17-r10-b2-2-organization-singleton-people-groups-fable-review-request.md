# MetalDocs R10-B2-2 — Organization Singleton Root / People / Groups — Independent Review Request

> **Status:** CANDIDATE / REVIEW REQUEST — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Authority baseline:** `71791dfecd4cd185684373ffcdccbf256138b741`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Authority note:** this packet proposes a B2-2 target for independent challenge. It does not amend R10 authority, handoff, program authority, ledger, code, schema, OpenAPI, frontend or Keycloak configuration.

---

# 0. Authority / scope

Read `AGENTS.md` and follow the repository authority chain before reviewing this packet.

Promoted authority already fixes:

```text
one company per deployment V1
singleton Tenant semantic root
Tenant.id immutable
expected_tenant_id ↔ DB Tenant.id startup/readiness fail-closed handshake
no universal tenant_id/company_id/deployment_id partition column
no Tenant RLS / tenant context / customer routing
Authentication != Organization != Authorization
Keycloak = V1 Authentication provider
R10-B2-1 = CLOSED / APPROVED
implementation = BLOCKED
```

B2-1 already owns and must not be rediscovered here:

```text
ProviderSubjectBinding
ApplicationSession
provider subject correlation authority
fresh-auth / assurance
provider-disable/live-session posture
provider reconciliation semantics
```

B2-3 later owns:

```text
Permission
Role
RoleAssignment
TenantScope | AreaScope
grant/revocation evidence
canonical grant evaluation
```

B2-4 later owns final cross-family transaction/locking/Audit/durable-intent matrix.

This packet asks only whether the minimum Organization persistent state is correct.

---

# 1. Root cause

Current auth/IAM shape historically mixed multiple authorities:

```text
human organizational identity
credential/provider identity
session state
roles/capabilities
Tenant plumbing
```

The promoted target separates them:

```text
Keycloak/provider → credential/authentication mechanism
Authentication     → provider binding + local Session + assurance
Organization       → company / people / groups identity and relationships
Authorization      → permissions / roles / grants
```

B2-2 must therefore avoid rebuilding a generic IAM User whose row again owns credentials, provider identity, authorization or customer-tenancy partitioning.

---

# 2. Target invariant

> Organization owns durable company/person/group identity and current organizational relationships. Authentication decides whether an external provider subject can authenticate as a User. Authorization decides what that User/Group may do. Neither property is embedded in Organization state.

Therefore:

```text
User != credential
User != provider account
User != RoleAssignment
User != Session
Group != Authorization scope
Area != hierarchy engine
Tenant != DB partition
```

The smallest candidate persistent Organization state is:

```text
Tenant
Area
User
UserProfile
Group
GroupMembership
```

`UserProfile` is subordinate Organization state, not a new bounded context.

---

# 3. B2-2-D1 — Tenant remains one singleton semantic root

Candidate:

```text
Tenant
  id           UUID PRIMARY KEY
  display_name TEXT NOT NULL
```

No V1 entitlement for:

```text
slug
code
status
ACTIVE/SUSPENDED/ERASED
generic settings JSONB
locale/timezone registry
tenant_id/company_id/deployment_id
```

`Tenant.id` is already promoted as immutable deployment↔DB identity anchor.

`display_name` is a real operator-mutable Organization fact and a concrete consumer of `tenant.settings.manage`; the permission does not imply a generic settings platform.

---

# 4. B2-2-D2 — exactly-one Tenant enforcement is split structurally

Target property:

```text
DB state contains at most one Tenant row
startup/readiness requires at least one Tenant row
```

Candidate enforcement:

```text
at-most-one:
  PostgreSQL uniqueness over one constant expression for every Tenant row
  (conceptually: UNIQUE INDEX tenant_singleton ON tenant ((true)))

at-least-one:
  promoted startup/readiness handshake fails closed when Tenant root is missing
```

Therefore:

```text
at-most-one + at-least-one = exactly one Tenant root in every serving deployment
```

No fake business `singleton_key` field and no trigger-heavy lifecycle are proposed.

Reviewer must attack whether this is a valid, minimal PostgreSQL representation and whether a materially simpler/falsifiable enforcement exists.

---

# 5. B2-2-D3 — no generic Tenant settings platform

Candidate law:

> Persist only typed Organization facts with a real consumer. `tenant.settings.manage` does not create entitlement for a schema-less key/value or JSON settings authority.

Today B2-2 needs only the singleton root identity and mutable `display_name` as Organization-owned Tenant state.

Future typed settings may be added by their actual semantic owner without changing the Tenant identity law.

---

# 6. B2-2-D4 — Area = stable UUID identity + immutable unique code + mutable name

Candidate:

```text
Area
  id   UUID PRIMARY KEY
  code TEXT NOT NULL
  name TEXT NOT NULL

UNIQUE(code)
```

Mutation law:

```text
id   = immutable
code = immutable V1
name = mutable
```

Reason for `code` as stable business fact:

- frozen numbering may use `{AREA}`;
- Authorization scopes reference Area identity;
- Approval `RoleInArea` consumes Area identity;
- Controlled Information persists Document Area.

The code is not the technical PK, but changing it silently can alter future governed numbering/organizational semantics. If a real re-code requirement appears, it should be an explicit material operation/reopen rather than casual mutation.

Uniqueness is deployment-wide under the single-company substrate.

---

# 7. B2-2-D5 — no Area hierarchy or lifecycle V1

Candidate deliberately omits:

```text
parent_area_id
Area tree / org chart
ACTIVE/INACTIVE/ARCHIVED
disabled_at
```

No promoted consumer currently requires hierarchy or a distinct Area lifecycle.

V1 supports:

```text
create Area
rename Area
reference Area
```

Deletion/retirement must respect later FK/use-case rules; a real requirement to preserve an Area while prohibiting future assignment is the trigger for a minimal lifecycle.

---

# 8. B2-2-D6 — User is a minimal stable organizational identity root

Candidate:

```text
User
  id          UUID PRIMARY KEY
  disabled_at TIMESTAMPTZ NULL
```

No User fields for:

```text
username
email
password/provider credentials
issuer/subject
provider account status
roles/permissions/capabilities
tenant_id/company_id
home_area/department
```

Meaning:

```text
User.id = stable MetalDocs organizational participant identity
```

B2-1 references this User through `ProviderSubjectBinding.user_id`.

---

# 9. B2-2-D7 — User eligibility is reversible `disabled_at`, not a state machine

Candidate:

```text
disabled_at IS NULL     → User organizationally eligible
disabled_at IS NOT NULL → User organizationally ineligible
```

Same User may be disabled and later re-enabled; Audit owns transition history.

No V1 states:

```text
ACTIVE
INACTIVE
SUSPENDED
TERMINATED
OFFBOARDED
ERASED
```

because no accepted consumer distinguishes those meanings today.

B2-1 is consumed as follows:

```text
Session issuance requires:
  accepted ProviderSubjectBinding
  + current User eligibility

User offboarding:
  makes User ineligible
  + B2-4 must ensure local Session revocation is immediate/race-safe
```

B2-2 fixes the eligibility fact; B2-4 fixes the final transaction/locking realization.

---

# 10. B2-2-D8 — separate User root from erasable UserProfile

Candidate:

```text
UserProfile
  user_id      UUID PRIMARY KEY REFERENCES User(id)
  display_name TEXT NOT NULL
  email        TEXT NULL
```

`UserProfile` is strict 1:1 subordinate Organization state; no separate UUID is proposed.

Purpose:

```text
User        → stable participant identity referenced by governed history
UserProfile → current human-readable/contact enrichment that can be erased/scrubbed
```

This separation exists to satisfy the already-promoted privacy requirement:

```text
retained PII-minimized/non-PII historical skeleton
+
separately erasable human-readable/user enrichment
```

It does not itself decide final B6 Audit field classification or create a privacy workflow.

Reviewer must attack whether this split is essential or accidental complexity, and whether governed references can instead safely target a PII-bearing User row without creating an erasure dead-end.

---

# 11. B2-2-D9 — profile attributes are never technical identity

`UserProfile.email` and `display_name` are current human attributes only.

They are not:

```text
technical PK
provider correlation key
Authentication binding authority
Authorization subject key
```

Candidate therefore does **not** freeze `UNIQUE(email)` as an identity invariant.

Any operational duplicate-email validation must not acquire identity semantics or bypass B2-1's explicit provider-subject correlation laws.

`username` is absent from Organization V1; provider username/display attributes may appear in provider/UI projections when needed, never as User identity.

---

# 12. B2-2-D10 — no User→Area membership/home-area V1

Candidate omits:

```text
User.area_id
home_area_id
department_id
UserAreaMembership
```

Existing promoted consumers do not require an independent organizational-placement fact:

```text
Authorization Area permissions → RoleAssignment scope = AreaScope (B2-3)
Approval RoleInArea             → consumes Authorization + Area
Document Area                   → Controlled Information owns it
```

Adding `User.home_area` now would duplicate/blur those meanings without an accepted second consumer.

Reopen trigger: a real organizational requirement for workforce placement independent of access grants (for example reporting/HR routing) that cannot be represented by existing facts.

---

# 13. B2-2-D11 — Group is a flat company-wide organizational identity

Candidate:

```text
Group
  id   UUID PRIMARY KEY
  name TEXT NOT NULL

UNIQUE(name)
```

No V1:

```text
code
parent_group_id
group hierarchy
group type
dynamic rule
area_id/scope
provider group mirror
ACTIVE/INACTIVE lifecycle
```

Groups are flat per frozen authority.

Group itself owns no Authorization scope. B2-3 may assign the same Group different RoleAssignments at `TenantScope` or `AreaScope`; Group identity remains company-wide.

`name` uniqueness is deployment-wide because one company exists in the deployment.

Reviewer must attack whether mutable `name` + total uniqueness is sufficient or whether Group requires a stable immutable code/business key for a real consumer.

---

# 14. B2-2-D12 — GroupMembership is current User↔Group truth, not history engine

Candidate relational fact:

```text
GroupMembership
  user_id  UUID NOT NULL REFERENCES User(id)
  group_id UUID NOT NULL REFERENCES Group(id)

UNIQUE(user_id, group_id)
```

Semantics:

```text
row exists → current member
row absent → not current member
```

No V1:

```text
ACTIVE/REMOVED state
joined_at/left_at semantic history
membership tombstones
nested group memberships
```

Membership transition history belongs to Audit. Approval already snapshots resolved participants when the relevant Step activates; historical membership is therefore not required as a second durable authority for Approval.

Open design point for independent review:

> Does B1's UUID-identity substrate require `GroupMembership.id UUID PRIMARY KEY`, or is this relationship correctly identified by the pair `(user_id,group_id)` because no consumer references Membership as an independent entity?

The candidate prefers **no surrogate UUID** unless the reviewer proves a concrete consumer/invariant requiring one.

B2-4 later owns exact membership mutation/Audit transaction boundaries.

---

# 15. Offboarding vs privacy cleanup

These remain distinct:

## Offboarding

```text
User.disabled_at = now
+ B2-4 local Session revocation/coherence
```

Offboarding does not automatically imply privacy erasure, GroupMembership deletion or RoleAssignment deletion; the final cross-owner disposition policy belongs to B2-4 after B2-3 defines grants.

## Lawful privacy cleanup

May erase/scrub current profile/enrichment and eligible Authentication state while retained governed evidence remains valid by reference to stable MetalDocs User/domain facts.

B2-2 does not introduce:

```text
UserErasureRequest
PrivacyCase
PrivacyWorkflow
generic privacy engine
```

B6/R10-C own the later Audit/restore privacy proofs.

---

# 16. Deployment-wide uniqueness re-derivation

Former per-Tenant uniqueness is re-derived as:

```text
Tenant                          → singleton
Area.code                       → deployment-wide UNIQUE
Group.name                      → deployment-wide UNIQUE
GroupMembership(user_id,group_id) → one current relationship
UserProfile.user_id             → one profile/User
User.id                         → UUID PK inside product DB
```

No replacement tenant/company partition column is needed.

---

# 17. Candidate persistence / mutation classification

B2-4 owns final classification, but this packet proposes the semantic direction:

```text
Tenant
  SEMANTIC AUTHORITY
  id immutable
  display_name mutable

Area
  SEMANTIC AUTHORITY
  id/code immutable
  name mutable

User
  SEMANTIC AUTHORITY
  id immutable
  disabled_at mutable/reversible eligibility

UserProfile
  SEMANTIC AUTHORITY subordinate enrichment
  mutable
  erasable/scrubbable

Group
  SEMANTIC AUTHORITY
  id immutable
  name mutable

GroupMembership
  SEMANTIC AUTHORITY current relationship
  create/delete
  historical transition = Audit
```

Reviewer should challenge whether `UserProfile` should instead be classified as attributed support rather than semantic authority; the critical property is that current human-readable Organization profile data has one owner and is erasable. Do not let classification vocabulary create a second owner.

---

# 18. Explicitly rejected/deferred B2-2 complexity

```text
Tenant slug/routing key
Tenant lifecycle/deletion/tombstones
generic Tenant settings registry
Area hierarchy
Area lifecycle without consumer
User credentials/provider state
User username as product identity
User home Area / org chart
employee/HR directory platform
Group hierarchy/nesting
provider group synchronization
dynamic groups
Group Area scope
Group lifecycle without consumer
membership-history state machine
privacy workflow/platform
tenant/company partition columns
RLS compensation
```

Future evidence may add a bounded fact without reopening unrelated Organization identity laws.

---

# 19. Required adversarial proof obligations

The independent reviewer must disposition B2-2-D1..D12 and attack at least:

```text
P1  constant-expression singleton uniqueness is valid/minimal PostgreSQL enforcement
P2  DB at-most-one + readiness at-least-one really proves exactly-one serving root
P3  Tenant root can remain minimal; tenant.settings.manage does not require generic settings store
P4  Area.code genuinely needs immutable deployment-wide uniqueness
P5  Area hierarchy/lifecycle can be omitted without breaking a frozen consumer
P6  User contains no credential/provider/AuthZ fact by necessity
P7  reversible User.disabled_at is sufficient V1 eligibility lifecycle
P8  User disable/re-enable cannot create identity/history ambiguity
P9  User/UserProfile split is necessary and not speculative privacy abstraction
P10 governed historical references can survive UserProfile erasure/scrub
P11 UserProfile fields are sufficient minimum; username/phone/title/etc. are not required
P12 email does not need identity-level uniqueness
P13 no accepted flow requires User.home_area / UserAreaMembership
P14 Group can be company-wide/flat without Area/scope/provider fields
P15 Group.name uniqueness is sufficient; stable Group code is not required
P16 Group lifecycle can be omitted
P17 GroupMembership requires or does not require a surrogate UUID
P18 current-only GroupMembership + Audit is sufficient historical model
P19 disabled User + retained memberships does not create a hidden access path once B2-3/B2-4 close
P20 offboarding may correctly defer membership/grant cleanup policy to B2-4
P21 no tenant_id/company_id is required on any B2-2 family
P22 lawful profile/auth cleanup can preserve governed history without a generic privacy platform
P23 UserProfile classification does not create duplicate authority
P24 no legacy IAM field has an evidenced V1 consumer accidentally removed
P25 all cross-owner FKs remain RESTRICT/NO ACTION and no cascade erases historical authorities
P26 subtractive pass: what else can be deleted from B2-2 without weakening a real property?
```

---

# 20. Strong alternatives to attack

Reviewer must compare against credible alternatives, not straw men.

## Alternative A — monolithic User row

```text
User(id, display_name, email, disabled_at)
```

Attack whether this is actually simpler and whether privacy can still erase PII without deleting or mutating historical organizational identity relied on by governed records.

## Alternative B — generic organization directory

```text
OrganizationUnit / Person / Position / Department / Manager / Membership / lifecycle
```

Attack whether any real V1 consumer justifies this enterprise-directory complexity.

## Alternative C — Keycloak as User/Group authority

Attack whether provider Users/Groups could replace MetalDocs `User`/`Group` while preserving product Organization identity, provider independence and canonical Authorization boundaries.

## Alternative D — no UserProfile split, Audit snapshots all human names

Attack whether duplicating PII snapshots into governed/Audit facts is actually safer or instead makes privacy erasure materially harder.

---

# 21. Reopen triggers

Examples of material evidence that may reopen a B2-2 decision:

```text
real employee/home-area fact independent of Authorization
organizational hierarchy needed by an accepted workflow/reporting consumer
Area retirement semantics needed to preserve history while blocking future use
Group lifecycle/hierarchy/dynamic membership required by a real consumer
stable Group code required by an external/business contract
multiple current profiles/addresses/contact identities with independent lifecycle
legal/privacy evidence showing the proposed User/UserProfile separation is insufficient
consumer requiring GroupMembership as independently addressable identity
```

Preference, legacy schema convenience, ERP-style expectations or hypothetical enterprise IAM features are not reopen evidence.

---

# 22. Requested independent review output

Required verdict:

```text
APPROVE R10-B2-2 ORGANIZATION SINGLETON / PEOPLE / GROUPS
```

or

```text
APPROVE R10-B2-2 ORGANIZATION SINGLETON / PEOPLE / GROUPS WITH MATERIAL FIXES
```

or

```text
DO NOT APPROVE R10-B2-2 ORGANIZATION SINGLETON / PEOPLE / GROUPS
```

Required output:

- verdict;
- BLOCKER / MAJOR / LOW;
- disposition B2-2-D1..D12;
- Tenant singleton/enforcement verdict;
- Area identity/lifecycle verdict;
- User lifecycle verdict;
- User/UserProfile privacy split verdict;
- email/username identity verdict;
- User→Area omission verdict;
- Group shape/lifecycle verdict;
- GroupMembership surrogate-ID/history verdict;
- offboarding/privacy boundary verdict;
- subtractive/YAGNI findings;
- any material reopen outside B2-2;
- exact corrected target if fixes are required;
- whether a bounded delta review is sufficient after correction.

Write review evidence to:

`docs/superpowers/analysis/2026-08-17-r10-b2-2-organization-singleton-people-groups-independent-fable-review.md`

Review artifacts are evidence only and must not alter target authority.
