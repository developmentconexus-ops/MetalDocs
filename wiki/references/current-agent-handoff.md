# Current Agent Handoff

> **Last verified:** 2026-08-17
> **Status:** ACTIVE — Cohesive Platform Redesign / **R9.5 FROZEN + GCR-REFINED + SINGLE-COMPANY-REFINED / R10-A CLOSED + SINGLE-COMPANY-REFINED / R10-B1 CLOSED + SINGLE-COMPANY-RESTRUCTURED / R10-B2 NEXT**
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Start with `AGENTS.md` and follow its authority chain. For the current stage read:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. **this file** for current status / next step
4. `wiki/architecture/cohesive-platform-redesign.md` — program authority / global coherence
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 product/domain authority including promoted GCR + single-company refinements
6. `wiki/architecture/r10-technical-architecture.md` — active R10 technical authority including the restructured B1 substrate
7. review artifacts only when auditing how a promoted decision was challenged

Git history is archive. Current code/schema/OpenAPI/module docs are current-state evidence only for target design.

---

## Current checkpoint

```text
R3–R9   = LOCKED

R9.5-1  = LOCKED / SINGLE-COMPANY-REFINED where tenancy wording changed
R9.5-2  = LOCKED / GCR-REFINED / SINGLE-COMPANY-REFINED
R9.5-3  = LOCKED (refined by R9.5-8)
R9.5-4  = LOCKED / SINGLE-COMPANY-REFINED where uniqueness wording changed
R9.5-5  = LOCKED (refined by R9.5-8; single-company privacy re-anchor applied)
R9.5-6  = LOCKED / SINGLE-COMPANY-REFINED (Tenant Portability deferred)
R9.5-7  = LOCKED / GCR-REFINED
R9.5-8  = CLOSED / APPROVED
R9.5    = FROZEN / GCR-REFINED / SINGLE-COMPANY-REFINED
reopen set = EMPTY

GCR                         = CLOSED / APPROVED
Single-Company Rebaseline   = CLOSED / APPROVED

R10-A   = CLOSED / APPROVED / GCR-REFINED / SINGLE-COMPANY-REFINED
R10-B   = IN PROGRESS / DESIGN ONLY
R10-B1  = CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED
R10-B2  = NEXT / DESIGN ONLY
R10-B3  = NOT STARTED
R10-B4  = NOT STARTED
R10-B5  = NOT STARTED
R10-B6  = NOT STARTED
R10-C   = NOT STARTED
R10-D   = NOT STARTED
R10-E   = NOT STARTED
R10-F   = NOT STARTED

implementation = BLOCKED
```

Promoted target authority:

- R3–R9.5 product/domain semantics: `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
- R10 technical architecture: `wiki/architecture/r10-technical-architecture.md`
- current program/global-coherence mirrors: `wiki/architecture/cohesive-platform-redesign.md`

---

## Single-Company Deployment / Tenancy Rebaseline — promoted outcome

V1 deployment invariant:

> **One MetalDocs deployment serves exactly one company. The same product codebase, build artifacts and migrations are reused for every deployment; customer-specific forks are forbidden. Shared/pooled multi-customer tenancy is deferred until measured evidence selects it.**

The current deployment serves Metal Nobre through configuration/data, never hardcoded product branches.

### Tenant meaning

`Tenant` remains the durable semantic company root because real consumers remain: company identity/settings, `tenant.settings.manage`, whole-company `TenantScope`, deployment↔database identity binding and a future productization/backfill anchor.

Binding definition:

> **Tenant is the single company/organization root of a deployment. Exactly one Tenant root row exists per MetalDocs product database. Tenant is a semantic root and Authorization scope target, never a database partition dimension in V1.**

There is no Tenant lifecycle in V1. `ACTIVE/SUSPENDED/ERASED`, `TenantDeletionRequest`, `TenantErasureRecord`, company-erasure tombstones and customer decommission workflows are deferred. Deployment stop/decommission belongs to operations.

`Tenant.id` is an immutable trust anchor. Editable company identity/settings are separate mutable facts.

### Deployment ↔ database identity handshake

Deployment configuration pins an `expected_tenant_id` UUID. Startup/readiness reads the singleton database Tenant root and MUST fail closed when:

```text
Tenant root is missing
more than one Tenant root exists
Tenant.id != expected_tenant_id
```

The response to mismatch is to correct deployment configuration or attach/restore the correct database — never mutate the Tenant UUID to make the check pass.

Company-root identity and deployment security profile are distinct properties. R10-C separately owns proof that an inspection-disabled dev/test profile cannot present itself as production.

### R10-B1 substrate after rebaseline

```text
one MetalDocs product-state PostgreSQL DB     = YES
canonical target product schema               = metaldocs
one company / Tenant root per deployment      = YES
universal tenant_id partition column          = NO
composite (tenant_id,id) PK/FK                 = NO
Tenant RLS / Tenant GUC context               = NO
cross-customer worker routing                  = NO
provider-owned DB atomicity dependency         = FORBIDDEN

