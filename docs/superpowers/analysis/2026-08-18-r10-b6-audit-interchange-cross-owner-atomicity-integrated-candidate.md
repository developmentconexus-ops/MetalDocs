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

Read authority in repository order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`
7. accepted non-final B3 + B5-approved B3 refinements
8. accepted corrected non-final B4 + acceptance record
9. accepted corrected non-final B5 + acceptance record

Current code/schema/OpenAPI/legacy ADRs remain current-state evidence only.

External comparison evidence only:

- NIST SP 800-92: <https://csrc.nist.gov/pubs/sp/800/92/final>
- RFC 8493 BagIt: <https://www.rfc-editor.org/rfc/rfc8493.html>
- NARA transfer metadata: <https://www.archives.gov/records-mgmt/policy/metadata-compiled>
- PostgreSQL isolation: <https://www.postgresql.org/docs/current/transaction-iso.html>
- PostgreSQL locking: <https://www.postgresql.org/docs/current/explicit-locking.html>
- PostgreSQL SHA-256 binary function precedent: <https://www.postgresql.org/docs/current/functions-binarystring.html>
- PostgreSQL SECURITY DEFINER safety guidance: <https://www.postgresql.org/docs/current/sql-createfunction.html>

Useful signal only:

- Audit/log infrastructure is evidence infrastructure, not domain-state authority;
- transfer packages benefit from explicit manifests/checksums independent of storage layout;
- transfer metadata must preserve enough meaning to identify and interpret content later;
- `READ COMMITTED` does not provide one snapshot across a multi-statement export, whereas `REPEATABLE READ` does;
- consistent lock order is the primary application defense against deadlock.

MetalDocs does not adopt event sourcing, BagIt, generic integration-bus semantics, global SERIALIZABLE, PKI/TSA/HSM, or provider-specific package identity merely because those mechanisms exist.

---

# 2. Known / Inferred / Unknown / Deferred

## 2.1 Known

### Audit

- AuditEvent is transversal timeline evidence, never canonical resource state;
- critical governed mutations cannot report success without required Audit in the same local DB commit;
- Audit is append-only, tamper-evident, queryable/exportable and has a separate retention regime;
- the immutable Audit skeleton surviving lawful User-profile erasure must be PII-minimized/non-human-readable;
- B2 grant/revoke/offboarding evidence must remain reconstructible after current rows disappear;
- Audit cannot become Approval/effectivity/AuthZ/retention-clock authority.

### Interchange

- Historical Migration, Governed Subject Export, External Repository `IMPORT_COPY`/`PUBLISH_COPY`, and Backup/Restore are distinct contracts;
- Tenant Portability Export remains deferred;
- ordinary `IMPORT_COPY` follows ordinary target lifecycle and permissions;
- Historical Migration never fabricates native Submission/Approval/Release/User facts;
- unknown history remains unknown;
- imported EFFECTIVE/SUPERSEDED/OBSOLETE requires proof;
- reliable legacy revision ordinals may map directly and next native REV must remain above every real imported ordinal;
- exact primary bytes are required for a real imported DocumentRevision; missing historical bytes may preserve evidence but cannot create a fake Revision;
- Historical Migration requires plan, true dry-run, deterministic per-item result, reconciliation, per-semantic-unit atomicity and cross-plan idempotency;
- source actors remain source snapshots/references; native migration actor is Migration/System principal;
- Governed Subject Export uses a provider-independent complete manifest with relationships/provenance/filenames/formats/sizes/SHA-256;
- a complete package fails closed if a required subject cannot lawfully be included;
- generated package output is temporary delivery output, not automatically Evidence/Artifact;
- signed packages are not required V1.

### B1–B5 integration

- one local PostgreSQL product DB/schema, typed FKs, RESTRICT/NO ACTION, serving role non-owner, default `READ COMMITTED`;
- cross-owner frozen atomicity uses one local transaction through owner seams;
- no owner hides a commit/imports another owner's persistence to obtain atomicity;
- provider/object-store calls never join MetalDocs commit;
- required durable effect intent is inserted in the owning local transaction;
- jobs/retries/outbox/projections never become semantic authority;
- B3 exact Submission, B4 governance/effectivity and B5 retention/hold/disposition remain authoritative;
- every Artifact has exactly one semantic retention root: one DocumentRevision or one Evidence.

## 2.2 Inferred corrected choices

1. Audit uses one small deployment-wide append hash chain; it is not event sourcing.
2. Committed Audit sequence is allocated transactionally from the chain head, not a non-transactional DB sequence.
3. Audit canonicalization reuses RFC 8785 JCS + SHA-256 from B3.
4. AuditChainHead is always the final contended semantic lock of an audited transaction.
5. V1 immutable Audit skeleton retention = `Indefinite`; finite chain pruning is not invented without requirement.
6. Audit never copies free-form domain reasons/comments/request bodies/profile attributes/provider claims/secrets into its immutable skeleton.
7. Human-readable User data is resolved at read time from erasable profile state; no `AuditEventEnrichment` table V1.
8. Audit event/facts schemas are versioned bounded product contracts, not arbitrary JSON dumps.
9. A narrow DB-owned Audit append primitive must structurally prevent ordinary serving SQL from directly inserting arbitrary chain rows or resetting AuditChainHead. Direct table INSERT/UPDATE privileges are not the target write contract.
10. Historical information needed by ongoing CI/Evidence behavior becomes target-owner imported state, not a permanent runtime dependency on Interchange process tables.
11. Permanent `DocumentRevision.history_kind = NATIVE|IMPORTED` and `Evidence.history_kind = NATIVE|IMPORTED` distinguish provenance without changing business lifecycle vocabulary.
12. Imported exact Revision content uses CI-owned `RevisionImportedContent`, never fake `RevisionSubmission`.
13. Imported lifecycle/governance interpretation uses CI-owned `RevisionImportedGovernanceSnapshot`; source-system/process mapping remains Interchange-owned.
14. Real historical ordinals without sufficient bytes use CI-owned permanent `RevisionOrdinalReservation`.
15. Native `RevisionDictionarySnapshot` remains mandatory for NATIVE revisions; imported revisions never resolve current dictionary merely to fill history. Imported dictionary state is present only if trustworthy source mapping exists and is represented explicitly in imported content/governance payload.
16. B5 native `cancelled_at/obsoleted_at` are **native MetalDocs transition instants**. For an already-terminal imported Revision they may remain NULL when imported governance proof carries the historical source terminal fact. A later native post-adoption terminal transition writes the corresponding native timestamp once.
17. Template origin accepts exact source kind `NATIVE_SUBMISSION|IMPORTED_REVISION_CONTENT` without restoring strong retention-payload FKs.
18. Imported current EFFECTIVE with unknown trustworthy effective/review anchor remains EFFECTIVE but is immediately due for Periodic Review; `adopted_as_current_at` is not relabeled as historical effectivity.
19. Imported Revision RetentionBinding is created in the migration unit despite no native Submission; trustworthy imported lifecycle anchors may drive retention, unknown anchors never silently do.
20. Historical Evidence uses `EvidenceImportedCapture`, never fake native captured-by/time. Imported retention uses trustworthy source captured/occurred fact only.
21. Historical Migration cross-plan idempotency uses a persistent closed typed `HistoricalSourceBinding`, not `target_type/target_id` polymorphism.
22. `InterchangeConnection` is stable logical external-repository identity; credentials/endpoints are provider/deployment mechanism. Rebinding to a different logical repository requires a new connection.
23. Governed export semantic snapshot uses one short `REPEATABLE READ` transaction; package assembly is asynchronous after commit.
24. Export manifest uses JCS + SHA-256; BagIt remains an optional future external contract.
25. Export creates no hidden retention/hold. Lawful disposition racing package build may make build fail visibly; export request does not silently preserve source forever.
26. PUBLISH_COPY pins exact source root + Artifact UUID/hash/format as snapshots but no strong Artifact FK. Missing/disposed source before external success produces visible failure.
27. IMPORT_COPY uses ordinary target-owner creation/revision/capture operations; successful target mutation and `RepositoryTransferReceipt` commit together.
28. B6 closes one same-local-commit matrix and a global partial lock order; operations may start later but never backtrack.

## 2.3 Unknown

- contractual finite Audit retention/pruning/checkpoint requirements;
- external cryptographic Audit anchoring/WORM/TSA/HSM;
- standardized BagIt/signed export requirement;
- explicitly partial export contract;
- provider-specific external receipt metadata beyond stable IDs/version snapshots;
- a real public integration event bus consumer;
- a historical source state not truthfully representable through current imported-target snapshots;
- exact R10-F legacy state/cutover mechanics.

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

Audit log, migration/integration state and async state can easily become shadow authorities. That produces best-effort Audit, fake native history, export snapshots that never existed, provider truth masquerading as product truth, and cross-owner partial commits.

Target invariant:

> **Every business/system fact has one semantic owner. Required Audit is atomically appended with the mutation but never owns that mutation. Historical Migration preserves source truth through explicit imported target-owner forms without fabricating native actions. External import/publish/export preserves exact provenance/content identity without storage coupling. Every local cross-owner invariant commits once through composition under one compatible lock order; external effects are durable intents followed by explicit receipts.**

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

### A — global event sourcing / Audit as history authority
Reject. Duplicates every domain authority and makes migration/recovery depend on one transversal log.

### B — best-effort Audit + generic integration/outbox engine
Reject. Allows governed success without evidence and collapses provenance/effect intent/process truth.

### C — DB triggers infer semantic Audit from CRUD
Reject. Persistence writes do not reliably encode Submission/Approval/Release/hold meaning.

### D — small semantic Audit + specialized Interchange contracts + explicit tx/lock matrix
**Recommended Global Maximum.** Smallest structure preserving auditability, historical truth, export integrity and local atomicity without a generic transversal platform.

Method outcome if accepted: **RESTRUCTURE NOW at design level**.

---

# 5. Audit model

## 5.1 AuditChainHead

```text
AuditChainHead
  singleton_key SMALLINT PRIMARY KEY CHECK singleton_key=1
  last_sequence BIGINT NOT NULL CHECK last_sequence >= 0
  last_hash BYTEA NOT NULL CHECK octet_length(last_hash)=32
