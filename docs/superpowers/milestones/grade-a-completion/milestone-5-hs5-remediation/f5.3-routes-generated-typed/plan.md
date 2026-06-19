# Plan — F5.3 routes_generated-typed

## Files

**Modified:**
- `internal/modules/templates/delivery/http/routes_generated.go`
  - Refactor `presignTemplateUpload(w, r, id, n, storageKey)` → return `(url, key string, ok bool)`;
    each caller writes its own op-specific typed response.
  - `PresignTemplateDocxUploadUrl` writes `templatesapi.PresignTemplateDocxUploadUrl200JSONResponse{Url: &url, StorageKey: &key}`.
  - `PresignTemplateSchemaUploadUrl` writes `templatesapi.PresignTemplateSchemaUploadUrl200JSONResponse{Url: &url, StorageKey: &key}`.
  - `PublishTemplateVersion` writes `templatesapi.PublishTemplateVersion200JSONResponse{PublishedVersionId, NextDraftId, NextDraftVersionNum}`. Drop `published_version_number` (undeclared per `openapi.yaml:1331`).

**New:**
- `internal/modules/templates/delivery/http/routes_typed_response_f53_test.go` — shape pin:
  exercises both presign endpoints + publish endpoint via the handler and asserts the JSON keys
  match exactly the declared 200 shape (no extra, no missing).

## Steps (TDD)

1. **Red:** add the shape-pin test for publish — assert `published_version_number` absent. With
   the current handler it FAILS (field still present). That is the behavioral red.
2. **Green:** apply the swap; the red test passes. Build green.
3. **Gate:** H-D grep on this file → 0; full templates suite green.

## Test strategy

The pin uses the existing httptest pattern in `routes_lifecycle_test.go`: build a handler with a
fake service that returns a known `PublishTemplateVersionResult`, drive the route, decode the
JSON, check `len(map) == 3` and that each declared key is present with the expected value.
Mirrors the surface F1.3 pinned for other operations.

For presign: drive both routes, assert the body has exactly `{url, storage_key}` with non-empty
string values; the helper returns identical values for both so a single fake suffices.
