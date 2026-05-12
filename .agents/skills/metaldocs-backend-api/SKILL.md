---
name: metaldocs-backend-api
description: Use when working on MetalDocs backend/API structure, OpenAPI, oapi-codegen, public HTTP routes, handler wiring, or route migrations. Triggers on changes under `api/openapi/v1/openapi.yaml`, `internal/modules/*/api/`, `internal/modules/*/delivery/http/`, `internal/modules/documents/approval/http/`, and on backend API questions involving route truth, RFC 9457, idempotency, authz, pagination, or generated frontend types.
---

# MetalDocs Backend/API

Read these docs first:

1. `wiki/README.md`
2. `wiki/architecture/backend-api-structure.md`
3. `wiki/architecture/api-contract.md`
4. `wiki/architecture/api-design-system.md`
5. `wiki/references/oapi-codegen.md`
6. The affected module doc and tech-debt register

## Required workflow

1. Build a route truth table from runtime code before changing routes.
2. Compare runtime routes to OpenAPI paths, `operationId`, generated `ServerInterface`, and module docs.
3. Classify mismatches before coding.
4. Use one canonical public contract per module.
5. For migrated modules, use generated wrapper routing instead of new raw route registrations.
6. After each `go generate`, inspect generated method names and signatures before wiring handlers.
7. Verify with OpenAPI lint, `GOFLAGS=-mod=mod go generate`, `go build`, targeted tests, and frontend API type generation when relevant.

## Stop rules

Stop and report when:

- generated signatures differ from the plan
- runtime and spec path parameter names differ
- a route has no real handler owner
- a spec route has no real runtime owner
- a module has ambiguous public route ownership
