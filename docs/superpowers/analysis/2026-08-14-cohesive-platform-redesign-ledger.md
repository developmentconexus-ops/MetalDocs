# MetalDocs Cohesive Platform Redesign — Active Decision Ledger

> **Status:** ACTIVE WIP — operator-approved decisions below are binding; unresolved items are explicit.
> **Date:** 2026-08-14
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** `wiki/standards/root-cause-global-maximum-method.md`
> **Program authority:** `wiki/architecture/cohesive-platform-redesign.md`
> **Implementation gate:** **CLOSED — design/documentation only.**

---

# 0. Fresh-session contract

Read in order:

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. this ledger
5. `wiki/references/current-agent-handoff.md`

Never revive old roadmaps/specs/ADRs/runtime concepts by inertia. Current code/schema/OpenAPI are migration evidence, not target-design authority.

Global maximum = **smallest professional architecture that correctly models the domain, preserves invariants and exposes clean extension seams.**

```text
product/domain semantics
→ invariants + lifecycle
→ Organization/AuthZ/Approval integration
→ whole-product completion
→ build-vs-buy
→ bounded contexts / ownership
→ data model + DB constraints
→ transaction/event contracts
→ API + frontend journeys
→ migration/delete map
→ implementation specification
→ implementation plan
→ code
```

Core rules:

1. every business fact has one canonical authority;
2. provider/editor/repository technology never becomes business identity;
3. Approval/Rendition/Release always bind the same exact immutable Submission/digest;
4. no generic BPM/ReBAC/low-code/object-platform engine without proven need;
5. no confirmed binary content without semantic ownership by a governed business object.

Historical root defect: browser QA proved humans could review edited bytes while freeze/render later selected blank-template bytes. The target must make that class impossible by construction.

---

# 1. LOCKED — Authentication / Organization / Authorization (R1–R2, R9)

## Authentication

Local AuthN remains V1 behind a future external-IdP seam. It owns sessions, credential lifecycle, lockout, activation and fresh-auth assurance. Conceptual principal exposes `user_id`, `tenant_id`, auth time/method/assurance.

Real MFA/passkeys/SSO/SAML/federation trigger re-evaluation of Keycloak/external IdP before rebuilding IdP features internally. Current fake/stub MFA coverage has no target entitlement.

## Organization

```text
Tenant
Area
User
Group
GroupMembership
```

Groups are flat V1; user may belong to multiple groups. Area is the one organizational truth reused by Document ownership, scoped authorization and Approval actor resolution.

## Authorization

Five tenant roles only:

```text
tenant_owner
area_manager
author
approver
viewer
```

One grant shape:

```text
RoleAssignment
  subject: User | Group
  role
  scope: TenantScope | AreaScope
  granted_by / granted_at
  revoked_by? / revoked_at? / reason?
```

Additive + default deny. No tenant-owner bypass, generic ACL/ReBAC graph, nested groups, deny engine or magic `"tenant"` scope sentinel.

Decision equation:

```text
Permission
+ required case/resource relationship
+ Domain Governance constraints
= ALLOW
```

RLS = tenant-isolation defense-in-depth only. DB constraints = structural invariants only. Current `system_admin` short-circuit/asserted-capability-GUC model is not target architecture.

