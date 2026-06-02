# Plan: Approval Route Admin — QA-Pass Remediation

## Summary
Preview QA against `/approval-routes` exposed five regressions surviving PR-1..5: a FE input transform that breaks every create (P1a), a 500 on canonical create payload (P1b), missing per-row `version` in the live list response combined with a missing FE seed (P2), silent failure UX on all non-412 errors (P3), and a read/write stage-shape divergence in OpenAPI (P5, with P4 latent on it). Fix in a 2-PR sequence: **PR-A** repairs the backend contract + observability + create-500 root cause; **PR-B** repairs the frontend ETag seeding, error UX, optimistic math, and profile_code transform. Close via re-run of `metaldocs-screen-qa` against `/approval-routes` with all seven gates green.

## User Story
As a `system_admin` operating the Approval Route Admin screen,
I want to create, edit, and deactivate routes from a cold page load,
So that governance configuration is actually usable without a backend dev opening a console to fish out the real error.

## Problem → Solution
**Current**: 5 of 7 QA gates blocked. Create → 400 then 500. Edit/Deactivate → 428 on cold load. All 4xx/5xx surface as silent failure. 88 vitest pass because they mock `routeAdminApi` and never see the real list payload.
**Desired**: All 7 gates pass against live `/approval-routes`. Error envelopes carry an actionable code, mutations show a toast, `If-Match` is always sent on edit/deactivate after a list read, and the OpenAPI list/write stage shape is unified.

## Metadata
- **Complexity**: Medium (cross-cutting BE+FE, no new subsystems, ~10 files)
- **Source PRD**: N/A (autonomous QA report from session)
- **PRD Phase**: N/A
- **Estimated Files**: 10–12

---

## UX Design

### Before
```
┌────────────────────────────────────────────────────┐
│ Administração de Rotas              [Nova rota] ←  │
│ ┌──────────────────────────────────────────────┐   │
│ │ QA Inbox Route   po   1 etapa   Ativa   …    │   │
│ └──────────────────────────────────────────────┘   │
│                                                    │
│ User clicks "Nova rota" → dialog                   │
│ Fills "qa-preview" → input shows "QA-PREVIEW"      │
│ Salvar → silently fails (400)                      │
│ Cancel + Edit existing → silently fails (428)      │
│ Cancel + Desativar      → silently fails (428)     │
│ No toast, no log, dialog stays open.               │
└────────────────────────────────────────────────────┘
```

### After
```
┌────────────────────────────────────────────────────┐
│ Administração de Rotas              [Nova rota]    │
│                                                    │
│ Type "qa-preview" → input keeps "qa-preview"       │
│ (visual uppercase via text-transform only)         │
│ Salvar → 201 → row appears, toast "Rota criada"    │
│                                                    │
│ Click Edit on cold-loaded row → dialog opens,      │
│ change name, Salvar → 200, toast "Rota atualizada" │
│ row name updates, version bumped                   │
│                                                    │
│ Click Desativar → dialog, reason → 200,            │
│ toast "Rota desativada", status flips              │
│                                                    │
│ Stale ETag (refetch between open and save) →       │
│ toast "Rota foi alterada, recarregue", row reloads │
└────────────────────────────────────────────────────┘
```

### Interaction Changes
| Touchpoint | Before | After | Notes |
|---|---|---|---|
| profile_code input | uppercased in state, sent uppercase | stored lowercase in state, displayed via `text-transform: uppercase` CSS | drop `.toUpperCase()` mutator |
| Cold-load Edit | 428 | 200 (If-Match seeded from list) | listRoutes seeds etagCache |
| Cold-load Desativar | 428 | 200 | same |
| Save fails (any non-412 4xx/5xx) | silent | toast with `resolveErrorMessage` text, dialog stays open | mutations onError |
| 412 stale ETag | toast (existing) | unchanged | already wired in mutationClient.ts:60-64 |
| BE 500 on create | opaque `internal.unknown` | actionable code (e.g. `validation.capability_unknown`, `validation.role_unknown`), 4xx not 500 | real cause surfaced; 500 only for genuine internal bugs |

---

## Mandatory Reading

