# Backend Platform Freeze Design

> **Status:** approved for planning
> **Date:** 2026-05-21
> **Scope:** MetalDocs backend API platform governance, hard legacy removal, controlled-documents rename, contract runtime alignment, frontend API contract cleanup, and wiki/skill workflow freeze.

## 1. Goal

Make the MetalDocs backend API a product-grade, explicit, frozen platform by removing legacy compatibility surfaces, eliminating ambiguous ownership, renaming `registry` to controlled documents, and enforcing a single contract path from OpenAPI through generated backend code, runtime routes, generated frontend types, and wiki module memory.

This is not feature work. It is a hard platform normalization program that makes future feature work safe.

## 2. Current Problem

MetalDocs accumulated transitional architecture while implementing version by version. Several transitional states became ambient defaults:

- Public runtime routes exist as raw `mux.HandleFunc` registrations even where generated backend packages already exist.
- OpenAPI contains both canonical `/api/v1/...` paths and legacy unprefixed `/documents...` paths.
- The frontend has generated types for many routes but still uses hand-written clients and response types for active product surfaces.
- Wiki pages, tech-debt registers, and workflow checks disagree with runtime/spec/generated truth.
- The historical `registry` module name hides the actual product concept: controlled documents.
- Permission guards and scripts still reference surfaces that are not real canonical product APIs.

The result is a system where runtime truth, contract truth, wiki truth, and frontend truth can all be individually plausible but collectively contradictory.

## 3. Platform Law

Every public HTTP surface must have exactly:

- One product/module owner.
- One OpenAPI path.
- One stable `operationId`.
- One canonical OpenAPI tag.
- One tag-scoped generated backend package.
- One runtime mount through the generated server boundary.
- One generated frontend type source.
- One wiki route-truth row.

Any public route that cannot satisfy this rule is either:

- Promoted into the canonical contract.
- Made explicitly internal and excluded from public API governance.
- Deleted.

Surface presence is not enough. A route is only contract-governed when runtime mounting, OpenAPI, generated backend code, generated frontend types, frontend wrappers, and wiki status all agree.

## 4. Canonical Module Ownership

| Module | Canonical API | Owns | Must not own |
|---|---|---|---|
| Controlled Documents | `/api/v1/controlled-documents*` | Controlled-document identity, code numbering, visibility grants, controlled-document lifecycle, atomic controlled-document plus first revision creation | Editor state, document content, approval decisions |
| Documents | `/api/v1/documents*` | Document instances, content, revisions, sessions, autosave, checkpoints, comments, fill-in, exports, view/reconstruct | Controlled-document numbering/identity, approval quorum/signoff decisions |
| Approval | `/api/v1/approval*` plus approved document-action routes | Approval routes, approval instances, inbox, signoff, publish/schedule/supersede/obsolete workflow | Document editing/content persistence, controlled-document identity |

## 5. Controlled Documents Rename

`registry` is retired as a product/module name. The canonical module is Controlled Documents.

| Layer | Target name |
|---|---|
| Product concept | Controlled Documents |
| Public API | `/api/v1/controlled-documents*` |
| OpenAPI tag | `controlled-documents` |
| Backend module directory | `internal/modules/controlleddocuments` |
| Generated Go package | `controlleddocumentsapi` |
| HTTP package alias | `controlleddocumentshttp` |
| Frontend feature | `frontend/apps/web/src/features/controlled-documents` |
| Query keys | `QK.controlledDocuments.*` |
| Database tables | keep `controlled_documents`, `cd_sequence_counters` |

The term `registry` is forbidden for controlled-document behavior after the freeze, except in migration notes, changelogs, and historical references that explain the rename.

## 6. Document-Scoped Approval Routes

Document-scoped approval routes are promoted as permanent product routes, not compatibility shims.

These routes remain because they are product-friendly and document-centric:

- `POST /api/v1/documents/{id}/submit`
- `POST /api/v1/documents/{id}/signoff`
- `POST /api/v1/documents/{id}/cancel`
- `POST /api/v1/documents/{id}/publish`
- `POST /api/v1/documents/{id}/schedule-publish`
- `POST /api/v1/documents/{id}/supersede`
- `POST /api/v1/documents/{id}/obsolete`
- `GET /api/v1/documents/{id}/approval-instance`

Ownership rule: these routes are approval-owned document actions. They must be tagged `controlled-documents` only if they mutate controlled-document identity. They must be tagged `documents` only if they mutate document content/editor state. Otherwise they are tagged `approval`, generated through `approvalapi`, mounted by the approval generated boundary, and documented in approval route truth tables.

## 7. Runtime Architecture Target

Each public backend module follows this shape:

```text
api/openapi/v1/openapi.yaml
  -> canonical tag and operationId
  -> internal/modules/<module>/api/cfg.yaml
  -> internal/modules/<module>/api/gen.go
  -> internal/modules/<module>/api/api.gen.go
  -> handler implements generated ServerInterface
  -> RegisterRoutes mounts generated ServerInterfaceWrapper only
  -> frontend types generated from the same OpenAPI contract
```

Raw public `mux.HandleFunc` route registration is not allowed in a generated public module after the freeze.

Internal sub-handlers may still exist, but they are implementation details behind the generated boundary.

## 8. Immediate Retirement List

The freeze retires these surfaces or names:

- `registry` as a module/product name.
- OpenAPI tag `registry`.
- Backend directory `internal/modules/registry`.
- Frontend feature directory `frontend/apps/web/src/features/registry`, unless a different future product concept named Registry is intentionally created.
- Unprefixed OpenAPI `/documents...` paths that are not actual runtime public APIs.
- Permission guard for public `POST /api/v1/documents` create.
- Wiki claims that documents or approval have no spec coverage when generated packages exist.
- Workflow checks that treat file/path presence as sufficient contract health.
- Hand-written frontend API request/response types where generated OpenAPI types exist.

Historical migrations are not rewritten. Historical references stay only as evidence.

## 9. Promote Or Delete Decisions

Some surfaces require explicit classification during execution:

| Surface | Decision required |
|---|---|
| `POST /api/v1/documents/{id}/pdf-complete` | Either make internal-only and remove from public route governance, or formalize as a platform/render callback contract. |
| `/registry` frontend route | Rename to `/controlled-documents`, or keep only as a short-lived redirect if browser history compatibility is required. |
| Approval hand-written frontend types | Replace with generated schema/operation types where OpenAPI coverage exists. |
| Controlled-documents local `useEffect` fetching | Move active views to TanStack Query hooks and generated types. |

## 10. Skill, Workflow, And Wiki Freeze

The hard platform freeze must update MetalDocs operating memory so future agents cannot reintroduce the same drift.

### Backend API Skill

`metaldocs-backend-api` must require:

- Route ownership proof before backend API changes.
- Runtime route table from actual registrations.
- OpenAPI path and `operationId` check.
- Generated backend package check.
- Runtime generated-boundary mount check.
- Generated frontend type check when frontend consumers exist.
- Wiki route-truth and module ownership sync.

The skill must fail the workflow on generated-but-raw-mounted public modules.

### Runtime Contract Prerequisite Skill

`runtime-contract-prereq` must classify and stop on:

- Duplicate OpenAPI namespaces.
- Runtime/spec path mismatch.
- Generated backend package present but wrapper not mounted.
- Frontend hand-written type drift for generated routes.
- Permission guard for a non-existent public route.
- Wiki status contradiction that affects implementation planning.

### Wiki Rules

Wiki module docs become contractual implementation memory. They must state:

- Canonical module name and owner.
- Public API surface.
- Runtime mount style.
- Generated package path.
- Frontend wrapper/type status.
- Persistence ownership.
- Legacy/retired surfaces.

