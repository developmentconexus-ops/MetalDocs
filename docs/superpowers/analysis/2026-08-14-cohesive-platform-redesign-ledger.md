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

Design sequence:

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

---

# 1. Core target principles

1. Every business fact has one canonical authority.
2. `Document` is stable governed identity; `DocumentRevision` is a business change cycle; `RevisionSubmission` is an immutable submitted candidate.
3. Approval, Rendition and Release always bind the **same exact Submission/digest**.
4. Supporting concerns (Audit, Notifications, Search) never become competing business-state authorities.
5. Current package/module boundaries have no entitlement to survive.
6. No generic BPM/ReBAC/expression/low-code/object-platform engine without a concrete requirement.
7. Domain truth is provider-independent: storage/editor/search/IdP/ERP/PLM are replaceable capabilities around the governance kernel.
8. No confirmed binary content exists without semantic ownership by a governed business object.

Historical root defect: browser QA proved humans could review edited content while freeze/render later selected blank-template bytes. Target architecture must make that class impossible by construction.

---

# 2. LOCKED — Authentication / Organization / Authorization

## Authentication

- local AuthN remains V1 behind a future external-IdP seam;
- opaque persisted sessions, revocation, lockout, credential lifecycle and fresh-auth assurance remain valid responsibilities;
- future principal concept exposes `user_id`, `tenant_id`, auth time/method/assurance;
- real MFA/passkeys/SSO/SAML/federation trigger re-evaluation of Keycloak/external IdP before rebuilding IdP capabilities internally;
- current fake/stub MFA coverage has no target entitlement.

## Organization

```text
Tenant
Area
User
Group
GroupMembership
```

Groups flat V1. User may belong to multiple groups. Area is one organizational truth reused by document ownership, scoped authorization and approval actor resolution.

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

Authorization equation:

```text
Permission
+ required case/resource relationship
+ Domain Governance constraints
= ALLOW
```

RLS = tenant-isolation defense-in-depth only. DB constraints = structural invariants only. Current `system_admin` capability short-circuit / asserted-capability GUC model is not target architecture.

---

# 3. LOCKED — Final Permission/role model (R9)

29 tenant Permissions remain locked for the R3–R9 operation set:

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
- `tenant_owner`: all 29 Permissions through normal Authorizer; still obeys relations, SoD, lifecycle, fresh-auth and tenant-operability constraints.

Relationship-authorized narrow operations do not become fake Permissions: own notifications/sessions/password, Approval exact case, Distribution exact assignment/ack, submitter withdrawal, system release/rendition/erasure.

Approval SoD V1:

1. actor cannot accept if Revision creator or Submission submitter;
2. same user cannot accept two distinct Steps of one ApprovalInstance;
3. reassignment target must remain active/qualified and satisfy SoD.

Last active tenant owner cannot be revoked/deactivated. Responsible owner of periodic-review-governed Documents must be reassigned before deactivation.

**R9.5 authorization delta rule:** Evidence/Dossier operations are new product operations discovered after R9; permissions are derived only after those lifecycles close, then the catalog/Golden Matrix receives a bounded delta. Do not invent permissions before operations.

---

# 4. LOCKED — Controlled Information configuration + lifecycle (R3–R5)

## Configuration

`DocumentProfile` → `DocumentType`:

```text
id
 tenant_id
 code           // immutable
 name
 description?
 category_id?
 status: ACTIVE | INACTIVE
```

`DocumentTypeCategory` is classification/navigation only. `GovernanceClass` deleted.

Approval configuration explicit:

```text
NoHumanApproval
or
UsePolicy(ApprovalPolicyID)
```

Template is a role of an ordinary governed Document. No parallel Template/TemplateVersion lifecycle. `TemplateUse` M:N; at most one UX default/type; source current EFFECTIVE template REV resolves once at derived-document creation and exact origin is pinned forever.

## Document / Revision

