# MetalDocs Cohesive Platform Redesign — Active Decision Ledger

> **Status:** ACTIVE WIP — operator-approved decisions below are binding; unresolved items are explicit.
> **Date:** 2026-08-16
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
3. Approval/Rendition/Release bind the same exact immutable RevisionSubmission/digest;
4. no generic BPM/ReBAC/low-code/object-platform engine without proven need;
5. no confirmed binary content without semantic ownership by a governed business object;
6. editor/browser state is never the authoritative persisted DRAFT truth;
7. Dossier is documentary context, never a hidden ERP/PLM/custom-object platform;
8. retention expiry never means automatic deletion;
9. imported historical truth is never rewritten as if it were native MetalDocs history;
10. launch scope favors simple invariant-preserving controls over speculative compliance/security platforms.

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

Groups are flat V1. Area is the organizational truth reused by Document ownership, scoped authorization and Approval actor resolution.

Five tenant roles only:

```text
tenant_owner
area_manager
author
approver
viewer
```

One additive/default-deny grant shape:

```text
RoleAssignment
  subject: User | Group
  role
  scope: TenantScope | AreaScope
  grant/revocation evidence
```

No tenant-owner bypass, generic ACL/ReBAC graph, nested groups, deny engine or magic scope sentinel.

Authorization equation:

```text
Permission
+ required case/resource relationship
+ Domain Governance constraints
= ALLOW
```

RLS = tenant-isolation defense-in-depth only. DB constraints = structural invariants. Current `system_admin` short-circuit/asserted-capability-GUC model is not target architecture.

R9's 29-permission catalog remains authority for the R3–R9 operation set:

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

Role bundles remain viewer / author / approver / area_manager / tenant_owner. Strict Approval SoD remains: creator/submitter cannot accept own Submission; same user cannot accept two Steps of one ApprovalInstance; reassignment must remain qualified and SoD-valid.

**R9.5 delta rule:** Evidence, Dossier, Retention, migration/export and any remaining operations receive one bounded authorization delta only after R9.5 operations close; operations first, permissions second.

---

# 2. LOCKED — Approval V1

Approval is a specialized sequential human workflow, not BPM:

```text
ApprovalPolicy(version)
  ordered Steps

Step:
  purpose: review | approval
  actor_rule: NamedUser | Group | RoleInArea
  completion: ANY | ALL
  requires_reauthentication
  due_in_days?
```

Human outcomes: `accept | return_for_changes`. Separate operations: `withdraw | cancel | reassign`. No normal terminal reject V1.

ApprovalInstance binds exactly one immutable `RevisionSubmission`. Return/withdraw terminates the attempt and returns the same REV to DRAFT; resubmission creates a new Submission and, when required, new ApprovalInstance. Submitted candidate bytes never reopen for mutation.

---

# 3. LOCKED — Controlled Information foundations (R3–R5)

`DocumentType` is tenant-scoped with immutable code, display fields, optional classification-only category and ACTIVE/INACTIVE lifecycle. GovernanceClass is deleted.

Approval config is explicit: `NoHumanApproval | UsePolicy(ApprovalPolicyID)`.

`Document` = stable governed identity. At most one EFFECTIVE + one open Revision V1. Official labels are `REV001`, `REV002`, ...

Revision states:

```text
DRAFT
SUBMITTED
EFFECTIVE
SUPERSEDED
OBSOLETE
CANCELLED
```

REV is a business change cycle; autosaves/checkpoints are technical history and never consume REV numbers. REV numbers never reuse. `REV002+` requires reason-for-change before first submit.

`RevisionSubmission` is immutable exact attempt identity, including under NoHumanApproval. Approval/Rendition/Release bind exact Submission/digest.

Document code/type/Area are stable V1. DocumentType numbering remains literals + `{TYPE}/{AREA}/{SEQ}`, TYPE or TYPE_AREA scope, minimum padding. No generic formula/reset/custom-metadata numbering.

Template is a role of an ordinary governed Document; no parallel TemplateVersion lifecycle. `TemplateUse` is M:N and exact source effective REV is pinned at derived-document creation.

TemplateSpec is optional structured-authoring state for applicable content, not universal Revision content.

---

# 4. LOCKED — Periodic Review / Rendition / Release

Periodic Review belongs Controlled Information: `Disabled | Every(n months)`. Due/overdue does not invalidate EFFECTIVE content. Immutable `PeriodicReviewRecord` binds exact current REV and outcome.

Rendition = immutable derived representation of exact Submission with output hash + generator/build provenance. Approval approves Submission, never renderer output bytes.

