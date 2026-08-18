# R10-B6 — Audit + Interchange + Cross-owner Atomicity — Integrated Candidate

> **Status:** NON-AUTHORITATIVE — SELF-REVIEWED CORRECTED CANDIDATE — OPERATOR REVIEW / R10 INTEGRATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Input working baseline:** B1/B2 promoted authority + B3/B4/B5 **ACCEPTED FOR R10 INTEGRATION / NON-FINAL**  
> **Input HEAD:** `777d560f1a77e12f7b55b0ea445876e91b046689`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Independent review:** deferred to Whole-R10 unless a material exception trigger appears.

This file is staging analysis. It does **not** independently ratify B6, promote B3–B5 to final authority, or silently rewrite the frozen R3–R9.5 ledger. It closes the relational/transactional `R10-B` design surface and records bounded refinements revealed by real pre-MetalDocs-history and cross-owner counterexamples.

---

# 1. Authority / evidence boundary

Authority order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`
7. accepted non-final B3 + B5-approved B3 refinements
8. accepted corrected non-final B4 + acceptance record
9. accepted corrected non-final B5 + acceptance record

Current code/schema/OpenAPI/legacy ADRs are current-state evidence only.

External comparison evidence only:

- NIST SP 800-92: <https://csrc.nist.gov/pubs/sp/800/92/final>
- RFC 8493 BagIt: <https://www.rfc-editor.org/rfc/rfc8493.html>
- NARA transfer metadata: <https://www.archives.gov/records-mgmt/policy/metadata-compiled>
- PostgreSQL isolation: <https://www.postgresql.org/docs/current/transaction-iso.html>
- PostgreSQL locking: <https://www.postgresql.org/docs/current/explicit-locking.html>
- PostgreSQL binary SHA-256: <https://www.postgresql.org/docs/current/functions-binarystring.html>
- PostgreSQL SECURITY DEFINER guidance: <https://www.postgresql.org/docs/current/sql-createfunction.html>

Useful signal only:

- Audit/log infrastructure is evidence infrastructure, not domain-state authority;
- transfer packages benefit from explicit manifests/checksums independent of storage layout;
- transfer metadata must preserve enough meaning to identify and interpret content later;
- `READ COMMITTED` does not provide one stable snapshot across a multi-statement complete export, while `REPEATABLE READ` does;
- consistent lock order is the primary application defense against deadlock.

MetalDocs does not adopt event sourcing, BagIt, a generic integration bus, global SERIALIZABLE, PKI/TSA/HSM, or provider-specific package identity merely because comparable systems expose them.

---

# 2. Known / Inferred / Unknown / Deferred

## 2.1 Known

### Audit

- AuditEvent is transversal timeline evidence, never canonical resource state;
- critical governed mutations cannot report success without required Audit in the same local DB commit;
- Audit is append-only, tamper-evident, queryable/exportable and has a separate retention regime;
- immutable Audit surviving User-profile erasure must be PII-minimized/non-human-readable;
- B2 access grant/revoke/offboarding evidence must remain reconstructible after current rows disappear;
- Audit cannot become Approval/effectivity/AuthZ/retention-clock authority.

### Interchange

- Historical Migration, Governed Subject Export, External Repository `IMPORT_COPY`/`PUBLISH_COPY`, and Backup/Restore are distinct;
- Tenant Portability Export remains deferred;
- ordinary `IMPORT_COPY` uses ordinary target lifecycle/permissions;
- Historical Migration never fabricates native Submission/Approval/Release/User facts;
- unknown history remains unknown;
- imported EFFECTIVE/SUPERSEDED/OBSOLETE requires source proof;
- reliable legacy revision ordinals may map directly and next native REV must remain above every real imported ordinal;
- exact primary bytes are required for a real imported DocumentRevision; missing historical bytes may preserve evidence but cannot create a fake Revision;
- Historical Migration requires plan, true dry-run, deterministic per-item outcomes, reconciliation, **atomicity per semantic import unit**, and cross-plan idempotency;
- partial batch success is valid;
- source actors remain source snapshots/references; Migration/System is the native actor of imported writes;
- Governed Subject Export uses provider-independent complete manifest with objects/relationships/provenance/filenames/formats/sizes/SHA-256;
- a complete package fails closed if a required subject cannot lawfully be included;
- generated package is temporary delivery output, not automatically Evidence/Artifact;
- signed export packages are not required V1.

### B1–B5

- one local PostgreSQL product DB/schema, typed FKs, RESTRICT/NO ACTION, serving role non-owner, default `READ COMMITTED`;
- cross-owner frozen atomicity uses one local transaction through owner seams;
- no owner hides commits/imports another owner's repository for atomicity;
- provider/object-store calls never join MetalDocs commit;
- required durable effect intent is inserted in the owning local transaction;
- jobs/retries/outbox/projections never become semantic authority;
- B3 exact Submission, B4 governance/effectivity and B5 retention/hold/disposition remain authority;
- every Artifact has exactly one semantic retention root: one DocumentRevision or one Evidence.

## 2.2 Inferred corrected choices

1. Audit uses one small deployment-wide append hash chain; it is not event sourcing.
2. Committed Audit sequence is allocated transactionally from the chain head, not a non-transactional sequence.
3. Audit canonicalization reuses RFC 8785 JCS + SHA-256 from B3.
4. AuditChainHead is always the final contended semantic lock of an audited transaction.
5. V1 immutable Audit skeleton retention = `Indefinite`; finite chain pruning is not invented without requirement.
6. Audit never copies free-form domain reasons/comments/request bodies/profile attributes/provider claims/secrets into immutable skeleton.
7. Human-readable User data is resolved at read time from erasable profile state; no `AuditEventEnrichment` V1.
8. Audit event/facts schemas are versioned bounded product contracts, not arbitrary JSON dumps.
9. A narrow DB-owned Audit append primitive structurally prevents ordinary serving SQL from direct event/head mutation.
10. Historical information needed by ongoing CI/Evidence behavior becomes target-owner imported state, not a permanent runtime dependency on Interchange process tables.
11. Permanent `DocumentRevision.history_kind = NATIVE|IMPORTED` and `Evidence.history_kind = NATIVE|IMPORTED` distinguish provenance without changing lifecycle vocabulary.
12. Imported exact Revision content uses CI-owned `RevisionImportedContent`, never fake `RevisionSubmission`.
13. Imported lifecycle/governance interpretation uses CI-owned `RevisionImportedGovernanceSnapshot`; migration process/source mapping remains Interchange-owned.
14. Real historical ordinals without sufficient bytes use CI-owned permanent `RevisionOrdinalReservation`.
15. Native `RevisionDictionarySnapshot` remains mandatory for NATIVE revisions; imported revisions never resolve current dictionary merely to fill history.
16. B5 `cancelled_at/obsoleted_at` mean native MetalDocs transition instants. An already-terminal imported Revision may retain NULL native timestamp when imported proof owns the historical fact; later native transition writes the native timestamp once.
17. Template origin accepts source `NATIVE_SUBMISSION|IMPORTED_REVISION_CONTENT` without restoring strong retention-payload FKs.
18. Imported current EFFECTIVE with unknown trustworthy effective/review anchor remains EFFECTIVE but is immediately due for Periodic Review; `adopted_as_current_at` is not historical effectivity.
19. Imported Revision RetentionBinding is created in the migration semantic unit despite no native Submission; unknown historical anchor never silently drives disposition.
20. Historical Evidence uses `EvidenceImportedCapture`, never fake native captured-by/time; unknown source capture/occurred anchor remains unknown.
21. Historical Migration idempotency uses a persistent closed typed `HistoricalSourceBinding`. Its long-lived canonical source identity is a SHA-256 digest of the normalized source namespace/entity identity, not necessarily raw human-readable source ID.
22. Raw/human-readable source labels/actors required as historical evidence live only in target-owned imported retention payload or an explicitly bounded process projection; they are not required by the permanent idempotency key.
23. Historical Migration plan items are grouped by immutable `semantic_unit_key`; APPLY commits all items in one unit or none. Per-item outcomes remain deterministic evidence inside the unit.
24. Creating a new historical Document identity is atomic with at least one exact imported Revision/content unit; no empty governed Document shell is created because a revision item later failed.
25. `InterchangeConnection` is stable logical external-repository identity; credentials/endpoints are provider/deployment mechanism. Different logical repository = new connection.
26. Governed export snapshot uses one short `REPEATABLE READ` transaction; package assembly happens after commit.
27. Export manifest uses JCS + SHA-256; BagIt remains optional future contract.
28. Export creates no hidden retention/hold. Lawful disposition racing package build may cause visible build failure; export request does not preserve source forever.
29. PUBLISH_COPY pins exact source root + Artifact UUID/hash/format as snapshots but no strong Artifact FK. Missing/disposed source before success is visible failure.
30. IMPORT_COPY uses ordinary target-owner operations; successful target mutation and `RepositoryTransferReceipt` commit together.
31. B6 closes one same-local-commit matrix and a global partial lock order; operations may start later but never backtrack.

## 2.3 Unknown

- contractual finite Audit retention/pruning/checkpoint;
- external cryptographic Audit anchoring/WORM/TSA/HSM;
- standardized BagIt/signed export requirement;
- explicitly partial export contract;
- provider-specific receipt metadata beyond stable IDs/version snapshots;
- real public integration-event-bus consumer;
- imported source state not truthfully representable through current target snapshots;
- exact R10-F cutover/state-mapping mechanics.

## 2.4 Deferred

```text
physical bytes / malware / ObjectLock / restore / delete verification → R10-C
worker/outbox schema / retry / lease / timer / DLQ / projection       → R10-D
HTTP/API/frontend/export/download/admin journeys                      → R10-E
legacy cutover / bootstrap recovery / final execution/deletion map    → R10-F
```

---

# 3. Root Cause / Target Invariant

Failure class: **transversal truth collapse**.

Audit log, migration/integration state and async state can become shadow authorities, producing best-effort Audit, fake native history, inconsistent exports, provider truth masquerading as product truth and cross-owner partial commits.

Target invariant:

> **Every business/system fact has one semantic owner. Required Audit is atomically appended with the governed mutation but never owns that mutation. Historical Migration preserves source truth through explicit imported target-owner forms without fabricating native actions. External import/publish/export preserves exact provenance/content identity without storage coupling. Every local cross-owner invariant commits once through composition under one compatible lock order; external effects are durable intents followed by explicit receipts.**

```text
AuditEvent != domain state
AuditEvent != outbox/event bus
outbox intent != external-effect receipt
migration execution != native User action
imported governance != ApprovalDecision/ReleaseRecord
export manifest != provider object layout
InterchangeConnection != credential store
composition != semantic owner
```

---

# 4. Alternatives / Method outcome

- **Global event sourcing / Audit as history authority:** reject — duplicate authority + recovery/migration complexity.
- **Best-effort Audit + generic integration/outbox engine:** reject — governed state may exist without mandatory evidence; meanings collapse.
- **DB triggers infer semantic Audit from CRUD:** reject — row mutation is not domain meaning.
- **Small semantic Audit + specialized Interchange + explicit tx/lock matrix:** **recommended Global Maximum**.

Method outcome if accepted: **RESTRUCTURE NOW at design level**.

---

# 5. Audit model

## AuditChainHead

```text
AuditChainHead
  singleton_key SMALLINT PRIMARY KEY CHECK singleton_key=1
  last_sequence BIGINT NOT NULL CHECK last_sequence >= 0
  last_hash BYTEA NOT NULL CHECK octet_length(last_hash)=32
