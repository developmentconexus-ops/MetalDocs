# Stage-1 Audit Artifact — Render Pipeline

> **Produced:** 2026-06-10
> **Scope:** `internal/modules/render`, `internal/platform/render`, `internal/platform/docgenv2`, `apps/docx-renderer`
> **Read-only snapshot.** No redesign proposals; all claims anchored to file:line.

---

## 1. Identity & Purpose

The render pipeline is responsible for converting a document revision's DOCX template and its resolved placeholder values into a frozen, immutable DOCX artifact and then a PDF. It spans two processes: the Go API/worker binary and the TypeScript `docx-renderer` sidecar service. The Go side owns placeholder resolution (via a registry of `ComputedResolver` implementations), freeze coordination (via `FreezeService`), outbox-based async dispatch, and Gotenberg-based PDF conversion. The TypeScript `docx-renderer` owns the actual DOCX token substitution and composition-block injection through the `@eigenpal/docx-js-editor` headless API, writing the resulting frozen DOCX to MinIO directly. The pipeline is split into two asynchronous phases following ADR 0015: a synchronous "Pin" phase (in-transaction, no network calls) and an async "Materialize" phase (worker → docx-renderer HTTP → MinIO → PDF outbox).

There is no "docgen v1" remnant in the codebase. The `internal/platform/docgenv2` package name is a historical artifact of the v2 naming era; the service it references is now called `docx-renderer`. The event type constant `EventTypePDFConvert = "docgen_v2_pdf"` carries the v2 name string on the wire.

---

## 2. File Inventory

### `internal/modules/render/resolvers/` — Computed-placeholder resolution engine

| File | Role |
|---|---|
| `resolver.go` | Core types: `ComputedResolver` interface, `ResolveInput`, `ResolvedValue`, reader interfaces (`RegistryReader`, `RevisionReader`, `WorkflowReader`, `DocumentReader`), helper guard functions |
| `registry.go` | Thread-safe `Registry` struct: `Register`, `Get`, `Known` — holds all `ComputedResolver` implementations at startup |
| `builtins.go` | `RegisterBuiltins` — registers the 8 built-in resolvers into a `*Registry` |
| `hash.go` | `hashInputs` — SHA-256 inputs hash for `ResolvedValue.InputsHash` using `encoding/json` + `crypto/sha256` |
| `doc_code.go` | `DocCodeResolver` — resolves `doc_code` (v1) from `RegistryReader.GetControlledDocument` |
| `doc_title.go` | `DocTitleResolver` — resolves `doc_title` (v1) from `DocumentReader.GetDocumentTitle` |
| `revision_number.go` | `RevisionNumberResolver` — resolves `revision_number` (v1) from `RevisionReader.GetRevisionNumber` |
| `effective_date.go` | `EffectiveDateResolver` — resolves `effective_date` (v1) from `RevisionReader.GetEffectiveFrom`, formats as `2006-01-02` |
| `controlled_by_area.go` | `ControlledByAreaResolver` — resolves `controlled_by_area` (v2) from `ResolveInput.AreaNameSnapshot`, falls back to `AreaCodeSnapshot` |
| `author.go` | `AuthorResolver` — resolves `author` (v1) from `RevisionReader.GetAuthor`, falls back to UserID if DisplayName empty |
| `approvers.go` | `ApproversResolver` — resolves `approvers` (v1) from `WorkflowReader.GetApprovers`; falls back to `"[aguardando aprovação]"` if no names |
| `approval_date.go` | `ApprovalDateResolver` — resolves `approval_date` (v1) from `WorkflowReader.GetFinalApprovalDate`; errors if date is zero |
| `resolver_test.go` | Contract test baseline |
| `contract_test.go` | Contract test for resolver key/version shape |
| `doc_code_test.go` | Unit tests for DocCodeResolver |
| `doc_title_test.go` | Unit tests for DocTitleResolver |
| `revision_number_test.go` | Unit tests for RevisionNumberResolver |
| `effective_date_test.go` | Unit tests for EffectiveDateResolver |
| `controlled_by_area_test.go` | Unit tests for ControlledByAreaResolver |
| `author_test.go` | Unit tests for AuthorResolver |
| `approvers_test.go` | Unit tests for ApproversResolver |
| `approval_date_test.go` | Unit tests for ApprovalDateResolver |
| `builtins_test.go` | Tests for RegisterBuiltins |

### `internal/modules/render/fanout/` — DOCX fanout client, outbox workers, reconstruction

| File | Role |
|---|---|
| `client.go` | `Client` — HTTP client to `POST /render/fanout` on docx-renderer; carries `FanoutRequest`/`FanoutResponse` shapes; sends `X-Service-Token` header |
| `reconstruction.go` | `ReconstructService` — forensic re-render: calls docx-renderer with stored inputs, compares hash to original, appends JSON entry to `documents.reconstruction_attempts` |
| `pdf_dispatcher.go` | `PDFDispatcher` — publishes `docgen_v2_pdf` event via `messaging.Publisher`; called post-approval |
| `pdf_dispatch_adapter.go` | `PDFDispatchAdapter` — bridges `approval/application.PDFDispatchInvoker` to `PDFDispatcher` by first reading the frozen DOCX S3 key via `DocxKeyReader` |
| `pdf_outbox_repository.go` | `PDFOutboxRepository` — CRUD for `metaldocs.pdf_dispatch_outbox`; `Enqueue` (ON CONFLICT DO NOTHING), `ClaimPending` (FOR UPDATE SKIP LOCKED), `MarkDispatched`, `MarkFailed` (retry or finalize), `ReadState`, `ResetStaleClaims` |
| `pdf_outbox_worker.go` | `PDFOutboxWorker` — polls `pdf_dispatch_outbox` every 5s, claims up to 10 rows, publishes `docgen_v2_pdf` events; exponential backoff (base 30s, cap 30m); max 5 attempts |
| `materialize_outbox_repository.go` | `MaterializeOutboxRepository` — identical CRUD pattern but targets `metaldocs.materialize_dispatch_outbox`; supports the async Pin→Materialize split (ADR 0015) |
| `materialize_outbox_worker.go` | `MaterializeOutboxWorker` — polls `materialize_dispatch_outbox` every 5s, publishes `docx_materialize` events; same retry/backoff shape as `PDFOutboxWorker` |
| `client_test.go` | Unit tests for fanout HTTP client |
| `pdf_dispatch_adapter_test.go` | Unit tests for adapter |
| `pdf_dispatcher_test.go` | Unit tests for dispatcher |
| `pdf_outbox_repository_test.go` | Unit tests for outbox repo |
| `pdf_outbox_worker_test.go` | Unit tests for outbox worker |
| `reconstruction_test.go` | Unit tests for ReconstructService |
| `reconstruction_drift_test.go` | Drift/regression test for reconstruction shape |

