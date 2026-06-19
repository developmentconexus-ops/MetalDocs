# Feature F6.2 — Spec

> **Milestone:** 6  ·  **Feature:** `f6.2-templates-query-typed`  ·  **Approved:** 2026-06-19

## Goal

Eliminate all `map[string]any` response sites in `internal/modules/templates/delivery/http/routes_query.go` by replacing them with oapi-codegen-generated typed response structs, and add the missing OpenAPI 200 body schemas for the three bare ops.

## Consumer contract

| Operation | Endpoint | 200 Wire shape | Generated type |
|-----------|----------|----------------|----------------|
| `listTemplates` | GET /api/v1/templates | `{data:{templates:[TemplateDTO]}, meta:{limit:int, offset:int}}` | `templatesapi.ListTemplatesResponse` (pre-exists) |
| `GetSystemBlankTemplate` | GET /api/v1/templates/system/blank | `{template_id:uuid, template_version_id:uuid, name:string}` | `templatesapi.SystemBlankTemplateResponse` (pre-exists) |
| `getTemplate` | GET /api/v1/templates/{id} | `{data:{template:TemplateDTO, latest_version:VersionDTO}}` | `templatesapi.GetTemplateResponse` (new) |
| `getTemplateDocxUrl` | GET /api/v1/templates/{id}/versions/{n}/docx-url | `{data:{url:string}}` | `templatesapi.GetTemplateDocxUrlResponse` (new) |
| `listTemplateAudit` | GET /api/v1/templates/{id}/audit | `{data:{audit:[TemplateAuditEvent]}, meta:{limit:int, offset:int}}` | `templatesapi.ListTemplateAuditResponse` (new) |

`getVersion` (line ~181) already emits a typed `VersionDTO` — not a `map[string]any` site; left unchanged.

## Schemas added to OpenAPI

- `GetTemplateResponse` — new, added to `components/schemas`
- `GetTemplateDocxUrlResponse` — new, added to `components/schemas`
- `TemplateAuditEvent` — new, added to `components/schemas`
- `ListTemplateAuditResponse` — new, added to `components/schemas`

Also attaches `content: application/json: schema: $ref` to the three bare 200 ops:
`getTemplate`, `getTemplateDocxUrl`, `listTemplateAudit`.

## Non-goals

- FE codegen regen (F6.6)
- `getVersion` handler — already typed, not a `map[string]any` site
- Pagination changes
- Any handler outside `routes_query.go`

## Validation Gate

| # | Criterion | Command |
|---|-----------|---------|
| 1 | 0 `map[string]any` in `routes_query.go` | `grep -nE 'map\[string\]any' internal/modules/templates/delivery/http/routes_query.go` |
| 2 | OpenAPI declares 200 schemas for `getTemplate`, `getTemplateDocxUrl`, `listTemplateAudit` | `grep -n 'GetTemplateResponse\|GetTemplateDocxUrlResponse\|ListTemplateAuditResponse' api/openapi/v1/openapi.yaml` |
| 3 | BE codegen fresh | `go generate ./internal/modules/templates/api/...` exits 0 |
| 4 | Build green | `go build ./...` |
| 5 | Templates tests green | `go test -count=1 ./internal/modules/templates/...` |
| 6 | Full suite green | `go test -count=1 ./...` |
| 7 | F6.2 typed-shape test green | `go test -count=1 -run TestQuery_TypedResponseShape ./internal/modules/templates/delivery/http/...` |

**Approved:** 2026-06-19
