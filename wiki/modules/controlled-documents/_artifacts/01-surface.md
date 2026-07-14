<!-- Last verified: 2026-06-11 | Scope: all Go source under internal/modules/controlleddocuments/ + db/migrations/ (canonical) | Verifier: adversarial-fix pass against actual source -->

### 1. File tree
- `module.go` — Module wiring for controlled-document service, HTTP handler, startup dependencies.

`api/`
- `api/gen.go` — `go:generate` entrypoint for oapi-codegen output (inferred from directive).
- `api/cfg.yaml` — oapi-codegen configuration for `registry` tagged server/models generation.
- `api/api.gen.go` — generated OpenAPI server/models for registry tag (generated file).

`application/`
- `application/service.go` — application service orchestration for controlled documents, sequence allocation, template checks, and revisions.
- `application/integration_test.go` — (undocumented)
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
| internal/modules/controlleddocuments/module.go:16 | type | Module | `struct` | (undocumented) |
| internal/modules/controlleddocuments/module.go:21 | type | Dependencies | `struct` | (undocumented) |
| internal/modules/controlleddocuments/module.go:27 | func | New | `func New(deps Dependencies) *Module` | (undocumented) |
| internal/modules/controlleddocuments/module.go:50 | method | RegisterRoutes | `func (m *Module) RegisterRoutes(mux *http.ServeMux)` | (undocumented) |
| internal/modules/controlleddocuments/module.go:54 | method | Service | `func (m *Module) Service() *application.ControlledDocumentService` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:23 | type | TemplateVersionChecker | `interface` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:27 | type | ProfileReader | `interface` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:31 | type | AreaReader | `interface` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:35 | type | ControlledDocument | `type ControlledDocument = controlleddocumentsdomain.ControlledDocument` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:36 | type | CDFilter | `type CDFilter = controlleddocumentsdomain.CDFilter` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:38 | type | ControlledDocumentService | `struct` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:53 | type | CreateControlledDocumentCmd | `struct` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:75 | type | CreateResult | `struct` | CreateResult is the atomic-create return: the persisted ControlledDocument |
| internal/modules/controlleddocuments/application/service.go:89 | func | NewControlledDocumentService | `func NewControlledDocumentService(db *sql.DB, docs controlleddocumentsdomain.ControlledDocumentRepository, seq controlleddocumentsdomain.SequenceAllocator, tplCheck TemplateVersionChecker, profiles ProfileReader, areas AreaReader, govLogger taxonomydomain.GovernanceLogger, docInit controlleddocumentsdomain.DocumentInitializer) *ControlledDocumentService` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:138 | method | WithDocumentInitializer | `func (s *ControlledDocumentService) WithDocumentInitializer(d controlleddocumentsdomain.DocumentInitializer) *ControlledDocumentService` | WithDocumentInitializer wires the DocumentInitializer adapter post-construction. |
| internal/modules/controlleddocuments/application/service.go:146 | method | Create | `func (s *ControlledDocumentService) Create(ctx context.Context, cmd CreateControlledDocumentCmd) (*CreateResult, error)` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:389 | method | PreviewCode | `func (s *ControlledDocumentService) PreviewCode(ctx context.Context, tenantID, profileCode, areaCode string) (string, error)` | PreviewCode returns the next auto-allocated CD code for (profile, area) |
| internal/modules/controlleddocuments/application/service.go:399 | method | PeekSeq | `func (s *ControlledDocumentService) PeekSeq(ctx context.Context, tenantID, profileCode, areaCode string) (int, error)` | PeekSeq returns the next sequence number that NextAndIncrement would |
| internal/modules/controlleddocuments/application/service.go:451 | method | Obsolete | `func (s *ControlledDocumentService) Obsolete(ctx context.Context, tenantID, controlledDocumentID string) error` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:455 | method | Supersede | `func (s *ControlledDocumentService) Supersede(ctx context.Context, tenantID, controlledDocumentID string) error` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:459 | method | List | `func (s *ControlledDocumentService) List(ctx context.Context, tenantID string, filter CDFilter) ([]ControlledDocument, bool, error)` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:472 | method | Get | `func (s *ControlledDocumentService) Get(ctx context.Context, tenantID, id string) (*ControlledDocument, error)` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:574 | type | CreateRevisionCmd | `struct` | (undocumented) |
| internal/modules/controlleddocuments/application/service.go:584 | method | CreateRevision | `func (s *ControlledDocumentService) CreateRevision(ctx context.Context, cmd CreateRevisionCmd) (*controlleddocumentsdomain.DocumentRef, error)` | CreateRevision creates a new document revision for an existing controlled |
| internal/modules/controlleddocuments/domain/controlled_document.go:10 | type | CDStatus | `type CDStatus string` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:13 | const | CDStatusActive | `const CDStatusActive CDStatus = "active"` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:14 | const | CDStatusObsolete | `const CDStatusObsolete CDStatus = "obsolete"` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:15 | const | CDStatusSuperseded | `const CDStatusSuperseded CDStatus = "superseded"` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:18 | type | ControlledDocument | `struct` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:36 | var | ErrCDNotFound | `var ErrCDNotFound error` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:37 | var | ErrCDCodeTaken | `var ErrCDCodeTaken error` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:38 | var | ErrCDArchivedCodeReuse | `var ErrCDArchivedCodeReuse error` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:39 | var | ErrSequenceCounterNotFound | `var ErrSequenceCounterNotFound error` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:40 | var | ErrCDNotActive | `var ErrCDNotActive error` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:41 | var | ErrActiveRevisionExists | `var ErrActiveRevisionExists error` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:42 | var | ErrManualCodeReasonRequired | `var ErrManualCodeReasonRequired error` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:43 | var | ErrOverrideReasonRequired | `var ErrOverrideReasonRequired error` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:101 | method | IsActive | `func (d ControlledDocument) IsActive() bool` | (undocumented) |
| internal/modules/controlleddocuments/domain/controlled_document.go:105 | func | AutoCode | `func AutoCode(profileCode, areaCode string, seq int) string` | (undocumented) |
| internal/modules/controlleddocuments/domain/document_initializer.go:19 | type | CloneTemplateRequest | `struct` | CloneTemplateRequest carries the user-supplied bits of an atomic CD-create |
| internal/modules/controlleddocuments/domain/document_initializer.go:51 | type | DocumentRef | `struct` | DocumentRef is the minimal handle the registry returns to callers after a |
| internal/modules/controlleddocuments/domain/document_initializer.go:61 | type | DocumentInitializer | `interface` | DocumentInitializer is the controlled-documents-owned port that the documents module |
| internal/modules/controlleddocuments/domain/port.go:8 | type | ControlledDocumentRepository | `interface` | (undocumented) |
| internal/modules/controlleddocuments/domain/port.go:23 | type | CDFilter | `struct` | (undocumented) |
| internal/modules/controlleddocuments/domain/resolution.go:5 | type | TemplateResolutionInput | `struct` | (undocumented) |
| internal/modules/controlleddocuments/domain/resolution.go:11 | type | TemplateVersionCandidate | `struct` | (undocumented) |
| internal/modules/controlleddocuments/domain/resolution.go:17 | type | TemplateResolutionResult | `struct` | (undocumented) |
| internal/modules/controlleddocuments/domain/resolution.go:23 | var | ErrProfileHasNoDefaultTemplate | `var ErrProfileHasNoDefaultTemplate error` | (undocumented) |
| internal/modules/controlleddocuments/domain/resolution.go:24 | var | ErrOverrideTemplateDeleted | `var ErrOverrideTemplateDeleted error` | (undocumented) |
| internal/modules/controlleddocuments/domain/resolution.go:25 | var | ErrOverrideNotPublished | `var ErrOverrideNotPublished error` | (undocumented) |
| internal/modules/controlleddocuments/domain/resolution.go:26 | var | ErrDefaultObsolete | `var ErrDefaultObsolete error` | (undocumented) |
| internal/modules/controlleddocuments/domain/resolution.go:27 | var | ErrTemplateProfileMismatch | `var ErrTemplateProfileMismatch error` | (undocumented) |
| internal/modules/controlleddocuments/domain/resolution.go:30 | func | Resolve | `func Resolve(in TemplateResolutionInput) (TemplateResolutionResult, error)` | (undocumented) |
| internal/modules/controlleddocuments/domain/sequence.go:8 | type | DBTX | `interface` | (undocumented) |
| internal/modules/controlleddocuments/domain/sequence.go:14 | type | DBExecutor | `type DBExecutor = DBTX` | (undocumented) |
| internal/modules/controlleddocuments/domain/sequence.go:16 | type | SequenceAllocator | `interface` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:22 | type | PostgresControlledDocumentRepository | `struct` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:26 | func | NewPostgresControlledDocumentRepository | `func NewPostgresControlledDocumentRepository(db *sql.DB) *PostgresControlledDocumentRepository` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:30 | method | GetByID | `func (r *PostgresControlledDocumentRepository) GetByID(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error)` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:51 | method | GetByCode | `func (r *PostgresControlledDocumentRepository) GetByCode(ctx context.Context, tenantID, profileCode, code string) (*controlleddocumentsdomain.ControlledDocument, error)` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:72 | method | CodeExists | `func (r *PostgresControlledDocumentRepository) CodeExists(ctx context.Context, tenantID, profileCode, code string) (bool, error)` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:87 | method | List | `func (r *PostgresControlledDocumentRepository) List(ctx context.Context, tenantID string, filter controlleddocumentsdomain.CDFilter) (items []controlleddocumentsdomain.ControlledDocument, hasMore bool, err error)` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:333 | method | Create | `func (r *PostgresControlledDocumentRepository) Create(ctx context.Context, doc *controlleddocumentsdomain.ControlledDocument) error` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:353 | method | CreateTx | `func (r *PostgresControlledDocumentRepository) CreateTx(ctx context.Context, tx controlleddocumentsdomain.DBTX, doc *controlleddocumentsdomain.ControlledDocument) error` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:432 | method | UpdateStatus | `func (r *PostgresControlledDocumentRepository) UpdateStatus(ctx context.Context, tenantID, id string, status controlleddocumentsdomain.CDStatus, updatedAt time.Time) error` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:450 | method | UpdateStatusTx | `func (r *PostgresControlledDocumentRepository) UpdateStatusTx(ctx context.Context, tx controlleddocumentsdomain.DBTX, tenantID, id string, status controlleddocumentsdomain.CDStatus, updatedAt time.Time) error` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:468 | method | CanRead | `func (r *PostgresControlledDocumentRepository) CanRead(ctx context.Context, tenantID, controlledDocumentID, actorUserID string) (bool, error)` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:513 | type | PostgresSequenceAllocator | `struct` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:517 | func | NewPostgresSequenceAllocator | `func NewPostgresSequenceAllocator(db *sql.DB) *PostgresSequenceAllocator` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:522 | method | EnsureCounter | `func (a *PostgresSequenceAllocator) EnsureCounter(ctx context.Context, tenantID, profileCode, areaCode string) error` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:541 | method | Peek | `func (a *PostgresSequenceAllocator) Peek(ctx context.Context, tenantID, profileCode, areaCode string) (int, error)` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:559 | method | NextAndIncrement | `func (a *PostgresSequenceAllocator) NextAndIncrement(ctx context.Context, tx controlleddocumentsdomain.DBTX, tenantID, profileCode, areaCode string) (int, error)` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:592 | type | PostgresTemplateVersionChecker | `struct` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:596 | func | NewPostgresTemplateVersionChecker | `func NewPostgresTemplateVersionChecker(db *sql.DB) *PostgresTemplateVersionChecker` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:600 | method | GetTemplateVersionState | `func (c *PostgresTemplateVersionChecker) GetTemplateVersionState(ctx context.Context, tenantID, templateVersionID string) (*string, string, error)` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:623 | type | TaxonomyProfileReader | `struct` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:627 | func | NewTaxonomyProfileReader | `func NewTaxonomyProfileReader(db *sql.DB) *TaxonomyProfileReader` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:631 | method | GetByCode | `func (r *TaxonomyProfileReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error)` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:665 | type | TaxonomyAreaReader | `struct` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:669 | func | NewTaxonomyAreaReader | `func NewTaxonomyAreaReader(db *sql.DB) *TaxonomyAreaReader` | (undocumented) |
| internal/modules/controlleddocuments/infrastructure/repository.go:671 | method | GetByCode | `func (r *TaxonomyAreaReader) GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.ProcessArea, error)` | (undocumented) |
| internal/modules/controlleddocuments/delivery/http/handler.go:30 | type | Handler | `struct` | (undocumented) |
| internal/modules/controlleddocuments/delivery/http/handler.go:37 | func | NewHandler | `func NewHandler(svc *application.ControlledDocumentService, db *sql.DB) *Handler` | (undocumented) |
| internal/modules/controlleddocuments/delivery/http/handler.go:68 | method | RegisterRoutes | `func (h *Handler) RegisterRoutes(mux *http.ServeMux)` | (undocumented) |
| internal/modules/controlleddocuments/delivery/http/routes.go:30 | method | ListControlledDocuments | `func (h *Handler) ListControlledDocuments(w http.ResponseWriter, r *http.Request, params controlleddocumentsapi.ListControlledDocumentsParams)` | (undocumented) |
| internal/modules/controlleddocuments/delivery/http/routes.go:70 | method | AtomicCreateControlledDocument | `func (h *Handler) AtomicCreateControlledDocument(w http.ResponseWriter, r *http.Request, params controlleddocumentsapi.AtomicCreateControlledDocumentParams)` | (undocumented) |
| internal/modules/controlleddocuments/delivery/http/routes.go:172 | method | PreviewControlledDocumentCode | `func (h *Handler) PreviewControlledDocumentCode(w http.ResponseWriter, r *http.Request, params controlleddocumentsapi.PreviewControlledDocumentCodeParams)` | (undocumented) |
| internal/modules/controlleddocuments/delivery/http/routes.go:197 | method | CreateControlledDocumentRevision | `func (h *Handler) CreateControlledDocumentRevision(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params controlleddocumentsapi.CreateControlledDocumentRevisionParams)` | (undocumented) |
| internal/modules/controlleddocuments/delivery/http/routes.go:245 | method | GetControlledDocument | `func (h *Handler) GetControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)` | (undocumented) |
| internal/modules/controlleddocuments/delivery/http/routes.go:266 | method | GetActiveDocument | `func (h *Handler) GetActiveDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)` | (undocumented) |
| internal/modules/controlleddocuments/delivery/http/routes.go:441 | method | ObsoleteControlledDocument | `func (h *Handler) ObsoleteControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)` | (undocumented) |
| internal/modules/controlleddocuments/delivery/http/routes.go:455 | method | SupersedeControlledDocument | `func (h *Handler) SupersedeControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID)` | (undocumented) |

