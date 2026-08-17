# R10 Technical Architecture — Active Stage Authority

> **Status:** ACTIVE — **R10-A CLOSED / APPROVED / GCR-REFINED; R10-B1 CLOSED / APPROVED / GCR-CLARIFIED; R10-B2 NEXT / DESIGN ONLY**
> **Promoted:** 2026-08-17
> **R10-B1 promotion ratified:** 2026-08-17
> **Global Coherence Review refinement ratified:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** `docs/engineering/standards/root-cause-global-maximum-method.md` — DevelopmentConexus Engineering Method v1.0.0 mirror
> **Frozen product/domain authority:** `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` for R3–R9.5
> **Program authority:** `wiki/architecture/cohesive-platform-redesign.md`
> **Implementation gate:** **CLOSED — no product implementation authorized.**

This page is the durable stage authority for R10. R10-A and R10-B1 remain closed; the 2026-08-17 Global Coherence Review made only the bounded amendments recorded below. The 8+3 ownership topology and all unaffected B1 substrate laws remain unchanged.

---

## 1. R10 decomposition and closure order

R10 is one integrated technical-design stage decomposed by failure class:

```text
R10-A  Ownership Topology & Dependency DAG              CLOSED / APPROVED / GCR-REFINED
R10-B  Transactional Domain State & DB Invariants      IN PROGRESS
  B1   Relational Substrate, Tenancy & Reference Law   CLOSED / APPROVED / GCR-CLARIFIED
  B2   Authentication / Organization / Authorization   NEXT / DESIGN ONLY
  B3   Controlled Information + Artifact relational core NOT STARTED
  B4   Approval + CI-owned Rendition/Release + Distribution NOT STARTED
  B5   Documentary Context / Records + Artifact closure NOT STARTED
  B6   Audit / Interchange / Cross-owner Atomicity      NOT STARTED
R10-C  Artifact / Records Physical Integrity            NOT STARTED
R10-D  Durable Async / Projections / External Effects   NOT STARTED
R10-E  Canonical Access / API / Frontend Journeys       NOT STARTED
R10-F  Historical Migration / Cutover / Final Deletion  NOT STARTED
```

B-blocks are design work packages, not ownership reassignment. In particular, Rendition/Release/effectivity remain Controlled Information-owned even though their relational-state design is sequenced with Approval/Distribution in B4.

Closure order:

```text
R10-A → B1 → B2 → B3 → B4 → B5 → B6 → R10-C → R10-D → R10-E → R10-F
```

Product implementation remains blocked until the integrated technical design and its operator/adversarial gates are complete.

---

# 2. R10-A — promoted ownership topology, GCR-refined

## 2.1 Business bounded contexts — exactly 8

### Authentication

Authentication owns product-facing authentication semantics:

```text
ProviderSubjectBinding
opaque MetalDocs application Session
application-session lifecycle / revocation
authentication-assurance facts
fresh-auth / reauthentication evidence
provider anti-corruption contract
```

Keycloak is the V1 Authentication provider and owns credential mechanisms:

```text
credential storage
password policy
provider account activation / enablement
provider lockout / brute-force protection
password recovery
MFA / passkeys
upstream OIDC / SAML / LDAP / AD federation
provider authentication session
provider-hosted authentication journeys
```

Stable provider identity is `issuer + subject`. Email, username and display name are attributes, not technical identity.

The published Authentication/provider result is an enumerated anti-corruption contract. It may expose facts such as:

```text
issuer
subject
authenticated_at
auth_time
acr?
amr?
```

It does **not** expose a generic claims map to Authorization/domain owners. Provider roles, realm/client roles, groups, organizations, permissions and arbitrary claim-to-permission mappings are forbidden as canonical MetalDocs Authorization inputs. There is no provider-role mapping table or claim-to-MetalDocs-permission bridge V1.

Keycloak Organizations or equivalent provider organization machinery may later be used only as AuthN routing/federation projection of MetalDocs Tenant state; it is never Organization or Authorization authority.

Candidate provider topology is one Keycloak realm per environment/application trust domain, not realm-per-Tenant. A real tenant-specific authentication-policy requirement that cannot be expressed within that topology is a reopen trigger.

Authentication owns no Organization membership or product grants.

### Organization

Owns organizational identity and Tenant lifecycle:

```text
Tenant
Area
User
Group
GroupMembership
Tenant settings/configuration facts
Tenant lifecycle: ACTIVE | SUSPENDED | ERASED
TenantDeletionRequest
TenantErasureRecord
erasure tombstone / erased-tenant reconciliation facts
```

`tenant.settings.manage` governs Tenant-family settings/configuration; no standalone `TenantSettings` bounded context or aggregate is implied by the permission.

There is **no mandatory Tenant DEK/key-custody fact family in V1**. If a later named application-layer Target Data family requires tenant-key lifecycle semantics, that decision re-enters through the GCR-R4 reopen trigger and must prove the protected data, erasure semantics and backup/restore posture before adding key-custody facts.

Records Governance supplies retention/hold blocker facts; it does not own Tenant state.

