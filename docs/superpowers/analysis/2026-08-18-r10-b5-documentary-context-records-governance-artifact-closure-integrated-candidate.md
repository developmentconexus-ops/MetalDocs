# R10-B5 — Documentary Context + Records Governance + Artifact Closure — Integrated Candidate

> **Status:** NON-AUTHORITATIVE — SELF-REVIEWED CORRECTED CANDIDATE — OPERATOR REVIEW / R10 INTEGRATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Input working baseline:** B1/B2 promoted authority + B3/B4 **ACCEPTED FOR R10 INTEGRATION / NON-FINAL**  
> **Input HEAD:** `942765b92be9400ecb46f0510fa18552e91fbaa2`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Independent review:** deferred to Whole-R10 unless a material exception trigger appears.

This file is staging analysis. It does **not** independently ratify B5, promote B3/B4 to final authority, or silently rewrite the frozen R3–R9.5 ledger. It consumes the current R10 working target, including the operator-approved B4 bounded Approval refinement.

B5 exposes three bounded B3 technical refinements through real Records-Governance counterexamples. They are explicit in §4 and require operator adjudication together with B5. No unrelated B3/B4 semantic is reopened.

---

# 1. Authority and evidence boundary

Authority path:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`
7. accepted non-final B3 candidate
8. accepted corrected non-final B4 candidate + B4 acceptance record

Current code/schema/OpenAPI/legacy module shape remains evidence only.

External references are comparison evidence only:

- M-Files object relationships: <https://userguide.m-files.com/user-guide/latest/eng/object_relationships.html>
- NARA records scheduling/disposition instructions: <https://www.archives.gov/records-mgmt/scheduling/instructions>
- Microsoft Purview disposition: <https://learn.microsoft.com/en-us/purview/disposition>
- AWS S3 Object Lock: <https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock.html>

Comparison signal only:

- relationships can organize documentary context without copying/owning the document;
- retention clocks can anchor on real lifecycle events;
- retention expiry and disposition decision are distinct;
- object-store retention/hold is physical enforcement, not Records-Governance authority.

MetalDocs does not import a generic records-management/ECM policy platform.

---

# 2. Known / Inferred / Unknown / Deferred

## 2.1 Known — frozen/promoted/accepted inputs

### Documentary Context

- `Dossier` = stable documentary context for an identifiable business subject;
- Dossier is not a physical folder and not ERP/CRM/PLM/PM/EAM master data;
- `DossierType` stays small: code/name/description/status + eligible DocumentTypes/EvidenceTypes;
- no Dossier custom fields/forms/workflow/ACL/completeness engine V1;
- Dossier stable key is unique within type; title may change;
- creation provenance is separate from zero..N ExternalReferences;
- ExternalReference uses connection + entity kind + external ID; one external identity cannot map to two Dossiers;
- Dossier↔Document is M:N over stable Document identity and never copies content, changes Document lifecycle/Area/AuthZ or grants access;
- every CAPTURED Evidence has exactly one immutable primary Dossier;
- DRAFT Evidence may correct primary Dossier with authorization on relevant scopes;
- secondary Dossier links are allowed;
- Dossier scope = one `TenantScope | AreaScope`; Evidence reuses primary Dossier scope;
- Dossier type/key/scope stable V1;
- Dossier lifecycle `ACTIVE ↔ ARCHIVED`; archive is reversible navigation state and never starts retention;
- no Dossier graph/hierarchy V1.

### Evidence

- EvidenceType has stable code/name/status, allowed formats and a small naming policy;
- naming tokens = `{TYPE}`, `{DOSSIER}`, `{REF}`, `{SEQ}`;
- user filename is provenance only;
- Evidence lifecycle = `DRAFT → CAPTURED → VOIDED`; VOIDED means invalid MetalDocs capture only;
- CAPTURED content/metadata immutable;
- external-world cancellation is separate;
- Evidence does not use REV/Approval/Release by default;
- exactly one primary Artifact per Evidence V1;
- no multi-file ArtifactPackage without real requirement.

### Records Governance

- no generic `Record` entity/declaration button;
- DocumentRevision gets RetentionBinding at first RevisionSubmission;
- Evidence gets RetentionBinding at CAPTURE;
- never-submitted Draft/staging/recovery state is not a retention subject;
- DocumentRevision retention unit includes governed immutable history: Submissions, Approval evidence, Renditions, Release/PeriodicReview evidence and referenced Artifacts;
- explicit type policy = `NoMinimum | KeepFor(value,DAYS|MONTHS|YEARS) | Indefinite`;
- no NULL-as-policy and no hardcoded legal periods;
- Document clock does not run while Revision is EFFECTIVE;
- Document anchor = superseded / obsoleted / cancelled for a submitted-never-released Revision;
- Evidence anchor = `CAPTURED_AT | OCCURRED_AT` by EvidenceType;
- Dossier archive never starts retention;
- RetentionBinding snapshots policy; later type changes do not recalculate old subjects;
- RetentionExtension only lengthens V1;
- expiry only creates disposition eligibility;
- no automatic delete;
- current EFFECTIVE Revision never disposition-eligible;
- disposition requires explicit authorized review, no active hold and verified removal before immutable DispositionRecord completion;
- LegalHold independent of retention;
- hold scopes V1 = Evidence | stable Document | Dossier;
- Document/Dossier hold materializes current RetentionBindings and future subjects entering live scope;
- unlink/lifecycle change cannot release already-materialized held subjects;
- hold blocks disposal, not normal business lifecycle;
- Artifact has no independent retention policy; preservation derives from governed subject relationships;
- provider WORM/ObjectLock/Purview remains enforcement only.

### B1–B4

- one PostgreSQL DB / `metaldocs`, UUID IDs, typed FKs, RESTRICT/NO ACTION, READ COMMITTED;
- no universal company/tenant partition column;
- live canonical Authorization remains authority;
- links never become grants;
- Artifact = exact bytes, provider-neutral, no confirmed semantic orphan;
- B3 Submission immutable exact candidate;
- B4 Approval/Rendition/Release bind same Submission;
- B4 Release is effectivity authority;
- B4 Distribution obligations/acknowledgements are immutable Revision-governance evidence.

## 2.2 Inferred corrected choices

1. Documentary Context and Records Governance remain separate owners; Dossier never becomes retention authority.
2. Dossier scope reuses B2 typed TenantScope/AreaScope; no magic scope sentinel.
3. DossierType eligibility is explicit current configuration; absence means not eligible for a new relation/capture.
4. Eligibility/status changes are future-facing; existing valid context remains.
5. Evidence separates mutable DRAFT working state from immutable CAPTURE payload.
6. Evidence canonical name is frozen at CAPTURE.
7. If naming uses `{SEQ}`, V1 has one monotonic sequence per EvidenceType; Dossier-local reset is a reopen trigger.
8. Native `OCCURRED_AT` CAPTURE requires known `occurred_at`; R10-F Historical Migration may preserve unknown explicitly and never make it silently disposition-eligible.
9. DocumentTypeRetentionRule and EvidenceTypeRetentionRule are Records-Governance-owned current configuration; no new permission family is added. Existing Tenant-wide `document_type.manage` / `evidence_type.manage` administer the corresponding rule under Records-Governance validation.
10. RetentionBinding is closed typed union DocumentRevision XOR Evidence; no generic subject registry.
11. Binding stores immutable policy snapshot only; anchor derives from canonical lifecycle facts.
12. RetentionExtension is append-only absolute `extended_until`, only after known anchor, only when it lengthens the current minimum, and never after DispositionFence exists.
13. LegalHoldSubject is immutable materialization evidence; release/unlink never deletes rows.
14. Direct Evidence hold requires existing Evidence RetentionBinding. Document/Dossier are future-materializing scopes.
15. Hold materialization and DispositionFence serialize on RetentionBinding; hold creation is all-or-nothing for current preservable scope.
16. DispositionFence is the immutable authorized irreversible barrier; DispositionRecord is written only after verified physical + semantic completion.
17. Business lifecycle and Records disposition are orthogonal; do not add `DISPOSED` to DocumentRevision/Evidence states.
18. Minimal DocumentRevision/Evidence identity skeleton survives disposition; governed retained payload/history may be removed.
19. Records-Governance control rows are not recursively made RetentionBindings V1.
20. DocumentRevision unit additionally includes B4 `SubmissionFeedback`, `SubmissionApprovalRequirement`, `ReleasePlan`, DistributionObligation and AcknowledgementRecord.
21. **One semantic Artifact retention root:** all semantic references to one Artifact must resolve to exactly one `DocumentRevision` or one `Evidence`. Same bytes used by another root get another Artifact semantic row/provenance identity; provider physical dedupe remains mechanism freedom.
22. A confirmed Artifact may have several references inside that one root (e.g. WorkingContent + same-Revision Submission), but cross-root reuse is rejected.
23. Artifact root/reachability is derived from typed references; no owner registry and no ref-count authority.
24. A final-reference removal deletes the Artifact semantic row in the same local semantic transaction; physical bytes may temporarily remain only as R10-C/D cleanup state when not part of governed disposition.

## 2.3 Unknown — kept unknown

- future Dossier-local Evidence sequence restart;
- stricter post-CAPTURE rules for secondary Dossier context changes;
- a future regulated multi-stage disposition-approval workflow;
- exact provider-neutral physical verification manifest shape in R10-C;
- final privacy classification of tiny post-disposition Evidence/Revision/Records-control skeleton fields in B6/R10-F.

These do not block current invariants.

## 2.4 Deferred

```text
Audit privacy + Interchange connection concrete family + final cross-owner matrix → B6
physical object deletion / ObjectLock/provider proof                            → R10-C
async disposition retry/lease/jobs/projections/notifications                   → R10-D
API/frontend journeys                                                           → R10-E
historical migration/cutover/legacy deletion                                   → R10-F
```

---

# 3. Root Cause

B5 prevents **ownership collapse**:

```text
context/folder
+ captured evidence
+ retention policy
+ legal preservation
+ physical deletion
```

must not become one generic repository/records engine.

Failure classes otherwise include:

- Dossier becoming content owner or access-grant source;
- Evidence forced through Document REV/Approval semantics;
- retention stored on Artifact/provider object;
- LegalHold reduced to ObjectLock flag;
- expiry becoming delete;
- polymorphic owner/subject registries;
- provenance FKs creating unintended indefinite retention;
- disposition deleting Revision identity and allowing REV-number reuse.

---

# 4. B3 bounded refinements revealed by B5

## 4.1 DocumentOrigin must not retain source-template retention unit forever

Accepted B3 working target used strong FK to source `RevisionSubmission`.

Real counterexample:

```text
Template REV004 / Submission S4 creates derived Document D
→ S4 later becomes lawfully disposition-eligible
→ surviving D.DocumentOrigin FK prevents deleting S4/Artifact forever
```

That invents indefinite template-source retention.

Corrected working shape:

```text
DocumentOrigin
  derived_document_id UUID PRIMARY KEY FK Document(id) RESTRICT
  source_template_revision_id UUID NOT NULL FK DocumentRevision(id) RESTRICT
  source_submission_id_snapshot UUID NOT NULL
  source_submission_digest BYTEA NOT NULL CHECK octet_length(...)=32
  source_artifact_sha256 BYTEA NOT NULL CHECK octet_length(...)=32
  source_content_format TEXT NOT NULL
  created_at TIMESTAMPTZ NOT NULL
