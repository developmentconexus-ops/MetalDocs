# R10-B5 — Documentary Context + Records Governance + Artifact Closure — Integrated Candidate

> **Status:** NON-AUTHORITATIVE CANDIDATE — OPERATOR REVIEW / R10 INTEGRATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Input working baseline:** B1/B2 promoted authority + B3/B4 **ACCEPTED FOR R10 INTEGRATION / NON-FINAL**  
> **Input HEAD:** `942765b92be9400ecb46f0510fa18552e91fbaa2`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Independent review:** deferred to Whole-R10 unless a material exception trigger appears.

This file is staging analysis. It does **not** independently ratify B5, does not promote B3/B4 to final authority, and does not silently rewrite the frozen R3–R9.5 ledger. It consumes the current R10 working target, including the operator-approved B4 bounded Approval refinement.

B5 produced three bounded B3 technical refinements through real Records-Governance counterexamples. They are called out explicitly in §4 and must be operator-adjudicated with B5; no unrelated B3 semantic is reopened.

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

External references used only as comparison evidence:

- M-Files object relationships: <https://userguide.m-files.com/user-guide/latest/eng/object_relationships.html>
- NARA records scheduling/disposition instructions: <https://www.archives.gov/records-mgmt/scheduling/instructions>
- Microsoft Purview disposition: <https://learn.microsoft.com/en-us/purview/disposition>
- AWS S3 Object Lock: <https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock.html>

Useful comparison result:

- documentary relationships can organize context without copying/owning the related document;
- retention clocks commonly anchor on real lifecycle events;
- retention expiry and disposition decision are separate;
- object-store retention/hold is physical enforcement, not business Records-Governance authority.

MetalDocs does not import a generic records-management/ECM policy platform from those systems.

---

# 2. Evidence → Known / Inferred / Unknown / Deferred

## 2.1 Known — frozen/promoted/accepted inputs

B5 must preserve:

### Documentary Context

- `Dossier` = stable documentary context for an identifiable business subject;
- Dossier is not a physical folder and not ERP/CRM/PLM/PM/EAM master data;
- `DossierType` remains small: code/name/description/status + eligible DocumentTypes/EvidenceTypes;
- no Dossier custom fields/forms/workflow/ACL/completeness engine V1;
- Dossier stable key is unique within DossierType across the deployment; title may change;
- `{DOSSIER}` resolves the stable Dossier key;
- creation provenance is separate from zero..N ExternalReferences;
- ExternalReference uses connection + entity kind + external ID; same external identity cannot map to two Dossiers;
- external source disappearance never deletes Dossier history;
- Dossier↔Document is M:N over stable Document identity, never copies content, changes lifecycle, Area or Authorization, and never grants access;
- every CAPTURED Evidence has exactly one immutable primary Dossier;
- DRAFT Evidence may correct primary Dossier subject to authorization;
- secondary Dossier links are allowed;
- Dossier scope = one `TenantScope | AreaScope`; Evidence reuses primary Dossier scope;
- Dossier type/key/scope stable V1;
- Dossier lifecycle `ACTIVE ↔ ARCHIVED`; archive is reversible navigation state and never starts retention;
- no Dossier hierarchy/graph V1.

### Evidence

- EvidenceType is deployment/company scoped with stable code/name/status, allowed formats and small canonical naming policy;
- naming tokens V1 = `{TYPE}`, `{DOSSIER}`, `{REF}`, `{SEQ}`;
- user filename is provenance only;
- Evidence lifecycle = `DRAFT → CAPTURED → VOIDED`, where VOIDED means invalid MetalDocs capture only;
- CAPTURED content/metadata immutable;
- external-world cancellation is not Evidence VOIDED;
- Evidence does not use REV/Approval/Release by default;
- exactly one primary Artifact per Evidence V1;
- no true multi-file package until real ArtifactPackage/PLM evidence appears.

### Records Governance

- there is no generic `Record` entity/declaration operation;
- DocumentRevision gets a RetentionBinding at first RevisionSubmission;
- Evidence gets a RetentionBinding at CAPTURE;
- never-submitted Drafts/staging/recovery snapshots are not retention subjects;
- DocumentRevision retention unit includes governed immutable history: Submissions, Approval evidence, Renditions, Release/PeriodicReview evidence and referenced Artifacts;
- DocumentType/EvidenceType choose explicit `NoMinimum | KeepFor(value,DAYS|MONTHS|YEARS) | Indefinite`;
- no NULL-as-retention-policy and no hardcoded legal periods;
- Document retention clock does not run while Revision is EFFECTIVE;
- native Document anchor = superseded / obsoleted / cancelled after a submitted-but-never-released revision;
- EvidenceType chooses `CAPTURED_AT | OCCURRED_AT` anchor;
- Dossier archive never starts retention;
- RetentionBinding snapshots policy; later type changes do not silently recalculate old subjects;
- RetentionExtension can only lengthen retention V1;
- retention expiry only creates disposition eligibility;
- no automatic delete;
- current EFFECTIVE Revision is never disposition eligible;
- disposition requires explicit authorization/review, no active hold and verified physical removal before immutable DispositionRecord completion;
- LegalHold is independent of retention;
- LegalHold scopes V1 = Evidence | stable Document | Dossier;
- Document/Dossier holds materialize current RetentionBindings and continue materializing newly entering governed retention subjects while hold remains active and scope remains live;
- unlink/lifecycle changes never release already-materialized held subjects;
- holds block disposal, not normal business lifecycle;
- Dossier hold is documentary-context scope, not generic custodian/eDiscovery graph;
- Artifact has no independent retention policy; preservation derives from retained/held subjects referencing it;
- provider WORM/ObjectLock/Purview is physical enforcement only.

### B1/B2/B3/B4

- one PostgreSQL DB / schema `metaldocs`, UUID IDs, typed FKs, cross-owner RESTRICT/NO ACTION, READ COMMITTED;
- no universal tenant/company partition columns;
- canonical Authorization remains live and scope-aware;
- Dossier/Evidence relationship links never grant access;
- Artifact = immutable exact bytes; provider location is not identity;
- no confirmed semantic orphan Artifact;
- B3 Submission is immutable exact candidate;
- B4 Approval/Rendition/Release bind same Submission;
- B4 winning Release is effectivity authority;
- B4 Distribution obligations/acknowledgements are immutable revision-governance evidence.

## 2.2 Inferred corrected choices

