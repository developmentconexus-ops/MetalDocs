# R10-A Ownership Topology — Fable Adjudication and Corrected Target

> **Status:** ADJUDICATED CORRECTED CANDIDATE — AWAITING COLD DELTA REVIEW — **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Candidate packet:** `docs/superpowers/analysis/2026-08-17-r10-a-ownership-topology-fable-review-request.md` @ `f51f6bfa`
> **Independent review:** `docs/superpowers/analysis/2026-08-17-r10-a-independent-fable-review.md` @ `c0bde261`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Authority note:** this artifact records adjudication and the corrected R10-A candidate. It does not promote R10-A into the active ledger/architecture authority. Promotion remains blocked until an independent cold delta review finds no material contradiction.
> **Implementation gate:** **CLOSED.** R10-B and product implementation remain blocked.

---

## 1. Stage result

The independent verdict was:

```text
APPROVE R10-A WITH MATERIAL FIXES
```

Adjudication outcome:

```text
R9.5 reopen set                 = EMPTY
R10-A topology direction        = CONFIRMED
R10-A material findings         = ADJUDICATED
R10-A corrected candidate       = DEFINED HERE
R10-A cold delta review         = REQUIRED / PENDING
R10-A authority promotion       = BLOCKED
R10-B                           = BLOCKED
implementation                  = BLOCKED
```

The review confirmed Authentication, Organization, Authorization, Controlled Information, Approval, Documentary Context, Records Governance, Distribution, Artifact, Audit and Interchange as sustainable owners. `Dictionary` was the only proposed owner that failed the Structural Inversion/subtractive test as presented.

---

## 2. Finding adjudication

| Finding | Adjudication | Corrected decision |
|---|---|---|
| F1 — tenant DEK/key-custody facts ownerless | **ACCEPT** | Organization owns tenant key-custody lifecycle facts required by Tenant erasure/restore semantics. Cryptographic algorithms, KEK integration and wrap/unwrap remain platform mechanisms. Records Governance supplies the lawfulness/blocker facts; it does not own key state. |
| F2 — Authorization filtering contract ambiguous | **ACCEPT** | Authorization owns Permission/Role/RoleAssignment, grant evaluation and the composable authorization/filter contract shape. Each semantic owner owns its resource/case relationship predicates. Authorization must not become a generic relationship graph or second domain authority. |
| F3 — frozen-fact owner inventory incomplete | **ACCEPT / CLOSE DEFECT CLASS** | R10-A closes with the full fact-to-owner inventory in §4, not a patch for only the examples found by the reviewer. |
| F4 — standalone Dictionary owner unsupported | **ACCEPT / RESTRUCTURE NOW** | Delete `Dictionary` as a target supporting owner. Tenant Dictionary and the bounded System Value Catalog become an internal Controlled Information capability because the frozen launch semantics evidence only the governed authoring/revision lifecycle as their business consumer. Keep tenant-managed values and product-shipped values as distinct internal fact classes. `dictionary.manage` remains a semantic Permission. Reopen only on a real second business consumer or materially independent lifecycle. |
| F5 — provenance ambiguity at Interchange | **ACCEPT** | Object-level creation/source provenance belongs to the semantic owner of the object. Interchange owns only transfer/batch/attempt/reconciliation/package provenance. Neither duplicates the other. |
| F6 — Notifications overstated as rebuildable projection | **ACCEPT** | Search remains a rebuildable projection. Notifications moves to attributed support and owns its durable delivery/inbox/read state only; it owns no approval, acknowledgement, authorization or other business truth. |
| F7 — Audit same-commit seam implicit | **ACCEPT / BOUND TO STAGE** | R10-A requires a published transactionally composable Audit append seam for critical governed mutations. Audit remains the single timeline authority; producers never own AuditEvent meaning. Exact transaction/port/DB mechanics are deferred to R10-B/R10-D. |

No finding supplies evidence satisfying the R9.5 reopen contract.

---

# 3. Corrected target ownership topology

## 3.1 Business bounded contexts — exactly 8

### A. Authentication

Owns credential/session identity facts only:

```text
local credential / identity binding
activation
opaque session
lockout / revocation
fresh-auth / reauthentication assurance
```

Does not own Organization membership, Tenant lifecycle or authorization grants.

### B. Organization

Owns organizational identity and Tenant lifecycle:

```text
Tenant
Area
User
Group
GroupMembership
Tenant lifecycle: ACTIVE | SUSPENDED | ERASED
TenantDeletionRequest
TenantErasureRecord
erasure tombstone / erased-tenant reconciliation facts
tenant key-custody lifecycle facts required for lawful DEK preservation/destruction
```

Key-custody ownership means lifecycle authority only. KEK provider integration, cryptographic primitives, wrapping/unwrapping and secret material handling remain platform mechanisms. Records Governance owns retention/hold blockers and supplies the lawfulness input to erasure coordination.

### C. Authorization

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

**Predicate ownership rule:** Authorization owns how grant evaluation composes with a resource/case predicate; it does **not** own each domain predicate's business meaning.

Initial relationship ownership:

| Relationship/predicate meaning | Semantic owner |
|---|---|
| Tenant/Area/User/Group membership and organizational scope | Organization |
| Document/Revision ownership, Area and lifecycle relationships | Controlled Information |
| Dossier scope, Evidence primary-Dossier scope and contextual links | Documentary Context |
| Approval actor qualification, participant snapshot and SoD constraints | Approval, consuming Organization/Authorization facts |
| Hold-to-retention-subject materialization and disposition blockers | Records Governance, consuming subject facts from CI/DC |
| Distribution obligation/audience/acknowledgement relationships | Distribution; these never grant access |
| Imported/exported target-object relationships | the target object's semantic owner; Interchange never redefines them |

Search, export, timeline and other surfaces consume the composed canonical result. No surface may recreate grant or relationship semantics independently.

### D. Controlled Information

Owns the single governed Document/Revision lifecycle:

```text
DocumentType
Document
DocumentRevision
WorkingContent
WorkingSnapshot technical checkpoint
EditorSession
RevisionSubmission
numbering / revision labels
DocumentType approval configuration
Template designation / TemplateUse
TemplateSpec when applicable
EditorialComment
PeriodicReview policy + PeriodicReviewRecord
Rendition business identity/provenance for exact Submission
OfficialRepresentationPolicy
ReleasePlan / ReleaseRecord / effectivity
Tenant Dictionary values
System Value Catalog descriptors/resolution contract
value snapshot bound to the governed revision lifecycle
```

Tenant Dictionary and System Value Catalog remain distinct internally:

```text
Tenant Dictionary     = tenant-managed mutable configuration
System Value Catalog  = product-owned bounded catalog
```

They share no mutable lifecycle merely because both produce values. They reside under Controlled Information because their frozen launch consumer is the Document/Revision authoring lifecycle, which snapshots decision-relevant values so later provider mutation cannot rewrite history.

Controlled Information does **not** own Approval decisions, Evidence/Dossier, retention/hold/disposition, distribution acknowledgement, Audit timeline, transfer process truth, Search, Notifications, storage-provider identity or generic workflow.

### E. Approval

Owns:

```text
ApprovalPolicy(version)
ordered ApprovalStep
actor rule
completion ANY | ALL
ApprovalInstance
activated participant snapshot
ApprovalDecision
reassignment / cancellation / oversight
fresh-auth requirement at decision time
strict SoD
```

It binds an exact RevisionSubmission but never owns Document effectivity.

### F. Documentary Context

Owns:

```text
EvidenceType
Evidence
Evidence naming policy and allowed-format policy reference
DossierType
Dossier
Dossier stable key / scope
creation provenance
ExternalReference
Dossier ↔ Document contextual link
Evidence primary-Dossier relationship
Evidence secondary context links
Evidence occurred/captured/source facts
```

Contextual links never grant access. CAPTURED Evidence reuses its immutable primary Dossier scope. Document lifecycle remains Controlled Information authority.

### G. Records Governance

Owns records-preservation/disposition meaning:

```text
retention policy/rule semantics
RetentionBinding + policy snapshot
RetentionExtension
retention anchors / eligibility facts
LegalHold
materialized held-subject relationship
prospective hold materialization
Disposition authorization / eligibility / completion
DispositionRecord
retention/hold blocker facts used by Tenant erasure
```

Where DocumentType/EvidenceType selects a retention policy, the type owner owns the selected reference while Records Governance owns the policy meaning and all resulting retention-subject lifecycle facts.

### H. Distribution

Owns:

```text
released-document distribution obligation
audience snapshot / historical denominator
AcknowledgementRecord
coverage / completion semantics
```