### `internal/platform/render/gotenberg/` — Gotenberg HTTP client

| File | Role |
|---|---|
| `client.go` | `Client` — `ConvertHTMLToPDF` (Chromium route) and `ConvertDocxToPDF`/`ConvertDocxToPDFWithOptions` (LibreOffice route); 30s timeout; 64 MiB PDF body limit; paper size `A4`/`Letter` override |
| `client_test.go` | Unit tests |

### `internal/platform/docgenv2/` — Template/snapshot readers for the docgenv2 era (renamed "templates" module integration)

| File | Role |
|---|---|
| `template_reader.go` | `TemplateReader` — reads published template version DOCX + schema from legacy `template_versions`/`templates` tables; fetches schema JSON from MinIO (1 MiB cap); uses `systemTemplateTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"` for system templates |
| `templates_reader.go` | `TemplatesTemplateReader` — reads published DOCX key from `templates_template_version`/`templates_template` tables (new templates module); schema always returns `""`; `FanoutTemplateReader` chains primary→secondary with `sql.ErrNoRows` fallback |
| `templates_snapshot_reader.go` | `TemplatesSnapshotReader` — implements `documents/application.SnapshotTemplateReader` from `templates_template_version`; loads `placeholder_schema` JSON and DOCX key for snapshot creation; `CompositionJSON` hardcoded to `{}` |
| `template_reader_test.go` | Unit tests for TemplateReader |
| `templates_snapshot_reader_test.go` | Unit tests for TemplatesSnapshotReader |

### `apps/docx-renderer/` — TypeScript/Node DOCX substitution + composition service

| File | Role |
|---|---|
| `src/index.ts` | Entry point: builds Fastify app, registers service-auth hook, health route, S3 factory, and routes; listens on `DOCX_RENDERER_PORT` (default 3100) |
| `src/env.ts` | Zod schema for all env vars; `loadEnv()` throws on invalid config with secret fields redacted in error |
| `src/service-auth.ts` | Fastify `onRequest` hook: validates `X-Service-Token` header via `timingSafeEqual`; exempts `/health` |
| `src/s3.ts` | `makeS3Client` — builds `minio.Client` from env; `getObjectBuffer`/`putObjectBuffer` helpers |
| `src/routes/index.ts` | Route registration entry: only registers fanout route |
| `src/routes/fanout.ts` | `POST /render/fanout` handler: validates body with Zod schema, fetches body DOCX from S3, calls `fanout()`, uploads frozen DOCX to `tenants/{tenantId}/revisions/{revisionId}/frozen.docx`, returns `{content_hash, final_docx_s3_key, unreplaced_vars, size_bytes}` |
| `src/render/fanout.ts` | `fanout()` function: builds `SubBlockRegistry`, renders header/footer sub-blocks via `compositionConfig`, merges with `placeholderValues` into `variables`, calls `processTemplateDetailed` from eigenpal, SHA-256s result, returns `{buffer, contentHash, unreplacedVars}` |
| `src/render/subblocks/registry.ts` | `SubBlockRegistry` class: `register`, `render`, `keys` — throws on unknown sub-block key |
| `src/render/subblocks/builtins.ts` | `registerV1Builtins` — registers 5 built-in sub-block renderers |
| `src/render/subblocks/doc_header_standard.ts` | `DocHeaderStandard` (`doc_header_standard`) — OOXML table with title, doc code, effective date, revision number |
| `src/render/subblocks/revision_box.ts` | `RevisionBox` (`revision_box`) — OOXML table from `revision_history` array in resolved values |
| `src/render/subblocks/approval_signatures_block.ts` | `ApprovalSignaturesBlock` (`approval_signatures_block`) — OOXML table with approver name + signed_at |
| `src/render/subblocks/footer_page_numbers.ts` | `FooterPageNumbers` (`footer_page_numbers`) — OOXML `PAGE`/`NUMPAGES` field elements |
| `src/render/subblocks/footer_controlled_copy_notice.ts` | `FooterControlledCopyNotice` (`footer_controlled_copy_notice`) — "CONTROLLED COPY — WHEN PRINTED" notice; overridable via `params.notice_text` |
| `build.mjs` | esbuild bundle script |
| `tsconfig.json` | TypeScript config |
| `vitest.config.ts` | Vitest test config |
| `package.json` | npm manifest; depends on vendored `@eigenpal/docx-js-editor@0.2.0` |
| `Dockerfile` | Multi-stage build: `node:20.11-alpine` builder → runtime; healthcheck on `/health`; exposes 3100; runs as `node` user |
| `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` | Vendored eigenpal tarball |
| `src/render/subblocks/__tests__/*.test.ts` | Unit tests for each sub-block renderer (7 files) |
| `src/render/__tests__/fanout.test.ts` | Unit tests for `fanout()` function |
| `src/routes/__tests__/fanout.test.ts` | Unit tests for the HTTP route handler |
| `test/fixtures.ts` | Shared test fixtures |
| `test/health.test.ts` | Health endpoint smoke test |
| `test/s3.smoke.test.ts` | S3 integration smoke test |

