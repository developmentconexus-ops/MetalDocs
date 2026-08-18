# R10-B6 — Audit + Interchange + Cross-owner Atomicity — Integrated Candidate

> **Status:** NON-AUTHORITATIVE CANDIDATE — OPERATOR REVIEW / R10 INTEGRATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Input working baseline:** B1/B2 promoted authority + B3/B4/B5 **ACCEPTED FOR R10 INTEGRATION / NON-FINAL**  
> **Input HEAD:** `777d560f1a77e12f7b55b0ea445876e91b046689`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Independent review:** deferred to Whole-R10 unless a material exception trigger appears.

This file is staging analysis. It does **not** independently ratify B6, promote B3–B5 to final authority, or silently rewrite the frozen R3–R9.5 ledger. It closes the relational/transactional `R10-B` design surface and explicitly records bounded refinements exposed by real pre-MetalDocs-history counterexamples.

---

# 1. Authority and evidence boundary

Authority path:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`
7. accepted non-final B3 candidate + B5-approved B3 refinements
8. accepted corrected non-final B4 candidate + B4 acceptance record
9. accepted corrected non-final B5 candidate + B5 acceptance record

Current code/schema/OpenAPI/legacy ADR/module shape remains evidence only.

External comparison evidence only:

- NIST SP 800-92 log-management guidance: <https://csrc.nist.gov/pubs/sp/800/92/final>
- RFC 8493 BagIt file-package/manifests: <https://www.rfc-editor.org/rfc/rfc8493.html>
- NARA electronic-record transfer metadata guidance: <https://www.archives.gov/records-mgmt/policy/metadata-compiled>
- PostgreSQL isolation: <https://www.postgresql.org/docs/current/transaction-iso.html>
- PostgreSQL explicit locking/deadlocks: <https://www.postgresql.org/docs/current/explicit-locking.html>

Comparison signal only:

- log/audit management is a distinct evidence infrastructure, not a substitute for domain state;
- transfer packages benefit from explicit payload manifests + checksums independent of storage layout;
- transferred records need enough metadata to identify and interpret them later;
- PostgreSQL `READ COMMITTED` may expose different snapshots to successive statements, while `REPEATABLE READ` provides one transaction snapshot;
- consistent lock acquisition order is the primary defense against application deadlocks.

MetalDocs does not adopt event sourcing, BagIt, a generic integration bus, PKI/TSA/HSM, or global stronger isolation merely because comparable systems expose those mechanisms.

---

# 2. Known / Inferred / Unknown / Deferred

## 2.1 Known — frozen/promoted/accepted inputs

### Audit

- Audit is append-only transversal timeline evidence, never canonical owner of resource state;
- critical governed mutations cannot report success without required Audit evidence in the same local DB commit;
- Audit is tamper-evident, queryable and exportable;
- Audit has its own retention regime, separate from B5 Records Governance;
- the immutable Audit portion surviving lawful user/data-subject PII erasure must be PII-minimized/non-human-readable;
- human-readable User enrichment may be erased without rewriting immutable governance evidence;
- B2 grant/revoke/offboarding evidence must remain forensically reconstructible after current grant/membership rows are deleted;
- Audit cannot become retention-clock/effectivity/approval/authorization authority.

### Interchange

- `Historical Migration`, `Governed Subject Export`, `External Repository IMPORT_COPY/PUBLISH_COPY`, and `Backup/Restore` are distinct contracts;
- Tenant Portability Export remains deferred;
- ordinary `IMPORT_COPY` follows normal target lifecycle/permissions;
- Historical Migration is privileged pre-MetalDocs-history import and never fabricates native facts;
- imported lifecycle/approval/effectivity facts are imported evidence, not fake ApprovalInstance/Decision/Release/internal-User actions;
- unknown source facts remain unknown;
- imported EFFECTIVE/SUPERSEDED/OBSOLETE require source proof;
- reliable numeric legacy revision ordinals may map directly; arbitrary labels preserve source label and map deterministically;
- next native REV must always be above the highest real imported historical ordinal;
- exact primary bytes are required to create a target DocumentRevision; missing historical bytes may preserve source-history evidence but do not create a fake Revision;
- Historical Migration has first-class plan/batch semantics, true dry-run, deterministic per-item outcomes, reconciliation and per-semantic-unit atomicity; partial batch success is valid;
- same source identity + same content reuses; conflicting content/state fails closed;
- source actors remain source snapshots/references; migration writes are attributed to Migration/System principal;
- Governed Subject Export packages Document/Evidence/Dossier independently with versioned provider-independent manifest, object/relationship/provenance metadata, canonical filenames, ContentFormats, sizes and SHA-256;
- a package claiming completeness fails closed rather than silently omit required unauthorized subjects;
- generated export package is temporary delivery output, not automatically Artifact/Evidence;
- V1 does not require signed export packages.

### Cross-owner / B1–B5

- one local PostgreSQL product-state DB/schema;
- ordinary isolation remains `READ COMMITTED` unless a narrower operation proves stronger isolation necessary;
- cross-owner frozen atomicity uses one local transaction through composable owner seams;
- no nested hidden commits and no owner imports another owner's repository for atomicity;
- provider databases/object stores never join MetalDocs atomic commit;
- required durable async intent is inserted in the same local transaction as the business fact when a later effect is necessary;
- jobs/outbox/retries/leases/projections never become semantic authority;
- B3 exact Submission, B4 Approval/Rendition/Release/Distribution and B5 Retention/Hold/Disposition laws remain integration inputs;
- B5 Artifact closure = exactly one semantic retention root per Artifact: one DocumentRevision or one Evidence.

## 2.2 Inferred candidate choices

1. Audit uses one deployment-wide append-only hash chain with one small serialized chain-head row; it is not event sourcing.
2. Audit sequence is allocated transactionally from the chain head, not from a non-transactional PostgreSQL sequence, so committed chain sequence is contiguous.
3. Audit hashes a versioned canonical event envelope using RFC 8785 JCS + SHA-256, reusing the canonicalization stack already selected by B3.
4. AuditChainHead is always the final contended semantic lock acquired by an audited transaction; no domain lock may be acquired after it.
5. Audit immutable V1 skeleton retention is `Indefinite`; no Audit-pruning/checkpoint subsystem is created without a real finite-retention requirement.
6. Audit payload does not copy free-form domain text, request bodies, passwords/tokens, IP/user-agent, provider claims or full resource snapshots. Domain reasons/comments/content remain with their owning evidence.
7. Audit uses stable internal UUID/code snapshots for actor/resource/forensic facts. Human-readable UserProfile resolution is read-time enrichment, not stored immutable Audit duplication.
8. Audit operation/fact schemas are small versioned product contracts. `facts JSONB` is bounded by the event schema and is not a generic dump field.
9. Historical imported content needed by ongoing Controlled Information behavior must be target-owner state, not an Interchange table that CI must query forever.
10. `DocumentRevision` therefore gains immutable `history_kind = NATIVE | IMPORTED` on the permanent identity skeleton.
11. Historical exact Revision content is represented by CI-owned immutable `RevisionImportedContent`, not a fabricated `RevisionSubmission`.
12. Imported source lifecycle/governance facts required to interpret a Revision are represented by CI-owned immutable `RevisionImportedGovernanceSnapshot`; Interchange owns the migration process/provenance that created it, not ongoing CI meaning.
13. Historical numeric ordinals that are real but lack sufficient bytes for a DocumentRevision are reserved by CI-owned `RevisionOrdinalReservation`; native allocation is above both Revision rows and reservations.
14. Template origin can reference either native Submission content or imported Revision content through a source-content-kind snapshot without reintroducing strong retention payload FKs.
15. Imported current EFFECTIVE Revision with unknown trustworthy source effectivity does not invent `effective_at`; `adopted_as_current_at` is an adoption fact. For Periodic Review, lack of a trustworthy prior anchor makes the Revision immediately due (not invalid) until an explicit review establishes a new anchor.
16. Imported Revision RetentionBinding is created in the migration semantic unit even though no native first Submission exists. Imported retention anchor derives from trustworthy imported lifecycle evidence; unknown anchor never silently becomes disposition-eligible.
17. Historical Evidence likewise requires an imported capture shape instead of fake native `captured_by`/`captured_at`; B5 is bounded-refined with immutable `Evidence.history_kind` + `EvidenceImportedCapture`.
18. Interchange source idempotency uses a persistent typed `HistoricalSourceBinding`; it is a closed target union, not `target_type/target_id` generic polymorphism.
19. External repository connection semantics use a stable MetalDocs `InterchangeConnection` identity; credentials/endpoints remain deployment/provider mechanism. Rebinding one connection to a different logical external repository is forbidden; create a new connection instead.
20. Governed export semantic snapshot is created inside one short PostgreSQL `REPEATABLE READ` transaction; package materialization happens after commit through R10-D mechanisms.
21. Governed export manifest uses JCS + SHA-256 and exact Artifact hashes; MetalDocs does not adopt BagIt V1 absent a consumer requiring that external packaging contract.
22. Export does not create hidden LegalHold/retention. If source bytes are lawfully disposed before package build finishes, build fails visibly/reconciles rather than silently changing retention semantics.
23. PUBLISH_COPY request snapshots exact source identity/hash/format but does not retain a strong Artifact FK forever. If source is disposed before external effect succeeds, transfer fails visibly/reconciles.
24. IMPORT_COPY calls ordinary target-owner seams after bytes are staged/validated; it never bypasses normal target governance because an external connector fetched the bytes.
25. B6 closes an explicit same-local-commit matrix and a global partial lock order. Operations may start at a later class, but may never backtrack to an earlier class.

## 2.3 Unknown — kept unknown

- a future contractual finite Audit retention period and chain-pruning/checkpoint scheme;
- external cryptographic anchoring/WORM/TSA/HSM for stronger DBA-level non-repudiation;
- a future BagIt or other standardized export packaging requirement;
- a future explicitly partial Governed Subject Export contract;
- provider-specific External Repository receipt metadata beyond bounded stable IDs/version facts;
- future non-MetalDocs consumer requiring a public integration event bus;
- imported historical source states not faithfully expressible in current target lifecycle without a new explicit semantic mapping;
- exact R10-F legacy/cutover mechanics for historical state maps.

These are reopen triggers, not V1 defaults.

## 2.4 Deferred

```text
physical storage / malware / ObjectLock / restore / byte-delete proof → R10-C
worker schemas / retries / leases / timers / DLQ / projections       → R10-D
HTTP/API/frontend/export/download/admin journeys                     → R10-E
legacy cutover / bootstrap recovery / final migration execution map   → R10-F
```

---

# 3. Root Cause

B6 prevents **transversal truth collapse**.

Without an explicit boundary, three cross-cutting concerns tend to become shadow authorities:

```text
Audit log
integration/migration state
async/outbox state
```

Typical defects:

- business state succeeds while mandatory Audit is lost;
- Audit payload is later queried as the resource authority;
- Audit becomes event sourcing accidentally;
- outbox/job state is treated as proof an external effect occurred;
- historical source approval/effectivity is rewritten as native MetalDocs actions;
- migration time is substituted for unknown historical time;
- export reads multiple `READ COMMITTED` snapshots and emits a package representing no real database state;
- generic integration registries duplicate target-owner identity;
- a reverse lock edge through Audit/composition creates deadlocks;
- one owner commits locally and a second owner is “repaired later” even though the invariant required atomicity.

---

# 4. Target invariant

> **Every business/system fact remains owned by exactly one semantic owner. Required Audit evidence is appended in the same local transaction as the governed mutation but never becomes the mutation's authority. Historical Migration preserves external truth as explicit imported target-owner state/evidence without fabricating native actions. External imports/publishes/exports preserve exact provenance and content identity without provider-storage coupling. Every B1–B5 invariant that spans local owners is committed once through composition under one compatible lock order, while external effects are represented only by durable intent + later explicit receipt.**

```text
AuditEvent != domain state
AuditEvent != outbox/event bus
outbox intent != effect receipt
Migration execution != native User action
imported governance != ApprovalDecision/ReleaseRecord
export manifest != provider key listing
InterchangeConnection != credential store
REPEATABLE READ export snapshot != global DB isolation policy
composition != semantic owner
```

---

# 5. Credible alternatives / Global Maximum

## A — central event sourcing / Audit as global history authority

Reject. Creates a second authority for every domain, couples all recovery/migration to one log and dramatically increases accidental complexity.

## B — async best-effort Audit + generic integration/outbox engine

Reject. Business facts can succeed without mandatory evidence; provenance/effect intent/process truth become one ambiguous transversal subsystem.

## C — DB triggers audit every CRUD write

Reject. Captures persistence mutations, not business meaning; cannot correctly express exact Submission/Approval/Release/hold semantics and makes the database infer domain authority.

## D — small semantic Audit + specialized Interchange contracts + explicit transaction/lock matrix

**Recommended Global Maximum.** Smallest structure that preserves forensic evidence, imported truth and cross-owner atomicity without introducing a generic event/integration platform.

---

# 6. Audit semantic model

## 6.1 AuditChainHead

One row per deployment/database:

```text
AuditChainHead
  singleton_key SMALLINT PRIMARY KEY CHECK singleton_key = 1
  last_sequence BIGINT NOT NULL CHECK last_sequence >= 0
  last_hash BYTEA NOT NULL CHECK octet_length(last_hash)=32
