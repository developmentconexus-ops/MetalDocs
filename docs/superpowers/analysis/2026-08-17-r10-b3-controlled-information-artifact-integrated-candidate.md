# R10-B3 — Controlled Information + Artifact Relational Core — Integrated Candidate

> **Status:** NON-AUTHORITATIVE CANDIDATE — OPERATOR REVIEW / R10 INTEGRATION PENDING  
> **Date:** 2026-08-17  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Input authority HEAD:** `111299f167f7bc959ab8d5ec9215474cc460e21c`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Independent review:** deferred to Whole-R10 unless a material exception trigger appears.

This file is staging analysis. It does **not** alter the frozen product/domain ledger, `wiki/architecture/r10-technical-architecture.md`, current handoff, or any promoted authority. If accepted by the operator, B3 becomes **ACCEPTED FOR R10 INTEGRATION / NON-FINAL**, not independently ratified. Later R10 blocks may reopen only the materially implicated part when a real counterexample appears.

---

# 1. Authority and evidence boundary

Authority used, in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`

Current code/schema/OpenAPI/module docs are current-state evidence only.

External references are evidence/comparison only, never MetalDocs authority:

- PostgreSQL conditional uniqueness / partial unique indexes: <https://www.postgresql.org/docs/current/ddl-constraints.html>
- PostgreSQL READ COMMITTED / row-lock semantics: <https://www.postgresql.org/docs/current/transaction-iso.html> and <https://www.postgresql.org/docs/current/explicit-locking.html>
- RFC 8785 JSON Canonicalization Scheme: <https://www.rfc-editor.org/rfc/rfc8785.html>
- M-Files document-template and version-history posture: <https://userguide.m-files.com/user-guide/latest/eng/using_template.html> and <https://userguide.m-files.com/user-guide/latest/eng/object_history.html>
- Veeva QualityDocs Periodic Review comparison: <https://quality.veevavault.help/en/lr/72024/>
- SharePoint versioning/content-approval/checkout comparison: <https://learn.microsoft.com/en-us/sharepoint/governance/versioning-content-approval-and-check-out-planning>

Comparison result: mature systems reinforce version separation, explicit review records, document-role templates and concurrency/version guards. MetalDocs does **not** copy their lifecycle/workflow engines; its frozen `DocumentRevision` / `WorkingContent` / `RevisionSubmission` semantics remain authority.

---

# 2. Evidence → Known / Inferred / Unknown / Deferred

## 2.1 Known — frozen/promoted inputs

B3 must preserve:

- `DocumentType`: immutable code, display fields, optional classification-only category, ACTIVE/INACTIVE;
- `Document`: stable governed identity with stable code / DocumentType / Area;
- official business revisions `REV001`, `REV002`, ...; numbers never reuse;
- `REV002+` requires reason-for-change before first SUBMIT;
- Revision states `DRAFT | SUBMITTED | EFFECTIVE | SUPERSEDED | OBSOLETE | CANCELLED`;
- at most one open Revision and at most one EFFECTIVE Revision per Document;
- format-agnostic mutable `WorkingContent` protected by one monotonic `working_version` OCC generation;
- whole-WorkingContent replacement semantics;
- immutable `RevisionSubmission` for every submit attempt, including NoHumanApproval;
- same-REV return/resubmit creates a new Submission, never mutates the prior attempt;
- Submission digest binds exact source Artifact hash + governed state + decision-relevant structured/template provenance, never provider location;
- `Artifact`: immutable exact-byte identity/facts with SHA-256, size, ContentFormat/media type and technical provenance;
- no confirmed orphan Artifact library;
- one current primary Artifact for a Revision WorkingContent; immutable historical Submission attempts preserve their exact source Artifact;
- Template is a role of ordinary governed Document, not a parallel lifecycle;
- `TemplateUse` is M:N; derived creation pins exact source effective REV/content;
- optional `TemplateSpec` only where structured authoring requires it;
- `EditorialComment` is DRAFT collaboration state; unresolved comments block SUBMIT;
- Periodic Review = `Disabled | Every(n months)`; overdue does not invalidate EFFECTIVE; immutable record binds exact reviewed REV and outcome;
- Tenant Dictionary values are pinned for a new REV and do not silently re-resolve on same-REV return/resubmit;
- System Value Catalog remains small/product-owned;
- B1: one PostgreSQL product DB / schema `metaldocs`, UUID PKs, typed FKs, no universal tenant partition, cross-owner RESTRICT/NO ACTION, READ COMMITTED, local same-commit atomicity;
- B2 identity/access/User/Area/lock laws remain closed.

## 2.2 Inferred — technical candidate choices

These are candidate realization choices, not product-authority rewrites:

1. `revision_no` is the only persisted official REV ordinal; `REV%03d` is derived.
2. Revision-varying governed metadata belongs to WorkingContent/Submission, not stable Document identity.
3. `WorkingContent` is one current mutable row per open Revision and owns the sole OCC generation.
4. Submission uses a bounded versioned manifest rather than format/provider-specific snapshot columns.
5. The manifest canonical form is RFC 8785 JCS and its digest is SHA-256.
6. System Value Catalog remains static product code; no editable catalog table V1.
7. Template role uses a typed relational subtype (`TemplateDocument`) rather than a Template aggregate.
8. Periodic Review uses `Document.responsible_user_id` as current responsibility authority and snapshots that User on the immutable review record.
9. A confirmed `Artifact` row is provider-neutral semantic exact-byte state; temporary provider staging is R10-C mechanism state.
10. SHA-256 is not a global Artifact-row uniqueness key; identical bytes may have distinct capture/provenance identities while physical deduplication remains mechanism freedom.
11. Numbering configuration becomes immutable after its first committed allocation; versioned numbering policy is deferred until a real post-use change requirement exists.

## 2.3 Unknown — bounded, not converted into defaults

- whether a future UX needs a canonical default TemplateUse per DocumentType;
- whether a real future requirement needs numbering-policy mutation after first use;
- whether Periodic Review needs a richer outcome vocabulary;
- whether a future authoring model needs same-REV introduction of new Tenant Dictionary dependencies after Revision creation;
- terminal cleanup timing for non-authoritative WorkingContent/EditorSession rows after a Revision can no longer return to DRAFT.

These Unknowns do not block B3. The candidate prepares seams, not dormant capabilities.

## 2.4 Deferred by stage ownership

```text
ApprovalPolicy / ApprovalInstance / ApprovalDecision / fresh-auth → B4
Rendition / Release / effectivity transaction                    → B4
Distribution                                                     → B4
Evidence / Dossier / Records Governance Artifact relations       → B5
final Artifact ownership/disposition closure                     → B5
Audit final fields / Interchange / cross-owner matrix            → B6
malware / physical storage / relocation / restore                → R10-C
async execution / projections / external provider effects        → R10-D
API / frontend / editor journeys                                 → R10-E
historical migration / cutover / deletion                        → R10-F
```

---

# 3. Root Cause

Current state fragments one controlled-information truth across separate `controlled_documents`, multiple `documents` representations, technical `document_revisions`, editor/provider state and a parallel Template lifecycle.

The structural failure class is **split-brain governed identity**:

```text
business Document identity
!= business Revision identity
!= mutable editor/autosave generation
!= immutable submitted attempt
!= exact bytes
```

Those concepts are legitimately different, but each needs one authority and explicit transitions. Renaming/tightening the existing tables would preserve the condition that historically allowed editor truth, freeze truth and approval truth to diverge.

---

# 4. Target invariant

> For each Document there is one stable governed identity. For each business change cycle there is one DocumentRevision. While that Revision is DRAFT there is one mutable WorkingContent protected by one monotonic OCC generation. SUBMIT atomically consumes exactly one accepted WorkingContent generation and creates one immutable RevisionSubmission whose manifest/digest and source Artifact identify the exact governed candidate that every downstream decision may reference.

Corollaries:

```text
Document           != bytes
DocumentRevision   != autosave/checkpoint
WorkingContent     != browser/editor session
RevisionSubmission != Revision
Artifact           != provider location/key/version
Template           != parallel aggregate/lifecycle
PeriodicReview     != mutable last-reviewed summary
EditorialComment   != provider comment identity
```

---

# 5. Constraints inherited from B1/B2

- one MetalDocs PostgreSQL DB / `metaldocs` schema;
- UUID technical identity;
- no universal `tenant_id/company_id/deployment_id` partition column;
- typed FKs only for business relations;
- cross-owner FK actions `RESTRICT | NO ACTION` only;
- frozen vocabulary `TEXT + CHECK` by default;
- canonical SHA-256 `BYTEA` with 32-byte length check;
- JSONB only for bounded owned whole snapshots/provenance, never an escape hatch;
- ordinary product mutations under explicit local transactions;
- default isolation `READ COMMITTED`; narrow locks/constraints/CAS before stronger isolation;
- B2 User/Area eligibility and lock classes precede B3 locks when both are needed;
- current code/schema inconvenience is not target authority.

---

# 6. Credible alternatives

## A — evolve legacy tables in place

Keep `controlled_documents`, current `documents`, technical `document_revisions` and parallel TemplateVersion; tighten constraints.

**Local Maximum / reject.** Small migration delta but preserves duplicate identities and incompatible mutation laws.

## B — one fat DocumentRevision row

Put mutable content, hashes, template data, approval/rendition fields and review summaries on one row.

**Local Maximum / reject.** Fewer tables but mutable DRAFT and immutable review evidence become column conventions instead of structural boundaries.

## C — small relational kernel separated by mutation law

```text
Document
→ DocumentRevision
→ WorkingContent + working_version
→ immutable RevisionSubmission
→ exact Artifact

