# MetalDocs — Single-Company Deployment / Tenancy Rebaseline — Adjudicated Corrected Target

> **Status:** ADJUDICATED CORRECTED CANDIDATE — **PENDING BOUNDED DELTA CHECK — NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Authority baseline before rebaseline:** `b2926f5a2d885ea8cc8a48f1261a1d8750498020`
> **Candidate packet:** `docs/superpowers/analysis/2026-08-17-single-company-deployment-tenancy-rebaseline-fable-review-request.md` @ `cba89d9d`
> **Independent review:** `docs/superpowers/analysis/2026-08-17-single-company-deployment-tenancy-rebaseline-independent-fable-review.md` @ `1acd5128`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Authority note:** this artifact records operator adjudication and corrected target only. It does not amend R9.5, R10-A, R10-B1, current-agent-handoff, code, schema, OpenAPI, frontend or deployment.

---

# 1. Independent review result

```text
VERDICT = APPROVE SINGLE-COMPANY DEPLOYMENT/TENANCY REBASELINE
          WITH MATERIAL FIXES

BLOCKER = 0
MAJOR   = 3
LOW     = 6
```

The review confirmed the root cause: the promoted tenant-qualified PK/FK, RLS, tenant-context, background tenant-routing, Tenant customer lifecycle, Tenant Portability Export and tenant-namespaced key invariant all derive from the pooled/shared-customer deployment premise. The clarified V1 requirement removes that premise.

The operator adjudicates every finding below under the Method. Findings are evidence, not authority.

---

# 2. Operator adjudication

| Finding | Decision | Corrected target |
|---|---|---|
| M1 — deployment↔database identity handshake | **ACCEPT / NARROW** | Keep one durable singleton `Tenant` root UUID in the DB and one configured `expected_tenant_id`; startup/readiness fails closed on mismatch. Do not create a Deployment aggregate or mix this company-root identity with deployment security-profile identity. |
| M2 — privacy obligations hidden inside Tenant erasure | **ACCEPT / RE-ANCHOR** | Defer customer/company deletion lifecycle, but retain user/data-subject privacy obligations: user offboarding, PII-minimized immutable Audit skeleton, separately erasable human-readable enrichment, restore non-resurrection proof, and GCR-R4 crypto-erasure reopen trigger if real immutable Target Data appears. Do not invent a generic privacy workflow/aggregate without a real requirement. |
| M3 — orphaned permissions | **ACCEPT** | Remove `tenant.export` and `tenant.deletion.request` from the frozen V1 base permission catalog. Base catalog becomes 27; R9.5 delta remains 16; five roles unchanged. Reinstate only if Tenant Portability or customer-lifecycle features re-enter. |
| L1 — singleton root enforcement | **ACCEPT WITH CORRECTION** | Enforce exactly one `Tenant` root row structurally. Do **not** retain an `ACTIVE` state or one-state lifecycle. |
| L2 — Tenant vocabulary | **ACCEPT** | Retain `Tenant`, `TenantScope`, `tenant_owner`, `tenant.settings.manage` consistently. No partial rename. `Tenant` means the single company/organization root of one deployment, not a DB partition dimension. |
| L3 — DB least privilege | **ACCEPT** | Keep non-owner/NOSUPERUSER serving role and separate non-serving maintenance trust surface. Remove RLS-specific `NOBYPASSRLS`/per-Tenant-iteration authority wording when RLS is removed. |
| L4 — storage layout | **ACCEPT** | Opaque immutable keys remain invariant; tenant prefix is no longer an invariant and is not required to be removed. R10-C may preserve/change layout by the simplest safe migration. |
| L5 — Tenant lifecycle | **ACCEPT** | Defer the whole `ACTIVE/SUSPENDED/ERASED` customer lifecycle. No vestigial `SUSPENDED` maintenance state. Deployment stop/maintenance is operations. |
| L6 — uniqueness sweep | **ACCEPT** | B2–B5 must explicitly rederive every former “unique within tenant” law as deployment-wide or owner-specific uniqueness. No blind find/replace. |