Stable `Document` identity with at most one effective + one open Revision.

Official labels:

```text
REV001
REV002
REV003
...
```

States:

```text
DRAFT
SUBMITTED
EFFECTIVE
SUPERSEDED
OBSOLETE
CANCELLED
```

Autosaves/checkpoints are technical working history, not business Revisions.

REV allocated when change cycle starts and never reused. `REV002+` requires reason-for-change before submission.

## RevisionSubmission

Immutable exact attempt:

```text
REV002
  Submission #1 → digest A → returned
  Submission #2 → digest B → released
```

Return/allowed withdrawal closes old attempt and returns **same REV** to DRAFT. Resubmission creates new Submission and, when required, new ApprovalInstance. After completed Approval V1 does not reopen candidate content: cancel candidate and create a new REV if content must change.

`SUPERSEDED` = newer Revision of same Document became effective. `OBSOLETE` = Document retired without successor, terminal V1.

## Numbering

`Document.code` immutable tenant-wide identity. Type/Area/code immutable V1.

DocumentType numbering language V1:

```text
literals + {TYPE} + {AREA} + {SEQ}
sequence_scope: TYPE | TYPE_AREA
sequence_width: minimum zero padding
```

No year/month/custom-fields/formulas/scripts/resets. Normal Create has no manual code override; legacy preservation belongs to explicit import/migration.

## TemplateSpec

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

**R9.5 refinement:** TemplateSpec/structured-authoring state is conditional capability for applicable content, not the universal definition of Revision content.

---

# 5. LOCKED — Approval / Periodic Review / Release (R4–R6)

## Approval

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

ApprovalInstance binds one exact RevisionSubmission. Participants materialize when Step activates and are snapshotted; action revalidates current qualification/SoD.

## Periodic Review

Owned by Controlled Information, not Approval:

```text
Disabled
or Every(n months)
```

Cadence from actual Effectivity; overdue does not invalidate content. Immutable `PeriodicReviewRecord` against exact effective REV with `confirmed_current | change_required`. `change_required` does not auto-create a REV. Review requires responsible-owner relation + `document.review_periodic` + exact REV still current.

## Rendition / Release

Rendition = immutable derived representation of exact Submission with own output hash + generator/build provenance.

Approval approves Submission, never output bytes. Release is automatic/system-owned; no human publish button. Optional `ReleasePlan.not_before`; actual `effective_at = released_at`.

Winning release transaction atomically:

```text
candidate REV -> EFFECTIVE
prior REV -> SUPERSEDED
Document.effective_revision_id -> candidate
Document.open_revision_id -> null
ReleaseRecord
outbox/events
```

Candidate may be cancelled after Approval and before Release; evidence remains historical.

### R6-12 REOPENED AND REPLACED BY R9.5-1

The prior universal rule **“OFFICIAL_PDF mandatory for every Release” is retired** because format-agnostic Documents provide concrete counterexamples (XLSX, SVG, CAD, native PDF).

Locked replacement:

```text
OfficialRepresentationPolicy =
    SourceOnly
  | RequireRendition(ContentFormat)
```

V1 supports at most one required derived rendition. The exact primary source Artifact is always frozen by Submission. Required rendition, when configured, must derive from that exact Submission. Rich multi-rendition requirement sets are deferred until proven.

Examples:

```text
DOCX → source DOCX + RequireRendition(PDF)
PDF  → SourceOnly
XLSX → SourceOnly (PDF may be optional/viewable)
SVG  → SourceOnly (PNG preview may be optional)
CAD  → source native; optional/required viewable only if type policy says so
```

All other R6 Release invariants remain locked.

---

# 6. LOCKED — Distribution / Values / Audit / Notifications / Search (R7)

## Distribution

Controlled obligation/acknowledgement, not AuthZ or Training/LMS.

Document configuration:

```text
None
or ReadAcknowledgement {
  targets: User | Group
  due_in_days?
  requires_reauthentication
}
```