```

Genesis = sequence 0 + 32 zero bytes.

## 5.2 AuditEvent

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

Actor XOR:

```text
USER   → actor_user_id only
SYSTEM → system_actor_code only
```

`system_actor_code`, `operation_code`, `resource_kind` and `facts_schema` are product-owned closed/bounded vocabularies at the admitted schema version. Facts schema defines exact keys/types; unknown arbitrary facts fail validation.

`resource_kind/resource_id` intentionally has no FK. Audit attribution is non-authoritative and the resource row/payload may later be lawfully disposed.

`occurred_at` is trusted server/application transaction time, never a client-supplied historical timestamp.

## 5.3 Hash

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

Cross-runtime golden vectors are mandatory.

## 5.4 DB-enforced append seam

Target DB privilege posture:

```text
ordinary serving role:
  NO direct INSERT/UPDATE/DELETE on AuditEvent
  NO direct UPDATE on AuditChainHead
  SELECT only where backend query path requires it
  EXECUTE only on narrow Audit-owned append primitive
```

The append primitive executes inside the caller's existing transaction and:

1. locks the singleton head;
2. obtains previous sequence/hash;
3. enforces next committed sequence;
4. accepts the bounded event envelope/canonical bytes from the Audit application seam;
5. inserts immutable event;
6. advances head monotonically;
7. exposes no reset/backfill API to ordinary serving trust.

Exact SQL-function vs equivalently strong database-owned primitive is implementation-spec work. If a SECURITY DEFINER function is chosen, ownership/search-path/PUBLIC EXECUTE must follow PostgreSQL safety guidance. Integrity validation recomputes the canonical hash independently; a malformed supplied hash cannot silently become a valid chain.

Multiple required events in one business transaction append in stable caller-defined semantic order while holding the head once.

## 5.5 Terminal lock rule

Audit append happens only after all owner locks/writes and required durable intents:

```text
... owner work
→ durable intent insert(s)
→ AuditChainHead
→ Audit event(s)
→ COMMIT
```

**No owner/domain lock may be acquired after AuditChainHead.**

Rollback removes business mutation + Audit events + head advance together.

---

# 6. Audit privacy / retention

Immutable skeleton classification:

| Data | Treatment |
|---|---|
| event id/sequence/time/hash/schema | non-PII technical/governance evidence |
| operation/resource codes | product codes |
| actor/subject User UUID | PII-minimized stable pseudonymous skeleton allowed to survive UserProfile erasure |
| role/scope/Area/Tenant/assignment ids, outcome/digest codes | bounded non-human-readable facts |
| User name/email/username/profile | forbidden from immutable Audit |
| historical source actor free text | imported retained payload, not Audit |
| approval/comment/return/disposition reason text | remains owning-domain evidence; Audit stores its evidence-id/outcome only |
| raw fresh-auth/JWT/password/token/provider claims | forbidden |
| IP/user-agent/request bodies/headers | telemetry/security logging, not semantic Audit V1 |

Read UI may resolve current UserProfile; when erased/unavailable, show stable opaque User identity.

No immutable erasable enrichment copy is introduced V1.

Audit skeleton retention V1:

```text
Indefinite
```

This is a product choice, not a claimed legal term. Finite retention would require explicit chain pruning/checkpoint semantics and therefore reopens B6.

Audit query/export never answers canonical resource state by “latest event”; owner reads remain mandatory.

---

# 7. B3 bounded refinements from historical truth

## 7.1 B3-R4 — RevisionOrdinalReservation

Real historical ordinal without sufficient exact bytes cannot become a Revision but cannot be reused.

```text
RevisionOrdinalReservation
  id UUID PRIMARY KEY
  document_id UUID NOT NULL FK Document(id) RESTRICT
  revision_no INTEGER NOT NULL CHECK revision_no >= 1
  reserved_at TIMESTAMPTZ NOT NULL
  UNIQUE(document_id,revision_no)
