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
7. Dossier is documentary context, never a hidden ERP/PLM/custom-object platform;
8. retention expiry never means automatic deletion; disposition requires eligibility + no active hold + explicit authorization + verified physical deletion.

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

**R9.5 delta rule:** Evidence/Dossier/Retention operations receive their bounded authorization delta only after all R9.5 operations close; operations first, permissions second.

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

`Document` is stable governed identity. At most one EFFECTIVE + one open Revision V1. Official labels are `REV001`, `REV002`, ...

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

# 6. LOCKED — Tenant lifecycle / Platform Security (R8 refined by R9.5-5)

`PlatformOperator` / `SystemPrincipal` are outside tenant RBAC and gain no implicit tenant-content access.

Tenant lifecycle remains exactly:

```text
ACTIVE
SUSPENDED
ERASED
```

Deletion request is a separate grace/cancel process. Onboarding creates Tenant + initial User + `tenant_owner @ Tenant` + single-use activation credential; platform operator never chooses tenant-owner password.

Suspension revokes sessions, blocks login/business mutations and is reversible. Tenant export/deletion request require fresh-auth.

Terminal erasure is now explicitly retention-aware:

```text
deletion request reaches execute_after
→ evaluate RetentionBindings / active LegalHolds
→ if blockers exist: request remains pending/blocked; Tenant is NOT ERASED
→ if no blockers: suspend/revoke sessions
→ erase eligible live tenant rows/blobs
→ destroy Tenant DEK that is no longer needed
→ preserve allowed non-PII audit/platform skeleton
→ Tenant ERASED
→ TenantErasureRecord
```

No new tenant state is introduced for retention. Suspension remains independent. Audit Trail itself is not deleted. Platform KEK wraps per-tenant DEK; no per-document key hierarchy/rotation/HSM product V1.

A DEK required to keep legally retained content intelligible is never destroyed while that obligation remains. Backup/restore must reapply erasure tombstones and restore/reconcile retention/hold facts before cleanup/service resumes.

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

Object-store versioning is defense-in-depth, never `REVxxx`. Object Lock/WORM/legal-hold support is optional physical enforcement consumed by MetalDocs Retention governance, never business authority.

Production baseline = encrypted transport + provider encryption at rest. Do not encrypt every Artifact with Tenant DEK V1; application Artifact crypto-shred is introduced only if Retention/Erasure later proves a concrete need.

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

`Dossier` is a stable documentary context for an identifiable business subject, not the ERP/PLM object itself and not a physical folder.

Tenant-scoped `DossierType` V1 is deliberately small:

```text
code
name
description?
status: ACTIVE | INACTIVE
eligible DocumentTypes
eligible EvidenceTypes
```

No custom fields, forms, formulas, workflow, custom ACL/lifecycle or generic relation schema. Eligibility is real validation; no required-evidence/completeness checklist V1.

Dossier has stable human/business `key`, unique within `(tenant, DossierType)`. Title may change; key does not. `{DOSSIER}` resolves the stable key. No generic Dossier numbering engine V1; creator/integration supplies key.

Creation provenance is separate from zero..N ExternalReferences. `ExternalReference = connection + entity_kind + external_id + optional display reference`; same external identity cannot silently point to two Dossiers. No heuristic auto-merge; ambiguity fails closed.

External master/status fields remain source-system projections, not canonical Dossier state. Source disappearance never deletes documentary history.

Dossier↔Document is M:N over stable Document identity. The relation never copies content, changes Document type/Area/lifecycle/AuthZ or grants access. Exact-REV usage evidence is a distinct future/explicit concept.

Every CAPTURED Evidence has exactly one immutable `primary_dossier`; DRAFT may correct it. Evidence may relate secondarily to other Dossiers without duplication. Primary Dossier supplies Evidence naming/context/scope.

Dossier uses exactly one `TenantScope | AreaScope`; Evidence reuses primary Dossier scope V1. No multi-area ACL. Dossier type/key/scope are stable V1.

Lifecycle is only:

```text
ACTIVE ↔ ARCHIVED
```

ARCHIVED is reversible MetalDocs navigation state, not external business status. Archiving never deletes/obsoletes related content or mutates external systems.

Document/Evidence links preserve link/unlink history. No Dossier-to-Dossier hierarchy/graph V1. Search/timeline are projections.

Boundaries are explicit: transaction calculations/workflows stay ERP/CRM; BOM/part/CAD/change management stays PLM; schedules/resources stay PM; work orders/assets operations stay EAM/CMMS. MetalDocs owns documentary context.