| Priority | File | Lines | Why |
|---|---|---|---|
| P0 | `internal/modules/documents/approval/http/route_admin_handler.go` | 1-265 | Handler for all 4 routes; `mapStageRequests` defaults `required_capability` to `workflow.sign` (line 246) — likely root cause of 500 if registry rejects it |
| P0 | `internal/modules/documents/approval/application/route_admin_service.go` | full | Create/Update/Deactivate orchestrators; where idempotency store, capability check, and DB tx live |
| P0 | `internal/modules/documents/approval/http/errors.go` | 1-260 | `MapErrorToResponse` + `WriteError`; >=500 already logs via `slog.Error` (errors.go:242-247) — confirm sink, add missing 4xx mappings for unknown role/capability |
| P0 | `api/openapi/v1/openapi.yaml` | 5596-5699 | StageRequest (canonical: `name`/`order`/`quorum`/`drift_policy`), StageSummary (legacy: `label`/`quorum_kind`), RouteSummary (already has `version`) |
| P0 | `internal/modules/documents/approval/http/contracts/route.go` | 150-210 | Go structs `ListRouteItem`+`ListStageItem`+`ListRoutesResponse`; field json tags |
| P0 | `frontend/apps/web/src/features/approval/api/routeAdminApi.ts` | full | Transport layer; `listRoutes` at line 28 doesn't seed etagCache; `seedRouteEtag` at 91-100 already correct format `"v<N>"` |
| P0 | `frontend/apps/web/src/features/approval/api/mutationClient.ts` | full | Already parses ETag from response header (line 45-48), routes 412 to toast, throws otherwise |
| P0 | `frontend/apps/web/src/features/approval/queries/useRouteAdminMutations.ts` | 60-180 | All three mutations; `onError` only rolls back cache, no toast |
| P0 | `frontend/apps/web/src/features/approval/pages/route-admin/RouteEditorDialog.tsx` | 290-320 | `setDraft(... event.target.value.toUpperCase())` at line 314 |
| P1 | `frontend/apps/web/src/lib/api/problem.ts` | 140-170 | `resolveErrorMessage(err)` signature — use in mutations onError |
| P1 | `frontend/apps/web/src/lib/api/errorMessages.ts` | full | pt-BR error code mapping; add `validation.role_unknown` etc. if new |
| P1 | `internal/modules/documents/approval/repository/postgres_approval_repository.go` | 440-490 | List query already selects `r.version`; confirms BE has the field but stale binary may not |
| P2 | `wiki/modules/approval.md` | full | Module Arc42; route truth table referenced by 4 ADRs |
| P2 | `wiki/decisions/0018-approval-route-lifecycle.md` | full | ADR added in PR-3; covers If-Match contract — update if FE seeding becomes part of the contract |
| P2 | `scripts/start-api.ps1` | full | Script-truth policy. PR rebuild rule applies — must use `-Build` to validate plan |

## External Documentation
| Topic | Source | Key Takeaway |
|---|---|---|
| TanStack Query mutation onError | https://tanstack.com/query/v5/docs/framework/react/guides/mutations | onError fires for any thrown/rejected error; return value ignored. Safe to call `toast.error()` inside. |
| RFC 9457 problem+json | local: `wiki/architecture/api-contract.md` | Already canonical; no new patterns needed. |

No external research needed beyond confirming TanStack onError semantics — feature uses established internal patterns.

---

## Patterns to Mirror

### NAMING_CONVENTION
```ts
// SOURCE: frontend/apps/web/src/features/approval/api/routeAdminApi.ts:20-26
export type RouteSummary = components['schemas']['RouteSummary'];
export type ListRoutesResponse = components['schemas']['ListRoutesResponse'];
export type CreateRouteRequest = components['schemas']['CreateRouteRequest'];
// Never re-type locally; always pull from codegen.
```

### ERROR_HANDLING — backend mapping
```go
// SOURCE: internal/modules/documents/approval/http/errors.go:200-235
// Add new branches here when domain returns a new sentinel; do NOT let it
// fall through to default 500.
switch {
case errors.Is(err, application.ErrRouteRoleUnknown):
    statusCode = http.StatusBadRequest
    code = approvalCodeValidationRoleUnknown
case errors.Is(err, application.ErrRouteCapabilityUnknown):
    statusCode = http.StatusBadRequest
    code = approvalCodeValidationCapabilityUnknown
// existing branches ...
}
```

### ERROR_HANDLING — frontend toast
```ts
// SOURCE: existing pattern in frontend/apps/web/src/features/approval/queries/useSignoffMutation.ts (see signoff flow)
onError: (err, _vars, context) => {
  if (context?.previous !== undefined) {
    queryClient.setQueryData(queryKey, context.previous);
  }
  // 412 already shows toast inside mutationClient; skip duplicate.
  if (err instanceof ApprovalError && err.status === 412) return;
  toast.error(resolveErrorMessage(err));
},
```

### LOGGING_PATTERN
```go
// SOURCE: internal/modules/documents/approval/http/errors.go:242-247
if prob.Status >= http.StatusInternalServerError {
    slog.Error("approval handler error",
        slog.Int("status", prob.Status),
        slog.String("code", string(prob.Code)),
        slog.Any("error", err),
    )
}
// New: log INFO on create with idempotency_key + tenant_id for trace.
```

### REPOSITORY_PATTERN
```go
// SOURCE: internal/modules/documents/approval/repository/postgres_approval_repository.go:444
SELECT r.id, r.name, r.tenant_id::text, r.profile_code, r.active, r.version,
       r.created_at, r.created_at AS updated_at, ...
// r.version already selected; only mapping/serialization layer may be dropping it.
```

