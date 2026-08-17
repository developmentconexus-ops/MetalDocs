# MetalDocs Cohesive Platform Redesign — Active Decision Ledger

> **Status:** ACTIVE — R3–R9.5 operator-approved decisions below are binding; R9.5 is **FROZEN**; R10 is **NEXT**.
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** `docs/engineering/standards/root-cause-global-maximum-method.md` — byte-equivalent mirror of DevelopmentConexus Engineering Method v1.0.0
> **Program authority:** `wiki/architecture/cohesive-platform-redesign.md`
> **Implementation gate:** **CLOSED — design/documentation only.**

Git history is the archive. Current code/schema/OpenAPI are current-state and migration evidence, not target authority.

Global maximum = **the smallest professional architecture that correctly models the domain, preserves invariants and exposes clean extension seams.**

```text
product/domain semantics
→ invariants + lifecycle
→ Organization/AuthZ/Approval integration
→ whole-product completion (R9.5)       [FROZEN]
→ bounded contexts / ownership          [R10 NEXT]
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
3. Approval/Rendition/Release bind the same exact immutable `RevisionSubmission` / digest;
4. `DRAFT` is mutable persisted working truth; `SUBMITTED` is the immutable review boundary;
5. no confirmed binary content exists without semantic ownership by a governed business object;
6. Dossier is documentary context, never a hidden ERP/PLM/custom-object platform;
7. retention expiry never means automatic deletion;
8. an active LegalHold prevents disposal of every governed retention subject within its live scope;
9. imported historical truth is never rewritten as if it were native MetalDocs history;
10. projections/jobs/providers may fail without rewriting canonical business truth;
11. no generic BPM/ReBAC/low-code/object-platform engine without proven need;
12. launch scope favors simple invariant-preserving controls over speculative compliance/security platforms.

---

# 1. LOCKED — Authentication / Organization / Authorization (R1–R2, R9 + R9.5 delta)

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

## R9 base catalog — 29 permissions

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

## LOCKED R9.5 bounded delta — 16 permissions

```text
evidence_type.manage

evidence.read
evidence.create
evidence.edit
evidence.capture
evidence.void

dossier_type.manage

dossier.read
dossier.create
dossier.manage

retention.extend
legal_hold.manage
disposition.manage

historical_migration.manage
governed_subject.export
external_repository.publish
```

No new role is introduced. Ordinary `IMPORT_COPY` reuses the normal target-object operations and receives no provider-specific permission.

R9.5 role-bundle additions:

```text
viewer:
  evidence.read
  dossier.read

author:
  evidence.read
  evidence.create
  evidence.edit
  evidence.capture
  dossier.read
  dossier.create
  dossier.manage

approver:
  evidence.read
  dossier.read

area_manager:
  all author R9.5 additions
  evidence.void

tenant_owner:
  all 16 R9.5 permissions