```

Database guard rejects any same `(document_id,revision_no)` across `DocumentRevision` and reservation.

Native next ordinal:

```text
max(DocumentRevision ordinals ∪ Reservation ordinals) + 1
```

Permanent minimal identity only; no fake lifecycle.

## 7.2 B3-R5/R6 — imported Revision identity/content/governance

Permanent skeleton adds:

```text
DocumentRevision.history_kind TEXT NOT NULL CHECK NATIVE|IMPORTED
```

`NATIVE` continues ordinary B3 rules. Historical Migration creates `IMPORTED` only through privileged CI seam.

### RevisionImportedContent

```text
RevisionImportedContent
  revision_id UUID PRIMARY KEY FK DocumentRevision(id) RESTRICT
  primary_artifact_id UUID NOT NULL FK Artifact(id) RESTRICT
  manifest_schema TEXT NOT NULL
  manifest_payload JSONB NOT NULL CHECK jsonb_typeof(...)='object'
  content_digest BYTEA NOT NULL CHECK octet_length(content_digest)=32
  adopted_at TIMESTAMPTZ NOT NULL
```

Digest = `SHA256(UTF8(manifest_schema) || 0x00 || RFC8785_JCS(manifest_payload))`.

Manifest freezes exact Artifact hash/size/format/media type and trustworthy imported governed/structured metadata. Unknown historical fields remain unknown; current dictionary/system values are never resolved merely to fill history.

`RevisionDictionarySnapshot` remains mandatory for `history_kind=NATIVE`; for `IMPORTED`, it exists only when trustworthy historical dictionary state can actually be represented. Otherwise imported manifest/governance explicitly records the absence/unknown.

### RevisionImportedGovernanceSnapshot

```text
RevisionImportedGovernanceSnapshot
  revision_id UUID PRIMARY KEY FK DocumentRevision(id) RESTRICT
  source_system_code_snapshot TEXT NOT NULL
  source_object_id_snapshot TEXT NULL
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

