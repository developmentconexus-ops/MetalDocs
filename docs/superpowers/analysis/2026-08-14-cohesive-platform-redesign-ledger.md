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

Final authorization equation:

```text
Permission
+ required case/resource relationship
+ Domain Governance constraints
= ALLOW
```

RLS = tenant-isolation defense-in-depth only. DB constraints = structural invariants only. Current `system_admin` capability short-circuit / asserted-capability GUC authorization model is not target architecture.

---

# 3. LOCKED — Final Permission/role model (R9)

29 tenant Permissions:

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

Relationship-authorized narrow operations do **not** become fake Permissions: own notifications/sessions/password, Approval exact case, Distribution exact assignment/ack, submitter withdrawal, system release/rendition/erasure.

Approval SoD V1:

1. actor cannot accept if Revision creator or Submission submitter;
2. same user cannot accept two distinct Steps of one ApprovalInstance;
3. reassignment target must remain active/qualified and satisfy SoD.

Last active tenant owner cannot be revoked/deactivated. Responsible owner of periodic-review-governed Documents must be reassigned before deactivation.

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

**R9.5 refinement:** TemplateSpec/structured-authoring state is conditional capability for formats/authoring modes that need it; it is not the universal definition of Revision content.

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

Human outcomes: `accept | return_for_changes`. Separate admin/lifecycle ops: `withdraw | cancel | reassign`. No normal terminal reject V1.

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

### REOPENED BY MATERIAL R9.5 COUNTEREXAMPLE

The prior universal rule **“OFFICIAL_PDF mandatory for every Release” is no longer locked**. Format-agnostic Documents (XLSX, SVG, CAD, native PDF, etc.) prove that PDF cannot be a universal semantic requirement.

Current direction to close in R9.5-1:

> Submission always freezes exact primary source Artifact. Required official/viewable Renditions are determined by content/document policy, not by a universal PDF rule.

Examples under consideration:

```text
DOCX → source DOCX + required official/viewable PDF
PDF  → source PDF may itself be official representation
XLSX → source XLSX; PDF may be optional/viewable
SVG  → source SVG; preview may be PNG, not necessarily PDF
CAD  → native source; optional provider-specific viewable rendition
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

**R9.5 retention refinement pending:** terminal erasure must respect legal-retention/hold obligations; retention block will define what is legally eligible for deletion vs retained/anonymized.

Backup/restore must reapply erasure tombstones before service availability.

---

# 8. LOCKED — R9.5 Whole-Product North Star

R9.5 was added because storage, editor, non-DOCX content and enterprise-context use cases can materially alter architecture and therefore must be closed **before** bounded contexts/filesystem/data model.

The previously proposed R10-A topology is **not approved** and is paused.

## R9.5-NS01 — MetalDocs product boundary

> **MetalDocs is the system of record for identity, governance, revision, evidence and documentary context. Physical storage, authoring/editor technology, viewers and upstream enterprise systems are replaceable providers/connectors around that kernel.**

Do not replace MetalDocs with M-Files/Nuxeo/Alfresco, and do not build a generic M-Files clone.

Build ourselves only what constitutes the governance product: controlled identity/revision/submission/approval/release/evidence/context/provenance/distribution/audit semantics.

Use specialist technologies for object storage, Office editing, CAD/PLM, ERP, malware scanning, etc.

## R9.5-NS02 — Format-agnostic controlled documents

A governed Document is **not a DOCX**. It may have native source content such as:

```text
DOCX
PDF
XLSX
SVG/PNG
DWG/STEP
XML
other allowed formats
```

The governing question is whether the information needs stable official identity + `REVxxx` lifecycle, not which MIME type it has.

A Revision has one **primary source content** in V1. Rich multi-file packages are not assumed. Independent governed items (e.g. mechanical drawing, electrical drawing, manual) should normally be separate Documents related through context rather than arbitrary children of one REV.

## R9.5-NS03 — Artifact is technical content identity, never a user business object

Conceptually:

```text
Artifact
  immutable bytes/content identity
  hash
  media type
  size
  provenance/location handled below domain boundary
```

Artifact is referenced by business objects. Users do **not** create/browse orphan Artifacts as library items.

### NO ORPHAN CONTENT invariant

An upload/authoring operation always occurs in the context of a previously registered semantic object:

```text
Document:
  create Document + REV
  → then author/upload primary Artifact