Release snapshots concrete users. Later Group membership never rewrites historical denominator. Explicit `AcknowledgementRecord` completes obligation; notification read/view/download never does. New effective REV supersedes pending old-REV assignments and materializes a new cohort. Distribution never edits RoleAssignments.

## Values

Product System Value Catalog V1:

```text
document_code
revision_label
revision_title
document_type_code
document_area_code
document_area_name
revision_created_by_name
```

Tenant Dictionary is mutable source data; referenced values resolve/snapshot when a **new REV** is created. Same-REV return/resubmit does not silently re-resolve. Historical content never uses live dictionary values.

## Audit

Domain evidence stays authoritative (`RevisionSubmission`, `ApprovalDecision`, `PeriodicReviewRecord`, `ReleaseRecord`, `DistributionAssignment`, `AcknowledgementRecord`, RoleAssignment history...). AuditEvent is transversal timeline only.

Critical governed mutation cannot report success without durable audit intent/event in the same commit boundary. Usage telemetry may be async.

Audit Trail remains append-only, tamper-evident, exportable, with explicit User/System actors.

## Notifications / Search

Notifications are delivery projection; Notification READ only means notification read. “Minhas Pendências” queries business authorities.

Search is rebuildable/eventually-consistent discovery projection. Official Library = effective content; Working Search = open content under current authorization. Stale search never grants access; canonical endpoint rechecks AuthZ. No Elasticsearch/OpenSearch requirement yet.

---

# 7. LOCKED — Tenant lifecycle / Platform Security (R8)

PlatformOperator/SystemPrincipal exist outside tenant RBAC and gain no implicit tenant-content access.

Tenant lifecycle:

```text
ACTIVE
SUSPENDED
ERASED
```

Deletion request separate: `PENDING | CANCELLED | EXECUTED` with grace period. Tenant remains active until execution and may cancel.

Onboarding creates Tenant + initial User + `tenant_owner @ Tenant` + single-use time-limited activation credential. Platform operator never chooses tenant-owner password.

Suspension revokes sessions, blocks login/business mutations and is reversible. Business jobs respect suspension; lifecycle/security jobs may continue.

Tenant export and deletion request are same-tenant owner operations requiring fresh-auth.

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

Audit Trail itself is not deleted. Sensitive retained payload must be erasable/unreadable. Platform KEK wraps per-tenant DEK; no per-document key hierarchy/rotation/HSM product V1.

**Retention refinement pending:** terminal erasure must respect legal-retention/hold obligations; retention block defines what is legally eligible for deletion vs retained/anonymized.

Backup/restore must reapply erasure tombstones before service availability.

---

# 8. LOCKED — R9.5 Whole-Product North Star

R9.5 exists because storage, editor, non-DOCX content and enterprise-context use cases materially affect architecture and therefore must close **before** bounded contexts/filesystem/data model.

The previously proposed R10-A topology is **not approved** and remains paused.

## Product boundary

> **MetalDocs is the system of record for identity, governance, revision, evidence and documentary context. Physical storage, authoring/editor technology, viewers and upstream ERP/PLM/repositories are replaceable providers/connectors around that kernel.**

Do not replace MetalDocs with M-Files/Nuxeo/Alfresco and do not build a generic M-Files clone.

Build ourselves only what constitutes the governance product. Use specialist technologies for object storage, Office editing, CAD/PLM, ERP, malware scanning, etc.

## Format-agnostic controlled documents

A governed Document is **not a DOCX**. It may use native source content such as DOCX/PDF/XLSX/SVG/PNG/CAD/XML/etc. subject to type/content policy.

The governing question is whether information needs stable official identity + `REVxxx` lifecycle, not MIME type.

## Dossier boundary

`Dossier` is a small future-facing documentary context for things such as Venda, Produto, Projeto, Equipamento, Customer/Case. It may relate Controlled Documents and Evidence and carry opaque external references to ERP/PLM.

