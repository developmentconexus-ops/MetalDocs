# Stage-1 Audit Artifact — module-documents-core

> **Area:** `internal/modules/documents` (excluding `approval/` subtree — a sibling mapper owns it)
> **Produced:** 2026-06-10 — read-only pass, branch `qa/iam-area-membership`
> **Method:** Exhaustive file read + grep-verified inbound/outbound imports. No runtime execution (Docker down); runtime-unverified items tagged explicitly.
> **Framing refs:** `wiki/architecture/backend-blueprint.md`, `wiki/standards/backend-canon.md`, `wiki/architecture/backend-target-architecture.md`.

---

## 1. Identity & Purpose

`documents` is the controlled-document-instance lifecycle module. It owns the `documents` table and its surrounding technical tables (revisions, sessions, checkpoints, placeholder values, comments, exports) and enforces the state machine `draft → under_review → approved → published → superseded | obsolete`. A document is an instance of a template version bound to a `controlled_documents` entry; it carries a business revision number (`revision_number`, zero-based → `REV00`, `REV01`, …), a placeholder snapshot, and an autosave artifact chain stored on S3. The module exposes the `CreateDocumentTx` port consumed by the `controlled-documents` module for atomic CD+document creation (ADR 0011). It also contains the editor session/autosave loop, the placeholder fill-in/freeze/materialize pipeline, the PDF export cache layer, and background sweeper jobs. The `approval/` sub-tree (same filesystem root, separate mapper) owns all approval-state-machine logic; this artifact covers only the core subtree.

---

## 2. File Inventory

### `internal/modules/documents/` (root)

| File | Role |
|---|---|
| `module.go` | Module struct, `New()` factory, `RegisterRoutes`, `RegisterRoutesWithRateLimit`, `buildLegacyMux`, `placeholderOptionsIAMAdapter` — composition root |
| `module_wrapper_test.go` | Wrapper-level integration test |

### `internal/modules/documents/domain/`

| File | Role |
|---|---|
| `model.go` | Core types: `Document`, `Session`, `Revision`, `Checkpoint`, `PendingUpload`, `RevisionHistoryItem`; `DocumentStatus` typed const (3 values: `draft`/`under_review`/`archived`); `SessionStatus`; all domain error sentinels |
| `state.go` | `CanTransitionDocument` pure function — covers only pre-approval transitions (`draft→under_review`, `→archived`) |
| `snapshot.go` | `TemplateSnapshot`, `SnapshotHashes`; `Hashes()` computing sha256 over each snapshot field |
| `values_hash.go` | `ComputeValuesHash(map[string]any) (string, error)` — sorted-key sha256 over placeholder values; returns error on marshal failure |
| `composite_hash.go` | `ComputeCompositeHash` — sha256 for PDF export cache key over `(contentHash, templateVersionID, grammarVer, rendererVer, RenderOptions)` |
| `comment.go` | `Comment`, `CommentCreateInput`, `CommentUpdateInput` types |
| `export.go` | `Export`, `ExportResult` types; `NewExport` constructor; export error sentinels; `allowedPaperSizes` set |
| `errors.go` | Two additional domain errors: `ErrValidationFailed`, `ErrEffectiveDateMissing` (complement `model.go` sentinels) |
| `snapshot_test.go` | Unit tests |
| `values_hash_test.go` | Unit tests |
| `composite_hash_test.go` | Unit tests |
| `model_test.go` | Unit tests |

### `internal/modules/documents/application/`

| File | Role |
|---|---|
| `service.go` | `Service` struct and all primary use-case methods: `CreateDocument` (legacy/duplicate only), `cloneIntoTx` (atomic CD create entry), `GetDocument`, `RenameDocument`, `DuplicateDocument`, `ListDocumentsPaginated`, `DocumentStats`, session ops, presign/commit autosave, `SyncArtifactMetadata`, checkpoint/restore, `Archive`, `SignedRevisionURL`, comments CRUD; all repository and collaborator interfaces (`Repository`, `Presigner`, `TemplateReader`, `FormValidator`, `Audit`, `ControlledDocumentReader`, `ControlledDocumentDuplicator`, `ProfileDefaultTemplateReader`); three constructors (`New`, `NewService`, `NewServiceWithSnapshot`) |
| `ports.go` | `CapabilityChecker` interface — consumer port for tier-1 authz (`CanDo`, `IsSystemAdmin`) |
| `snapshot_service.go` | `SnapshotService` — `ResolveTemplate` (loads snapshot + placeholder list pre-INSERT), `SnapshotFromTemplate` (deprecated post-commit path retained for backfill) |
| `freeze_service.go` | `FreezeService` — `Pin` (in-tx: validate + resolve computed placeholders + write freeze marker + enqueue outbox), `Materialize` (async: fanout call to docx-renderer), `Freeze` (legacy synchronous path) |
| `fillin_service.go` | `FillInService` — `SetPlaceholderValue`, `GetPlaceholderValues`, `GetFillInSchema`; `SnapshotSchemaReader` (reads `placeholder_schema_snapshot`); dual-format schema parser (`[...]` vs `{"placeholders":[...]}`) |
| `fillin_authz.go` | `requireDocEditDraft` — opens read-committed tx, seeds GUCs, resolves area via `LoadDocumentAreaCode`, calls `authz.Require(CapDocumentEdit, areaCode)` |
| `cd_initializer.go` | `CDDocumentInitializer` — adapter exposing `CloneTemplate`, `ResolveTemplateStorageKey`, `Exists` to the controlled-documents atomic create port |
| `context_builder.go` | `DocumentContextBuilder` — builds `resolvers.ResolveInput` for computed placeholder resolution (reads area, CD ID, area name, active approval instance from DB) |
| `document_area.go` | `LoadDocumentAreaCode` — canonical area resolver (documents + controlled_documents JOIN); single source of truth for area lookup used by authz guards across the module |
| `export_service.go` | `ExportService` — `ExportPDF` (composite-hash cache, `ConvertPDF` call, insert export row), `SignedDocxURL`, `GetDocumentSummary` |
| `view_service.go` | `ViewService` — `GetViewURL` (read-only tx + `authz.Require(document.view)`, checks viewable statuses, reads `final_pdf_s3_key`) |
| `reconstruct_service.go` | `ReconstructionService` — `GetReconstruction` (read-only tx + `authz.Require(document.edit)`, delegates to `ReconstructionRunner`) |
| `list_options.go` | `ListOptions = repository.ListOptions` type alias |
| `iam_user_options.go` | `IAMUserOptionsReader` consumer port; `UserOption` view type |
| `service_test.go`, `service_pagination_test.go`, `service_caps_test.go`, `service_cd_test.go` | Unit tests |
| `snapshot_service_test.go`, `snapshot_resolver_test.go`, `snapshot_seeder_test.go`, `snapshot_wire_test.go` | Unit tests |
| `freeze_service_test.go`, `freeze_pin_test.go`, `freeze_idempotency_test.go` | Unit tests |
| `fillin_service_test.go`, `export_service_test.go` | Unit tests |
| `context_builder_test.go`, `cd_initializer_test.go`, `autosave_commit_branches_test.go` | Unit tests |
| `create_document_snapshot_integration_test.go` | Integration test: snapshot atomicity |