---

# 12. LOCKED — R9.5-5 Retention / Records / Legal Hold

## RT-01 — No duplicate Record identity

Do not create a generic `Record` entity/lifecycle. Existing governed subjects become retention subjects automatically:

- CAPTURED Evidence creates a `RetentionBinding`;
- DocumentRevision creates a `RetentionBinding` at its first `RevisionSubmission`;
- a DRAFT Revision never submitted remains recovery/working data rather than records-retention data.

The DocumentRevision retention unit includes the governed Revision history necessary to prove it: immutable Submissions, Approval evidence, relevant Renditions, Release/PeriodicReview evidence and referenced Artifacts. WorkingSnapshots/staging/abandoned drafts remain under recovery/GC policy, not records retention.

## RT-02 — Small explicit retention configuration

DocumentType and EvidenceType each choose an explicit rule:

```text
NoMinimum
KeepFor(value, DAYS | MONTHS | YEARS)
Indefinite
```

No NULL-as-policy, hardcoded statutory periods, generic FilePlan/cutoff expression language, formula/script rules or Dossier-level retention inheritance V1.

MetalDocs supplies the mechanism; tenant compliance/legal/business owners supply the actual periods applicable to their obligations.

## RT-03 — Retention anchors

A DocumentRevision retention clock does **not** run while the Revision is EFFECTIVE. The anchor is determined by the approved lifecycle:

```text
SUPERSEDED → superseded_at
OBSOLETE   → obsoleted_at
CANCELLED after at least one Submission → cancelled_at
```

A cancelled never-submitted DRAFT does not become a records-retention subject.

EvidenceType chooses only:

```text
CAPTURED_AT
OCCURRED_AT
```

as V1 anchors. `OCCURRED_AT` makes `Evidence.occurred_at` mandatory before capture, including historical imports.

Dossier archive/status never starts retention; Dossier is not retention authority.

## RT-04 — Policy snapshot and extensions

Retention policy is snapshotted into `RetentionBinding`; later type-policy changes do not silently recalculate existing records.

An explicit audited `RetentionExtension` may only lengthen an existing retain-until date with reason/authority/evidence. Generic retroactive shortening is not a V1 operation; if ever required, it needs a separately designed high-risk correction flow.

## RT-05 — Expiry means eligibility, not deletion

Retention expiry derives only:

```text
EligibleForDisposition
```

Current EFFECTIVE revisions are never disposition-eligible regardless of date.

V1 has **no automatic deletion cron**. Physical disposition requires explicit authorized review/decision and is complete only after substantive DB payload + managed Artifacts are verifiably removed. Completion creates immutable `DispositionRecord` evidence. If provider deletion is blocked/fails, disposition remains incomplete.

## RT-06 — Legal Hold is independent

`LegalHold` is separate from retention duration and records explicit apply/release facts, actor/time, reason and optional case reference.

V1 legal-hold scopes:

```text
Evidence
stable Document
Dossier
```

Never Artifact directly.

Document and Dossier holds materialize concrete governed subjects in `LegalHoldItems`. New governed subjects entering that scope while the hold remains active are also materialized. Later unlink/lifecycle changes never release an already-held item implicitly.

Disposition rule:

```text
retention requirement ended
AND zero active LegalHolds
AND subject lifecycle permits destruction
AND explicit disposition authorized
= may attempt physical deletion
```

Hold blocks destruction, **not** normal business lifecycle: a Document may supersede, Evidence may be VOIDED for wrong capture, Dossier may archive, while held facts remain preserved.

Legal Hold V1 covers confirmed governed records, not transient autosaves/staging. Full eDiscovery/ESI preservation is a future separate capability.

## RT-07 — Artifact follows its retention subjects

Artifact has no independent business retention policy. Its preservation/deletion follows all retention subjects referencing it. No semantic/cross-tenant dedup makes this tractable.

Provider Object Lock/WORM/Purview is enforcement only. MetalDocs remains semantic authority:

- if MetalDocs requires preservation, absence of provider lock never permits deletion;
- if MetalDocs allows disposition but provider still locks, physical disposition is blocked/incomplete.

Deployment may choose an enforcement posture such as DB-only, WORM-governance or WORM-compliance. Provider-specific mode names do not appear in DocumentType/EvidenceType business configuration.

If physical WORM is mandatory for a deployment, creation of a retained record cannot report final success while required enforcement failed; exact choreography is R10.

## RT-08 — Dossier, Audit and external copies

