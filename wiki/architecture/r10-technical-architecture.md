# R10 Technical Architecture — Active Stage Authority

> **Status:** ACTIVE — **R10-A CLOSED / APPROVED / GCR + SINGLE-COMPANY-REFINED; R10-B1 CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED; R10-B2 CLOSED / APPROVED / INTEGRATED; R10-B3 NEXT / DESIGN ONLY**
> **Promoted:** 2026-08-17
> **R10-B1 promotion ratified:** 2026-08-17
> **Global Coherence Review refinement ratified:** 2026-08-17
> **Single-Company Deployment / Tenancy Rebaseline ratified:** 2026-08-17
> **R10-B2-1 promotion ratified:** 2026-08-17
> **R10-B2 integrated promotion ratified:** 2026-08-17
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
  B2   Authentication / Organization / Authorization             CLOSED / APPROVED / INTEGRATED
    B2-1 Authentication binding / Session / assurance            CLOSED / APPROVED
    B2-2 Organization singleton root / people / groups           CLOSED / APPROVED / INTEGRATED
    B2-3 Authorization                                           CLOSED / APPROVED / INTEGRATED
    B2-4 B2 coherence / constraints / transactions / privacy     CLOSED / APPROVED / INTEGRATED
  B3   Controlled Information + Artifact relational core         NEXT / DESIGN ONLY
  B4   Approval + CI-owned Rendition/Release + Distribution      NOT STARTED
  B5   Documentary Context / Records + Artifact closure           NOT STARTED
  B6   Audit / Interchange / Cross-owner Atomicity                NOT STARTED
R10-C  Artifact / Records Physical Integrity                     NOT STARTED
R10-D  Durable Async / Projections / External Effects            NOT STARTED
R10-E  Canonical Access / API / Frontend Journeys                NOT STARTED
R10-F  Historical Migration / Cutover / Final Deletion           NOT STARTED
```

Closure order: `R10-A → B1 → B2 → B3 → B4 → B5 → B6 → C → D → E → F`.

B2-2/B2-3/B2-4 were closed as one integrated batch after full independent review and bounded delta review; they are not separate reopen gates. B-blocks sequence design work; they do not move semantic ownership. Product implementation remains blocked until integrated R10 closes.

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
UserProfile
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
Permission static product catalog
Role static product catalog
RoleAssignment
subject: User | Group
scope: TenantScope | AreaScope
grant/revocation evidence contract
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

Exactly one durable `Tenant` root exists. B2 fixes structural exactly-one enforcement. `Tenant.id` is immutable UUID; editable identity/settings are separate mutable facts.

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

# 7. R10-B2-1 — Authentication binding / ApplicationSession / assurance — promoted

R10-B2-1 is **CLOSED / APPROVED**. It fixes the minimum MetalDocs-owned authentication state around Keycloak without moving credential authority, Organization identity or Authorization into the provider.

> **Integrated closure note:** references in this section that originally routed exact Organization/AuthZ/locking/privacy work to B2-2/B2-3/B2-4 are closure provenance only. Section 8 below is now the promoted integrated B2 authority and no B2 substage remains open.

## 7.1 Exactly two Authentication semantic persistent families

```text
ProviderSubjectBinding
ApplicationSession
```

No third semantic family exists for provider sync state, provider account mirror, provider claims, assurance-event history, provider tokens, provider Organizations/groups/roles or retry/error state. Provider execution truth belongs to R10-D mechanism persistence. Approval/domain consumers persist their own decision evidence.

## 7.2 ProviderSubjectBinding

Target shape:

```text
ProviderSubjectBinding

id          UUID PRIMARY KEY
user_id     UUID NOT NULL REFERENCES Organization.User(id)
issuer      TEXT NOT NULL
subject     TEXT NOT NULL
created_at  TIMESTAMPTZ NOT NULL
disabled_at TIMESTAMPTZ NULL

