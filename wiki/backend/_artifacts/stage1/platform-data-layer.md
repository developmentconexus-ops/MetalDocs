# Stage-1 Audit Artifact — Platform Data Layer

> **Area:** platform-data-layer
> **Packages covered:** `internal/platform/db`, `internal/platform/migrate`, `internal/platform/bootstrap`, `internal/platform/objectstore`, `internal/platform/storage`, `internal/platform/cache`
> **Produced:** 2026-06-10
> **Status:** Stage-1 (truth map only — no redesign proposals)

---

## 1. Identity & purpose

The platform data layer is the set of cross-cutting packages that establish database connectivity, run schema migrations at startup, wire all module-level dependencies into a single dependency bundle passed to the composition root, manage blob-storage presigning and download for documents and templates, and provide the raw MinIO object-store adapter consumed by the PDF conversion pipeline.

The layer owns five distinct concerns: (1) a thin Postgres connection factory (`db/postgres`); (2) a forward-only SQL migration runner keyed on `public.schema_migrations` (`migrate`); (3) three dependency-injection bootstrap factories — for the API server, the outbox worker, and the River jobs worker — that assemble every concrete infrastructure adapter and return it as a typed struct (`bootstrap`); (4) two MinIO presigner value objects for tenant-namespaced presigned PUT/GET URLs, plus helper functions for canonical object-key construction (`objectstore`); and (5) a raw MinIO blob-store adapter that satisfies the `pdfObjectStore` interface consumed by `platform/servicebus.GotenbergPDFClient` (`storage/minio`).

`platform/cache` is an empty placeholder — a `.gitkeep` only, no Go code — and therefore has no behavior. It is a drift bait violation of REQ-TOP-3 in `backend-target-architecture.md`.

---

## 2. File inventory

### `internal/platform/db`

| File | Role |
|---|---|
| `internal/platform/db/.gitkeep` | Empty directory marker; no Go package declared here |
| `internal/platform/db/postgres/connect.go` | Package `postgres`. Exports `Open(ctx, dsn) (*sql.DB, error)` — the sole connection factory. Hard-codes pool settings (25/25 max conns, 30 min lifetime, 5 min idle). |

### `internal/platform/migrate`

| File | Role |
|---|---|
| `internal/platform/migrate/migrate.go` | Package `migrate`. Exports `Apply(ctx, db, dir, log)` — reads `*.sql` files from `dir`, skips versions already in `public.schema_migrations`, enforces explicit BEGIN/COMMIT guard, executes under `pg_advisory_lock`. |
| `internal/platform/migrate/migrate_test.go` | Unit tests using `go-sqlmock` covering skip logic, advisory lock, explicit transaction guard, and schema_migrations missing-table tolerance. |
| `internal/platform/migrate/revision_number_zero_based_integration_test.go` | Integration test (build tag `integration`) validating the two-phase zero-based revision number shift migration logic against a live DB. Skips when `DATABASE_URL`/`METALDOCS_DATABASE_URL` are unset. |

### `internal/platform/bootstrap`

| File | Role |
|---|---|
| `internal/platform/bootstrap/api.go` | Package `bootstrap`. `APIDependencies` struct bundles all module repos, audit, messaging publisher, Gotenberg, MinIO clients, and cleanup. `BuildAPIDependencies` switches on `repoMode` (postgres vs memory) to wire concrete adapters or in-memory stubs. `buildMinioClients` creates two separate MinIO client instances: one for internal signed ops, one for browser-reachable presigning. |
| `internal/platform/bootstrap/jobs.go` | `JobsDependencies` struct + `BuildJobsDependencies`: opens its own Postgres connection, runs `MigrateRiverSchema`, builds a River client bundle. `MigrateRiverSchema` is also exported for reuse in `main.go`. |
| `internal/platform/bootstrap/worker.go` | `WorkerDependencies` struct + `BuildWorkerDependencies`: opens Postgres, wires the outbox consumer, optionally wires the Gotenberg PDF converter, reads `METALDOCS_FANOUT_URL` and `METALDOCS_DOCX_RENDERER_SERVICE_TOKEN` from env. |
| `internal/platform/bootstrap/api_test.go` | Tests for Gotenberg health-check status (up/skipped/down), MinIO invalid public endpoint error. |
| `internal/platform/bootstrap/worker_test.go` | Tests for `workerClaimLease` floor enforcement. |

### `internal/platform/objectstore`

