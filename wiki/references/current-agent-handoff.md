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

Do not restore/resume historical roadmaps/specs/ADRs or current runtime concepts by inertia.

## Where we are

R3–R9 closed the principal controlled-information/governance model: Document/REV/Submission, specialized Approval, Periodic Review, Release, Distribution/Acknowledgement, Audit/Notifications/Search, tenant lifecycle/security and the final 29-Permission/five-role authorization model.

A whole-product review then surfaced material missing requirements before technical architecture: storage/repositories, format-agnostic content, external-file authoring, EigenPal/editor semantics, business documentary context, retention and migration.

Therefore:

```text
R3–R9   LOCKED principal domain/governance
R9.5    ACTIVE whole-product completion
R10-A   PAUSED / previous topology proposal NOT approved
R10+    BLOCKED until R9.5 freeze
```

## Newly locked R9.5 north star

> MetalDocs is the system of record for **identity, governance, revision, evidence and documentary context**. Physical storage, editors/viewers and upstream ERP/PLM/repositories are replaceable capabilities/providers around that kernel.

Key decisions:

- do not replace MetalDocs with M-Files/Nuxeo/Alfresco and do not build a generic ECM/M-Files clone;
- `Document` is format-agnostic, not synonymous with DOCX;
- governed sources may be DOCX/PDF/XLSX/SVG/PNG/CAD/XML/etc. subject to type/content policy;
- V1 direction = one primary source content Artifact per DocumentRevision;
- Artifact is immutable technical content identity (hash/media type/size/provenance), never a user-facing business object;
- **NO ORPHAN CONTENT:** no generic upload bucket or “upload then classify later” flow;
- Document flow: create semantic Document/REV first, then author/upload its primary Artifact;
- Evidence flow: create a typed Evidence record first, then capture/upload its Artifact;
- examples of Evidence types: Nota Fiscal, XML NF-e, Comprovante de Entrega, Foto de Inspeção, Certificado de Teste, Documento enviado pelo cliente;
- Evidence is conceptually captured business/process evidence, not automatically a `REVxxx` change-controlled Document;
- if information itself requires stable official revision/lifecycle, use Document;
- `Dossier` is a deliberately small future-facing documentary context for things such as Venda, Produto, Projeto, Equipamento, Customer/Case; it may relate Documents/Evidence and external references;
- Dossier does not replace ERP/PLM. BOM/where-used/EBOM/MBOM/CAD dependency/ECR/ECO etc. are explicit PLM-integration boundary triggers;
- EigenPal is an authoring provider for DOCX, never Document identity/domain;
- storage distinguishes Managed Artifact Store (MinIO/AWS S3) from External Repository Connector (SharePoint/OneDrive/etc.); SharePoint Embedded may be a future Microsoft enterprise profile rather than an S3-shaped provider;
- external repository edits never silently mutate an EFFECTIVE MetalDocs Revision; adoption/import produces an explicit new DRAFT REV;
- storage-version IDs never equal business `REVxxx`;
- retention/Legal Hold belongs to MetalDocs governance semantics; provider WORM/Purview are physical enforcement when used;
- future ideas reopen the kernel only when they create a material identity/historical-truth/invariant counterexample.

## Material reconsiderations

The universal R6 requirement `OFFICIAL_PDF mandatory for every Release` is **reopened**, not silently discarded. XLSX/SVG/CAD/native-PDF examples prove PDF is not universally the official semantic representation.

R9.5-1 must replace it with a format/type-aware rule: every Submission freezes exact primary source Artifact; required official/viewable Renditions depend on content/document policy.

TemplateSpec is also refined as optional structured-authoring state for applicable formats, not the universal definition of Revision content.

Tenant erasure still stands, but R9.5 Retention/Legal Hold must refine what may legally be deleted vs retained/anonymized.

## Exact next step — R9.5-1 Content Model

Design only. Close:

1. exact semantics/lifecycle of `Artifact`, `DocumentRevision`, `Evidence` and `EvidenceType`;
2. no-orphan creation-before-upload invariant;
3. allowed media/content policy ownership;
4. one-primary-source rule and any justified exception;
5. Evidence lifecycle — immutable capture vs replace-before-finalization vs narrow version semantics;
6. provenance/external references;
7. native source vs required official/viewable Renditions by format/type;
8. format-independent RevisionContent + submission digest;
9. source/view/download implications;
10. conditional TemplateSpec/structured authoring relationship.

Then: R9.5-2 Storage/Repositories → R9.5-3 Authoring/EigenPal → R9.5-4 Dossier/Context → R9.5-5 Retention/Legal Hold → R9.5-6 Import/Migration/Export → R9.5-7 Attestation/Content Security → R9.5-8 whole-product adversarial freeze.

Only after R9.5-8 resumes R10 technical architecture/filesystem/data model.

## Documentation rule

The active ledger is the single detailed WIP authority. Git history is the archive. No product implementation is authorized.