---

## 3. Public Surface

### Exported Go types/functions consumed elsewhere

**`internal/modules/render/resolvers`**
- `ComputedResolver` interface — consumed by `FreezeService` (`documents/application/freeze_service.go:139`)
- `Registry`, `NewRegistry`, `RegisterBuiltins` — wired in `main.go:384-385` and `apps/worker/cmd/metaldocs-worker/main.go:79-80`
- `ResolveInput` — constructed by `DocumentContextBuilder` (`documents/application/context_builder.go`), consumed by all resolvers
- `ResolvedValue` — returned by every resolver `Resolve` call
- `TenantID`, `RevisionID`, `ControlledDocumentID`, `ApprovalInstanceID` — strong-typed IDs used in `ResolveInput`
- `RevisionReader`, `WorkflowReader`, `RegistryReader`, `DocumentReader` interfaces — implemented in `documents/repository/resolver_readers.go`
- `AuthorInfo`, `ApproverInfo` — data shapes for resolver reader returns

**`internal/modules/render/fanout`**
- `Client`, `NewClient`, `FanoutRequest`, `FanoutResponse` — used by `FreezeService` and `ReconstructService`
- `FanoutClient` interface (in `fanout` package) — satisfied by `*Client`; also declared separately as `application.FanoutClient` in `freeze_service.go`
- `PDFDispatcher`, `NewPDFDispatcher` — wired in `main.go:392`
- `PDFDispatchAdapter`, `NewPDFDispatchAdapter` — wired in `main.go:393`
- `PDFOutboxRepository`, `NewPDFOutboxRepository` — wired in `main.go:455`, `worker/main.go:88`
- `PDFOutboxWorker`, `NewPDFOutboxWorker` — wired in `main.go:488`
- `MaterializeOutboxRepository`, `NewMaterializeOutboxRepository` — wired in `main.go:456`
- `MaterializeOutboxWorker`, `NewMaterializeOutboxWorker` — wired in `main.go:491`
- `ReconstructService`, `NewReconstructService`, `ReconstructionEntry`, `EngineVersions` — wired in `main.go:422-425`
- `FanoutInputsReader` interface — implemented by `documents/repository.FanoutInputsReader`

**`internal/platform/render/gotenberg`**
- `Client`, `NewClient` — constructed in `bootstrap/worker.go:83` and `bootstrap/api.go`
- `ConvertHTMLToPDF`, `ConvertDocxToPDF`, `ConvertDocxToPDFWithOptions` — called via `servicebus.GotenbergPDFClient`

**`internal/platform/docgenv2`**
- `TemplateReader`, `NewTemplateReader` — wired in `main.go:403`
- `TemplatesTemplateReader`, `NewTemplatesTemplateReader` — wired in `main.go:403`
- `FanoutTemplateReader`, `NewFanoutTemplateReader` — wired as `docDeps.TplRead` in `main.go:402-405`
- `TemplatesSnapshotReader`, `NewTemplatesSnapshotReader` — wired as `docSnapshotReader` in `main.go:397`

### HTTP routes exposed by `apps/docx-renderer`

| Method | Path | Auth | Purpose |
|---|---|---|---|
| `GET` | `/health` | None | Liveness probe; returns `{status:"ok", version}` |
| `POST` | `/render/fanout` | `X-Service-Token` (timing-safe compare) | DOCX token substitution + composition-block injection; reads body DOCX from MinIO, writes frozen DOCX to MinIO, returns `{content_hash, final_docx_s3_key, unreplaced_vars, size_bytes}` |

### HTTP routes exposed by Go API (render-related)

| Method | Path | AuthZ | Handler |
|---|---|---|---|
| `POST` | `/api/v1/documents/{id}/reconstruct` | `CapDocumentEdit` (area-grade) | `documentshttp.ReconstructHandler.HandleReconstruct` → `ReconstructionService.GetReconstruction` → `ReconstructService.Reconstruct` → `fanout.ReconstructService.Reconstruct` |

---

## 4. Logic Flows

### Flow 1: Approval-triggered async freeze (Pin → Materialize → PDF)

The primary production path. Triggered when an approval signoff completes.

1. **Approval decision** commits in `approvalapp.DecisionService` inside a transaction (`approval/application`).
2. **Pin** (`documents/application/freeze_service.go:191`): inside the same transaction, `FreezeService.Pin` is called.
   - Reads snapshot via `SnapshotReader.ReadSnapshotWithFreezeAt` (`freeze_service.go:192`).
   - If already pinned (`valuesFrozenAt != nil`), returns early — idempotent (`freeze_service.go:197-199`).
   - Calls `pinValidateAndHash` (`freeze_service.go:201`): validates required non-computed placeholders, calls each `ComputedResolver.Resolve` for computed placeholders (`freeze_service.go:144`), upserts resolved values via `FillInWriter.UpsertValue` (`freeze_service.go:150`), computes `values_hash` SHA-256 (`freeze_service.go:166-174`), writes freeze marker via `FreezeFinalizer.WriteFreeze` (`freeze_service.go:176`).
   - Enqueues `materialize_dispatch_outbox` row via `MaterializeOutboxEnqueuer.Enqueue` (`freeze_service.go:218`), with `ON CONFLICT (tenant_id, revision_id) DO NOTHING`.
