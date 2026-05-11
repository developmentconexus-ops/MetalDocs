# taxonomy — dependency map (03-deps)

_Generated: 2026-05-11. Read-only artifact._

---

### 1. Imports OUT

Internal packages this module imports (stdlib and `github.com/*` excluded).

| Imported package | First seen in (file:line) | Symbols used | Purpose |
|---|---|---|---|
| `metaldocs/internal/platform/authn` | `delivery/http/routes_areas.go:11` | `authn.UserIDFromContext` | Extract actor user ID for archive operations |
| `metaldocs/internal/platform/httpresponse` | `delivery/http/routes_profiles.go:13` | `httpresponse.WriteJSON`, `httpresponse.WriteError` | Canonical JSON response helpers |
| `metaldocs/internal/platform/tenant` | `delivery/http/routes_profiles.go:14` | `tenant.DevTenantID` | Fallback tenant ID when X-Tenant-ID header is absent |

**Checklist — PRESENT vs ABSENT:**

| Package | Status |
|---|---|
| `internal/platform/authn` | **PRESENT** — `routes_areas.go:11`, `routes_profiles.go:12` |
| `internal/platform/httpresponse` | **PRESENT** — `routes_profiles.go:13` |
| `internal/platform/tenant` | **PRESENT** — `routes_profiles.go:14` |
| `internal/platform/authz` | **ABSENT** — no import; capability gating is done in `apps/api/cmd/metaldocs-api/permissions.go`, not inside this module |
| `internal/audit` | **ABSENT** — module uses its own `governance_events` table via `DBGovernanceLogger`; no dependency on the shared audit module |
| `internal/modules/iam` | **ABSENT** — no import anywhere in `internal/modules/taxonomy/` |

---

### 2. Imports IN

Other internal MetalDocs packages that import `metaldocs/internal/modules/taxonomy/...`. Capped at 50.

| Importer package | File:line | Symbols used | Notes |
|---|---|---|---|
| `apps/api/cmd/metaldocs-api` | `main.go:54` | `taxonomy` (package) | Module wiring — calls `taxonomy.New(...)` |
| `apps/api/cmd/metaldocs-api` | `main.go:55` | `taxonomydomain "...taxonomy/domain"` | `taxonomydomain.DocumentProfile` in `profileDefaultsAdapter` |
| `apps/api/cmd/metaldocs-api` | `main.go:56` | `taxonomyinfra "...taxonomy/infrastructure"` | `NewTemplateVersionChecker`, `NewProfileRepository` |
| `apps/api/cmd/metaldocs-api/permissions.go` | `:158,:166,:174` | path strings only | Capability gate for `/api/v2/taxonomy/*` routes |
| `internal/modules/registry` (module) | `module.go:12` | `taxonomyapp "...taxonomy/application"` | `taxonomyapp.NewDBGovernanceLogger` reused directly |
| `internal/modules/registry/application` | `service.go:13` | `taxonomydomain "...taxonomy/domain"` | `DocumentProfile`, `ProcessArea`, `GovernanceEvent`, `GovernanceLogger`, sentinel errors |
| `internal/modules/registry/application` | `service_test.go:13` | `taxonomydomain "...taxonomy/domain"` | Test fixtures using taxonomy domain types |
| `internal/modules/registry/infrastructure` | `repository.go:15` | `taxonomydomain "...taxonomy/domain"` | `TaxonomyProfileReader.GetByCode`, `TaxonomyAreaReader.GetByCode` — reads `document_profiles` and `document_process_areas` directly |
| `internal/modules/registry/delivery/http` | `routes.go:17` | `taxonomydomain "...taxonomy/domain"` | Binding taxonomy domain types in HTTP handlers |
| `internal/modules/registry/delivery/http` | `routes_contract_test.go:20` | `taxonomydomain "...taxonomy/domain"` | Contract test fixtures |

**Key verifications:**

| Expected importer | Status |
|---|---|
| `apps/api/cmd/metaldocs-api/main.go` — module wiring | **PRESENT** — `main.go:197` `taxonomy.New(...)`, `:201` `RegisterRoutes` |
| `apps/api/cmd/metaldocs-api/permissions.go` — capability gate | **PRESENT** — `:158,:166,:174` (path-string match; no Go import needed) |
| `internal/modules/registry/...` — profile + area FKs | **PRESENT** — `registry/infrastructure/repository.go:15`, `registry/application/service.go:13` |
| `internal/modules/documents/...` or `documents_v2` | **ABSENT** — `internal/modules/documents` imports taxonomy **zero times**. The profile adapter (`profileDefaultsAdapter`) lives in `apps/api/cmd/metaldocs-api/main.go:508` and calls `taxonomyinfra.NewProfileRepository` directly, then passes itself into `documents.Dependencies.ProfileDefaults`. |
| `internal/modules/templates_v2/...` | **ABSENT** — templates_v2 uses `doc_type_code` (its own field mapping to profile code); it imports no taxonomy packages |
| `internal/modules/approval/...` | **ABSENT** — `documents/approval` reads `process_area_code_snapshot` from the `documents` table (a denormalised snapshot), not from taxonomy directly |

---

### 3. DI / wiring touchpoints

