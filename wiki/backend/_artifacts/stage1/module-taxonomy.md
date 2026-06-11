# Stage-1 Audit Artifact — module-taxonomy

**Produced:** 2026-06-10 | **Branch at audit time:** qa/iam-area-membership | **Mode:** read-only code walk

---

## 1. Identity & purpose

`internal/modules/taxonomy` owns the **flat, code-keyed classification catalog** for the MetalDocs QMS:
three entities — `DocumentFamily`, `DocumentProfile`, `ProcessArea` — each backed by its own Postgres
table.  Families group profiles at the top level; profiles (per-tenant) bind to a family and carry a
`default_template_version_id` used by the document-creation wizard; areas (per-tenant) model the
operational org chart with an optional `parent_code` self-FK and application-layer cycle prevention.

The module is the **stable FK anchor** for two downstream modules: controlled-documents
(`document_profiles.code` + `process_areas.code` → CD code prefix `{profile}-{area}-{seq}`) and
documents (live read of `process_areas.name` for snapshot; `profileDefaultsAdapter` for default template
resolution).  It exposes 16 HTTP routes under `/api/v1/taxonomy/*` mounted via `taxonomyapi.HandlerWithOptions`
(oapi-codegen generated router from the OpenAPI spec — T-009 was resolved between the initial wiki and
the current code state).  Code immutability is a first-class QMS contract enforced at both the handler
layer and by DB triggers on all three tables.

---

## 2. File inventory

### `internal/modules/taxonomy/` (package root)

| File | Role |
|---|---|
| `module.go` | Composition root — wires repos, services, handler; exposes `Module` + `Dependencies` structs |
| `module_test.go` | Smoke-tests `New(...)` wiring when `TplChecker` is nil |

### `domain/`

| File | Role |
|---|---|
| `family.go` | `DocumentFamily` aggregate struct, `FamilyCode` type, sentinel errors, `NewDocumentFamily` constructor, `(*DocumentFamily).Deactivate` |
| `profile.go` | `DocumentProfile` aggregate struct (12 fields incl. `TenantID`, `ArchivedAt`), `ProfileCode` type, sentinel errors, `NewDocumentProfile`, `IsActive`, `Archive` |
| `area.go` | `ProcessArea` aggregate struct (incl. `TenantID`, `ParentCode`), `AreaCode` type, sentinel errors, `NewProcessArea`, `IsActive`, `Archive`, `trimOptionalAreaCode` |
| `constructors.go` | Shared helper `trimOptionalString` used by profile + area constructors |
| `port.go` | Repository interfaces `FamilyRepository`, `ProfileRepository`, `AreaRepository`; `TemplateVersionChecker`; `GovernanceLogger`; `FamilyTx`; `GovernanceEvent` struct; `GovernanceEventType` constants (11 event types) |
| `family_test.go` | Unit tests for `NewDocumentFamily` + `Deactivate` |
| `profile_test.go` | Unit tests for `NewDocumentProfile` + `Archive` |
| `area_test.go` | Unit tests for `NewProcessArea` + `Archive` |

### `application/`

| File | Role |
|---|---|
| `family_service.go` | `FamilyService` — List, Get, Create, Update (tx + `SELECT FOR UPDATE`), Deactivate (tx + `SELECT FOR UPDATE` + `HasActiveProfilesTx`); emits govLogger on all mutations |
| `profile_service.go` | `ProfileService` — List, Get, Create, Update, SetDefaultTemplate (tx), Archive (tx); emits govLogger on all mutations; panics if govLogger nil at construction |
| `area_service.go` | `AreaService` — List, Get, Create, Update, SetParent (tx + cycle check via `ListAncestorsTx`), Archive (tx); emits govLogger on all mutations |
| `governance_logger.go` | `DBGovernanceLogger` — legacy sink writing to `governance_events` table; marked `Deprecated` in favour of `AuditGovernanceAdapter` |
| `audit_governance_adapter.go` | `AuditGovernanceAdapter` — implements `domain.GovernanceLogger` by routing events to `auditdomain.Writer` (canonical `metaldocs.audit_events` sink) |
| `governance_payload.go` | `marshalGovernancePayload` helper — marshals `map[string]string` to JSON bytes |
| `family_service_test.go` | Unit tests for `FamilyService` (list, get, create, update, deactivate flows) |
| `profile_service_test.go` | Unit tests for `ProfileService` (list, get, create, archive flows) |
| `area_service_test.go` | Unit tests for `AreaService` (SetParent valid / cycle / nil parent) |
| `immutability_test.go` | Integration test (`//go:build integration`) — verifies `trg_document_profiles_code_immutable` fires on direct UPDATE of `code` |

### `infrastructure/`

| File | Role |
|---|---|
| `repository.go` | `ProfileRepository` + `AreaRepository` impls (`*sql.DB`-backed); `taxonomyTx` wrapper; `ListAncestors` + `ListAncestorsTx` recursive-CTE queries; `maxTaxonomyListRows = 1000`, `maxTaxonomyTreeDepth = 20`; `stringPtrToNull`, `nullStringPtr`, `areaCodePtrToNull`, `nullAreaCodePtr` helpers |
| `family_repository.go` | `FamilyRepository` impl; `familyTx` wrapper; `HasActiveProfiles` + `HasActiveProfilesTx` (both have `tenant_id` predicate); `GetByCodeForUpdate` uses `SELECT FOR UPDATE` inside caller-supplied tx |
| `authz_guc.go` | `setAuthzGUC` — sets `metaldocs.tenant_id` and `metaldocs.actor_id` session-local GUCs on a `*sql.Tx`; fails if either is absent from context |
| `template_version_checker.go` | `TemplateVersionChecker` — JOIN `templates_template_version v` + `templates_template t` on `v.id = $1 AND t.tenant_id = $2`; returns `(isPublished bool, profileCode string, err error)` |
| `authz_guc_test.go` | Unit tests for `setAuthzGUC` (missing tenant / missing actor paths) |
| `family_repository_test.go` | Unit tests for `FamilyRepository` (list, get, create, deactivate flows using a fake `*sql.DB`) |
| `template_version_checker_test.go` | Unit tests for `TemplateVersionChecker` (published / unpublished / no-rows paths) |

