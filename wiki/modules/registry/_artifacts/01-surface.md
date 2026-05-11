### 1. File tree
- `module.go` — Module wiring for registry service, HTTP handler, startup migration hook (inferred from top-level types/functions).

`api/`
- `api/gen.go` — `go:generate` entrypoint for oapi-codegen output (inferred from directive).
- `api/cfg.yaml` — oapi-codegen configuration for `registry` tagged server/models generation.
- `api/api.gen.go` — generated OpenAPI server/models for registry tag (generated file).

`application/`
- `application/service.go` — application service orchestration for controlled documents, sequence allocation, template checks, and revisions.
- `application/migration.go` — startup backfill routine from legacy `documents` rows into `controlled_documents`.
- `application/integration_test.go` — (undocumented)
- `application/migration_integration_test.go` — (undocumented)
- `application/service_test.go` — (undocumented)
- `application/tenant_isolation_test.go` — (undocumented)

`delivery/http/`
- `delivery/http/handler.go` — HTTP handler wiring and route registration to generated OpenAPI wrapper.
- `delivery/http/routes.go` — strict-server endpoint implementations and request/response mapping.
- `delivery/http/routes_contract_test.go` — (undocumented)

`domain/`
- `domain/controlled_document.go` — core controlled-document entity, statuses, and domain errors.
- `domain/document_initializer.go` — document initialization port and DTOs for atomic create/revision flows.
- `domain/port.go` — repository port and list filter contract.
- `domain/resolution.go` — template resolution rules and related domain errors.
- `domain/sequence.go` — sequence allocator and DB executor ports.
- `domain/autocode_test.go` — (undocumented)
- `domain/resolution_test.go` — (undocumented)
- `domain/sequence_bench_test.go` — (undocumented)
- `domain/sequence_test.go` — (undocumented)

`infrastructure/`
- `infrastructure/repository.go` — PostgreSQL adapters for repository, sequence allocation, template-state checks, and taxonomy readers.