1. Documentary Context and Records Governance remain separate owners; Dossier never becomes retention authority.
2. Dossier scope uses the same typed `TenantScope | AreaScope` shape already proven by B2, with no magic scope sentinel.
3. DossierType eligibility is explicit current configuration; absence of an eligibility row means that type is not eligible for new relation/capture under that DossierType.
4. Existing links/evidence remain valid if eligibility/type status later changes; current config governs future use only.
5. Evidence separates mutable DRAFT working state from immutable CAPTURE state so capture immutability and later disposition are structural rather than column convention.
6. Evidence canonical name is frozen at CAPTURE. DRAFT primary-Dossier correction therefore cannot silently rename already-captured evidence.
7. If naming policy uses `{SEQ}`, V1 uses one monotonic series per EvidenceType. Committed values never decrement/reuse; Dossier-local sequence reset is a reopen trigger, not launch machinery.
8. Native CAPTURE under `OCCURRED_AT` requires a known `occurred_at`; Historical Migration may preserve unknown explicitly in R10-F and such a subject never becomes silently deletion-eligible.
9. DocumentTypeRetentionRule and EvidenceTypeRetentionRule are Records-Governance-owned current configuration and must be explicit before the corresponding type can produce new retention subjects.
10. RetentionBinding is a closed typed union: exact DocumentRevision XOR exact Evidence, never generic `subject_type/id`.
11. RetentionBinding contains immutable policy snapshot only; the retention clock anchor is derived from canonical CI/Evidence lifecycle facts, avoiding a duplicated timestamp authority.
12. RetentionExtension V1 is an append-only absolute `extended_until` floor, allowed only after a known retention anchor exists and only when it lengthens the current minimum.
13. `LegalHoldSubject(hold,binding)` is immutable materialization evidence; releasing/unlinking never deletes old materialization rows.
14. Direct Evidence hold requires an existing Evidence RetentionBinding. Document/Dossier holds are the two future-materializing scopes.
15. Hold creation/materialization and disposition fence serialize on the same RetentionBinding so an active hold and irreversible disposition cannot cross silently.
16. Disposition uses an immutable `DispositionFence` as the authorized irreversible barrier and a separate immutable `DispositionRecord` only after physical/removal verification. Worker retry/lease/error state remains R10-D mechanism state.
17. Business lifecycle state and Records disposition are orthogonal: a disposed DocumentRevision remains SUPERSEDED/OBSOLETE/CANCELLED; disposed Evidence retains its last business state CAPTURED/VOIDED while a DispositionRecord says its retained payload was lawfully disposed.
18. Minimal DocumentRevision/Evidence identity skeletons may survive disposition; governed retention payload/history is what disposition removes.
19. RetentionBinding / LegalHold / RetentionExtension / Disposition skeletons are not recursively declared as new retention subjects V1; creating “retention of retention of retention” would be an unsupported records platform.
20. DocumentRevision retention unit additionally includes B4 `SubmissionFeedback`, `SubmissionApprovalRequirement`, `ReleasePlan`, DistributionObligation and AcknowledgementRecord because those facts have no independent governed meaning after their Revision history is disposed.
21. Artifact physical deletion is based on surviving typed semantic reachability, never a mutable ref-count authority.
22. A confirmed Artifact that loses its final typed semantic reference in a semantic transaction must be removed from Artifact semantic state in that same transaction; physical byte cleanup/deletion is R10-C/R10-D mechanism work subject to disposition rules.

## 2.3 Unknown — intentionally not converted to defaults

- whether a future product needs Dossier-local Evidence sequence reset rather than per-EvidenceType sequence;
- whether a future product needs direct whole-company/Area/User-specific external Dossier identity beyond the accepted ExternalReference shape;
- whether Evidence secondary Dossier links need stricter post-CAPTURE mutation rules than current contextual add/remove semantics;
- whether a future regulatory profile requires a separate approval workflow for disposition rather than the single explicit authorized decision/fence V1;
- exact provider-neutral physical deletion verification manifest shape owned with R10-C;
- final privacy classification of the tiny post-disposition Evidence/Revision/Records-Governance skeleton fields, to be challenged again by B6/R10-F.

These do not block B5 because current invariants can be enforced without implementing the future capability.

## 2.4 Deferred

```text
Audit field-level privacy / Interchange connection target / final cross-owner matrix → B6
physical object deletion / object-lock/scanner/provider verification               → R10-C
async disposition worker / retries / leases / notifications / projections         → R10-D
API/frontend Dossier/Evidence/hold/disposition journeys                           → R10-E
historical migration / cutover / legacy deletion                                  → R10-F
```

---

# 3. Root cause

B5 is not primarily a “records tables” problem. The structural failure class is **ownership collapse**:

```text
folder/context
+ governed captured evidence
+ retention policy
+ legal preservation
+ physical storage deletion
```

can easily collapse into one generic repository/record engine.

That creates predictable defects:

- Dossier accidentally owns/grants access to linked Documents;
- Evidence is forced through Document revision/approval semantics it does not need;
- retention policy becomes Artifact/storage metadata instead of business Records Governance;
- LegalHold becomes a provider object-lock flag and loses Document/Dossier scope semantics;
- disposition becomes automatic expiry deletion;
- generic `owner_type/id` or `subject_type/id` registries become second authorities;
- historical provenance keeps whole retention units alive forever through accidental FKs;
- deleting a retained Revision row can allow forbidden REV ordinal reuse.

Target structure must separate those concerns while preserving exact reachability and disposal correctness.

---

# 4. B3 bounded refinements revealed by B5

These are real cross-stage counterexamples, not implementation convenience.

## 4.1 DocumentOrigin must survive lawful source-template disposition without retaining the entire source unit forever

Accepted B3 candidate used:

```text
DocumentOrigin
  derived_document_id
  source_template_submission_id FK RevisionSubmission
```

Counterexample:

```text
Template REV004 / Submission S4
→ creates derived Document D
→ years later Template REV004 retention expires
→ disposition wants to delete S4 + source Artifact
→ D.DocumentOrigin FK makes S4 undeletable forever
```

That silently invents indefinite retention of every template source ever used.

Corrected candidate:

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

At derived creation B3/B4 still prove the source Submission is the current effective template source. The immutable origin copies exact source provenance. Later source retention disposal cannot rewrite the derived Document and does not require keeping the source Submission/Artifact physically present.

`source_template_revision_id` remains a typed FK because the minimal DocumentRevision identity skeleton survives disposition. Submission UUID/digest/hash are historical snapshots, not floating business references.

## 4.2 DocumentRevision needs canonical terminal timestamps for retention anchors

Accepted B3 candidate persists Revision state but no domain-owned `cancelled_at` / `obsoleted_at`.

Audit cannot become the retention anchor authority.

Refinement:

```text
DocumentRevision
  ...
  cancelled_at TIMESTAMPTZ NULL
  obsoleted_at TIMESTAMPTZ NULL
```

State-coupled law:

```text
CANCELLED → cancelled_at NOT NULL, obsoleted_at NULL
OBSOLETE  → obsoleted_at NOT NULL, cancelled_at NULL
other native states → both NULL
```

Those timestamps are written exactly once in the same CI lifecycle transaction as the state transition.

SUPERSEDED anchor is **not duplicated**: native supersession time is the `ReleaseRecord.released_at` whose `prior_effective_revision_id` is that Revision.

Historical Migration may preserve imported/unknown source anchors under R10-F rules rather than fabricating native timestamps.

## 4.3 DocumentRevision identity skeleton must not retain `dictionary_snapshot` forever after disposition

Accepted B3 candidate stores immutable `dictionary_snapshot` on the same DocumentRevision row whose ordinal must survive forever to prevent REV reuse.

Counterexample:

```text
REV002 retention unit lawfully disposed
→ delete DocumentRevision row: forbidden, ordinal may be reused
→ keep row: immutable dictionary_snapshot survives forever
→ clear snapshot: rewrites historical immutable state
```

Refinement:

```text
DocumentRevision
  id / document_id / revision_no / state / creator / timestamps
  // minimal identity/lifecycle skeleton

RevisionDictionarySnapshot
  revision_id UUID PRIMARY KEY FK DocumentRevision(id) RESTRICT
  snapshot JSONB NOT NULL CHECK jsonb_typeof(...)='object'
```

`RevisionDictionarySnapshot` remains immutable while present, participates in B3 Submission manifest construction, and is part of the DocumentRevision retention unit. It may be removed only by completed Records disposition. The Revision identity row remains.

No other B3 core semantic is reopened.

---

# 5. Target invariant

> **Dossier supplies documentary context without taking ownership or granting access. Evidence supplies a small capture lifecycle with one exact primary Artifact and immutable captured state. Records Governance automatically binds submitted DocumentRevisions and captured Evidence to an immutable retention-policy snapshot, materializes active LegalHolds over the exact RetentionBindings in scope, and permits irreversible disposition only through a serialized explicit fence with no active hold and verified removal. Artifact persistence remains exact-byte state whose survival is determined only by surviving typed semantic references.**

Corollaries:

```text
Dossier                 != folder/storage container
Dossier link            != access grant
Dossier                 != retention subject by itself
Evidence                != DocumentRevision
CAPTURE                  != Approval/Release
RetentionBinding        != generic Record
Retention policy        != Artifact policy
LegalHold               != provider ObjectLock
expiry                  != deletion
DispositionFence        != worker/job state
DispositionRecord       != AuditEvent
Artifact reachability   != ref_count column
post-disposition state  != business lifecycle rewrite
```

---

# 6. Credible alternatives

## A — Dossier as folder/binder + generic Record aggregate

```text
Dossier owns files/documents
Record(owner_type,owner_id)
RetentionPolicy rules engine
```

**Reject — Local Maximum / overgeneralized.** It makes context into ownership and recreates a generic ECM/records platform.

## B — generic polymorphic retention/artifact registries

```text
ArtifactOwner(owner_type,owner_id)
RetentionSubject(subject_type,subject_id)
```

**Reject.** Smaller table count, larger authority ambiguity. It bypasses B1 typed-reference law and creates an extensibility platform with no real consumer.

## C — typed Documentary Context + specialized Evidence + closed-union RetentionBinding + materialized holds + explicit disposition fence

**Recommended Global Maximum.** It preserves only the real distinct lifecycle/authority classes and leaves provider deletion, jobs and future record classes outside the kernel.

---

# 7. Documentary Context relational candidate

## 7.1 DossierType

```text
DossierType
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  name TEXT NOT NULL
  description TEXT NULL
  status TEXT NOT NULL CHECK ACTIVE|INACTIVE
```

Laws:

- code immutable;
- display/status mutable;
- INACTIVE blocks new Dossier creation and new eligibility-dependent relations, never invalidates existing context;
- no custom field/form/workflow/ACL/completeness metadata.

## 7.2 DossierType eligibility

```text
DossierTypeDocumentType
  dossier_type_id UUID NOT NULL FK DossierType(id) RESTRICT
  document_type_id UUID NOT NULL FK DocumentType(id) RESTRICT
  PRIMARY KEY(dossier_type_id,document_type_id)

DossierTypeEvidenceType
  dossier_type_id UUID NOT NULL FK DossierType(id) RESTRICT
  evidence_type_id UUID NOT NULL FK EvidenceType(id) RESTRICT
  PRIMARY KEY(dossier_type_id,evidence_type_id)
```

Current configuration only. Removing eligibility never rewrites existing Dossier relations/captured evidence.

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

Scope XOR:

```text
exactly one of tenant_scope_id / area_scope_id is non-NULL
```

Mutation law:

- id/type/key/scope/created facts immutable;
- title mutable;
- archive/re-enable changes `archived_at`; Audit later owns transition timeline;
- archive never deletes/unlinks and never starts retention.

No Dossier hierarchy or physical repository path.

## 7.4 DossierDocumentLink

```text
DossierDocumentLink
  dossier_id UUID NOT NULL FK Dossier(id) RESTRICT
  document_id UUID NOT NULL FK Document(id) RESTRICT
  linked_by_user_id UUID NOT NULL FK User(id) RESTRICT
  linked_at TIMESTAMPTZ NOT NULL

  PRIMARY KEY(dossier_id,document_id)
```

Current contextual relationship.

Add requires:

- Dossier ACTIVE;
- DocumentType eligible for DossierType;
- actor authorized for Dossier management and separately authorized for the Document relationship needed to create the link.

Unlink removes current context only. It never removes already-materialized LegalHoldSubject rows.

Cross-Area/cross-scope links are not forbidden merely because scopes differ; every target remains independently authorized. No transitive grant.

## 7.5 DossierExternalReference — relationship owned by Documentary Context, connection identity supplied by B6 Interchange

Conceptual target:

```text
DossierExternalReference
  id UUID PRIMARY KEY
  dossier_id UUID NOT NULL FK Dossier(id) RESTRICT
  connection_id UUID NOT NULL FK InterchangeConnection(id) RESTRICT  // B6 target
  entity_kind TEXT NOT NULL
  external_id TEXT NOT NULL
  created_at TIMESTAMPTZ NOT NULL

  UNIQUE(connection_id,entity_kind,external_id)
```

Mapping fields are immutable in normal serving paths. External system disappearance/status change never deletes/rekeys the Dossier.

B6 may name the concrete Interchange connection family but may not weaken the one-external-identity→one-Dossier law without a reopen.

---

# 8. Evidence relational candidate

## 8.1 EvidenceType

```text
EvidenceType
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  name TEXT NOT NULL
  description TEXT NULL
  status TEXT NOT NULL CHECK ACTIVE|INACTIVE
  naming_pattern TEXT NOT NULL
```

Code immutable. Current naming config affects future CAPTURE only.

Closed naming parser admits only literals + `{TYPE}` / `{DOSSIER}` / `{REF}` / `{SEQ}`. It is not a formula/custom-field engine.

## 8.2 Allowed formats

```text
EvidenceTypeAllowedFormat
  evidence_type_id UUID NOT NULL FK EvidenceType(id) RESTRICT
  content_format TEXT NOT NULL CHECK closed ContentFormat vocabulary
  PRIMARY KEY(evidence_type_id,content_format)
```

CAPTURE proves exact Artifact format is currently allowed. Existing CAPTURED Evidence is never invalidated by later config change.

## 8.3 EvidenceSequence — minimum `{SEQ}` mechanism

```text
EvidenceSequence
  evidence_type_id UUID PRIMARY KEY FK EvidenceType(id) RESTRICT
  next_value BIGINT NOT NULL CHECK next_value >= 1
```

Used only if naming pattern contains `{SEQ}`.

Candidate V1 scope = one monotonic series per EvidenceType across the deployment. This is smaller and collision-safe whether or not `{DOSSIER}` is present in the pattern.

If a real requirement later needs Dossier-local restart, replace this bounded mechanism deliberately; do not smuggle Dossier into identity now.

## 8.4 Evidence — stable identity/lifecycle skeleton

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

Laws:

- DRAFT may change `primary_dossier_id` through explicit Evidence edit;
- CAPTURE requires primary Dossier and sets/finalizes canonical name/sequence;
- after CAPTURE, type/primary Dossier/canonical name/sequence are immutable;
- post-disposition Evidence row is the minimal business-identity/lifecycle skeleton; DispositionRecord is the independent records-lifecycle fact.

`UNIQUE(canonical_name)` applies where canonical_name is non-NULL.

## 8.5 EvidenceDraft — mutable pre-capture state

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

Mutable only while Evidence = DRAFT.

Artifact replacement is whole primary-content replacement for Evidence DRAFT. Old Artifact semantic rows may survive only if another typed reference still exists; otherwise semantic Artifact deletion occurs in the same transaction and provider-byte cleanup is later mechanism work.

No separate Evidence revision/autosave lifecycle.

## 8.6 EvidenceCapture — immutable retained payload

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

Immutable while present.

Native CAPTURE transaction:

1. locks Evidence + current Draft;
2. proves Dossier/EvidenceType eligibility and scope;
3. proves allowed format;
4. if EvidenceType retention anchor is OCCURRED_AT, requires `occurred_at`;
5. allocates `{SEQ}` if required and computes canonical name;
6. establishes exact confirmed Artifact ownership;
7. inserts immutable EvidenceCapture;
8. sets Evidence primary Dossier/name/sequence/state CAPTURED;
9. deletes EvidenceDraft;
10. inserts RetentionBinding;
11. materializes active Dossier holds in scope;
12. commits all semantic facts together.

No fake Approval/Release/Revision is introduced.

## 8.7 EvidenceVoidRecord

```text
EvidenceVoidRecord
  evidence_id UUID PRIMARY KEY FK Evidence(id) RESTRICT
  voided_by_user_id UUID NOT NULL FK User(id) RESTRICT
  reason TEXT NOT NULL
  voided_at TIMESTAMPTZ NOT NULL
```

Immutable. VOID transition never mutates EvidenceCapture payload.

EvidenceCapture + EvidenceVoidRecord are part of Evidence retention payload and may be removed only by completed Records disposition; Evidence identity skeleton remains.

## 8.8 Secondary Dossier links

```text
EvidenceSecondaryDossierLink
  evidence_id UUID NOT NULL FK Evidence(id) RESTRICT
  dossier_id UUID NOT NULL FK Dossier(id) RESTRICT
  linked_by_user_id UUID NOT NULL FK User(id) RESTRICT
  linked_at TIMESTAMPTZ NOT NULL

  PRIMARY KEY(evidence_id,dossier_id)
```

Secondary link may not duplicate primary Dossier.

Current contextual relation; add/remove does not copy Evidence or change its primary scope. Add requires DossierType eligibility and canonical authorization on both sides.

An active Dossier hold materializes the Evidence RetentionBinding when a captured Evidence enters that Dossier through a secondary link. Later unlink never removes the materialized hold subject.

---

# 9. Retention policy configuration

## 9.1 DocumentTypeRetentionRule

```text
DocumentTypeRetentionRule
  document_type_id UUID PRIMARY KEY FK DocumentType(id) RESTRICT
  mode TEXT NOT NULL CHECK NO_MINIMUM|KEEP_FOR|INDEFINITE
  value INTEGER NULL CHECK value IS NULL OR value > 0
  unit TEXT NULL CHECK unit IS NULL OR unit IN (DAYS,MONTHS,YEARS)
```

CHECK:

```text
NO_MINIMUM | INDEFINITE → value/unit NULL
KEEP_FOR                 → value/unit NOT NULL
```

## 9.2 EvidenceTypeRetentionRule

```text
EvidenceTypeRetentionRule
  evidence_type_id UUID PRIMARY KEY FK EvidenceType(id) RESTRICT
  mode TEXT NOT NULL CHECK NO_MINIMUM|KEEP_FOR|INDEFINITE
  value INTEGER NULL CHECK value IS NULL OR value > 0
  unit TEXT NULL CHECK unit IS NULL OR unit IN (DAYS,MONTHS,YEARS)
  anchor_kind TEXT NOT NULL CHECK CAPTURED_AT|OCCURRED_AT
```

No DossierType retention rule exists.

Every active type capable of producing a new retention subject must have exactly one explicit rule. Missing rule fails closed before SUBMIT/CAPTURE; NULL never means NoMinimum.

Rule changes affect future RetentionBindings only.

---

# 10. RetentionBinding — one closed retention-subject family

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

Subject XOR:

```text
exactly one of document_revision_id / evidence_id is non-NULL
```

Structural uniqueness:

```text
UNIQUE(document_revision_id) WHERE document_revision_id IS NOT NULL
UNIQUE(evidence_id) WHERE evidence_id IS NOT NULL
```

Policy-shape CHECK mirrors the explicit type rule shape.

Mutation law: immutable/append-only while present.

Binding timing:

- first RevisionSubmission transaction creates DocumentRevision binding if absent;
- later same-REV resubmissions reuse that binding and never resnapshot policy;
- Evidence CAPTURE creates Evidence binding;
- never-submitted Revision and Evidence DRAFT have no binding.

Retention clock anchor is derived from canonical facts:

### DocumentRevision

```text
SUPERSEDED → ReleaseRecord.released_at where prior_effective_revision_id = Revision
OBSOLETE   → DocumentRevision.obsoleted_at
CANCELLED  → DocumentRevision.cancelled_at, only for a binding that already exists
DRAFT/SUBMITTED/EFFECTIVE → no running clock / not disposition eligible
```

### Evidence

```text
CAPTURED_AT → EvidenceCapture.captured_at
OCCURRED_AT → EvidenceCapture.occurred_at
```

No mutable `retain_until` authority is persisted on Binding. Eligibility is deterministic derivation from immutable policy snapshot + canonical anchor + monotonic extensions.

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

Append-only.

V1 creation requires:

- known retention anchor;
- policy not INDEFINITE;
- `extended_until` strictly later than current derived minimum retention end, including prior extensions.

Effective minimum end:

```text
max(base policy due, all RetentionExtension.extended_until)
```

No retroactive shortening. If future product needs pre-anchor policy extension, reopen deliberately; active LegalHold already provides the preservation seam before a clock begins.

Whole-company tenant-owner-only `retention.extend` remains Tenant-wide regardless of subject Area.

---

# 12. LegalHold

## 12.1 LegalHold

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

Scope XOR: exactly one Evidence / Document / Dossier.

Scope and creation facts immutable. Release is terminal for that Hold row. A future preservation need creates a new Hold rather than reactivating an old one.

`legal_hold.manage` remains Tenant-wide tenant-owner-only.

## 12.2 LegalHoldSubject — immutable materialized preservation set

```text
LegalHoldSubject
  legal_hold_id UUID NOT NULL FK LegalHold(id) RESTRICT
  retention_binding_id UUID NOT NULL FK RetentionBinding(id) RESTRICT
  materialized_at TIMESTAMPTZ NOT NULL

  PRIMARY KEY(legal_hold_id,retention_binding_id)
```

Rows are never deleted because a link disappears or hold is released. Active blocking condition is:

```text
LegalHold.released_at IS NULL
AND LegalHoldSubject row exists for Binding
```