### SERVICE_PATTERN
```go
// SOURCE: internal/modules/documents/approval/http/route_admin_handler.go:241-260
func mapStageRequests(stages []contracts.StageRequest) []domain.Stage {
    out := make([]domain.Stage, 0, len(stages))
    for _, s := range stages {
        cap := strings.TrimSpace(s.RequiredCapability)
        if cap == "" {
            cap = "workflow.sign"  // ← suspect default; replace with "document.signoff" or fail-fast
        }
        // ...
    }
}
```

### TEST_STRUCTURE — Go handler
```go
// SOURCE: internal/modules/documents/approval/http/route_admin_handler_test.go:34-77
type fakeRouteAdminService struct {
    listResult application.ListRoutesResult
    listErr    error
    // ...
}
mux.HandleFunc("POST /api/v1/approval/routes", h.CreateRouteHandler)
// Use table-driven sub-tests; one slot per error sentinel.
```

### TEST_STRUCTURE — Vitest mutation
```ts
// SOURCE: frontend/apps/web/src/features/approval/pages/route-admin/RouteAdminPage.test.tsx
// Add tests that DO NOT mock routeAdminApi — instead mock fetch and assert
// real request shape + ETag seed behavior + toast emission.
const fetchMock = vi.fn().mockResolvedValueOnce({
  ok: true,
  json: async () => ({ routes: [{ id, name, profile_code, active, version: 3, /* ... */ }], total: 1 }),
  headers: new Headers(),
});
// After list resolves, etagCache.get(id) must equal '"v3"'.
```

---

## Files to Change

| File | Action | Justification |
|---|---|---|
| `api/openapi/v1/openapi.yaml` | UPDATE | Unify `StageSummary` field names with `StageRequest` (canonical: `name`/`order`/`quorum`); deprecate `label`/`quorum_kind` or document compat shim |
| `internal/modules/documents/approval/http/contracts/route.go` | UPDATE | Adjust `ListStageItem` json tags to match new canonical schema; keep `Label` field but tag as `json:"name"` (or rename) |
| `internal/modules/documents/approval/http/route_admin_handler.go` | UPDATE | `mapListRoute` writes canonical field names; `mapStageRequests` default capability fix (P1b suspect) |
| `internal/modules/documents/approval/http/errors.go` | UPDATE | Add 4xx mappings for `ErrRouteRoleUnknown`, `ErrRouteCapabilityUnknown`, `ErrRouteAreaUnknown`; ensure no domain sentinel falls to default 500 |
| `internal/modules/documents/approval/application/route_admin_service.go` | UPDATE | Wrap unknown role/capability/area lookups with sentinel errors; add structured INFO log on create with idempotency_key + tenant_id |
| `internal/modules/documents/approval/application/errors.go` (new sentinels) | UPDATE | Define `ErrRouteRoleUnknown` etc. (or extend existing file) |
| `internal/modules/documents/approval/http/route_admin_handler_test.go` | UPDATE | Add table-driven cases for new sentinels → 400; for list response with `version` populated |
| `scripts/start-api.ps1` | UPDATE | Pipe API stdout/stderr to `logs/api.log` (or rotate) so slog output is inspectable from same session |
| `frontend/apps/web/src/lib/api-types/index.d.ts` | UPDATE | Regenerate codegen after OpenAPI change |
| `frontend/apps/web/src/features/approval/api/routeAdminApi.ts` | UPDATE | `listRoutes` seeds `etagCache` per row via `seedRouteEtag(r.id, r.version)` |
| `frontend/apps/web/src/features/approval/queries/useRouteAdminMutations.ts` | UPDATE | All three `onError` callbacks call `toast.error(resolveErrorMessage(err))` for non-412; fix optimistic `route.version + 1` to guard against undefined (defense-in-depth) |
| `frontend/apps/web/src/features/approval/pages/route-admin/RouteEditorDialog.tsx` | UPDATE | Remove `.toUpperCase()` at line 314; add `text-transform: uppercase` to input CSS for display only |
| `frontend/apps/web/src/features/approval/pages/route-admin/RouteAdmin.module.css` | UPDATE | Add `.profileCodeInput { text-transform: uppercase; }` |
| `frontend/apps/web/src/features/approval/pages/route-admin/RouteAdminPage.test.tsx` | UPDATE | Add tests that drive real fetch shape (no `routeAdminApi` mock) for: list seeds ETag, mutation error shows toast, profile_code submits lowercase |
| `wiki/modules/approval.md` | UPDATE | Refresh route truth table + stage shape if OpenAPI changes; bump `Last verified` |
| `wiki/decisions/0018-approval-route-lifecycle.md` | UPDATE | Add clarification that list response carries `version` for seed; ETag format remains `"v<N>"` |