UNIQUE (issuer, subject)
UNIQUE (user_id) WHERE disabled_at IS NULL
```

Cross-owner FK actions are `RESTRICT` / `NO ACTION` only.

Stable provider identity is `(issuer, subject)`. `user_id`, `issuer`, and `subject` are immutable mapping fields. Email, username, display name and provider role/group/org facts are never identity.

Acceptance state is represented only by `disabled_at`:

```text
disabled_at IS NULL     → binding currently accepted
disabled_at IS NOT NULL → binding currently disabled
```

The same mapping may be disabled and re-enabled by clearing `disabled_at`; Audit owns transition history. Re-enable never resurrects revoked Sessions.

V1 permits at most one currently accepted binding per User. Upstream federation occurs inside Keycloak and normally presents one MetalDocs-facing issuer+subject. Simultaneously active MetalDocs-facing issuers are a reopen trigger.

Total `UNIQUE(issuer,subject)` is deliberate: while the correlation row exists, one stable provider subject cannot be handed over to a different MetalDocs User. Same-subject re-trust for the same User uses the same row. Subject handover/User merge/provider-subject reuse is exceptional reopen-tier behavior, never a normal V1 flow. After lawful erasure of the row, a later correlation is a new binding decision under the same creation laws; the structural no-recorrelation guarantee was deliberately surrendered by erasure.

## 7.3 Binding creation / correlation authority

Provider subject selection must have **causal and verifiable correlation** to the exact provisioning/correlation intent, or be an explicit trusted-human correlation decision.

Permitted authority:

1. provider operation for the exact intent creates/returns the subject;
2. an intent-bound provider-side correlation/idempotency mechanism proves which subject resulted from that exact intent;
3. explicit trusted human correlation designates the subject.

Never subject-selection authority:

```text
email
username
display name
similar name
provider "already exists" + matching attribute
```

Human attributes may be corroborating display information only. Provider conflict/already-exists cannot silently adopt an account; it requires an intent-bound proof or explicit trusted correlation.

Valid provider authentication without an accepted binding never creates a MetalDocs User or Session and never auto-binds by attributes.

## 7.4 ApplicationSession

Target shape:

```text
ApplicationSession

id                          UUID PRIMARY KEY
subject_binding_id          UUID NOT NULL REFERENCES ProviderSubjectBinding(id)
credential_digest           BYTEA NOT NULL
created_at                  TIMESTAMPTZ NOT NULL
expires_at                  TIMESTAMPTZ NOT NULL
revoked_at                  TIMESTAMPTZ NULL
latest_reauthenticated_at   TIMESTAMPTZ NULL
latest_provider_auth_time   TIMESTAMPTZ NULL
latest_acr                  bounded nullable representation
latest_amr                  bounded nullable representation

UNIQUE (credential_digest)
CHECK (expires_at > created_at)
```

Cross-owner/reference actions use `RESTRICT` / `NO ACTION` where applicable.

Browser/session credential properties:

- high-entropy opaque bearer credential;
- database stores only a one-way verifier/digest, never replayable raw bearer;
- row disclosure must not yield a browser-replayable credential;
- absolute lifetime is finite;
- multiple Sessions per User are allowed;
- `revoked_at` is terminal; revoked Session never becomes active again;
- expiry is derived from `expires_at`; no persisted ACTIVE/EXPIRED enum;
- reauthentication never resurrects revoked/expired Session.

ApplicationSession contains no persisted Tenant/company dimension, duplicated `user_id`, roles, permissions, groups, provider access/refresh/ID-token authority, email/username, IP/User-Agent/LastSeen semantic state or idle-timeout state. Runtime authenticated context resolves `Session → Binding → User`; canonical Authorization reads current authority and is never snapshotted into Session.

`session.manage` UX may later require bounded support/mechanism telemetry such as device/IP/last-seen; R10-E may add it only for a concrete UX/security consumer. Such telemetry is not ApplicationSession semantic authority.

Provider token material may exist transiently/request-scoped if a later verified Keycloak journey needs it; it never becomes normal ApplicationSession authority.

## 7.5 Fresh-auth / assurance contract

Session `latest_*` assurance fields are **evidence inputs only**. Their presence or non-NULL value never by itself satisfies `requires_reauthentication`.

`requires_reauthentication` means an explicit provider authentication challenge completed for a bounded operation context. Consumer satisfaction must use an explicit bounded rule:

```text
one-shot evidence tied to the operation
OR
an explicitly configured freshness window owned by the consuming policy authority
```

No implicit/unbounded freshness window exists. Initial login does not automatically satisfy a later reauthentication requirement.

Forced reauth must validate the **same `(issuer,subject)`** as the Session's accepted binding. Different subject fails closed and requires a new login/new Session. Callback/final assurance update succeeds only if the Session remains non-revoked/non-expired; reauth cannot revive Session state.

Authentication may publish transient/value-object `FreshAuthEvidence` carrying bounded facts such as `session_id`, local `verified_at`, provider `auth_time?`, `acr?`, `amr?`. It does not persist a competing assurance-event history. Approval/B4 owns approval freshness policy and snapshots its consumed evidence in its own decision authority.

## 7.6 Provider disable, availability and offboarding

Provider-only disable/removal does not synchronously revoke every existing MetalDocs Session by assumption. V1 contract:

```text
provider-only disable/removal
→ new login fails
→ forced reauth fails
→ reconciliation may disable binding + revoke Sessions earlier
→ otherwise existing ApplicationSession survives no longer than local revoke or finite expires_at
```

The finite absolute Session TTL is therefore also the maximum provider-only-disable staleness bound absent earlier reconciliation. R10-E/deployment security configuration chooses the actual TTL value with that consequence explicit.

Keycloak outage:

```text
existing established ApplicationSessions may continue locally
new login fails visibly
forced reauth fails visibly
provider provisioning/reconciliation retries through R10-D
```

MetalDocs User offboarding is authoritative for MetalDocs access: Organization marks the User ineligible and local Sessions are revoked immediately without waiting for Keycloak. Provider disable/provisioning is an asynchronous external effect.

Role/group/grant changes do **not** require Session revocation because Session contains no AuthZ snapshot; canonical Authorization sees current grants on the next check.

## 7.7 Binding disable/re-enable/replacement and Session revocation

Disabling an accepted binding revokes all Sessions referencing that binding in the same local MetalDocs transaction. Replacement to a new provider subject occurs only after the new subject is causally/explicitly confirmed, then atomically:

```text
old binding disabled
new binding created/enabled
old-bound Sessions revoked
required Audit
```

Re-enable is also an acceptance mutation. It permits future Session issuance; it never revives terminally revoked Sessions.

`UNIQUE(user_id) WHERE disabled_at IS NULL` is the structural DB backstop against two simultaneously enabled bindings for one User.

## 7.8 Provider provisioning / reconciliation

Semantic outcomes remain the six required cases:

1. User exists / provider subject absent → User remains; no binding; no login.
2. provider subject exists / binding absent → no access until causally proven or explicitly trusted correlation creates binding.
3. binding exists / provider subject removed or disabled → reconciliation may disable binding and revoke Sessions; later trusted restoration may re-enable the same row.
4. duplicate `(issuer,subject)` attempt → idempotent only when it is the same existing mapping; conflicting mapping fails closed under total uniqueness.
5. provider unavailable → existing Sessions continue subject to local rules; provider operations retry.
6. uncertain provider response → never fabricate binding truth; reconcile by exact intent-bound correlation or retry.

No semantic `provider_sync_status` / provider-shadow FSM is introduced. Provider attempts, retry, lease, error and pending-correlation mechanism state belong to R10-D. Provider calls occur outside MetalDocs DB transactions; no XA/2PC/cross-provider atomicity claim exists.

Typical provisioning choreography:

```text
local tx #1:
  Organization.User mutation
  required Audit
  durable provider-provisioning intent
