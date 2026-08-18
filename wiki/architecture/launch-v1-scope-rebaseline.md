# Launch V1 Scope Rebaseline — Active R10 Overlay

> **Status:** ACTIVE / OPERATOR-APPROVED CURRENT-R10 OVERLAY  
> **Date:** 2026-08-18  
> **Applies to:** Launch V1 target only  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Detailed rationale:** `docs/superpowers/analysis/2026-08-18-launch-v1-records-governance-defer-rebaseline.md`

This page is the durable authority overlay for the Launch V1 scope. Earlier R9.5/R10 documents remain historical/accepted design evidence, but the rules below supersede their Launch-V1 Records-Governance assumptions wherever they conflict.

## 1. Launch invariant

> **Launch V1 preserves confirmed governed history and exposes no governed physical deletion/disposition. SUPERSEDED, OBSOLETE, CANCELLED and VOIDED never imply deletion. Only temporary/mechanism state is eligible for ordinary GC.**

## 2. Records Governance deferred

The following are **not Launch V1 implementation scope**:

```text
Records Governance bounded context/module
DocumentTypeRetentionRule
EvidenceTypeRetentionRule
RetentionBinding
RetentionExtension
LegalHold
LegalHoldSubject
DispositionFence
DispositionRecord
retention clocks / expiry eligibility
governed physical deletion workflow
Records-driven ObjectLock/WORM
eDiscovery/custodian machinery
```

Do not create dormant tables, permissions, modules, flags or jobs for them.

Method outcome: `DEFER SAFELY`.

Reopen only on concrete regulatory, contractual, customer or operational evidence requiring finite retention, legal preservation hold or governed destruction.

## 3. Launch topology

Active business bounded contexts:

```text
Authentication
Organization
Authorization
Controlled Information
Approval
Documentary Context
Distribution
```

Supporting semantic owners:

```text
Artifact
Audit
Interchange
```

`Records Governance` is a future capability, not an empty launch owner.

## 4. Launch permission catalog

Remove:

```text
retention.extend
legal_hold.manage
disposition.manage
```

Launch semantic permission count = **40**. No new role.

## 5. B5 launch interpretation

B5 Launch scope becomes:

```text
Documentary Context
+ Evidence
+ Artifact semantic ownership closure
```

Keep Dossier/Evidence semantics and one semantic Artifact root (`DocumentRevision` or `Evidence`). Remove all RetentionBinding/Hold/Disposition persistence and transactions from the launch target.

## 6. B3/B6 bounded consequences

Keep:

- native `DocumentRevision.cancelled_at` / `obsoleted_at` lifecycle facts;
- `RevisionOrdinalReservation`;
- `history_kind=NATIVE|IMPORTED`;
- `RevisionImportedContent` and imported target-owner governance state;
- imported Evidence capture/provenance state;
- native/imported timestamp truth separation;
- no current-dictionary fabrication of historical state.

Simplify for Launch:

- dictionary snapshot returns to immutable `DocumentRevision.dictionary_snapshot`; no separate `RevisionDictionarySnapshot` launch table;
- `DocumentOrigin` uses a closed typed exact source reference: `NATIVE_SUBMISSION | IMPORTED_REVISION_CONTENT`; no retention-survival snapshot machinery;
- Historical Migration creates no RetentionBinding/Hold state.

Artifact terminology: keep **one semantic root per Artifact**, not “retention root”.

## 7. B6 transaction-matrix consequence

Remove Records-only paths from the Launch transaction/lock graph:

```text
RetentionBinding creation
Hold materialization
RetentionExtension
LegalHold activation/release
DispositionFence
Disposition completion
governed-delete intent
```

Keep same-commit Audit, Interchange truth, migration semantic-unit atomicity, export consistency, owner transaction seams, durable real-effect intents, and AuditChainHead as final semantic lock.

## 8. Remaining R10 scope

```text
R10-B = INTEGRATED DESIGN BLOCK COMPLETE / NON-FINAL / LAUNCH-SCOPE-REBASELINED
R10-C = NEXT / SIMPLIFIED DESIGN ONLY
R10-D = shortened: no retention/hold/disposition async work
R10-E = shortened: no retention/hold/disposition API/UI
R10-F = shortened: no launch Records-Governance live-state migration
```

R10-C Launch focus:

```text
ManagedArtifactStore conformance
Local dev/test + AWS S3 reference production
staging / integrity / malware / format validation
canonical SHA-256 and exact-byte confirmation
physical binding / no-overwrite
backup/restore exact-byte integrity
cleanup of unconfirmed/temporary mechanism state
```

No governed physical disposition or Records-driven ObjectLock/WORM in Launch V1.

## 9. Implementation gate

Implementation remains **BLOCKED** until shortened R10-C/D/E/F complete, Whole-R10 Global Coherence Review + cold independent review are resolved, the operator ratifies the final target, and an implementation plan is authored.
