# Phase 1 — Surface scan (taxonomy)

Total exported symbols: 80 · HTTP operations: 16 · migrations touching taxonomy tables: 19

## 1. File tree

```
internal/modules/taxonomy/
├── module.go                                  — composition root, wires services + handler
├── application/
│   ├── area_service.go                        — AreaService use cases (list/get/create/update/set-parent/archive)
│   ├── area_service_test.go                   — AreaService parent-cycle + archive tests
│   ├── family_service.go                      — FamilyService use cases (list/get/create/update/deactivate)
│   ├── family_service_test.go                 — FamilyService CRUD/deactivation tests
│   ├── governance_logger.go                   — DB-backed governance event logger (legacy sink)
│   ├── immutability_test.go                   — code immutability tests
│   ├── profile_service.go                     — ProfileService + TemplateVersionChecker iface
│   └── profile_service_test.go                — ProfileService template-binding + archive tests
├── delivery/http/
│   ├── handler.go                             — Handler struct, route registration on net/http ServeMux
│   ├── routes_areas.go                        — area HTTP handlers + error mapping
│   ├── routes_families.go                     — family HTTP handlers + error mapping
│   ├── routes_families_contract_test.go       — family contract tests
│   ├── routes_profiles.go                     — profile HTTP handlers + error mapping
│   └── routes_profiles_contract_test.go       — profile contract tests
├── domain/
│   ├── area.go                                — ProcessArea entity + sentinel errors + IsActive/Archive
│   ├── area_test.go
│   ├── family.go                              — DocumentFamily entity + sentinel errors + Deactivate
│   ├── family_test.go
│   ├── port.go                                — Repository ifaces + GovernanceLogger iface + GovernanceEvent
│   ├── profile.go                             — DocumentProfile entity + sentinel errors + IsActive/Archive
│   └── profile_test.go
└── infrastructure/
    ├── family_repository.go                   — SQL impl FamilyRepository (incl HasActiveProfiles)
    ├── repository.go                          — SQL impl ProfileRepository + AreaRepository (+ ListAncestors)
    ├── template_version_checker.go            — SQL impl IsPublished(template_version_id) → bool, profile_code
    └── template_version_checker_test.go
```

## 2. Public surface

