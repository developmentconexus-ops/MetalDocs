# MetalDocs Cohesive Platform Redesign — Active Decision Ledger

> **Status:** ACTIVE WIP — operator-approved decisions below are binding; unresolved items are explicit.
> **Date:** 2026-08-14
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** `wiki/standards/root-cause-global-maximum-method.md`
> **Program authority:** `wiki/architecture/cohesive-platform-redesign.md`
> **Implementation gate:** **CLOSED — design/documentation only.**

Git history is the archive. Current code/schema/OpenAPI are migration evidence, not target authority.

Global maximum = **the smallest professional architecture that correctly models the domain, preserves invariants and exposes clean extension seams.**

```text
product/domain semantics
→ invariants + lifecycle
→ Organization/AuthZ/Approval integration
→ whole-product completion (R9.5)
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

Core invariants:

1. every business fact has one canonical authority;
2. provider/editor/repository technology never becomes business identity;
3. Approval, Rendition and Release bind the same exact immutable RevisionSubmission/digest;
4. no generic BPM/ReBAC/low-code/object-platform engine without proven need;
5. no confirmed binary content without semantic ownership by a governed business object;
6. editor/browser state is never the authoritative persisted DRAFT truth;
7. Dossier is documentary context, never a hidden ERP/PLM/custom-object platform.

---

# 1. LOCKED — Authentication / Organization / Authorization (R1–R2, R9)

Authentication owns local V1 credentials, activation, opaque sessions, lockout/revocation and fresh-auth assurance behind a future external-IdP seam. Real MFA/passkeys/SSO/SAML/federation trigger re-evaluation of Keycloak/external IdP. Current stub MFA has no target entitlement.

Organization:

```text
Tenant
Area
User
Group
GroupMembership
```

Groups are flat V1; Area is the one organizational truth reused by Document ownership, scoped authorization and Approval actor resolution.

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

RLS = tenant-isolation defense-in-depth only. DB constraints = structural invariants. Current `system_admin` short-circuit/asserted-capability-GUC model is not target architecture.

Final R9 tenant Permission Catalog for the R3–R9 operation set:

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
- `viewer`: `document.read_effective`;
- `author`: viewer + history/working/create/edit/comment/submit/periodic-review qualification;
- `approver`: viewer + `approval.act`, no blanket draft access;
- `area_manager`: author + revision cancel/obsolete/owner manage + approval act/oversee/reassign/cancel + distribution manage/oversee in Area;
- `tenant_owner`: all 29 through normal Authorizer; still obeys relationships, SoD, lifecycle, fresh-auth and tenant-operability constraints.

Approval SoD V1:
1. actor cannot accept if Revision creator or Submission submitter;
2. same user cannot accept two distinct Steps of one ApprovalInstance;
3. reassignment target must remain active/qualified and satisfy SoD.

Last active tenant owner cannot be revoked/deactivated. Responsible owner of review-governed Documents must be reassigned before deactivation.

**R9.5 delta rule:** new Evidence/Dossier operations receive Permissions only after all R9.5 operations close; operations first, authorization delta second.

---

# 2. LOCKED — Approval V1

Approval is a specialized sequential human workflow, not BPM.

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

Human outcomes: `accept | return_for_changes`. Separate operations: `withdraw | cancel | reassign`. No normal terminal reject V1.

Participants materialize when a Step activates and are snapshotted; action revalidates current qualification/SoD.

ApprovalInstance binds exactly one immutable `RevisionSubmission`. Return-for-changes/allowed withdrawal terminates that attempt and returns the **same REV** to DRAFT. Resubmission creates a new Submission and, when required, a new ApprovalInstance. After completed Approval, V1 never reopens submitted candidate bytes for editing.

---

# 3. LOCKED — Controlled Information foundations (R3–R5)

`DocumentType` replaces DocumentProfile and is tenant-scoped with immutable code, display fields, optional classification-only category and ACTIVE/INACTIVE lifecycle. GovernanceClass is deleted.

Approval configuration is explicit:

```text
NoHumanApproval
or UsePolicy(ApprovalPolicyID)
```

`Document` is stable governed identity. At most one EFFECTIVE + one open Revision V1.

Official business labels: `REV001`, `REV002`, ...

Revision states:

```text
DRAFT
SUBMITTED
EFFECTIVE
SUPERSEDED
OBSOLETE
CANCELLED
```

REV is a business change cycle; autosaves/checkpoints are technical working history and never consume REV numbers. REV numbers never reuse after cancellation. `REV002+` requires reason-for-change before first submit.

`RevisionSubmission` is immutable exact attempt identity and exists even under NoHumanApproval. Approval/Rendition/Release bind exact Submission/digest.

Document code is tenant-wide unique/immutable. Document type/Area/code are immutable V1. Numbering language:

```text
literals + {TYPE} + {AREA} + {SEQ}
sequence_scope: TYPE | TYPE_AREA
sequence_width: minimum zero padding
```

No year/month/custom metadata/formulas/scripts/resets. Normal Create has no manual-code override; legacy preservation belongs to explicit import/migration.

Template is a role of an ordinary governed Document; no parallel TemplateVersion lifecycle. `TemplateUse` is M:N and exact source effective REV is pinned at derived-document creation.

TemplateSpec owns structured authoring/fill contract only and is optional structured-authoring state for applicable content, **not** the universal definition of Revision content.

---

# 4. LOCKED — Periodic Review / Rendition / Release (R6 refined by R9.5-1)

Periodic Review belongs Controlled Information: `Disabled | Every(n months)`. Cadence starts from actual Effectivity; overdue does not invalidate content. Immutable `PeriodicReviewRecord` binds exact effective REV with `confirmed_current | change_required`. Completion requires responsible-owner relation + permission + exact REV still current.

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

Universal mandatory OFFICIAL_PDF is retired. Locked V1 replacement:

```text
OfficialRepresentationPolicy =
    SourceOnly
  | RequireRendition(ContentFormat)