It does not become ERP or PLM. BOM/where-used/EBOM/MBOM/CAD dependency/ECR/ECO/ECN are explicit PLM-integration boundary triggers. No generic custom-object low-code platform or graph engine V1.

## Authoring/provider boundary

EigenPal remains a strong DOCX in-app editor candidate, but Document/Revision identity does not depend on EigenPal. Future editor providers must not require redesigning Document/REV/Submission.

## Storage classes

Distinguish:

1. **Managed Artifact Store** — MetalDocs logically owns content; S3-compatible adapters such as MinIO/AWS S3 are natural providers.
2. **External Repository Connector** — SharePoint/OneDrive/Google Drive/CMIS repository; import-copy/publish-copy/reference are explicit; no silent two-way mutation of an EFFECTIVE REV.
3. **Future platform profile** — SharePoint Embedded/M365 may become a Microsoft-enterprise content-backend/Office integration profile but is not forced into the ordinary S3 ArtifactStore abstraction.

Storage-provider version IDs are never business `REVxxx`.

If externally published/referenced bytes change, existing EFFECTIVE MetalDocs REV never changes in place. Future explicit adoption/import creates a new DRAFT REV.

## Retention north star

Provider WORM/Purview may provide physical enforcement, but business retention/hold semantics belong to MetalDocs governance if MetalDocs owns the record.

## Freeze rule against infinite scope

A future idea reopens the governance kernel only if it creates a material counterexample that breaks identity, historical truth or a locked invariant. Provider/editor/connector/context requirements should land on existing seams rather than continually reopen the kernel.

---

# 9. LOCKED — R9.5-1 Content Model

## CM-01 — Artifact

`Artifact` is the immutable technical identity of exact bytes/content, not a user-facing business object.

Conceptually it carries at least identity, tenant, content hash, byte size, media/content format and technical provenance. Business meaning never lives on Artifact.

Any byte change creates a new Artifact. Filename equality does not imply content equality.

Artifact identity and any `RevisionSubmission.submission_digest` are **provider/location independent**. MinIO bucket, S3 key/version, SharePoint URL, presigned URL, server hostname, etc. never become business identity or digest inputs.

Storage migration therefore never creates a new business REV solely because bytes moved providers.

## CM-02 — Confirmed-content ownership; staging is allowed

There is no user-visible orphan Artifact library and no confirmed Artifact without semantic ownership.

However, a UX may begin with an upload **inside a known context such as a Dossier**:

```text
user drops file
→ temporary staging upload
→ detect ContentFormat
→ choose/confirm compatible EvidenceType
→ collect required semantic reference if any
→ allocate canonical evidence filename
→ create Evidence + confirm Artifact atomically/consistently
→ attach Evidence to Dossier
```

Temporary staging bytes are not confirmed Artifacts/business content and may be garbage-collected if classification/confirmation fails.

Document path may likewise create Document/REV first and then author/upload primary content.

Invariant:

```text
CONFIRMED Artifact
  must be owned by
    DocumentRevision
    or Evidence
```

No generic “upload and leave unclassified” bucket V1.

## CM-03 — EvidenceType

Evidence must become a semantic object, never a loose attachment.

Tenant-scoped `EvidenceType` V1 concept:

```text
id
 tenant_id
 code          // stable identifier
 name
 description?
 status: ACTIVE | INACTIVE
 allowed_formats[]
 naming_policy
```

Examples:

```text
NOTA_FISCAL
XML_NFE
COMPROVANTE_ENTREGA
FOTO_INSPECAO
CERTIFICADO_TESTE
DOC_CLIENTE
COMPROVANTE_PAGAMENTO
```

EvidenceType is configurable enough for different industries but is **not** a generic custom-field/schema/workflow engine V1.

## CM-04 — Evidence naming is system-governed

