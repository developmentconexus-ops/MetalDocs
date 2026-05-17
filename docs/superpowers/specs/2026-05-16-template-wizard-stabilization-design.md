# Template Wizard Stabilization Design

> Date: 2026-05-16
> Scope: Finish `/templates/new` as a production-grade wizard before moving to the next screen.
> Source review: `superpowers:requesting-code-review` on `codex/release-v2-name-polish`

## Goal

Finish the Template Creation Wizard as a focused, reviewable PR. The wizard should create profile-scoped or generic blank templates through the real `/api/v1/templates` contract, show only behavior backed by current runtime capability, and leave no known contract or migration hazard in the template creation path.

This design intentionally splits release-wide V2 naming cleanup into later PRs. The first repair PR fixes the wizard and the shared template surfaces it depends on. It does not try to make the global release inventory gate pass across the whole repository.

## Truth Classification

The repair follows the MetalDocs truth hierarchy:

1. Runtime truth: actual handlers, migrations, and wizard behavior.
2. Contract truth: OpenAPI, generated backend API, and generated frontend types.
3. Wiki truth: template module docs, wizard backlog, sync artifacts.
4. Execution truth: verification scripts, tests, runtime smoke, and review gates.

Review findings are classified as:

| Finding | Classification | PR 1 Action |
|---|---|---|
| `POST /api/v1/templates` spec/partial/runtime/frontend response mismatch | shared contract prerequisite | Fix in PR 1 |
| Empty slug can be bypassed via stepper | screen-local implementation | Fix in PR 1 |
| Raw `fetch` on active template wizard/catalog API paths | frontend API/TanStack prerequisite | Fix create path and catalog in PR 1 |
| Template table rename disables upgraded-DB tripwire function | database module-local implementation | Fix in PR 1 |
| Templates wiki still references `/templates/v2/placeholder-catalog` | wiki-memory drift | Fix in PR 1 |
| Nested screenshot artifacts under `frontend/apps/web/frontend/apps/web` | workflow/artifact cleanup | Fix in PR 1 |
| Global V2 inventory still has thousands of unexpected hits | workflow/tooling gap | Defer to release cleanup PR |
| Documents/registry e2e still references old V2 URLs | workflow/tooling gap | Defer unless it blocks wizard verification |
| Docgen-v2 naming is compatibility-sensitive | shared contract/deployment prerequisite | Defer to docgen classification PR |

## PR Boundary

### In Scope

- `/templates/new` wizard behavior and tests.
- `frontend/apps/web/src/features/templates/api/templates.ts` create-template path and placeholder catalog wrapper.
- `api/openapi/v1/partials/templates.yaml` and `api/openapi/v1/openapi.yaml` for template create response truth.
- Generated template backend API and generated frontend API types.
- Templates runtime handler code needed to match the fixed contract.
- Templates table rename post-baseline migration if kept in the branch.
- Fresh baseline/reference data only where needed to keep fresh and upgraded DB truth aligned.
- `wiki/backlog/novo-template-wizard.md`, `wiki/modules/templates.md`, and narrowly related sync artifacts.
- Wizard smoke evidence and screenshot artifact placement.

### Out of Scope

- Making `scripts/check-release-v2-names.ps1 -FailOnUnexpected` pass globally.
- Docgen-v2 service/package/env/event renaming or compatibility work.
- Documents and registry naming cleanup not needed by `/templates/new`.
- Historical docs cleanup outside active template wizard/module truth.
- Broad template editor refactors beyond API paths touched by the wizard.

## Contract Design

`POST /api/v1/templates` must have one canonical contract.

Request:

- `key`: required non-empty technical identifier submitted by the wizard.
- `name`: required display name.
- `description`: optional.
- `doc_type_code`: optional; present for profile-scoped templates, omitted for generic templates.
- `Idempotency-Key`: required by the frontend create path and runtime idempotency wrapper.

Response:

```json
{
  "data": {
    "template": { "...": "TemplateDTO fields" },
    "version": { "...": "VersionDTO fields" }
  }
}
```

The OpenAPI partial, bundled OpenAPI, generated Go server surface, generated frontend `paths` type, handler implementation, and `templates.ts` wrapper must agree on this shape. The `createTemplate` wrapper must derive its request and response types from the generated frontend API types after codegen.

No frontend code should depend on a hand-written create response shape that contradicts OpenAPI.

## Wizard Behavior Design

Step 1 keeps profile or generic scope selection. If a profile is selected, submit sends `doc_type_code`.

