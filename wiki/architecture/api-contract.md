# API Contract (Spec-as-Source-of-Truth)

> **Operational guide.** For the design system contract (error envelope, pagination, idempotency, two-tier authz, list filtering) see [`architecture/api-design-system.md`](api-design-system.md).

> **Last verified:** 2026-05-12
> **Scope:** OpenAPI spec location, backend codegen (oapi-codegen v2), frontend codegen (openapi-typescript v7), runtime enforcement gaps, CI drift guard, per-module migration status.
> **Out of scope:** Auth/IAM mechanics (`modules/iam.md`), approval-specific request shapes (`modules/approval.md`), frontend API call patterns (`architecture/frontend-structure.md §7`).
> **Key files:**
> - `api/openapi/v1/openapi.yaml:1` — single source of truth; OpenAPI 3.0.3
> - `redocly.yaml:1` — lint config (pre-existing rule suppressions documented inline)
> - `internal/modules/registry/api/cfg.yaml:1` — registry codegen config (include-tags: registry)
> - `internal/modules/registry/api/gen.go:1` — `//go:generate` invocation for registry
> - `internal/modules/registry/api/api.gen.go:1` — generated; DO NOT EDIT
> - `internal/modules/templates_v2/api/cfg.yaml:1` — templates codegen config (include-tags: templates)
> - `internal/modules/templates_v2/api/gen.go:1` — `//go:generate` invocation for templates_v2
> - `internal/modules/templates_v2/api/api.gen.go:1` — generated; DO NOT EDIT
> - `internal/modules/documents/api/cfg.yaml:1` — documents codegen config (include-tags: documents)
> - `internal/modules/documents/api/gen.go:1` — `//go:generate` invocation for documents (bootstrap only)
> - `internal/modules/documents/api/api.gen.go:1` — generated; DO NOT EDIT
> - `internal/modules/documents/approval/http/contracts/strictjson.go:23` — `Decode` helper; `DisallowUnknownFields` pattern used at handler boundaries
> - `internal/modules/registry/delivery/http/handler.go:72` — `ServerInterfaceWrapper` wiring pattern (registry)
> - `internal/modules/templates_v2/delivery/http/handler.go:32` — `ServerInterfaceWrapper` wiring pattern (templates_v2)
> - `migrations/0183_documents_name_not_empty.sql:27` — DB invariant floor for `documents.name`
> - `.github/workflows/api-contract.yml:1` — CI drift guard (3 jobs)
> - `frontend/apps/web/package.json:13` — `gen:api` script (`openapi-typescript`)

---

## 1. Spec location

`api/openapi/v1/openapi.yaml` is the **single source of truth** for all MetalDocs HTTP contracts. OpenAPI 3.0.3. Path prefix conventions: `/api/v1` (platform/auth) and `/api/v2` (business modules).

New endpoints MUST be authored in the spec first. Handlers implement; spec governs.

---

## 2. Backend codegen — oapi-codegen v2

Each migrated module has an `internal/modules/<x>/api/` directory with three files:

| File | Purpose |
|------|---------|
| `cfg.yaml` | Codegen config: package name, output file, `include-tags` filter |
| `gen.go` | Single-line `//go:generate` comment — no production code |
| `api.gen.go` | Generated output — **never hand-edit** |

The generated file provides:
- Go request/response types for all operations in scope.
- `ServerInterface` — one method per operation; handler struct must implement.
- `ServerInterfaceWrapper` — stdlib `net/http` adapter that parses path/query params and calls `ServerInterface` methods.
- `StrictServerInterface` (line ~1608 in registry gen) — higher-level variant where input/output are typed structs; handler returns `(ResponseObject, error)` instead of writing to `http.ResponseWriter`.

**Regenerate:**

```bash
GOFLAGS=-mod=mod go generate ./internal/modules/registry/api/...
GOFLAGS=-mod=mod go generate ./internal/modules/templates_v2/api/...
GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...
```