### `internal/modules/documents/repository/`

| File | Role |
|---|---|
| `repository.go` | `Repository` struct + all core SQL: `CreateDocument` (legacy non-tx), `CreateDocumentTx` (atomic: authz GUCs, advisory lock, snapshot write, placeholder seeding), `GetDocument` (LEFT JOIN revisions for artifact metadata), `UpdateDocumentName`/`UpdateDocumentNameTx`, `ListDocuments`, `ListDocumentsForUser`, `ListDocumentsPaginated` (keyset cursor `(updated_at DESC, id DESC)`), `CountDocuments`, `StatsByStatus`, `StatsByArea`, `UpdateDocumentStatus`, `MarkArchived`, `Unarchive`, all session ops, presign/commit/sync, checkpoint/restore, revision history, comments CRUD; helpers `loadDocumentArea`, `loadDocumentAreaBySession`, `buildDocumentFilter` |
| `snapshot_repository.go` | `SnapshotRepository` — `WriteSnapshot`, `ReadSnapshot`, `ReadSnapshotWithFreezeAt`, `WriteFreeze`, `WriteFinalDocx`; `requireRowsAffected` guard; `DBTX` interface |
| `fillin_repository.go` | `FillInRepository` — `SeedDefaults`, `UpsertValue`, `ListValues`; `PlaceholderValue` struct; schema-qualified table support for tests |
| `export_repository.go` | `InsertExport` (ON CONFLICT DO NOTHING + re-read on conflict), `GetExportByHash` |
| `resolver_readers.go` | `RevisionReader` — implements `resolvers.RevisionReader`: `GetRevisionNumber`, `GetEffectiveFrom`, `GetAuthor`, `GetDocumentTitle` — all querying `documents` table |
| Integration and unit test files (10 files) | Coverage for create, list, stats, autosave, commit, revision history, tripwire, tenant isolation, fillin, snapshot |

### `internal/modules/documents/delivery/http/`

| File | Role |
|---|---|
| `handler.go` | `Handler` struct; all main HTTP handlers: `listDocuments`, `documentStats`, `getDocument`, `renameDocument`, `finalizeDocument` (with idempotency store), `archiveDocument`, `duplicateDocument`, all session ops, presign/commit autosave, checkpoint ops, `listRevisionHistory`, `signedRevisionURL`, comments CRUD; `authorizeDocumentScope`; `withAdminCtx`; `mapErr`, `httpErr`, `httpErrDetail`; `RegisterRoutes`, `RegisterRoutesWithRateLimit`; `documentDetailResponse` DTO + `toDocumentDetailResponse` |
| `export_handler.go` | `ExportHandler` — `exportPDF`, `exportDocxURL`; `mapExportErr`; paper size allow-list |
| `generated_adapter.go` | `GeneratedServerAdapter` — bridges every codegen `ServerInterface` method to legacy mux via `forward()` — all typed params discarded |
| Test files (7 files) | Handler, comments, pagination, problem, export, finalize wiring, adapter tests |

### `internal/modules/documents/http/` (legacy sub-package, `package documentshttp`)

| File | Role |
|---|---|
| `fillin_handler.go` | `FillInHandler` — `GetFillInSchema`, `ListPlaceholderValues`, `PutPlaceholderValue`; separate RFC 9457 error writers |
| `placeholder_options_handler.go` | `PlaceholderOptionsHandler` — `HandleGetOptions` — enum values for `user`-type placeholders |
| `view_handler.go` | `ViewHandler` — `HandleView` — PDF status + presigned URL |
| `reconstruct_handler.go` | `ReconstructHandler` — `HandleReconstruct` — DOCX reconstruction entry for editor |
| `pdf_webhook_handler.go` | `PDFWebhookHandler` — `HandlePDFComplete` — HMAC-SHA256 authenticated inbound webhook from docgen worker; resolves canonical tenant by document ID |
| Test files (5 files) | fillin, placeholder options, view, reconstruct, pdf webhook, rbac tests |

### `internal/modules/documents/api/`

| File | Role |
|---|---|
| `gen.go` | `//go:generate` directive for oapi-codegen |
| `api.gen.go` | Generated: `ServerInterface`, `HandlerWithOptions`, `StdHTTPServerOptions`, all request/response types, `DocumentSummaryStatus` enum (8 values), `ListDocumentsParams`, `FinalizeDocumentParams`, etc. |

### `internal/modules/documents/jobs/`

| File | Role |
|---|---|
| `orphan_pending_sweeper.go` | `StartOrphanPendingSweeper` — background goroutine ticker, `DeleteExpiredPending` with `authz.WithBackgroundBypass` |
| `session_sweeper.go` | `StartSessionSweeper` — background goroutine ticker, `ExpireStaleSessions` with `authz.WithBackgroundBypass` |
| `jobs_test.go` | Integration placeholder (skipped) |