### Evidence hold

Direct Evidence hold requires existing RetentionBinding and materializes that one Binding immediately. It does not create a speculative hold over Evidence DRAFT/staging.

### Document hold

Creation materializes every current RetentionBinding for all Revisions of stable Document. First future Submission on a new Revision materializes the new Binding while Hold remains active.

### Dossier hold

Creation materializes current RetentionBindings for:

- every Revision of each currently linked Document;
- every CAPTURED Evidence whose primary Dossier is this Dossier;
- every CAPTURED Evidence with this Dossier as active secondary context.

Future entering subjects materialize while hold remains active:

- first Submission for a linked Document;
- Evidence CAPTURE in that Dossier;
- later Document/Evidence contextual link into that Dossier.

Unlink/archive/lifecycle changes never remove already-materialized subjects.

A future subject created after stable Document/Evidence genuinely left Dossier live scope is not captured unless another active direct Document/Evidence hold applies.

---

# 13. Hold-materialization concurrency contracts

B1 remains READ COMMITTED; no generic SERIALIZABLE or distributed lock service.

The target must make “hold creation vs subject entering scope” linearizable.

Canonical roots:

```text
Document mutation / first Submission / Document hold → Document row
Evidence capture / Evidence hold                    → Evidence row
Dossier links / Dossier hold                        → Dossier row
```

When an operation spans Document/Evidence + Dossier, acquire subject root first then Dossier rows in ascending UUID order.

Examples:

### First Document Submission

```text
lock Document
create first RetentionBinding
materialize active direct Document holds
load linked Dossiers
lock linked Dossiers ascending
materialize active Dossier holds
```

### DossierDocumentLink add

```text
lock Document
lock Dossier
insert link
materialize active Dossier holds over all existing Document Revision Bindings
```

This prevents link-vs-first-submission missed-subject races.

### Evidence CAPTURE

```text
lock Evidence
lock primary + current secondary Dossiers ascending
create RetentionBinding
materialize active Dossier holds
```

### Secondary Evidence link add

```text
lock Evidence
lock target Dossier
insert link
if Evidence already has RetentionBinding, materialize active Dossier holds
```

### Dossier hold create

```text
lock Dossier
insert Hold
load current in-scope RetentionBindings deterministically
lock Bindings ascending
prove no active DispositionFence for each
insert LegalHoldSubject rows
```

Hold creation never acquires Document/Evidence roots after RetentionBinding locks, avoiding a reverse lock cycle.

Exact SQL lock modes/order remain an implementation-spec proof obligation; if the composed B1–B5 lock graph cannot be proven acyclic, B5 reopens before code.

---

# 14. Disposition eligibility

## 14.1 Derived eligibility

DocumentRevision binding can be eligible only when:

```text
state ∈ SUPERSEDED|OBSOLETE|CANCELLED
AND retention anchor known
AND policy != INDEFINITE
AND now >= max(base due, extensions)
AND no active LegalHoldSubject
AND no existing DispositionFence
```

Current EFFECTIVE, DRAFT or SUBMITTED is never eligible.

Evidence binding can be eligible only when:

```text
state ∈ CAPTURED|VOIDED
AND anchor known
AND policy != INDEFINITE
AND now >= max(base due, extensions)
AND no active LegalHoldSubject
AND no existing DispositionFence
```

`NO_MINIMUM` means base due = anchor, not “delete automatically”.

## 14.2 DispositionFence — explicit decision + irreversible serialization barrier

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

Immutable.

Fence creation transaction:

```text
lock subject root
lock RetentionBinding FOR UPDATE
recompute eligibility from canonical state
prove no active hold
insert DispositionFence
commit
```

After Fence exists:

- new LegalHold materialization for this Binding fails visibly rather than claiming protection it cannot guarantee;
- R10-D may retry physical removal but cannot clear/change business authorization;
- if provider deletion partially fails, Fence remains and blocks contradictory future hold/disposition operations until recovery completes.

This is not a generic disposition workflow state machine.

## 14.3 DispositionRecord — completion authority

```text
DispositionRecord
  id UUID PRIMARY KEY
  disposition_fence_id UUID NOT NULL UNIQUE FK DispositionFence(id) RESTRICT
  completed_at TIMESTAMPTZ NOT NULL
  verification_schema TEXT NOT NULL
  verification_manifest JSONB NOT NULL CHECK jsonb_typeof(...)='object'
```

Immutable.

Record may be inserted only after R10-C proves every Artifact/content element in the unit is either:

```text
DELETED_AND_VERIFIED
OR
PRESERVED_SHARED_WITH_SURVIVING_TYPED_REFERENCE
```

Provider bucket/key/version is not business identity. The bounded verification manifest identifies exact Artifact IDs/hashes and disposition outcome without turning provider coordinates into Records authority.

No DispositionRecord is written merely because a job was queued, a DELETE call returned once, or retention expired.

---

# 15. DocumentRevision retention unit — integrated B3+B4 closure

For one retained DocumentRevision, the governed unit includes at minimum:

### B3 / Controlled Information

- `RevisionDictionarySnapshot`;
- all `RevisionSubmission` rows and exact Submission manifests/digests;
- Submission source Artifact references;
- applicable immutable TemplateSpec/submitted structured provenance that is not already solely inside manifest;
- immutable PeriodicReviewRecord rows;
- other immutable revision-level governed evidence proven by B3.

Excluded as operational/non-authoritative:

- mutable WorkingContent after terminal cleanup;
- EditorSession;
- never-authoritative staging/WorkingSnapshots.

### B4 / Approval + Release + Distribution

- `SubmissionApprovalRequirement`;
- `ApprovalInstance` / StepInstances / Participants / StepDecisions / reassignment evidence;
- `SubmissionFeedback`;
- `ReleasePlan`;
- `Rendition` rows + their output Artifacts;
- `ReleaseRecord` for this candidate Revision;
- `DistributionObligation` rows created by this Release;
- `AcknowledgementRecord` rows;
- exact fresh-auth bounded decision evidence stored by Approval.

Configuration shared across multiple subjects — e.g. ApprovalPolicy/PolicyVersion, DocumentType, User identity — is **not** automatically swallowed into this retention unit. References may remain to surviving configuration/identity skeletons.

### Why Distribution evidence belongs

A DistributionObligation/Acknowledgement has no independent subject once the Revision/Release it describes is gone. Retaining it separately would create orphan compliance evidence; deleting it earlier would destroy the Revision's governed history. It therefore follows the Revision unit.

---

# 16. Evidence retention unit

Evidence unit includes:

- immutable EvidenceCapture;
- exact primary Artifact;
- EvidenceVoidRecord when present;
- capture provenance/original filename held inside EvidenceCapture;
- any Evidence-owned immutable metadata required to interpret the captured subject.

Stable Evidence skeleton remains after disposition:

```text
Evidence.id
Evidence.evidence_type_id
last business lifecycle state
creation identity/timestamp
minimal naming/sequence/context fields required by product integrity
```

The exact surviving field list gets a B6/R10-F privacy-minimization challenge before final R10 ratification. No profile/email/display-name snapshot is added to Evidence.

Records-Governance DispositionRecord, RetentionBinding, extensions and released Hold history are not recursively part of the disposed Evidence unit.

