# Plan 10 Legacy Purge + Rename Sweep Implementation Spec

> **Date:** 2026-05-13
> **Mode:** implementation-only after approval
> **Status:** draft for confirmation
> **Scope:** Plan 10 only: mechanical legacy purge, `_v2` rename, `/api/v2` to `/api/v1`, schema cleanup, dead-surface deletion, and low-risk doc notes.
> **Hard rules:** no fallbacks, no guessed defaults, evidence-first artifact reading, no Plan 12/13 feature work, no architecture redesign.

## 1. Baseline Verification

### Roadmap state

`wiki/backlog/roadmap.md` is internally inconsistent. The execution table marks Plan 10 as `pending`, while the Plan 10 body currently says `done 2026-05-13` and reuses the Plan 8 commit list. Repo evidence contradicts the body status:

- `internal/modules/templates_v2/` exists.
- `internal/modules/templates/` does not exist.
- `api/openapi/v1/openapi.yaml` still contains many `/api/v2/*` paths.
- `internal/`, `frontend/`, and `packages/` still contain Plan 10 target strings.
- Plan 10 backlog rows remain open in the module backlogs.

Plan 10 should start with a roadmap drift correction commit that sets Plan 10, Plan 12, and Plan 13 back to `pending`, and aligns Plan 11 based on verification.

### Plans 4-8 prerequisite check

`wiki/backlog/roadmap.md` confirms these prerequisites are complete:

- Plan 4: done 2026-05-11.
- Plan 5: done 2026-05-11.
- Plan 6a: done 2026-05-11.
- Plan 7: done 2026-05-11.
- Plan 8: done 2026-05-13.

### Plan 11 verification note

Repo evidence shows Plan 11 was at least materially executed:

- `docs/superpowers/specs/2026-05-11-plan-11-editor-frontend.md` exists.
- `docs/superpowers/plans/2026-05-13-plan-09r-recovery.md` and recent commits include Plan 11-related docs.
- `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx` imports `MetalDocsEditor` from `@metaldocs/editor-ui`.
- `packages/editor-ui/test/templatePlugin.wiring.test.tsx` exists and exercises `template-draft`, `document-edit`, and `readonly` gating.

However, editor-ui backlog rows R-004/R-005/R-006 remain open and are part of Plan 10. Treat Plan 11 table status as roadmap drift, not as proof that Plan 10 editor cleanup already landed.

## 2. Required Evidence Read

Evidence read for this spec:

- `wiki/backlog/roadmap.md` Plan 10 and prerequisite statuses.
- `wiki/README.md` wiki index and drift policy.
- `CLAUDE.md` local startup, wiki, frontend, and backend/API workflows.
- `wiki/architecture/backend-api-structure.md` route truth table, OpenAPI-first, generated-wrapper policy.
- `wiki/architecture/api-contract.md` spec/codegen/frontend type generation contract.
- `wiki/architecture/api-design-system.md` `/api/v1` API design convention, RFC 9457, idempotency, authz.
- `wiki/architecture/frontend-structure.md` frontend API/type/query conventions.
- `wiki/decisions/0002-zone-purge.md` editable zones removal context.
- `wiki/modules/templates_v2-tech-debt.md` T-012 and `wiki/backlog/templates_v2-refactor.md` R-100/R-101/R-012.
- `wiki/modules/templates_v2/_artifacts/04-persistence.md` section 6 for editable zone residual evidence.
- `wiki/modules/approval-tech-debt.md` T-007/T-008/T-009/T-011 and `wiki/backlog/approval-refactor.md` R-007/R-008/R-009/R-011.
- `wiki/modules/approval/_artifacts/01-surface.md` section 1 for `infra/signature` and helper surface.
- `wiki/modules/approval/_artifacts/04-persistence.md` sections 1, 3, and 6 for `document_v2_id`, NOT VALID FKs, GUC helpers, and tripwire context.
- `wiki/modules/approval/_artifacts/02-flow-signoff.md` section 6 for `setAuthzGUC` runtime path.
- `wiki/modules/registry-tech-debt.md` T-010 and `wiki/backlog/registry-refactor.md` R-010/R-100.
- `wiki/modules/registry/_artifacts/03-deps.md` section 3 for the second repository instance at `apps/api/cmd/metaldocs-api/main.go`.
- `wiki/modules/auth-tech-debt.md` T-012 and `wiki/backlog/auth-refactor.md` R-012.
- `wiki/modules/auth/_artifacts/03-deps.md` section 4 for `OriginProtection` config surface.
- `wiki/backlog/editor-ui-eigenpal-refactor.md` R-004/R-005/R-006.
- `wiki/modules/taxonomy-tech-debt.md` T-013/T-015 and `wiki/backlog/taxonomy-refactor.md` R-013/R-015.
- `wiki/modules/taxonomy/_artifacts/04-persistence.md` section 4 for redundant PK evidence.
- `wiki/modules/documents-tech-debt.md` and `wiki/backlog/documents-refactor.md` R-100.
- `wiki/tests/system-acceptance-test.md` for final acceptance routine and routes needing v1 update.