---

## 3. Public Surface

### Cross-boundary Go symbols (consumed outside the module)

| Symbol | File:line | Kind | Consumer(s) |
|---|---|---|---|
| `New(Dependencies) *Module` | `module.go:54` | func | `apps/api/cmd/metaldocs-api/main.go` |
| `(*Module).RegisterRoutes` | `module.go:118` | method | main.go |
| `(*Module).RegisterRoutesWithRateLimit` | `module.go:132` | method | main.go |
| `(*Module).Repo() *repository.Repository` | `module.go:146` | method | main.go (worker wiring) |
| `application.CapabilityChecker` | `application/ports.go:12` | interface | `apps/api/internal/wiring/documents.go:28` |
| `application.NewCDDocumentInitializer(*Service)` | `application/cd_initializer.go:19` | func | main.go (wires to controlled-documents) |
| `application.LoadDocumentAreaCode` | `application/document_area.go:33` | func | approval subtree services |
| `application.NewFreezeService(...)` | `application/freeze_service.go:70` | func | main.go |
| `repository.New(*sql.DB)` | `repository/repository.go:36` | func | `module.go:58` |
| `repository.NewFillInRepository` | `repository/fillin_repository.go:31` | func | `module.go:84` |
| `repository.NewSnapshotRepository` | `repository/snapshot_repository.go:26` | func | platform/docgenv2 wiring |
| `repository.NewRevisionReader` | `repository/resolver_readers.go:16` | func | main.go (context builder wiring) |
| `repository.DBTX` | `repository/snapshot_repository.go:19` | interface | freeze_service.go, fillin_repository.go |
| `domain.Document` | `domain/model.go:25` | type | handler DTOs, approval subtree, controlled-documents |
| `domain.ComputeValuesHash` | `domain/values_hash.go:11` | func | freeze_service.go |
| `domain.ComputeCompositeHash` | `domain/composite_hash.go:30` | func | export_service.go |
| `domain.ErrNotFound` (and 11 other sentinels) | `domain/model.go:112–132` | vars | handler mapErr, approval subtree |
| `jobs.StartOrphanPendingSweeper` | `jobs/orphan_pending_sweeper.go:12` | func | main.go |
| `jobs.StartSessionSweeper` | `jobs/session_sweeper.go:12` | func | main.go |

### HTTP routes (core subtree; approval subtree excluded)

| Method | Path | Handler file:line | Authz |
|---|---|---|---|
| GET | `/api/v1/documents` | `delivery/http/handler.go:178` | `isSystemAdmin` or `created_by` scope |
| GET | `/api/v1/documents/stats` | `delivery/http/handler.go:224` | same |
| GET | `/api/v1/documents/{id}` | `delivery/http/handler.go:316` | `authorizeDocumentScope` (admin OR owner) |
| PATCH | `/api/v1/documents/{id}` | `delivery/http/handler.go:400` | `authorizeDocumentScope` |
| POST | `/api/v1/documents/{id}/finalize` | `delivery/http/handler.go:435` | `authorizeDocumentScope` + tier-2 `document.submit` in SubmitService |
| POST | `/api/v1/documents/{id}/archive` | `delivery/http/handler.go:605` | `authorizeDocumentScope` |
| POST | `/api/v1/documents/{id}/duplicate` | `delivery/http/handler.go:621` | `authorizeDocumentScope` |
| POST | `/api/v1/documents/{id}/session/acquire` | `delivery/http/handler.go:647` | `authorizeDocumentScope` |
| POST | `/api/v1/documents/{id}/session/heartbeat` | `delivery/http/handler.go:678` | `authorizeDocumentScope` |
| POST | `/api/v1/documents/{id}/session/release` | `delivery/http/handler.go:702` | `authorizeDocumentScope` |
| POST | `/api/v1/documents/{id}/session/force-release` | `delivery/http/handler.go:726` | `authorizeDocumentScope` |
| POST | `/api/v1/documents/{id}/autosave/presign` | `delivery/http/handler.go:752` | `authorizeDocumentScope` + rate-limit when wired |
| POST | `/api/v1/documents/{id}/autosave/commit` | `delivery/http/handler.go:~800` | `authorizeDocumentScope` + rate-limit |
| GET | `/api/v1/documents/{id}/checkpoints` | `delivery/http/handler.go:~870` | `authorizeDocumentScope` |
| POST | `/api/v1/documents/{id}/checkpoints` | `delivery/http/handler.go:~890` | `authorizeDocumentScope` |
| POST | `/api/v1/documents/{id}/checkpoints/{version}/restore` | `delivery/http/handler.go:~910` | `authorizeDocumentScope` |
| GET | `/api/v1/documents/{id}/revision-history` | `delivery/http/handler.go:~950` | `authorizeDocumentScope` |
| GET | `/api/v1/documents/{id}/revisions/{rid}/url` | `delivery/http/handler.go:~975` | `authorizeDocumentScope` |
| GET | `/api/v1/documents/{id}/comments` | `delivery/http/handler.go:~990` | `authorizeDocumentScope` |
| POST | `/api/v1/documents/{id}/comments` | `delivery/http/handler.go:~1005` | `authorizeDocumentScope` |
| PATCH | `/api/v1/documents/{id}/comments/{library_id}` | `delivery/http/handler.go:~1030` | `authorizeDocumentScope` |
| DELETE | `/api/v1/documents/{id}/comments/{library_id}` | `delivery/http/handler.go:1058` | `authorizeDocumentScope` |
| POST | `/api/v1/documents/{id}/export/pdf` | `delivery/http/export_handler.go:40` | tenant from ctx only (no doc-ownership check) |
| GET | `/api/v1/documents/{id}/export/docx-url` | `delivery/http/export_handler.go:41` | same |
| GET | `/api/v1/documents/{id}/fill-in-schema` | `http/fillin_handler.go:38` | tier-2 `document.edit` in service |
| GET | `/api/v1/documents/{id}/placeholders` | `http/fillin_handler.go:39` | tier-2 `document.edit` |
| PUT | `/api/v1/documents/{id}/placeholders/{pid}` | `http/fillin_handler.go:40` | tier-2 `document.edit` |
| GET | `/api/v1/documents/{id}/view` | `http/view_handler.go:31` | tier-2 `document.view` (viewable statuses only) |
| POST | `/api/v1/documents/{id}/reconstruct` | `http/reconstruct_handler.go:28` | tier-2 `document.edit` |
| GET | `/api/v1/documents/{id}/placeholder-options/{pid}` | `http/placeholder_options_handler.go:35` | no authz guard — tenant from ctx only |
| POST | `/api/v1/documents/{id}/pdf-complete` | `http/pdf_webhook_handler.go:38` | HMAC-SHA256 `X-Docgen-Signature`; no session auth |

