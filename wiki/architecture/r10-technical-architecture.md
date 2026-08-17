# R10 Technical Architecture — Active Stage Authority

> **Status:** ACTIVE — **R10-A CLOSED / APPROVED / GCR + SINGLE-COMPANY-REFINED; R10-B1 CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED; R10-B2 IN PROGRESS; R10-B2-1 CLOSED / APPROVED; R10-B2-2 NEXT / DESIGN ONLY**
> **Promoted:** 2026-08-17
> **R10-B1 promotion ratified:** 2026-08-17
> **Global Coherence Review refinement ratified:** 2026-08-17
> **Single-Company Deployment / Tenancy Rebaseline ratified:** 2026-08-17
> **R10-B2-1 promotion ratified:** 2026-08-17
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
  B2   Authentication / Organization / Authorization             IN PROGRESS / DESIGN ONLY
    B2-1 Authentication binding / Session / assurance            CLOSED / APPROVED
    B2-2 Organization singleton root / people / groups           NEXT / DESIGN ONLY
    B2-3 Authorization                                           NOT STARTED
    B2-4 B2 coherence / constraints / transactions / privacy     NOT STARTED
  B3   Controlled Information + Artifact relational core         NOT STARTED
  B4   Approval + CI-owned Rendition/Release + Distribution      NOT STARTED
  B5   Documentary Context / Records + Artifact closure           NOT STARTED
  B6   Audit / Interchange / Cross-owner Atomicity                NOT STARTED