```

Genesis:

```text
last_sequence = 0
last_hash = 32 zero bytes
```

Mutation law: mutable only inside Audit append; serving application cannot reset/reseed it.

## 6.2 AuditEvent

```text
AuditEvent
  id UUID PRIMARY KEY
  sequence_no BIGINT NOT NULL UNIQUE CHECK sequence_no >= 1
  occurred_at TIMESTAMPTZ NOT NULL

  actor_kind TEXT NOT NULL CHECK USER|SYSTEM
  actor_user_id UUID NULL FK User(id) RESTRICT
  system_actor_code TEXT NULL

  operation_code TEXT NOT NULL
  resource_kind TEXT NOT NULL
  resource_id UUID NOT NULL

  facts_schema TEXT NOT NULL
  facts JSONB NOT NULL CHECK jsonb_typeof(facts)='object'

  correlation_id UUID NULL

  prev_hash BYTEA NOT NULL CHECK octet_length(prev_hash)=32
  event_hash BYTEA NOT NULL CHECK octet_length(event_hash)=32
```

Actor union:

```text
USER   → actor_user_id only
SYSTEM → system_actor_code only
```

`resource_kind/resource_id` is intentionally non-FK forensic attribution. B1 permits generic attribution in Audit because Audit does not own or constrain resource lifecycle; the target row may later be disposed/deleted.

## 6.3 Canonical hash

Versioned canonical payload:

```text
payload = {
  id,
  sequence_no,
  occurred_at,
  actor_kind,
  actor_user_id?,
  system_actor_code?,
  operation_code,
  resource_kind,
  resource_id,
  facts_schema,
  facts,
  correlation_id?
}