## 3. Current-State Findings

### Plan 10 not executed

Current grep confirms these anchors are still live:

- `internal/modules/templates_v2/` exists; `internal/modules/templates/` does not.
- `api/openapi/v1/openapi.yaml` has `/api/v2/templates`, `/api/v2/taxonomy`, `/api/v2/controlled-documents`, `/api/v2/documents`, and `/api/v2/approval` paths.
- `internal/modules/documents/approval/repository/postgres_approval_repository.go` still references `document_v2_id`.
- `internal/modules/documents/approval/application/read_service.go` still selects and joins on `document_v2_id`.
- `internal/modules/documents/approval/application/obsolete_service.go` still filters by `document_v2_id`.
- `internal/modules/documents/repository/resolver_readers.go`, `internal/modules/registry/delivery/http/routes.go`, `internal/modules/jobs/stuck_instance_watchdog/job.go`, and `internal/test/e2e_seed.go` still reference `document_v2_id`.
- `packages/editor-ui/src/index.ts` still exports `createOutlinePlugin`.
- `packages/editor-ui/src/plugins/OutlinePlugin.tsx` still exists.
- `packages/editor-ui/src/plugins/mergefieldPlugin.ts` still exists.
- `packages/editor-ui/src/types.ts` still has `onLockLost`.
- `internal/modules/documents/approval/infra/signature/` still exists.
- `apps/api/cmd/metaldocs-api/main.go` still constructs a second registry repository instance for `cdRepo`.

### Already closed or already wired rows

Do not duplicate these:

- Documents R-100 is already done: `wiki/modules/documents-v2.md` is absent.
- Taxonomy T-013/R-013 is already closed: `migrations/0188_tripwire_extend.sql` defines `reject_families_code_update()` and `trg_reject_families_code_update`.
- Auth T-012/R-012 is already wired: `apps/api/cmd/metaldocs-api/main.go` constructs `security.NewOriginProtection(...)` from `authCfg.OriginProtection`, `SessionCookieName`, and `TrustedOrigins`, and wraps the handler with `originProtection.Wrap(...)`. Keep this as verification/backlog closure only unless tests reveal drift.
- Registry `profile_sequence_counters` has a prior drop migration in `migrations/0182_cd_sequence_per_area.sql`, but old create/grant migrations and comments still reference it. Plan 10 must verify no runtime usage before adding any additional migration. Historical migration references are allowed to remain unless the grep gate explicitly targets runtime paths.

## 4. Workstream Batching

### PR 10.0: Roadmap Drift Correction and Route Truth Prep

Goal: remove planning ambiguity before code movement.

Files:

- Modify: `wiki/backlog/roadmap.md`
- Create: `docs/superpowers/specs/2026-05-13-plan-10-purge-rename.md`

Actions:

1. Correct Plan 10/12/13 roadmap statuses to pending if confirmed by repo evidence.
2. Correct Plan 11 status based on repo verification, without expanding Plan 10 scope.
3. Build route truth tables from runtime code for templates, documents, registry, taxonomy, and approval before changing route strings.
4. Compare each runtime route to `api/openapi/v1/openapi.yaml`, generated `api.gen.go`, frontend generated types, and module docs.

Verification:

```powershell
rg -n "Plan 10|Plan 11|Plan 12|Plan 13" wiki/backlog/roadmap.md
rg -n "/api/v2/" internal api/openapi frontend wiki/tests
```

Expected:

- Roadmap status reflects actual implementation state.
- `/api/v2/` still exists at this stage; this PR only documents truth before the sweep.

### PR 10.1: Templates Module Rename and Template URL Canonicalization

Goal: move `internal/modules/templates_v2/` to `internal/modules/templates/`, update Go imports, update package aliases, and move template public routes to `/api/v1/templates/*`.

Files likely touched:

- Move: `internal/modules/templates_v2/` -> `internal/modules/templates/`
- Modify: `apps/api/cmd/metaldocs-api/main.go`
- Modify: `internal/platform/docgenv2/*` template readers if they import the old module path.
- Modify: `api/openapi/v1/openapi.yaml`
- Modify: `api/openapi/v1/partials/templates*.yaml` if present.
- Regenerate: `internal/modules/templates/api/api.gen.go`
- Modify/regenerate: `frontend/apps/web/src/lib/api-types/index.d.ts`
- Modify: `frontend/apps/web/src/features/templates/**`
- Modify wiki links from `templates_v2` to `templates` only when they describe the renamed module.
- Retire: `wiki/modules/templates-v2.md` after inbound links are repointed.

Stop rules:

- Stop if `internal/modules/templates/` already exists with non-equivalent code.
- Stop if moving creates an import cycle.
- Stop if generated `ServerInterface` signatures differ from the route truth table.
- Stop if any template route has no runtime owner or no spec owner.

Verification:

```powershell
$env:GOFLAGS = "-mod=mod"
go generate ./internal/modules/templates/api/...
go test ./internal/modules/templates/... -count=1
go test ./apps/api/... -count=1
cd frontend/apps/web; pnpm gen:api; pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Grep gate for this PR:

```powershell
rg -n "internal/modules/templates_v2|modules/templates_v2|templates_v2" internal apps api frontend packages wiki --glob "!wiki/modules/templates_v2/_artifacts/**"
rg -n "/api/v2/templates|/api/v2/signed|/api/v2/templates/v2" internal api/openapi frontend wiki/tests packages
```

Expected:

- No active Go import path uses `metaldocs/internal/modules/templates_v2`.
- No active public template URL uses `/api/v2`.
- Historical artifact paths may remain only under `_artifacts` if not linked as current truth.

### PR 10.2: Global `/api/v2` to `/api/v1` Contract Sweep

Goal: canonicalize remaining business module URLs to `/api/v1/*` with no parallel aliases.

Modules:

- documents
- registry / controlled-documents
- taxonomy
- approval and document-scoped approval routes
- IAM area memberships only if runtime/spec still expose `/api/v2/iam/*`; do not expand IAM semantics beyond the URL sweep.

Files likely touched:

- `api/openapi/v1/openapi.yaml`
- `api/openapi/v1/partials/*.yaml`
- `internal/modules/documents/**`
- `internal/modules/documents/approval/**`
- `internal/modules/registry/**`
- `internal/modules/taxonomy/**`
- `internal/modules/iam/**` only for v2 URL cleanup if present.
- `apps/api/cmd/metaldocs-api/permissions.go`
- `frontend/apps/web/src/features/**/api/**`
- `frontend/apps/web/src/features/**/queries/**`
- `frontend/apps/web/src/lib/api-types/index.d.ts`
- `wiki/tests/system-acceptance-test.md`

Stop rules:

- Stop if any runtime route has no matching OpenAPI path.
- Stop if a generated method name/signature changes in a way the handler does not implement.
- Stop if frontend needs to hand-write a type that should come from OpenAPI.

Verification:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = "-mod=mod"
go generate ./internal/modules/.../api/...
go test ./internal/modules/documents/... -count=1
go test ./internal/modules/registry/... -count=1
go test ./internal/modules/taxonomy/... -count=1
go test ./apps/api/... -count=1
cd frontend/apps/web; pnpm gen:api; pnpm.cmd tsc --noEmit -p tsconfig.build.json; pnpm test
```

Grep gate:

```powershell
rg -n "/api/v2/" internal frontend api/openapi wiki/tests
```

Expected:

- No `/api/v2/` remains in the gated paths except explicitly documented historical notes under `wiki/tests` if needed.

### PR 10.3: Approval Schema Column Rename and FK Validation

Goal: rename `approval_instances.document_v2_id` to `document_id`, align indexes/constraints, update code/tests, and validate previously NOT VALID FKs.

Files likely touched:

- Add migration after `0193`: `0194_approval_document_id_rename.sql`
- Add migration after rename: `0195_approval_validate_iam_fks.sql` or combine only if ordering remains clear.
- Modify: `internal/modules/documents/approval/repository/postgres_approval_repository.go`
- Modify: `internal/modules/documents/approval/repository/errors.go`
- Modify: `internal/modules/documents/approval/application/read_service.go`
- Modify: `internal/modules/documents/approval/application/obsolete_service.go`
- Modify: `internal/modules/documents/repository/resolver_readers.go`
- Modify: `internal/modules/registry/delivery/http/routes.go`
- Modify: `internal/modules/jobs/stuck_instance_watchdog/job.go`
- Modify: `internal/test/e2e_seed.go`
- Modify tests that list `document_v2_id` columns.

Migration ordering:

1. Rename column.
2. Rename unique constraint and indexes to `document_id` names.
3. Update runtime code after migration exists.
4. Validate NOT VALID FKs after column rename migration.

Stop rules:

- Stop if any FK/index/constraint name is different in current migrations than expected.
- Stop if a migration would need destructive drop/recreate where `ALTER ... RENAME` is sufficient.

Verification:

```powershell
rg -n "document_v2_id|approval_instances_document_v2_id" internal apps tests migrations api frontend
GOFLAGS=-mod=mod go test ./internal/modules/documents/approval/... -count=1
go test ./internal/modules/jobs/... -count=1
go test ./internal/test/... -count=1
```

Expected:

- No runtime or test code references `document_v2_id`.
- Historical migrations may still contain the original creation name, but follow-up migrations must rename live schema objects.

### PR 10.4: Templates Editable Zones and Registry Legacy Cleanup

Goal: drop residual live template `editable_zones` column only after reference grep is clean; verify whether `profile_sequence_counters` is already retired.

Files likely touched:

- Add migration after approval migrations: `0196_drop_templates_editable_zones.sql`
- Optional migration only if live schema evidence requires it: `0197_drop_profile_sequence_counters_residual.sql`
- Modify template domain/repository code only if current grep finds live references outside historical migrations/artifacts.
- Modify registry docs/backlog if `profile_sequence_counters` is already retired by `0182`.

Preconditions:

```powershell
rg -n "editable_zones" internal apps api frontend packages migrations --glob "!migrations/0120_templates_v2_init.sql" --glob "!migrations/0157_drop_editable_zones.sql"
rg -n "profile_sequence_counters" internal apps api frontend packages tests migrations --glob "!migrations/0124_registry_controlled_documents.sql" --glob "!migrations/0128_grants_new_tables.sql" --glob "!migrations/0182_cd_sequence_per_area.sql"
```

Stop rules:

- Stop if `editable_zones` is still referenced by runtime code.
- Stop if `profile_sequence_counters` has runtime use or if only historical migrations reference it.
- Stop instead of deleting if table existence in live DB cannot be proven from migrations.

Verification:

```powershell
rg -n "editable_zones" internal apps api frontend packages
rg -n "profile_sequence_counters" internal apps api frontend packages tests
```

Expected:

- `editable_zones` appears only in historical migration context and ADR/wiki history.
- `profile_sequence_counters` appears only in historical migration context or not at all.

### PR 10.5: Editor UI and Approval Mechanical Cleanup

Goal: delete dead editor surfaces and approval naming/GUC duplication without changing behavior.

Editor files likely touched:

- Modify: `packages/editor-ui/src/index.ts`
- Delete: `packages/editor-ui/src/plugins/OutlinePlugin.tsx` only if consumer grep is clean.
- Move: `packages/editor-ui/src/plugins/mergefieldPlugin.ts` -> `packages/editor-ui/src/plugins/sidebarModelData.ts`
- Modify: `packages/editor-ui/src/plugins/sidebarModelBridge.ts`
- Modify: `packages/editor-ui/src/types.ts`
- Modify tests under `packages/editor-ui/test/`

Approval files likely touched:

- Move: `internal/modules/documents/approval/infra/signature/` -> `internal/modules/documents/approval/infrastructure/signature/`
- Modify imports in approval HTTP and tests.
- Delete or privatize `WithMembershipContext` only if all call sites can use `setAuthzGUC` or a single helper path.
- Modify tests tied only to `WithMembershipContext`.

Stop rules:

- Stop if `OutlinePlugin.tsx` has a production consumer.
- Stop if `onLockLost` is part of a documented public adapter contract still used by downstream packages.
- Stop if `WithMembershipContext` has behavior not covered by `setAuthzGUC` plus existing service tx boundaries.

Verification:

```powershell
rg -n "createOutlinePlugin|OutlinePlugin|mergefieldPlugin|onLockLost" packages frontend
rg -n "approval/infra/signature|WithMembershipContext" internal apps tests
cd packages/editor-ui; pnpm test
GOFLAGS=-mod=mod go test ./internal/modules/documents/approval/... -count=1
```

Expected:

- No dead editor exports remain.
- Approval signature imports use `infrastructure/signature`.
- One GUC helper path remains for approval service txs.

### PR 10.6: Registry Module Boundary and Taxonomy PK Cleanup

Goal: stop constructing a second registry repository in the composition root, and remove redundant taxonomy code-only PKs only if code proves tenant-safe assumptions.

Registry files likely touched:

- Modify: `internal/modules/registry/module.go`
- Modify: `apps/api/cmd/metaldocs-api/main.go`

Registry design:

- Keep module ownership of the repository.
- Store the existing `repo` in `Module` as a private field typed to the smallest interface needed by external wiring.
- Expose a narrow method such as `ControlledDocumentReader()` or `RepositoryReader()` only if it returns an interface needed by `cdRegistryAdapter` and documents wiring.
- Do not expose the concrete infrastructure type unless existing patterns force it.

Taxonomy precondition:

```powershell
rg -n "WHERE\s+code\s*=|WHERE [^\n]*(document_profiles|document_process_areas)[^\n]*code\s*=|GetByCode\(" internal apps tests
```

Stop rules:

- Stop if app code contains cross-tenant `WHERE code = $X` assumptions for profiles or areas.
- Stop if a dependent FK still references `code` alone and cannot be migrated mechanically to `(tenant_id, code)`.
- Stop if Postgres cannot drop the old PK without breaking existing FKs in the current migration chain.

Migration ordering:

1. Add or confirm tenant-safe composite uniqueness exists.
2. Migrate dependent FKs if any reference `code` alone.
3. Drop old code-only PK.
4. Promote `(tenant_id, code)` to primary key only after dependent FKs are aligned.

Verification:

```powershell
go test ./internal/modules/registry/... -count=1
go test ./internal/modules/taxonomy/... -count=1
rg -n "NewPostgresControlledDocumentRepository" apps internal/modules/registry
```

Expected:

- Composition root no longer constructs the second registry repository instance.
- Taxonomy PK cleanup lands only if preconditions are fully proven; otherwise defer with exact evidence.

### PR 10.7: Wiki, Acceptance Test, and Final Sweep

Goal: align docs/tests after code is stable.

Files likely touched:

- `wiki/tests/system-acceptance-test.md`
- `wiki/backlog/roadmap.md`
- Module docs/backlogs touched by Plan 10.
- `wiki/README.md` after wiki-curator pass.

Actions:

1. Update acceptance-test URLs from `/api/v2` to `/api/v1`.
2. Add explicit TODO block only if a manual harness route cannot be executed in this plan.
3. Mark Plan 10 backlog rows closed with commit references after implementation lands.
4. Update `wiki/backlog/roadmap.md` Plan 10 to done only after grep gates and tests pass.
5. Dispatch wiki-curator for link/path/route updates.

Verification:

```powershell
rg -n "/api/v2/" wiki/tests
rg -n "templates_v2|documents-v2|templates-v2" wiki --glob "!wiki/modules/*/_artifacts/**"
```

Expected:

- Wiki current-truth docs describe `/api/v1` and `templates`, not `templates_v2`.
- Historical artifacts may preserve old names only as evidence.

## 5. Full Verification Gate

Run after all workstream commits land:

```powershell
go test ./...
go build ./...
cd frontend/apps/web; pnpm gen:api; pnpm.cmd tsc --noEmit -p tsconfig.build.json; pnpm test; pnpm build
cd ../../..
cd packages/editor-ui; pnpm test; pnpm build
```

OpenAPI/codegen gate:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = "-mod=mod"
go generate ./...
cd frontend/apps/web; pnpm gen:api
```

Grep gates:

```powershell
rg -n "/api/v2/" internal frontend api/openapi wiki/tests
rg -n "internal/modules/templates_v2|metaldocs/internal/modules/templates_v2" internal apps api frontend packages
rg -n "document_v2_id|approval_instances_document_v2_id" internal apps tests api frontend
rg -n "editable_zones" internal apps api frontend packages
rg -n "profile_sequence_counters" internal apps api frontend packages tests
rg -n "createOutlinePlugin|OutlinePlugin|mergefieldPlugin|onLockLost" packages frontend
rg -n "approval/infra/signature|WithMembershipContext" internal apps tests
```

Acceptance gate:

- Run `wiki/tests/system-acceptance-test.md` after route and frontend builds are stable.
- Record any non-executed harness item as explicit deferred evidence, not silent omission.

## 6. Deferred Items Allowed Only With Evidence

Allowed deferrals inside Plan 10:

- Taxonomy T-015 if code or FK precondition proves tenant-safe PK migration is not mechanical.
- Registry R-100 if live DB/table existence cannot be proven beyond historical migration references.
- Any route/codegen change where generated signatures conflict with runtime ownership.

Not allowed:

- Keeping `/api/v2/*` aliases.
- Shipping both `templates_v2` and `templates` module dirs.
- Hand-writing frontend API types covered by OpenAPI.
- Adding Plan 12 screen work.
- New CSRF/auth architecture beyond the already wired `OriginProtection` verification.

## 7. Proposed Commit/PR Order

1. `docs(plan-10): correct roadmap drift and add purge rename spec`
2. `chore(templates): rename templates_v2 module and v1 routes`
3. `chore(api): canonicalize business routes under api v1`
4. `chore(approval): rename document_v2_id to document_id`
5. `chore(db): drop legacy editable zones after reference sweep`
6. `chore(editor-approval): remove dead editor and approval surfaces`
7. `chore(registry-taxonomy): close boundary leak and tenant-safe pk cleanup`
8. `docs(wiki): refresh Plan 10 links and acceptance URLs`

## 8. Approval Gate

Do not execute implementation until this spec is reviewed and explicitly approved.