+ small typed configuration / provenance / review adjuncts
```

**Recommended Global Maximum.** Removes duplicate authority without creating generic ECM/BPM/object-platform machinery.

---

# 7. Integrated relational candidate

Names are semantic target names. Exact SQL naming is implementation-spec work.

## 7.1 Artifact core — supporting semantic owner `Artifact`

```text
Artifact
  id UUID PRIMARY KEY
  sha256 BYTEA NOT NULL CHECK octet_length(sha256)=32
  size_bytes BIGINT NOT NULL CHECK size_bytes >= 0
  content_format TEXT NOT NULL CHECK closed ContentFormat vocabulary
  media_type TEXT NOT NULL
  technical_provenance JSONB NOT NULL CHECK jsonb_typeof(...)='object'
  confirmed_at TIMESTAMPTZ NOT NULL
```

Mutation law:

- all semantic fields immutable after insert;
- provider bucket/key/version/URL absent;
- `technical_provenance` is bounded/provider-neutral provenance, not physical location authority;
- no global `UNIQUE(sha256)` V1;
- staging, malware result, managed physical location, relocation and restore are R10-C;
- no `owner_type/owner_id` generic business registry.

### Confirmation seam

Temporary provider staging may precede classification. A successful **semantic confirmation** inserts the immutable Artifact row only in the same local DB transaction that establishes its first typed governed owner relation. DB rollback may leave temporary staging for later mechanism cleanup, but never a confirmed orphan semantic Artifact.

B3 proves the Controlled Information side. B5 must close the final union when Evidence/Records relationships are designed.

---

## 7.2 Controlled Information configuration

### DocumentTypeCategory

```text
DocumentTypeCategory
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  name TEXT NOT NULL
  description TEXT NULL