### Authorization

Owns access-grant meaning and canonical evaluation:

```text
Permission
Role
RoleAssignment
subject: User | Group
scope: Tenant | Area
grant / revocation evidence
canonical grant evaluation
composable authorization/filter contract shape
```

Authorization owns the composition/evaluation contract, not domain relationship meaning. Each semantic owner owns its resource/case predicates. No role bypass, generic ACL/ReBAC graph, provider permission engine, provider-role bridge or deny engine V1.

Relationship-predicate ownership:

| Relationship meaning | Semantic owner |
|---|---|
| Tenant/Area/User/Group membership and organizational scope | Organization |
| Document/Revision ownership-responsibility, Area and lifecycle | Controlled Information |
| Dossier scope, Evidence primary-Dossier scope and contextual links | Documentary Context |
| Approval participant qualification, snapshot and SoD | Approval, consuming Organization/Authorization facts |
| Hold-to-subject materialization and disposition blockers | Records Governance |
| Distribution obligation/audience/acknowledgement | Distribution; never grants access |
| Imported/exported target-object relationships | target object's semantic owner |

Search/export/timeline/API surfaces consume the composed canonical result and may not re-derive policy independently.

### Controlled Information

Owns the single governed Document/Revision lifecycle:

```text
DocumentType
Document
DocumentRevision
Document owner / responsibility relationship
WorkingContent
WorkingSnapshot technical checkpoint
EditorSession
RevisionSubmission
numbering / revision labels / reason-for-change
DocumentType approval/representation/review configuration
Template designation / TemplateUse / TemplateSpec
EditorialComment
PeriodicReview policy + PeriodicReviewRecord
Rendition business identity/provenance for exact Submission
OfficialRepresentationPolicy
ReleasePlan / ReleaseRecord / effectivity
Tenant Dictionary values
System Value Catalog descriptors/resolution contract
value snapshots bound to governed revision lifecycle
```

The `Document owner / responsibility relationship` is a durable Controlled Information fact family because the frozen permission catalog includes `document.owner.manage`. R10-A intentionally does **not** decide its participant type, cardinality or physical representation. Those details belong to R10-B unless implementation evidence exposes a genuinely semantic ambiguity requiring a new decision loop. The relationship never bypasses capability-based Authorization.

Tenant Dictionary and System Value Catalog remain distinct internal fact classes:

```text
Tenant Dictionary     = tenant-managed mutable values
System Value Catalog  = product-owned bounded descriptors/resolvers
```

No standalone Dictionary owner exists in V1. The frozen launch consumer is the governed authoring/revision lifecycle; a real second business consumer or independent lifecycle is the reopen trigger.

DocumentType stores the frozen retention-rule value directly; R10-A does not introduce a `RetentionPolicy` entity.

### Approval

Owns:

```text
ApprovalPolicy(version)
ordered ApprovalStep configuration
actor rule / ANY|ALL / requires_reauthentication / due_in_days
ApprovalInstance bound to exact RevisionSubmission
activated participant snapshot
ApprovalDecision
attestation evidence: actor / Step / policy version / decision / server time / AuthN assurance
return-for-changes rationale
withdraw / cancel / reassign / oversight
strict SoD
```

Approval never owns Document effectivity. Human-readable manifestations of Approval facts are Controlled Information Renditions of Approval authority, not a second Approval record. Reauthentication consumes Authentication-owned assurance/fresh-auth facts; Approval never challenges local passwords itself.

### Documentary Context

Owns:

```text
EvidenceType
Evidence
Evidence lifecycle DRAFT → CAPTURED → VOIDED
external-world cancellation fact distinct from VOIDED
Evidence naming / allowed-format / retention-rule / retention-anchor configuration
DossierType
Dossier
Dossier stable key / scope
Dossier lifecycle ACTIVE ↔ ARCHIVED
creation provenance
ExternalReference
Dossier ↔ Document contextual link
Evidence primary-Dossier relationship
Evidence secondary context links
Evidence → exactly one primary Artifact relationship
```

Dossier/context links never grant access. CAPTURED Evidence uses exactly one immutable primary Dossier and reuses its scope. EvidenceType stores the frozen retention-rule value directly and selects `CAPTURED_AT | OCCURRED_AT`; no standalone `RetentionPolicy` entity exists.

### Records Governance

Owns preservation/disposition meaning:

```text
retention-rule vocabulary/meaning
RetentionBinding + snapped rule/anchor
retention clocks / eligibility
RetentionExtension
LegalHold
materialized held-subject relationship
prospective hold materialization
Disposition authorization / eligibility / completion
DispositionRecord
retention/hold blocker facts used by Tenant erasure
```

DocumentType/EvidenceType own their configured rule values; Records Governance owns the rule meaning and resulting retention-subject lifecycle. R10-A introduces no separately versioned `RetentionPolicy` aggregate.

### Distribution

Owns:

```text
released-document distribution obligation
audience snapshot / historical denominator
AcknowledgementRecord
coverage / completion semantics
```

Distribution never grants access. Notification delivery/read state never substitutes for acknowledgement.

