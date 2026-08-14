# MetalDocs Cohesive Platform Redesign — Active Decision Ledger

> **Status:** ACTIVE WIP — operator-approved decisions below are binding; unresolved items are explicit.
> **Date:** 2026-08-14
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** `wiki/standards/root-cause-global-maximum-method.md`
> **Program authority:** `wiki/architecture/cohesive-platform-redesign.md`
> **Implementation gate:** **CLOSED — design/documentation only.**

Git history is the archive for prior verbose ledger forms and superseded runtime/ADR assumptions. Current code/schema/OpenAPI are migration evidence, not target-design authority.

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
6. editor/browser state is never the authoritative persisted DRAFT truth.

---

# 1. LOCKED — Authentication / Organization / Authorization (R1–R2, R9)

Authentication owns local V1 credentials, activation, opaque sessions, lockout/revocation and fresh-auth assurance behind a future external-IdP seam. Real MFA/passkeys/SSO/SAML/federation trigger re-evaluation of Keycloak/external IdP before rebuilding IdP capabilities internally. Current stub MFA has no target entitlement.

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

RLS = tenant-isolation defense-in-depth only. DB constraints = structural invariants. Current `system_admin` short-circuit / asserted-capability-GUC model is not target architecture.

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

- `viewer`: `document.read_effective`.
- `author`: viewer + history/working/create/edit/comment/submit/periodic-review qualification.
- `approver`: viewer + `approval.act`; no blanket draft access.
- `area_manager`: author + revision cancel/obsolete/owner manage + approval act/oversee/reassign/cancel + distribution manage/oversee in Area.
- `tenant_owner`: all 29 through normal Authorizer; still obeys relationships, SoD, lifecycle, fresh-auth and tenant-operability constraints.

Approval SoD V1:

1. actor cannot accept if Revision creator or Submission submitter;
2. same user cannot accept two distinct Steps of one ApprovalInstance;
3. reassignment target must remain active/qualified and satisfy SoD.

Last active tenant owner cannot be revoked/deactivated. Responsible owner of review-governed Documents must be reassigned before deactivation.

**R9.5 delta rule:** new Evidence/Dossier operations receive Permissions only after their lifecycles close; operations first, authorization delta second.

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

REV is a business change cycle; autosaves/checkpoints are technical working history and never consume REV numbers. REV numbers never reuse after cancellation. `REV002+` requires reason-for-change before first submit.

`RevisionSubmission` is immutable exact attempt identity and exists even under NoHumanApproval. Approval/Rendition/Release bind exact Submission/digest.

Document code is tenant-wide unique/immutable. Document type/Area/code are immutable V1. Numbering language is deliberately small:

```text
literals + {TYPE} + {AREA} + {SEQ}
sequence_scope: TYPE | TYPE_AREA
sequence_width: minimum zero padding
```

No year/month/custom metadata/formulas/scripts/resets. Normal Create has no manual-code override; legacy preservation belongs to explicit import/migration.

Template is a role of an ordinary governed Document; no parallel TemplateVersion lifecycle. `TemplateUse` is M:N and exact source effective REV is pinned at derived-document creation.

TemplateSpec owns structured authoring/fill contract only. It is optional structured-authoring state for applicable content, **not** the universal definition of Revision content.

---

# 4. LOCKED — Periodic Review / Rendition / Release (R6 refined by R9.5-1)

Periodic Review belongs Controlled Information:

```text
Disabled
or Every(n months)
```

Cadence starts from actual Effectivity; overdue does not invalidate content. Immutable `PeriodicReviewRecord` binds exact effective REV with `confirmed_current | change_required`. Completion requires responsible-owner relation + permission + exact REV still current.

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

The prior universal mandatory OFFICIAL_PDF rule is retired. Locked V1 replacement:

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

Audit Trail itself is not deleted. Platform KEK wraps per-tenant DEK; no per-document key hierarchy/rotation/HSM product V1. R9.5 Retention must refine legal-retention/hold interaction before erasure mechanics freeze.

---

# 7. LOCKED — R9.5 Whole-Product North Star

R9.5 exists because storage, editor, non-DOCX content and enterprise-context use cases materially affect architecture. The previous R10-A topology is **not approved** and remains paused.

> **MetalDocs is the system of record for identity, governance, revision, evidence and documentary context. Physical storage, authoring/editor technology, viewers and upstream ERP/PLM/repositories are replaceable providers/connectors around that kernel.**

Do not replace MetalDocs with M-Files/Nuxeo/Alfresco and do not build a generic ECM/M-Files clone.

