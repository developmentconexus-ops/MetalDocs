# Data-flow Trace — PUT /api/v2/templates/{id}/versions/{n}/schema

## Task

For operation `UNVERIFIED` (HTTP `PUT /api/v2/templates/{id}/versions/{n}/schema`) in module `internal/modules/templates_v2`, trace call flow and runtime behavior.

### 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | `UNVERIFIED (path/op not declared)` | `api/openapi/v1/openapi.yaml:2857`, `api/openapi/v1/partials/templates.yaml:77` |
| Generated server stub | `n/a — no generated PUT /schema stub found` | `internal/modules/templates_v2/api/api.gen.go:584`, `internal/modules/templates_v2/api/api.gen.go:1216` |
| Handler | `Handler.updateSchemas` | `internal/modules/templates_v2/delivery/http/routes_schema.go:11` |
| Route registration | `mux.HandleFunc("PUT /api/v2/templates/{id}/versions/{n}/schema", h.updateSchemas)` | `internal/modules/templates_v2/delivery/http/handler.go:49` |

### 2. Call chain

1. `internal/modules/templates_v2/delivery/http/routes_schema.go:11` `Handler.updateSchemas` — parse path vars and request body; invoke authz callback.
   ? calls: `internal/modules/templates_v2/delivery/http/routes_schema.go:21` `h.authz(r, tenantID, "*", "template.edit")`

2. `internal/modules/templates_v2/delivery/http/routes_schema.go:36` `h.svc.UpdateSchemas(...UpdateSchemasCmd)` — enter application layer.
   ? calls: `internal/modules/templates_v2/application/schema.go:20` `Service.UpdateSchemas`

3. `internal/modules/templates_v2/application/schema.go:21` `s.repo.GetVersion` — load target version.
   ? calls: `internal/modules/templates_v2/repository/postgres.go:189` `Repository.GetVersion`
   ? DB: `internal/modules/templates_v2/repository/postgres.go:199` `QueryRowContext`

4. `internal/modules/templates_v2/application/schema.go:32` `ValidatePlaceholders` — validate placeholder schema.
   ? calls: `internal/modules/templates_v2/application/schema.go:84` `ValidatePlaceholders`

5. `internal/modules/templates_v2/application/schema.go:35` resolver registry check — only if `s.resolvers != nil`; validate `resolver_key` exists in registry.
   ? calls: `internal/modules/templates_v2/application/schema.go:36` `s.resolvers.Known()`

6. `internal/modules/templates_v2/application/schema.go:58` `s.repo.UpdateVersion` — persist metadata/placeholder schema on version row.
   ? calls: `internal/modules/templates_v2/repository/postgres.go:229` `Repository.UpdateVersion`
   ? DB: `internal/modules/templates_v2/repository/postgres.go:253` `ExecContext` (`UPDATE templates_v2_template_version ... WHERE id = $1`)

7. `internal/modules/templates_v2/application/schema.go:62` `s.repo.AppendAudit` — append audit event.
   ? calls: `internal/modules/templates_v2/repository/postgres.go:311` `Repository.AppendAudit`
   ? DB: `internal/modules/templates_v2/repository/postgres.go:323` `ExecContext` (`INSERT INTO templates_v2_audit_log ...`)

8. `internal/modules/templates_v2/delivery/http/routes_schema.go:50` `writeJSON` — return 200 JSON envelope with `data.version`.

- Transaction boundary: `none found` (`Repository` uses `*sql.DB` and direct `QueryRowContext`/`ExecContext`; no `BeginTx` in this flow) (`internal/modules/templates_v2/repository/postgres.go:22`, `:199`, `:253`, `:323`).
- Authz calls in chain: handler callback only (`h.authz`) (`internal/modules/templates_v2/delivery/http/routes_schema.go:21`). No `authz.Require`/`authz.RequireAll` in traced files (`UNVERIFIED elsewhere outside traced flow`).
- Idempotency interactions: `none found for this route` (no idempotency middleware/handler usage in traced endpoint).

### 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `TemplateVersion.status` | `draft required` | `unchanged` | `UpdateSchemas` enforces draft and updates schemas only | `handler action "template.edit" via h.authz` (`internal/modules/templates_v2/application/schema.go:25`, `:55-56`; `internal/modules/templates_v2/delivery/http/routes_schema.go:21`) |

### 4. SQL touched

