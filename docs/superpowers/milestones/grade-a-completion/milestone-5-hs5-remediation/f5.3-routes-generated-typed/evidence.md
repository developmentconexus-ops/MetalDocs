# Evidence — F5.3 routes_generated-typed (H-D sites)

> **Status:** CLOSED 2026-06-16 · structural lift from `map[string]any` to strict-server generated 200 types.

## Change

| File | Change |
|------|--------|
| `internal/modules/templates/delivery/http/routes_generated.go:97-103` | `PresignTemplateDocxUploadUrl` / `PresignTemplateSchemaUploadUrl` each write their op-specific generated 200 type (`templatesapi.PresignTemplateDocxUploadUrl200JSONResponse` / `PresignTemplateSchemaUploadUrl200JSONResponse`). |
| `internal/modules/templates/delivery/http/routes_generated.go:128 (was)` | Removed `map[string]any{"url":…, "storage_key":…}`. Helper refactored to return `(url, key string, ok bool)`; the per-op write happens in the caller using the strict-server typed struct. |
| `internal/modules/templates/delivery/http/routes_generated.go:238 (was)` | `PublishTemplateVersion` now writes `templatesapi.PublishTemplateVersion200JSONResponse{PublishedVersionId, NextDraftId, NextDraftVersionNum}` — **drops the undeclared `published_version_number`** field per `openapi.yaml:1331`. M1/F1.3 declared-fields-only leak closed at the wire-shape source. |
| `internal/modules/templates/delivery/http/routes_typed_response_f53_test.go` | **new** — shape pins for both presign endpoints (route-level `httptest`) and for the publish strict-server response struct (unit-level JSON-key assertion; the handler binding to that struct is the compile-time guarantee). |

## TDD record

Refactor-class for presign (wire shape identical before/after — both old `map[string]any{"url":…, "storage_key":…}` and the new typed struct serialize to the same JSON). The strict-server type binding is what changes — the contract is now compiler-enforced.

For publish the behavior change is intentional and spec-correct: dropping the undeclared
`published_version_number`. The unit test
`TestPublishTemplateVersion200JSONResponse_DeclaredKeysOnly` asserts the generated type emits
exactly the 3 declared fields and explicitly fails if `published_version_number` reappears —
this is the regression guard. Honest labeling: the *runtime red* was the H-D grep (showed
non-zero before, 0 after); the assertion test is a forward-looking guard, not a red→green proof
of behavior change (the strict-server struct does not have a `PublishedVersionNumber` field at
all, so the leak is impossible to re-introduce without first amending `openapi.yaml` + re-running
`gen.go`).

## Validation Gate results (real output)

1. **H-D grep on the file → 0:**
   `grep -n "map\[string\]any" internal/modules/templates/delivery/http/routes_generated.go`
   → **0 matches** (exit 1). Combined with M1/F1.2/F1.3 already-closed scope, F5.3 closes both
   remaining H-D sites; **H-D = 0 in this file.**
2. **Shape pins green:**
   - `TestPresignTemplateDocxUploadUrl_TypedResponseShape` → `PASS`
   - `TestPresignTemplateSchemaUploadUrl_TypedResponseShape` → `PASS`
   - `TestPublishTemplateVersion200JSONResponse_DeclaredKeysOnly` → `PASS`
   (Confirmed before AND after the swap — the strict-server struct's shape is the spec; pinning it
   was the durable contract.)
3. **Build** — `go build ./...` → `BUILD OK`.
4. **Module suite** — `go test -count=1 ./internal/modules/templates/...` → all packages `ok`
   (api/application/delivery/http/domain/infrastructure/repository).

## Fixture-vs-real

Pure unit + route-level `httptest` against the existing fake repo + real `application.Service`.
No live SQL involved (the H-D family is wire-shape only). No fixture stood in for a live path.

## Defers

None. Both cited H-D sites in `routes_generated.go` closed at the file. Mention-don't-fix: the
file `routes_create.go:59` (TemplateDTO mapping) still references `published_version_number` —
that is the **TemplateDTO** field (declared in OpenAPI, see `index.d.ts:2640`), unrelated to the
publish 200 leak this feature closed; not in F5.3 scope, not an H-D finding.