```

Derived creation still validates the exact current-effective source Submission under B3/B4 serialization. Immutable provenance is copied once. Source Submission/Artifact may later be disposed without rewriting derived history.

## 4.2 DocumentRevision needs canonical terminal timestamps

B3 state alone cannot supply retention anchor and Audit cannot become domain authority.

Refinement:

```text
DocumentRevision
  ...
  cancelled_at TIMESTAMPTZ NULL
  obsoleted_at TIMESTAMPTZ NULL
```

```text
CANCELLED → cancelled_at required, obsoleted_at NULL
OBSOLETE  → obsoleted_at required, cancelled_at NULL
other native states → both NULL
```

Written exactly once with the CI state transition.

Native SUPERSEDED timestamp is **not duplicated**; canonical anchor is B4 `ReleaseRecord.released_at` where `prior_effective_revision_id` is this Revision.

## 4.3 Permanent Revision identity must not permanently retain dictionary payload

B3 currently co-locates immutable `dictionary_snapshot` with the Revision row that must survive to prevent ordinal reuse.

Counterexample:

```text
dispose REV002 unit
→ delete Revision row: REV002 may be reused
→ keep row: dictionary snapshot retained forever
→ clear JSON: immutable history rewritten
```

Refinement:

```text
DocumentRevision
  id / document_id / revision_no / state / creator / lifecycle timestamps