---

## 2.2 Supporting semantic owners — exactly 3

### Artifact

Owns exact-byte and physical-content truth:

```text
Artifact identity
canonical SHA-256
size
closed ContentFormat catalog
media type
technical provenance
staging / validation / confirmation facts
managed physical-location facts
relocation verification / cutover facts
restore byte-integrity facts
```

Controlled Information owns the Revision/WorkingContent → primary Artifact relationship. Documentary Context owns the Evidence → primary Artifact relationship. Artifact owns the referenced byte identity and physical truth, not either business lifecycle.

No confirmed orphan Artifact exists. Storage providers remain replaceable mechanisms.

Production confirmation adds one promoted GCR safety property: untrusted bytes cannot become `CONFIRMED Artifact` without successful malware inspection. Scanner selection and physical validation ordering belong R10-C; Artifact owns the staging/validation/confirmation fact, not the scanner mechanism.

### Audit

Owns:

```text
append-only AuditEvent timeline
tamper-evidence / chain meaning
audit query/export semantics
Audit Trail separate retention regime
```

Critical governed mutation must be able to append durable Audit intent/event in the same local commit through a published transactionally composable seam. Exact transaction/storage mechanism belongs to R10-B/R10-D. Domain records remain the authority for their own facts.

GCR-R4 adds a B6 proof obligation: the immutable Audit state permitted to survive Tenant erasure must be a PII-minimized/non-PII skeleton. Human-readable/user enrichment must resolve through separately erasable state or read/projection enrichment. If B6 proves a real immutable Target Data family must remain stored yet become unintelligible after lawful erasure, the DEK/key-custody decision reopens before crypto-erasure machinery is introduced.

### Interchange

Owns only enumerated transfer-boundary process truth:

```text
Historical Migration batch / plan / dry-run / per-item outcome / reconciliation
Tenant Portability Export package process
Governed Subject Export package process
External Repository IMPORT_COPY / PUBLISH_COPY attempt/process truth
transfer/package/reconciliation identity
transfer-level source provenance
```

Object-level provenance belongs to the semantic owner of the object. Historical lifecycle/approval/effectivity evidence attached to an imported Revision is Controlled Information imported governance evidence, never fabricated native ApprovalDecision/ApprovalInstance/ReleaseRecord. Interchange owns the transfer process, not imported business truth.

Interchange is not an ESB, generic ETL, workflow, connector platform or external master-data authority.

---

## 2.3 Attributed support / projections / mechanisms

### Notifications — attributed support

Target placement:

```text
internal/support/notifications
```

Owns durable delivery/attempt/inbox/read state only. Producers resolve business meaning and recipients before handing off delivery intent. Notifications owns no Approval, Distribution acknowledgement, Authorization or other business truth.

### Search — rebuildable projection

Target placement:

```text
internal/projections/search
```

Search owns no canonical business truth. Projection membership never grants access; queries reapply canonical Authorization plus owner-supplied relationship predicates.

### Commodity mechanisms

Not semantic owners:

```text
jobs / schedulers
outbox / queue / leases / DLQ
workers
Keycloak / IdP provider client and adapter
storage provider clients
malware scanner client / inspection mechanism
rendering provider clients
external repository adapters
RLS
HTTP / OpenAPI / codegen
cache
rate limiting
observability
backup image transport
PlatformOperator / SystemPrincipal execution machinery
```

No mandatory V1 KEK/wrap-unwrap/Tenant-DEK mechanism exists. Such machinery requires a named protected Target Data family and a reopened decision.

PlatformOperator/SystemPrincipal remain outside tenant RBAC and gain no implicit tenant-content authority.

---

# 3. R10-A dependency and coordination rules

`internal/composition` is the outer application layer for concrete cross-owner use cases. Composition may coordinate owners; it owns no durable business meaning and no semantic owner imports it.

R10-A requires transactionally composable published application seams wherever frozen semantics require one local MetalDocs DB commit. R10-B decides the concrete UnitOfWork/Tx/schema mechanism.

Material seams:

1. **CI ↔ Approval** — exact Submission references and Approval reads are composition/read-contract mediated; no mutual package authority.
2. **Local transactional composition** — published application seams must permit one local MetalDocs DB transaction across exactly the owners required by a frozen atomicity invariant.
3. **Audit append** — Audit publishes a transactionally composable append seam; producers never own Audit storage/meaning.
4. **Artifact confirmation** — caller supplies an opaque semantic-owner reference; Artifact does not import CI/DC. R10-B supplies the structural no-orphan backstop; R10-C supplies physical validation/inspection proof.
5. **Records prospective hold materialization** — Records consumes published CI/DC subject facts/events/read seams; it does not acquire underlying lifecycle authority.
6. **Historical Migration** — target owners expose narrow privileged migration-grade application seams; Interchange calls owners, never the reverse.
7. **Notifications** — producers resolve recipients/business meaning before delivery intent; Notifications does not invent policy.
8. **Authorization filtering** — Authorization composes owner-supplied predicates; Search/export/timeline/API consumers do not rederive visibility.
9. **Tenant erasure/restore** — Organization owns Tenant/tombstone lifecycle, Records Governance owns blockers, Authentication owns application-session revocation, Artifact owns byte truth, Audit owns the surviving permitted non-PII skeleton; composition owns none of them.
10. **Authentication provider** — Keycloak/provider owns credential mechanics; MetalDocs owns User, provider binding, app Session and assurance facts. No operation assumes atomic commit across provider persistence and the MetalDocs product DB.