---

## 4. Logic Flows

### 4.1 `GET /api/v1/documents` — List Documents (keyset-paginated)

1. `handler.go:178` — `tenantIDFromReq(r)` → `tenant.FromContext`.
2. `handler.go:185` — `h.isSystemAdmin(ctx, callerUserID, tenantID)` → `CapabilityChecker.IsSystemAdmin`.
3. `handler.go:190` — `parseListOptions`: parses `cursor` (opaque), `limit` (1–100, default 20), `status[]`, `area_code`, `profile_code`, `q`, `include_archived`. Non-admin: sets `opts.CreatedBy = callerUserID`.
4. `handler.go:197` — `h.svc.ListDocumentsPaginated(ctx, tenantID, effectiveUserID, opts)`.
5. `service.go:~505` — delegates to `repo.ListDocumentsPaginated(ctx, tenantID, opts)` + `repo.CountDocuments(ctx, tenantID, opts)`.
6. `repository.go:465` — `buildDocumentFilter` assembles `WHERE` clause; `pagination.DecodeCursor(opts.Cursor)` decodes `(updated_at, id)` pair; appends `AND (updated_at, id) < ($n::timestamptz, $n)`; `LIMIT limit+1` probe.
7. `repository.go:511` — trims probe row, sets `hasMore=true` when `len(out) > limit`.
8. `handler.go:208` — builds `nextCursor = pagination.EncodeCursor(last.UpdatedAt.UTC().RFC3339Nano, last.ID)`.
9. Returns `200 {items, page:{next_cursor, has_more}, total}`.

### 4.2 `PATCH /api/v1/documents/{id}` — Rename Document

1. `handler.go:400` — `withAdminCtx(r)` enriches context with `iamdomain.RolesFromContext` roles.
2. `handler.go:403` — `authorizeDocumentScope`: resolves `tenantID`, checks `isSystemAdmin`; if not admin calls `svc.IsDocumentOwner`; 403 otherwise.
3. `handler.go:411` — JSON-decode `{name}`; length guard `isValidBoundedText(req.Name, 255)`.
4. `handler.go:420` — `h.svc.RenameDocument(ctx, tenantID, userID, docID, req.Name)`.
5. `service.go:556` — validates name non-empty/≤255. Loads document; guards `Status == draft`.
6. `service.go:575` — when `s.db != nil` (production): `BeginTx` → `repo.UpdateDocumentNameTx` → `audit.WriteTx(ctx, tx, ...)` → `tx.Commit`. (When `s.db == nil` / test path: `UpdateDocumentName` + `audit.Write` outside tx.)
7. `repository.go:305` — `UpdateDocumentNameTx`: `authz.SeedTxIdentity` + `authz.Require(CapDocumentEdit, areaCode)` + `UPDATE documents SET name=$2, updated_at=now()`.
8. `handler.go:426` — re-fetches document via `svc.GetDocument`; returns **`200 {document_object}`** (see drift §11).

### 4.3 `POST /api/v1/documents/{id}/finalize` — Submit for Review

1. `handler.go:435` — reads `Idempotency-Key` header; 400 if missing/invalid.
2. `handler.go:451` — `idempotency.RequestHash(r)` hashes body for payload-change detection.
3. `handler.go:474` — `idempStore.BeginReplay`: `ErrConflict` (same key, different hash) → 422; live replay → return replay body + `Idempotent-Replay: true`.
4. `handler.go:501` — `withAdminCtx(r)` + `authorizeDocumentScope` (tier-1).
5. `handler.go:512` — inline SQL: `SELECT revision_version, revision_number, controlled_document_id FROM documents WHERE id=$1 AND tenant_id=$2 AND status='draft'`; `ErrNoRows` → 409.
6. `handler.go:529` — inline SQL: `SELECT profile_code FROM controlled_documents WHERE id=$1 AND tenant_id=$2`.
7. `handler.go:545` — inline SQL: `SELECT id FROM approval_routes WHERE tenant_id=$1 AND profile_code=$2 AND active=true ORDER BY version DESC LIMIT 1`; `ErrNoRows` → 409.
8. `handler.go:564` — inline SQL: `SELECT COALESCE(content_hash,'') FROM document_revisions WHERE document_id=$1 ORDER BY created_at DESC, id DESC LIMIT 1`; `ErrNoRows` treated as empty string (not an error).
9. `handler.go:575` — `h.submitSvc.SubmitRevisionForReview(ctx, h.db, SubmitRequest{...})` (approval subtree): tx → authz.SeedTxIdentity → authz.Require(CapDocumentSubmit, areaCode) → InsertInstance → InsertStageInstances → `UPDATE documents SET status='under_review'` (fires `enforce_snapshot_on_submit_trg`) → emit governance_events → Commit.
10. `handler.go:591` — `idempStore.CompleteReplay(handle, 201, body)`.
11. Returns `201 {"instance_id": "..."}`.

### 4.4 `PUT /api/v1/documents/{id}/placeholders/{pid}` — Fill-in Placeholder Value

