# Feature F6.1 — Evidence

> **Milestone:** 6 — HS-5 Contract Sweep  ·  **Feature:** `f6.1-templates-lifecycle-typed`  ·  **Closed:** 2026-06-19
> **Contract:** `spec.md` (consumer contract + Validation Gate).

## What was implemented

- `api/openapi/v1/openapi.yaml`:
  - Added 4 component schemas: `TemplateVersionEnvelope`, `ArchiveTemplateResponse`, `TemplateApprovalConfig`, `UpsertTemplateApprovalConfigResponse`.
  - Attached `200.content.application/json.schema: $ref` to: `submitTemplateVersion`, `reviewTemplateVersion`, `archiveTemplate`, `upsertTemplateApprovalConfig`.
  - Tightened `ApproveTemplateVersionResponse` — promoted `next_draft` from optional to required (still `nullable: true`) so codegen emits the field without `omitempty`, preserving the wire-shape `"next_draft": null` on reject paths.
- `internal/modules/templates/api/api.gen.go`: regenerated via `GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/...`. Added types: `TemplateVersionEnvelope`, `ArchiveTemplateResponse`, `TemplateApprovalConfig`, `UpsertTemplateApprovalConfigResponse`, plus the matching strict-server `*200JSONResponse` aliases.
- `internal/modules/templates/delivery/http/routes_lifecycle.go`: swapped 5 `writeJSON(...map[string]any...)` sites to typed `*Response` constructions. `submitForReview` → `TemplateVersionEnvelope`. `review` → `TemplateVersionEnvelope`. `approve` → `ApproveTemplateVersionResponse` with anonymous-struct `NextDraft` pointer (parsed via `uuid.Parse(res.NextDraft.ID)`). `archiveTemplate` → `ArchiveTemplateResponse`. `upsertApprovalConfig` → `UpsertTemplateApprovalConfigResponse` (parses `cfg.TemplateID` via `uuid.Parse`).
- `internal/modules/templates/delivery/http/routes_typed_response_f61_test.go` (new): `TestLifecycle_TypedResponseShape` with 6 subtests (submit / review / approve_accept_publish / approve_reject / archive / upsert_approval_config). Each subtest decodes the response body with `json.NewDecoder + DisallowUnknownFields()` into the matching generated `*Response` type and asserts a typed field round-trips non-zero.

Producer matches consumer contract: the OpenAPI 200 schemas + the generated Go types are now the single source of truth; the handler emits exactly those types; the contract test decodes strict against them.

Commit: pending (single commit at end of F6.1 — see below).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Codegen regen | `GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/...` then `git diff --stat internal/modules/templates/api/api.gen.go` | `1 file changed, 149 insertions(+), 94 deletions(-)`. Manual inspection of new types confirms field names + types match expectation (`TemplateVersionEnvelope.Data.Version: VersionDTO`, `TemplateApprovalConfig.TemplateId: openapi_types.UUID`, etc.). | real |
| TDD failing-test order (honest record) | `go test -count=1 -run 'TestLifecycle_TypedResponseShape' ./internal/modules/templates/delivery/http/...` BEFORE handler swap | **PASS** (`ok metaldocs/internal/modules/templates/delivery/http 2.739s`). Interpretation: today's `map[string]any` literals already emit wire-byte-identical JSON to the new typed types — the spec interview (Q1) anticipated this. The test was authored to be drift-future-proof (strict-decode into the generated type) rather than to fail-then-pass; the handler swap below converts the proof from "wire happens to match" to "code path is typed at the language level". Documented for the validator: this is wire-already-correct structural retype, not a regression-test added after a bug. | real |
| Static (build) | `go build ./...` | clean (no output) | real |
| Targeted test (templates module) | `go test -count=1 ./internal/modules/templates/...` | `ok metaldocs/internal/modules/templates/application 1.337s`; `ok .../delivery/http 2.863s`; all packages PASS. | real |
| H-D Grep A on touched file | `grep -nE 'map\[string\]any' internal/modules/templates/delivery/http/routes_lifecycle.go` | 0 hits | real |
| Whole-repo regression | `go test -count=1 ./...` | All packages `ok`; 0 FAIL lines in output. | real |
| Wire-shape parity sentinels (existing tests) | `go test -count=1 -run 'TestSubmitForReview\|TestReview_Accept\|TestApprove_\|TestArchiveTemplate_\|TestUpsertApprovalConfig_' ./internal/modules/templates/delivery/http/...` | All green (the existing `routes_lifecycle_test.go` decoders are unchanged — `data.version.status`, `data.next_draft`, `data.template.archived_at`, `data.approval_config.approver_role`). | real |
| Runtime proof — wire JSON shape | Existing `routes_lifecycle_test.go` happy-paths exercise the full mux through `httptest.NewRecorder`; bodies are decoded into structs that mirror the wire keys/values. Combined with the new strict-decode F6.1 test, this is end-to-end wire proof at the handler boundary. | Real-handler-via-mux. NOT a live HTTP roundtrip against a deployed binary — labeled as `real (fixture-DB, real-handler-via-mux)`. The runtime emission path (Go encoding/json over the typed struct) is exactly what the deployed binary would execute. | real (handler-level) |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| 1. 5 handlers serialize generated `*Response`/envelope; 0 `map[string]any` in `routes_lifecycle.go` | YES | grep row above |
| 2. OpenAPI declares 200 body schemas for all 5 ops | YES | YAML edits to lines 1497–1504 (submit), 1532–1539 (review), 1588–1595 (archive), 1610–1617 (upsert); approve already aligned + tightened next_draft required | real |
| 3. BE codegen fresh — no uncommitted diff after regen | YES | `git diff --exit-code internal/modules/templates/api/api.gen.go` exit 0 after final regen; diff captured before initial commit | real |
| 4. Build + existing tests green from clean | YES | `go build ./...` clean; `go test -count=1 ./...` 0 FAIL | real |
| 5. Wire-shape sentinel tests pass without decoder modification | YES | targeted run row above | real |
| 6. NEW typed-shape test green | YES | F6.1 test row above | real (with TDD-order caveat documented) |

## Review disposition

- **Spec-compliance review:** PASS — every Validation Gate criterion has a re-runnable command. No scope creep: only the 5 cited sites + the 4 schema declarations + the 1 ApproveTemplate next_draft `required` tightening (in-scope: necessary to preserve wire-byte-identical output). No FE codegen touched in F6.1 (deferred to F6.6 per plan).
- **Code-quality review:** PASS — typed-struct construction follows the existing `routes_generated.go:74` pattern (`var resp X; resp.Data.Foo = dto; writeJSON(...)`). UUID parse follows existing `routes_mapping.go` convention. Anonymous-struct `NextDraft` literal uses the codegen-emitted exact field shape; no shadow type added.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| (none) | — | — |
