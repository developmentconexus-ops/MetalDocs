# R10-B3 — Controlled Information + Artifact Relational Core — Integrated Candidate

> **Status:** NON-AUTHORITATIVE CANDIDATE — OPERATOR REVIEW / R10 INTEGRATION PENDING  
> **Date:** 2026-08-17  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Input HEAD:** `111299f167f7bc959ab8d5ec9215474cc460e21c`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Independent review:** deliberately deferred to Whole-R10 unless a material exception trigger appears.

This file is staging analysis. It does **not** alter the frozen product/domain ledger, `wiki/architecture/r10-technical-architecture.md`, current handoff, or any promoted authority. It is intended to be challenged by later R10 blocks and eventually by the Whole-R10 independent review.

---

## 1. Authority and evidence boundary

Authority used, in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`

Current code/schema/OpenAPI/module docs are current-state evidence only.

External references are evidence/comparison only and never MetalDocs authority:

- PostgreSQL conditional uniqueness / partial unique indexes: <https://www.postgresql.org/docs/current/ddl-constraints.html>
- PostgreSQL row locking / READ COMMITTED considerations: <https://www.postgresql.org/docs/current/transaction-iso.html> and <https://www.postgresql.org/docs/current/explicit-locking.html>
- RFC 8785 JSON Canonicalization Scheme: <https://www.rfc-editor.org/rfc/rfc8785.html>
- M-Files templates as a role/property of ordinary documents: <https://userguide.m-files.com/user-guide/latest/eng/using_template.html>
- M-Files immutable version-history posture (rollback creates a new version): <https://userguide.m-files.com/user-guide/latest/eng/object_history.html>
- Veeva QualityDocs Periodic Review records bound to the steady-state document/version and stale-version safeguards: <https://quality.veevavault.help/en/lr/72024/>
- SharePoint versioning/content-approval comparison: <https://learn.microsoft.com/en-us/sharepoint/governance/versioning-content-approval-and-check-out-planning>

The comparison result is narrow: mature systems reinforce version separation, document-role templates, explicit review records and concurrency/version guards, but MetalDocs keeps its already-frozen `DocumentRevision` / `WorkingContent` / `RevisionSubmission` semantics rather than copying another product's lifecycle or workflow engine.

---

# 2. Evidence → Known / Inferred / Unknown / Deferred

## 2.1 Known — frozen/promoted inputs

The candidate must preserve:

- `DocumentType` with immutable code, display fields, optional classification-only category and ACTIVE/INACTIVE lifecycle;
- `Document` as stable governed identity with stable code / DocumentType / Area;
- official business revisions `REV001`, `REV002`, ...; REV numbers never reuse;
- `REV002+` reason-for-change before first SUBMIT;
- revision states `DRAFT | SUBMITTED | EFFECTIVE | SUPERSEDED | OBSOLETE | CANCELLED`;
- at most one open Revision and at most one EFFECTIVE Revision per Document;
- format-agnostic mutable `WorkingContent` protected by one monotonic `working_version` OCC generation;
- whole-WorkingContent replacement semantics;
- immutable `RevisionSubmission` for every attempt, including NoHumanApproval;
- same-REV return/resubmit creates a new Submission rather than mutating an old attempt;
- Submission digest binds exact source Artifact hash + governed state + decision-relevant structured/template provenance, never provider location;
- Artifact as immutable exact-byte technical identity/facts, provider-neutral;
- exactly one primary Artifact per DocumentRevision V1;
- Template as a role of ordinary governed Document, not a parallel lifecycle;
- `TemplateUse` M:N and exact source effective REV pinned at derived-document creation;
- optional `TemplateSpec` only where structured authoring requires it;
- `EditorialComment` as DRAFT collaboration state; unresolved comments block SUBMIT;
- Periodic Review `Disabled | Every(n months)`; overdue never invalidates EFFECTIVE content; immutable record binds exact reviewed REV and outcome;
- Tenant Dictionary values are pinned for a new REV and do not silently re-resolve on same-REV return/resubmit;
- small product-owned System Value Catalog;
- B1 substrate: one PostgreSQL product DB / `metaldocs` schema / UUID PKs / typed FKs / no universal tenant partition / READ COMMITTED / cross-owner RESTRICT-NO ACTION / one local transaction for frozen local atomicity;
- B2 identity/access laws, User/Area lifecycle and lock ordering remain closed.

## 2.2 Inferred — candidate technical choices

The following are technical realizations, not new frozen product requirements:

1. `revision_no` is the only persisted official REV ordinal; `REV%03d` is derived display.
2. `Document` carries only stable identity and current stewardship; revision-varying governed metadata belongs to WorkingContent/Submission.
3. `WorkingContent` is exactly one current row per business Revision and owns the one OCC generation.
4. Submission uses a bounded versioned manifest rather than a growing set of provider-/format-specific snapshot columns.
5. System Value Catalog remains static product code; no editable catalog table V1.
6. Template role is a typed relational subtype (`TemplateDocument`) so `TemplateUse` can reference the role structurally without a Template aggregate.
7. Periodic Review reuses `Document.responsible_user_id` as the canonical current responsible owner rather than creating a second owner authority.
8. A confirmed `Artifact` row represents confirmed exact-byte facts; temporary staging state is a later R10-C mechanism, not a second semantic Artifact lifecycle.
9. Artifact SHA-256 is not a global relational uniqueness key: two captures of identical bytes may carry different technical provenance later; physical deduplication remains provider/mechanism freedom.

## 2.3 Unknown — explicitly bounded

These do not block the relational candidate:

- whether a future UX needs exactly one default TemplateUse per DocumentType;
- whether a future product requirement needs numbering-policy changes after first allocation;
- whether Periodic Review later needs more verdict detail than the minimal V1 outcome set below;
- whether a future authoring model needs same-REV introduction of new Tenant Dictionary dependencies after revision creation.

Unknown remains unknown. The candidate prepares seams but does not add dormant machinery.

## 2.4 Deferred by R10 stage ownership

```text
ApprovalPolicy / ApprovalInstance / ApprovalDecision / fresh-auth → B4
Rendition / Release / effectivity transaction                    → B4
Distribution                                                     → B4
Evidence / Dossier / Records Governance Artifact relations       → B5
final Artifact ownership closure / disposition interaction       → B5
Audit final fields / Interchange / cross-owner matrix            → B6
malware / physical storage / relocation / restore                → R10-C
async execution / projections / external provider effects        → R10-D
API / frontend / editor journeys                                 → R10-E
historical migration / cutover / legacy deletion                 → R10-F
```

---

# 3. Root Cause

The current-state model fragments one controlled-information truth across separate `controlled_documents`, multiple `documents` representations, technical `document_revisions`, provider/editor state and a parallel Template lifecycle. That structure permits different bytes/metadata/revision notions to become authoritative at different points.

The structural failure class is **split-brain governed identity**:

```text
business document identity
≠ business revision identity
≠ editor/autosave generation
≠ submitted immutable attempt
≠ exact bytes
```

These facts must be distinct, but each must have exactly one owner and explicit transitions between them. Reusing today's tables with renamed columns would be a local maximum because it preserves the conditions that allowed editor truth, freeze truth and approval truth to diverge.

---

# 4. Target invariant

> For each Document there is one stable governed identity. For each business change cycle there is one DocumentRevision. While that Revision is DRAFT there is one mutable WorkingContent protected by one monotonic OCC generation. SUBMIT atomically consumes exactly one accepted WorkingContent generation and creates one immutable RevisionSubmission whose manifest/digest and source Artifact identify the exact governed candidate that every downstream decision may reference.

Corollaries:

```text
Document          != bytes
DocumentRevision  != autosave/checkpoint
WorkingContent    != browser/editor session state
RevisionSubmission!= DocumentRevision
Artifact          != provider location/key/version
Template          != parallel aggregate/lifecycle
PeriodicReview    != mutable last-reviewed summary
EditorialComment  != provider comment identity
```

---

# 5. Credible alternatives

## A. Evolve current tables in place

Keep `controlled_documents`, current `documents`, technical `document_revisions` and TemplateVersion, then tighten constraints.

**Rejected:** smallest migration delta, but preserves duplicate identities, provider/storage fields in business rows and parallel lifecycle paths. Local maximum.

## B. One fat DocumentRevision row

Put current content, metadata, hashes, template state, approval/rendition fields and periodic-review summaries onto one revision row.

**Rejected:** lowers table count while mixing incompatible mutation laws. Mutable DRAFT and immutable review evidence become column-state conventions rather than structural boundaries.

## C. Small relational kernel with mutation-law separation — recommended

```text
Document
→ DocumentRevision
→ WorkingContent + working_version
→ immutable RevisionSubmission
→ exact Artifact

