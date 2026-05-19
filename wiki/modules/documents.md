# Module: documents

> Living architecture doc. Arc42 (12 sections) + C4 (Context / Container) Mermaid diagrams + ADR links.

**Last verified:** 2026-05-19 (editor sidebar identification visual sync) | **Owner:** unassigned | **Status:** active | **Maturity:** L3

---

## 1. Introduction & Goals

`documents` is the controlled-document instance lifecycle module: it owns the `documents` table and its revision/checkpoint/comment/export/approval surfaces. A document is an instance filled from a template version, bound to a controlled-document entry, and moves through `draft â†’ under_review â†’ approved â†’ published â†’ superseded | obsolete`. The module sits between the registry module (which owns the controlled-document identity + numbering) and the approval sub-package (which materialises approval routes against a document revision).

### 1.1 Requirements overview

- Atomic create of a controlled document + its first draft revision â€” `wiki/decisions/0011-cd-atomic-create.md`
- Two-tier authorization on mutations â€” `wiki/decisions/0007-two-tier-authz.md`
- Eigenpal DOCX editor integration for `draft` editing â€” `wiki/decisions/0001-eigenpal-adoption.md`
- Placeholder snapshot pinned at submit â€” `wiki/concepts/placeholders.md`
- `{name}` single-brace token substitution at freeze â€” `wiki/concepts/token-syntax.md`
- Spec-as-source-of-truth for `/api/v1/documents/*` (in progress) â€” `wiki/decisions/0012-contract-first-api.md`

### 1.2 Quality Goals

| Rank | Goal | How verified |
|---|---|---|
| 1 | Authz isolation on regulated mutations | Tier-1 role gate + tier-2 `authz.Require` + Postgres tripwire; **T-003 closed Plan 5** â€” all 5 `documents` table mutations now call `authz.Require` + tripwire attached (migration 0188) |
| 2 | Atomic CD+draft create | `CreateDocumentTx` port; one transaction across `controlled_documents`, `documents`, `cd_sequence_counters` (ADR 0011) |
| 3 | Forward-only schema evolution | All `documents`-owned migrations are append-only (`migrations/0001..0183`); rename from `documents_v2` shipped as paired migrations 0167/0168 |

### 1.3 Stakeholders

| Role | Expectation |
|---|---|
| End user (document filler / approver) | Create / edit / submit / sign-off documents; library list filtered by ownership |
| Operator | Predictable migration story; audit trail intact; tripwire defends regulated paths |
| Developer (other modules) | Stable Go ports (`CreateDocumentTx`, `CapabilityChecker`, `SnapshotResolver`); typed errors over generic strings |

---

## 2. Architecture Constraints

- Language / runtime: Go 1.25.
- Persistence: Postgres only; row-level multi-tenancy via `tenant_id` on every owned table (`wiki/architecture/persistence.md`).
- Authz: two-tier per `wiki/decisions/0007-two-tier-authz.md`. Tripwire enforcer reads `metaldocs.asserted_caps` GUC.
- DOCX editor: `@eigenpal/docx-js-editor` from `vendor/eigenpal/` (`wiki/decisions/0001-eigenpal-adoption.md`).
- API contract: OpenAPI 3.0.3 via oapi-codegen (`wiki/decisions/0012-contract-first-api.md`) â€” **bootstrap only on this module**, handlers still mounted via stdlib `mux.HandleFunc`.
- Error envelope target: RFC 9457 Problem Details â€” current code emits legacy `{error:{code,message,â€¦}}` (T-001, mid-migration).

---

## 3. System Scope & Context (C4 Level 1)

```mermaid
C4Context
    title System Context â€” documents
    Person(user, "Filler / Approver", "Web client (React)")
    System_Boundary(b1, "MetalDocs API") {
        System(docs, "documents", "Document instance lifecycle: list, create, edit, finalize, approve, publish")
        System(registry, "registry", "Controlled-document identity + numbering")
        System(templates, "templates", "Template versions consumed at draft create")
        System(iam, "iam", "Capabilities + actor identity")
        System(render, "render", "DOCX â†’ PDF rendering")
        System(idemp, "platform/idempotency", "Stripe-style replay store")
    }
    SystemDb(db, "Postgres", "documents, document_revisions, document_checkpoints, document_comments, document_exports, approval_*, governance_events, metaldocs.pdf_dispatch_outbox")
    System_Ext(objstore, "Object Store", "DOCX + PDF artifacts")
    Rel(user, docs, "HTTPS /api/v1/documents/*")
    Rel(registry, docs, "Go: CreateDocumentTx port")
    Rel(docs, templates, "Go: template version reads")
    Rel(docs, iam, "Go: CapabilityChecker / ErrCapabilityDenied")
    Rel(docs, render, "Go: resolver readers (PDF)")
    Rel(docs, idemp, "Go: signoff idempotency store")
    Rel(docs, db, "SQL (pgx)")
    Rel(docs, objstore, "presigned upload/download (via platform/objectstore)")
```

### 3.1 Business Context

Quality-managed documents. Each instance traces back to a template + a controlled-document code (e.g. `POP-MAN-027`). Drafts are editable; once submitted, the placeholder catalog is frozen and approval signatures gate publication. Audit trail is the QMS evidence.

### 3.2 Technical Context

**Inbound HTTP (from FE):** see Â§5.3.

**Inbound Go (consumers â€” from `_artifacts/03-deps.md`):**
- `apps/api/cmd/metaldocs-api/main.go:22` â€” wires services + handler + approval handler
- `apps/api/internal/wiring/documents.go:6` â€” `NewCapabilityChecker` adapter at `:24` (ADR 0007 J2)
- `apps/worker/cmd/metaldocs-worker/main.go:8` â€” worker side
- `internal/modules/iam/integration_test.go:16`
- `internal/jobs/effective_date_publisher/job.go:8` Â· `internal/jobs/stuck_instance_watchdog/job.go:10`
- `internal/platform/docgenv2/templates_snapshot_reader.go:8`
- `internal/platform/objectstore/document_presigner.go:17`