R10-C  Artifact / Records Physical Integrity                     NOT STARTED
R10-D  Durable Async / Projections / External Effects            NOT STARTED
R10-E  Canonical Access / API / Frontend Journeys                NOT STARTED
R10-F  Historical Migration / Cutover / Final Deletion           NOT STARTED
```

Closure order: `R10-A → B1 → B2-1 → B2-2 → B2-3 → B2-4 → B3 → B4 → B5 → B6 → C → D → E → F`.

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

# 7. R10-B2-1 — Authentication binding / ApplicationSession / assurance — promoted

R10-B2-1 is **CLOSED / APPROVED**. It fixes the minimum MetalDocs-owned authentication state around Keycloak without moving credential authority, Organization identity or Authorization into the provider.

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

Provider token material may exist transiently/request-scoped if a later verified Keycloak journey needs it (for example logout mechanics); it never becomes normal ApplicationSession authority.

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

Authentication may publish transient/value-object `FreshAuthEvidence` carrying bounded facts such as `session_id`, local `verified_at`, provider `auth_time?`, `acr?`, `amr?`. It does not persist a competing assurance-event history. Approval/B4 owns approval freshness policy and snapshots its consumed evidence in its own decision authority. If Approval chooses one-shot evidence, single-consumption semantics belong to that consuming policy/transaction design rather than a new Authentication semantic table.

## 7.6 Provider disable, availability and offboarding

Provider-only disable/removal does not synchronously revoke every existing MetalDocs Session by assumption. V1 contract:

```text
provider-only disable/removal
→ new login fails
→ forced reauth fails
→ reconciliation may disable binding + revoke Sessions earlier
→ otherwise existing ApplicationSession survives no longer than local revoke or finite expires_at
```

The finite absolute Session TTL is therefore also the maximum provider-only-disable staleness bound absent earlier reconciliation. R10-E/deployment security configuration chooses the actual TTL value with that consequence explicit; B2-1 does not select a number.

Keycloak outage:

```text
existing established ApplicationSessions may continue locally
new login fails visibly
forced reauth fails visibly
provider provisioning/reconciliation retries through R10-D
```

MetalDocs User offboarding is different and authoritative for MetalDocs access: Organization marks the User ineligible and local Sessions are revoked immediately without waiting for Keycloak. Provider disable/provisioning is an asynchronous external effect.

Role/group/grant changes do **not** require Session revocation because Session contains no AuthZ snapshot; canonical Authorization sees current grants on the next check.

## 7.7 Binding disable/re-enable/replacement and Session revocation

Disabling an accepted binding is an Authentication authority mutation and revokes all Sessions referencing that binding in the same local MetalDocs transaction. Replacement to a new provider subject occurs only after that new subject is causally/explicitly confirmed, then atomically:

```text
old binding disabled
new binding created/enabled
old-bound Sessions revoked
required Audit
```

Re-enabling an old binding is also an acceptance mutation and participates in the same B2-4 concurrency/serialization discipline as disable/replacement/session issuance. Re-enable permits future Session issuance; it never revives terminally revoked Sessions.

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

B2-4 must classify exact persistence/mutation laws, but B2-1 fixes these semantic properties:

- `ProviderSubjectBinding` is Authentication semantic authority with immutable mapping fields and reversible acceptance (`disabled_at`); it is erasable under lawful user/data-subject cleanup and is not a governed-retention subject.
- `ApplicationSession` is Authentication semantic authority with mutable bounded assurance, finite expiry and terminal revocation; operational rows may be erased under lifecycle/privacy rules and are not governed-retention subjects.
- erase dependent ApplicationSessions before ProviderSubjectBinding under RESTRICT reference law;
- Audit/governed evidence must not FK-depend on Authentication rows for historical validity; governed decision evidence references MetalDocs User/domain authority, while Audit's surviving skeleton is independently privacy-safe per B6.

Lawful erasure of a binding row necessarily surrenders the DB-level structural guarantee against later re-correlation of that erased subject; a later binding is a new §7.3-governed correlation decision. B2-4/B6 must make this consequence explicit in the final persistence/privacy proof without inventing a privacy platform.

## 7.10 Concurrency invariants routed to B2-4

B2-1 declares the outcomes; B2-4 selects exact transaction/lock realization under B1 `READ COMMITTED`.

C1 — login/session issuance vs User offboarding:

```text
Either Session issuance commits before offboarding and is swept by offboarding revocation,
or offboarding commits first and issuance sees User ineligible and creates no Session.
Forbidden: offboarding reports success while a concurrently issued valid Session survives.
```

C2 — reauth callback vs Session revoke/expiry: final assurance update succeeds only against still-valid Session; revoke/expiry cannot be undone.

C3 — Session issuance vs binding disable/re-enable/replacement: issuance can commit only from a currently accepted binding under the same serialization discipline; disabling/replacing revokes affected Sessions in the same local tx. Re-enable is explicitly included in this discipline. Partial unique index remains the DB backstop for one enabled binding/User.

C4 — grant mutation vs existing Session: safe without Session revocation because AuthZ is never cached in Session.

A credible minimal B2-4 realization is row-lock serialization on the User eligibility row for C1 and on the binding row for C3 plus the DB uniqueness backstop; exact SQL/locking remains B2-4 authority.

## 7.11 Proof obligations

Later design/implementation must prove at minimum:

- total `(issuer,subject)` uniqueness and one-enabled-binding/User under concurrent writes;
- mapping fields cannot silently change correlation;
- disable/re-enable/replacement semantics preserve Session revocation and do not resurrect Sessions;
- no email/username/display-name or provider role/group/org attribute can select a subject or grant access;
- valid provider authentication without accepted binding cannot create User/Session;
- raw bearer never stored; Session-row disclosure is not replayable;
- every Session has finite absolute lifetime;
- no Session contains canonical AuthZ snapshots or provider-token authority;
- fresh-auth satisfaction is explicitly bounded and same-subject pinned;
- reauth cannot revive revoked/expired Session;
- provider outage does not invalidate established Sessions merely by outage;
- provider-only disable staleness is bounded by remaining Session lifetime and may be shortened by reconciliation;
- User offboarding revokes local access without provider dependency and races safely with issuance;
- binding disable/re-enable/replacement races safely with issuance;
- uncertain provider outcomes never fabricate Binding truth;
- provider mechanism state stays R10-D, not semantic provider-shadow state;
- lawful privacy cleanup can erase Session/Binding without rewriting retained governed evidence;
- no operation depends on atomicity across Keycloak/provider DB and MetalDocs DB.

## 7.12 Review / closure evidence

Evidence chain:

1. candidate — `docs/superpowers/analysis/2026-08-17-r10-b2-1-authentication-binding-session-assurance-fable-review-request.md` @ `9cba3acd`;
2. independent cold review — `docs/superpowers/analysis/2026-08-17-r10-b2-1-authentication-binding-session-assurance-independent-fable-review.md` @ `361f6c8b`, verdict `APPROVE ... WITH MATERIAL FIXES`, `BLOCKER=0`, `MAJOR=3`, `LOW=5`;
3. operator adjudication/corrected target — `docs/superpowers/analysis/2026-08-17-r10-b2-1-authentication-binding-session-assurance-adjudicated-corrected-target.md` @ `ee0a0ce0`;
4. bounded delta review — `docs/superpowers/analysis/2026-08-17-r10-b2-1-authentication-binding-session-assurance-corrected-target-fable-delta-review.md` @ `6593c471`, verdict `APPROVE R10-B2-1 ADJUDICATED CORRECTED TARGET`.

Final delta:

```text
BLOCKER = 0
MAJOR   = 0
prior findings closed = 8/8
new material contradiction = NONE
new concurrency counterexample = NONE
broad review required = NO
```

Successor notes from the delta review:

- **DL1 → B2-4/B6:** explicitly record that lawful Binding erasure surrenders the structural no-recorrelation guarantee for the erased subject; later correlation is a new trusted decision.
- **DL2 → B2-4:** re-enable is an acceptance mutation under the same C3 serialization discipline as disable/replacement.
- **DL3:** the current F2 wording is sufficient; a future wording cleanup may say the correlation marker is created by execution of the exact intent, but no pre-promotion amendment is required.

B2-1 reopens only on material evidence such as simultaneous MetalDocs-facing provider bindings becoming required, real provider subject reuse/handover between Users, an accepted immediate provider-initiated revocation requirement, a consumer proving additional Session semantic state is essential, or an assurance consumer that cannot be represented without changing Authentication ownership.

---

# 8. Exact next step — R10-B2-2 Organization

Open **R10-B2-2 — Organization singleton root / people / groups** in design-only mode. B2-1 is closed and must be consumed, not rediscovered.

B2-2 must decide the minimum persistent Organization state for:

```text
Tenant singleton root representation
  exact table/fact shape
  structural exactly-one-root enforcement
  Tenant.id immutable trust anchor
  editable company identity/settings facts
  startup/readiness consumer surface for expected_tenant_id handshake