3. **Materialize outbox worker** (`fanout/materialize_outbox_worker.go:53`): polls every 5s, claims pending rows with `FOR UPDATE SKIP LOCKED`, publishes `docx_materialize` events to messaging bus.
4. **Worker service** (`platform/worker/service.go:65`): `MaterializeJobRunner.Handle` is called.
   - Extracts `MaterializeFanoutPayload` (`worker/materialize_job_runner.go:59`).
   - Calls `MaterializeInvoker.Materialize` (implemented by `materializeInvokerAdapter.Materialize` → `FreezeService.Materialize`) outside a transaction (`materialize_job_runner.go:68`).
   - `FreezeService.Materialize` (`freeze_service.go:225`): reads snapshot, loads existing placeholder values, maps placeholder IDs to names, calls `fanout.Client.Fanout` (HTTP POST to docx-renderer) (`freeze_service.go:278`).
   - docx-renderer fetches body DOCX from MinIO, runs eigenpal `processTemplateDetailed`, uploads frozen DOCX to `tenants/{tenant}/revisions/{rev}/frozen.docx`, returns `{content_hash, final_docx_s3_key}`.
   - Back in `MaterializeJobRunner.Handle`, opens DB transaction, calls `WriteFinalDocxInTx` + `pdfOutbox.Enqueue` atomically (`materialize_job_runner.go:74-87`), commits.
5. **PDF outbox worker** (`fanout/pdf_outbox_worker.go:54`): polls every 5s, claims rows, publishes `docgen_v2_pdf` events.
6. **PDF job runner** (`platform/worker/pdf_job_runner.go:68`): calls `GotenbergPDFClient.ConvertPDF` (`servicebus/gotenberg_pdf.go:70`): opens frozen DOCX from MinIO, POSTs to Gotenberg LibreOffice route (`/forms/libreoffice/convert`), SHA-256s result, saves PDF to MinIO at `tenants/{tenant}/revisions/{rev}/final.pdf`, calls `PDFPersister.WritePDF` to persist `final_pdf_s3_key` + PDF hash.

### Flow 2: Synchronous Freeze (legacy path, kept for backward compatibility)

Defined at `freeze_service.go:302`, annotated "New code should use Pin + Materialize instead."

1. `FreezeService.Freeze` is called (within or without a transaction).
2. Calls `pinValidateAndHash` (same validation + resolution path as Pin).
3. Immediately calls `fanout.Client.Fanout` (HTTP to docx-renderer) — blocks the calling goroutine.
4. Calls `FinalDocxWriter.WriteFinalDocx` to persist `final_docx_s3_key` + `content_hash`.
5. No PDF outbox enqueue — the caller (approval path) must separately call `PDFDispatchAdapter.Dispatch` to enqueue the PDF job.

### Flow 3: PDF conversion (Gotenberg direct path)

1. `platform/worker/Service.RunOnce` receives a `docgen_v2_pdf` event from the outbox.
2. `PDFJobRunner.Handle` (`worker/pdf_job_runner.go:68`) extracts `PDFConvertPayload`.
3. Derives `docxKey` from payload (`FinalDocxS3Key` if set, else constructs path `tenants/{t}/revisions/{r}/frozen.docx`).
4. Calls `GotenbergPDFClient.ConvertPDF` (`servicebus/gotenberg_pdf.go:70`).
5. `GotenbergPDFClient` opens DOCX from MinIO via `pdfObjectStore.Open`, calls `gotenberg.Client.ConvertDocxToPDFWithOptions` (multipart POST to `/forms/libreoffice/convert`), saves PDF to `OutputKey`, returns `ContentHash`.
6. `PDFJobRunner` calls `PDFPersister.WritePDF` with `(tenantID, revisionID, outputKey, pdfHash, generatedAt)`.

### Flow 4: Forensic reconstruction

Triggered by `POST /api/v1/documents/{id}/reconstruct`.

1. `ReconstructionService.GetReconstruction` (`documents/application/reconstruct_service.go:26`) opens read-only transaction, seeds auth context, loads document's area code, requires `CapDocumentEdit` authz.
2. Calls `ReconstructionRunner.Reconstruct` → `fanout.ReconstructService.Reconstruct` (`fanout/reconstruction.go:62`).
3. `FanoutInputsReader.ReadForReconstruction` (`documents/repository/resolver_readers.go:112`) loads `body_docx_snapshot_s3_key`, `composition_config_snapshot`, `content_hash`, and stored placeholder values.
4. Calls `fanout.Client.Fanout` with stored inputs — re-runs docx-renderer.
5. Compares `resp.ContentHash` (hex) with `hex.EncodeToString(originalHash)` (`reconstruction.go:79`).
6. Marshals `ReconstructionEntry` and calls `ReconstructionWriter.AppendReconstruction` to write JSON blob to `documents.reconstruction_attempts`.
7. Returns `ReconstructionEntry` with `MatchesOriginal` field.

### Flow 5: Template reader fallback (docgenv2 dual-reader)

1. `FanoutTemplateReader.GetPublishedVersion` is called (`platform/docgenv2/templates_reader.go:53`).
2. Tries `TemplateReader.GetPublishedVersion` (queries `template_versions`/`templates` tables, legacy schema).
3. On `sql.ErrNoRows`, falls back to `TemplatesTemplateReader.GetPublishedVersion` (queries `templates_template_version`/`templates_template`, new schema).
4. Any other error from primary aborts without fallback.
5. Schema JSON is always `""` for the new templates module path; schema only present in legacy path (read from MinIO).

---

## 5. Dependencies

### Outbound (what this area imports)

**`internal/modules/render/resolvers`**
- `crypto/sha256`, `encoding/json` — inputs hashing
- No platform imports; pure domain logic only