## NOT Building

- **Tier-1 capability split** for `/api/v1/approval/routes` (BE-9 / F-001 ADR 0016 follow-up). Explicitly deferred; tracked separately.
- **Full module schema migration** to remove `additionalProperties: true`. The 6 affected schemas already in canonical form except StageSummary; rest is its own track.
- **E2E Playwright suite**. This plan adds vitest coverage at the boundary; Playwright is a separate workstream.
- **New seed routes** in `db/seeds/`. Existing "QA Inbox Route" is sufficient; if create-path needs additional fixture, add it as part of QA evidence, not the plan.
- **Refactoring `ListRoutesHandler` to use repository pagination**. Out of scope; only mapping layer changes.

---

## Step-by-Step Tasks

### Task 1: Pipe API stdout to log file
- **ACTION**: Edit `scripts/start-api.ps1` to redirect stdout+stderr of `metaldocs-api.exe` to `logs/api.log` with timestamped rotation.
- **IMPLEMENT**:
  ```powershell
  $logDir = Join-Path $PSScriptRoot "..\logs"
  if (-not (Test-Path $logDir)) { New-Item -ItemType Directory -Path $logDir | Out-Null }
  $logFile = Join-Path $logDir "api.log"
  & $apiBinary *>&1 | Tee-Object -FilePath $logFile -Append
  ```
- **MIRROR**: existing PowerShell error-handling style in `scripts/start-api.ps1`.
- **IMPORTS**: none (PowerShell built-ins).
- **GOTCHA**: `Tee-Object` blocks; if script must stay non-blocking, use `Start-Process -RedirectStandardOutput` instead and keep current shell free. Decide based on what `start-api.ps1` currently does.
- **VALIDATE**: rebuild + start API, hit a known 500, confirm `logs/api.log` shows the `slog.Error("approval handler error", ...)` line with structured fields.

### Task 2: Diagnose P1b — capture real create-500 error
- **ACTION**: With Task 1 log in place, replay POST `/api/v1/approval/routes` (canonical body, lowercase profile_code) and capture the slog line.
- **IMPLEMENT**: re-run the QA `fetch(...)` from preview session; tail `logs/api.log`; record exact error.
- **MIRROR**: errors.go:242-247 slog format.
- **IMPORTS**: n/a.
- **GOTCHA**: if log shows nothing, the error is rejected before the handler returns 5xx → unlikely; otherwise check that `MapErrorToResponse` falls to default 500 branch.
- **VALIDATE**: real cause documented in PR description; one of {capability unknown, role unknown, area unknown, idempotency store insert, tenant GUC missing}.

### Task 3: Add domain sentinels + 4xx mappings
- **ACTION**: For each unknown-reference cause surfaced in Task 2, add an `ErrRoute*Unknown` sentinel in application layer and a 400 branch in `MapErrorToResponse`.
- **IMPLEMENT**:
  ```go
  // application package
  var (
      ErrRouteRoleUnknown        = errors.New("approval route: required_role unknown for tenant")
      ErrRouteCapabilityUnknown  = errors.New("approval route: required_capability not registered")
      ErrRouteAreaUnknown        = errors.New("approval route: area_code unknown for tenant")
      ErrRouteStageInvalid       = errors.New("approval route: stage invalid")
  )
  ```
  ```go
  // errors.go MapErrorToResponse switch additions
  case errors.Is(err, application.ErrRouteRoleUnknown):
      statusCode = http.StatusBadRequest
      code = approvalCodeValidationRoleUnknown
  // ...
  ```
- **MIRROR**: errors.go:200-235 existing switch.
- **IMPORTS**: `"errors"` already there; add new approvalCode* consts.
- **GOTCHA**: existing `additionalProperties: true` on Problem schema lets extra fields ride; no OpenAPI change required for new codes — but document codes in `wiki/architecture/api-design-system.md` or `errorMessages.ts`.
- **VALIDATE**: `go test ./internal/modules/documents/approval/...` green; new `route_admin_handler_test.go` cases for each sentinel → 400.

### Task 4: Fix default required_capability
- **ACTION**: In `mapStageRequests` (route_admin_handler.go:241-260), replace `"workflow.sign"` default with `"document.signoff"`, OR fail-fast with `ErrRouteCapabilityUnknown` when blank (preferred — explicit > defaulted).
- **IMPLEMENT**:
  ```go
  cap := strings.TrimSpace(s.RequiredCapability)
  if cap == "" {
      // OpenAPI marks required_capability required; reject explicitly
      // instead of silently defaulting to a code that may not be registered.
      return nil, application.ErrRouteCapabilityUnknown
  }
  ```
  (Returning an error requires changing the function signature; alternatively keep void but populate empty and let domain layer reject — pick one based on actual code shape from Task 2.)