### `delivery/http/`

| File | Role |
|---|---|
| `handler.go` | `Handler` struct; `profileService`, `areaService`, `familyService` local interfaces; `NewHandler`; `RegisterRoutes` — mounts all 16 routes via `taxonomyapi.HandlerWithOptions(h, StdHTTPServerOptions{BaseURL: "/api/v1", ...})` |
| `routes_generated.go` | Adapter shim — implements `taxonomyapi.ServerInterface`; each method calls the private handler function; `ListTaxonomyFamilies` re-encodes the `include_inactive` query param into `r.URL.RawQuery` before forwarding |
| `routes_profiles.go` | Private handlers `listProfiles`, `createProfile`, `getProfile`, `updateProfile`, `setDefaultTemplate`, `archiveProfile`; `writeProfileError` switch (maps 8 sentinel/PG errors → RFC 9457 codes); `tenantIDFromRequest`, `parseIncludeArchived`; package-level `writeJSON`/`writeError` aliases to `httpresponse` |
| `routes_families.go` | Private handlers `listFamilies`, `createFamily`, `getFamily`, `updateFamily`, `deactivateFamily`; `writeFamilyError` switch; `parseBool` helper |
| `routes_areas.go` | Private handlers `listAreas`, `createArea`, `getArea`, `updateArea`, `archiveArea`; `writeAreaError` switch; `areaCodePtr` helper |
| `routes_profiles_contract_test.go` | Contract tests: error-envelope RFC 9457 shape, PATCH semantics, unique-violation 409 |
| `routes_families_contract_test.go` | Contract tests: unique-violation 409 for family create |
| `routes_areas_contract_test.go` | Contract tests: unique-violation 409 for area create, 404 on update of missing area |

### `api/`

| File | Role |
|---|---|
| `api.gen.go` | oapi-codegen generated file — `ServerInterface`, `ServerInterfaceWrapper`, `HandlerWithOptions`, generated models (`DocumentFamilyItem`, `DocumentProfileItem`, `ProcessAreaItem`, `Problem`, response aliases), `ListTaxonomyFamiliesParams`; `embedded-spec` included |
| `cfg.yaml` | oapi-codegen config — package `taxonomyapi`, `std-http-server: true`, `strict-server: true`, tag filter `taxonomy`, output `api.gen.go` |
| `gen.go` | `//go:generate` directive pointing at `api/openapi/v1/openapi.yaml` |

---

## 3. Public surface

### Exported Go symbols consumed by callers outside the module

| Symbol | Package | Consumed by |
|---|---|---|
| `Module`, `Dependencies` | `taxonomy` | `apps/api/cmd/metaldocs-api/main.go:314` |
| `(*Module).RegisterRoutes` | `taxonomy` | `main.go:315` |
| `NewProfileRepository` | `infrastructure` | `main.go:358` (standalone instance for `profileDefaultsAdapter`) |
| `NewTemplateVersionChecker` | `infrastructure` | `main.go:687` (passed in `Dependencies.TplChecker`) |
| `NewDBGovernanceLogger` (Deprecated) | `application` | `internal/modules/controlleddocuments/module.go:31` (legacy literal code path) |
| `NewAuditGovernanceAdapter` | `application` | `module.go:35` (wired when `AuditWriter != nil`) |
| `DocumentFamily`, `DocumentProfile`, `ProcessArea` | `domain` | `controlleddocuments/{application,delivery,infrastructure}`, `main.go` |
| `FamilyCode`, `ProfileCode`, `AreaCode` | `domain` | `controlleddocuments/application/service.go`, `main.go` |
| `GovernanceLogger`, `GovernanceEvent`, `GovernanceEventType` constants | `domain` | `controlleddocuments/module.go`, `application/*` |
| `FamilyRepository`, `ProfileRepository`, `AreaRepository`, `TemplateVersionChecker` | `domain` | interfaces only — used by service constructors |
| `ErrFamilyNotFound`, `ErrFamilyAlreadyInactive`, `ErrFamilyHasProfiles` | `domain` | `controlleddocuments` + handler error mappers |
| `ErrProfileNotFound`, `ErrProfileArchived`, `ErrProfileCodeImmutable`, `ErrTemplateNotPublished`, `ErrTemplateProfileMismatch` | `domain` | handler error mappers, `controlleddocuments` |
| `ErrAreaNotFound`, `ErrAreaArchived`, `ErrAreaParentCycle`, `ErrAreaParentCodeRequired`, `ErrAreaCodeImmutable` | `domain` | handler error mappers |

### HTTP routes

All 16 routes are mounted via `taxonomyapi.HandlerWithOptions` with `BaseURL: "/api/v1"` (`handler.go:43-51`).  The tier-1 authz capability dispatcher is in `apps/api/cmd/metaldocs-api/permissions.go:165-181`.