COMMIT

R10-D provider effect / reconciliation

local tx #2 after subject is proven:
  ProviderSubjectBinding mutation
  required Audit
COMMIT
```

## 7.9 Privacy / persistence classification

- `ProviderSubjectBinding` is Authentication semantic authority with immutable mapping fields and reversible acceptance; it is erasable under lawful user/data-subject cleanup and is not a governed-retention subject.
- `ApplicationSession` is Authentication semantic authority with mutable bounded assurance, finite expiry and terminal revocation; operational rows may be erased under lifecycle/privacy rules and are not governed-retention subjects.
- erase dependent ApplicationSessions before ProviderSubjectBinding under RESTRICT reference law;
- Audit/governed evidence must not FK-depend on Authentication rows for historical validity.

Lawful erasure of a binding row surrenders the DB-level structural guarantee against later re-correlation of that erased subject; a later binding is a new trusted correlation decision.

## 7.10 B2-1 review / closure evidence

Evidence chain:

1. candidate — `docs/superpowers/analysis/2026-08-17-r10-b2-1-authentication-binding-session-assurance-fable-review-request.md` @ `9cba3acd`;
2. independent cold review — `...-independent-fable-review.md` @ `361f6c8b`, verdict `APPROVE ... WITH MATERIAL FIXES`;
3. corrected target — `...-adjudicated-corrected-target.md` @ `ee0a0ce0`;
4. bounded delta review — `...-corrected-target-fable-delta-review.md` @ `6593c471`, verdict `APPROVE R10-B2-1 ADJUDICATED CORRECTED TARGET`.

Final delta: `BLOCKER=0`, `MAJOR=0`, `prior findings closed=8/8`, `new material contradiction=NONE`, `new concurrency counterexample=NONE`, broad review not required.

B2-1 reopens only on material evidence such as simultaneous MetalDocs-facing provider bindings becoming required, real provider subject reuse/handover between Users, an accepted immediate provider-initiated revocation requirement, a consumer proving additional Session semantic state is essential, or an assurance consumer that cannot be represented without changing Authentication ownership.

---

# 8. R10-B2 — Integrated Authentication / Organization / Authorization — promoted

R10-B2 is **CLOSED / APPROVED / INTEGRATED**. B2-2 Organization, B2-3 Authorization and B2-4 coherence were reviewed and closed as one system around the already-promoted B2-1 Authentication contract.

Integrated invariant:

> **A MetalDocs request acts as one eligible organizational User reached through one accepted provider binding and one valid local ApplicationSession. Effective product authority is derived live from current direct/group RoleAssignments over typed Tenant/Area scopes, static product Role→Permission bundles, domain-owned relationship predicates and domain governance constraints. Identity, authentication, organization, authorization, audit evidence and provider execution each have one owner.**

The target keeps one company per deployment, additive grants/default deny, no `tenant_owner` bypass, no provider role/group/org/permission authority, no universal partition column, no Tenant/Area/role/Permission RLS as canonical AuthZ, no generic ACL/ReBAC/deny engine and no nested groups.

## 8.1 Organization persistent state

### Tenant

```text
Tenant
  id           UUID PRIMARY KEY
  display_name TEXT NOT NULL