RevisionDictionarySnapshot
  revision_id UUID PRIMARY KEY FK DocumentRevision(id) RESTRICT
  snapshot JSONB NOT NULL CHECK jsonb_typeof(...)='object'
```

Snapshot is immutable while present, enters Submission manifest, and belongs to the Revision retention unit. Completed disposition may remove it while Revision identity skeleton survives.

No other B3/B4 semantic is reopened.

---

# 5. Target Invariant

> **Dossier supplies documentary context without ownership or grants. Evidence supplies one small DRAFT→CAPTURE boundary with one exact primary Artifact. Records Governance automatically binds first-submitted DocumentRevisions and captured Evidence to immutable policy snapshots, materializes active LegalHolds over exact RetentionBindings, and permits irreversible disposition only through a serialized explicit fence with no active/applicable hold and verified removal. Every Artifact has exactly one semantic retention root, derived from typed references.**

```text
Dossier != storage folder
Dossier link != access grant
Evidence != DocumentRevision
CAPTURE != Approval/Release
RetentionBinding != Record aggregate
LegalHold != ObjectLock
expiry != delete
DispositionFence != job state
DispositionRecord != AuditEvent
Artifact root != owner_type/id row
Records disposition != business lifecycle state
```

---

# 6. Alternatives / Global Maximum

## A — Dossier folder/binder + generic Record aggregate

Reject. Context becomes ownership; Records becomes a platform.

## B — generic ArtifactOwner/RetentionSubject registries

Reject. Fewer tables, more authority ambiguity and extensibility machinery.

## C — typed Dossier + specialized Evidence + closed-union Binding + materialized Hold + explicit Fence

**Recommended Global Maximum.** It models only the real lifecycle classes and keeps provider deletion/jobs outside semantic authority.

---

# 7. Documentary Context

## 7.1 DossierType

```text
DossierType
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  name TEXT NOT NULL
  description TEXT NULL
  status TEXT NOT NULL CHECK ACTIVE|INACTIVE
```

Code immutable. INACTIVE blocks new Dossiers and new eligibility-dependent relations; existing context remains.

## 7.2 Explicit eligibility

```text
DossierTypeDocumentType(dossier_type_id FK, document_type_id FK)
  PRIMARY KEY(dossier_type_id,document_type_id)

DossierTypeEvidenceType(dossier_type_id FK, evidence_type_id FK)
  PRIMARY KEY(dossier_type_id,evidence_type_id)
```

Current configuration only; removal never rewrites existing links/captures.

## 7.3 Dossier

```text
Dossier
  id UUID PRIMARY KEY
  dossier_type_id UUID NOT NULL FK DossierType(id) RESTRICT
  stable_key TEXT NOT NULL
  title TEXT NOT NULL
  tenant_scope_id UUID NULL FK Tenant(id) RESTRICT
  area_scope_id UUID NULL FK Area(id) RESTRICT
  archived_at TIMESTAMPTZ NULL
  created_by_user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL

  UNIQUE(dossier_type_id,stable_key)
```

Exactly one Tenant/Area scope. Type/key/scope immutable; title mutable; archive reversible. No physical path/hierarchy.

## 7.4 DossierDocumentLink

```text
DossierDocumentLink
  dossier_id UUID NOT NULL FK Dossier(id) RESTRICT
  document_id UUID NOT NULL FK Document(id) RESTRICT
  linked_by_user_id UUID NOT NULL FK User(id) RESTRICT
  linked_at TIMESTAMPTZ NOT NULL
  PRIMARY KEY(dossier_id,document_id)
```

Add requires current eligibility and canonical authorization on both sides. Link/unlink never grants access and never removes previously materialized hold subjects.

Cross-scope links are allowed if independently authorized; context is not scope inheritance.

## 7.5 DossierExternalReference

Relationship owned by Documentary Context; B6 supplies concrete typed Interchange connection family:

```text
DossierExternalReference
  id UUID PRIMARY KEY
  dossier_id UUID NOT NULL FK Dossier(id) RESTRICT
  connection_id UUID NOT NULL FK InterchangeConnection(id) RESTRICT // B6 family
  entity_kind TEXT NOT NULL
  external_id TEXT NOT NULL
  created_at TIMESTAMPTZ NOT NULL
  UNIQUE(connection_id,entity_kind,external_id)
```

Normal serving mapping immutable. External disappearance/status never deletes Dossier history.

---

# 8. Evidence

## 8.1 EvidenceType + formats

```text
EvidenceType
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  name TEXT NOT NULL
  description TEXT NULL
  status TEXT NOT NULL CHECK ACTIVE|INACTIVE
  naming_pattern TEXT NOT NULL

EvidenceTypeAllowedFormat
  evidence_type_id UUID NOT NULL FK EvidenceType(id) RESTRICT
  content_format TEXT NOT NULL CHECK closed ContentFormat
  PRIMARY KEY(evidence_type_id,content_format)
```

Code immutable. Closed naming grammar = literals + `{TYPE}/{DOSSIER}/{REF}/{SEQ}` only. Existing captured Evidence unaffected by later config changes.

## 8.2 EvidenceSequence

```text
EvidenceSequence
  evidence_type_id UUID PRIMARY KEY FK EvidenceType(id) RESTRICT
  next_value BIGINT NOT NULL CHECK next_value >= 1
