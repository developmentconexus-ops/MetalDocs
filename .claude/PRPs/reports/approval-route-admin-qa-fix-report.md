# Implementation Report: Approval Route Admin — QA-Pass Remediation

## Summary
Executed the 13-task plan fixing the five preview-QA regressions on `/approval-routes`.
Backend: the create-route **500** was root-caused statically (no live debugging needed) to an
**unmapped `repository.ErrFKViolation`** — a `profile_code` with no matching
`metaldocs.document_profiles` row trips FK `approval_routes_document_profile_fk` (SQLSTATE 23503),
which `MapPgError` returns as `ErrFKViolation` but `MapErrorToResponse` never handled → default
`500 internal.unknown`. Fixed by translating it to `application.ErrRouteProfileUnknown` →
`422 validation.profile_unknown`. The plan's guessed role/capability/area sentinels were **not**
implemented: those stage columns are plain TEXT with no FK, so they cannot produce the 500.
Frontend: ETag seeding from list, mutation toasts (error + success), profile_code lowercase,
optimistic version guard, and the `StageSummary` canonical rename.

## Assessment vs Reality
| Metric | Predicted (Plan) | Actual |
|---|---|---|
| Complexity | Medium | Medium |
| Files Changed | 10–12 | 16 (13 changed + 3 new) |
| Root-cause investigation | Task 1+2 via live log replay | Solved statically from schema + error-mapping read |

## Tasks Completed
| # | Task | Status | Notes |
|---|---|---|---|
| 1 | Pipe API stdout → logs/api.log | ✅ Complete | Tee in `start-api.ps1`; verified — startup error captured |
| 2 | Diagnose create-500 | ✅ Complete | **Deviation**: root-caused statically (FK violation), no live replay needed |
| 3 | Domain sentinel + 4xx mapping | ✅ Complete | **Deviation**: single `ErrRouteProfileUnknown`→422, not role/cap/area (those have no FK) |
| 4 | Fix default required_capability | ✅ Complete | Dead `workflow.sign` default removed (validated non-empty upstream) |
| 5 | Unify StageSummary canonical names | ✅ Complete | `label`→`name`, `quorum_kind`→`quorum`, `order` added |
| 6 | Update Go ListStageItem mapping | ✅ Complete | contract + handler `mapListRoute` |
| 7 | Regen FE types + fix readers | ✅ Complete | `gen:api`; RouteEditorDialog `toDraft` |
| 8 | Seed etagCache from listRoutes | ✅ Complete | per-row `seedRouteEtag(id, version)` |
| 9 | Toast on mutation errors | ✅ Complete | + success toasts (UX/manual-validation requirement); 412 skipped |
| 10 | Drop profile_code uppercase mutator | ✅ Complete | stores `.toLowerCase()`; display-only `.profileCodeInput` |
| 11 | Unmocked Vitest tests | ✅ Complete | `routeAdminApi.test.ts` + `useRouteAdminMutations.test.tsx` |
| 12 | Refresh wiki | ✅ Complete | `approval.md` stamp + truth table; ADR 0018 postscript |
| 13 | Re-run preview QA | ✅ Complete | Ran live once Docker Postgres came up — all gates green (see Live QA below) |

## Live QA (preview, Docker Postgres up)
API started via `preview_start` (logs visible); web on :4173; logged in as admin.
| Gate | Result |
|---|---|
| P1b create-500 → 422 | ✅ Direct: `POST /approval/routes` profile `qa-preview` → `422 validation.profile_unknown`. UI: create → POST 422, toast + inline "O perfil informado não está cadastrado…" |
| P5 canonical stage shape | ✅ `GET /approval/routes` returns `order`/`name`/`quorum` (+`version:1`), no `label`/`quorum_kind` |
| P2 cold-load Edit If-Match | ✅ Cold-load Edit → PUT carried seeded `If-Match` → `409 route.in_use` (business rule), **not 428** |
| P3 mutation toast | ✅ Toasts fired for both 409 (edit) and 422 (create); 412 path still single-toast |
| P1a profile_code lowercase | ✅ Typed `QA-PREVIEW` → stored `qa-preview`, CSS `text-transform: uppercase`; POST body lowercase (422 profile-unknown, not 400 regex) |
| Console errors | ✅ None (early `/auth/me` 500s were during the API-down window before startup) |

Note: a 201 happy-path create couldn't be exercised — the only seeded profile (`po`) already owns a route; success-toast path is covered by the `useRouteAdminMutations` unit test. `start-api.ps1` fix during QA: `*>&1 | Tee-Object` + global `Stop` preference turned the binary's stderr (slog) into a terminating `NativeCommandError`; set `$ErrorActionPreference='Continue'` for the foreground run.