---

# 17. Artifact global typed-reference closure

## 17.1 Known semantic Artifact references after B5

Closed V1 reference surface includes at least:

```text
WorkingContent.primary_artifact_id          → B3 mutable DRAFT authority
RevisionSubmission.source_artifact_id       → B3 immutable candidate
Rendition.output_artifact_id                → B4 immutable derived representation
EvidenceDraft.primary_artifact_id           → B5 mutable pre-capture state
EvidenceCapture.primary_artifact_id         → B5 immutable captured evidence
```

Future B6/C/D mechanism/provenance tables do not become semantic owners by convenience.

## 17.2 No generic owner registry

Do not create:

```text
ArtifactOwner(owner_type,owner_id)
Artifact.ref_count
Artifact.retention_until
```

Business relationships remain typed FKs from owning semantic objects.

## 17.3 Deferred global reachability guard

Because “confirmed Artifact has at least one typed semantic reference” is a global invariant across several tables, implementation must use one mechanically verified closed reference catalog plus database-level deferred enforcement or an equivalent mechanism that covers every serving path.

The catalog is **enforcement metadata**, not business authority, and must have parity tests against the actual typed FK surface.

At transaction end:

```text
confirmed Artifact with zero admitted typed semantic refs → constraint failure
```

Semantic replacement/disposition that removes the final ref must delete the Artifact semantic row in the same local transaction. Provider bytes may remain temporarily only as R10-C/D cleanup mechanism state.

## 17.4 Disposition and shared Artifact

Before provider deletion, compute surviving typed references excluding relations scheduled for this retention unit.

```text
surviving ref exists
→ keep Artifact row + bytes
→ verification manifest marks PRESERVED_SHARED_WITH_SURVIVING_TYPED_REFERENCE

no surviving ref
→ R10-C deletes/verifies bytes
→ final DB tx removes last semantic refs + Artifact row
```

No mutable count can substitute for actual typed reachability.

---

# 18. Final disposition transaction shape

External byte removal cannot join a PostgreSQL transaction.

Correct choreography:

```text
1. B5 Fence transaction
   authorize + recompute eligibility + prove no hold + insert DispositionFence

2. R10-C/D mechanism phase
   enumerate exact unit Artifacts
   preserve shared references
   delete exclusive physical bytes
   verify provider absence/integrity outcome
   retry/reconcile until terminal physical proof exists

3. B5 finalization transaction
   lock subject + Binding/Fence
   prove same fence, no contradictory semantic mutation
   delete retention-unit subordinate semantic rows in deterministic order
   delete Artifact rows whose last typed refs are being removed and physical deletion is verified
   keep shared Artifacts
   insert immutable DispositionRecord
   commit
```

If step 2 succeeds physically but step 3 fails, Fence remains; retry finalization. The system never fabricates a DispositionRecord before local semantic cleanup is committed.

Cross-owner DELETE remains explicit composition, never FK CASCADE across owners.

---

# 19. Post-disposition skeleton laws

## 19.1 DocumentRevision

DocumentRevision row survives because:

- `revision_no` must never reuse;
- later ReleaseRecords may reference prior Revision identity;
- derived DocumentOrigin may reference source template Revision identity;
- DispositionRecord/RetentionBinding needs a stable subject identity.

After completed disposition the revision retains only its minimal CI identity/lifecycle skeleton plus Records-Governance records. Retention payload rows are absent.

Do not add `DISPOSED` to DocumentRevision state. Records disposition is a separate axis.

## 19.2 Evidence

Evidence stable row survives as minimal identity/lifecycle skeleton so Records Governance can prove what was disposed without adding `DISPOSED` to the frozen Evidence lifecycle.

`EvidenceCapture` and its governed payload are absent after completed disposition.

## 19.3 No recursive records engine

Do not create RetentionBindings for:

```text
RetentionBinding
LegalHold
LegalHoldSubject
RetentionExtension
DispositionFence
DispositionRecord
```

Their minimal Records-Governance evidence is a separate system-governance skeleton. A future mandated retention regime for those control records is a new real consumer/reopen decision, not recursive default behavior.

---

# 20. Authorization / B2 coherence

### Dossier

```text
dossier.read / dossier.create / dossier.manage
```

Target scope = Dossier's typed TenantScope/AreaScope plus relationship/state predicates.

Dossier link never grants Document/Evidence access. Link creation/removal must separately authorize both the Dossier operation and the related target as required.

### Evidence

```text
evidence.read / evidence.create / evidence.edit / evidence.capture / evidence.void
```

Target scope = primary Dossier scope. Secondary links never widen authority.

DRAFT primary-Dossier move requires authority on both old/new relevant scopes.

### Records Governance

Frozen whole-company tenant-owner-only permissions remain Tenant-wide even when the subject belongs to an Area:

```text
retention.extend
legal_hold.manage
disposition.manage
```

Do not silently reinterpret them as Area-local because the retained subject has an Area.

No mechanism permissions:

```text
artifact.delete
storage.purge
objectlock.manage
retention.job.retry
```

Provider/storage actions happen under system mechanism authority after a canonical Records-Governance decision.

---

# 21. Persistence class × mutation law

| Family | Owner | Class / mutation law |
|---|---|---|
| DossierType | Documentary Context | semantic current config / code immutable, display/status mutable |
| DossierTypeDocumentType | Documentary Context | semantic current config / add-remove |
| DossierTypeEvidenceType | Documentary Context | semantic current config / add-remove |
| Dossier | Documentary Context | semantic identity / stable type-key-scope; title/archive mutable |
| DossierDocumentLink | Documentary Context | semantic current context / add-remove |
| DossierExternalReference | Documentary Context | semantic external identity relation / immutable normal serving path |
| EvidenceType | Documentary Context | semantic current config / code immutable |
| EvidenceTypeAllowedFormat | Documentary Context | semantic current config / add-remove |
| EvidenceSequence | Documentary Context | durable mechanism / monotonic counter |
| Evidence | Documentary Context | semantic identity + explicit lifecycle |
| EvidenceDraft | Documentary Context | semantic DRAFT working state / mutable |
| EvidenceCapture | Documentary Context | semantic captured payload / immutable while retained |
| EvidenceVoidRecord | Documentary Context | semantic evidence / immutable while retained |
| EvidenceSecondaryDossierLink | Documentary Context | semantic current context / add-remove |
| DocumentTypeRetentionRule | Records Governance | semantic current policy / mutable future-facing |
| EvidenceTypeRetentionRule | Records Governance | semantic current policy / mutable future-facing |
| RetentionBinding | Records Governance | semantic policy snapshot / immutable |
| RetentionExtension | Records Governance | semantic extension evidence / immutable append-only |
| LegalHold | Records Governance | semantic preservation scope / explicit release lifecycle |
| LegalHoldSubject | Records Governance | semantic materialized hold evidence / immutable |
| DispositionFence | Records Governance | semantic irreversible disposition authorization barrier / immutable |
| DispositionRecord | Records Governance | semantic completion evidence / immutable |
| Artifact | Artifact | exact-byte semantic identity / immutable while present |

---

# 22. Structural constraint envelope

