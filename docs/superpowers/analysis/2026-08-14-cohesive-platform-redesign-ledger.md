# MetalDocs Cohesive Platform Redesign — Active Decision Ledger

> **Status:** ACTIVE WIP — operator-approved decisions below are binding; unresolved items are explicit.
> **Date:** 2026-08-15
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
9. imported historical truth is never rewritten as if it were native MetalDocs history.

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

Groups are flat V1; Area is the organizational truth reused by Document ownership, scoped authorization and Approval actor resolution.

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

Final R9 Permission Catalog for the R3–R9 operation set remains 29 permissions:

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

Role bundles remain viewer / author / approver / area_manager / tenant_owner as previously locked. Strict Approval SoD remains: author/submitter cannot accept own Submission; same user cannot accept two Steps in one ApprovalInstance; reassignment must remain qualified and SoD-valid.

**R9.5 delta rule:** Evidence, Dossier, Retention, Import/Export and related operations receive one bounded authorization-catalog delta only after R9.5 operations close; operations first, permissions second.

---

# 2. LOCKED — Approval V1

Specialized sequential human workflow, not BPM:

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

`Document` = stable governed identity. At most one EFFECTIVE + one open Revision V1.

Official labels: `REV001`, `REV002`, ...

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

Document code/type/Area are stable V1. DocumentType numbering language remains literals + `{TYPE}/{AREA}/{SEQ}`, scope TYPE or TYPE_AREA, minimum sequence padding. No generic formula/reset/custom-metadata numbering.

Template is a role of an ordinary governed Document; no parallel TemplateVersion lifecycle. `TemplateUse` is M:N and exact source effective REV is pinned at derived-document creation.

TemplateSpec is optional structured-authoring state for applicable content, not universal Revision content.

---

# 4. LOCKED — Periodic Review / Rendition / Release

Periodic Review belongs Controlled Information: `Disabled | Every(n months)`. Due/overdue does not invalidate EFFECTIVE content. Immutable `PeriodicReviewRecord` binds exact current REV and outcome.

Rendition = immutable derived representation of exact Submission with output hash + generator/build provenance. Approval approves Submission, never renderer output bytes.

Release is automatic/system-owned; no publish button. Optional `ReleasePlan.not_before`; actual `effective_at = released_at`.

Winning release transaction atomically makes candidate EFFECTIVE, prior REV SUPERSEDED, swaps Document pointer, clears open pointer, writes ReleaseRecord and durable events.

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

Deletion request is a separate grace/cancel process. Terminal erasure is retention-aware:

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

No new tenant state for retention. A DEK needed to preserve legally retained intelligible content is never destroyed while that obligation remains. Backup/restore must reapply erasure tombstones and reconcile retention/hold facts before cleanup/service resumes.

---

# 7. LOCKED — R9.5 Whole-Product North Star

> **MetalDocs is the system of record for identity, governance, revision, evidence and documentary context. Physical storage, authoring/editor technology, viewers and upstream ERP/PLM/repositories are replaceable providers/connectors around that kernel.**

Do not replace MetalDocs with M-Files/Nuxeo/Alfresco and do not build a generic ECM clone.

A governed Document is format-agnostic. Dossier is a deliberately small documentary context, not ERP/PLM. EigenPal is a DOCX authoring provider, never Document identity.

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

Temporary/direct-presigned upload is allowed. Provider success does not confirm Artifact before integrity/content/semantic validation. Existing Artifact keys are never overwritten.

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

Document retention clock does not run while REV is EFFECTIVE. Anchor = superseded_at / obsoleted_at / cancelled_at for a submitted-but-never-released REV.

EvidenceType chooses `CAPTURED_AT | OCCURRED_AT`; OCCURRED_AT makes Evidence.occurred_at mandatory before capture. Dossier archive never starts retention.

Policy is snapshotted in RetentionBinding; later type changes do not silently recalculate existing records. `RetentionExtension` may only lengthen retention V1 with reason/authority. Generic retroactive shortening is excluded.

Expiry only makes a subject eligible for disposition. No automatic delete. Current EFFECTIVE REV is never disposition-eligible.