**Outbound Go (from `_artifacts/03-deps.md`):**
- `internal/modules/iam/domain` (`fillin_handler.go:16`) â€” typed `iamdomain.Capability` consts; cross-refs `iam` T-001
- `internal/modules/iam/application` (`handler.go:17`) â€” `ErrCapabilityDenied` sentinel; cross-refs `iam` T-009
- `internal/modules/iam/authz` (`fillin_handler.go:15`)
- `internal/modules/registry/domain`, `internal/modules/templates/domain`
- `internal/modules/render` (`resolver_readers.go:9`)
- `internal/platform/idempotency` (`approval/infrastructure/postgres_signoff_idemp_store.go:9`)
- `internal/platform/tenant`, `internal/platform/docgenv2`, `internal/platform/httpresponse` (`handler.go:20`), `internal/platform/ratelimit` (`module.go:13`), `internal/platform/servicebus` (`export_service.go:10`)

**DB tables owned (from `_artifacts/04-persistence.md`, full schemas in artifact):**
`documents`, `editor_sessions`, `document_revisions`, `document_checkpoints`, `document_placeholder_values`, `document_exports`, `document_comments`, `approval_routes`, `approval_route_stages`, `approval_instances`, `approval_stage_instances`, `approval_signoffs`, `governance_events`, `metaldocs.pdf_dispatch_outbox`.

**Frontend counterpart:** `frontend/apps/web/src/features/documents/` â€” out of scope for this backend Arc42 but cross-referenced in Â§12.

---

## 4. Solution Strategy

- **Clean-architecture layering** (delivery / application / domain / repository) â€” driver: project convention enforced by api-lint rule "ports live in application".
- **Two-tier authz with a Postgres tripwire** â€” driver: `wiki/decisions/0007-two-tier-authz.md`. Tier-1 role gate at HTTP edge, tier-2 `authz.Require(ctx, tx, cap, area)` in transaction, GUC-driven trigger on `approval_instances` + `approval_signoffs`.
- **Atomic CD+draft via a `CreateDocumentTx` port** â€” driver: `wiki/decisions/0011-cd-atomic-create.md`. Registry owns the call site; documents exposes only the port and the row writer.
- **Placeholder snapshot pinned at submit** â€” driver: `wiki/concepts/placeholders.md`. `application.SnapshotService` populates `placeholder_schema_snapshot`; trigger `enforce_snapshot_on_submit_trg` (`migrations/0152_*.sql:47`) blocks `under_review` transitions with a null snapshot.
- **Sub-package for approval mechanics** â€” driver: cohesion between document state transitions and the approval-instance + signoff lifecycle. Lives at `internal/modules/documents/approval/`; the top-level `internal/modules/approval/` path does **not** exist in workspace (resolved during Phase 3).
- **Codegen bootstrap only** â€” driver: ADR 0012 + spec drift (T-002). `api.gen.go` is generated and committed but routes are still mounted via `mux.HandleFunc`.

---

## 5. Building Block View (C4 Level 2 â€” Container)

### 5.1 Whitebox â€” documents

```mermaid
C4Container
    title Container View â€” documents module
    Container(http, "delivery/http", "Go stdlib mux", "Handler, ExportHandler â€” /api/v1/documents/* + /api/v1/documents/{id}/export/* + comments")
    Container(approvalHttp, "approval/http", "Go", "ApprovalHandler â€” /api/v1/documents/{id}/{submit,signoff,cancel,publish,schedule-publish,supersede,obsolete} + approval admin")
    Container(app, "application", "Go", "Service, SnapshotService, FreezeService, FillinService, ViewService, ExportService, DraftResolverService, ReconstructService, CDDocumentInitializer, SubmitService, DecisionService")
    Container(domain, "domain", "Go", "Document, Revision, Checkpoint, Snapshot, Comment, Export, CompositeHash, ValuesHash, errors, state machine")
    Container(repo, "repository", "Go + pgx", "Repository, FillinRepository, ExportRepository, SnapshotRepository, ResolverReaders, postgres_approval_repository, postgres_signoff_idemp_store")
    Container(jobs, "jobs", "Go", "OrphanPendingSweeper, SessionSweeper")
    Container(api, "api/api.gen.go", "Generated", "oapi-codegen ServerInterface (bootstrap only â€” not mounted)")
    ContainerDb(db, "Postgres", "Postgres", "14 owned tables (see Â§3.2)")
    System_Ext(iam, "iam", "Capabilities")
    System_Ext(registry, "registry", "CD identity")
    System_Ext(templates, "templates", "Template versions")
    System_Ext(render, "render", "DOCXâ†’PDF")
    System_Ext(idemp, "platform/idempotency", "Signoff replay store")
    Rel(http, app, "calls")
    Rel(approvalHttp, app, "calls SubmitService / DecisionService")
    Rel(app, domain, "uses")
    Rel(app, repo, "calls")
    Rel(repo, db, "SQL")
    Rel(jobs, repo, "sweeps")
    Rel(app, iam, "CapabilityChecker port adapter")
    Rel(registry, app, "CreateDocumentTx port")
    Rel(app, templates, "template version reads")
    Rel(app, render, "resolver readers (PDF)")
    Rel(app, idemp, "signoff idempotency")
```

### 5.2 Public surface (selected)

Full enumeration in `wiki/modules/documents/_artifacts/01-surface.md` (517 exported symbols). The rows below are the cross-boundary surface â€” symbols imported by other modules or registered as routes/ports.