canonical = RFC8785_JCS(payload)

event_hash = SHA256(
  prev_hash
  || 0x00
  || UTF8("metaldocs.audit-event.v1")
  || 0x00
  || canonical
)
```

`prev_hash` is stored separately and must equal the prior committed event hash. Cross-runtime golden vectors are required before implementation.

## 6.4 Append transaction law

For an audited mutation, all domain/records/effect-intent work is completed first. Audit append is last:

```text
lock AuditChainHead FOR UPDATE
read last_sequence/last_hash
for each required event in stable caller-defined order:
  allocate sequence = previous + 1
  compute canonical event/hash
  insert AuditEvent
  advance local previous hash/sequence
update AuditChainHead once to final sequence/hash
COMMIT
```

Rollback rolls back events + chain-head advance, so the committed Audit sequence has no transaction-created gaps.

**No domain/owner lock may be acquired after AuditChainHead.**

## 6.5 Serving trust / tamper evidence

Ordinary serving trust:

```text
AuditEvent INSERT/SELECT as needed
AuditEvent UPDATE/DELETE forbidden
AuditChainHead UPDATE only through narrow Audit append seam
```

Hash-chain verification detects mutation/deletion/reordering within the trust domain. V1 does **not** claim cryptographic non-repudiation against an attacker capable of rewriting the whole DB + application trust domain. External anchoring/signing remains a reopen trigger.

---

# 7. Audit privacy classification

Immutable V1 event skeleton stores only the minimum needed for forensic reconstruction.

| Field class | V1 treatment |
|---|---|
| event UUID / sequence / trusted time / hashes / schemas | non-PII technical/governance evidence |
| operation/resource kind | non-human-readable product code |
| actor User UUID / subject User UUID in bounded facts | PII-minimized stable pseudonymous identity skeleton accepted to survive UserProfile erasure |
| role code / scope kind / Area/Tenant UUID / assignment id / outcome codes / digests | bounded non-human-readable governance facts |
| User display name/email/username/profile | **not copied into immutable AuditEvent** |
| source actor free text from historical systems | **not copied into AuditEvent**; belongs imported Revision/Evidence retention payload |
| approval/comment/return/disposition reason text | **not copied**; authoritative domain evidence is referenced by UUID |
| JWT/token/password/fresh-auth raw provider data | forbidden |
| IP/user-agent/request body/arbitrary HTTP headers | not semantic Audit V1; telemetry/security logging may own them separately |

Read-time UI may resolve a surviving `actor_user_id` through current UserProfile. If enrichment is erased/unavailable, display a stable opaque User identifier rather than rewriting Audit.

No `AuditEventEnrichment` table is introduced V1.

---

# 8. Audit retention / query boundary

V1 immutable Audit skeleton retention:

```text
Indefinite
```

Reason: frozen authority requires a separate Audit regime but establishes no finite term. Designing chain segmentation/pruning now would add unsupported correctness machinery. A real finite-retention/legal deletion requirement reopens this decision.

Audit query/export reads AuditEvent; it never reconstructs canonical business state from “latest event”. Domain reads always return to the owning authority.

Audit export/package mechanics belong R10-D/E where material; exporting Audit does not convert it into B5 Evidence or a Governed Subject Export unless a future explicit requirement says so.

---

# 9. B3/B5 bounded refinements revealed by Historical Migration

These are material pre-MetalDocs-history counterexamples, not implementation convenience.

## 9.1 B3-R4 — reserve real historical ordinals without fake Revision

Counterexample:

```text
legacy REV007 has exact bytes → import real REV007
legacy REV008 ordinal/proof exists but bytes are lost
→ frozen rule forbids fake DocumentRevision REV008
→ B3 max(DocumentRevision.revision_no)+1 would create native REV008
```

That reuses a real historical ordinal.

Target:

```text
RevisionOrdinalReservation
  id UUID PRIMARY KEY
  document_id UUID NOT NULL FK Document(id) RESTRICT
  revision_no INTEGER NOT NULL CHECK revision_no >= 1
  reserved_at TIMESTAMPTZ NOT NULL
  UNIQUE(document_id,revision_no)
```

DB guard enforces no `(document_id,revision_no)` collision across `DocumentRevision` and `RevisionOrdinalReservation`.

Native allocation becomes:

```text
max(
  all DocumentRevision.revision_no,
  all RevisionOrdinalReservation.revision_no
) + 1
```

Reservation is a permanent minimal ordinal skeleton, not a fake lifecycle object.

## 9.2 B3-R5 — imported exact Revision content is not RevisionSubmission

`RevisionSubmission` means a native submit attempt. Historical Migration must not fabricate it.

Permanent Revision skeleton gains:

```text
DocumentRevision.history_kind TEXT NOT NULL CHECK NATIVE|IMPORTED
```

Immutable.

Imported exact content:

```text
RevisionImportedContent
  revision_id UUID PRIMARY KEY FK DocumentRevision(id) RESTRICT
  primary_artifact_id UUID NOT NULL FK Artifact(id) RESTRICT
  manifest_schema TEXT NOT NULL
  manifest_payload JSONB NOT NULL CHECK jsonb_typeof(...)='object'
  content_digest BYTEA NOT NULL CHECK octet_length(content_digest)=32
  adopted_at TIMESTAMPTZ NOT NULL