```

Genesis = sequence 0 + 32 zero bytes.

## AuditEvent

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

Actor XOR: USER → User FK only; SYSTEM → product-owned `system_actor_code` only.

`system_actor_code`, `operation_code`, `resource_kind`, `facts_schema` are admitted product vocabularies. Facts schema defines exact keys/types; arbitrary keys fail validation.

`resource_kind/resource_id` intentionally has no FK because Audit attribution does not own resource lifecycle and the target may later be disposed.

`occurred_at` is trusted server/application transaction time, never client-supplied historical time.

## Canonical hash

```text
payload = {
  id, sequence_no, occurred_at,
  actor_kind, actor_user_id?, system_actor_code?,
  operation_code, resource_kind, resource_id,
  facts_schema, facts, correlation_id?
}

canonical = RFC8785_JCS(payload)

event_hash = SHA256(
  prev_hash || 0x00 || UTF8("metaldocs.audit-event.v1") || 0x00 || canonical
)
```

Cross-runtime golden vectors mandatory.

## DB-enforced append seam

Ordinary serving role target:

```text
NO direct INSERT/UPDATE/DELETE on AuditEvent
NO direct UPDATE on AuditChainHead
SELECT only for allowed backend query surface
EXECUTE only on narrow Audit-owned append primitive
```

Append primitive executes inside caller transaction, locks singleton head, enforces next sequence/previous hash, writes immutable event(s), advances head monotonically, and exposes no serving reset/backfill path.

Exact SQL-function vs equivalently strong DB-owned primitive is implementation-spec work. If SECURITY DEFINER is chosen, safe owner/search-path/PUBLIC EXECUTE rules are mandatory. The append path must verify hash continuity and integrity validation independently recomputes JCS/hash so malformed canonical/hash input cannot silently validate.

Multiple required events in one transaction append in stable semantic order while holding the head once.

## Terminal lock law

```text
owner writes/locks
→ required durable Intents
→ AuditChainHead
→ Audit event(s)
→ COMMIT
```

No owner/domain lock after AuditChainHead. Rollback rolls business mutation + Audit/head back together.

---

# 6. Audit privacy / retention

| Data | V1 treatment |
|---|---|
| event id/sequence/time/hash/schema | non-PII technical/governance evidence |
| operation/resource codes | product codes |
| User UUID | PII-minimized stable pseudonymous skeleton allowed to survive UserProfile erasure |
| role/scope/Area/Tenant/assignment ids, outcomes, digests | bounded non-human-readable facts |
| User name/email/username/profile | forbidden from immutable Audit |
| historical source actor/label text | imported retention payload, not indefinite Audit |
| approval/comment/return/disposition reason | owning-domain evidence; Audit references evidence-id/outcome only |
| raw fresh-auth/JWT/password/token/provider claims | forbidden |
| IP/user-agent/request body/header | telemetry/security logging, not semantic Audit V1 |

Read UI may resolve current UserProfile; if absent, show opaque stable User identity. No `AuditEventEnrichment` V1.

Audit immutable skeleton retention V1 = **Indefinite**. This is a product choice, not a claimed statutory period. A finite term requires explicit chain pruning/checkpoint semantics and reopens B6.

Audit read/export never answers canonical business state by latest-event inference.

---

# 7. B3 bounded refinements from historical truth

## B3-R4 — RevisionOrdinalReservation

```text
RevisionOrdinalReservation
  id UUID PRIMARY KEY
  document_id UUID NOT NULL FK Document(id) RESTRICT
  revision_no INTEGER NOT NULL CHECK revision_no >= 1
  reserved_at TIMESTAMPTZ NOT NULL
  UNIQUE(document_id,revision_no)