Use `GOFLAGS=-mod=mod` when the project vendor directory is present; otherwise `go generate` will refuse to fetch the `oapi-codegen` binary dependency.

CI runs `go generate ./...` — see `api-contract.yml:27`.

---

## 3. Handler wiring pattern

Handlers do **not** implement `StrictServerInterface` directly. The current pattern uses `ServerInterfaceWrapper`:

```go
// internal/modules/registry/delivery/http/handler.go:72
generated := registryapi.ServerInterfaceWrapper{
    Handler: h,
    ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
        httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
    },
}
mux.HandleFunc("GET /api/v2/controlled-documents", generated.ListControlledDocuments)
// …
```

The handler struct (`*Handler`) implements `ServerInterface`; the wrapper handles route dispatch and param parsing.

---

## 4. Runtime enforcement gaps

oapi-codegen does **not** enforce:

- **Unknown fields:** Use `contracts.Decode` from `internal/modules/documents/approval/http/contracts/strictjson.go:23`. It sets `decoder.DisallowUnknownFields()`. Call it instead of `json.NewDecoder(r.Body).Decode(...)` at handler boundaries.
- **Required fields:** oapi-codegen generates pointer fields for optional and value fields for required, but does not produce 400 responses for missing required fields at runtime. Handlers must check explicitly (e.g., `missingAtomicCreateField` at `internal/modules/registry/delivery/http/routes.go:102`).

---

## 5. DB invariant floor

`documents.name` is enforced non-empty at the database level as a defense-in-depth layer below the HTTP contract:

```sql
-- migrations/0183_documents_name_not_empty.sql:27
ALTER TABLE documents
  ALTER COLUMN name SET NOT NULL,
  ADD CONSTRAINT documents_name_not_empty CHECK (length(trim(name)) > 0);
```

This prevents silent data corruption even if a future handler regression bypasses the spec-generated struct.

---

## 6. Frontend codegen — openapi-typescript v7

```bash
# frontend/apps/web
pnpm gen:api
# equivalent: openapi-typescript ../../../api/openapi/v1/openapi.yaml -o src/lib/api-types/index.d.ts
```

Output: `frontend/apps/web/src/lib/api-types/index.d.ts` — **never hand-edit**. The `api` client in `lib/api/client.ts` is typed against these generated `paths`.

---

## 7. CI drift guard

`.github/workflows/api-contract.yml` runs on every PR touching spec, generated files, or package manifests. Three jobs:

| Job | What it checks |
|-----|---------------|
| `backend-codegen-drift` | Runs `go generate ./...`; fails if `**/api.gen.go` has uncommitted changes |
| `frontend-codegen-drift` | Runs `npm run gen:api`; fails if `src/lib/api-types/` has uncommitted changes |
| `openapi-lint` | Runs `redocly lint` against the spec; config in `redocly.yaml` |

Pre-existing lint rule violations (133 errors at time of introduction) are suppressed in `redocly.yaml` pending a cleanup ticket. New violations from changed paths will still fail CI.

---

## 8. Legacy migration notes (superseded by Plan 8 baseline)

| Module | Path prefix | Codegen status | Handler migration |
|--------|------------|----------------|------------------|
| `registry` | `/api/v2/controlled-documents` | Full (`include-tags: registry`) | Complete (commits `aa867b6c`, `9fccd8e7`) |
| `templates_v2` | `/api/v2/templates` | Full (`include-tags: templates`) | Complete (commit `f7f9c58d`) |
| `documents` | `/api/v2/documents` | Bootstrap only (`include-tags: documents`) | Deferred — see below |
| `approval` | `/api/v2/approvals` | No spec coverage | Not started |
| `taxonomy` | `/api/v2/taxonomy` | No spec coverage | Not started |
| `iam` | `/api/v2/iam` | No spec coverage | Not started |
| `platform` | `/api/v1/auth`, `/api/v1/feature-flags` | No spec coverage | Not started |

