# Current Agent Handoff

> **Last verified:** 2026-08-14
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

R3–R9 principal governance is locked: format-independent Document/REV/Submission lifecycle, specialized Approval + SoD, Periodic Review, automatic Release, Distribution/Acknowledgement, Audit/Notifications/Search, tenant lifecycle/security and the R9 five-role/29-Permission model.

R9.5 is the whole-product completion pass added before technical architecture. Previous R10-A topology remains **not approved**.

### R9.5 north star

MetalDocs is system of record for **identity, governance, revision, evidence and documentary context**. Physical storage, editors/viewers and upstream ERP/PLM/repositories are providers/connectors around the kernel.

### Locked R9.5-1 Content Model

- `Document` is format-agnostic; DOCX is only one source format.
- `Artifact` = immutable exact bytes + canonical SHA-256; never user-facing business object and never provider/location identity.
- confirmed Artifact always belongs to a DocumentRevision or Evidence;
- staging upload may precede Evidence classification inside a Dossier, but staging is temporary/non-business;
- tenant-scoped `EvidenceType` defines semantic evidence type, allowed formats and small naming policy;
- MetalDocs generates canonical filenames; original upload filename is provenance only;
- naming tokens V1: `{TYPE}`, `{DOSSIER}`, `{REF}`, `{SEQ}`;
- Evidence lifecycle: `DRAFT → CAPTURED → VOIDED` only for invalid MetalDocs capture; CAPTURED content immutable;
- Evidence does not use REV/Approval/Release by default; revision-governed information should be Document;
- exactly one primary Artifact per DocumentRevision and per Evidence V1;
- product-owned `ContentFormat` catalog; DocumentType/EvidenceType choose allowed formats;
- format-independent `RevisionContent = primary_artifact + governed_metadata + optional structured_authoring`;
- Submission digest binds canonical business/content state and exact Artifact hash, never storage location;
- `OfficialRepresentationPolicy = SourceOnly | RequireRendition(ContentFormat)`, max one required rendition V1; universal mandatory PDF is retired.

### Locked R9.5-2 Storage / Repository

- one active Managed Artifact Store per deployment V1;
- first-class adapters: Local(dev/test), MinIO, AWS S3; other S3-compatible providers require conformance validation;
- provider migration = copy exact bytes + verify canonical hash + cutover; no new Artifact/REV and no permanent dual-write;
- physical object keys are opaque, immutable and tenant-namespaced; business filename never determines key/path;
- Artifact ID != content hash; no content-addressed/cross-tenant dedup V1;
- direct/presigned staging upload allowed; provider success does not confirm Artifact until integrity/content/semantic validation succeeds;
- canonical content hash = SHA-256; provider ETag/version/checksum is supporting evidence only;
- object versioning and Object Lock/WORM are defense/enforcement capabilities, never business revision/retention authority;
- production baseline = TLS + provider encryption at rest; Tenant DEK does not encrypt every Artifact V1;
- normal SharePoint/OneDrive/etc. are External Repository Connectors, not S3 providers;
- governed primary content V1 requires exact MetalDocs-managed copy; connector directions start with `IMPORT_COPY` / `PUBLISH_COPY`;
- external edits never mutate existing MetalDocs history; future adoption creates a new Artifact/new DRAFT REV where applicable;
- SharePoint Embedded is reserved as a future Microsoft-enterprise content-backend profile, not forced into S3 semantics;
- valid restore = Artifact DB fact + exact bytes + matching SHA-256; staging/incomplete uploads are garbage-collectable.

## Exact next step — R9.5-3 Authoring / EigenPal

Design only. Close:

1. authoritative working state of a DRAFT Revision;
2. working snapshots vs confirmed Artifacts;
3. in-app editing vs download/edit/upload replacement;
4. autosave and optimistic concurrency;
5. one-writer vs concurrent vs real-time collaboration V1;
6. EditorSession/presence semantics;
7. tracked changes and whether they are governed source content or review overlay;
8. comments/annotations vs Approval rationale;
9. return-for-changes/reviewer suggestions;
10. crash/offline recovery;
11. EigenPal anti-corruption/provider seam + upstream version strategy;
12. future editor providers without changing Document/REV/Submission truth.

Then R9.5-4 Dossier/Context → Retention → Import/Migration/Export → Attestation/Content Security → adversarial whole-product freeze. Only then resume R10 technical architecture.