| File | Symbol | Kind | Purpose |
|---|---|---|---|
| `internal/modules/documents/module.go:1` | `New`, `RegisterRoutes` | func | Module wiring entry |
| `internal/modules/documents/application/ports.go` | `CapabilityChecker` | iface | Consumer port â€” iam adapter implements (`apps/api/internal/wiring/documents.go:24`) |
| `internal/modules/documents/application/ports.go` | `CreateDocumentTx` | iface | Atomic CD+draft create port consumed by registry (ADR 0011) |
| `internal/modules/documents/application/service.go:26` | `Repository` | iface | Application-layer repo contract |
| `internal/modules/documents/application/service.go:81` | `Audit` | iface | Audit sink consumer port (adapter wired in main.go) |
| `internal/modules/documents/application/service.go:564` | `Service.RenameDocument` | func | Validate + UPDATE + audit (T-005: not transactional) |
| `internal/modules/documents/application/service.go:753` | `Service.UpdateDocumentStatus` | func | Status transitions; finalize tail-call lands here |
| `internal/modules/documents/application/snapshot_service.go` | `SnapshotService` | type | Populates `placeholder_schema_snapshot` (ADR 0008) |
| `internal/modules/documents/application/freeze_service.go` | `FreezeService` | type | `{name}` token substitution at freeze |
| `internal/modules/documents/application/fillin_service.go` | `FillinService` | type | Placeholder-value writes |
| `internal/modules/documents/application/fillin_authz.go:9` | typed `iamdomain.Capability` consts | imports | Cross-ref iam T-001 |
| `internal/modules/documents/application/cd_initializer.go` | `CDDocumentInitializer` | type | Registry-side hook for atomic CD create |
| `internal/modules/documents/approval/application/submit_service.go:43` | `SubmitService.SubmitRevisionForReview` | func | Tier-2 `authz.Require(string(iamdomain.CapDocumentSubmit), areaCode)` at `:85` |
| `internal/modules/documents/approval/application/decision_service.go` | `DecisionService` | type | Signoff approve/reject/publish/supersede/obsolete |
| `internal/modules/documents/delivery/http/handler.go:76` | `NewHandlerWithSubmit` | func | Wires db + submitSvc for atomic finalize |
| `internal/modules/documents/delivery/http/handler.go:145` | `listDocuments` | func | `GET /api/v1/documents` |
| `internal/modules/documents/delivery/http/handler.go:285` | `renameDocument` | func | `PATCH /api/v1/documents/{id}` (T-002 spec gap, T-004 dup route, T-005 audit-tx gap) |
| `internal/modules/documents/delivery/http/handler.go:316` | `finalizeDocument` | func | `POST /api/v1/documents/{id}/finalize` with HTTP idempotency header + replay support |
| `internal/modules/documents/delivery/http/handler.go:869` | `authorizeDocumentScope` | func | Role + ownership gate (tier-1) |
| `internal/modules/documents/delivery/http/handler.go:958..1009` | `mapErr` | func | Legacy envelope mapping (T-001) |
| `internal/modules/documents/repository/repository.go:76` | `CreateDocumentTx` impl | func | Tx-scoped CD+document insert (ADR 0011); asserts `document.create` before INSERT and `document.edit` before pointer/snapshot UPDATEs |
| `internal/modules/documents/repository/repository.go:216` | `UpdateDocumentName` | func | UPDATE inside tx with `authz.Require(CapDocumentEdit)` (T-003 closed Plan 5) |
| `internal/modules/documents/repository/repository.go:343` | `ListDocumentsPaginated` | func | LIMIT/OFFSET; pageSize cap 50 |
| `internal/modules/documents/repository/repository.go:376` | `CountDocuments` | func | Sibling COUNT(*) for list |
| `internal/modules/documents/approval/repository/postgres_approval_repository.go:34` | `InsertInstance` | func | Tripwire-gated INSERT into `approval_instances` |
| `internal/modules/documents/approval/application/idempotency.go:20` | `ComputeIdempotencyKey` | func | Internal deterministic key for submit (NOT HTTP header) |
| `internal/modules/documents/api/api.gen.go:1` | generated stubs (`ListDocuments`, `PostApiDocumentsIdFinalize`, â€¦) | iface | Codegen bootstrap (unmounted) |

The remaining ~480 symbols are domain DTOs, repository internal helpers, and codegen artifacts â€” treated as undocumented-on-purpose at module level (see tech-debt `Coverage stats`).

### 5.3 HTTP operations

Routes registered in `internal/modules/documents/delivery/http/handler.go` and `internal/modules/documents/approval/http/router.go`. **OperationID** column reflects current spec (`api/openapi/v1/openapi.yaml`); `â€”` = no spec op.

| Method | Path | OperationID | Handler | Authz |
|---|---|---|---|---|
| GET | `/api/v1/documents` | `listDocuments` | `Handler.listDocuments` (`handler.go:145`) | role: admin\|filler; filler scoped to `created_by` |
| GET | `/api/v1/documents/stats` | â€” | `Handler.documentStats` (`handler.go:174`) | role |
| GET | `/api/v1/documents/{id}` | `getDocument` | `Handler.getDocument` (`handler.go:114`) | role + ownership |
| PATCH | `/api/v1/documents/{id}` | â€” | `Handler.renameDocument` (`handler.go:285`) | role + ownership; **dup registration at `:86`+`:115`** (T-004) |
| POST | `/api/v1/documents/{id}/finalize` | `finalizeDocument` | `Handler.finalizeDocument` (`handler.go:403`) | role + ownership + tier-2 `authz.Require(string(iamdomain.CapDocumentSubmit), areaCode)` |
| GET | `/api/v1/documents/{id}/revision-history` | `getDocumentRevisionHistory` | `Handler.listRevisionHistory` (`handler.go:694`) | role + ownership |
| POST | `/api/v1/documents/{id}/archive` | â€” | `Handler.archiveDocument` | role |
| POST | `/api/v1/documents/{id}/duplicate` | â€” | `Handler.duplicateDocument` | role |
| GET/POST/PATCH/DELETE | `/api/v1/documents/{id}/comments[/{commentId}]` | â€” | comments CRUD | role + ownership |
| GET/POST | `/api/v1/documents/{id}/sessions` | â€” | session controller | role |
| GET/POST | `/api/v1/documents/{id}/checkpoints` | â€” | checkpoint controller | role |
| GET | `/api/v1/documents/{id}/revisions` | â€” | revisions URL handler | role |
| POST | `/api/v1/documents/{id}/export/pdf` | â€” | `ExportHandler` | role |
| GET | `/api/v1/documents/{id}/export/docx-url` | â€” | `ExportHandler` | role |
| POST | `/api/v1/documents/{id}/submit` | â€” | `ApprovalHandler` (`approval/http/router.go`) | tier-2 `document.submit` |
| POST | `/api/v1/documents/{id}/signoff` | â€” | `ApprovalHandler` | tier-2 `document.signoff` |
| POST | `/api/v1/documents/{id}/cancel` | â€” | `ApprovalHandler` | tier-2 |
| POST | `/api/v1/documents/{id}/publish` | â€” | `ApprovalHandler` | tier-2 `document.publish` |
| POST | `/api/v1/documents/{id}/schedule-publish` | â€” | `ApprovalHandler` | tier-2 |
| POST | `/api/v1/documents/{id}/supersede` | â€” | `ApprovalHandler` | tier-2 |
| POST | `/api/v1/documents/{id}/obsolete` | â€” | `ApprovalHandler` | tier-2 |
| GET | `/api/v1/documents/{id}/approval-instance` | `getApprovalInstanceByDocument` | `ApprovalHandler` | role |
| various | `/api/v1/approval-routes/*` | â€” | `ApprovalHandler` admin | role: admin |