1. `http/fillin_handler.go:40` — `PutPlaceholderValue` extracts `tenantID`, `actorID` from ctx.
2. Calls `requireDocEditDraft(ctx, db, tenantID, actorID, docID)` (`application/fillin_authz.go:19`): opens tx, `authz.WithCapCache` + `authz.SeedTxIdentity`, `LoadDocumentAreaCode`, `authz.Require(CapDocumentEdit, areaCode)`; tx rolled back after check.
3. `application/fillin_service.go` — `SetPlaceholderValue`: loads schema via `SnapshotSchemaReader`; validates placeholder ID and type (regex for `date`, format for `number`, IAM user check for `user` type); calls `FillInRepository.UpsertValue`.
4. `repository/fillin_repository.go:76` — `INSERT INTO document_placeholder_values ... ON CONFLICT (...) DO UPDATE SET value_text=$4, ...`.
5. Returns `200 {placeholder_id, value_text}`.

### 4.5 Atomic CD+Document Create — `CreateDocumentTx` Port

Called by `controlled-documents` (`CDDocumentInitializer.CloneTemplate`) inside a caller-owned transaction.

1. `application/cd_initializer.go:40` — receives caller-owned `*sql.Tx` from controlled-documents.
2. `service.go:~350` — `cloneIntoTx`: resolves template version ID (override priority → profile default); loads docx key via `TemplateReader.GetPublishedVersion`; `snapshotSvc.ResolveTemplate` returns `TemplateSnapshot` + required placeholder list.
3. `repository.go:124` — `CreateDocumentTx(ctx, tx, &doc, contentHash, docxKey, phs)`:
   a. `authz.WithCapCache(ctx)` + `authz.SeedTxIdentity(ctx, tx, tenantID, createdBy)`.
   b. `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))` on `"tenant:cdID"` for `revision_number` serialization.
   c. `authz.Require(CapDocumentCreate, docArea)` (tier-2 + Postgres tripwire).
   d. `INSERT INTO documents ... SELECT ... COALESCE(MAX(revision_number)+1, 0) ... RETURNING id`.
   e. `INSERT INTO editor_sessions ... RETURNING id`.
   f. `INSERT INTO document_revisions ... RETURNING id`.
   g. `authz.Require(CapDocumentEdit, docArea)` before pointer UPDATEs.
   h. `UPDATE editor_sessions SET last_acknowledged_revision_id=$1`.
   i. `UPDATE documents SET current_revision_id=$1, active_session_id=$2`.
   j. `UPDATE documents SET placeholder_schema_snapshot=..., body_docx_snapshot_s3_key=..., ...` (all 6 snapshot columns at once).
   k. `INSERT INTO document_placeholder_values ... ON CONFLICT DO NOTHING` (seed required placeholders).
4. Caller (controlled-documents) commits — CD row, sequence counter, and document row all commit atomically.

---

## 5. Dependencies

### Outbound (imports from outside `internal/modules/documents`)

| Import path | Used in | Why |
|---|---|---|
| `metaldocs/internal/modules/controlleddocuments/domain` | `application/service.go`, `application/cd_initializer.go`, `delivery/http/handler.go` | `ControlledDocument` type, `ErrCDNotFound`, `ErrCDNotActive`, template resolution |
| `metaldocs/internal/modules/iam/domain` | `application/ports.go`, `application/fillin_authz.go`, `delivery/http/handler.go`, `repository/repository.go` | Capability consts, `RolesFromContext`, `UserIDFromContext`, `WithAuthContext` |
| `metaldocs/internal/modules/iam/authz` | `repository/repository.go`, `application/fillin_authz.go`, `application/view_service.go`, `application/reconstruct_service.go`, `jobs/` | `SeedTxIdentity`, `Require`, `WithCapCache`, `WithBackgroundBypass` |
| `metaldocs/internal/modules/iam/application` | `delivery/http/handler.go` | `ErrCapabilityDenied` sentinel in `mapErr` |
| `metaldocs/internal/modules/templates/domain` | `application/service.go`, `repository/fillin_repository.go`, `application/snapshot_service.go` | `Placeholder` type, `ErrUnknownResolver` |
| `metaldocs/internal/modules/render/fanout` | `application/freeze_service.go`, `application/reconstruct_service.go`, `http/reconstruct_handler.go` | `FanoutRequest/Response`, `ReconstructionEntry` |
| `metaldocs/internal/modules/render/resolvers` | `application/context_builder.go`, `repository/resolver_readers.go` | `ResolveInput`, resolver interfaces, registry |
| `metaldocs/internal/platform/tenant` | `delivery/http/handler.go`, `http/` package | `tenant.FromContext` |
| `metaldocs/internal/platform/httpresponse` | `delivery/http/handler.go` | `WriteJSON` |
| `metaldocs/internal/platform/problem` | `delivery/http/handler.go`, `http/pdf_webhook_handler.go` | `problem.Write`, `problem.New`, `problem.Code*` consts |
| `metaldocs/internal/platform/idempotency` | `delivery/http/handler.go` | `New`, `BeginReplay`, `CompleteReplay`, `FailReplay`, `IsValidKey`, `RequestHash`, `ErrConflict` |
| `metaldocs/internal/platform/ratelimit` | `delivery/http/handler.go`, `delivery/http/export_handler.go`, `module.go` | `Middleware.Limit`, route key consts |
| `metaldocs/internal/platform/pagination` | `delivery/http/handler.go`, `repository/repository.go` | `EncodeCursor`, `DecodeCursor`, `ErrInvalidCursor` |
| `metaldocs/internal/platform/servicebus` | `application/export_service.go` | `ConvertPDFRequest`, `ConvertPDFResult` |
| `github.com/google/uuid` | `domain/comment.go`, `repository/repository.go` | UUID generation |
| `github.com/jackc/pgx/v5/pgconn` | `repository/repository.go` | `PgError` SQLSTATE 22P02 detection |
| `github.com/lib/pq` | `repository/repository.go` | `pq.Array` for `status = ANY($n)` |
| `github.com/oapi-codegen/runtime` | `api/api.gen.go` | Generated code runtime |

### Inbound (who imports `internal/modules/documents` — grep-verified)