+ small typed configuration/provenance/review adjuncts
```

**Recommended Global Maximum:** it removes duplicate authority without introducing generic ECM/BPM/object-platform machinery.

---

# 6. Integrated relational candidate

Names below are semantic target names; exact SQL naming/casing is implementation-spec work.

## 6.1 Artifact core — supporting semantic owner `Artifact`

```text
Artifact
  id UUID PRIMARY KEY
  sha256 BYTEA NOT NULL CHECK octet_length(sha256)=32
  size_bytes BIGINT NOT NULL CHECK size_bytes >= 0
  content_format TEXT NOT NULL CHECK closed ContentFormat vocabulary
  media_type TEXT NOT NULL
  confirmed_at TIMESTAMPTZ NOT NULL
```

Mutation law:

- row fields are immutable after insert;
- provider bucket/key/version/URL are absent;
- no global `UNIQUE(sha256)` V1;
- physical staging, scan evidence, storage location and relocation are R10-C;
- deletion of an unsubmitted Artifact identity is permitted only through typed-owner closure; B5 completes the global owner/disposition law before Whole-R10 ratification.

**Confirmation seam:** temporary bytes may exist before classification, but a successful local confirmation transaction inserts the `Artifact` semantic row only together with its first typed governed owner relationship. A DB failure leaves only temporary provider staging, never a confirmed orphan Artifact row.

No `owner_type/owner_id` generic registry is introduced.

---

## 6.2 Controlled Information configuration

### DocumentTypeCategory

```text
DocumentTypeCategory
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  name TEXT NOT NULL
  description TEXT NULL
```

- optional classification/navigation only;
- `code` immutable;
- no governance, workflow or AuthZ semantics;
- delete RESTRICT while referenced.

### DocumentType

```text
DocumentType
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  name TEXT NOT NULL
  description TEXT NULL
  category_id UUID NULL FK DocumentTypeCategory(id) RESTRICT
  status TEXT NOT NULL CHECK ACTIVE|INACTIVE
  numbering_pattern TEXT NOT NULL
  numbering_scope TEXT NOT NULL CHECK TYPE|TYPE_AREA
  sequence_padding INTEGER NOT NULL CHECK sequence_padding >= 1
