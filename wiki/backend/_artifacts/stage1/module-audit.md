# Stage-1 Audit Artifact — module-audit

> Produced: 2026-06-10. Read-only code scan. All claims carry file:line anchors. `[runtime-unverified]` marks claims that require a live Postgres instance to confirm.

---

## 1. Identity & purpose

`internal/modules/audit` is the regulated-action **append-only event sink** for MetalDocs. It exposes a write port (`domain.Writer`) for consumer modules to record regulated mutations (IAM admin operations, document lifecycle events, auth events, authz bypass events) and a read port (`domain.Reader`) plus an HTTP delivery layer for querying and exporting those events. Storage is two Postgres tables: `metaldocs.audit_events` (the immutable log) and `metaldocs.audit_export_jobs` (export job metadata and inline payload). The module deliberately owns no aggregate and no state machine; it is a side-effect sink whose callers have already enforced capability. Tamper-resistance is achieved through a SHA-256 row-hash chain (migration 0193) enforced at INSERT time via a Postgres advisory lock and validated by a scheduled job.

---

## 2. File inventory

### `internal/modules/audit/domain/`

| File | Role |
|---|---|
| `domain/port.go` | All domain types (`Event`, `ListEventsQuery`, `Cursor`, `IntegrityIssue`, `ExportJob`) and every interface (`Writer`, `Reader`, `Counter`, `IntegrityValidator`, `ExportJobRepository`); sentinel errors `ErrInvalidEvent`, `ErrExportJobNotFound`; `NewEvent` constructor. |

### `internal/modules/audit/application/`

| File | Role |
|---|---|
| `application/service.go` | `Service` struct; `NewService`; `WithExports` wiring method; `ListEvents`, `ExportEvents`, `GetExportStatus`, `LoadExportPayload`, `BuildSignedURL` orchestration; internal helpers `normalizeQuery`, `fetchAll`, `filterPayload`, `filterSummary`, `randomToken`. |
| `application/render.go` | `render`, `renderCSV`, `renderJSONL`, `canonicalPayload` — in-process serialization of `[]domain.Event` to CSV (with UTF-8 BOM) or JSONL bytes. |
| `application/service_test.go` | Unit tests for `NewService` panic guard, `ListEvents` normalization and tenant requirement, `GetExportStatus` actor requirement. Uses `captureReader` and `fakeExportRepo` stubs (no Postgres). |

### `internal/modules/audit/delivery/http/`

| File | Role |
|---|---|
| `delivery/http/handler.go` | `Handler` struct; `AuditQuerier` and `AuditExporter` interface narrowings; `NewHandler`, `WithExporter`, `RegisterRoutes`; route handlers `handleEvents`, `handleExport`, `handleExportSubresource`, `handleExportDownload`; helpers `parseListQuery`, `parseTime`, `encodeCursor`, `decodeCursor`, `buildEventResponses`, `auditTenantFromRequest`, `writeProblem`. |
| `delivery/http/handler_test.go` | Handler integration tests (unauthenticated rejection, tenant isolation, cursor pagination shape, limit clamp, invalid stored payload). Uses `memory.Writer` + `application.Service`. |
| `delivery/http/handler_export_test.go` | Export-path integration tests (multi-axis filter, cursor pagination with memory writer, CSV tenant-scoped download, actor scoping on status query, bad-format rejection). |

### `internal/modules/audit/infrastructure/postgres/`

| File | Role |
|---|---|
| `infrastructure/postgres/writer.go` | `Writer` struct; `NewWriter`; `Record` (opens own tx, calls `RecordTx`); `RecordTx` (acquires `pg_advisory_xact_lock`, runs CTE INSERT with hash-chain computation); `ValidateIntegrity` (validates last 10,000 rows of chain); `ListEvents` (parameterised SELECT with limit+1 probe); `CountEvents`; `buildListQuery`, `buildCountQuery`, `buildWhere` query builders. |
| `infrastructure/postgres/exports.go` | `ExportJobRepository`; `NewExportJobRepository`; `Save` (INSERT), `Get` (by tenant+id), `GetByDownloadToken` (by id+token); `scanJob` shared scanner. |
| `infrastructure/postgres/writer_test.go` | Sqlmock tests for `RecordTx` hash-chain INSERT, `ListEvents` exact-page/trim-probe regression guards (B1), `ValidateIntegrity` broken-chain detection and issue-limit cap. |

### `internal/modules/audit/infrastructure/memory/`

| File | Role |
|---|---|
| `infrastructure/memory/writer.go` | In-process `Writer`; `NewWriter`; `Record`, `RecordTx` (delegates to `Record`, ignores tx); `ListEvents` (mutex snapshot, sort, cursor-aware limit+1 probe); `CountEvents`; `matches` filter predicate; `ValidateIntegrity` (returns error — not supported). |
| `infrastructure/memory/exports.go` | In-process `ExportJobRepository`; `NewExportJobRepository`; `Save`, `Get`, `GetByDownloadToken`. |

### Marker files (no logic)

`application/.gitkeep`, `delivery/http/.gitkeep`, `domain/.gitkeep`, `infrastructure/.gitkeep` — layout markers only.