Evidence:
  create typed Evidence record
  → then capture/upload Artifact
```

No “upload now, classify later” generic bucket V1.

## R9.5-NS04 — Evidence is registered semantic information, not a loose attachment

Evidence represents captured evidence of a fact/process and is distinct from a change-controlled Document.

Examples:

```text
Nota Fiscal
XML NF-e
Comprovante de Entrega
Foto de Inspeção
Certificado de Teste
Documento enviado pelo cliente
```

Evidence must have a known semantic type before content capture/upload. Current direction for R9.5-1 is an explicit `EvidenceType` (or equivalent closed/tenant-managed classification authority) plus `Evidence` instance; exact configurability/cardinality/lifecycle still open.

A transaction receipt/evidence is normally captured, not edited through `REVxxx` like an instruction/procedure. If an item itself needs controlled revisioning, model it as Document instead.

## R9.5-NS05 — Dossier = documentary context, not ERP/PLM replacement

Introduce a deliberately small future-facing concept `Dossier` for the documentary context surrounding an external/local business subject.

Examples:

```text
Venda 889949
Produto MTR-400
Projeto Hotel Alpha
Equipamento XYZ
Customer/Case/etc.
```

A Dossier may relate Controlled Documents and Evidence and carry opaque external references to ERP/PLM/etc.

MetalDocs does **not** become authority for ERP financial/fiscal/stock data or PLM BOM/configuration/CAD-dependency semantics.

Boundary trigger: when requirements become BOM, where-used, EBOM/MBOM, CAD dependency, ECR/ECO/ECN, part configuration etc., integrate with a PLM rather than growing the MetalDocs kernel.

No generic custom-object low-code platform, graph database, custom lifecycle/formula/ACL engine for Dossiers V1.

## R9.5-NS06 — Authoring is a provider capability

EigenPal remains a strong DOCX in-app editor candidate, but Document/Revision identity does not depend on EigenPal.

The same DRAFT Revision may eventually support in-app authoring or external-file replacement subject to format/content policy.

Future editor providers (ONLYOFFICE/Office integration/etc.) must not require redesigning Document/REV/Submission.

Autosave/concurrency/tracked-changes/comments/collaboration semantics still require explicit R9.5-3 design.

## R9.5-NS07 — Storage classes

Distinguish:

1. **Managed Artifact Store** — MetalDocs logically owns content; S3-compatible adapters such as MinIO/AWS S3 are natural providers.
2. **External Repository Connector** — SharePoint/OneDrive/Google Drive/CMIS repository etc.; operations such as import-copy/publish-copy/reference are explicit; no silent two-way mutation of an EFFECTIVE REV.
3. **Future platform profile** — SharePoint Embedded/M365 may become an enterprise content-backend/Office integration profile but is not forced into the ordinary S3 ArtifactStore abstraction.

Storage-provider version IDs are never business `REVxxx`.

## R9.5-NS08 — External repository edits never mutate released truth silently

If an effective MetalDocs item is published/copied/referenced externally and external bytes change, existing MetalDocs REV never changes in place.

Possible future explicit flow:

```text
external version changed
→ surface drift/new external version
→ authorized import/adopt action
→ new MetalDocs REV DRAFT
```

No silent bidirectional sync V1.

## R9.5-NS09 — Retention semantics belong to MetalDocs governance

S3 Object Lock / Purview / repository retention may provide provider enforcement, but the business reason/state/policy must be MetalDocs semantic truth if MetalDocs owns the record.

Retention/Legal Hold is still an open whole-product block and must be designed before final tenant erasure/storage architecture.

## R9.5-NS10 — Freeze rule against infinite scope

A future idea reopens the governance kernel only if it provides a material counterexample that breaks identity, historical truth or a locked invariant.

Examples:

```text
AWS S3        → provider seam, does not reopen Document
SharePoint    → repository connector seam
XLSX editor   → authoring provider seam
Venda/Produto → Dossier context seam
BOM/ECO       → outside boundary; PLM integration
```

This rule exists specifically to permit future evolution without continually postponing a releasable architecture.

---

# 9. Build-vs-buy rulings

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
| AWS S3 | supported future managed storage target | cloud deployment |
| SharePoint normal | external repository connector candidate | customer repository integration requirement |
| SharePoint Embedded | future Microsoft-enterprise profile candidate | Microsoft-native storage/coauthoring/Purview strategy |
| PLM | integrate, do not build | BOM/CAD/configuration/change-management requirements |

---

# 10. Explicit target deletions / non-goals

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
- **universal mandatory OFFICIAL_PDF** — reopened by format-agnostic content;
- live Group membership as historical distribution denominator;
- notification read as acknowledgement;
- live Dictionary references from historical content;
- Audit/Search as business authority;
- system_admin/old 8-role+38-capability model/current RBAC DB bypass architecture;
- fake MFA;
- PlatformOperator as tenant role;
- generic upload bucket / orphan file library;
- “every file is Document”;
- “every Document is DOCX”;
- generic M-Files/Nuxeo clone;
- generic low-code business-object engine;
- PLM/BOM/CAD dependency/change-management kernel;
- silent external-repository two-way sync;
- current R10-A bounded-context/filesystem proposal as approved truth.

---

# 11. Remaining whole-product completion before technical architecture

## R9.5-1 — Content Model — **NEXT**

Close:

1. exact semantics of `Artifact`, `Document`, `DocumentRevision`, `Evidence`, `EvidenceType`, `Dossier` references;
2. one-primary-source invariant and whether any deliberate multi-file exception exists;
3. allowed-format/content-policy ownership (`DocumentType`, `EvidenceType`, etc.);
4. creation-before-upload invariant and upload/capture lifecycle;
5. whether Evidence is immutable capture, replaceable-before-finalization, or versioned under narrow conditions;
6. Evidence provenance and external references;
7. native source vs official/viewable Rendition requirements per format/type;
8. how format-independent `RevisionContent` and `submission_digest` are defined;
9. source download/view policy implications;
10. exact relationship of TemplateSpec/structured authoring to format-agnostic content.

Then:

### R9.5-2 — Storage / Repository Strategy

Managed Artifact Store (MinIO/AWS S3), provider capabilities, external repository connector semantics, SharePoint/Embedded profile boundary, immutable key/reference strategy.

### R9.5-3 — Authoring / EigenPal

Canonical working content, autosave/checkpoints, optimistic concurrency, external edit/upload, tracked changes/comments, collaboration, recovery, editor provider seam.

### R9.5-4 — Dossier / Context

DossierType, instances, Document/Evidence relations, ERP/PLM references, local/external origin and boundary against ERP/PLM replication.

### R9.5-5 — Retention / Records / Legal Hold

Retention authority, records, legal hold, physical-provider enforcement, tenant-erasure interaction.

### R9.5-6 — Import / Migration / Export

Legacy revisions/codes, imported evidence, external historical approvals, provenance and export packaging.

### R9.5-7 — Attestation + Content Security

Approval meaning statements, signature manifestation, malware/content validation, MIME/OOXML validation and safe download/view policy.

### R9.5-8 — Whole-product adversarial freeze

Only after this pass may the whole-product domain be marked closed and R10 technical architecture resume.

---

# 12. Technical-design queue after R9.5

R10: bounded contexts/dependency DAG/aggregate ownership/data model/table ownership/constraints/transactions/events/jobs/storage technical implementation/current-module and current-table migration map.

R11: APIs/OpenAPI/DTOs/problem semantics/frontend information architecture and complete journeys.

R12: proof/Golden Matrix/threat/invariant tests + final durable ADR/spec/wiki promotion + adversarial review.

R13: implementation specification + sequenced implementation plan, then code.

---

# 13. Implementation gate

- [x] Authentication / Organization / Authorization north star
- [x] Approval V1
- [x] R3–R5 controlled-information lifecycle/configuration/numbering/template foundations
- [x] R6 periodic review + release invariants (universal PDF requirement reopened only)
- [x] R7 Distribution/Values/Audit/Notifications/Search
- [x] R8 Tenant lifecycle/Security (retention interaction pending refinement)
- [x] R9 final authorization model
- [x] R9.5 whole-product north star / product boundary
- [ ] R9.5-1 Content Model
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

# 14. Exact next step

Continue **R9.5-1 — Content Model**.

Do not implement. Start from the newly locked invariants:

```text
no orphan uploads
format-agnostic Document
Evidence registered before capture/upload
Artifact is technical content identity
Dossier is context, not ERP/PLM replacement
Submission freezes exact source content independent of provider
```

The first design question is the exact semantic boundary and lifecycle of **Artifact vs DocumentRevision vs Evidence**.