```

Canonical digest:

```text
SHA256(
  UTF8(manifest_schema)
  || 0x00
  || RFC8785_JCS(manifest_payload)
)
```

Manifest contains exact primary Artifact hash/size/format/media type + known governed metadata/structured provenance. Unknown historical fields remain absent/explicitly unknown per schema; migration time never substitutes for a missing historical fact.

`RevisionImportedContent` is immutable retention payload under the same DocumentRevision Artifact root.

## 9.3 B3-R6 — imported governance snapshot belongs CI, not Interchange runtime dependency

CI must be able to interpret an imported Revision after migration machinery/process history is no longer in working context.

```text
RevisionImportedGovernanceSnapshot
  revision_id UUID PRIMARY KEY FK DocumentRevision(id) RESTRICT
  source_revision_label TEXT NULL
  source_state TEXT NOT NULL
  source_effective_at TIMESTAMPTZ NULL
  source_superseded_at TIMESTAMPTZ NULL
  source_obsoleted_at TIMESTAMPTZ NULL
  source_cancelled_at TIMESTAMPTZ NULL
  adopted_as_current_at TIMESTAMPTZ NULL
  source_actor_snapshot JSONB NULL CHECK NULL OR jsonb_typeof(...)='object'
  source_governance_snapshot JSONB NOT NULL CHECK jsonb_typeof(...)='object'
```

Bounded/versioned snapshot schema is required; it is not arbitrary metadata.

Rules:

- target `DocumentRevision.state` remains CI authority;
- imported state must be supported by explicit source proof/mapping;
- no ApprovalInstance/Decision/ReleaseRecord is synthesized;
- source actor names/references are source provenance, never internal User actors;
- if target state is imported EFFECTIVE and trustworthy source effectivity is unknown, keep `source_effective_at NULL` and set `adopted_as_current_at` to the real MetalDocs adoption instant;
- imported snapshot belongs to the DocumentRevision retention unit and may be disposed with that unit while permanent Revision identity/history_kind/ordinal survives.

## 9.4 B3-R1 extension — Template origin supports imported current-effective content

Accepted B5 shape is extended:

```text
DocumentOrigin
  derived_document_id
  source_template_revision_id
  source_content_kind TEXT CHECK NATIVE_SUBMISSION|IMPORTED_REVISION_CONTENT
  source_submission_id_snapshot UUID NULL
  source_content_digest BYTEA NOT NULL CHECK octet_length(...)=32
  source_artifact_sha256 BYTEA NOT NULL CHECK octet_length(...)=32
  source_content_format TEXT NOT NULL
  created_at
```

`NATIVE_SUBMISSION` requires source Submission-id snapshot; imported kind requires it NULL. Creation serializes against current Template effectivity and proves the selected exact source content authority for that source Revision.

## 9.5 Imported Periodic Review baseline

For current imported EFFECTIVE Revision:

```text
trusted source effective/review anchor known
→ normal due calculation from that trusted anchor

trusted source anchor unknown
→ preserve unknown historical effective time
→ adopted_as_current_at is not re-labeled as historical effective_at
→ Revision is immediately DUE for Periodic Review
```

Immediate due does not invalidate EFFECTIVE content. Once a native PeriodicReviewRecord occurs, later due scheduling uses the normal current-Revision review anchor.

## 9.6 B5 refinement — imported Revision RetentionBinding

`DocumentRevision` history kind `IMPORTED` with exact `RevisionImportedContent` creates its RetentionBinding in the same migration semantic unit; no first native Submission is required.

Imported retention anchor:

- while imported Revision is current EFFECTIVE → no running clock;
- when later superseded natively → B4 winning Release time is the new anchor;
- already-historical imported states may use only trustworthy imported lifecycle anchor from `RevisionImportedGovernanceSnapshot`;
- unknown historical anchor → no silent disposition eligibility.

---

# 10. B5-R1 — historical Evidence capture without fake native actor/time

B5 native `EvidenceCapture` requires native `captured_by_user_id` and `captured_at`. Historical Migration cannot invent them.

Permanent Evidence skeleton gains:

```text
Evidence.history_kind TEXT NOT NULL CHECK NATIVE|IMPORTED
```

Native CAPTURE remains unchanged.

Imported capture:

```text
EvidenceImportedCapture
  evidence_id UUID PRIMARY KEY FK Evidence(id) RESTRICT
  primary_artifact_id UUID NOT NULL FK Artifact(id) RESTRICT
  reference_value TEXT NULL
  governed_metadata JSONB NOT NULL CHECK jsonb_typeof(...)='object'
  occurred_at TIMESTAMPTZ NULL
  source_captured_at TIMESTAMPTZ NULL
  source_actor_snapshot JSONB NULL CHECK NULL OR jsonb_typeof(...)='object'
  source_provenance JSONB NOT NULL CHECK jsonb_typeof(...)='object'
  original_filename TEXT NULL
  adopted_at TIMESTAMPTZ NOT NULL
```

Exactly one native `EvidenceCapture` **or** `EvidenceImportedCapture` exists for CAPTURED Evidence according to `history_kind`.

Migration transaction creates:

```text
Evidence identity/history_kind IMPORTED
+ exact Artifact relation
+ EvidenceImportedCapture
+ immutable primary Dossier/name/sequence state
+ RetentionBinding
+ applicable active LegalHold materialization
```

No fake internal capture User/time.

Retention anchor:

```text
CAPTURED_AT → source_captured_at when trustworthy, else UNKNOWN
OCCURRED_AT → occurred_at when trustworthy, else UNKNOWN
```

Unknown anchor never silently becomes disposition-eligible. `adopted_at` is provenance, not substitute capture/occurred time.

---

# 11. Historical Migration semantic model

## 11.1 HistoricalMigrationSource

Stable source namespace:

```text
HistoricalMigrationSource
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  name TEXT NOT NULL
  source_kind TEXT NOT NULL
  status TEXT NOT NULL CHECK ACTIVE|INACTIVE
  created_at TIMESTAMPTZ NOT NULL
```

No credentials/provider endpoint secrets.

## 11.2 HistoricalMigrationPlan

```text
HistoricalMigrationPlan
  id UUID PRIMARY KEY
  source_id UUID NOT NULL FK HistoricalMigrationSource(id) RESTRICT
  mode TEXT NOT NULL CHECK CURRENT_STATE|FULL_HISTORY
  plan_schema TEXT NOT NULL
  plan_digest BYTEA NOT NULL CHECK octet_length(plan_digest)=32
  created_by_user_id UUID NOT NULL FK User(id) RESTRICT
  created_at TIMESTAMPTZ NOT NULL
```

Immutable after finalization. Plan digest covers a canonical ordered plan-item descriptor set.

## 11.3 HistoricalMigrationPlanItem

```text
HistoricalMigrationPlanItem
  id UUID PRIMARY KEY
  plan_id UUID NOT NULL FK HistoricalMigrationPlan(id) RESTRICT
  item_order BIGINT NOT NULL CHECK item_order >= 1
  source_entity_kind TEXT NOT NULL
  source_entity_id TEXT NOT NULL
  source_fingerprint BYTEA NOT NULL CHECK octet_length(source_fingerprint)=32
  target_kind TEXT NOT NULL CHECK DOCUMENT|DOCUMENT_REVISION|REVISION_ORDINAL|EVIDENCE|DOSSIER
  target_hint JSONB NOT NULL CHECK jsonb_typeof(...)='object'

  UNIQUE(plan_id,item_order)
  UNIQUE(plan_id,source_entity_kind,source_entity_id)