| File:line | Verb | Table(s) | Auth-area arg (if any) |
|---|---|---|---|
| `internal/modules/templates_v2/repository/postgres.go:190-197` | `SELECT` | `templates_v2_template_version` | none |
| `internal/modules/templates_v2/repository/postgres.go:236-252` | `UPDATE` | `templates_v2_template_version` | none |
| `internal/modules/templates_v2/repository/postgres.go:318-322` | `INSERT` | `templates_v2_audit_log` | none |

Tripwire pairing (authz before mutate SQL on same tx):
- `AUTHZ CHECK`: `internal/modules/templates_v2/delivery/http/routes_schema.go:21` (`h.authz(...)`)
- `MUTATING SQL`: `internal/modules/templates_v2/repository/postgres.go:253` (`ExecContext` UPDATE)
- `authz.Require` pairing: `VIOLATION` (no `authz.Require` call in traced flow before UPDATE; no same-tx authz guard).

### 5. Response shape

- 2xx schema ref: `UNVERIFIED` (operation missing from OpenAPI paths searched: `api/openapi/v1/openapi.yaml:2857`, `api/openapi/v1/partials/templates.yaml:77`).
- Runtime success envelope: legacy JSON `{"data":{"version":...}}` via `writeJSON` (`internal/modules/templates_v2/delivery/http/routes_schema.go:50-54`).
- Runtime error envelope: legacy JSON `{"error":{"code","message"}}` via `writeErr`/`writeMappedErr` (`internal/modules/templates_v2/delivery/http/handler.go:95-102`, `:108-118`; `internal/modules/templates_v2/delivery/http/routes_schema.go:17`, `:22`, `:32`, `:46`).
- RFC 9457: `not used in this endpoint`; RFC 9457 package exists globally (`internal/platform/problem/problem.go:1`, `:72`, `:79`) but not invoked in traced handler.
- Declared error statuses + Problem type URIs for this op: `UNVERIFIED` (op absent in OpenAPI files above).

### 6. Cross-references

- Authz wiring default (`AUTHZ BYPASSED` risk): `Handler.New` replaces nil authz with allow-all no-op (`internal/modules/templates_v2/delivery/http/handler.go:24-27`).
- ValidatePlaceholders catalog rule: fixed catalog of 7 names enforced by `placeholderCatalogSet` membership check (`internal/modules/templates_v2/application/schema.go:79-82`, `:103-105`).
- ValidatePlaceholders also enforces name regex + uniqueness + computed/name coupling:
  - name regex `^[a-z][a-z0-9_]{0,49}$` (`internal/modules/templates_v2/application/schema.go:77`, `:96-98`)
  - duplicate id/name rejected (`internal/modules/templates_v2/application/schema.go:91-93`, `:99-101`)
  - named placeholders must be `PHComputed` with `resolver_key == name` (`internal/modules/templates_v2/application/schema.go:106-108`)
  - additional constraint rejections: invalid regex/date/min-max/max_length/computed-resolver_key (`internal/modules/templates_v2/application/schema.go:110-136`)
  - select options rules in service: select requires options; non-select forbids options (`internal/modules/templates_v2/application/schema.go:47-52`)
- Domain placeholder types include fixed 7-type constants: `PHText`, `PHDate`, `PHNumber`, `PHSelect`, `PHUser`, `PHPicture`, `PHComputed` (`internal/modules/templates_v2/domain/schemas.go:13-20`).
- Catalog concept cross-reference: wiki states fixed 7-token catalog and backend rejects non-catalog names (`wiki/concepts/placeholders.md:27`, `:47`).
- Concurrency model:
  - pre-write stale check uses `ExpectedContentHash` vs loaded version hash (`internal/modules/templates_v2/application/schema.go:28-30`)
  - UPDATE predicate is only `WHERE id = $1` (no content-hash/version guard in SQL) (`internal/modules/templates_v2/repository/postgres.go:252`).
- Audit emission: yes, `AppendAudit` called with `Action: domain.AuditSaved` and `Details.kind = "schema"` (`internal/modules/templates_v2/application/schema.go:62-69`).
- Repository `SaveSchema`: `UNVERIFIED` (symbol not found in `internal/modules/templates_v2/repository/postgres.go`; available method is `UpdateVersion` at `:229`).
- ResolverRegistryReader runtime wiring in `main.go`: `UNVERIFIED`.
  - only `cmd/*/main.go` found is seed utility (`cmd/seed-test-document/main.go`).
  - service constructor leaves `resolvers` nil when not passed (`internal/modules/templates_v2/application/service.go:11-16`), and resolver-key registry validation is conditional on `s.resolvers != nil` (`internal/modules/templates_v2/application/schema.go:35`).

## Output

operation id `UNVERIFIED` · layer count in §2: `8` · tripwire pairing `VIOLATION`