```

Used only when `{SEQ}` occurs. Candidate V1 scope = EvidenceType across deployment. Committed values never reuse. Dossier-local reset is a reopen trigger.

## 8.3 Evidence identity/lifecycle

```text
Evidence
  id UUID PRIMARY KEY
  evidence_type_id UUID NOT NULL FK EvidenceType(id) RESTRICT
  status TEXT NOT NULL CHECK DRAFT|CAPTURED|VOIDED
  primary_dossier_id UUID NULL FK Dossier(id) RESTRICT
  canonical_name TEXT NULL
  sequence_no BIGINT NULL CHECK sequence_no IS NULL OR sequence_no >= 1
  created_by_user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL
```

DRAFT may change primary Dossier. CAPTURE freezes primary Dossier/name/sequence. After capture those fields immutable. `canonical_name` unique when non-NULL.

## 8.4 EvidenceDraft

```text
EvidenceDraft
  evidence_id UUID PRIMARY KEY FK Evidence(id) RESTRICT
  primary_artifact_id UUID NULL FK Artifact(id) RESTRICT
  reference_value TEXT NULL
  governed_metadata JSONB NOT NULL CHECK jsonb_typeof(...)='object'
  occurred_at TIMESTAMPTZ NULL
  original_filename TEXT NULL
  updated_by_user_id UUID NOT NULL FK User(id) RESTRICT
  updated_at TIMESTAMPTZ NOT NULL
```

Mutable only while DRAFT. Whole primary-content replacement; no Evidence REV/autosave lifecycle.

## 8.5 EvidenceCapture

```text
EvidenceCapture
  evidence_id UUID PRIMARY KEY FK Evidence(id) RESTRICT
  primary_artifact_id UUID NOT NULL FK Artifact(id) RESTRICT
  reference_value TEXT NULL
  governed_metadata JSONB NOT NULL CHECK jsonb_typeof(...)='object'
  occurred_at TIMESTAMPTZ NULL
  captured_by_user_id UUID NOT NULL FK User(id) RESTRICT
  captured_at TIMESTAMPTZ NOT NULL
  original_filename TEXT NULL
```

Immutable while retained.

Native CAPTURE in one local transaction:

1. lock Evidence/Draft;
2. validate primary Dossier + DossierTypeEvidenceType eligibility;
3. lock primary/current-secondary Dossiers in UUID order;
4. validate exact Artifact format;
5. require occurred_at when retention anchor is OCCURRED_AT;
6. allocate `{SEQ}` if required; compute globally unique canonical name;
7. establish exact Artifact reference rooted at this Evidence;
8. insert immutable EvidenceCapture;
9. freeze Evidence primary Dossier/name/sequence + state CAPTURED;
10. delete EvidenceDraft;
11. create RetentionBinding;
12. materialize active Dossier holds;
13. commit.

## 8.6 EvidenceVoidRecord

```text
EvidenceVoidRecord
  evidence_id UUID PRIMARY KEY FK Evidence(id) RESTRICT
  voided_by_user_id UUID NOT NULL FK User(id) RESTRICT
  reason TEXT NOT NULL
  voided_at TIMESTAMPTZ NOT NULL
```

Immutable. VOID never mutates EvidenceCapture.

## 8.7 Secondary Dossier context

```text
EvidenceSecondaryDossierLink
  evidence_id UUID NOT NULL FK Evidence(id) RESTRICT
  dossier_id UUID NOT NULL FK Dossier(id) RESTRICT
  linked_by_user_id UUID NOT NULL FK User(id) RESTRICT
  linked_at TIMESTAMPTZ NOT NULL
  PRIMARY KEY(evidence_id,dossier_id)
```

Cannot duplicate primary Dossier. Current contextual add/remove; does not widen Evidence scope or access. If captured Evidence enters an active held Dossier, materialize its Binding; unlink never removes prior materialization.

---

# 9. Retention Policy Configuration

```text
DocumentTypeRetentionRule
  document_type_id UUID PRIMARY KEY FK DocumentType(id) RESTRICT
  mode TEXT NOT NULL CHECK NO_MINIMUM|KEEP_FOR|INDEFINITE
  value INTEGER NULL
  unit TEXT NULL CHECK NULL OR DAYS|MONTHS|YEARS

EvidenceTypeRetentionRule
  evidence_type_id UUID PRIMARY KEY FK EvidenceType(id) RESTRICT
  mode TEXT NOT NULL CHECK NO_MINIMUM|KEEP_FOR|INDEFINITE
  value INTEGER NULL
  unit TEXT NULL CHECK NULL OR DAYS|MONTHS|YEARS
  anchor_kind TEXT NOT NULL CHECK CAPTURED_AT|OCCURRED_AT
```

Shape law:

```text
NO_MINIMUM/INDEFINITE → value/unit NULL
KEEP_FOR               → value > 0 + unit required
```

No DossierType retention policy.

Every type that can produce a new retention subject needs an explicit rule; missing rule fails first SUBMIT/CAPTURE. Current rule mutations use existing Tenant-wide type-management permission, not a new retention-policy permission.

---

# 10. RetentionBinding

```text
RetentionBinding
  id UUID PRIMARY KEY
  document_revision_id UUID NULL FK DocumentRevision(id) RESTRICT
  evidence_id UUID NULL FK Evidence(id) RESTRICT
  policy_mode TEXT NOT NULL CHECK NO_MINIMUM|KEEP_FOR|INDEFINITE
  policy_value INTEGER NULL
  policy_unit TEXT NULL
  evidence_anchor_kind TEXT NULL CHECK NULL OR CAPTURED_AT|OCCURRED_AT
  bound_at TIMESTAMPTZ NOT NULL
```

Exactly one DocumentRevision/Evidence subject.

```text
UNIQUE(document_revision_id) WHERE document_revision_id IS NOT NULL
UNIQUE(evidence_id) WHERE evidence_id IS NOT NULL
```

Immutable policy snapshot.

Timing:

- first RevisionSubmission creates binding if absent;
- same-REV resubmission never resnapshots;
- CAPTURE creates Evidence binding;
- never-submitted Revision/Evidence Draft has no binding.

Anchor derives from canonical facts:

```text
Document SUPERSEDED → B4 ReleaseRecord.released_at where prior_effective_revision_id = subject
Document OBSOLETE   → DocumentRevision.obsoleted_at
Document CANCELLED  → DocumentRevision.cancelled_at (binding already exists)
DRAFT/SUBMITTED/EFFECTIVE → no running clock