---

## 3. Public surface

### Exported types consumed elsewhere

| Package | Symbol | Kind | Consumers (verified by grep) |
|---|---|---|---|
| `domain` | `Event` | struct | All 7 non-test inbound importers |
| `domain` | `ListEventsQuery` | struct | All querying consumers |
| `domain` | `Cursor` | struct | `handler.go`, `service.go`, `writer.go` |
| `domain` | `Writer` | interface | `auth`, `iam`, `taxonomy`, `templates`, `controlleddocuments`, `bootstrap`, `main.go` adapters |
| `domain` | `Reader` | interface | `bootstrap`, `iam/admin_handler` |
| `domain` | `Counter` | interface | `bootstrap` |
| `domain` | `IntegrityValidator` | interface | `bootstrap`, `audit_integrity_validator/job.go` |
| `domain` | `ExportJobRepository` | interface | `bootstrap` |
| `domain` | `ExportJob` | struct | `application.Service`, `handler.go`, `bootstrap` |
| `domain` | `ExportFormat` / `ExportStatus` | string types | `application`, `handler.go`, `exports.go` |
| `domain` | `IntegrityIssue` / `IntegrityIssueKind` | struct / string type | `postgres/writer.go`, `jobs/audit_integrity_validator` |
| `domain` | `ErrInvalidEvent`, `ErrExportJobNotFound` | sentinel errors | `application`, `handler.go` |
| `domain` | `NewEvent` | constructor | `application/service.go:181` |
| `application` | `Service` | struct | `main.go:204-214`, `handler_test.go`, `handler_export_test.go` |
| `application` | `NewService` | constructor | `main.go:204` |
| `application` | `SignedURLBuilder` | function type | `main.go:206` |
| `application` | `ErrTenantRequired`, `ErrActorRequired`, `ErrInvalidFormat`, `ErrExportTooLarge`, `ErrExportRepoMissing`, `ErrCounterMissing`, `ErrExportTokenMismatch` | sentinel errors | `handler.go` |
| `application` | `SyncExportRowLimit`, `ExportTTL` | constants | `handler.go`, `service.go` |
| `delivery/http` | `Handler` | struct | `main.go:214,283` |
| `delivery/http` | `NewHandler`, `Handler.WithExporter`, `Handler.RegisterRoutes` | constructor + methods | `main.go:214,283` |
| `delivery/http` | `EventResponse` | struct | `handler_test.go` |
| `infrastructure/postgres` | `Writer` | struct | `bootstrap/api.go:105` |
| `infrastructure/postgres` | `NewWriter` | constructor | `bootstrap/api.go:105` |
| `infrastructure/postgres` | `ExportJobRepository` | struct | `bootstrap/api.go:106` |
| `infrastructure/postgres` | `NewExportJobRepository` | constructor | `bootstrap/api.go:106` |
| `infrastructure/memory` | `Writer` | struct | `bootstrap/api.go:139`, tests |
| `infrastructure/memory` | `NewWriter` | constructor | `bootstrap/api.go:139`, tests |
| `infrastructure/memory` | `ExportJobRepository` | struct | `bootstrap/api.go:140`, tests |
| `infrastructure/memory` | `NewExportJobRepository` | constructor | `bootstrap/api.go:140`, tests |

### HTTP routes

| Method | Path | Handler | Authz binding | Visibility |
|---|---|---|---|---|
| `GET` | `/api/v1/audit/events` | `Handler.handleEvents` (`handler.go:75`) | `CapAuditRead` (`permissions.go:232`) | `VisibilityPermissionGuarded` |
| `POST` | `/api/v1/audit/events/export` | `Handler.handleExport` (`handler.go:132`) | `CapAuditRead` (`permissions.go:233`) | `VisibilityPermissionGuarded` |
| `GET` | `/api/v1/audit/events/export/{id}` | `Handler.handleExportSubresource` → status branch (`handler.go:223`) | `CapAuditRead` (`permissions.go:234`) | `VisibilityPermissionGuarded` |
| `GET` | `/api/v1/audit/events/export/{id}/download` | `Handler.handleExportSubresource` → download branch → `handleExportDownload` (`handler.go:239`) | `CapAuditRead` (`permissions.go:234`) | `VisibilityPermissionGuarded` (tier-1); download token in query param (application-layer token gate) |

Route registration: `handler.go:69-73` mounts all four logical routes via three `mux.HandleFunc` calls; sub-resource routing is done inline by path parsing at `handler.go:232-246`.

---

## 4. Logic flows

### Flow 1: Write — Record a regulated event (fire-and-forget path)