### 2. Public surface
| File:line | Kind | Name | Signature / receiver | Doc comment first line |
|---|---|---|---|---|
| internal/modules/registry/module.go:15 | type | Module | `struct` | (undocumented) |
| internal/modules/registry/module.go:20 | type | Dependencies | `struct` | (undocumented) |
| internal/modules/registry/module.go:25 | func | New | `func New(deps Dependencies) *Module` | (undocumented) |
| internal/modules/registry/module.go:37 | method | RegisterRoutes | `func (m *Module) RegisterRoutes(mux *http.ServeMux)` | (undocumented) |
| internal/modules/registry/module.go:41 | method | RunStartupMigrations | `func (m *Module) RunStartupMigrations(ctx context.Context, db *sql.DB, logger *slog.Logger) error` | (undocumented) |
| internal/modules/registry/module.go:45 | method | Service | `func (m *Module) Service() *application.RegistryService` | (undocumented) |
| internal/modules/registry/application/service.go:16 | type | TemplateVersionChecker | `interface` | (undocumented) |
| internal/modules/registry/application/service.go:20 | type | ProfileReader | `interface` | (undocumented) |
| internal/modules/registry/application/service.go:24 | type | AreaReader | `interface` | (undocumented) |
| internal/modules/registry/application/service.go:28 | type | ControlledDocument | `type ControlledDocument = registrydomain.ControlledDocument` | (undocumented) |
| internal/modules/registry/application/service.go:29 | type | CDFilter | `type CDFilter = registrydomain.CDFilter` | (undocumented) |
| internal/modules/registry/application/service.go:31 | type | RegistryService | `struct` | (undocumented) |
| internal/modules/registry/application/service.go:43 | type | CreateControlledDocumentCmd | `struct` | (undocumented) |
| internal/modules/registry/application/service.go:63 | type | CreateResult | `struct` | CreateResult is the atomic-create return: the persisted ControlledDocument |
| internal/modules/registry/application/service.go:68 | func | NewRegistryService | `func NewRegistryService(... ) *RegistryService` | (undocumented) |
| internal/modules/registry/application/service.go:99 | method | WithDocumentInitializer | `func (s *RegistryService) WithDocumentInitializer(d registrydomain.DocumentInitializer) *RegistryService` | WithDocumentInitializer wires the DocumentInitializer adapter post-construction. |
| internal/modules/registry/application/service.go:104 | method | Create | `func (s *RegistryService) Create(ctx context.Context, cmd CreateControlledDocumentCmd) (*CreateResult, error)` | (undocumented) |
| internal/modules/registry/application/service.go:279 | method | PreviewCode | `func (s *RegistryService) PreviewCode(ctx context.Context, tenantID, profileCode, areaCode string) (string, error)` | PreviewCode returns the next auto-allocated CD code for (profile, area) |
| internal/modules/registry/application/service.go:289 | method | PeekSeq | `func (s *RegistryService) PeekSeq(ctx context.Context, tenantID, profileCode, areaCode string) (int, error)` | PeekSeq returns the next sequence number that NextAndIncrement would |
| internal/modules/registry/application/service.go:293 | method | Obsolete | `func (s *RegistryService) Obsolete(ctx context.Context, tenantID, controlledDocumentID string) error` | (undocumented) |
| internal/modules/registry/application/service.go:297 | method | Supersede | `func (s *RegistryService) Supersede(ctx context.Context, tenantID, controlledDocumentID string) error` | (undocumented) |
| internal/modules/registry/application/service.go:301 | method | List | `func (s *RegistryService) List(ctx context.Context, tenantID string, filter CDFilter) ([]ControlledDocument, error)` | (undocumented) |
| internal/modules/registry/application/service.go:305 | method | Get | `func (s *RegistryService) Get(ctx context.Context, tenantID, id string) (*ControlledDocument, error)` | (undocumented) |
| internal/modules/registry/application/service.go:320 | type | CreateRevisionCmd | `struct` | (undocumented) |
| internal/modules/registry/application/service.go:330 | method | CreateRevision | `func (s *RegistryService) CreateRevision(ctx context.Context, cmd CreateRevisionCmd) (*registrydomain.DocumentRef, error)` | CreateRevision creates a new document revision for an existing controlled |
| internal/modules/registry/application/migration.go:13 | func | BackfillLegacyDocuments | `func BackfillLegacyDocuments(ctx context.Context, db *sql.DB, logger *slog.Logger) error` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:10 | type | CDStatus | `type CDStatus string` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:13 | const | CDStatusActive | `const CDStatusActive CDStatus = "active"` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:14 | const | CDStatusObsolete | `const CDStatusObsolete CDStatus = "obsolete"` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:15 | const | CDStatusSuperseded | `const CDStatusSuperseded CDStatus = "superseded"` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:18 | type | ControlledDocument | `struct` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:35 | var | ErrCDNotFound | `var ErrCDNotFound error` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:36 | var | ErrCDCodeTaken | `var ErrCDCodeTaken error` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:37 | var | ErrCDArchivedCodeReuse | `var ErrCDArchivedCodeReuse error` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:38 | var | ErrSequenceCounterNotFound | `var ErrSequenceCounterNotFound error` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:39 | var | ErrCDNotActive | `var ErrCDNotActive error` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:40 | var | ErrManualCodeReasonRequired | `var ErrManualCodeReasonRequired error` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:41 | var | ErrOverrideReasonRequired | `var ErrOverrideReasonRequired error` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:44 | method | IsActive | `func (d ControlledDocument) IsActive() bool` | (undocumented) |
| internal/modules/registry/domain/controlled_document.go:48 | func | AutoCode | `func AutoCode(profileCode, areaCode string, seq int) string` | (undocumented) |
| internal/modules/registry/domain/document_initializer.go:11 | type | CloneTemplateRequest | `struct` | CloneTemplateRequest carries the user-supplied bits of an atomic CD-create |
| internal/modules/registry/domain/document_initializer.go:20 | type | DocumentRef | `struct` | DocumentRef is the minimal handle the registry returns to callers after a |
| internal/modules/registry/domain/document_initializer.go:30 | type | DocumentInitializer | `interface` | DocumentInitializer is the registry-owned port that the documents module |
| internal/modules/registry/domain/port.go:9 | type | ControlledDocumentRepository | `interface` | (undocumented) |
| internal/modules/registry/domain/port.go:19 | type | CDFilter | `struct` | (undocumented) |
| internal/modules/registry/domain/resolution.go:5 | type | TemplateResolutionInput | `struct` | (undocumented) |
| internal/modules/registry/domain/resolution.go:11 | type | TemplateVersionCandidate | `struct` | (undocumented) |
| internal/modules/registry/domain/resolution.go:17 | type | TemplateResolutionResult | `struct` | (undocumented) |
| internal/modules/registry/domain/resolution.go:23 | var | ErrProfileHasNoDefaultTemplate | `var ErrProfileHasNoDefaultTemplate error` | (undocumented) |
| internal/modules/registry/domain/resolution.go:24 | var | ErrOverrideTemplateDeleted | `var ErrOverrideTemplateDeleted error` | (undocumented) |
| internal/modules/registry/domain/resolution.go:25 | var | ErrOverrideNotPublished | `var ErrOverrideNotPublished error` | (undocumented) |
| internal/modules/registry/domain/resolution.go:26 | var | ErrDefaultObsolete | `var ErrDefaultObsolete error` | (undocumented) |
| internal/modules/registry/domain/resolution.go:27 | var | ErrTemplateProfileMismatch | `var ErrTemplateProfileMismatch error` | (undocumented) |
| internal/modules/registry/domain/resolution.go:30 | func | Resolve | `func Resolve(in TemplateResolutionInput) (TemplateResolutionResult, error)` | (undocumented) |
| internal/modules/registry/domain/sequence.go:8 | type | DBExecutor | `interface` | (undocumented) |
| internal/modules/registry/domain/sequence.go:13 | type | SequenceAllocator | `interface` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:18 | type | PostgresControlledDocumentRepository | `struct` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:22 | func | NewPostgresControlledDocumentRepository | `func NewPostgresControlledDocumentRepository(db *sql.DB) *PostgresControlledDocumentRepository` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:26 | method | GetByID | `func (r *PostgresControlledDocumentRepository) GetByID(ctx context.Context, tenantID, id string) (*registrydomain.ControlledDocument, error)` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:36 | method | GetByCode | `func (r *PostgresControlledDocumentRepository) GetByCode(ctx context.Context, tenantID, profileCode, code string) (*registrydomain.ControlledDocument, error)` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:46 | method | CodeExists | `func (r *PostgresControlledDocumentRepository) CodeExists(ctx context.Context, tenantID, profileCode, code string) (bool, error)` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:57 | method | List | `func (r *PostgresControlledDocumentRepository) List(ctx context.Context, tenantID string, filter registrydomain.CDFilter) ([]registrydomain.ControlledDocument, error)` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:133 | method | Create | `func (r *PostgresControlledDocumentRepository) Create(ctx context.Context, doc *registrydomain.ControlledDocument) error` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:137 | method | CreateTx | `func (r *PostgresControlledDocumentRepository) CreateTx(ctx context.Context, tx *sql.Tx, doc *registrydomain.ControlledDocument) error` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:184 | method | UpdateStatus | `func (r *PostgresControlledDocumentRepository) UpdateStatus(ctx context.Context, tenantID, id string, status registrydomain.CDStatus, updatedAt time.Time) error` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:199 | type | PostgresSequenceAllocator | `struct` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:203 | func | NewPostgresSequenceAllocator | `func NewPostgresSequenceAllocator(db *sql.DB) *PostgresSequenceAllocator` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:208 | method | EnsureCounter | `func (a *PostgresSequenceAllocator) EnsureCounter(ctx context.Context, tenantID, profileCode, areaCode string) error` | EnsureCounter initialises a sequence counter for the given tenant/profile/area combination. |
| internal/modules/registry/infrastructure/repository.go:224 | method | Peek | `func (a *PostgresSequenceAllocator) Peek(ctx context.Context, tenantID, profileCode, areaCode string) (int, error)` | Peek returns the current next_seq value without incrementing it. |
| internal/modules/registry/infrastructure/repository.go:239 | method | NextAndIncrement | `func (a *PostgresSequenceAllocator) NextAndIncrement(ctx context.Context, tx registrydomain.DBExecutor, tenantID, profileCode, areaCode string) (int, error)` | NextAndIncrement atomically increments and returns the next sequence number. |
| internal/modules/registry/infrastructure/repository.go:265 | type | PostgresTemplateVersionChecker | `struct` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:269 | func | NewPostgresTemplateVersionChecker | `func NewPostgresTemplateVersionChecker(db *sql.DB) *PostgresTemplateVersionChecker` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:273 | method | GetTemplateVersionState | `func (c *PostgresTemplateVersionChecker) GetTemplateVersionState(ctx context.Context, templateVersionID string) (*string, string, error)` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:295 | type | TaxonomyProfileReader | `struct` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:299 | func | NewTaxonomyProfileReader | `func NewTaxonomyProfileReader(db *sql.DB) *TaxonomyProfileReader` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:303 | method | GetByCode | `func (r *TaxonomyProfileReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error)` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:337 | type | TaxonomyAreaReader | `struct` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:341 | func | NewTaxonomyAreaReader | `func NewTaxonomyAreaReader(db *sql.DB) *TaxonomyAreaReader` | (undocumented) |
| internal/modules/registry/infrastructure/repository.go:343 | method | GetByCode | `func (r *TaxonomyAreaReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.ProcessArea, error)` | (undocumented) |
| internal/modules/registry/delivery/http/handler.go:30 | type | Handler | `struct` | (undocumented) |
| internal/modules/registry/delivery/http/handler.go:37 | func | NewHandler | `func NewHandler(svc *application.RegistryService, db *sql.DB) *Handler` | (undocumented) |
| internal/modules/registry/delivery/http/handler.go:67 | method | RegisterRoutes | `func (h *Handler) RegisterRoutes(mux *http.ServeMux)` | (undocumented) |
| internal/modules/registry/delivery/http/routes.go:23 | method | ListControlledDocuments | `func (h *Handler) ListControlledDocuments(w http.ResponseWriter, r *http.Request, params registryapi.ListControlledDocumentsParams)` | (undocumented) |
| internal/modules/registry/delivery/http/routes.go:43 | method | AtomicCreateControlledDocument | `func (h *Handler) AtomicCreateControlledDocument(w http.ResponseWriter, r *http.Request, params registryapi.AtomicCreateControlledDocumentParams)` | (undocumented) |
| internal/modules/registry/delivery/http/routes.go:127 | method | PreviewControlledDocumentCode | `func (h *Handler) PreviewControlledDocumentCode(w http.ResponseWriter, r *http.Request, params registryapi.PreviewControlledDocumentCodeParams)` | (undocumented) |
| internal/modules/registry/delivery/http/routes.go:148 | method | CreateControlledDocumentRevision | `func (h *Handler) CreateControlledDocumentRevision(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params registryapi.CreateControlledDocumentRevisionParams)` | (undocumented) |
| internal/modules/registry/delivery/http/routes.go:190 | method | GetControlledDocument | `func (h *Handler) GetControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)` | (undocumented) |
| internal/modules/registry/delivery/http/routes.go:205 | method | GetActiveDocument | `func (h *Handler) GetActiveDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)` | (undocumented) |
| internal/modules/registry/delivery/http/routes.go:328 | method | ObsoleteControlledDocument | `func (h *Handler) ObsoleteControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)` | (undocumented) |
| internal/modules/registry/delivery/http/routes.go:337 | method | SupersedeControlledDocument | `func (h *Handler) SupersedeControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)` | (undocumented) |