```

At most one required derived rendition V1. Exact primary source Artifact is always frozen by Submission; any required rendition must derive from that exact Submission.

---

# 5. LOCKED — Distribution / Values / Audit / Notifications / Search (R7)

Distribution = controlled obligation/acknowledgement, not AuthZ or LMS. Release snapshots concrete users; later Group membership never rewrites historical denominator. Explicit immutable `AcknowledgementRecord` completes obligation; notification read/view/download never does.

System Value Catalog remains a small product-owned contract. Tenant Dictionary is mutable source data resolved/snapshotted when a **new REV** is created; same-REV return/resubmit does not silently re-resolve.

Domain evidence records remain authorities. AuditEvent is transversal timeline only. Critical governed mutation cannot report success without durable audit intent/event in the same commit boundary. Audit stays append-only, tamper-evident and exportable with explicit User/System actors.

Notifications are delivery projection only. Search is rebuildable/eventually-consistent discovery projection and never grants canonical access. No Elasticsearch/OpenSearch requirement yet.

---

# 6. LOCKED — Tenant lifecycle / Platform Security (R8)

`PlatformOperator` / `SystemPrincipal` are outside tenant RBAC and gain no implicit tenant-content access.

Tenant lifecycle:

```text
ACTIVE
SUSPENDED
ERASED
```

Deletion request is a separate grace/cancel process. Onboarding creates Tenant + initial User + `tenant_owner @ Tenant` + single-use activation credential; platform operator never chooses tenant-owner password.

Suspension revokes sessions, blocks login/business mutations and is reversible. Tenant export/deletion request require fresh-auth.

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

Audit Trail itself is not deleted. Platform KEK wraps per-tenant DEK; no per-document key hierarchy/rotation/HSM product V1. R9.5-5 Retention must refine legal-retention/hold interaction before erasure mechanics freeze.

---

# 7. LOCKED — R9.5 Whole-Product North Star

R9.5 exists because storage, editor, non-DOCX content and enterprise-context use cases materially affect architecture. The previous R10-A topology is **not approved** and remains paused.

> **MetalDocs is the system of record for identity, governance, revision, evidence and documentary context. Physical storage, authoring/editor technology, viewers and upstream ERP/PLM/repositories are replaceable providers/connectors around that kernel.**

Do not replace MetalDocs with M-Files/Nuxeo/Alfresco and do not build a generic ECM/M-Files clone.

A governed Document is format-agnostic. `Dossier` is a deliberately small documentary context; BOM/where-used/EBOM/MBOM/CAD dependency/ECR/ECO are explicit PLM integration triggers.

EigenPal is a DOCX authoring provider, never Document identity.

Storage classes:
1. Managed Artifact Store — MetalDocs logically owns exact bytes.
2. External Repository Connector — SharePoint/OneDrive/etc. imports/publishes copies explicitly.
3. Future enterprise content profile — e.g. SharePoint Embedded/M365, designed explicitly rather than forced into S3 semantics.

Provider WORM/Purview may enforce retention physically; MetalDocs owns business retention/hold semantics if it owns the record.

Freeze rule: future ideas reopen the governance kernel only when they create a material identity/historical-truth/invariant counterexample. Otherwise they land on an existing provider/connector/context seam.

---

# 8. LOCKED — R9.5-1 Content Model

`Artifact` = immutable technical identity of exact bytes. It is not user-facing business identity and carries canonical SHA-256, size, ContentFormat/media type and technical provenance. Any byte change creates a new Artifact. Provider location/version/URL never enters Artifact business identity or Submission digest.

There is no confirmed orphan Artifact library. A UX may begin with temporary file staging inside a known Dossier, then classify and confirm semantic ownership. Staging is non-business and garbage-collectable.

Invariant:

```text
CONFIRMED Artifact
  must belong to
    DocumentRevision
    or Evidence