```

`Tenant.id` is immutable deployment↔DB trust anchor. `display_name` is mutable company identity/settings. No slug/status/customer lifecycle/generic settings JSON V1.

Structural at-most-one:

```text
UNIQUE INDEX ON tenant ((true))
```

Startup/readiness supplies at-least-one and `expected_tenant_id` matching. Missing root, multiple roots or mismatch fail closed. Combined serving invariant = exactly one Tenant root.

### Area

```text
Area
  id          UUID PRIMARY KEY
  code        TEXT NOT NULL UNIQUE
  name        TEXT NOT NULL
  disabled_at TIMESTAMPTZ NULL
```

`id/code` immutable; `name` mutable. `disabled_at IS NULL` accepts new references/assignments; non-NULL means retired while existing references remain valid. Retirement is reversible for the same Area identity.

Retired Area:

```text
existing Documents/history/grants remain valid
new Document Area assignment       → fail closed at Controlled Information
new AreaScope RoleAssignment       → fail closed at Authorization
new Approval policy Area reference → fail closed at Approval
```

No Area hierarchy, owner field, default approver role or generic metadata platform V1.

### User

```text
User
  id          UUID PRIMARY KEY
  disabled_at TIMESTAMPTZ NULL
```

`id` is immutable stable organizational identity. No username/email/provider subject/credential/role/capability/tenant_id/home_area/employee key. `disabled_at NULL` = eligible; non-NULL = ineligible. Disable/re-enable preserves identity.

### UserProfile

```text
UserProfile
  user_id      UUID PRIMARY KEY REFERENCES User(id)
  display_name TEXT NOT NULL
  email        TEXT NULL
```

Strict subordinate one-to-one state. `User` is stable governed identity; `UserProfile` is erasable human-readable/contact enrichment. Normally eligible User is profile-complete; absence means lawful erasure or bounded provisioning transition. Consumers use neutral/opaque fallback rather than fabricated data. Email/display name are attributes, never technical identity or binding authority; no `UNIQUE(email)` identity law.

### Group

```text
Group
  id   UUID PRIMARY KEY
  name TEXT NOT NULL UNIQUE
```

Flat, company-wide V1. No code, area scope, provider-group mirror, nested group, dynamic rule or retirement lifecycle.

Hard deletion remains allowed only when no live reference exists. B2 memberships/RoleAssignments must be absent; every persisted live cross-owner Group reference uses typed FK `Group(id)` with RESTRICT/NO ACTION. Known B4 consumers are ApprovalPolicy `Group` actor rules and live Distribution Group audience configuration. Historical snapshots resolve concrete Users.

### GroupMembership

```text
GroupMembership
  user_id  UUID NOT NULL REFERENCES User(id)
  group_id UUID NOT NULL REFERENCES Group(id)
  PRIMARY KEY (user_id, group_id)
```

Current truth only: row exists = current member. No surrogate UUID, interval/tombstone/history family or nested membership. Audit owns add/remove transition evidence.

## 8.2 Role / Permission static product authority

V1 persists no editable `permissions`, `roles`, `role_permissions` or custom-role bundle state. Authorization owns versioned-with-product static Permission and Role catalogs; DB persistence contains only current RoleAssignments.

Current V1 catalog = 43 permissions: 27 R9 base permissions after single-company removal of `tenant.export` and `tenant.deletion.request`, plus 16 R9.5 additions.

### Exact current Role→Permission bundles — single current technical home

#### viewer — 3

```text
document.read_effective
evidence.read
dossier.read
```

#### author — 15

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

#### approver — 4

```text
document.read_effective
approval.act
evidence.read
dossier.read
```

Approver has no blanket working/history access; exact Approval participation opens the case-specific Submission/evidence required by frozen relationship rules.

#### area_manager — 25

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

Area manager is operational, not RBAC/configuration administrative. It has no `access.manage`, `organization.manage`, tenant/config, audit/session or whole-company lifecycle administration.

#### tenant_owner — all 43

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

`tenant_owner` is an ordinary role bundle, never a bypass. Domain relationships/state/SoD/fresh-auth remain binding.

Mechanical implementation proof must keep:

```text
static Role codes
== RoleAssignment role_code CHECK vocabulary
== role↔scope CHECK vocabulary
```

Drift fails verification. CHECKs are enforcement, not a second catalog.

## 8.3 RoleAssignment — sole persisted Authorization family

```text
RoleAssignment
  id UUID PRIMARY KEY

  user_id  UUID NULL REFERENCES User(id)
  group_id UUID NULL REFERENCES Group(id)

  role_code TEXT NOT NULL

  tenant_scope_id UUID NULL REFERENCES Tenant(id)
  area_scope_id   UUID NULL REFERENCES Area(id)