Release is automatic/system-owned; no publish button. Optional `ReleasePlan.not_before`; actual `effective_at = released_at`. Winning release atomically makes candidate EFFECTIVE, prior REV SUPERSEDED, swaps pointers, writes ReleaseRecord and durable events.

Universal mandatory PDF is retired. Locked V1 policy:

```text
OfficialRepresentationPolicy = SourceOnly | RequireRendition(ContentFormat)
```

At most one required derived rendition V1. Exact primary source Artifact is always frozen by Submission.

---

# 5. LOCKED — Distribution / Values / Audit / Notifications / Search

Distribution = controlled obligation/acknowledgement, not AuthZ/LMS. Release snapshots concrete users; later Group membership never rewrites historical denominator. Explicit immutable AcknowledgementRecord completes obligation; notification read/view/download never does.

System Value Catalog remains small/product-owned. Tenant Dictionary values snapshot when a new REV is created; same-REV return/resubmit does not silently re-resolve.

Domain evidence records remain authorities. AuditEvent is transversal timeline only. Critical governed mutation cannot report success without durable audit intent/event in the same commit boundary. Audit remains append-only, tamper-evident and exportable.

Notifications = delivery projection only. Search = rebuildable/eventually-consistent discovery projection and never grants canonical access.

---

# 6. LOCKED — Tenant lifecycle / Platform Security (R8 refined by Retention)

PlatformOperator/SystemPrincipal are outside tenant RBAC and gain no implicit tenant-content access.

Tenant lifecycle remains exactly:

```text
ACTIVE
SUSPENDED
ERASED
```

Deletion request is separate. Terminal erasure is retention-aware:

```text
request reaches execute_after
→ evaluate RetentionBindings / active LegalHolds
→ blockers: request remains pending/blocked; Tenant != ERASED
→ no blockers: suspend/revoke sessions
→ erase eligible substantive rows/blobs
→ destroy no-longer-needed Tenant DEK
→ preserve allowed non-PII audit/platform skeleton
→ Tenant ERASED
→ TenantErasureRecord
```

No new tenant state for retention. A DEK needed to preserve legally retained intelligible content is not destroyed while that obligation remains. Backup/restore must reapply erasure tombstones and reconcile retention/hold facts before cleanup/service resumes.

---

# 7. LOCKED — R9.5 Whole-Product North Star

> **MetalDocs is the system of record for identity, governance, revision, evidence and documentary context. Physical storage, authoring/editor technology, viewers and upstream ERP/PLM/repositories are replaceable providers/connectors around that kernel.**

Do not replace MetalDocs with M-Files/Nuxeo/Alfresco and do not build a generic ECM clone.

A governed Document is format-agnostic. Dossier is deliberately small documentary context, not ERP/PLM. EigenPal is a DOCX authoring provider, never Document identity.

Storage classes:
1. Managed Artifact Store — exact MetalDocs-owned bytes.
2. External Repository Connector — explicit import/publish copies.
3. Future enterprise content profile — e.g. SharePoint Embedded/M365, designed explicitly rather than forced into S3 semantics.

Provider WORM/Purview may enforce retention physically; MetalDocs owns business retention/hold semantics if it owns the record.

Freeze rule: future ideas reopen the kernel only when they create a material identity/historical-truth/invariant counterexample. Otherwise use an existing provider/connector/context seam.

---

# 8. LOCKED — R9.5-1 Content Model

`Artifact` = immutable technical identity of exact bytes with canonical SHA-256, size, ContentFormat/media type and technical provenance. It is not user-facing business identity. Provider location/version/URL never enters Artifact business identity or Submission digest.

No confirmed orphan Artifact library. Temporary staging may precede classification, but confirmed Artifact must belong to DocumentRevision or Evidence.

Tenant-scoped `EvidenceType` defines stable code/name/status, allowed formats and a small canonical naming policy. User filename is provenance only. Evidence naming tokens V1: `{TYPE}`, `{DOSSIER}`, `{REF}`, `{SEQ}`.

Evidence lifecycle:

```text
DRAFT → CAPTURED → VOIDED // VOIDED only for invalid MetalDocs capture
```

CAPTURED content/metadata is immutable. External-world cancellation is a separate fact. Evidence does not use REV/Approval/Release by default.

Exactly one primary Artifact per DocumentRevision and Evidence V1. True indivisible multi-file packages are future ArtifactPackage/PLM triggers.

MetalDocs owns closed `ContentFormat` catalog. Format-independent RevisionContent:

```text
primary_artifact + governed_metadata + optional structured_authoring
```

Submission digest binds exact Artifact hash + governed state + decision-relevant structured/template provenance, never storage location.

---