```

Cross-table DB guard rejects collision with `DocumentRevision(document_id,revision_no)`.

Native next REV = max(Revision ordinals ∪ reservations)+1. Reservation is permanent ordinal identity only, never fake lifecycle.

## B3-R5/R6 — imported Revision provenance/content/governance

Permanent skeleton:

```text
DocumentRevision.history_kind TEXT NOT NULL CHECK NATIVE|IMPORTED
```

### RevisionImportedContent

```text
revision_id UUID PRIMARY KEY FK DocumentRevision RESTRICT
primary_artifact_id UUID NOT NULL FK Artifact RESTRICT
manifest_schema TEXT NOT NULL
manifest_payload JSONB NOT NULL bounded object
content_digest BYTEA32
adopted_at TIMESTAMPTZ NOT NULL
```

Digest = JCS + SHA-256. Manifest freezes exact Artifact facts and trustworthy imported governed/structured state. Unknown historical facts remain unknown. Current dictionary/system values are never resolved merely to fill imported history.

Native `RevisionDictionarySnapshot` remains mandatory for `history_kind=NATIVE`. For IMPORTED it exists only when trustworthy historical dictionary state can truly be represented; otherwise imported manifest/governance explicitly carries absence/unknown.

### RevisionImportedGovernanceSnapshot

```text
revision_id UUID PRIMARY KEY FK DocumentRevision RESTRICT
source_system_code_snapshot TEXT NOT NULL
source_object_id_snapshot TEXT NULL
source_revision_label TEXT NULL
source_state TEXT NOT NULL
source_effective_at TIMESTAMPTZ NULL
source_superseded_at TIMESTAMPTZ NULL
source_obsoleted_at TIMESTAMPTZ NULL
source_cancelled_at TIMESTAMPTZ NULL
adopted_as_current_at TIMESTAMPTZ NULL
source_actor_snapshot JSONB NULL bounded object
source_governance_snapshot JSONB NOT NULL bounded object
```

These fields are retained with the Revision unit, not indefinite Audit/idempotency state.

Target `DocumentRevision.state` remains CI authority. No synthetic ApprovalInstance/Decision/ReleaseRecord/internal User action.

Historical IMPORTED revisions do not use target `SUBMITTED`; that state asserts a native Submission boundary. Historical draft/submitted-like source material remains source evidence or enters ordinary import/native DRAFT when it must become editable.

V1 privileged historical target states = truthfully proven `EFFECTIVE|SUPERSEDED|OBSOLETE|CANCELLED`. New mappings require explicit reopen/R10-F rule.

### Native terminal timestamp clarification

`cancelled_at/obsoleted_at` are native MetalDocs transition times:

```text
NATIVE + CANCELLED → cancelled_at required
NATIVE + OBSOLETE  → obsoleted_at required
```

Already-terminal IMPORTED revisions may keep native timestamp NULL when imported snapshot proves historical terminal state. If an imported current revision later undergoes a native post-adoption terminal transition, native timestamp is written once. Later native supersession always uses B4 ReleaseRecord time.

## B3-R1 imported Template extension

```text
DocumentOrigin
  derived_document_id
  source_template_revision_id
  source_content_kind NATIVE_SUBMISSION|IMPORTED_REVISION_CONTENT
  source_submission_id_snapshot UUID NULL
  source_content_digest BYTEA32
  source_artifact_sha256 BYTEA32
  source_content_format TEXT
  created_at