Distribution never grants access. Notification read/view/download is never acknowledgement.

---

## 3.2 Supporting semantic owners — exactly 3

### I. Artifact

Owns exact-byte and physical-content truth:

```text
Artifact identity
canonical SHA-256
size
closed ContentFormat catalog
media type
technical provenance
staging / validation / confirmation state
managed physical-location facts
relocation verification / cutover facts
restore byte-integrity facts
```

`ContentFormat` authority lives here because it is the canonical technical classification of exact bytes. Domain owners reference it in their allowed-format/representation policies; they do not redefine the catalog.

No confirmed orphan Artifact exists. Artifact never validates semantic owner existence by importing CI/DC; owner-driven confirmation plus the R10-B invariant backstop enforces ownership without a reverse semantic dependency.

### J. Audit

Owns:

```text
append-only AuditEvent timeline
tamper-evidence / chain meaning
audit query/export semantics
Audit Trail's separate retention regime
```

Domain records remain authorities for domain facts. Critical governed mutation requires a durable Audit append through a published transactionally composable seam before success may be reported. Exact mechanism is R10-B/R10-D work.

### K. Interchange

Owns only transfer-boundary process truth for the enumerated launch contracts:

```text
Historical Migration batch / plan / dry-run / per-item outcome / reconciliation
Tenant Portability Export package process
Governed Subject Export package process
External Repository IMPORT_COPY / PUBLISH_COPY attempt/process truth
transfer attempt / package / reconciliation identity
transfer-level source provenance
```

**Provenance split:** source/creation/history facts that become facts of a Document, Revision, Evidence, Dossier, retention subject or other target object belong to that object's semantic owner. Interchange retains only the provenance proving how/when/from-where the transfer process acted.

Interchange is not an ESB, ETL platform, workflow, connector framework or external-system master-data owner.

---

## 3.3 Attributed non-business state / projections / mechanisms

### Notifications — attributed support

Notifications owns only non-business durable delivery state:

```text
delivery intent materialization
delivery attempt/status
user inbox state
read/unread state
```

Business producers resolve business meaning and recipients before emitting delivery intent. Notifications never queries policy to invent recipients and never turns delivery/read state into Approval, Distribution acknowledgement or other business evidence.

Target placement:

```text
internal/support/notifications
```

### Search — rebuildable projection

Search owns no canonical business state. Its index/projection is rebuildable and eventually consistent. Query surfaces must reapply canonical Authorization plus owner-supplied relationship predicates.

Target placement:

```text
internal/projections/search
```

### Commodity mechanisms