| File | Role |
|---|---|
| `internal/platform/objectstore/.gitkeep` | Empty directory marker from initial commit; superfluous now that Go files exist. |
| `internal/platform/objectstore/document_presigner.go` | `DocumentPresigner` struct. Methods: `PresignRevisionPUT`, `PresignObjectGET`, `AdoptTempObject` (copy + delete), `DeleteObject`, `HashObject` (streaming SHA-256 with size cap), `Exists`. Holds dual MinIO clients: `client` for data ops, `signingClient` for URL generation. `isNoSuchKeyErr` helper defined here. |
| `internal/platform/objectstore/document_presigner_export.go` | Adds `HeadObject` and `SizeObject` methods to `DocumentPresigner` via stat-only calls. Named `_export` — file was appended rather than merged. |
| `internal/platform/objectstore/templates_presigner.go` | `TemplatesPresigner` struct. Methods: `PresignPUT`, `PresignGET`, `HeadContentHash`, `Delete`. Mirrors the dual-client pattern from `DocumentPresigner`. Reuses `isNoSuchKeyErr` (defined in sibling file, same package). |
| `internal/platform/objectstore/template_keys.go` | Pure key-generation helpers: `TemplateDocxKey(tenantID, templateID, versionNum)` → `tenants/{t}/templates/{id}/v{n}.docx` and `TemplateSchemaKey(...)` → `...v{n}.schema.json`. |
| `internal/platform/objectstore/document_presigner_test.go` | Unit test asserting that `PresignRevisionPUT` and `PresignObjectGET` use the public (browser-reachable) MinIO client endpoint, not the internal endpoint. |
| `internal/platform/objectstore/templates_presigner_test.go` | Same dual-endpoint assertion for `TemplatesPresigner`. |
| `internal/platform/objectstore/template_keys_test.go` | Unit tests for canonical key format correctness. |

### `internal/platform/storage`

| File | Role |
|---|---|
| `internal/platform/storage/minio/store.go` | Package `minio`. `Store` struct backed by a single MinIO client. Methods: `EnsureBucket` (check-or-create with `autoCreateBucket` guard), `Save` (PutObject, content bytes), `Open` (GetObject + Stat), `Delete` (RemoveObject). Satisfies the `pdfObjectStore` interface in `platform/servicebus/gotenberg_pdf.go:49-52`. |

### `internal/platform/cache`

| File | Role |
|---|---|
| `internal/platform/cache/.gitkeep` | Empty directory marker only. No Go source files exist. Package has zero behavior. Present since initial commit (2026-03-16). |

---

## 3. Public surface

### `internal/platform/db/postgres`

```go
func Open(ctx context.Context, dsn string) (*sql.DB, error)
```

Used by all three bootstrap factories and `apps/api/cmd/metaldocs-e2e-seed/main.go`.

### `internal/platform/migrate`

```go
func Apply(ctx context.Context, db *sql.DB, dir string, log *slog.Logger) error
```

Called once at API startup (`apps/api/cmd/metaldocs-api/main.go:191`).

### `internal/platform/bootstrap`

```go
type APIDependencies struct { ... }       // full module dep bundle for metaldocs-api
type JobsDependencies struct { ... }      // River client bundle for metaldocs-jobs
type WorkerDependencies struct { ... }    // outbox consumer + PDF converter for metaldocs-worker

func BuildAPIDependencies(ctx context.Context, repoMode string, attachmentsCfg config.AttachmentsConfig) (APIDependencies, error)
func BuildJobsDependencies(ctx context.Context, cfg config.JobsConfig, workerFactory JobsWorkerFactory) (JobsDependencies, error)
func BuildWorkerDependencies(ctx context.Context, workerCfg config.WorkerConfig) (WorkerDependencies, error)
func MigrateRiverSchema(ctx context.Context, db *sql.DB, schema string) error
```

`MigrateRiverSchema` is also called directly from `main.go:439` for the API server's River client wiring path.

### `internal/platform/objectstore`