### 3. HTTP operations
| Method | Path | Handler symbol | Source file:line |
|---|---|---|---|
| POST | /api/v2/controlled-documents | `generated.AtomicCreateControlledDocument` | internal/modules/registry/delivery/http/handler.go:79 |
| POST | /api/v2/controlled-documents/{id}/revisions | `generated.CreateControlledDocumentRevision` | internal/modules/registry/delivery/http/handler.go:81 |
| GET | /api/v2/controlled-documents/preview-code | `generated.PreviewControlledDocumentCode` | internal/modules/registry/delivery/http/handler.go:83 |
| GET | /api/v2/controlled-documents | `generated.ListControlledDocuments` | internal/modules/registry/delivery/http/handler.go:84 |
| GET | /api/v2/controlled-documents/{id} | `generated.GetControlledDocument` | internal/modules/registry/delivery/http/handler.go:85 |
| GET | /api/v2/controlled-documents/{id}/active-document | `generated.GetActiveDocument` | internal/modules/registry/delivery/http/handler.go:86 |
| PUT | /api/v2/controlled-documents/{id}/obsolete | `generated.ObsoleteControlledDocument` | internal/modules/registry/delivery/http/handler.go:87 |
| PUT | /api/v2/controlled-documents/{id}/supersede | `generated.SupersedeControlledDocument` | internal/modules/registry/delivery/http/handler.go:88 |

### 4. Migration list
| Filename | Verb | Tables touched |
|---|---|---|
| migrations/0124_registry_controlled_documents.sql | CREATE TABLE, ALTER TABLE, CREATE INDEX, DROP TRIGGER, CREATE TRIGGER | `controlled_documents`, `profile_sequence_counters` |
| migrations/0126_documents_v2_bridge_columns.sql | ALTER TABLE (FK reference) | `controlled_documents` |
| migrations/0127_documents_v2_tenant_consistency_trigger.sql | CREATE FUNCTION/TRIGGER logic (SELECT existence check) | `controlled_documents` |
| migrations/0128_grants_new_tables.sql | GRANT | `controlled_documents`, `profile_sequence_counters` |
| migrations/0167_documents_bridge_and_state_columns.sql | ALTER TABLE (FK reference) | `controlled_documents` |
| migrations/0182_cd_sequence_per_area.sql | DROP TABLE, TRUNCATE TABLE, CREATE TABLE, GRANT | `profile_sequence_counters`, `controlled_documents`, `cd_sequence_counters` |
| migrations/0183_documents_name_not_empty.sql | UPDATE (backfill from JOIN source) | `controlled_documents` |