No finding creates a new bounded context, provider authority or framework.

---

# 3. Corrected SC-R1 — deployment model

```text
OUTCOME = RESTRUCTURE NOW
```

V1 deployment invariant:

> **One MetalDocs deployment serves exactly one company/organization.**

Current V1 deployment = Metal Nobre.

Productization invariant:

```text
one codebase
one product architecture
same build artifacts
same migrations
configuration/data vary per deployment
customer-specific forks = forbidden
```

A future second customer defaults to a second deployment stamp using the same product artifacts. A second customer does **not** automatically reopen pooled/shared tenancy; it triggers a deployment-economics/operations review.

Shared backend / DB-per-customer / pooled / hybrid remain future topology choices to be made only from measured evidence.

---

# 4. Corrected SC-R2 — Tenant retained as singleton semantic root

`Tenant` remains a durable Organization-owned semantic root because real consumers survive:

```text
company display identity
operator-managed company settings
company-wide Authorization scope
M1 deployment↔DB identity anchor
future productization/backfill anchor
```

Binding definition:

> **`Tenant` is the single company/organization root of one MetalDocs deployment. Exactly one Tenant root row exists per deployment. It is semantic company/root state and an Authorization scope target, never a row-partition dimension in V1.**

V1 Tenant has no customer lifecycle state machine.

Do not infer these merely from the noun `Tenant`:

```text
tenant_id on every table
RLS
request tenant context
customer routing
customer switcher
pooled deployment
```

Do not rename to `Company`/`Organization` now. If future commercialization proves a vocabulary defect, rename the full vocabulary deliberately in one future design stage; partial rename is forbidden.

---

# 5. Corrected M1 — deployment↔database identity handshake

Removing universal `tenant_id` must preserve the real property:

> A deployment must never silently serve a database belonging to another company/root.

Minimum structural mechanism:

```text
Database:
  exactly one Tenant root row
  Tenant.id = stable UUID

Deployment configuration:
  expected_tenant_id = stable UUID

Startup/readiness:
  load singleton Tenant root
  compare root.id to expected_tenant_id
  mismatch / missing / multiple roots = FAIL CLOSED
```

The same UUID is a future backfill anchor if pooled tenancy is ever deliberately selected.

This is **not** a new Deployment aggregate, company-directory service, control plane or per-row company column.

Deployment security profile (`production` vs explicit `dev/test`) remains a separate platform/R10-C property; do not collapse it into `Tenant` business state.

---

# 6. Corrected SC-R3 — remove tenant-qualified PK/FK substrate

Material B1 structural reopen:

Former default:

```text
tenant_id UUID NOT NULL
id UUID NOT NULL
PRIMARY KEY (tenant_id, id)
```

New V1 default:

```text
id UUID PRIMARY KEY
```

References use ordinary typed UUID FKs:

```text
target_id UUID NOT NULL
FOREIGN KEY (target_id) REFERENCES target_table(id)
```

Surviving laws:

```text
technical identity = UUID
business/provider/external identifiers never become PK by convenience
cross-owner FK proves existence/identity only, never authority
cross-owner DELETE/UPDATE = RESTRICT / NO ACTION only
cross-owner CASCADE / SET NULL / SET DEFAULT = forbidden
within-owner cascade remains exceptional/subordinate-only
universal polymorphic business registries remain forbidden
```

New anti-inertia law:

> Do not add `tenant_id`, `company_id`, `organization_id` or `deployment_id` to ordinary rows merely to preserve hypothetical shared multitenancy. A root reference exists only where the business relationship itself genuinely targets the company/root.

---

# 7. Corrected SC-R4 — remove Tenant RLS/context

Remove V1 tenant-isolation machinery:

```text
ENABLE/FORCE Tenant RLS
tenant GUC/request context
fail-closed tenant seeding
same-Tenant repository predicates as isolation law
per-worker SeedTxTenant class of behavior
```

Reason: the only company dataset in the product DB is already isolated by the deployment boundary. RLS had been promoted as Tenant isolation only, not canonical Authorization.

Do **not** compensate by adding Area/Role/Permission RLS. Authorization stays with Authorization/domain relationship predicates.

DB security that remains independently justified:

```text
ordinary serving role = non-owner + NOSUPERUSER
DDL/object ownership separate
maintenance/migration principal = separate non-serving trust surface
```

`NOBYPASSRLS` is not a promoted correctness property when no RLS exists.

---

# 8. Corrected SC-R5 — Keycloak posture

Keycloak remains the V1 Authentication provider.

V1 deployment posture:

```text
one Keycloak realm/trust domain for this deployment
one MetalDocs client/application
users of this company
```

Keycloak Organizations / multi-company routing is not required V1.

Future realm-level federation (OIDC/SAML/LDAP/AD) remains supported by provider configuration and does not require MetalDocs multitenancy.

`issuer` remains explicit provider identity in the binding; never hardcode a specific identity source into domain identity.

---

# 9. Corrected SC-R6 — Auth binding and session without tenant dimension

B2-1 target direction after rebaseline:

```text
AuthenticationSubjectBinding
  id UUID PK
  organization_user_id UUID FK
  issuer
  subject
  ...

UNIQUE (issuer, subject)
```

One User ↔ one-or-many provider subjects remains a B2 decision; do not force one-per-user merely because V1 is single-company.

`ApplicationSession` is deployment/company-bound by the deployment itself and therefore does not carry customer/Tenant routing state.

Still required:

```text
opaque app session
finite expiry/revocation
stable issuer+subject binding
no email/username auto-binding
no roles/permissions/groups/Area grants snapshot
structural anti-corruption contract
fresh-auth/assurance facts
provider reconciliation
```

No tenant selector, tenant-first routing or company switcher exists V1.

---

# 10. Corrected SC-R7 — company-wide Authorization scope survives

The distinction remains semantically real:

```text
TenantScope = whole company/root
AreaScope   = one Area
```

RoleAssignment remains:

```text
subject = User | Group
role = one of five frozen roles
scope = TenantScope | AreaScope
```

`TenantScope` is not a DB partition key. Its concrete relational representation is a B2 decision and need not carry a redundant `tenant_id` payload.

Five roles remain:

```text
tenant_owner
area_manager
author
approver
viewer
```

---

# 11. Corrected SC-R8 / M2 — defer customer lifecycle; retain data-subject privacy

Defer V1 customer/company lifecycle features:

```text
Tenant ACTIVE/SUSPENDED/ERASED
TenantDeletionRequest
TenantErasureRecord
Tenant-level erasure tombstones
Tenant-level restore reconciliation
customer/company delete workflow
```

Company deployment decommission is an operations concern.

This does **not** weaken user/data-subject privacy requirements.

Re-anchor to Organization User lifecycle/offboarding + Audit/restore proof:

```text
User deactivation/offboarding revokes app access/sessions
human-readable personal enrichment must be erasable where lawful
immutable Audit allowed to survive must be PII-minimized/non-PII
B6 classifies surviving Audit skeleton field-by-field
restore must not silently resurrect personal data that was lawfully erased
```

GCR-R4 trigger survives:

> If B6/R10-C proves a real immutable Target Data family must remain stored yet become unintelligible after lawful data-subject erasure, reopen the minimum tenant/data-key design with named Target Data and fail-closed enforcement.

Do not create `PrivacyCase`, generic privacy workflow, new privacy bounded context or user-erasure state machine absent a concrete requirement.

RetentionBinding, LegalHold, disposition, governed Artifact deletion, ordinary user offboarding, backup and restore remain V1 concerns.

---

# 12. Corrected M3 — permission catalog delta