```

Laws:

- code immutable;
- numbering grammar is only literals + `{TYPE}` / `{AREA}` / `{SEQ}`;
- numbering policy is mutable only before first allocation; after first committed allocation, pattern/scope/padding become immutable V1;
- no numbering-policy version family until a real requirement needs post-use changes;
- INACTIVE blocks creation of new Documents but does not invalidate existing Documents/Revisions;
- Approval and representation configuration stay out of B3.

The pattern parser is a closed domain parser with DB backstop/guard in the implementation spec, not a generic formula engine.

### DocumentNumberSeries

Durable mechanism owned by Controlled Information:

```text
DocumentNumberSeries
  id UUID PRIMARY KEY
  document_type_id UUID NOT NULL FK DocumentType(id) RESTRICT
  area_id UUID NULL FK Area(id) RESTRICT
  next_value BIGINT NOT NULL CHECK next_value >= 1
```

Structural uniqueness:

```text
UNIQUE(document_type_id)          WHERE area_id IS NULL
UNIQUE(document_type_id, area_id) WHERE area_id IS NOT NULL
```

Interpretation:

- `TYPE` scope uses the NULL-area series;
- `TYPE_AREA` scope uses the concrete Area series;
- allocation is transactional; committed values never decrement/reuse;
- final sequence and final rendered code are persisted on Document;
- preview is advisory UI/API behavior and reserves nothing (R10-E).

### TenantDictionaryValue

```text
TenantDictionaryValue
  id UUID PRIMARY KEY
  name TEXT NOT NULL UNIQUE
  value TEXT NOT NULL
  label TEXT NOT NULL
  description TEXT NULL
```

- company-level current truth; no `tenant_id` partition column;
- `name` immutable; value/display fields mutable;
- whole-company `dictionary.manage` target;
- relevant values are frozen into each new Revision snapshot below.

### System Value Catalog

No editable persistence V1. It remains a small static product catalog. Only resolved values that materially enter the submitted candidate are frozen in governed state/manifest.

---

## 6.3 Stable Document identity and responsibility

```text
Document
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  document_type_id UUID NOT NULL FK DocumentType(id) RESTRICT
  area_id UUID NOT NULL FK Area(id) RESTRICT
  sequence_no BIGINT NOT NULL CHECK sequence_no >= 1
  responsible_user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL
```

Mutation law:

- `id`, `code`, `document_type_id`, `area_id`, `sequence_no`, `created_at` immutable;
- `responsible_user_id` mutable through explicit `document.owner.manage` operation;
- no current/effective/open revision pointer duplicates authoritative Revision state;
- no provider/editor/storage identity;
- revision-varying title/governed metadata is not stored as mutable Document identity.

B2 coherence:

- new Document requires ACTIVE DocumentType, non-retired Area and enabled responsible User;
- existing Document remains valid after Area retirement or User offboarding; those lifecycle actions do not silently rewrite governed history;
- assigning a new responsible User requires that User to be enabled at commit;
- if future evidence requires “every live Document always has an enabled responsible User” across offboarding, that is a new cross-owner invariant and reopen trigger, not an implicit cascade.

AuthZ target classification:

- DocumentType/category/numbering/dictionary/TemplateUse configuration = Tenant-wide;
- Document content/revision/comment/periodic-review/owner operations = Area-targeted by immutable `Document.area_id` plus the applicable domain relationship/governance predicate;
- Artifact has no end-user mechanism permission family.

---

## 6.4 Business Revision

```text
DocumentRevision
  id UUID PRIMARY KEY
  document_id UUID NOT NULL FK Document(id) RESTRICT
  revision_no INTEGER NOT NULL CHECK revision_no >= 1
  state TEXT NOT NULL CHECK DRAFT|SUBMITTED|EFFECTIVE|SUPERSEDED|OBSOLETE|CANCELLED
  created_by_user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL
  dictionary_snapshot JSONB NOT NULL
```

Constraints:

```text
UNIQUE(document_id, revision_no)
UNIQUE(document_id) WHERE state IN ('DRAFT','SUBMITTED')
UNIQUE(document_id) WHERE state = 'EFFECTIVE'
```

Interpretation:

- `revision_no=1` renders `REV001`; no separate persisted revision label authority;
- at-most-one open = one `DRAFT|SUBMITTED` row;
- zero open is valid when no change cycle exists;
- at-most-one EFFECTIVE is structural; B4 owns the winning Release transition that makes it effective;
- new Revision allocation locks the Document and chooses `max(revision_no)+1`; cancelled/superseded ordinals never reuse;
- `dictionary_snapshot` is an immutable bounded whole snapshot of the relevant Tenant Dictionary values resolved for that Revision at creation;
- same-REV return/resubmit never re-resolves this snapshot;
- if future authoring proves same-REV introduction of new dictionary dependencies is required, reopen this snapshot rule deliberately rather than silently consulting current dictionary values.

Mixed-row immutable columns require DB-enforced immutability in the implementation spec; application convention alone is insufficient.

---

## 6.5 WorkingContent — sole mutable DRAFT authority

```text
WorkingContent
  revision_id UUID PRIMARY KEY FK DocumentRevision(id) RESTRICT
  primary_artifact_id UUID NOT NULL FK Artifact(id) RESTRICT
  governed_metadata JSONB NOT NULL
  structured_authoring JSONB NULL
  reason_for_change TEXT NULL
  working_version BIGINT NOT NULL CHECK working_version >= 1
  updated_by_user_id UUID NOT NULL FK User(id) RESTRICT
  updated_at TIMESTAMPTZ NOT NULL
