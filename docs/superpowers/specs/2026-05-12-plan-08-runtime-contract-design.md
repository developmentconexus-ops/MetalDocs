# Plan 8 Runtime Contract Completion Design

> **Status:** approved design direction, pending implementation plan
> **Date:** 2026-05-12
> **Owner:** leandro
> **Related superseded draft:** `docs/superpowers/specs/2026-05-12-plan-08-openapi.md`

## Goal

Plan 8 completes the OpenAPI contract for every currently mounted public runtime route in MetalDocs, then wires code generation where the runtime owner is structurally ready.

The target is a professional contract-first API baseline: the public contract describes the real server, generated interfaces do not ask handlers to implement routes they do not own, and frontend types come from one OpenAPI source.

## Source Of Truth

The route truth tables published in module wiki docs are the baseline for implementation:

- `wiki/modules/documents.md`
- `wiki/modules/approval.md`
- `wiki/modules/templates_v2.md`
- `wiki/modules/registry.md`
- `wiki/modules/taxonomy.md`
- `wiki/modules/audit.md`
- `wiki/modules/iam.md`
- `wiki/modules/auth.md`
- `wiki/architecture/api-contract.md`

Implementation must read the exact tech-debt rows and cited artifacts before closing a row. Runtime route registration always wins over stale OpenAPI entries.

## Scope

Plan 8 covers all mounted public runtime routes under:

- `/api/v2/documents/*`
- `/api/v2/approval/*`
- `/api/v2/templates*`
- `/api/v2/controlled-documents*`
- `/api/v2/taxonomy/*`
- `/api/v2/iam/*`
- `/api/v1/auth/*`
- `/api/v1/iam/*`
- `/api/v1/audit/*`

Plan 8 also removes or retags stale spec-only operations that have no mounted runtime owner and currently pollute generated module interfaces.

## Out Of Scope

- `/api/v2` to `/api/v1` rename sweep. That belongs to Plan 10.
- `_v2` package/module rename. That belongs to Plan 10.
- New business behavior, idempotency expansion, transaction redesign, or workflow redesign. Those belong to Plan 9 or later.
- Inventing endpoints to satisfy a desired API shape.
- Rewriting handlers to make generated code easier. Existing behavior stays the baseline.

## Design Principles

1. Runtime truth first.
   Every spec operation must map to a real mounted runtime route and a real handler owner.

2. One route, one public owner.
   A route may call internal services across modules, but its public HTTP contract belongs to exactly one handler owner.

3. Tags must match codegen ownership.
   `include-tags` should not pull unrelated or legacy paths into a module interface.

4. Wrapper wiring follows readiness.
   Fully aligned owners should mount through `ServerInterfaceWrapper`. Multi-owner modules may first become `Contracted only` where wrapper wiring would force an artificial handler boundary.

5. Plan 10 handles version cleanup.
   Current `/api/v2` runtime paths remain canonical for Plan 8.

## Route Classification

Every route receives one of these outcomes:

| Outcome | Meaning |
|---|---|
| `Contracted + codegen-wired` | OpenAPI path exists, operationId exists, generated method matches runtime owner, route mounts through `ServerInterfaceWrapper`. |
| `Contracted only` | OpenAPI path and operationId exist, but runtime remains raw for this plan because ownership spans a helper handler or split handler surface. |
| `Internal webhook` | Runtime route is not part of browser/client public API and is documented outside public client codegen. |
| `Retired from spec` | Existing spec path has no mounted runtime owner and must not generate module interface methods. |

## Key Corrections From The Superseded Draft

- `POST /api/v2/documents` is not a valid documents-owned public create route. ADR 0011 moved primary creation to `POST /api/v2/controlled-documents`.
- Documents has multiple real HTTP owners: core handler, export handler, fill-in handler, view handler, reconstruct handler, placeholder-options handler, PDF webhook handler, and approval submodule handler.
- Spec-only legacy paths such as `/documents/{documentId}/render/pdf` must not be tagged in a way that makes `internal/modules/documents/api` require `documents/delivery/http.Handler` to implement them.
- The current runtime version prefix is still `/api/v2` for business modules. Plan 8 must not perform the Plan 10 version rename.
- Taxonomy routes are `/api/v2/taxonomy/*` in current runtime, not `/api/v1/taxonomy/*`.