# 9. LOCKED — R9.5-2 Storage / Repository Strategy

Canonical content hash = SHA-256 of exact primary bytes. One active Managed Artifact Store per deployment V1. First-class adapters: Local(dev/test), MinIO, AWS S3. Other S3-compatible products require conformance validation.

Provider migration = copy exact bytes + verify canonical hash + cutover; no new Artifact/REV/Submission and no permanent dual-write V1.

Managed keys are opaque, immutable and tenant-namespaced. Business filename never determines path. Artifact ID != content hash; no content-addressed/cross-tenant dedup V1.

Temporary/direct-presigned upload is allowed. Provider success does not confirm Artifact before basic integrity/content/semantic validation. Existing Artifact keys are never overwritten.

Object-store versioning and Object Lock/WORM are defense/enforcement only, never REV/retention authority. Production baseline = encrypted transport + provider encryption at rest. Tenant DEK does not encrypt every Artifact V1.

Normal SharePoint/OneDrive/etc. are External Repository Connectors, not ManagedArtifactStore providers. Governed primary content V1 requires exact MetalDocs-managed copy. Connector directions begin `IMPORT_COPY` / `PUBLISH_COPY`. External edits never mutate existing MetalDocs history.

SharePoint Embedded remains a future Microsoft-enterprise content profile. Valid restore = Artifact DB fact + exact bytes + matching SHA-256.

---

# 10. LOCKED — R9.5-3 Authoring / EigenPal

Latest persisted `WorkingContent` is recoverable DRAFT truth; browser/editor is never authority. DRAFT uses monotonic technical `working_version`; immutable WorkingSnapshots are technical checkpoints, never REVxxx.

All governed DRAFT changes share one OCC version. Save/replacement requires `expected_working_version`; no last-write-wins. V1 uses one active in-app writer + OCC. EditorSession is a narrow heartbeat/staleness authoring lease.

External download/edit/upload holds no long checkout and fails on stale base; no automatic binary DOCX merge V1. In-app/external editing modify same DRAFT REV; provider is not business identity.

Submit requires/follows final successful flush, validates OCC and freezes exact persisted logical state into immutable RevisionSubmission. SUBMITTED rejects stale later autosaves/replacements.

Approval UI is read-only over exact Submission. Approval rationale is Approval evidence, not editor state. `EditorialComment` is product-owned DRAFT collaboration state; unresolved comments and, if enabled, tracked changes block submission V1.

Realtime Yjs/coauthoring deferred V1 but seam preserved. Preserve one EigenPal anti-corruption/provider adapter, exact dependency pin and MetalDocs fidelity corpus. Future Office/ONLYOFFICE providers cannot change core semantics.

---

# 11. LOCKED — R9.5-4 Dossier / Context

`Dossier` = stable documentary context for an identifiable business subject; not physical folder and not ERP/PLM entity.

Tenant-scoped `DossierType` stays small: code/name/description/status + eligible DocumentTypes/EvidenceTypes. No custom fields/forms/workflow/ACL engine and no required-evidence completeness engine V1.

Dossier has stable key unique within tenant+type; title may change. `{DOSSIER}` resolves stable key. No generic Dossier numbering V1.

Creation provenance is separate from zero..N ExternalReferences. ExternalReference uses connection + entity kind + external ID; same external identity cannot map to two Dossiers. No heuristic auto-merge.

External master fields/status remain projections, not canonical Dossier state. Source disappearance never deletes history.

Dossier↔Document is M:N over stable Document identity and never copies content or changes Document lifecycle/Area/AuthZ. Every CAPTURED Evidence has exactly one immutable primary Dossier; DRAFT may correct it; secondary Dossier links are allowed without duplication.

Dossier scope = one `TenantScope | AreaScope`; Evidence reuses primary Dossier scope. Dossier type/key/scope stable V1.

Lifecycle = `ACTIVE ↔ ARCHIVED`; archive is reversible navigation state, not external business status, and never deletes related content.

No Dossier-to-Dossier graph/hierarchy V1. Search/timeline are projections. ERP/CRM, PLM, PM and EAM/CMMS boundaries remain explicit.

---

# 12. LOCKED — R9.5-5 Retention / Records / Legal Hold

No generic `Record` entity/declaration button. Existing governed objects become retention subjects automatically.

- CAPTURED Evidence creates RetentionBinding.
- DocumentRevision creates RetentionBinding at first RevisionSubmission.
- Draft never submitted, staging and recovery WorkingSnapshots remain operational/GC data, not records retention.

DocumentRevision retention unit includes its governed immutable history: Submissions, Approval evidence, Renditions, Release/Review evidence and referenced Artifacts.