1. Consumer module (e.g., `iam/delivery/http/admin_handler.go:403`) builds an `auditdomain.Event` with a UUID id, `time.Now().UTC()`, actor, action, resource, payload JSON, trace ID, and tenant ID.
2. Consumer calls `auditdomain.Writer.Record(ctx, event)`. The error return is typically discarded (`_ = h.audit.Record(...)` at `admin_handler.go:403`) or logged only (`log.Printf` at `main.go:827`).
3. `postgres.Writer.Record` (`postgres/writer.go:27`) opens a new `*sql.Tx` via `db.BeginTx`.
4. `postgres.Writer.RecordTx` (`postgres/writer.go:44`) acquires a transaction-scoped advisory lock (`pg_advisory_xact_lock(90120260513004)` at `postgres/writer.go:45`) to serialize the hash-chain writes.
5. A CTE INSERT (`postgres/writer.go:49-66`) reads the current chain tail (`SELECT row_hash FROM metaldocs.audit_events ORDER BY audit_sequence DESC LIMIT 1`), sets `prev_hash` from it, and calls `metaldocs.audit_event_row_hash(...)` (defined in migration 0193) to compute the new `row_hash` in-database. The event row is inserted atomically.
6. The transaction commits at `postgres/writer.go:37`.
7. If the INSERT fails (PK collision, constraint violation, connection loss), the error bubbles up to the consumer which ignores it — the regulated action is already committed and the audit row is silently lost.

### Flow 2: Read — `GET /api/v1/audit/events`

1. Request arrives at `Handler.handleEvents` (`handler.go:75`) after passing the tier-1 capability check in the middleware stack (verifies `CapAuditRead` via `permissions.go:232`).
2. Handler confirms no path conflict with export routes (`handler.go:76-79`), checks method (`handler.go:80-83`).
3. `auditTenantFromRequest` (`handler.go:423`) extracts authenticated user ID (`authn.UserIDFromContext`) and tenant ID (`tenant.FromContext`); returns 401/403 on missing context values.
4. `parseListQuery` (`handler.go:316`) reads query params: `limit` (default 50, clamped to MaxLimit=100 via `pagination.ClampLimit`), `occurred_after`, `occurred_before` (RFC3339), `cursor` (base64 opaque), `actor_id`, `action`, `resource_type`, `resource_id`, `q`.
5. `Service.ListEvents` (`service.go:65`) calls `normalizeQuery` (trims whitespace, enforces tenant, applies default/cap on Limit at `service.go:94-99`), then delegates to `reader.ListEvents`.
6. `postgres.Writer.ListEvents` (`postgres/writer.go:142`) builds a parameterised SELECT via `buildListQuery` with a `LIMIT n+1` probe row. Returns `items[:limit], true` when the probe row is present, or `items, false` otherwise (AIP-158 compliant).
7. Handler decodes payloads, builds `EventResponse` slice, and emits `{"items":[...], "page":{"next_cursor":..., "has_more":...}}` at `handler.go:126-129`. The cursor encodes `(occurred_at RFC3339Nano, id)` via `pagination.EncodeCursor`.

### Flow 3: Export — `POST /api/v1/audit/events/export` (synchronous inline path)

1. Request reaches `Handler.handleExport` (`handler.go:132`) with `CapAuditRead` gate. JSON body parsed for `format` and `filter` fields (`handler.go:147-183`).
2. `Service.ExportEvents` (`service.go:105`) validates actor, format, and calls `counter.CountEvents` (same `postgres.Writer`) with cursor/limit zeroed to count all matching rows.
3. If `estimatedRows > SyncExportRowLimit (50,000)` (`service.go:134`), returns `ErrExportTooLarge`; handler maps to HTTP 501 (`handler.go:195`).
4. `fetchAll` (`service.go:239`) pages through all matching events using the reader with `pageSize=200`, accumulating into a slice.
5. `render` (`render.go:17`) serializes to CSV (with UTF-8 BOM, `render.go:28`) or JSONL (`render.go:57`). CSV columns: `occurred_at, actor_id, action, resource_type, resource_id, trace_id, payload_json`.
6. A `domain.ExportJob` is constructed with status=`ready`, a 24-byte random hex `DownloadToken` (`service.go:161`), `ExpiresAt = now + 15 min` (`service.go:163`), and the `Payload` bytes.
7. `exportRepo.Save` persists the job row (including the inline `BYTEA` payload) to `metaldocs.audit_export_jobs` (`exports.go:26`).
8. An audit-of-audit `domain.NewEvent` is emitted via `writer.Record` recording `audit.export.requested` (`service.go:181`). Error discarded.
9. Handler responds HTTP 202 with `{"export_id", "status", "signed_url", "expires_at"}`. The `signed_url` is `"/api/v1/audit/events/export/{id}/download?token={token}"` built by the `SignedURLBuilder` closure at `main.go:206-212`.

### Flow 4: Download — `GET /api/v1/audit/events/export/{id}/download?token=...`

1. `handleExportDownload` (`handler.go:281`) reads the `token` query parameter (no auth context required — the token is the credential).
2. `exporter.LoadExportPayload` (`service.go:217`) calls `exportRepo.GetByDownloadToken(ctx, exportID, token)`. On no-match or token mismatch, returns `ErrExportJobNotFound` → handler returns 404.
3. Expiry check: if `job.ExpiresAt` is non-zero and `now.After(job.ExpiresAt)`, returns `ErrExportJobNotFound` → 404 (`service.go:225`).
4. Handler sets `Content-Type` (`text/csv; charset=utf-8` or `application/x-ndjson`), `Content-Disposition: attachment; filename=...`, `Cache-Control: no-store`, and streams `job.Payload` bytes (`handler.go:307-313`).

