# R10 Technical Architecture — Active Stage Authority

> **Status:** ACTIVE — **R10-A CLOSED / APPROVED; R10-B NEXT / DESIGN ONLY**
> **Promoted:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** `docs/engineering/standards/root-cause-global-maximum-method.md` — DevelopmentConexus Engineering Method v1.0.0 mirror
> **Frozen product/domain authority:** `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` for R3–R9.5
> **Program authority:** `wiki/architecture/cohesive-platform-redesign.md`
> **Implementation gate:** **CLOSED — no product implementation authorized.**

This page is the durable stage authority for R10. The R3–R9.5 ledger remains binding for frozen product/domain semantics. Its historical statements that `R10 = NEXT` are superseded for current-stage routing by this page and the current handoff; its frozen semantic decisions are not superseded.

---

## 1. R10 decomposition and closure order

R10 is one integrated technical-design stage decomposed by failure class:

```text
R10-A  Ownership Topology & Dependency DAG          CLOSED / APPROVED
R10-B  Transactional Domain State & DB Invariants  NEXT
R10-C  Artifact / Records Physical Integrity        NOT STARTED
R10-D  Durable Async / Projections / External Effects NOT STARTED
R10-E  Canonical Access / API / Frontend Journeys   NOT STARTED
R10-F  Historical Migration / Cutover / Final Deletion NOT STARTED
```

Closure order:

```text
R10-A → R10-B → R10-C → R10-D → R10-E → R10-F
```

R10-E analysis may begin only when useful after ownership is stable, but it cannot close before the invariants/mechanisms it exposes are decided. Product implementation remains blocked until the integrated technical design and its operator/adversarial gates are complete.

---

# 2. R10-A — promoted ownership topology

## 2.1 Business bounded contexts — exactly 8

### Authentication

Owns local credential/session identity facts:

```text
credential / identity binding
activation
opaque session
lockout / revocation
fresh-auth / reauthentication assurance
```

External IdP remains an adapter seam. Authentication owns no grants, organization membership or Tenant lifecycle.

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
tenant key-custody lifecycle facts required for lawful DEK preservation/destruction
```

`tenant.settings.manage` governs Tenant-family settings/configuration; no standalone `TenantSettings` bounded context or aggregate is implied by the permission.

Key-custody ownership is lifecycle authority only. Cryptographic primitives, KEK integration, wrap/unwrap and secret-material handling remain platform mechanisms. Records Governance supplies retention/hold blocker facts; it does not own key state.

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

Authorization owns the composition/evaluation contract, not domain relationship meaning. Each semantic owner owns its resource/case predicates. No role bypass, generic ACL/ReBAC graph or provider permission engine V1.

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

Approval never owns Document effectivity. Human-readable manifestations of Approval facts are Controlled Information Renditions of Approval authority, not a second Approval record.

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

### Audit

Owns:

```text
append-only AuditEvent timeline
tamper-evidence / chain meaning
audit query/export semantics
Audit Trail separate retention regime
```

Critical governed mutation must be able to append durable Audit intent/event in the same local commit through a published transactionally composable seam. Exact transaction/storage mechanism belongs to R10-B/R10-D. Domain records remain the authority for their own facts.

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
storage provider clients
rendering provider clients
external repository adapters
RLS
HTTP / OpenAPI / codegen
cache
rate limiting
observability
crypto primitives / KEK integration / wrap-unwrap
backup image transport
PlatformOperator / SystemPrincipal execution machinery
```

PlatformOperator/SystemPrincipal remain outside tenant RBAC and gain no implicit tenant-content authority.

---

# 3. R10-A dependency and coordination rules

`internal/composition` is the outer application layer for concrete cross-owner use cases. Composition may coordinate owners; it owns no durable business meaning and no semantic owner imports it.

R10-A requires transactionally composable published application seams wherever frozen semantics require one local atomic commit. R10-B decides the concrete UnitOfWork/Tx/schema mechanism.