```

Tenant-scoped `EvidenceType` V1 has stable code/name/description/status, allowed formats and a small naming policy. It is not a custom schema/workflow engine.

Evidence canonical filename is generated by MetalDocs; original filename is provenance only. Naming tokens V1:

```text
{TYPE}
{DOSSIER}
{REF}
{SEQ}
```

`Evidence.reference?` is a narrow optional business reference, not metadata bag.

Evidence lifecycle:

```text
DRAFT
→ CAPTURED
→ VOIDED   // only invalid MetalDocs capture
```

DRAFT content may be replaced. CAPTURED content/metadata is immutable. Wrong capture is VOIDED with reason and replaced by new Evidence; external-world cancellation is a separate fact. Evidence does not use REV/Approval/Release by default; information needing revision governance should be Document.

Exactly one primary Artifact per DocumentRevision and per Evidence V1. True indivisible multi-file packages are future `ArtifactPackage`/PLM triggers.

MetalDocs owns a closed `ContentFormat` catalog. DocumentType/EvidenceType select allowed formats.

```text
RevisionContent
  primary_artifact
  governed_metadata
  structured_authoring?   // optional
```

Submission digest binds exact Artifact hash + governed state + decision-relevant structured/template provenance, never storage location.

---

# 9. LOCKED — R9.5-2 Storage / Repository Strategy

Artifact identity/hash and Submission digest are provider-independent. Canonical content hash = **SHA-256 of exact primary bytes**.

V1 has one active Managed Artifact Store per deployment, not per-tenant/per-document routing. First-class adapters:

```text
Local      // dev/test
MinIO
AWS S3
```

Other S3-compatible products require conformance validation before official support.

Provider migration = copy exact bytes + verify canonical hash + cut over physical location. It does not create new Artifact/REV/Submission. No permanent dual-write/active-active V1.

Managed keys are opaque, immutable and tenant-namespaced. Business/canonical filename never determines storage path. Artifact ID != content hash; no content-addressed or cross-tenant dedup V1.

Temporary staging/direct-presigned upload is allowed. Provider success alone does not confirm Artifact; confirmation requires integrity + content + semantic validation. Existing Artifact keys are never overwritten.

Object-store versioning is defense-in-depth, never `REVxxx`. Object Lock/WORM/legal-hold support is optional physical enforcement consumed by future MetalDocs Retention governance, never business authority.

Production baseline = encrypted transport + provider encryption at rest. Do not encrypt every Artifact with Tenant DEK V1; application Artifact crypto-shred is introduced only if Retention/Erasure proves a concrete need.

Normal SharePoint/OneDrive/etc. are External Repository Connectors, not ManagedArtifactStore providers. Governed primary content V1 requires an exact MetalDocs-managed copy. Connector directions begin with `IMPORT_COPY` / `PUBLISH_COPY`.

External edits never mutate existing EFFECTIVE REV/CAPTURED Evidence. Future adoption imports exact new bytes and creates new governed state/new DRAFT REV where applicable.

SharePoint Embedded is reserved as a future Microsoft-enterprise content-backend/coauthoring/Purview profile, not forced into S3 semantics.

Valid restore = Artifact DB fact + exact bytes + matching SHA-256. Tenant erasure/GC use MetalDocs inventory as semantic authority and tenant-namespaced storage as safety boundary.

---

# 10. LOCKED — R9.5-3 Authoring / EigenPal

Editor/browser state is never business authority. While a Revision is DRAFT, latest **persisted WorkingContent** is recoverable server truth.

DRAFT carries a monotonic technical `working_version`; `WorkingSnapshot` is an immutable technical snapshot and never a business REV. Old technical snapshots may be GC'd under recovery policy unless current or needed by later immutable evidence.

Any governed DRAFT change — primary bytes, structured values or governed Revision metadata — participates in the same working concurrency version. Editorial discussion that is not submitted content does not advance content version merely because comment state changed.

Every save/replacement requires `expected_working_version`; mismatch fails closed. No last-write-wins.

V1 uses one active in-app writer per DRAFT Revision **plus OCC**. `EditorSession` is a narrow heartbeat/staleness authoring lease, not generic distributed lock.

External download/edit/upload holds no long checkout. Replacement is based on known working_version and fails on stale base. No automatic binary DOCX merge V1.

In-app editing and external-file replacement modify the same DRAFT REV; authoring provider is not business identity. External replacement must satisfy content policy and structured-authoring parity.

Submit requires/follows final successful flush, validates OCC and freezes exact persisted logical state into immutable `RevisionSubmission`. After SUBMITTED, stale autosaves/replacements are rejected.

Approval UI is read-only over exact Submission. Current runtime `review → suggesting + autosave` has no target entitlement. Approval rationale/return reason is Approval evidence, not editor state.

`EditorialComment` is MetalDocs DRAFT collaboration state, not vendor DOCX-comment authority. V1 requires editorial comments resolved before submit. Tracked changes are optional provider markup; if enabled, unresolved tracked changes block submission. They are not mandatory V1 kernel requirements.

Browser/local recovery may supplement server snapshots but never overwrite a newer working_version silently.

Realtime Yjs/coauthoring is deferred V1. Future collaboration state remains authoring infrastructure and exports exact WorkingSnapshots; it never becomes REV/Submission identity.

Preserve one MetalDocs EigenPal anti-corruption/provider adapter. Pin exact editor version and require a MetalDocs fidelity/conformance corpus before upgrades. Future Office/ONLYOFFICE providers must not change Document/REV/Submission semantics.

---

# 11. LOCKED — R9.5-4 Dossier / Context

## DS-01 — Dossier meaning

`Dossier` is a **stable documentary context for an identifiable business subject**, not the ERP/PLM object itself and not a physical folder.

Examples:

```text
Venda 889949
Produto MTR-400
Projeto Hotel Alpha
Equipamento COL-0042
```

It groups documentary relationships while Document/Evidence retain independent identities/lifecycles.

## DS-02 — DossierType is deliberately small

Tenant-scoped `DossierType` V1:

```text
code
name
description?
status: ACTIVE | INACTIVE
eligible DocumentTypes
eligible EvidenceTypes
```

No custom fields, form builder, formulas, workflow, custom ACL, custom lifecycle or generic relation schema.

Eligibility is real validation, not just UX ranking. No required-evidence/completeness checklist V1.

## DS-03 — Stable key

Dossier has a stable human/business `key`, unique within `(tenant, DossierType)`. `title` may change; `key` does not V1.

`{DOSSIER}` in Evidence naming resolves this stable key, never mutable title.

V1 does not add a generic Dossier numbering engine. User/integration supplies the key. Add local numbering only when a proven recurring requirement appears.

## DS-04 — Creation provenance vs external identity

Do not model `origin = LOCAL | EXTERNAL` because a local Dossier may later acquire external identities.

Preserve creation provenance separately and allow zero..N `ExternalReference`s.

Conceptually:

```text
ExternalReference
  connection
  entity_kind
  external_id
  display_reference?
