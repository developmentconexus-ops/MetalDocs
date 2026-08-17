# R10 Technical Architecture — Active Stage Authority

> **Status:** ACTIVE — **R10-A CLOSED / APPROVED / GCR + SINGLE-COMPANY-REFINED; R10-B1 CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED; R10-B2 NEXT / DESIGN ONLY**
> **Promoted:** 2026-08-17
> **R10-B1 promotion ratified:** 2026-08-17
> **Global Coherence Review refinement ratified:** 2026-08-17
> **Single-Company Deployment / Tenancy Rebaseline ratified:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** `docs/engineering/standards/root-cause-global-maximum-method.md`
> **Frozen product/domain authority:** `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
> **Program authority:** `wiki/architecture/cohesive-platform-redesign.md`
> **Implementation gate:** **CLOSED — no product implementation authorized.**

This page is the durable technical-stage authority. Frozen product/domain semantics remain in the ledger and are not duplicated here unless needed to fix technical ownership, persistence or proof obligations.

---

## 1. R10 decomposition and closure order

```text
R10-A  Ownership Topology & Dependency DAG                       CLOSED / APPROVED / REFINED
R10-B  Transactional Domain State & DB Invariants                IN PROGRESS
  B1   Relational Substrate, Deployment Identity & Reference Law CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED
  B2   Authentication / Organization / Authorization             NEXT / DESIGN ONLY
  B3   Controlled Information + Artifact relational core         NOT STARTED
  B4   Approval + CI-owned Rendition/Release + Distribution      NOT STARTED
  B5   Documentary Context / Records + Artifact closure           NOT STARTED
  B6   Audit / Interchange / Cross-owner Atomicity                NOT STARTED
