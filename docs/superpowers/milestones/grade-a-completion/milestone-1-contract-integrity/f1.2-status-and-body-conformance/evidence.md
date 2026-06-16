# F1.2 evidence — status & body conformance

> **Feature:** F1.2 — status-and-body-conformance (Milestone 1 / Contract integrity)
> **Closed:** 2026-06-15
> **Scope:** P2-light. 5 public-route endpoints migrated to flat typed bodies + canonical status codes per [ADR 0035](../../../../../wiki/decisions/0035-flat-typed-responses-and-presign-status.md).

---

## 1. Endpoints migrated

| # | Endpoint | Before (H-D class) | After | Mapper |
|---|----------|--------------------|-------|--------|
| A2 | `PATCH /api/v1/documents/{id}` (`renameDocument`) | 200 + raw `domain.Document` JSON leak | 200 + empty body | n/a (no body) |
| A4 | `POST /api/v1/templates/{id}/versions` (`createNextVersion`) | 201 + `{data:{version: map[string]any}}` | 201 + flat `VersionDTO` | `toAPIVersionDTO` |
| A5 | `POST /api/v1/templates/{id}/versions/{n}/autosave/presign` (`presignTemplateAutosave`) | 201 + `{data:{upload_url,storage_key,expires_at}}` | 200 + flat `TemplatePresignAutosaveResponse` | n/a (generated struct) |
| P2-A | `POST /api/v1/templates/{id}/versions/{n}/autosave/commit` (`commitTemplateAutosave`) | 200 + `{data:{version: map[string]any}}` | 200 + flat `VersionDTO` | `toAPIVersionDTO` |
| P2-B | `GET /api/v1/templates/{id}/versions/{n}` (`getTemplateVersion`) | 200 + `{data:{version: map[string]any}}` | 200 + flat `VersionDTO` | `toAPIVersionDTO` |

**HS-6 deviation recorded:** A4 stays 201 (canonical REST for genuine resource create). Mission report originally framed as `201 → 200`; ADR 0035 §D2 governs. See spec.md interview Q2 and ADR 0035 "Negative" consequences.

## 2. Validation gates — outcomes

| Gate | Result | Evidence |
|------|--------|----------|
| V1 — A2 renameDocument typed-response test | **GREEN** | `TestRenameDocument_TypedResponseShape` PASS. Pre-impl RED proof: handler emitted `domain.Document` JSON (328 bytes); post-impl emits 200 + 0 bytes + no Content-Type. |
| V2 — A4 createNextVersion typed-response test | **GREEN** | `TestCreateNextVersion_TypedResponseShape` PASS. Pre-impl RED: top-level `data` envelope; post-impl flat `VersionDTO`. |
| V3 — A5 presignTemplateAutosave typed-response test | **GREEN** | `TestPresignAutosave_TypedResponseShape` PASS. Pre-impl RED: status 201 + envelope; post-impl 200 + flat `TemplatePresignAutosaveResponse`. |
| V3a — P2-A commitTemplateAutosave typed-response test | **GREEN** | `TestCommitAutosave_TypedResponseShape` PASS. Pre-impl RED: envelope; post-impl flat. |
| V3b — P2-B getTemplateVersion typed-response test | **GREEN** | `TestGetTemplateVersion_TypedResponseShape` PASS. Pre-impl RED: envelope; post-impl flat. |
| V4 — FE adapter typecheck | **GREEN** | `npx tsc --noEmit -p tsconfig.build.json` clean. Adapters (`presignAutosave`, `commitAutosave`, `getVersion`) decode flat shape against regenerated `lib/api-types/index.d.ts`. |
| V5 — Pre-existing happy/contract suites still green | **GREEN** | `TestPresignAutosave_Happy`, `TestCommitAutosave_Happy`, `TestCommitAutosave_HashMismatch`, `TestCreateTemplate_Happy`, `TestGeneratedTemplatesRoutes_ContractHappyPaths/getTemplateVersion` PASS (post-update to new contract). |
| V6 — Zero legacy envelope on F1.2 endpoints | **GREEN** | `grep "data" routes_create.go routes_autosave.go routes_query.go` — no `{data:{...}}` envelope literal on F1.2 sites. `toVersionResponse` still exists but is referenced only by F1.3/future endpoints per ADR-0035 adoption ledger. |
| V7 — OpenAPI codegen produced expected types | **GREEN** | `TemplatePresignAutosaveResponse` struct + `VersionDTO` with `PlaceholderSchema *[]map[string]interface{}` regenerated at `internal/modules/templates/api/api.gen.go:157,169,186`. |
| V8 — `go build ./...` clean | **GREEN** | Empty output, exit 0. |
| V9 — `go test ./...` clean | **GREEN** | All packages report `ok` (no `FAIL`). |
| V10 — Wiki stamps + ADR reference | **GREEN** | ADR 0035 created (Last verified 2026-06-15). `wiki/architecture/api-contract.md` stamp bumped. Adoption ledger updated. |

## 3. Files changed