Evidence CAPTURED_AT → EvidenceCapture.captured_at
Evidence OCCURRED_AT → EvidenceCapture.occurred_at
```

Binding never duplicates anchor timestamp.

---

# 11. RetentionExtension

```text
RetentionExtension
  id UUID PRIMARY KEY
  retention_binding_id UUID NOT NULL FK RetentionBinding(id) RESTRICT
  extended_until TIMESTAMPTZ NOT NULL
  reason TEXT NOT NULL
  created_by_user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL
```

Append-only. Requires known anchor, non-Indefinite policy, no DispositionFence, and `extended_until` later than current base/extended minimum.

Effective due = max(base due, all extensions). No shortening.

`retention.extend` remains Tenant-wide tenant-owner-only even for Area-scoped subject.

---

# 12. LegalHold

## 12.1 Hold scope

```text
LegalHold
  id UUID PRIMARY KEY
  evidence_id UUID NULL FK Evidence(id) RESTRICT
  document_id UUID NULL FK Document(id) RESTRICT
  dossier_id UUID NULL FK Dossier(id) RESTRICT
  reason TEXT NOT NULL
  created_by_user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL
  released_by_user_id UUID NULL FK User(id) RESTRICT
  released_at TIMESTAMPTZ NULL
  release_reason TEXT NULL
```

Exactly one scope. Scope/creation facts immutable; release terminal for that row. `legal_hold.manage` Tenant-wide.

Direct Evidence hold requires existing RetentionBinding. Document/Dossier holds can capture future subjects.

## 12.2 Materialized held subjects

```text
LegalHoldSubject
  legal_hold_id UUID NOT NULL FK LegalHold(id) RESTRICT
  retention_binding_id UUID NOT NULL FK RetentionBinding(id) RESTRICT
  materialized_at TIMESTAMPTZ NOT NULL
  PRIMARY KEY(legal_hold_id,retention_binding_id)
```

Rows never disappear due to unlink/release. Active blocking = hold unreleased + materialization row.

Document hold materializes all current Revision bindings and future first-submitted Revision bindings.

Dossier hold materializes current non-disposed bindings for:

- Revisions of linked Documents;
- captured Evidence whose primary Dossier is this Dossier;
- captured Evidence with this active secondary link;

and materializes future subjects entering live scope.

A completed DispositionRecord means the old binding no longer has preservable payload and is not newly materialized by a later hold. An **active DispositionFence without completed record** causes hold activation/materialization to fail, never partial-best-effort.

Hold activation is all-or-nothing for the current preservable scope: if any current subject is fenced/inconsistent, the whole new Hold creation rolls back visibly.

---

# 13. Hold/Scope Concurrency and Lock Order

B1 remains READ COMMITTED.

Canonical order for operations spanning context + retention:

```text
subject root (Document or Evidence)
→ ordered Dossier rows
→ ordered RetentionBinding rows
→ Hold/Fence child rows
```

Dossier-scoped hold creation starts at Dossier and may lock Bindings, but never acquires Document/Evidence roots afterwards.

### First Document Submission

```text
lock Document
load + lock currently linked Dossiers ascending
perform B3/B4 Submission snapshots
if first submission:
  create Binding
  materialize direct Document holds
  materialize active locked-Dossier holds
```

### DossierDocumentLink add

```text
lock Document
lock Dossier
insert link
lock existing Document Revision Bindings ascending
materialize active Dossier holds
```

### Evidence CAPTURE

```text
lock Evidence
lock primary/secondary Dossiers ascending
create Capture + Binding
materialize active Dossier holds
```

### Secondary Evidence link add

```text
lock Evidence
lock target Dossier
insert link
if Binding exists: lock it + materialize active Dossier holds
```

### Dossier Hold create

```text
lock Dossier
insert Hold
resolve current non-disposed in-scope Bindings
lock Bindings ascending
if any active Fence without completion → rollback whole Hold
insert materializations
```

Exact SQL modes/full B1–B5 wait-for proof remains implementation-spec obligation. No reverse order may be introduced for convenience.

---

# 14. Disposition Eligibility + Fence

## 14.1 Eligibility

DocumentRevision eligible only if:

```text
state ∈ SUPERSEDED|OBSOLETE|CANCELLED
anchor known
policy != INDEFINITE
now >= max(base due, extensions)
no active/applicable Hold
no DispositionFence
```

DRAFT/SUBMITTED/EFFECTIVE never eligible.

Evidence eligible only if:

```text
state ∈ CAPTURED|VOIDED
anchor known
policy != INDEFINITE
now >= max(base due, extensions)
no active/applicable Hold
no DispositionFence
```

`NO_MINIMUM` means base due = anchor, never auto-delete.

## 14.2 DispositionFence

```text
DispositionFence
  id UUID PRIMARY KEY
  retention_binding_id UUID NOT NULL UNIQUE FK RetentionBinding(id) RESTRICT
  authorized_by_user_id UUID NOT NULL FK User(id) RESTRICT
  reason TEXT NOT NULL
  eligibility_anchor_at TIMESTAMPTZ NOT NULL
  eligibility_due_at TIMESTAMPTZ NOT NULL
  created_at TIMESTAMPTZ NOT NULL
```

Immutable explicit decision + irreversible semantic barrier.

Fence transaction:

```text
lock subject root
lock RetentionBinding
recompute lifecycle/anchor/policy/extensions
reconcile existing materialized holds
re-evaluate active direct Document/Evidence/Dossier scopes that currently apply
materialize any missing applicable active hold and FAIL
prove zero active/applicable hold
prove eligible
insert Fence
commit
```

Thus disposition never relies solely on “materialization happened earlier”; it fails closed if a live applicable Hold exists.

After Fence:

- no RetentionExtension;
- no new hold materialization over this subject;
- async/provider retries may proceed but cannot rewrite authorization;
- if physical deletion partly fails, Fence remains until recovery/finalization.

No generic disposition workflow state machine.

---

# 15. DispositionRecord and Physical Choreography

```text
DispositionRecord
  id UUID PRIMARY KEY
  disposition_fence_id UUID NOT NULL UNIQUE FK DispositionFence(id) RESTRICT
  completed_at TIMESTAMPTZ NOT NULL
  verification_schema TEXT NOT NULL
  verification_manifest JSONB NOT NULL CHECK jsonb_typeof(...)='object'