### 3. HTTP operations
Routes are registered via a single generated `controlleddocumentsapi.HandlerWithOptions` call at `internal/modules/controlleddocuments/delivery/http/handler.go:95`. The handler struct (`Handler`) implements the generated `StrictServerInterface`; each method below is the corresponding implementation in `delivery/http/routes.go`.

| Method | Path | Handler symbol | Source file:line |
|---|---|---|---|
| POST | /api/v1/controlled-documents | `Handler.AtomicCreateControlledDocument` | internal/modules/controlleddocuments/delivery/http/routes.go:70 |
| POST | /api/v1/controlled-documents/{id}/revisions | `Handler.CreateControlledDocumentRevision` | internal/modules/controlleddocuments/delivery/http/routes.go:197 |
| GET | /api/v1/controlled-documents/preview-code | `Handler.PreviewControlledDocumentCode` | internal/modules/controlleddocuments/delivery/http/routes.go:172 |
| GET | /api/v1/controlled-documents | `Handler.ListControlledDocuments` | internal/modules/controlleddocuments/delivery/http/routes.go:30 |
| GET | /api/v1/controlled-documents/{id} | `Handler.GetControlledDocument` | internal/modules/controlleddocuments/delivery/http/routes.go:245 |
| GET | /api/v1/controlled-documents/{id}/active-document | `Handler.GetActiveDocument` | internal/modules/controlleddocuments/delivery/http/routes.go:266 |
| PUT | /api/v1/controlled-documents/{id}/obsolete | `Handler.ObsoleteControlledDocument` | internal/modules/controlleddocuments/delivery/http/routes.go:441 |
| PUT | /api/v1/controlled-documents/{id}/supersede | `Handler.SupersedeControlledDocument` | internal/modules/controlleddocuments/delivery/http/routes.go:455 |

### 4. Migration list
> **Note:** Migrations 0001–0202 have been consolidated into a curated baseline. The canonical `db/migrations/` directory begins at **0203**. Migrations 0124–0183 listed in earlier versions of this document exist only in `archive/migrations/` and are no longer applied by the current migration runner. The table below lists only canonical (active) migrations that affect tables owned or consumed by this module.

| Filename | Verb | Tables touched |
|---|---|---|
| db/migrations/0210_controlled_documents_capability_namespace.sql | (capability namespace update) | capability/authz tables scoped to controlled documents |
| db/migrations/0225_authz_p2_document_lifecycle_grants.sql | GRANT | document lifecycle capability grants |
| db/migrations/0229_authz_p12_rename_document_lifecycle_caps.sql | UPDATE/RENAME | capability names for document lifecycle |
| db/migrations/0232_drop_document_access_policies.sql | DROP | document access policy tables (legacy policy rows) |

> **Legacy reference (archive only):** The original `controlled_documents` and `cd_sequence_counters` tables, triggers, and grants were introduced in `archive/migrations/0124_registry_controlled_documents.sql` through `archive/migrations/0183_documents_name_not_empty.sql`. These are historical record only; the schema they created is now part of the curated baseline.
