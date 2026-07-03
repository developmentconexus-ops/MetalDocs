# API Contract (Spec-as-Source-of-Truth)

> **Operational guide.** For the design system contract (error envelope, pagination, idempotency, two-tier authz, list filtering) see [`architecture/api-design-system.md`](api-design-system.md).

> **Last verified:** 2026-07-02 (approval document-mutation cluster wire-truth repair: 6 operations —
> `recordDocumentSignoff`, `publishDocument`, `scheduleDocumentPublish`, `obsoleteDocument`,
> `cancelDocumentApproval`, `supersedeDocument` — got declared `requestBody`/200 schemas matching
> `approval/http/contracts` (ADR 0035 flat bodies); required `If-Match` header declared on the 5
> operations whose handlers hard-require it; BE (`approval/api`) + FE (`gen:api`) codegen regenerated.
> Prior: 2026-06-20 M8 / F8.6 — §5b H-D gate **widened to the full public-route surface**
> (presence, observability, approval/http — not just `delivery/http/`); recorded health + declared-dynamic
> metrics exemptions; **mechanical CI guard** added (`tools/cilint` `noresponsemap`, laundering-resistant).
> Prior: 2026-06-20 M7 / F7.1–F7.5 — HS-2 contract completion: audit/auth/search/documents 200 bodies typed
> (generated models + ADR-0012 hand-rolled structs); 4 documents 200 schemas declared + BE/FE codegen
> regenerated; honest two-part H-D gate defined in §5b. Prior: 2026-06-19 M6)
> **Scope:** OpenAPI spec location, backend codegen (oapi-codegen v2), frontend codegen (openapi-typescript v7), runtime enforcement gaps, CI drift guard, and freeze-law contract checks.
> **Out of scope:** Auth/IAM mechanics (`modules/iam.md`), approval-specific request shapes (`modules/approval.md`), frontend API call patterns (`architecture/frontend-structure.md section 7`).
> **Key files:**
> - `api/openapi/v1/openapi.yaml:1` - single source of truth; OpenAPI 3.0.3
> - `redocly.yaml:1` - lint config (pre-existing rule suppressions documented inline)
> - `internal/modules/controlleddocuments/api/cfg.yaml:1` - controlled-documents codegen config (include-tags: controlled-documents)
> - `internal/modules/controlleddocuments/api/gen.go:1` - `//go:generate` invocation for controlled-documents
> - `internal/modules/controlleddocuments/api/api.gen.go:1` - generated; DO NOT EDIT
> - `internal/modules/templates/api/cfg.yaml:1` - templates codegen config (include-tags: templates)
> - `internal/modules/templates/api/gen.go:1` - `//go:generate` invocation for templates
> - `internal/modules/templates/api/api.gen.go:1` - generated; DO NOT EDIT
> - `internal/modules/documents/api/cfg.yaml:1` - documents codegen config (include-tags: documents)
> - `internal/modules/documents/api/gen.go:1` - `//go:generate` invocation for documents
> - `internal/modules/documents/api/api.gen.go:1` - generated; DO NOT EDIT
> - `internal/modules/documents/approval/http/contracts/strictjson.go:23` - `Decode` helper; `DisallowUnknownFields` pattern used at handler boundaries
> - `internal/modules/controlleddocuments/delivery/http/handler.go:95` - `HandlerWithOptions` wiring pattern (controlled-documents)
> - `internal/modules/templates/delivery/http/handler.go:32` - `ServerInterfaceWrapper` wiring pattern (templates)
> - `migrations/0183_documents_name_not_empty.sql:27` - DB invariant floor for `documents.name`
> - `.github/workflows/api-contract.yml:1` - CI drift guard (3 jobs)
> - `frontend/apps/web/package.json:13` - `gen:api` script (`openapi-typescript`)

---

## 1. Spec location

`api/openapi/v1/openapi.yaml` is the **single source of truth** for all MetalDocs HTTP contracts. OpenAPI 3.0.3. Current public routes use the `/api/v1` prefix.

New endpoints MUST be authored in the spec first. Handlers implement; spec governs.

---

## 2. Backend codegen - oapi-codegen v2

Each migrated module has an `internal/modules/<x>/api/` directory with three files:

| File | Purpose |
|------|---------|
| `cfg.yaml` | Codegen config: package name, output file, `include-tags` filter |
| `gen.go` | Single-line `//go:generate` comment - no production code |
| `api.gen.go` | Generated output - **never hand-edit** |

The generated file provides:
- Go request/response types for all operations in scope.
- `ServerInterface` - one method per operation; handler struct must implement.
- `ServerInterfaceWrapper` - stdlib `net/http` adapter that parses path/query params and calls `ServerInterface` methods.
- `StrictServerInterface` (line ~1608 in controlled-documents gen) - higher-level variant where input/output are typed structs; handler returns `(ResponseObject, error)` instead of writing to `http.ResponseWriter`.

**Regenerate:**

```bash
GOFLAGS=-mod=mod go generate ./internal/modules/controlleddocuments/api/...
GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/...
GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...
```