```

Cross-owner FK actions = RESTRICT / NO ACTION.

Structural subject XOR: exactly one `user_id | group_id` non-NULL.

Structural scope XOR: exactly one `tenant_scope_id | area_scope_id` non-NULL.

No generic polymorphic subject/scope registry.

Role vocabulary CHECK contains exactly:

```text
tenant_owner
area_manager
author
approver
viewer
```

Role↔scope compatibility:

```text
tenant_owner → TenantScope only
area_manager → AreaScope only
author       → TenantScope | AreaScope
approver     → TenantScope | AreaScope
viewer       → TenantScope | AreaScope
```

DB CHECK makes illegal pairs unrepresentable; application validates the same invariant for friendly failure.

Every B3–B5/R10-E permission check declares whether the target is Tenant-wide or Area-targeted. Tenant-owner-only whole-company permissions remain Tenant-wide even when a resource has an Area relation.

Duplicate current-grant backstops:

```text
UNIQUE(user_id, role_code, tenant_scope_id)
  WHERE user_id IS NOT NULL AND tenant_scope_id IS NOT NULL
UNIQUE(user_id, role_code, area_scope_id)
  WHERE user_id IS NOT NULL AND area_scope_id IS NOT NULL
UNIQUE(group_id, role_code, tenant_scope_id)
  WHERE group_id IS NOT NULL AND tenant_scope_id IS NOT NULL
UNIQUE(group_id, role_code, area_scope_id)
  WHERE group_id IS NOT NULL AND area_scope_id IS NOT NULL
```

Current-truth mutation law:

```text
INSERT → grant exists
DELETE → grant revoked
```

Grant shape is immutable while row exists; change = revoke + new grant. No retained revoked interval or temporal-grant scheduler V1. Required grant/revocation Audit is in the same local transaction; re-grant creates new RoleAssignment UUID and evidence.

RoleAssignment needs UUID because the XOR union has no single NULL-free composite PK; four partial uniques cannot jointly be a PostgreSQL table PK. GroupMembership's NULL-free pair remains sufficient relationship identity.

`tenant_scope_id → Tenant(id)` is semantic whole-company scope, not partitioning.

## 8.4 Canonical Authorization evaluation / administration

No semantic persistence of:

```text
user_permissions
effective_permissions
cached group-expanded grants
Session roles/permissions
materialized ACL
provider-role mapping
```

Live evaluation:

```text
current direct User RoleAssignments
UNION
current GroupMemberships → current Group RoleAssignments
→ static Role → Permission bundle
→ scope match
→ domain relationship predicate when required
→ domain governance constraints
→ ALLOW or default DENY
```

Role/grant/membership changes take effect on the next canonical check without Session regeneration.

Scope application:

```text
Tenant-wide check → qualifying TenantScope assignment required
Area-targeted check → qualifying TenantScope OR matching AreaScope
```

Administration permissions:

```text
tenant.settings.manage → Tenant editable identity/settings
organization.manage    → Area/User/UserProfile/Group identity & lifecycle
access.manage          → GroupMembership + RoleAssignment
session.manage         → explicit administrative ApplicationSession management
```

Exact frozen bundles grant these administration permissions only to `tenant_owner`. Since `tenant_owner` is TenantScope-only:

```text
Organization administration        = TenantScope tenant_owner V1
GroupMembership administration     = TenantScope access.manage only
RoleAssignment administration      = TenantScope access.manage only
```

There is **no Area-local RBAC administrator V1**. Tenant owner may grant AreaScope roles to Users/Groups; access administration itself is not delegated.

New direct RoleAssignment to disabled User, new GroupMembership for disabled User and new AreaScope RoleAssignment to retired Area all fail closed. Existing AreaScope grants remain valid after Area retirement.

## 8.5 User offboarding / re-enable

Offboarding is destructive access teardown in one local MetalDocs transaction:

```text
BEGIN
lock User
set User.disabled_at = now
revoke all ApplicationSessions for User
  // terminal revoked_at; do not erase Session rows here