User-provided filenames are provenance only, never official naming authority.

Preserve both when useful:

```text
original_filename = "nota senhor osvaldo final.pdf"
canonical_filename = "NF-1001-889949-001.pdf"
```

EvidenceType owns a deliberately small naming language V1:

```text
{TYPE}
{DOSSIER}
{REF}
{SEQ}
```

Extension is derived from validated ContentFormat; sequence width is configured similarly to Document numbering.

`Evidence.reference?` is a narrow optional business reference for the Evidence itself (examples: invoice number, certificate number, romaneio number). It is not a generic metadata bag.

Do not add arbitrary customer/seller/date/formula/script tokens V1.

Human upload and future ERP/PLM integration use the **same naming semantics**: provider/source does not decide filename.

## CM-05 — Evidence classification UX

File format narrows valid EvidenceTypes but never silently determines business meaning.

Examples:

```text
PDF  → Nota Fiscal / Orçamento / Comprovante / Laudo / Documento Cliente...
XML  → XML NF-e / XML CT-e / other XML evidence types...
JPEG → Foto Entrega / Foto Inspeção / Documento Cliente...
```

Future AI may suggest a likely EvidenceType, but classification remains explicit unless a trusted integration supplies the semantic type by contract.

DossierType may later prioritize/restrict relevant EvidenceTypes; exact rule belongs to R9.5-4 Dossier design.

## CM-06 — Evidence lifecycle

Evidence does not use `REVxxx`.

Minimal V1 lifecycle:

```text
DRAFT
  ↓ capture
CAPTURED
  ↓ administrative invalidation when capture itself was wrong
VOIDED
```

While DRAFT, content may be replaced. CAPTURED Evidence is immutable in its governed content/metadata.

If captured incorrectly, preserve history: VOID prior Evidence with reason and create a new Evidence; do not rewrite the captured record.

`VOIDED` means the **MetalDocs capture is invalid**, not that the external-world subject was later cancelled. Example: an NF-e that was legitimately captured and later cancelled remains valid historical Evidence; cancellation is another domain fact/evidence, not retroactive invalidation of capture.

Evidence does not use Approval/Release by default. If information itself needs controlled revision/approval lifecycle, model it as Document.

Evidence capture is auditable and may carry User or System/Integration provenance.

## CM-07 — One primary Artifact V1

Exactly one primary Artifact per `DocumentRevision` and exactly one primary Artifact per `Evidence` V1.

Do not create arbitrary attachments collections where original/derived/corrected files become ambiguous.

Independent governed information should normally be separate Documents/Evidence related through Dossier/context.

True indivisible multi-file packages (for example some CAD assemblies) are an explicit future trigger for `ArtifactPackage` or, more likely, PLM/PDM integration. Do not contaminate V1 preemptively.

## CM-08 — Product ContentFormat catalog

MetalDocs owns a closed product `ContentFormat` catalog that normalizes accepted formats/media types/extensions and later provider capabilities.

Examples may include DOCX, PDF, XLSX, SVG, PNG, JPEG, XML, DWG, STEP, etc.

This is product capability metadata, not tenant scripts/plugins.

`DocumentType` and `EvidenceType` select allowed formats. Validation must never trust filename extension alone; content-security details are R9.5-7.

## CM-09 — Format-independent RevisionContent

Conceptual shape:

```text
RevisionContent
  primary_artifact
  governed_metadata
  structured_authoring?   // optional
```

For template-driven DOCX, `structured_authoring` may snapshot TemplateSpec/field values/value provenance. For externally uploaded XLSX/PDF/etc., it may be absent.

TemplateSpec remains fully useful without being universal content identity.

## CM-10 — Provider-independent Submission digest

`RevisionSubmission.submission_digest` binds a canonical representation of at least:

```text
Document/Revision identity needed for attestation
primary Artifact content hash
governed Revision metadata
structured-authoring state when present
template/source provenance when decision-relevant
```