### Flow 5: Integrity validation job

1. `main.go:544-549` conditionally registers the `audit_integrity_validator` scheduler job if `ENABLE_JOB_AUDIT_INTEGRITY_VALIDATOR != "false"` and `deps.AuditValidator != nil`.
2. On each scheduler tick, `job.go:18` (in `internal/modules/jobs/audit_integrity_validator/`) calls `validator.ValidateIntegrity(ctx)`.
3. `postgres.Writer.ValidateIntegrity` (`postgres/writer.go:77`) runs a window query on the last 10,000 rows, computing expected `prev_hash` and `row_hash` via `metaldocs.audit_event_row_hash(...)` and comparing to stored values.
4. Issues are collected up to `auditIntegrityIssueLimit=256`; any mismatch is returned as an `IntegrityIssue`.
5. The job logs each violation with `slog.ErrorContext` and returns `ErrIntegrityViolation` to the scheduler (`job.go:31-34`). No alert/metric is emitted beyond the log line.

---

## 5. Dependencies

### Outbound (packages the audit module imports)

| Package | Where imported | Why |
|---|---|---|
| `metaldocs/internal/modules/audit/domain` | `application/`, `delivery/http/`, both `infrastructure/` pkgs | Own domain types and interfaces |
| `metaldocs/internal/platform/authn` | `delivery/http/handler.go:17` | Extract authenticated user ID from context |
| `metaldocs/internal/platform/httpresponse` | `delivery/http/handler.go:18` | `WriteJSON` for JSON response encoding |
| `metaldocs/internal/platform/pagination` | `delivery/http/handler.go:19`, `infrastructure/postgres/writer.go:11` | `ClampLimit`, `DefaultLimit`, `EncodeCursor`, `DecodeCursor` |
| `metaldocs/internal/platform/problem` | `delivery/http/handler.go:20` | RFC 9457 problem+json emission |
| `metaldocs/internal/platform/tenant` | `delivery/http/handler.go:21` | `tenant.FromContext` for tenant ID extraction |
| `metaldocs/internal/platform/sqlescape` | `infrastructure/postgres/writer.go:12` | `LikeEscape` for LIKE query sanitisation |
| `metaldocs/internal/modules/jobs/scheduler` | `jobs/audit_integrity_validator/job.go:10` | `scheduler.JobFunc` type for the validator job |
| `github.com/google/uuid` | `application/service.go:13` | UUID generation for export job IDs |
| `database/sql` | `domain/port.go:5`, `infrastructure/postgres/writer.go`, `infrastructure/memory/writer.go` | `Writer.RecordTx` accepts `*sql.Tx` |
| Standard library (`context`, `crypto/rand`, `encoding/csv`, `encoding/hex`, `encoding/json`, `errors`, `fmt`, `io`, `log/slog`, `net/http`, `sort`, `strconv`, `strings`, `sync`, `time`) | Multiple files | General purpose |

The audit module has **no outbound imports of other MetalDocs domain or application modules**. It is a leaf in the import graph.

### Inbound (who imports audit — verified by grep)

| Importer | Package path | What they import | Call pattern |
|---|---|---|---|
| `apps/api/cmd/metaldocs-api/main.go` | `auditdomain`, `auditapp`, `auditdelivery` | Full wiring: `NewService`, `WithExports`, `NewHandler`, `WithExporter`, `RegisterRoutes`; `bypassAuditAdapter`, `documentsAuditAdapter` | Bootstrap + adapter construction |
| `internal/platform/bootstrap/api.go` | `auditdomain`, `auditmemory`, `auditpg` | `NewWriter`, `NewExportJobRepository`, fills `Deps` struct | Platform-layer dependency assembly |
| `internal/modules/iam/delivery/http/admin_handler.go` | `auditdomain` | `Writer.Record` (user/role mutations), `AuditEventLister` interface (recent events panel) | `Record` fire-and-forget; `ListEvents` in overview |
| `internal/modules/iam/delivery/http/people_handler.go` | `auditdomain` | `Writer.Record` | User management mutations |
| `internal/modules/iam/delivery/http/routes_memberships.go` | `auditdomain` | `Writer.Record` | Membership mutations |
| `internal/modules/iam/delivery/http/sessions_handler.go` | `auditdomain` | `Writer.Record` | Session force-logout |
| `internal/modules/auth/delivery/http/handler.go` | `auditdomain` | `Writer.Record` for login, logout, password-change events via `recordAudit` helper | `Record` fire-and-forget (logged on error) |
| `internal/modules/taxonomy/application/audit_governance_adapter.go` | `auditdomain` | `AuditGovernanceAdapter` wraps `Writer.Record` to satisfy `domain.GovernanceLogger` | Governance event routing to audit sink |
| `internal/modules/taxonomy/module.go` | `auditdomain` | Wires `AuditGovernanceAdapter` | Module initialization |
| `internal/modules/controlleddocuments/module.go` | `auditdomain` | Passes `Writer` to taxonomy for governance logging | Module initialization |
| `internal/modules/templates/repository/postgres.go` | `auditdomain` | `Writer.Record` and `Writer.RecordTx` for template lifecycle events | Repository-layer audit write |
| `internal/modules/jobs/audit_integrity_validator/job.go` | `auditdomain` | `IntegrityValidator.ValidateIntegrity` | Scheduled integrity check |

