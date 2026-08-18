# MetalDocs Launch V1 — Records Governance Defer Rebaseline

> **Status:** OPERATOR-APPROVED CURRENT-R10 OVERLAY / LAUNCH-SCOPE REBASELINE  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED until the shortened R10-C/D/E/F + Whole-R10 review/ratification complete

This record preserves the previously accepted B5/B6 Records-Governance design as **historical design knowledge**, but removes that capability from the **Launch V1 implementation target**. It is an explicit bounded current-R10 overlay; it does not silently rewrite the frozen R9.5 ledger or delete the prior B5/B6 candidates/acceptance records.

## 1. Trigger / Method outcome

Operator challenge: the Launch V1 product needs to become usable quickly, and there is no current evidenced launch requirement for finite records retention, Legal Hold, or lawful governed-content disposition.

DevelopmentConexus Method result:

```text
Retention / LegalHold / governed disposition in Launch V1
→ BOUNDED REOPEN
→ DEFER SAFELY
```

Reason:

- no named launch customer/regulatory/contractual requirement currently requires finite retention or governed destruction;
- if Launch V1 exposes no governed-content deletion, LegalHold adds no immediate preservation property because governed history is already preserved;
- retention clocks/extensions primarily feed a disposition capability that Launch V1 will not expose;
- implementing the full Records-Governance subsystem now would add tables, permissions, APIs, jobs, physical-delete protocols and UI without a current consumer;
- stable `DocumentRevision`, `Evidence`, `Artifact` and lifecycle identities already provide a clean seam to add Records Governance later without dismantling the current kernel.

This follows the Method's YAGNI law: prepare the seam, not the entire future capability.

## 2. Launch V1 governing invariant

> **MetalDocs Launch V1 preserves confirmed governed history and does not offer governed physical disposition/deletion. Business lifecycle states such as SUPERSEDED, OBSOLETE, CANCELLED and VOIDED never imply physical deletion. Only temporary/mechanism data is eligible for ordinary garbage collection.**

Launch V1 therefore distinguishes:

```text
GOVERNED / SEMANTIC HISTORY
  DocumentRevision
  RevisionSubmission / imported exact Revision content
  Approval / feedback / fresh-auth decision evidence
  Rendition / Release
  Distribution / Acknowledgement
  Evidence capture/imported capture
  Dossier context
  confirmed Artifact bytes
  Audit
  imported-history provenance

  → preserved; no governed physical delete operation V1

TEMPORARY / MECHANISM STATE
  failed/unconfirmed staging uploads
  abandoned render/export temporary output
  cache/preview scratch data
  retry/lease/job/outbox mechanism state subject to its own operational retention
  orphaned provider staging that never became confirmed governed content

  → bounded GC/reconciliation allowed
```

## 3. Launch ownership topology refinement

Previously accepted R10-A carried eight business bounded contexts including `Records Governance`.

Current Launch V1 working target has **seven active business bounded contexts**:

```text
Authentication
Organization
Authorization
Controlled Information
Approval
Documentary Context
Distribution
```

Supporting semantic owners remain:

```text
Artifact
Audit
Interchange
```

`Records Governance` is **DEFERRED FUTURE CAPABILITY**, not an empty launch bounded context/module.

Reopen trigger: a concrete regulatory, contractual, customer or operational requirement for finite retention, legal preservation hold, or governed disposal/destruction.

## 4. Authorization catalog refinement

Remove these dormant Launch V1 permissions:

```text
retention.extend
legal_hold.manage
disposition.manage
```

Current Launch V1 semantic permission count becomes:

```text
43 - 3 = 40
```

No role is added. The future Records-Governance capability will add permissions only when the capability returns.

All other B2 AuthN/Organization/AuthZ semantics remain unchanged.

## 5. B5 current Launch V1 target

B5 is reinterpreted for Launch V1 as:

```text
Documentary Context
+ Evidence
+ Artifact semantic ownership closure
```

Keep:

- `DossierType` / `Dossier`;
- Dossier↔Document contextual M:N relationship;
- `EvidenceType` / `Evidence` small `DRAFT → CAPTURED → VOIDED` lifecycle;
- exact primary Dossier and optional secondary contextual links;
- Evidence exact Artifact ownership/capture immutability;
- ExternalReference/provenance seams;
- one semantic Artifact root: one DocumentRevision or one Evidence;
- no generic `owner_type/id`, generic Record, folder ownership, ACL inheritance or ref-count authority.

Defer completely from Launch V1 implementation:

```text
DocumentTypeRetentionRule
EvidenceTypeRetentionRule
RetentionBinding
RetentionExtension
LegalHold
LegalHoldSubject
DispositionFence
DispositionRecord
retention clocks / expiry eligibility
governed physical delete workflow
ObjectLock/WORM mapping driven by Records Governance
eDiscovery/custodian preservation machinery
```

No disabled placeholder tables/flags are created for these deferred capabilities.

## 6. B3 refinements after the defer

### 6.1 KEEP — native lifecycle instants

Keep:

```text
DocumentRevision.cancelled_at
DocumentRevision.obsoleted_at
```

They are legitimate domain lifecycle facts independent of retention. Native supersession time remains B4 ReleaseRecord authority.

### 6.2 SIMPLIFY — dictionary snapshot

The B5-induced separation:

```text
RevisionDictionarySnapshot
```

was introduced primarily so dictionary payload could be disposed while permanent Revision identity survived.

With no governed disposition in Launch V1, revert the Launch working target to the smaller B3 shape:

```text
DocumentRevision.dictionary_snapshot JSONB immutable
```

No separate `RevisionDictionarySnapshot` table in Launch V1.

