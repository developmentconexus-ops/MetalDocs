# Feature F6.2 — Plan

> **Milestone:** 6  ·  **Feature:** `f6.2-templates-query-typed`

## Steps

1. **Add OpenAPI schemas + 200 content refs** — add `GetTemplateResponse`, `GetTemplateDocxUrlResponse`, `TemplateAuditEvent`, `ListTemplateAuditResponse` to `api/openapi/v1/openapi.yaml` components.schemas; attach `content` blocks to the three bare 200 responses.

2. **Run codegen** — `$env:GOFLAGS="-mod=mod"; go generate ./internal/modules/templates/api/...`

3. **Swap handler sites** — edit `routes_query.go` to replace all 5 `map[string]any` writeJSON calls with typed struct assignments using the generated types.

4. **Write TDD test** — create `routes_typed_response_f62_test.go` with `TestQuery_TypedResponseShape` subtests for all 5 ops.

5. **Verify** — `go build ./...`, `go test -count=1 ./internal/modules/templates/...`, `go test -count=1 ./...`, grep confirms 0 hits.

6. **Author evidence.md**

7. **Commit** — `feat(m6/f6.2): templates query typed responses + OpenAPI 200 schemas`

## Files changed

- `api/openapi/v1/openapi.yaml` — add schemas + 200 content refs
- `internal/modules/templates/api/api.gen.go` — regenerated
- `internal/modules/templates/delivery/http/routes_query.go` — handler swaps
- `internal/modules/templates/delivery/http/routes_typed_response_f62_test.go` — new test file
- `docs/superpowers/milestones/grade-a-completion/milestone-6-hs5-contract-sweep/f6.2-templates-query-typed/` — spec, plan, evidence