| File:line | Kind | Name | Signature / receiver | Doc |
|---|---|---|---|---|
| internal/modules/taxonomy/module.go:12 | type | Module | `type Module struct` | (undocumented) |
| internal/modules/taxonomy/module.go:16 | type | Dependencies | `type Dependencies struct` | (undocumented) |
| internal/modules/taxonomy/module.go:21 | func | New | `func New(deps Dependencies) *Module` | (undocumented) |
| internal/modules/taxonomy/module.go:36 | method | RegisterRoutes | `func (m *Module) RegisterRoutes(mux *http.ServeMux)` | (undocumented) |
| internal/modules/taxonomy/application/area_service.go:10 | type | AreaService | `type AreaService struct` | (undocumented) |
| internal/modules/taxonomy/application/area_service.go:16 | func | NewAreaService | `func NewAreaService(areas domain.AreaRepository, govLogger domain.GovernanceLogger) *AreaService` | (undocumented) |
| internal/modules/taxonomy/application/area_service.go:24 | method | List | `func (s *AreaService) List(ctx, tenantID, includeArchived) ([]ProcessArea, error)` | (undocumented) |
| internal/modules/taxonomy/application/area_service.go:28 | method | Get | `func (s *AreaService) Get(ctx, tenantID, code) (*ProcessArea, error)` | (undocumented) |
| internal/modules/taxonomy/application/area_service.go:32 | method | Create | `func (s *AreaService) Create(ctx, *ProcessArea) error` | (undocumented) |
| internal/modules/taxonomy/application/area_service.go:36 | method | Update | `func (s *AreaService) Update(ctx, *ProcessArea) error` | (undocumented) |
| internal/modules/taxonomy/application/area_service.go:40 | method | SetParent | `func (s *AreaService) SetParent(ctx, tenantID, areaCode, parentCode *string, actorID) error` | (undocumented) |
| internal/modules/taxonomy/application/area_service.go:81 | method | Archive | `func (s *AreaService) Archive(ctx, tenantID, areaCode, actorID) error` | (undocumented) |
| internal/modules/taxonomy/application/family_service.go:11 | type | FamilyService | `type FamilyService struct` | (undocumented) |
| internal/modules/taxonomy/application/family_service.go:15 | func | NewFamilyService | `func NewFamilyService(families domain.FamilyRepository) *FamilyService` | (undocumented) |
| internal/modules/taxonomy/application/family_service.go:19 | method | List | `func (s *FamilyService) List(ctx, includeInactive) ([]DocumentFamily, error)` | (undocumented) |
| internal/modules/taxonomy/application/family_service.go:23 | method | Get | `func (s *FamilyService) Get(ctx, code) (*DocumentFamily, error)` | (undocumented) |
| internal/modules/taxonomy/application/family_service.go:27 | method | Create | `func (s *FamilyService) Create(ctx, *DocumentFamily) error` | (undocumented) |
| internal/modules/taxonomy/application/family_service.go:32 | method | Update | `func (s *FamilyService) Update(ctx, *DocumentFamily) (*DocumentFamily, error)` | (undocumented) |
| internal/modules/taxonomy/application/family_service.go:48 | method | Deactivate | `func (s *FamilyService) Deactivate(ctx, code) error` | (undocumented) |
| internal/modules/taxonomy/application/profile_service.go:11 | type | TemplateVersionChecker | `type TemplateVersionChecker interface { IsPublished(ctx, versionID) (bool, string, error) }` | (undocumented) |
| internal/modules/taxonomy/application/profile_service.go:15 | type | ProfileService | `type ProfileService struct` | (undocumented) |
| internal/modules/taxonomy/application/profile_service.go:22 | func | NewProfileService | `func NewProfileService(profiles, tplCheck, govLogger) *ProfileService` | (undocumented) |
| internal/modules/taxonomy/application/profile_service.go:33 | method | List | `func (s *ProfileService) List(ctx, tenantID, includeArchived) ([]DocumentProfile, error)` | (undocumented) |
| internal/modules/taxonomy/application/profile_service.go:37 | method | Get | `func (s *ProfileService) Get(ctx, tenantID, code) (*DocumentProfile, error)` | (undocumented) |
| internal/modules/taxonomy/application/profile_service.go:41 | method | Create | `func (s *ProfileService) Create(ctx, *DocumentProfile) error` | (undocumented) |
| internal/modules/taxonomy/application/profile_service.go:45 | method | Update | `func (s *ProfileService) Update(ctx, *DocumentProfile) error` | (undocumented) |
| internal/modules/taxonomy/application/profile_service.go:49 | method | SetDefaultTemplate | `func (s *ProfileService) SetDefaultTemplate(ctx, tenantID, profileCode, templateVersionID, actorID) error` | (undocumented) |
| internal/modules/taxonomy/application/profile_service.go:87 | method | Archive | `func (s *ProfileService) Archive(ctx, tenantID, profileCode, actorID) error` | (undocumented) |
| internal/modules/taxonomy/application/governance_logger.go:10 | type | DBGovernanceLogger | `type DBGovernanceLogger struct` | (undocumented) |
| internal/modules/taxonomy/application/governance_logger.go:14 | func | NewDBGovernanceLogger | `func NewDBGovernanceLogger(db *sql.DB) *DBGovernanceLogger` | (undocumented) |
| internal/modules/taxonomy/application/governance_logger.go:18 | method | Log | `func (l *DBGovernanceLogger) Log(ctx, GovernanceEvent) error` | (undocumented) |
| internal/modules/taxonomy/domain/area.go:8 | type | ProcessArea | struct fields: Code,TenantID,Name,Description,ParentCode,OwnerUserID,DefaultApproverRole,ArchivedAt,CreatedAt | (undocumented) |
| internal/modules/taxonomy/domain/area.go:21 | var | ErrAreaNotFound | `errors.New("process area not found")` | (undocumented) |
| internal/modules/taxonomy/domain/area.go:22 | var | ErrAreaArchived | `errors.New("process area is archived")` | (undocumented) |
| internal/modules/taxonomy/domain/area.go:23 | var | ErrAreaParentCycle | `errors.New("area parent assignment creates cycle")` | (undocumented) |
| internal/modules/taxonomy/domain/area.go:24 | var | ErrAreaCodeImmutable | `errors.New("area code is immutable")` | (undocumented) |
| internal/modules/taxonomy/domain/area.go:27 | method | IsActive | `func (a *ProcessArea) IsActive() bool` | (undocumented) |
| internal/modules/taxonomy/domain/area.go:31 | method | Archive | `func (a *ProcessArea) Archive(now time.Time) error` | (undocumented) |
| internal/modules/taxonomy/domain/family.go:8 | type | DocumentFamily | struct fields: Code,Name,Description,IsActive,CreatedAt | (undocumented) |
| internal/modules/taxonomy/domain/family.go:17 | var | ErrFamilyNotFound | `errors.New("family not found")` | (undocumented) |
| internal/modules/taxonomy/domain/family.go:18 | var | ErrFamilyAlreadyInactive | `errors.New("family is already inactive")` | (undocumented) |
| internal/modules/taxonomy/domain/family.go:19 | var | ErrFamilyHasProfiles | `errors.New("family has active profiles and cannot be deactivated")` | (undocumented) |
| internal/modules/taxonomy/domain/family.go:22 | method | Deactivate | `func (f *DocumentFamily) Deactivate() error` | (undocumented) |
| internal/modules/taxonomy/domain/profile.go:8 | type | DocumentProfile | struct fields: Code,TenantID,FamilyCode,Name,Description,Alias,ReviewIntervalDays,DefaultTemplateVersionID,OwnerUserID,EditableByRole,ArchivedAt,CreatedAt | (undocumented) |
| internal/modules/taxonomy/domain/profile.go:24 | var | ErrProfileNotFound | `errors.New("profile not found")` | (undocumented) |
| internal/modules/taxonomy/domain/profile.go:25 | var | ErrProfileCodeImmutable | `errors.New("profile code is immutable")` | (undocumented) |
| internal/modules/taxonomy/domain/profile.go:26 | var | ErrProfileArchived | `errors.New("profile is archived")` | (undocumented) |
| internal/modules/taxonomy/domain/profile.go:27 | var | ErrTemplateNotPublished | `errors.New("template version is not published")` | (undocumented) |
| internal/modules/taxonomy/domain/profile.go:28 | var | ErrTemplateProfileMismatch | `errors.New("template version belongs to different profile")` | (undocumented) |
| internal/modules/taxonomy/domain/profile.go:31 | method | IsActive | `func (p *DocumentProfile) IsActive() bool` | (undocumented) |
| internal/modules/taxonomy/domain/profile.go:35 | method | Archive | `func (p *DocumentProfile) Archive(now) error` | (undocumented) |
| internal/modules/taxonomy/domain/port.go:5 | type | ProfileRepository | iface: GetByCode, List, Create, Update | (undocumented) |
| internal/modules/taxonomy/domain/port.go:12 | type | AreaRepository | iface: GetByCode, List, Create, Update, ListAncestors | (undocumented) |
| internal/modules/taxonomy/domain/port.go:20 | type | GovernanceLogger | iface: Log(ctx, GovernanceEvent) error | (undocumented) |
| internal/modules/taxonomy/domain/port.go:24 | type | GovernanceEvent | struct: TenantID, EventType, ActorUserID, ResourceType, ResourceID, Reason, PayloadJSON | (undocumented) |
| internal/modules/taxonomy/domain/port.go:34 | type | FamilyRepository | iface: GetByCode, List, Create, Update, HasActiveProfiles | (undocumented) |
| internal/modules/taxonomy/infrastructure/family_repository.go:11 | type | FamilyRepository | struct | (undocumented) |
| internal/modules/taxonomy/infrastructure/family_repository.go:15 | func | NewFamilyRepository | `func NewFamilyRepository(db *sql.DB) *FamilyRepository` | (undocumented) |
| internal/modules/taxonomy/infrastructure/family_repository.go:19 | method | GetByCode | SQL by code | (undocumented) |
| internal/modules/taxonomy/infrastructure/family_repository.go:38 | method | List | SQL list (filter is_active) | (undocumented) |
| internal/modules/taxonomy/infrastructure/family_repository.go:67 | method | Create | INSERT into document_families | (undocumented) |
| internal/modules/taxonomy/infrastructure/family_repository.go:75 | method | Update | UPDATE document_families | (undocumented) |
| internal/modules/taxonomy/infrastructure/family_repository.go:91 | method | HasActiveProfiles | EXISTS check on document_profiles | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:11 | type | ProfileRepository | struct | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:15 | func | NewProfileRepository | `func NewProfileRepository(db *sql.DB) *ProfileRepository` | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:19 | method | GetByCode | SQL by tenant+code | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:54 | method | List | SQL list tenant-scoped | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:102 | method | Create | INSERT document_profiles | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:127 | method | Update | UPDATE document_profiles | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:166 | type | AreaRepository | struct | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:170 | func | NewAreaRepository | `func NewAreaRepository(db *sql.DB) *AreaRepository` | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:174 | method | GetByCode | SQL by tenant+code | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:207 | method | List | SQL list tenant-scoped | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:253 | method | Create | INSERT document_process_areas | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:275 | method | Update | UPDATE document_process_areas | (undocumented) |
| internal/modules/taxonomy/infrastructure/repository.go:308 | method | ListAncestors | recursive CTE walk parent_code | (undocumented) |
| internal/modules/taxonomy/infrastructure/template_version_checker.go:9 | type | TemplateVersionChecker | struct | (undocumented) |
| internal/modules/taxonomy/infrastructure/template_version_checker.go:20 | func | NewTemplateVersionChecker | `func NewTemplateVersionChecker(db *sql.DB) *TemplateVersionChecker` | (undocumented) |
| internal/modules/taxonomy/infrastructure/template_version_checker.go:24 | method | IsPublished | SQL check on template_versions | (undocumented) |
| internal/modules/taxonomy/delivery/http/handler.go:36 | type | Handler | struct | (undocumented) |
| internal/modules/taxonomy/delivery/http/handler.go:42 | func | NewHandler | constructor | (undocumented) |
| internal/modules/taxonomy/delivery/http/handler.go:50 | method | RegisterRoutes | mux.HandleFunc bindings | (undocumented) |