| Method | Path | Handler method | Tier-1 cap | Notes |
|---|---|---|---|---|
| GET | `/api/v1/taxonomy/profiles` | `ListTaxonomyProfiles` | `taxonomy.view` | `?include_archived=bool` |
| POST | `/api/v1/taxonomy/profiles` | `CreateTaxonomyProfile` | `taxonomy.manage` | |
| GET | `/api/v1/taxonomy/profiles/{code}` | `GetTaxonomyProfile` | `taxonomy.view` | |
| PATCH | `/api/v1/taxonomy/profiles/{code}` | `UpdateTaxonomyProfile` | `taxonomy.manage` | |
| DELETE | `/api/v1/taxonomy/profiles/{code}` | `ArchiveTaxonomyProfile` | `taxonomy.manage` | sets `archived_at` |
| PUT | `/api/v1/taxonomy/profiles/{code}/default-template` | `SetTaxonomyProfileDefaultTemplate` | `taxonomy.manage` | checks `IsPublished` + profile match |
| GET | `/api/v1/taxonomy/areas` | `ListTaxonomyAreas` | `taxonomy.view` | `?include_archived=bool` |
| POST | `/api/v1/taxonomy/areas` | `CreateTaxonomyArea` | `taxonomy.manage` | |
| GET | `/api/v1/taxonomy/areas/{code}` | `GetTaxonomyArea` | `taxonomy.view` | |
| PUT | `/api/v1/taxonomy/areas/{code}` | `UpdateTaxonomyArea` | `taxonomy.manage` | includes `parent_code` |
| DELETE | `/api/v1/taxonomy/areas/{code}` | `ArchiveTaxonomyArea` | `taxonomy.manage` | sets `archived_at` |
| GET | `/api/v1/taxonomy/families` | `ListTaxonomyFamilies` | `taxonomy.view` | `?include_inactive=bool` |
| POST | `/api/v1/taxonomy/families` | `CreateTaxonomyFamily` | `taxonomy.manage` | |
| GET | `/api/v1/taxonomy/families/{code}` | `GetTaxonomyFamily` | `taxonomy.view` | |
| PATCH | `/api/v1/taxonomy/families/{code}` | `UpdateTaxonomyFamily` | `taxonomy.manage` | |
| DELETE | `/api/v1/taxonomy/families/{code}` | `DeactivateTaxonomyFamily` | `taxonomy.manage` | sets `is_active=FALSE` |

Tier-2 in-tx `authz.Require(CapTaxonomyManage)` is wired in all six write methods across `FamilyRepository`, `ProfileRepository`, and `AreaRepository` (`infrastructure/family_repository.go:113,135,196,250`; `infrastructure/repository.go:162,202,449,485,531,604`).  Archive/deactivate paths have tier-1 only.  DB tripwire `trg_require_cap_asserted` is attached to all three owned tables via `migrations/0188_tripwire_extend.sql:211-224`.

---

## 4. Logic flows

### Flow 1 — `listFamilies` (canonical read path)

1. HTTP GET `/api/v1/taxonomy/families[?include_inactive=bool]` arrives at `routes_generated.go:56` (`ListTaxonomyFamilies`); the codegen wrapper re-encodes the decoded `params.IncludeInactive` back into `r.URL.RawQuery` and delegates to `routes_families.go:20` (`listFamilies`).
2. `parseBool` (`routes_families.go:117`) parses `include_inactive`; malformed value → 400.
3. `FamilyService.List` (`family_service.go:22`) calls `FamilyRepository.List` (`family_repository.go:61`).
4. Repository opens a fresh `sql.Tx`, calls `setAuthzGUC` (`authz_guc.go:14`) to set `metaldocs.tenant_id` + `metaldocs.actor_id` session-local GUCs, then calls `authz.Require(ctx, tx, CapTaxonomyView, "tenant")`.
5. SQL: `SELECT code, name, description, is_active, created_at FROM metaldocs.document_families [WHERE is_active = TRUE] ORDER BY code ASC` — no tenant predicate (table is global).
6. Returns `{"items": [...DocumentFamily]}` with HTTP 200.

Note: `listFamilies` does not call `tenantIDFromRequest` — families have no `tenant_id` column.  The GUC step still fires to satisfy the tripwire; `authz.Require` uses the session user's roles from context.

### Flow 2 — `createProfile` (tenanted write path)

1. POST `/api/v1/taxonomy/profiles` — body `{code, family_code, name, ...}`.
2. `routes_profiles.go:70` (`createProfile`): decodes JSON body → `profileUpsertRequest`.  Calls `tenantIDFromRequest` (`routes_profiles.go:255`) → `tenant.FromContext(r.Context())` — tenant is session-bound; returns `ErrTenantMissing` (500) if absent.
3. Builds `domain.DocumentProfile{TenantID, Code, FamilyCode, ...}`; validates `code != ""`.
4. `ProfileService.Create` (`profile_service.go:53`): calls `domain.NewDocumentProfile` for normalization + validation (trims, sets alias default, checks required fields).
5. `ProfileRepository.Create` (`repository.go:152`): opens `sql.Tx`, calls `setAuthzGUC` then `authz.Require(ctx, tx, CapTaxonomyManage, "tenant")`.  INSERT into `metaldocs.document_profiles`.  DB tripwire `trg_require_cap_asserted` fires on INSERT.  On success commits.
6. `ProfileService.Create` then calls `s.govLogger.Log(...)` with event type `profile.created` (`profile_service.go:70`); if `AuditWriter` was wired at startup, this routes to `AuditGovernanceAdapter.Log` → `audit_events` table; otherwise `DBGovernanceLogger.Log` → `governance_events` table.
7. Returns 201 with the `DocumentProfile` JSON.  Error mapping via `writeProfileError` (`routes_profiles.go:230`): PG `23505` → 409 `PROFILE_ALREADY_EXISTS`; PG `23503` → 409 `FAMILY_NOT_FOUND`; PG `23514` → 400 `VALIDATION_ERROR`.