Spec gaps (missing `operationId`s on regulated paths) are enumerated in T-002 and `wiki/backlog/contract-first-followups.md`.

## API Route Truth Table (Plan 8 Baseline)

| Method | Path | Runtime owner (file:line) | Handler method | Spec path | operationId | Codegen method | Status | Notes |
|---|---|---|---|---|---|---|---|---|
| GET | `/api/v1/documents` | `internal/modules/documents/delivery/http/handler.go:83` | `h.listDocuments` | `/api/v1/documents` | `listDocuments` | `ListDocuments` | Aligned | Also re-registered at `handler.go:112` |
| GET | `/api/v1/documents/stats` | `internal/modules/documents/delivery/http/handler.go:84` | `h.documentStats` | `/api/v1/documents/stats` | `documentStats` | `DocumentStats` | Aligned | Also re-registered at `handler.go:113` |
| GET | `/api/v1/documents/{id}` | `internal/modules/documents/delivery/http/handler.go:86` | `h.getDocument` | `/api/v1/documents/{id}` | `getDocument` | `GetDocument` | Aligned | 200 response is now schema-backed by `DocumentDetailResponse`, including embedded JSON `FormDataJSON` |
| PATCH | `/api/v1/documents/{id}` | `internal/modules/documents/delivery/http/handler.go:87` | `h.renameDocument` | â€” | â€” | â€” | Spec missing | Also re-registered at `handler.go:116` |
| POST | `/api/v1/documents/{id}/finalize` | `internal/modules/documents/delivery/http/handler.go:88` | `h.finalizeDocument` | `/api/v1/documents/{id}/finalize` | â€” | `PostApiV2DocumentsIdFinalize` | Aligned | Also re-registered at `handler.go:117` |
| POST | `/api/v1/documents/{id}/archive` | `internal/modules/documents/delivery/http/handler.go:89` | `h.archiveDocument` | `/api/v1/documents/{id}/archive` | â€” | `PostApiV2DocumentsIdArchive` | Aligned | Also re-registered at `handler.go:118` |
| POST | `/api/v1/documents/{id}/duplicate` | `internal/modules/documents/delivery/http/handler.go:90` | `h.duplicateDocument` | â€” | â€” | â€” | Spec missing | Also re-registered at `handler.go:119` |
| POST | `/api/v1/documents/{id}/session/acquire` | `internal/modules/documents/delivery/http/handler.go:92` | `h.acquireSession` | `/api/v1/documents/{id}/session/acquire` | â€” | `PostApiV2DocumentsIdSessionAcquire` | Aligned | Also re-registered at `handler.go:121` |
| POST | `/api/v1/documents/{id}/session/heartbeat` | `internal/modules/documents/delivery/http/handler.go:93` | `h.heartbeatSession` | `/api/v1/documents/{id}/session/heartbeat` | â€” | `PostApiV2DocumentsIdSessionHeartbeat` | Aligned | Also re-registered at `handler.go:122` |
| POST | `/api/v1/documents/{id}/session/release` | `internal/modules/documents/delivery/http/handler.go:94` | `h.releaseSession` | `/api/v1/documents/{id}/session/release` | â€” | `PostApiV2DocumentsIdSessionRelease` | Aligned | Also re-registered at `handler.go:123` |
| POST | `/api/v1/documents/{id}/session/force-release` | `internal/modules/documents/delivery/http/handler.go:95` | `h.forceReleaseSession` | `/api/v1/documents/{id}/session/force-release` | â€” | `PostApiV2DocumentsIdSessionForceRelease` | Aligned | Also re-registered at `handler.go:124` |
| POST | `/api/v1/documents/{id}/autosave/presign` | `internal/modules/documents/delivery/http/handler.go:97` | `h.presignAutosave` | `/api/v1/documents/{id}/autosave/presign` | â€” | `PostApiV2DocumentsIdAutosavePresign` | Aligned | Also re-registered (wrapped) at `handler.go:126` |
| POST | `/api/v1/documents/{id}/autosave/commit` | `internal/modules/documents/delivery/http/handler.go:98` | `h.commitAutosave` | `/api/v1/documents/{id}/autosave/commit` | â€” | `PostApiV2DocumentsIdAutosaveCommit` | Aligned | Also re-registered (wrapped) at `handler.go:130` |
| GET | `/api/v1/documents/{id}/checkpoints` | `internal/modules/documents/delivery/http/handler.go:100` | `h.listCheckpoints` | `/api/v1/documents/{id}/checkpoints` | â€” | `GetApiV2DocumentsIdCheckpoints` | Aligned | Also re-registered at `handler.go:135` |
| POST | `/api/v1/documents/{id}/checkpoints` | `internal/modules/documents/delivery/http/handler.go:101` | `h.createCheckpoint` | `/api/v1/documents/{id}/checkpoints` | â€” | `PostApiV2DocumentsIdCheckpoints` | Aligned | Also re-registered at `handler.go:136` |
| POST | `/api/v1/documents/{id}/checkpoints/{version}/restore` | `internal/modules/documents/delivery/http/handler.go:102` | `h.restoreCheckpoint` | `/api/v1/documents/{id}/checkpoints/{versionNum}/restore` | â€” | `PostApiV2DocumentsIdCheckpointsVersionNumRestore` | Signature mismatch | Runtime param `{version}` differs from spec/codegen `{versionNum}` |
| GET | `/api/v1/documents/{id}/revisions/{rid}/url` | `internal/modules/documents/delivery/http/handler.go:104` | `h.signedRevisionURL` | `/api/v1/documents/{id}/revisions/{rid}/url` | â€” | `GetApiV2DocumentsIdRevisionsRidUrl` | Aligned | Also re-registered at `handler.go:139` |
| GET | `/api/v1/documents/{id}/comments` | `internal/modules/documents/delivery/http/handler.go:105` | `h.listComments` | `/api/v1/documents/{id}/comments` | `listDocumentComments` | `ListDocumentComments` | Aligned | Response is now schema-backed by `DocumentCommentResponse[]` |
| POST | `/api/v1/documents/{id}/comments` | `internal/modules/documents/delivery/http/handler.go:106` | `h.createComment` | `/api/v1/documents/{id}/comments` | `createDocumentComment` | `CreateDocumentComment` | Aligned | Request/response now use named comment schemas |
| PATCH | `/api/v1/documents/{id}/comments/{libraryID}` | `internal/modules/documents/delivery/http/handler.go:107` | `h.updateComment` | `/api/v1/documents/{id}/comments/{libraryID}` | `updateDocumentComment` | `UpdateDocumentComment` | Aligned | Request/response now use named comment schemas |
| DELETE | `/api/v1/documents/{id}/comments/{libraryID}` | `internal/modules/documents/delivery/http/handler.go:108` | `h.deleteComment` | `/api/v1/documents/{id}/comments/{libraryID}` | `deleteDocumentComment` | `DeleteDocumentComment` | Aligned | 204 no-content delete |
| POST | `/api/v1/documents/{id}/export/pdf` | `internal/modules/documents/delivery/http/export_handler.go:40` | `h.exportPDF` | `/api/v1/documents/{id}/export/pdf` | `exportDocumentPDF` | `ExportDocumentPDF` | Aligned | Also registered via rate-limit wrapper at `export_handler.go:45` |
| GET | `/api/v1/documents/{id}/export/docx-url` | `internal/modules/documents/delivery/http/export_handler.go:41` | `h.exportDocxURL` | `/api/v1/documents/{id}/export/docx-url` | `getDocumentDocxURL` | `GetDocumentDocxURL` | Aligned | Also re-registered at `export_handler.go:49` |
| GET | `/api/v1/documents/{id}/fill-in-schema` | `internal/modules/documents/http/fillin_handler.go:36` | `h.GetFillInSchema` | â€” | â€” | â€” | Spec missing |  |
| GET | `/api/v1/documents/{id}/placeholders` | `internal/modules/documents/http/fillin_handler.go:37` | `h.ListPlaceholderValues` | â€” | â€” | â€” | Spec missing |  |
| PUT | `/api/v1/documents/{id}/placeholders/{pid}` | `internal/modules/documents/http/fillin_handler.go:38` | `h.PutPlaceholderValue` | â€” | â€” | â€” | Spec missing |  |
| GET | `/api/v1/documents/{id}/view` | `internal/modules/documents/http/view_handler.go:30` | `h.HandleView` | â€” | â€” | â€” | Spec missing |  |
| POST | `/api/v1/documents/{id}/reconstruct` | `internal/modules/documents/http/reconstruct_handler.go:27` | `h.HandleReconstruct` | â€” | â€” | â€” | Spec missing |  |