These are not semantic owners:

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
crypto primitives / KEK integration / wrap-unwrap implementation
backup image transport
```

---

# 4. Frozen fact → exactly-one-owner closure inventory

This inventory is normative for the corrected candidate and exists to close F3's defect class. It groups every frozen durable/business fact family rather than mirroring current tables/modules.

| Frozen fact family | Exactly one semantic owner | Boundary note |
|---|---|---|
| local credentials, activation, session, lockout/revocation, fresh-auth assurance | Authentication | external IdP remains adapter seam |
| Tenant, Area, User, Group, GroupMembership | Organization | credentials excluded |
| Tenant lifecycle state | Organization | exactly ACTIVE/SUSPENDED/ERASED |
| deletion request, erasure completion record, erased-tenant tombstone/reconciliation facts | Organization | retention blockers are inputs from Records Governance |
| tenant DEK/key-custody lifecycle facts needed for lawful preservation/destruction | Organization | crypto/KEK/wrap mechanics remain platform |
| Permission, Role, RoleAssignment, grant/revocation, typed scope | Authorization | no bypass/ReBAC/deny engine |
| canonical grant evaluation + composition contract shape | Authorization | domain predicate meaning remains with each domain |
| DocumentType, numbering, document approval/representation/review policy references | Controlled Information | referenced Approval/Records/Artifact authorities remain external |
| stable Document identity + Area relationship | Controlled Information | contextual Dossier link does not change it |
| DocumentRevision identity/state/REV number | Controlled Information | one open + at most one EFFECTIVE V1 |
| WorkingContent, working_version OCC truth, WorkingSnapshot, EditorSession | Controlled Information | editor/provider state is not authority |
| immutable RevisionSubmission + decision-relevant digest/provenance snapshot | Controlled Information | Approval/Rendition/Release bind this identity |
| Template designation, TemplateUse, TemplateSpec | Controlled Information | no parallel template lifecycle |
| EditorialComment | Controlled Information | DRAFT collaboration only |
| PeriodicReview policy/record | Controlled Information | no separate review BC |
| Rendition identity/output hash/generator provenance for exact Submission | Controlled Information | renderer is mechanism |
| ReleasePlan, ReleaseRecord, EFFECTIVE/SUPERSEDED transition | Controlled Information | release coordinator owns nothing |
| Tenant Dictionary values | Controlled Information | tenant-managed configuration, snapshot by consuming revision lifecycle |
| bounded System Value Catalog descriptors/resolution contract | Controlled Information | product-owned; distinct internally from Tenant Dictionary |
| EvidenceType + evidence naming/allowed-format policy reference | Documentary Context | `ContentFormat` catalog itself is Artifact authority |
| Evidence lifecycle/content metadata/source/occurred/captured facts | Documentary Context | primary bytes referenced from Artifact |
| DossierType, Dossier identity/stable key/scope | Documentary Context | no custom-object platform |
| creation provenance + ExternalReference | Documentary Context | transfer-attempt provenance stays Interchange |
| Dossier↔Document and Evidence contextual relationships | Documentary Context | never grants access |
| retention policy/rule semantics | Records Governance | type owners may hold selected policy reference |
| RetentionBinding snapshot/anchor/eligibility + RetentionExtension | Records Governance | later policy mutation does not rewrite bindings |
| LegalHold + materialized/prospective held-subject relationships | Records Governance | underlying subject lifecycle remains CI/DC |
| Disposition decision/authorization/completion + DispositionRecord | Records Governance | physical byte truth stays Artifact |
| distribution obligation/audience snapshot/coverage | Distribution | access remains canonical AuthZ |
| AcknowledgementRecord | Distribution | Notifications never substitutes |
| Artifact identity/hash/size/media type/technical provenance | Artifact | no user-facing file-library identity |
| closed ContentFormat catalog | Artifact | consumers reference; no duplicate catalogs |
| Artifact staging/validation/confirmation + managed-location facts | Artifact | provider is mechanism |
| relocation/cutover + byte restore-integrity facts | Artifact | no new Artifact/REV/Submission on relocation |
| AuditEvent/tamper-evident timeline/query/export | Audit | not second domain-state authority |
| Audit Trail separate retention regime | Audit | retention mechanics may reuse platform machinery without moving authority |
| Historical Migration batch/plan/dry-run/per-item/reconciliation truth | Interchange | imported target facts written through target owners |
| portability/governed-export package process/manifests | Interchange | exported objects retain original owners |
| external IMPORT_COPY/PUBLISH_COPY attempt/process truth | Interchange | connectors are adapters |
| transfer/batch/attempt provenance | Interchange | object-level provenance belongs target owner |
| historical imported object-level provenance/governance facts | target semantic owner of the imported object/fact | never fabricated as native MetalDocs Approval/Release/history |
| imported retention anchor/unknown-anchor meaning | Records Governance | unknown never becomes deletion-eligible silently |
| notification delivery/inbox/read state | Notifications support | non-business, non-rebuildable state |
| search index/discovery state | Search projection | rebuildable; no authority |
| provider/job/outbox/worker retry/lease/DLQ state | owning mechanism, attributed to producer intent | mechanism does not acquire business meaning |
| backup bytes/image transport | platform operations | restore must reapply Organization tombstones and reconcile Records/Artifact facts before service resumes |

**Completeness rule for R10-B–R10-F:** if a later table/event/API introduces a durable fact not classifiable by this inventory, it is a material R10-A contradiction until explicitly adjudicated. Later stages may refine representation, not silently create a twelfth semantic authority.

---

# 5. Cross-owner coordination and dependency DAG

## 5.1 Coordination rule

`internal/composition` is the outer application layer for concrete cross-owner use cases. Composition may coordinate owners; it never owns durable domain meaning. No semantic owner imports `composition`.

R10-A requires **transactionally composable published application seams** wherever frozen semantics require one local atomic commit. This is a topology constraint, not the final transaction API. R10-B decides concrete UnitOfWork/Tx/DB boundaries and constraints.

### Submission coordination

```text
CI accepted WorkingContent
→ CI immutable RevisionSubmission
→ Records Governance first-submission RetentionBinding
→ Approval ApprovalInstance when configured
→ Audit durable append
→ durable async intents when required
```

Terminal authority: Controlled Information owns Submission. No generic workflow owner appears.

### Release coordination

```text
exact CI RevisionSubmission
+ Approval satisfaction when configured
+ required CI Rendition ready
+ governing preconditions
→ CI atomic effectivity transition
```

The coordinator reads Approval and calls CI; CI and Approval do not import each other to implement the orchestration.

### Disposition coordination

```text
Records Governance eligibility + no active hold
+ governed subject still permits disposal
+ Artifact verified physical removal
→ Records Governance DispositionRecord completion
```

### Tenant erasure / restore reconciliation

```text
Organization deletion request/tombstone state
+ Records Governance retention/hold blocker facts
+ Authentication session revocation
+ owner-attributed substantive data destruction
+ Artifact byte destruction/integrity facts
+ Organization key-custody lifecycle transition when lawful
→ Organization Tenant ERASED / TenantErasureRecord
```

On restore, Organization tombstones are reapplied first; Records Governance and Artifact facts are reconciled before cleanup/service resumes. Backup transport does not own this process meaning.

## 5.2 Explicit material seams

1. **CI ↔ Approval:** interaction is composition-mediated and/or reference/read-contract based; no mutual package authority.
2. **Local transactional composition:** owner application seams must permit one local DB transaction across the exact owners required by a frozen atomicity invariant. Exact mechanism is R10-B.
3. **Audit append:** Audit publishes a transactionally composable append seam; producers never write Audit storage as their own schema/table authority.
4. **Artifact confirmation:** caller supplies an opaque semantic-owner reference; Artifact does not import CI/DC to validate ownership. R10-B supplies the structural no-orphan backstop.
5. **Records prospective hold materialization:** Records consumes published CI/DC subject-entered-scope facts/events/read seams; no Records back-edge acquires subject lifecycle authority.
6. **Historical Migration:** target owners expose narrowly privileged migration-grade application seams; Interchange calls them. Target owners never depend on Interchange.
7. **Notifications:** producers resolve recipients/business meaning using Organization/Authorization/domain contracts before delivering intent; Notifications does not query policy to invent semantics.
8. **Authorization filtering:** Authorization composes owner-supplied predicates; Search/export/timeline consume the canonical result rather than re-deriving visibility.

## 5.3 Semantic dependency direction

Arrows below mean **consumer depends on published facts/contracts from provider**, not necessarily a direct Go import.

```text
Authentication       → Organization
Authorization        → Organization
ControlledInformation→ Organization, Authorization, Artifact
Approval             → Organization, Authorization, ControlledInformation(reference/read contract)
DocumentaryContext   → Organization, Authorization, Artifact, ControlledInformation(reference/read contract)
RecordsGovernance    → ControlledInformation + DocumentaryContext published subject facts
Distribution         → Organization, Authorization, ControlledInformation release facts
Audit                ← append intents from all critical mutating owners through its published seam
Interchange          → Organization, Authorization + applicable target owners
Notifications        ← producer-resolved delivery intents
Search               → owner events/read models + Authorization composition contract
composition          → any owners required by one concrete cross-owner use case
```

Package dependency must remain acyclic through interface inversion, reference types, published read/application contracts and composition. R10-B/R10-D may choose the concrete shape but may not introduce semantic cycles.

---

# 6. Corrected target filesystem classification

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
    <only concrete cross-owner application use cases>

  platform/
    <commodity db/http/async/crypto/provider/observability mechanics>
```