R10-C  Artifact / Records Physical Integrity                     NOT STARTED
R10-D  Durable Async / Projections / External Effects            NOT STARTED
R10-E  Canonical Access / API / Frontend Journeys                NOT STARTED
R10-F  Historical Migration / Cutover / Final Deletion           NOT STARTED
```

Closure order: `R10-A → B1 → B2 → B3 → B4 → B5 → B6 → C → D → E → F`.

B-blocks sequence design work; they do not move semantic ownership. Product implementation remains blocked until integrated R10 closes.

---

# 2. R10-A — promoted ownership topology after refinements

The V1 ownership set remains exactly **8 business bounded contexts + 3 supporting semantic owners**.

## 2.1 Business bounded contexts

### Authentication

Owns:

```text
ProviderSubjectBinding
opaque MetalDocs ApplicationSession
application-session lifecycle/revocation
authentication-assurance / fresh-auth evidence
provider anti-corruption contract
```

Keycloak owns credential mechanisms, password policy/recovery, provider lockout, MFA/passkeys, upstream federation and provider authentication journeys/session. Stable provider identity is `issuer + subject`; email/username/display name are attributes. Provider roles/groups/organizations/permissions/arbitrary claims cannot become canonical MetalDocs Authorization input; no mapping bridge exists V1.

Keycloak Organizations/company switching are not V1 product requirements. Each deployment carries its own provider configuration/trust domain.

### Organization

Owns:

```text
Tenant singleton company root
Tenant settings/configuration
Area
User
Group
GroupMembership
User lifecycle / offboarding
privacy-sensitive user/profile enrichment state needed by product operation
```

Binding law: exactly one Tenant root exists per product DB; `Tenant.id` is immutable; Tenant editable identity/settings are mutable facts. Tenant is a whole-company semantic root, **not a V1 DB partition dimension**.

Deferred V1 customer-company facts:

```text
Tenant ACTIVE/SUSPENDED/ERASED
TenantDeletionRequest
TenantErasureRecord
customer-company erasure tombstones
```

Deployment maintenance/decommission is operations. No mandatory Tenant DEK/key-custody family exists V1.

### Authorization

Owns:

```text
Permission
Role
RoleAssignment
subject: User | Group
scope: TenantScope | AreaScope
grant/revocation evidence
canonical grant evaluation
composable authorization/filter contract shape
```

`TenantScope` means the whole company represented by the singleton root; it is not DB tenancy. Relationship predicate meaning remains with each semantic domain. No bypass, generic ACL/ReBAC graph, provider permission engine or deny engine V1.

### Controlled Information

Owns DocumentType, Document/Revision, owner/responsibility, WorkingContent/Snapshots/EditorSession, RevisionSubmission, numbering, template role/use/spec, EditorialComment, periodic review, Rendition, representation policy, Release/effectivity, Tenant Dictionary and System Value Catalog semantics exactly as frozen in the ledger.

### Approval

Owns versioned policy/steps, exact-Submission ApprovalInstance, activated participant snapshot, ApprovalDecision, attestation/fresh-auth evidence, return/withdraw/cancel/reassign/oversight and strict SoD. Approval never owns Document effectivity.

### Documentary Context

Owns EvidenceType/Evidence, DossierType/Dossier, ExternalReference, Dossier↔Document context, Evidence primary/secondary Dossier context and Evidence→primary Artifact relationship. Context never grants access.

### Records Governance

Owns retention-rule meaning, RetentionBinding, RetentionExtension, LegalHold/materialization, disposition eligibility/authorization and DispositionRecord. No V1 customer-company erasure workflow exists. Where user/data-subject privacy intersects retained governed records, privacy-sensitive enrichment must be separated rather than rewriting retained evidence.

### Distribution

Owns released-document obligations, audience snapshot/historical denominator, AcknowledgementRecord and completion semantics; never grants access.

## 2.2 Supporting semantic owners

### Artifact

Owns immutable exact-byte identity, canonical SHA-256, size/format/media type, staging/validation/confirmation, managed physical location, relocation verification and restore byte integrity. No confirmed orphan Artifact. Production confirmation requires successful malware inspection; scanner mechanism/order belongs R10-C.

### Audit

Owns append-only AuditEvent timeline, tamper evidence, query/export semantics and its separate retention regime. Critical governed mutation appends Audit in the same local DB commit when frozen-required.

Privacy proof: immutable Audit state surviving lawful user/data-subject PII erasure must be PII-minimized/non-PII; human-readable enrichment is separately erasable/read-derived. If B6 proves named immutable Target Data must remain stored but become unintelligible, GCR-R4 reopens before crypto-erasure machinery is added.

### Interchange

Owns Historical Migration batch/plan/dry-run/outcomes/reconciliation, Governed Subject Export process and External Repository IMPORT_COPY/PUBLISH_COPY process truth. Tenant Portability Export is deferred; whole-stamp movement uses Backup/Restore unless a real portability/product-exit contract appears.

## 2.3 Support / projections / mechanisms

```text
Notifications → internal/support/notifications   // attributed delivery/read support only
Search        → internal/projections/search      // rebuildable projection only
Composition   → cross-owner coordination, no durable meaning
Platform      → DB/HTTP/async/observability/provider mechanisms
```

Not semantic owners: workers, outbox/queue/leases/DLQ, Keycloak client, storage client, malware scanner, renderer, external-repo adapters, HTTP/codegen, cache/rate-limit/observability, backup transport, PlatformOperator/SystemPrincipal execution machinery.

Tenant RLS/Tenant-context seeding/customer routing are **not V1 target mechanisms**. No compensating Area/role/Permission RLS is introduced.

---

# 3. R10-A coordination / dependency rules

`internal/composition` may coordinate owners but owns no business fact. Semantic owners expose transactionally composable application seams whenever one local DB atomic invariant requires multiple owners.

Material seams:

1. CI↔Approval exact Submission/evidence contracts; no mutual authority.
2. One local MetalDocs DB transaction for frozen cross-owner atomicity.
3. Audit publishes transactionally composable append.
4. Artifact confirmation uses opaque owner reference; Artifact does not import CI/DC.
5. Records prospective hold materialization consumes owner facts without taking lifecycle ownership.
6. Historical Migration calls narrow privileged target-owner seams; target owners do not depend on Interchange.
7. Notifications receives already-resolved delivery intent; it invents no policy.
8. Authorization composes owner predicates; Search/export/timeline/API do not rederive visibility.
9. User offboarding/privacy: Organization owns User/profile lifecycle, Authentication owns Session revocation, Audit owns surviving skeleton, Records governs retained evidence, Artifact owns bytes.
10. Authentication provider: no operation assumes atomic commit across Keycloak/provider persistence and MetalDocs DB.
11. Deployment↔database identity: configured `expected_tenant_id` is compared with the singleton DB Tenant root at startup/readiness; missing/multiple/mismatch fails closed. This is a boundary check, not a Deployment aggregate.

The package/import DAG remains acyclic.

---

# 4. Target filesystem and legacy disposition

```text
internal/
  modules/
    authentication/
    organization/
    authorization/
    controlledinformation/
    approval/
    documentarycontext/
    recordsgovernance/
    distribution/
    audit/
    interchange/
  support/
    artifacts/
    notifications/
  projections/
    search/
  composition/
  platform/