```

Immutable.

Correct choreography:

```text
1. B5 Fence transaction
   explicit authorize + fail-closed eligibility/hold check

2. R10-C/D mechanism phase
   enumerate exact Artifact set rooted at subject
   delete provider objects
   verify bytes are no longer retrievable under ManagedArtifactStore contract
   retry/reconcile as required

3. B5 finalization transaction
   lock subject + Binding/Fence
   prove same Fence + physical verification
   delete retention-unit subordinate semantic rows deterministically
   delete rooted Artifact semantic rows whose final refs disappear
   insert immutable DispositionRecord
   commit
```

No provider call is inside PostgreSQL transaction. If physical deletion succeeds and DB finalization fails, Fence remains; finalization is retried. No DispositionRecord is fabricated from a queued job/one DELETE response/expiry.

---

# 16. DocumentRevision Retention Unit

For one DocumentRevision, retained unit includes:

### B3

- `RevisionDictionarySnapshot`;
- all RevisionSubmissions/manifests/digests;
- exact Submission source Artifacts;
- revision-owned immutable submitted/template structured provenance where not solely represented in manifest;
- PeriodicReviewRecords;
- other immutable Revision-governance evidence.

Operational/non-authoritative WorkingContent/EditorSession/never-authoritative snapshots are not retained record history and should already be GC-able when terminal.

### B4

- SubmissionApprovalRequirement;
- ApprovalInstance / StepInstances / Participants / StepDecisions / reassignment evidence;
- SubmissionFeedback;
- ReleasePlan;
- Renditions + output Artifacts;
- ReleaseRecord for this candidate Revision;
- DistributionObligations created by its Release;
- AcknowledgementRecords;
- bounded fresh-auth decision evidence.

Shared configuration/identity such as ApprovalPolicyVersion, DocumentType and User skeleton is not automatically swallowed into the unit.

Distribution evidence follows the Revision because it has no independent governed subject after that Revision/Release history is disposed.

---

# 17. Evidence Retention Unit and Post-Disposition Skeleton

Evidence retained payload includes:

- EvidenceCapture;
- exact primary Artifact;
- EvidenceVoidRecord if present;
- captured metadata/provenance/original filename.

After completed disposition:

```text
Evidence stable identity/lifecycle skeleton remains
EvidenceCapture absent
EvidenceVoidRecord may be removed with retained payload
DispositionRecord + RetentionBinding remain Records evidence
```

Do not add `DISPOSED` to Evidence lifecycle; Records disposition is an orthogonal axis.

DocumentRevision row likewise survives disposition to preserve ordinal/non-reuse and stable references. Do not add `DISPOSED` to Revision state.

The exact tiny surviving Evidence/Revision/Records field set receives B6/R10-F privacy-minimization challenge before final R10 ratification.

Records-control rows (`RetentionBinding`, holds, extensions, fence, disposition record) are not recursively retention-bound in V1.

---

# 18. Artifact Global Closure — Single Semantic Retention Root

## 18.1 Typed reference surface

At B5 the admitted semantic Artifact refs are:

```text
WorkingContent.primary_artifact_id      → root = DocumentRevision
RevisionSubmission.source_artifact_id   → root = DocumentRevision
Rendition.output_artifact_id            → root = RevisionSubmission → DocumentRevision
EvidenceDraft.primary_artifact_id       → root = Evidence
EvidenceCapture.primary_artifact_id     → root = Evidence
```

No generic `ArtifactOwner`, `owner_type/id`, ref-count or Artifact retention fields.

## 18.2 Root law

For every Artifact row:

```text
all existing semantic references MUST resolve to exactly one root:
  one DocumentRevision
  OR one Evidence
```

A cross-root reference attempt fails.

Consequences:

- same Artifact may be referenced by WorkingContent and multiple same-Revision Submissions when bytes truly remain unchanged;
- same bytes reused by a new Revision/Evidence get a **new Artifact semantic row/provenance identity**;
- provider may deduplicate storage physically without changing Artifact identity;
- disposition cannot race a later cross-subject attachment because such attachment is structurally illegal.

## 18.3 No-orphan / root guard

Implementation must have one mechanically verified closed typed-reference catalog plus deferred DB enforcement (or equivalent all-path control):

```text
committed Artifact has zero semantic refs     → fail
refs resolve to >1 semantic retention root   → fail
```

Enforcement catalog is mechanism metadata, not business authority, and must parity-test against actual typed FK surface.

If a normal DRAFT replacement removes the final reference, the same local semantic transaction deletes the Artifact row; provider cleanup follows asynchronously.

For Records disposition, all semantic refs for the fenced root are removed together after verified provider deletion, then Artifact rows are deleted.

---

# 19. Authorization / B2 Coherence

### Dossier

`dossier.read/create/manage` target Dossier's typed scope + domain relationship/state. Link never grants target access.

### Evidence

`evidence.read/create/edit/capture/void` target primary Dossier scope. Secondary links never widen authority. DRAFT primary move requires authorization on both old/new relevant scopes.

### Records Governance

Frozen permissions remain Tenant-wide tenant-owner-only even when subject is Area-scoped:

```text
retention.extend
legal_hold.manage
disposition.manage
```

No `artifact.delete`, `storage.purge`, `objectlock.manage`, `retention.retry` mechanism permissions.

Type retention rules use existing Tenant-wide `document_type.manage` / `evidence_type.manage`; no new permission is invented.

---

# 20. Persistence Class × Mutation Law

| Family | Owner | Law |
|---|---|---|
| DossierType | Documentary Context | current config; code immutable |
| DossierType eligibility joins | Documentary Context | current config add/remove |
| Dossier | Documentary Context | stable type/key/scope; title/archive mutable |
| DossierDocumentLink | Documentary Context | current context add/remove |
| DossierExternalReference | Documentary Context | stable external mapping |
| EvidenceType | Documentary Context | current config; code immutable |
| EvidenceTypeAllowedFormat | Documentary Context | current config |
| EvidenceSequence | Documentary Context | monotonic mechanism |
| Evidence | Documentary Context | stable identity + lifecycle |
| EvidenceDraft | Documentary Context | mutable DRAFT state |
| EvidenceCapture | Documentary Context | immutable retained payload |
| EvidenceVoidRecord | Documentary Context | immutable retained evidence |
| EvidenceSecondaryDossierLink | Documentary Context | current context |
| type RetentionRules | Records Governance | current future-facing policy |
| RetentionBinding | Records Governance | immutable policy snapshot |
| RetentionExtension | Records Governance | append-only lengthening |
| LegalHold | Records Governance | stable scope + terminal release |
| LegalHoldSubject | Records Governance | immutable materialization |
| DispositionFence | Records Governance | immutable irreversible barrier |
| DispositionRecord | Records Governance | immutable completion evidence |
| Artifact | Artifact | immutable exact-byte identity while present |

---

# 21. Structural Constraint Envelope

```text
DossierType.code                                      UNIQUE
Dossier(dossier_type_id,stable_key)                   UNIQUE
Dossier scope                                         XOR Tenant/Area
eligibility/link pairs                                PK
DossierExternalReference(connection,kind,external_id) UNIQUE