### Flow 3 — `deactivateFamily` (state transition with tx locking)

1. DELETE `/api/v1/taxonomy/families/{code}`.
2. `routes_families.go:90` (`deactivateFamily`) calls `FamilyService.Deactivate` (`family_service.go:124`).
3. Service calls `FamilyRepository.BeginTx` (`family_repository.go:180`) to obtain a `*sql.Tx` wrapped as `domain.FamilyTx`.
4. `GetByCodeForUpdate` (`family_repository.go:188`): inside the tx, sets GUCs + `authz.Require(CapTaxonomyManage)`, then `SELECT ... FROM document_families WHERE code=$1 FOR UPDATE` — acquires a row-level write lock.
5. Resolves `tenantID` from context (`family_service.go:140-143`); calls `HasActiveProfilesTx` (`family_repository.go:218`) inside the **same tx**: `SELECT EXISTS(SELECT 1 FROM document_profiles WHERE tenant_id=$1 AND family_code=$2 AND archived_at IS NULL)` — now scoped to the caller's tenant.
6. If profiles exist → returns `ErrFamilyHasProfiles` → 409.
7. `(*DocumentFamily).Deactivate()` (`domain/family.go:45`) sets `IsActive=false`; returns `ErrFamilyAlreadyInactive` if already false.
8. `FamilyRepository.UpdateTx` (`family_repository.go:242`) executes `UPDATE document_families SET is_active=FALSE WHERE code=$4` inside the tx.  DB tripwire fires on UPDATE.
9. `tx.Commit()` — lock released.
10. After commit, `govLogger.Log(...)` emits `family.deactivated` event (`family_service.go:167`).
11. Returns 204.

### Flow 4 — `SetDefaultTemplate` (transactional profile update with cross-table check)

1. PUT `/api/v1/taxonomy/profiles/{code}/default-template` body `{template_version_id}`.
2. `routes_profiles.go:172` (`setDefaultTemplate`): validates body not empty; resolves `tenantID` + `actorUserID` from context.
3. `ProfileService.SetDefaultTemplate` (`profile_service.go:109`): begins tx via `ProfileRepository.BeginTx`.
4. `GetByCodeForUpdate` inside tx: locks profile row; checks `profile.IsActive()` (returns `ErrProfileArchived` → 409 if archived).
5. `TemplateVersionChecker.IsPublished` (`template_version_checker.go:30`): separate `*sql.DB` read (outside the profile tx) — joins `templates_v2_template_version` + `templates_v2_template` filtered by `v.id=$1 AND t.tenant_id=$2`.  Returns `(published bool, profileCode string, err error)`.
6. Validates `published=true` (else `ErrTemplateNotPublished` → 409) and `templateProfileCode == string(profileCode)` (else `ErrTemplateProfileMismatch` → 409).
7. Sets `profile.DefaultTemplateVersionID = &templateVersionID`; `ProfileRepository.UpdateTx` writes inside the tx; commits.
8. `govLogger.Log(...)` emits `profile.default_template_change`.
9. Returns 200 `{}`.

### Flow 5 — `SetParent` for area (cycle-safe hierarchy update)

1. Area update path via `updateArea` (`routes_areas.go:91`) sets `ParentCode` on the area struct, then calls `AreaService.Update` (`area_service.go:72`) — this is the non-transactional path; it does a `GetByCode` (no lock) then `Update` (no lock).
2. The dedicated `SetParent` method (`area_service.go:111`) is **not exposed via any HTTP route** — no handler in `routes_areas.go` calls it.  It is invoked only from `area_service_test.go:49` in tests.  The cycle check (ancestor walk via recursive CTE `ListAncestorsTx`) is therefore **not exercised in the production HTTP path**.
3. The production `updateArea` handler calls `AreaService.Update`, which does a `GetByCode` (separate tx) + `AreaRepository.Update` (separate tx) with no cycle check.  Parent assignment via the HTTP API is subject to cycles at the DB level only if the `fk_area_parent_tenant` FK constraint rejects a non-existent parent — it does not prevent cycles.

---

## 5. Dependencies

### Outbound — imports by taxonomy code

| Import | Used in | Why |
|---|---|---|
| `metaldocs/internal/platform/tenant` | `infrastructure/authz_guc.go`, `delivery/http/routes_profiles.go`, `infrastructure/template_version_checker.go`, `application/family_service.go` | Session-bound `tenant_id` from context; `DevTenantID` in integration test only |
| `metaldocs/internal/platform/authn` | `application/family_service.go`, `application/profile_service.go`, `application/area_service.go` | `UserIDFromContext` for `ActorUserID` in governance events |
| `metaldocs/internal/platform/httpresponse` | `delivery/http/routes_profiles.go` | `WriteJSON`, `WriteError` aliased as package-level vars |
| `metaldocs/internal/platform/problem` | `delivery/http/routes_profiles.go` | `problem.Code` type, `CodeValidationError`, `CodeInternalError` |
| `metaldocs/internal/modules/iam/authz` | `infrastructure/family_repository.go`, `infrastructure/repository.go` | `authz.Require(ctx, tx, cap, area)` — tier-2 in-tx authz check |
| `metaldocs/internal/modules/iam/domain` | `infrastructure/family_repository.go`, `infrastructure/repository.go`, `infrastructure/authz_guc.go` | `CapTaxonomyView`, `CapTaxonomyManage`, `UserIDFromContext` |
| `metaldocs/internal/modules/audit/domain` | `application/audit_governance_adapter.go`, `module.go` | `auditdomain.Writer`, `auditdomain.Event` — canonical audit sink |
| `metaldocs/internal/modules/taxonomy/api` | `delivery/http/handler.go`, `delivery/http/routes_generated.go` | oapi-codegen generated router + `ServerInterface` |
| `database/sql` | `infrastructure/*`, `application/governance_logger.go` | Postgres access via `*sql.DB` / `*sql.Tx` |
| `github.com/jackc/pgx/v5/pgconn` | `delivery/http/routes_*.go` | `*pgconn.PgError` for PG error code mapping |
| `github.com/google/uuid` | `application/audit_governance_adapter.go` | `uuid.NewString()` for audit event ID |