```

- optional classification/navigation only;
- code immutable;
- no governance/workflow/AuthZ semantics;
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
- numbering grammar = literals + `{TYPE}` / `{AREA}` / `{SEQ}` only;
- `{SEQ}` appears exactly once;
- `TYPE_AREA` requires `{AREA}` so independent per-Area sequences cannot produce structurally identical codes for equal sequence numbers;
- final `Document.code UNIQUE` remains the global collision backstop;
- numbering config mutable only before first committed allocation, then immutable V1;
- no numbering-policy-version family until real post-use mutation is required;
- INACTIVE blocks new Document creation but never invalidates existing Documents/Revisions;
- Approval/representation config remains outside B3.

A closed parser validates the pattern; it is not a formula/reset/custom-metadata engine.

### DocumentNumberSeries

Durable allocation mechanism owned by Controlled Information:

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

- `TYPE` → `area_id IS NULL`;
- `TYPE_AREA` → concrete Area;
- series is created on first allocation if absent;
- first successful series/allocation commit is also the structural point after which DocumentType numbering config cannot change;
- allocation monotonically advances `next_value`; committed sequence values never decrement/reuse;
- preview is read-only/advisory and reserves nothing (R10-E).

A DB invariant guard in the implementation spec must prove the series shape matches `DocumentType.numbering_scope`; application convention alone is insufficient.

### TenantDictionaryValue

```text
TenantDictionaryValue
  id UUID PRIMARY KEY
  name TEXT NOT NULL UNIQUE
  value TEXT NOT NULL
  label TEXT NOT NULL
  description TEXT NULL
```

- company-level current truth; no tenant partition column;
- name immutable; value/display mutable;
- whole-company `dictionary.manage` target;
- relevant values freeze into each new Revision snapshot.

### System Value Catalog

No editable persistence V1. Small static product catalog. Resolved system values that materially affect the submitted candidate enter WorkingContent/Submission governed state; the catalog itself does not become mutable DB authority.

---

## 7.3 Stable Document identity / numbering provenance / responsibility

```text
Document
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  document_type_id UUID NOT NULL FK DocumentType(id) RESTRICT
  area_id UUID NOT NULL FK Area(id) RESTRICT
  number_series_id UUID NOT NULL FK DocumentNumberSeries(id) RESTRICT
  sequence_no BIGINT NOT NULL CHECK sequence_no >= 1
  responsible_user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL

  UNIQUE(number_series_id, sequence_no)
```

Mutation law:

- `id`, code, type, Area, series, sequence and creation instant immutable;
- `responsible_user_id` mutable only through explicit `document.owner.manage`;
- no current/effective/open Revision pointer duplicates authoritative Revision state;
- no provider/editor/storage identity;
- revision-varying title/governed metadata lives in WorkingContent/Submission.

Numbering consistency backstop:

- `number_series_id` must reference this same `document_type_id`;
- if DocumentType scope is TYPE, series Area must be NULL;
- if scope is TYPE_AREA, series Area must equal `Document.area_id`.

Because this invariant crosses rows, the implementation spec must use a database-enforced guard/constraint mechanism covering all write paths; a service-only check is insufficient.

B2 coherence:

- new Document requires ACTIVE DocumentType, non-retired Area and enabled responsible User at commit;
- existing Document remains valid after Area retirement/User offboarding; those lifecycle changes do not rewrite governed history;
- new responsibility assignment requires enabled User;
- if future evidence requires every live Document to always reference an enabled responsible User across offboarding, that is a new cross-owner invariant/reopen trigger, not an implicit cascade.

AuthZ target classification:

- DocumentType/category/numbering/dictionary/TemplateUse configuration = Tenant-wide;
- Document content/revision/comment/periodic-review/owner operations = Area-targeted by immutable `Document.area_id` plus domain relationship/governance predicate;
- Artifact has no end-user mechanism permission family.

---

## 7.4 Business DocumentRevision

```text
DocumentRevision
  id UUID PRIMARY KEY
  document_id UUID NOT NULL FK Document(id) RESTRICT
  revision_no INTEGER NOT NULL CHECK revision_no >= 1
  state TEXT NOT NULL CHECK DRAFT|SUBMITTED|EFFECTIVE|SUPERSEDED|OBSOLETE|CANCELLED
  created_by_user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL
  dictionary_snapshot JSONB NOT NULL CHECK jsonb_typeof(...)='object'

  UNIQUE(document_id, revision_no)
```

Conditional uniqueness:

```text
UNIQUE(document_id) WHERE state IN ('DRAFT','SUBMITTED')
UNIQUE(document_id) WHERE state = 'EFFECTIVE'
```

Laws:

- `revision_no=1` renders `REV001`; no second persisted label authority;
- zero open is valid when no business change cycle exists;
- new Revision creation locks Document, proves no open Revision, allocates `max(revision_no)+1`;
- cancelled/superseded ordinals never reuse;
- dictionary snapshot is an immutable bounded map of relevant values resolved for that Revision at creation;
- same-REV return/resubmit never re-resolves it;
- if same-REV introduction of new dictionary dependencies becomes a real requirement, this rule reopens deliberately rather than silently consulting current dictionary state.

Immutable identity/snapshot columns need DB enforcement; state follows the explicit lifecycle state machine.

---

## 7.5 WorkingContent — sole mutable DRAFT authority

```text
WorkingContent
  revision_id UUID PRIMARY KEY FK DocumentRevision(id) RESTRICT
  primary_artifact_id UUID NOT NULL FK Artifact(id) RESTRICT
  governed_metadata JSONB NOT NULL CHECK jsonb_typeof(...)='object'
  structured_authoring JSONB NULL CHECK NULL OR jsonb_typeof(...)='object'
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