| Importer | What it uses |
|---|---|
| `apps/api/cmd/metaldocs-api/main.go` | `documents.New`, routes, `Repo()`, `Service`, `NewCDDocumentInitializer`, `jobs.Start*`, `NewFreezeService`, handler wiring |
| `apps/api/cmd/metaldocs-api/reauth.go` | Application types (reauth flow) |
| `apps/api/internal/wiring/documents.go` | `application.CapabilityChecker`, `NewCapabilityChecker` adapter |
| `apps/worker/cmd/metaldocs-worker/main.go` | Worker-side repository and service refs |
| `apps/jobs/cmd/metaldocs-jobs/main.go` | Scheduled publish job wiring (approval subtree) |
| `internal/modules/iam/integration_test.go` | Cross-module integration test |
| `internal/modules/jobs/stuck_instance_watchdog/job.go` | Approval repo types from documents |
| `internal/platform/objectstore/document_presigner.go` | Implements `application.Presigner` |
| `internal/platform/docgenv2/templates_snapshot_reader.go` | Implements `application.SnapshotTemplateReader` |

---

## 6. Persistence

### Tables owned (core subtree; approval tables excluded)

| Table | Key columns | Notes |
|---|---|---|
| `documents` | `id UUID PK`, `tenant_id`, `status`, `controlled_document_id`, `current_revision_id`, `active_session_id`, `revision_number`, `revision_title`, `process_area_code_snapshot`, `profile_code_snapshot`, `code`, `placeholder_schema_snapshot` + hash cols, `body_docx_snapshot_s3_key`, `final_pdf_s3_key`, `values_frozen_at`, `archived_at` | Tripwire `trg_require_cap_asserted` on INSERT/UPDATE (migration `0188_tripwire_extend.sql:196–199`) |
| `editor_sessions` | `id UUID PK`, `tenant_id`, `document_id`, `user_id`, `status`, `expires_at`, `last_acknowledged_revision_id` | `tenant_id` added by migration `0211_editor_sessions_tenant_id.sql` |
| `document_revisions` | `id UUID PK`, `document_id`, `parent_revision_id`, `session_id`, `storage_key`, `content_hash`, `form_data_snapshot`, `file_size_bytes`, `page_count`, `page_count_source` | Autosave artifact chain; never governs business revision history |
| `document_checkpoints` | `id UUID PK`, `document_id`, `revision_id`, `version_num`, `label`, `created_by` | Editor save points |
| `document_placeholder_values` | `tenant_id`, `revision_id`, `placeholder_id`, `value_text`, `source`, `computed_from`, `resolver_version`, `inputs_hash` | **T-009**: `revision_id REFERENCES documents(id)` — FK target is wrong (migration `0152_placeholder_fillin_columns.sql:51`) |
| `document_exports` | `id UUID PK`, `document_id`, `revision_id`, `composite_hash UNIQUE w/ document_id`, `storage_key`, `size_bytes`, `paper_size`, `landscape`, `docgen_v2_ver` | PDF export cache |
| `document_comments` | `id UUID PK`, `tenant_id`, `document_id`, `library_comment_id`, `parent_library_id`, `author_id`, `author_display`, `content_json`, `resolved_at`, `resolved_by` | Review comments; unresolved comments block final approval (approval subtree check) |

### Key query patterns

- **Keyset cursor**: `repository.go:465` — `WHERE (updated_at, id) < ($n::timestamptz, $n) ORDER BY updated_at DESC, id DESC LIMIT n+1`.
- **Advisory lock**: `repository.go:134` — `pg_advisory_xact_lock(hashtextextended($1, 0))` for `revision_number` allocation; auto-releases at commit/rollback.
- **Tripwire GUC path**: `authz.SeedTxIdentity` sets `metaldocs.tenant_id`/`actor_id`; `authz.Require` appends to `metaldocs.asserted_caps`; trigger `enforce_capability_asserted` reads GUC and raises if capability absent.
- **RowsAffected enforcement**: `snapshot_repository.go:43` — `requireRowsAffected` called on all `WriteSnapshot`, `WriteFreeze`, `WriteFinalDocx`, `MarkArchived`, `Unarchive` writes; returns error on 0 rows.
- **Snapshot trigger**: `migrations/0152_*.sql:47` — `enforce_snapshot_on_submit_trg` blocks `status='under_review'` UPDATE when `placeholder_schema_snapshot IS NULL`.

---

## 7. Config & Environment

The documents module reads **no environment variables directly**. All configuration arrives through the `Dependencies` struct passed to `module.New()` and wired in `apps/api/cmd/metaldocs-api/main.go`.

Notable defaults applied in the module:
- `module.go:72–79`: `DocgenVer` defaults to `"docgen-v2@0.4.0"`, `GrammarVer` defaults to `"grammar-v1"` when empty strings are passed.
- `handler.go:106–108`: `idempotency.New(db, "POST /api/v1/documents/{id}/finalize")` is constructed inline in the handler when `idempFinalize` is nil.

Indirect env dependencies (resolved in main.go / platform packages):
- Object store credentials → `internal/platform/objectstore`.
- Idempotency store → `internal/platform/idempotency` (Postgres-backed).
- PDF webhook HMAC secret → passed as `string` argument from main.go env read.
- Docgen service URL → `internal/platform/servicebus`.

---

## 8. Concurrency & Async

| Mechanism | File:line | Notes |
|---|---|---|
| Goroutine ticker — session sweep | `jobs/session_sweeper.go:12` | `StartSessionSweeper` — background goroutine; `authz.WithBackgroundBypass`; `ExpireStaleSessions` every `interval` |
| Goroutine ticker — orphan pending sweep | `jobs/orphan_pending_sweeper.go:12` | `StartOrphanPendingSweeper` — background goroutine; `authz.WithBackgroundBypass`; `DeleteExpiredPending` every `interval` |
| Transactional outbox enqueue | `application/freeze_service.go:215` | `Pin()` calls `materializeOutbox.Enqueue(ctx, tx, ...)` inside approval tx — outbox pattern; async materialization runs in worker |
| Idempotency replay store | `delivery/http/handler.go:474` | `BeginReplay/CompleteReplay/FailReplay`; deferred `FailReplay` cleanup guarded by `idempReleased` flag |
| Advisory lock | `repository/repository.go:134` | `pg_advisory_xact_lock` inside create tx for `revision_number` serialization |
| Approval tx composition | `delivery/http/handler.go:575` | `h.db` passed to `SubmitRevisionForReview`; approval subtree opens the tx and commits atomically |
| `defer tx.Rollback()` | throughout repository | Standard deferred rollback; commit path returns before defer fires |

