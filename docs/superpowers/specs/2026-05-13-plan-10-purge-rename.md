# Plan 10 Spec - Legacy Purge + Rename Sweep

## Scope
- Execute Plan 10 only.
- Mechanical/minimal changes only: rename, move, route prefix sweep, schema cleanup, dead-surface retirement.
- No fallback aliases and no `/api/v2/*` runtime paths.

## Workstream Batching
1. Batch A+B (largest blast radius):
- `internal/modules/templates_v2` -> `internal/modules/templates` rename and import rewrites.
- `/api/v2/*` -> `/api/v1/*` sweep across backend routes, frontend callsites, OpenAPI, and test/wiki runtime references.

2. Batch C (schema and repository alignment):
- `approval_instances.document_v2_id` -> `document_id` migration and code updates.
- constraint/index rename alignment.
- `templates_v2_template_version.editable_zones` drop migration after reference grep is clean.
- `VALIDATE CONSTRAINT` migration for scoped approval FKs.

3. Batch D+E (mechanical cleanup):
- editor-ui dead-surface cleanup and file rename.
- approval signature package path rename (`infra/signature` -> `infrastructure/signature`).
- docs tail cleanup and backlog state sync.

## Migration Order
1. `0194_approval_document_id_rename.sql`
- Renames `approval_instances.document_v2_id` to `document_id`.
- Renames related constraints/index names to `...document_id...`.
- Must land before runtime paths that depend on new column names.

2. `0195_approval_validate_iam_fks.sql`
- Validates previously `NOT VALID` approval FK constraints.
- Safe after rename alignment and before any drop operations.

3. `0196_drop_templates_editable_zones.sql`
- Drops `templates_v2_template_version.editable_zones`.
- Must run only after grep confirms no runtime/sql references.

## Verification Gates
- `go generate ./...`
- `go test ./...`
- `go build ./apps/api/cmd/metaldocs-api ./apps/worker/cmd/metaldocs-worker`
- `pnpm --dir frontend/apps/web gen:api`
- `pnpm --dir frontend/apps/web build`
- Grep gates:
  - no `metaldocs/internal/modules/templates_v2` imports in `internal/`, `apps/`, `tests/`
  - no `/api/v2/` in `internal/`, `frontend/`, `api/openapi/`, `wiki/tests/`
  - no `document_v2_id` in runtime/test SQL paths

## Current Verification Status
- Passed: backend generate/test/build, frontend generate/build, grep gates.
- Known non-blocking baseline issue: `pnpm --dir frontend/apps/web test` has pre-existing unrelated failures.