DocumentType/EvidenceType choose explicit:

```text
NoMinimum
KeepFor(value, DAYS|MONTHS|YEARS)
Indefinite
```

No hardcoded legal periods and no NULL-as-policy.

Document retention clock does not run while REV is EFFECTIVE. Anchor = superseded_at / obsoleted_at / cancelled_at for a submitted-but-never-released REV. EvidenceType chooses `CAPTURED_AT | OCCURRED_AT`; Dossier archive never starts retention.

Policy is snapshotted in RetentionBinding; later type changes do not silently recalculate existing records. `RetentionExtension` may only lengthen retention V1. Generic retroactive shortening is excluded.

Expiry only makes a subject eligible for disposition. No automatic delete. Current EFFECTIVE REV is never disposition-eligible. Physical disposition requires explicit authorized review, no active hold, and verified removal before immutable DispositionRecord says disposal completed.

LegalHold is independent of retention. V1 hold scopes: Evidence, stable Document, Dossier. Document/Dossier holds materialize concrete held subjects; unlink/lifecycle changes cannot make previously held content escape preservation. Holds block disposal, not normal business lifecycle.

Hold V1 covers confirmed governed record content, not transient autosaves/staging/full eDiscovery ESI. Artifact has no independent retention policy; preservation derives from subjects referencing it.

Provider WORM/Object Lock/Purview is enforcement only. DossierType has no retention policy. Audit Trail remains a separate retention regime. Tenant terminal erasure is blocked until relevant RetentionBindings/Holds permit destruction. No post-termination Retention Vault without real requirement.

---

# 13. LOCKED — R9.5-6 Import / Migration / Export

Ordinary `IMPORT_COPY` follows normal lifecycle. **Historical Migration** is a privileged path for pre-MetalDocs history and never fabricates native facts.

Every migrated object carries explicit source provenance. Historical lifecycle/approval/effectivity facts are imported governance evidence, not synthetic ApprovalDecision/ApprovalInstance/ReleaseRecord/internal-user actions.

Imported EFFECTIVE/SUPERSEDED/OBSOLETE states require imported proof. If source effectivity date is unknown, preserve unknown and record explicit `adopted_as_current_at`; never invent a date.

Reliable numeric legacy ordinals may map directly (`7 → REV007`); current-state-only import may begin at REV007 with gaps. Arbitrary labels map deterministically to REVxxx while preserving source label. Next native REV is always above highest imported ordinal.

Privileged migration may preserve safe unique legacy Document codes. Otherwise mapping is explicit and source code remains provenance.

Migration modes: `CURRENT_STATE | FULL_HISTORY`. A target DocumentRevision requires exact primary Artifact bytes. Missing historical content may remain imported-history evidence but does not create a fake Revision. No silent content-format conversion.

Historical actors stay source snapshots/references; optional user correlation is provenance only. Migration writes are attributed to Migration/System principal.

Dossier migration uses stable key + ExternalReference uniqueness. Historical Evidence separates occurred/captured facts from migration time. Imported retention uses trustworthy historical anchor when known; unknown anchor never silently becomes deletion-eligible. Migration never automatically disposes old content or replays old notifications/jobs/distribution side effects.

Historical Migration uses first-class batch/plan semantics with true dry-run, deterministic per-item outcomes and reconciliation report. Same source identity + same content → REUSE; conflicting content/state → fail closed. Atomicity is per semantic import unit; partial batch success is allowed/reconciled. No magical whole-migration rollback promise.

Export contracts remain distinct:

```text
Backup
Tenant Portability Export
Governed Subject Export
External Repository PUBLISH_COPY
```

`tenant.export` produces provider-independent tenant business/governance state + exact Artifacts, not deployment internals. It includes current DRAFT state needed to continue business but excludes staging, GC-able old WorkingSnapshots, projections, jobs/outbox and secrets.

Governed Subject Export may package Document/Evidence/Dossier independently. Every portability/governed package has a versioned provider-independent manifest with objects, relationships, provenance, canonical filenames, ContentFormats/sizes and SHA-256 values. Manifest never depends on provider keys/URLs/version IDs.

Generated export bundle is temporary delivery output, not automatically Evidence. Retention/Hold does not forbid authorized export and export does not release hold/change retention/count as acknowledgement. Export completeness must be explicit and authorization-safe.

Compatible MetalDocs→MetalDocs portability may preserve native authorities when its source/package trust requirements are satisfied by the eventual implementation contract. V1 does **not** require cryptographically signed portability packages.

---

# 14. LOCKED — R9.5-7 Launch Attestation + Basic Content Safety (YAGNI)