## Module Decisions

### Registry

Registry remains the owner of controlled-document creation, numbering, and revision creation.

Routes are already wrapper-mounted and aligned. Plan 8 should keep them `Contracted + codegen-wired`, confirm registry responses include current RFC 9457 error references, and avoid changing creation ownership.

### Documents

Documents primary creation is not part of this module contract.

Include these public routes in OpenAPI:

- core document routes: list, stats, detail, rename, finalize, archive, duplicate
- editor session routes
- autosave routes
- checkpoint routes
- revision signed URL
- comments CRUD
- export routes
- fill-in schema and placeholder value routes
- view route
- reconstruct route
- placeholder-options route

Treat `POST /api/v2/documents/{id}/pdf-complete` as an internal webhook unless client usage is discovered during implementation.

Implementation must not force every route into `documents/delivery/http.Handler`. A clean implementation may use separate generated tags/packages for sub-owners or mark helper surfaces `Contracted only` until a deliberate module-level contract handler is designed.

### Approval

Approval owns all mounted routes in `internal/modules/documents/approval/http/router.go`.

Plan 8 should author all 16 runtime approval routes under their current paths and add a generated package under `internal/modules/documents/approval/api/`. Wiring may use `ServerInterfaceWrapper` if generated signatures exactly match handler methods; otherwise the route remains `Contracted only` and the mismatch is documented for a follow-up design decision.

### Templates V2

Templates has an existing generated core plus hand-rolled routes.

Plan 8 should add every mounted template route to the spec using current paths and current request/response behavior. It should not rename `templates_v2` or change route prefixes.

### Taxonomy

Taxonomy requires full 16-route spec authoring from actual handler/domain shapes.

Field names must come from the current handlers and domain structs. No new taxonomy API shape is introduced in Plan 8.

### Audit

Audit gets an operationId and codegen package if the generated wrapper can map directly to `GET /api/v1/audit/events`.

### IAM And Auth

Auth and IAM admin routes already have partial spec coverage through `/api/v1` server-prefixed paths. Plan 8 should add missing operationIds and add OpenAPI coverage for mounted IAM area-membership routes.

If raw dispatcher handlers make wrapper wiring risky, these routes may be `Contracted only` for Plan 8, with generated frontend types still produced.

## Implementation Strategy

Use small PR-sized phases:

1. Clean failed exploratory implementation changes and preserve this design.
2. Spec hygiene: remove stale spec-only operations from module codegen scope; fix tags and operationIds.
3. Low-risk modules: registry, audit, auth, IAM.
4. Documents runtime contract by real route owner.
5. Approval full route contract.
6. Templates_v2 remaining runtime routes.
7. Taxonomy full route contract.
8. Frontend type generation, full drift check, wiki and roadmap sync.

Each phase must run `go build ./...` before commit. Any failing build stops the phase.

## Verification Gates

- Runtime route scan: `rg -n "mux\\.(HandleFunc|Handle)\\(" internal/modules apps/api -g"*.go" -g"!*_test.go" -g"!api.gen.go"`
- OpenAPI lint using the repo-supported command.
- Module codegen with `$env:GOFLAGS = "-mod=mod"`.
- Generated `ServerInterface` inspection after each `go generate`.
- Targeted module tests after each module phase.
- `go build ./...` after every phase.
- Frontend OpenAPI type generation after backend specs settle.
- TypeScript compile check after frontend generation.

## Stop Rules

Stop and report when:

- a generated method name differs from the planned delegation method
- a generated signature requires parameters the runtime handler cannot accept without behavior changes
- a spec path has no runtime owner
- a runtime route has no real handler method
- a path parameter name differs between runtime and spec
- wrapper wiring would force a helper route into the wrong handler owner
- `go build ./...` fails after a phase

## Success Criteria

- Every mounted public runtime route is represented in OpenAPI or explicitly classified as internal.
- No stale spec-only operation generates methods into the wrong module interface.
- Registry remains the sole primary controlled-document creation owner.
- Documents helper routes are included without collapsing unrelated handler owners.
- Taxonomy, approval, templates_v2, audit, IAM, auth, registry, and documents have route truth aligned with spec status.
- Frontend generated API types reflect the completed public contract.