Step 2 shows the real technical key generated from the name. The key is not a fake sequence code. Names that slugify to an empty key, such as `!!!`, must keep the wizard at Step 2 regardless of whether the user clicks the footer button or the stepper. The error is displayed inline near the name field.

Step 3 exposes one implemented choice: blank template. DOCX import is visibly disabled and rendered outside the radio group with clear unavailable semantics. There must be one tabbable radio before selection.

Step 4 is read-only public visibility. It states the real current backend behavior and does not expose role/area/user-count controls until the create contract supports those fields.

Step 5 summarizes the real key, scope, blank start, public visibility, author, and draft version. Submit calls the shared API wrapper with an idempotency key and redirects to `/templates/{id}/versions/{n}` on success.

## Frontend API Design

The active create-template path uses `apiFetch` or the generated `api` client from `frontend/apps/web/src/lib/api/client.ts`.

Minimum PR 1 wrapper coverage:

- `createTemplate` sends `/api/v1/templates`.
- Method is `POST`.
- `Idempotency-Key` is sent.
- Payload includes `key`, `name`, optional `description`, and optional `doc_type_code`.
- Response parses `data.template` and `data.version`.
- Error handling flows through the canonical API error path.

`fetchPlaceholderCatalog` should also use the canonical API client surface, because it is part of the template wizard/editor supporting surface and currently bypasses auth/error/tracing behavior.

Other raw `fetch` calls in the template editor may remain only if they are outside the wizard PR boundary and are documented as follow-up debt.

## Database Design

If the branch keeps the table rename from:

- `templates_v2_template` to `templates_template`
- `templates_v2_template_version` to `templates_template_version`
- `templates_v2_approval_config` to `templates_approval_config`
- `templates_v2_audit_log` to `templates_audit_log`

then upgraded databases must preserve the same governance behavior as fresh baseline databases.

The post-baseline migration must:

- live under `db/migrations/` unless explicitly classified otherwise;
- be idempotent;
- write one `public.schema_migrations` row;
- rename tables, indexes, sequences, and constraints as needed;
- replace or update `public.enforce_capability_asserted()` so `TG_TABLE_NAME` matches the renamed template tables;
- keep `trg_require_cap_asserted` attached to the renamed tables;
- leave historical migrations unchanged.

The fresh baseline and product reference data must match the post-migration runtime object names.

## Wiki And Evidence Design

Update only active truth:

- `wiki/backlog/novo-template-wizard.md` reflects the final wizard behavior and deferred capabilities.
- `wiki/modules/templates.md` reflects canonical `/api/v1/templates/placeholder-catalog` and current generated operation names.
- Any sync log entries clearly name the change context and do not perform a broad release sweep.

Remove or relocate accidental files under:

```text
frontend/apps/web/frontend/apps/web/...
```

Wizard evidence belongs under the intended design-source artifact directory for the template wizard.

## Verification Design

PR 1 cannot be considered complete until these checks pass or a blocker is explicitly classified:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module templates
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/templates
$env:GOFLAGS="-mod=mod"; go generate ./internal/modules/templates/api/...
go test ./internal/modules/templates/... -count=1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1
```

Frontend:

```powershell
cd frontend/apps/web
pnpm.cmd gen:api
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm.cmd vitest run src/features/templates/api src/features/templates/components/wizard src/features/templates/pages/TemplateWizardPage.test.tsx
```

Runtime smoke:

- Open `/templates/new`.
- Select a profile and create a blank template.
- Confirm `POST /api/v1/templates` includes `doc_type_code`, non-empty `key`, and `Idempotency-Key`.
- Confirm names like `!!!` cannot advance or submit.
- Confirm Step 3 has one selectable blank radio and DOCX import is disabled outside the radio group.
- Confirm Step 4 is honest public visibility.
- Confirm redirect to `/templates/{id}/versions/{n}`.

## Follow-Up PRs

After PR 1, separate PRs should handle:

1. Release V2 inventory gate classification and global cleanup.
2. Documents/registry e2e URL cleanup and inventory scope update.
3. Docgen-v2 compatibility classification and rename/defer decision.
4. Remaining wiki release sweep outside active template wizard/module truth.

## Acceptance Criteria

- `/templates/new` is truthful, keyboard-accessible, and backed by real runtime capability.
- Template creation contract has one source of truth across OpenAPI, generated code, runtime, and frontend wrapper.
- Upgraded and fresh databases enforce the same template capability tripwire after table rename.
- Focused wizard/API tests cover the review failures.
- Template wizard/module wiki truth no longer contradicts runtime route/operation truth.
- The PR summary clearly names deferred release-wide V2 cleanup items.