Use `GOFLAGS=-mod=mod` when the project vendor directory is present; otherwise `go generate` will refuse to fetch the `oapi-codegen` binary dependency.

CI runs `go generate ./...` - see `api-contract.yml:27`.

---

## 3. Handler wiring pattern

Handlers do **not** implement `StrictServerInterface` directly. The current pattern uses `HandlerWithOptions` (previously `ServerInterfaceWrapper`):

```go
// internal/modules/controlleddocuments/delivery/http/handler.go:95
controlleddocumentsapi.HandlerWithOptions(h, controlleddocumentsapi.StdHTTPServerOptions{
    BaseRouter: mux,
    BaseURL: "/api/v1",
    // ...
})
```

The handler struct (`*Handler`) implements `ServerInterface`; the generated `HandlerWithOptions` handles route dispatch and param parsing.

---

## 4. Runtime enforcement gaps

oapi-codegen does **not** enforce:

- **Unknown fields:** Use `contracts.Decode` from `internal/modules/documents/approval/http/contracts/strictjson.go:23`. It sets `decoder.DisallowUnknownFields()`. Call it instead of `json.NewDecoder(r.Body).Decode(...)` at handler boundaries.
- **Required fields:** oapi-codegen generates pointer fields for optional and value fields for required, but does not produce 400 responses for missing required fields at runtime. Handlers must check explicitly (e.g., `missingAtomicCreateField` at `internal/modules/controlleddocuments/delivery/http/routes.go:102`).

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

## 5b. Response-body typing gate (H-D — honest two-part)

**Rule:** no public delivery route may emit a `map[string]<T>` **response literal** — for **any** value
type `T` (`map[string]any`, `map[string]string`, `map[string]int`, …). Every 200/201 body
is a typed struct — a generated model (oapi-codegen modules) or a hand-rolled typed struct (pre-codegen
modules per ADR 0012, e.g. auth/search, and deliberately off-spec routes).

**Type scope widened (F9.4 — post-M8 5th miss).** The M8 gate named `map[string]any` only; the post-M8
re-audit found three documents-handler response literals (`duplicateDocument`, comment list/create/update,
`signedRevisionURL`) that evaded it by emitting `map[string]string`. The class the gate enforces is "no
untyped map response body", not one value type — the rule and the `noresponsemap` analyzer now flag any
`map[string]<T>` reaching a 2xx body writer. The non-response allowlist below is unchanged (those are
non-response uses regardless of value type).

**Why two parts:** the historical one-liner `grep -rEn 'writeJSON.*map\[string\]any'` ("Grep A") is
**necessary but not sufficient** — it is blind to:
- the `writeFillInJSON` / `WriteJSON` (capital) writer aliases, and
- built-then-written locals (`page := map[string]any{...}` / `payload := map[string]any{...}` on one or
  many lines, written on a later line).

In M6 Grep A read 0 while **10** response-literal sites survived behind exactly these patterns. The
honest gate therefore measures in two parts:

**Scope — the FULL public-route surface, not just `delivery/http/`.** The M8 re-audit (4th miss)
proved the path-scoped grep was blind to public routes registered OUTSIDE `internal/modules/*/delivery/http/`
— presence (`internal/modules/iam/presence/`), metrics/health (`internal/platform/observability/`), and
the approval HTTP package (`internal/modules/documents/approval/http/`). The gate now covers every package
that registers a public route:

```bash
ROUTE_PATHS='internal/modules/*/delivery/http/ internal/modules/documents/approval/http/ internal/modules/iam/presence/ internal/platform/observability/'

# Part A — necessary (the one-liner). Must be 0. Any map[string]<T>, not just any.
grep -rEn 'write(JSON|FillInJSON)|WriteJSON' $ROUTE_PATHS --include='*.go' | grep -v _test.go | grep -E 'map\[string\]'

# Part B — completeness (closes the blindspot). Every surviving hit must be on the
# NON-RESPONSE allowlist below; zero response literals.
grep -rEn 'map\[string\]' $ROUTE_PATHS --include='*.go' | grep -v _test.go
```

**Mechanical guard (F8.6 — laundering-resistant, runs in CI).** `tools/cilint` ships the `noresponsemap`
analyzer: it flags a `map[string]<T>` composite literal (any value type — F9.4) reaching a 2xx body writer
(`writeJSON` / `writeFillInJSON` / `WriteJSON`) on any registered-route package — **including** built-then-written locals
(`page := map[string]any{...}; writeJSON(w, 200, page)`) that Grep A is blind to. Run `go run ./tools/cilint
./internal/...`; CI runs it via `.github/workflows/invariants.yml`. Suppress a deliberately off-spec route
with `//cilint:allow-responsemap <reason>` on the writer-call line.

A **response literal** = a `map[string]<T>` (any value type) passed (directly, or via a built local such
as `page :=` / `payload :=`, on one line or many) to `writeJSON` / `writeFillInJSON` / `WriteJSON` (or any
2xx body writer). These are forbidden — convert to a typed body.