```go
// DocumentPresigner
func NewDocumentPresigner(client, signingClient *minio.Client, bucket string, ttl time.Duration, maxSizeBytes int64) *DocumentPresigner
func (p *DocumentPresigner) PresignRevisionPUT(ctx, tenantID, docID, contentHash string) (url, key string, err error)
func (p *DocumentPresigner) PresignObjectGET(ctx, storageKey string) (string, error)
func (p *DocumentPresigner) AdoptTempObject(ctx, tmpKey, finalKey string) error
func (p *DocumentPresigner) DeleteObject(ctx, key string) error
func (p *DocumentPresigner) HashObject(ctx, key string) (string, error)
func (p *DocumentPresigner) Exists(ctx, key string) (bool, error)
func (p *DocumentPresigner) HeadObject(ctx, key string) (bool, error)     // in document_presigner_export.go
func (p *DocumentPresigner) SizeObject(ctx, key string) (int64, error)    // in document_presigner_export.go

// TemplatesPresigner
func NewTemplatesPresigner(client, signingClient *minio.Client, bucket string, maxSizeBytes int64) *TemplatesPresigner
func (p *TemplatesPresigner) PresignPUT(ctx, key string, expires time.Duration) (string, error)
func (p *TemplatesPresigner) PresignGET(ctx, key string, expires time.Duration) (string, error)
func (p *TemplatesPresigner) HeadContentHash(ctx, key string) (string, error)
func (p *TemplatesPresigner) Delete(ctx, key string) error

// Key helpers
func TemplateDocxKey(tenantID, templateID string, versionNum int) string
func TemplateSchemaKey(tenantID, templateID string, versionNum int) string
```

There are no HTTP routes in this area.

### `internal/platform/storage/minio`

```go
func NewStore(cfg config.AttachmentsConfig) (*Store, error)
func (s *Store) EnsureBucket(ctx context.Context) error
func (s *Store) Save(ctx context.Context, storageKey string, content []byte) error
func (s *Store) Open(ctx context.Context, storageKey string) (io.ReadCloser, error)
func (s *Store) Delete(ctx context.Context, storageKey string) error
```

---

## 4. Logic flows

### Flow 1: API startup — DB connect, migrate, bootstrap

1. `main.go:180` calls `bootstrap.BuildAPIDependencies(ctx, repoMode, attachmentsCfg)`.
2. `bootstrap/api.go:74` switches on `repoMode`. For `"postgres"`:
   a. `config.LoadPostgresConfig()` reads `DATABASE_URL` or `PGHOST`/`PGPORT`/`PGDATABASE`/`PGUSER`/`PGPASSWORD` (`config/postgres.go:15-57`).
   b. `pgdb.Open(ctx, pgCfg.DSN)` opens `database/sql` via `pgx/stdlib`, sets pool limits, pings with 5s timeout (`db/postgres/connect.go:12-31`).
   c. If `attachmentsCfg.Provider == "minio"`, calls `buildMinioClients` to create two `*minio.Client` instances (internal + public endpoints) and also `miniostore.NewStore` + `servicebus.NewGotenbergPDFClient` if Gotenberg is enabled.
   d. Returns the fully populated `APIDependencies` struct including `Cleanup: func() { db.Close() }`.
3. `main.go:186-194`: if `deps.SQLDB != nil` and `METALDOCS_SKIP_STARTUP_MIGRATIONS != "true"`, reads `METALDOCS_MIGRATIONS_DIR` (default `"db/migrations"`) and calls `migrate.Apply`.
4. `migrate.Apply` (`migrate/migrate.go:31-95`):
   a. Acquires a session-level Postgres advisory lock `0x4D444D4947528000` via `pg_advisory_lock`.
   b. Reads applied versions from `public.schema_migrations` via `loadApplied`; if the table does not exist (PostgreSQL error `42P01`) returns an empty map.
   c. Scans `db/migrations/*.sql` for files matching `^\d{4}_`; sorts lexically.
   d. For each unapplied file: calls `requireExplicitTransactionGuard` (must start with `BEGIN`, must end with `COMMIT`); executes the full file body as a single `ExecContext`.
   e. Releases advisory lock in a deferred call (uses `context.Background()` to survive parent cancellation).

### Flow 2: Jobs binary startup — River schema migration

1. `apps/jobs/cmd/metaldocs-jobs/main.go` calls `bootstrap.BuildJobsDependencies(ctx, cfg, workerFactory)`.
2. `bootstrap/jobs.go:25-66`:
   a. `pgdb.Open` for a dedicated connection pool.
   b. `MigrateRiverSchema(ctx, db, cfg.RiverSchema)` — delegates to `rivermigrate.New(riverdatabasesql.New(db), ...)` then `migrator.Migrate(ctx, DirectionUp, nil)`. Schema is the `METALDOCS_JOBS_RIVER_SCHEMA` env var (empty = default River schema).
   c. Calls the injected `workerFactory(db)` to build the River workers registry.
   d. Returns `JobsDependencies{River, SQLDB, Cleanup}`.