Area
  identity / stable fields
  lifecycle if any is actually required
  deployment-wide uniqueness law

User
  technical identity
  authentication-eligibility/offboarding state
  erasable profile/enrichment boundary
  relationship to Area if any real semantic requirement exists
  deletion/disable/offboarding semantics without provider-state mirroring

Group
  flat-group identity
  deployment-wide uniqueness law
  lifecycle if any is actually required

GroupMembership
  User↔Group relationship
  mutation/evidence semantics
  no nested groups

B2-1 integration facts
  User eligibility consumed by Session issuance
  offboarding must revoke ApplicationSessions race-safely
  ProviderSubjectBinding FK target/integrity
  no provider subject/email/username/provider group becomes Organization authority
```

B2-2 must not design Authorization grants/RoleAssignment (B2-3), final cross-owner lock/transaction matrix (B2-4), provider provisioning mechanics (R10-D), frontend admin journeys (R10-E), or resurrect customer Tenant lifecycle/partitioning/RLS.

Keep these B2-4 successor obligations visible while designing Organization state:

- exactly-one Tenant root proof and immutable Tenant UUID;
- C1 login/session issuance vs User offboarding serialization;
- C3 binding acceptance mutations vs Session issuance serialization;
- lawful Binding-erasure no-recorrelation consequence;
- same-commit Audit/durable-intent points for Organization mutations.

Current IAM/auth/security schema/code remain current-state evidence only. No product implementation is authorized.