```

Provider placement:

```text
Keycloak / IdP adapter                 → Authentication infrastructure
Local / AWS S3 / compatible adapters   → Artifact infrastructure
malware scanner adapter                → Artifact/platform validation mechanism
EigenPal / rendering provider adapters → Controlled Information infrastructure/execution
SharePoint / external repo adapters    → Interchange infrastructure
```

Durable storage entitlement = `ManagedArtifactStore` port + conformance, not provider list. Object key layout is provider freedom; tenant/company prefix is not a V1 isolation invariant.

Legacy target disposition:

```text
approval              → Approval
audit                 → Audit
auth                  → Authentication provider binding/app Session/assurance; local credentials delete/migrate
documents + controlleddocuments → Controlled Information
distribution          → Distribution
iam                   → Organization + Authorization
jobs                  → platform async/composition; customer routing removed
notifications         → support/notifications
render                → CI Rendition semantics + provider infra
search                → projections/search
security              → split Authentication/platform; tenant-RLS/context + DEK/KEK legacy machinery has no V1 entitlement
taxonomy              → Area→Organization, DocumentType→CI, GovernanceClass deleted
templates             → CI template role/use
tokens                → CI Tenant Dictionary/System Value Catalog
```

---

# 5. R10-A closure / reopen record

R10-A topology remains exactly 8+3. GCR refined Authentication/provider and key-custody assumptions; Single-Company Rebaseline redefined Tenant, deferred customer lifecycle/portability and removed shared-tenancy substrate assumptions without creating a new owner.

Single-company review chain:

```text
candidate                cba89d9d
independent review       1acd5128  APPROVE WITH MATERIAL FIXES
corrected target          31a57e5b
bounded delta review      c87751f3  APPROVE
BLOCKER                   0
MAJOR                     0
prior findings closed     9/9
new material contradiction NONE
```

R10-A reopens only on material evidence: an unowned frozen fact, independent lifecycle/consumer requiring boundary change, genuine relationship graph complexity, independent Rendition semantic consumer, material transfer-boundary split, indivisible multi-file Artifact semantics, named immutable Target Data requiring key lifecycle, or real customer-lifecycle/portability facts requiring promotion.

---

# 6. R10-B1 — single-company relational substrate, deployment identity and reference law

R10-B1 is **CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED**. The one-company-per-deployment requirement invalidated only the former pooled/shared-tenancy PK/FK/RLS/routing laws.

## 6.1 PostgreSQL topology / singleton Tenant root / identity

```text
one MetalDocs product-state PostgreSQL database
canonical product-state schema = metaldocs
one company / Tenant root per deployment = YES
schema-per-bounded-context = NO
```

Provider DBs retain separate authority. No MetalDocs invariant may require cross-database atomicity with provider-owned persistence; no XA/2PC.

Exactly one durable `Tenant` root exists. B2 must choose structural exactly-one enforcement and prove it can reject a duplicate. `Tenant.id` is immutable UUID; editable identity/settings are separate mutable facts.

Deployment config pins `expected_tenant_id`. Startup/readiness fails closed on missing root, multiple roots or mismatch. Never mutate root UUID to satisfy configuration.

For ordinary durable entities:

```text
id UUID PRIMARY KEY
```

No universal partition column. Do not replace removed `tenant_id` with `company_id`, `organization_id` or `deployment_id` by reflex. Reference the Tenant only where the relationship itself has semantic meaning.

Business/provider/external identifiers never become technical PKs. Singleton Tenant UUID is the defined future backfill anchor if a deliberate pooled design ever wins.

## 6.2 Typed references / FK actions

Ordinary typed UUID FK:

```text
FOREIGN KEY (target_id) REFERENCES target_table(id)
```

Cross-owner FK proves existence/identity only; never transfers authority.

```text
cross-owner DELETE/UPDATE = RESTRICT | NO ACTION only
cross-owner CASCADE / SET NULL / SET DEFAULT = FORBIDDEN
```

Within one owner cascade is non-default and only for strictly subordinate state without independent historical meaning.

Generic polymorphic business registries remain rejected; typed relationships belong to the owner of the relationship. Audit resource attribution is allowed because Audit is non-authoritative for resource state.

## 6.3 Persistence primitives

```text
technical IDs     = UUID
business instants = TIMESTAMPTZ
canonical SHA-256 = BYTEA + octet_length(hash)=32
frozen vocabulary = TEXT + CHECK by default
real unknown      = NULL
```

`JSONB` only for bounded whole snapshots or genuinely variable provider-neutral provenance; not an escape hatch. Historical snapshots are never silently rewritten.

## 6.4 Persistence class × mutation law

Every persisted family declares:

```text
semantic class: SEMANTIC AUTHORITY | ATTRIBUTED SUPPORT | DURABLE MECHANISM | EPHEMERAL MECHANISM | REBUILDABLE PROJECTION
mutation law:   MUTABLE | IMMUTABLE/APPEND-ONLY | TERMINAL/TOMBSTONED | REBUILDABLE | explicit state machine
```

Examples:

```text
Tenant.id                = SEMANTIC AUTHORITY + IMMUTABLE identity anchor
Tenant editable settings = SEMANTIC AUTHORITY + MUTABLE
RevisionSubmission       = SEMANTIC AUTHORITY + IMMUTABLE
ApprovalDecision         = SEMANTIC AUTHORITY + IMMUTABLE
AuditEvent               = SEMANTIC AUTHORITY + APPEND-ONLY
async/outbox intent      = DURABLE MECHANISM + constrained operational state
```

“The application normally does not update it” is never sufficient proof of immutability.

## 6.5 Database security / Authorization boundary

No Tenant RLS stack exists V1 because one company exists in the DB. Do not introduce Area/role/Permission RLS as compensation. Canonical Authorization remains application/domain authority.

Serving least privilege remains:

```text
ordinary serving role = NOSUPERUSER + not owner of protected product tables
DDL/object ownership  = separate
```

`NOBYPASSRLS` is not a V1 authority requirement because no Tenant RLS exists.

## 6.6 Durable async intent / claim surface

Required durable intent is inserted in the same local transaction as the business mutation when future async work is necessary.

Claim surfaces expose bounded mechanism facts only:

```text
intent identity
kind / due time
lease/claim state
opaque target reference
```

No `tenant_id` customer-routing field. No arbitrary business content merely for dispatch convenience.

## 6.7 Database roles / maintenance

Migration/backfill/restore uses a distinct non-serving maintenance trust surface. Elevated maintenance identity must never be ordinary serving identity, request-reachable, product Authorization or implicit content bypass. No per-Tenant iteration rule exists V1.

## 6.8 Transaction / isolation law

Ordinary product mutations use explicit local MetalDocs PostgreSQL transactions.

```text
single-owner → owner service boundary
cross-owner frozen atomic use case:
  composition opens one transaction
  → published owner seams share it
  → one COMMIT or ROLLBACK