- **MIRROR**: validation pattern in `req.Validate()` calls inside the same handler.
- **IMPORTS**: `"metaldocs/internal/modules/documents/approval/application"`.
- **GOTCHA**: OpenAPI marks `required_capability` as required on StageRequest, so blank should already be caught by `req.Validate()`. Confirm; if true this branch is dead.
- **VALIDATE**: existing test for create with omitted capability returns 400 not 500.

### Task 5: Unify StageSummary canonical field names
- **ACTION**: In `api/openapi/v1/openapi.yaml:5615-5629`, rename `label` → `name`, `quorum_kind` → `quorum`; reorder for required block. Drop `members` if absent in domain (already absent in struct; only added in mapping). If a back-compat shim is needed, keep `label` as a deprecated alias for one release.
- **IMPLEMENT**:
  ```yaml
  StageSummary:
    type: object
    additionalProperties: true
    required: [name, order, required_role, required_capability, area_code, quorum, drift_policy]
    properties:
      name: { type: string }
      order: { type: integer, minimum: 1 }
      required_role: { type: string }
      required_capability: { type: string }
      area_code: { type: string }
      quorum: { $ref: '#/components/schemas/QuorumKind' }
      quorum_m: { type: integer, minimum: 1, nullable: true }
      drift_policy: { $ref: '#/components/schemas/DriftPolicy' }
  ```
- **MIRROR**: `StageRequest` definition at openapi.yaml:5596-5614 (canonical baseline).
- **IMPORTS**: n/a.
- **GOTCHA**: this is a contract change. Existing FE list code reads `stage.label`/`stage.quorum_kind` in `RouteEditorDialog.tsx:86-90`. Must update FE *in the same change*; otherwise list breaks. No external API consumers documented.
- **VALIDATE**: `pnpm exec openapi-typescript api/openapi/v1/openapi.yaml -o frontend/apps/web/src/lib/api-types/index.d.ts` produces new types; `pnpm tsc` fails until FE-side reads updated; fix in Task 7.

### Task 6: Update Go ListRouteItem mapping to canonical names
- **ACTION**: Rename `ListStageItem.Label` → `Name`, `QuorumKind` → `Quorum`; update json tags accordingly. Drop the synthesized `Members: []string{stage.RequiredRole}` if it's only a legacy carry-over.
- **IMPLEMENT**:
  ```go
  // contracts/route.go
  type ListStageItem struct {
      Order              int             `json:"order"`
      Name               string          `json:"name"`
      RequiredRole       string          `json:"required_role"`
      RequiredCapability string          `json:"required_capability"`
      AreaCode           string          `json:"area_code"`
      Quorum             QuorumKind      `json:"quorum"`
      QuorumM            *int            `json:"quorum_m,omitempty"`
      DriftPolicy        DriftPolicyKind `json:"drift_policy"`
  }
  ```
  ```go
  // route_admin_handler.go mapListRoute
  stages = append(stages, contracts.ListStageItem{
      Order:              stage.Order,
      Name:               stage.Name,
      RequiredRole:       stage.RequiredRole,
      RequiredCapability: stage.RequiredCapability,
      AreaCode:           stage.AreaCode,
      Quorum:             contracts.QuorumKind(stage.Quorum),
      QuorumM:            stage.QuorumM,
      DriftPolicy:        contracts.DriftPolicyKind(stage.DriftPolicy),
  })
  ```
- **MIRROR**: existing `RouteSummary` struct + handler tests; ensure golden JSON snapshots updated.
- **IMPORTS**: existing.
- **GOTCHA**: `repository.Route.Stages` field shape must already carry `Order` and `Name`; if it uses `Label`, fix at the repository scan layer too. Check `postgres_approval_repository.go:444+`.
- **VALIDATE**: `go test ./internal/modules/documents/approval/...`; live GET response now includes `version` + canonical stage names. Confirm with `curl http://localhost:8081/api/v1/approval/routes`.

### Task 7: Regenerate FE OpenAPI types + fix readers
- **ACTION**: Run codegen, then update `RouteEditorDialog.tsx:82-92` to read canonical field names.
- **IMPLEMENT**:
  ```ts
  // RouteEditorDialog.tsx
  stages: route.stages.map((stage) => ({
      uid: crypto.randomUUID(),
      label: stage.name,                       // FE-local field name stays `label` for now
      requiredRole: stage.required_role ?? '',
      requiredCapability: stage.required_capability || 'document.signoff',
      areaCode: stage.area_code ?? '',
      quorumKind: stage.quorum,                // was stage.quorum_kind
      m: String(stage.quorum_m ?? 1),
      driftPolicy: stage.drift_policy,
  })),
  ```
