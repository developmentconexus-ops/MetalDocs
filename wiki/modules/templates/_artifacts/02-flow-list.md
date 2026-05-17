# templates - Flow Trace: GET /api/v1/templates (ListTemplates)
<!-- Read-only trace. Facts only. -->
Generated: 2026-05-10
Updated: 2026-05-17 for wizard DOCX import + permission simplification.

## Header
- Operation: `GET /api/v1/templates`
- API op name: `ListTemplates`
- Module: `internal/modules/templates`
- Primary path: Route -> `delivery/http` handler -> `application.Service.ListTemplates` -> `repository.Repository.ListTemplates` -> SQL `templates_template`

## Flow

### 1) ROUTE REGISTRATION
- Generated oapi-codegen mount exists: `m.HandleFunc(http.MethodGet+" "+options.BaseURL+"/api/v1/templates", wrapper.ListTemplates)` at `internal/modules/templates/api/api.gen.go:954`.
- Generated wrapper dispatch calls module handler interface method: `siw.Handler.ListTemplates(w, r)` at `internal/modules/templates/api/api.gen.go:631-634`.
- Module route registration mounts generated wrapper endpoint directly: `mux.HandleFunc("GET /api/v1/templates", generated.ListTemplates)` at `internal/modules/templates/delivery/http/handler.go:40`.
- In module wiring, `generated` is a `templatesapi.ServerInterfaceWrapper{Handler: h, ...}` at `internal/modules/templates/delivery/http/handler.go:32-37`.
- Strict-server codegen wiring exists (`StrictServerInterface`, `NewStrictHandler`, `strictHandler.ListTemplates`) at `internal/modules/templates/api/api.gen.go:1193-1199`, `1228-1240`, `1276-1291`.
- Runtime use of `NewStrictHandler(...)` for templates route mounting is UNVERIFIED from inspected wiring; module register path uses `ServerInterfaceWrapper` (`internal/modules/templates/delivery/http/handler.go:31-40`).

### 2) HANDLER
- Generated module handler entrypoint: `func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request)` delegates to `h.listTemplates(w, r)` at `internal/modules/templates/delivery/http/routes_generated.go:15-17`.
- Concrete handler signature: `func (h *Handler) listTemplates(w http.ResponseWriter, r *http.Request)` at `internal/modules/templates/delivery/http/routes_query.go:12`.
- Query parse and validation:
- `limit` parsed by `readQueryInt(q.Get("limit"), 50)` and returns `400 invalid_limit` when parse fails (`routes_query.go:20-24`, `214-223`).
- `offset` parsed by `readQueryInt(q.Get("offset"), 0)` and returns `400 invalid_offset` when parse fails (`routes_query.go:25-29`, `214-223`).
- `doc_type` is `strings.TrimSpace(q.Get("doc_type"))`; non-empty value is passed as `*string` (`routes_query.go:31-35`).
- `area` and actor-area visibility inputs are no longer parsed by the active list path.
- Request to service is built as `application.ListFilter{TenantID, DocTypeCode, Limit, Offset}` at `internal/modules/templates/delivery/http/routes_query.go`.

### 3) AUTHZ
- Handler calls authz function for list: `h.authz(r, tenantID, "*", "template.view")` at `internal/modules/templates/delivery/http/routes_query.go:14`.
- If authz fails, handler writes mapped error and returns at `routes_query.go:14-17`.
- Handler constructor replaces nil authz with no-op: `if authz == nil { authz = func(...){ return nil } }` at `internal/modules/templates/delivery/http/handler.go:24-27`.
- App wiring passes nil authz: `tv2http.New(tv2Svc, nil).Register(mux)` at `apps/api/cmd/metaldocs-api/main.go:329`.
- AUTHZ BYPASSED (nil authz at wiring site).
- `authz.Require(...)` call for this path is UNVERIFIED; no such call appears in traced handler path.

### 4) TENANT SCOPING
- `tenantID` for list is derived by `tenantIDFromReq(r)` at `internal/modules/templates/delivery/http/routes_query.go:13`.
- `tenantIDFromReq` returns `X-Tenant-ID` header when non-empty; otherwise returns `tenant.DevTenantID` at `internal/modules/templates/delivery/http/handler.go:84-89`.
- Tenant is carried into filter as `ListFilter.TenantID` at `internal/modules/templates/delivery/http/routes_query.go:38`.
- Repository enforces tenant predicate in SQL: `WHERE tenant_id = $1` at `internal/modules/templates/repository/postgres.go:94`.

### 5) SERVICE LAYER
- Service method signature: `func (s *Service) ListTemplates(ctx context.Context, f ListFilter) ([]*domain.Template, error)` at `internal/modules/templates/application/queries.go:28`.
- Service behavior for list is pass-through: `return s.repo.ListTemplates(ctx, f)` at `internal/modules/templates/application/queries.go:29`.
- No additional list business rules in service path are present in this method (`queries.go:28-30`).