**`internal/modules/render/fanout`**
- `metaldocs/internal/platform/messaging` — `Publisher`, `Event`, `PDFConvertPayload`, `MaterializeFanoutPayload`, `EventTypePDFConvert`, `EventTypeMaterializeFanout`
- `database/sql` — outbox repository queries
- `net/http` — fanout HTTP client
- `github.com/google/uuid` — event ID generation
- `log/slog` — structured logging in workers
- `metaldocs/internal/modules/documents/domain` — `v2dom` (reconstruction type constraints)
- `metaldocs/internal/modules/iam/authz` — `ReconstructionService` authz check

**`internal/platform/render/gotenberg`**
- `net/http`, `mime/multipart` — multipart HTTP to Gotenberg
- No internal MetalDocs imports

**`internal/platform/docgenv2`**
- `github.com/minio/minio-go/v7` — schema file read from MinIO
- `database/sql` — template/snapshot queries
- `metaldocs/internal/modules/documents/application` — `SnapshotTemplateReader` interface
- `metaldocs/internal/modules/documents/domain` — `TemplateSnapshot`, `ErrSnapshotTemplateNotFound`

**`apps/docx-renderer`**
- `@eigenpal/docx-js-editor` (vendored `0.2.0.tgz`) — `processTemplateDetailed` headless DOCX substitution
- `fastify@4.26.2` — HTTP server
- `minio@^7.1.3` — MinIO client
- `zod@3.23.8` — env + request body validation
- `jszip@3.10.1` — DOCX ZIP manipulation (transitive via eigenpal)
- `node:crypto` — SHA-256 content hash, `timingSafeEqual` auth comparison

### Inbound (who imports this area — verified with grep)

| Importer | What it imports |
|---|---|
| `apps/api/cmd/metaldocs-api/main.go` | `modules/render/fanout`, `modules/render/resolvers`, `platform/docgenv2`, `platform/render/gotenberg` (via bootstrap) |
| `apps/worker/cmd/metaldocs-worker/main.go` | `modules/render/fanout`, `modules/render/resolvers` |
| `internal/modules/documents/application/freeze_service.go` | `modules/render/fanout`, `modules/render/resolvers` |
| `internal/modules/documents/application/context_builder.go` | `modules/render/resolvers` |
| `internal/modules/documents/application/reconstruct_service.go` | `modules/render/fanout` |
| `internal/modules/documents/http/reconstruct_handler.go` | `modules/render/fanout` (for `ReconstructionEntry` type) |
| `internal/modules/documents/repository/resolver_readers.go` | `modules/render/fanout`, `modules/render/resolvers` |
| `internal/platform/bootstrap/worker.go` | `platform/render/gotenberg` |
| `internal/platform/bootstrap/api.go` | `platform/render/gotenberg` |
| `internal/platform/worker/pdf_pipeline_test.go` | `platform/render/gotenberg` (test) |

---

## 6. Persistence

### Tables written/read by this area

**`metaldocs.pdf_dispatch_outbox`** — managed entirely by `PDFOutboxRepository`
- Columns: `id` (uuid PK), `tenant_id`, `revision_id`, `content_hash`, `status` (pending/processing/dispatched/failed), `attempts`, `next_retry_at`, `claimed_at`, `dispatched_at`, `last_error`, `created_at`
- Unique constraint: `(tenant_id, revision_id)` — deduplication via `ON CONFLICT DO NOTHING`
- Claim pattern: `FOR UPDATE SKIP LOCKED` (`pdf_outbox_repository.go:48-57`)
- Stale-claim reset: rows stuck in `processing` > 5 minutes reset to `pending` (`pdf_outbox_repository.go:147`)
- Read by `ViewService` via `PDFOutboxStateReader.ReadState` to report `pdf_status=failed` to clients

**`metaldocs.materialize_dispatch_outbox`** — managed entirely by `MaterializeOutboxRepository`
- Same column shape and query patterns as `pdf_dispatch_outbox`
- `ON CONFLICT (tenant_id, revision_id) DO NOTHING`
- Unique constraint: `(tenant_id, revision_id)`

**`documents`** — read (not written) by this area
- `final_pdf_s3_key` — written by `PDFPersister.WritePDF` (lives in `documents/repository`)
- `final_docx_s3_key`, `content_hash` — written by `FinalDocxWriter.WriteFinalDocx` (lives in `documents/repository`)
- `body_docx_snapshot_s3_key`, `composition_config_snapshot`, `content_hash` — read by `FanoutInputsReader.ReadForReconstruction` (`resolver_readers.go:117`)
- `revision_number`, `effective_from`, `created_by`, `name`, `process_area_code_snapshot`, `controlled_document_id`, `area_name_snapshot` — read by `RevisionReader` and `DocumentContextBuilder` for resolver inputs
- `reconstruction_attempts` (JSONB column) — appended by `ReconstructionWriter.AppendReconstruction`

**`template_versions`, `templates`** (legacy schema) — read by `TemplateReader.GetPublishedVersion` (`docgenv2/template_reader.go:29`)

**`templates_template_version`, `templates_template`** (new schema) — read by `TemplatesTemplateReader.GetPublishedVersion` (`docgenv2/templates_reader.go:22`) and `TemplatesSnapshotReader.LoadForSnapshot` (`docgenv2/templates_snapshot_reader.go:27`)

**`approval_signoffs`, `approval_instances`** — read by `WorkflowReader.GetApprovers` and `GetFinalApprovalDate` (`resolver_readers.go:66-103`) for approval resolver inputs

### MinIO / object storage