```

Tenant-scoped type management, retention/hold/disposition, Historical Migration, governed-subject export and external publication remain tenant-owner-only in V1 absent a real second consumer. A future dedicated records/export/integration role is a reopen trigger, not a launch TODO.

Strict Approval SoD remains: creator/submitter cannot accept own Submission; the same user cannot accept two Steps of one ApprovalInstance; reassignment must remain qualified and SoD-valid. `tenant_owner` is always a bundle of grants, never a bypass.

Do not create mechanism/provider permissions such as `xlsx.upload`, `docx.edit`, `artifact.replace`, `external_edit`, `storage.migrate`, `sharepoint.import`, `import.copy` or `renderer.retry`.

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

Normal SharePoint/OneDrive/etc. are External Repository Connectors, not ManagedArtifactStore providers. Governed primary content V1 requires exact MetalDocs-managed copy. Connector directions begin `IMPORT_COPY` / `PUBLISH_COPY`. External edits never mutate existing MetalDocs immutable history.

SharePoint Embedded remains a future Microsoft-enterprise content profile. Valid restore = Artifact DB fact + exact bytes + matching SHA-256.

---

# 10. LOCKED — R9.5-3 Authoring / EigenPal, refined by R9.5-8

`WorkingContent` is the format-agnostic mutable persisted authority of an open DRAFT Revision; browser/editor/provider state is never authority. Conceptually:

```text
WorkingContent
= current primary Artifact
+ governed metadata
+ optional structured authoring state
+ working_version
```

`structured_authoring` is absent when the chosen format/authoring model does not require it. EigenPal is one DOCX authoring provider around WorkingContent; it does not define WorkingContent semantics.

DRAFT uses monotonic technical `working_version`; immutable WorkingSnapshots are technical checkpoints, never REVxxx. All governed DRAFT changes share one OCC version. Save/replacement requires the caller's observed `expected_working_version`; no overlapping-operation last-write-wins. V1 uses one active in-app writer + OCC. EditorSession is a narrow heartbeat/staleness authoring lease.

While a Revision remains `DRAFT`, an authorized actor may deliberately replace its current WorkingContent with another supported content state. MetalDocs does **not** track or infer arbitrary file ancestry from prior downloads and does not require long checkout. OCC prevents races between overlapping operations; it does not prohibit a deliberate later replacement after the actor has observed the current DRAFT state.

Replacement is whole-WorkingContent. A mutation replacing the primary Artifact MUST in the same OCC step disposition any structured-authoring state and decision-relevant provenance whose validity depends on that representation; immutable historical/template-seeding provenance remains preserved. No automatic binary DOCX/XLSX merge exists in V1.

Submit requires/follows final successful flush, validates OCC and freezes **one coherent accepted WorkingContent version** into immutable `RevisionSubmission`: primary Artifact, governed metadata and all decision-relevant structured/template provenance MUST correspond to that same state. `SUBMITTED` rejects every later autosave/edit/upload/replacement for that attempt.

Approval review resolves only the exact Submission being decided. Approval UI is read-only; rationale is Approval evidence, not editor state. Editors/viewers/previews/renditions are representations, never content or approval authority. If an inline representation is unsupported or fidelity is not established for governed review, use a supported inspection path for that same exact Submission, including exact-source download where appropriate; do not invent a second canonical review artifact.

`EditorialComment` is product-owned DRAFT collaboration state; unresolved comments and, if enabled, tracked changes block submission V1.

Realtime Yjs/coauthoring/WOPI-style collaboration remains deferred behind a seam. Preserve one EigenPal anti-corruption/provider adapter, exact dependency pin and MetalDocs fidelity corpus. Future Office/ONLYOFFICE providers cannot change core semantics.

---

# 11. LOCKED — R9.5-4 Dossier / Context

`Dossier` = stable documentary context for an identifiable business subject; not physical folder and not ERP/PLM entity.

Tenant-scoped `DossierType` stays small: code/name/description/status + eligible DocumentTypes/EvidenceTypes. No custom fields/forms/workflow/ACL engine and no required-evidence completeness engine V1.

Dossier has stable key unique within tenant+type; title may change. `{DOSSIER}` resolves stable key. No generic Dossier numbering V1.

Creation provenance is separate from zero..N ExternalReferences. ExternalReference uses connection + entity kind + external ID; same external identity cannot map to two Dossiers. No heuristic auto-merge.

External master fields/status remain projections, not canonical Dossier state. Source disappearance never deletes history.

Dossier↔Document is M:N over stable Document identity and never copies content or changes Document lifecycle/Area/AuthZ. Links are documentary context only and **never grant access**. Every linked target is authorized against its own canonical scope/relationship.

Every CAPTURED Evidence has exactly one immutable primary Dossier; DRAFT may correct it subject to authorization on both old/new relevant scopes. Secondary Dossier links are allowed without duplication. Dossier scope = one `TenantScope | AreaScope`; Evidence reuses primary Dossier scope. Dossier type/key/scope are stable V1.

Lifecycle = `ACTIVE ↔ ARCHIVED`; archive is reversible navigation state, not external business status, and never deletes related content.

No Dossier-to-Dossier graph/hierarchy V1. Search/timeline/export projections reapply canonical AuthZ and may not turn contextual links into transitive grants. ERP/CRM, PLM, PM and EAM/CMMS boundaries remain explicit.

---

# 12. LOCKED — R9.5-5 Retention / Records / Legal Hold, refined by R9.5-8

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

Expiry only makes a subject eligible for disposition. No automatic delete. Current EFFECTIVE REV is never disposition-eligible. Physical disposition requires explicit authorized review, no active hold, and verified physical removal of the governed retention unit before immutable DispositionRecord says disposal completed.

LegalHold is independent of retention. V1 hold scopes: Evidence, stable Document, Dossier. Evidence hold targets that subject. An active Document/Dossier hold materializes concrete retention subjects already in scope **and continues materializing newly entering governed retention subjects while the hold remains active and they are within that live scope**. Unlink/lifecycle changes cannot release already-materialized held subjects. A subject created after it has genuinely left a Dossier hold's live scope is not implicitly captured by that Dossier hold; a direct Document/Evidence hold is the existing seam when broader preservation is required.

Holds block disposal, not normal business lifecycle. Dossier hold is documentary-context scope, not a generic custodian/ESI graph. Hold V1 covers confirmed governed record content, not never-submitted DRAFT autosaves/staging/full eDiscovery ESI. Artifact has no independent retention policy; preservation derives from subjects referencing it.

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

Generated export bundle is temporary delivery output, not automatically Evidence. Retention/Hold does not forbid authorized export and export does not release hold/change retention/count as acknowledgement.

Export completeness is explicit and authorization-safe. A contract that claims a complete package MUST fail closed rather than silently omit required linked subjects the actor is not authorized to read. Deliberately partial export, if ever supported, must identify itself as partial and define that contract explicitly rather than masquerading as complete.

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

Explicitly deferred until a concrete customer/security/regulatory trigger:

```text
malware scanning / ClamAV / quarantine / periodic rescans
ArtifactSecurityAssessment / CDR / advanced active-content inspection
ICP-Brasil / PKI / DocuSign / Adobe Sign / RFC3161 / TSA / HSM
cryptographically signed export packages / signing-key lifecycle
custom portable export encryption
macro-enabled Office support
full custom renderer sandbox/egress platform
eDiscovery / ESI preservation
```

The staging→validation→confirmation seam already leaves a natural place to add malware inspection later without redesigning Document/Evidence/Artifact/Submission.

---

# 15. LOCKED — Build-vs-buy / launch non-goals

No external ECM/JCR kernel, generic BPM/ReBAC/low-code engine, mandatory realtime collaboration, generic PLM/ERP/PM/CMMS features, generic multi-cloud/BYOS/cross-tenant dedup/silent sync, provider identity in domain or universal PDF rule.

Additional launch non-goals: long checkout/offline-file ancestry tracking; binary DOCX/XLSX merge or semantic spreadsheet comparison; Office calculation engine; universal renderer; ArtifactPackage before a real indivisible multi-file requirement; Dossier custom fields/forms/workflow/ACL/hierarchy/completeness engine; generic Record declaration workflow; automatic retention deletion; post-termination Retention Vault; generic migration transformation engine; retroactive side-effect replay; raw backup as portability contract; export of deployment secrets; antivirus/quarantine platform; PKI/signature infrastructure; signed-package infrastructure; advanced content-security platform; generic eDiscovery/ESI preservation.

Prepare seams only where a real future trigger is already evidenced; do not implement the future capability now.

---

# 16. LOCKED — R9.5-8 Whole-Product Adversarial Freeze

Review evidence:

- candidate packet: `docs/superpowers/analysis/2026-08-17-r9.5-8-whole-product-adversarial-freeze.md`;
- independent review of record: `docs/superpowers/analysis/2026-08-17-r9.5-8-independent-adversarial-challenge.md`;
- independent verdict: `VERDICT: APPROVE / FREEZE R9.5`;
- operator disposition: **ACCEPTED / RATIFIED — 2026-08-17**.

All 15 mandatory end-to-end scenarios survived independent adversarial attack without a material counterexample. Independent review disposition:

- B1 hold-scope unlink observation — **ACCEPTED / NO CHANGE**: post-scope future subjects are outside a Dossier hold; already-materialized subjects remain held; direct Document/Evidence hold is the existing broader-preservation seam.
- B2 never-submitted DRAFT outside hold — **ACCEPTED / DECLARED DEFERRAL**: eDiscovery/ESI preservation remains an explicit future trigger.
- B3 WorkingContent replacement holism — **APPLIED** in R9.5-3 above: primary-Artifact replacement dispositions dependent structured state/provenance in the same OCC step.

The bounded reopen of the former R9.5-3 phrase “external edit fails on stale base” is resolved and closed. The target does not track arbitrary download ancestry. Mutable-DRAFT replacement + OCC + immutable Submission is the frozen semantics.

The bounded 16-permission authorization delta in §1 is frozen. Final YAGNI/deletion pass found no additional material capability required and no remaining concept removable without weakening a distinct material property.

Reopen set after independent challenge: **EMPTY**.

```text
R9.5-1 = LOCKED
R9.5-2 = LOCKED
R9.5-3 = LOCKED (refined by R9.5-8)
R9.5-4 = LOCKED
R9.5-5 = LOCKED (refined by R9.5-8)
R9.5-6 = LOCKED
R9.5-7 = LOCKED
R9.5-8 = CLOSED / APPROVED
R9.5   = FROZEN
```

Future evidence reopens only the minimal decision actually invalidated under the DevelopmentConexus Engineering Method.

---

# 17. NEXT — R10 Technical Architecture

R10 is now unblocked for **design only**. It must descend from the frozen product/domain semantics rather than rediscover or rewrite them for implementation convenience.

R10 owns the technical realization needed to prove, at minimum:

1. bounded contexts/module ownership and published dependency DAG;
2. filesystem/package ownership and deletion/rename map for legacy modules;
3. target data model, table ownership, keys and DB constraints;
4. transaction boundaries and durable event/outbox contracts;
5. coherent/atomic creation of RevisionSubmission from one accepted WorkingContent version;
6. one OCC guard across every DRAFT-mutating path and DRAFT→SUBMITTED transition;
7. late autosave/upload cannot mutate SUBMITTED truth;
8. idempotent/atomic Release preserving exactly one EFFECTIVE Revision;
9. Artifact staging/confirmation/integrity, provider relocation and restore correctness;
10. LegalHold prospective materialization and verified disposition;
11. retention-aware tenant erasure and restore tombstone reconciliation;
12. canonical AuthZ on cross-scope queries/projections/exports;
13. Historical Migration idempotency/reconciliation without fabricated truth;
14. idempotent external publish/job effects and truthful failure reporting;
15. explicit authorization-safe export completeness;
16. APIs and frontend journeys derived from these authorities rather than defining parallel semantics.

R10 should first perform an integrated decomposition/ownership pass before selecting schemas/files/endpoints. Technology and topology remain subordinate to the frozen invariants.

---

# 18. Implementation gate

**NO PRODUCT IMPLEMENTATION YET.** R10 and subsequent technical design are documentation/design work. Product implementation remains blocked until the integrated technical design is complete, durable target specs/ADRs are promoted as required, material adversarial ambiguity is closed, the operator approves the integrated design, and an implementation plan is authored from that accepted target.