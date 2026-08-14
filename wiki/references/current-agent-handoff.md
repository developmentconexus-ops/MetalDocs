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

R9.5 is the whole-product completion pass. Previous R10-A topology remains **not approved** until R9.5 closes.

### Locked R9.5-1 — Content Model

- Document is format-agnostic; Artifact is immutable exact bytes + canonical SHA-256 and never a user-facing business object.
- confirmed Artifact always belongs to DocumentRevision or Evidence; staging is temporary/non-business.
- EvidenceType defines semantic type, allowed formats and small canonical naming policy; user filename is provenance only.
- Evidence lifecycle `DRAFT → CAPTURED → VOIDED` for wrong capture only; CAPTURED content immutable.
- exactly one primary Artifact per DocumentRevision and per Evidence V1.
- `RevisionContent = primary_artifact + governed_metadata + optional structured_authoring`.
- `OfficialRepresentationPolicy = SourceOnly | RequireRendition(ContentFormat)`, max one required rendition V1; universal mandatory PDF retired.

### Locked R9.5-2 — Storage / Repository

- one active Managed Artifact Store per deployment V1;
- Local(dev/test), MinIO and AWS S3 first-class; other S3-compatible products require conformance validation;
- provider migration copies/verifies bytes and cuts over without new Artifact/REV;
- object keys opaque/immutable/tenant-namespaced; no semantic content-addressed/cross-tenant dedup V1;
- direct/presigned staging upload allowed; provider success does not confirm Artifact before integrity/content/semantic validation;
- object versioning/Object Lock are defense/enforcement only, never REV/retention authority;
- production TLS + provider encryption at rest; Tenant DEK does not encrypt every Artifact V1;
- normal SharePoint/OneDrive/etc. are External Repository Connectors using `IMPORT_COPY` / `PUBLISH_COPY`, not managed-store providers;
- governed primary content requires exact MetalDocs-managed copy; external edits never mutate existing MetalDocs history;
- SharePoint Embedded reserved as future Microsoft-enterprise content profile;
- valid restore requires Artifact DB fact + exact bytes + matching SHA-256.

### Locked R9.5-3 — Authoring / EigenPal

- browser/editor never DRAFT authority; latest persisted `WorkingContent` is recoverable server truth;
- DRAFT has technical monotonic `working_version`; WorkingSnapshots are immutable technical snapshots and never REVxxx;
- governed DRAFT saves use OCC with `expected_working_version`; no last-write-wins;
- V1 one active in-app writer per DRAFT + OCC; EditorSession is narrow heartbeat/staleness lease;
- external download/edit/upload holds no long checkout and fails on stale base; no binary DOCX automerge V1;
- in-app and external editing modify same DRAFT REV; editor/provider is not business identity;
- submission freezes exact final persisted state and rejects stale autosaves after SUBMITTED;
- Approval view is read-only over exact Submission; current suggesting+autosave review behavior has no target entitlement;
- EditorialComment is product-owned DRAFT collaboration state; Approval rationale is separate evidence;
- unresolved comments and, if enabled, tracked changes block submission V1;
- realtime Yjs/coauthoring deferred but seam preserved; CRDT never becomes REV/Submission identity;
- preserve EigenPal anti-corruption adapter + exact pin + MetalDocs fidelity corpus; future Office/ONLYOFFICE providers cannot change core semantics.

### Locked R9.5-4 — Dossier / Context

- Dossier = stable documentary context for a subject such as Venda, Produto, Projeto or Equipamento; not a folder and not the ERP/PLM object itself.
- `DossierType` is tenant-scoped and deliberately small: code/name/description/status + eligible DocumentTypes/EvidenceTypes; no custom fields/forms/workflow/ACL engine.
- Dossier has stable `key`, unique within tenant+type; title may change. `{DOSSIER}` Evidence naming resolves the stable key.
- V1 has no generic Dossier numbering engine; creator/integration supplies key.
- creation provenance is separate from zero..N ExternalReferences; do not use a mutually exclusive LOCAL/EXTERNAL origin.
- ExternalReference = connection + entity kind + external ID; same external identity cannot map silently to two Dossiers. No heuristic auto-merge.
- external master fields/status remain source-system projections, not canonical Dossier state; source disappearance never deletes documentary history.
- Dossier↔Document is M:N over stable Document identity; it never copies content, changes Document lifecycle/Area/AuthZ or grants access.
- every CAPTURED Evidence has exactly one immutable `primary_dossier`; DRAFT can correct it. Evidence may relate secondarily to other Dossiers without duplication.
- Dossier scope V1 = exactly one `TenantScope | AreaScope`; Evidence reuses primary Dossier scope. No multi-area ACL.
- Dossier type/key/scope stable V1.
- lifecycle = `ACTIVE ↔ ARCHIVED`; archival is reversible MetalDocs navigation state, not Sale/Product/Project/etc. lifecycle and does not delete related content.
- relationships preserve link/unlink history; no Dossier-to-Dossier graph/hierarchy V1.
- Search/timeline are projections over canonical relationships/events.
- ERP/CRM, PLM, project-management and EAM/CMMS boundaries are explicit: MetalDocs owns documentary context, not those systems' operational domains.

## Exact next step — R9.5-5 Retention / Records / Legal Hold

Design only. Close:

1. which authorities retention applies to: released/superseded/obsolete Document revisions, CAPTURED/VOIDED Evidence, Dossier, Audit, Artifact;
2. policy ownership/configuration and snapshot/inheritance rules;
3. retention start trigger without a generic rules language;
4. expiry semantics: eligibility for disposition vs automatic deletion;
5. Legal Hold scope, apply/release authority and durable evidence;
6. interaction with Document lifecycle and Evidence invalidation;
7. interaction with tenant deletion/erasure and backups;
8. mapping to S3/MinIO Object Lock or Microsoft Purview as enforcement, never authority;
9. immutable disposition/deletion evidence;
10. whether separate Record declaration is necessary or released Documents/CAPTURED Evidence are already governed records;
11. the smallest policy model that works across Documents and Evidence without becoming a generic records-management engine;
12. what information may survive tenant erasure while a valid retention/hold obligation exists.

Then: R9.5-6 Import/Migration/Export → R9.5-7 Attestation/Content Security → R9.5-8 adversarial whole-product freeze. Only after that resume R10 technical architecture/filesystem/data model.