---

## 6. Persistence

### Tables

**`metaldocs.audit_events`** (primary log)

Created in archived `migrations/0004_init_audit_events.sql`. Extended by archived `0190_audit_events_tenant_id.sql` (adds `tenant_id TEXT NOT NULL DEFAULT ''`) and archived `0193_audit_events_hash_chain.sql` (adds `audit_sequence BIGSERIAL`, `prev_hash TEXT NOT NULL DEFAULT ''`, `row_hash TEXT NOT NULL DEFAULT ''`). Current canonical shape is in `db/baseline/0001_current_schema.sql:875`.

| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | Consumer-generated UUID (or timestamp-string in older consumers) |
| `occurred_at` | TIMESTAMPTZ NOT NULL | Set to `time.Now().UTC()` by consumer |
| `actor_id` | TEXT NOT NULL | User/system identifier |
| `action` | TEXT NOT NULL | Free-text event label (e.g. `auth.login`, `document.renamed`) |
| `resource_type` | TEXT NOT NULL | Domain entity type |
| `resource_id` | TEXT NOT NULL | Entity primary key |
| `payload` | JSONB NOT NULL DEFAULT '{}' | Unstructured event context; no size constraint |
| `trace_id` | TEXT NOT NULL | HTTP trace correlation |
| `tenant_id` | TEXT NOT NULL DEFAULT '' | Added in 0190; used in all WHERE clauses |
| `audit_sequence` | BIGSERIAL | Added in 0193; unique; ORDER BY anchor for hash chain |
| `prev_hash` | TEXT NOT NULL DEFAULT '' | SHA-256 hex of previous row's row_hash; empty for first row |
| `row_hash` | TEXT NOT NULL DEFAULT '' | SHA-256 hash of all 10 fields; computed by `metaldocs.audit_event_row_hash()` |

Indexes (from baseline): `idx_audit_events_occurred_at (occurred_at DESC)`, `idx_audit_events_actor_time (actor_id, occurred_at DESC)`, `idx_audit_events_resource_time (resource_type, resource_id, occurred_at DESC)`, `idx_audit_events_tenant_id (tenant_id)`, `idx_audit_events_audit_sequence (audit_sequence) UNIQUE`.

Constraints: `audit_events_prev_hash_shape CHECK (prev_hash = '' OR prev_hash ~ '^[0-9a-f]{64}$') NOT VALID`, `audit_events_row_hash_shape CHECK (row_hash = '' OR row_hash ~ '^[0-9a-f]{64}$') NOT VALID`.

Grants: `GRANT INSERT, SELECT ON TABLE metaldocs.audit_events TO metaldocs_app` (added in archived 0193). The baseline schema carries no explicit GRANT blocks — the grant lives in the archived migration ledger.

**`metaldocs.audit_export_jobs`** (export job metadata + inline payload)

Created in `db/migrations/0224_audit_export_jobs_pr6.sql`. Not yet in the curated baseline (migration is a forward-only delta). [runtime-unverified: whether this table exists in the current deployed database — it is a recent migration not reflected in the baseline file].

| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | UUID from `application.Service.ExportEvents` |
| `tenant_id` | UUID NOT NULL | Scoped; compared with TEXT on read (`postgres/exports.go:53`) |
| `actor_id` | TEXT NOT NULL | Owner; `GetExportStatus` enforces `job.ActorID == actorID` |
| `format` | TEXT NOT NULL CHECK (format IN ('csv','jsonl')) | |
| `filter_json` | JSONB NOT NULL | Serialised filter for record-keeping |
| `status` | TEXT NOT NULL DEFAULT 'pending' CHECK (...) | `pending`, `running`, `ready`, `failed` |
| `object_key` | TEXT (nullable) | Reserved for future async/S3 path |
| `download_token` | TEXT (nullable) | 24-byte random hex; used as URL credential |
| `expires_at` | TIMESTAMPTZ (nullable) | 15 minutes from creation; checked by `LoadExportPayload` |
| `error_message` | TEXT (nullable) | |
| `estimated_rows` | BIGINT NOT NULL DEFAULT 0 | Pre-render count from `CountEvents` |
| `actual_rows` | BIGINT NOT NULL DEFAULT 0 | Post-render actual count |
| `payload` | BYTEA (nullable) | Inline rendered CSV/JSONL blob |
| `created_at` | TIMESTAMPTZ NOT NULL DEFAULT now() | |
| `completed_at` | TIMESTAMPTZ (nullable) | |

Indexes: `ix_audit_export_jobs_tenant_actor (tenant_id, actor_id, created_at DESC)`, `ix_audit_export_jobs_download_token (download_token) WHERE download_token IS NOT NULL`.

No GRANT statement in migration 0224 — access to this table is [runtime-unverified: relies on existing role or schema-level grant].

### Query patterns