### Inbound — verified by grep

| Importer | Path | What is used |
|---|---|---|
| `apps/api/cmd/metaldocs-api/main.go` | `:64-66` | `taxonomy.Module`, `taxonomydomain.*`, `taxonomyinfra.NewProfileRepository`, `taxonomyinfra.NewTemplateVersionChecker` |
| `internal/modules/controlleddocuments/module.go` | `:12-13` | `taxonomyapp.NewDBGovernanceLogger` (legacy literal code path), `taxonomydomain.*` |
| `internal/modules/controlleddocuments/infrastructure/repository.go` | `:19` | `taxonomydomain.DocumentProfile`, `taxonomydomain.ProcessArea` |
| `internal/modules/controlleddocuments/delivery/http/routes.go` | `:19` | `taxonomydomain.*` |
| `internal/modules/controlleddocuments/application/service.go` | `:19` | `taxonomydomain.*` |
| `internal/modules/controlleddocuments/application/service_test.go` | `:17` | `taxonomydomain.*` |
| `internal/modules/controlleddocuments/delivery/http/routes_contract_test.go` | `:22` | `taxonomydomain.*` |

---

## 6. Persistence

### Tables owned

| Table | Schema | PK | Tenant scoped | Migration created |
|---|---|---|---|---|
| `document_families` | `metaldocs` | `code TEXT` | No — global catalog | `0023_init_document_family_and_profile_registry.sql:1-7` |
| `document_profiles` | `metaldocs` | `code TEXT` (cross-tenant unique); `ux_document_profiles_tenant_code (tenant_id, code)` UNIQUE | Yes — `tenant_id UUID NOT NULL DEFAULT DevTenantID` | `0023:9-17` base; `0122_taxonomy_extend_document_profiles.sql:4-49` tenant + governance extension |
| `document_process_areas` | `metaldocs` | `code TEXT` (cross-tenant unique); `ux_process_areas_tenant_code (tenant_id, code)` UNIQUE; self-FK `fk_area_parent_tenant (tenant_id, parent_code) → (tenant_id, code)` | Yes — `tenant_id UUID NOT NULL DEFAULT DevTenantID` | `0025_init_document_taxonomy.sql:1-7` base; `0123_taxonomy_extend_process_areas.sql:3-75` tenant + hierarchy extension |

### Tables read (cross-module joins)

| Table | Schema | Used in | Why |
|---|---|---|---|
| `templates_v2_template_version` | public | `template_version_checker.go:15-21` | Check `status = 'published'` |
| `templates_v2_template` | public | `template_version_checker.go:17` | Resolve `doc_type_code` (profile code) + tenant filter |
| `governance_events` | (default/public) | `application/governance_logger.go:28` | Legacy parallel audit sink written by `DBGovernanceLogger` |
| `audit_events` | `metaldocs` | `application/audit_governance_adapter.go:33` | Canonical sink via `auditdomain.Writer.Record` |

### Key migration files

| Migration | Taxonomy content |
|---|---|
| `0023_init_document_family_and_profile_registry.sql` | Creates `document_families` + `document_profiles`; seeds 10 families; adds `document_profile_code`/`document_family_code` columns to `documents` |
| `0024_grant_document_registry_privileges.sql` | GRANT to `metaldocs_app` |
| `0025_init_document_taxonomy.sql` | Creates `document_process_areas` + legacy `document_subjects` (orphan — no longer referenced by any Go code); adds `process_area_code`/`subject_code` to `documents` |
| `0026_grant_document_taxonomy_privileges.sql` | GRANT |
| `0122_taxonomy_extend_document_profiles.sql` | Adds `tenant_id`, governance columns, `profile_code_format` CHECK, `ux_document_profiles_tenant_code` UNIQUE index, `trg_document_profiles_code_immutable` trigger |
| `0123_taxonomy_extend_process_areas.sql` | Adds `tenant_id`, `parent_code`, hierarchy columns, `ux_process_areas_tenant_code` UNIQUE index, `fk_area_parent_tenant` self-FK, `area_code_format` CHECK, `trg_process_areas_code_immutable` trigger |
| `0161_grant_families_write_privileges.sql` | GRANT write privileges on `document_families` to `metaldocs_app` |
| `0175_documents_area_name_snapshot.sql` | Adds `area_name_snapshot` column to `documents` for live read |
| `0188_tripwire_extend.sql:211-257` | Attaches `trg_require_cap_asserted` to all 3 taxonomy tables; adds `trg_reject_families_code_update` on `document_families` |

### Query patterns