### 6.3 SIMPLIFY — DocumentOrigin

The B5 retention-safe source snapshots were primarily needed so a source Submission/Artifact retention unit could later be deleted.

With no governed disposal, Launch V1 may use a closed typed exact source reference:

```text
DocumentOrigin
  derived_document_id
  source_kind NATIVE_SUBMISSION | IMPORTED_REVISION_CONTENT
  source_submission_id FK RevisionSubmission NULL
  source_imported_content_revision_id FK DocumentRevision NULL
```

Closed-union validation ensures exactly one source kind/reference. Native source points to exact `RevisionSubmission`; imported source points to exact CI-owned imported Revision content identity. No generic polymorphic owner registry.

If future Records Governance requires source payload deletion, provenance-surviving snapshot semantics reopen then.

### 6.4 KEEP — imported-history truth refinements

Keep B6 history distinctions because they solve real migration truth, not retention:

```text
RevisionOrdinalReservation
DocumentRevision.history_kind = NATIVE | IMPORTED
RevisionImportedContent
RevisionImportedGovernanceSnapshot / equivalent target-owner imported governance state
Evidence.history_kind = NATIVE | IMPORTED
EvidenceImportedCapture / equivalent target-owner imported capture state
native vs imported timestamps remain distinct
current Tenant Dictionary is never resolved to fabricate historical state
```

Remove only the prior requirement that imported Revision/Evidence create `RetentionBinding`, because no Binding exists in Launch V1.

## 7. Artifact rule terminology refinement

Keep the structural rule, but rename its meaning from **semantic retention root** to simply **semantic root**:

> every confirmed Artifact belongs to exactly one semantic root: one DocumentRevision or one Evidence.

Multiple typed references inside the same root are allowed; cross-root reuse of the same Artifact row is rejected. Identical bytes across roots may use separate Artifact semantic rows with the same SHA-256; provider-level physical dedupe remains mechanism freedom.

This rule remains useful independently of Records Governance because it gives clear ownership and prevents cross-object lifecycle coupling.

## 8. B6 / transaction matrix refinement

Remove all same-commit paths that existed only for Records Governance:

```text
SUBMIT → RetentionBinding / Hold materialization
Evidence CAPTURE → RetentionBinding / Hold materialization
Dossier link → Hold materialization
RetentionExtension
LegalHold activation/release
DispositionFence
Disposition completion
physical delete intent for governed content
```

Keep B6:

- same-commit PII-minimized Audit for required governed mutations;
- Historical Migration / Governed Export / IMPORT_COPY / PUBLISH_COPY contracts;
- semantic-unit migration atomicity;
- exact imported-history/native-history separation;
- one local transaction through published owner seams;
- durable async intents for real external effects;
- `AuditChainHead` as final semantic lock;
- whole admitted-write lock graph proof before implementation.

Historical Migration imported Revision/Evidence no longer writes Records-Governance state.

## 9. R10-C/D/E/F scope reduction

### R10-C — Artifact Physical Integrity

Current Launch scope:

```text
ManagedArtifactStore provider-neutral conformance
Local dev/test provider
AWS S3 reference production provider
staging → integrity/malware/format validation → Artifact confirmation
canonical SHA-256 / size / exact-byte integrity
one current physical binding per Artifact
backup/restore integrity
failed/unconfirmed staging cleanup
```

Not Launch scope:

```text
governed physical disposition
disposition delete verification
Records-driven Object Lock/WORM
multi-cloud/BYOS/active-active
runtime provider-migration product feature unless independently required
```

### R10-D — Async

Remove retention timers, hold materialization jobs and disposition workers. Keep only async mechanisms required by current product effects (renderer, notifications, search projections, provider copy/export operations, reconciliation).

### R10-E — Access/API/Frontend

No Retention/LegalHold/Disposition APIs/screens/journeys in Launch V1.

### R10-F — Migration/Cutover

No launch migration of retention rules/holds/disposition history into live target authorities. Prior source records-governance facts, if present, remain explicit source/migration evidence until the capability is deliberately introduced.

## 10. Future Records-Governance seam

The deferred capability can later attach cleanly to stable identities:

```text
DocumentRevision.id
Evidence.id
Artifact semantic ownership
canonical lifecycle timestamps
Audit
```

A future bounded reopen may introduce:

```text
RetentionRule
RetentionBinding(DocumentRevision | Evidence)
RetentionExtension
LegalHold + materialized subjects
Disposition eligibility / authorization / completion
provider WORM/ObjectLock enforcement mapping
```

without changing Document/Revision/Evidence identity semantics.

## 11. Reopen triggers

Bring Records Governance back only on material evidence such as:

- law/regulation requiring minimum or maximum retention periods;
- customer/contract requiring deletion after a defined period;
- legal preservation/hold requirement;
- storage/compliance requirement for lawful governed disposal;
- operational scale/cost that makes indefinite governed preservation unacceptable;
- named eDiscovery/records-management consumer.

A hypothetical future enterprise customer is not sufficient by itself.

## 12. Current status after rebaseline

```text
R9.5 = FROZEN historical product/domain authority

R10-A/B2/B3/B5/B6
  = current R10 working target with this explicit Launch V1 bounded overlay

R10-B
  = INTEGRATED DESIGN BLOCK COMPLETE / NON-FINAL / LAUNCH-SCOPE-REBASELINED

Records Governance
  = DEFERRED FUTURE CAPABILITY

R10-C
  = NEXT / SIMPLIFIED DESIGN ONLY

implementation
  = BLOCKED until shortened C/D/E/F + Whole-R10 GCR/cold review/operator ratification + implementation plan
```