DossierType has no RetentionPolicy. Dossier may scope a LegalHold, while each related record keeps its own retention schedule.

Audit Trail remains under a separate retention regime defined by R7/R8; Document/Evidence policies do not govern Audit. Retention/hold/disposition operations themselves emit Audit evidence.

`PUBLISH_COPY` to a normal external repository never replaces canonical MetalDocs retention. A future SharePoint Embedded profile may map MetalDocs retention/hold semantics to Purview enforcement while preserving MetalDocs as the semantic authority.

## RT-09 — Tenant erasure interaction

Terminal tenant erasure evaluates all retention/hold blockers before `ERASED`.

While protected subjects remain:

```text
TenantDeletionRequest remains pending/blocked
Tenant != ERASED
required DEK material is not destroyed
```

No `RETENTION_PENDING` tenant state is added. Suspension remains independent.

R10 should minimize the retained surface during long blocked periods, but V1 does not invent a post-termination Retention Vault/custody subsystem without concrete economic/regulatory need.

Backup/restore must restore/reconcile `RetentionBinding`, `LegalHold`, hold items and disposition facts before cleanup; tenant-erasure tombstones retain precedence.

---

# 13. Build-vs-buy rulings

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
| full eDiscovery/records platform | do not build V1 | ESI preservation, legal discovery workflows or transfer/file-plan requirements become explicit |

---

# 14. Explicit target deletions / non-goals

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
- duplicate generic `Record` identity;
- hardcoded statutory retention periods;
- generic FilePlan/cutoff rules engine;
- retention expiry auto-delete;
- Dossier retention inheritance;
- generic query-based legal hold / full eDiscovery V1;
- provider WORM/Purview as business retention authority;
- previous R10-A topology as approved truth.

---

# 15. Remaining whole-product completion before technical architecture

- [x] R9.5-1 Content Model
- [x] R9.5-2 Storage / Repository Strategy
- [x] R9.5-3 Authoring / EigenPal
- [x] R9.5-4 Dossier / Context
- [x] R9.5-5 Retention / Records / Legal Hold
- [ ] **R9.5-6 Import / Migration / Export — NEXT**
- [ ] R9.5-7 Attestation + Content Security
- [ ] R9.5-8 Whole-product adversarial freeze

## R9.5-6 must close

1. how existing external/legacy Documents enter MetalDocs without pretending their history originated here;
2. preservation of legacy codes and revision labels (`REV07`, `7`, arbitrary legacy labels) versus normalization to target `REVxxx`;
3. whether migration can import current state only, full historical revisions, or both;
4. how externally approved/released records are represented without fabricating MetalDocs ApprovalDecision/ReleaseRecord;
5. import provenance, source-system identifiers and exact artifact hashes;
6. Evidence and Dossier import, including historical `occurred_at` and retention anchors;
7. duplicate/conflict detection and idempotent replay of large migrations;
8. migration validation, dry-run, reconciliation and rollback/abort boundaries;
9. generic user-facing import versus privileged migration/import paths;
10. export scope: one Document/Dossier versus tenant-wide export;
11. export package content, manifests, hashes, relationships and provenance so another system can verify what was exported;
12. whether export is backup, interoperability package or evidentiary package — these purposes must not be conflated;
13. external repository import/publish interaction with ordinary migration/export;
14. retention/legal-hold constraints on export and disposition-related exports.

Then R9.5-7 Attestation + Content Security.

---

# 16. Technical-design queue after R9.5

R10: bounded contexts/dependency DAG/aggregate ownership/data model/table ownership/constraints/transactions/events/jobs/storage implementation/current-module & current-table migration map.

R11: APIs/OpenAPI/DTOs/problem semantics/frontend IA and complete journeys.

R12: proof/Golden Matrix/threat/invariant tests + final durable ADR/spec/wiki promotion + adversarial review.

R13: implementation specification + sequenced implementation plan, then code.

Until all remaining gates close: **NO PRODUCT IMPLEMENTATION.**

---

# 17. Exact next step

Continue **R9.5-6 — Import / Migration / Export**.

Preserve:

```text
legacy/external history may be preserved but never impersonated as native MetalDocs history
Artifact hash/provenance remains exact and provider-independent
Document/Evidence/Dossier identities remain semantic
retention anchors must respect actual historical facts, not migration date
imports must be replay-safe/reconcilable
export must be verifiable and purpose-specific
```

Design the smallest import/export model that can onboard real companies and integrate repositories/ERP/PLM without turning migration into a second mutable authority.