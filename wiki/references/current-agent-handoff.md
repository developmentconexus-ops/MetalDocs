# Current Agent Handoff

> **Last verified:** 2026-08-15
> **Status:** ACTIVE — Cohesive Platform Redesign
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Implementation:** **BLOCKED — design/documentation only**

## Read order

1. `AGENTS.md`
2. `wiki/standards/root-cause-global-maximum-method.md`
3. `wiki/architecture/cohesive-platform-redesign.md`
4. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
5. this file

Git history is archive. Do not revive historical specs/ADRs/runtime concepts by inertia.

## Current checkpoint

R3–R9 principal governance is locked. R9.5 whole-product completion has now locked:

1. **Content Model** — format-agnostic Document; immutable Artifact; semantic Evidence/EvidenceType; one primary Artifact; canonical naming; format-aware official representation.
2. **Storage/Repositories** — one Managed Artifact Store/deployment; Local/MinIO/AWS S3 adapters; provider-independent hashes/keys; external repositories use explicit import/publish copies; SharePoint Embedded future profile.
3. **Authoring/EigenPal** — persisted WorkingContent + technical working_version/OCC; one writer V1; external editing without long checkout; immutable Submission boundary; Approval read-only; realtime collaboration deferred behind seam.
4. **Dossier/Context** — stable context key/scope; small DossierType; M:N Document links; CAPTURED Evidence has immutable primary Dossier; ExternalReferences; explicit ERP/PLM/PM/CMMS boundaries.
5. **Retention/Records/Legal Hold** — RetentionBinding on CAPTURED Evidence / first submitted Revision; type-scoped policy snapshots; explicit disposition; LegalHold over Evidence/Document/Dossier; WORM/Purview enforcement only; tenant erasure blocked while obligations survive.
6. **Import/Migration/Export** — ordinary import distinguished from privileged Historical Migration; external history never fabricated as native MetalDocs events; current-state/full-history modes; deterministic revision/code mapping; MigrationBatch dry-run/idempotency/reconciliation; tenant portability and governed-subject packages are provider-independent manifests with exact hashes and no secrets/runtime internals.

Previous R10-A topology remains **not approved** and blocked until R9.5 closes.

## Key migration/export truths

- Every migrated object carries explicit source provenance.
- Imported approval/effectivity facts are imported-governance evidence, never synthetic native ApprovalDecision/ReleaseRecord.
- Unknown historical dates stay unknown; `adopted_as_current_at` is a separate true fact.
- Target DocumentRevision requires exact source bytes; missing legacy bytes never produce fake Revision rows.
- Numeric source ordinals may preserve continuity (`7 → REV007`); arbitrary source labels map deterministically while preserving original label.
- No silent format normalization/transformation.
- Historical source actors remain source snapshots; migration execution actor is Migration/System principal.
- Migration never replays past notifications/jobs/distribution side effects.
- Retention on imported content uses trustworthy historical anchors when known; migration never auto-disposes expired records.
- Backup ≠ Tenant Portability Export ≠ Governed Subject Export ≠ PUBLISH_COPY.
- Portability exports business/governance authorities + exact Artifacts/current DRAFT state, excludes secrets/staging/jobs/outbox/rebuildable projections.
- Export packages use versioned provider-independent manifest + SHA-256 inventory. Package signature/encryption is R9.5-7.

## Exact next step — R9.5-7 Attestation + Content Security

Design only. Close:

1. exact semantic statement/evidence created by Approval `accept` and `return_for_changes`;
2. application attestation vs formal electronic/digital signature boundary;
3. immutable ApprovalReceipt/Decision evidence binding actor identity + assurance/fresh-auth + Submission digest;
4. how approval/effectivity manifests in human-readable official renditions without altering approved source truth;
5. portability/export package signature/integrity/encryption boundary;
6. upload quarantine/staging versus Document/Evidence lifecycle;
7. real content-type detection, allowed-format validation, size/complexity limits and parser hardening;
8. malware scanning and fail-closed behavior when scanner unavailable;
9. Office macros/external relationships, PDF active content, archive bombs and other risky content features;
10. safe preview/view/download headers and policy;
11. rendering/conversion sandbox and outbound-network policy for untrusted content;
12. security evidence/audit facts without making scanner telemetry business authority.

After R9.5-7: **R9.5-8 whole-product adversarial freeze + bounded authorization delta**. Only after explicit approval does R10 technical architecture/filesystem/data model resume.