Bounded versioned schemas apply to actor/governance snapshots.

Target `DocumentRevision.state` remains CI authority. Historical source state/actors/times are evidence snapshots only.

No fake ApprovalInstance/Decision/ReleaseRecord/internal User action.

### Imported state law

Historical `IMPORTED` revisions do not use target `SUBMITTED`, because that state asserts a native Submission boundary. Historical draft/submitted-like source material remains source evidence or enters ordinary import/native DRAFT if it needs live editing.

V1 historical target states admitted by privileged migration are terminal/current governed states that can be truthfully proven (`EFFECTIVE|SUPERSEDED|OBSOLETE|CANCELLED`). Any extra state mapping is a reopen/explicit R10-F rule, not silent coercion.

### B5-R2 native timestamp refinement

`DocumentRevision.cancelled_at/obsoleted_at` mean **native MetalDocs transition times**:

```text
history_kind=NATIVE + CANCELLED → cancelled_at required
history_kind=NATIVE + OBSOLETE  → obsoleted_at required
```

For an imported Revision already terminal at adoption, the native timestamp may remain NULL while imported governance carries trustworthy source terminal evidence. If an imported current Revision later undergoes a native post-adoption cancel/obsolete transition, the corresponding native timestamp is written once.

Supersession caused later by native B4 Release always uses `ReleaseRecord.released_at`.

## 7.3 B3-R1 extension — imported Template source

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

Native kind requires Submission-id snapshot; imported kind requires it NULL. Creation still serializes against current template effectivity and validates the exact source content authority.

## 7.4 Imported Periodic Review / retention

Imported EFFECTIVE:

```text
trustworthy source effective/review anchor known → normal review due calculation
anchor unknown                              → preserve unknown + immediately DUE
```

DUE never invalidates EFFECTIVE. First native PeriodicReviewRecord establishes the next ordinary review anchor.

Imported Revision with exact content creates B5 RetentionBinding in the migration item transaction despite no native Submission.

Retention anchor:

- current EFFECTIVE → no running clock;
- later native supersession → B4 ReleaseRecord time;
- already-historical imported state → trustworthy imported lifecycle timestamp only;
- unknown imported historical anchor → no silent disposition eligibility.

---

# 8. B5 bounded refinement — historical Evidence

Permanent skeleton adds:

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

CAPTURED Evidence has exactly one `EvidenceCapture` for NATIVE or `EvidenceImportedCapture` for IMPORTED.

Migration atomic unit creates Evidence identity + exact Artifact root + imported capture + frozen primary Dossier/name/sequence + RetentionBinding + applicable active hold materialization.

No fake `captured_by_user_id` or fake historical `captured_at`.

Retention:

```text
CAPTURED_AT → trustworthy source_captured_at only
OCCURRED_AT → trustworthy occurred_at only
unknown     → not silently disposition-eligible
```

`adopted_at` is provenance, not substitute historical time.

---

# 9. Historical Migration model

## HistoricalMigrationSource

```text
id UUID PK
code TEXT UNIQUE immutable
name TEXT
source_kind TEXT
status ACTIVE|INACTIVE
created_at TIMESTAMPTZ
```

Stable source namespace, no credentials/endpoints/secrets.

## HistoricalMigrationPlan + Item

Plan is created as a complete immutable unit; there is no selectable partially built plan.

```text
HistoricalMigrationPlan
  id UUID PK
  source_id UUID FK HistoricalMigrationSource RESTRICT
  mode CURRENT_STATE|FULL_HISTORY
  plan_schema TEXT
  plan_digest BYTEA32
  created_by_user_id UUID FK User RESTRICT
  created_at TIMESTAMPTZ

HistoricalMigrationPlanItem
  id UUID PK
  plan_id UUID FK Plan RESTRICT
  item_order BIGINT >= 1
  source_entity_kind TEXT
  source_entity_id TEXT
  source_fingerprint BYTEA32
  target_kind DOCUMENT|DOCUMENT_REVISION|REVISION_ORDINAL|EVIDENCE|DOSSIER
  target_hint JSONB bounded object
  UNIQUE(plan_id,item_order)
  UNIQUE(plan_id,source_entity_kind,source_entity_id)
```