- All reads and writes use `*sql.DB.BeginTx` + `setAuthzGUC` + `authz.Require` before the DML.  Every method creates its own tx; there is no service-layer tx spanning multiple repos.
- Writes that need atomicity use explicit `BeginTx`/`GetByCodeForUpdate` (SELECT FOR UPDATE)/DML/`Commit` within one repo method or the service opens a `domain.FamilyTx` and passes it to `*Tx` repo variants.
- List queries for profiles and areas have a hard cap of `LIMIT 1000` (`repository.go:20` constant `maxTaxonomyListRows`).  `listFamilies` has no LIMIT.
- Ancestor walk uses a bounded recursive CTE with `depth < maxTaxonomyTreeDepth` (20) to guard against runaway recursion.

---

## 7. Config & environment

No taxonomy-specific environment variables or config keys exist (`module.go` receives only `*sql.DB`, `TemplateVersionChecker`, and `auditdomain.Writer`).

`tenant.DevTenantID` (`ffffffff-ffff-ffff-ffff-ffffffffffff`) is a compile-time constant used as the default `tenant_id` value in migrations `0122:5` and `0123:4`, and referenced in the integration test (`immutability_test.go:47`).  It is not used as a runtime fallback in taxonomy HTTP handlers — `tenant.FromContext` will return an error if context lacks a session tenant.

---

## 8. Concurrency & async

- **No goroutines or channels** in any taxonomy code path.
- **Transactions with row locking** are used for all mutating flows that require atomicity: `FamilyService.Update` and `FamilyService.Deactivate` use `SELECT FOR UPDATE` via `GetByCodeForUpdate`; `ProfileService.SetDefaultTemplate`, `ProfileService.Archive`, `AreaService.SetParent`, and `AreaService.Archive` follow the same pattern.
- **`AreaService.Update`** (the only area mutation path reachable via HTTP) does a `GetByCode` + `Update` across two discrete transactions — no row lock.  A concurrent archive between the read and the write would allow updating an archived area.
- **Governance log writes** (`govLogger.Log`) are called after `tx.Commit()` in all service methods.  A process crash between commit and log would silently drop the governance event — no outbox, no retry.  This applies to both `DBGovernanceLogger` (INSERT to `governance_events`) and `AuditGovernanceAdapter` (`audit_events`).
- **`TemplateVersionChecker.IsPublished`** executes on a raw `*sql.DB` connection outside the profile tx (`template_version_checker.go:37`).  A concurrent template-version status change between the check and the profile UPDATE would pass the stale published state.  [runtime-unverified: frequency of this race in practice]

---

## 9. Error handling & observability

### Error mapping

Each aggregate has a dedicated `write*Error(w, err)` function:

- `writeProfileError` (`routes_profiles.go:230`): maps 7 sentinel errors + 3 PG codes (23503, 23505, 23514) to RFC 9457 `application/problem+json` codes.
- `writeFamilyError` (`routes_families.go:98`): maps 3 sentinel errors + 2 PG codes (23505, 23514).
- `writeAreaError` (`routes_areas.go:155`): maps 5 sentinel errors + 2 PG codes (23505, 23514).

All three use `httpresponse.WriteError` which calls `problem.Write` — RFC 9457 `application/problem+json` content type with `status`, `title`, `code` fields.

### Logging

- `slog.Error("taxonomy profile error", "err", err)` at `routes_profiles.go:250` — only for unmatched 500 paths.
- `slog.Error("taxonomy family error", "err", err)` at `routes_families.go:113` — same pattern.
- `slog.WarnContext` at `module.go:37` when `AuditWriter` is nil at construction time.
- No structured logging of request IDs, tenant IDs, or timing in any handler.
- No metrics or distributed tracing integration.

### Governance audit

Two parallel sinks depending on `Dependencies.AuditWriter`:
- If `AuditWriter != nil` (production): `AuditGovernanceAdapter` → `metaldocs.audit_events` via `auditdomain.Writer.Record`.
- If `AuditWriter == nil` (legacy / test): `DBGovernanceLogger` → `governance_events` (legacy table, `Deprecated` comment at `governance_logger.go:15`).

Coverage: all 11 `GovernanceEventType` constants are emitted — `family.created/updated/deactivated` (`family_service.go:53,110,167`), `profile.created/updated/default_template_change/archived` (`profile_service.go:70,96,155,194`), `area.created/updated/parent_changed/archived` (`area_service.go:59,98,160,199`).

---

## 10. Legacy / duplication / smell flags

- **F-01 — `document_subjects` orphan table** | `archive/migrations/0025_init_document_taxonomy.sql:9-16` | Migration `0025` created `metaldocs.document_subjects` (a `process_area_code`-keyed subject catalog with `is_active`, FKs from `documents`).  Zero Go code references this table — no repository, no domain type, no route.  The table exists in the DB schema with no owner.  Smell: dead schema surface; cross-tenant blast risk if ever written to via raw SQL.

- **F-02 — `AreaService.SetParent` unreachable via HTTP** | `application/area_service.go:111`; `delivery/http/routes_areas.go` (no call) | `SetParent` implements a transactional, cycle-safe parent assignment with `SELECT FOR UPDATE` + recursive CTE ancestor check.  The production `updateArea` handler calls `AreaService.Update` instead, which writes `parent_code` without any cycle check or row lock.  `SetParent` is called only from `area_service_test.go:49`.  The carefully built safety mechanism is dead code in production.  RF-016 (missing ADR for cycle prevention) compounds this — the spec-level design is undocumented AND the implementation is bypassed.