```

Semantics:

```text
WorkingContent
= exact current primary Artifact
+ bounded governed metadata
+ optional bounded structured-authoring state
+ reason-for-change
+ one monotonic working_version
```

Laws:

- only mutable while owning Revision is DRAFT;
- every governed DRAFT mutation uses caller-observed `expected_working_version`;
- successful mutation increments `working_version` exactly once;
- provider/browser/editor state is never authoritative;
- replacing primary Artifact is whole-WorkingContent replacement and atomically removes/invalidate representation-dependent structured state/provenance that no longer applies;
- `REV002+` requires non-blank `reason_for_change` before first SUBMIT;
- tracked-change state, when an enabled authoring model has it, is part of structured/governed state and the same OCC generation; no separate collaboration authority;
- no business autosave/checkpoint revision family.

Technical `WorkingSnapshot` remains absent from the minimum candidate. It may be added as an explicitly technical recovery/checkpoint family only if a concrete recovery consumer proves the need; it can never consume `REVxxx` or become approval truth.

### EditorSession

A narrow optional persisted lease is justified by the already-frozen “one active in-app writer + OCC” posture:

```text
EditorSession
  id UUID PRIMARY KEY
  revision_id UUID NOT NULL FK DocumentRevision(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  acquired_at TIMESTAMPTZ NOT NULL
  expires_at TIMESTAMPTZ NOT NULL
  released_at TIMESTAMPTZ NULL
```

At most one non-expired logical writer is enforced by the application/lease contract; correctness never depends on the lease. OCC remains the race arbiter. No long checkout.

---

## 6.6 Immutable RevisionSubmission

```text
RevisionSubmission
  id UUID PRIMARY KEY
  revision_id UUID NOT NULL FK DocumentRevision(id) RESTRICT
  source_artifact_id UUID NOT NULL FK Artifact(id) RESTRICT
  accepted_working_version BIGINT NOT NULL
  manifest_schema TEXT NOT NULL
  manifest_payload JSONB NOT NULL
  submission_digest BYTEA NOT NULL CHECK octet_length(submission_digest)=32
  submitted_by_user_id UUID NOT NULL FK User(id) RESTRICT
  submitted_at TIMESTAMPTZ NOT NULL

UNIQUE(revision_id, accepted_working_version)
```

Mutation law: immutable/append-only; serving product paths have no UPDATE/DELETE.

No `UNIQUE(submission_digest)`: two legitimate attempts may submit the same governed candidate and remain distinct attempt identities.

### Manifest V1

The manifest is a bounded product schema, not arbitrary metadata. Conceptual payload:

```json
{
  "document": {
    "id": "uuid",
    "code": "...",
    "document_type_code": "...",
    "area_id": "uuid"
  },
  "revision": {
    "id": "uuid",
    "number": 2
  },
  "source": {
    "sha256": "lowercase-hex",
    "size_bytes": "decimal-string",
    "content_format": "...",
    "media_type": "..."
  },
  "governed_metadata": {},
  "dictionary_snapshot": {},
  "structured_authoring": null,
  "reason_for_change": "...",
  "template_provenance": null
}
```

Attempt metadata (`submitted_at`, submitter, Submission UUID), provider location/key/version, renderer output and future Approval evidence are deliberately excluded from the candidate digest.

### Canonical digest

Recommended candidate algorithm:

```text
canonical_payload = RFC8785_JCS(manifest_payload)
submission_digest = SHA256(
  UTF8(manifest_schema)
  || 0x00
  || canonical_payload
)
```

Why this candidate beats alternatives:

- custom concatenation is fragile and format-coupled;
- deterministic CBOR is credible but introduces an additional serialization/tooling stack without a demonstrated consumer;
- JCS is a published deterministic JSON canonicalization scheme and matches MetalDocs' Go/TypeScript/JSON ecosystem.

Manifest inputs must satisfy I-JSON/JCS rules. Potentially >53-bit integral values use canonical decimal strings rather than depending on JavaScript numeric precision. High-precision domain numbers, if later admitted into governed metadata, must have a domain-defined canonical string form before inclusion.

Proof requires golden vectors across Go and TypeScript: same semantic manifest → byte-identical canonical form/digest; each included field mutation changes digest; provider/storage-location mutation does not.

---

## 6.7 Template role / use / specification / origin

### TemplateDocument

```text
TemplateDocument
  document_id UUID PRIMARY KEY FK Document(id) RESTRICT
```

Presence means an ordinary governed Document currently has Template role. There is no Template identity/version/lifecycle aggregate.

### TemplateUse

```text
TemplateUse
  template_document_id UUID NOT NULL FK TemplateDocument(document_id) RESTRICT
  target_document_type_id UUID NOT NULL FK DocumentType(id) RESTRICT
  PRIMARY KEY(template_document_id, target_document_type_id)
```

- current M:N eligibility/configuration;
- whole-company `template_use.manage` target;
- no `is_default` V1 because no promoted requirement currently proves a default-template fact;
- if UX later proves a default is required, an `is_default` fact plus partial unique index can be added without changing Template identity.

### TemplateSpec

```text
TemplateSpec
  revision_id UUID PRIMARY KEY FK DocumentRevision(id) RESTRICT
  spec JSONB NOT NULL
```

- only applicable to a Revision whose Document has Template role;
- bounded structured-authoring schema, never generic custom-object metadata;
- mutable only while that Revision is DRAFT and only in the same WorkingContent OCC mutation;
- exact submitted spec is frozen in the RevisionSubmission manifest;
- no TemplateSpec version counter independent from REV/Submission.

DB/application enforcement must reject TemplateSpec on a non-template Document.

### DocumentOrigin

```text
DocumentOrigin
  derived_document_id UUID PRIMARY KEY FK Document(id) RESTRICT
  source_template_submission_id UUID NOT NULL FK RevisionSubmission(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL
```

Mutation law: immutable.

This single source Submission pins the exact source template Revision and exact source Artifact without four parallel provenance pointers. The derived Document's own WorkingContent becomes content authority immediately after creation; later template revisions never silently rebind it.

**B4 successor obligation:** create-from-template must serialize with effectivity change and validate that `source_template_submission_id` is the winning source of the Template Document's current EFFECTIVE Revision at commit. B3 establishes the stale-source invariant; B4 supplies Release/effectivity mechanics.

---

## 6.8 EditorialComment — material DRAFT collaboration state

EditorialComment is retained because unresolved comments change SUBMIT eligibility.

```text
EditorialComment
  id UUID PRIMARY KEY
  revision_id UUID NOT NULL FK DocumentRevision(id) RESTRICT
  author_user_id UUID NOT NULL FK User(id) RESTRICT
  body TEXT NOT NULL
  created_at TIMESTAMPTZ NOT NULL
  resolved_at TIMESTAMPTZ NULL
  resolved_by_user_id UUID NULL FK User(id) RESTRICT
```

Laws:

- comment body/author/creation identity immutable;
- resolution is a terminal product state change for that comment;
- no provider/library comment ID as business identity;
- no threads/edit/delete/generic annotation platform V1 without evidence;
- create comment and resolve comment both require `expected_working_version` and atomically increment WorkingContent `working_version`, because they alter SUBMIT eligibility;
- SUBMIT sees zero unresolved EditorialComments for that exact Revision under the same OCC/lock boundary.

If a future editor needs location anchors, only a bounded provider-neutral anchor shape may enter this table; provider-native IDs remain adapter/support state.

---

## 6.9 Periodic Review

### Policy

```text
PeriodicReviewPolicy
  document_id UUID PRIMARY KEY FK Document(id) RESTRICT
  interval_months INTEGER NOT NULL CHECK interval_months > 0
```

Interpretation:

- row absent = `Disabled`;
- row present = `Every(n months)`;
- no policy inheritance/default engine from DocumentType V1;
- `Document.responsible_user_id` is the canonical current responsible owner; no duplicate review-owner field.

### Record

Minimal V1 candidate outcome set:

```text
KEEP_EFFECTIVE
CHANGE_REQUIRED
```

This intentionally does not copy richer QMS workflow verdict sets. `CHANGE_REQUIRED` records the governance conclusion; creating a new Revision or obsoleting the Document remains an explicit later operation with its own authorization/lifecycle semantics.

```text
PeriodicReviewRecord
  id UUID PRIMARY KEY
  revision_id UUID NOT NULL FK DocumentRevision(id) RESTRICT
  reviewed_by_user_id UUID NOT NULL FK User(id) RESTRICT
  responsible_user_id UUID NOT NULL FK User(id) RESTRICT
  outcome TEXT NOT NULL CHECK KEEP_EFFECTIVE|CHANGE_REQUIRED
  policy_interval_months INTEGER NOT NULL CHECK policy_interval_months > 0
  reviewed_at TIMESTAMPTZ NOT NULL
```

Mutation law: immutable/append-only; no UPDATE/DELETE serving path.

- `responsible_user_id` snapshots the canonical responsible owner at review time;
- `policy_interval_months` snapshots the policy used for that review evidence;
- record must bind the current EFFECTIVE Revision at commit;
- overdue does not alter Revision state or access;
- no mutable `last_reviewed_at` or `review_due_at` authority column is required.

Candidate due function:

```text
anchor = max(
  current EFFECTIVE Revision effective_at,
  latest PeriodicReviewRecord.reviewed_at for that same EFFECTIVE Revision
)
next_due_at = anchor + PeriodicReviewPolicy.interval_months
```

`effective_at` is B4 Release-owned state. R10-D may surface due/overdue projections/jobs but never owns this semantic calculation.

**B4 successor obligation:** Periodic Review and winning Release must conflict on the same per-Document serialization point so a review cannot commit against a Revision that ceased to be current EFFECTIVE during the transaction.

External comparison: Veeva QualityDocs similarly models a separate Periodic Review record around the steady-state document/version and offers guards against reviews when a later version exists. MetalDocs keeps a much smaller record rather than importing Veeva's configurable workflow/change-control platform.

---

# 7. Persistence class × mutation law

| Family | Class | Mutation law |
|---|---|---|
| Artifact | supporting semantic authority | fields immutable; typed-owner-controlled deletion only when lawful |
| DocumentTypeCategory | semantic authority | code immutable; display mutable |
| DocumentType | semantic authority | code immutable; numbering immutable after first allocation; display/category/status mutable |
| DocumentNumberSeries | durable mechanism | mutable monotonic counter |
| TenantDictionaryValue | semantic authority | name immutable; value/display mutable |
| Document | semantic authority | identity/type/Area/sequence immutable; responsibility mutable |
| DocumentRevision | semantic authority | identity/ordinal/snapshot immutable; explicit lifecycle state machine |
| WorkingContent | semantic authority | mutable DRAFT only through OCC |
| EditorSession | ephemeral/attributed mechanism | lease lifecycle; never correctness authority |
| RevisionSubmission | semantic authority | immutable/append-only |
| TemplateDocument | semantic current role | add/remove only through typed constraints; no history authority |
| TemplateUse | semantic current configuration | add/remove current relationship |
| TemplateSpec | semantic DRAFT state | mutable DRAFT only through same OCC; Submission freezes exact state |
| DocumentOrigin | semantic provenance | immutable |
| EditorialComment | semantic DRAFT collaboration | append + terminal resolve |
| PeriodicReviewPolicy | semantic current policy | mutable/disable by explicit operation |
| PeriodicReviewRecord | semantic evidence | immutable/append-only |

Immutable rows must be structurally non-updatable by the serving trust surface. Mixed mutable rows require DB-enforced immutable-column guards; “the service does not normally update this” is not proof.

---

# 8. Structural constraint envelope

```text
DocumentTypeCategory.code                 UNIQUE
DocumentType.code                         UNIQUE
TenantDictionaryValue.name                UNIQUE
Document.code                             UNIQUE

DocumentRevision(document_id, revision_no) UNIQUE
DocumentRevision one DRAFT|SUBMITTED       partial UNIQUE(document_id)
DocumentRevision one EFFECTIVE             partial UNIQUE(document_id)

DocumentNumberSeries TYPE                  partial UNIQUE(document_type_id) WHERE area_id IS NULL
DocumentNumberSeries TYPE_AREA             partial UNIQUE(document_type_id, area_id) WHERE area_id IS NOT NULL

WorkingContent.revision_id                 exactly one row / PK
RevisionSubmission(revision_id,
                   accepted_working_version) UNIQUE

TemplateDocument.document_id               PK
TemplateUse(template_document_id,
            target_document_type_id)        PK
TemplateSpec.revision_id                    PK
DocumentOrigin.derived_document_id          PK
PeriodicReviewPolicy.document_id            PK

Artifact.sha256                             length = 32 bytes
RevisionSubmission.submission_digest        length = 32 bytes
```

PostgreSQL partial unique indexes are the intended structural mechanism for conditional uniqueness; no application-only “check then insert” is accepted for one-open / one-effective / NumberSeries uniqueness.

No cross-owner CASCADE/SET NULL is introduced.

---

# 9. Transaction contracts

## 9.1 Atomic Document + REV001 creation

Pre-transaction mechanism may stage/generate initial bytes. The authoritative local commit is:

```text
BEGIN
  validate enabled responsible User under B2 lock law
  validate non-retired Area
  lock/read ACTIVE DocumentType + immutable numbering configuration
  allocate one NumberSeries value
  render final immutable Document.code

  if template-derived:
    validate TemplateUse
    serialize/validate exact effective source template Submission

  confirm/insert exact Artifact semantic row
  insert Document
  insert DocumentRevision REV001 / DRAFT + dictionary snapshot
  insert WorkingContent working_version=1
  insert optional TemplateSpec / DocumentOrigin

  // B6 adds required Audit append; R10-D adds required durable intent where needed
COMMIT
```

There is no successful “empty governed Document shell” without a coherent REV001 WorkingContent primary Artifact. Provider staging may fail independently before this transaction; provider cleanup is mechanism work.

Allocation + code + Document + REV001 + WorkingContent either all commit or all roll back.

## 9.2 New Revision creation

```text
BEGIN
  lock Document FOR UPDATE
  prove no DRAFT|SUBMITTED Revision exists
  allocate next revision_no = max+1
  resolve fresh relevant Tenant Dictionary snapshot
  seed initial primary Artifact/content from the authorized source state
  insert DocumentRevision DRAFT
  insert WorkingContent with fresh working_version=1
COMMIT
```

No separate revision counter or parent graph V1.

## 9.3 DRAFT mutation CAS

Every content/metadata/spec/comment-eligibility mutation follows one contract:

```text
expected_working_version = N

BEGIN
  prove Revision = DRAFT
  CAS/lock WorkingContent at N
  apply the complete governed mutation
  atomically invalidate representation-dependent state when required
  working_version = N + 1
COMMIT
```

Exactly one of two concurrent callers with the same observed N may win.

## 9.4 SUBMIT freeze

```text
BEGIN
  lock Document / target Revision / WorkingContent in canonical order
  prove Revision = DRAFT
  prove caller expected_working_version = N
  prove final provider flush is already represented by WorkingContent N
  prove primary Artifact is confirmed exact-byte state
  prove REV002+ reason-for-change
  prove zero unresolved EditorialComment
  prove no forbidden tracked-change state when applicable

  build bounded manifest from one coherent state
  canonicalize with manifest_schema + RFC8785 JCS
  compute SHA-256 submission_digest

  insert immutable RevisionSubmission(
    revision,
    source Artifact,
    accepted_working_version=N,
    manifest,
    digest,
    submitter/time
  )

  transition Revision DRAFT -> SUBMITTED
  increment WorkingContent working_version N -> N+1
  invalidate/release active EditorSession mechanism
COMMIT
```

**The SUBMIT increment is mandatory.** It consumes the pre-submit OCC generation. A late autosave that observed N cannot later become valid merely because B4 returns the same Revision to DRAFT. Any return/reopen operation must preserve monotonicity and must never reset `working_version`.

This closes the adversarial late-write-after-return counterexample without creating a second epoch/generation mechanism.

---

# 10. Lock / concurrency law

B1 remains READ COMMITTED. B2 lock classes remain authoritative and precede B3 locks when required.

B3 canonical order after applicable B2 `User → Area` eligibility locks:

```text
DocumentType
→ DocumentNumberSeries
→ Document row(s), UUID order when more than one
→ DocumentRevision
→ WorkingContent
→ EditorialComment child set, UUID order when a set lock is necessary
→ Artifact row only when existing-row coordination is required
```

Classes may be skipped but never acquired backwards.

Lock strength rules:

- User/Area eligibility checks use the B2 reader strength (`FOR SHARE`) where applicable;
- DocumentType creation/allocation reads hold a lock compatible with concurrent readers but conflicting with numbering-policy mutation;
- NumberSeries allocation updates/locks the exact series row;
- lifecycle mutations on a target Document use `FOR UPDATE` on Document as the per-document serialization root;
- source Template validation uses a lock that conflicts with B4 effectivity change on that same source Document;
- Periodic Review uses the same target Document serialization root as B4 Release;
- DRAFT saves that never acquire Document after WorkingContent may use the narrower CAS path; they must not later acquire an earlier lock class.

No global SERIALIZABLE, advisory-lock framework, distributed lock or long checkout is justified.

---

# 11. Authority / boundary

```text
Controlled Information owns
  DocumentType/category/numbering/dictionary semantics
  stable Document identity and responsibility
  business Revision
  mutable WorkingContent/OCC
  exact Submission
  template role/use/spec/origin
  EditorialComment
  periodic-review semantics

Artifact owns
  exact-byte immutable facts
  confirmation contract
  later physical integrity/location mechanics

Organization owns
  User / Area identity and lifecycle

Authorization owns
  live grant evaluation

Approval later owns
  human decision over exact Submission

Audit later owns
  transversal timeline, never current Document state
```

Artifact therefore has no product permission such as `artifact.read`, `artifact.replace`, `storage.migrate` or provider-specific operations. Access is authorized through the owning business object.

---

# 12. B1/B2 coherence

B3 introduces no:

- universal `tenant_id/company_id/deployment_id` partition column;
- Tenant/Area/role/Permission RLS policy engine;
- cross-owner CASCADE/SET NULL;
- provider role/group/claim authority;
- custom role/permission family;
- provider/editor/storage identity as business identity.

B3 permission target classification:

```text
Tenant-wide:
  document_type.manage
  template_use.manage
  dictionary.manage

Area-targeted by Document.area_id:
  document.read_*
  document.create
  document.edit
  document.comment
  document.submit
  document.cancel_revision
  document.obsolete
  document.review_periodic
  document.owner.manage
```

This does not redefine role bundles; it only supplies the domain target classification B2 required successor stages to declare.

---

# 13. Proof obligations

| Claim | Falsification/proof obligation |
|---|---|
| one open Revision | concurrent new-Revision attempts cannot commit two DRAFT/SUBMITTED rows |
| one EFFECTIVE | competing B4 Releases hit structural one-effective backstop |
| numbering uniqueness | concurrent same-series creates get distinct committed values/codes |
| no sequence reuse | cancel/obsolete/rollback behavior never decrements a committed allocation |
| OCC | two mutations using same expected N produce exactly one commit |
| coherent SUBMIT | no committed SUBMITTED Revision exists without its exact immutable Submission and vice versa |
| late-write exclusion | pre-SUBMIT autosave cannot mutate SUBMITTED or a later returned DRAFT after SUBMIT consumed N |
| exact source | Submission.source_artifact equals WorkingContent primary Artifact at accepted generation |
| digest determinism | cross-runtime golden vectors produce identical canonical bytes/digest |
| digest sensitivity | mutating every included semantic input changes digest |
| provider independence | moving provider/key/location does not change digest |
| comment gate | SUBMIT concurrent with comment create/resolve cannot bypass unresolved-comment invariant |
| Template origin | later template Revision never rewrites existing DocumentOrigin |
| stale template defense | create-from-template cannot commit source that ceased to be current EFFECTIVE during transaction |
| periodic-review staleness | review cannot commit against Revision that ceased to be current EFFECTIVE during transaction |
| immutable evidence | serving paths cannot UPDATE/DELETE Submission, Origin or ReviewRecord |
| Artifact no-orphan | confirmation always creates typed owner; replacement/disposition cannot leave confirmed unowned semantic row after B5 closure |
| B2 lifecycle | disabled User/retired Area cannot become new responsibility/scope target where prohibited |

Architecture proof before implementation: counterexample analysis and B1/B2 coherence. Implementation proof later: real PostgreSQL concurrency tests, negative constraint tests, restart/retry tests, digest golden vectors and end-to-end exact-Submission identity checks.

---

# 14. Adversarial challenge

## F1 — late autosave survives a return to DRAFT

Counterexample if SUBMIT does not advance OCC:

1. writer observes `working_version=N`;
2. SUBMIT freezes N and marks SUBMITTED;
3. B4 later returns the REV to DRAFT;
4. stale writer with N saves successfully.

**Candidate correction:** SUBMIT consumes N and advances to N+1. Return never resets generation.

## F2 — Template source becomes stale during derived-document create

Reading “current effective template” and committing later without shared serialization can pin a source that was superseded meanwhile.

**Candidate correction:** source Template Document is locked/validated against B4 effectivity at commit; DocumentOrigin stores exact source Submission.

## F3 — Periodic Review records stale effective content

Review reads current effective REV; Release switches it before review commit.

**Candidate correction:** Periodic Review and B4 Release share the Document serialization root.

## F4 — Artifact replacement creates confirmed orphan bytes

Swapping WorkingContent pointer without disposing/retaining the previous typed relation can leave semantic Artifact state with no governed owner.

**Candidate disposition:** B3 requires CI-local typed-owner closure on replacement; B5 must close the final union across DocumentRevision/Evidence/Records before Whole-R10 ratification. No generic owner registry is added pre-emptively.

## F5 — JSONB becomes a hidden generic object platform

`governed_metadata`, `structured_authoring`, manifest and TemplateSpec could become escape hatches.

**Candidate correction:** every JSONB field is a bounded owned whole snapshot/schema. No arbitrary field-definition/custom-object engine, and schema/version expansion is a material owner decision.

## F6 — hash uniqueness conflates provenance

Global `UNIQUE(sha256)` would make same bytes from distinct captures share one row and could silently collapse technical provenance.

**Candidate correction:** canonical SHA-256 proves exact bytes; UUID remains Artifact row identity; provider-level dedupe remains mechanism freedom.

## F7 — responsible owner offboarding causes hidden cascade

Automatically reassigning/cancelling Documents during User offboarding would create new cross-owner semantics not frozen in B2/B3.

**Candidate correction:** stable User UUID reference may remain; new responsibility assignment requires enabled User. A stronger always-enabled-owner invariant is a reopen trigger.

## F8 — numbering config mutation reinterprets history

If pattern/scope changes after allocations, reconstructing why codes/sequences exist requires policy history/versioning.

**Candidate correction:** numbering policy freezes after first allocation V1. Future real need reopens into explicit policy versioning rather than adding it speculatively now.

---

# 15. Essential vs accidental complexity / YAGNI

## Essential — retained

- stable Document identity;
- business Revision lineage;
- transactional numbering;
- exact bytes;
- one mutable DRAFT authority;
- OCC generation;
- immutable Submission attempts;
- template source provenance;
- editorial submit gate;
- periodic-review evidence;
- typed Organization relationships;
- structural conditional uniqueness.

## Accidental — removed/deferred

- separate ControlledDocument aggregate;
- Template/TemplateVersion lifecycle;
- generic numbering formula/reset engine;
- generic metadata/custom-object framework;
- business autosave/checkpoint versions;
- generic annotation/thread platform;
- universal tenant partition/RLS substrate;
- provider storage keys in business rows;
- provider/editor revision IDs as business identity;
- mandatory PDF/render fields in B3;
- generic BPM/workflow for Periodic Review;
- default-template machinery without a real consumer;
- ArtifactPackage/multi-file PLM abstraction;
- content-addressed dedupe as business authority.

Prepare seams, not future capability.

---

# 16. Decision — candidate only

**Proposed outcome: `RESTRUCTURE NOW` at target-design level.**

Adopt the small relational kernel:

```text
Artifact

DocumentTypeCategory
DocumentType
DocumentNumberSeries
TenantDictionaryValue

Document
DocumentRevision
WorkingContent
EditorSession
EditorialComment
RevisionSubmission

TemplateDocument
TemplateUse
TemplateSpec
DocumentOrigin

PeriodicReviewPolicy
PeriodicReviewRecord
```

with three primary transaction contracts:

```text
atomic Document + REV001 creation
one-CAS DRAFT mutation
coherent SUBMIT freeze
```

and one per-Document lifecycle serialization root consumed by later B4 Release/PeriodicReview/source-template validation.

This candidate is **not independently ratified**. If operator accepts it, the next state is:

```text
R10-B3 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
```

not `CLOSED / APPROVED` in the old micro-review sense. It remains challengeable by B4–F and the final Whole-R10 review.

Implementation remains BLOCKED.

---

# 17. Reopen triggers

Reopen only on material evidence:

- true indivisible multi-file governed content requiring ArtifactPackage;
- legitimate multiple simultaneous open business Revisions per Document;
- real post-use numbering-policy changes requiring versioned numbering policy;
- category begins driving governance rather than classification/navigation;
- a concrete default-template consumer requires canonical default semantics;
- same-REV authoring must introduce/re-resolve new dictionary dependencies after Revision creation;
- richer Periodic Review verdict/state semantics have a real consumer;
- responsibility must be Group-/role-based rather than one User;
- real-time coauthoring/merge semantics invalidate one-writer+OCC;
- provider/editor representation becomes incapable of final coherent flush before SUBMIT;
- cryptographic/signature requirements demand a different canonical submission representation;
- B4/B5 proves the proposed Document serialization root or typed Artifact closure cannot preserve its invariants without a structural change.

Implementation inconvenience, current schema shape or a hypothetical future ECM feature is not a reopen trigger.