Plan digest uses a versioned canonical ordered descriptor list with JCS + SHA-256.

## Execution / Outcome

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

Transient attempts/errors/retries stay R10-D.

## HistoricalSourceBinding

Cross-plan source idempotency:

```text
HistoricalSourceBinding
  id UUID PK
  source_id UUID FK HistoricalMigrationSource RESTRICT
  source_entity_kind TEXT
  source_entity_id TEXT
  source_fingerprint BYTEA32

  document_id UUID NULL FK Document RESTRICT
  document_revision_id UUID NULL FK DocumentRevision RESTRICT
  revision_ordinal_reservation_id UUID NULL FK RevisionOrdinalReservation RESTRICT
  evidence_id UUID NULL FK Evidence RESTRICT
  dossier_id UUID NULL FK Dossier RESTRICT

  bound_at TIMESTAMPTZ
  UNIQUE(source_id,source_entity_kind,source_entity_id)
```

Exactly one target FK.

Same source identity + same fingerprint/coherent target = REUSE. Different/conflicting fingerprint/state = fail closed. No heuristic merge.

---

# 10. Dry-run / APPLY law

### DRY_RUN

Uses the same validation contracts as APPLY for source fingerprints, mapping, formats, uniqueness, target eligibility and retention preconditions, but creates **zero target semantic rows, Artifacts, RetentionBindings or domain governance facts**. It persists only the Interchange dry-run execution/outcomes.

### APPLY per item

1. re-read/verify source fingerprint against finalized plan;
2. existing matching source binding + coherent target → REUSE;
3. conflict → fail item closed;
4. stage exact bytes outside DB transaction where necessary;
5. open one target semantic transaction;
6. call privileged target-owner migration seam;
7. create target imported content/governance/capture + RetentionBinding/holds where required;
8. confirm Artifact with one semantic retention root in same transaction;
9. create HistoricalSourceBinding;
10. insert final APPLY item outcome;
11. insert any required durable intent;
12. append Audit event(s) last;
13. commit.

Partial batch success is valid and reconciled per item. No whole-batch rollback promise.

Migration/System is native Audit actor for actual imported mutation. Source actor remains imported retained provenance only.

---

# 11. Governed Subject Export

## 11.1 Subject request

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

## 11.2 Complete closure semantics

`COMPLETE` V1 is explicit and non-recursive beyond the following bounded graph:

### Document export

Includes:

- stable Document identity/config provenance required to interpret it;
- every Revision identity skeleton for the Document, including ordinal reservations;
- every **currently present** retained Revision unit payload/history from B3/B4/B5;
- DispositionRecord/records skeleton where a historical payload was already lawfully disposed;
- exact current/present Artifact bytes referenced by included retained units;
- relevant Template origin and Dossier relationship metadata as references, but does **not** recursively export unrelated Dossier contents.

A disposed Revision is represented truthfully as identity/disposition evidence with no invented/missing bytes.

### Evidence export

Includes Evidence identity + currently present native/imported capture payload/Artifact + records/disposition evidence + primary/secondary Dossier relationship identities. It does not recursively export all other Dossier contents.

### Dossier export

Includes Dossier identity/provenance + all currently linked Documents (`DossierDocumentLink`) + all Evidence for which the Dossier is primary or secondary, each using its bounded Document/Evidence closure above. Traversal stops there; no Dossier-to-Dossier/transitive context graph is invented.

If any required included subject is not authorized/readable under canonical AuthZ, the COMPLETE export fails closed.

No PARTIAL export V1.

## 11.3 Stable snapshot — narrow isolation exception

One short PostgreSQL `REPEATABLE READ` transaction:

1. resolve exact complete closure under one DB snapshot;
2. apply canonical AuthZ to every required subject;
3. enumerate identities/relationships/provenance/hashes without provider locations;
4. build canonical manifest;
5. insert `GovernedSubjectExport` + immutable snapshot;
6. insert package-build durable intent;
7. append Audit last;
8. commit.

No byte copying while the DB transaction is open. Serialization/retry failure is visible/retryable, not reason for global isolation change.

## 11.4 Snapshot / receipt

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

Manifest includes package version, root, objects, relationships, provenance, canonical filenames, formats/media types, sizes, SHA-256, content/governance digests and `completeness=COMPLETE`. No provider bucket/key/URL/version identity.

Manifest digest = JCS + SHA-256.

Package output is temporary delivery output, not semantic Artifact/Evidence. Before receipt, package builder verifies every emitted file against snapshot hashes. Temporary location/expiry/download token is R10-D/E.

Export never changes LegalHold/retention or counts as Distribution acknowledgement.