The target package/import DAG must be acyclic. Semantic dependency does not require direct Go import; interface inversion, references, events and composition may preserve ownership while avoiding cycles.

---

# 4. Target filesystem classification

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
    <concrete cross-owner application use cases only>

  platform/
    <commodity db/http/async/observability/security/provider mechanics>
```

Within a semantic owner, use `domain/`, `application/`, `infrastructure/`, `delivery/`, `api/`, `public/` only when real consumers justify them. `public/` is not mandatory.

Provider placement:

```text
Keycloak / IdP adapter                 → Authentication infrastructure
Local / AWS S3 / compatible adapters   → Artifact infrastructure
malware scanner adapter                → Artifact/platform validation mechanism
EigenPal / rendering provider adapters → Controlled Information infrastructure/execution
SharePoint / external repo adapters    → Interchange infrastructure
```

The durable storage entitlement is `ManagedArtifactStore` port + conformance, not the provider list. No self-hosted production provider is frozen without a real deployment consumer.

---

# 5. Legacy backend disposition fixed by R10-A + GCR

| Current module | Promoted target disposition |
|---|---|
| `approval` | converge to Approval V1 |
| `audit` | retain Audit semantic owner; redesign durable skeleton later |
| `auth` | converge to Authentication provider-binding/app-session/assurance; local credential machinery becomes migration/deletion evidence |
| `controlleddocuments` | delete as target BC; identity/numbering responsibilities → Controlled Information |
| `distribution` | retain Distribution semantic owner |
| `documents` | delete legacy BC; responsibilities → Controlled Information |
| `iam` | delete/split → Organization + Authorization |
| `jobs` | delete as BC; owner-attributed work + platform async/composition |
| `notifications` | reclassify → `support/notifications` |
| `render` | dismantle; Rendition/value semantics → Controlled Information, providers → infrastructure |
| `search` | reclassify → `projections/search` |
| `security` | delete as BC; product-facing AuthN facts → Authentication, commodity security → platform; legacy DEK/KEK machinery has no V1 target entitlement |
| `taxonomy` | dismantle; Area → Organization, DocumentType/classification → Controlled Information, GovernanceClass deleted |
| `templates` | delete parallel lifecycle; template role/use → Controlled Information |
| `tokens` | delete standalone owner; Tenant Dictionary/System Value Catalog → Controlled Information |

No current module survives merely by inertia. Exact table/API/frontend cutover mechanics remain later R10 work.

---

# 6. Surface classification fixed by R10-A

The single OpenAPI + generated-owner-surface structural pattern may survive, but target ownership is:

```text
auth                         → Authentication provider binding / app Session / assurance
iam                          → Organization + Authorization
documents / controlled-docs  → Controlled Information
templates / tokens           → Controlled Information
taxonomy Area                → Organization
taxonomy document typing     → Controlled Information
approval                     → Approval
distribution                 → Distribution
audit                        → Audit
search                       → Search projection
notifications                → Notifications support
security                     → retire/split across Authentication/platform
documentary context          → new owner surface
records governance           → new owner surface
interchange                  → new owner surface
configuration/health/obs     → platform
```

There is no generic `/artifacts` business API merely because Artifact is a semantic supporting owner; confirmed Artifact must be attached to a governed owner. Exact paths, operationIds, DTOs and frontend journeys are R10-E.

Keycloak-hosted/themed login/recovery/MFA journeys replace custom MetalDocs credential UX where appropriate; R10-E must not rebuild credential administration journeys through provider admin APIs.

---

# 7. R10-A closure proof, GCR amendment and reopen record

R10-A originally closed after independent topology review, Method adjudication, cold/global coherence review, final completeness correction and mechanical fact/permission sweep.

Original final closure:

```text
BLOCKER                    = 0
remaining topology defect = 0
duplicate owners           = 0
invented fact families     = 0
RetentionPolicy entity     = ABSENT / PASS
```

The later 2026-08-17 Global Coherence Review did **not** reopen topology. It produced a bounded amendment of Authentication/provider facts and removed the unsupported mandatory Organization key-custody fact family.

GCR evidence chain:

1. `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-fable-review-request.md`;
2. `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-independent-fable-review.md`, verdict `APPROVE GCR MINIMAL REOPEN SET WITH MATERIAL FIXES`;
3. `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-adjudicated-corrected-target.md`;
4. `docs/superpowers/analysis/2026-08-17-global-coherence-minimal-reopen-corrected-target-fable-delta-review.md`, verdict `APPROVE GCR ADJUDICATED CORRECTED TARGET`.

Final GCR delta:

```text
BLOCKER = 0
MAJOR   = 0
prior findings closed = 11/11
new material contradiction = NONE
fifth material local maximum = NONE
```

R10-A topology remains exactly 8+3. Review artifacts remain evidence, not parallel authority; this page contains the promoted amended outcome.

---

# 8. R10-A invariants and reopen triggers

Promoted invariants:

1. every frozen durable/business fact family maps to exactly one semantic owner;
2. mechanisms/providers never become business authority by convenience;
3. no target bounded context is retained merely because a current module exists;
4. the 8 business bounded contexts + 3 supporting semantic owners are the V1 ownership set; adding a twelfth semantic owner is material and must return to the Method decision loop;
5. composition coordinates but owns no durable meaning;
6. canonical Authorization evaluation composes domain-owned relationship predicates rather than centralizing domain semantics;
7. no standalone Dictionary or `RetentionPolicy` entity exists without a real independent consumer/lifecycle;
8. Document owner/responsibility semantics belong to Controlled Information, while its concrete representation remains an R10-B question unless product semantics force a separate decision;
9. Notifications owns attributed delivery/read state but no business truth; Search is rebuildable projection only;
10. Keycloak/provider is Authentication mechanism/credential authority only; Organization and Authorization remain MetalDocs authorities;
11. no mandatory Tenant DEK/key-custody fact family exists without a named application-layer Target Data family requiring it.

R10-A reopens only on material evidence such as:

- a frozen/business fact that cannot be owned coherently by the promoted set;
- a real second consumer or independent lifecycle that invalidates the Dictionary collapse;
- a relationship model that materially exceeds frozen Tenant/Area authorization semantics;
- a genuine independent Rendition semantic consumer;
- Evidence/Dossier lifecycle evidence requiring a different boundary;
- an external-transfer requirement proving Interchange must split or a different owner must exist;
- an indivisible multi-file requirement that materially changes Artifact semantics;
- implementation evidence showing the acyclic/transactionally-composable ownership assumptions cannot preserve a frozen invariant;
- a material provider/authentication requirement showing one-realm-per-trust-domain cannot satisfy real tenant policy without provider Organization becoming product authority;
- a named immutable Target Data family proving tenant key-custody lifecycle is essential.

Package naming preference, current schema inconvenience, provider capability or implementation convenience are not reopen evidence.

---

# 9. R10-B1 — promoted relational substrate, tenancy and reference law

R10-B1 is **CLOSED / APPROVED**. The GCR changed no structural B1 law; it only clarified product-state DB scope and forbade cross-provider-DB atomicity assumptions.

## 9.1 PostgreSQL topology and identity

```text
one MetalDocs product-state PostgreSQL database
canonical target product-state schema = metaldocs
schema-per-Tenant = NO
schema-per-bounded-context = NO
```

Provider-owned products such as Keycloak retain separate persistence authority, migrations, credentials and lifecycle. Physical co-location on one PostgreSQL server/cluster does not merge logical persistence authority.

**No MetalDocs invariant may depend on cross-database atomicity between the MetalDocs product-state database and any provider-owned database.** No XA/2PC/distributed transaction is introduced to simulate such atomicity.

PostgreSQL namespace is mechanism, never semantic authority. Target product state does not use `public` as a second business-state namespace.

For a durable tenant-owned entity:

```text
tenant_id UUID NOT NULL
id        UUID NOT NULL
PRIMARY KEY (tenant_id, id)
```

`id` is opaque technical identity. No second global `UNIQUE(id)` is required absent a real consumer. Business/provider/external identifiers — Document code, REV label, Dossier key, Artifact hash, provider subject, external ID, provider key/URL — never become technical PKs by convenience.

Tenant is the root; genuinely global/product facts may lack `tenant_id` when their owner semantics require it. B2+ must not mechanically tenant-key every table for uniformity. Provider-subject binding tenant dimension/cardinality is a B2 decision and must be explicit; global `UNIQUE(issuer, subject)` is not assumed.

## 9.2 Same-tenant references and FK action law

Where relational existence is an invariant, tenant-owned references use tenant-qualified identity:

```text
FOREIGN KEY (tenant_id, target_id)
REFERENCES target_table(tenant_id, id)
```

A cross-owner FK proves existence, same Tenant and identity only; it never transfers lifecycle/business authority.

Across semantic owners the FK action set is closed:

```text
ON DELETE RESTRICT   = allowed
ON DELETE NO ACTION  = allowed
ON UPDATE RESTRICT   = allowed
ON UPDATE NO ACTION  = allowed

