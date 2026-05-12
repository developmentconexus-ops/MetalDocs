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
- module wiki and tech-debt notes

Classify mismatches as:

- product/API ownership issue
- spec drift
- runtime handler drift
- generated-code drift
- legacy debt

### 4. Pick the canonical pattern before implementation

Use the canonical target from `wiki/architecture/backend-api-structure.md`.

Rules:

- one public contract per module
- OpenAPI first
- one generated package per module tag
- generated wrapper routing for migrated modules
- no accidental new raw routes in fully migrated modules

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
npx tsc --noEmit
```

After each `go generate`, inspect generated `ServerInterface` method names and signatures before writing delegation or wrapper code.

## Stop rules

Stop and report when:

- a generated method name or signature differs from the plan
- a route has no corresponding real handler
- a spec operation has no real route owner
- path parameter names differ between spec and runtime
- a module appears to expose competing public ownership for the same behavior

## Output expectations

When finishing backend/API work, report:

1. The module contract decision
2. The route truth mismatches found
3. The code/spec/wiki changes made
4. Verification status
5. Wiki docs updated
