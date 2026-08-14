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

R3–R9 principal governance is locked: format-independent Document/REV/Submission, specialized Approval + SoD, Periodic Review, automatic Release, Distribution/Acknowledgement, Audit/Notifications/Search, tenant lifecycle/security and the five-role/29-Permission model.

R9.5 is the whole-product completion pass. Previous R10-A topology remains **not approved** until this pass closes.

### R9.5 north star

MetalDocs is system of record for **identity, governance, revision, evidence and documentary context**. Physical storage, editors/viewers and upstream ERP/PLM/repositories are providers/connectors around that kernel.

### Locked R9.5-1 — Content Model

- Document is format-agnostic; Artifact is immutable exact bytes + canonical SHA-256 and is never a user-facing business object.
- confirmed Artifact always belongs to DocumentRevision or Evidence; staging is temporary/non-business.
- EvidenceType defines semantic evidence type, allowed formats and small canonical naming policy; user filename is provenance only.
- Evidence lifecycle `DRAFT → CAPTURED → VOIDED` for wrong capture only; CAPTURED content immutable.
- exactly one primary Artifact per DocumentRevision and per Evidence V1.
- `RevisionContent = primary_artifact + governed_metadata + optional structured_authoring`.
- `OfficialRepresentationPolicy = SourceOnly | RequireRendition(ContentFormat)`, max one required rendition V1; universal mandatory PDF retired.

### Locked R9.5-2 — Storage / Repository

- one active Managed Artifact Store per deployment V1;
- first-class adapters Local(dev/test), MinIO and AWS S3; other S3-compatible providers require conformance validation;
- provider migration copies/verifies bytes and cuts over without creating new Artifact/REV;
- physical keys opaque/immutable/tenant-namespaced; no content-addressed/cross-tenant dedup V1;
- direct/presigned staging upload allowed; provider success does not confirm Artifact before integrity/content/semantic validation;
- object versioning/Object Lock are defense/enforcement capabilities, never REV/retention authority;
- production baseline TLS + provider encryption at rest; Tenant DEK does not encrypt every Artifact V1;
- normal SharePoint/OneDrive/etc. are External Repository Connectors using `IMPORT_COPY` / `PUBLISH_COPY`, not managed-store providers;
- governed primary content requires an exact managed copy; external edits never mutate existing MetalDocs history;
- SharePoint Embedded reserved as future Microsoft-enterprise content profile;
- valid restore requires Artifact DB fact + exact bytes + matching SHA-256.

### Locked R9.5-3 — Authoring / EigenPal

- browser/editor is never DRAFT authority; latest persisted `WorkingContent` is recoverable server truth;
- DRAFT has monotonic technical `working_version`; WorkingSnapshots are immutable technical snapshots and never REVxxx;
- all governed DRAFT changes share the same OCC version; server save/replacement requires `expected_working_version`; no last-write-wins;
- V1 uses one active in-app writer per DRAFT plus OCC; `EditorSession` is a narrow heartbeat/staleness authoring lease;
- external download/edit/upload holds no long checkout and must fail on stale base; no automatic DOCX binary merge V1;
- in-app and external replacement modify the same DRAFT REV; editor/provider is not persisted business identity;
- submission performs/follows final successful flush, freezes exact persisted logical state, rejects all later stale autosaves and makes REV SUBMITTED;
- Approval view is read-only over exact Submission; current `review → suggesting + autosave` behavior has no target entitlement;
- EditorialComment is MetalDocs DRAFT collaboration state, separate from Approval rationale and vendor DOCX comment authority;
- unresolved editorial comments and, if enabled, tracked changes block submission V1;
- realtime Yjs/coauthoring deferred V1 but preserved as an authoring-infrastructure seam; CRDT state never becomes REV/Submission identity;
- preserve one EigenPal anti-corruption/provider adapter, pin exact version and validate upgrades against a MetalDocs fidelity corpus;
- future Office/ONLYOFFICE authoring must not change Document/REV/Submission semantics.

## Exact next step — R9.5-4 Dossier / Context

Design only. Close the minimum common context model using adversarial examples `Venda`, `Produto`, `Projeto`, `Equipamento`:

1. Dossier semantic role versus Document/Evidence and source ERP/PLM entity;
2. DossierType configurability without a generic custom-object engine;
3. local-created vs externally-originated identity;
4. stable display/business key and external references;
5. Document/Evidence relation cardinality, including multi-Dossier cases;
6. whether Dossier-to-Dossier relations are needed V1;
7. lifecycle/close/archive semantics and treatment of external source status;
8. what minimal metadata belongs in MetalDocs vs remains external projection;
9. allowed/recommended EvidenceTypes/DocumentTypes by DossierType;
10. search/navigation/activity timeline;
11. ERP/PLM synchronization and source-of-truth boundaries;
12. explicit trigger where a requirement has become ERP/CRM/PLM rather than documentary context.

Then: R9.5-5 Retention/Legal Hold → R9.5-6 Import/Migration/Export → R9.5-7 Attestation/Content Security → R9.5-8 adversarial whole-product freeze. Only after that resume R10 technical architecture/filesystem/data model.