Within an owner, add `domain/`, `application/`, `infrastructure/`, `delivery/`, `api/` or `public/` only where a real responsibility/consumer justifies the package. `public/` is not mandatory.

Provider placement remains subordinate to ownership:

```text
Local / MinIO / AWS S3                  → Artifact infrastructure
EigenPal / renderer / Gotenberg         → Controlled Information infrastructure/execution
SharePoint / external repository        → Interchange infrastructure
KEK / wrap-unwrap / crypto primitives   → platform security mechanism
```

---

# 7. Corrected legacy module disposition

| Current module | Corrected target disposition |
|---|---|
| `approval` | converge to Approval V1 |
| `audit` | retain Audit semantic owner; redesign durability around published in-tx append seam |
| `auth` | rename/converge to Authentication |
| `controlleddocuments` | delete as BC; stable identity/numbering/effectivity meaning into Controlled Information |
| `distribution` | retain Distribution semantic owner; rederive contracts |
| `documents` | delete legacy BC shape; governed lifecycle into Controlled Information |
| `iam` | delete/split into Organization + Authorization |
| `jobs` | delete as module; rehome owner-attributed work and keep scheduler/retry machinery in platform |
| `notifications` | move to `support/notifications`; retain delivery/inbox/read state only |
| `render` | dismantle: Rendition/value-catalog semantics → Controlled Information; provider execution → infrastructure/platform as applicable |
| `search` | move to `projections/search`; canonical access remains external |
| `security` | dismantle: tenant lifecycle + key-custody lifecycle facts → Organization; AuthN facts → Authentication where applicable; crypto/KEK/operational controls → platform |
| `taxonomy` | delete/split: Area → Organization; DocumentType/classification meaning → Controlled Information; deleted GovernanceClass stays deleted |
| `templates` | delete parallel lifecycle; template role/TemplateUse/TemplateSpec/value authoring semantics → Controlled Information |
| `tokens` | delete target standalone owner; Tenant Dictionary → Controlled Information |

