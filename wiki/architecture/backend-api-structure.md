# Backend/API Structure

> **Last verified:** 2026-05-12
> **Scope:** Canonical backend/API structure rules for MetalDocs: module HTTP ownership, OpenAPI-first workflow, generated package layout, route truth requirements, migration rules, and verification gates.
> **Out of scope:** API behavior conventions such as RFC 9457, pagination, idempotency, and authz details (`architecture/api-design-system.md`); operational oapi-codegen usage (`references/oapi-codegen.md`); ADR rationale (`decisions/0012-contract-first-api.md`).
> **Key files:**
> - `api/openapi/v1/openapi.yaml:1` - single public HTTP contract
> - `internal/modules/controlleddocuments/api/cfg.yaml:1` - canonical per-module codegen config
> - `internal/modules/controlleddocuments/api/gen.go:1` - canonical `//go:generate` file
> - `internal/modules/controlleddocuments/delivery/http/handler.go:67` - canonical `ServerInterfaceWrapper` pattern
> - `wiki/architecture/api-contract.md:1` - contract-first architecture overview
> - `wiki/architecture/api-design-system.md:1` - shared API behavior conventions

---

## 1. Why this document exists

MetalDocs is standardizing on contract-first HTTP modules. Without one structural rulebook, each module can drift into a different route shape, different handler ownership model, and different migration style. This document defines the structural contract all backend/API work must follow so agents and humans make the same architectural choices before changing code.

## 2. The canonical module pattern

Every business module that exposes public HTTP endpoints should converge to this shape:

```text
OpenAPI path + operationId
  -> tag-scoped generated package at internal/modules/<module>/api/
  -> module HTTP handler implements generated ServerInterface
  -> RegisterRoutes wires only ServerInterfaceWrapper
  -> frontend types generate from the same OpenAPI contract
```

Canonical generated package layout:

```text
internal/modules/<module>/api/
  cfg.yaml
  gen.go
  api.gen.go
```

Rules:

- `api/openapi/v1/openapi.yaml` is the only public HTTP contract.
- Each module owns one OpenAPI tag and one generated Go package.
- Generated files are committed and never hand-edited.
- `RegisterRoutes` for a migrated module must mount generated wrapper handlers, not raw `mux.HandleFunc` business routes.

## 3. Route ownership rules

Before changing or migrating routes, identify the module that owns the public contract.

Rules:

- A public route belongs to exactly one module contract.
- If two modules appear to expose overlapping mutations, stop and resolve the product/API ownership first.
- Internal helper handlers may live in multiple Go files, but the public module contract must still be singular and explicit.
- Route ownership must be documented in the module wiki when a module has more than one internal HTTP handler file.

Acceptable internal structures:

1. One HTTP handler struct owns all public routes for the module.
2. A module-level contract handler coordinates stable internal sub-handlers.

The second pattern is acceptable only when it is a deliberate architecture choice. It is not a shortcut for avoiding cleanup.

## 4. OpenAPI-first rules

All public HTTP work starts from the spec.

Rules:

- Add or change the OpenAPI path before implementing a new public route.
- Every public operation must have a stable `operationId`.
- Every operation must use the module's canonical tag.
- Path parameter names must match runtime expectations exactly.
- If a runtime path and spec path differ, stop and resolve the mismatch before codegen wiring.

Examples of mismatches that must be resolved before implementation:

- competing create routes for one resource
- runtime `{version}` versus spec `{versionNum}`
- route exists in code but not in spec
- operation exists in spec but has no real handler

## 5. Route truth table gate

Before migrating or restructuring a module, build a route truth table from actual runtime code.

Minimum columns:

| Method | Runtime path | Owning file | Runtime handler | Spec path | OperationId | Generated method | Notes |
|---|---|---|---|---|---|---|---|

The truth table is required when:

- migrating a raw module to OpenAPI/codegen
- completing partial codegen adoption
- rewriting a plan that touches multiple module routes
- investigating route drift or ownership ambiguity

## 6. Legacy raw route policy

Raw `mux.HandleFunc` registrations are allowed only for modules that are not yet fully migrated.

Rules:

- Legacy raw routes must be treated as migration debt, not as the target architecture.
- Once a module is declared fully migrated, new public routes for that module must not be added as raw `mux.HandleFunc`.
- Mixing generated wrapper routes and new raw routes in a fully migrated module is not allowed.
- Partially migrated modules must record the remaining raw routes in their tech-debt register or migration plan.

## 7. Verification gates

Backend/API structural changes are not complete until all applicable verification passes.

Required checks:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = "-mod=mod"
go generate ./internal/modules/<module>/api/...
go build ./...
go test ./internal/modules/<module>/... -count=1
```

If frontend generated types are affected:

```powershell
cd frontend/apps/web
pnpm gen:api
npx tsc --noEmit
```

Additional gate:

- After each `go generate`, inspect the generated `ServerInterface` method names and signatures before adding delegation or wrapper wiring.

## 8. Required cross-reads

Read these alongside this document:

- `wiki/architecture/api-contract.md`
- `wiki/architecture/api-design-system.md`
- `wiki/references/oapi-codegen.md`
- relevant module doc and tech-debt register under `wiki/modules/`

## 9. See also

- `wiki/decisions/0012-contract-first-api.md`
- `wiki/backlog/roadmap.md`
- `docs/superpowers/specs/2026-05-12-backend-api-governance-design.md`