## 3. HTTP operations

| Method | Path | Handler | Source |
|---|---|---|---|
| GET    | /api/v1/taxonomy/profiles | listProfiles | handler.go:51 |
| POST   | /api/v1/taxonomy/profiles | createProfile | handler.go:52 |
| GET    | /api/v1/taxonomy/profiles/{code} | getProfile | handler.go:53 |
| PATCH  | /api/v1/taxonomy/profiles/{code} | updateProfile | handler.go:54 |
| DELETE | /api/v1/taxonomy/profiles/{code} | archiveProfile | handler.go:55 |
| PUT    | /api/v1/taxonomy/profiles/{code}/default-template | setDefaultTemplate | handler.go:56 |
| GET    | /api/v1/taxonomy/areas | listAreas | handler.go:58 |
| POST   | /api/v1/taxonomy/areas | createArea | handler.go:59 |
| GET    | /api/v1/taxonomy/areas/{code} | getArea | handler.go:60 |
| PUT    | /api/v1/taxonomy/areas/{code} | updateArea | handler.go:61 |
| DELETE | /api/v1/taxonomy/areas/{code} | archiveArea | handler.go:62 |
| GET    | /api/v1/taxonomy/families | listFamilies | handler.go:64 |
| POST   | /api/v1/taxonomy/families | createFamily | handler.go:65 |
| GET    | /api/v1/taxonomy/families/{code} | getFamily | handler.go:66 |
| PATCH  | /api/v1/taxonomy/families/{code} | updateFamily | handler.go:67 |
| DELETE | /api/v1/taxonomy/families/{code} | deactivateFamily | handler.go:68 |