```text
DossierType.code                                      UNIQUE
Dossier(dossier_type_id,stable_key)                   UNIQUE
Dossier scope                                         XOR Tenant/Area
DossierTypeDocumentType pair                          PK
DossierTypeEvidenceType pair                          PK
DossierDocumentLink pair                              PK
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
RetentionBinding.document_revision_id                 partial UNIQUE
RetentionBinding.evidence_id                          partial UNIQUE
RetentionBinding subject                              XOR DocumentRevision/Evidence
RetentionExtension                                   append-only; monotonic effective due
LegalHold scope                                       XOR Evidence/Document/Dossier
LegalHoldSubject(hold,binding)                         PK
DispositionFence.retention_binding_id                 UNIQUE
DispositionRecord.disposition_fence_id                UNIQUE
```

Cross-row DB/application guards must prove:

- Dossier/Evidence eligibility on new links/capture;
- Evidence primary Dossier immutable after CAPTURE;
- EvidenceCapture exists for CAPTURED/VOIDED unless completed DispositionRecord exists;
- secondary Dossier cannot equal primary;
- native OCCURRED_AT CAPTURE has occurred_at;
- first Submission/CAPTURE cannot succeed without RetentionBinding;
- binding policy snapshot shape valid;
- extension strictly lengthens known due;
- LegalHoldSubject belongs to Hold scope at materialization time;
- no new materialization can cross an active DispositionFence;
- DispositionFence cannot be created when active hold exists or subject is ineligible;
- DispositionRecord requires completed physical verification contract;
- confirmed Artifact cannot commit without an admitted typed semantic reference.

No cross-owner CASCADE/SET NULL.

---

# 23. Core transaction contracts

## 23.1 First Document Submission — B3/B4/B5 composed commit

```text
BEGIN
  B2 eligibility/Authorization as required
  lock Document / open Revision / WorkingContent under existing B3 order
  load current DocumentTypeRetentionRule
  prove explicit rule exists

  B3 freeze exact WorkingContent → immutable RevisionSubmission
  B4 insert SubmissionApprovalRequirement + ReleasePlan snapshots

  if no RetentionBinding for Revision:
    insert immutable binding from current retention rule
    materialize active direct Document holds
    lock linked Dossiers ascending
    materialize active Dossier holds

  required B6 Audit/durable intents later compose here
COMMIT
```

Same-REV resubmission creates new Submission/B4 attempt state but **does not** create or resnapshot RetentionBinding.

## 23.2 Evidence CAPTURE

See §8.6. CAPTURE + exact Artifact relation + RetentionBinding + active Dossier-hold materialization is one semantic local commit.

## 23.3 Document/Dossier LegalHold creation

```text
lock stable scope root
insert Hold
resolve current exact RetentionBindings in scope
lock Bindings ascending
prove no active DispositionFence
insert LegalHoldSubject rows
commit
```

## 23.4 RetentionExtension

```text
lock RetentionBinding
resolve canonical anchor
compute current due incl. existing extensions
prove requested extended_until is later
insert immutable extension
commit
```

## 23.5 DispositionFence

```text
lock subject root
lock RetentionBinding
recompute lifecycle/anchor/policy/extensions
prove eligible
prove zero active holds
insert immutable Fence
commit
```

## 23.6 Disposition completion

See §18. Provider effects are never inside local transaction; final DispositionRecord and semantic deletion commit together after verified external state.

---

# 24. Cross-stage coherence challenges

## C1 — first Submission returns to DRAFT

Binding already exists because records status begins at first Submission. Return-for-changes does not remove/resnapshot Binding. Retention clock still does not run until terminal anchor.

**Pass.** No B3/B4 reopen.

## C2 — template source Revision disposed while derived Document survives

Old B3 hard FK to source Submission would force indefinite source retention.

**Finding / bounded B3 refinement:** DocumentOrigin becomes self-contained source provenance snapshot + source Revision skeleton FK (§4.1).

## C3 — Revision row must survive but dictionary snapshot should not

Old B3 co-location makes ordinal preservation conflict with lawful disposal.

**Finding / bounded B3 refinement:** `RevisionDictionarySnapshot` split (§4.3).

## C4 — retention anchor for cancelled/obsolete Revision

Audit cannot be domain authority.

**Finding / bounded B3 refinement:** state-coupled CI timestamps (§4.2).

## C5 — Release supersession anchor

Do not duplicate timestamp into Binding. `ReleaseRecord.released_at` is canonical native supersession anchor.

**Pass / no new authority.**

## C6 — Approval/feedback/distribution history disposed separately

Would create orphan compliance facts or incomplete Revision history.

**Closure:** all subject-specific immutable B4 evidence follows Revision retention unit.

## C7 — Dossier link grants access

Rejected by frozen semantics. Every linked resource remains independently authorized.

**Pass.**

## C8 — Dossier hold misses a concurrently entering Revision/Evidence

Solved by stable Dossier root + subject/link lock ordering/materialization contracts.

**Proof required before implementation.**

## C9 — hold arrives while disposition worker deletes

Solved by RetentionBinding serialization + immutable DispositionFence. Hold wins before Fence or materialization fails after Fence; no silent crossing.

**Pass at architecture level.**

## C10 — Artifact shared by disposed and surviving subject

Deleting by subject would destroy surviving truth.

**Closure:** physical deletion only for zero-surviving-typed-reference Artifact; shared byte identity remains.

## C11 — delete DocumentRevision row after disposal

Would permit `max(revision_no)+1` to reuse an ordinal.

**Closure:** Revision identity skeleton never deleted by normal disposition.

## C12 — provider ObjectLock treated as LegalHold authority

Provider mechanism cannot represent Document/Dossier live scopes or future materialization.

**Closure:** MetalDocs Hold owns semantics; provider lock remains optional R10-C enforcement.

---

# 25. Proof obligations

| Claim | Required falsification/proof |
|---|---|
| Dossier never grants access | actor with Dossier access but no Document/Evidence access cannot read linked target |
| Dossier key scope | same key allowed across different types, rejected within same type |
| eligibility future-only | removing eligible type does not invalidate existing link/capture, blocks new one |
| Evidence capture immutability | every CAPTURED payload mutation/update path rejected |
| one primary Artifact | CAPTURE cannot commit without exact primary Artifact; two primary payloads impossible |
| native occurred anchor | OCCURRED_AT type cannot native-capture unknown occurred_at |
| sequence monotonic | concurrent CAPTURE using `{SEQ}` gets distinct committed values; rollback does not reuse committed values |
| first-submission binding | no successful first RevisionSubmission without same-commit RetentionBinding |
| no resnapshot | same-REV resubmit after policy change retains original Binding policy |
| first-capture binding | no successful CAPTURE without same-commit RetentionBinding |
| effective no clock | current EFFECTIVE Revision never becomes disposition eligible regardless elapsed wall time |
| superseded anchor | due derives from winning ReleaseRecord time, not mutable revision field |
| extension monotonic | extension cannot shorten due or apply to unknown anchor/Indefinite policy |
| hold current scope | Hold creation materializes all current Bindings in exact scope |
| hold future scope | first future Submission/CAPTURE/link while Hold active materializes new Binding |
| unlink cannot release | removing Dossier link leaves prior LegalHoldSubject intact |
| hold/disposition race | concurrency test proves either active hold blocks Fence or committed Fence blocks new materialization |
| no auto-delete | expiry alone causes no semantic/provider deletion |
| disposition truth | no DispositionRecord before verified physical phase + final semantic cleanup |
| revision ordinal | disposed REV row remains and next revision number never reuses it |
| origin survives source disposal | derived DocumentOrigin remains intelligible after source Submission/Artifact disposal |
| dictionary disposal | disposed Revision can remove dictionary snapshot without mutating Revision identity row |
| Artifact no-orphan | deferred DB enforcement rejects confirmed Artifact with zero typed semantic refs |
| Artifact shared preservation | disposal cannot delete bytes/Artifact while any surviving typed ref exists |
| Evidence lifecycle vs disposition | disposed CAPTURED/VOIDED Evidence does not gain a fake DISPOSED business state |
| tenant/area posture | no universal tenant partition or generic scope/object registry re-enters B5 |