It never binds storage location/provider identifiers.

Approval/Rendition/Release therefore remain stable across MinIO→S3 migration or other physical relocation.

## CM-11 — DocumentType content policy

DocumentType may select allowed source formats and official representation policy, but never names editor/storage technologies.

No `editor=EigenPal` or `storage=MinIO` in business configuration.

Official representation V1:

```text
SourceOnly
or RequireRendition(ContentFormat)
```

At most one required derived rendition V1.

## CM-12 — Integration semantics

Trusted integrations operate in domain terms, not generic file-drop terms.

Example Sankhya flow:

```text
Ensure/resolve Dossier Venda 889949
Create Evidence type=XML_NFE reference=<business reference> provenance=SANKHYA
capture exact Artifact
MetalDocs allocates canonical filename
```

This keeps human upload and automated capture semantically identical.

## CM-13 — Search/discovery consequence

Documents/Evidence/Dossiers are discoverable business objects. Artifact is not the primary discovery unit.

A search for a sale/product context should eventually return Dossier and its related Documents/Evidence, not a raw list of object-store keys.

---

# 10. Build-vs-buy rulings

| Technology/class | Ruling | Revisit trigger |
|---|---|---|
| M-Files/Nuxeo/Alfresco as core | do not adopt as MetalDocs kernel | only if product strategy changes to customization on another ECM |
| Jackrabbit/JCR as kernel | no | stack/domain economics materially change |
| CMIS | possible connector mechanism, not domain model | a target external repository supports CMIS well |
| Keycloak/external IdP | no now | SSO/federation/real MFA/passkeys/tenant IdP |
| OpenFGA/SpiceDB | no now | arbitrary relationship graph/service split |
| Camunda/BPMN | no now | true generic process-engine requirement |
| Temporal | no now | durable orchestration outgrows economical DB jobs/outbox |
| Elasticsearch/OpenSearch | no requirement yet | measured search needs require it |
| EigenPal | preferred DOCX authoring candidate; adapter/provider only | DOCX editor needs change |
| ONLYOFFICE/Office web | future authoring option, not V1 commitment | in-app XLSX/PPTX/full Office editing becomes requirement |
| MinIO | current managed storage provider candidate | deployment/provider choice |
| AWS S3 | future managed storage target | cloud deployment |
| SharePoint normal | external repository connector candidate | customer repository integration requirement |
| SharePoint Embedded | future Microsoft-enterprise profile candidate | Microsoft-native storage/coauthoring/Purview strategy |
| PLM | integrate, do not build | BOM/CAD/configuration/change-management requirements |

---

# 11. Explicit target deletions / non-goals

No entitlement to survive:

- current `documents` / `controlleddocuments` / parallel `templates` target split;
- ControlledDocument duplicate identity, DocumentProfile, behavioral Family, GovernanceClass;
- TemplateVersion parallel lifecycle, Template MetadataSchema policy bundle, CompositionJSON without requirement;
- user-visible technical versioning instead of `REVxxx`;
- autosaves as business Revisions;
- BPMN/CEL/M-of-N/generic delegation/terminal reject;
- Approval ownership of document state/release/periodic review;
- ReleaseGeneration required domain identity;
- universal mandatory FINAL_DOCX;
- universal mandatory OFFICIAL_PDF;
- live Group membership as historical distribution denominator;
- notification read as acknowledgement;
- live Dictionary references from historical content;
- Audit/Search as business authority;
- system_admin/old 8-role+38-capability model/current RBAC DB bypass architecture;
- fake MFA;
- PlatformOperator as tenant role;
- generic confirmed upload bucket / orphan file library;
- user filename as canonical business naming;
- “every file is Document”;
- “every Document is DOCX”;
- generic M-Files/Nuxeo clone;
- generic low-code business-object engine;
- PLM/BOM/CAD dependency/change-management kernel;
- silent external-repository two-way sync;
- arbitrary multi-file DocumentRevision/Evidence payload V1;
- current R10-A bounded-context/filesystem proposal as approved truth.