Physical disposition requires explicit authorized review, no active hold, and verified removal of substantive DB payload + managed Artifacts before immutable DispositionRecord says disposal completed.

LegalHold is independent of retention and stores explicit apply/release facts, reason, actor/time and optional case reference. V1 hold scopes: Evidence, stable Document, Dossier. Document/Dossier holds materialize concrete held subjects; unlink/lifecycle changes cannot make previously held content escape preservation. Holds block disposal, not normal business lifecycle.

Hold V1 covers confirmed governed record content, not transient autosaves/staging/full eDiscovery ESI.

Artifact has no independent retention policy; preservation derives from retention subjects that reference it.

Provider WORM/Object Lock/Purview is enforcement only. Deployment may require DB-only, WORM-governance or WORM-compliance posture; provider mode never leaks into DocumentType/EvidenceType. If required physical WORM cannot be established, record-finalization cannot falsely report compliant success.

DossierType has no retention policy. Audit Trail remains a separate retention regime.

Tenant terminal erasure is blocked until relevant RetentionBindings/Holds permit destruction. No new Tenant state. Do not build a post-termination Retention Vault until a real product/economic requirement proves it.

---

# 13. LOCKED — R9.5-6 Import / Migration / Export

## Import vs Historical Migration

Ordinary `IMPORT_COPY` follows normal lifecycle. **Historical Migration** is a privileged path for pre-MetalDocs history and never fabricates native facts.

Every migrated object carries explicit source provenance. Historical lifecycle/approval/effectivity facts are represented as imported governance evidence, **not** synthetic `ApprovalDecision`, `ApprovalInstance`, `ReleaseRecord` or fake internal-user actions.

Imported EFFECTIVE/SUPERSEDED/OBSOLETE states require explicit imported proof. Native states continue to require native mechanisms.

If source effectivity date is unknown, preserve unknown and record explicit `adopted_as_current_at`; never invent a historical date. Adoption time may be used as operational periodic-review basis only when historical effectivity is unavailable, while keeping the distinction visible.

## Revision/code migration

Reliable numeric legacy ordinals may map directly (`7 → REV007`) and current-state-only import may legitimately begin at REV007 with gaps. Arbitrary legacy labels map deterministically to REVxxx order while preserving exact `source_revision_label`. Next native REV is always above highest imported target ordinal.

Historical migration may preserve safe, unique legacy Document codes through its privileged path. If preservation is impossible, mapping is explicit and original source code remains provenance; never silently sanitize.

Migration modes:

```text
CURRENT_STATE
FULL_HISTORY
```

A target DocumentRevision requires exact primary Artifact bytes. Missing historical content may be recorded as imported-history evidence but does not create a fake Revision. Historical bytes are preserved exactly; no silent DOC→DOCX/PDF→DOCX/CAD→PDF normalization. Any future transform preserves original + explicit derivation provenance.

Historical actor/creator identity remains source snapshot/reference. Optional correlation to a current User is provenance only and never turns an external event into a native user action. Migration writes are attributed to Migration/System principal.

## Dossier / Evidence / retention migration

Dossier migration uses stable key + ExternalReference uniqueness and deterministic replay. Historical Evidence separates trustworthy occurred/captured facts from migration time.

Imported RetentionBinding snapshots target policy at import but may use trustworthy historical anchor. Unknown anchor never silently makes content deletion-eligible. Migration never automatically disposes already-expired content.

External hold/retention claims remain provenance unless explicitly adopted as active MetalDocs LegalHold; trusted MetalDocs→MetalDocs portability is the special case that may preserve native authorities.

Migration never replays historical notifications, jobs, Distribution assignments or other old side effects. Current obligations require explicit post-import action.

Legacy template content receives TemplateSpec/TemplateUse authority only after target structured-authoring validation succeeds.

## Batch safety

Historical Migration uses first-class batch/plan semantics with true dry-run, deterministic per-item outcomes and reconciliation report.

Same source identity + same content → REUSE/ALREADY_IMPORTED. Same source identity + conflicting content/state → fail closed.

