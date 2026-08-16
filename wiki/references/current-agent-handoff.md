# Current Agent Handoff

> **Last verified:** 2026-08-16
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
6. **Import/Migration/Export** — ordinary import distinguished from privileged Historical Migration; external history never fabricated as native MetalDocs events; current-state/full-history modes; deterministic revision/code mapping; MigrationBatch dry-run/idempotency/reconciliation; provider-independent portability/governed-subject manifests with hashes and no secrets/runtime internals.
7. **Launch Attestation + Basic Content Safety** — deliberately YAGNI: approval binds exact Submission/digest with actor/step/policy/time/assurance evidence; return-for-changes requires reason; application approval does not claim ICP-Brasil/qualified signature; approval may be manifested in a Rendition without mutating source bytes; upload only needs supported-format allowlist, basic size/type coherence and non-execution/download-safe behavior for launch.

Previous R10-A topology remains **not approved** and blocked until R9.5-8 closes.

## Explicitly deferred from launch

These are future triggers, **not hidden V1 TODOs**:

```text
malware scanning / ClamAV / quarantine / periodic rescans
ArtifactSecurityAssessment/CDR/advanced active-content analysis
ICP-Brasil/PKI/DocuSign/Adobe Sign/RFC3161/TSA/HSM
cryptographically signed export packages / package-signing-key lifecycle
custom portable export encryption
macro-enabled Office formats
full custom renderer sandbox/egress platform
eDiscovery/ESI preservation
```

The staging→validation→confirmation seam remains sufficient to add malware inspection later without redesigning Document/Evidence/Artifact/Submission.

## Important launch truths

- `ApprovalDecision` always applies to one exact immutable RevisionSubmission/digest; changed bytes/state require a new Submission.
- preserve actor, Step, ApprovalPolicy version, server time and required AuthN/fresh-auth evidence.
- MetalDocs V1 claims authenticated application approval, not a legal-signature level it has not implemented.
- no stamping/mutation of approved source bytes; human-readable approval manifestation is derived.
- DocumentType/EvidenceType only accept explicitly supported ContentFormats.
- client filename is provenance only; canonical naming is MetalDocs-owned.
- unsupported/complex formats need not be rendered inline; safe download is sufficient for V1.
- basic validation is required; a security platform is not.

## Exact next step — R9.5-8 Whole-Product Adversarial Freeze

Design only. Do **not** invent new feature scope. Test the locked model end-to-end and resolve only material contradictions.

Mandatory adversarial cases include:

1. in-app DOCX procedure from DRAFT → Submission → Approval → Release;
2. externally edited/uploaded XLSX governed Document;
3. native PDF/SVG/CAD-style controlled source without universal PDF assumption;
4. Evidence upload/classification/naming inside Sale Dossier;
5. Product Dossier with mechanical/electrical/manual Documents and inspection Evidence without drifting into PLM;
6. stale autosave / stale external edit conflict;
7. return-for-changes + same REV + new immutable Submission;
8. MinIO→S3 physical relocation with unchanged business identity;
9. SharePoint IMPORT_COPY/PUBLISH_COPY and external drift without silent mutation;
10. historical migration with incomplete bytes/dates/approval evidence;
11. retention expiry + LegalHold + explicit disposition;
12. tenant deletion blocked by retained/held content;
13. external ERP/PLM object disappearing while Dossier history remains;
14. cross-scope AuthZ, case visibility and strict SoD;
15. renderer/storage/job failure without corrupting business truth.

R9.5-8 must also produce the **bounded R9 authorization delta** for Evidence/Dossier/Retention/Import/Export operations and run one final YAGNI/deletion pass.

Only after explicit approval of R9.5-8 may R10 technical architecture/filesystem/data model resume.