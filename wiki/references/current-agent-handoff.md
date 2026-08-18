# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — Cohesive Platform Redesign / **R10-B COMPLETE NON-FINAL + LAUNCH-V1 RECORDS-GOVERNANCE DEFER REBASELINE + R10-C NEXT / SIMPLIFIED**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file — current checkpoint / exact next step
4. `wiki/architecture/cohesive-platform-redesign.md` — active program/global-coherence authority
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 historical/product-domain authority
6. `wiki/architecture/r10-technical-architecture.md` — promoted R10 authority through integrated B2
7. accepted B3/B4/B5/B6 candidates + B4/B5/B6 acceptance records
8. **`docs/superpowers/analysis/2026-08-18-launch-v1-records-governance-defer-rebaseline.md` — CURRENT LAUNCH-V1 OVERLAY; read after earlier accepted design because it deliberately supersedes only the launch-scope Records-Governance portions**
9. older review/current-state artifacts only when auditing how a decision was challenged

Git history is archive. Current code/schema/OpenAPI/module docs remain current-state/migration evidence only.

If earlier R9.5/R10 documents describe RetentionBinding, LegalHold, Disposition or Records Governance as Launch V1 target, apply the explicit Launch-V1 rebaseline above. Do not treat that as an unresolved contradiction: the rebaseline is the operator-approved bounded current-R10 overlay.

---

## Current checkpoint

```text
R9.5    = FROZEN / historical product-domain authority
R10-A   = CLOSED / APPROVED historically; Launch V1 topology rebaselined by current overlay

R10-B1  = CLOSED / APPROVED / SINGLE-COMPANY-RESTRUCTURED
R10-B2  = CLOSED / APPROVED / INTEGRATED; Launch permission catalog rebaselined 43 → 40
R10-B3  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R10-B4  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL
R10-B5  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL; Records-Governance portion DEFERRED for Launch
R10-B6  = ACCEPTED FOR R10 INTEGRATION / NON-FINAL; transaction matrix rebaselined accordingly
R10-B   = INTEGRATED DESIGN BLOCK COMPLETE / NON-FINAL / LAUNCH-SCOPE-REBASELINED

R10-C   = NEXT / SIMPLIFIED DESIGN ONLY
R10-D   = NOT STARTED
R10-E   = NOT STARTED
R10-F   = NOT STARTED

implementation = BLOCKED
```

Whole-R10 Global Coherence Review + cold independent review remain required before final R10 ratification. The rebaseline is intended to **shorten** C/D/E/F and the later implementation surface, not bypass the final review gate.

---

## Launch V1 invariant — operator approved

> **Launch V1 preserves confirmed governed history and exposes no governed physical deletion/disposition. SUPERSEDED, OBSOLETE, CANCELLED and VOIDED are lifecycle facts, never deletion commands. Only temporary/mechanism state is eligible for ordinary GC.**

Therefore these are **DEFERRED FUTURE CAPABILITY**, not Launch placeholders:

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
governed physical delete workflow
Records-driven ObjectLock/WORM
eDiscovery/custodian machinery
```

Do not create disabled tables, feature flags, dormant permissions or empty modules for them.

Reopen only on a concrete regulatory, contractual, customer or operational requirement for finite retention, legal preservation hold or governed destruction.

---

## Launch ownership / permissions after rebaseline

Seven active business bounded contexts:

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

`Records Governance` = future bounded capability.

Launch semantic permission catalog removes:

```text
retention.extend
legal_hold.manage
disposition.manage
```

Current Launch V1 count = **40**. No new role is introduced.

---

## Current R10-B kernel that remains in Launch

### Controlled Information / Approval / Release

```text
Document
→ DocumentRevision
→ WorkingContent + monotonic OCC
→ immutable RevisionSubmission
→ one sequential Governance Step route
→ SubmissionFeedback
→ optional Rendition
→ automatic Release
→ Distribution obligations / explicit acknowledgement
```

Keep the approved bounded Approval refinement: there is no structural `review|approval` Step discriminator; collaboration and fresh-auth are orthogonal.

### Documentary Context / Evidence

- Dossier is stable documentary context only; never folder/content owner/access grant.
- Dossier↔Document is contextual M:N over stable Document identity.
- Evidence remains a small `DRAFT → CAPTURED → VOIDED` lifecycle with exact immutable captured Artifact/payload.
- no generic Record declaration.

### Artifact

Every confirmed Artifact has exactly one **semantic root**: one DocumentRevision or one Evidence. Multiple references inside the same root are allowed; cross-root Artifact-row reuse is rejected. Same bytes may exist as distinct Artifact rows; physical dedupe is provider mechanism freedom.

### Audit / Interchange

Keep B6:

- same-local-commit required PII-minimized tamper-evident Audit;
- Audit is never state/outbox/event sourcing;
- Historical Migration / Governed Export / IMPORT_COPY / PUBLISH_COPY remain distinct;
- migration never fabricates native Approval/Release/User history;
- semantic-unit migration atomicity;
- complete export uses bounded consistent snapshot + provider-independent exact manifest;
- one local transaction through owner seams; provider effects remain async mechanisms;
- AuditChainHead final semantic lock; admitted-write lock graph must be mechanically proven before implementation.

---

## B3/B5/B6 refinements after Launch rebaseline

### Keep

- `DocumentRevision.cancelled_at` / `obsoleted_at` as native lifecycle facts;
- `RevisionOrdinalReservation`;
- `DocumentRevision.history_kind = NATIVE | IMPORTED`;
- `RevisionImportedContent`;
- target-owner imported Revision governance state;
- Evidence native/imported history distinction and imported capture state;
- native vs imported timestamps remain distinct;
- current Tenant Dictionary is never resolved to fabricate historical state.

### Simplify

**Dictionary:** no separate `RevisionDictionarySnapshot` table in Launch; immutable dictionary snapshot returns to `DocumentRevision.dictionary_snapshot` because governed disposition is absent.

**Template origin:** use a closed typed exact source reference:

```text
source_kind = NATIVE_SUBMISSION | IMPORTED_REVISION_CONTENT
native      → exact RevisionSubmission FK
imported    → exact imported Revision content identity
```

No retention-survival provenance snapshot machinery is needed until governed disposal returns.

**Historical Migration:** no RetentionBinding/Hold materialization for imported Revision/Evidence because Records Governance is absent.

---

## Exact next step — simplified R10-C Artifact Physical Integrity

Open **R10-C** only for physical integrity required by the launch product:

```text
ManagedArtifactStore provider-neutral conformance
Local first-class dev/test provider
AWS S3 reference production provider
staging → upload completion → integrity/malware/format validation → semantic confirmation
canonical full-byte SHA-256 + size verification
physical location binding without leaking provider identity into Artifact
no-overwrite / immutable confirmed bytes
backup/restore exact-byte integrity
failed/unconfirmed staging and temporary export/render cleanup
privacy non-resurrection reconciliation where restore is implicated
```

Explicitly **not** R10-C Launch scope:

```text
governed physical disposition
DispositionFence / DispositionRecord delete verification
LegalHold / retention clocks
Records-driven ObjectLock/WORM
multi-cloud/BYOS/active-active
content-addressed business identity
permanent dual-write
```

Keep later routing:

```text
async retry/lease/jobs/projections/effects → R10-D
API/frontend/viewer/download journeys      → R10-E
historical cutover/bootstrap/legacy delete → R10-F
```

Current implementation is evidence only. Implementation remains **BLOCKED** until the shortened remaining R10 stages, Whole-R10 review, operator ratification and implementation plan are complete.
