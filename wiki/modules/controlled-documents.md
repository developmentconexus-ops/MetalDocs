# Module: Controlled Documents

> Living architecture doc. Arc42 (12 sections) + C4 (Context/Container) Mermaid diagrams. Supersedes the 2026-05-07 stub.

**Last verified:** 2026-08-12 (A5 transaction lifecycle: `PostgresControlledDocumentRepository` now owns a `platform/db.TxRunner`; standalone `Create` uses `runner.Do`, while `CreateTx` remains the explicit caller-owned `db.Tx` seam) | **Prior:** 2026-08-09 (Phase G — cross-link added: ADR 0093 (Accepted design ruling, not implemented) rules that this module, `documents`, and `templates` will merge into a single **Controlled Information** bounded context; no code/schema change authorized yet; execution owned by #94/A9. Content below otherwise re-verified 2026-06-15) | **Prior:** 2026-06-15 (M4/F4.2 — port ADR 0030 cross-link: `PostgresTemplateVersionChecker` deleted; CD now consumes templates-owned `TemplateVersionPort.GetTemplateVersionState` — `status := "published"` hardcode removed) | **Prior:** 2026-06-12 (Wave 2.12 sync — db==nil authz-bypass class-B branch in Create DELETED (authz now unconditional); DBTX local interface replaced with db.Tx; sequence.go no longer imports database/sql; nosqltxindomain CI guard. Prior Wave 2 sync: GetActiveInstance service+repo extraction, ActiveDocumentInstance domain type, ErrNoActiveInstance, govLogger.LogTx in changeStatus, AuditWriter nil-panic guard in module.go, RLS migration 0234) | **Owner:** leandro | **Status:** active | **Maturity:** L2

> **Key files:**
> - `internal/modules/controlleddocuments/module.go:27` - module wiring (`New`, dependencies)
> - `internal/modules/controlleddocuments/application/service.go:146` - service `Create` (legacy literal struct identifier noted in Historical Literal Key Notes)
> - `internal/modules/controlleddocuments/application/service.go:451` - `Obsolete` / `Supersede` via `changeStatus`
> - `internal/modules/controlleddocuments/delivery/http/handler.go:49` - `injectTenant` middleware (reads tenant via `tenant.FromContext`)
> - `internal/modules/controlleddocuments/delivery/http/handler.go:61` - `tenantIDFromContext` (local context accessor)
> - `internal/modules/controlleddocuments/delivery/http/routes.go:70` - `AtomicCreateControlledDocument` handler
> - `internal/modules/controlleddocuments/delivery/http/routes.go:266` - `GetActiveDocument` handler (delegates to service — delivery layer is now SQL-free for this path; Wave 2)
> - `internal/modules/controlleddocuments/application/service.go:497` - `GetActiveInstance` (new service method — Wave 2; authz read-check then delegates to repo)
> - `internal/modules/controlleddocuments/infrastructure/repository.go:36,48` - Postgres repository and constructor; the constructor wires `platform/db.TxRunner`
> - `internal/modules/controlleddocuments/infrastructure/repository.go:385,405` - standalone `Create` via `runner.Do` and caller-owned `CreateTx` seam
> - `internal/modules/controlleddocuments/infrastructure/repository.go:582` - `GetActiveInstance` (new repo method — Wave 2; extracted FULL OUTER JOIN query from delivery layer)
> - `internal/modules/controlleddocuments/domain/port.go:8` - `ActiveDocumentInstance` domain type (Wave 2); `GetActiveInstance` on `ControlledDocumentRepository` interface at `:37`
> - `internal/modules/controlleddocuments/domain/controlled_document.go:37` - `ErrNoActiveInstance` sentinel (Wave 2)
> - `internal/modules/controlleddocuments/delivery/http/routes.go:595` - `tenantIDFromRequest` -> `tenant.FromContext`
> - `internal/modules/controlleddocuments/domain/document_initializer.go:61` - `DocumentInitializer` port (consumed by documents)
> - `internal/modules/controlleddocuments/infrastructure/repository.go:484` - `UpdateStatus` (lifecycle mutation)
> - `db/baseline/0001_current_schema.sql` - consolidated schema baseline (includes initial controlled-documents and per-area sequence tables; replaces legacy `migrations/0124…0183` series)
> - `db/migrations/0210_controlled_documents_capability_namespace.sql` - capability namespace rename (post-baseline)

---

## 1. Introduction & Goals

The **controlled-documents** module owns the catalog of code-numbered Controlled Documents (CDs). Each row in `public.controlled_documents` is a numbered slot binding a (`profile_code`, `process_area_code`) pair to a chain of `documents` revisions. The CD itself carries no content - it owns the identity, the per-(tenant, profile, area) sequence number, and the lifecycle status (`active | obsolete | superseded`).

### 1.1 Requirements overview

- Atomic CD + first-revision create - source: [wiki/decisions/0011-cd-atomic-create.md](decisions/0011-cd-atomic-create.md)
- Per-(profile, area) monotonic 3-digit sequence (`DC-RH-001`) - source: [wiki/concepts/controlled-documents.md](concepts/controlled-documents.md)
- `Idempotency-Key` replay safety on creation paths - source: ADR 0011
- Preview endpoint returns next code without sequence reservation - source: ADR 0011
- Active-document lookup tolerates published-only state (FULL OUTER JOIN) - source: E10 fix (commit 1dfcf3da)

### 1.2 Quality Goals

| Rank | Goal | How verified |
|---|---|---|
| 1 | Numbering integrity          no duplicate / out-of-band CD codes | `cd_sequence_counters` PK on (tenant_id, profile_code, process_area_code); UNIQUE (tenant_id, profile_code, code) on `controlled_documents`; `domain/sequence_test.go` |
| 2 | No orphan slots          CD row + first revision either both commit or both roll back | `application/integration_test.go` covering Create+CloneTemplate in single `*sql.Tx` (`service.go:243-257`) |
| 3 | Replay safety on creation paths | `internal/platform/idempotency` middleware + body-hash check; `routes_contract_test.go` |

### 1.3 Stakeholders

| Role | Expectation |
|---|---|
| Document author | Can issue a new CD via `POST /api/v1/controlled-documents` and immediately edit the first revision |
| Quality manager | `controlled_documents.code` is stable, audit-traceable, area-isolated |
| Operator | Per-(tenant, profile, area) sequence is monotonic; lifecycle transitions are recoverable |
| Frontend developer | `GET .../active-document` returns 200 with at least one document handle whenever the CD has any revision |

---

## 2. Architecture Constraints

- Language / runtime: Go 1.25
- Persistence: Postgres (per [wiki/architecture/data-model.md](architecture/data-model.md))
- API contract: OpenAPI 3.0.3 via oapi-codegen v2 - canonical spec partial `api/openapi/v1/partials/controlled-documents.yaml`; canonical public API prefix remains `/api/v1/controlled-documents*`.
- Authz: two-tier per [wiki/decisions/0007-two-tier-authz.md](decisions/0007-two-tier-authz.md); Plan 5 wired `authz.Require` + tripwire on `controlled_documents` and `cd_sequence_counters` (T-001/T-004 closed); `Obsolete`/`Supersede` use dedicated lifecycle capability constants from `iamdomain` (legacy literal identifiers; see Historical Literal Key Notes)
- Idempotency: shared platform `internal/platform/idempotency` per ADR 0011
- Numbering: 3-segment `{PROFILE}-{AREA}-{NNN}` per ADR 0011
- Error envelope: **RFC 9457** `application/problem+json` via `httpresponse.WriteError` -> `problem.Write` (T-003 closed Plan 7); `ErrTemplateProfileMismatch` -> 422 `template_invalid` via direct `problem.Write` (T-007 closed Plan 7)

---

## 3. System Scope & Context — module-scoped (C4 Level 1)

> System-level context lives in [`wiki/diagrams/c4-context.md`](../diagrams/c4-context.md). The diagram below is **module-scoped**: it shows controlled-documents' collaborators (taxonomy, documents, approval, idempotency) and the two owned Postgres tables.

```mermaid
C4Context
    title System Context — Controlled Documents (module-scoped)
    Person(author, "Author / QA", "Web client")
    System_Boundary(b1, "MetalDocs") {
        System(controlleddocuments, "Controlled Documents", "Controlled-document catalog: code generation + lifecycle")
        System_Ext(taxonomy, "Taxonomy", "Profiles + Areas (FK targets)")
        System_Ext(documents, "Documents", "Implements DocumentInitializer port; owns content")
        System_Ext(approval, "Approval", "Reads CD via approval_instances")
        System_Ext(idempotency, "platform/idempotency", "Replay store")
    }
    SystemDb_Ext(db, "Postgres", "controlled_documents, cd_sequence_counters")
    Rel(author, controlleddocuments, "HTTP /api/v1/controlled-documents/*")
    Rel(controlleddocuments, taxonomy, "ProfileReader / AreaReader (Go calls)")
    Rel(controlleddocuments, documents, "DocumentInitializer.CloneTemplate (in-tx)")
    Rel(controlleddocuments, idempotency, "Require middleware")
    Rel(controlleddocuments, db, "SQL")
    Rel(approval, controlleddocuments, "reads controlled_documents via cross-module SQL")
```

### 3.1 Business Context

Controlled-documents owns the **identity** of every controlled document under QMS. A row exists from the moment a numbered slot is issued; it is the durable anchor that approval, audit, and downstream PDFs key against. Taxonomy owns the abstract classification (families -> profiles -> areas); controlled-documents owns the concrete catalog.

### 3.2 Technical Context

**Inbound:** 8 HTTP routes under `/api/v1/controlled-documents/*` (see section 5.3). Go consumers: `internal/modules/documents` (imports `controlleddocumentsdomain` for `ControlledDocument`, `DocumentInitializer`, `DocumentRef`, `CloneTemplateRequest`).

**Outbound:** Postgres (`controlled_documents`, `cd_sequence_counters`, reads of `documents`, `approval_instances`, `document_revisions`, `document_profiles`, `document_process_areas`, `templates_template_version`). Cross-module Go: `taxonomy/domain`, `taxonomy/application` (governance logger), `platform/idempotency`, `platform/authn`, `platform/httpresponse`, `platform/tenant`.

---

## 4. Solution Strategy

- **Atomic multi-row create in a single `*sql.Tx`**          driver: ADR 0011 (eliminate orphan slot risk). Controlled-documents opens the tx, allocates the sequence, inserts the CD, calls `DocumentInitializer.CloneTemplate` (documents materializes the first revision inside the same tx), commits, then emits governance events.
- **Controlled-documents-owned port for the cross-module call**          driver: avoid circular imports. `domain/document_initializer.go:61` defines the interface; `documents/application/cd_initializer.go` implements it; main.go injects via `WithDocumentInitializer` post-construction.
- **Per-(tenant, profile, area) sequence counter table**          driver: ADR 0011 (replaces single `profile_sequence_counters` keyed only on profile; no cross-area counter bleed).
- **Shared idempotency platform on POST create + revisions**          driver: ADR 0011 (Stripe-style key replay; body-hash conflict detection).
- **FULL OUTER JOIN for active-document lookup** - driver: E10 fix; published-only CDs must return 200 with `published_document_id` so frontend renders download link.

---

## 5. Building Block View — module-scoped (C4 Level 2 — Container)

> System-level container topology lives in [`wiki/diagrams/c4-container-backend.md`](../diagrams/c4-container-backend.md). The diagram below decomposes the internal Go packages of controlled-documents (service/repository/sequence allocator/template-version checker/cross-module ports).

### 5.1 Whitebox — Controlled Documents

```mermaid
C4Container
    title Container View — Controlled Documents (module-internal packages)
    Container(http, "HTTP Handlers", "Go (net/http + oapi-codegen)", "8 routes under /api/v1/controlled-documents")
    Container(svc, "ControlledDocumentsService", "Go", "Create / CreateRevision / Obsolete / Supersede / List / Get / PreviewCode")
    Container(repo, "PostgresControlledDocumentRepository", "Go + database/sql", "CRUD; standalone Create via platform/db.TxRunner; CreateTx accepts caller tx")
    Container(seq, "PostgresSequenceAllocator", "Go + database/sql", "EnsureCounter / NextAndIncrement / Peek")
    Container(tpl, "PostgresTemplateVersionChecker", "Go + database/sql", "Validates override template state")
    Container(taxRead, "TaxonomyProfileReader / AreaReader", "Go + database/sql", "Tenant-scoped FK validation")
    Container(initPort, "DocumentInitializer port", "Go interface", "Controlled-documents-owned; implemented by documents")
    ContainerDb(db1, "controlled_documents", "Postgres", "tenant_id, code, status, owner")
    ContainerDb(db2, "cd_sequence_counters", "Postgres", "next_seq per (tenant, profile, area)")
    Container_Ext(idemp, "platform/idempotency", "Go", "POST replay store")
    Container_Ext(govLog, "taxonomy AuditGovernanceAdapter", "Go", "audit_events sink via auditdomain.Writer (Wave 2.12: DBGovernanceLogger deleted)")
    Rel(http, svc, "calls")
    Rel(http, idemp, "Require middleware on POST")
    Rel(svc, repo, "CreateTx / UpdateStatus / GetByID / List")
    Rel(svc, seq, "NextAndIncrement")
    Rel(svc, tpl, "GetTemplateVersionState")
    Rel(svc, taxRead, "GetByCode")
    Rel(svc, initPort, "CloneTemplate (in-tx)")
    Rel(svc, govLog, "Log (post-commit)")
    Rel(repo, db1, "SQL")
    Rel(seq, db2, "SQL")
```

### 5.2 Public surface (selected)

Full list in `_artifacts/01-surface.md` (92 exported symbols). Anchors below:

| File | Symbol | Kind | Purpose |
|---|---|---|---|
| `module.go:15` | `Module`, `New` | type + ctor | DI entry; builds repo, allocator, readers, service, handler |
| `application/service.go:38` | service struct | type | Use-case orchestrator (legacy literal struct identifier noted in Historical Literal Key Notes) |
| `application/service.go:146` | `Create` | method | Atomic CD + first-revision create |
| `application/service.go:389` | `PreviewCode` | method | Read-only next-code peek |
| `application/service.go:451` | `Obsolete` | method | Lifecycle transition via `changeStatus` |
| `application/service.go:455` | `Supersede` | method | Lifecycle transition via `changeStatus` |
| `application/service.go:584` | `CreateRevision` | method | New revision on existing CD |
| `domain/controlled_document.go:18` | `ControlledDocument` | struct | Domain entity |
| `domain/controlled_document.go:13-15` | `CDStatusActive/Obsolete/Superseded` | const | Status enum |
| `domain/controlled_document.go:48` | `AutoCode` | func | Format `{profile}-{area}-{NNN}` |
| `domain/document_initializer.go:61` | `DocumentInitializer` | iface | Cross-module port (documents implements) |
| `domain/document_initializer.go:27` | `NewCloneTemplateRequest` | func | Validates clone-template requests before crossing the documents port |
| `domain/document_initializer.go:51` | `DocumentRef` | struct | Handle returned by the documents port after a successful atomic create |
| `domain/sequence.go:13` | `SequenceAllocator` | iface | Counter port |
| `domain/resolution.go:30` | `Resolve` | func | Template-version resolution (default vs override) |
| `infrastructure/repository.go:36,48` | `PostgresControlledDocumentRepository`, `NewPostgresControlledDocumentRepository` | type + ctor | Holds the SQL pool and a `platform/db.TxRunner` |
| `infrastructure/repository.go:385` | `Create` | method | Standalone insert in a repository-owned `runner.Do` transaction, including tier-2 authz |
| `infrastructure/repository.go:405` | `CreateTx` | method | Insert in caller-owned `db.Tx`; authz remains the service's mandatory precondition |
| `infrastructure/repository.go:484` | `UpdateStatus` | method | Lifecycle UPDATE |
| `infrastructure/repository.go:582` | `GetActiveInstance` | method | Active+published document projection via the documents-owned read port |
| `infrastructure/repository.go:661` | `NextAndIncrement` | method | Sequence allocation |
| `domain/port.go:8` | `ActiveDocumentInstance` | struct | Active+published document result type (Wave 2) |
| `domain/controlled_document.go:37` | `ErrNoActiveInstance` | sentinel | No active document instance (Wave 2) |
| `application/service.go:497` | `GetActiveInstance` | method | Read-check + repo delegate (Wave 2) |

### 5.3 HTTP operations

All routes registered via `Handler.RegisterRoutes` (`delivery/http/handler.go:67`) onto the shared `http.ServeMux`.

| Method | Path | OperationID | Handler | Authz |
|---|---|---|---|---|
| POST | `/api/v1/controlled-documents` | `atomicCreateControlledDocument` | `routes.go:70` | create capability (tier-1 + in-tx `authz.Require`; sets `metaldocs.tenant_id`/`actor_id` before sequence/CD writes) |
| POST | `/api/v1/controlled-documents/{id}/revisions` | `createControlledDocumentRevision` | `routes.go:197` | create capability (tier-1) |
| GET | `/api/v1/controlled-documents/preview-code` | `previewControlledDocumentCode` | `routes.go:172` | (read; resolver mapping outside module) |
| GET | `/api/v1/controlled-documents` | `listControlledDocuments` | `routes.go:23` | (read) |
| GET | `/api/v1/controlled-documents/{id}` | `getControlledDocument` | `routes.go:190` | (read) |
| GET | `/api/v1/controlled-documents/{id}/active-document` | `getActiveDocument` | `routes.go:266` | (read; tenant from `tenant.FromContext` via `injectTenant` middleware) |
| PUT | `/api/v1/controlled-documents/{id}/obsolete` | `obsoleteControlledDocument` | `routes.go:441` | obsolete capability (legacy literal constant; see Historical Literal Key Notes)          T-001 closed Plan 5 |
| PUT | `/api/v1/controlled-documents/{id}/supersede` | `supersedeControlledDocument` | `routes.go:455` | supersede capability (legacy literal constant; see Historical Literal Key Notes)          T-001 closed Plan 5 |

## API Route Truth Table (Plan 8 Baseline)

| Method | Path | Runtime owner (file:line) | Handler method | Spec path | operationId | Codegen method | Status | Notes |
|---|---|---|---|---|---|---|---|---|
| GET | `/api/v1/controlled-documents` | `internal/modules/controlleddocuments/delivery/http/handler.go:95` | `controlleddocumentsapi.HandlerWithOptions` dispatch | `/api/v1/controlled-documents` | `listControlledDocuments` | `ListControlledDocuments` | Aligned | Generated boundary mounted |
| POST | `/api/v1/controlled-documents` | `internal/modules/controlleddocuments/delivery/http/handler.go:95` | `controlleddocumentsapi.HandlerWithOptions` dispatch | `/api/v1/controlled-documents` | `atomicCreateControlledDocument` | `AtomicCreateControlledDocument` | Aligned | Idempotency middleware wraps POST path |
| GET | `/api/v1/controlled-documents/preview-code` | `internal/modules/controlleddocuments/delivery/http/handler.go:95` | `controlleddocumentsapi.HandlerWithOptions` dispatch | `/api/v1/controlled-documents/preview-code` | `previewControlledDocumentCode` | `PreviewControlledDocumentCode` | Aligned |  |
| GET | `/api/v1/controlled-documents/{id}` | `internal/modules/controlleddocuments/delivery/http/handler.go:95` | `controlleddocumentsapi.HandlerWithOptions` dispatch | `/api/v1/controlled-documents/{id}` | `getControlledDocument` | `GetControlledDocument` | Aligned |  |
| POST | `/api/v1/controlled-documents/{id}/revisions` | `internal/modules/controlleddocuments/delivery/http/handler.go:95` | `controlleddocumentsapi.HandlerWithOptions` dispatch | `/api/v1/controlled-documents/{id}/revisions` | `createControlledDocumentRevision` | `CreateControlledDocumentRevision` | Aligned | Idempotency middleware wraps POST path |
| GET | `/api/v1/controlled-documents/{id}/active-document` | `internal/modules/controlleddocuments/delivery/http/handler.go:95` | `controlleddocumentsapi.HandlerWithOptions` dispatch | `/api/v1/controlled-documents/{id}/active-document` | `getActiveDocument` | `GetActiveDocument` | Aligned |  |
| PUT | `/api/v1/controlled-documents/{id}/obsolete` | `internal/modules/controlleddocuments/delivery/http/handler.go:95` | `controlleddocumentsapi.HandlerWithOptions` dispatch | `/api/v1/controlled-documents/{id}/obsolete` | `obsoleteControlledDocument` | `ObsoleteControlledDocument` | Aligned |  |
| PUT | `/api/v1/controlled-documents/{id}/supersede` | `internal/modules/controlleddocuments/delivery/http/handler.go:95` | `controlleddocumentsapi.HandlerWithOptions` dispatch | `/api/v1/controlled-documents/{id}/supersede` | `supersedeControlledDocument` | `SupersedeControlledDocument` | Aligned |  |

Module contract status: Generated boundary mounted
Owner: leandro

---

## 6. Runtime View

### 6.1 atomicCreateControlledDocument

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant Idem as idempotency.Require
    participant H as Handler
    participant S as ControlledDocumentsService
    participant Seq as SequenceAllocator
    participant Repo as CDRepository
    participant Init as DocumentInitializer
    participant DocRepo as documents.CreateDocumentTx
    participant Log as govLogger
    C->>Idem: POST /controlled-documents (Idempotency-Key, body)
    Idem->>Idem: BeginReplay (body hash — idempotency.Require middleware)
    Idem->>H: forward
    H->>S: Create(cmd)
    S->>Repo: TaxonomyProfileReader.GetByCode
    S->>Repo: TaxonomyAreaReader.GetByCode
    S->>S: BeginTx
    S->>S: setAuthzGUC(tenant_id, actor_id)
    S->>S: authz.Require(<create capability>, tenant)
    S->>Seq: NextAndIncrement(tx)
    S->>Repo: CreateTx (controlled_documents)
    S->>Init: CloneTemplate(tx, ...)
    Init->>DocRepo: CreateDocumentTx
    DocRepo-->>Init: documentID
    Init-->>S: DocumentRef
    S->>S: Commit
    S->>Log: emit governance events
    S-->>H: CreateResult
    H-->>Idem: 201 AtomicCreateResponse
    Idem->>Idem: CompleteReplay
    Idem-->>C: 201
```

State transitions:

| Entity | From | To | Trigger | Capability |
|---|---|---|---|---|
| `controlled_documents` (new row) | (none) | `active` | `CreateTx` | create capability (tier-1 IAM middleware + in-tx `authz.Require`) |
| `documents` (new row) | (none) | `draft` | `CreateDocumentTx` (in same tx) | same |
| `cd_sequence_counters.next_seq` | N | N+1 | `NextAndIncrement` | same |
| `governance_events` | (none) | event row (post-commit) | `govLogger.Log` | same |

Failure modes:

| Condition | HTTP | Code |
|---|---|---|
| Missing `Idempotency-Key` | 400 | `IDEMPOTENCY_KEY_REQUIRED` |
| Same key, different body | 422 | `IDEMPOTENCY_KEY_CONFLICT` |
| Profile/area archived | 409 | `PROFILE_ARCHIVED` / `AREA_ARCHIVED` |
| Override template mismatch | 409 | `TEMPLATE_PROFILE_MISMATCH` |
| Code uniqueness violation | 409 | `CONTROLLED_DOCUMENT_CODE_TAKEN` |
| Template/profile mismatch | 422 | `template_invalid` (`routes.go:560-561`          T-007 closed Plan 7) |

Detail: `_artifacts/02-flow-atomic-create.md`.

### 6.2 getActiveDocument

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as Handler
    participant S as ControlledDocumentsService
    participant R as PostgresControlledDocumentRepository
    participant DB as Postgres
    C->>H: GET /controlled-documents/{id}/active-document  (tenant from context via injectTenant)
    H->>S: GetActiveInstance(tenantID, id)
    S->>S: authz read-check (CanDo / actor context)
    S->>R: GetActiveInstance(ctx, tenantID, id)
    R->>DB: FULL OUTER JOIN (documents active LEFT JOIN documents published) + approval lookup
    DB-->>R: ActiveDocumentInstance or nil
    R-->>S: *ActiveDocumentInstance | nil
    S-->>H: *ActiveDocumentInstance | ErrNoActiveInstance
    alt nil (no active/published)
        H-->>C: 404
    else
        H-->>C: 200 ActiveDocumentResponse (all fields optional)
    end
```

Wave 2 note: delivery layer is now SQL-free for this path — the active-instance projection is delegated to `PostgresControlledDocumentRepository.GetActiveInstance` (`infrastructure/repository.go:582`) and orchestrated via `ControlledDocumentService.GetActiveInstance`. The domain type `ActiveDocumentInstance` (`domain/port.go:8`) and sentinel `ErrNoActiveInstance` (`domain/controlled_document.go:37`) complete the extraction.

State transitions: none (read).

Tripwire pairing: VIOLATION          no `authz.Require`, no `metaldocs.assert_caps`, no GUC; tenant is now sourced from `tenant.FromContext` (via `injectTenant` middleware) rather than `X-Tenant-ID` header (Plan 3 fix). Authz gap on this read path persists          see T-006.

Detail: `_artifacts/02-flow-get-active.md`.

### 6.3 obsoleteControlledDocument / supersedeControlledDocument

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as Handler
    participant S as ControlledDocumentsService
    participant Repo as CDRepository
    participant DB as Postgres
    C->>H: PUT /controlled-documents/{id}/obsolete
    H->>S: Obsolete(tenant, id)
    S->>Repo: GetByID
    Repo->>DB: SELECT controlled_documents
    DB-->>Repo: row
    Repo-->>S: doc
    alt doc.IsActive() == false
        S-->>H: ErrCDNotActive
        H-->>C: 409 CONTROLLED_DOCUMENT_NOT_ACTIVE
    else
        S->>Repo: UpdateStatus(obsolete) [in tx]
        Repo->>DB: UPDATE controlled_documents
        DB-->>Repo: ok
        S->>S: govLogger.LogTx(ctx, tx, event) [in-tx governance — Wave 2]
        S->>DB: COMMIT
        Repo-->>S: ok
        S-->>H: nil
        H-->>C: 204
    end
```

State transitions:

| Entity | From | To | Trigger | Guard | Audit emitted? |
|---|---|---|---|---|---|
| `controlled_documents` | `active` | `obsolete` | `Obsolete` op | `ErrCDNotActive` if not active | **YES          T-002 closed Plan 6a** |
| `controlled_documents` | `active` | `superseded` | `Supersede` op | `ErrCDNotActive` if not active | **YES          T-002 closed Plan 6a** |

Tripwire pairing: active          `authz.Require(<obsolete-or-supersede capability>, ...)` called inside `changeStatus` tx (`service.go:534`); `trg_require_cap_asserted` on `controlled_documents` (UPDATE, OR-logic accepts either lifecycle capability, `db/baseline/0001_current_schema.sql:3783-3786`). T-001/T-004 closed Plan 5.

Detail: `_artifacts/02-flow-obsolete.md`.

---

## 7. Deployment View

- Binary: single Go server (`apps/api/cmd/metaldocs-api`)
- Process: `:8081` (see [wiki/references/local-dev-startup.md](references/local-dev-startup.md))
- Migrations: applied at startup; historical controlled-documents schema is consolidated into `db/baseline/0001_current_schema.sql`; incremental migrations in `db/migrations/` (select post-baseline entries: `0210_controlled_documents_capability_namespace.sql`, `0225_authz_p2_document_lifecycle_grants.sql`, `0229_authz_p12_rename_document_lifecycle_caps.sql`, `0234_rls_controlled_documents_audit_events.sql` (Wave 2: ENABLE + FORCE ROW LEVEL SECURITY + NULL-permissive `tenant_isolation` policy on `public.controlled_documents`; ADR 0027 Tier 1, F-12, D-3, REQ-TEN-1; effective for NOSUPERUSER+NOBYPASSRLS prod role only — dev superuser BYPASSRLS skips policy) — see `_artifacts/04-persistence.md` section 6)
- Legacy controlled-documents maintenance (`application/migration.go`, `BackfillLegacyDocuments`) has been removed from the codebase. No legacy maintenance hook exists in the current module.
- Environment: module reads no env vars directly (`_artifacts/03-deps.md` section 4)

---

## 8. Cross-cutting Concepts

### 8.1 Authentication & Authorization

- Tier 1 (HTTP edge): IAM `CapabilityService` resolves create capability (legacy literal constant; see Historical Literal Key Notes) for POST routes (`apps/api/cmd/metaldocs-api/permissions.go:185-186`; reseeded in `migrations/0165_role_capabilities_reseed.sql` for `editor`, `author`, `system_admin`).
- Tier 2 (in-tx `authz.Require`): standalone repository `Create` checks the create capability inside its `runner.Do` transaction (`repository.go:385-396`). The normal service-owned paths check once before calling `CreateTx` (`service.go:480-494,559-664`); `CreateTx` deliberately does not duplicate that query. `changeStatus` checks obsolete/supersede capabilities. Plan 5 (T-001/T-004 closed).
- Tier 3 (Postgres `enforce_capability_asserted` tripwire): `db/baseline/0001_current_schema.sql:3783-3786` (consolidated; migration 0188 no longer exists as a standalone file) attaches `trg_require_cap_asserted` to `controlled_documents` (INSERT + UPDATE with OR-logic) and `cd_sequence_counters` (`0001_current_schema.sql:3776-3779`).
- See [wiki/concepts/authz-tiers.md](concepts/authz-tiers.md) and [wiki/decisions/0007-two-tier-authz.md](decisions/0007-two-tier-authz.md).

### 8.2 Error envelope

RFC 9457 `application/problem+json`. `httpresponse.WriteError` at `internal/platform/httpresponse/response.go:16-18` calls `problem.Write(w, problem.New(status, code, message))`. `ErrTemplateProfileMismatch` uses a direct `problem.Write` call at `routes.go:560-561` (422 `template_invalid`). T-003 and T-007 both closed Plan 7.

### 8.3 Idempotency

`Idempotency-Key` header required on POST create + POST revisions. Middleware: `internal/platform/idempotency/middleware.go:80` (`Require`). Store: `postgres_store.go:74` (`BeginReplay`). Body-hash conflict → 422 `IDEMPOTENCY_KEY_CONFLICT`. PUT lifecycle routes are NOT covered (T-008).

### 8.4 Pagination

`List` uses domain `CDFilter` (`domain/port.go:19`)          simple LIMIT/OFFSET. No cursor pagination (not required by current consumers).

### 8.5 Audit / Governance

Create path emits governance events post-commit via `s.govLogger.Log(...)` (`service.go:267-271`). Wave 2: `changeStatus` now emits `govLogger.LogTx(ctx, tx, event)` **before commit** (`service.go:583`) — governance write is now in-tx for lifecycle transitions. Wired from `taxonomyapp.NewAuditGovernanceAdapter(deps.AuditWriter)` (`module.go:35`) — cross-module sink coupling (T-008). Wave 2: `module.go:27` now panics when `AuditWriter` is nil (`DBGovernanceLogger` nil-fallback removed — fail-loud by design). `taxonomydomain` import removed from module.go (no longer needed after fallback deletion). Obsolete / Supersede audit gap (T-002) is closed per Plan 6a; see tech-debt register.

### 8.6 Concurrency / Transactions

The standard application flow is service-owned: `ControlledDocumentService` uses `platform/db.TxRunner.Do` and passes the resulting `*sql.Tx` through `CreateTx`, sequence allocation, and `DocumentInitializer.CloneTemplate`, preserving the atomic CD + first-revision guarantee from ADR 0011. The repository also exposes standalone `Create`; its constructor now creates a `TxRunner`, and `Create` uses `runner.Do` for begin/commit/rollback plus request-identity seeding. `CreateTx(ctx, db.Tx, doc)` remains the explicit caller-owned seam and performs storage only after the service's mandatory in-tx authz gate.

### 8.7 Tenant scoping

`controlled_documents.tenant_id NOT NULL`; `cd_sequence_counters` PK includes `tenant_id`. Indexes lead with `tenant_id` (`ix_controlled_documents_tenant_area`, `ix_controlled_documents_tenant_profile`). Tenant is enforced via query-argument predicate in every WHERE clause. Wave 2 (migration 0234): `public.controlled_documents` now has ENABLE + FORCE ROW LEVEL SECURITY with a NULL-permissive `tenant_isolation` policy (`current_setting('metaldocs.tenant_id', true)` compared to `tenant_id::text`; NULL GUC = no restriction on system paths without GUC). RLS is effective only for NOSUPERUSER+NOBYPASSRLS production roles (dev Docker metaldocs_app is superuser and bypasses). T-005 (GUC+RLS backstop absent) is partially addressed for `controlled_documents` — `cd_sequence_counters` not yet covered by RLS. `migrations/0127_documents_v2_tenant_consistency_trigger.sql` cross-checks tenant on `documents_v2` writes (legacy bridge).

Tenant is sourced from `tenant.FromContext` via the `injectTenant` thin middleware (`delivery/http/handler.go:48`), which reads the value injected by auth middleware (from `auth_sessions.tenant_id`). This replaces the prior `X-Tenant-ID` header reads (Plan 3 sweep). See `wiki/architecture/tenant-context.md`.

### 8.8 Numbering invariants

- `cd_sequence_counters.next_seq` is monotonic per (tenant, profile_code, process_area_code).
- Code format `{PROFILE}-{AREA}-{NNN}` zero-padded to 3 digits (`domain/sequence.go` -> `AutoCode`).
- `controlled_documents.code` immutability is enforced by the `trg_controlled_documents_code_immutable` trigger calling `reject_code_update()` (`db/baseline/0001_current_schema.sql:3755-3758`; historical origin in `archive/migrations/0124_registry_controlled_documents.sql`).

---

## 9. Architecture Decisions

| Decision | Link / Status |
|---|---|
| Atomic CD + first-revision create + per-area numbering + `Idempotency-Key` | [wiki/decisions/0011-cd-atomic-create.md](decisions/0011-cd-atomic-create.md) |
| Two-tier authz | [wiki/decisions/0007-two-tier-authz.md](decisions/0007-two-tier-authz.md) (tier-2/3 wired Plan 5          T-001/T-004 closed) |
| Contract-first API (OpenAPI + oapi-codegen) | [wiki/decisions/0012-contract-first-api.md](decisions/0012-contract-first-api.md) (422 `template_invalid` spec/handler drift          **CLOSED Plan 7 T-007**) |
| Which CD lifecycle events emit audit | tech-debt: missing-ADR (T-002) |
| Capability granularity (separate create / obsolete / supersede capabilities) | implemented Plan 5 (migration 0187 + dedicated lifecycle capability constants in `domain/model.go`); missing standalone ADR          ADR-TODO per Plan 13 |
| RFC 9457 envelope adoption | **CLOSED Plan 7** (T-003 + T-007) |
| GUC-based tenant scoping vs query-arg only | tech-debt: missing-ADR (T-005) |
| Read-path authz contract (e.g. `GetActiveDocument` tenant source) | tech-debt: missing-ADR (T-006) |
| Where controlled-documents audit sink should live (own logger vs shared taxonomy sink) | implementation debt closed in tech-debt (T-008); standalone ADR still missing |
| Documents DI cycle resolution (`WithDocumentInitializer` setter) | tech-debt: missing-ADR (T-010) |
| `PostgresTemplateVersionChecker` deleted; CD now consumes templates-owned `TemplateVersionPort.GetTemplateVersionState` port (consumer; `status := "published"` hardcode removed — M4/F4.2) | [`wiki/decisions/0030-template-version-state-port.md`](../decisions/0030-template-version-state-port.md) |
| OpenAPI partial directory (`v1/` for `/api/v1/` routes) | tech-debt: missing-ADR (T-011) |
| Exported-symbol doc-comment policy | tech-debt: missing-ADR (T-012) |

---

## 10. Quality Requirements

| Goal | Scenario | Pass criteria |
|---|---|---|
| Numbering integrity | Two concurrent `POST /controlled-documents` for same (profile, area) | Both succeed with distinct sequential codes; `cd_sequence_counters.next_seq` advances by 2 |
| No orphan slot | `CreateDocumentTx` returns error mid-tx | `controlled_documents` row NOT visible; `next_seq` NOT advanced |
| Replay safety | Replay POST with same `Idempotency-Key` and body | 201 with original CD payload; no new row |
| Body-hash conflict | Same key, different body | 422 `IDEMPOTENCY_KEY_CONFLICT`; no new row |
| Active-document publish-only | CD with only published revisions | 200 with `published_document_id` set; `document_id` omitted |
| Active-document empty CD | CD with no revisions | 404 |
| Lifecycle guard | `PUT .../obsolete` on already-obsolete CD | 409 `CONTROLLED_DOCUMENT_NOT_ACTIVE` |
| Code immutability | Direct SQL `UPDATE controlled_documents SET code = ...` | Postgres trigger raises (`reject_code_update`) |

---

## 11. Risks & Technical Debt

Detail in [wiki/modules/controlled-documents-tech-debt.md](modules/controlled-documents-tech-debt.md). Severity rubric: see top of that file (concrete triggers; do not invent local definitions).

- Critical: 2
- Major: 7
- Minor: 5

Top 3 (by severity, then blast-radius):

1. T-006          `GetActiveDocument` has no authz check for the read path; residual gap after Plan 3 header-trust fix.
2. T-005          Tenant scoping relies on query-arg only; no GUC + RLS backstop on controlled-documents-owned tables.
3. T-009          DI cycle resolved via post-construction setter; latent order-of-construction contract remains.
(T-001 closed Plan 5: lifecycle authz wired. T-002 closed Plan 6a: obsolete/supersede audit gap closed. T-004 closed Plan 5: tripwire attached to `controlled_documents` + `cd_sequence_counters`. T-008 closed Plan 6a per tech-debt register.)

### Coverage stats

- Public symbols undocumented: 79 / 94 (computed from `_artifacts/01-surface.md`; deferred to T-012; `application/migration.go` and `BackfillLegacyDocuments` removed — file no longer exists)
- Operations missing C4 placement: 0 / 8
- Cross-deps missing in section 5/section 8: 0 / 13 (IN-edges) + 0 / 6 (OUT-edges)
- State transitions missing in section 6: 0 / 2 (Obsolete, Supersede both in section 6.3; Create in section 6.1)
- Decisions without ADR link: 9 / 12

---

## 12. Glossary

| Term | Definition |
|---|---|
| Controlled Document (CD) | A numbered slot in `controlled_documents` binding (profile_code, process_area_code) to a chain of `documents` revisions |
| Slot | A CD row before/without document content; the durable identity |
| Sequence counter | `cd_sequence_counters.next_seq`, monotonic per (tenant, profile, area) |
| AutoCode | Format `{PROFILE}-{AREA}-{NNN}` (e.g. `DC-RH-001`) |
| Active document | The current non-terminal `documents` row for a CD (draft/under_review/approved/rejected/scheduled) OR the most recent `published` row |
| `DocumentInitializer` | Controlled-documents-owned port the documents module implements to materialize the first revision inside the controlled-documents tx |

---

## Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Postgres unavailable mid atomic-create | Tx aborts; no CD row, no first revision — caller sees 500 | `application/service.go:146` `Create` rolls back the whole tx | Per ADR 0011: atomic guarantee — no orphan slots possible by design |
| Sequence counter collision (concurrent create on same `(tenant, profile, area)`) | One side wins via `NextAndIncrement`; loser retries | `infrastructure/repository.go:661` uses `SELECT … FOR UPDATE` plus `UPDATE … RETURNING` in the caller tx | Caller retries with fresh `Idempotency-Key`; race is bounded |
| Idempotency replay (`Idempotency-Key` header reused) | 201 with stored response; no duplicate CD/document | `platform/idempotency.Require` middleware | Expected; safe network-retry path |
| Body-hash mismatch on replay | 409 from idempotency middleware | Idempotency store detects payload change | Caller must use a fresh key for a different payload |
| FK validation: profile / area unknown | 4xx from `TaxonomyProfileReader` / `AreaReader` | `taxRead.GetByCode` returns not-found | Confirm taxonomy data; backfill missing rows |
| Template-version state invalid for override | Atomic create rejects with conflict | `PostgresTemplateVersionChecker.GetTemplateVersionState` returns non-publishable state | Caller picks a published template version |
| `DocumentInitializer.CloneTemplate` fails inside tx | Whole atomic create rolls back; no CD; caller sees 500 | Cross-module port returns err; tx rollback | Investigate documents-module clone error; replay with same idempotency key |
| Obsolete/Supersede emit no audit event | Audit trail missing lifecycle entry | T-002 (registry-legacy critical) — known gap | Wire `auditdomain.Writer.Record` on `changeStatus` |
| Active-document lookup returns nothing for published-only CD | Frontend cannot render download link | Before E10 fix: missing row; now mitigated by FULL OUTER JOIN in `getActiveDocument` | Regression test on published-only CDs |
| Sequence counter not initialized for new `(profile, area)` | First create on a pair must call `EnsureCounter` | `service.go:146` ensures counter before `NextAndIncrement` | Self-healing — counter row created lazily |
| Cross-module governance log lag | Governance event for create lands after CD row | Logger writes post-commit (taxonomy `DBGovernanceLogger`) | Replay outbox if event missing; T-008 tracks dual-sink coupling |

## Cross-links

- Related ADRs: [0011](decisions/0011-cd-atomic-create.md), [0007](decisions/0007-two-tier-authz.md), [0012](decisions/0012-contract-first-api.md)
- Related concepts: [controlled-documents](concepts/controlled-documents.md), [authz-tiers](concepts/authz-tiers.md)
- Related modules: [documents](modules/documents.md), [taxonomy](modules/taxonomy.md), [approval](modules/approval.md)
- Workflow: [user-onboarding](workflows/user-onboarding.md) Step 5
- Backlog: [controlled-documents-refactor](backlog/controlled-documents-refactor.md)
- Tech debt: [controlled-documents-tech-debt](modules/controlled-documents-tech-debt.md)

### Historical Literal Key Notes

- Legacy literal capability keys in historical artifacts/migrations: `registry.create`, `registry.obsolete`, `registry.supersede`.
- Legacy literal service/capability identifiers preserved in historical references: `RegistryService`, `CapRegistryCreate`, `CapRegistryObsolete`, `CapRegistrySupersede`.

## Changelog

- 2026-06-12 - Wave 2 module sync: `GetActiveDocument` delivery handler extracted — inline SQL moved to `PostgresControlledDocumentRepository.GetActiveInstance` (`infrastructure/repository.go:517`) and `ControlledDocumentService.GetActiveInstance` (`application/service.go:497`); new domain type `ActiveDocumentInstance` (`domain/port.go:8`) and sentinel `ErrNoActiveInstance` (`domain/controlled_document.go:37`); delivery layer now SQL-free for this path. `changeStatus` now calls `govLogger.LogTx(ctx, tx, event)` before commit (`service.go:583`) — governance write is in-tx for lifecycle transitions. `module.go:27` now panics when `AuditWriter` is nil (`DBGovernanceLogger` nil-fallback removed; `taxonomydomain` import dropped from module.go). Migration 0234: ENABLE + FORCE RLS + NULL-permissive `tenant_isolation` policy on `public.controlled_documents` (ADR 0027 Tier 1). §5.2 public surface, §6.2 flow, §6.3 flow, §7 deployment, §8.5 governance, §8.7 tenant scoping, key files all updated.
- 2026-05-29 - Frontend surface removed: legacy `/controlled-documents` list/detail/explorer pages and Rail "Registro" nav entry deleted. CD identity and lifecycle (create / revise / publish / obsolete / supersede) are now driven exclusively through the Documents flow (`/documents`, `/documents/new`, `/documents/:id`). Backend module (routes, contract, authz, repositories) is unchanged; FE still consumes `features/controlled-documents/api`, `queries/`, and `types.ts` from the Documents flow.
- 2026-05-25 - Backend quality-bar sync: repository transaction ports now use the module-level `domain.DBTX` interface instead of exposing `*sql.Tx` in repository contracts; clone-template/document-ref constructors validate invalid zero-value port payloads; application/repository error paths wrap underlying errors with operation context and governance warnings use `slog.WarnContext` with tenant/actor fields.
- 2026-05-21 - Runtime mount canonicalization: controlled-documents now mounts public routes through `controlleddocumentsapi.HandlerWithOptions`; idempotency remains route-scoped to the two POST operations, and missing `Idempotency-Key` is normalized to `IDEMPOTENCY_KEY_REQUIRED`.
- 2026-05-20 - Create-revision conflict sync: `POST /api/v1/controlled-documents/{id}/revisions` now preserves the database-owned single-active-sibling invariant (`ux_documents_cd_active`) but translates that collision to `409 ACTIVE_REVISION_ALREADY_EXISTS` instead of surfacing a generic internal error when a second active revision is attempted concurrently.
- 2026-05-20 - Active-document approval-instance hardening: `GET /api/v1/controlled-documents/{id}/active-document` now treats `documents.status` as the only source of truth for `approval_state`, enriches `approval_instance_id` only when the active lineage row is actually `under_review`, and returns `500 INTERNAL_ERROR` if that secondary lookup fails instead of silently omitting review context.
- 2026-05-20 - Active-document scheduled-state sync: `GET /api/v1/controlled-documents/{id}/active-document` now derives `approval_state` from the governed `documents.status` of the active lineage row, so a scheduled replacement remains visible to `/documents/:id` as `scheduled` instead of drifting back to `approved`.
- 2026-05-20 - Canonical sibling-state sync: `GET /api/v1/controlled-documents/{id}/active-document` remains the technical lookup consumed by `/documents/:id` to decide whether a revision branch is open; frontend now treats every returned active approval state (`draft`, `under_review`, `approved`, `scheduled`, `rejected`) as branch-active context.
- 2026-05-20 - Active-document publish-only sync: `GET /api/v1/controlled-documents/{id}/active-document` no longer synthesizes `approval_state="draft"` when a controlled document has only a published revision. The controlled-documents FULL OUTER JOIN now leaves the active side absent in publish-only state and returns just `published_document_id`, matching the OpenAPI contract and the canonical `/documents/:id` publish flow.
- 2026-05-18 - Frontend contract sync: the controlled-document detail contract remains the source of truth for `visibility`, and the editor sidebar now consumes that runtime field directly through generated frontend types instead of a controlled-documents-local handwritten omission.
- 2026-05-15 - Database foundation sync: removed startup migration alias references (`RunStartupMigrations`), confirmed legacy maintenance is explicit recovery-only, and aligned startup notes with the current DB bootstrap workflow.
- 2026-05-15 - Runtime repair: atomic create now primes `metaldocs.tenant_id`/`metaldocs.actor_id` and asserts `registry.create` (legacy literal capability key) inside the caller-owned transaction before sequence/CD writes, restoring Plan 5 tripwire pairing for `/api/v1/controlled-documents`.

- 2026-05-11 - Plan 3 sweep: all `X-Tenant-ID` header reads replaced with `tenant.FromContext`; `injectTenant` middleware documented; section 5.3 T-006 note updated; section 6.2 sequence + tripwire note updated; section 8.7 tenant-scoping paragraph added; Key files updated.
- 2026-05-11 - initial Arc42 + C4 publish; supersedes 2026-05-07 stub.