CASCADE               = forbidden
SET NULL               = forbidden
SET DEFAULT            = forbidden
any other action       = forbidden unless this law is deliberately reopened
```

One owner's FK can never delete or mutate another owner's durable state. Cross-owner deletion/disposition/erasure is explicit coordinated behavior through the owning authorities.

Within one owner, cascade remains non-default and is allowed only for strictly subordinate state with no independent historical meaning.

Generic polymorphic business reference registries (`resource_type/resource_id`, generic Record/Object registries) are rejected. Typed relationships belong to the owner of the relationship. Audit generic resource attribution remains legal because Audit is explicitly non-authoritative for resource state.

## 9.3 Persistence primitives

Defaults:

```text
technical IDs     = UUID
business instants = TIMESTAMPTZ
canonical SHA-256 = BYTEA with octet_length(hash)=32
frozen vocabulary = TEXT + CHECK by default
real unknown      = NULL
```

Do not use empty strings, zero UUIDs, zero numbers or fabricated `UNKNOWN` values to avoid legitimate nullability. Historical Migration preserves unknown as unknown.

`JSONB` is allowed for bounded whole snapshots or genuinely variable provider-neutral provenance whose atomic meaning is the whole value; it is not an escape hatch for unmodeled business state. PostgreSQL ENUM is not the default.

Historical snapshots preserve what was true at the governed moment and are never silently rewritten by later mutation of their source.

## 9.4 Orthogonal persistence classification

Every persisted fact/table declares two independent dimensions.

Semantic persistence class:

```text
SEMANTIC AUTHORITY
ATTRIBUTED SUPPORT
DURABLE MECHANISM
EPHEMERAL MECHANISM
REBUILDABLE PROJECTION
```

Mutation law:

```text
MUTABLE
IMMUTABLE / APPEND-ONLY
TERMINAL / TOMBSTONED
REBUILDABLE
or an explicit constrained state machine
```

The classes describe state; they do not create semantic owners. Examples:

```text
RevisionSubmission      = SEMANTIC AUTHORITY + IMMUTABLE
ApprovalDecision        = SEMANTIC AUTHORITY + IMMUTABLE
AuditEvent              = SEMANTIC AUTHORITY + APPEND-ONLY
Notification inbox/read = ATTRIBUTED SUPPORT + MUTABLE
async/outbox intent     = DURABLE MECHANISM + constrained operational state
job lease               = EPHEMERAL MECHANISM
Search index            = REBUILDABLE PROJECTION + REBUILDABLE
```

B2–B6 must classify every table/fact family they close and choose a falsifiable enforcement strategy proportionate to immutable/terminal claims. “The application normally does not update it” is not sufficient proof of immutability.

## 9.5 Tenant isolation law

Tenant-owned `SEMANTIC AUTHORITY` state receives the full isolation stack:

```text
application/repository tenant predicate
+ same-tenant relational keys/FKs
+ ENABLE ROW LEVEL SECURITY
+ FORCE ROW LEVEL SECURITY
+ missing tenant context = FAIL CLOSED
```

Tenant-owned `ATTRIBUTED SUPPORT` receives the same default stack unless its owning block proves a narrower representation that preserves the identical tenant-isolation claim. R10-D must explicitly exercise or decline this clause when closing Notifications persistence; it may not become an unexamined escape hatch.

RLS is Tenant isolation only. It must not encode roles, Area, Dossier links, Approval participant logic, Document ownership or other canonical Authorization predicates.

`DURABLE MECHANISM`, `EPHEMERAL MECHANISM` and `REBUILDABLE PROJECTION` isolation posture is chosen by their owning later block, but no mechanism may use a globally readable surface to expose arbitrary tenant business content.

## 9.6 Durable async intent and global claim surface

When a tenant business mutation requires future async work, the required durable intent is inserted in the same tenant-scoped transaction but remains `DURABLE MECHANISM`, not business authority.

A globally claimable mechanism surface may expose only routing/mechanism facts such as:

```text
intent identity
tenant_id
intent kind / due time
lease/claim state
opaque target reference
```

It must not expose arbitrary tenant business content merely to make dispatch convenient.

Execution pattern:

```text
global claim of routing metadata
→ obtain tenant_id + opaque target
→ BEGIN ordinary tenant-scoped transaction
→ seed Tenant/system execution context
→ load canonical business content under normal isolation/AuthZ
→ execute owner/application use case
→ COMMIT/ROLLBACK
```

R10-D owns claim/retry/DLQ/external-effect execution and must preserve both global claimability and tenant-content isolation. It may not solve dispatch by granting a worker implicit all-tenant content visibility.

## 9.7 Database roles and platform maintenance

Ordinary serving runtime DML — API, workers and normal jobs — uses a role that is:

```text
NOSUPERUSER
NOBYPASSRLS
not owner of protected product tables
```

DDL/object ownership is separate. No database role per bounded context is introduced without a real deployment/trust boundary.

True cross-tenant migration/backfill/restore is a distinct non-serving maintenance trust surface. It may use per-Tenant iteration or a narrowly scoped maintenance principal with the minimum technical bypass required, but that principal:

```text
must never be an ordinary serving role
must never be reachable from a request path
must not imply product Authorization
must not become a SystemPrincipal content bypass
```

Concrete maintenance credential/process/rotation/cutover mechanics belong to R10-F / implementation / operations.

## 9.8 Transaction and isolation law

Ordinary tenant-owned mutations execute in explicit local MetalDocs PostgreSQL transactions.

Single-owner mutation: the owner application service owns the boundary.

Cross-owner atomic use case inside the MetalDocs product-state DB:

```text
composition opens one PostgreSQL transaction
→ invokes owner-specific published application seams
→ every participant uses the same transaction
→ one COMMIT or one ROLLBACK
```

No semantic owner hides an independent nested commit inside a frozen atomic operation, and no owner imports another owner's repository to obtain atomicity. Concrete UnitOfWork/Tx interface shape is deferred to the relevant B-blocks.

Provider-side work (Keycloak, object-store external effects, repositories, etc.) is never part of this local atomicity claim; it uses durable intent/idempotency/reconciliation in the owning later stage.

Default isolation is `READ COMMITTED`. `SERIALIZABLE` is not a global correctness substitute; use the narrowest sufficient invariant mechanism (`UNIQUE`, partial UNIQUE, CHECK, FK, CAS/working_version, `SELECT ... FOR UPDATE`, atomic UPDATE) and raise isolation only for a demonstrated failure class.

## 9.9 Same-commit Audit and durable intent

Where frozen authority requires Audit evidence, a critical governed mutation cannot report success unless its Audit append is durable in the same MetalDocs DB commit.

Where the mutation necessarily creates future async/provider work, its required durable intent is inserted in that same commit.

```text
BEGIN
  authoritative business/support facts
  required Audit append
  required durable mechanism intent