EvidenceType.code                                     UNIQUE
EvidenceTypeAllowedFormat pair                        PK
EvidenceSequence.evidence_type_id                     PK
Evidence.canonical_name                               UNIQUE WHERE NOT NULL
EvidenceDraft.evidence_id                             PK
EvidenceCapture.evidence_id                           PK
EvidenceVoidRecord.evidence_id                        PK
EvidenceSecondaryDossierLink pair                     PK

DocumentTypeRetentionRule.document_type_id            PK
EvidenceTypeRetentionRule.evidence_type_id            PK
RetentionBinding subject                              XOR Revision/Evidence
RetentionBinding.document_revision_id                 partial UNIQUE
RetentionBinding.evidence_id                          partial UNIQUE
LegalHold scope                                       XOR Evidence/Document/Dossier
LegalHoldSubject(hold,binding)                         PK
DispositionFence.retention_binding_id                 UNIQUE
DispositionRecord.disposition_fence_id                UNIQUE
```

Cross-row enforcement must prove:

- eligibility on new Dossier relations/capture;
- Evidence primary Dossier/name/sequence immutable after CAPTURE;
- CAPTURED/VOIDED has EvidenceCapture unless completed disposition exists;
- secondary Dossier != primary;
- native OCCURRED_AT CAPTURE has occurred_at;
- first Submission/CAPTURE cannot succeed without Binding;
- resubmit never resnapshots Binding;
- extension only lengthens and cannot cross Fence;
- hold materialization matches scope and cannot cross active Fence;
- Fence cannot exist while an active/applicable Hold protects subject;
- DispositionRecord requires verified physical/final semantic completion;
- Artifact cannot commit orphaned or cross-root.

No cross-owner CASCADE/SET NULL.

---

# 22. Core Transaction Contracts

## First Document Submission

```text
lock Document + existing B3 Revision/WorkingContent roots
load explicit DocumentTypeRetentionRule
load + lock linked Dossiers ascending
B3 freeze RevisionSubmission
B4 snapshot Approval/Release requirements
if first Submission:
  insert RetentionBinding
  materialize active direct Document holds
  materialize active locked-Dossier holds