3. `MigrateRiverSchema` is also exposed publicly and called directly from `main.go:439` when the API binary needs River enqueuer access (River schema may or may not have been migrated by the jobs binary yet).

### Flow 3: Document presign flow (upload path)

1. Handler calls `docPresigner.PresignRevisionPUT(ctx, tenantID, docID, contentHash)`.
2. `objectstore/document_presigner.go:41-51`: constructs object key `tenants/{tenantID}/documents/{docID}/revisions/{contentHash}.docx`.
3. Calls `signingClient.PresignedPutObject(ctx, bucket, key, ttl)` — uses the **public** MinIO client endpoint so the returned URL is browser-reachable.
4. Returns URL and key to the handler; the handler returns both to the client. The browser then PUTs the file directly to MinIO, bypassing the API.
5. After the client confirms upload, the handler calls `docPresigner.HashObject(ctx, tmpKey)` to compute a server-side SHA-256 over the object (with a `maxSizeBytes+1` read limit to detect oversize).
6. On commit: handler calls `docPresigner.AdoptTempObject(ctx, tmpKey, finalKey)` which copies via `CopyObject` then silently drops the temp key with `RemoveObject`.

### Flow 4: PDF conversion via GotenbergPDFClient

1. An outbox worker dequeues a PDF-convert job and calls `pdfConverter.ConvertPDF(ctx, req)`.
2. `servicebus/gotenberg_pdf.go:70-108`:
   a. `store.Open(ctx, req.DocxKey)` streams the DOCX from MinIO (`storage/minio/store.go:65-75`).
   b. `io.ReadAll` buffers the entire DOCX in memory.
   c. `converter.ConvertDocxToPDFWithOptions(ctx, docx, paperSize, landscape)` sends the bytes to Gotenberg.
   d. Computes SHA-256 of the returned PDF bytes.
   e. `store.Save(ctx, req.OutputKey, pdf)` writes the PDF to MinIO (`storage/minio/store.go:54-63`).
   f. Returns `ConvertPDFResult{OutputKey, ContentHash, SizeBytes}`.
3. The `pdfObjectStore` interface (`gotenberg_pdf.go:49-52`) is satisfied by `*storage/minio.Store`. No local-storage fallback for PDF conversion.

### Flow 5: Bootstrap memory mode (development / testing)

1. `BuildAPIDependencies(ctx, "memory", ...)` takes the `default` branch in `bootstrap/api.go:127-155`.
2. Wires `authmemory.NewRepository()`, `auditmemory.NewWriter()`, `auditmemory.NewExportJobRepository()`, `nooppub.NewPublisher()`. No Postgres, no MinIO.
3. Calls `authRepo.UpsertUserAndAssignRole` for each user/role pair from `authn.DevRoleMap()` — seeds in-memory state without a database.
4. `StatusProvider` is `observability.NewStaticRuntimeStatusProvider` (always returns 200 OK).
5. `SQLDB`, `PDFConverter`, `MinioClient`, `MinioPublicClient`, `MinioBucket` are all nil; callers guard on nil before using.

---

## 5. Dependencies

### Outbound — imports by each package

**`internal/platform/db/postgres`**
- `database/sql` (stdlib)
- `github.com/jackc/pgx/v5/stdlib` (blank-import side-effect: registers "pgx" driver)

**`internal/platform/migrate`**
- `database/sql`, `os`, `path/filepath`, `regexp`, `sort`, `strings`, `log/slog`
- `github.com/jackc/pgx/v5/pgconn` (for `PgError` code check on missing table)

**`internal/platform/bootstrap`**
- `metaldocs/internal/platform/db/postgres` (open)
- `metaldocs/internal/platform/migrate` (Apply, in api.go indirectly through main.go; MigrateRiverSchema in jobs.go)
- `metaldocs/internal/platform/config` (LoadPostgresConfig, LoadAttachmentsConfig, LoadGotenbergConfig, LoadWorkerConfig, LoadJobsConfig)
- `metaldocs/internal/platform/messaging`, `messaging/noop`, `messaging/outbox/postgres`
- `metaldocs/internal/platform/storage/minio` (miniostore.NewStore)
- `metaldocs/internal/platform/servicebus` (GotenbergPDFClient)
- `metaldocs/internal/platform/render/gotenberg` (Client)
- `metaldocs/internal/platform/authn`, `tenant`, `observability`, `jobs/river`
- `metaldocs/internal/modules/audit`, `auth`, `iam` infrastructure packages
- `github.com/minio/minio-go/v7` + `credentials`
- `github.com/riverqueue/river`, `riverdriver/riverdatabasesql`, `rivermigrate`