### Final R9 Permission Catalog for the R3–R9 operation set

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
tenant.export
tenant.deletion.request
```

Role bundles:

- `viewer`: `document.read_effective`.
- `author`: viewer + history/working/create/edit/comment/submit/periodic-review qualification.
- `approver`: viewer + `approval.act`; no blanket draft access.
- `area_manager`: author + revision cancel/obsolete/owner manage + approval act/oversee/reassign/cancel + distribution manage/oversee in Area.
- `tenant_owner`: all 29 through normal Authorizer; still obeys relations, SoD, lifecycle, fresh-auth and tenant-operability constraints.

Narrow self/case operations do not become fake Permissions: own notifications/sessions/password, exact Approval case, exact Distribution assignment/ack, submitter withdrawal, system release/rendition/erasure.

Approval SoD V1:

1. actor cannot accept if Revision creator or Submission submitter;
2. same user cannot accept two distinct Steps of one ApprovalInstance;
3. reassignment target must remain active/qualified and satisfy SoD.

Last active tenant owner cannot be revoked/deactivated. Responsible owner of review-governed Documents must be reassigned before deactivation.

**R9.5 delta rule:** new Evidence/Dossier operations receive Permissions only after their lifecycles close; operations first, authorization delta second.

---

# 2. LOCKED — Approval V1

Specialized sequential human workflow, not BPM.

```text
ApprovalPolicy(version)
  ordered ApprovalStep[]

Step:
  purpose: review | approval
  actor_rule: NamedUser | Group | RoleInArea
  completion: ANY | ALL
  requires_reauthentication
  due_in_days?
```

Human outcomes: `accept | return_for_changes`. Separate ops: `withdraw | cancel | reassign`. No normal terminal reject V1.

Participants materialize when a Step activates and are snapshotted as evidence; action revalidates current qualification/SoD.

ApprovalInstance binds exactly one immutable `RevisionSubmission`. Return-for-changes/allowed withdrawal terminates the attempt and returns the **same REV** to DRAFT. Resubmission creates a new Submission and, when required, a new ApprovalInstance. After completed Approval V1 does not reopen candidate content; cancel candidate + create a new REV if content must change.

---

# 3. LOCKED — Controlled Information configuration / lifecycle / authoring foundation (R3–R5)

## DocumentType / classification

`DocumentProfile` → tenant-scoped `DocumentType`:

```text
id
 tenant_id
 code           // immutable
 name
 description?
 category_id?
 status: ACTIVE | INACTIVE
```

`DocumentTypeCategory` is navigation/reporting only. `GovernanceClass` deleted.

Approval config is explicit:

```text
NoHumanApproval
or UsePolicy(ApprovalPolicyID)
```

## Document / Revision

`Document` = stable governed identity. At most one EFFECTIVE + one open Revision V1.

Official business labels:

```text
REV001
REV002
REV003
...
```

Revision states:

```text
DRAFT
SUBMITTED
EFFECTIVE
SUPERSEDED
OBSOLETE
CANCELLED
```

Autosaves/checkpoints are technical working history, never business REV numbers. REV is allocated when a change cycle starts and never reused. `REV002+` requires reason-for-change before first submit.

`SUPERSEDED` = newer Revision of same Document became effective. `OBSOLETE` = Document retired without successor; terminal V1.

## RevisionSubmission

Immutable exact attempt:

```text
REV002
  Submission #1 → digest A → returned
  Submission #2 → digest B → released
```

Approval/Rendition/Release always bind exact Submission/digest. No mutable submitted bytes.

## Numbering

`Document.code` tenant-wide unique/immutable; Document type/Area/code immutable V1.

DocumentType numbering language:

```text
literals + {TYPE} + {AREA} + {SEQ}
sequence_scope: TYPE | TYPE_AREA
sequence_width: minimum zero padding
```

No year/month/custom metadata/formulas/scripts/resets. Normal Create has no manual code override; legacy preservation belongs to explicit import/migration.

## Templates / structured authoring

Template is a role of an ordinary governed Document. No parallel TemplateVersion lifecycle. `TemplateUse` is M:N; source current EFFECTIVE template REV resolves once at derived-document creation and exact origin is pinned forever.

TemplateSpec owns structured authoring/fill contract only:

```text
TemplateField
  key
  label
  value_type: text | date | number | choice | user | image
  source: user_input | system(key) | dictionary(key)
  typed constraints...
  visible_if?