```

`target_hint` is bounded mapping input only; no credentials/file bytes/free-form governance dumps.

## 11.4 HistoricalMigrationExecution

```text
HistoricalMigrationExecution
  id UUID PRIMARY KEY
  plan_id UUID NOT NULL FK HistoricalMigrationPlan(id) RESTRICT
  mode TEXT NOT NULL CHECK DRY_RUN|APPLY
  requested_by_user_id UUID NOT NULL FK User(id) RESTRICT
  requested_at TIMESTAMPTZ NOT NULL
  completed_at TIMESTAMPTZ NULL
  status TEXT NOT NULL CHECK RUNNING|COMPLETED|PARTIAL|FAILED
```

Process state only; target truth never inferred from execution status.

## 11.5 HistoricalMigrationItemOutcome

One semantic final outcome per execution/item:

```text
HistoricalMigrationItemOutcome
  id UUID PRIMARY KEY
  execution_id UUID NOT NULL FK HistoricalMigrationExecution(id) RESTRICT
  plan_item_id UUID NOT NULL FK HistoricalMigrationPlanItem(id) RESTRICT
  outcome TEXT NOT NULL CHECK
    WOULD_CREATE|WOULD_REUSE|CONFLICT|INVALID|
    CREATED|REUSED|FAILED
  reason_code TEXT NULL
  target_id_snapshot UUID NULL
  recorded_at TIMESTAMPTZ NOT NULL

  UNIQUE(execution_id,plan_item_id)
```

Transient attempts/retries/logs remain R10-D.

## 11.6 HistoricalSourceBinding — cross-plan idempotency

```text
HistoricalSourceBinding
  id UUID PRIMARY KEY
  source_id UUID NOT NULL FK HistoricalMigrationSource(id) RESTRICT
  source_entity_kind TEXT NOT NULL
  source_entity_id TEXT NOT NULL
  source_fingerprint BYTEA NOT NULL CHECK octet_length(source_fingerprint)=32

  document_id UUID NULL FK Document(id) RESTRICT
  document_revision_id UUID NULL FK DocumentRevision(id) RESTRICT
  revision_ordinal_reservation_id UUID NULL FK RevisionOrdinalReservation(id) RESTRICT
  evidence_id UUID NULL FK Evidence(id) RESTRICT
  dossier_id UUID NULL FK Dossier(id) RESTRICT

  bound_at TIMESTAMPTZ NOT NULL

  UNIQUE(source_id,source_entity_kind,source_entity_id)
```

Exactly one target FK.

Same source identity:

```text
same source_fingerprint + compatible target truth → REUSE
changed/conflicting fingerprint/state             → CONFLICT / fail closed
```

No heuristic merge.

---

# 12. Historical Migration dry-run/apply law

Dry-run:

- validates source fingerprint/format/mapping/uniqueness/permissions/retention preconditions using the same target validation contracts as APPLY;
- creates no target Document/Revision/Evidence/Dossier/Artifact/RetentionBinding/hold/domain Audit facts;
- records deterministic `WOULD_* | CONFLICT | INVALID` outcomes only.

APPLY:

1. re-read/verify exact source fingerprint against finalized plan item;
2. if existing `HistoricalSourceBinding` matches same source fingerprint and target is coherent → REUSE;
3. if binding/fingerprint/target conflicts → fail item closed;
4. prepare/stage exact bytes outside semantic DB transaction where needed;
5. open one semantic target transaction;
6. call the target owner's privileged migration seam;
7. create target-owned imported content/governance/capture facts + RetentionBinding/holds where required;
8. confirm Artifact + one semantic retention root in same transaction;
9. create/update `HistoricalSourceBinding` for the exact source identity;
10. insert APPLY item outcome;
11. append required Audit event(s) last;
12. commit.

Partial batch success is valid. Each item is independently reconcilable.

No whole-batch rollback promise.

---

# 13. Governed Subject Export

## 13.1 Semantic request

Closed subject union:

```text
GovernedSubjectExport
  id UUID PRIMARY KEY
  document_id UUID NULL FK Document(id) RESTRICT
  evidence_id UUID NULL FK Evidence(id) RESTRICT
  dossier_id UUID NULL FK Dossier(id) RESTRICT
  requested_by_user_id UUID NOT NULL FK User(id) RESTRICT
  requested_at TIMESTAMPTZ NOT NULL
```

Exactly one subject.

## 13.2 Stable complete snapshot — narrow isolation exception

The snapshot transaction uses PostgreSQL `REPEATABLE READ` **only for this operation**.

Within one short transaction:

1. resolve the requested subject closure under one stable database snapshot;
2. apply canonical Authorization to every required object/relation;
3. if any required subject is unauthorized/unavailable, fail the complete export request before success;
4. enumerate exact retained/current content identities + hashes without provider locations;
5. build canonical versioned export manifest;
6. insert `GovernedSubjectExport` + immutable snapshot;
7. insert package-build durable intent;
8. append Audit event last;
9. commit.

No physical bytes are copied while the DB transaction is open.

PostgreSQL serialization/retry errors are fail-visible/retryable; they do not justify global isolation change.

## 13.3 GovernedExportSnapshot

```text
GovernedExportSnapshot
  export_id UUID PRIMARY KEY FK GovernedSubjectExport(id) RESTRICT
  manifest_schema TEXT NOT NULL
  manifest_payload JSONB NOT NULL CHECK jsonb_typeof(...)='object'
  manifest_digest BYTEA NOT NULL CHECK octet_length(manifest_digest)=32
  snapshotted_at TIMESTAMPTZ NOT NULL
```

Manifest V1 contains at minimum:

```text
package schema/version
root subject identity
included object identities/types
relationships/context
provenance
canonical filenames
ContentFormat/media type
size_bytes
sha256
content/governance identity digests where applicable
completeness = COMPLETE
```

No provider bucket/key/URL/version identifier.

Digest = `SHA256(UTF8(manifest_schema) || 0x00 || RFC8785_JCS(manifest_payload))`.

No partial export V1.

## 13.4 Package result

Actual package is temporary delivery output and is not a semantic Artifact/Evidence.

Successful build records only provider-independent receipt:

```text
GovernedExportPackageReceipt
  export_id UUID PRIMARY KEY FK GovernedSubjectExport(id) RESTRICT
  package_sha256 BYTEA NOT NULL CHECK octet_length(package_sha256)=32
  size_bytes BIGINT NOT NULL CHECK size_bytes >= 0
  completed_at TIMESTAMPTZ NOT NULL
```

Temporary package storage/location/expiry/download token is R10-D/E mechanism state.

Package build must verify every emitted file against the snapshot hash before writing receipt.

Export does not create/release LegalHold, change retention, or count as Distribution acknowledgement.

---

# 14. InterchangeConnection / ExternalReference closure

```text
InterchangeConnection
  id UUID PRIMARY KEY
  code TEXT NOT NULL UNIQUE
  name TEXT NOT NULL
  connector_kind TEXT NOT NULL
  status TEXT NOT NULL CHECK ACTIVE|INACTIVE
  created_at TIMESTAMPTZ NOT NULL