```

No owner hides nested commit or imports another owner's repository to obtain atomicity. Provider-side work is outside local atomicity and uses durable intent/idempotency/reconciliation.

Default isolation = `READ COMMITTED`; use narrowest sufficient UNIQUE/CHECK/FK/CAS/lock/atomic-update mechanism before raising isolation.

## 6.9 Same-commit Audit / durable intent

```text
BEGIN
  authoritative business/support facts
  required Audit append
  required durable mechanism intent
COMMIT
```

or all local facts roll back. External/provider effects execute later and reconcile truthfully.

## 6.10 Background work

One deployment serves one company. No Tenant enumeration, Tenant-context seeding or cross-customer routing.

```text
claim/query due work from this deployment DB
→ load canonical state through owner/application boundaries
→ execute
```

Outbox, idempotency, lease, retry and DLQ survive where independently required.

## 6.11 Package assignment / namespace

```text
B2 → Authentication + Organization + Authorization
B3 → Artifact + Controlled Information + WorkingContent + Submission
B4 → Approval + CI-owned Rendition/Release/effectivity + Distribution
B5 → Documentary Context + Records Governance + Artifact closure
B6 → Audit + Interchange + cross-owner tx/imported-history DB coherence
R10-D → Notifications + Search + async mechanism persistence/execution
```

`metaldocs` remains final product-state namespace. R10-F owns target-vs-legacy mapping, including removal/cutover of legacy `tenant_id`/RLS/context/routing/customer-lifecycle surfaces.

## 6.12 Closure proofs

Later design/implementation must prove:

- exactly-one Tenant root and duplicate-root rejection;
- immutable Tenant UUID;
- fail-closed startup/readiness on missing/multiple/mismatched `expected_tenant_id`;
- every cross-owner FK uses only RESTRICT/NO ACTION;
- no universal company partition column reappears under another name without semantic need;
- ordinary serving pool is non-owner/NOSUPERUSER;
- required Audit/durable intent shares authoritative mutation commit;
- every B-block classifies semantic persistence + mutation law;
- no provider DB atomicity assumption;
- R10-D claim surfaces stay bounded mechanism metadata without customer routing;
- R10-F maintenance elevation is non-serving and request-unreachable;
- B2–B5 mechanically re-derive former per-tenant uniqueness to actual deployment/semantic scope.

## 6.13 Reopen triggers

B1 reopens on material identity/transaction/reference/maintenance evidence, not current-schema inconvenience.

Shared/pooled customer tenancy re-enters only on measured evidence:

1. stamp fleet infra+ops cost materially exceeds sustainable economics;
2. upgrade/patch/backup orchestration exceeds operations capacity despite automation;
3. genuine cross-company product capability requires shared plane;
4. self-service provisioning latency/cost becomes a demonstrated commercial blocker;
5. contractual/compliance requirement selects shared tenancy/hosting.

A second customer alone triggers a deployment-economics review, not automatic pooling. A deliberate future stage chooses stamps vs shared-app/database-per-company vs pooled vs hybrid. Singleton `Tenant.id` is the backfill anchor if pooled partitioning wins.

---

# 7. Exact next step — R10-B2

Open **R10-B2 — Authentication / Organization / Authorization State** in design-only mode. Use four coherent packages without shrinking the detailed checklist.

## B2-1 — Authentication binding / Session / assurance

```text
ProviderSubjectBinding representation
  issuer + subject stable boundary
  deployment-wide duplicate-binding rejection
  one User ↔ one/multiple provider subjects
