---
name: metaldocs-backend-api
description: Use this skill for ANY MetalDocs backend or API work that touches public HTTP routes, OpenAPI, oapi-codegen, handler wiring, API contracts, generated api packages, route migrations, or module HTTP structure. Triggers on changes under `api/openapi/v1/openapi.yaml`, `internal/modules/*/api/`, `internal/modules/*/delivery/http/`, `internal/modules/documents/approval/http/`, and on requests about RFC 9457, idempotency, pagination, authz, tenant context, or frontend API type generation.
---

# MetalDocs Backend/API

You are working on the canonical backend/API surface of MetalDocs. Use this skill before changing any public HTTP contract or module route wiring.

## Read first

Read these docs in order:

1. `wiki/README.md`
2. `wiki/architecture/backend-api-structure.md`
3. `wiki/architecture/api-contract.md`
4. `wiki/architecture/api-design-system.md`
5. `wiki/references/oapi-codegen.md`
6. The affected module doc and tech-debt register under `wiki/modules/`

If frontend generated API types, feature API wrappers, TanStack Query hooks, or cache invalidation are affected, also use `.agents/skills/metaldocs-tanstack-query/SKILL.md`.

## Workflow

### 1. Orient the module

- Identify the affected module or modules.
- Read the module wiki doc and tech-debt register.
- Check whether the module is raw, partially migrated, or fully migrated.

### 2. Build route truth before editing

Build a route truth table from actual runtime registrations before changing routes.

Minimum columns:

| Method | Runtime path | Owning file | Runtime handler | Spec path | OperationId | Generated method | Notes |
|---|---|---|---|---|---|---|---|

Use runtime code as the source of truth for what currently exists.

### 3. Compare all contract surfaces

Compare:

- runtime route registrations
- OpenAPI path and tag ownership
- `operationId`
- generated `ServerInterface`
- generated frontend `paths` types when relevant
- module wiki and tech-debt notes

Classify mismatches as:

- product/API ownership issue
- spec drift
- runtime handler drift
- generated-code drift
- frontend contract drift
- legacy debt

### 4. Pick the canonical pattern before implementation

Use the canonical target from `wiki/architecture/backend-api-structure.md`.

Rules:

- one public contract per module
- OpenAPI first
- one generated package per module tag
- generated wrapper routing for migrated modules
- freeze check: generated module package exists for each touched generated public module (example: controlled-documents at `internal/modules/controlleddocuments/api/` with `package controlleddocumentsapi`)
- freeze check: runtime is mounted via generated wrapper/`HandlerWithOptions` from the generated package for each touched generated public module; no raw public mux route ownership for canonical generated endpoints (example: `/api/v1/controlled-documents*`)
- freeze check: runtime routes, OpenAPI operations, generated server interface, and frontend generated wrappers/types stay aligned for touched endpoints
- no accidental new raw routes in fully migrated modules
- no frontend cache/query work from hand-written response shapes when OpenAPI should own the type

If the runtime shape and target pattern conflict, stop and update the design or plan before coding.

### 5. Verification gates

Run the applicable checks:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = "-mod=mod"
go generate ./internal/modules/<module>/api/...
go build ./...
go test ./internal/modules/<module>/... -count=1
```

If frontend API types are affected:

```powershell
cd frontend/apps/web
pnpm gen:api
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

After each `go generate`, inspect generated `ServerInterface` method names and signatures before writing delegation or wrapper code.

After `pnpm gen:api`, inspect affected feature API wrappers and query hooks. If generated path/type changes require cache key or invalidation changes, switch to `metaldocs-tanstack-query`.

## Stop rules

Stop and report when:

- a generated method name or signature differs from the plan
- a route has no corresponding real handler
- a spec operation has no real route owner
- path parameter names differ between spec and runtime
- a module appears to expose competing public ownership for the same behavior
- frontend needs a hand-written type for a route that should be covered by OpenAPI
- a backend route change changes frontend cache semantics without a TanStack invalidation plan
- generated package/module identity contradicts canonical module ownership (for example legacy `registry` naming presented as canonical without explicit historical/migration label)
- hard fail: a public module has generated backend package/artifacts but public runtime ownership remains raw-mounted instead of generated wrapper/`HandlerWithOptions`
- runtime mounting pattern conflicts with freeze requirements (generated wrapper/`HandlerWithOptions` vs raw mux ownership for touched generated public routes)
- runtime/spec/generated/frontend-wrapper alignment cannot be proven for touched endpoints

## Output expectations

When finishing backend/API work, report:

1. The module contract decision
2. The route truth mismatches found
3. The code/spec/wiki changes made
4. Frontend type/query impact
5. Verification status
6. Wiki docs updated