Module contract status: Contracted
Owner: leandro

---

## 6. Runtime View

### 6.1 listDocuments (read) â€” `GET /api/v1/documents`

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as Handler
    participant S as Service
    participant R as Repository
    participant DB as Postgres
    C->>H: GET /api/v1/documents?page&pageSize&status&...
    H->>H: hasAnyRole(admin|filler) â€” 403 else
    H->>H: parseListOptions (cap pageSize=50)
    H->>H: effectiveUserID/CreatedBy scoping (non-admin)
    H->>S: ListDocumentsPaginated(opts)
    S->>R: ListDocumentsPaginated(opts)
    R->>DB: SELECT â€¦ FROM documents WHERE tenant_id=$1 [AND created_by=$N] ORDER BY updated_at DESC LIMIT/OFFSET
    DB-->>R: rows
    R-->>S: items
    S->>R: CountDocuments(opts)
    R->>DB: SELECT COUNT(*) FROM documents WHERE â€¦
    DB-->>S: total
    S-->>H: items + total
    H-->>C: 200 {items, page, pageSize, total}
```

Source: `_artifacts/02-flow-listDocuments.md`. No transaction; read-only.

### 6.2 renameDocument (write) â€” `PATCH /api/v1/documents/{id}`

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as Handler
    participant S as Service
    participant R as Repository
    participant A as Audit sink
    participant DB as Postgres
    C->>H: PATCH /api/v1/documents/{id} {name}
    H->>H: authorizeDocumentScope (admin OR filler+owner)
    H->>H: json.Decode req.Name (400 on fail)
    H->>S: RenameDocument(tenantID, userID, docID, name)
    S->>S: validate name (â‰¤255 chars, non-empty)
    S->>R: GetDocument
    S->>S: guard doc.Status==Draft else 409
    S->>R: UpdateDocumentName  (ExecContext â€” no tx)
    R->>DB: UPDATE documents SET name=$2, updated_at=now() WHERE id=$1 AND tenant_id=$3
    S->>A: audit.Write("document.renamed", ...)  (OUTSIDE tx â€” T-005)
    H-->>C: 204 No Content
```

Source: `_artifacts/02-flow-renameDocument.md` (corrected against `handler.go:285..308` during Phase 6.75 â€” response is `204 No Content`, not a re-fetched JSON body). Spec drift (T-002) â€” route absent from openapi.yaml. Duplicate registration (T-004). No tier-2 authz, no tripwire on `documents` table (T-003). Latent: error path at `handler.go:303-304` calls `httpErr(w, status, msg)` **twice** for a single `RenameDocument` error â€” second call writes a header after a status is already written; visible as a "superfluous WriteHeader" log only.