commit all semantic facts together
```

Same-REV resubmit creates new Submission/B4 attempt but no new Binding/policy snapshot.

## Evidence CAPTURE

As §8.5: capture + exact Artifact ref + Binding + current Dossier-hold materialization in one local commit.

## Hold creation

Lock stable scope root, lock exact current non-disposed Bindings, fail whole operation if any active Fence blocks preservability, insert materializations atomically.

## RetentionExtension

Lock Binding, resolve anchor/due, prove no Fence and strictly later due, append extension.

## DispositionFence

Lock subject + Binding, reconcile materialized and currently applicable active holds, recompute eligibility, insert Fence.

## Disposition completion

Provider deletion/verification outside DB, then one final semantic deletion + Artifact-row deletion + DispositionRecord commit.

---

# 23. Cross-Stage Findings / Coherence

### C1 — return-to-DRAFT after first Submission

Binding remains; no resnapshot; clock not running. **Pass.**

### C2 — source template disposed while derived Document survives

Hard FK to source Submission creates accidental indefinite retention. **Bounded B3 refinement §4.1.**

### C3 — Revision ordinal vs dictionary payload

Permanent row + payload co-location defeats either non-reuse or disposition. **Bounded B3 refinement §4.3.**

### C4 — cancelled/obsolete retention anchor

Audit cannot own domain timestamp. **Bounded B3 refinement §4.2.**

### C5 — superseded anchor

Use existing B4 ReleaseRecord timestamp; no duplicate authority. **Pass.**

### C6 — B4 evidence disposed separately

Would orphan/incompletely retain Revision governance. **Closure: follows Revision unit.**

### C7 — hold misses concurrently entering subject

Stable subject/Dossier roots + standardized order + materialization transaction make it linearizable. **Proof required.**

### C8 — hold arrives during disposition

Binding serialization + reconciled live scopes + Fence make race fail closed. **Pass at design level.**

### C9 — Artifact cross-root sharing races deletion

**Self-review finding:** cross-root reuse is removed. One Artifact has one semantic retention root; identical bytes in another root use another Artifact row. Provider dedupe remains transparent.

### C10 — delete whole Revision row

Would allow ordinal reuse. Revision skeleton survives. **Pass.**

### C11 — ObjectLock treated as LegalHold authority

Cannot represent Document/Dossier/future subject scope. Provider remains mechanism. **Pass.**

---

# 24. Proof Obligations

| Claim | Falsification/proof |
|---|---|
| Dossier link not grant | actor with Dossier access only cannot read linked target |
| Dossier key | duplicate within type fails; same key different type allowed |
| eligibility future-only | existing context survives config removal; new relation fails |
| capture immutable | every post-CAPTURE payload update path rejected |
| one Artifact root | cross-Revision/Evidence reuse of same Artifact ID rejected |
| same-root reuse | same-Revision WorkingContent→Submission reference allowed |
| native occurred anchor | OCCURRED_AT native capture without occurred_at fails |
| sequence concurrency | concurrent `{SEQ}` capture gets distinct committed values |
| first-submission binding | no first Submission commits without Binding |
| no resnapshot | same-REV resubmit after policy edit retains original rule |
| first-capture binding | no CAPTURE commits without Binding |
| EFFECTIVE no clock | current effective Revision never disposition-eligible |
| superseded anchor | due derives from ReleaseRecord time |
| extension | cannot shorten / use unknown anchor / cross Fence |
| hold current scope | creation materializes every current preservable Binding |
| hold future scope | new Submission/CAPTURE/link materializes while Hold active |
| hold activation atomic | one fenced subject rolls back whole new Hold creation |
| unlink | prior materialization survives unlink |
| hold/fence race | exactly one ordering wins; no silent disposal under active hold |
| no auto-delete | expiry alone causes no deletion |
| disposition truth | no record before provider proof + semantic finalization |
| ordinal non-reuse | disposed Revision row survives; next ordinal never reuses |
| source origin | derived origin survives source Submission/Artifact disposal |
| dictionary disposal | snapshot removable without Revision mutation |
| Artifact orphan | committed zero-ref Artifact rejected |
| lifecycle orthogonality | disposition adds no fake DISPOSED business state |
| tenancy posture | no universal tenant/object registry re-enters B5 |

---

# 25. Adversarial Challenge

- **Dossier as folder:** reject; context is not ownership/location.
- **Evidence as Document:** reject; REV/Approval/Release is accidental complexity for capture.
- **Generic Record table:** reject; creates unused declaration/subject platform.
- **Retention on Artifact:** reject; business subject owns policy.
- **Expiry→delete:** reject; hold/review/physical proof still required.
- **ObjectLock = Hold:** reject; provider lacks business scope/future materialization.
- **Keep template Submission forever:** reject; creates unrequested indefinite retention.
- **Delete Revision row:** reject; breaks ordinal non-reuse/stable references.
- **Keep dictionary snapshot forever:** reject; defeats disposition.
- **Artifact ref_count:** reject; can drift and becomes second authority.
- **Cross-root Artifact sharing:** reject; creates disposition/late-reference race with no real V1 consumer.
- **Hold after Fence best-effort:** reject; cannot claim preservation after irreversible barrier.
- **Recursive retention of control rows:** reject; creates non-terminating records platform.

---

# 26. Essential vs Accidental Complexity

## Essential

- Dossier context + typed scope;
- M:N document context without grants;
- mutable Evidence DRAFT + immutable CAPTURE;
- exact Evidence Artifact;
- explicit type retention rules;
- immutable per-subject Binding;
- lifecycle-derived anchor;
- monotonic extension;
- LegalHold + exact materialized set/future materialization;
- explicit disposition fence + verified completion;
- one semantic Artifact retention root;
- minimal post-disposition identity skeleton.

## Accidental/deferred

- folder ownership/hierarchy;
- Dossier object-platform features;
- generic Record declaration;
- generic retention expression/jurisdiction engine;
- polymorphic owner/subject registry;
- provider lock as business authority;
- auto deletion;
- eDiscovery/custodian graph;
- recursive records retention;
- ref-count authority;
- multi-file package without consumer;
- disposition BPM;
- Dossier-local sequence reset without requirement.

---

# 27. Candidate Decision

Recommended B5 Global Maximum:

```text
Documentary Context
  DossierType + explicit eligible types
  Dossier + typed scope
  contextual Document/Evidence relationships
  stable ExternalReference mapping

Evidence
  EvidenceType + allowed formats + bounded naming
  mutable EvidenceDraft
  immutable EvidenceCapture / Void evidence
  one primary Dossier + optional secondary context

Records Governance
  explicit type RetentionRules
  immutable RetentionBinding snapshot
  append-only RetentionExtension
  LegalHold + immutable materialized Bindings
  explicit DispositionFence
  verified immutable DispositionRecord

Artifact
  exact bytes only
  one semantic retention root
  typed reachability only
  no own retention policy / no owner registry
```

With bounded B3 refinements:

```text
1. DocumentOrigin strong source-Submission FK
   → source Revision skeleton FK + exact immutable Submission/digest/hash snapshot

2. DocumentRevision terminal anchor facts
   → canonical cancelled_at / obsoleted_at

3. dictionary_snapshot on permanent Revision row
   → separate immutable RevisionDictionarySnapshot retention payload
```

No B4 semantic reopen required.

---

# 28. Reopen Triggers

Reopen only on real evidence such as:

- Dossier becomes an independent business master requiring fields/workflow;
- real Dossier hierarchy/graph;
- Evidence needs REV/Approval lifecycle;
- true indivisible multi-file Evidence;
- a third real retention-subject family makes two-member closed union inadequate;
- jurisdiction/conditional retention requires rule engine;
- retention shortening/retroactive recalculation becomes required;
- Dossier-local Evidence numbering is real requirement;
- disposition needs regulated multi-stage approval/eSignature;
- LegalHold must cover ESI/custodians/unconfirmed drafts;
- Records control records need own formal retention regime;
- B6/R10-F proves post-disposition skeleton retains disallowed Target Data;
- concurrency proof shows hold/fence/materialization cannot be linearized under B1 law.

Legacy schema/provider inconvenience is not a reopen trigger.

---

# 29. Whole-R10 Posture

B5 remains **NON-AUTHORITATIVE** until operator adjudication.

If accepted:

```text
R10-B3 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
  + bounded B5-driven refinements above
R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R10-B5 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
implementation = BLOCKED
next = R10-B6 integrated candidate + B3/B4/B5/B6 coherence challenge
```

Whole-R10 Global Coherence Review + cold independent review remain before final ratification.