```

Native kind requires Submission snapshot; imported kind requires NULL. Creation still serializes against current template effectivity and exact source content authority.

## Imported Periodic Review / retention

Imported current EFFECTIVE:

```text
trusted source effectivity/review anchor known → normal due calculation
unknown                                      → preserve unknown + immediately DUE
```

DUE does not invalidate EFFECTIVE. First native PeriodicReviewRecord establishes later ordinary review anchor.

Imported Revision with exact content creates RetentionBinding inside migration semantic unit even without native Submission.

Retention anchor: current EFFECTIVE none; later native supersession B4 Release time; already-historical imported state uses trustworthy source lifecycle timestamp only; unknown anchor never silently disposition-eligible.

---

# 8. B5 bounded refinement — historical Evidence

```text
Evidence.history_kind TEXT NOT NULL CHECK NATIVE|IMPORTED
```

Native EvidenceCapture unchanged.

```text
EvidenceImportedCapture
  evidence_id UUID PRIMARY KEY FK Evidence RESTRICT
  primary_artifact_id UUID NOT NULL FK Artifact RESTRICT
  reference_value TEXT NULL
  governed_metadata JSONB NOT NULL bounded object
  occurred_at TIMESTAMPTZ NULL
  source_captured_at TIMESTAMPTZ NULL
  source_actor_snapshot JSONB NULL bounded object
  source_provenance JSONB NOT NULL bounded object
  original_filename TEXT NULL
  adopted_at TIMESTAMPTZ NOT NULL
```

CAPTURED Evidence has exactly one native or imported capture authority according to history_kind.

Migration unit atomically creates Evidence identity + exact Artifact root + imported capture + frozen Dossier/name/sequence + RetentionBinding + applicable Holds.

No fake internal captured-by/time. CAPTURED_AT uses trustworthy source_captured_at only; OCCURRED_AT uses trustworthy occurred_at only; unknown never silently disposition-eligible. `adopted_at` is not substitute historical capture time.

---

# 9. Historical Migration model

## Source namespace

```text
HistoricalMigrationSource
  id UUID PK
  code TEXT UNIQUE immutable
  name TEXT
  source_kind TEXT
  status ACTIVE|INACTIVE
  created_at TIMESTAMPTZ
```

No credentials/endpoints/secrets.

## Complete Plan / Items

Plan is inserted/finalized as one complete immutable semantic plan; no selectable partially built plan.

```text
HistoricalMigrationPlan
  id UUID PK
  source_id UUID FK Source RESTRICT
  mode CURRENT_STATE|FULL_HISTORY
  plan_schema TEXT
  plan_digest BYTEA32
  created_by_user_id UUID FK User RESTRICT
  created_at TIMESTAMPTZ

HistoricalMigrationPlanItem
  id UUID PK
  plan_id UUID FK Plan RESTRICT
  semantic_unit_key UUID NOT NULL
  item_order BIGINT >= 1
  source_entity_kind TEXT NOT NULL
  source_entity_key_digest BYTEA32 NOT NULL
  source_fingerprint BYTEA32 NOT NULL
  target_kind DOCUMENT|DOCUMENT_REVISION|REVISION_ORDINAL|EVIDENCE|DOSSIER
  target_hint JSONB bounded object
  UNIQUE(plan_id,item_order)
  UNIQUE(plan_id,source_entity_kind,source_entity_key_digest)
```

`source_entity_key_digest = SHA256(canonical source namespace + entity kind + normalized source identity)`; long-lived idempotency does not require raw human-readable identifier.

Plan digest covers ordered units/items using versioned JCS + SHA-256.

A semantic unit groups all items that must commit together. Example:

```text
new historical Document
  Document source item
  + first exact Revision source item
  = one semantic_unit_key