delete all GroupMemberships for User
delete all direct User RoleAssignments
append required Audit
insert durable provider-disable intent when required
COMMIT
```

ProviderSubjectBinding remains because issuer+subject→User correlation remains truthful. Provider effect is post-commit R10-D work. Group RoleAssignments remain; removing memberships removes inherited access.

Re-enable:

```text
BEGIN
lock User
clear User.disabled_at
append required Audit
insert provider-enable durable intent when required
COMMIT
```

No prior membership/direct grant is restored and no revoked Session resurrects. Default deny holds until explicit fresh access configuration. No separate temporary-suspension state V1; intentional automatic restoration is a reopen trigger.

Area retirement/re-enable only changes future assignability; existing refs/grants remain valid and retirement does not act as deny-all.

Group hard delete requires no membership rows, no Group RoleAssignments and no live cross-owner typed references.

## 8.6 Deterministic B2 lock law

B1 isolation remains READ COMMITTED. B2 uses narrow row locks + FK/UNIQUE/CHECK enforcement; no global SERIALIZABLE/advisory-lock framework.

Canonical acquisition order; classes may be skipped but never revisited backwards:

```text
1. User row
2. ProviderSubjectBinding rows for that User, ascending id
3. Area row
4. child sets in ascending PK order:
   ApplicationSession → id
   GroupMembership    → (user_id, group_id)
   RoleAssignment     → id
```

Group deletion is isolated:

```text
Group FOR UPDATE
→ GroupMembership rows ascending user_id
→ Group RoleAssignments ascending id
```

A Group-deletion transaction never then acquires User/Binding/Area locks. Concurrent Group-subject RoleAssignment or membership inserts serialize against Group deletion through the FK/Group-row conflict; implementation specs must state both cases explicitly.

Lock modes:

```text
eligibility/acceptance readers → FOR SHARE
lifecycle mutators             → FOR UPDATE
```

`FOR KEY SHARE` is insufficient for `disabled_at` serialization because non-key updates may take `FOR NO KEY UPDATE` without the required conflict.

Required race outcomes:

- Session issuance vs offboarding: issuance first is swept; offboarding first blocks issuance.
- binding disable/re-enable/replacement vs issuance: issuance only from accepted binding; disable/replacement revokes affected Sessions; re-enable never revives them.
- GroupMembership/direct User grant vs offboarding: mutation first is removed; offboarding first causes mutation to fail.
- AreaScope grant vs Area retirement: grant first survives as existing; retirement first blocks new grant.
- re-enable restores eligibility only, never deleted access rows/revoked Sessions.
- GroupMembership + Group RoleAssignment need no atomic coupling; effective group authority exists exactly when both current facts exist.

B2 guarantees fail-closed future Session resolution/Authorization after lifecycle commit. It does not claim global cancellation of a request that already completed its relevant authn/authz decision unless the business transaction shares a specific lock/invariant.

## 8.7 Persistence class × mutation law

```text
Tenant             SEMANTIC AUTHORITY — id immutable; display_name mutable
Area               SEMANTIC AUTHORITY — id/code immutable; name/disabled_at mutable
User               SEMANTIC AUTHORITY — id immutable; disabled_at mutable
UserProfile        SEMANTIC AUTHORITY subordinate enrichment — mutable/erasable
Group              SEMANTIC AUTHORITY — id immutable; name mutable; hard delete only unreferenced
GroupMembership    SEMANTIC AUTHORITY current relationship — INSERT/DELETE
RoleAssignment     SEMANTIC AUTHORITY current grant — immutable shape while present; INSERT/DELETE
ProviderSubjectBinding promoted Authentication semantic authority
ApplicationSession     promoted Authentication semantic authority
```

No historical grant/membership interval family. Audit is transition timeline, not current grant authority.

## 8.8 Audit / durable provider-intent / privacy

Administrative B2 mutation that changes identity, eligibility, binding acceptance or effective access appends required Audit evidence in the **same local transaction**.

At minimum this covers Tenant display/settings, Area create/rename/retire/re-enable, User create/offboard/re-enable, governed UserProfile mutation/erasure, Group create/rename/delete, GroupMembership add/remove, RoleAssignment grant/revoke, ProviderSubjectBinding acceptance/replacement, administrative Session revocation and offboarding.

Grant/revocation Audit must preserve enough PII-minimized facts for forensic reconstruction after RoleAssignment deletion: assignment id, subject reference, role, scope, actor, operation and trusted time subject to B6 final field classification.

Provider-side effect pattern:

```text
BEGIN
  local semantic truth
  required Audit
  required durable provider intent
COMMIT
→ R10-D executes/retries/reconciles provider effect
```

No provider HTTP call participates in local DB atomicity.

Privacy separation:

```text
erasable when lawful:
  UserProfile
  ApplicationSession rows after lifecycle/evidence need ends
  ProviderSubjectBinding when lawful