- **Read:** `body_docx_s3_key` (template DOCX) — fetched by docx-renderer from `DOCX_RENDERER_S3_BUCKET` (`routes/fanout.ts:52-56`)
- **Written:** `tenants/{tenantId}/revisions/{revisionId}/frozen.docx` — frozen DOCX uploaded by docx-renderer (`routes/fanout.ts:65-70`)
- **Read:** frozen DOCX fetched by `GotenbergPDFClient.ConvertPDF` from MinIO via `pdfObjectStore.Open` (`servicebus/gotenberg_pdf.go:73`)
- **Written:** `tenants/{tenantId}/revisions/{revisionId}/final.pdf` — PDF stored by `GotenbergPDFClient.ConvertPDF` (`servicebus/gotenberg_pdf.go:99`)
- **Read:** schema JSON for legacy templates — fetched by `TemplateReader.GetPublishedVersion` from MinIO (`docgenv2/template_reader.go:46`)

---

## 7. Config & Environment

### Go side (API binary and worker)

| Variable | Where parsed | Default | Effect |
|---|---|---|---|
| `METALDOCS_FANOUT_URL` | `main.go:362`, `worker/main.go` (via bootstrap) | `""` | If empty: freeze/fanout disabled, `log.Fatal` if called in approval path (`main.go:363`). If set: `fanout.Client` constructed; freeze, materialize, reconstruct wired. |
| `METALDOCS_DOCX_RENDERER_SERVICE_TOKEN` | `main.go:366`, `bootstrap/worker.go:48` | `""` | Required when `METALDOCS_FANOUT_URL` is set; sent as `X-Service-Token` header to docx-renderer |
| `METALDOCS_GOTENBERG_URL` | `config/gotenberg.go:18` | `""` | If empty: Gotenberg disabled; PDF conversion won't run. If set: `GotenbergPDFClient` constructed. |
| `METALDOCS_ATTACHMENTS_PROVIDER` | `config.LoadAttachmentsConfig()` | — | Must be `minio` for Gotenberg path to activate (`bootstrap/worker.go:76`) |
| `METALDOCS_ATTACHMENTS_*` | MinIO config loader | — | Endpoint, access key, secret, bucket for object storage |

### TypeScript `docx-renderer` (parsed via Zod in `src/env.ts`)

| Variable | Default | Notes |
|---|---|---|
| `DOCX_RENDERER_PORT` | `3100` | HTTP listen port |
| `DOCX_RENDERER_SERVICE_TOKEN` | required (>=16 chars) | Pre-shared token for service-to-service auth |
| `DOCX_RENDERER_S3_ENDPOINT` | `http://minio:9000` | MinIO endpoint |
| `DOCX_RENDERER_S3_ACCESS_KEY` | required (>=3 chars) | MinIO access key |
| `DOCX_RENDERER_S3_SECRET_KEY` | required (>=3 chars) | MinIO secret key |
| `DOCX_RENDERER_S3_BUCKET` | `metaldocs-docx-v2` | S3 bucket for DOCX read/write |
| `DOCX_RENDERER_S3_USE_SSL` | `false` | SSL flag for MinIO client |
| `DOCX_RENDERER_GOTENBERG_URL` | `http://gotenberg:3000` | Gotenberg URL — present in env schema but **not consumed by any route handler**; unused in current implementation [runtime-unverified: confirm not used by a Gotenberg call inside docx-renderer] |
| `LOG_LEVEL` | `info` | Fastify logger level |
| `VERSION` | `dev` | Returned in `/health` |

---

## 8. Concurrency & Async

### Go side

- **`PDFOutboxWorker.Run`** (`pdf_outbox_worker.go:41`): launched as a goroutine in `main.go:488-489` inside `startOutboxWorker` wrapper. The wrapper runs an infinite restart loop with exponential backoff (1s → 1m) on unexpected exit. The worker itself uses `time.NewTicker(5s)` and calls `tick` (reset stale claims, claim batch, dispatch each row synchronously).
- **`MaterializeOutboxWorker.Run`** (`materialize_outbox_worker.go:40`): same goroutine pattern via `startOutboxWorker` wrapper (`main.go:491-492`). Publishes `docx_materialize` events.
- **`PDFOutboxRepository.ClaimPending`** (`pdf_outbox_repository.go:46`): uses `FOR UPDATE SKIP LOCKED` — safe for concurrent worker instances. Stale-claim reset (`ResetStaleClaims`) reclaims rows stuck in `processing` for > 5 minutes.
- **`MaterializeOutboxRepository.ClaimPending`**: identical pattern.
- **`Registry` (resolvers)** (`registry.go:5`): protected by `sync.RWMutex`; `Register` takes write lock, `Get`/`Known` take read lock. The registry is populated at startup (`RegisterBuiltins`) and never modified at runtime, so read contention is zero in practice.
- **`worker.Service.RunOnce`** (`worker/service.go:42`): processes events sequentially within a batch; no per-event goroutines. Invoked from the worker binary's ticker loop.
- No River jobs are directly related to the render pipeline; River is used for scheduled approval publish only.

### TypeScript side

- Fastify uses Node.js single-threaded async I/O with async/await.
- `fanout()` calls `Promise.all` for concurrent header and footer sub-block rendering (`render/fanout.ts:27-47`).
- `processTemplateDetailed` (eigenpal) is synchronous and blocks the event loop for the duration of DOCX processing [runtime-unverified: no async split observed].

---

## 9. Error Handling & Observability

### Error handling patterns