**`internal/platform/objectstore`**
- `github.com/minio/minio-go/v7`
- `metaldocs/internal/modules/documents/domain` (ErrUploadMissing)
- `metaldocs/internal/modules/templates/domain` (ErrUploadMissing)
- `crypto/sha256`, `encoding/hex`, `io`, `log`, `net/url`, `strings`, `time`

**`internal/platform/storage/minio`**
- `github.com/minio/minio-go/v7` + `credentials`
- `metaldocs/internal/platform/config` (AttachmentsConfig)
- `bytes`, `io`, `context`

### Inbound — verified with grep

| Package | Imported by |
|---|---|
| `platform/db/postgres` | `bootstrap/api.go`, `bootstrap/jobs.go`, `bootstrap/worker.go`, `apps/api/cmd/metaldocs-e2e-seed/main.go` |
| `platform/migrate` | `apps/api/cmd/metaldocs-api/main.go` (only consumer) |
| `platform/bootstrap` | `apps/api/cmd/metaldocs-api/main.go`, `apps/worker/cmd/metaldocs-worker/main.go`, `apps/jobs/cmd/metaldocs-jobs/main.go` |
| `platform/objectstore` | `apps/api/cmd/metaldocs-api/main.go`, `internal/platform/objectstore/template_keys_test.go` |
| `platform/storage/minio` | `bootstrap/api.go`, `bootstrap/worker.go` |
| `platform/cache` | No imports (empty package, no files) |

---

## 6. Persistence

### `internal/platform/migrate`

Reads from and writes to `public.schema_migrations`. The `SELECT version FROM public.schema_migrations` query is the idempotency check. Each migration SQL file must `INSERT INTO public.schema_migrations (version) VALUES ('NNNN')` inside its own `BEGIN/COMMIT` block — enforced by convention, not by the runner.

Advisory lock key: `0x4D444D4947528000` (a named constant in `migrate.go:24`).

Migration files are in `db/migrations/` (0203 through 0233 as of 2026-06-10). Prerequisite extensions in `db/prerequisites/0001_extensions.sql` (only `pgcrypto`). Curated baseline in `db/baseline/0001_current_schema.sql`. Reference data in `db/reference-data/0001_product_reference_data.sql`. Dev seeds in `db/dev-seeds/0001_local_dev_seed.sql`.

The runner does **not** apply the prerequisite, baseline, or reference-data files — those are applied by the bootstrap script `scripts/dev-bootstrap-baseline.ps1`. The runner only applies files from the `db/migrations/` tail.

### `internal/platform/bootstrap`

`MigrateRiverSchema` writes River's own schema tables via `rivermigrate` — these are separate from the product `public.schema_migrations` ledger.

### Other packages in this area

`objectstore`, `storage/minio`, `db/postgres`, `cache` — no direct table access. They interact with Postgres only via the `*sql.DB` handle already open; `storage/minio` and `objectstore` interact with MinIO only.

---

## 7. Config & environment

### `internal/platform/db/postgres` (via `config/postgres.go`)

| Variable | Required | Default | Notes |
|---|---|---|---|
| `DATABASE_URL` | Preferred | — | Full DSN; takes priority over PG* vars |
| `PGHOST` | Required if no DSN | — | |
| `PGPORT` | No | `5432` | |
| `PGDATABASE` | Required if no DSN | — | |
| `PGUSER` | Required if no DSN | — | |
| `PGPASSWORD` | Required if no DSN | — | |
| `PGSSLMODE` | No | `require` | |

Pool settings are hard-coded in `db/postgres/connect.go:17-21`: 25 max open, 25 max idle, 30-minute lifetime, 5-minute idle timeout. Not configurable via env.

### `internal/platform/migrate` (called from `main.go`)

| Variable | Required | Default | Notes |
|---|---|---|---|
| `METALDOCS_MIGRATIONS_DIR` | No | `"db/migrations"` | Directory of `*.sql` migration files |
| `METALDOCS_SKIP_STARTUP_MIGRATIONS` | No | `""` (false) | Set to `"true"` to skip migrate.Apply at startup |