Authentication ↔ User integrity
opaque ApplicationSession lifecycle
fresh-auth / assurance representation
structural anti-corruption proof:
  no provider role/group/org/permission consumption
  no generic claims map
  no provider-role/claim-to-permission bridge
provider provisioning/binding reconciliation:
  User exists / subject absent
  subject exists / binding absent
  binding exists / subject removed or disabled
  duplicate attempt
  provider unavailable
  uncertain-response retry
provider-side disable vs already-live Session posture
```

## B2-2 — Organization

```text
singleton Tenant root representation
structural exactly-one-root enforcement
Tenant.id immutable trust anchor
Tenant editable identity/settings persistence
startup/readiness consumer surface for expected_tenant_id handshake
Area / User / Group / GroupMembership state
User lifecycle / offboarding / Session-revocation coordination
privacy-sensitive erasable profile/enrichment boundary needed by B6/R10-C
```

No V1 customer `ACTIVE/SUSPENDED/ERASED`, TenantDeletionRequest, TenantErasureRecord or tombstone state.

## B2-3 — Authorization

```text
Permission / Role / RoleAssignment
User|Group principals
TenantScope whole-company | AreaScope
role/grant/revocation evidence
canonical grant-evaluation read model
```

## B2-4 — coherence / constraints / transactions / privacy hooks

```text
semantic persistence + mutation-law classification for every B2 family
Tenant root identity immutability proof
former per-tenant uniqueness re-derivation
ordinary typed FK application
membership/grant/User-offboarding/Session tx boundaries
race-proof Session issuance vs offboarding/revocation
required same-commit Audit/durable-intent points
cross-owner coherence cases
```

B2 does **not** design local credentials/MFA/password/lockout, provider role mappings, universal company partition columns, Tenant/Area/role/Permission RLS, company selector/switching/routing, customer deletion/tombstones/Portability, Tenant DEK/KEK or XA/2PC.

### Successor obligations

- **B3–B5:** re-derive all former “unique within tenant” constraints to actual deployment/semantic scope; do not add replacement company column by reflex.
- **B6:** field-by-field PII-minimized/non-PII Audit skeleton proof; human-readable enrichment separately erasable/read-derived; reopen GCR-R4 only for named immutable Target Data.
- **R10-C:** prove singleton-root handshake integration, restore non-resurrection of lawfully erased user PII, ManagedArtifactStore conformance/key-layout freedom, scanner/parser ordering and production-profile integrity.
- **R10-D:** preserve outbox/idempotency/lease/retry/DLQ without customer routing; provider retry/reconciliation; no distributed transaction.
- **R10-E:** no company selector/switching V1; provider-hosted/themed auth journeys.
- **R10-F:** cut over legacy shared-tenancy `tenant_id`/RLS/context/routing/customer-lifecycle/portability plus local credential/DEK machinery only from accepted mappings.

Current code/schema/runtime are evidence only. No product implementation is authorized.