- **F-03 — Governance log written after tx.Commit (no outbox)** | `application/family_service.go:161`, `application/profile_service.go:193`, `application/area_service.go:199` (and all other govLogger calls post-commit) | After every mutating service method, `govLogger.Log` is called outside the committed transaction.  A crash, OOM, or context cancellation between commit and `Log` silently drops the audit event.  No outbox pattern, no retry queue.  For a QMS-regulated audit trail this is a data-loss path on every mutating operation.

- **F-04 — `DBGovernanceLogger` is deprecated but still the active path when `AuditWriter` is nil** | `application/governance_logger.go:15-17`; `module.go:37-39` | `NewDBGovernanceLogger` is marked `// Deprecated` and the `slog.Warn` at construction time signals it should not be used.  In `main.go:685-689`, `AuditWriter: deps.AuditWriter` is passed — if `deps.AuditWriter` is nil (e.g. a build with an uninitialised dependency), the legacy logger silently activates.  The two sinks (`governance_events` vs `audit_events`) have divergent schemas and no shared retention policy.  Also `internal/modules/controlleddocuments/module.go:31` hard-imports `taxonomyapp.NewDBGovernanceLogger` (the deprecated path) for its own governance logging — a cross-module coupling on a Deprecated symbol.

- **F-05 — Hard `LIMIT 1000` on profile and area list queries with TODO comment** | `infrastructure/repository.go:113,401` | `ProfileRepository.List` and `AreaRepository.List` apply `LIMIT 1000` via the compile-time constant `maxTaxonomyListRows`.  Both sites carry `// TODO: add pagination instead of returning the full profile catalog.`  `FamilyRepository.List` has no LIMIT at all (`family_repository.go:75-99`).  Three inconsistent bounds on related list operations.

- **F-06 — Redundant PRIMARY KEY on `code` alongside `(tenant_id, code)` UNIQUE index** | `archive/migrations/0023:10`; `archive/migrations/0025:2`; `archive/migrations/0122:31-32`; `archive/migrations/0123:26-28` | `document_profiles` and `document_process_areas` carry `code TEXT PRIMARY KEY` (cross-tenant uniqueness) plus a separate `UNIQUE (tenant_id, code)` index added by the 0122/0123 migrations.  The PK alone makes cross-tenant code collisions impossible, so the UNIQUE index adds an index for tenant-filtered lookups but no new logical uniqueness.  Structural artefact of the phased migration; not a correctness defect but carries storage + planner overhead and confuses readers.

- **F-07 — `AreaService.Update` performs two separate transactions (read + write) without row locking** | `application/area_service.go:73-88` | `Update` calls `GetByCode` (new tx, no FOR UPDATE) then `AreaRepository.Update` (new tx).  A concurrent archive between the two calls would allow updating an archived area.  Contrast with `SetDefaultTemplate` and `Archive` which both use `GetByCodeForUpdate` inside a single tx.  Inconsistent locking discipline across service methods.

- **F-08 — `FamilyRepository.GetByCode` (non-update read) opens a tx and sets `CapTaxonomyView` GUC** | `infrastructure/family_repository.go:29-59` | The non-update `GetByCode` opens a `BeginTx`, sets GUCs, calls `authz.Require(CapTaxonomyView)`, runs the SELECT, then the deferred `Rollback` discards the tx.  All reads on `ProfileRepository` and `AreaRepository` follow the same pattern.  The tx is used solely to scope the GUC `set_config(..., true)` (transaction-local).  This means every single-row read acquires and immediately discards a Postgres transaction.  The overhead is low for a catalog that is read infrequently, but the pattern departs from a simple `QueryRowContext` on `*sql.DB`.

- **F-09 — `routes_generated.go` re-encodes query param that was already parsed by codegen** | `delivery/http/routes_generated.go:56-66` | `ListTaxonomyFamilies` receives `params.ListTaxonomyFamiliesParams` (codegen-parsed `include_inactive`), then re-encodes the value back into `r.URL.RawQuery` before calling the private `listFamilies` which re-parses it via `parseBool(r.URL.Query().Get("include_inactive"))`.  The codegen-parsed value is thrown away; the raw query string is re-parsed.  The adapter shim defeats the purpose of the generated type-safe parameter.

- **F-10 — `TemplateVersionChecker.IsPublished` runs outside the profile tx** | `infrastructure/template_version_checker.go:30-48` | `IsPublished` uses a bare `c.db.QueryRowContext` (no tx, no GUC setup) against `templates_v2_template_version`.  A template version status change that occurs between the check and the profile `UPDATE` inside `SetDefaultTemplate`'s tx would pass the stale published state.  [runtime-unverified: actual frequency] Also, the templates table lives in `public` schema and is cross-tenant scoped only by `t.tenant_id` — no tier-2 authz assert on the templates read.

- **F-11 — No Go doc comments on any exported symbol** | All files under `internal/modules/taxonomy/` | Every exported type, function, method, and error variable in the module lacks a `// TypeName ...` Go doc comment.  `go doc` returns signatures only.  80 exported symbols total.

- **F-12 — `document_profiles.is_active` column orphaned after `0122` migration** | `archive/migrations/0023:14` vs `archive/migrations/0122:13` | Migration `0023` created `document_profiles` with `is_active BOOLEAN NOT NULL DEFAULT TRUE`.  Migration `0122` added `archived_at TIMESTAMPTZ NULL` as the soft-archive field (per ADR 0010).  Go code uses `archived_at IS NULL` exclusively.  The original `is_active` column on `document_profiles` is never written or read by any Go code — it is an orphaned column from the pre-ADR-0010 schema.  (Compare: `document_families.is_active` is still active and read/written because families use the boolean pattern, predating ADR 0010.)

---

## 11. Wiki drift