**Documents module note:** codegen bootstrap landed (commit `81e7ec23`) — `api.gen.go` is generated and up to date. Handler migration is blocked by spec-handler drift (missing spec ops for `renameDocument`, `duplicateDocument`, comments CRUD; orphaned spec ops with no handler). Details and migration template: `wiki/backlog/contract-first-followups.md`.

---

## 9. Module migration status (Plan 8 baseline)

| Module | Path prefix | Contract status | Notes |
|--------|------------|-----------------|-------|
| `documents` | `/api/v2/documents` | `Partial` | Mixed aligned+raw surface; includes one path signature mismatch (`{version}` vs `{versionNum}`) |
| `approval` | `/api/v2/approval`, `/api/v2/documents/* approval routes` | `Raw` | Runtime routes are mounted but not yet represented in OpenAPI/codegen |
| `templates_v2` | `/api/v2/templates`, `/api/v2/signed` | `Partial` | Core generated routes aligned; several runtime routes still spec-missing |
| `registry` | `/api/v2/controlled-documents` | `Wrapper-only` | Fully mounted through `ServerInterfaceWrapper` and aligned with spec/codegen |
| `taxonomy` | `/api/v2/taxonomy` | `Raw` | Runtime routes present; no OpenAPI coverage yet |
| `audit` | `/api/v1/audit` | `Partial` | Runtime path aligns with spec (`/audit/events` + `/api/v1` server prefix); operationId missing |
| `iam` | `/api/v1/iam`, `/api/v2/iam` | `Partial` | v1 admin routes aligned; v2 area-membership routes are spec-missing |
| `auth/platform` | `/api/v1/auth` | `Partial` | Auth runtime routes align with spec paths, but operationIds are missing |

### Route truth tables

- [Documents route table](../modules/documents.md#api-route-truth-table-plan-8-baseline)
- [Approval route table](../modules/approval.md#api-route-truth-table-plan-8-baseline)
- [Templates_v2 route table](../modules/templates_v2.md#api-route-truth-table-plan-8-baseline)
- [Registry route table](../modules/registry.md#api-route-truth-table-plan-8-baseline)
- [Taxonomy route table](../modules/taxonomy.md#api-route-truth-table-plan-8-baseline)
- [Audit route table](../modules/audit.md#api-route-truth-table-plan-8-baseline)
- [IAM route table](../modules/iam.md#api-route-truth-table-plan-8-baseline)
- [Auth route table](../modules/auth.md#api-route-truth-table-plan-8-baseline)

---

## 10. Adding a new module

1. Author operations in `api/openapi/v1/openapi.yaml` with a new `tags: [<module>]` value.
2. Lint: `npx @redocly/cli lint api/openapi/v1/openapi.yaml`.
3. Create `internal/modules/<x>/api/cfg.yaml`:
   ```yaml
   package: <x>api
   generate:
     models: true
     std-http-server: true
     strict-server: true
     embedded-spec: true
   output: api.gen.go
   output-options:
     include-tags:
       - <module-tag>
   ```
4. Create `internal/modules/<x>/api/gen.go`:
   ```go
   package <x>api
   //go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml ../../../../api/openapi/v1/openapi.yaml
   ```
5. Run `GOFLAGS=-mod=mod go generate ./internal/modules/<x>/api/...`.
6. Implement `ServerInterface` on the handler struct; wire via `ServerInterfaceWrapper` (registry pattern at `handler.go:72`).
7. Commit `api.gen.go` — CI drift check will verify it stays in sync.

---

## See also

- `wiki/architecture/backend-api-structure.md` - canonical backend/API structure rules and migration discipline
- `wiki/decisions/0012-contract-first-api.md` - ADR: why spec-as-source-of-truth was adopted and root cause of the `documents.name` bug
- `wiki/backlog/contract-first-followups.md` - deferred handler migrations + documents spec/handler gap inventory
- `wiki/references/oapi-codegen.md` - operational how-to (regenerate, vendor mode, add module)
- `wiki/architecture/frontend-structure.md §7` - frontend API call patterns using generated types