retained skeleton when governed history requires it:
  User.id
  User.disabled_at
  governed domain User UUID references
  PII-minimized/non-PII Audit skeleton
```

Offboarding is not privacy erasure. Session revocation is not Session deletion. Lawful Binding erasure surrenders structural no-recorrelation for the erased subject. B6 finalizes Audit field privacy; R10-C proves restore non-resurrection. No generic privacy workflow/platform.

## 8.9 Bootstrap / recovery / naming

Default deny + no bypass requires a distinct **non-serving/request-unreachable maintenance trust surface** for initial `tenant_owner` RoleAssignment seeding and admin-lockout recovery. It is never a permanent authorization bypass. R10-F specifies the procedure; R10-E may add UX warnings/guards only as defense-in-depth.

Required display names/codes reject unusable blank/whitespace forms and deliberately normalize inputs in implementation specs. Human display casing is not identity; Group name case-insensitive uniqueness is not implied; Tenant display name is not routing key; Area name is not Area identity.

## 8.10 Review / closure evidence

Integrated evidence chain:

```text
candidate                  b814f672  integrated B2 candidate
independent full review    34a567fd  APPROVE WITH MATERIAL FIXES
corrected target           2908a884  operator-adjudicated corrected candidate
bounded delta review       507075a8  APPROVE

full-review result:
  BLOCKER = 0
  MAJOR   = 3
  LOW     = 5

delta result:
  BLOCKER = 0
  MAJOR   = 0
  LOW     = 2 non-blocking notes
  prior findings closed = 3/3 MAJOR + 5/5 LOW + D15
  A1 tenant-owner-only access administration = APPROVE
  exact 5×43 bundle verification = MATCH
  deadlock under corrected law = NONE
  new material contradiction = NONE
  B2-1 reopen = NO
  reopen outside B2 = NO
  broad review required = NO
```

The two final LOW notes are not promotion conditions:

- Group-subject RoleAssignment insert vs Group deletion uses the same FK/Group-row conflict already proven; implementation spec names it explicitly.
- this R10 page is the **single current technical home** for the exact 5×43 Role→Permission bundles; staging/review artifacts and the frozen ledger remain provenance, not parallel current bundle authority.

## 8.11 Successor obligations

B3–B5/R10-E permission check sites declare Tenant-wide vs Area-targeted. Tenant-owner-only whole-company families remain Tenant-wide.

B4 must consume retired-Area new-policy rejection, typed Group FKs with RESTRICT for ApprovalPolicy/Distribution live configuration, concrete-User snapshots and bounded fresh-auth policy.

B5 has no speculative Group requirement; any real future Group reference obeys the same typed-FK RESTRICT law.

B6 finalizes Audit skeleton field-by-field privacy, grant/revoke forensic fields and same-commit cross-owner Audit matrix.

R10-C proves restore non-resurrection of lawfully erased user PII.

R10-D executes provider provisioning/disable/enable/reconciliation durable intents with retry/idempotency/lease/DLQ without becoming semantic authority.

R10-E consumes provider-hosted auth journeys, Session TTL, per-check-site scope classification, neutral historical actor fallback and optional last-admin UX protection.

R10-F/operations specifies initial tenant_owner seed, lockout recovery, legacy 8-role/capability and dual-grant-table cutover, legacy tenant_id/RLS/context removal, and static-catalog↔DB-CHECK parity gate.

## 8.12 Integrated proof obligations

Later implementation specification/tests must prove at minimum:

1. Tenant singleton at-most-one DB constraint + at-least-one/matching readiness handshake.
2. Tenant.id immutable.
3. Area.code immutable/deployment-wide unique; retirement/re-enable semantics.
4. User eligibility is one `disabled_at` fact; no provider/AuthZ/PII identity duplication.
5. UserProfile absence is valid erasable enrichment with neutral fallback.
6. Group hard delete fails on any live cross-owner reference.
7. GroupMembership pair PK prevents duplicates.
8. static catalogs contain exactly the accepted 5 roles/43 permissions and exact bundles in §8.2.
9. catalog↔role CHECK↔role-scope CHECK drift is mechanically detected.
10. RoleAssignment subject XOR and scope XOR hold at DB level.
11. illegal role↔scope pairs fail at DB level.
12. four partial uniqueness backstops reject duplicate current grants.
13. RoleAssignment UUID is stable technical PK; grant shape immutable while present.
14. Authorization is live/additive/default-deny with no effective-permission semantic store or Session AuthZ snapshot.
15. ordinary V1 access administration is TenantScope tenant-owner-only.
16. disabled User cannot receive new direct grant, membership or Session.
17. retired Area cannot receive new AreaScope grant/reference while existing refs/grants remain valid.
18. offboarding revokes Sessions, deletes memberships/direct grants and appends Audit in one local transaction.
19. offboarding retains binding correlation and provider work is durable post-commit choreography.
20. re-enable restores eligibility only and never old access rows/Sessions.
21. canonical lock order/modes eliminate reviewed deadlock/race classes under READ COMMITTED.
22. Group deletion vs concurrent membership/grant/reference creation fails closed.
23. same-commit Audit exists for material B2 identity/eligibility/access mutations.
24. grant/revoke Audit remains forensic after current RoleAssignment deletion without becoming current AuthZ authority.
25. provider calls never participate in local DB atomicity.
26. privacy cleanup erases enrichment/auth state without breaking governed User references.
27. no universal tenant/company/deployment partition column re-enters through B2.
28. no provider role/group/org/claim bridge re-enters Authorization.
29. no custom-role/permission/ACL/ReBAC/deny/nested-group platform appears without reopen evidence.
30. maintenance bootstrap/recovery is non-serving/request-unreachable and does not become a bypass.

## 8.13 Reopen triggers

B2 reopens only on material evidence such as simultaneously active multiple MetalDocs-facing provider identities/User; legitimate provider subject handover/reuse; mandatory immediate provider-initiated Session revocation; real HR/workforce Area placement independent of Authorization; nested/dynamic/scoped Groups; custom roles/bundles; temporal grants; explicit deny semantics; arbitrary resource sharing requiring ReBAC-class machinery; temporary User suspension with intentional automatic authority restoration; materially different Area retirement semantics; or a new permission that changes frozen bundles.

Current implementation inconvenience or legacy table shape is never a reopen trigger.

---

# 9. Exact next step — R10-B3 Controlled Information + Artifact relational core

Open **R10-B3 — Controlled Information + Artifact relational core** in design-only batch mode. R10-B2 is closed; do not rediscover Authentication/Organization/Authorization decisions.

B3 begins with one integrated sweep/decomposition, not microdecisions. It must derive the minimum relational state, constraints and same-commit transaction laws for the B1-assigned surface:

```text
Artifact core
  immutable exact-byte identity/hash/size/format/media-type facts needed by CI
  staging/confirmation ownership seam needed by Submission
  no provider-specific key/layout authority
  no B5 Documentary Context/Records artifact relationships yet