```

No generic expression language. Source anchors and TemplateSpec must agree before template submission.

**R9.5 refinement:** TemplateSpec/field-value state is conditional structured-authoring state, not universal Revision content.

---

# 4. LOCKED — Periodic Review / Rendition / Release (R6 refined by R9.5-1)

Periodic Review belongs Controlled Information:

```text
Disabled
or Every(n months)
```

Cadence starts from actual Effectivity and restarts after completed review. Due/overdue does not invalidate EFFECTIVE content. Immutable `PeriodicReviewRecord` binds exact current REV with `confirmed_current | change_required`. Review requires responsible-owner relation + `document.review_periodic` + exact REV still current.

Rendition = immutable derived representation of exact Submission with output hash + generator/build provenance. Approval approves Submission, never renderer output bytes.

Release is automatic/system-owned; no publish button. Optional `ReleasePlan.not_before`; actual `effective_at = released_at`.

Winning transaction atomically:

```text
candidate REV -> EFFECTIVE
prior REV -> SUPERSEDED
Document.effective_revision_id -> candidate
Document.open_revision_id -> null
ReleaseRecord
outbox/events
```

Candidate may be cancelled after Approval and before Release; historical evidence remains.

### R6-12 replacement

Universal mandatory `OFFICIAL_PDF` is retired. Locked V1 policy:

```text
OfficialRepresentationPolicy =
    SourceOnly
  | RequireRendition(ContentFormat)
```

At most one required derived rendition V1. Exact primary source Artifact is always frozen by Submission. Required rendition must derive from that exact Submission.

Examples:

```text
DOCX → source DOCX + RequireRendition(PDF)
PDF  → SourceOnly
XLSX → SourceOnly
SVG  → SourceOnly
CAD  → native source + optional/required viewable only if type policy says so
```

---

# 5. LOCKED — Distribution / Values / Audit / Notifications / Search (R7)

Distribution = obligation/acknowledgement, not AuthZ or LMS.

```text
DistributionConfiguration =
  None
  | ReadAcknowledgement {
      targets: User | Group
      due_in_days?
      requires_reauthentication
    }