```

Code immutable. This row is the stable MetalDocs identity of one logical external repository/source connection.

Credentials, tokens, endpoints and secret material are provider/deployment configuration keyed by connection id and never stored as business identity.

If configuration points to a materially different logical repository, create a new `InterchangeConnection`; do not reinterpret historical external IDs under the old connection.

B5 `DossierExternalReference.connection_id` becomes a typed FK to this family. `(connection_id, entity_kind, external_id)` remains unique for Dossier mapping.

---

# 15. External Repository COPY process

## 15.1 RepositoryTransfer

```text
RepositoryTransfer
  id UUID PRIMARY KEY
  connection_id UUID NOT NULL FK InterchangeConnection(id) RESTRICT
  direction TEXT NOT NULL CHECK IMPORT_COPY|PUBLISH_COPY
  requested_by_user_id UUID NOT NULL FK User(id) RESTRICT
  requested_at TIMESTAMPTZ NOT NULL

  external_entity_kind TEXT NULL
  external_entity_id TEXT NULL

  source_document_revision_id UUID NULL FK DocumentRevision(id) RESTRICT
  source_evidence_id UUID NULL FK Evidence(id) RESTRICT
  source_artifact_id_snapshot UUID NULL
  source_sha256 BYTEA NULL CHECK NULL OR octet_length(source_sha256)=32
  source_content_format TEXT NULL

  intended_target_kind TEXT NULL CHECK NULL OR DOCUMENT|EVIDENCE
```

Shape:

- `PUBLISH_COPY` pins one internal source root + exact Artifact id/hash/format snapshot; no strong Artifact FK;
- `IMPORT_COPY` pins external entity identity + intended target kind and does not pre-create target business state.

This row is process truth, not proof the external effect occurred.

## 15.2 Transfer effect

Same request transaction:

```text
RepositoryTransfer
+ Audit request event
+ durable external-effect intent
COMMIT
```

R10-D worker performs provider interaction.

### PUBLISH_COPY

Before external write, worker proves the pinned Artifact still exists with exact expected hash/format. If not, fail visibly; do not resurrect disposed content or invent retention.

### IMPORT_COPY

Worker retrieves/stages/validates bytes, then calls ordinary target `Document`/`Evidence` creation/capture seams under the requesting actor's authorized operation semantics. The connector does not bypass target governance.

## 15.3 RepositoryTransferReceipt

Only successful external effect creates immutable receipt:

```text
RepositoryTransferReceipt
  transfer_id UUID PRIMARY KEY FK RepositoryTransfer(id) RESTRICT
  external_entity_kind TEXT NOT NULL
  external_entity_id TEXT NOT NULL
  external_version_snapshot TEXT NULL
  target_document_id UUID NULL FK Document(id) RESTRICT
  target_evidence_id UUID NULL FK Evidence(id) RESTRICT
  completed_at TIMESTAMPTZ NOT NULL
```

For PUBLISH_COPY, receipt identifies external created/updated copy. For IMPORT_COPY, receipt identifies both external source and the exact resulting MetalDocs target.

No receipt = no semantic claim that the external effect succeeded.

Provider URLs remain mechanism/display data, not receipt identity.

---

# 16. Required Audit event classes

Audit is not “log every mutation”. Mandatory V1 same-commit semantic event classes:

### Authentication / Organization / Authorization

- Tenant settings mutation;
- Area create/rename/retire/re-enable;
- User create/offboard/re-enable;
- UserProfile governed mutation/erasure;
- ProviderSubjectBinding acceptance/replacement/disable/erasure where admitted;
- administrative Session revocation / offboarding;
- Group create/rename/delete;
- GroupMembership add/remove;
- RoleAssignment grant/revoke, including bounded forensic subject/role/scope facts.

### Controlled Information / Approval / Distribution

- Document creation;
- responsibility change;
- Revision creation where material to governance timeline;
- SUBMIT;
- Revision cancel/obsolete;
- PeriodicReviewRecord creation;
- ApprovalStepDecision;
- approval cancel/withdraw/reassign;
- Rendition semantic success;
- Release;
- Distribution acknowledgement.

### Documentary Context / Evidence / Records

- Dossier create/archive/re-enable;
- Dossier↔Document link/unlink;
- Evidence CAPTURE/VOID;
- captured Evidence secondary-Dossier link/unlink because it can affect live-hold scope;
- RetentionExtension;
- LegalHold activate/release;
- DispositionFence;
- completed DispositionRecord.

### Interchange

- Historical Migration APPLY item CREATE/REUSE/CONFLICT where a semantic target/binding decision occurs;
- Governed Subject Export snapshot accepted;
- external COPY request;
- external COPY successful receipt.

Not mandatory semantic Audit V1 by default:

```text
WorkingContent autosave/edit
EditorSession heartbeat/lease
EvidenceDraft edits
EditorialComment / SubmissionFeedback ordinary collaboration
search/view/download
notification delivery/read
worker retry/lease/DLQ
projection rebuild
provider health/telemetry
```

Those facts have their own authority or are operational telemetry. A later compliance requirement may promote a specific event class deliberately.

---

# 17. Same-local-commit matrix

`Audit` below means required event(s) in the same local transaction, appended last. `Intent` means durable mechanism intent in the same transaction only when a later external/async effect is required.

| Operation | Atomic semantic set |
|---|---|
| User offboarding | User disable + local Session revoke + memberships/direct grants revoke + required binding state + Audit + provider-disable Intent |
| Document + REV001 create | numbering allocation + Document + Revision + RevisionDictionarySnapshot + WorkingContent + optional Template origin/spec + Audit |
| First SUBMIT | WorkingContent generation consume + RevisionSubmission + B4 ApprovalRequirement/ReleasePlan + Approval init where required + B5 RetentionBinding + applicable Hold materialization + Audit + downstream evaluation Intent if needed |
| Same-REV resubmit | new RevisionSubmission + new B4 per-Submission requirements/Approval init + Audit + evaluation Intent; existing RetentionBinding unchanged |
| Approval RETURN/CANCEL/WITHDRAW | CI Revision return-to-DRAFT generation law + Approval terminal state + Audit |
| Approval ACCEPT decision | immutable decision + Step/Instance transition + Audit + release-evaluation Intent if final gate can now progress |
| Rendition semantic success | Artifact confirmation + Rendition owner relation + Audit + release-evaluation Intent |
| Release | candidate EFFECTIVE + predecessor SUPERSEDED + ReleaseRecord + DistributionObligations + Audit + notification/search Intents |
| Periodic Review | exact current Revision review record + Audit |
| Evidence CAPTURE | EvidenceCapture/ImportedCapture + Artifact root + Evidence state/name/primary dossier + RetentionBinding + applicable Holds + Audit |
| Dossier link enters active hold | link + necessary LegalHoldSubject rows + Audit |
| captured Evidence secondary link enters hold | link + necessary LegalHoldSubject rows + Audit |
| LegalHold activation | hold + complete current materialization + Audit |
| RetentionExtension | extension + Audit |
| DispositionFence | fence + Audit + physical-delete Intent |
| Disposition completion | retained payload cleanup + Artifact semantic cleanup + DispositionRecord + Audit |
| Historical Migration APPLY item | target-owner imported state/content/provenance + source binding + RetentionBinding/holds where required + item outcome + Audit |
| Governed Export snapshot | export request + immutable complete manifest snapshot + package-build Intent + Audit under one REPEATABLE READ transaction |
| PUBLISH/IMPORT COPY request | RepositoryTransfer + external-effect Intent + Audit |
| COPY success | RepositoryTransferReceipt + target relation if import + Audit |

No provider call/object-store transfer is inside these DB atomic boundaries.

---

# 18. Cross-owner composition law

`internal/composition` may orchestrate one local transaction but owns no durable business meaning.

Each semantic owner exposes transaction-aware application seams that:

- accept the shared transaction/context;
- do not commit/rollback independently;
- do not open hidden nested transactions for required semantic writes;
- do not import another owner's persistence adapter/repository;
- return domain-specific success/failure to composition;
- make all mandatory local invariants visible before the final Audit append.

Audit exposes a transaction-aware append seam only. R10-D exposes durable-intent insert seam only. Neither may call domain owners back after their lock class is reached.

---

# 19. Final B1–B6 lock-order law

The target is a **partial order**, not a generic lock manager.

## 19.1 Main governed-content branch

When applicable:

```text
B2 eligibility roots required by the operation
  User / Binding / Area according to promoted B2 sub-order