```

The same external identity is unique within tenant/connection/entity kind and cannot silently point to two Dossiers.

An integration may attach an external reference to an existing local Dossier when correlation is known. No heuristic auto-merge based on name/date/value similarity; ambiguity fails closed and requires explicit resolution.

## DS-05 — External master data stays external

External source status/fields never become canonical Dossier lifecycle automatically.

ERP/PLM/CRM fields may appear through read projections/cache, but Dossier canonical metadata stays small:

```text
type
key
title
scope
creation provenance
external references
relationships
```

If the external source becomes unavailable/deletes its object, Dossier/Documents/Evidence survive; reference/projection becomes stale/unavailable rather than deleting history.

## DS-06 — Document relationships

Dossier relates stable `Document` identity, not one specific Revision by default.

`Dossier ↔ Document` is M:N. One Document may serve multiple Dossiers without copying content.

A relation never changes Document type/Area/lifecycle/AuthZ and never grants access. Authorizer still filters each related item; inaccessible items need not be revealed.

Exact-REV usage evidence is a different future/explicit concept and is not overloaded into the normal Dossier→Document link.

## DS-07 — Evidence primary context

Every CAPTURED Evidence has exactly one immutable `primary_dossier`; while Evidence is DRAFT the primary context may be corrected.

Primary Dossier provides Evidence's naming/context/scope authority. Evidence may also have secondary relationships to other Dossiers without duplicating Evidence or Artifact.

Changing the primary Dossier after capture would rewrite historical context; correct by voiding/recreating Evidence instead.

## DS-08 — Scope

Dossier uses exactly one typed scope V1:

```text
TenantScope
or AreaScope
```

No multi-area ACL/scope list. Evidence reuses primary Dossier scope V1. Dossier type/key/scope are stable V1.

Dossier relationship never grants Document access and does not edit RoleAssignments.

## DS-09 — Lifecycle

Dossier lifecycle is intentionally tiny:

```text
ACTIVE ↔ ARCHIVED
```

ARCHIVED is reversible MetalDocs organization/navigation state, not Sale/Product/Project/Equipment business status.

Archiving does not delete Evidence, obsolete Documents or mutate external systems.

## DS-10 — Relationship history

Document/Evidence linkage/unlinkage must preserve historical/audit facts (`linked_at/by`, later unlink/reason as appropriate). Unlink never means the relationship never existed.

Dossier-to-Dossier hierarchy/graph relationships do **not** enter V1. Future repeated needs may justify a bounded relation model later.

## DS-11 — Search / timeline

Dossier is a first-class discovery object. Search and activity timeline are projections over canonical Dossier/Document/Evidence/external-reference facts and events; they are never authorities.

## DS-12 — Explicit product boundaries

ERP/CRM boundary: sale/order/customer calculations, financial/fiscal/stock/business-process state stay in ERP/CRM; MetalDocs owns documentary context, governed Documents/Evidence and optional projections.

PLM boundary: product documentation fits Dossier; BOM/part structure/where-used/EBOM/MBOM/CAD dependencies/ECR/ECO/ECN are PLM integration territory.

Project-management boundary: documents/evidence fit Dossier; Gantt/resource planning/critical path/budget/timesheets belong project-management systems.

EAM/CMMS boundary: equipment documents/evidence fit Dossier; work orders/preventive scheduling/spares/MTBF/downtime belong EAM/CMMS.

---

# 12. Build-vs-buy rulings

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
| Yjs/realtime collaboration | defer V1; seam only | proven simultaneous-authoring requirement |
| MinIO | V1/on-prem Managed Artifact Store candidate | deployment choice |
| AWS S3 | first-class cloud Managed Artifact Store target | cloud deployment |
| SharePoint normal | External Repository Connector | customer integration requirement |
| SharePoint Embedded | future Microsoft enterprise content profile | Microsoft-native storage/coauthoring/Purview strategy |
| PLM | integrate, do not build | BOM/CAD/configuration/change-management requirements |

---

# 13. Explicit target deletions / non-goals

No target entitlement to survive:

- current `documents` / `controlleddocuments` / parallel `templates` split;
- duplicate ControlledDocument identity / DocumentProfile / behavioral Family / GovernanceClass;
- parallel TemplateVersion lifecycle / template metadata-policy bundle / CompositionJSON without requirement;
- autosaves as business Revisions;
- editor/browser state as DRAFT authority;
- last-write-wins working saves;
- long external-edit checkout locks;
- automatic binary DOCX merge V1;
- vendor editor identity persisted as Document/Revision truth;
- approver editing/autosaving submitted bytes;
- mandatory realtime collaboration/CRDT V1;
- BPMN/CEL/M-of-N/generic delegation/terminal reject;
- Approval ownership of document state/release/review;
- universal mandatory FINAL_DOCX or OFFICIAL_PDF;
- Audit/Search/Notifications as business authorities;
- system_admin/old 8-role+38-capability/current RBAC DB bypass architecture;
- generic confirmed upload bucket/orphan file library;
- user filename as canonical naming;
- “every file is Document” / “every Document is DOCX”;
- generic ECM/M-Files clone / low-code object engine;
- Dossier custom-field/form/workflow/ACL engine;
- Dossier-to-Dossier graph/hierarchy V1;
- Dossier status mirroring ERP/PLM business lifecycle;
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

# 14. Remaining whole-product completion before technical architecture

- [x] R9.5-1 Content Model
- [x] R9.5-2 Storage / Repository Strategy
- [x] R9.5-3 Authoring / EigenPal
- [x] R9.5-4 Dossier / Context
- [ ] **R9.5-5 Retention / Records / Legal Hold — NEXT**
- [ ] R9.5-6 Import / Migration / Export
- [ ] R9.5-7 Attestation + Content Security
- [ ] R9.5-8 Whole-product adversarial freeze

## R9.5-5 must close

1. whether retention applies to Document, Revision/Release, Evidence, Dossier, Audit and/or Artifact;
2. retention policy ownership/configuration and inheritance/snapshot rules;
3. retention start trigger (release, capture, archival, external event, explicit date) without generic rules engine;
4. disposition semantics after retention expiry: eligible-for-disposition vs automatic deletion;
5. legal hold scope, apply/release authority and auditability;
6. interaction with Document SUPERSEDED/OBSOLETE and Evidence VOIDED;
7. interaction with tenant deletion/erasure and backups;
8. provider enforcement mapping to S3 Object Lock / MinIO WORM / Purview without making provider authoritative;
9. immutable retention evidence and deletion/disposition records;
10. whether V1 needs formal record declaration or whether released Documents + CAPTURED Evidence already serve as governed records;
11. whether retention is tenant-wide/type-based/context-based and how to avoid a generic records-management engine;
12. exact data allowed to survive tenant erasure when legal hold/retention requires survival.

Then R9.5-6 Import / Migration / Export.

---

# 15. Technical-design queue after R9.5

R10: bounded contexts/dependency DAG/aggregate ownership/data model/table ownership/constraints/transactions/events/jobs/storage implementation/current-module & current-table migration map.

R11: APIs/OpenAPI/DTOs/problem semantics/frontend IA and complete journeys.

R12: proof/Golden Matrix/threat/invariant tests + final durable ADR/spec/wiki promotion + adversarial review.

R13: implementation specification + sequenced implementation plan, then code.

Until all remaining gates close: **NO PRODUCT IMPLEMENTATION.**

---

# 16. Exact next step

Continue **R9.5-5 — Retention / Records / Legal Hold**.

Preserve:

```text
business retention authority stays in MetalDocs
provider WORM/Purview = enforcement only
Document/Evidence keep their approved lifecycles
Dossier is context, not record container lifecycle
Tenant erasure cannot silently violate a legal hold/retention obligation
no generic records-management rules engine without proven need
```

Design the smallest retention/hold model that works for controlled Documents and captured Evidence, survives storage-provider changes, and reconciles legally required preservation with terminal tenant erasure.