- **MIRROR**: existing translation block in the same file.
- **IMPORTS**: regenerated `components['schemas']['StageSummary']`.
- **GOTCHA**: tests in `RouteAdminPage.test.tsx` mock `routeAdminApi` so they won't fail on shape drift — must add an unmocked test (Task 11) before relying on green CI.
- **VALIDATE**: `pnpm tsc --noEmit -p frontend/apps/web/tsconfig.build.json` exit 0.

### Task 8: Seed etagCache from listRoutes
- **ACTION**: In `routeAdminApi.ts:28-46`, after `res.json()` resolves, walk `data.routes` and call `seedRouteEtag(r.id, r.version)` for each row.
- **IMPLEMENT**:
  ```ts
  export async function listRoutes(): Promise<ListRoutesResponse> {
    const res = await fetch(BASE);
    if (res.ok) {
      const data = (await res.json()) as ListRoutesResponse;
      for (const r of data.routes ?? []) {
        if (typeof r.version === 'number') {
          seedRouteEtag(r.id, r.version);
        }
      }
      return data;
    }
    // ... existing error path unchanged
  }
  ```
- **MIRROR**: `seedRouteEtag` itself at routeAdminApi.ts:91-100 (monotonic guard built in).
- **IMPORTS**: `seedRouteEtag` is already exported from the same file — no import change.
- **GOTCHA**: if backend `version` is `0` for legacy seed rows, `seedRouteEtag` skips (guard at line 92 `if (!Number.isFinite(version) || version <= 0) return`). DB migration may be needed for the seed; flag it but don't block the FE wiring.
- **VALIDATE**: after page load, `etagCache.get('<route-id>')` returns `"v<N>"`; subsequent Edit click results in PUT with `If-Match: "v<N>"` → 200.

### Task 9: Wire toast on mutation errors
- **ACTION**: Add `toast.error(resolveErrorMessage(err))` (skipping 412) to `onError` in all three mutations.
- **IMPLEMENT**:
  ```ts
  // useRouteAdminMutations.ts — repeat for create, update, deactivate
  import { toast } from 'sonner';
  import { ApprovalError } from '../api/mutationClient';
  import { resolveErrorMessage } from '../../../lib/api';

  onError: (err, _vars, context) => {
    if (context?.previous !== undefined) {
      queryClient.setQueryData(queryKey, context.previous);
    }
    if (err instanceof ApprovalError && err.status === 412) return; // toast already fired in mutationClient
    toast.error(resolveErrorMessage(err));
  },
  ```
- **MIRROR**: `mutationClient.ts:62-64` for the 412 path (do not duplicate it).
- **IMPORTS**: `toast` from `sonner`; `ApprovalError` from `'../api/mutationClient'`; `resolveErrorMessage` from `'../../../lib/api'`.
- **GOTCHA**: do not import `toast` in a way that breaks SSR; project is CSR-only via Vite, safe.
- **VALIDATE**: vitest case where `mutationFn` throws an ApprovalError(400) asserts a toast was emitted.

### Task 10: Drop profile_code uppercase mutator
- **ACTION**: In `RouteEditorDialog.tsx:314`, remove `.toUpperCase()`. Add CSS class for visual uppercase.
- **IMPLEMENT**:
  ```tsx
  // RouteEditorDialog.tsx:312-316
  <input
    id="route-profile-code"
    className={styles.profileCodeInput}
    value={draft.profileCode}
    onChange={(event) =>
      setDraft((prev) => ({ ...prev, profileCode: event.target.value }))
    }
    ...
  />
  ```
  ```css
  /* RouteAdmin.module.css */
  .profileCodeInput {
    text-transform: uppercase;
  }
  ```
- **MIRROR**: existing `<input>` rows in the same component (no transform).
- **IMPORTS**: none new.
- **GOTCHA**: `text-transform: uppercase` only affects display, not the underlying value. Confirm by typing a value and submitting — wire log inside the same task to assert lowercase reaches the network.
- **VALIDATE**: vitest input simulation; assert request body `profile_code` is lowercase.

### Task 11: Add unmocked Vitest tests for the real shape
- **ACTION**: Add tests to `RouteAdminPage.test.tsx` (or sibling file) that mock `fetch` directly — not `routeAdminApi` — to assert:
  1. List response with `version: 3` causes `etagCache` to hold `"v3"`.
  2. Create with `profileCode='qa-preview'` → POST body contains `profile_code: 'qa-preview'` (lowercase).
  3. PUT after list-only load includes `If-Match: "v3"` header.
  4. POST returning 400 with `code: 'validation.role_unknown'` triggers `toast.error` with mapped pt-BR text.
- **IMPLEMENT**:
  ```ts
  vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));
  const fetchMock = vi.fn();
  vi.stubGlobal('fetch', fetchMock);
  // ...arrange list response, render page, await render, assertions
  ```