- **Go outbox workers**: errors from `ResetStaleClaims` and `ClaimPending` are logged with `slog.Warn` and the tick is skipped — no crash. Dispatch errors trigger `MarkFailed` with exponential backoff (base 30s, cap 30m); on `finalize=true` (attempt >= 5), row moves to terminal `failed` status and `slog.Error` is emitted.
- **`MaterializeJobRunner.Handle`**: errors before the transaction are returned and handled by `worker.Service.markFailure`, which schedules a retry or dead-letters the event using the worker's `MaxAttempts` config.
- **`fanout.Client.Fanout`** (`client.go:58`): non-200 responses return an error containing status code and body; no RFC 9457 parsing.
- **`GotenbergPDFClient.ConvertPDF`**: wraps all errors with `fmt.Errorf("%s: %w")`; body size capped at 64 MiB for PDF, 4 KiB for error responses.
- **`ApprovalDateResolver`**: errors if `approvalDate.IsZero()` (`approval_date.go:30`) — hard failure, not a blank fallback.
- **`FreezeService.Materialize`**: errors if snapshot not yet pinned (`freeze_service.go:235`).
- **`ReconstructService`** (`documents/application/reconstruct_service.go`): authz failure surfaces as `authz.ErrCapDenied` (403); not-found is `v2dom.ErrNotFound` (404); other errors logged at `slog.ErrorContext` level by the HTTP handler.
- **`TemplatesSnapshotReader`**: maps `sql.ErrNoRows` to `domain.ErrSnapshotTemplateNotFound`.

### RFC 9457 usage

- `writeReconstructError` in `documents/http/reconstruct_handler.go:52` delegates to `writeFillInError` which applies the standard RFC 9457 response pattern used across the documents module. However, `fanout.Client.Fanout` returns a raw `fmt.Errorf` with status code string — no structured error body from docx-renderer is propagated as a typed error.

### Logging / Observability

- Go workers use `log/slog` structured logging for batch and event outcomes (`worker/service.go:89-94`).
- `PDFOutboxWorker` / `MaterializeOutboxWorker` use `slog.Warn` and `slog.Error` with `id`, `revision_id`, `tenant_id` fields.
- `MaterializeJobRunner` uses `slog.InfoContext` on success (`materialize_job_runner.go:90`).
- `DocumentContextBuilder` uses `slog.WarnContext` on approval instance lookup failure (best-effort, non-fatal) (`context_builder.go:63`).
- No metrics or traces instrumentation observed in this area — all observability is log-based.
- docx-renderer: Fastify built-in logger at `LOG_LEVEL` env-configured level; no custom metrics.

---

## 10. Legacy / Duplication / Smell Flags

- **`pdf_outbox_worker.go` and `materialize_outbox_worker.go` are near-identical clones.** Both implement the same tick/claim/dispatch/backoff loop against two different outbox tables using the same algorithm and constants (5s poll, 10 batch, 5 max attempts, 5m stale threshold, same exponential backoff formula). WHERE: `internal/modules/render/fanout/pdf_outbox_worker.go` and `internal/modules/render/fanout/materialize_outbox_worker.go`. WHY: This is structural duplication that should be a single generic outbox worker parameterized on table and payload type; any bug fix or tuning change must be applied twice.

- **`pdf_outbox_repository.go` and `materialize_outbox_repository.go` are near-identical clones.** All six methods (`Enqueue`, `ClaimPending`, `MarkDispatched`, `MarkFailed`, `ReadState`, `ResetStaleClaims`) share identical logic differing only in table name (`pdf_dispatch_outbox` vs `materialize_dispatch_outbox`) and row type (`OutboxRow` vs `MaterializeOutboxRow`). WHERE: `internal/modules/render/fanout/pdf_outbox_repository.go` and `internal/modules/render/fanout/materialize_outbox_repository.go`. WHY: Classic copy-paste duplication; all future schema/behavior changes must be applied to both.

- **`FreezeService.Freeze` is a synchronous legacy path explicitly superseded by Pin+Materialize.** The comment at `freeze_service.go:300` reads "New code should use Pin (in-tx) + Materialize (async worker) instead" but `Freeze` remains exported and in use. WHERE: `internal/modules/documents/application/freeze_service.go:302`. WHY: The synchronous path still holds a `fanout` client call on the hot path and blocks the calling transaction; unclear whether any callsite still exercises this path.

- **`EngineVersions` hardcoded to `"local"` strings in both API and worker wiring.** WHERE: `apps/api/cmd/metaldocs-api/main.go:424` and `apps/worker/cmd/metaldocs-worker/main.go` (not visible in the read slice, but reconstruction wiring in the API binary). WHY: The forensic reconstruction record carries `eigenpal_ver="local"` and `docxtemplater_ver="local"` always, making the `MatchesOriginal` comparison meaningless as a cross-version forensics tool.

- **`GetFinalApprovalDate` on `WorkflowReader` interface is revision-scoped and cannot disambiguate multiple approval instances.** The comment is explicit at `resolver.go:70-72`. WHERE: `internal/modules/render/resolvers/resolver.go:70`. WHY: A revision with multiple approval instances (e.g. rework cycles) will return `MAX(signed_at)` across all of them (`resolver_readers.go:93-101`), which is not deterministic for the correct approval date.

- **`FanoutTemplateReader` has no interface boundary for testing.** The `TODO` at `docgenv2/templates_reader.go:42` flags that the concrete types `*TemplateReader` and `*TemplatesTemplateReader` are directly embedded in the struct, making unit testing impossible without real DB/S3 dependencies. WHERE: `internal/platform/docgenv2/templates_reader.go:44-47`.

- **`pdf_dispatch_outbox` claim query is cross-tenant.** The `TODO` at `pdf_outbox_repository.go:43` states "thread tenant scope through the worker claim path before adding a tenant_id predicate here; the current worker intentionally drains the shared outbox across tenants." WHERE: `internal/modules/render/fanout/pdf_outbox_repository.go:43`. WHY: This is intentional today but documented as a known future multi-tenant concern.