---

## 9. Error Handling & Observability

### Error handling

- **`mapErr(err) (int, problem.Code)`** at `delivery/http/handler.go:1158` — sentinel-switch covering all domain errors, `ErrCapabilityDenied`, OCC errors, export errors; fallthrough to `500 internal_error`.
- **`httpErr(w, status, code)`** at `delivery/http/handler.go:1202` — delegates to `problem.Write(w, problem.New(status, code, string(code)))`. T-001 (RFC 9457 migration) is closed; `httpErr` now emits `application/problem+json`.
- **`httpErrDetail`** at `delivery/http/handler.go:1208` — adds `WithDetail` for runtime context (e.g., missing profile name, route name in approval-route-missing case).
- **`writeFillInError`** in `http/fillin_handler.go` — separate RFC 9457 writer for the legacy `http/` sub-package; maps `authz.ErrCapDenied`, `domain.ErrNotFound`, defaults to 500.
- **Postgres tripwire raises**: propagate as `pgconn.PgError`; no explicit sentinel in `mapErr` — surface as `500 internal_error`.
- **`string.HasPrefix(err.Error(), "form_data_invalid")`** at `handler.go:1195` — brittle string-prefix match for form validation errors; not a sentinel pattern.

### Observability

- **stdlib `log.Printf`**: used in `jobs/` sweepers, `handler.go` for non-critical paths (finalize idempotency errors, duplicate failures). No structured fields.
- **`log/slog`**: `view_handler.go`, `reconstruct_handler.go` use `slog.ErrorContext` for 500-class errors.
- **Governance events**: `EventEmitter.Emit` (approval subtree) → INSERT `governance_events`; QMS audit trail.
- **`Audit.Write` / `Audit.WriteTx`**: consumer-port interface at `service.go:82–85`; called for rename (tx-wrapped), archive, duplicate, autosave-commit, session-force-release, checkpoint ops. T-007 (no compile-time contract with audit domain) remains open.
- **No trace correlation** in this module; request-id propagation is upstream middleware responsibility.
- **No metrics emitted** in this module; covered by HTTP observability middleware upstream.

---

## 10. Legacy / Duplication / Smell Flags

- **FLAG-01 · Dual route registration — 18 copy-pasted `HandleFunc` calls**
  - WHAT: `RegisterRoutes` and `RegisterRoutesWithRateLimit` at `delivery/http/handler.go:112–176` register the same 18 non-rate-limited routes twice. Only the 2 presign/commit and 1 export-pdf routes differ (rate-limit wrapper). The remaining 15 routes are verbatim duplicates.
  - WHERE: `delivery/http/handler.go:112–176`
  - WHY: stdlib mux last-wins makes this safe today but divergence risk on every future route addition. Tech-debt T-004 registers PATCH as the known duplicate; this flag extends the observation to the entire pattern. RF-XXX (no registered RF ID).

- **FLAG-02 · Legacy `http/` sub-package coexists with `delivery/http/`**
  - WHAT: `internal/modules/documents/http/` (`package documentshttp`) has its own HTTP utility helpers (`tenantID`, `actorID`, `requestID`, `writeFillInJSON`, `writeFillInError`) parallel to `delivery/http/handler.go`'s `tenantIDFromReq`, `userIDFromReq`, `httpErr`. Two uncoordinated error writer families exist; the legacy package predates the canonical delivery layout.
  - WHERE: `http/*.go` vs `delivery/http/handler.go`
  - WHY: Maintenance surface: a change to RFC 9457 problem shapes must be applied in two places.

