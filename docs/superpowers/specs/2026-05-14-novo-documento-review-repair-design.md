# Novo Documento Review Repair Design

> Date: 2026-05-14
> Status: approved for implementation planning
> Scope: Repair code-review findings for PR 1 visibility and PR 2 blank template.
> Out of scope: author-safe IAM people search endpoint, external sharing, generic ACL platform.

## Context

The implementation for novo-documento visibility and blank-template creation is directionally aligned with MetalDocs architecture:

- Registry owns controlled-document visibility because it owns controlled-document identity and atomic create.
- Visibility is enforced server-side.
- Blank document creation uses a real immutable system-owned template version.
- People picker remains deferred because no author-safe IAM search endpoint exists.

The code review found issues that block merge readiness. The repair should harden rollout safety, API truth, authz consistency, and wizard honesty without redesigning the feature into a broad ACL system.

## Goals

- Keep existing controlled documents visible after migration.
- Return persisted visibility grants accurately from registry reads.
- Preserve server-side visibility filtering before pagination.
- Use explicit template authz and immutable-template error semantics.
- Make the wizard area-visibility UI match the submitted contract.
- Keep people and external sharing deferred until real backend capability exists.

## Non-Goals

- Do not add an IAM user search endpoint in this repair.
- Do not implement external sharing.
- Do not introduce edit-permission semantics through visibility.
- Do not create a reusable access-policy subsystem.
- Do not alter the blank-template strategy away from a real system-owned template version.

## Registry Visibility Repair

Migration `0198_controlled_document_visibility.sql` must be rollout-safe:

- Add `visibility_scope TEXT NOT NULL DEFAULT 'company'`.
- Existing controlled documents remain company-visible after migration.
- New documents can still be restricted because create requests explicitly persist restricted visibility and grant rows.

Repository reads must return stored visibility truth:

- For `company`, responses return `{ scope: "company", areaCodes: [], userIds: [] }`.
- For `restricted`, responses return area and user grants from `controlled_document_area_grants` and `controlled_document_user_grants`.
- Detail reads may load grants for the single controlled document.
- List reads should batch-load grants for returned controlled-document IDs to avoid N+1 behavior.
- Visibility filtering remains SQL-side and happens before limit/offset pagination.

Service filtering should be conservative:

- Only inject actor-based read filtering when an actor user ID exists.
- Non-request service contexts should not accidentally become more restrictive because of an empty actor ID.

## Template Blank Repair

The system blank endpoint should follow template read conventions:

- `GET /api/v1/templates/system/blank` requires `template.view`.
- The endpoint may still resolve the system-owned sentinel row, but it should not bypass the normal read capability boundary.
- The cross-tenant system-template model should be explicit in code through constants and tests.

System template mutations should use a dedicated immutable error:

- Add `domain.ErrSystemTemplateImmutable`.
- Return that error from create-next-version, submit, review, approve, publish, archive, and any other currently protected system-owned mutation path.
- Map it to HTTP `409` with code `SYSTEM_TEMPLATE_IMMUTABLE`.
- Tests should assert the specific immutable conflict behavior, not archived semantics.

## Wizard Repair

The wizard should match the contract honestly:

- `Toda empresa` remains exclusive and clears restricted grants.
- `Areas selecionadas` shows real area grant controls from taxonomy data already loaded by Step 2.
- The selected document process area remains selected and locked for restricted visibility.
- Additional area grants can be selected and submitted in `visibility.areaCodes`.
- `Pessoas especificas` remains disabled or otherwise non-actionable until an author-safe IAM endpoint exists.
- `Compartilhamento externo` remains disabled/deferred.

No runtime fake data or mock backend behavior should be introduced.

## Tests and Verification

Required focused tests:

- Migration/repository behavior proving existing/default company visibility remains visible.
- Registry read tests proving restricted grant rows round-trip into API responses.
- Registry list/detail tests proving filtering uses stored grants while responses expose stored grants.
- Template HTTP tests proving `SYSTEM_TEMPLATE_IMMUTABLE` conflict behavior.
- Frontend wizard tests proving additional area selection reaches `visibility.areaCodes`.

Required verification commands:

```powershell
go test ./internal/modules/registry/... ./internal/modules/templates/... -count=1
cd frontend/apps/web
pnpm.cmd test src/features/documents/state/__tests__/wizard.reducer.test.ts src/features/documents/pages/NewDocumentWizardPage.test.tsx
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

Before final completion, rerun the contract/runtime gates touched by this repair:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module registry
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module templates
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```

## Acceptance Criteria

- Existing controlled documents are not hidden by the visibility migration.
- API responses reflect persisted visibility grant rows.
- Restricted visibility filtering remains backend-enforced.
- Blank template endpoint requires read authorization.
- System-owned template mutation errors are explicit and auditable.
- Wizard area grant controls submit real selected area codes.
- People and external sharing remain visibly deferred with no fake runtime behavior.