Remove two V1 base permissions whose target capabilities are deferred:

```text
tenant.export
tenant.deletion.request
```

Result:

```text
R9 base catalog = 27 permissions
R9.5 bounded delta = 16 permissions
V1 total = 43 semantic permissions
roles = unchanged (5)
```

`tenant.settings.manage` remains because company/root settings remain a live V1 capability.

`tenant_owner` remains the whole-company role bundle, never an Authorization bypass.

If Tenant Portability Export or customer deletion lifecycle later re-enters, the corresponding permissions re-enter only through that feature's formal reopen.

---

# 13. Corrected SC-R9 — defer Tenant Portability Export

Defer:

```text
Tenant Portability Export
```

Keep independently required contracts:

```text
Backup / Restore
Governed Subject Export
External Repository PUBLISH_COPY
Historical Migration
IMPORT_COPY
```

Authorization-safe export completeness remains binding for contracts that claim completeness.

Moving a whole deployment to another host/environment can use deployment backup/restore of the same schema until a real portability/product-exit consumer requires a distinct package contract.

---

# 14. Corrected SC-R10 — background work without customer routing

Remove pooled-customer discovery/routing:

```text
Tenant enumeration
per-Tenant due-work loops
tenant_id in globally claimable async routing metadata
tenant re-entry context
```

Keep async correctness:

```text
transactional outbox / durable intent
same-commit intent insertion when required
idempotency
lease/claim state
retry
DLQ
truthful external-effect outcome/reconciliation
```

V1 execution shape:

```text
claim due work from this deployment's product DB
→ execute under ordinary application/system authorization rules
→ commit/reconcile
```

R10-D still owns execution mechanics.

---

# 15. Corrected SC-R11 — storage key namespace

Retain:

```text
provider key opaque to business semantics
immutable object key for an existing confirmed Artifact
no overwrite
canonical SHA-256 integrity
```

Remove as V1 invariant:

```text
tenant-namespaced key prefix
```

Do not mandate prefix removal either. R10-C may preserve a harmless existing prefix during migration or select a cleaner layout based on the simplest safe implementation.

Deployment-scoped bucket/account/container isolation is operations/provider configuration, not business identity.

---

# 16. Uniqueness rederivation obligation

B2–B5 must explicitly review every prior `tenant`-scoped uniqueness rule.

Examples of likely deployment-wide forms:

```text
AuthenticationSubjectBinding: UNIQUE(issuer, subject)
DocumentType.code: unique deployment-wide
EvidenceType.code: unique deployment-wide
Dossier key: unique within DossierType (unless later semantics prove otherwise)
Area code/name: deployment-wide as B2 decides
Group code/name: deployment-wide as B2 decides
numbering scopes: TYPE / TYPE_AREA without Tenant partition prefix
```

These are successor decisions; this rebaseline does not invent every final constraint in advance.

---

# 17. B1 corrected substrate direction

If promoted, R10-B1 changes materially only where pooled tenancy created the law.

Target substrate:

```text
one MetalDocs product-state PostgreSQL DB
one metaldocs schema
exactly one singleton Tenant/company root
startup/readiness expected_tenant_id handshake

UUID id PRIMARY KEY
ordinary typed UUID FKs
no universal tenant/company/deployment partition column

cross-owner FK = authority-neutral
cross-owner FK actions = RESTRICT / NO ACTION only
no generic polymorphic business registry

TIMESTAMPTZ / BYTEA SHA-256 / TEXT+CHECK / NULL primitives unchanged
persistence-class × mutation-law classification unchanged

no Tenant RLS
no tenant request GUC/context

serving DB role = non-owner + NOSUPERUSER
separate non-serving maintenance trust surface

READ COMMITTED default
explicit local transactions
same-commit required Audit + durable intent
no cross-provider-DB atomicity

outbox/idempotency/lease/retry/DLQ remain
no cross-customer routing machinery
```

Proof obligations added:

```text
singleton Tenant root structural control can be shown to fire
startup/readiness fails closed on expected_tenant_id mismatch
no universal tenant/company/deployment partition columns survive target by inertia
no RLS/tenant-context residues are treated as canonical AuthZ
```

Proof obligations removed as V1 target claims:

```text
same-Tenant composite FK census
Tenant RLS negative proof
tenant-context fail-closed proof
per-Tenant worker discovery proof
```

---

# 18. R10-A bounded implications

8 business bounded contexts + 3 supporting semantic owners remain unchanged.

Organization retains:

```text
Tenant singleton company root
Tenant settings/configuration
Area
User
Group
GroupMembership
```

Organization defers:

```text
Tenant customer lifecycle
TenantDeletionRequest
TenantErasureRecord
Tenant erasure tombstones
```

Re-anchor privacy/offboarding semantics to `User` + Audit/restore successor obligations.

Interchange defers Tenant Portability Export process truth but retains Historical Migration, Governed Subject Export and External Repository import/publish process truth.

No new owner is created.

---

# 19. B2 scope after promotion — corrected candidate

B2 should be re-derived as four internal packages rather than retaining a customer-lifecycle block that no longer exists:

```text
B2-1 Authentication
  provider subject binding
  application Session
  assurance/fresh-auth
  provider lifecycle/reconciliation
  structural anti-corruption proof

B2-2 Organization
  singleton Tenant root + settings
  Area / User / Group / GroupMembership
  User lifecycle/offboarding
  user/data-subject privacy hooks only to the minimum proven semantics

B2-3 Authorization
  Permission / Role / RoleAssignment
  TenantScope | AreaScope
  grant/revocation evidence
  canonical grant evaluation

B2-4 Coherence
  provisioning/offboarding
  session revocation
  membership/grant lifecycle
  singleton-root / expected_tenant_id handshake consumer
  deployment-wide uniqueness sweep
  transaction boundaries
  required same-commit Audit / durable intents
```

Remove from B2 target scope:

```text
tenant dimension in binding/session uniqueness
same-Tenant FK/RLS application
Tenant ACTIVE/SUSPENDED/ERASED
TenantDeletionRequest / TenantErasureRecord / Tenant tombstones
customer tenant routing
password/MFA credential tables
Tenant DEK/KEK machinery
XA/2PC
```

---

# 20. Productization seam and future pooled triggers

Future productization is preserved by:

```text
same codebase/build/migrations
configuration-only company identity/provider endpoints
singleton Tenant UUID as deployment identity/backfill anchor
provider-independent Authentication
provider-independent Artifact storage
clean domain boundaries
repeatable deployment stamp
```

A second customer triggers deployment-economics review, not automatic pooling.

Shared/pooled tenancy re-enters only on measured evidence such as:

```text
fleet infrastructure/ops cost materially breaks target margins
stamp upgrade/backup/patch orchestration exceeds ops capacity despite automation
real cross-company product capability appears
self-service signup economics/provisioning latency demonstrably require pooling
contractual/customer/compliance requirement selects a shared tenancy model
```

Then run a deliberate design stage comparing stamps / DB-per-company / shared app / pooled / hybrid with then-current evidence.

If pooled wins, singleton Tenant UUID is the explicit backfill anchor; adding tenant partition columns/RLS becomes a deliberate migration then, not a permanent V1 tax now.

---

# 21. Confirmed non-reopens

Absent new material evidence, this rebaseline does **not** reopen:

```text
modular monolith
8 + 3 ownership topology
Keycloak selection
Document / Revision / WorkingContent / RevisionSubmission
Approval / SoD
Artifact identity
ManagedArtifactStore + conformance
production malware-inspection gate
Dossier / Evidence
RetentionBinding / LegalHold / Disposition
Audit ownership and same-commit law
Historical Migration
Distribution
Search / Notifications classifications
outbox / durable-intent principles
no OpenFGA / no BPM / no speculative Temporal prerequisite
```