### 6.3 finalizeDocument (state transition) â€” `POST /api/v1/documents/{id}/finalize`

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as Handler
    participant SS as SubmitService
    participant AR as ApprovalRepo
    participant EM as EventEmitter
    participant S as Service
    participant DB as Postgres
    C->>H: POST /api/v1/documents/{id}/finalize
    H->>H: hasAnyRole(admin|filler) â€” 403 else
    H->>H: load draft guard (status==draft else 409)
    H->>H: resolve approval route + content hash (409 if missing)
    H->>SS: SubmitRevisionForReview(...)
    SS->>DB: BeginTx
    SS->>DB: setAuthzGUC (actor, tenant)
    SS->>SS: loadDocumentAreaCode
    SS->>DB: authz.Require(ctx, tx, string(iamdomain.CapDocumentSubmit), areaCode)
    SS->>AR: InsertInstance â†’ INSERT approval_instances (tripwire-gated)
    SS->>AR: InsertStageInstances â†’ INSERT approval_stage_instances (eligible_actor_ids)
    SS->>DB: UPDATE documents SET status='under_review' WHERE status='draft'  (fires enforce_snapshot_on_submit_trg)
    SS->>EM: Emit â†’ INSERT governance_events
    SS->>DB: Commit
    SS-->>H: {InstanceID}
    H->>S: Finalize â†’ UpdateDocumentStatus (audit-only tail call)
    H-->>C: 201 {"instanceId": ...}