- **MIRROR**: existing `RouteAdminPage.test.tsx:1-50` (file already imports React Query + render helpers).
- **IMPORTS**: `vi`, `vi.stubGlobal`, `etagCache` from `'../../api/etagCache'`.
- **GOTCHA**: `etagCache` is module-state; reset between tests with `etagCache.clear()` in `beforeEach`.
- **VALIDATE**: `pnpm exec vitest run src/features/approval` shows ≥ 4 new tests passing.

### Task 12: Refresh wiki
- **ACTION**: Bump `wiki/modules/approval.md` route truth table (stage shape) and `Last verified:` stamp; append a postscript to `wiki/decisions/0018-approval-route-lifecycle.md` documenting the list-seeds-ETag rule.
- **IMPLEMENT**: edit files; do not rebuild module wiki (single bounded change).
- **MIRROR**: existing stamp format in other wiki module docs.
- **IMPORTS**: n/a.
- **GOTCHA**: post-implementation invocation of `wiki-curator` agent per CLAUDE.md §2 is preferred, but manual edit is acceptable when the change is bounded.
- **VALIDATE**: file:line anchors in updated docs resolve; rebuild not required.

### Task 13: Re-run Preview QA
- **ACTION**: `start-api.ps1 -Build`, `pnpm --prefix frontend/apps/web run dev`, navigate `/approval-routes` as admin, run all seven gates of `metaldocs-screen-qa`.
- **IMPLEMENT**: follow `wiki/quality/screen-qa-checklist.md`.
- **MIRROR**: prior QA report style (this session's earlier output).
- **IMPORTS**: n/a.
- **GOTCHA**: ETag cache persists across hot-reload but resets on full page reload — test both.
- **VALIDATE**: all 7 gates green; QA report attached to PR.

---

## Testing Strategy

### Unit Tests

| Test | Input | Expected Output | Edge Case? |
|---|---|---|---|
| `listRoutes_seedsEtagCachePerRow` | list response `[{id:'a',version:3},{id:'b',version:1}]` | `etagCache.get('a')==='"v3"' && etagCache.get('b')==='"v1"'` | yes (cold load) |
| `listRoutes_skipsZeroVersion` | row with `version:0` | `etagCache.get('id') === undefined` | yes (legacy seed) |
| `createRoute_sendsLowercaseProfileCode` | input `'qa-preview'` | fetch body `profile_code: 'qa-preview'` | no |
| `updateRoute_includesIfMatchAfterList` | list then update on same row | PUT has `If-Match: "v3"` | yes (cold-load edit) |
| `mutationOnError_400_emitsToast` | mock fetch 400 with role_unknown | `toast.error` called once with pt-BR text | yes |
| `mutationOnError_412_doesNotDuplicateToast` | mock fetch 412 | exactly one toast (from mutationClient, not the mutation onError) | yes |
| `mapErrorToResponse_routeRoleUnknown` (Go) | `application.ErrRouteRoleUnknown` | `prob.Status==400 && prob.Code=="validation.role_unknown"` | yes |
| `mapErrorToResponse_routeCapabilityUnknown` (Go) | same | 400 + canonical code | yes |
| `mapListRoute_canonicalStageNames` (Go) | repository.Route with stages | JSON marshals with `name` (not `label`) and `quorum` (not `quorum_kind`) | yes |
| `createHandler_emptyCapability` | StageRequest with `RequiredCapability:""` | 400 not 500 | yes |

### Edge Cases Checklist
- [ ] Empty input → 400 with mapped code (no 500)
- [ ] Maximum size input → OpenAPI maxLength enforced
- [ ] Invalid types → JSON decode error → 400
- [ ] Concurrent access → If-Match mismatch → 412 with toast
- [ ] Network failure → caught by mutationClient → toast emitted
- [ ] Permission denied → 403 → toast with mapped text
- [ ] Legacy seed row with `version:0` → seedRouteEtag no-op, but BE still returns 200 on read; edit attempt → 428 with explicit toast directing the user to refresh

---

## Validation Commands

### Static Analysis
```bash
go vet ./internal/modules/documents/approval/...
corepack pnpm --prefix frontend/apps/web tsc --noEmit -p tsconfig.build.json
```
EXPECT: zero output.

### Unit Tests
```bash
go test ./internal/modules/documents/approval/... -race -count=1
corepack pnpm --prefix frontend/apps/web exec vitest run src/features/approval
```
EXPECT: all pass, new tests included.

### Full Test Suite
```bash
go test ./... -race -count=1
corepack pnpm --prefix frontend/apps/web exec vitest run
```
EXPECT: no regressions.

### OpenAPI Codegen
```bash
corepack pnpm --prefix frontend/apps/web exec openapi-typescript ../../../api/openapi/v1/openapi.yaml -o src/lib/api-types/index.d.ts
```
EXPECT: file updates; `git diff` shows only StageSummary rename.

### Browser Validation
```powershell
.\scripts\start-api.ps1 -Build
# in another shell
corepack pnpm --prefix frontend/apps/web run dev -- --port 4173 --strictPort
```
EXPECT: API on :8081, web on :4173. Navigate /approval-routes, run all 7 QA gates.

### Manual Validation
- [ ] Cold-load `/approval-routes` → list renders with existing rows
- [ ] Inspect `etagCache` via devtools: every row id has a `"v<N>"` entry
- [ ] Create route with `qa-preview` → 201, toast success, row appears
- [ ] Edit existing row → 200, toast success, name updates
- [ ] Deactivate row → 200, toast success, status flips to "Inativa"
- [ ] Force stale ETag (refetch then save) → 412 toast "Rota foi alterada"
- [ ] Console: no errors (HydrateFallback warn acceptable)
- [ ] Network: no 4xx/5xx during happy path

---

## Acceptance Criteria
- [ ] All 13 tasks completed
- [ ] All validation commands pass
- [ ] ≥ 4 new unmocked Vitest tests added
- [ ] ≥ 3 new Go handler tests added (per-sentinel)
- [ ] No type errors, no lint errors
- [ ] QA report attached showing 7/7 gates green
- [ ] Wiki updated (`approval.md`, `0018-approval-route-lifecycle.md`)

## Completion Checklist
- [ ] Code follows discovered patterns (errors.go switch, sonner toast, codegen-only types)
- [ ] Error handling matches codebase style (RFC 9457 + ApprovalError)
- [ ] Logging follows codebase conventions (`slog.Error` with structured fields)
- [ ] Tests follow existing patterns (vitest mocks fetch not routeAdminApi; Go table-driven sub-tests)
- [ ] No hardcoded values (capability default removed, not replaced with another magic string without justification)
- [ ] Documentation updated (`wiki/modules/approval.md`, ADR 0018)
- [ ] No unnecessary scope additions (Tier-1 cap split deferred)
- [ ] Self-contained — every patch grounded in this plan's file:line anchors

## Risks

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| BE stale binary masks real fix (P2 already "fixed" in code) | High | Low | Task 1 + Task 13 force `-Build`; QA evidence rule |
| StageSummary rename breaks downstream consumers we don't know about | Low | High | grep entire FE for `quorum_kind`/`stage.label`; no external API clients per route truth table |
| Capability default change breaks legacy "QA Inbox Route" or seeded data | Medium | Medium | Existing seed uses `document.signoff` (already canonical); migration not needed |
| New 4xx codes need pt-BR mapping | High | Low | Task 9 + add entries to `errorMessages.ts` |
| FE seed reads `version` but legacy seed has `0` → `seedRouteEtag` skips | Medium | Medium | Acceptable for legacy row; document with explicit toast on cold-edit attempt |
| OpenAPI codegen drifts tests that mock `components['schemas']` shape | Medium | Low | Regenerate first, re-run vitest, fix call sites in Task 7 |
| `slog` output sink isn't stdout | Medium | Low | Task 1 captures both stdout+stderr; if slog goes elsewhere, trace `cmd/metaldocs-api/main.go` for handler init |
| `additionalProperties: true` on Problem hides typo'd code consts | Low | Low | unit test asserts exact code string |

## Notes

- **Why P1b root cause is left to Task 2 rather than guessed**: opaque `internal.unknown` envelope hides the real error; guessing risks adding sentinels that don't fire. Capture once, then fix. Task 1 (log piping) is the prerequisite that makes Task 2 a 5-minute investigation rather than a multi-PR detective story.
- **Why 2 PRs not 5**: scope is genuinely bounded — single module, one screen, ~10 files. PR-1..5 were larger because they spanned OpenAPI + ADR + FE rewrite + wiki rebuild. This is a regression sweep, not a redesign. Splitting into 2 PRs (BE contract+root-cause, FE wiring+UX) keeps reviews tight without artificial fragmentation.
- **Why StageSummary unification is in scope**: P5 is the *cause* of the FE/BE friction that surfaces P2 + P4 cosmetics. Leaving it festers another QA cycle. Cheap to fix in the same change.
- **Why no E2E suite added**: project already has 88 vitest tests; E2E is its own track per `wiki/quality/`. Adding unmocked vitest at the boundary is the right altitude — proves the contract without spinning up Playwright.
- **Wiki drift policy**: per CLAUDE.md §2, any code change referenced by wiki must update the doc's `Last verified` stamp in the same change. Task 12 owns this.
- **Hard-stop class re-check**: original QA report flagged this as hard-stop ("shared contract prerequisite + module-local FE+BE"). Plan respects that by sequencing BE-first; FE changes depend on regenerated codegen.
- **Closure path**: PR description must include the QA evidence block (7 gates green, network log clean, console clean, toast UX verified). Per CLAUDE.md "Evidence rule" — `implemented`/`fixed` insufficient without verification commands + outcomes.