Material seams:

1. **CI ↔ Approval** — exact Submission references and Approval reads are composition/read-contract mediated; no mutual package authority. Manifestation Renditions/system values may consume published Approval evidence without making CI a second Approval authority.
2. **Local transactional composition** — published application seams must permit one local DB transaction across exactly the owners required by a frozen atomicity invariant.
3. **Audit append** — Audit publishes a transactionally composable append seam; producers never own Audit storage/meaning.
4. **Artifact confirmation** — caller supplies an opaque semantic-owner reference; Artifact does not import CI/DC. R10-B supplies the structural no-orphan backstop.
5. **Records prospective hold materialization** — Records consumes published CI/DC subject facts/events/read seams; it does not acquire underlying lifecycle authority.
6. **Historical Migration** — target owners expose narrow privileged migration-grade application seams; Interchange calls owners, never the reverse.
7. **Notifications** — producers resolve recipients/business meaning before delivery intent; Notifications does not invent policy.
8. **Authorization filtering** — Authorization composes owner-supplied predicates; Search/export/timeline/API consumers do not rederive visibility.
9. **Tenant erasure/restore** — Organization owns Tenant/tombstone/key-custody lifecycle, Records Governance owns blockers, Authentication owns session revocation, Artifact owns byte truth; composition owns none of them.

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
Local / MinIO / AWS S3 adapters       → Artifact infrastructure
EigenPal / rendering provider adapters → Controlled Information infrastructure/execution
SharePoint / external repo adapters    → Interchange infrastructure
crypto / KEK providers                 → platform mechanism behind Organization-owned key lifecycle
```

---

# 5. Legacy backend disposition fixed by R10-A

| Current module | Promoted target disposition |
|---|---|
| `approval` | converge to Approval V1 |
| `audit` | retain Audit semantic owner; redesign durability later |
| `auth` | converge/rename to Authentication |
| `controlleddocuments` | delete as target BC; identity/numbering responsibilities → Controlled Information |
| `distribution` | retain Distribution semantic owner |
| `documents` | delete legacy BC; responsibilities → Controlled Information |
| `iam` | delete/split → Organization + Authorization |
| `jobs` | delete as BC; owner-attributed work + platform async/composition |
| `notifications` | reclassify → `support/notifications` |
| `render` | dismantle; Rendition/value semantics → Controlled Information, providers → infrastructure |
| `search` | reclassify → `projections/search` |
| `security` | delete as BC; key-custody lifecycle → Organization, AuthN facts → Authentication, commodity security → platform |
| `taxonomy` | dismantle; Area → Organization, DocumentType/classification → Controlled Information, GovernanceClass deleted |
| `templates` | delete parallel lifecycle; template role/use → Controlled Information |
| `tokens` | delete standalone owner; Tenant Dictionary/System Value Catalog → Controlled Information |

No current module survives merely by inertia. Exact table/API/frontend cutover mechanics remain later R10 work.

---

# 6. Surface classification fixed by R10-A

The single OpenAPI + generated-owner-surface structural pattern may survive, but target ownership is:

```text
auth                         → Authentication
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
security                     → retire/split across Authentication/Organization/platform
documentary context          → new owner surface
records governance           → new owner surface
interchange                  → new owner surface
configuration/health/obs     → platform
```

There is no generic `/artifacts` business API merely because Artifact is a semantic supporting owner; confirmed Artifact must be attached to a governed owner. Exact paths, operationIds, DTOs and frontend journeys are R10-E.

Original R10-A candidate/review-packet surface classifications are historical evidence where they differ from this promoted topology.

---

# 7. R10-A closure proof and review record

R10-A was independently attacked and corrected before promotion.

Evidence chain:

1. candidate/review request — `docs/superpowers/analysis/2026-08-17-r10-a-ownership-topology-fable-review-request.md` @ `f51f6bfa`;
2. independent adversarial review — `docs/superpowers/analysis/2026-08-17-r10-a-independent-fable-review.md` @ `c0bde261`, verdict `APPROVE R10-A WITH MATERIAL FIXES`;
3. adjudicated corrected target — `docs/superpowers/analysis/2026-08-17-r10-a-fable-adjudication-corrected-target.md` @ `74c1ba80`;
4. cold delta/global coherence review — `docs/superpowers/analysis/2026-08-17-r10-a-cold-delta-fable-review.md` @ `b8c6f494`, verdict `APPROVE R10-A CORRECTED TARGET WITH MATERIAL FIXES`;
5. final completeness correction — `docs/superpowers/analysis/2026-08-17-r10-a-final-completeness-correction.md` @ `5cb350d5`;
6. independent mechanical completeness sweep — `docs/superpowers/analysis/2026-08-17-r10-a-final-completeness-fable-review.md` @ `ba351578`, verdict `APPROVE R10-A COMPLETENESS CLOSURE WITH MATERIAL FIXES`;
7. operator adjudication accepted the two final mechanical closure fixes on 2026-08-17:
   - `Document owner / responsibility relationship` belongs to Controlled Information, without inventing participant type/cardinality at R10-A;
   - Tenant settings/configuration are subsumed by the Organization/Tenant fact family, without inventing a standalone owner.

The final mechanical sweep found:

```text
BLOCKER                    = 0
remaining topology defect = 0
duplicate owners           = 0
invented fact families     = 0
RetentionPolicy entity     = ABSENT / PASS
R9.5 reopen set            = EMPTY
```

The review files remain evidence, not parallel authority. This page contains the promoted outcome.

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
9. Notifications owns attributed delivery/read state but no business truth; Search is rebuildable projection only.

R10-A reopens only on material evidence such as:

- a frozen/business fact that cannot be owned coherently by the promoted set;
- a real second consumer or independent lifecycle that invalidates the Dictionary collapse;
- a relationship model that materially exceeds frozen Tenant/Area authorization semantics;
- a genuine independent Rendition semantic consumer;
- Evidence/Dossier lifecycle evidence requiring a different boundary;
- an external-transfer requirement proving Interchange must split or a different owner must exist;
- an indivisible multi-file requirement that materially changes Artifact semantics;
- implementation evidence showing the acyclic/transactionally-composable ownership assumptions cannot preserve a frozen invariant.

Package naming preference, current schema inconvenience, provider capability or implementation convenience are not reopen evidence.

---

# 9. Exact next step — R10-B

R10-B is **Transactional Domain State & DB Invariants**. It must derive concrete persistent state and invariant enforcement from R10-A ownership without reopening boundaries for implementation convenience.

At minimum R10-B must decide:

```text
semantic owner → table/aggregate ownership
identity/key/FK/reference rules across owners
Document/Revision/WorkingContent/Submission state model
Document owner/responsibility representation, only to the minimum semantics already frozen
one-open-Revision and exactly-one-EFFECTIVE DB backstops
WorkingContent working_version OCC enforcement
Submission atomicity and immutability
Approval policy/instance/snapshot/decision constraints + SoD backstops
Evidence/Dossier lifecycle and relationship constraints
RetentionBinding/LegalHold/Disposition state constraints
Artifact no-confirmed-orphan structural backstop
Tenant deletion/erasure/tombstone/key-custody durable state
local cross-owner transaction boundaries
same-commit Audit append requirement
outbox intent insertion points required by atomic business mutations
imported-history representation that cannot fabricate native Approval/Release facts
```

R10-B must not decide R10-C physical storage/relocation mechanics, R10-D worker/retry/external-effect execution, R10-E final API/frontend journeys, or R10-F migration/cutover plans except where a minimal schema seam is required now.

Current code/schema/OpenAPI remain evidence only. No product implementation is authorized.