- **`DOCX_RENDERER_GOTENBERG_URL` env var is declared but not consumed.** The env schema in `apps/docx-renderer/src/env.ts:13` declares `DOCX_RENDERER_GOTENBERG_URL` pointing to `http://gotenberg:3000`, but no route handler in `src/routes/` reads this value. Gotenberg conversion from the docx-renderer side was removed and the Go worker now calls Gotenberg directly. The env var declaration is dead configuration. WHERE: `apps/docx-renderer/src/env.ts:13`.

- **`approvers.go` contains a hardcoded Portuguese fallback string.** `"[aguardando aprovação]"` (`approvers.go:34`) is a locale-specific string baked into the resolver with no i18n mechanism. WHERE: `internal/modules/render/resolvers/approvers.go:34`. WHY: Internationalization is not currently a stated requirement, but it is a smell for a system that may need to support multiple languages.

- **`controlled_by_area` resolver is version 2 while all other built-in resolvers are version 1.** No resolver versioning/migration policy exists in code (flagged as T-003 in tech debt). WHERE: `internal/modules/render/resolvers/controlled_by_area.go:14`. WHY: The registry silently stores last-registered version per key (`registry.go:19`); if two resolvers with the same key but different versions are registered, the second silently wins.

- **`TemplateReader` uses a sentinel UUID tenant ID `"ffffffff-ffff-ffff-ffff-ffffffffffff"` for system templates** baked as a package-level constant. WHERE: `internal/platform/docgenv2/template_reader.go:13`. WHY: Magic UUID with no named domain concept; not a separate flag but worth noting as a pattern inconsistency with the rest of the tenant model.

- **Naming drift: package is called `docgenv2` but the service it references is `docx-renderer`.** The rename commit (`7996a460c`) updated the service name but left the Go package name unchanged. WHERE: `internal/platform/docgenv2/`. WHY: The package name leaks the v2 naming era; the `FanoutTemplateReader` concept in this package refers to template reading, not fanout dispatch — the name is misleading.

---

## 11. Wiki Drift

**`wiki/modules/render-fanout.md` (Last verified: 2026-06-01)**

1. **Key files list duplicates `apps/docx-renderer/src/routes/fanout.ts`** — appears twice in the Key files section (lines 10 and 17). Minor but indicates the list was copy-edited without a dedup pass.

2. **`internal/platform/worker/pdf_job_runner.go` is listed as "outbox consumer"** but the actual flow is: the outbox workers (`PDFOutboxWorker`, `MaterializeOutboxWorker`) publish events to the messaging bus; `pdf_job_runner.go` is a _messaging consumer_ (handles events from the outbox messaging layer), not a direct outbox consumer. The distinction matters for understanding the two-level dispatch architecture.

3. **Pipeline step 3 says "PDFDispatcher publishes a `docgen_v2_pdf` outbox event"** but in the async Pin+Materialize path (ADR 0015), step 3 is actually the `materialize_dispatch_outbox` enqueue (not a PDF event), and the PDF outbox is enqueued by `MaterializeJobRunner` after the fanout call. The wiki pipeline description mixes the legacy Freeze path with the current Pin+Materialize path.

4. **The wiki does not document `MaterializeOutboxWorker` or `materialize_dispatch_outbox`** — the two-phase async split (ADR 0015) is not reflected in the pipeline steps or key files list. `materialize_outbox_repository.go` and `materialize_outbox_worker.go` are entirely absent from the wiki.

5. **`wiki/modules/render-fanout-tech-debt.md` T-001 states "module page is currently a high-level stub"** — still accurate as of this audit; the module page has no flow diagrams, no route truth table for the HTTP contract, and no complete persistence map.

---

## 12. Open Questions

- **[runtime-unverified]** `DOCX_RENDERER_GOTENBERG_URL` is declared in `apps/docx-renderer/src/env.ts:13` but not consumed by any route handler in `src/routes/`. Confirm this env var is truly dead (no planned future use) and whether it should be removed to avoid operator confusion.

- **[runtime-unverified]** `FreezeService.Freeze` (legacy synchronous path at `freeze_service.go:302`): confirm whether any live callsite still invokes this method in production, or if it can be removed.

- **[runtime-unverified]** `TemplatesSnapshotReader.LoadForSnapshot` always returns `CompositionJSON: []byte("{}")` (`templates_snapshot_reader.go:49`). Confirm whether composition config is intentionally not stored for the templates module, or whether this is a missing feature.

- **[runtime-unverified]** The `approvers` resolver returns a raw `*sql.Rows` nil list when `approvalInstanceID == ""` (`resolver_readers.go:67-69`). This means the `approvers` token on a document without an approval instance resolves to `"[aguardando aprovação]"` silently. Confirm whether this is the intended behavior for pre-approval documents.

- **[runtime-unverified]** `MaterializeJobRunner` calls `FreezeService.Materialize` which calls the docx-renderer HTTP endpoint outside a transaction (by design). If the worker process is killed between the HTTP call returning and the DB transaction committing, the frozen DOCX exists in MinIO but `final_docx_s3_key` is not persisted. Confirm whether the materialize outbox row is eventually retried (it stays `processing` until `ResetStaleClaims` resets it after 5m) and whether the docx-renderer `PUT` to MinIO at the deterministic key `tenants/{t}/revisions/{r}/frozen.docx` makes the re-run idempotent.

- **[runtime-unverified]** `PDFOutboxWorker.dispatchOne` (`pdf_outbox_worker.go:79`) includes `ContentHash` in the `PDFConvertPayload` but `PDFJobRunner.Handle` (`pdf_job_runner.go:77`) does not use it — it only uses `FinalDocxS3Key` to find the DOCX. Confirm whether the `ContentHash` field in the PDF payload is vestigial or reserved for future validation.
