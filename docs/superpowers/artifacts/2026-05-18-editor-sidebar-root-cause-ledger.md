# Editor Sidebar Root Cause Ledger

> Last updated: 2026-05-19
> Scope: Runtime evidence and mismatch classification for the editor right sidebar hardening plan.

## Runtime Snapshot

Docker Desktop was reinstalled during this session after the previous engine entered a broken WSL/Desktop state. Current runtime was rebuilt through the canonical MetalDocs path:

- `docker version`: client/server responding on `desktop-linux`
- `docker compose -f deploy/compose/docker-compose.yml --env-file .env up -d postgres redis minio`: passed
- `scripts/dev-bootstrap-baseline.ps1 -WithDevSeed`: passed
- `scripts/seed-system-blank-template.ps1`: passed
- `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents`: passed

Current health evidence:

```text
metaldocs-api process: running
GET /api/v1/health/live: 200
GET /api/v1/auth/me: 200
GET /api/v1/controlled-documents: 200
```

Current authenticated user:

```json
{"userId":"admin","tenantId":"ffffffff-ffff-ffff-ffff-ffffffffffff","username":"admin","displayName":"Administrator","mustChangePassword":false,"roles":["system_admin"]}
```

Current database counts after clean bootstrap:

```text
users: 6
profiles: 0
areas: 0
controlled_documents: 0
documents: 0
editor_sessions: 0
```

Current controlled documents response:

```json
{"items":[]}
```

## Runtime Reset Impact

The pre-reset editor document IDs used earlier in this session no longer exist after the curated bootstrap reset. This invalidates direct reuse of payloads from `d79a1974-ecdd-444e-a5c4-740e46a9bc52` and similar documents.

That is a `runtime prerequisite` discovery, not a feature behavior. The next E2E pass must create taxonomy and documents through real runtime flows before validating editor sidebar states.

## Classification

| ID | Symptom | Classification | Truth layer used | Root cause status | Fix status |
|---|---|---|---|---|---|
| B1 | `POST /api/v1/taxonomy/*` returned `500 internal server error` from UI setup | runtime prerequisite | runtime + wiki + code | confirmed before reset; must revalidate on clean runtime | local patch needs professional review |
| B2 | New editor document opened readonly with "Another user is editing this document" | shared contract prerequisite | runtime + code | partially confirmed before reset; must reproduce with fresh document | not fixed |
| B3 | Editor sidebar draft metadata showed empty profile/area | module-local implementation | runtime + code + contract | confirmed before reset | local patch needs contract review |
| B4 | Revision timeline showed `REVNaN - undefined` | shared contract prerequisite | runtime + OpenAPI/generated types + code | confirmed before reset | local patch needs contract review |
| B5 | Draft sidebar profile label sometimes showed raw `pop` before hydration | screen-local implementation | browser + TanStack Query | unconfirmed on clean runtime | not fixed |
| B6 | Approval instance query was requested while document was draft in one runtime log | screen-local implementation | logs + frontend code | unconfirmed on clean runtime | not fixed |
| B7 | `scripts/start-api.ps1` held the shell until timeout while API was healthy | workflow/tooling gap | execution truth | confirmed on clean runtime | not fixed |
| B8 | Approval route creator / automatic route assignment needed verification after legacy-memory concern | runtime prerequisite + shared contract prerequisite | runtime + DB + OpenAPI + frontend code + wiki | confirmed: runtime route catalog exists and finalize resolves by `controlled_documents.profile_code`; contract/wiki/frontend drift remains | investigation added; fixes pending |
| B9 | Approval route admin UI still defaults stage `required_capability` to legacy `doc.signoff` while runtime IAM uses `document.signoff` | screen-local implementation + wiki-memory drift | frontend code + DB runtime + IAM model + wiki | confirmed by source scan; current seeded runtime row is correct (`document.signoff`) | not fixed |

## Approval Route Creator / Profile Binding Debug Addendum

User memory was correct: MetalDocs still has an approval route catalog and the intended product model is profile-driven route assignment, not author-selected approvers.

Runtime truth gathered on 2026-05-19:

```text
approval_routes:
- profile_code=pop
- name=POP E2E Approval Route
- active=true
- version=1

approval_route_stages:
- stage_order=1
- name=Aprovacao
- required_role=approver
- required_capability=document.signoff
- area_code=general
- quorum=any_1_of
- on_eligibility_drift=fail_stage
```

Code/runtime truth:

- `POST /api/v1/documents/{id}/finalize` reads the document's `controlled_document_id`, loads `controlled_documents.profile_code`, then selects the newest active row in `approval_routes` for that profile before calling `SubmitRevisionForReview`.
- `RouteAdminService` supports create/update/deactivate of routes and persists route stages.
- Frontend route admin exists at `/approval-routes` through `frontend/apps/web/src/features/approval/pages/RouteAdminPage.tsx`.
- OpenAPI/generated surfaces now include `/api/v1/approval/routes`, but some wiki/module text still says these routes are spec missing.

Mismatches to carry into full debug:

- `screen-local implementation`: `RouteAdminPage` uses legacy `doc.signoff` defaults/fallbacks, while current IAM capability truth is `document.signoff`.
- `wiki-memory drift`: `wiki/modules/approval.md`, `wiki/concepts/authz-tiers.md`, and older workflow notes still reference `doc.submit` / `doc.signoff` in places where current code uses `document.submit` / `document.signoff`.
- `shared contract prerequisite`: route admin should be reviewed against generated OpenAPI/frontend types; `frontend/features/approval/api/approvalTypes.ts` remains handwritten despite generated route operations existing.
- `runtime prerequisite`: E2E setup must ensure a profile has an active route and at least one eligible actor in the route stage area/role before testing submit/finalize.

Important product conclusion:

- The document author should not need to manually choose approvers in the editor finalize flow.
- The route is selected automatically by document profile.
- Manual route selection that remains in older approval panels should be treated as legacy/screen-specific and must not be reintroduced into the governed editor sidebar flow without a new product decision.

## Next Evidence Needed

- Create taxonomy data through the real UI/API path on the clean runtime.
- Create a document through the real document wizard.
- Re-check `documents.created_by`, `documents.active_session_id`, and `editor_sessions.user_id`.
- Validate editor sidebar draft behavior against runtime API payloads.
- Validate route admin UI/API creates a route using current capability namespace (`document.signoff`) and that finalize automatically binds by profile without manual approver selection.
- Only then decide whether B2/B5/B6 require code changes.