## 4. Migration list

| Filename | Verb | Tables touched |
|---|---|---|
| migrations/0023_init_document_family_and_profile_registry.sql | CREATE TABLE, FK | document_families, document_profiles |
| migrations/0024_grant_document_registry_privileges.sql | GRANT | document_families, document_profiles |
| migrations/0025_init_document_taxonomy.sql | CREATE TABLE, FK | document_process_areas |
| migrations/0026_grant_document_taxonomy_privileges.sql | GRANT | document_process_areas |
| migrations/0027_init_document_profile_schema_and_governance.sql | FK | document_profiles |
| migrations/0029_seed_metal_nobre_document_registry.sql | INSERT seed | document_families, document_profiles, document_process_areas |
| migrations/0032_deactivate_legacy_document_registry.sql | UPDATE | document_profiles, document_families |
| migrations/0034_rename_metal_nobre_document_labels.sql | UPDATE | document_profiles |
| migrations/0035_add_document_profile_alias.sql | ALTER + UPDATE | document_profiles |
| migrations/0046_add_document_code_sequence.sql | FK | document_profiles |
| migrations/0075_create_template_drafts_and_audit.sql | FK | document_profiles |
| migrations/0122_taxonomy_extend_document_profiles.sql | ALTER | document_profiles |
| migrations/0123_taxonomy_extend_process_areas.sql | ALTER + FK | document_process_areas |
| migrations/0124_registry_controlled_documents.sql | FK | document_profiles, document_process_areas |
| migrations/0125_registry_iam_user_process_areas_governance_events.sql | FK | document_process_areas |
| migrations/0134_approval_routes.sql | FK | document_profiles |
| migrations/0161_grant_families_write_privileges.sql | GRANT | document_families |
| migrations/0175_documents_area_name_snapshot.sql | trigger reads | document_process_areas |
| migrations/0182_cd_sequence_per_area.sql | FK | document_profiles, document_process_areas |