```

Release snapshots concrete users. Later Group membership never rewrites historical denominator. Explicit immutable `AcknowledgementRecord` completes obligation; notification read/view/download never does. New effective REV supersedes pending prior-REV obligations and materializes fresh assignments.

System Value Catalog V1:

```text
document_code
revision_label
revision_title
document_type_code
document_area_code
document_area_name
revision_created_by_name
```

Tenant Dictionary is mutable source data; values resolve/snapshot when a **new REV** is created. Same-REV return/resubmit does not silently re-resolve. Historical content never depends on live Dictionary state.

Domain evidence (`RevisionSubmission`, ApprovalDecision, PeriodicReviewRecord, ReleaseRecord, DistributionAssignment, AcknowledgementRecord, RoleAssignment history...) remains authority. AuditEvent is transversal timeline only. Critical governed mutation cannot report success without durable audit intent/event in same commit boundary. Usage telemetry may be async.

Audit Trail remains append-only, tamper-evident, exportable, with explicit User/System actors.

Notifications = delivery projection only. Search = rebuildable/eventually-consistent projection; stale result never grants canonical resource access. No Elasticsearch/OpenSearch requirement yet.

---

# 6. LOCKED — Tenant lifecycle / Platform Security (R8)

`PlatformOperator` / `SystemPrincipal` are outside tenant RBAC and gain no implicit tenant-content access.

Tenant lifecycle:

```text
ACTIVE
SUSPENDED
ERASED
```

Deletion request is separate: `PENDING | CANCELLED | EXECUTED` with grace period. Onboarding creates Tenant + initial User + `tenant_owner @ Tenant` + single-use time-limited activation credential; platform operator never chooses tenant-owner password.

Suspension revokes sessions, blocks login/business mutations and is reversible. Business jobs respect suspension; lifecycle/security jobs may continue.

Tenant export/deletion request require fresh-auth.

Erasure direction:

```text
request due
→ suspend
→ revoke sessions
→ erase eligible live tenant rows/blobs
→ destroy Tenant DEK
→ preserve allowed non-PII audit/platform skeleton
→ ERASED
→ TenantErasureRecord
```

Audit Trail itself is not deleted. Platform KEK wraps per-tenant DEK; no per-document key hierarchy/rotation/HSM product V1.

**Retention refinement pending:** terminal erasure must respect legal-retention/hold obligations.

Backup/restore must reapply erasure tombstones before service availability.

---

# 7. LOCKED — R9.5 Whole-Product North Star

R9.5 exists because storage, editor, non-DOCX content and enterprise-context use cases materially affect architecture. The previously proposed R10-A topology is **not approved** and remains paused.

> **MetalDocs is the system of record for identity, governance, revision, evidence and documentary context. Physical storage, authoring/editor technology, viewers and upstream ERP/PLM/repositories are replaceable providers/connectors around that kernel.**

Do not replace MetalDocs with M-Files/Nuxeo/Alfresco and do not build a generic M-Files clone. Build only the governance product; buy/reuse object storage, Office editing, CAD/PLM, ERP, malware scanning etc.

A governed Document is format-agnostic. The question is whether information needs stable official identity + `REVxxx`, not whether it is DOCX.

`Dossier` is a deliberately small documentary context for Venda/Produto/Projeto/Equipamento/Case etc. It may relate Documents/Evidence and external ERP/PLM references, but never becomes ERP/PLM. BOM/where-used/EBOM/MBOM/CAD dependency/ECR/ECO are explicit PLM integration triggers.

EigenPal is a DOCX authoring provider, never Document identity.

Storage classes:

1. Managed Artifact Store — MetalDocs logically owns exact bytes.
2. External Repository Connector — SharePoint/OneDrive/etc. import/publish copies explicitly.
3. Future enterprise content profile — e.g. SharePoint Embedded/M365, designed explicitly rather than forced into S3 semantics.

Provider WORM/Purview may enforce retention physically, but MetalDocs owns business retention/hold semantics if it owns the record.

Freeze rule: a future idea reopens the kernel only when it provides a material identity/historical-truth/invariant counterexample. Otherwise land it on an existing provider/connector/context seam.

---

# 8. LOCKED — R9.5-1 Content Model

## Artifact

`Artifact` = immutable technical identity of exact bytes. It is not a user-facing business object and carries no business meaning.

Conceptually: tenant, exact SHA-256, byte size, ContentFormat/media type, technical provenance. Any byte change creates a new Artifact.

Artifact identity and `RevisionSubmission.submission_digest` are provider/location independent. Bucket/key/version/URL never enter business identity/digest.

## Staging / semantic ownership

There is no confirmed orphan Artifact library. UX may begin with a file drop inside a known Dossier:

```text
file drop
→ temporary staging
→ detect ContentFormat
→ choose/confirm compatible EvidenceType
→ collect narrow semantic reference when needed
→ allocate canonical evidence filename
→ create Evidence + confirm Artifact
→ relate Evidence to Dossier
```

Temporary staging bytes are non-business and garbage-collectable.

Invariant:

```text
CONFIRMED Artifact
  must belong to
    DocumentRevision
    or Evidence
```

## EvidenceType / naming

Tenant-scoped EvidenceType V1:

```text
id
 tenant_id
 code
 name
 description?
 status: ACTIVE | INACTIVE
 allowed_formats[]
 naming_policy
```

Examples: NOTA_FISCAL, XML_NFE, COMPROVANTE_ENTREGA, FOTO_INSPECAO, CERTIFICADO_TESTE, DOC_CLIENTE.

Not a generic custom-field/schema/workflow engine.

User filename is provenance only. Official filename is MetalDocs-generated.

Evidence naming language V1:

```text
{TYPE}
{DOSSIER}
{REF}
{SEQ}
```

Extension comes from validated ContentFormat. `Evidence.reference?` is a narrow optional business reference, not metadata bag. Human upload and ERP/PLM capture use identical naming semantics.

Format narrows candidate EvidenceTypes but does not silently determine business meaning; future AI may suggest. Trusted integration may supply type by contract.

## Evidence lifecycle

Evidence does not use REVxxx.

```text
DRAFT
  ↓ capture