A governed Document is format-agnostic. `Dossier` is a deliberately small documentary context for Venda/Produto/Projeto/Equipamento/Case etc.; it may relate Documents/Evidence and external references but never becomes ERP/PLM. BOM/where-used/EBOM/MBOM/CAD dependency/ECR/ECO are explicit PLM integration triggers.

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

Tenant-scoped `EvidenceType` V1 has stable code/name/description/status, allowed formats and small naming policy. It is not a custom schema/workflow engine.

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
  ↓ capture
CAPTURED
  ↓ invalid-capture administration
VOIDED
```

DRAFT content may be replaced. CAPTURED content/metadata is immutable. Wrong capture is VOIDED with reason and replaced by new Evidence; external-world cancellation is a separate fact. Evidence does not use REV/Approval/Release by default; information needing revision governance should be Document.

Exactly one primary Artifact per DocumentRevision and per Evidence V1. True indivisible multi-file packages are future `ArtifactPackage`/PLM triggers.

MetalDocs owns a closed `ContentFormat` catalog. DocumentType/EvidenceType select allowed formats. `RevisionContent` is format-independent:

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

Normal SharePoint/OneDrive/etc. are External Repository Connectors, not ManagedArtifactStore providers. Governed primary content V1 requires an exact MetalDocs-managed copy. Connector directions begin with:

```text
IMPORT_COPY
PUBLISH_COPY
```

External edits never mutate existing EFFECTIVE REV/CAPTURED Evidence. Future adoption imports exact new bytes and creates new governed state/new DRAFT REV where applicable.

SharePoint Embedded is reserved as a future Microsoft-enterprise content-backend/coauthoring/Purview profile, not forced into S3 semantics.

Valid restore = Artifact DB fact + exact bytes + matching SHA-256. Tenant erasure/GC use MetalDocs inventory as semantic authority and tenant-namespaced storage as safety boundary.

---

# 10. LOCKED — R9.5-3 Authoring / EigenPal

## AU-01 — Persisted working truth

Editor/browser state is never business authority. While a Revision is DRAFT, the latest **persisted WorkingContent** is the recoverable server truth.

DRAFT carries a monotonic technical `working_version`. It is not a REV or user-visible business version.

`WorkingSnapshot` is an immutable technical snapshot of logical RevisionContent at a working_version. Autosaves/checkpoints do not create REVxxx. Old technical snapshots may be garbage-collected according to recovery policy unless still current or needed by later immutable evidence.

Any governed DRAFT change — primary bytes, structured values or governed Revision metadata — participates in the same working concurrency version. Editorial discussion that is not part of submitted content does not advance content version merely because comment state changed.

## AU-02 — OCC; no last-write-wins

Every server save/replacement requires `expected_working_version`. Mismatch fails closed. No stale callback/tab/external upload silently overwrites newer persisted work.

V1 uses one active in-app writer per DRAFT Revision **plus OCC as the correctness backstop**. Other authorized users may inspect latest persisted working state.

`EditorSession` is a narrow authoring lease/presence record with heartbeat/staleness, not a generic distributed-lock engine.

## AU-03 — External editing

Download/edit/upload does not hold a long checkout lock. Download/replacement is based on a known working_version; stale-base replacement fails with conflict. No automatic binary DOCX merge V1.

In-app editing and external-file replacement modify the same DRAFT business REV. Authoring method/provider is not persisted business identity.

External replacement must satisfy DocumentType content policy and may not silently destroy TemplateSpec/structured-authoring parity.

## AU-04 — Submission boundary

Submit requires/follows a final successful flush, validates the expected working_version and freezes the exact persisted logical state into immutable `RevisionSubmission`.

After REV becomes SUBMITTED, any later autosave/stale editor callback is rejected. Submission candidate bytes never mutate.

Return-for-changes closes the old Approval attempt and reopens the same REV DRAFT for new WorkingContent; the prior Submission remains immutable forever.

## AU-05 — Approval is read-only content review

Approval UI renders the exact Submission source/preview read-only. Current runtime behavior that maps review to vendor `suggesting` mode with autosave has **no target entitlement**.

Approval rationale/return reason belongs to Approval evidence, not editor comments or source bytes.

Future inline approver suggestions, if proven, must be separate `ReviewSuggestion`/annotation state against exact Submission and must never mutate it.

## AU-06 — Editorial comments / tracked changes

`EditorialComment` is MetalDocs collaboration state attached to DRAFT Revision/working context, not vendor DOCX-comment authority. This keeps `document.comment` product-owned and format-independent.

V1 requires editorial comments resolved before submission so the submitted candidate is clean and unambiguous.

Tracked changes are optional authoring-provider markup, not a MetalDocs lifecycle primitive. If enabled, unresolved tracked changes block submission V1. They are not mandatory V1 requirements; vendor licensing/capability remains explicit build-vs-buy choice.

## AU-07 — Recovery

Browser/local recovery may supplement server snapshots but never silently overwrite a newer server working_version. If local recovery base is stale, surface conflict/export/reconciliation rather than applying automatically.

## AU-08 — Real-time collaboration deferred

Realtime coauthoring/Yjs is deferred V1. The seam remains deliberate: future collaboration state is authoring infrastructure and periodically/at-submit exports exact WorkingSnapshots; it never becomes REV/Submission identity.

Future collaborative submit must quiesce/freeze provider state, export exact source bytes, persist final WorkingSnapshot and only then create immutable Submission.

## AU-09 — EigenPal/provider boundary

Preserve one MetalDocs anti-corruption/provider adapter around EigenPal. No direct vendor API imports across product surfaces.

Pin exact editor dependency/version and require a MetalDocs fidelity/conformance corpus before upgrades. Upstream guarantees are supporting evidence, not MetalDocs proof.

Future AuthoringProviders such as Office/ONLYOFFICE may add/replace capabilities without changing Document/REV/Submission semantics.

V1 authoring north star:

```text
REV004 DRAFT
    ↓