Controlled Information configuration
  DocumentType
  optional DocumentTypeCategory navigation/classification only
  numbering configuration / sequence allocation facts
  Tenant Dictionary + System Value Catalog persistence only where frozen semantics require it

Document / Revision
  stable Document identity/code/type/Area/responsibility
  DocumentRevision identity / REV labels / lifecycle
  exactly-one-open and at-most-one-effective structural laws
  immutable/stable identity fields vs mutable draft-cycle facts

Working content
  format-agnostic WorkingContent authority
  working_version OCC
  technical snapshots/checkpoints/editor-session state only where a real invariant requires persistence
  no autosave/checkpoint as business Revision

RevisionSubmission
  immutable exact submission attempt
  governed submission digest/content identity
  source Artifact relationship / frozen source bytes
  same-REV return/resubmit creates a new Submission without mutating old attempt
  NoHumanApproval still creates Submission

Template role/use
  template = governed Document role, not parallel lifecycle
  TemplateUse M:N + at-most-one UX default/type where still frozen
  TemplateSpec / structured-authoring state only where applicable
  immutable DocumentOrigin/provenance when creating from template

Editorial / periodic-review CI state
  EditorialComment if material
  PeriodicReviewPolicy / PeriodicReviewRecord / responsible-owner relation
  stale-review protection against changed effective REV

Atomicity / constraints
  code allocation + Document + REV001 creation
  draft WorkingContent/OCC mutation
  SUBMIT freeze: semantic content + exact Artifact/Submission identity in one coherent boundary
  reason-for-change / numbering / immutable code and Area/DocumentType laws
  former per-tenant uniqueness re-derived to deployment/semantic scope
  required Audit/durable intent points routed to B6/R10-D without duplicating authority
```

B3 must explicitly separate what belongs later:

```text
Approval policy/instance/decision/fresh-auth consumption → B4
Rendition + effectivity/Release                          → B4
Distribution                                             → B4
Evidence/Dossier/Records Governance artifact relations  → B5
Audit/Interchange final cross-owner matrix               → B6
malware/storage physical integrity/relocation/restore    → R10-C
async execution/projections/provider effects             → R10-D
API/frontend journeys                                    → R10-E
historical cutover/deletion                              → R10-F
```

The first B3 deliverable is an **integrated intake/decomposition and candidate relational system**, followed by one serious adversarial review at the batch level. Do not return to per-table/per-field review ceremony unless a genuinely independent failure class requires it.

Current documents/controlleddocuments/templates/render schema/code remain current-state evidence only. Product implementation remains **BLOCKED**.