technical identity                             = UUID id PRIMARY KEY
business/provider/external id as PK             = NO
ordinary typed FK                              = YES
cross-owner FK                                 = authority-neutral
cross-owner FK actions                         = RESTRICT / NO ACTION only
cross-owner CASCADE/SET NULL/SET DEFAULT       = FORBIDDEN
universal polymorphic business registry        = FORBIDDEN

serving DB role                                = non-owner / NOSUPERUSER
canonical Authorization via RLS                = NO
maintenance trust surface                      = separate / non-serving

default transaction isolation                 = READ COMMITTED
cross-owner frozen atomicity                   = one local MetalDocs PostgreSQL transaction
mandatory Audit append                         = same commit when frozen-required
mandatory durable async intent                 = same commit when required
async execution/retry/DLQ                      = R10-D
```

Do not mechanically replace removed `tenant_id` with `company_id`, `organization_id` or `deployment_id`. A reference to the singleton Tenant appears only when that relationship itself has product meaning.

### Async after rebaseline

Remove tenant enumeration, Tenant seeding and `tenant_id` routing metadata. Keep reliability machinery that has an independent failure class:

```text
transactional outbox / durable intent
idempotency
claim/lease
retry
DLQ
truthful external-effect state
```

A worker operates against this deployment's product DB and canonical application/system-execution boundaries; it does not select among customer companies.

### Keycloak after rebaseline

Keycloak remains the V1 AuthN provider. Keycloak Organizations, tenant selector/company switching and realm-per-customer machinery are not V1 product requirements. Each deployment has its own provider configuration/trust domain. `issuer` remains explicit provider identity data so federation/provider migration remains open.

### Authorization after rebaseline

The five roles remain:

```text
tenant_owner
area_manager
author
approver
viewer
```

`TenantScope` remains a real semantic whole-company scope distinct from `AreaScope`; it is not a DB-tenancy mechanism.

Two base permissions are removed because their V1 capabilities are deferred:

```text
tenant.export
tenant.deletion.request
```

Therefore:

```text
R9 base catalog = 27
R9.5 delta       = 16
total V1         = 43
roles            = 5
```

`tenant.settings.manage` remains because company settings are live V1 product state.

### Privacy / data-subject obligations survive customer-lifecycle deferral

Deferring company deletion does **not** defer user/data-subject privacy. The target continues to require:

```text
User offboarding + application-session revocation
separately erasable human-readable/user enrichment
PII-minimized/non-PII immutable Audit skeleton
B6 field-by-field Audit classification
restore must not silently resurrect lawfully erased user/data-subject PII
GCR-R4 reopen if a named immutable Target Data family genuinely requires crypto-erasure
```

Do not invent a generic PrivacyCase/privacy-workflow platform without a real requirement. Retention, LegalHold, Disposition, backup/restore and governed-record preservation remain unchanged.

### Export / storage changes

Tenant Portability Export is deferred. Backup/Restore, Governed Subject Export, Historical Migration and External Repository `IMPORT_COPY` / `PUBLISH_COPY` remain.

Tenant-namespaced object-store keys are no longer an isolation invariant. Opaque immutable keys, exact-byte hash identity, no overwrite and provider independence remain; R10-C may preserve or change current prefixes according to the safest migration path.

### Future shared-tenancy triggers

A second customer alone triggers a deployment-economics review, not automatic pooling. Shared/pooled tenancy re-enters the Method only on measured evidence such as:

1. stamp fleet infrastructure/operations cost materially exceeds sustainable economics;
2. upgrade/patch/backup orchestration exceeds operations capacity despite automation;
3. a genuine cross-company product capability appears;
4. self-service provisioning latency/cost is a demonstrated commercial blocker;
5. a real contractual/compliance requirement selects a shared hosting/tenancy model.

A future design stage then chooses stamps vs shared app + DB-per-company vs pooled vs hybrid using then-current evidence. Singleton `Tenant.id` is the defined backfill anchor if pooling wins.

Review evidence chain:

1. candidate — `docs/superpowers/analysis/2026-08-17-single-company-deployment-tenancy-rebaseline-fable-review-request.md` @ `cba89d9d`;
2. independent cold review — `docs/superpowers/analysis/2026-08-17-single-company-deployment-tenancy-rebaseline-independent-fable-review.md` @ `1acd5128`, verdict `APPROVE ... WITH MATERIAL FIXES`;
3. operator adjudication/corrected target — `docs/superpowers/analysis/2026-08-17-single-company-deployment-tenancy-rebaseline-adjudicated-corrected-target.md` @ `31a57e5b`;
4. bounded delta review — `docs/superpowers/analysis/2026-08-17-single-company-deployment-tenancy-rebaseline-corrected-target-fable-delta-review.md` @ `c87751f3`, verdict `APPROVE ... ADJUDICATED CORRECTED TARGET`.

Final delta:

```text
BLOCKER = 0
MAJOR   = 0
prior findings closed = 9/9
new material contradiction = NONE
```

---

## R10-A promoted outcome after single-company refinement

The semantic ownership topology remains exactly unchanged:

### Business bounded contexts — 8

```text
Authentication
Organization
Authorization
Controlled Information
Approval
Documentary Context
Records Governance
Distribution
```

### Supporting semantic owners — 3

```text
Artifact
Audit
Interchange
```

### Attributed support / projections

```text
Notifications → internal/support/notifications
Search        → internal/projections/search
```

The rebaseline changed deployment/tenancy realization, not the 8+3 domain topology.

Organization owns the singleton Tenant root/settings plus Area/User/Group/GroupMembership and User lifecycle/offboarding semantics. It no longer owns a V1 customer lifecycle/deletion/tombstone family.

Interchange no longer has a V1 Tenant Portability Export process; its remaining transfer authorities are Historical Migration, Governed Subject Export and External Repository copy processes.

---

## R10-B design-package map

```text
B2 → Authentication + Organization + Authorization relational state
B3 → Artifact relational core + Controlled Information + WorkingContent + Submission
B4 → Approval + Controlled Information-owned Rendition/Release/effectivity relational state + Distribution
B5 → Documentary Context + Records Governance + Artifact second-consumer/no-confirmed-orphan closure
B6 → Audit relational state + Interchange batch/plan/outcome state + cross-owner transaction matrix + imported-history/global DB coherence

