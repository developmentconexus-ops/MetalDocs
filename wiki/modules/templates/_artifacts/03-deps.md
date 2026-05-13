# templates — Cross-dependency map
<!-- Phase 3 artifact. Read-only. Facts only. -->
Generated: 2026-05-10

---

## 1. Imports OUT (internal MetalDocs packages only)

| Imported package | First seen in (file:line) | Symbols used | Purpose |
|---|---|---|---|
| `internal/modules/iam/domain` | `delivery/http/handler.go:10` | `UserIDFromContext` | Extract actor user ID from request context |
| `internal/platform/httpresponse` | `delivery/http/handler.go:13` | `WriteJSON`, `ReadJSON` | Canonical JSON response writer |
| `internal/platform/tenant` | `delivery/http/handler.go:14` | `DevTenantID` | Fall-through tenant ID for dev environment |

Notes:
- All other intra-module imports (e.g. `templates/domain`, `templates/application`, `templates/api`) are self-imports and are excluded per instructions.
- The `application`, `domain`, and `repository` sub-packages have no external MetalDocs dependencies beyond those listed above (they depend only on stdlib and third-party drivers).

---

## 2. Imports IN (packages that import templates)

| Importer package | File:line of import | Symbols used | Notes |
|---|---|---|---|
| `internal/modules/documents/application` | `service.go:19` | `templatesdomain` (type aliases) | Document service holds `TemplateVersion` references |
| `internal/modules/documents/application` | `fillin_service.go:15` | `templatesdomain.Placeholder`, `templatesdomain.PHType` etc. | Fill-in service resolves placeholder schema from template domain types |
| `internal/modules/documents/application` | `freeze_service.go:15` | `tmpldom.VersionStatus`, `tmpldom.TemplateVersion` | Freeze service checks template version status before freezing |
| `internal/modules/documents/application` | `draft_resolver_service.go:10` | `tmpldom.Placeholder`, `tmpldom.PHComputed` | Draft resolver converts eigenpal native placeholder list to domain structs |
| `internal/modules/documents/application` | `snapshot_service.go:8` | `templatesdomain.TemplateSnapshot` | Snapshot service uses template snapshot type |
| `internal/modules/documents/repository` | `repository.go:16` | `templatesdomain.TemplateVersion` | Repo stores version foreign key; maps template domain types |
| `internal/modules/documents/repository` | `fillin_repository.go:9` | `templatesdomain.Placeholder` | Fill-in repository uses placeholder domain type |
| `internal/modules/documents/http` | `fillin_handler.go:17` | `templatesdomain.PHType`, `templatesdomain.Placeholder` | HTTP handler maps placeholder types to API response |
| `internal/modules/documents/http` | `placeholder_options_handler.go:8` | `templatesdomain.PHSelect`, `templatesdomain.Placeholder` | Options handler filters select-type placeholders |
| `internal/platform/objectstore` | `templates_presigner.go:13` | `domain.ErrUploadMissing` | Presigner adapter implements `application.Presigner`; uses domain error sentinel |
| `internal/platform/docgenv2` | `templates_reader.go` (no Go import; raw SQL only) | — | `TemplatesV2TemplateReader` reads directly from `templates_template_version` table; no Go import of templates packages |
| `internal/platform/docgenv2` | `templates_snapshot_reader.go` (no Go import; raw SQL only) | — | `TemplatesV2SnapshotReader` reads directly from DB tables; no Go import of templates packages |
| `apps/api/cmd/metaldocs-api` | `main.go:34–36` | `tv2app.New`, `tv2repo.New`, `tv2http.New` | DI composition root; constructs and registers all templates layers |
| `tests/docx_v2` | `scaffold_smoke_test.go:8` | `domain.Placeholder` (unclear: exact symbol) | Smoke test scaffolds placeholder domain objects |

Total external importers (non-self): 14 entries across 5 packages. None exceed 50.

---

## 3. DI / wiring touchpoints

| Site | File:line | What is wired |
|---|---|---|
| `apps/api/cmd/metaldocs-api/main.go` | `327` | `objectstore.NewTemplatesV2Presigner(deps.MinioClient, deps.MinioBucket, 25*1024*1024)` — constructs presigner adapter |
| `apps/api/cmd/metaldocs-api/main.go` | `328` | `tv2app.New(tv2repo.New(deps.SQLDB), tv2Presigner, realClock{}, realUUIDGen{})` — constructs repository and service; `resolvers` arg is omitted (nil) |
| `apps/api/cmd/metaldocs-api/main.go` | `329` | `tv2http.New(tv2Svc, nil).Register(mux)` — constructs handler with nil authz func (bypassed); registers all routes on the global mux |

Notes:
- `authz AuthzFunc` is wired as `nil`, causing the handler to fall through to the no-op `func(*http.Request, string, string, string) error { return nil }`.
- `ResolverRegistryReader` (the `resolvers` variadic arg to `tv2app.New`) is not passed, so resolver validation in `UpdateSchemas` is skipped at runtime.
- No wire/fx DI framework is used; construction is manual in `main.go`.

---

## 4. Configuration surface

No calls to `env.Get`, `os.Getenv`, `viper`, or feature-flag helpers were found anywhere in `internal/modules/templates/**/*.go`.

The only runtime-configurable value surfaces at construction time via `main.go`:

| Name | Read at (file:line) | Required? | Default |
|---|---|---|---|
| `deps.MinioClient` | `main.go:327` (passed to `NewTemplatesV2Presigner`) | Yes | — (set by DI deps loader) |
| `deps.MinioBucket` | `main.go:327` | Yes | — (set by DI deps loader) |
| Max object size | `main.go:327` literal `25*1024*1024` | n/a | 25 MiB hard-coded at wiring site |

No GUC `SET LOCAL` or PostgreSQL session-level config calls were found in module SQL.

---

## 5. Test surface

| Test file | Subject (file under test) | Kind |
|---|---|---|
| `application/approval_config_test.go` | `application/approval_config.go` | unit |
| `application/autosave_test.go` | `application/autosave.go` | unit |
| `application/create_test.go` | `application/create.go` | unit |
| `application/fakes_test.go` | shared in-memory repo fake for application tests | unit (helper) |
| `application/lifecycle_test.go` | `application/lifecycle.go` | unit |
| `application/queries_test.go` | `application/queries.go` | unit |
| `application/schema_test.go` | `application/schema.go` | unit |
| `application/visibility_graph_test.go` | `application/visibility_graph.go` | unit |
| `delivery/http/errors_test.go` | `delivery/http/errors.go` | unit |
| `delivery/http/routes_autosave_test.go` | `delivery/http/routes_autosave.go` | unit |
| `delivery/http/routes_catalog_test.go` | `delivery/http/routes_catalog.go` | unit |
| `delivery/http/routes_contract_test.go` | HTTP response shape contracts | unit |
| `delivery/http/routes_create_test.go` | `delivery/http/routes_create.go` | unit |
| `delivery/http/routes_lifecycle_test.go` | `delivery/http/routes_lifecycle.go` | unit |
| `domain/approval_test.go` | `domain/approval.go` | unit |
| `domain/schemas_test.go` | `domain/schemas.go` | unit |
| `domain/template_test.go` | `domain/template.go` | unit |
| `domain/version_test.go` | `domain/version.go` | unit |
| `repository/postgres_integration_test.go` | `repository/postgres.go` | integration |