---

# 26. Adversarial challenge

## F1 — Dossier should simply be a folder/container

That would make contextual membership imply ownership/location and invites inherited access.

**Reject.** Dossier stays stable context only.

## F2 — Evidence can reuse DocumentRevision lifecycle

It would add REV/Approval/Release machinery to captures that need only mutable DRAFT → immutable CAPTURE.

**Reject.** Separate Evidence is essential complexity.

## F3 — generic Record table is simpler

It reduces table count but introduces a second subject-identity registry and a declaration lifecycle with no consumer.

**Reject.** RetentionBinding is a closed union of the two real V1 retention subjects.

## F4 — store retention on Artifact

One Artifact can be referenced by multiple subjects with different retention/hold states. Artifact-level policy becomes ambiguous and provider-driven.

**Reject.** Preservation derives from subject references.

## F5 — expiry can enqueue delete automatically

A hold may appear, eligibility may be wrong, or explicit disposition review may decide not to dispose.

**Reject.** Expiry only makes eligible.

## F6 — LegalHold can be just Object Lock/WORM

Provider cannot represent Dossier/Document scope or future subject materialization, and a provider migration would move business authority.

**Reject.** Provider enforcement remains mechanism.

## F7 — keep source Template Submission forever to preserve DocumentOrigin

This silently changes retention to “indefinite while any derived document exists”, a requirement never frozen.

**Reject.** Immutable origin provenance snapshot is smaller and lifecycle-independent.

## F8 — delete whole Revision row on disposition

Breaks non-reuse of revision ordinals and downstream revision identity references.

**Reject.** Keep minimal identity skeleton.

## F9 — keep full Revision row including dictionary snapshot forever

Defeats lawful Records disposition.

**Reject.** Split retention payload from permanent ordinal skeleton.

## F10 — use Artifact ref_count for deletion

Counter can drift under repair/migration/concurrency and becomes a second authority.

**Reject.** Typed reachability is truth.

## F11 — allow LegalHold to materialize after DispositionFence and “do best effort”

Would claim preservation over content already entering irreversible removal.

**Reject.** Fence makes the conflict explicit/fail-closed.

## F12 — retain Records-Governance records through the same RetentionBinding mechanism

Creates recursive retention semantics with no terminating owner.

**Reject.** V1 retains minimal control evidence directly; future dedicated regime requires a real trigger.

---

# 27. Essential vs accidental complexity / YAGNI

## Essential

- stable Dossier context + typed scope;
- M:N Dossier↔Document context without grants;
- specialized Evidence capture boundary;
- exact Evidence Artifact;
- explicit retention rule snapshot per governed subject;
- lifecycle-derived retention anchor;
- append-only extensions;
- LegalHold scope + exact materialized subject set;
- future hold materialization for Document/Dossier scope;
- explicit no-hold disposition fence;
- verified completion record;
- typed Artifact reachability;
- minimal post-disposition identity skeletons.

## Accidental / deferred

- physical folder ownership semantics;
- Dossier hierarchy/graph;
- Dossier custom-object/forms/workflow/completeness engine;
- generic `Record` aggregate/declaration;
- generic retention expressions/jurisdiction matrix;
- generic polymorphic subject/Artifact owner registry;
- provider ObjectLock as business authority;
- automatic expiration deletion;
- eDiscovery/custodian/ESI graph;
- recursive retention of Records-Governance control rows;
- dynamic ref-count authority;
- multi-file ArtifactPackage without real consumer;
- disposition BPM approval engine;
- Dossier-local Evidence sequence reset without requirement.

---

# 28. Candidate decision

Recommended integrated B5 Global Maximum:

```text
Documentary Context
  DossierType
  Dossier + typed scope
  explicit type eligibility
  contextual Document links
  stable ExternalReferences

Evidence
  EvidenceType + allowed formats + bounded naming
  Evidence DRAFT working state
  immutable EvidenceCapture
  immutable Void evidence
  one primary Dossier / optional secondary context

Records Governance
  explicit type retention rules
  immutable RetentionBinding snapshot
  append-only RetentionExtension
  LegalHold + materialized RetentionBindings
  explicit DispositionFence
  verified immutable DispositionRecord

Artifact
  exact bytes only
  no own retention policy
  no generic owner registry
  survival by typed semantic reachability
```

With three bounded B3 refinements:

```text
1. DocumentOrigin hard FK to source Submission
   → self-contained exact source provenance snapshot + source Revision skeleton FK

2. DocumentRevision terminal retention anchors
   → canonical cancelled_at / obsoleted_at CI facts

3. dictionary_snapshot co-located on permanent Revision identity
   → separate immutable RevisionDictionarySnapshot retention payload
```

No B4 semantic reopen is required by B5.

---

# 29. Reopen triggers

Reopen only on real evidence such as:

- Dossier becomes a true independent business master requiring fields/workflow beyond documentary context;
- real Dossier hierarchy or Dossier-to-Dossier relation is required;
- Evidence needs independent REV/Approval lifecycle;
- true indivisible multi-file Evidence requires ArtifactPackage;
- a third real retention-subject family makes the two-member closed union materially inadequate;
- jurisdiction/conditional retention requires a real policy rules engine;
- required retention shortening/retroactive policy recalculation appears;
- Dossier-local Evidence sequence reset is a real user/business requirement;
- disposition requires regulated multi-stage approval/eSignature beyond one authorized decision;
- LegalHold must cover ESI/custodians/unconfirmed drafts beyond governed retention subjects;
- Records control records require their own formal retention regime;
- B6/R10-F privacy review proves the proposed post-disposition skeleton retains disallowed Target Data;
- SQL/concurrency proof shows hold materialization + disposition cannot be linearized under B1 lock law.

Legacy table shape, provider limitations or implementation convenience are not reopen triggers.

---

# 30. Whole-R10 posture

B5 remains **NON-AUTHORITATIVE CANDIDATE** until operator review/adjudication.

If accepted:

```text
R10-B3 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
  + bounded B5-driven technical refinements in §4

R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL

R10-B5 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL

implementation = BLOCKED
next = R10-B6 integrated candidate + B3/B4/B5/B6 coherence challenge
```

Whole-R10 independent cold review remains deferred until the integrated design is complete unless a material exception trigger appears.