Existing wiki docs checked: `wiki/modules/taxonomy.md` (last verified 2026-06-07) and `wiki/modules/taxonomy-tech-debt.md` (last verified 2026-06-07).

1. **`taxonomy.md` Key files line 12 claims `FamilyService` "constructor takes no govLogger"** — Code at `application/family_service.go:13-19` shows `govLogger domain.GovernanceLogger` field present and `NewFamilyService(families, govLogger)` accepting it since commit `115cb6359`.  T-004 is correctly marked closed in the tech-debt register but the key-files paragraph was never updated.

2. **`taxonomy.md` Key files line 13 claims `ProfileService` "Create/Update do not call [govLogger]"** — Code at `profile_service.go:70` (`Create`) and `:96` (`Update`) both call `s.govLogger.Log(...)` since commit `20bf2067a`.  T-005 is correctly marked closed in the tech-debt register but the key-files paragraph was not updated.

3. **`taxonomy.md` Key files line 14 claims `AreaService` "Create/Update do not [call govLogger]"** — Code at `area_service.go:59` (`Create`) and `:98` (`Update`) both call `s.govLogger.Log(...)` since commit `20bf2067a`.  T-005 closed but key-files text was not updated.

4. **`taxonomy.md` Key files line 18 claims `HasActiveProfiles` has "no tenant predicate"** — Code at `family_repository.go:170-173` shows `WHERE tenant_id = $1 AND family_code = $2 AND archived_at IS NULL` with an explicit `tenantID` parameter accepted via `HasActiveProfiles(ctx, tenantID, familyCode)`.  The interface at `domain/port.go:69` also shows the `tenantID string` parameter.  T-007 description in tech-debt says "cross-tenant SELECT" and "no tenant predicate" but both the interface and implementation now have the predicate.  The T-007 TOCTOU race concern (no tx / no lock) was also resolved: `Deactivate` now uses `GetByCodeForUpdate` + `HasActiveProfilesTx` inside a single tx (commits `5d9e46ba0` + `58a71b5aa`).  T-007 should be marked closed but the wiki still lists it as open.

5. **`taxonomy.md` §5.3 and `taxonomy.md` §6.3 describe `deactivateFamily` as three discrete SQL calls with no tx and no row lock** — Code at `family_service.go:124-179` uses `BeginTx`, `GetByCodeForUpdate` (FOR UPDATE), `HasActiveProfilesTx` (same tx), `UpdateTx` (same tx), `Commit`.  The "NO tx · NO row lock · TOCTOU race window" annotation in the sequence diagram is no longer accurate.

6. **`taxonomy.md` §2 claims "No OpenAPI spec, no oapi-codegen — divergence from ADR 0012 (T-009)"** and **Key files line 15 says "16 routes mounted on raw `net/http.ServeMux`"** — The module now has `internal/modules/taxonomy/api/` with `api.gen.go`, `cfg.yaml`, `gen.go`.  `handler.go:43-51` calls `taxonomyapi.HandlerWithOptions(h, StdHTTPServerOptions{BaseURL: "/api/v1", ...})`.  `routes_generated.go:10` has the compile-time assertion `var _ taxonomyapi.ServerInterface = (*Handler)(nil)`.  T-009 description in tech-debt says "raw `net/http.ServeMux` for 16 routes; no spec, no codegen" — this is no longer accurate; the spec was added and the routes are now codegen-mounted.  T-009 should be marked closed.

7. **`taxonomy.md` §3.2 Inbound interfaces line about `taxonomyapp.NewDBGovernanceLogger` states it is "reused by `registry/module.go:31`"** — The actual importer is `internal/modules/controlleddocuments/module.go:12,31` (the `controlleddocuments` module, not a package named `registry`).  The wiki uses an older internal name for the controlled-documents module.

8. **`taxonomy.md` main.go anchor `main.go:197-201,225,508-524`** — Actual line numbers as of the current tree: `taxonomyModule` construction at `main.go:314-315`; standalone `profileRepo` at `main.go:358`; `profileDefaultsAdapter` use at `main.go:412`; `profileDefaultsAdapter` type definition at `main.go:908-924`.  None of these match the `197-201,225,508-524` ranges cited in the wiki.

---

## 12. Open questions

1. **[runtime-unverified]** `document_subjects` table (created by migration `0025`) — is it referenced by any other module (e.g. legacy CK5 path, raw-SQL scripts) or is it safe to drop?  Docker is down; cannot verify live schema against pg_depend.

2. **[runtime-unverified]** The `TemplateVersionChecker` reads `templates_v2_template_version` with no GUC setup and no tier-2 authz assert.  Whether the Postgres tripwire (`trg_require_cap_asserted`) is attached to that table and would block the read cannot be verified without a live run.

3. **[runtime-unverified]** `document_profiles.is_active` column — confirmed as an orphan column in Go code, but whether any active DB view, trigger, or third-party integration reads it cannot be confirmed without inspecting live `pg_views` / `information_schema`.

4. **[runtime-unverified]** The `governance_events` table schema and its tenant scoping — `DBGovernanceLogger` inserts with `tenant_id` but the table DDL is not in the taxonomy migration set.  Which migration creates `governance_events` and whether it has an index on `tenant_id` is not confirmed from code inspection alone.

5. `AreaService.SetParent` is thoroughly tested and correctly implemented, but no HTTP handler calls it.  Whether `UpdateArea` is intentionally the area's parent-change mechanism (accepting the missing cycle check) or whether `SetParent` was intended as a dedicated endpoint that was never wired is undocumented.  There is no corresponding ADR (T-016 / RF-016).