---

# 12. External Repository Interchange

## InterchangeConnection

```text
InterchangeConnection
  id UUID PK
  code TEXT UNIQUE immutable
  name TEXT
  connector_kind TEXT product-supported bounded code
  status ACTIVE|INACTIVE
  created_at TIMESTAMPTZ
```

Logical connection identity only. Credentials/tokens/endpoints/secrets live in deployment/provider config keyed by connection id. If endpoint configuration changes the logical repository identity, create a new connection.

B5 `DossierExternalReference.connection_id` becomes FK here; `(connection_id,entity_kind,external_id)` remains unique.

## RepositoryTransfer

```text
RepositoryTransfer
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
  source_sha256 BYTEA NULL CHECK NULL OR octet_length(...)=32
  source_content_format TEXT NULL

  intended_target_kind TEXT NULL CHECK NULL OR DOCUMENT|EVIDENCE
  requested_target_document_id UUID NULL FK Document(id) RESTRICT
```

PUBLISH pins one internal source root + exact Artifact snapshot, no Artifact FK. IMPORT pins external source + intended ordinary target operation; optional existing Document target allows an ordinary new-revision/import workflow rather than a generic transformation engine.

Request transaction:

```text
RepositoryTransfer
+ external-effect durable intent
+ Audit
COMMIT
```

No request row is success proof.

### PUBLISH_COPY

Worker re-proves pinned Artifact UUID/hash/format still exists before provider write. If disposed/missing, transfer fails visibly; no resurrection/hidden retention.

### IMPORT_COPY

Worker fetches/stages/validates exact bytes, then invokes ordinary target owner operation under current authorized semantics. Connector identity does not bypass Document/Evidence governance.

## RepositoryTransferReceipt

```text
RepositoryTransferReceipt
  transfer_id UUID PK FK RepositoryTransfer RESTRICT
  external_entity_kind TEXT
  external_entity_id TEXT
  external_version_snapshot TEXT NULL
  target_document_id UUID NULL FK Document RESTRICT
  target_document_revision_id UUID NULL FK DocumentRevision RESTRICT
  target_evidence_id UUID NULL FK Evidence RESTRICT
  completed_at TIMESTAMPTZ
```

For IMPORT_COPY, target semantic creation/revision/capture **and Receipt commit in the same local target transaction**. If target creation commits, receipt cannot be missing for a transfer claimed successful.

For PUBLISH_COPY, only external provider confirmation creates receipt.

No receipt = no semantic success claim. Provider URLs remain display/mechanism data.

---

# 13. Required semantic Audit classes

Audit does not log everything. Same-commit V1 Audit is mandatory for:

### B2
Tenant/Area/User lifecycle/settings; governed UserProfile mutation/erasure; binding acceptance/replacement/disable/erasure where admitted; admin Session revoke/offboard; Group create/rename/delete; membership add/remove; RoleAssignment grant/revoke.

Grant/revoke facts preserve assignment id, subject stable id, role, scope and operation without human profile duplication.

### B3/B4
Document creation; responsibility change; Revision creation when materially entering governance; SUBMIT; Revision cancel/obsolete; PeriodicReviewRecord; ApprovalStepDecision; approval cancel/withdraw/reassign; Rendition semantic success; Release; explicit Distribution acknowledgement.

### B5
Dossier create/archive/re-enable; Dossier↔Document link/unlink; Evidence CAPTURE/VOID; captured secondary-Dossier link/unlink; RetentionExtension; LegalHold activate/release; DispositionFence; completed DispositionRecord.

### B6
Historical Migration APPLY semantic outcome; accepted Governed Export snapshot; COPY request; successful COPY receipt.

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

A later requirement may promote a specific class; no generic “audit every row” rule.

---

# 14. Same-local-commit matrix

`Audit` below means required immutable event(s), appended last. `Intent` is same-commit durable mechanism only when future work is required.

| Operation | Atomic semantic set |
|---|---|
| User offboarding | User disable + local Session revoke + membership/direct-grant revoke + admitted binding state + Audit + provider-disable Intent |
| Document + REV001 | number allocation + Document + Revision + native RevisionDictionarySnapshot + WorkingContent + template adjuncts + Audit |
| first SUBMIT | WC generation consume + RevisionSubmission + ApprovalRequirement/ReleasePlan + Approval init + RetentionBinding + active Hold materialization + Audit + evaluation Intent where needed |
| same-REV resubmit | new Submission + new B4 per-Submission requirements/Approval init + Audit + evaluation Intent; existing Binding unchanged |
| Approval RETURN/CANCEL/WITHDRAW | Revision DRAFT return generation law + Approval terminal state + Audit |
| Approval ACCEPT | immutable decision + Step/Instance transition + Audit + final-gate evaluation Intent if needed |
| Rendition success | Artifact confirm + Rendition relation + Audit + release evaluation Intent |
| Release | candidate EFFECTIVE + predecessor SUPERSEDED + ReleaseRecord + DistributionObligations + Audit + notification/search Intents |
| Periodic Review | exact current-Revision record + Audit |
| native/imported Evidence CAPTURE | capture authority + Artifact root + Evidence frozen identity/context + RetentionBinding + applicable Holds + Audit |
| Dossier/secondary Evidence link entering hold | relation + necessary HoldSubject materialization + Audit |
| LegalHold activation | Hold + complete current materialization + Audit |
| RetentionExtension | extension + Audit |
| DispositionFence | fence + Audit + physical-delete Intent |
| Disposition completion | retained payload/Artifact semantic cleanup + DispositionRecord + Audit |
| Historical Migration APPLY item | target imported state/content/provenance + source binding + Binding/holds + item outcome + Audit |
| Governed Export snapshot | export root + immutable complete manifest + build Intent + Audit under REPEATABLE READ |
| COPY request | RepositoryTransfer + external-effect Intent + Audit |
| IMPORT_COPY success | ordinary target semantic creation/revision/capture + RepositoryTransferReceipt + Audit in one local target transaction |
| PUBLISH_COPY success | RepositoryTransferReceipt + Audit |

