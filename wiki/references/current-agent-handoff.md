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

- Document is format-agnostic; Artifact = immutable exact bytes + canonical SHA-256 and never user-facing business identity.
- confirmed Artifact always belongs to DocumentRevision or Evidence; staging is temporary/non-business.
- EvidenceType defines semantic type, allowed formats and small canonical naming policy; user filename is provenance only.
- Evidence lifecycle `DRAFT → CAPTURED → VOIDED` for wrong capture only; CAPTURED content immutable.
- one primary Artifact per DocumentRevision/Evidence V1.
- `RevisionContent = primary_artifact + governed_metadata + optional structured_authoring`.
- `OfficialRepresentationPolicy = SourceOnly | RequireRendition(ContentFormat)`, max one required rendition V1.

### Locked R9.5-2 — Storage / Repository

- one active Managed Artifact Store per deployment V1;
- Local(dev/test), MinIO and AWS S3 first-class; other S3-compatible providers require conformance validation;
- provider migration copies/verifies bytes and cuts over without new Artifact/REV;
- object keys opaque/immutable/tenant-namespaced; no content-addressed/cross-tenant dedup V1;
- direct/presigned staging upload allowed; provider success does not confirm Artifact before integrity/content/semantic validation;
- object versioning/Object Lock are defense/enforcement only, never REV/retention authority;
- production TLS + provider encryption at rest; Tenant DEK does not encrypt every Artifact V1;
- normal SharePoint/OneDrive/etc. are External Repository Connectors using `IMPORT_COPY` / `PUBLISH_COPY`;
- governed primary content requires exact MetalDocs-managed copy; external edits never mutate existing MetalDocs history;
- SharePoint Embedded reserved as future Microsoft-enterprise content profile;
- valid restore requires Artifact DB fact + exact bytes + matching SHA-256.

### Locked R9.5-3 — Authoring / EigenPal

- latest persisted `WorkingContent` is recoverable DRAFT truth; browser/editor never authority;
- technical monotonic `working_version` + immutable WorkingSnapshots; never REVxxx;
- governed saves use OCC with `expected_working_version`; no last-write-wins;
- one active in-app writer per DRAFT + OCC V1; external edit/download/upload uses stale-base conflict, no long checkout/automerge;
- editor/provider is not business identity;
- submission freezes exact final persisted state and rejects stale autosaves after SUBMITTED;
- Approval view is read-only over exact Submission;
- EditorialComment is product-owned DRAFT state, separate from Approval rationale;
- unresolved comments/tracked changes block submit when those capabilities are enabled;
- realtime Yjs deferred; preserve provider seam;
- EigenPal stays behind one anti-corruption adapter + pinned version + MetalDocs fidelity corpus.

### Locked R9.5-4 — Dossier / Context

- Dossier = stable documentary context for Venda/Produto/Projeto/Equipamento/etc., not folder or ERP/PLM object.
- DossierType = small tenant config with code/name/status + eligible DocumentTypes/EvidenceTypes; no custom-object engine.
- stable key unique tenant+type; title mutable; `{DOSSIER}` uses key; no generic numbering engine V1.
- creation provenance separate from zero..N unique ExternalReferences; no heuristic auto-merge.
- external master/status data remains source-system projection; source disappearance never deletes documentary history.
- Dossier↔Document is M:N over stable Document identity and never grants/changes Document access/lifecycle.
- every CAPTURED Evidence has one immutable primary Dossier; secondary Dossier relationships allowed without duplication.
- scope = one TenantScope|AreaScope; Evidence reuses primary Dossier scope; no multi-area ACL.
- lifecycle only `ACTIVE ↔ ARCHIVED`; archive is navigation state and never deletes content.
- link/unlink history is preserved; no Dossier graph/hierarchy V1.
- ERP/CRM/PLM/PM/CMMS operational domains remain external.

### Locked R9.5-5 — Retention / Records / Legal Hold

- no duplicate generic Record entity; CAPTURED Evidence and first-submitted DocumentRevision become retention subjects through `RetentionBinding`.
- working snapshots/staging/unsubmitted drafts stay under GC/recovery, not records retention.
- DocumentType/EvidenceType explicitly choose `NoMinimum | KeepFor(value, DAYS|MONTHS|YEARS) | Indefinite`; no hardcoded legal periods or rules engine.
- Document retention clock does not run while EFFECTIVE; anchors are SUPERSEDED/OBSOLETE/CANCELLED(after submission). Evidence anchor = CAPTURED_AT or OCCURRED_AT.
- policy is snapshotted; future config changes do not silently recalc old records. `RetentionExtension` may lengthen, not generically shorten, existing obligations.
- expiry means only eligible-for-disposition. Current EFFECTIVE content is never eligible and there is no automatic delete cron V1.
- disposition requires explicit authorization/review + zero active holds + successful verified removal of substantive payload/Artifacts; completion creates immutable DispositionRecord.
- LegalHold is independent of retention duration and may scope Evidence, stable Document or Dossier. Document/Dossier holds materialize concrete held subjects; unlink/lifecycle change never releases held history.
- hold blocks destruction, not normal business lifecycle. Full eDiscovery/transient-autosave preservation is outside V1.
- Artifact has no independent business retention policy; provider WORM/Object Lock/Purview is enforcement only.
- DossierType has no retention inheritance; Audit remains separate regime.
- tenant terminal erasure is blocked while retained/held subjects remain; no new tenant state is added and required DEK material is not destroyed prematurely.
- backup/restore must restore/reconcile retention/hold/disposition facts before cleanup; erasure tombstones retain precedence.

## Exact next step — R9.5-6 Import / Migration / Export

Design only. Close:

1. how legacy/external Documents enter without fabricating native MetalDocs history;
2. legacy code/revision-label preservation vs target REVxxx normalization;
3. current-state-only versus full-history import;
4. externally approved/released history and how to represent it without fake ApprovalDecision/ReleaseRecord;
5. import provenance/source identifiers/exact Artifact hashes;
6. Evidence/Dossier import and historical occurred_at/retention anchors;
7. duplicate/conflict detection and replay-safe/idempotent migration;
8. dry-run, validation, reconciliation and abort/rollback boundary;
9. privileged migration/import path versus normal user operations;
10. export scope and package semantics;
11. manifests/hashes/relationships/provenance needed for verifiable export;
12. separation of backup vs interoperability export vs evidentiary package;
13. interaction with external repository IMPORT_COPY/PUBLISH_COPY;
14. retention/legal-hold constraints on export and disposition evidence.

Then: R9.5-7 Attestation/Content Security → R9.5-8 adversarial whole-product freeze. Only after that resume R10 technical architecture/filesystem/data model.