---

# 12. Remaining whole-product completion before technical architecture

- [x] **R9.5-1 Content Model**
- [ ] **R9.5-2 Storage / Repository Strategy — NEXT**
- [ ] R9.5-3 Authoring / EigenPal
- [ ] R9.5-4 Dossier / Context
- [ ] R9.5-5 Retention / Records / Legal Hold
- [ ] R9.5-6 Import / Migration / Export
- [ ] R9.5-7 Attestation + Content Security
- [ ] R9.5-8 Whole-product adversarial freeze

## R9.5-2 — Storage / Repository Strategy

Close:

1. Managed Artifact Store contract and what the domain is allowed to assume;
2. MinIO vs AWS S3 adapter semantics and portability subset;
3. whether Artifact physical location is one active location, replicated, or migration-aware;
4. immutable object key/reference strategy and deduplication stance;
5. upload staging/commit/garbage-collection mechanics conceptually;
6. provider capabilities: versioning, WORM/Object Lock, legal hold, presigned access, checksums;
7. encryption-at-rest responsibility and tenant DEK interaction;
8. external repository connector contract (SharePoint/OneDrive/Google Drive/CMIS);
9. import-copy vs publish-copy vs reference semantics;
10. external drift detection/adoption and prohibition on silent mutation of EFFECTIVE truth;
11. SharePoint Embedded boundary as optional enterprise profile;
12. backup/restore and tenant-erasure implications.

Then R9.5-3 Authoring/EigenPal.

---

# 13. Technical-design queue after R9.5

R10: bounded contexts/dependency DAG/aggregate ownership/data model/table ownership/constraints/transactions/events/jobs/storage technical implementation/current-module and current-table migration map.

R11: APIs/OpenAPI/DTOs/problem semantics/frontend information architecture and complete journeys.

R12: proof/Golden Matrix/threat/invariant tests + final durable ADR/spec/wiki promotion + adversarial review.

R13: implementation specification + sequenced implementation plan, then code.

---

# 14. Implementation gate

- [x] Authentication / Organization / Authorization north star
- [x] Approval V1
- [x] R3–R5 controlled-information lifecycle/configuration/numbering/template foundations
- [x] R6 periodic review + release invariants with R9.5 representation refinement
- [x] R7 Distribution/Values/Audit/Notifications/Search
- [x] R8 Tenant lifecycle/Security (retention interaction pending refinement)
- [x] R9 final authorization model
- [x] R9.5 whole-product north star / product boundary
- [x] R9.5-1 Content Model
- [ ] R9.5-2 Storage/Repositories
- [ ] R9.5-3 Authoring/Editor
- [ ] R9.5-4 Dossier/context
- [ ] R9.5-5 Retention/Legal Hold
- [ ] R9.5-6 Import/Migration/Export
- [ ] R9.5-7 Attestation/Content Security
- [ ] R9.5-8 adversarial whole-product freeze
- [ ] R10 technical architecture
- [ ] R11 API/frontend journeys
- [ ] R12 final proof/spec promotion/review
- [ ] operator approval of integrated code-ready design
- [ ] R13 implementation spec/plan

Until all remaining gates close: **NO PRODUCT IMPLEMENTATION.**

---

# 15. Exact next step

Continue **R9.5-2 — Storage / Repository Strategy**.

Do not implement. Preserve R9.5-1 truths:

```text
Artifact != storage location
confirmed Artifact always has semantic owner
staging bytes are temporary/non-business
canonical naming is generated by MetalDocs
storage provider/version never equals REVxxx
Submission digest never binds provider key/location
```

The first decision is the exact boundary between **Managed Artifact Store** and **External Repository Connector**, and what guarantees MetalDocs requires from a managed provider.