| Site (file:line) | What is wired |
|---|---|
| `apps/api/cmd/metaldocs-api/main.go:197` | `taxonomy.New(taxonomy.Dependencies{DB: deps.SQLDB, TplChecker: ...})` — constructs the full module |
| `apps/api/cmd/metaldocs-api/main.go:199` | `taxonomyinfra.NewTemplateVersionChecker(deps.SQLDB)` — passed as `TplChecker` into `taxonomy.New` |
| `apps/api/cmd/metaldocs-api/main.go:201` | `taxonomyModule.RegisterRoutes(mux)` — mounts `/api/v2/taxonomy/*` routes |
| `apps/api/cmd/metaldocs-api/main.go:225` | `profileRepo := taxonomyinfra.NewProfileRepository(deps.SQLDB)` — second, standalone instance for `profileDefaultsAdapter` |
| `apps/api/cmd/metaldocs-api/main.go:280` | `ProfileDefaults: &profileDefaultsAdapter{profileRepo: profileRepo}` — adapter injected into `documents.Dependencies` |
| `apps/api/cmd/metaldocs-api/main.go:508-524` | `profileDefaultsAdapter` struct definition — bridges `taxonomydomain.DocumentProfile.DefaultTemplateVersionID` → `documents_v2` `ProfileDefaultTemplateReader` interface |
| `internal/modules/registry/module.go:31` | `taxonomyapp.NewDBGovernanceLogger(deps.DB)` — registry reuses taxonomy's governance logger implementation |
| `internal/modules/registry/module.go:29-30` | `NewTaxonomyProfileReader(deps.DB)`, `NewTaxonomyAreaReader(deps.DB)` — registry has its own SQL readers for taxonomy tables (no runtime dependency on taxonomy module) |
| `internal/modules/taxonomy/module.go:22-31` | Internal wiring: `NewProfileRepository`, `NewAreaRepository`, `NewFamilyRepository`, `NewDBGovernanceLogger`, service constructors, `NewHandler` |

---

### 4. Configuration surface

`n/a — module reads no env vars or config keys directly.`

The two `os.Getenv` calls in `application/immutability_test.go:21,23` (`DATABASE_URL`, `METALDOCS_DATABASE_URL`) are integration test scaffolding only, behind `//go:build integration`.

---

### 5. Test surface

| Test file | Subject | Kind |
|---|---|---|
| `application/area_service_test.go` | `AreaService` CRUD + archive rules | unit |
| `application/family_service_test.go` | `FamilyService` CRUD | unit |
| `application/profile_service_test.go` | `ProfileService` CRUD + SetDefaultTemplate + archive | unit |
| `application/immutability_test.go` | DB check-constraint blocks code mutation | integration |
| `domain/area_test.go` | `ProcessArea` value-object validation | unit |
| `domain/family_test.go` | `DocumentFamily` value-object validation | unit |
| `domain/profile_test.go` | `DocumentProfile` value-object validation | unit |
| `delivery/http/routes_families_contract_test.go` | Families HTTP contract (shapes / status codes) | contract |
| `delivery/http/routes_profiles_contract_test.go` | Profiles HTTP contract (shapes / status codes) | contract |
| `infrastructure/template_version_checker_test.go` | `TemplateVersionChecker.IsPublished` | unit |

---

### 6. Cross-module data contracts

#### `document_profiles.code`
Used as part of the CD code prefix via `registrydomain.AutoCode(profileCode, areaCode, seq)`.

- `internal/modules/registry/domain/controlled_document.go:48` — `AutoCode` function definition: `fmt.Sprintf("%s-%s-%03d", strings.ToUpper(profileCode), strings.ToUpper(areaCode), seq)`
- `internal/modules/registry/application/service.go:168,182` — call sites passing `cmd.ProfileCode` into `AutoCode`
- `internal/modules/registry/infrastructure/repository.go:28,42,48,59,68,147,214,229,253` — `profile_code` column referenced in INSERT/SELECT/WHERE throughout the registry repository

#### `document_profiles.default_template_version_id`
Read by the `documents_v2` wizard via the `profileDefaultsAdapter`.

- `apps/api/cmd/metaldocs-api/main.go:515-524` — `profileDefaultsAdapter.GetDefaultTemplateVersionID` reads `profile.DefaultTemplateVersionID` from the taxonomy `ProfileRepository.GetByCode` result
- `internal/modules/taxonomy/infrastructure/repository.go:21-22` — SELECT includes `default_template_version_id`; scanned at line 29
- `internal/modules/registry/infrastructure/repository.go:305-306` — `TaxonomyProfileReader` also SELECTs `default_template_version_id` (read by registry on CD create for template validation path)

#### `process_areas.code`
Used as the second segment of the CD code prefix and as a filter/FK in registry.

- `internal/modules/registry/domain/controlled_document.go:48` — `AutoCode` second argument
- `internal/modules/registry/application/service.go:164,168,178,182` — `cmd.ProcessAreaCode` passed to `seq.NextAndIncrement` and `AutoCode`
- `internal/modules/registry/infrastructure/repository.go:73,78,214,216,229,253` — `process_area_code` used in queries

#### `process_areas.name`
Snapshotted into `documents.area_name_snapshot` at document creation time.

- `internal/modules/documents/repository/repository.go:94-96` — live SQL: `SELECT name FROM metaldocs.document_process_areas WHERE tenant_id=$1::uuid AND code=$2` scanned into `areaName`; inserted as `area_name_snapshot` at `:101`
- `internal/modules/documents/application/context_builder.go:42,70` — reads `area_name_snapshot` back from the `documents` table to build eigenpal document context

#### `document_profiles.family_code`
FK to `document_families.code`; surfaced in profile reads.

- `internal/modules/taxonomy/infrastructure/repository.go:21` — SELECT includes `family_code`; scanned into `profile.FamilyCode` at line 31
- `internal/modules/registry/infrastructure/repository.go:305` — `TaxonomyProfileReader` SELECT also includes `family_code`; scanned at line 316 (`&profile.FamilyCode`)
- `internal/modules/registry/domain/sequence_test.go:44` — integration test seed INSERT references `family_code` column

---

**Summary: 3 OUT edges · 10 IN edges · 9 DI touchpoints · 0 config keys**
