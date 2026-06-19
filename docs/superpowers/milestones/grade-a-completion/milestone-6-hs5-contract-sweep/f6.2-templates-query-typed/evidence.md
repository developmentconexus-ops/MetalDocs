# Feature F6.2 — Evidence

> **Milestone:** 6  ·  **Feature:** `f6.2-templates-query-typed`  ·  **Date:** 2026-06-19

## Validation Gate Results

| # | Criterion | Result |
|---|-----------|--------|
| 1 | 0 `map[string]any` in `routes_query.go` | PASS — 0 hits |
| 2 | OpenAPI declares 200 schemas for `getTemplate`, `getTemplateDocxUrl`, `listTemplateAudit` | PASS |
| 3 | BE codegen fresh | PASS — `go generate ./internal/modules/templates/api/...` exits 0 |
| 4 | Build green | PASS — `go build ./...` exits 0 |
| 5 | Templates tests green | PASS — `go test -count=1 ./internal/modules/templates/...` all ok |
| 6 | Full suite green | PASS — `go test -count=1 ./...` all ok |
| 7 | F6.2 typed-shape test green | PASS — `TestQuery_TypedResponseShape` 5/5 subtests pass |

## Grep confirmation

```
grep -nE 'map\[string\]any' internal/modules/templates/delivery/http/routes_query.go
(no output — 0 hits)
```

## Changes made

### `api/openapi/v1/openapi.yaml`
- Added 4 new component schemas: `GetTemplateResponse`, `GetTemplateDocxUrlResponse`, `TemplateAuditEvent`, `ListTemplateAuditResponse`
- Added `content: application/json: schema:` blocks to 3 bare 200 responses: `getTemplate`, `getTemplateDocxUrl`, `listTemplateAudit`

### `internal/modules/templates/api/api.gen.go`
- Regenerated via `go generate ./internal/modules/templates/api/...`
- New types emitted: `GetTemplateResponse`, `GetTemplateDocxUrlResponse`, `TemplateAuditEvent`, `ListTemplateAuditResponse`

### `internal/modules/templates/delivery/http/routes_query.go`
- All 5 `map[string]any` writeJSON sites replaced with typed struct assignments
- `time` import removed (OccurredAt is now `time.Time` in codegen type, assigned directly)
- `github.com/google/uuid` import added (required for UUID parsing in `GetSystemBlankTemplate` and `listAudit`)
- `listAudit`: inline `[]map[string]any` slice replaced with `[]templatesapi.TemplateAuditEvent`

### `internal/modules/templates/delivery/http/routes_query_test.go`
- Pre-existing test `TestGetSystemBlankTemplate_RequiresTemplateViewAuthz`: version ID fixed from `"blank-ver-1"` (non-UUID) to `"22222222-2222-4222-8222-222222222222"` — required because handler now calls `uuid.Parse(ver.ID)` where it previously passed the string verbatim through `map[string]any`

### `internal/modules/templates/delivery/http/routes_typed_response_f62_test.go`
- New test file — `TestQuery_TypedResponseShape` with 5 subtests: `list_templates`, `get_system_blank`, `get_template`, `get_docx_url`, `list_audit`
- Reuses `decodeStrict[T]` from f61 test (same package)

## Notes

- `getVersion` handler left unchanged — it already emits a typed `VersionDTO` directly (not a `map[string]any` site)
- `TemplateAuditEvent.OccurredAt` in codegen is `time.Time` (not string); `event.OccurredAt.UTC()` assigned directly
- `TemplateAuditEvent.Details` is `*map[string]interface{}`; domain `event.Details` is `map[string]any` — cast directly when non-nil
- UUID parsing errors in `listAudit` return 500 with `INTERNAL_ERROR`; consistent with `toAPITemplateDTO`/`toAPIVersionDTO` pattern