### 6) REPO LAYER
- Repository method signature: `func (r *Repository) ListTemplates(ctx context.Context, f application.ListFilter) ([]*domain.Template, error)` at `internal/modules/templates/repository/postgres.go:88`.
- SQL source table is `templates_template` (`postgres.go:93`).
- Selected columns:
- `id::text, tenant_id, doc_type_code, key, name, description, latest_version, published_version_id::text, created_by, system_owned, created_at, archived_at` (`postgres.go`).
- Join logic: no JOIN in `ListTemplates` query (`postgres.go:89-103`).
- WHERE clauses:
- Tenant filter: `tenant_id = $1` (`postgres.go:94`).
- Optional doc type: `($2::text IS NULL OR doc_type_code = $2)` (`postgres.go:95`).
- System-owned templates excluded: `system_owned = false`.
- No creator-scoped area or visibility predicates are applied; template availability is profile/document-type/lifecycle + IAM driven, not `areas`/`visibility`/`specific_areas` driven.
- Sort/pagination: `ORDER BY created_at DESC LIMIT $3 OFFSET $4`.
- Query args mapping: `$1=f.TenantID`, `$2=f.DocTypeCode`, `$3=f.Limit`, `$4=f.Offset`.
- Row mapping uses `scanTemplate(rows)` loop (`postgres.go:122-128`).

### 7) ERROR ENVELOPE
- Handler writes errors via `writeErr(...)` and `writeMappedErr(...)` (`internal/modules/templates/delivery/http/handler.go:95-102`, `108-118`).
- `writeErr` payload shape is legacy envelope:
- `{ "error": { "code": <code>, "message": <message> } }` (`handler.go:96-100`).
- Response content-type comes from `httpresponse.WriteJSON` as `application/json` (`internal/platform/httpresponse/response.go:8-11`).
- `MapErr` maps domain errors to status/code (`internal/modules/templates/delivery/http/errors.go:10-44`).
- RFC 9457 infrastructure exists (`problem.Write` sets `application/problem+json`) at `internal/platform/problem/problem.go:72-82`, but this list handler path uses `writeErr`/`WriteJSON` instead.

### 8) RESPONSE MAPPING
- Repository DB row -> domain mapping (`scanTemplate`):
- Scan targets `domain.Template` fields and SQL intermediates (`internal/modules/templates/repository/mappers.go:14-26`).
- Nullable `published_version_id` and `archived_at` mapped when valid; `system_owned` maps into the domain template for list exclusion and system template handling.
- Handler domain -> API JSON mapping:
- Each `*domain.Template` transformed by `toTemplateResponse(tpl)` in list loop (`internal/modules/templates/delivery/http/routes_query.go:49-52`).
- `toTemplateResponse` emits fields: `id`, `tenant_id`, `doc_type_code`, `key`, `name`, `description`, `latest_version`, `published_version_id`, `created_by`, `created_at`, `archived_at`. It does not expose `areas`, `visibility`, or `specific_areas`.
- Final success envelope: `{ "data": { "templates": [...] }, "meta": { "limit": <int>, "offset": <int> } }` (`internal/modules/templates/delivery/http/routes_query.go:54-62`).

## Summary Table

| Stage | Confirmed behavior | Evidence |
|---|---|---|
| Route mount | GET `/api/v1/templates` mounted to generated wrapper method | `internal/modules/templates/delivery/http/handler.go:40`; `internal/modules/templates/api/api.gen.go:954` |
| Dispatch | Wrapper calls `Handler.ListTemplates`; handler delegates to `listTemplates` | `internal/modules/templates/api/api.gen.go:631-634`; `internal/modules/templates/delivery/http/routes_generated.go:15-17` |
| Authz | `h.authz(..., "template.view")` called, but nil wiring installs no-op | `internal/modules/templates/delivery/http/routes_query.go:14`; `internal/modules/templates/delivery/http/handler.go:24-27`; `apps/api/cmd/metaldocs-api/main.go:329` |
| Tenant scope | Header `X-Tenant-ID` else `tenant.DevTenantID`; SQL has `tenant_id = $1` | `internal/modules/templates/delivery/http/handler.go:84-89`; `internal/modules/templates/repository/postgres.go:94` |
| Service | List is pass-through to repo | `internal/modules/templates/application/queries.go:28-30` |
| Repo query | `templates_template` SELECT with tenant, non-system-owned, optional doc_type, and `LIMIT/OFFSET`; no template-use visibility filters | `internal/modules/templates/repository/postgres.go` |
| Error envelope | Legacy `{error:{code,message}}` via `application/json`, not Problem+JSON on this path | `internal/modules/templates/delivery/http/handler.go:95-102`; `internal/platform/httpresponse/response.go:8-11`; `internal/platform/problem/problem.go:72-82` |
| Response mapping | DB row -> `domain.Template` -> `toTemplateResponse` -> `data.templates` + `meta` | `internal/modules/templates/repository/mappers.go:14-42`; `internal/modules/templates/delivery/http/routes_create.go:89-109`; `internal/modules/templates/delivery/http/routes_query.go:49-62` |