- authoritative/mutable only while Revision = DRAFT;
- persists through SUBMITTED when return-for-changes remains possible, but Submission is review authority during SUBMITTED;
- every governed DRAFT mutation uses caller-observed `expected_working_version`;
- successful governed mutation increments `working_version` exactly once;
- browser/editor/provider state never becomes authority;
- replacing primary Artifact is **whole-WorkingContent replacement**;
- representation-dependent structured-authoring state and TemplateSpec content that is no longer valid for the new representation must be atomically replaced/cleared in that same OCC mutation;
- immutable historical `DocumentOrigin` is never cleared by replacement;
- REV002+ requires nonblank reason-for-change before first SUBMIT;
- tracked-change state, when supported, is governed state under the same OCC generation;
- no business autosave/checkpoint Revision family.

Terminal retention/deletion of WorkingContent after a Revision can no longer return to DRAFT is deferred to B5/R10-F. No later consumer may treat terminal WorkingContent as official history; immutable Submission/Release facts own that history.

### WorkingSnapshot

Absent from minimum B3. It may be introduced only as technical recovery/checkpoint state when a concrete recovery consumer proves need; it never consumes REV numbers or becomes approval truth.

### EditorSession

Narrow mechanism justified by frozen one-active-in-app-writer posture:

```text
EditorSession
  id UUID PRIMARY KEY
  revision_id UUID NOT NULL FK DocumentRevision(id) RESTRICT
  user_id UUID NOT NULL FK User(id) RESTRICT
  acquired_at TIMESTAMPTZ NOT NULL
  expires_at TIMESTAMPTZ NOT NULL
  released_at TIMESTAMPTZ NULL
```

It is a lease/UX mechanism, never correctness authority. OCC remains the race arbiter. No long checkout.

---

## 7.6 Immutable RevisionSubmission

```text
RevisionSubmission
  id UUID PRIMARY KEY
  revision_id UUID NOT NULL FK DocumentRevision(id) RESTRICT
  source_artifact_id UUID NOT NULL FK Artifact(id) RESTRICT
  accepted_working_version BIGINT NOT NULL CHECK accepted_working_version >= 1
  manifest_schema TEXT NOT NULL
  manifest_payload JSONB NOT NULL CHECK jsonb_typeof(...)='object'
  submission_digest BYTEA NOT NULL CHECK octet_length(submission_digest)=32
  submitted_by_user_id UUID NOT NULL FK User(id) RESTRICT
  submitted_at TIMESTAMPTZ NOT NULL

  UNIQUE(revision_id, accepted_working_version)
```

Mutation law: immutable/append-only. Serving product path has no UPDATE/DELETE.

No `UNIQUE(submission_digest)`: separate legitimate attempts may submit the same governed candidate and remain different attempt identities.

### Manifest V1

Bounded product schema, conceptually:

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
  "template_spec": null,
  "template_provenance": null
}
```

`template_spec` is populated only when the submitted Revision is a template Revision with applicable structured specification. `template_provenance` is populated for derived Documents where `DocumentOrigin` is relevant.

Attempt metadata (Submission UUID, submitter, timestamp), provider location/key/version, renderer output and Approval evidence are excluded from the candidate digest.

### Canonical digest

Candidate:

```text
canonical_payload = RFC8785_JCS(manifest_payload)
submission_digest = SHA256(
  UTF8(manifest_schema)
  || 0x00
  || canonical_payload
)
```

Why:

- custom field concatenation is fragile/format-coupled;
- deterministic CBOR is credible but adds a second serialization stack without a demonstrated consumer;
- JCS provides deterministic JSON canonicalization compatible with the Go/TypeScript/JSON boundary already present.

Manifest input obeys I-JSON/JCS constraints. Potentially unsafe large integers use canonical decimal strings; future high-precision domain numbers require a domain-defined canonical string before inclusion.

Proof requires cross-runtime golden vectors:

- same semantic manifest → byte-identical canonical bytes/digest;
- mutation of each included semantic field changes digest;
- provider/storage relocation does not change digest.

---

## 7.7 Template role / use / specification / origin

### TemplateDocument

```text
TemplateDocument
  document_id UUID PRIMARY KEY FK Document(id) RESTRICT
```

Presence means an ordinary governed Document has Template role. No Template ID/version/lifecycle aggregate exists.

### TemplateUse

```text
TemplateUse
  template_document_id UUID NOT NULL FK TemplateDocument(document_id) RESTRICT
  target_document_type_id UUID NOT NULL FK DocumentType(id) RESTRICT

  PRIMARY KEY(template_document_id, target_document_type_id)
```

- current M:N eligibility/configuration;
- whole-company `template_use.manage` target;
- no `is_default` V1 because no promoted requirement currently proves a canonical default-template fact;
- if a real default consumer appears, the fact can be added with one partial unique invariant without changing Template identity.

### TemplateSpec

```text
TemplateSpec
  revision_id UUID PRIMARY KEY FK DocumentRevision(id) RESTRICT
  spec JSONB NOT NULL CHECK jsonb_typeof(...)='object'