Atomicity is per semantic import unit (e.g. one Document history or one Evidence), not one giant tenant transaction. Partial batch success is allowed and reconciled. No generic post-commit “rollback whole migration” promise; safety comes from dry-run, fail-closed apply, idempotent replay and explicit correction/expurgo when justified.

## Export distinctions

The following are separate contracts:

```text
Backup
Tenant Portability Export
Governed Subject Export
External Repository PUBLISH_COPY
```

`tenant.export` produces provider-independent tenant business/governance state + exact Artifacts, not raw deployment internals.

Portability includes current DRAFT WorkingContent required to continue business but excludes GC-able historical WorkingSnapshots, staging, Search/cache projections, jobs/outbox and provider delivery internals.

Export excludes password/session/activation secrets, connector credentials, KMS/DEK/KEK material, presigned URLs and deployment secrets.

Governed Subject Export may package a Document/Evidence/Dossier scope independently of backup/tenant-wide export.

Every portability/governed package has a versioned provider-independent manifest describing export identity/time/scope, objects, relationships, provenance, canonical filenames, ContentFormats/sizes and exact SHA-256 values. Manifest never depends on MinIO/S3 keys, URLs or provider version IDs.

Generated export bundle is a temporary delivery artifact, not automatically governed Evidence. Retention/Legal Hold blocks deletion, not authorized export. Export never releases a hold, changes retention or counts as acknowledgement.

Export completeness must be explicit and authorization-safe; a subset cannot masquerade as a complete tenant/Dossier representation.

Trusted compatible MetalDocs→MetalDocs portability may preserve native Approval/Release/Retention/Hold authorities with package/source-deployment provenance. All packages still undergo manifest/hash/schema/conflict validation. Package attestation/signature/encryption is deferred to R9.5-7.

---

# 14. Build-vs-buy / explicit non-goals

Current rulings remain: no external ECM kernel, no JCR kernel, no generic BPM/ReBAC/low-code engine, no mandatory realtime collaboration, no generic PLM/ERP/PM/CMMS features, no generic multi-cloud/BYOS/dedup/silent sync, no provider identity in domain, no universal PDF rule.

Additional non-goals:

- no fake native history during migration;
- no generic migration transformation engine;
- no giant tenant-wide migration transaction;
- no retroactive side-effect replay;
- no raw backup as portability contract;
- no export of secrets/provider internals as customer business data.

---

# 15. Remaining whole-product completion before technical architecture

- [x] R9.5-1 Content Model
- [x] R9.5-2 Storage / Repository Strategy
- [x] R9.5-3 Authoring / EigenPal
- [x] R9.5-4 Dossier / Context
- [x] R9.5-5 Retention / Records / Legal Hold
- [x] R9.5-6 Import / Migration / Export
- [ ] **R9.5-7 Attestation + Content Security — NEXT**
- [ ] R9.5-8 Whole-product adversarial freeze

R9.5-7 must close:

1. exact semantic statement/evidence created by Approval `accept` and `return_for_changes`;
2. whether V1 attestation is an application-level authenticated approval or a formal digital/electronic signature;
3. immutable ApprovalReceipt/decision evidence and how it binds actor + assurance + Submission digest;
4. manifestation of approval/effectivity in official/readable renditions without mutating approved source truth;
5. portability/export package integrity, signature and optional encryption boundary;
6. upload/content quarantine lifecycle versus Document/Evidence business lifecycle;
7. file-type detection, allowed-format validation, size/complexity limits and parser hardening;
8. malware scanning and behavior when scanner is unavailable;
9. risky format features such as Office macros/external relationships, PDF active content and archive bombs;
10. safe preview/view/download policy and Content-Disposition/content-type controls;
11. rendering/conversion sandbox/network policy for untrusted content;
12. security evidence and audit facts without turning telemetry into business authority.

Then R9.5-8 performs the final whole-product adversarial freeze and bounded R9 authorization delta before R10 technical architecture resumes.

---

# 16. Implementation gate

**NO PRODUCT IMPLEMENTATION.** R10 bounded contexts/filesystem/data model remains blocked until R9.5-8 is explicitly approved.