```

Source: `_artifacts/02-flow-finalizeDocument.md`. Full tripwire defense-in-depth on approval tables. Runtime now enforces HTTP `Idempotency-Key`, stores replay responses through `internal/platform/idempotency`, and replays as `201 { instanceId }` with `Idempotent-Replay: true`.

### State transitions

| From | To | Trigger | Authz cap (tier-2) | Surface |
|---|---|---|---|---|
| draft | under_review | `POST .../finalize` (handler) â†’ `SubmitService.SubmitRevisionForReview` | `document.submit` (`submit_service.go:85`) | `approval_instances` INSERT + `documents.status` UPDATE in one tx |
| under_review | approved | `POST .../signoff` final-stage | `document.signoff` | `approval_signoffs` INSERT |
| approved | published | `POST .../publish` | `document.publish` | `documents.status='published'` + governance event |
| published | superseded | `POST .../supersede` | `document.supersede` | new revision created; old marked superseded |
| published | obsolete | `POST .../obsolete` | `document.obsolete` | `documents.status='obsolete'` |
| under_review | draft (rejected) | `POST .../signoff` reject | `document.signoff` | `approval_instances.status='rejected'`; `documents.status` rollback |

### Failure modes (current legacy envelope, T-001)

| Condition | HTTP | Legacy `error.code` | Source |
|---|---|---|---|
| Missing role / ownership | 403 | `forbidden` | `handler.go:871, :887` |
| Validation (name length) | 400 | `invalid_name` | `service.go:565-567` |
| JSON decode | 400 | `invalid_body` | `handler.go:296` |
| Not found | 404 | `not_found` | `mapErr` `:966-972` |
| State guard (not draft) | 409 | `invalid_state_transition` | `service.go:572`, `handler.go:351` |
| No active approval route | 409 | route msg | `handler.go:387` |
| Capability denied (in-tx) | 403 | mapped from `iamapp.ErrCapabilityDenied` | `handler.go:17` import |
| Server error | 500 | `internal_error` | `mapErr` `:1009` |

Target shape: RFC 9457 `application/problem+json` (T-001 backlog R-001).

---

## 7. Deployment View

- Binary: `apps/api/cmd/metaldocs-api` (single Go server); worker side mounted via `apps/worker/cmd/metaldocs-worker`.
- Process: container on port `:8081` (per `scripts/start-api.ps1`).
- Migrations: applied at startup; documents-owned files live in `migrations/` chronological set (0001..0183 to date) â€” full enumeration in `_artifacts/04-persistence.md`. Forward-only (IP-006).
- Environment: documents reads no env/config vars directly (verified in `_artifacts/03-deps.md`). All knobs come through DI from `apps/api/cmd/metaldocs-api/main.go`.
- Background jobs:
  - `internal/jobs/effective_date_publisher/job.go:8` â€” promotes approved â†’ published at effective date
  - `internal/jobs/stuck_instance_watchdog/job.go:10` â€” watchdog for stalled approval instances
  - `internal/modules/documents/jobs/orphan_pending_sweeper.go`, `session_sweeper.go` â€” orphan + session cleanup

---

## 8. Cross-cutting Concepts

### 8.1 Authentication & Authorization
- Tier-1 (HTTP edge): role gate via `hasAnyRole(roleAdmin, roleDocumentFiller)` (`handler.go:870`) + ownership (`IsDocumentOwner` `:880`). Role strings `system_admin` / `document_filler` at `handler.go:26, :28`.
- Tier-2 (in-tx): `authz.Require(ctx, tx, string(iamdomain.CapDocumentSubmit), areaCode)` (`submit_service.go:85`). Capability value `"document.submit"` (Plan 4: previously `"doc.submit"`, closed T-008 / iam T-001).
- Typed capabilities: `internal/modules/documents/application/fillin_authz.go:9` consumes `iamdomain.Capability` consts. Module now uses typed namespace exclusively (T-008 closed).
- Capability adapter: `internal/modules/documents/application/ports.go` declares `CapabilityChecker`; impl `capabilityServiceAdapter` at `apps/api/internal/wiring/documents.go:14`; `NewCapabilityChecker` factory at `:24` (ADR 0007 J2 amendment).
- Postgres tripwire: `enforce_capability_asserted` function (`migrations/0142b_role_capabilities_v2_enforce.sql:67`), triggers on `approval_instances` (`:201`) and `approval_signoffs` (`:207`). Plan 5 migration `0188_tripwire_extend.sql:196-199` additionally attaches `trg_require_cap_asserted` to `public.documents` (INSERT `CapDocumentCreate`; UPDATE `CapDocumentEdit`). Reads `metaldocs.asserted_caps` GUC set by `setAuthzGUC` (`approval/application/authz_guc.go:11`). **T-003 closed.**
- Sentinel: `iamapp.ErrCapabilityDenied` imported at `handler.go:17`; `authz.ErrCapDenied` (struct) also imported (iam T-009 closed by Plan 4 â€” renamed from `authz.ErrCapabilityDenied`).

### 8.2 Error envelope
- Current: legacy `{error:{code,message,details,trace_id}}` via `httpErr` + `mapErr` (`handler.go:958..1013`).
- Target: RFC 9457 Problem+JSON via `internal/platform/httpresponse.WriteProblem`.
- Status: mid-migration (T-001).

### 8.3 Idempotency
- `internal/platform/idempotency` provides Stripe-style header replay store.
- Used by signoff via `approval/infrastructure/postgres_signoff_idemp_store.go:9`.
- Used on finalize: handler requires `Idempotency-Key`, hashes the request payload, checks/records replay entries through `internal/platform/idempotency`, and returns `Idempotent-Replay: true` on replay. The submit path still computes its internal deterministic key (`approval/application/idempotency.go:20`) for approval-side consistency.

### 8.4 Pagination
- Offset only: `page` + `pageSize` (default 1 / 20, cap 50). Repo `LIMIT/OFFSET` at `repository.go:343`.
- Response body: `{items, page, pageSize, total}` at `handler.go:167`; generated counterpart `DocumentListResponse` at `api.gen.go:266`.
- Cursor pagination not adopted (IP-003 not-applicable today).

### 8.5 Logging & Observability
- No structured trace correlation wired in this module (IP-007 system-wide gap â€” not documents-specific debt).
- Governance events emitted via `EventEmitter.Emit` â†’ INSERT `governance_events` (`approval/application/events.go:35`); QMS audit trail.
- Audit sink via consumer port (`Audit` iface at `application/service.go:81`); adapter wired in main.go (T-007 latent).

### 8.6 Concurrency / Transactions
- Repository methods accept `context.Context` and accept tx where the SubmitService composes one (`approval/application/submit_service.go:68`).
- `Service.RenameDocument` does **not** wrap UPDATE + audit in a tx (T-005).
- Atomic CD+draft create uses an explicit tx via `CreateDocumentTx` port (ADR 0011). Because migration 0188 tripwires `documents` INSERT as `document.create` and UPDATE as `document.edit`, `CreateDocumentTx` asserts both capabilities in-order inside the caller-owned tx.
- One-active-instance constraint enforced by partial unique index `ux_approval_instances_active` on `approval_instances(document_id) WHERE status='in_progress'` (`migrations/0135_*.sql:33`, renamed by 0194).

### 8.7 Snapshot & freeze (placeholder lifecycle)
- `SnapshotService` populates `placeholder_schema_snapshot` (ADR 0008; `wiki/concepts/placeholders.md`).
- Trigger `enforce_snapshot_on_submit_trg` (`migrations/0152_*.sql:47`) blocks `documents.status='under_review'` UPDATE when snapshot is null. Fires in finalize tx (step Â§6.3 box `UPDATE documents`).
- `FreezeService` performs `{name}` token substitution at freeze (`wiki/concepts/token-syntax.md`).
- `document_placeholder_values` schema bug surfaced: `revision_id REFERENCES documents(id)` (T-009).

---

## 9. Architecture Decisions

| Decision | Link / Status |
|---|---|
| Eigenpal DOCX editor adopted | `wiki/decisions/0001-eigenpal-adoption.md` |
| Two-tier authz (HTTP + in-tx + tripwire) | `wiki/decisions/0007-two-tier-authz.md` |
| Fixed 7-token placeholder catalog + snapshot | `wiki/concepts/placeholders.md` (ADR 0008) |
| Atomic CD+draft create via `CreateDocumentTx` port | `wiki/decisions/0011-cd-atomic-create.md` |
| Contract-first OpenAPI via oapi-codegen | `wiki/decisions/0012-contract-first-api.md` (bootstrap only on this module) |
| `{name}` single-brace token syntax | `wiki/concepts/token-syntax.md` (ADR 0003) |
| Duplicate route registration (`handler.go:86/:115`) | `tech-debt: missing-ADR` (T-004) |
| `audit.Write` outside SQL UPDATE tx in rename | `tech-debt: missing-ADR` (T-005) |
| Audit emission via consumer-port interface (no audit/domain import) | `tech-debt: missing-ADR` (T-007) |
| Capability namespace unified to typed `iamdomain.Capability` (`document.*`); `doc.submit` literal replaced with `string(iamdomain.CapDocumentSubmit)` | Plan 4 (2026-05-11) â€” closed T-008 |
| `document_placeholder_values.revision_id` FK target | `tech-debt: missing-ADR` (T-009) |

---

## 10. Quality Requirements

| Goal | Scenario | Pass criteria |
|---|---|---|
| Authz isolation (approval) | A user without `document.submit` calls `POST /finalize` | 403; `approval_instances` unchanged; tripwire would fire even if tier-2 bypassed |
| Authz isolation (documents table) | A user with stale role token calls `PATCH /documents/{id}` | 403 from role gate; in-tx `authz.Require` + tripwire trigger on `documents` table (T-003 closed Plan 5) |
| Atomic CD create | Crash mid-registry insert | Whole tx rolls back; `cd_sequence_counters` unchanged (ADR 0011) |
| Audit completeness on rename | Crash between UPDATE and audit.Write | **T-005: row mutated, no audit row** â€” fails today |
| Replay safety on finalize | Client retries finalize with the same `Idempotency-Key` and unchanged payload | Second call replays `201 { instanceId }` with `Idempotent-Replay: true`; mismatched payload under the same key is rejected by the shared idempotency store |
| Snapshot guard on submit | Submit with null placeholder snapshot | DB trigger `enforce_snapshot_on_submit_trg` raises; 500 surfaces as mapped error |
| Multi-tenant isolation | Cross-tenant id guessed | Every owned table carries `tenant_id`; repo queries scope (`repository.go:343, :376`) |

---

## 11. Risks & Technical Debt

Pointer-only. Body in `wiki/modules/documents-tech-debt.md`. Severity rubric in template `tech-debt-register.md`.

Summary counts (T-008 closed Plan 4; all rows counted by tally including closed):
- Critical: 1
- Major: 7
- Minor: 4

Top 3 (by severity, then blast radius):
1. OpenAPI spec drift on `/api/v1/documents/*` â€” see tech-debt T-002 (Critical; blocks RFC 9457 migration; multiple handlers without spec ops)
2. `renameDocument` audit write outside SQL tx â€” see tech-debt T-005 (Major; audit-trail atomicity broken) â€” replaces T-003 in top-3 (T-003 closed Plan 5)
3. `renameDocument` audit write outside SQL tx â€” see tech-debt T-005 (Major; audit-trail atomicity broken)

---

## 12. Glossary

| Term | Definition |
|---|---|
| Document | Governed revision row in `documents`, filled from a template version, bound to a controlled-document code |
| Governed revision | Business revision lineage stored in `documents`, grouped by `controlled_document_id`, with zero-based `revision_number` (`0` = `REV00`) and frozen `revision_title` captured at finalize |
| Technical revision | Autosave / artifact lineage in `document_revisions`; never the source of governed sidebar history |
| Checkpoint | Editor autosave point in `document_checkpoints` |
| Snapshot (placeholder schema) | Pinned placeholder catalog stored in `placeholder_schema_snapshot`; required for `under_review` |
| Approval instance | Row in `approval_instances`; one active per document via `ux_approval_instances_active` |
| Stage instance | Row in `approval_stage_instances`; materialised approval route stages with `eligible_actor_ids` |
| Signoff | Row in `approval_signoffs`; per-stage approve/reject |
| Tripwire | Postgres trigger `enforce_capability_asserted` reading `metaldocs.asserted_caps` GUC |
| `setAuthzGUC` | Helper at `approval/application/authz_guc.go:11` that primes the GUC for tripwire |
| `document.submit` | Tier-2 capability string asserted at `submit_service.go:85`; renamed from `doc.submit` in Plan 4 (migration 0186) |

---

## Cross-links

- Related ADRs: `wiki/decisions/0001-eigenpal-adoption.md`, `wiki/decisions/0007-two-tier-authz.md`, `wiki/decisions/0011-cd-atomic-create.md`, `wiki/decisions/0012-contract-first-api.md`
- Related concepts: `wiki/concepts/placeholders.md`, `wiki/concepts/token-syntax.md`
- Upstream template publisher: [`wiki/modules/templates.md`](templates.md) â€” publishes the `template_version` rows (with `placeholder_schema`) that documents instantiates from; `documents` snapshots `placeholder_schema` at create time (Â§8.7 of that doc)
- Frontend counterpart: `frontend/apps/web/src/features/documents/` â€” Library, Wizard, Editor, Published view (see `wiki/architecture/frontend-structure.md`)
- Predecessor stub: `retired documents module stub` â€” DEPRECATED, retired by R-100
- Backlog: `wiki/backlog/documents-refactor.md`, `wiki/backlog/contract-first-followups.md`
- Tech debt: `wiki/modules/documents-tech-debt.md`
- Iam cross-refs: `wiki/modules/iam-tech-debt.md` T-001 (namespaces), T-006 (RFC 9457), T-009 (`ErrCapabilityDenied`)
- Auth cross-ref: [`wiki/modules/auth.md Â§8.8`](auth.md) â€” `authdomain.CurrentUserFromContext` is the IN-edge this module reads after middleware injection; Â§8.1 of auth.md covers the middleware that sets the context value
- See also: [`modules/audit.md`](audit.md) â€” documents emits audit events via `documentsAuditAdapter` (`main.go:477-492`); T-005 (rename outside tx) and T-007 (latent consumer port) are the open gaps in the consumer-side register
- See also: [`modules/registry.md`](registry.md) â€” registry owns the CD identity (`controlled_documents`); documents exposes the `CreateDocumentTx` port that registry calls inside the atomic create transaction (ADR 0011)
- See also: [`modules/taxonomy.md`](taxonomy.md) â€” documents reads `process_areas.name` live at document-create time for the `area_name_snapshot` column (`repository.go:94-101`); taxonomy Â§8.9 documents this cross-module data contract

- Research artifacts: `wiki/modules/documents/_artifacts/00-context.md` â€¦ `06-selfreview.md`

## Changelog (this doc)

- 2026-05-19 - Editor sidebar identification layout sync: the editor sidebar now labels document identity as `Identificacao`, renders identity fields as stacked editorial labels (`Codigo`, `Tipo`, `Area responsavel`, `Visibilidade`), removes the extra outer sidebar padding, and intentionally does not show page count/file size until DB/API/TanStack runtime contract work exposes those values truthfully.
- 2026-05-19 - Editor sidebar revision-title/density sync: `REV00` now uses the canonical initial governed title `Criacao do documento` when `revisionTitle` is omitted; later governed revisions still require `revisionTitle` at formal submission. The editor sidebar renders governed revision rows as code/title/date without inline workflow status, keeps draft approvers hidden, and collapses long governed histories without using technical `document_revisions`.
- 2026-05-18 - Governed sidebar sync: `documents.revision_title` is now part of the runtime model and required on `POST /api/v1/documents/{id}/finalize`; the editor sidebar reads governed history from `GET /api/v1/documents/{id}/revision-history`, and that history is sourced from `documents` lineage by `controlled_document_id`, not from technical `document_revisions`.
- 2026-05-18 - Approval/registry sidebar boundary sync: the editor consumes `GET /api/v1/documents/{id}/approval-instance` only for `under_review`, and visibility is resolved from the registry-controlled document contract instead of a documents-local duplicate field.
- 2026-05-18 - Approval/review comments hardening sync: final approval now stops server-side with `approval.unresolved_comments` when active document comments remain unresolved; the signoff dialog maps that conflict inline and the editor keeps comment-load failures visible with a persistent retry banner instead of toast-only feedback.
- 2026-05-18 - Comments contract sync: `/api/v1/documents/{id}/comments*` now has named OpenAPI request/response schemas, generated backend/frontend types, and the editor comments wrapper consumes those generated types instead of handwritten payload rows.
- 2026-05-17 - Review comments lifecycle audit sync: recorded backend/API prerequisite that unresolved active comments must block approval/release server-side before frontend can claim enforcement.
- 2026-05-15 - Novo Documento runtime repair: `CreateDocumentTx` now asserts `document.edit` before the initial pointer/snapshot `documents` UPDATEs so the Plan 5 tripwire accepts atomic blank-template creation.

- 2026-05-10 â€” Full Arc42 + C4 rewrite via `metaldocs-module-doc` skill (Phases 0â€“8). Supersedes prior FE-leaning doc; FE Key files now live under `wiki/architecture/frontend-structure.md` cross-link.