COMMIT
```

or all local facts roll back. External/provider effects happen after commit and are reconciled; no cross-provider atomicity is implied.

## 9.10 Background work discovery

Only two shapes are lawful for discovering tenant work under fail-closed semantic/support isolation:

1. enumerate Tenants from a root/platform surface, then seed each Tenant and query due work under ordinary isolation;
2. consume a tenant-written platform routing/due intent that exposes only bounded routing metadata, then re-enter tenant-scoped execution.

A scheduler/job does not receive a third option that fails open RLS on tenant semantic/support tables.

## 9.11 R10-B design-package assignment

```text
B2 → Authentication + Organization + Authorization relational state
B3 → Artifact relational core + Controlled Information + WorkingContent + Submission
B4 → Approval + Controlled Information-owned Rendition/Release/effectivity relational state + Distribution
B5 → Documentary Context + Records Governance + Artifact second-consumer/no-confirmed-orphan closure
B6 → Audit relational state + Interchange batch/plan/outcome state + cross-owner tx matrix + imported-history/global DB coherence

R10-D → Notifications attributed-support persistence details + Search projection persistence + async mechanism persistence/execution
```

This map sequences design work; it does not move R10-A ownership.

Physical Artifact storage/relocation remains R10-C; worker/retry/external effects remain R10-D; API/frontend remain R10-E; migration/backfill/cutover execution remains R10-F.

## 9.12 Target namespace vs migration provenance

`metaldocs` is the final target product-state namespace even though the current system already contains legacy `metaldocs.*` tables. During transition, namespace occupancy is not proof that a table is target state.

R10-F must carry an explicit target-vs-legacy manifest/choreography (staging/rename/table-level mapping as chosen there). Current namespace inconvenience does not reopen the B1 target topology.

## 9.13 Closure evidence and surviving proof obligations

R10-B1 evidence chain:

1. candidate — `docs/superpowers/analysis/2026-08-17-r10-b1-relational-substrate-fable-review-request.md` @ `a3bb4ac8`;
2. independent cold review — `docs/superpowers/analysis/2026-08-17-r10-b1-independent-fable-review.md` @ `b38f598b`, verdict `APPROVE R10-B1 WITH MATERIAL FIXES`;
3. Method adjudication/corrected target — `docs/superpowers/analysis/2026-08-17-r10-b1-fable-adjudication-corrected-target.md` @ `92cba574`;
4. bounded delta review — `docs/superpowers/analysis/2026-08-17-r10-b1-corrected-target-fable-delta-review.md` @ `f0273b58`, verdict `APPROVE R10-B1 CORRECTED TARGET`.

The later GCR left B1 structural law unchanged and added only the product-state DB wording and no-cross-provider-DB-atomicity restriction.

The following claims must remain falsifiable through later design/implementation:

- same-Tenant composite-reference proof for every protected relationship;
- census proving every cross-owner FK uses only `RESTRICT`/`NO ACTION` and can neither delete nor mutate another owner's durable state;
- fail-closed RLS negative proof under the actual serving-class non-owner/NOBYPASSRLS role;
- proof that ordinary serving pools actually use the non-bypass role;
- rollback proof that required Audit and durable-intent rows share the authoritative MetalDocs mutation commit;
- proof that every closed B-block assigns both semantic persistence class and mutation law;
- proof that no promoted invariant requires atomicity across a provider DB and the MetalDocs product-state DB;
- R10-D proof that global dispatch surfaces expose routing metadata, not arbitrary tenant content;
- R10-D closure evidence explicitly exercising or declining the narrower-representation clause for Notifications while preserving the same tenant-isolation claim;
- R10-F proof that any maintenance bypass is non-serving, least-privilege and unreachable from request paths.

## 9.14 R10-B1 reopen triggers

Reopen B1 only on material evidence such as:

- a real tenant-owned relationship that cannot preserve same-Tenant integrity under the composite-key/reference law;
- a real global technical-identity consumer requiring global `UNIQUE(id)` rather than tenant-qualified identity;
- evidence that the one-product-state-schema/local-transaction premise cannot preserve a frozen invariant;
- an async/external-effect requirement that cannot remain globally claimable without exposing tenant content despite the routing-only seam;
- a serving operation that genuinely cannot function under fail-closed Tenant context and cannot be expressed by per-Tenant iteration or routing intent;
- implementation evidence that cross-owner `RESTRICT`/`NO ACTION` makes a frozen lifecycle impossible rather than merely explicit;
- a trust/deployment boundary that materially justifies schema/role separation beyond the promoted law.

Current schema inconvenience, migration cost, provider capability, package naming or hypothetical scale are not reopen evidence.

---

# 10. Exact next step — R10-B2

Open **R10-B2 — Authentication / Organization / Authorization State** in design-only mode.

B2 must derive the minimum target persistent state for the three promoted owners under B1's substrate law. The GCR removes local credential/key-custody machinery from scope and adds provider-binding/reconciliation obligations.

At minimum B2 must decide:

```text
Authentication provider-subject binding representation
  stable issuer + subject boundary
  explicit Tenant dimension in uniqueness law
  whether one User may bind one or multiple provider subjects