### `internal/platform/bootstrap` — MinIO / attachments (`config/attachments.go`)

| Variable | Required | Default | Notes |
|---|---|---|---|
| `METALDOCS_STORAGE_PROVIDER` | No | `"local"` | One of: `memory`, `local`, `minio` |
| `METALDOCS_ATTACHMENTS_SIGNING_SECRET` | Yes | — | Min 32 bytes |
| `METALDOCS_ATTACHMENTS_ROOT` | No | `"non_git/attachments"` | Local storage root |
| `METALDOCS_ATTACHMENTS_DOWNLOAD_TTL_SECONDS` | No | `300` | Min 30 |
| `APP_ENV` | No | `"local"` | |
| `METALDOCS_MINIO_ENDPOINT` | If MinIO | — | Internal cluster endpoint |
| `METALDOCS_MINIO_PUBLIC_ENDPOINT` | No | = `ENDPOINT` | Browser-reachable endpoint; defaults to internal |
| `METALDOCS_MINIO_ACCESS_KEY` | If MinIO | — | |
| `METALDOCS_MINIO_SECRET_KEY` | If MinIO | — | |
| `METALDOCS_MINIO_BUCKET` | If MinIO | — | |
| `METALDOCS_MINIO_USE_SSL` | No | `false` | |
| `METALDOCS_MINIO_AUTO_CREATE_BUCKET` | No | `false` | |

### `internal/platform/bootstrap` — Gotenberg

| Variable | Required | Default | Notes |
|---|---|---|---|
| `METALDOCS_GOTENBERG_URL` | No | `""` | Empty = Gotenberg disabled; PDF conversion omitted |

### `internal/platform/bootstrap` — Repository mode

| Variable | Required | Default | Notes |
|---|---|---|---|
| `METALDOCS_REPOSITORY` | No | `"memory"` | `"postgres"` or `"memory"` |

### `internal/platform/bootstrap` — Worker env (read directly, not via config struct)

| Variable | Read in | Notes |
|---|---|---|
| `METALDOCS_FANOUT_URL` | `bootstrap/worker.go:47` | Passed through to `WorkerDependencies.FanoutURL` |
| `METALDOCS_DOCX_RENDERER_SERVICE_TOKEN` | `bootstrap/worker.go:48` | Passed through to `WorkerDependencies.FanoutToken` |

### `internal/platform/bootstrap` — Jobs

| Variable | Required | Default | Notes |
|---|---|---|---|
| `METALDOCS_JOBS_RIVER_SCHEMA` | No | `""` (River default) | Postgres schema for River tables |
| `METALDOCS_JOBS_ENABLED` | No | `true` | |
| `METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` | No | `10` | River queue concurrency |

---

## 8. Concurrency & async

**`migrate.Apply`** — single-threaded execution under a Postgres advisory lock. No goroutines. The lock guarantees at-most-one runner across all API instances during startup.

**`db/postgres/connect.go`** — `*sql.DB` is safe for concurrent use. Pool limits are set at creation; all goroutines in the process share the single pool returned by `Open`.

**`bootstrap`** — each `Build*` function is called once at startup and is not goroutine-safe by design (single composition root pattern). The returned `Cleanup` functions are not goroutine-safe either; they are called in `defer deps.Cleanup()` from the main goroutine.

**`objectstore/DocumentPresigner`** — `*minio.Client` is documented as goroutine-safe; the presigner struct is therefore safe for concurrent use after construction.

**`storage/minio.Store`** — same: `*minio.Client` is goroutine-safe; `Store` carries no mutable state after construction.

**No goroutines, channels, outbox writes, or timers** are present in any file in this area. All background-processing concerns (River jobs, outbox relaying, PDF conversion) are initiated in `main.go` or the worker binary — not inside these platform packages.

---

## 9. Error handling & observability

**`migrate.Apply`** — all errors are wrapped with `fmt.Errorf("migrate: ...: %w", err)` giving a clear operation context chain. The advisory lock release uses `context.Background()` to survive context cancellation, and only sets `retErr` if the unlock fails and no prior error is already in flight (`migrate.go:44-48`).

**`db/postgres/connect.go`** — if `PingContext` fails, `db.Close()` is called (error silently discarded: `_ = db.Close()`) before returning the wrapped error.

