# Templates backlog

> **Last verified:** 2026-05-13
> **Scope:** Deferred items for the Templates feature (List screen + future screens).

## List screen — deferred items

- [ ] `TemplateDTO` missing `updated_at` field — `formatRelative` currently falls back to `created_at`. Add field in backend and update `TemplatesListPage.tsx:43`.
- [ ] Resolve `created_by` user_id → display name. Currently shows raw UUID passed as `author` to `TemplateCard`. Requires a user lookup (batch or eager join in API response).
- [x] Card grid gap aligned to tokenized 16px (`var(--sp-4)`) for list cards (pre-existing in code; re-verified in Plan 12.1 on 2026-05-13).
- [x] Mobile tab clipping at 375px fixed via horizontal scroll support on `TabBar` (`overflow-x: auto` + hidden scrollbar) (2026-05-13).
- [ ] `formatRelative` helper inlined in `TemplatesListPage.tsx:17` — promote to `lib/utils/formatRelative.ts` when a second caller appears.

## Integration Audit — 2026-05-13

Screen audited: `frontend/apps/web/design-source/templates/`

Gate evidence:
- `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/templates` -> pass
- `scripts/check-module-contract-sync.ps1 -Module templates` -> surfaces present; manual drift review required

| Item | Source | Runtime/API reality | Frontend reality | Classification | Action |
|---|---|---|---|---|---|
| Templates list fetch | design + keep list | `GET /api/v1/templates` exists and returns `200`; route/spec/generated files align | `useTemplatesQuery` calls `templates.listTemplates()` and the list page renders cards from the response | `screen-local integration fix` | keep in Plan 12.1; verify response mapping against real payload |
| Status tabs + counts | design + NOTES keep list | runtime supports draft/published/archived classification via template/version fields | tabs are computed locally from fetched rows; no shared contract change needed | `implemented and aligned` | keep in Plan 12.1 |
| Card title/version/status shell | design + keep list | list endpoint exposes name, latest version, published/archive fields | `TemplateCard` renders these already | `implemented and aligned` | keep in Plan 12.1 |
| Relative timestamp | backlog + current UI | API row does not expose `updated_at`; backlog confirms current fallback to `created_at` | `TemplatesListPage` formats `created_at` directly | `missing backend capability` | split prerequisite if true "updated" semantics are required; otherwise keep current fallback and document it |
| Author display name | backlog + keep list | API currently returns `created_by` UUID, not display name | card shows raw UUID-derived first token | `missing backend capability` | prerequisite backend/API work before this can match the design intent |
| Card description text | NOTES cut list | no `description` contract on list behavior that the screen can trust today | current list page does not render description | `defer` | keep out of Plan 12.1 until backend/product contract exists |
| Bound profile pills | NOTES cut list | module/docs say no list endpoint returns this association | no list UI wiring present | `defer` | leave out |
| "Em revisão" tab | NOTES cut list | templates lifecycle has `in_review`, but the list screen contract/design was already normalized to `Arquivados` | current UI follows the normalized 4-tab version | `defer` | leave out unless product wants review-tab semantics later |
| "Novo template" flow | design + keep list | `POST /api/v1/templates` exists and wizard route exists | button routes to `/templates/new`; wizard calls `templates.createTemplate()` | `screen-local integration fix` | include in Plan 12.1 scope only if the handoff from list -> wizard needs polish |
| Route ownership for `/templates` | current frontend routing | backend target route is healthy | `routes.tsx` has duplicate `path: "templates"` entries, including a redirect route that can shadow the list route | `screen-local integration fix` | fix in the screen PR before trusting list navigation |
| Frontend API layer | current frontend code | `v1` templates surface is real and contract-backed | feature contains both `api/templates.ts` and legacy `api/templates.ts` with incompatible shapes and old endpoints | `implemented but legacy-wired` | keep Plan 12.1 on `templates.ts`; do not broaden into full legacy purge unless needed by the touched screen path |
| TanStack wiring | current frontend code | runtime route is healthy and auth/session gate passes | list page already uses TanStack query, but only the list flow uses the newer path | `implemented but legacy-wired` | local cleanup is allowed if directly needed by the list screen |
| Mobile tab clipping | backlog | no backend dependency | UI/layout issue in `TabBar` and list styles | `screen-local integration fix` | can stay inside Plan 12.1 |
| Card grid gap token mismatch | backlog | no backend dependency | styling delta only | `screen-local integration fix` | can stay inside Plan 12.1 |
| Inlined `formatRelative` helper | backlog | no backend dependency | local utility concern only | `screen-local integration fix` | low priority; only do it if the screen diff already touches the helper |

Ready for implementation:
- templates list fetch and mapping
- status tabs and counts
- card shell rendering
- route fix for `/templates`
- mobile tab clipping
- card grid spacing
- list-to-wizard handoff polish

Prerequisites:
- backend/API support for `updated_at` if the screen must show true last-updated time
- backend/API support for author display name instead of raw `created_by`

Deferred:
- card description text
- bound profile pills
- any "in review" tab semantics beyond the current 4-tab screen

Verification needed next:
- rerun `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/templates` before implementation if startup/auth drift reappears
- inspect the real `GET /api/v1/templates` payload while implementing to confirm whether `meta.limit/offset` and list envelope still match `templates.listTemplates()`
- keep implementation on the `templates.ts` path; treat legacy `api/templates.ts` as separate cleanup unless the touched screen path still depends on it

## Plan 12.1 implementation update — 2026-05-13

Completed in this PR (screen-local):
- templates list path hardening on `templates.listTemplates()` for envelope/meta tolerance
- status tabs/counts behavior kept on real list data with deterministic status counting
- card shell consistency hardened for empty/missing name and author values
- mobile tab clipping fix in `TabBar.module.css`
- list -> wizard handoff verified unchanged (`/templates/new` route wiring remains correct)

Still prerequisite (not implemented in this PR):
- backend/API `updated_at` semantics for true last-updated display
- backend/API author display name instead of raw `created_by`

Still deferred (not implemented in this PR):
- card description text
- bound profile pills
- any `em revisao` tab semantics beyond the current 4-tab behavior