→ current configuration/allocation roots
  DocumentType / EvidenceType / Approval config / NumberSeries / EvidenceSequence

→ business subject roots
  existing Document(s) ordered UUID
  → existing Evidence(s) ordered UUID
  → Dossier(s) ordered UUID

→ business child/execution rows
  DocumentRevision / WorkingContent / ApprovalInstance+Step / EvidenceDraft+Capture

→ RetentionBinding(s) ordered UUID

→ Artifact existing-row coordination only where genuinely required

→ durable-intent inserts

→ AuditChainHead ALWAYS LAST
```

Classes may be skipped. A transaction may start at a later class if it will never acquire an earlier class.

## 19.2 Group live-configuration/snapshot branch

B2 Group deletion remains isolated:

```text
Group
→ GroupMembership / Group RoleAssignment children
```

and never acquires Document/Evidence/Dossier/Audit-domain roots afterwards except AuditChainHead last.

B4 Release/Approval participant resolution may lock/read live Group dependencies **after** its Document/Approval root because the reverse Group-maintenance path is forbidden from later acquiring those governed-content roots. This creates a DAG edge, not a cycle.

## 19.3 Approval return rule

A path that can return/cancel/withdraw a Revision must acquire the Document/Revision root **before** Approval execution rows.

Pure ACCEPT on a non-terminal intermediate Step may remain Approval-local and then append Audit; it must not subsequently acquire Document.

## 19.4 Dossier/Records rule

For relations/holds involving an existing subject:

```text
Document or Evidence subject root
→ relevant Dossier roots UUID order
→ RetentionBinding
→ AuditChainHead
```

Dossier-wide Hold activation may start at Dossier and then acquire RetentionBindings in deterministic order; it must not backtrack to Document/Evidence roots after those bindings are locked.

## 19.5 Audit rule

```text
NO owner/domain lock after AuditChainHead is acquired.
```

This is structural and must be testable in the implementation spec.

---

# 20. Governed Export isolation law

Global product isolation remains `READ COMMITTED`.

Only the short semantic export-snapshot transaction uses `REPEATABLE READ`, because a complete manifest spanning many relationships must reflect one actual database snapshot without long row-lock fan-out.

The transaction ends after:

```text
closure resolved
+ AuthZ checked
+ immutable manifest persisted
+ build Intent persisted
+ Audit appended
```

Package bytes are assembled after commit.

No other B6 operation gains stronger isolation by default.

---

# 21. Persistence / mutation classes

| Family | Semantic class | Mutation law |
|---|---|---|
| AuditChainHead | supporting semantic mechanism for Audit integrity | narrowly mutable monotonic head |
| AuditEvent | semantic Audit authority | immutable append-only |
| RevisionOrdinalReservation | CI semantic identity skeleton | immutable/permanent |
| DocumentRevision.history_kind | CI semantic identity | immutable |
| RevisionImportedContent | CI semantic imported content authority | immutable retention payload |
| RevisionImportedGovernanceSnapshot | CI semantic imported governance snapshot | immutable retention payload |
| Evidence.history_kind | Documentary Context semantic identity | immutable |
| EvidenceImportedCapture | Documentary Context semantic imported capture authority | immutable retention payload |
| HistoricalMigrationSource | Interchange semantic configuration | code immutable; display/status mutable |
| HistoricalMigrationPlan/Item | Interchange semantic process plan | immutable after finalization |
| HistoricalMigrationExecution | Interchange semantic process truth | constrained state machine |
| HistoricalMigrationItemOutcome | Interchange semantic evidence | immutable final per execution/item |
| HistoricalSourceBinding | Interchange source↔target provenance authority | immutable identity/fingerprint binding |
| GovernedSubjectExport | Interchange semantic request | immutable |
| GovernedExportSnapshot | Interchange semantic export identity | immutable |
| GovernedExportPackageReceipt | Interchange semantic success evidence | immutable |
| InterchangeConnection | Interchange semantic connection identity | code immutable; name/status mutable |
| RepositoryTransfer | Interchange semantic request/process truth | immutable request |
| RepositoryTransferReceipt | Interchange semantic success evidence | immutable |
| outbox/job/retry/lease/package location | durable/ephemeral mechanism | R10-D |

---

# 22. Enforcement / proof obligations

Before implementation, prove at minimum:

### Audit

1. serving role cannot UPDATE/DELETE AuditEvent;
2. chain head cannot be reset through serving paths;
3. rollback of audited mutation leaves neither domain state nor Audit event/head advance;
4. deleting/mutating/reordering an AuditEvent makes integrity validation fail;
5. two concurrent audited commits receive distinct contiguous committed sequence numbers;
6. canonical hash golden vectors match across runtimes/DB validation tooling;
7. no audited path acquires a domain lock after AuditChainHead;
8. mandatory operation census has no uncovered write path.

### Imported history

9. missing-byte historical REV ordinal reservation prevents later ordinal reuse;
10. imported Revision can be EFFECTIVE without fake RevisionSubmission/ReleaseRecord only when imported proof exists;
11. unknown source effectivity remains unknown; adoption time is not substituted;
12. native successor of imported EFFECTIVE revision supersedes it atomically through normal B4 Release;
13. imported Template content can create a derived Document without fake Submission and without retaining source payload forever;
14. imported Evidence creates no fake captured-by User/time;
15. imported unknown retention anchor never becomes silently disposition-eligible;
16. same historical source identity + same fingerprint reuses; conflicting fingerprint fails closed;
17. dry-run creates zero target semantic rows/Artifacts/Bindings.

### Export / external repository

18. complete export fails closed when one required linked subject is unauthorized;
19. manifest is stable under concurrent unrelated writes and corresponds to one REPEATABLE READ snapshot;
20. package build verifies every file hash before receipt;
21. manifest/package contains no provider storage location identifiers;
22. external PUBLISH without receipt never reads as completed;
23. IMPORT_COPY cannot bypass ordinary target permission/governance;
24. disposed/missing pinned source causes visible transfer failure, not resurrection.

### Transactions / locks

25. whole B1–B6 wait-for graph has no cycle under admitted transaction paths;
26. multi-row same-class locking uses deterministic order;
27. no nested owner commit can produce partial local success;
28. every same-commit matrix row has a negative failure probe proving rollback across all local semantic participants;
29. durable-intent failure rolls back the owning semantic mutation when the intent is required;
30. provider/external failure after commit does not rewrite domain truth and is recoverable through R10-D reconciliation.

---

# 23. Adversarial challenge

## F1 — Audit becomes hidden event sourcing

**Attack:** consumers start reading latest AuditEvent to infer state.  
**Closure:** Audit resource refs are non-authoritative; owner read contract remains mandatory. Architecture/static guard should forbid product-state code from using Audit as canonical state.

## F2 — indefinite Audit retains PII forever

**Attack:** free-form facts/profile names leak into immutable chain.  
**Closure:** bounded schema + field classification forbids human-readable/profile/request payload. Stable UUID skeleton only. A real requirement to erase even stable identifier skeleton reopens privacy/crypto decision.

## F3 — global Audit head becomes throughput/deadlock bottleneck

**Attack:** every audited transaction serializes at one row.  
**Closure:** head is final lock and critical section is tiny. No current throughput evidence justifies segmented chains. Measured contention is reopen trigger.

## F4 — Historical Migration fakes native history

**Attack:** implementation creates System-submitted RevisionSubmission/ReleaseRecord for convenience.  
**Closure:** history_kind + imported content/governance families provide lawful exact alternatives; synthetic native facts are forbidden.

## F5 — CI depends forever on Interchange

**Attack:** current imported Revision viewer/template/review reads migration tables.  
**Closure:** ongoing imported content/governance facts are target-owner snapshots; Interchange owns process/source mapping only.

## F6 — missing historical ordinal is reused

**Attack:** bytes missing means no Revision row, then next native revision reuses old ordinal.  
**Closure:** permanent RevisionOrdinalReservation participates in allocation guard.

## F7 — imported Evidence invents capture actor/time

**Attack:** migration user/time inserted as capture user/time.  
**Closure:** `EvidenceImportedCapture` separates source captured/occurred facts from `adopted_at` and source actor snapshot.

## F8 — export package is internally inconsistent

**Attack:** successive READ COMMITTED queries observe changing Dossier links/revisions.  
**Closure:** narrow REPEATABLE READ snapshot persisted before package assembly.

## F9 — export secretly blocks disposition forever

**Attack:** strong Artifact FKs/package leases become accidental retention.  
**Closure:** no hidden hold. Build may fail visibly if source is lawfully disposed before completion.

## F10 — PUBLISH_COPY truth conflates request with success

**Attack:** request row is treated as external publication proof.  
**Closure:** immutable Receipt is the only success evidence.

## F11 — generic integration engine appears through target/source unions

**Attack:** arbitrary `target_type/id` and free-form action/plugin graph emerges.  
**Closure:** closed typed FKs for admitted target families; four frozen transfer contracts remain distinct.

## F12 — Audit chain lock creates reverse deadlock

**Attack:** Audit append callback calls owner again.  
**Closure:** AuditChainHead is terminal leaf; no callback/domain lock after acquisition.

---

# 24. Essential vs accidental complexity

## Essential

- same-commit minimal Audit evidence;
- tamper-evident append chain;
- PII-minimized immutable skeleton;
- imported vs native truth distinction;
- exact imported content authority;
- historical ordinal preservation;
- true dry-run/idempotency/reconciliation;
- provider-independent complete export manifest;
- explicit external-copy success receipt;
- durable effect intent distinct from semantic result;
- final same-commit matrix and lock DAG;
- narrow stable-snapshot isolation for complete export.

## Accidental/deferred

- event sourcing;
- generic event/integration bus as domain;
- DB-trigger semantic audit;
- free-form Audit payload dumps;
- finite Audit pruning/checkpoints without requirement;
- external cryptographic audit anchors;
- BagIt requirement without consumer;
- signed export packages;
- partial export V1;
- generic transformation/sync engine;
- provider URLs/keys as business identity;
- whole-batch migration transaction;
- global SERIALIZABLE;
- distributed lock service;
- generic plugin workflow around transfers.

---

# 25. Reopen triggers

Reopen the implicated decision only if real evidence shows:

- measured single AuditChainHead contention is material;
- immutable Audit stable identifiers must themselves be lawfully erased;
- external cryptographic/timestamp non-repudiation is contractually required;
- Audit must have finite retention;
- a consumer requires BagIt/another standardized signed package;
- a real partial-export contract is required;
- a new imported historical state cannot be represented truthfully by imported target-owner snapshots;
- a new retention-subject family enters Historical Migration;
- external repository synchronization becomes bidirectional continuous semantic sync rather than explicit copies;
- a real cross-repository transaction/trust boundary requires earlier independent review;
- the whole B1–B6 lock graph cannot be proven acyclic without disproportionate machinery;
- `REPEATABLE READ` export cannot produce the required complete package semantics under real workload/evolution evidence.

Implementation inconvenience, existing legacy schema, current Audit/outbox code shape, or hypothetical future integrations do not reopen B6.

---

# 26. Candidate decision

DevelopmentConexus Method outcome:

> **RESTRUCTURE NOW at design level:** keep domain owners authoritative; replace best-effort/transversal ambiguity with a minimal same-commit tamper-evident Audit, explicit imported-history target-owner forms, specialized Interchange process/receipt contracts, a narrow stable-snapshot export transaction and one final B1–B6 same-commit/lock law.

If operator-accepted:

```text
R10-B3 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
         + B5 refinements B3-R1/R2/R3
         + B6 refinements B3-R4/R5/R6 + B3-R1 imported-content extension

R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL

R10-B5 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
         + B6 Evidence imported-capture bounded refinement

R10-B6 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED

R10-B = INTEGRATED DESIGN BLOCK COMPLETE / NON-FINAL

implementation = BLOCKED
next = R10-C — Artifact / Records Physical Integrity
```

Before R10-C starts, perform one integrated B1↔B6 Global Coherence Review of this corrected candidate. Whole-R10 cold independent review remains deferred until B6/C/D/E/F integration is complete unless a material exception trigger appears.