The earlier broad proposal for antivirus/quarantine/PKI/signed packages/sandbox infrastructure was explicitly rejected as over-engineered for pre-launch V1. The approved launch scope is deliberately small.

## Attestation

1. `ApprovalDecision` always binds the exact `RevisionSubmission` and its digest. Changing the governed bytes/state creates different content and existing approval does not apply.
2. Preserve actor, Step, ApprovalPolicy version, decision, trusted server timestamp and required AuthN assurance/fresh-auth evidence.
3. `return_for_changes` requires a reason V1.
4. MetalDocs V1 claims **authenticated application approval**, not ICP-Brasil/qualified digital signature or other legal-signature level it does not actually implement.
5. Approval/effectivity may be manifested in a human-readable PDF/certificate Rendition, but source bytes approved by humans are never modified merely to stamp approval.
6. NoHumanApproval never fabricates a human/System approver.

## Basic content safety

1. `DocumentType`/`EvidenceType` accept only explicitly supported `ContentFormat`s.
2. Upload/import has basic size limits and coherent format validation; client filename is never authoritative and canonical naming remains MetalDocs-owned.
3. MetalDocs does not intentionally execute user-uploaded content. Formats not supported for safe in-app preview are download-only or use an existing controlled viewer/rendition path.
4. Macro-enabled Office formats (`DOCM/XLSM/PPTM`) are outside the normal V1 support set unless explicitly reconsidered later.
5. Rendering remains a supporting service receiving content and returning a derived result; it receives no business authority. Advanced custom renderer-sandbox infrastructure is not a V1 product requirement.

## Explicitly deferred — triggers, not hidden launch TODOs

Do **not** build for V1 without a concrete customer/security/regulatory trigger:

```text
malware scanning / ClamAV
quarantine lifecycle
periodic malware rescan / ArtifactSecurityAssessment
CDR / advanced active-content inspection
ICP-Brasil / PKI / DocuSign / Adobe Sign
RFC3161/TSA / HSM
cryptographically signed export packages / signing-key lifecycle
custom export encryption format
macro-enabled Office support
full custom renderer sandbox/egress platform
eDiscovery / ESI preservation
```

The staging→validation→confirmation seam already leaves a natural place to add malware inspection later without redesigning Document/Evidence/Artifact/Submission.

---

# 15. Build-vs-buy / launch non-goals

Current rulings remain: no external ECM kernel, no JCR kernel, no generic BPM/ReBAC/low-code engine, no mandatory realtime collaboration, no generic PLM/ERP/PM/CMMS features, no generic multi-cloud/BYOS/dedup/silent sync, no provider identity in domain, no universal PDF rule.

Additional launch non-goals include fake native migration history, generic migration transformation engine, retroactive side-effect replay, raw backup as portability contract, export of deployment secrets, antivirus platform, PKI/signature infrastructure, signed-package infrastructure and advanced content-security platform before a real requirement exists.

---

# 16. Remaining whole-product completion before technical architecture

- [x] R9.5-1 Content Model
- [x] R9.5-2 Storage / Repository Strategy
- [x] R9.5-3 Authoring / EigenPal
- [x] R9.5-4 Dossier / Context
- [x] R9.5-5 Retention / Records / Legal Hold
- [x] R9.5-6 Import / Migration / Export
- [x] R9.5-7 Launch Attestation + Basic Content Safety
- [ ] **R9.5-8 Whole-product adversarial freeze — NEXT**

R9.5-8 must not introduce speculative features. Its job is to:

1. adversarially test R3–R9.5 against representative end-to-end scenarios and failure cases;
2. identify genuine contradictions/gaps only;
3. run a final YAGNI/deletion pass and distinguish launch requirements from future seams;
4. close the bounded authorization delta for new Evidence/Dossier/Retention/Import/Export operations;
5. produce the final whole-product capability/domain freeze;
6. determine whether any material issue still blocks descent into R10 technical architecture.

Representative adversarial cases must include at least:

```text
DOCX in-app controlled procedure
external XLSX controlled document
native PDF / SVG / CAD-style source
Evidence upload inside Sale Dossier
Product Dossier with mechanical/electrical/manual documents
MinIO→S3 relocation
SharePoint import/publish copy
return-for-changes + resubmit
stale external edit / stale autosave
historical migration with incomplete history
retention expiry + LegalHold
tenant deletion blocked by retained content
external-system disappearance
cross-scope authorization / SoD
provider/render/job failure without business-truth corruption
```

---

# 17. Implementation gate

**NO PRODUCT IMPLEMENTATION.** R10 bounded contexts/filesystem/data model remains blocked until R9.5-8 is explicitly approved.