```

Thus target never contains an empty Document shell because its required first exact Revision failed.

FULL_HISTORY additional exact revisions/reservations may be separate semantic units after Document identity exists. CURRENT_STATE creates Document + exact current Revision in one unit.

## Execution / outcomes

```text
HistoricalMigrationExecution
  id UUID PK
  plan_id UUID FK Plan RESTRICT
  mode DRY_RUN|APPLY
  requested_by_user_id UUID FK User RESTRICT
  requested_at TIMESTAMPTZ
  completed_at TIMESTAMPTZ NULL
  status RUNNING|COMPLETED|PARTIAL|FAILED

HistoricalMigrationItemOutcome
  id UUID PK
  execution_id UUID FK Execution RESTRICT
  plan_item_id UUID FK Item RESTRICT
  outcome WOULD_CREATE|WOULD_REUSE|CONFLICT|INVALID|CREATED|REUSED|FAILED
  reason_code TEXT NULL
  target_id_snapshot UUID NULL
  recorded_at TIMESTAMPTZ
  UNIQUE(execution_id,plan_item_id)
```

One semantic unit transaction emits final outcomes for all its items. Transient retries/errors remain R10-D.

## HistoricalSourceBinding

```text
HistoricalSourceBinding
  id UUID PK
  source_id UUID FK HistoricalMigrationSource RESTRICT
  source_entity_kind TEXT
  source_entity_key_digest BYTEA32
  source_fingerprint BYTEA32

  document_id UUID NULL FK Document RESTRICT
  document_revision_id UUID NULL FK DocumentRevision RESTRICT
  revision_ordinal_reservation_id UUID NULL FK RevisionOrdinalReservation RESTRICT
  evidence_id UUID NULL FK Evidence RESTRICT
  dossier_id UUID NULL FK Dossier RESTRICT

  bound_at TIMESTAMPTZ
  UNIQUE(source_id,source_entity_kind,source_entity_key_digest)
```

Exactly one target FK per source item. Same source digest + same fingerprint/coherent target = REUSE; conflict = fail closed. No heuristic merge.

Raw source actors/labels necessary as historical evidence belong target imported retention payload, not this permanent mapping.

---

# 10. Dry-run / APPLY law

### DRY_RUN

Runs same source-fingerprint/mapping/format/uniqueness/target-eligibility/retention checks as APPLY but creates **zero target semantic rows, Artifacts, RetentionBindings or domain governance facts**. Only Interchange execution/outcomes persist.

### APPLY per semantic unit

1. re-read and verify every unit item's source fingerprint;
2. existing matching SourceBinding + coherent target → REUSE item;
3. any conflict fails unit closed before target mutation;
4. stage exact bytes outside DB tx where required;
5. open one semantic unit transaction;
6. call target-owner privileged migration seams in global lock order;
7. create target imported identities/content/governance/capture + RetentionBinding/holds;
8. confirm Artifacts with one semantic retention root;
9. create all unit SourceBindings;
10. insert all unit final per-item outcomes;
11. insert required durable intents;
12. append required migration Audit event(s) last;
13. commit.

Any local failure rolls the **whole semantic unit** back. Batch may still be PARTIAL because other units committed.

Migration/System is native Audit actor; source actors remain imported provenance.

---

# 11. Governed Subject Export

## Root

```text
GovernedSubjectExport
  id UUID PK
  document_id UUID NULL FK Document RESTRICT
  evidence_id UUID NULL FK Evidence RESTRICT
  dossier_id UUID NULL FK Dossier RESTRICT
  requested_by_user_id UUID FK User RESTRICT
  requested_at TIMESTAMPTZ
```

Exactly one root.

## COMPLETE closure

### Document
Includes stable Document identity; all Revision identity skeletons + ordinal reservations; every currently present retained Revision-unit payload/history; disposition/records skeleton where payload was lawfully disposed; exact included Artifact bytes; relevant Template origin and Dossier relationship references without recursively exporting Dossier contents.

Disposed revisions are exported truthfully as identity/disposition evidence, never as missing-byte fiction.

### Evidence
Includes Evidence identity + currently present native/imported capture payload/Artifact + records/disposition evidence + primary/secondary Dossier relationship identities. No recursive other-Dossier content.

### Dossier
Includes Dossier identity/provenance + all currently linked Documents + all Evidence where Dossier is primary or secondary, each using bounded Document/Evidence closure. Traversal stops; no Dossier graph/transitive expansion.

Any required subject unauthorized/unavailable under canonical AuthZ makes COMPLETE export fail closed. No PARTIAL V1.

## Stable snapshot

One short PostgreSQL `REPEATABLE READ` transaction:

1. resolve complete closure under one snapshot;
2. apply canonical AuthZ to every required subject;
3. enumerate identities/relationships/provenance/hashes without provider locations;
4. build canonical manifest;
5. insert export + immutable snapshot;
6. insert build Intent;
7. append Audit last;
8. commit.

No byte copying while transaction open. Serialization failure is visible/retryable, not reason for global isolation change.

## Snapshot / package receipt

```text
GovernedExportSnapshot
  export_id UUID PK FK GovernedSubjectExport RESTRICT
  manifest_schema TEXT
  manifest_payload JSONB bounded object
  manifest_digest BYTEA32
  snapshotted_at TIMESTAMPTZ