- **FLAG-03 · `GeneratedServerAdapter` discards all typed codegen parameters**
  - WHAT: `delivery/http/generated_adapter.go` implements `documentsapi.ServerInterface` but every method calls `a.legacy.ServeHTTP(w, r)`, ignoring typed path UUIDs and query param structs parsed by the generated wrapper. The generated validation and types provide zero runtime benefit until handlers migrate.
  - WHERE: `delivery/http/generated_adapter.go:1–143`
  - WHY: ADR 0012 interim state. Documented as T-010 (now partially stale — the mount point was added; T-010's "not mounted" claim is no longer accurate).

- **FLAG-04 · `domain/model.go` defines only 3 of 8 live `DocumentStatus` values**
  - WHAT: `model.go:8–13` defines `draft`, `under_review`, `archived`. Statuses `approved`, `published`, `superseded`, `obsolete`, `scheduled`, `rejected` appear only in `api.gen.go:82–89` as `DocumentSummaryStatus`. The domain state machine (`state.go`) covers only pre-approval transitions.
  - WHERE: `domain/model.go:8–13`, `domain/state.go`
  - WHY: Incomplete domain model; the full lifecycle is split between the domain layer (pre-approval) and the approval subtree (post-approval) with no single authoritative list of all valid statuses.

- **FLAG-05 · Minimal `New()` constructor is a subset of `NewService` and silently incomplete**
  - WHAT: `application/service.go:120` — `New(r, p, t, fv, a) *Service` omits `caps`, `controlledDocumentReader`, `profileTemplates`. Any caller that reaches `CreateDocument`, `DuplicateDocument`, or cap checks will hit a nil guard sentinel error at runtime, not a construction-time panic.
  - WHERE: `application/service.go:120–128`
  - WHY: Constructor surface footgun; `New` vs `NewService` naming does not signal the completeness difference.

- **FLAG-06 · Inline SQL in `finalizeDocument` handler bypasses repository layer**
  - WHAT: `delivery/http/handler.go:512–573` contains four raw `db.QueryRowContext` calls (documents status+revision, profile code from CD, active approval route, latest content hash) directly in the HTTP handler body.
  - WHERE: `delivery/http/handler.go:512–573`
  - WHY: SQL in the delivery layer breaks the layering invariant (backend-canon §1.2); cannot be unit-tested without a real DB; duplicates query logic that could be centralized in repository.

- **FLAG-07 · `document_placeholder_values.revision_id` FK targets wrong table (T-009)**
  - WHAT: Migration `0152_placeholder_fillin_columns.sql:51` declares `revision_id REFERENCES documents(id)`. The column name implies a reference to `document_revisions(id)`.
  - WHERE: `migrations/0152_placeholder_fillin_columns.sql:51`
  - WHY: FK constraint enforces nothing useful in practice (UUID collision probability negligible). T-009 registered.

- **FLAG-08 · `SnapshotService.SnapshotFromTemplate` is deprecated but retained**
  - WHAT: `application/snapshot_service.go:46` marked `// Deprecated:` — post-commit snapshot write, retained only for backfill scripts. No reference to which scripts.
  - WHERE: `application/snapshot_service.go:46`
  - WHY: Dead code risk if backfill scripts have run; no compile-time callsite enumeration.

- **FLAG-09 · `placeholder_options_handler.go` has no authz guard**
  - WHAT: `http/placeholder_options_handler.go:38` `HandleGetOptions` reads only `tenantID` from context; no tier-1 role gate and no tier-2 `authz.Require` call before querying placeholder schema and IAM user list.
  - WHERE: `http/placeholder_options_handler.go:38`
  - WHY: Any authenticated tenant-member can enumerate IAM user lists and placeholder schemas without ownership or capability check. Potential information disclosure.

- **FLAG-10 · Error sentinel definitions split across two domain files**
  - WHAT: Twelve sentinels in `domain/model.go:112–132`; two more (`ErrValidationFailed`, `ErrEffectiveDateMissing`) in `domain/errors.go:5–6`. No structural reason for the split.
  - WHERE: `domain/errors.go:1–7`, `domain/model.go:112`
  - WHY: Minor inconsistency; new contributors looking for all domain errors will miss `errors.go`.

- **FLAG-11 · `ForceReleaseSession` audit write outside transaction**
  - WHAT: `application/service.go:803–808` calls `repo.ForceReleaseSession` (own internal tx) then `audit.Write` in a separate call outside any encompassing transaction.
  - WHERE: `application/service.go:803–808`
  - WHY: Same atomicity gap as T-005/rename (which was fixed). `ForceReleaseSession` was not included in Plan 6a scope.

- **FLAG-12 · `Archive` audit write outside transaction**
  - WHAT: `application/service.go:843–848` calls `repo.MarkArchived` (which opens and commits its own tx) then `audit.Write` outside that tx.
  - WHERE: `application/service.go:843–848`
  - WHY: Same atomicity gap: archive is committed before audit event; a crash between the two leaves a mutated row with no audit trail.

- **FLAG-13 · `err.Error()` string-prefix match for form validation errors**
  - WHAT: `delivery/http/handler.go:1195` — `strings.HasPrefix(err.Error(), "form_data_invalid")` in `mapErr`. Not a sentinel pattern; brittle against error message changes.
  - WHERE: `delivery/http/handler.go:1195`
  - WHY: Any wrapping of the error (e.g., `fmt.Errorf("...: %w", formErr)`) will cause the match to fail, silently producing a 500 instead of the intended 422.

---

## 11. Wiki Drift

| Wiki doc claim | Code reality |
|---|---|
| `documents.md §6.2` sequence diagram (~line 324): `H-->>C: 204 No Content` for `renameDocument` | `delivery/http/handler.go:426–432` calls `svc.GetDocument` and returns `httpresponse.WriteJSON(w, http.StatusOK, doc)` — response is `200 + document body`, not `204 No Content`. |
| `documents.md §8.4` (~line 423): "Offset only: `page` + `pageSize` (default 1/20, cap 50). Repo LIMIT/OFFSET at `repository.go:343`." | `repository.go:465` uses **keyset cursor** pagination (`(updated_at, id) < (...)`, not LIMIT/OFFSET). Handler `parseListOptions` uses `limit` parameter (default 20, cap 100, not 50). Response shape is `{items, page:{next_cursor, has_more}, total}`. The offset-pagination description is stale. |
| `documents-tech-debt.md T-010` (~line 85): "`api.gen.go` is generated and committed but no route is mounted via the generated `ServerInterface`" | `module.go:120` and `module.go:134` mount routes via `documentsapi.HandlerWithOptions(dhttp.NewGeneratedServerAdapter(legacyMux), ...)`. The generated wrapper IS mounted. T-010 "not mounted" claim is stale. FLAG-03 captures the current accurate state (mounted but typed params discarded). |

---

## 12. Open Questions

- **[runtime-unverified]** FLAG-09 (placeholder options no-authz): whether the absence of an authz guard on `HandleGetOptions` is an intentional product decision (all tenant users may see user lists) or an oversight. Cannot be determined from code alone.
- **[runtime-unverified]** PDF webhook route (`POST /api/v1/documents/{id}/pdf-complete`) is not middleware-protected — only HMAC guards it. Whether this route is network-isolated (not exposed externally) is a deployment question outside the code.
- **[runtime-unverified]** `SnapshotService.SnapshotFromTemplate` deprecated method (FLAG-08): whether any production backfill script still calls it and whether it is safe to remove.
- **[runtime-unverified]** `FreezeService.Freeze` (legacy synchronous path at `freeze_service.go:300`): whether any production call site still invokes `Freeze` rather than the `Pin`+`Materialize` async split.
- **[genuine unknown]** `domain/state.go:CanTransitionDocument` covers only 3 statuses. Whether this reflects an intentional domain boundary (pre-approval in documents domain, post-approval in approval domain) or an incomplete state machine is not resolvable from code alone — both interpretations are consistent with the structure.