**`objectstore/DocumentPresigner`** — nil-client guard on every method returns a descriptive error. `isNoSuchKeyErr` inspects both `minio.ErrorResponse.Code` and string-contains fallbacks for robustness across MinIO SDK versions (`document_presigner.go:152-168`). `AdoptTempObject` silently logs (uses the legacy `"log"` package, not `slog`) cleanup failures when the temp object is missing after copy: `log.Printf("objectstore: adopt tmp cleanup failed ...")` (`document_presigner.go:80`).

**`storage/minio/store.go`** — errors wrapped with `fmt.Errorf("... minio: %w", err)`. No nil guards; `Store` is expected to always have a valid client (ensured by `NewStore`).

**RFC 9457** — not applicable; this layer never writes HTTP responses.

**`slog`** — `migrate.Apply` accepts a `*slog.Logger` and logs `Info("applying migration", ...)` and `Info("migrations done", ...)`. All other packages in this area perform no logging (except the `log.Printf` smell noted above).

**No metrics or tracing calls** exist in any file in this area.

---

## 10. Legacy / duplication / smell flags

- **`platform/cache` is an empty placeholder** (`internal/platform/cache/.gitkeep` only; present since initial commit 2026-03-16, `912879cba`). Zero Go files. Zero behavior. Violates REQ-TOP-3 ("every platform package either has production consumers or does not exist"). This is the `C4` gap explicitly called out in `backend-blueprint.md:186`. RF-7 (per area-specific instructions).

- **`platform/storage` has one file, no Go package-level interface, and `platform/db` has only a subdirectory** — `internal/platform/db/.gitkeep` is a directory marker leftover from the initial commit; the real package is `internal/platform/db/postgres/connect.go`. The `db` directory itself declares no Go package, which means the tree path `internal/platform/db` is not a Go package — only `internal/platform/db/postgres` is. The extra nesting level (`db/postgres`) adds path depth without adding namespace value given there is only one DB driver in the codebase. RF-7 applies; the nesting is shallow enough that consolidation to `internal/platform/dbconn` or similar could be worth evaluating in Stage 2.

- **`document_presigner_export.go` splits methods of `DocumentPresigner` across two files** (`document_presigner.go` + `document_presigner_export.go`). The `_export` suffix implies the file was appended to add methods later rather than merged. This is a minor maintenance friction: the split is not standard Go convention, and the `_export` name conventionally signals a test-helper export file (per Go convention for `export_test.go` patterns). File: `internal/platform/objectstore/document_presigner_export.go:1-38`.

- **`log.Printf` in `objectstore` instead of `slog`** (`document_presigner.go:10, 80`). The rest of the MetalDocs codebase uses `log/slog` for structured logging. The legacy `log.Printf` call in `AdoptTempObject` is inconsistent with the canonical observability pattern and loses request context (`internal/platform/objectstore/document_presigner.go:80`).

- **`DocumentPresigner` and `TemplatesPresigner` duplicate the dual-client MinIO pattern** and both embed the content-hash-via-streaming logic (`HashObject` vs `HeadContentHash`). The logic is not identical (different error types returned: `domain.ErrUploadMissing` in each, but from different module domains) so extraction to a shared helper would require resolving the domain-error coupling. This duplication is not accidental — it was produced by sequential feature additions. `internal/platform/objectstore/document_presigner.go:99-136` vs `templates_presigner.go:47-80`.

- **`bootstrap/api.go` imports module infrastructure directly** (`auditmemory`, `auditpg`, `authmemory`, `authpg`, `iampg`). This is the intended composition-root pattern for a modular monolith, but it means `bootstrap` is a transitive dependency of 3+ bounded contexts. Any module infrastructure change recompiles bootstrap. The file has grown to 216 lines; it is not yet a god-file but is the largest in this area.

- **`METALDOCS_FANOUT_URL` and `METALDOCS_DOCX_RENDERER_SERVICE_TOKEN` are read directly from `os.Getenv` in `bootstrap/worker.go:47-48`** rather than through the typed config system (`platform/config`). This breaks the 12-Factor typed-config-at-startup contract for these two values. `internal/platform/bootstrap/worker.go:47-48`.

