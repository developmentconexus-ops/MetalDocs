# Current Agent Handoff

> **Last verified:** 2026-08-17
> **Status:** ACTIVE — Cohesive Platform Redesign / **R9.5 FROZEN / R10 NEXT**
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Start with `AGENTS.md` and follow its authority chain. For the current stage read:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. **this file** for current status / next step
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. R9.5-8 review evidence only when auditing the freeze:
   - `docs/superpowers/analysis/2026-08-17-r9.5-8-whole-product-adversarial-freeze.md`
   - `docs/superpowers/analysis/2026-08-17-r9.5-8-independent-adversarial-challenge.md`

Git history is archive. Do not revive historical specs/ADRs/runtime concepts by inertia.

## Current checkpoint

R3–R9 principal governance is locked. R9.5 whole-product completion is now **FROZEN**:

1. **Content Model** — format-agnostic Document/WorkingContent; immutable Artifact; semantic Evidence/EvidenceType; one primary Artifact; canonical naming; format-aware official representation.
2. **Storage/Repositories** — one Managed Artifact Store/deployment; Local/MinIO/AWS S3 adapters; provider-independent hashes/keys; explicit external `IMPORT_COPY`/`PUBLISH_COPY`; SharePoint Embedded future profile.
3. **Authoring/EigenPal** — DRAFT is mutable persisted WorkingContent with one `working_version`/OCC; authorized DRAFT edit/upload/replacement is allowed; MetalDocs does not track arbitrary offline-file ancestry; replacement is whole-WorkingContent; Submission freezes one coherent exact state; SUBMITTED is immutable; Approval is read-only over exact Submission; realtime collaboration remains deferred.
4. **Dossier/Context** — stable context key/scope; small DossierType; M:N Document links; links never grant access; CAPTURED Evidence has immutable primary Dossier; ExternalReferences; explicit ERP/PLM/PM/CMMS boundaries.
5. **Retention/Records/Legal Hold** — RetentionBinding on CAPTURED Evidence / first submitted Revision; policy snapshots; explicit disposition; LegalHold over Evidence/Document/Dossier; active Document/Dossier holds prospectively materialize newly entering retention subjects within their live scope; tenant erasure remains blocked while obligations survive.
6. **Import/Migration/Export** — ordinary import distinguished from privileged Historical Migration; external history never fabricated as native MetalDocs facts; current-state/full-history modes; deterministic revision/code mapping; MigrationBatch dry-run/idempotency/reconciliation; provider-independent portability/governed-subject manifests; export completeness is explicit and authorization-safe.
7. **Launch Attestation + Basic Content Safety** — exact-Submission authenticated application approval; no false legal-signature claim; no source-byte stamping; supported-format allowlist/basic validation; security/compliance platforms remain trigger-based.
8. **Whole-Product Adversarial Freeze** — all 15 mandatory scenarios independently challenged; no material counterexample; bounded 16-permission R9.5 authorization delta accepted; final YAGNI/deletion pass accepted; independent verdict `APPROVE / FREEZE R9.5`; operator disposition ratified.

```text
R9.5-1 = LOCKED
R9.5-2 = LOCKED
R9.5-3 = LOCKED (refined by R9.5-8)
R9.5-4 = LOCKED
R9.5-5 = LOCKED (refined by R9.5-8)
R9.5-6 = LOCKED
R9.5-7 = LOCKED
R9.5-8 = CLOSED / APPROVED
R9.5   = FROZEN
R10    = NEXT / DESIGN ONLY
implementation = BLOCKED
```

## R9.5-8 disposition of independent review

Review of record:

`docs/superpowers/analysis/2026-08-17-r9.5-8-independent-adversarial-challenge.md`

Disposition:

- B1 hold-scope unlink observation — **ACCEPTED / NO CHANGE**;
- B2 never-submitted DRAFT outside hold — **ACCEPTED / DECLARED DEFERRAL**;
- B3 whole-WorkingContent replacement wording — **APPLIED** to the active ledger;
- reopen set — **EMPTY**;
- final verdict — **APPROVE / FREEZE R9.5**.

The review artifacts are evidence, not parallel target authority. The promoted semantics now live in the active ledger.

## Locked R9.5 authorization delta

The R9 base catalog is extended by exactly these 16 semantic permissions:

```text
evidence_type.manage

evidence.read
evidence.create
evidence.edit
evidence.capture
evidence.void

dossier_type.manage

dossier.read
dossier.create
dossier.manage

retention.extend
legal_hold.manage
disposition.manage

historical_migration.manage
governed_subject.export
external_repository.publish
```

No new role or provider-specific permission engine is introduced. Ordinary `IMPORT_COPY` uses normal target-object permissions. `tenant_owner` remains a grant bundle, never a bypass.

## Explicitly deferred from launch

These remain future triggers, **not hidden V1 TODOs**:

```text
malware scanning / ClamAV / quarantine / periodic rescans
ArtifactSecurityAssessment/CDR/advanced active-content analysis
ICP-Brasil/PKI/DocuSign/Adobe Sign/RFC3161/TSA/HSM
cryptographically signed export packages / package-signing-key lifecycle
custom portable export encryption
macro-enabled Office formats
full custom renderer sandbox/egress platform
eDiscovery/ESI preservation
realtime coauthoring / WOPI-style collaboration
true indivisible multi-file ArtifactPackage without a supported format requiring it
```

## Exact next step — R10 Technical Architecture

Open **R10** as an integrated technical-design stage. Do **not** start product implementation.

R10 must derive the technical realization from the frozen product/domain authority and begin with a whole-stage decomposition before microdecisions. At minimum it must disposition:

```text
bounded contexts / module owners / dependency DAG
filesystem/package ownership and legacy deletion/rename map
target data model and table ownership
DB constraints / invariant backstops
transaction boundaries
durable events / outbox / async ownership
Artifact staging/storage/relocation/restore mechanics
WorkingContent OCC + Submission atomicity
Release idempotency / exactly-one-EFFECTIVE enforcement
Retention / LegalHold / disposition mechanics
tenant erasure / backup-restore reconciliation
canonical AuthZ across query/search/export surfaces
Historical Migration transaction/idempotency contracts
external publish/job effect truth
API contracts
frontend journeys
final migration/delete map
```

R10 may choose mechanisms only after ownership/invariants are explicit. Current runtime/code/schema/OpenAPI remain evidence, not target entitlement.

## Implementation gate

**CLOSED.** R10 and subsequent stages remain design/documentation work. Product implementation begins only after the integrated technical design is complete, durable target specs/ADRs are promoted as required, material adversarial ambiguity is closed, the operator approves the integrated design, and an implementation plan is authored from that accepted target.