Authentication ↔ Organization User binding integrity
opaque MetalDocs application Session representation and lifecycle
fresh-auth / authentication-assurance representation
structural provider anti-corruption contract proof:
  no provider role/group/org/permission consumption
  no generic claims map into Authorization/domain owners
  no provider-role mapping / claim-to-permission bridge
provider-binding/provisioning lifecycle + idempotent reconciliation for:
  User exists / provider subject absent
  provider subject exists / binding absent
  binding exists / provider subject removed or disabled
  duplicate issuer+subject attempt
  provider unavailable
  retry after uncertain provider response
provider-side disable vs already-live MetalDocs Session posture

Tenant / Area / User / Group / GroupMembership table ownership and lifecycle
Tenant settings/configuration persistence without inventing a new authority
Tenant lifecycle ACTIVE/SUSPENDED/ERASED durable representation
TenantDeletionRequest / TenantErasureRecord / tombstone + restore-reconciliation state

Permission / Role / RoleAssignment representation
User|Group principal references
Tenant|Area typed scope representation
role/grant/revocation evidence
canonical grant-evaluation read model needed by later owners

same-Tenant FK and RLS application under B1
semantic persistence + mutation-law classification for every B2 fact family
transaction boundaries for membership/grant/lifecycle mutations
required Audit/durable-intent insertion points
```

B2 must preserve these boundaries:

- Authentication owns provider binding/app Session/assurance, not credentials, Organization membership or grants;
- Organization owns Tenant/Area/User/Group/Tenant lifecycle, not credentials, permissions or mandatory key-custody state;
- Authorization owns Permission/Role/RoleAssignment/grant evaluation, not domain relationship predicates;
- RLS remains Tenant isolation only;
- PlatformOperator/SystemPrincipal remain outside tenant RBAC with no implicit tenant-content authority;
- no `tenant_owner` bypass; flat groups only V1; exactly five frozen tenant roles; RoleAssignment subject = User|Group and scope = Tenant|Area; additive/default-deny grants;
- Keycloak is the V1 AuthN provider but its roles/groups/organizations/permissions never become canonical product authority;
- no OpenFGA/SpiceDB, generic ACL/ReBAC graph, deny engine, nested groups or speculative enterprise identity machinery without a real trigger;
- no password/MFA/credential-storage target tables;
- no mandatory Tenant DEK/KEK/wrap-unwrap infrastructure;
- no XA/2PC with provider persistence.

### Named successor obligations from GCR

**B6 — Audit:** classify the immutable post-erasure Audit skeleton field-by-field. Prove surviving fields are PII-minimized/non-PII; user-readable enrichment is separately erasable/projection-only. If a real immutable Target Data family must remain stored yet become unintelligible, reopen GCR-R4 before adding crypto-erasure machinery.

**R10-C — Artifact/storage/content safety:** define `ManagedArtifactStore` conformance and its execution environment; choose the S3 client deliberately; decide scanner vs parser/validator ordering; prove production malware inspection cannot be bypassed; make deployment-profile declaration single-sourced so inspection-disabled dev/test cannot present as production; define staged-byte cleanup; record deletion/replacement condition for any transitional MinIO/dev-CI endpoint.

**R10-D — provider/external effects:** implement provider provisioning/retry/reconciliation mechanics only after B2 closes semantic lifecycle; no distributed transaction.

**R10-E — authentication UX:** use provider-hosted/themed login, recovery and MFA journeys where appropriate; do not rebuild credential UX through Keycloak admin APIs.

**R10-F — migration/cutover:** remove legacy local credential and DEK/KEK machinery according to accepted target mappings; provider persistence remains separate authority.

Current IAM/auth/security tables, schema and runtime are evidence only. No schema/code implementation is authorized.