- `RecordTx`: CTE INSERT (`postgres/writer.go:49-74`) with `pg_advisory_xact_lock` serialisation.
- `ListEvents`: parameterised SELECT with `LIMIT n+1` probe and keyset cursor `(occurred_at, id) < ($n, $n+1)` (`postgres/writer.go:194-212`).
- `CountEvents`: `SELECT COUNT(*) FROM metaldocs.audit_events WHERE ...` (`postgres/writer.go:215`).
- `ValidateIntegrity`: window query over the last 10,000 rows with `ROW_NUMBER()` and `LAG()` (`postgres/writer.go:78-100`).
- `ExportJobRepository.Save`: plain INSERT (`exports.go:27-49`).
- `ExportJobRepository.Get` / `GetByDownloadToken`: single-row SELECT (`exports.go:52-59`).
- Retention purge: `DELETE FROM metaldocs.audit_events WHERE occurred_at < $1` executed by a goroutine in `main.go:585-587` every 24 hours when `AUDIT_RETENTION_DAYS > 0`.

### Migration files (audit-relevant, in order)

| Migration | Location | What it does |
|---|---|---|
| `0004_init_audit_events.sql` | `archive/migrations/` | Creates `metaldocs.audit_events` (original 8 columns), 3 btree indexes |
| `0005_grant_workflow_audit_privileges.sql` | `archive/migrations/` | `GRANT INSERT ON metaldocs.audit_events TO metaldocs_app` |
| `0190_audit_events_tenant_id.sql` | `archive/migrations/` | Adds `tenant_id TEXT NOT NULL DEFAULT ''`, `idx_audit_events_tenant_id` |
| `0193_audit_events_hash_chain.sql` | `archive/migrations/` | Adds `audit_sequence`, `prev_hash`, `row_hash`; defines `metaldocs.audit_event_row_hash()`; backfills chain; adds CHECK constraints; upgrades grant to `GRANT INSERT, SELECT` |
| `0224_audit_export_jobs_pr6.sql` | `db/migrations/` | Creates `metaldocs.audit_export_jobs` with 2 indexes |

---

## 7. Config & environment

| Variable | Where read | Semantics |
|---|---|---|
| `AUDIT_RETENTION_DAYS` | `main.go:575` — `strconv.Atoi(os.Getenv("AUDIT_RETENTION_DAYS"))` | When `> 0`, enables a 24-hour goroutine that hard-deletes rows older than N days. Default: disabled (0). |
| `ENABLE_JOB_AUDIT_INTEGRITY_VALIDATOR` | `main.go:544` — `jobEnabled("ENABLE_JOB_AUDIT_INTEGRITY_VALIDATOR")` where `jobEnabled` checks `!strings.EqualFold(..., "false")` | When not set to `"false"`, registers the integrity validator scheduled job. Default: enabled (truthy unless explicitly `"false"`). |

The audit module's own packages (`domain`, `application`, `delivery/http`, `infrastructure/*`) read no environment variables directly.

---

## 8. Concurrency & async

- **Advisory lock serialisation (`postgres.Writer.RecordTx`):** each INSERT acquires `pg_advisory_xact_lock(90120260513004)` (`postgres/writer.go:45`) within the transaction. This serialises hash-chain writes across all connections on the same Postgres cluster. The lock is released automatically when the transaction commits or rolls back.

- **Memory writer mutex:** `memory.Writer` uses `sync.Mutex` (`memory/writer.go:16`) around `Record` and the snapshot copy in `ListEvents`. `memory.ExportJobRepository` uses its own `sync.Mutex` (`memory/exports.go:5`).

- **Retention purge goroutine (`main.go:576`):** an anonymous goroutine runs a 24-hour ticker; terminates on context cancellation. Issued against the main `*sql.DB` directly — not through the domain port.

- **`fetchAll` in `ExportEvents`:** iterates reader pages synchronously in the request goroutine (`application/service.go:239-261`). No additional goroutines; the entire export is blocking in the HTTP handler goroutine.

- **Integrity validator job:** invoked by the scheduler's own goroutine on each epoch. No goroutines spawned by the job itself.

There is no outbox, no background worker for export, no channels, and no timer inside the audit module packages themselves. The async export path mentioned in `application/service.go:30` ("async worker lands in a later PR") is **not implemented**.

---

## 9. Error handling & observability

### Error patterns

- **Domain errors:** `ErrInvalidEvent` (`domain/port.go:25`), `ErrExportJobNotFound` (`domain/port.go:170`) are sentinel errors tested with `errors.Is`. `NewEvent` validates required fields and returns `ErrInvalidEvent`-wrapped errors (`domain/port.go:46`).
- **Application errors:** `ErrTenantRequired`, `ErrReaderRequired`, `ErrActorRequired`, `ErrInvalidFormat`, `ErrExportTooLarge`, `ErrExportRepoMissing`, `ErrCounterMissing`, `ErrExportTokenMismatch` (`application/service.go:18-27`). All tested with `errors.Is` in handler and tests.
- **Infrastructure errors:** wrapped with `fmt.Errorf("context: %w", err)` throughout (`postgres/writer.go`). `sql.ErrNoRows` is mapped to `domain.ErrExportJobNotFound` in `scanJob` (`exports.go:83`).
- **RFC 9457 (Problem Details):** all 4xx/5xx paths in `handler.go` emit `application/problem+json` via `writeProblem(w, problem.New(...))`. Status 405 is the only exception — it calls `w.WriteHeader(405)` with no body (`handler.go:82`).
- **Fire-and-forget audit emission:** consumer call sites either discard the `Record` error entirely (`admin_handler.go:403`) or log it without propagation (`main.go:827`, `auth/handler.go:211`).