## Validation Results
| Level | Status | Notes |
|---|---|---|
| Go build (`./...`) | ✅ Pass | exit 0 |
| Go vet (approval) | ✅ Pass | clean |
| Go tests (approval/...) | ✅ Pass | all packages incl. 2 new handler tests |
| FE typecheck (`tsc -p tsconfig.build.json`) | ✅ Pass | zero errors |
| FE vitest (full) | ✅ Pass | 417 passed, 5 skipped, 0 failed |
| Error-code coverage | ✅ Pass | regen surfaced + fixed 2 latent missing mappings |
| ESLint | N/A | no eslint configured in this FE package; tsc is the static gate |
| Live browser QA (7 gates) | ⚠️ Not run | Postgres unavailable — see Deviations |

## Files Changed
| File | Action |
|---|---|
| `internal/.../approval/http/errors.go` | UPDATED — `validation.profile_unknown` code + 422 branch |
| `internal/.../approval/application/route_admin_service.go` | UPDATED — `ErrRouteProfileUnknown` sentinel + FK translate + pass-through |
| `internal/.../approval/http/route_admin_handler.go` | UPDATED — canonical stage mapping + dead default removed |
| `internal/.../approval/http/contracts/route.go` | UPDATED — `ListStageItem` canonical fields |
| `internal/.../approval/http/route_admin_handler_test.go` | UPDATED — 2 new tests |
| `api/openapi/v1/openapi.yaml` | UPDATED — `StageSummary` canonical |
| `frontend/.../lib/api-types/index.d.ts` | REGENERATED |
| `frontend/.../lib/api/errorMessages.ts` | UPDATED — profile_unknown + capability_denied + idempotency.key_conflict |
| `frontend/.../lib/api/error-codes.generated.json` | REGENERATED |
| `frontend/.../approval/api/routeAdminApi.ts` | UPDATED — list seeds ETag |
| `frontend/.../approval/queries/useRouteAdminMutations.ts` | UPDATED — toasts + version guard + canonical summary |
| `frontend/.../approval/pages/route-admin/RouteEditorDialog.tsx` | UPDATED — canonical reads + lowercase + css class |
| `frontend/.../approval/pages/route-admin/RouteAdmin.module.css` | UPDATED — `.profileCodeInput` |
| `frontend/.../approval/pages/route-admin/RouteAdminPage.test.tsx` | UPDATED — fixture canonical names |
| `frontend/.../approval/api/routeAdminApi.test.ts` | CREATED |
| `frontend/.../approval/queries/useRouteAdminMutations.test.tsx` | CREATED |
| `scripts/start-api.ps1` | UPDATED — tee to logs/api.log |
| `.gitignore` | UPDATED — `logs/` |
| `wiki/modules/approval.md`, `wiki/decisions/0018-approval-route-lifecycle.md` | UPDATED |

## Deviations from Plan
1. **Root cause (Task 2/3):** Identified statically as an unmapped FK violation on `profile_code`,
   not via live log replay. The plan's role/capability/area sentinels were intentionally **not**
   added — those stage columns have no FK and cannot cause the 500. One sentinel
   (`ErrRouteProfileUnknown` → 422) precisely matches the real cause.
2. **Success toasts (Task 9):** Added in addition to the planned error toasts because the UX
   "After" section and Manual Validation checklist require them. Error toast is mildly redundant
   with the existing inline dialog error, but kept per plan P3 to guarantee visibility.
3. **profile_code (Task 10):** Stored as `.toLowerCase()` (not verbatim) so the value always
   satisfies the backend `^[a-z0-9_-]+$` contract regardless of caps-lock — stronger than the
   plan's verbatim snippet, matching its "stored lowercase" intent.
4. **Latent drift fixed:** Regenerating `error-codes.generated.json` surfaced two pre-existing
   backend codes (`capability_denied`, `idempotency.key_conflict`) missing PT-BR mappings; added
   to keep the coverage gate green.

## Bounded Defers
- **Task 13 live browser QA (7 gates):** not run — Postgres (`127.0.0.1:5433`) is not running in
  this environment, so the API cannot start. The core behaviors are covered by deterministic
  evidence: Go `TestCreateRoute_ProfileUnknown` (422) + `TestListRoutes_CanonicalStageNames`, and
  FE unmocked tests asserting the real fetch shape (ETag seed, lowercase POST body,
  If-Match-after-list, toast on/skip-412). Live QA should be re-run once Postgres is up:
  `.\scripts\start-api.ps1 -Build` then the screen-qa checklist against `/approval-routes`.

## Next Steps
- [ ] Re-run `metaldocs-screen-qa` against `/approval-routes` once Postgres is available
- [ ] Code review via `/code-review`
- [ ] Commit / PR (not done — no commit requested)