- **Two separate `*minio.Client` instantiation paths exist** for presigning: `bootstrap/api.go:158-178` (`buildMinioClients`) creates two clients for `objectstore.*Presigner`; `bootstrap/api.go:97-100` and `worker.go:79` call `miniostore.NewStore` which creates a **third** client from the same credentials for PDF byte I/O. The result is that under full Gotenberg + MinIO mode, `BuildAPIDependencies` creates three separate MinIO client instances from the same config. `internal/platform/bootstrap/api.go:85-103`.

- **`MigrateRiverSchema` is called from two separate callers** with no guard against double execution: `bootstrap/jobs.go:36` (inside `BuildJobsDependencies`) and `main.go:439` (directly, for the API binary's enqueuer-only River bundle). `rivermigrate` is idempotent, so correctness is preserved, but the duplication is fragile. `internal/platform/bootstrap/jobs.go:68-79` and `apps/api/cmd/metaldocs-api/main.go:439`.

- **`internal/platform/objectstore/.gitkeep` is superfluous** — the directory already contains Go source files so the marker adds noise. Not a correctness issue. `internal/platform/objectstore/.gitkeep`.

- **TODO comment in `migrate.go:30`**: `// TODO: fail closed when a migration file omits explicit BEGIN/COMMIT so mixed transactional semantics cannot slip through.` — the guard `requireExplicitTransactionGuard` already exists and already fails on missing `BEGIN`. The TODO comment is stale; the described behavior is already implemented. `internal/platform/migrate/migrate.go:30`.

---

## 11. Wiki drift

No existing wiki doc covers this area directly. The following wiki documents make claims that can be checked against code:

- `wiki/architecture/backend-blueprint.md:175`: "pgx pool (`platform/db/postgres`)" — minor inaccuracy: the code uses `database/sql` with the pgx stdlib driver (`pgx/stdlib`), not `pgxpool`. There is no pgxpool import anywhere in `db/postgres`. The blueprint describes it as "pgx pool" which implies `pgxpool.Pool`; the actual implementation is `*sql.DB` with pgx as the underlying driver. This is a meaningful distinction: `pgxpool` offers named parameters, row scanning improvements, and explicit pool control; `database/sql` via pgx stdlib is more limited. `internal/platform/db/postgres/connect.go:6,13`.

- `wiki/database/migration-policy.md` claims "post-baseline migrations must be forward-only and idempotent". `migrate.Apply` does not enforce idempotency of the SQL within a migration file — it only skips files whose version is already in `schema_migrations`. The idempotency claim is a convention enforced by authors, not by the runner. Accurate as documentation policy, but could be read as a runtime guarantee. Not a code bug but a precision gap worth noting if the wiki is used as a specification.

- `wiki/architecture/backend-blueprint.md:186` ("C4. Caching — `platform/cache` directory exists but is empty"): accurate as of this audit. No change.

---

## 12. Open questions

- **[runtime-unverified]** Whether Postgres pool settings (25/25 max conns, `connect.go:17-21`) remain appropriate under production load. The values are hard-coded and cannot be tuned without a code change. No documented rationale exists for the specific numbers.

- **[runtime-unverified]** Whether `MinIOAutoCreateBucket` (`METALDOCS_MINIO_AUTO_CREATE_BUCKET`) is safe to enable in production. `EnsureBucket` in `storage/minio/store.go:37-52` creates the bucket when `autoCreateBucket` is true and the bucket is missing. If this runs against a production MinIO with stricter IAM policies, the create call may fail; the flag default is `false`, which is the safe default.

- **[runtime-unverified]** Whether `MigrateRiverSchema` called twice (once in `main.go:439` for the API binary and once in `BuildJobsDependencies` for the jobs binary) causes any race condition in a multi-instance deployment where both API and jobs start simultaneously. `rivermigrate` should be idempotent and uses its own locking, but this has not been confirmed against the River library version in vendor.

- **[runtime-unverified]** Whether the `MaxSizeBytes` cap in `HashObject` / `HeadContentHash` (defaulting to 25 MB when zero) produces a usable SHA-256 hash when content is exactly 25 MB (the `io.LimitReader(obj, limit+1)` reads up to `limit+1` bytes, then the `n > limit` check triggers — so at exactly `limit` bytes it passes, at `limit+1` it fails). This edge case is correct but untested.

- **Genuine unknown**: the `internal/platform/db/.gitkeep` marker predates the `db/postgres/` subdirectory creation. It is unclear whether the top-level `db/` directory was intended to eventually hold a package or only ever existed to hold the `postgres/` subdirectory. If the latter, the `.gitkeep` is permanent drift.