WorkingContent (working_version N)
    ↓
EigenPal OR external edit
    ↓
Save/Replace(expected=N)
    ↓
validated immutable Artifact + governed state
    ↓
WorkingSnapshot N+1
    ↓
final flush/OCC
    ↓
RevisionSubmission #k
    ↓
IMMUTABLE
```

---

# 11. Build-vs-buy rulings

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

# 12. Explicit target deletions / non-goals

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
- vendor DOCX comments as sole product collaboration authority;
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

# 13. Remaining whole-product completion before technical architecture

- [x] R9.5-1 Content Model
- [x] R9.5-2 Storage / Repository Strategy
- [x] R9.5-3 Authoring / EigenPal
- [ ] **R9.5-4 Dossier / Context — NEXT**
- [ ] R9.5-5 Retention / Records / Legal Hold
- [ ] R9.5-6 Import / Migration / Export
- [ ] R9.5-7 Attestation + Content Security
- [ ] R9.5-8 Whole-product adversarial freeze

## R9.5-4 must close

1. exact semantic role of Dossier versus Document/Evidence and versus ERP/PLM entity;
2. DossierType configurability without becoming custom-object platform;
3. local-created versus externally-originated Dossier identity;
4. stable display/business key and canonical naming implications;
5. external references and source-system identity mapping;
6. Document/Evidence membership/relationship cardinality;
7. whether one Document/Evidence may relate to multiple Dossiers;
8. Dossier-to-Dossier relations and whether V1 needs any;
9. lifecycle/close/archive semantics and whether external source status is mirrored or merely projected;
10. what metadata belongs on Dossier versus remains external projection;
11. allowed/recommended EvidenceTypes/DocumentTypes by DossierType;
12. search/navigation/activity timeline semantics;
13. ERP/PLM synchronization boundary and conflict/source-of-truth rules;
14. precise triggers where requirements stop being Dossier and become ERP/CRM/PLM domain.

Then R9.5-5 Retention / Records / Legal Hold.

---

# 14. Technical-design queue after R9.5

R10: bounded contexts/dependency DAG/aggregate ownership/data model/table ownership/constraints/transactions/events/jobs/storage implementation/current-module & current-table migration map.

R11: APIs/OpenAPI/DTOs/problem semantics/frontend IA and complete journeys.

R12: proof/Golden Matrix/threat/invariant tests + final durable ADR/spec/wiki promotion + adversarial review.

R13: implementation specification + sequenced implementation plan, then code.

Until all remaining gates close: **NO PRODUCT IMPLEMENTATION.**

---

# 15. Exact next step

Continue **R9.5-4 — Dossier / Context**.

Preserve:

```text
Dossier = documentary context, not source ERP/PLM object
Document/Evidence keep independent identities/lifecycles
Artifact stays technical only
external systems may supply references/projections but cannot rewrite MetalDocs governed truth
no generic custom-object platform V1
```

Use Venda, Produto, Projeto and Equipamento as adversarial examples and keep the minimum common model that supports all four without reproducing ERP/PLM.