GovernedExportPackageReceipt
  export_id UUID PK FK GovernedSubjectExport RESTRICT
  package_sha256 BYTEA32
  size_bytes BIGINT >= 0
  completed_at TIMESTAMPTZ
```

Manifest contains package version, root, objects, relationships, provenance, canonical filenames, formats/media types, sizes, SHA-256, content/governance digests, disposition truth where applicable and `completeness=COMPLETE`. No provider bucket/key/URL/version identity.

Manifest digest = JCS + SHA-256.

Package output is temporary delivery output, not semantic Artifact/Evidence. Builder verifies each emitted file against snapshot before receipt. Temporary location/expiry/download token is R10-D/E.

Export never changes LegalHold/retention or counts as Distribution acknowledgement.

---

# 12. External Repository Interchange

## InterchangeConnection

```text
id UUID PK
code TEXT UNIQUE immutable
name TEXT
connector_kind TEXT bounded supported code
status ACTIVE|INACTIVE
created_at TIMESTAMPTZ
```

Logical external repository identity only. Credentials/tokens/endpoints/secrets are deployment/provider config keyed by id. Materially different logical repository requires new connection.

B5 DossierExternalReference FK closes here; `(connection_id,entity_kind,external_id)` remains unique.

## RepositoryTransfer

```text
id UUID PK
connection_id UUID FK InterchangeConnection RESTRICT
direction IMPORT_COPY|PUBLISH_COPY
requested_by_user_id UUID FK User RESTRICT
requested_at TIMESTAMPTZ

external_entity_kind TEXT NULL
external_entity_id TEXT NULL

source_document_revision_id UUID NULL FK DocumentRevision RESTRICT
source_evidence_id UUID NULL FK Evidence RESTRICT
source_artifact_id_snapshot UUID NULL
source_sha256 BYTEA NULL/BYTEA32
source_content_format TEXT NULL

intended_target_kind TEXT NULL CHECK NULL OR DOCUMENT|EVIDENCE
requested_target_document_id UUID NULL FK Document RESTRICT
```

PUBLISH pins one internal source root + exact Artifact snapshot, no Artifact FK. IMPORT pins external source + intended ordinary target operation; optional existing Document target supports ordinary new-revision import.

Request transaction = RepositoryTransfer + external-effect Intent + Audit. Request is never success proof.

PUBLISH worker re-proves pinned Artifact UUID/hash/format still exists. Missing/disposed → visible failure, no resurrection/hidden retention.

IMPORT worker fetches/stages/validates bytes then invokes ordinary target owner under current authorization/governance.

## RepositoryTransferReceipt

```text
transfer_id UUID PK FK RepositoryTransfer RESTRICT
external_entity_kind TEXT
external_entity_id TEXT
external_version_snapshot TEXT NULL
target_document_id UUID NULL FK Document RESTRICT
target_document_revision_id UUID NULL FK DocumentRevision RESTRICT
target_evidence_id UUID NULL FK Evidence RESTRICT
completed_at TIMESTAMPTZ
```

For IMPORT_COPY, target semantic creation/revision/capture **and receipt commit in one local target transaction**. For PUBLISH_COPY, only confirmed external provider success creates receipt.

No receipt = no success claim. Provider URLs remain display/mechanism data.

External IDs used as semantic receipt/reference must be stable provider object identifiers, not transient URL/path/display name where a stronger provider object identity exists.

---

# 13. Mandatory semantic Audit classes

Audit does not mean “every row”. Required same-commit classes:

### Governance configuration

- Tenant settings;
- DocumentType/category/numbering governance changes and first numbering freeze;
- ApprovalPolicy/version activation/deactivation and DocumentType approval binding changes;
- TemplateUse changes;
- Tenant Dictionary value changes;
- DossierType/EvidenceType eligibility/naming/format configuration;
- DocumentType/EvidenceType retention-rule changes;
- Distribution live audience configuration changes;
- InterchangeConnection create/status/identity-safe configuration transition;
- Historical Migration plan finalization and APPLY semantic units;
- Governed Subject Export accepted snapshot.

### B2 identity/access

Area/User lifecycle; governed UserProfile mutation/erasure; provider binding acceptance/replacement/disable/erasure where admitted; admin Session revoke/offboard; Group create/rename/delete; membership add/remove; RoleAssignment grant/revoke.

Grant/revoke facts preserve assignment id, subject stable id, role, scope and operation without profile duplication.

### B3/B4

Document creation; responsibility change; Revision creation when entering governed lifecycle; SUBMIT; Revision cancel/obsolete; PeriodicReviewRecord; ApprovalStepDecision; approval cancel/withdraw/reassign; Rendition semantic success; Release; explicit Distribution acknowledgement.

### B5

Dossier create/title/archive/re-enable; Dossier↔Document link/unlink; Evidence CAPTURE/VOID; captured secondary-Dossier link/unlink; RetentionExtension; LegalHold activate/release; DispositionFence; completed DispositionRecord.

### B6 transfer outcomes

COPY request; successful COPY receipt; Governed Export package receipt where a complete package is successfully materialized.

Not mandatory semantic Audit by default:

```text
WorkingContent autosave/edit
EditorSession heartbeat/lease
EvidenceDraft edit
ordinary EditorialComment/SubmissionFeedback collaboration
search/view/download
notification delivery/read
worker retry/lease/DLQ
projection rebuild
provider health/telemetry
```

Later requirements may promote a specific class; no generic CRUD audit.

---

# 14. Same-local-commit matrix

`Audit` = required event(s) appended last. `Intent` = same-commit durable mechanism only when later work required.

| Operation | Atomic semantic set |
|---|---|
| User offboarding | User disable + local Session revoke + membership/direct-grant revoke + admitted binding state + Audit + provider-disable Intent |
| Document + REV001 | number allocation + Document + Revision + native RevisionDictionarySnapshot + WorkingContent + template adjuncts + Audit |
| first SUBMIT | WC generation consume + RevisionSubmission + ApprovalRequirement/ReleasePlan + Approval init + RetentionBinding + active Holds + Audit + evaluation Intent where needed |
| same-REV resubmit | new Submission + B4 per-Submission requirements/Approval init + Audit + evaluation Intent; existing Binding unchanged |
| Approval RETURN/CANCEL/WITHDRAW | Revision DRAFT return generation law + Approval terminal state + Audit |
| Approval ACCEPT | immutable decision + Step/Instance transition + Audit + final-gate evaluation Intent if needed |
| Rendition success | Artifact confirm + Rendition relation + Audit + release evaluation Intent |
| Release | candidate EFFECTIVE + predecessor SUPERSEDED + ReleaseRecord + DistributionObligations + Audit + notification/search Intents |
| Periodic Review | exact current-Revision record + Audit |
| native/imported Evidence CAPTURE | capture authority + Artifact root + frozen Evidence/context + RetentionBinding + applicable Holds + Audit |
| Dossier/secondary link enters hold | relation + HoldSubject materialization + Audit |
| LegalHold activation | Hold + complete current materialization + Audit |
| RetentionExtension | extension + Audit |
| DispositionFence | fence + Audit + physical-delete Intent |
| Disposition completion | retained payload/Artifact semantic cleanup + DispositionRecord + Audit |
| Historical Migration semantic unit | all target imported facts for unit + all source bindings + Bindings/holds + all item outcomes + Audit |
| Governed Export snapshot | export root + complete immutable manifest + build Intent + Audit under REPEATABLE READ |
| COPY request | RepositoryTransfer + external-effect Intent + Audit |
| IMPORT_COPY success | ordinary target semantic creation/revision/capture + RepositoryTransferReceipt + Audit in one target transaction |
| PUBLISH_COPY success | RepositoryTransferReceipt + Audit |

Single-owner governance configuration mutations above also append Audit in their owner transaction; only cross-owner sets are expanded in this table.

No provider/object transfer is inside DB commit.

---

# 15. Composition / lock law

`internal/composition` owns no durable meaning. Owner seams share transaction, do not commit independently, do not import another owner's repository, and expose all local failure before Audit.

Global target is partial order, not lock-manager framework.

## Main branch

```text
B2 eligibility roots according to promoted B2 order
→ configuration/allocation roots
  DocumentType / EvidenceType / Approval config / NumberSeries / EvidenceSequence