No provider call/object transfer is inside DB commit.

---

# 15. Composition / lock law

`internal/composition` owns no durable business meaning. Owner seams share a transaction, never commit independently, never import another owner's repository, and expose all local failure before Audit append.

Global target is a partial order, not a lock-manager framework.

## Main governed-content branch

When applicable:

```text
B2 eligibility roots according to promoted B2 sub-order
→ current configuration/allocation roots
  DocumentType / EvidenceType / Approval config / NumberSeries / EvidenceSequence
→ business roots
  Document(s) UUID order
  → Evidence(s) UUID order
  → Dossier(s) UUID order
→ child/execution
  Revision / WorkingContent / ApprovalInstance+Step / Evidence capture state
→ RetentionBinding(s) UUID order
→ Artifact existing-row coordination only when necessary
→ durable Intent inserts
→ AuditChainHead ALWAYS LAST
```

A transaction may begin later if it never acquires an earlier class.

## Group branch

Promoted B2 Group deletion remains:

```text
Group → memberships/group grants
```

and never later acquires governed Document/Evidence/Dossier roots. B4 may read/lock live Group dependencies after its Document/Approval root because the reverse path is prohibited, preserving DAG direction.

## Approval return

Any path that may return/cancel/withdraw a Revision locks Document/Revision before Approval execution rows. Pure intermediate ACCEPT may remain Approval-local and then Audit; it may not later acquire Document.

## Dossier/Records

```text
Document or Evidence root
→ relevant Dossier roots UUID order
→ RetentionBinding
→ Audit
```

Dossier-wide Hold activation may start at Dossier then acquire RetentionBindings deterministically; it must not backtrack to Document/Evidence after binding locks.

## Audit leaf

```text
NO owner/domain lock after AuditChainHead.
```

This is a proof obligation, not documentation convention.

---

# 16. Export isolation law

Product default remains `READ COMMITTED`.

Only short complete-export snapshot transaction uses `REPEATABLE READ` because many relationship reads must represent one real database snapshot without locking a large graph for package-build duration.

Transaction ends after manifest + build Intent + Audit. Package bytes are assembled after commit.

No other operation gets stronger isolation by default.

---

# 17. Persistence / mutation classes

| Family | Class | Mutation law |
|---|---|---|
| AuditChainHead | integrity-support semantic mechanism | narrow monotonic mutable head |
| AuditEvent | Audit semantic authority | immutable append-only |
| RevisionOrdinalReservation | CI semantic identity skeleton | permanent immutable |
| DocumentRevision.history_kind | CI semantic identity | immutable |
| RevisionImportedContent | CI imported content authority | immutable retention payload |
| RevisionImportedGovernanceSnapshot | CI imported governance authority | immutable retention payload |
| Evidence.history_kind | Documentary Context semantic identity | immutable |
| EvidenceImportedCapture | Documentary Context imported capture authority | immutable retention payload |
| HistoricalMigrationSource | Interchange config | code immutable; display/status mutable |
| HistoricalMigrationPlan/Item | Interchange plan | immutable complete plan |
| HistoricalMigrationExecution | Interchange process | constrained state machine |
| HistoricalMigrationItemOutcome | Interchange process evidence | immutable final per item/execution |
| HistoricalSourceBinding | source↔target provenance authority | immutable binding |
| GovernedSubjectExport | Interchange request | immutable |
| GovernedExportSnapshot | export semantic identity | immutable |
| GovernedExportPackageReceipt | export success evidence | immutable |
| InterchangeConnection | external logical identity | code immutable; name/status mutable |
| RepositoryTransfer | transfer request | immutable |
| RepositoryTransferReceipt | transfer success evidence | immutable |
| job/outbox/retry/lease/temp package location | mechanism | R10-D |

---

# 18. Proof obligations

Before implementation prove at minimum:

### Audit
1. serving trust cannot direct-write AuditEvent/reset head;
2. audited business rollback also rolls Audit/head back;
3. mutate/delete/reorder event → validator fails;
4. concurrent audited commits produce distinct contiguous committed sequences;
5. Go/TS/tool canonical hash goldens match;
6. mandatory operation census has no admitted uncovered path;
7. no audited path obtains a domain lock after Audit head;
8. Audit query cannot be used by canonical owner read paths to infer state.

### Imported history
9. missing-byte ordinal reservation prevents reuse;
10. imported Revision has no fake Submission/Approval/Release;
11. NATIVE dictionary snapshot required; imported unknown dictionary never resolved from current values;
12. imported terminal source state may retain NULL native terminal timestamp with imported proof;
13. unknown historical effectivity/capture remains unknown;
14. native successor Release supersedes imported EFFECTIVE normally;
15. imported Template origin works without fake Submission/indefinite source retention;
16. imported Evidence has no fake User/time;
17. unknown imported retention anchor never becomes eligible;
18. same source + same fingerprint reuses, changed fingerprint conflicts;
19. dry-run creates zero target semantic rows/Artifacts/Bindings.

### Export / transfers
20. Document/Evidence/Dossier closure matches the bounded COMPLETE contract;
21. any required unauthorized subject makes COMPLETE export fail closed;
22. disposed history exports truthful skeleton/disposition evidence, never missing-byte fiction;
23. concurrent writes cannot make the manifest span inconsistent snapshots;
24. package builder verifies each exact hash before receipt;
25. manifest/package has no provider storage identifiers;
26. no COPY receipt means no external success claim;
27. IMPORT target semantic mutation + receipt is atomic locally;
28. IMPORT cannot bypass normal target governance;
29. disposed/missing PUBLISH source fails visibly without resurrection.

### Transactions / concurrency
30. whole B1–B6 wait-for graph is acyclic under all admitted paths;
31. same-class multi-row locks use deterministic order;
32. no nested owner commit produces partial local success;
33. each matrix row has a negative probe proving local rollback across all semantic participants;
34. required Intent insert failure rolls back the business mutation;
35. provider failure after commit leaves domain truth intact and recoverable through R10-D.

---

# 19. Adversarial findings closed by candidate

- **Audit as hidden event sourcing:** forbidden owner-read dependency; Audit only attribution/evidence.
- **Indefinite Audit stores PII:** bounded schemas prohibit human profile/free text; stable UUID skeleton only.
- **Audit chain direct-write bypass:** serving trust targets narrow DB append primitive, not direct event/head writes.
- **Global head deadlock:** Audit is terminal leaf; measured throughput pressure is reopen trigger, not speculative segmentation.
- **Migration fakes native history:** explicit history kind/imported content/governance forms.
- **CI depends on Interchange forever:** ongoing imported facts are target-owner snapshots.
- **Historical ordinal reused:** permanent reservation participates in allocation.
- **Imported terminal timestamp fabricated:** native timestamp means native transition; historical source time stays imported evidence.
- **Imported Evidence invents captured-by/time:** dedicated imported capture.
- **Export is inconsistent:** short REPEATABLE READ manifest snapshot.
- **Export closure ambiguous:** explicit Document/Evidence/Dossier bounded COMPLETE closure.
- **Export secretly becomes hold:** no hidden preservation; race may fail package visibly.
- **COPY request treated as success:** immutable receipt only.
- **IMPORT target commits without receipt:** target mutation + receipt same local transaction.
- **Generic integration registry:** closed typed unions; four frozen contracts remain separate.
- **Audit reverse lock edge:** no owner call/lock after head acquisition.

---

# 20. Essential vs accidental complexity / reopen

Essential:

```text
same-commit bounded Audit
append tamper chain
PII-minimized skeleton
native vs imported truth
exact imported content/capture authority
historical ordinal preservation
true dry-run/idempotency/reconciliation
complete stable export manifest
explicit external-effect receipts
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

Reopen only on material evidence: measured Audit-head contention; finite/erasure requirement for Audit skeleton; external non-repudiation contract; standardized/partial export consumer; new truthful imported-state requirement; new retention subject in migration; continuous bidirectional sync requirement; cross-repository trust boundary; lock-graph proof failure; or real workload proving the export isolation solution insufficient.

Implementation inconvenience/current schema/hypothetical integrations do not reopen.

---

# 21. Candidate decision / next state if accepted

```text
R10-B3 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
         + B5 refinements B3-R1/R2/R3
         + B6 refinements B3-R4/R5/R6
         + B3-R1 imported-content extension
         + B3-R2 native-vs-imported terminal timestamp clarification

R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL

R10-B5 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
         + imported-Evidence bounded refinement

R10-B6 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED

R10-B = INTEGRATED DESIGN BLOCK COMPLETE / NON-FINAL
implementation = BLOCKED
next = R10-C — Artifact / Records Physical Integrity
```

Before operator acceptance, run one integrated B1↔B6 coherence review of this corrected candidate. Whole-R10 cold independent review remains deferred until integrated B6/C/D/E/F unless a material exception trigger appears.