---

# 22. Candidate authority amendments after delta approval

If the bounded delta review approves this corrected target, promotion must amend only the implicated authority/mirror text.

## R9.5 bounded deltas

```text
§1   base permission catalog 29 → 27; remove tenant.export / tenant.deletion.request;
     redefine TenantScope as singleton whole-company scope
§6   customer Tenant lifecycle/deletion/erasure deferred;
     re-anchor user/data-subject privacy + Audit skeleton/GCR-R4 obligations
§9   tenant-namespaced storage-key invariant removed; opaque/immutable key law retained
§13  Tenant Portability Export deferred; Backup/GSE/PUBLISH_COPY remain
```

## R10-A bounded deltas

```text
Tenant = singleton semantic root
Tenant customer lifecycle/deletion/tombstone fact families deferred
user/data-subject privacy obligations re-anchored
Tenant Portability Export process family deferred
8+3 unchanged
```

## R10-B1 material reopen

```text
remove composite tenant-qualified identity/FK laws
remove Tenant RLS/context laws
add singleton Tenant root + expected_tenant_id fail-closed handshake
retain id UUID, typed FKs, cross-owner action law, DB hardening, tx/audit/intent laws
remove cross-customer background discovery/routing assumptions
add future pooled-tenancy reopen triggers
```

## Program authority / handoff

Mirror only the amended laws; route B2 to the four-package scope in §19 and keep implementation blocked.

---

# 23. Bounded delta review request

The next reviewer must inspect only this corrected delta unless it creates material new evidence.

Required checks:

1. Does one-company-per-deployment reduce total V1 complexity without hardcoding Metal Nobre?
2. Is retaining singleton `Tenant` justified by real consumers and clearly separated from DB partitioning?
3. Is exactly-one-root enforcement coherent without an `ACTIVE` lifecycle?
4. Does expected_tenant_id startup/readiness fail-closed preserve wrong-database attachment safety without a new Deployment aggregate?
5. Does removal of tenant-qualified PK/FKs preserve all pooling-independent FK/identity laws?
6. Does removing Tenant RLS preserve actual V1 security while retaining ordinary DB least privilege?
7. Are Keycloak Organizations/routing correctly deferred without weakening federation?
8. Are binding/session tenant dimensions correctly removed while identity/assurance/AuthZ boundaries remain intact?
9. Does TenantScope remain necessary and non-partitioning?
10. Are customer lifecycle/deletion/erasure correctly deferred **without losing user/data-subject privacy obligations**?
11. Are `tenant.export` and `tenant.deletion.request` cleanly removed with no other permission orphaned?
12. Is Tenant Portability safely deferred while Backup/GSE/PUBLISH_COPY remain coherent?
13. Are async correctness mechanisms preserved while cross-customer routing is removed?
14. Is storage key layout correctly freed without weakening opaque/immutable key semantics?
15. Are deployment-wide uniqueness obligations explicitly routed to B2–B5?
16. Did the corrected target accidentally create another global root/company abstraction, dual vocabulary or deployment-control-plane requirement?
17. Is the future pooled-tenancy trigger set concrete enough?
18. Can B2 restart only after authority promotion without another broad review?

Required verdict:

```text
APPROVE SINGLE-COMPANY DEPLOYMENT/TENANCY ADJUDICATED CORRECTED TARGET
APPROVE ... WITH MATERIAL FIXES
DO NOT APPROVE ...
```

Write the result to:

`docs/superpowers/analysis/2026-08-17-single-company-deployment-tenancy-rebaseline-corrected-target-fable-delta-review.md`

---

# 24. Current gate

```text
current authority = unchanged at the pre-rebaseline promoted state
single-company rebaseline = adjudicated corrected candidate only
B2 = PAUSED pending bounded delta review + promotion
implementation = BLOCKED
```

No product implementation, schema/API/frontend change, Keycloak/deployment change or authority promotion is authorized by this artifact.