```

- valid only when owning Document has Template role;
- bounded authoring schema, never generic custom-object metadata;
- mutable only while that Revision is DRAFT and only in the same WorkingContent OCC mutation;
- whole-WorkingContent Artifact replacement clears/replaces representation-dependent TemplateSpec state when no longer valid;
- exact submitted spec is copied into Submission manifest;
- no independent TemplateSpec version counter.

DB/application enforcement must reject TemplateSpec on a non-template Document.

### DocumentOrigin

```text
DocumentOrigin
  derived_document_id UUID PRIMARY KEY FK Document(id) RESTRICT
  source_template_submission_id UUID NOT NULL FK RevisionSubmission(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL
```

Immutable.

One source Submission pins exact source template Revision + exact source Artifact without four parallel provenance pointers. The derived Document's own WorkingContent is content authority immediately after creation; later template revisions never silently rebind it.

**B4 successor obligation:** create-from-template must serialize with template effectivity change and prove `source_template_submission_id` is the winning source of the Template Document's current EFFECTIVE Revision at commit. B3 establishes stale-source prevention; B4 supplies Release/effectivity mechanics.

---

## 7.8 EditorialComment — material DRAFT state

Retained because unresolved comments affect SUBMIT eligibility.

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

- body/author/creation immutable;
- resolution is terminal for that comment;
- no provider/library comment ID as business identity;
- no generic thread/edit/delete/annotation platform V1;
- create and resolve both require `expected_working_version` and increment WorkingContent `working_version`, because they alter submit eligibility;
- SUBMIT proves zero unresolved comments for that Revision under the same OCC/lock boundary.

Future location anchors, if required, must be bounded/provider-neutral; provider-native IDs remain adapter/support state.

---

## 7.9 Periodic Review

### PeriodicReviewPolicy

```text
PeriodicReviewPolicy
  document_id UUID PRIMARY KEY FK Document(id) RESTRICT
  interval_months INTEGER NOT NULL CHECK interval_months > 0
```

- absent row = Disabled;
- present row = Every(n months);
- no DocumentType inheritance/default engine V1;
- `Document.responsible_user_id` remains current responsibility authority.

### PeriodicReviewRecord

Minimal candidate outcome vocabulary:

```text
KEEP_EFFECTIVE
CHANGE_REQUIRED
```

`CHANGE_REQUIRED` records a review conclusion only. Creating a new Revision or obsoleting a Document is a separate authorized lifecycle operation.

```text
PeriodicReviewRecord
  id UUID PRIMARY KEY
  revision_id UUID NOT NULL FK DocumentRevision(id) RESTRICT
  reviewed_by_user_id UUID NOT NULL FK User(id) RESTRICT
  responsible_user_id UUID NOT NULL FK User(id) RESTRICT
  outcome TEXT NOT NULL CHECK KEEP_EFFECTIVE|CHANGE_REQUIRED
  policy_interval_months INTEGER NOT NULL CHECK policy_interval_months > 0
  due_at TIMESTAMPTZ NOT NULL
  reviewed_at TIMESTAMPTZ NOT NULL
```

Immutable/append-only.

- record binds exact current EFFECTIVE Revision at commit;
- `responsible_user_id` snapshots current responsible owner at review time;
- `policy_interval_months` snapshots policy applied;
- `due_at` snapshots the derived due boundary the review addressed, preserving historical interpretability if policy later changes;
- overdue never mutates Revision state/access;
- no mutable `last_reviewed_at` / `review_due_at` authority column is required.

Candidate due calculation:

```text
anchor = max(
  current EFFECTIVE Revision effective_at,
  latest PeriodicReviewRecord.reviewed_at for that same EFFECTIVE Revision
)
next_due_at = anchor + current PeriodicReviewPolicy.interval_months
```

No structural “exactly one review record per cycle” is invented: the frozen product semantics only require immutable records, not a prohibition on an explicitly authorized additional review. R10-E may apply idempotency to the request journey; the domain record remains truthful.

`effective_at` comes from B4 Release-owned state. R10-D may surface due/overdue projections/jobs but never owns calculation semantics.

**B4 successor obligation:** Periodic Review and winning Release share the same per-Document serialization root so a review cannot commit against a Revision that ceased to be current EFFECTIVE during the transaction.

---

# 8. Persistence class × mutation law

| Family | Class | Mutation law |
|---|---|---|
| Artifact | supporting semantic authority | semantic fields immutable; typed-owner-controlled lawful deletion only |
| DocumentTypeCategory | semantic authority | code immutable; display mutable |
| DocumentType | semantic authority | code immutable; numbering immutable after first allocation; display/category/status mutable |
| DocumentNumberSeries | durable mechanism | mutable monotonic allocation counter |
| TenantDictionaryValue | semantic authority | name immutable; value/display mutable |
| Document | semantic authority | identity/type/Area/series/sequence immutable; responsibility mutable |
| DocumentRevision | semantic authority | identity/ordinal/snapshot immutable; explicit state machine |
| WorkingContent | semantic authority while open | mutable DRAFT only through OCC; review source only through Submission |
| EditorSession | ephemeral/attributed mechanism | lease lifecycle; never correctness authority |
| RevisionSubmission | semantic authority | immutable/append-only |
| TemplateDocument | semantic current role | add/remove typed role; no parallel history authority |
| TemplateUse | semantic current configuration | add/remove typed current relationship |
| TemplateSpec | semantic DRAFT state | same OCC as WorkingContent; Submission freezes exact accepted state |
| DocumentOrigin | semantic provenance | immutable |
| EditorialComment | semantic DRAFT collaboration | append + terminal resolve |
| PeriodicReviewPolicy | semantic current policy | mutable/disable by explicit operation |
| PeriodicReviewRecord | semantic evidence | immutable/append-only |

Immutable rows must be structurally non-updatable by serving trust. Mixed mutable rows require DB-enforced immutable-column guards. “The service normally does not update it” is not proof.

---

# 9. Structural constraint envelope

```text
DocumentTypeCategory.code                   UNIQUE
DocumentType.code                           UNIQUE
TenantDictionaryValue.name                  UNIQUE
Document.code                               UNIQUE
Document(number_series_id, sequence_no)     UNIQUE

DocumentRevision(document_id, revision_no)  UNIQUE
DocumentRevision one DRAFT|SUBMITTED         partial UNIQUE(document_id)
DocumentRevision one EFFECTIVE               partial UNIQUE(document_id)

DocumentNumberSeries TYPE                    partial UNIQUE(document_type_id) WHERE area_id IS NULL
DocumentNumberSeries TYPE_AREA               partial UNIQUE(document_type_id, area_id) WHERE area_id IS NOT NULL

WorkingContent.revision_id                   PK / one current working row
RevisionSubmission(revision_id,
                   accepted_working_version) UNIQUE

TemplateDocument.document_id                 PK
TemplateUse(template_document_id,
            target_document_type_id)          PK
TemplateSpec.revision_id                      PK
DocumentOrigin.derived_document_id            PK
PeriodicReviewPolicy.document_id              PK

Artifact.sha256                               exactly 32 bytes
RevisionSubmission.submission_digest          exactly 32 bytes

bounded JSONB families                        explicit object/shape checks
```

PostgreSQL partial unique indexes are the structural mechanism for one-open, one-effective and nullable-scope series uniqueness. No application-only check-then-insert is accepted for those invariants.

Cross-row NumberSeries/type/Area coherence and immutable-column enforcement require explicit DB guards in the implementation spec. No cross-owner CASCADE/SET NULL.

---

# 10. Transaction contracts

## 10.1 Atomic Document + REV001 creation

Provider/editor mechanism may prepare/stage initial bytes before local semantic commit. Authoritative transaction:

```text
BEGIN
  validate enabled responsible User using B2 lock law
  validate non-retired Area
  lock/read ACTIVE DocumentType
  validate numbering grammar/config
  get/create exact NumberSeries and allocate one sequence
  render final immutable Document.code
  prove code uniqueness

  if template-derived:
    validate TemplateUse
    serialize/validate exact effective source template Submission

  confirm/insert exact Artifact semantic row + bounded technical provenance
  insert Document with exact number_series_id + sequence
  insert DocumentRevision REV001 / DRAFT + dictionary snapshot
  insert WorkingContent working_version=1
  insert optional TemplateSpec / DocumentOrigin

  // B6 later composes required Audit append
  // R10-D later composes durable intent when an actual external/async effect exists
COMMIT
```

No successful empty governed Document shell: a successful creation produces coherent stable identity + REV001 + WorkingContent + primary Artifact.

Allocation/code/Document/REV001/WorkingContent either all commit or all roll back. Provider staging cleanup is mechanism work and cannot manufacture semantic success.

## 10.2 New Revision creation

```text
BEGIN
  lock Document FOR UPDATE
  prove no DRAFT|SUBMITTED Revision exists
  allocate revision_no = max(existing)+1
  resolve fresh relevant Tenant Dictionary snapshot
  establish exact initial primary Artifact/content from authorized source state
  insert DocumentRevision DRAFT
  insert WorkingContent working_version=1
COMMIT
```

No separate revision counter/parent graph V1.

## 10.3 DRAFT mutation CAS

Every mutation that changes candidate content **or submit eligibility** follows one contract:

```text
expected_working_version = N

BEGIN
  prove Revision = DRAFT
  CAS/lock WorkingContent at N
  apply complete governed mutation
  atomically clear/replace representation-dependent state when required
  working_version = N + 1
COMMIT
```

Two concurrent callers with the same observed N cannot both win.

## 10.4 SUBMIT freeze

```text
BEGIN
  lock Document / target Revision / WorkingContent in canonical order
  prove Revision = DRAFT
  prove caller expected_working_version = N
  prove final provider flush is represented by WorkingContent N
  prove current primary Artifact is confirmed
  prove REV002+ reason-for-change
  prove zero unresolved EditorialComment
  prove no forbidden tracked-change state when applicable

  load exact TemplateSpec / DocumentOrigin facts relevant to N
  build one bounded submission manifest from this coherent state
  canonicalize manifest with RFC 8785 JCS
  compute SHA-256 submission_digest

  insert immutable RevisionSubmission(
    Revision,
    source Artifact,
    accepted_working_version=N,
    manifest,
    digest,
    submitter/time
  )

  transition Revision DRAFT -> SUBMITTED
  increment WorkingContent working_version N -> N+1
  release/invalidate active EditorSession mechanism
COMMIT
```

### Mandatory OCC generation consumption

SUBMIT **must** consume/increment the DRAFT generation.

Without it:

```text
writer observes N
→ SUBMIT freezes N
→ B4 returns same REV to DRAFT
→ stale pre-SUBMIT writer still carrying N could save
```

Advancing to N+1 closes that race without introducing a second epoch/generation mechanism. B4 return-for-changes may reopen the same Revision but never reset/decrement `working_version`.

---

# 11. Lock / concurrency law

B1 stays READ COMMITTED. B2 lock classes remain authoritative and precede B3 where both apply.

After applicable B2 `User → Area` eligibility locks:

```text
DocumentType
→ DocumentNumberSeries
→ Document row(s), UUID order when >1
→ DocumentRevision
→ WorkingContent
→ EditorialComment child set, deterministic order when set locking is needed
→ Artifact row only when existing-row coordination is required
```

Classes may be skipped, never revisited backwards.

Rules:

- B2 User/Area eligibility reads use promoted B2 reader lock strength;
- NumberSeries allocation locks/updates the exact series row;
- per-Document lifecycle transitions use `Document FOR UPDATE` as serialization root;
- source Template validation locks the source Document in a way that conflicts with B4 effectivity change;
- Periodic Review locks the target Document using the same root as B4 Release;
- narrow DRAFT save may CAS WorkingContent without later acquiring an earlier lock class;
- no global SERIALIZABLE, advisory-lock framework, distributed lock or long checkout.

---

# 12. Authority / boundary

```text
Controlled Information owns
  DocumentType/category/numbering/dictionary semantics
  stable Document identity/responsibility
  business Revision
  mutable WorkingContent/OCC
  exact immutable Submission
  template role/use/spec/origin
  EditorialComment
  periodic-review semantics

Artifact owns
  exact-byte immutable facts
  semantic confirmation contract
  later physical integrity/location mechanics

Organization owns
  User / Area identity/lifecycle

Authorization owns
  live grants and canonical evaluation

Approval later owns
  human decision over exact Submission

Audit later owns
  transversal timeline, never current Document truth
```

Artifact gains no product mechanism permissions such as `artifact.read`, `artifact.replace` or `storage.migrate`; access is authorized through the owning business object.

---

# 13. B1/B2 coherence

B3 introduces no:

- universal tenant/company/deployment partition column;
- Tenant/Area/role/Permission RLS policy engine;
- cross-owner CASCADE/SET NULL;
- provider role/group/claim authority;
- custom role/permission family;
- provider/editor/storage identity as business identity.

Permission target classification supplied to B2 successor contract:

```text
Tenant-wide:
  document_type.manage
  template_use.manage
  dictionary.manage

Area-targeted by Document.area_id:
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
```

This does not change bundles; it classifies the domain target for canonical evaluation.

---

# 14. Enforcement strategy

Strongest reasonable enforcement covering all reachable paths:

- PK/FK/CHECK/partial UNIQUE for representable relational invariants;
- database immutable-column guards where a row mixes mutable/immutable facts;
- database cross-row guard for NumberSeries ↔ DocumentType ↔ Area coherence;
- non-owner/NOSUPERUSER serving role; immutable evidence tables have no serving UPDATE/DELETE privilege;
- application CAS using `expected_working_version` for caller-observed concurrency;
- row locks for lifecycle serialization/cross-row eligibility where UNIQUE alone cannot express the claim;
- bounded JSON schemas + DB coarse shape checks; JSON never bypasses owner validation;
- tests must show each control firing; artifact/schema existence alone is not proof.

---

# 15. Proof obligations

| Claim | Falsification/proof obligation |
|---|---|
| one open Revision | concurrent new-Revision attempts cannot commit two DRAFT/SUBMITTED rows |
| one EFFECTIVE | competing B4 Releases hit the partial unique backstop |
| numbering | concurrent same-series creates receive distinct committed sequence/code |
| series provenance | Document cannot pair a series with the wrong type/Area/scope |
| no committed sequence reuse | cancellation/deletion/rollback never decrements a committed allocation |
| OCC | two mutations using same expected N produce exactly one successful commit |
| coherent SUBMIT | SUBMITTED and its immutable Submission cannot commit independently |
| late-write exclusion | pre-SUBMIT write cannot mutate SUBMITTED or a later returned DRAFT after SUBMIT consumes N |
| exact source | Submission.source_artifact equals WorkingContent primary Artifact at accepted N |
| TemplateSpec coherence | submitted TemplateSpec belongs to same accepted WorkingContent generation |
| digest determinism | Go/TypeScript golden vectors produce identical canonical bytes/digest |
| digest sensitivity | mutation of each included semantic input changes digest |
| provider independence | storage key/provider/location change leaves digest unchanged |
| comment gate | SUBMIT concurrent with comment create/resolve cannot bypass unresolved-comment invariant |
| Template provenance | later Template Revision cannot rewrite existing DocumentOrigin |
| stale Template defense | create-from-template cannot commit a source that ceased to be current EFFECTIVE |
| stale periodic-review defense | review cannot commit against Revision that ceased to be current EFFECTIVE |
| immutable evidence | serving path cannot UPDATE/DELETE Submission, Origin or ReviewRecord |
| Artifact ownership | semantic confirmation has typed owner; B5 closes global owner/disposition union |
| B2 lifecycle | disabled User/retired Area cannot become prohibited new responsibility/reference target |

Architecture proof now: counterexample analysis + B1/B2 coherence. Implementation proof later: real PostgreSQL concurrency/negative constraint tests, privilege tests, restart/retry tests, digest golden vectors and exact-Submission E2E checks.

---

# 16. Adversarial challenge

## F1 — late autosave survives return-to-DRAFT

If SUBMIT does not advance OCC, a writer holding pre-submit N can become valid again after return.

**Closed in candidate:** SUBMIT consumes N → N+1; return never resets generation.

## F2 — Template source becomes stale during derived creation

Read current EFFECTIVE template, then concurrent Release supersedes it before derived creation commits.

**B3 contract:** immutable Origin points exact Submission; source Document serialization must conflict with B4 effectivity change. B4 must satisfy the other half.

## F3 — Periodic Review records stale effective content

Review reads current EFFECTIVE; concurrent Release switches it before review commit.

**B3 contract:** review and Release share target Document serialization root. B4 must satisfy the other half.

## F4 — Artifact replacement creates confirmed orphan state

Changing WorkingContent pointer can strand the prior confirmed Artifact semantic row.

**B3 contract:** CI replacement closes its typed ownership transition in the same local mutation. **B5 is a hard Whole-R10 prerequisite** for final global owner/disposition closure. No generic owner registry is added prematurely.

## F5 — JSONB becomes hidden generic object platform

Governed metadata/structured authoring/manifest/spec can become arbitrary custom-object storage.

**Closed:** each JSONB family is an owned bounded whole snapshot with coarse DB shape checks + explicit owner schema. New generic field-definition/custom-object semantics require a material decision.

## F6 — SHA uniqueness collapses provenance

Global unique hash would conflate identical bytes captured through materially different provenance.

**Closed:** UUID is Artifact row identity; SHA proves bytes; provider dedupe remains mechanism freedom.

## F7 — owner offboarding causes hidden cascade

Automatic Document reassignment/cancellation during User offboarding would create new product semantics.

**Closed:** stable User UUID reference may survive; new assignment requires enabled User. Stronger always-enabled-owner semantics is a reopen trigger.

## F8 — numbering config mutation reinterprets history

Changing pattern/scope after allocations requires reconstructing which policy produced historical codes.

**Closed for V1:** numbering config freezes at first committed allocation. Real mutation need reopens into explicit policy versioning.

## F9 — NumberSeries points at wrong scope/Area

A TYPE_AREA document could accidentally consume a TYPE series or another Area's series, producing a unique-looking but semantically false code.

**Closed structurally:** Document stores `number_series_id`; `(series,sequence)` is unique; DB cross-row guard must enforce type/scope/Area coherence for all write paths.

## F10 — Submission claims TemplateSpec but digest omits it

A separate TemplateSpec could affect governed authoring while not changing Submission identity.

**Closed:** submitted `template_spec` is explicitly part of manifest/digest when applicable and must come from the same accepted OCC generation.

---

# 17. Essential vs accidental complexity / YAGNI

## Essential — retain

- stable Document identity;
- business Revision lineage;
- transactional numbering/provenance;
- immutable exact bytes;
- one mutable DRAFT authority;
- one OCC generation;
- immutable Submission attempts/digest;
- Template source provenance;
- EditorialComment submit gate;
- Periodic Review evidence;
- typed User/Area relationships;
- structural conditional uniqueness.

## Accidental — remove/defer

- separate ControlledDocument aggregate;
- Template/TemplateVersion lifecycle;
- generic numbering formula/reset engine;
- generic metadata/custom-object framework;
- business autosave/checkpoint versions;
- generic annotation/thread platform;
- universal tenant partition/RLS substrate;
- provider storage keys in business rows;
- provider/editor revision IDs as business identity;
- mandatory PDF/rendition fields in B3;
- generic BPM for Periodic Review;
- canonical default-template machinery without a consumer;
- ArtifactPackage/multi-file PLM abstraction;
- content-addressed dedupe as business authority.

Prepare the seam, not the future capability.

---

# 18. Local vs Global Maximum

**Legacy-table adaptation** is the best answer inside current structure but preserves split-brain identity and parallel lifecycles: Local Maximum.

**Generic ECM/BPM/content platform** solves hypothetical futures by adding authorities and mechanisms with no current consumer: overengineered non-maximum.

**Small mutation-law-separated kernel** removes the root cause while keeping only evidenced seams: current Global Maximum candidate.

---

# 19. Decision — candidate only

Proposed Method outcome at target-design level:

> **RESTRUCTURE NOW** — converge Controlled Information on stable Document → business DocumentRevision → one OCC WorkingContent → immutable RevisionSubmission → exact Artifact, with small typed configuration/provenance/review adjuncts and three explicit transaction contracts: creation, DRAFT CAS mutation and coherent SUBMIT freeze.

Candidate family set:

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

This is **not independently ratified**. On operator acceptance:

```text
R10-B3 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
implementation = BLOCKED
next design block = R10-B4 candidate, which must challenge B3↔B4 coherence
```

No authority file is changed by this candidate alone.

---

# 20. Reopen triggers

Reopen only on material evidence:

- true indivisible multi-file governed content requiring ArtifactPackage;
- legitimate multiple simultaneous open business Revisions per Document;
- real post-use numbering-policy changes requiring versioned policy;
- category starts driving governance rather than classification/navigation;
- real canonical default-template consumer;
- same-REV authoring must introduce/re-resolve new dictionary dependencies;
- richer Periodic Review outcome/state semantics gains a real consumer;
- responsibility must be Group-/role-based rather than one User;
- realtime coauthoring/merge invalidates one-writer+OCC;
- provider/editor cannot guarantee final coherent flush before SUBMIT;
- signature/cryptographic requirements demand a different canonical Submission representation;
- B4/B5 proves the Document serialization root or typed Artifact closure cannot preserve frozen invariants;
- implementation evidence demonstrates a DB control cannot cover all admitted write paths without disproportionate accidental complexity.

Implementation inconvenience, current schema shape or hypothetical ECM features are not reopen triggers.

---

# 21. Whole-R10 review posture

No Fable/microreview is requested for this B3 candidate under the newly accepted working mode.

During B4–F:

- this candidate remains challengeable by material counterexample;
- later blocks must explicitly test their seams against B3 rather than assume B3 is final;
- a truly exceptional trust-boundary/irreversible/cross-repository blocker may trigger early independent review;
- otherwise independent cold review occurs on the integrated Whole-R10 design after Global Coherence Review and before final ratification.