**Non-response allowlist** (Part B survivors that are NOT response literals — keep):
- **Domain-mirror struct fields:** `audit AuditEventItem.Payload` (fed by a JSON decode buffer for
  arbitrary stored payload), `security signalItem.Evidence`.
- **Internal audit-emit params:** `recordAudit(... payload map[string]any)` in auth / iam / audit.
- **Command inputs:** `controlleddocuments formData`, `documents ContentFormData`.
- **Declared-dynamic metrics envelope (F8.2):** `observability MetricsResponse.{Runtime,Scheduler,DBPool}`
  — typed envelope whose inner metric blobs are intentionally `map[string]any` (free-form runtime stats,
  not a fixed schema). The envelope itself is a typed struct; only the dynamic leaves are maps.
- **Health/readiness probes (F8.6 — recorded exemption):** `internal/platform/observability/health.go`
  (`/api/v1/health/live`, `/api/v1/health/ready`). These are infra probes (k8s / load balancers), **not**
  the typed FE resource API — no generated client consumes them — and the readiness body is genuinely
  dynamic (a variable dependency-check array). Same category as the declared-dynamic metrics. The
  `noresponsemap` analyzer encodes this exemption by file (`noResponseMapExemptFiles`); it is recorded
  here, **not** silently passed. Any NEW health field stays inside this file or the exemption no longer applies.

> Anti-evasion: do not launder a response literal past Part A (e.g. `writeJSON(any(map[string]any{}))`),
> swap the value type to dodge an `any`-only check (`map[string]string` — the post-M8 evasion, now closed),
> or hide it in a helper — Part B will still flag any `map[string]<T>`, and any survivor that is not on
> the allowlist is a gate failure. Adding a response-shaped exception to the allowlist is forbidden; the
> allowlist is for non-response uses only.

---

## 6. Frontend codegen - openapi-typescript v7

```bash
# frontend/apps/web
pnpm gen:api
# equivalent: openapi-typescript ../../../api/openapi/v1/openapi.yaml -o src/lib/api-types/index.d.ts
```

Output: `frontend/apps/web/src/lib/api-types/index.d.ts` - **never hand-edit**. The `api` client in `lib/api/client.ts` is typed against these generated `paths`.
Local rule: use `pnpm gen:api` in `frontend/apps/web`.
CI rule: `frontend-codegen-drift` runs `npm run gen:api` in the same workspace and must produce equivalent output.

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

## 8. Historical route migration notes (superseded)

Historical snapshot only (captured 2026-05-15). This section is non-governing and must not be used as current contract truth for planning or implementation.

| Snapshot topic | Historical note |
|--------|------------------|
| controlled-documents | Was documented as fully wrapper-mounted with generated routes |
| templates | Was documented as generated core routes mounted with additional follow-up gaps |
| documents | Was documented as bootstrap-generated with deferred handler migration follow-ups |
| other modules | Historical migration posture changed over time; use runtime + spec + generated artifacts + route truth tables for current truth |

**Documents module note:** codegen bootstrap landed (commit `81e7ec23`) - `api.gen.go` is generated and up to date. Handler migration is blocked by spec-handler drift (missing spec ops for `renameDocument`, `duplicateDocument`, comments CRUD; orphaned spec ops with no handler). Details and migration template: `wiki/backlog/contract-first-followups.md`.

---

## 9. Freeze-law contract truth

Treat current truth as a four-way alignment requirement for touched public endpoints:

- runtime route registration and owner
- OpenAPI canonical namespace/path/operation/tag ownership
- generated backend package/interface + generated boundary mount (`ServerInterfaceWrapper`/`HandlerWithOptions`) for generated public modules
- generated frontend wrappers/types and module wiki status used for planning

If any of these conflict, stop and resolve the prerequisite before feature implementation.

### Route truth tables

- [Documents route table](../modules/documents.md#api-route-truth-table-plan-8-baseline)
- [Approval route table](../modules/approval.md#api-route-truth-table-plan-8-baseline)
- [templates route table](../modules/templates.md#api-route-truth-table-plan-8-baseline)
- [Controlled-documents route table](../modules/controlled-documents.md#api-route-truth-table-plan-8-baseline)
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
6. Implement `ServerInterface` on the handler struct; wire via `HandlerWithOptions` (controlled-documents pattern at `handler.go:95`).
7. Commit `api.gen.go` - CI drift check will verify it stays in sync.

---

## See also

- `wiki/architecture/backend-api-structure.md` - canonical backend/API structure rules and migration discipline
- `wiki/decisions/0012-contract-first-api.md` - ADR: why spec-as-source-of-truth was adopted and root cause of the `documents.name` bug
- `wiki/backlog/contract-first-followups.md` - deferred handler migrations + documents spec/handler gap inventory
- `wiki/references/oapi-codegen.md` - operational how-to (regenerate, vendor mode, add module)
- `wiki/architecture/frontend-structure.md section 7` - frontend API call patterns using generated types


