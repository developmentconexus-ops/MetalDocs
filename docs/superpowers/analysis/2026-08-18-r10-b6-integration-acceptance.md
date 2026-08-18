# R10-B6 — Integration Acceptance / Operator Adjudication

> **Status:** ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED

This record captures operator acceptance of:

`docs/superpowers/analysis/2026-08-18-r10-b6-audit-interchange-cross-owner-atomicity-integrated-candidate.md`

Acceptance means **working authority for continued R10 integration**, not final ratification. B3–B6 remain challengeable only by a material later-stage counterexample. Whole-R10 Global Coherence Review + cold independent review remain required before final R10 ratification.

---

## 1. Accepted Audit target

- `AuditEvent` is transversal forensic/timeline evidence and never canonical domain/resource state.
- required governed mutations append Audit in the **same local PostgreSQL transaction** as the business/system mutation; best-effort post-commit audit is not the target.
- one deployment-wide `AuditChainHead` orders committed events and supports a small append hash chain using RFC 8785 JCS + SHA-256.
- ordinary serving trust has no direct `INSERT/UPDATE/DELETE` on `AuditEvent` and no direct `UPDATE` on the chain head; Audit exposes only a narrow DB-owned append primitive usable inside the caller transaction.
- `AuditChainHead` is the final contended semantic lock of audited transactions; no domain lock may be acquired afterwards.
- the immutable Audit skeleton contains only stable IDs/codes/times/digests/outcomes and bounded product-owned facts schemas; free-form domain reasons/comments/request bodies/profile data/provider claims/secrets do not enter indefinite immutable Audit.
- User UUID may survive as a PII-minimized pseudonymous skeleton; human-readable UserProfile data is resolved at read time and remains separately erasable.
- Audit skeleton retention V1 is `Indefinite` as a product choice, not a claimed statutory period. Finite pruning/checkpoint semantics are a reopen trigger.
- Audit is not event sourcing, telemetry, outbox/durable intent, business state, retention-clock authority, Approval authority or Release/effectivity authority.

---

## 2. Accepted Interchange target

Interchange remains four distinct contracts rather than one generic integration engine:

```text
Historical Migration
Governed Subject Export
External Repository IMPORT_COPY
External Repository PUBLISH_COPY
```

Backup/Restore remains separate and Tenant Portability Export remains deferred.

### Historical Migration

- source history is never rewritten as synthetic native MetalDocs actions;
- source Approval/effectivity/actor facts are imported target-owned governance provenance, never fake `ApprovalInstance`, `ApprovalDecision`, `ReleaseRecord` or native User action;
- source unknowns remain unknown;
- migration persists plan + true dry-run/apply runs + deterministic per-item outcomes + reconciliation;
- APPLY is atomic per explicit **semantic import unit**, not per arbitrary row and not whole-batch;
- creating a new historical Document is atomic with at least one exact imported Revision/content unit so no empty governed Document shell can survive a failed first revision;
- partial batch success is valid and reconciled;
- persistent cross-plan idempotency uses a stable SHA-256 source-identity digest rather than requiring long-lived raw/human-readable source IDs;
- Migration/System principal is the native actor of the import operation; source human actors remain imported provenance only.

### Governed Subject Export

- export snapshot is one short `REPEATABLE READ` transaction so a `COMPLETE` manifest describes one consistent database state;
- package assembly happens after commit from exact immutable identities/hashes; no long DB transaction spans byte assembly/download;
- provider-independent manifest contains exact objects/relationships/provenance/canonical filenames/ContentFormats/sizes/SHA-256 and uses JCS + SHA-256;
- complete export fails closed if a required subject cannot lawfully be included; V1 has no masquerading partial export;
- generated package is temporary delivery output and is not automatically Artifact/Evidence/retention subject.

### External Repository copies

- `InterchangeConnection` is the stable logical repository identity; credentials/endpoints remain provider/deployment mechanism state;
- `PUBLISH_COPY` pins exact source identity/hash/format and records success only through immutable provider receipt after external confirmation;
- `IMPORT_COPY` uses ordinary target-owner lifecycle/permissions and commits target semantic creation plus immutable import receipt in one local transaction;
- provider URL/path/display name is never used as stronger business identity when a stable provider object identity exists.

---

## 3. Accepted B3/B5 bounded refinements exposed by B6

No unrelated B3/B4/B5 semantic is reopened.

### B3-R4 — RevisionOrdinalReservation

A historically real revision ordinal with insufficient exact bytes may not create a fake `DocumentRevision`, but the ordinal must still never be reused.

```text
RevisionOrdinalReservation
  document_id
  revision_no
```

Native next revision is above every `DocumentRevision` and reservation ordinal.

### B3-R5 — RevisionImportedContent

A historical Revision with exact bytes/content may exist without any native MetalDocs SUBMIT attempt. Therefore imported exact content is represented by immutable CI-owned `RevisionImportedContent`, never a fabricated `RevisionSubmission`.

### B3-R6 — imported governance/history kind

`DocumentRevision.history_kind = NATIVE | IMPORTED` distinguishes provenance without changing the Revision lifecycle vocabulary. Imported lifecycle/governance proof is CI-owned immutable imported governance state, not native Approval/Release history.

### B3-R1 imported-template extension