### Logging

- `slog.Error` on list/export/download/status failures in `handler.go:101,108,210,263,293`.
- `slog.Warn` on write-response failure `handler.go:438` and export write error `handler.go:312`.
- `slog.ErrorContext` / `slog.InfoContext` in `audit_integrity_validator/job.go:21,27,37`.
- `log.Printf` (unstructured, legacy) for retention purge failure `main.go:588` and documents adapter failure `main.go:827`.

No metrics are emitted by any audit module package. No distributed tracing spans are created. The `trace_id` field is stored per event (sourced from `requesttrace.Resolve(ctx)` via `traceIDFromContext` at `main.go:831`) but the audit module does not read or propagate it as an OTel span.

---

## 10. Legacy / duplication / smell flags

- **F-01: Limit clamp divergence between service (100) and wiki claim ([1..200]).** `application/service.go:94-99` clamps to `[1..100]` (default 50); the handler also calls `pagination.ClampLimit` (MaxLimit=100) at `handler.go:329`. Multiple sections of `wiki/modules/audit.md` (key files line 9, §5.2 line 147, §6.2 sequence diagram line 216, §8.4 line 281) state the range is `[1..200]`. The code is correct; the wiki doc is stale. Severity: low (documentation drift, no runtime impact).

- **F-02: Inline `BYTEA` payload in `audit_export_jobs` with no size guard.** `domain/port.go:158` (`Payload []byte`) and `exports.go:44` write the rendered blob directly into the row. For exports approaching 50,000 rows, the blob can reach tens of megabytes stored in a single Postgres row and loaded fully into API process memory. `ExportJob` struct holds the entire blob in-process. The `application/service.go:30` comment acknowledges "async worker lands in a later PR" but it is not implemented. Severity: medium (scalability and memory pressure for large exports).

- **F-03: `audit_export_jobs.tenant_id` column type mismatch between migration (UUID) and domain (string).** `db/migrations/0224_audit_export_jobs_pr6.sql:22` defines `tenant_id UUID NOT NULL`. `domain/port.go:146` declares `TenantID string`. `exports.go:43` passes `job.TenantID` as a string to the INSERT. Postgres will accept a UUID-string if the value is a valid UUID, but the application type and database type are not aligned — a non-UUID tenant_id value would fail at the DB layer. Severity: medium (schema/domain type mismatch that will surface on non-UUID tenant IDs or future schema hardening).

- **F-04: `memory.Writer.RecordTx` ignores the `*sql.Tx` argument.** `memory/writer.go:30-32` calls `w.Record(ctx, event)` unconditionally, bypassing the transaction entirely. In tests using the memory adapter, transactional audit writes are committed regardless of whether the surrounding transaction rolls back. This makes the memory adapter unsuitable for testing transactional rollback semantics. Severity: low (test fidelity gap; not a production path).

- **F-05: `application.Service` is an exported struct with public zero-value construction.** `service_test.go:36` constructs `&application.Service{}` directly (bypassing `NewService`) to test the nil-reader path. This works because `Service` is exported without unexported fields preventing direct construction. The test is intentional, but it couples the test to the struct's internal layout. Severity: info (minor test encapsulation smell; no production impact).

- **F-06: Sub-resource routing via inline string parsing rather than `ServeMux` patterns.** `handler.go:232-246` parses `strings.SplitN(tail, "/", 2)` to distinguish `/export/{id}` (status) from `/export/{id}/download` (download). Go 1.22+ `http.ServeMux` supports path parameters natively; the current approach is pre-1.22 style and is inconsistent with other modules that use oapi-codegen wiring. Severity: low (naming drift from sibling modules; not a bug).

- **F-07: Missing GRANT statement in `0224_audit_export_jobs_pr6.sql`.** The migration creates `metaldocs.audit_export_jobs` but includes no `GRANT INSERT, SELECT ON metaldocs.audit_export_jobs TO metaldocs_app`. Access works only if the app role inherits schema-level or PUBLIC grants. Sister table `audit_events` received its SELECT grant in 0193. Severity: low (latent; T-009 pattern repeated for the new table).

- **F-08: `audit_integrity_validator` job emits no metric on violation.** `job.go:27-34` logs the violation with `slog.ErrorContext` and returns an error to the scheduler. There is no counter, gauge, or alert sink. Operators cannot detect chain violations from a monitoring dashboard without log scraping. Severity: low (observability gap).