R10-F must map individual legacy jobs/tables/endpoints to these destinations before deletion; no indefinite compatibility owner survives.

---

# 8. Global Maximum / YAGNI result

The corrected candidate contains:

```text
8 business bounded contexts
3 supporting semantic owners
1 attributed durable support component (Notifications)
1 rebuildable projection (Search)
commodity mechanisms outside authority
```

No new Governance, Release, Workflow, Jobs, Rendering, Search, Notifications, Security, Connector Platform, generic Integration, BPM or ReBAC bounded context is introduced.

The subtractive change from the reviewed candidate is deliberate: standalone `Dictionary` disappears because no frozen second business consumer or independent lifecycle currently justifies the boundary. The seam remains cheap to extract later because value fact classes and their consumer contract stay explicit inside Controlled Information.

---

# 9. Proof obligations for cold delta review

The independent cold reviewer must attempt to falsify only the adjudicated/corrected delta plus its global coherence effects:

1. find any frozen durable/business fact still without exactly one owner;
2. find any fact now assigned to two authorities;
3. falsify Organization ownership of key-custody lifecycle facts without promoting crypto mechanism to authority;
4. find domain relationship semantics that Authorization would still need to own centrally;
5. show a frozen real second consumer or independent lifecycle that requires standalone Dictionary now;
6. show that moving Notifications out of `projections/` creates or hides business authority;
7. find an unavoidable semantic/package cycle in the explicit seams/DAG;
8. show that the transactional-composability constraint prejudges R10-B or, conversely, is too weak to preserve frozen atomicity;
9. find any provenance duplication between Interchange and target owners;
10. rerun Structural Inversion + subtractive/YAGNI on the final 8+3 topology;
11. verify the R9.5 reopen set remains EMPTY unless the strict five-part reopen contract is satisfied.

Legal cold-review outcomes:

```text
APPROVE R10-A CORRECTED TARGET
APPROVE R10-A CORRECTED TARGET WITH MATERIAL FIXES
DO NOT APPROVE R10-A CORRECTED TARGET
```

A reviewer finding remains evidence, not authority. No reviewer may promote this artifact, amend R9.5, open R10-B or implement product code as part of the review.

---

# 10. Reopen triggers after eventual R10-A promotion

Reopen only the implicated ownership decision on material evidence such as:

- a real second Tenant Dictionary/System Value consumer with an independent lifecycle that makes extraction from Controlled Information materially safer;
- a new resource relationship model that cannot be represented by the frozen Tenant/Area Authorization contract without centralizing domain semantics;
- Evidence/Dossier lifecycles becoming materially incompatible with one Documentary Context owner;
- a second real Rendition consumer proving Rendition needs an independent semantic owner;
- external-transfer semantics proving the enumerated Interchange contracts no longer share one coherent process-truth boundary;
- a new requirement changing Tenant/User/Area ownership;
- true indivisible multi-file Artifact semantics changing the Artifact boundary;
- implementation evidence showing a frozen atomicity invariant cannot be realized under the acyclic owner seams without distributed authority.

Package naming preference, current schema convenience, provider capabilities or migration cost are not reopen evidence.