→ business roots
  Document(s) UUID order
  → Evidence(s) UUID order
  → Dossier(s) UUID order
→ child/execution
  Revision / WorkingContent / ApprovalInstance+Step / Evidence capture
→ RetentionBinding(s) UUID order
→ Artifact existing-row coordination only if needed
→ durable Intents
→ AuditChainHead ALWAYS LAST
```

May start later if never backtracking.

## Group branch

B2 Group deletion stays `Group → membership/group grants` and never later acquires governed Document/Evidence/Dossier roots. B4 may read/lock live Group dependencies after Document/Approval root because reverse governed-root acquisition is forbidden.

## Approval return

Path that may return/cancel/withdraw a Revision locks Document/Revision before Approval rows. Pure intermediate ACCEPT may be Approval-local then Audit, never later Document.

## Dossier/Records

```text
Document or Evidence root
→ Dossier roots UUID order
→ RetentionBinding
→ Audit
```

Dossier-wide Hold may start at Dossier then deterministic Bindings and must not backtrack to subject roots.

## Audit leaf

**NO owner/domain lock after AuditChainHead.** This must be structurally testable.

---

# 16. Export isolation law

Product default remains `READ COMMITTED`.

Only short COMPLETE-export snapshot transaction uses `REPEATABLE READ`, because multi-relation closure must represent one real database snapshot without holding a large graph locked for package assembly.

Transaction ends after manifest + build Intent + Audit. Bytes assemble after commit. No other B6 operation gains stronger isolation by default.

---

# 17. Persistence / mutation classes

| Family | Class | Mutation law |
|---|---|---|
| AuditChainHead | integrity-support semantic mechanism | narrow monotonic head |
| AuditEvent | Audit semantic authority | immutable append-only |
| RevisionOrdinalReservation | CI identity skeleton | permanent immutable |
| DocumentRevision.history_kind | CI identity | immutable |
| RevisionImportedContent | CI imported content authority | immutable retention payload |
| RevisionImportedGovernanceSnapshot | CI imported governance authority | immutable retention payload |
| Evidence.history_kind | Documentary Context identity | immutable |
| EvidenceImportedCapture | imported capture authority | immutable retention payload |
| HistoricalMigrationSource | Interchange config | code immutable; display/status mutable |
| HistoricalMigrationPlan/Item | Interchange plan | immutable complete plan |
| HistoricalMigrationExecution | Interchange process | constrained state machine |
| HistoricalMigrationItemOutcome | process evidence | immutable final per item/execution |
| HistoricalSourceBinding | source↔target provenance | immutable digest binding |
| GovernedSubjectExport | export request | immutable |
| GovernedExportSnapshot | export semantic identity | immutable |
| GovernedExportPackageReceipt | export success evidence | immutable |
| InterchangeConnection | external logical identity | code immutable; name/status mutable |
| RepositoryTransfer | transfer request | immutable |
| RepositoryTransferReceipt | transfer success evidence | immutable |
| job/outbox/retry/lease/temp package location | mechanism | R10-D |

---

# 18. Proof obligations

### Audit
1. serving trust cannot direct-write AuditEvent/reset head;
2. audited business rollback also rolls Audit/head back;
3. mutate/delete/reorder event → validator fails;
4. concurrent audited commits produce distinct contiguous committed sequences;
5. canonical hash goldens match;
6. mandatory operation/config census has no admitted uncovered path;
7. no audited path locks domain after Audit head;
8. canonical owner reads never infer state from Audit.

### Imported history
9. missing-byte ordinal reservation prevents reuse;
10. new historical Document + first exact Revision is one semantic unit; failure leaves no empty Document shell;
11. imported Revision has no fake Submission/Approval/Release;
12. NATIVE dictionary snapshot required; imported unknown dictionary never filled from current values;
13. imported terminal state may retain NULL native transition time with imported proof;
14. unknown effectivity/capture remains unknown;
15. native successor Release supersedes imported EFFECTIVE normally;
16. imported Template works without fake Submission/indefinite source retention;
17. imported Evidence has no fake User/time;
18. unknown imported retention anchor never becomes eligible;
19. same source key digest + same fingerprint reuses; changed fingerprint conflicts;
20. long-lived source idempotency state does not require raw human-readable source ID;
21. dry-run creates zero target semantic rows/Artifacts/Bindings;
22. failure of any item in semantic unit rolls whole unit target state/outcomes/bindings back.

### Export/transfers
23. Document/Evidence/Dossier closure matches bounded COMPLETE contract;
24. any required unauthorized subject makes export fail closed;
25. disposed history exports truthful skeleton/disposition evidence;
26. concurrent writes cannot make manifest span inconsistent snapshots;
27. package verifies every hash before receipt;
28. manifest/package has no provider storage identifiers;
29. no COPY receipt means no success claim;
30. IMPORT target mutation + receipt is atomic;
31. IMPORT cannot bypass target governance;
32. disposed/missing PUBLISH source fails without resurrection.

### Transactions
33. whole B1–B6 wait-for graph is acyclic;
34. same-class multi-row locks deterministic;
35. no nested owner commit produces partial local success;
36. each matrix row has negative rollback probe;
37. required Intent failure rolls business mutation back;
38. provider failure after commit leaves canonical domain truth intact and R10-D-recoverable.

---

# 19. Adversarial findings closed

- Audit hidden event sourcing → owner reads cannot use Audit as state.
- Indefinite Audit PII → bounded non-human skeleton; free text/profile forbidden.
- Audit direct-write bypass → DB-owned append seam, no direct serving writes.
- Audit head deadlock → terminal leaf; measured throughput only can reopen segmentation.
- Migration fake native history → imported target-owner forms.
- CI depends forever on Interchange → ongoing imported facts target-owned.
- Missing historical ordinal reused → permanent reservation.
- Imported terminal time fabricated → native timestamp means native transition only.
- Imported Evidence fake actor/time → dedicated imported capture.
- Historical Document shell survives failed Revision → semantic-unit atomic grouping.
- Permanent migration idempotency stores PII → digest key, raw evidence lives in retained payload when needed.
- Export inconsistent → short REPEATABLE READ snapshot.
- Export closure ambiguous → explicit bounded COMPLETE graph.
- Export becomes hidden hold → no preservation side effect; build may fail visibly.
- COPY request treated as success → receipt only.
- IMPORT target without receipt → same local target transaction.
- Governance configuration missing from Audit → explicit config census.
- Generic integration registry → closed typed unions/four distinct contracts.
- Audit reverse lock → no owner lock after head.

---

# 20. Essential / accidental complexity and reopen

Essential:

```text
same-commit bounded Audit
append tamper chain
PII-minimized skeleton
native vs imported truth
exact imported content/capture
historical ordinal preservation
semantic-unit migration atomicity
true dry-run/idempotency/reconciliation
complete stable export manifest
external-effect receipt
final tx matrix + lock DAG
narrow REPEATABLE READ export snapshot
```

Rejected/deferred:

```text
event sourcing
generic integration/event bus domain
DB-trigger semantic inference
free-form Audit dumps
finite Audit pruning without requirement
external cryptographic anchors
BagIt/signatures without consumer
partial export V1
generic transform/sync engine
provider URL/key identity
whole-batch migration transaction
global SERIALIZABLE
distributed lock service
```

Reopen only on material evidence: measured Audit-head contention; finite/erasure requirement for Audit skeleton; external non-repudiation contract; standardized/partial export consumer; new truthful imported-state need; new retention subject in migration; continuous bidirectional sync requirement; cross-repository trust boundary; lock-graph proof failure; or real workload proving export isolation insufficient.

Implementation inconvenience/current schema/hypothetical integrations do not reopen.

---

# 21. Candidate decision / next state if accepted

```text
R10-B3 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
         + B5 refinements B3-R1/R2/R3
         + B6 refinements B3-R4/R5/R6
         + B3-R1 imported-content extension
         + B3-R2 native-vs-imported terminal-time clarification

R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL

R10-B5 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
         + imported-Evidence refinement

R10-B6 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED

R10-B = INTEGRATED DESIGN BLOCK COMPLETE / NON-FINAL
implementation = BLOCKED
next = R10-C — Artifact / Records Physical Integrity
```

This corrected candidate has undergone internal B1↔B6 coherence/adversarial review. Operator adjudication is still required before any router/status promotion. Whole-R10 cold independent review remains deferred until integrated B6/C/D/E/F unless a material exception trigger appears.