- **F-09: Retention purge goroutine bypasses the domain port.** `main.go:585-587` executes `DELETE FROM metaldocs.audit_events WHERE occurred_at < $1` directly via `deps.SQLDB.ExecContext`, not through any `domain.Writer` method. This means the domain `Writer` interface has no delete semantics, deletes are invisible to the audit module's own tests, and no audit row is emitted for the purge action itself. Severity: low (architectural inconsistency; no immediate regression risk).

- **F-10: `buildWhere` uses `ILIKE` for full-text search on `payload::text`.** `postgres/writer.go:260-264` appends `payload::text ILIKE $n ESCAPE '\'` when a `q` query param is supplied. This forces a full-table cast and sequential scan on a JSONB column for every search request; no GIN/GiST index covers this access path. Severity: medium (performance; will degrade as `audit_events` grows).

- **F-11: No Go doc comments on any exported symbol.** `domain/port.go`, `application/service.go`, `delivery/http/handler.go`, and both infrastructure packages contain zero Go doc comments on exported types, functions, methods, or constants. This is the T-012 item from the existing tech debt register, confirmed by direct code reading. Severity: info.

---

## 11. Wiki drift

The existing wiki docs (`wiki/modules/audit.md`, `wiki/modules/audit-tech-debt.md`) are substantially accurate and current. The following specific discrepancies were found between wiki claims and the code as read:

1. **`wiki/modules/audit.md` key files (line 9), §5.2 (line 147), §6.2 sequence diagram (line 216), §8.4 (line 281) state limit range is `[1..200]`, default 50.** Code reality: `application/service.go:94-99` clamps to `[1..100]`, default 50. The upper bound in both service and handler is 100 (matched by `pagination.MaxLimit=100`). The value 200 is not in the code.

2. **`wiki/modules/audit.md` §5.3 and API Route Truth Table (line 163) show only `GET /api/v1/audit/events` as the single route.** Code reality: `handler.go:69-73` registers three routes (`/api/v1/audit/events`, `/api/v1/audit/events/export`, `/api/v1/audit/events/export/`), providing four logical HTTP operations. The export routes and their `CapAuditRead` bindings (`permissions.go:233-234`) were added in PR-6 and are not reflected in the §5.3 table or the API route truth table. The §8.1 paragraph does mention export routes at `:230-231` but the canonical route truth table is incomplete.

3. **`wiki/modules/audit.md` §8.6 (line 288) states "postgres.Writer.Record calls db.ExecContext directly; ListEvents calls db.QueryContext directly. Neither accepts a `*sql.Tx`."** Code reality: `domain/port.go:98` defines `RecordTx(ctx context.Context, tx *sql.Tx, event Event) error` on `Writer`; `postgres.Writer` implements it at `postgres/writer.go:44`; `memory.Writer` implements it at `memory/writer.go:30`. The `RecordTx` method accepting `*sql.Tx` exists and is actively used by `bypassAuditAdapter` and `documentsAuditAdapter`. The wiki claim describes the state before `RecordTx` was added and has not been updated.

4. **`wiki/modules/audit.md` §8.5 (line 285) states "trace_id...defaults to `'trace-local'`".** Code reality: `main.go:831-833` calls `requesttrace.Resolve(ctx)` for the trace ID, not a hardcoded default. The value `"trace-local"` may be what `requesttrace.Resolve` returns when no trace header is present, but this is sourced from the `requesttrace` platform package, not from the audit module itself. [runtime-unverified: actual fallback value of `requesttrace.Resolve` when no header is set].

5. **`wiki/modules/audit-tech-debt.md` T-009 states "No explicit GRANT SELECT on audit_events."** Code reality: archived `migrations/0193_audit_events_hash_chain.sql:110` adds `GRANT INSERT, SELECT ON TABLE metaldocs.audit_events TO metaldocs_app`. T-009 was written before 0193 was applied. The SELECT grant now exists in the archived migration tree. The T-009 item in the tech debt register should be closed, as the grant is present in 0193.

---

## 12. Open questions

- **[runtime-unverified]** `metaldocs.audit_export_jobs` table: migration `0224_audit_export_jobs_pr6.sql` is in `db/migrations/` (not yet in baseline `db/baseline/0001_current_schema.sql`). Whether this migration has been applied to the current deployed database cannot be confirmed without a live Postgres connection.

- **[runtime-unverified]** Export job `tenant_id` UUID/string mismatch (F-03): whether existing tenant IDs are valid UUIDs that Postgres accepts when cast from the Go string type. If they are UUIDs the cast succeeds silently; if not, INSERTs fail at runtime.

- **[runtime-unverified]** Access to `metaldocs.audit_export_jobs` for `metaldocs_app`: migration 0224 contains no GRANT. Reads and writes will fail unless a schema-level or PUBLIC grant covers this table, or unless `metaldocs_app` is the table owner.

- **[runtime-unverified]** The `ENABLE_JOB_AUDIT_INTEGRITY_VALIDATOR` flag defaults to enabled (the `jobEnabled` function returns true unless the value is literally `"false"`). Whether the validator job runs in production and at what interval (scheduler epoch) cannot be determined without reading the scheduler configuration and a live run.

- **[runtime-unverified]** Performance of JSONL/CSV `payload::text ILIKE` full-text query (F-10): actual query plan and row scan cost under production data volumes are unknown.