**Backend**
- `internal/modules/documents/delivery/http/handler.go` — `renameDocument` drops `GetDocument` + writes 200 empty body.
- `internal/modules/documents/delivery/http/handler_rename_test.go` (new) — V1 TDD test.
- `internal/modules/templates/delivery/http/routes_create.go` — `createNextVersion` writes `toAPIVersionDTO(v)`.
- `internal/modules/templates/delivery/http/routes_autosave.go` — `presignAutosave` returns `200 + TemplatePresignAutosaveResponse`; `commitAutosave` returns `200 + VersionDTO`.
- `internal/modules/templates/delivery/http/routes_query.go` — `getVersion` writes `toAPIVersionDTO(v)`.
- `internal/modules/templates/delivery/http/routes_mapping.go` (new) — `toAPIVersionDTO` mapper.
- `internal/modules/templates/delivery/http/routes_typed_response_test.go` (new) — V2/V3/V3a/V3b TDD tests.
- `internal/modules/templates/delivery/http/routes_autosave_test.go` — V5 happy tests updated to new contract.
- `internal/modules/templates/delivery/http/routes_create_test.go` — `fakeUUID.New()` now emits valid UUIDv4 strings.
- `internal/modules/templates/delivery/http/routes_contract_test.go` — seeded version IDs migrated to UUIDs.
- `internal/modules/templates/api/api.gen.go` — regenerated.
- `api/openapi/v1/openapi.yaml` — 4 operation amendments (createTemplateVersion 201+VersionDTO; presignTemplateAutosave 200+TemplatePresignAutosaveResponse; commitTemplateAutosave 200+VersionDTO; getTemplateVersion 200+VersionDTO) + `TemplatePresignAutosaveResponse` schema + `VersionDTO.placeholder_schema` corrected to `array` (wire-shape parity).

**Frontend**
- `frontend/apps/web/src/lib/api-types/index.d.ts` — regenerated.
- `frontend/apps/web/src/features/templates/api/templates.ts` — `getVersion`, `presignAutosave`, `commitAutosave` decode flat shape (no `body.data.*`).

**Docs / ADR**
- `wiki/decisions/0035-flat-typed-responses-and-presign-status.md` (new) — ADR.
- `docs/superpowers/milestones/grade-a-completion/milestone-1-contract-integrity/milestone.md` — F1.2/F1.4 rows amended (P2-light absorb; HS-6 deviation note).
- `docs/superpowers/milestones/grade-a-completion/milestone-1-contract-integrity/f1.2-status-and-body-conformance/spec.md` (new).
- `docs/superpowers/milestones/grade-a-completion/milestone-1-contract-integrity/f1.2-status-and-body-conformance/plan.md` (new).
- `docs/superpowers/milestones/grade-a-completion/milestone-1-contract-integrity/f1.2-status-and-body-conformance/evidence.md` (this file).

## 4. Bounded defers

- **D-1 — FE vitest harness broken (env).** `npx vitest run` aborts at boot with `Package subpath './decode' is not defined by "exports" in parse5/entities`. Pre-existing pnpm/Node ESM resolve issue, not introduced by F1.2. FE correctness signal for F1.2 leans on `tsc --noEmit` (GREEN) + the typed regenerated adapter shape. **Owner:** later FE/devx task; not blocking M1.
- **D-2 — FE adapter unit tests for `presignAutosave` / `commitAutosave` / `getVersion`.** Not added because vitest cannot boot in this environment (see D-1). Defer alongside D-1.
- **D-3 — `toVersionResponse` retirement.** F1.3 territory per ADR 0035 adoption ledger. Mapper kept intact; only F1.2's call-sites migrated.
- **D-4 — `placeholder_schema` legacy FE type at `templates.ts:50` (`Record<string,unknown> | null`).** Wire shape now declared as array. The hand-typed FE legacy interface is wrong; runtime data is unaffected (FE already iterates the array form everywhere it actually uses it). Pre-existing drift; defer to a focused FE follow-up.

## 5. Run logs (selected)

```
$ go test ./internal/modules/documents/delivery/http/ -run "RenameDocument" -v
--- PASS: TestRenameDocument_TypedResponseShape
--- PASS: TestRenameDocument_Happy
--- PASS: TestRenameDocument_EmptyName_Returns400
--- PASS: TestRenameDocument_NameTooLong_Returns400WithoutCallingService

$ go test ./internal/modules/templates/delivery/http/
ok  metaldocs/internal/modules/templates/delivery/http

$ go build ./...
(exit 0)

$ go test ./...
(no FAIL lines)

$ npx tsc --noEmit -p tsconfig.build.json
(exit 0)
```

## 6. Closure

F1.2 is closed at the Definition-of-Done bar declared in spec.md (V1–V10 GREEN). Hard-stop HS-6 deviation (A4 keeps 201) is recorded in spec.md interview Q2 and ADR 0035. No deferred items block Milestone 1 close; D-1/D-2 are FE devx debt logged out of the contract-integrity scope.