No `Last verified` stamp may be bumped without checking runtime, OpenAPI, generated backend, generated frontend, and frontend wrapper truth for affected routes.

### Workflow Checks

The successor to `scripts/check-module-contract-sync.ps1` must fail on:

- Duplicate canonical path families.
- Generated packages without wrapper-mounted runtime routes.
- OpenAPI paths without runtime owners.
- Runtime public routes without OpenAPI operations.
- Stale module names such as `registry` for controlled-document code.
- Frontend hand-written request/response types for generated public routes.
- Wiki module status contradictions for the target module.

## 11. Phased Execution

### Phase 0: Freeze Charter

Write and approve the platform law, ownership map, rename decision, and forbidden states.

Exit gate:

- The design is approved.
- Controlled Documents, Documents, and Approval ownership is explicit.
- `registry` is formally deprecated as a module name.
- Raw public routes in generated modules are classified as migration debt.

### Phase 1: Contract Namespace Cleanup

Normalize OpenAPI before moving files.

Exit gate:

- No duplicate `/documents` versus `/api/v1/documents` namespace.
- Tags are canonical: `controlled-documents`, `documents`, `approval`.
- Frontend generated types no longer expose retired legacy path families.

### Phase 2: Rename Registry To Controlled Documents

Rename backend module, generated package, frontend feature, wiki pages, backlog/debt docs, and query naming.

Exit gate:

- Runtime code no longer imports `internal/modules/registry`.
- Public API remains `/api/v1/controlled-documents`.
- DB table names remain unchanged.
- No product/module references to `registry` remain outside historical notes.

### Phase 3: Runtime Generated Boundary

Promote documents and approval from generated-but-raw-mounted to generated-and-wrapper-mounted.

Exit gate:

- Documents public routes mount through `documentsapi.ServerInterfaceWrapper`.
- Approval public routes mount through `approvalapi.ServerInterfaceWrapper`.
- No public raw `mux.HandleFunc` remains in generated modules.
- Internal-only callbacks are classified and excluded or formally contracted.

### Phase 4: Frontend Contract Cleanup

Move active API wrappers to generated types and TanStack Query where appropriate.

Exit gate:

- Public API wrappers import generated `components` and `operations`.
- Hand-written mirrors of generated request/response shapes are removed.
- Query keys use product names: `controlledDocuments`, `documents`, `approval`.

### Phase 5: Wiki And Skill Freeze

Update skills, AGENTS guidance, backend API workflow, runtime prerequisite workflow, module docs, debt/backlog docs, DB dictionary cross-links, and contract checks.

Exit gate:

- Skills enforce the new backend platform law.
- Wiki module truth matches runtime/spec/generated/frontend truth.
- Contract checks fail on the drift classes that caused this freeze.

### Phase 6: Final Platform Gate

Run the full verification suite and mark the API platform frozen.

Exit gate:

- OpenAPI lint passes.
- `go generate` is clean.
- Controlled-documents, documents, approval, and permissions tests pass.
- Frontend API generation and typecheck pass.
- Wiki/module-contract sync passes.
- No forbidden `registry` product/module references remain outside historical notes.

## 12. Non-Goals

- Do not rewrite historical migrations.
- Do not redesign product behavior during the freeze.
- Do not introduce compatibility wrappers to preserve legacy internal names.
- Do not change database table names that are already domain-aligned.
- Do not merge frontend visual redesign work into the backend platform freeze.

## 13. Success Criteria

The freeze is successful when a new engineer or agent can answer these questions without archaeology:

- Which module owns this API route?
- Which OpenAPI operation defines it?
- Which generated backend package contains it?
- Where is it mounted at runtime?
- Which frontend generated type represents it?
- Which module wiki page documents it?
- Is it active, internal, or retired?

If any answer requires remembering old `registry` terminology or checking multiple contradictory sources, the freeze is not complete.