R10-D → Notifications attributed-support persistence + Search projection persistence + async mechanism persistence/execution
```

This sequencing does not move R10-A ownership.

---

## Exact next step — R10-B2

Open **R10-B2 — Authentication / Organization / Authorization State** in design-only mode, derived from the single-company substrate. Do not copy current auth/IAM/security tables.

B2 should be decomposed into a few coherent packages, while preserving this detailed must-decide surface.

### B2-1 — Authentication binding / Session / assurance

```text
ProviderSubjectBinding representation
  stable issuer + subject identity
  deployment-wide duplicate-binding rejection
  whether one User may bind one or multiple provider subjects
Authentication ↔ Organization User integrity
opaque MetalDocs ApplicationSession representation/lifecycle
fresh-auth / authentication-assurance representation
structural provider anti-corruption proof:
  no provider role/group/org/permission consumption
  no generic claims map into Authorization/domain owners
  no provider-role mapping / claim-to-permission bridge
provider binding/provisioning lifecycle + idempotent reconciliation:
  User exists / provider subject absent
  provider subject exists / binding absent
  binding exists / provider subject removed or disabled
  duplicate issuer+subject attempt
  provider unavailable
  retry after uncertain provider response
provider-side disable vs already-live MetalDocs Session posture
```

### B2-2 — Organization

```text
singleton Tenant root representation
structural exactly-one-root enforcement
Tenant.id = immutable trust anchor
Tenant editable identity/settings persistence
startup/readiness consumer surface for expected_tenant_id handshake
Area / User / Group / GroupMembership persistent state
User lifecycle / offboarding / session-revocation coordination
user/data-subject erasable profile/enrichment boundary needed by B6/R10-C privacy proof
```

No V1 `ACTIVE/SUSPENDED/ERASED`, TenantDeletionRequest, TenantErasureRecord or customer tombstone state.

### B2-3 — Authorization

```text
Permission / Role / RoleAssignment representation
User|Group subject representation
TenantScope whole-company representation
AreaScope representation
grant/revocation evidence
canonical grant-evaluation read surface for later owners
```

### B2-4 — B2 coherence / constraints / transactions

```text
semantic persistence class + mutation-law classification for every B2 family
Tenant root identity immutability proof
former per-tenant uniqueness re-derivation to deployment-wide/semantic scope
ordinary typed FK constraints under B1
transaction boundaries for membership/grant/User-offboarding/session mutations
race-proof new-session issuance vs User offboarding/revocation
required same-commit Audit/durable-intent insertion points
cross-owner coherence cases
```

B2 explicitly does **not** design:

```text
password hash / password-policy tables
credential activation/lockout/MFA/passkey persistence
provider roles/groups/orgs as product authority
claim-to-permission mappings
universal tenant_id/company_id/deployment_id partition columns
Tenant RLS / Area/role/Permission RLS compensation
tenant selector / company switching / cross-customer routing
Tenant customer lifecycle/deletion/tombstone state
Tenant Portability Export
Tenant DEK / key-custody / KEK / wrap-unwrap infrastructure
XA / 2PC across provider and product databases
```

B2 must preserve:

- Authentication ≠ Organization ≠ Authorization;
- exactly five frozen roles;
- flat Groups only V1;
- RoleAssignment subject = User|Group and scope = TenantScope|AreaScope;
- additive/default-deny grants; no `tenant_owner` bypass;
- domain relationship-predicate meaning remains outside Authorization;
- PlatformOperator/SystemPrincipal remain outside company RBAC with no implicit content authority;
- Keycloak is V1 AuthN provider but never canonical Organization/AuthZ authority;
- no OpenFGA/SpiceDB, generic ACL/ReBAC graph, deny engine or nested groups without a real trigger.

### Successor proof obligations

- **B3–B5:** mechanically re-derive every former “unique within tenant” constraint into its real deployment/semantic scope; do not blindly insert a replacement company column.
- **B6:** classify the immutable Audit skeleton field-by-field against user/data-subject privacy. Human-readable enrichment must be separately erasable/read-derived. Reopen GCR-R4 only if a real immutable Target Data family must remain stored but become unintelligible.
- **R10-C:** verify deployment↔database Tenant-root mismatch fails closed in startup/readiness integration; prove restore does not resurrect lawfully erased user/data-subject PII; define ManagedArtifactStore conformance, key-layout freedom, scanner/parser ordering and production profile integrity.
- **R10-D:** preserve outbox/idempotency/lease/retry/DLQ without customer routing; implement provider provisioning/retry/reconciliation after B2 fixes semantic lifecycle.
- **R10-E:** no tenant selector/company switching V1; use provider-hosted/themed login/recovery/MFA journeys where appropriate.
- **R10-F:** cut over legacy tenant_id/RLS/context plumbing, customer-lifecycle/portability surfaces and local credential/DEK machinery only from accepted target mappings; no provider DB atomicity.

Current IAM/auth/security tables, schema and runtime are evidence only. No schema/code implementation is authorized.

---

## Explicitly deferred from launch

```text
shared/pooled multi-customer deployment tenancy
shared backend tenant selector/company switching
Tenant customer lifecycle ACTIVE/SUSPENDED/ERASED
TenantDeletionRequest / TenantErasureRecord / customer tombstones
Tenant Portability Export
Keycloak Organizations as product tenancy machinery
Quarantine aggregate / periodic malware rescans / CDR / advanced active-content security
ArtifactSecurityAssessment domain
ICP-Brasil / PKI / DocuSign / Adobe Sign / RFC3161 / TSA / HSM
cryptographically signed export packages
custom portable export encryption
macro-enabled Office formats
full custom renderer sandbox/egress platform
eDiscovery / generic ESI preservation
realtime coauthoring / WOPI-style collaboration
true indivisible multi-file ArtifactPackage without a real triggering format
OpenFGA/SpiceDB without arbitrary relationship-sharing requirement
BPMN/Camunda/Flowable as Approval prerequisites
Temporal until R10-D proves repeated long-running durable-workflow/timer/retry/compensation machinery
mandatory application-layer tenant encryption / cryptographic erasure without a named Target Data family
self-hosted production object-store provider until a real deployment requires one
```

## Implementation gate

**CLOSED.** No product implementation starts in R10-B2. Product implementation begins only after the integrated R10 technical design is complete, durable target specs/ADRs are promoted as required, material adversarial ambiguity is closed, the operator approves the integrated design, and an implementation plan is authored from that accepted target.