Template origin may pin exact source kind `NATIVE_SUBMISSION | IMPORTED_REVISION_CONTENT` through immutable source Revision/content identity/digest/hash snapshots. It does not restore retention-blocking strong FKs to disposable source payload.

### B3-R2 native/imported terminal timestamp clarification

`cancelled_at` / `obsoleted_at` are **native MetalDocs transition instants**. An already-terminal imported Revision may retain NULL native transition time while imported governance state preserves the truthful historical source timestamp/unknown.

### B5 imported Evidence refinement

`Evidence.history_kind = NATIVE | IMPORTED`; historical Evidence uses target-owned immutable imported capture/provenance state rather than fake native `captured_by/captured_at`. Imported Evidence receives its RetentionBinding in the migration semantic unit; unknown historical anchor never silently becomes disposition-eligible.

### Imported dictionary rule

`RevisionDictionarySnapshot` remains mandatory for native revisions. Imported history never resolves current Tenant Dictionary merely to manufacture a historical snapshot that did not exist.

---

## 4. Accepted cross-owner atomicity model

B6 closes the R10-B same-local-commit model.

Canonical shape:

```text
composition opens one PostgreSQL transaction
→ published owner seams share the transaction
→ all frozen domain facts commit or roll back together
→ required durable async intents are inserted before Audit
→ required Audit event(s) append last
→ one COMMIT / ROLLBACK
```

No semantic owner imports another owner's repository or hides a nested commit to obtain atomicity. Provider/object-store/network effects never join the local DB transaction.

Material same-commit classes include at minimum:

- B2 administrative access/identity mutations + required Audit + provider-effect intents where applicable;
- Document creation / first Revision initialization;
- SUBMIT + Submission + ApprovalRequirement/ReleasePlan + first RetentionBinding/hold materialization + required Audit/intents;
- approval return/cancel/accept state transitions + required Audit/intents;
- Rendition semantic confirmation + Artifact/Rendition + required Audit/release-evaluation intent;
- winning Release + EFFECTIVE/SUPERSEDED + ReleaseRecord + Distribution obligations + Audit + projection/notification intents;
- Evidence CAPTURE + exact Artifact + immutable capture + RetentionBinding/holds + Audit;
- Dossier link mutations that enter active hold scope + materialization + Audit;
- LegalHold / RetentionExtension / DispositionFence + Audit and relevant durable intent;
- disposition semantic completion + retained-payload/Artifact semantic cleanup + DispositionRecord + Audit;
- Historical Migration semantic unit + target-owner imported state + retention/hold + reconciliation outcome + Audit;
- Governed Export immutable snapshot + Audit + package-build intent;
- IMPORT_COPY target creation + receipt + Audit;
- PUBLISH_COPY request truth + Audit + external-effect intent; external success receipt + Audit.

---

## 5. Accepted lock/isolation posture

Default product isolation remains `READ COMMITTED`.

One explicit bounded exception is accepted:

```text
Governed Subject Export snapshot → short REPEATABLE READ transaction
```

because a complete multi-statement export manifest must represent one consistent database snapshot without locking the whole subject graph during package assembly.

Global partial order remains owner/operation-specific but must never backtrack. The terminal invariant is:

```text
B2 eligibility/configuration roots
→ business subject roots
→ owner execution state
→ Dossier/context roots where required
→ RetentionBinding / Artifact coordination
→ durable intent insert
→ AuditChainHead
→ Audit append
→ COMMIT
```

Equivalent rows acquired together use deterministic ordering such as ascending UUID. Before implementation, the complete B1→B6 lock graph must be mechanically demonstrated acyclic under admitted write paths.

---

## 6. R10-B stage closure

With this acceptance:

```text
R10-B1 = CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED
R10-B2 = CLOSED / APPROVED / INTEGRATED
R10-B3 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R10-B4 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R10-B5 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R10-B6 = ACCEPTED FOR R10 INTEGRATION / NON-FINAL / NOT INDEPENDENTLY RATIFIED

R10-B = INTEGRATED DESIGN BLOCK COMPLETE / NON-FINAL
```

B6's self-review included the required internal whole-B coherence challenge. This does not replace the later Whole-R10 Global Coherence Review or cold independent review before final ratification.

---

## 7. Exact next step

```text
R10-C — Artifact / Records Physical Integrity = NEXT / DESIGN ONLY
implementation = BLOCKED
```

R10-C must consume the completed non-final R10-B model and close physical storage/integrity semantics without taking business authority from Artifact/Controlled Information/Records Governance.

At minimum it must cover:

- ManagedArtifactStore physical-location model and provider-neutral conformance;
- staging → malware inspection → semantic confirmation boundary;
- exact-byte verification on write/read/relocation/restore;
- storage relocation without changing Artifact identity or Submission digest;
- AWS S3/reference-production posture and local dev/test adapter parity where relevant;
- Object Lock/WORM interaction as enforcement of Records Governance, never policy authority;
- physical disposition/delete verification after B5 DispositionFence;
- orphan staging/failed-render/export cleanup as mechanism state;
- backup/restore integrity and non-resurrection constraints;
- provider failure/retry/reconciliation seams routed to R10-D where execution semantics belong.

Implementation remains **BLOCKED**.