CAPTURED
  ↓ administrative invalidation of wrong capture
VOIDED
```

DRAFT content may be replaced. CAPTURED governed content/metadata is immutable. Wrong capture is VOIDED with reason and replaced by new Evidence; history is never rewritten.

VOIDED means the MetalDocs capture itself was invalid, not that the external-world subject was later cancelled.

Evidence does not use Approval/Release by default; if information needs controlled revision/approval lifecycle, use Document.

Evidence capture is auditable and may have User or System/Integration provenance.

## One primary Artifact

Exactly one primary Artifact per DocumentRevision and one per Evidence V1. Independent governed information should be separate Documents/Evidence related through context. True indivisible multi-file packages are future `ArtifactPackage`/PLM trigger.

## ContentFormat / RevisionContent

MetalDocs owns a closed product ContentFormat catalog. DocumentType/EvidenceType select allowed formats; filename extension alone is never trusted.

```text
RevisionContent
  primary_artifact
  governed_metadata
  structured_authoring?   // optional
```

TemplateSpec/field values live in optional structured-authoring state where applicable.

Submission digest binds exact primary Artifact hash + governed state + decision-relevant authoring/template provenance, never storage location.

DocumentType may select allowed formats and `SourceOnly | RequireRendition(format)`, but never editor/storage provider names.

Trusted integrations operate in domain terms (`Create Evidence type=XML_NFE ... capture Artifact`), never generic file-drop semantics.

Documents/Evidence/Dossiers are discovery units; Artifact is not.

---

# 9. LOCKED — R9.5-2 Storage / Repository Strategy

## ST-01 — Provider-independent Artifact truth

Artifact identity/hash and any Submission digest are provider-independent. Physical storage location, provider object version, ETag and URL are infrastructure/supporting evidence only.

Product canonical content hash = **SHA-256 of exact primary bytes**.

## ST-02 — One Managed Artifact Store per deployment V1

V1 has one active Managed Artifact Store per MetalDocs deployment, not per-tenant/per-document routing.

First-class adapters:

```text
Local      // dev/test
MinIO
AWS S3
```

MinIO and AWS S3 share the managed-store contract. Other S3-compatible products require conformance validation before official support; do not promise universal compatibility.

## ST-03 — Provider migration

MinIO→S3 or other provider migration = copy exact bytes + verify canonical hash + cut over physical location. It does **not** create a new Artifact, REV or Submission because business content did not change.

Permanent dual-write/active-active replication is not V1.

## ST-04 — Opaque immutable physical keys

Managed object keys are opaque, immutable and tenant-namespaced, conceptually e.g.:

```text
tenants/{tenant_id}/artifacts/{artifact_id}
tenants/{tenant_id}/staging/{upload_id}
```

Do not encode Dossier/business filename/path in the storage key. Canonical business filename remains metadata.

Artifact ID != content hash. No content-addressed semantic dedup V1 and no cross-tenant dedup.

## ST-05 — Staging and direct upload

Temporary upload/staging is allowed and is not a confirmed Artifact.

Browser may upload directly to managed storage via a short-lived provider grant/presigned mechanism to avoid proxying large files through the Go API.

Provider upload success alone does not confirm an Artifact. Confirmation requires integrity + content + semantic validation.

Never overwrite an existing Artifact key; new content always gets a new upload/artifact identity.

## ST-06 — Minimal Managed Artifact Store contract

Business/domain code may rely only on a small provider-neutral capability such as:

```text
stage/put
complete + verify
read/open
stat
delete
short-lived upload/download grant where supported
```

Bucket creation, replication setup, lifecycle configuration, KMS policy etc. are deployment infrastructure, not domain operations.

## ST-07 — Versioning / WORM

Object-store versioning is optional defense-in-depth, never `REVxxx` authority. MetalDocs correctness does not depend on rewriting the same key.

Object Lock/WORM/legal-hold support is an optional provider enforcement capability consumed by future MetalDocs Retention governance; provider lock state does not become business retention authority.

## ST-08 — Encryption

Production baseline: encrypted transport + provider encryption at rest.

Do **not** encrypt every Artifact with the MetalDocs Tenant DEK V1. Tenant DEK remains for erasable-retained payloads such as protected Audit payloads. Application-level Artifact crypto-shred is introduced only if R9.5-5 Retention/Erasure proves a concrete need.

Avoid duplicating object-store encryption in a way that unnecessarily breaks presigned access, rendering, Office integration, range reads or future repository profiles.

## ST-09 — External repositories are not object stores

Normal SharePoint/OneDrive/Google Drive/CMIS repositories are **External Repository Connectors**, not ManagedArtifactStore providers.

Governed primary content V1 cannot rely only on an external reference. Before a DocumentRevision/Evidence becomes governed/final, MetalDocs captures an exact managed copy of the bytes.

Connector V1 semantic directions:

```text
IMPORT_COPY
PUBLISH_COPY
```

A bare external reference may decorate a Dossier, but does not satisfy `DocumentRevision.primary_artifact` or captured `Evidence.artifact`.

## ST-10 — External drift never mutates history

If an externally published/copied item changes, existing MetalDocs EFFECTIVE REV/CAPTURED Evidence does not change.

Future explicit adoption flow:

```text
external version changed
→ surface drift
→ authorized adopt/import
→ fetch exact external bytes
→ new Artifact
→ new DRAFT REV where applicable
```

No silent bidirectional sync V1.

## ST-11 — SharePoint Embedded is a future profile

SharePoint Embedded/M365 may become a Microsoft-enterprise content backend/coauthoring/Purview profile. It is deliberately **not** forced into the ordinary S3 adapter or normal repository-connector abstraction because its Entra/container/billing/compliance semantics are materially different.

R9.5/R10 must avoid provider-specific business identities so this profile remains possible later.

## ST-12 — Restore / erasure / GC

A valid restore requires:

```text
DB Artifact record exists
+ exact managed bytes exist
+ SHA-256(bytes) == Artifact.content_hash
```

Provider versions/backups never redefine domain lifecycle state.

Tenant erasure and cleanup use MetalDocs inventory as semantic authority and tenant-namespaced storage as safety boundary. Staging/incomplete uploads are explicitly garbage-collectable.

No generic multi-cloud routing, BYOS-per-tenant, permanent cross-provider replication, dedup engine or generic synchronization platform V1.

---

# 10. Build-vs-buy rulings

| Technology/class | Ruling | Revisit trigger |
|---|---|---|
| M-Files/Nuxeo/Alfresco as core | do not adopt as kernel | product strategy changes to customization on another ECM |
| Jackrabbit/JCR | no | stack/domain economics materially change |
| CMIS | connector mechanism only | target repository supports CMIS well |
| Keycloak/external IdP | no now | SSO/federation/real MFA/passkeys/tenant IdP |
| OpenFGA/SpiceDB | no now | arbitrary relationship graph/service split |
| Camunda/BPMN | no now | true generic process-engine requirement |
| Temporal | no now | durable orchestration outgrows economical DB jobs/outbox |
| Elasticsearch/OpenSearch | no requirement yet | measured search needs require it |
| EigenPal | preferred DOCX authoring candidate; provider/ACL only | DOCX authoring needs change |
| ONLYOFFICE/Office web | future authoring option | in-app XLSX/PPTX/full Office editing required |
| MinIO | V1/on-prem managed Artifact Store candidate | deployment choice |
| AWS S3 | first-class cloud managed Artifact Store target | cloud deployment |
| SharePoint normal | External Repository Connector | customer integration requirement |
| SharePoint Embedded | future Microsoft enterprise content profile | Microsoft-native storage/coauthoring/Purview strategy |
| PLM | integrate, do not build | BOM/CAD/configuration/change-management requirements |

---

# 11. Explicit target deletions / non-goals

No entitlement to survive:

- current `documents` / `controlleddocuments` / parallel `templates` target split;
- ControlledDocument duplicate identity, DocumentProfile, behavioral Family, GovernanceClass;
- parallel TemplateVersion lifecycle/MetadataSchema policy bundle/CompositionJSON without need;
- autosaves as business Revisions;
- BPMN/CEL/M-of-N/generic delegation/terminal reject;
- Approval ownership of document state/release/review;
- ReleaseGeneration required identity;
- universal mandatory FINAL_DOCX or OFFICIAL_PDF;
- live Group membership as historical distribution denominator;
- notification read as acknowledgement;
- live Dictionary references from historical content;
- Audit/Search as business authority;
- system_admin/old 8-role+38-capability/current RBAC DB bypass architecture;
- fake MFA / PlatformOperator as tenant role;
- generic confirmed upload bucket / orphan file library;
- user filename as canonical naming;
- “every file is Document” / “every Document is DOCX”;
- generic ECM/M-Files clone or low-code object engine;
- PLM/BOM/CAD dependency/change-management kernel;
- arbitrary multi-file Revision/Evidence payload V1;
- provider version = REV;
- content-addressed/cross-tenant dedup V1;
- BYOS-per-tenant/multi-cloud routing/permanent dual-write V1;
- SharePoint normal as S3 replacement;
- silent external repository sync;
- application-layer encryption of every Artifact without proven need;
- previous R10-A topology as approved truth.

---

# 12. Remaining whole-product completion before technical architecture

- [x] R9.5-1 Content Model
- [x] R9.5-2 Storage / Repository Strategy
- [ ] **R9.5-3 Authoring / EigenPal — NEXT**
- [ ] R9.5-4 Dossier / Context
- [ ] R9.5-5 Retention / Records / Legal Hold
- [ ] R9.5-6 Import / Migration / Export
- [ ] R9.5-7 Attestation + Content Security
- [ ] R9.5-8 Whole-product adversarial freeze

## R9.5-3 must close

1. what is authoritative working state while a REV is DRAFT;
2. relationship between working snapshots and confirmed Artifacts;
3. in-app DOCX editing vs external replacement/download-edit-upload;
4. autosave semantics, conflict detection and optimistic concurrency;
5. whether V1 permits one active writer, concurrent non-realtime writers, or realtime collaboration;
6. editor session/presence semantics;
7. tracked changes: governed content vs temporary review markup;
8. comments/annotations vs Approval rationale/evidence;
9. return-for-changes and reviewer suggestions;
10. crash/offline/recovery behavior;
11. EigenPal anti-corruption/provider seam and upstream-version strategy;
12. future editor providers without changing Document/REV/Submission truth.

Then R9.5-4 Dossier/Context.

---

# 13. Technical-design queue after R9.5

R10: bounded contexts/dependency DAG/aggregate ownership/data model/table ownership/constraints/transactions/events/jobs/storage implementation/current-module & current-table migration map.

R11: APIs/OpenAPI/DTOs/problem semantics/frontend IA and complete journeys.

R12: proof/Golden Matrix/threat/invariant tests + final durable ADR/spec/wiki promotion + adversarial review.

R13: implementation specification + sequenced implementation plan, then code.

Until all remaining gates close: **NO PRODUCT IMPLEMENTATION.**

---

# 14. Exact next step

Continue **R9.5-3 — Authoring / EigenPal**.

Preserve:

```text
Document != editor
REV != autosave
Artifact != storage location
DRAFT may evolve; Submission is immutable
CAPTURED Evidence is not editor-authored lifecycle
EigenPal is a DOCX provider, not domain authority
```

First close the authoritative DRAFT working-content model and concurrency before deciding collaboration features.